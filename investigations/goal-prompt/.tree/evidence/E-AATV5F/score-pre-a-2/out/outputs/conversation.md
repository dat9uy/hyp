# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/issue-113.md"
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
 "command": "hyp status 2>&1 | head -40; ls investigations; ls docs | head -30",
 "description": "Show hyp status and list trees and docs"
}
```

Result:

```text
no .tree store found; run 'hyp init'
checkpoint-store
checkpoint-store-design
id-citation
recipe-base
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
 "command": "git log --oneline | head -15; ls investigations/*; (cd investigations/checkpoint-store && hyp status 2>&1 | head -40); hyp prompt --help 2>&1 | head; ls cmd/hyp internal; grep -rn \"prompt\" cmd/hyp/*.go | head -20; ls /Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp",
 "description": "Survey trees, status, prompt command, code layout"
}
```

Result:

```text
a51f140 Merge pull request #137 from dat9uy/docs/discovery-and-build-goals
6191275 docs: name discovery and build Goals and rewrite the README around them
0ff087a Merge pull request #136 from dat9uy/docs/skill-change-is-a-build-goal
8c80f6c docs: treat a skill change as a build Goal and drop the skill-text gate
f3a4d05 Merge pull request #135 from dat9uy/feat/owner-flow-skills
bf3be84 refactor: name the Goal-writing skill hyp-owner-grill
ee3828f feat: publish the owner-flow skills in skills/
d295ed3 Merge pull request #133 from dat9uy/docs/out-of-goal-finding
d6ac7bc docs: name an out-of-Goal finding in the glossary
3ed9714 Merge pull request #132 from dat9uy/feat/one-subject-per-run
e3521a3 docs: state coverage and a Hypothesis subject as {base, patch, files}
f29b698 docs: record the select of 4047c01 in the checkpoint-store build tree
4047c01 docs: write ADR 0020's run subject as {base, patch, files}
03e9e3d docs: record the owner's answer to the step-5 stop
627c16b docs: record step 5's second review, its fixes and the stop of the checkpoint-store build tree
investigations/checkpoint-store:

investigations/checkpoint-store-design:

investigations/id-citation:
arms

investigations/recipe-base:
arms.txt
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A sto…
  G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a …
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and…
  G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH
[... 2874 more characters]
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-a-2/wt/docs/glossary.md"
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
 "command": "hyp prompt G-X 2>&1 | head -5; echo \"exit $?\"; grep -rln \"prompt\" internal/cli | head; grep -rn \"May change\" --include=*.go --include=*.md . 2>/dev/null | grep -v \"^./investigations\" | head -20; git branch -a; gh --version 2>&1 | head -1; ls /Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/hyp-owner-grill-eval /Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/trial-git | head",
 "description": "Check whether hyp prompt exists and where constraints headings are used"
}
```

Result:

```text
usage: hyp <command> [flags]

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
Statuses are derived from events; no command can set one.

exit 0
(eval):1: no matches found: --include=*.go
* work
gh version 2.103.3 (2026-10-26)
ls: /Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/hyp-owner-grill-eval: Operation not permitted
/Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/hyp-owner-grill-eval:

/Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/trial-git:
gh
gitconfig
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rIl \"May change\" . 2>/dev/null | grep -v \"^./.git/\" | head; echo ---; grep -rIn '\"prompt\"' internal cmd | head; echo ---; cat skills/hyp-goal-run/SKILL.md | head -80; ls skills/*; ls .tree 2>&1 | head -2; git ls-files | grep -c \"^investigations/.*events\"",
 "description": "Search repo for constraints headings and prompt command"
}
```

Result:

```text
./.claude/skills/hyp-owner-grill/SKILL.md
./.claude/skills/hyp-goal-run/SKILL.md
./skills/hyp-owner-grill/SKILL.md
./skills/hyp-goal-run/SKILL.md
---
---
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
- The recipe's checks may exist before
[... 9934 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rIl \"goal-prompt-261111\" . /Users/datguy/.hyp-sandbox/g146-pre-a-2 2>/dev/null | grep -v \"/.git/\" | head; echo ---; grep -rIl \"gh:113\\|gh:114\\|#113\" investigations docs 2>/dev/null | head; echo ---; sed -n 80,140p skills/hyp-goal-run/SKILL.md; echo; ls investigations/checkpoint-store/.tree investigations/checkpoint-store-design/.tree",
 "description": "Locate run-6 prompt and references to the issue"
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-pre-a-2/tmp/issue-113.md
---
---
## Never

- Never change the Goal or its met rule.
- Never record a choice the owner reserved, and never record an owner answer unless the owner gave those words in this session.

investigations/checkpoint-store-design/.tree:
events
evidence

investigations/checkpoint-store/.tree:
events
evidence
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -60; ls .tree/events; ls .tree/evidence | head -3; cd ../checkpoint-store-design && hyp status 2>&1 | head -12; git log --format='%h %ad %s' --date=iso -- ../checkpoint-store/.tree | tail -8",
 "description": "Show build Goal body and event files"
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
[... 9402 more characters]
```

## Grill

Checking one fact before I write the first round: skills already name `hyp prompt`, which the binary lacks.

### Tool Bash (cwd: investigations/checkpoint-store-design)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt && grep -rn \"prompt\" internal/skillcheck/*.go | head; grep -rn \"hyp prompt\" skills docs README.md AGENTS.md | cut -c1-140 | head; (go test ./internal/skillcheck -count=1 2>&1 | tail -5)",
 "description": "Check skillcheck handling of hyp prompt"
}
```

Result:

```text
ok  	github.com/dat9uy/hyp/internal/skillcheck	0.340s
```

### Tool Bash (cwd: .)

```text
{
 "command": "diff <(cat skills/hyp-owner-grill/SKILL.md) <(cat .claude/skills/hyp-owner-grill/SKILL.md) | head -20; diff -q skills/hyp-goal-run/SKILL.md .claude/skills/hyp-goal-run/SKILL.md; git status --short | head",
 "description": "Compare repo skills with installed copies"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

### Tool Bash (cwd: .)

```text
{
 "command": "diff skills/hyp-owner-grill/SKILL.md .claude/skills/hyp-owner-grill/SKILL.md | head -20; echo \"rc=$?\"; ls -la .claude/skills | head -5; grep -c \"hyp prompt\" skills/hyp-owner-grill/SKILL.md",
 "description": "Compare skill copies without process substitution"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

## Grill

I've read the issue and the repo. These facts shape the questions:

- **No store at the repo root.** `AGENTS.md` keeps trees under `investigations/<topic>/`. Issue #113 names no parent issue, so this Goal would go in a new store, `investigations/goal-prompt/`.
- **The binary lacks `hyp prompt`.** `hyp prompt` and the constraints-block warning on `hyp add goal` exist neither in the binary nor in the repo's `skills/` text. Only the installed grill skill, `.claude/skills/hyp-owner-grill`, describes them. That makes #113 a build Goal: the candidate is a hyp commit, and `go test` is the instrument.
- **G-D1CH1R has no constraints block.** Its body is plain prose, and a Goal is never edited. The skill says `hyp prompt` refuses a Goal without a valid block. So "the prompt given G-D1CH1R as it stood" cannot be produced from G-D1CH1R itself.
- **Run 6's prompt isn't on this machine.** It isn't in the repo or in `tmp/`. The issue only quotes run 5's.

## Round 1

**1. Your actor name.** Which actor name should the owner writes use (lowercase letters, digits and hyphens)? The earlier trees used `owner-dat`. *Recommend: `owner-dat`.*

**2. The baseline for the "clause for clause" line.** Today the line has no baseline file.
- **Option A.** You give me run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) as a file. It goes under the new tree's `.tree/evidence/`, and the line names it by sha256. The test builds a synthetic store with G-D1CH1R's lines and body plus a constraints block that reproduces run 6's `Never` lines. It asserts that the generated prompt equals the baseline except for the actor, the owner answer, and the branch and commit.
  - *Example:* the baseline's `Never open or merge a PR.` must come out of the Goal's `Never:` line, not from code.
- **Option B.** The baseline is run 5's text from the issue plus the differences you list. This is weaker, because I would be guessing what run 6 changed.
- **Principle:** a Goal is never edited, so G-D1CH1R can't gain a block. Only a synthetic Goal can test this line.

*Recommend A. Please give me the file, or its text pasted here.*

**3. What the run may do.** For the `May change:`, `Read only:` and `Never:` lines.
- **May change.** My draft: commits on one local branch of this repository, plus the texts the change makes wrong (`skills/`, `docs/`, `AGENTS.md`, `README.md`).
- **Read only.** My draft: `investigations/checkpoint-store` and `investigations/checkpoint-store-design`, which the baseline reads.
- **Never.** My draft:
  - never merge a PR;
  - never replace the installed `~/.local/bin/hyp`, so the run builds to a temp path;
  - never edit a Goal or an old event.
- **Open point.** Run 5 said "Never open or merge a PR". May this run open a PR (unmerged), or never?

*Recommend: never open or merge a PR, with the owner opening it from the branch.*

**4. Scope.** The Goal would hold these four things:
- (a) `hyp prompt`, whose output follows the installed grill skill's description;
- (b) the `hyp add goal` warning and the `hyp prompt` refusal for a body with no valid constraints block;
- (c) the skill-text fold-in from your 2026-11-11 comment: `hyp-goal-run` no longer says "discovery", and `hyp-prototype` no longer says "design";
- (d) syncing the repo's `skills/` copies with the installed grill and goal-run skills, so the shipped text matches the binary.

It would exclude #114, a typed `hyp add goal` field, and a "Never" line that forbids what the Goal needs. *Recommend: all four (a) to (d) in one Goal.* They share one test route, and ADR 0021 makes the skill changes build work anyway.

**5. Should the run test the store-path question?** The prompt prints the store's path.
- **Rival 1:** an absolute path, e.g. `/Users/x/wt/investigations/t`.
- **Rival 2:** a path relative to the repository root, e.g. `investigations/t`.
- **Scenario that separates them:** run `hyp prompt` in a worktree or sandbox clone at a different location, and check that the printed path still resolves there.
- **Principle:** a prompt is used in a clone as often as in the original.

I would record this as a DesignQuestion with both rivals and let the run test it. I would not ask you to pick. *Recommend: yes, test it.*

After your answers I'll run the second-reader review and the preflight scan, then show the routing table. I'll write nothing before you approve that table.

## Owner, message 2

1. Use `owner-standin`.

2. I can't hand you run 6's prompt file from here. Go with your synthetic-store test and widen the exceptions to the actor, the owner's answer, the branch, the PR and commit, and the run's own plan. Branch and PR become flags.

3. Agree with your drafts for May change, Read only and Never, including never opening or merging a PR. Make the constraints strict, one per line. The skill descriptions that name only one kind of Goal go under May change, with no line of their own.

4. Agree: all four, (a) to (d), in one Goal.

5. Agree, test it. I won't pick between the rivals.

On the store: if `AGENTS.md` keeps trees under `investigations/<topic>/`, then `investigations/goal-prompt/` is fine. Please note the single root store as a separate issue.

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader review of prompt design",
 "subagent_type": "general-purpose",
 "run_in_background": true,
 "prompt": "You are an independent reviewer of a design for the Go repo at /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt (the `hyp` CLI). Read-only: change nothing.\n\nThe design to be built (not yet in the code): a new command `hyp prompt <G-id> --actor <runner-actor> [--branch B] [--pr N] [--experiments N]` and a write-time warning in `hyp add goal`. Its rules are written in /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt/.claude/skills/hyp-owner-grill/SKILL.md, sections \"Writes\" (the constraints block paragraph) and \"Hand off\". Also read docs/replay-spec.md (§1 names, §5 refusals, exit codes) and docs/sandbox-spec.md §1.3 for exit codes, docs/glossary.md entries Goal, Owed, Checkpoint, Store, and skills/hyp-goal-run/SKILL.md for what the printed prompt is consumed by. Extra settled facts from the owner: the owner's flags --branch and --pr become flags of the prompt; the prompt's exceptions versus a historical baseline are the actor, the owner's answer, branch, PR, commit and the run's own plan; a Goal is never edited; there is no typed constraints field on `hyp add goal` — constraints live in the Goal body.\n\nTask: walk each rule through scenarios the designers probably did NOT draw, using the real code (internal/cli, internal/replay, internal/store, internal/event) to check facts. Scenarios to cover at least: (1) a store holding two Goals, one with a valid block and one without; `hyp prompt` for each; (2) a merged log from two actors where the Goal's add and a later checkpoint/answer come from different actor files and the \"N Results under the actor\" count; (3) a base/branch that moves mid-run, and a store reached through a git worktree or sandbox clone at another path (what does \"the store's path\" resolve to; what does exit 3 mean there); (4) a Goal body with CRLF, tabs, trailing spaces, a heading in the middle of prose, headings repeated, an empty block, or text after the block; (5) a Goal whose id is unknown or whose node is not a Goal; (6) a Goal already complete or whose OWED is already empty at prompt time; (7) --branch/--pr/--experiments with bad values (space, 0, negative, non-numeric); (8) body length limits in `hyp add goal` and whether the warning versus refusal split is consistent with how other warnings in the code behave.\nReport only concrete findings: for each, the scenario, what the rules give, what is wrong or underspecified, and the file:line you checked. Flag anything you could not verify. Under 600 words."
}
```

Result:

```text
[redacted: the loaded grill skill]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"body\\|Body\" internal/cli/*.go | grep -i \"goal\\|max\\|len\" | head -15; grep -rn \"rune\" internal/cli/add*.go internal/replay/validate*.go 2>/dev/null | head -8; ls internal/cli | head -40; grep -n \"exit\\b\\|Exit\" docs/sandbox-spec.md | sed -n 1,8p",
 "description": "Preflight scan of goal body handling and cli layout"
}
```

Result:

```text
internal/cli/cli.go:63:  add goal --title T [--body B] [--alias A] [--line L]... [--cite N=KEY]...
internal/cli/goalline_test.go:86:	for name, body := range files {
internal/cli/goalline_test.go:87:		sum := sha256.Sum256([]byte(body))
internal/cli/goalline_test.go:102:// A Hypothesis may name its subject, the files that embody its claim. A
internal/cli/write_test.go:651:func TestGoalBodyHoldsLongerTextThanOtherNodes(t *testing.T) {
internal/cli/write_test.go:656:	want(t, hyp(t, "add", "goal", "--title", "G", "--body", strings.Repeat("x", 2001), "--why", "w"), 1, "body")
internal/cli/write_test.go:657:	g := must(t, "add", "goal", "--title", "G", "--body", strings.Repeat("x", 2000), "--why", "w")
internal/cli/write.go:253:	maxBody := event.MaxText
internal/cli/write.go:255:		maxBody = event.MaxGoalBody
internal/cli/write.go:257:	body := w.fs.String("body", "", fmt.Sprintf("body (at most %d runes)", maxBody))
internal/cli/write.go:725:	body := w.fs.String("body", "", fmt.Sprintf("the Hypothesis's revised body (required, at most %d runes)", event.MaxText))
(eval):1: no matches found: internal/cli/add*.go
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
28:`hyp bundle run` and `hyp bundle known-answers` find REF in `recipes/` or `rubrics/`; a rubric is a bundle with one Standalone command and no step (replay spec §7.6), so `hyp bundle evidence` refuses it with exit 2, and a ref found in both libraries is refused with exit 1
[... 5623 more characters]
```

## Grill

The second-reader review is running. The preflight scan found one fact that matters: a Goal body holds at most 2000 runes (`event.MaxGoalBody`), so the constraints block has to fit inside that limit. I'll show the routing table once the review comes back.

## Owner, message 3

<task-notification>
<task-id>aa2622630d1718e10</task-id>
<tool-use-id>toolu_012qAS7Ee344kTA1GyL5yX55</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-a-2-wt/766dc026-fbda-445f-8ddd-3d6745ec2e1b/tasks/aa2622630d1718e10.output</output-file>
<status>completed</status>
<summary>Agent "Second-reader review of prompt design" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Findings from reading the rules against the code. I ran nothing, and I could not check every code path I cite.

1. **Exit 3 and the "store's path"**
   - `store.FindRoot` (`internal/store/store.go:47-62`) walks up from the cwd to the first directory holding `.tree/events`. It never calls git, so there is no "repository" test. The only failure is `ErrNoStore` ("no .tree store found; run 'hyp init'").
   - The Hand off text says exit 3 means "no repository holds the store". That is wrong. The real condition is "no `.tree/events` at or above the cwd".
   - `cli.go:186-195` maps any error not marked refusal or bad-arguments to exit 3. So `prompt` must wrap `ErrNoStore` with `herr.Env`. Any other unclassified error, such as an unreadable events directory or a corrupt line, also gives exit 3. The text does not say so.
   - The path is `filepath.Abs` of the cwd. It is not symlink-resolved (`EvalSymlinks` is not called). The same store can print as two different paths. Run from a worktree, a sandbox clone or a `.hyp-sandbox/&lt;x&gt;/wt` checkout, it prints that checkout's path.
   - The spec should say whether `prompt` prints the cwd-found path as is and whether the runner must reach the same path. I could not check whether `hyp-goal-run` assumes a fixed path.

2. **Merged log, actor files, and the "N Results under the actor" count**
   - Events merge by smallest ULID, tie broken by actor name (`replay-spec.md:155-156`). The Goal's add can sit in `owner.jsonl` while the checkpoint or answer sits in another actor's file. `prompt` must read the merged projection, not one file.
   - A new runner actor has zero Results, so the stop clause "N Results under the actor" starts at 0. That holds only if the actor is really new. If the actor already has Results, the count is already partly spent.
   - `prompt` does not check this. The actor regex `^[a-z0-9][a-z0-9-]{0,39}$` (`replay-spec.md:23`) is checked at write time, and it is unclear whether `prompt` applies it. A bad `--actor` should be exit 2.
   - Not stated: whether retracted Results count, and whether Results on other Goals in a multi-Goal store count. Both are open.
   - Also open: is the count scoped to this Goal's Results, or to all of the actor's Results?

3. **Two Goals in one store**
   - Rule 1 is clear: `prompt G-a` works and `prompt G-b` gives exit 1. The prompt must scope the stop clauses and the OWED reference to one Goal (`hyp status &lt;G-id&gt;`). `hyp-goal-run/SKILL.md:13` already requires this.
   - Not stated: whether the prompt reminds the runner to pass `--goal` on its checkpoint.
   - Not stated: whether the prompt shows the store path or the Goal id first when the store holds many Goals.

4. **Moving base or branch**
   - `--branch` and `--pr` are stored only as text in the prompt. The rules do not say whether `prompt` checks that the branch exists or that the PR number is real.
   - The Hand off text says to run `hyp prompt` on "the log the run starts from" and that the Goal's add must be "merged or on the branch the prompt names". So `prompt` run on a different branch would give exit 1 (unknown Goal) with no hint that the branch is the problem.
   - Not specified: whether `prompt` can detect that the Goal is absent from the current branch. If the answer is no, the message should name the possibility.

5. **Block parsing edge cases**
   - The rule says the block runs "from the first line that starts with a heading to the end". A heading in the middle of prose, such as "Never: do X" inside a paragraph, starts the block early. Everything after it must be valid.
   - So prose text after any heading-led line makes the Goal invalid. Because a Goal is never edited, that makes it permanently unusable. The warning at `add goal` time is the only chance to fix it.
   - CRLF: the skill says to write the body with `\n` line ends, and the parser must treat `\r` as non-empty. I found no normalisation of the body in `write.go`, but did not check every path, so I could not verify it.
   - The warning therefore must use the same parser as `prompt`. The code must not duplicate the logic in two places.
   - "Starts with" is ambiguous on leading whitespace. A line with an indent before `Never:` would be a non-heading, non-empty line and invalid. Say so explicitly.
   - An empty block, a missing heading (the block needs all three), and a trailing empty line need fixed rules. The rules say "at least once" and "non-empty line starts with one of them", but not whether the final newline counts.
   - Repeated headings are allowed. The rule does not say whether the `prompt` output must preserve their order. "Line for line" implies yes.
   - A body with no heading at all: the warning stays silent ("a heading but no valid block"), but `prompt` refuses it. That is consistent, but a Goal with no constraints is silently unusable. The grill skill's table says the heading is required, so `add goal` could warn on any body that lacks a valid block.

6. **Unknown id, or a node that is not a Goal**
   - Both should be exit 1. I found no existing helper to reuse that I verified. `resolve.go` likely handles id lookup, but I did not read it.
   - The rules are silent here. They should name the messages and the exit code.
   - An alias or short id for the Goal should resolve as other commands do.

7. **Goal already complete, or OWED empty**
   - The glossary says a Goal is complete when every live DesignQuestion under it is answered, and "met" is complete with no line under OWED.
   - The rules do not say whether `prompt` refuses a complete Goal, warns, or prints anyway. The stop clause "the runner owes nothing" would fire at once.
   - I recommend printing with no refusal, because a Goal can reopen after an owner return. The spec should say so.
   - Also note that a Goal with no DesignQuestion yet is not complete and has empty OWED. The prompt would stop at once there too, so a first run is not stopped by that clause. Not verified in `replay.go:530`.

8. **Bad flag values**
   - The rules name only "a `--branch` with a space" as exit 2. `--pr 0`, `--pr -1`, `--pr abc`, `--experiments 0`, `--experiments -3`, and an empty `--branch ""` are undefined.
   - `--experiments 0` would give a stop clause that fires immediately. Treat it as exit 2.
   - Repeated flags: `fs.String` takes the last value silently.
   - The branch charset is not fixed. Git allows many characters. Pick a conservative set, otherwise the prompt text could carry shell-special characters.
   - A `--pr` clause with no `--branch` is allowed, but the rules do not say what the runner is told then.

9. **Body length and warning versus refusal**
   - The Goal body limit is `MaxGoalBody = 2000` runes (`event.go:117`), counted as runes, and over-limit input is refused. The constraints block is part of those 2000 runes. A long body leaves little room, and the warning can fire only after the body was accepted.
   - Warnings elsewhere (`write.go:406-410`, `800-805`, `958-962`) are collected as `warn: …` lines and printed to stderr after a successful write, exit 0. The new warning fits that pattern.
   - Note the Goal-add path prints warnings only if `err == nil`, and the Goal branch builds `warns` only for Experiment kinds at `write.go:385-392`. The new warning needs a new place for Goal kind.
   - Because a Goal is never edited, a warning is a weak guard for something that later makes the Goal unusable with `prompt`. A refusal at add time would be stricter. The designers chose a warning, so a body that only mentions a heading in prose is not blocked. That is a reasonable split, but state the reason in the spec.

10. **Docs the design must update**
    - `replay-spec.md` has no `prompt` command. The AGENTS.md rule that a state machine has one spec means the spec needs the new refusals and exit codes.
    - `skillcheck` fails on any `hyp` command named in `skills/` that does not exist, so the Hand off text cannot land before the command does.</result>
<usage><subagent_tokens>44014</subagent_tokens><tool_uses>6</tool_uses><duration_ms>40806</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 44,64p internal/store/store.go; sed -n 383,394p internal/cli/write.go; sed -n 183,196p internal/cli/cli.go",
 "description": "Verify reviewer's claims against source"
}
```

Result:

```text
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
		}
		if p.Aims, err = w.keys("--aims", aims, event.KeyLimit, event.KeyReading); err != nil {
			return draft{}, err
		}
		if kind == event.KindExperiment {
			if warns, err = admitted(w.st.AdmitExp(p, facts, rubricFacts, noTarget)); err != nil {
				return draft{}, err
			}
		}
		return draft{event.TypeNodeAdd, "", p}, nil
	})
	if err == nil && kind == event.KindGoal {
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
```

## Grill

Your answers are applied: actor `owner-standin`, all four scope items (a) to (d), the store at `investigations/goal-prompt/`, and the branch and PR as flags. The second reader found real gaps, and I checked them against the source. Two of them change what I told you, so I'm flagging them before the routing table.

**What I verified**
- **Exit 3.** `store.FindRoot` (`internal/store/store.go:47`) only walks up to a `.tree/events` directory and never calls git. So "exit 3 means no repository holds the store" in the installed grill skill is wrong. The real condition is that no store sits at or above the working directory. Item (d) will correct that text.
- **Warnings.** `hyp add goal` builds warnings only for Experiments (`internal/cli/write.go:385`). The constraints-block warning needs a new place for Goals, and it must use the same parser as `hyp prompt` so the two cannot disagree.
- **Body limit.** The Goal body limit is 2000 runes, and the block counts inside it.

**Two corrections to my round 1**
- **The baseline is run 5's text, not run 6's.** Since run 6's file isn't available, the synthetic-store test compares with run 5's prompt as the issue quotes it. Its paths become placeholders and it is pinned by sha256. #113's "Done means" names run 6, so it would not be literally met. AGENTS.md then wants a new placeholder issue stating the unmet part. See A below.
- **The skill's stop clauses differ from the prompts that actually ran.** Run 5's prompt had four clauses: owes nothing, the "add rival Hypotheses" clause, 5 Experiments recorded, and "cannot take the next step honestly". The installed skill describes only "owes nothing, or N Results under the actor". Run 6's checkpoint also cites a clause "selected the candidate on Q-PC04ED". Someone has to pick the clause set, and it is you. See B.

## Round 2

**A. The baseline.** Confirm that the baseline is run 5's quoted text, with placeholders and a sha256. The exceptions are: actor, owner answer, branch, PR, commit, and the run's own plan. #113 then needs a placeholder issue stating "match run 6's prompt", because that part is unmet. *Recommend: yes. The issue text is below, since I can't file it from here.*

**B. Which stop clauses the generated prompt always carries.**
- (i) The runner owes nothing: stop and report.
- (ii) N Results recorded.
- (iii) "you cannot take the next step honestly and say why".
- (iv) The run-5 "every runner item is `add rival Hypotheses`…" clause. It only fits discovery runs.
- (v) "selected the candidate on Q-…". It names a particular question, so it would be the run's own plan, not generated.

*Recommend: (i), (ii) and (iii) always; drop (iv) and (v). `--experiments N` sets the number in (ii).*

**C. What (ii) counts.** The flag is `--experiments`, but the skill says "Results".
- **Option 1:** non-retracted Results recorded under this actor on this Goal since the prompt was printed. Example: a new actor starts at 0 and stops at 5.
- **Option 2:** Experiments the actor added, as run 5's text said.
- **Principle:** a Result is what a run produces, and an Experiment can sit planned. *Recommend: Option 1, with the prompt saying "Results".*

**D. A complete Goal.** Should `hyp prompt` print for a Goal that is already complete or already owes the runner nothing, or refuse?
- **Print:** the run stops at once and reports. An owner return can reopen a Goal.
- **Refuse:** exit 1 with a message.

*Recommend: print.*

**E. A malformed constraints block.** A Goal is never edited, so a body that is accepted with a bad block becomes permanently unusable for `hyp prompt`.
- **Today's skill:** `hyp add goal` only warns, and only when the body has a heading but no valid block.
- **Option 1:** keep the warning, and also warn when the body has no heading at all.
- **Option 2:** refuse (exit 1) at add time for any body without a valid block. This would break `hyp add goal` for Goals that need no run, such as a design record.

*Recommend: Option 1.*

**F. Flag bounds.** The `[N]` line would refuse these with exit 2:
- `--actor` outside the replay-spec name pattern.
- `--branch` empty, with a leading `-`, or outside `[A-Za-z0-9._/-]`.
- `--pr` or `--experiments` that is not a positive integer.

*Recommend: yes, as written.*

**G. The single-root-store issue.** This is the draft for you to file:
> **Title:** Where do Goals live: one root store, or one store per topic?
> **Event:** #113's grill put its Goal in `investigations/goal-prompt/` because AGENTS.md keeps trees under `investigations/<topic>/`. ADR 0020 says one store holds many Goals, and the grill skill says a new purpose goes in the same store.
> **Scope:** core. **Question:** should a repository keep one root store, or one store per topic, and which text says so?

Also: I'll keep **G** out of this Goal, since you asked for it as a separate issue.

When you've answered A to F, I'll show the routing table.

## Owner, message 4

A. Yes, with run 5's quoted text as the baseline, placeholders and a sha256. File the placeholder issue for the unmet "match run 6's prompt" part.

B. Agree: (i), (ii) and (iii) always, with (iv) and (v) dropped. `--experiments N` sets N in (ii). Earlier I wanted a cap on the run's Experiments as a flag defaulting to 5, so that default stands.

C. Agree: Option 1, and the prompt says "Results".

D. Agree: print.

E. Agree: Option 1.

F. Agree, as written.

G. The draft looks fine as a separate issue, and I'll file it myself. It stays out of this Goal.

### Tool Bash (cwd: .)

```text
{
 "command": "timeout 30 gh issue create --title \"Owner flow: a generated /goal prompt matches run 6's prompt clause for clause\" --body \"Part of #113.\n\n## Unmet goal\n#113 Done means: the prompt generated for G-D1CH1R as it stood before run 6 matches run 6's prompt (goal-prompt-261111-2129-g1b-run-6) clause for clause, except the actor, the owner's answer, the branch, the PR, the commit and the run's own plan.\n\n## Why it is unmet\nRun 6's prompt file is not available. The build Goal for #113 tests against run 5's prompt as quoted in #113 (placeholders, pinned by sha256) on a synthetic store, because G-D1CH1R has no constraints block and a Goal is never edited.\n\n## Scope\nCore.\n\n## Done means\nThe generated prompt is compared with run 6's prompt file, pinned by sha256, on a store cut of G-D1CH1R before run 6 plus a constraints block.\" 2>&1 | tail -3",
 "description": "Try to file the placeholder issue"
}
```

Result:

```text
To get started with GitHub CLI, please run:  gh auth login
Alternatively, populate the GH_TOKEN environment variable with a GitHub API authentication token.
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd \"$TMPDIR\" && cat > run5-prompt.txt <<'EOF'\n/goal Run the hyp tree in <tree> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G-id>.\nDone when the latest printed `hyp status` shows `runner owes nothing: stop and report to the owner` under OWED, or every runner item under OWED is `add rival Hypotheses` on a Question for which no owner checkpoint since the latest stop checkpoint asks for a new Hypothesis, or 5 Experiments have been recorded since this goal began, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.\nNever change <the investigated repository> or its issues; read it only, and work in throwaway copies.\nNever put application code or real source data in the public hyp repository.\nTreat <legacy code> as legacy. <...> may be test cases.\nNever open or merge a PR.\nEOF\nshasum -a 256 run5-prompt.txt; wc -c run5-prompt.txt; echo $TMPDIR",
 "description": "Write run-5 baseline text and hash it"
}
```

Result:

```text
840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3  run5-prompt.txt
     874 run5-prompt.txt
/tmp/claude-501
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt
```

## Grill

I couldn't file the placeholder issue. `gh` isn't authenticated and the machine has no network. Please file it yourself. This is the text:

> **Title:** Owner flow: a generated /goal prompt matches run 6's prompt clause for clause
> Part of #113.
> **Unmet goal:** #113's Done means: the prompt generated for G-D1CH1R before run 6 matches run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) clause for clause, except the actor, the owner's answer, the branch, the PR, the commit and the run's own plan.
> **Why it is unmet:** run 6's file is not available. The build Goal tests against run 5's quoted prompt on a synthetic store instead.
> **Scope:** core. **Done means:** the same comparison, against run 6's file pinned by sha256.

## Routing table

I ran the mistakes checklist on it:
- **Store:** the issue has no parent, so this is a new store at `investigations/goal-prompt/`, which is where `AGENTS.md` keeps trees.
- **Test method:** no installed skill owns it.
- **Constraints:** the block uses directories, and no `Never:` line forbids anything the run needs.

**Baseline:** `run-5-prompt.txt` is run 5's quoted text with its placeholders kept. Its sha256 is `840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3` (874 bytes). I'll copy that file into the new tree's evidence before the Goal add.

| # | Question | Goes to |
|---|---|---|
| 1 | Purpose, scope, constraints | Goal title and body (below) |
| 2 | The shape of the prompt, its stop clauses, and what N counts | Lines 1, 2, 6, 7 |
| 3 | Refusals, and warn versus refuse | Lines 3, 4 |
| 4 | Flag bounds | Line 5 |
| 5 | No store, exit 3 | Line 8 |
| 6 | Store path, absolute or repo-relative | DesignQuestion, rivals R1 and R2 below |
| 7 | Skill descriptions, the skill sync and the docs | `May change:`, no line |
| 8 | Run 6's prompt and the single root store | Two issues, filed by you |
| 9 | Owner's actor, the exit-3 wording, the `FindRoot` facts, the Goal body limit | Looked up |

**DesignQuestion Q: "What form of store path does `hyp prompt` print?"**
- **R1:** the absolute path of the store root.
- **R2:** the path relative to the repository root. R2 needs a git toplevel, and with none it exits 3.
- **Separating scenario:** create a worktree and a clone at other paths, run `hyp prompt` in each, and check that each printed path resolves to the store holding the Goal.
- **Candidates:** each rival is a commit meeting all eight lines. The two differ only in the path form.

**Goal title:** "hyp prints the /goal prompt for any Goal from its log, and warns when a Goal's constraints block is unusable"

**Goal body:**
```text
Builds #113: hyp prompt prints the /goal prompt for any Goal from the log, and hyp add goal warns about a constraints block no prompt could use. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Constraints live in the Goal body under three headings and no typed field is added (owner, 2026-11-11). A prompt may narrow them, never widen them. Out of scope: what a commit-candidate Goal needs from hyp (#114), one root store versus one per topic, and matching run 6's prompt (a placeholder issue).

May change: this repository, as commits on one local branch
May change: the texts the change makes wrong, by directory: skills/, docs/, AGENTS.md, README.md, including a skill description that names only one kind of Goal
Read only: investigations/checkpoint-store and investigations/checkpoint-store-design
Never: open or merge a PR
Never: replace the installed hyp binary; build to a temporary path
Never: edit a Goal or an event already in a log
```
This body is well under the 2000-rune limit and passes its own block rule.

**Lines.** Each is a build-Goal line, so a test on the candidate decides it.

1. **[P]** For a Goal whose body ends in a valid block, `hyp prompt G --actor A` exits 0 and prints, in order:
   - the store path;
   - the actor;
   - the Goal id;
   - the skill name `hyp-goal-run`;
   - the stop clauses: the runner owes nothing; N Results recorded under the actor, non-retracted and on this Goal, with N = `--experiments` and default 5; you cannot take the next step honestly and say why;
   - the block, line for line in body order.

   `--branch` and `--pr` each add one clause after the block. Two runs over one log print identical bytes.
2. **[P]** Baseline `run-5-prompt.txt`, sha256 `840d813d…e3a3`. On a synthetic store whose Goal has G-D1CH1R's title and lines, with a body ending in a block of run 5's constraint lines, the output matches the baseline clause for clause. It differs only in:
   - the actor, the owner answer, the branch, the PR, the commit and the run's own plan;
   - the dropped "add rival Hypotheses" clause;
   - "Results" for "Experiments";
   - run 5's "Treat … as legacy" sentence, which prints as its `Read only:` line.
3. **[N]** `hyp prompt` exits 1, with a message naming the rule and nothing on stdout, for:
   - an unknown id;
   - a node that is not a Goal;
   - a Goal body without a valid block.

   A valid block starts at the first line that begins at column 0 with a heading. It runs to the end of the body, holds all three headings, and every non-empty line starts at column 0 with a heading. A line of spaces, a tab or a lone CR is not empty.
4. **[N]** `hyp add goal` warns (a `warn:` line on stderr, exit 0, naming the rule) for exactly the bodies line 3 refuses, including a body with no heading. It never refuses the add, and a log holding such a body replays, with `hyp check` reporting 0 errors.
5. **[N]** `hyp prompt` exits 2 for:
   - an `--actor` outside the replay-spec name pattern, or none;
   - a `--branch` that is empty, starts with `-`, or has a character outside `[A-Za-z0-9._/-]`;
   - a `--pr` or `--experiments` that is not a positive integer.
6. **[P]** The output depends only on the merged projection. In a two-Goal store, `hyp prompt G-a` names only G-a's id and block, and points to `hyp status G-a`. With the Goal's add and its checkpoints in two actors' files, it prints the same bytes as when the events share one file.
7. **[P]** For a Goal that is complete or whose OWED already says the runner owes nothing, `hyp prompt` still prints and exits 0.
8. **[N]** With no store at or above the working directory, `hyp prompt` exits 3 and prints nothing on stdout.

The lines name no process, history or owner act. Red on base, the removal check and a test pair per rule go in each Experiment's verdict rule.

**Two things for you to confirm**
- **Line 2's extra exceptions.** The two I added beyond your list are the dropped (iv) clause plus "Results" for "Experiments", and the "Treat … as legacy" sentence printing as a `Read only:` line.
- **Line 8's wording.** It states the real exit-3 condition from the source, not the grill skill's "no repository holds the store".

Please confirm each row, or move it. After that I'll show the exact write commands, with `owner-standin` on the owner writes, one by one for your approval. I'll write nothing before then.

## Owner, message 5

The table matches my positions, so I confirm every row and both extra points, with one change to row 8.

- **Row 8:** I won't file anything on GitHub in this session. Keep the placeholder issue text and the root-store issue text as notes in the hand-off, and I'll file both later.
- **Line 2's extra exceptions:** confirmed, both of them.
- **Line 8's wording:** confirmed. State the real exit-3 condition from the source.

Show me the write commands.

### Tool Bash (cwd: .)

```text
{
 "command": "cd \"$TMPDIR\" && cat > lines.txt <<'EOF'\n[P] For a Goal whose body ends in a valid constraints block, `hyp prompt G --actor A` exits 0 and prints in order: the store path; the actor; the Goal id; the skill name hyp-goal-run; the stop clauses (the runner owes nothing; N Results recorded under the actor, non-retracted and on this Goal, N from --experiments, default 5; you cannot take the next step honestly and say why); the block, line for line in body order. --branch and --pr each add one clause after the block. Two runs over one log print identical bytes.\n[P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Goal has G-D1CH1R's title and lines and a body ending in a block of run 5's constraint lines, the output matches the baseline clause for clause and differs from it only in: actor, owner answer, branch, PR, commit, the run's own plan; the dropped add-rival-Hypotheses clause; Results for Experiments in the count clause; and run 5's Treat-as-legacy sentence, which prints as its Read only: line.\n[N] `hyp prompt` exits 1, with a message naming the rule and nothing on stdout, for an unknown id, a node that is not a Goal, and a Goal body without a valid block. A valid block starts at the first line that begins at column 0 with May change:, Read only: or Never:, runs to the end of the body, holds all three headings, and every non-empty line in it starts at column 0 with one. A line of spaces, a tab or a lone carriage return is not empty.\n[N] `hyp add goal` warns (a warn: line on stderr, exit 0, naming the rule) for exactly the bodies the previous line makes `hyp prompt` refuse, a body with no heading included. It never refuses the add, and a log holding such a body replays with `hyp check` reporting 0 errors.\n[N] `hyp prompt` exits 2 for an --actor outside the replay-spec name pattern or missing; a --branch that is empty, starts with -, or has a character outside A-Za-z0-9._/-; and a --pr or --experiments that is not a positive integer.\n[P] The output depends only on the merged projection. In a two-Goal store, `hyp prompt G-a` names only G-a's id and block and points to `hyp status G-a`; with the Goal's add and its checkpoints in two actors' files it prints the same bytes as when the events share one file.\n[P] For a Goal that is complete, or whose OWED already says the runner owes nothing, `hyp prompt` still prints and exits 0.\n[N] With no store (.tree/events) at or above the working directory, `hyp prompt` exits 3 and prints nothing on stdout.\nEOF\nwhile IFS= read -r l; do printf '%s' \"$l\" | wc -m; done < lines.txt\ncat > body.txt <<'EOF'\nBuilds #113: hyp prompt prints the /goal prompt for any Goal from the log, and hyp add goal warns about a constraints block no prompt could use. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Constraints live in the Goal body under three headings and no typed field is 
[... 762 more characters]
```

Result:

```text
     520
     525
     446
     276
     231
     274
     123
     118
     987 body.txt
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp 2>&1 | sed -n 1,60p",
 "description": "Read hyp command usage for exact write flags"
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
[... 1669 more characters]
```

## Grill

Nothing has been written yet. Here are the commands in order. Only step 2 is an owner write, so I need your approval of that exact command before I run it.

The runner actor for steps 1, 3, 4 and 5 is `claude-grill-113`, unique to this session.

**1. Runner: create the store and the baseline file**
```bash
mkdir -p investigations/goal-prompt && cd investigations/goal-prompt
HYP_ACTOR=claude-grill-113 hyp init
mkdir -p .tree/evidence/baseline
cp "$TMPDIR/run5-prompt.txt" .tree/evidence/baseline/run-5-prompt.txt
shasum -a 256 .tree/evidence/baseline/run-5-prompt.txt   # must print 840d813d…e3a3
```

**2. Owner write: needs your approval**

I'll write it with your actor, `owner-standin`:
```bash
HYP_ACTOR=owner-standin hyp add goal \
  --title "hyp prints the /goal prompt for any Goal from its log, and warns when a Goal's constraints block is unusable" \
  --body "$(cat "$TMPDIR/body.txt")" \
  --line "$(sed -n 1p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 2p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 3p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 4p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 5p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 6p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 7p "$TMPDIR/lines.txt")" \
  --line "$(sed -n 8p "$TMPDIR/lines.txt")" \
  --why "Owner flow #113: a generated /goal prompt and the run's constraints, confirmed in the grill session"
```
- **Files.** `$TMPDIR/body.txt` and `$TMPDIR/lines.txt` hold exactly the Goal body and the eight lines from the table you confirmed.
- **Line changes.** Line 3 now spells out the three headings. Line 4 says "the previous line" instead of "line 3". Every line is 118 to 525 runes, and the body is 987 runes, under the 2000 limit.
- **Cites.** No line cites a key.

**3. Runner: the contested question and its rivals**

`G-…` is the id that step 2 prints.
```bash
export HYP_ACTOR=claude-grill-113
hyp add dq --parent G-… --title "What form of store path does hyp prompt print?" \
  --body "Scenario that separates the rivals: create a git worktree and a clone at other paths, run hyp prompt in each, and check that each printed path resolves to the store holding the Goal. Each rival is a commit that meets all eight lines and differs from the other only in this form. The owner does not pick between them." \
  --why "Testable choice routed to a DesignQuestion; owner declined to choose"
hyp add hyp --parent Q-… --title "Print the absolute path of the store root" --body "..." --why "Rival 1"
hyp add hyp --parent Q-… --title "Print the store path relative to the repository root; with no git toplevel, exit 3" --body "..." --why "Rival 2"
```

**4. Runner: link and checkpoint**
```bash
hyp link G-… external_ref gh:113 --why "The issue this Goal builds"
hyp checkpoint --goal G-… --summary "Grill ended: one build Goal, eight lines, one DesignQuestion with two rivals; placeholder issue for run 6's prompt and the root-store issue are filed by the owner later" --why "end of grill"
```

**5. Runner: check and commit**
```bash
hyp status G-…
hyp check        # must print: check: 0 errors
git add investigations/goal-prompt/.tree/events investigations/goal-prompt/.tree/evidence
git commit -m "docs: record the grill for #113 in investigations/goal-prompt"
```
This runs after you approve step 2. The commit message uses the conventional format with no AI trailer, as `AGENTS.md` requires.

**Hand-off notes for you to carry**
- **Prompt generator.** `hyp prompt` doesn't exist in the binary yet, because this Goal builds it. So I can't print the run's `/goal` prompt with it. I'll write it by hand in the skill's shape, flag it as hand-written, and mark where `hyp prompt` should replace it later.
- **Placeholder issue.** The text is above. It says "Part of #113", so #113 stays open until it is filed.
- **Root-store issue.** The draft is in my round 2 message. You file both later; I file nothing.

Do you approve step 2 as written?

## Owner, message 6

Yes, I approve step 2 as written, run with `HYP_ACTOR=owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p investigations/goal-prompt && cd investigations/goal-prompt && HYP_ACTOR=claude-grill-113 hyp init && mkdir -p .tree/evidence/baseline && cp \"$TMPDIR/run5-prompt.txt\" .tree/evidence/baseline/run-5-prompt.txt && shasum -a 256 .tree/evidence/baseline/run-5-prompt.txt && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"hyp prints the /goal prompt for any Goal from its log, and warns when a Goal's constraints block is unusable\" \\\n  --body \"$(cat \"$TMPDIR/body.txt\")\" \\\n  --line \"$(sed -n 1p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 2p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 3p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 4p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 5p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 6p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 7p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 8p \"$TMPDIR/lines.txt\")\" \\\n  --why \"Owner flow #113: a generated /goal prompt and the run's constraints, confirmed in the grill session\" 2>&1",
 "description": "Init store, copy baseline, add the Goal as owner"
}
```

Result:

```text
Exit code 2
initialised .tree in /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt/investigations/goal-prompt
840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3  .tree/evidence/baseline/run-5-prompt.txt
--line names G-D1CH1R, which is no node here; cite an id this store holds, or write <name>:<id> for another store's
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 1,60p internal/cli/idcite.go | grep -n \"regexp\\|MustCompile\\|[A-Z]\\{4,\\}\" | head; ls investigations/goal-prompt/.tree/events",
 "description": "Check which tokens the id-cite rule matches"
}
```

Result:

```text
26:		if !strings.ContainsRune("GQHER", rune(text[i])) || text[i+1] != '-' {
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 10,70p internal/cli/idcite.go",
 "description": "Read id-cite token rule"
}
```

Result:

```text
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
		if len(rest) < 6 || !isCrockford(rest[:6]) || len(rest) > 6 && isWordByte(rest[6]) {
			continue
		}
		token := text[i : i+8]
		if _, err := resolveRef(p, token); err != nil {
			out = append(out, token)
		}
	}
	return out
}

func hasPlaceholder(s string) bool {
	for _, m := range placeholders {
		if strings.HasPrefix(s, m) {
			return true
		}
	}
	return false
}

// isCrockford reports whether s is upper-case Crockford base32, the
// alphabet of a ULID tail and so of a display id.
func isCrockford(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'A' && c <= 'Z') || strings.IndexByte("ILOU", c) >= 0 {
			return false
		}
	}
	return true
```

## Grill

The Goal add was refused (exit 2) and nothing was appended. Line 2 contains the bare id `G-D1CH1R`, and this store has no such node. The store's cite rule refuses an id-shaped token unless it is a node here or is written `<tree>:<id>`.

**What did happen:** step 1 completed. The store `investigations/goal-prompt` exists, and the baseline file is in place with the expected sha256 `840d813d…e3a3`.

**The one change.** I checked every line, the body and the title against the cite rule (`internal/cli/idcite.go`). Only line 2 trips it. I'd write `G-D1CH1R` as `checkpoint-store:G-D1CH1R` in line 2, which is the form the rule asks for and names the tree it comes from. Nothing else changes, and the line stays at about 540 runes.

Because this alters the text you approved, I won't run it on the old approval. This is line 2 as it would be written:

> [P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Goal has checkpoint-store:G-D1CH1R's title and lines and a body ending in a block of run 5's constraint lines, the output matches the baseline clause for clause and differs from it only in: actor, owner answer, branch, PR, commit, the run's own plan; the dropped add-rival-Hypotheses clause; Results for Experiments in the count clause; and run 5's Treat-as-legacy sentence, which prints as its Read only: line.

The command is the same as before, with `HYP_ACTOR=owner-standin`, the same title, body and `--why`, and lines 1 and 3 to 8 unchanged.

Do you approve this command with the changed line 2?

## Owner, message 7

Yes, I approve the command with the changed line 2, run with `HYP_ACTOR=owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt/investigations/goal-prompt && sed -i.bak \"2s/G-D1CH1R's title/checkpoint-store:G-D1CH1R's title/\" \"$TMPDIR/lines.txt\" && rm \"$TMPDIR/lines.txt.bak\" && sed -n 2p \"$TMPDIR/lines.txt\" | cut -c1-200 && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"hyp prints the /goal prompt for any Goal from its log, and warns when a Goal's constraints block is unusable\" \\\n  --body \"$(cat \"$TMPDIR/body.txt\")\" \\\n  --line \"$(sed -n 1p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 2p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 3p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 4p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 5p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 6p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 7p \"$TMPDIR/lines.txt\")\" \\\n  --line \"$(sed -n 8p \"$TMPDIR/lines.txt\")\" \\\n  --why \"Owner flow #113: a generated /goal prompt and the run's constraints, confirmed in the grill session\" 2>&1",
 "description": "Fix line 2 and add the Goal as owner"
}
```

Result:

```text
[P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Goal has checkpoint-store:G-D1CH1R's title and lines and a body endin
G-5WZZH1 node.add
  line G-5WZZH1#1: [P] For a Goal whose body ends in a valid constraints block, `hyp prompt G --actor A` exits 0 and prints in order: the store path; the actor; the Goal id; the skill name hyp-goal-run; the stop clauses (the runner owes nothing; N Results recorded under the actor, non-retracted and on this Goal, N from --experiments, default 5; you cannot take the next step honestly and say why); the block, line for line in body order. --branch and --pr each add one clause after the block. Two runs over one log print identical bytes.
  line G-5WZZH1#2: [P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Goal has checkpoint-store:G-D1CH1R's title and lines and a body ending in a block of run 5's constraint lines, the output matches the baseline clause for clause and differs from it only in: actor, owner answer, branch, PR, commit, the run's own plan; the dropped add-rival-Hypotheses clause; Results for Experiments in the count clause; and run 5's Treat-as-legacy sentence, which prints as its Read only: line.
  line G-5WZZH1#3: [N] `hyp prompt` exits 1, with a message naming the rule and nothing on stdout, for an unknown id, a node that is not a Goal, and a Goal body without a valid block. A valid block starts at the first line that begins at column 0 with May change:, Read only: or Never:, runs to the end of the body, holds all three headings, and every non-empty line in it starts at column 0 with one. A line of spaces, a tab or a lone carriage return is not empty.
  line G-5WZZH1#4: [N] `hyp add goal` warns (a warn: line on stderr, exit 0, naming the rule) for exactly the bodies the previous line makes `hyp prompt` refuse, a body with no heading included. It never refuses t
[... 908 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-5WZZH1 &&\nhyp add dq --parent $G --title \"What form of store path does hyp prompt print?\" \\\n  --body \"Scenario that separates the rivals: create a git worktree and a clone at other paths, run hyp prompt in each, and check that each printed path resolves to the store holding the Goal. Each rival is a commit that meets all eight lines and differs from the other only in this form. The owner does not pick between them.\" \\\n  --why \"Testable choice routed to a DesignQuestion; owner declined to choose\" 2>&1 | head -3",
 "description": "Add the store-path DesignQuestion"
}
```

Result:

```text
Q-A60ZRP node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-5WZZH1 && Q=Q-A60ZRP &&\nhyp add hyp --parent $Q --title \"Print the absolute path of the store root\" --body \"hyp prompt prints filepath.Abs of the directory holding .tree/events, as store.FindRoot returns it. It needs no git. Used from a worktree or clone at another path it prints that checkout's own path.\" --why \"Rival 1 of the store-path question\" 2>&1 | head -2 &&\nhyp add hyp --parent $Q --title \"Print the store path relative to the repository root; with no git toplevel, exit 3\" --body \"hyp prompt prints the store root relative to the git toplevel that holds it, so the same text names the store in any clone. With no git toplevel above the store it exits 3 (an environment problem).\" --why \"Rival 2 of the store-path question\" 2>&1 | head -2 &&\nhyp link $G external_ref gh:113 --why \"The issue this Goal builds\" 2>&1 | head -2 &&\nhyp checkpoint --goal $G --summary \"Grill ended: one build Goal with eight lines and one DesignQuestion with two rivals (store path form). Owner files later: the placeholder issue for matching run 6's prompt, and the one-root-store issue. hyp prompt does not exist in the binary yet, so the run's /goal prompt is written by hand.\" --why \"end of grill\" 2>&1 | head -2 &&\nhyp status $G 2>&1 | head -30; hyp check 2>&1 | tail -3",
 "description": "Add rivals, link, checkpoint, then status and check"
}
```

Result:

```text
H-9CXWQB node.add
H-S3BZCN node.add
G-5WZZH1 edge.add
G-5WZZH1 checkpoint
GOAL G-5WZZH1 hyp prints the /goal prompt for any Goal from its log, and … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-5WZZH1#1 no candidate — [P] For a Goal whose body ends in a valid constraints block, `hyp prompt G --actor A` exits 0 and prints in order: the store path; th…
  G-5WZZH1#2 no candidate — [P] Baseline run-5-prompt.txt, sha256 840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3. On a synthetic store whose Go…
  G-5WZZH1#3 no candidate — [N] `hyp prompt` exits 1, with a message naming the rule and nothing on stdout, for an unknown id, a node that is not a Goal, and a G…
  G-5WZZH1#4 no candidate — [N] `hyp add goal` warns (a warn: line on stderr, exit 0, naming the rule) for exactly the bodies the previous line makes `hyp prompt…
  G-5WZZH1#5 no candidate — [N] `hyp prompt` exits 2 for an --actor outside the replay-spec name pattern or missing; a --branch that is empty, starts with -, or …
  G-5WZZH1#6 no candidate — [P] The output depends only on the merged projection. In a two-Goal store, `hyp prompt G-a` names only G-a's id and block and points …
  G-5WZZH1#7 no candidate — [P] For a Goal that is complete, or whose OWED already says the runner owes nothing, `hyp prompt` still prints and exits 0.
  G-5WZZH1#8 no candidate — [N] With no store (.tree/events) at or above the working directory, `hyp prompt` exits 3 and prints nothing on stdout.
FRONTIER
  Q-A60ZRP open/untested What form of store path does hyp prompt print?
    H-9CXWQB proposed Print the absolute path of the store root
    H-S3BZCN proposed Print the store path relative to the repository root; with …
STALE
OWED
  runner: Q-A60ZRP test H-9CXWQB
  runner: Q-A60ZRP test H-S3BZCN
AWAITING OWNER
RECENT
  2026-11-12 14:42 claude-grill-113 edge.add G-5WZZH1 — The issue this Goal builds
CHECKPOINT 2026-11-12 14:42 claude-grill-113: Grill ended: one build Goal wit
[... 223 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-a-2/wt && git add investigations/goal-prompt/.tree/events investigations/goal-prompt/.tree/evidence && git status --short && git commit -q -m \"docs: record the grill for #113 in investigations/goal-prompt\" && git log --oneline | head -1",
 "description": "Commit the new tree's events and evidence"
}
```

Result:

```text
A  investigations/goal-prompt/.tree/events/claude-grill-113.jsonl
A  investigations/goal-prompt/.tree/events/owner-standin.jsonl
A  investigations/goal-prompt/.tree/evidence/baseline/run-5-prompt.txt
?? investigations/goal-prompt/.tree/.gitignore
1737485 docs: record the grill for #113 in investigations/goal-prompt
```

## Grill

The grill is written and committed as `1737485` on branch `work`. `hyp check` prints `0 errors, 0 warnings`.

## What is in the tree

- **Store:** `investigations/goal-prompt/`
- **Goal:** `G-5WZZH1`, with eight lines. The owner write used `owner-standin` and your approved command.
- **DesignQuestion:** `Q-A60ZRP`, "What form of store path does hyp prompt print?" with two rivals:
  - `H-9CXWQB`: print the absolute path.
  - `H-S3BZCN`: print the path relative to the repository root, or exit 3 with no git toplevel.
- **Link:** `external_ref gh:113`.
- **Baseline:** `.tree/evidence/baseline/run-5-prompt.txt`, sha256 `840d813d6f8e6b0c1502f3d629fc091b551dd90e5f823e5190ee35b055d7e3a3`.
- **Checkpoint:** "end of grill" is recorded.

I left `.tree/.gitignore` out of the commit. It is untracked, and I did not check whether other trees commit theirs.

`hyp status G-5WZZH1` shows the runner owing a test of each rival, nothing awaiting you, and every line still without a candidate.

## One change from what you approved

Line 2 now reads `checkpoint-store:G-D1CH1R`, because the store refuses a bare id from another tree. You approved that change before I ran it.

## Hand-off notes

`hyp prompt` does not exist yet, so I wrote the `/goal` prompt by hand in the grill skill's shape. It follows your answers: B's clause set with N = 5, the Results count, and `--branch` and `--pr` left out until the run has them. Replace it with `hyp prompt G-5WZZH1 --actor <runner-actor>` once the command exists.

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=<runner-actor>. Its Goal is G-5WZZH1.
Done when the latest printed `hyp status G-5WZZH1` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under <runner-actor> since this goal began, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
May change: this repository, as commits on one local branch
May change: the texts the change makes wrong, by directory: skills/, docs/, AGENTS.md, README.md, including a skill description that names only one kind of Goal
Read only: investigations/checkpoint-store and investigations/checkpoint-store-design
Never: open or merge a PR
Never: replace the installed hyp binary; build to a temporary path
Never: edit a Goal or an event already in a log
```

This text was written by hand, not generated. Use a runner name different from `claude-grill-113`.

**For you to file later (I filed nothing):**
- **Placeholder issue.** "Owner flow: a generated /goal prompt matches run 6's prompt clause for clause". Its text is in my earlier message. It says "Part of #113", so #113 stays open until you file it.
- **Root-store issue.** "Where do Goals live: one root store, or one store per topic?" The draft is in my round 2 message.

**Facts the run should know**
- **Exit 3.** The installed grill skill says exit 3 means "no repository holds the store". The code only walks up to a `.tree/events` directory, so the real condition is no store at or above the working directory. Line 8 states it that way, and item (d) corrects the skill text.
- **The id-cite rule.** A bare id from another tree is refused, so any later text that names one must be written `<tree>:<id>`.
- **Body limit.** A Goal body holds at most 2000 runes. Ours is 987.
