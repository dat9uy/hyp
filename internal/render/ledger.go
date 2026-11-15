package render

import (
	"fmt"
	"slices"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
)

// StopLine is the sentence OWED ends with when the runner owes nothing on
// the Goal: the stop clause a run reads, and the one hyp prompt quotes.
const StopLine = "runner owes nothing: stop and report to the owner"

// owedLine is one ledger item as hyp status prints it: its question, then
// itemText.
func (r renderer) owedLine(o replay.Owed) string {
	return r.id(o.Question) + " " + r.itemText(o)
}

// itemText names one ledger item. hyp status prints it under OWED or
// AWAITING OWNER, and a refused select prints it as what the runner owes
// first, so the two never name an item two ways.
func (r renderer) itemText(o replay.Owed) string {
	p := r.p
	switch o.Kind {
	case replay.OwedRivals:
		if out := r.ruledOut(o.Question); out != "" {
			return "add rival Hypotheses; ruled out: " + out
		}
		return "add rival Hypotheses"
	case replay.OwedUntested:
		return "test " + r.withAlias(o.Key)
	case replay.OwedUndecided:
		return "decide " + r.withAlias(o.Key) + ", inconclusive with no planned Experiment"
	case replay.OwedPremise:
		t := p.Nodes[o.To].Parent
		return "premise " + r.id(o.Key) + " assumes " + r.id(o.To) + ", but " + r.id(t) + " selected " + r.id(p.SelectedHyp(t))
	case replay.OwedPlanned:
		return "run " + r.id(o.Key) + " " + Text(p.Nodes[o.Key].Title)
	case replay.OwedLimit, replay.OwedReading, replay.OwedReferred:
		res := event.KeyResult(o.Key)
		text := ""
		if l, ok := p.Limit(o.Key); ok {
			text = l.Text
		}
		if l, ok := p.Reading(o.Key); ok {
			text = l.Answer
		}
		what := map[replay.OwedKind]string{replay.OwedLimit: "limit", replay.OwedReading: "apply answer", replay.OwedReferred: "limit referred, accept or return"}[o.Kind]
		state := ""
		if o.Kind != replay.OwedReferred {
			state = " [" + o.State
			if len(o.AimedBy) > 0 {
				state += " by " + r.ids(o.AimedBy)
			}
			state += "]"
		}
		return fmt.Sprintf("%s %s%s — %s", what, r.limitLabel(res, o.Key), state, Text(text))
	case replay.OwedChoose:
		return "choose: " + r.named(p.Questions[o.Question].Candidates)
	case replay.OwedSelect:
		return "select " + r.withAlias(o.Key) + ", the only candidate"
	case replay.OwedStale:
		return "stale " + r.withAlias(o.Key) + ": reaffirm it, or re-run what its select rests on"
	case replay.OwedAnswer:
		l, _ := p.Reading(o.Key)
		return "answer reading " + r.limitLabel(l.Result, o.Key) + ": " + Text(l.Text)
	}
	return string(o.Kind)
}

// waitText names one thing question dq's turn waits on. FRONTIER prints it
// under the question, and a refusal prints it after the question, so both
// name the wait in the same words.
func (r renderer) waitText(dq string, w replay.Wait) string {
	p := r.p
	key := r.limitLabel(event.KeyResult(w.Key), w.Key)
	switch w.Kind {
	case replay.WaitRunner:
		return "owes the runner first: " + r.itemText(w.Item)
	case replay.WaitReading:
		if len(w.Reach) == 0 {
			return fmt.Sprintf("waits on open reading %s for the owner; it holds %s's decision", key, r.id(dq))
		}
		return fmt.Sprintf("waits on open reading %s for the owner; its answer can move %s", key, r.named(w.Reach))
	case replay.WaitReferred:
		return fmt.Sprintf("waits on referred limit %s for the owner; an accept releases the select, a return sends its case to run", key)
	case replay.WaitPremise:
		t := p.Nodes[w.Premise.To].Parent
		return fmt.Sprintf("waits on premise %s: it assumes %s, in %s %s; decide %s first", r.withAlias(w.Premise.From), r.withAlias(w.Premise.To), r.id(t), p.Questions[t].State, r.id(t))
	case replay.WaitSubquestion:
		return fmt.Sprintf("waits on a sub-question of %s not answered", r.withAlias(w.Key))
	}
	return "waits on " + string(w.Kind) + " " + w.Key
}

