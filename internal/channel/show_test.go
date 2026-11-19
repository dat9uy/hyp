package channel

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/testutil"
)

// trace is the trace folder of the clone that holds dir.
func trace(t *testing.T, dir string) logstore.Trace {
	t.Helper()
	tr, err := logstore.TraceOf(dir)
	if err != nil {
		t.Fatal(err)
	}
	return tr
}

// reads are the read events of actor in the clone's trace, oldest first.
func reads(t *testing.T, dir, actor string) []events.Read {
	t.Helper()
	evs, _, err := trace(t, dir).Events()
	if err != nil {
		t.Fatal(err)
	}
	var out []events.Read
	for _, ev := range evs {
		if ev.Type != events.TypeRead || ev.Actor != actor {
			continue
		}
		var p events.Read
		if err := events.Decode(ev, &p); err != nil {
			t.Fatal(err)
		}
		out = append(out, p)
	}
	return out
}

// stored is the output of the trace named hash.
func stored(t *testing.T, dir, hash string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(trace(t, dir).Outputs(), hash))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// TestShowPrintsWhatTheActorGot: the owner's show holds each read of the
// runner byte for byte, as the runner got it, after the log moved on.
func TestShowPrintsWhatTheActorGot(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	_, before, _ := run(t, "status")
	// a run moves the log on, so a fresh status differs from the first
	c := f.leaf(t, map[string]string{"counter": "ok\n"})
	code, ran, ranErr := run(t, "run", c, f.goal+"/line-1")
	if code != 0 {
		t.Fatalf("run: code %d, stderr %q", code, ranErr)
	}
	_, after, _ := run(t, "status")
	if after == before {
		t.Fatal("test setup: the run did not change what status prints")
	}
	got := reads(t, f.r.Dir, "runner-1")
	if len(got) != 3 {
		t.Fatalf("%d reads of runner-1, want 3: status, run, status", len(got))
	}
	for i, want := range []string{before, ran, after} {
		if got[i].Role != "runner" || got[i].Exit != 0 || stored(t, f.r.Dir, got[i].Stdout) != want {
			t.Fatalf("read %d = %+v, holds %q, want %q", i+1, got[i], stored(t, f.r.Dir, got[i].Stdout), want)
		}
	}
	if got[1].Command[0] != "run" || got[1].Command[1] != c {
		t.Fatalf("read 2 command = %v", got[1].Command)
	}

	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	code, shown, errs := run(t, "show", "runner-1")
	if code != 0 || errs != "" {
		t.Fatalf("show: code %d, stderr %q", code, errs)
	}
	at := 0
	for _, want := range []string{"runner-1: 3 reads, oldest first\n", "command: hyp status\n", before, "command: hyp run " + c, ran, "command: hyp status\n", after} {
		i := strings.Index(shown[at:], want)
		if i < 0 {
			t.Fatalf("show lacks, in order, %q:\n%s", want, shown)
		}
		at += i + len(want)
	}
	// show is a read of its own actor, not of the actor it shows
	if n := len(reads(t, f.r.Dir, "runner-1")); n != 3 {
		t.Fatalf("show added a read to runner-1: %d", n)
	}
	own := reads(t, f.r.Dir, "owner-1")
	if last := own[len(own)-1]; last.Command[0] != "show" || stored(t, f.r.Dir, last.Stdout) != shown {
		t.Fatalf("show's own trace = %+v", last)
	}
}

