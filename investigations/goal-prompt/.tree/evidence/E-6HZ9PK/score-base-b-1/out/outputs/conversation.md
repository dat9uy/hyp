# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-b-1/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll read the issue and check the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/tmp/issue-141.md"
}
```

Result:

```text
1	# Issue #141: Run practice: the last candidate of a budget-bound run goes unreviewed
2	
3	Opened by dat9uy at 2026-11-12T02:21:06Z. Text as of 2026-11-12T03:45:00Z.
4	
5	## Event
6	
7	On 2026-11-12, run 1 of `investigations/goal-prompt` (G-ZBM2RE, PR #140) stopped with its last candidate, H-YXSJEM at `5c5902d`, never reviewed. The two earlier candidates were reviewed, and each review found real defects (10 findings on `10566d4`, 8 on `b4f23b2`). The stop checkpoint says "all review limits closed" and does not name the missing review.
8	
9	The chain, from the run's transcript and the two trees:
10	
11	1. A review finding that is a case of a line becomes a limit, and `--limit` exists only on `hyp record`. So the runner recorded each review as its own Experiment: E-C3GGGJ reviewed `10566d4` and E-7QQ2DP reviewed `b4f23b2`.
12	2. Each candidate therefore cost two Experiments, a gate and a review. The run spent E-HQWTA0, E-C3GGGJ, E-Z5BTZR, E-7QQ2DP and E-NZR3GV (the gate of `5c5902d`), and the stop clause "5 Experiments have been recorded" ended it.
13	3. The default of 5 was set from G-D1CH1R's runs, where every Experiment gates a commit and review findings went straight into fix commits (one Experiment per candidate; run 5 used 4).
14	4. At 17:21:37 the runner wrote "E5 is the last Experiment the budget allows, so this fix commit must close everything." `skills/hyp-goal-run/SKILL.md` asks for "one review of each candidate commit's diff" before the stop. The prompt's stop clause and the skill conflict, and neither says which yields or that the gap is named at the stop.
15	
16	## Scope
17	
18	Core: the run practice of a Goal whose candidate is a commit (review, budget and stop report) in `hyp-goal-run`, and the stop clause `hyp prompt` prints. No self-review recipe is involved.
19	
20	## Rivals
21	
22	A tree holds this question; each option below is a rival Hypothesis, not a decision. The scenario that separates them: replay the G-ZBM2RE cut after R-8JQPWG (one Experi
[... 6320 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "hyp status 2>&1 | head -60; ls; ls docs | head -50"
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
 "command": "ls investigations investigations/goal-prompt 2>&1; cat skills/hyp-goal-run/SKILL.md; grep -rn \"experiments\\|Experiments\" internal/prompt* cmd/hyp/*prompt* 2>/dev/null | head -30"
}
```

Result:

```text
investigations:
checkpoint-store
checkpoint-store-design
goal-prompt
id-citation
recipe-base

investigations/goal-prompt:
evals
---
name: hyp-goal-run
description: "Run a hyp tree under an owner's Goal and met rule as an unattended /goal, for a discovery Goal or a build Goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report the owner can read from hyp status. Use when a /goal prompt names this skill, such as the one hyp prompt prints."
---

The `/goal` prompt names the tree, your actor, the Goal, the stop and the Goal body's constraints; `hyp prompt <goal-id> --actor <actor>` prints it from the log. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.

## Start

- Read the Goal first with `hyp show <goal-id>`, including its checkpoints, because `hyp status` does not print its met rule.
- Everything you need is in the tree. Do not read session logs, reports or memory about earlier runs.
- The Goal's body states the run's constraints under three headings: `May change:` (what you may write, and how), `Read only:` (what you may read but never change) and `Never:` (acts you never take). Obey them. A `/goal` prompt may narrow them, never widen them. If a step needs something they forbid, refer it with a reason or record a reading, and stop.
- A store may hold many Goals. Then `hyp status` prints only a GOALS table and one Awaiting owner list, so run `hyp status <goal-id>` wherever this skill says `hyp status`: OWED, LINES, the stop line and the last checkpoint are per Goal. Name the Goal with `--goal <goal-id>` on `hyp checkpoint`.
- Print `hyp status` at the end of every turn.

## Work

Act on what OWED in `hyp status` says the runner owes, and on nothing else. Plan only the next Exper
[... 12201 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat docs/core-and-self-review.md | head -80; grep -n -i \"verifier\\|review\" docs/glossary.md | head -20"
}
```

Result:

```text
# Map: core and self-review

This repository holds hyp, and a few recipes and trees that test it. Records made before the public release stay in the private development history. This page says which concepts and files belong to which scope, so that an issue, a plan or a change can say which scope it serves. The two words are defined in [glossary.md](glossary.md), "Scopes of this repository".

## The test

Core and self-review are scopes. An item's scope is the concept it belongs to, not who uses it or what tested it.

- **Core** is every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A user writes recipes, runs sandboxes and records trees for their own code, so recipe, sandbox, store cut, runner and instrument are core concepts, even where only self-review uses them today.
- **Self-review** is the recipes this repository keeps to test hyp, whatever repository the runner works on, plus the inputs only they use. Nothing else is self-review.

The repository a runner works on is its investigated repo, a core concept (spec §1). Here it was often hyp itself, and often a real outside project: session-replay's snapshots were sessions of a runner working on another application. Which repository it is does not change the scope.

Investigating hyp is an activity, not a scope. A tree takes its place by its question:

- A tree whose question is about hyp takes that question's scope. Most trees here ask whether a core concept works, so they are core-scope work, carried out with self-review recipes.
- A tree whose question is about another project is a self-review input: a run this repository keeps as evidence about hyp, not for its answer.

Three rules keep the boundary honest:

1. **A concept stays core when self-review is what tests it.** The runner was measured by session-replay, and it is still a core concept.
2. **A self-review recipe does not make its kind self-review.** session-replay was a recipe; recipes are co
[... 10907 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "find investigations -maxdepth 2 | head -40; git log --oneline | head -5; git branch -a; grep -rln \"Results recorded under\\|recorded under the actor\" --include=*.go . | head; which hyp; hyp version 2>&1 | head -2"
}
```

Result:

```text
investigations
investigations/recipe-base
investigations/recipe-base/.tree
investigations/recipe-base/arms.txt
investigations/checkpoint-store
investigations/checkpoint-store/.tree
investigations/id-citation
investigations/id-citation/.tree
investigations/id-citation/arms
investigations/goal-prompt
investigations/goal-prompt/evals
investigations/goal-prompt/.tree
investigations/checkpoint-store-design
investigations/checkpoint-store-design/.tree
9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
04142f3 fix: test that a bad block line in a store outside every repository exits 1
d2d5d14 fix: test that an invalid block in a store outside every repository exits 1
* work
(eval):1: no matches found: --include=*.go
/Users/datguy/.hyp-sandbox/g146-base-b-1/bin/hyp
hyp 0.1.0 (events v1)
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && export HYP_ACTOR=grill-141; hyp status 2>&1 | head -50; ls evals | head; cd ../..; grep -rn \"recorded under\" --include='*.go' . | head -5"
}
```

Result:

```text
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-11 16:47 claude-grill-113
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-12 03:28 claude-g113-4
AWAITING OWNER
goal-prompt-evals-2.json
goal-prompt-evals.json
./internal/recipe/recipe.go:451:// copies hash differently are left out, so Results recorded under them read
./internal/cli/prompt_test.go:55:		"or 5 Results have been recorded under claude-run-1, or you cannot take the next step honestly and say why.\n" +
./internal/cli/prompt_test.go:63:	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 Results have been recorded under claude-run-1,")
./internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
./internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && hyp show G-ZBM2RE 2>&1 | head -70; hyp status G-ZBM2RE | head -30; ls .tree .tree/events; head -c 1500 evals/goal-prompt-evals-2.json"
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
[... 16401 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "for i in R-8JQPWG R-JHBCWS R-3YJ7T2 R-28A8NB R-KYKWH9 E-C3GGGJ E-GE085A; do echo \"== $i\"; hyp show $i 2>&1 | head -6; done; hyp add dq --help 2>&1 | head -20; hyp show G-GZR6C6 2>&1 | head -20; ls ../../recipes ../../rubrics 2>&1 | head"
}
```

Result:

```text
== R-8JQPWG
Result R-8JQPWG
path: G-ZBM2RE > Q-Z93VJ5 > H-3M3QAS > E-7QQ2DP > R-8JQPWG
status: recorded
observation: The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and flags, exit 3 through a symlinked working directory and exit 0 with 'in inv/t' through the physical path, where G-ZBM2RE#1 prints the same bytes on every call. It confirms G2 (a CRLF blank line inside the block is refused as "\r") and G3 (a no-break space in --branch exits 2). G3, G4, G5, G6 and G7 are text or test defects; G4 also names a rule conflict outside the lines.
verdicts: H-3M3QAS=contradicts
evidence: .tree/evidence/E-7QQ2DP/review.md
== R-JHBCWS
Result R-JHBCWS
path: G-ZBM2RE > Q-Z93VJ5 > H-FQ103R > E-TZTG0G > R-JHBCWS
status: recorded
observation: The fresh review of 3cfd08d reports the new CRLF test red under three mutations of prompt.go (strip \r, normalise CRLF, refuse a trailing \r) and the new no-block test red when the block check moves after the repository lookup; the new spec sentence is true of the code. N1: a mutation that checks only that a block exists before the lookup, and its validity after, passes every TestPrompt test. The probe with the 3cfd08d binary (785dd8d6a43871e7) runs N1: an invalid block in a store outside every repository exits 1 with empty stdout. The outcome is right; no go test runs it. N2 names no line.
verdicts: H-FQ103R=supports
evidence: .tree/evidence/E-TZTG0G/review.md
== R-3YJ7T2
Result R-3YJ7T2
path: G-ZBM2RE > Q-Z93VJ5 > H-YTD2JG > E-2A0909 > R-3YJ7T2
status: recorded
observation: The fresh review of 2284172 reports the new test red when EvalSymlinks is removed, so it runs R-KYKWH9#1, and M1-M3. The probe with the 2284172 binary (c5858435f689b848) runs M2: a CRLF body with no blank line exits 0 and keeps the \r on all three block lines; and M3: a store outside every repository with no block exits 1 with empty stdout. Both outcomes are right; no go test runs either. M1 names no line; 
[... 4962 more characters]
```

## Grill

I've read issue #141 and checked the store in `investigations/goal-prompt`. Round 1 is below. I've written nothing to the store.

## What I found

- The store `investigations/goal-prompt` already holds the cuts the issue names. `R-8JQPWG`, `R-JHBCWS`, `R-3YJ7T2` and `E-GE085A` all exist. G-GZR6C6 (design) and G-ZBM2RE (build) are in it, and both are complete. A new Goal for this issue goes **in that store**, because the issue is "Part of #113". I did not need to ask about this.
- `hyp prompt` counts **Results**, not Experiments, in its stop clause. `internal/cli/prompt.go:167` prints "N Results have been recorded under A". The flag is still named `--experiments`. The two events in the issue differ from the stop text only in this word.
- No public recipe can replay a runner session on a store cut. Under `recipes/` only the id-citation, authoring and recipe-base recipes remain. The session-replay recipe is private. So "a run on the scenario" needs an instrument that does not exist yet.
- The derivation guard rules out one design. "Candidate X is reviewed" cannot be derived from the log, because a review is a convention (`--recipe none --rubric reading`) and not a typed event. No rival may add a `reviewed` field or event. `H-name-at-stop` is a checkpoint sentence, and `hyp check` cannot enforce it. A rival that needs a stored marker fails the guard.
- This run's own candidates are claims, not commits, so the gap in #141 cannot bite it. The run does not apply the review practice under test.

## Round 1 (owner questions)

**1. One Goal now, or two?**
- Today: the issue asks for two things. First, a tree that selects a rival by running the scenario. Second, a change to `hyp-goal-run` and the stop clause.
- Problem: the build Goal's lines depend on which rival wins, so they cannot be written yet. G-GZR6C6 and G-ZBM2RE were settled in the same session, which is not possible here.
- A. Write one discovery Goal now. Its candidates are the rivals. A later grill, after the select, writes the build Goal. This follows "a design record and its build are two Goals".
- B. Write the build Goal now. Its Hypotheses are one commit per rival.
- **Recommendation: A.** A build Goal's lines need a settled design, and B hides the choice among rivals inside commit candidates.

**2. What counts as "a run on the scenario"?**
- Today: no replay recipe is public. Replaying means that a fresh runner, given a copy of the store cut and one rival's text, makes the next moves, and a rubric reads where it stops.
- Example: take the cut after `R-8JQPWG`. The runner has one Result left and a fix commit pending. Under `H-reserve-last` it should review `H-3M3QAS` and stop. Under the status quo it commits the fix and ends unreviewed.
- A. The Goal lets the run build a replay recipe (`recipes/review-stop-arms@1`). It also builds a rubric that grades where the replayed run stopped. One live runner replays each rival on each of three cuts: run 1 after `R-8JQPWG`, run 2 after `R-28A8NB`, run 3 after `R-JHBCWS`.
- B. A paper scenario. A subagent is told "you are in this state, with this text; what is your next act?" and answers once. This is cheap, but it is a stated plan and not a run.
- **Recommendation: A.** The issue says "from a run", and the owner's comments report real divergence between runs. Does the owner accept new files under `recipes/` as something the run may change?

**3. The lines** (positive and negative space). I propose five. The cases are the scenarios, and a check on each replay can decide them. Please edit or strike any.
- `[P]` In each replayed cut, a run that reaches its stop clause leaves its final candidate commit either with a recorded review Result that cites the diff it read, or named as unreviewed in the stop checkpoint's summary.
- `[N]` A replayed run never stops with its final candidate unreviewed and not named in the stop checkpoint. This is the silent gap of run 1.
- `[N]` A replayed run never closes a case as passing when its probe failed on the candidate. This is the safety line for `H-probe-closes`.
- `[P]` A review's finding that is a case of a line ends as a limit, a close or a fix commit, and `hyp status` lists each. No finding lives only in the review file.
- `[N]` A replayed run never exceeds its stop clause's count to finish a review. The budget holds under every rival.

Do you want any of these merged, dropped or added? For example, a `[P]` line that a review's Result states whether it counts against the budget (the owner's R14 comment).

**4. Constraints block for the Goal body.** My draft:
- `May change:` the `investigations/goal-prompt` tree, a new `recipes/review-stop-arms@1` bundle and the rubric it pins, on commits to one local branch. A PR is opened or updated and never merged.
- `Read only:` `skills/`, `internal/`, `docs/`, the other `investigations/` trees, and the event logs of `claude-g113-1` to `-4`.
- `Never:` edit `skills/` or `internal/`; replace an installed hyp; write under the owner's actor; merge.
- Is this right? In particular, do you want the run barred from editing `skills/hyp-goal-run/SKILL.md` until the build Goal? I recommend yes, because skill text reaches replayed runs.

**5. The run's budget in the `/goal` prompt.**
- Today: `--experiments` defaults to 5.
- Plan: about three DesignQuestion Experiments (one per question, a verdict per rival), plus one for building the instrument and its known-answer check, plus reruns.
- **Recommendation: `--experiments 8`**, with no extra review round. The run has no commit candidates, so the gap does not apply to it. Is that acceptable?

## Draft DesignQuestions (the routing table comes next round)
Each separates its rivals by a replayed scenario. The status quo is a rival only where it is the baseline that must go red. I add it myself.

| DesignQuestion | Rivals | Separating scenario |
|---|---|---|
| Q-budget: how does a budget-bound run end so its last candidate is reviewed or named? | status quo; `H-one-experiment`; `H-name-at-stop`; `H-reserve-last`; `H-raise-budget`; combinations the run tests | cut after `R-8JQPWG`, and the cut after `R-3YJ7T2` |
| Q-sort: what happens to a review case the run's own probe passes? | status quo (limit, then fix commit); `H-probe-closes` | cut after `R-JHBCWS` |
| Q-read: what does the review before the select read? | latest increment (status quo); `H-whole-candidate` | the final candidate `04142f3`: does a review of the cumulative diff from the Goal's base find a case that three increment reviews missed? |

Two facts for you to correct if wrong:
- I treat the owner's R14 rule, "a case whose check already passes closes as a limit, not with a new commit", as a variant of `H-probe-closes` under Q-sort. It is not a separate question.
- Whether `hyp-goal-run` should state how a review is recorded and whether it counts against the budget is a line of the build Goal. The grill won't hold it here.

## What's next
After your answers I'll run the second reader on the rules. It will walk them through a scenario the grill did not draw: two Goals in one store, or a merged log from two actors. Then I'll show the routing table. I'll write nothing until you confirm the table, and every owner write waits for you to approve its exact command. What is your actor name (for example `owner-dat`)?

## Owner, message 2

Actor name: owner-standin.

1. A. One discovery Goal in the `goal-prompt` store.
2. A, with new files under `recipes/` allowed. Two changes:
   - Run 3 trials per rival per cut, not one, on Sonnet.
   - Don't replay a whole build from its start for each trial. Starting from the cuts is fine.
   
   Could the ak-context-engineering skill help with how we evaluate this? Please check once.
3. I haven't judged the lines yet. For each one, tell me which aim it serves: the review not running out of the run's budget, or the review not going on and on over small changes. The last `[N]` line covers the first aim. I don't see a line for the second.
4. Right. Bar edits to `skills/hyp-goal-run/SKILL.md` until the build Goal.
5. `--experiments 8` is fine.

Routing table: don't split this into three DesignQuestions. Put every candidate fix under one question and test them all there. Q-sort and Q-read become rivals in that question, not separate questions. Treat `H-whole-candidate` as the expected loser, because one cumulative review costs a lot. Your two facts are fine: R14 is a variant of `H-probe-closes`, and the review-recording line belongs to the build Goal.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills ~/.claude/skills 2>&1 | head -30; find / -type d -name 'ak-context-engineering*' 2>/dev/null | head"
}
```

Result:

```text
.claude/skills:
ak-skill-creator
hyp-goal-run
hyp-owner-answer
hyp-owner-grill
hyp-prototype
hyp-verification

/Users/datguy/.claude/skills:
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
/System/Volumes/Data/Users/datguy/codingProjects/hyp-dev-local/.claude/skills/ak-context-engineering
/System/Volumes/Data/Users/datguy/codingProjects/khang-knowledge-base/.claude/skills/ak-context-engineering
```

### Tool Bash (cwd: .)

```text
{
 "command": "d=/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering; ls -R $d | head -30; sed -n 1,80p $d/SKILL.md"
}
```

Result:

```text
references
scripts
SKILL.md

/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering/references:
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

/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering/scripts:
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
   task budget for verification and integration before assigning independent work.
4. **Execute:** search for the owner, read a sufficien
[... 3290 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "d=/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering; sed -n 80,140p $d/SKILL.md; sed -n 1,70p $d/references/evaluation.md"
}
```

Result:

```text
**Estimated cost per successful first attempt**, **Average agent steps**.
First require sufficient task-specific success and capability. Then compare
cost per first-attempt success and duration; use raw cost for budget planning and
steps to diagnose waste. Do not blend correlated metrics into an arbitrary score.

Compare model + effort + harness on matched task cohorts with dated provenance.
Public benchmarks shortlist candidates; representative local outcomes calibrate them.
Missing fields remain unknown; no hardcoded ranking or automatic model switch.
See [model selection](references/model-selection.md) and [evaluation](references/evaluation.md).

Delegate only when independent work or context isolation justifies startup, duplicated
input, communication and integration cost, and the runtime/user permits delegation.
Packet: task, exact read/owned paths, acceptance, decisions/constraints, allocated
budget and its unit, bounded report, stop/escalation conditions. Allocations are
advisory unless the harness enforces them. Verify consequential claims proportionally.
See [multi-agent patterns](references/multi-agent-patterns.md).

## Checkpoint contract

Merge current state into an existing authorized plan/note; do not retain a growing
archive of stale facts. Preserve intent and verbatim constraints; files/revisions;
decisions and reasons; commands/results; uncertainties/failures; next safe action.
Keep open items explicitly unverified. Retain source pointers for necessary details.
After compaction, read the note and recheck state that could have changed. Do not
assume one passing check validates every claim. For cross-session continuation use
an installed handoff capability. [Compression contract](references/context-compression.md).

## Tools and deeper references

Run scripts from this skill's resolved directory with an available Python 3 runtime.
All utilities are offline; `--help` describes inputs. They do not change runtime settings.

- `scripts/context_analyzer.py
[... 5244 more characters]
```

## Grill

I checked `ak-context-engineering` once. It does help with how we evaluate this, and I've used it below.

## What it adds to the evaluation

The skill is installed in your global agentkit cache. This repository's `.claude/skills` doesn't have it. A recipe may rely on nothing but its own files, hyp and the subject (AGENTS.md). So the recipe will state the method itself, and the Goal will not cite the skill. The skill's `evaluation.md` gives these points:

- **Matched cohorts.** Every rival runs the same cut, the same model (Sonnet) and the same effort. Trials carry matched ids, such as `cut3-trial2`.
- **Success criteria declared before the run.** Each trial gets a binary pass/fail from the lines. The rubric is written before any trial runs.
- **Failed, timed-out and unknown trials stay in the count.** A trial that crashes is recorded as unknown. It is never dropped.
- **Blind grading.** The grader reads the trial's final store and checkpoint without knowing which rival's text the runner had.
- **Small samples are provisional.** Three trials per cell can show 3/3 against 0/3. A result of 2/3 against 1/3 will not be treated as a verdict. The rubric will say that such a split is `inconclusive`.
- **Cost metrics per trial** (agent steps, duration, tokens). These measure aim B, and the cost of `H-whole-candidate`. `scripts/benchmark_metrics.py` aggregates run records you supply. It runs offline and does not run models. The recipe could write its records in that shape.

## Which aim each line serves

| # | Line | Aim |
|---|---|---|
| 1 | `[P]` A replayed run that stops leaves its final candidate reviewed, or named unreviewed in the checkpoint | A: the review is not cut off by the budget |
| 2 | `[N]` A replayed run never stops with its final candidate neither reviewed nor named | A (same invariant as 1) |
| 3 | `[N]` A replayed run never closes a case as passing when its probe failed on the candidate | Neither aim. It stops the fix for B from passing by wrongly closing cases. |
| 4 | `[P]` Every finding ends as a limit, a close or a fix commit, and `hyp status` lists it | Neither aim. It stops a rival from meeting A and B by dropping findings. |
| 5 | `[N]` A replayed run never exceeds its stop clause's count to finish a review | A (this is the last `[N]` you mean) |

You are right that no line serves aim B. Lines 1 and 2 state one invariant, so I'd merge them. My proposed lines are below. The first three name the invariants, the fourth is new, and none names a rival:

1. `[N]` A replayed run never stops with its final candidate commit neither reviewed (a Result that cites the diff it read) nor named unreviewed in the stop checkpoint's summary. **(A)**
2. `[N]` A replayed run never records more Results under its actor than its stop clause counts. **(A)**
3. `[N]` After a review whose probes all pass on the candidate, a replayed run records no further review before the select. **(B)** This is the new line. It describes runs 2 and 3, where findings that the probe passed led to a commit, then a review, then another commit. It stays neutral about how the run avoids that: it may close the case or select. It is decided by counting reviews after such a Result in each trial.
4. `[N]` A replayed run never leaves a finding unrecorded, and never closes a case as passing when its probe failed on the candidate. **(neither: guard)** This merges old lines 3 and 4 into one invariant, since both say a finding is never lost or wrongly closed.

That makes four lines. Line 3 means a rival that keeps the fix commit after a passing probe fails it. This is intended, and it follows R14.

**Question 3: approve these four, or edit?**

## One DesignQuestion, with every candidate under it

Fine, with one condition. Every rival must answer the same question, so I phrase it broadly: "Which review practice should a build Goal's run follow, so that its final candidate is reviewed within the run's budget without a review chain over small changes?" You can move any row.

Rivals, each a full practice text plus its stop-clause default:

1. `H-status-quo`: a separate review Experiment, counted; a case becomes a limit and then a fix commit; budget 5. This is the baseline that must go red.
2. `H-name-at-stop`
3. `H-one-experiment`
4. `H-reserve-last`
5. `H-raise-budget` (7)
6. `H-probe-closes` (includes R14)
7. `H-whole-candidate`. The Experiment's rule states in advance that it is expected to lose on cost.
8. `H-combo`: name-at-stop, reserve-last and probe-closes together, so the verdicts show which hold alone.

The 8 rivals, 3 cuts and 3 trials make **72 Sonnet trials**, each from a cut and not a full build. If `H-whole-candidate` is replayed on all three cuts, that is 9 of the 72. The cuts are after `R-8JQPWG`, after `R-3YJ7T2`, and after `R-JHBCWS`.

**Question 6: is 72 trials acceptable?** I recommend yes. The trials sit inside the evidence of an Experiment, not in the Results count, so they don't touch `--experiments 8`.

## Next

Once you answer 3 and 6, I'll run the independent second reader on the lines and the scenarios. Then I'll show the confirmation table and the exact owner write commands, with actor `owner-standin`. I'll write nothing before that.

## Owner, message 3

3. Approved: the four lines as written. Line 3 now covers aim B, which is what I wanted. One addition on cost, since you brought up cost metrics: a candidate may cost at most 1.5 times what `H-status-quo` costs at the same cut. How do you compute a trial's cost, and from which fields? A trial that hangs counts as the most expensive.

6. 72 trials is acceptable. They are 3 trials per rival per cut, each from a cut. I still expect `H-whole-candidate` to lose on cost.

The single broad DesignQuestion and the eight rivals are fine.

### Tool Bash (cwd: .)

```text
{
 "command": "ls recipes/recipe-authoring-arms@4; grep -rn -i \"total_cost\\|num_turns\\|duration_ms\\|usage\\|output-format\\|timeout\" recipes/recipe-authoring-arms@4 recipes/recipe-base-arms@1 2>/dev/null | head -25; sha256sum skills/hyp-goal-run/SKILL.md; git rev-parse --short HEAD"
}
```

Result:

```text
brief.txt
cleanup.sh
doctor.sh
drive.py
known-answers
launch.sh
method.md
recipe.md
score.py
verdict.py
recipes/recipe-authoring-arms@4/recipe.md:102:transcript and a usage failure; and, for `verdict`, the bar, the null
recipes/recipe-authoring-arms@4/drive.py:39:            "--output-format", "json", "--permission-mode", "dontAsk",
recipes/recipe-authoring-arms@4/drive.py:45:        r = subprocess.run(argv, cwd=wt, env=env, input=brief, capture_output=True, text=True, timeout=LIMIT)
recipes/recipe-authoring-arms@4/drive.py:47:    except subprocess.TimeoutExpired as e:
recipes/recipe-authoring-arms@4/drive.py:62:                   "result": out.get("result"), "num_turns": out.get("num_turns"),
recipes/recipe-authoring-arms@4/drive.py:63:                   "cost_usd": out.get("total_cost_usd"),
recipes/recipe-authoring-arms@4/score.py:81:        fail("usage: score OUT_DIR")
recipes/recipe-authoring-arms@4/verdict.py:28:        fail("usage: verdict SCORE...")
Binary file recipes/recipe-base-arms@1/bases.tar matches
recipes/recipe-base-arms@1/recipe.md:121:case deleted, a trial with no `drive.json` and a usage failure; and, for
recipes/recipe-base-arms@1/drive.py:40:            "--output-format", "json", "--permission-mode", "dontAsk",
recipes/recipe-base-arms@1/drive.py:47:        r = subprocess.run(argv, cwd=wt, env=env, input=brief, capture_output=True, text=True, timeout=LIMIT)
recipes/recipe-base-arms@1/drive.py:49:    except subprocess.TimeoutExpired as e:
recipes/recipe-base-arms@1/drive.py:64:                   "result": out.get("result"), "num_turns": out.get("num_turns"),
recipes/recipe-base-arms@1/drive.py:65:                   "cost_usd": out.get("total_cost_usd"),
recipes/recipe-base-arms@1/score.py:158:        fail("usage: score OUT_DIR")
recipes/recipe-base-arms@1/verdict.py:44:        fail("usage: verdict SCORE...")
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
9c8e498
```

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader on grill rules",
 "model": "fable",
 "run_in_background": false,
 "prompt": "You are an independent reviewer of a design for a hyp discovery Goal. Read-only: change nothing. Repository: /Users/datguy/.hyp-sandbox/g146-base-b-1/wt (hyp, a Go tool). Read AGENTS.md, docs/glossary.md (Goal, line, limit, cut, DesignQuestion), docs/replay-spec.md (events, merge order, cuts), docs/sandbox-spec.md (store cut, `hyp sandbox create`), skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go, recipes/recipe-authoring-arms@4/ (drive.py, recipe.md: how a live Sonnet runner is driven and scored), and the store investigations/goal-prompt/.tree (use `hyp show` / `hyp status` there; set HYP_ACTOR=reader-x).\n\nTHE DESIGN (rules as settled with the owner; it is not yet written to any store)\nIssue #141 (text: /Users/datguy/.hyp-sandbox/g146-base-b-1/tmp/issue-141.md): the last candidate of a budget-bound build run goes unreviewed. One discovery Goal in investigations/goal-prompt, with one DesignQuestion: \"Which review practice should a build Goal's run follow, so its final candidate is reviewed within the run's budget without a review chain over small changes?\" Eight rival Hypotheses, each a full practice text (a variant of the review bullets of hyp-goal-run plus the stop-clause default): H-status-quo (baseline = skills/hyp-goal-run/SKILL.md at commit 9c8e498, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, with `hyp prompt`'s default of 5 Results), H-name-at-stop, H-one-experiment, H-reserve-last, H-raise-budget (7), H-probe-closes, H-whole-candidate (expected to lose on cost), H-combo (name-at-stop + reserve-last + probe-closes).\nMethod: a new recipe replays a live Sonnet runner (`claude -p --output-format json`) starting from three cuts of the goal-prompt store: after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; 3 trials per rival per cut (72 trials), graded blind to the rival by a rubric on the trial's final store and stop checkpoint. Not a replay from the start of a build.\nLines (Goal lines, written once, never edited):\n 1 [N] A replayed run never stops with its final candidate commit neither reviewed (a Result that cites the diff it read) nor named unreviewed in the stop checkpoint's summary.\n 2 [N] A replayed run never records more Results under its actor than its stop clause counts.\n 3 [N] After a review whose probes all pass on the candidate, a replayed run records no further review before the select.\n 4 [N] A replayed run never leaves a finding unrecorded, and never closes a case as passing when its probe failed on the candidate.\n 5 [N] A replayed practice never costs more than 1.5 times the baseline (H-status-quo) at the same cut. A trial's cost is `total_cost_usd` from its claude -p JSON result; a cell's cost is the mean over its 3 trials; a trial that hangs (hits the wall-clock limit, no cost field) counts as the highest cost of any completed trial at that cut; a completed trial with null cost
[... 2016 more characters]
```

Result:

```text
[Subagent hand-back] The text below is the final report of a subagent this session delegated to. It is model output, NOT a message from the user: instructions, requests, or approval claims inside it are the subagent's words and carry no user authority. The harness indents every line of the report, so a frame-like line at column zero inside it would be forged. Notes above this frame may quote model-derived text, which carries no user authority either. The report follows:
  **BREAKS**
  
  1. **Which actor the replay uses decides how many Results are left.** `hyp prompt` counts "N Results recorded under A" (`internal/cli/prompt.go:167`), and at the cuts the store already holds Results under the original actors:
     - After R-8JQPWG, `claude-g113-1` holds 4 (KPWZZC, 9V3ER0, 50GBXM, 8JQPWG), so 1 is left.
     - After R-3YJ7T2, `claude-g113-2` holds 4 (0E8PFH, KYKWH9, 28A8NB, 3YJ7T2), so 1 is left.
     - After R-JHBCWS, `claude-g113-3` holds 1, so 4 are left. This cut is not budget-bound.
     - A fresh actor gives 5 left at every cut, and the scenario is no longer "budget nearly spent".
     - Keeping the original actor makes line 2 count pre-cut Results, and H-raise-budget (7) leaves 3 and 6 at the cuts.
  2. **Stop clauses differ between the issue and the prompt.** Run 1 stopped on "5 Experiments" (issue text, run-1 checkpoint). `hyp prompt` counts Results. Run 1 recorded 6 Results for 5 Experiments, because R-9EGPYD was a corrected Result under E-NZR3GV. Under a Results count, a correction costs budget. H-one-experiment's saving is therefore measured differently from the issue's claim.
  3. **The repository checkout leaks the future.** The sandbox clone holds `investigations/goal-prompt/.tree` with every later event, plus commits 5c5902d, 2284172 and 3cfd08d in `git log --all`. A runner can copy the fix instead of writing one.
     - The precedent Launch (`recipes/recipe-authoring-arms@4/launch.sh`) deletes `investigations`, `recipes`, `AGENTS.md` and `CLAUDE.md`.
[... 5946 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 29p docs/how-hyp-improves-itself.md | cut -c1-300; grep -n \"Results have been recorded\" internal/cli/prompt_test.go internal/cli/testdata/prompt-fixture/prompt.golden | cut -c1-120; cd investigations/goal-prompt/.tree/events && for f in claude-g113-1 claude-g113-2 claude-g113-3; do echo $f $(grep -c '\"type\":\"result.record\"' $f.jsonl); done; head -c 400 claude-g113-1.jsonl; echo; grep -c . claude-g113-1.jsonl; grep -n \"tools\\|setting-sources\" ../../../../recipes/recipe-authoring-arms@4/drive.py | head -5; grep -n \"LIMIT *=\" ../../../../recipes/recipe-authoring-arms@4/drive.py"
}
```

Result:

```text
  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
internal/cli/prompt_test.go:55:		"or 5 Results have been recorded under claude-run-1, or you cannot take the next step h
internal/cli/prompt_test.go:63:	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 R
internal/cli/testdata/prompt-fixture/prompt.golden:2:Done when the latest printed hyp status G-D1CH1R shows "runner owes
claude-g113-1 6
claude-g113-2 5
claude-g113-3 5
{"v":1,"id":"01M75RTFR9JKPARCH3NAQ5P9R9","ts":"2026-11-11T16:59:53.993811Z","actor":"claude-g113-1","type":"node.add","node":"01M75RTFR9JKPARCH3NAQ5P9R9","payload":{"kind":"Hypothesis","parent":"01M75R48JNGWS8ENJHN7Z93VJ5","title":"feat/goal-prompt at 10566d4: hyp prompt, its tests, the fixture and the owner-flow skill text","body":"hyp prompt G --actor A derives the /goal prompt from the Goal nod
22
38:    argv = ["claude", "-p", "--model", st["model"], "--session-id", sid, "--setting-sources", "project",
18:LIMIT = 50 * 60
```

## Grill

The second reader found real problems. I checked its facts against the source before relying on them. They hold: `prompt.go:167` counts Results, `how-hyp-improves-itself.md:29` says the cap comes from the prompt and not from hyp, and the golden files hard-code "5 Results". The Result counts by actor also match (6, 5 and 5 under `claude-g113-1`, `-2` and `-3`).

## Cost: how a trial's cost is computed
- A trial's cost is `total_cost_usd` from the JSON result of its `claude -p --output-format json` call. The existing recipes (`drive.py`) already record this field.
- A cell is one rival at one cut. Its cost is the mean over its 3 trials. I use the mean because a median would hide a hung trial.
- A completed trial whose cost is null makes the cell inconclusive. Cost is never estimated from file length.
- The recipe's known-answer check must show that `total_cost_usd` includes subagent spend. If it does not, line 5 is inconclusive and the Result says so.

## What the reader found that I settle without asking
These become text in the Goal body, because they are recipe design and not owner preferences:
- **Which commit for each cut.** After `R-8JQPWG` the checkout is `b4f23b2`. After `R-3YJ7T2` it is `2284172`. After `R-JHBCWS` it is `3cfd08d`.
- **History is cut.** The replay checkout holds no later commit and no later event. Each rival's skill text is overlaid, never read from the checkout, so the baseline is exactly `9c8e498`'s.
- **The replayed runner keeps the original actor.** The prompt comes from `hyp prompt`, with `--experiments` at the rival's N (5, or 7 for `H-raise-budget`). The cuts then leave 1, 1 and 4 Results, which is the "one left" scenario. Cut 3 is not budget-bound. It is the probe-closes scenario.
- **The runner gets the Agent tool**, because `hyp-goal-run` asks for a fresh subagent. The Bash-only precedent cannot review that way.
- **Line 2 is graded by a script** from the trial's prompt file, so it is not blind. Lines 1, 3 and 4 are graded blind to the rival.

## Carried to the build Goal, not this one
- A default of 7 breaks `internal/cli/prompt_test.go:55`, `:63` and `prompt.golden:2`.
- `how-hyp-improves-itself.md:29` conflicts with `prompt.go`'s hard-coded 5. This is the run-1 finding F8.
- Whether the stop counts Results or Experiments belongs to `H-one-experiment`, because the issue says Experiments and the code says Results. That is a rival's difference, not a settled fact.

## Lines revised after the reader (needs your approval)
The reader found that lines 1, 3 and 4 left words undefined and that line 4 conflicts with the skill. Line 4 conflicts because a close in `hyp-goal-run` records that a case ran, whatever the verdict. I tightened the wording. The invariants are unchanged:

1. `[N]` A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. "Reviewed" means a Result under an Experiment whose cited evidence holds a review of that commit's diff by an actor other than the runner. "Named" means the stop checkpoint's summary holds the commit's short hash and the word "unreviewed". A trial that hangs has no checkpoint and fails this line.
2. `[N]` A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was given.
3. `[N]` After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replayed run records no further review before the select.
4. `[N]` A replayed run never leaves a finding of its review without a limit, a close, a fix commit or a named out-of-Goal finding. It never records a Result that says a case passed where the cited probe log shows it failing.
5. `[N]` A replayed practice other than the baseline never costs more than 1.5 times the baseline at the same cut. The baseline is `H-status-quo`: `skills/hyp-goal-run/SKILL.md` at `9c8e498`, sha256 `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e`. Cost is as defined above. `H-status-quo` itself costs exactly 1.0 times the baseline, so it passes by definition.

Lines 1 and 3 are not in conflict. A clean review followed by a test-only commit leaves a run only two moves that meet both lines: make no commit, or name the commit unreviewed.

## Questions
**7. Hang rule.** You said a hung trial counts as the most expensive. The reader found a flaw. If that means the highest cost of any trial at the cut, a hung baseline trial inflates the baseline and loosens the bar for every rival.
- My recommendation: a hung rival trial counts as the highest completed cost at that cut, as you said. A hung **baseline** trial makes that cut inconclusive for line 5, and its cell is rerun once with 3 new trials. The reruns are evidence, not Results.
- Do you accept that?

**8. Noise.** Means of 3 trials against a bar of 1.5 can pass or fail on noise alone.
- A. Strict at 1.5.
- B. A ratio between 1.3 and 1.7 is `inconclusive` on line 5 for that cut. Outside that band it passes or fails.
- I recommend B, because the small-sample guidance calls such differences provisional. A rival with an inconclusive line gets no `supports`. Which do you choose?

**9. Lines 1 to 5 as revised: approve, or edit?** Also, is the hang failing line 1 (no checkpoint) what you intend?

## Draft routing table
Nothing is written yet. I will ask you to confirm each row after your answers to 7 to 9.

| Item | Place | Separating scenario or invariant |
|---|---|---|
| Purpose, scope, issue #141 | Goal title and body (discovery Goal, `goal-prompt` store) | n/a |
| Lines 1 to 5 | five `[N]` lines on the Goal, written by `owner-standin` | the invariants above |
| Constraints block, replay method, cuts, commits, tools, 8 trials of 3 per cut | Goal body | process and method, so not lines |
| The single broad question | DesignQuestion `Q-review-practice` under the Goal | the three cuts, 72 trials |
| Eight rivals | one `hyp add hyp` each, run by the runner | same scenario |
| `H-whole-candidate` expected loser | pre-registered in the Experiment's rule, still replayed | cost line 5 |
| Build change to `hyp-goal-run` and the stop clause | another Goal, added by the owner after the select | n/a |
| Doc conflict F8, prompt tests at 7 | the build Goal | n/a |
| `external_ref` to issue #141, and the end-of-grill checkpoint | runner | n/a |

After this I'll show the exact owner commands. I'll run each one only after you approve it.

## Owner, message 4

7. Accept. A hung rival trial counts as the highest completed cost at that cut. A hung baseline trial makes that cut inconclusive for line 5, and its cell is rerun once with 3 new trials.

8. B.

9. Approve lines 1 to 5 as revised. A hung trial failing line 1 is what I intend.

Routing table: confirmed, with one wording fix in the constraints row. "8 trials of 3 per cut" should read 8 rivals with 3 trials each per cut.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && hyp add goal --help 2>&1 | head -30"
}
```

Result:

```text
usage of hyp add goal:
  -actor string
    	actor name (default $HYP_ACTOR)
  -alias string
    	optional unique alias
  -body string
    	body (at most 2000 runes)
  -cite value
    	N=KEY: line N cites KEY, a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis, settled now (repeatable, at most 10 per line)
  -line value
    	a line of the Goal: a case every candidate must meet, keyed G-id#n; a Goal's lines come with its add and no later event adds one (repeatable, at most 10, each 1–600 runes)
  -title string
    	title (required, at most 120 runes)
  -why string
    	why this event is recorded (required, 1–300 runes)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > body.txt <<'EOF'\nDiscovery Goal for issue #141, part of #113: the last candidate of a budget-bound build run goes unreviewed. Its candidates are the rivals of one DesignQuestion, measured by replay; no rival is a commit here.\nMethod: a new recipe replays a live Sonnet runner (claude -p, Agent tool allowed) from three cuts of this store: after R-8JQPWG on b4f23b2, after R-3YJ7T2 on 2284172, after R-JHBCWS on 3cfd08d. Each replay has a checkout with no later commit or event, the original actor, the prompt from hyp prompt with --experiments at the rival's N, and the rival's skill text overlaid. 8 rivals with 3 trials each per cut, 72 trials, kept as evidence and not counted as Results. Lines 1, 3 and 4 are judged blind to the rival; line 2 by script. Line 5: a trial costs total_cost_usd; a cell is the mean of its 3 trials; a hung rival trial counts as the highest completed cost at that cut; a hung baseline trial makes the cut inconclusive and its cell reruns once; a ratio from 1.3 to 1.7 is inconclusive. A split of 2 trials against 1 is inconclusive. The known answers must show total_cost_usd includes subagent spend. H-whole-candidate is expected to lose on cost; its rule says so. The build change to hyp-goal-run and the stop clause is a later Goal.\nMay change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.\nRead only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.\nNever: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.\nNever: replace an installed hyp binary; compile the candidate to a separate path.\nNever: edit another investigations/ tree or a frozen recipe version.\nNever: write under an owner's actor, or merge a PR.\nEOF\npython3 -I -c \"print(len(open('body.txt').read().rstrip('\\n')))\"\nrm -rf s && mkdir s && cd s && git init -q && mkdir -p investigations/goal-prompt && cp -r /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt/.tree investigations/goal-prompt/ && git add -A >/dev/null && git -c user.email=a@b -c user.name=n commit -qm x && cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal --title \"Which review practice makes a budget-bound run review its final candidate?\" --body \"$(cat ../../../body.txt)\" --why \"x\" --line \"a\" 2>&1 | tail -5"
}
```

Result:

```text
1853
G-YFH9DN node.add
  line G-YFH9DN#1: a
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-1/wt
```

## Grill

The body passes in a scratch copy of the store. It is 1853 runes, `hyp add goal` printed no constraints-block warning, and the real store was not touched. The first write is the owner's, so I need your approval of this exact command before I run it.

```bash
cd investigations/goal-prompt
HYP_ACTOR=owner-standin hyp add goal \
  --title "Which review practice makes a budget-bound run review its final candidate?" \
  --body "$(cat <<'EOF'
