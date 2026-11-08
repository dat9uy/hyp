package render

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/testutil"
)

var update = flag.Bool("update", false, "rewrite the golden files")

func golden(t *testing.T, name, got string) {
	t.Helper()
	path := filepath.Join("testdata", name+".golden")
	if *update {
		if err := os.MkdirAll("testdata", 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(got), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run go test ./internal/render/ -update)", err)
	}
	if got != string(want) {
		t.Fatalf("%s differs:\n--- got\n%s--- want\n%s", name, got, want)
	}
}

// withoutRecent drops the lines between RECENT and CHECKPOINT.
func withoutRecent(s string) string {
	var out []string
	skip := false
	for _, l := range strings.Split(s, "\n") {
		switch {
		case l == "RECENT":
			skip = true
		case strings.HasPrefix(l, "CHECKPOINT"):
			skip = false
		case skip:
			continue
		}
		out = append(out, l)
	}
	return strings.Join(out, "\n")
}

func TestStatusGoldenChain(t *testing.T) {
	c := testutil.BuildChain()
	b := c.B
	b.Emit(event.TypeCheckpoint, c.G, event.Checkpoint{Summary: "all three questions answered"})
	before := Status(replay.Replay(b.Evs, nil, ""), DefaultMaxLines)
	golden(t, "status_before", before)

	r := b.Result(c.E3, map[string]string{c.H3: event.VerdictContradicts})
	contradicted := Status(replay.Replay(b.Evs, nil, ""), DefaultMaxLines)
	golden(t, "status_contradicted", contradicted)
	for _, want := range []string{"STALE", " stale ← ", event.DisplayID(event.KindResult, r)} {
		if !strings.Contains(contradicted, want) {
			t.Fatalf("contradicted status lacks %q", want)
		}
	}

	b.Retract(r)
	retracted := Status(replay.Replay(b.Evs, nil, ""), DefaultMaxLines)
	golden(t, "status_retracted", retracted)
	if withoutRecent(before) != withoutRecent(retracted) {
		t.Fatalf("retract did not restore the status screen:\n%s\n%s", before, retracted)
	}
}

func TestStatusMax60(t *testing.T) {
	b := testutil.New()
	g := b.Goal("big investigation")
	for i := 0; i < 40; i++ {
		q := b.DQ(g, fmt.Sprintf("question %d", i))
		for j := 0; j < 3; j++ {
			b.Hyp(q, fmt.Sprintf("answer %d.%d", i, j))
		}
	}
	p := replay.Replay(b.Evs, nil, "")
	lines := strings.Split(strings.TrimSuffix(Status(p, 60), "\n"), "\n")
	if len(lines) > 60 {
		t.Fatalf("%d lines", len(lines))
	}
	if !strings.HasPrefix(lines[0], "GOAL") || lines[1] != "FRONTIER" || !strings.Contains(lines[len(lines)-1], "truncated") {
		t.Fatalf("bad layout: first %q, second %q, last %q", lines[0], lines[1], lines[len(lines)-1])
	}
	dqLines := 0
	for _, l := range lines {
		if strings.HasPrefix(l, "  Q-") {
			dqLines++
		}
	}
	if dqLines != 40 {
		t.Fatalf("frontier kept %d of 40 DQ lines", dqLines)
	}
	if all := strings.Count(Status(p, 0), "\n"); all <= 60 {
		t.Fatalf("unlimited status has only %d lines", all)
	}
}

func TestShowMarkdownDiff(t *testing.T) {
	c := testutil.BuildChain()
	before := replay.Replay(c.B.Evs, nil, "")
	r := c.B.Result(c.E3, map[string]string{c.H3: event.VerdictContradicts})
	after := replay.Replay(c.B.Evs, nil, "")

	show := Show(after, c.H1, []string{c.Q1, c.G})
	for _, want := range []string{"Hypothesis H-", "path: G-", "stale: H-", "edge out: depends_on H-", "history:"} {
		if !strings.Contains(show, want) {
			t.Fatalf("show lacks %q:\n%s", want, show)
		}
	}
	md := Markdown(after)
	for _, want := range []string{"# G-", "## Q-", "- H-", "  - E-", "    - R-"} {
		if !strings.Contains(md, want) {
			t.Fatalf("markdown lacks %q:\n%s", want, md)
		}
	}
	diff := Diff(before, after)
	for _, want := range []string{"absent → recorded", "answered → reopened/selected_stale", "EVENTS", event.DisplayID(event.KindResult, r)} {
		if !strings.Contains(diff, want) {
			t.Fatalf("diff lacks %q:\n%s", want, diff)
		}
	}
}

