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

// After the record ref is deleted, the repository is a git repository with
// no record ref: with no git on PATH hyp reads and writes the work-tree
// store, as it does with git.
func TestRecordRefDeletedWithNoGitUsesTheWorkTree(t *testing.T) {
	repo := recordRepo(t)
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
	want(t, hyp(t, "status"), 0, g)
	t.Setenv("PATH", t.TempDir())
	want(t, hyp(t, "status"), 0, g)
	g2 := must(t, "add", "goal", "--title", "written with no git", "--why", "w")
	want(t, hyp(t, "show", g2), 0, "written with no git")
}

// A reftable clone whose record came by hyp sync, or by a plain git fetch,
// is a record repository: with no git on PATH hyp status, hyp init and a
// write exit 3, and no event reaches only the work tree.
func TestSyncedReftableRecordNeedsGit(t *testing.T) {
	origin := recordRepo(t)
	must(t, "add", "goal", "--title", "in the record", "--why", "w")
	base, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	for _, how := range []string{"hyp sync", "git fetch"} {
		clone := filepath.Join(base, strings.ReplaceAll(how, " ", "-"))
		if out, err := exec.Command("git", "clone", "-q", "--ref-format=reftable", origin, clone).CombinedOutput(); err != nil {
			t.Skipf("git cannot make a reftable clone: %v %s", err, out)
		}
		t.Chdir(filepath.Join(clone, "s"))
		if how == "hyp sync" {
			want(t, hyp(t, "sync", "origin"), 0, "synced")
		} else {
			gitIn(t, clone, "fetch", "-q", "origin", "refs/hyp/record:refs/hyp/record")
		}
		path := os.Getenv("PATH")
		t.Setenv("PATH", t.TempDir())
		want(t, hyp(t, "status"), 3, "cannot read the record")
		want(t, hyp(t, "init"), 3, "cannot read the record")
		want(t, hyp(t, "add", "goal", "--title", "written with no git", "--why", "w"), 3, "cannot read the record")
		t.Setenv("PATH", path)
		if out := hyp(t, "status").stdout; !strings.Contains(out, "in the record") || strings.Contains(out, "written with no git") {
			t.Fatalf("%s: with git back hyp status prints:\n%s", how, out)
		}
	}
}

// Two stores whose names share a prefix of their sha256 each read their own
// events: s2330 and s130636 share the first 4 bytes.
func TestStoresWithACommonNameHashPrefixReadTheirOwnEvents(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	names := []string{"s2330", "s130636"}
	for _, n := range names {
		writeFile(t, filepath.Join(repo, n, "README.md"), n+"\n")
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	goals := map[string]string{}
	for _, n := range names {
		t.Chdir(filepath.Join(repo, n))
		must(t, "init")
		goals[n] = must(t, "add", "goal", "--title", "goal of "+n, "--why", "w")
	}
	for _, n := range names {
		t.Chdir(filepath.Join(repo, n))
		want(t, hyp(t, "show", goals[n]), 0, "goal of "+n)
		want(t, hyp(t, "status", goals[n]), 0, "goal of "+n)
	}
}

// A store folder reached through a symbolic link inside the repository is
// the store it links to: hyp show prints its events, and a write adds to it,
// not to a second store.
func TestStoreFolderThroughASymlinkIsTheStoreItLinksTo(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "the goal", "--why", "w")
	if err := os.Symlink("s", filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(repo, "link"))
	want(t, hyp(t, "show", g), 0, "the goal")
	g2 := must(t, "add", "goal", "--title", "second", "--why", "w")
	if names := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); strings.Contains(names, "link/") {
		t.Fatalf("a write through the link started a second store in the record:\n%s", names)
	}
	t.Chdir(filepath.Join(repo, "s"))
	want(t, hyp(t, "show", g2), 0, "second")
}

