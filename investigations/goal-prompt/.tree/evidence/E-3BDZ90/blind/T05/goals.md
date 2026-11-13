# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal, and a Goal body holds the run's constraints

Body:

```text
Builds #113 (core): `hyp prompt` prints the /goal prompt for any Goal from the log, and a Goal body ends with the run's constraints block. The design is the selected rival of this store's design record. The candidate is a hyp commit; go test decides lines 1 to 7. The descriptions of hyp-goal-run and hyp-prototype name one kind of Goal: they are texts this change makes wrong, edited under May change, with no line. Test skill text with ak-skill-creator's eval flow, update mode: original against candidate by sha256, fresh Sonnet agents, 3 trials, at most 48 runs: 6 routing prompts (natural, indirect, near-miss negatives, fixed holdout) and 2 functional cases for line 8, plus quick_validate and lint_cruft. Cases live in internal/skillcheck/testdata; records under .tree/evidence. Fixture for line 3: the log of investigations/checkpoint-store up to the owner's answer of 2026-11-11 14:31, its Goal body given a block. Clause list for line 3 (run 6's text is not in the repository; rebuilt from run 5): Run the hyp tree in <store> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G>. Done when `hyp status <G>` shows `runner owes nothing: stop and report to the owner` under OWED, or <N> Results are recorded under <actor> in <G>, or you cannot take the next step honestly and say why. Then the block, then the branch and PR clauses. Review each candidate commit once with a fresh subagent given the lines, the code and the settled keys. Sort each finding: a case of a line is a limit; a question about a line's meaning is a reading; a defect in text the run may edit is a fix commit; anything else is an out-of-Goal finding.
May change: commits on one local branch of the hyp repository, with a PR that is never merged; the texts the change makes wrong (internal/, cmd/, skills/, docs/, AGENTS.md, README.md)
Read only: the other trees under investigations/
Never: merge a PR
Never: replace the installed hyp binary; build to a temporary path
```

Lines:

1. [P] `hyp prompt G --actor A` on a Goal whose body ends with a valid constraints block exits 0 and prints, from the log alone and in this order: the store directory relative to the repository root, A, G's canonical id, the skill hyp-goal-run, three stop clauses (`hyp status G` shows `runner owes nothing: stop and report to the owner` under OWED; N Results are recorded under A in G, 5 by default; the runner cannot take the next step honestly), then the block's lines unchanged.
2. [P] `--results N` sets N; `--branch B` and `--pr N` each add one clause after the block and change no other text. The same log and flags print the same bytes twice, and `hyp prompt` appends no event and writes no file.
3. [P] On the fixture named in the Goal body, the prompt equals the body's clause list clause for clause. It differs only in the actor, the owner's answer, the branch, the PR, the commit, the run's own plan, and wording no step depends on.
4. [N] `hyp prompt` prints no prompt and exits 1, naming the cause, when the body has no valid block (a missing heading or the first bad line), when the prompt would exceed 4000 characters, or when A already has a Result in G (choose a new actor). It exits 2 for a bad argument (a non-Goal or unknown id, an invalid actor, a --branch with a space) and 3 when no repository holds the store.
5. [N] `hyp add goal` is refused with exit 1, naming the rule, and appends nothing, when its body has a line that starts with May change:, Read only: or Never: but no valid block, a near-miss heading (other case, a leading space, a space before the colon) or a carriage return in the block. A body with no heading line is written, and replay accepts every body.
6. [N] A body with a block that `hyp add goal` accepts is never refused by `hyp prompt` for its block, and a block `hyp prompt` refuses is never accepted by `hyp add goal`.
7. [N] In a store with two Goals, the prompt for G names only G's id and G's block; a Result under A in the other Goal never counts toward G's N, and a retracted Result under A in G still counts.
8. [P] hyp-goal-run states the build practice (branch, a PR never merged, one review round per commit, how a finding is sorted) and hyp-owner-grill's hand-off names `hyp prompt` and no longer has the skill write the prompt; on matched functional cases a fresh Sonnet consumer under the candidate meets each case's assertions at least as often as under the base hash 91009f6320f2fa5ca12958a106a4d38f81815dca10bae00b2f79a80792bd868c of hyp-owner-grill, and the candidate's hash differs from it.

## Goal in investigations/goal-prompt, by owner-standin: Record the /goal prompt design: who writes it

Design record for #113 (owner flow). Question: does hyp print the /goal prompt from the log, or does a skill write it? The owner settled it in the grill session of 2026-11-12: hyp prints it (`hyp prompt`). No run checks this Goal's candidates, so it has no lines; its cases go to the lines of the Goal that builds the design.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Who writes the /goal prompt for a Goal: hyp from the log, or a skill?

Today the repo skill hyp-owner-grill has the skill write the prompt by hand, and nothing prints it. The prompt must serve any Goal and carry its constraints block. Scenarios that separate the rivals: (1) generate the prompt twice from one log and compare bytes; (2) generate it for a body whose block is invalid and observe the exit status; (3) generate it for a Goal in a two-Goal store.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: hyp prompt prints the /goal prompt from the log

A command reads the log and prints the store directory, the actor, the Goal id, the skill, the stop clauses and the Goal body block unchanged. Same log and flags give the same bytes; an invalid block exits 1; it writes nothing.

## Hypothesis in investigations/goal-prompt, by claude-grill-113: A skill writes the /goal prompt from hyp show and hyp status

The owner-flow skill composes the prompt text by hand from the tree and copies the block lines. The text varies between sessions, and no exit status refuses an invalid block.

## Experiment in investigations/goal-prompt, by claude-grill-113: Settle the prompt form by the three scenarios, in the grill session

Rule, fixed before the Result: the owner settled in the session of 2026-11-12 that hyp prints the prompt (rival 1). Rival 2 is contradicted when any scenario cannot hold for a prompt a skill writes: (1) the same log gives the same bytes; (2) an invalid block is refused with exit 1; (3) a two-Goal store gives only the asked Goal. No run executes these; the scenarios file records the argument.

## DesignQuestion in investigations/goal-prompt, by claude-grill-113: Which hyp commit implements the #113 owner flow: hyp prompt and the constraints block?

The candidate is a hyp commit. The run adds each commit as a Hypothesis with its subject, and a fix as a new Hypothesis that supersedes it.
