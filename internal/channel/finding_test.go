package channel

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/testutil"
)

var findingIDRe = regexp.MustCompile(`F-[0-9A-Z]{6}`)

// addFinding adds a finding, which names no Goal, and returns its id.
func addFinding(t *testing.T, extra ...string) string {
	t.Helper()
	args := append([]string{"finding", "add", "--fact", "the select reaches an old build\nonly by cherry-pick", "--evidence", "R-1"}, extra...)
	code, out, errs := run(t, args...)
	id := findingIDRe.FindString(out)
	if code != 0 || id == "" || out != "recorded "+id+"  not folded\n" {
		t.Fatalf("finding add: code %d, stdout %q, stderr %q", code, out, errs)
	}
	return id
}

// newGoal adds a written Goal as the owner and returns its id.
func newGoal(t *testing.T) string {
	t.Helper()
	code, out, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "Find it.", "--line", "gap:it fails")
	fields := strings.Fields(out)
	if code != 0 || len(fields) < 2 {
		t.Fatalf("add goal: code %d, stdout %q, stderr %q", code, out, errs)
	}
	return fields[1]
}

func TestFindingAddFoldRefute(t *testing.T) {
	r := inRepo(t)
	keepStoreOut(t, r.Dir)
	goal := newGoal(t)
	other := newGoal(t)
	before := testutil.Snapshot(t, r.Dir)
	// any agent, with a role outside the four, adds and folds
	t.Setenv("HYP_ACTOR", "reader-1")
	t.Setenv("HYP_ROLE", "reader")
	f1 := addFinding(t)
	f2 := addFinding(t)

	_, list, _ := run(t, "finding", "list")
	for _, want := range []string{f1 + "  not folded  by reader-1 at ", "  evidence: R-1\n",
		"  fact: the select reaches an old build\n    only by cherry-pick\n",
		"    hyp finding fold " + f1 + " --goal G\n", "    hyp finding refute " + f2 + "\n",
		"record a new finding:\n  hyp finding add --fact TEXT --evidence TEXT"} {
		if !strings.Contains(list, want) {
			t.Fatalf("finding list lacks %q:\n%s", want, list)
		}
	}
	if _, all, _ := run(t, "status"); strings.Contains(all, "findings:") || !strings.Contains(all, "2 not folded: hyp finding list --not-folded\n") {
		t.Fatalf("the list of every Goal prints findings, or no count:\n%s", all)
	}

	code, out, errs := run(t, "finding", "fold", f1, f2, "--goal", goal)
	if code != 0 || !strings.Contains(out, "recorded "+f1+"  folded into "+goal+"\n") ||
		!strings.Contains(out, "recorded "+f2+"  folded into "+goal+"\n") || !strings.Contains(out, goal+"  written  ") {
		t.Fatalf("fold: code %d, stdout %q, stderr %q", code, out, errs)
	}
	code, out, errs = run(t, "finding", "fold", f1, "--goal", other)
	if code != 0 || !strings.HasPrefix(out, "recorded "+f1+"  folded into "+other+"\n") ||
		strings.Contains(out, f1+"  folded into "+goal) {
		t.Fatalf("fold again: code %d, stdout %q, stderr %q", code, out, errs)
	}
	// a folded finding may be folded again, and the fold line names no Goal
	_, list, _ = run(t, "finding", "list")
	if !strings.Contains(list, "    hyp finding fold "+f1+" --goal G\n") || strings.Contains(list, "--goal "+goal) ||
		strings.Contains(list, "--goal "+other) {
		t.Fatalf("finding list lacks the fold line of a folded finding, or names a Goal:\n%s", list)
	}
	_, status, _ := run(t, "status", goal)
	if strings.Contains(status, f1) || !strings.Contains(status, "findings:\n  "+f2+"  folded into "+goal+"  by reader-1 at ") {
		t.Fatalf("status G lists a finding folded into another Goal, or lacks its own:\n%s", status)
	}

	if code, _, errs := run(t, "finding", "fold", f1); code != 2 || !strings.Contains(errs, "finding fold needs --goal G") {
		t.Fatalf("fold with no Goal: code %d, stderr %q", code, errs)
	}
	if code, _, errs := run(t, "finding", "refute", f2); code != 0 {
		t.Fatalf("refute: code %d, stderr %q", code, errs)
	}
	for _, args := range [][]string{{"finding", "fold", f2, "--goal", goal}, {"finding", "refute", f2}, {"finding", "fold", f1, f2, "--goal", goal}} {
		code, _, errs := run(t, args...)
		if code != 1 || !strings.Contains(errs, "refused: Not refuted: "+f2+" is refuted") {
			t.Fatalf("%v: code %d, stderr %q", args, code, errs)
		}
	}
	// a refused fold of two findings folds neither
	_, list, _ = run(t, "finding", "list")
	if !strings.Contains(list, f1+"  folded into "+other+"  ") || !strings.Contains(list, f2+"  refuted  ") ||
		strings.Contains(list, "hyp finding refute "+f2) {
		t.Fatalf("finding list after the refusals:\n%s", list)
	}
	if _, nf, _ := run(t, "finding", "list", "--not-folded"); strings.Contains(nf, f1) || !strings.Contains(nf, "no finding is not folded.\n") {
		t.Fatalf("finding list --not-folded:\n%s", nf)
	}
	_, status, _ = run(t, "status", goal)
	if !strings.Contains(status, goal+"  written  ") {
		t.Fatalf("a finding changed the Goal's state:\n%s", status)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("an act on a finding changed the repository's git state")
	}
}

