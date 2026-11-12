# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Run practice: a budget-bound build run ends with its final candidate reviewed

Body:

```text
Chooses one change to the run practice of a build Goal (hyp-goal-run's Build Goals and hyp prompt's stop clause) so a budget-bound run ends with its final candidate reviewed, without running out of budget or dragging on over small changes. Issue #141. A rival is an Arm patch of skills/hyp-goal-run/SKILL.md or internal/cli/prompt.go, never a landed change. The base, today's skill with the default of 5, is the cost reference, not a rival. The owner adds Goal B, the build of the chosen rule, later.
Test: ak-skill-creator eval flow, Sonnet, 3 trials per arm per cut, replayed from three cuts of this store, never a whole build. A cut ends at the event next after R-8JQPWG (base b4f23b2), R-3YJ7T2 (base 2284172) or R-JHBCWS (base 3cfd08d), by event id. Each trial has a fresh actor, the prompt of hyp prompt G-ZBM2RE and a cap counted from zero. The replay recipe is a self-review instrument: the base on the first cut must first reproduce run 1's unreviewed stop. Cost is the median of 3 trials per arm and cut; a base median of +infinity makes the cut inconclusive for line 3. Line 1 must hold in all 3 trials, line 2 in 2 of 3. After the alone results the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3. With two or more survivors the owner chooses.
May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; it never merges a PR. Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.
Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace an installed binary; build an arm only inside its sandbox.
Never: write under an owner's actor; refer with a reason or record a reading, and stop.
Never: edit another investigations/ tree, ak-skill-creator or another repository.
```

Lines:

1. [N] On each cut, a run under the candidate rule never stops, a 30-minute timeout included, with a candidate commit that has neither a Result citing a fresh subagent's review of that commit's diff nor its short hash named as unreviewed in the stop checkpoint. A trial that never stops fails this line.
2. [P] On each cut, a run under the candidate rule stops with its final candidate reviewed (a Result citing a fresh subagent's review of that commit's diff) and with nothing in OWED of hyp status G-ZBM2RE for the runner, except a select that an open reading of the cut holds back. It does so within the --experiments cap its own prompt printed, counted from zero under a fresh actor.
3. [N] On every cut, the candidate's cost is never more than 1.5 times the base's cost on that cut. The base is skills/hyp-goal-run/SKILL.md with sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e and the default of 5; the candidate's patch differs from it. Cost is the median over 3 trials of billed USD with subagents; a trial that does not stop in 30 minutes, or has no cost reading, is +infinity.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: Which change to the run practice makes a budget-bound build run end with its final candidate reviewed, without running o

Scenario: replay the three cuts (after R-8JQPWG, R-3YJ7T2, R-JHBCWS), 3 Sonnet trials per arm each, and compare each arm with the base on the three lines. Issue #141 and its owner comments.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: A candidate's review is part of its gate Experiment: one Result carries the test verdicts and the revi

A candidate's review is part of its gate Experiment: one Result carries the test verdicts and the review's limits, so a candidate costs one Result. Patch: skills/hyp-goal-run/SKILL.md says the review of a candidate is recorded in its gate Experiment and Result.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: Keep the practice; the stop checkpoint names each candidate commit without a review, so the owner sees t

Keep the practice; the stop checkpoint names each candidate commit without a review, so the owner sees the gap. Patch: skills/hyp-goal-run/SKILL.md Stop section names each unreviewed candidate by short hash.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: With one Result left, the runner reviews the current candidate instead of committing another fix, and st

With one Result left, the runner reviews the current candidate instead of committing another fix, and stops with the limits recorded. Patch: skills/hyp-goal-run/SKILL.md reserves the last Result for the review.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: The default of --experiments rises from 5 to 7 to fit a gate and a review per candidate.

The default of --experiments rises from 5 to 7 to fit a gate and a review per candidate. Patch: internal/cli/prompt.go default 7, with its tests.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: A case the run's own probe runs and passes on the candidate is closed in the review's Result; a limit st

A case the run's own probe runs and passes on the candidate is closed in the review's Result; a limit stays for a case no check ran or the candidate fails. Patch: skills/hyp-goal-run/SKILL.md sorting rule for findings.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: The review before the select reads the cumulative diff from the candidate's subject base, not the lat

The review before the select reads the cumulative diff from the candidate's subject base, not the latest increment. The owner expects it to lose on cost; it is tested anyway. Patch: skills/hyp-goal-run/SKILL.md review step.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-limit-no-commit: A case whose check already passes closes as a limit with a Result, not with a new commit that only ad

A case whose check already passes closes as a limit with a Result, not with a new commit that only adds a test or spec text. Patch: skills/hyp-goal-run/SKILL.md sorting rule for findings.
