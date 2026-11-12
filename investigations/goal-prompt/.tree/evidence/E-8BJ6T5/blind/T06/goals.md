# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints a run's /goal prompt from the log, and a Goal body records the run's constraints

Body:

```text
Builds #113: `hyp prompt` prints a run's /goal prompt from the log; the constraints block below is its read side. Owner decision 2026-11-11: constraints live in the body under May change:, Read only: and Never:, and a typed field waits until a real run misreads or loses one. Every block line needs text after its colon. One checker serves `hyp add goal` (warns, still appends) and `hyp prompt` (refuses, exit 1). hyp prints the run cap and never enforces it; ADR 0022 supersedes ADR 0017's cap line in part. Lines 8 and 9 are an ak-skill-creator eval of the hyp-owner-grill Hand off: 2 cases, 3 Sonnet trials each, candidate and base skill, graded by eval_skill.py. The description edits have no line and no routing traces run: a known gap. A commit Goal's other needs from hyp are #114.
May change: this repository (hyp): commits on one local branch
May change: one pull request, opened and never merged
May change: internal/cli and internal/replay for the command and the shared block checker
May change: the descriptions of hyp-goal-run and hyp-prototype, which name one Goal kind
May change: skills/hyp-owner-grill Hand off, and the block definition in each skill that states it
May change: README.md, docs/glossary.md, docs/replay-spec.md, docs/core-and-self-review.md, docs/how-hyp-improves-itself.md
May change: a new docs/adr/0022, and a new recipes/hyp-owner-grill-eval@1
Read only: investigations/checkpoint-store and the other published trees
Read only: the ak-skill-creator skill, named on the eval command's External: line and never copied
Never: replace the installed ~/.local/bin/hyp
Never: edit a stored Goal or event
Never: add a typed constraints field to `hyp add goal`
Never: store a Result count, an actor baseline or a parsed constraint field
```

Lines:

