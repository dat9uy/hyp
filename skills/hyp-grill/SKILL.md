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
- hyp stopped a Goal for the owner (`Two candidates` or `Budget used`), or a
  selected Goal is to be reviewed: that is the re-grill, in
  [re-grill.md](re-grill.md).
- A runner needs the run rules: give it [run-rules.md](run-rules.md), the run
  part of this skill. It is a part of this skill, not a second skill.
- Not for: writing or running the build itself. The runner does that.

## What hyp checks, and what this skill holds

hyp holds the state machine. `hyp help` lists every command, and
`hyp status G --role grill` prints the acts the grill may take now, as ready
command lines with their keys filled in. hyp's rule is: "An agent runs a line
hyp printed. A skill names no write command." So this skill shows only reads
(`hyp help`, `hyp status`, `hyp try`, `hyp show`); every act that writes, you
copy from a line hyp printed. Set `HYP_ACTOR` to your actor name and
`HYP_ROLE=grill` for every command. The owner records the Goal itself, with the
act `hyp help` lists for adding a Goal; you draft the words and lines it takes.
In that draft, set `HYP_ACTOR` to the owner's actor, never a placeholder: the
actor of the `goal.add` events in the store, which `.tree/events/<actor>.jsonl`
names. Ask the owner once, only when the store has no Goal yet or several actors
added Goals. The owner's land, unselect and select of a Goal are allowed only
to the actor that added it, so a wrong actor locks the owner out of them.

## Decisions

A decision of the owner is held in one of three places: the Goal's words, a
line, or an example. Nothing else carries it to the runner.

A decision that no example checks is a gap: the grill reports it under
"Not covered" in the hand-off, with where the decision came from, so the owner
sees what the examples cannot catch. It is never passed to the runner as an
extra rule.

## The issue holds the decisions

A chat, a local page and a memory note are read by no one else, so what the
grill learns goes to the issues, in the same session, with `gh issue comment`.
The issue is the owner's record, not a rule for the runner:

- **Each decision of the owner goes to the issue you grill**, with the owner's
  words quoted and kept apart from your proposals. Say which part of the issue
  a decision replaces. Your inference is a proposal: never write it as the
  owner's words, in the issue or anywhere else.
