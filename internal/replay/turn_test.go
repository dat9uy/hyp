package replay

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// waitKinds lists what question q's turn waits on.
func waitKinds(t Turn) []WaitKind {
	var out []WaitKind
	for _, w := range t.Waits {
		out = append(out, w.Kind)
	}
	return out
}

func wantTurn(t *testing.T, p Projection, q string, kind TurnKind, waits ...WaitKind) {
	t.Helper()
	got := p.Turn(q)
	if got.Kind != kind || !slices.Equal(waitKinds(got), waits) {
		t.Fatalf("turn of %s = %s %v, want %s %v", q, got.Kind, waitKinds(got), kind, waits)
	}
}

// A question with no candidate waits on runner work, a question with one is
// owed its select, an answered or deferred one is closed.
func TestTurnFollowsTheQuestion(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	e := b.Exp(h, event.RecipeNone)
	wantTurn(t, replayB(b), q, TurnWaits, WaitRunner)
	b.Result(e, map[string]string{h: con})
	b.Hyp(q, "h2")
	b.Emit(event.TypeHypothesisAbandon, h, event.Empty{})
	p := replayB(b)
	if p.Turn(q).Waits[0].Item.Kind != OwedUntested {
		t.Fatalf("turn = %+v, want it to wait on the untested rival", p.Turn(q))
	}

	b2 := testutil.New()
	q2 := b2.DQ(b2.Goal("g"), "q")
	h2 := b2.Hyp(q2, "h")
	b2.Supports(h2)
	p = replayB(b2)
	wantTurn(t, p, q2, TurnOwed)
	if got := p.Turn(q2).Candidates; !slices.Equal(got, []string{h2}) {
		t.Fatalf("candidates = %v, want [%s]", got, h2)
	}
	b2.Answer(q2, h2, "")
	wantTurn(t, replayB(b2), q2, TurnClosed)

	b3 := testutil.New()
	q3 := b3.DQ(b3.Goal("g"), "q")
	b3.Defer(q3)
	wantTurn(t, replayB(b3), q3, TurnClosed)
}

// A lone candidate's turn waits on its premise's open question and on its
// own open sub-question, in that order, and is owed once both are answered.
func TestTurnWaitsOnPremiseAndSubquestion(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q, q2 := b.DQ(g, "q"), b.DQ(g, "q2")
	h, dep := b.Hyp(q, "h"), b.Hyp(q2, "dep")
	b.Supports(h)
	b.Supports(dep)
	b.Edge(h, event.EdgeDependsOn, dep)
	b.Supports(h) // a fresh basis with dep supported
	sub := b.DQ(h, "sub")
	sh := b.Hyp(sub, "sh")
	b.Supports(sh)
	p := replayB(b)
	wantTurn(t, p, q, TurnWaits, WaitPremise, WaitSubquestion)
	wantRefused(t, p.AdmitSelect(q, h), "decide "+dq(q2)+" first")
	b.Answer(q2, dep, "")
	p = replayB(b)
	wantTurn(t, p, q, TurnWaits, WaitSubquestion)
	wantRefused(t, p.AdmitSelect(q, h), "has a sub-question not answered")
	b.Answer(sub, sh, "")
	wantTurn(t, replayB(b), q, TurnOwed)
}

// A select whose basis changed while its premise's question stays answered
// reopens its question as stale. The runner owes the stale node, the turn
// waits on it, and a reaffirm answers the question again: no select of the
// same Hypothesis is owed, and nothing is left owed by no one.
func TestStaleSelectOwesTheRunner(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q, q2 := b.DQ(g, "q"), b.DQ(g, "q2")
	h, dep := b.Hyp(q, "h"), b.Hyp(q2, "dep")
	rd := b.Supports(dep)
	b.Answer(q2, dep, "")
	b.Edge(h, event.EdgeDependsOn, dep)
	b.Supports(h)
	b.Answer(q, h, "")
	b.Retract(rd)
	b.Reaffirm(h) // its snapshot holds dep unsupported
	b.Supports(dep)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQReopened, ReasonSelectedStale)
	wantStatus(t, p, q2, DQAnswered, "")
	wantOwed(t, p, q, "runner stale "+h)
	wantTurn(t, p, q, TurnWaits, WaitRunner)
	wantRefused(t, p.AdmitSelect(q, h), "owes the runner first: stale "+hid(h))
	b.Reaffirm(h)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantOwed(t, p, q)
}

// A stale select whose premise's question is open waits on that question,
// and owes no stale item: deciding the premise comes first.
func TestStaleSelectWaitsOnAnOpenPremise(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q, q2 := b.DQ(g, "q"), b.DQ(g, "q2")
	h, dep := b.Hyp(q, "h"), b.Hyp(q2, "dep")
	rd := b.Supports(dep)
	b.Answer(q2, dep, "")
	b.Edge(h, event.EdgeDependsOn, dep)
	b.Supports(h)
	b.Answer(q, h, "")
	b.Retract(rd)
	p := replayB(b)
	wantStatus(t, p, q, DQReopened, ReasonSelectedStale)
	wantOwed(t, p, q)
	wantTurn(t, p, q, TurnWaits, WaitPremise)
}

// A select of one of two candidates reopens awaiting_subquestions when the
// selected Hypothesis gets a sub-question. The turn waits on that
// sub-question, so no choice is owed again, and answering it answers the
// question.
func TestReopenedChoiceWaitsOnTheSelectedSubquestion(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	b.As("a").Select(q, h1, "", false) // the owner's choice between two
	wantStatus(t, replayB(b), q, DQAnswered, "")
	sub := b.DQ(h1, "sub")
	p := replayB(b)
	wantStatus(t, p, q, DQReopened, ReasonAwaitingSubquestions)
	wantOwed(t, p, q)
	wantTurn(t, p, q, TurnWaits, WaitSubquestion)
	if k := p.Turn(q).Waits[0].Key; k != h1 {
		t.Fatalf("waits on a sub-question of %s, want %s", k, h1)
	}
	wantRefused(t, p.AdmitSelect(q, h2), hid(h1)+" has a sub-question not answered")
	sh := b.Hyp(sub, "sh")
	b.Supports(sh)
	b.Answer(sub, sh, "")
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantTurn(t, p, q, TurnClosed)
}

// A select whose Hypothesis is no longer supported leaves it no candidate,
// so the select is never owed again: a contradicts makes the other
// candidate's select owed, and a retracted support leaves runner work.
func TestSelectNoLongerSupportedIsNotOwedAgain(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	b.As("a").Select(q, h1, "", false)
	b.Contradicts(h1)
	p := replayB(b)
	wantStatus(t, p, q, DQReopened, ReasonSelectedNotSupport)
	wantOwed(t, p, q, "runner select "+h2)
	wantTurn(t, p, q, TurnOwed)
	wantRefused(t, p.AdmitSelect(q, h1), "owes a select of "+hid(h2)+", the only candidate, not "+hid(h1))

	b2 := testutil.New()
	q2 := b2.DQ(b2.Goal("g"), "q")
	h := b2.Hyp(q2, "h")
	r := b2.Supports(h)
	b2.Answer(q2, h, "")
	b2.Retract(r) // its Experiment is planned again
	p = replayB(b2)
	wantStatus(t, p, q2, DQReopened, ReasonSelectedNotSupport)
	wantTurn(t, p, q2, TurnWaits, WaitRunner)
	if it := p.Turn(q2).Waits[0].Item; it.Kind != OwedPlanned {
		t.Fatalf("waits on %+v, want the planned Experiment", it)
	}
	wantRefused(t, p.AdmitSelect(q2, h), "owes the runner first: planned ")
}