func TestShowTraceHoldsFailures(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	code, out, errs := run(t, "run", "HEAD", "G-NOPE00/line-1")
	if code != 2 || out != "" || errs == "" {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	got := reads(t, f.r.Dir, "runner-1")
	if len(got) != 1 || got[0].Exit != 2 || stored(t, f.r.Dir, got[0].Stderr) != errs || stored(t, f.r.Dir, got[0].Stdout) != "" {
		t.Fatalf("reads = %+v", got)
	}
	code, shown, _ := run(t, "show", "runner-1")
	if code != 0 || !strings.Contains(shown, "exit 2\n") || !strings.Contains(shown, "stderr, "+strconv.Itoa(len(errs))+" bytes:\n"+errs) || !strings.Contains(shown, "stdout, 0 bytes") {
		t.Fatalf("show:\n%s", shown)
	}
}

func TestShowNoReads(t *testing.T) {
	newFx(t, map[string]string{"line-1/a": counterEx})
	code, out, _ := run(t, "show", "nobody-1")
	if code != 0 || out != "no read of nobody-1 is recorded\n" {
		t.Fatalf("code %d, stdout %q", code, out)
	}
}

func TestShowBadArgumentsExit2(t *testing.T) {
	newFx(t, map[string]string{"line-1/a": counterEx})
	for _, args := range [][]string{{"show"}, {"show", "a-1", "b-1"}, {"show", "Not An Actor"}} {
		if code, out, _ := run(t, args...); code != 2 || out != "" {
			t.Errorf("hyp %v: code %d, stdout %q", args, code, out)
		}
	}
}

// TestReadNeedsAnActor: a read prints only to an actor it can record; with no
// actor it is a bad argument and leaves no trace.
func TestReadNeedsAnActor(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	evs, _, err := logstore.Events(f.r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	t.Setenv("HYP_ACTOR", "")
	for _, args := range [][]string{{"status"}, {"show", "runner-1"}} {
		if code, out, errs := run(t, args...); code != 2 || out != "" || !strings.Contains(errs, "actor") {
			t.Errorf("hyp %v: code %d, stdout %q, stderr %q", args, code, out, errs)
		}
	}
	if code, _, _ := run(t, "status", "--actor", "Bad Actor"); code != 2 {
		t.Errorf("a bad actor: code %d", code)
	}
	after, _, _ := logstore.Events(f.r.Dir)
	if len(after) != len(evs) {
		t.Fatalf("%d events after reads with no actor, want %d", len(after), len(evs))
	}
	// a bad role after a valid actor is traced with its exit code
	t.Setenv("HYP_ACTOR", "owner-3")
	if code, _, _ := run(t, "status", "--role", "king"); code != 2 || len(reads(t, f.r.Dir, "owner-3")) != 1 || reads(t, f.r.Dir, "owner-3")[0].Exit != 2 {
		t.Fatalf("--role king: code %d, reads %+v", code, reads(t, f.r.Dir, "owner-3"))
	}
	// --actor beats the environment
	if code, _, _ := run(t, "status", "--actor", "owner-2"); code != 0 || len(reads(t, f.r.Dir, "owner-2")) != 1 || len(reads(t, f.r.Dir, "owner-3")) != 1 {
		t.Fatalf("--actor owner-2: code %d", code)
	}
}

// TestUnrecordedReadPrintsNothing: the trace is written before the print, so
// a read hyp cannot record prints nothing of its result and exits 3.
func TestUnrecordedReadPrintsNothing(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	outputs := trace(t, f.r.Dir).Outputs()
	if err := os.Chmod(outputs, 0o500); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.Chmod(outputs, 0o755) })
	code, out, errs := run(t, "status")
	if code != 3 || out != "" || !strings.Contains(errs, "could not record what this command printed") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
}

// TestValuesPrintAsHeld: run prints an example's words, and status prints an
// output's first line, as the store holds them: no quotes added, no escapes.
func TestValuesPrintAsHeld(t *testing.T) {
	words := `The owner says "hello" to hyp\n.`
	ex := "-- words --\n" + words + "\n-- run --\ncat counter\n-- floor: contains --\nzzz\n"
	f := newFx(t, map[string]string{"line-1/a": ex})
	c := f.leaf(t, map[string]string{"counter": `say "ok" \t` + "\n"})
	code, out, errs := run(t, "run", c, f.goal+"/line-1")
	if code != 0 || !strings.Contains(out, "\n  words: "+words+"\n") {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	// the outputs table prints at a stop: use the budget
	for _, v := range []string{"x2", "x3"} {
		if code, _, errs := run(t, "run", f.leaf(t, map[string]string{"counter": v + "\n"}), f.goal+"/line-1"); code != 0 {
			t.Fatalf("run %s: code %d, stderr %q", v, code, errs)
		}
	}
	_, status, _ := run(t, "status")
	if !strings.Contains(status, ` bytes: say "ok" \t`+"\n") {
		t.Fatalf("status does not print the output as held:\n%s", status)
	}
}

// TestShowAcrossWorktrees: a read in one git worktree of a clone shows from
// every other worktree of that clone, before any commit or merge.
func TestShowAcrossWorktrees(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	testutil.Git(t, f.r.Dir, "add", "-f", ".tree")
	testutil.Git(t, f.r.Dir, "commit", "-q", "-m", "the store")
	wt := filepath.Join(t.TempDir(), "wt")
	testutil.Git(t, f.r.Dir, "worktree", "add", "-q", "-b", "runner", wt)
	t.Chdir(wt)
	code, read, errs := run(t, "status")
	if code != 0 || !strings.Contains(read, f.goal) {
		t.Fatalf("status in the worktree: code %d, stdout %q, stderr %q", code, read, errs)
	}
	t.Chdir(f.r.Dir)
	t.Setenv("HYP_ACTOR", "owner-1")
	code, shown, _ := run(t, "show", "runner-1")
	if code != 0 || !strings.Contains(shown, "stdout, "+strconv.Itoa(len(read))+" bytes:\n"+read) {
		t.Fatalf("show in the main checkout: code %d:\n%s", code, shown)
	}
	// a separate clone has a trace of its own
	other := filepath.Join(t.TempDir(), "clone")
	testutil.Git(t, f.r.Dir, "clone", "-q", f.r.Dir, other)
	if trace(t, other).Dir == trace(t, f.r.Dir).Dir || trace(t, wt).Dir != trace(t, f.r.Dir).Dir {
		t.Fatal("the trace folder is not one per clone")
	}
}

// TestShowKeepsReadsInTheLog: a read event an earlier hyp wrote into the log
// still replays, and show prints it beside the trace's reads, oldest first.
func TestShowKeepsReadsInTheLog(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	old := "what an earlier hyp printed\n"
	hash, err := logstore.PutOutput(f.r.Dir, []byte(old))
	if err != nil {
		t.Fatal(err)
	}
	empty, err := logstore.PutOutput(f.r.Dir, nil)
	if err != nil {
		t.Fatal(err)
	}
	payload, _ := json.Marshal(events.Read{Role: "runner", Command: []string{"status"}, Stdout: hash, Stderr: empty})
	if _, err := logstore.Append(f.r.Dir, "runner-1", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeRead, Payload: payload}, nil
	}); err != nil {
		t.Fatal(err)
	}
	_, now, _ := run(t, "status")
	code, shown, _ := run(t, "show", "runner-1")
	at := strings.Index(shown, old)
	if code != 0 || at < 0 || !strings.Contains(shown[at:], now) || !strings.HasPrefix(shown, "runner-1: 2 reads") {
		t.Fatalf("show: code %d:\n%s", code, shown)
	}
}