// Frontier questions are never dropped. Past five OWED lines, ledger items
// are dropped like other detail, and the trailer counts only them.
func TestStatusFrontierOverBudgetKeepsEveryQuestion(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	for i := 0; i < 70; i++ {
		b.DQ(g, fmt.Sprintf("q%d", i))
	}
	out := Status(replay.Replay(b.Evs, nil, ""), 60)
	if strings.Count(out, "\n  Q-") != 70 || strings.Count(out, "\n  runner: ") != 5 ||
		!strings.Contains(out, "… 65 lines truncated") {
		t.Fatalf("frontier-only overflow:\n%s", out)
	}
}

func TestStatusVerifierInsufficientShowsWhy(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Supports(h1)
	sel := b.As("a").Select(q, h1, "", false)
	b.As("b").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	b.Evs[len(b.Evs)-1].Why = "no outputs were run; run the transformation twice"
	p := replay.Replay(b.Evs, nil, "")
	out := Status(p, 0)
	for _, want := range []string{"open/verifier_insufficient", "verifier: no outputs were run; run the transformation twice"} {
		if !strings.Contains(out, want) {
			t.Fatalf("status lacks %q:\n%s", want, out)
		}
	}
	show := Show(p, q, []string{g})
	if !strings.Contains(show, "verifier: insufficient — no outputs were run") {
		t.Fatalf("show lacks verifier line:\n%s", show)
	}
}

func TestSupersedeHintAndPredecessorEvidence(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	old := b.Hyp(q, "old")
	e := b.Exp(old, event.RecipeNone)
	b.Result(e, map[string]string{old: event.VerdictSupports})
	nw := b.Hyp(q, "new")
	b.Edge(nw, event.EdgeSupersedes, old)
	b.Supports(nw)
	p := replay.Replay(b.Evs, nil, "")
	hint := event.DisplayID(event.KindHypothesis, nw) + " supported (supersedes " + event.DisplayID(event.KindHypothesis, old) + ")"
	if out := Status(p, 0); !strings.Contains(out, hint) {
		t.Fatalf("status lacks %q:\n%s", hint, out)
	}
	show := Show(p, nw, []string{q, g})
	for _, want := range []string{"predecessor evidence (not counted):", "(" + event.DisplayID(event.KindExperiment, e) + ")"} {
		if !strings.Contains(show, want) {
			t.Fatalf("show lacks %q:\n%s", want, show)
		}
	}
	if p.Hyps[nw].Status != replay.HypSupported {
		t.Fatalf("derivation changed: %s", p.Hyps[nw].Status)
	}
}

func TestPredecessorEvidenceAlsoCounted(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	old := b.Hyp(q, "old")
	nw := b.Hyp(q, "new")
	e := b.Exp(old, event.RecipeNone, nw)
	both := b.Result(e, map[string]string{old: event.VerdictContradicts, nw: event.VerdictSupports})
	only := b.Result(e, map[string]string{old: event.VerdictSupports})
	b.Edge(nw, event.EdgeSupersedes, old)
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	show := Show(p, nw, []string{q, g})
	bothID := event.DisplayID(event.KindResult, both)
	onlyID := event.DisplayID(event.KindResult, only)
	var sawBoth, sawOnly bool
	// The Result that judged nw also prints under "judged by:", so read only this section.
	_, preds, _ := strings.Cut(show, "predecessor evidence")
	for _, line := range strings.Split(preds, "\n") {
		switch {
		case strings.Contains(line, bothID):
			sawBoth = true
			if !strings.HasSuffix(line, "(also counted for this hypothesis)") {
				t.Fatalf("line lacks qualifier: %q", line)
			}
		case strings.Contains(line, onlyID):
			sawOnly = true
			if strings.Contains(line, "also counted") {
				t.Fatalf("predecessor-only result is qualified: %q", line)
			}
		}
	}
	if !sawBoth || !sawOnly {
		t.Fatalf("results missing from show (both=%v only=%v):\n%s", sawBoth, sawOnly, show)
	}
	if p.Hyps[nw].Status != replay.HypSupported {
		t.Fatalf("derivation changed: %s", p.Hyps[nw].Status)
	}
}

func TestShowPrintsFullWhy(t *testing.T) {
	long := strings.Repeat("abcdefghij", 24) + "END-OF-WHY"
	edgeWhy := strings.Repeat("klmnopqrst", 24) + "END-OF-EDGE-WHY"
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	b.Evs[len(b.Evs)-1].Why = long
	h2 := b.Hyp(q, "h2")
	b.Edge(h2, event.EdgeDependsOn, h)
	b.Evs[len(b.Evs)-1].Why = edgeWhy
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	show := Show(p, h, []string{q, g})
	for _, want := range []string{long, edgeWhy} {
		if !strings.Contains(show, want) {
			t.Fatalf("show lacks full why %q:\n%s", want, show)
		}
	}
	status := Status(p, 0)
	if strings.Contains(status, "END-OF-EDGE-WHY") {
		t.Fatalf("status no longer cuts the why:\n%s", status)
	}
	var cut bool
	for _, line := range strings.Split(status, "\n") {
		if strings.Contains(line, "edge.add") && strings.HasSuffix(line, "…") {
			cut = true
		}
	}
	if !cut {
		t.Fatalf("status RECENT lacks a cut edge.add line:\n%s", status)
	}
}

