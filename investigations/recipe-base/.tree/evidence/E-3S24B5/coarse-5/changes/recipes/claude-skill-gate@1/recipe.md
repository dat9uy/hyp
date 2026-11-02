---
name: claude-skill-gate
version: 1
from: base:claude-runner-gate@1 656b478b4a4716b3498b10d3c13e5572bfea2a41359bba3c8bfe61f07e649c63
format: 1
subject: a fresh claude -p runner (sonnet) started from a store cut of a hyp tree, with the hyp skills as they are or with a variant laid over them
---

# claude-skill-gate

A whole runner gate in one bundle. One sandbox per trial: Launch gives a
fresh Claude Code runner a brief and a method, Drive runs it and keeps its
transcript, `score` counts the trial's stops, and `verdict` applies a gate
to every trial's score in one batched call. It drives `claude -p` only; a
runner for another harness is another base. Create each sandbox with
`--inputs FILE`, and with `--store SRC --cut EVENT` when the runner starts
from a tree at a store cut; `--hyp BIN` pins the hyp the runner calls.

## Launch

Run: ./launch.sh

Reads the inputs file, `KEY=VALUE` lines `arm`, `trial`, `model` and, for an
arm other than `base`, `variant` (the absolute path of a directory of
`<skill>/SKILL.md` files), and exits 2 when one is missing or the variant
holds anything else or names a skill that is not there. It copies the inputs to this recipe's tmp as
`state.env`, deletes the paths in `STRIP` from `wt/`, copies the two hyp
skills into `wt/.claude/skills/`, lays the variant's files over them, deletes
`wt/skills/` so the base text cannot be read around the variant, and writes `brief.txt` followed by
`method.md` to this recipe's tmp as the runner's brief. 

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
`changes/` and `changes.json`. Then `score` runs on that OUT_DIR, as a
Standalone command, once per trial, and `verdict` once over all of them.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: none

Run verdict: ./verdict.py
External: none

- `score OUT_DIR` prints `key<TAB>value` lines: `arm`, `trial`,
  `skill_loaded`, `failing_hyp`, `f_reading_pin`, `f_body_limit` and
  `end_state` (`recorded`, `registered` or `none`, read from the events
  files the runner added or changed). The script's docstring defines each
  field. It exits 2 when OUT_DIR lacks `state.env`, `drive.json`,
  `transcript.jsonl` or `changes.json`.
- `verdict --base ARM --trials N --repro R --pass P --row F... SCORE...`
  reads one score file per trial; the method's gate is
  `--base base --trials 5 --repro 3 --pass 1` with a `--row f_...` for each
  row asked about. An unloaded trial is never a reproduction in base and
  always a failure in a variant arm. It prints one line per row and arm,
  `verdict: repro=…`, then `verdict: ARM=…` for every other arm. The
  script's docstring states the rule. It exits 2 on a usage error, a
  missing field, no base arm, or an arm without exactly N distinct trials.

`known-answers/` holds, for `score`, a clean trial (the refusal text in a
non-error result does not count), a pin refusal with an Experiment but no
Result, a body-limit refusal with no skill loaded and a Result, changed paths
that are not the runner's events, `score-seam` (a `drive.json` as `drive.py`
wrote it on a stub `claude`), a missing file and a usage failure; and, for
`verdict`, passing and failing variants, a gate with no reproducing row, a
second row that reproduces alone, unloaded trials in base and in a variant,
and each usage failure. Every expected output was written by hand from the
rules in `METHOD.md`.
