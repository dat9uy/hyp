---
name: skill-text-arms
version: 3
from: skill-text-arms@2 c4ec8536155fef10ff67aedd712f6de7ddf5714d83e00ff86cc59c38dfcda303
format: 1
subject: A fresh sonnet runner with hyp and its two skills (base, or a SKILL.md variant) runs the id-citation tree's e-arms from that tree's store cut; skill-text stops are counted from its transcript and changes, and the skill-text gate is applied (self-review)
---

# skill-text-arms

The skill-text gate's instrument (AGENTS.md, "Improving hyp with hyp").
One sandbox per trial: `hyp sandbox create --store
investigations/id-citation/.tree --cut 01M6A7K0994S8115J0ZHBMN916` puts
the id-citation tree in `wt/.tree` as it was at the last `node.add` before e-arms;
Launch strips the clone down to hyp, its two skills (the arm's variant over
them) and what e-arms needs; Drive runs a `claude -p` runner that
pre-registers e-arms, runs it and records its Result; `score` reads one
trial's fingerprints; `verdict` applies the gate.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm` (`base`,
or a directory under `variants/`), `trial`, `model`, and for each arm
binary `today`, `text-warn`, `text-refuse` and `cite`, `bin_<arm>` (an
absolute path) and `sha_<arm>` (its sha256). It exits 2 when a line is
missing, a binary is not a file or its sha256 differs, the arm is unknown
or `wt/.tree` is missing. The binaries are built once by the operator from
the id-citation investigation's arm patches with `CGO_ENABLED=0 go build
-trimpath`, and each is named by the sha256 the inputs file gives; the
runner never builds them.

It copies each binary to `wt/arms/arm-<arm>` and writes
`wt/arms/arms.sha256`; keeps `recipes/id-citation-arms@1` from `wt/` and
deletes `AGENTS.md`, `CLAUDE.md`, `.claude`, `.agentkit`, `plans`,
`investigations` and every other recipe; copies `skills/hyp-prototype` and
`skills/hyp-verification` into `wt/.claude/skills/`, then, for an arm other
than `base`, each `variants/<arm>/<skill>/SKILL.md` over them; and writes
`fixture/` as `wt/fixture/`. The method is not written to `wt/`: the
runner's brief is `brief.txt` followed by `method.md`. These files
are written before the manifest, so they are not changes. Every product of
the trial is written in Drive.

`method.md` is the id-citation tree's e-arms method as that
investigation wrote it (5184 bytes), with these edits, so that it names nothing the runner
cannot have: the tree is "this repository's tree (`.tree/`)", not
`investigations/id-citation`; the method, fixture and cases file are
"given", not "committed"; `fixture.sh` is not named; the arms are named by
sha256 in `arms/arms.sha256` and built already as `arms/arm-<arm>`, with
no patch path; "phase 3" is reworded; and the Disclosure and Correction
sections, written after the id-citation run, are dropped. It stays over a body's
600 runes (4191 bytes), which S1 needs. `fixture/` is that
investigation's fixture unchanged (`fixture.tar`, `cases.tsv`, `ids.txt`).

`brief.txt` is new. The brief of the recipe-authoring instrument this
recipe was first made from told its runner not to create a tree and not
to run the Experiment, so none of its trials reached these stops.

Version 2: @1 wrote the method as `wt/METHOD.md`,
which gave it a home a plans-less author lacks, so S1 could not show; a
verifier judged that limit open. @2 carries the method only in
the brief, and Doctor fails when `METHOD.md` is in `wt/`; its first
paragraph says "This method" for "This file". @1's brief can
never reach P5 (no runner writes a Launch); @2 keeps that brief, so P5 is
not probed by either version.

Version 3: @2's `score-usage` case held `expect.json` but no `input/`,
because git keeps no empty directory, so the known-answer gate refused
the bundle (`known answer score-usage: input/ is missing`) and no run of
it could start. @3 adds `input/README.txt` to that case, and this text no
longer names plan files, commits or tracker issues. The scripts, the
method, the fixture and every expected output are unchanged.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run; the deleted paths
are gone; `recipes/` holds only `id-citation-arms@1`, with the hash e-arms
pinned (3fb1e139…); `hyp events` in `wt/` ends at the cut; `fixture/` and the four binaries matching `arms/arms.sha256` are there; and
`METHOD.md` is absent, `.claude/` holds only `skills/`, which for `base` holds only the two hyp
skills. Git-ignored leftovers under the host's `investigations/` cannot
reach `wt/`: the clone holds only committed files, and Doctor checks that
`investigations` is gone.

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

Kept by hand at remove: `--keep tmp/skill-text-arms/state.env --keep
tmp/skill-text-arms/drive.json --keep tmp/skill-text-arms/transcript.jsonl`.
The runner's tree and files reach OUT_DIR through `changes/`. Then `score`
runs on that OUT_DIR, as a Standalone command.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for the runner to end.

## Standalone

Run score: ./score.py

Run verdict: ./verdict.py

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`,
  `trial`, `skill_loaded`, `failing_hyp` (@3's definition), the 0/1
  fingerprints `f_reading_pin` (P6), `f_body_limit` (S1) and
  `f_launch_product` (P5), the reading `r_keep_tree` (P4, never a failure
  count), `exp_added`, `result_recorded` and `end_state` (`recorded`,
  `registered` or `none`). The script's docstring defines each field. It
  exits 2 when OUT_DIR lacks `state.env`, `transcript.jsonl` or
  `changes.json`.
- `verdict --row ROW [--row ROW]... SCORE...` applies the gate to every
  trial: per row, base reproduces at 3 or more of 5 and a text arm passes
  at 1 or fewer of 5; a trial with `skill_loaded` 0 counts against its arm
  (never a reproduction for base, always a failure for a text arm). It
  prints `verdict: repro=…` and `verdict: none=…` (any row reproduces, or
  none does) and `verdict: ARM=…` for each text arm. It exits 2 when a
  field or the base arm is missing, or an arm has other than 5 trials. The
  script's docstring states the rule.

`known-answers/` holds, for `score`, a trial that hits every fingerprint
(the body refusal, the reading-pin refusal, a Launch that writes into wt
and a re-run remove, with three `--keep` values under one directory), a
clean trial that loaded no skill and recorded its Result, a trial with no
transcript and a usage failure; and, for `verdict`, a reproducing base
(`repro=supports`), a base whose third failure loaded no skill
(`repro=contradicts`), a passing and a failing text arm, an arm with four
trials and a missing field. Every expected output was written by hand from
the rules above.
