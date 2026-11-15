package replay

import (
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// owedKinds lists the ledger's items on question q as "who kind key".
func owedKinds(p Projection, q string) []string {
	var out []string
	for _, o := range p.Owed() {
		if o.Question == q {
			out = append(out, string(o.Who)+" "+string(o.Kind)+" "+o.Key)
		}
	}
	return out
}

func wantOwed(t *testing.T, p Projection, q string, want ...string) {
	t.Helper()
	if got := owedKinds(p, q); !slices.Equal(got, want) {
		t.Fatalf("owed on %s = %q, want %q", q, got, want)
	}
}

// refusalFacts names a refusal by its facts, not its words: "admitted",
// "select H", "choose", the turn's kind, or "waits <kind> <key>" for the
// first thing the turn waits on. Package render owns the words.
func refusalFacts(r *Refusal) string {
	switch {
	case r == nil:
		return "admitted"
	case r.Kind != RefuseSelect:
		return string(r.Kind)
	case len(r.Owed) == 1:
		return "select " + r.Owed[0]
	case r.Choose:
		return "choose"
	case r.Turn.Kind != TurnWaits:
		return string(r.Turn.Kind)
	}
	switch w := r.Turn.Waits[0]; w.Kind {
	case WaitRunner:
		return "waits runner " + string(w.Item.Kind) + " " + w.Item.Key
	case WaitPremise:
		return "waits premise " + w.Premise.From + ">" + w.Premise.To
	default:
		return "waits " + string(w.Kind) + " " + w.Key
	}
}

func wantRefusal(t *testing.T, r *Refusal, want string) {
	t.Helper()
	if got := refusalFacts(r); !strings.HasPrefix(got, want) {
		t.Fatalf("refusal = %q (%+v), want %q", got, r, want)
	}
}

func dq(id string) string  { return event.DisplayID(event.KindDesignQuestion, id) }
func hid(id string) string { return event.DisplayID(event.KindHypothesis, id) }

// An inconclusive rival with no planned Experiment is owed by the runner, and
// it holds the lone candidate's select until an Experiment is planned for it.
func TestUndecidedRivalHoldsTheSelect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Result(b.Exp(h2, event.RecipeNone), map[string]string{h2: inc})
	p := replayB(b)
	wantOwed(t, p, q, "runner undecided "+h2)
	wantRefusal(t, p.SelectRefusal(q, h1), "waits runner undecided "+h2)

	e := b.Exp(h2, event.RecipeNone)
	wantOwed(t, replayB(b), q, "runner planned "+e)
	b.Result(e, map[string]string{h2: con})
	p = replayB(b)
	wantOwed(t, p, q, "runner select "+h1)
	wantRefusal(t, p.SelectRefusal(q, h1), "admitted")
	wantRefusal(t, p.SelectRefusal(q, h2), "select "+h1)
}

// A choice is admitted for any candidate and refused for a Hypothesis that is
// not one; once answered, nothing owes a select.
func TestAdmitSelectOnATradeoff(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3")
	b.Supports(h1)
	b.Supports(h2)
	b.Supports(h3)
	e3 := b.Exp(h3, event.RecipeNone)
	c3 := b.Result(e3, map[string]string{h3: con})
	p := replayB(b)
	wantOwed(t, p, q, "owner choose "+q)
	wantRefusal(t, p.SelectRefusal(q, h1), "admitted")
	wantRefusal(t, p.SelectRefusal(q, h2), "admitted")
	wantRefusal(t, p.SelectRefusal(q, h3), "choose")

	sel := b.As("a").Select(q, h2, "", false)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantOwed(t, p, q)
	wantRefusal(t, p.SelectRefusal(q, h1), "closed")

	// A new candidate reopens it, owes the choice again, and a select admits.
	b.Retract(c3)
	b.Emit(event.TypeExperimentWithdraw, e3, event.Empty{})
	p = replayB(b)
	wantStatus(t, p, q, DQReopened, ReasonNewCandidate)
	wantOwed(t, p, q, "owner choose "+q)
	wantRefusal(t, p.SelectRefusal(q, h3), "admitted")
	b.Select(q, h3, sel, false)
	wantStatus(t, replayB(b), q, DQAnswered, "")
}

