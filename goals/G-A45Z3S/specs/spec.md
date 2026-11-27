# After Stable refuses a run on H0, every output hyp prints of H0's cell names the output the place saved

The spec of Goal G-A45Z3S. It has no model, and the kept model in
`goals/model/` does not change: no Goal moves between states on another event
than before, and the kept model has no step for what a refusal prints.

## Line 1: the refusal names H0's first output

- A write act reads the cells the clone's trace keeps as files, under the
  lock, as it reads the events. The Stable refusal of a run on H0 prints the
  first output as the bytes the owner's place saved in this clone's trace.

## Line 2: hyp status names it too

- `hyp status` lists an unstable cell of H0's row with the hash of the output
  the place saved, not `(unknown)`. Replay reads a Result before it has the
  cells, so giving the cells names that hash.

## Line 3: a commit other than H0 is unchanged

- The first output of a refused run on another commit is the live cell's
  output from the Results, as before.
