package logstore

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

// TestTraceIsOutsideTheRepository: the trace folder is under the state
// folder, one per clone, shared by its worktrees, and writing it changes no
// file of the repository and no git state.
func TestTraceIsOutsideTheRepository(t *testing.T) {
	state := t.TempDir()
	t.Setenv("XDG_STATE_HOME", state)
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	before := testutil.Snapshot(t, r.Dir)
	tr, err := TraceOf(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tr.Dir, filepath.Join(state, "hyp", "traces")+string(filepath.Separator)) {
		t.Fatalf("trace folder %s is not under %s", tr.Dir, state)
	}
	hash, err := tr.PutOutput([]byte("what hyp printed"))
	if err != nil {
		t.Fatal(err)
	}
	ev, err := tr.Append("runner-1", events.Event{Type: events.TypeRead, Payload: []byte(`{"role":"runner","command":["status"],"exit":0,"stdout":"` + hash + `","stderr":""}`)})
	if err != nil || ev.Actor != "runner-1" || !ValidID(ev.ID) {
		t.Fatalf("Append = %+v, %v", ev, err)
	}
	evs, errs, err := tr.Events()
	if err != nil || len(errs) != 0 || len(evs) != 1 || evs[0].ID != ev.ID {
		t.Fatalf("Events = %v, %v, %v", evs, errs, err)
	}
	if data, err := os.ReadFile(filepath.Join(tr.Outputs(), hash)); err != nil || string(data) != "what hyp printed" {
		t.Fatalf("output = %q, %v", data, err)
	}
	if clone, err := os.ReadFile(filepath.Join(tr.Dir, "clone")); err != nil || strings.TrimSpace(string(clone)) != tr.Clone {
		t.Fatalf("clone file = %q, %v", clone, err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("the trace changed the repository's git state")
	}
	if out := testutil.Git(t, r.Dir, "status", "--porcelain", "--untracked-files=all"); out != "" {
		t.Fatalf("git sees a change: %q", out)
	}
	wt := filepath.Join(t.TempDir(), "wt")
	testutil.Git(t, r.Dir, "worktree", "add", "-q", "-b", "other", wt)
	if other, err := TraceOf(wt); err != nil || other.Dir != tr.Dir {
		t.Fatalf("the worktree's trace = %+v, %v, want %s", other, err, tr.Dir)
	}
}

// TestTraceFallsBackToHome: with no absolute $XDG_STATE_HOME the trace is
// under $HOME/.local/state; with neither, it is an environment problem.
func TestTraceFallsBackToHome(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_STATE_HOME", "relative")
	tr, err := TraceOf(r.Dir)
	if err != nil || !strings.HasPrefix(tr.Dir, filepath.Join(home, ".local", "state", "hyp", "traces")) {
		t.Fatalf("TraceOf = %+v, %v", tr, err)
	}
	t.Setenv("HOME", "")
	if _, err := TraceOf(r.Dir); fault.KindOf(err) != fault.Environment {
		t.Fatalf("no state folder: err = %v, want an environment problem", err)
	}
}
