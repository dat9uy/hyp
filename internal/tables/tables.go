package tables

import (
	"slices"
	"sort"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/spec"
)

// Goal state names.
const (
	StateWritten  = spec.StateWritten
	StateOpen     = spec.StateOpen
	StateSelected = spec.StateSelected
	StateEnded    = spec.StateEnded
)

// The stops of a Goal. A land event names one of the three.
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

	Findings []string // the finding.add ids its findings section names
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
	Close      *Close                 // the Goal's close, or nil
	Unselects  []Unselect             // the owner's unselects, in merge order
	Rejected   map[string]bool        // the row trees the owner's unselects reject: never selected again
	Why        string                 // the why of its latest goal.add
	Named      []string               // the finding.add ids its latest goal.add names

	// seen lists the ids of the Goal's events that replay applied, in merge
	// order: the snapshot an end names.
	seen []string

	// epoch counts the example-changing places so far, and the unselects: the
	// budget counts the Results recorded in the current epoch.
	epoch int

	// cells are the cells the clone's trace keeps as files, by
	// logstore.CellKey: H0's cells of the places, and the cells a try ran
	// again. Tables.UseCells sets them.
	cells map[string]logstore.Cell

	// Rests marks a Goal added with --from a Goal in work: it rests on that
	// Goal's select (Rest). Its goal.add marks it so.
	Rests bool
	// Rest is what a resting Goal in work reads of its premise; Calculate
	// sets it.
	Rest *Rest
	// premise are the cells, by example key, of a Goal that stands on its
	// premise's commit: its H0 cells there.
	premise map[string]LiveCell
}

// Select is a row of the Selects table.
type Select struct {
	ID      string // the select event id, or for a calculated select the Result that made the row met on every line
	Commit  string
	RowTree string
	Time    string
	Actor   string

	// Calculated marks a select no event holds: this clone's Results show
	// the row as the one candidate (Calculate). A select event of the
	// older form is not calculated.
	Calculated bool
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
	Why   string // the why the owner gave the land

	// The evidence the land carries, when it lands a commit: the commit, its
	// tree, its row tree and each example's verdict there. A land of the
	// older form carries none.
	Commit   string
	Tree     string
	RowTree  string
	Verdicts []events.LandVerdict
	Read     *events.CommitRead // what the land read of the commit; nil on a land of the older form
}

// Close is a row of the Closes table: the owner's close of a Goal, which ends
// it at any stop or none.
type Close struct {
	ID    string // the close event id
	Time  string
	Actor string
	Why   string // the why the owner gave the close
}

// Findings are the finding.add ids the Goal's record names: those its latest
// goal.add names, then those the current version of each placed example
// names, in key order, each once. They are in git, so every clone reads the
// same ids, whether or not its trace holds the findings.
func (g *Goal) Findings() []string {
	out := slices.Clone(g.Named)
	for _, p := range g.Placed() {
		for _, id := range p.Findings {
			if !slices.Contains(out, id) {
				out = append(out, id)
			}
		}
	}
	return out
}

// Seen is the snapshot an end of the Goal names: the ids of the Goal's
// events that replay applied, in merge order.
func (g *Goal) Seen() []string { return slices.Clone(g.seen) }

// EndID is the id of the land or close that ended the Goal, or empty.
func (g *Goal) EndID() string {
	switch {
	case g.Land != nil:
		return g.Land.ID
	case g.Close != nil:
		return g.Close.ID
	}
	return ""
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
	Refused   string // empty, "Stable" or "Same base"
	Differing []string
	Cells     []events.ExampleCell
	Unstable  []UnstableCell     // set only on a Result refused by Stable
	Read      *events.CommitRead // what its run read of git; nil on a Result recorded before runs carried it
	Base      *events.BaseRead   // the base newer than H0 its run measured; nil when it stood on H0
	Move      string             // the selected commit its run, on a merge of it, moved from
	epoch     int                // the Goal's epoch when it was recorded
	wasH0     bool               // its row tree was H0's when it was recorded
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
	findings        map[string]*Finding // by finding.add event id
	order           []*Finding          // every finding, in merge order
	Reads           map[string][]Read   // by actor, in merge order
	RefusedAtReplay []Refused
}

