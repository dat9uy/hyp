# Review of 04142f3 (against d2d5d14): hyp prompt, a bad block line outside every repository

Scope: internal/cli/prompt_test.go only (+6/-1), in TestPromptOutsideARepository. Checked against internal/cli/prompt.go at 04142f3, docs/replay-spec.md §7.5 "Constraints block" and Goal lines G-ZBM2RE#1-#7.

## Gate on a scratch copy of 04142f3
- `go vet ./...` is clean. `go test ./... -count=1` passes in every package.
- TestPromptOutsideARepository runs and passes. It is not skipped: the temporary directory under /private/var/folders has no .git above it.

## Does the new assertion go red when the behaviour breaks? Yes
Each mutant below was applied to internal/cli/prompt.go in a scratch copy. I ran `go test ./internal/cli -run TestPrompt` once with the 04142f3 test file and once with the d2d5d14 test file.

| Mutant | 04142f3 tests | d2d5d14 tests |
|---|---|---|
| A: a "starts with no heading" problem is refused only after the repository check passes; missing-block and missing-heading problems are still refused first | red at prompt_test.go:125, exit 3 (want 1) | green |
| B: trailing lines that hold only whitespace are trimmed from the block, so `"  "` counts as empty | red at :125, exit 3 | green |
| C: a line with no heading is never a problem | red at :125 and in TestPromptRefusesAnInvalidConstraintsBlock | red, in TestPromptRefusesAnInvalidConstraintsBlock |

Mutant A is the defect the earlier review asked to pin. The parent's tests miss it and the new assertion catches it. The request is met.

## Is the comment now true? Yes
The comment says a Goal with no valid block exits 1 even where no repository holds the store, "whether the block is missing, lacks a heading or holds a line that starts with no heading". The test now has one assertion for each of those three cases: `noBlock`, `badBlock` and `badLine`. Each assertion checks exit 1, an empty stdout and a problem-specific stderr substring. The claim that the block is checked first holds as an observable outcome: a store outside every repository would otherwise exit 3. That matches spec §7.5 ("It checks the block first, so a Goal with no valid block exits 1 wherever the store is").

## Findings

### Q1: the bad line outside a repository is whitespace only, so a text bad line can still be deferred to exit 3 (non-blocking)
- Goal line: G-ZBM2RE#3
- The new case's bad line is `"  "`, a trailing line of two spaces. Spec §7.5 counts that as a bad line, so the request was met as worded. It is still the edge case, not the usual kind of bad line, such as prose after the block. Mutant B shows that the test also depends on the whitespace rule.
- Mutant D: defer only the bad lines whose text is not whitespace. A `"  "` or `"\t"` line is still refused first; a line like `"The end."` is refused only after the repository check. All TestPrompt tests stay green under D.
- Failing input under D: a store outside every repository (`hyp init` in a scratch directory with no .git above it) and a Goal whose body is `Builds.\nMay change: a.\nRead only: b.\nNever: c.\nThe end.`, then `hyp prompt G --actor r`.
  - Expected (line 3, spec §7.5): exit 1, empty stdout, stderr names `"The end."`.
  - Actual under D: exit 3, empty stdout, stderr "no repository holds the store …".
  - At 04142f3 (unmutated) the code is correct. The inside-repository table test (prompt_test.go:218) pins the `"The end."` text, but not outside a repository.
- Confirmed by running: mutant D built into a scratch binary and run as above, and the TestPrompt suite run under D.
- Mutant D is contrived: it needs code that tells a whitespace bad line from a text bad line. I found no such branch in the code at 04142f3, so this is a gap in coverage, not a defect.
- Cheap fix: use a text bad line, e.g. `"Builds.\nMay change: a.\nRead only: b.\nNever: c.\nThe end."`, and check `strings.Contains(bl.stderr, `starts with no heading: "The end."`)`.

### Q2: outside a repository, the stderr check does not pin which bad line is named (low)
- Goal line: G-ZBM2RE#3 ("naming … the bad line")
- The new assertion checks only the substring "starts with no heading", not the quoted line `"  "`. A mutant that names the wrong line would pass this assertion. The table test inside a repository (prompt_test.go:219-221) pins the quoted line and does not depend on where the store is, because the message comes from the same constraintsBlock call.
- Confirmed by reading only. I ran no mutant for this; the table test would catch it.
- Folding the quoted line into the substring, as in Q1's fix, closes this too.

No defect in the product code and no wrong claim in the diff. Goal lines #1, #2, #4, #5, #6 and #7 are not touched by this diff.

Status: DONE_WITH_CONCERNS
Summary: The new assertion goes red on the target mutant (bad-line refusal deferred past the repository check, exit 3), which the parent's tests miss, and the comment now matches the three cases the test pins. One non-blocking gap: the bad line used is whitespace only, so a mutant that defers only text bad lines (e.g. "The end.") outside a repository survives all prompt tests.
