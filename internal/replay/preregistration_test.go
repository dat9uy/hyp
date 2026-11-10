package replay

import (
	"reflect"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// A Result gives a verdict on every target of its Experiment and on nothing
// else; a missing one is refused, so no mark is needed (ADR 0018).
func TestResultGivesAVerdictOnEveryTarget(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h1, h2, h3 := b.Hyp(q, "h1"), b.Hyp(q, "h2"), b.Hyp(q, "h3")
	e := b.Exp(h1, event.RecipeNone, h2, h3)
	if got, want := replayB(b).Exps[e].Targets, []string{h1, h2, h3}; !reflect.DeepEqual(got, want) {
		t.Fatalf("Targets = %v, want %v", got, want)
	}
	partial := b.Result(e, map[string]string{h1: sup, h2: con})
	full := b.Result(e, map[string]string{h1: sup, h2: con, h3: inc})
	p := replayB(b)
	wantRejected(t, p, partial, "bad verdicts: "+event.DisplayID(event.KindExperiment, e)+" targets "+event.DisplayID(event.KindHypothesis, h3)+"; give it a verdict, inconclusive when the run could not decide")
	if len(p.Rejections) != 1 {
		t.Fatalf("rejections = %+v, want only the partial Result", p.Rejections)
	}
	if !sameIDs(p.Exps[e].Live, full) {
		t.Fatalf("live Results = %v, want [%s]", p.Exps[e].Live, full)
	}
}

func sameIDs(got []string, want ...string) bool { return reflect.DeepEqual(got, want) }

// An Experiment under a DesignQuestion measures: it pins a recipe, has no
// targets and no rubric, and its Results give no verdict.
func TestMeasuringExperimentSitsUnderItsQuestion(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	measure := func(p event.NodeAdd) string {
		p.Kind, p.Parent, p.Title = event.KindExperiment, q, "count rows"
		return b.Emit(event.TypeNodeAdd, "", p)
	}
	hash := strings.Repeat("a", 64)
	ok := measure(event.NodeAdd{Recipe: "count@1", RecipeHash: hash})
	none := measure(event.NodeAdd{Recipe: event.RecipeNone})
	targeted := measure(event.NodeAdd{Recipe: "count@1", RecipeHash: hash, Targets: []string{h}})
	ruled := measure(event.NodeAdd{Recipe: "count@1", RecipeHash: hash, Rubric: event.RubricReading})
	counts := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: ok, Verdicts: map[string]string{}, Observation: "h stores 3 rows"})
	judged := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: ok, Verdicts: map[string]string{h: sup}, Observation: "o"})
	p := replayB(b)
	wantRejected(t, p, none, "bad experiment: an Experiment under a DesignQuestion measures, so it pins a recipe")
	wantRejected(t, p, targeted, "bad experiment: an Experiment under a DesignQuestion has no targets and no rubric; it measures")
	wantRejected(t, p, ruled, "bad experiment: an Experiment under a DesignQuestion has no targets and no rubric; it measures")
	wantRejected(t, p, judged, "bad verdicts: "+event.DisplayID(event.KindExperiment, ok)+" has no targets; it measures")
	if len(p.Rejections) != 4 {
		t.Fatalf("rejections = %+v, want the four above", p.Rejections)
	}
	x := p.Exps[ok]
	if x.Status != ExpRun || x.Targets != nil || x.Shape != ShapeMeasure || !sameIDs(x.Live, counts) {
		t.Fatalf("measuring Experiment = %+v, want run, no targets, %q, live [%s]", x, ShapeMeasure, counts)
	}
	if got := p.QuestionOf(ok); got != q {
		t.Fatalf("QuestionOf = %s, want %s", got, q)
	}
	// It moves no status and the question stays answered by what it was.
	if p.Hyps[h].Status != HypSupported {
		t.Fatalf("h = %s, want supported", p.Hyps[h].Status)
	}
}

// The shape is derived from the parent kind, the recipe, the rubric and the
// records; an older decided_by reading reads as rubric reading.
func TestShapeIsDerived(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	hash := strings.Repeat("b", 64)
	rec := []event.SubjectFile{{Path: ".tree/evidence/E-AAAAAA/command.json", SHA256: hash}}
	exp := func(p event.NodeAdd) string {
		p.Kind, p.Parent, p.Title = event.KindExperiment, h, "e"
		return b.Emit(event.TypeNodeAdd, "", p)
	}
	cases := []struct {
		id, want string
	}{
		{exp(event.NodeAdd{Recipe: "run@1", RecipeHash: hash, Rubric: "rule@1", RubricHash: hash}), ShapeRunAndRule},
		{exp(event.NodeAdd{Recipe: "run@1", RecipeHash: hash, Rubric: event.RubricReading}), ShapeReading},
		{exp(event.NodeAdd{Recipe: event.RecipeNone, Rubric: event.RubricReading}), ShapeReading},
		{exp(event.NodeAdd{Recipe: event.RecipeNone, DecidedBy: event.DecidedReading}), ShapeReading},
		{exp(event.NodeAdd{Recipe: event.RecipeNone, Rubric: event.RubricReading, Records: rec}), ShapeReassess},
		{exp(event.NodeAdd{Recipe: event.RecipeNone, Rubric: "rule@1", RubricHash: hash, Records: rec}), ShapeReassess},
	}
	p := replayB(b)
	wantNoRejections(t, p)
	for i, c := range cases {
		if got := p.Exps[c.id].Shape; got != c.want {
			t.Errorf("case %d: shape %q, want %q", i, got, c.want)
		}
	}
	if p.Nodes[cases[3].id].Rubric != event.RubricReading {
		t.Fatalf("decided_by reading reads as rubric %q, want reading", p.Nodes[cases[3].id].Rubric)
	}
}

func TestRubricAndRecordsRefusals(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	hash := strings.Repeat("c", 64)
	exp := func(p event.NodeAdd) string {
		p.Kind, p.Parent, p.Title = event.KindExperiment, h, "e"
		if p.Recipe == "" {
			p.Recipe = event.RecipeNone
		}
		return b.Emit(event.TypeNodeAdd, "", p)
	}
	cases := []struct{ id, want string }{
		{exp(event.NodeAdd{Rubric: "rule@1"}), "bad rubric: missing rubric_hash"},
		{exp(event.NodeAdd{Rubric: "Rule"}), "invalid rubric"},
		{exp(event.NodeAdd{RubricHash: hash}), "bad rubric: rubric_hash without a rubric bundle"},
		{exp(event.NodeAdd{Rubric: event.RubricReading, DecidedBy: event.DecidedReading}), "bad rubric: rubric and decided_by both given"},
		{exp(event.NodeAdd{Records: []event.SubjectFile{{Path: "a", SHA256: "zz"}}}), "bad records:"},
		{exp(event.NodeAdd{Records: []event.SubjectFile{{Path: "a", SHA256: hash}, {Path: "a", SHA256: hash}}}), "bad records: a named twice"},
		{b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "x", Rubric: event.RubricReading}), "unknown field"},
	}
	p := replayB(b)
	for _, c := range cases {
		wantRejected(t, p, c.id, c.want)
	}
}
