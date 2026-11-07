package cli

import (
	"bytes"
	"os"
	"path/filepath"
	"runtime/debug"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/recipe"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/store"
)

type result struct {
	code           int
	stdout, stderr string
}

func hyp(t *testing.T, args ...string) result {
	t.Helper()
	var out, errb bytes.Buffer
	code := Run(args, &out, &errb)
	return result{code, out.String(), errb.String()}
}

// must runs a command that must succeed and returns the display id it printed.
func must(t *testing.T, args ...string) string {
	t.Helper()
	r := hyp(t, args...)
	if r.code != 0 {
		t.Fatalf("hyp %v: exit %d, stderr %q", args, r.code, r.stderr)
	}
	return strings.Fields(r.stdout)[0]
}

func want(t *testing.T, r result, code int, substr string) {
	t.Helper()
	if r.code != code || !strings.Contains(r.stdout+r.stderr, substr) {
		t.Fatalf("exit %d (want %d), stdout %q, stderr %q (want %q)", r.code, code, r.stdout, r.stderr, substr)
	}
}

// setup creates a store with a Goal and one DQ and returns their display ids.
func setup(t *testing.T) (dir, g, q string) {
	t.Helper()
	dir = t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	g = must(t, "add", "goal", "--title", "G", "--why", "start")
	q = must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	return
}

func logLines(t *testing.T, dir string) int {
	t.Helper()
	files, _ := filepath.Glob(filepath.Join(dir, ".tree", "events", "*.jsonl"))
	n := 0
	for _, f := range files {
		data, _ := os.ReadFile(f)
		n += bytes.Count(data, []byte("\n"))
	}
	return n
}

func TestInitTwiceFails(t *testing.T) {
	t.Chdir(t.TempDir())
	want(t, hyp(t, "init"), 0, "initialised .tree")
	want(t, hyp(t, "init"), 3, "already exists")
}

func TestMissingWhyIsUsage(t *testing.T) {
	setup(t)
	want(t, hyp(t, "add", "goal", "--title", "G"), 2, "missing --why")
}

func TestMissingActorIsUsage(t *testing.T) {
	setup(t)
	t.Setenv("HYP_ACTOR", "")
	want(t, hyp(t, "add", "goal", "--title", "G", "--why", "w"), 2, "actor")
}

func TestStatusFlagRejected(t *testing.T) {
	_, _, q := setup(t)
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "T", "--why", "w", "--status", "supported"), 2, "status")
}

func TestOverCapTitleRefused(t *testing.T) {
	_, _, q := setup(t)
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", strings.Repeat("ế", 121), "--why", "w"), 1, "refused:")
}

func TestWrongParentRefused(t *testing.T) {
	_, g, _ := setup(t)
	want(t, hyp(t, "add", "hyp", "--parent", g, "--title", "T", "--why", "w"), 1, "refused: wrong parent kind")
}

func TestCrossDQSupersedeRefused(t *testing.T) {
	_, g, q := setup(t)
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H2", "--why", "w")
	want(t, hyp(t, "supersede", h2, "--old", h1, "--why", "w"), 1, "refused: cross-DQ supersedes")
}

