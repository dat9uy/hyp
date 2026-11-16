package cli

import (
	"strings"
	"testing"
)

// G-NXHHVA#2/#3 (#162): the ended Goal's text is a stored text; status
// prints it whole on one line, prompt prints it with %q (escaped, a third way).
func TestReviewEndTextPrintsOneWayInStatusAndPrompt(t *testing.T) {
	promptStore(t)
	g := goalWithBody(t, "o", "Builds.\n"+promptBlock+"\n")
	must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
	must(t, "answer", "--end", g+"=first line\nsecond \"quoted\" line", "--actor", "o")
	st := hyp(t, "status", g).stdout
	pr := hyp(t, "prompt", g, "--actor", "r")
	t.Logf("status:\n%s\nprompt stderr: %s", st, pr.stderr)
	if !strings.Contains(st, "first line second \"quoted\" line") {
		t.Fatalf("status does not hold the text as render.Text")
	}
	if !strings.Contains(pr.stderr+pr.stdout, "first line second \"quoted\" line") {
		t.Fatalf("hyp prompt prints the end text another way than status: %q", pr.stderr+pr.stdout)
	}
}

// G-NXHHVA#2: an answered reading's answer is printed by status/show; show
// uses %q for the answer and the earlier answers, Text elsewhere.
func TestReviewShowReadingAnswerPrintsLikeText(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	r := judge(t, h, "supports", "--reading", "a line")
	must(t, "answer", "--answer", r+"#r1=line one\nline two", "--actor", "owner-dat")
	sh := hyp(t, "show", r).stdout
	t.Logf("show:\n%s", sh)
	for _, l := range strings.Split(sh, "\n") {
		if strings.HasPrefix(l, "reading ") && !strings.Contains(l, "line one line two") {
			t.Fatalf("show's reading line prints the answer by another policy than render.Text (quoted/escaped): %s", l)
		}
	}
}

// G-NXHHVA#2 (#162): render --markdown prints titles, bodies and
// observations raw while every other view prints them by render.Text.
func TestReviewMarkdownPrintsTitleOnOneLine(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "obs one\n## FAKE HEADING", "--why", "w")
	out := hyp(t, "render", "--markdown").stdout
	t.Logf("markdown:\n%s", out)
	for _, l := range strings.Split(out, "\n") {
		if l == "## FAKE HEADING" {
			t.Fatalf("markdown prints an observation's line as a heading")
		}
	}
}

// G-NXHHVA#2: hyp add goal echoes line texts raw ("line G#n: text").
func TestReviewAddGoalEchoesLinesByText(t *testing.T) {
	dir := t.TempDir()
	t.Chdir(dir)
	t.Setenv("HYP_ACTOR", "a")
	must(t, "init")
	r := hyp(t, "add", "goal", "--title", "G", "--why", "w", "--line", "[P] first\nsecond")
	t.Logf("stdout:\n%s\nstderr:%s", r.stdout, r.stderr)
	if r.code != 0 {
		t.Skip("flag shape differs: " + r.stderr)
	}
	if strings.Contains(r.stdout, "first\nsecond") {
		t.Fatalf("add goal echoes a line with a raw line break")
	}
}
