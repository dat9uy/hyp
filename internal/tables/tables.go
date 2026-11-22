package tables

import (
	"slices"
	"sort"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
)

// Goal state names.
const (
	StateWritten  = "written"
	StateOpen     = "open"
	StateSelected = "selected"
	StateLanded   = "landed"
)

// The stops of a Goal (section 11). A land event names one of the three.
const (
	StopSelected      = "selected"
	StopTwoCandidates = "two candidates"
	StopBudgetUsed    = "budget used"
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
	Results    []*Result              // in merge order, refused ones included
	Select     *Select                // the Goal's current select, or nil
	Land       *Land                  // the Goal's land, or nil
	Unselects  []Unselect             // the owner's unselects, in merge order

	// epoch counts the example-changing places so far, and the unselects: the
	// budget counts the Results recorded in the current epoch.
	epoch int
}

// Select is a row of the Selects table.
type Select struct {
	ID      string // the select event id
	Commit  string
	RowTree string
	Time    string
	Actor   string
}

// Unselect is a row of the Unselects table: the owner's unselect of a
// selected Goal, which puts it back to open. Select is the select it undid.
// Results and Versions hold what had run before it: the count of the Goal's
// Results then, and every example version placed then (its example key and
// hash), each of which has H0's cell from its placement. The New cell guard
// reads them.
type Unselect struct {
	ID       string // the unselect event id
	Select   Select // the select it undid
	Time     string
	Actor    string
	Results  int
	Versions map[string]bool // by example key and hash: line-1/a@<hash>
}

// Land is a row of the Lands table: the owner's land of a Goal, which ends it.
// Stop is the stop it ended and Cause the event of the act that caused it.
type Land struct {
	ID    string // the land event id
	Stop  string
	Cause string
	Time  string
	Actor string
}

// Result is a row of the Results table.
type Result struct {
	ID        string // the result.record event id
	Line      string
	Commit    string
	RowTree   string
	Actor     string
	Role      string
	Time      string
	Retracted bool
	Refused   string // empty, or "Stable"
	Differing []string
	Cells     []events.ExampleCell
	Unstable  []UnstableCell // set only on a Result refused by Stable
	epoch     int            // the Goal's epoch when it was recorded
	wasH0     bool           // its row tree was H0's when it was recorded
}

// Refused is an event that failed a guard at its place in the merge order.
type Refused struct {
	EventID string
	Node    string // the event's node: a goal.add event id, or a Result id for a retract
	Err     error  // the guard's refusal
}

// Read is a row of the Reads table: what hyp printed to one actor for one
// command, from its read event.
type Read struct {
	ID      string // the read event id
	Time    string
	Actor   string
	Role    string
	Command []string
	Exit    int
	Stdout  string // the output hash of what it printed to stdout
	Stderr  string // the output hash of what it printed to stderr
}

// Tables are the calculated tables, kept in memory for one command.
type Tables struct {
	Goals           map[string]*Goal // by EventID
	results         map[string]*Result
	Reads           map[string][]Read // by actor, in merge order
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
		return nil, fault.BadArg(fault.NoGoal, name)
	case 1:
		return found, nil
	}
	return nil, fault.BadArg(fault.GoalTwice, name, n)
}

// GuardDraft refuses a replace of a Goal that is not written. A new Goal, with
// no node, always passes. node is the goal.add event id of the Goal to replace.
func GuardDraft(t Tables, node string) error {
	if node == "" {
		return nil
	}
	g, ok := t.Goals[node]
	if !ok {
		return fault.Refuse(fault.GuardDraft, fault.NoGoalToReplace, GoalID(node))
	}
	if g.State != StateWritten {
		return fault.Refuse(fault.GuardDraft, fault.NotWritten, g.ID, g.State)
	}
	return nil
}

// MaxLines is the most lines a Goal may list.
const MaxLines = 10

