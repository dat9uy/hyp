package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Who owes an item of the ledger.
const (
	OwedRunner = "runner"
	OwedOwner  = "owner"
)

// What an owed item is.
const (
	OwedRivals    = "rivals"    // an open question has no live rival and no open reading that can revive one
	OwedUntested  = "untested"  // a live Hypothesis has no Result and no Experiment that would give one
	OwedUndecided = "undecided" // a live Hypothesis is inconclusive and no planned Experiment would decide it
	OwedPlanned   = "planned"   // a planned Experiment has no Result
	OwedLimit     = "limit"     // an open limit under a question that has a candidate, or a line of the Goal a candidate has not met
	OwedReading   = "reading"   // a closed reading no Result applies
	OwedReferred  = "referred"  // a limit referred to the owner, not yet accepted
	OwedPremise   = "premise"   // a candidate assumes a Hypothesis its question did not select
	OwedStale     = "stale"     // a reopened question's select is stale and nothing else owed renews it
	OwedSelect    = "select"    // a lone candidate with nothing else owed on its question
	OwedChoose    = "choose"    // the owner chooses between two or more candidates
)

// How far an aimable item (a limit or a reading) has got.
const (
	AimNone      = "owed"              // no live Experiment aims at it
	AimPlanned   = "in_progress"       // an aiming Experiment has no Result yet
	AimUnsettled = "aimed_not_settled" // every aiming Experiment has a Result, and none settled it
)

// Owed is one item of the ledger: open work that someone owes on a question.
// Settling an item is always an explicit event; owing one never changes a
// status.
type Owed struct {
	Who, Kind, Question string
	Key                 string   // a node id, or a limit or reading key
	State               string   // for a limit or reading: AimNone, AimPlanned or AimUnsettled
	AimedBy             []string // live Experiments that aim at the key
	To                  string   // for a premise: the depends_on target Key assumes
}

