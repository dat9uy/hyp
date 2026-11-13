# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-a-1/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and checking the store.

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
 "command": "(hyp status 2>&1 | head -40); ls; ls .tree 2>&1 | head; ls docs | head -50"
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
ls: .tree: No such file or directory
adr
core-and-self-review.md
glossary.md
how-hyp-improves-itself.md
install.md
principles.md
replay-spec.md
sandbox-spec.md
```

## Grill

No store exists here, and the issue's prompt generator may already be in the skill text. I'll check what's real in the code and docs.

### Tool Bash (cwd: .)

```text
{
 "command": "which hyp; hyp 2>&1 | head -40; grep -rn \"prompt\" cmd internal --include=*.go -il | head; ls investigations skills skills/*; git log --oneline | head -15; wc -l docs/*.md docs/adr/* | tail -40"
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-base-a-1/bin/hyp
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
  render --markdown [--as-of REF|TIME]        the whole tree as Markdown, untruncate
[... 4046 more characters]
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-1/wt/docs/glossary.md"
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
 "command": "grep -rn \"hyp prompt\\|May change\\|constraints block\" --include=* -r . -l 2>/dev/null | grep -v \"^./.git/\" | head -20; ls cmd/hyp internal; echo; cat skills/hyp-goal-run/SKILL.md; cat docs/adr/0021-a-skill-change-is-a-build-goal.md"
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
- The recipe's 
[... 13340 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rniE \"hyp prompt|May change|constraints block\" . --exclude-dir=.git -l | head -20; echo ---; grep -rn \"\\\"prompt\\\"\" internal cmd | head; echo ---; ls internal/cli internal/render investigations/checkpoint-store investigations/checkpoint-store-design; git log --oneline -3 -- skills/hyp-goal-run; sed -n 1,80p docs/adr/0020-one-store-holds-many-goals.md"
}
```

Result:

```text
./investigations/recipe-base/.tree/evidence/E-3S24B5/coarse-3/transcript.jsonl
./investigations/checkpoint-store/.tree/events/claude-g1b-5.jsonl
./.claude/skills/hyp-prototype/SKILL.md
./.claude/skills/hyp-owner-grill/SKILL.md
./.claude/skills/ak-skill-creator/references/writing-effective-instructions.md
./.claude/skills/hyp-goal-run/SKILL.md
./docs/adr/0019-scope-change-folds-into-goal-lines.md
./docs/adr/0002-known-answers-gate-every-recipe-run.md
./docs/adr/0015-the-select-is-the-gate.md
./docs/adr/0011-a-confirm-counts-while-its-dossier-stands.md
./skills/hyp-prototype/SKILL.md
./skills/hyp-owner-grill/SKILL.md
./skills/hyp-goal-run/SKILL.md
---
---
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

internal/render:
markdown.go
render_test.go
show.go
showlimit.go
status.go
testdata

investigations/checkpoint-store:

investigations/checkpoint-store-design:
ee3828f feat: publish the owner-flow skills in skills/
---
status: proposed
---

# One store holds many Goals

Until now a tree had one Goal (spec §5 rule 4: "only one Goal per log"). A new purpose was a new tree, and that tree cited the old one in prose as `<tree>:ID`, which hyp does not check (ADR 0016). After the multi-tenant tree completed on 2026-11-10, the owner wanted the build that follows a tree to run from the tree instead of from a plan in prose. Every problem with that came from the one-Goal rule:

- A build had no checked way to cite what
[... 6723 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1 && ls; ls tmp; find . -name \"*goal-prompt*\" -not -path \"./wt/.git/*\" 2>/dev/null | head; cd wt/investigations/checkpoint-store && hyp status 2>&1 | head -70"
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
  2026-1
[... 1293 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt && grep -n \"prompt\" skills/*/SKILL.md | cut -c1-300"
}
```

Result:

```text
skills/hyp-goal-run/SKILL.md:3:description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report
skills/hyp-goal-run/SKILL.md:6:The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how 
skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a 
skills/hyp-goal-run/SKILL.md:68:Stop on the clauses the `/goal` prompt names. Stopping is not success.
skills/hyp-owner-grill/SKILL.md:35:| how the tree is run (process, review, order, cadence) | the Goal body, an Experiment's verdict rule, or the `/goal` prompt; never a line | only where it is a preference |
skills/hyp-owner-grill/SKILL.md:51:- **A line carries its own cases.** A line that compares output with a baseline names the baseline by sha256 and asserts that it differs from the candidate; other lines name no baseline. Name each exception it allows in its own words, not only in a later prompt.
skills/hyp-owner-grill/SKILL.md:96:The Goal body ends with the run's constraints, one heading per line, so a run and a prompt generator can find them:
skills/hyp-owner-grill/SKILL.md:108:Write the `/goal` prompt for hyp-goal-run, under 4000 characters: the tree, the runner's actor, the Goal id, the stop clauses, and the Goal body's `May change:`, `Read only:` and `Never:` lines copied as they are. Put general practice in the skill, not the prompt.
skills/hyp-owner-grill/SKILL.md:114:
[... 115 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls -la .claude/skills; for s in hyp-goal-run hyp-owner-grill hyp-prototype; do echo \"== $s\"; diff <(cat skills/$s/SKILL.md) <(cat .claude/skills/$s/SKILL.md) | head -60; done; sed -n 90,120p skills/hyp-owner-grill/SKILL.md"
}
```

Result:

```text
total 0
drwxr-xr-x@ 8 datguy  staff  256 Nov 12 21:19 .
drwxr-xr-x@ 4 datguy  staff  128 Nov 12 21:19 ..
drwxr-x---@ 9 datguy  staff  288 Nov 12 21:19 ak-skill-creator
drwxr-xr-x@ 3 datguy  staff   96 Nov 12 21:19 hyp-goal-run
drwxr-xr-x@ 3 datguy  staff   96 Nov 12 21:19 hyp-owner-answer
drwxr-xr-x@ 3 datguy  staff   96 Nov 12 21:19 hyp-owner-grill
drwxr-xr-x@ 4 datguy  staff  128 Nov 12 21:19 hyp-prototype
drwxr-xr-x@ 4 datguy  staff  128 Nov 12 21:19 hyp-verification
== hyp-goal-run
diff: /dev/fd/11: Operation not permitted
== hyp-owner-grill
diff: /dev/fd/11: Operation not permitted
== hyp-prototype
diff: /dev/fd/11: Operation not permitted
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

A body that names another tree's ids writes each as `<tree>:<id
[... 1263 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1 && grep -rl \"goal-prompt\" . 2>/dev/null | head; ls ~/.claude/plans 2>/dev/null | head; cd wt && hyp add goal --title x --body \"May change: a\" --why t 2>&1 | head -5; grep -n \"func.*[Gg]oal\" internal/cli/write.go | head -20; sed -n 1,60p internal/cli/help_test.go | head -40"
}
```

Result:

```text
./tmp/issue-113.md
missing actor: pass --actor or set HYP_ACTOR
1040:func (w *writer) checkpointGoal(named string) (string, error) {
package cli

import (
	"os"
	"testing"
)

// groupSubs lists each group command's subcommands. A help request on any of
// them, or on the group itself, must exit 0 with usage on stdout.
var groupSubs = map[string][]string{
	"add":     {"goal", "dq", "hyp", "exp"},
	"bundle":  {"hash", "check", "new", "run", "evidence", "known-answers"},
	"sandbox": {"create", "drive", "remove"},
}

func TestHelpRequestExitsZeroOnStdout(t *testing.T) {
	t.Chdir(t.TempDir())
	var cases [][]string
	for _, name := range Commands() {
		for _, flag := range []string{"-h", "--help"} {
			cases = append(cases, []string{name, flag})
			for _, sub := range groupSubs[name] {
				cases = append(cases, []string{name, sub, flag})
			}
		}
	}
	for group := range groupSubs {
		cases = append(cases, []string{group, "help"})
	}
	for _, args := range cases {
		r := hyp(t, args...)
		if r.code != 0 || r.stdout == "" || r.stderr != "" {
			t.Errorf("hyp %v: exit %d, stdout %q, stderr %q", args, r.code, r.stdout, r.stderr)
		}
	}
	if _, err := os.Stat(".tree"); !os.IsNotExist(err) {
		t.Fatalf("a help request created .tree (stat err %v)", err)
	}
}
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"1.3\\|exit 1\\|exit 3\" docs/sandbox-spec.md | head -8; grep -n \"Body\" internal/replay/*.go | grep -i goal | head -8; grep -n \"warn\" internal/cli/write.go | head -8; grep -n \"Commands()\" -A12 internal/cli/cli.go | head -30; sed -n 1,30p docs/core-and-self-review.md | cut -c1-200"
}
```

Result:

```text
28:`hyp bundle run` and `hyp bundle known-answers` find REF in `recipes/` or `rubrics/`; a rubric is a bundle with one Standalone command and no step (replay spec §7.6), so `hyp bundle evidence` refuses it with exit 2, and a ref found in both libraries is refused with exit 1. `hyp recipe` runs as a hidden alias of `hyp bundle`, because frozen bundles call it from their scripts (ADR 0018).
32:A sandbox is a directory plus its record, not process isolation. Each command hyp starts leads its own process group, which its row records as `pgid` (ADR 0002). Before Cleanup, `hyp sandbox remove` stops the groups of commands a killed hyp left unfinished (§4); after Cleanup, it stops the groups of the Launch, Doctor, Drive and Cleanup commands that still have members: SIGTERM, up to 5 seconds, then SIGKILL. A timeout or Ctrl-C kills the running command's group (its row says `interrupted` for Ctrl-C), and an Evidence or Standalone group ends with its command; create still runs Cleanup after a Ctrl-C. Limits: a daemon that leaves its group, as Postgres's postmaster does by calling `setsid`, still needs its recipe's Cleanup, and the leftover check still applies; a group id reused after its group ended is not told apart; what a check cannot see is listed under [replay-spec.md](replay-spec.md) §7.6 and ADR 0003. "Runs in a sandbox" means only that the command runs while the sandbox exists and gets that sandbox's `HYP_SANDBOX_*` variables (§1.3).
36:hyp runs every command in a bundle, and the operator runs none of a bundle's scripts directly. A command is what a `Run:` line names; every other file in the bundle, a helper script included, is hashed content (ADR 0002). Hashing never parses: a bundle whose `recipe.md` breaks a command rule, such as a `Run:` line outside a section or shell syntax, still hashes, so every pin to it still verifies, and `hyp check` names it as a bundle that does not open. `hyp bundle new` writes `format: 1` into every bundle it makes, a first version from i
[... 7165 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -60 | cut -c1-900"
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
[... 7556 more characters]
```

## Grill

Round 1 is below. No store exists in this repository, so nothing is written yet, and `hyp prompt` is not in the installed binary.

**What I found**
- `skills/hyp-owner-grill/SKILL.md` in the repo still tells a run to hand-write the `/goal` prompt, under 4000 characters. The copy I'm running names `hyp prompt <G-id> --actor … [--branch] [--pr] [--experiments]`, which `hyp` doesn't have. So this Goal builds that command and the warning on `hyp add goal`. #113 is core scope. The owner flow's skills and docs make claims the command doesn't back yet.
- **G-D1CH1R's body has no `May change:` / `Read only:` / `Never:` block.** A Goal is never edited, so a generator that reads the recorded block must refuse it. The Done-means line "generated prompt for G-D1CH1R matches run 6's" therefore can't be met from the recorded place.
- Run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) is not on this machine. I only have the run-5 text quoted in the issue.

I'll put the tree in `investigations/goal-prompt/`, with a new `hyp init` there. Each tree here has its own directory, and the repo root has no store.

**Questions for you**

1. **Your actor name** (for example `owner-dat`).

2. **How is Done-means 1 pinned?**
   - Today it says "matches run 6 clause for clause". G-D1CH1R can't produce a prompt, because it has no block, and I don't have run 6's text.
   - A: you paste run 6's prompt. I pin it as a synthetic fixture under `internal/*/testdata`: G-D1CH1R's body plus a block holding run 6's `Never:` lines. Line 1 then asserts the exact bytes. *Example:* the fixture's `Never: never open or merge a PR` appears in the output verbatim.
   - B: drop "matches run 6". The exact text is whatever you approve in Q3, and a Goal with no block is refused.
   - I recommend A, because B loses the check against a real hand-written prompt. Core tests may read only synthetic data, so the fixture must be synthetic either way.

3. **Which stop clauses does the printed prompt carry?** Run 5's had four:
   - (a) the runner owes nothing
   - (b) 5 Experiments recorded
   - (c) "cannot take the next step honestly"
   - (d) only `add rival Hypotheses` items are owed, with no owner checkpoint asking for a new one
   - The copy of the skill I'm running names only (a) and "N Results under the actor, 5 by default".
   - I recommend (a), (b) and (c). Clause (d) is tied to the multi-tenant tree's discovery state. Run 6 also had a Goal-specific clause, "selected the candidate on Q-PC04ED", which can't be derived from the Goal alone.
   - **One word for one thing.** The clause counts Results the runner's actor recorded, which hyp derives from the log. The flag is `--experiments`, though an Experiment and a Result are different nodes in the glossary. I recommend `--results`.

4. **What do `--branch` and `--pr` add?** The rule is that a prompt may narrow the block, never widen it.
   - `--branch feat/x` adds "Commit only to branch feat/x." That narrows `May change:`.
   - `--pr 140` could add "Update PR #140; never merge it." That is allowed only if `May change:` already allows pushing.
   - I recommend `--branch` as a narrowing clause only, and `--pr` as a narrowing clause that never states anything the block lacks. If the block doesn't allow a PR, `hyp prompt` refuses the flag with exit 2 or 1. Do you agree, or should the clauses be fixed text?

5. **The constraints block for this Goal.** My draft:
   ```
   May change: this repository (hyp): commits on one local branch, and a PR that is opened and never merged; and the texts the change makes wrong (skills/hyp-goal-run, skills/hyp-prototype, skills/hyp-owner-grill, README, the glossary, the specs, a new ADR 0022).
   Read only: investigations/checkpoint-store and the other published trees; GitHub is not reachable.
   Never: replace the installed ~/.local/bin/hyp; edit a stored Goal or event; add a typed constraints field to `hyp add goal`.
   ```
   - The typed field is your 2026-11-11 decision to wait until a real run misreads or loses a block. That is Goal body text, not a line.
   - Is a new ADR 0022 wanted, recording that decision? I recommend yes, short.

6. **How are the skill-text lines tested?** ADR 0021 leaves this to the Goal. The skill edits are the `hyp-goal-run` and `hyp-prototype` descriptions, which currently name one Goal kind, and the `hyp-owner-grill` Hand off section, which will use `hyp prompt`.
   - A: a `go test` over `skills/*/SKILL.md`. It checks that the descriptions name no single Goal kind, that the Hand off names `hyp prompt`, and that the existing skillcheck passes with `hyp prompt` present. Cheap and deterministic.
   - B: an ak-skill-creator eval, as ADR 0021 mentions.
   - I recommend A. Wording a model might misread isn't what these lines pin.

**Planned DesignQuestions** (not your choices; you only say whether the run should test each)
- **Generator:** a `hyp prompt` command, or a prompt the skill writes. *Scenario:* a Goal with a malformed block. A command exits 1 naming the rule, and a skill can't. Two calls on a merged log give the same bytes.
- **A Goal with no block (G-D1CH1R):** refuse with exit 1, print with no constraint lines, or print an explicit "no constraints recorded". *Scenario:* G-D1CH1R in a store. I recommend refuse, because a prompt with no limits reads as permission.
- **`hyp add goal` on a body with a heading but no valid block:** warn at exit 0 or refuse at exit 1. *Scenario:* add the malformed Goal and see whether a Goal that can never be prompted now exists.

**Draft lines**, all `[P]`/`[N]` for the hyp commit:
- `[P]` For a valid block, `hyp prompt` prints the pinned text, derived from the log alone and stored nowhere. It gives the same bytes on a merged log and writes no event.
- `[N]` For a Goal with no valid block, `hyp prompt` exits 1 and names the rule.
- `[N]` A prompt never drops, reorders or widens a block line.
- `[N]` Exit codes: 2 for bad arguments, 3 when no repository holds the store.
- `[P]` `hyp add goal` reports a malformed block as the DesignQuestion settles.
- `[P]` The skill lines from Q6.

Everything you've seen here is routed in the table I'll show once you've answered.

## Owner, message 2

1. My actor name is owner-standin.

2. I won't choose between A and B, because that is eval detail. Read the ak-skill-creator eval and audit flow and decide from it. On matching run 6, widen the exceptions to the actor, the owner's answer, branch, PR and commit, and the run's own plan. Branch and PR become flags. A Goal with no block is refused with exit 1 and the missing headings named, but adding such a Goal stays allowed.

3. Accept (a), (b) and (c). The prompt doesn't carry the owner's answers, because the runner reads OWED. The cap is a flag with a default of 5. I accept `--results` as the name, given your glossary point.

4. Agree. Make `--branch` and `--pr` flags that only narrow.

5. Use the strict form, with one constraint per line. Split your long `May change:` line into separate lines. The skill descriptions that name one kind of Goal go under May change, with no line of their own. Yes to a short ADR 0022.

6. Same as Q2: I won't pick A or B. Read the ak-skill-creator flow and decide. My only preferences are Sonnet for the test agents, 3 trials, and a modest budget.

**DesignQuestions**
- **Generator:** a hyp command that reads the log. Settle it in the session and don't test it.
- **Goal with no block:** refuse with exit 1.
- **`hyp add goal` on a malformed block:** warn at exit 0, because adding stays allowed.

Your draft lines are fine.

## Grill

I'll read the ak-skill-creator eval and audit flow before choosing the skill-text instrument.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/.claude/skills/ak-skill-creator && find . -type f | head -50; wc -c SKILL.md; cat SKILL.md | head -200"
}
```

Result:

```text
./references/writing-effective-instructions.md
./references/troubleshooting-guide.md
./references/benchmark-optimization-guide.md
./references/plugin-marketplace-hosting.md
./references/mcp-skills-integration.md
./references/skillmark-benchmark-criteria.md
./references/structure-organization-criteria.md
./references/creator-consumer-evaluation.md
./references/plugin-marketplace-overview.md
./references/validation-checklist.md
./references/skill-creation-workflow.md
./references/script-quality-criteria.md
./references/production-skill-standard.md
./references/skill-design-patterns.md
./references/script-dependency-strategy.md
./references/long-horizon-scaffold-templates.md
./references/token-efficiency-criteria.md
./references/plugin-marketplace-troubleshooting.md
./references/distribution-guide.md
./references/agentkit-kit-skill-contract.md
./references/metadata-quality-criteria.md
./references/skill-anatomy-and-requirements.md
./references/plugin-marketplace-sources.md
./references/long-horizon-repository-architecture.md
./references/cross-marketplace-distribution.md
./references/yaml-frontmatter-reference.md
./references/prompt-cruft-patterns.md
./references/skill-ecosystem-portability-and-safety.md
./references/evaluation-tools.md
./references/plugin-marketplace-schema.md
./references/long-horizon-building-loop.md
./references/semantic-routing-evaluation.md
./references/session-usage-analysis.md
./references/testing-and-iteration.md
./references/advisory-supervision.md
./LICENSE-MIT.txt
./NOTICE.txt
./scripts/resource_validation.py
./scripts/init_skill.py
./scripts/skill_standard_report.py
./scripts/lint_cruft.py
./scripts/package_skill.py
./scripts/quick_validate.py
./scripts/requirements.txt
./scripts/session_usage.py
./scripts/semantic_routing_providers.py
./scripts/eval_skill.py
./scripts/semantic_routing_contract.py
./scripts/encoding_utils.py
./scripts/eval_contract.py
    9510 SKILL.md
---
name: ak-skill-creator
description: "Create, update, audit, validat
[... 9438 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/.claude/skills/ak-skill-creator/references && cat testing-and-iteration.md | head -150; echo ======; sed -n 1,120p evaluation-tools.md"
}
```

Result:

```text
Exit code 1
# Testing and iteration

Evaluate what a consumer accomplishes. Separate structural validity, routing,
functional quality and cost. Begin with a small case set and expand from failures.

## Capture cases before optimizing

Use `assets/eval-cases.example.json` as the concrete contract. Author cases at
`evals/evals.json` in the target skill source; disposable results live beside it
or in the project's reports directory. Cases record id, split (train or holdout),
realistic prompt, expected_output, optional files and assertions with stable ids/text.
Use fabricated or approved redacted fixtures and preserve accepted user constraints.

Cover a normal task, boundary and failure/recovery when applicable. For mutations,
check input preservation, contained edits, retry behavior and honest partial status.
Keep holdout prompts outside author/rewrite context, with the split fixed across
iterations. The bundled public example demonstrates the contract; it is not an
independent holdout for evaluating this creator.

## Functional comparison

1. Snapshot before updating. Create compares no-skill with candidate; update compares
   original with candidate and reruns cases for behavior that must remain unchanged.
2. Use fresh sessions and clean workspaces with identical prompts, tools, fixtures
   and settings. Ensure the baseline cannot discover the candidate via global or
   project skills or inherited context. Consumers do not receive expected answers.
3. Record effective model id, runtime/version, snapshot hash, run/session id,
   available tools, input version and output location. Requested aliases are not
   evidence of the effective model; record what the provider reports.
4. Save artifacts and redacted execution traces. Check mechanical properties with
   code and judgment-dependent quality with an independent evidence-based rubric.
   Compare anonymized outputs when using a judge. Success claims are not artifacts.
5. Record tokens, duration, tool calls, retries and
[... 4879 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "sed -n 1,140p references/evaluation-tools.md; python3 -I scripts/eval_skill.py --help 2>&1 | head -30; ls assets 2>&1 | head"
}
```

Result:

```text
# Evaluation tooling

`scripts/eval_skill.py` uses standard-library Python. It validates cases, grades
existing artifacts and summarizes records; it never calls providers or executes
commands. Run from this skill's root or use the script's absolute path from any cwd.

## Cases and assertions

'''bash
python3 scripts/eval_skill.py validate assets/eval-cases.example.json
python3 scripts/eval_skill.py validate /absolute/target-skill/evals/evals.json
'''

Assertions have `id`, `text`, and optional `kind`:

| Kind | Required fields | Evidence |
|---|---|---|
| `file-exists` | `path` | Contained file exists |
| `json-equals` | `path`, `expected` | Parsed artifact matches expected JSON |
| `text-contains` | `path`, `expected` string | Required contract fragment exists |
| `rubric` (default) | `text` | Independent review with cited artifact evidence |

Use exact equality for machine contracts, not wording preferences. Paths are
relative to the output root; escaping paths and symlinks fail. Missing or malformed
artifacts fail. Rubric checks stay ungraded until reviewed. For richer checks,
run the existing domain verifier and retain its command, exit status and output.

## Run and grade

Discover available model runners and current flags from their help. Use a fresh
context with explicit output location. Do not assume provider credentials, model
names, isolation flags or delegation tools. If unavailable, record behavioral
verification as blocked and continue structural work with that limitation stated.

After an actual consumer run:

'''bash
python3 scripts/eval_skill.py grade assets/eval-cases.example.json \
  --case-id normalize-valid --outputs /absolute/run/outputs \
  --run-id unique-session-id --variant candidate \
  --model effective-model-id --runtime observed-runtime-version \
  --skill /absolute/skill-snapshot > /absolute/run/record.json
'''

No-skill: `--variant without_skill`, omit `--skill`. Other variants require a hash.
Optional observed metrics: `--tokens`, `--
[... 4687 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt && grep -rn -i \"/goal\\|goal prompt\" README.md docs/*.md docs/adr/*.md skills/*/SKILL.md | cut -c1-260 | head -50; ls internal/skillcheck; grep -n \"discovery\" skills/*/SKILL.md | cut -c1-200"
}
```

Result:

```text
README.md:164:For a Goal you want run unattended, discovery or build, three skills split the work. They run in Claude Code, whose `/goal` command keeps a session going until its condition holds.
README.md:166:1. **`/hyp-owner-grill`** questions you about the purpose until it is a Goal a run can finish. It looks up the facts itself, sends questions a run could test to DesignQuestions with rivals, derives the lines from what must and must never happen, 
README.md:167:2. **`/goal <that prompt>`** runs the Goal with the `hyp-goal-run` skill. The run does only what OWED says it owes, tests before it asks, and stops with a report you can read from `hyp status`.
README.md:168:3. **`/hyp-owner-answer`** walks you through what the stopped run left under AWAITING OWNER, explains each item and records only your answers. The next `/goal` run picks up from there.
docs/how-hyp-improves-itself.md:30:  stated by the `/goal` prompt or the recipe, not by hyp.
docs/adr/0017-the-ledger-is-the-stall-guard.md:20:- **A run's cap stays outside hyp.** A `/goal` prompt or a recipe states how many Experiments a run may take.
docs/adr/0021-a-skill-change-is-a-build-goal.md:9:AGENTS.md gated skill text apart from code. A skill-text change shipped only when a session replay from the event's store cut showed the failure in 3 or more of 5 trials without the change and in 1 or fewer of 
skills/hyp-goal-run/SKILL.md:3:description: "Run a hyp tree under an owner's Goal and met rule as an unattended discovery /goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the
skills/hyp-goal-run/SKILL.md:6:The `/goal` prompt names the tree, your actor, the Goal and the stop. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the
skills/hyp-goal-run/SKILL.md:12:- The Goal's body states the run's constraints under three headings: 
[... 2423 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 26,32p docs/how-hyp-improves-itself.md; sed -n 16,24p docs/adr/0017-the-ledger-is-the-stall-guard.md; sed -n 160,172p README.md | cut -c1-700; grep -n \"Part of\\|exit\" docs/replay-spec.md | sed -n 1,12p | cut -c1-160"
}
```

Result:

```text
- **The ledger.** `hyp status` lists under OWED what the tree owes and who
  owes it, and marks an item an Experiment aimed at and did not settle
  ([replay-spec.md](replay-spec.md) §8.1). It is the one account of progress
  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
  stated by the `/goal` prompt or the recipe, not by hyp.
- **Design marks.** `hyp signals` lists the Experiments set up against no live
  rival (one-rival) or against a dead one (D1). A Result that leaves a

- **`wasted`, `pending`, `no-movement` and `spin` are removed.** A second account of progress either agrees with the ledger, and adds nothing, or disagrees with it, and is a false alarm.
- **`one-rival`, `D1` and `unjudged` stay.** Each is already a write-time refusal or warning, or a mark that `hyp status` and `hyp show` print. `hyp signals` lists them for the whole tree.
- **`hyp signals --json` keeps the facts of each round:** its Experiment, cut, parent, targets, question, statuses at add, its marks and its Results with their status moves. It drops `pending`, `wasted`, `closed`, `line` and `moved`.
- **A run's cap stays outside hyp.** A `/goal` prompt or a recipe states how many Experiments a run may take.

## Consequences

- Old trees print fewer signal lines. No event, payload, status or ledger item changes.
From there the [hyp-prototype skill](skills/hyp-prototype/SKILL.md) guides the next rounds: experiments, verdicts, selects and checks. When `hyp status` lists something under AWAITING OWNER, it is yours to answer. [example.sh](skills/hyp-prototype/example.sh) runs a complete investigation in a throwaway directory.

### The owner flow: grill, run, answer

For a Goal you want run unattended, discovery or build, three skills split the work. They run in Claude Code, whose `/goal` command keeps a session going until its condition holds.

1. **`/hyp-owner-grill`** questions you about the purpose until it is a Goal a run can finish. It looks up the facts itself, send
[... 2696 more characters]
```

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader on prompt rules",
 "subagent_type": "general-purpose",
 "run_in_background": true,
 "prompt": "You are an independent reviewer of a design for a new command in the Go repo at /Users/datguy/.hyp-sandbox/g146-base-a-1/wt (hyp: an append-only event store under .tree/ with a replay engine). Do NOT edit any file. Read docs/glossary.md, docs/adr/0020-one-store-holds-many-goals.md, docs/replay-spec.md (§4, §5, §8), internal/cli/cli.go, internal/cli/write.go, internal/event (Goal body limit, MaxGoalBody), internal/store (how the store/repo is found), internal/skillcheck.\n\nRULES under review (settled with the repo owner):\n1. New command: `hyp prompt <G-id> --actor A [--branch B] [--pr P] [--results N]` (N default 5). Read-only: it appends no event and writes no file. Everything it prints is derived from the merged log plus its flags; nothing new is stored.\n2. It prints a `/goal` prompt: the store's path, the actor A, the Goal id, the skill name hyp-goal-run, stop clauses (the runner owes nothing under OWED; or N Results recorded under actor A; or the run cannot take the next step honestly), then the Goal body's constraints block line for line, then one clause for --branch (\"Commit only to branch B.\") and one for --pr (\"Do not merge PR #P.\"). It carries no owner answers or tree state.\n3. The constraints block of a Goal body runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body. It must contain each of the three headings at least once, and every non-empty line in it must start with one of them. A line of spaces, a tab or a lone carriage return is not empty. A heading may repeat.\n4. `hyp prompt` on a Goal with no valid block (no heading, a missing heading, or a non-heading non-empty line inside the block) exits 1 and names the missing headings; stdout stays empty. Bad arguments (branch with a space, unknown or non-Goal id, bad --results) exit 2. No repository holding the store: exit 3.\n5. `hyp add goal` still appends a Goal whose body has a heading but no valid block, exits 0 and prints a `warn:` line naming the rule. A Goal is never edited afterwards.\n6. A store can hold many Goals (ADR 0020); OWED, lines and checkpoints are per Goal.\n\nTASK: walk each rule through scenarios the owner did not draw, at least: (a) two Goals in one store, both with blocks, and `hyp prompt` for one of them; a body of one Goal that names the other's id; (b) a merged log from two actors, where the same Goal add appears in an events file from an actor other than the owner, and Results recorded by actor A in another worktree's log; what 'N Results under actor A' counts when a Result is retracted or its Experiment withdrawn; (c) the base or branch moving mid-run, and a prompt generated from a log that does not yet contain the Goal's add (unmerged branch); (d) a body with CRLF line ends, a trailing newline, a heading mid-prose, a body at the maximum length, headings with different case, a `N
[... 441 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a62a9f583598cd85a (internal ID - do not mention to user. Use SendMessage with to: 'a62a9f583598cd85a', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-1-wt/ea4155e0-5290-43bb-8480-42acd6eefa1d/tasks/a62a9f583598cd85a.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

## Grill

I read ak-skill-creator's eval flow and have settled the skill-test choice. Here is round 2, with a second reader still running in the background.

**How I decided the skill-test instrument**
- ak-skill-creator says to check mechanical properties with code, and instruction changes with matched original-vs-candidate runs by fresh consumers.
- You said the description edits get no line, so only one skill's behaviour changes: the `hyp-owner-grill` Hand off. That gets one eval:
  - 2 cases: a Goal with a valid block, and a Goal `hyp prompt` refuses.
  - 3 Sonnet trials per case, on the candidate skill and on the base skill: 12 runs.
  - The runner spawns the Sonnet consumers as fresh subagents, and `eval_skill.py` grades the saved outputs.
  - The cases live in `recipes/hyp-owner-grill-eval@1`, and the script is named on the command's `External:` line, as ADR 0021 says.
  - No routing traces and no holdout, to keep the budget modest.
- One gap I'll state in the Goal body, not hide: ak asks for routing traces on a metadata change, and the description edits get none. They keep their `Use when…` sentences unchanged.
- The bundle tests a hyp skill, so it is self-review. I'll add it to the scope map.

**Preflight finding.** ADR 0017 and `docs/how-hyp-improves-itself.md` say a run's cap is "stated by the `/goal` prompt or the recipe, not by hyp". Printing a default of 5 makes that false. I'll put both under May change, and ADR 0022 supersedes that line in part. hyp prints the cap and never enforces it.

**Questions**

1. **Run 6's text is still missing.**
   - You widened the exceptions to the actor, the owner's answer, branch, PR, commit and the run's own plan. That still compares against run 6's prompt, which isn't on this machine.
   - Paste its text or a path. I pin it as a synthetic fixture: G-D1CH1R's body plus a block holding run 6's `Never:` lines, with the run-specific parts left out.
   - Without it, the exact bytes of line 1 are my wording, and Done-means 1 is only checked against run 5's quoted text. Do you accept that fallback, or will you supply it?

2. **The exact flag clauses.** I propose pure restrictions that can't widen anything. `--branch feat/x` adds "Commit only to branch feat/x." `--pr 140` adds "Do not merge PR #140." Do you accept this text? Clause (c) would read: "Stop when you cannot take the next step honestly, and say why."

3. **A body with no heading at all.** `hyp add goal` allows it, as you said. Does it warn? I propose a `warn:` only when a heading is present but the block is invalid, so old-style bodies stay silent. `hyp prompt` still refuses them with exit 1.

**Routing table** (confirm each row or move it)

| Question | Goes to | Separating scenario |
|---|---|---|
| Generator: a `hyp prompt` command that reads the log, or a skill-written prompt | DesignQuestion, rivals "command" and "skill-written", settled in this session, not tested (Experiment `--recipe none`, then the select) | session evidence only: a malformed block exits 1 for a command, and a skill can't refuse; two calls give the same bytes |
| A Goal with no block | DesignQuestion, rivals "refuse, exit 1", "print without constraints", "print with 'none recorded'"; chosen: refuse | G-D1CH1R in a store |
| `hyp add goal` on a malformed block | DesignQuestion, rivals "warn at exit 0", "refuse at exit 1"; chosen: warn | add the malformed Goal and see whether prompting is possible |
| Stop clauses (a)(b)(c), `--results` default 5 | line 1's pinned text | |
| `--branch` and `--pr` only narrow | line 2 | |
| Skill-test instrument | Goal body, plus the verdict rule | |
| Description edits | May change only, no line | |
| Typed constraints field | Goal body `Never:`, no line | |
| Delivery needs of a commit Goal | #114, a separate Goal | |

**Lines, final draft**
1. `[P]` For a valid block, `hyp prompt` prints the pinned text. That text holds the store path, actor, Goal id, skill, clauses (a)(b)(c) with N=5, and the block line for line, with no owner answer or tree state.
2. `[P]` `--results N` changes only N. `--branch` and `--pr` each add exactly their one clause after the block and nothing else.
3. `[P]` The output depends only on the merged log and the flags, so the same bytes come from any actor-file order. The command appends no event and changes no file.
4. `[N]` A Goal with no valid block exits 1, names the missing headings, and prints nothing on stdout.
5. `[P]` `hyp add goal` with a heading but an invalid block appends the Goal, exits 0 and prints a `warn:` line naming the rule. A valid block prints none.
6. `[N]` Bad arguments (a branch with a space, an unknown or non-Goal id, a missing actor, a bad `--results`) exit 2, and no repository holding the store exits 3.
7. `[P]` At the candidate, each of 3 fresh Sonnet trials on a Goal with a valid block ends with `hyp prompt`'s output byte for byte, and adds no sentence.
8. `[N]` On a Goal that `hyp prompt` refuses, no trial hand-writes a prompt. Each reports the refusal to the owner.

**Goal body.** It states: the owner decision of 2026-11-11 (typed field waits), the ADR 0022 amendment of ADR 0017, the routing-trace gap above, and the 12-run eval budget.

The constraints block, one constraint per line:
```
May change: this repository (hyp): commits on one local branch
May change: one pull request, opened and never merged
May change: skills/hyp-goal-run and skills/hyp-prototype descriptions, which name one Goal kind
May change: skills/hyp-owner-grill, its Hand off section
May change: README.md, docs/glossary.md, docs/replay-spec.md, docs/core-and-self-review.md, docs/how-hyp-improves-itself.md
May change: a new docs/adr/0022 that supersedes in part ADR 0017's cap line
May change: a new recipes/hyp-owner-grill-eval@1
Read only: investigations/checkpoint-store and the other published trees
Read only: ~/.claude/skills/ak-skill-creator, which is named on the eval command's External: line and never copied
Never: replace the installed ~/.local/bin/hyp
Never: edit a stored Goal or event
Never: add a typed constraints field to `hyp add goal`
```

I'll show the exact write commands after you answer these three questions and the second reader reports.

## Owner, message 3

<task-notification>
<task-id>a62a9f583598cd85a</task-id>
<tool-use-id>toolu_011YyTqxXY4nBcWLZ91bjHzN</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-1-wt/ea4155e0-5290-43bb-8480-42acd6eefa1d/tasks/a62a9f583598cd85a.output</output-file>
<status>completed</status>
<summary>Agent "Second reader on prompt rules" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Gaps are ranked by severity. I read the source and did not run anything. Facts I could not verify are listed at the end.

## High

**1. "N Results recorded under actor A" has no baseline, and it counts the wrong scope.**
- Scenarios: (a), (b), (f).
- Why the rules fail:
  - Actors are long-lived (`ValidActor`, `internal/store/store.go:32,222`), and `Merge` keeps every event ever written (`internal/replay/replay.go:13`). The same actor's Results from earlier runs, or from another Goal in the same store, already satisfy the N-Results stop clause when the run starts. The run could stop at once.
  - Rule 6 says lines and OWED are per Goal, but the Result count in rule 2 is not scoped to a Goal.
  - Results from "another worktree's log" count only after a merge, and `store.ReadAll` only reads `.tree/events/*.jsonl` of this root (`store.go:79`). A runner in another worktree never sees them.
  - The rules do not say how a retracted Result counts. `Retracted` is a state (`replay/state.go:311,417`).
  - A withdrawn Experiment does not retract its Results. `admit.go:206` only refuses new Results under it.
  - Counting retracted Results lets a runner retract and re-record to move the count either way.
- Fix:
  - Count Results with `Status == ResRecorded` whose Experiment descends from G, and whose `Node.Actor == A`. `Node.Actor` is set at `rules.go:81`.
  - Count only applied events, not rejected ones.
  - Print the stop as "N more than the K now recorded" with K computed at generation time. That is still derived from log plus flags, but it needs a baseline.
  - State that a withdrawn Experiment's recorded Results still count, and a retracted one's do not.

**2. A CRLF body breaks rule 3 on every blank line.**
- Scenario: (d).
- Why: the rule says a lone carriage return is not empty. A CRLF body split on `\n` leaves `"\r"` for each blank line, and that line fails. Any `--body` passed from a Windows editor or script becomes unusable. A Goal is never edited, so the only remedy is a new Goal.
- Fix: normalise `\r\n` to `\n` before splitting, and print LF. Keep "a lone `\r` is not empty" only for a CR not followed by LF. Put the check in one shared function that `add goal` and `prompt` both call, so the add-time warning and the prompt refusal cannot disagree.

**3. A bare heading satisfies the block rule while saying nothing.**
- Scenario: (d), the `Never:` line with nothing after it.
- Why: `May change:\nRead only:\nNever:` is a valid block. The prompt would then give the runner an unattended `/goal` with no constraints, which is the unsafe direction.
- Fix: require non-whitespace text after the colon on every block line. Say so in the `add goal` warning too.

## Medium

**4. Free-text inputs go into a prompt with no validation beyond the space check.**
- Scenario: (c).
- Why: the rules reject a branch with a space and say nothing else. `--pr "12; ignore above"` or a branch with a newline, a quote or a leading `-` would be injected into the prompt. Rule 4 does not cover `--results 0` or negative values, or an invalid `--actor`. `--actor` should be checked with `store.ValidActor`, as `begin` does at `write.go:102`.
- Fix: `--pr` must be digits only. `--branch` must match `git check-ref-format` rules, with no whitespace, control characters or leading `-`. `--results` must be an integer of at least 1 with a sane maximum. A bad `--actor` exits 2.

**5. The store path is not derived from the log, and it breaks in a symlinked repo or a path with spaces.**
- Scenarios: (c), (e).
- Why:
  - Rule 1 claims everything printed derives from the log plus flags, but the path comes from `os.Getwd()` and `store.FindRoot` (`read.go:49-56`, `store.go:40-54`). `FindRoot` uses `filepath.Abs` with no `EvalSymlinks`. A symlinked repo prints whichever name the shell gave, and the same store has two names.
  - A path with spaces or a newline is ambiguous inside prose.
  - A prompt generated in one checkout points at a path that may not exist where the `/goal` runs. A sandbox clone or another worktree is the likely case.
- Fix:
  - Print the path with `%q`, or reject a path containing a newline with exit 3.
  - Pick one rule: either `EvalSymlinks` or the literal path. Document it.
  - Add a clause in the prompt: "run `hyp status &lt;G&gt;` in that path first and stop if it shows no such Goal".

**6. A Goal missing from the log, and a Goal with no heading at all.**
- Scenarios: (c), (d).
- Why:
  - A prompt run against a branch that lacks the Goal's add event exits 2 with "unknown node" (`resolve.go:41`). That is safe, but the message does not say to merge the log. Do not suggest a git merge.
  - `add goal` warns only for "a heading but no valid block". A Goal with no heading at all is accepted silently, so the owner learns of the problem only when `prompt` exits 1. Old Goals, such as `investigations/checkpoint-store`, will hit the same exit.
- Fix: the unknown-id message should say the log may not hold the Goal's add. Have `add goal` also warn when the body has no heading at all, or state explicitly that this is intended.

## Low

**7. A heading in prose silently starts the block early.**
- Scenario: (d).
- Why: a prose paragraph beginning `Never: …` makes every later prose line invalid. That refuses safely, but a Goal can never be edited.
- Fix: the `add goal` warning should print the offending line number and the first non-heading line.

**8. Heading matching is under-specified.**
- Scenario: (d).
- Why: the rules do not say whether matching is case-sensitive, whether leading whitespace is allowed, or whether a non-ASCII colon counts. Unicode in the block is not a problem: `textLimit` counts runes (`validate.go:905`), so 2000 runes can be up to 8000 bytes, and printed text is only the body.
- Fix: state "byte-exact, at column 0, ASCII `:`". Then `never:`, ` Never:` and a fullwidth colon are all non-headings, which rule 4 already refuses safely. Any control or escape characters in the body should be stripped or refused, since they would reach a terminal.

**9. A trailing newline can double blank lines in output.**
- Scenario: (d).
- Why: a trailing `\n` produces a final empty element after splitting, which is empty and valid. Printing "line for line" can then add a stray blank line.
- Fix: trim one trailing newline before printing, and end output with exactly one.

**10. Two Goals in one store, and one Goal naming the other's id.**
- Scenario: (a).
- Why: this is defined and safe. The block is the tail of the body and only headed lines survive, so prose naming `G-other` before the block is not printed. A constraint line such as `Never: touch G-B` is printed as written and is only text. The display id is the last 6 characters of the ULID (`event.go:496`), so a collision gives "ambiguous node" and exit 2 (`resolve.go:36`).
- Fix: print the resolved display id. Check `Kind == Goal`, because `resolveRef` accepts any node.

**11. A Goal that is withdrawn or complete.**
- Scenario: (f).
- Why: `goalState` returns complete or in_progress only (`replay.go:511-531`). There is no withdrawn Goal state, since only Experiments and DQs are withdrawn. The rules say nothing about prompting for a complete Goal. OWED may be empty, so the first stop clause fires immediately and the run does nothing. That is harmless.
- Fix: either print a `warn:` on stderr, exit 0, for a complete Goal, or say it is intended.

**12. Who added the Goal is not checked.**
- Scenario: (b).
- Why: the owner is not authenticated (ADR 0020 "Unchanged"; principles §7 rule 7). A Goal added from any actor's events file yields a valid prompt. This is consistent with the rest of hyp.
- Fix: no code change. Optionally print "Goal added by X" so the human sees it.

**13. Plumbing the new command in.**
- `help_test.go:19` runs `-h` and `--help` for every name in `Commands()`. It requires exit 0, stdout and an empty stderr with no store, so `hyp prompt -h` must be handled before the store lookup.
- The usage text and `skillcheck` need `prompt` registered in `commands` (`cli.go:92`). Mentioning `hyp prompt` in `hyp-goal-run` is then valid.
- No store is exit 3: `ErrNoStore` is a plain error, and `Run` maps it through the default branch (`cli.go:176`).
- The writers check arguments before they look for the store, so exit 2 beats exit 3 (`write.go:88-112`). Keep that order.

## Could not verify
- Whether `/goal` has a prompt length limit.
- Whether `parse` interleaves flags and positionals. I assumed it does.
- Whether `Node.Actor` for a Result is the recording actor. It looks that way at `rules.go:81`, but I did not trace it.
- Whether the Goal body's `Actor` is used anywhere for owner checks. I saw no use.</result>
<usage><subagent_tokens>60420</subagent_tokens><tool_uses>10</tool_uses><duration_ms>100739</duration_ms></usage>
</task-notification>

## Grill

The second reader has finished. Its findings are claims I haven't checked, so I'm verifying the ones my rules depend on against the source.

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 28,36p internal/store/store.go; sed -n 38,56p internal/store/store.go; sed -n 76,84p internal/replay/rules.go; grep -n \"ResRecorded\\|Retracted\" internal/replay/state.go | head -8; grep -n \"EvalSymlinks\" -r internal | head -3; sed -n 100,104p internal/cli/write.go"
}
```

Result:

```text
// ErrNoStore means no .tree/events directory was found above the start dir.
var ErrNoStore = errors.New("no .tree store found; run 'hyp init'")

var actorRe = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,39}$`)

// LineError is a line of an events file that could not be parsed.
type LineError struct {
	File string
	Err  error
}

// Dir returns the .tree directory under root.
func Dir(root string) string { return filepath.Join(root, ".tree") }

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
	event.TypeResultRecord: {
		validate: (*Facts).validateResult,
		apply: func(st *State, ev event.Event) {
			var p event.ResultRecord
			_ = event.DecodePayload(ev, &p)
			st.addNode(&Node{ID: ev.ID, Kind: event.KindResult, Parent: p.Parent, Actor: ev.Actor,
				AddedBy: ev.ID, Verdicts: p.Verdicts, VerdictOrder: event.VerdictOrder(ev.Payload), Observation: p.Observation, Evidence: p.Evidence,
				EvidenceHashes: p.EvidenceHashes, Limits: p.Limits, Closes: p.Closes, Readings: p.Readings, Applies: p.Applies,
				Subjects: runSubjects(p)})
311:	Retracted map[string]bool
371:			Retracted:    map[string]bool{},
416:	ResRecorded  ResStatus = "recorded"
417:	ResRetracted ResStatus = "retracted"
internal/recipe/recipe_test.go:692:		r, err := filepath.EvalSymlinks(p)
internal/recipe/recipe.go:365:	resolved, err := filepath.EvalSymlinks(dir)
internal/recipe/recipe.go:674:	p, err := filepath.EvalSymlinks(filepath.Join(root, cmd))
		return nil, usageErr("missing actor: pass --actor or set HYP_ACTOR")
	}
	if !store.ValidActor(*w.actor) {
		return nil, usageErr("invalid actor %q (want ^[a-z0-9][a-z0-9-]{0,39}$)", *w.actor)
	}
```

## Grill

The second reader found real gaps. I checked these against the source: `Node.Actor` for a Result is the recording actor (`internal/replay/rules.go:81`), actors persist across runs (`internal/store/store.go:31`), `FindRoot` doesn't resolve symlinks, and Results have `recorded` and `retracted` states. No answers from you have arrived since round 2, so nothing is written.

**Fixes I'm adding to the rules** (they only tighten the draft):
- **Shared validator.** `hyp add goal` and `hyp prompt` use one block validator, so the warning and the refusal can't disagree.
- **CRLF.** `\r\n` is read as `\n`, and output uses `\n`. A lone `\r` still isn't empty. Without this, a CRLF body fails on every blank line, and a Goal is never edited.
- **Headings.** A heading is byte-exact at column 0 with an ASCII colon. The `add goal` warning names the first offending line.
- **Flags.** `--pr` is digits only. `--branch` has no whitespace or control characters and no leading `-`. `--results` is an integer of at least 1. A bad `--actor` fails the actor pattern. All of these exit 2, and they exit before any store lookup. `hyp prompt -h` is handled before the store lookup, because `help_test` requires exit 0 with no store.
- **Output.** It ends with exactly one newline. The id must resolve to a Goal, and the printed id is the resolved display id.
- **A complete Goal.** It prints the prompt, exits 0, and writes a `warn:` line on stderr.

**Two defects that need your decision**

1. **The Results cap has no baseline.** An actor name that was used before already has Results in the log, so a reused actor satisfies "N Results" at once and the run stops immediately. The count also ignored the Goal and retractions.
   - Count the actor's `recorded` Results under Experiments of this Goal. A retracted Result doesn't count, and a withdrawn Experiment's recorded Results do.
   - If actor A already wrote any event in the store, `hyp prompt` exits 1 and says to choose a new actor. The actor is the run's identity, so the count then needs no stored baseline. Nothing is stored.
   - **Question:** do you accept the refusal, and one new line (`[N]`, below)?

2. **A bare heading passes the block rule.** `May change:`, `Read only:` and `Never:` with nothing after each colon is a valid block. It would give an unattended run no constraints at all.
   - I recommend that every block line needs non-whitespace text after the colon. The warning on add and the refusal in `hyp prompt` both name this rule.
   - **Question:** do you accept this tightening? It changes the block definition the three skills state.

**Two smaller questions for you**
- **The store path in the prompt.** It comes from the shell, so a symlinked repo prints under whichever name the shell had. I recommend printing it as found, quoted, and adding a first step to the prompt: "run `hyp status <G-id>` there and stop if the Goal is missing". That covers a prompt generated where the Goal's add isn't merged yet. Do you accept the extra step?
- **Round 2 is still open.** I still need your answers on run 6's text, the exact `--branch` and `--pr` clauses, and whether a body with no heading at all warns on add. The second reader also recommends warning there.

**Lines as they now stand** (9 of 10):
1. `[P]` For a valid block, `hyp prompt` prints the pinned text: the store path, actor, Goal id, skill, clauses (a), (b) and (c) with N=5, the first step above, and the block line for line. It carries no owner answer or tree state.
2. `[P]` `--results N` changes only N. `--branch` and `--pr` each add exactly their one clause after the block and nothing else.
3. `[P]` The output depends only on the merged log and the flags, so the same bytes come from any actor-file order. The command appends no event and changes no file. The Result count is as above.
4. `[N]` A Goal with no valid block exits 1 and names the missing headings or the first bad line. This covers a bare heading and CRLF handling, and nothing is printed on stdout.
5. `[P]` `hyp add goal` with a heading but an invalid block appends the Goal, exits 0 and prints a `warn:` naming the rule and line. A valid block prints none.
6. `[N]` Bad arguments exit 2. This covers an invalid actor, a bad `--pr`, `--branch` or `--results`, and an unknown or non-Goal id. No repository holding the store exits 3.
7. `[N]` An actor that already has an event in the store exits 1 and says to choose a new one.
8. `[P]` The `hyp-owner-grill` Hand off eval: each of 3 Sonnet trials ends with `hyp prompt`'s output byte for byte, and adds no sentence.
9. `[N]` On a Goal that `hyp prompt` refuses, no trial hand-writes a prompt. Each reports the refusal.

The routing table from round 2 stands. The Generator row is unchanged. The DesignQuestion for a Goal with no block still records "refuse, exit 1", and the one for `hyp add goal` still records "warn, exit 0". The constraints block gets two extra lines:
```
Never: store a Result count, an actor baseline or a parsed constraint field
May change: internal/cli and internal/replay for the command and the shared block validator
```

**Still missing before any write:** your answers to the round 2 questions (run 6's text, the flag clause wording, the no-heading warning) and to questions 1 and 2 above and the path step.

## Owner, message 4

1. Accept. `hyp prompt` refuses a reused actor and says to choose a new one. Add the `[N]` line for it.

2. Accept. Every block line needs non-whitespace text after its colon. Update the block definition in the three skills to match.

3. Accept the extra first step, `hyp status <G-id>`, which stops the run if the Goal is missing. Print the store path as found, quoted.

4. I have no run 6 text to give you, so work from the run-5 text you have. For the `--branch` and `--pr` clauses, use your narrowing-only wording. A body with no heading at all warns on add and still appends the Goal at exit 0.

5. Your 9 lines and the two extra constraint lines are fine. On line 8 and line 9, the trial design is yours to settle from the ak-skill-creator flow.

## Grill

Everything is settled except the exact prompt text. I'm checking the Goal body limit before drafting it, since the body must carry that text.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn \"MaxGoalBody\\|MaxLine\" internal/event/*.go | grep -v _test | head; grep -n \"Never\\|actor\" internal/cli/cli.go | head -5; mkdir -p $TMPDIR && echo $TMPDIR"
}
```

Result:

```text
internal/event/event.go:117:	MaxGoalBody     = 2000
internal/event/event.go:128:	MaxLine         = 600 // a line on a Goal's node.add (ADR 0020); a line an old checkpoint wrote holds at most MaxLimitItem
55:Every write command needs --why and an actor (--actor or $HYP_ACTOR).
/tmp/claude-501
```

## Grill

The body is capped at 2000 runes and each line at 600, so I'm drafting both and counting them. Then I'll dry-run the add in a throwaway store under a scratch actor, not the owner's.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/g113 && cd $TMPDIR/g113 && rm -rf .tree && cat > body.txt <<'EOF'\nBuilds #113: `hyp prompt` prints a run's /goal prompt from the log; the constraints block below is its read side. Owner decision 2026-11-11: constraints live in the body under May change:, Read only: and Never:, and a typed field waits until a real run misreads or loses one. Every block line needs text after its colon. One checker serves `hyp add goal` (warns, still appends) and `hyp prompt` (refuses, exit 1). hyp prints the run cap and never enforces it; ADR 0022 supersedes ADR 0017's cap line in part. Lines 8 and 9 are an ak-skill-creator eval of the hyp-owner-grill Hand off: 2 cases, 3 Sonnet trials each, candidate and base skill, graded by eval_skill.py. The description edits have no line and no routing traces run: a known gap. A commit Goal's other needs from hyp are #114.\nMay change: this repository (hyp): commits on one local branch\nMay change: one pull request, opened and never merged\nMay change: internal/cli and internal/replay for the command and the shared block checker\nMay change: the descriptions of hyp-goal-run and hyp-prototype, which name one Goal kind\nMay change: skills/hyp-owner-grill Hand off, and the block definition in each skill that states it\nMay change: README.md, docs/glossary.md, docs/replay-spec.md, docs/core-and-self-review.md, docs/how-hyp-improves-itself.md\nMay change: a new docs/adr/0022, and a new recipes/hyp-owner-grill-eval@1\nRead only: investigations/checkpoint-store and the other published trees\nRead only: the ak-skill-creator skill, named on the eval command's External: line and never copied\nNever: replace the installed ~/.local/bin/hyp\nNever: edit a stored Goal or event\nNever: add a typed constraints field to `hyp add goal`\nNever: store a Result count, an actor baseline or a parsed constraint field\nEOF\npython3 - <<'EOF'\nb=open('body.txt').read().rstrip('\\n'); print(len(b))\nEOF\ngrep -rn \"May change:\" /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/skills -l"
}
```

Result:

```text
1765
/Users/datguy/.hyp-sandbox/g146-base-a-1/wt/skills/hyp-owner-grill/SKILL.md
/Users/datguy/.hyp-sandbox/g146-base-a-1/wt/skills/hyp-goal-run/SKILL.md
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g113 && cat > lines.py <<'EOF'\nL=[\n\"[P] For a Goal whose body ends with a valid block, `hyp prompt G --actor A` exits 0 and prints exactly: `/goal Run the hyp tree in \\\"<store path as found>\\\" with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is G.` / `First run `hyp status G` there; if it shows no such Goal, stop and say so.` / `Done when the latest `hyp status G` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under A, or you cannot take the next step honestly and say why.` / `At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.` / the block, line for line. No owner answer or tree state.\",\n\"[P] `--results N` changes only the 5. `--branch B` adds exactly `Commit only to branch B.` and `--pr P` adds exactly `Do not merge PR #P.`, each as one line after the block, and nothing else changes.\",\n\"[P] The output depends only on the merged log and the flags: the same bytes for any order of the actors' events files. `hyp prompt` appends no event and changes no file. The Results counted are the actor's recorded Results under this Goal; a retracted one does not count.\",\n\"[N] A Goal with no valid block (no heading, a missing heading, a bare heading with no text after its colon, or a non-empty line in the block that starts with no heading) makes `hyp prompt` exit 1, name the missing heading or the first bad line, and print nothing on stdout. CRLF line ends read as LF.\",\n\"[P] `hyp add goal` with a body that has no heading, or a heading but no valid block, appends the Goal, exits 0 and prints a `warn:` line that names the rule and the first bad line. A body with a valid block prints no warning. One checker serves both commands.\",\n\"[N] Bad arguments exit 2: an invalid actor, a `--pr` that is not digits, a `--branch` with whitespace, a control character or a leading `-`, a `--results` below 1, and an id that is not a Goal. No repository holding the store exits 3. `hyp prompt -h` exits 0 with no store.\",\n\"[N] `hyp prompt` for an actor that already wrote any event in the store exits 1, names the rule and tells the owner to choose a new actor.\",\n\"[P] At the candidate, each of 3 fresh Sonnet trials of the hyp-owner-grill Hand off, given a finished grill whose Goal has a valid block, ends with the output of `hyp prompt` for that Goal byte for byte and adds no sentence.\",\n\"[N] At the candidate, no one of 3 fresh Sonnet trials of the Hand off, given a Goal that `hyp prompt` refuses, writes a prompt by hand. Each tells the owner of the refusal.\",\n]\nfor i,l in enumerate(L,1): print(i,len(l))\nimport json; json.dump(L,open('lines.json','w'))\nEOF\npython3 lines.py"
}
```

Result:

```text
1 684
2 199
3 271
4 300
5 259
6 273
7 138
8 224
9 172
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g113 && python3 - <<'EOF'\nimport json\nL=json.load(open('lines.json'))\nL[0]=\"[P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run the hyp tree in \\\"<store path as found>\\\" with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is G.` / `First run `hyp status G` there; if no such Goal shows, stop.` / `Done when the latest `hyp status G` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under A, or you cannot take the next step honestly and say why.` / `At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.` / the block, line for line; no owner answer.\"\nprint(len(L[0])); json.dump(L,open('lines.json','w'))\nEOF\nexport HYP_ACTOR=scratch-dryrun; hyp init >/dev/null && args=(add goal --title \"dry run\" --body \"$(cat body.txt)\" --why \"dry run in scratch\"); python3 - <<'EOF' > cmd.sh\nimport json,shlex\nL=json.load(open('lines.json'))\nprint(' '.join('--line '+shlex.quote(l) for l in L))\nEOF\neval hyp add goal --title dry --body \"\\\"\\$(cat body.txt)\\\"\" $(cat cmd.sh | sed 's/^/ /' | head -0) --why dry 2>&1 | head; "
}
```

Result:

```text
620
head: illegal line count -- 0
G-9CPFM3 node.add
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g113 && rm -rf .tree && python3 - <<'EOF'\nimport json,subprocess,os\nL=json.load(open('lines.json'))\nL[0]=L[0].replace('<store path as found>','<store path>').replace('; no owner answer.','.')\nprint([len(l) for l in L]); json.dump(L,open('lines.json','w'))\nenv=dict(os.environ,HYP_ACTOR='scratch-dryrun')\nsubprocess.run(['hyp','init'],env=env,capture_output=True)\ncmd=['hyp','add','goal','--title','dry','--body',open('body.txt').read().rstrip('\\n'),'--why','dry']\nfor l in L: cmd+=['--line',l]\nr=subprocess.run(cmd,env=env,capture_output=True,text=True); print(r.returncode,r.stdout,r.stderr)\nEOF\nrm -rf $TMPDIR/g113/.tree"
}
```

Result:

```text
[594, 199, 271, 300, 259, 273, 138, 224, 172]
0 G-KD8KC2 node.add
  line G-KD8KC2#1: [P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run the hyp tree in "<store path>" with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is G.` / `First run `hyp status G` there; if no such Goal shows, stop.` / `Done when the latest `hyp status G` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under A, or you cannot take the next step honestly and say why.` / `At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.` / the block, line for line.
  line G-KD8KC2#2: [P] `--results N` changes only the 5. `--branch B` adds exactly `Commit only to branch B.` and `--pr P` adds exactly `Do not merge PR #P.`, each as one line after the block, and nothing else changes.
  line G-KD8KC2#3: [P] The output depends only on the merged log and the flags: the same bytes for any order of the actors' events files. `hyp prompt` appends no event and changes no file. The Results counted are the actor's recorded Results under this Goal; a retracted one does not count.
  line G-KD8KC2#4: [N] A Goal with no valid block (no heading, a missing heading, a bare heading with no text after its colon, or a non-empty line in the block that starts with no heading) makes `hyp prompt` exit 1, name the missing heading or the first bad line, and print nothing on stdout. CRLF line ends read as LF.
  line G-KD8KC2#5: [P] `hyp add goal` with a body that has no heading, or a heading but no valid block, appends the Goal, exits 0 and prints a `warn:` line that names the rule and the first bad line. A body with a valid block prints no warning. One checker serves both commands.
  line G-KD8KC2#6: [N] Bad arguments exit 2: an invalid actor, a `--pr` that is not digits, a `--branch` with whitespace, a control character or a leading `-`, a `--results` below 1, and an id that is not a Goal
