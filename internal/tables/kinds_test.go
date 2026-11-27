package tables

import (
	"reflect"
	"runtime"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/spec"
	"github.com/dat9uy/hyp/internal/testutil"
)

// TestEveryKindHasOneReplayEntry: every kind of events.Kinds has a replay
// entry with a handler, a kind that changes a Goal's state above all, and
// with the type its payload decodes into; no
// entry names a type outside the closed set; an entry names a frozen Goal
// exactly when its kind's entry says an end freezes it; and only land and
// close name a snapshot.
func TestEveryKindHasOneReplayEntry(t *testing.T) {
	for _, k := range events.Kinds {
		r, ok := kindReplays[k.Type]
		if !ok || r.apply == nil {
			if k.Changes {
				t.Errorf("kind %s changes a Goal's state and has no replay handler", k.Type)
			} else {
				t.Errorf("kind %s has no replay handler", k.Type)
			}
			continue
		}
		if r.payload == nil {
			t.Errorf("kind %s names no payload type", k.Type)
		}
		if (r.goal != nil) != k.Frozen {
			t.Errorf("kind %s: its replay entry names a frozen Goal: %v, its kind entry says frozen: %v", k.Type, r.goal != nil, k.Frozen)
		}
		ends := k.Type == events.TypeLand || k.Type == events.TypeClose
		if (r.end != nil) != ends {
			t.Errorf("kind %s: its replay entry names a snapshot: %v, want %v", k.Type, r.end != nil, ends)
		}
	}
	for typ := range kindReplays {
		if _, ok := events.KindOf(typ); !ok {
			t.Errorf("a replay entry names type %q, which is outside the closed set", typ)
		}
	}
}

// funcName is the full name of the function f, as the runtime prints it.
func funcName(f any) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// guardSet is the set of the guards of refusals.
func guardSet(t *testing.T, refusals []testutil.Refusal) map[fault.Guard]bool {
	t.Helper()
	out := map[fault.Guard]bool{}
	for _, r := range refusals {
		if !r.GuardConst {
			t.Errorf("%s: replay refuses under a guard that is no constant, %s", r.At, r.Guard)
			continue
		}
		out[fault.Guard(r.Guard)] = true
	}
	return out
}

// TestReplayEntriesNameTheGuardsReplayChecks: the guards of each kind's
// replay entry are exactly those its handler's code can refuse an event
// under, with the form fault of a payload that does not decode for every
// kind and the snapshot's Not ended for a kind an end freezes; and
// ChecksAtReplay, which hyp spec prints, says yes for a guard of the spec
// exactly when the code replay runs can refuse under it. The guards are read
// from the source of every function replay can reach.
func TestReplayEntriesNameTheGuardsReplayChecks(t *testing.T) {
	code := testutil.ReadCode(t)
	snapshot := guardSet(t, code.Reachable(t, funcName(guardInSnapshot)))
	if !snapshot[fault.GuardNotEnded] {
		t.Fatalf("the snapshot check refuses under %v, not Not ended", snapshot)
	}
	// decodePayloads decodes every kind's payload with events.Decode, and
	// replayWith refuses an event under what it returned
	decode := guardSet(t, code.Reachable(t, funcName(events.Decode)))
	if !decode[fault.NoGuard] || len(decode) != 1 {
		t.Fatalf("the payload decode refuses under %v, not only the form fault", decode)
	}
	for typ, k := range kindReplays {
		want := guardSet(t, code.Reachable(t, funcName(k.apply)))
		for g := range decode {
			want[g] = true
		}
		if k.goal != nil {
			for g := range snapshot {
				want[g] = true
			}
		}
		got := map[fault.Guard]bool{}
		for _, g := range k.guards {
			if got[g] {
				t.Errorf("kind %s names guard %q twice", typ, g)
			}
			got[g] = true
		}
		for g := range want {
			if !got[g] {
				t.Errorf("replay can refuse a %s under guard %q, which its replay entry does not name", typ, g)
			}
		}
		for g := range got {
			if !want[g] {
				t.Errorf("the replay entry of %s names guard %q, under which replay never refuses it", typ, g)
			}
		}
	}

	all := guardSet(t, code.Reachable(t, funcName(Replay)))
	if len(all) < 10 {
		t.Fatalf("replay refuses under only %d guards; the walk missed its code", len(all))
	}
	for _, g := range spec.Guards {
		if got := ChecksAtReplay(g.Name); got != all[g.Name] {
			t.Errorf("guard %q: ChecksAtReplay says %v, and replay's code can refuse under it: %v", g.Name, got, all[g.Name])
		}
		delete(all, g.Name)
	}
	for g := range all {
		t.Errorf("replay can refuse under guard %q, which has no entry in the spec", g)
	}
}

// TestGoalOfAddOfNewGoal:an add of a new Goal belongs to no Goal yet; a
// replace belongs to the Goal it replaces.
func TestGoalOfAddOfNewGoal(t *testing.T) {
	if got := goalOf(events.Event{ID: "A", Node: "A", Type: events.TypeGoalAdd}); got != "" {
		t.Errorf("goalOf(new goal.add) = %q, want empty", got)
	}
	if got := goalOf(events.Event{ID: "B", Node: "A", Type: events.TypeGoalAdd}); got != "A" {
		t.Errorf("goalOf(replace) = %q, want A", got)
	}
	if got := goalOf(events.Event{ID: "C", Node: "A", Type: events.TypeResultRecord}); got != "" {
		t.Errorf("goalOf(result.record) = %q, want empty", got)
	}
}
