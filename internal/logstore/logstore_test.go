package logstore

import (
	"encoding/json"
	"maps"
	"math/rand"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

func unselect(payload string) func([]events.Event) (events.Event, error) {
	return func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeUnselect, Payload: json.RawMessage(payload)}, nil
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
	if err := Find(root); fault.KindOf(err) != fault.Refusal || fault.CodeOf(err) != fault.NoStore {
		t.Fatalf("Find before Create = %v, want the no-store refusal", err)
	}
	if err := Create(root); err != nil {
		t.Fatal(err)
	}
	if fi, err := os.Stat(filepath.Join(r.Dir, ".tree/events")); err != nil || !fi.IsDir() {
		t.Fatalf(".tree/events missing under the top level: %v", err)
	}
	// hyp writes no output under .tree/: the outputs are in the trace
	if _, err := os.Stat(filepath.Join(r.Dir, ".tree/outputs")); !os.IsNotExist(err) {
		t.Fatalf(".tree/outputs: %v, want none", err)
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
	if _, err := TopLevel(t.TempDir()); fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
}

func TestUnionMergedFileReplaysSame(t *testing.T) {
	root := newStore(t)
	for i, actor := range []string{"a", "b", "a", "b", "a"} {
		if _, err := Append(root, actor, unselect(`{"role":"owner"}`)); err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
	}
	want, errs, err := Events(root)
	if err != nil || len(errs) != 0 || len(want) != 5 {
		t.Fatalf("Events = %d events, %v, %v", len(want), errs, err)
	}
	rng := rand.New(rand.NewSource(1))
	for _, actor := range []string{"a", "b"} {
		oldForm(t, root, actor)
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
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	oldForm(t, root, "a")
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
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
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
	if _, err := Append(root, "b", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	_, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{}, fault.Refuse(fault.GuardDraft, fault.NotWritten, "G-X", "open")
	})
	if fault.KindOf(err) != fault.Refusal {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".tree", "events", "a.jsonl")); !os.IsNotExist(err) {
		t.Fatalf("a.jsonl exists after a refused build: %v", err)
	}
}

func TestAppendFillsEnvelope(t *testing.T) {
	root := newStore(t)
	ev, err := Append(root, "a", unselect(`{"role":"owner"}`))
	if err != nil {
		t.Fatal(err)
	}
	if ev.V != 2 || !ValidID(ev.ID) || ev.Actor != "a" || ev.Node != ev.ID || ev.TS == "" {
		t.Fatalf("envelope not filled: %+v", ev)
	}
	next, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeUnselect, Node: "G-X", Payload: json.RawMessage(`{"role":"owner"}`)}, nil
	})
	if err != nil || next.Node != "G-X" || next.ID <= ev.ID {
		t.Fatalf("second event %+v, %v", next, err)
	}
}

func TestAppendRefusesBadActorAndLongWhy(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "Bad Actor", unselect(`{}`)); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("bad actor: %v", err)
	}
	_, err := Append(root, "a", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeUnselect, Payload: json.RawMessage(`{}`), Why: strings.Repeat("é", 601)}, nil
	})
	if fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("long why: %v", err)
	}
	if _, err := os.Stat(filepath.Join(root, ".tree", "events", "a.jsonl")); !os.IsNotExist(err) {
		t.Fatal("a refused why left a file")
	}
}

