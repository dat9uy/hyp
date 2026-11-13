# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal from the log and reads a run's constraints from the Goal body

Body:

```text
Builds #113 (core): the /goal prompt that hyp prints from the log for any Goal, and the run's constraints it copies from the Goal body. Design: the Goal prompt-design in this store (a command that reads the log; no owner answer in the prompt, because the runner reads OWED). The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Line 1b's baseline is the run-5 prompt of #113, kept as a fixture in a synthetic store under internal/cli/testdata.
Block rules, for docs/: the block runs from the first line that starts with a heading to the end of the body; headings match in exact case; all three appear; every non-empty line starts with one; a line of spaces, a tab or a lone CR is not empty; text before the block is not copied. hyp prompt and hyp add goal share one parser. Each run takes a new actor. --results counts applied Results under the actor on the named Goal, retracted included, default 5; the skill text says --experiments, so the skills change with it.
Skill text (ADR 0021): hyp-goal-run says "an unattended discovery /goal" and hyp-prototype says "Investigate an uncertain design", though both serve any Goal; hyp-owner-grill hands off with hyp prompt once it exists. Test the descriptions by ak-skill-creator's routing comparison (references/testing-and-iteration.md): the description at the base commit against the candidate's, in the real catalog; natural positives (one a /goal prompt for a build Goal), indirect requests, near-miss negatives, a fixed holdout; never name the target skill in a routing prompt; report TP, FP, FN, TN. Sonnet test agents, 3 trials, a modest budget; results stay outside the repository.
May change: this repository, as commits on one local branch, and the texts the change makes wrong: skills/, docs/, AGENTS.md, README.md
Read only: investigations/ and recipes/ of this repository
Never: merge a PR
Never: replace an installed hyp binary
Never: edit a written Goal or a frozen recipe version
```

Lines:

1. [P] For a Goal whose body ends in a valid constraints block and that has a live top-level Question, `hyp prompt G --actor A` prints, in order, the store path, A, G, the skill hyp-goal-run, the stop clauses, the block line for line, one fixed sentence (if two clauses disagree, follow the narrower one and name the conflict in the stop checkpoint), and one narrowing clause per flag.
2. [P] On the synthetic fixture Goal, the printed prompt differs from the baseline, the run-5 prompt of issue #113 (6 lines, LF, final newline, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3), in exactly: the actor, the owner answer, the branch, the PR, the commit, the run's plan, the dropped add-rival-Hypotheses clause, the Results count, the commands that name the Goal, and the fixed sentence.
3. [P] The stop clauses are: OWED owes the runner nothing, read with `hyp status G`; N Results recorded under the actor on G (applied events only, retracted Results counted, default 5, set by `--results`); and `you cannot take the next step honestly and say why`.
4. [N] A Goal whose body has no valid constraints block makes `hyp prompt` exit 1 with a message naming what is missing (a heading, or a non-empty line that starts with no heading), and prints no prompt. Such a Goal is still accepted by `hyp add goal` and by replay.
5. [N] `hyp prompt` exits 1 with a message naming the rule, and prints no prompt, for a Goal with no live top-level Question and for a prompt of more than 4000 characters.
6. [P] `hyp add goal` with a body that has a heading but no valid block appends the Goal, exits 0 and prints a warning; with a valid block, or with no heading, it prints none.
7. [N] A flag clause never widens the block: its words are only or never, never may. A `--branch` with a space, or a `--results` that is not a positive integer, exits 2; a store that no repository holds exits 3.
8. [N] In a store with two Goals, `hyp prompt G` counts Results and reads OWED for G only: a Result under the same actor on the other Goal is not counted, its OWED items are not read, and the prompt names `hyp status G` and `--goal G` for checkpoints.

## Goal in investigations/goal-prompt, by owner-standin: Record where the /goal prompt is generated: a hyp command that reads the log

Design record for #113 (core): where is the /goal prompt generated? Settled by argument in the session of 2026-11-12, with owner-standin as the evidence: a hyp command that reads the log, `hyp prompt`, and not a skill that writes the prompt from the tree. The prompt carries no owner answer, because the runner reads OWED. This Goal has no lines: no run checks its candidates. Its cases are the lines of the build Goal in this store, alias prompt-build.

## DesignQuestion in investigations/goal-prompt, by claude-gp-grill-1: Where is the /goal prompt generated?

Today a person writes the /goal prompt by hand from the tree (runs 3 to 5 of the multi-tenant tree changed it in three places only). Scenario that separates the rivals: print the prompt ten times for one fixture Goal, then print it for a Goal whose body has no valid constraints block.

## Hypothesis in investigations/goal-prompt, by claude-gp-grill-1: A hyp command reads the log and prints the prompt

hyp prompt G --actor A reads the log and the Goal body and prints the prompt. The same log gives the same bytes, and a Goal without a valid block gets exit 1.

## Hypothesis in investigations/goal-prompt, by claude-gp-grill-1: A skill writes the prompt from the tree

A skill reads hyp status and the Goal body, then writes the prompt as text. Its wording can differ between runs, and it has no exit code to refuse with.

## Experiment in investigations/goal-prompt, by claude-gp-grill-1: Scenarios argued in the session of 2026-11-12

Rule, fixed before the record; scenarios S1 to S3 are in the evidence file. h-command supports if a command meets S1 to S3. h-skill contradicts if a skill fails S1 or S2. No run happened: the evidence is the session, so the tree says the question was argued, not tested.

## DesignQuestion in investigations/goal-prompt, by claude-gp-grill-1: Which hyp commit implements the /goal prompt?

The candidate is a hyp commit. The run adds each commit as a Hypothesis; a fix is a new Hypothesis that supersedes it. Scenario that separates commits: go test on the base and on each commit, against the lines of prompt-build.
