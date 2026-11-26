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

// namingGoal adds a written Goal that names the findings, and returns its
// id.
func namingGoal(t *testing.T, findings ...string) string {
	t.Helper()
	args := []string{"add", "goal", "--base", "HEAD", "--words", "Find it.", "--line", "gap:it fails"}
	for _, f := range findings {
		args = append(args, "--finding", f)
	}
	code, out, errs := run(t, args...)
	fields := strings.Fields(out)
	if code != 0 || len(fields) < 2 || !strings.Contains(out, "findings: "+strings.Join(findings, " ")+"\n") {
		t.Fatalf("add goal: code %d, stdout %q, stderr %q", code, out, errs)
	}
	return fields[1]
}

func TestFindingAddNameRefute(t *testing.T) {
	r := inRepo(t)
	keepStoreOut(t, r.Dir)
	newGoal(t)
	// any agent, with a role outside the four, adds and refutes
	t.Setenv("HYP_ACTOR", "reader-1")
	t.Setenv("HYP_ROLE", "reader")
	f1 := addFinding(t)
	f2 := addFinding(t)
	f3 := addFinding(t)

	_, list, _ := run(t, "finding", "list")
	for _, want := range []string{f1 + "  not folded  by reader-1 at ", "  evidence: R-1\n",
		"  fact: the select reaches an old build\n    only by cherry-pick\n",
		"    hyp finding refute " + f2 + "\n",
		"record a new finding:\n  hyp finding add --fact TEXT --evidence TEXT"} {
		if !strings.Contains(list, want) {
			t.Fatalf("finding list lacks %q:\n%s", want, list)
		}
	}
	if strings.Contains(list, "finding fold") {
		t.Fatalf("finding list names a fold:\n%s", list)
	}
	if _, all, _ := run(t, "status"); strings.Contains(all, "findings:") || !strings.Contains(all, "3 not folded: hyp finding list --not-folded\n") {
		t.Fatalf("the list of every Goal prints findings, or no count:\n%s", all)
	}

	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	goal := namingGoal(t, f1, f2)
	_, status, _ := run(t, "status", goal)
	for _, f := range []string{f1, f2} {
		if !strings.Contains(status, "  "+f+"  folded into "+goal+"  by reader-1 at ") {
			t.Fatalf("status G lacks %s:\n%s", f, status)
		}
	}
	if !strings.Contains(status, "findings:\n") || !strings.Contains(status, "    fact: the select reaches an old build\n      only by cherry-pick\n") {
		t.Fatalf("status G lacks the facts:\n%s", status)
	}
	if code, out, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "Again.", "--line", "gap:x", "--finding", f1); code != 1 || out != "" ||
		!strings.Contains(errs, "refused: One in work: "+f1+" is named by "+goal+", a Goal in work") {
		t.Fatalf("a second Goal in work names f1: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "None.", "--line", "gap:x", "--finding", "F-NONE00"); code != 1 ||
		!strings.Contains(errs, "refused: no finding F-NONE00") {
		t.Fatalf("a Goal names a finding the trace lacks: code %d, stderr %q", code, errs)
	}

	t.Setenv("HYP_ACTOR", "reader-1")
	if code, _, errs := run(t, "finding", "refute", f3); code != 0 {
		t.Fatalf("refute: code %d, stderr %q", code, errs)
	}
	if code, _, errs := run(t, "finding", "refute", f3); code != 1 || !strings.Contains(errs, "refused: Not refuted: "+f3+" is refuted") {
		t.Fatalf("refute again: code %d, stderr %q", code, errs)
	}
	t.Setenv("HYP_ACTOR", "owner-1")
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "Refuted.", "--line", "gap:x", "--finding", f3); code != 1 ||
		!strings.Contains(errs, "refused: Not refuted: "+f3+" is refuted") {
		t.Fatalf("a Goal names a refuted finding: code %d, stderr %q", code, errs)
	}
	if _, nf, _ := run(t, "finding", "list", "--not-folded"); strings.Contains(nf, f1) || !strings.Contains(nf, "no finding is not folded.\n") {
		t.Fatalf("finding list --not-folded:\n%s", nf)
	}
	if _, status, _ := run(t, "status", goal); !strings.Contains(status, goal+"  written  ") {
		t.Fatalf("a finding changed the Goal's state:\n%s", status)
	}
}

