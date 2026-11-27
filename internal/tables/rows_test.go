package tables

import (
	"encoding/json"
	"fmt"
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/testutil"
)

var evSeq int

// nextID is a fresh event id that sorts after the ones before it.
func nextID() string {
	evSeq++
	return fmt.Sprintf("01M661QG%02d00000000000EVT%03d", evSeq%100, evSeq)
}

// result is a result.record event on goal.
func result(t *testing.T, goal, line, commit, rowTree string, cells ...events.ExampleCell) events.Event {
	t.Helper()
	return resultRefused(t, goal, line, commit, rowTree, "", cells...)
}

func resultRefused(t *testing.T, goal, line, commit, rowTree, refused string, cells ...events.ExampleCell) events.Event {
	t.Helper()
	return resultRead(t, goal, line, commit, rowTree, refused, nil, cells...)
}

// resultAbove is a result.record event whose run read that its commit
// descends from the commits above.
func resultAbove(t *testing.T, goal, line, commit, rowTree string, above []string, cells ...events.ExampleCell) events.Event {
	t.Helper()
	return resultRead(t, goal, line, commit, rowTree, "", &events.CommitRead{Above: above}, cells...)
}

func resultRead(t *testing.T, goal, line, commit, rowTree, refused string, read *events.CommitRead, cells ...events.ExampleCell) events.Event {
	t.Helper()
	p, err := json.Marshal(events.ResultRecord{Role: "runner", Line: line, Commit: commit, RowTree: rowTree, Refused: refused, Cells: cells, Read: read})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-19T00:00:00Z", Actor: "r", Type: events.TypeResultRecord, Node: goal, Payload: p}
}

// cell is a Result's cell for example a of line-1.
func cell(hash, verdict string) events.ExampleCell {
	return events.ExampleCell{Example: "a", Hash: hash, Probe: "p", Output: "o-" + verdict, Stderr: "e", Verdict: verdict}
}

// realGoal is a goal.add event whose H0 is the commit base of the repository.
func realGoal(t *testing.T, r *testutil.Repo, base string, lines ...events.Line) events.Event {
	t.Helper()
	tree, err := gitread.RowTree(r.Dir, base)
	if err != nil {
		t.Fatal(err)
	}
	p, _ := json.Marshal(events.GoalAdd{Role: "owner", Words: "w", Base: base, BaseTree: tree, Budget: 3, Lines: lines})
	return events.Event{V: 2, ID: id1, Actor: "a", Type: events.TypeGoalAdd, Node: id1, Payload: p}
}

// placeReal is a placement whose H0 cell sits on the Goal's H0 commit and tree.
func placeReal(t *testing.T, id, hash, verdict string, g events.Event) events.Event {
	t.Helper()
	var ga events.GoalAdd
	if err := json.Unmarshal(g.Payload, &ga); err != nil {
		t.Fatal(err)
	}
	p, _ := json.Marshal(events.ExamplePlace{Role: "grill", Line: "line-1", Example: "a", Hash: hash,
		H0: events.Cell{Commit: ga.Base, RowTree: ga.BaseTree, Probe: "p", Output: "o-" + verdict, Stderr: "e", Verdict: verdict}})
	return events.Event{V: 2, ID: id, TS: "2026-11-19T00:00:00Z", Actor: "g", Type: events.TypeExamplePlace, Node: g.ID, Payload: p}
}

func rowOf(t *testing.T, rows []Row, tree string) Row {
	t.Helper()
	for _, r := range rows {
		if r.RowTree == tree {
			return r
		}
	}
	t.Fatalf("no row for tree %s in %+v", tree, rows)
	return Row{}
}

func treeOf(t *testing.T, r *testutil.Repo, commit string) string {
	t.Helper()
	tree, err := gitread.RowTree(r.Dir, commit)
	if err != nil {
		t.Fatal(err)
	}
	return tree
}

// fork makes commits A <- B and A <- C on two branches and returns them.
func fork(t *testing.T) (r *testutil.Repo, a, b, c string) {
	t.Helper()
	r = testutil.NewRepo(t)
	a = r.Commit(t, map[string]string{"main.go": "1"})
	b = r.Commit(t, map[string]string{"b.go": "b"})
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "other", a)
	c = r.Commit(t, map[string]string{"c.go": "c"})
	return r, a, b, c
}

