---
name: skill-gate
version: 1
format: 1
subject: a hyp skill-text change, gated by 5 fresh claude -p runner trials per arm in hyp sandboxes
---

# skill-gate

The instrument of the pre-registered method in METHOD.md: a change to
`skills/*/SKILL.md` ships only when `base` shows a failure in 3 or more of
5 trials and the changed text in 1 or fewer of 5. One sandbox is one trial.

Per-run inputs: a file given to `hyp sandbox create --inputs FILE`, lines
`key=value` (blank lines and `#` lines ignored):

- `arm`: `base`, or the arm's name (`[a-z][a-z0-9-]{0,31}`).
- `trial`: the trial number, a positive integer.
- `brief`: absolute path of the file whose text is the runner's stdin.
- `variant`: absolute path of the arm's variant directory, holding
  `hyp-prototype/SKILL.md` and/or `hyp-verification/SKILL.md`; absent (or
  empty) when `arm=base`, and required otherwise.

The sandbox is made as `hyp sandbox create ID --host H --base COMMIT
--recipe skill-gate@1 --store SRC --cut EVENT --inputs FILE --hyp BIN`. The
runner's `hyp` is the sandbox's pinned `hyp BIN`, the same in every arm.
After `hyp sandbox drive ID`, remove with
`hyp sandbox remove ID OUT_DIR --keep tmp/skill-gate/state.json
--keep tmp/skill-gate/transcript.jsonl --keep tmp/skill-gate/summary.json
--keep tmp/skill-gate/events.json`. OUT_DIR then holds the runner's
transcript, its run summary, the tree's events as hyp prints them
(`events.json`), the state Launch wrote, and, from remove itself, every
file the runner added or changed under `changes/` (the tree's events
included). Then score each trial with the Evidence command, and judge all
trials with the Standalone command.

## Launch

Make the sandbox what the method says it is before the runner starts:
remove `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`, `plans` and
`PLAN.md`/`PLANS.md` from `wt/`, and `METHOD.md` and `investigations/` (the
method and earlier investigations are not the runner's brief); move the
two skills `wt/skills/hyp-prototype` and `wt/skills/hyp-verification` to
`wt/.claude/skills/` so the runner loads them as project skills; lay the arm's variant
`SKILL.md` files over them. The tree is already at the cut
(`--store`, `--cut`). State goes to `$HYP_SANDBOX_RECIPE_TMP/state.json`:
arm, trial, brief hash, the sha256 of each skill file the runner loads, and
the ids of the events at the cut.

Run: ./launch.py

## Doctor

Prove the sandbox holds only hyp's code, the two skills (with the arm's
variant) and the tree at the cut: no agent files, no `.claude` beyond the
skills, no plans, each skill file equals what Launch recorded, a variant file
is what the arm's variant directory holds, the tree's events are exactly
the cut's, `hyp` runs, and `claude` is on the `PATH`.

Run: ./doctor.py

## Drive

Start the runner: a fresh `claude -p --model sonnet` with project settings
only (`--setting-sources project`), the brief on stdin, in `wt/`, with
`HYP_ACTOR=runner-<trial>`. It is stopped after 50 minutes (its process
group is killed). Its stream-json session transcript goes to
`$HYP_SANDBOX_RECIPE_TMP/transcript.jsonl`, a summary of its run (exit, turns,
cost, timed out) to `summary.json`, and when it ends `hyp events --json` of
the tree to `events.json`. A runner that fails or times out is still a trial:
the drive exits 0 once the summary is written. A second drive of the same
sandbox is refused.

Run: ./drive.py

## Evidence

Score one trial: prints `key<TAB>value` lines `arm`, `trial`,
`skill_loaded`, `failing_hyp`, `f_reading_pin`, `f_body_limit`, `end_state`
(rules in METHOD.md, "Scoring one trial"). It reads the kept files in
OUT_DIR (the last argument).

Run: ./score.py
External: OUT_DIR of hyp sandbox remove (the subject); writes nothing

## Cleanup

Nothing is left running: Drive kills the runner's process group, and the
sandbox's remove stops any group that remains. The command succeeds when run
twice and when Launch never ran.

Run: ./cleanup.py

## Standalone

Judge all trials: `./verdict.py ROW... SCORE...`, where each ROW is a
fingerprint key (`f_reading_pin`, `f_body_limit`) and each SCORE is the
absolute path of one trial's score (the `stdout/0` of its record). Needs
exactly 5 trials per arm with distinct trial numbers and an arm `base`.
Prints `verdict: repro=supports|contradicts`, then
`verdict: ARM=supports|contradicts|inconclusive` for each variant arm
(sorted). Exit 0; exit 30 and no stdout for a refusal.

Run verdict: ./verdict.py
External: the SCORE files named by the arguments (read only); writes nothing
