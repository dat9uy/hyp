package acts

import (
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/tables"
)

// lines are the ready lines of role.
func lines(rl []ReadyLine, role string) []string {
	var out []string
	for _, l := range rl {
		if l.Role == role {
			out = append(out, l.Line)
		}
	}
	return out
}

func (e *runEnv) allowed(t *testing.T) []ReadyLine {
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
	if want := "hyp place " + e.goal + " line-2/<example>"; !slices.Contains(lines(rl, "grill"), want) {
		t.Fatalf("grill may %q, want %q", lines(rl, "grill"), want)
	}
	if got := lines(rl, "runner"); len(got) != 0 {
		t.Fatalf("runner may %q, want nothing on a Goal that is not open", got)
	}
	if got := lines(rl, "owner"); len(got) != 1 || !strings.HasPrefix(got[0], "hyp add goal "+e.goal+" ") {
		t.Fatalf("owner may %q, want the Draft replace", got)
	}
}

func TestAllowedOpenGoalRunnerMayRun(t *testing.T) {
	e := newRunEnv(t)
	rl := e.allowed(t)
	want := "hyp run <commit> " + e.goal + "/line-1"
	if got := lines(rl, "runner"); !slices.Equal(got, []string{want}) {
		t.Fatalf("runner may %q, want [%q]", got, want)
	}
	if got := lines(rl, "owner"); len(got) != 0 {
		t.Fatalf("owner may %q, want nothing on an open Goal with no stop", got)
	}
	for _, l := range rl {
		for _, banned := range []string{"land", "reset", "retract"} {
			if strings.HasPrefix(l.Line, "hyp "+banned) {
				t.Fatalf("printed %q, an act the binary lacks", l.Line)
			}
		}
	}
}

func TestAllowedOneCandidateRunnerMaySelect(t *testing.T) {
	e := newRunEnv(t)
	_, tree := e.metLeaf(t, "ok", true)
	want := "hyp select " + e.goal + " " + tree
	if got := lines(e.allowed(t), "runner"); !slices.Contains(got, want) {
		t.Fatalf("runner may %q, want %q", got, want)
	}
	// a candidate whose commit holds no spec is not offered: run on the spec commit first
	e2 := newRunEnv(t)
	e2.metLeaf(t, "ok", false)
	for _, l := range lines(e2.allowed(t), "runner") {
		if strings.HasPrefix(l, "hyp select") {
			t.Fatalf("offered %q without a spec", l)
		}
	}
}

func TestAllowedTwoCandidatesOnlyOwnerMaySelect(t *testing.T) {
	e := newRunEnv(t)
	_, t1 := e.metLeaf(t, "ok", true)
	_, t2 := e.metLeaf(t, "ok2", true)
	rl := e.allowed(t)
	for _, l := range lines(rl, "runner") {
		if strings.HasPrefix(l, "hyp select") || strings.HasPrefix(l, "hyp run") {
			t.Fatalf("runner may %q at a stop", l)
		}
	}
	owner := lines(rl, "owner")
	for _, tree := range []string{t1, t2} {
		if !slices.Contains(owner, "hyp select "+e.goal+" "+tree) {
			t.Fatalf("owner may %q, want a select of %s", owner, tree)
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
		t.Fatalf("runner may %q after the budget is used", got)
	}
	if st := Stand(e.goalRow(t), e.root); st.Stop != StopBudgetUsed || st.Used != 3 || st.Waits != "owner" {
		t.Fatalf("standing = %+v", st)
	}
}
