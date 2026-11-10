package replay

import "github.com/dat9uy/hyp/internal/event"

// Exp is the derived record of one Experiment (ADR 0008). Replay builds it
// once per projection, and every reader takes an Experiment's state from it
// instead of re-deriving it from the Experiment's Results.
type Exp struct {
	Status     ExpStatus         // ExpPlanned, ExpRun, ExpSuperseded or ExpWithdrawn
	Closed     ExpStatus         // ExpSuperseded or ExpWithdrawn when it is closed, even if a Result makes it run; "" otherwise
	Live       []string          // non-retracted Results, in creation order
	Judged     map[string]string // Hypothesis → verdict over Live; a later Result wins
	OneRival   bool              // the one-rival rule (oneRival) on its add-time rivals
	ReplacedBy []string          // Experiments with a live supersedes edge to it, in edge order
	// The pre-registration (preregistration.go): Targets are the
	// Hypotheses each Result gives a verdict on, none under a
	// DesignQuestion, and Shape is derived from the parent kind, the
	// targets and the rubric.
	Targets []string
	Shape   string
}

// expClosed is how Experiment id is closed: ExpSuperseded when a live
// supersedes edge points at it, otherwise ExpWithdrawn when it is withdrawn,
// otherwise "".
func (st *State) expClosed(id string) ExpStatus {
	switch {
	case st.superseded(id):
		return ExpSuperseded
	case st.Withdrawn[id]:
		return ExpWithdrawn
	}
	return ""
}

// expStatus is the status rule of Experiment id (spec §7.6), first match
// wins: run with a non-retracted Result, so a live Result beats a close;
// then superseded, withdrawn, and planned.
func (st *State) expStatus(id string) ExpStatus {
	if st.liveResult(id) != "" {
		return ExpRun
	}
	if c := st.expClosed(id); c != "" {
		return c
	}
	return ExpPlanned
}

// exps builds the Exp record of every Experiment from the stored state.
func (st *State) exps() map[string]*Exp {
	out := map[string]*Exp{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindExperiment {
			continue
		}
		e := &Exp{Status: st.expStatus(id), Closed: st.expClosed(id), Judged: map[string]string{}}
		if e.Closed == ExpSuperseded {
			e.ReplacedBy = st.Superseders(id)
		}
		for _, r := range n.Children {
			if st.Retracted[r] {
				continue
			}
			e.Live = append(e.Live, r)
			for h, v := range st.Nodes[r].Verdicts {
				e.Judged[h] = v
			}
		}
		e.OneRival = oneRival(st.RivalsAtAdd[id], n.Targets)
		e.Targets, e.Shape = targetsOf(st.Nodes, n), shapeOf(st.Nodes, n)
		out[id] = e
	}
	return out
}
