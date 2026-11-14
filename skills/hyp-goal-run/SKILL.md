---
name: hyp-goal-run
description: "Run a hyp tree under an owner's Goal as an unattended /goal, for a discovery Goal or a build Goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report the owner can read from hyp status. Use when a /goal prompt names this skill, such as the one hyp prompt prints."
metadata:
  version: "1.2.0"
---

The `/goal` prompt names the tree, your actor, the Goal, the stop and the Goal body's constraints; `hyp prompt <goal-id> --actor <actor>` prints it from the log. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.

## Start

- Read the Goal first with `hyp show <goal-id>`, including its body and checkpoints, because `hyp status` prints neither the body nor any checkpoint but the last.
- Everything you need is in the tree. Do not read session logs, reports or memory about earlier runs.
- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a step needs something they forbid, refer it with a reason or record a reading, and stop.
- A store may hold many Goals. Then `hyp status` prints only a GOALS table and one Awaiting owner list, so run `hyp status <goal-id>` wherever this skill says `hyp status`: OWED, LINES, the stop line and the last checkpoint are per Goal. Name the Goal with `--goal <goal-id>` on `hyp checkpoint`.
- Print `hyp status` at the end of every turn.

## Work

Act on what OWED in `hyp status` says the runner owes, and on nothing else. Plan only the next Experiment.

- Use the Goal's must-fail cases as known-answer flaw cases, in the owner's terms.
- When the Goal body names a base record under `.tree/evidence/`, the red-on-base Experiment pins it with `--records` and does not run the base again.
- The recipe's checks may exist before the Experiment, because `hyp add exp` pins them. Write and run no subject under test until `hyp add exp` has recorded the Experiment with its rubric.
- Try a new or changed check only on its cases' stand-ins, never on the subject under test, until that Experiment is recorded. If the subject under test shows a check is wrong anyway, do not just fix the check: write the case that shows it, in the Goal's terms with the outcome it must give, in a new version, and name that case in the Experiment body. A sentence that discloses a trial run pins nothing.
- Before you pin a recipe whose Drive runs an agent, run it once on one case while it is still a Draft, and read that run's transcript for reads or writes outside the subject. A Draft may be edited in place. A pinned recipe changes only by a new version, and every trial already run on it is lost.
- When an Experiment counts trials, its body states, before the run, the trials per cell (one case on one candidate), the threshold, and that a cell whose miss alone would contradict a Hypothesis is rerun once before the Result is recorded, with how the two runs together decide it. A pinned rubric reads both runs' records, and the Result cites both. A lone miss can be noise: at 2 of 3 per cell over 9 cells, a candidate that passes each cell 85% of the time is contradicted about 4 times in 10 by chance.
- `hyp record` and `hyp check` warn when a Result's text disagrees with its cited files: a verdict that a cited `verdict:` line contradicts, or a commit or hash in the observation or a limit that no cited file holds. Fix each one with a new Result that cites the file; a Result cannot change after it is recorded.

## Build Goals

A build Goal's candidate is a commit, and "candidate" in this section means that commit only. Two rules of "Work" read differently here: the candidate is written before its Experiment and only its gate waits for `hyp add exp`, and the stand-in a new test is tried on is the base. A round runs in this order:

1. **Write the commit, with its tests.**
   - A build candidate is written first: the commit, with its tests, exists before its Hypothesis. What waits for `hyp add exp` is the gate. Run no red check, no removal check and no review of that commit until its Experiment is recorded, and write no sentence in a body to disclose the build.
   - Write the tree with the installed hyp. Compile the candidate to a separate path, and never replace an installed binary.
2. **Add its Hypothesis and tie its subject.**
   - Each commit you test is one Hypothesis under the Goal's question, with the branch and the commit in its title. Tie its subject before the run with `hyp subject <H> --base <commit>`. A fix is a new commit and a new Hypothesis that supersedes the old one (`hyp supersede <new> --old <old>`).