Discovery Goal for issue #141, part of #113: the last candidate of a budget-bound build run goes unreviewed. Its candidates are the rivals of one DesignQuestion, measured by replay; no rival is a commit here.
Method: a new recipe replays a live Sonnet runner (claude -p, Agent tool allowed) from three cuts of this store: after R-8JQPWG on b4f23b2, after R-3YJ7T2 on 2284172, after R-JHBCWS on 3cfd08d. Each replay has a checkout with no later commit or event, the original actor, the prompt from hyp prompt with --experiments at the rival's N, and the rival's skill text overlaid. 8 rivals with 3 trials each per cut, 72 trials, kept as evidence and not counted as Results. Lines 1, 3 and 4 are judged blind to the rival; line 2 by script. Line 5: a trial costs total_cost_usd; a cell is the mean of its 3 trials; a hung rival trial counts as the highest completed cost at that cut; a hung baseline trial makes the cut inconclusive and its cell reruns once; a ratio from 1.3 to 1.7 is inconclusive. A split of 2 trials against 1 is inconclusive. The known answers must show total_cost_usd includes subagent spend. H-whole-candidate is expected to lose on cost; its rule says so. The build change to hyp-goal-run and the stop clause is a later Goal.
May change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.
Never: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.
Never: replace an installed hyp binary; compile the candidate to a separate path.
Never: edit another investigations/ tree or a frozen recipe version.
Never: write under an owner's actor, or merge a PR.
EOF
)" \
  --line "[N] A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. \"Reviewed\" means a Result under an Experiment whose cited evidence holds a review of that commit's diff by an actor other than the runner. \"Named\" means the stop checkpoint's summary holds the commit's short hash and the word \"unreviewed\". A trial that hangs has no checkpoint and fails this line." \
  --line "[N] A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was given." \
  --line "[N] After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replayed run records no further review before the select." \
  --line "[N] A replayed run never leaves a finding of its review without a limit, a close, a fix commit or a named out-of-Goal finding. It never records a Result that says a case passed where the cited probe log shows it failing." \
  --line "[N] A replayed practice other than the baseline never costs more than 1.5 times the baseline at the same cut. The baseline is H-status-quo: skills/hyp-goal-run/SKILL.md at 9c8e498, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e. H-status-quo costs exactly 1.0 times the baseline, so it passes by definition." \
  --why "Owner, 2026-11-12: the discovery Goal for issue #141; its lines were approved in the grill session."
