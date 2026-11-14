package cli

import (
	"os"
	"path/filepath"
	"testing"
)

// A git repository with no record ref, in the reftable ref format, with no
// git on PATH: line 9 says hyp reads and writes the work-tree .tree/ as on
// the base, where hyp needed no git at all.
func TestProbeReftableRepoWithNoRecordAndNoGitUsesTheWorkTree(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q", "--ref-format=reftable")
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "work tree", "--why", "w")
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dst)
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 0, g)
	g2 := must(t, "add", "goal", "--title", "written with no git", "--why", "w")
	want(t, hyp(t, "show", g2), 0, "written with no git")
}

// Control: a record repository in the reftable format works with git on
// PATH (line 1), and with no git on PATH every command exits 3.
func TestProbeReftableRecordRepo(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q", "--ref-format=reftable")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "a store\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	g := must(t, "add", "goal", "--title", "in the record", "--why", "w")
	gitIn(t, repo, "rev-parse", "--verify", "refs/hyp/record")
	want(t, hyp(t, "status"), 0, g)
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 3, "cannot read the record")
}

// Control: in a linked worktree of a record repository whose refs are
// packed, no git on PATH gives exit 3, not the work-tree store.
func TestProbeLinkedWorktreePackedRefsNoGit(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "in the record", "--why", "w")
	gitIn(t, repo, "pack-refs", "--all")
	if _, err := os.Stat(filepath.Join(repo, ".git", "refs", "hyp", "record")); err == nil {
		t.Fatal("ref still loose")
	}
	wt := filepath.Join(t.TempDir(), "wt")
	gitIn(t, repo, "worktree", "add", "-q", "--detach", wt)
	if err := os.MkdirAll(filepath.Join(wt, "s"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(wt, "s"))
	want(t, hyp(t, "status"), 0, g)
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 3, "cannot read the record")
	want(t, hyp(t, "add", "goal", "--title", "x", "--why", "w"), 3, "cannot read the record")
}
