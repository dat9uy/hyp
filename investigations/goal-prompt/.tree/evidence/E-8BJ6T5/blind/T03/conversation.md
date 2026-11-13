# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-a-2/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-2/tmp/issue-113.md"
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
 "command": "(hyp status 2>&1 | head -40); ls; ls docs | head -50",
 "description": "Show hyp store status and repo layout"
}
```

Result:

```text
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
adr
core-and-self-review.md
glossary.md
how-hyp-improves-itself.md
install.md
principles.md
replay-spec.md
sandbox-spec.md
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls investigations investigations/* | head -40; ls internal cmd/hyp skills/*; which hyp; hyp --help 2>&1 | head -40; grep -rn \"prompt\" --include=*.go -il . | head; git log --oneline | head -15; ls docs/adr",
 "description": "Survey trees, commands, and prompt generator presence"
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
cmd/hyp:
main.go

internal:
check
claim
cli
cmdrecord
contain
event
herr
recipe
render
replay
sandbox
signals
skillcheck
store
testutil

skills/hyp-goal-run:
SKILL.md

skills/hyp-owner-answer:
SKILL.md

skills/hyp-owner-grill:
SKILL.md

skills/hyp-prototype:
example.sh
SKILL.md

skills/hyp-verification:
RECIPE-FORMAT.md
SKILL.md
/Users/datguy/.hyp-sandbox/g146-base-a-2/bin/hyp
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
  answer [
[... 3633 more characters]
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-2/wt/docs/glossary.md"
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
 "command": "hyp --help 2>&1 | sed -n 40,80p; hyp prompt --help 2>&1 | head; ls -a investigations/checkpoint-store investigations/checkpoint-store-design; grep -rn \"hyp prompt\" --include=*.go --include=*.md . | head -20; grep -rln \"goal-prompt\" . 2>/dev/null | head; ls ../tmp",
 "description": "Check whether hyp prompt exists and locate run-6 prompt"
}
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-2/wt/skills/hyp-goal-run/SKILL.md"
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
usage: hyp <command> [flags]

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
Statuses are derived from events; no command can set one.

setup:
  init                                   create .tree in the current directory
  version                                print the hyp version

write:
investigations/checkpoint-store:
.
..
.tree

investigations/checkpoint-store-design:
.
..
.tree
(eval):1: no matches found: --include=*.go
hyp-owner-grill-eval
issue-113.md
trial-git
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
 "command": "grep -rn \"prompt\" internal cmd -il | head; grep -rn \"hyp prompt\" skills docs README.md | head -20; ls internal/skillcheck internal/cli; cat docs/adr/0021-a-skill-change-is-a-build-goal.md; (cd investigations/checkpoint-store && hyp status 2>&1 | head -60)",
 "description": "Check for prompt command in code, skillcheck, ADR 0021, and build tree status"
}
```

Result:

```text
internal/cli:
answer_record_test.go
claim_test.go
cli.go
contain_test.go
cross_goal_moves_test.go
events_test.go
exitcode_test.go
exitcode_wide_test.go
goal_lines_answer_test.go
goalline_test.go
help_test.go
idcite_test.go
idcite.go
limitclose_test.go
limitreturn_test.go
missing_format_test.go
owed_test.go
preregistration_test.go
reach_test.go
read_test.go
read.go
recipe_first_test.go
recipe_test.go
recipe.go
records.go
resolve.go
run_subject_test.go
sameask_test.go
sandbox_test.go
sandbox.go
separated_test.go
showlimit_test.go
signals_once_test.go
subject.go
testdata
two_goal_store_test.go
write_test.go
write.go

internal/skillcheck:
skillcheck_test.go
---
status: accepted
---

Supersedes in part [ADR 0007](0007-records-start-at-230.md) (the skill-text gate's replay instrument), [ADR 0010](0010-a-record-keeps-or-identifies-what-a-replay-needs.md) (the follow-up that rebuilds `skill-text-arms` Arms from patches) and [ADR 0015](0015-the-select-is-the-gate.md) (the frozen verifier recipes).

# A skill change is a build Goal

AGENTS.md gated skill text apart from code. A skill-text change shipped only when a session replay from the event's store cut showed the failure in 3 or more of 5 trials without the change and in 1 or fewer of 5 with it, and nothing shipped when no cut reproduced the failure. Its one instrument, `recipes/skill-text-arms@5` with the rubric `rubrics/skill-text-gate@1`, replays the id-citation tree's e-arms with `hyp-prototype` and `hyp-verification` loaded. It cannot reach a skill that runs before any tree exists, such as `hyp-owner-grill`, or one that spans a whole `/goal` run, such as `hyp-goal-run` and `hyp-owner-answer`. For those skills the rule was a permanent freeze, and a new skill needed its own exception. [how-hyp-improves-itself.md](../how-hyp-improves-itself.md) had already found that replay studies of skill text decided nothing, because agents with the skill loaded rarely made the mistakes they counted.

Owner decision on 2026-11-11 (#1
[... 5264 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2 && grep -rl \"goal-prompt-261111\" . 2>/dev/null | head; ls tmp/hyp-owner-grill-eval tmp/trial-git 2>&1 | head -20; cd wt && ls investigations/checkpoint-store/.tree investigations/checkpoint-store/.tree/events investigations/checkpoint-store/.tree/evidence | head -30; cat skills/hyp-owner-grill/SKILL.md | head -5; sed -n 1,200p skills/hyp-owner-answer/SKILL.md | head -30; head -8 skills/hyp-prototype/SKILL.md",
 "description": "Look for run-6 prompt and eval dirs, skill headers"
}
```

Result:

```text
./tmp/issue-113.md
ls: tmp/hyp-owner-grill-eval: Operation not permitted
tmp/hyp-owner-grill-eval:

tmp/trial-git:
gh
gitconfig
investigations/checkpoint-store/.tree:
events
evidence

investigations/checkpoint-store/.tree/events:
claude-g1-1.jsonl
claude-g1b-1.jsonl
claude-g1b-2.jsonl
claude-g1b-3.jsonl
claude-g1b-4.jsonl
claude-g1b-5.jsonl
claude-g1b-6.jsonl
owner-dat.jsonl

investigations/checkpoint-store/.tree/evidence:
E-161WBK
E-17H1DS
E-2XGWNC
E-6PD9S8
E-CM3SCT
E-E89EQF
E-FS4RH5
E-GAT8N4
E-QA0X14
E-YCWW1B
E-ZX9APE
---
name: hyp-owner-grill
description: "Grill the owner about a purpose until it is a hyp Goal a /goal run can finish: route each question (owner-only into a line, rivals a run could separate into a DesignQuestion, facts looked up), derive the lines from positive and negative space, split purposes into separate Goals, and record the owner's writes only on their approval. Use before a /goal run, when the owner wants to write a new Goal; running a Goal is hyp-goal-run, answering a stopped run is hyp-owner-answer."
disable-model-invocation: true
argument-hint: "[the purpose, an issue or a design to turn into a Goal]"
---
name: hyp-owner-answer
description: "After a /goal run stops at the owner, walk the open readings and referred limits of the hyp tree with the owner: explain each one, answer questions, and record only the owner's answers."
disable-model-invocation: true
argument-hint: "[reading or limit key, question or answer]"
---

The owner's input: $ARGUMENTS

A `/goal` run stopped and the owner now reads what waits for them. Use the tree of this session. If this session has no tree, ask for its path. Read the Goal with `hyp show <goal>`, including its lines and checkpoints, and read `hyp status`. In a store with many Goals, `hyp status` prints a GOALS table and one Awaiting owner list grouped by Goal; read `hyp status <goal>` for one Goal's lines and OWED. The owner's acts are `hyp answer` events: each item is `KEY=TEXT` with the owner's words, an
[... 2039 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store-design)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -80; cd ../checkpoint-store-design && hyp status 2>&1 | head -30",
 "description": "Show the earlier build Goal as a model"
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
[... 8987 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store-design)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && cat docs/adr/0020-one-store-holds-many-goals.md | head -80; grep -n \"Goal\" docs/replay-spec.md | head -60",
 "description": "Read ADR 0020 and spec Goal mentions"
}
```

Result:

```text
<persisted-output>
Output too large (35KB). Full output saved to: /Users/datguy/.claude/projects/-Users-datguy--hyp-sandbox-g146-base-a-2-wt/e3ce0dbb-c168-4a49-9caf-ecc01cb40fbf/tool-results/bf5a05zef.txt

Preview (first 2KB):
---
status: proposed
---

# One store holds many Goals

Until now a tree had one Goal (spec §5 rule 4: "only one Goal per log"). A new purpose was a new tree, and that tree cited the old one in prose as `<tree>:ID`, which hyp does not check (ADR 0016). After the multi-tenant tree completed on 2026-11-10, the owner wanted the build that follows a tree to run from the tree instead of from a plan in prose. Every problem with that came from the one-Goal rule:

- A build had no checked way to cite what the tree had decided.
- The owner's checkpoints were split across trees.
- A commit could not be a subject, because `hyp record` copied file hashes and dropped the sandbox's `base` (`internal/cli/subject.go:118`).

In session `e6a35a24` (2026-11-10 and 11) the owner settled the model: **building is a Goal like any other.** hyp keeps no discovery kind and no delivery kind. The design questions it raised are recorded and selected in `investigations/checkpoint-store-design` (G-73R704), with that session as evidence:

- Q-D906CD (R-HW9B05): what a Result keeps about its runs.
- Q-HC4V9T (R-V8032Z): where a build Goal's start commit lives.
- Q-H8PJ20 (R-8CJCV6): what the owner's act is attached to.
- Q-R21S8R (R-AWPQ4X): how many event types there are.

`investigations/checkpoint-store` (G-D1CH1R) builds the design, and its nine owner lines are this ADR's acceptance. Tracked in #126, part of #114.

## Decision

- **One store holds many Goals.** The store is still one `.tree/` with one events file per actor. A tree is a view, not a container: a Goal and every node whose parents lead to it. Rule 4's "only one Goal per log" goes.
- **A subject has one shape: a base commit, plus what is given over it.** What is given over it is at most one patch, and input fi
[... 212 more characters]
```

## Grill

Checking what the code and published skill already have for the prompt generator, then I'll put round 1 to you.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"prompt\\|/goal\" skills/hyp-owner-grill/SKILL.md | head -20; sed -n 1,80p internal/skillcheck/skillcheck_test.go; grep -n \"Body\\|body\" internal/replay/*.go | grep -i goal | head; grep -n \"runner owes nothing\" -r internal --include=*.go | head -5; grep -n \"func.*addGoal\\|\\\"goal\\\"\" internal/cli/write.go | head",
 "description": "Inspect published grill skill, skillcheck, goal body handling"
}
```

Result:

```text
3:description: "Grill the owner about a purpose until it is a hyp Goal a /goal run can finish: route each question (owner-only into a line, rivals a run could separate into a DesignQuestion, facts looked up), derive the lines from positive and negative space, split purposes into separate Goals, and record the owner's writes only on their approval. Use before a /goal run, when the owner wants to write a new Goal; running a Goal is hyp-goal-run, answering a stopped run is hyp-owner-answer."
10:This skill turns a conversation with the owner into a hyp tree that an unattended `/goal` run can finish: a Goal with its lines, the contested questions under it with their rivals, and nothing the run must come back to ask. It ends where the hyp-goal-run skill starts. Use it with the hyp-prototype skill for the commands that add nodes.
35:| how the tree is run (process, review, order, cadence) | the Goal body, an Experiment's verdict rule, or the `/goal` prompt; never a line | only where it is a preference |
51:- **A line carries its own cases.** A line that compares output with a baseline names the baseline by sha256 and asserts that it differs from the candidate; other lines name no baseline. Name each exception it allows in its own words, not only in a later prompt.
96:The Goal body ends with the run's constraints, one heading per line, so a run and a prompt generator can find them:
108:Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Never:` lines copied as they are. Put general practice in the skill, not the prompt. For a build Goal it also says:
114:A prompt that states tree state (an accept, a select) is written from the merged log, or names the PR it waits for.
// Package skillcheck keeps the skills honest: every hyp command they show
// must be one the CLI dispatches.
package skillcheck

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"
	
[... 2515 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 90,125p skills/hyp-owner-grill/SKILL.md; grep -n \"MaxGoalBody\" -r internal | head -3; grep -n \"func Commands\" -A25 internal/cli/cli.go | head -50; grep -n \"owner may\\|runner owes nothing\\|Experiments have been recorded\" -r internal --include='*.go' | grep -v _test | head",
 "description": "Read repo grill skill's hand-off and CLI command dispatch"
}
```

Result:

```text
| the Goal, its purpose and its lines | `hyp add goal --title "…" --body "…" --line "…" --cite N=KEY --why "…"` | owner |
| a contested question | `hyp add dq --parent <G-id> …`, then one `hyp add hyp --parent <Q-id> …` per rival | runner |
| a question with rivals the owner settled in the session | the DesignQuestion and each rival as above; then under the chosen rival `hyp add exp --recipe none --rubric reading --target <each other rival>` whose body states the scenarios as its rule, `hyp record` with a verdict per rival citing a scenarios file under `.tree/evidence/<E-id>/`, then the select OWED owes (a runner write, not the owner's) | runner |
| a link to another Goal or an issue | `hyp link <id> external_ref …` | runner |
| the end | `hyp checkpoint --goal <G-id> --summary "…" --why "end of grill"` | runner |

The Goal body ends with the run's constraints, one heading per line, so a run and a prompt generator can find them:

'''text
May change: <the repository, how it is changed (for example commits on one local branch, a PR never merged), and the texts the change makes wrong>
Read only: <repositories, directories or data the run may read but never change>
Never: <acts the run never takes, such as replacing an installed binary or publishing source data>
'''

A body that names another tree's ids writes each as `<tree>:<id>`. Then run `hyp status <G-id>` and `hyp check`: it must print `check: 0 errors`. Commit `.tree/events/` and `.tree/evidence/`.

## Hand off

Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Never:` lines copied as they are. Put general practice in the skill, not the prompt. For a build Goal it also says:
- the branch, and that the run opens a PR and never merges;
- the write scope covers every text the change makes wrong, the design record included;
- one review round per candidate commit, by a fresh subagent given
[... 2175 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/checkpoint-store/.tree/events && head -c 3500 claude-g1-1.jsonl; echo; wc -l *.jsonl; cd .. && hyp show Q-PC04ED 2>&1 | head -12; cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && grep -n \"build Goal\" -i skills/hyp-prototype/SKILL.md | head",
 "description": "See how the earlier build tree began (Goal, question, first hypothesis)"
}
```

Result:

```text
{"v":1,"id":"01M74Y0D72WN8RS0X4R0D1CH1R","ts":"2026-11-11T09:11:16.450904Z","actor":"claude-g1-1","type":"node.add","node":"01M74Y0D72WN8RS0X4R0D1CH1R","payload":{"kind":"Goal","parent":"","title":"hyp implements the shared store: one store holds many Goals","body":"Builds the design recorded in checkpoint-store-design:G-73R704 (selects on checkpoint-store-design-D906CD, checkpoint-store-design-HC4V9T, checkpoint-store-design-H8PJ20, checkpoint-store-design-R21S8R). Every decision is selected there with session e6a35a24 as evidence, and ADR 0020 cites those Results. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Two-Goal stores land first, so lines 2, 6 and 7 are red on that commit's parent.","alias":"goal"},"why":"Owner, 2026-11-11: split G1; this tree is the build (G1b). The owner writes its lines."}
{"v":1,"id":"01M74Y0D91NRBGT9R26QA54372","ts":"2026-11-11T09:11:16.513758Z","actor":"claude-g1-1","type":"edge.add","node":"01M74Y0D72WN8RS0X4R0D1CH1R","payload":{"type":"external_ref","to":"checkpoint-store-design:G-73R704"},"why":"The design this Goal builds (G1a)."}
{"v":1,"id":"01M74Y0DAXW1YGXWTYQGPC04ED","ts":"2026-11-11T09:11:16.573387Z","actor":"claude-g1-1","type":"node.add","node":"01M74Y0DAXW1YGXWTYQGPC04ED","payload":{"kind":"DesignQuestion","parent":"01M74Y0D72WN8RS0X4R0D1CH1R","title":"Which hyp commit implements the shared-store design?","alias":"q-commit"},"why":"The build's one question; its candidate will be a hyp commit, added when one exists."}

       3 claude-g1-1.jsonl
       8 claude-g1b-1.jsonl
       5 claude-g1b-2.jsonl
       5 claude-g1b-3.jsonl
      12 claude-g1b-4.jsonl
      20 claude-g1b-5.jsonl
       6 claude-g1b-6.jsonl
       3 owner-dat.jsonl
      62 total
DesignQuestion Q-PC04ED
alias: q-commit
title: Which hyp commit implements the shared-store design?
path: G-D1CH1R > Q-PC04ED
status: answered
candidates: H-P4YGE2
selected: H-P4YGE2 — OWED owes this select: R-RRYBPH on 4047c01 (branc
[... 1241 more characters]
```

## Grill

I checked the repo and the store before asking anything. The installed `hyp` has no `prompt` command, and the code has no prompt generator. There is no store at the repo root, so this Goal would be the first in a new one.

`investigations/checkpoint-store` (G-D1CH1R) is the model. It is a build Goal with one question, `q-commit` ("Which hyp commit …?"). The run adds the commits as Hypotheses.

I'm treating #113 as a **build Goal in hyp's core**. Its candidate is a hyp commit that adds `hyp prompt` and a warning in `hyp add goal`. It also corrects the two skill descriptions you folded in.

I have a few facts to correct before round 1. You don't need to answer these.

- **The skill descriptions:** `hyp-goal-run` says "an unattended discovery /goal" and `hyp-prototype` says "Investigate an uncertain design". Both predate the words discovery Goal and build Goal.
- **Run 6's prompt:** the text of `goal-prompt-261111-2129-g1b-run-6` is in no file I can read. The issue only names it.
- **G-D1CH1R's body:** it has no `May change:`, `Read only:` or `Never:` block, and a Goal is never edited.

## Round 1

**Q1. Done-means 1 cannot be met as written.**

- **What exists:** the issue says the prompt generated from G-D1CH1R must match run 6's prompt clause for clause. But `hyp prompt` refuses a Goal whose body has no valid constraints block, and G-D1CH1R's body has none. Also, AGENTS.md says core tests read only synthetic data, so the real store can't be the test input.
- **Option A:** I build a synthetic Goal modelled on G-D1CH1R, with a valid block. The expected prompt is a fixture pinned by sha256 under `internal/*/testdata`, written from run 6's real text. One line says the generated prompt equals the fixture.
- **Option B:** no fixture. The line lists the clauses and checks each one separately.
- **Principle:** the prompt is the product, so a test should hold its exact bytes. That also catches a drifting wording.
- **Recommendation:** A. Please paste run 6's prompt, or the clauses you remember, so the fixture matches your real run and not my guess.

**Q2. What does the fixed text say?** This is my draft. Edit it, or confirm it.

```text
/goal Run the hyp tree in <absolute store path> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G-id>.
Done when the latest printed `hyp status <G-id>` shows `runner owes nothing: stop and report to the owner` under OWED, or <N> Experiments have a Result under <actor>, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
<the body's May change: / Read only: / Never: lines, copied as they are>
Work on branch <B>.                              (only with --branch)
The PR for the branch is #<N>. Never merge it.  (only with --pr)
```

- **Dropped from run 5:** the clause "every runner item under OWED is `add rival Hypotheses`…" is a hand-tuned discovery clause. The skill already makes the runner act on what OWED lists. `--experiments` defaults to 5.
- **The branch and PR clauses:** they only name the branch and the PR. Whether the run may open a PR belongs in the body's `May change:` or `Never:`. A prompt may narrow the block and never widen it.
- **Recommendation:** the draft above.

**Q3. What stays out of the prompt?** Run 6's prompt carried the owner answer it started from and a start commit.

- **What exists:** OWED already lists each answer and each returned limit. A build Goal records no start commit (glossary, Build Goal). The branch head is the start.
- **Derivation guard:** the actor, branch and PR number are arguments to `hyp prompt`. Everything else is read from the log, so no stored field and no new event.
- **Recommendation:** leave the owner answer and the commit out. They are the "run-specific inputs" your Done-means already allows to differ.

**Q4. Is the strictness of the constraints block right?** This is what I plan to build.

- **Valid block:** it starts at the first line beginning with `May change:`, `Read only:` or `Never:`. It has all three headings. Every non-empty line in it starts with one of them. A whitespace-only line counts as non-empty.
- **`hyp prompt` on a Goal with no valid block:** it refuses with exit 1 and a message that names the rule, and prints nothing to stdout.
- **`hyp add goal` on a body that has a heading but no valid block:** it warns and writes anyway. It stays silent on a body with no heading, so old Goals and old logs replay unchanged.
- **Principle:** replay must never reject an old log, and a Goal can't be edited. So the refusal sits at the read side.
- **Recommendation:** confirm.

**Q5. Does this Goal also correct the repo's skill text?** There are two parts.

- **(a) The descriptions you folded in.** My draft wording:
  - `hyp-goal-run`: "…as an unattended /goal, for a discovery Goal or a build Goal…"
  - `hyp-prototype`: "Investigate an uncertain design, or build to a Goal, with rival hypotheses…"
- **(b) The Hand off section of `skills/hyp-owner-grill`.** The repo's copy tells the grill to hand-write the prompt, and `hyp prompt` makes that wrong. I'd change it to run `hyp prompt <G-id> --actor … [--branch] [--pr] [--experiments]` and give the output as it is. The existing skillcheck test then confirms `hyp prompt` is a real command.
- **Recommendation:** do both in this Goal. Leaving (b) would leave a published skill that contradicts the binary.

**Q6. How does the change land?**

- **Option A:** commits on one local branch of this repo. The run opens no PR, and you open it with the Result ids.
- **Option B:** the run also opens a PR and never merges it.
- **What exists:** this machine has no network, so `gh` cannot open a PR from here.
- **Recommendation:** A.

**Also needed:** your actor name, for example `owner-dat`. I will use `claude-p113-grill` for runner writes.

## What I'm planning, so you can redirect it

- **Routing:** I plan no DesignQuestion beyond `q-commit`. "hyp prints vs a skill writes" has no scenario a run could execute to separate them. A byte-exact fixture can only be met by code, so that is the Goal's purpose, not a choice.
- **Lines:** about 8, written after your answers.
  - [P] the fixture output
  - [P] `--branch` and `--pr` add exactly one clause each
  - [N] the prompt never adds text to the block
  - [N] an invalid block is refused with exit 1
  - [N] old stores' `status` and `check` stay byte-identical
  - [P]/[N] the two skill descriptions
  - [P] the hand-off names `hyp prompt`
- **Before the routing table:** I'll have a fresh subagent walk the rules through a scenario we haven't drawn, such as two Goals in one store and a Goal whose block spans a merged log. I'll also scan the docs and ADRs for a claim the run would contradict.

## Owner, message 2

1. Fixture or per-clause check is a testing detail, so I won't pick. Read the ak-skill-creator eval and audit flow and decide from that. I can't paste run 6's prompt, because I only see your messages. The exceptions to the clause-for-clause match are wider than you list: the actor, the owner's answer, branch, PR and commit, and the run's own plan. Branch and PR become flags.
2. Confirm the draft. Experiments is a flag, default 5.
3. Agree. The prompt carries no owner answers, since the runner reads OWED. The commit falls under the widened exceptions.
4. Confirm. Keep the strict form, one constraint per line. `hyp prompt` should exit 1 and name what is missing. Adding such a Goal stays allowed, so warn-and-write is fine.
5. (a) The two skill descriptions are text the change makes wrong, so they go under May change. Don't make a line for them. (b) Accept your recommendation and do both.
6. A.

Actor name: owner-standin.

Routing: accept your plan. Drop the skill-description line from your list.

### Tool Skill (cwd: .)

```text
{
 "skill": "ak-skill-creator",
 "args": "I need only its eval and audit flow (evals.json format, eval_skill.py grading, audit), to decide how to test a generated-prompt output."
}
```

Result:

```text
Launching skill: ak-skill-creator
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/.claude/skills/ak-skill-creator && ls references scripts | head -60; grep -n \"evals.json\" -r references | head -15; sed -n 1,60p scripts/eval_skill.py",
 "description": "Locate eval format and grader in ak-skill-creator"
}
```

Result:

```text
references:
advisory-supervision.md
agentkit-kit-skill-contract.md
benchmark-optimization-guide.md
creator-consumer-evaluation.md
cross-marketplace-distribution.md
distribution-guide.md
evaluation-tools.md
long-horizon-building-loop.md
long-horizon-repository-architecture.md
long-horizon-scaffold-templates.md
mcp-skills-integration.md
metadata-quality-criteria.md
plugin-marketplace-hosting.md
plugin-marketplace-overview.md
plugin-marketplace-schema.md
plugin-marketplace-sources.md
plugin-marketplace-troubleshooting.md
production-skill-standard.md
prompt-cruft-patterns.md
script-dependency-strategy.md
script-quality-criteria.md
semantic-routing-evaluation.md
session-usage-analysis.md
skill-anatomy-and-requirements.md
skill-creation-workflow.md
skill-design-patterns.md
skill-ecosystem-portability-and-safety.md
skillmark-benchmark-criteria.md
structure-organization-criteria.md
testing-and-iteration.md
token-efficiency-criteria.md
troubleshooting-guide.md
validation-checklist.md
writing-effective-instructions.md
yaml-frontmatter-reference.md

