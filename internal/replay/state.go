// Package replay derives every hyp status from an ordered event list. It is
// pure: no file I/O, no clock, no randomness and no goroutines.
package replay

import (
	"github.com/dat9uy/hyp/internal/event"
)

// Projection is what Replay returns.
type Projection struct {
	Nodes          map[string]*Node     // static fields from node.add / result.record
	Hyps           map[string]*Hyp      // Hypothesis → its derived record
	Questions      map[string]*Question // DesignQuestion → its derived record
	Goals          []GoalStatus         // one per Goal, in creation order
	Rejections     []Rejection
	Applied        []event.Event                      // events that applied, in merge order
	LastCheckpoint map[string]*event.Event            // Goal → its latest runner checkpoint; an old owner checkpoint is not one
	RivalsAtAdd    map[string][]string                // Experiment → live rivals of its parent just before its add
	Exps           map[string]*Exp                    // Experiment → its derived record (experiment.go)
	Results        map[string]*Res                    // Result → its derived record (result.go)
	Edges          []Edge                             // every edge in apply order; Live is false once removed
	Decisions      map[string]*Decision               // DQ → its Decision record (decision.go)
	Readings       []Reading                          // readings of non-retracted Results, in creation order
	Lines          []Limit                            // every Goal's lines, by Goal in creation order, then in order (spec §8.2)
	Subjects       map[string]event.HypothesisSubject // Hypothesis → its subject, the latest tie
}

// HypStatus, QuestionState, ExpStatus, ResStatus and GoalState are the status
// words of one node kind each, so a word of one kind cannot be compared with
// a word of another (ADR 0013).
type (
	HypStatus     string
	QuestionState string
	ExpStatus     string
	ResStatus     string
	GoalState     string
)

// Open reports whether a question in state s is open or reopened: the
// states that owe every kind of ledger item and may owe a select.
func (s QuestionState) Open() bool { return s == DQOpen || s == DQReopened }

// Obliges reports whether a sub-question in state s is still an obligation
// on its parent Hypothesis (ADR 0014): not withdrawn and not answered, so a
// deferred one counts.
func (s QuestionState) Obliges() bool { return s != DQWithdrawn && s != DQAnswered }

// Hyp is the derived record of one Hypothesis (spec §7.1, §7.3).
type Hyp struct {
	Status HypStatus
	Stale  *StaleCause // nil when not stale
}

// Question is the derived record of one DesignQuestion (spec §7.3, §7.4). Its
// Decision, in Projection.Decisions under the same id, holds what was stored.
type Question struct {
	State      QuestionState
	Reason     string      // why it is open or reopened; "" otherwise
	Stale      *StaleCause // nil when not stale
	Candidates []string    // its candidates (spec §7.4)
}

// Shown is any node's status as code that prints any node reads it.
type Shown struct {
	Word, Reason string
	Stale        *StaleCause
}

// Display returns the status of a node whose kind the caller does not know;
// ok is false when p holds no such node. It is for printing: code that knows
// the kind reads Hyps, Questions, Exps, Results or Goal.
func (p Projection) Display(id string) (s Shown, ok bool) {
	if h := p.Hyps[id]; h != nil {
		return Shown{Word: string(h.Status), Stale: h.Stale}, true
	}
	if q := p.Questions[id]; q != nil {
		return Shown{Word: string(q.State), Reason: q.Reason, Stale: q.Stale}, true
	}
	if e := p.Exps[id]; e != nil {
		return Shown{Word: string(e.Status)}, true
	}
	if r := p.Results[id]; r != nil {
		return Shown{Word: string(r.Status)}, true
	}
	if n := p.Nodes[id]; n != nil && n.Kind == event.KindGoal {
		return Shown{Word: string(p.GoalStatusOf(id).Status)}, true
	}
	return Shown{}, false
}

// StaleCause is one link of a stale chain such as H-5 ← H-3 contradicted ← R-9.
type StaleCause struct {
	Node, Reason string
	Next         *StaleCause
}

