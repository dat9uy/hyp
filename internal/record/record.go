// Package record owns where hyp keeps each store's events and the files they
// name. In a record repository, a git repository that holds the hidden ref
// Ref, they are commits on that ref, which no branch holds, so no branch
// act, clean or worktree removal touches them. Anywhere else, a sandbox
// clone included, they are the work-tree folder .tree/ beside the store, as
// before the record existed.
//
// It is the one package that names the store folder or git's own, and with
// package sandbox the one that runs git.
package record

import (
	"errors"
	"os"
	"path"
	"path/filepath"
	"strings"
)

// Names of the store folder and of git's, and the hidden ref.
const (
	TreeDir = ".tree"
	GitName = ".git"
	// Ref is the hidden ref whose commits hold the record. A plain git
	// fetch or clone does not bring it; hyp sync does.
	Ref = "refs/hyp/record"
	// EvidenceDir is the slash prefix of an evidence pointer, relative to
	// the store's folder.
	EvidenceDir = TreeDir + "/evidence/"
	// EventsDirRel is the slash path of a store's events folder, relative
	// to the store's folder.
	EventsDirRel = TreeDir + "/events"
	// GitAttributes is the content of the store's .gitattributes: git
	// merges an events file of two branches by keeping every line of both.
	GitAttributes = "events/*.jsonl merge=union\n"
	// GitIgnore is the content of the store's .gitignore. Git ignores every
	// file under evidence/, since the record keeps each file an event
	// names; index.sqlite is named because older hyp binaries write one.
	GitIgnore = "index.sqlite\nindex.sqlite-*\nevidence/\n"
	// evidenceIgnore is the line of GitIgnore that ignores the evidence.
	evidenceIgnore = "evidence/"
)

// ErrNoStore means store discovery found no store.
var ErrNoStore = errors.New("no .tree store found; run 'hyp init'")

// Dir returns the store folder under root.
func Dir(root string) string { return filepath.Join(root, TreeDir) }

// EventsDir returns the events folder of the store at root.
func EventsDir(root string) string { return filepath.Join(root, TreeDir, "events") }

// EvidenceRoot returns the evidence folder of the store at root.
func EvidenceRoot(root string) string { return filepath.Join(root, TreeDir, "evidence") }

// IsGitName reports whether a path component is git's folder or file.
func IsGitName(name string) bool { return name == GitName }

// IsStoreDirName reports whether a path component is a store folder.
func IsStoreDirName(name string) bool { return name == TreeDir }

// UnderStoreDir reports whether a slash path relative to a store's folder
// lies under its store folder.
func UnderStoreDir(rel string) bool { return strings.HasPrefix(rel, TreeDir+"/") }

// FindTop returns the repository root at or above start: the first
// directory that holds .git, a directory or, in a worktree, a file.
func FindTop(start string) (string, bool) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return "", false
	}
	for d := abs; ; d = filepath.Dir(d) {
		if _, err := os.Lstat(filepath.Join(d, GitName)); err == nil {
			return d, true
		}
		if filepath.Dir(d) == d {
			return "", false
		}
	}
}

// Store is one store found by discovery.
type Store struct {
	// Dir is the store's folder in the work tree: the directory that holds,
	// or would hold, .tree/.
	Dir string
	// Name is the store's folder as a slash path from the repository root,
	// "." for the root; "" outside a repository.
	Name string
	// Repo is the repository that holds Dir, nil outside one.
	Repo *Repo
	// InRecord reports whether the record holds this store.
	InRecord bool
	tip      string
	entries  map[string]entry
}

// RecordMode reports whether the store's writes go to the record.
func (s *Store) RecordMode() bool { return s.Repo != nil && s.Repo.Record }

// Locate runs store discovery from start. It walks up to the repository
// root, the first directory that holds .git, and stops at the first
// directory that the record holds as a store or whose work tree holds
// .tree/events. With no .git at or above start it walks to the filesystem
// root and reads the work tree only.
func Locate(start string) (*Store, error) {
	abs, err := filepath.Abs(start)
	if err != nil {
		return nil, err
	}
	top, inRepo := FindTop(abs)
	var repo *Repo
	if inRepo {
		if repo, err = openRepo(top); err != nil {
			return nil, err
		}
	}
	s := &Store{Repo: repo}
	if repo != nil && repo.Record {
		if err := s.readTip(); err != nil {
			return nil, err
		}
	}
	for dir := abs; ; dir = filepath.Dir(dir) {
		name := ""
		if inRepo {
			name = slashRel(top, dir)
		}
		if s.entries != nil && s.hasStore(name) {
			s.Dir, s.Name, s.InRecord = dir, name, true
			return s, nil
		}
		if fi, err := os.Stat(EventsDir(dir)); err == nil && fi.IsDir() {
			s.Dir, s.Name = dir, name
			return s, nil
		}
		if inRepo && dir == top {
			return nil, ErrNoStore
		}
		if filepath.Dir(dir) == dir {
			return nil, ErrNoStore
		}
	}
}

