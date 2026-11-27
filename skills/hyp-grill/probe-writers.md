# Probe writers

The Probe-writer part of hyp-grill, step 6 of [examples.md](examples.md).
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
  needs. Before the owner adds the Goal, print the block of any Goal of the
  store, and name the draft folder, `goals/draft-<name>/`, in place of
  `goals/<G>/`. Probe writers run in parallel, each in a worktree named by its
  example key, on a branch made from the record branch. Make each worktree
  yourself, one after another, before you start the writers, for `line-2/a`:
  `git worktree add -b probe/<G>-line-2-a <path>/probe-<G>-line-2-a record/<G>`.
  git can fail when two worktrees are made at the same moment, so do not let
  the writers or the runtime make them at once. Tell each writer its worktree
  path and to work only there.
- The Probe writer commits the example's `probe/` folder and returns the
  commit and the output of `hyp try`. In the record branch's worktree, bring each
  commit onto the record branch with `git cherry-pick`; each touches only
  its own folder, so the picks do not conflict. Then do step 7 of
  [examples.md](examples.md) there, because the writer's `hyp try` ran on its own
  branch. Remove each writer's worktree once its commit is picked. Start no Probe
  writer for an example the generic Probe can run.
