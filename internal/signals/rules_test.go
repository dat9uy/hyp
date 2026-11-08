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
	if s := r.Rounds[1]; s.OneRival || s.Line != e(x)+" pending; D1 at add: "+h(h2)+"=contradicted" {
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

// A round whose only Result is retracted has nothing to judge, so it is
// pending and never wasted.
func TestRetractedOnlyResultIsPending(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1 := b.Hyp(q, "h1")
	x := b.Exp(h1, event.RecipeNone)
	b.Retract(b.Result(x, map[string]string{h1: event.VerdictSupports}))
	r := compute(t, b)
	if s := r.Rounds[0]; !s.Pending || s.Wasted || len(s.Results) != 0 || s.Line != e(x)+" pending" {
		t.Fatalf("round 1 = %+v", s)
	}
	noSignal(t, r, "wasted:")
}

func r(id string) string { return event.DisplayID(event.KindResult, id) }

// spinRun gives h1 a Result that moves it to supported, then three that do
// not move it.
func spinRun() (b *testutil.B, h1 string, rs []string) {
	b = testutil.New()
	h1 = b.Hyp(b.DQ(b.Goal("g"), "q"), "h1")
	for range 4 {
		rs = append(rs, b.Supports(h1))
	}
	return b, h1, rs
}

func TestSpinOnThreeResultsThatDoNotMove(t *testing.T) {
	b, h1, rs := spinRun()
	want := "spin " + h(h1) + " supported→supported at " + r(rs[1]) + ", supported→supported at " + r(rs[2]) + ", supported→supported at " + r(rs[3])
	if got := detail(t, compute(t, b), "spin:"+h(h1)+":"+r(rs[3])); got != want {
		t.Fatalf("got %q\nwant %q", got, want)
	}
}

// A retracted Result does not count toward spin, as it does not count toward
// wasted: the Results either side of it are not three in a row.
func TestRetractedResultDoesNotCountTowardSpin(t *testing.T) {
	b, _, rs := spinRun()
	b.Retract(rs[2])
	noSignal(t, compute(t, b), "spin:")
}

// Spin reads each Result's statuses from when it was recorded, as wasted does,
// so retracting the Result that moved does not make the later ones move.
func TestRetractingTheMovingResultKeepsSpin(t *testing.T) {
	b, h1, rs := spinRun()
	b.Retract(rs[0])
	detail(t, compute(t, b), "spin:"+h(h1)+":"+r(rs[3]))
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

func TestUnjudgedSignal(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	ex := b.Exp(h1, event.RecipeNone, h2)
	if rep := Compute(b.Evs); len(rep.Rounds[0].Unjudged) != 0 || len(rep.Signals) != 0 {
		t.Fatalf("pending round: unjudged %v, signals %+v; want none", rep.Rounds[0].Unjudged, rep.Signals)
	}
	b.Result(ex, map[string]string{h1: event.VerdictSupports})
	rep := Compute(b.Evs)
	if got := strings.Join(rep.Rounds[0].Unjudged, ","); got != h(h2) {
		t.Fatalf("round unjudged = %q, want %q", got, h(h2))
	}
	want := Signal{Key: "unjudged:" + e(ex), Round: 1, Detail: "unjudged " + e(ex) + " targets " + h(h2) + ", which no Result judged"}
	found := false
	for _, s := range rep.Signals {
		found = found || s == want
	}
	if !found {
		t.Fatalf("signals = %+v, want to contain %+v", rep.Signals, want)
	}
}

// A closed round is neither pending nor wasted, and keeps its add-time
// one-rival and D1 marks.
func TestClosedRoundIsNotPendingOrWasted(t *testing.T) {
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
	byExp := map[string]Round{}
	for _, x := range r.Rounds {
		byExp[x.Exp] = x
	}
	for id, want := range map[string]string{old: "superseded by " + e(repl), gone: "withdrawn"} {
		x := byExp[e(id)]
		if x.Pending || x.Wasted || x.Closed != want || !strings.HasPrefix(x.Line, e(id)+" "+want) {
			t.Errorf("%s: pending %v wasted %v closed %q line %q, want closed %q", e(id), x.Pending, x.Wasted, x.Closed, x.Line, want)
		}
	}
	if !byExp[e(repl)].Pending {
		t.Errorf("open %s: want pending", e(repl))
	}
	detail(t, r, "D1:"+e(old))
	detail(t, r, "one-rival:"+e(gone))
	noSignal(t, r, "wasted:")
}
