# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-a-2/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue and checking the store.

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
 "command": "(which hyp; hyp status 2>&1 | head -40; ls; ls docs)",
 "description": "Check hyp store status and repo layout"
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-base-a-2/bin/hyp
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
 "command": "git log --oneline | head -5; hyp prompt --help 2>&1 | head; grep -rn \"prompt\" cmd internal --include=*.go -il | head; ls investigations skills/*; cat docs/glossary.md | head -150",
 "description": "Check prompt command existence, glossary"
}
```

Result:

```text
a51f140 Merge pull request #137 from dat9uy/docs/discovery-and-build-goals
6191275 docs: name discovery and build Goals and rewrite the README around them
0ff087a Merge pull request #136 from dat9uy/docs/skill-change-is-a-build-goal
8c80f6c docs: treat a skill change as a build Goal and drop the skill-text gate
f3a4d05 Merge pull request #135 from dat9uy/feat/owner-flow-skills
usage: hyp <command> [flags]

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
Statuses are derived from events; no command can set one.

setup:
  init                                   create .tree in the current directory
  version                                print the hyp version

write:
(eval):1: no matches found: --include=*.go
investigations:
checkpoint-store
checkpoint-store-design
id-citation
recipe-base

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
# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.

## Language

### Trees

**Tree**:
One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
_Avoid_: graph, log (the log is its events), project

