package replay

import (
	"slices"

	"github.com/dat9uy/hyp/internal/event"
)

// Limit returns the limit keyed k, on whichever question its Result is.
func (p Projection) Limit(k string) (Limit, bool) {
	for _, s := range p.Decisions {
		for _, l := range s.Limits {
			if l.Key == k {
				return l, true
			}
		}
	}
	return Limit{}, false
}

// UnclosedAims lists the open limits Experiment exp aims at that a Result
// recorded under it with closes and newLimits would leave open. A close
// records that the case ran, whatever the verdicts (ADR 0016), and a case
// the run did not reach is a new limit, so a Result that does neither may
// have run the case and not said so. It is a warning: an instrument failure
// leaves an aimed limit open on purpose.
func (p Projection) UnclosedAims(exp string, closes []string, newLimits int) []string {
	n := p.Nodes[exp]
	if n == nil || n.Kind != event.KindExperiment || newLimits > 0 {
		return nil
	}
	var out []string
	for _, k := range n.Aims {
		if key, ok := event.ParseKey(k); !ok || key.Kind != event.KeyLimit || slices.Contains(closes, k) {
			continue
		}
		if l, ok := p.Limit(k); ok && l.State == LimitOpen {
			out = append(out, k)
		}
	}
	return out
}