// aboveOf is what a run on commit reads of the rows of goal in tb.
func aboveOf(t *testing.T, r *testutil.Repo, tb Tables, goal, commit string) []string {
	t.Helper()
	above, err := tb.Goals[goal].Above(gitread.Ancestry{Dir: r.Dir}, commit)
	if err != nil {
		t.Fatal(err)
	}
	return above
}

func TestRowsLeavesFromAncestry(t *testing.T) {
	r, a, b, c := fork(t)
	g := realGoal(t, r, a, gap)
	evs := []events.Event{g, placeReal(t, idP1, hashA, "failed", g)}
	for _, x := range []struct{ commit, verdict string }{{a, "failed"}, {b, "met"}, {c, "failed"}} {
		above := aboveOf(t, r, Replay(evs), id1, x.commit)
		evs = append(evs, resultAbove(t, id1, "line-1", x.commit, treeOf(t, r, x.commit), above, cell(hashA, x.verdict)))
	}
	tb := Replay(evs)
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	rows := Rows(tb)
	if len(rows) != 3 {
		t.Fatalf("rows = %+v, want 3", rows)
	}
	if row := rowOf(t, rows, treeOf(t, r, b)); !row.Leaf || row.Verdicts["line-1/a"] != "met" {
		t.Fatalf("row B = %+v, want a leaf that is met", row)
	}
	if row := rowOf(t, rows, treeOf(t, r, c)); !row.Leaf || row.Verdicts["line-1/a"] != "failed" {
		t.Fatalf("row C = %+v, want a leaf that failed", row)
	}
	if row := rowOf(t, rows, treeOf(t, r, a)); row.Leaf {
		t.Fatalf("row A (H0) = %+v, want it not a leaf", row)
	}
}

func TestRowNotLeafWhenOneCommitHasDescendant(t *testing.T) {
	r := testutil.NewRepo(t)
	a := r.Commit(t, map[string]string{"main.go": "1"})
	x := r.Commit(t, map[string]string{"x.go": "x"})
	y := r.Commit(t, map[string]string{"goals/G-NXHHVA/note": "n"}) // same row tree as x
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "z", x)
	z := r.Commit(t, map[string]string{"z.go": "z"})
	g := realGoal(t, r, a, gap)
	evs := []events.Event{g, placeReal(t, idP1, hashA, "failed", g)}
	for _, c := range []string{x, y, z} {
		above := aboveOf(t, r, Replay(evs), id1, c)
		evs = append(evs, resultAbove(t, id1, "line-1", c, treeOf(t, r, c), above, cell(hashA, "failed")))
	}
	tb := Replay(evs)
	rows := Rows(tb)
	if row := rowOf(t, rows, treeOf(t, r, z)); !row.Leaf {
		t.Fatalf("row Z = %+v, want a leaf", row)
	}
	row := rowOf(t, rows, treeOf(t, r, x))
	if !slices.Equal(row.Commits, []string{x, y}) {
		t.Fatalf("row X commits = %v, want X and Y", row.Commits)
	}
	if row.Leaf {
		t.Fatalf("row X = %+v: commit X has a descendant Z in another row, so it is not a leaf", row)
	}
}

func TestCellUnderOldPlacementIsHistory(t *testing.T) {
	r, a, b, _ := fork(t)
	g := realGoal(t, r, a, gap)
	rb := treeOf(t, r, b)
	tb := Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g),
		result(t, id1, "line-1", b, rb, cell(hashA, "met")),
	})
	if got := rowOf(t, Rows(tb), rb).Verdicts["line-1/a"]; got != "met" {
		t.Fatalf("verdict under the current placement = %q, want met", got)
	}
	tb = Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g),
		result(t, id1, "line-1", b, rb, cell(hashA, "met")),
		placeReal(t, idP2, hashB, "failed", g),
	})
	rows := Rows(tb)
	if got := rowOf(t, rows, rb).Verdicts["line-1/a"]; got != NotRun {
		t.Fatalf("verdict of a cell under an old placement = %q, want %q", got, NotRun)
	}
	if got := rowOf(t, rows, treeOf(t, r, a)).Verdicts["line-1/a"]; got != "failed" {
		t.Fatalf("H0 verdict under the new placement = %q, want failed from its cell", got)
	}
}

