# Run rules

The run part of hyp-grill. A runner of a hyp Goal follows these rules. hyp
refuses what breaks its guards; these are the rules it cannot check.

1. **Run only the lines hyp prints.** Read `hyp status G --role runner` and
   copy the run line it prints, with your commit filled in. Run the select line
   only when hyp prints it for you.
2. **Check before you run.** `hyp try G LINE/EXAMPLE --worktree` shows an
   example's verdict on your uncommitted files. It records nothing and uses no
   run of the budget. When the record branch holds no spec and you write the
   Goal's model, follow the hyp-statechart skill: try each switch as a patch
   this way, and report each wrong build that passes every example with the
   outputs `hyp try` printed, never from the model alone.
3. **A refusal names its guard.** Read the guard's name and fix its cause. Never
   run again to get a different verdict.
4. **Keep every tree you ran.** The budget counts trees, and it cannot see a
   failed path you hide. Do not rewrite away a commit you ran.
5. **Leave the examples as the grill placed them.** An example that seems wrong
   goes to the owner; it is not yours to edit.
6. **Commit only what your acts wrote.** Add the event lines and output files
   your own acts wrote, each by name.
7. **Select when hyp offers it.** When hyp offers one candidate, select it,
   unless the hand-off asks you to look for rivals; then run each rival the
   hand-off names, and a second candidate stops the Goal for the owner. A rival
   is a second commit you run with `hyp run`: it uses runs of the budget. A
   hand-off that asks for no rivals holds back only that, never a switch tried
   with `hyp try --worktree`.
8. **Stop for the owner when hyp stops.** At `Two candidates` or
   `Budget used`, stop and report the state block.
9. **The PR body is the select's report.** Paste it as hyp printed it; hyp
   writes nothing to GitHub.
10. **A defect outside the Goal** is fixed only when the line needs it.
    Otherwise list it in the PR for the owner.
