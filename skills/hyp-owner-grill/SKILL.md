---
name: hyp-owner-grill
description: "Grill the owner about a purpose until it is a hyp Goal a /goal run can finish: route each question (owner-only into a line, rivals a run could separate into a DesignQuestion, facts looked up), derive the lines from positive and negative space, split purposes into separate Goals, and record the owner's writes only on their approval. Use before a /goal run, when the owner wants to write a new Goal; running a Goal is hyp-goal-run, answering a stopped run is hyp-owner-answer."
disable-model-invocation: true
argument-hint: "[the purpose, an issue or a design to turn into a Goal]"
---

The owner's input: $ARGUMENTS

This skill turns a conversation with the owner into a hyp tree that an unattended `/goal` run can finish: a Goal with its lines, the contested questions under it with their rivals, and nothing the run must come back to ask. It ends where the hyp-goal-run skill starts. Use it with the hyp-prototype skill for the commands that add nodes.

The grill is not done when the owner and you agree. It is done when every question is in its place in the tree, the owner approved each owner write, and nothing is left that only the owner can answer.

## Start

1. Find the store. Run `hyp status` at the repository root. A new purpose in a repository that has a store is a new Goal **in the same store**: one store holds many Goals. If the installed hyp refuses a second Goal in one log, the new Goal is a new tree under its own directory. Run `hyp init` only when there is no store.
2. Set `HYP_ACTOR` to a runner name unique to this session. Ask the owner for their actor name (lowercase letters, digits and hyphens, for example `owner-dat`) once.
3. Read the glossary of the repository, if it has one. Use its words. One concept gets one word, and every sense of the word has one shape. Before you name a thing by how it is stored (`bases`, `commits`), look for the domain word.

## Rounds

Work in rounds. A round asks every question whose prerequisites are settled, numbered, each with your recommendation. Wait for the answers, then compute the next round.

Every question opens with what exists today and the problem it settles, then gives one small worked example of each option and the principle behind it. A choice without its basis is not a question the owner can answer. Name why each mechanism is in the question; if you cannot, drop it.

Run the mistakes checklist below on each round before you send it, and again before Confirm the routing.

## Mistakes checklist

Each item is a mistake earlier grills made. Check the round or the table against every item, and fix what fails before the owner sees it.

