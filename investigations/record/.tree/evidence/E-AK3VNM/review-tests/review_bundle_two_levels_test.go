package cli

import (
	"path/filepath"
	"strings"
	"testing"
)

// One record repository, two worktrees on two branches. The branch of
// worktree w1 holds kit@1 beside the store (s/recipes/kit@1); the branch of
// worktree w2 holds kit@1, with other bytes, in the repository's top
// library (recipes/kit@1). Each worktree pins its own copy in a Result.
// The record keeps both copies, each at its own path. A Result that was
// not questionable when its write exited 0 must stay so, from every
// checkout, after the other worktree's write.
func TestReviewPinnedBundleAtTwoLibraryLevels(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	w1 := filepath.Join(top, "w1")
	gitIn(t, top, "init", "-q", "-b", "main", w1)
	writeFile(t, filepath.Join(w1, "s", "README.md"), "a store\n")
	gitIn(t, w1, "add", "-A")
	gitIn(t, w1, "commit", "-qm", "base")
	t.Chdir(filepath.Join(w1, "s"))
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "w")

	w2 := filepath.Join(top, "w2")
	gitIn(t, w1, "worktree", "add", "-q", "-b", "other", w2)

	// w1: kit@1 beside the store.
	bundleIn(t, filepath.Join(w1, "s"), "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho near\n"})
	q1 := must(t, "add", "dq", "--parent", g, "--title", "Q1", "--why", "w")
	h1 := must(t, "add", "hyp", "--parent", q1, "--title", "H1", "--why", "w")
	e1 := must(t, "add", "exp", "--parent", h1, "--title", "E1", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	writeFile(t, ".tree/evidence/"+e1+"/out.txt", "ran near\n")
	r1 := must(t, "record", "--exp", e1, "--verdict", h1+"=supports", "--observation", "ran", "--evidence", ".tree/evidence/"+e1+"/out.txt", "--why", "w")
	if out := hyp(t, "show", r1).stdout; strings.Contains(out, "questionable") {
		t.Fatalf("%s reads questionable right after its write:\n%s", r1, out)
	}

	// w2: kit@1 with other bytes in the top library, none beside the store.
	t.Chdir(filepath.Join(w2, "s"))
	bundleIn(t, w2, "kit", "", map[string]string{"run.sh": "#!/bin/sh\necho far\n"})
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H2", "--why", "w")
	e2 := hyp(t, "add", "exp", "--parent", h2, "--title", "E2", "--recipe", "kit@1", "--rubric", "reading", "--why", "w")
	if e2.code != 0 {
		t.Fatalf("w2 cannot add an Experiment on its kit@1: exit %d %s", e2.code, e2.stderr)
	}
	eid := strings.Fields(e2.stdout)[0]
	writeFile(t, ".tree/evidence/"+eid+"/out.txt", "ran far\n")
	r2 := hyp(t, "record", "--exp", eid, "--verdict", h2+"=supports", "--observation", "ran", "--evidence", ".tree/evidence/"+eid+"/out.txt", "--why", "w")
	if r2.code != 0 {
		t.Fatalf("w2 cannot record on its kit@1: exit %d %s", r2.code, r2.stderr)
	}
	r2id := strings.Fields(r2.stdout)[0]

	t.Logf("record files:\n%s", gitIn(t, w1, "ls-tree", "-r", "--name-only", "refs/hyp/record"))
	for _, wt := range []string{w1, w2} {
		t.Chdir(filepath.Join(wt, "s"))
		for _, id := range []string{r1, r2id} {
			if out := hyp(t, "show", id).stdout; strings.Contains(out, "questionable") {
				t.Errorf("from %s, %s reads questionable after both writes exited 0:\n%s", filepath.Base(wt), id, out)
			}
		}
	}
}
