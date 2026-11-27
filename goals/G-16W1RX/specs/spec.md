# A draft of examples lives in the clone's trace, and an add writes the Goal's folder from the bytes it hashed

The spec of Goal G-16W1RX. It has no model, and the kept model in
`goals/model/` does not change: AddGoal and Place move a Goal between the same
states on the same events as before. Only where an example set is read from,
and one Guard of Own path, change.

## Line 1: a draft is a named folder of the clone's trace

- `hyp draft path NAME` prints `drafts/NAME` in this clone's trace, under
  `~/.hyp/traces`, and creates nothing. No commit holds it.
- `hyp try --draft NAME --base COMMIT LINE/EXAMPLE`, `hyp add goal --draft NAME`
  and `hyp place G --draft NAME` read the example set from that folder as
  `LINE/EXAMPLE`. `--examples` is gone: no path of the repository holds a draft.
- A draft that is not there, or a name that is not an actor-style name, is a
  bad argument (exit 2).

## Line 2: one read of the draft

- An add or a place copies the draft once to a temporary folder. The hash of
  each example, the run on H0 and the copy to `goals/<G>/` all read that copy,
  so `goals/<G>/` holds the bytes whose hashes the event carries, even when
  the draft changes while the Probe runs. Committed does not apply to a draft,
  which is in no commit.
- A place with a draft also writes each example it placed to `goals/<G>/`,
  replacing an earlier version of that example's folder; the owner commits it.

## Line 3: Own path names a folder of no Goal

- A runner's change that adds or changes a file under `goals/X/`, where X is
  no Goal of the store (and not `model`), is refused by Own path with the code
  `no-goal-folder`. A change that only deletes such files passes. A path under
  a Goal of the store that is not the runner's own is still `other-goal-path`.

## Line 4: another Goal's folder is unchanged

- The refusal for a path under another Goal of the store is as before.

## Line 5: the grill skill

- `skills/hyp-grill` tells the grill to draft examples in a named draft with
  `hyp try --draft`, and names no `goals/draft-` folder.