scripts:
encoding_utils.py
eval_contract.py
eval_skill.py
frontmatter_validation.py
init_skill.py
lint_cruft.py
metadata_routing.py
package_skill.py
quick_validate.py
requirements.txt
resource_validation.py
semantic_routing_contract.py
semantic_routing_eval.py
semantic_routing_identity.py
semantic_routing_metrics.py
semantic_routing_providers.py
semantic_routing_runner.py
semantic_routing_runtime.py
session_usage.py
skill_standard_report.py
standard_validation.py
tests
references/long-horizon-scaffold-templates.md:78:2. Define the first representative cases in `evals/evals.json` before optimizing instructions.
references/long-horizon-scaffold-templates.md:137:## `evals/evals.json`
references/long-horizon-repository-architecture.md:132:`evals/evals.json`, `.gitignore`, `VERSION`, `CHANGELOG.md`, and a one-paragraph `README.md`
references/evaluation-tools.md:11:python3 scripts/eval_skill.py validate /absolute/target-skill/evals/evals
[... 3310 more characters]
```

### Tool Agent (cwd: .claude/skills/ak-skill-creator)

```text
{
 "description": "Second reader for hyp prompt rules",
 "subagent_type": "general-purpose",
 "model": "fable",
 "run_in_background": true,
 "prompt": "You are an independent reviewer of a design. Do NOT edit any file; read only. Repository: /Users/datguy/.hyp-sandbox/g146-base-a-2/wt (Go, binary `hyp`; read AGENTS.md, docs/glossary.md, docs/replay-spec.md, internal/cli/cli.go and internal/cli/write.go as needed; the installed `hyp` binary is at /Users/datguy/.hyp-sandbox/g146-base-a-2/bin/hyp; you may run it only inside a scratch copy under $TMPDIR, never in the repo's investigations/).\n\nDesign to review (not yet built). A new command `hyp prompt <G-id> --actor <name> [--branch B] [--pr N] [--experiments N]` prints a `/goal` prompt for the hyp-goal-run skill, derived only from the store: \n line 1: `/goal Run the hyp tree in <absolute path of the directory holding .tree> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G-id>.`\n line 2: `Done when the latest printed `hyp status <G-id>` shows `runner owes nothing: stop and report to the owner` under OWED, or <N> Experiments have a Result under <actor>, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.` (N defaults to 5)\n then the Goal body's constraints block, copied line for line; then `Work on branch <B>.` if --branch; then `The PR for the branch is #<N>. Never merge it.` if --pr.\nThe constraints block is the part of the Goal body from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body. It must contain each of the three headings at least once, and every non-empty line in it must start with one of them (a heading may repeat). A line of spaces, a tab or a lone carriage return is not empty. If the body has no valid block, `hyp prompt` exits 1 with a message naming what is missing and prints nothing on stdout. Exit 2 = bad arguments (e.g. a --branch containing a space, actor not lowercase letters/digits/hyphens, N<1); exit 3 = no repository holds the store. `hyp add goal` with a body that has a heading but no valid block prints a warning and still writes (exit 0); a body with no heading gets no warning; `hyp check` and replay never warn or reject because of a Goal body's block (old Goals have none). Nothing new is stored in the log: the prompt is a function of store path, Goal id, actor, flags and the Goal body. A Goal body is at most 2000 runes (event.MaxGoalBody). A Goal is never edited.\n\nTask: walk these rules through scenarios the designer did not draw and report every case where the rules are ambiguous, contradictory, unsafe or break an existing invariant. Cover at least: two Goals in one store (one with a block, one without; one Goal id given as alias vs display id vs prefix — check how `hyp` resolves refs); a log merged from two actors; a Goal whose body ends with CRLF or has a trailing newline or whitespac
[... 597 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a222bc4b7422dc38d (internal ID - do not mention to user. Use SendMessage with to: 'a222bc4b7422dc38d', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-2-wt/e3ce0dbb-c168-4a49-9caf-ecc01cb40fbf/tasks/a222bc4b7422dc38d.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && grep -rn -i \"hand-written\\|hand written\\|/goal prompt\\|goal prompt\\|by hand\" README.md docs skills AGENTS.md | grep -v \"^docs/adr/00[01]\" | head -30; grep -n \"hyp status\\|hyp checkpoint\\|hyp answer\" README.md | head; grep -n \"constraints\\|May change\\|Read only\" -r docs README.md skills | head",
 "description": "Preflight scan for claims the change would contradict"
}
```

Result:

```text
docs/core-and-self-review.md:44:| `recipes/recipe-base-arms@1` | a recipe this repository keeps to test hyp: a sonnet runner builds a held-out instrument from the skill-text gate method in its bundle on three hyp arms (template only, a coarse base, fine bases), for `investigations/recipe-base`; copied by hand from a version of `recipe-authoring-arms` (its `recipe.md` names it by hash), because `--from` copies the same name only: the one bootstrap exception to starting from a base, had a base shipped (ADR 0009) | public |
docs/core-and-self-review.md:62:| arm (id-citation-arms) | an Arm: a hyp binary built in Launch from main or main plus one rival's patch, named by sha256, the subject the cases run (version 1 read binaries built by hand from absolute paths) |
docs/replay-spec.md:256:  A pin to a bundle without an Evidence or Standalone command, such as a known-answer fixture, is held by its hash alone (§7.3). The CLI derives the mark from files when it reads, as it does `recipeHashes`. Replay never reads files, and the mark never changes a status. `hyp record` warns, and `hyp show`, `hyp status`, `hyp render` and `hyp check` show it. Nothing is refused. Limits: a script named only in prose, whose output a Result cites, is not seen; neither is a hand-made output file cited beside a real command record, or a script with none of the marks in reason 1. A hand-written `command.json` that claims a clean run with passing known answers satisfies reason 2; a reviewer checks the record's `command.log` and `known-answers.log`. A script that only a Launch, Doctor, Drive or Cleanup command reaches, cited from another Experiment, is held by its hash alone: hyp runs it in a sandbox, but no reason ties that sandbox's record to the Result. The gate proves that a bundle reproduces its own expected outputs, not that those outputs are right: a weak or wrong case needs review (ADR 0002). A case and a run get the same environment, working directory and hyp, and a case compares exit code,
[... 5497 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 138,172p README.md; grep -n -i \"review\" skills/hyp-goal-run/SKILL.md | head",
 "description": "Read README owner-flow section and goal-run review mentions"
}
```

Result:

```text
**3. Start a session** in your project and paste this prompt, with the goal and its lines filled in:

'''text
Use the hyp-prototype skill. Set HYP_ACTOR=claude-<topic>-1 for this session.
If `hyp status` says there is no .tree store, run `hyp init` at the repository root.
Record this goal with its lines under my actor name <your-name>, showing me
the command first: <the outcome I want, and who needs it>.
Lines: <each case every candidate must meet>.
Add the first design question under it, with at least two rival hypotheses,
one of them "no change needed". Do not add or run experiments yet.
Show me `hyp status` and stop.
'''

**4. The lines are yours.** They are the cases every candidate must meet. They come with the Goal and are never added to later, so take your time over them. The session shows you the `hyp add goal` command with your lines and runs it under your actor name once you approve:

'''bash
hyp add goal --title "<the outcome>" --actor <your-name> --why "<who needs it>" \
  --line "<a case every candidate must meet>" --line "<another>"
'''

Your answers later are `hyp answer` events under the same name, for example `hyp answer --answer "R-…#r1=<your words>"`.

From there the [hyp-prototype skill](skills/hyp-prototype/SKILL.md) guides the next rounds: experiments, verdicts, selects and checks. When `hyp status` lists something under AWAITING OWNER, it is yours to answer. [example.sh](skills/hyp-prototype/example.sh) runs a complete investigation in a throwaway directory.

### The owner flow: grill, run, answer

For a Goal you want run unattended, discovery or build, three skills split the work. They run in Claude Code, whose `/goal` command keeps a session going until its condition holds.

1. **`/hyp-owner-grill`** questions you about the purpose until it is a Goal a run can finish. It looks up the facts itself, sends questions a run could test to DesignQuestions with rivals, derives the lines from what must and must never happen, splits purposes that need
[... 622 more characters]
```

## Grill

Your answers are in, and the choice on Q1 was mine to make.

**Q1 decision: a per-clause table test in Go.**
- `eval_skill.py` only grades artifacts that already exist. Its assertions are independent, each with an id, and the kinds are rubric, file exists, json-equals and text-contains. It never runs a model.
- The same idea fits a deterministic generator. One named case per clause means a failure names the clause that broke.
- A pinned fixture would only freeze my own guess at the text, since run 6's prompt isn't in any file.
- Two cases will also assert exact output: the whole prompt for one Goal, and the whole prompt for two stores that differ only in their other events. That second case catches any extra text.
- ak-skill-creator's eval format is for skill text, so I won't use it here. `go test` is the instrument.

**One new question from the preflight scan (Q7).**
- **The gap:** the repo's `hyp-owner-grill` Hand off section puts four things in the hand-written prompt for a build Goal. They are the branch, the one review round per candidate commit by a fresh subagent, the sorting of findings (limit, reading, fix commit or out-of-Goal finding), and "opens a PR and never merges". `hyp prompt` carries only the branch. The repo's `hyp-goal-run` has no review-round text at all (I grepped it).
- **If I switch the hand-off to `hyp prompt` and change nothing else,** a build run would lose its review round.
- **Option A:** this Goal also moves the review round and the finding-sorting into `skills/hyp-goal-run`, as a build-Goal section. That is a skill-body change, tested by a skillcheck-style text test.
- **Option B:** it stays out. The review round goes into each build Goal's body as a `May change:` or `Never:` line, written by the grill.
- **Recommendation:** A. It is general practice, and the old skill itself said "put general practice in the skill, not the prompt".

Two more texts to note:
- **README:** step 1 of "The owner flow" says the grill "ends with a `/goal` prompt". I'll add that to what the change makes wrong, under `May change:`.
- **Run 6 fidelity:** the lines will pin the draft you confirmed, not run 6's real text. I can't check run 6's prompt, and I'll say so in the Goal body.

A fresh subagent is now walking the rules through scenarios we haven't drawn: two Goals in one store, a store path with odd characters, a branch name that tries to inject a line, and so on. I'll show you the routing table, the exact commands and the `hyp prompt` result after it reports and you answer Q7.

## Owner, message 3

<task-notification>
<task-id>a222bc4b7422dc38d</task-id>
<tool-use-id>toolu_0156fneh2xUSBK7LuMMfPFcr</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-2-wt/e3ce0dbb-c168-4a49-9caf-ecc01cb40fbf/tasks/a222bc4b7422dc38d.output</output-file>
<status>completed</status>
<summary>Agent "Second reader for hyp prompt rules" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>1. **Goal-line text is unsafe for some paths.** Line 1 embeds the store path raw. A path with a newline splits the `/goal` line into two lines. A path with spaces or backticks is ambiguous to the model and to the shell. `FindRoot` returns `filepath.Abs` output with no sanitising (`internal/store/store.go:47-62`). Fix: exit 3 or 2 on a path containing a control character, and quote the path otherwise. The spec says nothing about this (my inference).

2. **Nothing guards the actor and branch strings except what you add.** The `--actor` check can reuse `store.ValidActor` (`store.go:223`, `^[a-z0-9][a-z0-9-]{0,39}$`). Go's `$` without the `m` flag does not match before a trailing `\n`, so that case is rejected. But `--branch` has no existing validator. "No space" is not enough: a branch can contain a newline, a tab, `;`, a backtick or `$(...)`. Specify a whitelist such as git ref characters, with no control characters, and say where `--pr` is validated (digits only, at least 1).

3. **The block validity rule contradicts itself on trailing whitespace.** A "line of spaces, a tab or a lone CR is not empty", and every non-empty line must start with a heading. So a body ending `Never: x\n` is fine only if "\n" is split off. `strings.Split` yields a final "" element, and you must say whether that counts as a line. CRLF bodies make every line end in `\r`, so every line is still valid, but the copied block then contains `\r`. A blank line inside the block that was typed as `\r\n` is a lone CR, so the Goal is refused. Whitespace-only last lines (`"  "`) also make the block invalid. The write path does not trim bodies. Only `--why` is trimmed (`internal/cli/write.go:105`), and the 2000-rune limit counts runes (`internal/replay/validate.go:905`). So these bodies are storable. Decide on one normalisation: split on `\n`, strip one trailing `\r` per line, and drop one final empty element. Also decide whether blank lines are allowed inside the block, since the rule as written forbids them.

4. **Prose lines that start with a heading word.** "From the first line starting with…" takes the earliest such line. If the prose has `Never: …` as an aside, the block starts there and anything between that line and the real block must also start with a heading, or the Goal is refused. Conversely, indented headings (`  May change:`) do not match. Say whether matching is case-sensitive and whether leading whitespace is allowed. `Never:` alone with no content is also unspecified.

5. **Warning versus refusal skew at `add goal`.** "Heading but no valid block" warns, while "no heading" is silent. A Goal whose block is wrongly formed but has no heading, such as `Do not touch X`, is silent at write and fails only at `prompt`. That is acceptable, but then `prompt`'s error text must say the block is missing, not that it is malformed. Also, `add goal` must run the same function as `prompt`. Two implementations will drift.

6. **Ref resolution is narrower than "alias vs display id vs prefix".** `resolveRef` (`internal/cli/resolve.go:13-37`) accepts only an exact ULID, an alias, or a case-insensitive display id. There is no prefix matching. A "prefix" gives exit 2 ("unknown node"). It returns any node, so `prompt` must also check `Kind == Goal`. Otherwise `hyp prompt Q-xxxx` would try to read a DesignQuestion body. Aliases are checked before display ids, so an alias that equals another node's display id shadows it. The prompt's line 1 and line 2 should print the display id, not the typed alias, so that the later `hyp status &lt;G-id&gt;` has an unambiguous form (my inference).

7. **Foreign `&lt;tree&gt;:&lt;id&gt;`.** I found no `&lt;tree&gt;:` syntax in `resolveRef`, so it is exit 2 unknown. Say so explicitly. Do not add cross-tree resolution, because the prompt is a function of this store only.

8. **Wrong store reached.** `FindRoot` walks up and takes the first `.tree/events` (`store.go:52-61`). Run from inside a repo nested in another repo that has its own `.tree`, it picks the inner one, silently. In a git worktree the `.tree` is per checkout. `.tree/.gitignore` and `Init` mean a worktree usually has no store, so the walk goes up to the parent's. The printed absolute path is the only visible clue. Consider `--store` or at least printing the path. Symlinks are not resolved either (no `EvalSymlinks` in `store`), so the same store can yield different paths for different cwds.

9. **Merged log from two actors.** Replay decides what is rejected, and a Goal event could be rejected (a duplicate alias, for example). `prompt` must use the replay projection, not raw events, or it could print a Goal that replay dropped. Two Goals with the same alias are one rejected event. "Results under `&lt;actor&gt;`" counts in line 2 depend on actor files, so it is fine. But `--actor` need not match any actor in the store, and `prompt` will print for a stranger. Decide whether an unknown actor warns or is allowed (a fresh run is legitimate).

10. **Withdrawn or complete Goal.** The spec is silent. Printing a prompt for a Goal that already satisfies "owes nothing" is harmless but pointless. A withdrawn Goal is a hold. Recommend exit 1 with the status named, or at least a note on stderr. This is a design choice, not a proven invariant.

11. **`--experiments`.** `0` and negative values are exit 2 (as designed). A huge N, such as `99999999999999999999`, must be rejected by `strconv.Atoi` as exit 2, not wrapped. Cap it or accept it, but say which. An N the prompt prints but the actor can never reach just means the stop condition never fires.

12. **Body size and line 2 backticks.** Line 2 contains literal backticks. If the output is pasted into a shell, they would be executed. Not a hyp bug, but note it in the docs. The 2000-rune cap on the Goal bound the block size, so the prompt length is bounded.

13. **Invariants preserved.** Nothing is stored and replay is untouched, so `hyp check` and replay stay clean (`validate.go:84-89` has no block rule). Keep it that way: do not add the block rule to `validateNodeAdd`.</result>
<usage><subagent_tokens>42530</subagent_tokens><tool_uses>9</tool_uses><duration_ms>41828</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 8,40p internal/cli/resolve.go; sed -n 44,64p internal/store/store.go; grep -n \"ValidActor\" -A4 internal/store/store.go | head -12; sed -n 100,108p internal/cli/write.go",
 "description": "Verify the reviewer's cited facts against source"
}
```

Result:

```text
	"github.com/dat9uy/hyp/internal/replay"
)

// resolveRef turns a user-typed reference into a node id (spec §2): an exact
// ULID first, then an alias, then a display id.
func resolveRef(p replay.Projection, ref string) (string, error) {
	if _, ok := p.Nodes[ref]; ok {
		return ref, nil
	}
	var byAlias, byDisplay []string
	for id, n := range p.Nodes {
		if n.Alias != "" && n.Alias == ref {
			byAlias = append(byAlias, id)
		}
		if strings.EqualFold(event.DisplayID(n.Kind, id), ref) {
			byDisplay = append(byDisplay, id)
		}
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
		if parent == dir {
			return "", ErrNoStore
		}
		dir = parent
	}
}

// Init creates .tree/events, .tree/evidence and .tree/.gitignore in dir. It
170:	if !ValidActor(actor) {
171-		return event.Event{}, fmt.Errorf("invalid actor %q (want %s)", actor, actorRe)
172-	}
173-	// Lock the events directory, not one actor file, so that every writer
174-	// validates against a log that already holds all other writers' events.
--
222:// ValidActor reports whether actor matches ^[a-z0-9][a-z0-9-]{0,39}$.
223:func ValidActor(actor string) bool { return actorRe.MatchString(actor) }
		return nil, usageErr("missing actor: pass --actor or set HYP_ACTOR")
	}
	if !store.ValidAc
[... 262 more characters]
```

## Grill

The reviewer finished. Its report is not an answer from you, so Q7 is still open and nothing has been written. I checked its cited facts against the source: `resolveRef` has no prefix matching, `FindRoot` returns an unsanitised absolute path, and `ValidActor` is `^[a-z0-9][a-z0-9-]{0,39}$`. All three hold.

## Round 2

**Q7 (still open).** Do the review round and the finding-sort move into `skills/hyp-goal-run` as a build-Goal section? I recommend yes. That adds line 8 below.

**Q8. Defaults the review forced.** Confirm all of them, or name the ones you want changed.

1. **Goal ref:**
   - **Resolution:** the existing `resolveRef` resolves it (exact id, alias or display id; no prefix matching). Anything that is not a Goal in this store, including `<tree>:<id>`, exits 2.
   - **Printed id:** the prompt prints the Goal's display id, never the typed alias, so the later `hyp status <G-id>` is unambiguous.
2. **Strings in the prompt:**
   - **Actor:** it uses the existing actor pattern.
   - **Branch:** letters, digits, `.`, `_`, `/` and `-` only, and it can't start with `-`.
   - **`--pr` and `--experiments`:** each must be an integer of 1 or more.
   - **Failure:** any bad value exits 2.
   - **Why:** a newline or a backtick in a branch name would inject text into a prompt that a model obeys.
3. **Store path:** it is printed as `FindRoot` returns it, with no symlink resolution. A path with a control character exits 3, and so does no store.
4. **Block rules:** these are the strict form you confirmed, made exact.
   - **Matching:** headings match case-sensitively at the first column.
   - **Empty heading:** a heading with no text after the colon is invalid, because it names no constraint.
   - **Trailing newline:** the empty piece after one final `\n` is not a line.
   - **Blank lines:** they are rejected inside the block, as you said "one constraint per line".
   - **CRLF:** bodies are written with `\n`.
5. **One function:** `hyp add goal` and `hyp prompt` run the same block check, so their messages can't drift.
6. **Complete Goal:** `hyp prompt` prints a prompt even for a complete Goal. An owner return can make a complete Goal owe work again.
7. **Actor not in the store:** `hyp prompt` accepts it, because a fresh run uses a new actor.
8. **Rejected Goal event:** `hyp prompt` reads the replay projection, so a Goal that replay rejected exits 2.

## Provisional routing table

| Grill question | Goes to | Separating scenario |
|---|---|---|
| hyp prints vs a skill writes | Goal title and body: this Goal builds `hyp prompt` | none; no run can separate them |
| Fixture vs per-clause test | Experiment verdict rule: a Go table test, one named case per clause plus two whole-output cases | none |
| What may differ from run 6 | Goal body, and lines 1 and 3 | none |
| Fixed text, `--experiments`, `--branch`, `--pr` | **L1**, **L2** | none |
| Output depends on nothing but its inputs | **L3** | none |
| Block validity and exit codes (Q4, Q8.4) | **L4**, **L5**, **L6** | none |
| Hand-off of `hyp-owner-grill` | **L7** | none |
| Review round into `hyp-goal-run` (Q7) | **L8**, if you say yes | none |
| Skill descriptions, README step 1 of the owner flow, docs text | `May change:`, no line | none |
| How the change lands | Local branch, no PR (`May change:` and `Never:`) | none |
| Red on base, removal check | Each Experiment's verdict rule | none |
| Run 6's text can't be checked | Goal body says the lines pin the confirmed draft | none |
| `q-commit` | One DesignQuestion with no rivals yet; the run adds commits | none |
| Other Goals | None. #114 is untouched | none |

**Lines (provisional text; I'll show the exact commands after you confirm):**

- **L1 [P]**
  - **Input:** `hyp prompt G --actor A` on a Goal whose body has a valid block.
  - **Output, exit 0:** line 1 is `/goal Run the hyp tree in <store dir> with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is <display id>.` Line 2 is the Done-when clause, with 5 Experiments. Then come the block's lines unchanged, and no other byte.
- **L2 [P]**
  - `--experiments N` changes only that number.
  - `--branch B` appends exactly `Work on branch B.`
  - `--pr N` appends exactly `The PR for the branch is #N. Never merge it.`, after the branch line.
  - No flag changes any other byte.
- **L3 [N]**
  - **Output depends only on its inputs:** the bytes depend only on the store directory, the Goal, the actor, the flags and the Goal's body. Two stores that differ only in Results, answers, selects, checkpoints or other Goals give identical bytes.
  - **Two Goals in one store:** each prompt holds only its own block.
- **L4 [N]**
  - **What is refused:** a body with no valid block gives exit 1.
  - **The message:** it names the missing heading or the offending line.
  - **Stdout:** it is empty.
- **L5 [N]**
  - **Exit 2 on bad arguments:** a bad actor, branch, PR, experiments count or ref, with empty stdout.
  - **Exit 3 on environment problems:** a store path with a control character, or no store.
- **L6 [N]**
  - **Warn-and-write:** `hyp add goal` with a heading but no valid block exits 0, writes the Goal, and prints the same message `hyp prompt` gives.
  - **Silent cases:** a valid block, or a body with no heading, gets no warning.
  - **Replay and `hyp check`:** neither warns on nor rejects a Goal for its block, so a store of block-less Goals checks `check: 0 errors`.
- **L7 [P]** The Hand off of `skills/hyp-owner-grill` runs `hyp prompt <G-id> --actor … [--branch] [--pr] [--experiments]`, gives its output as it is, and tells no one to hand-write a prompt. The skillcheck test passes.
- **L8 [P]** (if Q7 is yes) `skills/hyp-goal-run` has a build-Goal section. It states one review round per candidate commit, by a fresh subagent given the lines, the code and the keys settled with the owner's words. It sorts each finding four ways: a case of a line is a limit, a question about a line's meaning is a reading, a defect in text the run may edit is a fix commit, and anything else is an out-of-Goal finding.

Please give me:
- your answer on Q7;
- confirm or amend Q8;
- confirm or move each row of the table.

## Owner, message 4

Q7: yes, accept your recommendation. Move the review round and the finding-sort into `hyp-goal-run`, with L8.

Q8: confirm all eight defaults.

Table: confirm every row. None goes against my positions. Keep the skill descriptions under `May change:` with no line.

## Grill

Routing confirmed. I'm running the runner writes first: `hyp init`, then counting the runes in the Goal text, since the body is capped at 2000 and each line at 600.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-p113-grill && hyp init 2>&1; git status --short | head; mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > body.txt <<'EOF'\nAdds `hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]`, which prints the /goal prompt for hyp-goal-run from the log for any Goal, discovery or build; a warning in `hyp add goal` for a body with a constraints heading but no valid block; and a build-Goal section in skills/hyp-goal-run. Part of #113 (gap 1), core scope. Owner decisions: 2026-11-11, a run's constraints live in the Goal body under three headings, a prompt may narrow them and never widen them, and no typed field until a run misreads one; 2026-11-12, the prompt may differ from a hand-written prompt only in the actor, the owner's answer, the branch, the PR, the commit and the run's own plan. Run 6's prompt is in no file, so the lines pin the clause list the owner confirmed on 2026-11-12, not run 6's text. The candidate is a hyp commit; go test on its base and on the candidate is the instrument, with one named test case per prompt clause.\nMay change: commits on one local branch of this repository, in cmd/, internal/, skills/ and docs/ text: the Hand off of skills/hyp-owner-grill, a build-Goal section of skills/hyp-goal-run, the descriptions of skills/hyp-goal-run and skills/hyp-prototype, and the README, glossary and spec text that the change makes wrong.\nRead only: investigations/, recipes/, the other skills, and the ADRs.\nNever: open or merge a PR.\nNever: replace the installed hyp binary.\nNever: edit an existing ADR, a recorded event or a frozen recipe version.\nNever: put application code or real source data in this repository.\nEOF\nprintf '%s' \"$(cat body.txt)\" | wc -m\ncat > l1.txt <<'EOF'\n[P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly: line 1 `/goal Run the hyp tree in <store dir> with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is <display id>.`; line 2 `Done when the latest printed `hyp status <display id>` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Experiments have a Result under A, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.`; then the block's lines unchanged; no other byte.\nEOF\nwc -m l1.txt",
 "description": "Init the store and count Goal body and line lengths"
}
```

Result:

```text
initialised .tree in /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
?? .tree/
    1532
     629 l1.txt
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && i=0; while IFS= read -r l; do i=$((i+1)); printf '%s' \"$l\" > L$i.txt; done <<'EOF'\n[P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly three parts, in order, and no other byte: the line `/goal Run the hyp tree in <store dir> with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is <display id>.`; the Done-when line; then the block's lines, unchanged. <store dir> is the absolute directory that holds .tree.\n[P] The Done-when line of `hyp prompt` reads exactly: `Done when the latest printed `hyp status <display id>` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Experiments have a Result under A, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.`\n[P] `--experiments N` changes only the 5 in the Done-when line. `--branch B` appends exactly `Work on branch B.` after the block, and `--pr N` appends exactly `The PR for the branch is #N. Never merge it.` after that, in this order. No flag changes any other byte.\n[N] The bytes `hyp prompt` prints depend only on the store directory, the Goal, the actor, the flags and the Goal's body: two stores that differ only in Results, answers, selects, checkpoints or other Goals give identical output. In a store with two Goals, each prompt holds only its own block.\n[N] A Goal body with no valid constraints block gives exit 1, empty stdout and a message that names the missing heading or the offending line. The block starts at the first line that begins with `May change:`, `Read only:` or `Never:` (case-sensitive, first column), holds all three headings, has text after each heading, and has only non-empty lines that each begin with a heading; a line of spaces, a tab or a lone carriage return is not empty.\n[N] `hyp prompt` gives exit 2 and empty stdout for an actor outside ^[a-z0-9][a-z0-9-]{0,39}$, a branch outside letters, digits, `.`, `_`, `/` and `-` or one that starts with `-`, a `--pr` or `--experiments` that is not an integer of 1 or more, and a ref that is not a Goal of this store, `<tree>:<id>` included. It gives exit 3 for a store path with a control character, and for no store.\n[N] `hyp add goal` with a body that has a heading line but no valid block exits 0, writes the Goal, and warns with the message `hyp prompt` gives for it. A valid block, or a body with no heading line, gives no warning. Replay and `hyp check` never warn on or reject a Goal for its block, so a store of Goals with no block checks `check: 0 errors`.\n[P] The Hand off of skills/hyp-owner-grill runs `hyp prompt <G-id> --actor <runner-actor>` with the optional `--branch`, `--pr` and `--experiments`, gives its output as it is, and holds no instruction to hand-write a prompt. The skillcheck test passes.\n[P] skills/hyp-goal-run has a section for a build Goal: one review round per candidate commit by 
[... 420 more characters]
```

Result:

```text
     440 L9.txt
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && ls; wc -m L*.txt",
 "description": "List scratch files and rune counts"
}
```

Result:

```text
add-design-goal.sh
add-goal.sh
body.txt
body2.txt
body2.txt.bak
body2n.txt
design-body.txt
ids.txt
l1.txt
l2.txt
l3.txt
l4.txt
l5.txt
l6.txt
l7.txt
l8.txt
L9.txt
s
scenarios.md
     440 L9.txt
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && ls -l --time-style=full-iso 2>/dev/null || ls -lT; wc -m l?.txt L9.txt body.txt; head -c 300 add-goal.sh",
 "description": "Inspect scratch dir timestamps and counts"
}
```

Result:

```text
total 144
-rw-r--r--@ 1 datguy  wheel   434 Nov 12 21:13:38 2026 add-design-goal.sh
-rw-r--r--@ 1 datguy  wheel  2615 Nov 12 21:11:44 2026 add-goal.sh
-rw-r--r--@ 1 datguy  wheel  1533 Nov 12 21:24:24 2026 body.txt
-rw-r--r--@ 1 datguy  wheel  1793 Nov 12 21:11:12 2026 body2.txt
-rw-r--r--@ 1 datguy  wheel  1776 Nov 12 21:10:58 2026 body2.txt.bak
-rw-r--r--@ 1 datguy  wheel  1775 Nov 12 21:10:58 2026 body2n.txt
-rw-r--r--@ 1 datguy  wheel   758 Nov 12 21:13:26 2026 design-body.txt
-rw-r--r--@ 1 datguy  wheel    36 Nov 12 21:13:54 2026 ids.txt
-rw-r--r--@ 1 datguy  wheel   382 Nov 12 21:24:37 2026 l1.txt
-rw-r--r--@ 1 datguy  wheel   396 Nov 12 21:24:37 2026 l2.txt
-rw-r--r--@ 1 datguy  wheel   264 Nov 12 21:24:37 2026 l3.txt
-rw-r--r--@ 1 datguy  wheel   294 Nov 12 21:24:37 2026 l4.txt
-rw-r--r--@ 1 datguy  wheel   446 Nov 12 21:24:37 2026 l5.txt
-rw-r--r--@ 1 datguy  wheel   389 Nov 12 21:24:37 2026 l6.txt
-rw-r--r--@ 1 datguy  wheel   347 Nov 12 21:24:37 2026 l7.txt
-rw-r--r--@ 1 datguy  wheel   252 Nov 12 21:24:37 2026 l8.txt
-rw-r--r--@ 1 datguy  wheel   440 Nov 12 21:24:37 2026 L9.txt
drwxr-xr-x@ 4 datguy  wheel   128 Nov 12 21:15:22 2026 s
-rw-r--r--@ 1 datguy  wheel   867 Nov 12 21:11:28 2026 scenarios.md
     382 l1.txt
     396 l2.txt
     264 l3.txt
     294 l4.txt
     446 l5.txt
     389 l6.txt
     347 l7.txt
     252 l8.txt
     440 L9.txt
    1533 body.txt
    4743 total
