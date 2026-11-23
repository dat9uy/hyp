# The Goal's model

The model part of hyp-grill, step 10 of "How to grill". Read it when the change
moves through states and you write the Goal's model with the hyp-statechart
skill: a TLA+ spec in `goals/<G>/specs/`, one variant per potential Hypothesis,
and the chart. The statechart tool, not you, says which rules each variant
breaks and which examples each variant would meet.

- Draft the model after the owner consents to the words and lines and
  before the examples are final, so that the variants the tool reports can
  still give the examples their direction.
- The hyp-statechart skill owns how to run the tool, how to try each
  variant as a patch with `hyp try G LINE/EXAMPLE --worktree`, and what
  its part of the hand-off holds. Follow it there; this step adds only
  what you ask the owner.
- Before the examples exist, each variant that no example separates is a
  reading of a line ("How to grill the examples", step 1).
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
no chart; a spec folder with no model is selected as before.
