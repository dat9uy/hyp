package record

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

// putVarint encodes v as a reftable varint.
func putVarint(v uint64) []byte {
	var buf [10]byte
	i := len(buf) - 1
	buf[i] = byte(v & 0x7f)
	for v >>= 7; v != 0; v >>= 7 {
		v--
		i--
		buf[i] = 0x80 | byte(v&0x7f)
	}
	return append([]byte(nil), buf[i:]...)
}

// oneRecordTable is a version 1 table whose one ref block holds rec.
func oneRecordTable(rec []byte) []byte {
	b := []byte("REFT\x01\x00\x00\x00")
	b = append(b, make([]byte, 16)...)
	n := 24 + 4 + len(rec) + 3 + 2
	b = append(b, 'r', byte(n>>16), byte(n>>8), byte(n))
	b = append(b, rec...)
	return append(b, 0, 0, 24, 0, 1)
}

// stack writes one table as the whole reftable stack in a new folder.
func stack(t *testing.T, table []byte) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "tables.list"), []byte("a.ref\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "a.ref"), table, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// A ref record whose lengths do not fit the block reads as unknown, never
// as a panic, so with no git the command exits 3.
func TestReftableHoldsRefusesLengthsPastTheBlock(t *testing.T) {
	for name, rec := range map[string][]byte{
		"prefix length past int":        slices.Concat(putVarint(1<<63), putVarint(1<<3), []byte("a"), putVarint(0)),
		"suffix length past int":        slices.Concat(putVarint(0), putVarint(^uint64(0)-7), []byte("a"), putVarint(0)),
		"symref target length past int": slices.Concat(putVarint(0), putVarint(1<<3|3), []byte("a"), putVarint(0), putVarint(^uint64(0)-999)),
		"symref target past the block":  slices.Concat(putVarint(0), putVarint(1<<3|3), []byte("a"), putVarint(0), putVarint(50)),
	} {
		t.Run(name, func(t *testing.T) {
			if _, ok := reftableRefSafe(t, oneRecordTable(rec)); ok {
				t.Fatal("a malformed table reads as known")
			}
			if _, ok := reftableHolds(stack(t, oneRecordTable(rec)), Ref); ok {
				t.Fatal("a malformed stack reads as known")
			}
		})
	}
}

// Every truncation and every one-byte change of a table git wrote reads
// without a panic: as git reads it, or as unknown.
func TestReftableHoldsNeverPanicsOnACorruptTable(t *testing.T) {
	repo := t.TempDir()
	if out, err := exec.Command("git", "init", "-q", "--ref-format=reftable", repo).CombinedOutput(); err != nil {
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
	var in strings.Builder
	for i := range 40 {
		fmt.Fprintf(&in, "create refs/a/t%03d %s\n", i, obj)
	}
	fmt.Fprintf(&in, "create %s %s\nsymref-create refs/hyp/s refs/a/t001\n", Ref, obj)
	uc := exec.Command("git", "-C", repo, "update-ref", "--stdin")
	uc.Stdin = strings.NewReader(in.String())
	if out, err := uc.CombinedOutput(); err != nil {
		t.Fatalf("update-ref: %v %s", err, out)
	}
	git(t, repo, "pack-refs", "--all")
	dir := filepath.Join(repo, ".git", "reftable")
	list, err := os.ReadFile(filepath.Join(dir, "tables.list"))
	if err != nil {
		t.Fatal(err)
	}
	tbl, err := os.ReadFile(filepath.Join(dir, strings.Fields(string(list))[0]))
	if err != nil {
		t.Fatal(err)
	}
	if holds, ok := reftableHolds(stack(t, tbl), Ref); !ok || !holds {
		t.Fatalf("the table git wrote: holds %v, ok %v", holds, ok)
	}
	for n := range len(tbl) {
		if _, ok := reftableRefSafe(t, tbl[:n]); ok && n < 24 {
			t.Fatalf("a table cut to %d bytes reads as known", n)
		}
	}
	for i := range len(tbl) {
		for _, v := range []byte{0x00, 0xff, tbl[i] ^ 0x80, tbl[i] + 1} {
			c := slices.Clone(tbl)
			c[i] = v
			reftableRefSafe(t, c)
		}
	}
}

// reftableRefSafe reads one table and fails the test when reading panics.
func reftableRefSafe(t *testing.T, table []byte) (typ byte, ok bool) {
	t.Helper()
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("reading a corrupt table panicked: %v", r)
		}
	}()
	typ, _, err := reftableRef(table, Ref)
	return typ, err == nil
}