[... 742 more characters]
```

## Grill

The dry run in a scratch store (since deleted) accepted the body at 1765 of 2000 runes and all nine lines. Line 1 is 594 of 600 runes. Nothing is written to the real tree yet.

**Settled for lines 8 and 9** (the trial design, from the ak-skill-creator flow)
- 2 cases: a finished grill whose Goal has a valid block, and one whose Goal `hyp prompt` refuses.
- 3 Sonnet trials per case, each a fresh subagent with the candidate skill and no way to find the base copy. The base skill gets the same cases and trial count, which makes 12 runs.
- `eval_skill.py` grades the saved outputs with exact-text assertions on the final prompt file. There are no rubric checks.
- The model id is recorded from what the runtime reports.
- Line 8 or 9 holds only when all 3 candidate trials pass. Red on base is each Experiment's verdict rule, not part of a line.
- A correction to what I said earlier: the block definition appears in two skills, `hyp-goal-run` and `hyp-owner-grill`, not three. I'll update those two.

**Routing, final** (confirm each row or move it)

| Question | Goes to |
|---|---|
| Generator: command or skill-written | DesignQuestion Q1: "command reads the log" (chosen) against "skill writes it by hand". Session evidence, no test. |
| A Goal with no block | DesignQuestion Q2: "refuse, exit 1" (chosen), "print with no constraint lines", "print with 'no constraints recorded'". Scenario: G-D1CH1R in a store. |
| `hyp add goal` on a body with no valid block | DesignQuestion Q3: "warn at exit 0, still append" (chosen) against "refuse at exit 1". Scenario: the malformed Goal is added, then prompted. The warning also covers a body with no heading. |
| Stop clauses (a)(b)(c), default 5, first step, store path | line 1 |
| Flags only narrow | line 2 |
| Derived from the log, read-only, Result count | line 3 |
| Refusals, shared checker, warning | lines 4 and 5 |
| Argument and store errors | line 6 |
| Reused actor refused | line 7 |
| Hand off eval | lines 8 and 9, plus the verdict rules |
| Description edits | May change only, no line |
| Typed field, ADR 0022, routing-trace gap, eval budget | Goal body |
| Commit Goal's other needs from hyp | #114, a separate Goal |

**The owner write, for your approval.** This is the only write under `owner-standin`. It goes in `investigations/goal-prompt/`, after the runner's `hyp init`.
```
HYP_ACTOR=owner-standin hyp add goal \
  --title "hyp prints a run's /goal prompt from the log, and a Goal body records the run's constraints" \
  --body "$(cat $TMPDIR/g113/body.txt)" \
  --line "<line 1>" … --line "<line 9>" \
  --why "Owner flow, #113: grilled 2026-11-12; constraints live in the body, a typed field waits."
