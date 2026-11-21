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
    hyp-statechart skill, which ships with hyp: a TLA+ spec in
    `goals/<G>/specs/`, one variant per potential Hypothesis, and the chart.
    The statechart tool, not you, says which rules each variant breaks and
    which examples each variant would meet:
    - Draft the model after the owner consents to the words and lines and
      before the examples are final, so that what the tool predicts can still
      change the examples.
    - The hyp-statechart skill owns how to run the tool, how to try each
      variant as a patch with `hyp try G LINE/EXAMPLE --worktree`, and what
      its part of the hand-off holds. Follow it there; this step adds only
      what you ask the owner.
    - For each tried variant that meets every example and prints something
      else on a step no example takes, ask the owner one question in plain
      English, at the example level and with no model word: show the case, what
      each build prints for it, and ask which output is wanted.
    - When the owner answers "either output is fine", list the case under
      "Not covered". Otherwise place an example that separates the two builds.
    - A prediction alone, never tried, never reaches the owner: only a tried
      variant does.
    - Stop trying a variant at its first failed example. A variant that fails
      one example meets not every example, so its other tries teach nothing.

    At select, hyp refuses a model that has a rule no variant config names, or
    no chart; a spec folder with no model is selected as before.

## How to grill the examples

Use this when the Goal is recorded and its lines need examples. The owner's
answers are the examples: you write them down, you do not design them. The
grill takes two rounds of questions, in this order: first the open question
about every line, then, after you write the floors, the split and ceiling
questions.

1. **Ask the open question about each line.** Ask, for each line, exactly
   "What would you do, and what would you see?" Name the line, by its key and
   words, with its question. Offer no draft and no options, and ask nothing yet
   about ceilings or splits.
2. **The answer is the example.** What the owner does is the `run`, in the
   owner's words (for example `sh greet.sh`). What the owner sees, the want,
   is the floor, as a literal. Write the example from the answer. Do not read
   it back to the owner and do not ask the owner to accept it.
3. **Two things to see is two rules.** When one line's answer names two
   separate things to see (for example a list and a count), warn that the line
   combines two rules, and propose to split it into two lines. Ask the owner
   whether to split; the owner decides. Raise a split only after the owner's
   answer names two things, never in advance. A split changes the lines, so the
   owner records it while the Goal is `written`. When the owner keeps one
   line, its example holds a floor for each thing.
4. **Ask the ceiling question about each line.** After you write a line's
   floor, read aloud the line's "must never" words, quoted exactly as the line
   has them (for example: line-1 says "never prints hello"), and ask: "What
   could a wrong build print that this floor still lets through?" Ask it word
   for word, for every line, also a line with no "never" words. Each wrong
   output the owner names becomes a `ceiling` section of that line's example.
   When the owner names none, the example has no ceiling. Ask the split and
   ceiling questions after the owner answered the open questions.
5. **One folder per example**, at `goals/<G>/<line>/<example>/`, with an
   `example.txtar` and a `probe/` folder. The `example.txtar` sections are
   `words` (the owner's sentence), `run` (the command run on the subject),
   `input/<path>` files, and at least one `floor: contains` or `floor: equals`,
   with any `ceiling`. Floors and ceilings are literals, byte for byte.
6. **Give the generic Probe to each example it can run.** The generic Probe,
   [probe/run](probe/run), runs `$HYP_RUN` in a copy of the subject, prints its
   stdout, then a last line `exit N`. It can run an example when the owner's
   command, run in a plain copy of the subject, shows what the owner sees.
   Copy it with `cp` into the example's `probe/run`, byte for byte, and keep it
   executable.
7. **Hand every other example to a Probe writer.** When the owner's answer
   names a state to set up before the command (for example "in a git
   repository with three commits"), the generic Probe cannot run it. Then:
   - Write and commit the `example.txtar`. Its `run` stays the owner's command,
     word for word; the setup goes in the Probe, never in the `run`.
   - Write no Probe code for it yourself. Start a Probe writer: an agent
     whose prompt is the probe-writer block that
     `hyp status G --role probe-writer` prints, copied word for word, followed
     by the example's key (for example `line-2/a`), its words and the setup it
     needs.
   - The Probe writer commits the example's `probe/` folder and returns the
     output of `hyp try`. Start no Probe writer for an example the generic
     Probe can run.
8. **Run `hyp try` before you place.** A gap example must fail on H0 and a
   guard example must be met on H0. Run `hyp try G LINE/EXAMPLE` and read its
   verdict, exit code, output and stderr. It records nothing and uses no run.
   When the verdict is not the one the line needs, fix the example and try
   again. Then commit the folder and place it, with the place line
   `hyp status G --role grill` prints for that line.
9. **Stop when the Goal has enough.** Once every line has an example that
   gives its line's verdict on H0, `hyp status G` shows the Goal `open`. Then
   stop asking: say that the Goal has enough examples, and give the hand-off.
   Ask no further question and offer no more examples; more examples come only
   from real outputs.
10. **Report the surfaces no example covers.** List every surface each line
    reaches: each file, output or command where the line's words can be true
    or false. Each surface that no example covers goes under "Not covered" in
    the hand-off. Listing a surface is not a question to the owner.

## How to re-grill

The re-grill reviews a Goal against its line, not against the build that met it.

- **When the owner reports something odd**, run `hyp status G` to see the
  Goal's lines, then ask the owner to state the odd thing in one of two
  forms, quoted in your question: "On input X, I must see Y." or "On input X,
  I must never see Y." You may read the candidate's outputs to find what
  looks off, but ask for the statement before you propose a fix.
- **Then compare Y with the lines' words.** When a line's words already ask
  for Y or forbid it, propose an adjust: a new version of that line's example
  that adds Y as a floor (must see) or as a ceiling (must never see). When Y
  needs words that no line has, propose a new Goal. Use the words "an adjust"
  or "a new Goal" in the proposal. The owner decides. An adjust of a selected
  Goal waits for the owner's unselect (see "After a select").
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
  new Goal: the owner lands this Goal, and the new Goal cites it. A new line
  is always a new Goal, never an edit of an open one.
- Use `hyp show <actor>` to read exactly what hyp printed to the runner.

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
