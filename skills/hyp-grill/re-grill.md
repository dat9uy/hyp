# Re-grill

The re-grill part of hyp-grill. Read it at a stop for the owner, when a
selected Goal is reviewed, or when a runner reports a placed example that no
build can meet.

The re-grill reviews a Goal against its lines, from what builds really
printed, at each stop for the owner: `Selected`, `Two candidates` and
`Budget used`. A runner's report of an example that no build can meet goes
straight to "When the Goal must end", stop or no stop. Read the table of
outputs `hyp status G` prints and each output it names, in the trace of the
clone that ran it; where this clone's trace lacks a cell, run the `hyp try`
line `hyp status G` prints for it. A met output matters as much as a failed
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

Every kept finding is a place, and a place restarts the budget: tell the
owner this cost before they keep it. At the `Selected` stop the owner first
unselects; the rows and verdicts stay. A changed `run`, input or Probe is a
new example key, and a place that changes no example is refused.

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

The ended Goal's builds stay commits on their branches. They are not
Hypotheses of the new Goal, and the hand-off never offers one to the runner.
A runner that builds on one picks it onto the new record branch, because
`hyp select` reads the Goal's spec from the selected commit. Report what
agents found about those builds, such as their verdicts or why a Probe
failed, as findings in the hand-off.
