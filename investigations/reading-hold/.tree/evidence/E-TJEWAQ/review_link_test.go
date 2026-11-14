package cli

import "testing"

// Premise candidate is released for hyp link NEW supersedes H as for hyp supersede.
func TestReviewLinkSupersedesPremiseCandidateReleased(t *testing.T) {
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
	next := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--why", "w")
	must(t, "link", next, "supersedes", h, "--why", "w")
}

// A held H can still be linked supersedes when no reading is open on it, and
// unlink is not refused.
func TestReviewLinkSupersedesUnheldAndUnlink(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	next := must(t, "add", "hyp", "--parent", q, "--title", "N", "--why", "w")
	must(t, "link", next, "supersedes", h, "--why", "w")
	must(t, "unlink", next, "supersedes", h, "--why", "w")
}
