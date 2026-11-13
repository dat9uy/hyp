# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?

Body:

```text
Design record for issue #141, part of #113. A run of a build Goal stopped twice with its last candidate commit unreviewed (G-ZBM2RE, runs 1 and 2), and run 3 spent each review on one more small commit. This Goal compares rival changes to the review practice of hyp-goal-run and the stop clause hyp prompt prints, and selects one or a combination. It builds nothing: the change to skills/ and internal/ is a later build Goal, which also carries the rule that hyp-goal-run states how a review is recorded and whether it counts against the budget.
The candidate is a rule, measured by replaying Sonnet runs from two starting points: the cut after R-8JQPWG (run 1, one Result left, a fix commit pending) and the cut after R-3YJ7T2 (run 2, one Result left, two limits whose cases pass). 3 trials per rival and per cut. The rivals are all under one question; H-whole-candidate is expected to lose on cost and is tested, not assumed.
May change: this tree's .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text and scratch builds for each rival rule, kept under .tree/evidence/<E-id>/
Read only: skills/, internal/, docs/, and every other investigations/ tree
Never: edit skills/ or internal/; the change is a later build Goal
Never: replace the installed hyp, or any binary outside .tree/evidence/
Never: write under an owner actor, or merge a PR
```

Lines:

1. [N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither reviewed nor named unreviewed. Reviewed: a Result that cites a fresh subagent report on that commit diff and names the commit. Named: the stop checkpoint holds the commit short sha and the word unreviewed. Fails if any of the 3 trials at either cut does.
2. [N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the stop checkpoint names it. Fails if any of the 3 trials at either cut does.
3. [N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix commit and no name in the stop checkpoint. Fails if any trial does. A cell where no trial produced such a finding is reported as not exercised, not as passed.
4. [N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own prompt prints. Fails if any of the 3 trials at either cut does.
5. [N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check that ran; a commit that only adds a test or text for such a case counts. Fails if 2 of the 3 trials at a cut do. A cell where no trial produced such a review is reported as not exercised, not as passed.
6. [N] A candidate rule mean total_cost_usd over its 3 trials at a cut is more than 1.5 times the mean of H-base at the same cut. H-base is skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) with internal/cli/prompt.go (sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed) at 9c8e498; a candidate differs from it. A hung trial counts at the highest cost of any finished trial at that cut; the hang limit is twice the slowest H-base trial there.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?

Today a review is its own Experiment whose Result counts against the run budget; hyp prompt defaults to 5 Results. Runs 1 and 2 of G-ZBM2RE stopped with the last candidate unreviewed; run 3 spent each review on one more small commit. Each rival is one rule. Test: replay Sonnet runs from the cut after R-8JQPWG (actor claude-g113-1) and the cut after R-3YJ7T2 (actor claude-g113-2), 3 trials each, with the rival own skill text and prompt binary and no event or commit after the cut. H-base must fail line 1 at the first cut, or the instrument is wrong.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-base: today's practice

Skill text and prompt as at 9c8e498 (sha256 cb15e2f2… and 21f93323…): one review of each candidate, its own Experiment, counted; default 5 Results; stop checkpoint need not name a gap. The baseline for line 6.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: the review is part of the gate Experiment

A candidate review is recorded in the gate Experiment's Result, with its limits, so a candidate costs one Experiment. Source: run 3's E-GE085A.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: the stop checkpoint names each unreviewed candidate

Keep the practice; the stop checkpoint names each candidate without a review (short sha and the word unreviewed), so the owner sees the gap. Holds alone here.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Result left, review instead of another fix

When one Result remains, the runner reviews the current candidate instead of committing another fix, and stops with recorded limits.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the --experiments default rises to 7

hyp prompt defaults to 7 Results, to fit a gate and a review per candidate. Prompt binary built with default 7; skill text as H-base.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's own probe passed closes in the review's Result

A case that the run's own probe ran and passed on the candidate is closed in the review's Result, with no new commit. A limit stays for a case no check ran or one the candidate fails.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the review reads the cumulative diff from the Goal's base

The review before the select reads the diff from the Goal's base, not the last increment. Expected to lose on cost (line 6); tested, not assumed.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-combo: one-experiment, name-at-stop, reserve-last and probe-closes together

The four rules together. Shows which rivals hold alone and which only together.