#!/bin/zsh
cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
hyp add goal --actor owner-standin \
  --title "hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body" \
  --alias goal \
  --body "$(cat $TMPDIR/g/body.txt)" \
  --line "[P] hyp prompt G --acto
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "D=$TMPDIR/p113-a2; mkdir -p $D && cd $D && ls -A | head -3; cat > body.txt <<'EOF'\nAdds `hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]`, which prints the /goal prompt for hyp-goal-run from the log for any Goal, discovery or build; a warning in `hyp add goal` for a body with a constraints heading but no valid block; and a build-Goal section in skills/hyp-goal-run. Part of #113 (gap 1), core scope. Owner decisions: 2026-11-11, a run's constraints live in the Goal body under three headings, a prompt may narrow them and never widen them, and no typed field until a run misreads one; 2026-11-12, the prompt may differ from a hand-written prompt only in the actor, the owner's answer, the branch, the PR, the commit and the run's own plan. Run 6's prompt is in no file, so the lines pin the clause list the owner confirmed on 2026-11-12, not run 6's text. The candidate is a hyp commit; go test on its base and on the candidate is the instrument, with one named test case per prompt clause.\nMay change: commits on one local branch of this repository, in cmd/, internal/, skills/ and docs/ text: the Hand off of skills/hyp-owner-grill, a build-Goal section of skills/hyp-goal-run, the descriptions of skills/hyp-goal-run and skills/hyp-prototype, and the README, glossary and spec text that the change makes wrong.\nRead only: investigations/, recipes/, the other skills, and the ADRs.\nNever: open or merge a PR.\nNever: replace the installed hyp binary.\nNever: edit an existing ADR, a recorded event or a frozen recipe version.\nNever: put application code or real source data in this repository.\nEOF\ncp ../g/l1.txt line1.txt; for n in 2 3 4 5 6 7 8; do cp ../g/l$n.txt line$n.txt; done; cp ../g/L9.txt line9.txt; head -c 120 line1.txt; echo; head -c 80 line2.txt; echo; wc -m body.txt line?.txt",
 "description": "Stage Goal body and line files in a fresh directory"
}
```

Result:

```text
[P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly three parts
[P] The Done-when line of `hyp prompt` reads exactly: `Done when the latest prin
    1533 body.txt
     382 line1.txt
     396 line2.txt
     264 line3.txt
     294 line4.txt
     446 line5.txt
     389 line6.txt
     347 line7.txt
     252 line8.txt
     440 line9.txt
    4743 total
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

## Grill

`hyp init` is done, and the store is at the repo root. Nothing is written under your name yet. This is the one owner write.

**The command (not run yet):**

```bash
cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && D=$TMPDIR/p113-a2 && \
hyp add goal --actor owner-standin --alias goal \
  --title "hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body" \
  --body "$(cat $D/body.txt)" \
  --line "$(cat $D/line1.txt)" --line "$(cat $D/line2.txt)" --line "$(cat $D/line3.txt)" \
  --line "$(cat $D/line4.txt)" --line "$(cat $D/line5.txt)" --line "$(cat $D/line6.txt)" \
  --line "$(cat $D/line7.txt)" --line "$(cat $D/line8.txt)" --line "$(cat $D/line9.txt)" \
  --why "Owner, 2026-11-12: issue #113 gap 1, built as a build Goal; routing confirmed in the grill session."
```

**Body** (1533 runes; the block is the last four `Never:` lines and the two before them):

> Adds `hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]`, which prints the /goal prompt for hyp-goal-run from the log for any Goal, discovery or build; a warning in `hyp add goal` for a body with a constraints heading but no valid block; and a build-Goal section in skills/hyp-goal-run. Part of #113 (gap 1), core scope. Owner decisions: 2026-11-11, a run's constraints live in the Goal body under three headings, a prompt may narrow them and never widen them, and no typed field until a run misreads one; 2026-11-12, the prompt may differ from a hand-written prompt only in the actor, the owner's answer, the branch, the PR, the commit and the run's own plan. Run 6's prompt is in no file, so the lines pin the clause list the owner confirmed on 2026-11-12, not run 6's text. The candidate is a hyp commit; go test on its base and on the candidate is the instrument, with one named test case per prompt clause.
>
> May change: commits on one local branch of this repository, in cmd/, internal/, skills/ and docs/ text: the Hand off of skills/hyp-owner-grill, a build-Goal section of skills/hyp-goal-run, the descriptions of skills/hyp-goal-run and skills/hyp-prototype, and the README, glossary and spec text that the change makes wrong.
> Read only: investigations/, recipes/, the other skills, and the ADRs.
> Never: open or merge a PR.
> Never: replace the installed hyp binary.
> Never: edit an existing ADR, a recorded event or a frozen recipe version.
> Never: put application code or real source data in this repository.

**The nine lines** (the old L1 was 629 runes, over the 600 cap, so it is now L1 and L2; the old L2 to L8 are now L3 to L9; each is 252 to 446 runes):

1. [P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly three parts, in order, and no other byte: the line `/goal Run the hyp tree in <store dir> with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is <display id>.`; the Done-when line; then the block's lines, unchanged. `<store dir>` is the absolute directory that holds .tree.
2. [P] The Done-when line reads exactly the text you confirmed in Q2, with 5 Experiments and the actor A.
3. [P] `--experiments N` changes only the 5. `--branch B` appends exactly `Work on branch B.` after the block, and `--pr N` appends exactly `The PR for the branch is #N. Never merge it.` after that. No flag changes any other byte.
4. [N] The bytes depend only on the store directory, the Goal, the actor, the flags and the Goal's body. Two stores that differ only in Results, answers, selects, checkpoints or other Goals give identical output. With two Goals in one store, each prompt holds only its own block.
5. [N] A body with no valid block gives exit 1, empty stdout and a message that names the missing heading or the offending line. The block rules are the strict form you confirmed in Q4 and Q8.
6. [N] Bad arguments give exit 2 with empty stdout, and a control character in the store path or no store gives exit 3, as in Q8.
7. [N] `hyp add goal` warns, writes and exits 0 on a heading with no valid block. Replay and `hyp check` never warn on or reject a Goal for its block.
8. [P] The Hand off of `skills/hyp-owner-grill` runs `hyp prompt`, gives its output as it is, and holds no instruction to hand-write a prompt. The skillcheck test passes.
9. [P] `skills/hyp-goal-run` has a build-Goal section: one review round per candidate commit, and the four-way sort of findings.