**Store**:
The `.tree/` directory of a repository: one events file per actor and the evidence files Results point at ([spe
[... 20926 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp help 2>&1 | sed -n 1,60p; grep -rln \"prompt\" cmd internal | head; hyp version; git status --short | head",
 "description": "Check for prompt command in binary and source"
}
```

Result:

```text
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
  show REF|R-id#n|G-id#n [--as-of REF|TIME]  one node, or one limit or line, in full, with its history
  diff --since REF|TIME                       status changes and events since a point
  render --markdown [--as-of REF|TIME]        the whole tree as Markdown, untruncated
  check                                       r
[... 1691 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat skills/hyp-goal-run/SKILL.md | head -120; ls investigations/checkpoint-store/.tree 2>/dev/null; ls docs/adr | tail -8",
 "description": "Read goal-run skill, list ADRs"
}
```

Result:

```text
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
- Try a new or changed check 
[... 10042 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2 && ls; ls tmp; grep -rl \"goal-prompt-261111\" . 2>/dev/null | head; cd wt/investigations/checkpoint-store && hyp status 2>&1 | head -60",
 "description": "Locate run-6 prompt and checkpoint-store tree status"
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
./recipes/hyp-owner-grill-eval/issue-a.md
./tmp/issue-113.md
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
  2026-11-11 14:37 claude-g1b-6 dq.select Q-PC04ED — OWED owes this select: R-RRYBPH on 404
[... 1344 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && grep -rn \"hyp prompt\" --include=* skills docs README.md AGENTS.md internal cmd 2>/dev/null | head -20; ls cmd/hyp internal; grep -rn \"May change\" --include=*.go . | head; grep -n \"May change\\|Read only\" -r skills docs | head -20; diff -rq skills .claude/skills 2>&1 | head",
 "description": "Find prompt references and constraints-block handling in source"
}
```

Result:

```text
(eval):1: no matches found: --include=*
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
(eval):1: no matches found: --include=*.go
skills/hyp-owner-grill/SKILL.md:34:| what a run may change, may only read, or must never do | the Goal body's `May change:`, `Read only:` and `Never:` headings | yes |
skills/hyp-owner-grill/SKILL.md:99:May change: <the repository, how it is changed (for example commits on one local branch, a PR never merged), and the texts the change makes wrong>
skills/hyp-owner-grill/SKILL.md:100:Read only: <repositories, directories or data the run may read but never change>
skills/hyp-owner-grill/SKILL.md:108:Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Never:` lines copied as they are. Put general practice in the skill, not the prompt. For a build Goal it also says:
skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a step needs something they forbid, refer it with a reason or record a reading, and stop.
Only in .claude/skills: ak-skill-creator
Files skills/hyp-owner-grill/SKILL.md and .claude/skills/hyp-owner-grill/SKILL.md differ
```

### Tool Bash (cwd: .)

```text
{
 "command": "diff skills/hyp-owner-grill/SKILL.md .claude/skills/hyp-owner-grill/SKILL.md; sed -n 100,140p skills/hyp-owner-grill/SKILL.md; ls internal/cli internal/replay internal/skillcheck | head -80; wc -l internal/cli/*.go | tail -1",
 "description": "Diff grill skill versions and list CLI sources"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n 'case \"' internal/cli/cli.go | head -60; grep -n \"warn\" internal/cli/write.go | head -20; grep -n \"Body\" internal/replay/state.go | head; grep -rn \"repository\\|exit 3\\|ExitEnv\" internal/herr/*.go | head",
 "description": "Inspect command dispatch, warnings, exit codes"
}
```

Result:

```text
35:		case "vcs.revision":
37:		case "vcs.modified":
159:	case "version":
162:	case "help", "-h", "--help":
296:	var warns []string
388:			if warns, err = admitted(w.st.AdmitExp(p, facts, rubricFacts, noTarget)); err != nil {
406:		for _, l := range warns {
414:// or bad-args finding (exit 2) is the error, and each warning becomes a
415:// "warn: " line to print once the event is written.
416:func admitted(fs []replay.Finding) (warns []string, err error) {
424:			warns = append(warns, "warn: "+f.Text)
427:	return warns, nil
487:	var warns []string
489:		warns = nil
514:		warns = admitWarns
530:			warns = append(warns, fmt.Sprintf("warn: %s aims at %s and this Result does not close it; a close records that the case ran, whatever the verdicts (add --closes %s), and a case the run did not reach is a new --limit",
536:		for _, l := range warns {
539:		warnResult(w.root, w.last.Node, stderr)
544:// warnResult prints the reasons Result id is uncontained (spec §7.6) and
547:func warnResult(root, id string, stderr io.Writer) {
550:		fmt.Fprintf(stderr, "warn: uncontained check skipped: %v\n", err)
555:		fmt.Fprintf(stderr, "warn: uncontained %s (%s)\n", event.DisplayID(event.KindResult, id), why)
782:	var warns []string
798:		warns, err = admitted(w.p.AdmitSelect(dq, h))
113:	ID, Kind, Parent, Title, Body, Alias, Recipe, RecipeHash, DecidedBy, Actor, AddedBy string
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp show G-D1CH1R 2>&1 | sed -n 1,40p; grep -rln \"g1b-run-6\\|goal-prompt\" . 2>/dev/null | grep -v \"^./.git/\" | head",
 "description": "Read G-D1CH1R body, find run-6 prompt in repo"
}
```

Result:

```text
no .tree store found; run 'hyp init'
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | sed -n 1,45p; cd ../.. ; grep -rl \"g1b-run-6\" . 2>/dev/null | grep -v \"^./.git/\" | head",
 "description": "Read G-D1CH1R from its tree"
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
[... 5463 more characters]
```

## Grill

I haven't written anything yet. Round 1 is below.

## What I found

- **No store at the repo root.** My plan is a new tree at `investigations/goal-prompt/`, as AGENTS.md puts core trees under `investigations/<topic>/`.
- **`hyp prompt` doesn't exist.** The installed binary and `cmd/`, `internal/` have no such command. #113 is the build of it.
- **Two copies of the grill skill differ.** The copy I'm running already describes `hyp prompt <G-id> --actor … [--branch] [--pr] [--experiments]`. It also describes a warn on `hyp add goal` for a bad constraints block, a refusal in `hyp prompt` (exit 1), and exits 2 and 3. `skills/hyp-owner-grill/SKILL.md` in the repo is the old hand-written version. I'm reading the loaded copy as the target shape, so the repo copy is a text the change makes wrong.
- **G-D1CH1R can't satisfy "Done means" bullet 1 as written.** Its body has no `May change:`, `Read only:` or `Never:` block. Goals are never edited, so a `hyp prompt` that refuses an invalid block would refuse it. The run-6 prompt text (`goal-prompt-261111-2129-g1b-run-6`) is not anywhere in the repo.

## Questions

**1. One Goal or two?**
- **Today:** the issue folds in the skill-description fixes (`hyp-goal-run` says "discovery /goal", `hyp-prototype` says "Investigate an uncertain design"). A skill change is a build Goal (ADR 0021).
- **Option A: one build Goal.** The candidate is one commit covering the command, the add-goal warning, the skill and doc text, and `skills/hyp-owner-grill`. It has one `May change:` block. Skillcheck fails on base when skill text names `hyp prompt`, so red-on-base holds for the skill text too.
- **Option B: two Goals.** One for the code and one for the skill text. This splits one commit across two runs for no gain.
- **Recommendation: A**, in a new tree `investigations/goal-prompt/`.

**2. "Done means" bullet 1 (match run 6 for G-D1CH1R).**
- **Problem:** it names history and an input the repo doesn't hold, and it conflicts with the refusal rule above. It can't be a line, because a runner can't meet it.
- **Option A: replace it with a synthetic-data test.** A Goal with a valid block produces every clause run 5 had that no step depends on removing, in the same order. The test pins that output on a synthetic Goal under `internal/cli/testdata`. The G-D1CH1R comparison is then not asked, and the Goal body says so.
- **Option B: let `hyp prompt` accept a Goal with no block** and print no constraint lines. That would produce a prompt with no `Never:` lines, which breaks bullet 3, "The 'Never' lines come from a recorded place".
- **Recommendation: A.** Do you accept that, and can you paste the run-6 prompt text so I can derive the synthetic fixture from it?

**3. Which stop clauses does the prompt print?**
- **Today:** run 5 had four. (i) The runner owes nothing. (ii) Every runner item is `add rival Hypotheses` on a Question with no owner request for a new Hypothesis. (iii) 5 Experiments recorded since the goal began. (iv) "cannot take the next step honestly and say why". The sentence "At each stop, record the stop checkpoint…" is already in the `hyp-goal-run` Stop section.
- **Option A:** print (i), (iii) and (iv), and drop (ii). Count (iii) as "N Results under the actor", which replay derives from the log. That needs no "since the goal began" marker, and 5 is the default.
- **Option B:** print all four. This adds a derived-state rule to the generator.
- **Recommendation: A.** Which do you want?

**4. Command or skill?**
- **Today:** your issue says "as something hyp prints or a skill writes from the tree", and the loaded skill already assumes `hyp prompt`.
- **Routing:** this is a testable choice, so it is a DesignQuestion. The rivals are `hyp prompt` and a skill that writes the prompt from `hyp show`. The scenario that separates them is one Goal body with a CRLF line in its block, given to both writers. I ask only whether a run should test it or the tree should record it as decided by this session.
- **Recommendation:** record it as decided, with the session as evidence.

**5. Block validity.**
- **Proposed rule:** the block runs from the first line starting with a heading to the end of the body. All three headings must appear at least once, and every non-empty line must start with one of them. A line of spaces, a tab or a lone carriage return is not empty.
- **Enforcement:** `hyp add goal` warns and does not refuse, because Goals already in logs must replay. It warns only when a heading is present and the block is invalid. `hyp prompt` exits 1 with a message that names the rule.
- **Question:** do you confirm warn at add and refuse at prompt?

**6. What the prompt carries and writes.**
- **Carries:** the store path, the actor, the Goal id, the skill name, the stop clauses and the block lines verbatim. `--branch` and `--pr` each add one narrowing clause after the block.
- **Does not carry:** owner answers or tree state, because OWED lists those.
- **Writes:** nothing. `hyp prompt` appends no event, which follows from the derivation guard.
- **Question:** do you confirm this? Is `owner-dat` your actor name?

## Next round
I'll draft the lines after your answers. They would be `[P]` for a valid block's exact output and `[N]` for exit 1, exit 2, exit 3 and no write.

I still owe the second reader, the self-application check (the Goal changes the tool that judges it) and the preflight scan. After those comes the routing table, and then the writes, which need your approval of each exact owner command.

I can't give you the `/goal` prompt yet. I'll produce it with `hyp prompt` once it exists, and until then the hand-off text is the hand-written form.

## Owner, message 2

1. A, one build Goal. The skill descriptions that name only one kind of Goal are text the change makes wrong, so they go under May change, with no line of their own.

2. A. A Goal body without the constraint headings is refused by `hyp prompt` with exit 1, naming what is missing, and adding such a Goal stays allowed. I can't paste the run-6 text. For the synthetic fixture, widen the exceptions to the actor, the owner's answer, branch, PR and commit, and the run's own plan. Branch and PR become flags.

3. A, as you recommend.

4. Agreed. Hyp writes the prompt as a command that reads the log, and the tree records that as decided in the session. No test is needed.

5. Confirmed. Warn at add, refuse at prompt with exit 1. Keep the strict form, one constraint per line.

6. Confirmed. The prompt does not carry the owner's answers, because the runner reads OWED. The Experiment cap is a flag with a default of 5. My actor name is owner-standin, not owner-dat.

## Grill

Thanks, I'll record all six answers as given. Next are the second reader, the self-application check and the preflight scan. Then I'll show you the routing table.

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader review of prompt rules",
 "subagent_type": "general-purpose",
 "prompt": "Repo: /Users/datguy/.hyp-sandbox/g146-base-a-2/wt (Go CLI `hyp`; read AGENTS.md, docs/replay-spec.md, docs/glossary.md, internal/cli/cli.go, write.go, read.go). READ ONLY: change nothing.\n\nA new read command `hyp prompt <G-id> --actor A [--branch B] [--pr N] [--experiments N]` is planned. Rules:\nR1. A Goal's body may end with a constraints block: it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body; it holds each of the three headings at least once; every non-empty line in it starts with one of them; a line of spaces, a tab or a lone carriage return is not empty (so invalid).\nR2. `hyp add goal` with a body that has a heading but no valid block succeeds and prints a warn line. A body with no heading adds silently.\nR3. `hyp prompt` on a Goal without a valid block exits 1 and names what is missing; bad arguments (e.g. --branch with a space) exit 2; no repository holding the store exits 3.\nR4. With a valid block it prints: the store path, actor A, the Goal id, skill name hyp-goal-run, stop clauses (runner owes nothing; N Results recorded under actor A, default 5, set by --experiments; cannot take the next step honestly), then the block's lines byte for byte; --branch and --pr each add one clause after the block that only narrows.\nR5. It writes nothing: no event, no file. It carries no owner answers or tree state.\nR6. Goals are never edited; a store may hold many Goals; logs from several actors/worktrees merge.\n\nWalk each rule through scenarios the designers did not draw, e.g.: two Goals in one store, one with a block and one without; a merged log from two actors where the Goal's add is on another branch not yet merged; a Goal body whose block text contains another tree's ids `tree:G-X`; CRLF or trailing-space bodies; a body with the headings appearing in the middle followed by prose; the actor already used for earlier Results (does \"N Results under actor\" count them?); store path when run from a worktree or a subdirectory; a Goal id that is an alias or unknown; stop count semantics. For each, say what the rules give and where they are ambiguous, contradictory or leave a state reachable that should be refused. Cite file:line for every fact about the code. Report at most 12 findings, most important first, each with scenario, outcome, and the source fact.",
 "run_in_background": true
}
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: aa7a5f4e467854f21 (internal ID - do not mention to user. Use SendMessage with to: 'aa7a5f4e467854f21', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-2-wt/85078edc-ac51-4204-aee9-42d61eec2054/tasks/aa7a5f4e467854f21.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 140,215p internal/cli/cli.go; grep -n \"Exit\\|exit\" internal/herr/*.go | head -20",
 "description": "Check CLI dispatch and exit codes for unknown refs"
}
```

Result:

```text
// Commands lists every command name Run dispatches, sorted.
func Commands() []string {
	names := []string{"version", "help"}
	for name := range commands {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}

// Run executes one hyp invocation and returns the process exit code: 0
// success, 1 refused, 2 usage error, 3 I/O or internal error. It is the only
// place a herr kind becomes an exit code.
func Run(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 {
		fmt.Fprint(stderr, usage)
		return 2
	}
	switch args[0] {
	case "version":
		fmt.Fprintln(stdout, versionLine())
		return 0
	case "help", "-h", "--help":
		fmt.Fprint(stdout, usage)
		return 0
	}
	name := args[0]
	if a, ok := aliases[name]; ok {
		name = a
	}
	cmd, ok := commands[name]
	if !ok {
		fmt.Fprint(stderr, usage)
		return 2
	}
	err := cmd(args[1:], stdout, stderr)
	if err == nil {
		return 0
	}
	if h, ok := errors.AsType[helpText](err); ok {
		fmt.Fprintln(stdout, string(h))
		return 0
	}
	fmt.Fprintln(stderr, err)
	// A recipe script's own exit code passes through unchanged (sandbox spec §1.3).
	if cf, ok := errors.AsType[*sandbox.CommandFailed](err); ok {
		return *cf.Row.Exit
	}
	switch herr.KindOf(err) {
	case herr.Refusal:
		return 1
	case herr.BadArguments:
		return 2
	default:
		return 3
	}
}

func cmdInit(args []string, stdout, stderr io.Writer) error {
	const use = "usage: hyp init"
	if len(args) == 1 && (args[0] == "-h" || args[0] == "--help") {
		return helpText(use)
	}
	if len(args) != 0 {
		return usageErr(use)
	}
	cwd, err := os.Getwd()
	if err != nil {
		return err
	}
	if err := store.Init(cwd); err != nil {
		return herr.Env("%v", err)
	}
	fmt.Fprintf(stdout, "initialised .tree in %s\n", cwd)
	return nil
}
2:// return an error of one kind; only cli.Run maps a kind to an exit code.
```

## Grill

The second reader is running on the prompt rules. I'll check each fact it reports against the source before I use it, then show the routing table.

## Owner, message 3

<task-notification>
<task-id>aa7a5f4e467854f21</task-id>
<tool-use-id>toolu_01NaSYcr3iHqDoDs7Phih6Qx</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-2-wt/85078edc-ac51-4204-aee9-42d61eec2054/tasks/aa7a5f4e467854f21.output</output-file>
<status>completed</status>
<summary>Agent "Second-reader review of prompt rules" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Findings, most important first. "Source" gives the code fact each one rests on. Nothing in `hyp prompt` exists yet, so the fixes are for the plan.

1. **Old Goals can never run, and the refusal has no fix.**
   - **Scenario:** a Goal added before this change has no block, or has a malformed one. Per R6 Goals are never edited.
   - **Outcome:** `hyp prompt` exits 1 for that Goal forever. Only `amend` revises a body, and it is for Hypotheses. The only remedy is a new Goal, and the old Goal's Results and Lines stay on the old id. R3 does not say the message should tell the owner to add a new Goal.
   - **Source:** `validateAmend` calls `f.need(ev.Node, event.KindHypothesis)` at internal/replay/validate.go:578. Lines come only with the Goal's add (internal/cli/write.go:264).

2. **"Starts with a heading" lets a prose line hijack the block.**
   - **Scenario:** the met-rule prose has a line like `Never: ship X is the aim...` before the real block. Or the headings appear mid-body, followed by prose.
   - **Outcome:**
     - R1 opens the block at the first such line, so every following prose line is invalid and the whole Goal is refused. The real block at the end never gets a chance.
     - R2 does warn here (heading present, block invalid). The warning comes after the Goal is already written and cannot be undone under R6.
     - R2's "has a heading" is also undefined for `  Never:` with leading spaces, or `x May change:` mid-line.
   - **Source:** a Goal body is free text up to MaxGoalBody = 2000 runes (internal/event/event.go:117; internal/replay/validate.go:83-87). `cmdAdd` only stores `*body` (internal/cli/write.go:287).

3. **"N Results recorded under actor A" is not scoped to the Goal and has no counter.**
   - **Scenario:**
     - (a) The store has two Goals, or actor A already recorded Results earlier. Do they count?
     - (b) A retracted Result.
     - (c) A Result whose add was rejected on a merged log.
   - **Outcome:**
     - R4 does not say whether the count is per Goal, per run or per store.
     - Nothing records when the run started, so a reused actor may already be at or over N and the stop clause is met at once.
     - Retracted Results still exist as nodes. Only the `Retracted` map and a status exclude them.
     - Results rejected at replay are absent from the projection.
     - No command gives the runner this count, and `hyp status G` is per Goal while the clause reads per actor.
   - **Source:** a Result Node carries `Actor: ev.Actor` and `Parent` (internal/replay/rules.go:79-83). `Retracted` is a separate map (rules.go:109-111; internal/replay/decision.go:133). `GoalOf` walks the parents (internal/replay/state.go:440-447). Actor is a filename, `&lt;actor&gt;.jsonl` (internal/store/store.go:211).

4. **Unknown, non-Goal, alias and display-id refs are not covered by R3, and an unmerged Goal gives exit 2, not 1 or 3.**
   - **Scenario:**
     - (a) The Goal's add is on a branch not yet merged.
     - (b) An unknown id.
     - (c) An id that names an Experiment.
     - (d) An alias or `g-abc123` in lower case.
   - **Outcome:**
     - `resolveRef` takes a ULID, then an alias, then a display id compared case-insensitively. It returns `BadArgs` for unknown or ambiguous, which is exit 2.
     - A Goal still on another branch is simply an unknown node.
     - R3 is silent on (a) to (c). `status` shows the pattern: a non-Goal gives `usageErr`, so exit 2.
     - Aliases are legal inputs, but the printed Goal id must be one form. Pick the display id or the ULID.
     - The same Goal can be reached by several spellings, so the prompt text varies.
   - **Source:** internal/cli/resolve.go:13-37. The kind check is at internal/cli/read.go:146-148. Exit-code mapping is at internal/cli/cli.go:188-194.

5. **Exit 3 versus other failures, and the store path.**
   - **Scenario:** run from a subdirectory, from a linked worktree, through a symlinked path, or in a nested directory that has its own `.tree`.
   - **Outcome:**
     - `FindRoot` walks up to the nearest ancestor holding `.tree/events` and returns an absolute path from `os.Getwd()`. Symlinks are not resolved.
     - `ErrNoStore` is a plain `errors.New`, so `KindOf` gives Environment, exit 3. That matches R3, but the exit 3 also covers a store that exists yet is unreadable.
     - A worktree has its own checked-out `.tree` because events are git-tracked. The printed path is the worktree's, not the main checkout's. A Goal added elsewhere is invisible there (finding 4).
     - The path may be relative-looking or differ between the owner's shell and the unattended run. R4 does not say whether it is a path to `.tree` or to the root.
   - **Source:** internal/store/store.go:46-62 and :30. `openEvents` is at internal/cli/read.go:45-58. docs/replay-spec.md:16-17 and :21 describe the store as git-tracked and found by walking up like `.git`.

6. **CRLF, lone `\r`, tabs, trailing space and "byte for byte".**
   - **Scenario:** a CRLF body, trailing spaces, or a blank line that is `\r`.
   - **Outcome:**
     - Splitting on `\n` leaves a `\r` at the end of every CRLF line. A heading line still passes ("starts with").
     - A CRLF blank line is a lone `\r`, which R1 calls invalid. So a pasted CRLF block with blank separators is rejected, though `\r` mid-line is fine.
     - A trailing-space-only line is invalid. A terminal newline is acceptable.
     - Byte-for-byte output means CR, tabs and trailing spaces reach the runner. The terminal may mangle them.
     - R1 and R4 do not say whether a final `\n` is added, or whether empty lines inside the block are printed.
   - **Source:** nothing in the code normalizes bodies. `textLimit` counts runes only (internal/replay/validate.go:905-914). The log keeps `Payload` as written, and `parseFile` trims only whole JSONL lines (internal/store/store.go:106).

7. **Block text that holds ids, including `tree:G-X`.**
   - **Scenario:** a block line mentions `tree:G-ABC123`, or a bare `G-ABC123` that is not a node.
   - **Outcome:**
     - The add-time cite check skips an id right after `:` or `-`. So `tree:G-X` and `containment:R-ABC123` pass at add (internal/cli/idcite.go:34-36).
     - A bare six-character id that is no node makes `add goal` fail with exit 2 (`BadArgs`, idcite.go:100). A block line naming a future or foreign id cannot be written unless it is prefixed.
     - `hyp prompt` should print such ids verbatim and must not resolve or validate them. R4 and R5 are silent.
     - Placeholder marks (`&lt;`, `{`, `…`, `...`, `XXX`) after `G-` also refuse (idcite.go:18-19). A block saying "Never: touch G-..." will fail the add.
   - **Source:** `body` is in `textFlags` (idcite.go:11). `checkCites` runs inside `w.run` (internal/cli/write.go:150).

8. **R2's warning path does not exist for a Goal today.**
   - **Scenario:** `hyp add goal` with a heading but no valid block.
   - **Outcome:**
     - Only Experiments produce `warns`, and they are printed only if no error (internal/cli/write.go:405-409). A Goal needs new code for the warning.
     - The Goal add also prints the `line` lines (write.go:394-398). The order of those lines and the warning needs fixing.
     - `Validate` and replay must not refuse a body for the block. The spec says CLI-only refusals exist so an old log cannot turn into rejections (docs/replay-spec.md:140). So R2 must be CLI-only and `hyp check` must not report it, or the old Goals in finding 1 would become errors.

9. **Branch and PR clauses are free text that may not only narrow.**
   - **Scenario:** `--branch` is `main`, or contains shell characters. `--pr` is `0` or `-1`. `--experiments 0` or negative.
   - **Outcome:**
     - R3 gives only the one example of a space. It does not say which branch syntax is valid (git ref rules), whether `--pr` must be positive, or whether `--experiments` may be 0.
     - `--experiments 0` is a clause already met at the start.
     - "Only narrows" is checked by nobody. A clause that comes after a block that says `May change: main` is not mechanically a narrowing.
     - The `parse` helper accepts flags in any position and a repeated flag silently takes the last value (internal/cli/write.go:43-62).
     - `--actor` for a read command is not handled by `writer.begin`. A read must re-do the `$HYP_ACTOR` fallback and `ValidActor` (internal/cli/write.go:96-104), or R3's exit 2 differs between commands.

10. **R5 versus the counts and the merged-log shape.**
    - **Scenario:** `hyp prompt` on a merged log where the actor's file has lines that did not parse.
    - **Outcome:**
      - `lineErrs` are skipped with only a warning on stderr in `status` (internal/cli/read.go:154-156). `prompt` might silently print a prompt from a partial log.
      - If the Goal add itself is one of the unparseable lines, the Goal is unknown, so exit 2 (finding 4) rather than a visible corruption.
      - R5 says "no owner answers or tree state", but the Goal id, actor and path are tree state of a kind. Define that as the only state.

11. **Many Goals, one with a block and one without.**
    - **Scenario:** two Goals in one store.
    - **Outcome:** each is read by its own Node body, so `prompt G1` works and `prompt G2` exits 1. Nothing is shared. The unscoped count (finding 3) and `checkpoint --goal`, which is required when there are more than one Goal (internal/cli/write.go:888), mean the prompt must tell the runner which Goal to checkpoint on. R4's stop-clause list does not say so. The runner's skill is named but its Goal argument is not.
    - **Source:** `st.Goals` is plural (internal/replay/state.go:317-320). `Node.Body` is set from the add payload (internal/replay/rules.go:49).

12. **Heading counts: three headings "at least once" versus a repeated heading and order.**
    - **Scenario:** a block with two `Never:` lines and no `Read only:`, or with the headings in any order.
    - **Outcome:**
      - R1 requires each heading to appear at least once, so the example is invalid.
      - R1 does not fix the order and does not say what follows a heading on the same line. `Never:` alone, with nothing after it, is a valid non-empty line that carries no constraint.
      - An empty `Never:` could be read as "nothing is forbidden". Say whether a heading must carry text.
      - Heading matching is case-sensitive and exact, so `Read-only:` or `May Change:` is a non-heading line and invalidates the block.</result>
<usage><subagent_tokens>66341</subagent_tokens><tool_uses>16</tool_uses><duration_ms>74303</duration_ms></usage>
</task-notification>

## Grill

The second reader's report arrived. It is agent output, not an answer from you. I'm checking its main facts against the source before I rely on them.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"MaxGoalBody\" -r internal | head -4; sed -n 10,40p internal/cli/idcite.go; sed -n 95,102p internal/cli/idcite.go; sed -n 10,40p internal/cli/resolve.go; sed -n 140,150p internal/cli/read.go; sed -n 884,890p internal/cli/write.go; sed -n 400,410p internal/cli/write.go",
 "description": "Verify second-reader facts against source"
}
```

Result:

```text
internal/cli/write.go:255:		maxBody = event.MaxGoalBody
internal/replay/validate.go:85:		maxBody = event.MaxGoalBody // a Goal's body holds the owner's met rule
internal/event/event.go:117:	MaxGoalBody     = 2000

// textFlags are the flags whose text a write command checks for cites.
var textFlags = []string{"why", "body", "observation", "summary", "limit", "line", "refers", "answer", "accept", "returns"}

// placeholders are the marks that make "X-" plus what follows a stand-in
// for an id that was never filled in.
var placeholders = []string{"<", "{", "…", "...", "XXX"}

// unresolvedCites names each id-shaped token in text that is no node of p
// (spec §5.1). A token starts at a word boundary with G-, Q-, H-, E- or R-
// and goes on with a placeholder mark, or with six Crockford-base32
// upper-case characters that end the word. A token right after ':' cites
// another store ("containment:R-ABC123") and is skipped.
func unresolvedCites(p replay.Projection, text string) []string {
	var out []string
	for i := 0; i+2 <= len(text); i++ {
		if !strings.ContainsRune("GQHER", rune(text[i])) || text[i+1] != '-' {
			continue
		}
		if i > 0 && (isWordByte(text[i-1]) || text[i-1] == ':' || text[i-1] == '-') {
			continue
		}
		rest := text[i+2:]
		if hasPlaceholder(rest) {
			word := text[i:]
			if n := strings.IndexAny(word, " \t\n"); n >= 0 {
				word = word[:n]
			}
			out = append(out, strings.TrimRight(word, ",;:)"))
			continue
		}
		}
	}
	if len(bad) == 0 {
		return nil
	}
	return herr.BadArgs("%s; cite an id this store holds, or write <name>:<id> for another store's", strings.Join(bad, "; "))
}

// resolveRef turns a user-typed reference into a node id (spec §2): an exact
// ULID first, then an alias, then a display id.
func resolveRef(p replay.Projection, ref string) (string, error) {
	if _, ok := p.Nodes[ref]; ok {
		return ref, nil
	}
	var byAlias, byDisplay []string
	for id, n := range p.Nodes {
		if n.Alias != "" && n.Alias == ref {
			byAlias = append(byAli
[... 1618 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"goal. prompt\\|/goal prompt\\|prompt generator\\|hyp prompt\" README.md docs skills AGENTS.md | cut -c1-220 | head -30; grep -n \"discovery\" skills/*/SKILL.md | head",
 "description": "Preflight scan of texts the change touches"
}
```

Result:

```text
README.md:166:1. **`/hyp-owner-grill`** questions you about the purpose until it is a Goal a run can finish. It looks up the facts itself, sends questions a run could test to DesignQuestions with rivals, derives the line
docs/adr/0017-the-ledger-is-the-stall-guard.md:20:- **A run's cap stays outside hyp.** A `/goal` prompt or a recipe states how many Experiments a run may take.
docs/how-hyp-improves-itself.md:30:  stated by the `/goal` prompt or the recipe, not by hyp.
skills/hyp-owner-grill/SKILL.md:35:| how the tree is run (process, review, order, cadence) | the Goal body, an Experiment's verdict rule, or the `/goal` prompt; never a line | only where it is a preference |
skills/hyp-owner-grill/SKILL.md:96:The Goal body ends with the run's constraints, one heading per line, so a run and a prompt generator can find them:
skills/hyp-owner-grill/SKILL.md:108:Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Neve
skills/hyp-goal-run/SKILL.md:3:description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner ca
skills/hyp-goal-run/SKILL.md:6:The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypo
skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you n
skills/hyp-goal-run/SKILL.md:68:Stop on the clauses the `/goal` prompt names. Stopping is not success.
skills/hyp-goal-run/SKILL.md:3:description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test be
[... 409 more characters]
```

## Grill

The second reader's findings match the source. I checked these:
- Unknown nodes, and nodes that aren't Goals, exit 2 (`resolve.go`, `read.go:146`).
- Goal bodies are capped at 2000 runes (`event.go:117`).
- A Goal body with a bare placeholder id such as `G-...` fails the add (`idcite.go`), so foreign ids need a `tree:` prefix.
- Only Experiments emit warnings today (`write.go:405`). The Goal-add warning is new code and must be CLI-only. The replay spec keeps old logs from turning into rejections.

## Second reader and preflight

**Looked up, so they go into lines and not to you:**
- An unknown, unmerged or non-Goal ref exits 2.
- The printed Goal id is the display id. The store path is the absolute directory holding `.tree`.
- Ids in a block print verbatim and are never resolved.
- A Goal added on a branch that isn't merged is unknown. The skill already says to run `hyp prompt` on the log the run starts from.

**Self-application:**
- `hyp prompt` feeds later runs and doesn't judge this Goal, so the lines don't judge themselves.
- The installed `hyp` can't warn on my own Goal add, so I will check the block rules by hand before the add.

**Preflight:**
- ADR 0017 and `how-hyp-improves-itself.md:30` say a run's cap is stated outside hyp. A flag keeps that true, because nothing is stored. I'd leave both texts alone.
- **Review round missing:** the repo's `hyp-goal-run` has no build-Goal review round. The old hand-written prompt carried it, and `hyp prompt` will not. Without a skill change the practice is lost.

## Round 2

**7. Reused actor.**
- **Problem:** the clause counts Results under the actor in this Goal. A reused actor may already be at N, so the stop is met at once.
- **Option A:** `hyp prompt` refuses (exit 1) an actor that already has a Result under this Goal. Replay derives this, so nothing is stored.
- **Option B:** no check. A crashed run can resume under its actor, but a reused actor stops early.
- **Recommendation:** A.

**8. Must a heading carry text?** For example, is `Never:` alone valid?
- **Option A:** valid. It is a non-empty line that starts with a heading.
- **Option B:** invalid. A heading with nothing after it carries no constraint, and a reader could take an empty `Never:` as "nothing forbidden".
- **Recommendation:** B. Headings are also case-sensitive and may come in any order.

**9. Docs and ADRs.** May the run add a new ADR (0022) and update the glossary, specs, README and install text? Existing ADRs stay unedited, as earlier runs referred such edits to you.
- **Recommendation:** yes to the first, no to editing existing ADRs.

**10. Branch and PR.** Does the run commit on `feat/goal-prompt`, open a PR, and never merge it?
- **Recommendation:** yes.

**11. Review-round text.** May the run add to `hyp-goal-run` the build-Goal practice that the old prompt carried?
- **What it covers:** one fresh-subagent review per candidate commit. It sorts findings into a limit, a reading, a fix commit or an out-of-Goal finding.
- **Recommendation:** yes. It is a skill text this change makes wrong, so it goes under May change and gets no line.

## Routing table (draft)

| # | Question | Goes to | Separating scenario or invariant |
|---|---|---|---|
| 1 | One Goal | Goal title and body. Skill descriptions go under `May change:`. | none |
| 2 | Run-6 match | Goal body. The synthetic fixture's exceptions are the actor, owner answer, branch, PR, commit and the run's plan. | none |
| 3 | Stop clauses | Line 1, which carries its own content | none |
| 4 | Command or skill | DesignQuestion "Where is the prompt produced?", rivals "hyp prompt reads the log" and "a skill writes it from `hyp show`", recorded as decided with this session as evidence | one Goal body with a CRLF block given to both writers |
| 5 | Block validity | Lines 3 and 4 | |
| 6 | Content and writes | Lines 1, 2 and 5 | |
| 7 | Reused actor | Line 8 (if you pick A) | |
| 8 | Heading text | Line 3 (if you pick B) | |
| 9–11 | Docs, PR, review text | Goal body `May change:` and `Never:` | |
| — | Second-reader facts | Lines 6 and 7, and body text | |

**Draft lines:**
- **1 [P]** For a Goal with a valid block, `hyp prompt G --actor A` exits 0. It prints these in order, each ending in `\n`:
  - the store's absolute root;
  - A;
  - the Goal's display id;
  - `hyp-goal-run`;
  - the stop clauses: the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; or it cannot take the next step honestly and says why;
  - the block's lines byte for byte.

  With no flags there is no branch or PR clause.
- **2 [P]** `--experiments N` sets N. `--branch B` and `--pr N` each add one clause after the block, and each only narrows. The output is the same for any flag order.
- **3 [N]** A block runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end. It is valid only if all three appear, each heading is case-sensitive and carries text, and every non-empty line starts with one. A line of spaces, a tab or a lone `\r` is not empty. On a body with no valid block, `hyp prompt` exits 1, names the missing heading or the bad line, and prints no prompt on stdout.
- **4 [P]** `hyp add goal` with a heading but no valid block succeeds and warns on stderr, naming the rule. A body with no heading adds silently. Replay and `hyp check` report no error for either.
- **5 [N]** `hyp prompt` writes nothing: `.tree` is byte-identical before and after. It carries no owner answer and no tree state beyond the Goal's id and body.
- **6 [N]** These exit 2:
  - a missing or invalid actor;
  - a branch git refuses, such as one with a space;
  - `--pr` or `--experiments` below 1;
  - a repeated flag;
  - an unknown node;
  - a node that is not a Goal.

  No `.tree` found exits 3.
- **7 [N]** In a two-Goal store, `hyp prompt` reads only the named Goal's body. The other Goal's block never appears, and a Goal without a block is refused while the other prints.
- **8 [N]** `hyp prompt` exits 1 for an actor that already has a Result under that Goal.

## Next step
Please confirm each row, or move it, and answer 7 to 11. Then I'll show the exact commands, with owner-standin on the Goal add, for your approval one by one. Nothing is written yet.

## Owner, message 4

7 A. 8 B. 9 yes to a new ADR 0022 and the doc updates, no to editing existing ADRs. 10 yes, the run commits on `feat/goal-prompt`, opens a PR and never merges it. 11 yes, the review-round text goes under May change.

I confirm the routing table and the draft lines as they stand, with no row to move. Nothing is pushed or posted in this session.

I approve the exact owner commands one at a time when you show them, with owner-standin as the actor.

## Grill

Thanks, I have your answers and the confirmation. I'll run the runner-side `hyp init` first, then draft the Goal's body and check it by hand, since the installed `hyp` can't warn on a bad block.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p investigations/goal-prompt && cd investigations/goal-prompt && HYP_ACTOR=claude-grill-113 hyp init && ls -a . && cd ../.. && git status --short",
 "description": "Create the new tree store"
}
```

Result:

```text
initialised .tree in /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/goal-prompt
.
..
.tree
?? investigations/goal-prompt/
```

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > body.txt <<'EOF'\nBuilds `hyp prompt` for issue #113: hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run's constraints under May change:, Read only: and Never:. The candidate is a hyp commit. go test is the instrument, and skillcheck decides the skill text.\nDone means, by the owner on 2026-11-11: for a Goal with a valid block, the prompt carries every clause the hand-written run-5 prompt had, except the actor, the owner's answer, the branch, the PR, the commit and the run's own plan. A test pins this on a synthetic Goal under internal/cli/testdata. checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it and no test compares it.\nThe session settled that hyp prompt is a command that reads the log, and that no skill writes the prompt.\nMay change: the hyp repository, by commits on the local branch feat/goal-prompt and then a PR that the run opens and never merges\nMay change: cmd, internal, the docs and a new ADR 0022\nMay change: skills, for the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal, the hand-off of hyp-owner-grill, and a build-Goal review round in hyp-goal-run\nRead only: the existing ADRs, every other tree under investigations, and recipes\nNever: replace the installed hyp binary\nNever: edit an existing ADR or a Goal\nNever: merge a PR or close #113 (write Part of #113)\nEOF\nwc -m body.txt; grep -vcE '^$' body.txt; awk 'NR>3' body.txt | grep -vcE '^(May change:|Read only:|Never:)'; grep -nE '(^|[^A-Za-z0-9:-])[GQHER]-' body.txt\ncat > l1.txt <<'EOF'\n[P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each ending in a newline and in this order: the store's absolute root, A, the Goal's display id, `hyp-goal-run`, the stop clauses (the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; it cannot take the next step honestly and says why), then the block's lines byte for byte. With no flag there is no branch or PR clause.\nEOF\ncat > l2.txt <<'EOF'\n[P] `--experiments N` sets N. `--branch B` and `--pr N` each add exactly one clause after the block, and each clause only narrows what the block allows. The output is the same for any order of the flags.\nEOF\ncat > l3.txt <<'EOF'\n[N] A body's block runs from its first line that starts with `May change:`, `Read only:` or `Never:` to the end, and is valid only if all three headings appear, each has text after it, and every non-empty line starts with one (headings are case-sensitive; a line of spaces, a tab or a lone carriage return is not empty). For a body with no valid block, `hyp prompt` exits 1, names the missing heading or the bad line, and prints no prompt on stdout.\nEOF\ncat > l4.txt <<'EOF'\n[P] `hyp add goal` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no heading adds without a warning. Replay and `hyp ch
[... 948 more characters]
```

Result:

```text
    1348 body.txt
10
0
     462 l1.txt
     204 l2.txt
     450 l3.txt
     219 l4.txt
     161 l5.txt
     241 l6.txt
     203 l7.txt
      97 l8.txt
    2037 total
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-2/wt
```

## Grill

I created the store with `hyp init` under `claude-grill-113`, in `investigations/goal-prompt/`. Nothing else is written.

I checked the Goal body by hand because the installed `hyp` has no block warning:
- **Length:** 1348 of 2000 runes.
- **Block start:** the first heading line is line 4.
- **Block shape:** lines 4 to 10 each start with a heading, all three headings appear, and there are no blank or whitespace lines.
- **Ids:** the only node id is written `checkpoint-store:G-D1CH1R`, so the cite check skips it.
- **Lines:** 8 lines, each under 600 runes. The longest is 462.

## Owner write 1 of 1: the Goal

This is the exact command. It runs under `owner-standin`, and I'll run it only after you approve it.

```bash
cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \
  --title "hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run's constraints" \
  --body $'Builds `hyp prompt` for issue #113: hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run\'s constraints under May change:, Read only: and Never:. The candidate is a hyp commit. go test is the instrument, and skillcheck decides the skill text.\nDone means, by the owner on 2026-11-11: for a Goal with a valid block, the prompt carries every clause the hand-written run-5 prompt had, except the actor, the owner\'s answer, the branch, the PR, the commit and the run\'s own plan. A test pins this on a synthetic Goal under internal/cli/testdata. checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it and no test compares it.\nThe session settled that hyp prompt is a command that reads the log, and that no skill writes the prompt.\nMay change: the hyp repository, by commits on the local branch feat/goal-prompt and then a PR that the run opens and never merges\nMay change: cmd, internal, the docs and a new ADR 0022\nMay change: skills, for the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal, the hand-off of hyp-owner-grill, and a build-Goal review round in hyp-goal-run\nRead only: the existing ADRs, every other tree under investigations, and recipes\nNever: replace the installed hyp binary\nNever: edit an existing ADR or a Goal\nNever: merge a PR or close #113 (write Part of #113)' \
  --line "[P] For a Goal with a valid constraints block, \`hyp prompt G --actor A\` exits 0 and prints, each ending in a newline and in this order: the store's absolute root, A, the Goal's display id, \`hyp-goal-run\`, the stop clauses (the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; it cannot take the next step honestly and says why), then the block's lines byte for byte. With no flag there is no branch or PR clause." \
  --line "[P] \`--experiments N\` sets N. \`--branch B\` and \`--pr N\` each add exactly one clause after the block, and each clause only narrows what the block allows. The output is the same for any order of the flags." \
  --line "[N] A body's block runs from its first line that starts with \`May change:\`, \`Read only:\` or \`Never:\` to the end, and is valid only if all three headings appear, each has text after it, and every non-empty line starts with one (headings are case-sensitive; a line of spaces, a tab or a lone carriage return is not empty). For a body with no valid block, \`hyp prompt\` exits 1, names the missing heading or the bad line, and prints no prompt on stdout." \
  --line "[P] \`hyp add goal\` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no heading adds without a warning. Replay and \`hyp check\` report no error for either Goal." \
  --line "[N] \`hyp prompt\` writes nothing: the bytes of \`.tree\` are the same before and after. It carries no owner answer and no tree state beyond the Goal's id and body." \
  --line "[N] These exit 2: a missing or invalid actor, a branch that git refuses (such as one with a space), \`--pr\` or \`--experiments\` below 1, a repeated flag, an unknown node, and a node that is not a Goal. A run where no \`.tree\` is found exits 3." \
  --line "[N] In a store with two Goals, \`hyp prompt\` reads only the named Goal's body: the other Goal's block never appears in the output, and a Goal without a valid block is refused while the other Goal prints." \
  --line "[N] \`hyp prompt\` exits 1, naming the actor, when the actor already has a Result under that Goal." \
  --why "Owner flow, issue #113: the run's constraints get a recorded place and hyp prints the /goal prompt from the log. Written after the grill of 2026-11-12, with the owner's approval of this command."
```

**Do you approve this exact command to run under `owner-standin`?**

After it runs, the runner writes follow:
1. The DesignQuestion "Where is the prompt produced?" under the Goal, with two rivals: "`hyp prompt` reads the log" and "a skill writes the prompt from `hyp show`".
2. Under the chosen rival, an Experiment with `--recipe none --rubric reading --target <the other rival>`. Its body states the CRLF scenario as the rule. A `hyp record` follows, citing a scenarios file under `.tree/evidence/<E-id>/`.
3. The select, once OWED owes it.
4. `hyp link` to #113 as an `external_ref`.
5. `hyp status`, `hyp check` and a checkpoint "end of grill".

I'll commit nothing without being asked, and push nothing, as you said.

## Owner, message 5

Approved. Run the command exactly as shown, under owner-standin.

### Tool Bash (cwd: .)

```text
{
 "command": "cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run's constraints\" \\\n  --body $'Builds `hyp prompt` for issue #113: hyp prints the /goal prompt for any Goal from the log, and a Goal body carries the run\\'s constraints under May change:, Read only: and Never:. The candidate is a hyp commit. go test is the instrument, and skillcheck decides the skill text.\\nDone means, by the owner on 2026-11-11: for a Goal with a valid block, the prompt carries every clause the hand-written run-5 prompt had, except the actor, the owner\\'s answer, the branch, the PR, the commit and the run\\'s own plan. A test pins this on a synthetic Goal under internal/cli/testdata. checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it and no test compares it.\\nThe session settled that hyp prompt is a command that reads the log, and that no skill writes the prompt.\\nMay change: the hyp repository, by commits on the local branch feat/goal-prompt and then a PR that the run opens and never merges\\nMay change: cmd, internal, the docs and a new ADR 0022\\nMay change: skills, for the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal, the hand-off of hyp-owner-grill, and a build-Goal review round in hyp-goal-run\\nRead only: the existing ADRs, every other tree under investigations, and recipes\\nNever: replace the installed hyp binary\\nNever: edit an existing ADR or a Goal\\nNever: merge a PR or close #113 (write Part of #113)' \\\n  --line \"[P] For a Goal with a valid constraints block, \\`hyp prompt G --actor A\\` exits 0 and prints, each ending in a newline and in this order: the store's absolute root, A, the Goal's display id, \\`hyp-goal-run\\`, the stop clauses (the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; it cannot take the next step honestly and says why), then the block's lines byte for byte. With no flag there is no branch or PR clause.\" \\\n  --line \"[P] \\`--experiments N\\` sets N. \\`--branch B\\` and \\`--pr N\\` each add exactly one clause after the block, and each clause only narrows what the block allows. The output is the same for any order of the flags.\" \\\n  --line \"[N] A body's block runs from its first line that starts with \\`May change:\\`, \\`Read only:\\` or \\`Never:\\` to the end, and is valid only if all three headings appear, each has text after it, and every non-empty line starts with one (headings are case-sensitive; a line of spaces, a tab or a lone carriage return is not empty). For a body with no valid block, \\`hyp prompt\\` exits 1, names the missing heading or the bad line, and prints no prompt on stdout.\" \\\n  --line \"[P] \\`hyp add goal\\` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no heading adds without a warning. Replay and \\`hyp check
[... 1117 more characters]
```

Result:

```text
G-VRG6SA node.add
  line G-VRG6SA#1: [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each ending in a newline and in this order: the store's absolute root, A, the Goal's display id, `hyp-goal-run`, the stop clauses (the runner owes nothing; N non-retracted Results recorded under A in this Goal, N being 5 by default; it cannot take the next step honestly and says why), then the block's lines byte for byte. With no flag there is no branch or PR clause.
  line G-VRG6SA#2: [P] `--experiments N` sets N. `--branch B` and `--pr N` each add exactly one clause after the block, and each clause only narrows what the block allows. The output is the same for any order of the flags.
  line G-VRG6SA#3: [N] A body's block runs from its first line that starts with `May change:`, `Read only:` or `Never:` to the end, and is valid only if all three headings appear, each has text after it, and every non-empty line starts with one (headings are case-sensitive; a line of spaces, a tab or a lone carriage return is not empty). For a body with no valid block, `hyp prompt` exits 1, names the missing heading or the bad line, and prints no prompt on stdout.
  line G-VRG6SA#4: [P] `hyp add goal` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no heading adds without a warning. Replay and `hyp check` report no error for either Goal.
  line G-VRG6SA#5: [N] `hyp prompt` writes nothing: the bytes of `.tree` are the same before and after. It carries no owner answer and no tree state beyond the Goal's id and body.
  line G-VRG6SA#6: [N] These exit 2: a missing or invalid actor, a branch that git refuses (such as one with a space), `--pr` or `--experiments` below 1, a repeated flag, an unknown node, and a node that is not a Goal. A run where no `.tree` is found exits 3.
  line G-VRG6SA#7: [N] In a store with two Goals, `hyp prompt` reads only the named Goal's body: the other Goal's block never appea
[... 213 more characters]
```

## Grill

The Goal is written as G-VRG6SA. Next come the runner writes: the DesignQuestion and its two rivals.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113; G=G-VRG6SA\nhyp add dq --parent $G --alias q-where --title \"Where is the /goal prompt produced?\" --body $'Rivals: a hyp command that reads the log, or a skill that writes the prompt from hyp show.\\nScenario that separates them: one Goal whose body block has CRLF line ends, given to both writers. The command prints the block byte for byte or refuses it with exit 1; a skill-written prompt can drift from the body and has no exit code to refuse with.' --why \"Contested question of the grill for issue #113; the owner settled it in the session.\" && \nhyp add hyp --parent q-where --alias h-command --title \"hyp prompt is a command that reads the log\" --body \"hyp prompt G --actor A prints the prompt from the Goal's body and the log, writes nothing, and exits 1 when the body has no valid constraints block.\" --why \"Rival 1 of Q-where.\" &&\nhyp add hyp --parent q-where --alias h-skill --title \"A skill writes the prompt from hyp show\" --body \"The owner-grill skill reads hyp show for the Goal and writes the prompt text itself, copying the constraints lines by hand.\" --why \"Rival 2 of Q-where.\"; hyp add --help 2>&1 | head -5",
 "description": "Add the DesignQuestion and its two rival Hypotheses"
}
```

Result:

```text
Q-H4QXXH node.add
H-ATB7RM node.add
H-JWWPMV node.add
usage: hyp add goal|dq|hyp|exp [flags]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp add exp --parent h-command --alias e-where --recipe none --rubric reading --target h-skill --title \"Settle Q-where by the session's scenarios\" --body $'Rule, fixed before the record: no run is made; the owner settled this in the session of 2026-11-12 and the session is the evidence. Scenarios in the evidence file: (1) a Goal body whose block has CRLF line ends; (2) a Goal body whose block has a bad line; (3) an actor that already has a Result. The command side gives each scenario a fixed output or an exit code. The skill side has no exit code and copies text by hand. H-command is supported when the command side gives a defined result in all three; H-skill is contradicted when the skill side gives none in at least one.' --why \"Records the owner's session decision as an argued, not tested, answer.\" 2>&1 | tee $TMPDIR/e.out; E=$(grep -o 'E-[A-Z0-9]\\{6\\}' $TMPDIR/e.out | head -1); echo $E; mkdir -p .tree/evidence/$E && cat > .tree/evidence/$E/scenarios.md <<'EOF'\n# Scenarios for Q-where, argued in the session of 2026-11-12 (not run)\n\n1. A Goal body whose constraints block has CRLF line ends and blank separator lines. Command: a blank CRLF line is a lone carriage return, so the block is invalid and `hyp prompt` exits 1 and names the line. Skill: the writer copies the text by hand and may normalise it; nothing refuses.\n2. A Goal body whose block has a line that starts with no heading. Command: exit 1 and names the line. Skill: no exit code; the prompt can carry the line or drop it.\n3. A runner actor that already has a Result under the Goal. Command: exit 1 naming the actor, derived from the log. Skill: it must recount by hand from hyp show.\n\nOwner decision, 2026-11-12: hyp prompt is a command that reads the log; no test is needed.\nEOF\nls .tree/evidence/$E",
 "description": "Add the Experiment and write the scenarios evidence file"
}
```

Result:

```text
E-A5JJZQ node.add
E-A5JJZQ
scenarios.md
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113; E=E-A5JJZQ\nhyp record --exp $E --verdict h-command=supports --verdict h-skill=contradicts --observation \"The owner settled Q-where in the session of 2026-11-12: hyp prompt is a command that reads the log. The three scenarios in the evidence file give the command a defined result each (output or exit 1) and the skill none. Argued, not run.\" --evidence .tree/evidence/$E/scenarios.md --why \"Records the session decision under the Experiment's pre-registered rule.\" 2>&1 | tail -8; hyp status G-VRG6SA 2>&1 | sed -n 1,40p",
 "description": "Record the Result and read the ledger"
}
```

Result:

```text
R-6HQ2BH result.record
GOAL G-VRG6SA hyp prints the /goal prompt for any Goal from the log, and … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-VRG6SA#1 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each endin…
  G-VRG6SA#2 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [P] `--experiments N` sets N. `--branch B` and `--pr N` each add exactly one clause after the block, a…
  G-VRG6SA#3 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [N] A body's block runs from its first line that starts with `May change:`, `Read only:` or `Never:` t…
  G-VRG6SA#4 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [P] `hyp add goal` with a body that has a heading but no valid block succeeds and warns on stderr, nam…
  G-VRG6SA#5 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [N] `hyp prompt` writes nothing: the bytes of `.tree` are the same before and after. It carries no own…
  G-VRG6SA#6 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [N] These exit 2: a missing or invalid actor, a branch that git refuses (such as one with a space), `-…
  G-VRG6SA#7 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [N] In a store with two Goals, `hyp prompt` reads only the named Goal's body: the other Goal's block n…
  G-VRG6SA#8 uncovered on Q-H4QXXH: H-ATB7RM (h-command) — [N] `hyp prompt` exits 1, naming the actor, when the actor already has a Result under that Goal.
FRONTIER
  Q-H4QXXH open/supported_unselected Where is the /goal prompt produced?
    candidates (1): H-ATB7RM (h-command)
    H-ATB7RM (h-command) supported hyp prompt is a command that reads the log
    H-JWWPMV (h-skill) contradicted A skill writes the prompt from hyp show
STALE
OWED
  runner: Q-H4QXXH limit G-VRG6SA#1 [owed] — [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0…
  runner: Q-H4QXXH limit G-VRG6SA#2 [owed] — [P] `--experiments N` sets N. `--branch B` and `--pr N` eac
[... 949 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113; hyp select Q-H4QXXH 2>&1 | head; hyp check 2>&1 | tail -3; grep -n \"top-level\\|top level\" ../../docs/replay-spec.md | head",
 "description": "Ask hyp what select is owed and check the spec on top-level lines"
}
```

Result:

```text
suggested: none (Q-H4QXXH (q-where) owes the runner first: limit G-VRG6SA#1)
check: 0 errors, 0 warnings
136:CLI error text is `refused: <reason>` on stderr, with exit code 1. A help request (`-h` or `--help` on any command, and `help` at the top level and on a group command) prints usage on stdout and exits 0 at every level, while a missing or unknown subcommand exits 2.
233:- `complete` when the Goal has at least one live DQ as a child and every live DQ whose parent is the Goal is `answered`; a Goal with no live top-level DQ is `in_progress`, because nothing has been asked yet. It reads no line: a line of the Goal is owed on each top-level question as a `limit` item (§8.1) and the select waits on it, and the owner's "met" is a complete Goal with no line under OWED. hyp stores no word for it. Sub-questions are covered through the `answered` rule of §7.4, so sub-questions under rejected rival hypotheses never block the Goal. (This clarifies the advice's "nothing stale" clause: stale on the selected path already prevents `answered`.)
344:Lines **[added]**: a line of the Goal (§4 lines) is a limit with the Goal for its Result. It has a hand and closers as a limit does, and is neither open nor closed: whether its case ran is per candidate, its coverage. `Open()` of a line means only that its hand is not an accept. **Coverage**: a Result covers a Hypothesis when it has a verdict on it, or when one of the Result's run subjects holds the Hypothesis's whole subject (§4 subject; `Projection.Covers`). Coverage says the bytes ran, not that the claim held. A line is **covered** on a question when a closer of the line covers each of its candidates, and **uncovered** there otherwise. An unaccepted, unreferred line is owed on each live top-level question where it is uncovered (`Projection.Uncovered`), so a new candidate is owed the line again. Coverage is not a state of the move table: it changes when a candidate is added or falls, with no move on the key, and no move reads it. `
[... 5698 more characters]
```

## Grill

The grill is not finished, and the fault is in the routing I proposed. You confirmed that table, but it has a defect I should have caught before any write.

## What went wrong

I put the "where is the prompt produced" DesignQuestion (Q-H4QXXH) under G-VRG6SA, the build Goal. Its rivals are designs, not commits. The spec (§8.1) owes every line of a Goal on every live top-level question with a candidate. So `hyp status` now owes the runner eight limits, G-VRG6SA#1 to #8, on a question whose candidate has no commit for any run to test. `hyp select Q-H4QXXH` also says `suggested: none`, because the limits come first. The select is not owed, and no run can ever clear those eight limits.

This is the grid problem in the skill: a design record and the build of that design are two Goals. I should have caught it before the routing table.

## State on disk

- **Written:**
  - G-VRG6SA, owner-standin, which you approved.
  - The DesignQuestion Q-H4QXXH.
  - Rivals H-ATB7RM (`h-command`, supported) and H-JWWPMV (`h-skill`, contradicted).
  - Experiment E-A5JJZQ and Result R-6HQ2BH, with the scenarios file in `.tree/evidence/E-A5JJZQ/`.
- **Integrity:** `hyp check` reports 0 errors and 0 warnings.
- **Not written:** no link, no checkpoint, no select and no commit. Events are append-only, so the DesignQuestion can only be withdrawn, not removed.

## Proposed fix

Two Goals, as the skill says:

1. **Keep G-VRG6SA as the build Goal.** The runner withdraws Q-H4QXXH from it. This is a runner write, `hyp withdraw`, and the Result stays in history. The runner adds a top-level question "Which hyp commit implements `hyp prompt`?" with no rival yet, as in G-D1CH1R. A run adds the commit Hypotheses. The eight lines then attach to a question that has commits.
2. **Add a design-record Goal**, with no lines, holding the same DesignQuestion, the same two rivals and the session evidence. That needs a second owner write: a new `hyp add goal` under `owner-standin`, which I would show you for approval. G-VRG6SA's body already says the session settled the command, and I would link the two Goals with `external_ref`.

The cheaper alternative is to leave things as they are. A run would then refer all eight lines and wait for you to accept each, which is eight owner acts for no gain, so I don't recommend it.

**Do you want the two-Goal fix?** If yes, I'll withdraw Q-H4QXXH and add the commit question now. Then I'll show you the exact `hyp add goal` command for the design record, wait for your approval, and continue to the link, the checkpoint and the hand-off.

The hand-off needs a note too. `hyp prompt` doesn't exist yet, so the `/goal` prompt for the build run can only be hand-written from the Goal body, until the first run builds the command.

## Owner, message 6

Yes, do the two-Goal fix. This is a concrete defect in my confirmed routing, so I accept your correction. Withdraw Q-H4QXXH, add the commit question, and then show me the exact `hyp add goal` command for the design record under owner-standin. I'll approve it then.

I have no position on the hand-off note. Go with your recommendation.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/goal-prompt && export HYP_ACTOR=claude-grill-113\nhyp withdraw Q-H4QXXH --why \"Misplaced: a design settled by argument has no commit for the build Goal's lines to run on. It moves to its own design-record Goal; this Goal asks which commit implements hyp prompt.\" 2>&1 | tail -3\nhyp add dq --parent G-VRG6SA --alias q-commit --title \"Which hyp commit implements hyp prompt?\" --body \"The candidate is a hyp commit on feat/goal-prompt: the prompt command, the warning on hyp add goal, and the skill and doc text the Goal's May change names. A run adds each commit as a Hypothesis whose subject is the base plus that commit's change; a fix is a new Hypothesis that supersedes it. Each Experiment's verdict rule asks for red on base and green on the candidate.\" --why \"The build Goal's top-level question; its rivals are the commits a run adds.\" 2>&1 | tail -2\nhyp status G-VRG6SA 2>&1 | sed -n 1,30p; hyp check | tail -2",
 "description": "Withdraw misplaced question and add the commit question"
}
```

Result:

```text
Q-H4QXXH dq.withdraw
Q-JSM74G node.add
GOAL G-VRG6SA hyp prints the /goal prompt for any Goal from the log, and … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 1, stale 0)
LINES
  G-VRG6SA#1 no candidate — [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each ending in a newline and in this orde…
  G-VRG6SA#2 no candidate — [P] `--experiments N` sets N. `--branch B` and `--pr N` each add exactly one clause after the block, and each clause only narrows wha…
  G-VRG6SA#3 no candidate — [N] A body's block runs from its first line that starts with `May change:`, `Read only:` or `Never:` to the end, and is valid only if…
  G-VRG6SA#4 no candidate — [P] `hyp add goal` with a body that has a heading but no valid block succeeds and warns on stderr, naming the rule; a body with no he…
  G-VRG6SA#5 no candidate — [N] `hyp prompt` writes nothing: the bytes of `.tree` are the same before and after. It carries no owner answer and no tree state bey…
  G-VRG6SA#6 no candidate — [N] These exit 2: a missing or invalid actor, a branch that git refuses (such as one with a space), `--pr` or `--experiments` below 1…
  G-VRG6SA#7 no candidate — [N] In a store with two Goals, `hyp prompt` reads only the named Goal's body: the other Goal's block never appears in the output, and…
  G-VRG6SA#8 no candidate — [N] `hyp prompt` exits 1, naming the actor, when the actor already has a Result under that Goal.
FRONTIER
  Q-JSM74G open/no_candidates Which hyp commit implements hyp prompt?
STALE
OWED
  runner: Q-JSM74G add rival Hypotheses
AWAITING OWNER
RECENT
  2026-11-12 14:12 claude-grill-113 dq.withdraw Q-H4QXXH — Misplaced: a design settled by argument has no commit for the build Goal's line…
  2026-11-12 14:12 claude-grill-113 result.record R-6HQ2BH — Records the session decision under the Experiment's pre-registered rule.
CHECKPOINT none
check: 0 errors, 0 warnings
```

## Grill

I withdrew Q-H4QXXH and added the commit question Q-JSM74G (`q-commit`). `hyp check` reports 0 errors. `hyp status` now shows the eight lines as `no candidate`, and OWED asks the runner only to `add rival Hypotheses` on Q-JSM74G. A run can finish that work.

## Owner write: the design-record Goal

This is the exact command, to run under `owner-standin` after you approve it.

I gave the Goal no lines, because no run checks its candidates. I gave it no constraints block either, because no run uses it, and a body with no heading adds silently.

```bash
cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \
  --alias goal-record \
  --title "Design record: hyp prompt is a command that reads the log" \
  --body $'Records the one design question of issue #113 that the owner settled by argument in the grill session of 2026-11-12: where the /goal prompt is produced. The rivals are a hyp command that reads the log and a skill that writes the prompt from hyp show. The session is the evidence; no run was made. It has no lines: no run checks its candidates, and its cases are the lines of the build Goal G-VRG6SA, which builds this design.' \
  --why "Owner flow, issue #113: a design settled by argument is its own Goal, apart from the build Goal G-VRG6SA. Written after the grill of 2026-11-12, with the owner's approval of this command."
```

**Do you approve this exact command to run under `owner-standin`?**

After it runs, I'll make these runner writes:
1. Re-add the DesignQuestion with its two rivals under the new Goal, since nodes can't move.
2. Add the Experiment, with a copy of the scenarios file under its own id.
3. Record the Result and the select.
4. Link the design-record Goal to G-VRG6SA, and G-VRG6SA to issue #113, as `external_ref`.
5. Run `hyp status` and `hyp check` for each Goal.
6. Write the `end of grill` checkpoint on each Goal.

On the hand-off, my recommendation stands. `hyp prompt` doesn't exist yet, so I'll give you a hand-written `/goal` prompt for the first build run, copied from the Goal body's block and the stop clauses. The run builds the command, and later runs use it.

## Owner, message 7

Approved. Run the command exactly as shown, under owner-standin.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt/investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --alias goal-record \\\n  --title \"Design record: hyp prompt is a command that reads the log\" \\\n  --body $'Records the one design question of issue #113 that the owner settled by argument in the grill session of 2026-11-12: where the /goal prompt is produced. The rivals are a hyp command that reads the log and a skill that writes the prompt from hyp show. The session is the evidence; no run was made. It has no lines: no run checks its candidates, and its cases are the lines of the build Goal G-VRG6SA, which builds this design.' \\\n  --why \"Owner flow, issue #113: a design settled by argument is its own Goal, apart from the build Goal G-VRG6SA. Written after the grill of 2026-11-12, with the owner's approval of this command.\"; echo \"exit $?\"",
 "description": "Add the design-record Goal as owner-standin"
}
```

Result:

```text
G-V6KQR7 node.add
exit 0
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113; D=G-V6KQR7\nhyp add dq --parent $D --alias q-produced --title \"Where is the /goal prompt produced?\" --body $'Rivals: a hyp command that reads the log, or a skill that writes the prompt from hyp show.\\nScenario that separates them: one Goal whose body block has CRLF line ends, given to both writers. The command prints the block byte for byte or refuses it with exit 1; a skill-written prompt can drift from the body and has no exit code to refuse with.' --why \"Contested question of the grill for issue #113; the owner settled it in the session.\" &&\nhyp add hyp --parent q-produced --alias h-cmd --title \"hyp prompt is a command that reads the log\" --body \"hyp prompt G --actor A prints the prompt from the Goal's body and the log, writes nothing, and exits 1 when the body has no valid constraints block.\" --why \"Rival 1 of q-produced.\" &&\nhyp add hyp --parent q-produced --alias h-skl --title \"A skill writes the prompt from hyp show\" --body \"The owner-grill skill reads hyp show for the Goal and writes the prompt text itself, copying the constraints lines by hand.\" --why \"Rival 2 of q-produced.\" &&\nhyp add exp --parent h-cmd --alias e-produced --recipe none --rubric reading --target h-skl --title \"Settle q-produced by the session's scenarios\" --body $'Rule, fixed before the record: no run is made; the owner settled this in the session of 2026-11-12 and the session is the evidence. Scenarios in the evidence file: (1) a Goal body whose block has CRLF line ends; (2) a Goal body whose block has a bad line; (3) an actor that already has a Result. The command side gives each scenario a fixed output or an exit code. The skill side has no exit code and copies text by hand. h-cmd is supported when the command side gives a defined result in all three; h-skl is contradicted when the skill side gives none in at least one.' --why \"Records the owner's session decision as an argued, not tested, answer.\" 2>&1 | tee $TMPDIR/e2.out\nE=$(grep -o 'E-[A-Z0-9]\\{6\\}' $TMPDIR/e2.out | head -1); echo $E; mkdir -p .tree/evidence/$E && cp .tree/evidence/E-A5JJZQ/scenarios.md .tree/evidence/$E/scenarios.md\nhyp record --exp $E --verdict h-cmd=supports --verdict h-skl=contradicts --observation \"The owner settled q-produced in the session of 2026-11-12: hyp prompt is a command that reads the log. The three scenarios in the evidence file give the command a defined result each (output or exit 1) and the skill none. Argued, not run.\" --evidence .tree/evidence/$E/scenarios.md --why \"Records the session decision under the Experiment's pre-registered rule.\" 2>&1 | tail -3\nhyp status $D 2>&1 | sed -n 1,20p",
 "description": "Build the design-record question, experiment and result"
}
```

Result:

```text
Q-C4SVE2 node.add
H-9XTF12 node.add
H-HYYNSX node.add
E-BYHYTV node.add
E-BYHYTV
R-YHNKGQ result.record
GOAL G-V6KQR7 Design record: hyp prompt is a command that reads the log — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
  Q-C4SVE2 open/supported_unselected Where is the /goal prompt produced?
    candidates (1): H-9XTF12 (h-cmd)
    H-9XTF12 (h-cmd) supported hyp prompt is a command that reads the log
    H-HYYNSX (h-skl) contradicted A skill writes the prompt from hyp show
STALE
OWED
  runner: Q-C4SVE2 select H-9XTF12 (h-cmd), the only candidate
AWAITING OWNER
RECENT
  2026-11-12 14:13 claude-grill-113 result.record R-YHNKGQ — Records the session decision under the Experiment's pre-registered rule.
CHECKPOINT none
```

## Grill

Now the select is owed. I'll take it, add the links, then run the status and check commands.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp select Q-C4SVE2 H-9XTF12 --why \"OWED owes this select: R-YHNKGQ supports h-cmd and contradicts h-skl, on the owner's session decision of 2026-11-12.\" 2>&1 | tail -2\nhyp link G-V6KQR7 external_ref G-VRG6SA --why \"The build Goal that builds this design.\" 2>&1 | tail -2\nhyp link G-VRG6SA external_ref G-V6KQR7 --why \"The design record this Goal builds.\" 2>&1 | tail -2\nhyp link G-VRG6SA external_ref \"dat9uy/hyp#113\" --why \"The issue this Goal serves.\" 2>&1 | tail -2\nhyp link G-V6KQR7 external_ref \"dat9uy/hyp#113\" --why \"The issue this record serves.\" 2>&1 | tail -2",
 "description": "Select the answer and link the Goals and the issue"
}
```

Result:

```text
Q-C4SVE2 dq.select
G-V6KQR7 edge.add
G-VRG6SA edge.add
G-VRG6SA edge.add
G-V6KQR7 edge.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-2/wt && git remote -v | head -2; cd investigations/goal-prompt && export HYP_ACTOR=claude-grill-113\nhyp status 2>&1 | sed -n 1,16p; hyp check | tail -2",
 "description": "Verify issue slug against git remote; store status"
}
```

Result:

```text
GOALS (hyp status G-… prints one Goal)
  G-VRG6SA hyp prints the /goal prompt for any Goal from the log, and … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 1, stale 0); runner owes 1; checkpoint none
  G-V6KQR7 Design record: hyp prompt is a command that reads the log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint none