func TestFindingFoldedIntoAnEndedGoal(t *testing.T) {
	inRepo(t)
	goal := newGoal(t)
	other := newGoal(t)
	f := addFinding(t)
	if code, _, errs := run(t, "finding", "fold", f, "--goal", goal); code != 0 {
		t.Fatalf("fold: code %d, stderr %q", code, errs)
	}
	if code, _, errs := run(t, "close", goal); code != 0 {
		t.Fatalf("close: code %d, stderr %q", code, errs)
	}
	_, list, _ := run(t, "finding", "list")
	if strings.Contains(list, f) || !strings.Contains(list, "1 finding folded into an ended Goal not shown; hyp finding list --all lists them.\n") {
		t.Fatalf("finding list shows a finding folded into an ended Goal:\n%s", list)
	}
	if _, all, _ := run(t, "finding", "list", "--all"); !strings.Contains(all, f+"  folded into "+goal+", ended  ") {
		t.Fatalf("finding list --all:\n%s", all)
	}
	code, _, errs := run(t, "finding", "fold", f, "--goal", goal)
	if code != 1 || !strings.Contains(errs, "refused: In work: "+goal+" is ended") {
		t.Fatalf("fold into an ended Goal: code %d, stderr %q", code, errs)
	}
	if code, out, errs := run(t, "finding", "fold", f, "--goal", other); code != 0 || !strings.HasPrefix(out, "recorded "+f+"  folded into "+other+"\n") {
		t.Fatalf("fold again into a Goal in work: code %d, stdout %q, stderr %q", code, out, errs)
	}
}

func TestFindingDraft(t *testing.T) {
	inRepo(t)
	newGoal(t)
	draft := "hyp add goal --base HEAD --words 'Cherry-pick old builds.' --line 'gap:select reaches an old build.'"
	f := addFinding(t, "--draft", draft)
	_, list, _ := run(t, "finding", "list")
	if !strings.Contains(list, f+"  not folded  ") || !strings.Contains(list, "  draft by the agent, which the owner may edit and run:\n    "+draft+"\n") {
		t.Fatalf("finding list lacks the draft:\n%s", list)
	}
	code, _, errs := run(t, "finding", "add", "--fact", "a", "--evidence", "b", "--draft", "rm -rf /")
	if code != 2 || !strings.Contains(errs, "--draft is an add goal command line") {
		t.Fatalf("a draft that is no add goal: code %d, stderr %q", code, errs)
	}
}

