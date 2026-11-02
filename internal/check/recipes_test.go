package check

import (
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
)

// A store under a bundle's known-answers/ is a case input, not a tree; the
// same directory name outside a bundle is searched as usual.
func TestFindStoresSkipsKnownAnswerInputs(t *testing.T) {
	top := t.TempDir()
	for _, d := range []string{
		".tree",
		"recipes/x@1/known-answers/c/input/.tree",
		"other/known-answers/.tree",
	} {
		if err := os.MkdirAll(filepath.Join(top, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(top, "recipes", "x@1", "recipe.md"), []byte("---\nname: x\nversion: 1\nformat: 1\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := findStores(top)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{top, filepath.Join(top, "other", "known-answers")}; !slices.Equal(got, want) {
		t.Fatalf("stores = %q, want %q", got, want)
	}
}

// leftover finds recipes/x/ for a pin to x@1 only when it hashes as the pin.
func TestLeftoverHashesAsThePin(t *testing.T) {
	top := t.TempDir()
	if err := os.MkdirAll(filepath.Join(top, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	dir := filepath.Join(top, "recipes", "x")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "recipe.md"), []byte("---\nname: x\nversion: 1\nformat: 1\n---\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, _, hash, err := recipe.HashOnly(dir)
	if err != nil {
		t.Fatal(err)
	}
	if got, version := leftover(top, replay.Pin{Ref: event.Ref{Name: "x", Version: "1"}, Hash: hash}); got != dir || version != "1" {
		t.Fatalf("leftover = %q %q, want %q 1", got, version, dir)
	}
	if got, _ := leftover(top, replay.Pin{Ref: event.Ref{Name: "x", Version: "1"}, Hash: "other"}); got != "" {
		t.Fatalf("leftover with another hash = %q, want none", got)
	}
	if got, _ := leftover(top, replay.Pin{Ref: event.Ref{Name: "y", Version: "1"}, Hash: hash}); got != "" {
		t.Fatalf("leftover for y@1 = %q, want none", got)
	}
}
