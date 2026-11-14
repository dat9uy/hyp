package replay

import "github.com/dat9uy/hyp/internal/event"

// Unseparated lists each pair of rivals under question dq that no
// non-retracted Result separates: none gives the two different verdicts.
// Two rivals are separated when some Result judged both and gave them
// different words (glossary, Rival). A pair is listed in child order.
func (p Projection) Unseparated(dq string) [][2]string {
	var rivals []string
	for _, h := range p.Nodes[dq].Children {
		if p.Nodes[h].Kind == event.KindHypothesis && p.Hyps[h] != nil && RivalStatus(p.Hyps[h].Status) {
			rivals = append(rivals, h)
		}
	}
	var out [][2]string
	for i, a := range rivals {
		for _, b := range rivals[i+1:] {
			if !p.separates(a, b) {
				out = append(out, [2]string{a, b})
			}
		}
	}
	return out
}

// separates reports whether a non-retracted Result gives a and b different
// verdicts.
func (p Projection) separates(a, b string) bool {
	for _, res := range p.LiveResults() {
		v := p.Nodes[res].Verdicts
		va, oka := v[a]
		vb, okb := v[b]
		if oka && okb && va != vb {
			return true
		}
	}
	return false
}

// LiveResults lists the non-retracted Results in merge order: the Results
// whose verdicts count.
func (p Projection) LiveResults() []string {
	var out []string
	for _, ev := range p.Applied {
		if ev.Type == event.TypeResultRecord && p.Results[ev.ID].Status != ResRetracted {
			out = append(out, ev.ID)
		}
	}
	return out
}
