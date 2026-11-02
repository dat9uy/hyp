---
name: claude-skill-gate
version: 1
from: base:claude-runner-gate@1 656b478b4a4716b3498b10d3c13e5572bfea2a41359bba3c8bfe61f07e649c63
format: 1
subject: a skill-text change (a variant of skills/*/SKILL.md) against the skills as they are, judged by a fresh claude -p runner started from a tree at a store cut
---

# claude-skill-gate

The instrument of the pre-registered skill-text gate (METHOD.md): a change
to a skill's text ships only when `base` shows a failure in 3 or more of 5
trials and the changed text in 1 or fewer of 5. One sandbox per trial,
made with `hyp sandbox create ID --host H --base COMMIT --recipe
claude-skill-gate@1 --store SRC --cut EVENT --inputs FILE`; the same
`--hyp BIN` in every arm. The store, the cut and the variant directories
are run-time inputs, not part of the bundle.

Inputs file, `KEY=VALUE` lines: `arm` (`base`, or an arm name matching
`[a-z][a-z0-9-]*`), `trial` (1 to 5, distinct within an arm), `model`
(`sonnet`) and, for an arm other than `base`, `variant=DIR`: an absolute
path to a directory of `<skill>/SKILL.md` files, one per hyp skill it
changes.

## Launch

Run: ./launch.sh

Checks the inputs (exit 2 on a missing line, or on a variant that is not a
directory of `hyp-prototype/` or `hyp-verification/` entries each holding a
`SKILL.md`), copies them to this recipe's tmp as `state.env`, deletes
`AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit` and `plans` from `wt/`,
copies the two hyp skills into `wt/.claude/skills/`, lays the variant's
`SKILL.md` files over them (and over `wt/skills/`, so no file shows the
old text), and copies `brief.txt` to this recipe's tmp as the runner's
brief. The tree at the cut is the one `sandbox create` copied in; Launch
leaves it alone.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and `hyp` run, the stripped paths are gone,
`.claude/` holds only `skills/` with both skills, and Launch wrote the
brief and `state.env`.

## Drive

Run: ./drive.py

Runs `claude -p --model MODEL --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
the brief on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (arm, trial, model, exit, num_turns, cost_usd,
permission_denials). It exits 1 when no transcript was written.

## Evidence

Kept by hand at remove: `--keep tmp/claude-skill-gate/state.env --keep
tmp/claude-skill-gate/drive.json --keep tmp/claude-skill-gate/transcript.jsonl`.
The sandbox is removed and what the runner added or changed (its tree's
events included) reaches OUT_DIR through `changes/` and `changes.json`.
Then `score` runs on that OUT_DIR once per trial, and `verdict` once over
all of them.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: none

Run verdict: ./verdict.py
External: none

- `score OUT_DIR` prints `arm`, `trial`, `skill_loaded`, `failing_hyp`,
  `f_reading_pin`, `f_body_limit` and `end_state` as `key<TAB>value`
  lines. The script's docstring defines each field. It exits 2 when OUT_DIR
  lacks `state.env`, `drive.json`, `transcript.jsonl` or `changes.json`.
- `verdict --row F [--row F]... SCORE...` reads one score file per trial
  and needs exactly 5 trials per arm with distinct trial numbers. Per row
  and arm, k counts the trials whose F is 1; a trial that never loaded a
  skill never counts in `base` and always counts in another arm. A row
  reproduces at base k >= 3; a variant arm passes a row at k <= 1. It
  prints `verdict: repro=supports|contradicts`, then `verdict: ARM=supports`
  (passes every reproducing row), `contradicts` (fails one) or
  `inconclusive` (no row reproduces). Exits 2 on a usage error, a missing
  key, a value that is not 0 or 1, no base arm, or a wrong trial count.

`known-answers/` holds, for `score`, a clean trial, a trial with no skill,
a trial with two failing hyp calls and both fingerprints, each fingerprint
alone, end_state `registered` and `recorded` in added and modified event
files, `score-seam` (the `drive.json` shape `drive.py` writes), each
missing file and a usage failure; and, for `verdict`, every threshold edge
and branch (supports, contradicts, inconclusive), the skill-loaded rule in
both directions, several rows and several arms, and each exit-2 error.
Every expected output was written by hand from the rules above.
