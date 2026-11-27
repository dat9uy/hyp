package acts

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
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

// selected runs a met leaf with a spec, which this clone then shows
// selected; it returns the leaf's commit and row tree.
func (e *runEnv) selected(t *testing.T) (commit, tree string) {
	t.Helper()
	commit, tree = e.metLeaf(t, "ok", true)
	if g := e.shown(t); g.State != tables.StateSelected {
		t.Fatalf("test setup: state %s, want selected", g.State)
	}
	return commit, tree
}

// A land on the selected commit carries its evidence: the commit, its tree,
// its row tree and each example's verdict, so a clone that replays git alone
// ends the Goal selected on that commit.
func TestLandOnTheSelectedCommitCarriesItsEvidence(t *testing.T) {
	e := newRunEnv(t)
	commit, tree := e.selected(t)
	shown := e.shown(t)
	rep, err := Land(e.landOpts("owner-1", "owner"))
	if err != nil {
		t.Fatal(err)
	}
	if rep.Stop != StopSelected || rep.Cause.Act != ActRun || rep.Cause.ID != shown.Select.ID || rep.Commit != commit || len(rep.Ran) != 0 {
		t.Fatalf("report = %+v, want the stop selected on %s caused by %s", rep, commit, shown.Select.ID)
	}
	ls := e.lands(t)
	var p events.Land
	if len(ls) != 1 || json.Unmarshal(ls[0].Payload, &p) != nil || p.Stop != StopSelected || p.Role != "owner" {
		t.Fatalf("lands = %d, payload %+v", len(ls), p)
	}
	cur, _ := shown.Current("line-1/a")
	wantTree := testutil.Git(t, e.root, "rev-parse", commit+"^{tree}")
	if p.Commit != commit || p.Tree != wantTree || p.RowTree != tree || len(p.Verdicts) != 1 ||
		p.Verdicts[0] != (events.LandVerdict{Line: "line-1", Example: "a", Hash: cur.Hash, Verdict: "met"}) {
		t.Fatalf("payload = %+v, want the evidence of %s", p, commit)
	}
	g := e.goalRow(t) // replay of git alone
	if g.State != tables.StateEnded || g.Land == nil || g.Select == nil || g.Select.Commit != commit {
		t.Fatalf("state = %s, land %+v, select %+v", g.State, g.Land, g.Select)
	}
	if st := Stand(g); st.Stop != StopSelected || st.Waits != "none" {
		t.Fatalf("standing = %+v", st)
	}
	if got := AllowedFor(g); len(got) != 0 {
		t.Fatalf("a landed Goal allows %+v", got)
	}
	// a landed Goal takes no write: Not ended refuses each
	_, err = Land(e.landOpts("owner-1", "owner"))
	refusedBy(t, err, "Not ended")
	_, err = Select(e.selectOpts(tree, "owner-1", "owner"))
	refusedBy(t, err, "Not ended")
	_, err = e.run(commit)
	refusedBy(t, err, "Not ended")
	if len(e.lands(t)) != 1 {
		t.Fatal("a refused land wrote an event")
	}
}

// A land on a commit this clone holds no Result of runs each example there
// first, and lands only when every one is met.
func TestLandRunsTheExamplesThisCloneLacks(t *testing.T) {
	e := newRunEnv(t)
	bad := e.child(t, e.base, map[string]string{"counter": "nope\n", "goals/" + e.goal + "/specs/spec.md": "s\n"})
	o := e.landOpts("owner-1", "owner")
	o.Commit = bad
	_, err := Land(o)
	refusedBy(t, err, "Evidence")
	if fault.CodeOf(err) != fault.NotMetOnCommit || len(e.lands(t)) != 0 {
		t.Fatalf("err = %v, lands = %d", err, len(e.lands(t)))
	}
	good := e.child(t, e.base, map[string]string{"counter": "ok\n", "goals/" + e.goal + "/specs/spec.md": "s\n"})
	o.Commit = good
	rep, err := Land(o)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Commit != good || len(rep.Ran) != 1 || rep.Ran[0] != "line-1/a" || len(rep.Verdicts) != 1 || rep.Verdicts[0].Verdict != "met" {
		t.Fatalf("report = %+v, want line-1/a run and met on %s", rep, good)
	}
	if len(e.results(t)) != 0 {
		t.Fatal("the land recorded a Result")
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
	if g := e.goalRow(t); g.State != tables.StateEnded || g.Select != nil {
		t.Fatalf("state = %s", g.State)
	}
}

func TestLandRefusals(t *testing.T) {
	e := newRunEnv(t)
	// an open Goal with no select and no commit named has no stop
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
	commit, _ := e.selected(t)
	g := e.shown(t)
	st := Stand(g)
	if st.Stop != StopSelected || st.Waits != "owner" {
		t.Fatalf("standing = %+v", st)
	}
	got := AllowedFor(g)
	if len(got) != 3 || got[0].Role != "owner" || got[0].Act != ActLand || got[0].Goal != e.goal || got[0].Commit != commit ||
		got[1].Role != "owner" || got[1].Act != ActUnselect || got[1].Goal != e.goal || got[1].Commit != commit ||
		got[2].Role != "owner" || got[2].Act != ActClose || got[2].Goal != e.goal {
		t.Fatalf("allowed = %+v, want only the owner's land, unselect and close", got)
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
	refusedBy(t, from(), "From ended")
	e.selected(t)
	refusedBy(t, from(), "From ended")
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
