package replay

import (
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// answered is a question whose selected h1 has two supporting Results and a
// sufficient verifier confirm, with a second, untested rival.
func answered() (b *testutil.B, q, h1, h2, sel, r2 string) {
	b = testutil.New()
	q = b.DQ(b.Goal("g"), "q")
	h1, h2 = b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	r2 = b.Supports(h1)
	sel = b.As("a").Select(q, h1, "", false)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As("a")
	return b, q, h1, h2, sel, r2
}

// The event of #24: a Result the verifier judged is retracted.
func TestRetractOfAJudgedResultReopens(t *testing.T) {
	b, q, _, _, _, r2 := answered()
	wantStatus(t, replayB(b), q, DQAnswered, "")
	b.Retract(r2)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQReopened, ReasonAwaitingVerifier)
}

func TestRetractOfALaterResultRestoresTheConfirm(t *testing.T) {
	b, q, h1, _, _, _ := answered()
	r := b.Supports(h1)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	b.Retract(r)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestNewResultOnARivalLapsesTheConfirm(t *testing.T) {
	b, q, _, h2, _, _ := answered()
	b.Contradicts(h2)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
}

func TestNewRivalLapsesAndAbandoningItRestores(t *testing.T) {
	b, q, _, _, _, _ := answered()
	h3 := b.Hyp(q, "h3")
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	b.Abandon(h3)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestAbandonedRivalLapsesTheConfirm(t *testing.T) {
	b, q, _, h2, _, _ := answered()
	b.Abandon(h2)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
}

func TestExternalRefAndIncomingLinkKeepTheConfirm(t *testing.T) {
	b, q, h1, _, _, _ := answered()
	b.Edge(h1, event.EdgeExternalRef, "gh:1")
	q2 := b.DQ(b.Evs[0].ID, "other")
	b.Edge(b.Hyp(q2, "points in"), event.EdgeDependsOn, h1)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

func TestOutgoingLinkLapsesAndRemovingItRestores(t *testing.T) {
	b, q, h1, _, _, _ := answered()
	q2 := b.DQ(b.Evs[0].ID, "other")
	dep := b.Hyp(q2, "dep")
	b.Supports(dep)
	b.Edge(h1, event.EdgeDependsOn, dep)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	b.Unedge(h1, event.EdgeDependsOn, dep)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestAmendLapsesAndAmendBackRestores(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old"})
	b.Supports(h)
	sel := b.As("a").Select(q, h, "", false)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	amend(b.As("a"), h, "new")
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	amend(b, h, "old")
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestAmendOfARivalKeepsTheConfirm(t *testing.T) {
	b, q, _, h2, _, _ := answered()
	amend(b, h2, "rival's body")
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// The latest Confirm whose Dossier matches is the one that counts: D1
// sufficient, D2 insufficient, then back at D1 the verdict is sufficient.
func TestLatestMatchingConfirmCounts(t *testing.T) {
	b, q, h1, _, sel, _ := answered()
	r := b.Supports(h1)
	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	wantStatus(t, replayB(b), q, DQReopened, ReasonVerifierInsufficient)
	b.As("a").Retract(r)
	p := replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	if c := p.Decisions[q].Counting(event.RoleVerifier); c.Verdict != event.ConfirmSufficient || c.Actor != "v" {
		t.Fatalf("counting = %+v, want v's sufficient", c)
	}
	if c := p.Decisions[q].Latest(event.RoleVerifier); c.Actor != "w" {
		t.Fatalf("latest = %+v, want w's", c)
	}
}

func TestNewVerifierConfirmOnTheSameDossierSupersedes(t *testing.T) {
	b, q, _, _, sel, _ := answered()
	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	wantStatus(t, replayB(b), q, DQReopened, ReasonVerifierInsufficient)
	b.As("x").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// A reopened question closes by appending: a fresh verifier confirm when the
// claim is unchanged.
func TestFreshConfirmClosesAReopenedQuestion(t *testing.T) {
	b, q, _, _, sel, r2 := answered()
	b.Retract(r2)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// ... and a new superseding Hypothesis when the claim changed.
func TestSupersedingHypothesisClosesAReopenedQuestion(t *testing.T) {
	b, q, h1, _, sel, r2 := answered()
	b.Retract(r2)
	h3 := b.Hyp(q, "h3")
	b.Edge(h3, event.EdgeSupersedes, h1)
	b.Supports(h3)
	s2 := b.Select(q, h3, sel, false)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingVerifier)
	b.As("w").Confirm(q, s2, event.RoleVerifier, event.ConfirmSufficient)
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}

// ownerAnswered is a trade-off select with a limit, confirmed by both roles,
// the owner accepting the limit.
func ownerAnswered() (b *testutil.B, q, h, sel string) {
	b = testutil.New()
	q = b.DQ(b.Goal("g"), "q")
	h = b.Hyp(q, "h")
	r := limitResult(b, h, []string{"l1"}, nil)
	sel = b.As("a").Select(q, h, "", true)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As("o").Emit(event.TypeConfirm, q, event.Confirm{Select: sel, Role: event.RoleOwner,
		Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r, 1)}})
	b.As("a")
	return b, q, h, sel
}

func TestOwnerConfirmAndLimitLapseAndRestore(t *testing.T) {
	b, q, h, sel := ownerAnswered()
	p := replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, LimitAccepted, "o")

	r := b.Supports(h)
	p = replayB(b)
	wantLimit(t, p, q, 0, LimitOpen, "")
	wantStatus(t, p, q, DQReopened, ReasonAwaitingVerifier)
	if !p.LimitsOf(q)[0].Gating {
		t.Fatal("a lapsed acceptance leaves the limit not gating")
	}

	// A fresh verifier is not enough: the owner's sign-off lapsed too.
	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingOwner)

	b.As("a").Retract(r)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantLimit(t, p, q, 0, LimitAccepted, "o")
}

func TestOwnerInsufficientAtTheSameDossierWithdrawsAcceptance(t *testing.T) {
	b, q, h, sel := ownerAnswered()
	r := b.Supports(h)
	// An owner's no at another Dossier leaves the acceptance at the first one.
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	b.As("a").Retract(r)
	p := replayB(b)
	wantLimit(t, p, q, 0, LimitAccepted, "o")
	wantStatus(t, p, q, DQAnswered, "")

	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	p = replayB(b)
	wantLimit(t, p, q, 0, LimitOpen, "")
	wantStatus(t, p, q, DQReopened, ReasonAwaitingOwner)
}

// Replaces the owner_required frozen at the select: two candidates now
// require the owner, whatever held at the select.
func TestOwnerRequiredFromCandidatesNow(t *testing.T) {
	b, q, _, h2, sel, _ := answered()
	b.Supports(h2)
	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQReopened, ReasonAwaitingOwner)
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestOwnerNotRequiredOnceCandidatesAreGone(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	sel := b.As("a").Select(q, h1, "", false) // two candidates here
	b.Contradicts(h2)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

func TestOwnerRequiredByTheSelectsFlags(t *testing.T) {
	for name, payload := range map[string]func(h string) event.DQSelect{
		"tradeoff":     func(h string) event.DQSelect { return event.DQSelect{Hypothesis: h, Tradeoff: true} },
		"scope change": func(h string) event.DQSelect { return event.DQSelect{Hypothesis: h, ScopeChange: true} },
	} {
		t.Run(name, func(t *testing.T) {
			b := testutil.New()
			q := b.DQ(b.Goal("g"), "q")
			h := b.Hyp(q, "h")
			b.Supports(h)
			sel := b.As("a").Emit(event.TypeDQSelect, q, payload(h))
			b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
			wantStatus(t, replayB(b), q, DQOpen, ReasonAwaitingOwner)
			b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmSufficient)
			wantStatus(t, replayB(b), q, DQAnswered, "")
		})
	}
}

// Replaces rule 16 and the re-verify flag: an old dq.request_verify is
// read-only history. It applies whatever it holds, including what rule 16
// refused, and changes no status, chain or limit.
func TestOldRequestVerifyHasNoEffect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	request := func(sel string) { b.Emit(event.TypeDQRequestVerify, q, event.DQRequestVerify{Select: sel}) }

	sel := b.As("a").Select(q, h, "", false)
	request(sel) // no verifier confirm yet: rule 16 refused this
	p := replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQOpen, ReasonAwaitingVerifier)

	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	b.As("a")
	request(sel)
	request(sel) // a request already open: rule 16 refused this
	request("not-a-select")
	b.Emit(event.TypeDQRequestVerify, q, map[string]any{"unknown": true})
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQOpen, ReasonVerifierInsufficient)
	if s := p.Decisions[q]; len(s.Chain) != 1 {
		t.Fatalf("chain = %v, want the one verdict", s.Chain)
	}

	b.As("w").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As("a")
	request(sel) // latest confirm not insufficient: rule 16 refused this
	p = replayB(b)
	wantNoRejections(t, p)
	wantStatus(t, p, q, DQAnswered, "")
}
