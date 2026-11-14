package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// failingGit puts on PATH a git that fails every command as git does when
// it cannot open the repository.
func failingGit(t *testing.T) {
	t.Helper()
	bin := t.TempDir()
	writeFile(t, filepath.Join(bin, "git"), "#!/bin/sh\necho 'fatal: not a git repository' >&2\nexit 128\n")
	if err := os.Chmod(filepath.Join(bin, "git"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Setenv("PATH", bin)
}

// In a record repository, hyp needs git to read the record: with no git on
// PATH, or a git that cannot open the repository, hyp status, hyp show and
// a write exit 3 and write nothing to the work-tree store, so no event a
// write exits 0 on is missing from the record.
func TestRecordNeedsGitToReadTheRecord(t *testing.T) {
	recordRepo(t)
	g := must(t, "add", "goal", "--title", "in the record", "--why", "w")
	path := os.Getenv("PATH")
	for _, c := range []struct {
		name string
		set  func()
	}{
		{"no git on PATH", func() { t.Setenv("PATH", t.TempDir()) }},
		{"git cannot open the repository", func() { failingGit(t) }},
	} {
		c.set()
		want(t, hyp(t, "status"), 3, "cannot read the record")
		want(t, hyp(t, "show", g), 3, "cannot read the record")
		want(t, hyp(t, "add", "goal", "--title", "written with no record", "--why", "w"), 3, "cannot read the record")
		t.Setenv("PATH", path)
		out := hyp(t, "status").stdout
		if !strings.Contains(out, g) || strings.Contains(out, "written with no record") {
			t.Fatalf("%s: with git back hyp status prints:\n%s", c.name, out)
		}
		if _, err := os.Stat(filepath.Join(".tree", "events")); err == nil {
			entries, _ := os.ReadDir(filepath.Join(".tree", "events"))
			for _, e := range entries {
				if b, _ := os.ReadFile(filepath.Join(".tree", "events", e.Name())); strings.Contains(string(b), "written with no record") {
					t.Fatalf("%s: the refused write reached the work-tree store %s", c.name, e.Name())
				}
			}
		}
		must(t, "add", "goal", "--title", "next", "--why", "w")
	}
}

// In a git repository with no record ref, hyp with no git on PATH reads and
// writes the work-tree store as before the record.
func TestNoGitInARepositoryWithNoRecordUsesTheWorkTree(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
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

// hyp import untracks only the evidence files no event names: an uncited
// file whose name is a glob that matches a cited file leaves the cited file
// tracked.
func TestImportKeepsCitedFileTrackedBesideAGlobName(t *testing.T) {
	repo := recordRepo(t)
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	_, _, ptr := chain(t, "cited\n")
	dir := filepath.Dir(ptr)
	for _, n := range []string{"ou[t].txt", "*.txt", "o?t.txt"} {
		writeFile(t, filepath.Join(other, dir, n), "uncited\n")
	}
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(filepath.Join(dst, ".tree", ".gitignore")); err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "add", "-A", "w")
	gitIn(t, repo, "commit", "-qm", "work-tree store")
	t.Chdir(dst)
	want(t, hyp(t, "import", "."), 0, "imported")
	tracked := strings.Split(gitIn(t, repo, "ls-files", "w/.tree/evidence"), "\n")
	if len(tracked) != 1 || tracked[0] != "w/"+ptr {
		t.Fatalf("hyp import left tracked %q; want only the cited w/%s", tracked, ptr)
	}
	for _, n := range []string{"ou[t].txt", "*.txt", "o?t.txt"} {
		if _, err := os.Stat(filepath.Join(dir, n)); err != nil {
			t.Fatalf("hyp import removed the uncited %s from disk: %v", n, err)
		}
	}
}

// hyp show PATH@SHA prints only bytes whose sha256 starts SHA: when the
// record keeps no such copy, it refuses.
func TestShowPickPrintsOnlyBytesOfThatSha(t *testing.T) {
	repo := recordRepo(t)
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	_, h, _, ptr := chainH(t, "one\n")
	writeFile(t, ptr, "two\n")
	must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--rubric", "reading", "--records", ptr, "--why", "w")
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	_ = os.Remove(filepath.Join(dst, ".tree", ".gitignore"))
	gitIn(t, repo, "add", "-A", "w")
	gitIn(t, repo, "commit", "-qm", "work-tree store")
	t.Chdir(dst)
	want(t, hyp(t, "import", "."), 0, "imported")
	printed := 0
	for _, data := range []string{"one\n", "two\n"} {
		r := hyp(t, "show", ptr+"@"+sha256Hex(data)[:12])
		if r.code == 0 {
			printed++
		}
		if r.code == 0 && r.stdout != data {
			t.Fatalf("hyp show %s@%s printed %q, not %q", ptr, sha256Hex(data)[:12], r.stdout, data)
		}
		if r.code != 0 && r.code != 1 {
			t.Fatalf("hyp show %s@%s: exit %d, stderr %q", ptr, sha256Hex(data)[:12], r.code, r.stderr)
		}
	}
	if printed == 0 {
		t.Fatal("hyp show refused both picks; the record keeps one of the two contents")
	}
}

// reftableRepo makes a git repository in the reftable ref format, or skips
// when the git on PATH cannot.
func reftableRepo(t *testing.T) string {
	t.Helper()
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", repo, "init", "-q", "--ref-format=reftable").CombinedOutput(); err != nil {
		t.Skipf("git cannot make a reftable repository: %v %s", err, out)
	}
	return repo
}

// A reftable repository with no record ref and no git on PATH reads and
// writes the work-tree store as before the record; a reftable record
// repository with no git on PATH exits 3.
func TestReftableRepositoryWithNoGit(t *testing.T) {
	repo := reftableRepo(t)
	other := t.TempDir()
	t.Chdir(other)
	must(t, "init")
	g := must(t, "add", "goal", "--title", "work tree", "--why", "w")
	dst := filepath.Join(repo, "w")
	if err := os.CopyFS(dst, os.DirFS(other)); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dst)
	path := os.Getenv("PATH")
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 0, g)
	g2 := must(t, "add", "goal", "--title", "written with no git", "--why", "w")
	want(t, hyp(t, "show", g2), 0, "written with no git")
	t.Setenv("PATH", path)

	rec := reftableRepo(t)
	writeFile(t, filepath.Join(rec, "s", "README.md"), "a store\n")
	gitIn(t, rec, "add", "-A")
	gitIn(t, rec, "commit", "-qm", "base")
	t.Chdir(filepath.Join(rec, "s"))
	must(t, "init")
	g3 := must(t, "add", "goal", "--title", "in the record", "--why", "w")
	want(t, hyp(t, "status"), 0, g3)
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 3, "cannot read the record")
	want(t, hyp(t, "add", "goal", "--title", "x", "--why", "w"), 3, "cannot read the record")
}
