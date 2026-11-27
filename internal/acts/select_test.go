package acts

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
)

func (e *runEnv) selectOpts(tree, actor, role string) SelectOpts {
	return SelectOpts{Dir: e.root, Actor: actor, Role: role, Goal: e.goal, Tree: tree}
}

// metLeaf runs line-1 on a leaf from H0 that prints value, and returns its
// commit and row tree. The leaf also holds a spec.
func (e *runEnv) metLeaf(t *testing.T, value string, withSpec bool) (commit, tree string) {
	t.Helper()
	files := map[string]string{"counter": value + "\n"}
	if withSpec {
		files["goals/"+e.goal+"/specs/spec.md"] = "spec\n"
	}
	commit = e.child(t, e.base, files)
	if _, err := e.run(commit); err != nil {
		t.Fatal(err)
	}
	tree, err := gitread.RowTree(e.root, commit)
	if err != nil {
		t.Fatal(err)
	}
	return commit, tree
}

func (e *runEnv) selects(t *testing.T) []events.Event {
	t.Helper()
	evs, _ := e.log(t)
	var out []events.Event
	for _, ev := range evs {
		if ev.Type == events.TypeSelect {
			out = append(out, ev)
		}
	}
	return out
}

// shown is the Goal as this clone shows it: replayed, with the select its
// Results show calculated.
func (e *runEnv) shown(t *testing.T) *tables.Goal {
	t.Helper()
	tb, _, err := ReadTables(e.root)
	if err != nil {
		t.Fatal(err)
	}
	g := tb.ByID(e.goal)
	if g == nil {
		t.Fatal("no Goal")
	}
	return g
}

// A lone candidate is selected by the Results that show it: no event holds
// the select, and no act of the runner writes one.
func TestOneCandidateIsSelectedWithNoEvent(t *testing.T) {
	e := newRunEnv(t)
	c, tree := e.metLeaf(t, "ok", true)
	g := e.shown(t)
	if g.State != tables.StateSelected || g.Select == nil || !g.Select.Calculated || g.Select.Commit != c || g.Select.RowTree != tree {
		t.Fatalf("state = %s, select = %+v; want selected on %s", g.State, g.Select, c)
	}
	if n := len(e.selects(t)); n != 0 {
		t.Fatalf("%d select events, want none", n)
	}
	// replay alone, as a clone that lacks the Results reads git, does not
	// show the select
	if r := e.goalRow(t); r.Select != nil {
		t.Fatalf("replay holds a select: %+v", r.Select)
	}
	// the runner selects nothing, and the owner's select is refused by
	// Selected
	_, err := Select(e.selectOpts(tree, "runner-1", "runner"))
	refusedBy(t, err, "Candidate")
	_, err = Select(e.selectOpts(tree, "owner-1", "owner"))
	refusedBy(t, err, "Selected")
	// a second candidate is calculated too: the Goal shows two candidates,
	// and no select
	e.metLeaf(t, "ok2", true)
	if g := e.shown(t); g.State != tables.StateOpen || g.Select != nil || Stand(g, e.root).Stop != StopTwoCandidates {
		t.Fatalf("state = %s, select = %+v, want open at two candidates", g.State, g.Select)
	}
}

func TestSelectRefusesNonCandidate(t *testing.T) {
	e := newRunEnv(t)
	_, okTree := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	// a failed leaf is a row of the Goal but not a candidate
	bad := e.child(t, e.base, map[string]string{"counter": "nope\n", "goals/" + e.goal + "/specs/spec.md": "s\n"})
	if _, err := e.run(bad); err != nil {
		t.Fatal(err)
	}
	badTree, _ := gitread.RowTree(e.root, bad)
	_, err := Select(e.selectOpts(badTree, "owner-1", "owner"))
	refusedBy(t, err, "Candidate")
	_, err = Select(e.selectOpts(strings.Repeat("0", 40), "owner-1", "owner"))
	refusedBy(t, err, "Candidate")
	if _, err := Select(e.selectOpts(okTree, "owner-1", "owner")); err != nil {
		t.Fatalf("the candidate itself was refused: %v", err)
	}
}

