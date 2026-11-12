# Review of ea697bf against G-908PRZ

Scope read: `git show ea697bf` (review-result.diff), lines.txt, question.txt, the candidate's `skills/hyp-goal-run/SKILL.md`, and the commit chain `2cea6c2..ea697bf`. The Goal's base SKILL.md (sha256 cb15e2f…) is the one at `2cea6c2`. `git diff 2cea6c2 ea697bf` changes only `skills/hyp-goal-run/SKILL.md` (3 lines added, 2 removed). The changes to `internal/cli/prompt.go`, `prompt_test.go` and `hyp-owner-grill/SKILL.md` that `git show ea697bf` prints only undo the sibling commit `d0e8591`, which raised the default N from 5 to 7. Measured from the base, the candidate's default N is 5, the same as the base's.

## Findings

### 1. `git show <commit>` is not "that candidate's diff" when the candidate sits on another commit. Label: G-908PRZ#1, G-908PRZ#2

New text, line 34: "Ask a fresh subagent for one review of the candidate commit's diff (`git show <commit>`)".

`git show` prints the commit's change against its parent, not against the Goal's base. This commit shows the problem itself. `git show ea697bf` prints changes to `internal/cli/prompt.go` (`-7 +5`), to `prompt_test.go` and to `hyp-owner-grill/SKILL.md`. Measured from the base, none of these changes belongs to this candidate. A reviewer given only `git show` would review those files as changes the candidate makes.

The same gap appears in the usual case. Line 29 says "A fix is a new commit and a new Hypothesis that supersedes the old one". So `git show <fix>` shows only the fix's delta. A defect from the first commit that the fix carries forward is never reviewed against the candidate that is selected. The owner's words say the review must read "that candidate's diff". Line 29 already ties a subject with `--base <commit>`, so the text can name `git diff <base> <commit>`, the form that `edb95c6` used.

### 2. The text does not say that the gate's Result waits for the review. Label: G-908PRZ#1, G-908PRZ#3

"in its gate Experiment, so the review costs no extra Result" holds only if the runner records the gate's Result after it saves the review. Line 23 says "a Result cannot change after it is recorded". A runner that runs the gate and records its Result first, which is the usual order under lines 31–33, must then record a second Result to cite the review. That costs the extra Result the design means to avoid and uses up the N-Result budget. The text should state the order: run the gate, get the review, save it, sort its findings, and only then record the one Result. That one Result cites the gate logs and the review file, and carries any `--limit` from the findings. Line 36 needs the same order, because a finding sorted as "a `--limit`" can only be recorded on a Result that is not yet recorded.

### 3. The pre-registered verdict rule does not include the review. Label: G-908PRZ#1

Line 32 lists what the verdict rule names "before you run anything": lines aimed at, red/green tests, removal checks, `go vet`/`go test`. The review is not in that list. The new text then requires "the Result that gives the candidate its verdict" to cite the review. The text does not say whether a finding can change that verdict. For example, it does not say whether a case of a line that the candidate fails makes the Result `contradicts`. Under `--rubric reading` the rule is the body, so a verdict the review changes would rest on a rule that was never registered. Add the review, and what its findings do to the verdict, to the line 32 list.

### 4. A review per fix commit is a review chain. Label: G-908PRZ#3 (and the owner's words "without a review chain on small changes")

The text says "Review each candidate", and a fix is a new candidate (line 29). So every small fix gets its own subagent review and its own gate Result. This is the chain the owner's words exclude. `edb95c6` had said "One review of the candidate you select replaces a review of each fix commit". The candidate removes that sentence and puts nothing in its place. Nothing in the new text lets a small fix keep the review of its predecessor, or asks for one review of the fix's whole diff from base. Whether this makes cost go above 1.5x the base median is for the trials to measure. The text alone gives no limit on it.

### 5. A candidate built but not yet gated when the budget ends is not covered. Label: G-908PRZ#1, G-908PRZ#4

The review lives "in its gate Experiment". A candidate that the run commits but never gates, for example a fix committed when the last Result is used, has no review Result. Line #1 counts "every candidate it built". The stop checkpoint now names unreviewed candidates ("Candidates not reviewed: <each by Hypothesis id and commit, or none>"), and that satisfies #4. But the text gives no rule to prevent the case under #1. Such a rule could be: "when one Result is left, gate and review the current candidate; commit no new fix" (compare `1265f98`).

### 6. "Review each candidate before any select of it" has two readings. Label: G-908PRZ#1, G-908PRZ#2

- Reading 1: review every candidate (as part of its gate).
- Reading 2: a candidate needs a review only if you will select it.

Under reading 2 a run can stop with gated candidates that it never reviewed, and only names them in the checkpoint. That passes #4 but fails #1. The next sentences ("A candidate is reviewed only when such a Result exists. Select no candidate that is not reviewed.") fit both readings. The new checkpoint slot "Candidates not reviewed" suggests that unreviewed candidates are expected. Make the text say which reading it means.

### 7. "Give it the code" does not say who reads the diff. Label: G-908PRZ#1, G-908PRZ#2

The owner's words require "a subagent whose transcript in the trial read that candidate's diff". The text says "Give it the code". A runner may paste a summary or a few hunks into the prompt. The text should say to give the subagent the full diff, or the command to run itself, so that the subagent's transcript holds the diff. "Never your summary of it" applies only to the reply, not to the input.

### 8. "the Result that gives the candidate its verdict" is unclear when a candidate has more than one Result. Label: none

A candidate can get verdicts from the gate Result and later from reassess Results (line 68) or limit-closing Results (line 70). "A candidate is reviewed only when such a Result exists" is clear enough. But "the Result" (singular) does not say which Result must cite the review. Say "its gate Experiment's Result".

### 9. The checkpoint summary may become too long. Label: none

The template now adds "Candidates not reviewed: <each by Hypothesis id and commit, or none>" before the keys under AWAITING OWNER and the out-of-Goal findings. `--summary` holds at most 600 runes (`event.MaxText`). A run with several unreviewed candidates and several owner keys can go over that limit and be refused at the stop. That is the step whose purpose is to keep the fork out of chat-only. A short commit form, or naming H ids only, keeps the summary within the limit.

### 10. The commit changes files outside its scope, with a net change of zero. Label: none

As one commit, ea697bf changes `internal/cli/prompt.go` and its test. The Goal permits changes to `skills/hyp-goal-run` and to text that the change makes wrong. The net change against the base is zero, so the candidate as a subject is in scope. But the commit message ("hyp-goal-run saves each candidate's review…") does not mention that the commit also undoes the default-N change of `d0e8591`. A reader of the log could not tell. This finding is the same problem as finding 1, seen from the history side.

### 11. Text elsewhere

`hyp-owner-grill/SKILL.md` line 116, measured from the base, is unchanged ("5 by default", "a build Goal's review round included, is in hyp-goal-run"), and it is still correct. No text in `docs/`, `AGENTS.md` or `README.md` describes the review step, so the change makes no other text wrong. The new text cites no ADR and no `docs/` file. The added inline code (`git show <commit>`) names no hyp command, so skillcheck is unaffected (consistent with the settled G-908PRZ#5, R-XT3JRJ).

## Overall

The design is sound: the review is saved word for word, cited in the gate Result, and is a condition of the select, and unreviewed candidates are named at the stop. But `git show <commit>` reviews the wrong diff for stacked or fix commits, and the text does not order the review before the gate Result is recorded. Fix both before the claims "that candidate's diff" and "no extra Result" can be relied on. The fix chain (finding 4) and the unbuilt-candidate gap (finding 5) are left for the trials on #1 and #3 to show.
