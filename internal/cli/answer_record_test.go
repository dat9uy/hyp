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

// Amending the selected Hypothesis of an answered question keeps it answered
// with the new body as the design, and the first body stays in the
// Hypothesis's history. The synthetic store's verifier confirms are history.
func TestAmendShowsTheNewDesign(t *testing.T) {
	copySynth(t)
	want(t, hyp(t, "show", "Q-WRT7N9"), 0, "status: answered")
	want(t, hyp(t, "show", "Q-WRT7N9"), 0, "design: Each process keeps an LRU of 10k entries; no shared state.")

	body := "Each process keeps an LRU of 10k entries, about 40MB, with no shared state."
	want(t, hyp(t, "amend", "H-WEE8KQ", "--body", body, "--why", "R-6B1D8V measured the memory", "--actor", "runner-2"),
		0, "H-WEE8KQ hypothesis.amend")
	show := hyp(t, "show", "Q-WRT7N9")
	for _, s := range []string{"status: answered", "design: " + body} {
		want(t, show, 0, s)
	}
	h := hyp(t, "show", "H-WEE8KQ")
	want(t, h, 0, "    body: "+body)                                                      // the revisions are listed
	want(t, h, 0, "    body: Each process keeps an LRU of 10k entries; no shared state.") // the first body is still visible
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

// refer hands limit key k to the owner with a runner checkpoint.
func refer(t *testing.T, k string) {
	t.Helper()
	must(t, "checkpoint", "--summary", "no test within the Goal settles it", "--refers", k, "--why", "w")
}

func TestLimitVisibleUntilClosedOrAccepted(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r1 := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--limit", "raw SQL org gets no row", "--why", "w")
	want(t, hyp(t, "show", q), 0, r1+"#1 open — raw SQL org gets no row")
	want(t, hyp(t, "status"), 0, "OPEN LIMITS")

	// The open limit on the candidate is the runner's: the select is refused
	// until the limit is closed or referred to the owner.
	want(t, hyp(t, "select", q, h, "--why", "w"), 1, "refused: "+q+" owes the runner first: limit "+r1+"#1")
	refer(t, r1+"#1")
	must(t, "select", q, h, "--why", "w")
	want(t, hyp(t, "show", q), 0, "status: answered")
	want(t, hyp(t, "show", q), 0, r1+"#1 open")
	want(t, hyp(t, "status"), 0, "  "+q+" limit referred, accept or refuse "+r1+"#1 — raw SQL org gets no row\n")

	// Owner gate: a sufficient owner confirm on a scope change must accept
	// the open limit.
	want(t, hyp(t, "select", q, h, "--scope-change", "--why", "w2"), 1, q+" is answered; hyp unselect it first")
	must(t, "unselect", q, "--why", "w2")
	must(t, "select", q, h, "--scope-change", "--why", "w2")
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
			if judgesSelected {
				refer(t, r+"#1")
			} else {
				e1 := must(t, "add", "exp", "--parent", h1, "--title", "E1", "--recipe", "none", "--decided-by", "reading", "--target", "none", "--why", "w")
				must(t, "record", "--exp", e1, "--verdict", h1+"=supports", "--observation", "o", "--why", "w")
			}
			sel := hyp(t, "select", q, h1, "--scope-change", "--why", "w")
			if !judgesSelected {
				want(t, sel, 0, "warn: 1 open limit judged other Hypotheses: "+r+"#1")
			}
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
	refer(t, r+"#1")
	must(t, "select", q, h, "--scope-change", "--why", "w")
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

func TestReadingAwaitsOwnerUntilOwnerCloses(t *testing.T) {
	_, g, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "read the line as code-only", "--why", "w")
	key := r + "#r1"

	want(t, hyp(t, "status"), 0, "  "+key+" reading: read the line as code-only")
	want(t, hyp(t, "show", r), 0, "reading "+key+" open — read the line as code-only")
	want(t, hyp(t, "show", g), 0, "reading "+key+" open")

	// The flag pair is a usage error, a bad reference and the recording actor are refusals.
	want(t, hyp(t, "checkpoint", "--summary", "s", "--closes", key, "--why", "w"), 2, "--closes needs --role owner")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "verifier", "--why", "w"), 2, "--role must be owner")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--closes", r+"#r2", "--actor", "owner-dat", "--why", "w"), 1, "has 1 readings")
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--closes", key, "--why", "w"), 1, "separation of duties")

	must(t, "checkpoint", "--summary", "code only", "--role", "owner", "--closes", key, "--actor", "owner-dat", "--why", "w")
	out := hyp(t, "status")
	want(t, out, 0, "AWAITING OWNER")
	if strings.Contains(out.stdout, key+" reading:") {
		t.Fatalf("closed reading still under AWAITING OWNER:\n%s", out.stdout)
	}
	want(t, hyp(t, "show", r), 0, "reading "+key+" closed by owner checkpoint")
	want(t, hyp(t, "show", r), 0, `answer: "code only" — `)
	want(t, hyp(t, "show", g), 0, "    closes: "+key)
	want(t, hyp(t, "check"), 0, "check: 0 errors, 0 warnings")
}

