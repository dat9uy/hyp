---
name: hyp-goal-run
description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report the owner can read from hyp status. Use when a /goal prompt names this skill."
---

The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.

## Start

- Read the Goal first with `hyp show <goal-id>`, including its checkpoints, because `hyp status` does not print its met rule.
- Everything you need is in the tree. Do not read session logs, reports or memory about earlier runs.
- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a step needs something they forbid, refer it with a reason or record a reading, and stop.
- A store may hold many Goals. Then `hyp status` prints only a GOALS table and one Awaiting owner list, so run `hyp status <goal-id>` wherever this skill says `hyp status`: OWED, LINES, the stop line and the last checkpoint are per Goal. Name the Goal with `--goal <goal-id>` on `hyp checkpoint`.
- Print `hyp status` at the end of every turn.

## Work

Act on what OWED in `hyp status` says the runner owes, and on nothing else. Plan only the next Experiment.

- Use the Goal's must-fail cases as known-answer flaw cases, in the owner's terms.
- The recipe's checks may exist before the verdict rule, because `hyp add exp` pins them. Write and run no candidate until the rule is recorded on the Experiment.
- Try a new or changed check only on its cases' stand-ins, never on a candidate, until the rule is recorded. If a candidate shows a check is wrong anyway, do not just fix the check: write the case that shows it, in the Goal's terms with the outcome it must give, in a new version, and name that case in the Experiment body. A sentence that discloses a trial run pins nothing.
- `hyp record` and `hyp check` warn when a Result's text disagrees with its cited files: a verdict that a cited `verdict:` line contradicts, or a commit or hash in the observation or a limit that no cited file holds. Fix each one with a new Result that cites the file; a Result cannot change after it is recorded.

## Before a trade-off

Before a Question goes to the owner as a trade-off, add one structure that is not among the candidates, and the cheapest variant of each contradicted rival that would hold. Run them through a new version of the pinned recipe and record the Result. If a new structure holds every check, say in a checkpoint that the rule does not decide this Question.

## Readings

- When a verdict rests on your reading of a Goal line, write the other reading as a strict check: a separate named command in the recipe, with its own known answers, that sets no verdict. Say so in the Experiment's verdict rule, and run it on the same candidates. Record the reading with `hyp record --reading`, naming the check and which candidates fail it. If the recipe cannot test the other reading, say that in the reading.
- When a verdict rests on the word of a reading that is still open under AWAITING OWNER, or on an owner answer another Result already reads, record the same reading on that Result too, word for word, with the same `--reading-of` key if it has one: a reading holds only its own Result's Question, and hyp prints identical readings as one question and warns when the owner answers only one. Record a copy only on a Result that judged a Hypothesis whose verdict differs from the strict check's outcome for it: a `supports` the check fails, or a `contradicts` the check passes. A Result whose judged Hypotheses the check gives the same outcome as the verdict gets no copy. The owner answers every copy in one answer, because `--answer` is repeatable.
- Before you record a reading, look for a test that can decide it. If a test fits the Goal's lines, write it, run it under the verdict rule, and record no reading. If a test needs a new Goal line, name the test, the fact it measures and the line in the reading, and ask the owner whether to run it. Ask the owner to choose a meaning only when no test can decide, and say why none can.
- Write each reading in ASD-STE100 Simplified Technical English: short sentences, active verbs, one meaning for each word, and the Goal's words for things. Use this form:

  ```text
  <Goal line>: "<the words>" has two meanings. Meaning 1: <...>. Meaning 2: <...>. This run used meaning <n>. <Check> tests meaning <n>; <candidates> fail it. Question: <one question the owner can answer in their own words>?
  ```

  Give one small example of each meaning, in the Goal's terms. If the line does not have two meanings, state the case it does not clearly cover, then the question. A reading holds at most 600 characters.

## Owner answers

- When the owner answers a reading, record it with their exact words, `hyp answer --answer "<R-id>#r<n>=<their words>" --actor <owner-actor>`, and treat their answer as the meaning of that line. If the owner's words answer a reading but do not name it, ask which reading before you write; an answer must name its key. Do not leave a question only in chat.
- If the owner's answer itself has two meanings, record a reading of that answer with `hyp record --reading-of <R-id>#r<n>="<the form above>"`. It narrows the answer and never reopens it; the answer stays in force.
- Only the owner changes an answer. When they give new words for a reading they already answered, record them with the same `hyp answer --answer "<R-id>#r<n>=<their new words>"`; hyp derives that it is a revision. The earlier answer and the Results that applied it stay. Apply the new answer with a Result that `--applies` it, and test a successor (`hyp supersede`) of any Hypothesis the old answer contradicted.
- A limit the owner accepts is recorded with `hyp answer --accept "<key>=<their words>" --actor <owner-actor>`, and a return with `--returns`. hyp admits an accept only of a limit you referred, and never under your actor.

