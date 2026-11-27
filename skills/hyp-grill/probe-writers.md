# Probe writers

The Probe-writer part of hyp-grill, step 6 of [examples.md](examples.md).
Read it when an example's reading needs a state set up before the command (for
example "in a git repository with three commits"), so the generic Probe cannot
run it. Then:

- Write the `example.txtar` in the draft. Its `run` stays the command the
  owner would type, word for word; the setup goes in the Probe, never in
  the `run`.
- Tell the writer that each setup act must succeed: the Probe checks its exit
  and exits 3 when it fails, so a refused setup shows as an environment
  problem on H0, not as the failed verdict a gap example needs. This holds for
  a Probe nested inside the setup too.
- Write no Probe code for it yourself. Start a Probe writer: an agent
  whose prompt is the probe-writer block that
  `hyp status G --role probe-writer` prints, copied word for word, followed
  by the example's key (for example `line-2/a`), its words and the setup it
  needs. Before the owner adds the Goal, print the block of any Goal of the
  store, and name the example's folder in the draft, `hyp draft path NAME`
  and `/line-2/a`, in place of `goals/<G>/line-2/a`. The draft is in no
  commit, so a Probe writer needs no worktree and no branch: it writes only
  the `probe/` folder of its own example in the draft, and writers of
  different examples run in parallel with no conflict. Tell each writer its
  example folder and to touch nothing else.
- The Probe writer returns the output of `hyp try --draft NAME --base COMMIT
  LINE/EXAMPLE`. Do step 7 of [examples.md](examples.md) yourself on the
  finished draft. Start no Probe writer for an example the generic Probe can
  run.