```
- The nine lines are exactly the strings shown in the dry run above. They are saved in `$TMPDIR/g113/lines.json`, and the body is in `$TMPDIR/g113/body.txt`. I'll pass them as argv from those files, so apostrophes and backticks need no shell quoting.
- There are no `--cite`s, since this Goal carries no settled key.

**Then the runner writes**, as `claude-grill-113`, after you approve:
1. `hyp init` in `investigations/goal-prompt/`.
2. `hyp add dq --parent <G-id>` for each of Q1 to Q3, then one `hyp add hyp` per rival.
3. Under each chosen rival, `hyp add exp --recipe none --rubric reading --target <each other rival>`. Its body gives the scenarios as its rule. A scenarios file goes under `.tree/evidence/<E-id>/`, and `hyp record` gives a verdict per rival. The chosen rival is `supports` and the others `contradicts`.
4. The select that OWED owes for each of Q1 to Q3.
5. `hyp link <G-id> external_ref gh:113`, and the same for #114.
6. `hyp status <G-id>` and `hyp check`, which must print `check: 0 errors`.
7. `hyp checkpoint --goal <G-id> --summary "…" --why "end of grill"`.

**Hand-off.** `hyp prompt` doesn't exist on base, so it can't print this Goal's prompt. I'll hand-write this Goal's `/goal` prompt once, from the on-disk Hand off section, using the same text that line 1 pins plus the branch and PR clauses. It is the last hand-written one. The prompt will name the Goal's id, so the Goal's add must be in the log the run starts from.

Do you confirm the routing table and approve the `hyp add goal` command above?

## Owner, message 5

Confirmed: the routing table matches my positions, and I approve the `hyp add goal` command under owner-standin as shown. Don't push, post on GitHub or open a PR in this session.

## Grill

Both approvals received. I'll create the tree directory and run `hyp init` as the runner, then the owner's `hyp add goal` with exactly the saved strings.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p investigations/goal-prompt && cd investigations/goal-prompt && export HYP_ACTOR=claude-grill-113 && hyp init && ls -a && python3 - <<'EOF'\nimport json,subprocess,os,glob\nd=glob.glob('/tmp/claude-501*/g113')+[os.path.join(os.environ.get('TMPDIR','/tmp'),'g113')]\nd=[x for x in d if os.path.exists(x+'/lines.json')][0]\nL=json.load(open(d+'/lines.json')); body=open(d+'/body.txt').read().rstrip('\\n')\ncmd=['hyp','add','goal','--title',\"hyp prints a run's /goal prompt from the log, and a Goal body records the run's constraints\",'--body',body,'--why','Owner flow, #113: grilled 2026-11-12; constraints live in the body, a typed field waits.']\nfor l in L: cmd+=['--line',l]\nr=subprocess.run(cmd,env=dict(os.environ,HYP_ACTOR='owner-standin'),capture_output=True,text=True)\nprint(r.returncode,r.stdout[:300],r.stderr)\nEOF"
}
```

