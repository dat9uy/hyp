---
name: skill-text-arms
version: 5
from: skill-text-arms@4 430d9cc597ee30c3bd3ac663dfea11a8471869e8c13ef62e39a52f9a368e536d
format: 1
subject: A fresh sonnet runner with hyp and its two skills (base, or a SKILL.md variant given as the trial's --patch) runs the id-citation tree's e-arms from that tree's store cut; skill-text stops are counted from its transcript and changes (self-review)
---

# skill-text-arms

The skill-text gate's instrument (AGENTS.md, "Improving hyp with hyp").
One sandbox per trial: `hyp sandbox create --store
investigations/id-citation --cut 01M6A7K0994S8115J0ZHBMN916` puts
the id-citation tree in `wt/.tree` as it was at the last `node.add` before e-arms;
Launch builds e-arms' four Arm binaries, strips the clone down to hyp, its
two skills (with the arm's SKILL.md patch applied) and what e-arms needs; Drive runs a `claude -p` runner that
pre-registers e-arms, runs it and records its Result; `score` reads one
trial's fingerprints. The gate itself, which turns the scores of all
trials into verdicts, is the rubric `rubrics/skill-text-gate@1`.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`, `trial`
and `model`. A SKILL.md variant is the trial's `--patch`: `arm` is `base`,
which takes no patch, or the ARM of the trial's one
`--patch ARM=FILE`, whose patch touches only `skills/*/SKILL.md`. Launch
exits 2 when a line is missing, `base` is given a patch, another arm is
given no patch, more than one or one of another name, the patch touches
another file or does not apply, or `wt/.tree` is missing.

It builds the four id-citation Arm binaries from `fixture/arms/`: the
commit in `fixture/arms/BASE` unpatched as the Arm `fixture/arms/base_arm`
names (`text-refuse`), and `today`, `text-warn` and `cite` as that commit
with `fixture/arms/<arm>.patch` applied alone. Each is built in two clean
copies of `wt/` checked out at that commit, at different paths, with
`CGO_ENABLED=0 go build -trimpath`, and Launch exits 2 when the commit is not
in the host's history, a patch does not apply, a build fails or the two
copies give different sha256 values. These binaries are the runner's
fixture, not this recipe's subject: they apply at their own commit, never
at the trial's `--base`, so a later base cannot break them. The runner
never builds them.

It copies each binary to `wt/arms/arm-<arm>` and writes
`wt/arms/arms.sha256`; applies the arm's patch to `wt/skills/`; keeps
`recipes/id-citation-arms@1` from `wt/` and deletes `AGENTS.md`,
`CLAUDE.md`, `.claude`, `.agentkit`, `plans`, `investigations` and every
other recipe; copies `skills/hyp-prototype` and `skills/hyp-verification`,
and any other skill the patch touches, into `wt/.claude/skills/`; and
writes `fixture/` as `wt/fixture/`. The method is not written to `wt/`: the
runner's brief is `brief.txt` followed by `method.md`. These files are
written before the manifest, so they are not changes. Every product of the
trial is written in Drive.

`method.md` is the id-citation tree's e-arms method as that
investigation wrote it (5184 bytes), with these edits, so that it names nothing the runner
cannot have: the tree is "this repository's tree (`.tree/`)", not
`investigations/id-citation`; the method, fixture and cases file are
"given", not "committed"; `fixture.sh` is not named; the arms are named by
sha256 in `arms/arms.sha256` and built already as `arms/arm-<arm>`, with
no patch path; "phase 3" is reworded; and the Disclosure and Correction
sections, written after the id-citation run, are dropped. It stays over a body's
600 runes (4191 bytes), which S1 needs. `fixture/` is that
investigation's fixture unchanged (`fixture.tar`, `cases.tsv`, `ids.txt`),
and `fixture/arms/` holds the Arm patches from
`investigations/id-citation/arms/` with `BASE` and `base_arm`.

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

Version 4: @3 read the four Arm binaries from absolute paths in the inputs
file (`bin_<arm>` and `sha_<arm>`), built by hand from patches that were
never published, and each SKILL.md variant from `variants/` in the bundle,
so it could not be run from the public repository. @4 builds the binaries
in Launch from `fixture/arms/`, and a SKILL.md variant is the trial's
`--patch`, kept and hashed by the sandbox. The brief, the method, Drive,
`score`, `verdict` and every expected output are unchanged.

Version 5: a recipe measures and never prints a verdict (ADR 0018), and
@4's `verdict` command printed the gate's verdicts, so no new Experiment
could pin @4. @5 drops `verdict`, its script and its five cases; they
moved unchanged into the rubric `rubrics/skill-text-gate@1`, whose cases
are named without the `verdict-` prefix. Doctor checks the pinned
id-citation recipe with `hyp bundle hash`, the name `hyp recipe hash`
now has. Launch, Drive, `score` and every other expected output are
unchanged.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3`, `go` and the pinned `hyp` run; the deleted
paths are gone; `recipes/` holds only `id-citation-arms@1`, with the hash
e-arms pinned (3fb1e139…); `hyp events` in `wt/` ends at the cut;
`fixture/` is there and `arms/` holds the four Arm binaries, each matching
`arms/arms.sha256`; `METHOD.md` is absent; and `.claude/` holds only
`skills/`, each `SKILL.md` there equal to wt's `skills/` copy. For `base`
it holds only the two hyp skills, each as the base commit has it; for a
text arm at least one `SKILL.md` differs from the base commit. Git-ignored
leftovers under the host's `investigations/` cannot reach `wt/`: the clone
holds only committed files, and Doctor checks that `investigations` is
gone.

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
External: OUT_DIR of hyp sandbox remove (the subject), read only; writes nothing

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`,
  `trial`, `skill_loaded`, `failing_hyp` (@3's definition), the 0/1
  fingerprints `f_reading_pin` (P6), `f_body_limit` (S1) and
  `f_launch_product` (P5), the reading `r_keep_tree` (P4, never a failure
  count), `exp_added`, `result_recorded` and `end_state` (`recorded`,
  `registered` or `none`). The script's docstring defines each field. It
  exits 2 when OUT_DIR lacks `state.env`, `transcript.jsonl` or
  `changes.json`.

`known-answers/` holds, for `score`, a trial that hits every fingerprint
(the body refusal, the reading-pin refusal, a Launch that writes into wt
and a re-run remove, with three `--keep` values under one directory), a
clean trial that loaded no skill and recorded its Result, a trial with no
transcript and a usage failure. Every expected output was written by hand
from the rules above.
