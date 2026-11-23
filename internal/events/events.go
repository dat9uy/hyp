// Package events is the v2 event envelope and the payload of each event type.
// A line that is not v2, or has a type outside the closed set, does not parse.
package events

import (
	"bytes"
	"encoding/json"

	"github.com/dat9uy/hyp/internal/fault"
)

// Version is the envelope version every line carries.
const Version = 2

// Event is one line of an events file.
type Event struct {
	V       int             `json:"v"`
	ID      string          `json:"id"`
	TS      string          `json:"ts"`
	Actor   string          `json:"actor"`
	Type    string          `json:"type"`
	Node    string          `json:"node"`
	Payload json.RawMessage `json:"payload"`
	Why     string          `json:"why"`
}

// The eleven event types. example.place is the old form of a place, one
// example per event: hyp writes examples.place, and replays both.
const (
	TypeGoalAdd       = "goal.add"
	TypeExamplePlace  = "example.place"
	TypeExamplesPlace = "examples.place"
	TypeResultRecord  = "result.record"
	TypeResultRetract = "result.retract"
	TypeSelect        = "select"
	TypeBudgetReset   = "budget.reset"
	TypeLand          = "land"
	TypeClose         = "close"
	TypeUnselect      = "unselect"
	TypeRead          = "read"
)

// Where an event kind is stored.
const (
	HomeGit   = "git"   // the store's .tree/events/, which git tracks and merges
	HomeTrace = "trace" // the clone's trace folder under $HOME/.hyp/traces
)

// Kind is the entry of one event type: whether it changes a Goal's state,
// which Goal it belongs to, and whether an end freezes it. One rule stores
// every kind: a kind that changes a Goal's state is stored in git, so every
// clone replays the same Goal states from git, and every other kind is stored
// in the clone's trace. hyp acts on these entries: the store writes an event
// only where its entry says, replay freezes an event only when its entry says
// so, and hyp spec prints them.
type Kind struct {
	Type    string
	Changes bool   // it changes a Goal's state: written, open, selected or ended
	Goal    string // the Goal the event belongs to, as its node names it
	Frozen  bool   // an end of its Goal refuses it at replay when the end's snapshot does not name it
	Writer  string // the act that writes it
}

// Home is where the kind is stored: git when it changes a Goal's state, else
// the clone's trace.
func (k Kind) Home() string {
	if k.Changes {
		return HomeGit
	}
	return HomeTrace
}

// Kinds are the entries of the eleven event types, in the order hyp spec
// prints them. A Result changes no Goal's state: the state is calculated from
// the places, selects and ends alone, so a clone that holds no Result replays
// the same state. A result.record that an earlier hyp wrote into git is still
// replayed from there.
var Kinds = []Kind{
	{TypeGoalAdd, true, "the Goal it adds, or the written Goal it replaces", true, "add goal"},
	{TypeExamplePlace, true, "the Goal its node names; the old form of a place, replayed and no longer written", true, "none"},
	{TypeExamplesPlace, true, "the Goal its node names", true, "place"},
	{TypeResultRecord, false, "the Goal its node names", false, "run"},
	{TypeResultRetract, false, "the Goal of the Result its node names", false, "retract, not built"},
	{TypeSelect, true, "the Goal its node names", true, "select"},
	{TypeBudgetReset, true, "the Goal its node names", false, "reset, not built"},
	{TypeLand, true, "the Goal its node names; it ends the Goal", false, "land"},
	{TypeClose, true, "the Goal its node names; it ends the Goal", false, "close"},
	{TypeUnselect, true, "the Goal its node names", true, "unselect"},
	{TypeRead, false, "none: it is what hyp printed to an actor, and changes no Goal", false, "every command that names an actor"},
}

// KindOf returns the entry of an event type, and false for a type outside the
// closed set.
func KindOf(typ string) (Kind, bool) {
	for _, k := range Kinds {
		if k.Type == typ {
			return k, true
		}
	}
	return Kind{}, false
}

// The kind of a line: a gap line must fail on H0, a guard line must be met.
const (
	KindGap   = "gap"
	KindGuard = "guard"
)

// Line is one line of a Goal.
type Line struct {
	Key  string `json:"key"`
	Text string `json:"text"`
	Kind string `json:"kind"`
}

// GoalAdd is the payload of goal.add. From is a Goal id or empty.
type GoalAdd struct {
	Role     string `json:"role"`
	Words    string `json:"words"`
	Base     string `json:"base"`
	BaseTree string `json:"base_tree"`
	Budget   int    `json:"budget"`
	From     string `json:"from"`
	Lines    []Line `json:"lines"`
}