Result:

```text
initialised .tree in /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/goal-prompt
.
..
.tree
0 G-N77C5E node.add
  line G-N77C5E#1: [P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run the hyp tree in "<store path>" with the hyp-goal-run skill, as HYP_ACTOR=A. Its Goal is G.` / `First run `hyp status G` there; if no such Goal shows, stop.` / `Done when the latest `hy
```

## Grill

The Goal is G-N77C5E. Next the runner writes: the three DesignQuestions with their rivals.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-N77C5E\nq(){ hyp add dq --parent $G --alias \"$1\" --title \"$2\" --body \"$3\" --why \"Settled by the owner in the grill session of 2026-11-12; recorded so the tree says it was argued, not tested.\" | head -1; }\nq gen \"Who produces a run's /goal prompt?\" \"Scenario: a Goal with a malformed block, and two calls on one merged log. A command can exit 1 naming the rule and give the same bytes twice; a skill-written prompt cannot refuse with an exit code.\"\nq noblock \"What does hyp prompt do for a Goal whose body has no valid constraints block?\" \"Scenario: G-D1CH1R, a stored Goal with no block, in a store. A prompt with no constraint lines reads as permission to a run.\"\nq addwarn \"What does hyp add goal do for a body with no valid constraints block?\" \"Scenario: add the malformed Goal, then ask for its prompt. A Goal is never edited, so a refused add leaves no unusable Goal, and an allowed add leaves one that hyp prompt refuses.\"\nhyp status $G | head -20"
}
```

