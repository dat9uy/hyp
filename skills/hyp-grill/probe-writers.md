# Probe writers

The Probe-writer part of hyp-grill, step 6 of "How to grill the examples".
Read it when an example's reading needs a state set up before the command (for
example "in a git repository with three commits"), so the generic Probe cannot
run it. Then:

- Write and commit the `example.txtar`. Its `run` stays the command the
  owner would type, word for word; the setup goes in the Probe, never in
  the `run`.
- Write no Probe code for it yourself. Start a Probe writer: an agent
  whose prompt is the probe-writer block that
  `hyp status G --role probe-writer` prints, copied word for word, followed
  by the example's key (for example `line-2/a`), its words and the setup it
  needs. Probe writers run in parallel, each in a worktree of its own,
  and the record branch stays checked out in the owner's checkout, so
  also tell it to commit on a branch made from the record branch before
  it writes, for `line-2/a`:
  `git checkout -b probe/<G>-line-2-a record/<G>`.
- The Probe writer commits the example's `probe/` folder and returns the
  commit and the output of `hyp try`. In the owner's checkout, bring each
  commit onto the record branch with `git cherry-pick`; each touches only
  its own folder, so the picks do not conflict. Then do step 7 of "How to grill the examples"
  there,
  because the writer's `hyp try` ran on its own branch. Start no Probe
  writer for an example the generic Probe can run.
