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

	w("Limit %s", label)
	w("text: %s", l.Text)
	w("result: %s under %s", r.id(l.Result), r.withAlias(p.Nodes[l.Result].Parent))
	w("question: %s", r.id(l.Question))
	state := l.State
	switch l.State {
	case replay.LimitOpen:
		switch {
		case l.Referred != "":
			state += ", referred: the owner accepts or returns it"
		case l.Returned != "":
			state += ", returned: the runner runs its case"
		default:
			state += ": the runner runs its case"
		}
	case replay.LimitClosed:
		state += " by " + r.closers(l.Closers)
	case replay.LimitAccepted:
		state += " by " + l.By
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
	if l.State == replay.LimitClosed {
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

	w("history:")
	for _, ev := range p.Applied {
		var what string
		switch ev.Type {
		case event.TypeCheckpoint:
			var c event.Checkpoint
			_ = event.DecodePayload(ev, &c)
			for _, f := range []struct {
				keys []string
				verb string
			}{{c.Refers, "refers"}, {c.Returns, "returns"}, {c.Accepts, "accepts"}} {
				if slices.Contains(f.keys, l.Key) {
					what = f.verb + ": " + Cut(c.Summary, event.MaxText)
				}
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
	return b.String()
}
