package cli

import (
	"crypto/sha256"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Finding A (G-SZCBXZ#1). The snapshot of a store is keyed by the record's
// tip and only the first 4 bytes of sha256(store name). Two store names
// whose sha256 share those 4 bytes (s2330 and s130636, found by a birthday
// search over "s<N>") read one snapshot, which was built for the other
// store and so leaves out this store's folder: the event the write exited 0
// on is not printed by hyp show / hyp status in its own store's folder.
func TestReviewSnapshotKeyCollision(t *testing.T) {
	a, b := "s2330", "s130636"
	ha, hb := sha256.Sum256([]byte(a)), sha256.Sum256([]byte(b))
	if string(ha[:4]) != string(hb[:4]) {
		t.Fatalf("precondition: %s and %s must share 4 sha256 bytes", a, b)
	}
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	for _, n := range []string{a, b} {
		writeFile(t, filepath.Join(repo, n, "README.md"), n+"\n")
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	goals := map[string]string{}
	for _, n := range []string{a, b} {
		t.Chdir(filepath.Join(repo, n))
		must(t, "init")
		goals[n] = must(t, "add", "goal", "--title", "goal of "+n, "--why", "w")
	}
	// Both reads happen at the same record tip.
	for _, n := range []string{a, b} {
		t.Chdir(filepath.Join(repo, n))
		res := hyp(t, "show", goals[n])
		if res.code != 0 || !strings.Contains(res.stdout, "goal of "+n) {
			t.Errorf("store %s: hyp show %s exit %d, stdout %q, stderr %q; want its Goal", n, goals[n], res.code, res.stdout, res.stderr)
		}
		st := hyp(t, "status", goals[n])
		if st.code != 0 || !strings.Contains(st.stdout, "goal of "+n) {
			t.Errorf("store %s: hyp status %s exit %d, stdout %q, stderr %q; want its Goal", n, goals[n], st.code, st.stdout, st.stderr)
		}
	}
}

// Finding B (G-SZCBXZ#10). The boundary test skips every directory named
// investigations, recipes or rubrics at any depth, but go builds a package
// in such a directory (for example internal/check/recipes). A non-test Go
// file there that names the store folder and runs git is not reported.
func TestReviewBoundarySkipsBuildablePackages(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "go.mod"), "module m\n")
	src := "package recipes\n\nimport \"os/exec\"\n\nvar dir = \".tr\" + \"ee\"\nvar s = \"x/." + "tree/events\"\n\nfunc f() { exec.Command(\"git\", \"status\").Run() }\n"
	for _, d := range []string{"recipes", "rubrics", "investigations"} {
		writeFile(t, filepath.Join(root, "internal", "check", d, "probe.go"), src)
	}
	bad := boundaryFindings(t, root)
	for _, d := range []string{"recipes", "rubrics", "investigations"} {
		found := false
		for _, b := range bad {
			if strings.HasPrefix(b, "internal/check/"+d+"/") {
				found = true
			}
		}
		if !found {
			t.Errorf("internal/check/%s/probe.go names the store folder and runs git; the boundary test reports nothing for it (findings: %q)", d, bad)
		}
	}
}

// Finding C (G-SZCBXZ#1). A store's folder reached through a symbolic link
// inside the repository is named by the link's path, so discovery there
// finds the empty work-tree folder hyp init left, not the record's store:
// hyp show prints no event the write exited 0 on, and a write there starts
// a second store in the record.
func TestReviewStoreFolderThroughSymlink(t *testing.T) {
	repo := recordRepo(t)
	g := must(t, "add", "goal", "--title", "the goal", "--why", "w")
	if err := os.Symlink("s", filepath.Join(repo, "link")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(repo, "link"))
	res := hyp(t, "show", g)
	if res.code != 0 || !strings.Contains(res.stdout, "the goal") {
		t.Errorf("hyp show %s through the link: exit %d, stdout %q, stderr %q; want the Goal", g, res.code, res.stdout, res.stderr)
	}
	must(t, "add", "goal", "--title", "second", "--why", "w")
	if names := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); strings.Contains(names, "link/") {
		t.Errorf("a write through the link started a second store in the record:\n%s", names)
	}
}

// Finding D (G-SZCBXZ#8, in tension with G-SZCBXZ#7). Two work-tree stores
// pin one bundle, recipes/kit@1 at the repository root, at the same hash.
// Store one is imported while the checkout holds that hash, so the record
// keeps the bundle; then the checkout's bundle is edited. Store two, still a
// work-tree store, replays its Result as questionable; once imported, the
// record's copy holds the pinned hash and the Result is no longer
// questionable, so the imported Goal does not print what its work-tree
// store replays to.
func TestReviewImportStatusDiffersWhenRecordHoldsTheBundle(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	gitIn(t, repo, "init", "-q")
	bundleIn(t, repo, "kit", "", map[string]string{"run.sh": "#!/bin/sh\n"})
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	type ids struct{ g, r string }
	got := map[string]ids{}
	for _, name := range []string{"one", "two"} {
		if err := os.Mkdir(filepath.Join(repo, name), 0o755); err != nil {
			t.Fatal(err)
		}
		t.Chdir(filepath.Join(repo, name))
		must(t, "init")
		dropRecord(t, repo)
		g := must(t, "add", "goal", "--title", "G "+name, "--why", "w")
		q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "w")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
		ptr := ".tree/evidence/" + e + "/out.txt"
		writeFile(t, ptr, "ran\n")
		r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "ran", "--evidence", ptr, "--why", "w")
		got[name] = ids{g, r}
	}
	t.Chdir(repo)
	want(t, hyp(t, "import", "one"), 0, "imported")
	writeFile(t, filepath.Join(repo, "recipes", "kit@1", "run.sh"), "#!/bin/sh\necho edited\n")
	t.Chdir(filepath.Join(repo, "two"))
	view := func() string {
		return hyp(t, "status", got["two"].g).stdout + hyp(t, "show", got["two"].r).stdout
	}
	before := view()
	if !strings.Contains(before, "questionable") {
		t.Fatalf("precondition: the work-tree store two should read its Result questionable:\n%s", before)
	}
	t.Chdir(repo)
	want(t, hyp(t, "import", "two"), 0, "imported")
	t.Chdir(filepath.Join(repo, "two"))
	if after := view(); after != before {
		t.Errorf("store two after import prints\n%s\nits work-tree store replayed to\n%s", after, before)
	}
}