func TestCheckpointSummaryInFull(t *testing.T) {
	long := strings.Repeat("abcdefghij", 52) + "END-OF-SUMMARY"
	b := testutil.New()
	g := b.Goal("g")
	b.Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: long})
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	if show := Show(p, g, nil); !strings.Contains(show, "    summary: "+long) {
		t.Fatalf("show lacks the full summary:\n%s", show)
	}
	var line string
	for _, l := range strings.Split(Status(p, 0), "\n") {
		if strings.HasPrefix(l, "CHECKPOINT ") {
			line = l
		}
	}
	if want := " (full: hyp show " + event.DisplayID(event.KindGoal, g) + ")"; !strings.Contains(line, "…") || !strings.HasSuffix(line, want) {
		t.Fatalf("cut summary lacks pointer %q: %q", want, line)
	}

	b2 := testutil.New()
	g2 := b2.Goal("g")
	b2.Emit(event.TypeCheckpoint, g2, event.Checkpoint{Summary: "short summary"})
	p2 := replay.Replay(b2.Evs, nil, "")
	for _, l := range strings.Split(Status(p2, 0), "\n") {
		if strings.HasPrefix(l, "CHECKPOINT ") && (!strings.HasSuffix(l, ": short summary") || strings.Contains(l, "(full:")) {
			t.Fatalf("short summary changed: %q", l)
		}
	}
}

func TestShowPrintsSelectKindAndWhy(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	hid := event.DisplayID(event.KindHypothesis, h1)
	b.Supports(h1)

	plain := b.Select(q, h1, "", false)
	p := replay.Replay(b.Evs, nil, "")
	if show := Show(p, q, []string{g}); !strings.Contains(show, "selected: "+hid+" — test\n") {
		t.Fatalf("plain select line wrong:\n%s", show)
	}

	// A second candidate makes the select a trade-off; no flag says so.
	b.Supports(b.Hyp(q, "h2"))
	trade := b.Select(q, h1, plain, false)
	b.Evs[len(b.Evs)-1].Why = "owner picked h1: DB-enforced revisions"
	p = replay.Replay(b.Evs, nil, "")
	if show := Show(p, q, []string{g}); !strings.Contains(show, "selected: "+hid+" (trade-off) — owner picked h1: DB-enforced revisions\n") {
		t.Fatalf("trade-off select line wrong:\n%s", show)
	}

	b.Select(q, h1, trade, true)
	p = replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	if show := Show(p, q, []string{g}); !strings.Contains(show, "selected: "+hid+" (trade-off, scope-change) — test\n") {
		t.Fatalf("two-kind select line wrong:\n%s", show)
	}
}

func TestStatusDecidedTradeoffLeavesAwaitingOwner(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	qid := event.DisplayID(event.KindDesignQuestion, q)
	h1id := event.DisplayID(event.KindHypothesis, h1)
	h2id := event.DisplayID(event.KindHypothesis, h2)

	before := Status(replay.Replay(b.Evs, nil, ""), 0)
	if !strings.Contains(before, "  "+qid+" choose: "+h1id+", "+h2id+" (no Result separates them)\n") {
		t.Fatalf("undecided trade-off missing from AWAITING OWNER:\n%s", before)
	}

	sel := b.As("a").Select(q, h1, "", true)
	b.As("b").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	after := Status(replay.Replay(b.Evs, nil, ""), 0)
	if strings.Contains(after, qid+" choose: ") {
		t.Fatalf("decided trade-off still under AWAITING OWNER:\n%s", after)
	}
	if !strings.Contains(after, "    trade-off decided: "+h1id+" over "+h2id+"\n") {
		t.Fatalf("decided trade-off line missing:\n%s", after)
	}
	_, owner, _ := strings.Cut(after, "AWAITING OWNER\n")
	if owner, _, _ = strings.Cut(owner, "RECENT\n"); owner != "" {
		t.Fatalf("AWAITING OWNER not empty:\n%s", after)
	}
}

func TestStatusContradictedChoiceAwaitsOwnerAgain(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	h3 := b.Hyp(q, "h3")
	b.Supports(h1)
	b.Supports(h2)
	b.Supports(h3)
	b.As("a").Select(q, h1, "", true)
	b.Contradicts(h1)
	qid := event.DisplayID(event.KindDesignQuestion, q)
	h2id := event.DisplayID(event.KindHypothesis, h2)
	h3id := event.DisplayID(event.KindHypothesis, h3)

	out := Status(replay.Replay(b.Evs, nil, ""), 0)
	if !strings.Contains(out, "  "+qid+" choose: "+h2id+", "+h3id+" (no Result separates them)\n") {
		t.Fatalf("trade-off line missing from AWAITING OWNER after the pick was contradicted:\n%s", out)
	}
	if strings.Contains(out, "trade-off decided") {
		t.Fatalf("contradicted pick still shown as decided:\n%s", out)
	}
}

