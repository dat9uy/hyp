# Plan for the hyp-goal-run /goal (Goal G-D1CH1R)

Source: the /goal prompt (A/prompt.txt) and the hyp-goal-run skill (A/skills/hyp-goal-run/SKILL.md). No hyp command was run and nothing was written to the store.

1. **Naming and tying each candidate commit; what a fix is.**
   - Each commit I test is one Hypothesis under the Goal's question. Its title carries the branch and the commit.
   - Before the run, I tie its subject with `hyp subject <H> [PATH…] --base <commit> [--patch FILE]`. A Result covers a candidate only if it gave that candidate a verdict, or if the whole subject (base, patch and files) is held by one of the Result's cited runs. So the tie comes before the run.
   - A fix is a new commit and a new Hypothesis that supersedes the old one, with `hyp supersede <new> --old <old>`.
   - I write the tree with the installed hyp. I compile each candidate to a separate path and never replace an installed binary.

2. **Each Experiment's verdict rule: what it holds and when I write it.**
   - I write the rule in the Experiment body before I run anything. It states the Goal lines the Experiment aims at, and for a build Goal it also holds:
     - each new test red on the base and green on the candidate;
     - a removal check for each refusal (remove the refusal and show which test fails);
     - `go vet ./...` and `go test ./... -count=1`, or the repository's own gate.
   - For skill text, the original and the candidate skills run on the same eval cases.
   - The recipe's checks may already exist, because `hyp add exp` pins them. I write and run no candidate until the rule is recorded on the Experiment.
   - I try a new or changed check only on its cases' stand-ins, never on a candidate, until the rule is recorded. If a candidate shows a check is wrong, I write a new recipe version with the case that shows it, in the Goal's terms with its required outcome, and name that case in the Experiment body.
   - If the repository keeps no recipe for the gate, the Experiment uses `--recipe none --rubric reading`, and its body states the gate as the rule.
   - When a verdict rests on my reading of a Goal line, the rule adds the other reading as a strict check: a separate named command with its own known answers that sets no verdict.

3. **Where I keep logs.**
   - Logs go under `.tree/evidence/<E-id>/`, and the Result cites them.
   - When reassessing records an earlier run kept, I pin them with `--records .tree/evidence/<E-id>` and `--recipe none`.

4. **Before the stop: the review, and how I sort its findings.**
   - Before the stop, I ask a fresh subagent for one review of each candidate commit's diff against the design and the Goal's lines. I give it the code, the lines, the keys already settled and the owner's words. I do not give it my conclusion. It labels each finding with the line it is a case of, or `none`.
   - Sorting:
     - a case of a line becomes a `--limit` that I close or, only if the case is void, refer;
     - a question about what a line means becomes a reading;
     - a defect in text I may edit becomes a fix commit;
     - anything else is an out-of-Goal finding. It is named in the stop checkpoint, gets no limit, refer or reading, and never holds the select.

5. **What my last message lists for the owner.**
   - It repeats the runner checkpoint: `Stopped: <stop clause>. Waiting for the owner: <keys under AWAITING OWNER>. Out-of-Goal findings: <one phrase each, or none>.`
   - It lists each line under LINES by its key, with the Results that ran it. It names each question where OWED still owes a line. A run that ends with lines owed is a normal result.
   - It lists every owner item with the exact command the owner runs: each AWAITING OWNER item, each out-of-Goal finding (a new Goal or an issue), merging the PR, reinstalling hyp, and each personal skill that names a command the change altered.
   - I print `hyp status G-D1CH1R` at the end of every turn.
