# Review of whole-candidate (edb95c6) against G-908PRZ#1 to #5

The commit changes one bullet of skills/hyp-goal-run/SKILL.md ("Before the stop ... each candidate commit's diff" becomes "Before the select ... the candidate's whole diff from the Goal's base commit"). `hyp subject --base`, `hyp supersede`, `hyp checkpoint` and the other commands the file names exist in `hyp --help`. The changed line names no new command or flag. No text in docs/, README.md or AGENTS.md cites the old wording.

F1. Label: G-908PRZ#1 (also #4)
Text: "One review of the candidate you select replaces a review of each fix commit."
Finding: Line #1 requires every candidate the run built to have a review Result. A fix is a new commit and a new Hypothesis, so each superseded commit is a candidate the run built. The text now says those commits need no review. This contradicts #1 directly, and it removes the instruction that kept #4 from firing.
Case: The run builds C1, C2 (supersedes C1), C3 (supersedes C2). It reviews only C3 and selects it. C1 and C2 have no review Result, so #1 fails in that trial.

F2. Label: G-908PRZ#4 (also #1)
Text: "Before the select, ask a fresh subagent ..." together with the Stop section.
Finding: The review is tied to the select event. A select happens only when OWED lists `select`. A run that ends on its budget, on an open reading or limit, or on a stop clause with no select never reaches the review step. The old text said "before the stop", which covered this path. The Stop section has no review requirement and does not say to name an unreviewed candidate in the checkpoint, which #4 allows as the only exception.
Case: Cut A, N=5. The run spends its 5 Results on Experiments and limits, owes a limit, and stops with the checkpoint "Stopped: N Results". No select was due, so no review was asked. A candidate stands with no review Result and the checkpoint does not name it. #4 fails.

F3. Label: G-908PRZ#2 (also #1)
Text: "ask a fresh subagent for one review" (no Result is mentioned).
Finding: Lines #1, #2 and #4 count a "review Result". The skill never says to record the subagent's review as a Result, under which Experiment, or that it spends one of the N budget Results. The change does not fix this gap. It also gives no order for budget: the select needs OWED to list `select`, and the review needs a Result, but nothing reserves a Result for the review.
Case: Cut B, N=5. The run uses Results 1 to 5 on candidate Experiments and closes the last limit with Result 5. OWED now lists `select`, but the budget is spent. The run selects with a chat-only review, or skips the review. #2 ("a review Result before its budget ends") fails.

F4. Label: G-908PRZ#4 (also #1)
Text: "a defect in text you may edit is a fix commit" (unchanged), now read with "one review of the candidate you select".
Finding: The review comes just before the select, and its findings can produce a fix commit. That fix is a new candidate and a new Hypothesis, and the review no longer covers it. The Stop section says "fix it with a commit before the stop", which adds one more unreviewed commit after the review. Nothing says to review again, or to name the candidate in the checkpoint. The aim of one review per run is not reached, and the final candidate can end unreviewed.
Case: Review of C3 finds a defect in the skill text. The run commits C4 and stops. C4 has no review Result and the checkpoint does not name it. #4 fails.

F5. Label: none
Text: "from the Goal's base commit (`git diff <base> <candidate>`)".
Finding: A Goal has no base commit. The base is set per Hypothesis with `hyp subject <H> --base <commit>`, and the run may use different bases. "The Goal's base" has two meanings: the first candidate's base, or the candidate's own subject base. A different base gives a different cumulative diff.
Case: A fix is built on a moved base, or the pushed branch is rebased. The run diffs from the first Hypothesis's base, and the reviewer sees changes that are not the candidate's own.

F6. Label: none
Text: "not of the last fix commit alone. One review ... replaces a review of each fix commit."
Finding: The second sentence can be read as an order ("skip the reviews of the fix commits") or as a description of scope. Under the first reading the run reviews only the final candidate. That is F1's break of #1. Under the second reading it still owes a review per candidate. The line should say which.
Case: One run reads "replaces" as an instruction and reviews once. Another reads it as scope and reviews each fix. The two trials give different review counts and costs.

F7. Label: G-908PRZ#3
Text: "whole diff from the Goal's base commit".
Finding: The review input grows with every fix, because it is the cumulative diff. The text sets no limit on the size of the diff or on the number of reviews of it. If F4 makes the run review again after each fix, each review is larger than the last one the base run asked for, which raises cost. Cost against the 1.5 times bound is not tested by the text. The run's own text does not break #3, but it gives no guard.
Case: A run with 4 fix commits that reviews after each one pays 4 whole-diff reviews. The median cost may be over 1.5 times the base's.

F8. Label: G-908PRZ#5
Finding: The change is one bullet with no new frontmatter. Nothing in it should fail `quick_validate.py`, `lint_cruft.py --routing` or skillcheck (no new `hyp` command is named). This review did not run those checks.

Summary: The change tells the run to review once, before the select, but it drops reviews of superseded candidates and of runs that stop before a select, never says to record the review as a Result, and leaves an unreviewed fix after the review, so the run can still break G-908PRZ#1, #2 and #4.