Result:

```text
Q-A5C4SG node.add
--body names G-D1CH1R, which is no node here; cite an id this store holds, or write <name>:<id> for another store's
Q-TX16PY node.add
GOAL G-N77C5E hyp prints a run's /goal prompt from the log, and a Goal bo… — in_progress (answered 0, open 2, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-N77C5E#1 no candidate — [P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run the hyp tree in "<store path>" with the hyp-goal-run s…
  G-N77C5E#2 no candidate — [P] `--results N` changes only the 5. `--branch B` adds exactly `Commit only to branch B.` and `--pr P` adds exactly `Do not merge PR…
  G-N77C5E#3 no candidate — [P] The output depends only on the merged log and the flags: the same bytes for any order of the actors' events files. `hyp prompt` a…
  G-N77C5E#4 no candidate — [N] A Goal with no valid block (no heading, a missing heading, a bare heading with no text after its colon, or a non-empty line in th…
  G-N77C5E#5 no candidate — [P] `hyp add goal` with a body that has no heading, or a heading but no valid block, appends the Goal, exits 0 and prints a `warn:` l…
  G-N77C5E#6 no candidate — [N] Bad arguments exit 2: an invalid actor, a `--pr` that is not digits, a `--branch` with whitespace, a control character or a leadi…
  G-N77C5E#7 no candidate — [N] `hyp prompt` for an actor that already wrote any event in the store exits 1, names the rule and tells the owner to choose a new a…
  G-N77C5E#8 no candidate — [P] At the candidate, each of 3 fresh Sonnet trials of the hyp-owner-grill Hand off, given a finished grill whose Goal has a valid bl…
  G-N77C5E#9 no candidate — [N] At the candidate, no one of 3 fresh Sonnet trials of the Hand off, given a Goal that `hyp prompt` refuses, writes a prompt by han…
