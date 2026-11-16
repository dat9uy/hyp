package cli

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// Two commands give one answer about a fact: hyp status and hyp show
// decide "separated" by one rule. Three candidates where one Result judges
// two of them leaves two pairs that no Result tells apart; show lists them
// and the choose line under AWAITING OWNER names the same pairs.
func TestStatusAndShowAgreeOnSeparation(t *testing.T) {
	_, _, q := setup(t)
	a := must(t, "add", "hyp", "--parent", q, "--title", "A", "--why", "w")
	b := must(t, "add", "hyp", "--parent", q, "--title", "B", "--why", "w")
	c := must(t, "add", "hyp", "--parent", q, "--title", "C", "--why", "w")
	for _, h := range []string{a, b, c} {
		judge(t, h, "supports")
	}
	e := must(t, "add", "exp", "--parent", a, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", b, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", a+"=supports", "--verdict", b+"=inconclusive", "--observation", "o", "--why", "w")

	show := hyp(t, "show", q)
	want(t, show, 0, "  "+a+" and "+c+": no Result gives them different verdicts")
	want(t, show, 0, "  "+b+" and "+c+": no Result gives them different verdicts")
	if strings.Contains(show.stdout, a+" and "+b+": no Result") {
		t.Fatalf("show lists a pair a Result separates:\n%s", show.stdout)
	}
	want(t, hyp(t, "status"), 0, "  "+q+" choose: "+a+", "+b+", "+c+" (no Result separates "+a+" and "+c+"; "+b+" and "+c+")\n")
}

// A depends_on edge between two Hypotheses of one question is no premise:
// the ledger owes no premise item and the select waits on none, so hyp
// status prints no assumes line for it either.
func TestStatusAssumesOnlyAPremise(t *testing.T) {
	_, _, q := setup(t)
	a := must(t, "add", "hyp", "--parent", q, "--title", "A", "--why", "w")
	b := must(t, "add", "hyp", "--parent", q, "--title", "B", "--why", "w")
	must(t, "link", a, "depends_on", b, "--why", "w")
	judge(t, a, "supports")
	judge(t, b, "supports")
	out := hyp(t, "status")
	want(t, out, 0, "  "+q+" choose: ")
	if strings.Contains(out.stdout, "assumes") {
		t.Fatalf("status prints a premise the ledger does not hold:\n%s", out.stdout)
	}
}

// A refused select names what the runner owes first in the words OWED
// prints for that item.
func TestRefusalNamesTheItemAsOwedDoes(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	judge(t, h1, "supports")
	judge(t, h2, "inconclusive")
	item := "decide " + h2 + ", inconclusive with no planned Experiment"
	want(t, hyp(t, "status"), 0, "  runner: "+q+" "+item+"\n")
	want(t, hyp(t, "select", q, h1, "--why", "w"), 1, "refused: "+q+" owes the runner first: "+item+"\n")
	want(t, hyp(t, "select", q), 0, "suggested: none ("+q+" owes the runner first: "+item+")\n")
}

// A refused select names what the question waits on in the words hyp
// status prints under the question.
func TestRefusalNamesTheWaitAsStatusDoes(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	judge(t, h2, "supports")
	k, _ := emptyReach(t, h1)
	wait := "waits on open reading " + k + " for the owner; it holds " + q + "'s decision"
	want(t, hyp(t, "status"), 0, "    "+wait+"\n")
	want(t, hyp(t, "select", q, h2, "--why", "w"), 1, "refused: "+q+" "+wait+"\n")
}

// hyp status prints a title whole: a FRONTIER line never shows less of a
// title than the state holds.
func TestStatusPrintsTitleWhole(t *testing.T) {
	_, _, q := setup(t)
	title := strings.Repeat("abcdefghij", 12) + " END-OF-TITLE"
	must(t, "add", "hyp", "--parent", q, "--title", title, "--why", "w")
	want(t, hyp(t, "status"), 0, " "+title+"\n")
}

