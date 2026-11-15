package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// The facts below are the ones hyp status and hyp show print beside the
// ledger. Each is derived once here, by the rule the ledger and the
// refusals read, so a renderer only lays it out and two commands never
// answer the same fact by two rules.

// OwedOf is the ledger's items on the questions of Goal g, in the order
// Owed gives them.
func (p Projection) OwedOf(g string) []Owed {
	var out []Owed
	for _, o := range p.Owed() {
		if p.GoalOf(o.Question) == g {
			out = append(out, o)
		}
	}
	return out
}

// UnseparatedCandidates lists each pair of question dq's candidates that no
// non-retracted Result separates, by the rule Unseparated reads: no Result
// judged both and gave them different verdicts. A pair is listed in
// candidate order.
func (p Projection) UnseparatedCandidates(dq string) [][2]string {
	cands := p.Questions[dq].Candidates
	var out [][2]string
	for i, a := range cands {
		for _, b := range cands[i+1:] {
			if !p.separates(a, b) {
				out = append(out, [2]string{a, b})
			}
		}
	}
	return out
}

// Premises lists the premises of question dq's candidates, in candidate and
// edge order: the depends_on edges the ledger owes a premise item on. An
// edge to a Hypothesis under dq itself is no premise.
func (p Projection) Premises(dq string) []Premise { return p.premisesOf(dq) }

// ContradictedBy lists the non-retracted Results that give Hypothesis h
// contradicts, in merge order.
func (p Projection) ContradictedBy(h string) []string {
	var out []string
	for _, res := range p.LiveResults() {
		if p.Nodes[res].Verdicts[h] == event.VerdictContradicts {
			out = append(out, res)
		}
	}
	return out
}

// JudgedBy lists the non-retracted Results that give Hypothesis h a
// verdict, in merge order.
func (p Projection) JudgedBy(h string) []string {
	var out []string
	for _, res := range p.LiveResults() {
		if _, ok := p.Nodes[res].Verdicts[h]; ok {
			out = append(out, res)
		}
	}
	return out
}

// Predecessors lists the Hypotheses that h supersedes, directly or through
// another, nearest first, by live supersedes edges.
func (p Projection) Predecessors(h string) []string {
	var out []string
	seen := map[string]bool{h: true}
	edges := p.LiveEdges()
	for queue := []string{h}; len(queue) > 0; queue = queue[1:] {
		for _, e := range edges {
			if e.From == queue[0] && e.Type == event.EdgeSupersedes && !seen[e.To] {
				seen[e.To] = true
				out = append(out, e.To)
				queue = append(queue, e.To)
			}
		}
	}
	return out
}

// Supersedes lists the Hypotheses h directly supersedes, by live edges.
func (p Projection) Supersedes(h string) []string {
	var out []string
	for _, e := range p.LiveEdges() {
		if e.From == h && e.Type == event.EdgeSupersedes {
			out = append(out, e.To)
		}
	}
	return out
}

// PredecessorResults lists the non-retracted Results that judged a
// predecessor of h (Predecessors), in merge order.
func (p Projection) PredecessorResults(h string) []string {
	preds := p.Predecessors(h)
	var out []string
	for _, res := range p.LiveResults() {
		if slices.ContainsFunc(preds, func(pr string) bool { _, ok := p.Nodes[res].Verdicts[pr]; return ok }) {
			out = append(out, res)
		}
	}
	return out
}

// AimedBy lists every Experiment that aims at key k, whatever its status,
// in creation order. The ledger's AimedBy counts only the live ones.
func (p Projection) AimedBy(k string) []string {
	var out []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeNodeAdd && p.Nodes[ev.ID] != nil && slices.Contains(p.Nodes[ev.ID].Aims, k) {
			out = append(out, ev.ID)
		}
	}
	return out
}

// NotJudgedByCloser lists the candidates of limit l's question that none of
// its closers gave a verdict.
func (p Projection) NotJudgedByCloser(l Limit) []string {
	var out []string
	for _, c := range p.Questions[l.Question].Candidates {
		if !slices.ContainsFunc(l.Closers, func(id string) bool { _, ok := p.Nodes[id].Verdicts[c]; return ok }) {
			out = append(out, c)
		}
	}
	return out
}

// ResultsOf lists Goal g's Results that match keep, in merge order.
func (p Projection) ResultsOf(g string, keep func(*Res) bool) []string {
	var out []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeResultRecord && p.GoalOf(ev.ID) == g && p.Results[ev.ID] != nil && keep(p.Results[ev.ID]) {
			out = append(out, ev.ID)
		}
	}
	return out
}

// Questionable lists Goal g's Results whose recipe changed since they were
// recorded.
func (p Projection) Questionable(g string) []string {
	return p.ResultsOf(g, func(r *Res) bool { return r.Questionable })
}

// Uncontained lists Goal g's Results with uncontained evidence.
func (p Projection) Uncontained(g string) []string {
	return p.ResultsOf(g, func(r *Res) bool { return len(r.Uncontained) > 0 })
}

// ChosenCandidate is the Hypothesis of question dq's current select when
// that select records the owner's choice: one of two or more candidates;
// "" otherwise.
func (p Projection) ChosenCandidate(dq string) string {
	if rec, ok := p.CurrentSelect(dq); ok && p.TradeOff(dq, rec.Hyp) {
		return rec.Hyp
	}
	return ""
}

// TradeOff reports whether Hypothesis h is one of two or more candidates of
// question dq, so a select of it is the owner's choice.
func (p Projection) TradeOff(dq, h string) bool {
	cands := p.Questions[dq].Candidates
	return len(cands) >= 2 && slices.Contains(cands, h)
}

// SelectedHyp is the Hypothesis of question dq's current select, or "".
func (p Projection) SelectedHyp(dq string) string {
	rec, _ := p.CurrentSelect(dq)
	return rec.Hyp
}