// GuardWholeLines refuses lines that are not whole: every line has a kind,
// gap or guard, the Goal has at least one gap line, and at most ten lines.
func GuardWholeLines(lines []Line) error {
	if len(lines) > MaxLines {
		return fault.Refuse(fault.GuardWholeLines, fault.TooManyLines, len(lines), MaxLines)
	}
	gaps := 0
	for i, l := range lines {
		switch l.Kind {
		case events.KindGap:
			gaps++
		case events.KindGuard:
		default:
			return fault.Refuse(fault.GuardWholeLines, fault.BadLineKind, i+1, l.Key, l.Kind)
		}
	}
	if gaps == 0 {
		return fault.Refuse(fault.GuardWholeLines, fault.NoGapLine)
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
		return fault.Refuse(fault.GuardWholeLines, fault.PayloadInvalid, err)
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

// LiveCell is a live cell on one row of a Goal: a cell of a Result that is not
// retracted and not refused by Stable, or the H0 cell of a placement, under
// the example hash that is current for its key. Result is empty for a
// placement's H0 cell.
type LiveCell struct {
	Key     string // line-3/a
	Hash    string // the example hash
	Probe   string
	Output  string
	Stderr  string
	Exit    int
	Verdict string
	Result  string // the result.record event id
}

// LiveCells is the one query for the live cells of a Goal on the row with
// rowTree, in the order they were made: a placement's H0 cell first, then the
// cells of the Results in merge order. The open state, Stable and Rows all
// read it. A placement's H0 cell lives on H0's row, and only while its row
// tree is the Goal's H0 row tree. A cell whose example hash is not the current
// placement's, or whose key's line the Goal no longer lists, is history.
func (g *Goal) LiveCells(rowTree string) []LiveCell {
	var out []LiveCell
	if rowTree == g.BaseTree {
		for _, p := range g.Placed() {
			if p.H0.RowTree == g.BaseTree {
				out = append(out, LiveCell{Key: p.Key, Hash: p.Hash, Probe: p.H0.Probe, Output: p.H0.Output,
					Stderr: p.H0.Stderr, Exit: p.H0.Exit, Verdict: p.H0.Verdict})
			}
		}
	}
	for _, r := range g.Results {
		if r.Retracted || r.Refused != "" || r.RowTree != rowTree {
			continue
		}
		for _, c := range r.Cells {
			key := r.Line + "/" + c.Example
			if cur, ok := g.Current(key); !ok || cur.Hash != c.Hash {
				continue
			}
			if _, ok := g.line(r.Line); !ok {
				continue
			}
			out = append(out, LiveCell{Key: key, Hash: c.Hash, Probe: c.Probe, Output: c.Output,
				Stderr: c.Stderr, Exit: c.Exit, Verdict: c.Verdict, Result: r.ID})
		}
	}
	return out
}

// Verdict is the verdict of the current placement of key on the row with
// rowTree: the first live cell's, because Stable keeps the first verdict, or
// NotRun when no cell is live.
func (g *Goal) Verdict(key, rowTree string) string {
	for _, c := range g.LiveCells(rowTree) {
		if c.Key == key {
			return c.Verdict
		}
	}
	return NotRun
}

// H0Verdict is the verdict of key on H0's row. A placement's cell reads
// NotRun after a Draft replace moved H0, until a run on H0 measures it.
func (g *Goal) H0Verdict(key string) string { return g.Verdict(key, g.BaseTree) }

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

// refresh recalculates the Goal's state. A selected or landed Goal stays so.
func (g *Goal) refresh() {
	switch {
	case g.State == StateSelected, g.State == StateLanded:
	case g.isOpen():
		g.State = StateOpen
	default:
		g.State = StateWritten
	}
}

// GuardNotLanded refuses every write to a landed Goal: the land ends it.
func GuardNotLanded(g *Goal) error {
	if g.State == StateLanded {
		return fault.Refuse(fault.GuardNotLanded, fault.Landed, g.ID)
	}
	return nil
}

// GuardNotSelected refuses a place, a run or a second select on a Goal that
// has a select.
func GuardNotSelected(g *Goal) error {
	if g.State == StateSelected {
		return fault.Refuse(fault.GuardSelected, fault.HasSelect, g.ID)
	}
	return nil
}

// GuardGrillPlaces refuses a placement by any role but the grill.
func GuardGrillPlaces(role string) error {
	if role != "grill" {
		return fault.Refuse(fault.GuardGrillPlaces, fault.NotGrill, role)
	}
	return nil
}

// GuardAddsOnlyTable refuses a placement of a hash the key has already held,
// current or earlier: a version only grows, so an old version never comes
// back. The content half of Adds only reads git and is the act's.
func GuardAddsOnlyTable(g *Goal, key, hash string) error {
	for _, old := range g.Placements[key] {
		if old.Hash != hash {
			continue
		}
		if cur, _ := g.Current(key); cur.Hash == hash {
			return fault.Refuse(fault.GuardAddsOnly, fault.AlreadyCurrent, hash, key, cur.Version)
		}
		return fault.Refuse(fault.GuardAddsOnly, fault.EarlierVersion, hash, old.Version, key)
	}
	return nil
}

// GuardPlace is the git-free check of an old-form example.place event, which
// a log written before the whole-set place holds; Replay calls it at the
// event's place in the merge order: not Selected, Grill places and the table
// half of Adds only. Not landed always holds. ev.Node is the goal.add event id
// of the Goal.
func GuardPlace(t Tables, ev events.Event) error {
	var p events.ExamplePlace
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardGrillPlaces, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToPlace, GoalID(ev.Node))
	}
	if err := GuardNotLanded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := GuardGrillPlaces(p.Role); err != nil {
		return err
	}
	return GuardAddsOnlyTable(g, p.Line+"/"+p.Example, p.Hash)
}

