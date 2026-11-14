package cli

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
)

// An open reading holds its question's decision, and its next Hypothesis
// while the work rests on it (spec §4 readings, §5.1, §8.1).

// judge records verdict on Hypothesis h in a new Experiment under h, with
// the extra flags, and returns the Result.
func judge(t *testing.T, h, verdict string, extra ...string) string {
	t.Helper()
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	args := append([]string{"record", "--exp", e, "--verdict", h + "=" + verdict, "--observation", "o", "--why", "w"}, extra...)
	return must(t, args...)
}

// emptyReach builds, under q, a Hypothesis that Result r supports with an
// open reading, and that another Result r2 contradicts, so the reading's
// answer can move no Hypothesis. It returns the reading's key and r2.
func emptyReach(t *testing.T, h string, r2Flags ...string) (k, r2 string) {
	t.Helper()
	k = judge(t, h, "supports", "--reading", "a line") + "#r1"
	r2 = judge(t, h, "contradicts", r2Flags...)
	return k, r2
}

func noOwed(t *testing.T, out result, item string) {
	t.Helper()
	owed := out.stdout
	if i := strings.Index(owed, "OWED\n"); i >= 0 {
		owed = owed[i:]
	}
	if j := strings.Index(owed, "AWAITING OWNER"); j >= 0 {
		owed = owed[:j]
	}
	if strings.Contains(owed, item) {
		t.Fatalf("OWED holds %q:\n%s", item, out.stdout)
	}
}

// Line 1 and line 6: a reading whose answer can move no Hypothesis still
// holds the lone candidate's select; the waits line and the refusal name it
// and say it holds the question's decision. Replay still applies a select
// from a merged log.
func TestOpenReadingWithNoReachHoldsSelect(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	judge(t, h2, "supports")
	k, _ := emptyReach(t, h1)

	out := hyp(t, "status")
	noOwed(t, out, "select "+h2)
	want(t, out, 0, "    waits on open reading "+k+" for the owner; it holds "+q+"'s decision\n")
	if strings.Contains(out.stdout, k+" for the owner; its answer can move") {
		t.Fatalf("the waits line names a Hypothesis the answer can move:\n%s", out.stdout)
	}
	refusal := "open reading " + k + " waits for the owner; it holds " + q + "'s decision"
	want(t, hyp(t, "select", q), 0, "suggested: none ("+refusal+")")
	want(t, hyp(t, "select", q, h2, "--why", "w"), 1, refusal)

	rawEvent(t, "m", event.TypeDQSelect, fullID(t, q), map[string]any{"hypothesis": fullID(t, h2)}, "merged")
	if c := hyp(t, "check"); strings.Contains(c.stdout, "rejected") {
		t.Fatalf("replay rejects a select from a merged log:\n%s", c.stdout)
	}
	want(t, hyp(t, "show", q), 0, "status: answered")
}

// Line 1: the owner's choice between two candidates waits too.
func TestOpenReadingWithNoReachHoldsChoice(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	h3 := must(t, "add", "hyp", "--parent", q, "--title", "H3", "--why", "w")
	judge(t, h2, "supports")
	judge(t, h3, "supports")
	k, _ := emptyReach(t, h1)

	out := hyp(t, "status")
	if strings.Contains(out.stdout, q+" choose") {
		t.Fatalf("a choice is owed while reading %s is open:\n%s", k, out.stdout)
	}
	want(t, hyp(t, "select", q, h2, "--why", "w"), 1, "open reading "+k+" waits for the owner")
}

// Line 1: a reading on a Result whose Experiment has no targets holds the
// select of the question it measured.
func TestOpenReadingOnNoTargetResultHoldsSelect(t *testing.T) {
	_, q, h := preStore(t)
	judge(t, h, "supports")
	e := must(t, "add", "exp", "--parent", q, "--title", "M", "--recipe", "tool@1", "--why", "w")
	k := must(t, "record", "--exp", e, "--observation", "stored rows 4", "--reading", "a line", "--why", "w") + "#r1"

	noOwed(t, hyp(t, "status"), "select "+h)
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "open reading "+k+" waits for the owner; it holds "+q+"'s decision")
}

