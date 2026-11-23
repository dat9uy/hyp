package tables

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// unselectEv is an unselect event on the Goal id1.
func unselectEv(t *testing.T, actor, role string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Unselect{Role: role})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-20T00:00:00Z", Actor: actor, Type: events.TypeUnselect, Node: id1, Payload: p}
}

func landEv(t *testing.T, stop string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Land{Role: "owner", Stop: stop, Cause: "x"})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-20T00:00:00Z", Actor: "a", Type: events.TypeLand, Node: id1, Payload: p}
}

// selectedLog is an open Goal whose row r1 is met and selected by the runner.
func selectedLog(t *testing.T) []events.Event {
	return []events.Event{
		goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		run(t, "r1", "met"), selectEv(t, "x", "runner", "commit-r1", "r1"),
	}
}

func TestUnselectOpensTheGoalAndRestartsTheBudget(t *testing.T) {
	tb := Replay(append(selectedLog(t), unselectEv(t, "a", "owner")))
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if g.State != StateOpen || g.Select != nil || len(g.Unselects) != 1 || g.Unselects[0].Select.RowTree != "r1" {
		t.Fatalf("Goal = %+v, want open with one unselect of r1", g)
	}
	if RunsUsed(g) != 0 {
		t.Fatalf("runs used = %d, want 0 after the unselect", RunsUsed(g))
	}
	if !g.metEverywhere("r1") || len(g.Results) != 1 || g.Results[0].Retracted {
		t.Fatal("the old row lost its cells")
	}
}

func TestReplayUnselectGuardsAgain(t *testing.T) {
	open := []events.Event{goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed"), run(t, "r1", "met")}
	cases := []struct {
		name  string
		evs   []events.Event
		guard string
	}{
		{"an open Goal never selected", append(open, unselectEv(t, "a", "owner")), "Selected"},
		{"a runner", append(selectedLog(t), unselectEv(t, "x", "runner")), "Owner unselects"},
		{"the grill", append(selectedLog(t), unselectEv(t, "g", "grill")), "Owner unselects"},
		{"an owner who did not add the Goal", append(selectedLog(t), unselectEv(t, "other", "owner")), "Goal's actor"},
		{"a second unselect", append(selectedLog(t), unselectEv(t, "a", "owner"), unselectEv(t, "a", "owner")), "Selected"},
		{"a landed Goal", append(selectedLog(t), landEv(t, StopSelected), unselectEv(t, "a", "owner")), "Not ended"},
	}
	for _, c := range cases {
		tb := Replay(c.evs)
		if len(tb.RefusedAtReplay) != 1 {
			t.Errorf("%s: refused = %+v, want the unselect refused", c.name, tb.RefusedAtReplay)
			continue
		}
		if got := guardOf(tb.RefusedAtReplay[0].Err); got != c.guard {
			t.Errorf("%s: refused by %q, want %q", c.name, got, c.guard)
		}
	}
}

// reselected replays evs after the selected log and its unselect, and
// reports whether New cell lets the runner select again. Replay does not
// check New cell, which reads the Results of the clone's trace: the select
// act does.
func reselected(t *testing.T, _ string, evs ...events.Event) bool {
	t.Helper()
	all := append(append(selectedLog(t), unselectEv(t, "a", "owner")), evs...)
	tb := Replay(all)
	return GuardNewCell(tb.Goals[id1], "runner") == nil
}

func TestRunnerReselectsOnlyAfterANewCell(t *testing.T) {
	if reselected(t, "r1") {
		t.Error("the runner reselected at once")
	}
	if reselected(t, "r1", run(t, "r1", "met")) {
		t.Error("a rerun of the same version on the old row counted as a new cell")
	}
	if !reselected(t, "r3", run(t, "r3", "met")) {
		t.Error("a run on a new row did not count")
	}
	v2 := placeOn(t, nextID(), id1, "g", "grill", "line-1", "a", hashB, "t", "failed")
	if !reselected(t, "r1", v2, run2(t, "r1", "a", hashB)) {
		t.Error("a new version run on the old row did not count")
	}
	// a new key whose folder has the bytes of an old one is a new version
	b := placeOn(t, nextID(), id1, "g", "grill", "line-1", "b", hashA, "t", "failed")
	if !reselected(t, "r1", b, run2(t, "r1", "b", hashA)) {
		t.Error("a new example key with an old hash did not count")
	}
}

func TestRunnerReselectRefusedForAnOldVersionOnAnOldRow(t *testing.T) {
	// rows r1 and r2 run version 1; version 2 runs on r1 only and r1 is selected
	v2 := placeOn(t, nextID(), id1, "g", "grill", "line-1", "a", hashB, "t", "failed")
	all := []events.Event{
		goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		run(t, "r2", "failed"), run(t, "r1", "met"), v2, run2(t, "r1", "a", hashB),
		selectEv(t, "x", "runner", "commit-r1", "r1"), unselectEv(t, "a", "owner"),
		run2(t, "r2", "a", hashB), // the pair is new; its version and its row are not
	}
	g := Replay(all).Goals[id1]
	if err := GuardNewCell(g, "runner"); err == nil || guardOf(err) != "New cell" {
		t.Fatalf("GuardNewCell = %v, want the runner's select refused by New cell", err)
	}
	// the owner selects at any time
	if err := GuardNewCell(g, "owner"); err != nil {
		t.Fatalf("the owner's select was refused: %v", err)
	}
	// replay does not check New cell: a select the act recorded stands
	all = append(all, selectEv(t, "x", "runner", "commit-r1", "r1"))
	if tb := Replay(all); tb.Goals[id1].State != StateSelected || len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("replay refused the select: %+v", tb.RefusedAtReplay)
	}
}

// run2 is a Result on line-1 of row tree with one cell of example ex.
func run2(t *testing.T, tree, ex, hash string) events.Event {
	t.Helper()
	c := cell(hash, "met")
	c.Example = ex
	return result(t, id1, "line-1", "commit-"+tree, tree, c)
}

func guardOf(err error) string { return string(fault.GuardOf(err)) }
