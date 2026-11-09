package replay

import "strings"

// SameAsk lists the keys of the readings that ask the owner the same question
// as reading k, k first: the same `of` and the same text, spaces trimmed.
// Each copy holds its own Result's question (§8.1), so a meaning that reaches
// two questions has a key on each; the owner answers them as one. A key that
// names no reading lists nothing.
func (p Projection) SameAsk(k string) []string {
	var base *Reading
	for i := range p.Readings {
		if p.Readings[i].Key == k {
			base = &p.Readings[i]
		}
	}
	if base == nil {
		return nil
	}
	out := []string{k}
	for _, l := range p.Readings {
		if l.Key != k && sameAsk(*base, l) {
			out = append(out, l.Key)
		}
	}
	return out
}

func sameAsk(a, b Reading) bool {
	return a.Of == b.Of && strings.TrimSpace(a.Text) == strings.TrimSpace(b.Text)
}

// Reading returns reading k, or false when no reading has that key.
func (p Projection) Reading(k string) (Reading, bool) {
	for _, l := range p.Readings {
		if l.Key == k {
			return l, true
		}
	}
	return Reading{}, false
}

// KeyLabel shows reading or limit key k as R-W6V06D#r1.
func (p Projection) KeyLabel(k string) string { return p.keyLabel(k) }