// GuardPlaceSet is the git-free check of an examples.place event, called by
// the place act and by Replay at the event's place in the merge order: not
// Selected, Grill places, and for each example of the set the table half of
// Adds only. The set names at least one example, each key once. Not landed
// always holds. ev.Node is the goal.add event id of the Goal.
func GuardPlaceSet(t Tables, ev events.Event) error {
	var p events.ExamplesPlace
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardGrillPlaces, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToPlace, GoalID(ev.Node))
	}
	if err := GuardNotLanded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := GuardGrillPlaces(p.Role); err != nil {
		return err
	}
	if len(p.Examples) == 0 {
		return fault.Refuse(fault.GuardAddsOnly, fault.SetUnchanged, g.ID)
	}
	seen := map[string]bool{}
	for _, x := range p.Examples {
		key := x.Line + "/" + x.Example
		if seen[key] {
			return fault.Refuse(fault.GuardAddsOnly, fault.KeyTwiceInSet, key)
		}
		seen[key] = true
		if err := GuardAddsOnlyTable(g, key, x.Hash); err != nil {
			return err
		}
	}
	return nil
}

// addPlacement adds one placement the guards allowed. The version is the
// count of the hashes the key has held, this one included.
func (g *Goal) addPlacement(ev events.Event, line, name, hash string, h0 events.Cell) {
	if g.Placements == nil {
		g.Placements = map[string][]Placement{}
	}
	key := line + "/" + name
	// GuardAddsOnlyTable refused every hash the key held, so this one is new.
	version := len(g.Placements[key]) + 1
	g.Placements[key] = append(g.Placements[key], Placement{
		Key: key, Line: line, Example: name, Hash: hash, Version: version,
		Time: ev.TS, Actor: ev.Actor, EventID: ev.ID, H0: h0,
	})
}

// place applies an old-form placement of one example. A new example key or a
// new hash on a key starts a new budget.
func (g *Goal) place(ev events.Event, p events.ExamplePlace) {
	g.addPlacement(ev, p.Line, p.Example, p.Hash, p.H0)
	g.epoch++
}

// placeSet applies a place of a set: every example it names, and one new
// budget for the whole set.
func (g *Goal) placeSet(ev events.Event, p events.ExamplesPlace) {
	for _, x := range p.Examples {
		g.addPlacement(ev, x.Line, x.Example, x.Hash, x.H0)
	}
	g.epoch++
}

// Replay applies the events in order. An event that fails a guard replay
// checks again is skipped and kept in RefusedAtReplay; it is never deleted
// from the log. A read event changes no Goal: it adds a row to Reads.
func Replay(evs []events.Event) Tables {
	t := Tables{Goals: map[string]*Goal{}, results: map[string]*Result{}, Reads: map[string][]Read{}}
	refuse := func(ev events.Event, err error) {
		t.RefusedAtReplay = append(t.RefusedAtReplay, Refused{EventID: ev.ID, Node: ev.Node, Err: err})
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
		case events.TypeExamplesPlace:
			if err := GuardPlaceSet(t, ev); err != nil {
				refuse(ev, err)
				continue
			}
			var p events.ExamplesPlace
			if err := events.Decode(ev, &p); err != nil {
				continue // GuardPlaceSet decoded it already
			}
			g := t.Goals[ev.Node]
			g.placeSet(ev, p)
			g.refresh()
		case events.TypeResultRecord:
			replayResult(&t, ev, refuse)
		case events.TypeResultRetract:
			replayRetract(&t, ev, refuse)
		case events.TypeSelect:
			replaySelect(&t, ev, refuse)
		case events.TypeLand:
			replayLand(&t, ev, refuse)
		case events.TypeUnselect:
			replayUnselect(&t, ev, refuse)
		case events.TypeRead:
			var p events.Read
			if err := events.Decode(ev, &p); err != nil {
				refuse(ev, fault.Refuse(fault.NoGuard, fault.PayloadInvalid, err))
				continue
			}
			t.Reads[ev.Actor] = append(t.Reads[ev.Actor], Read{
				ID: ev.ID, Time: ev.TS, Actor: ev.Actor, Role: p.Role, Command: p.Command,
				Exit: p.Exit, Stdout: p.Stdout, Stderr: p.Stderr,
			})
		}
	}
	return t
}

