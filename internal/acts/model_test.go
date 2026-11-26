package acts

import (
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
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

// selectModel commits files on a leaf from H0, runs line-1 on it and selects.
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
	tree, err := gitread.RowTree(e.root, commit)
	if err != nil {
		t.Fatal(err)
	}
	_, err = Select(e.selectOpts(tree, "runner-1", "runner"))
	return err
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
	holds := HeldBack(g, e.root)
	if len(holds) != 1 {
		t.Fatalf("holds = %v", holds)
	}
	want, _ := fault.As(selErr)
	got, _ := fault.As(holds[0].Err)
	if want == nil || got == nil || got.Guard != want.Guard || got.Code != want.Code || got.Code != fault.RuleNoVariant {
		t.Fatalf("held back by %v, select refused with %v", holds[0].Err, selErr)
	}
}
