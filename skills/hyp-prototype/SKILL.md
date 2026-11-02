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
  holds goes under that hypothesis.
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
  of the assumption marks every dependent stale automatically.
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
- `hyp signals` prints the waste signals for the whole tree (`wasted`,
  `no-movement`, `one-rival`, `D1`, `spin`, `verifier-loop`), one line each;
  a `one-rival` line names the rivals the next experiment could target.
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
- When a Result shows a limit of the design that is not yet tested or
  fixed, name it on the Result with `--limit "<the limit>"` (repeatable,
  at most 300 characters), not only in the observation. It stays under the
  question in `hyp show <Q-id>` and in `hyp status` (OPEN LIMITS) until a
  later Result closes it with `--closes <R-id>#<n>` or the owner accepts it.
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
- The suggestion is mechanical: it is the only supported, non-stale candidate,
  not a recommendation. Select only a suggested candidate, and only when the
  evidence would convince an independent verifier:
  `hyp select <Q-id> <H-id> --why "..."`. Add `--tradeoff` when the choice
  gives up something the owner cares about, and `--scope-change` when it
  changes the Goal's scope; either one requires the owner's confirmation. The
  owner is also required automatically when two or more candidates were
  supported at the time of the select.
- To change a selection, select the other candidate; the new select replaces
  the old one.

## Confirm

- Every select needs an independent verifier; follow [VERIFIER.md](VERIFIER.md).
  The CLI refuses a verifier confirm from the actor that made the select.
- `Q-… open/verifier_insufficient` with a `verifier:` line under it is the
  verifier's to-do. Do it (experiments, records, retracts). Then ask for a
  fresh verdict on the same select:
  `hyp request-verify <Q-id> --why "<which Results or changes address the to-do>"`.
  The question returns to `awaiting_verifier`; start a new verifier with a
  new actor name, following VERIFIER.md. Do not re-select to reach the
  verifier: a new select re-snapshots dependencies and drops any owner
  confirm on the old one.
- Amending the selected Hypothesis drops the select's sufficient verifier
  confirm and its owner confirm, because they judged the old text. The
  question returns to `awaiting_verifier`; start a new verifier as above.
- AWAITING OWNER lists two kinds of line. Stop and ask the owner for either.
  Show them `hyp show <Q-id>` and the rivals' `hyp show` output.
  - `Q-… trade-off: H-…, H-…` means two or more candidates are supported
    and the owner must choose; it appears before any trade-off select.
    Record the owner's pick as
    `hyp select <Q-id> <H-id> --tradeoff --why "<the owner's words>"`.
    The line then leaves AWAITING OWNER and FRONTIER shows
    `trade-off decided: H-… over H-…` under the question. Start the
    verifier; it judges a trade-off select without requiring the rivals
    to be excluded.
  - `Q-… awaiting_owner: H-…` appears after the verifier confirms a select
    that needs the owner. Record their answer under the owner's own actor
    name (lowercase letters, digits and hyphens, for example `owner-dat`):
    `hyp confirm <Q-id> --role owner --verdict sufficient --why "<their words>" --actor <owner-name>`.
    Show the owner the question's `limits:` lines. A sufficient owner
    confirm is refused while an open limit from a Result that judged the
    selected Hypothesis is not accepted; pass `--accept <R-id>#<n>` for
    each such limit the owner accepts, with their words in `--why`. A limit
    marked `does not gate` came from a Result on other rivals and needs no
    `--accept`.
- A question is answered only when its selection is supported, not stale,
  confirmed by the verifier and, when required, by the owner. The Goal
  completes when it has at least one top-level question and every one
  is answered.

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
  index (ADR 0006), and it is gitignored.

A complete walkthrough is in [example.sh](example.sh).
