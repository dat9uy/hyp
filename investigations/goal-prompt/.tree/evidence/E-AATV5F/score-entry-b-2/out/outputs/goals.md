# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: Keep a commit-candidate run's last candidate reviewed or named, without review chains

Body:

```text
For issue #141, part of #113. A run of a Goal whose candidate is a commit must end with its last candidate reviewed or named unreviewed in the stop checkpoint, and a review must neither run out of the run's budget nor chain on small changes. The candidate is a hyp commit: skills/hyp-goal-run, internal/cli/prompt.go and the texts they make wrong. One DesignQuestion holds six rivals, tested by replays on Sonnet, 3 trials per arm from two cuts, run 1 after R-8JQPWG (actor claude-g113-1) and run 2 after R-3YJ7T2, against the base at 9c8e498 (line 6 pins its sha256). H-whole-candidate is the expected loser. Combinations with H-name-at-stop run only for singles that fail line 2. Method: ak-skill-creator's eval flow; recipes/hyp-goal-run-eval@1 pins the grader, a price table, the hang cap min(3x the longest base trial, 30 min) and the median. A review Result is a Result on the candidate that cites a reviewer's report of its diff. This run names its own unreviewed candidates in its checkpoint.
May change: this repository, in commits on one local branch and at most one PR never merged: cmd/, internal/ and their testdata, and the texts the change makes wrong: skills/, docs/, AGENTS.md, README.md; one new bundle recipes/hyp-goal-run-eval@1. Installing a candidate skill or an arm's hyp into a sandbox is not a change to .claude/skills or the installed binary.
Read only: every investigations/ tree except goal-prompt (its cuts are read, never written back), plans/, and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or ~/.local/share/hyp/hyp-33896ec; build each arm's hyp only at a path of its own inside its sandbox.
Never: edit an existing recipes/ version, rubrics/, another investigations/ tree, ak-skill-creator or another repository.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [P] At each cut (run 1 after R-8JQPWG, run 2 after R-3YJ7T2), the candidate's final commit has a review Result in at least 2 of 3 trials.
2. [N] No trial ends with a final candidate that has no review Result and is not named as unreviewed in the stop checkpoint.
3. [P] Every finding of a review at the cut is, by the stop, fixed in a commit, closed by a Result that ran its case, referred, or named in the stop checkpoint.
4. [P] Where every finding's case already passes the run's own probe on the candidate, at least 2 of 3 trials make no new candidate commit after that review.
5. [N] No trial makes more than 2 new candidate commits from a cut; at that bound the run names its open findings in the stop checkpoint and does not fix a third time.
6. [P] At each cut the candidate's median trial cost is at most 1.5 times the base's median at the same cut, the base being skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e with internal/cli/prompt.go sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed; the candidate differs from it. A trial past the hang cap costs infinity.
7. [P] The stop clause hyp prompt prints and hyp-goal-run agree on whether a review counts against the budget and which yields when they meet: prompt_test.go asserts the exact clause sentence, a check asserts the phrase in hyp-goal-run that says how a review is recorded, and the clause holds for a Goal with no commit candidate.
8. [P] Each changed SKILL.md passes ak-skill-creator's quick_validate.py and lint_cruft.py --routing with no High finding, skillcheck passes, go test ./... -count=1 passes, and the G-ZBM2RE#6 eval case still passes.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: What change keeps a budget-bound run's last candidate reviewed or named, without review chains?

Separating scenarios: replay from cut 1 (run 1, claude-g113-1, after R-8JQPWG: one Result left, a fix commit pending) and cut 2 (run 2, after R-3YJ7T2: one Result left, both probes already pass), 3 trials per arm on Sonnet, against the base at 9c8e498. Measured: is the final candidate reviewed or named, how many new commits follow, and the cost against the base.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A candidate's review is part of its gate Experiment

One Result carries the test verdicts and the review's limits, so a candidate costs one Experiment. hyp-goal-run says how a review is recorded and that it does not count separately against the budget.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The stop checkpoint names each candidate without a review

The practice stays (a review is its own Experiment), and hyp-goal-run requires the stop checkpoint to name every candidate commit with no review Result, so the gap is visible to the owner. The floor of line 2; also tested alone.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: With one Experiment left, review the current candidate instead of committing another fix

When the budget allows one more Result, the run reviews the current candidate and stops with recorded limits, and a case whose check already passes closes as a limit and not with a new commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The default of --experiments rises to fit a gate and a review per candidate

internal/cli/prompt.go default 5 becomes a larger number (for example 7), with prompt_test.go and the prompt-fixture golden file changed with it. It breaks the completed G-ZBM2RE#1, which stays as written; the PR says so.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A case the run's own probe ran and passed is closed in the review's Result, not recorded as a limit

A limit stays only for a case no check ran or one the candidate fails, so a review that finds only passing cases leads to no fix commit. Run 3's cut after R-JHBCWS shows N1 and P1 closing on 3cfd08d and d2d5d14 with no commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The review before the select reads the cumulative diff from the Goal's base

The base is the cut's pinned commit, not main. Expected loser: one cumulative review should cost more than 1.5 times the base at each cut; its Experiment pre-registers that prediction and still runs the arm.
