# Review of 2284172 (fix on 5c5902d) against G-ZBM2RE and R-KYKWH9#1

Checked at 2284172 in a detached worktree (removed after). Tracked files untouched.

## Does the new test cover R-KYKWH9#1? Yes (confirmed by mutation)

Mutation: in internal/cli/prompt.go replace `storeDir, err := filepath.EvalSymlinks(r.root)` with `storeDir, err := r.root, error(nil)`.
`go test ./internal/cli -run TestPromptOutside -count=1` then FAILS:
  "through the real path: exit 3 ... through a link inside a repository: exit 0" (prints a prompt with "tree").
Unmutated: `go test ./internal/cli -run TestPrompt -count=1 -v` all PASS. So the test is red on the b4f23b2 behaviour and green on the fix; it does not pass vacuously on that axis. It asserts r.code==3 on the real path before comparing, so "both outcomes equal but wrong" is excluded. stderr equality is meaningful: it embeds the resolved store dir.

## Findings

M1 | none (R-KYKWH9 K2, already known) | The whole test body, including the new link half, sits after a t.Skipf when any ancestor of the temp dir holds `.git`. If TMPDIR is inside a repository, the case is silently not run, and `go test` still reports ok (not -v). Failing input: TMPDIR=<dir inside a repo>; expected: the link case still runs (or the test fails loudly / builds its own isolated root); actual: SKIP, exit 0. The link half does not need the skip to be global only if the real dir is outside a repo, so it cannot be rescued without a different root. Not confirmed by running (environment here has no repo above /var/folders); read from prompt_test.go lines 86-93.

M2 | G-ZBM2RE#3 (spec text) | New spec sentence "a CRLF body with no blank line in its block keeps the `\r` at the end of each block line" has no test. The only CRLF test (prompt_test.go:199) covers refusal on a blank CRLF line. By code reading it is true: constraintsBlock splits on "\n", headingOf uses HasPrefix, lines are written back with "\n" so output ends "\r\n". Failing input if someone later trims "\r": body "B\r\nMay change: a.\r\nRead only: b.\r\nNever: c.\r\n"; expected block lines end with `\r`; no test would go red. Not confirmed by running (the sandbox hook blocked my probe command that creates a `.git` dir); confirmed by reading the code only. Note the owner answered line 3 "CRLF bodies need not be accepted", so this is a stated behaviour without a gate, not a violation.

M3 | G-ZBM2RE#1 (spec text, minor imprecision) | Spec says "exits 3 when the store's directory cannot be resolved or no repository holds its real directory". True of code (EvalSymlinks error and repoRoot miss both herr.Env). Order claim elsewhere in the same paragraph (exit 1 for a bad block precedes exit 3) is also true: constraintsBlock runs before EvalSymlinks. But the spec never says that order; a Goal with an invalid block in a store outside any repository exits 1, not 3. Input: store outside any repo, body with no block. Expected by a reader of "It exits 3 when ...": 3 (or ambiguous); actual: 1. Not run; read from prompt.go. Low severity, ambiguity only.

## Spec truth otherwise
- "reads the path after resolving symbolic links ... same bytes": true, and TestPromptThroughASymlinkedDirectory goes red under the same mutation (confirmed: exit 3 through symlink).
- "does not read $HYP_ACTOR", "--branch space/control char exits 2", "--pr/--branch add clauses after the block": match prompt.go.
- Design (prompt carries no owner answers, same bytes in other stores): the diff touches nothing there.

Status: DONE_WITH_CONCERNS
