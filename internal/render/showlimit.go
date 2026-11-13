package render

import (
	"fmt"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// ShowLimit renders one limit for hyp show R-…#n: its text and state, whose
// turn it is while open, what its Result judged and the question's
// candidates now, the Experiments that aim at it, each closer with its
// verdicts, the candidates no closer judged, and the checkpoints that
// referred, returned or accepted it with their summaries. Every line is a
// stored fact or derived from them; a limit's category is never stored
// (ADR 0016).
func ShowLimit(p replay.Projection, l replay.Limit) string {
	r := renderer{p: p}
	var b strings.Builder
	w := func(format string, a ...any) { fmt.Fprintf(&b, format+"\n", a...) }
	label := r.limitLabel(l.Result, l.Key)
	if l.IsLine() {
		showLine(r, w, l, label)
		r.limitHistory(w, l)
		return b.String()
	}

	w("Limit %s", label)
	w("text: %s", l.Text)
	w("result: %s under %s", r.id(l.Result), r.withAlias(p.Nodes[l.Result].Parent))
	w("question: %s", r.id(l.Question))
	state := r.limitState(l)
	if l.Open() {
		switch {
		case l.Referred():
			state += ", referred: the owner accepts or returns it"
		case l.Returned():
			state += ", returned: the runner runs its case"
		default:
			state += ": the runner runs its case"
		}
	}
	w("state: %s", state)

	n := p.Nodes[l.Result]
	var judged []string
	for _, h := range n.VerdictOrder {
		judged = append(judged, fmt.Sprintf("%s %s, now %s", r.withAlias(h), n.Verdicts[h], p.Hyps[h].Status))
	}
	w("judged: %s", strings.Join(judged, "; "))
	cands := p.Questions[l.Question].Candidates
	if len(cands) == 0 {
		w("candidates: none")
	} else {
		w("candidates: %s", r.named(cands))
	}
	if l.Closed() {
		var unjudged []string
		for _, c := range cands {
			if !slices.ContainsFunc(l.Closers, func(id string) bool { _, ok := p.Nodes[id].Verdicts[c]; return ok }) {
				unjudged = append(unjudged, c)
			}
		}
		if len(unjudged) > 0 {
			w("not judged by a closer: %s", r.named(unjudged))
		}
	}

	var aimed []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeNodeAdd && slices.Contains(p.Nodes[ev.Node].Aims, l.Key) {
			aimed = append(aimed, r.id(ev.Node)+" "+string(p.Exps[ev.Node].Status))
		}
	}
	if len(aimed) > 0 {
		w("aimed by: %s", strings.Join(aimed, ", "))
	}

	r.limitHistory(w, l)
	return b.String()
}

// showLine renders a line of the Goal: its text, its hand when it has one,
// and for each live top-level question the candidates and whether a closer
// covers them all, then what each closer ran on. A line is never open or
// closed; the ledger owes it on a question with an uncovered candidate.
func showLine(r renderer, w func(string, ...any), l replay.Limit, label string) {
	p := r.p
	w("Line %s", label)
	w("text: %s", l.Text)
	if h := r.handWord(l); h != "" {
		w("hand: %s", h)
	}
	for _, q := range p.TopQuestions(l.Result) {
		cands := p.Questions[q].Candidates
		w("question %s: candidates %s", r.withAlias(q), map[bool]string{true: "none", false: r.named(cands)}[len(cands) == 0])
		switch un := p.Uncovered(l, q); {
		case len(un) > 0:
			w("  uncovered: %s", r.named(un))
		case len(cands) > 0:
			w("  covered")
		}
	}
	for _, c := range l.Closers {
		var hyps []string
		for _, q := range p.TopQuestions(l.Result) {
			for _, h := range p.Nodes[q].Children {
				if p.Covers(c, h) {
					hyps = append(hyps, h)
				}
			}
		}
		w("ran in %s: %s", r.id(c), r.ranOn(c, hyps))
	}
	for _, c := range l.Cites {
		now := ""
		if c.Now != c.Meaning {
			now = ", now " + c.Now
		}
		w("cites %s: %s%s", r.citeLabel(c.Key), c.Meaning, now)
	}
	if sel := p.NotTogether(l); len(sel) > 0 {
		w("(no one closer covers %s together)", strings.ReplaceAll(r.ids(sel), ", ", " and "))
	}
	var aimed []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeNodeAdd && slices.Contains(p.Nodes[ev.Node].Aims, l.Key) {
			aimed = append(aimed, r.id(ev.Node)+" "+string(p.Exps[ev.Node].Status))
		}
	}
	if len(aimed) > 0 {
		w("aimed by: %s", strings.Join(aimed, ", "))
	}
}

// limitHistory prints the events that recorded, closed, referred, returned
// or accepted limit l, oldest first.
func (r renderer) limitHistory(w func(string, ...any), l replay.Limit) {
	p := r.p
	w("history:")
	for _, ev := range p.Applied {
		var what string
		switch ev.Type {
		case event.TypeCheckpoint:
			var c event.Checkpoint
			_ = event.DecodePayload(ev, &c)
			for _, it := range c.Refers {
				if it.Key == l.Key {
					what = "refers: " + Cut(c.Summary, event.MaxSummary)
					if it.Text != "" {
						what = "refers: " + Cut(it.Text, event.MaxText)
					}
				}
			}
			for _, f := range []struct {
				keys []string
				verb string
			}{{c.Returns, "returns"}, {c.Accepts, "accepts"}} {
				if slices.Contains(f.keys, l.Key) {
					what = f.verb + ": " + Cut(c.Summary, event.MaxText)
				}
			}
			if l.IsLine() && len(c.Lines) > 0 && ev.Node == l.Result {
				what = "writes it"
			}
		case event.TypeAnswer:
			var a event.Answer
			_ = event.DecodePayload(ev, &a)
			for _, f := range []struct {
				items []event.KeyText
				verb  string
			}{{a.Returns, "returns"}, {a.Accepts, "accepts"}} {
				for _, it := range f.items {
					if it.Key == l.Key {
						what = f.verb + ": " + Cut(it.Text, event.MaxText)
					}
				}
			}
		case event.TypeNodeAdd:
			var a event.NodeAdd
			_ = event.DecodePayload(ev, &a)
			if l.IsLine() && ev.ID == l.Result && len(a.Lines) > 0 {
				what = "writes it"
			}
		case event.TypeConfirm:
			var c event.Confirm
			_ = event.DecodePayload(ev, &c)
			if slices.Contains(c.Accepts, l.Key) {
				what = "accepts: " + Cut(ev.Why, event.MaxText)
			}
		case event.TypeResultRecord:
			if ev.ID == l.Result {
				what = "records it"
			} else if slices.Contains(p.Nodes[ev.ID].Closes, l.Key) {
				what = "closes it: " + r.verdicts(p.Nodes[ev.ID].Verdicts)
			}
		}
		if what != "" {
			w("  %s %s %s — %s", Minute(ev.TS), ev.Actor, ev.Type, what)
		}
	}
}

// citeLabel shows a line's cite: a Hypothesis with its alias, or a key.
func (r renderer) citeLabel(c string) string {
	if n := r.p.Nodes[c]; n != nil {
		return r.withAlias(c)
	}
	return r.limitLabel(event.KeyResult(c), c)
}