3. **Record its Experiment, with the gate as its rubric.**
   - Pin the gate as a recipe when the repository keeps one. A recipe measures and never prints a verdict, so pin with it a rubric from `rubrics/` that turns its measures into verdicts. When the repository keeps no gate recipe, the Experiment on a commit uses `--recipe none --rubric reading`: its body states the gate as its rubric, and the Result cites the gate's logs.
   - Before `hyp add exp`, run each new test on the base, which is a stand-in and not a candidate, and write what it gives into the rubric: red, or green for a line the base already meets, where the removal check is what shows that the test bites. For a successor's new test, the red run is on the candidate it supersedes.
   - Before you run the gate, fix each Experiment's rubric: prose in its body with `--rubric reading`, or a pinned rubric bundle. Its body states the Goal's lines it aims at; for each new test, what it gives on the base, and green on the candidate; a removal check for each refusal (remove the refusal and show which test fails), where a removal whose pattern does not apply whole is a failed removal; `go vet ./...` and `go test ./... -count=1`, or the repository's own gate; and the review. For the review it names the diff command, `git show <commit>` with the candidate's commit, and states that a test the review returns that is a case of a line and red on the candidate contradicts the candidate. For skill text, run the original and the candidate skills on the same eval cases.
4. **Run the gate and the review side by side.**
   - Keep the logs under `.tree/evidence/<E-id>/` and cite them in the Result.
   - Review each candidate before any select of it, in its gate Experiment, so the review costs no extra Result. Ask a fresh subagent for one review of the candidate commit's diff (`git show <commit>`) against the design and the Goal's lines: each candidate's review reads its own commit, because the commits before it were reviewed as candidates of their own. Write the brief from the Experiment's rubric, not from memory, and give every candidate's reviewer the same brief apart from the commit: the code, the lines and the keys already settled, with the owner's words. Do not give it your conclusion, a list of places to probe or a finding of an earlier review.
   - Ask the reviewer to return each finding as a test file with its output on the candidate commit, run in a checkout of that commit that is not the run's branch, and to label each with the line it is a case of, or none. A finding it could not run, it says so, and why.
   - Save the subagent's reply word for word under `.tree/evidence/<E-id>/`, never your summary of it, and cite that file in the Result that gives the candidate its verdict. A candidate is reviewed only when such a Result exists. Select no candidate that is not reviewed.
5. **Run each returned test, sort the findings, and record the Result.**
   - Run each test the review returned yourself, on the candidate, before `hyp record`. Then sort each finding by what its test showed:
     - **Red, and a case of a line.** The verdict is `contradicts`, because the candidate does not meet the line completely. Record the case as a `--limit` on that Result, so the ledger owes it on the successor.
     - **Green.** Nothing: the finding was wrong.
     - **Not run, only described.** No limit. It stays in the saved review file, and you name it in the stop checkpoint.
     - **A case of a line that no test within the Goal can run.** A `--limit` you refer, with the reason it cannot run.
     - **Red, and the line's words do not say whether the line covers the case.** A reading. The case is in doubt when the words leave it open, and also whenever you and the reviewer place the finding differently. Do not settle it yourself either way: write no fix for the case in this run, and do not sort it out of the Goal. When cases in doubt are the candidate's only failures, its verdict is `inconclusive`: it is not known to meet the Goal until the owner answers. Stop on that question and plan no further Experiment for it, although OWED lists an `undecided` item, and say in the stop checkpoint that the candidate waits for the reading.
     - **Red, and a kind of case that the line's words, the Goal body or an owner's answer says no line covers.** An out-of-Goal finding, which you triage as "Out-of-Goal findings" below. Name those words, or the answer's key, in the checkpoint.
   - A `supports` Result records no limit whose case ran red on its own candidate. A limit's text names the case, never the outcome.

The run may push its branch and open or update one PR. It never merges.

## Out-of-Goal findings

