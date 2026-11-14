package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// The reftable reader finds Ref as git does, in tables of either hash, with
// many refs over several blocks, after a deletion, and after compaction.
func TestReftableHoldsReadsAsGitDoes(t *testing.T) {
	for _, format := range []string{"sha1", "sha256"} {
		t.Run(format, func(t *testing.T) {
			repo := t.TempDir()
			cmd := exec.Command("git", "init", "-q", "--ref-format=reftable", "--object-format="+format, repo)
			if out, err := cmd.CombinedOutput(); err != nil {
				t.Skipf("git cannot make a reftable repository: %v %s", err, out)
			}
			t.Setenv("GIT_COMMITTER_NAME", "t")
			t.Setenv("GIT_COMMITTER_EMAIL", "t@t")
			hc := exec.Command("git", "-C", repo, "hash-object", "-w", "--stdin")
			hc.Stdin = strings.NewReader("x\n")
			out, err := hc.Output()
			if err != nil {
				t.Fatal(err)
			}
			obj := strings.TrimSpace(string(out))
			dir := filepath.Join(repo, ".git", "reftable")
			check := func(want bool, when string) {
				t.Helper()
				holds, ok := reftableHolds(dir, Ref)
				if !ok || holds != want {
					t.Fatalf("%s: reftableHolds = %v, ok %v; want %v", when, holds, ok, want)
				}
			}
			check(false, "a new repository")
			var in strings.Builder
			for i := range 3000 {
				fmt.Fprintf(&in, "create refs/a/t%05d %s\n", i, obj)
			}
			fmt.Fprintf(&in, "create refs/hyp/recore %s\ncreate refs/hyp/recordz %s\n", obj, obj)
			uc := exec.Command("git", "-C", repo, "update-ref", "--stdin")
			uc.Stdin = strings.NewReader(in.String())
			if out, err := uc.CombinedOutput(); err != nil {
				t.Fatalf("update-ref: %v %s", err, out)
			}
			check(false, "many refs and names next to the record ref")
			git(t, repo, "update-ref", Ref, obj)
			check(true, "after update-ref")
			git(t, repo, "pack-refs", "--all")
			check(true, "after compaction")
			git(t, repo, "update-ref", "-d", Ref)
			check(false, "after delete")
			git(t, repo, "update-ref", Ref, obj)
			git(t, repo, "update-ref", "refs/a/t00001", obj)
			check(true, "after a later table writes another ref")
			git(t, repo, "pack-refs", "--all")
			check(true, "after a second compaction")
		})
	}
}

// A reftable stack hyp cannot read counts as unknown.
func TestReftableHoldsRefusesAMalformedTable(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tables.list"), []byte("a.ref\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.ref"), []byte("REFT\x01\x00\x10\x00garbage"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, ok := reftableHolds(dir, Ref); ok {
		t.Fatal("a malformed table reads as known")
	}
}

// Without git, a reftable stack hyp cannot read counts as holding the
// record, so the command stops with exit 3 and writes nothing past it.
func TestRefOnDiskCountsAnUnreadableReftable(t *testing.T) {
	top := t.TempDir()
	dir := filepath.Join(top, GitName, "reftable")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "tables.list"), []byte("a.ref\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.ref"), []byte("not a table"), 0o644); err != nil {
		t.Fatal(err)
	}
	if !refOnDisk(top) {
		t.Fatal("refOnDisk reads an unreadable reftable stack as no record")
	}
	if err := os.WriteFile(filepath.Join(dir, "tables.list"), nil, 0o644); err != nil {
		t.Fatal(err)
	}
	if refOnDisk(top) {
		t.Fatal("refOnDisk reads an empty reftable stack as holding the record")
	}
}
