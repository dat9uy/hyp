# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

### Tool Bash (cwd: .)

```text
{
 "command": "cat /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md; cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt && hyp status 2>&1 | head -60",
 "description": "Read issue and show hyp status"
}
```

Result:

```text
# Issue #141: Run practice: the last candidate of a budget-bound run goes unreviewed

Opened by dat9uy at 2026-11-12T02:21:06Z. Text as of 2026-11-12T03:45:00Z.

## Event

On 2026-11-12, run 1 of `investigations/goal-prompt` (G-ZBM2RE, PR #140) stopped with its last candidate, H-YXSJEM at `5c5902d`, never reviewed. The two earlier candidates were reviewed, and each review found real defects (10 findings on `10566d4`, 8 on `b4f23b2`). The stop checkpoint says "all review limits closed" and does not name the missing review.

The chain, from the run's transcript and the two trees:

1. A review finding that is a case of a line becomes a limit, and `--limit` exists only on `hyp record`. So the runner recorded each review as its own Experiment: E-C3GGGJ reviewed `10566d4` and E-7QQ2DP reviewed `b4f23b2`.
2. Each candidate therefore cost two Experiments, a gate and a review. The run spent E-HQWTA0, E-C3GGGJ, E-Z5BTZR, E-7QQ2DP and E-NZR3GV (the gate of `5c5902d`), and the stop clause "5 Experiments have been recorded" ended it.
3. The default of 5 was set from G-D1CH1R's runs, where every Experiment gates a commit and review findings went straight into fix commits (one Experiment per candidate; run 5 used 4).
4. At 17:21:37 the runner wrote "E5 is the last Experiment the budget allows, so this fix commit must close everything." `skills/hyp-goal-run/SKILL.md` asks for "one review of each candidate commit's diff" before the stop. The prompt's stop clause and the skill conflict, and neither says which yields or that the gap is named at the stop.

## Scope

Core: the run practice of a Goal whose candidate is a commit (review, budget and stop report) in `hyp-goal-run`, and the stop clause `hyp prompt` prints. No self-review recipe is involved.

## Rivals

A tree holds this question; each option below is a rival Hypothesis, not a decision. The scenario that separates them: replay the G-ZBM2RE cut after R-8JQPWG (one Experiment left, a fix commit pending) and see whether the run 
[... 6052 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls; ls investigations; ls investigations/goal-prompt 2>&1 | head; which hyp; hyp --help 2>&1 | head -40; sed -n 1,200p skills/hyp-goal-run/SKILL.md",
 "description": "Survey repo, investigations and goal-run skill"
}
```

Result:

```text
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
checkpoint-store
checkpoint-store-design
goal-prompt
id-citation
recipe-base
evals
/Users/datguy/.hyp-sandbox/g146-base-b-2/bin/hyp
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
  show REF|R-id#n|G-id#n [--as-of REF|TIME]  one node, or one limit or line, in full, with its his
[... 15226 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "ls -a investigations/goal-prompt investigations/goal-prompt/.tree 2>&1 | head -30; cd investigations/goal-prompt && hyp status 2>&1 | head -50; cd ../..; grep -rn \"experiments\\|Experiments have been\\|Results recorded\" internal/*/prompt*.go cmd 2>/dev/null | head -20; grep -rln \"prompt\" internal | head",
 "description": "Inspect goal-prompt tree and prompt code"
}
```

Result:

```text
investigations/goal-prompt:
.
..
.tree
evals

investigations/goal-prompt/.tree:
.
..
.gitignore
events
evidence
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-11 16:47 claude-grill-113
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-12 03:28 claude-g113-4
AWAITING OWNER
internal/cli/prompt_test.go:63:	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 Results have been recorded under claude-run-1,")
internal/cli/prompt_test.go:331:	want(t, hyp(t, "prompt", g, "--actor", "r", "--experiments", "0"), 2, "--experiments must be at least 1")
internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
internal/cli/prompt.go:123:	if *experiments < 1 {
internal/cli/prompt.go:124:		return usageErr("--experiments must be at least 1")
internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
internal/cli/testdata/prompt-fixture/investigations/checkpoint-store/.tree/events/claude-g1-1.jsonl
internal/cli/testdata/prompt-fixture/investigations/checkpoint-store/.tree/events/claude-g1b-4.jsonl
internal/cli/testdata/prompt-fixture/prompt.golden
internal/cli/prompt_test.go
internal/cli/cli.go
internal/cli/prompt.go
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt && hyp show G-ZBM2RE 2>&1 | head -80; hyp status G-ZBM2RE 2>&1 | head -60; ls .tree/events; cat .tree/.gitignore",
 "description": "Show existing Goal and its status"
}
```

Result:

```text
Goal G-ZBM2RE
alias: build
title: hyp prints the /goal prompt for a Goal from its log
body: Builds the design recorded in G-GZR6C6 (selects H-MCDPNB and H-RSDX0H, with grill session 285fa63e as evidence), for issue #113. The candidate is a hyp commit: the read command hyp prompt and the skill text it makes wrong. go test on the candidate decides lines 1 to 5; line 6 is an eval case in ak-skill-creator's format; line 7 runs its pinned scripts. A skill change follows ak-skill-creator's update flow: matched original and candidate consumer runs on the cases its edited lines touch. The fixture copies run 6's prompt (plans/reports/goal-prompt-261111-2129-g1b-run-6.md) and checkpoint-store:G-D1CH1R's events into testdata, because plans/ and investigations/ are not testdata.
May change: this repository, in commits on the branch the prompt names, that follow AGENTS.md (conventional messages, no AI trailers); the run may push that branch and open or update one PR, and never merges it. It changes cmd/, internal/ and their testdata, and the text the change makes wrong: docs/replay-spec.md, docs/glossary.md, docs/principles.md, AGENTS.md, README.md, and skills/ (hyp-owner-grill's hand-off and Goal-body template, hyp-owner-answer's end, hyp-goal-run's practice from run 6, and the descriptions of hyp-goal-run and hyp-prototype).
Read only: plans/, every investigations/ tree except goal-prompt, and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or ~/.local/share/hyp/hyp-33896ec; compile the candidate to a separate path.
Never: edit recipes/, rubrics/, another investigations/ tree, ak-skill-creator or another repository.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
line G-ZBM2RE#1 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [P] For a Goal whose body holds a valid constraints block, 
[... 14890 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt && sed -n 100,200p internal/cli/prompt.go; git log --oneline | head -30; ls docs docs/adr | head -60",
 "description": "Read prompt.go and docs list"
}
```

Result:

```text
// cmdPrompt prints the /goal prompt for a Goal, derived from the log: the
// store's path, the actor, the Goal, the hyp-goal-run skill, the stop clauses
// and the Goal body's constraints block. It reads no owner answer, so the
// same Goal node and flags give the same bytes in any store.
func cmdPrompt(args []string, stdout, stderr io.Writer) error {
	fs := readFlags("prompt", stderr)
	actor := fs.String("actor", "", "the runner's actor name (required; $HYP_ACTOR is not read)")
	branch := fs.String("branch", "", "the branch the run commits on")
	pr := fs.Int("pr", 0, "the number of the PR the run updates")
	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
	pos, err := parse(fs, args)
	if err != nil {
		return err
	}
	if len(pos) != 1 {
		return usageErr(promptUse)
	}
	if *actor == "" {
		return usageErr("missing --actor: hyp prompt does not read $HYP_ACTOR")
	}
	if !store.ValidActor(*actor) {
		return usageErr("invalid actor %q (want ^[a-z0-9][a-z0-9-]{0,39}$)", *actor)
	}
	if *experiments < 1 {
		return usageErr("--experiments must be at least 1")
	}
	if *pr < 0 {
		return usageErr("--pr must be a PR number")
	}
	if strings.ContainsFunc(*branch, unicode.IsSpace) || strings.ContainsFunc(*branch, unicode.IsControl) {
		return usageErr("--branch %q holds a space or a control character, which would break the prompt's lines", *branch)
	}
	r, err := openReader()
	if err != nil {
		return err
	}
	id, err := resolveRef(r.full, pos[0])
	if err != nil {
		return usageErr("%v", err)
	}
	n := r.full.Nodes[id]
	if n.Kind != event.KindGoal {
		return usageErr("hyp prompt takes a Goal: %s is a %s", pos[0], n.Kind)
	}
	g := event.DisplayID(event.KindGoal, id)
	block, problems, _ := constraintsBlock(n.Body)
	if len(problems) > 0 {
		return herr.Refused("hyp prompt %s: %s", g, strings.Join(problems, "; "))
	}
	// The working directory may reach the store through a symbolic link;
	// the store's place in its repository is read from it
[... 4633 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "find investigations/goal-prompt/evals | head -30; grep -n -i \"store cut\\|--as-of\\|cut\" docs/glossary.md | head -20; ls .claude/skills; grep -rn \"review\" docs/glossary.md | head; cd investigations/goal-prompt && hyp show R-8JQPWG 2>&1 | head -5; hyp show R-JHBCWS 2>&1 | head -12",
 "description": "Look at evals, glossary on cuts and reviews"
}
```

Result:

