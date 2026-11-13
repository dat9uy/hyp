# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: A budget-bound run reviews its last commit or names it, and no review runs on over small changes

Body:

```text
Design record for #141 (part of #113), tested by replayed runs: which rule of hyp-goal-run, or which default cap of hyp prompt, ends a budget-bound run of a build Goal with its last commit reviewed or named as unreviewed, with no review that goes on over small changes.
Aim 1, a review does not run out of the run's budget: lines 1 and 2. Aim 2, a review does not go on over small changes: lines 3 to 5.
Four DesignQuestions, each with today's rule as the base rival; each body names its cuts. A cut is the store at one event with the repository at that time's commit: S1 claude-g113-1 after R-8JQPWG; S2 claude-g113-2 after R-3YJ7T2; S3 claude-g113-3 after R-JHBCWS; S4 the add of H-VHG4N5 (04142f3).
Method: claude-sonnet-5-5, 3 trials for each rival and each cut, started from the cut, never from the build's start. Each trial has its own actor, a hyp prompt with --experiments set to the cap minus the Results already under the original actor at the cut, and the rule text installed over the checkout's copy. Scripts are kept as evidence with --recipe none.
Known answers first: the base rule on S1 and S2 must end with the last commit unreviewed, as runs 1 and 2 did, and one trial's cost must equal its transcript's. Until both hold, no verdict counts.
The owner adds the build Goal after the selects; it owns changes to skills/, docs/ and cmd/.
May change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, never merging it. It changes no file in cmd/, internal/, skills/ or docs/.
Read only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.
Never: commit to main or merge a PR.
Never: replace the installed hyp binary.
Never: merge a trial's events into investigations/goal-prompt/.tree/events.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a rule the run ends with a review Result for its last commit: a Result that cites a file naming that commit's sha with its findings or "no defect found".
2. [N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says that every commit, or every review limit, is closed or reviewed while one commit has no review Result.
3. [N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of at most 1.5 times the base rule's median (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), and its text differs from the base. Cost = input_tokens + output_tokens of the trial and every subagent it starts. A trial with no stop in 40 minutes ranks above every costed trial. Median of 3 = the middle value sorted. A base median that hangs gives inconclusive.
4. [N] Aim 2: no trial records more than 3 review Results.
5. [N] Aim 2: a trial closes a review finding as passed only with a cited file that ran the finding's case.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: How does a review count against the run's cap?

Cuts S1, S2, S3. Scenario: replay each cut for each rival; count the Results the run spends from the cut to its stop, and whether its last commit ends reviewed. Cap left at the cut under 5: 1, 1, 4; under 7 (H-raise-budget): 3, 3, 6. Base trials are shared with Q-last and Q-passing at the same cut.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What does a run do with one Result left and a fix commit pending?

Cuts S1, S2. Scenario: replay each cut with the cap left at 1; see whether the run stops with its last commit reviewed or named, and what it costs. Base trials are shared with Q-record at the same cut.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What happens to a review finding whose case the run's own probe already passed?

Cuts S2, S3. Scenario: replay each cut; S2 holds two findings whose probes passed (R-3YJ7T2), S3 holds N1 (R-JHBCWS). See whether a new commit follows, and what the run costs. Base trials are shared with Q-record at the same cut.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What does the review before the select read?

Cut S4, the add of H-VHG4N5 (04142f3). Scenario: replay S4 for each rival; see whether the cumulative review finds a defect the increment reviews did not, and what it costs against the Goal's base.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Base: a review is its own Experiment and its Result counts toward the cap

Today's rule: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed over the checkout's copy. A review is its own Experiment (--recipe none --rubric reading) with its own Result; cap default 5.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A commit's review is part of its gate Experiment: one Result

A commit's review is part of its gate Experiment. One Result carries the test verdicts and the review's limits, so a commit costs one Result. Rule text added to hyp-goal-run; cap default 5.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The default of --experiments rises from 5 to 7

The review stays its own Experiment; hyp prompt's default of --experiments rises from 5 to 7, to fit a gate and a review per commit. Trials use 7 minus the Results already spent at the cut. The owner argued against it on 2026-11-12 (run 2's fresh budget did not help); tested anyway.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Base: with one Result left, commit the fix and gate it

Today's rule: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed over the checkout's copy. With one Result left the run commits the pending fix and gates it.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: With one Result left, review the current candidate and stop with recorded limits

When one Result remains, the run reviews the current candidate instead of committing another fix, and stops with recorded limits. Rule text added to hyp-goal-run.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Base: a probed case is a limit, then a test commit

Today's rule: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed over the checkout's copy. A review finding whose case the run's probe passed is recorded as a limit; a test commit follows.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A case the run's probe passed is closed in the review's Result

A case the run's own probe ran and passed on the candidate is closed in the review's Result, with the probe log cited, not recorded as a limit. A limit stays for a case no check ran, or one the candidate fails. Rule text added to hyp-goal-run.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Base: the review before the select reads the last increment

Today's rule: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed over the checkout's copy. The review reads the diff of the last commit against its parent.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The review before the select reads the cumulative diff from the Goal's base

The review before the select reads the cumulative diff from the Goal's base, not the last increment. Expected to lose on cost (one cumulative review is large); tested anyway, by the owner's choice.