// GoalStatus summarises one Goal. Counts keys are DQ states plus "stale",
// counted over the nodes whose parents lead to the Goal.
type GoalStatus struct {
	Status GoalState
	Counts map[string]int
	GoalID string
}

// Rejection is an event that replay skipped because Validate refused it.
type Rejection struct {
	EventID string
	Reason  string
}

// Node holds the static fields of a node.
type Node struct {
	ID, Kind, Parent, Title, Body, Alias, Recipe, RecipeHash, DecidedBy, Actor, AddedBy string
	// Rubric is an Experiment's rubric (event.RubricReading, a NAME@VERSION
	// pinned by RubricHash, or ""); an old decided_by reading reads as
	// RubricReading. Records are its pinned records.
	Rubric, RubricHash string
	Records            []event.SubjectFile
	Targets            []string
	Children           []string
	Verdicts           map[string]string
	VerdictOrder       []string // Verdicts' keys in recorded order
	Observation        string
	Evidence           []string
	EvidenceHashes     map[string]string
	EvidenceSHA256     map[string]string // a Result's: each file pointer's sha256 when it was recorded
	Limits, Closes     []string
	Readings           []event.ReadingItem
	Aims, Applies      []string
	Subjects           []event.RunSubject // a Result's: one subject per cited run; an old Result's fingerprints read as one with no base
}

// Pin is a recipe ref and the bundle hash an event recorded for it. Pointer
// is the evidence pointer that recorded it, or "" for an Experiment's recipe.
type Pin struct {
	Ref           event.Ref
	Hash, Pointer string
	Rubric        bool // the Experiment's rubric pin, a bundle in rubrics/
}

// pins lists the pins the Result id depends on: its Experiment's recipe
// unless that is none, its Experiment's rubric when it names a bundle, then
// each evidence pointer with a recorded hash, in
// the order cited. A Result is questionable when any pin's ref is missing
// from the current recipe hashes or hashes differently.
func pins(nodes map[string]*Node, id string) []Pin {
	r := nodes[id]
	var pins []Pin
	if exp := nodes[r.Parent]; exp != nil && exp.Recipe != event.RecipeNone {
		// validate admits only name@version or none, so the parse holds.
		ref, _ := event.ParseRef(exp.Recipe)
		pins = append(pins, Pin{Ref: ref, Hash: exp.RecipeHash})
	}
	if exp := nodes[r.Parent]; exp != nil && exp.RubricHash != "" {
		// validate admits a rubric_hash only beside a name@version rubric.
		ref, _ := event.ParseRef(exp.Rubric)
		pins = append(pins, Pin{Ref: ref, Hash: exp.RubricHash, Rubric: true})
	}
	seen := map[string]bool{}
	for _, e := range r.Evidence {
		h, ok := r.EvidenceHashes[e]
		if !ok || seen[e] {
			continue
		}
		seen[e] = true
		ref, _ := event.EvidenceRef(e)
		pins = append(pins, Pin{Ref: ref, Hash: h, Pointer: e})
	}
	return pins
}

// Limit is one keyed limit with its derived state: a limit a Result
// recorded, keyed <result-id>#<n>, or a line of the Goal, keyed
// <goal-id>#<n>, whose Result is the Goal id and whose Question is "".
type Limit struct {
	Key, Result, Text string
	Question          string   // the question of its Result; "" for a line
	Hyps              []string // Hypotheses the Result has a verdict for, sorted; none for a line
	Hand              Hand     // the latest refer, return or accept (§8.2); Move is "" when none
	Closers           []string // every non-retracted Result that closes it, in creation order
	Cites             []Cite   // a line's cites, from its Goal's add; none for a limit
}

// Cite is one cite of a line: the cited key or Hypothesis id, what it meant
// when the line was written (Meaning) and what it would mean now (Now). Both
// are derived from the key's state; a change between them is a hyp check
// warning and changes no status (ADR 0020).
type Cite struct {
	Key, Meaning, Now string
}