```text
investigations/goal-prompt/evals
investigations/goal-prompt/evals/goal-prompt-evals.json
investigations/goal-prompt/evals/goal-prompt-evals-2.json
267:One independent, scored repeat of an attempt whose outcome can vary, such as a fresh agent given one task or one session replay from a cut. A rubric counts trials ("T correct of 6"). A trial can take a whole Experiment run, as a session replay does, or share one with other trials. hyp does not record trials; the Experiment's body or method says how many and how each is scored.
353:A Result whose rubric or check hyp cannot vouch for: a script outside what a versioned bundle hashes (one that ends in `.py` or `.sh`, has an execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`), a pinned bundle without a clean command record whose known answers passed, or a command record that matches no pin (spec §7.6).
359:A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. The sandbox is the only place hyp assembles a subject: from the base, a store cut, per-run inputs and one patch for each Arm. Code under test is built in its Launch ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). It is not process isolation: commands still run on the host.
383:_Avoid_: run, execute
397:**Store cut**:
ak-skill-creator
hyp-goal-run
hyp-owner-answer
hyp-owner-grill
hyp-prototype
hyp-verification
docs/glossary.md:3:The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.
docs/glossary.md:46:
[... 4162 more characters]
```

## Grill

I've read issue #141 and its three comments, `hyp-goal-run`, `prompt.go` and the `goal-prompt` tree. Nothing is written yet. This is round 1: the questions that decide the shape of the tree.

**Facts I checked, so they aren't questions:**
- The repository has no root store. The only store that holds these runs is `investigations/goal-prompt`. It has two complete Goals, G-GZR6C6 (design) and G-ZBM2RE (build). R-8JQPWG and R-JHBCWS, the cut points in the issue, exist there.
- The stop clause counts Results recorded under the actor, not Experiments. A gate and a review recorded as two Experiments cost two Results. `--experiments` is only the flag's name.
- G-ZBM2RE line 1 pins "default 5", and `prompt.golden` and `prompt_test.go` pin it too. If H-raise-budget wins, the build Goal that lands it has to change a closed Goal's pinned default. That is fine for a new Goal, because lines are never edited.
- Nothing in the log marks a Result as a review. `hyp status` cannot derive "this candidate is unreviewed". A mark would be a new stored field, so I propose none.

**Q1. One Goal or two?** The issue asks the tree to select a rival and also to change the skill. Today's stand-in:
- Discovery Goal. Its candidates are the rival practices, each as a patch to `skills/hyp-goal-run/SKILL.md` or `internal/cli/prompt.go`. The instrument is a replay of the issue's cut. Nothing is committed to `skills/` or `internal/`.
- Build Goal. Its candidate is a commit that lands the winner.

I recommend two Goals. Write the discovery Goal now, and let the owner add the build Goal after the select. Without that split, every rival needs its own commit and every line is owed on every commit, including the ones that lose. This also follows your rule that a design and its build are two Goals. The cost is one more grill after the select.
- **(a) Two Goals, discovery now (recommended).**
- **(b) One build Goal with one commit per rival.**

**Q2. Which store?** The skill says a new Goal goes in the same store. I recommend `investigations/goal-prompt`, because the issue is Part of #113 and its cuts replay that store. The alternative is a new tree, `investigations/goal-review-budget`. It would need a copy of the cut's events, and issue #138 says `sandbox create` can't yet use a base that holds a `.tree`.
- **(a) Same store, `investigations/goal-prompt` (recommended).**
- **(b) New tree, `investigations/goal-review-budget`.**

**Q3. Purposes outside this Goal.** I'd name these and not write them. Say whether each should become its own Goal later.
- **H-whole-candidate** (the review reads the cumulative diff). It changes what a review finds, not whether the last candidate is reviewed. No line of this Goal needs it, so it would have no line to serve here. I recommend a separate Goal.
- **A review mark** in the log, so `hyp status` could warn about an unreviewed candidate. It is the only way to detect the gap in the binary, as AGENTS.md prefers. It is a new stored field and a design question of its own. I recommend a separate Goal, not now.

**Q4. The run's constraints.** This is my proposed block. Edit anything.
- `May change:` the tree under `investigations/goal-prompt`, its evidence, and one new recipe version for the replay. Candidate patches go only to scratch copies. Commits go on one local branch, and the run opens one PR and never merges it.
- `Read only:` `skills/`, `internal/`, `docs/`, and the other trees.
- `Never:` commit a candidate patch to `skills/` or `internal/`, replace the installed hyp, edit a frozen recipe, add an event kind or stored field, write under the owner's actor.

**Q5. Run details.**
- **Run budget.** I recommend `--experiments 10`. The default 5 can't fit six rivals plus their combinations. It is the same budget problem this issue is about.
- **Branch name.** `feat/review-budget`, unless you prefer another.
- **Actor names.** Your actor name (e.g. `owner-dat`), and I'll pick a fresh runner name.

Round 2, after your answers, is the routing table: each rival with its separating scenario, and the draft lines. The lines are mostly the issue's "Done means" with a baseline pinned by sha256, plus a line that the skill states how a review is recorded and counted.

## Owner, message 2

1 a, two Goals. Discovery now, build after the select.

2 a, same store, `investigations/goal-prompt`.

3 I don't want these split off.
- **H-whole-candidate:** keep it in this Goal as a rival, since every candidate fix goes under the one question and gets tested. I expect it to lose, because one cumulative review costs a lot. Treat it as the expected loser.
- **Review mark:** I'm fine with testing it as a rival too. If hyp works out from the log that a review is owed, that's worth a test. Test it as a scratch patch only, so nothing lands in `skills/` or `internal/`.

4 Confirmed.

5
- **Budget:** `--experiments 10` is fine.
- **Branch:** `feat/review-budget` is fine.
- **Actor:** mine is `owner-standin`. Pick any fresh runner name.

I also have some points for round 2:
- **Test method:** replayed runs on Sonnet, 3 trials per candidate and per starting point. Don't replay a whole build from its start for each trial; that costs too much. Could the ak-context-engineering skill help with how we evaluate?
- **Aims:** each line should serve one of two aims. A review must not easily run out of the run's budget, and a review must not go on and on over small changes. Tell me which aim each line serves.
- **Cost:** a candidate may cost at most 1.5 times the base at the same starting point. How do you compute cost? If you use a median, how is it computed? A trial that hangs counts as the most expensive.

### Tool Bash (cwd: .)

```text
{
 "command": "ls ~/.claude/skills 2>&1 | head -30; find / -iname \"*context-engineering*\" -not -path \"*/proc/*\" 2>/dev/null | head; ls .claude/skills/ak-skill-creator .claude/skills/ak-skill-creator/scripts 2>&1 | head -40; shasum -a 256 skills/hyp-goal-run/SKILL.md internal/cli/prompt.go; grep -n -i \"cost\\|token\\|median\\|trial\\|timeout\" .claude/skills/ak-skill-creator/scripts/eval_skill.py | head -40",
 "description": "Check for context-engineering skill and eval cost support"
}
```

Result:

```text
hyp-goal-run
hyp-owner-answer
hyp-owner-grill
hyp-prototype
hyp-verification
public-doc-voice
synced
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/2.19.0/engineer/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/codex/2.19.0/engineer/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/omp/2.19.0/engineer/.omp/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/omp/2.19.0/engineer/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/pi/2.19.0/engineer/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/pi/2.19.0/engineer/.pi/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.agentkit/backups/20261105T121723Z-c6a93f00/data/root/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/.local/state/nvim/undo/%Users%datguy%codingProjects%context-engineering%main.py
/System/Volumes/Data/Users/datguy/codingProjects/hyp-dev-local/.claude/skills/ak-context-engineering
.claude/skills/ak-skill-creator:
assets
LICENSE-MIT.txt
LICENSE.txt
NOTICE.txt
references
scripts
SKILL.md

.claude/skills/ak-skill-creator/scripts:
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
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
21f9332383bc6b0eaf3d8e581eea0d397d0
[... 739 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering && ls; sed -n 1,25p SKILL.md | cut -c1-400; grep -n -i \"evaluat\\|eval\\b\\|cost\\|trial\" SKILL.md | head -20; cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/.claude/skills/ak-skill-creator/scripts && sed -n 1,60p session_usage.py | cut -c1-200; grep -n -i \"median\\|mean\\|cost\\|timeout\\|hang\" eval_contract.py session_usage.py | head -30",
 "description": "Read context-engineering skill and eval cost tooling"
}
```

Result:

```text
references
scripts
SKILL.md
---
name: ak-context-engineering
description: "Manage context, task budgets, compaction and model choice from outcome evidence, plus memory and agent systems."
when_to_use: "Invoke for token waste, context left, usage limits, long tasks, forgetful sessions or benchmark cost."
category: engineering
keywords: [context, tokens, limits, memory, optimization, compaction, budget, benchmarks]
argument-hint: "[topic or question]"
metadata:
  author: agentkit
  version: "1.3.1"
---

# Context Engineering

Finish the requested work correctly with less wasted context, cost, time and agent
steps. Optimize complete outcomes, including failures, retries and verification.
Never cut required scope or verification to meet an efficiency target.

## Operating loop

1. **Frame:** identify the outcome, constraints and acceptance checks. Reuse an
   accepted plan. Short tasks need no additional plan or ledger ceremony.
2. **Observe:** use available runtime telemetry. Distinguish context occupancy,
   cumulative task spend, provider quota and dollar cost. Missing values are unknown.
3. **Allocate:** reserve room for the next result, response and checkpoint. Reserve
4:when_to_use: "Invoke for token waste, context left, usage limits, long tasks, forgetful sessions or benchmark cost."
15:Finish the requested work correctly with less wasted context, cost, time and agent
24:   cumulative task spend, provider quota and dollar cost. Missing values are unknown.
52:## Cache and cost observations
55:retries and cost per successful task with model/runtime/settings and source identity.
79:Keep all four core metrics: **Average cost per task**, **Average task duration**,
80:**Estimated cost per successful first attempt**, **Average agent steps**.
82:cost per first-attempt success and duration; use raw cost for budget planning and
88:See [model selection](references/model-selection.md) and [evaluation](references/evaluation.md).
91:input, communication and integration cost, a
[... 4738 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator/scripts)

```text
{
 "command": "cd /Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering && sed -n 50,90p SKILL.md | cut -c1-500; sed -n 1,80p references/evaluation.md | cut -c1-500",
 "description": "Read evaluation guidance on cost"
}
```

Result:

