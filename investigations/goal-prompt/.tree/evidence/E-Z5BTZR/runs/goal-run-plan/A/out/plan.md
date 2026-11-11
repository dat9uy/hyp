# Plan for the hyp-goal-run consumer (answers from the prompt and the skill only)

1. Naming and tying each candidate commit, and what a fix is.
   - Candidate commits go on branch feat/one-subject-per-run, as the prompt's line 9 names. Commits follow AGENTS.md: conventional messages, no AI trailers. The prompt's line 3 limits what may change: the code, ADR 0020, and the spec, glossary, principles, skills/ and README.md text the change makes wrong.
   - Each candidate is tied before its run with `hyp subject <H> [PATH…] --base <commit> [--patch FILE]`. A Result covers a candidate only if it gives that candidate a verdict, or if one of its cited runs holds the candidate's whole subject (base, patch and files). The skill does not say how to name a candidate commit beyond this tie. I would use the commit hash that `hyp subject` takes as `--base`.
   - A fix to a check is not a silent edit. If a candidate shows a check is wrong, I do not just fix the check. I write the case that shows it, in the Goal's terms with the outcome it must give, as a new version of the pinned recipe, and name that case in the Experiment body. A sentence that discloses a trial run pins nothing. A recipe is never edited in place; a change is a new @<N>. Recipes are read only (prompt line 4).

2. What each Experiment's verdict rule holds, and when I write it.
   - The verdict rule is recorded on the Experiment before any candidate is written or run. The recipe's checks may exist earlier, because `hyp add exp` pins them. A new or changed check is tried only on its cases' stand-ins, never on a candidate, until the rule is recorded.
   - Each rule gives a verdict on every target. `inconclusive` is allowed when the run cannot decide. If an outcome is already known, the rule says so.
   - Rubric choice for a reassessment: `--rubric <name>@<version>` from rubrics/ (a new rubric version when an answer changes the rule), or `--rubric reading` with the rule in the body. A recipe measures and never prints a verdict.
   - When a verdict rests on a reading of a Goal line, the other reading is a separate strict check in the recipe, with its own known answers, that sets no verdict. The rule says so, and the check runs on the same candidates.

3. Where I keep logs.
   - The skill says the tree is the record. Each step's outcome goes into the store as a Result, a Question, a Hypothesis or an Experiment, and each runner checkpoint goes in with `hyp checkpoint --goal G-D1CH1R`.
   - Evidence from runs is kept under `.tree/evidence/<E-id>`, and reassessment can pin it with `--records .tree/evidence/<E-id>` and `--recipe none`.
   - The skill says not to read session logs, reports or memory about earlier runs, and to add nothing outside the tree. The prompt also says to write nothing to the store in this consumer run, so this run writes no logs.

4. What I do before the stop about a review, and how I sort its findings.
   - Before the stop, I record a runner checkpoint so the fork is not left only in chat: `hyp checkpoint --summary "Stopped: <stop clause>. Waiting for the owner: <keys under AWAITING OWNER>. Out-of-Goal findings: <one short phrase each, or none>" --why "<why>"`.
   - A defect in text I changed and may edit is fixed with a commit before the stop. It is not an out-of-Goal finding.
   - Each finding is sorted by whether it is a case of a Goal line. A case of a Goal line gets a limit, a refer, or a reading as the skill's Goal-line and Readings sections say. A case of none of the Goal's lines is an out-of-Goal finding. It gets no limit, no refer and no reading, never holds the select, and is named in the checkpoint.
   - The skill has no separate review step. This is my reading of its Stop and Goal-lines sections, and the owner should check it.

5. What my last message lists for the owner.
   - My last message repeats the runner checkpoint.
   - It lists each line under LINES by its key (`G-D1CH1R#<n>`), with the Results that ran it.
   - It names each Question where OWED still owes a line.
   - It names the keys under AWAITING OWNER, with each reading in full and the copies its answer can touch.
   - It ends with `hyp status G-D1CH1R` printed, since the skill says to print status at the end of every turn.
   - A run that ends with lines owed is a normal result. The stop is one of the prompt's clauses: OWED says "runner owes nothing", 5 Results under claude-g1b-6 are recorded, or I cannot take the next step honestly and say why.

Status notes: no command was run and nothing was written to the store. The answers come from the prompt and skills/hyp-goal-run/SKILL.md only.