// Line 8: a store whose folder name holds a colon is imported by PATH in
// every spelling: an argument that names a folder on disk is a PATH.
func TestImportStoreFolderWithAColon(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "a:b", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "a:b"))
	must(t, "init")
	dropRecord(t, repo)
	g := must(t, "add", "goal", "--title", "colon goal", "--why", "w")
	t.Chdir(repo)
	want(t, hyp(t, "import", "a:b"), 0, "imported")
	for _, arg := range []string{"a:b", "./a:b", filepath.Join(repo, "a:b")} {
		want(t, hyp(t, "import", arg), 0, "holds every event")
	}
	t.Chdir(filepath.Join(repo, "a:b"))
	want(t, hyp(t, "show", g), 0, "colon goal")
}

// Line 6: hyp import untracks an evidence file no event names, leaving it
// on disk, also when its staged bytes differ from both HEAD and the work
// tree, and a second import exits 0.
func TestImportUntracksAnUnnamedFileWithStagedChanges(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	dropRecord(t, repo)
	chain(t, "cited\n")
	stray := filepath.Join(repo, "s", ".tree", "evidence", "stray.txt")
	writeFile(t, stray, "one\n")
	gitIn(t, repo, "add", "-f", "-A")
	gitIn(t, repo, "commit", "-qm", "stores")
	writeFile(t, stray, "two\n")
	gitIn(t, repo, "add", "-f", "s/.tree/evidence/stray.txt")
	writeFile(t, stray, "three\n")
	t.Chdir(repo)
	want(t, hyp(t, "import", "s"), 0, "imported")
	want(t, hyp(t, "import", "s"), 0, "holds every event")
	if out := gitIn(t, repo, "ls-files", "s/.tree/evidence/stray.txt"); out != "" {
		t.Fatalf("after hyp import, git still tracks %s, which no event names", out)
	}
	if b, err := os.ReadFile(stray); err != nil || string(b) != "three\n" {
		t.Fatalf("hyp import changed the unnamed file on disk: %q %v", b, err)
	}
}

// Line 8: hyp import adds each named file with the bytes its event names.
// A file whose bytes on disk changed after the event, so that they no
// longer hash as the event stores, is left out and printed, never kept.
func TestImportLeavesOutBytesNoEventNames(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(top, "s"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(top, "s"))
	must(t, "init")
	_, _, ptr := chain(t, "the bytes the Result cited\n")
	gitIn(t, top, "init", "-q")
	gitIn(t, top, "add", "-A")
	gitIn(t, top, "commit", "-qm", "store")
	writeFile(t, ptr, "changed after the Result\n")
	res := hyp(t, "import", ".")
	want(t, res, 0, "imported")
	if !strings.Contains(res.stderr, ptr) {
		t.Fatalf("hyp import kept or dropped %s with no warning: stderr %q", ptr, res.stderr)
	}
	if out, _ := exec.Command("git", "-C", top, "show", "refs/hyp/record:s/"+ptr).CombinedOutput(); strings.Contains(string(out), "changed after the Result") {
		t.Fatalf("the record keeps bytes no event names at s/%s", ptr)
	}
}

// Line 1: on a filesystem that folds case, the folder s typed as S is the
// same folder, so hyp status there prints what s prints, and a write from
// there goes to the store s, which then prints it.
func TestStoreFolderEnteredWithOtherCase(t *testing.T) {
	repo := recordRepo(t)
	probe := filepath.Join(repo, "probe-lower")
	writeFile(t, probe, "x")
	_, err := os.Stat(filepath.Join(repo, "PROBE-LOWER"))
	os.Remove(probe)
	if err != nil {
		t.Skip("the filesystem does not fold case")
	}
	g, r, ptr := chain(t, "the bytes\n")
	views0 := views(t, g, r, ptr)
	t.Chdir(filepath.Join(repo, "S"))
	if res := hyp(t, "status", g); res.code != 0 || res.stdout != views0[0] {
		t.Fatalf("hyp status %s from S: exit %d, stdout %q, stderr %q; want what s prints:\n%s", g, res.code, res.stdout, res.stderr, views0[0])
	}
	id := must(t, "add", "goal", "--title", "from S", "--why", "w")
	t.Chdir(filepath.Join(repo, "s"))
	want(t, hyp(t, "show", id), 0, "from S")
	if ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); strings.Contains(ls, "S/.tree/") {
		t.Fatalf("the record holds two stores for the folder s:\n%s", ls)
	}
}

