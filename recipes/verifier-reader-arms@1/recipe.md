---
name: verifier-reader-arms
version: 1
format: 1
subject: A fresh sonnet verifier, on a hyp binary built from the sandbox base, with its brief from the base (base) or the base plus a VERIFIER.md patch (compare, session, reader), judges a recorded select on a synthetic store with three known wrong verdicts; what it answered and confirmed is counted from its transcripts, changes and a follow-up turn (self-review)
---

# verifier-reader-arms

The instrument of the verifier reader tree's `e-trials`. One sandbox per trial.
The store is a fiction about webhook retries, built by `fixture/generate.py` and
kept as a file of this bundle under `fixture/tree`, with `fixture/ids.txt` and
`fixture/items.tsv` naming its ids. Every Experiment is decided by reading with
text evidence, so no Result carries a mark. The question has a plain select and
one candidate. Three `contradicts` are known wrong and differ in kind: in two the
configuration the run used is not the one its Hypothesis prescribes, and the first
of these alone keeps a rival from being a candidate; in the third the
configuration matches, and the failure comes from the run's own start script. A
fourth `contradicts` is the known right one. No text in the store says which is
which.

`arms/` holds the three patches, each a subject over the sandbox base given as the
trial's `--patch`, and each touching only `skills/hyp-prototype/VERIFIER.md`:
`compare` has the one verifier write what a Hypothesis prescribes, what the run
used, and whether they are the same; `session` starts one fresh reader per Result
first, and the verifier reads their reports; `reader` is `session` with the reader
also sent to each Hypothesis its verdicts name. The hyp binary is the same in every
arm.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`, `trial` and
`model`. `arm` is `base`, which takes no patch, or the ARM of the trial's one
`--patch ARM=FILE`. Launch exits 2 when a line is missing, `base` is given a
patch, another arm is given no patch, more than one or one of another name, the
patch touches anything but `skills/hyp-prototype/VERIFIER.md` or does not apply,
or the arm does not build. It applies the patch, builds the binary twice from
copies of `wt` with `CGO_ENABLED=0 go build -trimpath`, requires equal sha256, and
puts it at this recipe's tmp as `bin/hyp`, with the sha256 in `arm.sha256`. It
writes the section after `## Brief for the verifier`, up to the next heading, as
`brief.txt`, with `<Q-id>` and `verifier-<topic>-1` filled, and the section after
`## Brief for a reader`, when there is one, as `reader-brief.txt`. It copies the
inputs file to `state.txt` and `fixture/items.tsv` to `items.tsv`. Then it deletes
everything in `wt/`, `.git` included, and copies `fixture/tree` to `wt/.tree`.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the arm's `hyp` run and the binary matches
`arm.sha256`; `hyp audit --help` exits non-zero; `wt/` holds only `.tree`; the
question reads `open/awaiting_verifier` with no `uncontained` or `questionable`
mark; the events hold no telling word; `brief.txt` names the question, holds no
placeholder Launch fills, and names no other id of `fixture/ids.txt`; only
`session` and `reader` have a reader brief, it holds `<R-id>`, and it holds the
words `on each Hypothesis` for `reader` and not for `session`.

## Drive

Run: ./drive.py

The docstring states every session, limit and file. A reader or verifier turn is
`claude -p --model MODEL --session-id SID --setting-sources project
--output-format json --permission-mode dontAsk --allowedTools Bash Read Glob
Grep` in `wt/`, without `HYP_*` or `CLAUDE_CODE_*` variables, with the arm's `hyp`
first on `PATH`, `HYP_ACTOR=reader-<arm>-<trial>-<R-id>` for a reader and
`HYP_ACTOR=verifier-<arm>-<trial>` for the verifier. A reader's final message is
its report, saved without a change to `wt/reader-reports/<R-id>.md` once every
reader has ended, and to `reader-report-<R-id>.md` at this recipe's tmp. A reader
that failed or gave no answer ends the trial's sessions, and score reads the trial
as void. The follow-up resumes the verifier's session with `--tools ""` and
`followup.txt`, which names no id. A tool call in it stops the catch, because
`--tools ""` is not relied on.

## Evidence

Kept by hand at remove, each as `--keep tmp/verifier-reader-arms/FILE`:
`state.txt`, `drive.json`, `transcript.jsonl`, `followup.jsonl`, `verifier.json`,
`followup.json`, `brief.txt`, `arm.sha256`, and for `session` and `reader`
`reader-brief.txt` and each `reader-<R-id>.jsonl` and `reader-report-<R-id>.md`
that Drive wrote. The trial's events reach OUT_DIR through `changes/`. Then `score`
runs on that OUT_DIR, as a Standalone command.

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
- `verdict SCORE...` applies the pre-registered rule, stated in its docstring. It
  always accepts `base` alone, and otherwise needs all three other arms on a valid
  cut, exiting 2 on anything else.

`known-answers/` holds, for `score`: a catch in each arm; one wrong verdict missed,
per wrong verdict; a false alarm on a right verdict, on the control and on both; a
`not-opened`; no confirm; a tool in the follow-up; all three answered `wrong` with
a `sufficient` confirm; a void trial by exit, budget, a follow-up timeout, a
follow-up with no answer line, a reader failure, a reader with no answer, a reader
using an alias for the `H-` id and a report that is missing; `unclear` taken as a
reader's answer; `readers_looked_beyond` by `show` of an H-id, by alias, by
`status` and by a loop, not by a redirected `show` of its own id or an errored
call; `reports_read` by Read, a loop and a `cd`, not by `ls` or `wc` or after the
confirm; `events_by_others`; a follow-up line with the verdict echoed, with a colon,
bolded, in code, with the ids reversed and with one line missing; opening by a glob
across folders, a loop, a `cd`, Read and the readers; sizes only, another folder, an
errored Read, one file only, after the confirm, and a refused confirm re-run; a
missing file and a usage failure. For `verdict`: a `supports` and a `contradicts`
case for each label; base alone at 0, 2, 3, 4 and 5; base at the limit with the
other arms; `inconclusive` by count and by false alarms; two and three supported;
arms refused on an invalid cut; a partial set; no base; four trials; a repeated
trial; an unknown arm; a missing field; a void dropped; every wrong verdict answered
`wrong` without a catch; and a usage failure. Every expected output was written by
hand from the rules above.
