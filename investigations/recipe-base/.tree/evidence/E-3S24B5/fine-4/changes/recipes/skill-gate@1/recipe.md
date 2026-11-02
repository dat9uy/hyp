---
name: skill-gate
version: 1
from: base:claude-runner@1 653da60f3ee4b3abdbed0924638749e22f538ac084678df8de445c77109d76cc
from: base:transcript-score@1 5719d450e84411b4c045ce64af158bff8adf69dd6f62a3ffc79cc0c4690f904a
from: base:gate-verdict@1 99ddae8b7f8e8647a234c0705298d7b64d9e894e9f7183eedf87dd7cdea59805
format: 1
subject: a fresh Claude Code runner (model sonnet), started from a tree at a store cut with the hyp skills as one arm has them
---

# skill-gate

The instrument of the skill-text gate (METHOD.md): a change to
`skills/*/SKILL.md` ships only when `base` shows a failure in 3 or more of 5
trials and the changed text in 1 or fewer of 5. One sandbox is one trial.

Create each trial's sandbox with
`hyp sandbox create ID --host H --base COMMIT --recipe skill-gate@1 --store SRC --cut EVENT --inputs FILE --hyp BIN`.
The store and the cut are the run's own; the recipe takes them as they are.
The inputs file has `KEY=VALUE` lines:

- `arm`: `base`, or a label `[a-z][a-z0-9-]{0,31}` other than `repro`;
- `trial`: a number; the 5 trials of an arm have distinct numbers;
- `model`: `sonnet`;
- `variant`: for an arm other than `base` only, the absolute path of a
  directory of `<skill>/SKILL.md` files, one per skill it changes
  (`hyp-prototype`, `hyp-verification`). It is laid over the skills in both
  `wt/.claude/skills/` and `wt/skills/`. `base` takes none.

## Launch

Run: ./launch.sh

Checks the inputs. Deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`,
`investigations` and `plans` from `wt/`, copies the two skills from `wt/skills/` into
`wt/.claude/skills/`, lays the variant over them, and keeps the brief
(`brief.txt`), `state.env` and `variant.sha256` (the hash of each variant
file, or `none`) in this recipe's tmp. The runner's hyp is the sandbox's
pinned one, the same in every arm.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the agent files
are gone, `.claude/` holds only `skills/`, both skill copies agree, the brief
exists and `hyp status` reads the tree at the cut in `wt/`.

## Drive

Run: ./drive.py

Runs a fresh `claude -p --model sonnet --setting-sources project
--permission-mode dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill`
in `wt/` with the brief on stdin and `HYP_ACTOR=runner-<arm>-<trial>`, and
stops it after 50 minutes. It copies the session transcript to this recipe's
tmp as `transcript.jsonl` and writes `drive.json` (exit, turns, cost).

## Evidence

Kept at remove: `--keep tmp/skill-gate/state.env --keep
tmp/skill-gate/drive.json --keep tmp/skill-gate/transcript.jsonl --keep
tmp/skill-gate/variant.sha256`. What the runner added or changed, its tree's
events included, reaches OUT_DIR through `changes/` and `changes.json`.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of one trial's hyp sandbox remove (argument); reads only

`score OUT_DIR` prints `arm`, `trial`, `skill_loaded`, `failing_hyp`,
`f_reading_pin`, `f_body_limit` and `end_state`; the script's docstring
defines each. It exits 2 when OUT_DIR lacks `state.env`,
`transcript.jsonl` or `sandbox.json`.

Run verdict: ./verdict.py
External: the score files given as arguments; reads only

`verdict --row ROW [--row ROW ...] SCORE...` reads the score of every
trial, exactly 5 per arm and one arm `base`, and judges the given rows
(`f_reading_pin`, `f_body_limit`). It prints `verdict: repro=supports|contradicts`
and `verdict: ARM=supports|contradicts|inconclusive` per variant arm, by the
rule in its docstring (base reproduces a row at k >= 3; a variant passes at
k <= 1; a trial that never loaded a skill is no reproduction in `base` and a
failure in any other arm). It exits 2 on a malformed input.

`known-answers/` holds hand-written trials and score sets that cover every
score field and every verdict branch.