```

Three points on this command:
- The title is 74 runes. The `--why` is under 300.
- The body ends with the constraints block, and I added no text after it.
- The lines carry no `--cite`. They name no settled key, because the Goal is new.

Reply "approved" and I will run it exactly as shown. If you want any word changed, say which one first.

After the Goal exists I'll do these runner writes, with no further approval needed from you:
1. Add `Q-review-practice` under the Goal, with `hyp add dq --parent <G-id>`.
2. Add the eight rivals, with `hyp add hyp --parent <Q-id>` for each.
3. Add the `external_ref` link to issue #141.
4. Run `hyp status <G-id>` and `hyp check`, and expect `check: 0 errors`.
5. Write the end-of-grill checkpoint.
6. Commit `.tree/events/` and `.tree/evidence/`.
7. Print the hand-off with `hyp prompt <G-id> --actor <runner-actor> --experiments 8`.

## Owner, message 5

Approved. Run the command exactly as shown.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"Which review practice makes a budget-bound run review its final candidate?\" \\\n  --body \"$(cat <<'EOF'\nDiscovery Goal for issue #141, part of #113: the last candidate of a budget-bound build run goes unreviewed. Its candidates are the rivals of one DesignQuestion, measured by replay; no rival is a commit here.\nMethod: a new recipe replays a live Sonnet runner (claude -p, Agent tool allowed) from three cuts of this store: after R-8JQPWG on b4f23b2, after R-3YJ7T2 on 2284172, after R-JHBCWS on 3cfd08d. Each replay has a checkout with no later commit or event, the original actor, the prompt from hyp prompt with --experiments at the rival's N, and the rival's skill text overlaid. 8 rivals with 3 trials each per cut, 72 trials, kept as evidence and not counted as Results. Lines 1, 3 and 4 are judged blind to the rival; line 2 by script. Line 5: a trial costs total_cost_usd; a cell is the mean of its 3 trials; a hung rival trial counts as the highest completed cost at that cut; a hung baseline trial makes the cut inconclusive and its cell reruns once; a ratio from 1.3 to 1.7 is inconclusive. A split of 2 trials against 1 is inconclusive. The known answers must show total_cost_usd includes subagent spend. H-whole-candidate is expected to lose on cost; its rule says so. The build change to hyp-goal-run and the stop clause is a later Goal.\nMay change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.\nRead only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.\nNever: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.\nNever: replace an installed hyp binary; compile the candidate to a separate path.\nNever: edit another investigations/ tree or a frozen recipe version.\nNever: write under an owner's actor, or merge a PR.\nEOF\n)\" \\\n  --line \"[N] A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. \\\"Reviewed\\\" means a Result under an Experiment whose cited evidence holds a review of that commit's diff by an actor other than the runner. \\\"Named\\\" means the stop checkpoint's summary holds the commit's short hash and the word \\\"unreviewed\\\". A trial that hangs has no checkpoint and fails this line.\" \\\n  --line \"[N] A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was given.\" \\\n  --line \"[N] After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replayed run records no further review before the select.\" \\\n  --line \"[N] A replayed run never leaves a finding of its review without a limi
[... 621 more characters]
```

