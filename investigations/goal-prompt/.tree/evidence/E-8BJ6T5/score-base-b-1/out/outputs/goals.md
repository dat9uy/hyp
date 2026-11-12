# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Decide how a run reviews each candidate commit of a build Goal within its budget

Body:

```text
Decide, by replaying two cuts of this store, which rule text lets the run of a build Goal keep every candidate commit reviewed within its Result budget, without reviews that continue because of small changes. Part of #113; the event and the rivals are in issue #141. The candidates are rule texts for skills/hyp-goal-run, given as throwaway patches over its text at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and the --experiments default. The status quo is the baseline. The run ends at the select; the build of the chosen rule is another Goal.

Trials: the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending). Each rival gets 3 trials on each cut, and the status quo gets 3 on each cut. Each trial uses a fresh actor, runs on Sonnet from the cut and its candidate commit, and stops on the prompt clause or after 60 minutes. No trial replays a build from its start. The cost of a trial is the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any trial on that cut that did not hang. The cost of a rule on a cut is the median of its 3 trial costs, taken for each cut alone. The drive.json token and dollar cost is reported and decides nothing.

May change: this tree's events and evidence in investigations/goal-prompt/.tree, and scratch sandboxes and trial actors
Read only: skills/, internal/, docs/ and the other trees of this repository
Never: edit skills/ or docs/, replace an installed binary, push a branch, open or merge a PR
```

Lines:

1. [N] On the cut after R-8JQPWG or the cut after R-JHBCWS, a run under the rule stops with a candidate commit that has no review Result, and its stop checkpoint does not name that candidate as unreviewed.
2. [P] The rule text states how a review of a candidate commit is recorded (which Experiment, which Result, which verdict) and whether that review counts against the stop clause.
3. [N] A review Result closes as passing a case that no check ran, or a case that the candidate fails.
4. [N] A run stops with a finding of a review that is not sorted: not a limit, not a close by a check that ran its case, not a fix commit, not a reading, and not an out-of-Goal finding named in the stop checkpoint.
5. [P] On the cut after R-JHBCWS, when every finding of a review of 3cfd08d is a case that a probe the run itself ran already passes on 3cfd08d, the run closes those cases by that probe and writes no further candidate commit.
6. [P] On the cut after R-8JQPWG, the run stops with its final candidate commit having a review Result.
7. [N] A candidate rule that differs from the status-quo rule (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) and costs more than 1.5 times its cost on the same cut is admitted. Cost is the median of 3 trials, each the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any non-hung trial on that cut.

## DesignQuestion in investigations/goal-prompt, by grill-141: How does a build Goal run keep every candidate reviewed within its budget, without reviews that go on over small changes

Scenario: replay the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending); 3 trials per rival per cut, a fresh actor each, on Sonnet. A rival is a throwaway patch over skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), or for raise-budget the --experiments default. The status quo is the baseline. A combination of rivals is added as a rival when the select needs it.

## Hypothesis in investigations/goal-prompt, by grill-141: Status quo: a review is its own Experiment and counts against the budget (baseline)

Keep skills/hyp-goal-run as it is at sha256 cb15e2f2. A review of a candidate is its own Experiment (--recipe none --rubric reading) with its own Result, and that Result counts in the stop clause. The baseline for line 7.

## Hypothesis in investigations/goal-prompt, by grill-141: One experiment: a candidate's review is part of its gate Experiment

One Result carries the test verdicts and the review's limits, so a candidate costs one Experiment. Patch: the skill says to record the review in the gate Experiment's Result, not in a separate Experiment.

## Hypothesis in investigations/goal-prompt, by grill-141: Name at stop: the stop checkpoint names each candidate without a review

Keep the practice. Patch: the skill asks the stop checkpoint to name each candidate commit with no review Result, so the gap is visible to the owner. Holds alone, and may combine with any rival.

## Hypothesis in investigations/goal-prompt, by grill-141: Reserve last: with one Experiment left, review the current candidate instead of another fix

Patch: when one Result remains in the stop clause, the run reviews the current candidate instead of committing another fix, records the findings as limits, and stops.

## Hypothesis in investigations/goal-prompt, by grill-141: Raise budget: the default of --experiments rises to fit a gate and a review per candidate

No skill patch. The trial's prompt uses a higher --experiments, for example 7, in place of the default 5 of hyp prompt.

## Hypothesis in investigations/goal-prompt, by grill-141: Probe closes: a case the run's own probe passes on the candidate is closed in the review's Result

Patch: a case that a probe the run itself ran passes on the candidate is closed in the review's Result, not recorded as a limit and not followed by a commit. A case no check ran, or one the candidate fails, stays a limit.

## Hypothesis in investigations/goal-prompt, by grill-141: Whole candidate: the review before the select reads the diff from the Goal's base (expected loser)

Patch: the review before the select reads the cumulative diff of the candidate from the Goal's base, not the latest increment. The owner expects it to lose on cost, one cumulative review being large.
