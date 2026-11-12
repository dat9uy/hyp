# Review of 3cfd08d against 2284172 (hyp prompt)

Scope: docs/replay-spec.md (one sentence added to the §7.5 "Constraints block" bullet) and internal/cli/prompt_test.go (two assertions added). prompt.go is unchanged. The diff against 2284172 changes no other code; the rest of the commit is files under investigations/goal-prompt/.tree.

## What I ran

All mutations ran in a scratch copy made with `git archive 3cfd08d` at
/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/e2fd7562-0eb8-46e2-aea5-6da5dfa67e3e/scratchpad/review/src.
The repository was not touched. After each mutation the original prompt.go was copied back, and a final `diff` showed it matched.

- Baseline: `go test ./internal/cli/ -run TestPrompt` passes. In the repository, `go vet ./...` is clean and `go test ./... -count=1` passes in every package, skillcheck included.
- M1: print each block line with `strings.TrimRight(l, "\r")`. The new CRLF assertion fails (prompt_test.go:77, stdout has no `\r`). **Red.**
- M2: change the body's `\r\n` to `\n` before splitting it into lines. The new CRLF assertion fails, and so does the old CRLF blank-line warning test (prompt_test.go:219). **Red.**
- M3: move the whole block check after the repository lookup. The new outside-repo assertion fails (prompt_test.go:115, exit 3, want 1). **Red.**
- M4: check before the repository lookup only that a block exists (`!found`), and check its validity (`len(problems) > 0`) after the lookup. **Every TestPrompt test passes, so this mutation survives** (see N1).
- M5: refuse any block line that ends in `\r`. The new CRLF assertion fails (exit 1). **Red.**
- I ran the built binary (/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/e2fd7562-0eb8-46e2-aea5-6da5dfa67e3e/scratchpad/review/hyp-3cfd08d) on a store outside every repository:
  - body with no block: exit 1
  - block missing "Read only:": exit 1
  - block with a line of two spaces: exit 1
  - valid block: exit 3

  stdout was empty every time. The new spec sentence holds for every invalid case, not only the missing-block case the test pins.
- `git status --short` at the end shows only the two untracked paths that were there before (claude-g113-3.jsonl and E-TZTG0G/). The tree is clean.

## Spec sentences

- "It checks the block first, so a Goal with no valid block exits 1 wherever the store is." **True of the code.** In cmdPrompt, `constraintsBlock` and the `herr.Refused` call come before `filepath.EvalSymlinks` and `repoRoot`, which raise both exit-3 cases. I confirmed this by reading and with the binary probe above. A store that `openReader` cannot open fails before the check, but then there is no Goal to read, so this does not contradict the sentence.
- Every other sentence in the bullet is unchanged from 2284172. This review did not re-check them.

## Findings

**N1. G-ZBM2RE#3: the new test pins exit 1 before exit 3 only for a missing block, not for an invalid one.**
- **The gap.** The spec sentence covers "no valid block", and the earlier review noted that the spec "did not say exit 1 comes before exit 3". The new assertion uses only the body "Builds, with no block.".
- **Failing input** (on mutation M4, not on the commit): a store outside every repository, with a Goal whose body is `x\nMay change: a.\nNever: c.`.
  - Expected: exit 1, empty stdout, stderr naming the missing "Read only:" line.
  - Actual on M4: exit 3, and every TestPrompt test still passes.
- **The commit itself is correct.** The binary probe gives exit 1 for this input.
- **Severity:** low. It is a missing test, not wrong behaviour.
- **Fix:** in TestPromptOutsideARepository, add one assertion with a body that has a heading but an invalid block, for example `"x\nMay change: a.\nNever: c."`. It should expect exit 1, empty stdout, and stderr containing `no "Read only:" line`.
- **How confirmed:** by running (mutation M4 and the binary probe).

The two items the earlier review asked for are now covered:
- (a) The CRLF byte-for-byte assertion goes red under M1, M2 and M5. Before this commit M1 would not have been caught: the only earlier CRLF test is the blank-line warning test, and M1 leaves it passing.
- (b) The no-block-outside-repo assertion goes red under M3.

**N2. None (test strength, informational).** The CRLF assertion checks a suffix of stdout and not stderr.
- **Failing input** (on a mutation, not on the commit): one that also printed the line `Builds.\r` before the block.
  - Expected: the test goes red.
  - Actual: the CRLF assertion would not see it.
- TestPromptPrintsTheGoalsRunPrompt already pins the full output exactly for an LF body, and that check would catch such a mutation.
- **How confirmed:** only by reading. No action needed.

I found no defect in the code or in the claims of the diff, measured against lines 1 to 7. Lines 6 and 7 are not affected by this diff: there is no change to skills and none to the fixture.

Status: DONE
Summary: Both new tests go red when their behaviour is broken (CRLF assertion under M1, M2, M5; exit-1-before-3 under M3), and the new spec sentence matches the code. One low gap remains (N1): only the missing-block case pins exit 1 before exit 3, so a mutation that checks block validity after the repository lookup survives.
