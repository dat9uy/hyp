# An add writes the Goal's folder before it records the Goal

The spec of Goal G-Z1YX7V. It has no model, and the kept model in
`goals/model/` does not change: AddGoal moves a Goal to open on the same event
as before. Only the order of two writes inside the act changes, and one print of
`hyp status`.

## Line 1: a folder that cannot be written records no Goal

- `hyp add goal --draft NAME` writes `goals/<G>/` from the draft's snapshot after
  the event's id is known and before the event is written. The id comes from the
  store's lock, so the folder name and the event agree.
- When the folder cannot be written, the add exits 3 and writes no event. A
  folder it began is removed.
- `hyp status`, with no Goal, prints `no Goal is in work.` when no Goal is
  written, open or selected, also when the store holds no Goal.

## Line 2: a recorded Goal has its folder

- A successful add records the Goal open, and `goals/<G>/` holds the examples
  whose hashes the goal.add carries, as before.
