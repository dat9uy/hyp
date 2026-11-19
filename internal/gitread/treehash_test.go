package gitread

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

func TestTreeObjectHashEqualsGitRoot(t *testing.T) {
	r := testutil.NewRepo(t)
	r.WriteExec(t, "bin/run", "#!/bin/sh\n")
	r.Symlink(t, "link", "a.txt")
	c := r.Commit(t, map[string]string{
		"a.txt":       "a",
		"a/b.txt":     "b",
		"a.b":         "dot",
		"goals/G/x":   "x",
		".tree/e.txt": "e",
		"z/deep/f":    "f",
	})
	entries, err := RootEntries(r.Dir, c)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) < 6 {
		t.Fatalf("root entries = %d, want at least 6: %+v", len(entries), entries)
	}
	want, err := Git(r.Dir, "rev-parse", c+"^{tree}")
	if err != nil {
		t.Fatal(err)
	}
	if got := TreeObjectHash(entries); got != strings.TrimSpace(string(want)) {
		t.Fatalf("TreeObjectHash = %s, want %s", got, want)
	}
}

func TestRowTreeIgnoresGoalsAndStore(t *testing.T) {
	r := testutil.NewRepo(t)
	c1 := r.Commit(t, map[string]string{"main.go": "package main", "goals/G/a": "1", ".tree/events/x.jsonl": "1"})
	c2 := r.Commit(t, map[string]string{"goals/G/a": "2", "goals/H/b": "3", ".tree/events/x.jsonl": "2"})
	t1, err := RowTree(r.Dir, c1)
	if err != nil {
		t.Fatal(err)
	}
	if len(t1) != 40 {
		t.Fatalf("row tree = %q, want 40 hex digits", t1)
	}
	t2, err := RowTree(r.Dir, c2)
	if err != nil {
		t.Fatal(err)
	}
	if t1 != t2 {
		t.Fatalf("a change under goals/ and .tree/ changed the row tree: %s vs %s", t1, t2)
	}
	c3 := r.Commit(t, map[string]string{"main.go": "package main // changed"})
	t3, err := RowTree(r.Dir, c3)
	if err != nil {
		t.Fatal(err)
	}
	if t3 == t1 {
		t.Fatal("a change outside goals/ and .tree/ kept the row tree")
	}
}

func TestRootEntriesNoSnapshotChange(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"a": "a"})
	before := testutil.Snapshot(t, r.Dir)
	if _, err := RowTree(r.Dir, c); err != nil {
		t.Fatal(err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("RowTree changed the repository")
	}
}

func TestDirTreeHashEqualsGit(t *testing.T) {
	r := testutil.NewRepo(t)
	r.WriteExec(t, "f/bin/run", "#!/bin/sh\n")
	r.WriteExec(t, "f/exec-only", "x")
	r.Symlink(t, "f/link", "bin/run")
	r.Symlink(t, "f/dir-link", "../f")
	c := r.Commit(t, map[string]string{
		"f/a.b":           "dot",
		"f/a/inner.txt":   "dir a",
		"f/a-b":           "dash",
		"f/a b":           "space",
		"f/\u00e9t\u00e9": "unicode",
		"f/nested/x/y/z":  "deep",
		"f/empty":         "",
		"f/crlf":          "line\r\n",
		"f/bin/data":      "\x00\x01\x02",
		"outside":         "not in the folder",
	})
	// An empty directory holds nothing git can store.
	if err := os.MkdirAll(filepath.Join(r.Dir, "f/hollow/deeper"), 0o755); err != nil {
		t.Fatal(err)
	}
	want, err := Git(r.Dir, "rev-parse", c+":f")
	if err != nil {
		t.Fatal(err)
	}
	got, err := DirTreeHash(filepath.Join(r.Dir, "f"))
	if err != nil {
		t.Fatal(err)
	}
	if got != strings.TrimSpace(string(want)) {
		t.Fatalf("DirTreeHash = %s, want git's %s", got, want)
	}
}

func TestDirTreeHashSeesModeAndBytes(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"f/a": "1"})
	base, err := DirTreeHash(filepath.Join(r.Dir, "f"))
	if err != nil {
		t.Fatal(err)
	}
	if len(base) != 40 {
		t.Fatalf("hash = %q, want 40 hex digits", base)
	}
	if err := os.Chmod(filepath.Join(r.Dir, "f/a"), 0o755); err != nil {
		t.Fatal(err)
	}
	exec, _ := DirTreeHash(filepath.Join(r.Dir, "f"))
	if exec == base {
		t.Fatal("an execute bit left the hash unchanged")
	}
	if err := os.WriteFile(filepath.Join(r.Dir, "f/a"), []byte("2"), 0o755); err != nil {
		t.Fatal(err)
	}
	if changed, _ := DirTreeHash(filepath.Join(r.Dir, "f")); changed == exec {
		t.Fatal("changed bytes left the hash unchanged")
	}
}

func TestDirTreeHashRefusesMissingAndSpecialFiles(t *testing.T) {
	if _, err := DirTreeHash(filepath.Join(t.TempDir(), "nope")); err == nil {
		t.Fatal("a missing folder was hashed")
	}
	f := filepath.Join(t.TempDir(), "file")
	if err := os.WriteFile(f, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := DirTreeHash(f); err == nil {
		t.Fatal("a file was hashed as a folder")
	}
}

func TestFirstDifferenceNamesTheFirstPath(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"f/a": "1", "f/d/b": "2", "f/d/c": "3"})
	tree, err := Git(r.Dir, "rev-parse", c+":f")
	if err != nil {
		t.Fatal(err)
	}
	th := strings.TrimSpace(string(tree))
	f := filepath.Join(r.Dir, "f")
	if d, err := FirstDifference(r.Dir, f, th); err != nil || d != "" {
		t.Fatalf("equal folders: %q, %v", d, err)
	}
	if err := os.WriteFile(filepath.Join(f, "d/c"), []byte("changed"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(f, "z-stray"), []byte("s"), 0o644); err != nil {
		t.Fatal(err)
	}
	if d, err := FirstDifference(r.Dir, f, th); err != nil || d != "d/c" {
		t.Fatalf("changed file: %q, %v; want d/c", d, err)
	}
	if err := os.Remove(filepath.Join(f, "a")); err != nil {
		t.Fatal(err)
	}
	if d, err := FirstDifference(r.Dir, f, th); err != nil || d != "a" {
		t.Fatalf("missing file: %q, %v; want a", d, err)
	}
	if err := os.WriteFile(filepath.Join(f, "a"), []byte("1"), 0o755); err != nil {
		t.Fatal(err)
	}
	if d, err := FirstDifference(r.Dir, f, th); err != nil || d != "a" {
		t.Fatalf("mode change: %q, %v; want a", d, err)
	}
}

func TestDirTreeHashWritesNoGitState(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"f/a": "1"})
	fresh := filepath.Join(r.Dir, "fresh")
	if err := os.MkdirAll(fresh, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(fresh, "only-on-disk"), []byte("never committed"), 0o644); err != nil {
		t.Fatal(err)
	}
	before := testutil.Snapshot(t, r.Dir)
	h, err := DirTreeHash(fresh)
	if err != nil || len(h) != 40 {
		t.Fatalf("hash = %q, %v", h, err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("DirTreeHash changed the repository")
	}
	if code, _, err := GitExit(r.Dir, "cat-file", "-e", h); err != nil || code == 0 {
		t.Fatalf("the tree object %s exists in the repository: code %d, %v", h, code, err)
	}
}