func TestShowDesignAndRevisions(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old body"})
	b.Supports(h)
	b.Select(q, h, "", false)
	b.Emit(event.TypeHypothesisAmend, h, event.HypothesisAmend{Body: "new body"})
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	if show := Show(p, q, []string{g}); !strings.Contains(show, "design: new body\n") {
		t.Fatalf("design line missing:\n%s", show)
	}
	show := Show(p, h, []string{q, g})
	for _, s := range []string{"body: new body\n", "    body: old body\n", "    body: new body\n"} {
		if !strings.Contains(show, s) {
			t.Fatalf("show %s missing %q:\n%s", h, s, show)
		}
	}
}

func TestShowMarksLimitThatDoesNotGate(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h2, event.RecipeNone),
		Verdicts: map[string]string{h2: event.VerdictContradicts}, Observation: "o", Limits: []string{"l1"}})
	b.Select(q, h1, "", true)
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	if s := Show(p, q, []string{g}); !strings.Contains(s, " open, does not gate (judges ") {
		t.Fatalf("show does not mark the rival limit:\n%s", s)
	}
	if s := Status(p, 60); !strings.Contains(s, "OPEN LIMITS\n") || !strings.Contains(s, " — l1") {
		t.Fatalf("status does not list the rival limit:\n%s", s)
	}
}

func TestShowAndStatusLimits(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	r1 := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h, event.RecipeNone),
		Verdicts: map[string]string{h: event.VerdictSupports}, Observation: "o", Limits: []string{"l1", "l2"}})
	r2 := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h, event.RecipeNone),
		Verdicts: map[string]string{h: event.VerdictSupports}, Observation: "o", Closes: []string{event.LimitKey(r1, 2)}})
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	checks := []struct {
		name, got, want string
	}{
		{"show q open", Show(p, q, []string{g}), " open — l1\n"},
		{"show q closed", Show(p, q, []string{g}), " closed by "},
		{"show r1", Show(p, r1, nil), "limit #1: l1\n"},
		{"show r2", Show(p, r2, nil), "closes: "},
		{"status header", Status(p, 60), "OPEN LIMITS\n"},
		{"status open", Status(p, 60), " — l1"},
	}
	for _, c := range checks {
		if !strings.Contains(c.got, c.want) {
			t.Fatalf("%s: missing %q:\n%s", c.name, c.want, c.got)
		}
	}
	if s := Status(p, 60); strings.Contains(s, " — l2") {
		t.Fatalf("closed limit listed in status:\n%s", s)
	}
}

func TestMarkdownDesignIsAmendedBody(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindHypothesis, Parent: q, Title: "h", Body: "old body"})
	b.Supports(h)
	b.Select(q, h, "", false)
	b.Emit(event.TypeHypothesisAmend, h, event.HypothesisAmend{Body: "new body"})
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	md := Markdown(p)
	if !strings.Contains(md, "Selected: "+renderer{p: p}.id(h)+"\nDesign: new body\n") || strings.Contains(md, "old body") {
		t.Fatalf("design not the amended body:\n%s", md)
	}
}

func TestMarkdownNoSelectUnchanged(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	b.Hyp(q, "h")
	md := Markdown(replay.Replay(b.Evs, nil, ""))
	if strings.Contains(md, "Design:") || strings.Contains(md, "Limits:") {
		t.Fatalf("question without select or limits gained lines:\n%s", md)
	}
}

func TestMarkdownLimitsMatchShow(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q1, q2 := b.DQ(g, "q1"), b.DQ(g, "q2")
	h1, h2 := b.Hyp(q1, "h1"), b.Hyp(q2, "h2")
	r1 := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h1, event.RecipeNone),
		Verdicts: map[string]string{h1: event.VerdictSupports}, Observation: "o", Limits: []string{"l1", "l2"}})
	b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h1, event.RecipeNone),
		Verdicts: map[string]string{h1: event.VerdictSupports}, Observation: "o", Closes: []string{event.LimitKey(r1, 1)}})
	sel := b.As("a").Select(q1, h1, "", true)
	b.As("v").Confirm(q1, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As("o").Emit(event.TypeConfirm, q1, event.Confirm{Select: sel, Role: event.RoleOwner,
		Verdict: event.ConfirmSufficient, Accepts: []string{event.LimitKey(r1, 2)}})
	b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h2, event.RecipeNone),
		Verdicts: map[string]string{h2: event.VerdictSupports}, Observation: "o", Limits: []string{"l3"}})
	p := replay.Replay(b.Evs, nil, "")
	if len(p.Rejections) != 0 {
		t.Fatalf("rejections: %v", p.Rejections)
	}
	md := Markdown(p)
	for _, q := range []string{q1, q2} {
		in := false
		n := 0
		for _, line := range strings.Split(Show(p, q, []string{g}), "\n") {
			if line == "limits:" {
				in = true
				continue
			}
			if !in {
				continue
			}
			if !strings.HasPrefix(line, "  ") {
				break
			}
			n++
			if want := "- " + strings.TrimPrefix(line, "  ") + "\n"; !strings.Contains(md, want) {
				t.Fatalf("markdown missing %q:\n%s", want, md)
			}
		}
		if n == 0 {
			t.Fatalf("show %s has no limits", q)
		}
	}
	for _, s := range []string{"closed by ", "accepted by o — l2", "open — l3"} {
		if !strings.Contains(md, s) {
			t.Fatalf("markdown missing %q:\n%s", s, md)
		}
	}
}

