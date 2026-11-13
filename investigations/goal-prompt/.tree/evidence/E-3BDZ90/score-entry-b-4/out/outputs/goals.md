# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: A budget-bound run reviews its last candidate without a review-and-fix chain that goes on

Body:

```text
Discovery Goal for issue #141 (part of #113). A run of a build Goal that reaches its Result budget has reviewed its last candidate commit, and its reviews do not go on because of small changes. The candidates are six changes to the run practice (skills/hyp-goal-run and the stop clause hyp prompt prints), each a Hypothesis under one DesignQuestion that names its cut. A cut is the store as of a Result, made with hyp sandbox create --store --cut; the run first checks that --base hides the commits after the cut. A trial replays a Sonnet consumer from the cut with the arm's rule, as the cut's real actor so the budget left is the real one, in a clone with no later commit and without the text of issue #141. When the stop clause fires, a new run starts with hyp prompt as printed, a new actor and a fresh budget of the same size. An independent grader on the strongest model reads each trial. Each arm runs 3 trials at each of its cuts, and the control (the unchanged hyp-goal-run) runs 3 at the same cut. Test each rival alone first; if none holds every line, the run may add one combination Hypothesis. H-name-at-stop and H-whole-candidate carry no preset verdict. The build that changes skills/ and the prompt is a later Goal.
May change: the investigations/goal-prompt tree (.tree/events and .tree/evidence), in commits on the branch the prompt names; scratch stores and clones under $TMPDIR. The run may push that branch and open or update one PR, and never merges it.
Read only: skills/, docs/, internal/, cmd/, plans/, and every investigations/ tree except goal-prompt.
Never: edit skills/, docs/, internal/ or cmd/; that is the build Goal.
Never: replace an installed hyp binary.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Never: commit to main or merge a PR.
```

Lines:

1. [P] For each Hypothesis, replaying its named cut with its rule gives, in at least 2 of 3 trials, a run whose last candidate commit has a review Result at the first stop clause of that arm's own prompt (5 Results under the actor; 7 for the raise-budget rival). Naming the gap in the stop checkpoint does not satisfy this line. The control, skills/hyp-goal-run/SKILL.md at 9c8e498 (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), runs at the same cut and is no candidate.
2. [N] A Hypothesis costs more than 1.5 times the control at its cut. A trial costs the Results recorded from the cut until the select or until the runner owes nothing, across restarts; 15 Results without that, 45 minutes or a kill costs infinity. An arm costs the median of its 3 trials, infinity last. It fails if its median is infinity or 2 times it exceeds 3 times the control's; a finite arm passes an infinite control. Its skill text differs from the control's (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e).

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: Which change to the run practice keeps a review from being cut off by the budget or dragged out by small fixes?

Each rival is a change to hyp-goal-run or to the stop clause hyp prompt prints, and each names its cut: the store as of one Result, replayed with hyp sandbox create --store --cut. The control arm (unchanged hyp-goal-run) runs at the same cut and is not a Hypothesis. The Goal's lines are the cases. Cuts: run 1 after R-8JQPWG (actor claude-g113-1, 1 Result left of 5); run 2 after R-3YJ7T2 (actor claude-g113-2, 1 left of 5); run 3 after R-JHBCWS (actor claude-g113-3, 4 left of 5); final after R-BSWFXG (04142f3 is the candidate, its review limit open).

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-name-at-stop: the stop checkpoint names each candidate commit that has no review

Keep the practice. At the stop, the checkpoint names each candidate commit without a review Result. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2. Expected by the owner: line 1 contradicts it alone, since naming is not a review; it stays a part of any combination. No preset verdict.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-reserve-last: with one Result left, review the current candidate instead of committing another fix

When one Result remains, the runner reviews the current candidate instead of committing another fix, and stops with the review's limits recorded. A case whose check already passes closes as a limit, not with a new commit. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-one-experiment: a candidate's review is part of its gate Experiment

One Experiment carries the gate verdicts and the review, and one Result carries the test verdicts and the review's limits, so a candidate costs one Result again. hyp-goal-run states how a review is recorded and that it counts as part of the gate. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-raise-budget: the default of --experiments rises to 7

The stop clause hyp prompt prints defaults to 7 Results under the actor, to fit a gate and a review per candidate. The trial prompt uses --experiments 7. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-probe-closes: a case the run's own probe passes is closed in the review's Result, not recorded as a limit

A case that the run's own probe runs and passes on the candidate is closed in the review Result, since a check ran it. A limit stays for a case no check ran or one the candidate fails. Cut: run 3 after R-JHBCWS, where N1 closes on 3cfd08d and no fix commit follows.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base

The review before the select reads the cumulative diff from the Goal's base to the candidate, not the latest increment. Cut: final, after R-BSWFXG, with 04142f3 as the candidate. The owner expects it to lose on cost; no preset verdict.
