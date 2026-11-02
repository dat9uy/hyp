package store

import (
	"bytes"
	"errors"
	"fmt"
	"io/fs"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// files maps each regular file under dir, slash-relative, to its bytes.
func files(t *testing.T, dir string) map[string][]byte {
	t.Helper()
	out := map[string][]byte{}
	err := filepath.WalkDir(dir, func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(dir, p)
		data, err := os.ReadFile(p)
		out[filepath.ToSlash(rel)] = data
		return err
	})
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		t.Fatal(err)
	}
	return out
}

func writeCut(t *testing.T, src, at string) string {
	t.Helper()
	c, err := CutAt(src, at)
	if err != nil {
		t.Fatalf("CutAt %s: %v", at, err)
	}
	dest := filepath.Join(t.TempDir(), "cut")
	if err := c.Write(dest); err != nil {
		t.Fatal(err)
	}
	return dest
}

// synthDir is the synthetic store generate.sh writes with the current hyp.
const synthDir = "../testutil/testdata/synth"

// TestCutMatchesExpected cuts the synthetic store at chosen events and checks
// each actor file against the first lines of the store's own file, and the
// evidence against the E-* folders of the Experiments the cut keeps. An actor
// with no event before the cut gets no file. e-short (E-8Y7V9P) is added at
// event 29, so a cut before it drops its command record.
func TestCutMatchesExpected(t *testing.T) {
	src := files(t, filepath.Join(synthDir, ".tree", "events"))
	evidence := files(t, filepath.Join(synthDir, ".tree", "evidence"))
	if len(src) != 4 || len(evidence) != 4 {
		t.Fatalf("synthetic store: %d actor files, %d evidence files; want 4 and 4", len(src), len(evidence))
	}
	for _, c := range []struct {
		name, at string
		lines    map[string]int // actor file → lines kept
		evidence bool           // E-8Y7V9P kept
	}{
		{"first event", "01M6A4QXKHN55JF4C0N6KMMKHH", map[string]int{"runner.jsonl": 1}, false},
		{"first verifier", "01M6A4QYVXYH98DYPZZQJJWA4A", map[string]int{"runner.jsonl": 19, "verifier-1.jsonl": 1}, false},
		{"e-mem added", "01M6A4QZ108X4P54TTZGMB1818", map[string]int{"runner.jsonl": 21, "verifier-1.jsonl": 1, "verifier-2.jsonl": 1}, false},
		{"before e-short", "01M6A4QZCPJK9PST0B1Y8E85PK", map[string]int{"runner.jsonl": 26, "verifier-1.jsonl": 1, "verifier-2.jsonl": 1, "verifier-3.jsonl": 1}, false},
		{"e-short added", "01M6A4QZECV6PYM2QPQS8Y7V9P", map[string]int{"runner.jsonl": 27, "verifier-1.jsonl": 1, "verifier-2.jsonl": 1, "verifier-3.jsonl": 1}, true},
		{"last event", "01M6A4R05P4KF4KA2HKD19W65Q", map[string]int{"runner.jsonl": 31, "verifier-1.jsonl": 3, "verifier-2.jsonl": 1, "verifier-3.jsonl": 1}, true},
	} {
		got := writeCut(t, synthDir, c.at)
		want := map[string][]byte{}
		for actor, n := range c.lines {
			lines := bytes.SplitAfter(src[actor], []byte("\n"))
			if n > len(lines) {
				t.Fatalf("%s: %s has %d lines, want %d", c.name, actor, len(lines), n)
			}
			want[actor] = bytes.Join(lines[:n], nil)
		}
		if g := files(t, filepath.Join(got, ".tree", "events")); !maps.EqualFunc(g, want, bytes.Equal) {
			t.Errorf("%s: events differ: got %v, want %v", c.name, lineCounts(g), c.lines)
		}
		wantEv := map[string][]byte{}
		if c.evidence {
			wantEv = evidence
		}
		if g := files(t, filepath.Join(got, ".tree", "evidence")); !maps.EqualFunc(g, wantEv, bytes.Equal) {
			t.Errorf("%s: evidence differs: got %v, want %v", c.name, slices.Sorted(maps.Keys(g)), slices.Sorted(maps.Keys(wantEv)))
		}
	}
}

// lineCounts maps each actor file to its number of lines, for a failure message.
func lineCounts(fs map[string][]byte) map[string]int {
	out := map[string]int{}
	for k, b := range fs {
		out[k] = bytes.Count(b, []byte("\n"))
	}
	return out
}

// writeLines writes raw event lines to <root>/.tree/events/<actor>.jsonl.
func writeLines(t *testing.T, root, actor string, lines ...string) {
	t.Helper()
	var b bytes.Buffer
	for _, l := range lines {
		b.WriteString(l + "\n")
	}
	if err := os.WriteFile(filepath.Join(root, ".tree", "events", actor+".jsonl"), b.Bytes(), 0o644); err != nil {
		t.Fatal(err)
	}
}

func goalLine(id, actor string) string {
	return fmt.Sprintf(`{"v":1,"id":%q,"ts":"2026-10-27T00:00:00Z","actor":%q,"type":"node.add","node":%q,"payload":{"kind":"Goal","parent":"","title":"g"},"why":"w"}`, id, actor, id)
}

func ids(evs []event.Event) []string {
	var out []string
	for _, ev := range evs {
		out = append(out, ev.ID)
	}
	return out
}