func TestStatusPlannedOneRivalIsOneLine(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Hyp(q, "h2")
	e := b.Exp(h1, event.RecipeNone)
	out := withoutRecent(Status(replay.Replay(b.Evs, nil, ""), 0))
	id := event.DisplayID(event.KindExperiment, e)
	var lines []string
	for _, l := range strings.Split(out, "\n") {
		if strings.Contains(l, id) && !strings.HasPrefix(l, "  runner: ") {
			lines = append(lines, l)
		}
	}
	if len(lines) != 1 || !strings.Contains(lines[0], "planned [one-rival]") {
		t.Fatalf("want one planned line with its one-rival mark for %s, got %q:\n%s", id, lines, out)
	}
}

// A listed Hypothesis may hold a sub-question beside its Experiments. Only an
// Experiment has marks, so the question must not be read as one.
func TestStatusHypothesisWithSubQuestion(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	b.Hyp(q, "h2")
	sub := b.DQ(h1, "sub")
	e := b.Exp(h1, event.RecipeNone)
	out := Status(replay.Replay(b.Evs, nil, ""), 0)
	subLine := event.DisplayID(event.KindDesignQuestion, sub) + " open"
	if !strings.Contains(out, subLine) {
		t.Fatalf("want the sub-question on the frontier (%q):\n%s", subLine, out)
	}
	marked := event.DisplayID(event.KindExperiment, e) + " planned [one-rival]"
	if !strings.Contains(out, marked) {
		t.Fatalf("want the sibling Experiment's mark (%q):\n%s", marked, out)
	}
}

// lineOf returns the FRONTIER line naming id, with its leading spaces, or "".
func lineOf(out, display string) string {
	frontier := out[strings.Index(out, "FRONTIER"):strings.Index(out, "STALE")]
	for _, l := range strings.Split(frontier, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), display+" ") {
			return l
		}
	}
	return ""
}

func indentOf(l string) int { return len(l) - len(strings.TrimLeft(l, " ")) }

// A sub-question prints one level beneath the Hypothesis that holds it, so the
// line order and the indent name its owner.
func TestStatusSubQuestionPrintsBeneathItsHypothesis(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	s1 := b.DQ(h1, "sub-of-h1")
	h3 := b.Hyp(s1, "h3")
	s2 := b.DQ(h2, "sub-of-h2")
	out := Status(replay.Replay(b.Evs, nil, ""), 0)
	H := func(id string) string { return event.DisplayID(event.KindHypothesis, id) }
	Q := func(id string) string { return event.DisplayID(event.KindDesignQuestion, id) }
	if got, want := indentOf(lineOf(out, Q(s1))), indentOf(lineOf(out, H(h1)))+2; got != want {
		t.Fatalf("sub-of-h1 at indent %d, want %d:\n%s", got, want, out)
	}
	if got, want := indentOf(lineOf(out, H(h3))), indentOf(lineOf(out, Q(s1)))+2; got != want {
		t.Fatalf("h3 at indent %d, want %d:\n%s", got, want, out)
	}
	order := []string{H(h1), Q(s1), H(h3), H(h2), Q(s2)}
	at := 0
	for _, id := range order {
		i := strings.Index(out[at:], id)
		if i < 0 {
			t.Fatalf("%s out of order, want %v:\n%s", id, order, out)
		}
		at += i
	}
	if strings.Contains(out, "out of play") {
		t.Fatalf("every open question is printed, but a count line shows:\n%s", out)
	}
}

