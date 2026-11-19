package gitread

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

// CopyWorktree copies what a commit of every change would hold: an edited
// file with its new bytes, a new file, an exec bit and a symlink, and not an
// ignored file, a deleted file or a skipped folder. It writes no git state.
func TestCopyWorktreeHoldsWhatACommitWould(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{".gitignore": "ignored\n", "a": "a\n", "gone": "g\n", "goals/x": "x\n"})
	write := func(rel, body string, perm os.FileMode) {
		if err := os.MkdirAll(filepath.Dir(filepath.Join(r.Dir, rel)), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.Dir, rel), []byte(body), perm); err != nil {
			t.Fatal(err)
		}
	}
	write("a", "edited\n", 0o644)
	write("sub/new", "new\n", 0o755)
	write("ignored", "i\n", 0o644)
	write("goals/y", "y\n", 0o644)
	r.Symlink(t, "link", "a")
	if err := os.Remove(filepath.Join(r.Dir, "gone")); err != nil {
		t.Fatal(err)
	}
	before := testutil.Snapshot(t, r.Dir)
	dest := filepath.Join(t.TempDir(), "subject")
	if err := CopyWorktree(r.Dir, dest, "goals"); err != nil {
		t.Fatal(err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("CopyWorktree changed .git")
	}
	files, err := diskFiles(dest)
	if err != nil {
		t.Fatal(err)
	}
	want := map[string]string{".gitignore": "100644", "a": "100644", "sub/new": "100755", "link": "120000"}
	if len(files) != len(want) {
		t.Fatalf("copied %v, want %v", files, want)
	}
	for p, mode := range want {
		if files[p].Mode != mode {
			t.Errorf("%s: mode %q, want %s", p, files[p].Mode, mode)
		}
	}
	if b, _ := os.ReadFile(filepath.Join(dest, "a")); string(b) != "edited\n" {
		t.Errorf("a holds %q, want the bytes on disk", b)
	}
}

// A nested repository in the checkout is refused, as an extracted tree
// refuses a submodule.
func TestCopyWorktreeRefusesNestedRepository(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a\n"})
	nested := filepath.Join(r.Dir, "inner")
	if out, err := exec.Command("git", "init", "-q", nested).CombinedOutput(); err != nil {
		t.Fatalf("git init: %v %s", err, out)
	}
	if err := os.WriteFile(filepath.Join(nested, "f"), []byte("f\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	err := CopyWorktree(r.Dir, filepath.Join(t.TempDir(), "subject"))
	if fault.KindOf(err) != fault.Refusal || fault.CodeOf(err) != fault.Submodule {
		t.Fatalf("err = %v, want a submodule refusal", err)
	}
}