```text
  still take precedence over this skill's efficiency heuristics.

## Cache and cost observations

When the host exposes them, record input/output tokens, cache reads/writes, elapsed time,
retries and cost per successful task with model/runtime/settings and source identity.
Missing telemetry remains `null`/unknown, never estimated from file length. Context occupancy
is not cumulative billed tokens or currency. Prefix stability, cache policy and reasoning
effort belong to prompt assembly/adapters or APIs; renaming a skill does not establish a
cache improvement. Compare matched original/candidate runs before changing runtime defaults.

## Capacity and task budgets

Context bands are advisory defaults: Green <50%, Yellow 50–<70%, Orange 70–<85%,
Red ≥85%. They are not calibrated quality thresholds for every model. Available
headroom and the next atomic step matter more than the band. Unknown telemetry
never becomes a percentage based on twenty or forty tool calls.

At Orange or insufficient headroom, finish a safe atomic boundary, checkpoint,
then use a supported compaction capability. In Red, prioritize recovery and the
checkpoint over new exploration. Never abandon a required safety check. If no
compaction tool is available, preserve continuation state without claiming to have
compacted or creating a new session without authorization.

Details: [runtime awareness](references/runtime-awareness.md),
[session protocol](references/session-protocol.md).

## Model selection and delegation

Keep all four core metrics: **Average cost per task**, **Average task duration**,
**Estimated cost per successful first attempt**, **Average agent steps**.
First require sufficient task-specific success and capability. Then compare
cost per first-attempt success and duration; use raw cost for budget planning and
steps to diagnose waste. Do not blend correlated metrics into an arbitrary score.

Compare model + effort + harness on matched task cohorts with dated provenance.
Public benchmark
[... 5308 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering && ls scripts references; shasum -a 256 references/evaluation.md scripts/benchmark_metrics.py SKILL.md 2>&1; cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt && hyp show R-JHBCWS >/dev/null; hyp show E-GE085A 2>&1 | head -12 | cut -c1-300",
 "description": "Hash evaluation files, check run 3 experiment"
}
```

Result:

```text
references:
context-compression.md
context-degradation.md
context-fundamentals.md
context-optimization.md
evaluation.md
memory-systems.md
model-selection.md
multi-agent-patterns.md
project-development.md
runtime-awareness.md
session-protocol.md
tool-design.md

scripts:
benchmark_metrics.py
benchmark_statistics.py
benchmark_validation.py
compression_evaluator.py
compression_grades.py
compression_probes.py
context_analyzer.py
context_budget.py
context_health.py
tests
fdd4a34ff3d1b937ad482a031160b8b9da40f73a20244258d6d38e77ed8d2d0b  references/evaluation.md
2a98cf72638a32e45811da40be66f5b85cb01fd04b7e14cac7297cf963e6942f  scripts/benchmark_metrics.py
ad2ce02b31b8d0595a934267ce726b489044443711f75622a7199e13e88d6383  SKILL.md
Experiment E-GE085A
alias: e-04142f3
title: Gate and review 04142f3 against the seven lines and R-DER6F8's limit
body: Supports H-VHG4N5 only if all hold on 04142f3, else contradicts. go vet ./... and go test ./... -count=1 pass; TestPrompt* run, none skipped. Red by mutation: refuse a bad block line only after the repository lookup and the bad-line check fails (exit 3, want 1). probe-p1.sh passes on the 04142
path: G-ZBM2RE > Q-Z93VJ5 > H-VHG4N5 > E-GE085A
status: run
recipe: none
rubric: reading
shape: reassess from kept records
records: .tree/evidence/E-NZR3GV/binaries-and-mapping.log e16a74c63651
records: .tree/evidence/E-NZR3GV/eval-summary.log 6195c7fc142a
records: .tree/evidence/E-NZR3GV/evals-hash.log 9fbbc3221e3c
```

## Grill

**Round 2: the routing table, the draft lines and your three method points.** Nothing is written yet. All actors and the branch are as you gave them: `owner-standin`, `feat/review-budget`, `--experiments 10`.

## Your three points

**Does ak-context-engineering help?** Yes. Its `references/evaluation.md` asks for:
- the same frozen inputs for every configuration;
- matched trial ids, with several independent trials;
- failed and timed-out runs recorded, never dropped;
- p50 and p95 reported;
- cost left unknown when unobserved, never estimated;
- small samples treated as provisional.

It is not installed under `.claude/skills/` here; it sits only in `~/.agentkit/cache/…/3.0.0/engineer/skills/ak-context-engineering`. I propose it goes under the Goal's `Read only:` line, pinned by sha256 like ak-skill-creator:
- `references/evaluation.md`: `fdd4a34f…2d0b`
- `scripts/benchmark_metrics.py`: `2a98cf72…942f`

**How I compute cost.**
- **Trial cost.** The cost of a trial is every token the replayed run uses from its starting point to its stop, subagent reviews included. The cost is priced in USD from the transcript's usage fields at one dated Sonnet rate table, with cache reads and writes priced separately. Raw tokens are reported beside it. A trial with no usage data counts as unknown.
- **Median.** There are 3 trials per candidate per starting point. Sort the 3 costs ascending and take the second. Starting points are never pooled.
- **Hangs.** A trial that hangs, or whose cost is unknown, counts as +∞, the most expensive. One hang leaves the median finite. Two hangs make it +∞, and the line fails.
- **The test.** The candidate passes when `median(candidate) ≤ 1.5 × median(base)` at that starting point. If the base's median is +∞, the ratio is undefined, and the run records that and refers it.

**Aims.** Each line serves one aim; the table shows which. A = a review must not easily run out of the budget. B = a review must not go on and on over small changes. The validators line from my first draft serves neither, so I dropped it. `skillcheck`, `quick_validate.py` and `go test` belong to the build Goal.

## Routing table

Starting points: **S1** is the goal-prompt store cut after R-8JQPWG (run 1, one Result left, a fix commit pending). **S2** is the cut after R-JHBCWS (run 3, reviews that found no defect still led to commits).

| # | Question | Goes to | Separating scenario |
|---|---|---|---|
| 1 | What run practice ends a budget-bound build run on a reviewed final candidate, without review going on over small changes? | DesignQuestion Q1 under the Goal | S1 and S2, 3 Sonnet trials each, per candidate and for the base |
| 1a | Rivals | H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate (expected loser, tested anyway), H-review-owed (scratch hyp patch only) | as above |
| 1b | Combinations | Goal body: H-name-at-stop is also run with each single that fails L1 alone. A combination is a new rival under Q1. | same |
| 2 | The base | Not a Hypothesis. The lines pin it by sha256: `skills/hyp-goal-run/SKILL.md` `cb15e2f2…750e`, `internal/cli/prompt.go` `21f93323…38ed`. | same |
| 3 | Test method (Sonnet, 3 trials, starting points) | Goal body, plus each Experiment's verdict rule | none, it's process |
| 4 | Constraints, branch, budget, actors | Goal body block and `hyp prompt` flags | none |
| 5 | Landing the winner | Another Goal, added by you after the select | none |

H-review-owed is checked against the derivation guard. The log already shows a candidate Hypothesis and a `--rubric reading` Result that targets it, so a derived "review owed" needs no new field. A gate Experiment looks the same, though, so the rival may need a distinguishing convention. If the run finds it does, it states the field and flags it to you. It proposes no field in advance.

## Draft lines (5; I'll trim each to 600 runes when I write them)

| Line | Aim | Check |
|---|---|---|
| **L1 [P]** At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that commit as unreviewed. A trial that hangs, or that says "all review limits closed" or "runner owes nothing" while a candidate has no review, does not meet it. The base (sha256 above) meets it in at most 1 of 3 trials at each starting point, so the counts differ. | A | An independent reviewer grades each transcript against a rubric, blind to the candidate. |
| **L2 [N]** From S2, a candidate run launches no more than one fresh review before its stop. The base launched two in run 3 and meets this in at most 1 of 3 trials. | B | Count review launches in the transcript. |
| **L3 [N]** At each starting point the candidate's median trial cost is at most 1.5 times the base's median, computed as above. | B | Usage fields, priced at the dated rates. |
| **L4 [N]** A candidate run ends with no review finding that is a case of a line left neither closed, referred, fixed nor named in its stop checkpoint. | A | A reviewer checks the findings against the checkpoint. |
| **L5 [P]** A Sonnet run given only the candidate's `hyp-goal-run` text and the printed prompt answers, in each of 3 trials, "how is a review recorded?" and "does it count toward the stop clause's N?" as the candidate's text says. The base text states neither. | A | Fixed questions with fixed expected answers. |

## Questions for you

1. **Cost denominator.** At S1 the base has one Result left and stops after a gate, without any review. Any candidate that adds a review adds a subagent run on top. A 1.5× limit may therefore fail every candidate that actually reviews. My recommendation is to keep your rule as stated. The run records every ratio, so a Goal where all fail still shows the numbers. The other option is to count the base's skipped review in its cost, priced from an earlier review such as E-7QQ2DP.
   - **(a) Keep as stated (recommended).**
   - **(b) Add the base's skipped review to its cost.**
2. **How many trials must meet L1?** I recommend 3 of 3, because a single hang or miss should count against a candidate. The alternative is 2 of 3.
   - **(a) 3 of 3 (recommended).**
   - **(b) 2 of 3.**
3. **Trial count and staging.** With the base and 7 rivals, there are 8 configurations × 2 starting points × 3 trials = 48 Sonnet trials, plus combinations. I recommend a staged plan: run the base first, then the singles, then combinations only for singles that fail L1 alone. Say if you want a cap.
   - **(a) Staged, no cap (recommended).**
   - **(b) Staged, with a cap on trials. Tell me the number.**
4. **Routing.** Confirm or move each row of the table, and the five lines with their aims. The first draft's validators line is moved out to the build Goal.

After your answers I run the second reader, apply the model to itself, scan the preflight text, and show the exact write commands for your approval.

## Owner, message 3

1 a, keep the 1.5× rule as stated. Record every ratio.

2 a, 3 of 3.

3 a, staged with no cap. Replay from the cut points only, never a whole build from its start.

