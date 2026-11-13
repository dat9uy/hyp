---
name: hyp-owner-answer
description: "After a /goal run stops at the owner, walk the open readings and referred limits of the hyp tree with the owner: explain each one, answer questions, and record only the owner's answers."
disable-model-invocation: true
argument-hint: "[reading or limit key, question or answer]"
---

The owner's input: $ARGUMENTS

A `/goal` run stopped and the owner now reads what waits for them. Use the tree of this session. If this session has no tree, ask for its path. Read the Goal with `hyp show <goal>`, including its lines and checkpoints, and read `hyp status`. In a store with many Goals, `hyp status` prints a GOALS table and one Awaiting owner list grouped by Goal; read `hyp status <goal>` for one Goal's lines and OWED. The owner's acts are `hyp answer` events: each item is `KEY=TEXT` with the owner's words, and an answer has no `--why` and no summary. One answer may name keys of several Goals in the store.

Write in ASD-STE100 Simplified Technical English: short sentences, active verbs, one meaning for each word, and the Goal's words for things. Name a Hypothesis by its alias. This rule covers each part of what you show, also the explanation of a quote. Quote the runner's or a recipe's words only briefly, then explain them in STE. Define each tree or recipe word the first time you use it, for example a check name, "structure", "shape", "host" or "stand-in".

## What the owner decides

Each owner move turns the direction of an item so that it serves the Goal, or starts a new Goal:

- A reading answer sets what a line or an answer means, so the runner judges by it.
- An accept says this tree does not run a case. A return sends the case back to run.
- A choose selects one of the candidates that meet every line, for the owner's own reasons. The runner records the select in the owner's words.

The Goal's lines are the owner's one move that no item asks for. Each other move answers what the ledger hands the owner under AWAITING OWNER.
- A new Goal is a new tree, for a purpose this Goal does not hold.

The owner never runs a case, never judges a verdict and never edits a recorded event. The Goal's lines are written once, so this tree never gets a new line. A recipe check is not a line: it counts toward a verdict only through a line or an owner answer, and the runner maps checks to a line when it closes it.

## Show each open reading

If the input is empty, show each open reading in this form. If the input names a reading, show only that reading.

- **Line:** the Goal line or owner answer, in its exact words.
- **Meaning 1 and Meaning 2:** state what the ambiguity is between. Give one small example of each meaning.
- **Effect:** which candidates fail under each meaning, and which meaning the run used.
- **Trade-off:** what each meaning makes easier or harder. Mark this as not tested if no check tests it.
- **Test:** look for a test first. Name the test that can separate the two meanings or the candidates they affect, the fact it measures, and the Goal line it needs, if any. If no test can decide, say why. Only then is this a choice for the owner.
- **Question:** one question the owner can answer in their own words. If a test can decide, the question is whether to run that test.

Do not ask the owner to decide what a test can decide. Do not recommend a meaning. A question that the owner must answer goes in the tree, not only in this chat.

## Do what the owner's words ask

- **A question.** Explain from the tree and the recipe. Write nothing to the tree. The reading stays open.
- **"I don't know", or no decision.** Leave the reading open. Do not close it, because a closed reading sets the meaning of the line. It blocks nothing. Propose a test that can give the owner the facts, and name the Goal line it runs under.
- **A test.** If the test fits the Goal's lines, run it as in "Apply the answers". If no line holds it, it cannot decide here: show the owner the fork in "A test outside the Goal's lines". The reading stays open until the owner answers it.
- **An answer.** Answer the reading with their exact words, under the actor of their earlier owner acts. When `hyp status` prints several keys on one reading line, they ask the same question: name every key in the same answer, with the same words. A `warn: … asks the same as …` line means a copy was left behind; answer it with the same words:

  ```text
  hyp answer --answer "<R-id>#r<n>=<their words>" [--answer "<R-id>#r<n>=<their words>" ...] --actor <owner-actor>
  ```

