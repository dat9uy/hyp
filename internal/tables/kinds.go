package tables

import (
	"slices"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// kindReplay is the replay entry of one event kind of events.Kinds: the type
// of its payload, how replay applies an event of the kind, the guards replay
// checks of it, which Goal an end freezes it on, and, for a kind that ends a
// Goal, the snapshot it names. Replay reaches a kind only through its entry,
// so a kind's replay is changed in one place.
//
// Replay decodes each payload once, in the type its entry names, before any
// guard sees the event: a payload that does not decode is refused with the
// decode error and never reaches apply. A guard never sees bytes, so no kind
// and no guard decides on its own what an undecodable payload is.
type kindReplay struct {
	// payload returns a pointer to a new zero value of the kind's payload
	// type, which replay decodes the event's payload into.
	payload func() any

	// apply checks the event against the guards replay checks for its kind
	// and applies it to t, or hands it to refuse. p is the decoded payload,
	// of the type payload returns.
	apply func(t *Tables, ev events.Event, p any, refuse func(events.Event, error))

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
	// snapshot the decoded payload p names, or nil for an end that names
	// none.
	end func(p any) []string
}

// kindReplays are the replay entries, keyed by event type. Every type of the
// closed set has one, and no other type does.
var kindReplays = map[string]kindReplay{
	events.TypeGoalAdd: {payload: func() any { return new(events.GoalAdd) },
		apply: replayGoalAdd, goal: replacedGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardWholeLines, fault.GuardDraft,
			fault.GuardWholeSet, fault.GuardAddsOnly, fault.GuardKind}},
	events.TypeExamplePlace: {payload: func() any { return new(events.ExamplePlace) },
		apply: replayExamplePlace, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerPlaces, fault.GuardSelected,
			fault.GuardAddsOnly}},
	events.TypeExamplesPlace: {payload: func() any { return new(events.ExamplesPlace) },
		apply: replayExamplesPlace, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerPlaces, fault.GuardSelected,
			fault.GuardGoalsActor, fault.GuardAddsOnly}},
	events.TypeResultRecord: {payload: func() any { return new(events.ResultRecord) },
		apply:  replayResult,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardSelected, fault.GuardOpen}},
	events.TypeSelect: {payload: func() any { return new(events.Select) },
		apply: replaySelect, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardSelected, fault.GuardGoalsActor}},
	events.TypeLand: {payload: func() any { return new(events.Land) },
		apply: replayLand, end: landSnapshot,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardStop, fault.GuardOwnerLands,
			fault.GuardGoalsActor, fault.GuardEvidence}},
	events.TypeClose: {payload: func() any { return new(events.Close) },
		apply: replayClose, end: closeSnapshot,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerClose, fault.GuardGoalsActor}},
	events.TypeUnselect: {payload: func() any { return new(events.Unselect) },
		apply: replayUnselect, goal: nodeGoal,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotEnded, fault.GuardOwnerUnsel, fault.GuardGoalsActor,
			fault.GuardSelected}},
	events.TypeRead: {payload: func() any { return new(events.Read) },
		apply:  replayRead,
		guards: []fault.Guard{fault.NoGuard}},
	events.TypeFindingAdd: {payload: func() any { return new(events.FindingAdd) },
		apply:  replayFindingAdd,
		guards: []fault.Guard{fault.NoGuard}},
	events.TypeFindingFold: {payload: func() any { return new(events.FindingFold) },
		apply: replayFindingFold,
		guards: []fault.Guard{fault.NoGuard, fault.GuardInWork, fault.GuardIssue, fault.GuardNotRefuted,
			fault.GuardNotDeclined}},
	events.TypeFindingRefute: {payload: func() any { return new(events.FindingRefute) },
		apply:  replayFindingRefute,
		guards: []fault.Guard{fault.NoGuard, fault.GuardNotRefuted, fault.GuardNotDeclined}},
	events.TypeFindingDecline: {payload: func() any { return new(events.FindingDecline) },
		apply: replayFindingDecline,
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
func landSnapshot(p any) []string { return p.(*events.Land).Seen }

// closeSnapshot is the snapshot a close names.
func closeSnapshot(p any) []string { return p.(*events.Close).Seen }

// decoded is an event's payload as replay decoded it: p, of the type its
// kind's entry names, or the error decoding it returned. Both are nil for
// an event of a type outside the closed set.
type decoded struct {
	p   any
	err error
}

// decodePayloads decodes the payload of every event once, in the type its
// kind's replay entry names. It is the only place in this package that
// decodes a payload.
func decodePayloads(evs []events.Event) []decoded {
	out := make([]decoded, len(evs))
	for i, ev := range evs {
		k, ok := kindReplays[ev.Type]
		if !ok {
			continue
		}
		p := k.payload()
		if err := events.Decode(ev, p); err != nil {
			out[i].err = err
			continue
		}
		out[i].p = p
	}
	return out
}
