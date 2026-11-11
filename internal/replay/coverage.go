package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Covers reports whether Result r ran on Hypothesis h: it has a verdict on h,
// or one of its run subjects holds h's whole subject (HoldsSubject). Coverage
// says the bytes ran, not that the claim held; only a verdict says that.
func (p Projection) Covers(r, h string) bool {
	n := p.Nodes[r]
	if n == nil {
		return false
	}
	if _, ok := n.Verdicts[h]; ok {
		return true
	}
	return HoldsSubject(n.Subjects, p.Subjects[h])
}

// Uncovered lists the candidates of question q that no closer of limit l
// covers, in candidate order.
func (p Projection) Uncovered(l Limit, q string) []string {
	var out []string
	for _, c := range p.Questions[q].Candidates {
		if !slices.ContainsFunc(l.Closers, func(r string) bool { return p.Covers(r, c) }) {
			out = append(out, c)
		}
	}
	return out
}

// TopQuestions lists the live DesignQuestions whose parent is Goal g, in
// creation order: the questions a line of g reaches.
func (p Projection) TopQuestions(g string) []string {
	var out []string
	for _, ev := range p.Applied {
		n := p.Nodes[ev.ID]
		if ev.Type != event.TypeNodeAdd || n == nil || n.Kind != event.KindDesignQuestion || n.Parent != g {
			continue
		}
		if q := p.Questions[ev.ID]; q != nil && q.State != DQWithdrawn {
			out = append(out, ev.ID)
		}
	}
	return out
}

// NotTogether lists the Hypotheses selected now on the live top-level
// questions of line l's Goal when two or more are selected and no one closer of line l
// covers them all; it is nil otherwise. It is a mark: a claim that depends
// on its host is covered only together with that host, so the owner reads
// it at choose. No item and no status reads it.
func (p Projection) NotTogether(l Limit) []string {
	var sel []string
	for _, q := range p.TopQuestions(l.Result) {
		if rec, ok := p.CurrentSelect(q); ok {
			sel = append(sel, rec.Hyp)
		}
	}
	if len(sel) < 2 {
		return nil
	}
	for _, r := range l.Closers {
		if !slices.ContainsFunc(sel, func(h string) bool { return !p.Covers(r, h) }) {
			return nil
		}
	}
	return sel
}
