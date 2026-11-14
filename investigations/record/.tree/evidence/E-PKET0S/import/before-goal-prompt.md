# G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and whether it carries owner answers — complete

Design record for issue #113, settled by the owner in the grill session of 2026-11-11 (hyp-owner-grill). Two questions: who writes the /goal prompt for a Goal, hyp or a skill; and whether the prompt carries the owner's answers. Each is a DesignQuestion with its rivals, decided in that session with the session as evidence, so the tree says it was argued, not tested. This Goal has no lines and no run: no run checks its candidates. Its cases are the lines of the build Goal in this store that cites its selects.

## Q-NCY8HY Who writes the /goal prompt for a Goal: hyp from the log, or a skill? — answered

Issue #113: runs 1 to 6 of checkpoint-store:G-D1CH1R had hand-written /goal prompts. Scenario that separates the rivals: generate the prompt twice from a fixture of checkpoint-store:G-D1CH1R's log cut at the owner answer of 2026-11-11 14:31, and once in a store with a second Goal, then compare the outputs byte for byte and with run 6's prompt clause for clause.

Selected: H-MCDPNB
Design: A read command derives the prompt from the log: store path, actor, Goal id, the hyp-goal-run skill, the stop clauses and the Goal body's constraints block. hyp-owner-grill's hand-off and hyp-owner-answer's end call it. Same log and flags, same bytes.

- H-MCDPNB hyp prints it: hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N] — supported
  - E-VJPMHG The owner's scenarios in the grill session decide who writes the prompt — run (recipe none, rule by reading)
    - R-FCEZY6 recorded H-MCDPNB=supports H-E5K8VG=contradicts: S1 and S2 as argued in the session: a read command over the log gives the same bytes for the same log and flags and can refuse a missing heading; agent-written text cannot. The owner chose h-cmd.
- H-E5K8VG A skill writes it: hyp-owner-grill and hyp-owner-answer compose the prompt from hyp status and hyp show — contradicted

## Q-MAKJ8R Does the /goal prompt carry the owner's answers? — answered

Run 6's prompt opened with the owner's answer it started from. Scenarios that separate the rivals: (S1) hyp-owner-answer records an answer, then its own runner checkpoint, then the prompt is made; (S2) a run writes a refer-only checkpoint after an answer and dies before closing the returned limit; (S3) one answer event settles a key of each of two Goals; (S4) at the 14:31 cut of checkpoint-store:G-D1CH1R, OWED already lists the returned limit checkpoint-store:R-9SMRB2#4 for the runner.

Selected: H-RSDX0H
Design: The prompt holds no owner answer text and is byte-identical before and after an answer event.

- H-RSDX0H No answers clause: the runner reads OWED, which already lists each returned limit and each answered reading to apply — supported
  - E-T7ZS23 The owner's scenarios in the grill session decide whether the prompt carries answers — run (recipe none, rule by reading)
    - R-26GPDW recorded H-RSDX0H=supports H-9Y4F4H=contradicts H-CHKW8Y=contradicts: S1 to S4 as argued in the session: position-based answers vanish after hyp-owner-answer's own checkpoint; state-based answers repeat what OWED already lists. The owner chose h-none.
- H-9Y4F4H Answers from state: each returned, unclosed limit and each unapplied answered reading of this Goal — contradicted
- H-CHKW8Y Answers by position: the answer events after the Goal's latest runner checkpoint — contradicted

# G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete

Builds the design recorded in G-GZR6C6 (selects H-MCDPNB and H-RSDX0H, with grill session 285fa63e as evidence), for issue #113. The candidate is a hyp commit: the read command hyp prompt and the skill text it makes wrong. go test on the candidate decides lines 1 to 5; line 6 is an eval case in ak-skill-creator's format; line 7 runs its pinned scripts. A skill change follows ak-skill-creator's update flow: matched original and candidate consumer runs on the cases its edited lines touch. The fixture copies run 6's prompt (plans/reports/goal-prompt-261111-2129-g1b-run-6.md) and checkpoint-store:G-D1CH1R's events into testdata, because plans/ and investigations/ are not testdata.
May change: this repository, in commits on the branch the prompt names, that follow AGENTS.md (conventional messages, no AI trailers); the run may push that branch and open or update one PR, and never merges it. It changes cmd/, internal/ and their testdata, and the text the change makes wrong: docs/replay-spec.md, docs/glossary.md, docs/principles.md, AGENTS.md, README.md, and skills/ (hyp-owner-grill's hand-off and Goal-body template, hyp-owner-answer's end, hyp-goal-run's practice from run 6, and the descriptions of hyp-goal-run and hyp-prototype).
Read only: plans/, every investigations/ tree except goal-prompt, and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or ~/.local/share/hyp/hyp-33896ec; compile the candidate to a separate path.
Never: edit recipes/, rubrics/, another investigations/ tree, ak-skill-creator or another repository.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.

## Q-Z93VJ5 Which hyp commit implements hyp prompt and the skill text it makes wrong? — answered

Each candidate is a Hypothesis whose subject is the base plus that commit's change; a fix is a new Hypothesis that supersedes it. The Goal's lines are its cases.

Selected: H-VHG4N5
Design: 04142f3 holds d2d5d14, adds a go test that a store outside every repository whose block holds a line that starts with no heading exits 1 with nothing on stdout (P1, R-DER6F8#1), and narrows the test comment to the three cases it pins. It changes no file under skills/, no spec text and no code.

Limits:

- R-9V3ER0#1 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F1, G-ZBM2RE#3: a whitespace-only line inside the constraints block, such as '   ' between May change: and Read only:, must make hyp prompt exit 1 and hyp add goal warn; 10566d4 treats it as empty.
- R-9V3ER0#2 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F2, G-ZBM2RE#1: a store with no repository above it; 10566d4 prints the store path as '.' with exit 0, though line 1 prints the path relative to the repository root.
- R-9V3ER0#3 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F3, G-ZBM2RE#1: --branch with a newline, such as 'b\nMay change: everything', adds a constraint-shaped line after the block; line 1 lets --branch add only its clause.
- R-9V3ER0#4 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F6, G-ZBM2RE#6: run 6's stop clause 'selected the candidate, or 5 Experiments recorded since this goal began' against the printed 'runner owes nothing, or 5 Results recorded under the actor'.
- R-9V3ER0#5 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F7, G-ZBM2RE#6: run 6's clause 'Its Experiment uses --recipe none --rubric reading' is stated by neither the printed prompt nor hyp-goal-run, and E-HQWTA0's eval case had no assertion for it.
- R-9V3ER0#6 closed by R-50GBXM (H-3M3QAS (h-b4f23b2) supports) — F10, G-ZBM2RE#2 and #4: the answer and two-Goal tests compare two prompts that would match even if the header named another Goal; neither checks the header names G alone.
- R-8JQPWG#1 closed by R-FRF929 (H-YXSJEM (h-5c5902d) supports) — G1, G-ZBM2RE#1: the same store reached through a symlinked working directory and through its physical path must print the same bytes; b4f23b2 exits 3 on one and 0 on the other.
- R-8JQPWG#2 closed by R-FRF929 (H-YXSJEM (h-5c5902d) supports) — G2, G-ZBM2RE#3: a CRLF body whose block holds a blank line, which is "\r" after the split on \n; b4f23b2 refuses it as a non-empty line with no heading.
- R-8JQPWG#3 closed by R-FRF929 (H-YXSJEM (h-5c5902d) supports) — G8, G-ZBM2RE#2 and #4: the tests check the Goal id in one of its four places; a prompt that names G in its header and another Goal in hyp show, hyp status or --goal would pass.
- R-KYKWH9#1 closed by R-28A8NB (H-YTD2JG (h-2284172) supports) — K1, G-ZBM2RE#1: a store whose real directory is outside any repository, reached through a link inside a repository, must give the same outcome as through its real path; no go test runs this case.
- R-3YJ7T2#1 closed by R-356DV0 (H-FQ103R (h-3cfd08d) supports) — M2, G-ZBM2RE#1: a CRLF body with no blank line in its block must print each block line byte for byte, with its carriage return; no go test runs this case.
- R-3YJ7T2#2 closed by R-356DV0 (H-FQ103R (h-3cfd08d) supports) — M3, G-ZBM2RE#3: a store outside every repository whose Goal body has no constraints block must exit 1 with nothing on stdout, not 3; no go test runs this case.
- R-JHBCWS#1 closed by R-27Y1EV (H-0KVRSQ (h-d2d5d14) supports) — N1, G-ZBM2RE#3: a store outside every repository whose Goal body has a heading but no valid block (no "Read only:" line) must exit 1 with nothing on stdout, not 3; no go test runs this case.
- R-DER6F8#1 closed by R-BSWFXG (H-VHG4N5 (h-04142f3) supports) — P1, G-ZBM2RE#3: a store outside every repository whose Goal body has all three headings and a block line that starts with none of them must exit 1 with nothing on stdout, not 3; no go test runs this case.
- R-BSWFXG#1 closed by R-HXEBAF (H-VHG4N5 (h-04142f3) supports) — Q1, G-ZBM2RE#3: a store outside every repository whose Goal body has all three headings and a block line of text that starts with no heading, such as "The end.", must exit 1 with nothing on stdout, naming that line, not 3.

- H-Q5P9R9 feat/goal-prompt at 10566d4: hyp prompt, its tests, the fixture and the owner-flow skill text — superseded
  - E-HQWTA0 Gate 10566d4 against G-ZBM2RE's seven lines — run (recipe none, rule by reading)
    - R-KPWZZC recorded H-Q5P9R9=supports: On 10566d4: go vet and go test ./... -count=1 pass; six TestPrompt* fail on base 2881e7d (no prompt command, exit 2) and pass on the candidate; both removal checks fail TestPromptRefusesAnInvalidConstraintsBlock; quick_validate 0 errors and lint_cruft --routing high=0 on the four changed SKILL.md, skillcheck passes. Evals 6d65ff69, blind haiku consumers and a sonnet reviewer: candidate passes every assertion except c04-strict (l6 26/27, hand-offs exact bytes); original skills fail 13 of 36 rubric assertions and both exact-prompt checks.
  - E-C3GGGJ Review 10566d4's diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-9V3ER0 recorded H-Q5P9R9=contradicts: The fresh review of 10566d4 reports F1-F10. The probe with the 10566d4 binary confirms F1: a whitespace-only line inside the block draws no add-goal warning and hyp prompt exits 0, where G-ZBM2RE#3 refuses a non-empty line with no heading. It confirms F2 (no repository: path '.') and F3 (a newline in --branch adds a constraint-shaped line). F4, F5, F8 and F9 name no line.
- H-3M3QAS feat/goal-prompt at b4f23b2: 10566d4 plus fixes for review findings F1-F3, F5, F7, F9, F10 — superseded
  - E-Z5BTZR Gate b4f23b2 against the seven lines and R-9V3ER0's limits — run (recipe none, rule by reading)
    - R-50GBXM recorded H-3M3QAS=supports: On b4f23b2: go vet and go test ./... -count=1 pass; each TestPrompt* fails on 2881e7d with b4f23b2's tests and passes on b4f23b2; the Prints and Refuses tests also fail on 10566d4 (F1: a line of spaces; F2: no repository; F3: newline in --branch); four removal checks each fail a TestPrompt* (the --branch one rerun after a script error, logged). Scripts: 0 errors, high=0, skillcheck passes. Evals 2aeb0abe, blind: candidate passes all but c04-strict (l6 27/28 with c15, c16 run 6's stop and c27 the Experiment setup); original skills fail 15 of 34 rubric assertions.
  - E-7QQ2DP Review b4f23b2's fix diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-8JQPWG recorded H-3M3QAS=contradicts: The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and flags, exit 3 through a symlinked working directory and exit 0 with 'in inv/t' through the physical path, where G-ZBM2RE#1 prints the same bytes on every call. It confirms G2 (a CRLF blank line inside the block is refused as "\r") and G3 (a no-break space in --branch exits 2). G3, G4, G5, G6 and G7 are text or test defects; G4 also names a rule conflict outside the lines.
- H-YXSJEM feat/goal-prompt at 5c5902d: b4f23b2 plus fixes for second-review findings G1-G8 — superseded
  - E-NZR3GV Gate 5c5902d against the seven lines and R-8JQPWG's limits — run (recipe none, rule by reading)
    - R-FRF929 recorded H-YXSJEM=supports: On 5c5902d: go vet and go test ./... -count=1 pass; each TestPrompt* fails on 2881e7d with 5c5902d's tests and passes on 5c5902d; the symlink test also fails on b4f23b2 (G1); a tab and a CRLF blank line are refused (G2, meaning 1); every Goal id is checked (G8); five removal checks each fail a TestPrompt*. strict-crlf fails (no verdict). Scripts: 0 errors, high=0, skillcheck passes. Evals 2aeb0abe, all five cases, blind: candidate passes all but c04-strict; original skills fail 14 of 36 rubric assertions and both exact-prompt checks.
    - R-9EGPYD recorded H-YXSJEM=supports: Same run as R-FRF929, recorded again to cite the evals file's hash: the eval cases are goal-prompt-evals-2.json, sha256 2aeb0abe, graded with eval_skill.py c47a38de; the candidate passes every assertion but c04-strict, and strict-crlf fails (no verdict).
  - E-ASC28P Apply the owner's line-3 and line-6 answers to 5c5902d's kept records — run (recipe none, reassess from kept records)
    - R-0E8PFH recorded H-YXSJEM=supports: Read E-NZR3GV's kept records under the owner's answers: go vet exit 0 and go test exit 0; skill scripts high=0 and skillcheck passes; the candidate passes every eval assertion except c04-strict (l6 27/28), which line 6's answer leaves without a verdict; strict-crlf fails, which line 3's answer leaves without a verdict. No other check fails.
  - E-H98Z3B Review 5c5902d's fix diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-KYKWH9 recorded H-YXSJEM=supports: The fresh review of 5c5902d's fix reports R-8JQPWG#1-#3 addressed and K1-K5. The probe with the 5c5902d binary (b83eb5a13cc1c56d) runs K1: a store whose real directory is outside any repository, linked into one, exits 3 through the link and through its real path, the same outcome; b4f23b2 (b6712e4e51925096) exits 0 and 3, so it breaks line 1. K2-K5 name no line: a test skip, two text gaps and a skill sentence. — uncontained
- H-YTD2JG feat/goal-prompt at 2284172: 5c5902d plus a test and spec text for review findings K1, K3, K5 — superseded
  - E-DYR1JM Gate 2284172 against the seven lines and R-KYKWH9#1 — run (recipe none, reassess from kept records)
    - R-28A8NB recorded H-YTD2JG=supports: On 2284172: go vet and go test ./... -count=1 pass; all eight TestPrompt* pass, none skipped. TestPromptOutsideARepository from 2284172 fails on b4f23b2's code (exit 0 through the link) and passes on 2284172; with EvalSymlinks removed it fails. probe-k1.sh with the 2284172 binary (c5858435f689b848) exits 3 on both paths. skills/ is identical to 5c5902d, so E-NZR3GV's evals 2aeb0abe and scripts hold for #6 and #7. strict-crlf fails (no verdict).
  - E-2A0909 Review 2284172's fix diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-3YJ7T2 recorded H-YTD2JG=supports: The fresh review of 2284172 reports the new test red when EvalSymlinks is removed, so it runs R-KYKWH9#1, and M1-M3. The probe with the 2284172 binary (c5858435f689b848) runs M2: a CRLF body with no blank line exits 0 and keeps the \r on all three block lines; and M3: a store outside every repository with no block exits 1 with empty stdout. Both outcomes are right; no go test runs either. M1 names no line; M3 also shows the spec does not say exit 1 comes before exit 3. — uncontained
- H-FQ103R feat/goal-prompt at 3cfd08d: 2284172 plus tests and spec text for review findings M2, M3 — superseded
  - E-Y7C84P Gate 3cfd08d against the seven lines and R-3YJ7T2's limits — run (recipe none, reassess from kept records)
    - R-356DV0 recorded H-FQ103R=supports: On 3cfd08d: go vet and go test ./... -count=1 pass; all eight TestPrompt* pass, none skipped. Stripping a trailing \r from printed block lines fails the CRLF check; checking the store's place before the block fails the no-block check (exit 3, want 1). probe-m2-m3.sh with the 3cfd08d binary (9e0156970672a0b7) passes M2 and M3. skills/ is identical to 5c5902d, so E-NZR3GV's evals 2aeb0abe and scripts hold for #6 and #7.
  - E-TZTG0G Review 3cfd08d's fix diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-JHBCWS recorded H-FQ103R=supports: The fresh review of 3cfd08d reports the new CRLF test red under three mutations of prompt.go (strip \r, normalise CRLF, refuse a trailing \r) and the new no-block test red when the block check moves after the repository lookup; the new spec sentence is true of the code. N1: a mutation that checks only that a block exists before the lookup, and its validity after, passes every TestPrompt test. The probe with the 3cfd08d binary (785dd8d6a43871e7) runs N1: an invalid block in a store outside every repository exits 1 with empty stdout. The outcome is right; no go test runs it. N2 names no line. — uncontained
- H-0KVRSQ feat/goal-prompt at d2d5d14: 3cfd08d plus a test for review finding N1 — superseded
  - E-WQ4EVS Gate d2d5d14 against the seven lines and R-JHBCWS's limit — run (recipe none, reassess from kept records)
    - R-27Y1EV recorded H-0KVRSQ=supports: On d2d5d14: go vet and go test ./... -count=1 pass; all eight TestPrompt* pass, none skipped. Checking only that a block exists before the repository lookup, and its validity after, fails the invalid-block check (exit 3, want 1). probe-n1.sh with the d2d5d14 binary (7b75abbede43c89c) passes N1: exit 1, empty stdout, stderr names the missing "Read only:" line. skills/ is identical to 5c5902d, so E-NZR3GV's evals 2aeb0abe and scripts hold for #6 and #7. — uncontained
  - E-ZMMVPN Review d2d5d14's fix diff against G-GZR6C6 and the lines; probe each finding — run (recipe none, rule by reading)
    - R-DER6F8 recorded H-0KVRSQ=supports: The fresh review of d2d5d14 reports the new invalid-block check red when only a missing block is refused before the repository lookup (exit 3, want 1), and the parent's tests green under that mutation, so the pin is new. P1: a mutation that refuses only a bad block line after the lookup passes every TestPrompt test, so the comment's "present but not valid" is pinned only for a missing heading. The probe with the d2d5d14 binary (7b75abbede43c89c) runs P1: a line of two spaces in the block of a store outside every repository exits 1 with empty stdout. The outcome is right; no go test runs it. — uncontained
- H-VHG4N5 feat/goal-prompt at 04142f3: d2d5d14 plus a test for review finding P1 — supported
  - E-GE085A Gate and review 04142f3 against the seven lines and R-DER6F8's limit — run (recipe none, reassess from kept records)
    - R-BSWFXG recorded H-VHG4N5=supports: On 04142f3: go vet and go test ./... -count=1 pass; all eight TestPrompt* pass, none skipped. Refusing a bad block line only after the repository lookup fails the bad-line check (exit 3, want 1). probe-p1.sh with the 04142f3 binary (e1b6ce2114f4689e) passes P1. skills/ is identical to 5c5902d, so E-NZR3GV's evals 2aeb0abe and scripts hold for #6 and #7. The fresh review finds no wrong outcome. Q1: a mutant that defers only a bad line of text passes every TestPrompt test. Q2's naming of the bad line is pinned by the table test inside a repository. — uncontained
  - E-ZG1BC8 Run Q1's case on the 04142f3 binary: a text bad line outside every repository — run (recipe none, reassess from kept records)
    - R-HXEBAF recorded H-VHG4N5=supports: probe-q1.sh with the 04142f3 binary (e1b6ce2114f4689e) runs Q1: a store outside every repository whose block ends with the line "The end." exits 1 with empty stdout, and stderr names "The end.". The outcome is right; no go test runs it outside a repository. The evals R-BSWFXG names for #6 are goal-prompt-evals-2.json, sha256 2aeb0abe, in evals-hash.log. — uncontained

# G-908PRZ A budget-bound build run ends with its final candidate reviewed, without a review chain on small changes — complete

For issue #141 (part of #113). The candidate is a hyp commit over main 9c8e498b87683e50a94a410b6a3cfecdd7a0d441 that changes the run practice of a build Goal (review, budget, stop) in skills/hyp-goal-run, or hyp prompt's default N. Owner's prediction: H-whole-candidate is contradicted on cost.
Cut A: this store as of 01M75T1YRF2R2Z99DX0E8JQPWG (R-8JQPWG), actor claude-g113-1, code at b4f23b2. Cut B: as of 01M76V52SZ9DMCYFVYFRJEGDSA (run 2's stop), a fresh actor, code at 3cfd08d.
A trial replays a cut under the recipe the run builds, on Sonnet, with the method stated in the Question's body.
A review Result gives a verdict on a candidate and cites a review written by a subagent whose transcript in the trial read that candidate's diff.
May change: this repository, in commits on the branch the prompt names, in a worktree separate from the checkout ~/.claude/skills links to, following AGENTS.md; the run may push that branch and open or update one PR, and never merges it.
May change: skills/hyp-goal-run and the text the change makes wrong (skills/, docs/, AGENTS.md, README.md); cmd/, internal/ and testdata for H-raise-budget only; a new recipes/ directory.
Read only: plans/, every investigations/ tree except goal-prompt, ak-skill-creator under .claude/skills.
Never: commit to main or merge a PR.
Never: replace an installed hyp; compile each candidate to a separate path.
Never: edit an existing recipe version, rubrics/, another tree, .claude/skills or another repository.
Never: let a trial reach a remote, or run one on a model other than Sonnet except the one Opus trial.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.

## Q-8YMFW1 Which run practice ends a budget-bound build run with its final candidate reviewed, without a review chain? — answered

Separating scenarios: replay cut A and cut B under each candidate. Method: sandbox at the cut's commit; the candidate's skills/ in wt/.claude/skills; claude -p on Sonnet with --setting-sources project; prompt printed in wt with N = the candidate's default; gh auth unset. Doctor: loaded hyp-goal-run sha256 = the candidate's; no later commit reachable. Same model, effort, claude version, timeout per trial. Rows (candidate, cut, trial) keep total_cost_usd, duration_ms, num_turns, pass; none dropped. Median: middle of 3 sorted costs; a trial with no result record counts highest.

Selected: H-SFEA2M
Design: The review of a candidate is recorded as its review Result: the subagent's reply saved word for word and cited by a Result that gives the candidate a verdict, inside its gate Experiment so it costs no extra Result. No select of a candidate without one; the stop checkpoint names each candidate that lacks one. Built as a commit that changes skills/hyp-goal-run. From R-4XNFHN: every arm asked subagents, but runs cited a summary or no Result.

Limits:

- R-S0BRXE#1 closed by R-T2WJ7V (H-SFEA2M (review-result) supports) — whole-candidate F1, G-908PRZ#1: a run builds C1, fixes to C2 and C3, and reviews only C3, as 'one review replaces a review of each fix commit' says; C1 and C2 end with no review Result.
- R-S0BRXE#2 accepted by owner-dat — whole-candidate F2, G-908PRZ#4: the review is tied to the select; a run that stops on its budget with no select owed asks for no review, and nothing names the unreviewed candidate.
- R-S0BRXE#3 accepted by owner-dat — name-at-stop F2, G-908PRZ#4: with many unreviewed candidates and AWAITING OWNER keys, the stop summary passes 600 runes; hyp refuses it, or the run trims the unreviewed names.
- R-S0BRXE#4 accepted by owner-dat — reserve-last F3, G-908PRZ#1: the rule fires only at exactly one Result left; Results spent on readings or limit closes leave zero at the next commit, which then has no review Result.
- R-S0BRXE#5 closed by R-T2WJ7V (H-SFEA2M (review-result) supports) — one-experiment F1, G-908PRZ#4: the gate-plus-review Result of C1 finds a text defect at the budget's end; the fix C2 is committed with no Result left and the stop checkpoint does not name it.
- R-S0BRXE#6 closed by R-T2WJ7V (H-SFEA2M (review-result) supports) — one-experiment F3, G-908PRZ#1: a review with cases of lines adds limit-closing Results, so a candidate costs more than one Result and the budget can end before the select.
- R-S0BRXE#7 accepted by owner-dat — raise-budget F1, G-908PRZ#4: with skill text unchanged, readings, limit closes and reassess Results fill all 7 before the review; the run stops on an unreviewed candidate it does not name.
- R-S0BRXE#8 accepted by owner-dat — raise-budget F3, G-908PRZ#3: a run that stopped at its 5th Result can record two more before the stop clause, so its median cost can pass 1.5 times the base at N 5.
- R-S0BRXE#9 accepted by owner-dat — probe-closes F4, G-908PRZ#1: k probes run before the review Result is recorded; with N 5 the budget ends before the review Result exists.
- R-S0BRXE#10 accepted by owner-dat — probe-closes F2, G-908PRZ#1: one passing probe of a narrow finding is recorded as --closes on a whole Goal line, which no replay of that line ran.
- R-4XNFHN#1 accepted by owner-dat — G-908PRZ#1 and #2: score reads 'a review written by a subagent' as a cited file holding the subagent's own text. No check ran the reading 'the runner's summary of that review'; by results-and-selects.txt it would not lift any candidate over all four lines.
- R-4XNFHN#2 closed by R-BHPHN5 (H-SFEA2M (review-result) supports) — G-908PRZ#3: total_cost_usd is the session's cost-state total; whether it holds the subagents' own cost was not checked, so an arm that spends more on reviews may be under-counted.
- R-KWYBA6#1 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1, #2: the text names git show <commit>, the delta against the parent, not git diff <base> <commit>; a fix commit's review misses defects carried from the commit before it.
- R-KWYBA6#2 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1, #3: the text does not order the review before the gate's Result is recorded; a run that records the gate first spends a second Result on the review.
- R-KWYBA6#3 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1: the pre-registered verdict rule does not name the review or what its findings do to the verdict.
- R-KWYBA6#4 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#3: each fix commit is a new candidate and gets its own review, a review chain on small changes; no text lets a fix reuse its predecessor's review.
- R-KWYBA6#5 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1: a candidate committed after the last Result is used is built but never gated, so never reviewed; the text has no rule against it.
- R-KWYBA6#6 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1, #2: 'Review each candidate before any select of it' reads as every candidate or only a candidate to be selected.
- R-KWYBA6#7 closed by R-RSV3E7 (H-SFEA2M (review-result) supports) — G-908PRZ#1, #2: 'Give it the code' lets a runner paste a summary, so the subagent's transcript may not hold the diff.
- R-YNB43F#1 accepted by owner-dat — G-908PRZ#3: the base rows are E-XVFAZ6's trials from earlier the same day, not run in the same batch as this arm; same model and claude version, but not interleaved.
- R-YNB43F#2 closed by R-BHPHN5 (H-SFEA2M (review-result) supports) — G-908PRZ#3: total_cost_usd is the session's cost-state total; whether it holds the subagents' own cost was not checked, so this arm's review cost may be under-counted.

- H-FXFZVB A candidate's review is part of its gate Experiment — contradicted
  - E-S9A619 Gate each candidate commit: line 5's skill scripts, skillcheck, go vet and go test — run (recipe none, rule by reading)
    - R-WPB1BB recorded H-FXFZVB=supports H-19CNVM=supports H-2497ZQ=supports H-AAHHZ1=supports H-XEJT0S=supports H-P97YYS=supports: On each commit, every changed SKILL.md (hyp-goal-run; hyp-owner-grill for d0e8591) gives quick_validate.py exit 0 'Skill is valid!' and lint_cruft.py --routing high=0; go vet and go test ./... -count=1 exit 0, skillcheck ok. d0e8591: TestPromptPrintsTheGoalsRunPrompt fails on 7825275's prompt.go, passes on the candidate, and fails with the default set back to 5.
  - E-XVFAZ6 Replay cut A and cut B three times per arm and score reviews, selects, stop names and cost — run (recipe run-review-replay@1, rule by reading)
    - R-4XNFHN recorded H-FXFZVB=contradicts H-19CNVM=contradicts H-2497ZQ=contradicts H-AAHHZ1=contradicts H-XEJT0S=contradicts H-P97YYS=contradicts: 42 trials, none timed out. a1 passes in 0 of 3 cut-A trials and b2 in 0 of 3 cut-B trials for every arm, base included, so #1 and #2 fail for all six. Cost ratio at cut A: one-experiment 1.44, name-at-stop 1.01, probe-closes 1.61, raise-budget 1.85, reserve-last 1.52, whole-candidate 0.87; at cut B all <=1.32. #4: reserve-last fails one trial (rr-reserve-last-a-1). Runs did ask subagents, but cited the runner's summary or no Result at all, so no Result met the review definition.
  - E-PMK8H9 Fresh subagent review of each candidate commit's diff against the Goal's lines — run (recipe none, rule by reading)
    - R-S0BRXE recorded H-FXFZVB=supports H-19CNVM=supports H-2497ZQ=supports H-AAHHZ1=supports H-XEJT0S=supports H-P97YYS=supports: Six fresh reviews (one per commit, diff vs 2cea6c2) report no text that tells a run to stop with an unreviewed candidate unnamed, and none that breaks #5. Ten cases of lines #1, #3, #4 that no check ran are the limits. Text defects (reserve-last F2, probe-closes F3/F7, whole-candidate F5, one-experiment F2) get fix commits on the candidates the replay leaves standing; findings of no line go to the stop checkpoint.
- H-19CNVM The stop checkpoint names each candidate without a review — contradicted
- H-2497ZQ With one Result left, review the current candidate instead of committing another fix — contradicted
- H-AAHHZ1 hyp prompt's default N rises to fit a gate and a review per candidate — contradicted
- H-XEJT0S A review case the run's probe passes closes in the review's Result — contradicted
- H-P97YYS The review before the select reads the cumulative diff from the Goal's base — contradicted
- H-SFEA2M feat/run-review: a review is a Result citing the subagent's own reply, before any select — supported
  - E-3CNWXD Gate H-SFEA2M's commit: line 5's skill scripts, skillcheck, go vet and go test — run (recipe none, rule by reading)
    - R-XT3JRJ recorded H-SFEA2M=supports: On ea697bf: quick_validate.py exit 0 and lint_cruft.py --routing high=0 on skills/hyp-goal-run/SKILL.md; go vet and go test ./... -count=1 exit 0, skillcheck included.
  - E-MKEBRG Fresh subagent review of H-SFEA2M's diff against the Goal's lines — run (recipe none, rule by reading)
    - R-KWYBA6 recorded H-SFEA2M=supports: Fresh subagent review of ea697bf, saved word for word: 11 findings, none shows by reading that the text breaks #4 or #5; it says the new checkpoint slot satisfies #4. Findings 1-7 are cases of #1-#4 that the E-2TJK2N replay runs; 8-10 are labelled none.
  - E-2TJK2N Replay cut A and cut B three times for H-SFEA2M and score against E-XVFAZ6's base — run (recipe run-review-replay@1, rule by reading)
    - R-YNB43F recorded H-SFEA2M=supports: Arm review-result at ea697bf, 6 Sonnet trials on claude 2.1.322, none timed out. a1 in 2 of 3 cut-A trials (a-3 spawned no subagent); b2 in 3 of 3 cut-B; median cost 0.6463 at A (ratio 1.4573) and 0.4428 at B (0.9405) against E-XVFAZ6's base rows; l4 in all 6. Strict flags equal the plain ones.
    - R-RSV3E7 recorded H-SFEA2M=supports: Same rule and trials as R-YNB43F, citing the inputs that pin ea697bf. kwyba6-cases.txt names where each R-KWYBA6 case ran in the trials: b-3 reviewed a 3-commit fix chain with git show (cases 1, 4); a-3 built a candidate with its last Result and left it unreviewed but named (case 5); every subagent ran the diff itself (case 7). — uncontained
    - R-T2WJ7V recorded H-SFEA2M=supports: Same rule and trials as R-YNB43F (inputs pin ea697bf). s0brxe-cases.txt names where three R-S0BRXE cases ran: b-3 reviewed every commit of a 3-commit fix chain (#1); a-3 committed a candidate at the budget's end and named it at the stop (#5); a-1 and a-2 recorded review limits and stopped before a select, while cut B selected within N in 3 of 3 (#6).
  - E-2A6FW1 Probe: does claude -p's total_cost_usd include a subagent's usage? — run (recipe none, rule by reading)
    - R-BHPHN5 recorded H-SFEA2M=supports: claude 2.1.322, Sonnet: modelUsage outputTokens 486 = main transcript 315 + subagent transcript 171; every token kind sums exactly, and costUSD equals total_cost_usd 0.0677498. total_cost_usd includes the subagent's usage, so #3's medians count review subagents.

# G-QXM27F hyp-owner-grill keeps a grill from repeating S1 to S6 of #146, and hyp add goal refuses a truncated sha256 — complete

For issue #146 (part of #113). The candidate is a hyp commit over main 132cb8e21ad5750266e11d413a172d1d9a891976 that changes skills/hyp-owner-grill to fix entries S1 to S6 of #146, and hyp add goal for line 9; S7 is another Goal. An entry whose line holds on the base is recorded as not a skill defect, citing that Result, with no text change; for S2 the stop names SKILL.md's [N] rule as an out-of-Goal finding.
Fixture A: the #113 grill; repository and hyp at a51f140; issue #113 as of 2026-11-11T16:20Z. Fixture B: the #141 grill; repository and hyp at 9c8e498; issue #141 as of 2026-11-12T03:45Z. A trial runs the grill under the candidate or base text in a sandbox; a stand-in owner gives the owner's replies in session 285fa63e or f50e4740 to the questions the grill asks, paraphrased in a bundle file.
May change: this repository, in commits on the branch the prompt names, in a worktree separate from the checkout ~/.claude/skills links to, following AGENTS.md; it may push that branch and open or update one PR.
May change: skills/hyp-owner-grill and every text the change makes wrong (skills/, docs/, AGENTS.md, README.md); cmd/, internal/ and testdata for line 9; a new investigations/goal-prompt/recipes/hyp-owner-grill-eval@N.
Read only: plans/, every investigations/ tree except goal-prompt, ak-skill-creator under .claude/skills, transcripts 285fa63e and f50e4740 for the stand-in's replies.
Never: commit to main or merge a PR.
Never: replace an installed hyp; build each candidate to its own path.
Never: edit a recipe version, rubrics/, another tree, .claude/skills or another repository.
Never: commit the transcripts 285fa63e or f50e4740 or text copied from them.
Never: let a trial reach a remote or write outside its sandbox, or run a trial or its subagents on any model but Sonnet.
Never: write under an owner's actor outside a trial's sandbox store, where the stand-in is owner-standin; if a step needs the owner, refer or record a reading, and stop.

## Q-CHAT7K Which change keeps a grill from repeating S1 to S6 of #146? — answered

A rival's candidate is one commit. Each fixture runs 3 trials under it and 3 under the base, a fresh Sonnet session each; the stand-in approves each owner write. A trial with no Goal with lines passes no line of 1 to 6. eval_skill.py (sha256 c47a38dee2456cc6172dff28472052c0f7b28d7baf310af6e83b5ca655d10a24, External:) grades; a reviewer who did not run the trial and does not know its arm grades each rubric assertion. Cost: the sum of total_cost_usd over a trial's calls; median: the middle of 3, a trial with no record counting highest.

Selected: H-RHHANS
Design: Commit a95a21e04b2f284bc237bab3653956d727d2f32b on feat/owner-grill (SKILL.md sha256 54a2c953ea6a5695d2904a2c7e6090c053392f314838dadf38579f85e3a97aed). Over main 132cb8e: hyp add goal refuses a cut hash before an ellipsis (39d2810), and SKILL.md keeps the base steps and adds one mistakes checklist of the entries red on base (store, test method, constraints), run before each round and before Confirm the routing. S2, S4 and S6 hold on base (R-B3TNBP) and are not in it.

- H-Z8WNJK Fix each entry of S1 to S6 where it arises in SKILL.md — superseded
- H-6M7FDJ One S1 to S6 checklist run before each round and before the routing — superseded
- E-6HZ9PK Measure the base hyp-owner-grill text on fixtures A and B, 3 trials each, to find which of S1 to S6 recur — run (recipe hyp-owner-grill-eval@1, measure only)
  - R-5M7RDM recorded  Six base trials under hyp-owner-grill-eval@1 ran to the stand-in's end (3 to 9 exchanges), cost_usd 1.02 to 1.75. Mechanical checks: a Goal with lines in every trial; l1 fails in base-a-1 (a store at the clone's top directory) and passes in the other five; l4 passes in all six. Not graded by a reviewer: in base-a-3 the grill listed the sandbox's recipe copy (its stand-in and cases files by name) and grepped issue-a.md there, so Drive's isolation leaks the eval. hyp-owner-grill-eval@2 denies those reads; the base is measured again under @2.
- E-8BJ6T5 Measure the base hyp-owner-grill text on fixtures A and B under @2, 3 trials each, to find which of S1 to S6 recur — run (recipe hyp-owner-grill-eval@2, measure only)
  - R-956X7J retracted  Six base trials under @2 ended at the stand-in's end (4 to 6 exchanges). table, with a blind Sonnet review of the rubric assertions: fixture a passes l1 1/3, l2 3/3, l3 0/3, l4 2/3, l5 0/3; fixture b passes l1 3/3, l4 3/3, l5 3/3, l6 3/3. Median cost_usd a 1.3349, b 1.0794 (literal 4.0429, 3.4299). Lines 1, 3 and 5 fail on base, so S1, S3 and S5 are skill defects. Lines 2, 4 and 6 hold on base, so S2, S4 and S6 are no skill defect, and the candidates leave their text unchanged.
  - R-B3TNBP recorded  Six base trials under @2 ended at the stand-in's end (4 to 6 exchanges). table, with a blind Sonnet review of the rubric assertions: fixture a passes l1 1/3, l2 3/3, l3 0/3, l4 2/3, l5 0/3; fixture b passes l1 3/3, l4 3/3, l5 3/3, l6 3/3. Median cost_usd a 1.3349, b 1.0794 (literal 4.0429, 3.4299). Lines 1, 3 and 5 fail on base, so S1, S3 and S5 are skill defects. Lines 2, 4 and 6 hold on base, so S2, S4 and S6 are no skill defect, and the candidates leave their text unchanged.
- H-PBAPW8 feat/owner-grill 6b7095e: entry edits for S1, S3 and S5, with line 9's refusal — contradicted
  - E-AATV5F Run both candidate texts on fixtures A and B, 3 trials each, and grade lines 1 to 7 against base — withdrawn (recipe hyp-owner-grill-eval@2, rule by reading)
  - E-7JGBSV Gate both candidate commits on lines 8 and 9: red on base, green, removal, vet, test, validate, lint, review — run (recipe none, rule by reading)
    - R-4Z8QS9 recorded H-PBAPW8=supports H-RHHANS=supports: Both 6b7095e and a95a21e, each in detached worktrees over 132cb8e: the Truncated tests fail on base with the candidate's test file (exit 0, want 1), pass on the candidate, and fail again with the truncatedHex call removed; go vet and go test ./... exit 0, skillcheck ok. quick_validate.py prints Skill is valid! and lint_cruft.py --routing high=0 (one medium numeric-cap) on skills/hyp-owner-grill; ids.txt holds both scripts' whole sha256 as line 8 pins. Lines 8 and 9 hold on both. The review's other findings go to the stop report.
  - E-3BDZ90 Run both candidate texts on fixtures A and B, 3 trials each, rerunning the three trials a usage limit cut — run (recipe hyp-owner-grill-eval@2, rule by reading)
    - R-1E0YVX retracted H-PBAPW8=contradicts H-RHHANS=supports: 12 arm trials (9 kept from E-AATV5F, 3 rerun), all at the stand-in's end; two blind Sonnet reviewers, table. preflight a95a21e: a l1 3/3 l2 3/3 l3 3/3 l4 3/3 l5 3/3; b l1 3/3 l4 2/3 l5 2/3 l6 3/3; median cost_usd 0.965x base on a, 1.227x on b. entry-edits 6b7095e: a l2 1/3 (two [N] lines say a write is refused), l1 3/3 l3 2/3 l4 2/3 l5 3/3; b l1 3/3 l4 2/3 l5 3/3 l6 3/3; cost 1.258x, 1.323x. By the rule, preflight meets lines 1 to 7 and entry-edits misses line 2.
    - R-1R2QN8 recorded H-PBAPW8=contradicts H-RHHANS=supports: 12 arm trials (9 kept from E-AATV5F, 3 rerun), all at the stand-in's end; two blind Sonnet reviewers, table. preflight a95a21e: a l1 3/3 l2 3/3 l3 3/3 l4 3/3 l5 3/3; b l1 3/3 l4 2/3 l5 2/3 l6 3/3; median cost_usd 0.965x base on a, 1.227x on b. entry-edits 6b7095e: a l2 1/3 (two [N] lines say a write is refused), l1 3/3 l3 2/3 l4 2/3 l5 3/3; b l1 3/3 l4 2/3 l5 3/3 l6 3/3; cost 1.258x, 1.323x. By the rule, preflight meets lines 1 to 7 and entry-edits misses line 2.
- H-RHHANS feat/owner-grill a95a21e: one checklist of S1, S3 and S5 before each round and the routing, with line 9's refusal — supported

# G-V5NP59 hyp keeps a text past a field's rune cap whole in its event — complete

For #142, #144, #153 (part of #113 and #146): a text the tree states (verdict rule, stop report, Goal body, question) is kept whole in the event that writes it; no pinned file holds a field's text. The candidate is a hyp commit over cf1639a5c4db273632ea183f8d299b6c52d08a73 that raises these caps; hyp show prints each raised field whole, and hyp status prints the last checkpoint's summary whole, so TestCheckpointSummaryInFull changes.
Caps: Experiment and DesignQuestion body 600 to 2000; runner checkpoint summary 600 to 2000; Goal body 2000 to 3000; every title 120 to 200. Fields sharing MaxText that stay 600: Hypothesis body and amend, observation, refer text, answer item text, old owner checkpoint summary; records path stays 300.
Fixtures: internal/cli/testdata/text-cap/ and its CASES file (field, runes, sha256 per draft), checked by go test.
Out of scope: #143, the count caps, #157.
May change: this repository, in commits on the branch the prompt names, in a worktree separate from the checkout ~/.claude/skills links to, following AGENTS.md; the run may push that branch and open or update one PR, and never merges it.
May change: cmd/, internal/ and their testdata, and the texts the change makes wrong (skills/, docs/, AGENTS.md, README.md).
Read only: plans/, every investigations/ tree except goal-prompt, and the drafts and CASES under internal/cli/testdata/text-cap/.
Never: commit to main or merge a PR.
Never: replace an installed hyp; compile each candidate to a separate path.
Never: edit an existing recipe version, rubrics/, another tree, .claude/skills or another repository.
Never: add a store-level constraints block (#160), or keep a field's text in a pinned file instead of its event.
Never: write to this store with a candidate binary; tree writes use the installed hyp, which keeps today's caps.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.

## Q-44F8FA How does hyp keep a text longer than a field's rune cap? — answered

Difference test on the issues' options. raise vs pin-file: no sentence; G-V5NP59#1 forbids a pinned text. raise vs keep: on e-hqwta0-650, raise meets G-V5NP59#1 and keep does not, because 600 refuses it and the stored retry drops c04-strict's reason; keep is the base. raise vs split: on g908prz-gate-1346, raise meets G-V5NP59#1 and split does not, because shared clauses (397) plus part 1 (290) and the base check (113) are 801 runes; split also combines with raise. derive and store block combine with raise: not rivals.

Selected: H-CVZZ85
Design: Commit 69f0979e9b00c3b5ceccb7030ce3a2476924d36a on feat/text-cap: a14faa8 (H-6RBHG4) plus the fixes from its review (R-SWVMYT): OneLine turns VT, FF, NEL, LS and PS into a space too, and a limit's history prints a referring summary through OneLine, its spaces kept.

Limits:

- R-ZGHGYM#1 closed by R-SWVMYT (H-6RBHG4 (h-raise-a14faa8) supports) — G-V5NP59#3: TestTextCapStopReportInStatusAndShow finds F4, F8 and G4 as substrings of all stdout, which display ids can also hold; it does not check that the CHECKPOINT line and the show summary line hold the draft whole.

- H-KFK9TX Raise each cap; the event holds the whole text — superseded
  - E-YSHWRE Gate the text-cap commit on lines 1-5: red on base, green, removal, vet, test, review — run (recipe none, rule by reading)
    - R-ZGHGYM recorded H-KFK9TX=supports: b3f223e over 2277aa9: the TextCap tests fail on base with the candidate's test file (go test exit 1; every CASES draft refused at exit 1 over limit) and pass on b3f223e; with the Experiment body cap back at 600 and the status summary cut at 200, four TextCap tests and TestCheckpointSummaryInFull fail; go vet and go test ./... exit 0. By the rule, lines 1-5 hold. The review finds a summary's newlines printed raw in status and show (none-line, a fix commit) and a weak F4/F8/G4 check (line 3).
- H-6RBHG4 feat/text-cap a14faa8: raise each cap; the event holds the whole text; a checkpoint summary prints whole on one line — superseded
  - E-S6QZDR Gate a14faa8 on lines 1-5 and R-ZGHGYM#1: red on base, green, removal, vet, test, review — run (recipe none, rule by reading)
    - R-SWVMYT recorded H-6RBHG4=supports: a14faa8 over 2277aa9: the TextCap tests fail on base with its test file (go test exit 1) and pass on a14faa8; the stop-report test reads each item from the summary's own line (R-ZGHGYM#1). With the Experiment body cap back at 600 and the summary printed raw, three TextCap tests and TestCheckpointSummaryOnOneLine fail; go vet and go test ./... exit 0. strict-bare-store: 5 of 7 drafts exit 2. By the rule, lines 1-5 hold. The review finds \v, \f, U+0085, U+2028, U+2029 pass OneLine and a limit history merges spaces (none-line; fix commit). — uncontained
- H-CVZZ85 feat/text-cap 69f0979: raise each cap; the event holds the whole text; a summary prints on one line for every line break — supported
  - E-1KTGMX Gate 69f0979 on lines 1-5: red on base, green, removal, vet, test, strict checks, review — run (recipe none, rule by reading)
    - R-DC28VG recorded H-CVZZ85=supports: 69f0979 over 2277aa9: the TextCap tests fail on base with its test file (go test exit 1) and pass on 69f0979; with the Experiment body cap back at 600 and OneLine back to LF, CR LF and CR, three TextCap tests and TestCheckpointSummaryOnOneLine fail; go vet and go test ./... exit 0. By the rule, lines 1-5 hold. strict-bare-store: 5 of 7 drafts exit 2. strict-one-line, rerun after its CASES read moved before setup: stand-ins give their known answers; the two goal-body drafts fail. The review finds no defect on any line.
  - E-TEX5AY Apply the owner answers on lines 1 and 2 to 69f0979 from E-1KTGMX's records — run (recipe none, reassess from kept records)
    - R-MEA8C2 recorded H-CVZZ85=supports: Under the owner answers, meaning 1 on lines 1 and 2, the kept records of 69f0979 decide: green.log ends go test exit 0; removal-check.log fails with the old caps put back; gate.log ends go test exit 0; review-reply.txt finds no defect on lines 1-5. strict-bare-store and strict-one-line test meaning 2, which the owner did not choose, so they set no verdict. H-CVZZ85 supports.