An out-of-Goal finding is a case of none of the Goal's lines and asks no question about what a line means. It gets no limit, no refer and no reading, and it never holds the select. A finding the reviewer labelled with a line is never sorted here by you alone: when you place it differently, it is a reading. Give each one the first destination that fits:

1. **Check that it is a defect.** Cite the file and line, or the run, that shows it. If the code or the run does not show it, drop it, and give the reason in one phrase in the checkpoint.
2. **Fix it on the run's branch** when the Goal's `May change:` lines let you edit what the fix touches: one commit, before the stop. On a build Goal it is a fix like any other, so a new Hypothesis that passes the gate and the review before any select.
3. **Fold it into a known issue** when an open issue in the repository's tracker already covers it: name the issue and write the comment that adds the finding.
4. **Propose a new issue** only when no open issue covers it: write its title and body, with the evidence from step 1.

You post nothing for steps 3 and 4: the owner decides what enters the tracker, so you write the comment or the issue and the owner posts it.

## Before a trade-off

Before a Question goes to the owner as a trade-off, add one structure that is not among the candidates, and the cheapest variant of each contradicted rival that would hold. A variant must pass the difference test against the contradicted rival it varies: "on <scenario>, it meets line <n> and <that rival> does not, because <mechanism>". The new structure states in its Hypothesis body the scenario on which its outcome differs from each candidate's, and the mechanism. Run them through a new version of the pinned recipe and record the Result. If a new structure holds every check, say in a checkpoint that the rubric does not decide this Question.

## Readings

- When a verdict rests on your reading of a Goal line, write the other reading as a strict check: a separate named command in the recipe, with its own known answers, that sets no verdict. Say so in the Experiment's body, and run it on the same candidates. Record the reading with `hyp record --reading`, naming the check and which candidates fail it. If the recipe cannot test the other reading, say that in the reading.
- When a verdict rests on the word of a reading that is still open under AWAITING OWNER, or on an owner answer another Result already reads, record the same reading on that Result too, word for word, with the same `--reading-of` key if it has one: a reading holds only its own Result's Question, and hyp prints identical readings as one question and warns when the owner answers only one. Record a copy only on a Result that judged a Hypothesis whose verdict differs from the strict check's outcome for it: a `supports` the check fails, or a `contradicts` the check passes. A Result whose judged Hypotheses the check gives the same outcome as the verdict gets no copy. The owner answers every copy in one answer, because `--answer` is repeatable.
- When the doubt is whether a line covers a case, ask about the kind of case, not the one scenario: name what sets it apart from the cases the line's words plainly cover, such as how the input is reached, which state it starts from or which version wrote it, so one answer places the next finding of that kind. Before you record such a reading, read the owner's answers in the history `hyp show <goal-id>` prints; when an answer already places the case, sort the finding by that answer and record no new reading.
- Before you record a reading, look for a test that can decide it. If a test fits the Goal's lines, write it, run it under the Experiment's rubric, and record no reading. If a test needs a new Goal line, name the test, the fact it measures and the line in the reading, and ask the owner whether to run it. Ask the owner to choose a meaning only when no test can decide, and say why none can.
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
2. For each Question, add one Experiment that names each answer it applies with `--aims`, and fix how its verdicts are decided before you run it: a rubric from `rubrics/` (`--rubric <name>@<version>`, a new rubric version when the answer changes it) or `--rubric reading` with the rubric in its body. A recipe measures and never prints a verdict. To reassess records an earlier run kept, pin them with `--records .tree/evidence/<E-id>` and `--recipe none`. If an outcome is already known, say so in the rubric. Give a verdict on every target, `inconclusive` when the run could not decide. Record its Result with `--applies` for each answer it applied. Old Results stay as they are.
3. If applying an answer raises a new question about what a Goal line or the owner's answer means, record it with `hyp record --reading` on that Result. A Result cannot change after it is recorded, so if it is already recorded, record a new Result under the same Experiment.
4. Act on each `limit` line under OWED, whichever Hypothesis its Result gave a verdict on: a limit is a case the Goal covers that no check ran. If a test within the Goal's lines can run the case on every candidate, add an Experiment with `--aims` at it, run it, and record `--closes` on the Result that ran the case. Before each `hyp record`, read the `limit` lines under OWED and add `--closes` for every one whose case this run ran: a Result cannot change, so a missed close costs a Result. A close records that the case ran, whatever the verdicts: a candidate that fails the case is contradicted and the limit is still closed, so never write a rubric that closes it only on a pass. A close covers every case the limit's text names; `hyp record` prints that text when it closes. Record each case you did not run as a new `--limit` on the closing Result, and refer that new key when its case is void or outside the Goal's lines. A claim that a case is void, out of scope or covered elsewhere is a limit and a refer, never a sentence in a body or an observation. When that claim rests on your reading of a Goal line, put the reading in the refer's text. Write one case per `--limit`. If no such test can, or the case is void for every candidate, refer it with your reason: `hyp checkpoint --refers "<key>=<why no test within the Goal can close it, or why the case is void>" --why "..."`; hyp refuses a refer with no reason. When the owner returns a referred limit, it is an ordinary `limit` line again.
5. End with the candidates left under each Question: none, one, or several, by alias.
6. Select only when OWED lists `select` for that Question: `hyp select <Q-id> <H-id> --why "..."`. Never confirm a select. Do not change the Goal. Add new structures only as "Before a trade-off" says, or when the owner asks.

