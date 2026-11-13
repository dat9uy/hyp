package store

import (
	"bytes"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/oklog/ulid/v2"
)

// TestAppendSortsAfterNewest: the newest event's id is from a clock far ahead
// of this one, as on a branch merged in from a machine whose clock ran fast.
// Each appended id still sorts after every id the store holds.
func TestAppendSortsAfterNewest(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	// 7ZZZZZZZZZ is the greatest ULID time, year 10889.
	future := "7ZZZZZZZZZ0000000000000000"
	f, err := os.OpenFile(filepath.Join(root, ".tree", "events", "z.jsonl"), os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString(`{"v":1,"id":"` + future + `","ts":"2026-10-27T00:00:00Z","actor":"z","type":"node.add","node":"` + future + `","payload":{"kind":"DesignQuestion","parent":"` + g.ID + `","title":"q"},"why":"w"}` + "\n")
	f.Close()
	var got []string
	for _, actor := range []string{"a", "b", "a"} {
		ev, err := Append(root, actor, nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
		if err != nil {
			t.Fatal(err)
		}
		got = append(got, ev.ID)
	}
	evs, _, err := Events(root)
	if err != nil {
		t.Fatal(err)
	}
	var all []string
	for _, ev := range evs {
		all = append(all, ev.ID)
	}
	if want := append([]string{g.ID, future}, got...); !slices.Equal(all, want) {
		t.Fatalf("merged ids %v, want %v: each appended id after %s", all, want, future)
	}
}

func gitRun(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-c", "user.name=t", "-c", "user.email=t@t"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git %v: %v: %s", args, err, out)
	}
}

// TestUnionMergeKeepsBothBranches: init writes .tree/.gitattributes holding
// one line; with it, git merges two branches that each appended to the same
// actor's events file with exit 0, and the file holds every line of both
// branches once and no other line.
func TestUnionMergeKeepsBothBranches(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q", "-b", "main")
	if err := Init(repo); err != nil {
		t.Fatal(err)
	}
	ga, err := os.ReadFile(filepath.Join(repo, ".tree", ".gitattributes"))
	if err != nil || string(ga) != "events/*.jsonl merge=union\n" {
		t.Fatalf(".tree/.gitattributes = %q (%v), want exactly the line events/*.jsonl merge=union", ga, err)
	}
	g, err := Append(repo, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	gitRun(t, repo, "add", "-A")
	gitRun(t, repo, "commit", "-qm", "goal")
	file := filepath.Join(repo, ".tree", "events", "a.jsonl")
	read := func() []string {
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		return strings.SplitAfter(strings.TrimSuffix(string(data), "\n"), "\n")
	}
	gitRun(t, repo, "checkout", "-q", "-b", "side")
	for range 2 {
		if _, err := Append(repo, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "side", "why")); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "commit", "-qam", "side")
	side := read()
	gitRun(t, repo, "checkout", "-q", "main")
	for range 2 {
		if _, err := Append(repo, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "main", "why")); err != nil {
			t.Fatal(err)
		}
	}
	gitRun(t, repo, "commit", "-qam", "main")
	main := read()
	cmd := exec.Command("git", "-c", "user.name=t", "-c", "user.email=t@t", "merge", "--no-edit", "side")
	cmd.Dir = repo
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git merge: %v: %s", err, out)
	}
	merged := read()
	want := map[string]bool{}
	for _, l := range append(side, main...) {
		want[strings.TrimSuffix(l, "\n")] = true
	}
	got := map[string]int{}
	for _, l := range merged {
		got[strings.TrimSuffix(l, "\n")]++
	}
	if len(merged) != len(want) {
		t.Errorf("merged file has %d lines, want %d", len(merged), len(want))
	}
	for l := range want {
		if got[l] != 1 {
			t.Errorf("line held %d times, want once: %s", got[l], l)
		}
	}
	for l := range got {
		if !want[l] {
			t.Errorf("line of neither branch: %q", l)
		}
	}
}

