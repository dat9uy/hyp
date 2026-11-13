# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: How a build-Goal run reviews its candidates within a cap: rival practices replayed from two cuts

Body:

```text
Discovery Goal for issue #141, part of #113. Core scope: the run practice of a build Goal in hyp-goal-run (review, cap, stop report) and the stop clause hyp prompt prints. Aims: a review must not easily run out of the run's cap, and must not drag on over small changes. Four DesignQuestions hold the rivals, each with today's practice as a rival: record (H-one-experiment), cap (H-name-at-stop, H-reserve-last, H-raise-budget), probe (H-probe-closes), scope (H-whole-candidate). Each rival is replayed alone by a Sonnet runner (claude-sonnet-5-5): 7 Arms, 2 cuts, 3 trials each, at most 50 trials, 30 minutes per trial, in ak-skill-creator's eval flow. An Arm is a patch over skills/hyp-goal-run kept in a new recipe, never a commit to skills/. A trial prints its prompt inside its sandbox, works only in the cut store at the sandbox root, and never pushes. Combinations of the rivals that hold, and the build of the selected practice, are later Goals. Evidence: the owner's comments on #141.
May change: this repository, in commits on the branch fix/goal-run-review, with one PR the run may open or update and never merges; investigations/goal-prompt (its tree and evals), a new versioned recipe under recipes/ with its Arm patches and inputs, and docs/ where the new recipe makes a row wrong.
Read only: skills/, internal/, cmd/, every investigations/ tree except goal-prompt, ak-skill-creator under .claude/skills, and the commits b4f23b2 and 3cfd08d.
Never: commit to main or merge a PR.
Never: edit skills/, internal/ or cmd/ in this repository, or replace ~/.local/bin/hyp.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Never: let a trial push, or read or write the committed investigations/goal-prompt/.tree in its sandbox.
```

Lines:

1. [P] On cut 1 (store cut at event 01M75T1YRF2R2Z99DX0E8JQPWG, base b4f23b2, actor claude-g113-1, cap 5) a trial ends with a review Result for its final candidate, or with a stop checkpoint that names that candidate by Hypothesis id or commit as unreviewed. A review Result cites the report of a fresh subagent that read the candidate's diff. The final candidate is the Hypothesis of the last commit on the trial's branch that changes a file outside investigations/.
2. [P] On cut 2 (store cut at event 01M76VXNN2ZW5WAXJ194JHBCWS, base 3cfd08d, actor claude-g113-3, cap 5) a trial ends with a review Result for its final candidate, or with a stop checkpoint that names that candidate as unreviewed, as line 1 defines them. A trial that adds no commit meets this line when the cut's review of 3cfd08d stands.
3. [N] A select is recorded on a candidate that has no review Result, in the trial's store or its cut.
4. [N] A trial stops at its cap with a candidate that has no review Result and that its stop checkpoint does not name by Hypothesis id or commit.
5. [N] On a cut, an Arm's median trial cost is more than 1.5 times the baseline Arm's median on that cut. Cost is a trial's tokens: input, cache_read, cache_write, output and reasoning summed as ak-skill-creator's session_usage.py (sha256 863c4f39…) counts them. The median is the middle of the 3 sorted costs. A trial with no stop after 30 minutes costs the highest completed cost on that cut and fails lines 1 to 4. Baseline: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e at cap 5; each Arm's text or cap differs from it.
6. [N] After a review whose findings are all cases that probes pass on the candidate (probes of the trial, or ones recorded in the cut's store), the trial's branch gets a new commit.
7. [N] A trial's actor has more Results recorded than its Arm's cap (5; 7 for H-raise-budget), counting the Results already in the cut.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: How is a candidate's review recorded, and does it count against the run's cap?

Today a review is its own Experiment and Result (runs 1 and 2), so a candidate costs a gate Result and a review Result. Scenario: replay cut 1 (R-8JQPWG) and cut 2 (R-JHBCWS) under each Arm; count Results recorded under the actor at the stop and whether each final candidate has a review Result.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What does the run do as its cap nears with a candidate unreviewed?

Run 1 stopped at 5 Results with 5c5902d unreviewed and unnamed; run 2 stopped with 3cfd08d unreviewed and named. Scenario: replay cut 1, where 4 of 5 Results are spent and a fix is pending (3 are left at cap 7); see whether the trial ends with the final candidate reviewed or named, and its cost against the baseline.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What does a review finding become when the run's own probe passes the case?

Today a finding that is a case of a line is a limit, so run 3's N1 and P1, which the code passed, each led to a fix commit and a new review. Scenario: replay cut 2, after R-JHBCWS; see whether the trial adds a commit after a review whose findings all pass their probes, and the chain's cost.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What does the review before the select read: the last increment or the whole candidate?

Since 10566d4 each review read only the latest fix diff, so the code at the final candidate was never read in one piece against the lines. Scenario: replay cut 1 and cut 2; see which diff the trial's fresh reviewer is given, whether findings differ, and the cost.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Baseline: a review is its own Experiment and Result

Today's practice: skills/hyp-goal-run/SKILL.md at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, cap 5, no patch. A review is recorded as a separate Experiment (recipe none, rubric reading) whose Result counts against the cap.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: the review is part of the candidate's gate Experiment

One Result carries the gate's test verdicts and the review's limits, so a candidate costs one Result. Arm: a patch over SKILL.md that says so; cap 5.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Baseline: no rule for the last Result of the cap

Today's practice: skills/hyp-goal-run/SKILL.md at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, cap 5, no patch. Nothing says what yields when one Result is left: the cap or the review.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: the stop checkpoint names each unreviewed candidate

Arm: a patch over SKILL.md that makes the stop checkpoint name each candidate with no review Result by Hypothesis id and commit; cap 5. Holds alone or with any other rival.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Result left the run reviews the current candidate

Arm: a patch over SKILL.md that, with one Result left, makes the run review its current candidate instead of committing another fix, record its limits, and stop; cap 5.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the cap is 7 Results

Arm: no text patch; hyp prompt --experiments 7 on the same cuts. Owner's comment of 2026-11-12 argues a fixed cap ends unreviewed anyway; the owner chose to test it.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Baseline: a finding that is a case of a line is a limit

Today's practice: skills/hyp-goal-run/SKILL.md at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, cap 5, no patch. A case the run's own probe passes is still recorded as a limit and closed by a test commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's probe passes is closed in the review's Result

Arm: a patch over SKILL.md. A case a probe ran and passed on the candidate is closed in the review's Result, not recorded as a limit; a limit stays for a case no check ran or one the candidate fails. Cap 5.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Baseline: the review reads the latest increment

Today's practice: skills/hyp-goal-run/SKILL.md at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, cap 5, no patch. The fresh reviewer is given the diff of the last commit against its parent.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the review before the select reads the cumulative diff

Arm: a patch over SKILL.md. The review before the select reads the diff from the Goal's base to the candidate, not the last increment. Cap 5.
