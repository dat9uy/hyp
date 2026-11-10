package replay

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// Each predicate on a status word names one composition the code once
// spelled out at every call site. The two senses of "open" for a question
// differ on deferred: the ledger works an open or reopened question, and a
// deferred sub-question still obliges its parent.
func TestStatusPredicates(t *testing.T) {
	for _, c := range []struct {
		s             QuestionState
		open, obliges bool
	}{
		{DQOpen, true, true},
		{DQReopened, true, true},
		{DQDeferred, false, true},
		{DQAnswered, false, false},
		{DQWithdrawn, false, false},
	} {
		if c.s.Open() != c.open || c.s.Obliges() != c.obliges {
			t.Errorf("%s: Open %v Obliges %v, want %v %v", c.s, c.s.Open(), c.s.Obliges(), c.open, c.obliges)
		}
	}
	for _, c := range []struct {
		s           HypStatus
		live, rival bool
	}{
		{HypProposed, true, true},
		{HypSupported, true, true},
		{HypInconclusive, true, true},
		{HypContradicted, true, false},
		{HypSuperseded, false, false},
		{HypAbandoned, false, false},
	} {
		if LiveHypStatus(c.s) != c.live || RivalStatus(c.s) != c.rival {
			t.Errorf("%s: live %v rival %v, want %v %v", c.s, LiveHypStatus(c.s), RivalStatus(c.s), c.live, c.rival)
		}
	}
	for _, c := range []struct {
		s      ExpStatus
		closed bool
	}{{ExpPlanned, false}, {ExpRun, false}, {ExpSuperseded, true}, {ExpWithdrawn, true}} {
		if (&Exp{Status: c.s}).Closed() != c.closed {
			t.Errorf("%s: Closed %v, want %v", c.s, !c.closed, c.closed)
		}
	}
	if !(Reading{}).Open() || (Reading{By: "c"}).Open() {
		t.Error("a reading is open exactly while no checkpoint closed it")
	}
}

// LiveResults lists the non-retracted Results in merge order.
func TestLiveResults(t *testing.T) {
	b := testutil.New()
	h := b.Hyp(b.DQ(b.Goal("g"), "q"), "h")
	r1 := b.Supports(h)
	r2 := b.Result(b.Exp(h, event.RecipeNone), map[string]string{h: inc})
	r3 := b.Supports(h)
	b.Retract(r2)
	if got := replayB(b).LiveResults(); !slices.Equal(got, []string{r1, r3}) {
		t.Fatalf("LiveResults = %v, want [%s %s]", got, r1, r3)
	}
}