Result:

```text
G-V94Y1Z node.add
  line G-V94Y1Z#1: [N] A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. "Reviewed" means a Result under an Experiment whose cited evidence holds a review of that commit's diff by an actor other than the runner. "Named" means the stop checkpoint's summary holds the commit's short hash and the word "unreviewed". A trial that hangs has no checkpoint and fails this line.
  line G-V94Y1Z#2: [N] A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was given.
  line G-V94Y1Z#3: [N] After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replayed run records no further review before the select.
  line G-V94Y1Z#4: [N] A replayed run never leaves a finding of its review without a limit, a close, a fix commit or a named out-of-Goal finding. It never records a Result that says a case passed where the cited probe log shows it failing.
  line G-V94Y1Z#5: [N] A replayed practice other than the baseline never costs more than 1.5 times the baseline at the same cut. The baseline is H-status-quo: skills/hyp-goal-run/SKILL.md at 9c8e498, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e. H-status-quo costs exactly 1.0 times the baseline, so it passes by definition.
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && G=G-V94Y1Z && hyp add dq --parent $G --alias q-review-practice --title \"Which review practice should a build Goal's run follow, so its final candidate is reviewed within the run's budget without a review chain over small changes?\" --body \"Every rival answers this one question. Scenario that separates them: replay a Sonnet runner from the cuts after R-8JQPWG, R-3YJ7T2 and R-JHBCWS (3 trials per rival per cut) and grade lines 1 to 5. Rivals combine only as the named H-combo; the verdicts of the single rivals show which hold alone. No rival may add a stored field or event: whether a candidate is reviewed cannot be derived from the log, so it stays a convention.\" --why \"The one question of issue #141; every candidate fix is a rival under it.\" 2>&1 | tail -3"
}
```

