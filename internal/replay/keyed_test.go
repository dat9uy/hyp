package replay

import (
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// A refer of a Result's limit that a Result closed is refused: its case
// ran. Retracting the closer opens the limit again, and a refer applies.
func TestReferOfAClosedLimit(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	h := b.Hyp(b.DQ(g, "q"), "h")
	r1 := limitResult(b, h, []string{"l1"}, nil)
	k := event.LimitKey(r1, 1)
	closer := limitResult(b, h, nil, []string{k})
	late := b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "no test", Refers: []string{k}})
	p := replayB(b)
	wantRejected(t, p, late, "its case ran, so there is nothing to refer")
	b.Retract(closer) // its Experiment is planned again
	b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "no test", Refers: []string{k}})
	p = replayB(b)
	if len(p.Rejections) != 1 {
		t.Fatalf("rejections = %+v, want only the refer before the retract", p.Rejections)
	}
	if l, _ := p.Limit(k); !l.Open() || !l.Referred() {
		t.Fatalf("limit = %+v, want open and referred", l)
	}
}

// A line of the Goal is never closed, so a runner refers it after a Result
// ran its case.
func TestReferOfALineThatRan(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	b.As("o").Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "lines", Role: event.RoleOwner, Lines: []string{"one tenant per row"}})
	q := b.DQ(g, "q")
	h := b.As("a").Hyp(q, "h")
	k := event.LimitKey(g, 1)
	limitResult(b, h, nil, []string{k})
	b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "no test", Refers: []string{k}})
	wantNoRejections(t, replayB(b))
}