// Line 1: a reopened question whose select is stale owes no stale item
// while a reading on it is open, and reaffirm of the question or of its
// Hypothesis exits 1, naming the reading.
func TestOpenReadingHoldsStaleAndReaffirm(t *testing.T) {
	// The question itself depends on T, so its select goes stale when T is
	// contradicted, and no candidate's premise waits on T's question.
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	judge(t, h, "supports")
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	tgt := must(t, "add", "hyp", "--parent", q2, "--title", "T", "--why", "w")
	judge(t, tgt, "supports")
	must(t, "select", q, h, "--why", "w")
	must(t, "link", q, "depends_on", tgt, "--why", "w")
	k := judge(t, h, "supports", "--reading", "a line") + "#r1"
	judge(t, tgt, "contradicts")

	out := hyp(t, "status")
	want(t, out, 0, "selected_stale")
	noOwed(t, out, q+" stale ")
	want(t, hyp(t, "reaffirm", h, "--why", "w"), 1, "open reading "+k+" waits for the owner")
	want(t, hyp(t, "reaffirm", q, "--why", "w"), 1, "open reading "+k+" waits for the owner")
}

// Line 2: an inconclusive Hypothesis that a Result with an open reading
// judged inconclusive is one the answer can move, so it owes no undecided
// item, whatever other Results judged; a planned Experiment on it is still
// owed. Another inconclusive Hypothesis is owed undecided.
func TestReadingOnInconclusiveOwesNoUndecided(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--alias", "h-one", "--title", "H1", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--alias", "h-two", "--title", "H2", "--why", "w")
	judge(t, h1, "inconclusive")
	judge(t, h1, "inconclusive", "--reading", "a line")
	judge(t, h2, "inconclusive")

	out := hyp(t, "status")
	noOwed(t, out, "decide "+h1)
	want(t, out, 0, q+" decide "+h2+" (h-two), inconclusive with no planned Experiment")
	e := must(t, "add", "exp", "--parent", h1, "--title", "Next", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	want(t, hyp(t, "status"), 0, q+" run "+e+" Next")
}

// Line 3: while a reading's answer can move a Hypothesis, the question's
// next Hypothesis waits: hyp add hyp, hyp supersede --old and hyp abandon of
// that Hypothesis exit 1, naming the reading. This is the event of #174: a
// supported Hypothesis, its reading open, and a new rival added at once.
func TestOpenReadingHoldsNextHypothesis(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	next := must(t, "add", "hyp", "--parent", q, "--title", "H next", "--why", "w")
	k := judge(t, h, "supports", "--reading", "a line") + "#r1"

	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w"), 1, "open reading "+k+" holds the next Hypothesis of "+q)
	want(t, hyp(t, "supersede", next, "--old", h, "--why", "w"), 1, "open reading "+k)
	want(t, hyp(t, "abandon", h, "--why", "w"), 1, "open reading "+k)

	// Replay still applies a new Hypothesis from a merged log.
	rawEvent(t, "m", event.TypeNodeAdd, "", map[string]any{"kind": event.KindHypothesis, "parent": fullID(t, q), "title": "merged"}, "merged")
	if c := hyp(t, "check"); strings.Contains(c.stdout, "rejected") {
		t.Fatalf("replay rejects a Hypothesis from a merged log:\n%s", c.stdout)
	}
}

// Line 3: an inconclusive Hypothesis the reading's Result judged
// inconclusive is one its answer can move, so it holds the next one.
func TestReadingOnInconclusiveHoldsNextHypothesis(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	k := judge(t, h, "inconclusive", "--reading", "a line") + "#r1"
	want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w"), 1, "open reading "+k+" holds the next Hypothesis of "+q)
	want(t, hyp(t, "abandon", h, "--why", "w"), 1, "open reading "+k)
}

