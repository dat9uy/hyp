# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-a-1/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-1/tmp/issue-113.md"
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
 "command": "(which hyp; hyp status 2>&1 | head -40; ls; ls .tree 2>&1 | head)",
 "description": "Show hyp status and repo layout"
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-base-a-1/bin/hyp
no .tree store found; run 'hyp init'
AGENTS.md
CLAUDE.md
cmd
docs
go.mod
go.sum
internal
investigations
README.md
recipes
skills
ls: .tree: No such file or directory
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls investigations investigations/* | head -40; hyp 2>&1 | head -40; hyp prompt --help 2>&1 | head -20; git log --oneline | head -15; ls internal; grep -rn \"prompt\" cmd/hyp/*.go | head",
 "description": "Survey investigations, CLI and prompt command"
}
```

Result:

```text
investigations:
checkpoint-store
checkpoint-store-design
id-citation
recipe-base

investigations/checkpoint-store:

investigations/checkpoint-store-design:

investigations/id-citation:
arms

investigations/recipe-base:
arms.txt
usage: hyp <command> [flags]

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
Statuses are derived from events; no command can set one.

setup:
  init                                   create .tree in the current directory
  version                                print the hyp version

write:
  add goal --title T [--body B] [--alias A] [--line L]... [--cite N=KEY]...
  add dq   --parent REF --title T [--body B] [--alias A]
  add hyp  --parent REF --title T [--body B] [--alias A] [--subject PATH]...
  add exp  --parent REF --title T --recipe NAME@V|none [--rubric NAME@V|reading] [--records PATH]... [--target REF|none]... [--alias A]
  record   --exp REF [--verdict REF=supports|contradicts|inconclusive]... --observation O [--evidence P]... [--limit L]... [--closes R#N]...
  retract REF | abandon REF | reaffirm REF
  supersede NEW --old OLD
  amend HYP --body B                     revise a Hypothesis body
  subject HYP [PATH...] [--base C]       tie a Hypothesis to the base and files that embody its claim
  link FROM depends_on|external_ref TO   unlink FROM TYPE TO
  select DQ [HYP]                        admitted only when hyp status owes it (no HYP: print the suggestion only)
  unselect DQ | defer DQ | withdraw DQ|EXP | resume DQ
  checkpoint [--summary S] [--goal G] [--refers K[=TEXT]]...   the runner's report: a summary, refers, or both
  answer [--answer R#rN=TEXT]... [--accept K=TEXT]... [--returns K=TEXT]...   the owner's act; no --why

read (REF/TIME: an event id, a node ref, or an RFC3339 time):
  status [--as-of REF|TIME] [--max-lines N] [--live]   orientation screen, 60 lines by default; --live hides contradicted rivals
  show REF|R-id#n|G-id#n [--as-of REF|TIME]  one node, or one limit or line, in full, with its
[... 3475 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "wc -l docs/glossary.md docs/principles.md skills/*/SKILL.md && sed -n 1,200p docs/glossary.md",
 "description": "Read glossary"
}
```

Result:

```text
<persisted-output>
Output too large (33.3KB). Full output saved to: /Users/datguy/.claude/projects/-Users-datguy--hyp-sandbox-g146-base-a-1-wt/560fedd4-84a6-43c0-b113-e41924c3198b/tool-results/bvxwfg01h.txt

Preview (first 2KB):
     439 docs/glossary.md
     140 docs/principles.md
      83 skills/hyp-goal-run/SKILL.md
     141 skills/hyp-owner-answer/SKILL.md
     123 skills/hyp-owner-grill/SKILL.md
     217 skills/hyp-prototype/SKILL.md
     108 skills/hyp-verification/SKILL.md
    1251 total
# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.

## Language

### Trees

**Tree**:
One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
_Avoid_: graph, log (the log is its events), project