// A sub-question prints only from an open or reopened question and through a
// live rival. Otherwise it prints nothing, and the count line accounts for it.
func TestStatusPrintsOnlyOpenWork(t *testing.T) {
	Q := func(id string) string { return event.DisplayID(event.KindDesignQuestion, id) }
	for _, tc := range []struct {
		name  string
		setup func(b *testutil.B, q, h1, h2 string)
	}{
		{"parent deferred", func(b *testutil.B, q, h1, h2 string) { b.Defer(q) }},
		{"parent withdrawn", func(b *testutil.B, q, h1, h2 string) { b.Withdraw(q) }},
		{"holder contradicted", func(b *testutil.B, q, h1, h2 string) { b.Contradicts(h2) }},
		{"holder abandoned", func(b *testutil.B, q, h1, h2 string) { b.Abandon(h2) }},
		{"parent answered by the rival", func(b *testutil.B, q, h1, h2 string) {
			b.Supports(h1)
			b.Answer(q, h1, "")
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			b := testutil.New()
			g := b.Goal("g")
			q := b.DQ(g, "q")
			h1 := b.Hyp(q, "h1")
			h2 := b.Hyp(q, "h2")
			sub := b.DQ(h2, "sub-of-h2")
			tc.setup(b, q, h1, h2)
			out := Status(replay.Replay(b.Evs, nil, ""), 0)
			if l := lineOf(out, Q(sub)); l != "" {
				t.Fatalf("want the sub-question pruned, got %q:\n%s", l, out)
			}
			if !strings.Contains(out, "(1 open question out of play; hyp render --markdown lists every question)") {
				t.Fatalf("want the count line:\n%s", out)
			}
		})
	}
}

// Printed open questions plus the count line equal the Goal line's open and
// reopened counts, whatever does not print.
func TestStatusCountLineAccountsForTheGoalCounts(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	h3 := b.Hyp(q, "h3")
	in := b.DQ(h1, "in play")
	dead := b.DQ(h2, "under a contradicted rival")
	b.Hyp(dead, "x")
	deep := b.DQ(b.Hyp(dead, "y"), "deep")
	gone := b.DQ(h3, "under an abandoned rival")
	b.Contradicts(h2)
	b.Abandon(h3)
	p := replay.Replay(b.Evs, nil, "")
	out := Status(p, 0)
	Q := func(id string) string { return event.DisplayID(event.KindDesignQuestion, id) }
	if lineOf(out, Q(in)) == "" {
		t.Fatalf("want the question under the live rival:\n%s", out)
	}
	for _, id := range []string{dead, deep, gone} {
		if lineOf(out, Q(id)) != "" {
			t.Fatalf("want %s pruned:\n%s", Q(id), out)
		}
	}
	if !strings.Contains(out, "(3 open questions out of play;") {
		t.Fatalf("want three counted:\n%s", out)
	}
	if got, want := p.Goal.Counts[string(replay.DQOpen)], 5; got != want {
		t.Fatalf("Goal counts %d open, want %d printed and counted", got, want)
	}
}

// A supported Hypothesis still owes a deferred sub-question, so it prints even
// though it is not open: nothing else names the work owed.
func TestStatusPrintsADeferredBlocker(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	h2 := b.Hyp(q, "h2")
	blocking := b.DQ(h1, "blocking")
	parked := b.DQ(h2, "parked under a proposed rival")
	b.Defer(blocking)
	b.Defer(parked)
	b.Supports(h1)
	p := replay.Replay(b.Evs, nil, "")
	if s := p.Hyps[h1].Status; s != replay.HypSupported {
		t.Fatalf("h1 is %s, want %s", s, replay.HypSupported)
	}
	out := Status(p, 0)
	Q := func(id string) string { return event.DisplayID(event.KindDesignQuestion, id) }
	H := func(id string) string { return event.DisplayID(event.KindHypothesis, id) }
	l := lineOf(out, Q(blocking))
	if !strings.Contains(l, " deferred blocking") {
		t.Fatalf("want the blocker's line, got %q:\n%s", l, out)
	}
	if got, want := indentOf(l), indentOf(lineOf(out, H(h1)))+2; got != want {
		t.Fatalf("blocker at indent %d, want %d:\n%s", got, want, out)
	}
	if l := lineOf(out, Q(parked)); l != "" {
		t.Fatalf("a deferred sub-question that blocks nothing printed %q:\n%s", l, out)
	}
}

// Truncation keeps the Hypothesis that holds a printed sub-question and may
// still drop one that holds nothing.
func TestStatusTruncationKeepsTheHolder(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	var holders, bare []string
	for i := range 6 {
		h := b.Hyp(q, fmt.Sprintf("holder %d", i))
		b.DQ(h, fmt.Sprintf("sub %d", i))
		holders = append(holders, h)
		bare = append(bare, b.Hyp(q, fmt.Sprintf("bare %d", i)))
	}
	out := Status(replay.Replay(b.Evs, nil, ""), 15)
	H := func(id string) string { return event.DisplayID(event.KindHypothesis, id) }
	for _, h := range holders {
		if lineOf(out, H(h)) == "" {
			t.Fatalf("holder %s dropped:\n%s", H(h), out)
		}
	}
	for _, h := range bare {
		if lineOf(out, H(h)) != "" {
			t.Fatalf("want bare %s dropped first:\n%s", H(h), out)
		}
	}
	if !strings.Contains(out, "lines truncated") {
		t.Fatalf("want the truncation trailer:\n%s", out)
	}
}

