package acts

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

// tryStateOf is what a command could change in the repository: git's
// directory, the files git sees as changed or new, and every file under
// .tree.
func tryStateOf(t *testing.T, root string) string {
	t.Helper()
	h := sha256.New()
	fmt.Fprintln(h, testutil.Snapshot(t, root))
	fmt.Fprintln(h, testutil.Git(t, root, "status", "--porcelain", "--untracked-files=all"))
	err := filepath.WalkDir(filepath.Join(root, ".tree"), func(p string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		if err != nil {
			return err
		}
		fmt.Fprintf(h, "%s %x\n", p, sha256.Sum256(b))
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}

func (e *runEnv) try(key string, worktree bool) (TryReport, error) {
	return Try(TryOpts{Dir: e.root, Goal: e.goal, Key: key, Worktree: worktree, ProbeTimeout: 20 * time.Second})
}

// h0Cell returns H0's cell of the latest placement of key.
func (e *runEnv) h0Cell(t *testing.T, key string) events.Cell {
	t.Helper()
	var c events.Cell
	found := false
	for _, ev := range e.placements(t) {
		var p events.ExamplePlace
		if err := json.Unmarshal(ev.Payload, &p); err != nil {
			t.Fatal(err)
		}
		if p.Line+"/"+p.Example == key {
			c, found = p.H0, true
		}
	}
	if !found {
		t.Fatalf("no placement of %s", key)
	}
	return c
}

// An example folder nobody committed runs on H0 while the checkout holds an
// edit nobody committed: try shows H0's output, stderr, exit code and
// verdict, changes nothing, and shows what place then records.
func TestTryOnH0ShowsWhatPlaceRecords(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	e := newRunEnv(t)
	ex := "-- words --\nw\n-- run --\ncat counter; echo warn >&2; exit 3\n-- floor: contains --\nok\n"
	e.write(t, "line-1/b", ex, fixtureProbe)
	writeFile(t, e.root, "counter", "ok\n")

	before := tryStateOf(t, e.root)
	rep, err := e.try("line-1/b", false)
	if err != nil {
		t.Fatal(err)
	}
	if after := tryStateOf(t, e.root); after != before {
		t.Fatal("try changed the repository")
	}
	if string(rep.Out.Stdout) != "bad" || string(rep.Out.Stderr) != "warn\n" || rep.Out.Exit != 3 || rep.Verdict != "failed" || rep.Commit != e.base || rep.Worktree {
		t.Fatalf("report = %+v, want H0's output bad, stderr warn, exit 3, failed", rep)
	}

	testutil.Git(t, e.root, "add", "goals")
	testutil.Git(t, e.root, "commit", "-q", "-m", "example b")
	if _, err := e.place("line-1/b"); err != nil {
		t.Fatal(err)
	}
	c := e.h0Cell(t, "line-1/b")
	if c.Output != rep.OutHash || c.Stderr != rep.ErrHash || c.Exit != rep.Out.Exit || c.Verdict != rep.Verdict || c.Probe != rep.ProbeHash {
		t.Fatalf("place recorded %+v; try showed %+v", c, rep)
	}
	if cur, ok := e.goalRow(t).Current("line-1/b"); !ok || cur.Hash != rep.Hash {
		t.Fatalf("placed %+v, try showed hash %s", cur, rep.Hash)
	}
}

// With worktree, try runs on the checkout's files: an edited file, a new
// file, not an ignored file, not a deleted one. It changes nothing and uses
// no run, and it shows what run records once the change is committed.
func TestTryWorktreeShowsWhatRunRecords(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	e := newRunEnv(t)
	ex := "-- words --\nw\n-- run --\ncat counter extra ignored gone 2>/dev/null; echo warn >&2\n-- floor: contains --\nok\n"
	e.write(t, "line-1/c", ex, fixtureProbe)
	e.commit(t)
	if _, err := e.place("line-1/c"); err != nil {
		t.Fatal(err)
	}
	e.r.Commit(t, map[string]string{".gitignore": "ignored\n", "gone": "gone\n"})
	writeFile(t, e.root, "counter", "ok\n")
	writeFile(t, e.root, "extra", "and-more\n")
	writeFile(t, e.root, "ignored", "ignored\n")
	if err := os.Remove(filepath.Join(e.root, "gone")); err != nil {
		t.Fatal(err)
	}

	before := tryStateOf(t, e.root)
	rep, err := e.try("line-1/c", true)
	if err != nil {
		t.Fatal(err)
	}
	if after := tryStateOf(t, e.root); after != before {
		t.Fatal("try --worktree changed the repository")
	}
	if string(rep.Out.Stdout) != "ok\nand-more" || rep.Verdict != "met" || !rep.Worktree {
		t.Fatalf("report = %+v, want the edited and the new file only, met", rep)
	}
	if n := len(e.results(t)); n != 0 {
		t.Fatalf("try recorded %d Results", n)
	}

	testutil.Git(t, e.root, "add", "-A")
	testutil.Git(t, e.root, "commit", "-q", "-m", "the change")
	head := testutil.Git(t, e.root, "rev-parse", "HEAD")
	run, err := e.run(head[:len(head)-1])
	if err != nil {
		t.Fatal(err)
	}
	res := e.results(t)
	var cell events.ExampleCell
	for _, c := range res[len(res)-1].Cells {
		if c.Example == "c" {
			cell = c
		}
	}
	if cell.Output != rep.OutHash || cell.Stderr != rep.ErrHash || cell.Exit != rep.Out.Exit || cell.Verdict != rep.Verdict || cell.Hash != rep.Hash || cell.Probe != rep.ProbeHash {
		t.Fatalf("run recorded %+v (report %+v); try showed %+v", cell, run, rep)
	}
}

func TestTryBadArguments(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	e := newRunEnv(t)
	for _, key := range []string{"line-1", "line-1/A", "line-9/a", "line-1/nothere"} {
		_, err := e.try(key, false)
		if fault.KindOf(err) != fault.BadArguments {
			t.Errorf("try %s: err = %v, want bad arguments", key, err)
		}
	}
	if _, err := Try(TryOpts{Dir: e.root, Goal: "G-NOPE00", Key: "line-1/a"}); fault.KindOf(err) != fault.BadArguments {
		t.Errorf("unknown Goal: err = %v, want bad arguments", err)
	}
}

// A Probe that prints no exit line is an environment problem, as in place.
func TestTryProbeFailureIsEnvironment(t *testing.T) {
	t.Setenv("HOME", t.TempDir())
	e := newRunEnv(t)
	e.write(t, "line-1/d", metOnH0, "#!/bin/sh\necho no-exit-line\n")
	if _, err := e.try("line-1/d", false); fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
}

func writeFile(t *testing.T, root, rel, body string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(root, rel), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}