**Store**:
The `.tree/` directory of a repository: one events file per actor and the evidence files Results point at ([spec §1](replay-spec.md#1-names-and-locations)). hyp finds it by walking up from the current directory, as git finds `.git`. In code: package `store`.
_Avoid_: database, index (hyp keeps none, ADR 0006)

**Event**:
One line of `.tree/events/<actor>.jsonl`: an id, a time, an actor, a type from a closed set and its payload ([spec §3](replay-spec.md#3-event-envelope), [spec §4](replay-spec.md#4-event-types-closed-set)). Events are only appended, and every status is derived from them. In code: `event.Event`.
_Avoid_:
[... 107 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp --help 2>&1 | sed -n 40,80p; grep -rln '\"prompt\"\\|hyp prompt' cmd internal skills docs | head; ls -a investigations/checkpoint-store investigations/checkpoint-store-design; (cd investigations/checkpoint-store && hyp status 2>&1 | head -70)",
 "description": "Check for prompt command and sample build tree"
}
```

Result:

```text
  bundle run --record DIR REF NAME [ARGS...]   run one Standalone command with no sandbox; exits with its exit code
  bundle evidence --record DIR REF OUT_DIR     run recipe REF's Evidence commands on remove's OUT_DIR; exits with the failing one's code
  bundle known-answers REF                     run every known-answers/NAME/ case of bundle REF, twice each; 1 on a wrong answer

sandbox (a disposable clone under ~/.hyp-sandbox/ID for one Experiment):
  sandbox create ID --host H --base REV --recipe NAME@VERSION|none [--recipe NAME@VERSION]... [--recipes DIR] [--store SRC --cut EVENT] [--inputs FILE] [--hyp BIN]
  sandbox drive ID                             run every recipe's Drive commands in the sandbox; exits with the failing one's code
  sandbox remove ID OUT_DIR [--keep REL]... | sandbox remove ID --discard
investigations/checkpoint-store:
.
..
.tree

investigations/checkpoint-store-design:
.
..
.tree
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A sto…
  G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a …
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and…
  G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Goal's lines come with …
  G-D1CH1R#5 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] One answer may settle key…
  G-D1CH1R#6 covered on Q-PC04ED; ran in R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M
[... 2075 more characters]
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-1/wt/skills/hyp-goal-run/SKILL.md"
}
```

Result:

```text
1	---
2	name: hyp-goal-run
3	description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report the owner can read from hyp status. Use when a /goal prompt names this skill."
4	---
5	
6	The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.
7	
8	## Start
9	
10	- Read the Goal first with `hyp show <goal-id>`, including its checkpoints, because `hyp status` does not print its met rule.
11	- Everything you need is in the tree. Do not read session logs, reports or memory about earlier runs.
12	- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a step needs something they forbid, refer it with a reason or record a reading, and stop.
13	- A store may hold many Goals. Then `hyp status` prints only a GOALS table and one Awaiting owner list, so run `hyp status <goal-id>` wherever this skill says `hyp status`: OWED, LINES, the stop line and the last checkpoint are per Goal. Name the Goal with `--goal <goal-id>` on `hyp checkpoint`.
14	- Print `hyp status` at the end of every turn.
15	
16	## Work
17	
18	Act on what OWED in `hyp status` says the runner owes, and on nothing else. Plan only the next Experiment.
19	
20	- Use the Goal's must-fail cases as known-answer flaw cases, in the owner's terms.
21	- The recipe's checks may exist before the verdict rule, because `hyp add exp` pins them. Write and run no candidate until the rule is rec
[... 9959 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1 && ls; ls tmp; find . -name '*goal-prompt*' 2>/dev/null | head; cd wt && grep -rn \"goal-prompt\\|hyp prompt\\|/goal\" --include=*.md --include=*.go . | grep -v \"^./skills/hyp-goal-run\" | head -40",
 "description": "Locate run-6 prompt and prompt references"
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
(eval):1: no matches found: --include=*.md
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -60; hyp render --markdown 2>&1 | grep -n \"^#\" | head -30",
 "description": "Show the build Goal as a model"
}
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
[... 7718 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt && grep -rniE \"goal-prompt|hyp prompt|May change:\" --exclude-dir=.git . | cut -c1-220 | head -40",
 "description": "Search repo for prompt references"
}
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt && cat docs/adr/0021-a-skill-change-is-a-build-goal.md; grep -n -iE \"discovery Goal|build Goal|Goal\\*\\*|Line\\*\\*|constraint\" docs/glossary.md | cut -c1-260",
 "description": "Read ADR 0021 and glossary terms"
}
```

Result:

```text
---
status: accepted
---

Supersedes in part [ADR 0007](0007-records-start-at-230.md) (the skill-text gate's replay instrument), [ADR 0010](0010-a-record-keeps-or-identifies-what-a-replay-needs.md) (the follow-up that rebuilds `skill-text-arms` Arms from patches) and [ADR 0015](0015-the-select-is-the-gate.md) (the frozen verifier recipes).

# A skill change is a build Goal

AGENTS.md gated skill text apart from code. A skill-text change shipped only when a session replay from the event's store cut showed the failure in 3 or more of 5 trials without the change and in 1 or fewer of 5 with it, and nothing shipped when no cut reproduced the failure. Its one instrument, `recipes/skill-text-arms@5` with the rubric `rubrics/skill-text-gate@1`, replays the id-citation tree's e-arms with `hyp-prototype` and `hyp-verification` loaded. It cannot reach a skill that runs before any tree exists, such as `hyp-owner-grill`, or one that spans a whole `/goal` run, such as `hyp-goal-run` and `hyp-owner-answer`. For those skills the rule was a permanent freeze, and a new skill needed its own exception. [how-hyp-improves-itself.md](../how-hyp-improves-itself.md) had already found that replay studies of skill text decided nothing, because agents with the skill loaded rarely made the mistakes they counted.

Owner decision on 2026-11-11 (#134): building is a Goal like any other (ADR 0020), and a change to `skills/` is building. The special gate was context that led agents to design gate policy instead of applying that model.

## Decision

- **A change to `skills/`, a new skill included, is a build Goal like any code change.** Its candidate is the commit; its lines state the skill's behaviour; each line is checked by its own test or by an instrument the Goal builds, red on base, and the select ends it. For a new skill, base is no skill.
- **How a skill is tested is that Goal's choice**, settled in its grill. One choice is ak-skill-creator's eval: cases in its `evals.json` format kept as a f
[... 2322 more characters]
```

Result:

```text
./.claude/skills/hyp-owner-grill/SKILL.md:34:| what a run may change, may only read, or must never do | the Goal body's `May change:`, `Read only:` and `Never:` headings | yes |
./.claude/skills/hyp-owner-grill/SKILL.md:96:The Goal body ends with the run's constraints block, so a run and `hyp prompt` can find it:
./.claude/skills/hyp-owner-grill/SKILL.md:99:May change: <the repository, how it is changed (for example commits on one local branch, a PR never merged), and the texts the change makes wrong>
./.claude/skills/hyp-owner-grill/SKILL.md:104:The block runs from the first line that starts with `May change:`, `Read only:` or `Never:` to the end of the body. It holds each of the three headings at least once, and eve
./.claude/skills/hyp-owner-grill/SKILL.md:113:hyp prompt <G-id> --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
./.claude/skills/hyp-owner-grill/SKILL.md:116:It prints the store's path, the actor, the Goal id, the skill, the stop clauses (the runner owes nothing, or N Results under the actor, 5 by default) and the Goal body's cons
./.claude/skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (
./skills/hyp-owner-grill/SKILL.md:34:| what a run may change, may only read, or must never do | the Goal body's `May change:`, `Read only:` and `Never:` headings | yes |
./skills/hyp-owner-grill/SKILL.md:99:May change: <the repository, how it is changed (for example commits on one local branch, a PR never merged), and the texts the change makes wrong>
./skills/hyp-owner-grill/SKILL.md:108:Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Ne
./skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` 
[... 103 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "diff skills/hyp-owner-grill/SKILL.md .claude/skills/hyp-owner-grill/SKILL.md; ls -la .claude/skills; git status --short | head; sed -n 1,30p internal/cli/*.go | head -5; ls internal/cli internal/replay | head -60",
 "description": "Diff repo skill vs running skill; list CLI code"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 1,80p skills/hyp-owner-answer/SKILL.md; grep -n \"stop clause\\|Done when\\|stop line\\|runner owes nothing\" -r internal docs --include=*.go --include=*.md | cut -c1-200 | head -30",
 "description": "Read owner-answer skill and stop-line sources"
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
[... 6086 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"runner owes nothing\" internal cmd docs skills | cut -c1-180 | head; grep -n \"func.*[Aa]ddGoal\\|\\\"goal\\\"\" internal/cli/*.go | grep -v _test | head; grep -n \"HYP_ACTOR\\|func Main\\|case \\\"status\\\"\" internal/cli/cli.go | head; grep -n \"^## \\|^### \" docs/replay-spec.md | head -50",
 "description": "Locate stop line, add goal and spec sections"
}
```

Result:

```text
internal/render/testdata/status_retracted.golden:5:  runner owes nothing: stop and report to the owner
internal/render/testdata/status_before.golden:5:  runner owes nothing: stop and report to the owner
internal/render/status.go:207:		// runner owes nothing there. Runner work prints under OWED, and a
internal/render/status.go:361:		add(kKeep, "  runner owes nothing: stop and report to the owner")
internal/render/status.go:644:		owes := "runner owes nothing"
internal/cli/reach_test.go:39:	want(t, out, 0, "runner owes nothing")
internal/cli/owed_test.go:17:// the runner owes nothing.
internal/cli/owed_test.go:29:	want(t, hyp(t, "status"), 0, "runner owes nothing: stop and report to the owner\n")
docs/principles.md:47:when the ledger owes it, and a run stops when the runner owes nothing. Nothing
docs/adr/0015-the-select-is-the-gate.md:75:Retiring the verifier must not make the runner a verifier in prose. Run 3 of the multi-tenant tree put a reading copy on `R-M4ABBF`, unde
internal/cli/write.go:236:	"goal": event.KindGoal,
internal/cli/write.go:888:	goal := w.fs.String("goal", "", "the Goal this checkpoint is on; required when the store holds more than one Goal")
55:Every write command needs --why and an actor (--actor or $HYP_ACTOR).
13:## 1. Names and locations
25:## 2. Ids
33:## 3. Event envelope
45:## 4. Event types (closed set)
94:## 5. Refusals
138:### 5.1 Write-time refusals
153:## 6. Merge order
161:## 7. Derived state (forward pass)
175:### 7.1 Hypothesis status (first matching rule wins)
187:### 7.2 Combined status graph (for cycle refusal)
191:### 7.3 Stale (snapshot rule, the minimal form of truth maintenance)
207:### 7.4 DesignQuestion state
231:### 7.5 Goal
236:### 7.6 Experiments and Results
263:## 8. Projection (what replay returns)
348:### 8.1 Ledger: what the tree owes
389:### 8.2 Keyed items: limits, lines and readings
418:## 9. Sandbox lifecycle and steps
422:### 9.1 Steps and sections
426:### 9.2 hyp runs every command
430:### 9.3 How a command run
[... 65 more characters]
```

## Grill

I've done the lookups I can do myself, so this is Round 1. I've written nothing to any store.

**What I found**
- The repo root has no `.tree`. Trees live under `investigations/<topic>/`, so a new purpose gets its own directory and `hyp init` there.
- The installed `hyp` has no `prompt` command. The `hyp-owner-grill` copy I'm running describes one, but `skills/hyp-owner-grill/SKILL.md` in the repo still says the skill writes the prompt by hand. Building `hyp prompt` therefore makes repo text wrong.
- G-D1CH1R's body is prose with no `May change:`, `Read only:` or `Never:` block. Under the rule in the running skill, `hyp prompt` would refuse it with exit 1.
- Run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) is not in the repo, and the issue quotes only run 5's. The run-6 checkpoints show its stop clauses differed from run 5's, for example "selected the candidate on Q-PC04ED".

**Round 1**

1. **Where the tree lives, and who writes.**
   - Today: no store at the root, and every existing tree sits in its own `investigations/` directory.
   - Recommendation: `investigations/goal-prompt/`, with the alias `goal`.
   - Please give me your owner actor name (lowercase letters, digits and hyphens, for example `owner-dat`) and say whether that directory name is fine.

2. **Run 6's prompt text.**
   - Problem: the Done-means line says the generated prompt matches run 6 "clause for clause". Without that text I can only guess its clauses. The stop clauses are the part that varies between runs 5 and 6.
   - Option (a): you paste run 6's prompt. I keep it as a file of the build's test data, and the line pins its sha256 and asserts the candidate's output differs only in the actor, the owner answer it starts from, and the branch and commit.
   - Option (b): I rebuild it from run 5 plus the run-6 checkpoints. This is cheaper, but the line would rest on my guess.
   - Principle: a line that compares output with a baseline names the baseline by hash.
   - Recommendation: (a).

3. **G-D1CH1R cannot be generated as it stands.**
   - Problem: the Done-means line says to generate from G-D1CH1R "as it stood before run 6". That body has no constraints block, a Goal is never edited, and `hyp prompt` refuses a Goal without a valid block. Run 6's constraints were hand-written in the prompt.
   - Option (a): the line is met on a synthetic Goal in `internal/*/testdata` with the same shape and a valid block. Core tests read only synthetic data, so this fits the rule anyway. The "Never lines come from a recorded place" clause is then proven on that Goal.
   - Option (b): `hyp prompt` gets a `--constraints FILE` escape for old Goals. That brings back the hand edit the owner rejected.
   - Option (c): drop the G-D1CH1R wording from Done-means and say the old Goal cannot be generated.
   - Recommendation: (a). Should I also note in the Goal body that G-D1CH1R is out by design?

4. **One Goal or two.** The work has two parts:
   - the `hyp prompt` command, with the warning from `hyp add goal` and the refusal on a bad block;
   - the skill text: `hyp-owner-grill`'s Hand off, `hyp-goal-run`'s "discovery" description, and `hyp-prototype`'s description.
   - Both parts are one hyp commit, and ADR 0021 says a skill change is a build Goal. One run could cover every candidate for every line if the skill lines are checked by `go test`, which reads `skills/` as skillcheck already does. This is a test I can name: it fails on any skill description that says "discovery" only.
   - Recommendation: one Goal. #114 (what a commit-candidate Goal needs from hyp) stays a separate Goal, which you add.

5. **Is the typed place for constraints already decided?**
   - Your 2026-11-11 comment settled this: headings in the Goal body, with no typed field until a run misreads one. I read that as closed and will not make it a DesignQuestion. It also satisfies the derivation guard, because hyp derives the block from the body and stores no copy.
   - Please confirm that I read it right.

Once 2 and 3 are answered I can draft the lines and the routing table, including which settled questions the tree records as DesignQuestions.

## Owner, message 2

1. My actor name is owner-standin. I lean toward one shared store at the repo root, since it may now hold many Goals. Does a root store break something concrete? If it does, use `investigations/goal-prompt/` with alias `goal`. Then keep the root-store idea as a separate issue.

2. (a) is fine. I can't paste run 6's prompt in this session, so please note that it has to be supplied. Widen the exceptions to the actor, the owner's answer, branch, PR and commit, and the run's own plan. Branch and PR become flags.

3. (a). Yes, note in the Goal body that G-D1CH1R is out by design. A Goal body without the constraint headings is refused with exit 1, naming what is missing, but adding such a Goal stays allowed.

4. One Goal. I don't want a line that fails on skill descriptions saying "discovery" only. Those descriptions are text the change makes wrong, so list them under May change. #114 stays separate.

5. Yes, you read it right. The headings stay in the Goal body, with no typed field.

## Grill

I'll test your root-store question on a scratch copy, not in the repo.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"walk\\|filepath.Dir\" internal/store/*.go | grep -v _test | head -12; grep -rn \"tree name\\|TreeName\\|<tree>:\" internal --include=*.go -l | head; S=$TMPDIR/rs && rm -rf $S && git clone -q . $S && cd $S && HYP_ACTOR=t hyp init 2>&1 | tail -2; ls -a | head -5; for d in investigations/checkpoint-store investigations/id-citation; do (cd $d; hyp status 2>&1 | head -2; hyp check 2>&1 | tail -1); done; hyp bundle check 2>&1 | tail -3; hyp status 2>&1 | head -3",
 "description": "Test a root store beside nested trees in a scratch clone"
}
```

Result:

```text
internal/store/cut.go:124:		if filepath.Dir(rel) == "." && strings.HasPrefix(rel, "E-") && !c.exps[rel] {
internal/store/store.go:46:// FindRoot walks up from start to the directory that contains .tree/events.
internal/store/store.go:56:		parent := filepath.Dir(dir)
(eval):1: no matches found: --include=*.go
initialised .tree in /tmp/claude-501/rs
.
..
.git
.gitignore
.tree
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
check: 0 errors, 6 warnings
GOAL G-4S36M5 A write command never appends text that cites a node id thi… — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
check: 0 errors, 0 warnings
warn: recipes/recipe-authoring-arms@4 score: no External: line
warn: recipes/recipe-authoring-arms@4 verdict: no External: line
recipe check: 6 trees, 6 pins, 0 drift
GOAL none (hyp add goal --title T --why W)
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"external_ref\" docs/replay-spec.md | head -5; grep -n \"tree name\\|name of a tree\\|directory name\\|basename\" docs/replay-spec.md docs/glossary.md | head",
 "description": "Find how a tree's name is derived"
}
```

Result:

```text
docs/replay-spec.md:64:| `edge.add`, `edge.remove` | the edge's `from` node id | `type` (`depends_on`, `supersedes` between two Hypotheses or two Experiments, `external_ref`), `to` (node id; for `external_ref` a string such as `gh:976` or a URL, at most 300 runes) |
docs/replay-spec.md:245:- A **pin** is a recipe ref (`name@version`) and the hash an event recorded for it at write time. A bundle is its ref and hash; a `recipes/<name>@<version>/` directory name is only the library's lookup key, and hyp opens each copy it makes by the ref it holds ([ADR 0005](adr/0005-a-bundle-is-its-ref-and-hash.md)). A Result depends on its Experiment's pin (`recipe` and `recipe_hash`, unless the recipe is `none`), on its Experiment's rubric pin (`rubric` and `rubric_hash`, when it names a bundle), and on one pin per entry of its own `evidence_hashes`, whose ref is the pointer's first segment under `recipes/` after cleaning the path. A Result is `questionable_evidence` when, for any pin it depends on, `recipeHashes[ref]` is missing or differs from the pinned hash. A new version of a recipe is a new ref, so it never flags Results recorded under an older one. `recipeHashes` leaves out a bundle that cannot be hashed (a symlink, a file or directory name outside `[A-Za-z0-9._-]`, or a `recipe.md` whose name or version does not match its directory; a name below a known answer's `known-answers/NAME/input/` may also hold `@`, so an input can hold a store with its pinned bundles) and a ref whose copies hash differently, and `hyp check` names each one. A bundle that hashes but does not open stays in `recipeHashes`, so its pins verify, and `hyp check` names it too (sandbox spec §1.2). This is a flag only and never changes any status.
```

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader on prompt-command rules",
 "subagent_type": "general-purpose",
 "prompt": "You are an independent reviewer. Repo: /Users/datguy/.hyp-sandbox/g146-base-a-1/wt (Go; the `hyp` tool; read AGENTS.md, docs/replay-spec.md, docs/glossary.md, skills/hyp-owner-grill/SKILL.md, .claude/skills/hyp-owner-grill/SKILL.md, skills/hyp-goal-run/SKILL.md, internal/cli/write.go, internal/render/status.go, internal/store/store.go). Do NOT edit anything in the repo; scratch work only under $TMPDIR (clone the repo there if you want to run hyp).\n\nWe plan to add a read command `hyp prompt <G-id> --actor A [--branch B] [--pr N] [--results N]` that prints a `/goal` prompt for an unattended run, derived only from the event log (no stored field). It prints: the store's path, actor, Goal id, the skill name hyp-goal-run, stop clauses (runner owes nothing; N Results recorded under the actor, default 5; cannot take the next step honestly), then the Goal body's \"constraints block\" copied line for line, then one clause each for --branch and --pr. The block rule: it runs from the first line of the Goal body that starts with `May change:`, `Read only:` or `Never:` to the end of the body; it must contain each of the three headings at least once, and every non-empty line in it must start with one of them (a line of spaces, a tab or a lone CR is not empty). A body without a valid block makes `hyp prompt` exit 1 naming what is missing. `hyp add goal` with a body that has a heading but no valid block still succeeds (exit 0) but warns; a body with no heading adds silently; replay never rejects either. Exit codes: 1 rule refusal, 2 bad arguments (e.g. a branch with a space, bad actor name), 3 environment (no store found). The repo root will hold a store (.tree at root) beside existing nested stores under investigations/*/.tree; a store may hold many Goals.\n\nWalk these rules through scenarios the designers did not draw, e.g.: two Goals in one store (which block, which counts); a merged log from two actors (is a Result 'under the actor' counted from the right events, does the Goal add exist in the log the run starts from); a base that moves mid-run (--branch); a Goal id that is an alias, retracted/abandoned/complete Goal, or not a Goal id; body with CRLF, Unicode whitespace, heading mid-line, heading in a prose paragraph before the block, the same heading twice, indented headings, case differences; output size; a nested store found first when running from investigations/x; --pr with a Never line that says never open a PR (widening). Report concrete problems with evidence from the code (file:line), and which rule needs a decision. Keep the report under 500 words; no praise.",
 "run_in_background": true
}
```

Result:

```text
[redacted: the loaded grill skill]
```

## Grill

**Round 1 is settled.** Your answers fix the actor (`owner-standin`), the shared store, the baseline exceptions and the one-Goal split. A second reviewer is walking the `hyp prompt` rules through scenarios we didn't draw. Its findings come back before I confirm the routing.

**Does a root store break something?** I tested on a scratch clone and found nothing that breaks.
- `hyp init` at the root works.
- From inside `investigations/checkpoint-store` and `investigations/id-citation`, `hyp status` and `hyp check` still find the nearer nested store. This matches the walk-up rule in `internal/store/store.go`.
- `hyp bundle check` still passes (6 trees, 6 pins, 0 drift).

So I'm using a root store and I'm not writing the separate issue. One text becomes wrong: AGENTS.md step 3 says "Under `investigations/<topic>/`, run `hyp init`". I'll list that line under May change.

**Round 2** has three owner decisions.

1. **The baseline line cannot be written yet.**
   - Problem: a line that compares output with a baseline must name the baseline by sha256, and lines are never edited. Run 6's prompt isn't here, so there is no hash.
   - Option (a): hold the whole Goal add until you supply the text.
   - Option (b): add the Goal now with structural lines for the prompt's contents, and carry the baseline match in a later Goal. The text goes into that Goal when you supply it.
   - Under (b), Done-means item 1 is unmet when this Goal completes. Per AGENTS.md, I must open a placeholder issue first that states the unmet goal, and the closing comment on #113 names that issue.
   - Recommendation: (b), since you can't supply the text now.

2. **Which stop clauses the prompt prints, and what `N` counts.**
   - Run 5 had four clauses: the runner owes nothing; or only add-rival items remain with no owner ask for a new Hypothesis; or 5 Experiments recorded; or "cannot take the next step honestly".
   - The skill copy I'm running prints only "owes nothing, or N Results under the actor".
   - Run 6's checkpoints show "cannot take the next step honestly" still in use.
   - My recommendation is three clauses: owes nothing; N Results recorded under the actor (default 5); and "cannot take the next step honestly and say why".
   - I'd drop the add-rival clause. You can keep it if you want it.
   - Counting Results by actor is something replay derives from the log, so nothing is stored. A new runner actor per run makes it equal to "since this goal began".
   - The unit: the flag is `--experiments`, but one Experiment can hold several Results. I recommend counting Results and renaming the flag `--results`. The flag exists only in skill text so far.

3. **What gets recorded as a DesignQuestion.** I propose exactly one: *how does `hyp prompt` find the constraints block?*
   - Rival A: from the first heading line to the end of the body, with every non-empty line starting with a heading.
   - Rival B: only the trailing run of heading lines.
   - The separating scenario is a body with a heading line inside prose, followed by more prose. A refuses it. B accepts the trailing lines.
   - The skill text in PR #135 already chose A. That gives a DQ decided by argument, with this session as evidence.
   - I'd record nothing else as a DQ. The flags, the exit codes and the no-stored-field rule are your decisions or derived from the log.

**Draft lines**, all checked by `go test` on the candidate:
- **[P] Print.** `hyp prompt G --actor A` on a Goal with a valid block exits 0. It prints, in order, the store's path, A, G, `hyp-goal-run`, the stop clauses, and the block's lines byte for byte. Two runs on the same log print identical bytes.
- **[P] Flags.** `--results N` changes only the count and defaults to 5. `--branch B` and `--pr N` each add one clause after the block and change nothing before it.
- **[N] Block.** A body without a valid block exits 1 and names what is missing. The three headings must each appear, and every non-empty line must start with one. Unreadable cases include spaces, tabs and lone CRs that look empty.
- **[N] Arguments and environment.** A bad actor name, or a `--branch` with a space, exits 2. A run with no store above it exits 3. An unknown or non-Goal id exits 1.
- **[N] Add.** `hyp add goal` with a body that has a heading but no valid block exits 0 and warns. A valid block or no heading at all gives no warning. Replay never rejects either.
- **[N] Read-only.** `hyp prompt` writes no event and no file, and stores no copy of the block, so the store is byte-identical before and after.
- **[N] One Goal.** In a store with two Goals, the output holds only the named Goal's id and block.

These need no line: the skill descriptions (you ruled that out), the skillcheck gate (it already fails on a missing command), and the baseline match (option (b) above).

**The hand-off for this grill.** The installed `hyp` has no `prompt`, so I can't print the prompt from the log. I'll write it by hand to the repo's `skills/hyp-owner-grill` rule: under 4000 characters and labelled hand-written. The built `hyp prompt` should reproduce it, which is a useful self-check for the run.

Please answer 1 to 3. After that comes the routing table for your confirmation.

## Owner, message 3

<task-notification>
<task-id>a3ca2cfede3aba7c4</task-id>
<tool-use-id>toolu_01K2J4VBxfViFPiwtbokUS6c</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-1-wt/560fedd4-84a6-43c0-b113-e41924c3198b/tasks/a3ca2cfede3aba7c4.output</output-file>
<status>completed</status>
<summary>Agent "Second reader on prompt-command rules" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Findings, most serious first. Everything below is from reading the code; I did not run a clone.

**Needs a decision**

1. **"N Results under the actor" has no baseline.** `Res`/`Node.Actor` is the actor of the add event (`internal/replay/rules.go:50`). Replay merges the actor files by ID (`replay.go:13`).
   - The count is cumulative. A second run with the same actor on the same Goal stops at once if it already has N Results.
   - The count has no Goal scope. A Result under a different Goal in the same store, or one under a retracted Result, would count unless you filter. Pick one: Results under this Goal, non-retracted, by that actor. Say that the prompt prints this.
   - A reused actor name on another branch or clone gives the same cumulative count.
   - The run's own `--actor` must differ from the owner's actor. `hyp prompt` could refuse an actor that wrote the Goal add, or warn.

2. **Two Goals in one store.** The block comes from the named Goal's body, which is fine. But `hyp status` shows only a GOALS table in that case (`skills/hyp-goal-run/SKILL.md:13`), so the prompt must say `hyp status &lt;G-id&gt;` and `hyp checkpoint --goal &lt;G-id&gt;`. Otherwise it does not fit the skill.
   - A Goal id typed as an alias or a display id (`G-xxxx`) resolves through `resolveRef` (`internal/cli/resolve.go`).
   - `resolveRef` accepts any node. Refuse a non-Goal with exit 1, or 2 for bad arguments. Decide which.
   - Print the resolved display id, not the alias. A later alias collision would change what it means.

3. **Goal lifecycle.** There is no Goal retract or abandon, only `in_progress` or `complete` (`replay/state.go:419`). So the retracted/abandoned case does not exist. Decide whether `complete` refuses (exit 1) or warns. Nothing is owed on a complete Goal, so the run stops at once under clause 1.

4. **Missing Goal add in the run's log.** The Goal add is in the owner's file, and the runner starts from a checkout. If the owner's branch is unmerged, `hyp prompt` reports "unknown node" at exit 2 (`resolve.go:36`).
   - That is an environment or state problem, not bad arguments. Name the cause in the message.
   - `--branch` has the same problem. If main moves mid-run, nothing in the prompt says to rebase or to re-read the Goal. The Goal body is immutable (no amend for a Goal in the spec), so the block cannot drift. Owner answers and checkpoints can.
   - Decide whether the prompt tells the run to start from `--branch` or merely commit to it. Also check `--branch` against git ref rules (`..`, a leading `-`), not only against spaces.

5. **Widening.** `--pr N` adds a clause that can contradict `Never: open a PR`. The skill says a prompt "may narrow, never widen" (`goal-run/SKILL.md:12`). Refuse if any `Never:` line matches "PR" or "pull request"? That is a heuristic. Alternative: print the clause as "only if the constraints allow it". Decide, and say which one the spec uses.

6. **Nested store found first.** `FindRoot` stops at the first `.tree/events` going up (`store.go:47`). From `investigations/x`, `hyp prompt` prints that nested store's path while the owner meant the root store. The prompt says "store path", so print it absolute and print the Goal id alongside, and make the Goal-not-found error name the store it searched. Consider a `--store` flag; there is no such flag now.

**Block rule gaps**

7. **Indentation.** "Starts with" has no stated trim rule. `  May change:` and `- May change:` do not start the block under a literal rule. A body that has the heading only indented gives "no valid block", and `add goal` warns. State that explicitly. Also say whether case matters (`may change:`).

8. **Heading in prose before the block.** "First line that starts with a heading" is unambiguous. But a prose line starting `Never: …` before a real block becomes the start, and the intervening prose lines fail the every-line rule. The error then blames the wrong line. Report the failing line number.

9. **Whitespace definition.** The rule says a line of spaces, a tab or a lone CR is not empty, so a trailing blank line before the end of the body is invalid. Many editors and `--body` heredocs add one. A trailing `\n` would be a refusal. Define "end of body" as trimmed, or accept the trailing newline only.
   - CRLF: a line ending `\r` is non-empty and does not start with a heading. Decide whether to strip one trailing CR per line, and print it as written or normalised. Copy-line-for-line then differs from the stored text.
   - Unicode spaces (NBSP, U+2028) split differently from `strings.Split` on `\n`. Fix the split to `\n` only.
   - A repeated heading is allowed ("at least once"), which is fine, but the run then has two `Never:` lines. State that this is intended.

10. **Size.** `MaxGoalBody` is 2000 runes (`event.go:116`). The grill skill says the prompt is under 4000 characters (`hyp-owner-grill/SKILL.md:110`). The block is at most 2000 runes, but the fixed text plus a long path or branch may exceed 4000 characters. A deep store path alone could do it. Decide whether `hyp prompt` refuses over 4000, and whether that is a rule refusal (exit 1) or a warning.

11. **Warn on `add goal`.** The warning path exists for Experiments (`write.go:416`, `admitted`). `add goal` has no admission step (`write.go:283-331`), so a new warn needs a new place, run before the add. Also, "heading present but no valid block" warns, but a body with no heading is silent, so the owner learns only when `hyp prompt` fails. Consider warning on both.</result>
<usage><subagent_tokens>50625</subagent_tokens><tool_uses>6</tool_uses><duration_ms>36612</duration_ms></usage>
</task-notification>

## Grill

Nothing from you yet on Round 2. The background reviewer has finished, and I'm checking its facts against the source before relying on them.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"MaxGoalBody\" -r internal | head -3; sed -n 25,45p internal/cli/resolve.go; grep -n \"in_progress\\|\\\"complete\\\"\" internal/replay/*.go | grep -v _test | head -5; sed -n 44,56p internal/store/store.go; sed -n 283,331p internal/cli/write.go | grep -n \"warn\\|admit\" ",
 "description": "Verify reviewer's code claims"
}
```

Result:

```text
internal/cli/write.go:255:		maxBody = event.MaxGoalBody
internal/replay/validate.go:85:		maxBody = event.MaxGoalBody // a Goal's body holds the owner's met rule
internal/event/event.go:117:	MaxGoalBody     = 2000
	}
	for _, matches := range [][]string{byAlias, byDisplay} {
		switch len(matches) {
		case 0:
			continue
		case 1:
			return matches[0], nil
		default:
			return "", herr.BadArgs("ambiguous node %s", ref)
		}
	}
	return "", herr.BadArgs("unknown node %s", ref)
}

// display returns the display id of node id, or id itself when unknown.
func display(p replay.Projection, id string) string {
	if n := p.Nodes[id]; n != nil {
		return event.DisplayID(n.Kind, id)
	}
	return id
}
internal/replay/owed.go:39:	AimPlanned   = "in_progress"       // an aiming Experiment has no Result yet
internal/replay/replay.go:511:// answered live top-level DQ to be complete, so a new Goal is in_progress.
internal/replay/state.go:419:	GoalComplete   GoalState = "complete"
internal/replay/state.go:420:	GoalInProgress GoalState = "in_progress"
func eventsDir(root string) string { return filepath.Join(root, ".tree", "events") }

// FindRoot walks up from start to the directory that contains .tree/events.
func FindRoot(start string) (string, error) {
	dir, err := filepath.Abs(start)
	if err != nil {
		return "", err
	}
	for {
		if fi, err := os.Stat(eventsDir(dir)); err == nil && fi.IsDir() {
			return dir, nil
		}
		parent := filepath.Dir(dir)
14:	var warns []string
```

## Grill

The reviewer finished. The completion notice is an automated event, not an answer from you, so **Round 2 is still unanswered** and I've written nothing.

I checked the reviewer's claims against the source. These hold:
- An unknown node is `BadArgs` (exit 2), and an alias or display id resolves to any node, not only a Goal.
- A Goal body is capped at 2000 runes.
- A Goal is only `in_progress` or `complete`.
- The store walk stops at the first `.tree/events` going up.

I did not re-verify the line-level claims about actor attribution and the add-goal warning path. The run will meet them as its own cases.

None of these change Round 2. They add six decisions, each a rule you choose. The run can't settle any of them by measuring, so none becomes a DesignQuestion.

**Round 3, with my recommendations**

1. **What "N Results under the actor" counts.**
   - Today a Result carries the actor of its add event, and replay merges every actor's file. A bare count would include other Goals' Results, retracted Results, and Results from an earlier run that reused the actor name.
   - Recommendation: count non-retracted Results under the named Goal by that actor, and have the prompt say so. The run starts a new actor name each time, so this is the run's own total.
   - Also: refuse an actor that wrote the Goal's add (exit 1). That is the owner's name.

2. **Which Goals and ids `hyp prompt` accepts.**
   - A non-Goal id exits 1, not the 2 that `resolveRef` gives. An unknown id exits 1 and names the cause, such as "the Goal's add is not in this log: merge the owner's branch".
   - The prompt prints the resolved `G-…` id, never the alias. In a multi-Goal store it tells the run to use `hyp status <G-id>` and `hyp checkpoint --goal <G-id>`.
   - A complete Goal is refused with exit 1, since the run would stop at once.

3. **`--pr` can contradict `Never: open a PR`.** hyp can't tell what a `Never:` line means, and a keyword match is a guess. Recommendation: the `--pr` clause is worded "the PR the run pushes to, only if May change allows it". The block still prints first, and the skill's rule that the block wins stays. Neither flag is checked against the block.

4. **`--branch` and `--pr` values.** Reject a value that is not a valid git ref name, such as `..`, a leading `-`, or a space. That is exit 2. A PR must be a positive integer. The prompt says to commit to the branch and doesn't say to rebase. The Goal body is immutable, but answers and checkpoints can change, so the run re-reads them with `hyp status`.

5. **The block rule, made exact.**
   - A heading must start at column 0, case-sensitive. `  May change:` and `- Never:` don't count.
   - The body splits on `\n` only. One trailing `\r` per line is stripped for the test, and the line prints as stored.
   - A trailing blank line at the very end of the body is allowed. A line of spaces, a tab or a lone CR elsewhere is not empty.
   - The refusal names the first failing line number.
   - A repeated heading is intended.

6. **Size and where the store prints.**
   - The block can be up to 2000 runes, and the skill sets 4000 characters for the whole prompt. Recommendation: `hyp prompt` refuses over 4000 with exit 1 and says by how much.
   - It prints the absolute store path and names it in every error. A nested store is found first, so the Goal-not-found error says which store it searched. I would add no `--store` flag yet.
   - On the add warning: warn on a body with a heading but no valid block, as the skill says. A body with no heading adds silently. Say if you want a warning there too.

All six become test cases in the lines above, not extra lines: 1 and 2 extend Print and the Arguments line, 4 the Flags line, and 5 the Block line. Line 3 needs no line, because the block is already printed first.

Round 2 is waiting on your answers about the baseline line, the stop clauses and the unit, and the one DesignQuestion. Round 3 is waiting on these six. After both, I'll send the routing table.

## Owner, message 4

I never received Round 2, only this Round 3 message, so please resend Round 2 and I'll answer it.

Round 3: I accept your recommendation on all six, 1 to 6. For 6, adding a Goal with no heading stays silent, with no warning there, and adding stays allowed.

## Grill

Round 3 is recorded: you accepted all six, and a Goal with no heading adds silently. Here is Round 2 again. It has three questions.

**Round 2**

**1. The baseline line cannot be written yet.**
- **Problem:** the Done-means line says the generated prompt matches run 6's prompt clause for clause. A line that compares output with a baseline must name the baseline by sha256, and lines are never edited. Run 6's prompt is not in the repo and you can't paste it now, so there is no hash to name.
- **Option (a):** hold the whole Goal add until you supply the text.
- **Option (b):** add the Goal now with structural lines (what the prompt contains, in what order, how flags and errors behave). The run-6 comparison goes into a later Goal, which you add once you supply the text, with its hash. Done-means item 1 is then unmet when this Goal completes. AGENTS.md says I must open a placeholder issue first that states the unmet goal, and #113's closing comment names it.
- **Recommendation:** (b), since the text can't come in this session. Under (b), the exceptions you set (actor, owner's answer, branch, PR, commit and the run's own plan) go into the later Goal's line and into this Goal's body as a note.

**2. Which stop clauses the prompt prints, and what `N` counts.**
- **What exists:** run 5's prompt had four clauses: the runner owes nothing; or only add-rival items remain with no owner request for a new Hypothesis; or 5 Experiments recorded; or "you cannot take the next step honestly and say why". The skill copy I'm running prints only "owes nothing, or N Results under the actor". Run 6's checkpoints still cite "cannot take the next step honestly".
- **Recommendation:** three clauses.
  1. `hyp status <G-id>` shows `runner owes nothing: stop and report to the owner`.
  2. N Results recorded under the actor (default 5).
  3. You cannot take the next step honestly; say why.
- **Dropped:** the add-rival clause. Keep it only if you want it.
- **Unit:** one Experiment can hold several Results. I recommend counting Results, using the Round 3 counting rule, and renaming the flag from `--experiments` to `--results`. The old name exists only in skill text so far.

**3. What the tree records as a DesignQuestion.** I propose exactly one: *how does `hyp prompt` find the constraints block?*
- **Rival A:** from the first line starting with a heading to the end of the body, with every non-empty line starting with a heading.
- **Rival B:** only the trailing run of heading lines.
- **Separating scenario:** a body with a prose line that starts `Never: …`, then more prose, then a real block. A refuses it and names the first failing line. B accepts the real trailing block.
- **Status:** PR #135's skill text already chose A, and Round 3 item 5 sharpened it. The DQ would be recorded as decided by argument, with this session as evidence. Per the skill, that means the DQ, both rivals, an Experiment under rival A (`--recipe none --rubric reading`) whose body states the scenarios, a Result with a verdict per rival, and then the select.
- **Alternative:** record no DesignQuestion. The block rule becomes a plain line and body text.
- **Recommendation:** record it. The scenario is a test a run can execute, so the skill says it is a DQ.

Please answer 1 to 3. After that I'll send the routing table, with one row per question, for you to confirm before I write anything.

## Owner, message 5

1. (b). Add the Goal now with structural lines. Put the run-6 comparison in a later Goal once I supply the text. My exceptions go into that later line and into this Goal's body as a note: the actor, the owner's answer, branch, PR and commit, and the run's own plan. Branch and PR are flags.

2. I accept the three clauses and dropping the add-rival clause. On the unit, I want the cap to be on the run's Experiments, as a flag with a default of 5. Round 3 item 1 already counts Results for "N Results under the actor", and that is fine. Keep my Experiments cap unless you show a concrete reason Experiments can't be counted under the actor. In that case, take your fix.

3. Record the DesignQuestion, as you recommend. Who writes the prompt (hyp, as a command that reads the log) and whether it carries my answers (no, the runner reads OWED) are settled here. Neither needs testing.

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 40,60p internal/replay/rules.go; grep -n \"Withdrawn\\|Actor \" internal/replay/state.go | head -12",
 "description": "Check Experiments carry an actor and withdrawn state"
}
```

Result:

```text
				// The derived records still hold the state before this event.
				// An Experiment under a DesignQuestion has no rivals: it
				// gives no verdict.
				st.RivalsAtAdd[ev.ID] = st.LiveRivals(st.Nodes[p.Parent].Parent, p.Parent)
			}
			rubric := p.Rubric
			if p.DecidedBy == event.DecidedReading {
				rubric = event.RubricReading // the older spelling (ADR 0018)
			}
			st.addNode(&Node{ID: ev.ID, Kind: p.Kind, Parent: p.Parent, Title: p.Title, Body: p.Body,
				Alias: p.Alias, Recipe: p.Recipe, RecipeHash: p.RecipeHash, DecidedBy: p.DecidedBy, Actor: ev.Actor,
				AddedBy: ev.ID, Targets: p.Targets, Aims: p.Aims, Rubric: rubric, RubricHash: p.RubricHash, Records: p.Records})
			if p.Alias != "" {
				st.Aliases[p.Alias] = ev.ID
			}
			if p.Kind == event.KindGoal {
				st.Goals = append(st.Goals, ev.ID)
				if len(p.Lines) > 0 {
					// The Goal's add counts as the one write of its lines, and
					// each cite keeps what it meant here (ADR 0020).
					cites := make([][]Cite, len(p.Lines))
46:func (s QuestionState) Obliges() bool { return s != DQWithdrawn && s != DQAnswered }
287:	ID, Hyp, Actor string
313:	Withdrawn map[string]bool // Experiment → it has an applied experiment.withdraw
315:	Hold      map[string]QuestionState // DQ → "", DQDeferred or DQWithdrawn
373:			Withdrawn:    map[string]bool{},
409:	DQWithdrawn QuestionState = "withdrawn"
414:	ExpWithdrawn  ExpStatus = "withdrawn"
```

## Grill

Both counts are derivable from the log, so I have no concrete reason against your Experiments cap. An Experiment node keeps the actor of its add event (`internal/replay/rules.go:50`), and replay tracks withdrawn Experiments. The cap counts Experiments added under the named Goal by the run's actor and not withdrawn. The flag stays `--experiments`, default 5. So clause 2 reads "N Experiments recorded under <actor>". Round 3 item 1's counting rule still holds, with Experiments in place of Results.

I've written nothing yet. Here is the routing table. Please confirm each row or move it.

| # | Question | Goes to | Notes |
|---|---|---|---|
| 1 | What the purpose is: `hyp prompt` prints the `/goal` prompt for any Goal; constraints come from the body headings | Goal title and body | Title: "hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body". The body carries your note: the exceptions (actor, owner's answer, branch, PR, commit, the run's own plan) and "G-D1CH1R is out by design" |
| 2 | What a run may change, read or never do | Body `May change:`, `Read only:`, `Never:` | May change: commits on one local branch of this repo covering `cmd/`, `internal/`, `skills/` (the hyp-owner-grill Hand off, the "discovery" wording in the hyp-goal-run and hyp-prototype descriptions), `docs/` and `README.md`, and AGENTS.md step 3's "under `investigations/<topic>/`". A PR may be opened and never merged. Read only: `investigations/*` and `recipes/*`. Never: replace `~/.local/bin/hyp`, edit another tree or a recorded event, add a Goal or a line, write under the owner's actor |
| 3 | Run-6 match | Later Goal (yours) | I open a placeholder issue first that states the unmet Done-means item 1 |
| 4 | Typed place for constraints | Settled (your 2026-11-11 comment) | No node, no line |
| 5 | Who writes the prompt, and whether it carries your answers | Settled: no test | `hyp prompt` reads the log; OWED carries the answers |
| 6 | Root store | Settled: looked up | Nothing broke; AGENTS.md step 3 text goes under May change |
| 7 | How `hyp prompt` finds the constraints block | **DesignQuestion**, decided by argument | Rivals: A = first heading line to the end, every non-empty line a heading. B = trailing run of heading lines. Scenario: a prose line starting `Never:`, more prose, then a real block. A refuses it and names the first failing line. B accepts the trailing block. Experiment under A (`--recipe none --rubric reading`) with the scenarios as its rule, then a Result with a verdict per rival. I then add the select OWED owes |
| 8 | The 6 Round 3 decisions | Lines 3 to 7 below | |
| 9 | The 3 stop clauses, the flags and the counting unit | Line 1 and Line 2 | |

**Lines.** All are checked by `go test` on the candidate. Each line pins one invariant, and none names history.
1. **[P]** `hyp prompt G --actor A` on a Goal with a valid block exits 0. It prints, in this order, the absolute store path, A, the resolved `G-…` id, `hyp-goal-run`, the three stop clauses, then the block's lines byte for byte. Two runs on one log print identical bytes. In a many-Goal store it names `hyp status <G-id>` and `hyp checkpoint --goal <G-id>`.
2. **[P]** `--experiments N` (default 5) changes only the number in the count clause. `--branch B` and `--pr N` each add one clause after the block and change nothing before it. The PR clause is worded "only if May change allows it".
3. **[N]** A Goal body without a valid block exits 1 and names the first failing line. Block = from the first column-0, case-sensitive line starting with `May change:`, `Read only:` or `Never:` to the end of the body. All three headings must appear. Every non-empty line must start with one. Spaces, tabs and a lone CR do not count as empty. Split on `\n` only, strip one trailing `\r` for the test, and allow one trailing blank at the very end.
4. **[N]** The Experiments cap counts non-withdrawn Experiments added under the named Goal by A. Another Goal's Experiments, or another actor's, are never counted. An actor that wrote the Goal's add is refused with exit 1.
5. **[N]** A non-Goal id, an unknown id, or a complete Goal exits 1 and names the cause. An unknown id names the store searched and says to merge the Goal's add into the log. An alias never prints.
6. **[N]** A bad actor name, a `--branch` that is not a valid git ref name, or a non-positive PR or experiments count exits 2. No store above the current directory exits 3. A prompt over 4000 characters exits 1 and says by how much.
7. **[N]** `hyp add goal` with a heading but no valid block exits 0 and warns. A valid block, or no heading, adds silently. Replay accepts both.
8. **[N]** `hyp prompt` appends no event, writes no file, and stores no copy of the block: the store is byte-identical before and after.

Out of lines, by your rulings: the skill descriptions, the run-6 match, and the rule that a flag can't widen the block.

**Hand-off.** The installed `hyp` has no `prompt`, so the final `/goal` prompt will be written by hand to the repo's `skills/hyp-owner-grill` rule. I'll label it hand-written. A build Goal also needs the review round the repo skill lists (a fresh subagent per candidate commit, findings sorted into limit, reading, fix commit or out-of-Goal finding), which goes into the prompt.

**Before the first write:** I'll run `hyp init` at the repo root as the runner. I'll then show the exact `hyp add goal` command under `owner-standin` and run it only after you approve that command.

Do you confirm the table, or do you want any row moved?

## Owner, message 6

I confirm the table and all eight lines as they stand. No row needs moving.

## Grill

Table confirmed. First I run `hyp init` at the repo root as the runner. It is not an owner write.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && hyp init --why \"store for the #113 build Goal: hyp prompt\" 2>&1; hyp status 2>&1 | head -3; git status --short",
 "description": "Init the root store as the runner"
}
```

Result:

```text
usage: hyp init
no .tree store found; run 'hyp init'
```

## Grill

`hyp init` takes no `--why`. I'll run it plain.

### Tool Bash (cwd: .)

```text
{
 "command": "hyp init 2>&1; HYP_ACTOR=claude-grill-113 hyp status 2>&1 | head -3; git status --short",
 "description": "Init the root store"
}
```

Result:

```text
initialised .tree in /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
GOAL none (hyp add goal --title T --why W)
?? .tree/
```

## Grill

The store exists. Now I draft the owner's command in a file and check the rune limits (body ≤ 2000, each line ≤ 600). I run nothing under your actor yet.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > body.txt <<'EOF'\nBuilds #113: hyp prints the /goal prompt for any Goal from the log (hyp prompt), and reads a run's constraints from the Goal body's headings. Owner decision 2026-11-11: no typed field until a real run misreads or loses a constraint. The candidate is a hyp commit; go test is the instrument. The change also fixes the texts it makes wrong: the hyp-owner-grill hand-off, the word discovery in the descriptions of hyp-goal-run and hyp-prototype, the usage text, the specs, the glossary and the README.\nOut by design: a Goal written before the headings, such as G-D1CH1R in investigations/checkpoint-store, has no block. hyp prompt refuses it with exit 1 and names what is missing; adding such a Goal stays allowed.\nNot here: matching run 6's prompt (goal-prompt-261111-2129-g1b-run-6) clause for clause. The owner must supply its text first; a later Goal, added by the owner, carries that comparison by sha256. Differences it allows: the actor, the owner's answer the run starts from, the branch, the PR, the commit and the run's own plan. Branch and PR are flags. What a commit-candidate Goal needs from hyp is #114.\nMay change: this repository, as commits on one local branch, in cmd/, internal/, skills/ (the hyp-owner-grill hand-off, the descriptions of hyp-goal-run and hyp-prototype), docs/, README.md and AGENTS.md step 3, where it says where trees live; a PR may be opened and is never merged\nRead only: investigations/ and recipes/, and every store under them\nNever: replace the installed hyp binary\nNever: edit another tree, a recorded event or this Goal\nNever: add a Goal or a line\nNever: write under the owner's actor\nEOF\npython3 -I - <<'EOF'\nb=open('body.txt').read().rstrip('\\n'); print(len(b))\nEOF",
 "description": "Draft Goal body and count runes"
}
```

Result:

```text
1626
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && cat > add-goal.sh <<'EOF'\n#!/bin/zsh\ncd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt\nhyp add goal --actor owner-standin \\\n  --title \"hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body\" \\\n  --alias goal \\\n  --body \"$(cat $TMPDIR/g/body.txt)\" \\\n  --line \"[P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and prints, in this order: the absolute store path, A, the resolved G- id, hyp-goal-run, three stop clauses (hyp status shows the runner owes nothing; N Experiments recorded under A; the run cannot take the next step honestly and says why), then the block's lines byte for byte. Two runs on one log print identical bytes. In a store with many Goals it names hyp status <G-id> and hyp checkpoint --goal <G-id>.\" \\\n  --line \"[P] --experiments N (default 5) changes only the number in the count clause. --branch B and --pr N each add exactly one clause after the block and change nothing before it. The PR clause says the PR applies only if May change allows it.\" \\\n  --line \"[N] A Goal body without a valid block makes hyp prompt exit 1 and name the first failing line. The block runs from the first column-0, case-sensitive line starting May change:, Read only: or Never: to the end of the body; it holds each heading at least once; every non-empty line starts with one. A line of spaces, a tab or a lone CR is not empty. Lines split on \\\\n only, one trailing \\\\r is ignored for the test, and one trailing blank at the very end is allowed.\" \\\n  --line \"[N] The Experiments cap counts only non-withdrawn Experiments added under the named Goal by actor A. Another Goal's Experiments and another actor's are never counted. An actor that wrote the Goal's add is refused with exit 1.\" \\\n  --line \"[N] A non-Goal id, an unknown id or a complete Goal makes hyp prompt exit 1 and name the cause. An unknown id names the store searched and says the Goal's add must be in the log. An alias never prints.\" \\\n  --line \"[N] A bad actor name, a --branch that is not a valid git ref name, or a non-positive --pr or --experiments exits 2. No store above the current directory exits 3. A prompt over 4000 characters exits 1 and says by how many.\" \\\n  --line \"[N] hyp add goal with a body that has a heading but no valid block exits 0 and warns. A valid block, or a body with no heading, adds silently. Replay accepts both logs.\" \\\n  --line \"[N] hyp prompt appends no event, writes no file and stores no copy of the block: the store is byte-identical before and after.\" \\\n  --why \"Owner flow, #113: the /goal prompt comes from the log, and a run's constraints from the Goal body.\"\nEOF\npython3 -I - <<'EOF'\nimport re\ns=open('add-goal.sh').read()\nfor m in re.finditer(r'--line \"(.*?)\" \\\\\\n',s): print(len(m.group(1)))\nEOF\ncd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt && hyp add goal --help 2>&1 | head -5",
 "description": "Draft add-goal comma
[... 26 more characters]
```

Result:

```text
485
236
465
225
201
221
168
126
usage of hyp add goal:
  -actor string
    	actor name (default $HYP_ACTOR)
  -alias string
    	optional unique alias
```

## Grill

The store is initialized. All 8 lines are within 600 runes (the longest is 485), and the body is 1626 of 2000. This is the exact owner write, saved as `$TMPDIR/g/add-goal.sh`, with `$TMPDIR/g/body.txt` as the body. **I have not run it.**

```text
hyp add goal --actor owner-standin --alias goal \
  --title "hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body" \
  --body "$(cat body.txt)" \
  --line "[P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and prints, in this order: the absolute store path, A, the resolved G- id, hyp-goal-run, three stop clauses (hyp status shows the runner owes nothing; N Experiments recorded under A; the run cannot take the next step honestly and says why), then the block's lines byte for byte. Two runs on one log print identical bytes. In a store with many Goals it names hyp status <G-id> and hyp checkpoint --goal <G-id>." \
  --line "[P] --experiments N (default 5) changes only the number in the count clause. --branch B and --pr N each add exactly one clause after the block and change nothing before it. The PR clause says the PR applies only if May change allows it." \
  --line "[N] A Goal body without a valid block makes hyp prompt exit 1 and name the first failing line. The block runs from the first column-0, case-sensitive line starting May change:, Read only: or Never: to the end of the body; it holds each heading at least once; every non-empty line starts with one. A line of spaces, a tab or a lone CR is not empty. Lines split on \n only, one trailing \r is ignored for the test, and one trailing blank at the very end is allowed." \
  --line "[N] The Experiments cap counts only non-withdrawn Experiments added under the named Goal by actor A. Another Goal's Experiments and another actor's are never counted. An actor that wrote the Goal's add is refused with exit 1." \
  --line "[N] A non-Goal id, an unknown id or a complete Goal makes hyp prompt exit 1 and name the cause. An unknown id names the store searched and says the Goal's add must be in the log. An alias never prints." \
  --line "[N] A bad actor name, a --branch that is not a valid git ref name, or a non-positive --pr or --experiments exits 2. No store above the current directory exits 3. A prompt over 4000 characters exits 1 and says by how many." \
  --line "[N] hyp add goal with a body that has a heading but no valid block exits 0 and warns. A valid block, or a body with no heading, adds silently. Replay accepts both logs." \
  --line "[N] hyp prompt appends no event, writes no file and stores no copy of the block: the store is byte-identical before and after." \
  --why "Owner flow, #113: the /goal prompt comes from the log, and a run's constraints from the Goal body."
```

The body, as stored:

```text
Builds #113: hyp prints the /goal prompt for any Goal from the log (hyp prompt), and reads a run's constraints from the Goal body's headings. Owner decision 2026-11-11: no typed field until a real run misreads or loses a constraint. The candidate is a hyp commit; go test is the instrument. The change also fixes the texts it makes wrong: the hyp-owner-grill hand-off, the word discovery in the descriptions of hyp-goal-run and hyp-prototype, the usage text, the specs, the glossary and the README.
Out by design: a Goal written before the headings, such as G-D1CH1R in investigations/checkpoint-store, has no block. hyp prompt refuses it with exit 1 and names what is missing; adding such a Goal stays allowed.
Not here: matching run 6's prompt (goal-prompt-261111-2129-g1b-run-6) clause for clause. The owner must supply its text first; a later Goal, added by the owner, carries that comparison by sha256. Differences it allows: the actor, the owner's answer the run starts from, the branch, the PR, the commit and the run's own plan. Branch and PR are flags. What a commit-candidate Goal needs from hyp is #114.
May change: this repository, as commits on one local branch, in cmd/, internal/, skills/ (the hyp-owner-grill hand-off, the descriptions of hyp-goal-run and hyp-prototype), docs/, README.md and AGENTS.md step 3, where it says where trees live; a PR may be opened and is never merged
Read only: investigations/ and recipes/, and every store under them
Never: replace the installed hyp binary
Never: edit another tree, a recorded event or this Goal
Never: add a Goal or a line
Never: write under the owner's actor
```

Two things to check before you approve:
- **The block follows the rules we settled.** It starts at the first column-0 heading, and every non-empty line from there is a heading line.
- **The Goal carries no `--cite`.** This store holds no settled keys yet.

Do you approve this exact command under `owner-standin`? After you do, I run it and then add the DesignQuestion, its two rivals, the decided-by-argument Experiment and the Result. Then I link the Goal to `gh:113`, run `hyp check`, and write the end-of-grill checkpoint.

## Owner, message 7

Approved. Run that exact command under `owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "zsh $TMPDIR/g/add-goal.sh 2>&1 | tail -20; hyp status 2>&1 | head -20",
 "description": "Run the approved owner command"
}
```

Result:

```text
--body names G-D1CH1R, which is no node here; cite an id this store holds, or write <name>:<id> for another store's
GOAL none (hyp add goal --title T --why W)
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && python3 -I - <<'EOF'\ns=open('body.txt').read()\nold=\"such as G-D1CH1R in investigations/checkpoint-store, has no block.\"\nassert old in s\ns=s.replace(old,\"such as checkpoint-store:G-D1CH1R, has no block.\")\nopen('body.txt','w').write(s); print(len(s.rstrip('\\n')))\nEOF\ngrep -n \"G-D1CH1R\" body.txt | cut -c1-200; grep -c \"\" body.txt",
 "description": "Fix the cross-tree id in the body"
}
```

Result:

```text
1608
2:Out by design: a Goal written before the headings, such as checkpoint-store:G-D1CH1R, has no block. hyp prompt refuses it with exit 1 and names what is missing; adding such a Goal stays allowed.
9
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

## Grill

The first attempt wrote nothing, and the store has no Goal yet. I changed the body as shown above and left the command and the 8 lines as you approved.

The body is now 1608 of 2000 runes. The only place it names G-D1CH1R is line 2, which now reads `checkpoint-store:G-D1CH1R`.

Do you approve running the same `add-goal.sh` command under `owner-standin` with this corrected body? I'll run it as soon as you say yes.

## Owner, message 8

Approved. Run `add-goal.sh` with the corrected body under `owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "zsh $TMPDIR/g/add-goal.sh 2>&1 | tail -20; hyp status 2>&1 | head -14",
 "description": "Run the approved owner command"
}
```

Result:

```text
G-NSXQ3E node.add
  line G-NSXQ3E#1: [P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and prints, in this order: the absolute store path, A, the resolved G- id, hyp-goal-run, three stop clauses (hyp status shows the runner owes nothing; N Experiments recorded under A; the run cannot take the next step honestly and says why), then the block's lines byte for byte. Two runs on one log print identical bytes. In a store with many Goals it names hyp status <G-id> and hyp checkpoint --goal <G-id>.
  line G-NSXQ3E#2: [P] --experiments N (default 5) changes only the number in the count clause. --branch B and --pr N each add exactly one clause after the block and change nothing before it. The PR clause says the PR applies only if May change allows it.
  line G-NSXQ3E#3: [N] A Goal body without a valid block makes hyp prompt exit 1 and name the first failing line. The block runs from the first column-0, case-sensitive line starting May change:, Read only: or Never: to the end of the body; it holds each heading at least once; every non-empty line starts with one. A line of spaces, a tab or a lone CR is not empty. Lines split on \n only, one trailing \r is ignored for the test, and one trailing blank at the very end is allowed.
  line G-NSXQ3E#4: [N] The Experiments cap counts only non-withdrawn Experiments added under the named Goal by actor A. Another Goal's Experiments and another actor's are never counted. An actor that wrote the Goal's add is refused with exit 1.
  line G-NSXQ3E#5: [N] A non-Goal id, an unknown id or a complete Goal makes hyp prompt exit 1 and name the cause. An unknown id names the store searched and says the Goal's add must be in the log. An alias never prints.
  line G-NSXQ3E#6: [N] A bad actor name, a --branch that is not a valid git ref name, or a non-positive --pr or --experiments exits 2. No store above the current directory exits 3. A prompt over 4000 characters exits 1 and says by how many.
  line G-NSXQ3E#7: [N] hyp add 
[... 1832 more characters]
```

## Grill

The Goal is written as `G-NSXQ3E` with 8 lines. Now the runner writes: the DesignQuestion and its two rivals.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-NSXQ3E\nhyp add dq --parent $G --alias q-block --title \"How does hyp prompt find a Goal body's constraints block?\" --body \"Scenario that separates the rivals: a body with a prose line that starts with Never:, more prose, then a real block of May change:, Read only: and Never: lines. Decided by argument in the grill session; the run builds the winner.\" --why \"A testable option set settled in the grill is still a DesignQuestion with rivals.\" 2>&1 | tail -3\nhyp add hyp --parent q-block --alias h-first-heading --title \"The block runs from the first heading line to the end of the body\" --body \"From the first column-0 line starting May change:, Read only: or Never: to the end of the body. Every non-empty line in it starts with a heading and all three headings appear. The scenario body is refused with exit 1 and the first failing line is named.\" --why \"Rival A, chosen in the grill: it matches the rule PR #135 wrote in hyp-owner-grill.\" 2>&1 | tail -3\nhyp add hyp --parent q-block --alias h-trailing-run --title \"The block is the trailing run of heading lines\" --body \"The block is the last run of consecutive heading lines at the end of the body; prose before it is ignored. The scenario body is accepted and the real block is printed.\" --why \"Rival B, rejected in the grill: a heading-like line in prose would pass unnoticed and the first failing line could not be named.\" 2>&1 | tail -3\nhyp status 2>&1 | sed -n '/FRONTIER/,/STALE/p'",
 "description": "Add the DesignQuestion and two rivals"
}
```

Result:

```text
Q-61DHX6 node.add
H-8VWGRN node.add
H-TYQZYP node.add
FRONTIER
  Q-61DHX6 open/untested How does hyp prompt find a Goal body's constraints block?
    H-8VWGRN (h-first-heading) proposed The block runs from the first heading line to the end of th…
    H-TYQZYP (h-trailing-run) proposed The block is the trailing run of heading lines
STALE
```

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp add exp --parent H-8VWGRN --recipe none --rubric reading --target H-TYQZYP --alias e-block-scenarios --title \"Scenarios that separate the two block rules, argued in the grill\" --body \"Rule, fixed before the Result: S1 = a body with a prose line starting Never:, a line of prose, then a block of May change:, Read only: and Never: lines. A refuses S1 (exit 1) and names the first failing line, which is the prose line after the first Never:. B accepts S1 and prints the trailing block. S2 = a body that is only a valid block: both accept. S3 = a block followed by a prose line: A refuses, B accepts. The owner chose A in the session because a heading-like line in prose must not pass unnoticed. Verdicts: A supports on S1 and S3; B contradicts. Evidence: the scenarios file.\" --why \"Records the rival choice as argued, not tested.\" 2>&1 | tail -3",
 "description": "Add the argued Experiment"
}
```

Result:

```text
E-9B59GK node.add
```

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && mkdir -p .tree/evidence/E-9B59GK && cat > .tree/evidence/E-9B59GK/scenarios.txt <<'EOF'\nScenarios for Q-61DHX6 (how hyp prompt finds the constraints block). Argued in the #113 grill session, not run.\n\nS1 body:\n  Context prose.\n  Never: this line is prose, not part of a block.\n  More prose.\n  May change: x\n  Read only: y\n  Never: z\nRule A (first heading line to end): refuses; the first failing line is \"More prose.\" (line 3).\nRule B (trailing run of heading lines): accepts and prints lines 4 to 6.\n\nS2 body: a valid block and nothing else. Rule A: accepts. Rule B: accepts.\n\nS3 body: a valid block, then \"Closing prose.\" Rule A: refuses at the last line. Rule B: accepts the block.\n\nverdict: H-8VWGRN supports\nverdict: H-TYQZYP contradicts\nEOF\nhyp record --exp E-9B59GK --verdict H-8VWGRN=supports --verdict H-TYQZYP=contradicts --observation \"Records the grill's argued choice under the Experiment's pre-registered rule: the owner chose rule A in session; S1 and S3 separate the rivals and rule B fails the refusal case.\" --evidence .tree/evidence/E-9B59GK/scenarios.txt --why \"Result of the argued DesignQuestion.\" 2>&1 | tail -6; hyp status 2>&1 | sed -n '/OWED/,/AWAITING/p'",
 "description": "Record the argued Result"
}
```

