package replay

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// answered is a question whose selected h1 has two supporting Results, with
// a second, untested rival. The select alone answers it (ADR 0015).
func answered() (b *testutil.B, q, h1, h2, sel, r2 string) {
	b = testutil.New()
	q = b.DQ(b.Goal("g"), "q")
	h1, h2 = b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	r2 = b.Supports(h1)
	sel = b.As("a").Select(q, h1, "", false)
	return b, q, h1, h2, sel, r2
}

func TestSelectAloneAnswers(t *testing.T) {
	b, q, h1, _, sel, _ := answered()
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	if got := p.Decisions[q].Selects[sel].Candidates; !slices.Equal(got, []string{h1}) {
		t.Fatalf("candidates at the select = %v, want [%s]", got, h1)
	}
}

// A Dossier change that adds no candidate leaves the question answered: a
// retract that keeps h1 supported, a new rival, a link, an amend.
func TestDossierChangeWithoutANewCandidateKeepsTheAnswer(t *testing.T) {
	b, q, h1, h2, _, r2 := answered()
	b.Retract(r2)
	b.Hyp(q, "h3")
	b.Contradicts(h2)
	q2 := b.DQ(b.Evs[0].ID, "other")
	dep := b.Hyp(q2, "dep")
	b.Supports(dep)
	b.Answer(q2, dep, "")
	b.Edge(h1, event.EdgeDependsOn, dep)
	amend(b, h1, "new body")
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

func TestNewCandidateReopensAndReselectAnswers(t *testing.T) {
	b, q, h1, h2, sel, _ := answered()
	b.Supports(h2)
	wantStatus(t, replayB(b), q, DQReopened, ReasonNewCandidate)
	b.Select(q, h1, sel, false)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// Retracting the contradicts that kept a rival out makes it a candidate again.
func TestRetractedDecidingContradictsIsANewCandidate(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	c := b.Contradicts(h2)
	b.As("a").Select(q, h1, "", false)
	wantStatus(t, replayB(b), q, DQAnswered, "")
	b.Retract(c)
	wantStatus(t, replayB(b), q, DQReopened, ReasonNewCandidate)
}

// A candidate that falls after the select does not reopen: the chosen one
// still stands.
func TestFallenCandidateKeepsTheAnswer(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	b.As("a").Select(q, h1, "", false) // the owner's choice between two
	wantStatus(t, replayB(b), q, DQAnswered, "")
	b.Contradicts(h2)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// The new_candidate reason comes after awaiting_subquestions.
func TestNewCandidateIsAfterAwaitingSubquestions(t *testing.T) {
	b, q, h1, h2, _, _ := answered()
	b.DQ(h1, "sub")
	b.Supports(h2)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingSubquestions)
}

// A trade-off is derived from the candidates, so an older select's stored
// tradeoff flag changes nothing.
func TestOldTradeoffFlagNeedsNoOwner(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	b.As("a").Emit(event.TypeDQSelect, q, event.DQSelect{Hypothesis: h, Tradeoff: true})
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// scopeAnswered is a scope-change select with a limit, confirmed by the
// owner, who accepts the limit.
func scopeAnswered() (b *testutil.B, q, h, sel string) {
	b = testutil.New()
	q = b.DQ(b.Goal("g"), "q")
	h = b.Hyp(q, "h")
	r := limitResult(b, h, []string{"l1"}, nil)
	sel = b.As("a").Select(q, h, "", true)
	b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
		Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r, 1)}})
	b.As("a")
	return b, q, h, sel
}

// A scope_change flag in an old log is history: the select is the gate, so
// the question is answered with no confirm (ADR 0019).
func TestOldScopeChangeFlagNeedsNoOwner(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	b.As("a").Select(q, h, "", true)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// An old owner confirm is replayed only as the accepts it made, at its place
// in the log (keyed.go). Nothing lapses: new evidence leaves the question
// answered and the accept standing, and a later insufficient confirm changes
// nothing.
func TestOldOwnerConfirmKeepsOnlyItsAccepts(t *testing.T) {
	b, q, h, sel := scopeAnswered()
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, limitAccepted, "o")

	b.Supports(h)
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, limitAccepted, "o")
}

// An old verifier confirm is read-only history: replay accepts it, whatever
// its verdict and actor, and it applies nothing.
func TestOldVerifierConfirmHasNoEffect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	sel := b.As("a").Select(q, h, "", false)
	// The same actor as the select: separation of duties refused this once.
	b.Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	b.As("v").Confirm(q, "not-a-select", event.RoleVerifier, event.ConfirmInsufficient)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

// Replaces rule 16 and the re-verify flag: an old dq.request_verify is
// read-only history. It applies whatever it holds and changes no status.
func TestOldRequestVerifyHasNoEffect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	request := func(sel string) { b.Emit(event.TypeDQRequestVerify, q, event.DQRequestVerify{Select: sel}) }

	sel := b.As("a").Select(q, h, "", false)
	request(sel)
	request("not-a-select")
	b.Emit(event.TypeDQRequestVerify, q, map[string]any{"unknown": true})
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}