func TestUnknownEdgeTypeRefused(t *testing.T) {
	dir, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	before := logLines(t, dir)
	r := hyp(t, "link", h1, "blocks", h2, "--why", "w")
	if r.code != 1 && r.code != 2 {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
	if logLines(t, dir) != before {
		t.Fatal("refused link changed the log")
	}
}

func TestUnknownRefRefused(t *testing.T) {
	setup(t)
	want(t, hyp(t, "abandon", "H-NOPE00", "--why", "w"), 2, "unknown node H-NOPE00")
}

func TestEvidenceOutsideStoreRefused(t *testing.T) {
	dir, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o",
		"--evidence", "notes.md", "--why", "w"), 2, "evidence notes.md not found")
	os.WriteFile(filepath.Join(dir, ".tree", "evidence", "log.txt"), []byte("x"), 0o644)
	want(t, hyp(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o",
		"--evidence", ".tree/evidence/log.txt", "--evidence", "https://example.com/run/1", "--why", "w"), 0, "result.record")
	os.MkdirAll(filepath.Join(dir, ".tree", "evidence", "E-RUN001"), 0o755)
	os.MkdirAll(filepath.Join(dir, "recipes", "boot"), 0o755)
	for _, ev := range []string{".tree/evidence/E-RUN001", "recipes/boot/"} {
		want(t, hyp(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o",
			"--evidence", ev, "--why", "w"), 2, "evidence "+ev+" is a directory; cite a file in it")
	}
	want(t, hyp(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o",
		"--evidence", ".tree/evidence", "--why", "w"), 2, "evidence .tree/evidence not found under .tree/evidence or recipes")
}

func TestHappyPath(t *testing.T) {
	dir, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--alias", "first", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	e := must(t, "add", "exp", "--parent", "first", "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=contradicts",
		"--observation", "H1 holds, H2 fails", "--why", "w"), 0, "R-")
	want(t, hyp(t, "select", q), 0, "suggested: "+h1)
	want(t, hyp(t, "select", q, h1, "--why", "pick the supported one"), 0, q+" dq.select")
	want(t, hyp(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "checked", "--actor", "b"), 0, q+" confirm")
	want(t, hyp(t, "checkpoint", "--summary", "Q answered by H1", "--why", "w"), 0, "checkpoint")

	p, _, err := store.Load(dir)
	if err != nil || len(p.Rejections) > 0 {
		t.Fatalf("load: %v %v", err, p.Rejections)
	}
	for id, n := range p.Nodes {
		if display(p, id) == q && p.Questions[id].State != replay.DQAnswered {
			t.Fatalf("%s (%s) is %+v, want answered", q, n.Title, p.Questions[id])
		}
	}
	if p.Goal.Status != replay.GoalComplete {
		t.Fatalf("goal = %s", p.Goal.Status)
	}
}

func TestSameActorVerifierRefused(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w"), 1, "refused: no current selection")
	must(t, "select", q, h1, "--why", "w")
	want(t, hyp(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w"), 1, "refused: separation of duties")
}

// A single file recipes/<name>.md is not a recipe: a ref only it could answer
// is not found, exit 2, whatever version its frontmatter names.
func TestSingleFileRecipeNotFound(t *testing.T) {
	dir, _, q := setup(t)
	os.MkdirAll(filepath.Join(dir, "recipes"), 0o755)
	os.WriteFile(filepath.Join(dir, "recipes", "boot.md"), []byte("---\nname: boot\nversion: 2\nformat: 1\n---\n"), 0o644)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	for _, ref := range []string{"boot@1", "boot@2"} {
		r := hyp(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", ref, "--decided-by", "reading", "--why", "w")
		want(t, r, 2, "recipe "+ref+" not found; searched ")
	}
}

func TestMissingRequiredTextIsUsage(t *testing.T) {
	_, _, q := setup(t)
	want(t, hyp(t, "add", "hyp", "--parent", q, "--why", "w"), 2, "missing --title")
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "   ", "--why", "w"), 2, "missing --title")
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "T", "--why", "  "), 2, "missing --why")
	want(t, hyp(t, "checkpoint", "--why", "w"), 2, "missing --summary")
}

func TestHelpExitsZeroAndUsagePrintsOnce(t *testing.T) {
	setup(t)
	want(t, hyp(t, "add", "goal", "-h"), 0, "-title")
	r := hyp(t, "add", "goal", "--title", "G", "--why", "w", "--bogus")
	if r.code != 2 || strings.Count(r.stderr, "bogus") != 1 {
		t.Fatalf("exit %d, stderr %q", r.code, r.stderr)
	}
}

