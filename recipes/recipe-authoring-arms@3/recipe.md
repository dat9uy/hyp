---
name: recipe-authoring-arms
version: 3
from: recipe-authoring-arms@2 58ae7f45ee4ac5c62bc59e675e23aae878fbef7b8fc859ce858a7967b467878a
format: 1
subject: A sonnet runner with only hyp and its two skills turns a written method into a recipe instrument, on two hyp arms (main, and main plus a first-bundle skeleton); stops are counted from its transcript and bundle (self-review, #235)
---

# recipe-authoring-arms

e-author's instrument for #235 (`investigations/recipe-authoring`). One
sandbox per trial: Launch strips the clone down to hyp, its two skills and
a method; Drive runs a `claude -p` runner on it with the sandbox's pinned
hyp (the arm, `--hyp BIN`); `score` counts the trial's stops and judges the
bundle it left; `verdict` applies the pre-registered bar to every trial.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`
(`main` or `scaffold`, the label of the `--hyp` binary), `trial` and
`model`. It deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`,
`plans`, `investigations` and `recipes` from `wt/`, copies
`skills/hyp-prototype` and `skills/hyp-verification` into
`wt/.claude/skills/`, and writes `method.md` as `wt/METHOD.md`. These
files are written before the manifest, so they are not changes.

`method.md` is #240's e-arms method (`plans/261101-0708-240-id-citation/assets/e-arms-method.md`
on main at 92d7c5d) with these edits, so that it names nothing the runner
cannot have: the fixture's and cases file's sha256 lines and the ids from
#240's own fixture are removed, the arms are described as binaries given at
run time, "phase 3" is reworded, a line says the tree is not part of the
task, and the Disclosure and Correction sections, written after #240's run,
are dropped.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run, the deleted paths
are gone, `METHOD.md` is there and `.claude/` holds only the two skills.

## Drive

Run: ./drive.py

Runs `claude -p --model MODEL --setting-sources project --permission-mode
dontAsk --allowedTools Bash Read Edit Write Glob Grep Skill` in `wt/` with
`brief.txt` on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables, and
stops it after 50 minutes. It copies the session transcript to this
recipe's tmp as `transcript.jsonl` and writes `drive.json` (exit, turns,
cost, denials). It exits 1 when no transcript was written.

## Evidence

Kept by hand at remove: `--keep tmp/recipe-authoring-arms/state.env
--keep tmp/recipe-authoring-arms/drive.json --keep
tmp/recipe-authoring-arms/transcript.jsonl`. The runner's bundle reaches
OUT_DIR through `changes/`. Then `score` runs on that OUT_DIR, as a
Standalone command: a known-answer input cannot hold the `@` paths under
`changes/recipes/`, so its cases carry each trial directory as a git bundle.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py

Run verdict: ./verdict.py

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`,
  `trial`, `failing_hyp` (Bash calls that end in error with hyp as their last command, the one whose exit status the call returns, after heredoc bodies are dropped), one
  0/1 fingerprint per stop hyp could carry (`f_sandbox_json`: the runner
  wrote a `sandbox.json`; `f_tar_input`: an archive in a known-answers
  input that `method.md` does not name; `f_go_file`: a `.go` or `.go.txt` file in a bundle;
  `f_recipe_none`: `hyp add exp --recipe none`; `f_weak_cases`: `hyp
  recipe check` warns about the bundle's cases), `stops` (their sum), and
  the end state: the bundle under `recipes/` rebuilt from `changes/`,
  `known_answers_exit` (`hyp recipe known-answers`), `add_exp_exit` (`hyp
  add exp --recipe` in a scratch tree) and `end_state` (`pass` when both
  are 0). Every hyp it calls is the hyp first on PATH, the operator's,
  the same judge for both arms. It exits 2 when OUT_DIR lacks
  `state.env`, `transcript.jsonl` or `changes.json`.
- `verdict SCORE...` reads score outputs and applies e-author's bar
  (the rule is in the script's docstring): with m and s the mean stops of
  main and scaffold and pm and ps their end-state passes, the null region
  (m <= 1 and pm >= 2/3 of main's trials) gives `h-binary=contradicts`,
  `h-partial=supports`, `h-null=supports`; else the bar (s <= m/2 and
  ps >= pm) gives `h-binary=supports`, `h-partial=contradicts`,
  `h-null=contradicts`; else every label is inconclusive. It exits 2 when
  a file lacks a field or names another arm, or an arm has no trial.

`known-answers/` holds, for `score`, a clean trial (a passing bundle, an
error that is not a hyp call, a success that is), a trial that hits every
fingerprint with a bundle that has no `format: 1` line, a trial with no
transcript and a usage failure; and, for `verdict`, the bar, the null
region, the inconclusive band and a missing arm. Every expected output was
written by hand from the rules above. Version 2: @1's score read
changes.json's mode as a number, but hyp writes it in ls form
(`-rwxr-xr-x`), so score crashed on every real trial; @1's fixture had
used numbers. @2 reads the ls form, and its fixture's changes.json files
use it. Version 3 (#235, after R-8Y0TZB): failing_hyp counted any
errored call that mentioned hyp, so a permission denial on a generator
script counted; it now counts only a call whose last command is hyp. And
f_tar_input fired on every trial for the fixture.tar the method itself
asks for; an archive method.md names no longer counts. Each trial directory is a git bundle
holding HEAD and main, made with fixed author dates.
