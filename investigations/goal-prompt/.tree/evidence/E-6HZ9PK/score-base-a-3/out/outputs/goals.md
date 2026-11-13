# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal from the log

Body:

```text
Builds issue #113: a `hyp prompt` command that prints, from the log, the /goal prompt for any Goal, and a check of the constraints block in a Goal body. Who produces the prompt is the design-record Goal in this store. The candidate is a hyp commit; go test is the instrument, with a synthetic fixture under internal/*/testdata.
Owner decisions, 2026-11-11: a run's constraints live in the Goal body under May change:, Read only: and Never:; a /goal prompt may narrow them, never widen them; a typed field waits until a real run misreads or loses one.
Gap: run 6's prompt (goal-prompt-261111-2129-g1b-run-6) was not found, so no line compares the output with it. A later Goal adds that line, with these allowed differences: the actor, the owner's answer, the branch, the PR, the commit and the run's own plan.
checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it.
If the run changes skill text, test it with the ak-skill-creator eval and audit flow: Sonnet test agents, 3 trials, a modest budget; the run decides the details. No line gates skill text.
May change: the hyp repository, as commits on one local branch, never merged; events in investigations/goal-prompt under the runner's own actor; the texts the change makes wrong: skills/hyp-owner-grill/SKILL.md (its hand-written 4000-character prompt step), the description lines of skills/hyp-goal-run/SKILL.md and skills/hyp-prototype/SKILL.md that name only one kind of Goal, and the docs that describe the prompt step, the new command or its exit codes
Read only: every other tree under investigations/, docs/adr/ and the .claude directory
Never: replace the installed hyp binary
Never: open or merge a PR
Never: edit or delete an event or a Goal in any tree
Never: put application code or real source data in this repository
```

Lines:

1. [P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the actor, the Goal id, the skill name hyp-goal-run, the stop clauses (the runner owes nothing under that Goal's OWED; or N Results on that Goal, recorded under the actor and not retracted, N from --experiments, default 5), the block copied line for line, then one clause for --branch and one for --pr if given. Two runs on the same log print the same bytes. It prints no owner answer.
2. [P] In a store with two Goals, the prompt for one names only that Goal's id and constraints block and counts only that Goal's Results, and equals the prompt for the same Goal in a store holding only its own events.
3. [N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule and what the block lacks (a missing heading, or the first line that does not start with one), and prints nothing to stdout.
4. [N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body, holds each heading at least once, and every line in it is non-empty and starts with a heading. A line of spaces, a tab or a carriage return is not empty, so a `\r` in the block makes it invalid. `hyp add goal` with a body that holds a heading but no valid block prints a warning that names the rule, exits 0 and still writes the Goal.
5. [N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or no actor. It exits 3 when no repository holds the store, and 1 for an unknown Goal id. Each prints a message and no prompt.

## Goal in investigations/goal-prompt, by owner-standin: Record who produces a Goal's /goal prompt

Design record for issue #113. Question: does hyp print a Goal's /goal prompt from a command that reads the log, or does the grill skill write it. Settled in the session of 2026-11-12: hyp prints it. The prompt carries no owner answer, because the runner reads OWED. The build is the other Goal in this store.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Who produces a Goal's /goal prompt?

Scenarios that separate the rivals: (a) the same log read twice gives the same bytes; (b) a Goal whose body has no valid constraints block gets no prompt and a refusal; (c) in a store with two Goals, only the asked Goal's items appear.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A hyp command prints the prompt from the log

hyp prompt <G-id> --actor <A> reads the log and prints the prompt, so scenarios (a), (b) and (c) are checked by go test.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: The grill skill writes the prompt from hyp show and hyp status

The grill skill composes the prompt as prose each session; no command reads the log for it.

## Experiment in investigations/goal-prompt, by claude-grill-113: Session decision on who produces the prompt

Rule, fixed before the record: the owner settled this question in the session of 2026-11-12, so this Experiment is an argument, not a run. H-BM50HC supports when a command that reads the log can meet scenarios (a), (b) and (c) of the question. H-QZMAYS contradicts when prose written per session cannot guarantee (a) or (b): two sessions can word it differently (a), and no exit code can refuse a Goal with no block (b). Evidence: scenarios.md in this Experiment's evidence directory.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Which hyp commit implements hyp prompt and the constraints-block check?

The candidate is a hyp commit that makes every line of this Goal hold. Rival commits are added by the run.
