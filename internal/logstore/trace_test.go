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

// TestTraceIsOutsideTheRepository: the trace folder is under $HOME/.hyp, one
// per clone, shared by its worktrees, and writing it changes no file of the
// repository and no git state.
func TestTraceIsOutsideTheRepository(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	state := filepath.Join(home, ".hyp")
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	before := testutil.Snapshot(t, r.Dir)
	tr, err := TraceOf(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(tr.Dir, filepath.Join(state, "traces")+string(filepath.Separator)) {
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

// TestTraceIgnoresXDGStateHome: the trace is under $HOME/.hyp whatever
// $XDG_STATE_HOME holds, and writing it leaves nothing under $HOME/.local or
// $XDG_STATE_HOME. With no absolute $HOME it is an environment problem.
func TestTraceIgnoresXDGStateHome(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	home := t.TempDir()
	t.Setenv("HOME", home)
	for _, state := range []string{"", "relative", t.TempDir()} {
		t.Setenv("XDG_STATE_HOME", state)
		tr, err := TraceOf(r.Dir)
		if err != nil || !strings.HasPrefix(tr.Dir, filepath.Join(home, ".hyp", "traces")+string(filepath.Separator)) {
			t.Fatalf("XDG_STATE_HOME=%q: TraceOf = %+v, %v", state, tr, err)
		}
		if _, err := tr.PutOutput([]byte("x")); err != nil {
			t.Fatal(err)
		}
		if filepath.IsAbs(state) {
			if entries, _ := os.ReadDir(state); len(entries) != 0 {
				t.Fatalf("$XDG_STATE_HOME holds %v", entries)
			}
		}
	}
	if _, err := os.Stat(filepath.Join(home, ".local")); !os.IsNotExist(err) {
		t.Fatalf("$HOME/.local: %v, want none", err)
	}
	for _, h := range []string{"", "relative"} {
		t.Setenv("HOME", h)
		if _, err := TraceOf(r.Dir); fault.KindOf(err) != fault.Environment {
			t.Fatalf("HOME=%q: err = %v, want an environment problem", h, err)
		}
	}
}

// TestGoneTraces: only a trace whose clone is gone is listed, and removing
// removes only those; a folder with no usable file clone is never listed.
func TestGoneTraces(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	live := testutil.NewRepo(t)
	live.Commit(t, map[string]string{"a": "a"})
	dead := testutil.NewRepo(t)
	dead.Commit(t, map[string]string{"a": "b"})
	var trs []Trace
	for _, r := range []*testutil.Repo{live, dead} {
		tr, err := TraceOf(r.Dir)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := tr.Append("runner-1", events.Event{Type: events.TypeRead, Payload: []byte(`{"role":"","command":["status"],"exit":0,"stdout":"","stderr":""}`)}); err != nil {
			t.Fatal(err)
		}
		trs = append(trs, tr)
	}
	base, err := TracesDir()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(base, "nameless"), 0o755); err != nil {
		t.Fatal(err)
	}
	if _, gone, err := GoneTraces(); err != nil || len(gone) != 0 {
		t.Fatalf("both live: %v, %v", gone, err)
	}
	if err := os.RemoveAll(dead.Dir); err != nil {
		t.Fatal(err)
	}
	b, gone, err := GoneTraces()
	if err != nil || b != base || len(gone) != 1 || gone[0].Clone != trs[1].Clone || gone[0].Key != filepath.Base(trs[1].Dir) {
		t.Fatalf("GoneTraces = %s, %v, %v; want the dead clone %s", b, gone, err, trs[1].Clone)
	}
	if _, err := os.Stat(trs[1].Dir); err != nil {
		t.Fatalf("listing removed the trace: %v", err)
	}
	removed, err := RemoveGone(base, gone)
	if err != nil || len(removed) != 1 {
		t.Fatalf("RemoveGone = %v, %v", removed, err)
	}
	if _, err := os.Stat(trs[1].Dir); !os.IsNotExist(err) {
		t.Fatalf("the dead clone's trace is still there: %v", err)
	}
	for _, d := range []string{trs[0].Dir, filepath.Join(base, "nameless")} {
		if _, err := os.Stat(d); err != nil {
			t.Fatalf("%s was removed: %v", d, err)
		}
	}
	// RemoveGone checks each clone again: one that exists keeps its trace
	if err := os.MkdirAll(trs[0].Dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if removed, err := RemoveGone(base, []Gone{{Key: filepath.Base(trs[0].Dir), Clone: trs[0].Clone}}); err != nil || len(removed) != 0 {
		t.Fatalf("RemoveGone of a live clone = %v, %v", removed, err)
	}
	if _, err := os.Stat(trs[0].Dir); err != nil {
		t.Fatalf("the live clone's trace was removed: %v", err)
	}
}
