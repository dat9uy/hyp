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

// OwnPath is a path under goals/ whose entry in a run's commit differs from
// H0's, and whether the commit lacks it.
type OwnPath struct {
	Path    string
	Deleted bool // the commit holds no file at the path that H0 holds
}

// GoalChange returns each file under goals/ whose entry in commit differs
// from its entry in h0: what a Goal's folder in commit holds that H0 does
// not. It reads the two trees, not the commits between them, so it is the
// same whichever branch brought each file.
func GoalChange(dir, commit, h0 string) ([]OwnPath, error) {
	if commit == h0 {
		return nil, nil
	}
	raw, err := Git(dir, "diff-tree", "-r", "--name-status", "-z", h0, commit)
	if err != nil {
		return nil, err
	}
	var out []OwnPath
	for _, p := range diffPaths(raw, "") {
		if strings.HasPrefix(p.path, "goals/") {
			out = append(out, OwnPath{Path: p.path, Deleted: p.deleted})
		}
	}
	return out, nil
}

// CommitTree reads a run's commit for Goal paths: its trees, and the
// commits newer than H0 below it.
type CommitTree struct {
	Dir, Commit, H0 string

	newer map[string][]string // the commits newer than H0 below Commit, with their parents
}

// TreeHash is the hash of the tree at path in the commit, or "" when the
// commit holds no tree there.
func (c *CommitTree) TreeHash(path string) (string, error) {
	return c.object(c.Commit, path, "tree")
}

// object is the id of the entry of kind at path in commit, or "" when
// there is none.
func (c *CommitTree) object(commit, path, kind string) (string, error) {
	out, err := Git(c.Dir, "ls-tree", "-z", commit, path)
	if err != nil {
		return "", err
	}
	for _, rec := range bytes.Split(out, []byte{0}) {
		meta, name, ok := strings.Cut(string(rec), "\t")
		if !ok || name != path {
			continue
		}
		f := strings.Fields(meta)
		if len(f) == 3 && f[1] == kind {
			return f[2], nil
		}
	}
	return "", nil
}

// Writer is who wrote the bytes a path holds in a run's commit.
type Writer struct {
	Commit string   // the commit that wrote them
	Old    bool     // a commit H0's history holds, which no run of the Goal wrote
	Events []string // the ids of the events whose files Commit adds under .tree/events/
}

// Writer finds the commit that wrote the bytes path holds in the commit:
// the walk from the commit goes to a parent that holds the same bytes there
// while one does, and the writer is the commit no parent of which does, or
// the first commit H0's history holds.
func (c *CommitTree) Writer(path string) (Writer, error) {
	if c.newer == nil {
		var err error
		if c.newer, err = revList(c.Dir, c.Commit, "^"+c.H0); err != nil {
			return Writer{}, err
		}
	}
	blob, err := c.object(c.Commit, path, "blob")
	if err != nil {
		return Writer{}, err
	}
	cur := c.Commit
	for {
		ps, newer := c.newer[cur]
		if !newer {
			return Writer{Commit: cur, Old: true}, nil
		}
		next := ""
		for _, p := range ps {
			b, err := c.object(p, path, "blob")
			if err != nil {
				return Writer{}, err
			}
			if b == blob {
				next = p
				break
			}
		}
		if next == "" {
			evs, err := c.addedEvents(cur, ps)
			return Writer{Commit: cur, Events: evs}, err
		}
		cur = next
	}
}

// addedEvents is the ids of the event files that commit, against its first
// parent, adds under .tree/events/: each file is named by its event's id.
func (c *CommitTree) addedEvents(commit string, parents []string) ([]string, error) {
	if len(parents) == 0 {
		return nil, nil
	}
	raw, err := Git(c.Dir, "diff-tree", "-r", "--name-status", "-z", parents[0], commit)
	if err != nil {
		return nil, err
	}
	fields := bytes.Split(raw, []byte{0})
	var ids []string
	for i := 0; i+1 < len(fields); i += 2 {
		st, p := string(fields[i]), string(fields[i+1])
		if st != "A" || !strings.HasPrefix(p, ".tree/events/") || !strings.HasSuffix(p, ".json") {
			continue
		}
		name := p[strings.LastIndex(p, "/")+1:]
		ids = append(ids, strings.TrimSuffix(name, ".json"))
	}
	return ids, nil
}

// changedPath is a path of a diff-tree --name-status output and whether the
// diff deletes it; for a merge it is deleted when every parent's diff deletes it.
type changedPath struct {
	path    string
	deleted bool
}

// diffPaths splits the NUL-separated output of diff-tree --name-status -z; a
// commit argument makes the first field the commit's own id. Each change is
// a status field, one letter per parent, then its path.
func diffPaths(raw []byte, commit string) []changedPath {
	var fields []string
	for i, f := range bytes.Split(raw, []byte{0}) {
		if len(f) == 0 || (i == 0 && string(f) == commit) {
			continue
		}
		fields = append(fields, string(f))
	}
	var paths []changedPath
	for i := 0; i+1 < len(fields); i += 2 {
		paths = append(paths, changedPath{path: fields[i+1], deleted: strings.Trim(fields[i], "D") == ""})
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
