package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// RefusalKind is why a write that reads the ledger is refused. Replay
// decides the refusal and stores its facts; package render phrases it, so
// the refusal and hyp status name a fact in the same words.
type RefusalKind string

// The values of RefusalKind.
const (
	RefuseEnded     RefusalKind = "ended"      // the node leads to an ended Goal; Ended holds the refusal text
	RefuseSelect    RefusalKind = "select"     // the ledger owes no select of Hyp on Question
	RefuseReaffirm  RefusalKind = "reaffirm"   // an open reading holds Question's decision, so Node is not reaffirmed
	RefuseNextHyp   RefusalKind = "next_hyp"   // an open reading holds Question's next Hypothesis
	RefuseRetireHyp RefusalKind = "retire_hyp" // an open reading's answer can move Hyp, so it is not superseded or abandoned
)

// Refusal is one refused write, with the facts that decide it.
type Refusal struct {
	Kind     RefusalKind
	Question string
	Hyp      string   // RefuseSelect: the Hypothesis named; RefuseRetireHyp: the one to retire
	Node     string   // RefuseReaffirm: the node to reaffirm
	Verb     string   // RefuseRetireHyp: "supersede" or "abandon"
	Owed     []string // RefuseSelect: the Hypotheses a select may name now (SelectOwed)
	Choose   bool     // RefuseSelect: the ledger owes the owner's choice
	Turn     Turn     // RefuseSelect, RefuseReaffirm: Question's turn
	Reading  Reading  // RefuseNextHyp, RefuseRetireHyp: the reading that holds it
	Ended    string   // RefuseEnded: the end's refusal text (EndedRefusal)
}

// SelectRefusal is the refusal of a select of Hypothesis h on question q,
// or nil when the ledger owes it: a select of h, or a choice on q with h a
// candidate. h must be a resolved id. It reads derived state, so Validate
// does not repeat it: replay still applies a select in a merged log.
func (p Projection) SelectRefusal(q, h string) *Refusal {
	if r := p.endedRefusal(q); r != nil {
		return r
	}
	owed, choose := p.SelectOwed(q)
	if slices.Contains(owed, h) {
		return nil
	}
	return &Refusal{Kind: RefuseSelect, Question: q, Hyp: h, Owed: owed, Choose: choose, Turn: p.Turn(q)}
}

// ReaffirmRefusal is the refusal of a reaffirm of node id while an open
// reading holds the decision of its question: id itself when it is a
// question, else its parent; nil when it is admitted. Replay still applies
// a reaffirm in a merged log.
func (p Projection) ReaffirmRefusal(id string) *Refusal {
	n := p.Nodes[id]
	if n == nil {
		return nil
	}
	if r := p.endedRefusal(id); r != nil {
		return r
	}
	q := id
	if n.Kind == event.KindHypothesis {
		q = n.Parent
	}
	if qs := p.Questions[q]; qs == nil || !qs.State.Open() || len(p.openReadings(q)) == 0 {
		return nil
	}
	return &Refusal{Kind: RefuseReaffirm, Question: q, Node: id, Turn: p.Turn(q)}
}

// AddHypRefusal is the refusal of a new Hypothesis under question q while
// an open reading holds q's next Hypothesis (NextHold), unless the ledger
// owes a premise item on q, which a new Hypothesis can settle; nil when it
// is admitted. Replay still applies such an add in a merged log.
func (p Projection) AddHypRefusal(q string) *Refusal {
	if r := p.endedRefusal(q); r != nil {
		return r
	}
	l, ok := p.NextHold(q)
	if !ok || p.owesPremise(q, "") {
		return nil
	}
	return &Refusal{Kind: RefuseNextHyp, Question: q, Reading: l}
}

// RetireRefusal is the refusal to supersede or abandon (verb) Hypothesis h
// while an open reading's answer can move it, unless h is the candidate a
// premise item on its question names; nil when it is admitted. Replay
// still applies such an event in a merged log.
func (p Projection) RetireRefusal(verb, h string) *Refusal {
	if r := p.endedRefusal(h); r != nil {
		return r
	}
	l, ok := p.MovingReading(h)
	if !ok || p.owesPremise(p.Nodes[h].Parent, h) {
		return nil
	}
	return &Refusal{Kind: RefuseRetireHyp, Question: p.Nodes[h].Parent, Hyp: h, Verb: verb, Reading: l}
}

// endedRefusal is admitEnded as a Refusal.
func (p Projection) endedRefusal(id string) *Refusal {
	if fs := p.admitEnded(id); len(fs) > 0 {
		return &Refusal{Kind: RefuseEnded, Ended: fs[0].Text}
	}
	return nil
}