1. **Store.** The Goal goes where the repository's agent instructions (AGENTS.md, CLAUDE.md) say trees live, such as `investigations/<topic>/`: in the store there that holds the Goal of the issue's parent issue, when one does, or else in a new store in that directory. No `hyp init` at the repository root unless the instructions name no directory for trees.
2. **Test method.** Where an installed skill owns the test method (for a skill, ak-skill-creator's eval flow: its cases, fixtures, matching and grading), the round asks only whether to use that flow and the owner's preferences of model, trial count and budget, never a choice among the method's parts.
3. **Constraints.** `May change:` covers the texts the change makes wrong by their directories (`skills/`, `docs/`, `AGENTS.md`, `README.md`), not as a list of files or parts. No `Never:` line forbids what `May change:` or the Goal's test route needs, such as a new recipe version under `recipes/`.

## Route each question before you ask it

| The question is | Its place | Ask the owner? |
|---|---|---|
| a fact about the code, the stores or the history | look it up yourself or with a subagent; check a subagent's fact before you rely on it | never |
| a choice between options a run could separate | a DesignQuestion with at least two rival Hypotheses, and the scenario that separates them in the body; this holds also when the owner settles it in the grill | never as a choice; ask only whether the run should test it |
| what the purpose must allow or must never allow | a line of the Goal | yes |
| the purpose, its scope, who asked and why | the Goal's title and body | yes |
| what a run may change, may only read, or must never do | the Goal body's `May change:`, `Read only:` and `Never:` headings | yes |
| how the tree is run (process, review, order, cadence) | the Goal body, an Experiment's verdict rule, or the `/goal` prompt; never a line | only where it is a preference |
| a purpose outside this Goal | a separate Goal, added by the owner | name it; the owner decides |

A testable option set is never put to the owner as a choice. If the owner settles such a question in the session anyway, it is still a DesignQuestion with its rivals, recorded as decided with the session as evidence (see Writes), so the tree says it was argued, not tested. A question becomes a DesignQuestion only when you can name the scenario a run would execute to separate its rivals, such as two Goals in one store, or a base that moves mid-run. A decision with alternatives but no such scenario, such as a preferred shape, bound or word, is the owner's: it is a line or body text, never a question recorded as decided. One DesignQuestion holds one question, and every rival under it answers that question; two grill questions are two DesignQuestions.

## The derivation guard

Before you propose any stored field, flag or event, ask: can replay derive this from the log at that position? If it can, do not propose it. A stored copy of a derived fact, such as a relation label beside a key or a role beside an actor, can disagree with what it copies. This mistake recurs more than any other in a design grill; check it on every field, before the owner has to.

## Derive the lines

Lines are the Goal's positive and negative space, written once and never edited.

- **One line per invariant, not per feature.** List each state the purpose must allow (positive, `[P]`) and each state it must never allow (negative, `[N]`).
- **Negative space comes from what the grill rejected.** Each rejected option that would produce a reachable state is a forbidden state, so it is an `[N]` line. A rejected option that is only a shape, such as an unknown field a strict decoder already refuses, gives no line.
- **Merge states one invariant guards.** At most 10 lines of 1 to 600 runes each; the cap forces the merging.
- **A line carries its own cases.** A line that compares output with a baseline names the baseline by sha256 and asserts that it differs from the candidate; other lines name no baseline. Name each exception it allows in its own words, not only in a later prompt.
- **A line may cite settled keys** (`--cite N=KEY`): a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis.

**For a build Goal** (the candidate is a commit), a line asserts one property of the candidate, checked by its own tests or one pinned instrument. `[P]` pins what a valid input produces, exactly. `[N]` pins that the forbidden state is refused at write, with exit 1 and a message that names the rule, and that a log holding it is rejected on replay. Before you write a line, it must pass six questions. Does it:

1. name only the candidate's behaviour?
2. get decided by the candidate's tests or one pinned instrument?
3. use an input that can be built on the candidate?
4. need only runner acts on the branch?
5. state one invariant in the P or N form?
6. fail on some commit that could be named?

A line that names history ("fails on the commit before"), the process, an owner act or a reading fails. Red on base, the removal check and "each rule has a test pair" are facts about the tests, so they go in each Experiment's verdict rule, never in a line. A runner cannot meet a line about history, and it will stop on it.

**For a discovery Goal** (the candidate is a claim, measured by a recipe and a rubric), a line names a case every candidate must meet that a check can run on every candidate.

**A Goal whose questions are all settled by argument in the session** (a design record) gets no lines: no run checks its candidates, so a line there is owed forever. Its cases go to the lines of the Goal that builds the design.

## Split purposes into Goals

hyp owes every line on every top-level question that has a candidate. So draw a grid: lines down the side, top-level questions across, and in each cell ask "could one run cover every candidate of this question for this line?" A cell no run could ever fill means the question and the line serve different purposes: put them in separate Goals. A quick proxy: group the top-level questions by what their candidate is (a claim argued with stand-ins, a commit in repository X, a commit in repository Y). Each group is one Goal. A design record and the build of that design are two Goals.

## Before the round closes

1. **A second reader.** Ask an independent reviewer, such as a fresh subagent on the strongest model, to walk every settled rule through a scenario the grill did not draw: two Goals in one store, a merged log from two actors, a base that moves mid-run. Give it the rules and the code, not your conclusion. Check each fact it reports against the source before you rely on it.
2. **Apply the model to itself.** If the Goal changes the tool or rules that will judge it, check its lines against that change.
3. **A preflight scan.** Read every store and text the change touches for a claim the run would contradict, such as a design that says "no store holds X" while one does. Each owner decision a scan can raise is asked now, not by the run.

## Confirm the routing

Run the mistakes checklist on the table first. Before any write, show the owner one table with a row per grill question: the question, where it goes (line, Goal body, DesignQuestion with its rivals, looked up, another Goal), and for a DesignQuestion the scenario that separates its rivals. Mark each line `[P]` or `[N]` and the invariant it states. The owner confirms each row or moves it. Which settled questions the tree records as DesignQuestions is the owner's call; the table is where they make it. Write nothing before they confirm.

## Writes

The grill ends in these writes, in this order. Show each owner write as the exact command with the owner's actor name; run it only after the owner approves that command in the session. Their approval is the owner's act; you only type it. Never write under the owner's name without it.

| Output | Command | Actor |
|---|---|---|
| a new store, if none exists | `hyp init` | runner |
| the Goal, its purpose and its lines | `hyp add goal --title "…" --body "…" --line "…" --cite N=KEY --why "…"` | owner |
| a contested question | `hyp add dq --parent <G-id> …`, then one `hyp add hyp --parent <Q-id> …` per rival | runner |
| a question with rivals the owner settled in the session | the DesignQuestion and each rival as above; then under the chosen rival `hyp add exp --recipe none --rubric reading --target <each other rival>` whose body states the scenarios as its rule, `hyp record` with a verdict per rival citing a scenarios file under `.tree/evidence/<E-id>/`, then the select OWED owes (a runner write, not the owner's) | runner |
| a link to another Goal or an issue | `hyp link <id> external_ref …` | runner |
| the end | `hyp checkpoint --goal <G-id> --summary "…" --why "end of grill"` | runner |

The Goal body ends with the run's constraints block, so a run and `hyp prompt` can find it:

```text
May change: <the repository, how it is changed (for example commits on one local branch, a PR never merged), and the texts the change makes wrong>
Read only: <repositories, directories or data the run may read but never change>
Never: <acts the run never takes, such as replacing an installed binary or publishing source data>
```

The block runs from the first line that starts with `May change:`, `Read only:` or `Never:` to the end of the body. It holds each of the three headings at least once, and every non-empty line in it starts with one of them, so write one constraint per line and put no other text after the block. A line of spaces, a tab or a lone carriage return is not empty: leave a blank line truly empty, and write the body with `\n` line ends. A heading may repeat, such as one `Never:` line per act. `hyp add goal` warns when a body has a heading but no valid block, and `hyp prompt` refuses that Goal.

A body that names another tree's ids writes each as `<tree>:<id>`. Then run `hyp status <G-id>` and `hyp check`: it must print `check: 0 errors`. Commit `.tree/events/` and `.tree/evidence/`.

## Hand off

hyp prints the `/goal` prompt for hyp-goal-run from the log. Choose a new runner actor, then run:

```text
hyp prompt <G-id> --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
```

It prints the store's path, the actor, the Goal id, the skill, the stop clauses (the runner owes nothing, or N Results under the actor, 5 by default) and the Goal body's constraints block, line for line; `--branch` and `--pr` add a clause each after the block. Give the owner its output as it is, and do not add to it. The prompt carries no owner answer: OWED lists each answer and each returned limit the runner applies. The practice of a run, a build Goal's review round included, is in hyp-goal-run. Run `hyp prompt` on the log the run starts from: the prompt names the Goal by its id, so the Goal's add must be in that log, merged or on the branch the prompt names. If `hyp prompt` refuses (exit 1), the Goal body has no valid constraints block; a Goal is never edited, so tell the owner, who adds a new Goal with a valid block. Exit 2 is a bad argument, such as a `--branch` with a space; exit 3 means no repository holds the store, so run it from the store inside its repository.

## Never

- Put a testable choice to the owner as a choice.
- Propose a stored field replay can derive.
- Write a line about history, process, an owner act or a reading.
- Add a purpose this Goal does not hold as a line; it is a new Goal.
- Write under the owner's actor without their approval of the exact command.
- End the grill with a question only the owner can answer still in chat and not in the tree.