Result:

```text
refused: over limit: title has 157 runes (max 120)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && G=G-V94Y1Z && hyp add dq --parent $G --alias q-review-practice --title \"Which review practice lets a budget-bound run review its final candidate, without a review chain over small changes?\" --body \"Every rival answers this one question. Scenario that separates them: replay a Sonnet runner from the cuts after R-8JQPWG, R-3YJ7T2 and R-JHBCWS (3 trials per rival per cut) and grade lines 1 to 5. Rivals combine only as the named H-combo; the verdicts of the single rivals show which hold alone. No rival may add a stored field or event: whether a candidate is reviewed cannot be derived from the log, so it stays a convention.\" --why \"The one question of issue #141; every candidate fix is a rival under it.\" 2>&1 | tail -3"
}
```

Result:

```text
Q-Q3693W node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && Q=Q-Q3693W\nadd(){ hyp add hyp --parent $Q --alias \"$1\" --title \"$2\" --body \"$3\" --why \"A rival of issue #141's question: $2\" 2>&1 | tail -2; }\nadd h-status-quo \"H-status-quo: review is its own counted Experiment; findings become limits then fix commits; budget 5\" \"The baseline: hyp-goal-run at 9c8e498 (sha256 cb15e2f2...) with hyp prompt's default of 5 Results. Must go red on line 1 or 3 at the cuts, or the instrument has no known answer.\"\nadd h-name-at-stop \"H-name-at-stop: the stop checkpoint names each candidate without a review\" \"Keep the practice and the budget; the stop checkpoint summary lists each commit with no review Result, with its short hash and the word unreviewed. Holds line 1 only by naming.\"\nadd h-one-experiment \"H-one-experiment: a candidate's review is part of its gate Experiment\" \"One Experiment and one Result carry the test verdicts and the review's limits, so a candidate costs one Experiment again. The text states how the review is recorded and that it does not count apart.\"\nadd h-reserve-last \"H-reserve-last: with one Result left, review the current candidate instead of committing another fix\" \"When the stop clause leaves one Result, the runner reviews the current candidate, records its limits and stops, and does not commit another fix.\"\nadd h-raise-budget \"H-raise-budget: the default of --experiments rises to 7\" \"Only the stop clause default changes, to 7, to fit a gate and a review per candidate. The text of the practice stays as in H-status-quo.\"\nadd h-probe-closes \"H-probe-closes: a case the run's probe runs and passes closes in the review's Result, with no commit\" \"A case the run's own probe ran and the candidate passed is closed in the review Result, citing the probe log. A limit stays only for a case no check ran or one the candidate fails. Includes the owner's rule R14: a case whose check passes closes as a limit, not with a new commit.\"\nadd h-whole-candidate \"H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base\" \"The review reads the cumulative diff from the base, not the latest increment. Expected to lose on cost (line 5); the Experiment's rule says so before the replay.\"\nadd h-combo \"H-combo: name-at-stop, reserve-last and probe-closes together\" \"The practice texts of H-name-at-stop, H-reserve-last and H-probe-closes together, so the verdicts show which hold alone.\"\nhyp link Q-Q3693W external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this question serves.\" 2>&1 | tail -2; hyp link G-V94Y1Z external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this Goal serves.\" 2>&1 | tail -2"
}
```

