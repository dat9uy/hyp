---
name: hyp-grill
description: "Grill the owner of a hyp v2 Goal into its words, lines and examples, and re-grill a Goal that stopped or was selected. Use when a real event points at a change and the owner wants it held as a hyp Goal, when hyp stops a Goal for the owner, or when a selected Goal is reviewed. Not for building or running a Goal's lines on commits: the runner does that, with the run rules this skill ships."
---

# hyp grill

This skill turns an owner's problem into a hyp Goal that a runner can work
without a prompt written by hand: the Goal's words, its gap and guard lines, and
examples placed on each line that fail or pass on H0 as their line needs. It
also covers the re-grill, which reviews a Goal after hyp stops it or after a
select, and ships the short list of run rules a runner follows that hyp cannot
check. The quality bar: every decision of the owner is held by the words, a
line or an example, and every decision that no example checks is reported.

## When to use

- The owner has a real event (a refusal, a warning, a stop, a flagged run) and
  wants the change it points at held as a Goal.
- hyp stopped a Goal for the owner (`Two candidates` or `Budget used`), a
  selected Goal is to be reviewed, or a runner reports a placed example that
  no build can meet: that is the re-grill, in [re-grill.md](re-grill.md).
- A runner needs the run rules: give it [run-rules.md](run-rules.md), the run
  part of this skill, not a second skill.
- Not for: writing or running the build itself. The runner does that.

## What hyp checks, and what this skill holds

hyp holds the state machine. `hyp help` lists every command, and
`hyp status G --role grill` prints the acts the grill may take now, as ready
command lines with their keys filled in. hyp's rule is: "An agent runs a line
hyp printed. A skill names no write command." So this skill shows only reads
(`hyp help`, `hyp status`, `hyp try`, `hyp show`); every act that writes, you
copy from a line hyp printed. Set `HYP_ACTOR` to your actor name and
`HYP_ROLE=grill` for every command. The owner records the Goal itself, with the
act `hyp help` lists for adding a Goal; you draft the words and lines it takes,
and the draft of examples it carries with `--draft NAME`. Git holds only
the owner's acts: you write no event, and the owner adds a new example to an
open Goal with the place act.

In the owner's draft, set `HYP_ACTOR` to the owner's actor, never a
placeholder: the actor of the `goal.add` events in the store, which the folder
`.tree/events/<actor>/` names (in an older store, `.tree/events/<actor>.jsonl`).
Ask the owner once, only when the store has no Goal yet or several actors
added Goals. The owner's place, land, unselect and close of a Goal are allowed
only to the actor that added it, so a wrong actor locks the owner out of them.

## Decisions

A decision of the owner is held in one of three places: the Goal's words, a
line, or an example. Nothing else carries it to the runner.

A decision that no example checks is a gap: the grill reports it under
"Not covered" in the hand-off, with where the decision came from, so the owner
sees what the examples cannot catch. It is never passed to the runner as an
extra rule.

## The issue holds the decisions

A chat, a local page and a memory note are read by no one else, so what the
grill learns goes to the issue you grill, in the same session, with
`gh issue comment`. When the Goal has no issue, the record PR's body holds it.
The issue is the owner's record, not a rule for the runner:

- **Each decision of the owner goes to the issue you grill**, with the owner's
  words quoted and kept apart from your proposals. Say which part of the issue
  a decision replaces. Your inference is a proposal: never write it as the
  owner's words, in the issue or anywhere else.