Result:

```text
R-7EWXBS result.record
OWED
  runner: Q-61DHX6 limit G-NSXQ3E#1 [owed] — [P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and…
  runner: Q-61DHX6 limit G-NSXQ3E#2 [owed] — [P] --experiments N (default 5) changes only the number in the count clause. --…
  runner: Q-61DHX6 limit G-NSXQ3E#3 [owed] — [N] A Goal body without a valid block makes hyp prompt exit 1 and name the firs…
  runner: Q-61DHX6 limit G-NSXQ3E#4 [owed] — [N] The Experiments cap counts only non-withdrawn Experiments added under the n…
  runner: Q-61DHX6 limit G-NSXQ3E#5 [owed] — [N] A non-Goal id, an unknown id or a complete Goal makes hyp prompt exit 1 and…
  runner: Q-61DHX6 limit G-NSXQ3E#6 [owed] — [N] A bad actor name, a --branch that is not a valid git ref name, or a non-pos…
  runner: Q-61DHX6 limit G-NSXQ3E#7 [owed] — [N] hyp add goal with a body that has a heading but no valid block exits 0 and …
  runner: Q-61DHX6 limit G-NSXQ3E#8 [owed] — [N] hyp prompt appends no event, writes no file and stores no copy of the block…
AWAITING OWNER
```

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && hyp select Q-61DHX6 2>&1 | head; hyp select Q-61DHX6 H-8VWGRN --why \"test\" 2>&1 | head -5; sed -n 207,231p docs/replay-spec.md | cut -c1-400",
 "description": "Probe select admission and DQ state rules"
}
```

Result:

```text
suggested: none (Q-61DHX6 (q-block) owes the runner first: limit G-NSXQ3E#1)
refused: Q-61DHX6 (q-block) owes the runner first: limit G-NSXQ3E#1
### 7.4 DesignQuestion state

Hold state is the latest of `dq.defer` / `dq.withdraw` / `dq.resume`; resume clears it. The latest in merge order wins: a withdraw is undone by a resume, and a resume of a DQ with no hold changes nothing. A defer of a withdrawn DQ is refused (§5 rule 29): the question no longer matters, so parking it says nothing; a resume brings it back first. A withdraw of a deferred DQ applies and replaces the 

One record, the question's `Decision`, owns everything below except the hold state: select history (each select with the candidates when it applied), conflicts, the limits under the question (§8) and the reason order. Render, signals, check and the CLI read it.

Selection is the current selection id (`""` when there is none) plus a conflict set:
- Applying `dq.select` or `dq.unselect` whose `prev_select` ≠ the current selection id adds both ids to the conflict set. The event still applies; the last one in merge order becomes current.
- A later select or unselect whose `prev_select` equals the current id clears the conflict set.
- **Candidates at the select**: when a `dq.select` applies, replay stores the question's candidates just before it, as it stores the live rivals at an Experiment's add (§8). It is a fact derived at apply time, not a payload field, so old logs get it too.
- A **candidate** is a live child Hypothesis whose status is `supported`; neither stale nor an open sub-question removes it. Two or more candidates are a **trade-off**, a state derived from the candidates and never recorded; an older select's stored `tradeoff` is ignored. The owner ends a trade-off with their choice, recorded as the select (ADR 0015): hyp admits that select only when the ledger ow

The state is decided by the first matching rule:
1. `withdrawn`, then `deferred` (from the hold state).
2. `answered` when all of 
[... 1998 more characters]
```

## Grill

I stopped before the link and the checkpoint, because there is a problem in the tree that I should have caught at the grid step.

**What happened**
- I recorded the block-rule DesignQuestion `Q-61DHX6` under the build Goal `G-NSXQ3E`, as the confirmed table says.
- hyp owes every Goal line on every top-level question that has a candidate. `Q-61DHX6` now has a candidate, `H-8VWGRN`.
- All eight lines are therefore owed on it. `OWED` lists them, and `hyp select Q-61DHX6` is refused with "owes the runner first: limit G-NSXQ3E#1".
- No run can cover them. A run's Results judge commits, and none can give `H-8VWGRN` a verdict on "exits 1 with exit code 3 when no store exists".
- The result is that the question stays open forever, or the run files eight refers that you must accept one by one.

This is the case the skill's grid rule warns about: an argued design question and a build with lines serve different purposes. My table put them in one Goal, and you confirmed it without that consequence being visible. That was my miss.

**What exists now**
- The Goal `G-NSXQ3E` (yours, approved) and its 8 lines.
- Runner events: `Q-61DHX6`, the rivals `H-8VWGRN` and `H-TYQZYP`, the Experiment `E-9B59GK`, the Result `R-7EWXBS`, and the evidence file `.tree/evidence/E-9B59GK/scenarios.txt`.
- Events are append-only, so none of this can be deleted. A DesignQuestion can be withdrawn, and a withdrawn question owes no line.

**Options**
- **(a) Move the record to its own Goal.**
  - I withdraw `Q-61DHX6`.
  - You add a second Goal in this store, titled for the block rule, with no lines and a valid constraints block. It is a design record, so the skill says it gets no lines.
  - I re-add the DesignQuestion, the rivals, the Experiment and the Result under it, and the select becomes owed there.
  - `G-NSXQ3E`'s body then cites it, but the body is already stored and can't change. The run finds it by `hyp status`.
  - The cost is one more owner write for you to approve.
- **(b) Withdraw `Q-61DHX6` and keep no DesignQuestion.**
  - The block rule lives only in line 3 and the body.
  - You lose the argued record. The rule for what the tree records said your call decides it.
- **(c) Leave it.** The run files eight refers and you accept them. I don't recommend this.

**Recommendation: (a).** It keeps the argued record you asked for, and the build Goal's OWED stays clean for the run.

Which do you choose? If it's (a), I'll show you the exact `hyp add goal` command for the design-record Goal for approval before I write anything.

## Owner, message 9

(a). Withdraw `Q-61DHX6` and show me the exact `hyp add goal` command for the design-record Goal. I'll approve it before anything is written. Don't re-add the DesignQuestion and the rest under it until I've approved that command.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && hyp withdraw Q-61DHX6 --why \"The argued block-rule question moves to its own design-record Goal: a Goal with lines owes every line on it, and no run can cover them. Owner chose this in the grill.\" 2>&1 | tail -3; hyp status G-NSXQ3E 2>&1 | sed -n '/FRONTIER/,/AWAITING/p'; hyp check 2>&1 | tail -3",
 "description": "Withdraw the misplaced DesignQuestion"
}
```

Result:

```text
Q-61DHX6 dq.withdraw
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
check: 0 errors, 0 warnings
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && cat > design-body.txt <<'EOF'\nDesign record for the build Goal G-NSXQ3E (#113): how hyp prompt finds a Goal body's constraints block. The owner settled it by argument in the grill session. The DesignQuestion, both rival Hypotheses, the argued Experiment and its Result sit under this Goal, and the select closes it. The build Goal carries the lines that check the chosen rule; this Goal has none, because no run checks its candidates.\nMay change: this store, under this Goal only: the DesignQuestion, the rival Hypotheses, one Experiment, its Result, the select and the end checkpoint, written by the runner as events and one evidence file under .tree/evidence/\nRead only: everything else, G-NSXQ3E included\nNever: edit G-NSXQ3E, add a Goal or a line\nNever: write under the owner's actor\nEOF\ncat > add-design-goal.sh <<'EOF'\n#!/bin/zsh\ncd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt\nhyp add goal --actor owner-standin \\\n  --alias goal-block-rule \\\n  --title \"How hyp prompt finds a Goal body's constraints block (design record for G-NSXQ3E)\" \\\n  --body \"$(cat $TMPDIR/g/design-body.txt)\" \\\n  --why \"Owner flow, #113: the block rule was settled by argument in the grill; a Goal with lines would owe every line on it, so the record has its own Goal with none.\"\nEOF\npython3 -I -c \"print(len(open('design-body.txt').read().rstrip('\\n')))\"",
 "description": "Draft the design-record Goal command"
}
```

Result:

```text
757
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

## Grill

This is the exact owner write. The body is 757 of 2000 runes.

```text
hyp add goal --actor owner-standin \
  --alias goal-block-rule \
  --title "How hyp prompt finds a Goal body's constraints block (design record for G-NSXQ3E)" \
  --body "$(cat body.txt)" \
  --why "Owner flow, #113: the block rule was settled by argument in the grill; a Goal with lines would owe every line on it, so the record has its own Goal with none."
```

The body, as stored:

```text
Design record for the build Goal G-NSXQ3E (#113): how hyp prompt finds a Goal body's constraints block. The owner settled it by argument in the grill session. The DesignQuestion, both rival Hypotheses, the argued Experiment and its Result sit under this Goal, and the select closes it. The build Goal carries the lines that check the chosen rule; this Goal has none, because no run checks its candidates.
May change: this store, under this Goal only: the DesignQuestion, the rival Hypotheses, one Experiment, its Result, the select and the end checkpoint, written by the runner as events and one evidence file under .tree/evidence/
Read only: everything else, G-NSXQ3E included
Never: edit G-NSXQ3E, add a Goal or a line
Never: write under the owner's actor
```

- **No lines.** This is the "settled by argument" case the skill gives no lines to. Its `hyp status` will show no LINES.
- **Valid constraints block.** The block starts at the first heading and holds each heading, and every non-empty line in it starts with one.
- **Id check.** The body names only `G-NSXQ3E`, which exists in this store, so the cross-tree refusal from the last attempt does not apply.

Do you approve this exact command under `owner-standin`? Once you do, I run it and re-add the DesignQuestion, the two rivals, the Experiment and the Result under the new Goal. I then make the select, link `G-NSXQ3E` to `gh:113`, run `hyp check`, and write the end-of-grill checkpoint.

## Owner, message 10

Approved. Run that exact command under `owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "sed -i '' 's#design-body.txt#design-body.txt#' $TMPDIR/g/add-design-goal.sh; zsh $TMPDIR/g/add-design-goal.sh 2>&1 | tail -5; hyp status 2>&1 | head -8",
 "description": "Run the approved design-record Goal command"
}
```

Result:

```text
G-9WATD9 node.add
GOALS (hyp status G-… prints one Goal)
  G-NSXQ3E hyp prints the /goal prompt for any Goal from the log, read… — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 1, stale 0); runner owes nothing; checkpoint none
  G-9WATD9 How hyp prompt finds a Goal body's constraints block (desig… — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint none
