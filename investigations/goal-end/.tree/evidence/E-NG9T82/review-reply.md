Review of commit 4df2b9e against G-NXHHVA. I wrote four tests and all four fail on this checkout, so each is a finding. The commit left the text policy (render.Text) only partly applied. I found no failure of line #1 in the stop sentence or in the refusal/status wording, and I did not probe those further.

Test file: /private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/c1ec756a-c57d-45a5-b4ca-dec5d80afc5c/scratchpad/review/internal/cli/review_channel_test.go. It uses the package's own helpers (setup, must, hyp, judge, promptStore, goalWithBody, promptBlock). It is the only file I added or changed, and gofmt reports it clean.

The commands, in order:
- `go test ./internal/cli/ -run 'TestReview' -count=1 -v`
- `go test ./internal/cli/ -run 'TestReviewShowReading' -count=1 -v`

The second command is a re-run of finding 2 after I tightened its check. The first run of `TestReviewShowReadingAnswerPrintsLikeText` passed because it only looked for the Text-form string somewhere in the output. That string appears in the history section of `show`, so the pass was not meaningful. I made the test check the `reading ...` line itself.

## Findings

### 1. `hyp prompt` on an ended Goal prints the end text with %q; `hyp status` prints it with render.Text. Label: G-NXHHVA#2 (also the #162 "prints three ways" event)

- **Where:** `internal/cli/prompt.go:150` uses `%q` on `e.Text`. Status prints the same text through `render.Text` (`status.go:136`).
- **What happens:** the same stored fact reaches the agent in two forms. A text with a line break and quotes reads `first line second "quoted" line` in status and `"first line\nsecond \"quoted\" line"` in prompt.

Output:
```
=== RUN   TestReviewEndTextPrintsOneWayInStatusAndPrompt
    review_channel_test.go:17: status:
        GOAL G-HBRA7D G — ended (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
        ENDED 2026-11-15 16:25 o: first line second "quoted" line
        FRONTIER
        STALE
        OWED
          runner owes nothing: stop and report to the owner
        AWAITING OWNER
        RECENT
          2026-11-15 16:25 o answer ends G-HBRA7D
        CHECKPOINT none

        prompt stderr: refused: hyp prompt G-HBRA7D: the Goal is ended, by o in answer 01M7G0E7SBRM3VGWQ43DK080X7: "first line\nsecond \"quoted\" line"; an ended Goal has no run
    review_channel_test.go:22: hyp prompt prints the end text another way than status: "refused: hyp prompt G-HBRA7D: the Goal is ended, by o in answer 01M7G0E7SBRM3VGWQ43DK080X7: \"first line\\nsecond \\\"quoted\\\" line\"; an ended Goal has no run\n"
--- FAIL: TestReviewEndTextPrintsOneWayInStatusAndPrompt (0.02s)
```

### 2. `hyp show` prints a reading's answer with %q on the `reading ...` line and with Text on the `answers:` history line of the same output. Label: G-NXHHVA#2 (#162)

- **Where:** `internal/render/show.go:324`, `l.Answer` through `%q`. The earlier-answer line at `show.go:332` does the same with `a.Text`. The `answers:` history line goes through `Text`.
- **What happens:** one command gives two forms of one fact.

Output of the tightened re-run (the first run's output for this test was the weak pass described above):
```
=== RUN   TestReviewShowReadingAnswerPrintsLikeText
    review_channel_test.go:34: show:
    review_channel_test.go:37: show's reading line prints the answer by another policy than render.Text (quoted/escaped): reading R-DNGRBM#r1 closed by owner answer 01M7G0EE74JPJNDGM8ABSNW5TW, answer: "line one\nline two" — a line
--- FAIL: TestReviewShowReadingAnswerPrintsLikeText (0.03s)
```
In the first run, the same `show` output also held `answers: R-BMTWGN#r1 — line one line two` on the history line below the quoted one.

### 3. `hyp render --markdown` prints observations raw, so a line break in a stored text becomes a markdown heading. Label: G-NXHHVA#2 and #3 (#162 "raw line breaks")

- **Where:** `internal/render/markdown.go` lines 31, 33, 38, 40, 61 and 75 print Title, Body and Observation without `Text`. The commit touched this file for one unrelated line.
- **What happens:** an observation `obs one\n## FAKE HEADING` produces a line the view reads as a heading, so the view states a section the state does not hold.

Output:
```
=== RUN   TestReviewMarkdownPrintsTitleOnOneLine
    review_channel_test.go:48: markdown:
        # G-P8R953 G — in_progress

        ## Q-FJN3E0 Q — open/supported_unselected

        - H-AXPPGS H — supported
          - E-5TNAPB E — run (recipe none, rule by reading)
            - R-9C3PSG recorded H-AXPPGS=supports: obs one
        ## FAKE HEADING
    review_channel_test.go:51: markdown prints an observation's line as a heading
--- FAIL: TestReviewMarkdownPrintsTitleOnOneLine (0.01s)
```
Caveat: the Markdown view might be meant to keep a text's own line breaks (its usage says "untruncated"). If so, that is a judgement for the owner. Either way, the commit's own claim "a stored text prints whole on one line" does not hold for this view.

### 4. `hyp add goal` echoes line texts raw, with line breaks. Label: G-NXHHVA#2 and #3 (#162 "raw line breaks")

- **Where:** `internal/cli/write.go:412`, `fmt.Fprintf(stdout, "  line %s: %s\n", ..., l.Text)`. The echo of limits and readings at `write.go:224` and `:231` does use `render.Text`.
- **What happens:** a line with an embedded newline prints a second, unlabelled line, so the agent reads text that is not a line of the Goal as one.

Output:
```
=== RUN   TestReviewAddGoalEchoesLinesByText
    review_channel_test.go:63: stdout:
        G-A710RJ node.add
          line G-A710RJ#1: [P] first
        second

        stderr:
    review_channel_test.go:68: add goal echoes a line with a raw line break
--- FAIL: TestReviewAddGoalEchoesLinesByText (0.00s)
FAIL
FAIL	github.com/dat9uy/hyp/internal/cli	0.413s
```

## Full test source

```go
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
```

## Not checked

- Line #1, "one block of code": the commit still leaves view-formatting text in `internal/cli`. Findings 1, 2 and 4 are one symptom of it. Because that wording is outside render, the channel is not one block. I did not write a separate test for this.
- I did not build or run the rest of the suite beyond the four `TestReview*` tests. I did not test the `validate.go` refusals, which the design says stay in replay.

Status: DONE_WITH_CONCERNS
Summary: Four red tests show render.Text is not the only text policy. `hyp prompt` and `hyp show` quote with %q, and `hyp render --markdown` and `hyp add goal` print stored texts raw. Each case gives the same fact two forms or lets a stored line break read as structure. Finding 3 depends on whether the Markdown view is meant to keep line breaks.