// UseCells gives every Goal the cells the clone's trace keeps as files. H0's
// cell of a place reads its outputs from them; a cell no Result of this clone
// holds can be one a try ran again.
func (t Tables) UseCells(cells map[string]logstore.Cell) {
	for _, g := range t.Goals {
		g.cells = cells
		g.nameH0Outputs()
	}
}

// TraceCell is the cell the clone's trace keeps as a file for the example
// with hash on the row with rowTree, if any.
func (g *Goal) TraceCell(hash, rowTree string) (logstore.Cell, bool) {
	c, ok := g.cells[logstore.CellKey(hash, rowTree)]
	return c, ok
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
// p is the payload of ev.
func GuardGoalAdd(t Tables, ev events.Event, p *events.GoalAdd) error {
	node := ev.Node
	if node == ev.ID {
		node = ""
	}
	if err := GuardDraft(t, node); err != nil {
		return err
	}
	if err := GuardWholeLines(p.Lines); err != nil {
		return err
	}
	if err := guardAddedExamples(*p); err != nil {
		return err
	}
	if node == "" {
		return nil
	}
	return GuardReplaceExamples(GoalID(node), len(p.Examples) > 0)
}

// GuardReplaceExamples is the half of Draft that a replace carries no
// example, since the owner places them: goal names the Goal replaced, and
// examples says whether the replace carries any.
func GuardReplaceExamples(goal string, examples bool) error {
	if examples {
		return fault.Refuse(fault.GuardDraft, fault.ReplaceExamples, goal)
	}
	return nil
}

// guardAddedExamples is the table half of the examples a goal.add carries:
// a new Goal only, each on a line of the Goal, each key once, each with H0's
// verdict its line's kind needs as the add recorded it. The act ran each on
// H0 (Kind) and read its folder (Committed, Adds only, Meetable).
func guardAddedExamples(p events.GoalAdd) error {
	seen := map[string]bool{}
	for _, x := range p.Examples {
		key := x.Line + "/" + x.Example
		kind := ""
		for _, l := range p.Lines {
			if l.Key == x.Line {
				kind = l.Kind
			}
		}
		if kind == "" {
			return fault.Refuse(fault.GuardWholeSet, fault.ExampleOffLine, key)
		}
		if seen[key] {
			return fault.Refuse(fault.GuardAddsOnly, fault.KeyTwiceInSet, key)
		}
		seen[key] = true
		if x.H0.Verdict != Needs(kind) || x.H0.RowTree != p.BaseTree {
			return fault.Refuse(fault.GuardKind, fault.AddedVerdictWrong, key, x.H0.Verdict, Needs(kind))
		}
	}
	return nil
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
// refused by Stable, or the H0 cell of a placement, under
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
	By      string // for H0's cell of a place: who saved its outputs in the trace, or empty when this clone's trace lacks them
}

// LiveCells is the one query for the live cells of a Goal on the row with
// rowTree, in the order they were made: a placement's H0 cell first, then the
// cells of the Results in merge order. The open state, Stable and Rows all
// read it. A placement's H0 cell lives on H0's row, and only while its row
// tree is the Goal's H0 row tree. A cell whose example hash is not the current
// placement's, or whose key's line the Goal no longer lists, is history.
func (g *Goal) LiveCells(rowTree string) []LiveCell {
	var out []LiveCell
	if rowTree == g.BaseTree && g.premise != nil {
		for _, p := range g.Placed() {
			if c, ok := g.premise[p.Key]; ok && c.Hash == p.Hash {
				out = append(out, c)
			}
		}
	} else if rowTree == g.BaseTree {
		for _, p := range g.Placed() {
			if p.H0.RowTree == g.BaseTree {
				// the outputs are the trace's alone, whatever the place names
				c := LiveCell{Key: p.Key, Hash: p.Hash, Probe: p.H0.Probe, Exit: p.H0.Exit, Verdict: p.H0.Verdict}
				if tc, ok := g.TraceCell(p.Hash, g.BaseTree); ok {
					c.Output, c.Stderr, c.By = tc.Output, tc.Stderr, tc.By
				}
				out = append(out, c)
			}
		}
	}
	for _, r := range g.Results {
		if r.Refused != "" || r.RowTree != rowTree {
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

// PlacedVerdict is H0's verdict on the current placement of key as git holds
// it: the verdict its place recorded, while that place's row tree is the
// Goal's H0 row tree, else NotRun.
func (g *Goal) PlacedVerdict(key string) string {
	if p, ok := g.Current(key); ok && p.H0.RowTree == g.BaseTree {
		return p.H0.Verdict
	}
	return NotRun
}

// isOpen reports whether every line has a placed example and every current
// placement's verdict on H0, as its place recorded it in git, is the one its
// line's kind needs. It reads no Result: a Result is in the trace of the
// clone that ran it, and every clone calculates the same state from git.
func (g *Goal) isOpen() bool {
	placed := map[string]bool{}
	for _, p := range g.Placed() {
		l, _ := g.line(p.Line)
		if g.PlacedVerdict(p.Key) != Needs(l.Kind) {
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

// refresh recalculates the Goal's state. A selected or ended Goal stays so.
func (g *Goal) refresh() {
	switch {
	case g.State == StateSelected, g.State == StateEnded:
	case g.isOpen():
		g.State = StateOpen
	default:
		g.State = StateWritten
	}
}

// GuardNotEnded refuses every write to an ended Goal: a land or a close ends
// it.
func GuardNotEnded(g *Goal) error {
	if g.State == StateEnded {
		return fault.Refuse(fault.GuardNotEnded, fault.Ended, g.ID)
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

// IsMove reports whether a run that names move, the selected commit its
// commit merges, moves g from its select: g shows selected on move by this
// clone's Results, with no select event.
func IsMove(g *Goal, move string) bool {
	return move != "" && g.State == StateSelected && g.Select != nil && g.Select.Calculated && g.Select.Commit == move
}

// GuardOwnerPlaces refuses a place by any role but the owner: a place is
// an owner act, and git holds only the owner's acts.
func GuardOwnerPlaces(role string) error {
	if role != "owner" {
		return fault.Refuse(fault.GuardOwnerPlaces, fault.NotOwnerPlace, role)
	}
	return nil
}

// guardPlacerAtReplay is Owner places at replay: a place of the older form,
// written before a place was the owner's act, names the grill, and replays
// as it was.
func guardPlacerAtReplay(role string) error {
	if role == "grill" {
		return nil
	}
	return GuardOwnerPlaces(role)
}

// GuardAddsOnlyTable refuses a placement of a hash the key has already held,
// current or earlier: a version only grows, so an old version never comes
// back. The current hash may be placed again only to measure it on H0 after
// a replace moved H0, when its place's row tree is no longer H0's. The
// content half of Adds only reads git and is the act's.
func GuardAddsOnlyTable(g *Goal, key, hash string) error {
	for _, old := range g.Placements[key] {
		if old.Hash != hash {
			continue
		}
		if cur, _ := g.Current(key); cur.Hash == hash {
			if cur.H0.RowTree != g.BaseTree {
				return nil
			}
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
// p is the payload of ev.
func GuardPlace(t Tables, ev events.Event, p *events.ExamplePlace) error {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToPlace, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := guardPlacerAtReplay(p.Role); err != nil {
		return err
	}
	return GuardAddsOnlyTable(g, p.Line+"/"+p.Example, p.Hash)
}

// GuardPlaceSet is the git-free check of an examples.place event, called by
// the place act and by Replay at the event's place in the merge order: not
// Selected, Grill places, and for each example of the set the table half of
// Adds only. The set names at least one example, each key once. Not landed
// always holds. ev.Node is the goal.add event id of the Goal.
// p is the payload of ev.
func GuardPlaceSet(t Tables, ev events.Event, p *events.ExamplesPlace) error {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToPlace, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardNotSelected(g); err != nil {
		return err
	}
	if err := guardPlacerAtReplay(p.Role); err != nil {
		return err
	}
	if err := GuardGoalActor(g, ev.Actor, p.Role); err != nil {
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
func (g *Goal) addPlacement(ev events.Event, line, name, hash string, h0 events.Cell, findings []string) {
	if g.Placements == nil {
		g.Placements = map[string][]Placement{}
	}
	key := line + "/" + name
	// GuardAddsOnlyTable refused every hash the key held but the current one
	// measured again on a moved H0, which keeps its version.
	version := 1
	if cur, ok := g.Current(key); ok {
		version = cur.Version + 1
		if cur.Hash == hash {
			version = cur.Version
		}
	}
	g.Placements[key] = append(g.Placements[key], Placement{
		Key: key, Line: line, Example: name, Hash: hash, Version: version,
		Time: ev.TS, Actor: ev.Actor, EventID: ev.ID, H0: h0, Findings: findings,
	})
}

// place applies an old-form placement of one example. A new example key or a
// new hash on a key starts a new budget.
func (g *Goal) place(ev events.Event, p events.ExamplePlace) {
	g.addPlacement(ev, p.Line, p.Example, p.Hash, p.H0, nil)
	g.epoch++
}

// placeSet applies a place of a set: every example it names, and one new
// budget for the whole set.
func (g *Goal) placeSet(ev events.Event, p events.ExamplesPlace) {
	for _, x := range p.Examples {
		g.addPlacement(ev, x.Line, x.Example, x.Hash, x.H0, x.Findings)
	}
	g.epoch++
}

// Replay applies the events in order. An event that fails a guard replay
// checks again is skipped and kept in RefusedAtReplay; it is never deleted
// from the log. A read event changes no Goal: it adds a row to Reads.
//
// An end, a land or a close, names its snapshot: the Goal's events replay
// applied when the owner ended it. An event of that Goal ahead of the end in
// the merge order that the snapshot does not name was written in another
// clone before the end and merged after it: Not ended refuses it, so the
// ended Goal holds what the owner saw in every clone. A land written before
// close existed names no snapshot and refuses nothing. The snapshot used is
// that of the end replay applies; an end that replay refuses names none, so
// Replay drops it and replays again. Replay decodes each payload once, before
// it replays, and every pass reads the payloads it decoded.
func Replay(evs []events.Event) Tables {
	ps := decodePayloads(evs)
	dropped := map[string]bool{}
	for {
		snaps := snapshots(evs, ps, dropped)
		t := replayWith(evs, ps, snaps)
		again := false
		for node, s := range snaps {
			if g := t.Goals[node]; g == nil || g.EndID() != evs[s.at].ID {
				dropped[evs[s.at].ID] = true
				again = true
			}
		}
		if !again {
			return t
		}
	}
}

// snapshot is the snapshot an end names: the end's place in the merge order
// and the ids it names.
type snapshot struct {
	at   int
	seen map[string]bool
}

// snapshots finds, for each Goal, its first end in the merge order that is
// not dropped and whose payload decoded, and keeps it when it names a
// snapshot. Every later end of the Goal fails Not ended. ps are the payloads
// of evs as decodePayloads decoded them.
func snapshots(evs []events.Event, ps []decoded, dropped map[string]bool) map[string]snapshot {
	out := map[string]snapshot{}
	first := map[string]bool{}
	for i, ev := range evs {
		if dropped[ev.ID] || first[ev.Node] {
			continue
		}
		k, ok := kindReplays[ev.Type]
		if !ok || k.end == nil || ps[i].err != nil {
			continue
		}
		seen := k.end(ps[i].p)
		first[ev.Node] = true
		if seen == nil {
			continue
		}
		s := snapshot{at: i, seen: map[string]bool{}}
		for _, id := range seen {
			s.seen[id] = true
		}
		out[ev.Node] = s
	}
	return out
}

// goalOf is the goal.add event id of the Goal an end freezes the event on,
// as its kind's replay entry names it, or empty for an event that changes no
// Goal, that adds a new Goal, or that no end freezes.
func goalOf(ev events.Event) string {
	k, ok := kindReplays[ev.Type]
	if !ok || k.goal == nil {
		return ""
	}
	return k.goal(ev)
}

// guardInSnapshot is the half of Not ended that replay checks of an event of
// the Goal node ahead of its end in the merge order: the snapshot s of the
// end endID names it.
func guardInSnapshot(ev events.Event, node string, s snapshot, endID string) error {
	if !s.seen[ev.ID] {
		return fault.Refuse(fault.GuardNotEnded, fault.NotInSnapshot, GoalID(node), endID)
	}
	return nil
}

// replayWith applies evs in order, with the snapshots snaps. ps are the
// payloads of evs as decodePayloads decoded them: an event whose payload did
// not decode is refused with the decode error, a form fault that names no
// guard, after the half of Not ended that its end's snapshot checks, and its
// kind's apply never sees it.
func replayWith(evs []events.Event, ps []decoded, snaps map[string]snapshot) Tables {
	t := Tables{Goals: map[string]*Goal{}, results: map[string]*Result{}, findings: map[string]*Finding{}, Reads: map[string][]Read{}}
	refuse := func(ev events.Event, err error) {
		t.RefusedAtReplay = append(t.RefusedAtReplay, Refused{EventID: ev.ID, Node: ev.Node, Err: err})
	}
	for i, ev := range evs {
		if node := goalOf(ev); node != "" {
			if s, ok := snaps[node]; ok && i < s.at {
				if err := guardInSnapshot(ev, node, s, evs[s.at].ID); err != nil {
					refuse(ev, err)
					continue
				}
			}
		}
		k, ok := kindReplays[ev.Type]
		if !ok {
			continue
		}
		if ps[i].err != nil {
			refuse(ev, ps[i].err)
			continue
		}
		k.apply(&t, ev, ps[i].p, refuse)
	}
	linkNames(&t)
	return t
}

// replayExamplePlace applies an old-form place of one example that
// GuardPlace allows.
func replayExamplePlace(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.ExamplePlace)
	if err := GuardPlace(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	g.place(ev, *p)
	g.seen = append(g.seen, ev.ID)
	g.refresh()
}

// replayExamplesPlace applies a place of a set that GuardPlaceSet allows.
func replayExamplesPlace(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.ExamplesPlace)
	if err := GuardPlaceSet(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	g.placeSet(ev, *p)
	g.seen = append(g.seen, ev.ID)
	g.refresh()
}

// replayRead adds a read to the reads of its actor. A read changes no Goal.
func replayRead(t *Tables, ev events.Event, payload any, _ func(events.Event, error)) {
	p := payload.(*events.Read)
	t.Reads[ev.Actor] = append(t.Reads[ev.Actor], Read{
		ID: ev.ID, Time: ev.TS, Actor: ev.Actor, Role: p.Role, Command: p.Command,
		Exit: p.Exit, Stdout: p.Stdout, Stderr: p.Stderr,
	})
}

// replayResult adds a Result at its place in the merge order, when the run
// guards that read only git allow it. A Result is not in the snapshot an end
// names: the snapshot names only events stored in git. A Result refused by
// Stable is kept, marked, and counts for nothing.
func replayResult(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.ResultRecord)
	if err := guardRunAtReplay(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	r := &Result{
		ID: ev.ID, Line: p.Line, Commit: p.Commit, RowTree: p.RowTree, Actor: ev.Actor, Role: p.Role,
		Time: ev.TS, Refused: p.Refused, Differing: p.Differing, Cells: p.Cells, Read: p.Read, epoch: g.epoch,
		wasH0: p.RowTree == g.BaseTree, Base: p.Base, Move: p.Move,
	}
	if p.Refused == RefusedStable {
		r.Unstable = g.unstableAt(*p)
	}
	g.Results = append(g.Results, r)
	t.results[ev.ID] = r
}

// restsOn reports whether the Goal the add p records rests on the Goal its
// From cites: the add marks it so, and the cited Goal is one of the store.
func (t *Tables) restsOn(p *events.GoalAdd) bool {
	_, ok := t.Goals[p.From]
	return p.Rests && p.From != "" && ok
}

func replayGoalAdd(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.GoalAdd)
	if err := GuardGoalAdd(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	if g, ok := t.Goals[ev.Node]; ok && ev.Node != ev.ID {
		// a replace carries no example: GuardGoalAdd refused one that does
		g.Words, g.Base, g.BaseTree, g.From, g.Lines, g.Why, g.Named = p.Words, p.Base, p.BaseTree, p.From, p.Lines, ev.Why, p.Findings
		g.Rests = t.restsOn(p)
		g.seen = append(g.seen, ev.ID)
		g.refresh()
		return
	}
	g := &Goal{
		ID: GoalID(ev.ID), EventID: ev.ID, Words: p.Words, Base: p.Base, BaseTree: p.BaseTree,
		Budget: p.Budget, Actor: ev.Actor, From: p.From, State: StateWritten, Lines: p.Lines,
		Why: ev.Why, Named: p.Findings, seen: []string{ev.ID}, Rests: t.restsOn(p),
	}
	t.Goals[ev.ID] = g
	if len(p.Examples) > 0 {
		// the add carries the examples: the Goal is open at once
		for _, x := range p.Examples {
			g.addPlacement(ev, x.Line, x.Example, x.Hash, x.H0, x.Findings)
		}
		g.epoch++
		g.refresh()
	}
}

// replaySelect applies a select the table guards allow at its place in the
// merge order: not Selected, the table half of Candidate, and Goal's actor for
// an owner. The git half of Candidate, the leaf mark, and Spec and Own path
// are trusted as written.
func replaySelect(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.Select)
	if err := GuardSelect(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	g.Select = &Select{ID: ev.ID, Commit: p.Commit, RowTree: p.RowTree, Time: ev.TS, Actor: ev.Actor}
	g.seen = append(g.seen, ev.ID)
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
// p is the payload of ev.
func GuardLand(t Tables, ev events.Event, p *events.Land) error {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToLand, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardOwnerLands(p.Role); err != nil {
		return err
	}
	if err := GuardGoalActor(g, ev.Actor, p.Role); err != nil {
		return err
	}
	if err := GuardLandedFirst(t, g, p.Commit, nil); err != nil {
		return err
	}
	if p.Commit != "" {
		return GuardEvidence(g, *p)
	}
	switch {
	case p.Stop == StopSelected && g.Select != nil:
	case (p.Stop == StopTwoCandidates || p.Stop == StopBudgetUsed) && g.Select == nil:
	default:
		return fault.Refuse(fault.GuardStop, fault.StopNotHeld, p.Stop, g.ID)
	}
	return nil
}

// GuardEvidence is the table half of a land that names a commit: the commit's
// row is not one an unselect rejected, the Goal's select, when an event holds
// one, is on that row, and the land carries a met verdict for every current
// example of the Goal at its current hash. The act ran every example this
// clone held no Result for; replay reads the verdicts the land carries, so
// every clone ends the Goal from git alone.
func GuardEvidence(g *Goal, p events.Land) error {
	if err := GuardEvidenceRow(g, p.Commit, p.RowTree); err != nil {
		return err
	}
	if g.Select != nil && !g.Select.Calculated && g.Select.RowTree != p.RowTree {
		return fault.Refuse(fault.GuardEvidence, fault.NotTheSelect, p.Commit, g.ID, g.Select.Commit)
	}
	verdicts := map[string]events.LandVerdict{}
	for _, v := range p.Verdicts {
		verdicts[v.Line+"/"+v.Example] = v
	}
	for _, x := range g.Placed() {
		v, ok := verdicts[x.Key]
		switch {
		case !ok || v.Hash != x.Hash:
			return fault.Refuse(fault.GuardEvidence, fault.EvidenceMissing, p.Commit, x.Key)
		case v.Verdict != "met":
			return fault.Refuse(fault.GuardEvidence, fault.NotMetOnCommit, x.Key, p.Commit, v.Verdict)
		}
	}
	return nil
}

// GuardEvidenceRow is the half of Evidence that reads no verdict: the Goal is
// not written, and the row of commit, rowTree, is not one an unselect
// rejected. The land act checks it before it runs a Probe.
func GuardEvidenceRow(g *Goal, commit, rowTree string) error {
	if g.State == StateWritten && !g.openAtReplay() {
		return fault.Refuse(fault.GuardEvidence, fault.LandNotOpen, g.ID, g.State)
	}
	if g.Rejected[rowTree] {
		return fault.Refuse(fault.GuardEvidence, fault.CommitUnselected, commit, g.ID)
	}
	return nil
}

// replayLand applies a land the table guards allow: the Goal is landed, and
// keeps the stop the land names.
func replayLand(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.Land)
	if err := GuardLand(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	g.Land = &Land{ID: ev.ID, Stop: p.Stop, Cause: p.Cause, Time: ev.TS, Actor: ev.Actor, Why: ev.Why,
		Commit: p.Commit, Tree: p.Tree, RowTree: p.RowTree, Verdicts: p.Verdicts, Read: p.Read}
	if p.Commit != "" && p.Stop == StopSelected && g.Select == nil {
		// the land carries the select it approves: every clone reads the
		// ended Goal selected on that commit
		g.Select = &Select{ID: ev.ID, Commit: p.Commit, RowTree: p.RowTree, Time: ev.TS, Actor: ev.Actor}
	}
	g.State = StateEnded
}

// GuardOwnerCloses refuses a close by any role but the owner.
func GuardOwnerCloses(role string) error {
	if role != "owner" {
		return fault.Refuse(fault.GuardOwnerClose, fault.NotOwnerClose, role)
	}
	return nil
}

// GuardClose is the check of a close event, called by the close act and by
// Replay at the event's place in the merge order: Not ended, Owner closes,
// then Goal's actor. A close needs no stop. Every guard reads only the
// tables. ev.Node is the goal.add event id.
// p is the payload of ev.
func GuardClose(t Tables, ev events.Event, p *events.Close) error {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToClose, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardOwnerCloses(p.Role); err != nil {
		return err
	}
	return GuardGoalActor(g, ev.Actor, p.Role)
}

// replayClose applies a close the guards allow: the Goal is ended, and keeps
// its rows, verdicts and select as they were.
func replayClose(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.Close)
	if err := GuardClose(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	g.Close = &Close{ID: ev.ID, Time: ev.TS, Actor: ev.Actor, Why: ev.Why}
	g.State = StateEnded
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
// p is the payload of ev.
func GuardUnselect(t Tables, ev events.Event, p *events.Unselect) error {
	g, ok := t.Goals[ev.Node]
	if !ok {
		return fault.Refuse(fault.NoGuard, fault.NoGoalToUnselect, GoalID(ev.Node))
	}
	if err := GuardNotEnded(g); err != nil {
		return err
	}
	if err := GuardOwnerUnselects(p.Role); err != nil {
		return err
	}
	if err := GuardGoalActor(g, ev.Actor, p.Role); err != nil {
		return err
	}
	if p.RowTree != "" {
		// an unselect that names its commit rejects that row; the act
		// checked that this clone showed the Goal selected on it, which
		// reads the Results of its trace
		if err := GuardSelectedOn(g, p.Commit, p.RowTree); err != nil {
			return err
		}
		if g.Rejected[p.RowTree] {
			return fault.Refuse(fault.GuardSelected, fault.AlreadyUnselected, g.ID, p.Commit)
		}
		return nil
	}
	return GuardHasSelect(g)
}

// GuardSelectedOn is the half of Selected that an unselect names the commit
// the Goal is selected on: the select, when the Goal has one, is on the row
// of commit, rowTree.
func GuardSelectedOn(g *Goal, commit, rowTree string) error {
	if g.Select != nil && g.Select.RowTree != rowTree {
		return fault.Refuse(fault.GuardSelected, fault.NotSelectedOn, g.ID, commit, g.Select.Commit)
	}
	return nil
}

// replayUnselect applies an unselect the guards allow: the Goal loses its
// select and is open again, its Results and cells stay as they were, and the
// budget's boundary moves to the unselect.
func replayUnselect(t *Tables, ev events.Event, payload any, refuse func(events.Event, error)) {
	p := payload.(*events.Unselect)
	if err := GuardUnselect(*t, ev, p); err != nil {
		refuse(ev, err)
		return
	}
	g := t.Goals[ev.Node]
	undone := Select{Commit: p.Commit, RowTree: p.RowTree}
	if g.Select != nil {
		undone = *g.Select
	}
	if g.Rejected == nil {
		g.Rejected = map[string]bool{}
	}
	g.Rejected[undone.RowTree] = true
	versions := map[string]bool{}
	for _, hist := range g.Placements {
		for _, p := range hist {
			versions[p.Key+"@"+p.Hash] = true
		}
	}
	g.Unselects = append(g.Unselects, Unselect{
		ID: ev.ID, Select: undone, Time: ev.TS, Actor: ev.Actor, Results: len(g.Results), Versions: versions,
	})
	g.Select = nil
	g.State = StateOpen
	g.seen = append(g.seen, ev.ID)
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
		if r.Refused != "" {
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
