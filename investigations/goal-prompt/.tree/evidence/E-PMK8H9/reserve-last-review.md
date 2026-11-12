Review of reserve-last (1265f98) against G-908PRZ#1 to #5. The diff adds one bullet to "Build Goals" in skills/hyp-goal-run/SKILL.md. No command or flag in the new text; `hyp prompt --experiments` and `hyp status` exist.

F1. Label: G-908PRZ#4
Text: "When one remains, do not commit another fix: ask for the review of the current candidate, record its Result ... and stop."
Finding: The rule assumes the current candidate has no review Result yet. It gives no instruction when the current candidate was already reviewed and one Result still remains (the review found a defect, which sends the run to the fix path). The text then reads as "review the same candidate again", a duplicate review, or as "stop with a known defect".
Case: N=5. Candidate B gets its gate Result (3) and review Result (4). The review finds a defect in the skill text. One Result remains. The run may not commit a fix; it asks for a second review of B (a chain of reviews on one candidate, against the Goal's aim), or it stops. The stop checkpoint must name nothing for G#4 because B has a review, but B ships with a known defect.

F2. Label: none (contradiction inside the skill)
Text: new bullet "do not commit another fix" vs the next bullet "a defect in text you may edit is a fix commit" and Stop 1 "A defect in text you changed and may edit is not [out-of-Goal]: fix it with a commit before the stop."
Finding: "record its Result with each finding sorted as below" sends a text-defect finding to "fix commit", while the new bullet and the claim forbid a commit. Two rules give opposite orders and the text names no winner and no fourth place for such a finding (not a limit, reading or out-of-Goal finding). A run may obey either.
Case: With one Result left the review finds a wrong command in the candidate's SKILL.md. One run reads Stop 1 and commits a fix (a candidate with no Result left, G#4 broken unless named). Another run obeys the new bullet and records nothing about it, losing the finding.

F3. Label: G-908PRZ#1
Text: "Count the Results left under your actor"; "When one remains".
Finding: Only the case "exactly one" is handled. The exact trigger also misses a run that reaches the last Result in another way: a Result that is not a fix commit (a reading, an apply-answer Result, a limit close, a gate Result of a new candidate) spends the reserved Result. Nothing says these must leave one Result for the review.
Case: N=5, two Results left, the run commits fix C (a new candidate). Its gate Result takes one; one remains: correct. But if instead the run uses the two for a limit-closing Result and a reading, then commits C at zero left. The "when one remains" test is never true at a commit, and C has no review Result at the stop.

F4. Label: G-908PRZ#1
Text: "ask for the review of the current candidate, record its Result" with "Before the stop, ask a fresh subagent for one review of each candidate commit's diff".
Finding: "each candidate" in the older bullet is not narrowed. With one Result left, a superseded earlier candidate with no review (its gate failed, so a fix replaced it) stays unreviewed. #1 requires "every candidate it built" to have a review Result. The new text reviews only "the current candidate" and a Result holds one candidate's review at best; the skill does not say whether one Result may record the reviews of two candidates.
Case: Candidate A fails its gate (Result 1), fix B supersedes it (gate Result 2, review Result 3, defect), fix C (gate Result 4). One left: review of C (5). A has no review Result and the run stops; #1 fails for that trial.

F5. Label: G-908PRZ#2
Text: "Count the Results left ... before the stop clause."
Finding: "before the stop clause" has two meanings: before the run reaches the stop (at each step), or at the point where the Stop section is read. There is also no way named to count: `hyp status` prints no Results-under-actor count and the prompt gives only N. A run that counts at the wrong time sees "one remains" after the commit.
Case: At cut B the run has 2 left, commits fix D, then counts: 2 left (gate and review still possible). It runs the gate, counts 1 left, asks for the review, which uses it, and stops without selecting: select needs OWED to list it and the run is out of Results. A run that miscounts by one (counts the prompt's N from the first Result of the tree, including other actors' Results) may stop one Result early or late.

F6. Label: G-908PRZ#4
Text: "A fix commit with no Result left for its review is a candidate the run cannot review."
Finding: The sentence states a consequence but gives no instruction for the case it describes. It does not say the commit is forbidden unless two Results remain (gate and review), though the owner-grill skill (line 116) says N has "room for a gate and a review per candidate". "Results left" is counted as one, but a new fix commit needs two (gate and review). So the threshold "one remains" allows a fix commit at two left only if the gate Result can be skipped.
Case: Two Results left, the run commits fix E. Its gate Result takes one; the review takes the other. Fine. But if the run commits at two left and the gate Experiment records two Results (a failed first run, then a rerun), no Result is left for the review and E is stopped unreviewed with no instruction to name it in the stop checkpoint.

F7. Label: none
Text: skills/hyp-owner-grill/SKILL.md line 116 and `hyp prompt` source say N defaults to 7; the installed `hyp prompt --help` says 5; the Goal line #3 says base default N 5.
Finding: Not caused by the change, but the new reserve rule depends on N; the skill says nothing of what N is, and the owner-grill text "room for a gate and a review per candidate" is now partly false: the run may leave a gate-less or review-less candidate. Check the installed binary against the source before a replay uses "the candidate's default N".
Case: A replay with the installed binary prints N=5 while the source prints 7; the trials count Results against different budgets.

F8. Label: G-908PRZ#5
Text: new bullet.
Finding: No defect found for #5: the bullet holds no wrapped inline code span, no hyp command and no ADR or docs citation; skillcheck should pass. Not run here (read-only task).
Case: none.

Summary: The bullet fixes the "fix with no room to review" case but conflicts with the fix-commit rule and Stop 1, covers only "exactly one left", leaves earlier unreviewed candidates and non-fix Results uncovered, and gives no way to count.