## Goal lines

- The Goal's lines are limits keyed on the Goal, `G-<id>#<n>`; `hyp status` prints them under LINES. A Goal's lines come with its add and no later event adds one. If the Goal's body states lines and none is typed, they cannot be added to this Goal: tell the owner, who may add a new Goal with them (`hyp add goal … --line "<line>"`). Do not write them for the owner.
- A line is owed as a `limit` on each top-level Question with a candidate no closer covers. Run its case like any limit: aim at it with `--aims G-<id>#<n>`, and record `--closes G-<id>#<n>` on each Result that ran the case. A Result covers a candidate it gave a verdict on, or one whose whole subject (base, patch and files) one of the Result's cited runs holds, so tie each candidate before the run: `hyp subject <H> [PATH…] --base <commit> [--patch FILE]`.
- A line is never edited. What a line means is a reading. A line no test within the Goal can run is a refer, and the owner accepts or returns it. A new line or a new purpose is a new Goal, so a new tree: tell the owner, and do not write it.

## Stop

Stop on the clauses the `/goal` prompt names. Stopping is not success.

1. Before your last message, record a runner checkpoint, so the fork does not live only in chat; the owner reads `hyp status`, which prints the last checkpoint, each reading in full and what each copy's answer can touch. Do not restate a reading's meanings in the checkpoint:

   ```text
   hyp checkpoint --goal <goal-id> --summary "Stopped: <stop clause>. Candidates not reviewed: <each by Hypothesis id and commit, or none>. Waiting for the owner: <the keys under AWAITING OWNER>. Review findings with no run: <each as one short phrase, or none>. Out-of-Goal findings: <each as one short phrase and its destination: fixed in <commit>, issue <id>, new issue, or dropped, or none>" --why "<why>"
   ```

   Name each out-of-Goal finding in the checkpoint so it lives in the log, not only in chat. Then run `hyp sync`, so the remote's record holds every event of the run, at every stop; if it exits non-zero, say so in your last message.

2. Your last message repeats the checkpoint, and lists each line under LINES by its key, with the Results that ran it, and names each question where OWED still owes a line. A run that ends with lines owed is a normal result.

3. Then list every owner item with the exact command the owner runs: each item under AWAITING OWNER, each out-of-Goal finding with what its destination asks of the owner (a fix commit to check, or a comment or issue to post, with its text), merging the PR, reinstalling hyp, and each personal skill that names a command the change altered.

## Never

- Never change the Goal.
- Never record a choice the owner reserved, and never record an owner answer unless the owner gave those words in this session.