// A limit has two facts, and no word joins them: its hand, which the move
// table stores and reads, and its closers, derived from Results' closes.
// Coverage of a line is a third, derived per candidate (Projection.Covers),
// and no move reads it.

// IsLine reports whether l is a line of the Goal.
func (l Limit) IsLine() bool { return l.Question == "" }

// Referred reports whether the latest hand is a refer: the owner's turn.
func (l Limit) Referred() bool { return l.Hand.Move == HandRefer }

// Returned reports whether the latest hand is a return: the runner's turn again.
func (l Limit) Returned() bool { return l.Hand.Move == HandReturn }

// Accepted reports whether the latest hand is an accept: this tree does not
// run the case.
func (l Limit) Accepted() bool { return l.Hand.Move == HandAccept }

// Closed reports whether a non-retracted Result closes a Result's limit. A
// line is never closed: whether its case ran is per candidate.
func (l Limit) Closed() bool { return !l.IsLine() && len(l.Closers) > 0 }

// Open reports whether the limit is neither closed nor accepted. For a line
// it means only that the owner has not accepted it.
func (l Limit) Open() bool { return !l.Closed() && !l.Accepted() }

// CurrentSelect returns the DQ's current select; ok is false when the DQ's
// last select or unselect is an unselect, or there is none.
func (p Projection) CurrentSelect(dq string) (rec SelectRec, ok bool) {
	sel := p.Decisions[dq]
	if sel == nil || sel.Current == "" {
		return SelectRec{}, false
	}
	return sel.Selects[sel.Current], true
}

// Reading is one reading of a Goal line recorded in a Result, with its derived
// state. By is the owner answer, or old owner checkpoint, that closed it, or
// "" while it is open, and Answer is the owner's words in force. Question is the
// DesignQuestion the Result belongs to. Applied is the Result that applies
// the answer: one whose applies names the key, or, when no Experiment aims
// at the reading, the first Result under its question after the close.
type Reading struct {
	Key, Result, Text string
	Of                string // the reading whose owner answer this one reads, or ""
	By                string // the owner answer or old owner checkpoint that closed it, or "" while open
	Answer            string // the answer in force: the latest of Answers
	Answers           []Answer
	Question          string
	Applied           string // the Result that applied the answer in force, or ""
	// Reach lists the live Hypotheses under Question whose standing the
	// answer can move, read from the Result's own verdicts: a candidate it
	// judged, or a contradicted rival whose only contradicts is its own. An
	// open reading makes its question wait only while Reach is not empty.
	Reach []string
}

// Open reports whether no owner answer has closed the reading yet.
func (l Reading) Open() bool { return l.By == "" }

// Answer is one owner answer to a reading: the close, then each revision, in
// merge order. Applied is the Result that applied it, or "".
type Answer struct {
	By, Text, Applied string
	ByAnswer          bool // By is an answer event; otherwise an old owner checkpoint
}

// LimitsOf lists the limits recorded under question dq, in creation order.
func (p Projection) LimitsOf(dq string) []Limit {
	if s := p.Decisions[dq]; s != nil {
		return s.Limits
	}
	return nil
}

// LiveEdges lists the live edges in apply order.
func (p Projection) LiveEdges() []Edge {
	var live []Edge
	for _, e := range p.Edges {
		if e.Live {
			live = append(live, e)
		}
	}
	return live
}

// Edge is one edge.add; Live turns false on a matching edge.remove.
type Edge struct {
	ID, From, Type, To string
	Live               bool
}

// SelectRec is one applied dq.select. Candidates is not typed: replay
// stores the question's candidates when the select applies, as it stores
// RivalsAtAdd, so a later candidate reopens the question (new_candidate).
type SelectRec struct {
	ID, Hyp, Actor string
	Why            string
	Candidates     []string // the question's candidates just before the select applied
}

// Snap records each depends_on target's state right after a basis event.
type Snap struct {
	EventID string
	Targets map[string]SnapVal
}

