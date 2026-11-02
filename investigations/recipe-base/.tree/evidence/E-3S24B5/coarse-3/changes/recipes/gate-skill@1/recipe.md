---
name: gate-skill
version: 1
from: base:claude-runner-gate@1 656b478b4a4716b3498b10d3c13e5572bfea2a41359bba3c8bfe61f07e649c63
format: 1
subject: a skill-text gate: fresh claude runners per arm, scored for two refusal fingerprints, and a 3-of-5 reproduce / 1-of-5 pass verdict
---

# gate-skill

A whole runner gate in one bundle. One sandbox per trial: Launch gives a
fresh Claude Code runner a brief and a method, Drive runs it and keeps its
transcript, `score` counts the trial's stops, and `verdict` applies a gate
to every trial's score in one batched call. It drives `claude -p` only; a
runner for another harness is another base. Create each sandbox with
`--inputs FILE`, and with `--store SRC --cut EVENT` when the runner starts
from a tree at a store cut; `--hyp BIN` pins the hyp the runner calls.

## Launch

Run: ./launch.sh

Reads the inputs file, `KEY=VALUE` lines `arm` and `trial` (and `model`,
which must be `sonnet` when given), and exits 2 when one is missing. The arm
`base` leaves the skills as the sandbox's commit has them; any other arm
needs `variants/ARM/SKILL-NAME/SKILL.md` in this bundle, laid over
`.claude/skills/` and `skills/`, else exit 2. It copies the inputs to this
recipe's tmp as `state.env` (model fixed to sonnet), deletes `AGENTS.md`,
`CLAUDE.md`, `.claude`, `.agentkit` and `plans` from `wt/`, copies the two
hyp skills into `wt/.claude/skills/`, and writes `brief.txt` followed by
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
  `end_state` (`recorded`, `registered` or `none`). The script's docstring
  defines each field. It exits 2 when OUT_DIR lacks `state.env`,
  `drive.json`, `transcript.jsonl` or `changes.json`.
- `verdict --row F [--row F]... SCORE...` reads one score file per trial,
  5 per arm with distinct trial numbers and a `base` arm. A row
  reproduces at k >= 3 on `base`; a variant arm passes it at k <= 1; a
  trial that never loaded a skill is never a base reproduction and always a
  variant failure. It prints `verdict: repro=…` and `verdict: ARM=…`
  (`supports`, `contradicts`, `inconclusive`). It exits 2 on a usage error,
  a missing field, no base arm, or an arm without exactly 5 trials.

`known-answers/` holds, for `score`, a clean trial, both fingerprints with
failing hyp calls, a Result, an Experiment without one, modified event files
with and without the runner's own events, a missing transcript, a usage
failure, and `score-seam`, whose `drive.json` is what `drive.py` wrote on a
stub `claude`; and, for `verdict`, every branch (reproducing and not,
passing and failing arms, skill-less trials in each arm, several rows, a
single row, and each exit-2 error). Every expected output was written by
hand from the docstrings.
