# Review of d2d5d14 against 3cfd08d (internal/cli/prompt_test.go only)

Scope: the new `badBlock` assertion in TestPromptOutsideARepository and its rewritten comment. Read with internal/cli/prompt.go and docs/replay-spec.md §7.5 "Constraints block". All mutations ran in a scratch copy (`git archive d2d5d14`). The repository was not touched.

## Does the new assertion go red when the behaviour breaks? Yes

- Baseline: `go test ./internal/cli -count=1` passes. TestPromptOutsideARepository runs here (`-v` shows PASS, not SKIP), because no `.git` sits above the macOS temp dir.
- Mutation A: in cmdPrompt, refuse right away only when `!found` (no block at all). Any other block problem is refused after `repoRoot`. Result: the new assertion fails at prompt_test.go:120 with `exit 3 (want 1)`, stderr `no repository holds the store ...`. With the parent's prompt_test.go (3cfd08d), the whole package passes under Mutation A. So this assertion is the only check that catches it, and it is a real new pin.
- Mutation B: print the block on stdout before refusing. Result: the new assertion fails (`stdout "May change: a.\nNever: c."`), and so does the existing TestPromptRefusesAnInvalidConstraintsBlock.
- Mutation C: stop reporting missing headings (`!seen[h] && false`). Result: the new assertion fails with exit 3.
- Original code restored: the package passes.

## Is the test comment true? Partly

"a Goal with no valid block is a refusal, exit 1, even where no repository holds the store" is true. It matches the code: the block is checked before EvalSymlinks and repoRoot. It also matches the spec ("It checks the block first, so a Goal with no valid block exits 1 wherever the store is"). "whether the block is missing or present but not valid" is true of the code, but the test pins only one way a block can be invalid. See P1.

## Findings

### P1: the "present but not valid" claim is pinned only for a missing heading, not for a bad line (G-ZBM2RE#3)
- Input: a store outside every repository, with a Goal body `"Builds.\nMay change: a.\nRead only: b.\nNever: c.\n  "`. The last line holds two spaces, so the block has a line that starts with no heading.
- Expected (line 3 and the spec): exit 1, nothing on stdout, naming the bad line.
- Mutation D: only the bad-line problem is refused after the repository check. Under it, the hyp binary built in scratch exits 3 with "no repository holds the store ...", while the original exits 1 with `refused: ... starts with no heading: "  "`. All of `go test ./internal/cli -run TestPrompt` still passes under Mutation D.
- Severity: low. The real code has one refusal path for every problem, so Mutation D needs a deliberate split. The earlier review asked only for the missing-"Read only:" case, and that case is delivered. Still, the comment claims more than the test proves. Fix: either add a second body to the same test (for example `"Builds.\nMay change: a.\nRead only: b.\nNever: c.\nnote"`) and check exit 1, empty stdout, and stderr containing "starts with no heading", or narrow the comment to "missing or missing a heading".
- Confirmed by running (Mutation D, scratch binaries hyp-orig and hyp-mutD).

No other defect found. The diff adds no code. The badBlock Goal is added before the link section, which uses only `g`, so the later assertions are unaffected. `must` on `add goal` passes even though the add prints a warning, which matches line 3 ("hyp add goal still accepts such a body with exit 0"). This test does not assert the warning text, but line 3's add-goal warning is out of this diff's scope.

Status: DONE_WITH_CONCERNS
Summary: The new assertion goes red under a mutation that refuses an invalid but present block only after the repository check (exit 3 instead of 1), and the parent test file misses that mutation, so the pin is real. One low gap (P1): the comment's "present but not valid" covers only a missing heading, and a mutation that defers the bad-line refusal still passes all prompt tests.
