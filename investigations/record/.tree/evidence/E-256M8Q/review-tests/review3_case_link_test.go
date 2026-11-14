package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Line 1: on a filesystem that folds case, the folder s typed as S is the
// store s, also when the repository holds a symbolic link "a" to s, whose
// name sorts before s. hyp status from S prints what s prints, and a write
// from S goes to the store s, not to a store named for the link.
func TestReviewStoreFolderOtherCaseBesideALink(t *testing.T) {
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
	res := hyp(t, "status", g)
	if res.code != 0 || res.stdout != views0[0] {
		t.Errorf("hyp status %s from S: exit %d, stdout %q, stderr %q; want what s prints:\n%s", g, res.code, res.stdout, res.stderr, views0[0])
	}
	add := hyp(t, "add", "goal", "--title", "from S", "--why", "w")
	t.Logf("hyp add goal from S: exit %d, stdout %q, stderr %q", add.code, add.stdout, add.stderr)
	ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record")
	if strings.Contains(ls, "a/.tree/") || strings.Contains(ls, "S/.tree/") {
		t.Errorf("the record holds a second store for the folder s:\n%s", ls)
	}
}
