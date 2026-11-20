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

// The nine event types. All but read change the state machine; read is the
// trace of what hyp printed to an actor, and changes no state.
const (
	TypeGoalAdd       = "goal.add"
	TypeExamplePlace  = "example.place"
	TypeResultRecord  = "result.record"
	TypeResultRetract = "result.retract"
	TypeSelect        = "select"
	TypeBudgetReset   = "budget.reset"
	TypeLand          = "land"
	TypeUnselect      = "unselect"
	TypeRead          = "read"
)

var types = map[string]bool{
	TypeGoalAdd: true, TypeExamplePlace: true, TypeResultRecord: true,
	TypeResultRetract: true, TypeSelect: true, TypeBudgetReset: true,
	TypeLand: true, TypeUnselect: true, TypeRead: true,
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

// Cell is a cell of H0's row, written by the placement that adds an example.
type Cell struct {
	Commit  string `json:"commit"`
	RowTree string `json:"row_tree"`
	Probe   string `json:"probe"`
	Output  string `json:"output"`
	Stderr  string `json:"stderr"`
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

// Land is the payload of land: the stop the land ends, and the event of the
// act that caused that stop (a select, or the result.record of the run after
// which the stop held).
type Land struct {
	Role  string `json:"role"`
	Stop  string `json:"stop"`
	Cause string `json:"cause"`
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
// name the bytes it printed, stored under .tree/outputs/ by their sha256.
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
// version other than 2 and a type outside the nine.
func ParseLine(b []byte) (Event, error) {
	var e Event
	if err := strictDecode(b, &e); err != nil {
		return Event{}, err
	}
	if e.V != Version {
		return Event{}, fault.Refuse(fault.NoGuard, fault.EventVersion, e.V, Version)
	}
	if !types[e.Type] {
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