FRONTIER
  Q-A5C4SG open/no_candidates Who produces a run's /goal prompt?
  Q-TX16PY open/no_candidates What does hyp add goal do for a body with no valid constrai…
STALE
OWED
  runner: Q-A5C4SG add rival Hypotheses
  r
[... 58 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113 && G=G-N77C5E\nhyp add dq --parent $G --alias noblock --title \"What does hyp prompt do for a Goal whose body has no valid constraints block?\" --body \"Scenario: checkpoint-store:G-D1CH1R, a stored Goal with no block, in a store. A prompt with no constraint lines reads as permission to a run.\" --why \"Settled by the owner in the grill session of 2026-11-12; recorded so the tree says it was argued, not tested.\" | head -1\nhyp status $G | sed -n '/FRONTIER/,/STALE/p'"
}
```

Result:

```text
Q-VW4S2C node.add
FRONTIER
  Q-A5C4SG open/no_candidates Who produces a run's /goal prompt?
  Q-TX16PY open/no_candidates What does hyp add goal do for a body with no valid constrai…
  Q-VW4S2C open/no_candidates What does hyp prompt do for a Goal whose body has no valid …
STALE
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nW=\"Rival recorded in the grill session of 2026-11-12.\"\nh(){ hyp add hyp --parent \"$1\" --alias \"$2\" --title \"$3\" --body \"$4\" --why \"$W\" | head -1; }\nh Q-A5C4SG cmd \"A hyp command prints the prompt from the log\" \"hyp prompt reads the merged log and the flags, refuses a Goal with no valid block with exit 1, and gives the same bytes on every call.\"\nh Q-A5C4SG skill \"A skill writes the prompt by hand\" \"The grill skill writes the prompt from the tree in prose; no exit code can refuse a malformed block and two writers give two texts.\"\nh Q-VW4S2C refuse \"hyp prompt refuses a Goal with no valid block, exit 1, headings named\" \"Nothing goes to stdout; the owner adds a new Goal with a valid block.\"\nh Q-VW4S2C silent \"hyp prompt prints the prompt with no constraint lines\" \"A prompt that lists no limits reads as permission to an unattended run.\"\nh Q-VW4S2C none \"hyp prompt prints the prompt with a line 'no constraints recorded'\" \"The run sees that nothing was recorded, but still starts.\"\nh Q-TX16PY warn \"hyp add goal warns at exit 0 and still appends the Goal\" \"The warning names the rule and the first bad line; adding stays allowed.\"\nh Q-TX16PY refuseadd \"hyp add goal refuses a body with no valid block, exit 1\" \"Nothing is appended, so no Goal exists that hyp prompt would refuse.\"\nhyp status G-N77C5E | sed -n '/FRONTIER/,/STALE/p'; hyp add exp -h 2>&1 | head -20; hyp record -h 2>&1 | head -5"
}
```

Result:

```text
H-5PMS5D node.add
H-1C8VGB node.add
H-H1SYJK node.add
H-YT16YZ node.add
H-JTQK12 node.add
H-PXCKGT node.add
H-HT1041 node.add
FRONTIER
  Q-A5C4SG open/untested Who produces a run's /goal prompt?
    H-5PMS5D (cmd) proposed A hyp command prints the prompt from the log
    H-1C8VGB (skill) proposed A skill writes the prompt by hand
  Q-TX16PY open/untested What does hyp add goal do for a body with no valid constrai…
    H-PXCKGT (warn) proposed hyp add goal warns at exit 0 and still appends the Goal
    H-HT1041 (refuseadd) proposed hyp add goal refuses a body with no valid block, exit 1
  Q-VW4S2C open/untested What does hyp prompt do for a Goal whose body has no valid …
    H-H1SYJK (refuse) proposed hyp prompt refuses a Goal with no valid block, exit 1, head…
    H-YT16YZ (silent) proposed hyp prompt prints the prompt with no constraint lines
    H-JTQK12 (none) proposed hyp prompt prints the prompt with a line 'no constraints re…
