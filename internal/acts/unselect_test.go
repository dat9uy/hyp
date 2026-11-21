package acts

import (
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/tables"
)

func (e *runEnv) unselectOpts(actor, role string) UnselectOpts {
	return UnselectOpts{Dir: e.root, Actor: actor, Role: role, Goal: e.goal}
}

func (e *runEnv) count(t *testing.T, typ string) int {
	t.Helper()
	evs, _ := e.log(t)
	n := 0
	for _, ev := range evs {
		if ev.Type == typ {
			n++
		}
	}
	return n
}

// hasAct reports whether the allowed acts of the Goal hold act for role.
func hasAct(rs []Ready, role, act string) bool {
	for _, r := range rs {
		if r.Role == role && r.Act == act {
			return true
		}
	}
	return false
}

func TestUnselectOpensTheGoal(t *testing.T) {
	e := newRunEnv(t)
	commit, tree := e.selected(t)
	sel := e.selects(t)
	rep, err := Unselect(e.unselectOpts("owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Goal != e.goal || rep.Select.ID != sel[0].ID || rep.Select.Commit != commit {
		t.Fatalf("report = %+v, want the select %s undone", rep, sel[0].ID)
	}
	g := e.goalRow(t)
	if g.State != tables.StateOpen || g.Select != nil || tables.RunsUsed(g) != 0 {
		t.Fatalf("state = %s, select %+v, runs %d", g.State, g.Select, tables.RunsUsed(g))
	}
	got := AllowedFor(g, e.root)
	if !hasAct(got, "owner", ActSelect) || hasAct(got, "runner", ActSelect) || !hasAct(got, "grill", ActPlace) {
		t.Fatalf("allowed = %+v, want the owner's select and no runner's select", got)
	}
	// the runner's select needs a new cell; a rerun of the old row is none
	_, err = Select(e.selectOpts(tree, "runner-1", "runner"))
	refusedBy(t, err, "New cell")
	if _, err := e.run(commit); err != nil {
		t.Fatal(err)
	}
	_, err = Select(e.selectOpts(tree, "runner-1", "runner"))
	refusedBy(t, err, "New cell")
	// a run on a new row is a new cell
	c2 := e.child(t, commit, map[string]string{"more": "1\n"})
	if _, err := e.run(c2); err != nil {
		t.Fatal(err)
	}
	t2, err := gitread.RowTree(e.root, c2)
	if err != nil {
		t.Fatal(err)
	}
	if !hasAct(AllowedFor(e.goalRow(t), e.root), "runner", ActSelect) {
		t.Fatal("no runner's select after a run on a new row")
	}
	if _, err := Select(e.selectOpts(t2, "runner-1", "runner")); err != nil {
		t.Fatal(err)
	}
	if g := e.goalRow(t); g.State != tables.StateSelected || g.Select.RowTree != t2 {
		t.Fatalf("state = %s, select %+v", g.State, g.Select)
	}
}

func TestUnselectRefusals(t *testing.T) {
	e := newRunEnv(t)
	// an open Goal has no select
	_, err := Unselect(e.unselectOpts("owner-1", "owner"))
	refusedBy(t, err, "Selected")
	e.selected(t)
	for _, role := range []string{"runner", "grill", "probe-writer"} {
		_, err = Unselect(e.unselectOpts("owner-1", role))
		refusedBy(t, err, "Owner unselects")
	}
	_, err = Unselect(e.unselectOpts("owner-2", "owner"))
	refusedBy(t, err, "Goal's actor")
	if e.count(t, events.TypeUnselect) != 0 {
		t.Fatal("a refused unselect wrote an event")
	}
	if _, err := Unselect(e.unselectOpts("owner-1", "owner")); err != nil {
		t.Fatal(err)
	}
	_, err = Unselect(e.unselectOpts("owner-1", "owner"))
	refusedBy(t, err, "Selected")
	// a landed Goal takes no unselect
	e2 := newRunEnv(t)
	e2.selected(t)
	if _, err := Land(e2.landOpts("owner-1", "owner")); err != nil {
		t.Fatal(err)
	}
	_, err = Unselect(e2.unselectOpts("owner-1", "owner"))
	refusedBy(t, err, "Not landed")
	if e.count(t, events.TypeUnselect) != 1 || e2.count(t, events.TypeUnselect) != 0 {
		t.Fatal("a refused unselect wrote an event")
	}
}