## Reassess

Apply each answered reading that `hyp status` lists under OWED as `apply answer`, and act on each limit it lists.

1. For each answer, write one line: the owner's words, the check, and the candidates that fail it. If the owner is in this session, show the lines and wait for their yes. If not, write the lines in the Experiment's body.
2. For each Question, add one Experiment that names each answer it applies with `--aims`, and fix how its verdicts are decided before you run it: a rubric from `rubrics/` (`--rubric <name>@<version>`, a new rubric version when the answer changes the rule) or `--rubric reading` with the rule in its body. A recipe measures and never prints a verdict. To reassess records an earlier run kept, pin them with `--records .tree/evidence/<E-id>` and `--recipe none`. If an outcome is already known, say so in the rule. Give a verdict on every target, `inconclusive` when the run could not decide. Record its Result with `--applies` for each answer it applied. Old Results stay as they are.
3. If applying an answer raises a new question about what a Goal line or the owner's answer means, record it with `hyp record --reading` on that Result. A Result cannot change after it is recorded, so if it is already recorded, record a new Result under the same Experiment.
4. Act on each `limit` line under OWED, whichever Hypothesis its Result gave a verdict on: a limit is a case the Goal covers that no check ran. If a test within the Goal's lines can run the case on every candidate, add an Experiment with `--aims` at it, run it, and record `--closes` on the Result that ran the case. A close records that the case ran, whatever the verdicts: a candidate that fails the case is contradicted and the limit is still closed, so never write a rule that closes it only on a pass. A close covers every case the limit's text names; `hyp record` prints that text when it closes. Record each case you did not run as a new `--limit` on the closing Result, and refer that new key when its case is void or outside the Goal's lines. A claim that a case is void, out of scope or covered elsewhere is a limit and a refer, never a sentence in a body or an observation. When that claim rests on your reading of a Goal line, put the reading in the refer's text. Write one case per `--limit`. If no such test can, or the case is void for every candidate, refer it with your reason: `hyp checkpoint --refers "<key>=<why no test within the Goal can close it, or why the case is void>" --why "..."`; hyp refuses a refer with no reason. When the owner returns a referred limit, it is an ordinary `limit` line again.
5. End with the candidates left under each Question: none, one, or several, by alias.
6. Select only when OWED lists `select` for that Question: `hyp select <Q-id> <H-id> --why "..."`. Never confirm a select. Do not change the Goal or add new structures unless the owner asks.

## Goal lines

- The Goal's lines are limits keyed on the Goal, `G-<id>#<n>`; `hyp status` prints them under LINES. A Goal's lines come with its add and no later event adds one. If the Goal's body states lines and none is typed, they cannot be added to this Goal: tell the owner, who may add a new Goal with them (`hyp add goal … --line "<line>"`). Do not write them for the owner.
- A line is owed as a `limit` on each top-level Question with a candidate no closer covers. Run its case like any limit: aim at it with `--aims G-<id>#<n>`, and record `--closes G-<id>#<n>` on each Result that ran the case. A Result covers a candidate it gave a verdict on, or one whose whole subject (base, patch and files) one of the Result's cited runs holds, so tie each candidate before the run: `hyp subject <H> [PATH…] --base <commit> [--patch FILE]`.
- A line is never edited. What a line means is a reading. A line no test within the Goal can run is a refer, and the owner accepts or returns it. A new line or a new purpose is a new Goal, so a new tree: tell the owner, and do not write it.

## Stop

Stop on the clauses the `/goal` prompt names. Stopping is not success.

1. Before your last message, record a runner checkpoint, so the fork does not live only in chat; the owner reads `hyp status`, which prints the last checkpoint, each reading in full and what each copy's answer can touch. Do not restate a reading's meanings in the checkpoint:

   ```text
   hyp checkpoint --summary "Stopped: <stop clause>. Waiting for the owner: <the keys under AWAITING OWNER>. Out-of-Goal findings: <one short phrase each, or none>" --why "<why>"
   ```

   An out-of-Goal finding is a case of none of the Goal's lines: it gets no limit, no refer and no reading, and it never holds the select. Name each one in the checkpoint so it lives in the log, not only in chat. A defect in text you changed and may edit is not one: fix it with a commit before the stop.

2. Your last message repeats the checkpoint, and lists each line under LINES by its key, with the Results that ran it, and names each question where OWED still owes a line. A run that ends with lines owed is a normal result.

## Never

- Never change the Goal or its met rule.
- Never record a choice the owner reserved, and never record an owner answer unless the owner gave those words in this session.
