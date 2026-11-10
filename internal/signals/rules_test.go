package signals

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/testutil"
)

func h(id string) string { return event.DisplayID(event.KindHypothesis, id) }
func e(id string) string { return event.DisplayID(event.KindExperiment, id) }

func detail(t *testing.T, r Report, key string) string {
	t.Helper()
	for _, s := range r.Signals {
		if s.Key == key {
			return s.Detail
		}
	}
	t.Fatalf("no signal %s in %v", key, r.Signals)
	return ""
}

// compute fails the test if any event is refused: Compute drops refused
// events, so a refused setup would pass a noSignal check for the wrong reason.
func compute(t *testing.T, b *testutil.B) Report {
	t.Helper()
	if _, p := replay.Build(b.Evs, nil, ""); len(p.Rejections) != 0 {
		t.Fatalf("setup refused: %v", p.Rejections)
	}
	return Compute(b.Evs)
}

func noSignal(t *testing.T, r Report, prefix string) {
	t.Helper()
	for _, s := range r.Signals {
		if strings.HasPrefix(s.Key, prefix) {
			t.Fatalf("unexpected signal %s: %s", s.Key, s.Detail)
		}
	}
}

// An Experiment with no target while a live rival exists is one-rival, which
// is what `hyp add exp --target none` stores.
func TestOneRivalWithNoTarget(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	x := b.Exp(h1, event.RecipeNone)
	want := "one-rival " + e(x) + " at add: " + h(h1) + "=proposed (parent), " + h(h2) + "=proposed; targets: none; could target: " + h(h2)
	if got := detail(t, Compute(b.Evs), "one-rival:"+e(x)); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

func TestTargetingALiveRivalIsNotOneRival(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Exp(h1, event.RecipeNone, h2)
	noSignal(t, Compute(b.Evs), "one-rival:")
}

func TestD1WhenATargetIsDead(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Contradicts(h2)
	x := b.Exp(h1, event.RecipeNone, h2)
	r := Compute(b.Evs)
	if got, want := detail(t, r, "D1:"+e(x)), "D1 "+e(x)+" at add: "+h(h2)+"=contradicted"; got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
	if s := r.Rounds[1]; s.OneRival {
		t.Fatalf("round 2 = %+v", s)
	}
}

// The verifier-loop signal was retired with the verifier role (ADR 0015):
// old insufficient verifier confirms in a row are history and fire nothing.
func TestOldVerifierVerdictsFireNoSignal(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	sel := b.As("a").Select(q, h1, "", false)
	b.As("v1").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	b.As("v2").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	noSignal(t, compute(t, b), "verifier-loop:")
}

// A retracted Result is not listed in its round.
func TestRetractedResultIsNotListed(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1 := b.Hyp(q, "h1")
	x := b.Exp(h1, event.RecipeNone)
	b.Retract(b.Result(x, map[string]string{h1: event.VerdictSupports}))
	if s := compute(t, b).Rounds[0]; len(s.Results) != 0 {
		t.Fatalf("round 1 = %+v", s)
	}
}

// Progress is the ledger's (ADR 0017): Results that move nothing fire no
// signal, however many there are.
func TestResultsThatMoveNothingFireNoSignal(t *testing.T) {
	b := testutil.New()
	h1 := b.Hyp(b.DQ(b.Goal("g"), "q"), "h1")
	for range 4 {
		b.Supports(h1)
	}
	if r := compute(t, b); len(r.Signals) != 0 {
		t.Fatalf("signals = %+v, want none", r.Signals)
	}
}

func TestEmptyLogHasEmptyLists(t *testing.T) {
	got, err := json.Marshal(Compute(nil))
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != `{"rounds":[],"signals":[]}` {
		t.Fatalf("got %s", got)
	}
}

// A closed round keeps its add-time one-rival and D1 marks.
func TestClosedRoundKeepsItsMarks(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3")
	b.Contradicts(h3)
	old := b.Exp(h1, event.RecipeNone, h3)
	repl := b.Exp(h1, event.RecipeNone, h2)
	gone := b.Exp(h2, event.RecipeNone)
	b.Edge(repl, event.EdgeSupersedes, old)
	b.Emit(event.TypeExperimentWithdraw, gone, event.Empty{})
	r := compute(t, b)
	detail(t, r, "D1:"+e(old))
	detail(t, r, "one-rival:"+e(gone))
}
