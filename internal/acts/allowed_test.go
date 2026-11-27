package acts

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/tables"
)

// lines are the allowed acts of role, each as its act and keys: the channel
// words them as command lines.
func lines(rl []Ready, role string) []Ready {
	var out []Ready
	for _, l := range rl {
		if l.Role == role {
			out = append(out, l)
		}
	}
	return out
}

func acted(rl []Ready, act string) bool {
	return slices.ContainsFunc(rl, func(r Ready) bool { return r.Act == act })
}

// allowed lists the Goal's allowed acts as this clone shows them, with the
// select its Results show.
func (e *runEnv) allowed(t *testing.T) []Ready {
	t.Helper()
	return AllowedFor(e.shown(t))
}

// owners lists only the acts of the owner that are not the place and the
// close, which an open Goal always allows.
func owners(rl []Ready) []Ready {
	var out []Ready
	for _, l := range lines(rl, "owner") {
		if l.Act != ActPlace && l.Act != ActClose {
			out = append(out, l)
		}
	}
	return out
}

func TestAllowedWrittenGoalOwnerMayPlace(t *testing.T) {
	e := newRunEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	if g := e.goalRow(t); g.State != tables.StateWritten {
		t.Fatalf("test setup: %s", g.State)
	}
	rl := e.allowed(t)
	if got := lines(rl, "grill"); len(got) != 0 {
		t.Fatalf("grill may %+v, want nothing: the grill writes no event", got)
	}
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %+v, want nothing on a Goal that is not open", got)
	}
	want := []Ready{{Goal: e.goal, Role: "owner", Act: ActAddGoal}, {Goal: e.goal, Role: "owner", Act: ActPlace}, {Goal: e.goal, Role: "owner", Act: ActClose}}
	if got := lines(rl, "owner"); !slices.Equal(got, want) {
		t.Fatalf("owner may %+v, want the Draft replace, the place and the close", got)
	}
}

func TestAllowedOpenGoalRunnerMayRun(t *testing.T) {
	e := newRunEnv(t)
	rl := e.allowed(t)
	want := Ready{Goal: e.goal, Role: "runner", Act: ActRun, Line: "line-1"}
	if got := lines(rl, "runner"); !slices.Equal(got, []Ready{want}) {
		t.Fatalf("runner may %+v, want [%+v]", got, want)
	}
	if got := lines(rl, "owner"); !slices.Equal(got, []Ready{{Goal: e.goal, Role: "owner", Act: ActPlace}, {Goal: e.goal, Role: "owner", Act: ActClose}}) {
		t.Fatalf("owner may %+v, want only the place and the close on an open Goal with no stop", got)
	}
	for _, l := range rl {
		if !slices.Contains([]string{ActAddGoal, ActPlace, ActRun, ActSelect, ActClose}, l.Act) {
			t.Fatalf("allowed %+v, an act the binary lacks", l)
		}
	}
}

// A lone candidate shows the Goal selected: the owner may land or unselect
// its commit, and the runner has no act.
func TestAllowedOneCandidateShowsSelected(t *testing.T) {
	e := newRunEnv(t)
	c, _ := e.metLeaf(t, "ok", true)
	rl := e.allowed(t)
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %+v, want nothing on a selected Goal", got)
	}
	want := []Ready{{Goal: e.goal, Role: "owner", Act: ActLand, Commit: c}, {Goal: e.goal, Role: "owner", Act: ActUnselect, Commit: c}}
	if got := owners(rl); !slices.Equal(got, want) {
		t.Fatalf("owner may %+v, want %+v", got, want)
	}
	// a candidate whose commit holds no spec is not selected: run on the spec commit first
	e2 := newRunEnv(t)
	e2.metLeaf(t, "ok", false)
	rl = e2.allowed(t)
	if acted(rl, ActSelect) || acted(rl, ActLand) || !acted(lines(rl, "runner"), ActRun) {
		t.Fatalf("allowed %+v without a spec, want only runs", rl)
	}
}

func TestAllowedTwoCandidatesOnlyOwnerMaySelect(t *testing.T) {
	e := newRunEnv(t)
	_, t1 := e.metLeaf(t, "ok", true)
	_, t2 := e.metLeaf(t, "ok2", true)
	rl := e.allowed(t)
	if got := lines(rl, "runner"); acted(got, ActSelect) || acted(got, ActRun) {
		t.Fatalf("runner may %+v at a stop", got)
	}
	owner := lines(rl, "owner")
	for _, tree := range []string{t1, t2} {
		if !slices.Contains(owner, Ready{Goal: e.goal, Role: "owner", Act: ActSelect, Tree: tree}) {
			t.Fatalf("owner may %+v, want a select of %s", owner, tree)
		}
	}
	if st := Stand(e.shown(t)); st.Stop != StopTwoCandidates || len(st.Candidates) != 2 {
		t.Fatalf("standing = %+v", st)
	}
}