// An old-form finding, recorded on a Goal, still replays: it reads not
// folded until a fold names it.
func TestFindingOldForm(t *testing.T) {
	r := inRepo(t)
	goal := newGoal(t)
	data, err := os.ReadFile(filepath.Join(r.Dir, ".tree", "events", "owner-1.jsonl"))
	if err != nil {
		t.Fatal(err)
	}
	m := regexp.MustCompile(`"id":"([0-9A-Z]{26})"`).FindSubmatch(data)
	if m == nil {
		t.Fatalf("no goal.add id in %s", data)
	}
	gid := string(m[1])
	root, err := logstore.TopLevel(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	trace, err := logstore.TraceOf(root)
	if err != nil {
		t.Fatal(err)
	}
	ts := gid[:10]
	lines := `{"v":2,"id":"` + ts + `ZZZZZZZZZZZZZZZY","ts":"2026-11-24T02:54:20Z","actor":"old","type":"finding.add","node":"` + gid + `","payload":{"role":"","fact":"old one","evidence":"e1"},"why":""}
{"v":2,"id":"` + ts + `ZZZZZZZZZZZZZZZZ","ts":"2026-11-24T02:54:21Z","actor":"old","type":"finding.add","node":"` + gid + `","payload":{"role":"","fact":"old two","evidence":"e2"},"why":""}
`
	if err := os.MkdirAll(filepath.Join(trace.Dir, "events"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trace.Dir, "events", "old.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	_, list, _ := run(t, "finding", "list", "--not-folded")
	for _, want := range []string{"F-ZZZZZY  not folded  by old at ", "  fact: old one\n", "F-ZZZZZZ  not folded  "} {
		if !strings.Contains(list, want) {
			t.Fatalf("finding list --not-folded lacks %q:\n%s", want, list)
		}
	}
	if code, out, errs := run(t, "finding", "fold", "F-ZZZZZY", "--goal", goal); code != 0 || !strings.HasPrefix(out, "recorded F-ZZZZZY  folded into "+goal+"\n") {
		t.Fatalf("fold of an old-form finding: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, status, _ := run(t, "status"); !strings.Contains(status, "1 not folded: hyp finding list --not-folded\n") || strings.Contains(status, "refused at replay") {
		t.Fatalf("status:\n%s", status)
	}
}

func TestFindingArguments(t *testing.T) {
	inRepo(t)
	goal := newGoal(t)
	if code, _, errs := run(t, "close", goal); code != 0 {
		t.Fatalf("close: code %d, stderr %q", code, errs)
	}
	f := addFinding(t)
	for _, c := range []struct {
		args []string
		code int
		errs string
	}{
		{[]string{"finding", "add", "--evidence", "R-1"}, 2, "finding add needs --fact TEXT"},
		{[]string{"finding", "add", "--fact", "a"}, 2, "finding add needs --evidence TEXT"},
		{[]string{"finding", "add", goal, "--fact", "a", "--evidence", "R-1"}, 2, "finding add takes the arguments"},
		{[]string{"finding", "fold", "--goal", goal}, 2, "finding fold takes the arguments"},
		{[]string{"finding", "fold", f, "--issue", "1"}, 2, "flag provided but not defined: -issue"},
		{[]string{"finding", "fold", f, "--goal", "G-NOGOAL"}, 1, "refused: In work: no Goal G-NOGOAL to fold into"},
		{[]string{"finding", "fold", f, "--goal", goal}, 1, "refused: In work: " + goal + " is ended"},
		{[]string{"finding", "refute", "F-1", "F-2"}, 2, "finding refute takes the arguments"},
		{[]string{"finding", "refute", "F-NONE"}, 1, "refused: no finding F-NONE"},
		{[]string{"finding", "list", "F-1"}, 2, "finding list takes the arguments"},
		{[]string{"finding"}, 2, "finding what?"},
		{[]string{"finding", "drop"}, 2, "finding what?"},
	} {
		code, out, errs := run(t, c.args...)
		if code != c.code || !strings.Contains(errs, c.errs) || strings.Contains(out, "recorded F-") {
			t.Errorf("%v: code %d, stdout %q, stderr %q; want code %d and %q", c.args, code, out, errs, c.code, c.errs)
		}
	}
}