// TestPruneListsThenRemovesGoneTraces: hyp prune names the trace of a
// repository that no longer exists, not a live one, and removes nothing;
// hyp prune --remove removes only that trace and leaves $HOME/.local alone.
func TestPruneListsThenRemovesGoneTraces(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	old := filepath.Join(home, ".local", "state", "hyp", "traces", "old")
	if err := os.MkdirAll(old, 0o755); err != nil {
		t.Fatal(err)
	}
	dead := inRepo(t)
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:g"); code != 0 {
		t.Fatalf("add goal: %s", errs)
	}
	deadTrace := trace(t, dead.Dir)
	live := inRepo(t)
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:g"); code != 0 {
		t.Fatalf("add goal: %s", errs)
	}
	liveTrace := trace(t, live.Dir)
	if code, out, errs := run(t, "prune"); code != 0 || !strings.HasPrefix(out, "no trace under "+filepath.Join(home, ".hyp", "traces")) {
		t.Fatalf("prune with nothing gone: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if err := os.RemoveAll(dead.Dir); err != nil {
		t.Fatal(err)
	}
	code, out, errs := run(t, "prune")
	if code != 0 || !strings.Contains(out, deadTrace.Clone) || strings.Contains(out, live.Dir) {
		t.Fatalf("prune: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, err := os.Stat(deadTrace.Dir); err != nil {
		t.Fatalf("prune removed a trace: %v", err)
	}
	code, out, errs = run(t, "prune", "--remove")
	if code != 0 || !strings.Contains(out, "removed 1 trace") || !strings.Contains(out, deadTrace.Clone) || strings.Contains(out, live.Dir) {
		t.Fatalf("prune --remove: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, err := os.Stat(deadTrace.Dir); !os.IsNotExist(err) {
		t.Fatalf("the gone repository's trace is still there: %v", err)
	}
	if _, err := os.Stat(liveTrace.Dir); err != nil {
		t.Fatalf("the live repository's trace is gone: %v", err)
	}
	if _, err := os.Stat(old); err != nil {
		t.Fatalf("the folder under ~/.local changed: %v", err)
	}
	if code, out, _ := run(t, "prune"); code != 0 || strings.Contains(out, deadTrace.Clone) {
		t.Fatalf("prune after --remove: code %d, stdout %q", code, out)
	}
	if code, _, errs := run(t, "prune", "extra"); code != 2 || errs == "" {
		t.Fatalf("prune extra: code %d, stderr %q", code, errs)
	}
}

// TestPruneNeedsAStore: like every command but add goal, prune in a
// repository with no store is refused, exit 1.
func TestPruneNeedsAStore(t *testing.T) {
	inRepo(t)
	if code, _, errs := run(t, "prune"); code != 1 || !strings.Contains(errs, "no store; run hyp add goal") {
		t.Fatalf("prune with no store: code %d, stderr %q", code, errs)
	}
}