AWAITING OWNER
```

## Grill

`G-9WATD9` is written. Now I re-add the DesignQuestion, rivals, Experiment and Result under it, then the select.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-9WATD9\nhyp add dq --parent $G --alias q-block --title \"How does hyp prompt find a Goal body's constraints block?\" --body \"Scenario that separates the rivals: a body with a prose line that starts with Never:, more prose, then a real block of May change:, Read only: and Never: lines. Decided by argument in the grill session; the build Goal G-NSXQ3E builds the winner.\" --why \"A testable option set settled in the grill is still a DesignQuestion with rivals.\" 2>&1 | tail -1\nhyp add hyp --parent q-block --alias h-first-heading --title \"The block runs from the first heading line to the end of the body\" --body \"From the first column-0 line starting May change:, Read only: or Never: to the end of the body. Every non-empty line in it starts with a heading and all three headings appear. The scenario body is refused with exit 1 and the first failing line is named.\" --why \"Rival A, chosen in the grill: it matches the rule PR #135 wrote in hyp-owner-grill.\" 2>&1 | tail -1\nhyp add hyp --parent q-block --alias h-trailing-run --title \"The block is the trailing run of heading lines\" --body \"The block is the last run of consecutive heading lines at the end of the body; prose before it is ignored. The scenario body is accepted and the real block is printed.\" --why \"Rival B, rejected in the grill: a heading-like line in prose would pass unnoticed and the first failing line could not be named.\" 2>&1 | tail -1\nhyp add exp --parent h-first-heading --recipe none --rubric reading --target h-trailing-run --alias e-block-scenarios --title \"Scenarios that separate the two block rules, argued in the grill\" --body \"Rule, fixed before the Result: S1 = a body with a prose line starting Never:, a line of prose, then a block of May change:, Read only: and Never: lines. A refuses S1 (exit 1) and names the first failing line, which is the prose line after the first Never:. B accepts S1 and prints the trailing block. S2 = a body that is only a valid block: both accept. S3 = a block followed by a prose line: A refuses, B accepts. The owner chose A in the session because a heading-like line in prose must not pass unnoticed. Verdicts: A supports on S1 and S3; B contradicts. Evidence: the scenarios file.\" --why \"Records the rival choice as argued, not tested.\" 2>&1 | tail -1",
 "description": "Re-add the DQ, rivals and Experiment under the design Goal"
}
```