func TestShowListsTheDossier(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "picked")
	h2 := b.Hyp(q, "rival")
	h3 := b.Hyp(q, "gone")
	b.Abandon(h3)
	sup := b.Supports(h1)
	over := b.Result(b.Exp(h1, event.RecipeNone, h2), map[string]string{h2: event.VerdictContradicts})
	s2 := b.Supports(h2)
	b.Edge(h1, event.EdgeExternalRef, "gh:1")
	before := Show(replay.Replay(b.Evs, nil, ""), q, []string{g})
	if strings.Contains(before, "dossier:") {
		t.Fatalf("no select, but a Dossier is listed:\n%s", before)
	}
	b.Select(q, h1, "", false)
	p := replay.Replay(b.Evs, nil, "")
	show := Show(p, q, []string{g})
	id := func(i string) string { return event.DisplayID(p.Nodes[i].Kind, i) }
	for _, want := range []string{
		"dossier:\n  " + id(h1) + " supported (selected) picked\n    " + id(sup) + " supports\n",
		"  " + id(h2) + " contradicted rival\n",
		"    " + id(over) + " contradicts [decides the owner requirement",
		"    " + id(s2) + " supports\n",
	} {
		if !strings.Contains(show, want) {
			t.Fatalf("missing %q in:\n%s", want, show)
		}
	}
	if strings.Contains(show, "gh:1\n  ") || strings.Contains(show, "link "+id(h1)) {
		t.Fatalf("external_ref is listed in the Dossier:\n%s", show)
	}
}

func TestShowNamesAContradictsOnARivalAsChangedSinceTheVerifier(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
	b.Supports(h1)
	b.Supports(h2)
	sel := b.As("a").Select(q, h1, "", false)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	r := b.As("a").Contradicts(h2)
	p := replay.Replay(b.Evs, nil, "")
	show := Show(p, q, []string{g})
	want := "changed since v: +" + event.DisplayID(p.Nodes[r].Kind, r)
	if !strings.Contains(show, want) {
		t.Fatalf("missing %q in:\n%s", want, show)
	}
}

func TestShowPrintsWhatChangedSinceALapsedConfirm(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	gone := b.Supports(h)
	b.Supports(h)
	sel := b.As("a").Select(q, h, "", true)
	b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmSufficient)
	show := Show(replay.Replay(b.Evs, nil, ""), q, []string{g})
	if strings.Contains(show, "changed since") {
		t.Fatalf("nothing changed, but:\n%s", show)
	}
	b.As("a").Retract(gone)
	added := b.Supports(h)
	rival := b.Hyp(q, "rival")
	b.Emit(event.TypeHypothesisAmend, h, event.HypothesisAmend{Body: "new"})
	p := replay.Replay(b.Evs, nil, "")
	show = Show(p, q, []string{g})
	id := func(i string) string { return event.DisplayID(p.Nodes[i].Kind, i) }
	for _, want := range []string{
		"verifier: sufficient — test\n",
		"changed since v: ",
		"changed since o: ",
		"+" + id(added),
		"+" + id(rival),
		"-" + id(gone),
		"body",
	} {
		if !strings.Contains(show, want) {
			t.Fatalf("missing %q in:\n%s", want, show)
		}
	}
}

// A lone candidate has no rival to be chosen over, so a --tradeoff select of it
// records no choice: the select still requires the owner, and FRONTIER prints
// no decided line.
func TestStatusLoneCandidateHasNoDecidedLine(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	b.As("a").Select(q, h, "", true)
	out := Status(replay.Replay(b.Evs, nil, ""), 0)
	if strings.Contains(out, "trade-off decided") {
		t.Fatalf("decided line for a lone candidate:\n%s", out)
	}
}

// hyp show lists a lone candidate that is not ready too; only hyp select
// withholds the suggestion.
func TestShowListsCandidates(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	b.Supports(h)
	b.DQ(h, "sub")
	out := Show(replay.Replay(b.Evs, nil, ""), q, nil)
	if want := "candidates: " + event.DisplayID(event.KindHypothesis, h) + "\n"; !strings.Contains(out, want) {
		t.Fatalf("missing %q:\n%s", want, out)
	}
}

