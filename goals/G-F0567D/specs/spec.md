# The statechart tool reads a rule's source as a reference to one Goal's line or one guard of hyp spec, and an over-budget check says it predicts nothing

The spec of Goal G-F0567D. It has no model, and the kept model in
`goals/model/` does not change: the change is to the statechart tool in
`skills/hyp-statechart/scripts/goal.ts` and `worker.ts`, and no step or rule of
hyp's machine moves.

## Line 1: a line is named only by a source that names this Goal

- A source is read as a list of `G-XXXXXX:line-N`. goal.md counts line N of the
  Goal whose spec folder the tool writes into as named only by a reference
  whose Goal is that Goal. `G-OTHER:line-1` names no line of this Goal.

## Line 2: a source of another form is refused

- A source is a list of `G-XXXXXX:line-N`, or `hyp spec, guard NAME`. For each
  rule with any other source, goal.md names the rule and its source, and the
  tool exits 3 when every rule holds.

## Line 3: an over-budget check predicts nothing

- When the model has more states than the chart budget, goal.md lists every
  example under "Examples to try on every variant", and the row "Separated by"
  reads `not predicted`, never `none`.

## Line 4: a model within the budget reads as before

- A source of either form on a model within the budget gives the goal.md it
  gave before.
