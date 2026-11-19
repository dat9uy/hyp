package gitread

import (
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