STALE
usage of hyp add exp:
  -actor string
    	actor name (default $HYP_ACTOR)
  -aims value
    	R-id#n of a limit, or R-id#rn of an answered reading, this Experiment means to settle (repeatable)
  -alias string
    	optional unique alias
  -body string
    	body (at most 600 runes)
  -decided-by string
    	reading: the older spelling of --rubric reading
  -parent string
    	parent node (required)
  -recipe string
    	NAME@VERSION or none (required)
  -records value
    	file or directory under .tree/evidence/ that an earlier run kept and this Experiment reads; hyp pins each file's sha256 (repeatable)
  -rubric string
    	NAME@VERSION of a rubric bundle, or reading: how measures become verdicts (required under a Hypothesis, refused under a DesignQuestion)
  -target value
usage of hyp record:
  -actor string
    	actor name (default $HYP_ACTOR)
  -applies value
    	R-id#rn of an answered reading whose answer these verdicts apply; the Experiment must aim at it (repeatable)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd $TMPDIR && rm -rf scr && cp -r /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/goal-prompt scr && cd scr && export HYP_ACTOR=claude-grill-113\nhyp add exp --parent H-5PMS5D --alias e1 --title \"Scenarios: command or skill\" --recipe none --rubric reading --target H-1C8VGB --body \"Rule: scenarios file decides.\" --why t | head -3\nE=$(hyp status | grep -o 'E-[A-Z0-9]*' | head -1); echo $E\nmkdir -p .tree/evidence/$E && echo \"scenarios\" > .tree/evidence/$E/scenarios.md\nhyp record --exp $E --verdict H-5PMS5D=supports --verdict H-1C8VGB=contradicts --observation \"Scenarios in scenarios.md.\" --evidence .tree/evidence/$E/scenarios.md --why t 2>&1 | head -5\nhyp status | sed -n '/FRONTIER/,/AWAITING/p' | cut -c1-200"
}
```

Result:

```text
E-9EN8M2 node.add
E-9EN8M2
R-EGEPJG result.record
FRONTIER
  Q-A5C4SG open/supported_unselected Who produces a run's /goal prompt?
    candidates (1): H-5PMS5D (cmd)
    H-5PMS5D (cmd) supported A hyp command prints the prompt from the log
    H-1C8VGB (skill) contradicted A skill writes the prompt by hand
  Q-TX16PY open/untested What does hyp add goal do for a body with no valid constrai…
    H-PXCKGT (warn) proposed hyp add goal warns at exit 0 and still appends the Goal
    H-HT1041 (refuseadd) proposed hyp add goal refuses a body with no valid block, exit 1
  Q-VW4S2C open/untested What does hyp prompt do for a Goal whose body has no valid …
    H-H1SYJK (refuse) proposed hyp prompt refuses a Goal with no valid block, exit 1, head…
    H-YT16YZ (silent) proposed hyp prompt prints the prompt with no constraint lines
    H-JTQK12 (none) proposed hyp prompt prints the prompt with a line 'no constraints re…
STALE
OWED
  runner: Q-A5C4SG limit G-N77C5E#1 [owed] — [P] For a valid block, `hyp prompt G --actor A` exits 0 and prints: `/goal Run …
  runner: Q-A5C4SG limit G-N77C5E#2 [owed] — [P] `--results N` changes only the 5. `--branch B` adds exactly `Commit only to…
  runner: Q-A5C4SG limit G-N77C5E#3 [owed] — [P] The output depends only on the merged log and the flags: the same bytes for…
  runner: Q-A5C4SG limit G-N77C5E#4 [owed] — [N] A Goal with no valid block (no heading, a missing heading, a bare heading w…
  runner: Q-A5C4SG limit G-N77C5E#5 [owed] — [P] `hyp add goal` with a body that has no heading, or a heading but no valid b…
  runner: Q-A5C4SG limit G-N77C5E#6 [owed] — [N] Bad arguments exit 2: an invalid actor, a `--pr` that is not digits, a `--b…
  runner: Q-A5C4SG limit G-N77C5E#7 [owed] — [N] `hyp prompt` for an actor that already wrote any event in the store exits 1…
  runner: Q-A5C4SG limit G-N77C5E#8 [owed] — [P] At the candidate, each of 3 fresh Sonnet trials of the hyp-owner-grill Hand…
  runner: Q-A5C4SG limit G-N77C5E#9 [owed] — [N] At the can
[... 357 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "rm -rf $TMPDIR/scr; cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/goal-prompt && hyp status | sed -n '1,2p;/OWED/,/AWAITING/p' | cut -c1-120; cd ../.. && git status --short"
}
```

Result:

```text
GOAL G-N77C5E hyp prints a run's /goal prompt from the log, and a Goal bo… — in_progress (answered 0, open 3, reopened 0
LINES
OWED
  runner: Q-A5C4SG test H-5PMS5D (cmd)
  runner: Q-A5C4SG test H-1C8VGB (skill)
  runner: Q-TX16PY test H-PXCKGT (warn)
  runner: Q-TX16PY test H-HT1041 (refuseadd)
  runner: Q-VW4S2C test H-H1SYJK (refuse)
  runner: Q-VW4S2C test H-YT16YZ (silent)
  runner: Q-VW4S2C test H-JTQK12 (none)
AWAITING OWNER
?? investigations/goal-prompt/
```

## Grill

I made a routing mistake and stopped before recording the settled DesignQuestions. The tree is half-written.

**What's in the real tree now** (`investigations/goal-prompt/`, uncommitted, nothing pushed or posted)
- Goal G-N77C5E with its nine lines, written under `owner-standin` as you approved.
- Q-A5C4SG (generator), Q-VW4S2C (a Goal with no block) and Q-TX16PY (`hyp add goal` on a bad block), with their seven rival Hypotheses. All three are unrecorded, with no Experiment or select.

**The mistake.** The skill says a design record and its build are two Goals. I skipped that split, so the three argued DesignQuestions sit under the build Goal. I tested what would happen on a scratch copy, since removed. After one of them gets a supported candidate, hyp owes all nine lines on that question:
```
runner: Q-A5C4SG limit G-N77C5E#1 [owed] … through #9 [owed]
```
A "command prints the prompt" claim argued in a session can never be covered by a commit run, so that select would block forever. The same happens to the other two. The skill says a Goal whose questions are all settled by argument gets no lines. The settled questions belong in a design record, and this Goal needs its own question: which hyp commit implements the design, as `Q-PC04ED` did for G-D1CH1R.

**The fix, in order**
1. You approve a second owner write: a design-record Goal in the same store, with no lines. The body's constraints block is written so it passes the new checker, and the Goal is never run:
   ```
   HYP_ACTOR=owner-standin hyp add goal \
     --title "Design record for #113: who prints a run's /goal prompt, and what hyp does on a Goal with no valid constraints block" \
     --body $'Records three questions the owner settled by argument in the grill session of 2026-11-12, with the session as evidence. G-N77C5E builds the design and carries its cases as lines. No run checks these candidates, so this Goal has no lines.\nMay change: this Goal\'s tree events and evidence, to record the session\'s decisions\nRead only: the repository\nNever: run this Goal as an unattended /goal' \
     --why "Owner flow, #113: design record split from its build, per the grill skill."
   ```
2. The runner withdraws my three misplaced DesignQuestions under G-N77C5E with `hyp withdraw`, so none stay in its OWED. The withdraw's `--why` points to the design record. The seven rival Hypotheses and the three questions stay in the log as history, because events are append-only.
3. The runner re-adds the three questions with the same rivals under the design record. For each it records a no-recipe Experiment, a scenarios file under `.tree/evidence/<E-id>/`, verdicts (chosen rival `supports`, others `contradicts`), and the select that OWED owes.
4. The runner adds one build question under G-N77C5E, "Which hyp commit implements #113?", with no Hypotheses. The run adds the rivals. Without that question the Goal would have nothing owed.
5. `hyp link` to #113 and #114, then `hyp status` and `hyp check` (target `check: 0 errors`). Then the end-of-grill `hyp checkpoint` on G-N77C5E, and a hand-written `/goal` prompt for this Goal, since `hyp prompt` isn't built yet.

Do you approve the step 1 command and this order? The only alternative I see is to leave the three questions under the build Goal and `hyp accept` the lines they owe. That leaves the argued decisions inside a Goal that's meant to build them, and I don't recommend it.

## Owner, message 6

Approved: the step 1 command under owner-standin as shown, and the order you gave. I take your recommendation over the `hyp accept` alternative. Don't push, post on GitHub or open a PR in this session.

## Grill

Approved. Running step 1 exactly as shown, then steps 2 to 4.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/goal-prompt && python3 - <<'EOF'\nimport subprocess,os\nbody=\"Records three questions the owner settled by argument in the grill session of 2026-11-12, with the session as evidence. G-N77C5E builds the design and carries its cases as lines. No run checks these candidates, so this Goal has no lines.\\nMay change: this Goal's tree events and evidence, to record the session's decisions\\nRead only: the repository\\nNever: run this Goal as an unattended /goal\"\nr=subprocess.run(['hyp','add','goal','--title',\"Design record for #113: who prints a run's /goal prompt, and what hyp does on a Goal with no valid constraints block\",'--body',body,'--why','Owner flow, #113: design record split from its build, per the grill skill.'],env=dict(os.environ,HYP_ACTOR='owner-standin'),capture_output=True,text=True)\nprint(r.returncode,r.stdout,r.stderr)\nEOF"
}
```

Result:

```text
0 G-HEDGTC node.add
```

## Grill

