# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-pre-b-3/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-b-3/tmp/issue-141.md"
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
 "command": "hyp status 2>&1 | head -40; ls investigations; ls docs | head -30"
}
```

Result:

```text
no .tree store found; run 'hyp init'
checkpoint-store
checkpoint-store-design
goal-prompt
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

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/investigations/goal-prompt && hyp status 2>&1 | head -80; ls; git -C .. log --oneline -3; hyp version 2>&1 | head -2"
}
```

Result:

```text
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-11 16:47 claude-grill-113
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-12 03:28 claude-g113-4
AWAITING OWNER
evals
9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
hyp 0.1.0 (events v1)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt && grep -n \"review\\|Experiments\\|budget\" skills/hyp-goal-run/SKILL.md | head -60; grep -rn \"experiments\\|Results recorded\" internal/*/prompt*.go | grep -v _test | head -20"
}
```

Result:

```text
6:The `/goal` prompt names the tree, your actor, the Goal, the stop and the Goal body's constraints; `hyp prompt <goal-id> --actor <actor>` prints it from the log. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.
34:- Before the stop, ask a fresh subagent for one review of each candidate commit's diff against the design and the Goal's lines. Give it the code, the lines and the keys already settled, with the owner's words; do not give it your conclusion. Ask it to label each finding with the line it is a case of, or none.
internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
internal/cli/prompt.go:123:	if *experiments < 1 {
internal/cli/prompt.go:124:		return usageErr("--experiments must be at least 1")
internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/investigations/goal-prompt && hyp status G-ZBM2RE 2>&1 | head -120; hyp show G-ZBM2RE 2>&1 | head -60"
}
```

Result:

```text
GOAL G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-ZBM2RE#1 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [P] For a Goal whose body holds a valid constrain…
  G-ZBM2RE#2 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [N] hyp prompt prints no text of an owner answer:…
  G-ZBM2RE#3 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [N] The constraints block runs from the first lin…
  G-ZBM2RE#4 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [N] In a store with two Goals, where one answer e…
  G-ZBM2RE#5 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [N] hyp prompt with no --actor exits 2 and never …
  G-ZBM2RE#6 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [P] For the testdata fixture (checkpoint-store:G-…
  G-ZBM2RE#7 covered on Q-Z93VJ5; ran in R-KPWZZC, R-50GBXM, R-FRF929, R-28A8NB, R-356DV0, R-27Y1EV, R-BSWFXG — [P] Each SKILL.md the candidate changes passes ak…
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
RECENT
  2026-11-12 03:28 claude-g113-4 dq.select Q-Z93VJ5 — H-VHG4N5 (04142f3) is the only candidate; lines G-ZBM2RE#1-#7 ran in R-BSWFXG w…
  2026-11-12 03:23 claude-g113-3 result.record R-HXEBAF — Records E-ZG1BC8 under its pre-registered rule; it closes R-BSWFXG#1 and cites …
  2026-11-12 03:23 claude-g113-3 result.record R-BSWFXG — Records E-GE085A under its pre-registered rule; Q1 is a case of line 3 no check…
  2026-11-12 03:19 claude-g113-3 edge.add H-VHG4N5 — 04142f3 holds d2d5d14 and adds the test R-DER6F8#1 asks for.
  2026-11-12 03:19 claude-g113-3 hypothesis.subject H-VHG4N5 — The candidate is the comm
[... 14743 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt && cat skills/hyp-goal-run/SKILL.md; sed -n 95,200p internal/cli/prompt.go; grep -rn \"experiments\\|Experiments\\|run cap\\|5 Results\" docs/how-hyp-improves-itself.md docs/glossary.md | head -20; ls skills/hyp-goal-run internal/cli | head -50"
}
```

Result:

```text
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

Act on what OWED in `hyp status` says the runner owes, and on nothing else. Plan only the next Experiment.

- Use the Goal's must-fail cases as known-answer flaw cases, in the owner's terms.
- The recipe's checks may exist befor
[... 19142 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n -i \"store cut\\|^### .*cut\\|--cut\\|--store\" docs/glossary.md docs/sandbox-spec.md | head -20; ls investigations/goal-prompt/evals investigations/goal-prompt/.tree 2>/dev/null | head -30; ls recipes; git log --oneline -8 -- skills/hyp-goal-run; ls ~/.claude/skills .claude/skills 2>/dev/null"
}
```

Result:

```text
docs/glossary.md:359:A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. The sandbox is the only place hyp assembles a subject: from the base, a store cut, per-run inputs and one patch for each Arm. Code under test is built in its Launch ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). It is not process isolation: commands still run on the host.
docs/glossary.md:397:**Store cut**:
docs/sandbox-spec.md:158:A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (see Arm in [glossary.md](glossary.md)), given as `--patch ARM=FILE`.
investigations/goal-prompt/.tree:
events
evidence

investigations/goal-prompt/evals:
goal-prompt-evals-2.json
goal-prompt-evals.json
authoring-stops@2
id-citation-arms@1
id-citation-arms@2
recipe-authoring-arms@4
recipe-base-arms@1
5c5902d fix: hyp prompt reads the store's real path and states its line rules
b4f23b2 fix: hyp prompt refuses a line of spaces, a store outside a repository and a bad branch
10566d4 feat: hyp prompt prints a Goal's /goal prompt from its constraints block
ee3828f feat: publish the owner-flow skills in skills/
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
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/investigations/goal-prompt && sed -n 397,402p ../../docs/glossary.md; hyp show R-8JQPWG 2>&1 | head -12; hyp show R-JHBCWS 2>&1 | head -8; hyp show E-HQWTA0 2>&1 | head -6; hyp show E-NZR3GV 2>&1|head -4; ls ../../.claude/skills/ak-skill-creator | head; hyp --help 2>&1 | grep -i -n \"link\\|signals\" | head"
}
```

Result:

```text
**Store cut**:
A copy of a tree's store as it stood at one event, given to a sandbox.
_Avoid_: snapshot (that is a different thing: a capture of a live session)

**Patch**:
A throwaway change over the subject's base that never lands in code, given to `hyp sandbox create` as `--patch ARM=FILE` for one Arm. An Arm has at most one; a patch is never versioned, stacked or rebased, and a change that lands is a commit used as the base instead. hyp checks that it applies to the base, records its sha256 and keeps it with the kept files. It is the subject, not a file of the recipe bundle ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)).
Result R-8JQPWG
path: G-ZBM2RE > Q-Z93VJ5 > H-3M3QAS > E-7QQ2DP > R-8JQPWG
status: recorded
observation: The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and flags, exit 3 through a symlinked working directory and exit 0 with 'in inv/t' through the physical path, where G-ZBM2RE#1 prints the same bytes on every call. It confirms G2 (a CRLF blank line inside the block is refused as "\r") and G3 (a no-break space in --branch exits 2). G3, G4, G5, G6 and G7 are text or test defects; G4 also names a rule conflict outside the lines.
verdicts: H-3M3QAS=contradicts
evidence: .tree/evidence/E-7QQ2DP/review.md
evidence: .tree/evidence/E-7QQ2DP/probe-b4f23b2.log
limit #1: G1, G-ZBM2RE#1: the same store reached through a symlinked working directory and through its physical path must print the same bytes; b4f23b2 exits 3 on one and 0 on the other.
limit #2: G2, G-ZBM2RE#3: a CRLF body whose block holds a blank line, which is "\r" after the split on \n; b4f23b2 refuses it as a non-empty line with no heading.
limit #3: G8, G-ZBM2RE#2 and #4: the tests check the Goal id in one of its four places; a prompt that names G in its header and another Goal in hyp show, hyp status or --goal would pass.
history:
  2026-11-11 17:21 claude-g113-1 result.record R-8JQPWG — Records E-7QQ2DP under its pre
