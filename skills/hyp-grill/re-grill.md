# Re-grill

The re-grill part of hyp-grill. It reviews a Goal against its lines, from what
builds really printed, at each stop for the owner (`Selected`,
`Two candidates` and `Budget used`) and when a selected Goal is reviewed. A
runner's report of an example that no build can meet goes straight to "When
the Goal must end", stop or no stop. At `Two candidates` and `Budget used`,
read the table of outputs `hyp status G` prints and each output it names, in
the trace of the clone that ran it; where this clone's trace lacks a cell, run
the `hyp try` line `hyp status G` prints for it. At `Selected`, read each
example's output on the selected commit C with
`hyp try G LINE/EXAMPLE --commit C`. A met output matters as much as a failed
one.

- **Discover ceilings.** In each met output, look for text the line's words
  forbid that its floors let through. Show the owner the example, its floors
  and that output, and ask whether the text must never appear. A kept one
  becomes a ceiling, copied byte for byte from the output, as a new version
  of that example.
- **Review floors.** When a build met an example but its output is not what
  the line means, the floor is too loose: add a floor or a ceiling, as a new
  version. When a surface the line reaches has no example, add a new example
  key. When a build the owner accepts fails only on a floor, the floor is too
  strict, and the Goal must end (below).

Every ceiling, floor or example the owner keeps is a place, and a place
restarts the budget: tell the owner this cost before they keep it. At the
`Selected` stop the owner first unselects; the rows and verdicts stay. A
changed `run`, input or Probe is a new example key, and a place that changes
no example is refused.

- **When the owner reports something odd**, ask the owner, as a question, to
  state it in one of two forms, quoted in your question: "On input X, I must
  see Y." or "On input X, I must never see Y." You may read the candidate's
  outputs to find what looks off, but ask for the statement before you
  propose a fix.
- **Then compare Y with the lines' words.** When a line's words already ask
  for Y or forbid it, propose an adjust: a new version of that line's example
  that adds Y as a floor (must see) or as a ceiling (must never see). When Y
  needs words that no line has, propose a new Goal: the owner ends this
  Goal, as "When the Goal must end" says, and the new Goal cites it. A new
  line is always a new Goal, never an edit of an open one. Use the words "an
  adjust" or "a new Goal" in the proposal. The owner decides.
- Use `hyp show <actor>` to read exactly what hyp printed to the runner.
- **Check a report against the code before you pass it on.** A runner's
  report, or another agent's, about what the build does is prose. Before you
  tell the owner what the build does or does not do, read the code the report
  names, or `hyp spec` on the build, and say which you read. A claim you did
  not check goes to the owner as the agent's words, never as a fact.

## At `Selected`: walk the difference

At `Selected`, every verdict is met and the owner wrote every floor, so
neither shows the owner anything new. The owner learns a want from a
consequence of the build. So show the owner what the build does beyond its
examples, as real runs, before the owner reads the code. This covers what the
owner put in a line and what the owner left out.

1. **Find the cases in the diff.** Start from `hyp status G`: under
   `kept model:` it lists each step and rule of the kept model that C adds or
   changes against H0, each rule with the words of the line it names. Write
   one case for each listed step or rule. Then read the diff between H0 and
   C, and write one case for each other changed part. A case is a command
   the owner would type that reaches that part. Skip a case that a placed example already
   holds: its verdict is met by rule.
2. **Write each case as an example folder that you do not place.** Put it
   under the line whose words are nearest to it, or under any line when none
   is near: the folder is never placed. Give it the generic Probe, or a Probe
   writer's Probe when the case needs a setup, as in
   [examples.md](examples.md). An example file needs a floor, so give it
   any short floor: the walk shows outputs, never the verdict. hyp cuts the
   Probe's last line `exit N` off before it compares, so no floor sees it.
3. **Ask about one case at a time.** Show the case and its output on H0, from
   `hyp try G LINE/CASE`. Ask: "What do you expect the build to print?" The
   owner may answer "I expect Y", "either is fine" or "I don't know". Each
   answer is welcome. Do not show the build's output before the owner answers.
4. **Then show the build's output**, from `hyp try G LINE/CASE --commit C`,
   byte for byte, beside H0's output. Write no summary of your own: the owner
   reads outputs and code, not prose about them.
5. **Each case ends in the owner's words:**
   - **It fits.** Keep nothing. `hyp try` saves cells in this clone's trace
     only, so the walk adds nothing to git. Commit no walk folder.
   - **Hold it, and a line's words cover it.** This is an adjust. The owner
     unselects. The case becomes a new example key on that line, and is
     placed. Its floor is the owner's answer, or, after "either is fine" or
     "I don't know", the part of an output the owner names. The place moves
     the budget's boundary, so the runner runs the selected build again, as
     run 1 of 3.
   - **Hold it, and no line's words cover it.** This is a new Goal. The owner
     lands G, then adds a Goal `--from G`, in the same record PR.
6. **Code reading stays behind the walk.** When the owner finds a consequence
   in the code that no case showed, record it with `hyp finding add`, with the
   case that would have shown it as its evidence.

### A partial fix at `Selected`

A partial fix, or a part the build left out, that the walk or the runner's
report shows is recorded as a finding at `Selected` by the orchestrator, the
session that starts the runner and reports to the owner, whatever the owner
answers: keep the build, unselect it or end the Goal. A PR's notes alone do
not hold it. Record it with `hyp finding add`, and name it in the tracking
issue's holder table when the Goal has one.

## When the Goal must end

A placed example cannot be repaired: a version only adds floors and
ceilings, a placed folder is never removed, and a line is met only when every
example is met. So when an example is wrong, the Goal ends, and a new Goal
with `--from G` holds the corrected example. Two causes lead here:

- **A floor that is too strict.** A build the owner accepts fails only on a
  floor.
- **A Probe that no correct build can run.** A placed Probe exits 3 on a build
  that keeps the Goal's rules, for example when its setup needs the H0
  behaviour that its line removes. Only a build shows this, because
  `hyp place` measures each Probe on H0 only.

The owner ends the Goal with the act that `hyp status G` offers: `hyp land G`
at a stop, `hyp close G` when the Goal has no stop, such as with 0 runs.

A Goal that ends with no select gives up its want. When the owner still wants
it, a new Goal holds it once its lines are known, and an issue holds it until
then. The `--why` of the land or close says which, or that the want is dropped.

The ended Goal's builds stay commits on their branches. They are not
Hypotheses of the new Goal, and the hand-off never offers one to the runner.
A runner that builds on one picks it onto the new record branch, because
hyp reads the Goal's spec from the selected commit. Record what
agents found about those builds, such as their verdicts or why a Probe
failed, as findings with `hyp finding add`, and name each in the new Goal:
with `--finding F` in its `add goal`, or in the `findings` section of one of
its examples.

## Before a land

Before a land, show the owner each decided part that no line builds: a
decision the owner made in the grill or the re-grill that no line of the Goal
holds. A decided want is a line of a written Goal, never only an issue or a
comment, so give each one as a drafted `add goal` with its words, lines and
why, to run `--from` the Goal after the land:

```text
hyp add goal --base COMMIT --from G --words TEXT --line gap:TEXT --why TEXT
```

The owner edits and runs it after `hyp land G`. Read
`hyp finding list --not-folded` with the owner before, and add
`--finding F` to the drafted `add goal` for each finding that bears on the
new Goal.