// At returns the store whose folder is dir, whether or not discovery from
// dir would stop there, or ErrNoStore when neither the record nor the work
// tree holds a store at dir.
func At(dir string) (*Store, error) {
	s, err := Locate(dir)
	if err != nil {
		return nil, err
	}
	abs, _ := filepath.Abs(dir)
	if s.Dir != abs {
		return nil, ErrNoStore
	}
	return s, nil
}

func slashRel(top, dir string) string {
	rel, err := filepath.Rel(top, dir)
	if err != nil {
		return "."
	}
	return filepath.ToSlash(rel)
}

// recordPath is the record path of a slash path relative to the store's
// folder.
func (s *Store) recordPath(rel string) string { return joinName(s.Name, rel) }

func joinName(name, rel string) string {
	if name == "." || name == "" {
		return rel
	}
	return path.Join(name, rel)
}

// hasStore reports whether the record holds a store named name: a tree at
// <name>/.tree/events.
func (s *Store) hasStore(name string) bool {
	e, ok := s.entries[joinName(name, EventsDirRel)]
	return ok && e.typ == "tree"
}

// Root is the directory that every read of the store reads below: the
// record's copy of the store when the record holds it, its work-tree folder
// otherwise.
func (s *Store) Root() (string, error) {
	if !s.InRecord {
		return s.Dir, nil
	}
	snap, err := s.Repo.snapshot(s.tip, s.Name, s.entries)
	if err != nil {
		return "", err
	}
	return filepath.Join(snap, filepath.FromSlash(s.Name)), nil
}

// HoldsFile reports whether the record holds the file at slash path rel
// relative to the store's folder.
func (s *Store) HoldsFile(rel string) bool {
	if s.entries == nil {
		return false
	}
	e, ok := s.entries[s.recordPath(rel)]
	return ok && e.typ == "blob"
}

// ReadFile returns the bytes of the file at slash path rel relative to the
// store's folder: the record's copy when the record holds the store, the
// work tree's otherwise.
func (s *Store) ReadFile(rel string) ([]byte, error) {
	root, err := s.Root()
	if err != nil {
		return nil, err
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
}

// WorkTreeEventFiles returns the events files of the store's work-tree
// folder, keyed by actor, and nil when it has none.
func (s *Store) WorkTreeEventFiles() (map[string][]byte, error) {
	return readEventFiles(s.Dir)
}

// EventFiles returns the store's events files as every read sees them,
// keyed by actor (the file's name without .jsonl).
func (s *Store) EventFiles() (map[string][]byte, error) {
	root, err := s.Root()
	if err != nil {
		return nil, err
	}
	return readEventFiles(root)
}

func readEventFiles(root string) (map[string][]byte, error) {
	files, err := filepath.Glob(filepath.Join(EventsDir(root), "*.jsonl"))
	if err != nil {
		return nil, err
	}
	out := map[string][]byte{}
	for _, f := range files {
		data, err := os.ReadFile(f)
		if err != nil {
			return nil, err
		}
		out[strings.TrimSuffix(filepath.Base(f), ".jsonl")] = data
	}
	return out, nil
}

// Init writes the store folder in dir: events/, evidence/, .gitignore and
// .gitattributes. In a git repository with no record and no sandbox mark it
// also writes the empty record, which makes it a record repository.
func Init(dir string) error {
	for _, sub := range []string{"events", "evidence"} {
		if err := os.MkdirAll(filepath.Join(Dir(dir), sub), 0o755); err != nil {
			return err
		}
	}
	if err := os.WriteFile(filepath.Join(Dir(dir), ".gitattributes"), []byte(GitAttributes), 0o644); err != nil {
		return err
	}
	if err := os.WriteFile(filepath.Join(Dir(dir), ".gitignore"), []byte(GitIgnore), 0o644); err != nil {
		return err
	}
	top, ok := FindTop(dir)
	if !ok {
		return nil
	}
	repo, err := openRepo(top)
	if err != nil {
		return err
	}
	return repo.ensureRecord()
}

// IgnoreEvidence adds the line that makes git ignore every file under the
// store's evidence folder to the .gitignore of the store at dir, keeping
// its other lines.
func IgnoreEvidence(dir string) error {
	p := filepath.Join(Dir(dir), ".gitignore")
	data, err := os.ReadFile(p)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, l := range strings.Split(string(data), "\n") {
		if strings.TrimSpace(l) == evidenceIgnore {
			return nil
		}
	}
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		data = append(data, '\n')
	}
	data = append(data, evidenceIgnore+"\n"...)
	if err := os.MkdirAll(Dir(dir), 0o755); err != nil {
		return err
	}
	return os.WriteFile(p, data, 0o644)
}
