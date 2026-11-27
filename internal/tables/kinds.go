package tables

import (
	"slices"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// kindReplay is the replay entry of one event kind of events.Kinds: how
// replay applies an event of the kind, the guards replay checks of it, which
// Goal an end freezes it on, and, for a kind that ends a Goal, the snapshot
// it names. Replay reaches a kind only through its entry, so a kind's replay
// is changed in one place.
type kindReplay struct {
	// apply checks the event against the guards replay checks for its kind
	// and applies it to t, or hands it to refuse.
	apply func(t *Tables, ev events.Event, refuse func(events.Event, error))

	// guards are every guard replay can refuse an event of the kind under:
	// each guard apply can refuse it under, NoGuard for a form fault, and
	// Not ended for a kind an end freezes, whose snapshot replay checks.
	// hyp spec prints a guard checked at replay exactly when an entry
	// names it.
	guards []fault.Guard

	// goal is the goal.add event id of the Goal an end freezes the event
	// on, or empty for an event that adds a new Goal. It is nil exactly
	// for a kind whose events.Kinds entry says no end freezes it.
	goal func(ev events.Event) string

	// end is set for a kind that ends a Goal: it returns the ids of the
	// snapshot the event names, nil for an end that names none, and false
	// for a payload that does not decode.
	end func(ev events.Event) ([]string, bool)
}

// kindReplays are the replay entries, keyed by event type. Every type of the
// closed set has one, and no other type does.
var kindReplays = map[string]kindReplay{
	events.TypeGoalAdd: {apply: replayGoalAdd, goal: replacedGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardWholeLines, fault.GuardDraft,
			fault.GuardWholeSet, fault.GuardAddsOnly, fault.GuardKind}},
	events.TypeExamplePlace: {apply: replayExamplePlace, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerPlaces, fault.GuardSelected,
			fault.GuardAddsOnly}},
	events.TypeExamplesPlace: {apply: replayExamplesPlace, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerPlaces, fault.GuardSelected,
			fault.GuardGoalsActor, fault.GuardAddsOnly}},
	events.TypeResultRecord: {apply: replayResult,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardSelected, fault.GuardOpen}},
	events.TypeSelect: {apply: replaySelect, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardSelected, fault.GuardGoalsActor}},
	events.TypeLand: {apply: replayLand, end: landSnapshot,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardStop, fault.GuardOwnerLands,
			fault.GuardGoalsActor, fault.GuardEvidence}},
	events.TypeClose: {apply: replayClose, end: closeSnapshot,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerClose, fault.GuardGoalsActor}},
	events.TypeUnselect: {apply: replayUnselect, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerUnsel, fault.GuardGoalsActor,
			fault.GuardSelected}},
	events.TypeRead: {apply: replayRead,
		guards: []fault.Guard{fault.NoGuard}},
	events.TypeFindingAdd: {apply: replayFindingAdd,
		guards: []fault.Guard{fault.NoGuard}},
	events.TypeFindingFold: {apply: replayFindingFold,
		guards: []fault.Guard{fault.NoGuard, fault.GuardInWork, fault.GuardIssue, fault.GuardNotRefuted,
			fault.GuardNotDeclined}},
	events.TypeFindingRefute: {apply: replayFindingRefute,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotRefuted, fault.GuardNotDeclined}},
	events.TypeFindingDecline: {apply: replayFindingDecline,
		guards: []fault.Guard{fault.NoGuard, fault.GuardOwnerDecl, fault.GuardNotRefuted, fault.GuardNotDeclined,
			fault.GuardNotFolded}},
}

// ChecksAtReplay reports whether replay can refuse an event under guard g:
// whether the replay entry of some kind names it. hyp spec prints it as a
// guard's checked at replay.
func ChecksAtReplay(g fault.Guard) bool {
	for _, k := range kindReplays {
		if slices.Contains(k.guards, g) {
			return true
		}
	}
	return false
}

// nodeGoal is the Goal an event's node names.
func nodeGoal(ev events.Event) string { return ev.Node }

// replacedGoal is the Goal a goal.add replaces; an add of a new Goal names
// itself and belongs to no Goal yet.
func replacedGoal(ev events.Event) string {
	if ev.Node == ev.ID {
		return ""
	}
	return ev.Node
}

// landSnapshot is the snapshot a land names.
func landSnapshot(ev events.Event) ([]string, bool) {
	var p events.Land
	if events.Decode(ev, &p) != nil {
		return nil, false
	}
	return p.Seen, true
}

// closeSnapshot is the snapshot a close names.
func closeSnapshot(ev events.Event) ([]string, bool) {
	var p events.Close
	if events.Decode(ev, &p) != nil {
		return nil, false
	}
	return p.Seen, true
}
