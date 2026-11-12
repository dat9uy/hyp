# Plan for the unattended run (Goal G-D1CH1R)

Sources: the /goal prompt (prompt.txt), skills/hyp-goal-run/SKILL.md, skills/hyp-prototype/SKILL.md and skills/hyp-verification/SKILL.md. Nothing was run and the store was not touched.

1. Naming and tying candidates, and what a fix is

- A candidate is a Hypothesis (`H-` id) under the top-level Question. Give it an easy name with `--alias` when adding it; use the alias anywhere.
- Before a run, tie each candidate's subject with `hyp subject <H> [PATH…] --base <commit> [--patch FILE]`. A Result covers a candidate it gave a verdict on, or one whose whole subject (base, patch and files) a cited run holds.
- A fix, an untested part of a claim, or a stronger instrument is a new Hypothesis, not an edit. Order: `hyp add hyp --parent <Q-id>`, then `hyp supersede <new H-id> --old <old H-id>`, then `hyp add exp --parent <new H-id>`. A fix tested on a supported Hypothesis moves nothing.
- The skills do not say how to name a candidate commit. The rule about naming a binary by sha256 is in AGENTS.md, which is not among the files I was given, so I cannot cite it.

2. Verdict rules: what they hold and when I write them

- Each Experiment gets its verdict rule when it is added, before any candidate runs. The rule is a rubric (`--rubric <name@version>`), or `--rubric reading` with the rule in the body. I write no candidate run until the rule is recorded on the Experiment.
- Checks may exist before the rule, because `hyp add exp` pins them. A new or changed check is tried only on the stand-ins of its cases, never on a candidate, until the rule is recorded.
- If a candidate shows a check is wrong, I do not just fix it. I write the case that shows it, in the Goal's terms with the outcome it must give, in a new recipe version, and name that case in the Experiment body.
- The Goal's must-fail cases are the known-answer flaw cases.
- Where a verdict rests on my reading of a Goal line, I write the other reading as a separate strict check with its own known answers. It sets no verdict. The rule names it, I run it on the same candidates, and I record the reading with `hyp record --reading`.
- A failed pre-registered pass rule is `contradicts`, not `inconclusive`.
- Before a trade-off goes to the owner, I add one structure outside the candidates, and the cheapest variant of each contradicted rival. If it holds every check, I say in a checkpoint that the rule does not decide the Question.

3. Where logs are kept

- Hand-observed evidence goes in `.tree/evidence/<E-id>/` with short file names: command output, logs, timings, screenshots, and the revision from `git rev-parse HEAD`. Scripted Evidence is written there with `hyp bundle evidence --record .tree/evidence/<E-id>/<name> REF OUT_DIR`.
- Longer text goes in those files and is pointed to with `--evidence`. A Result cites only files of its own Experiment's run, plus the records it pinned with `--records`.
- Conflict: this run's instructions say "Write nothing to the store", and the skill keeps logs under `.tree/evidence/`. Under this run's instructions I would keep no logs in the store and report this conflict to the owner instead of writing them.

4. Before the stop, and the review

- Before the last message I record a runner checkpoint with `hyp checkpoint --summary "Stopped: <stop clause>. Waiting for the owner: <keys under AWAITING OWNER>. Out-of-Goal findings: <one short phrase each, or none>" --why "<why>"`, with `--goal G-D1CH1R` as the prompt requires.
- Out-of-Goal findings are named in that checkpoint. They get no limit, no refer and no reading, and never hold the select.
- A defect in text I changed and may edit is fixed with a commit before the stop, not listed as a finding.
- The skills name no review step and say nothing about how to sort review findings. The sorting above (in-Goal case becomes a line-limit, out-of-Goal case goes into the checkpoint, defect in my own edited text is fixed by commit) is what the skills imply, but no review procedure is given. This gap is unresolved.

5. What the last message lists for the owner

- It repeats the checkpoint text.
- It lists each line under LINES by its key, with the Results that ran it.
- It names each Question where OWED still owes a line.
- `hyp status G-D1CH1R` is printed at the end of every turn, including the last.
- Stopping is not success. A run that ends with lines owed is a normal result.
- Stop clauses from the prompt: OWED says "runner owes nothing: stop and report to the owner", or 5 Results are recorded under claude-g1b-6, or I cannot take the next step honestly and must say why.
- Never change the Goal or its met rule. Never record an owner's choice or answer unless the owner gave those words in this session.

Unresolved questions:
- No skill covers review, so item 4 is inferred.
- The "write nothing to the store" rule conflicts with the evidence-location rule in item 3.
