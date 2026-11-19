package gitread

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
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

// CommitOf returns the full sha of the commit rev names.
func CommitOf(dir, rev string) (string, error) {
	out, err := Git(dir, "rev-parse", "--verify", rev+"^{commit}")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

// HasCommit reports whether rev names a commit this clone holds.
func HasCommit(dir, rev string) (bool, error) {
	code, _, err := GitExit(dir, "cat-file", "-e", rev+"^{commit}")
	if err != nil {
		return false, err
	}
	return code == 0, nil
}

// IsAncestor reports whether a is an ancestor of b, or b itself. A commit
// the clone lacks is an environment error.
func IsAncestor(dir, a, b string) (bool, error) {
	code, _, stderr, err := run(dir, nil, []string{"merge-base", "--is-ancestor", a, b})
	if err != nil {
		return false, err
	}
	switch code {
	case 0:
		return true, nil
	case 1:
		return false, nil
	}
	return false, herr.Env("git merge-base --is-ancestor %s %s: exit %d: %s", a, b, code, strings.TrimSpace(stderr))
}

// treeFile is one blob entry of ls-tree -r.
type treeFile struct {
	mode, hash, path string
}

// listFiles returns the entries of treeish, recursively, in git's order. A
// submodule entry (mode 160000) is a refusal naming its path.
func listFiles(dir, treeish string) ([]treeFile, error) {
	out, err := Git(dir, "ls-tree", "-r", "-z", treeish)
	if err != nil {
		return nil, err
	}
	var files []treeFile
	for _, rec := range bytes.Split(out, []byte{0}) {
		if len(rec) == 0 {
			continue
		}
		meta, name, ok := strings.Cut(string(rec), "\t")
		f := strings.Fields(meta)
		if !ok || len(f) != 3 {
			return nil, herr.Env("git ls-tree %s: cannot parse %q", treeish, rec)
		}
		if f[0] == "160000" {
			return nil, herr.Refused("%s: %s is a submodule (mode 160000)", treeish, name)
		}
		files = append(files, treeFile{mode: f[0], hash: f[2], path: name})
	}
	return files, nil
}

// entryPath returns where the tree entry name lands under dest. A name that
// would leave dest is a refusal.
func entryPath(dest, name string) (string, error) {
	clean := filepath.Clean(filepath.FromSlash(name))
	if name == "" || filepath.IsAbs(clean) || clean == ".." || strings.HasPrefix(clean, ".."+string(filepath.Separator)) {
		return "", herr.Refused("tree entry %q leaves the destination", name)
	}
	return filepath.Join(dest, clean), nil
}

// ExtractTree writes the files of treeish (a commit or a tree) under dest.
// It does not use git archive, whose export-ignore and export-subst
// attributes change the bytes: it lists the tree with ls-tree -r -z and reads
// every blob with one cat-file --batch process. Mode 100755 is written 0755,
// mode 120000 as a symlink, anything else 0644.
func ExtractTree(dir, treeish, dest string) error {
	files, err := listFiles(dir, treeish)
	if err != nil {
		return err
	}
	var req bytes.Buffer
	for _, f := range files {
		req.WriteString(f.hash + "\n")
	}
	var data []byte
	if len(files) > 0 {
		if data, err = GitStdin(dir, req.Bytes(), "cat-file", "--batch"); err != nil {
			return err
		}
	}
	if err := os.MkdirAll(dest, 0o755); err != nil {
		return herr.Env("extract %s: %v", treeish, err)
	}
	for _, f := range files {
		body, rest, err := nextBlob(data, f.hash)
		if err != nil {
			return herr.Env("git cat-file --batch %s: %v", treeish, err)
		}
		data = rest
		path, err := entryPath(dest, f.path)
		if err != nil {
			return err
		}
		if err := ensureDirs(dest, filepath.Dir(path)); err != nil {
			return err
		}
		if f.mode == "120000" {
			if err := os.Symlink(string(body), path); err != nil {
				return herr.Env("extract %s: %v", f.path, err)
			}
			continue
		}
		perm := os.FileMode(0o644)
		if f.mode == "100755" {
			perm = 0o755
		}
		if err := writeNew(path, body, perm); err != nil {
			return herr.Env("extract %s: %v", f.path, err)
		}
		// The umask must not change what the tree says.
		if err := os.Chmod(path, perm); err != nil {
			return herr.Env("extract %s: %v", f.path, err)
		}
	}
	return nil
}

// nextBlob reads one "<hash> blob <size>\n<bytes>\n" record of cat-file
// --batch from data and returns its bytes and what follows it.
func nextBlob(data []byte, want string) (body, rest []byte, err error) {
	nl := bytes.IndexByte(data, '\n')
	if nl < 0 {
		return nil, nil, fmt.Errorf("output ends before the object %s", want)
	}
	f := strings.Fields(string(data[:nl]))
	if len(f) == 2 && f[1] == "missing" {
		return nil, nil, fmt.Errorf("object %s is missing from this clone", want)
	}
	if len(f) != 3 || f[0] != want || f[1] != "blob" {
		return nil, nil, fmt.Errorf("unexpected record %q for %s", data[:nl], want)
	}
	size, err := strconv.Atoi(f[2])
	if err != nil || size < 0 || nl+1+size+1 > len(data) || data[nl+1+size] != '\n' {
		return nil, nil, fmt.Errorf("bad size in record %q", data[:nl])
	}
	return data[nl+1 : nl+1+size], data[nl+1+size+1:], nil
}

// SubtreeHash returns the hash of the entry name directly under the tree
// treeish, or "" when treeish has no such entry.
func SubtreeHash(dir, treeish, name string) (string, error) {
	entries, err := RootEntries(dir, treeish)
	if err != nil {
		return "", err
	}
	for _, e := range entries {
		if e.Name == name && e.Type == "tree" {
			return e.Hash, nil
		}
	}
	return "", nil
}

// ensureDirs makes every folder from dest down to dir. Each component that
// exists must be a real directory: a symlink, or a file of the same name
// that another tree entry made, would carry a later write outside dest (on a
// case-insensitive file system the entry a/b reaches through a symlink A).
func ensureDirs(dest, dir string) error {
	rel, err := filepath.Rel(dest, dir)
	if err != nil {
		return herr.Env("extract %s: %v", dir, err)
	}
	cur := dest
	if rel == "." {
		return nil
	}
	for _, part := range strings.Split(rel, string(filepath.Separator)) {
		cur = filepath.Join(cur, part)
		fi, err := os.Lstat(cur)
		switch {
		case os.IsNotExist(err):
			if err := os.Mkdir(cur, 0o755); err != nil {
				return herr.Env("extract %s: %v", cur, err)
			}
		case err != nil:
			return herr.Env("extract %s: %v", cur, err)
		case !fi.IsDir():
			return herr.Env("extract %s: a parent is not a real directory (symlink or file already written)", cur)
		}
	}
	return nil
}

// writeNew creates path with body and fails when anything is there already,
// so a case-colliding entry cannot overwrite another.
func writeNew(path string, body []byte, perm os.FileMode) error {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_EXCL, perm)
	if err != nil {
		return err
	}
	if _, err := f.Write(body); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}
