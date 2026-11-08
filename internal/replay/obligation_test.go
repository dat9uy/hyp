package replay

import (
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// obligation is a Goal with a question q and rivals h1 (supported) and h2
// (proposed), the start of every case below.
type obligation struct {
	b      *testutil.B
	g, q   string
	h1, h2 string
}

func newObligation() obligation {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	b.Supports(h1)
	return obligation{b: b, g: g, q: q, h1: h1, h2: h2}
}

// answered adds a sub-question of h with one Hypothesis that answers it.
func (o obligation) answered(h string) (sub, ans, sel string) {
	sub = o.b.DQ(h, "sub")
	ans = o.b.Hyp(sub, "ans")
	o.b.Supports(ans)
	sel = o.b.Answer(sub, ans, "")
	return sub, ans, sel
}

// A sub-question must be settled, whichever way, before its parent can be the
// answer (spec §7.4). The block is the question's, so the parent's status and
// everything reading it stay as its verdicts leave them.
func TestSubquestionIsAnObligation(t *testing.T) {
	cases := []struct {
		name string
		run  func(t *testing.T, o obligation)
	}{
		{"answered parent gains a sub-question", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")
			o.b.DQ(o.h1, "sub")
			p := replayB(o.b)
			wantStatus(t, p, o.q, DQReopened, ReasonAwaitingSubquestions)
			wantStatus(t, p, o.h1, HypSupported, "")
		}},
		{"selected parent has an open sub-question", func(t *testing.T, o obligation) {
			o.b.DQ(o.h1, "sub")
			o.b.Select(o.q, o.h1, "", false)
			p := replayB(o.b)
			wantStatus(t, p, o.q, DQOpen, ReasonAwaitingSubquestions)
			wantStatus(t, p, o.h1, HypSupported, "")
		}},
		{"unselected supported parent has an open sub-question", func(t *testing.T, o obligation) {
			o.b.DQ(o.h1, "sub")
			p := replayB(o.b)
			wantStatus(t, p, o.q, DQOpen, ReasonSupportedUnsel)
			wantStatus(t, p, o.h1, HypSupported, "")
			if got := p.Questions[o.q].Candidates; !slices.Equal(got, []string{o.h1}) {
				t.Fatalf("candidates = %v, want [%s]", got, o.h1)
			}
		}},
		{"a sub-question on a target leaves its dependent answered", func(t *testing.T, o obligation) {
			q2 := o.b.DQ(o.g, "q2")
			d := o.b.Hyp(q2, "d")
			o.b.Edge(d, event.EdgeDependsOn, o.h1)
			o.b.Supports(d)
			o.b.Answer(q2, d, "")
			o.b.DQ(o.h1, "sub")
			p := replayB(o.b)
			wantStatus(t, p, q2, DQAnswered, "")
			wantStale(t, p, d, nil)
		}},
		{"any answer discharges it, and a flip changes nothing on the parent", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			sub, ans, sel := o.answered(o.h1)
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")

			ans2 := o.b.Hyp(sub, "ans2")
			o.b.Contradicts(ans)
			o.b.Supports(ans2)
			wantStatus(t, replayB(o.b), o.q, DQReopened, ReasonAwaitingSubquestions)

			o.b.Answer(sub, ans2, sel)
			p := replayB(o.b)
			wantStatus(t, p, sub, DQAnswered, "")
			wantStatus(t, p, o.q, DQAnswered, "")
		}},
		{"withdrawing it releases the parent", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			sub := o.b.DQ(o.h1, "sub")
			wantStatus(t, replayB(o.b), o.q, DQReopened, ReasonAwaitingSubquestions)
			o.b.Withdraw(sub)
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")
		}},
		{"a premise goes through depends_on and reaffirm restores the answer", func(t *testing.T, o obligation) {
			sub, ans, sel := o.answered(o.h1)
			o.b.Edge(o.h1, event.EdgeDependsOn, ans)
			o.b.Reaffirm(o.h1)
			o.b.Answer(o.q, o.h1, "")
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")

			ans2 := o.b.Hyp(sub, "ans2")
			o.b.Contradicts(ans)
			o.b.Supports(ans2)
			p := replayB(o.b)
			wantStatus(t, p, o.q, DQReopened, ReasonSelectedStale)
			if p.Hyps[o.h1].Stale == nil {
				t.Fatalf("h1 should be stale: its premise fell")
			}

			o.b.Answer(sub, ans2, sel)
			o.b.Unedge(o.h1, event.EdgeDependsOn, ans)
			o.b.Edge(o.h1, event.EdgeDependsOn, ans2)
			o.b.Reaffirm(o.h1)
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")
		}},
		{"a basis taken while the target had an open sub-question stays good", func(t *testing.T, o obligation) {
			q2 := o.b.DQ(o.g, "q2")
			d := o.b.Hyp(q2, "d")
			o.b.Edge(d, event.EdgeDependsOn, o.h1)
			sub := o.b.DQ(o.h1, "sub")
			o.b.Supports(d)
			wantStale(t, replayB(o.b), d, nil)
			ans := o.b.Hyp(sub, "ans")
			o.b.Supports(ans)
			o.b.Answer(sub, ans, "")
			wantStale(t, replayB(o.b), d, nil)
		}},
		{"a deferred sub-question still blocks", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			o.b.Defer(o.b.DQ(o.h1, "sub"))
			wantStatus(t, replayB(o.b), o.q, DQReopened, ReasonAwaitingSubquestions)
		}},
		{"a sub-question under an unselected rival does not block", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			o.b.DQ(o.h2, "sub")
			wantStatus(t, replayB(o.b), o.q, DQAnswered, "")
		}},
		{"the Goal is not complete while the obligation is open", func(t *testing.T, o obligation) {
			o.b.Answer(o.q, o.h1, "")
			if got := replayB(o.b).Goal.Status; got != GoalComplete {
				t.Fatalf("Goal is %s before the sub-question", got)
			}
			sub := o.b.DQ(o.h1, "sub")
			if got := replayB(o.b).Goal.Status; got != GoalInProgress {
				t.Fatalf("Goal is %s with an open obligation", got)
			}
			o.b.Withdraw(sub)
			if got := replayB(o.b).Goal.Status; got != GoalComplete {
				t.Fatalf("Goal is %s after the release", got)
			}
		}},
		{"a sub-question's Hypothesis may depend on the parent", func(t *testing.T, o obligation) {
			sub := o.b.DQ(o.h1, "sub")
			h3 := o.b.Hyp(sub, "h3")
			o.b.Edge(h3, event.EdgeDependsOn, o.h1)
			o.b.Edge(sub, event.EdgeDependsOn, o.h1)
			wantNoRejections(t, replayB(o.b))
		}},
		{"depends_on between Hypotheses still refuses a cycle", func(t *testing.T, o obligation) {
			sub := o.b.DQ(o.h1, "sub")
			h3 := o.b.Hyp(sub, "h3")
			o.b.Edge(h3, event.EdgeDependsOn, o.h1)
			closing := o.b.Edge(o.h1, event.EdgeDependsOn, h3)
			p := replayB(o.b)
			if len(p.Rejections) != 1 || p.Rejections[0].EventID != closing || !strings.Contains(p.Rejections[0].Reason, "depends_on cycle") {
				t.Fatalf("rejections = %+v, want the closing edge %s refused as a depends_on cycle", p.Rejections, closing)
			}
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) { c.run(t, newObligation()) })
	}
}

// HasOpenSubquestion is what the suggestion reads: not withdrawn and not
// answered, so a deferred sub-question counts.
func TestHasOpenSubquestion(t *testing.T) {
	o := newObligation()
	if replayB(o.b).HasOpenSubquestion(o.h1) {
		t.Fatal("no sub-question yet")
	}
	sub := o.b.DQ(o.h1, "sub")
	if !replayB(o.b).HasOpenSubquestion(o.h1) {
		t.Fatal("open sub-question not seen")
	}
	o.b.Defer(sub)
	if !replayB(o.b).HasOpenSubquestion(o.h1) {
		t.Fatal("deferred sub-question not seen")
	}
	o.b.Withdraw(sub)
	if replayB(o.b).HasOpenSubquestion(o.h1) {
		t.Fatal("withdrawn sub-question still seen")
	}
}