func TestRowKeyedBySubjectTree(t *testing.T) {
	r := testutil.NewRepo(t)
	a := r.Commit(t, map[string]string{"main.go": "1"})
	r.Commit(t, map[string]string{"x.go": "x"})
	x := r.Commit(t, map[string]string{"goals/G-NXHHVA/a": "1"})
	y := r.Commit(t, map[string]string{"goals/G-NXHHVA/b": "2"})
	g := realGoal(t, r, a, gap)
	tx := treeOf(t, r, x)
	if tx != treeOf(t, r, y) {
		t.Fatal("test setup: x and y differ in the subject")
	}
	tb := Replay([]events.Event{
		g, placeReal(t, idP1, hashA, "failed", g),
		result(t, id1, "line-1", x, tx, cell(hashA, "failed")),
		result(t, id1, "line-1", y, tx, cell(hashA, "failed")),
	})
	rows := Rows(tb)
	row := rowOf(t, rows, tx)
	if !slices.Equal(row.Commits, []string{x, y}) {
		t.Fatalf("commits = %v, want x and y in one row", row.Commits)
	}
}

func TestRunOnH0DoesNotOpenAGoal(t *testing.T) {
	// line-2 has no example yet, so the Goal is written and a Draft replace may
	// move H0 and drop that line.
	first := []events.Event{goalAdd(t, id1, "", "w", gap, guard), placeEv(t, idP1, "line-1", "a", hashA, "failed")}
	p, _ := json.Marshal(events.GoalAdd{Role: "owner", Words: "w", Base: "c2", BaseTree: "t2", Budget: 3, Lines: []events.Line{gap}})
	replace := events.Event{V: 2, ID: id2, Actor: "a", Type: events.TypeGoalAdd, Node: id1, Payload: p}
	tb := Replay(append(first, replace))
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	// The placement's cell is on the old row tree, so the example reads not run.
	g := tb.Goals[id1]
	if g.State != StateWritten || g.H0Verdict("line-1/a") != NotRun {
		t.Fatalf("after the replace: state %s, verdict %s; want written and %q", g.State, g.H0Verdict("line-1/a"), NotRun)
	}
	// A run on H0 shows its cell on H0's row, but a Result is in the trace of
	// the clone that ran it, so it opens no Goal: the state reads git alone.
	tb = Replay(append(first, replace, result(t, id1, "line-1", "c2", "t2", cell(hashA, "failed"))))
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	if g = tb.Goals[id1]; g.State != StateWritten || g.H0Verdict("line-1/a") != "failed" {
		t.Fatalf("after run on H0: state %s, verdict %s; want written and failed", g.State, g.H0Verdict("line-1/a"))
	}
	// The grill places the same version again, measured on the new H0: git
	// holds its verdict there, and the Goal opens with its version kept.
	again := placeOn(t, idP3, id1, "g", "grill", "line-1", "a", hashA, "t2", "failed")
	tb = Replay(append(first, replace, again))
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g = tb.Goals[id1]
	if cur, _ := g.Current("line-1/a"); g.State != StateOpen || cur.Version != 1 {
		t.Fatalf("after the place again: state %s, version %d; want open and 1", g.State, cur.Version)
	}
}

func TestRowsReadNoGit(t *testing.T) {
	// The commits are in no clone: a row's leaf mark is what the Results
	// carry, so removing history changes no row.
	g := goalAdd(t, id1, "", "w", gap)
	tb := Replay([]events.Event{
		g, placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		resultAbove(t, id1, "line-1", "c1", "t1", []string{"c"}, cell(hashA, "met")),
		resultAbove(t, id1, "line-1", "c2", "t2", []string{"c", "c1"}, cell(hashA, "met")),
		resultAbove(t, id1, "line-1", "c3", "t3", []string{"c"}, cell(hashA, "met")),
	})
	rows := Rows(tb)
	if row := rowOf(t, rows, "t1"); row.Leaf {
		t.Fatalf("row t1 = %+v, want not a leaf: c2's run read that it descends from c1", row)
	}
	for _, tree := range []string{"t2", "t3"} {
		if row := rowOf(t, rows, tree); !row.Leaf {
			t.Fatalf("row %s = %+v, want a leaf", tree, row)
		}
	}
}
