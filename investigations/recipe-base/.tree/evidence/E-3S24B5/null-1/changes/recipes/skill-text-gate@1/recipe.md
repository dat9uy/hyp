---
name: skill-text-gate
version: 1
format: 1
subject: A change to a skill's text (skills/*/SKILL.md), tried by fresh claude runners from a recorded point of a hyp tree
---

# skill-text-gate

The instrument of the pre-registered skill-text gate (METHOD.md). A change
to a skill's text ships only when a fresh agent runner, started from a
recorded point of a hyp tree, shows a failure in 3 or more of 5 trials with
the skills as they are (`base`), and in 1 or fewer of 5 with the changed
text (a variant arm).

One trial is one sandbox. Trials and arms are separate runs of this
recipe; the verdict is the one Standalone command, `verdict`, over every
trial's score.

## The inputs file

`hyp sandbox create ID --host H --base REV --recipe skill-text-gate@1
--store SRC --cut EVENT --inputs FILE` takes one inputs file. Its lines up
to the first blank line are `key=value` headers; everything after the blank
line is the brief the runner gets on stdin.

```text
arm=base            # base, or the name of a variant: [a-z][a-z0-9-]{0,31}, not "repro"
trial=3             # a whole number from 1
variant=/abs/dir    # not for base; a directory holding <skill>/SKILL.md, one per skill the arm changes

<the brief>
```

Run the same recipe with `--store SRC --cut EVENT` for the tree's point.
The store, the cut, the variant directories and the brief are the run's
inputs; the recipe takes them as they are.

When the sandbox is removed, keep the four files Launch and Drive leave:

```text
hyp sandbox remove ID OUT_DIR \
  --keep tmp/skill-text-gate/transcript.jsonl \
  --keep tmp/skill-text-gate/summary.json \
  --keep tmp/skill-text-gate/cut-events.txt \
  --keep tmp/skill-text-gate/arm.tsv
```

## Launch

Run: ./launch.sh

Reads the inputs file. Removes the agent files from `wt/` (`AGENTS.md` and
`CLAUDE.md` anywhere, `.claude`, any `plans` directory, and `METHOD.md`), lays the
variant's `<skill>/SKILL.md` files over `wt/skills/` (nothing for `base`),
and gives `wt/.claude/skills/` a copy of each skill directory, the only
thing under `.claude`. The store cut was written to `wt/.tree` by hyp
before this runs. Writes under `$HYP_SANDBOX_RECIPE_TMP`: `brief.txt`,
`arm.tsv` (every skill file, its sha256, and whether the variant or the
commit supplied it) and `cut-events.txt` (the id of every event in the
cut, so the score can tell the runner's events from the tree's).

## Doctor

Run: ./doctor.sh

The sandbox holds both hyp skills, each under `skills/` and
`.claude/skills/` byte for byte; no agent file or plan; nothing under
`.claude` but the skills; `base` skills equal the commit's; a variant's
files equal their source and every other skill file equals the commit's;
the brief is not empty; `claude` and `hyp` run.

## Drive

Run: ./drive.py

Starts one fresh `claude -p` session in `wt/` with the model `sonnet`,
project settings only, and the brief on stdin; stops it after 50 minutes.
Keeps its session transcript as `transcript.jsonl` and a `summary.json`
(exit, timed_out, turns, cost, session id) under `$HYP_SANDBOX_RECIPE_TMP`.
A runner that fails or times out is a trial like any other; Drive fails
only when it cannot start the runner or finds no transcript. The runner
sees the same `hyp` in every arm.

## Evidence

Run: ./score.py
External: OUT_DIR of hyp sandbox remove (the trial's kept files); writes nothing

Prints `key<TAB>value` lines from one trial's kept files: `arm`, `trial`,
`skill_loaded`, `failing_hyp`, the fingerprints `f_reading_pin` and
`f_body_limit`, and `end_state`. Run it with
`hyp recipe evidence --record DIR skill-text-gate@1 OUT_DIR`; save its
`stdout/0` as that trial's score file.

## Cleanup

Run: ./cleanup.sh

Kills the runner's process group when Drive left one, and waits for it.
Exits 0 with nothing to do.

## Standalone

Run verdict: ./verdict.py
External: the score files named in the arguments (absolute paths), read only

`hyp recipe run --record DIR skill-text-gate@1 verdict --row ROW ...
SCORE...` reads every trial's score file and judges the rows given with
`--row` (each a fingerprint, `f_...`). Exactly 5 trials per arm, distinct
trial numbers, and a `base` arm are required, else exit 2 and no verdict.
A trial that never loaded a skill is not a reproduction in `base`, and a
failure in any other arm. It prints one `k: ROW ARM N` line per row and
arm, then `verdict: repro=supports|contradicts`, then
`verdict: ARM=supports|contradicts|inconclusive` for each variant arm in
name order.