// The select waits on a candidate's premise in an open question; once that
// question selects another Hypothesis, the premise is owed by the runner, and
// re-pointing the edge settles it.
func TestOpenPremiseAndPremiseItem(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	h := b.Hyp(q1, "h")
	t1, t2 := b.Hyp(q2, "t1"), b.Hyp(q2, "t2")
	b.Supports(t1)
	b.Supports(t2)
	b.Edge(h, event.EdgeDependsOn, t1)
	b.Supports(h)
	p := replayB(b)
	wantOwed(t, p, q1)
	wantOwed(t, p, q2, "owner choose "+q2)
	if pr := p.OpenPremise(q1); pr == nil || pr.From != h || pr.To != t1 {
		t.Fatalf("open premise = %+v, want %s on %s", pr, h, t1)
	}
	wantRefusal(t, p.SelectRefusal(q1, h), "waits premise "+h+">"+t1)

	b.As("a").Select(q2, t2, "", false)
	p = replayB(b)
	wantStatus(t, p, q2, DQAnswered, "")
	wantOwed(t, p, q1, "runner premise "+h)
	wantRefusal(t, p.SelectRefusal(q1, h), "waits runner premise "+h)

	b.Unedge(h, event.EdgeDependsOn, t1)
	b.Edge(h, event.EdgeDependsOn, t2)
	b.Reaffirm(h)
	p = replayB(b)
	wantOwed(t, p, q1, "runner select "+h)
	wantRefusal(t, p.SelectRefusal(q1, h), "admitted")
}

// A premise item is owed on an answered question too, so a premise that loses
// after the answer is not only listed; the status stays answered.
func TestPremiseItemOnAnAnsweredQuestion(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	h := b.Hyp(q1, "h")
	t1, t2 := b.Hyp(q2, "t1"), b.Hyp(q2, "t2")
	b.Supports(t1)
	b.Edge(h, event.EdgeDependsOn, t1)
	b.Supports(h)
	s2 := b.Answer(q2, t1, "")
	b.Answer(q1, h, "")
	wantStatus(t, replayB(b), q1, DQAnswered, "")
	b.Supports(t2)
	b.Select(q2, t2, s2, false)
	p := replayB(b)
	wantStatus(t, p, q1, DQAnswered, "")
	wantOwed(t, p, q1, "runner premise "+h)
}

func TestPremiseCycles(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	a1, a2 := b.Hyp(q1, "a1"), b.Hyp(q1, "a2")
	c1, c2 := b.Hyp(q2, "c1"), b.Hyp(q2, "c2")
	b.Edge(a1, event.EdgeDependsOn, c2)
	b.Edge(c1, event.EdgeDependsOn, a2)
	for _, h := range []string{a1, a2, c1, c2} {
		b.Supports(h)
	}
	p := replayB(b)
	wantNoRejections(t, p)
	if got := p.PremiseCycles(); len(got) != 1 || got[0] != [2]string{q1, q2} {
		t.Fatalf("premise cycles = %v, want [%s %s]", got, q1, q2)
	}
}

// An open limit is owed while its question has a candidate, whichever
// Hypothesis its Result judged: the runner owes it until it is referred,
// then the owner owes it, and an open limit stays owed after the answer.
func TestLimitOwedWhileQuestionHasCandidate(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	dead := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h2, event.RecipeNone),
		Verdicts: map[string]string{h2: con}, Observation: "o", Limits: []string{"instrument runs on SQLite"}})
	k := event.LimitKey(dead, 1)
	p := replayB(b)
	wantOwed(t, p, q, "runner limit "+k)
	wantRefusal(t, p.SelectRefusal(q, h1), "waits runner limit "+k)
	b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "no test within the Goal", Refers: []event.KeyText{{Key: k}}})
	p = replayB(b)
	// The select waits on a referred limit until the owner accepts or returns it.
	wantOwed(t, p, q, "owner referred "+k)
	wantRefusal(t, p.SelectRefusal(q, h1), "waits referred "+k)
	b.As("o").Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "accept", Role: event.RoleOwner, Accepts: []string{k}})
	p = replayB(b)
	wantOwed(t, p, q, "runner select "+h1)
	wantRefusal(t, p.SelectRefusal(q, h1), "admitted")

	b.As("a").Select(q, h1, "", false)
	r := limitResult(b, h1, []string{"one host"}, nil)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantOwed(t, p, q, "runner limit "+event.LimitKey(r, 1))
}
