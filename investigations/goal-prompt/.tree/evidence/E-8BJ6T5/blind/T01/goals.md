# Nodes the grill added

Stores at the end: ., internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/id-citation, investigations/recipe-base

## Goal in ., by owner-standin: hyp prints the /goal prompt for any Goal from the log, and checks the Goal body's constraints block

Body:

```text
Builds `hyp prompt` and the check of the Goal body's constraints block, as the installed hyp-owner-grill skill specifies them (issue #113; owner decision 2026-11-11: a run's constraints live in the Goal body, with no typed field). The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Updating skills/hyp-owner-grill, whose hand-off text still says to hand-write the prompt, and the description fixes of hyp-goal-run and hyp-prototype are texts this change makes wrong, so they fall under May change and are not lines.

May change: the hyp code and tests of this repository, as commits on one local branch
May change: skills/hyp-owner-grill, skills/hyp-goal-run and skills/hyp-prototype, including the hand-off text and the descriptions this change makes wrong
May change: the docs and specs this change makes wrong, in the same commits
Read only: the other investigation trees under investigations/ and every recipe under recipes/
Never: merge a PR
Never: replace the installed hyp binary
Never: edit a recipe version that is already frozen
```

Lines:

1. [P] 'hyp prompt <G> --actor A' on a Goal whose body ends in a valid constraints block exits 0 and prints, in this order, the store's path, the actor, the Goal's display id, the skill name hyp-goal-run, and the stop clauses: the runner owes nothing; N Results recorded under the actor, 5 by default; the run cannot take the next step honestly. Nothing else in the output varies by run except the flag values: it names no owner answer, no commit and no plan.
2. [P] The Goal body's block is printed byte for byte, line for line. After it come only a branch clause if --branch is given, then a PR clause if --pr is given. --experiments N sets N in the stop clause.
3. [N] hyp prompt refuses a Goal whose body has no valid block, and a prompt longer than 4000 runes: it exits 1, names the rule in its message and prints nothing to stdout. It never truncates.
4. [P] 'hyp add goal' with a body that has a line starting with 'May change:', 'Read only:' or 'Never:' but no valid block exits 0 and prints a warning that names the rule, by the same test hyp prompt uses. A body with a valid block, or with no such line, prints no warning.
5. [N] hyp prompt exits 2 for a missing or invalid actor ($HYP_ACTOR is the fallback; ^[a-z0-9][a-z0-9-]{0,39}$), a --branch that is empty, starts with - or holds a character outside letters, digits and . _ / -, a --pr or --experiments that is not an integer of at least 1, and a ref that is unknown, ambiguous or not a Goal. It exits 3 when no repository holds the store.
6. [P] The prompt is a function of the log and the flags: the same log and flags give the same bytes, an alias and an id of one Goal give the same bytes, hyp prompt writes no event and no file, and in a store with two Goals the output holds only the named Goal's block.
7. [P] A block runs from the first line that starts with a heading to the end of the body, and is valid when it holds each of the three headings at least once and every non-empty line starts with one. A heading may repeat and a blank line may be truly empty. A lowercase, indented or mid-line heading starts no block; a line of spaces, a tab or a lone carriage return, or any non-empty line that starts with no heading, makes the block invalid.

## DesignQuestion in ., by claude-prompt-1: Which hyp commit implements hyp prompt and the constraints block check?

Rivals are commits, each a Hypothesis tied to the base plus its change with hyp subject. They are added when they exist, because a one-rival Experiment is refused. A line is decided by go tests on the candidate.
