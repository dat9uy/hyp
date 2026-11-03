package sandbox

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// patchFor writes a patch over host's a.txt ("a\n") that changes it to body.
func patchFor(t *testing.T, from, body string) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "arm.diff")
	writeFile(t, p, "diff --git a/a.txt b/a.txt\n--- a/a.txt\n+++ b/a.txt\n@@ -1 +1 @@\n-"+from+"\n+"+body+"\n")
	return p
}

// A patch that applies is copied to patches/<ARM>.patch, recorded with the
// copy's sha256, passed to Launch in HYP_SANDBOX_PATCHES, never applied to
// wt/, and kept by remove with the same sha256.
func TestPatchKeptAndPassed(t *testing.T) {
	setup(t)
	fix := patchFor(t, "a", "fixed")
	root := recipeRoot(t, "arms", map[string]string{
		"recipe.md": bundle("arms", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\necho \"$HYP_SANDBOX_PATCHES\" > \"$HYP_SANDBOX_TMP/patches.txt\"\nfor f in \"$HYP_SANDBOX_PATCHES\"/*; do echo \"${f##*/}\"; done >> \"$HYP_SANDBOX_TMP/patches.txt\"\n",
	}, "launch.sh")
	rec, box, err := createWith(t, "pk", CreateOptions{Refs: []string{"arms@1"}, Recipes: root, Patches: []Patch{{Arm: "fix", File: fix}}})
	if err != nil {
		t.Fatal(err)
	}
	cp := filepath.Join(box, "patches", "fix.patch")
	sum, _ := fileSum(cp)
	if want, _ := fileSum(fix); len(rec.Patches) != 1 || rec.Patches[0] != (PatchSum{Arm: "fix", SHA256: want}) || sum != want {
		t.Fatalf("patches = %+v, copy sha256 %s", rec.Patches, sum)
	}
	if got, _ := os.ReadFile(filepath.Join(rec.TmpDir(), "patches.txt")); string(got) != rec.PatchesDir()+"\nfix.patch\n" {
		t.Errorf("Launch saw %q", got)
	}
	if got, _ := os.ReadFile(filepath.Join(rec.WTDir(), "a.txt")); string(got) != "a\n" {
		t.Errorf("wt/a.txt = %q; a patch is never applied to wt/", got)
	}
	if _, err := Remove(RemoveOptions{ID: "pk", OutDir: filepath.Join(t.TempDir(), "x"), Keep: []string{"patches"}}); code(err) != 2 {
		t.Errorf("--keep patches: err = %v, want 2", err)
	}
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "pk", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	kept, _ := fileSum(filepath.Join(out, "patches", "fix.patch"))
	var saved struct{ Patches []PatchSum }
	data, _ := os.ReadFile(filepath.Join(out, "sandbox.json"))
	if err := json.Unmarshal(data, &saved); err != nil || len(saved.Patches) != 1 || saved.Patches[0].SHA256 != kept {
		t.Errorf("kept patch sha256 %s, record %s", kept, data)
	}
}

// A patch that does not apply to the base is refused with 2 before Launch,
// and nothing is left behind.
func TestPatchThatDoesNotApply(t *testing.T) {
	setup(t)
	marker := filepath.Join(t.TempDir(), "launched")
	root := recipeRoot(t, "arms", map[string]string{
		"recipe.md": bundle("arms", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\n: > " + marker + "\n",
	}, "launch.sh")
	bad := patchFor(t, "not in a.txt", "x")
	_, box, err := createWith(t, "pb", CreateOptions{Refs: []string{"arms@1"}, Recipes: root, Patches: []Patch{{Arm: "bad", File: bad}}})
	if code(err) != 2 || !strings.Contains(err.Error(), "patch bad") || !strings.Contains(err.Error(), "does not apply") || !strings.Contains(err.Error(), "partial sandbox removed") {
		t.Fatalf("err = %v", err)
	}
	if exists(box) || exists(marker) {
		t.Errorf("sandbox left: %v, Launch ran: %v", exists(box), exists(marker))
	}
}

func TestPatchBadArguments(t *testing.T) {
	setup(t)
	good := patchFor(t, "a", "b")
	for name, ps := range map[string][]Patch{
		"bad arm":      {{Arm: "Fix", File: good}},
		"digit first":  {{Arm: "1x", File: good}},
		"repeated arm": {{Arm: "x", File: good}, {Arm: "x", File: good}},
		"missing file": {{Arm: "x", File: filepath.Join(t.TempDir(), "none.patch")}},
		"directory":    {{Arm: "x", File: t.TempDir()}},
	} {
		_, box, err := createWith(t, "pa", CreateOptions{Refs: []string{"none"}, Patches: ps})
		if code(err) != 2 || exists(box) {
			t.Errorf("%s: err = %v, sandbox left %v", name, err, exists(box))
		}
	}
	for _, s := range []string{"x", "=f", "x=", ""} {
		if _, err := ParsePatch(s); code(err) != 2 {
			t.Errorf("ParsePatch(%q) = %v, want 2", s, err)
		}
	}
	if p, err := ParsePatch("fix=a=b.patch"); err != nil || p != (Patch{Arm: "fix", File: "a=b.patch"}) {
		t.Errorf("ParsePatch = %+v, %v", p, err)
	}
}
