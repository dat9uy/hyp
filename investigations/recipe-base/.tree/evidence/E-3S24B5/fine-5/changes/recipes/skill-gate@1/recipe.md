---
name: skill-gate
version: 1
from: base:claude-runner@1 653da60f3ee4b3abdbed0924638749e22f538ac084678df8de445c77109d76cc
from: base:transcript-score@1 5719d450e84411b4c045ce64af158bff8adf69dd6f62a3ffc79cc0c4690f904a
from: base:gate-verdict@1 99ddae8b7f8e8647a234c0705298d7b64d9e894e9f7183eedf87dd7cdea59805
format: 1
subject: one trial of a fresh Claude Code runner (model sonnet) on a hyp sandbox at a pinned commit and store cut, with the hyp skills as one arm has them, and the 5-trial gate over the trials' scores
---

# skill-gate

The instrument of the skill-text gate (METHOD.md): a change to
`skills/*/SKILL.md` ships only when `base` shows a failure in 3 or more of
5 trials and the changed arm in 1 or fewer. One trial is one sandbox:

    hyp sandbox create ID --host H --base REV --recipe skill-gate@1 \
        --store SRC --cut EVENT --inputs FILE

The store, the cut and the base commit are the run's. `FILE` holds
`KEY=VALUE` lines `arm`, `trial`, `model=sonnet` and, for any arm but
`base`, `variant=DIR`: a directory with `<skill>/SKILL.md` for each of
`hyp-prototype` and `hyp-verification` it changes. `--hyp BIN` pins the
hyp the runner calls, the same in every arm.

## Launch

Run: ./launch.sh

Reads the inputs file and exits 2 when `arm`, `trial` or `model` is
missing, `model` is not `sonnet`, `base` has a `variant=` or another arm has
none, or the variant holds anything but `hyp-prototype/SKILL.md` and
`hyp-verification/SKILL.md`. It copies the inputs to this recipe's tmp as
`state.env`, deletes `AGENTS.md CLAUDE.md .claude .agentkit plans` from
`wt/`, copies the two skills of the commit into `wt/.claude/skills/`, lays
the arm's variant files over them (and over `wt/skills/`), and writes
`brief.txt` to this recipe's tmp as the runner's brief. `wt/.tree` is the
store cut, written by hyp before Launch.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the stripped
paths are gone, `.claude/` holds only `skills/` and the brief exists.

## Drive

Run: ./drive.py

Runs `claude -p --model MODEL --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
the brief on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (arm, trial, model, exit, turns, cost_usd, denials).
It exits 1 when no transcript was written.

## Evidence

Kept by hand at remove: `--keep tmp/<name>/state.env --keep
tmp/<name>/drive.json --keep tmp/<name>/transcript.jsonl`, with `<name>`
this bundle's name. What the runner changed reaches OUT_DIR through
`changes/` and `changes.json`.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: none

- `score OUT_DIR` scores one trial from its OUT_DIR, as `hyp sandbox remove`
  left it with `state.env` and `transcript.jsonl` kept. It prints
  `key<TAB>value` lines `arm`, `trial`, `skill_loaded`, `failing_hyp`,
  `f_reading_pin`, `f_body_limit` and `end_state`; the script's docstring
  defines each. It exits 2 when OUT_DIR lacks `state.env`,
  `transcript.jsonl` or `changes.json`.

Run verdict: ./verdict.py
External: none

- `verdict --row F [--row F]... SCORE...` reads the score of every trial
  (10 or more files, exactly 5 per arm) and prints, per arm and row, `k`,
  then `verdict: repro=…` and `verdict: ARM=…` as the docstring and METHOD.md
  state: base reproduces a row at k >= 3, a variant passes it at k <= 1, and
  a trial that loaded no skill never counts as base's reproduction and
  always counts as a variant's failure. It exits 2 on a malformed input.

`known-answers/` holds, for both commands, every field and every verdict
branch, written by hand from the docstrings and METHOD.md.