// hyp show prints a body on its own line: a line break in the body becomes
// a space, so a body's line never reads as a field of the view.
func TestShowPrintsBodyOnOneLine(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "the design\nstatus: complete", "--why", "w")
	out := hyp(t, "show", h)
	want(t, out, 0, "body: the design status: complete\n")
	for _, l := range strings.Split(out.stdout, "\n") {
		if l == "status: complete" {
			t.Fatalf("show prints a body line as a status field:\n%s", out.stdout)
		}
	}
}

// The stop sentence is written once: hyp prompt quotes the sentence that
// hyp status prints when the runner owes nothing, and no other Go file of
// hyp spells it.
func TestStopSentenceIsWrittenOnce(t *testing.T) {
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	sentence := "runner owes nothing: stop and report to the owner"
	var files []string
	for _, dir := range []string{"internal", "cmd"} {
		_ = filepath.WalkDir(filepath.Join(root, dir), func(path string, d os.DirEntry, err error) error {
			if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			if b, _ := os.ReadFile(path); strings.Contains(string(b), sentence) {
				rel, _ := filepath.Rel(root, path)
				files = append(files, rel)
			}
			return nil
		})
	}
	if len(files) != 1 {
		t.Fatalf("the stop sentence is spelled in %d Go files, want 1: %v", len(files), files)
	}

	promptStore(t)
	g := goalWithBody(t, "o", "Builds the design.\n"+promptBlock+"\n")
	status := hyp(t, "status", g)
	if !strings.Contains(status.stdout, "OWED\n  "+sentence+"\n") {
		t.Fatalf("status does not end OWED with the stop sentence:\n%s", status.stdout)
	}
	prompt := hyp(t, "prompt", g, "--actor", "r")
	m := regexp.MustCompile(`shows "([^"]*)" under OWED`).FindStringSubmatch(prompt.stdout)
	if prompt.code != 0 || m == nil || m[1] != sentence {
		t.Fatalf("prompt quotes %q (exit %d), want the sentence status prints", m, prompt.code)
	}
}

// An ended Goal's text prints in one form wherever hyp prints it: status,
// the prompt refusal, and the refusal of a write under the Goal.
func TestEndTextPrintsOneWay(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", "Builds.\n"+promptBlock+"\n")
	q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	must(t, "answer", "--end", g+"=first line\nsecond \"quoted\" line", "--actor", "o")
	text := `first line second "quoted" line`
	want(t, hyp(t, "status", g), 0, ": "+text+"\n")
	want(t, hyp(t, "prompt", g, "--actor", "r"), 1, `: "`+text+`"; an ended Goal has no run`)
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w"), 1, `: "`+text+`"; nothing more is written`)
}

// hyp show prints a reading's answer, and an earlier answer, whole on one
// line, as its history line prints the same answer.
func TestShowPrintsReadingAnswerOnOneLine(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	r := judge(t, h, "supports", "--reading", "a line")
	must(t, "answer", "--answer", r+"#r1=line one\nline two", "--actor", "o")
	must(t, "answer", "--answer", r+"#r1=line three\nline four", "--actor", "o")
	out := hyp(t, "show", r)
	want(t, out, 0, `answer: "line three line four" — a line`+"\n")
	want(t, out, 0, `  earlier answer: "line one line two" (`)
	if strings.Contains(out.stdout, `\n`) {
		t.Fatalf("show prints an escaped line break:\n%s", out.stdout)
	}
}

// hyp render --markdown prints a stored text on one line too, so a line of
// an observation never reads as a heading.
func TestMarkdownPrintsStoredTextOnOneLine(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "a body\n## BODY HEADING", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "obs one\n## FAKE HEADING", "--why", "w")
	out := hyp(t, "render", "--markdown")
	want(t, out, 0, "obs one ## FAKE HEADING")
	for _, l := range strings.Split(out.stdout, "\n") {
		if l == "## FAKE HEADING" || l == "## BODY HEADING" {
			t.Fatalf("markdown prints a stored text's line as a heading:\n%s", out.stdout)
		}
	}
}

// hyp add goal echoes each line on one line of its output.
func TestAddGoalEchoesALineOnOneLine(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	out := hyp(t, "add", "goal", "--title", "G", "--why", "w", "--line", "[P] first\nsecond")
	want(t, out, 0, "#1: [P] first second\n")
}
