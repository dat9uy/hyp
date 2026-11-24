# Run rules

The run part of hyp-grill. A runner of a hyp Goal follows these rules. hyp
refuses what breaks its guards; these are the rules it cannot check.

## Start

A runner starts on an open Goal from this prompt. The grill prints it at its
hand-off with the Goal, the actor and the issues the Goal serves filled in, and
the owner copies it and adds nothing else: what the runner is told comes only
from the Goal and these rules, so two runners of one Goal are told the same.
`<issues>` lists each issue as `#N`, separated by commas.

```text
Run the hyp Goal <G> as its runner.
- Your actor is <actor>. Set HYP_ACTOR=<actor> and HYP_ROLE=runner each
  time you call hyp.
- Work in your own worktree, on a new branch from the default branch.
- Follow the run rules of the hyp-grill skill (run-rules.md) and this
  repository's agent files, such as AGENTS.md.
- The Goal's placed examples are the whole task: start from what
  `hyp status <G> --role runner` prints. No issue, PR or hand-off is a
  source of work.
- Open a PR whose body is the select's report. Below it, add a line
  `Closes #N` for each of these issues, so the merge closes them: <issues>.
  They are only the issues to close, not a source of work. Do not merge the
  PR: the owner merges.
```

## Rules

1. **Run only the lines hyp prints.** Read `hyp status G --role runner` and
   copy the run line it prints, with your commit filled in. Run the select line
   only when hyp prints it for you. When every line is met on your commit and
   hyp prints no select line, the commit usually fails the Spec guard, and
   `hyp status` does not say so. Check that the commit holds a spec under
   `goals/G/specs/`, and when that spec is a model, that it holds a chart and
   that its check config lists no rule the statechart tool printed as not
   shown to fail. Fix it in a new commit and run its lines.
2. **Check before you run.** `hyp try G LINE/EXAMPLE --worktree` shows an
   example's verdict on your uncommitted files. It records nothing and uses no
   run of the budget. The select needs a spec in the commit, so when the
   record branch holds none, write the Goal's model before the first commit you
   run. Follow the hyp-statechart skill: try each variant the tool
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
6. **Commit only what your acts wrote.** Add the event lines your own acts
   wrote under `.tree/events/`, each by name. hyp keeps Results and outputs
   in the clone's trace, so a run leaves nothing to commit.
7. **Select when hyp offers it.** When hyp offers one candidate, select it.
   You run a rival, a second commit with `hyp run` that uses runs of the
   budget, only after the owner unselects the Goal; a second candidate then
   stops the Goal for the owner. A variant tried with `hyp try --worktree` is
   not a rival.
8. **Stop for the owner when hyp stops.** At `Two candidates` or
   `Budget used`, stop and report the state block.
9. **The PR body is the select's report.** Paste it as hyp printed it; hyp
   writes nothing to GitHub. Below it, add only one `Closes #N` line for each
   issue the Start prompt names, and the diffs of rule 2, each with its
   variant and the outputs `hyp try` printed.
10. **A defect outside the Goal** is fixed only when the line needs it.
    Otherwise record it as a finding (rule 11) and list it in the PR for the
    owner.
11. **A fact that no line of its Goal needs is a finding.** Record it with
    `hyp finding add --fact TEXT --evidence TEXT`: the fact and a pointer to
    its evidence, such as a commit, a path or a Result id. Then read
    `hyp status`, which prints the words of each Goal in work, and
    `hyp finding list`. Fold the finding into the one Goal in work whose want
    it bears on, with the ready `hyp finding fold` line the list prints. When
    no Goal in work holds that want, draft an `add goal` for the owner, with
    its words, lines and why, and attach it with
    `--draft "hyp add goal ..."`; the owner may edit and run it. Never hold a
    fact only in an issue, a PR or a comment.
