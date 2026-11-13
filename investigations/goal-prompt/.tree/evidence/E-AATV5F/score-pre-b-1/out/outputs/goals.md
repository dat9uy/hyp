# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Choose the review, budget and stop practice of a build Goal run, so its final candidate is reviewed or named

Body:

```text
Discovery Goal for issue #141 (Part of #113). The candidate is a practice: a variant of skills/hyp-goal-run/SKILL.md and of the stop clause in internal/cli/prompt.go. One DesignQuestion holds six rivals: H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes and H-whole-candidate. A combination is a rival only if a single rival fails a line. The build of the selected practice is a later Goal the owner adds.
Method: ak-skill-creator's eval flow. A trial replays from a cut of this store, on Sonnet, 3 trials per practice and cut, never a full build. Cuts: after R-8JQPWG (repository at b4f23b2; one Experiment left), after R-3YJ7T2 (2284172; both probes pass), after R-JHBCWS (3cfd08d; findings are passing cases). A trial runs under the original run's actor, so Results already recorded count against the budget. It takes its skill text and hyp binary from the practice under test, and only the code and store from the cut. The base is an Experiment with no target under the DesignQuestion; every closer cites it. Every line runs on every cut it names.
May change: investigations/goal-prompt (.tree, evidence, evals, and arms/ for candidate patches), in commits on the branch the prompt names; the run may push it and open or update one PR, and never merges. Nothing under skills/, internal/, docs/ or cmd/: a practice is a patch in arms/, applied only in a trial sandbox.
Read only: this repository outside investigations/goal-prompt, plans/, other investigations/ trees, and ak-skill-creator under .claude/skills, pinned by sha256.
Never: give a trial runner any event, file or Goal of this store after its cut.
Never: commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or any installed hyp; compile a candidate to a separate path.
Never: edit recipes/, another investigations/ tree, ak-skill-creator or another repository.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [N] On each cut (after R-8JQPWG, R-3YJ7T2, R-JHBCWS), a run stops on its budget with a candidate commit that has no review Result and is not named as unreviewed in the stop checkpoint.
2. [P] On the cut after R-3YJ7T2 and on the cut after R-JHBCWS, where each review's findings are only cases the run's own probe passes on the reviewed candidate, the run ends with its final candidate holding a review Result.
3. [P] The candidate's text states how a review is recorded (its own Experiment, or inside the gate's) and whether its Result counts against the budget.
4. [N] The prompt hyp prints and the skill's review rule, read together, do not say which yields when the budget and a review of the final candidate cannot both be met, or do not say that the stop checkpoint names the unreviewed candidate.
5. [N] On each cut, a candidate's mean cost per trial is more than 1.5 times the base's mean at the same cut. Base: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e with hyp sha256 c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c; a candidate's SKILL.md or hyp sha256 differs from these. Cost is dollars from the usage of the runner and every subagent it starts; a hung trial (over 60 minutes or 150 turns) costs the highest of any trial at that cut.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: Which practice reviews a build run's final candidate within budget, without chaining on small changes?

Today: hyp prompt stops a run at 5 Results per actor, and hyp-goal-run asks one review per candidate. --limit exists only on hyp record, so a review became its own Experiment: a gate plus a review cost two. Runs 1 to 3 of G-ZBM2RE (issue #141) ended with the last candidate unreviewed. Separating scenario: replay each practice from the cuts after R-8JQPWG (one Experiment left, a fix pending), R-3YJ7T2 (both probes pass) and R-JHBCWS (a review finds only passing cases), 3 trials each, and read how each run stops. Each rival is a whole practice.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: a candidate's review is part of its gate Experiment

One Result carries the test verdicts and the review's limits, so a candidate costs one Experiment again. Run 3 did this unprompted for 04142f3 (E-GE085A).

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: keep the practice; the stop checkpoint names each unreviewed candidate

The gap is visible to the owner. Run 2 did it without a rule; the scenario also checks whether it holds when the runner has no earlier gap to learn from. It must hold alone.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Experiment left, review the current candidate, not another fix

The run reviews the candidate it has, stops with the findings recorded as limits, and names what it did not fix. On run 2's cut both probes passed on 2284172, so a Result closing them followed by the select would have ended reviewed.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate

An edit to the default in internal/cli/prompt.go, for example 5 to 7, so the hyp sha256 differs from the base (a bare --experiments flag would not). Owner's note: run 2 had a fresh budget of 5 and still ended unreviewed, because each review led to a new commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's own probe passes is closed in the review's Result, not a limit

A limit stays for a case no check ran, or one the candidate fails. N1 and P1 of run 3 close on 3cfd08d and d2d5d14 and no commit follows. The rest of the budget text stays as the base.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base

Since 10566d4 every review read only the last fix. Owner's expectation: this rival loses, because one cumulative review costs a lot; a run still gives its verdict. The rest of the budget text stays as the base.
