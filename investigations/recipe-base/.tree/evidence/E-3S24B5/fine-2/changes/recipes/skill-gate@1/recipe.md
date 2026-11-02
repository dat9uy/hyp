---
name: skill-gate
version: 1
from: base:claude-runner@1 653da60f3ee4b3abdbed0924638749e22f538ac084678df8de445c77109d76cc
from: base:transcript-score@1 5719d450e84411b4c045ce64af158bff8adf69dd6f62a3ffc79cc0c4690f904a
from: base:gate-verdict@1 99ddae8b7f8e8647a234c0705298d7b64d9e894e9f7183eedf87dd7cdea59805
format: 1
subject: one trial of a fresh Claude Code runner on a hyp tree at a store cut, with or without a variant of the hyp skills' text
---

# skill-gate

The instrument of the skill-text gate (METHOD.md). A change to a skill's
text (`skills/*/SKILL.md`) ships only when the runner shows a failure in 3
or more of 5 trials with the skills as they are, and in 1 or fewer of 5
with the changed text.

One trial is one sandbox, made from this repository at a pinned commit
with a tree's store copied in at a store cut:

```
hyp sandbox create ID --host H --base COMMIT --recipe skill-gate@1 \
  --store SRC --cut EVENT --inputs FILE --hyp BIN
```

The store, the cut and the variant directories are given at run time.
`--hyp BIN` is the same binary in every arm. The inputs file holds
`KEY=VALUE` lines: `arm` (`base` or a label `[a-z][a-z0-9-]*`), `trial`
(1 to 5, distinct within an arm), `model` (`sonnet`) and, for an arm other
than `base`, `variant=DIR`, an absolute path to a directory holding
`<skill>/SKILL.md` for each skill that arm changes (`hyp-prototype` or
`hyp-verification`). A variant may be an example.

## Launch

Run: ./launch.sh

Deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit` and `plans` from
`wt/`, copies the two hyp skills into `wt/.claude/skills/`, and for an arm
other than `base` lays each variant `SKILL.md` over its skill in both
`wt/.claude/skills/` and `wt/skills/`. It records each laid file's sha256 in
`variant.sha256`, copies the inputs to `state.env`, and copies `brief.txt`
to this recipe's tmp as the runner's brief. Exits 2 when the inputs lack
`arm`, `trial` or `model`, or a variant is missing, empty, or holds anything
but a skill with a `SKILL.md`. The tree at the cut is already in `wt/.tree`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the stripped paths
are gone, `.claude/` holds only `skills/` with both skills, the laid
variant files match `variant.sha256`, and the brief exists.

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

Keep at remove: `--keep tmp/skill-gate/state.env --keep
tmp/skill-gate/drive.json --keep tmp/skill-gate/transcript.jsonl --keep
tmp/skill-gate/variant.sha256`. Remove adds `changes.json` and `changes/`:
every file the runner added or changed, its tree's events included.

Run: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial's kept files); writes only $HYP_OUT, nothing

Prints the trial's `key<TAB>value` lines: `arm`, `trial`, `skill_loaded`,
`failing_hyp`, `f_reading_pin`, `f_body_limit`, `end_state`. The script's
docstring defines each. It reads files only and calls no hyp. It exits 2
when OUT_DIR lacks `state.env`, `transcript.jsonl` or `changes.json`.
`end_state` reads only events files remove lists as added (the runner's own
actor file).

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run verdict: ./verdict.py
External: the score files given as arguments (absolute paths); writes nothing

`hyp recipe run --record DIR skill-gate@1 verdict --row f_reading_pin --row
f_body_limit SCORE...` reads the printed score of every trial, with the rows
to judge as `--row` arguments. The rule is fixed in the script: exactly 5
trials per arm with distinct trial numbers, a row reproduces when `base` has
k >= 3 trials with it at 1, a variant arm passes a row at k <= 1. A trial
that never loaded a skill is never a reproduction in `base` and always a
failure in any other arm. It prints `verdict: repro=supports|contradicts`
and `verdict: ARM=supports|contradicts|inconclusive` for each variant arm.
Exits 2 on a usage error or a score set that breaks the rule above.

## Known answers

`known-answers/` holds, for `./score.py`, a trial that records a Result
with both fingerprints and three failing hyp calls (a heredoc, a pipe, a
plain error and a non-hyp error among them), a trial that only registers an
Experiment (successful results holding refusal text, a modified events
file, ignored), a trial with no skill load and no events, and a trial with
no transcript. For `verdict`, a reproducing base with a passing and a failing
arm, a base that does not reproduce, unloaded trials in both kinds of arm at
the thresholds, two rows judged together, and six refusals. Every expected
output was written by hand from the docstrings.
