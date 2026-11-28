# Run rules

The run part of hyp-grill. A runner of a hyp Goal follows these rules. hyp
refuses what breaks its guards; these are the rules it cannot check.

## Start

A runner starts on an open Goal from its Start prompt. `hyp status G` prints
the Start prompt of an open Goal, ready to copy, with the Goal and the
runner's actor filled in. The owner copies it and adds nothing else: what the
runner is told comes only from the Goal and these rules, so two runners of one
Goal are told the same.

## Rules

1. **Run only the lines hyp prints.** Read `hyp status G --role runner` and
   copy the run line it prints, with your commit filled in. Your runs write
   only to this clone's trace, never to git. When every line is met on your
   commit, `hyp status G` shows the Goal `selected` on it, calculated from
   this clone's Results. When every line is met and hyp status does not show
   the Goal selected, it lists the commit under `held back`, with the guard
   that refuses it and its fix. Read the guard and make the fix that hyp
   status names in a new commit, then run its lines.
2. **Check before you run.** `hyp try G LINE/EXAMPLE --worktree` shows an
   example's verdict on your uncommitted files. It records nothing and uses no
   run of the budget. hyp shows a Goal selected only on a commit that holds
   a spec, so when the record branch holds none, write the Goal's model
   before the first commit you run. When the repository keeps a model in `goals/model/`, change it in your
   commit, and carry into it the draft the grill left in `goals/<G>/specs/`;
   keep no `.tla` file in the Goal's folder then. Follow the hyp-statechart
   skill: try each variant the tool predicts to meet every example as a patch
   this way, and report each potential Hypothesis that meets every example
   with the outputs `hyp try` printed, never from the model alone. Its patch
   goes in the PR body as a diff (rule 9).
3. **A refusal names its guard.** Read the guard's name and fix its cause. Never
   run again to get a different verdict.
4. **Keep every commit you ran.** The budget counts Hypotheses: each distinct
   subject you ran, other than H0. It cannot see a failed path you hide. Do not
   rewrite away a commit you ran. Keep work you have not committed in your
   worktree or in a commit, never in `git stash`: every worktree of a clone
   shares one stash list, so a pop can take another session's stash, and a
   lost stash is found only by hand.
   Each commit of your own change writes a path under `goals/<G>/`, such as
   its spec, and none under `.tree/`. hyp reads a commit below your run that
   writes nothing there as the base, not as your change, and measures every
   example on it. So a commit of code alone, with the spec in a later commit,
   makes `Same base` refuse the run. Put the code and the spec in one commit,
   or write under `goals/<G>/` in each commit.
5. **Leave the examples as the grill placed them.** An example that seems wrong
   goes to the owner; it is not yours to edit.
6. **Commit no event.** Git holds only the owner's acts. hyp keeps your
   Results and outputs in the clone's trace, so a run leaves nothing under
   `.tree/` to commit.
7. **Open the PR when hyp status shows the Goal selected.** When
   `hyp status G` shows the Goal `selected` on your commit, push the branch
   and open the PR. You run a rival, a second commit with `hyp run` that uses
   runs of the budget, only after the owner unselects that commit; a second
   candidate then stops the Goal for the owner. A variant tried with
   `hyp try --worktree` is not a rival.
8. **Stop for the owner when hyp stops.** At `Two candidates` or
   `Budget used`, stop and report the state block.
9. **The PR body is the report hyp status prints.** Paste what
   `hyp status G` prints under `PR body:` while the Goal is selected, as hyp
   printed it; hyp writes nothing to GitHub. Below it, add only the diffs of rule 2, each
   with its variant and the outputs `hyp try` printed.
10. **A defect outside the Goal** is fixed only when the line needs it.
    Otherwise record it as a finding (rule 11) and list it in the PR for the
    owner. A partial fix, or a part of a line's want that your build leaves
    out, is a finding too: record it before you open the PR and list its id
    there. A PR's notes alone do not hold it, because a PR is read once, at
    the merge.
11. **A fact that no line of its Goal needs is a finding.** Record it with
    `hyp finding add --fact TEXT --evidence TEXT`: the fact and a pointer to
    its evidence, such as a commit, a path or a Result id. Then read
    `hyp status`, which prints the words of each Goal in work. A Goal names
    the findings it is about: the owner with `--finding F` in `hyp add goal`,
    the grill in the `findings` section of an example. When a Goal in work
    holds the want the finding bears on, list the finding's id in the PR for
    the owner and the grill of that Goal. When no Goal in work holds that
    want, draft an `add goal` for the owner, with its words, lines and why,
    and attach it with `--draft "hyp add goal ..."`; the owner may edit it,
    add `--finding F` with the finding's id, and run it. Never hold a fact
    only in an issue, a PR or a comment.