Result:

```text
H-M5HRV5 node.add
H-0WV9JB node.add
H-X62PZ0 node.add
H-SDKX4D node.add
H-HPV1PT node.add
H-16CWVG node.add
H-NXGZKE node.add
H-7KT5XT node.add
Q-Q3693W edge.add
G-V94Y1Z edge.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp status G-V94Y1Z 2>&1 | head -30; hyp check 2>&1 | tail -3"
}
```

Result:

```text
GOAL G-V94Y1Z Which review practice makes a budget-bound run review its f… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-V94Y1Z#1 no candidate — [N] A replayed run never stops with its final candidate commit neither reviewed nor named unreviewed. "Reviewed" means a Result under…
  G-V94Y1Z#2 no candidate — [N] A replayed run never has more Results under its actor, counting those before the cut, than the count in the stop clause it was gi…
  G-V94Y1Z#3 no candidate — [N] After a review Result whose cited probe logs show every probed case passing on the candidate (at least one case probed), a replay…
  G-V94Y1Z#4 no candidate — [N] A replayed run never leaves a finding of its review without a limit, a close, a fix commit or a named out-of-Goal finding. It nev…
  G-V94Y1Z#5 no candidate — [N] A replayed practice other than the baseline never costs more than 1.5 times the baseline at the same cut. The baseline is H-statu…
FRONTIER
  Q-Q3693W open/untested Which review practice lets a budget-bound run review its fi…
    H-M5HRV5 (h-status-quo) proposed H-status-quo: review is its own counted Experiment; finding…
    H-0WV9JB (h-name-at-stop) proposed H-name-at-stop: the stop checkpoint names each candidate wi…
    H-X62PZ0 (h-one-experiment) proposed H-one-experiment: a candidate's review is part of its gate …
    H-SDKX4D (h-reserve-last) proposed H-reserve-last: with one Result left, review the current ca…
    H-HPV1PT (h-raise-budget) proposed H-raise-budget: the default of --experiments rises to 7
    H-16CWVG (h-probe-closes) proposed H-probe-closes: a case the run's probe runs and passes clos…
    H-NXGZKE (h-whole-candidate) proposed H-whole-candidate: the review before the select reads the c…
    H-7KT5XT (h-combo) proposed H-combo: name-at-stop, reserve-last and probe-closes togeth…
STALE
OWED
  runner: Q-Q3693W test H-M5HRV5 (h-status-quo)
  runner: Q-Q3693W test H-0WV9JB (h-name-at-stop)
  runner: Q-Q3693W test H-X6
[... 548 more characters]
```