AWAITING OWNER
check: 0 errors, 0 warnings
```

## Grill

There is no git remote, so I can't confirm that `dat9uy/hyp` is the repository name. I made it up. I'll unlink it and use the bare `#113` you gave me.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp unlink G-VRG6SA external_ref \"dat9uy/hyp#113\" --why \"The repository slug was a guess; the owner names the issue as #113.\" 2>&1 | tail -1\nhyp unlink G-V6KQR7 external_ref \"dat9uy/hyp#113\" --why \"The repository slug was a guess; the owner names the issue as #113.\" 2>&1 | tail -1\nhyp link G-VRG6SA external_ref \"#113\" --why \"The issue this Goal serves.\" 2>&1 | tail -1\nhyp link G-V6KQR7 external_ref \"#113\" --why \"The issue this record serves.\" 2>&1 | tail -1\nhyp checkpoint --goal G-V6KQR7 --summary \"Design record complete: Q-C4SVE2 answered by the owner's session decision (R-YHNKGQ); hyp prompt is a command that reads the log. No lines; its cases are the lines of G-VRG6SA.\" --why \"end of grill\" 2>&1 | tail -1\nhyp checkpoint --goal G-VRG6SA --summary \"End of grill. Runner owes: add rival commit Hypotheses on Q-JSM74G. The design question was moved to G-V6KQR7, selected there. hyp prompt does not exist yet, so the first run's /goal prompt is written by hand from this body.\" --why \"end of grill\" 2>&1 | tail -1\nhyp status G-VRG6SA 2>&1 | sed -n 1,3p; hyp status G-VRG6SA | sed -n '/OWED/,$p'; hyp check | tail -1; cd ../.. && git status --short",
 "description": "Fix issue links, write end-of-grill checkpoints, verify"
}
```

Result:

```text
G-VRG6SA edge.remove
G-V6KQR7 edge.remove
G-VRG6SA edge.add
G-V6KQR7 edge.add
G-V6KQR7 checkpoint
G-VRG6SA checkpoint
GOAL G-VRG6SA hyp prints the /goal prompt for any Goal from the log, and … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 1, stale 0)
LINES
  G-VRG6SA#1 no candidate — [P] For a Goal with a valid constraints block, `hyp prompt G --actor A` exits 0 and prints, each ending in a newline and in this orde…
