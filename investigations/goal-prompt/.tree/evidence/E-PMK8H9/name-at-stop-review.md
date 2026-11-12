Review of name-at-stop (2faee48) against G-908PRZ#1 to #5. Command check: `hyp checkpoint --goal --summary --why` all exist in the installed `hyp checkpoint --help`; no wrong command or flag.

F1. Label: G-908PRZ#4 (also #1, #2)
Text: "A candidate whose review did not run before the stop ..." (Build Goals) and "each candidate with no review Result" (Stop template).
Finding: The two sentences use two tests. The review bullet never says the subagent's review is recorded as a Result; it only says to ask for one and sort the findings. The new bullet says "review did not run"; the template says "no review Result". The lines count a "review Result". A run can read "review" as the subagent call, not a recorded Result.
Case: The run asks a subagent to review H-2, sorts the findings in chat, and records no Result. It lists H-2 as reviewed (or omits it), so the checkpoint is silent. The trial grader finds no review Result and the stop names no candidate: #4 breaks. The other reading lists H-2 as unreviewed and #1 fails even though a review ran.

F2. Label: G-908PRZ#4
Text: the Stop `--summary "... Candidates without a review: <each ... by Hypothesis id and commit, or none> ..."`.
Finding: `--summary` holds at most 600 runes (hyp checkpoint --help), and a longer one is refused. The template already holds "Stopped", "Waiting for the owner" (all AWAITING OWNER keys) and "Out-of-Goal findings". The new field adds about 90 fixed runes plus about 25 for each candidate (id plus commit). The text gives no short form and no rule for what to cut. The refusal is on the stop checkpoint itself, which the run must record last.
Case: At cut A the run built 6 Hypotheses with no review, and 4 readings wait under AWAITING OWNER. The summary passes 600 runes, hyp refuses it, and the run either trims the new field (the candidate goes unnamed, #4 breaks) or ends with no stop checkpoint.

F3. Label: G-908PRZ#4 (also #1)
Text: "each candidate with no review Result" and "A candidate whose review did not run".
Finding: "Candidate" is not tied to superseded commits. The skill makes each tested commit a Hypothesis, and a fix supersedes the old one. Line #1 says "every candidate it built". The review bullet says "each candidate commit's diff". A run can read "candidate" as the live or last one only, and skip superseded Hypotheses.
Case: H-1 (superseded by H-2) was never reviewed. The run reviews H-2, lists "none", and the stop names no gap. A grader counting every built candidate sees H-1 with no review Result and no name in the checkpoint: #4 breaks.

F4. Label: G-908PRZ#2 (also #1)
Text: the whole change. It names the gap and adds no step that makes the review happen or orders it.
Finding: The claim keeps the practice and adds only the name. Nothing stops a select of an unreviewed candidate, and nothing says to keep one Result of the budget for the final review. The review stays "Before the stop", with no link to the select or to the budget. The change fits #4 only; #1 and #2 rest on the unchanged bullet, which a run can still skip under a short budget.
Case: At cut B with N=5 the run spends 4 Results on a fix and a re-run, selects H-3 (OWED lists select) with no review Result, then stops and names H-3 in the checkpoint. #4 holds, but #2 fails: the select is not of a candidate with a review Result.

F5. Label: none
Text: Stop step 3 ("list every owner item with the exact command").
Finding: A named gap is not an owner item and has no command, so the final message may not repeat it. Step 2 says the last message repeats the checkpoint, which covers it. Low risk; the text is only less explicit than for the other fields.
Case: The run's last message lists only AWAITING OWNER items and the PR merge; the gap is in the checkpoint and in `hyp status` but not in the owner list.

F6. Label: none
Text: the Stop template, shared by every Goal; and skills/hyp-owner-answer/SKILL.md line 136, a checkpoint template without the field.
Finding: The field sits in a template that a discovery Goal also uses, where there is no candidate commit. "or none" covers it, but it is a build-only field in a general step. The hyp-owner-answer stop template has no such field. That run does not build, so this is not a contradiction.
Case: A discovery Goal run writes "Candidates without a review: none" in its checkpoint; no harm, only noise.

No finding on #3 (cost): the added bullet and field add no step or Result, so the cost change is a few output tokens. No finding on #5 from the diff: no frontmatter changed; the diff adds one bullet and one template field, and no `hyp` command name that skillcheck would fail on. I did not run quick_validate.py or lint_cruft.py.

Summary: the change names unreviewed candidates but leaves "review Result" undefined, can exceed the 600-rune summary cap, and does not touch the paths that break #1 and #2.