// replayResult adds a Result the run guards allow at its place in the merge
// order. A Result refused by Stable is kept, marked, and counts for nothing.
func replayResult(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	if err := GuardRun(*t, ev); err != nil {
		refuse(ev, err)
		return
	}
	var p events.ResultRecord
	if err := events.Decode(ev, &p); err != nil {
		return // GuardRun decoded it already
	}
	g := t.Goals[ev.Node]
	r := &Result{
		ID: ev.ID, Line: p.Line, Commit: p.Commit, RowTree: p.RowTree, Actor: ev.Actor, Role: p.Role,
		Time: ev.TS, Refused: p.Refused, Differing: p.Differing, Cells: p.Cells, epoch: g.epoch,
		wasH0: p.RowTree == g.BaseTree,
	}
	if p.Refused != "" {
		r.Unstable = g.unstableAt(p)
	}
	g.Results = append(g.Results, r)
	t.results[ev.ID] = r
	g.refresh()
}

// replayRetract marks a Result retracted. Only the actor that ran it may
// retract it (Own Result); a retract that names no Result is skipped.
func replayRetract(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	r, ok := t.results[ev.Node]
	if !ok {
		refuse(ev, fault.Refuse(fault.NoGuard, fault.NoResultToRetract, ev.Node))
		return
	}
	if r.Actor != ev.Actor {
		refuse(ev, fault.Refuse(fault.GuardOwnResult, fault.NotOwnResult, r.ID, r.Actor, ev.Actor))
		return
	}
	r.Retracted = true
	for _, g := range t.Goals {
		if slices.Contains(g.Results, r) {
			g.refresh()
		}
	}
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

// replaySelect applies a select the table guards allow at its place in the
// merge order: not Selected, the table half of Candidate, and Goal's actor for
// an owner. The git half of Candidate, the leaf mark, and Spec and Own path
// are trusted as written.
func replaySelect(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	if err := GuardSelect(*t, ev); err != nil {
		refuse(ev, err)
		return
	}
	var p events.Select
	if err := events.Decode(ev, &p); err != nil {
		return // GuardSelect decoded it already
	}
	g := t.Goals[ev.Node]
	g.Select = &Select{ID: ev.ID, Commit: p.Commit, RowTree: p.RowTree, Time: ev.TS, Actor: ev.Actor}
	g.State = StateSelected
}

// GuardOwnerLands refuses a land by any role but the owner.
func GuardOwnerLands(role string) error {
	if role != "owner" {
		return fault.Refuse(fault.GuardOwnerLands, fault.NotOwner, role)
	}
	return nil
}

// GuardLand is the git-free check of a land event, called by Replay at the
// event's place in the merge order: Not landed, Owner lands, Goal's actor, and
// the table half of Stop: the stop it names is one of the three, a select
// for Selected and none for the other two. That two candidates or the budget
// held when the act ran reads git, and is trusted as written, as the leaf mark
// of a select is. ev.Node is the goal.add event id.
func GuardLand(t Tables, ev events.Event) error {
	var p events.Land
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardStop, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToLand, GoalID(ev.Node))
	}
	if err := GuardNotLanded(g); err != nil {
		return err
	}
	if err := GuardOwnerLands(p.Role); err != nil {
		return err
	}
	if err := GuardGoalActor(g, ev.Actor, p.Role); err != nil {
		return err
	}
	switch {
	case p.Stop == StopSelected && g.Select != nil:
	case (p.Stop == StopTwoCandidates || p.Stop == StopBudgetUsed) && g.Select == nil:
	default:
		return fault.Refuse(fault.GuardStop, fault.StopNotHeld, p.Stop, g.ID)
	}
	return nil
}