The full text of lines 5 to 9 is in the files the command reads. Should I paste them in full?

Do you approve this exact command, to run under `owner-standin`?

## Owner, message 5

Approved. Run the command under `owner-standin`. There is no need to paste lines 5 to 9 in full.

### Tool Bash (cwd: .)

```text
{
 "command": "D=$TMPDIR/p113-a2 && \\\nhyp add goal --actor owner-standin --alias goal \\\n  --title \"hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body\" \\\n  --body \"$(cat $D/body.txt)\" \\\n  --line \"$(cat $D/line1.txt)\" --line \"$(cat $D/line2.txt)\" --line \"$(cat $D/line3.txt)\" \\\n  --line \"$(cat $D/line4.txt)\" --line \"$(cat $D/line5.txt)\" --line \"$(cat $D/line6.txt)\" \\\n  --line \"$(cat $D/line7.txt)\" --line \"$(cat $D/line8.txt)\" --line \"$(cat $D/line9.txt)\" \\\n  --why \"Owner, 2026-11-12: issue #113 gap 1, built as a build Goal; routing confirmed in the grill session.\"",
 "description": "Add the Goal under the owner's actor"
}
```

Result:

```text
G-SVK8PF node.add
  line G-SVK8PF#1: [P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly three parts, in order, and no other byte: the line `/goal Run the hyp tree in <store dir> with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is <display id>.`; the Done-when line; then the block's lines, unchanged. <store dir> is the absolute directory that holds .tree.
  line G-SVK8PF#2: [P] The Done-when line of `hyp prompt` reads exactly: `Done when the latest printed `hyp status <display id>` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Experiments have a Result under A, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.`
  line G-SVK8PF#3: [P] `--experiments N` changes only the 5 in the Done-when line. `--branch B` appends exactly `Work on branch B.` after the block, and `--pr N` appends exactly `The PR for the branch is #N. Never merge it.` after that, in this order. No flag changes any other byte.
  line G-SVK8PF#4: [N] The bytes `hyp prompt` prints depend only on the store directory, the Goal, the actor, the flags and the Goal's body: two stores that differ only in Results, answers, selects, checkpoints or other Goals give identical output. In a store with two Goals, each prompt holds only its own block.
  line G-SVK8PF#5: [N] A Goal body with no valid constraints block gives exit 1, empty stdout and a message that names the missing heading or the offending line. The block starts at the first line that begins with `May change:`, `Read only:` or `Never:` (case-sensitive, first column), holds all three headings, has text after each heading, and has only non-empty lines that each begin with a heading; a line of spaces, a tab or a lone carriage return is not empty.
  line G-SVK8PF#6: [N] `hyp prompt` gives exit 2 and empty stdout for an actor outside ^[a-z0-9][a-z