// Line 1: on a filesystem that folds Unicode normalization, the folder
// café typed with a decomposed é is the same folder as the composed name it
// was made with, so hyp status there prints what that folder prints, and a
// write from there goes to the same store.
func TestStoreFolderEnteredWithOtherNormalization(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	nfc, nfd := "café", "café"
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, nfc, "README.md"), "a store\n")
	if _, err := os.Stat(filepath.Join(repo, nfd)); err != nil {
		t.Skip("the filesystem does not fold Unicode normalization")
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, nfc))
	must(t, "init")
	g, r, ptr := chain(t, "the bytes\n")
	views0 := views(t, g, r, ptr)
	t.Chdir(filepath.Join(repo, nfd))
	if res := hyp(t, "status", g); res.code != 0 || res.stdout != views0[0] {
		t.Fatalf("hyp status %s from the decomposed spelling: exit %d, stdout %q, stderr %q; want what the folder prints:\n%s", g, res.code, res.stdout, res.stderr, views0[0])
	}
	id := must(t, "add", "goal", "--title", "from NFD", "--why", "w")
	t.Chdir(filepath.Join(repo, nfc))
	want(t, hyp(t, "show", id), 0, "from NFD")
	stores := map[string]bool{}
	for _, l := range strings.Split(gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"), "\n") {
		if i := strings.Index(l, "/.tree/events/"); i >= 0 {
			stores[l[:i]] = true
		}
	}
	if len(stores) != 1 {
		t.Fatalf("the record holds %d stores for one folder: %v", len(stores), stores)
	}
}

// Line 8: an import that added an event without a named file's bytes,
// because its source held other bytes or none, is followed by an import
// from a commit that holds those bytes: it adds them to the event the
// record holds, and hyp show prints them. Before, hyp show says the store
// holds no copy and names hyp import.
func TestImportAddsNamedBytesAnEarlierImportLacked(t *testing.T) {
	for _, how := range []string{"changed", "removed"} {
		t.Run(how, func(t *testing.T) {
			gitConfig(t)
			top, err := filepath.EvalSymlinks(t.TempDir())
			if err != nil {
				t.Fatal(err)
			}
			if err := os.MkdirAll(filepath.Join(top, "s"), 0o755); err != nil {
				t.Fatal(err)
			}
			t.Chdir(filepath.Join(top, "s"))
			must(t, "init")
			_, _, ptr := chain(t, "the bytes the Result cited\n")
			gitIn(t, top, "init", "-q")
			// hyp init makes git ignore the evidence folder, so the commit
			// holds the cited file only when it is added with -f.
			gitIn(t, top, "add", "-f", "-A")
			gitIn(t, top, "commit", "-qm", "store")
			if how == "changed" {
				writeFile(t, ptr, "changed after the Result\n")
			} else if err := os.Remove(ptr); err != nil {
				t.Fatal(err)
			}
			want(t, hyp(t, "import", "."), 0, "imported")
			want(t, hyp(t, "show", ptr), 3, "hyp import")
			want(t, hyp(t, "import", "HEAD:."), 0, "added 1 files it lacked")
			want(t, hyp(t, "show", ptr), 0, "the bytes the Result cited")
			tip := gitIn(t, top, "rev-parse", "refs/hyp/record")
			want(t, hyp(t, "import", "HEAD:."), 0, "holds every event\n")
			if again := gitIn(t, top, "rev-parse", "refs/hyp/record"); again != tip {
				t.Fatalf("a third import moved the record from %s to %s", tip, again)
			}
		})
	}
}