OWED
  runner: Q-JSM74G add rival Hypotheses
AWAITING OWNER
RECENT
  2026-11-12 14:13 claude-grill-113 edge.add G-VRG6SA — The issue this Goal serves.
  2026-11-12 14:13 claude-grill-113 edge.remove G-VRG6SA — The repository slug was a guess; the owner names the issue as #113.
  2026-11-12 14:13 claude-grill-113 edge.add G-VRG6SA — The issue this Goal serves.
  2026-11-12 14:13 claude-grill-113 edge.add G-VRG6SA — The design record this Goal builds.
  2026-11-12 14:12 claude-grill-113 dq.withdraw Q-H4QXXH — Misplaced: a design settled by argument has no commit for the build Goal's line…
  2026-11-12 14:12 claude-grill-113 result.record R-6HQ2BH — Records the session decision under the Experiment's pre-registered rule.
CHECKPOINT 2026-11-12 14:13 claude-grill-113: End of grill. Runner owes: add rival commit Hypotheses on Q-JSM74G. The design question was moved to G-V6KQR7, selected there. hyp prompt does not exist yet, so the first run's /goal prompt is writte… (full: hyp show G-VRG6SA)
check: 0 errors, 0 warnings
?? investigations/goal-prompt/
```

## Grill

The grill is done. The tree is at `investigations/goal-prompt/`, and `hyp check` reports 0 errors and 0 warnings. Nothing is committed or pushed.

## What is in the tree

| Goal | State |
|---|---|
| **G-VRG6SA**, the build Goal | OWED asks the runner to `add rival Hypotheses` on Q-JSM74G, "Which hyp commit implements `hyp prompt`?". It holds 8 lines (4 `[P]`, 4 `[N]`) and a constraints block. |
| **G-V6KQR7**, the design record | Complete, with no lines. Q-C4SVE2 is answered: `hyp prompt` is a command that reads the log. The select is recorded on R-YHNKGQ, with the session as evidence. |

- **Links:** the two Goals are linked to each other and to `#113` as `external_ref`.
- **Checkpoints:** both Goals have an "end of grill" checkpoint.
- **Withdrawn:** Q-H4QXXH, with its Result R-6HQ2BH and its rivals, stays in history.
- **Issue link:** I first linked `dat9uy/hyp#113`, which was a repository name I guessed. I unlinked it and linked the bare `#113`. Both steps are in the log.