func TestAppendWithoutStoreIsRefusal(t *testing.T) {
	r := testutil.NewRepo(t)
	if _, err := Append(r.Dir, "a", unselect(`{}`)); fault.KindOf(err) != fault.Refusal {
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
	files, err := os.ReadDir(filepath.Dir(OutputPath(root, a)))
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
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
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
	line := `{"v":2,"id":"01M661QG600000000000000001","ts":"t","actor":"other","type":"unselect","node":"n","payload":{"role":"owner"},"why":""}`
	appendRaw(t, root, "a", line+"\n")
	evs, errs, err := Events(root)
	if err != nil || len(evs) != 0 || len(errs) != 1 {
		t.Fatalf("events %d, errs %v, err %v", len(evs), errs, err)
	}
}

// oldForm moves the actor's events from their files of one event into
// <actor>.jsonl, the one file per actor an earlier hyp appended to.
func oldForm(t *testing.T, root, actor string) {
	t.Helper()
	dir := filepath.Join(root, ".tree", "events", actor)
	paths, err := filepath.Glob(filepath.Join(dir, "*.json"))
	if err != nil {
		t.Fatal(err)
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		appendRaw(t, root, actor, string(data))
	}
	if err := os.RemoveAll(dir); err != nil {
		t.Fatal(err)
	}
}

// storeFiles is every file under .tree/events with its bytes.
func storeFiles(t *testing.T, root string) map[string]string {
	t.Helper()
	files := map[string]string{}
	base := filepath.Join(root, ".tree", "events")
	err := filepath.WalkDir(base, func(p string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		data, err := os.ReadFile(p)
		rel, _ := filepath.Rel(base, p)
		files[rel] = string(data)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return files
}

func TestAppendAddsOneFileAndChangesNone(t *testing.T) {
	root := newStore(t)
	for i, actor := range []string{"a", "a", "b"} {
		before := storeFiles(t, root)
		ev, err := Append(root, actor, unselect(`{"role":"owner"}`))
		if err != nil {
			t.Fatalf("append %d: %v", i, err)
		}
		after := storeFiles(t, root)
		if len(after) != len(before)+1 {
			t.Fatalf("append %d: %d files, want %d", i, len(after), len(before)+1)
		}
		for name, data := range before {
			if after[name] != data {
				t.Fatalf("append %d changed %s", i, name)
			}
		}
		name := filepath.Join(actor, ev.ID+".json")
		if !strings.HasSuffix(after[name], "\n") || strings.Count(after[name], "\n") != 1 {
			t.Fatalf("append %d: %s = %q, want one line", i, name, after[name])
		}
	}
	evs, errs, err := Events(root)
	if err != nil || len(errs) != 0 || len(evs) != 3 {
		t.Fatalf("Events = %d events, %v, %v", len(evs), errs, err)
	}
}

func TestOldAndNewFilesReplayTogether(t *testing.T) {
	root := newStore(t)
	for i := 0; i < 2; i++ {
		if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
			t.Fatal(err)
		}
	}
	oldForm(t, root, "a")
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	evs, errs, err := Events(root)
	if err != nil || len(errs) != 0 || len(evs) != 3 {
		t.Fatalf("Events = %d events, %v, %v", len(evs), errs, err)
	}
	if !slices.IsSortedFunc(evs, func(x, y events.Event) int { return strings.Compare(x.ID, y.ID) }) {
		t.Fatal("events of the old and the new files are not in id order")
	}
}

func TestEventFileOfAnotherActorIsReported(t *testing.T) {
	root := newStore(t)
	ev, err := Append(root, "a", unselect(`{"role":"owner"}`))
	if err != nil {
		t.Fatal(err)
	}
	from := filepath.Join(root, ".tree", "events", "a", ev.ID+".json")
	to := filepath.Join(root, ".tree", "events", "b")
	if err := os.MkdirAll(to, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(from, filepath.Join(to, ev.ID+".json")); err != nil {
		t.Fatal(err)
	}
	evs, errs, err := Events(root)
	if err != nil || len(evs) != 0 || len(errs) != 1 {
		t.Fatalf("Events = %d events, %v, %v; want one line error", len(evs), errs, err)
	}
	if want := filepath.Join("b", ev.ID+".json"); errs[0].File != want || errs[0].Line != 1 {
		t.Fatalf("line error = %+v, want %s line 1", errs[0], want)
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

func TestDuplicateLineReplaysOnce(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	for _, data := range storeFiles(t, root) {
		appendRaw(t, root, "a", data+data)
	}
	evs, errs, err := Events(root)
	if err != nil || len(errs) != 0 {
		t.Fatalf("Events: %v, %v", errs, err)
	}
	if len(evs) != 1 {
		t.Fatalf("Events = %d events, want the duplicated line once", len(evs))
	}
}

func TestSameIDAndActorOrdersByLineBytes(t *testing.T) {
	line := func(why string) string {
		b, err := json.Marshal(events.Event{
			V: events.Version, ID: "01M661QG600000000000NXHHVA", TS: "2026-11-19T00:00:00Z", Actor: "a",
			Type: events.TypeUnselect, Node: "01M661QG600000000000NXHHVA", Payload: json.RawMessage(`{"role":"owner"}`), Why: why,
		})
		if err != nil {
			t.Fatal(err)
		}
		return string(b)
	}
	first, second := line("a"), line("b")
	if first >= second {
		t.Fatalf("fixture lines are not in byte order: %q %q", first, second)
	}
	for name, order := range map[string][2]string{"ascending": {first, second}, "descending": {second, first}} {
		root := newStore(t)
		appendRaw(t, root, "a", order[0]+"\n"+order[1]+"\n")
		evs, errs, err := Events(root)
		if err != nil || len(errs) != 0 || len(evs) != 2 {
			t.Fatalf("%s: Events = %d events, %v, %v", name, len(evs), errs, err)
		}
		if evs[0].Why != "a" || evs[1].Why != "b" {
			t.Errorf("%s: order = %q, %q; two events with one id and actor go by line bytes whatever their place in the file", name, evs[0].Why, evs[1].Why)
		}
	}
}

func TestWithLockReadsAndWritesNothing(t *testing.T) {
	root := newStore(t)
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	before := storeFiles(t, root)
	called := 0
	err := WithLock(root, func(evs []events.Event) error {
		called++
		if len(evs) != 1 {
			t.Errorf("WithLock passed %d events, want 1", len(evs))
		}
		return nil
	})
	if err != nil || called != 1 {
		t.Fatalf("WithLock = %v, called %d times", err, called)
	}
	if after := storeFiles(t, root); !maps.Equal(after, before) {
		t.Fatal("WithLock wrote to the log")
	}
	// the error of f comes back unchanged, and the lock is released
	want := fault.Refuse(fault.NoGuard, fault.Code("stop"))
	if err := WithLock(root, func([]events.Event) error { return want }); err != want {
		t.Fatalf("WithLock returned %v, want f's error", err)
	}
	if _, err := Append(root, "a", unselect(`{"role":"owner"}`)); err != nil {
		t.Fatalf("Append after WithLock: %v", err)
	}
}

func TestWithLockHoldsAppendBack(t *testing.T) {
	root := newStore(t)
	done := make(chan error, 1)
	err := WithLock(root, func([]events.Event) error {
		go func() {
			_, err := Append(root, "a", unselect(`{"role":"owner"}`))
			done <- err
		}()
		time.Sleep(200 * time.Millisecond)
		evs, _, err := Events(root)
		if err != nil || len(evs) != 0 {
			t.Errorf("an Append wrote while WithLock held the lock: %d events, %v", len(evs), err)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if evs, _, _ := Events(root); len(evs) != 1 {
		t.Fatalf("Append after the lock: %d events, want 1", len(evs))
	}
}

func TestWithLockNeedsAStore(t *testing.T) {
	r := testutil.NewRepo(t)
	err := WithLock(r.Dir, func([]events.Event) error { return nil })
	if err == nil || fault.KindOf(err) != fault.Refusal {
		t.Fatalf("WithLock with no store = %v, want the no-store refusal", err)
	}
}

func TestPutOutputInACloneWithoutOutputsDir(t *testing.T) {
	// git keeps no empty folder, so a clone of a store has no .tree/outputs.
	root := newStore(t)
	if err := os.RemoveAll(filepath.Join(root, ".tree", "outputs")); err != nil {
		t.Fatal(err)
	}
	name, err := PutOutput(root, []byte("hello"))
	if err != nil {
		t.Fatalf("PutOutput without .tree/outputs: %v", err)
	}
	if _, err := os.Stat(OutputPath(root, name)); err != nil {
		t.Fatalf("output file: %v", err)
	}
}
