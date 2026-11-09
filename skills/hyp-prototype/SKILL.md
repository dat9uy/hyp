---
name: hyp-prototype
description: Investigate an uncertain design with rival hypotheses and bounded experiments, keeping the whole investigation in a hyp tree (.tree/) instead of prose. Use for design questions, prototypes, and resuming an investigation another session started.
---

# hyp-prototype

The `.tree/` store is the only record of the investigation. Every status
(supported, stale, answered, …) is derived by `hyp` from recorded events; you
never type one. Run `hyp help` for every command and flag.

## Start or resume

1. Run `hyp status` first, every session. It fits in 60 lines and shows the
   Goal, the open questions (FRONTIER), stale chains, what awaits the owner,
   recent events and the last checkpoint. Use `hyp show <id>` for detail and
   `hyp diff --since <id|RFC3339>` for what changed.
2. If it says `no .tree store found`, run `hyp init` at the repository root,
   then `hyp add goal --title "<observable outcome>" --why "<who asked, why>"`.
3. Set `HYP_ACTOR` to a name unique to this session, for example
   `export HYP_ACTOR=claude-<topic>-1`. Every write needs it and a `--why`.

## Frame

- One `hyp add dq --parent <goal-or-hyp> --title "<question>" --why "..."` per
  material design question. A question that only matters if a hypothesis
  holds goes under that hypothesis. It blocks: its parent cannot become the
  answer until the sub-question is settled, whichever way it is settled.
- Rival answers are siblings: `hyp add hyp --parent <Q-id> --title "..." --why "..."`.
  Add at least two when the answer is genuinely uncertain.
- New information goes in a new node. A fix, an untested part of a claim
  or a stronger instrument is a new Hypothesis that supersedes the old one,
  and it exists before the Experiment that tests it. Run, in this order:
  `hyp add hyp --parent <Q-id> --title "..." --why "..."`, then
  `hyp supersede <new H-id> --old <old H-id> --why "..."`, then
  `hyp add exp --parent <new H-id> ...`. A Hypothesis that is already
  supported can only move by being contradicted, so a fix tested on it
  moves nothing and the round reads as a stall.
- Never edit a node's events. `hyp amend <H-id> --body "<the whole revised body>" --why "<which Results the new wording matches>"`
  only makes a body say exactly what its recorded Results tested (clearer
  wording, or a narrower claim). It never adds a claim that still needs a test.
- When one answer assumes another, say so:
  `hyp link <H-or-Q-id> depends_on <H-id> --why "..."`. Then a contradiction
  of the assumption marks every dependent stale automatically. This link, not
  a sub-question, is how to say a claim rests on an answer.
- Link issues or URLs with `hyp link <id> external_ref gh:976 --why "..."`.

## Experiment

- Choose the experiment whose outcome would exclude a rival (strong
  inference), not one that can only confirm your favourite. Do not test a
  contradicted Hypothesis again, as parent or target; if its contradiction
  was wrong, `hyp retract` that Result first. An Experiment under a
  supported Hypothesis can only move it by contradicting it.
- Run the behaviour. An experiment that passes by its own construction is not
  a test: hashing a set of files and then checking that the hash changes when
  those files change proves nothing about outputs.
- `hyp add exp --parent <H-id> --title "<what you will observe>" --recipe <name@version|none> --target <sibling H-id> --why "..."`.
  List every sibling the experiment also tests with `--target` (repeatable);
  the parent hypothesis is tested implicitly, so do not list it. When the
  question has live rivals and the experiment targets none of them, hyp
  refuses and lists them: name the rival the result would count against, or
  pass `--target none` to proceed without one.
- `hyp signals` prints the design marks for the whole tree (`one-rival`,
  `D1`, `unjudged`), one line each; a `one-rival` line names the rivals the
  next experiment could target. Progress is what OWED in `hyp status` shows.
- Ids look like `Q-GRM5AG` (the first word a write prints). Give a node an
  easier name with `--alias` when you add it, and use the alias anywhere.
- Use a recipe from `recipes/` when one fits (`name@version` from its
  frontmatter). A verdict a measure decides needs a recipe that holds it
  (RECIPE-FORMAT.md, Verdict scripts). `--recipe none` with
  `--decided-by reading` is for a verdict decided by reading. Run the
  experiment with the hyp-verification skill.

## Record

- Record every result, including inconclusive ones, with one verdict per
  targeted hypothesis:
  `hyp record --exp <E-id> --verdict <H-id>=supports --verdict <H2-id>=contradicts --observation "<what happened>" --evidence .tree/evidence/<E-id>/run.log --why "..."`.
- A limit is a case the Goal covers that no check ran. When a Result
  leaves such a case, name it on the Result with `--limit "<the case>"`
  (repeatable, at most 300 characters), not only in the observation. Close
  it with an Experiment that aims at it (`--aims <R-id>#<n>`) and runs the
  case on every candidate; its Result names `--closes <R-id>#<n>`, and its
  verdicts separate the rivals. Write one case per `--limit`: "not X or Y"
  is two limits. A close covers every case the limit's text names, and
  `hyp record` prints that text when it closes it. If the Result did not
  run some of them, record each one it did not run as a `--limit` on the
  same Result, and refer that new key when its case is void for every
  candidate or outside the Goal's lines. A claim that a case is void, out
  of scope or covered elsewhere is a limit and a refer, never a sentence in
  an Experiment body or an observation. The scope of a run goes in the Experiment
  body, a procedure note or a commit goes in the observation, and a premise
  is a `depends_on` edge; none of these is a limit.