// Owed derives the ledger in question creation order. An open or reopened
// question owes every kind of item; an answered one owes only the limit,
// reading, referred and premise items, so evidence that lands after the
// answer is owed and not only listed. An open reading is owed by the owner
// too; hyp status prints it under AWAITING OWNER from its own record, so it
// is not repeated here.
func (p Projection) Owed() []Owed {
	aimers := map[string][]string{}
	for _, ev := range p.Applied {
		n := p.Nodes[ev.ID]
		if ev.Type != event.TypeNodeAdd || n == nil || n.Kind != event.KindExperiment || closedExp(p.Exps[ev.ID]) {
			continue
		}
		for _, k := range n.Aims {
			aimers[k] = append(aimers[k], ev.ID)
		}
	}
	aim := func(o Owed) Owed {
		o.AimedBy = aimers[o.Key]
		o.State = AimNone
		if len(o.AimedBy) > 0 {
			o.State = AimUnsettled
			for _, e := range o.AimedBy {
				if p.Exps[e].Status == ExpPlanned {
					o.State = AimPlanned
				}
			}
		}
		return o
	}

	var out []Owed
	for _, ev := range p.Applied {
		dq := ev.ID
		n := p.Nodes[dq]
		if ev.Type != event.TypeNodeAdd || n == nil || n.Kind != event.KindDesignQuestion {
			continue
		}
		q := p.Questions[dq]
		if q == nil {
			continue
		}
		open := q.State.Open()
		if !open && q.State != DQAnswered {
			continue
		}
		var mine []Owed
		add := func(o Owed) { o.Question = dq; mine = append(mine, o) }

		// An open reading makes the question wait only while its answer can
		// move a Hypothesis there (its Reach).
		waiting := len(p.openReadings(dq)) > 0
		if open {
			var live []string
			for _, h := range n.Children {
				if hs := p.Hyps[h]; hs != nil && hs.Status != HypSuperseded && hs.Status != HypAbandoned {
					live = append(live, h)
				}
			}
			// With no live rival, new rivals are owed, unless an open reading
			// whose answer may revive one waits for the owner.
			rivals := slices.ContainsFunc(live, func(h string) bool { return p.Hyps[h].Status != HypContradicted })
			if !rivals && !waiting {
				add(Owed{Who: OwedRunner, Kind: OwedRivals, Key: dq})
			}
			for _, h := range live {
				switch {
				case p.Hyps[h].Status == HypProposed && !p.hasLiveExp(h):
					add(Owed{Who: OwedRunner, Kind: OwedUntested, Key: h})
				case p.Hyps[h].Status == HypInconclusive && !p.hasPlannedExp(h):
					add(Owed{Who: OwedRunner, Kind: OwedUndecided, Key: h})
				}
				for _, e := range p.Nodes[h].Children {
					if p.Nodes[e].Kind == event.KindExperiment && p.Exps[e].Status == ExpPlanned {
						add(Owed{Who: OwedRunner, Kind: OwedPlanned, Key: e})
					}
				}
			}
			// A measuring Experiment sits under the question itself.
			for _, e := range n.Children {
				if p.Nodes[e].Kind == event.KindExperiment && p.Exps[e].Status == ExpPlanned {
					add(Owed{Who: OwedRunner, Kind: OwedPlanned, Key: e})
				}
			}
		}
		for _, l := range p.LimitsOf(dq) {
			switch {
			case !l.Open():
			case l.Referred():
				add(Owed{Who: OwedOwner, Kind: OwedReferred, Key: l.Key})
			case len(q.Candidates) > 0:
				// A limit names a case no check ran; it may apply to every
				// candidate, whichever Hypotheses its Result judged.
				add(aim(Owed{Who: OwedRunner, Kind: OwedLimit, Key: l.Key}))
			}
		}
		// A line of the Goal is owed on each live top-level question with a
		// candidate no closer covers; a referred one is the owner's there.
		if n.Parent == p.Goal.GoalID {
			for _, l := range p.Lines {
				switch {
				case !l.Open():
				case l.Referred():
					add(Owed{Who: OwedOwner, Kind: OwedReferred, Key: l.Key})
				case len(p.Uncovered(l, dq)) > 0:
					add(aim(Owed{Who: OwedRunner, Kind: OwedLimit, Key: l.Key}))
				}
			}
		}
		for _, l := range p.Readings {
			if l.Question == dq && l.By != "" && l.Applied == "" {
				add(aim(Owed{Who: OwedRunner, Kind: OwedReading, Key: l.Key}))
			}
		}
		for _, pr := range p.premisesOf(dq) {
			if rec, ok := p.CurrentSelect(p.Nodes[pr.To].Parent); ok && rec.Hyp != pr.To {
				add(Owed{Who: OwedRunner, Kind: OwedPremise, Key: pr.From, To: pr.To})
			}
		}
		// A select whose basis changed reopens its question as stale. When
		// no other runner item and no premise's open question accounts for
		// it, the runner reaffirms the stale node or re-runs what it rests
		// on; a select of the same Hypothesis is never owed again.
		if open && q.Reason == ReasonSelectedStale && !RunnerOwes(mine) && p.OpenPremise(dq) == nil {
			add(Owed{Who: OwedRunner, Kind: OwedStale, Key: p.staleNode(dq)})
		}
		// A select or a choice is owed only when the question's turn is
		// owed (turn.go): nothing it waits on, and a candidate.
		if t := p.turnOf(dq, mine); t.Kind == TurnOwed {
			if len(t.Candidates) == 1 {
				add(Owed{Who: OwedRunner, Kind: OwedSelect, Key: t.Candidates[0]})
			} else {
				add(Owed{Who: OwedOwner, Kind: OwedChoose, Key: dq})
			}
		}
		out = append(out, mine...)
	}
	return out
}

