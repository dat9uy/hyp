# Review of raise-budget (d0e8591, base 2cea6c2)

Read: the diff; internal/cli/prompt.go and prompt_test.go at d0e8591; skills/hyp-goal-run/SKILL.md (unchanged); a search of skills/, docs/, AGENTS.md, README.md and internal/cli at d0e8591 for the default N. The only place that states the default is skills/hyp-owner-grill/SKILL.md line 116 (now "7 by default"); README.md:149, cli.go:86 and promptUse name the flag but not its value. No text that states 5 is left, except the fixture golden (F5).

F1. Label: G-908PRZ#4 (also #1, #2)
Text: skills/hyp-goal-run/SKILL.md, "Before the stop, ask a fresh subagent for one review of each candidate commit's diff"; prompt.go `--experiments` default 7.
Finding: the change only raises a number. SKILL.md is untouched, so nothing tells the run to keep budget for the review. The stop clause is "N Results recorded under the actor", and the run's own Results also count: the instrument, every limit-closing Experiment, every reading Result and every reassess Result. "7 = a gate and a review per candidate" holds only for a run with no other Results. The claim's reason lives in the owner-grill text, which the runner never reads.
Case: cut A, one candidate. The run records 1 instrument Result, 1 gate Result, 3 limit-closing Results and 2 reading/reassess Results (7). The prompt now says stop. The review was not yet asked and is not a Result. The run stops with a candidate that has no review Result, and its stop checkpoint names the 7-Result clause, not the candidate. That breaks #4 and fails #1.

F2. Label: G-908PRZ#1, #2
Text: owner-grill line 116, "room for a gate and a review per candidate".
Finding: the arithmetic is 2 Results per candidate, so 7 fits 3 candidates. A tree with 6 rivals (the run-review Goal has 6) needs 12. The default does not scale with candidates, and the text implies it does. "A review" is also not defined as a Result: the SKILL.md review is a subagent call, while the lines speak of a "review Result". If the review is not recorded as a Result, the stated room is wrong; if it is, SKILL.md does not say where it is recorded.
Case: a Goal with 4 candidates. 4 gates + 4 reviews = 8 > 7. The 7th Result ends the run during candidate 4's review, so cut B ends with no selected candidate that has a review Result.

F3. Label: G-908PRZ#3
Text: default N 5 -> 7, with no change to the skill's stop or review text.
Finding: the cost line compares the candidate to a base at N 5. A run that used to stop at its 5th Result can now run 2 more Results before the stop clause fires. Gates are `go test` runs and each Result adds turns. A run that does not stop at the owes-nothing clause grows in cost with N and nothing in the change limits it. The base is the sole cheaper arm, so a rise of about 40% in Results (7/5) plus the review subagent can reach 1.5x median cost.
Case: at cut A, the base run stops at Result 5 for 1.0 unit of cost. The candidate uses 7 Results and one review, for 1.6 units or more. The median of 3 is then above 1.5x the base.

F4. Label: G-908PRZ#4 (the Goal's "without a chain of reviews on small changes")
Text: SKILL.md "ask a fresh subagent for one review of each candidate commit"; owner-grill "a review per candidate".
Finding: a fix commit after a review (SKILL.md: "a defect in text you changed ... fix it with a commit before the stop") is a new candidate commit by this wording. Raising N gives room for it to be gated and reviewed again, and no text says a small fix is exempt. "Each candidate commit" has two meanings: each commit the run builds, or each candidate under a Question. The change makes the room for a chain, not the end of a chain.
Case: the review of candidate C1 finds a typo. The run commits C2, runs its gate (Result 5) and asks a second review (Result 6) and a third on C3 (Result 7). The chain on small changes the Goal wants to avoid is now inside the budget.

F5. Label: none
Text: internal/cli/prompt_test.go, TestPromptForTheCheckpointStoreFixture (args now add `--experiments 5`) and the comment "Run 6's prompt named 5 Results".
Finding: not a defect. The fixture still pins the bytes of run 6, and the explicit flag is the right way to keep the golden file. But the default path is now pinned by one test only (TestPromptPrintsTheGoalsRunPrompt, "or 7 Results"), and no test pins that the owner-grill text and the flag default agree. The text and the code can drift again without a failing test; skillcheck checks command names, not numbers.
Case: a later change sets the default to 6 in prompt.go and the test; owner-grill still says "7 by default", and `go test` passes.

F6. Label: none
Text: owner-grill line 116, "N Results under the actor, 7 by default, room for a gate and a review per candidate".
Finding: the line is a very long single sentence with a parenthesis that now has three items. "Room for" reads as a promise to the owner that the run will review each candidate; it is only a count. The claim says "for example 7", but the text states 7 as the rule and gives no rule for choosing another N when candidates are more than 3.
Case: the owner reads the line, expects a review of every candidate, and runs `hyp prompt` for a 5-rival Goal with the default. See F2.

F7. Label: G-908PRZ#5
Text: skills/hyp-owner-grill/SKILL.md is the only SKILL.md changed.
Finding: no defect found by reading. The front matter and description are unchanged, and the added words do not cite an ADR or a docs file. The 0 errors / no High finding result must still be measured with the two named scripts; I did not run them. hyp-goal-run is unchanged, so replays of hyp-goal-run at the candidate use the same SKILL.md hash as the base (cb15e2f2...), and the only difference between arms is N.
Case: none found; a replay will differ only by the N in the prompt.

Summary: raise-budget changes only a number, so a run that spends Results on limits and readings can still stop at N with its last candidate unreviewed (G-908PRZ#4, #1), costs more by up to 7/5 (G-908PRZ#3), and gets room for a review chain on small fixes.
