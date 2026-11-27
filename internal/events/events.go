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

// The fifteen event types. example.place is the old form of a place, one
// example per event: hyp writes examples.place, and replays both.
const (
	TypeGoalAdd        = "goal.add"
	TypeExamplePlace   = "example.place"
	TypeExamplesPlace  = "examples.place"
	TypeResultRecord   = "result.record"
	TypeResultRetract  = "result.retract"
	TypeSelect         = "select"
	TypeBudgetReset    = "budget.reset"
	TypeLand           = "land"
	TypeClose          = "close"
	TypeUnselect       = "unselect"
	TypeRead           = "read"
	TypeFindingAdd     = "finding.add"
	TypeFindingFold    = "finding.fold"
	TypeFindingRefute  = "finding.refute"
	TypeFindingDecline = "finding.decline"
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
	Step    string // the step of the kept model whose name its step line carries, or "" for none
}

// Home is where the kind is stored: git when it changes a Goal's state, else
// the clone's trace.
func (k Kind) Home() string {
	if k.Changes {
		return HomeGit
	}
	return HomeTrace
}

// Kinds are the entries of the fifteen event types, in the order hyp spec
// prints them. Git holds only the owner's acts. A Result changes no state
// that replay of git gives: replay calculates the state from the adds,
// places, selects, unselects and ends alone, so every clone replays the same
// state. A clone whose trace holds the Results that show one candidate shows
// the Goal selected on it, with no event. hyp reads an event of a kind that
// changes no state replay gives only from the trace, never from git.
var Kinds = []Kind{
	{TypeGoalAdd, true, "the Goal it adds, or the written Goal it replaces; its payload names the findings the Goal is about, and may carry the Goal's examples with their H0 verdicts", true, "add goal", "AddGoal"},
	{TypeExamplePlace, true, "the Goal its node names; the old form of a place, replayed and no longer written", true, "none", "Place"},
	{TypeExamplesPlace, true, "the Goal its node names; each example names the findings its findings section names", true, "place, by the owner; a place of the grill's is the older form, replayed", "Place"},
	{TypeResultRecord, false, "the Goal its node names; it carries what the run read of git about its commit: the commits of the Goal's rows it descends from, its spec folder and its kept model", false, "run", "Run"},
	{TypeResultRetract, false, "the Goal of the Result its node names", false, "retract, not built", "Retract"},
	{TypeSelect, true, "the Goal its node names", true, "select, by the owner among two candidates; a select of the runner's is the older form, replayed", "Select"},
	{TypeBudgetReset, true, "the Goal its node names", false, "reset, not built", "BudgetReset"},
	{TypeLand, true, "the Goal its node names; it ends the Goal, and may carry the evidence of the commit it lands, with what it read of that commit's spec folder and kept model", false, "land", "Land"},
	{TypeClose, true, "the Goal its node names; it ends the Goal", false, "close", "Close"},
	{TypeUnselect, true, "the Goal its node names; it names the commit it rejects", true, "unselect", "Unselect"},
	{TypeRead, false, "none: it is what hyp printed to an actor, and changes no Goal", false, "every command that names an actor", ""},
	{TypeFindingAdd, false, "none: its node is its own id, and a finding names no Goal: a Goal names it; the old form's node names the Goal it was recorded on; its payload may name an example and a commit that holds it", false, "finding add", "FindingAdd"},
	{TypeFindingFold, false, "the Goal its payload names, which the fold counts as a name of the finding by that Goal; the old form names an issue", false, "none: the old form of a name, replayed and no longer written", "FindingFold"},
	{TypeFindingRefute, false, "the finding its node names", false, "finding refute", "FindingRefute"},
	{TypeFindingDecline, false, "the finding its node names", false, "finding decline", "FindingDecline"},
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

// GoalAdd is the payload of goal.add. From is a Goal id or empty. Findings
// are the finding.add ids of the findings the Goal names: the Goal's record
// keeps only their ids, so it reads whole on a clone whose trace lacks them.
type GoalAdd struct {
	Role     string   `json:"role"`
	Words    string   `json:"words"`
	Base     string   `json:"base"`
	BaseTree string   `json:"base_tree"`
	Budget   int      `json:"budget"`
	From     string   `json:"from"`
	Lines    []Line   `json:"lines"`
	Findings []string `json:"findings,omitempty"`

	// Examples are the examples the owner adds the Goal with, each with its
	// hash and H0's verdict, as a place names them: a Goal added with its
	// examples is open at once. A goal.add written before add goal carried
	// examples names none.
	Examples []PlacedExample `json:"examples,omitempty"`
}

// Cell is H0's verdict on an example, written by the placement that adds it.
// A place names no output: H0's outputs go to the clone's trace. hyp reads
// no output by the Output and Stderr of a place; they stay in the type so
// that the place events that name them still decode.
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

// PlacedExample is one example of an examples.place event. Findings are the
// finding.add ids of the findings its findings section names.
type PlacedExample struct {
	Line     string   `json:"line"`
	Example  string   `json:"example"`
	Hash     string   `json:"hash"`
	H0       Cell     `json:"h0"`
	Findings []string `json:"findings,omitempty"`
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
	Steps   string `json:"steps,omitempty"` // the hash of the run's step log, when its Probe's subject wrote one
}

// ResultRecord is the payload of result.record. Refused is empty, or
// "Stable" with the example keys whose verdict differed in Differing. Read
// is what the run read of git about its commit; a Result recorded before
// runs carried it holds none.
type ResultRecord struct {
	Role      string        `json:"role"`
	Line      string        `json:"line"`
	Commit    string        `json:"commit"`
	RowTree   string        `json:"row_tree"`
	Refused   string        `json:"refused"`
	Differing []string      `json:"differing,omitempty"`
	Cells     []ExampleCell `json:"cells"`
	Read      *CommitRead   `json:"read,omitempty"`
}

// CommitRead is what a run read of git about its commit when it recorded
// its Result: the act reads git, and its event carries what it read, so no
// state or print reads git history again. Rows, leaves, candidates, the
// Spec guard of the calculated select and the prints of a selected commit
// are calculated from it.
type CommitRead struct {
	// Above is each commit of the Goal's rows, H0's included, that the
	// commit descends from, as git showed it when the run was recorded.
	Above []string `json:"above,omitempty"`

	// What the commit's spec folder, goals/G/specs/, holds.
	SpecFiles int      `json:"spec_files"`        // the number of its files
	Model     bool     `json:"model"`             // it holds a .tla file
	Unnamed   []string `json:"unnamed,omitempty"` // the rules of its check configs no variant config names
	Named     []string `json:"named,omitempty"`   // the rules some variant config names
	Charts    []Chart  `json:"charts,omitempty"`  // its charts, outside the variant folders

	// What the commit holds of the kept model, goals/model/.
	Keeps bool       `json:"keeps"`           // the commit keeps a model there
	Steps []KeptStep `json:"steps,omitempty"` // each step of it, in module order
	Kept  []KeptItem `json:"kept,omitempty"`  // what it adds or changes against H0
}

// Chart is one chart (.mmd) of a spec folder: its path below the folder and
// its text.
type Chart struct {
	Path string `json:"path"`
	Text string `json:"text"`
}

// KeptStep is one step of the kept model: the name of its definition and
// the command it takes.
type KeptStep struct {
	Name    string `json:"name"`
	Command string `json:"command"`
}

// KeptItem is one step or rule of the kept model that a commit adds or
// changes. A step is a definition tagged with the command it takes; a rule
// is a name its check configs check, with the source its comment names.
type KeptItem struct {
	Rule    bool   `json:"rule,omitempty"`    // a rule; else a step
	Name    string `json:"name"`              // the name of its definition
	Command string `json:"command,omitempty"` // a step's command
	Source  string `json:"source,omitempty"`  // a rule's source, as its comment names it; empty when none
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
// which the stop held, or that showed the Goal selected), and the snapshot: the ids of the Goal's events that
// replay applied when the owner landed. A land written before close existed
// names no snapshot.
type Land struct {
	Role  string   `json:"role"`
	Stop  string   `json:"stop"`
	Cause string   `json:"cause"`
	Seen  []string `json:"seen,omitempty"`

	// The evidence the land approves, when it lands a commit: the commit,
	// its tree, its row tree, and each current example's verdict on it. A
	// land written before land carried evidence names none.
	Commit   string        `json:"commit,omitempty"`
	Tree     string        `json:"tree,omitempty"`
	RowTree  string        `json:"row_tree,omitempty"`
	Verdicts []LandVerdict `json:"verdicts,omitempty"`

	// Read is what the land read of git about the commit it lands, its
	// spec folder and kept model, so every clone prints the ended Goal from
	// the event. A land written before land carried it holds none.
	Read *CommitRead `json:"read,omitempty"`
}

// LandVerdict is one example's verdict on the commit a land names: the
// example key's line and name, its hash, and its verdict there.
type LandVerdict struct {
	Line    string `json:"line"`
	Example string `json:"example"`
	Hash    string `json:"hash"`
	Verdict string `json:"verdict"`
}

// Close is the payload of close: the owner's act that ends a Goal that is
// not ended, at any stop or none, and the snapshot as a land names it.
type Close struct {
	Role string   `json:"role"`
	Seen []string `json:"seen"`
}

// Unselect is the payload of unselect: the owner's act that puts a selected
// Goal back to open. Commit and RowTree name the commit it rejects and its
// row tree: no clone shows the Goal selected on that row again. An unselect
// written before it named its commit names none, and undoes the select
// event it follows.
type Unselect struct {
	Role    string `json:"role"`
	Commit  string `json:"commit,omitempty"`
	RowTree string `json:"row_tree,omitempty"`
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

// FindingAdd is the payload of finding.add: a fact an agent found, and a
// pointer to its evidence, such as a Result id, a commit or a path. Its node
// is its own id: a finding names no Goal; a Goal names it. The old form's
// node names the Goal it was recorded on. Draft
// is an add goal command line the agent drafted for the owner, or empty.
// Role is the role the act declared, or empty. Example and Commit, set
// together or not at all, name an example as G-XXXXXX/line-N/name and the
// full sha of a commit that holds it: the finding is about the steps of the
// kept model that example's run section takes on that commit.
type FindingAdd struct {
	Role     string `json:"role"`
	Fact     string `json:"fact"`
	Evidence string `json:"evidence"`
	Draft    string `json:"draft,omitempty"`
	Example  string `json:"example,omitempty"`
	Commit   string `json:"commit,omitempty"`
}

// FindingFold is the payload of finding.fold, the old form of a name, which
// replay still reads and no act writes: the findings it folds, by their
// finding.add ids, and the goal.add id of the Goal it folds them into, which
// counts as a name of each by that Goal. Its oldest form names the number of
// an issue in place of a Goal.
type FindingFold struct {
	Role     string   `json:"role"`
	Findings []string `json:"findings"`
	Goal     string   `json:"goal,omitempty"`
	Issue    int      `json:"issue,omitempty"`
}

// FindingRefute is the payload of finding.refute, whose node is the
// finding.add id of the finding it refutes.
type FindingRefute struct {
	Role string `json:"role"`
}

// FindingDecline is the payload of finding.decline, whose node is the
// finding.add id of the finding the owner declines; the event's why is the
// owner's reason not to fix it.
type FindingDecline struct {
	Role string `json:"role"`
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