- A weakness you observe against a hypothesis belongs in its verdict
  (`contradicts` or `inconclusive`), not only in the observation text; status
  is derived from verdicts and never reads prose.
- A failed pre-registered pass rule is `contradicts`, not `inconclusive`.
  `inconclusive` means the experiment ran as designed and still could not
  decide the claim, for example a count inside a pre-registered inconclusive
  band. Its next round must change the instrument or the power, or add a rival.
- `--why` holds at most 300 characters and `--observation` at most 600. Put
  anything longer (logs, reasoning, screenshots) in files under
  `.tree/evidence/<E-id>/` and point to them with `--evidence` (repeatable).
- A wrong record is undone with `hyp retract <R-id> --why "..."`, never by
  editing files.

## Revise

After each record, run `hyp status` again. For each line under STALE, recheck
the dependent before relying on it:

- run a new experiment on it; or
- re-point the dependency: `hyp unlink <id> depends_on <old> --why "..."`,
  `hyp link <id> depends_on <new> --why "..."`, then
  `hyp reaffirm <id> --why "<the new reason it still holds>"`; or
- give it up: `hyp abandon <H-id> --why "..."`.

Never reaffirm without a new reason. A reaffirm cannot clear staleness while
the assumption itself is unsupported; `hyp check` warns when one had no effect.
A question you must park is `hyp defer <Q-id>` (undo with `hyp resume`); one
that no longer matters is `hyp withdraw <Q-id>`.

## Select

- `hyp select <Q-id>` with no hypothesis writes nothing and prints the tool's
  suggestion: one candidate, `trade-off: owner decides between …`, or
  `suggested: none (<reason>)`.
- Select only when OWED lists `select` for that Hypothesis:
  `hyp select <Q-id> <H-id> --why "..."`. hyp refuses a select that the tree
  does not owe and names what comes first: a runner item on the question, an
  open reading, or a premise in a question that is still open. Decide that
  question first.
- A gap you know at select time is a `--limit` on a Result, never only words
  in `--why`.
- An open limit under the question holds the select, whichever Hypothesis
  its Result judged, and `hyp select` names it. Run its case on the
  candidates. If no test within the Goal's lines can close it, or its case
  is void for every candidate, refer it:
  `hyp checkpoint --refers <R-id>#<n> --summary "<why no test can close it>"`.
  The owner then accepts it.
- The select answers the question while it stands: the selected Hypothesis is
  supported and not stale, it has no open sub-question, and no candidate is
  new since the select. A new candidate reopens the question
  (`new_candidate`); OWED then lists `select` or `choose` again.
- Add `--scope-change` when the select changes the Goal's scope. Only then
  does the owner confirm: `Q-… awaiting_owner: H-… (scope change)` waits
  for `hyp confirm <Q-id> --role owner --verdict sufficient --why "<their words>" --actor <owner-name>`,
  with `--accept <R-id>#<n>` for each open limit they accept.
- To change a selection, select the other candidate when OWED lists it. When
  the evidence changes the claim, add a new Hypothesis that `supersedes` the
  old one and select that.
- AWAITING OWNER lists the lines that stop the work. Stop and ask the owner.
  Show them `hyp show <Q-id>` and the rivals' `hyp show` output.
  - `Q-… choose: H-…, H-…` means two or more candidates are supported and the
    owner must choose. It appears only when OWED lists nothing for the runner
    on that question and no reading on it is open, because either may change
    the candidates. Record the owner's choice as
    `hyp select <Q-id> <H-id> --why "<the owner's words>"`. That select
    answers the question; no confirm follows.
  - `R-…#r1 reading: …` is a reading of the Goal that a verdict rests on. It
    changes no status, but it holds back new rivals and the owner's choice on
    its question until the owner answers. Carry on with other work and ask
    the owner at the next stop. Record it with `hyp record --reading "<the
    line>: <how you read it>"`. Record the owner's answer under their own
    actor name (lowercase letters, digits and hyphens, for example
    `owner-dat`):
    `hyp checkpoint --role owner --closes <R-id>#r1 --summary "<their words>" --why "<why>" --actor <owner-name>`.
    If they give new words for a reading they already answered, use
    `--revises <R-id>#r1` instead.
- The Goal completes when it has at least one top-level question and every
  one is answered.

## End of session

1. `hyp checkpoint --summary "<what is next, at most 600 characters>" --why "end of session"`.
2. `hyp check` must print `check: 0 errors`. Fix any rejected event by
   recording a correct one; never edit the log.
3. Commit `.tree/events/` and `.tree/evidence/` with the code they describe.

## Never

- Type a status anywhere, in a flag, a title or a commit message.
- Keep a parallel Markdown account, progress log or status table; the tree
  and its evidence files are the account.
- Edit or delete anything under `.tree/events/`.
- Commit `.tree/index.sqlite`; older hyp binaries write it, hyp keeps no
  index, and it is gitignored.

A complete walkthrough is in [example.sh](example.sh).
