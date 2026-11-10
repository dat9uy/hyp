package cli

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/herr"
	"github.com/dat9uy/hyp/internal/replay"
)

// wantCode checks the exit code, and for 1 the refused: prefix, which no
// other code may carry.
func wantCode(t *testing.T, r result, code int) {
	t.Helper()
	if r.code != code || (code == 1) != strings.HasPrefix(r.stderr, "refused:") {
		t.Fatalf("exit %d (want %d), stderr %q", r.code, code, r.stderr)
	}
}

// A node ref that names nothing, or more than one node, is fixed on the
// command line.
func TestNodeRefIsBadArgument(t *testing.T) {
	setup(t)
	wantCode(t, hyp(t, "show", "H-NOPE00"), 2)
	p := replay.Projection{Nodes: map[string]*replay.Node{
		"a": {ID: "a", Kind: "hypothesis", Alias: "twin"},
		"b": {ID: "b", Kind: "hypothesis", Alias: "twin"},
	}}
	if _, err := resolveRef(p, "twin"); herr.KindOf(err) != herr.BadArguments {
		t.Fatalf("ambiguous ref: %v (kind %d), want bad arguments", err, herr.KindOf(err))
	}
}

// An --evidence pointer that names no file is a bad argument; a recipe it
// names that breaks a rule is a refusal.
func TestEvidencePointerCodes(t *testing.T) {
	repo, tree, q := libraryStore(t)
	bundleIn(t, repo, "bad", "## Launch\n\nRun: launch.sh\n", map[string]string{"launch.sh": "#!/bin/sh\n"})
	writeFile(t, filepath.Join(repo, "recipes", "single.md"), "---\nname: single\nversion: 1\nformat: 1\n---\n")
	writeFile(t, filepath.Join(repo, "recipes", "boot@1", "cases", "known.txt"), "k\n")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	rec := func(ev string) result {
		return hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--evidence", ev, "--why", "w")
	}
	wantCode(t, rec("notes.md"), 2)
	wantCode(t, rec(filepath.Join(tree, ".tree", "evidence", "x")), 2)
	wantCode(t, rec("recipes/nope@1/a.txt"), 2)
	wantCode(t, rec("recipes/single@1/a.txt"), 2)
	wantCode(t, rec("recipes/boot@1/cases/missing.txt"), 2)
	wantCode(t, rec("recipes/boot@1/cases"), 2)
	wantCode(t, rec("recipes/bad@1/launch.sh"), 1)
}

// hyp bundle hash DIR: no recipe.md is a bad argument, a bundle that breaks
// a rule is a refusal, and one that cannot be read is the environment.
func TestRecipeHashDirCodes(t *testing.T) {
	dir := t.TempDir()
	wantCode(t, hyp(t, "bundle", "hash", dir), 2)
	bundleIn(t, dir, "bad", "## Launch\n\nRun: launch.sh\n", map[string]string{"launch.sh": "#!/bin/sh\n"})
	wantCode(t, hyp(t, "bundle", "hash", filepath.Join(dir, "recipes", "bad@1")), 1)
	bundleIn(t, dir, "shut", "", nil)
	md := filepath.Join(dir, "recipes", "shut@1", "recipe.md")
	if err := os.Chmod(md, 0); err != nil {
		t.Fatal(err)
	}
	defer os.Chmod(md, 0o644)
	if _, err := os.ReadFile(md); err == nil {
		t.Skip("running as a user who reads mode-0 files")
	}
	wantCode(t, hyp(t, "bundle", "hash", filepath.Join(dir, "recipes", "shut@1")), 3)
}

// hyp sandbox drive: a copy that is missing is the environment; one outside
// the root, or one that no longer opens, is refused.
func TestDriveCopyCodes(t *testing.T) {
	if _, err := exec.LookPath("lsof"); err != nil {
		t.Skip("lsof not on PATH")
	}
	home := t.TempDir()
	t.Setenv("HOME", home)
	h := t.TempDir()
	gitIn(t, h, "init", "-q", "-b", "main")
	writeFile(t, filepath.Join(h, "a.txt"), "a\n")
	gitIn(t, h, "add", "a.txt")
	gitIn(t, h, "-c", "user.name=t", "-c", "user.email=t@t", "commit", "-qm", "init")
	sha := gitIn(t, h, "rev-parse", "HEAD")
	stub := filepath.Join(t.TempDir(), "hyp")
	if err := os.WriteFile(stub, []byte("#!/bin/sh\necho hyp 0.0.0-test\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	lib := t.TempDir()
	bundleIn(t, lib, "drv", "## Drive\n\nRun: ./drive.sh\n", map[string]string{"drive.sh": "#!/bin/sh\n"})
	create := func(id string) (meta, copyDir string) {
		must(t, "sandbox", "create", id, "--host", h, "--base", sha, "--hyp", stub, "--recipe", "drv@1", "--recipes", lib)
		box := filepath.Join(home, ".hyp-sandbox", id)
		return filepath.Join(box, "sandbox.json"), filepath.Join(box, "recipes", "drv")
	}

	_, cp := create("c1")
	writeFile(t, filepath.Join(cp, "recipe.md"), "---\nname: drv\nversion: 1\nformat: 1\n---\n\n## Drive\n\nRun: drive.sh\n")
	wantCode(t, hyp(t, "sandbox", "drive", "c1"), 1)

	meta, cp := create("c2")
	data, err := os.ReadFile(meta)
	if err != nil {
		t.Fatal(err)
	}
	// A copy an older record names outside the sandboxes root is refused.
	writeFile(t, meta, strings.Replace(string(data), `"name": "drv",`, `"name": "drv", "path": "`+t.TempDir()+`",`, 1))
	wantCode(t, hyp(t, "sandbox", "drive", "c2"), 1)

	_, cp = create("c3")
	if err := os.RemoveAll(cp); err != nil {
		t.Fatal(err)
	}
	wantCode(t, hyp(t, "sandbox", "drive", "c3"), 3)
}