func TestReadingHoldsSixHundredRunes(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	over := strings.Repeat("é", 601)
	want(t, hyp(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", over, "--why", "w"), 1, "reading")
	must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", over[:len(over)-2], "--why", "w")
}

func TestRetractedResultDropsItsReading(t *testing.T) {
	_, _, q := setup(t)
	h := must(t, "add", "hyp", "--parent", q, "--title", "H", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h+"=supports", "--observation", "o", "--reading", "a reading", "--why", "w")
	must(t, "retract", r, "--why", "w")
	if out := hyp(t, "status"); strings.Contains(out.stdout, "reading:") {
		t.Fatalf("retracted Result's reading still shows:\n%s", out.stdout)
	}
	want(t, hyp(t, "checkpoint", "--summary", "s", "--role", "owner", "--closes", r+"#r1", "--actor", "owner-dat", "--why", "w"), 1, "is retracted")
}

// An answered reading the runner has not applied is the runner's item, so the
// owner is not asked to choose until a Result applies it.
func TestChooseWaitsForAnswersToApply(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--reading", "a reading", "--why", "w")
	want(t, hyp(t, "status"), 0, "  candidates (2): "+h1+", "+h2+" (trade-off)\n")
	notIn(t, " choose: ") // the open reading waits for the owner
	must(t, "checkpoint", "--summary", "my words", "--role", "owner", "--closes", r+"#r1", "--actor", "owner-dat", "--why", "w")
	want(t, hyp(t, "status"), 0, "runner: "+q+" apply answer "+r+"#r1 [owed] — my words\n")
	notIn(t, " choose: ")

	// With no Experiment aiming at the reading, a Result under the question
	// after the answer applies it, as before aims existed.
	e2 := must(t, "add", "exp", "--parent", h1, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "  "+q+" choose: "+h1+", "+h2+" (no Result separates them)\n")
}

func TestStatusNamesTheCandidatesLeft(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--body", "b", "--alias", "h-one", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--alias", "h-two", "--why", "w")
	if out := hyp(t, "status"); strings.Contains(out.stdout, "candidates (") {
		t.Fatalf("candidates line with none supported:\n%s", out.stdout)
	}
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "  candidates (2): "+h1+" (h-one), "+h2+" (h-two) (trade-off)\n")
	e2 := must(t, "add", "exp", "--parent", h2, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", h1, "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h2+"=contradicts", "--verdict", h1+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "  candidates (1): "+h1+" (h-one)\n")
}

func TestChooseMarkedWhenNoResultSeparatesCandidates(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--title", "H2", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=supports", "--observation", "o", "--why", "w")
	want(t, hyp(t, "status"), 0, "choose: "+h1+", "+h2+" (no Result separates them)")

	// One Result supporting h1 and not h2 separates them, though both stay supported.
	e2 := must(t, "add", "exp", "--parent", h1, "--title", "E2", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	must(t, "record", "--exp", e2, "--verdict", h1+"=supports", "--verdict", h2+"=inconclusive", "--observation", "o", "--why", "w")
	out := hyp(t, "status")
	want(t, out, 0, "choose: "+h1+", "+h2+"\n")
}

func TestHypothesisPageShowsWhatJudgedIt(t *testing.T) {
	_, _, q := setup(t)
	h1 := must(t, "add", "hyp", "--parent", q, "--alias", "h-one", "--title", "H1", "--body", "b", "--why", "w")
	h2 := must(t, "add", "hyp", "--parent", q, "--alias", "h-two", "--title", "H2", "--body", "b", "--why", "w")
	e := must(t, "add", "exp", "--parent", h1, "--title", "E", "--recipe", "none", "--decided-by", "reading", "--target", h2, "--why", "w")
	r := must(t, "record", "--exp", e, "--verdict", h1+"=supports", "--verdict", h2+"=contradicts", "--observation", "h-two failed the line", "--reading", "a long reading that names how the line was read and must print in full under awaiting owner", "--why", "w")

	// h2 is judged only as a target, so nothing lists this Result among its children.
	page := hyp(t, "show", h2)
	want(t, page, 0, "judged by:\n  "+r+" ("+e+") contradicts — h-two failed the line\n")
	want(t, page, 0, "    reading "+r+"#r1 open — a long reading")
	want(t, hyp(t, "show", q), 0, "  "+h2+" (h-two) contradicted")
	st := hyp(t, "status")
	want(t, st, 0, h1+" (h-one) supported")
	want(t, st, 0, "must print in full under awaiting owner")

	must(t, "retract", r, "--why", "w")
	if out := hyp(t, "show", h2); strings.Contains(out.stdout, "judged by:") {
		t.Fatalf("a retracted Result still judges:\n%s", out.stdout)
	}
}