4 Confirmed. All five table rows and the five lines with their aims stand, and the validators line stays in the build Goal. Pinning the ak-context-engineering files under `Read only:` is fine.

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader of the grill design",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer of a design for a hyp \"discovery Goal\". Read-only: change no file, write nothing under investigations/. Repository: /Users/datguy/.hyp-sandbox/g146-base-b-2/wt (Go; the `hyp` binary is on PATH). Read AGENTS.md, docs/glossary.md, skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go and issue text /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md. The store `investigations/goal-prompt/.tree` holds the runs the issue talks about (use `hyp status`, `hyp show`, `hyp events --json` from that directory).\n\nSettled rules to test (the owner confirmed them):\n- Goal G (new, in the store investigations/goal-prompt, which already holds Goals G-GZR6C6 and G-ZBM2RE). One top-level DesignQuestion Q1: \"What run practice ends a budget-bound build run on a reviewed final candidate, without review going on over small changes?\" Rival Hypotheses (each a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never committed): H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate, H-review-owed. Combinations with H-name-at-stop are new rivals added by the run.\n- Starting points: S1 = store cut after Result R-8JQPWG (run 1, one Result left of 5, a fix commit pending); S2 = store cut after R-JHBCWS (run 3). Replay from the cut only, 3 Sonnet trials per candidate and per start, never a whole build.\n- Lines (written once, never edited): L1 [P] at S1 and S2 in each of 3 trials the candidate run ends with a review Result for its final candidate commit or a stop checkpoint naming it unreviewed; a hang, or \"all review limits closed\"/\"runner owes nothing\" while a candidate has no review, does not meet it; the base (skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, internal/cli/prompt.go sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed) meets it in at most 1 of 3 trials. L2 [N] from S2 the candidate run launches at most one fresh review before its stop. L3 [N] at each start the median trial cost (USD from usage fields at a dated rate table, subagents included, a hang or unknown counts as +infinity, median of 3 = second smallest) of the candidate is at most 1.5 times the base's median. L4 [N] a run ends with no review finding that is a case of a line left neither closed, referred, fixed nor named in the stop checkpoint. L5 [P] a Sonnet run given only the candidate's hyp-goal-run text and the printed prompt answers \"how is a review recorded?\" and \"does it count toward the stop clause's N?\" as the candidate's text says; the base text states neither.\n- hyp owes every line on every top-level question that has a candidate. Statuses are derived from the log by replay.\n- The run is an unattended /goal run with `hyp prompt` printing its stop clause (10 R
[... 1653 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a90a57c8ff3424ae8 (internal ID - do not mention to user. Use SendMessage with to: 'a90a57c8ff3424ae8', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-2-wt/ad21ff27-cd6c-4cc9-a5d1-5a687d1f91c0/tasks/a90a57c8ff3424ae8.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn -i \"run cap\\|no run cap\\|default 5\\|5 Results\\|5 Experiments\\|--experiments\" docs skills README.md AGENTS.md | cut -c1-260 | head -20; git log -1 --format=%H",
 "description": "Scan texts that mention the run budget"
}
```

Result:

```text
skills/hyp-owner-grill/SKILL.md:113:hyp prompt <G-id> --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
9c8e498b87683e50a94a410b6a3cfecdd7a0d441
```

## Grill

The second reader is still running. My own scan of the budget text found nothing that conflicts with the Goal. No doc or skill other than `hyp-owner-grill`'s hand-off command mentions the run budget or its default. I'll write nothing until the reader's report comes back, then show you the exact commands to approve.

## Owner, message 4

<task-notification>
<task-id>a90a57c8ff3424ae8</task-id>
<tool-use-id>toolu_01Ae94KyrzBJUPrSe46dNX4W</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-2-wt/ad21ff27-cd6c-4cc9-a5d1-5a687d1f91c0/tasks/a90a57c8ff3424ae8.output</output-file>
<status>completed</status>
<summary>Agent "Second reader of the grill design" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>I found six concrete defects and one gap, ranked most severe first.

**1. The "base" hashes name HEAD, not the commit a replay starts from.**
- `sha256sum skills/hyp-goal-run/SKILL.md internal/cli/prompt.go` gives the L1 hashes (cb15e2f2…, 21f93323…). They match HEAD.
- `git show b4f23b2:skills/hyp-goal-run/SKILL.md | sha256sum` gives a6152711…. `git show b4f23b2:internal/cli/prompt.go | sha256sum` gives 674dc260….
- S1's `--base` has to be about b4f23b2, because the fix commit is still pending. At that commit neither file matches L1's base.
- The same gap hits candidates. A scratch patch written against HEAD's SKILL.md need not apply at S1's base, and `hyp sandbox create --patch` checks that it applies. L1 needs the same candidate at both starts, so each candidate needs a patch per start.
- Which files the Sonnet trial actually reads is not stated. `~/.claude/skills` is a symlink to the host checkout (docs/install.md:74-75), not the sandbox clone. Nothing says the trial's skills come from the patched sandbox. `--recipe none` has no Launch step to link them.
- Fix: pin the base by commit and by the hash of the files it holds at that commit.

**2. The budget a trial starts with is undefined, so "budget-bound" is not guaranteed.**
- `internal/cli/prompt.go:167` prints "N Results have been recorded under ACTOR" as an absolute count.
- The cut at R-8JQPWG holds 4 Results by claude-g113-1. A trial under a new actor gets a fresh budget of 5, which is run 2's case, not "one left".
- A trial under claude-g113-1 inherits the 4 in the cut. Nothing says which one is meant, or what `--experiments` the trial prompt gets.
- At S2 (R-JHBCWS) run 3 had recorded only 1 Result. The count is 4 left, so the trial may end on "runner owes nothing" and never meet the cap. L1 and L2 would then be judged on runs the cap never bound.
- The issue and the lines say "Experiments", but the flag `--experiments` counts Results.

**3. L5 cannot be met by some candidates and carries two invariants.**
- L5 asks a model two questions and expects them answered "as the candidate's text says". The base states neither, so L5 contradicts any candidate that leaves `hyp-goal-run` unchanged. That covers H-raise-budget (only `prompt.go`) and H-name-at-stop alone.
- Whether the answer matches "the candidate's text" is also unspecified.
- Fix: split L5 into two lines.

**4. L4 and L1 have no pinned instrument.**
- L4 asks whether a review finding "is a case of a line". That is a reading, not a script check. Labelling comes from the trial's own reviewer subagent.
- "Fixed" and "named in the stop checkpoint" are not derived by `hyp status`. Only the closed, referred and open states are derived.
- L1's "final candidate commit", "names it unreviewed" and "hang" need a timeout and a parsing rule. None is given.
- "All review limits closed" is a phrase from run 1's checkpoint, which is a weak matcher.
- Whether L4 fails on any nameable commit is unchecked. The base's sort rule may already pass it.

**5. L3 can pass vacuously and is not N-form.**
- If the base's median is +infinity (2 of 3 trials hang), "at most 1.5 times" holds for any candidate.
- It is a ratio, not a forbidden event.
- A cost that is unknown counts as infinity, so a candidate with missing usage fields fails by default.

**6. H-review-owed cannot be built under "Never".**
- Nothing in the log marks an Experiment as a review. A gate and a review are both `--recipe none --rubric reading`. `hyp status` would need a new stored field, or a guess from body text.
- If the guess is chosen, AGENTS.md says a tree-detectable failure belongs in the binary and is gated by `go test`, not by 3 Sonnet trials. The select would then not reflect how that candidate is gated.

**7. Budget arithmetic for the run itself.**
- 8 candidates (7 rivals plus the base) × 2 starts × 3 trials is 48 trials before any combinations.
- Rule 32 (replay-spec.md:132) and the `aims` rules allow one Experiment to target 8 Hypotheses and one Result to carry 8 verdicts. A Result can also `--closes` G#1 to G#5 together, so 10 Results are enough only if the run batches.
- A Result cannot be edited. One hang in a batch means a new Result.
- Each new combination rival owes all 5 lines, and L3 needs the base's trials cited from an earlier Result.
- L3 and L2 therefore can't be closed by a Result that doesn't cite the base's median, and a candidate is covered only if the Result "gave a verdict on" it.

**Checked and sound**
- `hyp status --as-of R-8JQPWG` replays the third Goal correctly. replay-spec.md:158 cuts at the first event whose id equals `EVENT`, and a Result id is its event id. The output shows G-ZBM2RE `in_progress`, runner owes 1, and R-8JQPWG's 3 open limits.
- `hyp status` with no Goal id in a 3-Goal store prints only the GOALS table and the AWAITING OWNER list. The skill covers this ("run `hyp status &lt;goal-id&gt;`").
- G's lines are keyed `G-&lt;id&gt;#n`. Rule 32 stops G's Results from closing G-ZBM2RE's lines, so the two Goals' lines do not mix.
- The merge is a k-way merge by ULID (replay-spec.md:150-160). A cut or `--as-of` with a later Goal in the log is unaffected by events after the stop.
- A line is capped at 600 runes (rule 31). L1's text must stay under it, and the two hashes fit.
- At S1, `claude-g113-1` has exactly 4 Results, so "one left of 5" is true if the actor is reused (see 2).
- H-raise-budget changing `prompt.go` does not alter the run's own cap, because trials use a scratch binary via `--hyp`. The run's own cap comes from the installed `hyp prompt`.
- Not checked: whether `--recipe none` with `--store/--cut` creates a usable sandbox end to end. I did not run `hyp sandbox create`.</result>
<usage><subagent_tokens>62527</subagent_tokens><tool_uses>7</tool_uses><duration_ms>105335</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "for c in b4f23b2 3cfd08d d2d5d14; do echo $c; git show $c:skills/hyp-goal-run/SKILL.md | shasum -a 256 | cut -c1-16; git show $c:internal/cli/prompt.go | shasum -a 256 | cut -c1-16; done; cd investigations/goal-prompt; for a in claude-g113-1 claude-g113-2 claude-g113-3; do echo $a; hyp events --json | python3 -I -c \"\nimport sys,json\nd=json.load(sys.stdin)\nev=d.get('events',d) if isinstance(d,dict) else d\nn=0\nfor e in ev:\n    if e.get('kind')=='result.record' and e.get('actor')==sys.argv[1]:\n        n+=1; print(n,e.get('id'))\n\" $a; done 2>&1 | head -30; hyp events --json | head -c 400",
 "description": "Verify base hashes and Result counts per actor"
}
```

Result:

```text
b4f23b2
(eval):1: bad substitution
e3b0c44298fc1c14
674dc260b4863afd
3cfd08d
(eval):1: bad substitution
e3b0c44298fc1c14
21f9332383bc6b0e
d2d5d14
(eval):1: bad substitution
e3b0c44298fc1c14
21f9332383bc6b0e
claude-g113-1
claude-g113-2
claude-g113-3
{"format":1,"events":[{"v":1,"id":"01M75R06QEHM6EG2PSEEGZR6C6","ts":"2026-11-11T16:45:32.782596Z","actor":"owner-dat","type":"node.add","node":"01M75R06QEHM6EG2PSEEGZR6C6","payload":{"kind":"Goal","parent":"","title":"Record how hyp makes the /goal prompt: who writes it, and whether it carries owner answers","body":"Design record for issue #113, settled by the owner in the grill session of 2026-11
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt && for c in b4f23b2 3cfd08d; do echo $c; git show \"${c}:skills/hyp-goal-run/SKILL.md\" | shasum -a 256 | cut -c1-16; done; cd investigations/goal-prompt; hyp events --json | python3 -I -c \"\nimport sys,json,collections\nev=json.load(sys.stdin)['events']\nc=collections.Counter(); order=[]\nfor e in ev:\n    if e['type']=='result.record':\n        c[e['actor']]+=1; print(e['actor'],c[e['actor']],e['id'][-6:])\n\" | grep -E \"g113-[123]\"",
 "description": "Recompute hashes and per-actor Result counts"
}
```

Result:

```text
b4f23b2
a6152711c2c930b2
3cfd08d
cb15e2f2e207b5a6
claude-g113-1 1 KPWZZC
claude-g113-1 2 9V3ER0
claude-g113-1 3 50GBXM
claude-g113-1 4 8JQPWG
claude-g113-1 5 FRF929
claude-g113-1 6 9EGPYD
claude-g113-2 1 0E8PFH
claude-g113-2 2 KYKWH9
claude-g113-2 3 28A8NB
claude-g113-2 4 3YJ7T2
claude-g113-2 5 356DV0
claude-g113-3 1 JHBCWS
claude-g113-3 2 27Y1EV
claude-g113-3 3 DER6F8
claude-g113-3 4 BSWFXG
claude-g113-3 5 HXEBAF
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt && for c in b4f23b2 3cfd08d; do for f in skills/hyp-goal-run/SKILL.md internal/cli/prompt.go; do printf \"%s %s \" $c $f; git show \"${c}:${f}\" | shasum -a 256 | cut -d' ' -f1; done; done; git rev-parse b4f23b2 3cfd08d; hyp --help | grep -n \"version\" | head -2; echo $TMPDIR",
 "description": "Get full hashes and commits for both starts"
}
```

Result:

```text
b4f23b2 skills/hyp-goal-run/SKILL.md a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab
b4f23b2 internal/cli/prompt.go 674dc260b4863afdec8ea2a27383ebe909b0736e25278d79aea78ecc4738d08b
3cfd08d skills/hyp-goal-run/SKILL.md cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e
3cfd08d internal/cli/prompt.go 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed
b4f23b2e88f4444319b8261dba7710752665b91c
3cfd08d584b49216881349355240c97cb0bed1c8
8:  version                                print the hyp version
38:  bundle check                           every pin to a versioned bundle, in every tree under the repo root, still matches it
/tmp/claude-501
```

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/dry && cd $TMPDIR/dry && rm -rf .tree && git init -q . 2>/dev/null; cat > $TMPDIR/goal-cmd.sh <<'EOF'\nhyp add goal --title \"Select the run practice that ends a budget-bound build run on a reviewed final candidate, without endless review\" --body $'Discovery Goal for issue #141 (part of #113), from the comments of runs 1 to 3 of G-ZBM2RE (observation entries R10, R14). The candidate is a run practice: a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never a commit. The landing of the winner is a separate build Goal that the owner adds after the select.\\nStart S1: the store cut after R-8JQPWG, repository at b4f23b2 (its skills/hyp-goal-run/SKILL.md sha256 a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab, internal/cli/prompt.go sha256 674dc260b4863afdec8ea2a27383ebe909b0736e25278d79aea78ecc4738d08b), actor claude-g113-1, who then has 4 Results and one left. Start S2: the cut after R-JHBCWS, repository at 3cfd08d (those files sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e and 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed), actor claude-g113-3, 4 Results left. A trial replays only from a cut, never a whole build, as a Sonnet run with the prompt that the candidate own hyp prints for the Goal with that actor and no --experiments flag. It reads hyp-goal-run from the patched sandbox clone; a trial that read another copy is void. Each candidate and the base get 3 trials at each start. A candidate patch is written once per start and always states how a review is recorded and whether it counts toward the stop clause.\\nPlan, staged: the base first, which also fixes the wall-clock cap that makes a trial a hang; then the seven singles; then H-name-at-stop combined with each single that fails line 1 alone, each combination a new rival under the question. Batch the trials of a stage into one Experiment and one Result. Cost is USD priced from the usage fields of the whole trial, subagents included, at one dated Sonnet rate table, with tokens reported beside it; a hang or an unknown cost is +infinity; the median of 3 is the second smallest. Record every ratio. Known-answer cases: the base for line 1, a patch that stops after a review without recording its findings for line 4, a patch that adds two more reviews for lines 2 and 3. H-whole-candidate is expected to lose; test it anyway. H-review-owed may need a stored field; if it does, state the field and refer it, and add none.\\nMay change: the tree under investigations/goal-prompt (its .tree events and evidence) and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push that branch and open or update one PR, and never merges it. Candidate patches and scratch binaries live in scratch copies and in evidence files only.\\nRead only: skills/, internal/, docs/, plans/, every investigations/ tree except goal-prompt, ak-skill-creator u
[... 2718 more characters]
```

Result:

```text
no .tree store found; run 'hyp init'
exit=3
no .tree store found; run 'hyp init'
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/dry && HYP_ACTOR=runner-scratch hyp init --why dry; echo init=$?; ls -a; HYP_ACTOR=runner-scratch bash $TMPDIR/goal-cmd.sh; echo exit=$?; hyp status 2>&1 | head -6 | cut -c1-220; hyp check 2>&1 | tail -3",
 "description": "Init scratch store and dry-run Goal add"
}
```

Result:

```text
usage: hyp init
init=2
.
..
.git
no .tree store found; run 'hyp init'
exit=3
no .tree store found; run 'hyp init'
no .tree store found; run 'hyp init'
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

## Grill

Dry-running the Goal command in a scratch store (not the real tree) to check its format before I show it to you.

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/dry && hyp init; echo init=$?; HYP_ACTOR=runner-scratch bash $TMPDIR/goal-cmd.sh; echo exit=$?; hyp status 2>&1 | head -6 | cut -c1-220; hyp check 2>&1 | tail -3",
 "description": "Init scratch store without flags, dry-run Goal add"
}
```

Result:

```text
initialised .tree in /tmp/claude-501/dry
init=0
--body names G-ZBM2RE, which is no node here; --body names R-8JQPWG, which is no node here; --body names R-JHBCWS, which is no node here; cite an id this store holds, or write <name>:<id> for another store's
exit=2
GOAL none (hyp add goal --title T --why W)
check: 0 errors, 0 warnings
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/dry && sed -e 's/G-ZBM2RE/goal-zbm/g; s/R-8JQPWG/r-8jq/g; s/R-JHBCWS/r-jhb/g' $TMPDIR/goal-cmd.sh > $TMPDIR/goal-dry.sh && HYP_ACTOR=runner-scratch bash $TMPDIR/goal-dry.sh; echo exit=$?; hyp check | tail -2; GID=$(hyp status | grep -o 'G-[A-Z0-9]*' | head -1); hyp prompt $GID --actor r1 2>&1 | head -4 | cut -c1-160; hyp show $GID 2>&1 | grep -c \"^line\"; hyp show $GID | grep '^line' | awk '{print length($0)}'",
 "description": "Dry-run with ids neutralised; check lines and prompt"
}
```

Result:

```text
Exit code 1
refused: over limit: body has 3431 runes (max 2000)
exit=1
check: 0 errors, 0 warnings
(eval):1: failed to change group ID: operation not permitted
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat > $TMPDIR/goal-cmd.sh <<'EOF'\nhyp add goal --title \"Select the run practice that ends a budget-bound build run on a reviewed final candidate, without endless review\" --body $'Discovery Goal for issue #141 (part of #113), from runs 1 to 3 of G-ZBM2RE. The candidate is a run practice: a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never a commit. Landing the winner is a separate build Goal that the owner adds after the select.\\nS1: the store cut after R-8JQPWG, repository at b4f23b2 (SKILL.md sha256 a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab, prompt.go sha256 674dc260b4863afdec8ea2a27383ebe909b0736e25278d79aea78ecc4738d08b), actor claude-g113-1, one Result left of 5. S2: the cut after R-JHBCWS, repository at 3cfd08d (SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, prompt.go sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed), actor claude-g113-3, 4 left.\\nA trial replays only from a cut, never a whole build: a Sonnet run with the prompt the candidate own hyp prints for the Goal and that actor, with no --experiments flag, reading hyp-goal-run from the patched sandbox clone. Each candidate and the base get 3 trials at each start. Every candidate patch states how a review is recorded and whether it counts. Stages: the base, which also fixes the hang cap; the singles; H-name-at-stop with each single that fails line 1 alone. One Experiment per stage. Record every ratio. H-whole-candidate is expected to lose; test it. If H-review-owed needs a stored field, state it and add none.\\nMay change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push that branch and open or update one PR, and never merges it. Candidate patches stay in scratch copies and evidence files.\\nRead only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).\\nNever: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.\\nNever: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.\\nNever: edit a frozen recipe or another investigations/ tree.\\nNever: add an event kind or a stored field to the log.\\nNever: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.' \\\n --line \"[P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that commit as unreviewed. A trial that hangs, or whose checkpoint says all review limits are closed or the runner owes nothing while a candidate has no review, does not meet it. At S1 the
[... 1849 more characters]
```

Result:

```text
Exit code 1
refused: over limit: body has 2399 runes (max 2000)
exit=1
check: 0 errors, 0 warnings
(eval):10: failed to change group ID: operation not permitted
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR && python3 -I - <<'EOF'\nimport re,os\np=os.path.expandvars('$TMPDIR/goal-cmd.sh')\ns=open(p).read()\nfor full in [\"a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab\",\"674dc260b4863afdec8ea2a27383ebe909b0736e25278d79aea78ecc4738d08b\",\"cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e\",\"21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed\"]:\n    s=s.replace(full,full[:16]+\"…\")\nrep=[\n(\" Landing the winner is a separate build Goal that the owner adds after the select.\",\" The owner adds the build Goal that lands the winner after the select.\"),\n(\" Every candidate patch states how a review is recorded and whether it counts. Stages: the base, which also fixes the hang cap; the singles; H-name-at-stop with each single that fails line 1 alone. One Experiment per stage. Record every ratio. H-whole-candidate is expected to lose; test it. If H-review-owed needs a stored field, state it and add none.\",\" Every candidate patch states how a review is recorded and whether it counts. Stages: the base (it fixes the hang cap), the singles, then H-name-at-stop with each single that fails line 1 alone; one Experiment per stage. Record every ratio. If H-review-owed needs a stored field, state it and add none.\"),\n(\" Candidate patches stay in scratch copies and evidence files.\",\"\"),\n(\", never a whole build: a Sonnet run\",\": a Sonnet run\"),\n(\" with the prompt the candidate own hyp prints for the Goal and that actor, with no --experiments flag, reading hyp-goal-run from the patched sandbox clone.\",\" with the prompt the candidate own hyp prints for the Goal and that actor (no --experiments flag), reading hyp-goal-run from the patched sandbox clone.\"),\n]\nfor a,b in rep:\n    assert a in s,a[:30]\n    s=s.replace(a,b)\nopen(p,'w').write(s)\nEOF\ncd $TMPDIR/dry && rm -rf .tree && hyp init >/dev/null && sed -e 's/G-ZBM2RE/goal-zbm/g; s/R-8JQPWG/r-8jq/g; s/R-JHBCWS/r-jhb/g' $TMPDIR/goal-cmd.sh > $TMPDIR/goal-dry.sh && HYP_ACTOR=runner-scratch bash $TMPDIR/goal-dry.sh; echo exit=$?",
 "description": "Shorten body and retry dry-run"
}
```

Result:

```text
refused: over limit: body has 2063 runes (max 2000)
exit=1
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR && python3 -I - <<'EOF'\nimport os\np=os.path.expandvars('$TMPDIR/goal-cmd.sh')\ns=open(p).read()\nfor a,b in [(\" Each candidate and the base get 3 trials at each start.\",\"\"),(\"the run may push that branch and open or update one PR, and never merges it\",\"the run may push it and open or update one PR, and never merges it\")]:\n    assert a in s; s=s.replace(a,b)\nopen(p,'w').write(s)\nEOF\ncd $TMPDIR/dry && rm -rf .tree && hyp init >/dev/null && sed -e 's/G-ZBM2RE/goal-zbm/g; s/R-8JQPWG/r-8jq/g; s/R-JHBCWS/r-jhb/g' $TMPDIR/goal-cmd.sh > $TMPDIR/goal-dry.sh && HYP_ACTOR=runner-scratch bash $TMPDIR/goal-dry.sh; echo exit=$?; hyp check | tail -1; GID=$(hyp status | grep -o 'G-[A-Z0-9]*' | head -1); hyp prompt $GID --actor r1 2>&1 | cut -c1-100; hyp show $GID | grep '^line' | wc -l",
 "description": "Retry dry-run, check prompt and lines"
}
```

Result:

```text
Exit code 1
G-5C5RJ4 node.add
  line G-5C5RJ4#1: [P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that commit as unreviewed. A trial that hangs, or whose checkpoint says all review limits are closed or the runner owes nothing while a candidate has no review, does not meet it. At S1 the base (the files pinned in the body) meets it in at most 1 of 3 trials, so the candidate differs from it; at S2 no baseline is compared.
  line G-5C5RJ4#2: [N] Aim B. From S2, in each of 3 trials a run under the candidate launches no more than 1 fresh review before it stops. The base (the S2 files pinned in the body) meets this in at most 1 of 3 trials, so the candidate differs from it.
  line G-5C5RJ4#3: [N] Aim B. At S1 and at S2 the median cost of the candidate over 3 trials never exceeds 1.5 times the base median at the same start. Cost is USD from the usage fields of the whole trial, subagents included, at one dated Sonnet rate table. A hang or an unknown cost is +infinity; the median of 3 is the second smallest; if either median is +infinity the line is not met.
  line G-5C5RJ4#4: [N] Aim A. A run under the candidate ends with no review finding that is a case of a Goal line left neither closed, referred, fixed by a later commit nor named in its stop checkpoint.
  line G-5C5RJ4#5: [P] Aim A. A Sonnet run given only the candidate hyp-goal-run text and its printed prompt answers, in each of 3 trials, how a review is recorded and whether it counts toward the stop clause count, as the candidate text states. The base text at b4f23b2 and at 3cfd08d states neither, so the candidate differs from it.
exit=0
check: 0 errors, 0 warnings
(eval):9: bad math expression: operator expected at `C5RJ4'
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/dry && hyp prompt G-5C5RJ4 --actor claude-g146-run --branch feat/review-budget --experiments 10 2>&1 | cut -c1-170; echo exit=$?",
 "description": "Print prompt from dry-run Goal"
}
```

Result:

```text
/goal Run the hyp tree in . with the hyp-goal-run skill, as HYP_ACTOR=claude-g146-run. Its Goal is G-5C5RJ4: read it with hyp show G-5C5RJ4, read hyp status G-5C5RJ4, and
Done when the latest printed hyp status G-5C5RJ4 shows "runner owes nothing: stop and report to the owner" under OWED, or 10 Results have been recorded under claude-g146-
May change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open
Read only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering referen
Never: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.
Never: edit a frozen recipe or another investigations/ tree.
Never: add an event kind or a stored field to the log.
Never: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch feat/review-budget.
exit=0
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat $TMPDIR/goal-cmd.sh",
 "description": "Print final Goal command to show owner"
}
```

Result:

```text
hyp add goal --title "Select the run practice that ends a budget-bound build run on a reviewed final candidate, without endless review" --body $'Discovery Goal for issue #141 (part of #113), from runs 1 to 3 of G-ZBM2RE. The candidate is a run practice: a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never a commit. The owner adds the build Goal that lands the winner after the select.\nS1: the store cut after R-8JQPWG, repository at b4f23b2 (SKILL.md sha256 a6152711c2c930b2…, prompt.go sha256 674dc260b4863afd…), actor claude-g113-1, one Result left of 5. S2: the cut after R-JHBCWS, repository at 3cfd08d (SKILL.md sha256 cb15e2f2e207b5a6…, prompt.go sha256 21f9332383bc6b0e…), actor claude-g113-3, 4 left.\nA trial replays only from a cut: a Sonnet run with the prompt the candidate own hyp prints for the Goal and that actor (no --experiments flag), reading hyp-goal-run from the patched sandbox clone. Every candidate patch states how a review is recorded and whether it counts. Stages: the base (it fixes the hang cap), the singles, then H-name-at-stop with each single that fails line 1 alone; one Experiment per stage. Record every ratio. If H-review-owed needs a stored field, state it and add none.\nMay change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open or update one PR, and never merges it.\nRead only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).\nNever: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.\nNever: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.\nNever: edit a frozen recipe or another investigations/ tree.\nNever: add
[... 1943 more characters]
```

## Grill

The second reader found real defects, so I changed the design. I haven't written anything to the real store. Your approval is needed on the Goal command below.

## What the reader found, and what I checked

I checked each of these against the source and the log:

- **The base hashes in L1 were wrong for S1.** The hashes I gave you are the files at `3cfd08d`, which is S2. S1's repository is `b4f23b2`, where `SKILL.md` has sha256 `a6152711…` and `prompt.go` has `674dc260…`. Each start now pins its own commit's files.
- **The trial's budget was undefined.** The stop clause counts Results under the actor, and the cut keeps that count. `claude-g113-1` has 4 Results at S1, so there is one left. `claude-g113-3` has 1 at S2, so there are 4 left. Each trial now runs under the original run's actor with the candidate's own `hyp prompt` and no `--experiments` flag.
- **L1 could not differ from the base at S2.** With 4 left, the base run also ends on a reviewed candidate. L1's "base meets it in at most 1 of 3" now applies at S1 only. S2 is judged by L2 (aim B). I kept 5 lines, as you confirmed.
- **The cost line could pass vacuously.** If the base's median is +∞, any candidate would pass. L3 now fails when either median is +∞.
- **The review-owed rival may need a stored field.** The body says: if H-review-owed needs one, the run states it and adds none.

The reader checked that `--as-of R-8JQPWG` replays correctly with a third Goal in the store. It did not run `hyp sandbox create` with a cut, so the first Experiment must prove the replay builds before any trial runs.

**Two caveats:**
- `hyp` caps a Goal body at 2000 runes. To fit, the four baseline hashes in the body are 16-hex prefixes with "…", the same abbreviation G-ZBM2RE's line 6 used. The two ak-context-engineering hashes are shortened to 8 hex digits and the `...` placeholder.
- The staging, the hang cap and the known-answer cases are not in the body. They go in each Experiment's verdict rule, which the run writes.

## Your approval

I dry-ran this exact command in a scratch store, with the three ids it cites replaced by placeholders. It exited 0 with `check: 0 errors, 0 warnings`, and `hyp prompt` printed the constraints block line for line. Run it from `investigations/goal-prompt`, with the actor `owner-standin`:

```bash
HYP_ACTOR=owner-standin hyp add goal --title "Select the run practice that ends a budget-bound build run on a reviewed final candidate, without endless review" --body $'Discovery Goal for issue #141 (part of #113), from runs 1 to 3 of G-ZBM2RE. The candidate is a run practice: a scratch patch to skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go or a scratch hyp binary, never a commit. The owner adds the build Goal that lands the winner after the select.\nS1: the store cut after R-8JQPWG, repository at b4f23b2 (SKILL.md sha256 a6152711c2c930b2…, prompt.go sha256 674dc260b4863afd…), actor claude-g113-1, one Result left of 5. S2: the cut after R-JHBCWS, repository at 3cfd08d (SKILL.md sha256 cb15e2f2e207b5a6…, prompt.go sha256 21f9332383bc6b0e…), actor claude-g113-3, 4 left.\nA trial replays only from a cut: a Sonnet run with the prompt the candidate own hyp prints for the Goal and that actor (no --experiments flag), reading hyp-goal-run from the patched sandbox clone. Every candidate patch states how a review is recorded and whether it counts. Stages: the base (it fixes the hang cap), the singles, then H-name-at-stop with each single that fails line 1 alone; one Experiment per stage. Record every ratio. If H-review-owed needs a stored field, state it and add none.\nMay change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open or update one PR, and never merges it.\nRead only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).\nNever: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.\nNever: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.\nNever: edit a frozen recipe or another investigations/ tree.\nNever: add an event kind or a stored field to the log.\nNever: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.' \
 --line "[P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that commit as unreviewed. A trial that hangs, or whose checkpoint says all review limits are closed or the runner owes nothing while a candidate has no review, does not meet it. At S1 the base (the files pinned in the body) meets it in at most 1 of 3 trials, so the candidate differs from it; at S2 no baseline is compared." \
 --line "[N] Aim B. From S2, in each of 3 trials a run under the candidate launches no more than 1 fresh review before it stops. The base (the S2 files pinned in the body) meets this in at most 1 of 3 trials, so the candidate differs from it." \
 --line "[N] Aim B. At S1 and at S2 the median cost of the candidate over 3 trials never exceeds 1.5 times the base median at the same start. Cost is USD from the usage fields of the whole trial, subagents included, at one dated Sonnet rate table. A hang or an unknown cost is +infinity; the median of 3 is the second smallest; if either median is +infinity the line is not met." \
 --line "[N] Aim A. A run under the candidate ends with no review finding that is a case of a Goal line left neither closed, referred, fixed by a later commit nor named in its stop checkpoint." \
 --line "[P] Aim A. A Sonnet run given only the candidate hyp-goal-run text and its printed prompt answers, in each of 3 trials, how a review is recorded and whether it counts toward the stop clause count, as the candidate text states. The base text at b4f23b2 and at 3cfd08d states neither, so the candidate differs from it." \
 --why "Owner, 2026-11-12: the discovery Goal for issue #141; lines and their aims confirmed in the grill session."
