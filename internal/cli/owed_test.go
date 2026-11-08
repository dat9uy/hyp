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
	want(t, hyp(t, "status"), 0, "verifier: "+q+" confirm the select\n")
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
	must(t, "record", "--exp", e3, "--verdict", h+"=supports", "--observation", "o3", "--closes", key, "--why", "w")
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
	must(t, "checkpoint", "--summary", "helpers too", "--role", "owner", "--closes", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" apply answer "+key+" [owed] — helpers too\n")

	e2 := must(t, "add", "exp", "--parent", h1, "--title", "apply", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--aims", key, "--why", "w")
	want(t, hyp(t, "status"), 0, "apply answer "+key+" [in_progress by "+e2+"]")

	// An aimed Result that does not say it applies the answer leaves it owed.
	plain := must(t, "add", "exp", "--parent", h1, "--title", "plain", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	want(t, hyp(t, "record", "--exp", plain, "--verdict", h1+"=supports", "--observation", "o", "--applies", key, "--why", "w"), 1, "not among")
	must(t, "record", "--exp", plain, "--verdict", h1+"=supports", "--observation", "o", "--why", "w")
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
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "needs prod data", "--why", "w")
	key := r + "#1"

	must(t, "checkpoint", "--summary", "no test in the Goal reaches prod", "--refers", key, "--why", "w")
	notIn(t, "runner: "+q+" limit")
	want(t, hyp(t, "status"), 0, "\n  "+q+" limit referred, accept or refuse "+key+" — needs prod data\n")
	want(t, hyp(t, "status"), 0, "runner: "+q+" select "+h+", the only candidate\n")

	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--closes", key, "--why", "w"), 1, "separation of duties")
	must(t, "checkpoint", "--summary", "accept", "--role", "owner", "--closes", key, "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "show", q), 0, key+" accepted by owner-dat")
	notIn(t, "limit referred")
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}
