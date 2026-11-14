package record

import (
	"bytes"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
)

// reftableHolds reports whether the reftable stack in the folder dir names
// ref, read without git: the newest table that has a record for ref
// decides, and a deletion record names no ref. ok is false when the stack
// cannot be read, so the caller cannot tell.
func reftableHolds(dir, ref string) (holds, ok bool) {
	list, err := os.ReadFile(filepath.Join(dir, "tables.list"))
	if err != nil {
		return false, false
	}
	names := strings.Fields(string(list))
	for i := len(names) - 1; i >= 0; i-- {
		b, err := os.ReadFile(filepath.Join(dir, names[i]))
		if err != nil {
			return false, false
		}
		typ, found, err := reftableRef(b, ref)
		if err != nil {
			return false, false
		}
		if found {
			return typ != 0, true
		}
	}
	return false, true
}

var errReftable = errors.New("reftable: malformed table")

// reftableRef scans the ref blocks of one table for ref and returns the
// value type of its record.
func reftableRef(b []byte, ref string) (typ byte, found bool, err error) {
	if len(b) < 24 || string(b[:4]) != "REFT" {
		return 0, false, errReftable
	}
	var hdr, hashLen int
	switch b[4] {
	case 1:
		hdr, hashLen = 24, 20
	case 2:
		if len(b) < 28 {
			return 0, false, errReftable
		}
		hdr = 28
		switch string(b[24:28]) {
		case "sha1":
			hashLen = 20
		case "s256":
			hashLen = 32
		default:
			return 0, false, errReftable
		}
	default:
		return 0, false, errReftable
	}
	blockStart, pos := 0, hdr
	for pos+4 <= len(b) && b[pos] == 'r' {
		blockLen := int(b[pos+1])<<16 | int(b[pos+2])<<8 | int(b[pos+3])
		end := blockStart + blockLen
		if end > len(b) || end < pos+6 {
			return 0, false, errReftable
		}
		restarts := int(binary.BigEndian.Uint16(b[end-2 : end]))
		recEnd := end - 2 - 3*restarts
		if recEnd < pos+4 {
			return 0, false, errReftable
		}
		var key []byte
		for p := pos + 4; p < recEnd; {
			prefix, n := reftableVarint(b[p:recEnd])
			if n == 0 {
				return 0, false, errReftable
			}
			p += n
			sv, n := reftableVarint(b[p:recEnd])
			if n == 0 {
				return 0, false, errReftable
			}
			p += n
			vt := byte(sv & 7)
			if prefix > uint64(len(key)) || sv>>3 > uint64(recEnd-p) {
				return 0, false, errReftable
			}
			suffix := int(sv >> 3)
			key = append(key[:prefix:prefix], b[p:p+suffix]...)
			p += suffix
			if _, n = reftableVarint(b[p:recEnd]); n == 0 {
				return 0, false, errReftable
			}
			p += n
			switch vt {
			case 0:
			case 1:
				p += hashLen
			case 2:
				p += 2 * hashLen
			case 3:
				l, n := reftableVarint(b[p:recEnd])
				if n == 0 || l > uint64(recEnd-p-n) {
					return 0, false, errReftable
				}
				p += n + int(l)
			default:
				return 0, false, errReftable
			}
			if p > recEnd {
				return 0, false, errReftable
			}
			if c := bytes.Compare(key, []byte(ref)); c == 0 {
				return vt, true, nil
			} else if c > 0 {
				return 0, false, nil
			}
		}
		// The next block starts after this one's padding, if any.
		pos = end
		for pos < len(b) && b[pos] == 0 {
			pos++
		}
		blockStart = pos
	}
	return 0, false, nil
}

// reftableVarint reads one reftable varint and returns it with the bytes it
// took, 0 when b ends inside it or it does not fit 63 bits.
func reftableVarint(b []byte) (uint64, int) {
	if len(b) == 0 {
		return 0, 0
	}
	v := uint64(b[0] & 0x7f)
	i := 0
	for b[i]&0x80 != 0 {
		i++
		if i >= len(b) || v >= 1<<56 {
			return 0, 0
		}
		v = (v+1)<<7 | uint64(b[i]&0x7f)
	}
	return v, i + 1
}
