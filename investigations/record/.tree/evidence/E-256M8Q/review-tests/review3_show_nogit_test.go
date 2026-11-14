package cli

import (
	"os"
	"strings"
	"testing"
)

// Line 9: where no .git is at or above the current directory, hyp reads the
// work-tree .tree/ as on the base. hyp show of a cited file removed from the
// work tree should report the missing file, not tell the user to run
// hyp import, a command for a record repository that cannot help here.
func TestReviewShowMissingFileWithNoGitNamesNoImport(t *testing.T) {
	gitConfig(t)
	t.Chdir(t.TempDir())
	must(t, "init")
	_, _, ptr := chain(t, "bytes\n")
	if err := os.Remove(ptr); err != nil {
		t.Fatal(err)
	}
	res := hyp(t, "show", ptr)
	t.Logf("hyp show %s: exit %d, stdout %q, stderr %q", ptr, res.code, res.stdout, res.stderr)
	if strings.Contains(res.stderr, "hyp import") {
		t.Errorf("with no git, hyp show of a removed cited file names hyp import: %q", res.stderr)
	}
}
