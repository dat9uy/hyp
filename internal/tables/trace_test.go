package tables

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
)

// placeInGit is an examples.place as hyp writes it now: the hashes and the
// row tree of H0 the set ran on, with no cell.
func placeInGit(t *testing.T, id, rowTree string, xs ...[3]string) events.Event {
	t.Helper()
	p := events.ExamplesPlace{Role: "grill", Commit: "c", RowTree: rowTree}
	for _, x := range xs {
		p.Examples = append(p.Examples, events.PlacedExample{Line: x[0], Example: x[1], Hash: x[2]})
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: id, Actor: "g", Type: events.TypeExamplesPlace, Node: id1, Payload: b}
}

// selectNaming is a select by the runner that names the examples it was
// judged by.
func selectNaming(t *testing.T, rowTree string, judged ...events.Judged) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Select{Role: "runner", Commit: "commit-" + rowTree, RowTree: rowTree, Examples: judged})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), Actor: "r", Type: events.TypeSelect, Node: id1, Payload: p}
}

// A clone whose trace holds no cell reads the Goal from git alone: the place
// opens it, because the place checked Kind on H0's row tree, and a select
// that names the current examples selects it with no Result in the clone.
func TestGitAloneOpensAndSelects(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap)
	place := placeInGit(t, idP1, "t", [3]string{"line-1", "a", hashA})
	tb := Replay([]events.Event{goal, place})
	if g := tb.Goals[id1]; g.State != StateOpen {
		t.Fatalf("state = %s, want open from the place alone", g.State)
	}
	tb = Replay([]events.Event{goal, place, selectNaming(t, "r1", events.Judged{Key: "line-1/a", Hash: hashA})})
	g := tb.Goals[id1]
	if len(tb.RefusedAtReplay) != 0 || g.State != StateSelected || len(g.Select.Judged) != 1 {
		t.Fatalf("Goal = %+v, refused %+v; want selected with no Result", g, tb.RefusedAtReplay)
	}
	if RunsUsed(g) != 0 {
		t.Fatalf("runs used = %d, want 0: no Result is in this clone", RunsUsed(g))
	}
}

// A place on a row tree that is no longer H0's opens nothing until a cell
// on H0's row says so.
func TestGitPlaceOnAnOldH0DoesNotOpen(t *testing.T) {
	tb := Replay([]events.Event{goalAdd(t, id1, "", "w", gap), placeInGit(t, idP1, "old", [3]string{"line-1", "a", hashA})})
	if g := tb.Goals[id1]; g.State != StateWritten {
		t.Fatalf("state = %s, want written", g.State)
	}
}

// A select that names other examples than the current placements is refused
// at replay in every clone.
func TestSelectNamingOtherExamplesIsRefused(t *testing.T) {
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap), placeInGit(t, idP1, "t", [3]string{"line-1", "a", hashA}),
		selectNaming(t, "r1", events.Judged{Key: "line-1/a", Hash: hashB}),
	})
	if len(tb.RefusedAtReplay) != 1 || tb.Goals[id1].State != StateOpen {
		t.Fatalf("refused %+v, state %s; want the select refused and the Goal open", tb.RefusedAtReplay, tb.Goals[id1].State)
	}
}

// The trace's saved cells are live on their row: the place's own first, then
// a cell a try ran again, which is marked.
func TestSavedCellsAreLive(t *testing.T) {
	save := func(again bool, verdict string) events.Event {
		p, err := json.Marshal(events.CellsSave{Role: "grill", Commit: "c", RowTree: "t", Again: again,
			Cells: []events.SavedCell{{Line: "line-1", Example: "a", Hash: hashA, Output: "o", Stderr: "e", Verdict: verdict}}})
		if err != nil {
			t.Fatal(err)
		}
		return events.Event{V: 2, ID: nextID(), Actor: "g", Type: events.TypeCellsSave, Node: id1, Payload: p}
	}
	goal := goalAdd(t, id1, "", "w", gap)
	place := placeInGit(t, idP1, "t", [3]string{"line-1", "a", hashA})
	g := Replay([]events.Event{goal, place, save(true, "failed")}).Goals[id1]
	if c, ok := g.SavedCell("line-1/a", "t"); !ok || !c.Again || g.H0Verdict("line-1/a") != "failed" {
		t.Fatalf("saved = %+v %v, want the cell run again", c, ok)
	}
	g = Replay([]events.Event{goal, place, save(false, "failed"), save(true, "met")}).Goals[id1]
	if c, ok := g.SavedCell("line-1/a", "t"); !ok || c.Again || g.H0Verdict("line-1/a") != "failed" {
		t.Fatalf("saved = %+v %v, want the place's own cell first", c, ok)
	}
}