// SnapVal is a target's (status, stale) pair inside a Snap.
type SnapVal struct {
	Status HypStatus
	Stale  bool
}

// Facts is what the applied events stored, and lookup indexes over it. A replay
// rule takes Facts and nothing else, so reading a derived status from one does
// not compile (spec §5.1, ADR 0013).
type Facts struct {
	Nodes     map[string]*Node
	Order     []string // node ids in creation order
	SeenIDs   map[string]bool
	Retracted map[string]bool
	Abandoned map[string]bool
	Withdrawn map[string]bool // Experiment → it has an applied experiment.withdraw
	Edges     []Edge
	Hold      map[string]QuestionState // DQ → "", DQDeferred or DQWithdrawn
	Decisions map[string]*Decision     // DQ → its Decision record (decision.go)
	// Goals are the store's Goals in creation order. A store holds many;
	// a tree is a view of one Goal and every node whose parents lead to it.
	Goals   []string
	Aliases map[string]string // alias → node id
	// Answers holds, for each reading key an owner answered, the
	// close and then each revise, in merge order (keyed.go).
	Answers map[string][]Answer
	// LimitHand holds, for each limit key a checkpoint referred, returned or
	// accepted, the latest of the three in merge order (keyed.go).
	LimitHand map[string]Hand
	// Lines are each Goal's lines, and LinesBy the event that wrote them:
	// the Goal's own add (its id), or an old owner checkpoint. A Goal is
	// absent from both before one did.
	Lines   map[string][]string
	LinesBy map[string]string
	// Subjects holds each Hypothesis's subject, the latest tie.
	Subjects map[string]event.HypothesisSubject

	// Lookup indexes over the stored state, maintained by apply so that a
	// full recompute per event stays fast.
	ResultsByHyp map[string][]string // hypothesis → Results with a verdict for it, in creation order
	EdgesFrom    map[string][]int    // node → indexes into Edges of edges it starts
	SupersededN  map[string]int      // Hypothesis or Experiment → number of live supersedes edges pointing at it
}

// State is Facts and what recompute and Step derive from them. Snaps,
// EverAnswered and RivalsAtAdd are stored, but record derived status at an
// event, so they stay out of Facts.
type State struct {
	Facts
	// recipeHashes are the current recipe hashes the last Step was given;
	// a Result whose pin they do not match is questionable.
	recipeHashes map[string]string
	Snaps        map[string][]Snap
	EverAnswered map[string]bool
	// The derived records, rebuilt wholesale by recompute, so a map read
	// before a Step keeps the state from before it.
	Hyps        map[string]*Hyp
	Questions   map[string]*Question
	goals       map[string]GoalState // Goal → its derived state
	Applied     []event.Event
	RivalsAtAdd map[string][]string // Experiment → live rivals of its parent just before its add
	// LineCites holds each Goal's lines' cites with what each meant at the
	// Goal's add: derived from Facts there and kept, as RivalsAtAdd is, so
	// no replay rule reads it.
	LineCites map[string][][]Cite
}

// NewState returns an empty state.
func NewState() *State {
	return &State{
		Facts: Facts{
			Nodes:        map[string]*Node{},
			SeenIDs:      map[string]bool{},
			Retracted:    map[string]bool{},
			Abandoned:    map[string]bool{},
			Withdrawn:    map[string]bool{},
			Hold:         map[string]QuestionState{},
			Decisions:    map[string]*Decision{},
			Aliases:      map[string]string{},
			Answers:      map[string][]Answer{},
			LimitHand:    map[string]Hand{},
			Subjects:     map[string]event.HypothesisSubject{},
			Lines:        map[string][]string{},
			LinesBy:      map[string]string{},
			ResultsByHyp: map[string][]string{},
			EdgesFrom:    map[string][]int{},
			SupersededN:  map[string]int{},
		},
		Snaps:        map[string][]Snap{},
		EverAnswered: map[string]bool{},
		Hyps:         map[string]*Hyp{},
		Questions:    map[string]*Question{},
		goals:        map[string]GoalState{},
		RivalsAtAdd:  map[string][]string{},
		LineCites:    map[string][][]Cite{},
	}
}

