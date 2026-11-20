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
   terms, with no mechanism in it.
2. **Ask product questions only.** Ask what the owner wants to be true, never
   how to build it: the file layout, the command shape and the names inside the
   build are the runner's to choose, and the owner vetoes them at merge.
3. **Write each option for the owner's choice.** Each option has a short label
   and a description. The description says what the option would make the
   runner build, and the cheapest build that meets it, so the owner sees the
   price of each answer.
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
   can be in (`written`, `open`, `selected`) and every command that prints that
   form, so the answer covers each one.
8. **Draft the words and lines.** One sentence of words. Each line is `gap`
   (false on H0, the build must make it true) or `guard` (true on H0, the build
   must keep it true). A Goal needs at least one gap line. The owner records it.
9. **Switch to a record branch before the Goal is recorded.** The Goal's events
   and examples reach the default branch only through a record PR, never by a
   write on it. Before you show the owner the act that adds the Goal, run
   `git fetch origin` and `git switch -c record/<issue> origin/main` (or the
   repository's default branch) in the checkout where the owner will run it,
   where `<issue>` is the number of the issue you grill. hyp gives the Goal
   its id only when the Goal is added, so rename the branch `record/<G>` with
   `git branch -m` once hyp prints it.

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
- **After a select**, the Goal refuses every further place. A gap the review
  finds is a new Goal that cites this one, grilled from the start.
- A new line is always a new Goal, never an edit of an open one.
- Use `hyp show <actor>` to read exactly what hyp printed to the runner.

## Hand-off

The grill ends with a record branch that holds the examples and the place
events, and a hand-off the owner and the runner read: the Goal's line, the
owner's decisions numbered with their words, and "Not covered" with each
decision no example checks. Before the record PR, the issues already hold those
decisions and the "Not covered" items they own.

## Done when

- `hyp status G` shows the Goal `open`, and every placed example gives its
  line's verdict on H0.
- Each decision of the owner is the words, a line or an example, or is listed
  under "Not covered".
- The issue you grill holds each decision of the owner in the owner's words,
  and each "Not covered" item another open issue owns is on that issue.

## Resources

- [run-rules.md](run-rules.md): the run part. Give it to the runner of the Goal.
- [probe/run](probe/run): the generic Probe, a POSIX `sh` script. hyp runs it
  on macOS and Linux; `hyp try` exercises it, so it has no test of its own.
