# The grill's moves for v2

The spec of Goal G-D2MVSR. It has no model: the change is skill text in
`skills/hyp-grill/SKILL.md`. It moves no Goal between states and changes no
act, guard, output or exit code of hyp.

## Line 1: the first grill of a written Goal's examples

`SKILL.md`, section "How to grill the examples":

- For each line, the grill asks exactly "What would you do, and what would you
  see?", one line of its message per Goal line, starting with the line's key
  and words. It offers no draft and no options.
- The owner's answer is the example: what the owner does is the `run`, what
  the owner sees is the floor. The grill does not read the example back and
  does not ask the owner to accept it, and it uses no words such as "confirm",
  "approve" or "look right".
- An answer that names two separate things to see gets a warning that the line
  combines two rules, and a split proposal. The owner decides; a kept line
  gets a floor for each thing.
- After each floor, the grill quotes the line's "must never" words and asks
  "What could a wrong build print that this floor still lets through?" Each
  wrong output the owner names is a `ceiling`.
- Once every line has an example with its line's verdict on H0 (the Goal is
  `open`), the grill stops asking. Its last message says the Goal has enough
  examples and holds no question mark.

Step 5 of "How to grill" ("A correction is not consent") now applies to the
owner's words and lines, not to the examples of a recorded Goal.

## Line 2: the re-grill when the owner reports something odd

`SKILL.md`, section "How to re-grill":

- The grill first asks the owner to state the odd thing as "On input X, I must
  see Y." or "On input X, I must never see Y.", before it reads outputs or
  names a cause.
- When a line's words already ask for Y or forbid it, the grill proposes an
  adjust: a new version of that line's example that adds Y as a floor or a
  ceiling. When Y needs new words, it proposes a new Goal. The owner decides.

## Line 3: variants and the Probe writer

- "How to grill", step 10: the grill stops trying a variant at its first
  failed example.
- "How to grill the examples", steps 6 and 7: an example whose owner command
  runs in a plain copy of the subject gets the generic Probe, copied byte for
  byte. An example that needs a setup keeps the owner's command as its `run`;
  the grill writes no Probe code for it and starts a Probe writer whose prompt
  is the probe-writer block of `hyp status G --role probe-writer`, word for
  word, with the example's key, words and setup.
- The section "The issue holds the decisions" applies when the Goal comes
  from an issue.