- **A "Not covered" item that another open issue owns** (it waits on that
  issue's change, or that issue decides it) goes to that issue too, with the
  owner's words, and the hand-off names that issue.
- **The issue closes when each want it tracks is held by an added Goal.** An
  issue tracks the process toward a want: the event, the owner's decisions,
  the split into Goals and where each stands. A Goal is the snapshot of one
  want. When the want splits into several Goals, the issue stays open until
  the last one is added: the record PR of each earlier Goal says "Part of #N",
  and the record PR of the last says "Closes #N". Decisions made after an add
  still go to that issue. A want whose lines are already known needs no issue.
- **A record cites only what others can read**: an issue, a PR or a commit.
  Never a gitignored or local file, such as a design page, a report or a
  memory note, in the Goal's words, a `--why` or a comment.

## A rethink page becomes an issue

A rethink page, which the hyp-rethinking skill draws, is scratch. It changes
while the owner talks, and it is a local file, so no record cites it. Its
items are where a Goal's lines come from, and every item counts, a Proposed
item too. Do not expect an issue to exist: when the owner wants a page's items
built, the grill opens the issue from the page itself, and the issue then
holds each item until a record holds it.

1. Check the page's data file with hyp-rethinking's `check_items.py` until it
   exits 0.
2. From this skill's folder, run
   `python3 scripts/draw_issue.py PAGE.json > body.md`. It draws the body: the
   page's sections that hold items, in page order, each item with its id, tag
   and text, and a holder table, `| Item | Tag | Holder |`, with one row for
   each item and its holder empty. It draws nothing while the data has a
   defect, and exits 1.
3. The grill opens the issue with that body:
   `gh issue create --title "PAGE TITLE" --body-file body.md`. When the page
   changes, draw the body again, carry over each holder already written, and
   replace the issue's body with `gh issue edit N --body-file body.md`.

Write each holder as the owner decides. Each tag has its kind of holder:

- **Owner said** and **Proposed**: a line of an open or written Goal,
  `G-XXXXXX line-N`.
- **Fact**: a finding, `F-XXXXXX`, when the page wants the fact gone;
  otherwise the Goal line whose why holds the fact as evidence.
- **Open**: the owner's answer, `owner: "WORDS"`.
- **Hypothetical**: an example drafted for a line, `G-XXXXXX line-N/NAME`, or
  `draft NAME line-N/NAME` before the Goal is added.
- **Any tag**: the owner's close of the item, `close: "WORDS"`, in the
  owner's words.

An item too vague for a sharp line is held by a line of a written Goal that
holds the item's words, and the grill sharpens that line with the owner
before the Goal opens. A finding never holds a vague item: a finding is a
fact, and a vague item is a want.

Run `python3 scripts/check_holders.py body.md` on the issue's current body
(`gh issue view N --json body -q .body > body.md`) at every split of the
items into Goals and before any report of done, in the clone that holds the
store. It lists each item with no holder (`holder-missing`), a holder whose
kind does not fit its tag (`holder-kind`), a Goal or line the store does not
hold (`holder-unknown`), and a Goal that ended with no select
(`holder-ended`), and exits 1. A report of done waits until it exits 0. The
talk needs no end: the items stay in the table, so the check reads them at
any time.

At the end of each Goal, the land and each remainder go to the tracking issue
with their records: the land with its commit, and each remainder, a part the
Goal left out or fixed only in part, with the finding or the written Goal that
holds it. Write the new holder into the table and run the check again. With no
tracking issue, each remainder is a finding only: record it with
`hyp finding add`, and open no issue for it.

## How to grill

When the owner already recorded the Goal (`hyp status G` shows it `written`
with no examples), its recording is the confirmation step 1 asks for: skip that
and steps 2 to 8. Read `hyp status G`, with its why and each finding it
names, and the event the why names. Make the worktree of step 9 on branch
`record/<G>`, then go on with step 10 and the examples.

1. **Read the event, the store and what the issue touches, then confirm the
   problem once.** Read the issue or event that points at the change,
   `hyp help`, `hyp status`, and the spec sections and code the issue touches.
   When the Goal you draft will cite a landed Goal with `--from`, read
   `hyp status` of that Goal: it prints what that Goal decided, its words,
   lines, examples and their verdicts. Read it there, not from the store's
   files. Then, before you draft words or lines:
   - Name each case the issue leaves open and ask it as a product question.
     Ask only the open cases, not every case you can think of.
   - Show the problem restated in the owner's terms with no mechanism in it,
     what is out of scope (name the sibling issues that own it), and a preview
     of what the owner would do and see when the change is done, marked
     "not built".
   - Ask once whether this matches. Draft nothing the owner has not confirmed.
2. **Ask product questions only.** Ask what the owner wants to be true, never
   how to build it: the file layout, the command shape and the names inside the
   build are the runner's to choose, and the owner vetoes them at merge.
3. **Write each option for the owner's choice.** Each option has a short label
   and a description. The description says what the option would make the
   runner build, and the cheapest build that meets it, so the owner sees the
   price of each answer. Check each description as you check a line (step 8)
   before you offer it: the owner picks by the label, so a loose rule in the
   description reaches the line unread.
4. **Record the owner's answer as given.** The label the owner picked is the
   recorded text: keep its exact words, never the option's description or your
   paraphrase of it.
5. **A correction is not consent.** When the owner corrects a draft of the
   words or lines, apply the correction and show the result again. The owner
   records the Goal only after agreeing to the draft. This step is about the
   owner's acts; the examples of a recorded Goal follow "How to grill the
   examples".
6. **Propose missing names.** When the Goal needs a name the owner did not give
   (an example name, a folder, a file), propose one and let the owner veto it;
   do not turn it into a question.
7. **A form question names every case.** When you ask about the form of what
   hyp prints (the state block, a report, an error), name every state a Goal
   can be in (`written`, `open`, `selected`, `ended`) and every command that
   prints that form, so the answer covers each one.
8. **Draft the words and lines.** One sentence of words. Each line is `gap`
   (false on H0, the build must make it true) or `guard` (true on H0, the build
   must keep it true). A Goal needs at least one gap line. The owner records it.
   When a line makes something wait for new evidence, or counts acts, check
   whether repeating an earlier act on the same bytes (the same example version
   on the same row) would meet its words. If it would, write the line in terms
   of what changed, not an act being recorded: the same check on the same bytes
   teaches nothing, so it never counts as new evidence. When the words allow
   more than one reading of what changed (for example, whether a version and a
   row that each ran before, but never together, count), show the owner a
   concrete case of each reading and let the owner choose. Place an example for
   the chosen reading. The command you draft carries the words and lines
   exactly as you showed them to the owner.
9. **Make a record branch in a worktree of its own.** The Goal's events and
   examples reach the default branch only through a record PR. Another session
   may work in the main checkout, so never switch its branch or commit there.
   Before you show the owner the act that adds the Goal, run `git fetch origin`
   and `git worktree add -b record/<issue> <path> origin/main` (or the default
   branch), with `<issue>` the issue you grill, or a short name of the want when
   it has no issue, and `<path>` a folder beside the repository. The owner runs
   the act in that path, and every commit of the grill and pick of a Probe
   writer's commit happens there. hyp gives the Goal its id only when it is
   added: rename the branch `record/<G>` then.
10. **Draft the model with `/hyp-statechart`.** When the change moves through
    states, write the Goal's model on the record branch with the
    hyp-statechart skill, which ships with hyp, after the owner consents to
    the words and lines and before the examples are final. When the
    repository keeps a model in `goals/model/`, the Goal's model is a change
    to it, drafted in the Goal's folder. Follow [model.md](model.md) for
    where the draft stays and what to ask the owner about each variant the
    statechart tool reports.

## How to grill the examples

When the owner agreed to the Goal's words and lines, and before the owner adds
the Goal, follow [examples.md](examples.md): you show the readings of each
line, the owner chooses, you draft one example per chosen reading in a draft
draft, try each on H0 with `hyp try --draft`, and the owner adds the Goal
with the draft, `hyp add goal ... --draft NAME`. It
also says when the Goal has enough.

## How to re-grill

At a stop for the owner (`Selected`, `Two candidates`, `Budget used`), when a
selected Goal is reviewed, or when a runner reports a placed example that no
build can meet, follow [re-grill.md](re-grill.md). It reviews the Goal from
what builds really printed and turns each change the owner keeps into an
adjust or a new Goal; at `Selected` it walks the difference with the owner;
and it says when a Goal must end because an example is wrong.

## Hand-off

The grill ends with a record branch that holds the Goal's
folder with the examples the owner's add copied there from the draft, and the owner's add,
and a hand-off for the owner: the Goal's lines, the owner's decisions
numbered with their words, and "Not covered" with each decision no example
checks. Before the record PR, the issues already hold those decisions and the
"Not covered" items they own. With no issue, the record PR's body holds them.

A runner works from the Goal: what `hyp status G --role runner` prints, which
holds the Goal's words, lines and placed examples, and the model in the spec
folder, with the run rules. If a runner reads the hand-off, it is
background, and no "Not covered" item is a task.

The hand-off ends with where the runner's prompt is: `hyp status G` prints
the Start prompt of an open Goal, ready to copy, with the Goal and the
runner's actor, `runner-` and the Goal id in lower case, filled in. Write no
runner prompt of your own, and add no choice to the hand-off for the runner:
the runner looks for no rivals until the owner unselects.

## Done when

- `hyp status G` shows the Goal `open`, and every placed example gives its
  line's verdict on H0.
- Every variant of the model is separated by an example, or the owner said
  either output is fine and it is listed under "Not covered".
- Each decision of the owner is the words, a line or an example, or is listed
  under "Not covered".
- The issue you grill, or the record PR's body when there is no issue, holds
  each decision of the owner in the owner's words, and each "Not covered" item another open issue owns is on that issue.
- The hand-off ends with the sentence that `hyp status G` prints the
  runner's Start prompt.

## Resources

- [run-rules.md](run-rules.md): the run part. Give it to the runner of the Goal.
- [re-grill.md](re-grill.md): the re-grill, at a stop for the owner.
- [model.md](model.md): drafting the Goal's model and asking about its
  variants.
- [examples.md](examples.md): the examples of each line, from the owner's
  choice of readings to the place.
- [probe-writers.md](probe-writers.md): handing an example the generic Probe
  cannot run to a Probe writer.
- The hyp-statechart skill, beside this one: the model, its variants and its
  chart, with the statechart tool.
- [scripts/draw_issue.py](scripts/draw_issue.py) `PAGE.json`: prints an
  issue body from a rethink page's data file, with every item and an empty
  holder table. Exit 1 when the data has a defect, 2 when a file cannot be
  read.
- [scripts/check_holders.py](scripts/check_holders.py) `BODY.md`: lists each
  item of an issue body whose holder is missing, of the wrong kind, unknown to
  the store or ended with no select. Exit 1 with defects, 2 when the body
  cannot be read or `hyp` cannot run. Both scripts use the Python 3 standard
  library only. Tests: `python3 -m unittest discover -s scripts/tests`.
- [probe/run](probe/run): the generic Probe, a POSIX `sh` script. hyp runs it
  on macOS and Linux; `hyp try` exercises it, so it has no test of its own.
