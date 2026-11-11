package cli

import (
	"strings"
	"testing"
)

// notIn fails when status prints s.
func notIn(t *testing.T, s string) {
	t.Helper()
	if out := hyp(t, "status").stdout; strings.Contains(out, s) {
		t.Fatalf("status still prints %q:\n%s", s, out)
	}
}

// The ledger walks one candidate from untested to its select, and says when
// the runner owes nothing.
func TestOwedRunnerWalksToSelect(t *testing.T) {
	_, _, q := setup(t)
	want(t, hyp(t, "status"), 0, "runner: "+q+" add rival Hypotheses\n")
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" test "+h+"\n")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" run "+e+" E\n")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" select "+h+", the only candidate\n")
	must(t, "select", q, h, "--why", "w")
	want(t, hyp(t, "status"), 0, "complete (answered 1")
	want(t, hyp(t, "status"), 0, "runner owes nothing: stop and report to the owner\n")
}

// A limit on the candidate is owed until a Result closes it. An Experiment
// that aims at it moves it to in_progress, and a Result that does not close
// it leaves it aimed_not_settled: aiming is intent, never the outcome.
func TestOwedLimitAimedThenClosed(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "one org only", "--why", "w")
	key := r + "#1"
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [owed] — one org only\n")
	notIn(t, "select "+h)

	want(t, hyp(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", r+"#9", "--why", "w"), 1, "")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [in_progress by "+e2+"]")
	want(t, hyp(t, "show", e2), 0, "aims: "+key)
	must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o2", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [aimed_not_settled by "+e2+"]")

	e3 := must(t, "add", "exp", "--parent", h, "--title", "E3", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	// The close prints the limit's whole text, so the runner rereads every
	// case it names while claiming them.
	want(t, hyp(t, "record", "--exp", e3, "--verdict", h+"=supports", "--observation", "o3", "--closes", key, "--why", "w"), 0, "  closes "+key+": one org only\n")
	notIn(t, "limit "+key)
	want(t, hyp(t, "status"), 0, "runner: "+q+" select "+h+", the only candidate\n")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// An Experiment under one question may close a limit under another when it
// aims at that limit.
func TestOwedCrossQuestionCloseNeedsAim(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "one org only", "--why", "w")
	key := r + "#1"

	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "ask")
	h2 := must(t, "add", "hyp", "--parent", q2, "--title", "H2", "--body", "b", "--why", "w")
	plain := must(t, "add", "exp", "--parent", h2, "--title", "plain", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	want(t, hyp(t, "record", "--exp", plain, "--verdict", h2+"=supports", "--observation", "o", "--closes", key, "--why", "w"), 1, "")

	aimed := must(t, "add", "exp", "--parent", h2, "--title", "aimed", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" limit "+key+" [in_progress by "+aimed+"]")
	r2 := must(t, "record", "--exp", aimed, "--verdict", h2+"=supports", "--observation", "o", "--closes", key, "--why", "w")
	want(t, hyp(t, "show", q), 0, key+" closed by "+r2)
}

// A closed reading is owed until a Result applies it. Only an answered
// reading can be aimed at, and only an aimed one can be applied.
func TestOwedReadingAppliedByAimedResult(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--reading", "code only or helpers too", "--why", "w")
	key := r + "#r1"

	want(t, hyp(t, "add", "exp", "--parent", h1, "--title", "early", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--aims", key, "--why", "w"), 1, "is open")
	must(t, "answer", "--answer", key+"=helpers too", "--actor", "owner-dat")
	want(t, hyp(t, "status"), 0, "runner: "+q+" apply answer "+key+" [owed] — helpers too\n")

	e2 := must(t, "add", "exp", "--parent", h1, "--title", "apply", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--aims", key, "--why", "w")
	want(t, hyp(t, "status"), 0, "apply answer "+key+" [in_progress by "+e2+"]")

	// An aimed Result that does not say it applies the answer leaves it owed.
	plain := must(t, "add", "exp", "--parent", h1, "--title", "plain", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	want(t, hyp(t, "record", "--exp", plain, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--applies", key, "--why", "w"), 1, "not among")
	must(t, "record", "--exp", plain, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "apply answer "+key+" [in_progress by "+e2+"]")

	r2 := must(t, "record", "--exp", e2, "--verdict", h1+"=supports", "--verdict", h2+"=contradicts", "--observation", "o", "--applies", key, "--why", "w")
	notIn(t, "apply answer "+key)
	notIn(t, "answered readings not applied")
	want(t, hyp(t, "show", r2), 0, "applies: "+key)
	want(t, hyp(t, "status"), 0, "runner: "+q+" select "+h1+", the only candidate\n")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// A limit no test within the Goal can close is referred to the owner. It
// leaves the runner's ledger, waits under AWAITING OWNER, and only the owner
// accepts it.
func TestOwedReferredLimitWaitsForOwner(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "needs prod data", "--why", "w")
	key := r + "#1"

	must(t, "checkpoint", "--summary", "no test in the Goal reaches prod", "--refers", key, "--why", "w")
	notIn(t, "runner: "+q+" limit")
	want(t, hyp(t, "status"), 0, "\n  "+q+" limit referred, accept or return "+key+" — needs prod data\n")
	notIn(t, "runner: "+q+" select")
	want(t, hyp(t, "status"), 0, "waits on referred limit "+key+" for the owner")
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "refused: referred limit "+key+" waits for the owner")

	want(t, hyp(t, "answer", "--answer", key+"=s", "--actor", "owner-dat"), 2, "--answer \""+key+"\": want <R-id>#r<n>")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--accept", key, "--why", "w"), 2, "--accept was retired")
	want(t, hyp(t, "answer", "--accept", key+"=s"), 1, "separation of duties")
	must(t, "answer", "--accept", key+"=accept", "--actor", "owner-dat")
	want(t, hyp(t, "show", g), 0, "    accepts: "+key+" — accept\n")
	want(t, hyp(t, "show", q), 0, key+" accepted by owner-dat")
	notIn(t, "limit referred")
	notIn(t, "waits on referred limit")
	want(t, hyp(t, "status"), 0, "runner: "+q+" select "+h+", the only candidate\n")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// A reading of an answer names the reading the owner answered. It narrows
// that answer and never reopens it, so it may read only an answered reading.
func TestReadingOfAnAnswer(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "shared logic fails when?", "--why", "w")
	first := r + "#r1"

	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	rec := func(of string) result {
		return hyp(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--reading-of", of+"=does logic include helpers?", "--why", "w")
	}
	want(t, rec(first), 1, "is open; the owner has not answered it")
	want(t, rec(r+"#1"), 2, "want <R-id>#r<n>")
	want(t, hyp(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--reading-of", first, "--why", "w"), 2, "want <R-id>#r<n>=TEXT")

	must(t, "answer", "--answer", first+"=any sharing fails", "--actor", "owner-dat")
	r2 := must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--reading", "a Goal line", "--reading-of", first+"=does logic include helpers?", "--why", "w")
	second := r2 + "#r2"
	want(t, hyp(t, "show", r2), 0, "reading "+r2+"#r1 open — a Goal line\n")
	want(t, hyp(t, "show", r2), 0, "reading "+second+" of "+first+" open — does logic include helpers?\n")
	want(t, hyp(t, "status"), 0, "  "+second+" of "+first+" reading: does logic include helpers?\n")
	want(t, hyp(t, "show", g), 0, "reading "+first+" closed by owner answer")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// hyp record prints the key of each limit and reading it records, so the
// runner names them later without counting flags.
func TestRecordPrintsItsKeys(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	out := hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "first", "--limit", "second", "--reading", "a line", "--why", "w")
	r := strings.Fields(out.stdout)[0]
	want(t, out, 0, r+" result.record\n  limit "+r+"#1: first\n  limit "+r+"#2: second\n  reading "+r+"#r1: a line\n")
	must(t, "answer", "--answer", r+"#r1"+"=yes", "--actor", "owner-dat")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	out = hyp(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--reading-of", r+"#r1=does yes cover helpers?", "--why", "w")
	r2 := strings.Fields(out.stdout)[0]
	want(t, out, 0, "  reading "+r2+"#r1 of "+r+"#r1: does yes cover helpers?\n")
}

// An owner revises an answer they gave with a new answer: the new words are
// in force from that answer on, the earlier answer stays in history, and the
// revised answer is owed until a later Result applies it. Whether an answer
// is a first one or a revision is derived from the reading's state.
func TestOwnerRevisesAnAnswer(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	exp := func() string {
		return must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	}
	r := must(t, "record", "--exp", exp(), "--verdict", h+"=supports", "--observation", "o", "--reading", "is a reader a logic?", "--why", "w")
	k := r + "#r1"
	owner := func(text string) result {
		return hyp(t, "answer", "--answer", k+"="+text, "--actor", "owner-dat")
	}

	want(t, owner("a reader is a logic"), 0, " answer\n")
	// The first answer is applied only by a Result that names it in applies.
	must(t, "record", "--exp", exp(), "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" apply answer "+k)
	ea := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--aims", k, "--why", "w")
	r2 := must(t, "record", "--exp", ea, "--verdict", h+"=supports", "--observation", "o", "--applies", k, "--why", "w")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--revises", k, "--why", "w"), 2, "--revises was retired")
	want(t, hyp(t, "answer", "--answer", r+"#1"+"=s"), 2, "want <R-id>#r<n>")

	want(t, owner("a reader is a logic of the file, not of a Section"), 0, " answer\n")
	want(t, hyp(t, "show", r), 0, ", answer: \"a reader is a logic of the file, not of a Section\" — is a reader a logic?\n")
	want(t, hyp(t, "show", r), 0, "revised by owner answer ")
	want(t, hyp(t, "show", r), 0, "  earlier answer: \"a reader is a logic\" (owner answer ")
	want(t, hyp(t, "show", r), 0, ", applied by "+r2+")\n")
	owed := "runner: " + q + " apply answer " + k
	want(t, hyp(t, "status"), 0, owed)

	// A revised answer is not applied by the next Result alone.
	must(t, "record", "--exp", exp(), "--verdict", h+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, owed)
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--aims", k, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--applies", k, "--why", "w")
	if out := hyp(t, "status"); strings.Contains(out.stdout, owed) {
		t.Fatalf("revised answer still owed after a Result applied it:\n%s", out.stdout)
	}
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// Two questions whose candidates assume each other's Hypotheses hold each
// other's select; hyp check names the cycle.
func TestCheckWarnsOnPremiseCycle(t *testing.T) {
	_, g, q1 := setup(t)
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "w")
	supported := func(q, title string) string {
		h := must(t, "add", "hyp", "--parent", q, "--title", title, "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
		must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--why", "w")
		return h
	}
	a1, a2 := supported(q1, "A1"), supported(q1, "A2")
	c1, c2 := supported(q2, "C1"), supported(q2, "C2")
	must(t, "link", a1, "depends_on", c2, "--why", "w")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
	must(t, "link", c1, "depends_on", a2, "--why", "w")
	want(t, hyp(t, "check"), 0, "warn: premise cycle between "+q1+" and "+q2+"; remove a depends_on edge to break it")
	want(t, hyp(t, "select", q1), 0, "suggested: none ("+a1+" assumes "+c2+", in "+q2+" open; decide "+q2+" first)")
}
