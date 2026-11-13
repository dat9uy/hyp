Issue #113: a /goal prompt that hyp makes from the tree, and the run's constraints.

- Where the new Goal's store goes: at first you lean toward one shared store at the repository root that would hold this Goal and every later Goal, because one store may now hold many Goals. If the grill shows a concrete thing that a root store breaks, you accept the place it proposes instead, and you want the root-store idea kept as a separate issue.
- How the change is tested: the test cases, the fixtures, which earlier run to compare with, how clauses are matched, held-out cases and similar details of an eval. You expect the grill to use the ak-skill-creator skill's eval and audit flow for that, because that skill already says how to do it. You do not want to be asked to choose among those details. If the grill asks you to choose one, say so, ask it to read that flow and decide from it, and do not pick an option. Your own preferences on testing are only these: test agents run on Sonnet, 3 trials are enough, and the budget stays modest.
- A Goal body without the run's constraint headings: the prompt command refuses it with exit 1 and names what is missing. Adding such a Goal stays allowed.
- Skill descriptions that name only one kind of Goal: no line for it; it is text the change makes wrong, under May change.
- Who writes the /goal prompt: hyp, as a command that reads the log. You are content to settle this in the session instead of testing it.
- Whether the prompt carries the owner's answers: no. The runner reads OWED. Settle it in the session too.
- The shape of the constraints block: the strict form, one constraint per line.
- Matching an earlier run's hand-written prompt clause for clause: widen the exceptions (the actor, the owner's answer, branch, PR and commit, and the run's own plan). Branch and PR become flags.
- A cap on the run's Experiments: a flag with a default of 5.