Result:

```text
refused: alias already used: "q-block"
refused: alias already used: "h-first-heading"
refused: alias already used: "h-trailing-run"
refused: alias already used: "e-block-scenarios"
```

## Grill

The old aliases still belong to the withdrawn nodes, so I'm using fresh ones and passing parents by id.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-9WATD9\nQ=$(hyp add dq --parent $G --alias dq-block-find --title \"How does hyp prompt find a Goal body's constraints block?\" --body \"Scenario that separates the rivals: a body with a prose line that starts with Never:, more prose, then a real block of May change:, Read only: and Never: lines. Decided by argument in the grill session; the build Goal G-NSXQ3E builds the winner.\" --why \"A testable option set settled in the grill is still a DesignQuestion with rivals.\" 2>&1 | awk '{print $1}'); echo Q=$Q\nA=$(hyp add hyp --parent $Q --alias hyp-first-heading --title \"The block runs from the first heading line to the end of the body\" --body \"From the first column-0 line starting May change:, Read only: or Never: to the end of the body. Every non-empty line in it starts with a heading and all three headings appear. The scenario body is refused with exit 1 and the first failing line is named.\" --why \"Rival A, chosen in the grill: it matches the rule PR #135 wrote in hyp-owner-grill.\" 2>&1 | awk '{print $1}'); echo A=$A\nB=$(hyp add hyp --parent $Q --alias hyp-trailing-run --title \"The block is the trailing run of heading lines\" --body \"The block is the last run of consecutive heading lines at the end of the body; prose before it is ignored. The scenario body is accepted and the real block is printed.\" --why \"Rival B, rejected in the grill: a heading-like line in prose would pass unnoticed and the first failing line could not be named.\" 2>&1 | awk '{print $1}'); echo B=$B\nE=$(hyp add exp --parent $A --recipe none --rubric reading --target $B --alias exp-block-scenarios --title \"Scenarios that separate the two block rules, argued in the grill\" --body \"Rule, fixed before the Result: S1 = a body with a prose line starting Never:, a line of prose, then a block of May change:, Read only: and Never: lines. A refuses S1 (exit 1) and names the first failing line, which is the prose line after the first Never:. B accepts S1 and prints the trailing block. S2 = a body that is only a valid block: both accept. S3 = a block followed by a prose line: A refuses, B accepts. The owner chose A in the session because a heading-like line in prose must not pass unnoticed. Verdicts: A supports on S1 and S3; B contradicts. Evidence: the scenarios file.\" --why \"Records the rival choice as argued, not tested.\" 2>&1 | awk '{print $1}'); echo E=$E\necho \"$Q $A $B $E\" > $TMPDIR/g/ids.txt",
 "description": "Re-add nodes with fresh aliases"
}
```

Result:

```text
Q=Q-XYXE51
A=H-4QWKFQ
B=H-H0T0RT
E=E-2KQ9JZ
```

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && mkdir -p .tree/evidence/E-2KQ9JZ && cat > .tree/evidence/E-2KQ9JZ/scenarios.txt <<'EOF'\nScenarios for Q-XYXE51 (how hyp prompt finds the constraints block). Argued in the #113 grill session, not run.\n\nS1 body:\n  Context prose.\n  Never: this line is prose, not part of a block.\n  More prose.\n  May change: x\n  Read only: y\n  Never: z\nRule A (first heading line to end): refuses; the first failing line is \"More prose.\" (line 3).\nRule B (trailing run of heading lines): accepts and prints lines 4 to 6.\n\nS2 body: a valid block and nothing else. Rule A: accepts. Rule B: accepts.\n\nS3 body: a valid block, then \"Closing prose.\" Rule A: refuses at the last line. Rule B: accepts the block.\n\nverdict: H-4QWKFQ supports\nverdict: H-H0T0RT contradicts\nEOF\nhyp record --exp E-2KQ9JZ --verdict H-4QWKFQ=supports --verdict H-H0T0RT=contradicts --observation \"Records the grill's argued choice under the Experiment's pre-registered rule: the owner chose rule A in session; S1 and S3 separate the rivals and rule B fails the refusal case.\" --evidence .tree/evidence/E-2KQ9JZ/scenarios.txt --why \"Result of the argued DesignQuestion.\" 2>&1 | tail -4\nhyp status G-9WATD9 2>&1 | sed -n '/OWED/,/AWAITING/p'; hyp select Q-XYXE51 2>&1 | head -3",
 "description": "Record the Result and check select suggestion"
}
```

