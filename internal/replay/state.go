// Package replay derives every hyp status from an ordered event list. It is
// pure: no file I/O, no clock, no randomness and no goroutines.
package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Projection is what Replay returns.
type Projection struct {
	Nodes          map[string]*Node      // static fields from node.add / result.record
	Status         map[string]NodeStatus // derived
	Goal           GoalStatus
	Rejections     []Rejection
	Applied        []event.Event // events that applied, in merge order
	LastCheckpoint *event.Event
	RivalsAtAdd    map[string][]string  // Experiment → live rivals of its parent just before its add
	Exps           map[string]*Exp      // Experiment → its derived record (experiment.go)
	Results        map[string]*Res      // Result → its derived record (result.go)
	Edges          []Edge               // every edge in apply order; Live is false once removed
	Decisions      map[string]*Decision // DQ → its Decision record (decision.go)
}

// NodeStatus is the derived state of one node.
type NodeStatus struct {
	Status    string      // hypothesis status, DQ state, experiment planned/run/superseded/withdrawn, result recorded/retracted, goal state
	Reason    string      // DQ open/reopened reason; "" otherwise
	Stale     *StaleCause // nil when not stale
	Suggested []string    // DQ only
}

// StaleCause is one link of a stale chain such as H-5 ← H-3 contradicted ← R-9.
type StaleCause struct {
	Node, Reason string
	Next         *StaleCause
}

// GoalStatus summarises the Goal. Counts keys are DQ states plus "stale".
type GoalStatus struct {
	Status string
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
	Tradeoff       bool
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
	Status string
	Stale  bool
}

// State is the stored state built by applying events.
type State struct {
	Nodes map[string]*Node
	// recipeHashes are the current recipe hashes the last Step was given;
	// a Result whose pin they do not match is questionable.
	recipeHashes map[string]string
	Order        []string // node ids in creation order
	SeenIDs      map[string]bool
	Retracted    map[string]bool
	Abandoned    map[string]bool
	Withdrawn    map[string]bool // Experiment → it has an applied experiment.withdraw
	Edges        []Edge
	Hold         map[string]string    // DQ → "", "deferred" or "withdrawn"
	Decisions    map[string]*Decision // DQ → its Decision record (decision.go)
	Snaps        map[string][]Snap
	EverAnswered map[string]bool
	Derived      map[string]NodeStatus
	Applied      []event.Event
	GoalID       string
	Aliases      map[string]string // alias → node id

	// Lookup indexes over the stored state, maintained by apply so that a
	// full recompute per event stays fast.
	ResultsByHyp map[string][]string // hypothesis → Results with a verdict for it, in creation order
	EdgesFrom    map[string][]int    // node → indexes into Edges of edges it starts
	SupersededN  map[string]int      // Hypothesis or Experiment → number of live supersedes edges pointing at it
	RivalsAtAdd  map[string][]string // Experiment → live rivals of its parent just before its add
}

// NewState returns an empty state.
func NewState() *State {
	return &State{
		Nodes:        map[string]*Node{},
		SeenIDs:      map[string]bool{},
		Retracted:    map[string]bool{},
		Abandoned:    map[string]bool{},
		Withdrawn:    map[string]bool{},
		Hold:         map[string]string{},
		Decisions:    map[string]*Decision{},
		Snaps:        map[string][]Snap{},
		EverAnswered: map[string]bool{},
		Derived:      map[string]NodeStatus{},
		Aliases:      map[string]string{},
		ResultsByHyp: map[string][]string{},
		EdgesFrom:    map[string][]int{},
		SupersededN:  map[string]int{},
		RivalsAtAdd:  map[string][]string{},
	}
}

// Derived status words.
const (
	HypProposed                = "proposed"
	HypSupported               = "supported"
	HypContradicted            = "contradicted"
	HypInconclusive            = "inconclusive"
	HypSuperseded              = "superseded"
	HypAbandoned               = "abandoned"
	HypAwaitingSubquestions    = "awaiting_subquestions"
	DQOpen                     = "open"
	DQReopened                 = "reopened"
	DQAnswered                 = "answered"
	DQDeferred                 = "deferred"
	DQWithdrawn                = "withdrawn"
	ExpPlanned                 = "planned"
	ExpRun                     = "run"
	ExpSuperseded              = "superseded"
	ExpWithdrawn               = "withdrawn"
	ResRecorded                = "recorded"
	ResRetracted               = "retracted"
	GoalComplete               = "complete"
	GoalInProgress             = "in_progress"
	ReasonConflicting          = "conflicting_selections"
	ReasonSelectedNotSupport   = "selected_not_supported"
	ReasonSelectedStale        = "selected_stale"
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
