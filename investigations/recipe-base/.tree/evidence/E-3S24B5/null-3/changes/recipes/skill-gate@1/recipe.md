---
name: skill-gate
version: 1
format: 1
subject: A change to a hyp skill's text (skills/*/SKILL.md), judged by fresh claude -p runners started from a recorded point of a hyp tree
---

# skill-gate

The instrument of the pre-registered method in `METHOD.md`: a change to a
skill's text ships only when `base` shows a failure in 3 or more of 5 trials
and the changed text in 1 or fewer of 5. One trial is one sandbox, one
runner, one score. The tree, the store cut, the brief and the variant
directories are inputs of a run, not part of this bundle.

## Inputs

`hyp sandbox create ID --host H --base REV --store SRC --cut EVENT --recipe skill-gate@1 --inputs FILE`.
FILE holds `key=value` lines, each key once:

- `arm`: `base`, or the arm's name (`[a-z][a-z0-9-]{0,31}`, not `repro`);
- `trial`: the trial number, a positive whole number; an arm's five trials
  carry distinct numbers;
- `variant`: for an arm other than `base`, the absolute path of a directory
  `variant/<skill>/SKILL.md`, one directory per skill it changes, each
  named after a skill of `skills/` and holding `SKILL.md` and nothing else;
  absent for `base`;
- `brief`: the absolute path of the brief the runner gets on stdin, the
  same file for every trial of a gate.

Run five trials of `base` and five of each variant arm, one sandbox each.
A variant directory is an example, not part of the bundle: Launch records
the sha256 of every `SKILL.md` it lays over in `trial.json`.

## Launch

Check the inputs. Lay the arm's `SKILL.md` files over the sandbox's
`skills/<skill>/SKILL.md`. Remove `AGENTS.md`, `CLAUDE.md`, `.claude`,
`.agentkit`, `plans/` and `investigations/` from `wt/`, then copy `skills/*`
to `.claude/skills/` so the runner loads exactly the two skills with the
variant over them. `wt/.tree` is the store at the cut, written by
`hyp sandbox create --store --cut` before this runs. Keep the brief and
write `trial.json` (arm, trial, commit, brief and skill hashes) under
`$HYP_SANDBOX_RECIPE_TMP`. The sandbox's manifest is made after Launch, so
`changes.json` lists only what the runner adds or changes.

Run: ./launch.sh

## Doctor

Nothing agent-shaped is left in `wt/`, `.claude` holds only the two skills,
equal to `skills/`, the tree is at the cut, and `claude`, `python3` and `hyp`
run.

Run: ./doctor.sh

## Drive

`hyp sandbox drive ID` runs the runner once: a fresh `claude -p` session
with the model `sonnet`, `--setting-sources project`, the brief on stdin,
`wt/` as its working directory, `HYP_ACTOR=runner-ARM-TRIAL`, the sandbox's
pinned `hyp` first on `PATH` (the same hyp in every arm), and permissions
bypassed (the sandbox is its disposable clone). It is stopped after 50
minutes. The runner exiting non-zero or timing out is data, recorded in
`run.json`; Drive fails only when it cannot start `claude` or find the
session transcript. A sandbox is driven once; a second Drive refuses.

Run: ./drive.py

## Evidence

After `hyp sandbox remove ID OUT_DIR` with
`--keep tmp/skill-gate/trial.json --keep tmp/skill-gate/run.json --keep tmp/skill-gate/transcript.jsonl --keep tmp/skill-gate/brief.md`,
OUT_DIR holds the transcript, the summary (`run.json`: exit, turns, cost),
`trial.json`, `sandbox.json`, the inputs, `tree/` (the store after the run, the
runner's events included) and `changes/` with every file the runner added or
changed.

`score` prints `key<TAB>value` lines for one trial:

- `arm`, `trial` from `trial.json`;
- `skill_loaded`: 1 when the transcript has a Skill tool call or a Read of
  a file named `SKILL.md`;
- `failing_hyp`: the number of Bash calls whose tool result is an error and
  whose exit status is a hyp's: the last command of the call, or an earlier
  one joined to it by `&&` alone, with a pipeline's status that of its last
  stage, is `hyp` (env assignments, `command`, `exec`, `env`, `time` and a
  path ending `/hyp` allowed); `hyp x | head`, `hyp x; echo` and
  `hyp x || true` do not count;
- `f_reading_pin`: 1 when an errored tool result holds `--decided-by reading
  contradicts recipe REF, an instrument whose checks decide the verdicts`;
- `f_body_limit`: 1 when an errored tool result holds `over limit: body has
  N runes (max 600)`;
- `end_state`: `recorded` when the runner's events hold a `result.record`,
  `registered` when they hold an Experiment `node.add` and no Result,
  else `none`.

Only `assistant` and `user` lines of the transcript are read, because
`api-request-blob` lines repeat every tool call. The runner's events are the
applied events of the kept store `OUT_DIR/tree` (`wt/.tree` after the run,
read through `hyp events --json`) stamped after `sandbox.json`'s `created`;
the cut's own events are older. Only an Experiment `node.add` and a
`result.record` are looked for.

Run: ./score.py
External: OUT_DIR of hyp sandbox remove (one trial); writes only $HYP_OUT

## Standalone

`verdict ROW... SCORE_FILE...` reads the scores of every trial: ROW is a
fingerprint name (`f_reading_pin`, `f_body_limit`), SCORE_FILE the absolute
path of one trial's `score` stdout. It needs at least one row, a `base` arm,
and exactly 5 trials per arm with distinct trial numbers (else exit 2). For
each row and arm, k counts the trials whose fingerprint is 1; a trial with
`skill_loaded` 0 never counts for `base` and always counts for any other
arm. A row reproduces when `base` has k >= 3; a variant arm passes a row at
k <= 1. It prints `k ROW ARM K` lines, then `verdict: repro=supports` when
any row reproduces, else `verdict: repro=contradicts`, and for each variant
arm `verdict: ARM=supports` when it passes every reproducing row,
`contradicts` when it fails one, `inconclusive` when no row reproduces. The
operator still types each verdict into `hyp record`.

Run verdict: ./verdict.py
External: the score files named in the arguments; writes nothing

## Cleanup

Kill the runner's process group if one is left. Safe to run twice and when
Launch never ran.

Run: ./cleanup.sh