// The acts on a finding change no git state: their events are in the trace.
func TestFindingActsKeepGitState(t *testing.T) {
	r := inRepo(t)
	keepStoreOut(t, r.Dir)
	newGoal(t)
	before := testutil.Snapshot(t, r.Dir)
	f := addFinding(t)
	if code, _, errs := run(t, "finding", "refute", f); code != 0 {
		t.Fatalf("refute: code %d, stderr %q", code, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("an act on a finding changed the repository's git state")
	}
}

// A Goal's names are in git: a clone whose trace lacks the finding prints its
// id, and says the trace does not hold it.
func TestFindingNamedOnAnotherClone(t *testing.T) {
	r := inRepo(t)
	newGoal(t)
	f := addFinding(t)
	goal := namingGoal(t, f)
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "goal")
	clone := filepath.Join(t.TempDir(), "clone")
	testutil.Git(t, r.Dir, "clone", "-q", r.Dir, clone)
	t.Chdir(clone)
	_, status, _ := run(t, "status", goal)
	if !strings.Contains(status, "findings:\n  "+f+"  not in this clone's trace\n") || strings.Contains(status, "cherry-pick") {
		t.Fatalf("status G on another clone:\n%s", status)
	}
}

func TestFindingNamedByAnEndedGoal(t *testing.T) {
	inRepo(t)
	newGoal(t)
	f := addFinding(t)
	addFinding(t) // a finding no Goal names, which waits
	goal := namingGoal(t, f)
	if code, _, errs := run(t, "close", goal); code != 0 {
		t.Fatalf("close: code %d, stderr %q", code, errs)
	}
	_, list, _ := run(t, "finding", "list")
	if strings.Contains(list, f+"  ") || !strings.Contains(list, "1 finding settled or ended not shown; hyp finding list --all lists them.\n") {
		t.Fatalf("finding list shows a finding of an ended Goal:\n%s", list)
	}
	if _, all, _ := run(t, "finding", "list", "--all"); !strings.Contains(all, f+"  ended in "+goal+", not settled  ") {
		t.Fatalf("finding list --all:\n%s", all)
	}
	if _, status, _ := run(t, "status"); !strings.Contains(status, "1 not folded: hyp finding list") {
		t.Fatalf("status counts an ended finding as waiting:\n%s", status)
	}
	if code, _, errs := run(t, "finding", "decline", f, "--why", "not wanted"); code != 1 ||
		!strings.Contains(errs, "refused: Not folded: "+f+" is named by "+goal) {
		t.Fatalf("decline of a finding an ended Goal named: code %d, stderr %q", code, errs)
	}
	// a later Goal may name it
	later := namingGoal(t, f)
	if _, all, _ := run(t, "finding", "list"); !strings.Contains(all, f+"  folded into "+later+"  ") {
		t.Fatalf("finding list after a later Goal names it:\n%s", all)
	}
}

