---
name: skill-gate
version: 1
from: base:claude-runner@1 653da60f3ee4b3abdbed0924638749e22f538ac084678df8de445c77109d76cc
from: base:transcript-score@1 5719d450e84411b4c045ce64af158bff8adf69dd6f62a3ffc79cc0c4690f904a
from: base:gate-verdict@1 99ddae8b7f8e8647a234c0705298d7b64d9e894e9f7183eedf87dd7cdea59805
format: 1
subject: a fresh Claude Code runner (model sonnet) started from a recorded point of a hyp tree, with the hyp skills as they are or with variant SKILL.md files laid over them
---

# skill-gate

The instrument of METHOD.md, "a skill-text gate": a change to a skill's
text ships only when the runner fails in 3 or more of 5 trials with the
skills as they are (`base`) and in 1 or fewer of 5 with the changed text.
One sandbox is one trial. Create it from this repository at the pinned
commit with the tree's store copied in:

    hyp sandbox create ID --host H --base REV --recipe skill-gate@1 \
      --store SRC --cut EVENT --inputs FILE

`--store` and `--cut` are required (Launch refuses a wt with no tree).
`--hyp BIN` pins the hyp the runner calls; use the same one in every arm.
The inputs file has `KEY=VALUE` lines `arm` (`base`, or a variant arm's
name), `trial` (a number) and, for a variant arm only, `variant=DIR`, an
absolute path to a directory of `<skill>/SKILL.md` files, one for each of
`hyp-prototype` and `hyp-verification` the arm changes and nothing else.

## Launch

Run: ./launch.sh

Reads the inputs file and exits 2 on a missing `arm` or `trial`, a trial
that is not a number, `base` with a `variant`, a variant arm without one,
an arm name that is not alias-shaped or is `repro`, and a variant that is
not a directory of `<skill>/SKILL.md` files for the two hyp skills. It
writes `state.env` (arm, trial, `model=sonnet`, and the sha256 of each
variant file) to this recipe's tmp, deletes `AGENTS.md`, `CLAUDE.md`,
`.claude`, `.agentkit`, `plans`, `METHOD.md` and any `recipes/skill-gate@*`
from `wt/`, copies the two skills to `wt/.claude/skills/`, lays the
arm's variant files over the skills in both `wt/.claude/skills/` and
`wt/skills/`, and writes `brief.txt` as the runner's brief. The tree at the
cut stays in `wt/.tree`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the stripped
paths are gone, `.claude/` holds only `skills/`, both skills are there and
equal to `wt/skills/`, `wt/.tree/events` exists, `state.env` names the arm,
trial and model `sonnet`, and the brief exists.

## Drive

Run: ./drive.py

Runs `claude -p --model sonnet --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
the brief on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (arm, trial, model, exit, turns, cost, denials). It
exits 1 when no transcript was written.

## Evidence

Kept by hand at remove, so that the trial's evidence holds the transcript,
the summary and every file the runner added or changed (`changes/` and
`changes.json`, its tree's events included):

    hyp sandbox remove ID OUT_DIR --keep tmp/skill-gate/state.env \
      --keep tmp/skill-gate/drive.json --keep tmp/skill-gate/transcript.jsonl

Then score each OUT_DIR and give the scores to the verdict (Standalone).

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial's evidence); reads files only

- `score OUT_DIR` prints `key<TAB>value` lines `arm`, `trial`,
  `skill_loaded`, `failing_hyp`, `f_reading_pin`, `f_body_limit` and
  `end_state`, as the script's docstring defines them. It exits 2 when
  OUT_DIR lacks `state.env`, `transcript.jsonl` or `changes.json`, or
  `state.env` lacks `arm` or `trial`.

Run verdict: ./verdict.py
External: score files, one per trial (the output of score); reads files only

- `verdict ROW... -- SCORE...` reads the score of every trial and judges
  the rows named (`f_reading_pin`, `f_body_limit`). It needs exactly 5
  trials with distinct numbers in each arm and a `base` arm. A row
  reproduces when `base` fails it in 3 or more trials; a variant arm passes
  a row at 1 or fewer. A trial that loaded no skill never counts as a
  reproduction for `base` and always counts as a failure for a variant. It
  prints `verdict: repro=…` and `verdict: ARM=…` for each variant arm
  (`supports`, `contradicts`, or `inconclusive` when no row reproduces).
  It exits 2 on a usage error or bad scores.

`known-answers/` holds, for score, a clean trial, each fingerprint, each
end state (`none`, `registered`, `recorded`), an events file that is
modified or not copied, and the exit-2 failures; for verdict, a passing
and a failing variant, the thresholds' boundaries, trials that loaded no
skill in either arm, two rows of which one reproduces, no reproduction,
and the exit-2 failures. Every expected output was written by hand from
the docstrings of the two scripts.
