# Plan for the /goal run on G-D1CH1R

Answers come from the /goal prompt (prompt.txt) and the hyp-goal-run skill (skills/hyp-goal-run/SKILL.md). No command was run and nothing was written to the store.

1. Naming and tying candidate commits, and what a fix is.
   Each candidate commit I test is one Hypothesis under the Goal's question. Its title carries the branch and the commit. Before the run I tie its subject with `hyp subject <H> --base <commit>`. A fix is a new commit and a new Hypothesis that supersedes the old one, with `hyp supersede <new> --old <old>`. I build candidates to a separate path and never replace the installed hyp. Every checkpoint I write carries `--goal G-D1CH1R`, and I act as actor claude-g1b-6.

2. What each Experiment's verdict rule holds, and when I write it.
   I write the verdict rule into the Experiment body before I run anything. The rule names the Goal lines it aims at. For each new test it states red on the base and green on the candidate. For each refusal it states a removal check: remove the refusal and show which test fails. It also requires `go vet ./...` and `go test ./... -count=1`, or the repository's own gate. Recipe checks may exist before the rule, since `hyp add exp` pins them. I run no candidate until the rule is recorded on the Experiment. A new or changed check runs only on its cases' stand-ins until then. If a candidate shows a check is wrong, I write the case that shows it, in the Goal's terms with the outcome it must give, in a new version, and I name that case in the Experiment body.

3. Where I keep logs.
   Logs go under `.tree/evidence/<E-id>/`, and the Result cites them. A recorded Result cannot change, so a correction is a new Result.

4. What I do before the stop about a review, and how I sort its findings.
   Before the stop, I ask a fresh subagent for one review of each candidate commit's diff. I give it the code, the Goal lines, the settled keys and the owner's words. I do not give it my conclusion. It labels each finding with the line it is a case of, or none. I sort each finding this way:
   - A case of a line becomes a `--limit` that I close. If the limit is void, I refer it.
   - A question about a line's meaning becomes a reading.
   - A defect in text I may edit becomes a fix commit, made before the stop.
   - Anything else is an out-of-Goal finding. It is named in the stop checkpoint, gets no limit, refer or reading, and never holds the select.

5. What my last message lists for the owner.
   My last message repeats the stop checkpoint, `hyp checkpoint --goal G-D1CH1R --summary "Stopped: <stop clause>. Waiting for the owner: <keys under AWAITING OWNER>. Out-of-Goal findings: <each, or none>" --why "<why>"`. It lists each line under LINES by its key, with the Results that ran it, and names each question where OWED still owes a line. Then it lists every owner item with the exact command the owner runs. The items are each AWAITING OWNER item, each out-of-Goal finding, merging PR #132 (the run never merges), reinstalling hyp, and each personal skill that names a command the change altered. I print `hyp status G-D1CH1R` at the end of every turn.
