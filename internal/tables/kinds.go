package tables

import "github.com/dat9uy/hyp/internal/events"

// kindReplay is the replay entry of one event kind of events.Kinds: how
// replay applies an event of the kind, which Goal an end freezes it on, and,
// for a kind that ends a Goal, the snapshot it names. Replay reaches a kind
// only through its entry, so a kind's replay is changed in one place.
type kindReplay struct {
	// apply checks the event against the guards replay checks for its kind
	// and applies it to t, or hands it to refuse.
	apply func(t *Tables, ev events.Event, refuse func(events.Event, error))

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
	events.TypeGoalAdd:        {apply: replayGoalAdd, goal: replacedGoal},
	events.TypeExamplePlace:   {apply: replayExamplePlace, goal: nodeGoal},
	events.TypeExamplesPlace:  {apply: replayExamplesPlace, goal: nodeGoal},
	events.TypeResultRecord:   {apply: replayResult},
	events.TypeSelect:         {apply: replaySelect, goal: nodeGoal},
	events.TypeLand:           {apply: replayLand, end: landSnapshot},
	events.TypeClose:          {apply: replayClose, end: closeSnapshot},
	events.TypeUnselect:       {apply: replayUnselect, goal: nodeGoal},
	events.TypeRead:           {apply: replayRead},
	events.TypeFindingAdd:     {apply: replayFindingAdd},
	events.TypeFindingFold:    {apply: replayFindingFold},
	events.TypeFindingRefute:  {apply: replayFindingRefute},
	events.TypeFindingDecline: {apply: replayFindingDecline},
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