// Line 3: a reading whose answer can move nothing still holds the next
// Hypothesis while no open limit that a Result on the question recorded is
// unreferred: none at all, only a referred one, or only a line of the Goal.
func TestReadingWithNoReachHoldsNextHypothesis(t *testing.T) {
	check := func(t *testing.T, q, k string) {
		t.Helper()
		noOwed(t, hyp(t, "status"), q+" add rival Hypotheses")
		want(t, hyp(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w"), 1, "open reading "+k+" holds the next Hypothesis of "+q)
	}
	t.Run("no limit", func(t *testing.T) {
		_, _, q := setup(t)
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		k, _ := emptyReach(t, h)
		check(t, q, k)
	})
	t.Run("a referred limit", func(t *testing.T) {
		_, _, q := setup(t)
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		k, r2 := emptyReach(t, h, "--limit", "case C")
		must(t, "checkpoint", "--refers", r2+"#1=void for every candidate", "--why", "w")
		check(t, q, k)
	})
	t.Run("a line of the Goal", func(t *testing.T) {
		dir := t.TempDir()
		t.Chdir(dir)
		t.Setenv("HYP_ACTOR", "a")
		must(t, "init")
		g := must(t, "add", "goal", "--title", "G", "--line", "case L", "--why", "start")
		q := must(t, "add", "dq", "--parent", g, "--title", "Q", "--why", "ask")
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		k, _ := emptyReach(t, h)
		check(t, q, k)
	})
}

// Line 3: a premise item on the question lets the runner add a Hypothesis,
// and supersede or abandon the candidate the premise item names.
func TestPremiseReleasesNextHypothesis(t *testing.T) {
	_, g, q := setup(t)
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	tgt := must(t, "add", "hyp", "--parent", q2, "--title", "T", "--why", "w")
	u := must(t, "add", "hyp", "--parent", q2, "--title", "U", "--why", "w")
	judge(t, tgt, "supports")
	judge(t, u, "supports")
	must(t, "select", q2, u, "--why", "the owner chose U")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	must(t, "link", h, "depends_on", tgt, "--why", "w")
	judge(t, h, "supports", "--reading", "a line")

	want(t, hyp(t, "status"), 0, q+" premise "+h+" assumes "+tgt)
	next := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	must(t, "supersede", next, "--old", h, "--why", "w")
}

// Line 4: when no open reading's answer can move a Hypothesis and an open
// limit a Result on the question recorded is unreferred, or the owner
// returned it, the next Hypothesis is released and rivals are owed.
func TestReadingWithNoReachReleasedByLimit(t *testing.T) {
	t.Run("open limit", func(t *testing.T) {
		_, _, q := setup(t)
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		emptyReach(t, h, "--limit", "case C")
		want(t, hyp(t, "status"), 0, "runner: "+q+" add rival Hypotheses")
		must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	})
	t.Run("returned limit", func(t *testing.T) {
		_, _, q := setup(t)
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
		_, r2 := emptyReach(t, h, "--limit", "case C")
		must(t, "checkpoint", "--refers", r2+"#1=void for every candidate", "--why", "w")
		must(t, "answer", "--returns", r2+"#1=run it", "--actor", "o")
		want(t, hyp(t, "status"), 0, "runner: "+q+" add rival Hypotheses")
		must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	})
}

// Line 5: once the owner answers every reading on the question, nothing is
// held, and the runner owes an apply answer item for each.
func TestAnsweredReadingsHoldNothing(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	k := judge(t, h, "supports", "--reading", "a line") + "#r1"
	must(t, "answer", "--answer", k+"=read it as code only", "--actor", "o")

	want(t, hyp(t, "status"), 0, q+" apply answer "+k)
	must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	must(t, "abandon", h, "--why", "w")
}

// Line 3, under the owner's answer that it holds every command that writes
// the same supersedes edge: hyp supersede is the only such command, so
// hyp link refuses the supersedes type and H stays live.
func TestLinkSupersedesHeldLikeSupersede(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	next := must(t, "add", "hyp", "--parent", q, "--title", "H next", "--why", "w")
	k := judge(t, h, "supports", "--reading", "a line") + "#r1"

	want(t, hyp(t, "supersede", next, "--old", h, "--why", "w"), 1, "open reading "+k)
	want(t, hyp(t, "link", next, "supersedes", h, "--why", "w"), 2, "use hyp supersede "+next+" --old "+h)
	if st := hyp(t, "show", h); strings.Contains(st.stdout, "status: superseded") {
		t.Fatalf("%s is superseded after a refused link:\n%s", h, st.stdout)
	}
}

// hyp link refuses the supersedes type even when nothing holds H, and
// hyp unlink still removes a supersedes edge to reopen H.
func TestLinkRefusesSupersedesUnlinkReopens(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	next := must(t, "add", "hyp", "--parent", q, "--title", "H next", "--why", "w")
	want(t, hyp(t, "link", next, "supersedes", h, "--why", "w"), 2, "use hyp supersede")
	must(t, "supersede", next, "--old", h, "--why", "w")
	must(t, "unlink", next, "supersedes", h, "--why", "w")
	if st := hyp(t, "show", h); strings.Contains(st.stdout, "status: superseded") {
		t.Fatalf("%s is still superseded after unlink:\n%s", h, st.stdout)
	}
}
