package replay

import "slices"

// TurnKind is where a question's decision stands (spec §8.1): the one value
// the ledger, a refused select and hyp status read, never stored.
type TurnKind string

// The values of TurnKind, in the order the rule tries them.
const (
	TurnClosed      TurnKind = "closed"       // answered, deferred or withdrawn
	TurnWaits       TurnKind = "waits"        // something must happen before a select or a choice
	TurnNoCandidate TurnKind = "no_candidate" // open with nothing to select and nothing to wait on; the ledger's rules always owe the runner work there, so it is a guard
	TurnOwed        TurnKind = "owed"         // the runner's select of a lone candidate, or the owner's choice
)

// WaitKind is one thing an open question's select or choice waits on.
type WaitKind string

// The values of WaitKind, in the order a Turn lists them.
const (
	WaitRunner      WaitKind = "runner"      // the runner owes an item on the question
	WaitReading     WaitKind = "reading"     // an open reading whose answer can move a Hypothesis there
	WaitReferred    WaitKind = "referred"    // a limit the runner referred, not yet accepted or returned
	WaitPremise     WaitKind = "premise"     // a candidate assumes a Hypothesis whose question is open
	WaitSubquestion WaitKind = "subquestion" // the selected or lone candidate has a sub-question not answered
)

// Wait is one thing a question waits on. Key is the runner item's key, the
// reading or limit key, the premise's candidate, or the Hypothesis whose
// sub-question is open.
type Wait struct {
	Kind    WaitKind
	Key     string
	Item    Owed     // WaitRunner: the first item the runner owes
	Reach   []string // WaitReading: the Hypotheses its answer can move
	Premise Premise  // WaitPremise
}

// Turn is a question's turn: its kind, everything it waits on when it
// waits, and its candidates when a select or a choice is owed (one
// candidate: the runner selects it; two or more: the owner chooses).
type Turn struct {
	Kind       TurnKind
	Waits      []Wait
	Candidates []string
}

// Turn derives question dq's turn from the ledger's other items on it.
func (p Projection) Turn(dq string) Turn {
	var items []Owed
	for _, o := range p.Owed() {
		if o.Question == dq && o.Kind != OwedSelect && o.Kind != OwedChoose {
			items = append(items, o)
		}
	}
	return p.turnOf(dq, items)
}

// turnOf is the turn rule, first match wins: closed unless the question is
// open or reopened; waits when anything below holds; no candidate; owed.
// items are the ledger's items on dq other than its select or choice. A
// select stands exactly when the question is answered, so a reopened
// question with a select never owes a select of it again: it waits.
func (p Projection) turnOf(dq string, items []Owed) Turn {
	q := p.Questions[dq]
	if q == nil || !q.State.Open() {
		return Turn{Kind: TurnClosed}
	}
	var ws []Wait
	if i := slices.IndexFunc(items, func(o Owed) bool { return o.Who == OwedRunner }); i >= 0 {
		ws = append(ws, Wait{Kind: WaitRunner, Key: items[i].Key, Item: items[i]})
	}
	for _, l := range p.openReadings(dq) {
		ws = append(ws, Wait{Kind: WaitReading, Key: l.Key, Reach: l.Reach})
	}
	if k := p.ReferredLimit(dq); k != "" {
		ws = append(ws, Wait{Kind: WaitReferred, Key: k})
	}
	if pr := p.OpenPremise(dq); pr != nil {
		ws = append(ws, Wait{Kind: WaitPremise, Key: pr.From, Premise: *pr})
	}
	if h := p.subquestionWait(dq); h != "" {
		ws = append(ws, Wait{Kind: WaitSubquestion, Key: h})
	}
	switch {
	case len(ws) > 0:
		return Turn{Kind: TurnWaits, Waits: ws}
	case len(q.Candidates) == 0:
		return Turn{Kind: TurnNoCandidate}
	}
	return Turn{Kind: TurnOwed, Candidates: q.Candidates}
}

// subquestionWait is the Hypothesis whose open sub-question question dq
// waits for (ADR 0014), or "": the current select's when the question
// reopened awaiting its sub-questions, or else the lone candidate's. A
// choice between two or more candidates does not wait for them.
func (p Projection) subquestionWait(dq string) string {
	q := p.Questions[dq]
	if rec, ok := p.CurrentSelect(dq); ok && q.Reason == ReasonAwaitingSubquestions {
		return rec.Hyp
	}
	if len(q.Candidates) == 1 && p.HasOpenSubquestion(q.Candidates[0]) {
		return q.Candidates[0]
	}
	return ""
}

// openReadings lists the open readings on question dq whose answer can move
// a Hypothesis there: each makes the select and choice wait, and stops new
// rivals being owed, since the answer may revive one.
func (p Projection) openReadings(dq string) []Reading {
	var out []Reading
	for _, l := range p.Readings {
		if l.Question == dq && l.Open() && len(l.Reach) > 0 {
			out = append(out, l)
		}
	}
	return out
}