// Line 1: on a filesystem that folds case, the folder s typed as S is the
// store s, also when the repository holds a link a to s, whose name sorts
// before s: a link to the folder is never its name.
func TestStoreFolderOtherCaseBesideALink(t *testing.T) {
	repo := recordRepo(t)
	probe := filepath.Join(repo, "probe-lower")
	writeFile(t, probe, "x")
	_, err := os.Stat(filepath.Join(repo, "PROBE-LOWER"))
	os.Remove(probe)
	if err != nil {
		t.Skip("the filesystem does not fold case")
	}
	if err := os.Symlink("s", filepath.Join(repo, "a")); err != nil {
		t.Fatal(err)
	}
	g, r, ptr := chain(t, "the bytes\n")
	views0 := views(t, g, r, ptr)
	t.Chdir(filepath.Join(repo, "S"))
	if res := hyp(t, "status", g); res.code != 0 || res.stdout != views0[0] {
		t.Fatalf("hyp status %s from S: exit %d, stdout %q, stderr %q; want what s prints:\n%s", g, res.code, res.stdout, res.stderr, views0[0])
	}
	id := must(t, "add", "goal", "--title", "from S", "--why", "w")
	t.Chdir(filepath.Join(repo, "s"))
	want(t, hyp(t, "show", id), 0, "from S")
	if ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); strings.Contains(ls, "a/.tree/") || strings.Contains(ls, "S/.tree/") {
		t.Fatalf("the record holds a second store for the folder s:\n%s", ls)
	}
}

// Line 7: an import that added an event without the bundle it pins, because
// its source lacked the bundle, is followed by an import from a commit that
// holds the bundle with the pinned hash: it adds the bundle, so the Result
// reads as the work-tree store read it from a checkout with or without the
// bundle; one more import adds nothing.
func TestImportAddsAPinnedBundleAnEarlierImportLacked(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, "s", "README.md"), "r\n")
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, "s"))
	must(t, "init")
	dropRecord(t, repo)
	bundleIn(t, filepath.Join(repo, "s"), "kit", "", nil)
	_, r := recipeChain(t)
	control := hyp(t, "show", r).stdout
	if strings.Contains(control, "questionable") {
		t.Fatalf("control: the work-tree store reads %s questionable:\n%s", r, control)
	}
	gitIn(t, repo, "add", "-f", "-A")
	gitIn(t, repo, "commit", "-qm", "store")
	if err := os.RemoveAll(filepath.Join(repo, "s", "recipes")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(repo)
	want(t, hyp(t, "import", "s"), 0, "the record lacks it")
	want(t, hyp(t, "import", "HEAD:s"), 0, "files it lacked")
	if ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); !strings.Contains(ls, "s/recipes/kit@1/recipe.md") {
		t.Fatalf("after an import from a commit that holds kit@1 with its pinned hash, the record lacks it:\n%s", ls)
	}
	tip := gitIn(t, repo, "rev-parse", "refs/hyp/record")
	want(t, hyp(t, "import", "HEAD:s"), 0, "holds every event\n")
	if again := gitIn(t, repo, "rev-parse", "refs/hyp/record"); again != tip {
		t.Fatalf("one more import moved the record from %s to %s", tip, again)
	}
	t.Chdir(filepath.Join(repo, "s"))
	without := hyp(t, "show", r).stdout
	gitIn(t, repo, "checkout", "--", "s/recipes")
	with := hyp(t, "show", r).stdout
	if strings.Contains(without, "questionable") || strings.Contains(with, "questionable") {
		t.Fatalf("%s reads questionable in the record, though the work-tree store read it fine:\nwithout the bundle:\n%s\nwith it:\n%s", r, without, with)
	}
}

// Where no git repository holds the current directory, hyp show of a cited
// file removed from the work tree reports it as on the base and names no
// hyp import, which only a record repository can use.
func TestShowMissingFileWithNoGitNamesNoImport(t *testing.T) {
	gitConfig(t)
	t.Chdir(t.TempDir())
	must(t, "init")
	_, _, ptr := chain(t, "bytes\n")
	if err := os.Remove(ptr); err != nil {
		t.Fatal(err)
	}
	res := hyp(t, "show", ptr)
	if res.code != 3 || strings.Contains(res.stderr, "hyp import") {
		t.Fatalf("with no git, hyp show of a removed cited file: exit %d, stderr %q; want exit 3 naming no hyp import", res.code, res.stderr)
	}
}
