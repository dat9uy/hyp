Review of 29f27cb against G-5W1F5S. I found no defects, and both probes I wrote passed.

**Diff summary.** The commit moves the supersede admission check into a shared helper, `admitSupersede`, in `internal/cli/write.go`. Both `hyp supersede --old H` and `hyp link NEW supersedes H` now call it. `hyp unlink` is untouched. The helper calls `p.AdmitRetire("supersede", to)` only when the target is a Hypothesis.

**Other writers of a supersedes edge.** `grep` over `internal` and `cmd` finds only `cmdSupersede` and `cmdLink` writing one from the CLI. The "every command that writes the same supersede edge" ruling is covered. `hyp link` stays outside `usage` text that lists `depends_on|external_ref`, but it accepts `supersedes`. That is a doc nit, not a defect.

**Findings.**
1. Label: none. Premise exemption through `hyp link`: when a premise item names H as the candidate, `hyp link NEW supersedes H` succeeds, as `hyp supersede` does. Test `TestReviewLinkSupersedesPremiseCandidateReleased` passes.
2. Label: none. A supersedes link on an H with no open reading is not refused, and `hyp unlink NEW supersedes H` is not refused. Test `TestReviewLinkSupersedesUnheldAndUnlink` passes.
3. The commit's own test, `TestLinkSupersedesHeldLikeSupersede`, also passes. I did not write it. It checks that the link exits 1 naming the open reading and that H stays unsuperseded.

**Not tested.** I did not check that replay still applies a held link event from a merged log. `admitSupersede` sits in the CLI write path only, and `AdmitRetire` is the same function used before. The replay behaviour is therefore unchanged.

**Run output.** `go test ./internal/cli -run '^TestReviewLink|^TestLinkSupersedesHeldLikeSupersede$' -v -count=1` printed:
```
=== RUN   TestLinkSupersedesHeldLikeSupersede
--- PASS: TestLinkSupersedesHeldLikeSupersede (0.02s)
=== RUN   TestReviewLinkSupersedesPremiseCandidateReleased
--- PASS: TestReviewLinkSupersedesPremiseCandidateReleased (0.06s)
=== RUN   TestReviewLinkSupersedesUnheldAndUnlink
--- PASS: TestReviewLinkSupersedesUnheldAndUnlink (0.02s)
PASS
ok  	github.com/dat9uy/hyp/internal/cli	0.469s
```
`go vet ./...` exited 0 with no output. `go test ./... -count=1` passed every package (`internal/cli` 28.8s, `internal/sandbox` 72.9s). I removed the worktree after the run.

**Test file** `internal/cli/review_link_test.go`, full text:
```go
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
```

Status: DONE
Summary: The commit closes the `hyp link` bypass on line 3 as the owner ruled, and I found no defect. `go vet` and the full `go test` are clean.
Concerns/Blockers: none.
