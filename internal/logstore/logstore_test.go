package logstore

import (
	"encoding/json"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/testutil"
)

func reset(payload string) func([]events.Event) (events.Event, error) {
	return func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeBudgetReset, Payload: json.RawMessage(payload)}, nil
	}
}

func newStore(t *testing.T) string {
	t.Helper()
	r := testutil.NewRepo(t)
	root, err := TopLevel(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestCreateAtTopLevel(t *testing.T) {
	r := testutil.NewRepo(t)
	sub := filepath.Join(r.Dir, "a", "b")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	root, err := TopLevel(sub)
	if err != nil {
		t.Fatal(err)
	}
	if root != r.Dir {
		t.Fatalf("TopLevel = %s, want %s", root, r.Dir)
	}
	if err := Find(root); err == nil || herr.KindOf(err) != herr.Refusal || !strings.Contains(err.Error(), "no store; run hyp add goal") {
		t.Fatalf("Find before Create = %v, want the no-store refusal", err)
	}
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{".tree/events", ".tree/outputs"} {
		if fi, err := os.Stat(filepath.Join(r.Dir, p)); err != nil || !fi.IsDir() {
			t.Fatalf("%s missing under the top level: %v", p, err)
		}
	}
	attrs, err := os.ReadFile(filepath.Join(r.Dir, ".tree", ".gitattributes"))
	if err != nil || string(attrs) != "events/*.jsonl merge=union\n" {
		t.Fatalf(".gitattributes = %q, %v", attrs, err)
	}
	if _, err := os.Stat(filepath.Join(sub, ".tree")); err == nil {
		t.Fatal(".tree created in the subdirectory")
	}
	if err := Find(root); err != nil {
		t.Fatal(err)
	}
}

func TestTopLevelOutsideRepoIsEnvironment(t *testing.T) {
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	if _, err := TopLevel(t.TempDir()); herr.KindOf(err) != herr.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
}

func TestUnionMergedFileReplaysSame(t *testing.T) {
	root := newStore(t)
	for i, actor := range []string{"a", "b", "a", "b", "a"} {
		if _, err := Append(root, actor, reset(`{"role":"owner"}`)); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	want, errs, err := Events(root)
	if err != nil || len(errs) != 0 || len(want) != 5 {
		t.Fatalf("Events = %d events, %v, %v", len(want), errs, err)
	}
	rng := rand.New(rand.NewSource(1))
	for _, actor := range []string{"a", "b"} {
		p := filepath.Join(root, ".tree", "events", actor+".jsonl")
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
		rng.Shuffle(len(lines), func(i, j int) { lines[i], lines[j] = lines[j], lines[i] })
		if err := os.WriteFile(p, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, errs, err := Events(root)
	if err != nil || len(errs) != 0 {
		t.Fatalf("Events after shuffle: %v, %v", errs, err)
	}
	ids := func(evs []events.Event) []string {
		var out []string
		for _, e := range evs {
			out = append(out, e.ID)
		}
		return out
	}
	if !slices.Equal(ids(got), ids(want)) {
		t.Fatalf("order changed:\n got %v\nwant %v", ids(got), ids(want))
	}
	if !slices.IsSorted(ids(got)) {
		t.Fatalf("ids not sorted: %v", ids(got))
	}
}

func TestTornLastLineIsReportedNotFatal(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "a", reset(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	appendRaw(t, root, "a", `{"v":2,"id":"01M661QG60000000`)
	evs, errs, err := Events(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || len(errs) != 1 {
		t.Fatalf("events %d, line errors %v; want 1 and 1", len(evs), errs)
	}
	if errs[0].File != "a.jsonl" || errs[0].Line != 2 {
		t.Fatalf("line error = %+v, want a.jsonl line 2", errs[0])
	}
}

func TestAppendAfterTornLineStartsNewLine(t *testing.T) {
	root := newStore(t)
	appendRaw(t, root, "a", `{"v":2,"id":"01M661QG60000000`)
	if _, err := Append(root, "a", reset(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	evs, errs, err := Events(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 || len(errs) != 1 {
		t.Fatalf("events %d, line errors %v; want the new event and the one torn line", len(evs), errs)
	}
}

func TestRefusedBuildWritesNothing(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "b", reset(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	_, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{}, herr.Refused("Draft: nope")
	})
	if herr.KindOf(err) != herr.Refusal {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".tree", "events", "a.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("a.jsonl exists after a refused build: %v", err)
	}
}

func TestAppendFillsEnvelope(t *testing.T) {
	root := newStore(t)
	ev, err := Append(root, "a", reset(`{"role":"owner"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.V != 2 || !ValidID(ev.ID) || ev.Actor != "a" || ev.Node != ev.ID || ev.TS == "" {
		t.Fatalf("envelope not filled: %+v", ev)
	}
	next, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeBudgetReset, Node: "G-X", Payload: json.RawMessage(`{"role":"owner"}`)}, nil
	})
	if err != nil || next.Node != "G-X" || next.ID <= ev.ID {
		t.Fatalf("second event %+v, %v", next, err)
	}
}

func TestAppendRefusesBadActorAndLongWhy(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "Bad Actor", reset(`{}`)); herr.KindOf(err) != herr.BadArguments {
		t.Fatalf("bad actor: %v", err)
	}
	_, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeBudgetReset, Payload: json.RawMessage(`{}`), Why: strings.Repeat("é", 601)}, nil
	})
	if herr.KindOf(err) != herr.BadArguments {
		t.Fatalf("long why: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".tree", "events", "a.jsonl")); !os.IsNotExist(err) {
		t.Fatal("a refused why left a file")
	}
}

func TestAppendWithoutStoreIsRefusal(t *testing.T) {
	r := testutil.NewRepo(t)
	if _, err := Append(r.Dir, "a", reset(`{}`)); herr.KindOf(err) != herr.Refusal {
		t.Fatalf("err = %v, want the no-store refusal", err)
	}
}

func TestPutOutputOnceBySha256(t *testing.T) {
	root := newStore(t)
	a, err := PutOutput(root, []byte("hello"))
	if err != nil {
		t.Fatal(err)
	}
	b, err := PutOutput(root, []byte("hello"))
	if err != nil || a != b {
		t.Fatalf("hashes %s %s, err %v", a, b, err)
	}
	if a != "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" {
		t.Fatalf("hash = %s, want sha256 of hello", a)
	}
	files, err := os.ReadDir(filepath.Join(root, ".tree", "outputs"))
	if err != nil || len(files) != 1 {
		t.Fatalf("outputs = %v, %v; want one file", files, err)
	}
	if body, err := os.ReadFile(OutputPath(root, a)); err != nil || string(body) != "hello" {
		t.Fatalf("stored %q, %v", body, err)
	}
	empty, err := PutOutput(root, nil)
	if err != nil || empty != "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" {
		t.Fatalf("empty output hash %s, %v", empty, err)
	}
}

func TestStoreWritesNoGitState(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	before := testutil.Snapshot(t, r.Dir)
	root, err := TopLevel(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	if _, err := Append(root, "a", reset(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := PutOutput(root, []byte("x")); err != nil {
		t.Fatal(err)
	}
	evs, _, err := Events(root)
	if err != nil || len(evs) != 1 {
		t.Fatalf("the act did not run: %d events, %v", len(evs), err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("the store changed the repository's git state")
	}
}

func TestIDAfterSortsAfterNewest(t *testing.T) {
	now := time.Now().UTC()
	last := newID(now.Add(time.Millisecond))
	got := idAfter(now, last)
	if got == "" || got <= last {
		t.Fatalf("idAfter = %q, want an id greater than %q", got, last)
	}
	if got := idAfter(now, ""); !ValidID(got) {
		t.Fatalf("idAfter with no last = %q", got)
	}
}

func TestEventWithWrongActorIsALineError(t *testing.T) {
	root := newStore(t)
	line := `{"v":2,"id":"01M661QG600000000000000001","ts":"t","actor":"other","type":"budget.reset","node":"n","payload":{"role":"owner"},"why":""}`
	appendRaw(t, root, "a", line+"\n")
	evs, errs, err := Events(root)
	if err != nil || len(evs) != 0 || len(errs) != 1 {
		t.Fatalf("events %d, errs %v, err %v", len(evs), errs, err)
	}
}

// appendRaw writes s to an actor's file without a store call, as a crash or a
// merge would leave it.
func appendRaw(t *testing.T, root, actor, s string) {
	t.Helper()
	f, err := os.OpenFile(filepath.Join(root, ".tree", "events", actor+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if _, err := f.WriteString(s); err != nil {
		t.Fatal(err)
	}
}