// The command is gone; a verifier can confirm again at any time, and an
// insufficient verdict lapses when the Dossier changes.
func TestRequestVerifyIsRemoved(t *testing.T) {
	setup(t)
	if got := hyp(t, "request-verify", "x", "--why", "w").code; got != 2 {
		t.Fatalf("request-verify exit = %d, want 2", got)
	}
	if help := hyp(t, "help").stdout; strings.Contains(help, "request-verify") {
		t.Fatalf("help names request-verify:\n%s", help)
	}
}

func TestAddExpRefusesGoalAsParent(t *testing.T) {
	_, g, _ := setup(t)
	want(t, hyp(t, "add", "exp", "--parent", g, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w"), 1, "refused: wrong parent kind: Experiment under Goal")
}

func TestAddExpRefusesUntargetedLiveRival(t *testing.T) {
	dir, _, q := setup(t)
	ha := must(t, "add", "hyp", "--parent", q, "--title", "A", "--alias", "h-authority", "--why", "w")
	must(t, "add", "hyp", "--parent", q, "--title", "D", "--alias", "h-definitions", "--why", "w")
	hc := must(t, "add", "hyp", "--parent", q, "--title", "C", "--alias", "h-catalog", "--why", "w")
	e1 := must(t, "add", "exp", "--parent", "h-authority", "--title", "e1", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e1, "--verdict", "h-authority=inconclusive", "--observation", "o", "--why", "w")
	e2 := must(t, "add", "exp", "--parent", "h-catalog", "--title", "e2", "--recipe", "none", "--decided-by", "reading", "--target", "h-definitions", "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", "h-catalog=supports", "--verdict", "h-definitions=inconclusive", "--observation", "o", "--why", "w")

	n := logLines(t, dir)
	add := []string{"add", "exp", "--parent", "h-definitions", "--title", "e3", "--recipe", "none", "--decided-by", "reading", "--why", "w"}
	r := hyp(t, add...)
	want(t, r, 1, "refused: live rivals of ")
	want(t, r, 1, "  "+ha+" inconclusive\n  "+hc+" supported\n")
	want(t, r, 1, "--target none to proceed without one")
	if got := logLines(t, dir); got != n {
		t.Fatalf("refused add wrote %d lines", got-n)
	}
	want(t, hyp(t, append(add, "--target", "h-catalog")...), 0, "node.add")
	want(t, hyp(t, append(add, "--target", "none")...), 0, "node.add")
	want(t, hyp(t, append(add, "--target", "none", "--target", "h-catalog")...), 2, "--target none cannot be combined")
}

func TestAddExpDeadSiblingsDoNotRefuse(t *testing.T) {
	_, _, q := setup(t)
	must(t, "add", "hyp", "--parent", q, "--title", "A", "--alias", "ha", "--why", "w")
	must(t, "add", "hyp", "--parent", q, "--title", "B", "--alias", "hb", "--why", "w")
	e := must(t, "add", "exp", "--parent", "hb", "--title", "e", "--recipe", "none", "--decided-by", "reading", "--target", "ha", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", "hb=contradicts", "--observation", "o", "--why", "w")
	want(t, hyp(t, "add", "exp", "--parent", "ha", "--title", "e2", "--recipe", "none", "--decided-by", "reading", "--why", "w"), 0, "node.add")

	must(t, "add", "hyp", "--parent", q, "--title", "C", "--alias", "hc", "--why", "w")
	want(t, hyp(t, "add", "exp", "--parent", "ha", "--title", "e3", "--recipe", "none", "--decided-by", "reading", "--target", "hb", "--why", "w"),
		1, "refused: live rivals of ")
}

func TestAddExpWarnsWhenParentOrTargetIsNotARival(t *testing.T) {
	_, _, q := setup(t)
	ha := must(t, "add", "hyp", "--parent", q, "--title", "A", "--alias", "ha", "--why", "w")
	hb := must(t, "add", "hyp", "--parent", q, "--title", "B", "--alias", "hb", "--why", "w")
	e := must(t, "add", "exp", "--parent", "hb", "--title", "e", "--recipe", "none", "--decided-by", "reading", "--target", "ha", "--why", "w")
	r1 := must(t, "record", "--exp", e, "--verdict", "hb=contradicts", "--observation", "o", "--why", "w")
	addExp := func(parent string, targets ...string) result {
		args := []string{"add", "exp", "--parent", parent, "--title", "x", "--recipe", "none", "--decided-by", "reading", "--why", "w"}
		for _, h := range targets {
			args = append(args, "--target", h)
		}
		return hyp(t, args...)
	}

	r := addExp("hb", "ha")
	want(t, r, 0, "node.add")
	if w := "warn: D1: parent " + hb + " is contradicted by " + r1 + "; if that was wrong, hyp retract <R-id> reopens it; otherwise supersede " + hb + " and test the new Hypothesis\n"; r.stderr != w {
		t.Fatalf("contradicted parent: stderr %q, want %q", r.stderr, w)
	}
	contradictedTarget := addExp("ha", "hb")
	want(t, contradictedTarget, 0, "warn: D1: target "+hb+" is contradicted by "+r1+";")
	if strings.Contains(contradictedTarget.stderr, ha) {
		t.Fatalf("live parent warned: %q", contradictedTarget.stderr)
	}

	hc := must(t, "add", "hyp", "--parent", q, "--title", "C", "--alias", "hc", "--why", "w")
	if r := addExp("ha", "hc"); r.code != 0 || r.stderr != "" {
		t.Fatalf("live parent and target: exit %d, stderr %q", r.code, r.stderr)
	}

	hd := must(t, "add", "hyp", "--parent", q, "--title", "D", "--alias", "hd", "--why", "w")
	must(t, "supersede", "hd", "--old", "hc", "--why", "w")
	if r := addExp("ha", "hc", "hd"); r.code != 0 || r.stderr != "warn: D1: target "+hc+" is superseded by "+hd+"; use "+hd+" instead\n" {
		t.Fatalf("superseded target: exit %d, stderr %q", r.code, r.stderr)
	}
	want(t, addExp("hc", "hd"), 0, "warn: D1: parent "+hc+" is superseded by "+hd+";")

	must(t, "abandon", "ha", "--why", "w")
	want(t, addExp("hd", "ha"), 0, "warn: D1: target "+ha+" is abandoned; it is no longer a rival\n")

	sig := hyp(t, "signals")
	for _, l := range []string{" at add: " + hb + "=contradicted", " at add: " + hc + "=superseded", " at add: " + ha + "=abandoned"} {
		want(t, sig, 0, l)
	}
}

func TestAmendCommand(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "old", "--why", "w")
	want(t, hyp(t, "amend", h, "--body", "new", "--why", "w"), 0, h+" hypothesis.amend")
	want(t, hyp(t, "amend", h, "--body", "new", "--why", "w"), 1, "body is unchanged")
	want(t, hyp(t, "amend", h, "--why", "w"), 2, "missing --body")
}

func TestLimitFlags(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r1 := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "raw SQL org gets no row", "--why", "w")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--closes", "nope", "--why", "w"), 2, "want <R-id>#<n>")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--closes", r1+"#2", "--why", "w"), 1, "bad limit ref")
	must(t, "select", q, h, "--tradeoff", "--why", "w")
	must(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w", "--actor", "v")
	r := hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--why", "w", "--actor", "o")
	want(t, r, 1, "open limits under")
	want(t, r, 1, "raw SQL org gets no row")
	must(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--accept", r1+"#1", "--why", "w", "--actor", "o")
	want(t, hyp(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--accept", r1+"#1", "--why", "w", "--actor", "v2"),
		1, "only a sufficient owner confirm accepts limits")
}

func TestBuildSuffix(t *testing.T) {
	rev := debug.BuildSetting{Key: "vcs.revision", Value: "d8239c7ceddf1be16497d42145073dcd63cf3e88"}
	dirty := debug.BuildSetting{Key: "vcs.modified", Value: "true"}
	clean := debug.BuildSetting{Key: "vcs.modified", Value: "false"}
	for _, c := range []struct {
		settings []debug.BuildSetting
		want     string
	}{
		{nil, ""},
		{[]debug.BuildSetting{rev}, " d8239c7ceddf"},
		{[]debug.BuildSetting{rev, dirty}, " d8239c7ceddf-dirty"},
		{[]debug.BuildSetting{clean}, ""},
	} {
		if got := buildSuffix(c.settings); got != c.want {
			t.Errorf("buildSuffix(%v) = %q, want %q", c.settings, got, c.want)
		}
	}
	// A go test binary carries no VCS stamp, so version prints the bare line.
	want(t, hyp(t, "version"), 0, "hyp 0.1.0 (events v1)")
}

func TestRecordHelpStatesListCaps(t *testing.T) {
	r := hyp(t, "record", "-h")
	cap := "repeatable, at most 10, each at most 300 runes"
	want(t, r, 0, cap)
	if n := strings.Count(r.stdout+r.stderr, cap); n != 2 {
		t.Fatalf("%q occurs %d times in record help, want 2 (--evidence and --limit)", cap, n)
	}
	want(t, r, 0, "repeatable, at most 10, each at most 600 runes")
}

// writeLib writes a recipe.md-only bundle at dir/recipes/<dirName>/.
func writeLib(t *testing.T, dir, dirName, body string) {
	t.Helper()
	p := filepath.Join(dir, "recipes", dirName)
	if err := os.MkdirAll(p, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(p, "recipe.md"), []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

// libraryStore makes a repository (repo/.git) whose library holds boot@1, and a
// store at repo/tree with no recipes of its own; the working directory is the
// store. It returns the repo, the store and the store's DesignQuestion.
func libraryStore(t *testing.T) (repo, tree, q string) {
	t.Helper()
	repo = t.TempDir()
	if err := os.MkdirAll(filepath.Join(repo, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeLib(t, repo, "boot@1", "---\nname: boot\nversion: 1\nformat: 1\n---\n")
	tree = filepath.Join(repo, "tree")
	if err := os.MkdirAll(tree, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Chdir(tree)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	g := must(t, "add", "goal", "--title", "G", "--why", "start")
	q = must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	return repo, tree, q
}

func noQuestionable(t *testing.T, tree string) {
	t.Helper()
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	for id, res := range p.Results {
		if res.Questionable {
			t.Fatalf("%s is questionable", id)
		}
	}
}

func TestExperimentCitesRepoRootLibrary(t *testing.T) {
	repo, tree, q := libraryStore(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	if _, err := os.Stat(filepath.Join(tree, "recipes")); !os.IsNotExist(err) {
		t.Fatalf("the store has its own recipes/ (%v)", err)
	}
	lib, err := recipe.Open(filepath.Join(repo, "recipes", "boot@1"))
	if err != nil {
		t.Fatal(err)
	}
	p, _, err := store.Load(tree)
	if err != nil {
		t.Fatal(err)
	}
	exps := 0
	for _, n := range p.Nodes {
		if n.Kind == event.KindExperiment {
			exps++
			if n.RecipeHash != lib.Hash {
				t.Fatalf("recipe_hash = %s, want the library bundle's %s", n.RecipeHash, lib.Hash)
			}
		}
	}
	if exps != 1 {
		t.Fatalf("%d Experiments, want 1", exps)
	}
	noQuestionable(t, tree)
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")

	// A new version beside the old one leaves the old Result clean.
	writeLib(t, repo, "boot@2", "---\nname: boot\nversion: 2\nformat: 1\n---\n\nAlso check the console.\n")
	noQuestionable(t, tree)
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
	want(t, hyp(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "boot@2", "--decided-by", "reading", "--why", "w"), 0, "node.add")
}

func TestLibraryConflictAndMissingRefRefused(t *testing.T) {
	_, tree, q := libraryStore(t)
	writeLib(t, tree, "boot@1", "---\nname: boot\nversion: 1\nformat: 1\n---\n\nA local edit.\n")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	near := filepath.Join("tree", "recipes", "boot@1") + " and "
	far := filepath.Join("recipes", "boot@1") + " differ"
	r := hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	want(t, r, 1, near)
	want(t, r, 1, far)
	r = hyp(t, "check")
	want(t, r, 0, "warn: recipe boot@1: ")
	want(t, r, 0, far)

	r = hyp(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "nope@1", "--why", "w")
	want(t, r, 2, "recipe nope@1 not found; searched ")
	want(t, r, 2, filepath.Join("tree", "recipes")+", ")
	if n := strings.Count(r.stdout+r.stderr, string(filepath.Separator)+"recipes"); n != 2 {
		t.Fatalf("searched %d directories, want 2: %s%s", n, r.stdout, r.stderr)
	}
}

func TestCheckReportsBundleThatDoesNotOpen(t *testing.T) {
	dir, _, q := setup(t)
	writeLib(t, dir, "boot@1", "---\nname: boot\nversion: 1\nformat: 1\n---\n\n## Launch\n\nRun: ./launch.sh\n")
	script := filepath.Join(dir, "recipes", "boot@1", "launch.sh")
	if err := os.WriteFile(script, []byte("#!/bin/sh\n"), 0o755); err != nil {
		t.Fatal(err)
	}
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "boot@1", "--decided-by", "reading", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")

	if err := os.Chmod(script, 0o644); err != nil {
		t.Fatal(err)
	}
	r := hyp(t, "check")
	want(t, r, 0, filepath.Join("recipes", "boot@1")+" does not open: ")
	want(t, r, 0, "is not executable (chmod +x")
	// Hashing never parses, so the pin still verifies.
	want(t, r, 0, "its pins still verify")
	if strings.Contains(r.stdout, "questionable evidence") {
		t.Fatalf("stdout %q: a bundle that hashes the same made its Result questionable", r.stdout)
	}
}

// An Experiment with no measure, --recipe none or a pinned recipe that is not
// an instrument, is written only when it says its verdicts are decided by
// reading; a pinned instrument needs no declaration and refuses one.
func TestAddExpDeclaration(t *testing.T) {
	dir, _, q := setup(t)
	bundleIn(t, dir, "env", "## Launch\n\nRun: ./up.sh\n", map[string]string{"up.sh": "#!/bin/sh\n"})
	bundleIn(t, dir, "tool", "## Standalone\n\nRun ok: ./ok.sh\n", map[string]string{
		"ok.sh":                        "#!/bin/sh\necho ok\n",
		"known-answers/ok/expect.json": `{"step": "standalone", "command": "ok", "exit": 0, "stdout": "ok\n"}`,
		"known-answers/ok/input/a":     "",
	})
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	addExp := func(extra ...string) result {
		t.Helper()
		return hyp(t, append([]string{"add", "exp", "--parent", h, "--title", "E", "--why", "w"}, extra...)...)
	}
	const path = "(RECIPE-FORMAT, Verdict scripts)"

	before := logLines(t, dir)
	want(t, addExp("--recipe", "none"), 1, "refused: recipe none holds no measure")
	want(t, addExp("--recipe", "none"), 1, path)
	want(t, addExp("--recipe", "env@1"), 1, "refused: recipe env@1 holds no measure")
	want(t, addExp("--recipe", "env@1"), 1, path)
	if n := logLines(t, dir); n != before {
		t.Fatalf("refused add exp wrote %d lines", n-before)
	}

	read := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, addExp("--recipe", "env@1", "--decided-by", "reading"), 0, "node.add")
	measured := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "tool@1", "--why", "w")

	want(t, addExp("--recipe", "tool@1", "--decided-by", "reading"), 2, "--decided-by reading contradicts recipe tool@1")
	want(t, addExp("--recipe", "none", "--decided-by", "measure"), 2, `invalid --decided-by "measure"`)

	want(t, hyp(t, "show", read), 0, "decided by: reading")
	if r := hyp(t, "show", measured); strings.Contains(r.stdout, "decided by") {
		t.Fatalf("show of a measured Experiment prints a declaration:\n%s", r.stdout)
	}
}

// hyp withdraw writes dq.withdraw for a DesignQuestion and
// experiment.withdraw for an Experiment; a Hypothesis is the wrong kind.
func TestWithdrawDispatchesByKind(t *testing.T) {
	_, g, q := setup(t)
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "withdraw", q2, "--why", "moot"), 0, q2+" dq.withdraw")
	want(t, hyp(t, "withdraw", e, "--why", "will not run"), 0, e+" experiment.withdraw")
	want(t, hyp(t, "withdraw", e, "--why", "again"), 1, "refused: already withdrawn")
	want(t, hyp(t, "withdraw", h, "--why", "w"), 1, "refused: wrong kind")
	want(t, hyp(t, "show", e), 0, "status: withdrawn")
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w"), 1,
		"refused: "+e+" is withdrawn; add a new Experiment for this Result")
}

// hyp supersede joins two Experiments under one question: the old one reads
// superseded, leaves hyp status, shows its replacement, and refuses a Result.
func TestSupersedeExperiment(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	old := must(t, "add", "exp", "--parent", h, "--title", "Old run", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	repl := must(t, "add", "exp", "--parent", h, "--title", "New run", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "supersede", repl, "--old", old, "--why", "replaces it"), 0, repl+" edge.add")
	want(t, hyp(t, "show", old), 0, "status: superseded (replaced by "+repl+")")
	want(t, hyp(t, "show", q), 0, "  "+old+" tests "+h+"\n    judged: superseded\n")
	st := hyp(t, "status")
	if strings.Contains(st.stdout, old) || !strings.Contains(st.stdout, repl+" planned New run") {
		t.Fatalf("status lists %s or not %s as planned:\n%s", old, repl, st.stdout)
	}
	want(t, hyp(t, "record", "--exp", old, "--verdict", h+"=supports", "--observation", "o", "--why", "w"), 1,
		"refused: "+old+" is superseded by "+repl)
	want(t, hyp(t, "record", "--exp", repl, "--verdict", h+"=supports", "--observation", "o", "--why", "w"), 0, "result.record")
	want(t, hyp(t, "supersede", old, "--old", repl, "--why", "w"), 1, "has the non-retracted Result")
	want(t, hyp(t, "supersede", h, "--old", old, "--why", "w"), 1, "refused: wrong kind: supersedes joins two Hypotheses or two Experiments")
}

// A Result from another worktree applied under a withdrawn Experiment makes
// it read run, and hyp check names it.
func TestCheckNamesRunExperimentClosedInMergedLog(t *testing.T) {
	dir, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	must(t, "withdraw", e, "--why", "will not run")
	p, _, err := store.Load(dir)
	if err != nil {
		t.Fatal(err)
	}
	var eid, hid string
	for id, n := range p.Nodes {
		switch n.Kind {
		case event.KindExperiment:
			eid = id
		case event.KindHypothesis:
			hid = id
		}
	}
	line := `{"v":1,"id":"7ZZZZZZZZZZZZZZZZZZZZZZZZZ","ts":"2026-11-01T00:00:00Z","actor":"b","type":"result.record","node":"7ZZZZZZZZZZZZZZZZZZZZZZZZZ",` +
		`"payload":{"parent":"` + eid + `","verdicts":{"` + hid + `":"supports"},"observation":"o"},"why":"from b"}` + "\n"
	if err := os.WriteFile(filepath.Join(dir, ".tree", "events", "b.jsonl"), []byte(line), 0o644); err != nil {
		t.Fatal(err)
	}
	want(t, hyp(t, "show", e), 0, "status: run")
	want(t, hyp(t, "check"), 0, "warn: "+e+" is withdrawn but reads run: its Result R-ZZZZZZ applied after the close")
}

// No current selection outranks a bad --accept, and a bad --accept outranks the
// limit gate, in the order the refusals have always had.
func TestConfirmRefusalOrder(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--observation", "o", "--limit", "one host", "--why", "w")
	owner := []string{"confirm", q, "--role", "owner", "--verdict", "sufficient", "--why", "w"}
	bad := append(append([]string{}, owner...), "--accept", "nope")
	want(t, hyp(t, bad...), 1, "refused: no current selection")
	must(t, "select", q, h1, "--why", "w")
	want(t, hyp(t, bad...), 2, "nope")
	want(t, hyp(t, owner...), 1, "open limits under")
}

func TestSelectWithNoHypothesisReadsCandidates(t *testing.T) {
	supported := func(q, title string) string {
		h := must(t, "add", "hyp", "--parent", q, "--title", title, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
		must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
		return h
	}
	// stalePremise makes h stale by contradicting a target it depends on.
	stalePremise := func(g, h string) {
		q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
		tgt := supported(q2, "T")
		must(t, "link", h, "depends_on", tgt, "--why", "w")
		e := must(t, "add", "exp", "--parent", tgt, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
		must(t, "record", "--exp", e, "--verdict", tgt+"=contradicts", "--observation", "o", "--why", "w")
	}
	t.Run("lone ready", func(t *testing.T) {
		_, _, q := setup(t)
		h := supported(q, "H")
		want(t, hyp(t, "select", q), 0, "suggested: "+h)
	})
	t.Run("lone stale", func(t *testing.T) {
		_, g, q := setup(t)
		stalePremise(g, supported(q, "H"))
		want(t, hyp(t, "select", q), 0, "suggested: none (")
	})
	t.Run("lone with an open sub-question", func(t *testing.T) {
		_, _, q := setup(t)
		h := supported(q, "H")
		must(t, "add", "dq", "--parent", h, "--title", "Sub", "--why", "w")
		want(t, hyp(t, "select", q), 0, "suggested: none (")
	})
	t.Run("lone with an open sub-question names it, then is suggested once settled", func(t *testing.T) {
		_, _, q := setup(t)
		h := supported(q, "H")
		sub := must(t, "add", "dq", "--parent", h, "--title", "Sub", "--why", "w")
		want(t, hyp(t, "select", q), 0, "suggested: none (awaiting_subquestions)")
		must(t, "withdraw", sub, "--why", "w")
		want(t, hyp(t, "select", q), 0, "suggested: "+h)
	})
	t.Run("two with one stale", func(t *testing.T) {
		_, g, q := setup(t)
		supported(q, "H1")
		stalePremise(g, supported(q, "H2"))
		want(t, hyp(t, "select", q), 0, "trade-off: owner decides between")
	})
	t.Run("two with one waiting", func(t *testing.T) {
		_, _, q := setup(t)
		supported(q, "H1")
		h2 := supported(q, "H2")
		must(t, "add", "dq", "--parent", h2, "--title", "Sub", "--why", "w")
		want(t, hyp(t, "select", q), 0, "trade-off: owner decides between")
	})
}

func TestGoalBodyHoldsLongerTextThanOtherNodes(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	want(t, hyp(t, "add", "goal", "--title", "G", "--body", strings.Repeat("x", 2001), "--why", "w"), 1, "body")
	g := must(t, "add", "goal", "--title", "G", "--body", strings.Repeat("x", 2000), "--why", "w")
	must(t, "add", "dq", "--parent", g, "--title", "Q", "--body", strings.Repeat("x", 600), "--why", "w")
	want(t, hyp(t, "add", "dq", "--parent", g, "--title", "Q", "--body", strings.Repeat("x", 601), "--why", "w"), 1, "body")
}
