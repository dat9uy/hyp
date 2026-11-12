# Nodes the grill added

Stores at the end: internal/cli/testdata/prompt-fixture/investigations/checkpoint-store, internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: A budget-bound run ends with its last candidate reviewed or named, and a review does not loop

Body:

```text
Builds a change to the run practice of a Goal whose candidate is a commit, for issue #141 (part of #113): a run that reaches its budget has a review Result for its final candidate or names it unreviewed, and a review does not go on over small changes. The candidate is a hyp commit that changes skills/hyp-goal-run/SKILL.md and, where a rival needs it, the stop clause in internal/cli/prompt.go. One DesignQuestion holds six rivals tested against each other and the base; H-whole-candidate is the expected loser. A combination is added only if no single rival holds the lines. Method: replayed runs on Sonnet (claude-sonnet-5-5), 3 trials for each candidate and cut, matched to the base as ak-context-engineering says; no replay of a whole build. Each Experiment's verdict rule fixes the hang cap before the run.
May change: this repository, in commits on the branch feat/review-budget from main; it may push the branch and open or update one PR, and never merges. It changes skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go with its tests and golden, and the texts the change makes wrong: skills/hyp-owner-grill/SKILL.md, docs/how-hyp-improves-itself.md, and the docs that state the stop clause.
Read only: plans/, every investigations/ tree except goal-prompt, ak-skill-creator under .claude/skills, and ak-context-engineering in ~/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering (SKILL.md sha256 ad2ce02b31b8d0595a934267ce726b489044443711f75622a7199e13e88d6383; references/evaluation.md sha256 fdd4a34ff3d1b937ad482a031160b8b9da40f73a20244258d6d38e77ed8d2d0b).
Never: commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or ~/.local/share/hyp/hyp-33896ec; compile the candidate to a separate path.
Never: edit recipes/, rubrics/, another investigations/ tree, ak-skill-creator, ak-context-engineering or another repository.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

Lines:

1. [P] At store cut R-8JQPWG with repository commit b4f23b2, under actor claude-g113-1 and the candidate's own hyp prompt default, each of 3 Sonnet trials ends with a review Result for its final candidate or a stop checkpoint that names that candidate as unreviewed. The baseline skill (sha256 a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab) does neither in at least one of 3 trials, and the candidate differs from it.
2. [P] At store cut R-JHBCWS with repository commit 3cfd08d, under actor claude-g113-3, where the review case N1 passes its probe, each of 3 Sonnet trials ends with the select of H-FQ103R and records no later candidate commit. The baseline skill (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) records a later commit in at least one of 3 trials, and the candidate differs from it.
3. [N] From either store cut, no trial records more than 2 fresh-reviewer runs before the select or the stop.
4. [N] At either store cut, a candidate's median trial cost exceeds 1.5 times the base's median, or it has 2 or more hangs in 3 trials. Cost is input-equivalent tokens from the cut to the stop, subagent turns included, weights input 1, cache-write 1.25, cache-read 0.1, output 5; a hang costs infinity; if the base median is infinite the verdict is inconclusive.
5. [P] hyp-goal-run states how a review is recorded (which Experiment, under which actor) and whether it counts against the stop clause's Results. An eval case gives a consumer only the skill text, and its answers to both questions match the selected rule.
6. [P] go test ./..., go vet ./... and skillcheck pass on the candidate. The cases of G-ZBM2RE lines 1 to 5 and 7 pass, except the default number of Results in the stop clause where the selected rule changes it; the test and golden change with it.

## DesignQuestion in investigations/goal-prompt, by claude-grill-141: Which change makes a budget-bound run end with its last candidate reviewed or named, with no review loop?

Scenario: replay from store cut R-8JQPWG (repository b4f23b2, actor claude-g113-1; one Result left under the base, a fix pending) and from cut R-JHBCWS (repository 3cfd08d, actor claude-g113-3; case N1 passes its probe), 3 Sonnet trials for each rival and the base. A rival is a commit that changes hyp-goal-run and, where needed, the stop clause. The base is the baseline of lines 1, 2 and 4, not a rival. Rivals can combine, only if no single rival holds the lines; the tree says which hold alone. The stop clause counts Results, not Experiments.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A candidate's review is part of its gate Result

One Result carries the gate's test verdicts and the review's limits, so a candidate costs one Result again. The stop clause counts Results, so a second Result under the same Experiment would not save budget.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The stop checkpoint names each candidate without a review

Keep the practice of a gate and a review per candidate. The stop checkpoint names every candidate that has no review Result, so the owner sees the gap. This rival gives line 1 its floor, and it does not make a review reachable.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: With one Result left, review the current candidate instead of committing another fix

When one Result remains, the run reviews the current candidate instead of committing a fix, records the findings as limits, and stops. The run does not end on an unreviewed commit.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: Raise the default of --experiments to fit a gate and a review per candidate

hyp prompt prints a higher default, for example 7. The owner argued from run 2 that a fixed budget still ends unreviewed while each review leads to a commit; this rival tests that argument.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: A case the run's own probe ran and passed closes in the review's Result, not as a limit

A case that the run's own probe runs and passes on the candidate is closed in the review's Result, because a check ran it. A limit stays for a case no check ran or one the candidate fails. No commit follows a passing case.

## Hypothesis in investigations/goal-prompt, by claude-grill-141: The review before the select reads the cumulative diff from the Goal's base

The review reads the whole candidate against the lines, not the latest increment. The owner expects it to lose on cost; the run tests that and does not assume it.
