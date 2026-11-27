package acts

import (
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
)

func TestCfgRulesReadsRuleSectionsOnly(t *testing.T) {
	cfg := `\* a comment that names INVARIANT Hidden
SPECIFICATION Spec
CONSTANTS
    Skip = FALSE
INVARIANTS TypeOK
    (* Gone *) Safe
INVARIANT Safe
PROPERTY Live \* trailing comment
CHECK_DEADLOCK FALSE
`
	got := CfgRules(cfg)
	if want := []string{"TypeOK", "Safe", "Live"}; !slices.Equal(got, want) {
		t.Fatalf("rules = %q, want %q", got, want)
	}
}

// landSpec is what the owner's land on commit, which meets every example,
// gives: nil, or the Spec or Own path refusal that also holds back the
// calculated select of that commit.
func (e *runEnv) landSpec(t *testing.T, commit string) error {
	t.Helper()
	_, err := Land(LandOpts{Dir: e.root, Actor: "owner-1", Role: "owner", Goal: e.goal, Commit: commit})
	return err
}

// selectModel commits files on a leaf from H0, runs line-1 on it and lands
// it, which reads the spec as a calculated select does.
func (e *runEnv) selectModel(t *testing.T, files map[string]string) error {
	t.Helper()
	all := map[string]string{"counter": "ok\n"}
	for p, s := range files {
		all["goals/"+e.goal+"/specs/"+p] = s
	}
	commit := e.child(t, e.base, all)
	if _, err := e.run(commit); err != nil {
		t.Fatal(err)
	}
	return e.landSpec(t, commit)
}

const (
	modelTLA  = "---- MODULE m ----\n====\n"
	checkCfg  = "SPECIFICATION Spec\nINVARIANTS\n    NamedRule\n    LonelyRule\n"
	namedCfg  = "SPECIFICATION Spec\nINVARIANTS\n    NamedRule\n"
	lonelyCfg = "SPECIFICATION Spec\nINVARIANT LonelyRule\n"
	chartMmd  = "stateDiagram-v2\n  [*] --> s0\n"
)

func TestSelectRefusesARuleNoVariantNames(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectModel(t, map[string]string{
		"m.tla": modelTLA, "m.cfg": checkCfg, "switches/named.cfg": namedCfg, "chart.mmd": chartMmd,
	})
	refusedBy(t, err, "Spec")
	f, _ := fault.As(err)
	if f.Code != fault.RuleNoVariant {
		t.Fatalf("code = %s", f.Code)
	}
	if rules, _ := f.Args[1].([]string); !slices.Equal(rules, []string{"LonelyRule"}) {
		t.Fatalf("rules = %v", f.Args[1])
	}
}

func TestSelectRefusesAModelWithNoChart(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectModel(t, map[string]string{
		"m.tla": modelTLA, "m.cfg": checkCfg, "switches/named.cfg": namedCfg, "switches/lonely.cfg": lonelyCfg,
		// a chart under switches/ is not the model's chart
		"switches/chart.mmd": chartMmd,
	})
	refusedBy(t, err, "Spec")
	if f, _ := fault.As(err); f.Code != fault.NoChart {
		t.Fatalf("code = %s", f.Code)
	}
}

