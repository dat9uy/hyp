package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Line 1 (store folder spelled another way): on a filesystem that is
// insensitive to Unicode normalization (APFS), the folder "café" typed
// with a decomposed é (NFD) is the same folder as the composed one (NFC)
// it was made with. hyp status from there must print what the folder
// prints, and a write from there must go to the same store.
func TestReviewStoreFolderEnteredWithOtherNormalization(t *testing.T) {
	gitConfig(t)
	repo, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	nfc := "caf\u00e9"
	nfd := "cafe\u0301"
	gitIn(t, repo, "init", "-q")
	writeFile(t, filepath.Join(repo, nfc, "README.md"), "a store\n")
	if _, err := os.Stat(filepath.Join(repo, nfd)); err != nil {
		t.Skip("the filesystem does not fold Unicode normalization")
	}
	gitIn(t, repo, "add", "-A")
	gitIn(t, repo, "commit", "-qm", "base")
	t.Chdir(filepath.Join(repo, nfc))
	must(t, "init")
	g, r, ptr := chain(t, "the bytes\n")
	views0 := views(t, g, r, ptr)
	t.Chdir(filepath.Join(repo, nfd))
	if res := hyp(t, "status", g); res.code != 0 || res.stdout != views0[0] {
		t.Errorf("hyp status %s from the NFD spelling: exit %d, stdout %q, stderr %q; want what the NFC folder prints", g, res.code, res.stdout, res.stderr)
	}
	res := hyp(t, "add", "goal", "--title", "from NFD", "--why", "w")
	t.Logf("hyp add goal from the NFD spelling: exit %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
	ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record")
	stores := map[string]bool{}
	for _, l := range strings.Split(ls, "\n") {
		if i := strings.Index(l, "/.tree/events/"); i >= 0 {
			stores[l[:i]] = true
		}
	}
	if len(stores) != 1 {
		t.Errorf("the record holds %d stores for one folder: %v", len(stores), stores)
	}
}

// Line 8: hyp import adds each event with the bytes of each file it names.
// When the checkout's copy of a cited file changed, the import from the
// checkout leaves the bytes out (as the candidate means it to). Importing
// then from a commit that holds the right bytes must add them, but the
// event is already in the record, so the import adds nothing and the
// record never gets the bytes that the commit holds.
func TestReviewImportFromRefAfterCheckoutImportAddsTheBytes(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(top, "s"), 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(filepath.Join(top, "s"))
	must(t, "init")
	_, _, ptr := chain(t, "the bytes the Result cited\n")
	gitIn(t, top, "init", "-q")
	gitIn(t, top, "add", "-A")
	gitIn(t, top, "commit", "-qm", "store")
	writeFile(t, ptr, "changed after the Result\n")
	res := hyp(t, "import", ".")
	t.Logf("hyp import .: exit %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
	res = hyp(t, "import", "HEAD:.")
	t.Logf("hyp import HEAD:.: exit %d, stdout %q, stderr %q", res.code, res.stdout, res.stderr)
	res = hyp(t, "show", ptr)
	if res.code != 0 || !strings.Contains(res.stdout, "the bytes the Result cited") {
		t.Errorf("hyp show %s after both imports: exit %d, stdout %q, stderr %q; want the cited bytes, which HEAD:. holds", ptr, res.code, res.stdout, res.stderr)
	}
}
