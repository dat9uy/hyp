package cli

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// caseInsensitive reports whether the filesystem holding dir folds case.
func caseInsensitive(t *testing.T, dir string) bool {
	t.Helper()
	p := filepath.Join(dir, "probe-lower")
	writeFile(t, p, "x")
	defer os.Remove(p)
	_, err := os.Stat(filepath.Join(dir, "PROBE-LOWER"))
	return err == nil
}

// G-SZCBXZ#1: in a record repository, an event a write exited 0 on is
// printed by hyp status and hyp show run in its store's folder. On a
// case-insensitive filesystem (macOS by default) the folder s can be
// entered as S: it is the same folder, and on the base hyp finds the same
// store there. The candidate names the store by the path as typed, so from
// S it finds no record store, falls back to the empty work-tree .tree/,
// and a write from S starts a second store "S" in the record that s never
// shows.
func TestReviewStoreFolderEnteredWithOtherCase(t *testing.T) {
	repo := recordRepo(t)
	if !caseInsensitive(t, repo) {
		t.Skip("filesystem is case-sensitive")
	}
	g, r, ptr := chain(t, "the bytes\n")
	want := views(t, g, r, ptr)

	t.Chdir(filepath.Join(repo, "S"))
	res := hyp(t, "status", g)
	if res.code != 0 || res.stdout != want[0] {
		t.Errorf("hyp status %s from S (the folder s): exit %d, stdout %q, stderr %q; want what s prints:\n%s", g, res.code, res.stdout, res.stderr, want[0])
	}
	g2 := hyp(t, "add", "goal", "--title", "from S", "--why", "w")
	if g2.code != 0 {
		t.Fatalf("hyp add goal from S: exit %d, stderr %q", g2.code, g2.stderr)
	}
	id := strings.TrimSpace(strings.Fields(g2.stdout)[0])

	t.Chdir(filepath.Join(repo, "s"))
	if res := hyp(t, "show", id); res.code != 0 {
		t.Errorf("a write from S exited 0 with %s, but hyp show %s in s: exit %d, stderr %q", id, id, res.code, res.stderr)
	}
	ls := gitIn(t, repo, "ls-tree", "-r", "--name-only", "refs/hyp/record")
	if strings.Contains(ls, "S/.tree/events/") {
		t.Errorf("the record now holds two stores for one folder:\n%s", ls)
	}
}

// G-SZCBXZ#8: hyp import adds every event "with the bytes of each file
// those events name". A Result written in a work-tree store stores its
// evidence file's sha256; when that file changed on disk after the
// Result, the candidate's import keeps the changed bytes under the path
// with no warning (hyp import passes no sha256 to the write, unlike every
// other write, which refuses bytes that do not hash as the event stores),
// and hyp show PATH then prints bytes no event names.
func TestReviewImportKeepsBytesNoEventNames(t *testing.T) {
	gitConfig(t)
	top, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	// The store is made before the folder is a git repository, so it is a
	// work-tree store, as a store made by an older hyp or in a plain folder.
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
	if res.code != 0 {
		t.Fatalf("hyp import .: exit %d, stderr %q", res.code, res.stderr)
	}
	if strings.Contains(res.stderr, ptr) {
		return // the import says it could not keep the cited bytes
	}
	show := hyp(t, "show", ptr)
	h := sha256.Sum256([]byte(show.stdout))
	got := hex.EncodeToString(h[:])
	w := sha256.Sum256([]byte("the bytes the Result cited\n"))
	if want := hex.EncodeToString(w[:]); got != want {
		t.Errorf("after hyp import, hyp show %s prints %q (sha256 %s), not the bytes the Result names (sha256 %s), and the import warned of nothing: stderr %q", ptr, show.stdout, got[:12], want[:12], res.stderr)
	}
}
