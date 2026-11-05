---
name: hyp-prototype-verifier
description: Independent check that a selected hypothesis is sufficiently demonstrated before its design question counts as answered.
---

# Verifier

Every `hyp select` needs a verifier confirm before the question is answered.
The verifier must be someone other than the selector: the CLI refuses a
`--role verifier` confirm from the actor that made the select.

## Parent: start the verifier

Start an independent agent session (a subagent in harnesses that have one)
with a different actor name, for example `verifier-<topic>-1`. A subagent
usually inherits the parent's `HYP_ACTOR`, so the brief tells it to pass
`--actor` explicitly on its one write. Give
it only the question id and the brief below. Do not pass your reasoning, your
conclusion or the transcript: the verifier judges the recorded evidence, not
your account of it.

A verifier must be able to read the evidence it confirms. If the harness
limits the verifier's tools, allow it to run `hyp` and to read the files
under `.tree/evidence/` and `recipes/` that the Results cite.

## Brief for the verifier

You are verifying design question `<Q-id>` in the hyp tree of this repository.

1. Run `hyp show <Q-id>`. Note the `selected:` line: the hypothesis, its kind in parentheses (none for a plain select), and the selector's why, which for a trade-off should hold the owner's words. Note the rivals among the question's children. If it prints `changed since <actor>:`, an earlier confirm judged different evidence, and the line names what the Dossier gained (`+`) and lost (`-`); read that line first, then judge the whole selection again.
   The `design:` line is the selected hypothesis's current body: judge that
   text, not an earlier one. Note any `limits:` lines and their state.
2. Run `hyp show` on the selected hypothesis, and on every Result listed
   under `dossier:` in the output of step 1, rivals' Results included. A
   Result is listed there because it has a verdict on a live hypothesis,
   wherever it sits in the tree. A line marked `decides the owner requirement`
   is a `contradicts` that alone keeps a rival from being a candidate, so check
   that verdict closely. If `hyp show` on the selected hypothesis prints
   `predecessor evidence (not counted):`, run `hyp show` on each Result
   listed there too. Those Results tested the claims this hypothesis
   supersedes; judge whether its text answers what they found.
3. Open every evidence pointer those Results list: files under
   `.tree/evidence/` or `recipes/`, or URLs.
4. Judge the selection by its kind:
   - **Plain select** (no kind shown), or `(scope-change)` alone:
     - Does the evidence demonstrate the hypothesis within the scope its
       title and the question state, not merely something nearby?
     - Were the rivals actually excluded by an experiment that targeted
       them, rather than never tested?
   - **`(trade-off)`**, alone or with `scope-change`: the owner chose
     between rivals that both meet the question, on properties they value
     differently. A supported rival is left on purpose. Do not require the
     rivals to be excluded. Judge:
     - Does the evidence demonstrate the selected hypothesis within the
       scope its title and the question state?
     - Does the evidence show that the remaining rivals really differ on
       the properties the select's why names? If they do not differ, the
       trade-off is not real; say so.
   - For each open limit, including those marked `does not gate`: is it
     outside what the question asks, or does the evidence leave a stated
     part of the question untested? A limit found on a rival can still
     leave part of the question untested. If it is the second, the verdict
     is insufficient and its why names the limit.

     A sufficient verdict hands the select to the owner confirm, which hyp
     requires after every trade-off select.
   - For each Result line hyp marks (`questionable:` or `uncontained:`):
     does the verdict rest on what the mark names? If it does, the verdict
     is insufficient, and its why names the mark.
   - For each Result whose verdict counts a recipe command's output: does
     the run's mode, the Evidence command or Standalone label the verdict
     counts, have a must-pass case and a must-fail case under the bundle's
     `known-answers/`? Are all of the cited record's `known_answers`
     outcomes `ok`? If either is no, the verdict is insufficient, and its
     why names the missing case or the failed one.
   - When that command is a verdict check (its cases print
     `verdict: LABEL=WORD` lines): does each label the verdict counts have
     a case printing `supports` and a case printing `contradicts`, and does
     `hyp recipe check` print no verdict warning for it? If not, the
     verdict is insufficient, and its why names the missing case.
5. Record your judgement, once, under your own actor name:
   `hyp confirm <Q-id> --role verifier --verdict sufficient --actor verifier-<topic>-1 --why "<reason, at most 300 characters>"`
   or the same with `--verdict insufficient`, where the `--why` states what evidence would make the selection sufficient. For a trade-off select, never ask for a rival to be excluded; ask only for evidence on the two questions above. The next agent reads it as its to-do list; `hyp status` shows it under the question as `verifier_insufficient`. When the to-do is done, that agent starts a new verifier.

Do not change the tree in any other way: no records, selects, links or
retracts. If something is wrong, say so in the `--why` of an insufficient
verdict and let the investigating session act on it.