Result:

```text
R-71E4JK result.record
OWED
  runner: Q-XYXE51 select H-4QWKFQ (hyp-first-heading), the only candidate
AWAITING OWNER
suggested: H-4QWKFQ
```

## Grill

The select is now owed. I make it as a runner write, then link, check and checkpoint.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp select Q-XYXE51 H-4QWKFQ --why \"OWED owes this select: R-71E4JK supports the first-heading rule and contradicts the trailing-run rule; the owner chose it in the grill.\" 2>&1 | tail -2\nhyp link G-NSXQ3E external_ref gh:113 --why \"The build Goal serves issue #113.\" 2>&1 | tail -1\nhyp link G-9WATD9 depends_on G-NSXQ3E --why \"x\" 2>&1 | tail -1\nhyp link G-9WATD9 external_ref gh:113 --why \"The design record serves issue #113.\" 2>&1 | tail -1\nhyp status 2>&1 | head -5; hyp check 2>&1 | tail -3",
 "description": "Select, link issue, check"
}
```

Result:

```text
Q-XYXE51 dq.select
G-NSXQ3E edge.add
refused: wrong kind: depends_on from Goal
G-9WATD9 edge.add
GOALS (hyp status G-… prints one Goal)
  G-NSXQ3E hyp prints the /goal prompt for any Goal from the log, read… — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 1, stale 0); runner owes nothing; checkpoint none
  G-9WATD9 How hyp prompt finds a Goal body's constraints block (desig… — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint none
