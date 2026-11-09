package cli

import (
	"strings"
	"testing"
)

// hyp show Q lists each pair of live rivals that no Result gives different
// verdicts, and leaves out a pair some Result separates and a rival that is
// contradicted.
func TestShowQuestionListsUnseparatedRivals(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--alias", "h-one", "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--alias", "h-two", "--title", "H2", "--body", "b", "--why", "w")
	h3 := must(t, "add", "hyp", "--parent", q, "--alias", "h-three", "--title", "H3", "--body", "b", "--why", "w")
	h4 := must(t, "add", "hyp", "--parent", q, "--alias", "h-four", "--title", "H4", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading",
		"--target", h2, "--target", h3, "--target", h4, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=supports",
		"--verdict", h3+"=inconclusive", "--verdict", h4+"=contradicts", "--observation", "o", "--why", "w")

	out := hyp(t, "show", q)
	want(t, out, 0, "not separated:\n  "+h1+" (h-one) and "+h2+" (h-two): no Result gives them different verdicts")
	if strings.Count(out.stdout, "no Result gives them") != 1 {
		t.Fatalf("want one unseparated pair:\n%s", out.stdout)
	}
}
