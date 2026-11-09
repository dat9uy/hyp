package cli

import (
	"strings"
	"testing"
)

// A Result under an Experiment that aims at an open limit, and that neither
// closes it nor names a new limit, warns: a close records that the case ran,
// whatever the verdicts. The warning changes no exit code.
func TestRecordWarnsWhenAnAimedLimitStaysOpen(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "missed deploy", "--why", "w")
	key := r + "#1"
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")

	want(t, hyp(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "failed the case", "--why", "w"), 0,
		"warn: "+e2+" aims at "+key+" and this Result does not close it; a close records that the case ran, whatever the verdicts (add --closes "+key+")")

	// A Result that names the case it did not reach as a new limit does not warn.
	if out := hyp(t, "record", "--exp", e2, "--verdict", h+"=inconclusive", "--observation", "instrument failed", "--limit", "the case did not run", "--why", "w"); out.code != 0 || strings.Contains(out.stderr, "aims at") {
		t.Fatalf("exit %d, stderr %q, want no aim warning", out.code, out.stderr)
	}

	// A close with contradicts is a close; the limit's line shows the verdicts.
	if out := hyp(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "failed the case", "--closes", key, "--why", "w"); out.code != 0 || strings.Contains(out.stderr, "aims at") {
		t.Fatalf("exit %d, stderr %q, want a clean close", out.code, out.stderr)
	}
	want(t, hyp(t, "show", q), 0, key+" closed by ")
	want(t, hyp(t, "show", q), 0, " ("+h+" contradicts) — missed deploy")
	notIn(t, "limit "+key)
}

// Every Result that closes a limit is named on its line, in creation order.
func TestLimitLineNamesEveryCloser(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--alias", "h-one", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--limit", "case C", "--why", "w")
	key := r + "#1"
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--target", "none", "--why", "w")
	c1 := must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o", "--closes", key, "--why", "w")
	e3 := must(t, "add", "exp", "--parent", h2, "--title", "E3", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--target", "none", "--why", "w")
	c2 := must(t, "record", "--exp", e3, "--verdict", h2+"=contradicts", "--observation", "o", "--closes", key, "--why", "w")
	want(t, hyp(t, "show", q), 0, key+" closed by "+c1+" ("+h+" (h-one) supports), "+c2+" ("+h2+" contradicts) — case C")
}