// Cell is H0's verdict on an example, written by the placement that adds it.
// A place names no output: H0's outputs go to the clone's trace. Output and
// Stderr are set only in a place an earlier hyp wrote, whose outputs are under
// .tree/outputs.
type Cell struct {
	Commit  string `json:"commit"`
	RowTree string `json:"row_tree"`
	Probe   string `json:"probe"`
	Output  string `json:"output,omitempty"`
	Stderr  string `json:"stderr,omitempty"`
	Exit    int    `json:"exit"`
	Verdict string `json:"verdict"`
}

// ExamplePlace is the payload of example.place.
type ExamplePlace struct {
	Role    string `json:"role"`
	Line    string `json:"line"`
	Example string `json:"example"`
	Hash    string `json:"hash"`
	H0      Cell   `json:"h0"`
}

// ExamplesPlace is the payload of examples.place: every example of the set
// that the place changed, each with its hash and H0's verdict. One place is one event, so a
// crash leaves the whole set placed or none of it.
type ExamplesPlace struct {
	Role     string          `json:"role"`
	Examples []PlacedExample `json:"examples"`
}

// PlacedExample is one example of an examples.place event.
type PlacedExample struct {
	Line    string `json:"line"`
	Example string `json:"example"`
	Hash    string `json:"hash"`
	H0      Cell   `json:"h0"`
}

// ExampleCell is the cell of one example in a Result.
type ExampleCell struct {
	Example string `json:"example"`
	Hash    string `json:"hash"`
	Probe   string `json:"probe"`
	Output  string `json:"output"`
	Stderr  string `json:"stderr"`
	Exit    int    `json:"exit"`
	Verdict string `json:"verdict"`
}

// ResultRecord is the payload of result.record. Refused is empty, or
// "Stable" with the example keys whose verdict differed in Differing.
type ResultRecord struct {
	Role      string        `json:"role"`
	Line      string        `json:"line"`
	Commit    string        `json:"commit"`
	RowTree   string        `json:"row_tree"`
	Refused   string        `json:"refused"`
	Differing []string      `json:"differing,omitempty"`
	Cells     []ExampleCell `json:"cells"`
}

// ResultRetract is the payload of result.retract.
type ResultRetract struct {
	Role string `json:"role"`
}

// Select is the payload of select.
type Select struct {
	Role    string `json:"role"`
	Commit  string `json:"commit"`
	RowTree string `json:"row_tree"`
}

// Land is the payload of land: the stop the land ends, the event of the act
// that caused that stop (a select, or the result.record of the run after
// which the stop held), and the snapshot: the ids of the Goal's events that
// replay applied when the owner landed. A land written before close existed
// names no snapshot.
type Land struct {
	Role  string   `json:"role"`
	Stop  string   `json:"stop"`
	Cause string   `json:"cause"`
	Seen  []string `json:"seen,omitempty"`
}

// Close is the payload of close: the owner's act that ends a Goal that is
// not ended, at any stop or none, and the snapshot as a land names it.
type Close struct {
	Role string   `json:"role"`
	Seen []string `json:"seen"`
}

// Unselect is the payload of unselect: the owner's act that puts a selected
// Goal back to open.
type Unselect struct {
	Role string `json:"role"`
}

// BudgetReset is the payload of budget.reset.
type BudgetReset struct {
	Role string `json:"role"`
}

// Read is the payload of read: what hyp printed to the event's actor for one
// command. Command is the command's arguments after "hyp"; Stdout and Stderr
// name the bytes it printed, stored in the trace's outputs/ by their sha256.
// Role is the role the command declared, or empty for a read that declared
// none.
type Read struct {
	Role    string   `json:"role"`
	Command []string `json:"command"`
	Exit    int      `json:"exit"`
	Stdout  string   `json:"stdout"`
	Stderr  string   `json:"stderr"`
}

// ParseLine parses one line of an events file. It refuses an unknown field, a
// version other than 2 and a type outside the closed set.
func ParseLine(b []byte) (Event, error) {
	var e Event
	if err := strictDecode(b, &e); err != nil {
		return Event{}, err
	}
	if e.V != Version {
		return Event{}, fault.Refuse(fault.NoGuard, fault.EventVersion, e.V, Version)
	}
	if _, ok := KindOf(e.Type); !ok {
		return Event{}, fault.Refuse(fault.NoGuard, fault.EventType, e.Type)
	}
	return e, nil
}

// Decode parses e's payload into dst, refusing an unknown field.
func Decode(e Event, dst any) error {
	if err := strictDecode(e.Payload, dst); err != nil {
		return fault.Refuse(fault.NoGuard, fault.PayloadDecode, e.Type, err)
	}
	return nil
}

func strictDecode(raw []byte, dst any) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return err
	}
	if dec.More() {
		return fault.Refuse(fault.NoGuard, fault.TrailingData)
	}
	return nil
}
