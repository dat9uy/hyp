package tables

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// closeEv is a close event on the Goal id1 that names the snapshot seen.
func closeEv(t *testing.T, actor, role string, seen ...string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Close{Role: role, Seen: seen})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-20T00:00:00Z", Actor: actor, Type: events.TypeClose, Node: id1, Payload: p}
}

// landSeen is a land event on the Goal id1 at the stop selected, by the
// select sel, that names the snapshot seen.
func landSeen(t *testing.T, sel string, seen ...string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.Land{Role: "owner", Stop: StopSelected, Cause: sel, Seen: seen})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: nextID(), TS: "2026-11-20T00:00:00Z", Actor: "a", Type: events.TypeLand, Node: id1, Payload: p}
}

func ids(evs ...events.Event) []string {
	var out []string
	for _, ev := range evs {
		out = append(out, ev.ID)
	}
	return out
}

func TestCloseEndsAnOpenGoal(t *testing.T) {
	add, place := goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed")
	tb := Replay([]events.Event{add, place, closeEv(t, "a", "owner", ids(add, place)...)})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if g.State != StateEnded || g.Close == nil || g.Land != nil {
		t.Fatalf("Goal = %+v, want ended by its close", g)
	}
}

func TestReplayCloseGuardsAgain(t *testing.T) {
	open := func() []events.Event {
		return []events.Event{goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed")}
	}
	cases := []struct {
		name  string
		evs   []events.Event
		guard string
	}{
		{"a runner's close", append(open(), closeEv(t, "r", "runner")), string(fault.GuardOwnerClose)},
		{"an owner who did not add the Goal", append(open(), closeEv(t, "someone", "owner")), string(fault.GuardGoalsActor)},
		{"a second close", append(open(), closeEv(t, "a", "owner"), closeEv(t, "a", "owner")), string(fault.GuardNotEnded)},
		{"a run after the close", append(open(), closeEv(t, "a", "owner"), run(t, "r1", "met")), string(fault.GuardNotEnded)},
		{"a land after the close", append(append(open(), run(t, "r1", "met")), closeEv(t, "a", "owner"),
			landEv(t, StopBudgetUsed)), string(fault.GuardNotEnded)},
	}
	for _, c := range cases {
		tb := Replay(c.evs)
		if len(tb.RefusedAtReplay) != 1 || guardOf(tb.RefusedAtReplay[0].Err) != c.guard {
			t.Fatalf("%s: refused %+v, want one refusal by %s", c.name, tb.RefusedAtReplay, c.guard)
		}
	}
}

func TestEndHoldsItsSnapshotAfterAMerge(t *testing.T) {
	add, place := goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed")
	// clone B's run sorts in ahead of A's close, which did not see it
	runB := run(t, "rB", "failed")
	tb := Replay([]events.Event{add, place, runB, closeEv(t, "a", "owner", ids(add, place)...)})
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != runB.ID ||
		guardOf(tb.RefusedAtReplay[0].Err) != string(fault.GuardNotEnded) {
		t.Fatalf("refused at replay: %+v, want B's run by Not ended", tb.RefusedAtReplay)
	}
	if g := tb.Goals[id1]; len(g.Results) != 0 || g.State != StateEnded {
		t.Fatalf("Goal = %+v, want ended with no Result", g)
	}

	// a land names its snapshot the same way
	runA, sel := run(t, "r1", "met"), selectEv(t, "x", "runner", "commit-r1", "r1")
	runB = run(t, "rB", "failed")
	tb = Replay([]events.Event{add, place, runB, runA, sel, landSeen(t, sel.ID, ids(add, place, runA, sel)...)})
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != runB.ID {
		t.Fatalf("refused at replay: %+v, want B's run", tb.RefusedAtReplay)
	}
	if g := tb.Goals[id1]; len(g.Results) != 1 || g.Results[0].ID != runA.ID || g.Land == nil {
		t.Fatalf("Goal = %+v, want landed with only A's run", g)
	}
}

func TestEndThatReplayRefusesNamesNoSnapshot(t *testing.T) {
	add, place := goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed")
	runB := run(t, "rB", "failed")
	// the runner's close fails Owner closes, so it holds the Goal to nothing
	tb := Replay([]events.Event{add, place, runB, closeEv(t, "r", "runner", ids(add, place)...)})
	if len(tb.RefusedAtReplay) != 1 || guardOf(tb.RefusedAtReplay[0].Err) != string(fault.GuardOwnerClose) {
		t.Fatalf("refused at replay: %+v, want only the runner's close", tb.RefusedAtReplay)
	}
	if g := tb.Goals[id1]; len(g.Results) != 1 || g.State == StateEnded {
		t.Fatalf("Goal = %+v, want open with B's run", g)
	}
}

func TestLandWithNoSnapshotRefusesNothing(t *testing.T) {
	// a land written before close existed names no snapshot
	tb := Replay(append(selectedLog(t), landEv(t, StopSelected)))
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	if g := tb.Goals[id1]; g.State != StateEnded || len(g.Results) != 1 {
		t.Fatalf("Goal = %+v, want ended with its run", g)
	}
}