```

**Do you approve this command, as written or with your edits?** I will not run it until you say so.

## What I run after your approval

These are runner writes under a fresh actor, `claude-grill-141`:
1. `hyp add dq --parent <G-id>` for Q1, with the S1 and S2 scenarios in its body.
2. Seven `hyp add hyp --parent <Q1-id>`: H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate (body notes your expectation that it loses) and H-review-owed.
3. `hyp link <G-id> external_ref` to issue #141.
4. `hyp check`, which must print `check: 0 errors`, then the end-of-grill checkpoint.
5. `hyp prompt <G-id> --actor <runner-actor> --branch feat/review-budget --experiments 10`, and I give you the printed prompt unchanged.

## Owner, message 5

Approved, run the command as written. Use `claude-grill-141` for the runner writes after it.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt && HYP_ACTOR=owner-standin bash $TMPDIR/goal-cmd.sh 2>&1 | cut -c1-200; echo exit=${pipestatus[1]}; hyp status 2>&1 | head -6 | cut -c1-160",
 "description": "Run approved owner Goal add in the real store"
}
```

Result:

```text
refused: over limit: body has 2004 runes (max 2000)
exit=1
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner ow
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothin
AWAITING OWNER
```

## Grill

The approved command was refused, and nothing was written. `hyp` rejected the body as 2004 runes against a limit of 2000. My dry run used short placeholder ids in place of `G-ZBM2RE`, `R-8JQPWG` and `R-JHBCWS`, so it undercounted by 4 runes. Any edit changes the command you approved, so I haven't run a changed version.

