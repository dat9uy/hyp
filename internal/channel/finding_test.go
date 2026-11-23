package channel

import (
	"regexp"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

var findingIDRe = regexp.MustCompile(`F-[0-9A-Z]{6}`)

// addFinding adds a finding on goal and returns its id.
func addFinding(t *testing.T, goal string) string {
	t.Helper()
	code, out, errs := run(t, "finding", "add", goal, "--fact", "the select reaches an old build\nonly by cherry-pick", "--evidence", "R-1")
	id := findingIDRe.FindString(out)
	if code != 0 || id == "" || !strings.HasPrefix(out, "recorded "+id+" on "+goal+": not folded\n") {
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
	before := testutil.Snapshot(t, r.Dir)
	// any agent, with a role outside the four, adds and folds
	t.Setenv("HYP_ACTOR", "reader-1")
	t.Setenv("HYP_ROLE", "reader")
	f1 := addFinding(t, goal)
	f2 := addFinding(t, goal)

	_, status, _ := run(t, "status", goal)
	for _, want := range []string{"findings:\n", "  " + f1 + "  not folded  by reader-1 at ", "    evidence: R-1\n",
		"    fact: the select reaches an old build\n      only by cherry-pick\n"} {
		if !strings.Contains(status, want) {
			t.Fatalf("status lacks %q:\n%s", want, status)
		}
	}
	if _, all, _ := run(t, "status"); strings.Contains(all, "findings:") {
		t.Fatalf("the list of every Goal prints findings:\n%s", all)
	}

	code, out, errs := run(t, "finding", "fold", f1, f2, "--issue", "1")
	if code != 0 || !strings.Contains(out, "recorded "+f1+" on "+goal+": folded into #1\n") ||
		!strings.Contains(out, "recorded "+f2+" on "+goal+": folded into #1\n") {
		t.Fatalf("fold: code %d, stdout %q, stderr %q", code, out, errs)
	}
	code, out, errs = run(t, "finding", "fold", f1, "--issue", "2")
	if code != 0 || !strings.HasPrefix(out, "recorded "+f1+" on "+goal+": folded into #2\n") ||
		strings.Contains(out, f1+"  folded into #1") {
		t.Fatalf("fold again: code %d, stdout %q, stderr %q", code, out, errs)
	}

	if code, _, errs := run(t, "finding", "fold", f1); code != 1 || !strings.Contains(errs, "refused: Issue: ") {
		t.Fatalf("fold with no issue: code %d, stderr %q", code, errs)
	}
	if code, _, errs := run(t, "finding", "refute", f2); code != 0 {
		t.Fatalf("refute: code %d, stderr %q", code, errs)
	}
	for _, args := range [][]string{{"finding", "fold", f2, "--issue", "3"}, {"finding", "refute", f2}, {"finding", "fold", f1, f2, "--issue", "3"}} {
		code, _, errs := run(t, args...)
		if code != 1 || !strings.Contains(errs, "refused: Not refuted: "+f2+" is refuted") {
			t.Fatalf("%v: code %d, stderr %q", args, code, errs)
		}
	}
	// a refused fold of two findings folds neither
	_, status, _ = run(t, "status", goal)
	if !strings.Contains(status, "  "+f1+"  folded into #2  ") || !strings.Contains(status, "  "+f2+"  refuted  ") {
		t.Fatalf("status after the refusals:\n%s", status)
	}
	if !strings.Contains(status, goal+"  written  ") {
		t.Fatalf("a finding changed the Goal's state:\n%s", status)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("an act on a finding changed the repository's git state")
	}
}

func TestFindingOnAnEndedGoal(t *testing.T) {
	inRepo(t)
	goal := newGoal(t)
	if code, _, errs := run(t, "close", goal); code != 0 {
		t.Fatalf("close: code %d, stderr %q", code, errs)
	}
	f := addFinding(t, goal)
	if code, _, errs := run(t, "finding", "fold", f, "--issue", "4"); code != 0 {
		t.Fatalf("fold on an ended Goal: code %d, stderr %q", code, errs)
	}
	_, status, _ := run(t, "status", goal)
	if !strings.Contains(status, goal+"  ended\n") || !strings.Contains(status, "  "+f+"  folded into #4  ") {
		t.Fatalf("status:\n%s", status)
	}
}

func TestFindingArguments(t *testing.T) {
	inRepo(t)
	goal := newGoal(t)
	for _, c := range []struct {
		args []string
		code int
		errs string
	}{
		{[]string{"finding", "add", "G-NOGOAL", "--fact", "a", "--evidence", "R-1"}, 1, "refused: no Goal G-NOGOAL to add a finding to"},
		{[]string{"finding", "add", goal, "--evidence", "R-1"}, 2, "finding add needs --fact TEXT"},
		{[]string{"finding", "add", goal, "--fact", "a"}, 2, "finding add needs --evidence TEXT"},
		{[]string{"finding", "add", "--fact", "a", "--evidence", "R-1"}, 2, "finding add takes the arguments"},
		{[]string{"finding", "fold", "--issue", "1"}, 2, "finding fold takes the arguments"},
		{[]string{"finding", "refute", "F-1", "F-2"}, 2, "finding refute takes the arguments"},
		{[]string{"finding", "refute", "F-NONE"}, 1, "refused: no finding F-NONE"},
		{[]string{"finding", "fold", "F-NONE", "--issue", "1"}, 1, "refused: no finding F-NONE"},
		{[]string{"finding"}, 2, "finding what?"},
		{[]string{"finding", "drop"}, 2, "finding what?"},
	} {
		code, out, errs := run(t, c.args...)
		if code != c.code || !strings.Contains(errs, c.errs) || strings.Contains(out, "recorded F-") {
			t.Errorf("%v: code %d, stdout %q, stderr %q; want code %d and %q", c.args, code, out, errs, c.code, c.errs)
		}
	}
}
