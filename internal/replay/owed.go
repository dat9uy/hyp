package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Who owes an item of the ledger.
const (
	OwedRunner   = "runner"
	OwedOwner    = "owner"
	OwedVerifier = "verifier"
)

// What an owed item is.
const (
	OwedRivals   = "rivals"   // an open question has no live rival and no open reading
	OwedUntested = "untested" // a live Hypothesis has no Result and no Experiment that would give one
	OwedPlanned  = "planned"  // a planned Experiment has no Result
	OwedLimit    = "limit"    // an open limit on a candidate
	OwedReading  = "reading"  // a closed reading no Result applies
	OwedReferred = "referred" // a limit referred to the owner, not yet accepted
	OwedSelect   = "select"   // a lone candidate with nothing else owed on its question
	OwedConfirm  = "confirm"  // a select waits for the verifier
	OwedRespond  = "respond"  // the verifier found a select insufficient
	OwedChoose   = "choose"   // the owner chooses between two or more candidates
)

// How far an aimable item (a limit or a reading) has got.
const (
	AimNone      = "owed"              // no live Experiment aims at it
	AimPlanned   = "in_progress"       // an aiming Experiment has no Result yet
	AimUnsettled = "aimed_not_settled" // every aiming Experiment has a Result, and none settled it
)

// Owed is one item of the ledger: open work that someone owes on an open or
// reopened question. Settling an item is always an explicit event; owing one
// never changes a status.
type Owed struct {
	Who, Kind, Question string
	Key                 string   // a node id, or a limit or reading key
	State               string   // for a limit or reading: AimNone, AimPlanned or AimUnsettled
	AimedBy             []string // live Experiments that aim at the key
}

// Owed derives the ledger in question creation order. An open reading and an
// owner confirm are owed by the owner too; hyp status prints those under
// AWAITING OWNER from their own records, so they are not repeated here.
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
		if q == nil || (q.State != DQOpen && q.State != DQReopened) {
			continue
		}
		var mine []Owed
		add := func(o Owed) { o.Question = dq; mine = append(mine, o) }

		var live []string
		for _, h := range n.Children {
			if hs := p.Hyps[h]; hs != nil && hs.Status != HypSuperseded && hs.Status != HypAbandoned {
				live = append(live, h)
			}
		}
		// With no live rival, new rivals are owed, unless an open reading on
		// the question's Results waits for the owner, whose answer may revive one.
		rivals := slices.ContainsFunc(live, func(h string) bool { return p.Hyps[h].Status != HypContradicted })
		waiting := slices.ContainsFunc(p.Readings, func(l Reading) bool { return l.Question == dq && l.By == "" })
		if !rivals && !waiting {
			add(Owed{Who: OwedRunner, Kind: OwedRivals, Key: dq})
		}
		for _, h := range live {
			if p.Hyps[h].Status == HypProposed && !p.hasLiveExp(h) {
				add(Owed{Who: OwedRunner, Kind: OwedUntested, Key: h})
			}
			for _, e := range p.Nodes[h].Children {
				if p.Nodes[e].Kind == event.KindExperiment && p.Exps[e].Status == ExpPlanned {
					add(Owed{Who: OwedRunner, Kind: OwedPlanned, Key: e})
				}
			}
		}
		for _, l := range p.LimitsOf(dq) {
			if l.State != LimitOpen || !overlaps(l.Hyps, q.Candidates) {
				continue
			}
			if l.Referred != "" {
				add(Owed{Who: OwedOwner, Kind: OwedReferred, Key: l.Key})
				continue
			}
			add(aim(Owed{Who: OwedRunner, Kind: OwedLimit, Key: l.Key}))
		}
		for _, l := range p.Readings {
			if l.Question == dq && l.By != "" && l.Applied == "" {
				add(aim(Owed{Who: OwedRunner, Kind: OwedReading, Key: l.Key}))
			}
		}
		// A select counts only while it names a candidate: once its Hypothesis
		// is contradicted, the question needs a select or a choice again.
		rec, selected := p.CurrentSelect(dq)
		selected = selected && slices.Contains(q.Candidates, rec.Hyp)
		if len(q.Candidates) == 1 && !selected && !slices.ContainsFunc(mine, func(o Owed) bool { return o.Who == OwedRunner }) {
			add(Owed{Who: OwedRunner, Kind: OwedSelect, Key: q.Candidates[0]})
		}
		// A trade-off is a state: two or more candidates. Choosing between them
		// is the owner's turn only once the runner owes nothing on the question
		// and no reading on it waits for the owner, since either may change
		// the candidates.
		if len(q.Candidates) >= 2 && !selected && !waiting && !slices.ContainsFunc(mine, func(o Owed) bool { return o.Who == OwedRunner }) {
			add(Owed{Who: OwedOwner, Kind: OwedChoose, Key: dq})
		}
		switch q.Reason {
		case ReasonAwaitingVerifier:
			add(Owed{Who: OwedVerifier, Kind: OwedConfirm, Key: dq})
		case ReasonVerifierInsufficient:
			add(Owed{Who: OwedRunner, Kind: OwedRespond, Key: dq})
		}
		out = append(out, mine...)
	}
	return out
}

// RunnerOwes reports whether the ledger holds an item the runner owes.
func RunnerOwes(items []Owed) bool {
	return slices.ContainsFunc(items, func(o Owed) bool { return o.Who == OwedRunner })
}

// hasLiveExp reports whether a planned or run Experiment has h as its parent
// or a target.
func (p Projection) hasLiveExp(h string) bool {
	for id, e := range p.Exps {
		n := p.Nodes[id]
		if closedExp(e) || n == nil {
			continue
		}
		if n.Parent == h || slices.Contains(n.Targets, h) {
			return true
		}
	}
	return false
}

func closedExp(e *Exp) bool {
	return e == nil || e.Closed != "" || e.Status == ExpSuperseded || e.Status == ExpWithdrawn
}

func overlaps(a, b []string) bool {
	return slices.ContainsFunc(a, func(x string) bool { return slices.Contains(b, x) })
}
