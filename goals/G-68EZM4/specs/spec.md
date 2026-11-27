# hyp names no event kind it does not act on

The spec of Goal G-68EZM4. It has no model, and the kept model in
`goals/model/` does not change: no command wrote `budget.reset` or
`result.retract`, so removing them moves no Goal between states. The kept
model has no step and no rule for either kind.

## Line 1: the spec lists neither kind

- `events.Kinds` holds thirteen kinds. `budget.reset` and `result.retract`
  are not in it, and their payload types are gone.
- The guard Own Result and the codes `no-result-to-retract` and
  `not-own-result` are gone from `internal/fault`, `internal/spec` and the
  texts in `internal/channel`.
- A Result has no retracted mark. A live cell is a cell of a Result that
  Stable did not refuse.

## Line 2: the budget's boundary

- `hyp spec` says: the boundary is the latest of the latest example-changing
  place and the latest unselect, else the Goal's start.
- It no longer says that a retracted Result counts.

## Line 3: an old event of either kind

- A line whose type is `budget.reset` or `result.retract` does not parse, as
  every type outside the closed set. `hyp status` lists it under
  `not read at replay:` as `unknown event type <type>`, from git or from the
  clone's trace.

## Line 4: the budget restarts as before

- The Goal's epoch moves at an example-changing place and at an unselect, as
  on H0. After either, `hyp status G` shows `runs: 0 of 3 used`.
