package example

import (
	"bytes"

	"github.com/dat9uy/hyp/internal/herr"
)

// Verdict is the result of a compare.
type Verdict string

// The two verdicts.
const (
	Met    Verdict = "met"
	Failed Verdict = "failed"
)

// Compare judges output, the stdout of the run with one final newline
// dropped, against ex: met when every contains floor is a substring, every
// equals floor equals the output and no ceiling is a substring; otherwise
// failed. The exit code and the stderr are never compared.
func Compare(output []byte, ex Example) Verdict {
	for _, f := range ex.Floors {
		switch f.Op {
		case OpContains:
			if !bytes.Contains(output, []byte(f.Text)) {
				return Failed
			}
		case OpEquals:
			if !bytes.Equal(output, []byte(f.Text)) {
				return Failed
			}
		default:
			return Failed
		}
	}
	for _, c := range ex.Ceilings {
		if bytes.Contains(output, []byte(c)) {
			return Failed
		}
	}
	return Met
}

// AddsOnly refuses a new version of an example that changes more than it
// adds floors or ceilings. oldProbe and newProbe are the tree hashes of the
// two versions' probe/ folders. The refusal names the first section or path
// that differs; such a change is a new example key, not a new version.
func AddsOnly(old, next Example, oldProbe, newProbe string) error {
	differ := func(what string) error {
		return herr.Refused("Adds only: %s differs from the current version; a change there is a new example key, not a new version", what)
	}
	switch {
	case oldProbe != newProbe:
		return differ("probe/")
	case old.Words != next.Words:
		return differ("the words section")
	case old.Run != next.Run:
		return differ("the run section")
	}
	oldIn := map[string]string{}
	for _, in := range old.Inputs {
		oldIn[in.Path] = in.Body
	}
	newIn := map[string]string{}
	for _, in := range next.Inputs {
		newIn[in.Path] = in.Body
	}
	for _, list := range [][]Input{old.Inputs, next.Inputs} {
		for _, in := range list {
			o, inOld := oldIn[in.Path]
			n, inNew := newIn[in.Path]
			if inOld != inNew || o != n {
				return differ("input/" + in.Path)
			}
		}
	}
	if old.HasExternal != next.HasExternal || old.External != next.External {
		return differ("the external section")
	}
	oldFloors := map[Floor]int{}
	for _, f := range next.Floors {
		oldFloors[f]++
	}
	for _, f := range old.Floors {
		if oldFloors[f]--; oldFloors[f] < 0 {
			return herr.Refused("Adds only: the new version drops floor: %s %q", f.Op, f.Text)
		}
	}
	ceilings := map[string]int{}
	for _, c := range next.Ceilings {
		ceilings[c]++
	}
	for _, c := range old.Ceilings {
		if ceilings[c]--; ceilings[c] < 0 {
			return herr.Refused("Adds only: the new version drops ceiling %q", c)
		}
	}
	if len(next.Floors)+len(next.Ceilings) <= len(old.Floors)+len(old.Ceilings) {
		return herr.Refused("Adds only: the new version adds nothing; it must add at least one floor or ceiling")
	}
	return nil
}