func TestAllowedBudgetUsedRunnerMayNotRun(t *testing.T) {
	e := newRunEnv(t)
	// until the third tree has a cell for every example, runs stay allowed
	for i, v := range []string{"x1", "x2", "x3"} {
		c := e.leaf(t, e.base, v)
		if _, err := e.run(c); err != nil {
			t.Fatal(err)
		}
		st := Stand(e.goalRow(t))
		if i < 2 && st.Stop != StopNone {
			t.Fatalf("after %d trees stop = %s", i+1, st.Stop)
		}
	}
	rl := e.allowed(t)
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %+v after the budget is used", got)
	}
	if st := Stand(e.goalRow(t)); st.Stop != StopBudgetUsed || st.Used != 3 || st.Waits != "owner" {
		t.Fatalf("standing = %+v", st)
	}
}

// A candidate with no spec at three Hypotheses still has an act: the run on
// a spec commit of the same Hypothesis, which costs no budget. With it the
// runner may select.
func TestAllowedCandidateWithoutSpecAtBudgetMayRunSpecCommit(t *testing.T) {
	e := newRunEnv(t)
	for _, v := range []string{"x1", "x2"} {
		if _, err := e.run(e.leaf(t, e.base, v)); err != nil {
			t.Fatal(err)
		}
	}
	commit, tree := e.metLeaf(t, "ok", false)
	st := Stand(e.goalRow(t))
	if st.Stop != StopNone || st.Used != 3 || len(st.Candidates) != 1 {
		t.Fatalf("test setup: standing = %+v", st)
	}
	want := Ready{Goal: e.goal, Role: "runner", Act: ActRun, Line: "line-1"}
	if got := lines(e.allowed(t), "runner"); !slices.Equal(got, []Ready{want}) {
		t.Fatalf("runner may %+v, want [%+v]", got, want)
	}

	spec := e.child(t, commit, map[string]string{"goals/" + e.goal + "/specs/spec.md": "spec\n"})
	if _, err := e.run(spec); err != nil {
		t.Fatalf("run on the spec commit: %v", err)
	}
	g := e.shown(t)
	if g.State != tables.StateSelected || g.Select.RowTree != tree || g.Select.Commit != spec {
		t.Fatalf("after the spec run state = %s, select = %+v, want selected on %s", g.State, g.Select, spec)
	}
	if got := lines(e.allowed(t), "runner"); len(got) != 0 {
		t.Fatalf("after the spec run runner may %+v, want nothing", got)
	}
}

// places is the owner's one allowed place: the Goal's whole example set.
func places(e *runEnv) []Ready {
	return []Ready{{Goal: e.goal, Role: "owner", Act: ActPlace}}
}

// placeLines are the owner's place lines among rl.
func placeLines(rl []Ready) []Ready {
	var out []Ready
	for _, l := range lines(rl, "owner") {
		if l.Act == ActPlace {
			out = append(out, l)
		}
	}
	return out
}

// The owner may place the whole set in every state that allows a place, and
// not once the Goal is selected.
func TestAllowedOwnerMayPlaceTheSet(t *testing.T) {
	// written: line-1 holds an example, line-2 and line-3 none
	e := newRunEnv(t, gapLine, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	if g := e.goalRow(t); g.State != tables.StateWritten {
		t.Fatalf("test setup: %s", g.State)
	}
	if got, want := placeLines(e.allowed(t)), places(e); !slices.Equal(got, want) {
		t.Fatalf("written: owner may place %+v, want %+v", got, want)
	}

	// open, with no stop, then with one candidate, two candidates and the select
	o := newRunEnv(t)
	if got, want := placeLines(o.allowed(t)), places(o); !slices.Equal(got, want) {
		t.Fatalf("open: owner may place %+v, want %+v", got, want)
	}
	_, tree := o.metLeaf(t, "ok", true)
	if got := placeLines(o.allowed(t)); len(got) != 0 {
		t.Fatalf("one candidate, selected: owner may place %+v, want nothing", got)
	}
	o.metLeaf(t, "ok2", true)
	if st := Stand(o.goalRow(t)); st.Stop != StopTwoCandidates {
		t.Fatalf("test setup: stop %s", st.Stop)
	}
	if got, want := placeLines(o.allowed(t)), places(o); !slices.Equal(got, want) {
		t.Fatalf("two candidates: owner may place %+v, want %+v", got, want)
	}
	if _, err := Select(o.selectOpts(tree, "owner-1", "owner")); err != nil {
		t.Fatal(err)
	}
	if got := placeLines(o.allowed(t)); len(got) != 0 {
		t.Fatalf("selected: owner may place %+v, want nothing", got)
	}
}
