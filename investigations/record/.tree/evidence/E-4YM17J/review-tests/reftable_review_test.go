package record

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

// gitVarint encodes v as a reftable varint, as git's put_var_int does.
func gitVarint(v uint64) []byte {
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

// table builds a version 1 table whose one ref block holds rec.
func table(rec []byte) []byte {
	b := []byte("REFT\x01\x00\x00\x00")
	b = append(b, make([]byte, 16)...) // min and max update index
	blockLen := 24 + 4 + len(rec) + 3 + 2
	b = append(b, 'r', byte(blockLen>>16), byte(blockLen>>8), byte(blockLen))
	b = append(b, rec...)
	b = append(b, 0, 0, 24) // one restart, at the first record
	b = append(b, 0, 1)     // restart count
	return b
}

// A table hyp cannot read must count as unknown (ok false), so refOnDisk
// counts it as holding the record and the command exits 3. A ref record
// whose varints do not fit an int must not panic.
func TestReftableHoldsDoesNotPanicOnHugeVarints(t *testing.T) {
	neg := uint64(1) << 63
	cases := map[string][]byte{
		"prefix length past int": append(append(gitVarint(neg), gitVarint(1<<3|0)...), append([]byte("a"), gitVarint(0)...)...),
		"symref target length past int": append(append(append(gitVarint(0), gitVarint(1<<3|3)...), 'a'), append(gitVarint(0), gitVarint(^uint64(0)-999)...)...),
	}
	for name, rec := range cases {
		t.Run(name, func(t *testing.T) {
			dir := t.TempDir()
			if err := os.WriteFile(filepath.Join(dir, "tables.list"), []byte("a.ref\n"), 0o644); err != nil {
				t.Fatal(err)
			}
			if err := os.WriteFile(filepath.Join(dir, "a.ref"), table(rec), 0o644); err != nil {
				t.Fatal(err)
			}
			defer func() {
				if r := recover(); r != nil {
					t.Fatalf("reftableHolds panicked: %v", fmt.Sprint(r))
				}
			}()
			holds, ok := reftableHolds(dir, Ref)
			if ok {
				t.Fatalf("a malformed table reads as known: holds %v", holds)
			}
		})
	}
}
