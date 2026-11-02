package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// copySynth copies the synthetic store (internal/testutil/testdata/synth: its
// events, evidence and recipes) into a temp dir and changes into it. The
// source is never written.
func copySynth(t *testing.T) {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "synth")
	if err := os.CopyFS(dir, os.DirFS("../testutil/testdata/synth")); err != nil {
		t.Fatal(err)
	}
	t.Chdir(dir)
}

// Amending the selected Hypothesis of an answered question reopens it for the
// verifier; a sufficient verdict answers it again with the new body as the
// design, and the first body stays in the Hypothesis's history.
func TestAmendShowsVerifiedDesign(t *testing.T) {
	copySynth(t)
	want(t, hyp(t, "show", "Q-WRT7N9"), 0, "status: answered")
	want(t, hyp(t, "show", "Q-WRT7N9"), 0, "design: Each process keeps an LRU of 10k entries; no shared state.")

	body := "Each process keeps an LRU of 10k entries, about 40MB, with no shared state."
	want(t, hyp(t, "amend", "H-WEE8KQ", "--body", body, "--why", "R-6B1D8V measured the memory", "--actor", "runner-2"),
		0, "H-WEE8KQ hypothesis.amend")
	want(t, hyp(t, "show", "Q-WRT7N9"), 0, "status: reopened/awaiting_verifier")

	must(t, "confirm", "Q-WRT7N9", "--role", "verifier", "--verdict", "sufficient",
		"--why", "The body states the memory R-6B1D8V measured", "--actor", "verifier-4")

	show := hyp(t, "show", "Q-WRT7N9")
	for _, s := range []string{"status: answered", "design: " + body} {
		want(t, show, 0, s)
	}
	h := hyp(t, "show", "H-WEE8KQ")
	want(t, h, 0, "    body: "+body)                                                      // the revisions are listed
	want(t, h, 0, "    body: Each process keeps an LRU of 10k entries; no shared state.") // the first body is still visible
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

func TestLimitVisibleUntilClosedOrAccepted(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r1 := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "raw SQL org gets no row", "--why", "w")
	want(t, hyp(t, "show", q), 0, r1+"#1 open — raw SQL org gets no row")
	want(t, hyp(t, "status"), 0, "OPEN LIMITS")

	// No owner needed: the limit stays visible but does not block the answer.
	must(t, "select", q, h, "--why", "w")
	must(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w", "--actor", "v")
	want(t, hyp(t, "show", q), 0, "status: answered")
	want(t, hyp(t, "show", q), 0, r1+"#1 open")

	// Owner gate: a sufficient owner confirm must accept the open limit.
	must(t, "select", q, h, "--tradeoff", "--why", "w2")
	must(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w", "--actor", "v")
	want(t, hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--why", "w", "--actor", "o"), 1, "open limits under")
	want(t, hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--accept", r1+"#1", "--why", "w", "--actor", "o"), 0, "")
	want(t, hyp(t, "show", q), 0, r1+"#1 accepted by o")
	if s := hyp(t, "status").stdout; strings.Contains(s, "OPEN LIMITS") {
		t.Fatalf("status still lists open limits after acceptance:\n%s", s)
	}

	// Close path: a later Result closes a limit; retracting it reopens the limit.
	e2 := must(t, "add", "exp", "--parent", h, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r2 := must(t, "record", "--exp", e2, "--verdict", h+"=supports", "--observation", "o2", "--limit", "second limit", "--why", "w")
	e3 := must(t, "add", "exp", "--parent", h, "--title", "E3", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r3 := must(t, "record", "--exp", e3, "--verdict", h+"=supports", "--observation", "o3", "--closes", r2+"#1", "--why", "w")
	want(t, hyp(t, "show", q), 0, r2+"#1 closed by "+r3)
	must(t, "retract", r3, "--why", "w")
	want(t, hyp(t, "show", q), 0, r2+"#1 open")
}

func TestRivalLimitDoesNotGateOwner(t *testing.T) {
	for _, judgesSelected := range []bool{false, true} {
		t.Run(fmt.Sprintf("judgesSelected=%v", judgesSelected), func(t *testing.T) {
			_, _, q := setup(t)
			h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--body", "b", "--why", "w")
			h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--why", "w")
			target := "none"
			if judgesSelected {
				target = h1
			}
			e := must(t, "add", "exp", "--parent", h2, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", target, "--why", "w")
			args := []string{"record", "--exp", e, "--verdict", h2 + "=contradicts", "--observation", "o", "--limit", "only tried one org", "--why", "w"}
			if judgesSelected {
				args = append(args, "--verdict", h1+"=supports")
			}
			r := must(t, args...)
			must(t, "select", q, h1, "--tradeoff", "--why", "w")
			must(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w", "--actor", "v")
			owner := hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--why", "w", "--actor", "o")
			if judgesSelected {
				want(t, owner, 1, r+"#1 only tried one org")
				want(t, hyp(t, "show", q), 0, r+"#1 open — only tried one org")
				return
			}
			want(t, owner, 0, "")
			want(t, hyp(t, "show", q), 0, r+"#1 open, does not gate (judges "+h2+") — only tried one org")
			want(t, hyp(t, "status"), 0, "OPEN LIMITS")
		})
	}
}

func TestInsufficientOwnerConfirmReopensLimit(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "only tried one org", "--why", "w")
	must(t, "select", q, h, "--tradeoff", "--why", "w")
	must(t, "confirm", q, "--role", "verifier", "--verdict", "sufficient", "--why", "w", "--actor", "v")
	must(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--accept", r+"#1", "--why", "w", "--actor", "o")
	want(t, hyp(t, "show", q), 0, r+"#1 accepted by o")

	// The owner's no reopens the limit they had accepted.
	must(t, "confirm", q, "--role", "owner", "--verdict", "insufficient", "--why", "changed my mind", "--actor", "o")
	want(t, hyp(t, "show", q), 0, "status: reopened/awaiting_owner")
	want(t, hyp(t, "show", q), 0, r+"#1 open — only tried one org")
	want(t, hyp(t, "status"), 0, "OPEN LIMITS")

	want(t, hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--why", "w", "--actor", "o"), 1, "open limits under")
	want(t, hyp(t, "confirm", q, "--role", "owner", "--verdict", "sufficient", "--accept", r+"#1", "--why", "w", "--actor", "o"), 0, "")
	want(t, hyp(t, "show", q), 0, r+"#1 accepted by o")
	if s := hyp(t, "status").stdout; strings.Contains(s, "OPEN LIMITS") {
		t.Fatalf("status still lists open limits after the second acceptance:\n%s", s)
	}
}
