package sandbox

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A replay needs only the pinned recipe, the base commit and the files the
// tree keeps: a Standalone run, a sandbox with a patch driven and removed,
// and Evidence on what it kept leave no host path in either record or in the
// == lines, apart from a directory argument outside the repository, and every
// kept patch and input hashes to the sha256 its record names.
func TestReplayNeedsNoHostPath(t *testing.T) {
	setup(t)
	origTemp := os.TempDir()
	tmpdir := t.TempDir()
	t.Setenv("TMPDIR", tmpdir)
	h, base := host(t)
	lib := filepath.Join(h, "recipes", "rep@1")
	files := map[string]string{
		"recipe.md": recipeMD("rep", "## Launch\n\nRun: ./launch.sh\n\n## Drive\n\nRun: ./drive.sh\n\n"+
			"## Evidence\n\nRun: ./ev.sh\n\nExternal: none\n\n## Standalone\n\nRun chk: ./chk.sh\n\nExternal: none\n"),
		"launch.sh": "#!/bin/sh\nset -e\narm=\"$HYP_SANDBOX_RECIPE_TMP/a\"\n" +
			"git -C \"$HYP_SANDBOX_WT\" worktree add -q --detach \"$arm\" HEAD\n" +
			"git -C \"$arm\" apply \"$HYP_SANDBOX_PATCHES/a.patch\"\n" +
			"while IFS= read -r l; do echo \"$l\"; done < \"$arm/a.txt\" > \"$HYP_SANDBOX_RECIPE_TMP/built\"\n",
		"drive.sh":                     "#!/bin/sh\nwhile IFS= read -r l; do echo \"$l\"; done < \"$HYP_SANDBOX_RECIPE_TMP/built\" > \"$HYP_SANDBOX_WT/driven.txt\"\n",
		"ev.sh":                        "#!/bin/sh\necho ok\n",
		"chk.sh":                       "#!/bin/sh\necho ok\n",
		"known-answers/ev/expect.json": expect("evidence", "./ev.sh", nil, 0, "ok\n", ""),
		"known-answers/ev/input/a":     "",
		"known-answers/sa/expect.json": expect("standalone", "chk", []string{"x"}, 0, "ok\n", ""),
		"known-answers/sa/input/a":     "",
	}
	for rel, body := range files {
		writeFile(t, filepath.Join(lib, rel), body)
	}
	for _, s := range []string{"launch.sh", "drive.sh", "ev.sh", "chk.sh"} {
		if err := os.Chmod(filepath.Join(lib, s), 0o755); err != nil {
			t.Fatal(err)
		}
	}

	// hyp recipe run: an absolute file, a relative file, an in-repo directory
	// and an out-of-repo directory.
	outside := t.TempDir()
	abs := filepath.Join(t.TempDir(), "cases.tsv")
	writeFile(t, abs, "a\tb\n")
	writeFile(t, filepath.Join(h, "data", "rel.txt"), "rel\n")
	if err := os.MkdirAll(filepath.Join(h, "sub"), 0o755); err != nil {
		t.Fatal(err)
	}
	ev := filepath.Join(h, "inv", ".tree", "evidence", "E-1")
	runRec := filepath.Join(ev, "run")
	var warn bytes.Buffer
	if _, err := Run(RunOptions{Dir: h, Ref: "rep@1", Step: "standalone", Label: "chk", Args: []string{abs, "data/rel.txt", "./sub", outside}, Record: runRec, Warn: &warn}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(warn.String(), "argument 4 names a directory outside the repository") {
		t.Errorf("warn = %q", warn.String())
	}

	// hyp sandbox create --patch a=FILE, drive, remove to an OUT_DIR inside the host.
	patch := patchFor(t, "a", "patched")
	rec, _, err := createWith(t, "e2e", CreateOptions{Host: h, Base: base, Refs: []string{"rep@1"}, Recipes: h, Patches: []Patch{{Arm: "a", File: patch}}})
	if err != nil {
		t.Fatal(err)
	}
	var drove bytes.Buffer
	if _, err := Drive("e2e", &drove); err != nil {
		t.Fatalf("drive: %v\n%s", err, drove.String())
	}
	if got, _ := os.ReadFile(filepath.Join(rec.WTDir(), "driven.txt")); string(got) != "patched\n" {
		t.Errorf("the patched arm built %q", got)
	}
	out := filepath.Join(ev, "out")
	if _, err := Remove(RemoveOptions{ID: "e2e", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	evRec := filepath.Join(ev, "evidence")
	if _, err := Run(RunOptions{Dir: h, Ref: "rep@1", Step: "evidence", Args: []string{out}, Record: evRec}); err != nil {
		t.Fatal(err)
	}

	// A1: no host prefix, apart from the out-of-repo directory word.
	var prefixes []string
	for _, p := range []string{os.Getenv("HOME"), tmpdir, origTemp, filepath.Dir(h)} {
		prefixes = append(prefixes, filepath.Clean(p))
		if r, err := filepath.EvalSymlinks(p); err == nil {
			prefixes = append(prefixes, r)
		}
	}
	scan := func(name, text string) {
		text = strings.ReplaceAll(text, outside, "OUTSIDE")
		for _, p := range prefixes {
			if strings.Contains(text, p) {
				t.Errorf("%s names %s:\n%s", name, p, text)
			}
		}
	}
	read := func(p string) string {
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		return string(data)
	}
	eqLines := func(p string) string {
		var keep []string
		for l := range strings.SplitSeq(read(p), "\n") {
			if strings.HasPrefix(l, "== ") {
				keep = append(keep, l)
			}
		}
		if len(keep) == 0 {
			t.Errorf("%s has no == line", p)
		}
		return strings.Join(keep, "\n")
	}
	scan("run command.json", read(filepath.Join(runRec, "command.json")))
	scan("run command.log", eqLines(filepath.Join(runRec, "command.log")))
	scan("evidence command.json", read(filepath.Join(evRec, "command.json")))
	scan("evidence command.log", eqLines(filepath.Join(evRec, "command.log")))
	scan("sandbox.json", read(filepath.Join(out, "sandbox.json")))
	scan("recipe.log", eqLines(filepath.Join(out, "recipe.log")))
	if !strings.Contains(read(filepath.Join(runRec, "command.json")), outside) {
		t.Error("the out-of-repo directory is not in the record as given")
	}

	// A2: each kept patch and input hashes to the sha256 its record names.
	saved := savedRecord(t, out)
	if len(saved.Patches) != 1 {
		t.Fatalf("patches = %+v", saved.Patches)
	}
	if sum, _ := fileSum(filepath.Join(out, "patches", "a.patch")); sum != saved.Patches[0].SHA256 {
		t.Errorf("kept patch sha256 %s, record %s", sum, saved.Patches[0].SHA256)
	}
	cr := readCommandRecord(t, runRec)
	if len(cr.Inputs) != 2 {
		t.Fatalf("inputs = %+v", cr.Inputs)
	}
	for _, in := range cr.Inputs {
		if sum, _ := fileSum(filepath.Join(runRec, filepath.FromSlash(in.Path))); !in.Kept || sum != in.SHA256 {
			t.Errorf("input %+v: kept copy hashes %s", in, sum)
		}
	}
}
