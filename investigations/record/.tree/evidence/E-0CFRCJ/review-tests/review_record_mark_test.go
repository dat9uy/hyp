package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The record mark outlives the record ref: after the ref is deleted, the
// repository is a git repository with no record ref, and hyp with git
// reads and writes the work-tree store; with no git on PATH it must do the
// same (as on the base), not exit 3.
func TestReviewStaleRecordMarkAfterRefDeleted(t *testing.T) {
	repo := recordRepo(t) // files ref backend; hyp init made the record
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "work tree", "--why", "w")
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "update-ref", "-d", "refs/hyp/record")
	t.Chdir(dst)
	want(t, hyp(t, "status"), 0, g) // with git: no record, work tree
	t.Setenv("PATH", t.TempDir())
	r := hyp(t, "status")
	if r.code != 0 {
		t.Errorf("no record ref, no git on PATH: hyp status exit %d, stderr %q; want 0 and the work-tree store", r.code, r.stderr)
	}
	r = hyp(t, "add", "goal", "--title", "written with no git", "--why", "w")
	if r.code != 0 {
		t.Errorf("no record ref, no git on PATH: hyp add exit %d, stderr %q; want 0", r.code, r.stderr)
	}
}

// A reftable clone whose record arrived through hyp sync carries no record
// mark, so with no git on PATH hyp takes it for a repository with no record
// and writes the work-tree store: the write exits 0 but the record, and so
// every other worktree, never sees the event.
func TestReviewSyncedReftableRecordHasNoMark(t *testing.T) {
	gitConfig(t)
	origin := recordRepo(t)
	must(t, "add", "goal", "--title", "in the record", "--why", "w")
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	clone := filepath.Join(base, "clone")
	if out, err := exec.Command("git", "clone", "-q", "--ref-format=reftable", origin, clone).CombinedOutput(); err != nil {
		t.Skipf("git cannot make a reftable clone: %v %s", err, out)
	}
	t.Chdir(filepath.Join(clone, "s"))
	want(t, hyp(t, "sync", "origin"), 0, "synced")
	if _, err := os.Stat(filepath.Join(clone, ".git", "hyp-record")); err != nil {
		t.Logf("after hyp sync the clone has no record mark: %v", err)
	}
	path := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	r := hyp(t, "status")
	if r.code != 3 {
		t.Errorf("reftable record repository, no git on PATH: hyp status exit %d, stdout %q stderr %q; want 3", r.code, r.stdout, r.stderr)
	}
	r = hyp(t, "init")
	if r.code != 3 {
		t.Errorf("reftable record repository, no git on PATH: hyp init exit %d, stderr %q; want 3", r.code, r.stderr)
	}
	r = hyp(t, "add", "goal", "--title", "written with no git", "--why", "w")
	if r.code != 3 {
		t.Errorf("reftable record repository, no git on PATH: hyp add exit %d, stdout %q; want 3", r.code, r.stdout)
	}
	g2 := r.stdout
	t.Setenv("PATH", path)
	if r.code == 0 && g2 != "" {
		wt := filepath.Join(base, "wt2")
		gitIn(t, clone, "worktree", "add", "-q", "--detach", wt)
		t.Chdir(filepath.Join(wt, "s"))
		if r := hyp(t, "status"); r.code == 0 && !strings.Contains(r.stdout, "written with no git") {
			t.Errorf("the write exited 0 with no git, but hyp status in another worktree does not print it:\n%s", r.stdout)
		}
	}
}