// TestCutFollowsMergeOrder: an actor file out of id order. The merge puts
// 01B (file b) before 01C and 01A (file a, in that order), so the cut at 01B
// keeps 01B alone; comparing ids would also keep 01A.
func TestCutFollowsMergeOrder(t *testing.T) {
	src := newStore(t)
	writeLines(t, src, "a", goalLine("01C", "a"), goalLine("01A", "a"))
	writeLines(t, src, "b", goalLine("01B", "b"))
	all, _, err := Events(src)
	if err != nil {
		t.Fatal(err)
	}
	for i, ev := range all {
		dest := writeCut(t, src, ev.ID)
		got, lineErrs, err := Events(dest)
		if err != nil || len(lineErrs) > 0 {
			t.Fatal(err, lineErrs)
		}
		if g, w := ids(got), ids(all[:i+1]); !slices.Equal(g, w) {
			t.Errorf("cut at %s: events %v, want %v", ev.ID, g, w)
		}
		cut, _, _ := Load(dest)
		asOf, _, _ := LoadAsOf(src, ev.ID)
		if g, w := ids(cut.Applied), ids(asOf.Applied); !slices.Equal(g, w) {
			t.Errorf("cut at %s: applied %v, as-of applied %v", ev.ID, g, w)
		}
	}
	if g := ids(all); !slices.Equal(g, []string{"01B", "01C", "01A"}) {
		t.Fatalf("merge order %v", g)
	}
	if _, err := os.Stat(filepath.Join(writeCut(t, src, "01B"), ".tree", "events", "a.jsonl")); !os.IsNotExist(err) {
		t.Errorf("a.jsonl written with no event before the cut: %v", err)
	}
}

// TestCutBreaksTiesByActor: two actors hold the same id. Replay applies the
// copy of the actor whose name sorts first, "a" before "a-b"; so does the cut.
func TestCutBreaksTiesByActor(t *testing.T) {
	src := newStore(t)
	writeLines(t, src, "a", goalLine("01X", "a"))
	writeLines(t, src, "a-b", goalLine("01X", "a-b"))
	all, _, err := Events(src)
	if err != nil || len(all) != 2 || all[0].Actor != "a" {
		t.Fatalf("events %v (%v)", all, err)
	}
	got := files(t, filepath.Join(writeCut(t, src, "01X"), ".tree", "events"))
	if want := goalLine("01X", "a") + "\n"; len(got) != 1 || string(got["a.jsonl"]) != want {
		t.Fatalf("cut kept %v, want a.jsonl alone", slices.Sorted(maps.Keys(got)))
	}
}

func TestCutDropsLaterEvidence(t *testing.T) {
	src := newStore(t)
	g, err := Append(src, "a", nodeAdd(t, event.KindGoal, "", "g", "w"))
	if err != nil {
		t.Fatal(err)
	}
	q, _ := Append(src, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "w"))
	h, _ := Append(src, "a", nodeAdd(t, event.KindHypothesis, q.ID, "h", "w"))
	add := nodeAdd(t, event.KindExperiment, h.ID, "e1", "w")
	e1, err := Append(src, "a", event.Event{Type: add.Type, Payload: []byte(`{"kind":"Experiment","parent":"` + h.ID + `","title":"e1","recipe":"none"}`), Why: "w"})
	if err != nil {
		t.Fatal(err)
	}
	e2, err := Append(src, "a", event.Event{Type: add.Type, Payload: []byte(`{"kind":"Experiment","parent":"` + h.ID + `","title":"e2","recipe":"none"}`), Why: "w"})
	if err != nil {
		t.Fatal(err)
	}
	ev := filepath.Join(src, ".tree", "evidence")
	d1, d2 := "E-"+e1.ID[len(e1.ID)-6:], "E-"+e2.ID[len(e2.ID)-6:]
	for _, rel := range []string{d1 + "/log.txt", d2 + "/log.txt", d2 + ".txt", "notes.md"} {
		p := filepath.Join(ev, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(rel), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got := slices.Sorted(maps.Keys(files(t, filepath.Join(writeCut(t, src, e1.ID), ".tree", "evidence"))))
	if want := []string{d1 + "/log.txt", "notes.md"}; !slices.Equal(got, want) {
		t.Errorf("evidence at the cut = %v, want %v", got, want)
	}
}

func TestCutRefusals(t *testing.T) {
	src := newStore(t)
	g, err := Append(src, "a", nodeAdd(t, event.KindGoal, "", "g", "w"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := CutAt(src, "01NOPE"); !errors.Is(err, ErrNoEvent) {
		t.Errorf("unknown event: %v, want ErrNoEvent", err)
	}
	if _, err := CutAt(t.TempDir(), g.ID); !errors.Is(err, ErrNoStore) {
		t.Errorf("no store: %v, want ErrNoStore", err)
	}
	f, _ := os.OpenFile(filepath.Join(src, ".tree", "events", "a.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	f.WriteString("{not json\n")
	f.Close()
	if _, err := CutAt(src, g.ID); err == nil || errors.Is(err, ErrNoEvent) {
		t.Errorf("bad line: %v, want a line error", err)
	}
	c, err := CutAt(newStoreWith(t, g), g.ID)
	if err != nil {
		t.Fatal(err)
	}
	dest := t.TempDir()
	if err := os.Mkdir(filepath.Join(dest, ".tree"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := c.Write(dest); err == nil {
		t.Error("Write over an existing .tree succeeded")
	}
}

// newStoreWith is a store holding the one event ev.
func newStoreWith(t *testing.T, ev event.Event) string {
	t.Helper()
	root := newStore(t)
	writeLines(t, root, ev.Actor, goalLine(ev.ID, ev.Actor))
	return root
}
