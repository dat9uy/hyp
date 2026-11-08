// Package replay derives every hyp status from an ordered event list. It is
// pure: no file I/O, no clock, no randomness and no goroutines.
package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Projection is what Replay returns.
type Projection struct {
	Nodes          map[string]*Node     // static fields from node.add / result.record
	Hyps           map[string]*Hyp      // Hypothesis → its derived record
	Questions      map[string]*Question // DesignQuestion → its derived record
	Goal           GoalStatus
	Rejections     []Rejection
	Applied        []event.Event // events that applied, in merge order
	LastCheckpoint *event.Event
	RivalsAtAdd    map[string][]string  // Experiment → live rivals of its parent just before its add
	Exps           map[string]*Exp      // Experiment → its derived record (experiment.go)
	Results        map[string]*Res      // Result → its derived record (result.go)
	Edges          []Edge               // every edge in apply order; Live is false once removed
	Decisions      map[string]*Decision // DQ → its Decision record (decision.go)
	Readings       []Reading            // readings of non-retracted Results, in creation order
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
		return Shown{Word: string(p.Goal.Status)}, true
	}
	return Shown{}, false
}

// StaleCause is one link of a stale chain such as H-5 ← H-3 contradicted ← R-9.
type StaleCause struct {
	Node, Reason string
	Next         *StaleCause
}

// GoalStatus summarises the Goal. Counts keys are DQ states plus "stale".
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
	Targets                                                                             []string
	Children                                                                            []string
	Verdicts                                                                            map[string]string
	VerdictOrder                                                                        []string // Verdicts' keys in recorded order
	Observation                                                                         string
	Evidence                                                                            []string
	EvidenceHashes                                                                      map[string]string
	Limits, Closes                                                                      []string
	Readings                                                                            []event.ReadingItem
	Aims, Applies                                                                       []string
}

// Pin is a recipe ref and the bundle hash an event recorded for it. Pointer
// is the evidence pointer that recorded it, or "" for an Experiment's recipe.
type Pin struct {
	Ref           event.Ref
	Hash, Pointer string
}

// pins lists the pins the Result id depends on: its Experiment's recipe
// unless that is none, then each evidence pointer with a recorded hash, in
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

// Limit is one named limit recorded in a Result, with its derived state.
type Limit struct {
	Key, Result, Text string
	Hyps              []string // Hypotheses the Result has a verdict for, sorted
	State             string   // LimitOpen, LimitClosed or LimitAccepted
	By                string   // the closing Result id, or the accepting owner's actor
	Gating            bool     // Gates the Hypothesis of the question's current select
	Referred          string   // the checkpoint that first referred it to the owner, or ""
}

// Gates reports whether the limit gates the owner confirm of a select of
// hyp: it is open and its Result judged hyp.
func (l Limit) Gates(hyp string) bool {
	return l.State == LimitOpen && slices.Contains(l.Hyps, hyp)
}

// UnjudgedAfter lists exp's targets that were live rivals at its add and that
// neither a non-retracted Result of exp nor extra judged, in target order.
// extra holds the verdicts of a Result about to be recorded, or is nil.
func (p Projection) UnjudgedAfter(exp string, extra map[string]string) []string {
	var judged map[string]string
	if e := p.Exps[exp]; e != nil {
		judged = e.Judged
	}
	return unjudged(p.Nodes[exp].Targets, p.RivalsAtAdd[exp], judged, extra)
}

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
// state. By is the owner checkpoint that closed it, or "" while it is open, and
// Answer is that checkpoint's summary, the owner's words. Question is the
// DesignQuestion the Result belongs to. Applied is the Result that applies
// the answer: one whose applies names the key, or, when no Experiment aims
// at the reading, the first Result under its question after the close.
type Reading struct {
	Key, Result, Text string
	Of                string // the reading whose owner answer this one reads, or ""
	By                string // the owner checkpoint that closed it, or "" while open
	Answer            string // the answer in force: the latest of Answers
	Answers           []Answer
	Question          string
	Applied           string // the Result that applied the answer in force, or ""
}

// Answer is one owner answer to a reading: the close, then each revision, in
// merge order. Applied is the Result that applied it, or "".
type Answer struct {
	By, Text, Applied string
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

// SelectRec is one applied dq.select.
type SelectRec struct {
	ID, Hyp, Actor string
	ScopeChange    bool
	Why            string
}

// Verdict is one insufficient verifier confirm: its event id and its actor.
type Verdict struct {
	ID, Actor string
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
	GoalID    string
	Aliases   map[string]string // alias → node id
	// ClosedReadings holds each reading key an applied owner checkpoint closed.
	ClosedReadings map[string]bool

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
	goal        GoalState
	Applied     []event.Event
	RivalsAtAdd map[string][]string // Experiment → live rivals of its parent just before its add
}

// NewState returns an empty state.
func NewState() *State {
	return &State{
		Facts: Facts{
			Nodes:          map[string]*Node{},
			SeenIDs:        map[string]bool{},
			Retracted:      map[string]bool{},
			Abandoned:      map[string]bool{},
			Withdrawn:      map[string]bool{},
			Hold:           map[string]QuestionState{},
			Decisions:      map[string]*Decision{},
			Aliases:        map[string]string{},
			ClosedReadings: map[string]bool{},
			ResultsByHyp:   map[string][]string{},
			EdgesFrom:      map[string][]int{},
			SupersededN:    map[string]int{},
		},
		Snaps:        map[string][]Snap{},
		EverAnswered: map[string]bool{},
		Hyps:         map[string]*Hyp{},
		Questions:    map[string]*Question{},
		RivalsAtAdd:  map[string][]string{},
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
	ReasonVerifierInsufficient = "verifier_insufficient"
	ReasonAwaitingVerifier     = "awaiting_verifier"
	ReasonAwaitingOwner        = "awaiting_owner"
	ReasonNoCandidates         = "no_candidates"
	ReasonUntested             = "untested"
	ReasonAllContradicted      = "all_contradicted"
	ReasonInconclusive         = "inconclusive"
	ReasonSupportedUnsel       = "supported_unselected"
	StaleReasonStale           = "stale"
	StaleReasonChanged         = "changed_since_basis"
	StaleReasonContradicts     = "contradicts"
	LimitOpen                  = "open"
	LimitClosed                = "closed"
	LimitAccepted              = "accepted"
)
