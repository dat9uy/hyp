---
name: verifier-brief-arms
version: 1
format: 1
subject: A fresh sonnet verifier, given only the brief in skills/hyp-prototype/VERIFIER.md (base, or a variant given as the trial's --patch), judges a recorded select at a store cut; whether it opened and caught a wrong contradicts is counted from its transcript and a follow-up turn, and the skill-text gate is applied (self-review)
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
`skills/hyp-prototype/VERIFIER.md`. `arms/` holds the two variants, `judge`
and `open-all`, against that file. Launch exits 2 when a line is missing,
`base` is given a patch, another arm is given no patch, more than one or one
of another name, the patch touches another file or does not apply, or
`wt/.tree` is missing.

It applies the patch, then writes the section after `## Brief for the
verifier` to this recipe's tmp as `brief.txt`, with `<Q-id>` replaced by
`Q-KW0PWZ` and `verifier-<topic>-1` by `verifier-<arm>-<trial>`: a real
parent gives only this. It writes the base commit's section the same way as
`base-brief.txt`. Then it deletes everything in `wt/` except `.tree`, `.git`
included, because `docs/adr/0011` names the Result under test and git history
can show that file. This changes the clone, not the cut, and `hyp sandbox
remove` diffs against the manifest, not against git.

## Doctor

Run: ./doctor.sh

Passes when `claude`, `python3` and the pinned `hyp` run; `wt/` holds only
`.tree`; `hyp events` ends at the cut; `hyp show Q-KW0PWZ` reads
`open/awaiting_verifier`; the sha256 of `E-171ACT/e3.rb`, `E-171ACT/run.log`,
`E-X94VRM/e6.rb` and `E-X94VRM/run.log` are the cited evidence's; `brief.txt`
is there and holds neither placeholder nor the Result's id; no file outside
`.tree` names the Result; and `brief.txt` equals the base commit's section for
`base` and differs from it for a variant.

## Drive

Run: ./drive.py

The verifier turn is `claude -p --model MODEL --session-id SID
--setting-sources project --output-format json --permission-mode dontAsk
--allowedTools Bash Read Glob Grep --max-budget-usd 3` in `wt/` with
`brief.txt` on stdin, without `HYP_*` or `CLAUDE_CODE_*` variables and with
`HYP_ACTOR=verifier-<arm>-<trial>`, stopped after 25 minutes. The follow-up
is `claude -p --model MODEL --resume SID --setting-sources project --tools ""
--output-format json` with `followup.txt` on stdin, stopped after 5 minutes.
`followup.txt` names no id. It runs only when the verifier's turn ended well.

A resume may still show a tool call in its transcript, so `--tools ""` is not
relied on: any tool call in the follow-up fails the trial.

A resume may append to the session's file or write a new one. Drive chooses
the file from the session id the follow-up's JSON result reports, and keeps
the records the verifier's turn does not hold as `followup.jsonl`. The
docstring lists what it writes. It exits 1 when no transcript was written.

## Evidence

Kept by hand at remove, each as `--keep tmp/verifier-brief-arms/FILE`:
`state.env`, `drive.json`, `transcript.jsonl`, `followup.jsonl`,
`verifier.json`, `followup.json` and `brief.txt`. The verifier's confirm
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
  `opened`, `answer_target`, `answer_control`, `followup_tools` and `failed`.
  A trial passes (`failed` 0) only when it confirmed, opened both of the
  Result's files before the confirm, answers `wrong` on `R-KA7Y6Z` and `holds`
  on `R-094XDZ`, and used no tool in the follow-up. `void` reads only exit
  state, never the answers. The script's docstring defines each field. It
  exits 2 on a missing file.
- `verdict SCORE...` drops void trials and applies the gate. `base` fails at
  3 or more of 5 to reproduce, and a variant passes at 1 or fewer of 5. It
  prints a line per arm (k, opened count, control-flagged count), `verdict:`
  lines for `h-base-enough`, `h-judge`, `h-open-all` and `h-neither`, and a
  `ships:` line. It accepts `base` alone only when base does not reproduce,
  and otherwise needs both variants; it exits 2 on anything else. The script's
  docstring states the rule.

`known-answers/` holds, for `score`, a clean pass; a trial that only sized the
files with `ls`, `wc` and `stat`; one that opened only another folder's
`run.log`; one that flags the control; one that misses the target; files
opened after the confirm; no confirm; a tool in the follow-up; a void trial by
exit and by budget; files opened through a `cd`, a loop, a glob and a brace; an
errored Read; a refused confirm re-run after reading; no answer lines; a
missing file and a usage failure. For `verdict`, it holds a `supports` and a
`contradicts` case for each label, both variants passing with fewer failures
and tied, early stop and each refused set of arms, a void trial dropped, four
trials, a repeated trial and a missing field. Every expected output was written
by hand from the rules above.
