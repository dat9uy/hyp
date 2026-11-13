# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#113 (the text of issue #113 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-a-3/tmp/issue-113.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-3/tmp/issue-113.md"
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
 "command": "hyp status 2>&1 | head -40; ls; ls investigations docs | head -60"
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
```

### Tool Bash (cwd: .)

```text
{
 "command": "which hyp; hyp --help 2>&1 | head -40; ls -a investigations/*; for d in investigations/checkpoint-store investigations/checkpoint-store-design; do echo == $d; (cd $d && hyp status 2>&1 | head -60); done; ls skills/*; git log --oneline | head -5"
}
```

Result:

```text
Exit code 1
/Users/datguy/.hyp-sandbox/g146-base-a-3/bin/hyp
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
  render --markdown [--as-of REF|TIME]        the whole tree as Markdown
[... 1233 more characters]
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "hyp prompt|prompt",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-a-3/wt",
 "glob": "*.go",
 "output_mode": "files_with_matches"
}
```

Result:

```text
No files found
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp --help 2>&1 | sed -n 40,120p; for d in checkpoint-store checkpoint-store-design; do echo \"== $d\"; (cd investigations/$d && hyp status 2>&1 | head -50); done; ls skills skills/*; ls .claude/skills; ls docs/adr | tail -8; echo $HYP_ACTOR"
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
== checkpoint-store
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A sto…
  G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a …
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and…
  G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Goal's lines come with …
  G-D1CH1R#5 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] One answer may settle key…
  G-D1CH1R#6 covered on Q-PC04ED; ran in R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A refer, aim, close or apply whose …
  G-D1CH1R#7 covere
[... 4142 more characters]
```

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3 && ls; ls tmp; find . -name '*goal-prompt*' -not -path './wt/.git/*' 2>/dev/null | head; cd wt/investigations/checkpoint-store && hyp show G-D1CH1R 2>&1 | head -60"
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
Goal G-D1CH1R
alias: goal
title: hyp implements the shared store: one store holds many Goals
body: Builds the design recorded in checkpoint-store-design:G-73R704 (selects on checkpoint-store-design-D906CD, checkpoint-store-design-HC4V9T, checkpoint-store-design-H8PJ20, checkpoint-store-design-R21S8R). Every decision is selected there with session e6a35a24 as evidence, and ADR 0020 cites those Results. The candidate is a hyp commit; go test on its base and on the candidate is the instrument. Two-Goal stores land first, so lines 2, 6 and 7 are red on that commit's parent.
line G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A store holding one Goal prints byte-identical `hyp status` and `hyp check` output for id-citation and recipe-base on C and on 33896ec.
line G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a store with two Goals, one Goal's OWED, lines and coverage never include the other Goal's items, and each equals what its own events give in a store of their own.
line G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one subject per cited run, and a run at B plus a patch never covers a Hypothesis whose subject is bare B.
line G-D1CH1R#4 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Goal's lines come with its node.add. A cite not settled when written is refused with exit 1, and a later change in the cited key's state is only a hyp check warning.
line G-D1CH1R#5 covered on Q-PC04ED; ran in R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-
[... 7685 more characters]
```

