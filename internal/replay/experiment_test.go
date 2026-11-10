package replay

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

func withdrawExp(b *testutil.B, e string) string {
	return b.Emit(event.TypeExperimentWithdraw, e, event.Empty{})
}

// An Experiment's status, first match wins: run, superseded, withdrawn,
// planned. Removing the supersedes edge reopens it.
func TestExpLifecycle(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	old := b.Exp(h1, event.RecipeNone, h2)
	repl := b.Exp(h1, event.RecipeNone, h2)
	gone := b.Exp(h2, event.RecipeNone, h1)
	b.Edge(repl, event.EdgeSupersedes, old)
	withdrawExp(b, gone)
	p := replayB(b)
	wantNoRejections(t, p)
	for _, c := range []struct {
		id             string
		status, closed ExpStatus
		by             []string
	}{
		{old, ExpSuperseded, ExpSuperseded, []string{repl}},
		{repl, ExpPlanned, "", nil},
		{gone, ExpWithdrawn, ExpWithdrawn, nil},
	} {
		x := p.Exps[c.id]
		if x.Status != c.status || x.Closed != c.closed || !reflect.DeepEqual(x.ReplacedBy, c.by) || statuses(p)[c.id].Status != string(c.status) {
			t.Errorf("%s: %+v (Status %q), want status %q closed %q by %v", c.id, x, statuses(p)[c.id].Status, c.status, c.closed, c.by)
		}
	}
	b.Unedge(repl, event.EdgeSupersedes, old)
	if x := replayB(b).Exps[old]; x.Status != ExpPlanned || x.Closed != "" || x.ReplacedBy != nil {
		t.Fatalf("after edge.remove: %+v, want planned again", x)
	}
}

// In a merged log a Result recorded under a closed Experiment still applies:
// the live Result wins, and Closed keeps the close for hyp check.
func TestExpLiveResultBeatsClose(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	e := b.Exp(h, event.RecipeNone)
	withdrawExp(b, e)
	r := b.Result(e, map[string]string{h: sup})
	p := replayB(b)
	wantNoRejections(t, p)
	if x := p.Exps[e]; x.Status != ExpRun || x.Closed != ExpWithdrawn || !reflect.DeepEqual(x.Live, []string{r}) {
		t.Fatalf("%+v, want run, closed withdrawn, live %s", x, r)
	}
	b.Retract(r)
	if x := replayB(b).Exps[e]; x.Status != ExpWithdrawn {
		t.Fatalf("Result retracted: %+v, want withdrawn", x)
	}
}

func TestExpCloseRefusals(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q, q2 := b.DQ(g, "q"), b.DQ(g, "q2")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q2, "h3")
	ran := b.Exp(h1, event.RecipeNone, h2)
	r := b.Result(ran, map[string]string{h1: sup, h2: inc})
	e1, e2 := b.Exp(h1, event.RecipeNone, h2), b.Exp(h2, event.RecipeNone, h1)
	other := b.Exp(h3, event.RecipeNone)
	gone := b.Exp(h1, event.RecipeNone, h2)
	withdrawExp(b, gone)
	b.Edge(e2, event.EdgeSupersedes, e1)
	st, p := Build(b.Evs, nil, "")
	wantNoRejections(t, p)

	mk := func(typ, node string, payload any) event.Event {
		saved := b.Evs
		b.Emit(typ, node, payload)
		ev := b.Evs[len(b.Evs)-1]
		b.Evs = saved
		return ev
	}
	sup := func(from, to string) event.Event {
		return mk(event.TypeEdgeAdd, from, event.Edge{Type: event.EdgeSupersedes, To: to})
	}
	for _, row := range []struct {
		name string
		ev   event.Event
		want string
	}{
		{"withdraw a Hypothesis", mk(event.TypeExperimentWithdraw, h1, event.Empty{}), "wrong kind"},
		{"withdraw twice", mk(event.TypeExperimentWithdraw, gone, event.Empty{}), "already withdrawn: " + gone},
		{"withdraw superseded", mk(event.TypeExperimentWithdraw, e1, event.Empty{}), "already superseded: " + e1},
		{"withdraw with a live Result", mk(event.TypeExperimentWithdraw, ran, event.Empty{}), "has the non-retracted Result " + r},
		{"withdraw with a payload", mk(event.TypeExperimentWithdraw, e2, map[string]string{"x": "y"}), "unknown field"},
		{"supersede onto a live Result", sup(e2, ran), "has the non-retracted Result " + r},
		{"supersede Experiment by Hypothesis", sup(h1, e2), "wrong kind: supersedes joins two Hypotheses or two Experiments"},
		{"supersede across questions", sup(other, e2), "cross-DQ supersedes"},
		{"supersede self", sup(e2, e2), "self edge"},
		{"supersede cycle", sup(e1, e2), "supersedes cycle"},
	} {
		if err := Validate(st, row.ev); err == nil || !strings.Contains(err.Error(), row.want) {
			t.Errorf("%s: got %v, want an error containing %q", row.name, err, row.want)
		}
	}
	if err := Validate(st, sup(gone, e2)); err != nil {
		t.Errorf("a withdrawn Experiment may supersede another: %v", err)
	}
}

func TestAdmitResultRefusesUnderClosedExperiment(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	old, repl, gone := b.Exp(h, event.RecipeNone), b.Exp(h, event.RecipeNone), b.Exp(h, event.RecipeNone)
	b.Edge(repl, event.EdgeSupersedes, old)
	withdrawExp(b, gone)
	st := stateB(b)
	ed := func(id string) string { return event.DisplayID(event.KindExperiment, id) }
	for exp, want := range map[string]string{
		old:  ed(old) + " is superseded by " + ed(repl) + "; record the Result there, or hyp unlink " + ed(repl) + " supersedes " + ed(old) + " to reopen " + ed(old),
		gone: ed(gone) + " is withdrawn; add a new Experiment for this Result",
	} {
		got := st.AdmitResult(event.ResultRecord{Parent: exp}, ResultFacts{}, nil)
		if len(got) != 1 || got[0] != (Finding{FindRefuse, want}) {
			t.Errorf("%s: %+v, want refusal %q", exp, got, want)
		}
	}
	if got := st.AdmitResult(event.ResultRecord{Parent: repl}, ResultFacts{}, nil); got != nil {
		t.Errorf("open Experiment: %+v, want nothing", got)
	}
}
