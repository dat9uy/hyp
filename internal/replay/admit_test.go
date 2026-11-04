package replay

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

func stateB(b *testutil.B) *State {
	st, _ := Build(b.Evs, nil, "")
	return st
}

func kinds(fs []Finding) []string {
	var out []string
	for _, f := range fs {
		out = append(out, f.Kind)
	}
	return out
}

func TestAdmitExpRecipeRules(t *testing.T) {
	instr := RecipeFacts{Instrument: true, HasCases: true}
	for _, c := range []struct {
		name string
		p    event.NodeAdd
		f    RecipeFacts
		want []string
		text string
	}{
		{"instrument without known answers", event.NodeAdd{Recipe: "x@1"}, RecipeFacts{Instrument: true}, []string{FindRefuse}, "recipe x@1 is an instrument with no known answers"},
		{"measured with decided-by", event.NodeAdd{Recipe: "x@1", DecidedBy: event.DecidedReading}, instr, []string{FindBadArgs}, "--decided-by reading contradicts recipe x@1"},
		{"recipe none, no decided-by", event.NodeAdd{Recipe: event.RecipeNone}, RecipeFacts{}, []string{FindRefuse}, "recipe none holds no measure"},
		{"not an instrument, no decided-by", event.NodeAdd{Recipe: "x@1"}, RecipeFacts{}, []string{FindRefuse}, "recipe x@1 holds no measure"},
		{"reading", event.NodeAdd{Recipe: event.RecipeNone, DecidedBy: event.DecidedReading}, RecipeFacts{}, nil, ""},
		{"measured", event.NodeAdd{Recipe: "x@1"}, instr, nil, ""},
		{"template placeholder", event.NodeAdd{Recipe: "x@1"}, RecipeFacts{Instrument: true, HasCases: true, Placeholder: []string{"check.sh", "recipe.md"}},
			[]string{FindRefuse}, "recipe x@1 still holds the template placeholder (hyp:placeholder) in check.sh, recipe.md"},
		{"placeholder on recipe none is ignored", event.NodeAdd{Recipe: event.RecipeNone, DecidedBy: event.DecidedReading}, RecipeFacts{Placeholder: []string{"recipe.md"}}, nil, ""},
	} {
		got := (*State)(nil).AdmitExp(c.p, c.f, false)
		if !reflect.DeepEqual(kinds(got), c.want) || (c.text != "" && !strings.HasPrefix(got[0].Text, c.text)) {
			t.Errorf("%s: findings %+v, want kinds %v starting %q", c.name, got, c.want, c.text)
		}
	}
}

func TestAdmitExpOneRival(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	st := stateB(b)
	reading := event.NodeAdd{Recipe: event.RecipeNone, DecidedBy: event.DecidedReading, Parent: h1}

	got := st.AdmitExp(reading, RecipeFacts{}, false)
	if len(got) != 1 || got[0].Kind != FindRefuse || !strings.HasPrefix(got[0].Text, "live rivals of "+event.DisplayID(event.KindHypothesis, h1)+" under ") {
		t.Fatalf("no target while %s is live: %+v, want the one-rival refusal", h2, got)
	}
	if got := st.AdmitExp(reading, RecipeFacts{}, true); got != nil {
		t.Fatalf("--target none: %+v, want nothing", got)
	}
	reading.Targets = []string{h2}
	if got := st.AdmitExp(reading, RecipeFacts{}, false); got != nil {
		t.Fatalf("targets the live rival: %+v, want nothing", got)
	}

	// The refusal and the record read one rule: an Experiment added with
	// --target none under the same state is marked one-rival.
	e := b.Exp(h1, event.RecipeNone)
	if !Replay(b.Evs, nil, "").Exps[e].OneRival {
		t.Fatal("Exps disagrees with AdmitExp's one-rival refusal")
	}
}