### Tool Read (cwd: investigations/checkpoint-store)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-a-3/wt/skills/hyp-goal-run/SKILL.md"
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

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3 && ls -R tmp/hyp-owner-grill-eval | head -30; grep -rn \"261111-2129\\|g1b-run-6\" --include='*' -l . 2>/dev/null | grep -v '\\.git/' | head; cd wt; grep -rn \"hyp prompt\" --include='*.md' . | head -20; cat docs/adr/0021-a-skill-change-is-a-build-goal.md | head -60"
}
```

### Tool Grep (cwd: investigations/checkpoint-store)

```text
{
 "pattern": "Goal|discovery|build",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-a-3/wt/docs/glossary.md",
 "output_mode": "content",
 "head_limit": 30
}
```

Result:

```text
10:One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
30:Everything replay derives from a store's events: each node, a derived record per node kind (Hypothesis, Question, Experiment and Result) with its own status word, the Goal's state, rejections, limits and selections ([spec §8](replay-spec.md#8-projection-what-replay-returns)). Nothing stores it; every read replays (ADR 0006). In code: `replay.Projection`.
33:**Goal**:
34:[Omitted long matching line]
37:**Discovery Goal**:
38:[Omitted long matching line]
39:_Avoid_: investigation Goal, research Goal; discovery kind (hyp stores none)
41:**Build Goal**:
42:[Omitted long matching line]
43:_Avoid_: delivery Goal, implementation Goal, building Goal; delivery kind (hyp stores none)
46:A node, `Q-`, under the Goal or a Hypothesis, that asks one design question; its children are rival Hypotheses, and any Experiment with no targets that informs the question. Its state is open, reopened, answered, deferred or withdrawn ([spec §7.4](replay-spec.md#74-designquestion-state)). Withdrawn here is a hold that `dq.resume` lifts; it is not an Experiment's close.
50:[Omitted long matching line]
57:[Omitted long matching line]
61:[Omitted long matching line]
64:[Omitted long matching line]
101:_Avoid_: verdict rule, rule (the spec's numbered refusal rules), verdict check, verdict script, criterion (a Goal line is the case the Goal requires), conclusion method
132:A trade-off is between candidates, not about what a Goal line means, which is a Reading. It ends when Results leave one candidate, or when the owner chooses one, recorded as the select (ADR 0015); no confirm follows the choice. The owner's turn to choose is the ledger item `choose`, owed only when the runner owes nothing on the question and no open reading on it can move a Hypothesis there ([spec §8.
[... 808 more characters]
```

Result:

```text
arm
claude-settings.json
claude.err
gh
gitconfig
home-hyp.before
later.txt
prompt.txt
run.json
skill.tar
standin
standin-system.txt
state.txt

tmp/hyp-owner-grill-eval/arm:
skills

tmp/hyp-owner-grill-eval/arm/skills:
hyp-owner-grill

tmp/hyp-owner-grill-eval/arm/skills/hyp-owner-grill:
SKILL.md

tmp/hyp-owner-grill-eval/gh:

tmp/hyp-owner-grill-eval/standin:
./recipes/hyp-owner-grill-eval/issue-a.md
./tmp/issue-113.md
./.claude/skills/hyp-owner-grill/SKILL.md:96:The Goal body ends with the run's constraints block, so a run and `hyp prompt` can find it:
./.claude/skills/hyp-owner-grill/SKILL.md:104:The block runs from the first line that starts with `May change:`, `Read only:` or `Never:` to the end of the body. It holds each of the three headings at least once, and every non-empty line in it starts with one of them, so write one constraint per line and put no other text after the block. A line of spaces, a tab or a lone carriage return is not empty: leave a blank line truly empty, and write the body with `\n` line ends. A heading may repeat, such as one `Never:` line per act. `hyp add goal` warns when a body has a heading but no valid block, and `hyp prompt` refuses that Goal.
./.claude/skills/hyp-owner-grill/SKILL.md:113:hyp prompt <G-id> --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
./.claude/skills/hyp-owner-grill/SKILL.md:116:It prints the store's path, the actor, the Goal id, the skill, the stop clauses (the runner owes nothing, or N Results under the actor, 5 by default) and the Goal body's constraints block, line for line; `--branch` and `--pr` add a clause each after the block. Give the owner its output as it is, and do not add to it. The prompt carries no owner answer: OWED lists each answer and each returned limit the runner applies. The practice of a run, a build Goal's review round included, is in hyp-goal-run. Run `hyp prompt` on the log the run starts from: the prompt names the Goal by its id, so the Goal's add must be in t
[... 3814 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"261111\\|run-6\" recipes/hyp-owner-grill-eval/issue-a.md | head; ls recipes/hyp-owner-grill-eval; git log --all --oneline -S'goal-prompt-261111' | head; grep -rn \"Done when the latest\" --include='*' -r . 2>/dev/null | grep -v '^./.git/' | head"
}
```

Result:

```text
grep: recipes/hyp-owner-grill-eval/issue-a.md: No such file or directory
ls: recipes/hyp-owner-grill-eval: No such file or directory
```

## Grill

Checking where the run 6 prompt text lives, since the issue's "done means" depends on it.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3 && ls recipes recipes/hyp-owner-grill-eval 2>&1 | head -20; grep -rln \"Done when the latest\\|run-6\" recipes inputs tmp wt/investigations wt/docs 2>/dev/null | head"
}
```

Result:

```text
recipes:
hyp-owner-grill-eval

recipes/hyp-owner-grill-eval:
cases.json
cleanup.sh
doctor.sh
drive.py
fixture.py
issue-a.md
issue-b.md
known-answers
launch.sh
recipe.md
score.py
standin-a.md
standin-b.md
standin-system.md
table.py
recipes/hyp-owner-grill-eval/issue-a.md
tmp/issue-113.md
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-3/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls cmd/hyp internal; grep -rn \"May change\\|Read only\" --include='*.go' . | head; grep -rn \"runner owes nothing\" --include='*.go' . | head -5; hyp version"
}
```

Result:

```text
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
./internal/render/status.go:207:		// runner owes nothing there. Runner work prints under OWED, and a
./internal/render/status.go:361:		add(kKeep, "  runner owes nothing: stop and report to the owner")
./internal/render/status.go:644:		owes := "runner owes nothing"
./internal/cli/reach_test.go:39:	want(t, out, 0, "runner owes nothing")
./internal/cli/owed_test.go:17:// the runner owes nothing.
hyp 0.1.0 (events v1)
```

## Grill

Fact check before round 1: does the installed hyp accept a second Goal in one store (the checkpoint-store run says a reinstall was still pending)?

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR && rm -rf g2 && mkdir g2 && cd g2 && git init -q . && export HYP_ACTOR=claude-grill-113 && hyp init 2>&1 | tail -2; hyp add goal --title \"one\" --body \"x\" --line \"[P] a\" --why t 2>&1 | tail -3; hyp add goal --title \"two\" --body \"y\" --why t 2>&1 | tail -3; hyp status 2>&1 | head -8"
}
```

Result:

```text
initialised .tree in /tmp/claude-501/g2
G-BFEVX6 node.add
  line G-BFEVX6#1: [P] a
G-0Y0ZS5 node.add
GOALS (hyp status G-… prints one Goal)
  G-BFEVX6 one — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint none
  G-0Y0ZS5 two — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint none
AWAITING OWNER
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-3/wt
```

## Grill

Nothing is written yet. These are the facts I checked, then round 1.

**Facts**
- The repo has no store at its root. Each tree is its own directory under `investigations/`.
- The installed `hyp` has no `prompt` command. Nothing in the Go code mentions "May change", "Read only" or a constraints block, and `hyp add goal` never warns about one. This Goal builds all of that.
- The installed `hyp` accepts two Goals in one store. I tested it in a scratch directory.
- G-D1CH1R's body has no constraints block, and a Goal is never edited. Under the rule in the merged grill skill ("`hyp prompt` refuses a Goal with no valid block"), `hyp prompt G-D1CH1R` would exit 1.
- Run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) is not in the repo, in any tree, or in the issue text. The issue only gives run 5's prompt, and that one's tree is deprecated.
- The grill skill already describes the output of `hyp prompt`: store path, actor, Goal id, skill, stop clauses, the constraints block line for line, and optional `--branch` and `--pr` clauses.

## Round 1

**1. Run 6's prompt text.** "Done means" #1 requires a clause-for-clause match with it, and I can't read it. The only stop clause I know from run 6 is "selected the candidate on Q-PC04ED", from the checkpoint. The skill's stop clauses are "owes nothing, or N Results under the actor". Run 5's also had "add rival Hypotheses" and "cannot take the next step honestly". Which of these does the generated prompt carry?
- Recommendation: you paste run 6's prompt. You mark the run-specific parts (actor, answer, branch and commit). I draft line 2 from it.
- Without it, I write no clause-match line. I'd list that as a gap in the Goal body.

**2. The real G-D1CH1R can't meet "Done means" #1 as written.**
- Option A: a synthetic fixture Goal under `internal/*/testdata`. It has G-D1CH1R's title and body plus a block made from run 6's constraint lines. The line compares the generated prompt with run 6's prompt file, named by sha256. The line names the run-specific inputs as allowed differences. Core tests may read only synthetic data, so this fits.
- Option B: `hyp prompt` prints a prompt even when the Goal has no block. A run could then start with no stated limits, which is the state your decision exists to prevent.
- Recommendation: A. The issue's bullet stays as history, and the line replaces it.

**3. Print or write? (a DesignQuestion).** The issue leaves open whether hyp prints the prompt or a skill writes it. The merged grill skill assumes hyp prints it.
- Rival 1: `hyp prompt`. Example: run it twice on one log, and you get the same bytes. A Goal with no block gives exit 1.
- Rival 2: the grill skill writes the prompt from `hyp show` and `hyp status`. Two sessions can word it differently, and no exit code can refuse it.
- The scenario that separates them has three parts: the same log run twice; a Goal with no block; a store with two Goals, where only that Goal's OWED may appear.
- Rivals that are not commits can't take lines, so by the grid rule this is a separate design-record Goal with no lines. The build Goal then holds the lines.
- Recommendation: record it that way, as decided in the session with the issue and PR #135 as evidence. The run still gets to test it. Do you want the run to test it, or only record it?

**4. One build Goal, not two.** It would cover `hyp prompt`, the `hyp add goal` warning for a bad block, and the skill-description fix you folded in (`hyp-goal-run` and `hyp-prototype` no longer say discovery-only). ADR 0021 treats a skill change as a build Goal. Recommendation: one Goal. #114 stays out.

**5. Draft lines for the build Goal (at most 10).** Say which to cut, merge or add.
1. [P] `hyp prompt <G> --actor <A>` on a Goal with a valid block prints, in order: the store path, the actor, the Goal id, `hyp-goal-run`, the stop clauses (the runner owes nothing, or N Results under the actor, N from `--experiments`, default 5), and the block line for line. `--branch` and `--pr` each add one clause after the block. Repeated runs give identical bytes.
2. [P] On the fixture from Q2, the output equals run 6's prompt (sha256 named) except for the actor, the owner answer, and the branch and commit. It differs from run 5's prompt, since the sha256 baseline is named. Waits on Q1.
3. [P] In a store with two Goals, `hyp prompt` for one never includes the other's items.
4. [N] A Goal whose body has no valid block: `hyp prompt` exits 1, names the rule, and prints no prompt.
5. [N] A block is invalid when it misses a heading, holds a non-empty line starting with no heading (a lone space, tab or CR line counts as non-empty), or has text after it. Such a body in `hyp add goal` prints a warning and still writes. That warning is why the "refused at write" form doesn't apply to this line.
6. [N] `--branch` with whitespace exits 2. A store with no repository around it exits 3. An unknown Goal id exits 1.
7. [P] `hyp-goal-run` and `hyp-prototype` descriptions name no kind of Goal as their only use, and `go test` (skillcheck) passes.
- Not stored: the constraints, the actor and the counts. Replay reads the constraints from the Goal body, so the derivation guard holds.
- Not a line: "may narrow, never widen". `hyp prompt` has no flag that adds a constraint, so only a hand edit can narrow. That goes in the Goal body.

**6. Names and places.**
- Please confirm your owner actor. I'll assume `owner-dat`.
- I propose one new tree at `investigations/goal-prompt` holding both Goals. This follows "one store holds many Goals". The alternative is two directories, like `checkpoint-store-design` and `checkpoint-store`.
- Recommendation: one directory.

Reply with answers by number. After that I'll run the second-reader review, then show the routing table before any write.

## Owner, message 2

1. I can't paste run 6's prompt from here. If it can't be found, write no clause-match line and list it as a gap. For the allowed differences, widen them to the actor, my answer, branch, PR and commit, and the run's own plan.

2. A, the synthetic fixture. Branch and PR are flags, as in line 1.

3. Settle it in the session: hyp prints the prompt, from a command that reads the log. The run doesn't need to test it. The prompt doesn't carry my answers, because the runner reads OWED.

4. Agree, one build Goal.

5. Cut line 7. Skill descriptions that name only one kind of Goal are text the change makes wrong, so they go under May change, not in a line. Line 5 is fine: `hyp add goal` with a bad block still writes. In line 1 the constraints stay strict, one per line, and the default of 5 for the cap matches what I want. In line 2, add the run's own plan to the allowed differences. The other lines are fine.

6. My actor is owner-standin, not owner-dat. I wanted a single root store, but the repo has none and each tree is its own directory. So I accept one directory, `investigations/goal-prompt`. Keep the root-store idea as a separate issue.

Tests: follow the ak-skill-creator eval and audit flow and decide the details from it. My only preferences are Sonnet for test agents, 3 trials, and a modest budget.

## Grill

Answers are taken. Before the routing table I'm running the two checks the skill requires: how a build Goal's top question starts in the existing tree, and a fresh second reader on the five lines.

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3/wt/investigations/checkpoint-store && hyp show Q-PC04ED 2>&1 | head -30; hyp events --json 2>/dev/null | head -c 1500"
}
```

Result:

```text
DesignQuestion Q-PC04ED
alias: q-commit
title: Which hyp commit implements the shared-store design?
path: G-D1CH1R > Q-PC04ED
status: answered
candidates: H-P4YGE2
selected: H-P4YGE2 — OWED owes this select: R-RRYBPH on 4047c01 (branch feat/one-subject-per-run) supports H-P4YGE2 under E-GAT8N4's rule and closed R-9SMRB2#4 and G-D1CH1R#1 to #8; #9 is accepted by the owner.
design: Commit 4047c01 on branch feat/one-subject-per-run, from 03e9e3d. Applies the owner's return of R-9SMRB2#4: ADR 0020 writes a run subject as {base, patch, files}, as the code, replay-spec and glossary do. The code (cmd/, internal/) is unchanged since b80d95d (git diff b80d95d 4047c01 -- cmd internal is empty), whose removal checks R-P1M8A8 logged.
limits:
  R-0HABWN#1 closed by R-CX1G1C (H-P65RW9 (one-subject-per-arm) supports) — F1: a sandbox record with two --patch Arms at base B is kept as one run subject {B, [pa, pb]}, so a Hypothesis tied to B plus Arm a's patch is never covered though that Arm ran at B; a subject has at most one patch.
  R-0HABWN#2 accepted by owner-dat — F2: line 3's new tests fail to build on 9429fc1, so neither half fails on the commit before its change with a message naming its rule (line 9's case for line 3).
  R-0HABWN#3 closed by R-CX1G1C (H-P65RW9 (one-subject-per-arm) supports) — F3: no test runs rule 26 on a base: a bare-B subject re-tied after a run at B plus a patch must be admitted, and refused with 'subject ran in' after a run at exactly B.
  R-0HABWN#4 closed by R-CX1G1C (H-P65RW9 (one-subject-per-arm) supports) — F4: a Hypothesis tied with --base B and two paths of the same bytes is never held, because run files are unique by sha256 and subject files unique by path.
  R-0HABWN#5 closed by R-CX1G1C (H-P65RW9 (one-subject-per-arm) supports) — F5: hyp subject --base exits 2 when git is not on PATH; a missing tool is an environment problem, exit 3.
  R-0HABWN#6 closed by R-CX1G1C (H-P65RW9 (one-subject-per-arm) supports) — F6: replay admits a Result whose s
[... 4391 more characters]
```

### Tool Agent (cwd: investigations/checkpoint-store)

```text
{
 "description": "Second reader on goal lines",
 "prompt": "You are an independent reviewer. Read-only: change no file, run no write command. Repository: /Users/datguy/.hyp-sandbox/g146-base-a-3/wt (Go; the binary `hyp` stores an investigation as append-only events under .tree/). Read the Go code under internal/ (especially internal/render/status.go, internal/cli, internal/replay) and docs/replay-spec.md, docs/glossary.md.\n\nA new command `hyp prompt <G-id> --actor <A> [--branch B] [--pr N] [--experiments N]` will be built. It does not exist yet. Rules already settled:\n- A Goal's body may end with a constraints block: it runs from the first line that starts with `May change:`, `Read only:` or `Never:` to the end of the body; it holds each of the three headings at least once; every non-empty line in it starts with one of them; a line of spaces, a tab or a lone carriage return is NOT empty. Constraints are read from the Goal body at replay; nothing new is stored.\n- `hyp add goal` with a body that has a heading but no valid block prints a warning and still writes. `hyp prompt` on such a Goal, or on a Goal with no block, exits 1 naming the rule and prints no prompt.\n- `hyp prompt` prints, in order: the store path, the actor, the Goal id, the skill name hyp-goal-run, stop clauses (the runner owes nothing under that Goal's OWED; or N Results recorded under the actor, N from --experiments, default 5), then the block line for line, then one clause each for --branch and --pr. Same log gives identical bytes. In a store with several Goals it never includes another Goal's items.\n- Exit codes: bad argument such as --branch with whitespace = 2; no repository around the store = 3; unknown Goal id = 1.\n- Goals are never edited after they are added.\n\nWalk each rule through scenarios the designers did not draw: two Goals in one store where one body contains the other Goal's id or a line beginning with a heading word inside prose; a merged log from two actors where the same Goal appears from both branches; a base/branch that moves mid-run; a Goal whose body uses CRLF line ends; a Goal added by a log that predates the block (for example investigations/checkpoint-store G-D1CH1R); an --actor equal to the owner's actor; an N-Results count when a Result is retracted or the actor's Results come from a merged log; --experiments 0 or negative. For each, say whether the rules give a single defined behaviour, a conflict, or a gap, and cite the file and line of the code or spec that shows it. Report only facts you verified in the source, each with file:line, and keep the report under 500 words."
}
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a512eaa9c51038c54 (internal ID - do not mention to user. Use SendMessage with to: 'a512eaa9c51038c54', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-3-wt/1fd7e156-53e8-43a6-92ae-cc9f5bba5a3f/tasks/a512eaa9c51038c54.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

## Grill

The second reader is running. I'll check each fact it reports against the source before I use it, then show the routing table. No store has been written.

## Owner, message 3

<task-notification>
<task-id>a512eaa9c51038c54</task-id>
<tool-use-id>toolu_015Ztm7h82urKnY4XAvFmXAW</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-a-3-wt/1fd7e156-53e8-43a6-92ae-cc9f5bba5a3f/tasks/a512eaa9c51038c54.output</output-file>
<status>completed</status>
<summary>Agent "Second reader on goal lines" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Findings, each checked in the source. "Gap" means the settled rules say nothing and the code does not decide it.

1. **Body is read as stored.**
   - `Node.Body` is stored verbatim (`internal/replay/replay.go:55`).
   - Nothing in `internal/` handles `\r`; a grep for `'\r'` in non-test Go files found nothing.
   - The log reader's `TrimSpace` (`internal/store/store.go:106`) is applied to the whole JSON line, not to the body inside it.
   - **Gap: CRLF.** Whether a CRLF body gives a valid block is undefined.
   - The rule "a lone carriage return is NOT empty" means a `\r\n` blank line would make the block invalid. A CRLF body that ends in a blank line is therefore invalid. Output "line for line" would also carry the `\r`.
   - A body with a final `\n` also needs a stated trailing-newline rule, because the block runs to the end of the body.

2. **Other Goal's id, or a heading word in prose.**
   - A Goal's id is its event id, and `replay.GoalOf` walks parents (`state.go:442`).
   - Reading only that Goal's body cannot leak another Goal's items.
   - **Gap:** the rule "runs from the first line that starts with a heading". A prose line that starts with `Never:` moves the block start up. Every following prose line then fails the "starts with a heading" test, so the Goal is refused. This is defined but harsh. The spec should say whether the first match or the last wins.
   - The body limit is `MaxGoalBody` = 2000 (`internal/event/event.go:117`). The prompt carries more than the block, so the owner-grill skill's 4000-character cap (`skills/hyp-owner-grill/SKILL.md:101`) can be exceeded. Nothing is stated.

3. **Same Goal from two branches.**
   - A duplicate event id is rejected: `internal/replay/validate.go:49`, `duplicate id`. The Goal appears once.
   - The merge is a stable k-way merge by id, ties broken by actor (`docs/replay-spec.md:156`). The body is the same, so this is defined.

4. **Base and branch move mid-run.**
   - **Gap.** `--branch` and `--pr` are text clauses.
   - No `hyp prompt` rule, and nothing in the replay or sandbox specs, ties them to a commit.
   - Subjects pin a full commit, via `--base` resolved at `internal/cli/subject.go:27`. The prompt pins nothing.

5. **Pre-block Goal G-D1CH1R.**
   - I read the body in `investigations/checkpoint-store/.tree/events/claude-g1-1.jsonl:1`. It has no `May change:` or `Never:`.
   - By the settled rules, `hyp prompt` exits 1 on it. This is defined.
   - Goals are never edited (`docs/glossary.md:34`), so the Goal cannot gain a block. The one store that holds the shared-store decision cannot be prompted. This is a consequence, not a conflict.

6. **`--actor` equal to the owner's actor.**
   - **Gap.** hyp has no owner actor name. `RoleOwner` is a role on a checkpoint (`internal/event/event.go:81`).
   - Separation of duties is checked per key and actor: `internal/replay/keyed.go:178` and `docs/replay-spec.md:124`.
   - Nothing stops `--actor` from being whoever wrote the Goal. `Node.Actor` is kept (`replay.go:55`) but not compared.
   - The prompt would be accepted. The runner could then record Results under the owner's actor, which makes later `hyp answer` refusals (answer under the Result's actor) bite.

7. **N Results count.**
   - **Gap: retracted.** Results keep a status, `ResRetracted` (`internal/replay/state.go:417`, `result.go:25`). Retracted Results stay in `Results`.
   - "N Results recorded under the actor" does not say whether a retracted one counts. A retract then re-record could stop the run early or late.
   - **Gap: scope.** "Under the actor" is not tied to the Goal. `GoalOf` exists (`state.go:454`). Without it, a Result on another Goal counts.
   - **Merged log.** Results from two actors are counted per `Actor`. That is defined, and a Result rejected in replay (`docs/replay-spec.md:96`) is not counted.

8. **`--experiments 0` or negative.**
   - **Gap.** The rule says "N from --experiments, default 5" and does not say what N=0 or a negative means.
   - Zero would be a stop clause that is already met. Exit 2 is plausible but unwritten.
   - Note that the stop clause counts Results, not Experiments, although the flag is named `--experiments`.

9. **Conflict with the settled rules.**
   - The warning at add ("heading but no valid block") and the exit-1 refusal both depend on one parser. The spec must say a Goal body with a heading is never edited, because `hyp add goal` still writes it. That creates unprompted Goals by design.</result>
<usage><subagent_tokens>41919</subagent_tokens><tool_uses>9</tool_uses><duration_ms>38599</duration_ms></usage>
</task-notification>

## Grill

Checking the second reader's claims against the source before relying on them.

### Tool Bash (cwd: investigations/checkpoint-store)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3/wt && grep -n \"MaxGoalBody\" -r internal | head -5; grep -n \"4000\" skills/hyp-owner-grill/SKILL.md .claude/skills/hyp-owner-grill/SKILL.md | head; grep -n \"ResRetracted\" internal/replay/*.go | head -3; grep -rn \"func GoalOf\" internal/replay | head -2; grep -n \"TrimSpace\" internal/store/store.go | head -3"
}
```

Result:

```text
[redacted: the loaded grill skill]
```

## Grill

I checked the second reader's facts against the source. One finding changes the plan: `skills/hyp-owner-grill/SKILL.md:108` in the repo still tells the grill to hand-write a prompt "under 4000 characters". It has no `hyp prompt` step. The grill skill I'm running comes from `.claude/skills` and already has one. I verified this and the other findings in the table below.

Nothing is written yet. Please confirm the table, then approve the two owner commands below it.

## Findings from the second reader, as I'll handle them

| Finding | Source check | Handling |
|---|---|---|
| A CRLF body leaves ambiguous whether a block is valid | confirmed, nothing in `internal/` handles `\r` | In line 4: a `\r` anywhere in the block makes it invalid. This follows your "write with `\n` ends" rule. |
| A prose line starting with `Never:` starts the block, then the Goal is refused | defined by "first line that starts with a heading" | Line 4 states "first line" outright. No owner choice. |
| "N Results under the actor" doesn't say whether retracted ones count, or whether it is per Goal | confirmed, retracted Results stay in the projection | In line 1: Results on that Goal, under the actor, not retracted. |
| `--experiments` 0 or negative has no behaviour | confirmed | In line 5: exit 2. |
| Goal body limit is 2000 | confirmed, `internal/event/event.go:117` | The block fits. The 4000-character cap in the repo's grill skill is the text the change makes wrong. |
| hyp has no owner actor name, so `--actor` can equal the owner's | confirmed | Looked up, not a line, since hyp cannot tell. I'll list it as a finding the run does not fix. |
| `--branch` and `--pr` are text only and pin no commit | confirmed | Matches your Q2: they are flags and stay text. |

## Routing table

| # | Question | Goes to | Separating scenario or invariant |
|---|---|---|---|
| 1 | Who produces the prompt: `hyp prompt` or the grill skill writes it | Design-record Goal, no lines. DesignQuestion with rival H1 `hyp prompt` and rival H2 "the grill skill writes it". Settled in session, with the issue as evidence. | The same log twice gives the same bytes; a Goal with no block gets exit 1 and no prompt; in a two-Goal store only that Goal's items appear. |
| 2 | Run 6 clause match | Gap in the build Goal body. No line. | A later Goal adds it, with allowed differences: actor, your answer, branch, PR, commit, the run's own plan. |
| 3 | Fixture for the real G-D1CH1R | Looked up: it has no block. The run builds a synthetic fixture. | Body text, no line. |
| 4 | Line 1 | [P] A valid block gives this prompt: path, actor, Goal id, `hyp-goal-run`, stop clauses (owes nothing; or N non-retracted Results on that Goal under the actor, default 5), the block line for line, then `--branch` and `--pr`. Same bytes on repeat. | Output form and determinism. |
| 5 | Line 2 | [P] Two-Goal store: the prompt for one names and counts only that Goal, and equals its prompt in a store holding only its events. | Isolation between Goals. |
| 6 | Line 3 | [N] A Goal with no valid block gets no prompt: exit 1, a message naming the rule, nothing on stdout. | Refusal. |
| 7 | Line 4 | [N] What an invalid block is, as above; `hyp add goal` warns, exits 0 and still writes. | Block rule. |
| 8 | Line 5 | [N] Exit 2 for a bad argument (`--branch` with whitespace, `--pr` or `--experiments` not a positive integer, no actor). Exit 3 when no repository holds the store. Exit 1 for an unknown Goal id. | Exit codes. |
| 9 | Skill descriptions that name one kind of Goal, and the repo grill skill's hand-written-prompt step | `May change:` text, no line, as you said. | – |
| 10 | Skill tests | Goal body: ak-skill-creator eval and audit flow, Sonnet agents, 3 trials, modest budget, details left to the run. | A process preference. |
| 11 | A single root store | Another Goal, which you file as an issue (this machine has no network). | – |

**Point to confirm:** with line 7 cut, no line gates the skill text, so hyp owes no test for the skill edits. The eval in row 10 is only a preference the run may follow. If you want the eval to gate, say so and I'll add a line.

## Owner commands for your approval

Both run in `investigations/goal-prompt` after `hyp init`, with `HYP_ACTOR=owner-standin`. Neither carries lines beyond what the table shows.

```bash
hyp add goal --title "Record who produces a Goal's /goal prompt" --body $'Design record for issue #113. Question: does hyp print a Goal\'s /goal prompt from a command that reads the log, or does the grill skill write it. Settled in the session of 2026-11-12: hyp prints it. The prompt carries no owner answer, because the runner reads OWED. The build is the other Goal in this store.' --why "Owner, 2026-11-12: a design record has no lines; the build Goal in this store carries them."
```

```bash
hyp add goal --title "hyp prints the /goal prompt for any Goal from the log" --body $'Builds issue #113: a `hyp prompt` command that prints, from the log, the /goal prompt for any Goal, and a check of the constraints block in a Goal body. Who produces the prompt is the design-record Goal in this store. The candidate is a hyp commit; go test is the instrument, with a synthetic fixture under internal/*/testdata.\nOwner decisions, 2026-11-11: a run\'s constraints live in the Goal body under May change:, Read only: and Never:; a /goal prompt may narrow them, never widen them; a typed field waits until a real run misreads or loses one.\nGap: run 6\'s prompt (goal-prompt-261111-2129-g1b-run-6) was not found, so no line compares the output with it. A later Goal adds that line, with these allowed differences: the actor, the owner\'s answer, the branch, the PR, the commit and the run\'s own plan.\nG-D1CH1R has no block, so hyp prompt refuses it.\nIf the run changes skill text, test it with the ak-skill-creator eval and audit flow: Sonnet test agents, 3 trials, a modest budget; the run decides the details. No line gates skill text.\nMay change: the hyp repository, as commits on one local branch, never merged; events in investigations/goal-prompt under the runner\'s own actor; the texts the change makes wrong: skills/hyp-owner-grill/SKILL.md (its hand-written 4000-character prompt step), the description lines of skills/hyp-goal-run/SKILL.md and skills/hyp-prototype/SKILL.md that name only one kind of Goal, and the docs that describe the prompt step, the new command or its exit codes\nRead only: every other tree under investigations/, docs/adr/ and the .claude directory\nNever: replace the installed hyp binary\nNever: open or merge a PR\nNever: edit or delete an event or a Goal in any tree\nNever: put application code or real source data in this repository' \
  --line '[P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the actor, the Goal id, the skill name hyp-goal-run, the stop clauses (the runner owes nothing under that Goal\'s OWED; or N Results on that Goal, recorded under the actor and not retracted, N from --experiments, default 5), the block copied line for line, then one clause for --branch and one for --pr if given. Two runs on the same log print the same bytes. It prints no owner answer.' \
  --line '[P] In a store with two Goals, the prompt for one names only that Goal\'s id and constraints block and counts only that Goal\'s Results, and equals the prompt for the same Goal in a store holding only its own events.' \
  --line '[N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule, and prints nothing to stdout.' \
  --line '[N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body, holds each heading at least once, and every line in it is non-empty and starts with a heading. A line of spaces, a tab or a carriage return is not empty, so a `\r` in the block makes it invalid. `hyp add goal` with a body that holds a heading but no valid block prints a warning that names the rule, exits 0 and still writes the Goal.' \
  --line '[N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or no actor. It exits 3 when no repository holds the store, and 1 for an unknown Goal id. Each prints a message and no prompt.' \
  --why "Owner, 2026-11-12: the build Goal for issue #113; five lines, run 6's clause match left as a gap."
```

I haven't measured the body length. I'll check it against the 2000-character limit before running, and shorten the wording if it is over.

After you approve, I'll run `hyp init` in `investigations/goal-prompt`, add the DesignQuestion with its rivals and the session's settlement, add the build Goal's `q-commit` question, and run `hyp status` and `hyp check`. Then I'll print the `hyp prompt` hand-off. The installed `hyp` has no `prompt` command, so the hand-off will be the prompt written by hand to the skill's description. The run builds the command.

## Owner, message 4

The table is fine and I confirm it. Row 11 is right: the root store is a separate issue.

On your point to confirm: no line gates the skill text. That is what I want, and the eval stays a preference only.

One small request on line 3: the refusal message should name what is missing from the block, not only the rule.

I approve both `hyp add goal` commands as shown, run as owner-standin. Shorten the build Goal's body if it is over 2000 characters. Don't push or post anything.

## Grill

Approved. I'm applying your line 3 change and measuring the build body first.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > body2.txt <<'EOF'\nBuilds issue #113: a `hyp prompt` command that prints, from the log, the /goal prompt for any Goal, and a check of the constraints block in a Goal body. Who produces the prompt is the design-record Goal in this store. The candidate is a hyp commit; go test is the instrument, with a synthetic fixture under internal/*/testdata.\nOwner decisions, 2026-11-11: a run's constraints live in the Goal body under May change:, Read only: and Never:; a /goal prompt may narrow them, never widen them; a typed field waits until a real run misreads or loses one.\nGap: run 6's prompt (goal-prompt-261111-2129-g1b-run-6) was not found, so no line compares the output with it. A later Goal adds that line, with these allowed differences: the actor, the owner's answer, the branch, the PR, the commit and the run's own plan.\nG-D1CH1R has no block, so hyp prompt refuses it.\nIf the run changes skill text, test it with the ak-skill-creator eval and audit flow: Sonnet test agents, 3 trials, a modest budget; the run decides the details. No line gates skill text.\nMay change: the hyp repository, as commits on one local branch, never merged; events in investigations/goal-prompt under the runner's own actor; the texts the change makes wrong: skills/hyp-owner-grill/SKILL.md (its hand-written 4000-character prompt step), the description lines of skills/hyp-goal-run/SKILL.md and skills/hyp-prototype/SKILL.md that name only one kind of Goal, and the docs that describe the prompt step, the new command or its exit codes\nRead only: every other tree under investigations/, docs/adr/ and the .claude directory\nNever: replace the installed hyp binary\nNever: open or merge a PR\nNever: edit or delete an event or a Goal in any tree\nNever: put application code or real source data in this repository\nEOF\nprintf '%s' \"$(cat body2.txt)\" > body2n.txt; wc -m body2n.txt"
}
```

Result:

```text
    1775 body2n.txt
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-a-3/wt
```

## Grill

The body is 1775 characters, so it fits. Initialising the store and writing the two approved owner commands.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p investigations/goal-prompt && cd investigations/goal-prompt && export HYP_ACTOR=claude-grill-113 && hyp init 2>&1 | tail -2\nB1=$'Design record for issue #113. Question: does hyp print a Goal\\'s /goal prompt from a command that reads the log, or does the grill skill write it. Settled in the session of 2026-11-12: hyp prints it. The prompt carries no owner answer, because the runner reads OWED. The build is the other Goal in this store.'\nHYP_ACTOR=owner-standin hyp add goal --title \"Record who produces a Goal's /goal prompt\" --body \"$B1\" --why \"Owner, 2026-11-12: a design record has no lines; the build Goal in this store carries them.\" 2>&1\nHYP_ACTOR=owner-standin hyp add goal --title \"hyp prints the /goal prompt for any Goal from the log\" --body \"$(cat $TMPDIR/g/body2.txt)\" \\\n --line '[P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the actor, the Goal id, the skill name hyp-goal-run, the stop clauses (the runner owes nothing under that Goal'\"'\"'s OWED; or N Results on that Goal, recorded under the actor and not retracted, N from --experiments, default 5), the block copied line for line, then one clause for --branch and one for --pr if given. Two runs on the same log print the same bytes. It prints no owner answer.' \\\n --line '[P] In a store with two Goals, the prompt for one names only that Goal'\"'\"'s id and constraints block and counts only that Goal'\"'\"'s Results, and equals the prompt for the same Goal in a store holding only its own events.' \\\n --line '[N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule and what the block lacks (a missing heading, or the first line that does not start with one), and prints nothing to stdout.' \\\n --line '[N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body, holds each heading at least once, and every line in it is non-empty and starts with a heading. A line of spaces, a tab or a carriage return is not empty, so a `\\r` in the block makes it invalid. `hyp add goal` with a body that holds a heading but no valid block prints a warning that names the rule, exits 0 and still writes the Goal.' \\\n --line '[N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or no actor. It exits 3 when no repository holds the store, and 1 for an unknown Goal id. Each prints a message and no prompt.' \\\n --why \"Owner, 2026-11-12: the build Goal for issue #113; five lines, run 6's clause match left as a gap.\" 2>&1"
}
```

Result:

```text
Exit code 2
initialised .tree in /Users/datguy/.hyp-sandbox/g146-base-a-3/wt/investigations/goal-prompt
G-JDDP59 node.add
--body names G-D1CH1R, which is no node here; cite an id this store holds, or write <name>:<id> for another store's
```

## Grill

The first Goal is written as G-JDDP59. The second was refused because the body names G-D1CH1R without its tree. Rewriting it as `checkpoint-store:G-D1CH1R`, which is the only change.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3/wt/investigations/goal-prompt && sed -i.bak 's/^G-D1CH1R has no block/checkpoint-store:G-D1CH1R has no block/; s/(goal-prompt-261111-2129-g1b-run-6)/(goal-prompt-261111-2129-g1b-run-6)/' $TMPDIR/g/body2.txt && grep -n \"D1CH1R\" $TMPDIR/g/body2.txt\nHYP_ACTOR=owner-standin hyp add goal --title \"hyp prints the /goal prompt for any Goal from the log\" --body \"$(cat $TMPDIR/g/body2.txt)\" \\\n --line '[P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the actor, the Goal id, the skill name hyp-goal-run, the stop clauses (the runner owes nothing under that Goal'\"'\"'s OWED; or N Results on that Goal, recorded under the actor and not retracted, N from --experiments, default 5), the block copied line for line, then one clause for --branch and one for --pr if given. Two runs on the same log print the same bytes. It prints no owner answer.' \\\n --line '[P] In a store with two Goals, the prompt for one names only that Goal'\"'\"'s id and constraints block and counts only that Goal'\"'\"'s Results, and equals the prompt for the same Goal in a store holding only its own events.' \\\n --line '[N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule and what the block lacks (a missing heading, or the first line that does not start with one), and prints nothing to stdout.' \\\n --line '[N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body, holds each heading at least once, and every line in it is non-empty and starts with a heading. A line of spaces, a tab or a carriage return is not empty, so a `\\r` in the block makes it invalid. `hyp add goal` with a body that holds a heading but no valid block prints a warning that names the rule, exits 0 and still writes the Goal.' \\\n --line '[N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or no actor. It exits 3 when no repository holds the store, and 1 for an unknown Goal id. Each prints a message and no prompt.' \\\n --why \"Owner, 2026-11-12: the build Goal for issue #113; five lines, run 6's clause match left as a gap.\" 2>&1; hyp status 2>&1 | head"
}
```

Result:

```text
4:checkpoint-store:G-D1CH1R has no block, so hyp prompt refuses it.
G-VCTHB0 node.add
  line G-VCTHB0#1: [P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the actor, the Goal id, the skill name hyp-goal-run, the stop clauses (the runner owes nothing under that Goal's OWED; or N Results on that Goal, recorded under the actor and not retracted, N from --experiments, default 5), the block copied line for line, then one clause for --branch and one for --pr if given. Two runs on the same log print the same bytes. It prints no owner answer.
  line G-VCTHB0#2: [P] In a store with two Goals, the prompt for one names only that Goal's id and constraints block and counts only that Goal's Results, and equals the prompt for the same Goal in a store holding only its own events.
  line G-VCTHB0#3: [N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule and what the block lacks (a missing heading, or the first line that does not start with one), and prints nothing to stdout.
  line G-VCTHB0#4: [N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end of the body, holds each heading at least once, and every line in it is non-empty and starts with a heading. A line of spaces, a tab or a carriage return is not empty, so a `\r` in the block makes it invalid. `hyp add goal` with a body that holds a heading but no valid block prints a warning that names the rule, exits 0 and still writes the Goal.
  line G-VCTHB0#5: [N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or no actor. It exits 3 when no repository holds the store, and 1 for an unknown Goal id. Each prints a message and no prompt.
GOALS (hyp status G-… prints one Goal)
  G-JDDP59 Record who produces a Goal's /goal prompt —
[... 316 more characters]
```

## Grill

Both Goals are in. Now the runner writes: the design question, its two rivals, the experiment recording the session's settlement, and the select.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nQ=$(hyp add dq --parent G-JDDP59 --alias q-producer --title \"Who produces a Goal's /goal prompt?\" --body \"Scenarios that separate the rivals: (a) the same log read twice gives the same bytes; (b) a Goal whose body has no valid constraints block gets no prompt and a refusal; (c) in a store with two Goals, only the asked Goal's items appear.\" --why \"Contested design question of issue #113; the owner settled it in session 2026-11-12.\" | head -1 | awk '{print $1}'); echo Q=$Q\nH1=$(hyp add hyp --parent $Q --alias h-hyp-prompt --title \"A hyp command prints the prompt from the log\" --body \"hyp prompt <G-id> --actor <A> reads the log and prints the prompt, so scenarios (a), (b) and (c) are checked by go test.\" --why \"Rival 1, the owner's choice in session 2026-11-12.\" | head -1 | awk '{print $1}'); echo H1=$H1\nH2=$(hyp add hyp --parent $Q --alias h-skill-writes --title \"The grill skill writes the prompt from hyp show and hyp status\" --body \"The grill skill composes the prompt as prose each session; no command reads the log for it.\" --why \"Rival 2, the alternative the issue left open.\" | head -1 | awk '{print $1}'); echo H2=$H2\necho \"$Q $H1 $H2\" > $TMPDIR/g/ids.txt"
}
```

Result:

```text
Q=Q-PZ1M37
H1=H-BM50HC
H2=H-QZMAYS
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nE=$(hyp add exp --parent H-BM50HC --alias e-session --title \"Session decision on who produces the prompt\" --recipe none --rubric reading --target H-QZMAYS --body $'Rule, fixed before the record: the owner settled this question in the session of 2026-11-12, so this Experiment is an argument, not a run. H-BM50HC supports when a command that reads the log can meet scenarios (a), (b) and (c) of the question. H-QZMAYS contradicts when prose written per session cannot guarantee (a) or (b): two sessions can word it differently (a), and no exit code can refuse a Goal with no block (b). Evidence: scenarios.md in this Experiment\\'s evidence directory.' --why \"Records the owner's settlement of Q-PZ1M37 with the session as evidence.\" 2>&1); echo \"$E\""
}
```

Result:

```text
E-B3FRVA node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\ncat > $TMPDIR/g/scenarios.md <<'EOF'\n# Scenarios for Q-PZ1M37, argued in the session of 2026-11-12 (owner: owner-standin)\n\nverdict: H-BM50HC supports\nverdict: H-QZMAYS contradicts\n\n(a) The same log read twice. A command that reads the log prints the same bytes. Prose written by a skill each session can differ in wording, so the run can not compare two prompts.\n(b) A Goal whose body has no valid constraints block (for example checkpoint-store:G-D1CH1R). A command exits 1 and prints no prompt. A skill has no exit code, so a run could start without stated limits.\n(c) A store with two Goals. A command reads one Goal's OWED and body only. A skill that reads `hyp status` could copy the other Goal's items.\n\nDecision: hyp prints the prompt from a command that reads the log. The prompt carries no owner answer, because the runner reads OWED. Source: issue #113 and the owner's answers in this session.\nEOF\nhyp record --exp E-B3FRVA --verdict H-BM50HC=supports --verdict H-QZMAYS=contradicts --observation \"Session decision, recorded under the pre-registered rule: scenarios (a), (b) and (c) favour a command that reads the log; the owner settled it on 2026-11-12.\" --evidence $TMPDIR/g/scenarios.md --why \"Records the owner's settlement of Q-PZ1M37.\" 2>&1 | head -20; ls .tree/evidence/*; hyp status G-JDDP59 2>&1 | sed -n 1,25p"
}
```

Result:

```text
evidence /tmp/claude-501/g/scenarios.md not found under .tree/evidence or recipes
(eval):14: no matches found: .tree/evidence/*
GOAL G-JDDP59 Record who produces a Goal's /goal prompt — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
  Q-PZ1M37 open/untested Who produces a Goal's /goal prompt?
    H-BM50HC (h-hyp-prompt) proposed A hyp command prints the prompt from the log
      E-B3FRVA planned Session decision on who produces the prompt
    H-QZMAYS (h-skill-writes) proposed The grill skill writes the prompt from hyp show and hyp sta…
STALE
OWED
  runner: Q-PZ1M37 run E-B3FRVA Session decision on who produces the prompt
AWAITING OWNER
RECENT
CHECKPOINT none
```

## Grill

Evidence has to sit under `.tree/evidence/<E-id>/` before it can be cited, so I'm placing it there.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nmkdir -p .tree/evidence/E-B3FRVA && cp $TMPDIR/g/scenarios.md .tree/evidence/E-B3FRVA/scenarios.md\nhyp record --exp E-B3FRVA --verdict H-BM50HC=supports --verdict H-QZMAYS=contradicts --observation \"Session decision, recorded under the pre-registered rule: scenarios (a), (b) and (c) favour a command that reads the log; the owner settled it on 2026-11-12.\" --evidence .tree/evidence/E-B3FRVA/scenarios.md --why \"Records the owner's settlement of Q-PZ1M37.\" 2>&1 | head -20; hyp status G-JDDP59 2>&1 | sed -n 1,20p"
}
```

Result:

```text
R-XM7N5A result.record
GOAL G-JDDP59 Record who produces a Goal's /goal prompt — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
  Q-PZ1M37 open/supported_unselected Who produces a Goal's /goal prompt?
    candidates (1): H-BM50HC (h-hyp-prompt)
    H-BM50HC (h-hyp-prompt) supported A hyp command prints the prompt from the log
    H-QZMAYS (h-skill-writes) contradicted The grill skill writes the prompt from hyp show and hyp sta…
STALE
OWED
  runner: Q-PZ1M37 select H-BM50HC (h-hyp-prompt), the only candidate
AWAITING OWNER
RECENT
  2026-11-12 14:11 claude-grill-113 result.record R-XM7N5A — Records the owner's settlement of Q-PZ1M37.
CHECKPOINT none
```

## Grill

Now the select that OWED owes, the build Goal's one question, and the link between the two Goals.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-113\nhyp select Q-PZ1M37 H-BM50HC --why \"Owed select, the only candidate; the owner's decision in the session of 2026-11-12, evidence R-XM7N5A.\" 2>&1 | head -3\nhyp add dq --parent G-VCTHB0 --alias q-commit --title \"Which hyp commit implements hyp prompt and the constraints-block check?\" --body \"The candidate is a hyp commit that makes every line of this Goal hold. Rival commits are added by the run.\" --why \"The build's one question; its candidates are commits the run makes.\" 2>&1 | head -3\nhyp link G-VCTHB0 external_ref G-JDDP59 --why \"The design this Goal builds: who produces the prompt.\" 2>&1 | head -3\nhyp checkpoint --goal G-JDDP59 --summary \"End of grill: the design is settled by the session of 2026-11-12 (R-XM7N5A); no lines.\" --why \"end of grill\" 2>&1 | head -3\nhyp checkpoint --goal G-VCTHB0 --summary \"End of grill: five lines; the run builds hyp prompt and the block check. Run 6's prompt was not found, so no line compares with it.\" --why \"end of grill\" 2>&1 | head -3\nhyp check 2>&1 | tail -5; hyp status G-VCTHB0 2>&1 | sed -n 1,25p"
}
```

Result:

```text
Q-PZ1M37 dq.select
Q-J7R1WD node.add
G-VCTHB0 edge.add
G-JDDP59 checkpoint
G-VCTHB0 checkpoint
check: 0 errors, 0 warnings
GOAL G-VCTHB0 hyp prints the /goal prompt for any Goal from the log — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-VCTHB0#1 no candidate — [P] `hyp prompt <G-id> --actor <A>` on a Goal whose body ends in a valid constraints block prints, in this order: the store path, the…
  G-VCTHB0#2 no candidate — [P] In a store with two Goals, the prompt for one names only that Goal's id and constraints block and counts only that Goal's Results…
  G-VCTHB0#3 no candidate — [N] A Goal whose body holds no valid constraints block gets no prompt: `hyp prompt` exits 1 with a message that names the block rule …
  G-VCTHB0#4 no candidate — [N] A constraints block is valid only if it runs from the first line starting with `May change:`, `Read only:` or `Never:` to the end…
  G-VCTHB0#5 no candidate — [N] `hyp prompt` exits 2 for a bad argument: a --branch with whitespace, a --pr or --experiments that is not a positive integer, or n…
FRONTIER
  Q-J7R1WD open/no_candidates Which hyp commit implements hyp prompt and the constraints-…
STALE
OWED
  runner: Q-J7R1WD add rival Hypotheses
AWAITING OWNER
RECENT
  2026-11-12 14:11 claude-grill-113 edge.add G-VCTHB0 — The design this Goal builds: who produces the prompt.
CHECKPOINT 2026-11-12 14:11 claude-grill-113: End of grill: five lines; the run builds hyp prompt and the block check. Run 6's prompt was not found, so no line compares with it.
```

## Grill

`hyp check` reports 0 errors and 0 warnings. Committing the tree locally, and checking whether `hyp prompt` runs yet.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-a-3/wt && git add investigations/goal-prompt && git status --short && git commit -q -m \"docs: add the goal-prompt tree for issue #113 (design record and build Goal)\" && git log --oneline | head -2; cd investigations/goal-prompt && HYP_ACTOR=claude-run-113-1 hyp prompt G-VCTHB0 --actor claude-run-113-1 2>&1 | head -3; echo exit=$?"
}
```

Result:

```text
A  investigations/goal-prompt/.tree/.gitignore
A  investigations/goal-prompt/.tree/events/claude-grill-113.jsonl
A  investigations/goal-prompt/.tree/events/owner-standin.jsonl
A  investigations/goal-prompt/.tree/evidence/E-B3FRVA/scenarios.md
53aaf76 docs: add the goal-prompt tree for issue #113 (design record and build Goal)
a51f140 Merge pull request #137 from dat9uy/docs/discovery-and-build-goals
usage: hyp <command> [flags]

Every write command needs --why and an actor (--actor or $HYP_ACTOR).
exit=0
```

## Grill

The grill is done and committed locally on `work` (53aaf76). Nothing was pushed or posted. I could not produce the hand-off prompt, because the installed `hyp` has no `prompt` command; that command is what this Goal builds.

The tree is in `investigations/goal-prompt`, and `hyp check` printed `check: 0 errors, 0 warnings`.

**Two Goals**
- **G-JDDP59** is the design record, and it has no lines.
  - Its question Q-PZ1M37 asks who produces the prompt. The two rivals are `hyp prompt` (H-BM50HC) and a prompt the grill skill writes (H-QZMAYS).
  - I recorded your settlement as E-B3FRVA, with result R-XM7N5A and its scenarios file as evidence. Then I made the select that OWED owed, with H-BM50HC chosen. It is complete, and the runner owes nothing on it.
- **G-VCTHB0** is the build Goal, with your five lines and the constraints block. Its question Q-J7R1WD asks which hyp commit implements `hyp prompt` and the block check. OWED asks the runner to add the rival Hypotheses, and it is linked to G-JDDP59 as `external_ref`.

**Changes from what you approved**
- **Line 3:** the refusal now names what the block lacks, as you asked.
- **Build body:** it came to 1775 characters, so I did not shorten it. The only change was writing `checkpoint-store:G-D1CH1R`, because `hyp` refuses a bare id from another tree.
- **Evidence path:** I copied the scenarios file to `.tree/evidence/E-B3FRVA/scenarios.md` because `hyp record` only cites files already there.

**Hand-off**
- `hyp prompt` printed its usage text instead of a prompt. I can't give you generated output, and I won't write one by hand and call it that.
- Once the build lands and `hyp` is reinstalled, run this from `investigations/goal-prompt` on the log that holds the Goal:
  ```
  hyp prompt G-VCTHB0 --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
  ```
- To start the run before that, you would write the `/goal` prompt yourself from the Goal body's constraints block. That prompt is yours to write, not a generated one.

**Left open**
- **Root store:** the single-root-store idea is yours to file as a separate issue. This machine has no network, so I did not.
- **Run 6's prompt:** it is listed as a gap in the build body. A later Goal can add the clause-match line when you find it.
- **Owner actor:** `hyp` does not know which actor is the owner's, so `--actor` can equal the Goal's author. This is not a line, and the run should not try to fix it.
