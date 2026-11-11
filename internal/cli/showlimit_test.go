package cli

import "testing"

// hyp show R-…#n prints a limit: its text and state, whose turn it is, what
// its Result judged, the aimers, the closers with their verdicts and the
// checkpoints that handed it over, with their summaries.
func TestShowLimitKey(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--alias", "h-one", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "missed deploy", "--why", "w")
	key := r + "#1"

	want(t, hyp(t, "show", key), 0, "Limit "+key+"\ntext: missed deploy\nresult: "+r+" under "+e+"\nquestion: "+q+"\nstate: open: the runner runs its case\njudged: "+h+" (h-one) supports, now supported\ncandidates: "+h+" (h-one)\n")
	must(t, "checkpoint", "--summary", "outside the Goal", "--refers", key, "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: open, referred: the owner accepts or returns it\n")
	want(t, hyp(t, "show", key), 0, "checkpoint — refers: outside the Goal\n")
	must(t, "answer", "--returns", key+"=Line B covers it", "--actor", "owner-dat")
	want(t, hyp(t, "show", key), 0, "state: open, returned: the runner runs its case\n")
	want(t, hyp(t, "show", key), 0, "owner-dat answer — returns: Line B covers it\n")

	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	want(t, hyp(t, "show", key), 0, "aimed by: "+e2+" planned\n")
	c := must(t, "record", "--exp", e2, "--verdict", h+"=contradicts", "--observation", "failed", "--closes", key, "--why", "w")
	want(t, hyp(t, "show", key), 0, "state: closed by "+c+" ("+h+" (h-one) contradicts)\n")
	want(t, hyp(t, "show", key), 0, "result.record — closes it: "+h+"=contradicts\n")

	want(t, hyp(t, "show", r+"#2"), 2, "names no limit of a live Result")
}

// A closed limit names each candidate that no closer judged, since the case
// never ran on it.
func TestShowLimitNamesCandidatesNoCloserJudged(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "case C", "--why", "w")
	key := r + "#1"
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--aims", key, "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "ran", "--closes", key, "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--alias", "h-two", "--why", "w")
	e3 := must(t, "add", "exp", "--parent", h2, "--title", "E3", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
	must(t, "record", "--exp", e3, "--verdict", h2+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "show", key), 0, "not judged by a closer: "+h2+" (h-two)\n")
}
