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

func wantAdmitted(t *testing.T, fs []Finding) {
	t.Helper()
	if slices.ContainsFunc(fs, func(f Finding) bool { return f.Kind != FindWarn }) {
		t.Fatalf("findings = %+v, want the select admitted", fs)
	}
}

func wantRefused(t *testing.T, fs []Finding, want string) {
	t.Helper()
	if len(fs) != 1 || fs[0].Kind != FindRefuse || !strings.Contains(fs[0].Text, want) {
		t.Fatalf("findings = %+v, want a refusal with %q", fs, want)
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
	wantRefused(t, p.AdmitSelect(q, h1), dq(q)+" owes the runner first: undecided "+hid(h2))

	e := b.Exp(h2, event.RecipeNone)
	wantOwed(t, replayB(b), q, "runner planned "+e)
	b.Result(e, map[string]string{h2: con})
	p = replayB(b)
	wantOwed(t, p, q, "runner select "+h1)
	wantAdmitted(t, p.AdmitSelect(q, h1))
	wantRefused(t, p.AdmitSelect(q, h2), "owes a select of "+hid(h1)+", the only candidate, not "+hid(h2))
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
	wantAdmitted(t, p.AdmitSelect(q, h1))
	wantAdmitted(t, p.AdmitSelect(q, h2))
	wantRefused(t, p.AdmitSelect(q, h3), hid(h3)+" is not a candidate of "+dq(q))

	sel := b.As("a").Select(q, h2, "", false)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantOwed(t, p, q)
	wantRefused(t, p.AdmitSelect(q, h1), dq(q)+" is answered; hyp unselect it first")

	// A new candidate reopens it, owes the choice again, and a select admits.
	b.Retract(c3)
	b.Emit(event.TypeExperimentWithdraw, e3, event.Empty{})
	p = replayB(b)
	wantStatus(t, p, q, DQReopened, ReasonNewCandidate)
	wantOwed(t, p, q, "owner choose "+q)
	wantAdmitted(t, p.AdmitSelect(q, h3))
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
	wantRefused(t, p.AdmitSelect(q1, h), hid(h)+" assumes "+hid(t1)+", in "+dq(q2)+" open; decide "+dq(q2)+" first")

	b.As("a").Select(q2, t2, "", false)
	p = replayB(b)
	wantStatus(t, p, q2, DQAnswered, "")
	wantOwed(t, p, q1, "runner premise "+h)
	wantRefused(t, p.AdmitSelect(q1, h), "premise "+hid(h)+" assumes "+hid(t1)+", but "+dq(q2)+" selected "+hid(t2))

	b.Unedge(h, event.EdgeDependsOn, t1)
	b.Edge(h, event.EdgeDependsOn, t2)
	b.Reaffirm(h)
	p = replayB(b)
	wantOwed(t, p, q1, "runner select "+h)
	wantAdmitted(t, p.AdmitSelect(q1, h))
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
	if fs := p.AdmitSelect(q, h1); len(fs) != 1 || fs[0].Kind != FindRefuse || !strings.Contains(fs[0].Text, "limit "+event.DisplayID(event.KindResult, dead)+"#1") {
		t.Fatalf("findings = %+v, want a refusal naming the rival's limit", fs)
	}
	b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "no test within the Goal", Refers: []string{k}})
	p = replayB(b)
	// The select waits on a referred limit until the owner accepts or returns it.
	wantOwed(t, p, q, "owner referred "+k)
	if fs := p.AdmitSelect(q, h1); len(fs) != 1 || !strings.Contains(fs[0].Text, "referred limit "+event.DisplayID(event.KindResult, dead)+"#1 waits for the owner") {
		t.Fatalf("findings = %+v, want a refusal naming the referred limit", fs)
	}
	b.As("o").Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "accept", Role: event.RoleOwner, Accepts: []string{k}})
	p = replayB(b)
	wantOwed(t, p, q, "runner select "+h1)
	if fs := p.AdmitSelect(q, h1); len(fs) != 0 {
		t.Fatalf("findings = %+v, want none once the limit is accepted", fs)
	}

	b.As("a").Select(q, h1, "", false)
	r := limitResult(b, h1, []string{"one host"}, nil)
	p = replayB(b)
	wantStatus(t, p, q, DQAnswered, "")
	wantOwed(t, p, q, "runner limit "+event.LimitKey(r, 1))
}
