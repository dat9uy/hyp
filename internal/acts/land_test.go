package acts

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/tables"
)

func (e *runEnv) landOpts(actor, role string) LandOpts {
	return LandOpts{Dir: e.root, Actor: actor, Role: role, Goal: e.goal}
}

func (e *runEnv) lands(t *testing.T) []events.Event {
	t.Helper()
	evs, _ := e.log(t)
	var out []events.Event
	for _, ev := range evs {
		if ev.Type == events.TypeLand {
			out = append(out, ev)
		}
	}
	return out
}

// selected runs a met leaf with a spec and selects it; it returns the leaf's
// commit and row tree.
func (e *runEnv) selected(t *testing.T) (commit, tree string) {
	t.Helper()
	commit, tree = e.metLeaf(t, "ok", true)
	if _, err := Select(e.selectOpts(tree, "runner-1", "runner")); err != nil {
		t.Fatal(err)
	}
	return commit, tree
}

func TestLandAfterSelectNamesTheSelect(t *testing.T) {
	e := newRunEnv(t)
	commit, tree := e.selected(t)
	sel := e.selects(t)
	rep, err := Land(e.landOpts("owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Stop != StopSelected || rep.Cause.Act != ActSelect || rep.Cause.ID != sel[0].ID {
		t.Fatalf("report = %+v, want the stop selected caused by %s", rep, sel[0].ID)
	}
	ls := e.lands(t)
	var p events.Land
	if len(ls) != 1 || json.Unmarshal(ls[0].Payload, &p) != nil || p.Stop != StopSelected || p.Cause != sel[0].ID || p.Role != "owner" {
		t.Fatalf("lands = %d, payload %+v", len(ls), p)
	}
	g := e.goalRow(t)
	if g.State != tables.StateLanded || g.Land == nil || g.Select == nil {
		t.Fatalf("state = %s, land %+v", g.State, g.Land)
	}
	if st := Stand(g, e.root); st.Stop != StopSelected || st.Waits != "none" || st.Cause.ID != sel[0].ID {
		t.Fatalf("standing = %+v", st)
	}
	if got := AllowedFor(g, e.root); len(got) != 0 {
		t.Fatalf("a landed Goal allows %+v", got)
	}
	// a landed Goal takes no write: Not landed refuses each
	_, err = Land(e.landOpts("owner-1", "owner"))
	refusedBy(t, err, "Not landed")
	_, err = Select(e.selectOpts(tree, "runner-1", "runner"))
	refusedBy(t, err, "Not landed")
	_, err = e.run(commit)
	refusedBy(t, err, "Not landed")
	if len(e.lands(t)) != 1 {
		t.Fatal("a refused land wrote an event")
	}
}

func TestLandAfterBudgetUsedNamesTheLastRun(t *testing.T) {
	e := newRunEnv(t)
	for _, v := range []string{"x1", "x2", "x3"} {
		if _, err := e.run(e.leaf(t, e.base, v)); err != nil {
			t.Fatal(err)
		}
	}
	results, _ := e.log(t)
	last := ""
	for _, ev := range results {
		if ev.Type == events.TypeResultRecord {
			last = ev.ID
		}
	}
	rep, err := Land(e.landOpts("owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Stop != StopBudgetUsed || rep.Cause.Act != ActRun || rep.Cause.ID != last {
		t.Fatalf("report = %+v, want budget used caused by %s", rep, last)
	}
	if g := e.goalRow(t); g.State != tables.StateLanded || g.Select != nil {
		t.Fatalf("state = %s", g.State)
	}
}

func TestLandRefusals(t *testing.T) {
	e := newRunEnv(t)
	// an open Goal has no stop
	_, err := Land(e.landOpts("owner-1", "owner"))
	refusedBy(t, err, "Stop")
	e.selected(t)
	for _, role := range []string{"runner", "grill", "probe-writer"} {
		_, err := Land(e.landOpts(role+"-1", role))
		refusedBy(t, err, "Owner lands")
	}
	_, err = Land(e.landOpts("owner-2", "owner"))
	refusedBy(t, err, "Goal's actor")
	if len(e.lands(t)) != 0 {
		t.Fatal("a refused land wrote an event")
	}
}

func TestStopWaitsForTheOwnerAfterSelect(t *testing.T) {
	e := newRunEnv(t)
	e.selected(t)
	g := e.goalRow(t)
	st := Stand(g, e.root)
	if st.Stop != StopSelected || st.Waits != "owner" {
		t.Fatalf("standing = %+v", st)
	}
	got := AllowedFor(g, e.root)
	if len(got) != 2 || got[0].Role != "owner" || got[0].Act != ActLand || got[0].Goal != e.goal ||
		got[1].Role != "owner" || got[1].Act != ActUnselect || got[1].Goal != e.goal {
		t.Fatalf("allowed = %+v, want only the owner's land and unselect", got)
	}
}

func TestAddGoalFromNeedsALandedGoal(t *testing.T) {
	e := newRunEnv(t)
	from := func() error {
		o := opts(e.root, "HEAD", gapLine)
		o.From = e.goal
		_, err := AddGoal(o)
		return err
	}
	refusedBy(t, from(), "From landed")
	e.selected(t)
	refusedBy(t, from(), "From landed")
	if _, err := Land(e.landOpts("owner-1", "owner")); err != nil {
		t.Fatal(err)
	}
	if err := from(); err != nil {
		t.Fatal(err)
	}
	evs, _ := e.log(t)
	var p events.GoalAdd
	// from holds the cited Goal's full goal.add id, which no other Goal shares.
	if err := events.Decode(evs[len(evs)-1], &p); err != nil || p.From != e.goalRow(t).EventID {
		t.Fatalf("from = %q, %v", p.From, err)
	}
}
