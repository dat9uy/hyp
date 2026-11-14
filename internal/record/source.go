package record

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/dat9uy/hyp/internal/herr"
)

// ExtractAt writes to dest the files of commit or tree rev that a read of
// the store whose folder is the slash path name needs: every file but those
// under another store's folder. dest gets a .git folder, so its root is
// the repository root of every path in it.
func (r *Repo) ExtractAt(rev, name, dest string) error {
	if _, err := gitOut(r.Top, nil, nil, "rev-parse", "--verify", "--quiet", rev+"^{tree}"); err != nil {
		return herr.BadArgs("%s names no commit or tree in %s", rev, r.Top)
	}
	entries, err := r.listTree(rev)
	if err != nil {
		return err
	}
	type want struct{ path, mode string }
	bySHA := map[string][]want{}
	var shas []string
	for p, e := range entries {
		if e.typ != "blob" || !keepPath(name, p) {
			continue
		}
		if _, ok := bySHA[e.sha]; !ok {
			shas = append(shas, e.sha)
		}
		bySHA[e.sha] = append(bySHA[e.sha], want{p, e.mode})
	}
	if err := os.MkdirAll(filepath.Join(dest, GitName), 0o755); err != nil {
		return err
	}
	return r.readBlobs(shas, func(sha string, data []byte) error {
		for _, w := range bySHA[sha] {
			dst := filepath.Join(dest, filepath.FromSlash(w.path))
			if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
				return err
			}
			mode := os.FileMode(0o644)
			if w.mode == "100755" {
				mode = 0o755
			}
			if err := os.WriteFile(dst, data, mode); err != nil {
				return err
			}
		}
		return nil
	})
}

// Name is the slash path of dir from the repository root.
func (r *Repo) Name(dir string) (string, bool) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return "", false
	}
	rel, err := filepath.Rel(r.Top, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return "", false
	}
	return filepath.ToSlash(rel), true
}

// TrackedEvidence lists the files git tracks under the evidence folder of
// the store at dir, as slash paths relative to dir.
func (r *Repo) TrackedEvidence(dir string) ([]string, error) {
	out, err := gitRaw(dir, nil, nil, "ls-files", "-z", "--", EvidenceDir)
	if err != nil {
		return nil, err
	}
	var files []string
	for _, f := range bytes.Split(out, []byte{0}) {
		if len(f) > 0 {
			files = append(files, string(f))
		}
	}
	return files, nil
}

// Untrack removes each slash path relative to dir from git's index,
// leaving the file on disk.
func (r *Repo) Untrack(dir string, rels []string) error {
	for len(rels) > 0 {
		n := min(len(rels), 200)
		args := append([]string{"rm", "--cached", "--quiet", "--"}, rels[:n]...)
		if _, err := gitOut(dir, nil, nil, args...); err != nil {
			return err
		}
		rels = rels[n:]
	}
	return nil
}

// OpenRecord opens the repository whose work tree holds dir for a write
// that makes it a record repository when it is not one: hyp import. It is
// refused in a sandbox clone.
func OpenRecord(dir string) (*Repo, error) {
	repo, ok, err := OpenRepo(dir)
	if err != nil {
		return nil, err
	}
	if !ok || repo.Common == "" {
		return nil, herr.Refused("no git repository holds %s, so it can hold no record", dir)
	}
	if _, err := os.Stat(filepath.Join(repo.Common, sandboxMark)); err == nil {
		return nil, herr.Refused("%s is a sandbox clone, which keeps no record", repo.Top)
	}
	if err := repo.ensureRecord(); err != nil {
		return nil, err
	}
	return repo, nil
}

// StoreNamed returns the store the record names name in repo, read at the
// record's tip, whether or not the record holds it yet. Dir is the folder
// name names in the work tree.
func (r *Repo) StoreNamed(name string) (*Store, error) {
	s := &Store{Repo: r, Name: name, Dir: filepath.Join(r.Top, filepath.FromSlash(name))}
	if err := s.readTip(); err != nil {
		return nil, err
	}
	s.InRecord = s.hasStore(name)
	return s, nil
}
