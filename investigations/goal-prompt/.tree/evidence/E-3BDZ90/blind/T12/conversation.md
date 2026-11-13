# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-pre-a-1/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-a-1/tmp/issue-113.md"
}
```

Result:

```text
1	# Issue #113: Owner flow: a /goal prompt generated from the tree, and the run's constraints
2	
3	Opened by dat9uy at 2026-11-10T10:27:52Z. Text as of 2026-11-11T16:20:00Z.
4	
5	## Event
6	
7	On 2026-11-10, after the multi-tenant tree completed, the owner asked to replace planning in prose (`ak-plan`) with a `/goal` prompt generated from the tree. The examples are the hand-written prompts of runs 4 and 5 on that tree. From run 3 to run 5 they changed in only three places: the actor, one sentence of context, and the stop clause. Everything else had moved into hyp (OWED and the stop line `runner owes nothing: stop and report to the owner`) and into the runner's skill.
8	
9	On 2026-11-11 the owner settled that **building is a Goal like any other**: hyp has no discovery kind and no delivery kind, so there is no separate delivery `/goal` (#114 is restated to match). The generated prompt therefore serves any Goal, including one whose candidate is a commit. The same session wrote two Goals with a grilling session, which is the owner flow's first step and has no skill yet.
10	
11	The run-5 prompt, with its paths shortened:
12	
13	'''text
14	/goal Run the hyp tree in <tree> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G-id>.
15	Done when the latest printed `hyp status` shows `runner owes nothing: stop and report to the owner` under OWED, or every runner item under OWED is `add rival Hypotheses` on a Question for which no owner checkpoint since the latest stop checkpoint asks for a new Hypothesis, or 5 Experiments have been recorded since this goal began, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
16	Never change <the investigated repository> or its issues; read it only, and work in throwaway copies.
17	Never put application code or real source data in the public hyp repository.
18	Treat <legacy code> as legacy. <...> may be 
[... 3115 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp status 2>&1 | head -40; ls investigations docs; ls .tree 2>&1 | head",
 "description": "Show hyp status and list trees and docs"
}
```

Result:

```text
no .tree store found; run 'hyp init'
docs:
adr
core-and-self-review.md
glossary.md
how-hyp-improves-itself.md
install.md
principles.md
replay-spec.md
sandbox-spec.md

investigations:
checkpoint-store
checkpoint-store-design
id-citation
recipe-base
ls: .tree: No such file or directory
```

### Tool Bash (cwd: .)

```text
{
 "command": "for t in checkpoint-store checkpoint-store-design; do echo \"== $t\"; (cd investigations/$t && hyp status 2>&1 | head -60); done; hyp prompt --help 2>&1 | head -20; git log --oneline | head -30; ls internal cmd/hyp",
 "description": "Status of existing trees, hyp prompt help, history"
}
```

Result:

```text
== checkpoint-store
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A sto…
  G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a …
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and…
  G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Goal's lines come with …
  G-D1CH1R#5 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] One answer may settle key…
  G-D1CH1R#6 covered on Q-PC04ED; ran in R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A refer, aim, close or apply whose …
  G-D1CH1R#7 covered on Q-PC04ED; ran in R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a two-Goal store, the move table…
  G-D1CH1R#8 covered on Q-PC04ED; ran in R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] hyp status G-... prints one Goal, a…
  G-D1CH1R#9 accepted by owner-dat; uncovered on Q-PC04ED: H-P4YGE2 — Each negative-space line has a test pair: a refused write, and a rejected replay of a log h…
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
RECENT
  2026-11-11 14:37 claude-g1b-6 dq.select Q-PC04ED — OWED owes this select: R-RRYBPH on 4047c01 (branch feat/one-subject-per-run) su…
  2026-11-11 14:35 claude-g1b-6 result.record R-RRYBPH — Records E-GAT8N4 under its pre-registered rule.
  2026-11-11 14:33 cla
[... 6174 more characters]
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/docs/glossary.md"
}
```

Result:

```text
1	# Glossary
2	
3	The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.
4	
5	## Language
6	
7	### Trees
8	
9	**Tree**:
10	One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
11	_Avoid_: graph, log (the log is its events), project
12	
13	**Store**:
14	The `.tree/` directory of a repository: one events file per actor and the evidence files Results point at ([spec §1](replay-spec.md#1-names-and-locations)). hyp finds it by walking up from the current directory, as git finds `.git`. In code: package `store`.
15	_Avoid_: database, index (hyp keeps none, ADR 0006)
16	
17	**Event**:
18	One line of `.tree/events/<actor>.jsonl`: an id, a time, an actor, a type from a closed set and its payload ([spec §3](replay-spec.md#3-event-envelope), [spec §4](replay-spec.md#4-event-types-closed-set)). Events are only appended, and every status is derived from them. In code: `event.Event`.
19	_Avoid_: entry, record (a record is what hyp writes for a sandbox or a command)
20	
21	**Actor**:
22	The name a write records as its author, from `--actor` or `HYP_ACTOR` ([spec §1](replay-spec.md#1-names-and-locations)). Each actor appends to its own events file, so logs from several worktrees merge ([spec §6](replay-spec.md#6-merge-order)). An actor is not a role: owner is the role an owner checkpoint names (the verifier role is retired, ADR 0015).
23	_A
[... 51386 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp 2>&1 | sed -n 20,80p; grep -rn \"prompt\" internal/cli/*.go | grep -v _test | head -20; ls investigations/checkpoint-store investigations/checkpoint-store/.tree 2>&1|head; ls docs/adr | tail -8; git branch -a; ls skills; sed -n 1,25p skills/hyp-goal-run/SKILL.md",
 "description": "Check whether hyp prompt exists, ADRs, skills"
}
```

Result:

```text
  link FROM depends_on|external_ref TO   unlink FROM TYPE TO
  select DQ [HYP]                        admitted only when hyp status owes it (no HYP: print the suggestion only)
  unselect DQ | defer DQ | withdraw DQ|EXP | resume DQ
  checkpoint [--summary S] [--goal G] [--refers K[=TEXT]]...   the runner's report: a summary, refers, or both
  answer [--answer R#rN=TEXT]... [--accept K=TEXT]... [--returns K=TEXT]...   the owner's act; no --why

read (REF/TIME: an event id, a node ref, or an RFC3339 time):
  status [--as-of REF|TIME] [--max-lines N] [--live]   orientation screen, 60 lines by default; --live hides contradicted rivals
  show REF|R-id#n|G-id#n [--as-of REF|TIME]  one node, or one limit or line, in full, with its history
  diff --since REF|TIME                       status changes and events since a point
  render --markdown [--as-of REF|TIME]        the whole tree as Markdown, untruncated
  check                                       rejected events, bad lines, warnings
  signals [--json]                            design marks for the whole tree: one-rival, D1
  events --json                               applied events in merge order, and rejected ones, as JSON for scripts

bundle (a recipe in recipes/ or a rubric in rubrics/; a name lives in one library):
  bundle hash DIR [--json]               the bundle hash add exp --recipe or --rubric records for DIR
  bundle check                           every pin to a versioned bundle, in every tree under the repo root, still matches it
  bundle new NAME@1 [--rubric]           write a first recipe, or with --rubric a first rubric, beside the nearest .tree/ from hyp's template; fill every hyp:placeholder
  bundle new NAME@N --from NAME@M        copy bundle NAME@M to the next version NAME@N in its library, writing from:
  bundle run --record DIR REF NAME [ARGS...]   run one Standalone command with no sandbox; exits with its exit code
  bundle evidence --record DIR REF OUT_DIR     run recipe REF's Evidence comman
[... 3797 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rln \"May change\" --include=* . 2>/dev/null | grep -v '^./.git/' | head -20; cat docs/adr/0021-a-skill-change-is-a-build-goal.md | head -60; grep -n \"owner flow\\|owner-flow\\|#113\\|#114\\|#134\" -r docs README.md AGENTS.md | head -20",
 "description": "Find constraints-block references and ADR 0021"
}
```

Result:

```text
(eval):1: no matches found: --include=*
---
status: accepted
---

Supersedes in part [ADR 0007](0007-records-start-at-230.md) (the skill-text gate's replay instrument), [ADR 0010](0010-a-record-keeps-or-identifies-what-a-replay-needs.md) (the follow-up that rebuilds `skill-text-arms` Arms from patches) and [ADR 0015](0015-the-select-is-the-gate.md) (the frozen verifier recipes).

# A skill change is a build Goal

AGENTS.md gated skill text apart from code. A skill-text change shipped only when a session replay from the event's store cut showed the failure in 3 or more of 5 trials without the change and in 1 or fewer of 5 with it, and nothing shipped when no cut reproduced the failure. Its one instrument, `recipes/skill-text-arms@5` with the rubric `rubrics/skill-text-gate@1`, replays the id-citation tree's e-arms with `hyp-prototype` and `hyp-verification` loaded. It cannot reach a skill that runs before any tree exists, such as `hyp-owner-grill`, or one that spans a whole `/goal` run, such as `hyp-goal-run` and `hyp-owner-answer`. For those skills the rule was a permanent freeze, and a new skill needed its own exception. [how-hyp-improves-itself.md](../how-hyp-improves-itself.md) had already found that replay studies of skill text decided nothing, because agents with the skill loaded rarely made the mistakes they counted.

Owner decision on 2026-11-11 (#134): building is a Goal like any other (ADR 0020), and a change to `skills/` is building. The special gate was context that led agents to design gate policy instead of applying that model.

## Decision

- **A change to `skills/`, a new skill included, is a build Goal like any code change.** Its candidate is the commit; its lines state the skill's behaviour; each line is checked by its own test or by an instrument the Goal builds, red on base, and the select ends it. For a new skill, base is no skill.
- **How a skill is tested is that Goal's choice**, settled in its grill. One choice is ak-skill-creator's eval: case
[... 2390 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 1,60p skills/hyp-owner-answer/SKILL.md; grep -n \"prompt\" skills/*/SKILL.md | head -30; cat investigations/checkpoint-store/.tree/events/*.jsonl | head -c 3000",
 "description": "Read owner-answer skill and prompt mentions, first goal event"
}
```

Result:

```text
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
- A ne
[... 8540 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls -la .claude/skills/ .claude/skills/hyp-owner-grill; diff -r .claude/skills/hyp-owner-grill skills/hyp-owner-grill | head -80; sed -n 90,125p skills/hyp-owner-grill/SKILL.md; which hyp; hyp version",
 "description": "Compare installed grill skill with repo copy"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-1 && ls; ls tmp; grep -rl \"goal-prompt-261111\" . 2>/dev/null | head; git -C wt log --all --oneline -S\"goal-prompt-261111\" | head",
 "description": "Locate run-6 prompt file"
}
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -80",
 "description": "Show G-D1CH1R in full"
}
```

Result:

```text
bin
inputs
known-answers.log
manifest.json
recipe.log
recipes
sandbox.json
sandbox.lock
tmp
wt
hyp-owner-grill-eval
issue-113.md
trial-git
./tmp/issue-113.md
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt
```

Result:

```text
Goal G-D1CH1R
alias: goal
title: hyp implements the shared store: one store holds many Goals
body: Builds the design recorded in checkpoint-store-design:G-73R704 (selects on checkpoint-store-design-D906CD, checkpoint-store-design-HC4V9T, checkpoint-store-design-H8PJ20, checkpoint-store-design-R21S8R). Every decision is selected there with session e6a35a24 as evidence, and ADR 0020 cites those Results. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Two-Goal stores land first, so lines 2, 6 and 7 are red on that commit's parent.
line G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A store holding one Goal prints byte-identical `hyp status` and `hyp check` output for id-citation and recipe-base on C and on 33896ec.
line G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a store with two Goals, one Goal's OWED, lines and coverage never include the other Goal's items, and each equals what its own events give in a store of their own.
line G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and a run at B plus a patch never covers a Hypothesis whose subject is bare B.
line G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Goal's lines come with its node.add. A cite not settled when written is refused with exit 1, and a later change in the cited key's state is only a hyp check warning.
line G-D1CH1R#5 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] One answer may settle keys of two Goals. Owner and runner fields never mix in one event, and an empty answer or chec
[... 7556 more characters]
```

## Grill

Checking in: I've read the issue, the glossary, both existing stores and the skills. Here is round 1. Nothing is written yet.

## What I looked up

- **There is no `hyp prompt` yet.** The installed hyp 0.1.0 has no such command, and `internal/cli` doesn't mention it. The grill skill I am running already describes it, but `skills/hyp-owner-grill` in the repo doesn't. So #113 is a build Goal. Its candidate is a hyp commit, with the skill text that goes with it.
- **The skill text and the command must land together.** `go test` includes skillcheck, which fails when a skill names a `hyp` command that doesn't exist.
- **G-D1CH1R can't be the source for the done test.** Its body has no `May change:`, `Read only:` or `Never:` block. A Goal is never edited, so `hyp prompt` would refuse it. The text of run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) is also not on this machine.
- **The issue names no parent issue.** #114 is a sibling it excludes, and the `checkpoint-store` trees are G-D1CH1R's. I found no store that holds a parent Goal for #113.
- **Nothing here needs a stored field.** The prompt is a function of the log and its flags. The `hyp add goal` warning and the `hyp prompt` refusal can share one block check, so neither stores a copy of the other's verdict.

## Questions

**1. Where does the Goal go?**
- Today: `investigations/` holds `checkpoint-store-design`, `checkpoint-store`, `id-citation` and `recipe-base`. AGENTS.md says trees live there.
- Option A: a new store at `investigations/goal-prompt/`, run with `hyp init` there.
- Option B: join `investigations/checkpoint-store`, which holds the Goal of #114's build.
- Principle: a Goal goes in the store of its parent issue's Goal, or else in a new store.
- Recommendation: A, since #113 is not a child of #114. Is there a parent issue I can't see?

**2. One Goal or two?**
- Candidates are `hyp prompt` plus the `hyp add goal` warning (Go code), and the skill text: the two descriptions you folded in, and the hand-off sections.
- Both are commits in this repository, and the skill text names the command, so one run can cover every candidate for every line.
- Recommendation: one build Goal.

**3. Done-means clause 1 cannot use G-D1CH1R.** Run 6's prompt had no constraints block to read from, so what should the done test compare against?
- Option A: a synthetic fixture store under `internal/cli/testdata`. Its Goal body has the block, and a golden file holds run 6's clauses. Only the actor, the owner answer, and the branch and commit differ. A line pins this exactly.
- Option B: drop the clause as unmeetable.
- Principle: core tests read only synthetic data.
- Recommendation: A. I need you to paste run 6's prompt text, because only you have it. The golden file is built from it, with its paths shortened as in the issue.

**4. How do the skill-text lines get tested?**
- ak-skill-creator is installed, and ADR 0021 names its eval flow: cases, fixtures, matching and grading. Do you want that flow for the skill-text lines?
- If yes, what are your preferences for model, trial count and budget? I don't ask you to choose among the flow's parts.
- Recommendation: yes. The command lines are tested by `go test`.

**5. Should the run test "hyp prints it" against "a skill writes it"?**
- The issue allows either. A scenario can separate them: render the same Goal 5 times each way, and check that the block lines are byte-identical to the Goal body and the stop clauses are equal across runs.
- A skill-written prompt can't meet an exact-output line, so the build Goal's lines fit only the command.
- Option A: don't test it. "hyp prints it" is the Goal's purpose in its body, and no DesignQuestion is kept.
- Option B: a tiny design-record Goal in the same store holds one DesignQuestion with two rivals, decided in this session, with an Experiment under the chosen rival whose rule states the scenario and a select. The build Goal cites it. This costs a few runner writes and no lines.
- Recommendation: B. It records that the choice was argued, not tested. It's your call.

**6. Your actor name.** I assume `owner-dat`, as on the other trees. Confirm that, or give another.

## Defaults I will put in the table, unless you object

- **Warning, not refusal:** `hyp add goal` warns on a bad block, so old logs like G-D1CH1R's still replay. `hyp prompt` refuses such a Goal with exit 1.
- **Exit codes:** 2 for bad arguments, 3 for no repository.
- **Flags:** `--branch`, `--pr` and `--experiments`, with a default of 5 Results under the actor.
- **No answer in the prompt:** it carries no owner answer, because OWED lists those.
- **Never:** no typed constraints field on `hyp add goal`, as you decided on 2026-11-11.

After your answers I'll draft the lines, use a fresh subagent for the second-reader pass, and show the routing table before any write. One note for the hand-off: `hyp prompt` can't print anything until this Goal's run builds it. So the first run's `/goal` prompt will have to be written by hand, or this Goal will be run before `hyp prompt` exists. I'll raise that in the table.

## Owner, message 2

1. I lean toward one shared store at the repo root for this Goal and every later one, since a store may now hold many Goals. Does a root store break something concrete? If it does, tell me what, and I'll take `investigations/goal-prompt/`. I'd then want the root-store idea kept as its own issue.
2. Agree, one build Goal.
3. I can't paste run 6's prompt, since I have no files here. Go with a synthetic fixture and golden file (option A). Widen the exceptions beyond what you listed: the actor, the owner's answer, the branch, the PR, the commit, and the run's own plan may all differ. Branch and PR become flags.
4. Yes, use the ak-skill-creator eval and audit flow, and decide its parts from that flow. My only preferences are Sonnet for the test agents, 3 trials, and a modest budget.
5. Option A, no test. hyp writes the prompt, as a command that reads the log. We settle that here, with no DesignQuestion. The prompt carries no owner answers, because the runner reads OWED.
6. Use `owner-standin`.

Defaults:
- **Bad block:** `hyp add goal` stays allowed. `hyp prompt` refuses a Goal without the constraint headings, exits 1, and names what is missing.
- **Flags:** `--experiments` defaults to 5, and `--branch` and `--pr` are flags. The other defaults are fine.
- **Skill descriptions that name only one kind of Goal:** no line for them. They go under May change, as text this change makes wrong.

### Tool Bash (cwd: .)

```text
{
 "command": "cat .gitignore 2>/dev/null; grep -n \"walk\\|nearest\\|repo root\\|repository root\" docs/replay-spec.md | head -20; grep -rn \"investigations\" --include=*.go internal | grep -v _test | head; grep -n \"root\" docs/adr/0020-one-store-holds-many-goals.md | head",
 "description": "Check root-store implications in specs and code"
}
```

Result:

```text
.claude/
/hyp
*.sqlite
*.sqlite-*
*.pyc
.*-scratch
/plans/
.agentkit/
16:- Store directory in the investigated repo: `.tree/`, found by walking up from the current directory, the way git finds `.git`.
21:- Recipes: `recipes/<name>@<version>/`, in the directory that contains `.tree/` or in any parent up to the repository root (the first directory holding `.git`). See `skills/hyp-verification/RECIPE-FORMAT.md`. [sandbox-spec.md](sandbox-spec.md) §1 says when hyp runs a recipe's commands and whether a sandbox exists then.
(eval):1: no matches found: --include=*.go
```

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader scenario review",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer of a design, read-only. Repo: /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt (Go CLI `hyp`). Do not edit anything and do not run `hyp init` or any hyp write command. Read-only commands like `hyp status` inside investigations/* are fine.\n\nBackground: hyp stores an investigation as append-only events under `.tree/`; a store may hold many Goals (docs/adr/0020-one-store-holds-many-goals.md, docs/glossary.md, docs/replay-spec.md). A Goal's body may end with a \"constraints block\" with headings `May change:`, `Read only:`, `Never:`.\n\nRules settled with the owner (to be built as a new command `hyp prompt`):\nR1. `hyp prompt <G-id> --actor <runner-actor> [--branch B] [--pr N] [--experiments N]` is read-only: it appends no event and prints a `/goal` prompt for the skill hyp-goal-run (skills/hyp-goal-run/SKILL.md) derived only from the log and flags. It prints: the store's path, the actor, the Goal id, the skill name, stop clauses (runner owes nothing; or N Results recorded under the actor, default 5; or the runner cannot take the next step honestly), then the Goal body's constraints block line for line; --branch and --pr each add one clause after the block. The prompt carries no owner answer and no commit.\nR2. The block runs from the first line of the Goal body that starts with one of the three headings to the end of the body. It must hold each of the three headings at least once, and every non-empty line in it must start with one of them. A line of only spaces, a tab or a lone carriage return is not empty. A heading may repeat.\nR3. `hyp prompt` on a Goal whose body has no valid block exits 1 and names what is missing. Bad arguments (e.g. a --branch containing a space) exit 2. No repository holds the store: exit 3.\nR4. `hyp add goal` stays allowed for any body; it only warns (exit 0) when a body has a heading but no valid block. Old logs whose Goals have no block (e.g. investigations/checkpoint-store G-D1CH1R) must still replay.\nR5. No stored field: nothing new in the event schema.\nR6. Skills may not cite ADRs or docs/ files; `go test` includes internal/skillcheck, which fails when a file under skills/ names a `hyp` command that does not exist.\n\nTask: walk each rule through scenarios the designers did NOT draw, and report only concrete failures, ambiguities or contradictions, each with the exact input, what the rules say happens, and why that is wrong or undecided. Required scenarios at least: (a) a store with two Goals, `hyp prompt` for each, including a Goal-id that is an alias, ambiguous or missing; (b) a merged log from two actors, where the Goal's add event is on an unmerged branch of another worktree; (c) the store found by walking up from a subdirectory or from a nested store (investigations/*/ inside a repo that also has a root store); (d) a Goal body whose prose, before the real block, has a
[... 680 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a4ec16fc1c871895e (internal ID - do not mention to user. Use SendMessage with to: 'a4ec16fc1c871895e', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-a-1-wt/ca6abbd9-20aa-47b2-9376-6d042c58cf7b/tasks/a4ec16fc1c871895e.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"/goal\\|discovery /goal\\|uncertain design\" README.md docs/*.md skills/*/SKILL.md AGENTS.md | cut -c1-220 | head -40",
 "description": "Preflight scan for texts the change makes wrong"
}
```

Result:

```text
README.md:164:For a Goal you want run unattended, discovery or build, three skills split the work. They run in Claude Code, whose `/goal` command keeps a session going until its condition holds.
README.md:166:1. **`/hyp-owner-grill`** questions you about the purpose until it is a Goal a run can finish. It looks up the facts itself, sends questions a run could test to DesignQuestions with rivals, derives the line
README.md:167:2. **`/goal <that prompt>`** runs the Goal with the `hyp-goal-run` skill. The run does only what OWED says it owes, tests before it asks, and stops with a report you can read from `hyp status`.
README.md:168:3. **`/hyp-owner-answer`** walks you through what the stopped run left under AWAITING OWNER, explains each item and records only your answers. The next `/goal` run picks up from there.
docs/how-hyp-improves-itself.md:30:  stated by the `/goal` prompt or the recipe, not by hyp.
skills/hyp-goal-run/SKILL.md:3:description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner ca
skills/hyp-goal-run/SKILL.md:6:The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypo
skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you n
skills/hyp-goal-run/SKILL.md:68:Stop on the clauses the `/goal` prompt names. Stopping is not success.
skills/hyp-owner-answer/SKILL.md:3:description: "After a /goal run stops at the owner, walk the open readings and referred limits of the hyp tree with the owner: explain each one, answer questions, and record only the ow
skills/hyp-owner-answer/SKILL.md:10:A `/goal` run stopped and the owner now reads what waits for 
[... 1657 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sleep 240; echo done",
 "description": "Wait for reviewer",
 "timeout": 300000
}
```

Result:

```text
<tool_use_error>Blocked: sleep 240 followed by: echo done. To wait for a condition, use Monitor with an until-loop (e.g. `until <check>; do sleep 2; done`). To wait for a command you started, use run_in_background: true. Do not chain shorter sleeps to work around this block.</tool_use_error>
```

## Grill

I'm waiting on the independent reviewer before I show the routing table. My answer to your question 1 stands on what I've checked so far.

**Does a root store break something concrete?** I found no mechanical break, but it does contradict a rule.
- **No mechanical break.** hyp finds the nearest `.tree/` by walking up. Stores under `investigations/*` would still win inside their own directories. Recipes resolve from the store's directory up to the repository root, which is where `recipes/` already sits. I did not check every test that scans `investigations/*`.
- **A rule it contradicts.** AGENTS.md says trees live under `investigations/<topic>/`. A root store would contradict that until AGENTS.md changes. That change is a different purpose from #113, so it can't be a line of this Goal.

So I'll use `investigations/goal-prompt/`, as you said you would if there was a conflict. The root-store idea should be its own issue. This machine has no network, so I can't open it. I'll list it in the stop checkpoint as an out-of-Goal finding, and you can open it from there.

I have applied your other answers:
- **Actors:** `owner-standin` is the owner, and the `/goal` run's actor is chosen at hand-off.
- **No DesignQuestion for "hyp writes the prompt".** It goes in the Goal body as the purpose. The build Goal has one question, "which hyp commit?", with no Hypotheses until the run adds them.
- **Done-means exceptions:** actor, owner answer, branch, PR, commit and the run's plan may differ. `--branch` and `--pr` are flags.
- **`hyp prompt`** refuses a Goal without the headings (exit 1) and names what is missing.
- **Skill descriptions:** no line for them. They go under `May change:`.

One thing the table will ask you about: the stop clauses. The run-5 prompt in the issue has a clause that stops when the only open runner item is `add rival Hypotheses`. A build Goal starts with exactly that item, so that clause would stop it before it begins. I'll propose leaving it out.

## Owner, message 3

<task-notification>
<task-id>a4ec16fc1c871895e</task-id>
<tool-use-id>toolu_018Cq1RnFDvqDVzaqgq5GU5T</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-a-1-wt/ca6abbd9-20aa-47b2-9376-6d042c58cf7b/tasks/a4ec16fc1c871895e.output</output-file>
<status>completed</status>
<summary>Agent "Second-reader scenario review" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>**Review of the `hyp prompt` rules: concrete failures, ambiguities and contradictions, most severe first**

Facts below come from `internal/cli/cli.go`, `resolve.go`, `read.go`, `write.go`, `idcite.go`, `internal/store/store.go`, `internal/herr/herr.go` and `internal/replay/replay.go`. I did not run anything.

**Wiring and exit codes**
- A command is registered in the `commands` map in `cli.go`. `Commands()` lists it automatically, so skillcheck sees it.
- `usage` in `cli.go` is a separate constant, and `help_test` probably checks it.
- Read commands use `openReader()` or `openEvents()`. `openEvents` calls `store.FindRoot(cwd)`, then `store.Events`.
- `openReader` also loads recipe hashes and runs `contain.Mark`, which `hyp prompt` does not need.
- `herr.Refused` gives exit 1, `herr.BadArgs` gives 2, and `herr.Env` or any error of no kind gives 3. `ErrNoStore` is a plain error, so it already maps to 3.
- `usageErr` is the helper in the cli package for exit 2.

**1. Unknown, ambiguous and non-Goal ids are unspecified (a).**
- `resolveRef` returns `herr.BadArgs` for an unknown or ambiguous ref, so exit 2.
- R3 gives exit 1 only for a bad block.
- `hyp status` also exits 2 when the node is not a Goal ("takes a Goal").
- The rules must say that a missing, ambiguous or non-Goal id exits 2, and that an alias or a case-insensitive display id (`g-abc123`) is accepted. They must also say which form the prompt prints. An alias resolves only in this store, so printing the canonical display id is the safe choice.

**2. The prompt is not "derived only from the log and flags" (c).**
- The store path comes from `os.Getwd()` plus `FindRoot`, which picks the nearest ancestor store.
- From a subdirectory, or from a nested `investigations/x/` store, the printed path differs.
- A runner launched elsewhere may resolve a different store.
- `FindRoot` uses `filepath.Abs`, not `EvalSymlinks`, so `/var` and `/private/var` differ on macOS. A worktree path may not exist later.
- Decide: print the absolute root with symlinks resolved, and say that this is the one cwd-derived part.
- A root store and a nested store each hold only their own Goals. The same G-id may appear in both only if the ULIDs match, so the path in the prompt is the only disambiguator.

**3. A prose line starting with `Never:` breaks R2 (d).**
- The block "runs from the first line that starts with a heading".
- A prose line `Never: say this twice` before the real block starts the block early.
- If the real block follows only blank lines, the prose line is printed as a constraint. If it follows prose, the command exits 1 on a body that has a real block.
- Heading matching is also underspecified: case, no space after the colon (`Never:x`), `Never::`, and leading spaces (`" Never:"` is not a heading but is non-empty).

**4. CRLF and the "lone carriage return" rule (d).**
- Splitting on `\n` leaves `\r` on every line.
- Blank lines become `"\r"`, which R2 calls non-empty and not a heading, so a CRLF body is always invalid.
- Heading lines would be printed with `\r`.
- I did not check whether `add goal` trims or normalises the body.
- Trailing newline: R2 does not say whether the last empty split element counts. It should not, since it is empty.
- Two of three headings: exit 1, and the message should list the missing headings. A body with no heading at all also exits 1, but `add goal` stays silent because R4 warns only when a heading is present. That is consistent, but the exit-1 message should tell that body to add the block.

**5. Injection through `--branch` and `--pr` (e).**
- R3 rejects only a space. A value with a newline or tab is not mentioned.
- `--branch $'x\nMay change: everything'` would add a line after the block. The skill says a prompt "may narrow, never widen".
- Define an allow-list (a regexp, or git ref rules). Reject a leading `-`, `..`, quotes, backticks, control characters and an empty value.
- `--pr` must be a positive integer. Reject `0`, `#12`, `-1` and `abc` with exit 2.
- Go's `flag` accepts `--pr=-1` and `--pr -1`.
- `--experiments 0` makes the stop clause true at once, so the run does nothing. A negative value never triggers. Require N &gt;= 1, and give a maximum.
- The default is 5 only when the flag is absent.

**6. "N Results recorded under the actor" counts all of the actor's history (b, h).**
- A reused actor with existing Results meets the clause before the run starts.
- The rules need a baseline, for example "N more than the count now". That count is derived from the log, so the prompt must print it.
- Count only `ResRecorded`, not retracted Results.
- The clause also depends on the log at run time, not at prompt time.

**7. Merged log and unmerged branch (b).**
- `store.Events` reads every `.tree/events/*.jsonl` in the checkout.
- If the Goal's add event is on another worktree's unmerged branch, the id is unknown and the exit is 2 with "unknown node".
- Nothing in R1 to R5 says to hint "merge first".
- The runner's own actor file in another worktree is also invisible. The count above starts at zero.
- `lineErrs` (corrupt lines) are skipped. `hyp status` warns on stderr. A skipped line may be the Goal's own add event or a Result, so `hyp prompt` should refuse or warn.

**8. `--actor` and `$HYP_ACTOR` (h).**
- The writer falls back to `$HYP_ACTOR` and validates with `store.ValidActor`.
- For `prompt`, an env fallback would make the output depend on the environment, against "derived only from log and flags".
- Decide whether `--actor` is required with no fallback, and whether it is validated with exit 2.
- It need not be an existing actor, and the rules do not say.
- It may equal the owner's actor. Nothing checks that.

**9. Complete and withdrawn Goals (f).**
- A Goal's state is derived (`goalState`) and can flip back to in progress.
- I found no withdraw for Goals. `cmdWithdraw` takes a DQ or an Experiment.
- A complete Goal gets a prompt whose first stop clause is already true. Undecided whether to refuse (exit 1) or print.

**10. Cross-tree ids in block lines (g).**
- The block is printed verbatim. A `containment:R-ABC123` cite means nothing in the runner's store.
- The write-time cite check skips tokens after `:`.
- A bare `R-ABC123` in the block is resolved nowhere. The prompt does not say which store it means.
- Say that block text is printed unresolved.

**11. Block size and R4.**
- A Goal body is capped at 2000 runes (`MaxGoalBody`). The three headings share that cap with the met rule.
- The `add goal` warning must go through `admitted()` in `write.go` (CLI-side), not replay validation. Otherwise old logs such as `G-D1CH1R` would change on replay.

**12. R6.**
- skillcheck only checks that a `hyp &lt;word&gt;` in a backtick or fence names a registered command.
- A SKILL.md edit that mentions `hyp prompt` before the command exists fails the test. Order the commits accordingly.</result>
<usage><subagent_tokens>42528</subagent_tokens><tool_uses>7</tool_uses><duration_ms>65440</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"MaxGoalBody\" internal | head -5; grep -n \"BadArgs\" internal/cli/resolve.go | head; grep -n \"func FindRoot\" -A12 internal/store/store.go | head -25; grep -n \"takes a Goal\\|not a Goal\" internal/cli/*.go | head -5; grep -n \"func ValidActor\" -A6 internal/store/*.go | head -10",
 "description": "Verify reviewer's claims against source"
}
```

Result:

```text
internal/cli/write.go:255:		maxBody = event.MaxGoalBody
internal/replay/validate.go:85:		maxBody = event.MaxGoalBody // a Goal's body holds the owner's met rule
internal/event/event.go:117:	MaxGoalBody     = 2000
33:			return "", herr.BadArgs("ambiguous node %s", ref)
36:	return "", herr.BadArgs("unknown node %s", ref)
47:func FindRoot(start string) (string, error) {
48-	dir, err := filepath.Abs(start)
49-	if err != nil {
50-		return "", err
51-	}
52-	for {
53-		if fi, err := os.Stat(eventsDir(dir)); err == nil && fi.IsDir() {
54-			return dir, nil
55-		}
56-		parent := filepath.Dir(dir)
57-		if parent == dir {
58-			return "", ErrNoStore
59-		}
internal/cli/read.go:147:			return usageErr("hyp status takes a Goal: %s is a %s", pos[0], p.Nodes[id].Kind)
internal/store/store.go:223:func ValidActor(actor string) bool { return actorRe.MatchString(actor) }
```

## Grill

The second reader is done. I checked its main facts against the source before using them: a Goal body is capped at 2000 runes, an unknown or ambiguous id exits 2, and `FindRoot` returns a plain absolute path. I have folded its findings into the lines below. Nothing is written yet.

## What the second reader changed

- **Unknown ids exit 2, not 1.** A missing, ambiguous or non-Goal id exits 2, and an alias or lower-case id resolves. The prompt prints the Goal's display id.
- **`--actor` is required.** `$HYP_ACTOR` is ignored, so the output depends only on the log and the flags.
- **Flag values can't inject text.** `--branch` and `--pr` are checked, because a newline in `--branch` could add a line after the block and widen it. `--experiments` must be at least 1.
- **CRLF bodies are invalid.** A `\r` is a lone carriage return, so a CRLF body has no valid block. This matches the skill's rule.
- **The warning lives only at write time.** It goes in `hyp add goal`, not in replay, so old logs such as G-D1CH1R's still replay.
- **The body shares a 2000-rune cap with the met rule.** The block must fit in what is left.

## Routing table

| # | Question | Place | Scenario / invariant |
|---|---|---|---|
| 1 | Where the Goal lives | looked up, then your answer: new store `investigations/goal-prompt/` | none |
| 2 | One Goal for the command and the skill text | Goal | none |
| 3 | Purpose: hyp writes the prompt, for any Goal, and the prompt carries no owner answer, commit or plan | Goal title and body, settled here with no DesignQuestion | none |
| 4 | Which hyp commit builds it | DesignQuestion `q-commit`, no Hypotheses yet. The run adds the candidate commits, as in `checkpoint-store`. | none |
| 5 | Done clause 1, with run 6's text unavailable | line 1, using a fixture and a golden file | `[P]` exact output |
| 6 | Flags and their values | lines 2 and 4 | `[P]` flag clauses; `[N]` bad arguments exit 2 |
| 7 | A Goal with no valid block | line 3 | `[N]` exit 1, naming what is missing |
| 8 | `hyp prompt` writes nothing | line 5 | `[N]` the `.tree/` bytes are unchanged |
| 9 | Many Goals in one store | line 6 | `[P]` isolation |
| 10 | `hyp add goal` stays allowed and warns | line 7 | `[P]` warning agrees with the refusal |
| 11 | A runner obeys the block | line 8, by ak-skill-creator eval | `[P]` |
| 12 | The grill hands off the output verbatim | line 9, by ak-skill-creator eval | `[P]` |
| 13 | Descriptions of hyp-goal-run and hyp-prototype | `May change:`, no line | none |
| 14 | A root store for later Goals | an out-of-Goal finding, named in the stop checkpoint | none |
| 15 | The run's constraints | the Goal body's block | none |

Lines 3 and 4 refuse command input, not a state in the log, so the "rejected on replay" half of the `[N]` form does not apply to them. Line 7 is `[P]` because the owner decided that `hyp add goal` stays allowed.

The golden file cannot come from run 6, so it will be written from the run-5 prompt in the issue. So the done test is met against a fixture, not against run 6's text. That is the cost of option A.

## Four defaults to confirm

- **Stop clauses.** The prompt has: the runner owes nothing; N Results recorded under the actor after it starts (default 5); it cannot take the next step honestly. It also names the stop checkpoint and the lines with the Results that ran them. The run-5 clause "every runner item is `add rival Hypotheses`" is left out, because a build Goal starts with exactly that item. Run 6 stopped on its select, which "runner owes nothing" already covers.
- **A reused actor** is not refused. The clause says "after you start", and the grill skill already says to choose a new actor.
- **A complete Goal** still gets a prompt, and its first stop clause is already true.
- **Block text** is printed as written. Ids like `<tree>:<id>` are not resolved.

## Writes, in order

I will run the first and third with `HYP_ACTOR=claude-grill-113`. I will not run the second until you approve it.

1. `cd investigations/goal-prompt && hyp init` (runner)
2. The Goal, which you approve. I wrote the body to fit the 2000-rune cap, and each line is under 600 runes:

```text
hyp add goal --actor owner-standin --why "Owner, 2026-11-12: #113, hyp prints the /goal prompt for any Goal" \
  --title "hyp prints the /goal prompt for any Goal from the tree" \
  --body $'Purpose (#113, core): hyp prints the /goal prompt for hyp-goal-run for any Goal, discovery or build, with a command that reads the log: hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]. Settled in the grill: hyp writes the prompt, and no run tests a skill-written one.\nA run\'s constraints are the Goal body\'s May change:, Read only: and Never: lines. The prompt copies them line for line; a /goal prompt may narrow them, never widen them. The prompt carries no owner answer, no commit and no plan, because OWED lists what the runner owes.\nThe candidate is a hyp commit with the skill text it makes wrong, which includes the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal. Lines 1 to 7 are decided by go test on a synthetic fixture store and a golden prompt written from the run-5 prompt in #113; lines 8 and 9 by ak-skill-creator\'s eval flow with Sonnet test agents, 3 trials and a modest budget.\nMay change: this repository by commits on one local branch, and a PR the run opens and never merges; the texts the change makes wrong, such as skills/, docs/, AGENTS.md and README.md; tests and testdata under internal/; a new recipe version under recipes/ for the eval; this Goal\'s own tree under investigations/goal-prompt/\nRead only: every other tree under investigations/; the installed ak-skill-creator skill and its eval_skill.py\nNever: replace the installed hyp binary\nNever: merge a PR\nNever: add a typed constraints field to hyp add goal\nNever: edit or delete a recorded event' \
  --line "[P] On the fixture store, hyp prompt <G> --actor A for a Goal whose body ends with a valid block exits 0 and prints exactly the golden prompt with the store's absolute path, A and the Goal's display id substituted: skill hyp-goal-run, the stop clauses, then the block's lines verbatim and in order. A Goal named by alias or in lower case prints its display id. The hyp usage lists the command." \
  --line "[P] --branch B adds one branch clause and --pr N one PR clause after the block, branch first; with both, two clauses and no other change; --experiments N sets the number in the Results stop clause (5 when absent); with no flag the output has neither clause." \
  --line "[N] A Goal body with no valid block makes hyp prompt exit 1 with nothing on stdout and a message that names the missing heading or the first invalid line. Invalid: no heading, one or two of the three headings, a non-empty line in the block that starts with no heading, a line of spaces, a tab or a lone carriage return, CRLF line ends." \
  --line "[N] Bad arguments make hyp prompt exit 2 with nothing on stdout: no --actor (\$HYP_ACTOR is ignored) or an invalid actor name; a missing, ambiguous or non-Goal id; a --branch that is empty, starts with -, holds .., a space or a control character; a --pr that is not a positive integer; --experiments below 1." \
  --line "[N] hyp prompt appends no event and changes no file: the bytes of .tree/ are equal before and after, whether it exits 0, 1 or 2." \
  --line "[P] In a store with two Goals, the prompt for one holds only that Goal's id and block, and equals the prompt from a store that holds only that Goal's events, except for the store path." \
  --line "[P] hyp add goal appends any body. It exits 0 with a warning naming the missing part when the body has a heading and hyp prompt would exit 1 on it; a body with no heading gets no warning. hyp check of a log with such a Goal prints no error." \
  --line "[P] A runner given the hyp-goal-run skill and the prompt hyp prompt prints for a build Goal fixture whose block forbids an act, such as merging a PR or changing a read-only directory, does not take it, and when a step needs it, refers it with a reason or records a reading and stops, in each eval case." \
  --line "[P] A grill given the hyp-owner-grill skill, in a repository whose hyp has hyp prompt, ends by giving the owner the command's output exactly and adds no clause, in each eval case."
```

3. The runner writes:
   - `hyp add dq --parent <G-id> --title "Which hyp commit makes hyp print the /goal prompt for any Goal?" --alias q-commit`
   - `hyp link <G-id> external_ref gh:113`
   - `hyp check`
   - `hyp checkpoint --goal <G-id> --summary "…" --why "end of grill"`

## Hand-off caveat

`hyp prompt` can't print this Goal's prompt, because this run builds it. At hand-off I will write the first run's `/goal` prompt by hand, in the shape the new command will print. I'll copy the constraints block line for line. The installed grill skill expects `hyp prompt` and will not work for later Goals until this Goal's commit is installed.

Please confirm:
- the table, or move rows;
- the four defaults;
- the exact `hyp add goal` command above, to be run as `owner-standin`.

## Owner, message 4

Confirmed. The table is fine as it stands, and I accept the four defaults. Row 1 is fine as long as row 14 keeps the root store as its own issue. Go ahead and run the `hyp add goal` command as `owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/investigations/goal-prompt && cd /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/investigations/goal-prompt && HYP_ACTOR=claude-grill-113 hyp init && ls -a",
 "description": "Create the new store for the Goal"
}
```

Result:

```text
initialised .tree in /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/investigations/goal-prompt
.
..
.tree
```

## Grill

Now the owner write, exactly as approved.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp add goal --actor owner-standin --why \"Owner, 2026-11-12: #113, hyp prints the /goal prompt for any Goal\" \\\n  --title \"hyp prints the /goal prompt for any Goal from the tree\" \\\n  --body $'Purpose (#113, core): hyp prints the /goal prompt for hyp-goal-run for any Goal, discovery or build, with a command that reads the log: hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]. Settled in the grill: hyp writes the prompt, and no run tests a skill-written one.\\nA run\\'s constraints are the Goal body\\'s May change:, Read only: and Never: lines. The prompt copies them line for line; a /goal prompt may narrow them, never widen them. The prompt carries no owner answer, no commit and no plan, because OWED lists what the runner owes.\\nThe candidate is a hyp commit with the skill text it makes wrong, which includes the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal. Lines 1 to 7 are decided by go test on a synthetic fixture store and a golden prompt written from the run-5 prompt in #113; lines 8 and 9 by ak-skill-creator\\'s eval flow with Sonnet test agents, 3 trials and a modest budget.\\nMay change: this repository by commits on one local branch, and a PR the run opens and never merges; the texts the change makes wrong, such as skills/, docs/, AGENTS.md and README.md; tests and testdata under internal/; a new recipe version under recipes/ for the eval; this Goal\\'s own tree under investigations/goal-prompt/\\nRead only: every other tree under investigations/; the installed ak-skill-creator skill and its eval_skill.py\\nNever: replace the installed hyp binary\\nNever: merge a PR\\nNever: add a typed constraints field to hyp add goal\\nNever: edit or delete a recorded event' \\\n  --line \"[P] On the fixture store, hyp prompt <G> --actor A for a Goal whose body ends with a valid block exits 0 and prints exactly the golden prompt with the store's absolute path, A and the Goal's display id substituted: skill hyp-goal-run, the stop clauses, then the block's lines verbatim and in order. A Goal named by alias or in lower case prints its display id. The hyp usage lists the command.\" \\\n  --line \"[P] --branch B adds one branch clause and --pr N one PR clause after the block, branch first; with both, two clauses and no other change; --experiments N sets the number in the Results stop clause (5 when absent); with no flag the output has neither clause.\" \\\n  --line \"[N] A Goal body with no valid block makes hyp prompt exit 1 with nothing on stdout and a message that names the missing heading or the first invalid line. Invalid: no heading, one or two of the three headings, a non-empty line in the block that starts with no heading, a line of spaces, a tab or a lone carriage return, CRLF line ends.\" \\\n  --line \"[N] Bad arguments make hyp prompt exit 2 with nothing on stdout: no --actor (\\$HYP_ACTOR is ignored) or an invalid actor name; a missing, ambiguous or non-Goal id; a --branch tha
[... 1338 more characters]
```

Result:

```text
G-KYCW82 node.add
  line G-KYCW82#1: [P] On the fixture store, hyp prompt <G> --actor A for a Goal whose body ends with a valid block exits 0 and prints exactly the golden prompt with the store's absolute path, A and the Goal's display id substituted: skill hyp-goal-run, the stop clauses, then the block's lines verbatim and in order. A Goal named by alias or in lower case prints its display id. The hyp usage lists the command.
  line G-KYCW82#2: [P] --branch B adds one branch clause and --pr N one PR clause after the block, branch first; with both, two clauses and no other change; --experiments N sets the number in the Results stop clause (5 when absent); with no flag the output has neither clause.
  line G-KYCW82#3: [N] A Goal body with no valid block makes hyp prompt exit 1 with nothing on stdout and a message that names the missing heading or the first invalid line. Invalid: no heading, one or two of the three headings, a non-empty line in the block that starts with no heading, a line of spaces, a tab or a lone carriage return, CRLF line ends.
  line G-KYCW82#4: [N] Bad arguments make hyp prompt exit 2 with nothing on stdout: no --actor ($HYP_ACTOR is ignored) or an invalid actor name; a missing, ambiguous or non-Goal id; a --branch that is empty, starts with -, holds .., a space or a control character; a --pr that is not a positive integer; --experiments below 1.
  line G-KYCW82#5: [N] hyp prompt appends no event and changes no file: the bytes of .tree/ are equal before and after, whether it exits 0, 1 or 2.
  line G-KYCW82#6: [P] In a store with two Goals, the prompt for one holds only that Goal's id and block, and equals the prompt from a store that holds only that Goal's events, except for the store path.
  line G-KYCW82#7: [P] hyp add goal appends any body. It exits 0 with a warning naming the missing part when the body has a heading and hyp prompt would exit 1 on it; a body with no heading gets no warning. hyp check of a log with such a Goal prints no error
[... 529 more characters]
```

## Grill

Goal is `G-KYCW82`. Now the runner writes.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && hyp add dq --parent G-KYCW82 --alias q-commit --title \"Which hyp commit makes hyp print the /goal prompt for any Goal?\" --body \"The build's one question; its candidate is a hyp commit, added by the run when one exists.\" --why \"The build's one question, as in checkpoint-store; no Hypothesis yet.\" && hyp link G-KYCW82 external_ref gh:113 --why \"The issue this Goal serves.\" && hyp check && hyp status G-KYCW82 | head -30",
 "description": "Add the DesignQuestion, link the issue, check"
}
```

Result:

```text
Q-3NHTH3 node.add
G-KYCW82 edge.add
check: 0 errors, 0 warnings
GOAL G-KYCW82 hyp prints the /goal prompt for any Goal from the tree — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-KYCW82#1 no candidate — [P] On the fixture store, hyp prompt <G> --actor A for a Goal whose body ends with a valid block exits 0 and prints exactly the golde…
  G-KYCW82#2 no candidate — [P] --branch B adds one branch clause and --pr N one PR clause after the block, branch first; with both, two clauses and no other cha…
  G-KYCW82#3 no candidate — [N] A Goal body with no valid block makes hyp prompt exit 1 with nothing on stdout and a message that names the missing heading or th…
  G-KYCW82#4 no candidate — [N] Bad arguments make hyp prompt exit 2 with nothing on stdout: no --actor ($HYP_ACTOR is ignored) or an invalid actor name; a missi…
  G-KYCW82#5 no candidate — [N] hyp prompt appends no event and changes no file: the bytes of .tree/ are equal before and after, whether it exits 0, 1 or 2.
  G-KYCW82#6 no candidate — [P] In a store with two Goals, the prompt for one holds only that Goal's id and block, and equals the prompt from a store that holds …
  G-KYCW82#7 no candidate — [P] hyp add goal appends any body. It exits 0 with a warning naming the missing part when the body has a heading and hyp prompt would…
  G-KYCW82#8 no candidate — [P] A runner given the hyp-goal-run skill and the prompt hyp prompt prints for a build Goal fixture whose block forbids an act, such …
  G-KYCW82#9 no candidate — [P] A grill given the hyp-owner-grill skill, in a repository whose hyp has hyp prompt, ends by giving the owner the command's output …
FRONTIER
  Q-3NHTH3 open/no_candidates Which hyp commit makes hyp print the /goal prompt for any G…
STALE
OWED
  runner: Q-3NHTH3 add rival Hypotheses
AWAITING OWNER
RECENT
  2026-11-12 14:41 claude-grill-113 edge.add G-KYCW82 — The issue this Goal serves.
CHECKPOINT none
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "HYP_ACTOR=claude-grill-113 hyp checkpoint --goal G-KYCW82 --summary \"Grill of #113 done. G-KYCW82 is a build Goal with 9 lines and one question, Q-3NHTH3, with no candidate yet. Out-of-Goal finding for the owner to open as its own issue: one shared store at the repository root for this and every later Goal, which AGENTS.md's investigations/<topic>/ rule does not allow today. hyp prompt does not exist yet, so the first run's /goal prompt is written by hand.\" --why \"end of grill\" && hyp check && cd ../.. && git status --short",
 "description": "Record end-of-grill checkpoint and check"
}
```

Result:

```text
G-KYCW82 checkpoint
check: 0 errors, 0 warnings
?? investigations/goal-prompt/
```

## Grill

The grill is done. The Goal is in a new store at `investigations/goal-prompt/`, and `hyp check` prints `check: 0 errors, 0 warnings`.

## What was written

| Write | Result |
|---|---|
| `hyp init` (runner `claude-grill-113`) | store created |
| `hyp add goal` (as `owner-standin`, the command you approved) | **G-KYCW82**, with 9 lines |
| `hyp add dq` (runner) | **Q-3NHTH3** (`q-commit`), no Hypotheses yet. OWED says `add rival Hypotheses`, which is the run's first act. |
| `hyp link` | external_ref `gh:113` |
| `hyp checkpoint` | end of grill. Its summary also names the root-store finding. |

## Not done

- **Nothing is committed.** `investigations/goal-prompt/` is untracked. Say the word and I'll commit it with a `docs:` message. I only commit when you ask.
- **The root-store issue is not opened.** This machine has no network. It is recorded as an out-of-Goal finding in the checkpoint. The issue text would be: *"One shared store at the repository root for this and every later Goal. AGENTS.md says trees live under `investigations/<topic>/`, so the rule and the checklist in hyp-owner-grill would change with it."*

## Hand-off

`hyp prompt` doesn't exist yet, because this Goal builds it. So I wrote this prompt by hand in the shape the command will print, and copied the constraints block line for line. The run's branch name is my suggestion. Change it if you like. Use a new runner actor.

```text
/goal Run the hyp tree in /Users/datguy/.hyp-sandbox/g146-pre-a-1/wt/investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-prompt-1. Its Goal is G-KYCW82.
Done when the latest printed `hyp status G-KYCW82` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under claude-prompt-1 after you start, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
May change: this repository by commits on one local branch, and a PR the run opens and never merges; the texts the change makes wrong, such as skills/, docs/, AGENTS.md and README.md; tests and testdata under internal/; a new recipe version under recipes/ for the eval; this Goal's own tree under investigations/goal-prompt/
Read only: every other tree under investigations/; the installed ak-skill-creator skill and its eval_skill.py
Never: replace the installed hyp binary
Never: merge a PR
Never: add a typed constraints field to hyp add goal
Never: edit or delete a recorded event
Work on the branch feat/hyp-prompt.
```

The prompt names the Goal by its id, so its add must be in the log the run starts from. That means committing `investigations/goal-prompt/` first, or running on a branch that holds it.

For the lines the run owes the owner, the stop checkpoint will list the Results that ran each one. Lines 8 and 9 depend on the ak-skill-creator eval at Sonnet, 3 trials and a modest budget, as you set.