func TestRunnerCannotSelectAmongTwo(t *testing.T) {
	e := newRunEnv(t)
	_, t1 := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	if g := e.shown(t); len(tables.Candidates(g, gitread.Ancestry{Dir: e.root})) != 2 || g.State != tables.StateOpen {
		t.Fatalf("test setup: want two candidates and an open Goal, state %s", g.State)
	}
	_, err := Select(e.selectOpts(t1, "runner-1", "runner"))
	refusedBy(t, err, "Candidate")
	if n := len(e.selects(t)); n != 0 {
		t.Fatalf("%d selects recorded", n)
	}
}

func TestOwnerSelectsAmongTwo(t *testing.T) {
	e := newRunEnv(t)
	_, t1 := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	rep, err := Select(e.selectOpts(t1, "owner-1", "owner"))
	if err != nil || rep.RowTree != t1 {
		t.Fatalf("report = %+v, %v", rep, err)
	}
	var p events.Select
	_ = json.Unmarshal(e.selects(t)[0].Payload, &p)
	if p.Role != "owner" {
		t.Fatalf("role = %s", p.Role)
	}
}

func TestOwnerOfOtherGoalRefused(t *testing.T) {
	e := newRunEnv(t)
	_, tree := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	_, err := Select(e.selectOpts(tree, "owner-2", "owner"))
	refusedBy(t, err, "Goal's actor")
	if _, err := Select(e.selectOpts(tree, "owner-1", "owner")); err != nil {
		t.Fatalf("the Goal's own owner was refused: %v", err)
	}
}

// A lone candidate with no spec is not selected: hyp status holds it back
// with the Spec refusal.
func TestNoSpecHoldsTheSelectBack(t *testing.T) {
	e := newRunEnv(t)
	e.metLeaf(t, "ok", false)
	g := e.shown(t)
	if g.State != tables.StateOpen {
		t.Fatalf("state = %s, want open", g.State)
	}
	held := HeldBack(g, e.root)
	if len(held) != 1 {
		t.Fatalf("held back = %+v, want one", held)
	}
	refusedBy(t, held[0].Err, "Spec")
}

func TestOtherGoalPathHoldsTheSelectBack(t *testing.T) {
	e := newRunEnv(t)
	// A Result recorded on a commit whose own change writes another Goal's
	// path: run would refuse it, so the Result is appended by hand.
	c := e.child(t, e.base, map[string]string{
		"counter": "ok\n", "goals/" + e.goal + "/specs/spec.md": "s\n", "goals/G-OTHER1/note": "n\n",
	})
	tree, _ := gitread.RowTree(e.root, c)
	cur, _ := e.goalRow(t).Current("line-1/a")
	payload, _ := json.Marshal(events.ResultRecord{Role: "runner", Line: "line-1", Commit: c, RowTree: tree,
		Cells: []events.ExampleCell{{Example: "a", Hash: cur.Hash, Probe: "p", Output: "o", Stderr: "e", Verdict: "met"}}})
	if _, err := logstore.Append(e.root, "runner-1", func([]events.Event) (events.Event, error) {
		return events.Event{Type: events.TypeResultRecord, Node: e.goalRow(t).EventID, Payload: payload}, nil
	}); err != nil {
		t.Fatal(err)
	}
	g := e.shown(t)
	held := HeldBack(g, e.root)
	if g.State != tables.StateOpen || len(held) != 1 {
		t.Fatalf("state = %s, held back = %+v", g.State, held)
	}
	refusedBy(t, held[0].Err, "Own path")
}

func TestSelectReturnsTheSelectedGoal(t *testing.T) {
	e := newRunEnv(t)
	_, tree := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	rep, err := Select(e.selectOpts(tree, "owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	// the channel prints the PR body from the Goal as the log holds it now
	g := rep.Selected
	if g == nil || g.Select == nil || g.ID != e.goal || g.Select.Commit != rep.Commit || g.Select.RowTree != tree || rep.RowTree != tree {
		t.Fatalf("report = %+v", rep)
	}
}

func TestSelectWritesNoGitState(t *testing.T) {
	e := newRunEnv(t)
	_, tree := e.metLeaf(t, "ok", true)
	e.metLeaf(t, "ok2", true)
	before := testutil.Snapshot(t, e.root)
	rep, err := Select(e.selectOpts(tree, "owner-1", "owner"))
	if err != nil || len(e.selects(t)) != 1 || rep.Selected == nil {
		t.Fatalf("select did not record: %v", err)
	}
	if testutil.Snapshot(t, e.root) != before {
		t.Fatal("select changed the repository's git state")
	}
}
