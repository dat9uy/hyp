---
name: claude-skill-gate
version: 1
from: base:claude-runner-gate@1 656b478b4a4716b3498b10d3c13e5572bfea2a41359bba3c8bfe61f07e649c63
format: 1
subject: hyp's skills (skills/*/SKILL.md), as a fresh Claude Code runner (claude -p, model sonnet) uses them from a recorded point of a hyp tree
---

# claude-skill-gate

The instrument of the pre-registered skill-text gate (METHOD.md): a change
to a skill's text ships only when `base` shows a failure in 3 or more of 5
trials and the changed text in 1 or fewer of 5. One sandbox per trial: Launch
sets the arm's skills, Drive runs the runner and keeps its transcript,
`score` scores one trial's evidence and `verdict` applies the gate to all of
them. It drives `claude -p` only.

Create each sandbox from this repository at a pinned commit, with the tree's
store at the store cut, and with the same hyp in every arm:

    hyp sandbox create ID --host H --base REV --recipe claude-skill-gate@1 \
      --store SRC --cut EVENT --inputs FILE --hyp BIN

The inputs FILE has `KEY=VALUE` lines `arm` (an alias-like name; `base` is
the arm with the skills as the commit has them), `trial` (a number, distinct
within the arm), `model` (`sonnet`) and, for every arm but base, `variant`:
the absolute path of a directory holding `<skill>/SKILL.md`, one per skill
it changes (`hyp-prototype`, `hyp-verification`). The variant directories
are inputs of the run, named by the caller.

## Launch

Run: ./launch.sh

Reads the inputs file and exits 2 when `arm`, `trial` or `model` is
missing, when `base` has a `variant=`, or when another arm's variant is not
an absolute directory holding only `<skill>/SKILL.md` for the two skills.
It copies the inputs to this recipe's tmp as `state.env`, deletes
`AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit` and `plans` from `wt/`,
copies the commit's two skills into `wt/.claude/skills/`, lays the
variant's `SKILL.md` files over them, writes their hashes to `variant.sha256`
and copies `brief.txt` to this recipe's tmp as the runner's brief. The tree
at the cut is in `wt/.tree` already (hyp wrote it).

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the stripped paths
are gone, `.claude/` holds only `skills/` with both skills, each variant
file is the one `variant.sha256` names, and the brief exists.

## Drive

Run: ./drive.py

Runs `claude -p --model MODEL --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
the brief on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (arm, trial, model, exit, num_turns, cost_usd,
denials). It exits 1 when no transcript was written.

## Evidence

Kept by hand at remove, per trial: `--keep tmp/claude-skill-gate/state.env
--keep tmp/claude-skill-gate/drive.json --keep
tmp/claude-skill-gate/transcript.jsonl --keep
tmp/claude-skill-gate/variant.sha256`. Remove copies the sandbox record,
`changes.json` and every file the runner added or changed (its tree's
events included) to OUT_DIR. Then `score` runs on that OUT_DIR as a
Standalone command, once per trial, and `verdict` once over all the scores.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial's evidence), read only

Run verdict: ./verdict.py
External: the score files of every trial, read only

- `score OUT_DIR` prints `arm`, `trial`, `skill_loaded`, `failing_hyp`,
  `f_reading_pin`, `f_body_limit` and `end_state` as `key<TAB>value` lines.
  The script's docstring defines each field. It exits 2 when OUT_DIR lacks
  `state.env`, `drive.json`, `transcript.jsonl`, `sandbox.json` or
  `changes.json`.
- `verdict --row ROW... SCORE...` reads one score file per trial and judges
  the given rows (fingerprint keys such as `f_reading_pin`). A row
  reproduces when `base` has k >= 3 of 5; a variant arm passes it at k <= 1;
  a trial without a loaded skill is a failure of a variant arm and never a
  reproduction in base. It prints `verdict: repro=supports|contradicts` and
  `verdict: ARM=supports|contradicts|inconclusive` per variant arm. The
  script's docstring states the rule. It exits 2 on a usage error, a
  missing field or an arm without exactly 5 distinct trials.

`known-answers/` holds, for `score`, every field and fingerprint (a hit and
a miss of each), end states `recorded`, `registered` and `none`, events
from before the sandbox, a piped and a non-hyp error, and each missing
file; and, for `verdict`, every branch: reproduces and passes, reproduces
and fails, no reproduction, several rows and arms, an unloaded skill in base
and in a variant arm, and each refusal. Every expected output was written
by hand from the METHOD's rules.
