# Nodes the grill added

Stores at the end: internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/goal-prompt, investigations/id-citation, investigations/recipe-base

## Goal in investigations/goal-prompt, by owner-standin: hyp prints the /goal prompt for any Goal, and a run's constraints have a checked place in the Goal body

Body:

```text
Build Goal for #113 (core scope), asked by the owner: hyp prints the /goal prompt for any Goal from the log, and a run's constraints have a checked place in the Goal body. The candidate is a commit in this repository. It covers: hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N, default 5]; the constraints-block check in hyp add goal; the descriptions of hyp-goal-run and hyp-prototype, and the build-Goal practice in hyp-goal-run (owner, 2026-11-11). Skill text is tested with the ak-skill-creator eval and audit flow, which decides cases, fixtures, matching and grading: test agents on Sonnet, 3 trials per case, a modest budget; a plain go test check covers the two descriptions.
Pending, not a line: the prompt equals run 6's prompt (goal-prompt-261111-2129-g1b-run-6) clause for clause, except the actor, the owner's answer, branch, PR, commit and the run's own plan. A line must name that file by sha256 and lines are never edited, so it goes into a new Goal when the file is found. The run must not invent it.
Outside this Goal: #114, #134, a typed constraints field, moving trees to a root store. Run 1 uses a hand-written prompt, because hyp prompt does not exist yet.
May change: the hyp repository as commits on one local branch, feat/goal-prompt, with a PR opened and never merged; the investigations/goal-prompt store; and the texts the change makes wrong, under skills/, docs/, AGENTS.md and README.md
Read only: every other directory under investigations/, recipes/, and .claude/
Never: merge a PR
Never: write to ~/.local/bin; build with go build -o "$TMPDIR/hyp" and run tests from there
Never: add a typed constraints field to hyp add goal
Never: take on #114 or #134
```

Lines:

1. [P] hyp prompt <G-id> --actor A names the store's absolute path, the actor, the Goal id and the skill hyp-goal-run, then prints the Goal body's constraints block line for line, unchanged, after the stop clauses; --branch and --pr each add one clause after the block.
2. [P] The stop clauses are scoped to the Goal: 'runner owes nothing' under that Goal's OWED, or N Results (default 5, --experiments N) recorded under the actor on questions that lead to that Goal; a Result under the actor on another Goal of the store is not counted, and a retracted Result still counts.
3. [N] hyp prompt exits 1, with a message that names the rule and nothing on stdout, for a Goal whose body has no valid constraints block and for a prompt over 4000 runes. The block runs from the first column-0 line that starts with 'May change:', 'Read only:' or 'Never:' to the end of the body; it holds each heading at least once, every non-empty line in it starts with a heading (a line of spaces, a tab or a lone CR is not empty), and a heading may repeat.
4. [N] hyp prompt exits 2 for an invalid actor, a --branch with a leading dash or a character outside [A-Za-z0-9._/-], a non-positive --pr or --experiments, or an id that is not a Goal; it exits 3 when no repository holds the store.
5. [N] hyp add goal exits 1, names the rule and appends nothing for a body with a column-0 heading line but no valid constraints block; a body with no column-0 heading line is added; a merged log that holds such a Goal replays without rejection (spec 5.1: text rules are write-time), and hyp prompt refuses it.
6. [N] hyp prompt appends no event and writes no file, and hyp add goal stores no field for the block: replay derives it from the body.
7. [P] The descriptions of skills/hyp-goal-run and skills/hyp-prototype each name a build Goal as well as a discovery Goal, and skills/hyp-goal-run states the build-Goal practice: the branch and a PR that is opened and never merged, one review round per candidate commit, and sorting each finding into a limit, a reading, a fix commit or an out-of-Goal finding.

## DesignQuestion in investigations/goal-prompt, by claude-113-1: Which hyp commit makes hyp print the /goal prompt and check the constraints block?

The build's one question; its candidate is a hyp commit on feat/goal-prompt, added by the run when one exists.
