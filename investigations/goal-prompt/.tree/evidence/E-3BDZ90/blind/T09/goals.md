# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal from the log, and a run reads its constraints from the Goal body

Body:

```text
Issue #113 (core). hyp prompt prints the /goal prompt for any Goal from the log and copies the constraints block of the Goal body, so a run reads what it may change from the recorded place. A build Goal: the candidate is a hyp commit and go test is the instrument. The skill part (hyp-goal-run, hyp-prototype, hyp-owner-grill; ADR 0021) is tested with the ak-skill-creator eval and audit flow: cases kept in a recipes/<skill>-eval@N bundle, graded by eval_skill.py; test agents Sonnet, 3 trials, modest budget. ADR 0017 and docs/how-hyp-improves-itself.md say the cap is stated outside hyp; they change to: printed by hyp prompt, enforced by nothing.
Not met here: the first Done means bullet of #113 (the prompt matches run 6 clause for clause). Run 6 text is not available, so no line can run it. For the Goal that adds that baseline, the allowed differences are: the actor, the owner answer the run starts from, the branch, the PR, the commit, and the plan of the run.
May change: this repository, as commits on one local branch feat/goal-prompt
May change: a PR for that branch, opened and never merged
May change: cmd/ and internal/, with their testdata
May change: a new recipes/<skill>-eval@N bundle for the skill tests
May change: the texts the change makes wrong, by their directories: skills/, docs/, AGENTS.md, README.md
May change: investigations/goal-prompt/, through hyp commands only
Read only: every other store under investigations/
Read only: recipes/, apart from the new eval bundle
Never: replace the installed hyp binary
Never: edit a frozen recipe version
Never: merge a PR
```

Lines:

1. [P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and prints exactly: the store path, A, the Goal id and the skill hyp-goal-run; three stop clauses (runner owes nothing under OWED; N Results recorded under A in the tree of this Goal since the run began, N being 5 or --experiments; cannot take the next step honestly, and say why); the lines of the block verbatim and in order; then one clause each for --branch and --pr, each opening "Within the constraints above:". No flag alters a block line, and the log is unchanged.
2. [N] hyp prompt exits 1, prints nothing to stdout and names the rule in its message when the Goal has no valid block (a block runs from the first line starting May change:, Read only: or Never: to the end of the body, holds all three headings, and every non-empty line starts with one; the message names the missing heading or the first offending line number and its first characters), or when the output would be 4000 runes or more (the message names the limit and the length).
3. [P] hyp add goal with a body that holds no constraints block exits 0 and the log replays. With a heading present and the block invalid it also exits 0 and prints a warn: line naming the first offending line.
4. [N] In a store with two Goals, the prompt of one Goal never holds items of the other Goal, and it equals the prompt printed from a store holding only the events of that Goal.
5. [N] hyp prompt exits 2 and prints nothing to stdout for a --branch outside [A-Za-z0-9._/-]+, a --pr or --experiments that is not a positive integer, a missing --actor, and an id that is unknown or not a Goal. It exits 3 when no repository holds the store.
6. [P] When --actor equals the actor that added the Goal, hyp prompt prints the same stdout and exit 0 and writes one warn: line to stderr that names separation of duties.
7. [P] A consumer agent given the candidate hyp-owner-grill and hyp-goal-run skills and the candidate hyp hands off a Goal by running hyp prompt and gives its output unchanged, in each of 3 trials of the ak-skill-creator eval.
8. [P] The descriptions of hyp-goal-run and hyp-prototype route the run and the tree work of a build Goal to them as they route those of a discovery Goal: the ak-skill-creator eval has positive cases for both kinds of Goal and adjacent negative cases, and all pass.
9. [P] The ak-skill-creator audit of each changed skill reports no error-level finding.

## DesignQuestion in investigations/goal-prompt, by claude-g113-grill-1: Which hyp commit makes hyp print the /goal prompt for any Goal and read a run's constraints from the Goal body?

Candidate: a commit in this repository. Rival Hypotheses are the commits a run builds; the lines of the Goal are owed on each.
