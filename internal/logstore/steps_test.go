package logstore

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
)

// With HYP_STEP_LOG set, each event hyp writes appends one step line, step
// <Action> <json>, in order; a read appends none.
func TestStepLogGetsOneLinePerEvent(t *testing.T) {
	root := newStore(t)
	log := filepath.Join(t.TempDir(), "steps")
	t.Setenv(StepLogVar, log)
	first, err := Append(root, "a", reset(`{"role":"owner"}`))
	if err != nil {
		t.Fatal(err)
	}
	tr, err := TraceOf(root)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := tr.Append("a", events.Event{Type: events.TypeRead, Payload: json.RawMessage(`{}`)}); err != nil {
		t.Fatal(err)
	}
	second, err := tr.Append("a", events.Event{Type: events.TypeFindingAdd, Payload: json.RawMessage(`{}`)})
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile(log)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if len(lines) != 2 {
		t.Fatalf("step log holds %d lines, want 2:\n%s", len(lines), data)
	}
	for i, want := range []struct {
		action string
		ev     events.Event
	}{{"BudgetReset", first}, {"FindingAdd", second}} {
		head, body, ok := strings.Cut(strings.TrimPrefix(lines[i], "step "), " ")
		if !strings.HasPrefix(lines[i], "step ") || !ok || head != want.action {
			t.Fatalf("line %d = %q, want step %s <json>", i+1, lines[i], want.action)
		}
		var ev events.Event
		if err := json.Unmarshal([]byte(body), &ev); err != nil || ev.ID != want.ev.ID {
			t.Fatalf("line %d json = %q (%v), want the event %s", i+1, body, err, want.ev.ID)
		}
	}
}

// With HYP_STEP_LOG unset, hyp writes no step line anywhere.
func TestNoStepLogWhenUnset(t *testing.T) {
	root := newStore(t)
	t.Setenv(StepLogVar, "")
	if _, err := Append(root, "a", reset(`{"role":"owner"}`)); err != nil {
		t.Fatal(err)
	}
	for _, dir := range []string{root, os.Getenv("HOME")} {
		filepath.WalkDir(dir, func(p string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() {
				return nil
			}
			if b, _ := os.ReadFile(p); strings.HasPrefix(string(b), "step ") || strings.Contains(string(b), "\nstep ") {
				t.Errorf("%s holds a step line", p)
			}
			return nil
		})
	}
}