## Before the run starts

`hyp prompt` does not exist, so I could not generate the prompt and wrote it by hand from the Goal body's block and the stop clauses. Commit `investigations/goal-prompt/.tree/` before the run, because the run reads the log it starts from and the Goal's add must be in it. Then use this prompt, with a fresh runner actor:

```text
/goal Run the hyp tree in investigations/goal-prompt of the hyp repository with the hyp-goal-run skill, as HYP_ACTOR=claude-run-113. Its Goal is G-VRG6SA. The store holds two Goals; use hyp status G-VRG6SA wherever the skill says hyp status, and name the Goal with --goal on hyp checkpoint.
Done when the latest printed hyp status G-VRG6SA shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under claude-run-113 in this Goal, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
Work on branch feat/goal-prompt. When the candidate is ready, open a PR and never merge it.
May change: the hyp repository, by commits on the local branch feat/goal-prompt and then a PR that the run opens and never merges
May change: cmd, internal, the docs and a new ADR 0022
May change: skills, for the descriptions of hyp-goal-run and hyp-prototype that name one kind of Goal, the hand-off of hyp-owner-grill, and a build-Goal review round in hyp-goal-run
Read only: the existing ADRs, every other tree under investigations, and recipes
Never: replace the installed hyp binary
Never: edit an existing ADR or a Goal
Never: merge a PR or close #113 (write Part of #113)
```

Once the first run builds `hyp prompt`, later runs can use it.