// staleNode is the node a stale select of question dq rests on: the
// selected Hypothesis when it is stale, otherwise the question itself.
func (p Projection) staleNode(dq string) string {
	if rec, ok := p.CurrentSelect(dq); ok && p.Hyps[rec.Hyp] != nil && p.Hyps[rec.Hyp].Stale != nil {
		return rec.Hyp
	}
	return dq
}

// Premise is a live depends_on edge from a candidate of a question to a
// Hypothesis under another question.
type Premise struct {
	From, To string
}

// premisesOf lists the premises of question dq's candidates, in candidate
// and edge order. An edge to a Hypothesis under dq itself is no premise:
// the choice between them is dq's own.
func (p Projection) premisesOf(dq string) []Premise {
	var out []Premise
	for _, c := range p.Questions[dq].Candidates {
		for _, e := range p.Edges {
			if !e.Live || e.Type != event.EdgeDependsOn || e.From != c {
				continue
			}
			if t := p.Nodes[e.To]; t != nil && t.Kind == event.KindHypothesis && t.Parent != dq {
				out = append(out, Premise{From: c, To: e.To})
			}
		}
	}
	return out
}

// PremiseCycles lists each pair of questions, in question creation order,
// whose candidates assume each other's Hypotheses. Neither owes a select
// while the other is open, so each holds the other until an edge goes.
func (p Projection) PremiseCycles() [][2]string {
	var qs []string
	for _, ev := range p.Applied {
		if n := p.Nodes[ev.ID]; ev.Type == event.TypeNodeAdd && n != nil && n.Kind == event.KindDesignQuestion {
			qs = append(qs, ev.ID)
		}
	}
	assumes := func(a, b string) bool {
		return slices.ContainsFunc(p.premisesOf(a), func(pr Premise) bool { return p.Nodes[pr.To].Parent == b })
	}
	var out [][2]string
	for i, a := range qs {
		for _, b := range qs[i+1:] {
			if assumes(a, b) && assumes(b, a) {
				out = append(out, [2]string{a, b})
			}
		}
	}
	return out
}

// OpenPremise returns the first premise of question dq whose target's
// question is open or reopened, or nil. While it stands, dq's turn waits:
// the premise's question is decided first.
func (p Projection) OpenPremise(dq string) *Premise {
	for _, pr := range p.premisesOf(dq) {
		if t := p.Questions[p.Nodes[pr.To].Parent]; t != nil && t.State.Open() {
			return &pr
		}
	}
	return nil
}

// RunnerOwes reports whether the ledger holds an item the runner owes.
func RunnerOwes(items []Owed) bool {
	return slices.ContainsFunc(items, func(o Owed) bool { return o.Who == OwedRunner })
}

// hasLiveExp reports whether a planned or run Experiment targets h.
func (p Projection) hasLiveExp(h string) bool {
	for id, e := range p.Exps {
		n := p.Nodes[id]
		if closedExp(e) || n == nil {
			continue
		}
		if slices.Contains(e.Targets, h) {
			return true
		}
	}
	return false
}

// hasPlannedExp reports whether a planned Experiment targets h.
func (p Projection) hasPlannedExp(h string) bool {
	for id, e := range p.Exps {
		n := p.Nodes[id]
		if closedExp(e) || n == nil || e.Status != ExpPlanned {
			continue
		}
		if slices.Contains(e.Targets, h) {
			return true
		}
	}
	return false
}

// closedExp reports whether e is missing or closed (Exp.Closed).
func closedExp(e *Exp) bool { return e == nil || e.Closed() }

// ReferredLimit is the key of the first open limit under question dq, or
// line of the Goal when dq is a top-level question, whose latest hand is a
// refer, or "". The question's select and choice wait for it until the
// owner accepts or returns it.
func (p Projection) ReferredLimit(dq string) string {
	ls := p.LimitsOf(dq)
	if n := p.Nodes[dq]; n != nil && n.Parent == p.Goal.GoalID {
		ls = append(slices.Clone(ls), p.Lines...)
	}
	for _, l := range ls {
		if l.Open() && l.Referred() {
			return l.Key
		}
	}
	return ""
}
