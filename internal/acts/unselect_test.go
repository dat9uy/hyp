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
	rep, err := Unselect(e.unselectOpts("owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Goal != e.goal || rep.Select.Commit != commit || rep.Select.RowTree != tree {
		t.Fatalf("report = %+v, want the select of %s undone", rep, commit)
	}
	evs, _ := e.log(t)
	var p events.Unselect
	if err := events.Decode(evs[len(evs)-1], &p); err != nil || p.Commit != commit || p.RowTree != tree {
		t.Fatalf("payload = %+v, %v, want it to name %s", p, err, commit)
	}
	g := e.shown(t)
	if g.State != tables.StateOpen || g.Select != nil || tables.RunsUsed(g) != 0 || !g.Rejected[tree] {
		t.Fatalf("state = %s, select %+v, runs %d", g.State, g.Select, tables.RunsUsed(g))
	}
	got := AllowedFor(g)
	if hasAct(got, "owner", ActSelect) || hasAct(got, "runner", ActSelect) || !hasAct(got, "owner", ActPlace) || !hasAct(got, "runner", ActRun) {
		t.Fatalf("allowed = %+v, want the owner's place, the runner's run and no select", got)
	}
	// the rejected row is never selected again, and a land on it is refused
	if _, err := e.run(commit); err != nil {
		t.Fatal(err)
	}
	if g := e.shown(t); g.State != tables.StateOpen {
		t.Fatalf("after a rerun of the rejected row state = %s", g.State)
	}
	o := e.landOpts("owner-1", "owner")
	o.Commit = commit
	_, err = Land(o)
	refusedBy(t, err, "Evidence")
	// a run on a new row is a new cell, and its row is selected
	c2 := e.child(t, commit, map[string]string{"more": "1\n"})
	if _, err := e.run(c2); err != nil {
		t.Fatal(err)
	}
	t2, err := gitread.RowTree(e.root, c2)
	if err != nil {
		t.Fatal(err)
	}
	if g := e.shown(t); g.State != tables.StateSelected || g.Select.RowTree != t2 {
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
	// a commit the Goal is not selected on
	other := e.unselectOpts("owner-1", "owner")
	other.Commit = e.base
	_, err = Unselect(other)
	refusedBy(t, err, "Selected")
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
	refusedBy(t, err, "Not ended")
	if e.count(t, events.TypeUnselect) != 1 || e2.count(t, events.TypeUnselect) != 0 {
		t.Fatal("a refused unselect wrote an event")
	}
}
