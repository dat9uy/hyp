package testutil

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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

// TestCommitStartsNoMaintenance checks that a commit in a test repository
// leaves no background git process behind, which could write under the git
// directory after the commit returns and break a Snapshot.
func TestCommitStartsNoMaintenance(t *testing.T) {
	r := NewRepo(t)
	trace := filepath.Join(t.TempDir(), "trace.json")
	t.Setenv("GIT_TRACE2_EVENT", trace)
	r.Commit(t, map[string]string{"a.txt": "a"})
	b, err := os.ReadFile(trace)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(b), `"maintenance"`) {
		t.Fatalf("the commit started git maintenance:\n%s", b)
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
