# The Goal's model

The model part of hyp-grill, step 10 of "How to grill". Read it when the change
moves through states and you write the Goal's model with the hyp-statechart
skill: a TLA+ spec in `goals/<G>/specs/`, one variant per potential Hypothesis,
and the chart. The statechart tool, not you, says which rules each variant
breaks and which examples each variant would meet.

When the repository keeps a model of its machine in `goals/model/`, the
Goal's model is a change to that kept model, not a new whole model:

- Start from the kept module. Add or change only the steps, prints and rules
  the Goal's lines touch, and give each rule you add the source of its line
  as `G:line-N`.
- Keep the draft with the Goal. Write it as a copy of the kept module in
  `goals/<G>/specs/`, never in `goals/model/` on the record branch: the
  record PR merges before any select, and the kept model on a commit must
  describe that commit's code. The runner carries the change into
  `goals/model/` in the commit it selects.
- A repository with no kept model writes a Goal's own whole model in
  `goals/<G>/specs/`.

For every model:

- Draft it after the owner consents to the words and lines and before the
  examples are final, so that the variants the tool reports can still give
  the examples their direction.
- The hyp-statechart skill owns how to run the tool, how to try each
  variant as a patch with `hyp try G LINE/EXAMPLE --worktree`, and what
  its part of the hand-off holds. Follow it there; this step adds only
  what you ask the owner.
- Before the examples exist, each variant that no example separates is a
  reading of a line ([examples.md](examples.md), step 1).
- A variant reaches the owner as a question only when it changes what the
  owner sees: a print, a refusal or an exit code. A variant that differs only
  inside the build, or in free text that no act of hyp reads, is the runner's
  choice: name it in the hand-off and ask nothing.
- After the examples are placed, for each tried variant that meets every
  example and prints something else on a step no example takes, ask the
  owner one question in plain English, at the example level and with no
  model word: show the case, what each build prints for it, and ask which
  output is wanted.
- When the owner answers "either output is fine", list the case under
  "Not covered". Otherwise place an example that separates the two builds.
- Stop trying a variant at its first failed example. A variant that fails
  one example meets not every example, so its other tries teach nothing.

At select, hyp refuses a model that has a rule no variant config names, or
no chart; a spec folder with no model passes this check.