func TestFindingDecline(t *testing.T) {
	inRepo(t)
	newGoal(t)
	f1 := addFinding(t)
	f2 := addFinding(t)
	held := addFinding(t)
	goal := namingGoal(t, held)

	if code, out, errs := run(t, "finding", "decline", f1); code != 1 || out != "" ||
		!strings.Contains(errs, "refused: Owner declines: a decline needs --why TEXT") {
		t.Fatalf("decline with no why: code %d, stdout %q, stderr %q", code, out, errs)
	}
	t.Setenv("HYP_ROLE", "runner")
	if code, _, errs := run(t, "finding", "decline", f1, "--why", "not wanted"); code != 1 ||
		!strings.Contains(errs, "refused: Owner declines: only the owner declines a finding; the role is runner") {
		t.Fatalf("decline by a runner: code %d, stderr %q", code, errs)
	}
	t.Setenv("HYP_ROLE", "owner")
	if code, _, errs := run(t, "finding", "decline", held, "--why", "not wanted"); code != 1 ||
		!strings.Contains(errs, "refused: Not folded: "+held+" is named by "+goal) {
		t.Fatalf("decline of a named finding: code %d, stderr %q", code, errs)
	}
	if code, out, errs := run(t, "finding", "decline", f1, "--why", "not wanted"); code != 0 || out != "recorded "+f1+"  declined\n" {
		t.Fatalf("decline: code %d, stdout %q, stderr %q", code, out, errs)
	}
	for _, args := range [][]string{{"add", "goal", "--base", "HEAD", "--words", "D.", "--line", "gap:x", "--finding", f1},
		{"finding", "refute", f1}, {"finding", "decline", f1, "--why", "still not wanted"}} {
		code, _, errs := run(t, args...)
		if code != 1 || !strings.Contains(errs, "refused: Not declined: "+f1+" is declined") {
			t.Fatalf("%v: code %d, stderr %q", args, code, errs)
		}
	}
	if code, _, errs := run(t, "finding", "refute", f2); code != 0 {
		t.Fatalf("refute: code %d, stderr %q", code, errs)
	}
	if code, _, errs := run(t, "finding", "decline", f2, "--why", "not wanted"); code != 1 ||
		!strings.Contains(errs, "refused: Not refuted: "+f2+" is refuted") {
		t.Fatalf("decline of a refuted finding: code %d, stderr %q", code, errs)
	}

	_, list, _ := run(t, "finding", "list")
	if strings.Contains(list, f1) || !strings.Contains(list, "1 finding declined not shown; hyp finding list --all lists them.\n") {
		t.Fatalf("finding list shows a declined finding:\n%s", list)
	}
	_, all, _ := run(t, "finding", "list", "--all")
	if !strings.Contains(all, f1+"  declined  by ") || !strings.Contains(all, "  why: not wanted\n") ||
		strings.Contains(all, "hyp finding refute "+f1) {
		t.Fatalf("finding list --all:\n%s", all)
	}
	if _, status, _ := run(t, "status"); !strings.Contains(status, "\n0 not folded: hyp finding list") {
		t.Fatalf("status counts a declined finding as not folded:\n%s", status)
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
// folded until a Goal names it. An old fold event counts as a name by its
// Goal.
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
	lines := `{"v":2,"id":"` + ts + `ZZZZZZZZZZZZZZZX","ts":"2026-11-24T02:54:20Z","actor":"old","type":"finding.add","node":"` + gid + `","payload":{"role":"","fact":"old one","evidence":"e1"},"why":""}
{"v":2,"id":"` + ts + `ZZZZZZZZZZZZZZZY","ts":"2026-11-24T02:54:21Z","actor":"old","type":"finding.add","node":"` + gid + `","payload":{"role":"","fact":"old two","evidence":"e2"},"why":""}
`
	if err := os.MkdirAll(filepath.Join(trace.Dir, "events"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(trace.Dir, "events", "old.jsonl"), []byte(lines), 0o644); err != nil {
		t.Fatal(err)
	}
	_, list, _ := run(t, "finding", "list", "--not-folded")
	for _, want := range []string{"F-ZZZZZX  not folded  by old at ", "  fact: old one\n", "F-ZZZZZY  not folded  "} {
		if !strings.Contains(list, want) {
			t.Fatalf("finding list --not-folded lacks %q:\n%s", want, list)
		}
	}
	fold := `{"v":2,"id":"` + ts + `ZZZZZZZZZZZZZZZZ","ts":"2026-11-24T02:54:22Z","actor":"old","type":"finding.fold","node":"` + ts + `ZZZZZZZZZZZZZZZZ","payload":{"role":"","findings":["` + ts + `ZZZZZZZZZZZZZZZX"],"goal":"` + gid + `"},"why":""}
`
	f, err := os.OpenFile(filepath.Join(trace.Dir, "events", "old.jsonl"), os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := f.WriteString(fold); err != nil {
		t.Fatal(err)
	}
	f.Close()
	if _, status, _ := run(t, "status", goal); !strings.Contains(status, "findings:\n  F-ZZZZZX  folded into "+goal+"  by old at ") {
		t.Fatalf("status G lacks the finding an old fold folded into it:\n%s", status)
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
		{[]string{"finding", "fold", f, "--goal", goal}, 2, "finding what?"},
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