// SelectBlock names why question q owes no select or choice: the first
// thing its turn waits on, or else q's state.
func SelectBlock(p replay.Projection, q string) string {
	r := renderer{p: p}
	return r.selectBlock(q, p.Turn(q))
}

func (r renderer) selectBlock(q string, t replay.Turn) string {
	if t.Kind == replay.TurnWaits {
		return r.withAlias(q) + " " + r.waitText(q, t.Waits[0])
	}
	qs := r.p.Questions[q]
	state := string(qs.State)
	if qs.Reason != "" {
		state += "/" + qs.Reason
	}
	switch {
	case qs.State == replay.DQAnswered:
		return fmt.Sprintf("%s is answered; hyp unselect it first to change the choice", r.withAlias(q))
	case len(qs.Candidates) == 0:
		return fmt.Sprintf("%s has no candidate (%s)", r.withAlias(q), state)
	}
	return fmt.Sprintf("%s owes no select (%s)", r.withAlias(q), state)
}

// RefusalText phrases a refusal that replay decided from the ledger.
func RefusalText(p replay.Projection, f *replay.Refusal) string {
	r := renderer{p: p}
	switch f.Kind {
	case replay.RefuseEnded:
		return f.Ended
	case replay.RefuseSelect:
		switch {
		case len(f.Owed) == 1:
			return fmt.Sprintf("%s owes a select of %s, the only candidate, not %s", r.withAlias(f.Question), r.withAlias(f.Owed[0]), r.withAlias(f.Hyp))
		case f.Choose:
			return fmt.Sprintf("%s is not a candidate of %s; the owner chooses between %s", r.withAlias(f.Hyp), r.withAlias(f.Question), r.named(f.Owed))
		}
		return r.selectBlock(f.Question, f.Turn)
	case replay.RefuseReaffirm:
		return fmt.Sprintf("%s cannot be reaffirmed: %s", r.withAlias(f.Node), r.selectBlock(f.Question, f.Turn))
	case replay.RefuseNextHyp:
		why := "no open limit that a Result on " + r.withAlias(f.Question) + " recorded is unreferred"
		if len(f.Reading.Reach) > 0 {
			why = "its answer can move " + r.named(f.Reading.Reach)
		}
		return fmt.Sprintf("open reading %s holds the next Hypothesis of %s for the owner: %s", r.limitLabel(f.Reading.Result, f.Reading.Key), r.withAlias(f.Question), why)
	case replay.RefuseRetireHyp:
		return fmt.Sprintf("cannot %s %s: open reading %s waits for the owner, and its answer can move %s", f.Verb, r.withAlias(f.Hyp), r.limitLabel(f.Reading.Result, f.Reading.Key), r.withAlias(f.Hyp))
	}
	return string(f.Kind)
}

// premiseGroup is a Hypothesis that candidates rest on, with the
// candidates that assume it.
type premiseGroup struct {
	to string
	by []string
}

// premiseGroups lays out the premises replay derived (Premises) by their
// target, in the order they come.
func premiseGroups(prs []replay.Premise) []premiseGroup {
	var out []premiseGroup
	for _, pr := range prs {
		i := slices.IndexFunc(out, func(g premiseGroup) bool { return g.to == pr.To })
		if i < 0 {
			out = append(out, premiseGroup{to: pr.To})
			i = len(out) - 1
		}
		if !slices.Contains(out[i].by, pr.From) {
			out[i].by = append(out[i].by, pr.From)
		}
	}
	return out
}
