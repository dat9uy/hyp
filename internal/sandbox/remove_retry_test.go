package sandbox

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A remove killed while it saved to OUT_DIR left only its staging
// directory, and its record says removing. Removing again to the same
// OUT_DIR clears the staging directory, saves a whole record and deletes
// the sandbox.
func TestRemoveRetriesAfterAKilledSave(t *testing.T) {
	setup(t)
	rec, box := create(t, "r1")
	rec.Removing = now()
	if err := writeJSON(filepath.Join(box, "sandbox.json"), rec); err != nil {
		t.Fatal(err)
	}
	out := filepath.Join(t.TempDir(), "out")
	staging := out + ".partial-r1"
	if err := os.MkdirAll(filepath.Join(staging, "tree"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(staging, "half"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Remove(RemoveOptions{ID: "r1", OutDir: out + "/"}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if !exists(filepath.Join(out, "sandbox.json")) || exists(filepath.Join(out, "half")) {
		t.Error("OUT_DIR is not a fresh whole save")
	}
	if exists(staging) || exists(box) {
		t.Error("the staging directory or the sandbox is left")
	}
}

// A remove killed after it renamed its save into OUT_DIR and before the
// delete left the sandbox and a whole record. Removing again to the same
// OUT_DIR names --discard, which finishes; OUT_DIR holding another
// sandbox's record is only an existing path.
func TestRemoveAfterAKilledDeleteNamesDiscard(t *testing.T) {
	setup(t)
	rec, box := create(t, "r2")
	out := filepath.Join(t.TempDir(), "out")
	saved := *rec
	saved.Removed = now()
	if err := os.MkdirAll(out, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := writeJSON(filepath.Join(out, "sandbox.json"), &saved); err != nil {
		t.Fatal(err)
	}
	_, err := Remove(RemoveOptions{ID: "r2", OutDir: out})
	if code(err) != 2 || !strings.Contains(err.Error(), "hyp sandbox remove r2 --discard") {
		t.Fatalf("err = %v, want code 2 naming --discard", err)
	}
	other := filepath.Join(t.TempDir(), "other")
	create(t, "r3")
	if _, err := Remove(RemoveOptions{ID: "r3", OutDir: other}); err != nil {
		t.Fatal(err)
	}
	_, err = Remove(RemoveOptions{ID: "r2", OutDir: other})
	if code(err) != 2 || !strings.Contains(err.Error(), "exists; choose a new path") {
		t.Fatalf("err = %v, want the plain existing-path refusal", err)
	}
	if _, err := Remove(RemoveOptions{ID: "r2", Discard: true}); err != nil {
		t.Fatal(err)
	}
	if exists(box) {
		t.Error("discard left the sandbox")
	}
}
