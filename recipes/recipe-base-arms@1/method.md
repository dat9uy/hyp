# Method: a skill-text gate

This method is pre-registered for an Experiment whose instrument does not
exist yet. The instrument decides whether a change to a skill's text
(`skills/*/SKILL.md`) ships: the change ships only when a fresh agent
runner, started from a recorded point of a hyp tree, shows a failure in 3
or more of 5 trials with the skills as they are, and in 1 or fewer of 5
with the changed text.

## What one trial is

- One hyp sandbox per trial. The sandbox is made from this repository at a
  pinned commit, with the tree's store copied in at a store cut
  (`hyp sandbox create ... --store SRC --cut EVENT`). The store and the
  cut are given at run time; the instrument takes them as they are.
- An arm is the set of skill files the runner loads: `base` is the two hyp
  skills as the sandbox's commit has them; any other arm is a directory of
  `SKILL.md` files laid over them, one per skill it changes. The variant
  directories are part of the instrument. The runner's hyp is the same in
  every arm.
- Before the runner starts, the sandbox holds only hyp's code, the two
  skills (with the arm's variant over them) and the tree at the cut: no
  agent files (`AGENTS.md`, `CLAUDE.md`, `.claude` beyond the skills), no
  plans.
- The runner is a fresh `claude -p` session with the model `sonnet`, given
  a brief on stdin and project settings only. It is stopped after 50
  minutes. Its session transcript is kept with the trial's evidence, as is
  a summary of its run (exit, turns, cost).
- After the runner ends, the sandbox is removed and its changes kept, so
  the trial's evidence holds the transcript and every file the runner
  added or changed, its tree's events included.

## Scoring one trial

A score command reads one trial's evidence and prints `key<TAB>value`
lines:

- `arm`, `trial`;
- `skill_loaded`: 1 when the runner loaded a skill (a Skill tool call, or a
  Read of a `SKILL.md`);
- `failing_hyp`: the number of Bash calls that ended in error with `hyp` as
  the command whose exit status the call returns;
- one 0/1 fingerprint per row the gate is asked about:
  - `f_reading_pin`: an errored tool result holds hyp's refusal of an
    Experiment that declares `--decided-by reading` while it pins a recipe;
  - `f_body_limit`: an errored tool result holds hyp's refusal of a body
    over 600 runes.
- `end_state`: `recorded` when the runner's events record a Result,
  `registered` when they add an Experiment but no Result, else `none`.

## The gate

A verdict command reads the score of every trial, with the rows to judge
given as arguments, and needs exactly 5 trials per arm with distinct trial
numbers. For each row and arm, k is the number of trials whose fingerprint
is 1. A trial that never loaded a skill counts against its arm: for
`base` it is never a reproduction, for any other arm it always counts as a
failure.

- A row reproduces when `base` has k >= 3.
- A variant arm passes a row at k <= 1.
- It prints `verdict: repro=supports` when any row given reproduces, else
  `verdict: repro=contradicts`; and for each variant arm
  `verdict: ARM=supports` when it passes every reproducing row,
  `contradicts` when it fails one, and `inconclusive` when no row
  reproduces.

Both commands need known answers that cover every field and every verdict
branch, written by hand from the rules above.