// AWAITING OWNER names at most one step per question, and only for a question
// FRONTIER prints. OPEN LIMITS lists every open limit regardless.
func TestStatusAwaitingOwnerFollowsFrontier(t *testing.T) {
	type tree struct {
		b                      *testutil.B
		q, h1, h2, sub, s1, s2 string
	}
	// limit makes h and rival candidates, and records an open limit on a third
	// rival it contradicts, so the limit is listed but no runner owes it.
	limit := func(b *testutil.B, q, h, rival string) {
		e := b.Exp(h, "none", rival)
		b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: e,
			Verdicts:    map[string]string{h: event.VerdictSupports, rival: event.VerdictSupports},
			Observation: "o"})
		third := b.Hyp(q, "third")
		e3 := b.Exp(third, "none")
		b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: e3,
			Verdicts:    map[string]string{third: event.VerdictContradicts},
			Observation: "o", Limits: []string{"a limit"}})
	}
	// top: q has two supported rivals and an open limit.
	top := func() tree {
		b := testutil.New()
		q := b.DQ(b.Goal("g"), "q")
		h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
		limit(b, q, h1, h2)
		return tree{b: b, q: q, h1: h1, h2: h2}
	}
	// nested: the two supported rivals and the limit are under sub, on q's rival h2.
	nested := func() tree {
		b := testutil.New()
		q := b.DQ(b.Goal("g"), "q")
		h1, h2 := b.Hyp(q, "h1"), b.Hyp(q, "h2")
		sub := b.DQ(h2, "sub")
		s1, s2 := b.Hyp(sub, "s1"), b.Hyp(sub, "s2")
		limit(b, sub, s1, s2)
		return tree{b: b, q: q, h1: h1, h2: h2, sub: sub, s1: s1, s2: s2}
	}
	confirmed := func(x tree, q, h string, tradeoff bool) {
		sel := x.b.As("a").Select(q, h, "", tradeoff)
		x.b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	}
	cases := []struct {
		name  string
		build func() tree
		want  string // the AWAITING OWNER line kind, or "" for none
	}{
		{"open, no select", func() tree { return top() }, "choose"},
		{"open, plain select before the verifier", func() tree {
			x := top()
			x.b.As("a").Select(x.q, x.h1, "", false)
			return x
		}, ""},
		{"open, plain select, verifier confirmed", func() tree {
			x := top()
			confirmed(x, x.q, x.h1, false)
			return x
		}, "awaiting_owner"},
		{"open, trade-off select before the verifier", func() tree {
			x := top()
			x.b.As("a").Select(x.q, x.h1, "", true)
			return x
		}, ""},
		{"open, trade-off select, verifier confirmed", func() tree {
			x := top()
			confirmed(x, x.q, x.h1, true)
			return x
		}, "awaiting_owner"},
		{"deferred, no select", func() tree {
			x := top()
			x.b.Defer(x.q)
			return x
		}, ""},
		{"deferred, plain select, verifier confirmed", func() tree {
			x := top()
			confirmed(x, x.q, x.h1, false)
			x.b.Defer(x.q)
			return x
		}, ""},
		{"deferred, trade-off select, verifier confirmed", func() tree {
			x := top()
			confirmed(x, x.q, x.h1, true)
			x.b.Defer(x.q)
			return x
		}, ""},
		{"resumed, plain select, verifier confirmed", func() tree {
			x := top()
			confirmed(x, x.q, x.h1, false)
			x.b.Defer(x.q)
			x.b.Resume(x.q)
			return x
		}, "awaiting_owner"},
		{"withdrawn", func() tree {
			x := top()
			x.b.Withdraw(x.q)
			return x
		}, ""},
		{"sub in play", func() tree { return nested() }, "choose"},
		{"sub, parent deferred", func() tree {
			x := nested()
			x.b.Defer(x.q)
			return x
		}, ""},
		{"sub, parent withdrawn", func() tree {
			x := nested()
			x.b.Withdraw(x.q)
			return x
		}, ""},
		{"sub, rival abandoned", func() tree {
			x := nested()
			x.b.Abandon(x.h2)
			return x
		}, ""},
		{"sub, rival contradicted", func() tree {
			x := nested()
			x.b.Contradicts(x.h2)
			return x
		}, ""},
		{"sub deferred, parent in play", func() tree {
			x := nested()
			x.b.Defer(x.sub)
			return x
		}, ""},
		{"sub awaiting the owner, parent deferred", func() tree {
			x := nested()
			confirmed(x, x.sub, x.s1, false)
			x.b.Defer(x.q)
			return x
		}, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			out := Status(replay.Replay(c.build().b.Evs, nil, ""), 0)
			_, rest, _ := strings.Cut(out, "AWAITING OWNER\n")
			section, limits, ok := strings.Cut(rest, "OPEN LIMITS\n")
			if !ok {
				t.Fatalf("OPEN LIMITS missing, though it is a ledger:\n%s", out)
			}
			if !strings.Contains(limits, " — a limit\n") {
				t.Fatalf("limit missing from OPEN LIMITS:\n%s", out)
			}
			got := ""
			if lines := strings.Split(strings.TrimSpace(section), "\n"); section != "" {
				if len(lines) > 1 {
					t.Fatalf("more than one owner step:\n%s", out)
				}
				got = strings.TrimSuffix(strings.Fields(lines[0])[1], ":")
			}
			if got != c.want {
				t.Fatalf("owner step = %q, want %q:\n%s", got, c.want, out)
			}
		})
	}
}
