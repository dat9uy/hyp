package gitread

import (
	"crypto/sha1" //nolint:gosec // git's own object id, never a security check
	"encoding/hex"
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// excludedFromRow are the root entries that make no part of a commit's
// subject: the Goal folders and the store.
var excludedFromRow = map[string]bool{"goals": true, ".tree": true}

// TreeObjectHash returns the sha1 of the git tree object that holds entries.
// Entries are sorted by name, a directory as name + "/", the way git sorts
// them; the mode of a directory is written 40000. hyp writes no object.
func TreeObjectHash(entries []Entry) string {
	sorted := append([]Entry(nil), entries...)
	key := func(e Entry) string {
		if e.Type == "tree" {
			return e.Name + "/"
		}
		return e.Name
	}
	sort.Slice(sorted, func(i, j int) bool { return key(sorted[i]) < key(sorted[j]) })
	var body strings.Builder
	for _, e := range sorted {
		mode := strings.TrimLeft(e.Mode, "0")
		raw, err := hex.DecodeString(e.Hash)
		if err != nil || len(raw) != sha1.Size {
			// A hash that is not a sha1 cannot name an entry of a tree hyp
			// accepts (the object format is checked before any read).
			raw = make([]byte, sha1.Size)
		}
		fmt.Fprintf(&body, "%s %s\x00", mode, e.Name)
		body.Write(raw)
	}
	h := sha1.New() //nolint:gosec
	h.Write([]byte("tree " + strconv.Itoa(body.Len()) + "\x00"))
	h.Write([]byte(body.String()))
	return hex.EncodeToString(h.Sum(nil))
}

// RowTree returns the row tree of commit: the tree hash of its root without
// the entries goals and .tree. Two commits that differ only under those two
// folders have one row tree.
func RowTree(dir, commit string) (string, error) {
	entries, err := RootEntries(dir, commit)
	if err != nil {
		return "", err
	}
	var kept []Entry
	for _, e := range entries {
		if e.Mode == "160000" {
			return "", herr.Refused("%s: %s is a submodule (mode 160000)", commit, e.Name)
		}
		if !excludedFromRow[e.Name] {
			kept = append(kept, e)
		}
	}
	return TreeObjectHash(kept), nil
}