func TestAdmitExpWarnings(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	r := b.Contradicts(h2)
	p := event.NodeAdd{Recipe: "x@1", Parent: h1, Targets: []string{h2}}
	got := stateB(b).AdmitExp(p, RecipeFacts{Instrument: true, HasCases: true, GoWarnings: []string{"a.go: go file"}}, false)
	want := []Finding{
		{FindWarn, "D1: target " + event.DisplayID(event.KindHypothesis, h2) + " is contradicted by " + event.DisplayID(event.KindResult, r) +
			"; if that was wrong, hyp retract <R-id> reopens it; otherwise supersede " + event.DisplayID(event.KindHypothesis, h2) + " and test the new Hypothesis"},
		{FindWarn, "recipe x@1: a.go: go file"},
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("findings\n%+v\nwant\n%+v", got, want)
	}
}

func TestAdmitResultPinRules(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	e := b.Exp(h, "x@1") // pins hash "aaa"
	none := b.Exp(h, event.RecipeNone)
	st := stateB(b)
	ed := event.DisplayID(event.KindExperiment, e)
	for _, c := range []struct {
		name    string
		exp     string
		f       RecipeFacts
		refHash map[string]string
		text    string
	}{
		{"unchanged", e, RecipeFacts{Hash: "aaa"}, map[string]string{"x@1": "aaa"}, ""},
		{"recipe none", none, RecipeFacts{Err: errors.New("ignored")}, nil, ""},
		{"no longer loads", e, RecipeFacts{Err: errors.New("gone")}, nil, "recipe x@1 of " + ed + " no longer loads (gone)"},
		{"changed since pinned", e, RecipeFacts{Hash: "bbb"}, nil, "recipe x@1 changed since " + ed + " pinned it (recorded aaa, now bbb)"},
		{"evidence hashed differently", e, RecipeFacts{Hash: "aaa"}, map[string]string{"x@1": "ccc"}, "recipe x@1 changed while recording: its evidence hashed ccc, " + ed + " pinned aaa"},
	} {
		got := st.AdmitResult(event.ResultRecord{Parent: c.exp}, c.f, c.refHash)
		if c.text == "" {
			if got != nil {
				t.Errorf("%s: %+v, want nothing", c.name, got)
			}
			continue
		}
		if len(got) != 1 || got[0].Kind != FindRefuse || !strings.HasPrefix(got[0].Text, c.text) {
			t.Errorf("%s: %+v, want a refusal starting %q", c.name, got, c.text)
		}
	}
}

func TestAdmitConfirm(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	r := limitResult(b, h, []string{"l1"}, nil)
	key := event.LimitKey(r, 1)
	owner := func(accepts ...string) event.Confirm {
		return event.Confirm{Role: event.RoleOwner, Verdict: event.ConfirmSufficient, Accepts: accepts}
	}
	refused := func(fs []Finding, want string) {
		t.Helper()
		if len(fs) != 1 || fs[0].Kind != FindRefuse || !strings.Contains(fs[0].Text, want) {
			t.Fatalf("findings = %+v, want a refusal with %q", fs, want)
		}
	}

	refused(replayB(b).AdmitConfirm(q, owner()), "no current selection")
	refused(replayB(b).AdmitConfirm(h, owner()), "no current selection")

	b.As("a").Select(q, h, "", true)
	p := replayB(b)
	refused(p.AdmitConfirm(q, owner()), "open limits under "+event.DisplayID(event.KindDesignQuestion, q)+" are not accepted:\n  "+event.DisplayID(event.KindResult, r)+"#1 l1\n")
	for name, c := range map[string]event.Confirm{
		"accepted":           owner(key),
		"owner insufficient": {Role: event.RoleOwner, Verdict: event.ConfirmInsufficient},
		"verifier":           {Role: event.RoleVerifier, Verdict: event.ConfirmSufficient},
	} {
		if fs := p.AdmitConfirm(q, c); len(fs) != 0 {
			t.Fatalf("%s: findings = %+v, want none", name, fs)
		}
	}
}
