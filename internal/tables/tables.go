package tables

import (
	"sort"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
)

// Goal state names.
const (
	StateWritten  = "written"
	StateOpen     = "open"
	StateSelected = "selected"
)

// NotRun is the verdict a row shows for a cell that is not live.
const NotRun = "not run"

// Placement is a row of the Placements table: one example folder placed on
// an example key. The latest placement of a key is its current one.
type Placement struct {
	Key     string // line key and example name: line-3/a
	Line    string
	Example string
	Hash    string // the example folder's tree hash
	Version int    // @n: the place of Hash among the distinct hashes of Key
	Time    string
	Actor   string
	EventID string
	H0      events.Cell // H0's cell, written by the placement
}

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

	Placements map[string][]Placement // by example key, in placement order
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

// Current returns the current placement of an example key.
func (g *Goal) Current(key string) (Placement, bool) {
	hist := g.Placements[key]
	if len(hist) == 0 {
		return Placement{}, false
	}
	return hist[len(hist)-1], true
}

// line returns the Goal's line with key.
func (g *Goal) line(key string) (Line, bool) {
	for _, l := range g.Lines {
		if l.Key == key {
			return l, true
		}
	}
	return Line{}, false
}

// Placed lists the current placements of the lines the Goal still lists, in
// key order. A placement whose line key is gone is history and is left out.
func (g *Goal) Placed() []Placement {
	var out []Placement
	for key := range g.Placements {
		p, _ := g.Current(key)
		if _, ok := g.line(p.Line); ok {
			out = append(out, p)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Key < out[j].Key })
	return out
}

// H0Verdict is the verdict of the current placement of key on H0's row: its
// cell's verdict while the cell's row tree is the Goal's H0 row tree, and
// NotRun after a Draft replace moved H0, or when key was never placed.
func (g *Goal) H0Verdict(key string) string {
	p, ok := g.Current(key)
	if !ok || p.H0.RowTree != g.BaseTree {
		return NotRun
	}
	return p.H0.Verdict
}

// Needs is the verdict H0 must give an example of a line of this kind: a gap
// example fails, a guard example is met.
func Needs(kind string) string {
	if kind == events.KindGuard {
		return "met"
	}
	return "failed"
}

// isOpen reports whether every line has a placed example and every current
// placement has a live cell on H0's row with the verdict its line's kind
// needs.
func (g *Goal) isOpen() bool {
	placed := map[string]bool{}
	for _, p := range g.Placed() {
		l, _ := g.line(p.Line)
		if g.H0Verdict(p.Key) != Needs(l.Kind) {
			return false
		}
		placed[p.Line] = true
	}
	for _, l := range g.Lines {
		if !placed[l.Key] {
			return false
		}
	}
	return len(g.Lines) > 0
}

// refresh recalculates the Goal's state. A selected Goal stays selected.
func (g *Goal) refresh() {
	switch {
	case g.State == StateSelected:
	case g.isOpen():
		g.State = StateOpen
	default:
		g.State = StateWritten
	}
}

// GuardNotSelected refuses a place, a run or a second select on a Goal that
// has a select.
func GuardNotSelected(g *Goal) error {
	if g.State == StateSelected {
		return herr.Refused("Selected: %s has a select; it takes no place, run or second select", g.ID)
	}
	return nil
}

// GuardOwnCase refuses a placement by any role but the grill.
func GuardOwnCase(role string) error {
	if role != "grill" {
		return herr.Refused("Own case: only the grill places an example; the role is %q", role)
	}
	return nil
}

// GuardAddsOnlyTable refuses a placement of the hash that is already current
// on its key. The content half of Adds only reads git and is the act's.
func GuardAddsOnlyTable(g *Goal, key, hash string) error {
	if cur, ok := g.Current(key); ok && cur.Hash == hash {
		return herr.Refused("Adds only: %s is already the current version of %s (@%d); a new version adds a floor or a ceiling", hash, key, cur.Version)
	}
	return nil
}

// GuardPlace is the git-free check of an example.place event, called by Replay
// at the event's place in the merge order: not Selected, Own case and the
// table half of Adds only. The place act calls the same three guards in spec
// order around the guards that read git. Not ended always holds. ev.Node is
// the goal.add event id of the Goal.
func GuardPlace(t Tables, ev events.Event) error {
	var p events.ExamplePlace
	if err := events.Decode(ev, &p); err != nil {
		return herr.Refused("Own case: %v", err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return herr.Refused("no Goal %s to place on", GoalID(ev.Node))
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := GuardOwnCase(p.Role); err != nil {
		return err
	}
	return GuardAddsOnlyTable(g, p.Line+"/"+p.Example, p.Hash)
}

// place applies a placement the guards allowed. The version is the place of
// the hash among the distinct hashes the key has held.
func (g *Goal) place(ev events.Event, p events.ExamplePlace) {
	if g.Placements == nil {
		g.Placements = map[string][]Placement{}
	}
	key := p.Line + "/" + p.Example
	version := 0
	seen := map[string]bool{}
	for _, old := range g.Placements[key] {
		if old.Hash == p.Hash {
			version = old.Version
		}
		seen[old.Hash] = true
	}
	if version == 0 {
		version = len(seen) + 1
	}
	g.Placements[key] = append(g.Placements[key], Placement{
		Key: key, Line: p.Line, Example: p.Example, Hash: p.Hash, Version: version,
		Time: ev.TS, Actor: ev.Actor, EventID: ev.ID, H0: p.H0,
	})
}

// Replay applies the events in order. An event that fails a guard replay
// checks again is skipped and kept in RefusedAtReplay; it is never deleted
// from the log.
func Replay(evs []events.Event) Tables {
	t := Tables{Goals: map[string]*Goal{}}
	refuse := func(ev events.Event, err error) {
		t.RefusedAtReplay = append(t.RefusedAtReplay, Refused{EventID: ev.ID, Reason: herr.Text(err)})
	}
	for _, ev := range evs {
		switch ev.Type {
		case events.TypeGoalAdd:
			replayGoalAdd(&t, ev, refuse)
		case events.TypeExamplePlace:
			if err := GuardPlace(t, ev); err != nil {
				refuse(ev, err)
				continue
			}
			var p events.ExamplePlace
			if err := events.Decode(ev, &p); err != nil {
				continue // GuardPlace decoded it already
			}
			g := t.Goals[ev.Node]
			g.place(ev, p)
			g.refresh()
		}
	}
	return t
}

func replayGoalAdd(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	if err := GuardGoalAdd(*t, ev); err != nil {
		refuse(ev, err)
		return
	}
	var p events.GoalAdd
	if err := events.Decode(ev, &p); err != nil {
		return // GuardGoalAdd decoded it already
	}
	if g, ok := t.Goals[ev.Node]; ok && ev.Node != ev.ID {
		g.Words, g.Base, g.BaseTree, g.From, g.Lines = p.Words, p.Base, p.BaseTree, p.From, p.Lines
		g.refresh()
		return
	}
	t.Goals[ev.ID] = &Goal{
		ID: GoalID(ev.ID), EventID: ev.ID, Words: p.Words, Base: p.Base, BaseTree: p.BaseTree,
		Budget: p.Budget, Actor: ev.Actor, From: p.From, State: StateWritten, Lines: p.Lines,
	}
}
