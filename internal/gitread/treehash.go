package gitread

import (
	"crypto/sha1" //nolint:gosec // git's own object id, never a security check
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
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

// blobHash is the sha1 of the git blob object that holds b.
func blobHash(b []byte) string {
	h := sha1.New() //nolint:gosec
	h.Write([]byte("blob " + strconv.Itoa(len(b)) + "\x00"))
	h.Write(b)
	return hex.EncodeToString(h.Sum(nil))
}

// diskBlob returns the tree entry of the file or symlink at path, as git
// would store it with no filter: mode 100755 when the owner execute bit is set,
// otherwise 100644; a symlink is mode 120000 and holds its target. Anything
// else, a socket or a device, is a refusal.
func diskBlob(path, name string, info fs.FileInfo) (Entry, error) {
	switch m := info.Mode(); {
	case m&fs.ModeSymlink != 0:
		target, err := os.Readlink(path)
		if err != nil {
			return Entry{}, herr.Env("read link %s: %v", path, err)
		}
		return Entry{Mode: "120000", Type: "blob", Hash: blobHash([]byte(target)), Name: name}, nil
	case m.IsRegular():
		b, err := os.ReadFile(path)
		if err != nil {
			return Entry{}, herr.Env("read %s: %v", path, err)
		}
		mode := "100644"
		if m.Perm()&0o100 != 0 {
			mode = "100755"
		}
		return Entry{Mode: mode, Type: "blob", Hash: blobHash(b), Name: name}, nil
	}
	return Entry{}, herr.Refused("%s is not a regular file, a symlink or a folder", path)
}

// DirTreeHash returns the git tree hash of the folder at path, as git would
// hash its bytes with no filter. An empty folder is skipped, because git
// stores none. hyp writes no object.
func DirTreeHash(path string) (string, error) {
	h, _, err := dirTree(path)
	return h, err
}

// dirTree returns the tree hash of the folder at path and whether it holds
// anything git can store.
func dirTree(path string) (string, bool, error) {
	des, err := os.ReadDir(path)
	if err != nil {
		return "", false, herr.Refused("read folder %s: %v", path, err)
	}
	var entries []Entry
	for _, de := range des {
		child := filepath.Join(path, de.Name())
		info, err := os.Lstat(child)
		if err != nil {
			return "", false, herr.Env("read %s: %v", child, err)
		}
		if info.IsDir() {
			h, any, err := dirTree(child)
			if err != nil {
				return "", false, err
			}
			if any {
				entries = append(entries, Entry{Mode: "40000", Type: "tree", Hash: h, Name: de.Name()})
			}
			continue
		}
		e, err := diskBlob(child, de.Name(), info)
		if err != nil {
			return "", false, err
		}
		entries = append(entries, e)
	}
	if len(entries) == 0 {
		return TreeObjectHash(nil), false, nil
	}
	return TreeObjectHash(entries), true, nil
}

// diskFiles lists every file and symlink under root, by slash path.
func diskFiles(root string) (map[string]Entry, error) {
	out := map[string]Entry{}
	err := filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err != nil {
			return herr.Env("read %s: %v", p, err)
		}
		if d.IsDir() {
			return nil
		}
		info, err := os.Lstat(p)
		if err != nil {
			return herr.Env("read %s: %v", p, err)
		}
		rel, err := filepath.Rel(root, p)
		if err != nil {
			return herr.Env("read %s: %v", p, err)
		}
		e, err := diskBlob(p, rel, info)
		if err != nil {
			return err
		}
		out[filepath.ToSlash(rel)] = e
		return nil
	})
	return out, err
}

// FirstDifference names the first path, in byte order, at which the folder
// at diskPath differs from the tree treeHash of the repository at dir, by
// mode or bytes; it returns "" when they are equal.
func FirstDifference(dir, diskPath, treeHash string) (string, error) {
	disk, err := diskFiles(diskPath)
	if err != nil {
		return "", err
	}
	files, err := listFiles(dir, treeHash)
	if err != nil {
		return "", err
	}
	committed := make(map[string]treeFile, len(files))
	paths := make([]string, 0, len(files)+len(disk))
	for _, f := range files {
		committed[f.path] = f
		paths = append(paths, f.path)
	}
	for p := range disk {
		if _, ok := committed[p]; !ok {
			paths = append(paths, p)
		}
	}
	sort.Strings(paths)
	for _, p := range paths {
		d, onDisk := disk[p]
		c, inGit := committed[p]
		if onDisk != inGit || (onDisk && (d.Mode != c.mode || d.Hash != c.hash)) {
			return p, nil
		}
	}
	return "", nil
}
