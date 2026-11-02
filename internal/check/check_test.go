package check

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/store"
)

func appendTo(t *testing.T, root, actor string, ev event.Event) event.Event {
	t.Helper()
	out, err := store.Append(root, actor, ev)
	if err != nil {
		t.Fatal(err)
	}
	return out
}

func node(kind, parent string) event.Event {
	raw, _ := json.Marshal(event.NodeAdd{Kind: kind, Parent: parent, Title: kind})
	return event.Event{Type: event.TypeNodeAdd, Payload: raw, Why: "why"}
}

func edge(from, to string) event.Event {
	raw, _ := json.Marshal(event.Edge{Type: event.EdgeDependsOn, To: to})
	return event.Event{Type: event.TypeEdgeAdd, Node: from, Payload: raw, Why: "why"}
}

func TestCheckCountsBadLineAndCrossActorCycle(t *testing.T) {
	root := t.TempDir()
	if err := store.Init(root); err != nil {
		t.Fatal(err)
	}
	g := appendTo(t, root, "a", node(event.KindGoal, ""))
	q1 := appendTo(t, root, "a", node(event.KindDesignQuestion, g.ID))
	q2 := appendTo(t, root, "a", node(event.KindDesignQuestion, g.ID))
	h1 := appendTo(t, root, "a", node(event.KindHypothesis, q1.ID))
	h2 := appendTo(t, root, "a", node(event.KindHypothesis, q2.ID))
	appendTo(t, root, "a", edge(h1.ID, h2.ID))

	// Actor b's worktree did not yet have a's edge, so its reverse edge was
	// valid there; after the merge it closes a cycle.
	raw, _ := json.Marshal(event.Edge{Type: event.EdgeDependsOn, To: h1.ID})
	reverse, _ := json.Marshal(event.Event{V: 1, ID: "7ZZZZZZZZZZZZZZZZZZZZZZZZZ", TS: "2026-10-24T00:00:00Z",
		Actor: "b", Type: event.TypeEdgeAdd, Node: h2.ID, Payload: raw, Why: "from b"})
	data := append(reverse, '\n')
	data = append(data, []byte("{not json\n")...)
	if err := os.WriteFile(filepath.Join(root, ".tree", "events", "b.jsonl"), data, 0o644); err != nil {
		t.Fatal(err)
	}

	report, errs := Run(root)
	last := report[len(report)-1]
	if errs != 2 || !strings.HasPrefix(last, "check: 2 errors") {
		t.Fatalf("errors = %d, report:\n%s", errs, strings.Join(report, "\n"))
	}
	joined := strings.Join(report, "\n")
	for _, want := range []string{"error: b.jsonl:2 ", "error: rejected 7ZZZZZZZZZZZZZZZZZZZZZZZZZ: depends_on cycle"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("report lacks %q:\n%s", want, joined)
		}
	}
}

func TestCheckCleanStore(t *testing.T) {
	root := t.TempDir()
	store.Init(root)
	appendTo(t, root, "a", node(event.KindGoal, ""))
	report, errs := Run(root)
	if errs != 0 || report[len(report)-1] != "check: 0 errors, 0 warnings" {
		t.Fatalf("report: %v", report)
	}
}

// Every synthetic tree under internal/testutil/testdata, written by the
// current hyp, replays with no rejected event.
func TestRunTreesReplayWithoutRejections(t *testing.T) {
	trees, _ := filepath.Glob("../testutil/testdata/*/.tree")
	if len(trees) == 0 {
		t.Fatal("no run trees found")
	}
	for _, tree := range trees {
		if report, errs := Run(filepath.Dir(tree)); errs != 0 {
			t.Errorf("%s: %v", tree, report)
		}
	}
}
