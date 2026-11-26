package example

import (
	"bytes"
	"regexp"
	"slices"
	"strings"

	"github.com/dat9uy/hyp/internal/fault"
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

// ExitLine is the last line a Probe prints, "exit N". hyp cuts it off the
// Probe's stdout before it compares, so the output never ends with it.
var ExitLine = regexp.MustCompile(`^exit (0|[1-9][0-9]*)$`)

// Meetable refuses an example that no output can meet: a floor whose whole
// text is an exit line, two equals floors that differ, an equals floor that
// does not hold the text of a contains floor, or a floor that holds the text
// of a ceiling. It names the first such floor.
func Meetable(ex Example) error {
	for _, f := range ex.Floors {
		if ExitLine.MatchString(f.Text) {
			return fault.Refuse(fault.GuardMeetable, fault.FloorIsExitLine, f.Op, f.Text)
		}
	}
	var equals *Floor
	for i, f := range ex.Floors {
		if f.Op != OpEquals {
			continue
		}
		if equals == nil {
			equals = &ex.Floors[i]
		} else if equals.Text != f.Text {
			return fault.Refuse(fault.GuardMeetable, fault.EqualsFloorsDiffer, equals.Text, f.Text)
		}
	}
	for _, f := range ex.Floors {
		if equals != nil && f.Op == OpContains && !strings.Contains(equals.Text, f.Text) {
			return fault.Refuse(fault.GuardMeetable, fault.EqualsLacksContains, equals.Text, f.Text)
		}
		for _, c := range ex.Ceilings {
			if strings.Contains(f.Text, c) {
				return fault.Refuse(fault.GuardMeetable, fault.FloorHoldsCeiling, f.Op, f.Text, c)
			}
		}
	}
	return nil
}

// AddsOnly refuses a new version of an example that changes more than it
// adds floors or ceilings. oldProbe and newProbe are the tree hashes of the
// two versions' probe/ folders. The refusal names the first section or path
// that differs; such a change is a new example key, not a new version.
func AddsOnly(old, next Example, oldProbe, newProbe string) error {
	section := func(name string) error { return fault.Refuse(fault.GuardAddsOnly, fault.SectionDiffers, name) }
	switch {
	case oldProbe != newProbe:
		return fault.Refuse(fault.GuardAddsOnly, fault.PathDiffers, "probe/")
	case old.Words != next.Words:
		return section("words")
	case old.Run != next.Run:
		return section("run")
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
				return section("input/" + in.Path)
			}
		}
	}
	if old.HasExternal != next.HasExternal || old.External != next.External {
		return section("external")
	}
	if !slices.Equal(old.Findings, next.Findings) {
		return section("findings")
	}
	oldFloors := map[Floor]int{}
	for _, f := range next.Floors {
		oldFloors[f]++
	}
	for _, f := range old.Floors {
		if oldFloors[f]--; oldFloors[f] < 0 {
			return fault.Refuse(fault.GuardAddsOnly, fault.DropsFloor, f.Op, f.Text)
		}
	}
	ceilings := map[string]int{}
	for _, c := range next.Ceilings {
		ceilings[c]++
	}
	for _, c := range old.Ceilings {
		if ceilings[c]--; ceilings[c] < 0 {
			return fault.Refuse(fault.GuardAddsOnly, fault.DropsCeiling, c)
		}
	}
	if len(next.Floors)+len(next.Ceilings) <= len(old.Floors)+len(old.Ceilings) {
		return fault.Refuse(fault.GuardAddsOnly, fault.AddsNothing)
	}
	return nil
}
