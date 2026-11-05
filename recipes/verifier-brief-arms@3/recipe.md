---
name: verifier-brief-arms
version: 3
from: verifier-brief-arms@2 5f877f36258b8b2b00fbbd4ad04528ad6460de5d8cc00434098933b1854b587f
format: 1
subject: A fresh sonnet verifier, given only the brief in skills/hyp-prototype/VERIFIER.md (base, or a variant given as the trial's --patch, which starts a fresh reader per Result first), judges a recorded select at a store cut; whether it opened and caught a wrong contradicts is counted from its transcript and a follow-up turn, and the skill-text gate is applied (self-review)
---

# verifier-brief-arms

The skill-text gate's instrument for the verifier's brief. One sandbox per
trial. The store is given to `hyp sandbox create --store` with `--cut
01M5Q03CGKKXTZRYKYZW4PYKPJ`, and is never kept in this bundle: its evidence
is application code, so the store, the transcripts and the tree that pins
this recipe stay private. Known answers are synthetic. The cut is the select
of `Q-KW0PWZ` just before its verifier ran. The Result under test is
`R-KA7Y6Z`, whose `contradicts` two independent readers judged wrong; the
control is `R-094XDZ`, a `contradicts` in the same question that one reader
of the same audit judged to hold.

## Launch

Run: ./launch.sh

Reads the inputs file (`--inputs FILE`), `KEY=VALUE` lines: `arm`, `trial`
and `model`. `arm` is `base`, which takes no patch, or the ARM of the trial's
one `--patch ARM=FILE`, whose patch touches only
`skills/hyp-prototype/VERIFIER.md`. `arms/` holds the one variant, `reader`,
the same bytes as `verifier-reader-arms@1`'s, against that file. Launch exits 2 when a line is missing,
`base` is given a patch, another arm is given no patch, more than one or one
of another name, the patch touches another file or does not apply, or
`wt/.tree` is missing.

It applies the patch, then writes the section after `## Brief for the
verifier`, up to the next heading, to this recipe's tmp as `brief.txt`, with `<Q-id>` replaced by
`Q-KW0PWZ` and `verifier-<topic>-1` by `verifier-<arm>-<trial>`: a real
parent gives only this. It writes the base commit's section the same way as
`base-brief.txt`, and the section after `## Brief for a reader`, when there is
one, as `reader-brief.txt`, with `items.tsv` listing each R-id and H-id under
the cut's `dossier:`. Then it deletes everything in `wt/` except `.tree`, `.git`
included, because `docs/adr/0011` names the Result under test and git history
can show that file. This changes the clone, not the cut, and `hyp sandbox
remove` diffs against the manifest, not against git.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run; `wt/` holds only
`.tree`; `hyp events` ends at the cut; `hyp show Q-KW0PWZ` reads
`open/awaiting_verifier`; the sha256 of `E-171ACT/e3.rb`, `E-171ACT/run.log`,
`E-X94VRM/e6.rb` and `E-X94VRM/run.log` are the cited evidence's; `brief.txt`
is there and holds neither placeholder, the Result's id, nor a line starting
`## `; for `reader`, `reader-brief.txt` holds `<R-id>` and `on each Hypothesis` and no
heading, and `items.tsv` holds the Dossier's 14 verdicts on 8 Results; for `base`,
neither file is there; no file outside
`.tree` names the Result; and `brief.txt` equals the base commit's section for
`base` and differs from it for a variant.

## Drive

Run: ./drive.py

For `reader`, one reader per Result in `items.tsv` runs first, as `claude -p
--model MODEL --session-id SID --setting-sources project --output-format json
--permission-mode dontAsk --allowedTools Bash Read Glob Grep --max-budget-usd 0.25`
in `wt/` with `HYP_ACTOR=reader-<arm>-<trial>-<R-id>`, stopped after 5 minutes,
and none is started that could end after minute 30. Its final message is its
report, saved without a change as `reader-report-<R-id>.md` at this recipe's tmp,
and to `wt/reader-reports/<R-id>.md` once every reader has ended. A reader that
failed or gave no answer ends the trial's sessions, and score reads it as void.

The verifier turn is the same command with `--max-budget-usd 1`, with
`brief.txt` on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=verifier-<arm>-<trial>`, stopped after 20 minutes. The follow-up
is `claude -p --model MODEL --resume SID --setting-sources project --tools ""
--output-format json --max-budget-usd 0.50` with `followup.txt` on stdin, stopped
after 5 minutes. These budgets bound a trial at $1.50 for `base` and $3.50 for
`reader`, because the run's spending cap leaves little room.
`followup.txt` names no id. It runs only when the verifier's turn ended well.

A resume may still show a tool call in its transcript, so `--tools ""` is not
relied on: any tool call in the follow-up fails the trial.

A resume may append to the session's file or write a new one. Drive chooses
the file from the session id the follow-up's JSON result reports, and keeps
the records the verifier's turn does not hold as `followup.jsonl`. The
docstring lists what it writes. It exits 1 when the verifier ran and no transcript
was written.

## Evidence

Kept by hand at remove, each as `--keep tmp/verifier-brief-arms/FILE`:
`state.txt`, `drive.json`, `transcript.jsonl`, `followup.jsonl`,
`verifier.json`, `followup.json` and `brief.txt`, and for `reader`
`reader-brief.txt`, `items.tsv` and each `reader-<R-id>.jsonl` and
`reader-report-<R-id>.md` that Drive wrote. The verifier's confirm
reaches OUT_DIR through `changes/`. Then `score` runs on that OUT_DIR, as a
Standalone command.

## Cleanup

Run: ./cleanup.sh

Nothing is left running: Drive waits for both turns to end.

## Standalone

Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the subject), read only; writes nothing

Run verdict: ./verdict.py
External: none; reads only the score files given as arguments

- `score OUT_DIR` prints `key<TAB>value` lines for one trial: `arm`, `trial`,
  `void`, `confirmed`, `confirm_verdict`, `opened_run_log`, `opened_script`,
  `opened`, `answer_target`, `answer_control`, `followup_tools`, `readers_failed`,
  `readers_no_answer`, `readers_looked_beyond`, `reports_read` and `failed`.
  A trial passes (`failed` 0) only when it confirmed, opened both of the
  Result's files before the confirm or through a reader, answers `wrong` on `R-KA7Y6Z` and `holds`
  on `R-094XDZ`, and used no tool in the follow-up. `void` reads exit state and whether a reader failed or gave no answer, never the
  follow-up's answers. The script's docstring defines each field. It
  exits 2 on a missing file.
- `verdict SCORE...` drops void trials and applies the gate. `base` fails at
  3 or more of 5 to reproduce, and `reader` passes at 1 or fewer of 5. It
  prints a line per arm (k, opened count, control-flagged count), a
  `reproduced:` line, `verdict:` lines for `h-null` and `h-reader`, and a
  `ships:` line. It accepts `base` alone only when base does not reproduce, and
  otherwise needs `reader`; it exits 2 on anything else. The script's docstring
  states the rule.

`known-answers/` holds, for `score`, a clean pass; a trial that only sized the
files with `ls`, `wc` and `stat`; one that opened only another folder's
`run.log`; one that flags the control; one that misses the target; files
opened after the confirm; no confirm; a tool in the follow-up; a void trial by
exit and by budget; files opened through a `cd`, a loop, a glob and a brace; an
errored Read; a refused confirm re-run after reading; no answer lines; a
missing file and a usage failure; and for readers, a pass through a reader, a
reader of another Result, one file only, an errored read, an `unclear` report, a
missed target, a flagged control, `readers_looked_beyond`, `reports_read` in part,
by a loop and by listing only, and a void trial by reader failure, no answer, an
alias, a partial answer and a missing report. For `verdict`, it holds a `supports` and a
`contradicts` case for each label, and an `inconclusive` for `h-reader`, base
alone refused when it reproduces, the variant refused after an early stop, no
base, an unknown arm, a void trial dropped, a void trial leaving four, four
trials, a repeated trial, a missing field and a usage failure. Every expected output was written
by hand from the rules above.
