package cli

import (
	"strings"
	"testing"
)

// Two open readings with the same text ask the owner one question: AWAITING
// OWNER prints them as one line, each key with the question it holds and what
// the answer can touch there, and a
// checkpoint that answers one of them warns about the other. A reading with
// other words stays on its own line.
func TestSameAskReadingsPrintAndCloseAsOne(t *testing.T) {
	_, g, q1 := setup(t)
	q2 := must(t, "add", "dq", "--parent", g, "--title", "Q2", "--why", "ask")
	rec := func(q, text string) string {
		h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
		e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
		return must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", text, "--why", "w")
	}
	r1 := rec(q1, "must a reader have a stored record?")
	r2 := rec(q2, " must a reader have a stored record? ")
	r3 := rec(q2, "another question")
	k1, k2, k3 := r1+"#r1", r2+"#r1", r3+"#r1"

	out := hyp(t, "status")
	want(t, out, 0, "  "+k1+" ("+q1+"), "+k2+" ("+q2+") reading: must a reader have a stored record?")
	want(t, out, 0, "  "+k3+" reading: another question")
	if strings.Count(out.stdout, "stored record?") != 1 {
		t.Fatalf("same question printed more than once:\n%s", out.stdout)
	}
	// Under the grouped line, each key names its question, the candidates
	// and what its Result judged. Each reading's answer can move the
	// candidate its Result supported, so it holds Q1's lone select and Q2's
	// choice, and FRONTIER names each one with what it can move.
	want(t, out, 0, "    "+k1+": "+q1+" candidates H-")
	want(t, out, 0, "    "+k2+": "+q2+" candidates H-")
	want(t, out, 0, " supports, now supported")
	want(t, out, 0, "    waits on open reading "+k1+" for the owner; its answer can move H-")
	want(t, out, 0, "    waits on open reading "+k2+" for the owner; its answer can move H-")
	want(t, out, 0, "    waits on open reading "+k3+" for the owner; its answer can move H-")

	want(t, hyp(t, "checkpoint", "--summary", "yes", "--role", "owner", "--closes", k1, "--actor", "owner-dat", "--why", "w"), 0,
		"warn: "+k2+" asks the same as "+k1+" and stays open; add --closes "+k2)
	// One copy left open prints alone, as any reading does.
	want(t, hyp(t, "status"), 0, "  "+k2+" reading: ")

	both := hyp(t, "checkpoint", "--summary", "yes", "--role", "owner", "--closes", k2, "--closes", k3, "--actor", "owner-dat", "--why", "w")
	if both.code != 0 || strings.Contains(both.stderr, "asks the same") {
		t.Fatalf("closing the last copy warned: exit %d, stderr %q", both.code, both.stderr)
	}
	want(t, hyp(t, "checkpoint", "--summary", "no", "--role", "owner", "--revises", k1, "--actor", "owner-dat", "--why", "w"), 0,
		"warn: "+k2+" asks the same as "+k1+" and keeps its earlier answer; add --revises "+k2)
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// --live leaves a contradicted Hypothesis out of FRONTIER and counts it under
// its question, but keeps one that holds a marked Experiment. Plain status
// lists them all.
func TestStatusLiveHidesContradictedRivals(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--alias", "h-one", "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--alias", "h-two", "--title", "H2", "--body", "b", "--why", "w")
	h3 := must(t, "add", "hyp", "--parent", q, "--alias", "h-three", "--title", "H3", "--body", "b", "--why", "w")
	// A planned Experiment under h-three that names no rival is marked one-rival.
	must(t, "add", "exp", "--parent", h3, "--title", "E3", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--target", h3, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=contradicts", "--verdict", h3+"=contradicts", "--observation", "o", "--why", "w")

	plain := hyp(t, "status")
	want(t, plain, 0, h2+" (h-two) contradicted")
	live := hyp(t, "status", "--live")
	want(t, live, 0, h1+" (h-one) supported")
	want(t, live, 0, h3+" (h-three) contradicted")
	want(t, live, 0, "    (1 contradicted hidden)")
	if strings.Contains(live.stdout, "(h-two)") {
		t.Fatalf("--live printed a contradicted rival:\n%s", live.stdout)
	}
	if strings.Contains(plain.stdout, "hidden") {
		t.Fatalf("plain status hid a Hypothesis:\n%s", plain.stdout)
	}
}
