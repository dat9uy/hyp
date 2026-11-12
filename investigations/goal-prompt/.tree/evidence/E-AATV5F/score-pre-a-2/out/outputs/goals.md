# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal from its log, and warns when a Goal's constraints block is unusable

Body:

```text
Builds #113: hyp prompt prints the /goal prompt for any Goal from the log, and hyp add goal warns about a constraints block no prompt could use. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Constraints live in the Goal body under three headings and no typed field is added (owner, 2026-11-11). A prompt may narrow them, never widen them. Out of scope: what a commit-candidate Goal needs from hyp (#114), one root store versus one per topic, and matching run 6's prompt (a placeholder issue).

May change: this repository, as commits on one local branch
May change: the texts the change makes wrong, by directory: skills/, docs/, AGENTS.md, README.md, including a skill description that names only one kind of Goal
Read only: investigations/checkpoint-store and investigations/checkpoint-store-design
Never: open or merge a PR
Never: replace the installed hyp binary; build to a temporary path
Never: edit a Goal or an event already in a log
```

Lines:

1. [P] For a Goal whose body ends in a valid constraints block, `hyp prompt G --actor A` exits 0 and prints in order: the store path; the actor; the Goal id; the skill name hyp-goal-run; the stop clauses (the runner owes nothing; N Results recorded under the actor, non-retracted and on this Goal, N from --experiments, default 5; you cannot take the next step honestly and say why); the block, line for line in body order. --branch and --pr each add one clause after the block. Two runs over one log print identical bytes.
2. [P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Goal has checkpoint-store:G-D1CH1R's title and lines and a body ending in a block of run 5's constraint lines, the output matches the baseline clause for clause and differs from it only in: actor, owner answer, branch, PR, commit, the run's own plan; the dropped add-rival-Hypotheses clause; Results for Experiments in the count clause; and run 5's Treat-as-legacy sentence, which prints as its Read only: line.
3. [N] `hyp prompt` exits 1, with a message naming the rule and nothing on stdout, for an unknown id, a node that is not a Goal, and a Goal body without a valid block. A valid block starts at the first line that begins at column 0 with May change:, Read only: or Never:, runs to the end of the body, holds all three headings, and every non-empty line in it starts at column 0 with one. A line of spaces, a tab or a lone carriage return is not empty.
4. [N] `hyp add goal` warns (a warn: line on stderr, exit 0, naming the rule) for exactly the bodies the previous line makes `hyp prompt` refuse, a body with no heading included. It never refuses the add, and a log holding such a body replays with `hyp check` reporting 0 errors.
5. [N] `hyp prompt` exits 2 for an --actor outside the replay-spec name pattern or missing; a --branch that is empty, starts with -, or has a character outside A-Za-z0-9._/-; and a --pr or --experiments that is not a positive integer.
6. [P] The output depends only on the merged projection. In a two-Goal store, `hyp prompt G-a` names only G-a's id and block and points to `hyp status G-a`; with the Goal's add and its checkpoints in two actors' files it prints the same bytes as when the events share one file.
7. [P] For a Goal that is complete, or whose OWED already says the runner owes nothing, `hyp prompt` still prints and exits 0.
8. [N] With no store (.tree/events) at or above the working directory, `hyp prompt` exits 3 and prints nothing on stdout.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: What form of store path does hyp prompt print?

Scenario that separates the rivals: create a git worktree and a clone at other paths, run hyp prompt in each, and check that each printed path resolves to the store holding the Goal. Each rival is a commit that meets all eight lines and differs from the other only in this form. The owner does not pick between them.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: Print the absolute path of the store root

hyp prompt prints filepath.Abs of the directory holding .tree/events, as store.FindRoot returns it. It needs no git. Used from a worktree or clone at another path it prints that checkout's own path.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: Print the store path relative to the repository root; with no git toplevel, exit 3

hyp prompt prints the store root relative to the git toplevel that holds it, so the same text names the store in any clone. With no git toplevel above the store it exits 3 (an environment problem).
