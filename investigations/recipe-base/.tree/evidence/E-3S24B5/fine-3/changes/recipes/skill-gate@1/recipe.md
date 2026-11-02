---
name: skill-gate
version: 1
from: base:claude-runner@1 653da60f3ee4b3abdbed0924638749e22f538ac084678df8de445c77109d76cc
from: base:transcript-score@1 5719d450e84411b4c045ce64af158bff8adf69dd6f62a3ffc79cc0c4690f904a
from: base:gate-verdict@1 99ddae8b7f8e8647a234c0705298d7b64d9e894e9f7183eedf87dd7cdea59805
format: 1
subject: a fresh claude -p runner started from a store cut of a hyp tree, with the hyp skills as the arm names them
---

# skill-gate

A sandbox's Launch, Doctor, Drive and Cleanup for one trial of a fresh
Claude Code runner. It drives `claude -p` only; a runner for another
harness is another base. Create the sandbox with `--inputs FILE`, and
with `--store SRC --cut EVENT` when the runner starts from a tree at a
store cut; `--hyp BIN` pins the hyp the runner calls.

A Standalone check that scores one trial OUT_DIR, as `hyp sandbox remove`
left it with the runner's `state.env` and `transcript.jsonl` kept. It reads
the Claude Code transcript format.

A Standalone check that applies a reproduce-then-pass gate to the score
files of an Experiment's trials, one batched call after every trial ran.
Its thresholds are command-line arguments, so the Experiment's body
pre-registers the rule by quoting the command line.

## Launch

Run: ./launch.sh

Reads the inputs file, `KEY=VALUE` lines `arm`, `trial` and `model`, and
`variant=DIR` (absolute) for every arm but `base`; exits 2 when one is
missing, or when DIR is not a directory of `<skill>/SKILL.md` files naming
`hyp-prototype` or `hyp-verification`. It copies the inputs to this
recipe's tmp as `state.env`, deletes `AGENTS.md CLAUDE.md .claude
.agentkit plans` from `wt/`, copies the two hyp skills as the commit has
them into `wt/.claude/skills/`, lays the variant's `SKILL.md` files over
them (and over `wt/skills/`, so the runner reads one text), and writes
`brief.txt` followed by `method.md` to this recipe's tmp as the runner's
brief. The `base` arm lays nothing over. The sandbox is created with
`--store SRC --cut EVENT` for the tree's store cut and `--hyp BIN`, the
same hyp in every arm.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the stripped
paths are gone, `.claude/` holds only `skills/`, `skills/<s>/SKILL.md` and
`.claude/skills/<s>/SKILL.md` are the same text, and the brief exists.

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

- `score OUT_DIR` prints `key<TAB>value` lines `arm`, `trial`,
  `skill_loaded`, `failing_hyp`, `f_reading_pin`, `f_body_limit` and
  `end_state` (`recorded`, `registered` or `none`, from the events of the
  actor `runner-ARM-TRIAL` in `changes/.tree/events/`). The script's
  docstring defines each field. It exits 2 when OUT_DIR lacks `state.env`,
  `transcript.jsonl` or `changes.json`.

`known-answers/score-*` hold a clean trial, a trial with no skill load and
two failing hyp calls, each fingerprint alone and both together, a limit
that is not the body's, and end states with events of another actor. Every
expected output was written by hand from the docstring.

Run verdict: ./verdict.py
External: none

- `verdict --row F [--row F]... SCORE...` reads one score file per trial.
  A row reproduces when `base` has k >= 3 of 5 trials with `F` = 1, and a
  variant arm passes it at k <= 1; a trial with `skill_loaded` 0 never
  reproduces in `base` and always fails in a variant. The script's
  docstring states the rule. It prints `verdict: repro=…` and `verdict:
  ARM=…` for each other arm. It exits 2 on a usage error, a missing field,
  no base arm, or an arm without exactly 5 distinct trials.

`known-answers/verdict-*` hold a passing and a failing variant, a base that
does not reproduce, several rows and arms, no-skill trials in each arm at
their thresholds, and every refusal. Every expected output was written by
hand from the docstring.