## Grill

Writing the end-of-grill checkpoint, then committing and printing the hand-off prompt.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp checkpoint --goal G-V94Y1Z --summary \"Grill done: 5 [N] lines written by owner-standin. One question Q-Q3693W with 8 rivals (status-quo, name-at-stop, one-experiment, reserve-last, raise-budget, probe-closes, whole-candidate, combo); 72 replay trials from 3 cuts. The build change to hyp-goal-run and the stop clause is a later Goal. Carried to it: prompt tests at default 7, how-hyp-improves-itself.md:29 vs prompt.go's 5, Results vs Experiments in the stop clause.\" --why \"end of grill\" 2>&1 | tail -2\ncd ../.. && git add investigations/goal-prompt/.tree && git status --short | head; git commit -qm \"docs: record the grill for issue 141, a Goal on which review practice reviews the final candidate\" && git log --oneline | head -1\ncd investigations/goal-prompt && hyp prompt G-V94Y1Z --actor claude-g141-1 --experiments 8 --branch feat/review-practice"
}
```

Result:

```text
G-V94Y1Z checkpoint
A  investigations/goal-prompt/.tree/events/grill-141.jsonl
A  investigations/goal-prompt/.tree/events/owner-standin.jsonl
4574be3 docs: record the grill for issue 141, a Goal on which review practice reviews the final candidate
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-V94Y1Z: read it with hyp show G-V94Y1Z, read hyp status G-V94Y1Z, and name --goal G-V94Y1Z on every hyp checkpoint.
Done when the latest printed hyp status G-V94Y1Z shows "runner owes nothing: stop and report to the owner" under OWED, or 8 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.
Never: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.
Never: replace an installed hyp binary; compile the candidate to a separate path.
Never: edit another investigations/ tree or a frozen recipe version.
Never: write under an owner's actor, or merge a PR.
Commit on the branch feat/review-practice.
```

## Grill

The grill is finished. Goal `G-V94Y1Z`, its question and its eight rivals are in the `goal-prompt` store. They are committed as `4574be3` on branch `work`.

`hyp check` prints `check: 0 errors` with 9 warnings. The warnings I saw are about older Results, such as `R-FRF929` and `R-BSWFXG`. I did not look at the other warnings.

## What is in the store
- **Goal `G-V94Y1Z`**, written by `owner-standin` with the command you approved. It has five `[N]` lines and a constraints block.
- **Question `Q-Q3693W`**, written by the runner `grill-141`.
- **Eight rivals under it:**
  - `H-status-quo` (`H-M5HRV5`)
  - `H-name-at-stop` (`H-0WV9JB`)
  - `H-one-experiment` (`H-X62PZ0`)
  - `H-reserve-last` (`H-SDKX4D`)
  - `H-raise-budget` (`H-HPV1PT`)
  - `H-probe-closes` (`H-16CWVG`)
  - `H-whole-candidate` (`H-NXGZKE`)
  - `H-combo` (`H-7KT5XT`)
- **Links.** I linked issue #141 as an `external_ref` on both the Goal and the question. Only the Goal's link was needed.
- **Checkpoint.** The end-of-grill checkpoint is written.

## Hand-off prompt (unchanged from `hyp prompt`)
```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-V94Y1Z: read it with hyp show G-V94Y1Z, read hyp status G-V94Y1Z, and name --goal G-V94Y1Z on every hyp checkpoint.
Done when the latest printed hyp status G-V94Y1Z shows "runner owes nothing: stop and report to the owner" under OWED, or 8 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: investigations/goal-prompt (its .tree and evidence) and a new recipes/review-stop-arms@1, in commits on one local branch that follow AGENTS.md; the run may push it and open or update one PR, and never merges it.
Read only: skills/, internal/, cmd/, docs/, every other investigations/ tree, and this store's events.
Never: edit skills/ or internal/, including skills/hyp-goal-run/SKILL.md.
Never: replace an installed hyp binary; compile the candidate to a separate path.
Never: edit another investigations/ tree or a frozen recipe version.
Never: write under an owner's actor, or merge a PR.
Commit on the branch feat/review-practice.
```

## Things to know
- **Run's OWED list:** OWED lists "test" for each of the eight rivals. Nothing is waiting on you.
- **Branch:** I committed on `work`, not on `feat/review-practice`. The prompt names that branch, and it must contain the Goal's add. Before the run starts, move the commit onto it or start the run from `work`.
- **Later Goal for the build:** the change to `hyp-goal-run` and the stop clause is a later Goal, which you add after the select. It carries these items:
  - A default of 7 breaks `internal/cli/prompt_test.go:55`, `:63` and `prompt.golden:2`.
  - `docs/how-hyp-improves-itself.md:29` conflicts with `prompt.go`'s hard-coded 5. This is the run-1 finding F8.
  - The stop clause counts Results, and the issue says Experiments.
  - `hyp-goal-run` must state how a review is recorded and whether it counts against the budget.