// replayLand applies a land the table guards allow: the Goal is landed, and
// keeps the stop the land names.
func replayLand(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	if err := GuardLand(*t, ev); err != nil {
		refuse(ev, err)
		return
	}
	var p events.Land
	if err := events.Decode(ev, &p); err != nil {
		return // GuardLand decoded it already
	}
	g := t.Goals[ev.Node]
	g.Land = &Land{ID: ev.ID, Stop: p.Stop, Cause: p.Cause, Time: ev.TS, Actor: ev.Actor}
	g.State = StateLanded
}

// GuardOwnerUnselects refuses an unselect by any role but the owner.
func GuardOwnerUnselects(role string) error {
	if role != "owner" {
		return fault.Refuse(fault.GuardOwnerUnsel, fault.NotOwnerUnselect, role)
	}
	return nil
}

// GuardHasSelect is the Selected guard of unselect: the Goal has a select.
// An open Goal, never selected or already unselected, has none.
func GuardHasSelect(g *Goal) error {
	if g.State != StateSelected {
		return fault.Refuse(fault.GuardSelected, fault.NoSelect, g.ID, g.State)
	}
	return nil
}

// GuardUnselect is the check of an unselect event, called by the unselect act
// and by Replay at the event's place in the merge order: Not landed, Owner
// unselects, Goal's actor, then Selected (the Goal has a select). Every guard
// reads only the tables. ev.Node is the goal.add event id.
func GuardUnselect(t Tables, ev events.Event) error {
	var p events.Unselect
	if err := events.Decode(ev, &p); err != nil {
		return fault.Refuse(fault.GuardOwnerUnsel, fault.PayloadInvalid, err)
	}
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToUnselect, GoalID(ev.Node))
	}
	if err := GuardNotLanded(g); err != nil {
		return err
	}
	if err := GuardOwnerUnselects(p.Role); err != nil {
		return err
	}
	if err := GuardGoalActor(g, ev.Actor, p.Role); err != nil {
		return err
	}
	return GuardHasSelect(g)
}

// replayUnselect applies an unselect the guards allow: the Goal loses its
// select and is open again, its Results and cells stay as they were, and the
// budget's boundary moves to the unselect.
func replayUnselect(t *Tables, ev events.Event, refuse func(events.Event, error)) {
	if err := GuardUnselect(*t, ev); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	versions := map[string]bool{}
	for _, hist := range g.Placements {
		for _, p := range hist {
			versions[p.Key+"@"+p.Hash] = true
		}
	}
	g.Unselects = append(g.Unselects, Unselect{
		ID: ev.ID, Select: *g.Select, Time: ev.TS, Actor: ev.Actor, Results: len(g.Results), Versions: versions,
	})
	g.Select = nil
	g.State = StateOpen
	g.epoch++ // the unselect starts a new budget
	g.refresh()
}

// HasNewCell reports whether a run since the Goal's latest unselect recorded
// a live cell whose example version, or whose row, had not run before that
// unselect. A Goal never unselected needs none, so it reports true. A version
// is an example key at one hash: two keys whose folders hold the same bytes
// are two versions. A version placed before the unselect ran on H0 when it
// was placed; a row ran before
// when a Result before the unselect, not refused by Stable, named its tree.
func (g *Goal) HasNewCell() bool {
	if len(g.Unselects) == 0 {
		return true
	}
	u := g.Unselects[len(g.Unselects)-1]
	oldRows := map[string]bool{g.BaseTree: true}
	for _, r := range g.Results[:u.Results] {
		if r.Refused == "" {
			oldRows[r.RowTree] = true
		}
	}
	for _, r := range g.Results[u.Results:] {
		if r.Retracted || r.Refused != "" {
			continue
		}
		for _, c := range r.Cells {
			if !oldRows[r.RowTree] || !u.Versions[r.Line+"/"+c.Example+"@"+c.Hash] {
				return true
			}
		}
	}
	return false
}

// GuardNewCell is the New cell guard of select: after an unselect, the runner
// selects only once a run recorded a new cell (HasNewCell). The owner selects
// at any time.
func GuardNewCell(g *Goal, role string) error {
	if role == "owner" || g.HasNewCell() {
		return nil
	}
	u := g.Unselects[len(g.Unselects)-1]
	return fault.Refuse(fault.GuardNewCell, fault.NoNewCell, g.ID, u.ID)
}