// The status words of each node kind.
const (
	HypProposed     HypStatus = "proposed"
	HypSupported    HypStatus = "supported"
	HypContradicted HypStatus = "contradicted"
	HypInconclusive HypStatus = "inconclusive"
	HypSuperseded   HypStatus = "superseded"
	HypAbandoned    HypStatus = "abandoned"

	DQOpen      QuestionState = "open"
	DQReopened  QuestionState = "reopened"
	DQAnswered  QuestionState = "answered"
	DQDeferred  QuestionState = "deferred"
	DQWithdrawn QuestionState = "withdrawn"

	ExpPlanned    ExpStatus = "planned"
	ExpRun        ExpStatus = "run"
	ExpSuperseded ExpStatus = "superseded"
	ExpWithdrawn  ExpStatus = "withdrawn"

	ResRecorded  ResStatus = "recorded"
	ResRetracted ResStatus = "retracted"

	GoalComplete   GoalState = "complete"
	GoalInProgress GoalState = "in_progress"
)

// Why a question is open, why a node is stale, and a limit's state.
const (
	ReasonConflicting          = "conflicting_selections"
	ReasonSelectedNotSupport   = "selected_not_supported"
	ReasonSelectedStale        = "selected_stale"
	ReasonAwaitingSubquestions = "awaiting_subquestions"
	ReasonNewCandidate         = "new_candidate"
	ReasonNoCandidates         = "no_candidates"
	ReasonUntested             = "untested"
	ReasonAllContradicted      = "all_contradicted"
	ReasonInconclusive         = "inconclusive"
	ReasonSupportedUnsel       = "supported_unselected"
	StaleReasonStale           = "stale"
	StaleReasonChanged         = "changed_since_basis"
	StaleReasonContradicts     = "contradicts"
)

// goalOf returns the Goal that node id's parents lead to, or "" when none
// does.
func goalOf(nodes map[string]*Node, id string) string {
	for n := nodes[id]; n != nil; n = nodes[n.Parent] {
		if n.Kind == event.KindGoal {
			return n.ID
		}
	}
	return ""
}

// GoalOf returns the Goal that node id's parents lead to, or "" when none
// does. A tree is the view of one Goal: the Goal and every node with it as
// GoalOf.
func (p Projection) GoalOf(id string) string { return goalOf(p.Nodes, id) }

// KeyGoal returns the Goal that limit, line or reading key k leads to: a
// line's own Goal, or the Goal of a limit's or reading's Result.
func (p Projection) KeyGoal(k string) string { return goalOf(p.Nodes, event.KeyResult(k)) }

// EventOf reports whether event ev belongs to Goal g: its node leads to g,
// or, for an answer, which is on no node, one of its keys does.
func (p Projection) EventOf(ev event.Event, g string) bool {
	if ev.Node != "" {
		return goalOf(p.Nodes, ev.Node) == g
	}
	var a event.Answer
	if ev.Type != event.TypeAnswer || event.DecodePayload(ev, &a) != nil {
		return false
	}
	for _, items := range [][]event.KeyText{a.Answers, a.Accepts, a.Returns} {
		for _, it := range items {
			if p.KeyGoal(it.Key) == g {
				return true
			}
		}
	}
	return false
}

// GoalStatusOf returns Goal g's summary, or a zero GoalStatus when g is not
// a Goal.
func (p Projection) GoalStatusOf(g string) GoalStatus {
	for _, gs := range p.Goals {
		if gs.GoalID == g {
			return gs
		}
	}
	return GoalStatus{}
}

// LinesOf lists Goal g's lines, in order.
func (p Projection) LinesOf(g string) []Limit {
	var out []Limit
	for _, l := range p.Lines {
		if l.Result == g {
			out = append(out, l)
		}
	}
	return out
}