My proposed fix is one phrase in the `May change:` line. Replace "open or update one PR" with "open one PR", which saves 10 runes. It slightly narrows what the run may do: it could no longer update the PR once opened. If you'd rather keep that, say so and I'll cut a different 4 or more runes. For example, I could drop "(it fixes the hang cap)" instead.

Everything else stays as approved, including the five lines and the `--why`. The edited line would read:

`May change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open one PR, and never merges it.`

Do you approve the command with that one change? Once you do, I'll run it. Then I'll run the `claude-grill-141` writes: Q1 and its seven rivals, the link to #141, `hyp check`, the end-of-grill checkpoint, and `hyp prompt`.

## Owner, message 6

Keep the ability to update the PR. Cut "(it fixes the hang cap)" instead, and any other filler you need, so the body fits 2000 runes. Run it with that edit under `owner-standin`, and I approve it.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt && sed -i '' 's/ (it fixes the hang cap)//' $TMPDIR/goal-cmd.sh && grep -c \"hang cap\" $TMPDIR/goal-cmd.sh; HYP_ACTOR=owner-standin bash $TMPDIR/goal-cmd.sh 2>&1 | cut -c1-200; hyp status 2>&1 | head -5 | cut -c1-140",
 "description": "Run approved edited Goal add as owner-standin"
}
```

Result:

```text
0
G-BW3XJK node.add
  line G-BW3XJK#1: [P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate commit, or with a stop checkpoint that names that 
  line G-BW3XJK#2: [N] Aim B. From S2, in each of 3 trials a run under the candidate launches no more than 1 fresh review before it stops. The base (the S2 files pinned in the body) meets this in at m
  line G-BW3XJK#3: [N] Aim B. At S1 and at S2 the median cost of the candidate over 3 trials never exceeds 1.5 times the base median at the same start. Cost is USD from the usage fields of the whole t
  line G-BW3XJK#4: [N] Aim A. A run under the candidate ends with no review finding that is a case of a Goal line left neither closed, referred, fixed by a later commit nor named in its stop checkpoin
  line G-BW3XJK#5: [P] Aim A. A Sonnet run given only the candidate hyp-goal-run text and its printed prompt answers, in each of 3 trials, how a review is recorded and whether it counts toward the sto
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0,
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
  G-BW3XJK Select the run practice that ends a budget-bound build run … — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn
