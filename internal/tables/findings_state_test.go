package tables

import (
	"testing"

	"github.com/dat9uy/hyp/internal/spec"
)

// A finding's state follows from the Goals that name it.
func TestFindingStateFollowsItsNames(t *testing.T) {
	work := &Goal{ID: "G-WORK00", State: StateOpen}
	landed := &Goal{ID: "G-LAND00", State: StateEnded, Land: &Land{Stop: StopSelected}}
	used := &Goal{ID: "G-USED00", State: StateEnded, Land: &Land{Stop: StopBudgetUsed}}
	closed := &Goal{ID: "G-CLOS00", State: StateEnded, Close: &Close{}}
	for _, c := range []struct {
		names []*Goal
		want  string
	}{
		{nil, spec.FindingNotFolded},
		{[]*Goal{work}, spec.FindingFolded},
		{[]*Goal{landed, work}, spec.FindingFolded},
		{[]*Goal{landed}, spec.FindingSettled},
		{[]*Goal{closed, landed}, spec.FindingSettled},
		{[]*Goal{closed}, spec.FindingEnded},
		{[]*Goal{used, closed}, spec.FindingEnded},
	} {
		f := &Finding{Names: c.names}
		if got := f.State(); got != c.want {
			t.Errorf("named by %v: state %s, want %s", c.names, got, c.want)
		}
		if err := GuardNotFolded(f); (err == nil) != (len(c.names) == 0) {
			t.Errorf("named by %v: Not folded = %v", c.names, err)
		}
	}
	if f := (&Finding{Names: []*Goal{landed}, Refuted: true}); f.State() != spec.FindingRefuted {
		t.Errorf("a refuted finding reads %s", f.State())
	}
}
