I found one defect in commit 33f18e0. All other checks of lines #1 to #6 came out clean. The full suite passes on the commit: `go vet ./...` is clean and `go test ./... -count=1` passes in every package.

## Finding 1

**Label: none.** The line it defeats is G-5W1F5S#3, but that line names only `hyp supersede --old H` and `hyp abandon H`. Whether line #3 also covers the same edge written through `hyp link` is a reading for the owner.

**Statement:** `hyp link NEW supersedes H` writes the same supersedes `edge.add` event as `hyp supersede NEW --old H`, and it exits 0. So a Hypothesis that an open reading's answer can move gets superseded although the new hold refuses `hyp supersede` in the same tree.

**Cause:** the check `AdmitRetire("supersede", to)` is only in `cmdSupersede` (`internal/cli/write.go:764`). `cmdLink` (`write.go:789`) passes the edge type through and runs no write-time rule.

**Test file:** `/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/ab7de3cc-6f57-44ea-a0b0-2b60787fe4d3/scratchpad/review/internal/cli/reading_hold_link_test.go`
```go
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
```

**Output of `go test ./internal/cli -run TestLinkSupersedesHeldLikeSupersede -v -count=1` on 33f18e0:**
```
=== RUN   TestLinkSupersedesHeldLikeSupersede
    reading_hold_link_test.go:24: hyp link H-98MRWT supersedes H-E2HBYH exits 0 while open reading R-PW63Q0#r1 can move H-E2HBYH; H-E2HBYH now:
        status: superseded
--- FAIL: TestLinkSupersedesHeldLikeSupersede (0.03s)
FAIL
FAIL	github.com/dat9uy/hyp/internal/cli	0.481s
FAIL
```

## Probes that passed (no finding)
- **Line #2:** an inconclusive Hypothesis is not owed `undecided` when a planned Experiment sits under it with `--target none`. The parent counts as a target.
- **Line #6 on a sub-question** (a question under a Hypothesis): the status line "waits on open reading … it holds Q2's decision" appears, and the `hyp select` refusal says the same.
- **Code read with no defect found:**
  - Reach now includes an inconclusive Hypothesis that the reading's Result judged inconclusive.
  - `NextHold` reads only limits that Results recorded, so a Goal line is not counted. A returned limit counts as unreferred.
  - The premise exception applies to add, supersede and abandon.
  - The stale item is suppressed while a reading is open.
  - `AdmitReaffirm` covers both the question and its Hypotheses.
  - `hyp confirm` is retired and cannot write a reaffirm.
- **Docs:** the commit updates spec §4 readings, §5, §5.1, §7.3 and §8.1. The glossary changes the rivals, undecided and stale rows and adds a new Reaffirm entry (plus a Reach entry). It also updates `owner-guide.md` and both skills. The hyp-goal-run skill no longer tells the runner to copy readings by hand.

The review worktree is still at `/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/ab7de3cc-6f57-44ea-a0b0-2b60787fe4d3/scratchpad/review` (detached at 33f18e0). It holds the test file above and is uncommitted. Remove it with `git worktree remove --force` on that path when you are done. I committed nothing and did not touch the installed binary.

Status: DONE_WITH_CONCERNS
Summary: One defect: `hyp link NEW supersedes H` writes the same supersede event as `hyp supersede` and gets around the new hold on a Hypothesis an open reading's answer can move, because only `cmdSupersede` calls `AdmitRetire`. All other checks of lines #1 to #6 passed, and the full test suite is green.
Concerns: whether line #3 covers the supersede edge written by `hyp link` is the owner's reading to give. The fix would be to call `AdmitRetire` from `cmdLink` when the edge type is `supersedes`.
