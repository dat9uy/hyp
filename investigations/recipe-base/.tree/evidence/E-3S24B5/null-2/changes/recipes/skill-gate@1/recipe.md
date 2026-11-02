---
name: skill-gate
version: 1
format: 1
subject: A fresh sonnet runner with hyp and its two skills (base, or a variant directory of SKILL.md files laid over them) is started from a hyp tree at a store cut; its failures are scored from its transcript and changes, and the skill-text gate (3 of 5 on base, 1 of 5 or fewer on a variant) is applied
---

# skill-gate

The instrument of METHOD.md's skill-text gate: it decides whether a change
to a skill's text (`skills/*/SKILL.md`) ships. One sandbox per trial, five
trials per arm. `score` reads one trial's evidence; `verdict` reads every
trial's score and applies the gate. The tree, the store cut, the brief and
the variant directories are inputs of a run, not files of this bundle.

## Launch

Run: ./launch.sh

Create the sandbox with `hyp sandbox create ID --host H --base REV
--recipe skill-gate@1 --store SRC --cut EVENT --inputs FILE`. `REV` pins
the commit the sandbox is made from, and the cut puts the tree at `EVENT`
in `wt/.tree`.

The inputs file is `KEY=VALUE` lines, all required except `variant` for
`base`: `arm` (`base`, or a label `[a-z][a-z0-9-]{0,31}` other than
`repro`), `trial` (a whole number), `brief` (an absolute path to the
runner's brief, a file), `cut` (the event id the store was cut at) and, for
an arm other than `base`, `variant` (an absolute path to a directory
holding `<skill>/SKILL.md` for each skill the arm changes; `<skill>` is
`hyp-prototype` or `hyp-verification`). Launch exits 2 when a line is
missing or wrong, or `wt/.tree` is missing.

Launch deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`, `plans`,
`investigations`, `recipes` and `METHOD.md` from `wt/`, leaving hyp's code,
the two skills and the tree. It copies `skills/hyp-prototype` and
`skills/hyp-verification` into `wt/.claude/skills/`, and for an arm other
than `base` copies each variant `SKILL.md` over both that copy and
`wt/skills/<skill>/SKILL.md`, so the runner cannot read the base text of a
skill the arm changes. The brief goes to this recipe's tmp, never to `wt/`.
These changes come before the manifest, so they are not changes of the
trial. Every product of the trial is written in Drive.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and `hyp` run; the deleted paths are gone;
`.claude/` holds only `skills/`, which holds exactly the two hyp skills;
each skill's text is the base text, or the variant's for the skills the arm
changes, in both places; `hyp events` in `wt/` ends at the inputs file's
`cut`; and the brief is a non-empty file.

## Drive

Run: ./drive.py

Runs `claude -p --model sonnet --setting-sources project` in `wt/` with the
brief on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (exit, turns, cost). It exits 1 when no transcript was
written.

## Evidence

Kept at remove: `--keep tmp/skill-gate/state.env --keep
tmp/skill-gate/drive.json --keep tmp/skill-gate/transcript.jsonl`. The
runner's files, its tree's events included, reach OUT_DIR through
`changes/`. Then `score` runs on that OUT_DIR.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of one trial (hyp sandbox remove's, with the three kept files); reads only that directory

Run verdict: ./verdict.py
External: the score outputs of every trial, as absolute paths; reads only those files

- `score OUT_DIR` prints `key<TAB>value` lines: `arm`, `trial`,
  `skill_loaded`, `failing_hyp`, `f_reading_pin`, `f_body_limit` and
  `end_state` (`recorded`, `registered` or `none`). The script's docstring
  defines each. It exits 2 when OUT_DIR lacks `state.env`,
  `transcript.jsonl` or `changes.json`.
- `verdict --row ROW [--row ROW]... SCORE...` where ROW is a fingerprint
  name (`f_reading_pin`, `f_body_limit`) and each SCORE is one trial's
  score output. It needs exactly 5 trials per arm with distinct trial
  numbers, and a `base` arm. A row reproduces when `base` has k >= 3, where
  a trial that loaded no skill never counts for `base`; a variant arm
  passes a row at k <= 1, where a trial that loaded no skill always counts
  as a failure. It prints `verdict: repro=supports|contradicts` and
  `verdict: ARM=supports|contradicts|inconclusive` for each variant arm
  (`inconclusive` when no row reproduces). It exits 2 on bad input.

`known-answers/` holds, for `score`, a trial that hits both fingerprints,
a clean trial that loaded no skill and recorded a Result, a registered-only
trial, a trial with no transcript and a usage failure; and, for `verdict`,
a reproducing base with a passing and a failing variant, a base that does
not reproduce (`inconclusive`), a base whose third failure loaded no skill,
a variant whose unloaded trial fails it, a row that reproduces beside one
that does not, and bad inputs. Every expected output was written by hand
from METHOD.md.
