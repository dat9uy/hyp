# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Which review practice makes a budget-bound run review its final candidate?

Body:

```text
Discovery Goal for issue #141, part of #113: the last candidate of a budget-bound build run goes unreviewed. Its candidates are the rivals of one DesignQuestion, measured by replay; no rival is a commit here.
Method: a new recipe replays a live Sonnet runner (claude -p, Agent tool allowed) from three cuts of this store: after R-8JQPWG on b4f23b2, after R-3YJ7T2 on 2284172, after R-JHBCWS on 3cfd08d. Each replay has a checkout with no later commit or event, the original actor, the prompt from hyp prompt with --experiments at the rival's N, and the rival's skill text overlaid. 8 rivals with 3 trials each per cut, 72 trials, kept as evidence and not counted as Results. Lines 1, 3 and 4 are judged blind to the rival; line 2 by script. Line 5: a trial costs total_cost_usd; a cell is the mean of its 3 trials; a hung rival trial counts as the highest completed cost at that cut; a hung baseline trial makes the cut inconclusive and its cell reruns once; a ratio from 1.3 to 1.7 is inconclusive. A split of 2 trials against 1 is inconclusive. The known answers must show total_cost_usd includes subagent spend. H-whole-candidate is expected to lose on cost; its rule says so. The build change to hyp-goal-run and the stop clause is a later Goal.
May change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.
Never: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.
Never: replace an installed hyp binary; compile the candidate to a separate path.
Never: edit another investigations/ tree or a frozen recipe version.
Never: write under an owner's actor, or merge a PR.
```

Lines:

1. [N] A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. "Reviewed" means a Result under an Experiment whose cited evidence holds a review of that commit's diff by an actor other than the runner. "Named" means the stop checkpoint's summary holds the commit's short hash and the word "unreviewed". A trial that hangs has no checkpoint and fails this line.
2. [N] A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was given.
3. [N] After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replayed run records no further review before the select.
4. [N] A replayed run never leaves a finding of its review without a limit, a close, a fix commit or a named out-of-Goal finding. It never records a Result that says a case passed where the cited probe log shows it failing.
5. [N] A replayed practice other than the baseline never costs more than 1.5 times the baseline at the same cut. The baseline is H-status-quo: skills/hyp-goal-run/SKILL.md at 9c8e498, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e. H-status-quo costs exactly 1.0 times the baseline, so it passes by definition.

## DesignQuestion in investigations/goal-prompt, by grill-141: Which review practice lets a budget-bound run review its final candidate, without a review chain over small changes?

Every rival answers this one question. Scenario that separates them: replay a Sonnet runner from the cuts after R-8JQPWG, R-3YJ7T2 and R-JHBCWS (3 trials per rival per cut) and grade lines 1 to 5. Rivals combine only as the named H-combo; the verdicts of the single rivals show which hold alone. No rival may add a stored field or event: whether a candidate is reviewed cannot be derived from the log, so it stays a convention.

## Hypothesis in investigations/goal-prompt, by grill-141: H-status-quo: review is its own counted Experiment; findings become limits then fix commits; budget 5

The baseline: hyp-goal-run at 9c8e498 (sha256 cb15e2f2...) with hyp prompt's default of 5 Results. Must go red on line 1 or 3 at the cuts, or the instrument has no known answer.

## Hypothesis in investigations/goal-prompt, by grill-141: H-name-at-stop: the stop checkpoint names each candidate without a review

Keep the practice and the budget; the stop checkpoint summary lists each commit with no review Result, with its short hash and the word unreviewed. Holds line 1 only by naming.

## Hypothesis in investigations/goal-prompt, by grill-141: H-one-experiment: a candidate's review is part of its gate Experiment

One Experiment and one Result carry the test verdicts and the review's limits, so a candidate costs one Experiment again. The text states how the review is recorded and that it does not count apart.

## Hypothesis in investigations/goal-prompt, by grill-141: H-reserve-last: with one Result left, review the current candidate instead of committing another fix

When the stop clause leaves one Result, the runner reviews the current candidate, records its limits and stops, and does not commit another fix.

## Hypothesis in investigations/goal-prompt, by grill-141: H-raise-budget: the default of --experiments rises to 7

Only the stop clause default changes, to 7, to fit a gate and a review per candidate. The text of the practice stays as in H-status-quo.

## Hypothesis in investigations/goal-prompt, by grill-141: H-probe-closes: a case the run's probe runs and passes closes in the review's Result, with no commit

A case the run's own probe ran and the candidate passed is closed in the review Result, citing the probe log. A limit stays only for a case no check ran or one the candidate fails. Includes the owner's rule R14: a case whose check passes closes as a limit, not with a new commit.

## Hypothesis in investigations/goal-prompt, by grill-141: H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base

The review reads the cumulative diff from the base, not the latest increment. Expected to lose on cost (line 5); the Experiment's rule says so before the replay.

## Hypothesis in investigations/goal-prompt, by grill-141: H-combo: name-at-stop, reserve-last and probe-closes together

The practice texts of H-name-at-stop, H-reserve-last and H-probe-closes together, so the verdicts show which hold alone.
