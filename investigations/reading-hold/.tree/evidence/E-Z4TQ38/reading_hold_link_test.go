package cli

import (
	"strings"
	"testing"
)

// hyp link NEW supersedes H writes the same supersedes edge.add that
// hyp supersede NEW --old H writes, so it must be held the same way while
// an open reading's answer can move H.
func TestLinkSupersedesHeldLikeSupersede(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--why", "w")
	next := must(t, "add", "hyp", "--parent", q, "--title", "H next", "--why", "w")
	k := judge(t, h, "supports", "--reading", "a line") + "#r1"

	// The supersede command is held.
	want(t, hyp(t, "supersede", next, "--old", h, "--why", "w"), 1, "open reading "+k)

	// The same edge through hyp link is not.
	out := hyp(t, "link", next, "supersedes", h, "--why", "w")
	if out.code == 0 {
		st := hyp(t, "show", h)
		t.Fatalf("hyp link %s supersedes %s exits 0 while open reading %s can move %s; %s now:\n%s",
			next, h, k, h, h, firstLine(st.stdout, "status:"))
	}
	want(t, out, 1, "open reading "+k)
}

func firstLine(s, prefix string) string {
	for _, l := range strings.Split(s, "\n") {
		if strings.HasPrefix(strings.TrimSpace(l), prefix) {
			return l
		}
	}
	return s
}
