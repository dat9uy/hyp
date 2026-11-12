# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run's constraints

Body:

```text
Builds `hyp prompt` for issue #113: hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run's constraints under May change:, Read only: and Never:. The candidate is a hyp commit. go test is the instrument, and skillcheck decides the skill text.
Done means, by the owner on 2026-11-11: for a Goal with a valid block, the prompt carries every clause the hand-written run-5 prompt had, except the actor, the owner's answer, the branch, the PR, the commit and the run's own plan. A test pins this on a synthetic Goal under internal/cli/testdata. checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it and no test compares it.
The session settled that hyp prompt is a command that reads the log, and that no skill writes the prompt.
May change: the hyp repository, by commits on the local branch feat/goal-prompt and then a PR that the run opens and never merges
May change: cmd, internal, the docs and a new ADR 0022
May change: skills, for the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal, the hand-off of hyp-owner-grill, and a build-Goal review round in hyp-goal-run
Read only: the existing ADRs, every other tree under investigations, and recipes
Never: replace the installed hyp binary
Never: edit an existing ADR or a Goal
Never: merge a PR or close #113 (write Part of #113)
```

Lines:

1. [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each ending in a newline and in this order: the store's absolute root, A, the Goal's display id, `hyp-goal-run`, the stop clauses (the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; it cannot take the next step honestly and says why), then the block's lines byte for byte. With no flag there is no branch or PR clause.
2. [P] `--experiments N` sets N. `--branch B` and `--pr N` each add exactly one clause after the block, and each clause only narrows what the block allows. The output is the same for any order of the flags.
3. [N] A body's block runs from its first line that starts with `May change:`, `Read only:` or `Never:` to the end, and is valid only if all three headings appear, each has text after it, and every non-empty line starts with one (headings are case-sensitive; a line of spaces, a tab or a lone carriage return is not empty). For a body with no valid block, `hyp prompt` exits 1, names the missing heading or the bad line, and prints no prompt on stdout.
4. [P] `hyp add goal` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no heading adds without a warning. Replay and `hyp check` report no error for either Goal.
5. [N] `hyp prompt` writes nothing: the bytes of `.tree` are the same before and after. It carries no owner answer and no tree state beyond the Goal's id and body.
6. [N] These exit 2: a missing or invalid actor, a branch that git refuses (such as one with a space), `--pr` or `--experiments` below 1, a repeated flag, an unknown node, and a node that is not a Goal. A run where no `.tree` is found exits 3.
7. [N] In a store with two Goals, `hyp prompt` reads only the named Goal's body: the other Goal's block never appears in the output, and a Goal without a valid block is refused while the other Goal prints.
8. [N] `hyp prompt` exits 1, naming the actor, when the actor already has a Result under that Goal.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Where is the /goal prompt produced?

Rivals: a hyp command that reads the log, or a skill that writes the prompt from hyp show.
Scenario that separates them: one Goal whose body block has CRLF line ends, given to both writers. The command prints the block byte for byte or refuses it with exit 1; a skill-written prompt can drift from the body and has no exit code to refuse with.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt is a command that reads the log

hyp prompt G --actor A prints the prompt from the Goal's body and the log, writes nothing, and exits 1 when the body has no valid constraints block.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A skill writes the prompt from hyp show

The owner-grill skill reads hyp show for the Goal and writes the prompt text itself, copying the constraints lines by hand.

## Experiment in investigations/goal-prompt, by claude-grill-113: Settle Q-where by the session's scenarios

Rule, fixed before the record: no run is made; the owner settled this in the session of 2026-11-12 and the session is the evidence. Scenarios in the evidence file: (1) a Goal body whose block has CRLF line ends; (2) a Goal body whose block has a bad line; (3) an actor that already has a Result. The command side gives each scenario a fixed output or an exit code. The skill side has no exit code and copies text by hand. H-command is supported when the command side gives a defined result in all three; H-skill is contradicted when the skill side gives none in at least one.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Which hyp commit implements hyp prompt?

The candidate is a hyp commit on feat/goal-prompt: the prompt command, the warning on hyp add goal, and the skill and doc text the Goal's May change names. A run adds each commit as a Hypothesis whose subject is the base plus that commit's change; a fix is a new Hypothesis that supersedes it. Each Experiment's verdict rule asks for red on base and green on the candidate.

## Goal in investigations/goal-prompt, by owner-standin: Design record: hyp prompt is a command that reads the log

Records the one design question of issue #113 that the owner settled by argument in the grill session of 2026-11-12: where the /goal prompt is produced. The rivals are a hyp command that reads the log and a skill that writes the prompt from hyp show. The session is the evidence; no run was made. It has no lines: no run checks its candidates, and its cases are the lines of the build Goal G-VRG6SA, which builds this design.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Where is the /goal prompt produced?

Rivals: a hyp command that reads the log, or a skill that writes the prompt from hyp show.
Scenario that separates them: one Goal whose body block has CRLF line ends, given to both writers. The command prints the block byte for byte or refuses it with exit 1; a skill-written prompt can drift from the body and has no exit code to refuse with.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt is a command that reads the log

hyp prompt G --actor A prints the prompt from the Goal's body and the log, writes nothing, and exits 1 when the body has no valid constraints block.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A skill writes the prompt from hyp show

The owner-grill skill reads hyp show for the Goal and writes the prompt text itself, copying the constraints lines by hand.

## Experiment in investigations/goal-prompt, by claude-grill-113: Settle q-produced by the session's scenarios

Rule, fixed before the record: no run is made; the owner settled this in the session of 2026-11-12 and the session is the evidence. Scenarios in the evidence file: (1) a Goal body whose block has CRLF line ends; (2) a Goal body whose block has a bad line; (3) an actor that already has a Result. The command side gives each scenario a fixed output or an exit code. The skill side has no exit code and copies text by hand. h-cmd is supported when the command side gives a defined result in all three; h-skl is contradicted when the skill side gives none in at least one.
