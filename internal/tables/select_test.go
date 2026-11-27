package tables

import (
	"encoding/json"
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/testutil"
)

// selectEv is a select event on the Goal id1.
func selectEv(t *testing.T, actor, role, commit, rowTree string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Select{Role: role, Commit: commit, RowTree: rowTree})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-19T00:00:00Z", Actor: actor, Type: events.TypeSelect, Node: id1, Payload: p}
}

func TestCandidateIsLeafMetOnEveryLine(t *testing.T) {
	r, a, b, c := fork(t)
	g := realGoal(t, r, a, gap)
	tb := Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g),
		result(t, id1, "line-1", b, treeOf(t, r, b), cell(hashA, "met")),
		result(t, id1, "line-1", c, treeOf(t, r, c), cell(hashA, "failed")),
	})
	cands := Candidates(tb.Goals[id1])
	if len(cands) != 1 || cands[0].RowTree != treeOf(t, r, b) {
		t.Fatalf("candidates = %+v, want only the leaf row B that is met", cands)
	}
	// B met but no longer a leaf once a commit below it is a row of the Goal.
	testutil.Git(t, r.Dir, "checkout", "-q", "main")
	d := r.Commit(t, map[string]string{"d.go": "d"})
	tb = Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g),
		result(t, id1, "line-1", b, treeOf(t, r, b), cell(hashA, "met")),
		resultAbove(t, id1, "line-1", d, treeOf(t, r, d), []string{a, b}, cell(hashA, "failed")),
	})
	if cands := Candidates(tb.Goals[id1]); len(cands) != 0 {
		t.Fatalf("candidates = %+v, want none: B is not a leaf", cands)
	}
}

func TestNotCandidateWithUnrunLine(t *testing.T) {
	r, a, b, _ := fork(t)
	g := realGoal(t, r, a, gap, guard)
	// line-2/a is placed, but B has only a line-1 cell.
	p2, _ := json.Marshal(events.ExamplePlace{Role: "grill", Line: "line-2", Example: "a", Hash: hashB,
		H0: events.Cell{Commit: a, RowTree: treeOf(t, r, a), Probe: "p", Output: "o", Stderr: "e", Verdict: "met"}})
	place2 := events.Event{V: 2, ID: idP2, TS: "2026-11-19T00:00:00Z", Actor: "g", Type: events.TypeExamplePlace, Node: g.ID, Payload: p2}
	tb := Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g), place2,
		result(t, id1, "line-1", b, treeOf(t, r, b), cell(hashA, "met")),
	})
	goal := tb.Goals[id1]
	if len(goal.Placed()) != 2 {
		t.Fatalf("test setup: placed = %+v", goal.Placed())
	}
	// once line-2 is run on B too, B is a candidate: the test sees the rule, not an empty answer
	run2 := result(t, id1, "line-2", b, treeOf(t, r, b), cell(hashB, "met"))
	run2.Payload = []byte(`{"role":"runner","line":"line-2","commit":"` + b + `","row_tree":"` + treeOf(t, r, b) + `","refused":"","cells":[{"example":"a","hash":"` + hashB + `","probe":"p","output":"o","stderr":"e","exit":0,"verdict":"met"}]}`)
	both := Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g), place2,
		result(t, id1, "line-1", b, treeOf(t, r, b), cell(hashA, "met")), run2,
	})
	if cands := Candidates(both.Goals[id1]); len(cands) != 1 {
		t.Fatalf("candidates = %+v, want B once both lines are run", cands)
	}
	if cands := Candidates(goal); len(cands) != 0 {
		t.Fatalf("candidates = %+v, want none while line-2 is not run on B", cands)
	}
}

func TestSelectMakesGoalSelected(t *testing.T) {
	_, g := openGoal(t, run(t, "r1", "met"))
	if g.State != StateOpen || g.Select != nil {
		t.Fatalf("test setup: %+v", g)
	}
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		run(t, "r1", "met"), selectEv(t, "a", "owner", "commit-r1", "r1"),
	})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	got := tb.Goals[id1]
	if got.State != StateSelected || got.Select == nil || got.Select.Commit != "commit-r1" || got.Select.RowTree != "r1" {
		t.Fatalf("Goal = %+v, want selected with the select's commit and row", got)
	}
}

func TestReplaySelectGuardsAgain(t *testing.T) {
	base := func() []events.Event {
		return []events.Event{goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed"), run(t, "r1", "met"), run(t, "r2", "failed")}
	}
	cases := []struct {
		name string
		ev   events.Event
		keep bool
	}{
		{"owner of the Goal selects a met row", selectEv(t, "a", "owner", "commit-r1", "r1"), true},
		{"runner selects a met row", selectEv(t, "x", "runner", "commit-r1", "r1"), true},
		// Candidate reads the Results of the clone's trace, which another
		// clone lacks: the act checks it, and replay keeps the select
		{"a row that is not met", selectEv(t, "a", "runner", "commit-r2", "r2"), true},
		{"a row no Result names", selectEv(t, "a", "runner", "commit-r9", "r9"), true},
		{"owner who did not add the Goal", selectEv(t, "other", "owner", "commit-r1", "r1"), false},
		{"a role that selects nothing", selectEv(t, "a", "grill", "commit-r1", "r1"), false},
	}
	for _, c := range cases {
		tb := Replay(append(base(), c.ev))
		if len(tb.RefusedAtReplay) != 0 && c.keep {
			t.Errorf("%s: refused at replay: %+v", c.name, tb.RefusedAtReplay)
		}
		if got := tb.Goals[id1].State == StateSelected; got != c.keep {
			t.Errorf("%s: selected = %v, want %v (refused: %+v)", c.name, got, c.keep, tb.RefusedAtReplay)
		}
	}
	// a second select is refused: the first stands
	tb := Replay(append(base(), selectEv(t, "a", "owner", "commit-r1", "r1"), selectEv(t, "a", "owner", "commit-r1", "r1")))
	if len(tb.RefusedAtReplay) != 1 || tb.Goals[id1].Select == nil {
		t.Fatalf("refused = %+v, want the second select refused", tb.RefusedAtReplay)
	}
}

func TestUnstableKeepsBothOutputs(t *testing.T) {
	first := result(t, id1, "line-1", "c1", "r1", events.ExampleCell{Example: "a", Hash: hashA, Probe: "p", Output: "out-first", Stderr: "e", Verdict: "met"})
	second := resultRefused(t, id1, "line-1", "c1", "r1", "Stable", events.ExampleCell{Example: "a", Hash: hashA, Probe: "p", Output: "out-second", Stderr: "e", Verdict: "failed"})
	p, _ := json.Marshal(events.ResultRecord{Role: "runner", Line: "line-1", Commit: "c1", RowTree: "r1", Refused: "Stable", Differing: []string{"line-1/a"}, Cells: []events.ExampleCell{{Example: "a", Hash: hashA, Probe: "p", Output: "out-second", Stderr: "e", Verdict: "failed"}}})
	second.Payload = p
	later := retract(t, first) // a retract after the refusal must not change what it kept
	_, g := openGoal(t, first, second, later)
	var got []UnstableCell
	for _, r := range g.Results {
		got = append(got, r.Unstable...)
	}
	want := []UnstableCell{{Key: "line-1/a", Tree: "r1", First: "out-first", Second: "out-second"}}
	if !slices.Equal(got, want) {
		t.Fatalf("unstable = %+v, want %+v", got, want)
	}
}