G-HEDGTC is the design record. Step 2: withdraw the three misplaced questions. Step 3: re-add them under the design record and record each as settled in the session.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nfor q in Q-A5C4SG Q-TX16PY Q-VW4S2C; do hyp withdraw $q --why \"Placed under the build Goal G-N77C5E by mistake: a question settled by argument belongs in the design record G-HEDGTC, since the build Goal's lines are owed on it and no commit run can cover it.\" | head -2; done\nhyp withdraw -h 2>&1 | head -5; hyp status G-N77C5E | sed -n '/FRONTIER/,/AWAITING/p'"
}
```

Result:

```text
Q-A5C4SG dq.withdraw
Q-TX16PY dq.withdraw
Q-VW4S2C dq.withdraw
usage of hyp withdraw:
  -actor string
    	actor name (default $HYP_ACTOR)
  -why string
    	why this event is recorded (required, 1–300 runes)
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cat > $TMPDIR/design.py <<'EOF'\nimport subprocess,os,re\nenv=dict(os.environ,HYP_ACTOR='claude-grill-113')\nD='G-HEDGTC'\nW='Settled by the owner in the grill session of 2026-11-12; the session is the evidence.'\ndef run(*a):\n    r=subprocess.run(['hyp',*a],env=env,capture_output=True,text=True)\n    if r.returncode: raise SystemExit(f\"FAIL {a[:3]}: {r.stdout}{r.stderr}\")\n    return r.stdout\ndef nid(out): return re.match(r'(\\S+) ',out).group(1)\nQS=[\n (\"gen\",\"Who produces a run's /goal prompt?\",\n  \"Scenario: a Goal with a malformed block, and two calls on one merged log. A command can exit 1 naming the rule and give the same bytes twice; a skill-written prompt cannot refuse with an exit code.\",\n  [(\"cmd\",\"A hyp command prints the prompt from the log\",\"hyp prompt reads the merged log and the flags, refuses a Goal with no valid block with exit 1, and gives the same bytes on every call.\"),\n   (\"skill\",\"A skill writes the prompt by hand\",\"The grill skill writes the prompt from the tree in prose; no exit code can refuse a malformed block and two writers give two texts.\")],0),\n (\"noblock\",\"What does hyp prompt do for a Goal whose body has no valid constraints block?\",\n  \"Scenario: checkpoint-store:G-D1CH1R, a stored Goal with no block, in a store. A prompt with no constraint lines reads as permission to an unattended run.\",\n  [(\"refuse\",\"hyp prompt refuses a Goal with no valid block, exit 1, headings named\",\"Nothing goes to stdout; the owner adds a new Goal with a valid block.\"),\n   (\"silent\",\"hyp prompt prints the prompt with no constraint lines\",\"A prompt that lists no limits reads as permission to an unattended run.\"),\n   (\"none\",\"hyp prompt prints the prompt with a line 'no constraints recorded'\",\"The run sees that nothing was recorded, but still starts.\")],0),\n (\"addwarn\",\"What does hyp add goal do for a body with no valid constraints block?\",\n  \"Scenario: add the malformed Goal, then ask for its prompt. A refused add leaves no unusable Goal; an allowed add leaves one that hyp prompt refuses, and a Goal is never edited.\",\n  [(\"warn\",\"hyp add goal warns at exit 0 and still appends the Goal\",\"The warning names the rule and the first bad line; adding stays allowed, as the owner decided.\"),\n   (\"refuseadd\",\"hyp add goal refuses a body with no valid block, exit 1\",\"Nothing is appended, so no Goal exists that hyp prompt would refuse.\")],0),\n]\nRULE={ \"gen\":\"Rule: the rival the scenarios support is supported, the other contradicted. The owner settled it in session: only a command can refuse with an exit code and give one text; a skill-written prompt can do neither.\",\n \"noblock\":\"Rule: the rival the scenarios support is supported, the others contradicted. The owner settled it in session: a prompt with no limits reads as permission, so a Goal with no block is refused with exit 1 and the missing headings named.\",\n \"addwarn\":\"Rule: the rival the scena
[... 2503 more characters]
```

Result:

```text
gen Q-5XQQZ5 ['H-H6QT1E', 'H-GGK1FJ'] E-3RF017 R-75W1ZE result.record
0 Q-5XQQZ5 dq.select 
noblock Q-JV0W60 ['H-6S1EGG', 'H-NH2ZCK', 'H-S4AP5W'] E-5B0901 R-0QGDSH result.record
0 Q-JV0W60 dq.select 
addwarn Q-25NHCT ['H-Z1J0RK', 'H-X3V94C'] E-TDFGA0 R-KJRR5P result.record
0 Q-25NHCT dq.select
```

## Grill

The three questions are recorded and selected. Next: the build question, the links, and `hyp check`.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp add dq --parent G-N77C5E --alias commit --title \"Which hyp commit implements #113?\" --body \"Each commit is a Hypothesis whose subject is the base plus that commit's change; a fix is a new Hypothesis that supersedes it. The run adds the rivals. The design it builds is recorded in G-HEDGTC, selected on Q-5XQQZ5, Q-JV0W60 and Q-25NHCT.\" --why \"The build Goal's own question; without it nothing is owed to the run.\" | head -1\nhyp link G-N77C5E external_ref gh:113 --why \"The issue this Goal serves (#113); #114 holds what a commit Goal needs from hyp.\" | head -1\nhyp link G-N77C5E external_ref gh:114 --why \"Separate Goal's issue: what a commit Goal needs from hyp.\" | head -1\nhyp link G-HEDGTC external_ref gh:113 --why \"The issue whose design questions this Goal records.\" | head -1\nhyp status G-N77C5E | sed -n '1,2p;/FRONTIER/,/AWAITING/p' | cut -c1-140; hyp status G-HEDGTC | sed -n '1,3p;/OWED/,/AWAITING/p'; hyp check"
}
```

Result:

```text
Q-WVQPHZ node.add
G-N77C5E edge.add
G-N77C5E edge.add
G-HEDGTC edge.add
GOAL G-N77C5E hyp prints a run's /goal prompt from the log, and a Goal bo… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdr
LINES
FRONTIER
  Q-WVQPHZ open/no_candidates Which hyp commit implements #113?
STALE
OWED
  runner: Q-WVQPHZ add rival Hypotheses
AWAITING OWNER
GOAL G-HEDGTC Design record for #113: who prints a run's /goal prompt, an… — complete (answered 3, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
check: 0 errors, 0 warnings
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp checkpoint --goal G-N77C5E --summary \"End of grill, 2026-11-12. Build Goal for #113: nine lines, one open question (Q-WVQPHZ) with no rivals yet; the run adds them. Its design is the design record G-HEDGTC (Q-5XQQZ5, Q-JV0W60, Q-25NHCT, all selected). Q-A5C4SG, Q-TX16PY and Q-VW4S2C were first put here by mistake and are withdrawn. Waiting for the owner: nothing.\" --why \"end of grill\" | head -1\nhyp checkpoint --goal G-HEDGTC --summary \"End of grill, 2026-11-12. Three questions settled by the owner in session, each selected with the session as evidence. No lines: G-N77C5E builds the design.\" --why \"end of grill\" | head -1\nhyp check; cd ../.. && git add investigations/goal-prompt && git status --short && git commit -q -m \"docs: record the #113 build Goal and its design record in investigations/goal-prompt\" && git log --oneline -1"
}
```

Result:

```text
G-N77C5E checkpoint
G-HEDGTC checkpoint
check: 0 errors, 0 warnings
A  investigations/goal-prompt/.tree/.gitignore
A  investigations/goal-prompt/.tree/events/claude-grill-113.jsonl
A  investigations/goal-prompt/.tree/events/owner-standin.jsonl
A  investigations/goal-prompt/.tree/evidence/E-3RF017/scenarios.md
A  investigations/goal-prompt/.tree/evidence/E-5B0901/scenarios.md
A  investigations/goal-prompt/.tree/evidence/E-TDFGA0/scenarios.md
879f788 docs: record the #113 build Goal and its design record in investigations/goal-prompt
```

## Grill

The grill is finished, and the tree is committed locally as `879f788` on `work`. Nothing was pushed, posted or opened as a PR. `hyp check` prints `check: 0 errors, 0 warnings`.

**What the tree holds** (`investigations/goal-prompt/`)
- **G-N77C5E** is the build Goal.
  - It has the nine lines and the constraints block.
  - Its one open question is Q-WVQPHZ, "Which hyp commit implements #113?", with no rivals yet. OWED says "runner: Q-WVQPHZ add rival Hypotheses".
  - It is linked to gh:113 and gh:114.
- **G-HEDGTC** is the design record.
  - It has no lines.
  - Its three questions are Q-5XQQZ5 (command, not skill), Q-JV0W60 (refuse a Goal with no block) and Q-25NHCT (`hyp add goal` warns).
  - Each is selected, with the session as evidence in `.tree/evidence/E-3RF017`, `E-5B0901` and `E-TDFGA0`.
- **Withdrawn questions.** My three misplaced questions, Q-A5C4SG, Q-TX16PY and Q-VW4S2C, stay in the log as withdrawn. Their withdraw reasons say why.
- **Both Goals have an end-of-grill checkpoint.** Nothing waits on you under AWAITING OWNER.

**Things to know**
- G-N77C5E's body was written before the design record existed, so it names no G-HEDGTC id. The link is in Q-WVQPHZ's body and in the checkpoint.
- The block-definition tightening reaches two skills, `hyp-goal-run` and `hyp-owner-grill`, not three. The run updates both.
- `hyp prompt` doesn't exist on base, so I hand-wrote this prompt from the on-disk Hand off section, under 4000 characters. Its stop clauses match line 1's pinned text. Generate future prompts with `hyp prompt` once it is built.
- I wrote the actor and branch as `claude-run-113-1` and `feat/hyp-prompt`. Change them if you prefer. The run must use an actor that has no events yet.
- Your "no PR in this session" applies to me only. The block lets the run open one PR and never merge it.

**The `/goal` prompt**
```text
/goal Run the hyp tree in "/Users/datguy/.hyp-sandbox/g146-base-a-1/wt/investigations/goal-prompt" with the hyp-goal-run skill, as HYP_ACTOR=claude-run-113-1. Its Goal is G-N77C5E. Its design record is G-HEDGTC, already selected; do not reopen it.
First run `hyp status G-N77C5E` there; if no such Goal shows, stop. The tree is committed on the branch `work` at 879f788.
Done when the latest `hyp status G-N77C5E` shows `runner owes nothing: stop and report to the owner` under OWED, or 5 Results have been recorded under claude-run-113-1, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
May change: this repository (hyp): commits on one local branch
May change: one pull request, opened and never merged
May change: internal/cli and internal/replay for the command and the shared block checker
May change: the descriptions of hyp-goal-run and hyp-prototype, which name one Goal kind
May change: skills/hyp-owner-grill Hand off, and the block definition in each skill that states it
May change: README.md, docs/glossary.md, docs/replay-spec.md, docs/core-and-self-review.md, docs/how-hyp-improves-itself.md
May change: a new docs/adr/0022, and a new recipes/hyp-owner-grill-eval@1
Read only: investigations/checkpoint-store and the other published trees
Read only: the ak-skill-creator skill, named on the eval command's External: line and never copied
Never: replace the installed ~/.local/bin/hyp
Never: edit a stored Goal or event
Never: add a typed constraints field to `hyp add goal`
Never: store a Result count, an actor baseline or a parsed constraint field
Commit on the branch feat/hyp-prototype-prompt-113 cut from `work`; open one PR from it and never merge it. Each candidate commit is a Hypothesis under Q-WVQPHZ, tied with `hyp subject`; a fix is a new Hypothesis that supersedes it. Red on base and green on the candidate belong in each Experiment's verdict rule, not in a line.
The write scope covers every text the change makes wrong, ADR 0017 and docs/how-hyp-improves-itself.md included.
Run one review round per candidate commit with a fresh subagent given the lines, the code and the keys already settled with the owner's words. Sort each finding: a case of a line is a limit; a question about a line's meaning is a reading; a defect in text you may edit is a fix commit; anything else is an out-of-Goal finding, named in the stop checkpoint and never holding the select.
For lines 8 and 9, spawn the Sonnet trials as fresh subagents (2 cases, 3 trials each, candidate and base skill, 12 runs) and grade the saved outputs with eval_skill.py.
```
Two caveats about that prompt:
- It names the branch `feat/hyp-prototype-prompt-113` in the instruction, while the notes above say `feat/hyp-prompt`. They differ, and I'd settle on one before you run it. Which do you want?
- A prompt that states tree state must come from the merged log. This one names commit `879f788` for that reason.

## Owner, message 7

Use `feat/hyp-prototype-prompt-113`, the name already in the prompt. Change the notes to match.

## Grill

Done: the branch is `feat/hyp-prototype-prompt-113` everywhere. The only place I had written another name was my own note, "I wrote the actor and branch as `claude-run-113-1` and `feat/hyp-prompt`". The prompt and the tree already used your name. Nothing in the repository mentions the other name, so no file needs editing.

The `/goal` prompt above is final as printed. These are the values the notes and the prompt now agree on:
- **Tree:** `investigations/goal-prompt`, committed on `work` at `879f788`.
- **Goal:** G-N77C5E, with the design record G-HEDGTC already selected.
- **Runner actor:** `claude-run-113-1`. It has no events yet, so the actor-reuse rule won't refuse it.
- **Branch:** `feat/hyp-prototype-prompt-113`, cut from `work`.
- **PR:** the run opens one and never merges it.

Nothing has been pushed, posted or opened as a PR. The grill is complete and nothing is waiting on you.