1. [P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run the hyp tree in "<store path>" with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is G.` / `First run `hyp status G` there; if no such Goal shows, stop.` / `Done when the latest `hyp status G` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under A, or you cannot take the next step honestly and say why.` / `At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.` / the block, line for line.
2. [P] `--results N` changes only the 5. `--branch B` adds exactly `Commit only to branch B.` and `--pr P` adds exactly `Do not merge PR #P.`, each as one line after the block, and nothing else changes.
3. [P] The output depends only on the merged log and the flags: the same bytes for any order of the actors' events files. `hyp prompt` appends no event and changes no file. The Results counted are the actor's recorded Results under this Goal; a retracted one does not count.
4. [N] A Goal with no valid block (no heading, a missing heading, a bare heading with no text after its colon, or a non-empty line in the block that starts with no heading) makes `hyp prompt` exit 1, name the missing heading or the first bad line, and print nothing on stdout. CRLF line ends read as LF.
5. [P] `hyp add goal` with a body that has no heading, or a heading but no valid block, appends the Goal, exits 0 and prints a `warn:` line that names the rule and the first bad line. A body with a valid block prints no warning. One checker serves both commands.
6. [N] Bad arguments exit 2: an invalid actor, a `--pr` that is not digits, a `--branch` with whitespace, a control character or a leading `-`, a `--results` below 1, and an id that is not a Goal. No repository holding the store exits 3. `hyp prompt -h` exits 0 with no store.
7. [N] `hyp prompt` for an actor that already wrote any event in the store exits 1, names the rule and tells the owner to choose a new actor.
8. [P] At the candidate, each of 3 fresh Sonnet trials of the hyp-owner-grill Hand off, given a finished grill whose Goal has a valid block, ends with the output of `hyp prompt` for that Goal byte for byte and adds no sentence.
9. [N] At the candidate, no one of 3 fresh Sonnet trials of the Hand off, given a Goal that `hyp prompt` refuses, writes a prompt by hand. Each tells the owner of the refusal.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Who produces a run's /goal prompt?

Scenario: a Goal with a malformed block, and two calls on one merged log. A command can exit 1 naming the rule and give the same bytes twice; a skill-written prompt cannot refuse with an exit code.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: What does hyp add goal do for a body with no valid constraints block?

Scenario: add the malformed Goal, then ask for its prompt. A Goal is never edited, so a refused add leaves no unusable Goal, and an allowed add leaves one that hyp prompt refuses.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: What does hyp prompt do for a Goal whose body has no valid constraints block?

Scenario: checkpoint-store:G-D1CH1R, a stored Goal with no block, in a store. A prompt with no constraint lines reads as permission to a run.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A hyp command prints the prompt from the log

hyp prompt reads the merged log and the flags, refuses a Goal with no valid block with exit 1, and gives the same bytes on every call.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A skill writes the prompt by hand

The grill skill writes the prompt from the tree in prose; no exit code can refuse a malformed block and two writers give two texts.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt refuses a Goal with no valid block, exit 1, headings named

Nothing goes to stdout; the owner adds a new Goal with a valid block.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt prints the prompt with no constraint lines

A prompt that lists no limits reads as permission to an unattended run.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt prints the prompt with a line 'no constraints recorded'

The run sees that nothing was recorded, but still starts.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp add goal warns at exit 0 and still appends the Goal

The warning names the rule and the first bad line; adding stays allowed.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp add goal refuses a body with no valid block, exit 1

Nothing is appended, so no Goal exists that hyp prompt would refuse.

## Goal in investigations/goal-prompt, by owner-standin: Design record for #113: who prints a run's /goal prompt, and what hyp does on a Goal with no valid constraints block

Records three questions the owner settled by argument in the grill session of 2026-11-12, with the session as evidence. G-N77C5E builds the design and carries its cases as lines. No run checks these candidates, so this Goal has no lines.
May change: this Goal's tree events and evidence, to record the session's decisions
Read only: the repository
Never: run this Goal as an unattended /goal

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Who produces a run's /goal prompt?

Scenario: a Goal with a malformed block, and two calls on one merged log. A command can exit 1 naming the rule and give the same bytes twice; a skill-written prompt cannot refuse with an exit code.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A hyp command prints the prompt from the log

hyp prompt reads the merged log and the flags, refuses a Goal with no valid block with exit 1, and gives the same bytes on every call.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A skill writes the prompt by hand

The grill skill writes the prompt from the tree in prose; no exit code can refuse a malformed block and two writers give two texts.

## Experiment in investigations/goal-prompt, by claude-grill-113: Scenarios settled in the grill session: gen

Rule: the rival the scenarios support is supported, the other contradicted. The owner settled it in session: only a command can refuse with an exit code and give one text; a skill-written prompt can do neither.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: What does hyp prompt do for a Goal whose body has no valid constraints block?

Scenario: checkpoint-store:G-D1CH1R, a stored Goal with no block, in a store. A prompt with no constraint lines reads as permission to an unattended run.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt refuses a Goal with no valid block, exit 1, headings named

Nothing goes to stdout; the owner adds a new Goal with a valid block.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt prints the prompt with no constraint lines

A prompt that lists no limits reads as permission to an unattended run.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt prints the prompt with a line 'no constraints recorded'

The run sees that nothing was recorded, but still starts.

## Experiment in investigations/goal-prompt, by claude-grill-113: Scenarios settled in the grill session: noblock

Rule: the rival the scenarios support is supported, the others contradicted. The owner settled it in session: a prompt with no limits reads as permission, so a Goal with no block is refused with exit 1 and the missing headings named.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: What does hyp add goal do for a body with no valid constraints block?

Scenario: add the malformed Goal, then ask for its prompt. A refused add leaves no unusable Goal; an allowed add leaves one that hyp prompt refuses, and a Goal is never edited.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp add goal warns at exit 0 and still appends the Goal

The warning names the rule and the first bad line; adding stays allowed, as the owner decided.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp add goal refuses a body with no valid block, exit 1

Nothing is appended, so no Goal exists that hyp prompt would refuse.

## Experiment in investigations/goal-prompt, by claude-grill-113: Scenarios settled in the grill session: addwarn

Rule: the rival the scenarios support is supported, the other contradicted. The owner settled it in session: adding a Goal with no valid block stays allowed, and the add warns.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Which hyp commit implements #113?

Each commit is a Hypothesis whose subject is the base plus that commit's change; a fix is a new Hypothesis that supersedes it. The run adds the rivals. The design it builds is recorded in G-HEDGTC, selected on Q-5XQQZ5, Q-JV0W60 and Q-25NHCT.
