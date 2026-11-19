package testutil

import "testing"

func TestSnapshotSeesNewCommit(t *testing.T) {
	r := NewRepo(t)
	if r.Dir == "" {
		t.Fatal("NewRepo returned no directory")
	}
	before := Snapshot(t, r.Dir)
	if before == "" {
		t.Fatal("empty snapshot")
	}
	r.Commit(t, map[string]string{"a.txt": "a"})
	if Snapshot(t, r.Dir) == before {
		t.Fatal("snapshot did not change after a commit")
	}
}

func TestSnapshotSeesConfigChange(t *testing.T) {
	r := NewRepo(t)
	before := Snapshot(t, r.Dir)
	if before == "" {
		t.Fatal("empty snapshot")
	}
	git(t, r.Dir, "config", "x.y", "z")
	if Snapshot(t, r.Dir) == before {
		t.Fatal("snapshot did not change after a config change")
	}
}

func TestCommitHashIsStable(t *testing.T) {
	files := map[string]string{"a.txt": "a", "d/b.txt": "b"}
	a := NewRepo(t).Commit(t, files)
	b := NewRepo(t).Commit(t, files)
	if a == "" || a != b {
		t.Fatalf("commit hashes differ or are empty: %q %q", a, b)
	}
}