AWAITING OWNER
check: 0 errors, 0 warnings
```

## Grill

Both Goals are consistent. The refused `depends_on` was a probe I didn't need, since a Goal can't depend on a Goal, so it changed nothing. Now the end-of-grill checkpoints.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp checkpoint --goal G-NSXQ3E --summary \"End of grill for #113. Eight lines, written by owner-standin. The block-rule question is decided in G-9WATD9 (select of H-4QWKFQ): the block runs from the first heading line to the end of the body. The run-6 match is not here: it waits for the owner to supply run 6's prompt text and becomes a later Goal. The built hyp prompt is meant to reproduce the hand-written prompt below.\" --why \"end of grill\" 2>&1 | tail -1\nhyp checkpoint --goal G-9WATD9 --summary \"End of grill for #113. The block rule is selected: H-4QWKFQ, first heading line to the end of the body, decided by argument in the grill session. This Goal has no lines. The build is G-NSXQ3E.\" --why \"end of grill\" 2>&1 | tail -1\nhyp check 2>&1 | tail -2; cat .gitignore; git status --short",
 "description": "Write end-of-grill checkpoints and check status"
}
```

Result:

```text
G-NSXQ3E checkpoint
G-9WATD9 checkpoint
check: 0 errors, 0 warnings
.claude/
/hyp
*.sqlite
*.sqlite-*
*.pyc
.*-scratch
/plans/
.agentkit/
?? .tree/
```

