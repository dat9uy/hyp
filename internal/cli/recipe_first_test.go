package cli

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/store"
)

// hyp bundle new NAME@1 writes a first bundle that fails its known answers
// and that add exp refuses until every placeholder is replaced; filled, it
// passes both.
func TestRecipeNewFirstVersionFlow(t *testing.T) {
	dir, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	addExp := func() result {
		t.Helper()
		return hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "g@1", "--rubric", "reading", "--target", "none", "--why", "w")
	}

	bundle := filepath.Join(dir, "recipes", "g@1")
	r := hyp(t, "bundle", "new", "g@1")
	want(t, r, 0, bundle)

	r = hyp(t, "bundle", "known-answers", "g@1")
	want(t, r, 1, "FAIL must-pass: exit 1, want 0")
	want(t, r, 1, "refused: recipe g@1: 2 of 2 known answers failed")
	want(t, addExp(), 1, "refused: recipe g@1 still holds the template placeholder (hyp:placeholder) in check.sh, recipe.md")
	p, _, err := store.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	for _, n := range p.Nodes {
		if n.Kind == event.KindExperiment {
			t.Fatalf("Experiment %s added from a placeholder bundle", n.ID)
		}
	}

	// A second first version of the name is refused; a next version needs
	// --from.
	want(t, hyp(t, "bundle", "new", "g@1"), 1, "recipe g exists at "+bundle)
	want(t, hyp(t, "bundle", "new", "g@2"), 2, "missing --from: g@2 is not a first version")
	want(t, hyp(t, "bundle", "new", "g@2", "--from", "g@1"), 0, "from: g@1 ")

	// Fill the bundle.
	writeFile(t, filepath.Join(bundle, "recipe.md"), "---\nname: g\nversion: 1\nformat: 1\nsubject: the value file of a case input\n---\n\n# g\n\n"+
		"## Launch\n\n## Doctor\n\n## Drive\n\n## Evidence\n\n## Standalone\n\nRun check: ./check.sh\nExternal: none\n\n## Cleanup\n")
	writeFile(t, filepath.Join(bundle, "check.sh"), "#!/bin/sh\nif [ \"$(cat \"$1/value\")\" = ok ]; then echo pass; exit 0; fi\necho fail\nexit 1\n")
	if err := os.Chmod(filepath.Join(bundle, "check.sh"), 0o755); err != nil {
		t.Fatal(err)
	}
	for name, value := range map[string]string{"must-pass": "ok\n", "must-fail": "no\n"} {
		input := filepath.Join(bundle, "known-answers", name, "input")
		if err := os.Remove(filepath.Join(input, "README.txt")); err != nil {
			t.Fatal(err)
		}
		writeFile(t, filepath.Join(input, "value"), value)
	}
	want(t, hyp(t, "bundle", "known-answers", "g@1"), 0, "ok must-fail\nok must-pass\n")
	if r := addExp(); r.code != 0 || r.stderr != "" {
		t.Fatalf("filled bundle: %+v", r)
	}
}

// Outside a store, hyp bundle new NAME@1 fails as other store commands do,
// and a bad ref is still a bad argument.
func TestRecipeNewFirstVersionOutsideStore(t *testing.T) {
	t.Chdir(t.TempDir())
	want(t, hyp(t, "bundle", "new", "g@1"), 3, "no .tree store found; run 'hyp init'")
	want(t, hyp(t, "bundle", "new", "g@2"), 2, "missing --from")
	want(t, hyp(t, "bundle", "new", "g"), 2, "bad recipe ref")
	if _, err := os.Stat("recipes"); !os.IsNotExist(err) {
		t.Fatalf("recipes/ written outside a store (%v)", err)
	}
}