AWAITING OWNER
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 G=G-BW3XJK\nhyp add dq --parent $G --alias q-practice --title \"What run practice ends a budget-bound build run on a reviewed final candidate, without review going on over small changes?\" --body \"Separating scenarios: S1 (cut after R-8JQPWG: one Result left, a fix commit pending) and S2 (cut after R-JHBCWS: reviews that found no defect still led to commits). Replay each from the cut with 3 Sonnet trials per candidate and for the base, and compare how each run ends, how many fresh reviews it launches and what it costs. Each rival is a scratch patch; the Goal body pins the base.\" --why \"The question of issue #141, held as rival Hypotheses for the run to test.\" 2>&1 | tail -3\nQ=$(hyp status $G | grep -o 'Q-[A-Z0-9]*' | head -1); echo Q=$Q\nadd(){ hyp add hyp --parent $Q --alias \"$1\" --title \"$2\" --body \"$3\" --why \"A rival of issue #141 named by the owner, to be tested from the cuts.\" 2>&1 | head -2; }\nadd h-one-experiment \"H-one-experiment: a candidate's review is part of its gate Experiment\" \"One Experiment and one Result carry the gate's verdicts and the review's limits, so a candidate costs one Result again.\"\nadd h-name-at-stop \"H-name-at-stop: the stop checkpoint names each candidate without a review\" \"Keep the practice of a review as its own Experiment; the stop checkpoint names every candidate with no review, so the gap is visible. Also tested combined with each single that fails line 1 alone.\"\nadd h-reserve-last \"H-reserve-last: with one Result left the run reviews the current candidate instead of committing another fix\" \"The run stops with the review's findings recorded as limits, not with a new commit.\"\nadd h-raise-budget \"H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate\" \"A scratch change to prompt.go's default (for example 7). Owner's comment on run 2 argues against it: each review so far led to a new commit.\"\nadd h-probe-closes \"H-probe-closes: a case the run's own probe ran and passed on the candidate is closed in the review's Result, not recorded as a limit\" \"A limit stays for a case no check ran or one the candidate fails. Run 3's N1 and P1 would close on 3cfd08d with no further commit.\"\nadd h-whole-candidate \"H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base\" \"Owner expects it to lose on cost (one cumulative review is large); it is tested anyway.\"\nadd h-review-owed \"H-review-owed: hyp status OWED lists a review for a candidate that has none\" \"A scratch hyp binary derives the owed review from the log. A gate Experiment and a review look alike (--recipe none --rubric reading); if the log cannot tell them apart without a new stored field, the run states the field and refers it, and adds none.\"\nhyp link $G external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this Goal serves.\" 2>&1 | tail -1\nhyp check | tail -2; hyp status $G | head 
[... 91 more characters]
```

Result:

```text
refused: over limit: title has 122 runes (max 120)
Q=
add hyp: wrong number of arguments
add hyp: wrong number of arguments
add hyp: wrong number of arguments
add hyp: wrong number of arguments
add hyp: wrong number of arguments
add hyp: wrong number of arguments
add hyp: wrong number of arguments
G-BW3XJK edge.add
warn: uncited R-BSWFXG: 2aeb0abe is in the observation but in no cited file
check: 0 errors, 9 warnings
GOAL G-BW3XJK Select the run practice that ends a budget-bound build run … — in_progress (answered 0, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-BW3XJK#1 no candidate — [P] Aim A. At S1 and at S2, in each of 3 trials a Sonnet run under the candidate ends with a review Result for its final candidate co…
  G-BW3XJK#2 no candidate — [N] Aim B. From S2, in each of 3 trials a run under the candidate launches no more than 1 fresh review before it stops. The base (the…
  G-BW3XJK#3 no candidate — [N] Aim B. At S1 and at S2 the median cost of the candidate over 3 trials never exceeds 1.5 times the base median at the same start. …
  G-BW3XJK#4 no candidate — [N] Aim A. A run under the candidate ends with no review finding that is a case of a Goal line left neither closed, referred, fixed b…
  G-BW3XJK#5 no candidate — [P] Aim A. A Sonnet run given only the candidate hyp-goal-run text and its printed prompt answers, in each of 3 trials, how a review …
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
RECENT
  2026-11-12 14:17 claude-grill-141 edge.add G-BW3XJK — The issue this Goal serves.
CHECKPOINT none
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 G=G-BW3XJK\nhyp add dq --parent $G --alias q-practice --title \"What run practice ends a budget-bound build run on a reviewed final candidate, without endless review?\" --body \"Separating scenarios: S1 (cut after R-8JQPWG: one Result left, a fix commit pending) and S2 (cut after R-JHBCWS: reviews that found no defect still led to commits). Replay each from the cut with 3 Sonnet trials per candidate and for the base, and compare how each run ends, how many fresh reviews it launches and what it costs. Each rival is a scratch patch; the Goal body pins the base.\" --why \"The question of issue #141, held as rival Hypotheses for the run to test.\" 2>&1 | tail -2\nQ=$(hyp status $G | grep -o 'Q-[A-Z0-9]\\{6\\}' | head -1); echo Q=$Q",
 "description": "Add DesignQuestion with shorter title"
}
```

Result:

```text
Q-QJXJT3 node.add
Q=Q-QJXJT3
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 G=G-BW3XJK Q=Q-QJXJT3\nadd(){ hyp add hyp --parent $Q --alias \"$1\" --title \"$2\" --body \"$3\" --why \"A rival of issue #141 named by the owner, to be tested from the cuts.\" 2>&1 | head -1; }\nadd h-one-experiment \"H-one-experiment: a candidate's review is part of its gate Experiment\" \"One Experiment and one Result carry the gate's verdicts and the review's limits, so a candidate costs one Result again.\"\nadd h-name-at-stop \"H-name-at-stop: the stop checkpoint names each candidate without a review\" \"Keep a review as its own Experiment; the stop checkpoint names every candidate with no review, so the gap is visible. Also tested combined with each single that fails line 1 alone.\"\nadd h-reserve-last \"H-reserve-last: with one Result left the run reviews the current candidate, not another fix\" \"The run stops with the review's findings recorded as limits, not with a new commit.\"\nadd h-raise-budget \"H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate\" \"A scratch change to prompt.go's default (for example 7). Run 2 argues against it: each review so far led to a new commit.\"\nadd h-probe-closes \"H-probe-closes: a case the run's own probe passed on the candidate is closed in the review's Result, not a limit\" \"A limit stays for a case no check ran or one the candidate fails. Run 3's N1 and P1 would close on 3cfd08d with no further commit.\"\nadd h-whole-candidate \"H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base\" \"The owner expects it to lose on cost, since one cumulative review is large; it is tested anyway.\"\nadd h-review-owed \"H-review-owed: hyp status OWED lists a review for a candidate that has none\" \"A scratch hyp binary derives the owed review from the log. A gate and a review look alike (--recipe none --rubric reading); if the log cannot tell them apart without a new stored field, the run states the field and refers it, and adds none.\"\nhyp check | tail -1; hyp status $G | sed -n 8,22p | cut -c1-160",
 "description": "Add the seven rival Hypotheses"
}
```

Result:

```text
H-6KFK19 node.add
H-F17VMA node.add
H-B8H98M node.add
H-QB7E71 node.add
H-MSQPSP node.add
H-TT3AAP node.add
H-Y7QCEM node.add
check: 0 errors, 9 warnings
FRONTIER
  Q-QJXJT3 open/untested What run practice ends a budget-bound build run on a review…
    H-6KFK19 (h-one-experiment) proposed H-one-experiment: a candidate's review is part of its gate …
    H-F17VMA (h-name-at-stop) proposed H-name-at-stop: the stop checkpoint names each candidate wi…
    H-B8H98M (h-reserve-last) proposed H-reserve-last: with one Result left the run reviews the cu…
    H-QB7E71 (h-raise-budget) proposed H-raise-budget: the default of --experiments rises to fit a…
    H-MSQPSP (h-probe-closes) proposed H-probe-closes: a case the run's own probe passed on the ca…
    H-TT3AAP (h-whole-candidate) proposed H-whole-candidate: the review before the select reads the c…
    H-Y7QCEM (h-review-owed) proposed H-review-owed: hyp status OWED lists a review for a candida…