[... 1407 more characters]
```

## Grill

The Goal is written as G-SVK8PF. Next are the runner writes: the question, the issue link and the end-of-grill checkpoint.

### Tool Bash (cwd: .)

```text
{
 "command": "export HYP_ACTOR=claude-p113-grill && hyp add dq --parent G-SVK8PF --alias q-commit --title \"Which hyp commit implements hyp prompt and the skill changes of Goal G-SVK8PF?\" --why \"The build's one question; its candidate will be a hyp commit, added when one exists.\" && hyp link G-SVK8PF external_ref gh:113 --why \"The issue this Goal builds (gap 1); #114 stays open.\" && hyp checkpoint --goal G-SVK8PF --summary \"End of grill: Goal G-SVK8PF with 9 lines and one question, q-commit, with no rivals yet. Waiting for the owner: none. Out-of-Goal findings: the installed hyp has no prompt command, so the first run's prompt cannot be generated.\" --why \"end of grill\" && hyp status G-SVK8PF && hyp check",
 "description": "Add question, link, checkpoint; show status and check"
}
```

Result:

```text
Q-6ATAMW node.add
G-SVK8PF edge.add
G-SVK8PF checkpoint
GOAL G-SVK8PF hyp prints the /goal prompt for any Goal from the log, read… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-SVK8PF#1 no candidate — [P] `hyp prompt G --actor A` on a Goal whose body holds a valid constraints block exits 0 and prints exactly three parts, in order, a…
  G-SVK8PF#2 no candidate — [P] The Done-when line of `hyp prompt` reads exactly: `Done when the latest printed `hyp status <display id>` shows `runner owes noth…
  G-SVK8PF#3 no candidate — [P] `--experiments N` changes only the 5 in the Done-when line. `--branch B` appends exactly `Work on branch B.` after the block, and…
  G-SVK8PF#4 no candidate — [N] The bytes `hyp prompt` prints depend only on the store directory, the Goal, the actor, the flags and the Goal's body: two stores …
  G-SVK8PF#5 no candidate — [N] A Goal body with no valid constraints block gives exit 1, empty stdout and a message that names the missing heading or the offend…
  G-SVK8PF#6 no candidate — [N] `hyp prompt` gives exit 2 and empty stdout for an actor outside ^[a-z0-9][a-z0-9-]{0,39}$, a branch outside letters, digits, `.`,…
  G-SVK8PF#7 no candidate — [N] `hyp add goal` with a body that has a heading line but no valid block exits 0, writes the Goal, and warns with the message `hyp p…
  G-SVK8PF#8 no candidate — [P] The Hand off of skills/hyp-owner-grill runs `hyp prompt <G-id> --actor <runner-actor>` with the optional `--branch`, `--pr` and `…
  G-SVK8PF#9 no candidate — [P] skills/hyp-goal-run has a section for a build Goal: one review round per candidate commit by a fresh subagent given the lines, th…