- **A "Not covered" item that another open issue owns** (it waits on that
  issue's change, or that issue decides it) goes to that issue too, with the
  owner's words, and the hand-off names that issue.
- **A record cites only what others can read**: an issue, a PR or a commit.
  Never a gitignored or local file, such as a design page, a report or a
  memory note, in the Goal's words, a `--why` or a comment.

## How to grill

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
   can be in (`written`, `open`, `selected`, `landed`) and every command that
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
9. **Make a record branch before the Goal is recorded.** The Goal's events
   and examples reach the default branch only through a record PR, never by a
   write on it. Before you show the owner the act that adds the Goal, run
   `git fetch origin` and `git checkout -b record/<issue> origin/main` (or the
   repository's default branch) in the checkout where the owner will run it,
   where `<issue>` is the number of the issue you grill. hyp gives the Goal
   its id only when the Goal is added, so rename the branch `record/<G>` with
   `git branch -m` once hyp prints it.
10. **Draft the model with `/hyp-statechart`.** When the change moves through
    states, write the Goal's model on the record branch with the
    hyp-statechart skill, which ships with hyp, after the owner consents to
    the words and lines and before the examples are final. Follow
    [model.md](model.md) for when to draft it and what to ask the owner about
    each variant the statechart tool reports.

## How to grill the examples

Use this when the Goal is recorded and its lines need examples. You propose
the examples; the owner chooses their direction: which reading of each line an
example checks. A reading is one part or one meaning of a line's words. Ask in
two rounds: first the readings of every line, then, after you write the
floors, the split question. Ask no question about ceilings: a ceiling is text
a build really printed, from H0 here (step 7) or from candidates at the
re-grill.

While a question is open, each message that ends your turn repeats it in
full: the owner may read only the last message, not the status before it.

1. **Show the readings of each line, and let the owner choose.** Start the
   part of the message for each Goal line with one line that gives its key and
   words, and put its readings under it. Each reading shows a case, what the
   owner would do, and the output it expects, so that the owner chooses between
   outputs, not between descriptions:
   - **Where a model exists**, each variant that no example separates is one
     reading of the line its rules come from. Show the case as two outputs,
     the model's print and the variant's print, each marked "model print, not
     built", and ask which output is wanted.
   - **A line no variant reaches** gets its readings from its words: name the
     parts or meanings the line can check, show beside each the output you
     expect, marked "not built", and ask which readings the line means. A
     reading in words alone is not enough to write a floor from.

   A model print is a direction, never a verdict: the example written from the
   choice is still tried on H0 before it is placed (step 7), and a model result
   never counts at select. Use plain English and no model word other than the
   "model print, not built" mark.
2. **Write one example per chosen reading.** Its `words` section is the
   reading in the exact words the owner picked, so the select report says what
   each example checks. Its `run` is the command the owner would type (for
   example `sh greet.sh`). Write floors from the part where the wanted output
   differs from the other, as short literals: a model print is a guess at the
   real text, so a long literal copied from it fails on bytes the owner never
   chose. Never copy the rejected output into a ceiling: a wrong build prints
   the same fault in other words. A rejected case the owner wants checked is
   its own example (for example "on a branch, no warning"), whose ceiling
   reuses a literal the owner wants elsewhere, such as a sibling's floor.
   When the owner answers "either output is fine", write no example for that
   reading and list it under "Not covered". The choice is the decision: do not
   ask the owner to approve the file again.
3. **Two things to see is two rules.** When the readings the owner chose for
   one line are separate things to see (for example a list and a count), warn
   that the line combines two rules, and propose to split it into two lines.
   Ask the owner whether to split; the owner decides. Raise a split only after
   the owner's choice names two things, never in advance. A split changes the
   lines, so the owner records it while the Goal is `written`. When the owner
   keeps one line, it holds one example per reading.
4. **One folder per example**, at `goals/<G>/<line>/<example>/`, with an
   `example.txtar` and a `probe/` folder. A new reading is a new example key,
   never a new version of an existing one. The `example.txtar` sections are
   `words` (the reading the owner chose), `run` (the command run on the subject),
   `input/<path>` files, and at least one `floor: contains` or `floor: equals`,
   with any `ceiling`. Floors and ceilings are literals, byte for byte.
5. **Give the generic Probe to each example it can run.** The generic Probe,
   [probe/run](probe/run), runs `$HYP_RUN` in a copy of the subject, prints its
   stdout, then a last line `exit N`. It can run an example when the example's
   command, run in a plain copy of the subject, shows what the owner sees.
   Copy it with `cp` into the example's `probe/run`, byte for byte, and keep it
   executable.
6. **Hand every other example to a Probe writer.** When the reading needs a
   state set up before the command (for example "in a git repository with
   three commits"), the generic Probe cannot run it. Follow
   [probe-writers.md](probe-writers.md): write no Probe code yourself, start a
   Probe writer from the block `hyp status G --role probe-writer` prints, and
   bring its commit onto the record branch.
7. **Run `hyp try` before you place.** A gap example must fail on H0 and a
   guard example must be met on H0. Run `hyp try G LINE/EXAMPLE` and read its
   verdict, exit code, output and stderr. It records nothing and uses no run.
   When the verdict is not the one the line needs, fix the example and try
   again. A gap example's H0 output shows the behaviour the line wants gone:
   text of it that a half-fixed build could still print becomes a ceiling,
   copied byte for byte. Name each such ceiling in the hand-off with its H0
   text; the owner may drop it. When every line holds its examples, commit the folders and place
   the whole set with the one place line `hyp status G --role grill` prints,
   `hyp place G`. hyp runs each new or changed example on H0 and places the
   set only when every example gives its line's verdict there. Otherwise it
   places none of the set and names each example that disagrees, each line
   that holds no example and each folder it cannot place, with the guard that
   refused it. Fix every cause and place again. A placed example stays in the
   set: never remove its folder.
8. **Stop when the Goal has enough.** The Goal has enough when every line
   holds an example that gives its line's verdict on H0 (`hyp status G` shows
   the Goal `open`), and every variant of the model is separated by an example
   or the owner said either output is fine (listed under "Not covered"). Then
   stop asking: say "This is enough.", and give the hand-off.
   Ask no further question and offer no more examples; more examples come only
   from real outputs.
9. **Report the surfaces no example covers.** List every surface each line
   reaches: each file, output or command where the line's words can be true
   or false. Each surface that no example covers goes under "Not covered" in
   the hand-off. Listing a surface is not a question to the owner.

## How to re-grill

At a stop for the owner (`Selected`, `Two candidates`, `Budget used`), or when
a selected Goal is reviewed, follow [re-grill.md](re-grill.md): it reviews the
Goal from what builds really printed, finds ceilings and loose floors, and
turns each kept finding into an adjust or a new Goal.

## Hand-off

The grill ends with a record branch that holds the examples and the place
events, and a hand-off for the owner: the Goal's lines, the owner's decisions
numbered with their words, and "Not covered" with each decision no example
checks. Before the record PR, the issues already hold those decisions and the
"Not covered" items they own.

A runner works from the Goal: what `hyp status G --role runner` prints, which
holds the Goal's words, lines and placed examples, and the model in the spec
folder, with the run rules. If a runner reads the hand-off, it is
background, and no "Not covered" item is a task.

The hand-off ends with the runner's prompt, ready to copy: the Start prompt
of the run rules, word for word, in one `text` block, with `<G>`, `<actor>`
and `<issues>` filled in. `<issues>` is the issue you grill and each other
issue the owner names as settled by the Goal. Propose the runner's actor, such as `runner-` and the Goal
id in lower case, and let the owner veto it; it is never the owner's actor.
Change no other word of the prompt, write no runner prompt of your own, and
add no choice to the hand-off for the runner: the runner looks for no rivals
until the owner unselects.

## Done when

- `hyp status G` shows the Goal `open`, and every placed example gives its
  line's verdict on H0.
- Every variant of the model is separated by an example, or the owner said
  either output is fine and it is listed under "Not covered".
- Each decision of the owner is the words, a line or an example, or is listed
  under "Not covered".
- The issue you grill holds each decision of the owner in the owner's words,
  and each "Not covered" item another open issue owns is on that issue.
- The hand-off ends with the Start prompt of the run rules, filled in with the
  Goal, the runner's actor and the issues the Goal serves.

## Resources

- [run-rules.md](run-rules.md): the run part. Give it to the runner of the Goal.
- [re-grill.md](re-grill.md): the re-grill, at a stop for the owner.
- [model.md](model.md): drafting the Goal's model and asking about its
  variants.
- [probe-writers.md](probe-writers.md): handing an example the generic Probe
  cannot run to a Probe writer.
- The hyp-statechart skill, beside this one: the model, its variants and its
  chart, with the statechart tool.
- [probe/run](probe/run): the generic Probe, a POSIX `sh` script. hyp runs it
  on macOS and Linux; `hyp try` exercises it, so it has no test of its own.
