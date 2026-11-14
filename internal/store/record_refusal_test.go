package store

import (
	"encoding/json"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// In a git repository, a write naming a file that neither the work tree
// nor the record holds is refused and writes nothing, since the record
// could not keep the bytes the event names.
func TestAppendRefusesANamedFileNoOneHolds(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	t.Setenv("GIT_CONFIG_GLOBAL", filepath.Join(t.TempDir(), "none"))
	t.Setenv("GIT_CONFIG_NOSYSTEM", "1")
	repo := t.TempDir()
	gitRun(t, repo, "init", "-q")
	root := filepath.Join(repo, "s")
	if err := Init(root); err != nil {
		t.Fatal(err)
	}
	g, err := Append(root, "a", nodeAdd(t, event.KindGoal, "", "goal", "why"))
	if err != nil {
		t.Fatal(err)
	}
	q, err := Append(root, "a", nodeAdd(t, event.KindDesignQuestion, g.ID, "q", "why"))
	if err != nil {
		t.Fatal(err)
	}
	h, err := Append(root, "a", nodeAdd(t, event.KindHypothesis, q.ID, "h", "why"))
	if err != nil {
		t.Fatal(err)
	}
	raw, err := json.Marshal(event.NodeAdd{Kind: event.KindExperiment, Parent: h.ID, Title: "e", Recipe: event.RecipeNone, Rubric: event.RubricReading,
		Records: []event.SubjectFile{{Path: ".tree/evidence/E-AAAAAA/gone.txt", SHA256: strings.Repeat("0", 64)}}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Append(root, "a", event.Event{Type: event.TypeNodeAdd, Payload: raw, Why: "why"})
	if err == nil || !strings.Contains(err.Error(), "in neither the work tree nor the record") {
		t.Fatalf("Append of an Experiment pinning a file no one holds: %v, want a refusal", err)
	}
	evs, _, err := Events(mustRoot(t, root))
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 3 {
		t.Fatalf("the store holds %d events after the refusal, want 3", len(evs))
	}
}

func mustRoot(t *testing.T, dir string) string {
	t.Helper()
	root, err := FindRoot(dir)
	if err != nil {
		t.Fatal(err)
	}
	return root
}
