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

func (e *runEnv) allowed(t *testing.T) []Ready {
	t.Helper()
	_, tb := e.log(t)
	return AllowedFor(tb.ByID(e.goal), e.root)
}

func TestAllowedWrittenGoalGrillMayPlace(t *testing.T) {
	e := newRunEnv(t, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	if g := e.goalRow(t); g.State != tables.StateWritten {
		t.Fatalf("test setup: %s", g.State)
	}
	rl := e.allowed(t)
	if want := (Ready{Goal: e.goal, Role: "grill", Act: ActPlace, Line: "line-2"}); !slices.Contains(lines(rl, "grill"), want) {
		t.Fatalf("grill may %+v, want %+v", lines(rl, "grill"), want)
	}
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %+v, want nothing on a Goal that is not open", got)
	}
	if got := lines(rl, "owner"); len(got) != 1 || got[0] != (Ready{Goal: e.goal, Role: "owner", Act: ActAddGoal}) {
		t.Fatalf("owner may %+v, want the Draft replace", got)
	}
}

func TestAllowedOpenGoalRunnerMayRun(t *testing.T) {
	e := newRunEnv(t)
	rl := e.allowed(t)
	want := Ready{Goal: e.goal, Role: "runner", Act: ActRun, Line: "line-1"}
	if got := lines(rl, "runner"); !slices.Equal(got, []Ready{want}) {
		t.Fatalf("runner may %+v, want [%+v]", got, want)
	}
	if got := lines(rl, "owner"); len(got) != 0 {
		t.Fatalf("owner may %+v, want nothing on an open Goal with no stop", got)
	}
	for _, l := range rl {
		if !slices.Contains([]string{ActAddGoal, ActPlace, ActRun, ActSelect}, l.Act) {
			t.Fatalf("allowed %+v, an act the binary lacks", l)
		}
	}
}

func TestAllowedOneCandidateRunnerMaySelect(t *testing.T) {
	e := newRunEnv(t)
	_, tree := e.metLeaf(t, "ok", true)
	want := Ready{Goal: e.goal, Role: "runner", Act: ActSelect, Tree: tree}
	if got := lines(e.allowed(t), "runner"); !slices.Contains(got, want) {
		t.Fatalf("runner may %+v, want %+v", got, want)
	}
	// a candidate whose commit holds no spec is not offered: run on the spec commit first
	e2 := newRunEnv(t)
	e2.metLeaf(t, "ok", false)
	if got := lines(e2.allowed(t), "runner"); acted(got, ActSelect) {
		t.Fatalf("offered %+v without a spec", got)
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
	if st := Stand(e.goalRow(t), e.root); st.Stop != StopTwoCandidates || len(st.Candidates) != 2 {
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
		st := Stand(e.goalRow(t), e.root)
		if i < 2 && st.Stop != StopNone {
			t.Fatalf("after %d trees stop = %s", i+1, st.Stop)
		}
	}
	rl := e.allowed(t)
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %+v after the budget is used", got)
	}
	if st := Stand(e.goalRow(t), e.root); st.Stop != StopBudgetUsed || st.Used != 3 || st.Waits != "owner" {
		t.Fatalf("standing = %+v", st)
	}
}

// places are the grill's allowed places, one per line key, in order.
func places(e *runEnv, keys ...string) []Ready {
	var out []Ready
	for _, k := range keys {
		out = append(out, Ready{Goal: e.goal, Role: "grill", Act: ActPlace, Line: k})
	}
	return out
}

// The grill may place on every line in every state that allows a place, a
// line that holds an example included, and on no line once the Goal is
// selected.
func TestAllowedGrillMayPlaceOnEveryLine(t *testing.T) {
	// written: line-1 holds an example, line-2 and line-3 none
	e := newRunEnv(t, gapLine, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	if g := e.goalRow(t); g.State != tables.StateWritten {
		t.Fatalf("test setup: %s", g.State)
	}
	if got, want := lines(e.allowed(t), "grill"), places(e, "line-1", "line-2", "line-3"); !slices.Equal(got, want) {
		t.Fatalf("written: grill may %+v, want %+v", got, want)
	}

	// open, with no stop, then with one candidate, two candidates and the select
	o := newRunEnv(t)
	if got, want := lines(o.allowed(t), "grill"), places(o, "line-1"); !slices.Equal(got, want) {
		t.Fatalf("open: grill may %+v, want %+v", got, want)
	}
	_, tree := o.metLeaf(t, "ok", true)
	if got, want := lines(o.allowed(t), "grill"), places(o, "line-1"); !slices.Equal(got, want) {
		t.Fatalf("one candidate: grill may %+v, want %+v", got, want)
	}
	o.metLeaf(t, "ok2", true)
	if st := Stand(o.goalRow(t), o.root); st.Stop != StopTwoCandidates {
		t.Fatalf("test setup: stop %s", st.Stop)
	}
	if got, want := lines(o.allowed(t), "grill"), places(o, "line-1"); !slices.Equal(got, want) {
		t.Fatalf("two candidates: grill may %+v, want %+v", got, want)
	}
	if _, err := Select(o.selectOpts(tree, "owner-1", "owner")); err != nil {
		t.Fatal(err)
	}
	if got := lines(o.allowed(t), "grill"); len(got) != 0 {
		t.Fatalf("selected: grill may %+v, want nothing", got)
	}
}