// TestClockBehindOnTwoClones: two clones of a store whose newest id is from a
// clock far ahead each append one event under the same actor. The two ids
// differ, both sort after the newest, and the events file holding both
// clones' lines, in either order, replays every event with no rejection.
func TestClockBehindOnTwoClones(t *testing.T) {
	root := newStore(t)
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	// A far time below the greatest, so the next millisecond exists.
	future := "7ZZZZZZZZY0000000000000000"
	writeLines(t, root, "z", `{"v":1,"id":"`+future+`","ts":"2026-10-27T00:00:00Z","actor":"z","type":"node.add","node":"`+future+`","payload":{"kind":"DesignQuestion","parent":"`+g.ID+`","title":"q"},"why":"w"}`)
	base, err := os.ReadFile(filepath.Join(root, ".tree", "events", "a.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	var added [2]string
	var ids [2]string
	for i := range 2 {
		clone := filepath.Join(t.TempDir(), "clone")
		if err := os.CopyFS(clone, os.DirFS(root)); err != nil {
			t.Fatal(err)
		}
		ev, err := Append(clone, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
		if err != nil {
			t.Fatal(err)
		}
		data, _ := os.ReadFile(filepath.Join(clone, ".tree", "events", "a.jsonl"))
		added[i], ids[i] = string(data[len(base):]), ev.ID
	}
	if ids[0] == ids[1] || ids[0] <= future || ids[1] <= future {
		t.Fatalf("clone ids %v, want two different ids after %s", ids, future)
	}
	var applied [2][]string
	for i, order := range [][2]int{{0, 1}, {1, 0}} {
		merged := newStore(t)
		data := string(base) + added[order[0]] + added[order[1]]
		if err := os.WriteFile(filepath.Join(merged, ".tree", "events", "a.jsonl"), []byte(data), 0o644); err != nil {
			t.Fatal(err)
		}
		zs, _ := os.ReadFile(filepath.Join(root, ".tree", "events", "z.jsonl"))
		if err := os.WriteFile(filepath.Join(merged, ".tree", "events", "z.jsonl"), zs, 0o644); err != nil {
			t.Fatal(err)
		}
		p, lineErrs, err := Load(merged)
		if err != nil || len(lineErrs) > 0 || len(p.Rejections) > 0 {
			t.Fatalf("order %v: %v %v %v", order, err, lineErrs, p.Rejections)
		}
		applied[i] = ids2(p.Applied)
	}
	if !slices.Equal(applied[0], applied[1]) || len(applied[0]) != 4 {
		t.Errorf("applied %v and %v, want the same four events", applied[0], applied[1])
	}
}

func ids2(evs []event.Event) []string {
	var out []string
	for _, ev := range evs {
		out = append(out, ev.ID)
	}
	return out
}

// TestSameIDLinesAnyOrder: one events file holds two lines with one id, as a
// hand edit could write. Whichever comes first in the file, replay as of that
// id applies the same event as the store cut at it, and the cut writes the
// same bytes. The pairs differ in the actor field, and in key order alone.
func TestSameIDLinesAnyOrder(t *testing.T) {
	const x = "01JB0000000000000000000005"
	goal := func(actor, title string) string {
		return `{"v":1,"id":"` + x + `","ts":"2026-10-27T00:00:00Z","actor":"` + actor + `","type":"node.add","node":"` + x + `","payload":{"kind":"Goal","parent":"","title":"` + title + `"},"why":"w"}`
	}
	keysSwapped := `{"id":"` + x + `","v":1,"ts":"2026-10-27T00:00:00Z","actor":"a","type":"node.add","node":"` + x + `","payload":{"kind":"Goal","parent":"","title":"y"},"why":"w"}`
	for _, pair := range [][2]string{{goal("b", "x"), goal("a", "y")}, {goal("a", "y"), keysSwapped}} {
		var cuts [2]map[string][]byte
		for i, lines := range [][2]string{pair, {pair[1], pair[0]}} {
			src := newStore(t)
			writeLines(t, src, "a", lines[0], lines[1])
			asOf, _, err := LoadAsOf(src, x)
			if err != nil {
				t.Fatal(err)
			}
			dest := writeCut(t, src, x)
			cut, _, err := Load(dest)
			if err != nil {
				t.Fatal(err)
			}
			if a, c := asOf.Nodes[x], cut.Nodes[x]; a == nil || c == nil || a.Title != c.Title {
				t.Errorf("order %d: as-of node %+v, cut node %+v", i, a, c)
			}
			cuts[i] = files(t, filepath.Join(dest, ".tree", "events"))
		}
		if !maps.EqualFunc(cuts[0], cuts[1], bytes.Equal) {
			t.Errorf("cut bytes depend on line order: %q vs %q", cuts[0]["a.jsonl"], cuts[1]["a.jsonl"])
		}
	}
}

// TestAppendInGreatestMillisecond: in the greatest ULID millisecond two
// clones still append different ids after the newest, and one below the
// greatest ULID still gets an id after it. After the greatest ULID, or after
// an id above it that replay admits, no ULID sorts, so the append is refused
// and writes nothing.
func TestAppendInGreatestMillisecond(t *testing.T) {
	for _, c := range []struct {
		last            string
		refused, differ bool
	}{
		{"7ZZZZZZZZZ0000000000000000", false, true},
		{"7ZZZZZZZZZZZZZZZZZZZZZZZZY", false, false},
		{"7ZZZZZZZZZZZZZZZZZZZZZZZZZ", true, false},
		{"80000000000000000000000000", true, false},
	} {
		root := newStore(t)
		writeLines(t, root, "z", goalLine(c.last, "z"))
		var ids []string
		for range 2 {
			clone := filepath.Join(t.TempDir(), "clone")
			if err := os.CopyFS(clone, os.DirFS(root)); err != nil {
				t.Fatal(err)
			}
			ev, err := Append(clone, "a", nodeAdd(t, event.KindGoal, "", "g", "why"))
			if c.refused {
				if err == nil || !strings.HasPrefix(err.Error(), "refused: ") || lines(t, clone, "a") != 0 {
					t.Fatalf("after %s: err %v, %d lines; want a refusal that writes nothing", c.last, err, lines(t, clone, "a"))
				}
				continue
			}
			if err != nil || ev.ID <= c.last {
				t.Fatalf("after %s: id %s, err %v", c.last, ev.ID, err)
			}
			ids = append(ids, ev.ID)
		}
		if c.differ && ids[0] == ids[1] {
			t.Errorf("two clones appended one id %s", ids[0])
		}
	}
}

// TestAppendAfterRejectedID: a line whose id replay rejects for its form, such
// as a lower-case ULID or a short junk id that sorts after every applied id,
// neither blocks an append nor counts as the newest id; the appended id sorts
// after every id of the id form.
func TestAppendAfterRejectedID(t *testing.T) {
	for _, bad := range []string{"01mzzzzzzzzzzzzzzzzzzzzzzz", "zz", "7ZZZZZZZZZZZZZZZZZZZZZZZZ"} {
		root := newStore(t)
		g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "g", "why"))
		if err != nil {
			t.Fatal(err)
		}
		writeLines(t, root, "z", goalLine(bad, "z"))
		ev, err := Append(root, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
		if err != nil || ev.ID <= g.ID {
			t.Errorf("after rejected id %q: id %s, err %v; want an id after %s", bad, ev.ID, err, g.ID)
		}
	}
}

// TestAppendAfterRejectedIDKeepsClock: a line whose id replay rejects for its
// form does not move the appended id off the clock. A lower-case ULID such as
// 01mzzzzzzzzzzzzzzzzzzzzzzz sorts after every upper-case id of today; when it
// counted as the newest id, the appended id took the millisecond after it.
// The appended id's millisecond is the clock's, read before and after Append.
func TestAppendAfterRejectedIDKeepsClock(t *testing.T) {
	root := newStore(t)
	writeLines(t, root, "z", goalLine("01mzzzzzzzzzzzzzzzzzzzzzzz", "z"))
	before := ulid.Timestamp(time.Now())
	ev, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "g", "why"))
	after := ulid.Timestamp(time.Now())
	if err != nil {
		t.Fatal(err)
	}
	u, err := ulid.ParseStrict(ev.ID)
	if err != nil {
		t.Fatal(err)
	}
	if ms := u.Time(); ms < before || ms > after {
		t.Errorf("after a rejected lower-case id: appended id %s has millisecond %d, want the clock's, %d to %d", ev.ID, ms, before, after)
	}
}