## Grill

The skill says to commit `.tree/events/` and `.tree/evidence/`. I commit them on the current branch `work`, using the conventional format and no AI trailer, per AGENTS.md.

### Tool Bash (cwd: .)

```text
{
 "command": "git add .tree && git commit -q -m \"docs: grill #113 into a build Goal for hyp prompt and a design record for its block rule\" && git log --oneline | head -2; git status --short; find .tree -type f | head",
 "description": "Commit the store"
}
```

Result:

```text
4ff1dba docs: grill #113 into a build Goal for hyp prompt and a design record for its block rule
a51f140 Merge pull request #137 from dat9uy/docs/discovery-and-build-goals
.tree/evidence/E-2KQ9JZ/scenarios.txt
.tree/evidence/E-9B59GK/scenarios.txt
.tree/.gitignore
.tree/events/claude-grill-113.jsonl
.tree/events/owner-standin.jsonl
```

## Grill

The grill is done: both Goals are committed as `4ff1dba` on `work`, and `hyp check` reports 0 errors.

## Hand-off

The installed `hyp` has no `prompt`, so this prompt is **hand-written** to the repo's `skills/hyp-owner-grill` rule. Once the build exists, `hyp prompt G-NSXQ3E --actor claude-build-113 --branch feat/goal-prompt` should reproduce it. The branch name and runner actor are my picks; change them if you like.

```text
/goal Run the hyp tree in /Users/datguy/.hyp-sandbox/g146-base-a-1/wt (store: .tree at the repository root) with the hyp-goal-run skill, as HYP_ACTOR=claude-build-113. Its Goal is G-NSXQ3E. The block-rule question is decided in G-9WATD9 (select of H-4QWKFQ); read it, do not change it.
Done when the latest printed `hyp status G-NSXQ3E` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Experiments under G-NSXQ3E have been added by claude-build-113 and not withdrawn, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names with --goal G-NSXQ3E, and list the lines under LINES with the Results that ran them.
Commit on the local branch feat/goal-prompt. Open a PR when a candidate commit is ready, and never merge it.
The write scope covers every text the change makes wrong, the design record's neighbours in docs/ included.
After each candidate commit, run one review round by a fresh subagent: give it the eight lines, the code and the keys already settled with the owner's words. Sort each finding: a case of a line is a limit; a question about a line's meaning is a reading; a defect in text you may edit is a fix commit; anything else is an out-of-Goal finding, named in the stop checkpoint, and it never holds the select.
May change: this repository, as commits on one local branch, in cmd/, internal/, skills/ (the hyp-owner-grill hand-off, the descriptions of hyp-goal-run and hyp-prototype), docs/, README.md and AGENTS.md step 3, where it says where trees live; a PR may be opened and is never merged
Read only: investigations/ and recipes/, and every store under them
Never: replace the installed hyp binary
Never: edit another tree, a recorded event or this Goal
Never: add a Goal or a line
Never: write under the owner's actor
```

## What waits on you

- **Run 6's prompt text.** It isn't in the repo and can't be generated. You supply it, and then you add the later Goal that carries the baseline comparison by sha256.
- **The placeholder issue.** AGENTS.md says an issue that closes with its goal unmet gets a new issue first. #113's Done-means item 1 (the run-6 match) will be unmet when `G-NSXQ3E` completes. This machine has no network, so I could not open that issue. Please open it, stating that unmet goal, before #113 closes.
- **AGENTS.md step 3.** It still says trees live under `investigations/<topic>/`. The run may change it, as the Goal's May change line says.
- **Evidence file `E-9B59GK/scenarios.txt`.** It belongs to the withdrawn first attempt and is harmless. It stays in the log.

Store state: `G-NSXQ3E` is in progress and the runner owes nothing yet. `G-9WATD9` is complete, with `Q-XYXE51` answered.
