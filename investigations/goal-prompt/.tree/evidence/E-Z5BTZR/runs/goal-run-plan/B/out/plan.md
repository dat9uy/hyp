# Plan: hyp-goal-run answers (Goal G-D1CH1R, B)

Source: the /goal prompt (`B/prompt.txt`) and `B/skills/hyp-goal-run/SKILL.md`. No command was run and nothing was written to the store.

1. **Naming and tying each candidate commit, and what a fix is.** Each commit I test is one Hypothesis under the Goal's question. Its branch and commit go in its title. Before the run I tie its subject with `hyp subject <H> --base <commit>`. A fix is a new commit and a new Hypothesis that supersedes the old one, recorded with `hyp supersede <new> --old <old>`. I write the tree with the installed hyp and compile each candidate to a separate path. I never replace an installed binary. The prompt puts the bin directory on PATH, and I use that hyp to write the tree.

2. **Each Experiment's verdict rule, and when I write it.** I write the rule into the Experiment's body before I run anything. I write no candidate and run no candidate until the rule is recorded on the Experiment. The recipe's checks may exist earlier, because `hyp add exp` pins them. The rule states:
   - the Goal's lines the Experiment aims at;
   - for each new test, red on the base and green on the candidate;
   - for each refusal, a removal check that removes the refusal and shows which test fails;
   - `go vet ./...` and `go test ./... -count=1`, or the repository's own gate;
   - for skill text, the original and the candidate skills run on the same eval cases.
   An Experiment that runs the repository's own gate on a commit uses `--recipe none --rubric reading`, unless the Goal names a pinned recipe. A new or changed check runs only on its cases' stand-ins, never on a candidate, until its rule is recorded.

3. **Where logs are kept.** Logs go under `.tree/evidence/<E-id>/`, and the Result cites them.

4. **What I do before the stop about a review, and how I sort its findings.** Before the stop, I ask a fresh subagent for one review of each candidate commit's diff. I give it the code, the Goal's lines, the keys already settled and the owner's words. I do not give it my conclusion. I ask it to label each finding with the line it is a case of, or none. I sort each finding this way:
   - a case of a line becomes a `--limit` that I close, or refer if the case is void;
   - a question about what a line means becomes a reading;
   - a defect in text I may edit becomes a fix commit;
   - anything else is an out-of-Goal finding. I name it in the stop checkpoint. It never holds the select.

5. **What my last message lists for the owner.** My last message repeats the stop checkpoint. It lists each line under LINES by its key, with the Results that ran it, and names each Question where OWED still owes a line. Then it lists every owner item with the exact command the owner runs:
   - each item under AWAITING OWNER;
   - each out-of-Goal finding, as a new Goal or an issue;
   - merging the PR;
   - reinstalling hyp;
   - each personal skill that names a command the change altered.
   A run that ends with lines owed is a normal result.

Status: DONE, with `out/plan.md` written with the five answers.
