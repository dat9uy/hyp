# Run rules

The run part of hyp-grill. A runner of a hyp Goal follows these rules. hyp
refuses what breaks its guards; these are the rules it cannot check.

## Start

A runner starts on an open Goal from this prompt. The owner fills in the Goal
and the actor, and adds nothing else: what the runner is told comes only from
the Goal and these rules, so two runners of one Goal are told the same.

```text
Run the hyp Goal <G> as its runner.
- Your actor is <actor>. Set HYP_ACTOR=<actor> each time you call hyp.
- Work in your own worktree, on a new branch from the default branch.
- Follow the run rules of the hyp-grill skill (run-rules.md) and this
  repository's agent files, such as AGENTS.md.
- The Goal's placed examples are the whole task: start from what
  `hyp status <G> --role runner` prints. No issue, PR or hand-off is a
  source of work.
- Open a PR whose body is the select's report. Do not merge it: the owner
  merges.
```

## Rules

1. **Run only the lines hyp prints.** Read `hyp status G --role runner` and
   copy the run line it prints, with your commit filled in. Run the select line
   only when hyp prints it for you.
2. **Check before you run.** `hyp try G LINE/EXAMPLE --worktree` shows an
   example's verdict on your uncommitted files. It records nothing and uses no
   run of the budget. When the record branch holds no spec and you write the
   Goal's model, follow the hyp-statechart skill: try each variant the tool
   predicts to meet every example as a patch this way, and report each
   potential Hypothesis that meets every example with the outputs `hyp try`
   printed, never from the model alone. Its patch
   goes in the PR body as a diff (rule 9).
3. **A refusal names its guard.** Read the guard's name and fix its cause. Never
   run again to get a different verdict.
4. **Keep every commit you ran.** The budget counts Hypotheses: each distinct
   subject you ran, other than H0. It cannot see a failed path you hide. Do not
   rewrite away a commit you ran.
5. **Leave the examples as the grill placed them.** An example that seems wrong
   goes to the owner; it is not yours to edit.
6. **Commit only what your acts wrote.** Add the event lines and output files
   your own acts wrote, each by name.
7. **Select when hyp offers it.** When hyp offers one candidate, select it.
   You run a rival, a second commit with `hyp run` that uses runs of the
   budget, only after the owner unselects the Goal; a second candidate then
   stops the Goal for the owner. A variant tried with `hyp try --worktree` is
   not a rival.
8. **Stop for the owner when hyp stops.** At `Two candidates` or
   `Budget used`, stop and report the state block.
9. **The PR body is the select's report.** Paste it as hyp printed it; hyp
   writes nothing to GitHub. Below it, add only the diffs of rule 2, each
   with its variant and the outputs `hyp try` printed.
10. **A defect outside the Goal** is fixed only when the line needs it.
    Otherwise list it in the PR for the owner.
