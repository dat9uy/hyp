Review of probe-closes (709e453) against G-908PRZ#1 to #5 and the claim. Flags checked: `--limit`, `--closes`, `--evidence`, `--aims`, `hyp checkpoint --refers` all exist; skillcheck sees no new command. Findings are about meaning and fit, not syntax.

F1. Label: G-908PRZ#4
Text: "A case the probe runs and the candidate passes is closed in the review's Result ... record no `--limit` for it and commit no fix."
Finding: "closed" names no command. In hyp, `--closes` closes an existing key (`<R-id>#<n>` or `G-<id>#<n>`). A review finding has no key. Two meanings: (a) a prose sentence in the observation, (b) `--closes G-908PRZ#n`. Meaning (a) breaks the rule in hyp-prototype and in this skill's step 4 that a claim a case is covered or void "is a limit and a refer, never a sentence in an observation". The log then holds no typed trace that the case ran, and the next reader sees nothing. Meaning (b) closes the whole Goal line for the candidate, not the one finding (see F2).
Case: the reviewer says "the text lets a run stop before the review; case of #4". The run replays one stand-in, it passes, and the run writes "closed by probe" in the observation. hyp status shows no closed item; the stop checkpoint has nothing to name; the owner cannot tell the case was settled.

F2. Label: G-908PRZ#1
Text: same sentence, read as `--closes` on a line.
Finding: A Result that closes G-908PRZ#n covers the candidate it gave a verdict on, so one passing probe of a narrow finding would mark the whole line (for example, "3 of 3 trials, cut A") as run. A line's case is the line, not the finding. The Result would also need an Experiment that `--aims` at the line, and the review Experiment's rule is written before the review, so it does not aim there. hyp refuses the close (replay-spec refusal 22) or the run closes a line that no replay ran.
Case: finding "the stop checkpoint may omit a candidate" (case of #4). The run does a text grep probe that passes, and records `--closes G-908PRZ#4` on the review Result. #4 now reads covered for the candidate with no trial at cut A or B.

F3. Label: G-908PRZ#4
Text: "Probe each finding ... with a check on the candidate before you record the review's Result."
Finding: This conflicts with "Work": "Try a new or changed check only on its cases' stand-ins, never on a candidate, until the rule is recorded", and with ADR 0015 item 3 (a check built by the judge after seeing the candidate). The probe is a new check, written after the finding, run on the candidate, and the run both writes and judges it. The text names no step that records a verdict rule for the probe, nor which Experiment holds it. A weak probe passes and the case leaves the ledger with no limit and no refer, so the select is not held.
Case: finding "the skill gives two meanings for 'stop clause'". The run writes a probe that greps for the phrase, it passes, and the case closes. The meaning problem stays and no owner or later run sees it.

F4. Label: G-908PRZ#1
Text: same instruction, with the budget of N Results.
Finding: Each probe that is recorded needs an Experiment and a Result under the run's actor, and the review Result comes after all probes. A candidate with k findings takes up to k+1 Results before its review Result exists. With default N 5 (build, gate, probes, review) the budget can end before the review Result is recorded. The text gives no cap on probes and no rule to record the review first. The old order recorded the review, then limits.
Case: cut A, N=5: Results used: gate (1), review subagent Result is deferred, 3 findings are probed (2 to 4 and a probe batch), budget ends. The run stops with a candidate with no review Result, and the checkpoint may not name it. #1 and #4 fail.

F5. Label: G-908PRZ#3
Text: same instruction.
Finding: A probe on skill text is a replay or an eval case run, and the Goal's own method for skill text ("run the original and the candidate skills on the same eval cases") is costly. "A check" has no bound on cost. Several probes per candidate add claude -p runs to every trial, and the median cost against the base can exceed 1.5 times. The base had no probe step.
Case: the review gives 4 findings labeled #2. The run replays a stand-in for each (4 runs on Sonnet) before it records the review. Total cost passes 1.5 times the base median.

F6. Label: G-908PRZ#1
Text: "a case the candidate fails is a fix commit" and, in the same bullet, "a defect in text you may edit is a fix commit".
Finding: The claim says no fix follows a passing case, but a failing case now always gets a fix commit. Each fix commit is a new Hypothesis and a new candidate, and "Before the stop ... one review of each candidate commit" makes it owe its own review, which can produce more findings and probes: a chain of reviews. The Goal wants no chain on small changes. Before this change a failing case could be a closed limit (step 4: "a candidate that fails the case is contradicted and the limit is still closed"). The new sentence also differs from step 4 and from "Goal lines": a failed case there is a verdict, not a fix.
Case: the probe fails on a one-word wording case. The run commits a fix, supersedes, reviews the new commit; that review finds a new case, which is probed, which fails again. Budget ends with the last candidate unreviewed.

F7. Label: G-908PRZ#4
Text: "No fix commit follows a passing case" against "a defect in text you may edit is a fix commit".
Finding: One finding can be both a case of a line and a defect in text. The new sentence says a passing probe means no fix, but the probe tests a case, not the wording. The two sentences give opposite orders for the same finding, and the text does not say which wins. The defect stays in the shipped skill.
Case: reviewer: "step 3 says 'a Result covers' and 'a Result gave a verdict' with two meanings; case of #2". The probe replays the cut and passes. By the new rule, no fix; by the old rule, fix. The run picks one.

F8. Label: none
Text: "A case the probe runs and the candidate passes ... cite the probe's log there" and "Sort the rest: a case no check ran is a `--limit`".
Finding: "the probe" and "a check" are not the same noun; "the probe" is never introduced (the claim says "the run's own probe"). A pre-existing check, such as `go test` in the gate that already ran, can read as "a check ran it" even though it did not test the case. Then every case that sits in a file the gate loads would close.
Case: finding "the refusal text for X is missing" with the gate green. The run says the gate ran it and cites go-test.log. No test touched that refusal.

F9. Label: none
Text: "cite the probe's log there" with `--evidence`.
Finding: `--evidence` takes at most 10 pointers (each at most 300 runes). The review Result already cites the review and the gate logs; a review with several probes can pass 10, and the text gives no way to cite more. hyp refuses the Result, so the review Result is not recorded.
Case: 1 review file, 1 gate log, 9 probe logs gives 11 pointers; `hyp record` refuses; the candidate has no review Result at the stop.

F10. Label: none
Text elsewhere: hyp-prototype "Record" ("A gap you know at select time is a `--limit` on a Result, never only words in `--why`"; "A claim that a case is ... covered elsewhere is a limit and a refer, never a sentence"), and README line 110 ("A Result records the cases that its run did not cover. A later Result closes a limit when it runs the case").
Finding: The change introduces a closing of a case with no limit key, which no other skill or doc describes. hyp-goal-run now disagrees with hyp-prototype on how a covered case is shown. No text outside the changed bullet is made false, but the two skills now give the owner two different traces.
Case: a run reading both skills may record a limit for a passing case (hyp-prototype) and close it, or record none (hyp-goal-run); the ledger differs by which skill the run read last.

Summary: the change adds a probe step with no cap, no key to close and no pre-registered rule, so a run can close a case with prose, spend its budget before the review Result, and leave a candidate unreviewed (G-908PRZ#1, #4) at higher cost (#3); the "fail gives a fix commit" half can start the very chain of reviews the Goal wants to avoid.
