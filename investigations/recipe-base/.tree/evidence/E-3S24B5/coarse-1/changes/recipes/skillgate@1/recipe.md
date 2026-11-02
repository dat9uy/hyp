---
name: skillgate
version: 1
from: base:claude-runner-gate@1 656b478b4a4716b3498b10d3c13e5572bfea2a41359bba3c8bfe61f07e649c63
format: 1
subject: the skill-text gate: a change to skills/*/SKILL.md ships only when base shows a failure in 3 of 5 trials and the changed text in 1 or fewer
---

# skillgate

A whole runner gate in one bundle. One sandbox per trial: Launch gives a
fresh Claude Code runner a brief and a method, Drive runs it and keeps its
transcript, `score` counts the trial's stops, and `verdict` applies a gate
to every trial's score in one batched call. It drives `claude -p` only; a
runner for another harness is another base. Create each sandbox with
`--inputs FILE`, and with `--store SRC --cut EVENT` when the runner starts
from a tree at a store cut; `--hyp BIN` pins the hyp the runner calls.

## Launch

Run: ./launch.sh

Reads the inputs file, `KEY=VALUE` lines `arm`, `trial` and `model`
(required), `variant` (a directory of `<skill>/SKILL.md` files, required
when arm is not `base`) and `brief` (a file replacing `brief.txt`). Exits 2
when one is missing or the variant holds no SKILL.md of a hyp skill. It
copies the inputs to this recipe's tmp as `state.env`, deletes the paths in
`STRIP` from `wt/`, copies the two hyp skills into `.claude/skills/`, lays
the variant's SKILL.md files over them (and over `skills/`), and writes the
brief followed by `method.md` to this recipe's tmp as the runner's stdin.

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

- `score OUT_DIR` prints `arm`, `trial`, `skill_loaded`, `failing_hyp`,
  `f_reading_pin`, `f_body_limit` and `end_state` (`recorded`,
  `registered` or `none`, from the runner's own `.tree/events` lines in
  `changes/`). The script's docstring defines each field. It exits 2 when
  OUT_DIR lacks `state.env`, `drive.json`, `transcript.jsonl` or
  `changes.json`.
- `verdict --base ARM --trials N --repro R --pass P --row F... SCORE...`
  judges each row on every arm: base reproduces a row at k >= R, a variant
  passes at k <= P, a trial that never loaded a skill is never a base
  failure and always a variant failure. Run it with `--base base --trials 5
  --repro 3 --pass 1 --row f_reading_pin --row f_body_limit`. It exits 2 on
  a usage error, a missing field, no base arm, or an arm without exactly N
  distinct trials.

`known-answers/` holds hand-written cases for every `score` field (hits,
no hits, a non-error result holding the text, events of other actors,
uncopied files, a missing transcript, usage, and a `drive.json` as
`drive.py` writes it) and for every `verdict` branch (reproduce and pass or
fail, no reproduction, unloaded base and variant trials, several variants,
and each input error).