- **A new answer to a reading they already answered.** `hyp show <R-id>` prints `closed by owner checkpoint …, answer: "…"` for it. Give the new answer with the same command and their new words, naming every key that asks it; hyp derives that it is a revision from the reading's state:

  ```text
  hyp answer --answer "<R-id>#r<n>=<their new words>" [--answer ...] --actor <owner-actor>
  ```

  If hyp prints `warn: … revises …, which G-…#n cites`, a line of another Goal was written on the earlier answer; tell the owner.

  A revision changes no earlier event or verdict. The earlier answer stays in history, and the runner owes the reading again as `apply answer` under OWED.

The owner names the reading: "Answer <key>: …" answers a reading that waits under AWAITING OWNER, or gives new words for one already answered. If the owner gives words that answer a reading but do not name it, ask which reading they answer before you write; an answer must name its key. If you are not sure that the words are an answer, ask. Show the owner each write under their actor before you run it.

## Show each referred limit

A line `Q-… limit referred, accept or return R-…#n` under AWAITING OWNER is a limit, and one with a `G-…#n` key is a line of the Goal: a case the Goal covers that no check ran on the Hypotheses its Result judged. The runner referred it because it says no test within the Goal's lines can run the case, or the case is void for every candidate. If the input is empty, show each referred limit after the readings. If the input names a limit key, show only that limit.

Run `hyp show <key>` first; for a line, `<key>` is `G-…#n`. It prints the limit's state, what its Result judged and their status now, the candidates, the Experiments that aim at it, and the runner's refer summary. Show the limit in this order:

- **In plain words:** what the tree knows and what it does not know about the case, in STE and in the Goal's words. Give one small example that passes and one that fails. Then quote the limit's text and the runner's refer summary briefly, in their exact words. Say which Goal line or owner answer the refer rests on, if it names one.
- **Ran already?** If an Experiment aims at the limit and has a Result, the case may have run without a close. A close records that the case ran, whatever the verdicts, so such a limit is the runner's to close, not the owner's to answer. Tell the owner, and do not ask them to accept it.
- **What each answer can change:** a table with one row for each outcome. Use the candidates and the judged Hypotheses with their status now, as `hyp show` prints them:

  | Answer | Result of the run | Ends with |
  |---|---|---|
  | Accept | nothing runs | what is selected, or what the owner chooses between |
  | Return | the case fails on <Hypotheses> | … |
  | Return | the case holds on <Hypotheses> | … |

  A contradicted Hypothesis is not tested again, so a case that can revive it runs on its successor; name the successor in each Return row it affects. Then write one STE sentence that names the rows that change the outcome, or says that no row changes it. Mark each row that rests on your reading of the limit's text, and not on what `hyp show` prints.
- **A return also needs:** each thing that running the case needs and that only the owner can ask for, for example a new structure or a new Hypothesis. If the case needs a line this Goal does not have, it is not a return: show the fork in "A test outside the Goal's lines". If it needs nothing, say so.
- **Test:** name the test within the Goal's lines that can run the case on every candidate, and the fact it measures. If no test can, say why.
- **Question:** one question in the owner's words: does this tree run the case or not?

Do not recommend an answer. Then do what the owner's words ask:

- **Accept: this tree does not run the case.** hyp admits an accept only of a referred limit, and only under an actor other than the one that referred it. Record their exact words. Name every limit key they accept:

  ```text
  hyp answer --accept "<key>=<their words>" [--accept "<key>=<their words>" ...] --actor <owner-actor>
  ```

  An accept means only that this Goal does not run the case. A later Goal carries the case by citing the accepted key on one of its lines.

- **Return: the Goal covers the case, so run it.** Record their exact words. Name every limit key they return:

  ```text
  hyp answer --returns "<key>=<their words>" [--returns "<key>=<their words>" ...] --actor <owner-actor>
  ```

  If "A return also needs" named something, ask the owner for it in the same answer, and put their words for it in the return's text, so no second ask follows. A returned limit is the runner's `limit` line under OWED again, and it holds the select until a Result closes it.
