package record

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"

	"github.com/dat9uy/hyp/internal/herr"
)

// File is one file a write adds to the record: its slash path relative to
// the repository root, and where its bytes are, a file on disk (Src) or a
// blob the repository already holds (Blob).
type File struct {
	Path string
	Src  string
	Blob string
	Exec bool
}

// StoreFile is the File of the file at slash path rel relative to the
// store's folder, read from src.
func (s *Store) StoreFile(rel, src string) File {
	fi, err := os.Stat(src)
	return File{Path: s.recordPath(rel), Src: src, Exec: err == nil && fi.Mode()&0o111 != 0}
}

// RepoFile is the File of the file at absolute path abs in the repository.
// ok is false when abs is outside the repository's work tree.
func (s *Store) RepoFile(abs string) (File, bool) {
	if s.Repo == nil {
		return File{}, false
	}
	return FileUnder(s.Repo.Top, abs)
}

// FileUnder is the File of the file at absolute path abs below top, the
// root of a repository or of a copy of one; ok is false outside top.
func FileUnder(top, abs string) (File, bool) {
	rel, err := filepath.Rel(top, abs)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return File{}, false
	}
	fi, err := os.Stat(abs)
	return File{Path: filepath.ToSlash(rel), Src: abs, Exec: err == nil && fi.Mode()&0o111 != 0}, true
}

// Lock takes the store's write lock and returns its release. In a record
// repository it is one lock for the whole clone, in its common git
// directory, so writes from every worktree take turns; it rereads the
// record once the lock is held. Elsewhere it locks the store's events
// folder.
func (s *Store) Lock() (func(), error) {
	var f *os.File
	var err error
	if s.RecordMode() {
		if err := os.MkdirAll(s.Repo.hypDir(), 0o755); err != nil {
			return nil, err
		}
		if f, err = os.OpenFile(s.Repo.hypDir("record.lock"), os.O_CREATE|os.O_RDWR, 0o644); err != nil {
			return nil, err
		}
	} else if f, err = os.Open(EventsDir(s.Dir)); err != nil {
		return nil, err
	}
	if err := syscall.Flock(int(f.Fd()), syscall.LOCK_EX); err != nil {
		f.Close()
		return nil, err
	}
	release := func() {
		syscall.Flock(int(f.Fd()), syscall.LOCK_UN)
		f.Close()
	}
	if s.RecordMode() {
		if err := s.readTip(); err != nil {
			release()
			return nil, err
		}
		s.InRecord = s.hasStore(s.Name)
	}
	return release, nil
}

// afterStep runs after each step of a record write; a test sets it to stop
// the process there.
var afterStep = func(step string) {}

// Append adds lines, keyed by actor, to the store's events files, with the
// files the events name and the commits their subjects name. In a record
// repository it is one commit on Ref, which only its last step, a
// compare-and-swap of Ref, makes visible: a write stopped before that step
// leaves the record as it was. A file whose path the record already holds
// with other bytes is refused, since a file an event named never changes.
// Elsewhere it appends to the work-tree events files and syncs. The caller
// holds Lock.
func (s *Store) Append(lines map[string][][]byte, files []File, commits []string) error {
	if !s.RecordMode() {
		return s.appendWorkTree(lines)
	}
	r := s.Repo
	idx, err := os.CreateTemp(r.hypDir(), "index-*")
	if err != nil {
		return err
	}
	idx.Close()
	defer os.Remove(idx.Name())
	env := []string{"GIT_INDEX_FILE=" + idx.Name()}
	if s.tip != "" {
		if _, err := gitOut(r.Top, env, nil, "read-tree", s.tip); err != nil {
			return err
		}
	} else if _, err := gitOut(r.Top, env, nil, "read-tree", "--empty"); err != nil {
		return err
	}
	var info bytes.Buffer
	actors := make([]string, 0, len(lines))
	for a := range lines {
		actors = append(actors, a)
	}
	slices.Sort(actors)
	for _, actor := range actors {
		p := s.recordPath(EventsDirRel + "/" + actor + ".jsonl")
		var data []byte
		if e, ok := s.entries[p]; ok {
			if err := r.readBlobs([]string{e.sha}, func(_ string, b []byte) error { data = b; return nil }); err != nil {
				return err
			}
		}
		if len(data) > 0 && data[len(data)-1] != '\n' {
			data = append(data, '\n')
		}
		for _, l := range lines[actor] {
			data = append(append(data, l...), '\n')
		}
		sha, err := gitOut(r.Top, nil, data, "hash-object", "-w", "--stdin")
		if err != nil {
			return err
		}
		fmt.Fprintf(&info, "100644 %s\t%s\n", sha, p)
	}
	seen := map[string]bool{}
	for _, f := range files {
		if seen[f.Path] {
			continue
		}
		seen[f.Path] = true
		sha := f.Blob
		if sha == "" {
			if sha, err = gitOut(r.Top, nil, nil, "hash-object", "-w", "--no-filters", "--", f.Src); err != nil {
				return err
			}
		}
		if e, ok := s.entries[f.Path]; ok {
			if e.sha != sha {
				return herr.Refused("%s differs from the copy the record holds; a file an event names never changes, so write it under a new path", f.Path)
			}
			continue
		}
		mode := "100644"
		if f.Exec {
			mode = "100755"
		}
		fmt.Fprintf(&info, "%s %s\t%s\n", mode, sha, f.Path)
	}
	if _, err := gitOut(r.Top, env, info.Bytes(), "update-index", "--add", "--index-info"); err != nil {
		return err
	}
	afterStep("objects")
	tree, err := gitOut(r.Top, env, nil, "write-tree")
	if err != nil {
		return err
	}
	afterStep("tree")
	var parents []string
	if s.tip != "" {
		parents = append(parents, s.tip)
	}
	for _, c := range commits {
		id, ok := r.IsCommit(c)
		if !ok || slices.Contains(parents, id) || (s.tip != "" && r.isAncestor(id, s.tip)) {
			continue
		}
		parents = append(parents, id)
	}
	c, err := r.commitTree(tree, parents, "hyp: "+strings.Join(actors, " "))
	if err != nil {
		return err
	}
	afterStep("commit")
	if _, err := gitOut(r.Top, nil, nil, "update-ref", Ref, c, s.tip); err != nil {
		return err
	}
	afterStep("ref")
	return s.readTip()
}

// appendWorkTree appends each actor's lines to its work-tree events file
// and syncs it.
func (s *Store) appendWorkTree(lines map[string][][]byte) error {
	for actor, ls := range lines {
		f, err := os.OpenFile(filepath.Join(EventsDir(s.Dir), actor+".jsonl"), os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			return err
		}
		var b bytes.Buffer
		for _, l := range ls {
			b.Write(l)
			b.WriteByte('\n')
		}
		if _, err := f.Write(b.Bytes()); err != nil {
			f.Close()
			return err
		}
		if err := f.Sync(); err != nil {
			f.Close()
			return err
		}
		if err := f.Close(); err != nil {
			return err
		}
	}
	return nil
}
