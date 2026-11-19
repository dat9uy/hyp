package gitread

import (
	"bytes"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// Entry is one entry of a git tree as `git ls-tree` prints it.
type Entry struct {
	Mode string // as git prints it, for example 040000 or 100644
	Type string // blob, tree or commit
	Hash string // 40 hex digits
	Name string
}

// RootEntries lists the entries of commit's root tree.
func RootEntries(dir, commit string) ([]Entry, error) {
	out, err := Git(dir, "ls-tree", "-z", commit)
	if err != nil {
		return nil, err
	}
	var entries []Entry
	for _, rec := range bytes.Split(out, []byte{0}) {
		if len(rec) == 0 {
			continue
		}
		meta, name, ok := strings.Cut(string(rec), "\t")
		f := strings.Fields(meta)
		if !ok || len(f) != 3 {
			return nil, herr.Env("git ls-tree %s: cannot parse %q", commit, rec)
		}
		entries = append(entries, Entry{Mode: f[0], Type: f[1], Hash: f[2], Name: name})
	}
	return entries, nil
}
