package tables

import (
	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
)

// Goal state names.
const StateWritten = "written"

// Line is a row of the Lines table.
type Line = events.Line

// Goal is a row of the Goals table.
type Goal struct {
	ID       string // G- and the last six characters of EventID
	EventID  string // the goal.add event that made it
	Words    string
	Base     string
	BaseTree string
	Budget   int
	Actor    string // of the first goal.add; a Draft replace keeps it
	From     string
	State    string
	Lines    []Line
}

// Refused is an event that failed a guard at its place in the merge order.
type Refused struct {
	EventID string
	Reason  string
}

// Tables are the calculated tables, kept in memory for one command.
type Tables struct {
	Goals           map[string]*Goal // by EventID
	RefusedAtReplay []Refused
}

// GoalID is the Goal id of a goal.add event id.
func GoalID(eventID string) string {
	tail := eventID
	if len(tail) > 6 {
		tail = tail[len(tail)-6:]
	}
	return "G-" + tail
}

// ByID returns the Goal whose id is id, or nil.
func (t Tables) ByID(id string) *Goal {
	for _, g := range t.Goals {
		if g.ID == id {
			return g
		}
	}
	return nil
}

// Find resolves a Goal named by its id (G-NXHHVA) or by its goal.add event id.
// A short id that matches two Goals is ambiguous, and an unknown one is a bad
// argument.
func (t Tables) Find(name string) (*Goal, error) {
	if g, ok := t.Goals[name]; ok {
		return g, nil
	}
	var found *Goal
	n := 0
	for _, g := range t.Goals {
		if g.ID == name {
			found = g
			n++
		}
	}
	switch n {
	case 0:
		return nil, herr.BadArgs("no Goal %s", name)
	case 1:
		return found, nil
	}
	return nil, herr.BadArgs("%s matches %d Goals; name the event id", name, n)
}

// GuardDraft refuses a replace of a Goal that is not written. A new Goal, with
// no node, always passes. node is the goal.add event id of the Goal to replace.
func GuardDraft(t Tables, node string) error {
	if node == "" {
		return nil
	}
	g, ok := t.Goals[node]
	if !ok {
		return herr.Refused("Draft: no Goal %s to replace", GoalID(node))
	}
	if g.State != StateWritten {
		return herr.Refused("Draft: %s is %s; only a written Goal may be replaced", g.ID, g.State)
	}
	return nil
}

// MaxLines is the most lines a Goal may list.
const MaxLines = 10

// GuardWholeLines refuses lines that are not whole: every line has a kind,
// gap or guard, the Goal has at least one gap line, and at most ten lines.
func GuardWholeLines(lines []Line) error {
	if len(lines) > MaxLines {
		return herr.Refused("Whole lines: %d lines; a Goal lists at most %d", len(lines), MaxLines)
	}
	gaps := 0
	for i, l := range lines {
		switch l.Kind {
		case events.KindGap:
			gaps++
		case events.KindGuard:
		default:
			return herr.Refused("Whole lines: line %d (%s) has kind %q; want gap or guard", i+1, l.Key, l.Kind)
		}
	}
	if gaps == 0 {
		return herr.Refused("Whole lines: the Goal has no gap line")
	}
	return nil
}

// GuardGoalAdd is the git-free check of a goal.add event, called by the add
// goal act and again by Replay at the event's place in the merge order, so the
// two never drift. ev.Node is empty or equal to ev.ID for a new Goal, and the
// goal.add event id of the Goal to replace otherwise.
func GuardGoalAdd(t Tables, ev events.Event) error {
	var p events.GoalAdd
	if err := events.Decode(ev, &p); err != nil {
		return herr.Refused("Whole lines: %v", err)
	}
	node := ev.Node
	if node == ev.ID {
		node = ""
	}
	if err := GuardDraft(t, node); err != nil {
		return err
	}
	return GuardWholeLines(p.Lines)
}

// Replay applies the events in order. An event that fails a guard replay
// checks again is skipped and kept in RefusedAtReplay; it is never deleted
// from the log.
func Replay(evs []events.Event) Tables {
	t := Tables{Goals: map[string]*Goal{}}
	for _, ev := range evs {
		if ev.Type != events.TypeGoalAdd {
			continue
		}
		if err := GuardGoalAdd(t, ev); err != nil {
			t.RefusedAtReplay = append(t.RefusedAtReplay, Refused{EventID: ev.ID, Reason: herr.Text(err)})
			continue
		}
		var p events.GoalAdd
		if err := events.Decode(ev, &p); err != nil {
			continue // GuardGoalAdd decoded it already
		}
		if g, ok := t.Goals[ev.Node]; ok && ev.Node != ev.ID {
			g.Words, g.Base, g.BaseTree, g.From, g.Lines = p.Words, p.Base, p.BaseTree, p.From, p.Lines
			continue
		}
		t.Goals[ev.ID] = &Goal{
			ID: GoalID(ev.ID), EventID: ev.ID, Words: p.Words, Base: p.Base, BaseTree: p.BaseTree,
			Budget: p.Budget, Actor: ev.Actor, From: p.From, State: StateWritten, Lines: p.Lines,
		}
	}
	return t
}
