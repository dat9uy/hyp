package tables

import (
	"testing"

	"github.com/dat9uy/hyp/internal/events"
)

// TestEveryKindHasOneReplayEntry: every kind of events.Kinds has a replay
// entry with a handler, a kind that changes a Goal's state above all; no
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

// TestGoalOfAddOfNewGoal: an add of a new Goal belongs to no Goal yet; a
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
