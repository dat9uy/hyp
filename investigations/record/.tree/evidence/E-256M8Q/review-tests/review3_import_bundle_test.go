package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Line 7 (with line 8): a work-tree store whose Result pins kit@1 is
// imported first from a checkout that lacks the bundle, so the record lacks
// it, then from the commit that holds it with the pinned hash. A later
// import adds a named file an earlier one lacked; it must also add the
// pinned bundle, so the Result's questionable flag reads the same from a
// checkout with or without recipes/kit@1, and as the work-tree store reads.
func TestReviewLaterImportAddsAPinnedBundleAnEarlierImportLacked(t *testing.T) {
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
	first := hyp(t, "import", "s")
	t.Logf("hyp import s: exit %d, stdout %q, stderr %q", first.code, first.stdout, first.stderr)
	second := hyp(t, "import", "HEAD:s")
	t.Logf("hyp import HEAD:s: exit %d, stdout %q, stderr %q", second.code, second.stdout, second.stderr)
	if ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record"); !strings.Contains(ls, "s/recipes/kit@1/recipe.md") {
		t.Errorf("after an import from a commit that holds kit@1 with its pinned hash, the record lacks it:\n%s", ls)
	}
	t.Chdir(filepath.Join(repo, "s"))
	without := hyp(t, "show", r).stdout
	gitIn(t, repo, "checkout", "--", "s/recipes")
	with := hyp(t, "show", r).stdout
	if strings.Contains(without, "questionable") != strings.Contains(with, "questionable") {
		t.Errorf("%s reads questionable=%v from a checkout without recipes/kit@1 and questionable=%v from one with it", r, strings.Contains(without, "questionable"), strings.Contains(with, "questionable"))
	}
	if strings.Contains(without, "questionable") {
		t.Errorf("%s reads questionable in the record, though the work-tree store read it fine:\n%s", r, without)
	}
}
