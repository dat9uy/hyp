---
name: verifier-audit-arms
version: 1
format: 1
subject: A fresh sonnet verifier, on a hyp binary and brief built from the sandbox base (base) or the base plus a recorded-audit patch (audit, readers), judges a recorded select on a synthetic store with two known wrong verdicts; what it opened, audited and caught is counted from its transcripts, changes and a follow-up turn (self-review)
---

# verifier-audit-arms

The instrument of the verifier audit tree's `e-trials`. One sandbox per trial.
The store is a fiction about webhook retries, built by `fixture/generate.py`
and kept as a file of this bundle under `fixture/tree`, with `fixture/ids.env`
and `fixture/items.tsv` naming its ids. Every Experiment is decided by reading
with text evidence, so no Result carries a mark. The question has a plain select
and one candidate. Two `contradicts` are known wrong and differ in kind; the
first alone keeps a rival from being a candidate. A third `contradicts` is the
known right one. No text in the store says which is which.

`arms/` holds the two patches, each a subject over the sandbox base given as the
trial's `--patch`: `audit` adds `hyp audit`, the coverage rule and a brief where
one verifier audits every verdict; `readers` is the same binary with a brief
where a fresh reader per Result audits under the verifier's name first.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`, `trial` and
`model`. `arm` is `base`, which takes no patch, or the ARM of the trial's one
`--patch ARM=FILE`. Launch exits 2 when a line is missing, `base` is given a
patch, another arm is given no patch, more than one or one of another name, the
patch does not apply, or the arm does not build. It applies the patch, builds
the binary twice from copies of `wt` with `CGO_ENABLED=0 go build -trimpath`,
requires equal sha256, and puts it at this recipe's tmp as `bin/hyp`, with the
sha256 in `arm.sha256`. It writes the section after `## Brief for the verifier`
as `brief.txt`, with `<Q-id>` and `verifier-<topic>-1` filled, and the section
after `## Brief for a reader`, when there is one, as `reader-brief.txt`. Then it
deletes everything in `wt/`, `.git` included, and copies `fixture/tree` to
`wt/.tree`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the arm's `hyp` run and the binary matches
`arm.sha256`; `wt/` holds only `.tree`; the question reads
`open/awaiting_verifier` with no `uncontained` or `questionable` mark; the
events hold no telling word; `brief.txt` names the question, holds no placeholder
Launch fills, and names no other id of `fixture/ids.env`; `hyp audit --help`
exits 0 for a patched arm and non-zero for `base`; and only `readers` has a
reader brief.

## Drive

Run: ./drive.py

The docstring states every session, limit and file. A reader or verifier turn is
`claude -p --model MODEL --session-id SID --setting-sources project
--output-format json --permission-mode dontAsk --allowedTools Bash Read Glob
Grep` in `wt/`, without `HYP_*` or `CLAUDE_CODE_*` variables, with the arm's
`hyp` first on `PATH` and `HYP_ACTOR=verifier-<arm>-<trial>`. The follow-up
resumes the verifier's session with `--tools ""` and `followup.txt`, which names
no id. A tool call in it fails the trial, because `--tools ""` is not relied on.

## Evidence

Kept by hand at remove, each as `--keep tmp/verifier-audit-arms/FILE`:
`state.env`, `drive.json`, `transcript.jsonl`, `followup.jsonl`, `verifier.json`,
`followup.json`, `brief.txt`, `arm.sha256`, and for `readers` `reader-brief.txt`
and each `reader-<R-id>.jsonl`. The trial's events reach OUT_DIR through
`changes/`. Then `score` runs on that OUT_DIR, as a Standalone command.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for every turn to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the subject), read only; writes nothing

Run verdict: ./verdict.py
External: none; reads only the score files given as arguments

- `score OUT_DIR` prints `key<TAB>value` lines for one trial. Its docstring
  defines each field and what opens a file. It exits 2 on a usage error or a
  missing file.
- `verdict SCORE...` applies the pre-registered rule, stated in its docstring.
  It always accepts `base` alone, and otherwise needs both audit arms on a valid
  cut, exiting 2 on anything else.

`known-answers/` holds, for `score`: a clean catch; a catch by readers' opening;
files sized by `wc`, `ls` and `stat`; one wrong verdict caught; a flagged control;
files opened after the confirm; no confirm; a tool in the follow-up; a void trial
by exit and by budget; a `holds` audit on a wrong verdict with a `wrong` follow-up,
and one later changed; a `wrong` audit on the control; a verifier that audited in
the readers arm; a refused confirm re-run after reading; opening by a glob across
folders, a loop, a `cd` and Read; a `cd` into the evidence directory with a glob, a
loop, sizes only and another folder; an errored Read; a file of the same name in
another folder; no answer lines; a missing file and a usage failure. For
`verdict`: a `supports` and a `contradicts` case for each label; base alone on a
valid and an invalid cut; one audit arm refused; audit arms refused on an invalid
cut; both audit arms supported; `inconclusive` by count and by flagged control;
abandon by contradiction and by `holds`; four trials; a missing field; a void
dropped; no base and a usage failure. Every expected output was written by hand
from the rules above.