[... 2523 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "head -c 1500 evals/goal-prompt-evals-2.json; echo; hyp show E-NZR3GV 2>&1 | sed -n 5,30p | cut -c1-300; ls .tree/evidence/E-NZR3GV 2>/dev/null | head"
}
```

Result:

```text
{
 "skill_name": "hyp-owner-flow",
 "evals": [
  {
   "id": "l6-run6-clauses",
   "split": "holdout",
   "prompt": "No consumer run. The artifacts are the prompt that hyp prompt G-D1CH1R --actor claude-g1b-6 --branch feat/one-subject-per-run --pr 132 prints for the testdata fixture (prompt.txt), hyp show G-D1CH1R on the fixture (goal-show.txt), and the skills the prompt names with the skills hyp-goal-run names (skills/). Run 6's hand-written prompt (plans/reports/goal-prompt-261111-2129-g1b-run-6.md, sha256 94f61179c63cb43ef63169553541d82c5d75d3b7e78cfe7ac9ccf954ca6285e3) is the reference. Excepted clauses: the actor, the owner answer, the branch, the PR, the commit and the run plan (the Experiment's aims and verdict rule, that run's owner items).",
   "expected_output": "Each non-excepted clause of run 6's prompt is stated by prompt.txt or by a SKILL.md under skills/.",
   "files": [],
   "assertions": [
    {
     "id": "c01-tree",
     "text": "States that the run works in the hyp tree at investigations/checkpoint-store. Evidence: prompt.txt or a skills/*/SKILL.md file."
    },
    {
     "id": "c02-skill",
     "text": "States that the run uses the hyp-goal-run skill. Evidence: prompt.txt or a skills/*/SKILL.md file."
    },
    {
     "id": "c03-goal",
     "text": "States that the run's Goal is G-D1CH1R. Evidence: prompt.txt or a skills/*/SKILL.md file."
    },
    {
     "id": "c04-design",
     "text": "States the design the Goal builds (ADR 0020 / checkpoint-store-de
path: G-ZBM2RE > Q-Z93VJ5 > H-YXSJEM > E-NZR3GV
status: run
recipe: none
rubric: reading
shape: rule by reading
aims: R-8JQPWG#1
aims: R-8JQPWG#2
aims: R-8JQPWG#3
aims: G-ZBM2RE#1
aims: G-ZBM2RE#2
aims: G-ZBM2RE#3
aims: G-ZBM2RE#4
aims: G-ZBM2RE#6
aims: G-ZBM2RE#7
children:
  R-FRF929 recorded H-YXSJEM=supports
  R-9EGPYD recorded H-YXSJEM=supports
history:
  2026-11-11 17:23 claude-g113-1 node.add E-NZR3GV — Pre-registers the rule before any run on 5c5902d; c04-strict and strict-crlf set no ve
[... 178 more characters]
```

## Grill

I've read #141 and its three owner comments, the `hyp-goal-run` skill, `hyp prompt`, and the `investigations/goal-prompt` store. I've written nothing yet. Round 1 is below.

**Settled by lookup, so not asked:**
- **Store:** `investigations/goal-prompt`. It already holds G-GZR6C6 and G-ZBM2RE for the parent issue #113, and two Goals already share it.
- **Scope:** core. It covers the run practice in `hyp-goal-run` and the stop clause `hyp prompt` prints.
- **Premise from your comment 2:** it holds. `skills/hyp-goal-run/SKILL.md` ("Build Goals") asks for one review per candidate and sorts a case of a line into a `--limit`. A limit needs `hyp record`, so it needs an Experiment. No text says whether a review is its own Experiment.
- **Rivals:** all six (H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate) can be separated by a replayed run. They go in the tree as rivals, not as your choice.

**Q1. One Goal or two?**
- *Today:* #141's "Done means" asks for a selected rival and for a changed skill and prompt. A Goal's lines are written once and never edited.
- *Problem:* the changed text can't have lines until a rival is selected.
- *Option A (two Goals):*
  - Goal A is a discovery Goal. Its candidates are the rival rules. It is tested by replaying the cuts of run 1 (after R-8JQPWG), run 2 (after R-3YJ7T2) and run 3 (after R-JHBCWS).
  - Goal B is a build Goal for the commit that edits `hyp-goal-run` and `hyp prompt`. You add it after A selects, with lines that follow the selected rival.
- *Option B (one Goal):* A and B are one build Goal, with lines written before the rival is known.
- *Principle:* the grid rule. A claim and a commit can't both be covered by one run.
- **Recommend A.** I would write only Goal A now and name Goal B in its checkpoint, so you add it.

**Q2. Lines of Goal A.** I propose two, both citing run 1's R-8JQPWG cut as a scenario:
- **[N]** On each replay cut, a run under the candidate rule never reaches its stop with a candidate commit that has neither a Result recording a fresh review of its diff nor a name in the stop checkpoint as unreviewed. Run 1's "all review limits closed" is exactly this state.
- **[P]** On run 2's cut after R-3YJ7T2, both of whose probes pass, a run under the candidate rule ends with `hyp status` showing "runner owes nothing". The final candidate has a review Result, and the budget is not exceeded.
- *Why two:*
  - The [N] line is the "Done means" of #141: reviewed or named.
  - Without the [P] line, H-name-at-stop alone would meet every line, and the tree would learn nothing about the others.
  - With it, H-name-at-stop alone is contradicted, which answers "which hold alone".
- Do you want the [P] line? **Recommend yes.** Do you want to change either line?

**Q3. Is "a passing case costs no new commit" a purpose or a rival?**
- You wrote in comment 2 that a case whose check already passes closes as a limit, not with a new commit. Run 3's N1 and P1 are examples.
- *As a line:* every candidate rule must meet it.
- *As a rival:* it is DQ2 below, against "limit, then a fix commit" and H-probe-closes.
- *Principle:* a run can separate it, so it is a rival. A line would fix the answer before the run.
- **Recommend rival.** Do you agree?

**Q4. The tree's proposed shape.** This is for your yes or no. It is a structure, not a choice among options.
- **DQ1:** how a budget-bound run ends reviewed or named. Rivals are H-one-experiment, H-name-at-stop, H-reserve-last and H-raise-budget.
- **DQ2:** what a finding does when the run's own probe passes the case. Rivals are limit then fix commit (today), limit closed with no commit, and H-probe-closes.
- **DQ3:** what the pre-select review reads. Rivals are the latest increment (today) and H-whole-candidate. Its scenario is a defect planted across two commits, since the real runs never showed one.
- Should the run also test the combinations, such as H-name-at-stop with each other DQ1 rival? That is seven arms instead of four. **Recommend yes, but only for H-name-at-stop**, because #141 says only that rival combines.

**Q5. Test method.** `hyp-goal-run` is a skill, so ak-skill-creator's eval flow owns the method. Each case is a store cut with the original skill and each rival's patched skill, graded blind. The earlier goal-prompt evals were static artifacts with no consumer run.
- Do you want the flow used for consumer runs on the cuts? **Recommend yes.**
- What model, how many trials per case, and what budget? **Recommend** the strongest model, 2 trials per case, and a cap of about 40 consumer runs. A DQ1 arm needs 3 cuts, so seven arms × 3 cuts × 2 trials = 42 runs. That is over the cap, so I suggest running only one trial on the combination arms.

**Q6. Constraints block for Goal A.**
- *May change:*
  - Commits on one local branch, plus the `investigations/goal-prompt` store.
  - A new recipe version under `recipes/`, with its rubric, for the replay.
  - Nothing in `skills/`, `cmd/`, `internal/` or `docs/`: the rival rules are throwaway Arm patches, and the real edit is Goal B's.
- *Read only:* `plans/`, every other `investigations/` tree, and ak-skill-creator pinned by sha256.
- *Never:*
  - Merge a PR.
  - Replace the installed binary.
  - Write under the owner's actor.
- Do you want to confirm this, or change it?

**Your actor name.** I will use `owner-dat`, as in the store. Tell me if it is different.

## Owner, message 2

1. Go with A, two Goals.

2. Put all the candidate fixes under one question, not three. How does each of your two lines serve my two aims? I want a review that doesn't easily run out of the run's budget, and one that doesn't drag on over small changes. I don't see either aim in [N] or [P]. Keep [P] if it covers them, otherwise tell me what's missing.

3. Agree, make it a rival.

4. See my answer to 2 on putting the fixes under one question. I expect H-whole-candidate to lose, because one cumulative review costs a lot. Treat it as the expected loser, but test it. Combinations are your call.

5. Yes to the eval flow, but use Sonnet with 3 trials per candidate and per starting point, not the strongest model and 2 trials. Replay from the cuts only. I don't want a full replay of a whole build from its start for every trial. A candidate may cost at most 1.5× the base at the same starting point. How is cost computed? If you use a median, how do you compute it? A hung trial counts as the most expensive. Could the ak-context-engineering skill help with how we evaluate?

6. Confirmed.

My actor name is owner-standin, not owner-dat. Please use that.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills ~/.claude/skills; find ~/.claude ~/.hyp-sandbox -maxdepth 4 -iname \"*context-eng*\" 2>/dev/null | head; grep -n -i \"median\\|cost\\|token\\|trial\\|timing\" .claude/skills/ak-skill-creator/SKILL.md | head -30"
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
96:evidence; `references/evaluation-tools.md` owns executable records and observed cost.
123:| Layout, metadata fields and context loading | `references/skill-anatomy-and-requirements.md`, `references/yaml-frontmatter-reference.md`, `references/token-efficiency-criteria.md` |
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/.claude/skills/ak-skill-creator && grep -n -i \"cost\\|median\\|token\\|timeout\\|hung\\|duration\\|trial\\|runs_per\" references/evaluation-tools.md | head -50; ls scripts | head -30"
}
```

Result:

```text
46:Optional observed metrics: `--tokens`, `--duration-ms`, `--tool-calls`, `--corrections`.
67:    "input_tokens": null, "output_tokens": null,
68:    "cache_read_tokens": null, "cache_write_tokens": null,
71:  "cost": {
83:cached tokens. Do not sum fields using an assumed universal billing formula.
85:Cost records carry currency, evidence source and pricing date (YYYY-MM-DD).
87:actual savings from estimated or missing costs. Cache state is cold, warm, disabled
110:outputs/traces. Missing metrics or uncovered targets never become zero-cost passes.
113:and cost provenance. They expose `comparison_context`, `missing_evidence`,
123:their aggregate tokens as a proven cost improvement.
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
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "sed -n 40,125p references/evaluation-tools.md"
}
```

Result:

```text
  --run-id unique-session-id --variant candidate \
  --model effective-model-id --runtime observed-runtime-version \
  --skill /absolute/skill-snapshot > /absolute/run/record.json
'''

No-skill: `--variant without_skill`, omit `--skill`. Other variants require a hash.
Optional observed metrics: `--tokens`, `--duration-ms`, `--tool-calls`, `--corrections`.
Omitted metrics stay null. For execution that did not complete, set `--status blocked`
or `--status failed` and retain the reason/provider error beside the record.

### Import observed configuration and usage

Add `--observations /absolute/run/observations.json` to `grade` to import a
provider-neutral evidence object. This file supplies observations only; it cannot
override run identity, skill hash, status or check verdicts. The parser in
`scripts/eval_contract.py` owns accepted fields and validation.

Illustrative shape (null means unavailable, not zero):

'''json
{
  "effective_effort": null,
  "cache_state": null,
  "fingerprints": {
    "configuration": null, "cases": null, "tools": null, "catalog": null
  },
  "usage": {
    "input_tokens": null, "output_tokens": null,
    "cache_read_tokens": null, "cache_write_tokens": null,
    "source": null, "semantics": null
  },
  "cost": {
    "amount": null, "currency": null, "source": null, "pricing_date": null
  },
  "loaded_references": null
}
'''

Use observed effective effort rather than a requested alias. Fingerprints identify
the actual configuration, frozen cases, available tools and catalog; do not put raw
settings or secrets into records. `loaded_references: []` means the trace showed no
references read; null means reads were not observed. Usage source and semantics
describe how the runner mapped provider fields, including whether input includes
cached tokens. Do not sum fields using an assumed universal billing formula.

Cost records carry currency, evidence source and pricing date (YYYY-MM-DD).
Separate measured billing from an estimate in the source descr
[... 2395 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/.claude/skills/ak-skill-creator/scripts && sed -n 1,40p session_usage.py"
}
```

Result:

```text
#!/usr/bin/env python3
"""Measure one skill's invocations in local agent transcripts; read-only, never prints message text.

Scans Claude Code (`<config>/projects/**/*.jsonl`) and Codex (`<home>/sessions/**/*.jsonl`)
transcripts, slices each invocation from its trigger to the next human prompt, and emits
JSON metrics (tokens, turns, tool calls, errors, latency, delegation, skill files read)
plus heuristic signals. Signals are review leads, not verdicts.
"""

import argparse
import json
import os
import re
import statistics
import sys
from collections import Counter, defaultdict
from datetime import datetime
from pathlib import Path

from encoding_utils import configure_utf8_console

MECHANICAL_TOOLS = {
    'Read', 'Grep', 'Glob', 'LS', 'Bash', 'PowerShell', 'WebFetch', 'WebSearch', 'NotebookRead',
    'exec_command', 'shell', 'exec', 'read_file', 'list_dir', 'grep_files', 'view_image', 'wait',
}
DELEGATION_TOOLS = {'Agent', 'Task'}
WAIT_TOOLS = {'sleep', 'wait', 'Monitor', 'ScheduleWakeup', 'write_stdin'}
NON_HUMAN_PREFIXES = ('<system-reminder>', '<task-notification>', '<local-command', 'Caveat:', '<environment_context>',
                      '# AGENTS.md instructions', '# Files mentioned by the user', 'This session is being continued')
INJECTED_EVENT = re.compile(r'<[\w-]+-event\b')
INTERRUPT_PREFIX = '[Request interrupted'
# Only the head of a tool output is inspected, so a failed exit printed inside file content is not an error.
EXIT_CODE = re.compile(r'^.{0,300}?(?:[Ee]xit code:?|exited with code|exit status|"exit_code"\s*:)\s*[1-9]', re.S)
TOKEN_FIELDS = ('input', 'cache_read', 'cache_write', 'output', 'reasoning')


class SkillId:
    """Names and directory slug that identify one skill across runtimes."""

    def __init__(self, raw):
```

## Grill

Your actor is `owner-standin`, and I'll use it in every owner write. I've written nothing yet.

## Answer to 2: my two lines did not serve your aims

Neither line mentions your aims. [N] says "reviewed or named", which is #141's "Done means". [P] checks one cut for a reviewed final candidate within the cap. Nothing in them measures drag on small changes. Nothing in them limits the cost of the review itself, and no line covers a budget the review easily exhausts on the other two cuts. Here is a revised set of three lines, each tied to an aim:

| # | Line | Serves |
|---|---|---|
| 1 | **[N]** On each cut, a run under the candidate rule never stops with a candidate commit that has neither a Result recording a fresh review of its diff nor a name in the stop checkpoint as unreviewed. | the Done means of #141 |
| 2 | **[P]** On each of the three cuts, a run under the candidate rule ends with `hyp status` showing "runner owes nothing" and its final candidate reviewed. It does so within the `--experiments` cap it was prompted with, counted from the cut. | aim (a): a review that doesn't run out of the budget. The version of line 2 I proposed last round checked only run 2's cut. Run 1's cut has one Result left, and run 3's cut is where reviews kept leading to commits. |
| 3 | **[N]** On any cut, a candidate's cost is more than 1.5× the base rule's cost on the same cut. | aim (b): a review that doesn't drag on. Cost is the one measure of drag that doesn't depend on which rival wins. |

- **Why line 3 is a cost line.** I don't use a line like "no fix commit for a passing case". That is the DQ2 rival you agreed should be tested, not a purpose.
- **What line 3 compares.** The base is today's `hyp-goal-run` skill with the default of 5, run on the same cut.
- **How lines 1 and 2 are scored over 3 trials.** This is my pre-registered default:
  - An [N] line (line 1) must hold in all 3 trials.
  - A [P] line (line 2) must hold in at least 2 of 3.
  - You can veto this when you confirm the routing.

## Answer to 4: one question, seven rivals

- **The question:** "Which change to the run practice makes a budget-bound run end with its final candidate reviewed, without running out of budget or dragging on?"
- **Rivals:**
  - H-one-experiment
  - H-name-at-stop
  - H-reserve-last
  - H-raise-budget
  - H-probe-closes
  - H-whole-candidate
  - H-limit-no-commit: from your comment 2, a passing case closes as a limit with no new commit.
- **Base:** today's skill is the cost reference, not a rival.
- **H-whole-candidate:** its verdict rule says the expected outcome is that it loses on line 3. It is still tested.
- **Planted defect:** with only these three lines, the rival's possible benefit is not measured, so I drop the planted-defect scenario. If you want a line for that benefit, it is a new line.
- **Combinations:** the Goal body lets the run add one combination only after the alone results. The combination is H-name-at-stop plus a rival that fails line 1's naming clause and meets lines 2 and 3.
  - Hypotheses added by the run stay inside the Goal and don't need your approval.
  - I don't pre-register all six combinations, because hyp owes every line on each open Hypothesis and that would force 54 more runs.
- **Arms:** seven rivals plus the base is 8 arms. At 3 cuts × 3 trials, that is 72 consumer runs on Sonnet.

## Answer to 5: how cost is computed

- **Eval flow.** I'll use ak-skill-creator's eval flow with Sonnet and 3 trials per arm per cut, replaying from the cuts only.
- **Per trial.** The measured billed USD from the transcript's usage, with its source and pricing date recorded.
  - `references/evaluation-tools.md` forbids summing fields with an assumed billing formula, and a missing metric is never zero.
  - A trial with no cost reading is treated like a hung trial.
- **Per arm and cut.** The median of the 3 trials: sort them and take the second.
- **Hung trial.** A trial that doesn't stop within the pre-registered timeout, or that has no cost reading, counts as +∞. Two hung trials make the median +∞, which fails the 1.5× cap. This follows your rule.
- **Comparison.** A candidate passes line 3 on a cut when `median(candidate) ≤ 1.5 × median(base)` at that cut. If the base median is +∞, that cut is inconclusive for line 3 and the run says so.
- **ak-context-engineering.** It is not installed. It isn't in `.claude/skills`, in `~/.claude/skills`, or in the skill list, so I could not consult it. If you have it, install it under `.claude/skills` and I'll read it before the writes. Meanwhile I'm using ak-skill-creator's `references/evaluation-tools.md` and `scripts/session_usage.py` for cost provenance.

## What I need from you

- Do you confirm the three lines, the 2-of-3 and all-3 rule, the median with a hung trial as +∞, and the 7 rivals?
- Do you want a wall-clock timeout for a hung trial of, say, 30 minutes per trial? I'll write that as a preference unless you name another figure.
- Will you install ak-context-engineering? If not, I'll proceed without it.

After your answers I'll run the second reader and the preflight scan, then show the full routing table and the exact `hyp add goal` command under `owner-standin` for your approval.

## Owner, message 3

Confirmed, with one fix: line 3 as written reads backwards. It says a candidate's cost *is* more than 1.5× the base. It should say that on every cut the candidate's cost is never more than 1.5× the base's cost on that cut. Please reword it that way.

The rest is confirmed:
- The 2-of-3 and all-3 scoring.
- The median, with a hung trial counted as +∞.
- The 7 rivals and the base.
- The 30-minute timeout.

I can't install ak-context-engineering, so go ahead without it. You asked once, which is enough.

Show me the routing table and the exact command under `owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "sha256sum skills/hyp-goal-run/SKILL.md; git rev-parse HEAD; hyp add goal --help 2>&1 | head -20; hyp add dq --help 2>&1 | head; hyp add hyp --help 2>&1 | head; echo $HYP_ACTOR"
}
```

Result:

```text
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
9c8e498b87683e50a94a410b6a3cfecdd7a0d441
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
usage of hyp add dq:
  -actor string
    	actor name (default $HYP_ACTOR)
  -alias string
    	optional unique alias
  -body string
    	body (at most 600 runes)
  -parent string
    	parent node (required)
  -title string
usage of hyp add hyp:
  -actor string
    	actor name (default $HYP_ACTOR)
  -alias string
    	optional unique alias
  -body string
    	body (at most 600 runes)
  -parent string
    	parent node (required)
  -subject value
```

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader of draft Goal",
 "subagent_type": "general-purpose",
 "prompt": "You are an independent reviewer of a draft hyp Goal. Repository: /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt (read only; change nothing). Read AGENTS.md, docs/glossary.md, docs/replay-spec.md (Goal lines, limits, OWED, select rules), docs/sandbox-spec.md (store cut, Arm, patch), skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go, and the store investigations/goal-prompt (`hyp show`/`hyp status` inside it work; do not write to it).\n\nDraft Goal (discovery Goal, store investigations/goal-prompt, third Goal there):\nPurpose: choose a change to the run practice (hyp-goal-run skill text and/or the `hyp prompt` stop clause default) so a budget-bound run on a build Goal ends with its final candidate commit reviewed, without running out of budget and without dragging on over small changes. Source: issue #141.\nOne DesignQuestion with 7 rival Hypotheses (H-one-experiment: a review is part of the gate Experiment; H-name-at-stop; H-reserve-last: with one Result left, review instead of another fix commit; H-raise-budget: default --experiments 7; H-probe-closes: a case the run's own probe passes is closed in the review Result, no limit; H-whole-candidate: review reads cumulative diff from the Goal base; H-limit-no-commit: a passing case closes as a limit with no new commit). The base is today's skill (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) and default 5, at commit 9c8e498; the base is the cost reference, not a rival.\nTest route: ak-skill-creator eval flow; consumer runs on Sonnet, 3 trials per arm per starting point, replayed from three store cuts of investigations/goal-prompt only (after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS), never a whole build from its start. Cost per trial = measured billed USD from the transcript; hung trial (30 min timeout) or missing cost = +infinity; cost per arm and cut = median of 3.\nLines: (1) [N] on each cut, a run under the candidate never stops with a candidate commit that has neither a Result recording a fresh review of its diff nor a name in the stop checkpoint as unreviewed. (2) [P] on each cut, a run under the candidate ends with `hyp status` showing \"runner owes nothing\" and its final candidate reviewed, within the --experiments cap it was prompted with, counted from the cut; scored at least 2 of 3 trials. (3) [N] on every cut the candidate's cost is never more than 1.5x the base's cost on that cut. Lines 1 is scored all 3 trials.\n\nTask: walk each line and the rivals through scenarios the author did NOT draw: two Goals in one store (the cut store holds G-GZR6C6 and G-ZBM2RE), a merged log from two actors, a base that moves mid-run (the skill and prompt change on main), a cut whose actor's earlier Results already count toward the cap, a rival that is a prompt flag change (H-raise-budget) versus skill text, hyp's own OWED/select rules (can a select be owed or admitted for this Goal as
[... 397 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a236fb085a3cd5b06 (internal ID - do not mention to user. Use SendMessage with to: 'a236fb085a3cd5b06', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-b-3-wt/0a25c3cb-af93-470b-b964-61b66b6c44a2/tasks/a236fb085a3cd5b06.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

## Grill

The second reader is running. I'll wait for its report before showing you the table and command.

### Tool ToolSearch (cwd: .)

```text
{
 "query": "select:Monitor",
 "max_results": 1
}
```

Result:

```text

```

## Grill

The second reader is still running. I'm waiting for its report. After that I'll show you the routing table and the exact `hyp add goal` command under `owner-standin`.

## Owner, message 4

<task-notification>
<task-id>a236fb085a3cd5b06</task-id>
<tool-use-id>toolu_01WwU6EufMehEJj5dD8pan7k</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-b-3-wt/0a25c3cb-af93-470b-b964-61b66b6c44a2/tasks/a236fb085a3cd5b06.output</output-file>
<status>completed</status>
<summary>Agent "Second reader of draft Goal" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Concrete problems, each with its evidence. I did not write to the store.

1. **The base commit already holds the finished feature.** The draft pins the base at 9c8e498, and `internal/cli/prompt.go` there already has the merged fixes. The commits that fixed it, 10566d4, b4f23b2 and 5c5902d, are in its history (`git log` on `prompt.go`). A consumer replayed from a cut would be asked to build code that already exists, so the "final candidate commit" is undefined. Each cut needs its own `--base`: the commit under review at that cut. Per R-8JQPWG's observation, that commit is b4f23b2. The draft's single base would break the "base is the cost reference" claim too.

2. **A cut after R-8JQPWG and a cut after R-3YJ7T2 have different owner state.**
   - The owner's answers are the `owner-dat` events dated 2026-11-12T02:2x and 02:3x (`.tree/events/owner-dat.jsonl` lines 3 and 4). R-8JQPWG was recorded 2026-11-11 17:21, so its cut precedes them.
   - There, `hyp status G-ZBM2RE` would show readings under AWAITING OWNER. The runner owes nothing while the owner holds them (the first checkpoint on G-ZBM2RE says "Waiting for the owner").
   - A run can therefore satisfy Line 2 without any review. The draft does not say whether such a stop counts.

3. **The cap in Line 2 is ambiguous.** The prompt's stop clause counts "Results recorded under &lt;actor&gt;" over the whole log (`internal/cli/prompt.go:167`). It does not count from a cut.
   - claude-g113-1 already holds 6 Results, g113-2 holds 5 and g113-3 holds 5. A consumer that reuses one of those actors starts over its cap.
   - The draft must force a fresh actor. "Counted from the cut" then means "from zero", and Line 2's "within the cap it was prompted with" has to name that actor.

4. **H-raise-budget makes Line 2 and Line 3 incomparable across arms.** `--experiments` defaults to 5 (`prompt.go:109`).
   - The prompt is printed by the hyp binary, so the arm is a patch to `prompt.go`. That would break the existing prompt tests, which the Arm's own gate must still pass.
   - An arm prompted with 7 passes Line 2 "within its cap" at a larger cap. It also costs more under Line 3. The cap should be fixed to one number for the Line 2 check, or the line should say it is scored against each arm's own cap.

5. **Line 1 cannot be decided by a check.**
   - "A Result recording a fresh review" has no typed field. Fresh, review and diff live only in free text and in a subagent the skill asks for (`SKILL.md:34`).
   - "A name in the stop checkpoint as unreviewed" is free text. The checkpoint summary is limited to 600 runes (`docs/replay-spec.md:92`) and must also carry the stop clause, the AWAITING OWNER keys and the out-of-Goal findings (`SKILL.md:83-88`). A name may not fit.
   - A hung or timed-out trial has no stop at all, so Line 1 passes it vacuously.
   - The line needs a named grader and a rule, such as a rubric that reads the review's evidence file.

6. **Line 3's cost measure is under-specified.**
   - The review is run by a subagent. The draft does not say whether "billed USD from the transcript" includes subagent usage. If it doesn't, every review rival looks cheap.
   - A median of 3 with two +infinity values is +infinity. If the base's median is +infinity, "≤1.5x base" passes trivially for any candidate. If only the candidate's is, it fails.
   - The base does no extra review, so a review rival may exceed 1.5x by construction. The draft does not say what happens if every review arm fails Line 3.

7. **OWED and select for this Goal.**
   - A Goal's lines are owed on every candidate of its top-level question (`docs/replay-spec.md` §8.1, `limit` row). A line is covered per candidate by a Result that gave a verdict on it, or one whose cited run holds its whole subject (base, patch, files). That means 7 rivals × 3 cuts, each rival tied with `hyp subject &lt;H&gt; --base &lt;commit&gt; --patch FILE`.
   - Two Hypotheses that are skill-text-only or prompt-only still need a patch each. The "H-name-at-stop" rival has no clear patch target, because the stop clause is in `prompt.go:167` and the checkpoint text is in `SKILL.md`.
   - The base is not a Hypothesis. Its trials need an Arm with no patch, or no node. The draft says only that it is a "cost reference".
   - With two or more surviving candidates, hyp owes `choose` to the owner, not `select` to the runner (§8.1, `choose` row). "Select admitted" holds only if exactly one rival survives. The draft should say the owner is expected to choose.

8. **Two Goals in one store.** The cut holds G-GZR6C6 and G-ZBM2RE, and `hyp status` without a Goal prints only the GOALS table (`SKILL.md:13`). The consumer's prompt must be printed with `hyp prompt G-ZBM2RE`, and Line 2's "hyp status" must mean `hyp status G-ZBM2RE`. Otherwise "runner owes nothing" is read from the wrong Goal; both Goals currently show it.

9. **A merged log from two actors.** Event files are per actor (`.tree/events/*.jsonl`), so a cut "after R-X" needs an order. The draft does not say it is by event id (ULID). It should say so, because the Result ids R-8JQPWG, R-3YJ7T2 and R-JHBCWS sit in different actor files.

10. **A base that moves.** The skill and prompt change on main between cuts. H-whole-candidate says "cumulative diff from the Goal base", but the Goal has no recorded base, only per-Hypothesis `hyp subject --base`. The rival is ambiguous unless it names that source.</result>
<usage><subagent_tokens>50565</subagent_tokens><tool_uses>8</tool_uses><duration_ms>58509</duration_ms></usage>
</task-notification>

## Grill

The reader's report is automated and not an answer from you, so nothing is approved yet. I'll check its facts against the source before relying on any of them.

### Tool Bash (cwd: .)

```text
{
 "command": "for c in b4f23b2 5c5902d 2284172 3cfd08d 9c8e498; do git cat-file -t $c 2>&1 | tr '\\n' ' '; done; echo; grep -n \"choose\" docs/replay-spec.md | head -8; grep -n -i \"store cut\\|--store\\|--at\\|cut\" docs/sandbox-spec.md | head -15; cd investigations/goal-prompt && for r in R-8JQPWG R-3YJ7T2 R-JHBCWS R-FRF929 R-KPWZZC; do echo $r; grep -h \"\\\"$r\\|${r#R-}\" .tree/events/*.jsonl | head -1 | cut -c1-160; done; ls .tree/events"
}
```

Result:

```text
commit commit commit commit commit 
90:**[added] readings**: a Result names a line of the Goal that a verdict rests on and how the run read it (`readings`). A reading's key is `<result-id>#r<n>`, and `n` counts the Result's `readings` from 1. Its states are open and closed, and both are derived. An owner answer closes it: its `answers` names the key, and the item's text is the owner's answer, which `hyp show` prints beside the closed reading (an old owner checkpoint closed it with `closes` and its `summary`). An owner changes an answer they gave with a later answer that names the key again: whether an item is a first answer or a revision is derived from the reading's state, so the old `closes` and `revises` merge (ADR 0020). `hyp answer` warns, at exit 0, when it revises a key that a line of another Goal cites. The answers are kept in merge order, the close first, and the latest is the answer in force; `hyp show` prints it beside the reading and each earlier one beneath, with the Result that applied it. A revision never reopens the reading and changes no event before it: an earlier Result keeps its verdicts, and a Hypothesis it contradicted stays contradicted until a new Result judges a successor. In an old log, a `revises` that sorts before its key's close is refused, and a second `closes` of an answered reading is rejected; two revisions both apply, the later in force. A reading of a retracted Result is left out. It refuses, at exit 1, a key that names no reading, a retracted Result, and an answer under the actor that recorded the Result. An open reading prints under AWAITING OWNER and changes no status: it never changes a question's state or the Goal's. It holds back ledger items on its question only while its **reach** is not empty: the live Hypotheses under that question whose standing the answer can move, read from the Result's own verdicts, which are a candidate the Result judged, and a contradicted rival whose only `contradicts` is that Result's. While the r
[... 6662 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "mkdir -p $TMPDIR/g141 && cd $TMPDIR/g141 && cat > build.py <<'EOF'\nimport json\ntitle=\"Run practice: a budget-bound build run ends with its final candidate reviewed\"\nbody_head=(\"Chooses one change to the run practice of a build Goal (the Build Goals text of hyp-goal-run and the stop clause hyp prompt prints) so a budget-bound run ends with its final candidate reviewed, does not run out of budget on reviews, and does not drag on over small changes. Issue #141, part of #113. Each rival is a throwaway Arm patch of skills/hyp-goal-run/SKILL.md or internal/cli/prompt.go, never a landed change. The base is today's skill (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) with the default of 5; it is the cost reference, not a rival. The owner adds Goal B, a build Goal for the chosen rule, after the choose.\\n\"\n\"Test: ak-skill-creator's eval flow, Sonnet, 3 trials per arm per cut, replayed from three cuts of this store and never a whole build. A cut ends at the event next after R-8JQPWG (base b4f23b2), R-3YJ7T2 (base 2284172) or R-JHBCWS (base 3cfd08d), ordered by event id. Each trial uses a fresh actor, the prompt from hyp prompt G-ZBM2RE, and a cap counted from zero. The replay recipe is a self-review instrument: base on the first cut must first reproduce run 1's unreviewed stop. Cost per trial is the billed USD of the transcript with its subagents, with source and pricing date; a trial that does not stop in 30 minutes, or has no cost reading, is +infinity; per arm and cut it is the median of 3. A base median of +infinity makes that cut inconclusive for line 3. Scoring: line 1 in all 3 trials, line 2 in 2 of 3. After the alone results the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3. With two or more survivors the owner chooses.\\n\")\nblock=(\"May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; the run may push that branch and open or update one PR, and never merges it. Arms are patches in the sandbox only, so no text under skills/, cmd/, internal/, docs/, AGENTS.md or README.md changes.\\n\"\n\"Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/, and ak-skill-creator under .claude/skills, pinned by sha256.\\n\"\n\"Never: commit to main or merge a PR.\\n\"\n\"Never: replace ~/.local/bin/hyp or any installed binary; build an arm only inside its sandbox.\\n\"\n\"Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.\\n\"\n\"Never: edit another investigations/ tree, ak-skill-creator or another repository.\")\nbody=body_head+block\nL=[\n\"[N] On each cut, a run under the candidate rule never stops, a 30-minute timeout included, with a candidate commit that has neither a Result citing a fresh subagent's review of that commit's diff nor its short hash named as un
[... 3223 more characters]
```

Result:

```text
77 2478 body runes (<=2000)
300 line runes (<=600)
380 line runes (<=600)
418 line runes (<=600)
261 H-one-experiment
207 H-name-at-stop
210 H-reserve-last
145 H-raise-budget
218 H-probe-closes
223 H-whole-candidate
187 H-limit-no-commit
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt
```

## Grill

The body is 2478 runes against a 2000 limit, so I'm tightening it.

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && python3 -I - <<'EOF'\nimport json\nd=json.load(open(\"data.json\"))\nhead=(\"Chooses one change to the run practice of a build Goal (hyp-goal-run's Build Goals text and the stop clause hyp prompt prints) so a budget-bound run ends with its final candidate reviewed, without running out of budget or dragging on over small changes. Issue #141, part of #113. A rival is a throwaway Arm patch of skills/hyp-goal-run/SKILL.md or internal/cli/prompt.go, never a landed change. The base is today's skill and the default of 5: the cost reference, not a rival. The owner adds Goal B, the build of the chosen rule, after the choose.\\n\"\n\"Test: ak-skill-creator's eval flow, Sonnet, 3 trials per arm per cut, replayed from three cuts of this store, never a whole build. A cut ends at the event next after R-8JQPWG (base b4f23b2), R-3YJ7T2 (base 2284172) or R-JHBCWS (base 3cfd08d), ordered by event id. Each trial has a fresh actor, the prompt of hyp prompt G-ZBM2RE and a cap counted from zero. The replay recipe is a self-review instrument: the base on the first cut must first reproduce run 1's unreviewed stop. Cost is the median of 3 trials per arm and cut; a base median of +infinity makes the cut inconclusive for line 3. Line 1 must hold in all 3 trials, line 2 in 2 of 3. After the alone results the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3. With two or more survivors the owner chooses.\\n\")\nblock=(\"May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; the run may push that branch and open or update one PR, and never merges it. Arms are sandbox patches, so no text under skills/, cmd/, internal/ or docs/ changes.\\n\"\n\"Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/, and ak-skill-creator under .claude/skills, pinned by sha256.\\n\"\n\"Never: commit to main or merge a PR.\\n\"\n\"Never: replace an installed binary; build an arm only inside its sandbox.\\n\"\n\"Never: write under an owner's actor; refer with a reason or record a reading, and stop.\\n\"\n\"Never: edit another investigations/ tree, ak-skill-creator or another repository.\")\nd[\"body\"]=head+block\nprint(len(d[\"body\"]))\njson.dump(d,open(\"data.json\",\"w\"))\nEOF"
}
```

Result:

```text
2113
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && python3 -I - <<'EOF'\nimport json\nd=json.load(open(\"data.json\"))\nb=d[\"body\"]\nfor a,c in [(\" Issue #141, part of #113.\",\" Issue #141.\"),\n (\"; the run may push that branch and open or update one PR, and never merges it\",\"; it never merges a PR\"),\n (\" The owner adds Goal B, the build of the chosen rule, after the choose.\",\" The owner adds Goal B, the build of the chosen rule, later.\"),\n (\"ordered by event id. \",\"by event id. \"),\n (\"Arms are sandbox patches, so no text under skills/, cmd/, internal/ or docs/ changes.\",\"Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.\"),\n (\"Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/, and ak\",\"Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak\")]:\n    assert a in b; b=b.replace(a,c)\nd[\"body\"]=b; print(len(b))\njson.dump(d,open(\"data.json\",\"w\"))\nEOF"
}
```

Result:

```text
2021
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && python3 -I - <<'EOF'\nimport json\nd=json.load(open(\"data.json\"))\nb=d[\"body\"]\nfor a,c in [(\" (hyp-goal-run's Build Goals text and the stop clause hyp prompt prints)\",\" (hyp-goal-run's Build Goals and hyp prompt's stop clause)\"),\n (\"Test: ak-skill-creator's eval flow, Sonnet,\",\"Test: ak-skill-creator eval flow, Sonnet,\"),\n (\"; never a whole build\",\", never a whole build\") if False else (\"The base is today's skill and the default of 5: the cost reference, not a rival.\",\"The base, today's skill with the default of 5, is the cost reference, not a rival.\")]:\n    assert a in b; b=b.replace(a,c)\nd[\"body\"]=b; print(len(b))\njson.dump(d,open(\"data.json\",\"w\"))\n# emit command\nimport shlex\nL=d[\"lines\"]\ncmd=\"HYP_ACTOR=owner-standin hyp add goal --title %s --body %s\"%(shlex.quote(d[\"title\"]),shlex.quote(b))\nfor l in L: cmd+=\" --line \"+shlex.quote(l)\ncmd+=\" --why \"+shlex.quote(\"Owner, 2026-11-12: the goal of issue #141, confirmed in the grill session; lines 1-3 serve the aims: no unreviewed final candidate, no budget run-out, no drag.\")\nopen(\"goal.cmd\",\"w\").write(cmd)\nprint(len(cmd))\nEOF\npython3 -I -c \"print(len('Owner, 2026-11-12: the goal of issue #141, confirmed in the grill session; lines 1-3 serve the aims: no unreviewed final candidate, no budget run-out, no drag.'))\""
}
```

Result:

```text
2007
3489
159
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && python3 -I - <<'EOF'\nimport json,shlex\nd=json.load(open(\"data.json\"))\nb=d[\"body\"].replace(\"A rival is a throwaway Arm patch\",\"A rival is an Arm patch\")\nd[\"body\"]=b; print(len(b)); assert len(b)<=2000\njson.dump(d,open(\"data.json\",\"w\"))\ncmd=\"HYP_ACTOR=owner-standin hyp add goal --title %s --body %s\"%(shlex.quote(d[\"title\"]),shlex.quote(b))\nfor l in d[\"lines\"]: cmd+=\" --line \"+shlex.quote(l)\ncmd+=\" --why \"+shlex.quote(\"Owner, 2026-11-12: the goal of issue #141, confirmed in the grill session; lines 1-3 serve the aims: no unreviewed final candidate, no budget run-out, no drag.\")\nopen(\"goal.cmd\",\"w\").write(cmd)\nprint(b); print(); [print(l,\"\\n\") for l in d[\"lines\"]]\nEOF"
}
```

Result:

```text
1998
Chooses one change to the run practice of a build Goal (hyp-goal-run's Build Goals and hyp prompt's stop clause) so a budget-bound run ends with its final candidate reviewed, without running out of budget or dragging on over small changes. Issue #141. A rival is an Arm patch of skills/hyp-goal-run/SKILL.md or internal/cli/prompt.go, never a landed change. The base, today's skill with the default of 5, is the cost reference, not a rival. The owner adds Goal B, the build of the chosen rule, later.
Test: ak-skill-creator eval flow, Sonnet, 3 trials per arm per cut, replayed from three cuts of this store, never a whole build. A cut ends at the event next after R-8JQPWG (base b4f23b2), R-3YJ7T2 (base 2284172) or R-JHBCWS (base 3cfd08d), by event id. Each trial has a fresh actor, the prompt of hyp prompt G-ZBM2RE and a cap counted from zero. The replay recipe is a self-review instrument: the base on the first cut must first reproduce run 1's unreviewed stop. Cost is the median of 3 trials per arm and cut; a base median of +infinity makes the cut inconclusive for line 3. Line 1 must hold in all 3 trials, line 2 in 2 of 3. After the alone results the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3. With two or more survivors the owner chooses.
May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; it never merges a PR. Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.
Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace an installed binary; build an arm only inside its sandbox.
Never: write under an owner's actor; refer with a reason or record a reading, and stop.
Never: edit another investigations/ tree, ak-skill-creator or another reposito
[... 1175 more characters]
```

## Grill

The routing table and the exact command are below. I've run nothing against the store. The command waits for your approval.

## Changes since your last answer
The second reader's findings were checked against the source. These changes go into the Goal text:
- **Line 3:** reworded as you asked. It now reads "never more than 1.5 times the base's cost on that cut". It also names the base by sha256, as the rules require for a comparison.
- **Line 1:** a trial that never stops fails it. "Reviewed" means a Result citing a fresh subagent's review of that commit's diff. "Named" means the commit's short hash is marked unreviewed in the stop checkpoint.
- **Line 2:** it no longer demands "owes nothing". On the first cut, readings R-KPWZZC and R-50GBXM are still open, because your answers came later. Line 2 therefore allows a select that an open reading holds back.
- **Cuts:** each has its own base: b4f23b2, 2284172 and 3cfd08d. Cuts are ordered by event id.
- **Trial setup:** each trial uses a fresh actor, with the cap counted from zero.
- **Replay recipe:** it is a self-review instrument. Before any verdict counts, the base on the first cut must reproduce run 1's unreviewed stop.
- **Choose:** with two or more survivors, hyp owes the owner a `choose`, not the runner a `select`.

## Routing table

| Grill question | Place | Separating scenario |
|---|---|---|
| One Goal or two | Goal A now (discovery). Goal B (the build) is yours to add later. | none |
| Reviewed or named at the stop | **Line 1 [N]**: no unreviewed, unnamed final candidate on any cut | none |
| Reviewed final candidate within the budget (aim a) | **Line 2 [P]** | none |
| Cost never above 1.5× the base (aim b) | **Line 3 [N]**, with the base named by sha256 | none |
| 7 rivals: H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate, H-limit-no-commit | **One DesignQuestion** with 7 rivals | Replay the three cuts, 3 Sonnet trials per arm, scored on the three lines. H-whole-candidate is the expected loser on line 3 and is tested anyway. |
| Combinations | Goal body: the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3 | after the alone results |
| Test method, model, trials, cost, timeout, scoring | Goal body | none |
| Constraints | Goal body block | none |
| Plain facts (store, scope, premise) | looked up | none |
| Goal B, the build of the chosen rule | another Goal, which you add | none |

## Command 1: your write, under `owner-standin`
```sh
HYP_ACTOR=owner-standin hyp add goal \
  --title 'Run practice: a budget-bound build run ends with its final candidate reviewed' \
  --body '<body below, 1998 runes>' \
  --line '<line 1>' --line '<line 2>' --line '<line 3>' \
  --why 'Owner, 2026-11-12: the goal of issue #141, confirmed in the grill session; lines 1-3 serve the aims: no unreviewed final candidate, no budget run-out, no drag.'
```
I run it from `investigations/goal-prompt`. The full command, with the body and lines in place, is in `$TMPDIR/g141/goal.cmd`.

**Body**
```text
Chooses one change to the run practice of a build Goal (hyp-goal-run's Build Goals and hyp prompt's stop clause) so a budget-bound run ends with its final candidate reviewed, without running out of budget or dragging on over small changes. Issue #141. A rival is an Arm patch of skills/hyp-goal-run/SKILL.md or internal/cli/prompt.go, never a landed change. The base, today's skill with the default of 5, is the cost reference, not a rival. The owner adds Goal B, the build of the chosen rule, later.
Test: ak-skill-creator eval flow, Sonnet, 3 trials per arm per cut, replayed from three cuts of this store, never a whole build. A cut ends at the event next after R-8JQPWG (base b4f23b2), R-3YJ7T2 (base 2284172) or R-JHBCWS (base 3cfd08d), by event id. Each trial has a fresh actor, the prompt of hyp prompt G-ZBM2RE and a cap counted from zero. The replay recipe is a self-review instrument: the base on the first cut must first reproduce run 1's unreviewed stop. Cost is the median of 3 trials per arm and cut; a base median of +infinity makes the cut inconclusive for line 3. Line 1 must hold in all 3 trials, line 2 in 2 of 3. After the alone results the run may add H-name-at-stop with one rival that fails only line 1's naming clause and meets lines 2 and 3. With two or more survivors the owner chooses.
May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; it never merges a PR. Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.
Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace an installed binary; build an arm only inside its sandbox.
Never: write under an owner's actor; refer with a reason or record a reading, and stop.
Never: edit another investigations/ tree, ak-skill-creator or another repository.
```

**Lines**
1. **[N]** On each cut, a run under the candidate rule never stops, a 30-minute timeout included, with a candidate commit that has neither a Result citing a fresh subagent's review of that commit's diff nor its short hash named as unreviewed in the stop checkpoint. A trial that never stops fails this line.
2. **[P]** On each cut, a run under the candidate rule stops with its final candidate reviewed (a Result citing a fresh subagent's review of that commit's diff) and with nothing in OWED of `hyp status G-ZBM2RE` for the runner, except a select that an open reading of the cut holds back. It does so within the `--experiments` cap its own prompt printed, counted from zero under a fresh actor.
3. **[N]** On every cut, the candidate's cost is never more than 1.5 times the base's cost on that cut. The base is `skills/hyp-goal-run/SKILL.md` with sha256 `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e` and the default of 5; the candidate's patch differs from it. Cost is the median over 3 trials of billed USD with subagents; a trial that does not stop in 30 minutes, or has no cost reading, is +infinity.

## After your approval, my writes as the runner
I use my own runner actor for these, in order:
1. `hyp add dq --parent <G-id> …`
2. Seven `hyp add hyp --parent <Q-id> …` commands, one per rival, with their bodies drafted in `$TMPDIR/g141/data.json`.
3. `hyp link <G-id> external_ref https://github.com/dat9uy/hyp/issues/141`
4. `hyp status <G-id>`, then `hyp check`, which must print `check: 0 errors`.
5. `hyp checkpoint --goal <G-id> --summary … --why "end of grill"`
6. A commit of `.tree/events/` and `.tree/evidence/`.
7. `hyp prompt <G-id> --actor <new runner actor>`, handed to you unchanged.

**Do you approve command 1 exactly as shown?** You can also change a line, the body or a constraint first.

## Owner, message 5

Approved: command 1 exactly as shown, under `owner-standin`. The routing table matches what I confirmed, so I have nothing to move. Go ahead with your runner writes. Do not push or open a PR in this session.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-3/wt/investigations/goal-prompt && bash -c \"$(cat $TMPDIR/g141/goal.cmd)\" 2>&1 | tail -15"
}
```

Result:

```text
G-9T240G node.add
  line G-9T240G#1: [N] On each cut, a run under the candidate rule never stops, a 30-minute timeout included, with a candidate commit that has neither a Result citing a fresh subagent's review of that commit's diff nor its short hash named as unreviewed in the stop checkpoint. A trial that never stops fails this line.
  line G-9T240G#2: [P] On each cut, a run under the candidate rule stops with its final candidate reviewed (a Result citing a fresh subagent's review of that commit's diff) and with nothing in OWED of hyp status G-ZBM2RE for the runner, except a select that an open reading of the cut holds back. It does so within the --experiments cap its own prompt printed, counted from zero under a fresh actor.
  line G-9T240G#3: [N] On every cut, the candidate's cost is never more than 1.5 times the base's cost on that cut. The base is skills/hyp-goal-run/SKILL.md with sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e and the default of 5; the candidate's patch differs from it. Cost is the median over 3 trials of billed USD with subagents; a trial that does not stop in 30 minutes, or has no cost reading, is +infinity.
```

## Grill

The Goal is written as G-9T240G. Now the runner writes. First I'll check the flags.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp add hyp --help 2>&1 | tail -12; hyp add dq --help 2>&1 | tail -6"
}
```

Result:

```text
  -alias string
    	optional unique alias
  -body string
    	body (at most 600 runes)
  -parent string
    	parent node (required)
  -subject value
    	a file or directory that embodies the claim; hyp ties the Hypothesis to each file's sha256 (repeatable)
  -title string
    	title (required, at most 120 runes)
  -why string
    	why this event is recorded (required, 1–300 runes)
  -parent string
    	parent node (required)
  -title string
    	title (required, at most 120 runes)
  -why string
    	why this event is recorded (required, 1–300 runes)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && python3 -I - <<'EOF'\nimport json,subprocess\nd=json.load(open(__import__(\"os\").environ[\"TMPDIR\"]+\"/g141/data.json\"))\ndef run(a):\n    r=subprocess.run(a,capture_output=True,text=True); print(\" \".join(a[:4]),\"->\",r.returncode,r.stdout.strip()[:200],r.stderr.strip()[:300]); return r\nq=d[\"dq\"]\nr=run([\"hyp\",\"add\",\"dq\",\"--parent\",\"G-9T240G\",\"--title\",q[0][:120],\"--body\",q[1],\"--why\",\"The one question of issue #141; every rival answers it.\"])\nEOF\nhyp status G-9T240G 2>&1 | head -12"
}
```

Result:

```text
hyp add dq --parent -> 0 Q-R9KP1Q node.add 
GOAL G-9T240G Run practice: a budget-bound build run ends with its final … — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-9T240G#1 no candidate — [N] On each cut, a run under the candidate rule never stops, a 30-minute timeout included, with a candidate commit that has neither a…
  G-9T240G#2 no candidate — [P] On each cut, a run under the candidate rule stops with its final candidate reviewed (a Result citing a fresh subagent's review of…
  G-9T240G#3 no candidate — [N] On every cut, the candidate's cost is never more than 1.5 times the base's cost on that cut. The base is skills/hyp-goal-run/SKIL…
FRONTIER
  Q-R9KP1Q open/no_candidates Which change to the run practice makes a budget-bound build…
STALE
OWED
  runner: Q-R9KP1Q add rival Hypotheses
AWAITING OWNER
RECENT
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && python3 -I - <<'EOF'\nimport json,subprocess,os\nd=json.load(open(os.environ[\"TMPDIR\"]+\"/g141/data.json\"))\nfor name,desc,patch in d[\"H\"]:\n    t=name+\": \"+desc\n    r=subprocess.run([\"hyp\",\"add\",\"hyp\",\"--parent\",\"Q-R9KP1Q\",\"--alias\",name.lower(),\"--title\",t[:120],\"--body\",desc+\" \"+patch,\"--why\",\"A rival of issue #141, as the issue and its owner comments state it.\"],capture_output=True,text=True)\n    print(name,r.returncode,r.stdout.strip()[:60],r.stderr.strip()[:200])\nEOF\nhyp link G-9T240G external_ref https://github.com/dat9uy/hyp/issues/141 2>&1 | tail -2; hyp status G-9T240G 2>&1 | sed -n 5,30p; hyp check 2>&1 | tail -5"
}
```

Result:

```text
H-one-experiment 0 H-9D8HD1 node.add 
H-name-at-stop 0 H-BY60MR node.add 
H-reserve-last 0 H-9FKT38 node.add 
H-raise-budget 0 H-6YV91H node.add 
H-probe-closes 0 H-K9DR79 node.add 
H-whole-candidate 0 H-VCECYB node.add 
H-limit-no-commit 0 H-YQE1G2 node.add 
missing --why
  G-9T240G#3 no candidate — [N] On every cut, the candidate's cost is never more than 1.5 times the base's cost on that cut. The base is skills/hyp-goal-run/SKIL…
FRONTIER
  Q-R9KP1Q open/untested Which change to the run practice makes a budget-bound build…
    H-9D8HD1 (h-one-experiment) proposed H-one-experiment: A candidate's review is part of its gate …
    H-BY60MR (h-name-at-stop) proposed H-name-at-stop: Keep the practice; the stop checkpoint name…
    H-9FKT38 (h-reserve-last) proposed H-reserve-last: With one Result left, the runner reviews th…
    H-6YV91H (h-raise-budget) proposed H-raise-budget: The default of --experiments rises from 5 t…
    H-K9DR79 (h-probe-closes) proposed H-probe-closes: A case the run's own probe runs and passes …
    H-VCECYB (h-whole-candidate) proposed H-whole-candidate: The review before the select reads the c…
    H-YQE1G2 (h-limit-no-commit) proposed H-limit-no-commit: A case whose check already passes closes…
STALE
OWED
  runner: Q-R9KP1Q test H-9D8HD1 (h-one-experiment)
  runner: Q-R9KP1Q test H-BY60MR (h-name-at-stop)
  runner: Q-R9KP1Q test H-9FKT38 (h-reserve-last)
  runner: Q-R9KP1Q test H-6YV91H (h-raise-budget)
  runner: Q-R9KP1Q test H-K9DR79 (h-probe-closes)
  runner: Q-R9KP1Q test H-VCECYB (h-whole-candidate)
  runner: Q-R9KP1Q test H-YQE1G2 (h-limit-no-commit)
AWAITING OWNER
RECENT
CHECKPOINT none
warn: uncontained R-BSWFXG (script outside recipes/: .tree/evidence/E-ZMMVPN/probe-p1.sh)
warn: uncontained R-HXEBAF (script outside recipes/: .tree/evidence/E-ZG1BC8/probe-q1.sh)
warn: uncited R-FRF929: 2aeb0abe is in the observation but in no cited file
warn: uncited R-BSWFXG: 2aeb0abe is in the observation but in no cited file
check: 0 errors, 9 wa
[... 6 more characters]
```

## Grill

The link needs a `--why`. I'll add it, then write the checkpoint.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && hyp link G-9T240G external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this Goal serves.\" 2>&1 | tail -2; hyp checkpoint --goal G-9T240G --summary \"Grill done: 3 lines written by owner-standin; Q-R9KP1Q holds 7 rivals, none tested. Goal B (the build of the chosen rule) is for the owner to add after the choose. Out-of-Goal: none. The replay recipe is a self-review instrument and must first reproduce run 1's unreviewed stop.\" --why \"end of grill\" 2>&1 | tail -2; hyp check 2>&1 | tail -1; hyp status G-9T240G | grep -i \"checkpoint\"; cd ../.. && git status --short | head"
}
```

Result:

```text
G-9T240G edge.add
G-9T240G checkpoint
check: 0 errors, 9 warnings
    H-BY60MR (h-name-at-stop) proposed H-name-at-stop: Keep the practice; the stop checkpoint name…
CHECKPOINT 2026-11-12 14:52 claude-grill-141: Grill done: 3 lines written by owner-standin; Q-R9KP1Q holds 7 rivals, none tested. Goal B (the build of the chosen rule) is for the owner to add after the choose. Out-of-Goal: none. The replay recip… (full: hyp show G-9T240G)
?? investigations/goal-prompt/.tree/events/claude-grill-141.jsonl
?? investigations/goal-prompt/.tree/events/owner-standin.jsonl
```

### Tool Bash (cwd: .)

```text
{
 "command": "git add investigations/goal-prompt/.tree && git commit -q -m \"docs: add the Goal and rivals for the budget-bound review practice of #141\" && git log --oneline -1 && cd investigations/goal-prompt && hyp prompt G-9T240G --actor claude-g141-1 --branch work"
}
```

Result:

```text
26dc196 docs: add the Goal and rivals for the budget-bound review practice of #141
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-9T240G: read it with hyp show G-9T240G, read hyp status G-9T240G, and name --goal G-9T240G on every hyp checkpoint.
Done when the latest printed hyp status G-9T240G shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; it never merges a PR. Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.
Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace an installed binary; build an arm only inside its sandbox.
Never: write under an owner's actor; refer with a reason or record a reading, and stop.
Never: edit another investigations/ tree, ak-skill-creator or another repository.
Commit on the branch work.
```

## Grill

The grill is done. Goal G-9T240G is in `investigations/goal-prompt` with its three lines, one question (Q-R9KP1Q) and seven untested rivals. `hyp check` prints 0 errors, but with 9 warnings. Two I read are about run scripts under `.tree/evidence/` outside `recipes/` and an uncited hash in an old Result. I did not look at the other seven.

I committed the events as `26dc196` on branch `work`. I did not push or open a PR.

**What was written**
- **Goal G-9T240G** was written under `owner-standin`, with the command you approved.
- **Lines:**
  - #1 [N]: no unreviewed, unnamed final candidate.
  - #2 [P]: a reviewed final candidate within the cap.
  - #3 [N]: cost never above 1.5× the base.
- **Question Q-R9KP1Q:** which change to the run practice makes a budget-bound build run end with its final candidate reviewed, without running out of budget or dragging on over small changes. The runner writes under `claude-grill-141`.
- **Rivals:**
  - H-9D8HD1 (one-experiment)
  - H-BY60MR (name-at-stop)
  - H-9FKT38 (reserve-last)
  - H-6YV91H (raise-budget)
  - H-K9DR79 (probe-closes)
  - H-VCECYB (whole-candidate)
  - H-YQE1G2 (limit-no-commit)
- **Link and checkpoint:** the Goal is linked to issue #141 and ends with the "end of grill" checkpoint.

The DesignQuestion's title was cut at 120 runes, and events can't be edited. It reads fine but ends mid-phrase.

**Hand-off prompt**
Here is the output of `hyp prompt`, unchanged. It uses runner actor `claude-g141-1` and branch `work`:

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-9T240G: read it with hyp show G-9T240G, read hyp status G-9T240G, and name --goal G-9T240G on every hyp checkpoint.
Done when the latest printed hyp status G-9T240G shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: the investigations/goal-prompt store, in commits on one local branch, and a new recipe version under recipes/ with its rubric under rubrics/; it never merges a PR. Arms are sandbox patches: no text under skills/, cmd/, internal/ or docs/ changes.
Read only: plans/, every other investigations/ tree, skills/, cmd/, internal/, docs/ and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: replace an installed binary; build an arm only inside its sandbox.
Never: write under an owner's actor; refer with a reason or record a reading, and stop.
Never: edit another investigations/ tree, ak-skill-creator or another repository.
Commit on the branch work.
```

**Things for you to know**
- **Cap:** the default cap of 5 Results is likely too small for this Goal, which needs a recipe, 8 arms and 3 cuts. Pass `--experiments N` to `hyp prompt` if you want a larger one. I left it at the default. I think the run will stop at the cap with lines still owed. Under the skill, that is a normal result.
- **Replay recipe:** the run has to write it as a new version under `recipes/`. It is a self-review instrument, so it must first reproduce run 1's unreviewed stop on the first cut before any verdict counts.
- **Goal B:** you add the build Goal for the chosen rule after the choose. If two or more rivals survive, hyp owes you the choose.
- **ak-context-engineering:** not installed, so I proceeded without it.