- **A question, "I don't know", or no decision.** Explain from the tree. Write nothing. The limit stays referred. While it is referred, it holds the select and the choice on its question, so tell the owner that the question waits for their answer.

If the owner's words answer a limit but do not name it, ask which key before you write. hyp refuses `--returns` on a limit that is not referred or is already closed; tell the owner what `hyp show <R-id>#<n>` prints. Show the owner each write under their actor.

## Apply the answers

Apply only when the owner asks you to reassess, or to update the tree. A closed reading changes no status until a new Result records the verdicts. A returned limit is applied as hyp-goal-run's Reassess step 4 says for a `limit` line.

Reassess as the hyp-goal-run skill's "Reassess" section says; it is the one copy of these steps, shared with the `/goal` run. The owner is in this session, so show the answer lines of step 1 and wait for their yes.

Run `hyp check` and `hyp bundle check` after the Results. If the run's instruction expects new structures or variants, ask the owner.

## Show the candidates

If no reading is open, show each open Question with the candidates left, by alias.

- **None.** Say so. Each rival failed, or has no Result yet.
- **One.** It is the suggested select. Show each `limit` line under OWED for it. Run a test within the Goal's lines that can close one, if the owner asks you to reassess. Show each referred limit under AWAITING OWNER as "Show each referred limit" says.
- **Several.** `hyp status` prints `Q-… choose: …` under AWAITING OWNER only when the runner owes nothing on that Question. Until then, apply what the runner owes first.
   The owner can choose one by their own reasons, in their own words, or ask for a test or a question that rules more out. Name the test that would separate them, if one exists. If that test is under no Goal line, show the fork below.

## A test outside the Goal's lines

The runner or a recipe can name a check that separates candidates but that no Goal line holds, for example a check the recipe prints and counts for no verdict. That check cannot decide in this tree: a verdict on it would drop a candidate for a case the Goal never asked for. Show the owner the two ways, with what each costs, and do not recommend one:

- **A preference between designs that both meet the Goal.** The owner chooses. The check can run as one Experiment that measures it, and its Result is evidence for the choice, not a verdict. The owner's choose names the reason in their words.
- **A requirement this Goal lacks.** It is a new purpose, so a new Goal in the same store: the owner adds it with `hyp add goal --title … --line "<the new case>" --cite <n>=<key> --actor <owner-actor>`, citing the settled keys of this Goal it builds on, and every case runs again under it. A move never crosses Goals: in this Goal, the runner refers each item the new Goal takes over, and the owner accepts it with words that name the new Goal.

Ask one question in the owner's words: would they reject a design that fails this check? Yes points to a new Goal; no points to a choice. Do not write either until the owner says which.

## End each turn

If this turn recorded a Result or changed what a reading does to the candidates, record a runner checkpoint before you end, under the runner's actor, not the owner's:

```text
hyp checkpoint --goal <G-id> --summary "Stopped: <what waits>. Waiting for the owner: <the keys under AWAITING OWNER>" --why "<why>"
```

The owner reads `hyp status`, which prints the last checkpoint, so the fork must not live only in chat. A checkpoint holds only a summary and refers; the owner's acts are `hyp answer`.

State what changed in the tree, what is still open, and the next question for the owner. Print `hyp status`.

When the owner's answers leave work for the runner, the next run starts from a fresh `/goal` prompt. Print it with `hyp prompt <goal> --actor <new runner actor> [--branch <branch>] [--pr <N>]` and give the owner its output as it is. Do not add the owner's answers to it: OWED lists each answer and each returned limit that the next run applies. Run it on the log the next run starts from, with the answers merged or on the branch the prompt names. If it exits 1, the Goal body has no valid constraints block: tell the owner, who adds a new Goal. Exit 2 is a bad argument; exit 3 means no repository holds the store.