STALE
OWED
  runner: Q-QJXJT3 test H-6KFK19 (h-one-experiment)
  runner: Q-QJXJT3 test H-F17VMA (h-name-at-stop)
  runner: Q-QJXJT3 test H-B8H98M (h-reserve-last)
  runner: Q-QJXJT3 test H-QB7E71 (h-raise-budget)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 G=G-BW3XJK\nhyp check | grep -c \"BW3XJK\\|QJXJT3\"; \nhyp checkpoint --goal $G --summary \"Grill done: 5 lines written by owner-standin, 7 rivals under Q-QJXJT3, issue #141 linked. The run tests from the cuts S1 and S2, staged: base, singles, then H-name-at-stop combinations. Out-of-Goal: landing the winner is a build Goal the owner adds after the select; a derived review mark may need a stored field and is referred, not added.\" --why \"end of grill\" 2>&1 | tail -1\nhyp check | head -1; hyp check | tail -1; hyp prompt $G --actor claude-g141-run-1 --branch feat/review-budget --experiments 10; echo exit=$?; git status --short | head",
 "description": "Checkpoint, check and print the /goal prompt"
}
```

Result:

```text
0
G-BW3XJK checkpoint
warn: uncontained R-KYKWH9 (script outside recipes/: .tree/evidence/E-H98Z3B/probe-k1.sh)
check: 0 errors, 9 warnings
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-run-1. Its Goal is G-BW3XJK: read it with hyp show G-BW3XJK, read hyp status G-BW3XJK, and name --goal G-BW3XJK on every hyp checkpoint.
Done when the latest printed hyp status G-BW3XJK shows "runner owes nothing: stop and report to the owner" under OWED, or 10 Results have been recorded under claude-g141-run-1, or you cannot take the next step honestly and say why.
May change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).
Never: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.
Never: edit a frozen recipe or another investigations/ tree.
Never: add an event kind or a stored field to the log.
Never: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch feat/review-budget.
exit=0
?? .tree/events/claude-grill-141.jsonl
?? .tree/events/owner-standin.jsonl
```

## Grill

The grill for #141 is finished. The new Goal G-BW3XJK, Q1 with its seven rivals and the end-of-grill checkpoint are written. `hyp check` prints `check: 0 errors`, with 9 warnings, none of them on the new nodes.

## What the tree holds

- **Goal G-BW3XJK**, written by `owner-standin`, with the five lines you confirmed, each tagged with its aim. I cut "(it fixes the hang cap)" from the body, as you said, which brought it under the 2000-rune limit. The body says to record every ratio.
- **Q-QJXJT3** (alias `q-practice`), whose body holds the S1 and S2 scenarios.
- **Seven rival Hypotheses under Q-QJXJT3**:

  | Id | Rival |
  |---|---|
  | H-6KFK19 | H-one-experiment |
  | H-F17VMA | H-name-at-stop |
  | H-B8H98M | H-reserve-last |
  | H-QB7E71 | H-raise-budget |
  | H-MSQPSP | H-probe-closes |
  | H-TT3AAP | H-whole-candidate (your expectation that it loses is in its body) |
  | H-Y7QCEM | H-review-owed |

- **The link** to https://github.com/dat9uy/hyp/issues/141 as `external_ref`, and the end-of-grill checkpoint under `claude-grill-141`.

## Not done

- **Not committed.** `.tree/events/owner-standin.jsonl` and `.tree/events/claude-grill-141.jsonl` are untracked, and I did not commit them. The skill says to commit `.tree/events/` and `.tree/evidence/`, but you didn't ask me to commit. Run the commit when you want it. The prompt below names `feat/review-budget`, and `hyp prompt` needs the Goal's add in the log the run starts from.
- **Landing the winner.** The build Goal that lands the winner is yours to add after the select.

## The `/goal` prompt, as printed

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-run-1. Its Goal is G-BW3XJK: read it with hyp show G-BW3XJK, read hyp status G-BW3XJK, and name --goal G-BW3XJK on every hyp checkpoint.
Done when the latest printed hyp status G-BW3XJK shows "runner owes nothing: stop and report to the owner" under OWED, or 10 Results have been recorded under claude-g141-run-1, or you cannot take the next step honestly and say why.
May change: the tree under investigations/goal-prompt and one new recipe version for the replay, in commits on the branch feat/review-budget; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, docs/, plans/, every other investigations/ tree, ak-skill-creator under .claude/skills pinned by sha256, and ak-context-engineering references/evaluation.md (sha256 fdd4a34f ...) and scripts/benchmark_metrics.py (sha256 2a98cf72 ...).
Never: commit a candidate patch to skills/ or internal/, commit to main or merge a PR.
Never: replace ~/.local/bin/hyp or any installed hyp; compile every candidate to a separate path.
Never: edit a frozen recipe or another investigations/ tree.
Never: add an event kind or a stored field to the log.
Never: write under an owner actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch feat/review-budget.
```

## Known risks for the run

- **Replay is untested.** No one has run `hyp sandbox create` with a store cut yet. The first Experiment must prove that the replay builds before any trial runs.
- **L3 may fail every reviewing candidate.** At S1 the base stops without a review, so a candidate that adds one may exceed 1.5 times the base's cost. You chose to keep the rule, so the run records every ratio.
- **The run's own budget is tight.** 10 Results are enough only if each stage's trials are batched into one Experiment and one Result, which the body says.
