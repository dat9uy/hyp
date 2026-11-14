package cli

import (
	"strings"
	"testing"
)

// An open reading holds its question only while its answer can move a
// Hypothesis there, read from its own Result's verdicts (spec §8.1).

// A reading on a Result whose candidate another Result later contradicted
// can move nothing, yet it holds the question's decision and its next
// Hypothesis while no open limit on the question is unreferred: no rivals
// are owed, and the waits line says it holds the decision.
func TestReadingOnDeadRivalHoldsNextHypothesis(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--alias", "h-one", "--title", "H", "--body", "b", "--why", "w")
	e1 := must(t, "add", "exp", "--parent", h, "--title", "E1", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	r1 := must(t, "record", "--exp", e1, "--verdict", h+"=supports", "--observation", "o", "--reading", "a line", "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "o", "--why", "w")

	out := hyp(t, "status")
	want(t, out, 0, "    waits on open reading "+r1+"#r1 for the owner; it holds "+q+"'s decision")
	want(t, out, 0, "runner owes nothing")
	if strings.Contains(out.stdout, "add rival Hypotheses") {
		t.Fatalf("rivals owed while a reading holds the next Hypothesis:\n%s", out.stdout)
	}
}

// A reading on the Result whose contradicts alone keeps a rival out may
// revive it, so rivals waits for the owner.
func TestReadingOnDecidingContradictsHoldsRivals(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=contradicts", "--observation", "o", "--reading", "a line", "--why", "w")

	out := hyp(t, "status")
	want(t, out, 0, "    waits on open reading "+r+"#r1 for the owner; its answer can move "+h)
	want(t, out, 0, "runner owes nothing")
}

// The select of a lone candidate waits on a reading on its supporting Result:
// the answer may drop the only candidate.
func TestReadingHoldsLoneCandidateSelect(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "a line", "--why", "w")

	want(t, hyp(t, "select", q), 0, "suggested: none (open reading "+r+"#r1 waits for the owner; its answer can move "+h+")")
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "open reading "+r+"#r1 waits for the owner")
	if out := hyp(t, "status"); strings.Contains(out.stdout, "select "+h) {
		t.Fatalf("select owed while a reading can drop the candidate:\n%s", out.stdout)
	}
}
