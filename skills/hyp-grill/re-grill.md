# Re-grill

The re-grill part of hyp-grill. Read it at a stop for the owner, or when a
selected Goal is reviewed.

The re-grill reviews a Goal against its lines, from what builds really
printed, at each stop for the owner: `Selected`, `Two candidates` and
`Budget used`. Read the table of outputs `hyp status G` prints and each
output it names, in the trace of the clone that ran it; where this clone's
trace lacks a cell, run the `hyp try` line `hyp status G` prints for it. A
met output matters as much as a failed one.

- **Discover ceilings.** In each met output, look for text the line's words
  forbid that its floors let through. Show the owner the example, its floors
  and that output, and ask whether the text must never appear. A kept one
  becomes a ceiling, copied byte for byte from the output, as a new version
  of that example.
- **Review floors.** When a build met an example but its output is not what
  the line means, the floor is too loose: add a floor or a ceiling, as a new
  version. When a surface the line reaches has no example, add a new example
  key. When a build the owner accepts fails only on a floor, the floor is too
  strict. A version cannot loosen it, a placed folder is never removed, and a
  line is met only when every example is met, so the Goal ends: the owner
  lands it, and a new Goal with `--from G` holds the corrected example.

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
  needs words that no line has, propose a new Goal: the owner lands this
  Goal, and the new Goal cites it. A new line is always a new Goal, never an
  edit of an open one. Use the words "an adjust" or "a new Goal" in the
  proposal. The owner decides.
- Use `hyp show <actor>` to read exactly what hyp printed to the runner.
