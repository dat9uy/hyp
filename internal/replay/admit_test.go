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

func TestAdmitExpBundleRules(t *testing.T) {
	instr := RecipeFacts{Instrument: true, HasCases: true}
	rec := []event.SubjectFile{{Path: ".tree/evidence/E-AAAAAA/out.txt", SHA256: strings.Repeat("a", 64)}}
	for _, c := range []struct {
		name  string
		p     event.NodeAdd
		f, rb RecipeFacts
		want  []string
		text  string
	}{
		{"instrument without known answers", event.NodeAdd{Recipe: "x@1"}, RecipeFacts{Instrument: true}, RecipeFacts{}, []string{FindRefuse}, "recipe x@1 is an instrument with no known answers"},
		{"recipe prints verdicts", event.NodeAdd{Recipe: "x@1", Rubric: event.RubricReading}, RecipeFacts{Instrument: true, HasCases: true, Verdicts: true}, RecipeFacts{},
			[]string{FindRefuse}, "recipe x@1: a recipe measures and never prints a verdict; put the rule in a rubric"},
		{"reading", event.NodeAdd{Recipe: event.RecipeNone, Rubric: event.RubricReading}, RecipeFacts{}, RecipeFacts{}, nil, ""},
		{"measured", event.NodeAdd{Recipe: "x@1"}, instr, RecipeFacts{}, nil, ""},
		{"run and rule", event.NodeAdd{Recipe: "x@1", Rubric: "r@1"}, instr, RecipeFacts{}, nil, ""},
		{"not a rubric", event.NodeAdd{Recipe: "x@1", Rubric: "r@1"}, instr, RecipeFacts{NotRubric: []string{"it has a step section", "verdict a: no supports case"}},
			[]string{FindRefuse}, "rubric r@1 is not a rubric: it has a step section; verdict a: no supports case"},
		{"rubric with nothing to read", event.NodeAdd{Recipe: event.RecipeNone, Rubric: "r@1"}, RecipeFacts{}, RecipeFacts{},
			[]string{FindRefuse}, "rubric r@1 has no measures to read: pin a recipe or --records"},
		{"reassess from records", event.NodeAdd{Recipe: event.RecipeNone, Rubric: "r@1", Records: rec}, RecipeFacts{}, RecipeFacts{}, nil, ""},
		{"template placeholder", event.NodeAdd{Recipe: "x@1"}, RecipeFacts{Instrument: true, HasCases: true, Placeholder: []string{"check.sh", "recipe.md"}}, RecipeFacts{},
			[]string{FindRefuse}, "recipe x@1 still holds the template placeholder (hyp:placeholder) in check.sh, recipe.md"},
		{"rubric placeholder", event.NodeAdd{Recipe: "x@1", Rubric: "r@1"}, instr, RecipeFacts{Placeholder: []string{"rule.py"}},
			[]string{FindRefuse}, "rubric r@1 still holds the template placeholder (hyp:placeholder) in rule.py"},
		{"placeholder on recipe none is ignored", event.NodeAdd{Recipe: event.RecipeNone, Rubric: event.RubricReading}, RecipeFacts{Placeholder: []string{"recipe.md"}}, RecipeFacts{}, nil, ""},
	} {
		got := (*State)(nil).AdmitExp(c.p, c.f, c.rb, false)
		if !reflect.DeepEqual(kinds(got), c.want) || (c.text != "" && !strings.HasPrefix(got[0].Text, c.text)) {
			t.Errorf("%s: findings %+v, want kinds %v starting %q", c.name, got, c.want, c.text)
		}
	}
}

// Under a Hypothesis an Experiment pins a rubric; under a DesignQuestion it
// measures, so it has no rubric and no target.
func TestAdmitExpParentRules(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	st := stateB(b)
	for _, c := range []struct {
		name     string
		p        event.NodeAdd
		noTarget bool
		want     []string
		text     string
	}{
		{"hypothesis without rubric", event.NodeAdd{Parent: h, Recipe: event.RecipeNone}, false, []string{FindRefuse}, NoRubric},
		{"hypothesis with reading", event.NodeAdd{Parent: h, Recipe: event.RecipeNone, Rubric: event.RubricReading}, false, nil, ""},
		{"question measures", event.NodeAdd{Parent: q, Recipe: "x@1"}, false, nil, ""},
		{"question with rubric", event.NodeAdd{Parent: q, Recipe: "x@1", Rubric: event.RubricReading}, false, []string{FindBadArgs}, "an Experiment under a DesignQuestion measures"},
		{"question with target", event.NodeAdd{Parent: q, Recipe: "x@1", Targets: []string{h}}, false, []string{FindBadArgs}, "an Experiment under a DesignQuestion measures"},
		{"question with target none", event.NodeAdd{Parent: q, Recipe: "x@1"}, true, []string{FindBadArgs}, "an Experiment under a DesignQuestion measures"},
	} {
		got := st.AdmitExp(c.p, RecipeFacts{Instrument: true, HasCases: true}, RecipeFacts{}, c.noTarget)
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
	reading := event.NodeAdd{Recipe: event.RecipeNone, Rubric: event.RubricReading, Parent: h1}

	got := st.AdmitExp(reading, RecipeFacts{}, RecipeFacts{}, false)
	if len(got) != 1 || got[0].Kind != FindRefuse || !strings.HasPrefix(got[0].Text, "rivals of "+event.DisplayID(event.KindHypothesis, h1)+" under ") {
		t.Fatalf("no target while %s is live: %+v, want the one-rival refusal", h2, got)
	}
	if got := st.AdmitExp(reading, RecipeFacts{}, RecipeFacts{}, true); got != nil {
		t.Fatalf("--target none: %+v, want nothing", got)
	}
	reading.Targets = []string{h2}
	if got := st.AdmitExp(reading, RecipeFacts{}, RecipeFacts{}, false); got != nil {
		t.Fatalf("targets the rival: %+v, want nothing", got)
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
	p := event.NodeAdd{Recipe: "x@1", Rubric: event.RubricReading, Parent: h1, Targets: []string{h2}}
	got := stateB(b).AdmitExp(p, RecipeFacts{Instrument: true, HasCases: true, GoWarnings: []string{"a.go: go file"}}, RecipeFacts{}, false)
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
		got := st.AdmitResult(event.ResultRecord{Parent: c.exp}, ResultFacts{Recipe: c.f}, c.refHash)
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