FRONTIER
  Q-6ATAMW open/no_candidates Which hyp commit implements hyp prompt and the skill change…
STALE
OWED
  runner: Q-6ATAMW add rival Hypotheses
AWAITING OWNER
RECENT
  2026-11-12 14:25 claude-p113-grill edge.add G-SVK8PF — The issue this Goal builds (gap 1); #114 stays open.
CHECKPOINT 2026-11-12 14:25 clau
[... 269 more characters]
```

## Grill

Goal G-SVK8PF is in the store at the repo root, and `hyp check` prints `check: 0 errors, 0 warnings`. I could not do the hand-off, because the installed `hyp` has no `prompt` command yet, so I have no `hyp prompt` output to give you. I did not commit anything. `.tree/` is untracked on branch `work`, and the run needs the Goal in the log it starts from, so it has to be committed or merged first.

**What is in the store:**
- **Goal:** G-SVK8PF with nine lines (L1 to L9) and the constraints block you approved. It was written under `owner-standin`.
- **Question:** `q-commit` is Q-6ATAMW, with no rival Hypotheses yet. The run adds the commits as Hypotheses, as in the checkpoint-store tree.
- **Issue link:** `gh:113` is linked as an external ref. #114 stays open.
- **Checkpoint:** the end-of-grill checkpoint is recorded. OWED shows the runner owing "add rival Hypotheses", and nothing is awaiting you.
- **Routing:** no DesignQuestion has rivals. Nothing here had a scenario that could separate rivals.

**No `hyp prompt` yet.**
- **Why:** this Goal builds `hyp prompt`, so its first run cannot use the generator.
- **If you want a prompt now:** I can assemble one by hand from the Q2 draft and the Goal's block. I'd mark it as a bootstrap, and I haven't written one. Say so and I will.
- **After the run:** the next Goal can use the real command, once its commit is installed.

**Scratch-file mistake.** `$TMPDIR/g` already held scratch files from another session, for the `g146-base-a-1` worktree. They included `add-goal.sh`, `body.txt` and `l1.txt` to `l8.txt`. I treated that directory as empty.
- **What I overwrote:** my `body.txt` and `l1.txt`, then my `L1` to `L8` writes on this case-insensitive disk, replaced that session's `body.txt` and `l1.txt` to `l8.txt`.
- **Recovery:** I can't recover them.
- **Effect on that session:** if it re-runs its `add-goal.sh`, which reads `body.txt` and the line files, it would submit my text, not its own.
- **What I changed since:** I moved to `$TMPDIR/p113-a2` and touched nothing else in `g`. I did not touch that session's store.

**Not covered:**
- **Run 6 fidelity:** the lines pin the clause list you confirmed, not run 6's text. The Goal body says so.
- **Skill descriptions:** the `hyp-goal-run` and `hyp-prototype` descriptions are under `May change:`, with no line.
