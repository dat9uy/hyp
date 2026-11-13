# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Which practice ends a budget-bound run with its last candidate reviewed or named, without the review going on

Body:

```text
Discovery Goal for issue #141 (Part of #113): which practice makes a budget-bound unattended run end with its last candidate commit reviewed, or named unreviewed in the stop checkpoint, without the review going on over small changes. A review must not easily run out of the run's budget, and must not go on and on because of small changes. The candidates are practice texts for skills/hyp-goal-run and the stop clause hyp prompt prints. The cuts are cuts of G-ZBM2RE's history in this store: run 1 after R-8JQPWG, run 2 after R-3YJ7T2, run 3 after R-JHBCWS. A trial copies the log and evidence through that Result into a bundle repository holding only the ancestors of the commit under review, uses a fresh actor with hyp prompt --experiments set to the Results the original run had left (1, 1 and 4; H-raise-budget gets that number plus 2), and runs Sonnet to its stop, with a Sonnet reviewer. 3 trials per rival and per start; the base is the current skill text. The trial prompt adds: no push, no network. Combinations are tested only if no single rival holds. The build of the chosen practice is a separate Goal the owner adds.
May change: the .tree/events and .tree/evidence of investigations/goal-prompt, through hyp commands under the run's actor; scratch bundle repositories and worktrees outside this repository. The run edits no file under skills/, internal/, docs/ or recipes/.
Read only: skills/, internal/, docs/, recipes/, every other investigations/ tree, G-ZBM2RE's and G-GZR6C6's history in this store, and ak-skill-creator under .claude/skills (eval_skill.py, for token counts).
Never: push, open or update a PR, post to GitHub, or merge.
Never: merge a trial's events into this store; a trial works on a copy.
Never: replace ~/.local/bin/hyp or an installed skill.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [P] At each of three starts (G-ZBM2RE's cut after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; a fresh actor; hyp prompt --experiments set to the Results the original run had left: 1, 1, 4), in all 3 trials of a candidate the run reaches its stop and ends with the last commit it tested either reviewed (a Result of an Experiment whose title or body names that commit cites a review file) or named, by its short hash and the word "unreviewed", in the stop checkpoint's summary. A trial with no stop fails this.
2. [P] At the second and third starts, at least 2 of the 3 trials of a candidate end with the last commit they tested reviewed, as line 1 defines "reviewed", and not merely named.
3. [N] After the cut, no trial records a third consecutive cycle of a review of a commit followed by a fix commit that supersedes it. History copied from before the cut does not count; a Result that merges a commit's gate and its review counts as the review.
4. [N] No trial selects a candidate whose commit no review Result cites, as line 1 defines "review"; a run in that state stops and names the commit.
5. [P] The candidate's hyp-goal-run text states how a review of a candidate commit is recorded (its command, and whether it has its own Experiment) and whether its Result counts against the run's budget, in the skill's own words, and cites no ADR and no file under docs/.
6. [P] At each start, the mean over 3 trials of a candidate's cost, the total tokens of the trial's agent and every subagent it spawns, is at most 1.5 times the mean of the base's 3 trials at that start. A trial not stopped at the cap, 3 times the largest base trial, is scored at the cap. The base is hyp-goal-run SKILL.md with sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and every candidate's text differs from it.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What practice ends a budget-bound run with its last candidate reviewed or named, without the review going on?

Scenario: replay G-ZBM2RE's cut after R-8JQPWG (run 1, 1 Result left, a fix pending), after R-3YJ7T2 (run 2, 1 left, the last commit not needed) and after R-JHBCWS (run 3, 4 left, a probe-passed case); 3 Sonnet trials per rival and per start, against the current skill text as base. Each rival is a practice for hyp-goal-run and the stop clause hyp prompt prints. Issue #141.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: a review is part of its candidate's gate Experiment

One Experiment and one Result carry the candidate's test verdicts and the review's limits, so a candidate costs one Result again. The skill states that a review is not its own Experiment and counts as part of the gate's Result.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: keep the practice; the stop names each candidate with no review

A review stays its own Experiment and Result. When the stop clause is reached, the stop checkpoint names each candidate commit without a review by its short hash and the word unreviewed, and the run selects none.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Result left, review the current candidate instead of another fix

When the budget leaves one Result, the run reviews the current candidate and records its limits, and commits no further fix; it stops with the limits recorded and the candidate reviewed.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate

hyp prompt's default rises (for example 5 to 7) and the skill states that a review is its own Experiment counting against the budget. In a trial the same effect is --experiments set to the Results the original run had left plus 2.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's own probe passes on the candidate closes in the review's Result

A case that a probe run in the review ran and passed on the candidate is closed in that review's Result, since a check ran it, and is not recorded as a limit. A limit stays for a case no check ran or one the candidate fails, so no fix commit follows a passing case.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the last review before the select reads the cumulative diff from the Goal's base

The review before the select reads the whole diff from the Goal's base, not the latest increment, so the final code is read once against the lines. Expected to lose on cost; that is a prediction, not a rule.
