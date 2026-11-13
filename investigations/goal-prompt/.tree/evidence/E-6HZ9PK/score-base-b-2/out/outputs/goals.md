# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Select the run practice that ends a budget-bound build run on a reviewed final candidate, without endless review

Body:

```text
Discovery Goal for issue #141 (part of #113), from runs 1 to 3 of G-ZBM2RE. The candidate is a run practice: a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never a commit. The owner adds the build Goal that lands the winner after the select.
S1: the store cut after R-8JQPWG, repository at b4f23b2 (SKILL.md sha256 a6152711c2c930b2…, prompt.go sha256 674dc260b4863afd…), actor claude-g113-1, one Result left of 5. S2: the cut after R-JHBCWS, repository at 3cfd08d (SKILL.md sha256 cb15e2f2e207b5a6…, prompt.go sha256 21f9332383bc6b0e…), actor claude-g113-3, 4 left.
A trial replays only from a cut: a Sonnet run with the prompt the candidate own hyp prints for the Goal and that actor (no --experiments flag), reading hyp-goal-run from the patched sandbox clone. Every candidate patch states how a review is recorded and whether it counts. Stages: the base, the singles, then H-name-at-stop with each single that fails line 1 alone; one Experiment per stage. Record every ratio. If H-review-owed needs a stored field, state it and add none.
May change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).
Never: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.
Never: edit a frozen recipe or another investigations/ tree.
Never: add an event kind or a stored field to the log.
Never: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that commit as unreviewed. A trial that hangs, or whose checkpoint says all review limits are closed or the runner owes nothing while a candidate has no review, does not meet it. At S1 the base (the files pinned in the body) meets it in at most 1 of 3 trials, so the candidate differs from it; at S2 no baseline is compared.
2. [N] Aim B. From S2, in each of 3 trials a run under the candidate launches no more than 1 fresh review before it stops. The base (the S2 files pinned in the body) meets this in at most 1 of 3 trials, so the candidate differs from it.
3. [N] Aim B. At S1 and at S2 the median cost of the candidate over 3 trials never exceeds 1.5 times the base median at the same start. Cost is USD from the usage fields of the whole trial, subagents included, at one dated Sonnet rate table. A hang or an unknown cost is +infinity; the median of 3 is the second smallest; if either median is +infinity the line is not met.
4. [N] Aim A. A run under the candidate ends with no review finding that is a case of a Goal line left neither closed, referred, fixed by a later commit nor named in its stop checkpoint.
5. [P] Aim A. A Sonnet run given only the candidate hyp-goal-run text and its printed prompt answers, in each of 3 trials, how a review is recorded and whether it counts toward the stop clause count, as the candidate text states. The base text at b4f23b2 and at 3cfd08d states neither, so the candidate differs from it.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What run practice ends a budget-bound build run on a reviewed final candidate, without endless review?

Separating scenarios: S1 (cut after R-8JQPWG: one Result left, a fix commit pending) and S2 (cut after R-JHBCWS: reviews that found no defect still led to commits). Replay each from the cut with 3 Sonnet trials per candidate and for the base, and compare how each run ends, how many fresh reviews it launches and what it costs. Each rival is a scratch patch; the Goal body pins the base.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: a candidate's review is part of its gate Experiment

One Experiment and one Result carry the gate's verdicts and the review's limits, so a candidate costs one Result again.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: the stop checkpoint names each candidate without a review

Keep a review as its own Experiment; the stop checkpoint names every candidate with no review, so the gap is visible. Also tested combined with each single that fails line 1 alone.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Result left the run reviews the current candidate, not another fix

The run stops with the review's findings recorded as limits, not with a new commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate

A scratch change to prompt.go's default (for example 7). Run 2 argues against it: each review so far led to a new commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's own probe passed on the candidate is closed in the review's Result, not a limit

A limit stays for a case no check ran or one the candidate fails. Run 3's N1 and P1 would close on 3cfd08d with no further commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base

The owner expects it to lose on cost, since one cumulative review is large; it is tested anyway.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-review-owed: hyp status OWED lists a review for a candidate that has none

A scratch hyp binary derives the owed review from the log. A gate and a review look alike (--recipe none --rubric reading); if the log cannot tell them apart without a new stored field, the run states the field and refers it, and adds none.