func TestSelectRecordsAModelWithOldVariantFolderAndChart(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectModel(t, map[string]string{
		"m.tla": modelTLA, "m.cfg": checkCfg, "switches/named.cfg": namedCfg, "switches/lonely.cfg": lonelyCfg, "chart.mmd": chartMmd,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSelectReadsVariantConfigsUnderVariants(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectModel(t, map[string]string{
		"m.tla": modelTLA, "m.cfg": checkCfg, "variants/named.cfg": namedCfg, "variants/lonely.cfg": lonelyCfg, "chart.mmd": chartMmd,
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSelectRefusalNamesTheVariantsFolder(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectModel(t, map[string]string{
		"m.tla": modelTLA, "m.cfg": checkCfg, "variants/named.cfg": namedCfg, "chart.mmd": chartMmd,
	})
	refusedBy(t, err, "Spec")
	f, _ := fault.As(err)
	if f.Code != fault.RuleNoVariant {
		t.Fatalf("code = %s", f.Code)
	}
	if dir, _ := f.Args[2].(string); !strings.HasSuffix(dir, "/specs/variants/") {
		t.Fatalf("folder = %v", f.Args[2])
	}
}

func TestSelectRecordsASpecWithNoModel(t *testing.T) {
	e := newRunEnv(t)
	// configs and no .tla file are no model: nothing to check
	if err := e.selectModel(t, map[string]string{"notes.md": "notes\n", "m.cfg": checkCfg}); err != nil {
		t.Fatal(err)
	}
}

func TestUnnamedKeepsRuleOrder(t *testing.T) {
	m := Model{Rules: []string{"A", "B", "C"}, Named: []string{"B"}}
	if got := strings.Join(m.Unnamed(), ","); got != "A,C" {
		t.Fatalf("unnamed = %s", got)
	}
}

// TestHeldBackNamesWhatSelectRefuses: a lone candidate that hyp select
// refuses is held back with the refusal select printed.
func TestHeldBackNamesWhatSelectRefuses(t *testing.T) {
	e := newRunEnv(t)
	selErr := e.selectModel(t, map[string]string{"m.tla": modelTLA, "m.cfg": checkCfg, "variants/named.cfg": namedCfg})
	tb, _, err := ReadTables(e.root)
	if err != nil {
		t.Fatal(err)
	}
	g, err := tb.Find(e.goal)
	if err != nil {
		t.Fatal(err)
	}
	holds := HeldBack(g)
	if len(holds) != 1 {
		t.Fatalf("holds = %v", holds)
	}
	want, _ := fault.As(selErr)
	got, _ := fault.As(holds[0].Err)
	if want == nil || got == nil || got.Guard != want.Guard || got.Code != want.Code || got.Code != fault.RuleNoVariant {
		t.Fatalf("held back by %v, select refused with %v", holds[0].Err, selErr)
	}
}

const (
	keptTLA = "---- MODULE kept ----\nOldRule == x \\in {0}\n\\* source: model only\nNewRule ==\n    x \\in {0, 1}\n====\n"
	keptCfg = "SPECIFICATION Spec\nINVARIANTS OldRule NewRule\n"
	newCfg  = "SPECIFICATION Spec\nINVARIANTS NewRule\n"
)

// selectKept commits the kept model files and the spec files on a leaf from
// H0, runs line-1 on it and selects.
func (e *runEnv) selectKept(t *testing.T, kept, spec map[string]string) error {
	t.Helper()
	all := map[string]string{"counter": "ok\n"}
	for p, s := range kept {
		all[KeptDir+p] = s
	}
	for p, s := range spec {
		all["goals/"+e.goal+"/specs/"+p] = s
	}
	commit := e.child(t, e.base, all)
	if _, err := e.run(commit); err != nil {
		t.Fatal(err)
	}
	return e.landSpec(t, commit)
}

func TestSelectRefusesANewKeptRuleNoVariantNames(t *testing.T) {
	e := newRunEnv(t)
	err := e.selectKept(t, map[string]string{"kept.tla": keptTLA, "kept.cfg": keptCfg},
		map[string]string{"variants/new.cfg": newCfg, "chart.mmd": chartMmd})
	refusedBy(t, err, "Spec")
	f, _ := fault.As(err)
	if f.Code != fault.RuleNoVariant {
		t.Fatalf("code = %s", f.Code)
	}
	if rules, _ := f.Args[1].([]string); !slices.Equal(rules, []string{"OldRule"}) {
		t.Fatalf("rules = %v", f.Args[1])
	}
}

func TestSelectRecordsAKeptChangeWhoseRulesAreNamed(t *testing.T) {
	e := newRunEnv(t)
	both := "SPECIFICATION Spec\nINVARIANTS OldRule NewRule\n"
	err := e.selectKept(t, map[string]string{"kept.tla": keptTLA, "kept.cfg": keptCfg},
		map[string]string{"variants/both.cfg": both, "chart.mmd": chartMmd})
	if err != nil {
		t.Fatal(err)
	}
}

func TestSelectRefusesAKeptChangeWithNoChart(t *testing.T) {
	e := newRunEnv(t)
	both := "SPECIFICATION Spec\nINVARIANTS OldRule NewRule\n"
	err := e.selectKept(t, map[string]string{"kept.tla": keptTLA, "kept.cfg": keptCfg},
		map[string]string{"variants/both.cfg": both})
	refusedBy(t, err, "Spec")
	if f, _ := fault.As(err); f.Code != fault.NoChart {
		t.Fatalf("code = %s", f.Code)
	}
}

func TestKeptChangeIsNewAndChangedRulesOnly(t *testing.T) {
	e := newRunEnv(t)
	from := e.child(t, e.base, map[string]string{KeptDir + "kept.tla": keptTLA, KeptDir + "kept.cfg": keptCfg})
	changed := strings.Replace(keptTLA, "{0, 1}", "{0, 1, 2}", 1) + "Added == TRUE\n"
	to := e.child(t, from, map[string]string{
		KeptDir + "kept.tla": changed,
		KeptDir + "kept.cfg": "SPECIFICATION Spec\nINVARIANTS OldRule NewRule Added\n",
	})
	got, err := KeptChange(e.root, from, to)
	if err != nil {
		t.Fatal(err)
	}
	if want := []string{"NewRule", "Added"}; !slices.Equal(got, want) {
		t.Fatalf("kept change = %q, want %q", got, want)
	}
	if got, _ := KeptChange(e.root, from, from); len(got) != 0 {
		t.Fatalf("kept change of a commit against itself = %q", got)
	}
}

func TestKeptDiffListsChangedStepsAndRuleSources(t *testing.T) {
	e := newRunEnv(t)
	base := "---- MODULE kept ----\n\\* command: toy add\nInc == n' = n + 1\n\\* source: model only\nKeeps == n >= 0\n====\n"
	cfg := "SPECIFICATION Spec\nINVARIANTS Keeps\n"
	from := e.child(t, e.base, map[string]string{KeptDir + "kept.tla": base, KeptDir + "kept.cfg": cfg})
	changed := strings.Replace(base, "====\n",
		"\\* command: toy fold\nFold == n' = n\n\\* source: G-ONE:line-1\nNew == n < 3\n====\n", 1)
	to := e.child(t, from, map[string]string{
		KeptDir + "kept.tla": changed,
		KeptDir + "kept.cfg": "SPECIFICATION Spec\nINVARIANTS Keeps New\n",
	})
	got, err := KeptDiff(e.root, from, to)
	if err != nil {
		t.Fatal(err)
	}
	want := []KeptItem{{Name: "Fold", Command: "toy fold"}, {Rule: true, Name: "New", Source: "G-ONE:line-1"}}
	if !slices.Equal(got, want) {
		t.Fatalf("kept diff = %+v, want %+v", got, want)
	}
}
