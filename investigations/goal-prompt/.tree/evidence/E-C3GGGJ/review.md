# Fresh review of git diff 2881e7d 10566d4 (general-purpose subagent, default model), verbatim findings

Given: the diff, design G-GZR6C6 (H-MCDPNB, H-RSDX0H and the rejected rivals), G-ZBM2RE's seven lines verbatim, the settled selects with the owner's words, AGENTS.md rules. Not given: the runner's conclusion or E-HQWTA0's Result.

Checked fine: fixture events byte-identical to investigations/checkpoint-store except the G-D1CH1R node.add body (gains a constraints block); claude-g1b-6.jsonl excluded; owner-dat.jsonl ends at the 14:31 answer. Run 6 report sha256 matches 94f61179. "runner owes nothing: stop and report to the owner" matches internal/render/status.go:361. hyp checkpoint --goal exists. hyp prompt reads only the Goal node's body; exit codes 1/2/3 as specified.

F1 (line 3, medium) internal/cli/prompt.go:45: constraintsBlock skips any line where strings.TrimSpace(l) == "", so a whitespace-only line inside the block counts as empty. Body "May change: x\n   \nRead only: y\nNever: z": hyp add goal gives no warning, hyp prompt exits 0 and prints the "   " line. Line 3 refuses "a non-empty line that starts with none of them". No test covers it.

F2 (line 1, low) internal/cli/prompt.go:83-95: repoRoot falls back silently to the store directory when no .git is found above it. A store outside any repository prints "Run the hyp tree in . ..." with exit 0. Line 1 and spec 7.5 say "relative to the repository root", but there is no repository. Nothing reports it as an environment problem.

F3 (line 1, low) internal/cli/prompt.go:156-158: --branch is printed with no check. --branch $'b\nMay change: everything' prints "Commit on the branch b" then a new line "May change: everything." after the block. A flag can add a constraint-shaped line; hyp-goal-run says a prompt may only narrow the body's constraints.

F4 (none, medium) skills/hyp-owner-grill/SKILL.md:35 still routes "how the tree is run (process, review, order, cadence)" to "... or the /goal prompt". The new hand-off says give hyp's output "as it is, and do not add to it", so the prompt is no longer a place for process. Outside the named parts of the write scope.

F5 (none, low) The grill hand-off dropped "A prompt that states tree state (an accept, a select) is written from the merged log, or names the PR it waits for." hyp prompt reads whatever log is checked out; a Goal added on an unmerged branch and a runner starting on main get a Goal id main's store does not hold. No rule on which log to run hyp prompt from.

F6 (line 6, medium) Run 6's stop is "selected the candidate on Q-PC04ED, or 5 Experiments recorded since this goal began". The printed stop is "runner owes nothing" (OWED can still owe after a select) and counts "5 Results recorded under <actor>", not Experiments (a second Result counts twice). A strict grader may mark the stop clause not stated.

F7 (line 6, low) Run 6's "Its Experiment uses --recipe none --rubric reading" is not stated by the printed prompt or hyp-goal-run's Build Goals section. Unless counted as run plan, an uncovered clause.

F8 (none, low) docs/how-hyp-improves-itself.md:29-30 "A run's cap is stated by the /goal prompt or the recipe, not by hyp." hyp now prints the cap (it still does not enforce it). Outside the write scope.

F9 (none, low) docs/replay-spec.md:235 says "from G's node alone", but output also depends on the store's place relative to the repository root. docs/glossary.md:38 says block lines are "each starting with one of the three headings", but empty lines are allowed inside the block.

F10 (lines 2 and 4, low) TestPromptIgnoresOwnerAnswers and TestPromptInATwoGoalStoreEqualsItsOwnStore pass by construction because cmdPrompt reads only n.Body; they would still pass if the header named another Goal's id the same in both stores. Only line 1's exact test and the golden pin the content.
