package gitread

import (
	"bytes"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
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
			return nil, fault.EnvFault(fault.LsTreeParse, commit, string(rec))
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

// OnNoRef reports whether commit is a commit this clone holds that no branch
// or tag holds: no ref under refs/heads, refs/remotes or refs/tags contains
// it. Such a commit is kept only until git gc, and a clone of the branches
// and tags lacks it. A commit this clone lacks is not reported: there is
// nothing here to warn about.
func OnNoRef(dir, commit string) (bool, error) {
	if ok, err := HasCommit(dir, commit); err != nil || !ok {
		return false, err
	}
	out, err := Git(dir, "for-each-ref", "--count=1", "--format=%(refname)", "--contains", commit,
		"refs/heads", "refs/remotes", "refs/tags")
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(string(out)) == "", nil
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
	return false, fault.EnvFault(fault.GitExit, []string{"merge-base", "--is-ancestor", a, b}, code, strings.TrimSpace(stderr))
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
			return nil, fault.EnvFault(fault.LsTreeParse, treeish, string(rec))
		}
		if f[0] == "160000" {
			return nil, fault.Refuse(fault.NoGuard, fault.Submodule, treeish, name)
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
		return "", fault.Refuse(fault.NoGuard, fault.EntryLeaves, name)
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
		return fault.EnvFault(fault.Extract, treeish, err)
	}
	for _, f := range files {
		body, rest, err := nextBlob(data, f.hash)
		if err != nil {
			return fault.EnvFault(fault.CatFileBatch, treeish, err)
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
				return fault.EnvFault(fault.Extract, f.path, err)
			}
			continue
		}
		perm := os.FileMode(0o644)
		if f.mode == "100755" {
			perm = 0o755
		}
		if err := writeNew(path, body, perm); err != nil {
			return fault.EnvFault(fault.Extract, f.path, err)
		}
		// The umask must not change what the tree says.
		if err := os.Chmod(path, perm); err != nil {
			return fault.EnvFault(fault.Extract, f.path, err)
		}
	}
	return nil
}

// nextBlob reads one "<hash> blob <size>\n<bytes>\n" record of cat-file
// --batch from data and returns its bytes and what follows it.
func nextBlob(data []byte, want string) (body, rest []byte, err error) {
	nl := bytes.IndexByte(data, '\n')
	if nl < 0 {
		return nil, nil, fault.EnvFault(fault.BatchEnds, want)
	}
	f := strings.Fields(string(data[:nl]))
	if len(f) == 2 && f[1] == "missing" {
		return nil, nil, fault.EnvFault(fault.BatchMissing, want)
	}
	if len(f) != 3 || f[0] != want || f[1] != "blob" {
		return nil, nil, fault.EnvFault(fault.BatchRecord, string(data[:nl]), want)
	}
	size, err := strconv.Atoi(f[2])
	if err != nil || size < 0 || nl+1+size+1 > len(data) || data[nl+1+size] != '\n' {
		return nil, nil, fault.EnvFault(fault.BatchSize, string(data[:nl]))
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
		return fault.EnvFault(fault.Extract, dir, err)
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
				return fault.EnvFault(fault.Extract, cur, err)
			}
		case err != nil:
			return fault.EnvFault(fault.Extract, cur, err)
		case !fi.IsDir():
			return fault.EnvFault(fault.NotRealDir, cur)
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

// Ancestry is the repository at Dir as the tables read it: commit ancestry
// and the presence of a commit.
type Ancestry struct{ Dir string }

// IsAncestor reports whether a is an ancestor of b, or b itself.
func (a Ancestry) IsAncestor(x, y string) (bool, error) { return IsAncestor(a.Dir, x, y) }

// HasCommit reports whether the clone holds the commit.
func (a Ancestry) HasCommit(commit string) (bool, error) { return HasCommit(a.Dir, commit) }

// OwnPath is a path the runner's own change touches and the commit that wrote
// it.
type OwnPath struct {
	Path   string
	Commit string
}

// OwnChange returns the paths of the runner's own change of commit: what the
// runner wrote, not what it took in from main.
//
// The base is found by walking first parents from commit, commit excluded, to
// the first commit that is h0 or a merge commit. The paths are the union of
// each non-merge commit's change against its parent on that walk, plus, when
// commit is itself a merge, the paths that differ from every parent. It never
// diffs from h0 when a merge lies between.
func OwnChange(dir, commit, h0 string) ([]OwnPath, error) {
	if commit == h0 {
		return nil, nil
	}
	var out []OwnPath
	seen := map[string]bool{}
	add := func(c string, paths []string) {
		for _, p := range paths {
			if !seen[p] {
				seen[p] = true
				out = append(out, OwnPath{Path: p, Commit: c})
			}
		}
	}
	cur := commit
	for {
		parents, err := parentsOf(dir, cur)
		if err != nil {
			return nil, err
		}
		args := []string{"diff-tree", "-r", "--name-only", "-z", cur}
		if len(parents) > 1 {
			args = []string{"diff-tree", "-r", "-c", "--name-only", "-z", cur}
		}
		raw, err := Git(dir, args...)
		if err != nil {
			return nil, err
		}
		add(cur, diffPaths(raw, cur))
		if len(parents) == 0 {
			return out, nil
		}
		next := parents[0]
		if next == h0 {
			return out, nil
		}
		nextParents, err := parentsOf(dir, next)
		if err != nil {
			return nil, err
		}
		if len(nextParents) > 1 {
			return out, nil
		}
		cur = next
	}
}

// diffPaths splits the NUL-separated output of diff-tree --name-only -z; a
// commit argument makes the first field the commit's own id.
func diffPaths(raw []byte, commit string) []string {
	var paths []string
	for i, f := range bytes.Split(raw, []byte{0}) {
		if len(f) == 0 || (i == 0 && string(f) == commit) {
			continue
		}
		paths = append(paths, string(f))
	}
	return paths
}

// parentsOf reads the parents of commit with cat-file -p.
func parentsOf(dir, commit string) ([]string, error) {
	out, err := Git(dir, "cat-file", "-p", commit)
	if err != nil {
		return nil, err
	}
	var parents []string
	for _, line := range strings.Split(string(out), "\n") {
		if line == "" {
			break // the header ends at the first blank line
		}
		if p, ok := strings.CutPrefix(line, "parent "); ok {
			parents = append(parents, p)
		}
	}
	return parents, nil
}
