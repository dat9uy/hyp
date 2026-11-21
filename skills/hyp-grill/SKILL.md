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
  selected Goal is to be reviewed: that is the re-grill.
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

1. **Read the event and the store.** Read the issue or event that points at the
   change, `hyp help`, and `hyp status`. Restate the problem in the owner's
   terms, with no mechanism in it. When the Goal you draft will cite a landed
   Goal with `--from`, read `hyp status` of that Goal: it prints what that
   Goal decided, its words, lines, examples and their verdicts. Read it there,
   not from the store's files.
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
5. **A correction is not consent.** When the owner corrects a draft, apply the
   correction and show the result again. Run no act on a correction; wait for
   the owner's consent to that act.
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
    hyp-statechart skill, which ships with hyp: a TLA+ spec in
    `goals/<G>/specs/`, one variant per potential Hypothesis, and the chart.
    The statechart tool, not you, says which rules each variant breaks and
    which examples each variant would meet:
    - Draft the model after the owner consents to the words and lines and
      before the examples are final, so that what the tool predicts can still
      change the examples.
    - Run the tool on the check config. It writes one variant config per
      variant under `specs/variants/`, prints each rule no variant breaks as
      not shown to fail, prints a walk of the model for each example, and
      predicts from the model alone whether each example is met under each
      variant.
    - Try each variant the tool predicts to meet every example: write it as a
      patch and run each example on it with `hyp try G LINE/EXAMPLE
      --worktree`.
    - For each other variant, try one example the tool predicts to fail on
      that variant's patch: one spot check for each other variant.
    - Give the tool what `hyp try` printed: the verdict line of each tried
      example goes in `specs/tried/<variant>/<line>/<example>.txt`, and the
      tool's report sets each tried cell beside its prediction.
    - Where a tried verdict differs from its prediction, fix the model, not an
      example, and run the tool again.
    - Say so in one line of the hand-off for each tried verdict that differed
      from its prediction, and what you fixed in the model.
    - For each tried variant that meets every example and prints something
      else on a step no example takes, ask the owner one question in plain
      English, at the example level and with no model word: show the case, what
      each build prints for it, and ask which output is wanted.
    - When the owner answers "either output is fine", list the case under
      "Not covered". Otherwise place an example that separates the two builds.
    - A prediction alone, never tried, never reaches the owner: only a tried
      variant does.

    At select, hyp refuses a model that has a rule no variant config names, or
    no chart; a spec folder with no model is selected as before.

## How to write examples

1. **List the surfaces before the examples.** Before you write any example,
   list every surface the line reaches: each file, output or command where the
   line's words can be true or false. Then give each surface an example, or
   report it under "Not covered".
2. **One folder per example**, at `goals/<G>/<line>/<example>/`, with an
   `example.txtar` and a `probe/` folder. The `example.txtar` sections are
   `words` (the owner's sentence), `run` (the command run on the subject),
   `input/<path>` files, and at least one `floor: contains` or `floor: equals`,
   with any `ceiling`. Floors are literals, byte for byte.
3. **Use the generic Probe.** Copy [probe/run](probe/run) from this skill into
   the example's `probe/run` and keep it executable. It runs `$HYP_RUN` in the
   subject, prints its stdout, then a last line `exit N`. Write a new Probe only
   when the example needs one that this one cannot be.
4. **Make the check able to fail.** A gap example must fail on H0 and a guard
   example must be met on H0. `hyp try` shows the verdict on H0 before you
   place; `hyp try --worktree` can also show it on the files of a checkout.
5. **Run `hyp try` before you place.** Run `hyp try G LINE/EXAMPLE` and read its
   verdict, exit code, output and stderr. It records nothing and uses no run.
   Only then commit the folder and place it, with the place line
   `hyp status G --role grill` prints for that line.

## How to re-grill

The re-grill reviews a Goal against its line, not against the build that met it.

- **At a stop for the owner** (`Two candidates`, `Budget used`), read the table
  of outputs `hyp status G` prints. If an example let a build pass that the
  line does not mean, or failed one it does, the fix is an example: a new
  version of an example may only add a floor or a ceiling; a changed `run`,
  input or Probe is a new example key. A placement that changes an example moves
  the budget's boundary.
- **After a select**, the Goal refuses every place. When the review finds a
  gap that an example on an existing line can hold, the owner unselects the
  Goal first. The Goal is then open again, its rows and verdicts stay, and you
  place the new example as at any other stop. A gap that needs a new line is a
  new Goal: the owner lands this Goal, and the new Goal cites it.
- A new line is always a new Goal, never an edit of an open one.
- Use `hyp show <actor>` to read exactly what hyp printed to the runner.

## Hand-off

The grill ends with a record branch that holds the examples and the place
events, and a hand-off for the owner: the Goal's line, the owner's decisions
numbered with their words, and "Not covered" with each decision no example
checks. Before the record PR, the issues already hold those decisions and the
"Not covered" items they own.

A runner works from the Goal: what `hyp status G --role runner` prints, the
placed examples and the model in the spec folder, with the run rules. If a
runner reads the hand-off, it is background, and no "Not covered" item is a
task.

The owner starts a runner with the prompt in the Start section of the run
rules, with the Goal and the runner's actor filled in. Do not write a runner
prompt of your own, and add no choice to the hand-off for the runner: the
runner looks for no rivals until the owner unselects.

## Done when

- `hyp status G` shows the Goal `open`, and every placed example gives its
  line's verdict on H0.
- Each decision of the owner is the words, a line or an example, or is listed
  under "Not covered".
- The issue you grill holds each decision of the owner in the owner's words,
  and each "Not covered" item another open issue owns is on that issue.

## Resources

- [run-rules.md](run-rules.md): the run part. Give it to the runner of the Goal.
- The hyp-statechart skill, beside this one: the model, its variants and its
  chart, with the statechart tool.
- [probe/run](probe/run): the generic Probe, a POSIX `sh` script. hyp runs it
  on macOS and Linux; `hyp try` exercises it, so it has no test of its own.
