---
name: recipe-base-arms
version: 1
format: 1
subject: A sonnet runner with only hyp and its two skills turns the skill-text gate method into a recipe instrument, on three hyp arms (null, coarse base, fine bases); stops, end state and lines edited past the bases are counted (self-review, #270)
---

# recipe-base-arms

e-base's instrument for #270 (`investigations/recipe-base`). One sandbox
per trial: Launch strips the clone down to hyp, its two skills and a
method; Drive runs a `claude -p` runner on it with the sandbox's pinned
hyp (the arm, `--hyp BIN`); `score` counts the trial's stops, judges the
bundle it left and measures how much of it the runner wrote past the
bases it named; `verdict` applies the pre-registered rule to every trial.

This bundle was copied by hand from `recipe-authoring-arms@3` (hash
ddb06ff031eb47c316d095248e8ebfd3a2f25b69a02e91992468c91b33ef827f), not
with `hyp recipe new`, because `--from` copies the same name only and
no base exists yet: the one bootstrap exception the scope map records.
`launch.sh`, `doctor.sh` and `cleanup.sh` are @3's, Launch also refusing
an arm other than `null`, `coarse` or `fine`; `drive.py` is @3's plus
`HYP_ACTOR=runner-<arm>-<trial>`; `score.py` is @3's plus `model`,
`cost_usd`, `base_from`, `edited_lines`, `unused_base_files` and `seam`.

The arms are hyp binaries named by sha256 in
`investigations/recipe-base/arms.txt`: `null` is main, `coarse` embeds
`base:claude-runner-gate@1`, and `fine` embeds `base:claude-runner@1`,
`base:transcript-score@1`, `base:gate-verdict@1` and
`base:fixture-check@1`, all from the unmerged prototype branch
`proto/270-base`. `bases.tar` holds those five bases as the arms embed
them, at `<name>@<N>/<path>`, so `score` can tell copied lines from
written ones.

`method.md` is a skill-text gate method written for this instrument from
the AGENTS.md skill-text gate: a fresh sonnet runner at a store cut, with
the skills as they are or a SKILL.md variant over them, scored from its
transcript, and passed at 3 or more of 5 failing without the change and 1
or fewer of 5 with it. It names no file of the held-out recipe it
describes. `brief.txt` is @3's, with the method's own nouns, and one
neutral line every arm gets: "hyp's own help may list starting points for
a recipe." The brief is byte-identical across arms, so only the binary
differs; on `null` the help lists no base.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`
(`null`, `coarse` or `fine`, the label of the `--hyp` binary), `trial`
and `model`. It deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`,
`plans`, `investigations` and `recipes` from `wt/`, copies
`skills/hyp-prototype` and `skills/hyp-verification` into
`wt/.claude/skills/`, and writes `method.md` as `wt/METHOD.md`. These
files are written before the manifest, so they are not changes.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the deleted paths
are gone, `METHOD.md` is there and `.claude/` holds only the two skills.

## Drive

Run: ./drive.py

Runs `claude -p --model MODEL --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
`brief.txt` on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=runner-<arm>-<trial>`, and stops it after 50 minutes. It copies
the session transcript to this recipe's tmp as `transcript.jsonl` and
writes `drive.json` (exit, turns, cost, denials). It exits 1 when no
transcript was written.

## Evidence

Kept by hand at remove: `--keep tmp/recipe-base-arms/state.env --keep
tmp/recipe-base-arms/drive.json --keep
tmp/recipe-base-arms/transcript.jsonl`. The runner's bundle reaches
OUT_DIR through `changes/`. Then `score` runs on that OUT_DIR, as a
Standalone command: a known-answer input cannot hold the `@` paths under
`changes/recipes/`, so its trial cases carry each trial directory as a git
bundle.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py
External: reads OUT_DIR, and runs the hyp first on PATH (the operator's) on the runner's bundle in a scratch repository

Run verdict: ./verdict.py
External: none

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`,
  `trial`, `model`, `cost_usd`, `failing_hyp` and @3's fingerprints,
  `stops`, the end state (`bundle`, `known_answers_exit`,
  `add_exp_exit`, `end_state`, judged by the operator's hyp, the same
  judge for every arm), and the copy measures `base_from`,
  `edited_lines`, `unused_base_files` and `seam`. The script's docstring
  defines each field. It exits 2 when OUT_DIR lacks `state.env`,
  `drive.json`, `transcript.jsonl` or `changes.json`.
- `verdict SCORE...` applies e-base's rule (stated in the script's
  docstring): a base arm beats null when its end-state passes are no
  fewer and its mean stops at least 1 lower; h-null supports when neither
  beats null; one that beats alone supports; when both beat, the lower
  mean `edited_lines` supports unless the two are within 10%, when both
  are inconclusive. With 3 trials per arm and a base arm's gap in [0.5,
  1.5] it prints `extend: yes` and every label inconclusive. It exits 2
  when a trial's model is not `sonnet`, a field or an arm is missing, or
  the arms' trial counts differ or are not 3 or 5.

`known-answers/` holds, for `score`, a null trial (a hand-made bundle, one
failing hyp call, a case input holding a transcript), a fine trial copied
from `base:gate-verdict@1` with two lines edited, a subject written and a
case deleted, a trial with no `drive.json` and a usage failure; and, for
`verdict`, h-null supported (a base arm with fewer stops but fewer
passes), coarse alone beating null, fine alone at the bar with 5 trials,
both beating with coarse editing less, both within 10%, the extension, a
trial run on another model, a missing arm and unequal trial counts. Every
expected output was written by hand from the rules above and checked
against the script.
