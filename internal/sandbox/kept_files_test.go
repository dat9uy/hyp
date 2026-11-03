package sandbox

import (
	"os"
	"path/filepath"
	"slices"
	"testing"
)

// Every name remove writes into OUT_DIR is in KeptFiles, so keepPaths
// reserves it.
func TestRemoveWritesOnlyKeptFiles(t *testing.T) {
	setup(t)
	in := filepath.Join(t.TempDir(), "params.json")
	writeFile(t, in, `{"n": 1}`+"\n")
	root := recipeRoot(t, "use", map[string]string{
		"recipe.md": bundle("use", "Run: ./launch.sh", "", ""),
		"launch.sh": "#!/bin/sh\necho launched\n",
	}, "launch.sh")
	_, box, err := createWith(t, "kept", CreateOptions{Refs: []string{"use@1"}, Recipes: root, Inputs: in, Patches: []Patch{{Arm: "fix", File: patchFor(t, "a", "b")}}})
	if err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(box, keptKnownAnswers), "log\n")
	out := filepath.Join(t.TempDir(), "out")
	if _, err := Remove(RemoveOptions{ID: "kept", OutDir: out}); err != nil {
		t.Fatal(err)
	}
	ents, err := os.ReadDir(out)
	if err != nil {
		t.Fatal(err)
	}
	if len(ents) < 5 {
		t.Fatalf("OUT_DIR holds %d names, want at least sandbox.json, recipe.log, known-answers.log, inputs and patches", len(ents))
	}
	for _, e := range ents {
		if !slices.Contains(KeptFiles, e.Name()) {
			t.Errorf("remove wrote %s, which KeptFiles does not list", e.Name())
		}
	}
}
