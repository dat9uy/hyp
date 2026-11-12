# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-entry-b-1/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and finding the stores.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-entry-b-1/tmp/issue-141.md"
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
 "command": "ls investigations/ && for d in investigations/*/; do echo \"== $d\"; (cd $d && hyp status 2>&1 | head -15); done",
 "description": "List investigation stores and their status"
}
```

Result:

```text
checkpoint-store
checkpoint-store-design
goal-prompt
id-citation
recipe-base
== investigations/checkpoint-store-design/
GOAL G-73R704 Record the shared-store design: what hyp stores and derives… — complete (answered 4, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
RECENT
  2026-11-11 09:11 claude-g1-1 dq.select Q-R21S8R — Owed select, the only candidate; the owner's decision in session e6a35a24 (2026…
  2026-11-11 09:11 claude-g1-1 dq.select Q-H8PJ20 — Owed select, the only candidate; the owner's decision in session e6a35a24 (2026…
  2026-11-11 09:11 claude-g1-1 dq.select Q-HC4V9T — Owed select, the only candidate; the owner's decision in session e6a35a24 (2026…
  2026-11-11 09:11 claude-g1-1 dq.select Q-D906CD — Owed select, the only candidate; the owner's decision in session e6a35a24 (2026…
  2026-11-11 09:10 claude-g1-1 result.record R-AWPQ4X — Session decision, recorded as the Experiment's Result.
  2026-11-11 09:10 claude-g1-1 result.record R-8CJCV6 — Session decision, recorded as the Experiment's Result.
  2026-11-11 09:10 claude-g1-1 result.record R-V8032Z — Session decision, recorded as the Experiment's Result.
  2026-11-11 09:10 claude-g1-1 result.record R-HW9B05 — Session decision, recorded as the Experiment's Result.
== investigations/checkpoint-store/
GOAL G-D1CH1R hyp implements the shared store: one store holds many Goals — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-D1CH1R#1 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [P] A sto…
  G-D1CH1R#2 covered on Q-PC04ED; ran in R-Q7ZZW8, R-M2PZ8N, R-872FV1, R-YWCB7C, R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] In a …
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW141, R-P1M8A8, R-RRYBPH — [N] A Result keeps one su
[... 3641 more characters]
```

## Grill

The issue is part of #113, whose Goals are in `investigations/goal-prompt`, so the new Goal goes in that store. Next I'll read the texts the change touches.

### Tool Bash (cwd: .)

```text
{
 "command": "cat skills/hyp-goal-run/SKILL.md && git log --oneline -5 -- skills/hyp-goal-run && grep -rn \"experiments\\|Results recorded\" internal/cli/prompt*.go | head -40",
 "description": "Read goal-run skill and prompt code"
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
[... 13307 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/investigations/goal-prompt && hyp status G-ZBM2RE 2>&1 | head -60; hyp show G-ZBM2RE 2>&1 | head -80",
 "description": "Show the existing goal-prompt Goal"
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

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt && hyp status G-GZR6C6 -h >/dev/null 2>&1; (cd investigations/goal-prompt && hyp show G-GZR6C6 | head -60; hyp show R-8JQPWG 2>&1 | head -20; hyp show R-JHBCWS 2>&1 | head -30); grep -n \"Verifier\\|Store cut\\|store cut\\|Runner\\b\" docs/glossary.md | head; grep -n \"limit\" docs/glossary.md | head; hyp record -h 2>&1 | head -40; hyp add exp -h 2>&1 | head -30",
 "description": "Inspect precedent design Goal, cited results, glossary, record flags"
}
```

Result:

```text
Goal G-GZR6C6
alias: design
title: Record how hyp makes the /goal prompt: who writes it, and whether it carries owner answers
body: Design record for issue #113, settled by the owner in the grill session of 2026-11-11 (hyp-owner-grill). Two questions: who writes the /goal prompt for a Goal, hyp or a skill; and whether the prompt carries the owner's answers. Each is a DesignQuestion with its rivals, decided in that session with the session as evidence, so the tree says it was argued, not tested. This Goal has no lines and no run: no run checks its candidates. Its cases are the lines of the build Goal in this store that cites its selects.
path: G-GZR6C6
status: complete
edge out: external_ref https://github.com/dat9uy/hyp/issues/113
children:
  Q-NCY8HY (q-who) answered Who writes the /goal prompt for a Goal: hyp from the log, o…
  Q-MAKJ8R (q-answers) answered Does the /goal prompt carry the owner's answers?
history:
  2026-11-11 16:45 owner-dat node.add G-GZR6C6 — Owner, 2026-11-11: the design record for #113's prompt generator; its build is a second Goal in this store.
  2026-11-11 16:47 claude-grill-113 edge.add G-GZR6C6 — The issue whose design questions this record settles.
  2026-11-11 16:47 claude-grill-113 checkpoint G-GZR6C6 — end of grill
    summary: Design record complete: Q-NCY8HY selects H-MCDPNB (hyp prints the prompt) and Q-MAKJ8R selects H-RSDX0H (no answers clause), with grill session 285fa63e as evidence. G-ZBM2RE builds them.
  2026-11-11 16:48 claude-grill-113 edge.remove G-GZR6C6 — Wrong repository in the URL; the remote is dat9uy/hyp.
  2026-11-11 16:48 claude-grill-113 edge.add G-GZR6C6 — Issue #113 at the repository's real remote.
Result R-8JQPWG
path: G-ZBM2RE > Q-Z93VJ5 > H-3M3QAS > E-7QQ2DP > R-8JQPWG
status: recorded
observation: The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and flags, exit 3 through a symlinked working directory and exit 0 with 'in inv/t' through the ph
[... 10149 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 100,200p internal/cli/prompt.go; grep -n \"run cap\\|cap\\b\\|Experiments recorded\\|no run cap\" docs/how-hyp-improves-itself.md | head; grep -rn \"experiments\\|5 Results\\|stop clause\" skills/*/SKILL.md | head -20",
 "description": "Read prompt stop clause code and related texts"
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
[... 3064 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 20,40p docs/how-hyp-improves-itself.md; sed -n 397,400p docs/glossary.md; grep -n \"cut\" docs/sandbox-spec.md | head -12; ls .claude/skills/ak-skill-creator | head -20; ls investigations/goal-prompt/.tree/evidence | head -40; hyp show H-YTD2JG 2>&1 | head -5; (cd investigations/goal-prompt && hyp show H-YTD2JG | head -6)",
 "description": "Preflight lookups: cap text, store cut, eval flow, evidence"
}
```

Result:

```text
the run goes well:

- **Write-time refusals.** The binary refuses events that would corrupt the
  tree or skip a step, for example an experiment that ignores live rivals
  without saying so, or a select the ledger does not owe. The list is
  owned by [replay-spec.md](replay-spec.md) §5.
- **The ledger.** `hyp status` lists under OWED what the tree owes and who
  owes it, and marks an item an Experiment aimed at and did not settle
  ([replay-spec.md](replay-spec.md) §8.1). It is the one account of progress
  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
  stated by the `/goal` prompt or the recipe, not by hyp.
- **Design marks.** `hyp signals` lists the Experiments set up against no live
  rival (one-rival) or against a dead one (D1). A Result that leaves a
  target without a verdict is refused, not marked (ADR 0018).
- **The select is the gate.** `hyp select` is admitted only when the ledger
  owes it, and the owner decides trade-offs
  ([ADR 0015](adr/0015-the-select-is-the-gate.md)). What the owner decides,
  and what they never do, is in [principles.md](principles.md#7-who-does-what).

A run that reaches a select the ledger owes, with no refusal it had to work around, is a success.
It needs no further study.
**Store cut**:
A copy of a tree's store as it stood at one event, given to a sandbox.
_Avoid_: snapshot (that is a different thing: a capture of a live session)

46:- No shell runs a command. Its first word is a relative path containing `/` to an executable script inside the bundle, and the rest are plain words; quotes, `$`, pipes, redirects and globs are refused.
158:A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (see Arm in [glossary.md](glossary.md)), given as `--patch ARM=FILE`.
assets
LICENSE-MIT.txt
LICENSE.txt
NOTICE.txt
references
scripts
SKILL.md
E-2A0909
E-7QQ
[... 685 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/investigations/goal-prompt && ls .tree/events | head -3; grep -rhE '\"type\":\"result.record\"' .tree/events 2>/dev/null | head -2 | cut -c1-300",
 "description": "Peek at event file format"
}
```

Result:

```text
claude-g113-1.jsonl
claude-g113-2.jsonl
claude-g113-3.jsonl
{"v":1,"id":"01M75S97KWYDNTGA2N56KPWZZC","ts":"2026-11-11T17:07:57.180003Z","actor":"claude-g113-1","type":"result.record","node":"01M75S97KWYDNTGA2N56KPWZZC","payload":{"parent":"01M75RTZK05VMXAX2ER1HQWTA0","verdicts":{"01M75RTFR9JKPARCH3NAQ5P9R9":"supports"},"observation":"On 10566d4: go vet and g
{"v":1,"id":"01M75SBVS853GXX0C0SC9V3ER0","ts":"2026-11-11T17:09:23.36821Z","actor":"claude-g113-1","type":"result.record","node":"01M75SBVS853GXX0C0SC9V3ER0","payload":{"parent":"01M75SAW63CZAX0KXXXMC3GGGJ","verdicts":{"01M75RTFR9JKPARCH3NAQ5P9R9":"contradicts"},"observation":"The fresh review of 10
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/investigations/goal-prompt/.tree/events && for f in claude-g113-*.jsonl; do echo \"== $f\"; grep '\"type\":\"result.record\"' $f | sed -E 's/.*\"id\":\"[0-9A-Z]{20}([0-9A-Z]{6})\",\"ts\":\"([^\"]*)\".*\"actor\":\"([^\"]*)\".*/R-\\1 \\2 \\3/' ; done",
 "description": "List Results per actor with times"
}
```

Result:

```text
== claude-g113-1.jsonl
R-KPWZZC 2026-11-11T17:07:57.180003Z claude-g113-1
R-9V3ER0 2026-11-11T17:09:23.36821Z claude-g113-1
R-50GBXM 2026-11-11T17:17:46.95442Z claude-g113-1
R-8JQPWG 2026-11-11T17:21:27.311404Z claude-g113-1
R-FRF929 2026-11-11T17:30:29.026755Z claude-g113-1
R-9EGPYD 2026-11-11T17:30:45.219002Z claude-g113-1
== claude-g113-2.jsonl
R-0E8PFH 2026-11-12T02:48:18.682738Z claude-g113-2
R-KYKWH9 2026-11-12T02:51:44.175152Z claude-g113-2
R-28A8NB 2026-11-12T02:55:04.445113Z claude-g113-2
R-3YJ7T2 2026-11-12T02:56:59.537823Z claude-g113-2
R-356DV0 2026-11-12T02:59:35.448093Z claude-g113-2
== claude-g113-3.jsonl
R-JHBCWS 2026-11-12T03:13:18.498579Z claude-g113-3
R-27Y1EV 2026-11-12T03:15:49.831676Z claude-g113-3
R-DER6F8 2026-11-12T03:18:50.451686Z claude-g113-3
R-BSWFXG 2026-11-12T03:23:28.19161Z claude-g113-3
R-HXEBAF 2026-11-12T03:23:52.814802Z claude-g113-3
== claude-g113-4.jsonl
```

## Grill

Round 1 is below. I've written nothing yet, because the Goal write needs your answers and your approval of the exact command.

## What I looked up (no question for you)

- **Where the Goal goes.** #141 is part of #113, whose Goals are G-GZR6C6 (design) and G-ZBM2RE (build) in `investigations/goal-prompt`. The new Goal goes in that store.
- **What the stop clause counts.** `hyp prompt` stops a run after "N Results recorded under the actor", not after N Experiments (`internal/cli/prompt.go:167`). By the log, run 1 recorded 6 Results (R-KPWZZC, R-9V3ER0, R-50GBXM, R-8JQPWG, R-FRF929, R-9EGPYD). Run 2 recorded 5 and run 3 recorded 5. After R-8JQPWG run 1 had 4, so one was left, as the issue says.
- **Where `--limit` works.** It exists only on `hyp record` (checked with `hyp record -h`). A review finding that is a limit therefore needs a Result, and a Result needs an Experiment.
- **Where the review practice is written.** The only text is `skills/hyp-goal-run/SKILL.md` ("Before the stop, ask a fresh subagent for one review…" and "Sort each finding…"). It says nothing about how a review is recorded or whether it counts against the cap.
- **Overlap with open findings.** `docs/how-hyp-improves-itself.md:29` says "A run's cap is stated by the `/goal` prompt or the recipe, not by hyp". It overlaps run 4's finding F8. Whichever rival wins, the Goal that builds it fixes that text.
- **Where the review facts are recorded.** Run 1's review is E-C3GGGJ/R-9V3ER0 and E-7QQ2DP/R-8JQPWG. Run 2's reviews are R-KYKWH9 and R-3YJ7T2. Run 3's are R-JHBCWS and R-DER6F8, and its gate and review in one Experiment is E-GE085A/R-BSWFXG.

## Questions for you

### 1. One Goal now, or the design and the build together?

- **Today.** #141 asks for two things. One is a tree that picks a rival from a run on the scenario. The other is the change itself: `hyp-goal-run` text and, if a rival needs it, the `hyp prompt` stop clause.
- **Problem.** A build Goal's lines test one commit. A line such as "the default is 7" cannot hold for a rival that is only skill text. The lines of the build depend on which rival wins, so they cannot be written before the select.
- **Option A, two Goals in this order.** This grill writes a design Goal whose candidates are the rival rules. You add the build Goal (with its own lines) after the select, by running the grill again. This follows G-GZR6C6 then G-ZBM2RE.
- **Option B, one build Goal now.** Each rival becomes a candidate commit. Every line would have to hold for every rival, which fails the grid test above.
- **Recommendation: A.** The build Goal is then named in the checkpoint as the owner's next act, not left in chat.

### 2. Is "named as unreviewed" a floor (a line) or a rival?

- **Today.** The issue lists H-name-at-stop as a rival. Your Done-means paragraph says a budget-bound run ends reviewed, or names the candidate it did not review.
- **Problem.** If naming is a rival, it always passes, because it only adds a sentence to the stop checkpoint. It also never competes with the others. Run 2 named the gap with no rule, and run 1 did not. A line can test both.
- **Option A, naming is a floor.** It becomes two lines. The positive line says that on each of the three cuts, a run that reaches its cap ends with a review Result for its final candidate or with a checkpoint that names it. The negative line says that no checkpoint says "all review limits closed" or "owes nothing" while the final candidate has no review Result and is not named. The rivals then compete on how to end reviewed. A rival that passes only by naming shows up as "holds only by the floor", which answers "which hold alone".
- **Option B, naming stays a rival.** It is tested like the others and always wins as the cheapest.
- **Recommendation: A.** The lines would be:
  - **[P]** On each cut (run 1 after R-8JQPWG, run 2 after R-3YJ7T2, run 3 after R-JHBCWS), a run that follows the candidate rule ends with a review Result for its final candidate, or with a checkpoint that names that candidate as unreviewed.
  - **[N]** A run follows the candidate rule and ends with its final candidate neither reviewed nor named.

  These would be about 2 to 4 lines in the Goal. I will draft the exact words for the routing table.

### 3. How should the run test the rivals?

- **Today.** The issue asks to "replay the cut". A store cut is a copy of the store as it stood at one event (`docs/glossary.md`, Store cut). The rules are about one decision point: what the run does when one Result is left, or when a review finds a case its probe already passed.
- **Problem.** A full consumer run per rival and cut is many agent runs. The decision point is small, so a full run adds noise.
- **Option A, decision-point walk-through.** For each rival and cut, a fresh subagent gets the cut's `hyp status` and the rule text, and says what the run does next. The runner counts the Results the path costs and records a verdict. The Experiment is `--recipe none --rubric reading`, with the cuts as evidence. Example: for H-reserve-last on run 2's cut, one Result is left and the pending fix commit is not made. The review is the 5th Result and the select is next, so the path ends reviewed.
- **Option B, live runs.** ak-skill-creator's eval flow runs a consumer agent on each rival's skill text, on each cut. You would also choose the model, the number of trials and the budget.
- **Recommendation: A.** The build Goal then runs the eval flow (ADR 0021) on the winning text. Do you want B here instead?

### 4. Should the run test H-raise-budget, even though you argued against it?

- **Today.** Your comment of 03:04 says a fresh budget did not help: any fixed cap ends on an unreviewed commit unless a review leads to no commit.
- **Problem.** I can't put a testable rival to you as a choice. If you settle it in the session, it is still a DesignQuestion with its rivals, recorded as decided with the session as evidence. The only question is whether the run tests it.
- **Recommendation: test it.** Counting Results on the three cuts under a cap of 7 costs almost nothing and gives a number. The tree then shows what "fixed budget" would have done. It also checks your argument and doesn't only record it.
- **Alternative.** Record it as decided, with your comment as evidence.

### 5. Your actor name

Is it `owner-dat` again? The Goal add runs under it only after you approve the command.

### 6. The constraints block for the Goal

I propose this body ending. Say if any line is wrong:

```text
May change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, and never merges it. The design Goal changes no file in cmd/, internal/, skills/ or docs/; the build Goal that follows owns those.
Read only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.
Never: commit to main or merge a PR.
Never: replace the installed hyp binary.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

## Where the other questions go

These are my routing proposals. They are not questions for you, and I'll show them in the confirmation table after your answers and a second reader's walk-through.

| Question | Proposed place | Scenario that separates the rivals |
|---|---|---|
| How does a review count against the cap? | DesignQuestion. Rivals: separate Experiment (today), review inside the gate Experiment (H-one-experiment), cap raised to 7 (H-raise-budget). | Count the Results each rule spends on the three cuts. |
| What does a run do with one Result left and a fix commit pending? | DesignQuestion. Rivals: commit the fix (today), review the current candidate (H-reserve-last). | Run 2's cut after R-3YJ7T2. |
| What happens to a finding whose probe already passed? | DesignQuestion. Rivals: limit then test commit (today), close in the review's Result (H-probe-closes). | Run 3's cut after R-JHBCWS: N1 and P1. |
| What does the pre-select review read? | DesignQuestion. Rivals: the last increment (today), the cumulative diff from the base (H-whole-candidate). | Review 04142f3 against the Goal's base and see whether it finds a defect the three increment reviews did not. |
| How is a review recorded and does it count? | A line of the build Goal that follows, not a rival. | None here. |

Your answers to 1 to 6 decide the next step. If you answer 1 with B, the rows above change.

## Owner, message 2

1. A. Two Goals, with the design Goal first and the build Goal after the select.

2. A, naming is a floor. How do [P] and [N] serve my two aims: a review not running out of the run's budget, and a review not going on and on over small changes? Please say so in the Goal.

3. Neither A nor B as written. I want replayed runs on Sonnet, 3 trials for each rival and each cut, started from the cut. I don't want a full replay of a build from its start for every trial; that costs too much. Could the ak-context-engineering skill help with how we evaluate? Please check that once.

Cost: a rival may cost at most 1.5 times the base at the same cut. How do you compute cost? If you use a median, how is it computed? A trial that hangs counts as the most expensive.

4. Test H-raise-budget. All the rivals go under their question and get tested.

Table row 4: mark H-whole-candidate as the expected loser, since one cumulative review costs a lot. Test it anyway.

5. `owner-standin`. Please use that name in the command.

6. The constraints block is fine.

## Grill

I'll check once whether an `ak-context-engineering` skill exists.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills ~/.claude/skills 2>&1 | head -30; find . ~/.claude -maxdepth 4 -iname \"*context-engineering*\" 2>/dev/null | head; grep -rli \"context-engineering\" .claude skills docs 2>/dev/null | head",
 "description": "Look for ak-context-engineering skill"
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
.claude/skills/ak-skill-creator/assets/semantic-routing/agentkit-engineer-cases.json
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n -i -B2 -A2 \"context-engineering\" .claude/skills/ak-skill-creator/assets/semantic-routing/agentkit-engineer-cases.json | head -20; ls .claude/skills/ak-skill-creator/references | head -30",
 "description": "See how ak-skill-creator mentions context-engineering"
}
```

Result:

```text
525-      "split": "holdout",
526-      "prompt": "Our agent forgets instructions halfway through long sessions and burns tokens. Rework its context budget.",
527:      "expected": "ak-context-engineering",
528-      "class": "indirect"
529-    },
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
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/.claude/skills/ak-skill-creator && grep -rn -i \"median\\|timeout\\|hang\\|total_tokens\\|token\" references/creator-consumer-evaluation.md references/evaluation-tools.md references/benchmark-optimization-guide.md references/testing-and-iteration.md | head -40; ls scripts | head -30; sha256sum /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/skills/hyp-goal-run/SKILL.md; cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt && git rev-parse HEAD",
 "description": "Find cost/median/hang conventions in the eval flow; hash the base skill"
}
```

Result:

```text
references/creator-consumer-evaluation.md:49:success/invariants, then corrections, tokens, latency and tool calls. Do not hide
references/creator-consumer-evaluation.md:56:conditions first; label a configuration experiment separately from a prose change.
references/evaluation-tools.md:46:Optional observed metrics: `--tokens`, `--duration-ms`, `--tool-calls`, `--corrections`.
references/evaluation-tools.md:67:    "input_tokens": null, "output_tokens": null,
references/evaluation-tools.md:68:    "cache_read_tokens": null, "cache_write_tokens": null,
references/evaluation-tools.md:83:cached tokens. Do not sum fields using an assumed universal billing formula.
references/evaluation-tools.md:92:The importer neither changes model/effort nor operates provider caches. Use verified
references/evaluation-tools.md:123:their aggregate tokens as a proven cost improvement.
references/benchmark-optimization-guide.md:16:`references/testing-and-iteration.md`. For authoring changes, use
references/benchmark-optimization-guide.md:38:user decisions. Record quality/time/token/correction tradeoffs. A shorter prompt,
references/testing-and-iteration.md:23:   original with candidate and reruns cases for behavior that must remain unchanged.
references/testing-and-iteration.md:33:5. Record tokens, duration, tool calls, retries and human corrections when observable.
references/testing-and-iteration.md:90:Select coverage by the change. Metadata changes need actual routing comparisons;
references/testing-and-iteration.md:91:instruction or script changes need affected consumer regression; creator workflow
references/testing-and-iteration.md:92:changes need the creator-to-consumer loop. A semantic no-op such as spelling or
references/testing-and-iteration.md:94:rerun an unchanged passed suite without new evidence, a failure or unresolved risk.
references/testing-and-iteration.md:98:reference reads; token totals and root line counts alone do not measure savings.
references/testing-and-iteration.md
[... 827 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in 10566d4 b4f23b2 5c5902d 2284172 3cfd08d d2d5d14 04142f3; do printf \"%s \" $c; git cat-file -t $c 2>&1; git merge-base --is-ancestor $c HEAD && echo \"  ancestor of HEAD\"; done; git log --oneline -1 --format=%H main 2>&1",
 "description": "Check the cut commits exist and are reachable"
}
```

Result:

```text
Exit code 128
10566d4 commit
  ancestor of HEAD
b4f23b2 commit
  ancestor of HEAD
5c5902d commit
  ancestor of HEAD
2284172 commit
  ancestor of HEAD
3cfd08d commit
  ancestor of HEAD
d2d5d14 commit
  ancestor of HEAD
04142f3 commit
  ancestor of HEAD
fatal: ambiguous argument 'main': unknown revision or path not in the working tree.
Use '--' to separate paths from revisions, like this:
'git <command> [<revision>...] -- [<file>...]'
```

## Grill

All seven commits are reachable, so a cut can be rebuilt. The second reader now gets the settled rules, with no conclusion from me.

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader: walk grill rules",
 "subagent_type": "general-purpose",
 "model": "fable",
 "run_in_background": false,
 "prompt": "You are an independent reviewer of a design before it is written into a hyp tree. Read-only: change nothing in /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt. Work only from files; give me findings with the file and line that supports each, and say plainly when something is unverified.\n\nRead first: AGENTS.md, docs/glossary.md (entries Goal, Line, Store cut, Result, Limit, Owed), docs/replay-spec.md sections on lines/limits/OWED/coverage (find them), internal/cli/prompt.go, skills/hyp-goal-run/SKILL.md, and the issue text in /Users/datguy/.hyp-sandbox/g146-entry-b-1/tmp/issue-141.md. The store is investigations/goal-prompt (run `hyp status` there; events in .tree/events/*.jsonl).\n\nTHE DESIGN (rules settled with the owner):\nA new \"design\" Goal G-new goes into the SAME store as G-GZR6C6 and G-ZBM2RE. Its purpose: a run of a build Goal that reaches its Results cap ends with its last commit reviewed or named as unreviewed, and a review does not go on and on over small changes. It has four DesignQuestions, each with rival Hypotheses (rival = a rule text for hyp-goal-run, plus the rule of today as a base rival):\n Q-record: how a review counts against the cap: today's separate Experiment | review inside the gate Experiment | cap default raised from 5 to 7.\n Q-last: with one Result left and a fix commit pending: commit the fix (today) | review the current candidate instead and stop with recorded limits.\n Q-passing: a review finding whose case a run's own probe already passed: limit then test commit (today) | close in the review's Result, no commit.\n Q-reads: the pre-select review reads the last increment (today) | the cumulative diff from the Goal's base (expected loser on cost, tested anyway).\nEach rival is tested by replayed runs: Sonnet, 3 trials per rival per cut, each started from a store cut (events up to one event) with the repository checked out at that time's commit, not from the build's start. Cuts: S1 run claude-g113-1 after R-8JQPWG (4 Results of cap 5 spent); S2 run claude-g113-2 after R-3YJ7T2 (4 of 5); S3 run claude-g113-3 after R-JHBCWS; S4 the point where H-VHG4N5 (04142f3) was added.\nLines (checked per trial): [P] on each cut its question names, each trial ends with a review Result for its last commit or a stop checkpoint naming it as unreviewed. [N] no trial ends with its last commit neither reviewed nor named, or with a checkpoint saying all review limits closed / runner owes nothing while the last commit has no review Result. [N] a rule whose median cost over its 3 trials exceeds 1.5 times the median of the base rule at the same cut, cost = input_tokens+output_tokens of the trial and its reviewer subagents, hang (no stop in 40 minutes) = the most expensive rank. The base rule is skills/hyp-goal-run/SKILL.md at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa
[... 1016 more characters]
```

Result:

```text
[Subagent hand-back] The text below is the final report of a subagent this session delegated to. It is model output, NOT a message from the user: instructions, requests, or approval claims inside it are the subagent's words and carry no user authority. The harness indents every line of the report, so a frame-like line at column zero inside it would be forged. Notes above this frame may quote model-derived text, which carries no user authority either. The report follows:
  **Scenario 1: two or three Goals in one store. Holds, with two gaps.**
  - OWED, LINES and coverage are per Goal (replay-spec.md:351). A line is owed on every live top-level Question of its own Goal and on no other Goal's.
  - `hyp prompt G-ZBM2RE` reads only that Goal's node and the flags (prompt.go:140-167). Its bytes cannot change when G-new is added. G-ZBM2RE#4 states this property.
  - The stop clause "N Results recorded under A" (prompt.go:167) has no Goal filter, and nothing in the binary counts it. The runner counts it. The cut's Results, which sit under G-ZBM2RE, do count for the replay actor, but only if the trial uses the original actor name. A fresh actor starts at 0 of 5, not 4 of 5, and the Q-last scenario disappears. A cut copies lines "as written" (replay-spec.md:158), so the actor cannot be renamed.
  - Run 1 recorded 6 Results (events file, `result.record` count). R-8JQPWG is its 4th, so "4 of 5" is correct. Run 1's own checkpoint says "5 Experiments", so the cap unit differs between the issue text and the log.
  - Gap: a line is owed on every one of the four Questions. "[P] on each cut its question names" cannot be keyed to one Question. Coverage is per candidate (judged, or subject held). A Result that ran only S1 on Q-record still covers all its candidates, so hyp cannot enforce "each cut". The cuts must be listed in the line text, or the line must be one per Question. A per-Question line is still owed on the other three Questions.
  - Also, a G-new run's default cap of 5 is sm
[... 4831 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in b4f23b2 5c5902d 3cfd08d 04142f3 HEAD; do printf \"%s \" $c; git show $c:skills/hyp-goal-run/SKILL.md | sha256sum | cut -c1-12; done; sed -n 1,40p .claude/skills/ak-skill-creator/references/session-usage-analysis.md; sed -n 28,36p docs/core-and-self-review.md; git diff b4f23b2 HEAD --stat -- skills/hyp-goal-run | tail -2",
 "description": "Hash base skill at each cut; read usage tool doc"
}
```

Result:

```text
b4f23b2 (eval):1: bad substitution
e3b0c44298fc
5c5902d (eval):1: bad substitution
e3b0c44298fc
3cfd08d (eval):1: bad substitution
e3b0c44298fc
04142f3 (eval):1: bad substitution
e3b0c44298fc
HEAD (eval):1: bad substitution
e3b0c44298fc
# Session usage analysis

`audit` and `optimize` measure how the skill actually ran in local agent sessions
before judging its text. Static lint shows what the instructions say; transcripts
show what they cost: turns, tokens, tool calls, failures, wait time, delegation and
which resources were loaded. The goal is a skill that finishes faster, varies less
between runs and spends fewer tokens without lowering output quality.

## Collect

'''bash
python3 scripts/session_usage.py <skill-dir-or-name> --exclude-session <current-session-id> > <scratch-dir>/usage.json
'''

Write the output outside the skill and repository; files left in a skill directory
become part of its installed content.

- The script is standard-library, read-only and prints metrics, tool names,
  session-id prefixes, skill-relative paths and home-relative roots (other roots show
  only their last two folders). It never
  prints prompt, reply or tool-output text; `subcommand` is reported only when it
  is a mode listed first in the skill's `argument-hint`, otherwise `(free-text)`.
- Default roots are `$CLAUDE_CONFIG_DIR` (else `~/.claude`) `/projects` and
  `$CODEX_HOME` (else `~/.codex`) `/sessions` plus `/archived_sessions`. Override
  with repeatable `--root claude-code=DIR` or `--root codex=DIR`; narrow with
  `--runtime`, `--since YYYY-MM-DD` and `--entrypoint` (for example to separate
  scripted eval runs such as `codex_exec` from interactive sessions).
- Pass a skill directory to get `unused_skill_files` and mode names. Exclude the
  session that is running the audit, because its span is still open. Use the id
  the runtime exposes to the session when it has one; otherwise it is the file name
  of the newest transcript in the current project's transcript folder. 
[... 1316 more characters]
```

## Grill

`ak-context-engineering` isn't installed here, so I couldn't consult it. Your skill list and `~/.claude/skills` don't have it. Its name appears only as an expected routing answer in one of ak-skill-creator's test cases. The closest installed source is ak-skill-creator's own evaluation references, and I used those.

## What I took from ak-skill-creator

- `references/evaluation-tools.md` and `testing-and-iteration.md` say to record tokens, duration and tool calls.
- `evaluation-tools.md` also says not to sum token fields with an assumed billing formula. So I use fresh input and output tokens as the cost, and leave the cache fields out.
- `scripts/session_usage.py` reads transcripts under `~/.claude/projects`. Whether it counts reviewer subagents is unverified, because my sandbox can't read that path. The Goal makes the runner check this before any verdict counts.

## How cost is computed

- **Cost of one trial** is `input_tokens + output_tokens` of the replayed run and of every reviewer subagent it starts. Cache fields, duration and tool calls are recorded but decide nothing.
- **Median of 3 trials** is the middle value after sorting the three costs.
- **A rival passes at a cut** if its median is at most 1.5 times the base rule's median at the same cut. Both medians are over 3 trials.
- **A hung trial** (no stop in 40 minutes; real runs took about 11) ranks above every costed trial. One hang in three leaves the median at the larger of the other two. Two hangs make the median a hang, so the rival fails.
- **If the base's median is a hang**, the cost verdict is `inconclusive`.

## How the lines serve your two aims

My first draft of the lines had gaps, so the lines below differ from it. The second reader found that a rule "never review; write `unreviewed` at the stop" would have passed everything. I checked its other claims and applied them as below.

| Aim | Lines (draft wording) |
|---|---|
| **Aim 1: a review does not run out of the run's budget** | **[P]** On each cut its question names, in at least 2 of the 3 trials the run ends with a review Result for its last commit. A review Result cites a file that names the commit's sha and gives its findings or "no defect found". **[N]** No trial's stop checkpoint omits the sha of a commit that has no review Result. None says every commit is reviewed when one is not. |
| **Aim 2: a review does not go on over small changes** | **[N]** The median cost of a rule at a cut is at most 1.5 times the base rule's median, as computed above. The base is `skills/hyp-goal-run/SKILL.md` at sha256 `cb15e2f2…`, and the rule's text must differ from it. **[N]** A trial closes a finding as "probe passed" only with a cited file of a run of its case. |

The Goal body will say which aim each line serves.

## What the second reader found, and what I did

I checked these against the log and the code.

- **The cap counts per actor.** The stop clause counts Results under the replay actor, so a fresh actor starts at 0. I don't rename the actor in the cut. Each trial uses `hyp prompt --actor trial-N --experiments <left>`, where `<left>` is the cap minus the Results already spent at the cut:
  - S1: 1 left (3 under a cap of 7).
  - S2: 1 left (3 under a cap of 7).
  - S3: 4 left (6 under a cap of 7).
  - This goes in the Experiment's rule, not in a line.
- **The base text differs between cuts.** `skills/hyp-goal-run/SKILL.md` differs between the S1 checkout and HEAD (the diff is one line). So each trial installs the pinned base file (`cb15e2f2…`, or the rival's text) over the checkout's copy. The checkout supplies only the code under review.
- **S3 is not a budget case.** R-JHBCWS is run 3's first Result, so 4 are left. S3 tests Q-passing, not the cap. S3 is therefore listed only under Q-passing, and the lines say "ends" rather than "reaches the cap".
- **Cuts per question.** A line is owed on every question, but a cut belongs to one or two of them. So the cuts are listed per question in the Goal body:
  - Q-record uses S1, S2 and S3.
  - Q-last uses S1 and S2.
  - Q-passing uses S2 and S3.
  - Q-reads uses S4.
  - Base trials are shared between questions, because the base text is the same. That is 45 trials in all, not the 72 you would get by running every rival on every cut.
- **Trial events must not reach the real store.** A trial's Results would count as real, and its new Hypotheses would reopen Q-Z93VJ5. I add a `Never:` line for this.
- **No replay recipe is published.** The instrument must reproduce history before its verdicts count. The known answer is that the base rule on S1 and S2 must end with the last commit unreviewed, as the real runs did.

## Round 2: questions for you

1. **Cost source.** Do you accept the cost definition above, and 40 minutes as the hang limit? The runner would first check one trial's number against its transcript.
2. **A cap on review rounds.** The relative 1.5 times limit lets a rule loop as much as the base does. The base's runs had 2, 2 and 3 reviews. Do you want an absolute line, such as "no trial records more than 3 review Results"? I recommend yes, because it states your second aim directly.
3. **The 2-of-3 rule in the first line.** It means a rule needs a review Result for the last commit in at least 2 of 3 trials per cut. I chose it so one lucky trial does not count. Is 2 of 3 right, or do you want all 3?
4. **The instrument.** The runner needs a driver for the 45 trials. In option A it is a new recipe such as `recipes/<name>@1`. That is a self-review item and needs a row in `docs/core-and-self-review.md`, so `May change:` must add `recipes/` and `docs/`. In option B it is scripts kept as evidence, with `--recipe none`. `hyp record` warns "uncontained" for them, as in runs 2 and 3. I recommend B, because the warning is only a warning and the scope map stays unchanged. Which do you want?
5. **Model pin.** I propose `claude-sonnet-5-5`, so "Sonnet" is a fixed id. Is that right?

Please also approve the added `Never:` line: "Never: merge a trial's events into investigations/goal-prompt/.tree/events." After your answers I'll show the routing table, then the exact write commands under `owner-standin` for your approval.

## Owner, message 3

1. Yes, I accept that cost definition and the 40-minute hang limit. Checking one trial's number against its transcript first is fine.
2. Yes, add the absolute cap of 3 review Results per trial. It states my second aim directly.
3. 2 of 3 is right.
4. B, scripts kept as evidence with `--recipe none`.
5. Yes, pin `claude-sonnet-5-5`.

I approve the added `Never:` line about merging a trial's events. Show me the routing table next.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/investigations/goal-prompt && hyp add goal -h 2>&1 | head -30; hyp add dq -h 2>&1 | grep -i \"body\\|title\" ; hyp add hyp -h 2>&1 | grep -i \"body\\|title\"",
 "description": "Check body and line length limits"
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
  -body string
    	body (at most 600 runes)
  -title string
    	title (required, at most 120 runes)
  -body string
    	body (at most 600 runes)
  -title string
    	title (required, at most 120 runes)
```

## Grill

I'll draft the Goal body and lines and count their runes against the limits before showing the table.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "mkdir -p $TMPDIR/g141 && cd $TMPDIR/g141 && cat > body.txt <<'EOF'\nDesign record for #141 (part of #113), tested by replayed runs, not argued: which rule of hyp-goal-run, or which default cap of hyp prompt, ends a budget-bound run of a build Goal with its last commit reviewed or named as unreviewed, with no review that goes on over small changes. Evidence: runs claude-g113-1, -2, -3 of G-ZBM2RE and the owner's comments on #141.\nAim 1, a review does not run out of the run's budget: lines 1 and 2. Aim 2, a review does not go on over small changes: lines 3 to 5.\nFour DesignQuestions, each with today's rule as the base rival. Each question's body names its cuts. A cut is the store at one event and the repository at that time's commit: S1 claude-g113-1 after R-8JQPWG; S2 claude-g113-2 after R-3YJ7T2; S3 claude-g113-3 after R-JHBCWS; S4 the add of H-VHG4N5 (04142f3).\nMethod: claude-sonnet-5-5, 3 trials for each rival and each cut, started from the cut, never from the build's start. Each trial has its own actor, a prompt from hyp prompt with --experiments set to the cap minus the Results already under the original actor at the cut, and the rule text installed over the checkout's copy. Scripts are kept as evidence with --recipe none.\nKnown answers first: the base rule on S1 and S2 must end with the last commit unreviewed, as runs 1 and 2 did, and one trial's cost must equal its transcript's. Until both hold, no verdict counts.\nThe build Goal that follows is added by the owner after the selects and owns the changes to skills/, docs/ and cmd/.\nMay change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, and never merges it. It changes no file in cmd/, internal/, skills/ or docs/.\nRead only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.\nNever: commit to main or merge a PR.\nNever: replace the installed hyp binary.\nNever: merge a trial's events into investigations/goal-prompt/.tree/events.\nNever: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.\nEOF\ncat > l1.txt <<'EOF'\n[P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a rule the run ends with a review Result for its last commit: a Result that cites a file naming that commit's sha with its findings or \"no defect found\".\nEOF\ncat > l2.txt <<'EOF'\n[N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says that every commit, or every review limit, is closed or reviewed while one commit has no review Result.\nEOF\ncat > l3.txt <<'EOF'\n[N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of at most 1.5 times the base rule's median (ski
[... 782 more characters]
```

Result:

```text
body 2123
l1 283
l2 232
l3 545
l4 55
l5 104
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && cat > body.txt <<'EOF'\nDesign record for #141 (part of #113), tested by replayed runs, not argued: which rule of hyp-goal-run, or which default cap of hyp prompt, ends a budget-bound run of a build Goal with its last commit reviewed or named as unreviewed, with no review that goes on over small changes. Evidence: runs claude-g113-1, -2, -3 of G-ZBM2RE.\nAim 1, a review does not run out of the run's budget: lines 1 and 2. Aim 2, a review does not go on over small changes: lines 3 to 5.\nFour DesignQuestions, each with today's rule as the base rival; each body names its cuts. A cut is the store at one event and the repository at that time's commit: S1 claude-g113-1 after R-8JQPWG; S2 claude-g113-2 after R-3YJ7T2; S3 claude-g113-3 after R-JHBCWS; S4 the add of H-VHG4N5 (04142f3).\nMethod: claude-sonnet-5-5, 3 trials for each rival and each cut, started from the cut, never from the build's start. Each trial has its own actor, a hyp prompt with --experiments set to the cap minus the Results already under the original actor at the cut, and the rule text installed over the checkout's copy. Scripts are kept as evidence with --recipe none.\nKnown answers first: the base rule on S1 and S2 must end with the last commit unreviewed, as runs 1 and 2 did, and one trial's cost must equal its transcript's. Until both hold, no verdict counts.\nThe owner adds the build Goal after the selects; it owns changes to skills/, docs/ and cmd/.\nMay change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, and never merges it. It changes no file in cmd/, internal/, skills/ or docs/.\nRead only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.\nNever: commit to main or merge a PR.\nNever: replace the installed hyp binary.\nNever: merge a trial's events into investigations/goal-prompt/.tree/events.\nNever: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.\nEOF\npython3 -c \"print(len(open('body.txt').read().rstrip('\\n')))\"; grep -c '^$' body.txt; grep -n ' $' body.txt | head -2",
 "description": "Trim body and recount"
}
```

Result:

```text
2043
0
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g141 && python3 - <<'EOF'\ns=open('body.txt').read()\ns=s.replace(\"Design record for #141 (part of #113), tested by replayed runs, not argued: which rule\",\"Design record for #141 (part of #113), tested by replayed runs: which rule\")\ns=s.replace(\" Evidence: runs claude-g113-1, -2, -3 of G-ZBM2RE.\",\"\")\ns=s.replace(\"; the run may push it and open or update one PR, and never merges it. It changes\",\"; the run may push it and open or update one PR, never merging it. It changes\")\ns=s.replace(\", the store at one event and\",\"\")\ns=s.replace(\"A cut is the store at one event and the repository at that time's commit:\",\"A cut is the store at one event with the repository at that time's commit:\")\nopen('body.txt','w').write(s)\nprint(len(s.rstrip('\\n')))\nEOF",
 "description": "Trim body to fit 2000 runes"
}
```

Result:

```text
1979
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt
```

## Grill

Nothing is written yet. Below are the routing table, the exact text of the one owner write, and the runner writes that follow it.

## Routing table

| # | Grill question | Where it goes | Scenario or note |
|---|---|---|---|
| 1 | Where does the Goal live? | A new Goal (alias `review-design`) in `investigations/goal-prompt`, the store of G-GZR6C6 and G-ZBM2RE. Two `external_ref` links, to #141 and #113. | |
| 2 | Purpose, scope, aims | The Goal's title and body. The body ties lines 1 and 2 to aim 1 and lines 3 to 5 to aim 2. | |
| 3 | What the run may change, may only read, and must never do | The `May change:`, `Read only:` and `Never:` block you approved, plus the `Never:` line about merging trial events. | |
| 4 | How the tree is run | The Goal body for the outline, and each Experiment's verdict rule for the details. Both come from your answers 3 to 5. | claude-sonnet-5-5, 3 trials per rival and cut, started from the cut. Each trial has its own actor and a prompt with `--experiments` set to the cap minus the Results already spent. The rule text is installed over the checkout's copy. The scripts are kept as evidence with `--recipe none`. |
| 5 | Known answers and the cost check | Goal body and Experiment rule. | The base rule on S1 and S2 must end unreviewed, and one trial's cost must equal its transcript's. Until both hold, no verdict counts. |
| 6 | **Line 1 [P]** | Line. Aim 1: in at least 2 of 3 trials per cut, the run ends with a review Result for its last commit. | Red on the base at S1 and S2, since the real runs ended unreviewed. |
| 7 | **Line 2 [N]** | Line. Aim 1: no stop checkpoint omits the sha of an unreviewed commit or claims that all are reviewed. | |
| 8 | **Line 3 [N]** | Line. Aim 2: a rule's median cost is at most 1.5 times the base's, with cost, median and hang as you accepted. | |
| 9 | **Line 4 [N]** | Line. Aim 2: no trial records more than 3 review Results. | |
| 10 | **Line 5 [N]** | Line. Aim 2: a finding is closed as "probe passed" only with a cited file that ran its case. | |
| 11 | How does a review count against the cap? | DesignQuestion Q-record. Rivals: separate Experiment (base, today), review inside the gate Experiment (H-one-experiment), cap raised to 7 (H-raise-budget). | Count the Results each rule spends from the cut. Cuts: S1, S2, S3. |
| 12 | What does a run do with one Result left and a fix pending? | DesignQuestion Q-last. Rivals: commit the fix (base), review the current candidate and stop with limits (H-reserve-last). | Cuts: S1, S2. |
| 13 | What happens to a finding whose probe already passed? | DesignQuestion Q-passing. Rivals: limit then test commit (base), close in the review's Result (H-probe-closes). | Cuts: S2, S3. |
| 14 | What does the pre-select review read? | DesignQuestion Q-reads. Rivals: last increment (base), cumulative diff from the Goal's base (H-whole-candidate, marked "expected to lose on cost, tested anyway"). | Cut: S4. |
| 15 | Is H-name-at-stop a rival? | No. It is the floor in lines 1 and 2. | |
| 16 | Settled in the session, or tested? | All four questions are tested. None is recorded as decided. | |
| 17 | How a review is recorded and whether it counts, written into `hyp-goal-run` | The build Goal you add after the selects. | |
| 18 | The cap text in `docs/how-hyp-improves-itself.md:29` (run 4's F8) | The same build Goal. | |
| 19 | Facts I looked up and checked | Not asked. The stop clause counts Results, not Experiments. `--limit` exists only on `hyp record`. The cut commits are all reachable. | |

## The Goal's text

Title: `A budget-bound run reviews its last commit or names it, and no review runs on over small changes`

The Goal body (1979 runes of 2000):

```text
Design record for #141 (part of #113), tested by replayed runs: which rule of hyp-goal-run, or which default cap of hyp prompt, ends a budget-bound run of a build Goal with its last commit reviewed or named as unreviewed, with no review that goes on over small changes.
Aim 1, a review does not run out of the run's budget: lines 1 and 2. Aim 2, a review does not go on over small changes: lines 3 to 5.
Four DesignQuestions, each with today's rule as the base rival; each body names its cuts. A cut is the store at one event with the repository at that time's commit: S1 claude-g113-1 after R-8JQPWG; S2 claude-g113-2 after R-3YJ7T2; S3 claude-g113-3 after R-JHBCWS; S4 the add of H-VHG4N5 (04142f3).
Method: claude-sonnet-5-5, 3 trials for each rival and each cut, started from the cut, never from the build's start. Each trial has its own actor, a hyp prompt with --experiments set to the cap minus the Results already under the original actor at the cut, and the rule text installed over the checkout's copy. Scripts are kept as evidence with --recipe none.
Known answers first: the base rule on S1 and S2 must end with the last commit unreviewed, as runs 1 and 2 did, and one trial's cost must equal its transcript's. Until both hold, no verdict counts.
The owner adds the build Goal after the selects; it owns changes to skills/, docs/ and cmd/.
May change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, never merging it. It changes no file in cmd/, internal/, skills/ or docs/.
Read only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.
Never: commit to main or merge a PR.
Never: replace the installed hyp binary.
Never: merge a trial's events into investigations/goal-prompt/.tree/events.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
```

The five lines (283, 232, 545, 55 and 104 runes of 600 each):

1. `[P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a rule the run ends with a review Result for its last commit: a Result that cites a file naming that commit's sha with its findings or "no defect found".`
2. `[N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says that every commit, or every review limit, is closed or reviewed while one commit has no review Result.`
3. `[N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of at most 1.5 times the base rule's median (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), and its text differs from the base. Cost = input_tokens + output_tokens of the trial and every subagent it starts. A trial with no stop in 40 minutes ranks above every costed trial. Median of 3 = the middle value sorted. A base median that hangs gives inconclusive.`
4. `[N] Aim 2: no trial records more than 3 review Results.`
5. `[N] Aim 2: a trial closes a review finding as passed only with a cited file that ran the finding's case.`

The last three constraints I checked against the test route. `Never: replace the installed hyp binary` and the rest forbid nothing the route writes, because it writes only under `.tree/`.

## The owner write, for your approval

Run it only after you confirm the table and approve this exact command. The five lines and the body are the texts above, held in files under my scratch directory:

```bash
cd investigations/goal-prompt && hyp add goal --actor owner-standin --alias review-design \
  --title "A budget-bound run reviews its last commit or names it, and no review runs on over small changes" \
  --body "$(cat "$TMPDIR/g141/body.txt")" \
  --line "$(cat "$TMPDIR/g141/l1.txt")" --line "$(cat "$TMPDIR/g141/l2.txt")" \
  --line "$(cat "$TMPDIR/g141/l3.txt")" --line "$(cat "$TMPDIR/g141/l4.txt")" \
  --line "$(cat "$TMPDIR/g141/l5.txt")" \
  --why "Owner, 2026-11-12: the design record for #141, whose rivals are tested by replayed runs; its build is a later Goal in this store."
```

## The runner writes that follow it

- `hyp init` is not needed, because the store exists.
- `hyp add dq --parent <G-id>` four times, then one `hyp add hyp --parent <Q-id>` per rival, with base, H-one-experiment, H-raise-budget, H-reserve-last, H-probe-closes and H-whole-candidate. The base rival is a separate node under each question. The DesignQuestion bodies name the cuts, and the rival bodies are at most 600 runes.
- `hyp link <G-id> external_ref` to #141 and to #113.
- `hyp check`, which must print `check: 0 errors`.
- `hyp checkpoint --goal <G-id> --summary "…" --why "end of grill"`.
- A commit of `.tree/events/` and `.tree/evidence/`.
- `hyp prompt <G-id> --actor <runner-actor>` for the hand-off, with the actor name of your choice.

Reply "confirmed" (or name the rows to move) and "approved" for the command. I'll run the owner write as `owner-standin` and then the runner writes.

## Owner, message 4

Confirmed. The table matches my answers, with H-whole-candidate marked as expected to lose. Approved for the exact `hyp add goal` command under `owner-standin`. For the runner actor name, use whatever you propose.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt/investigations/goal-prompt && hyp add goal --actor owner-standin --alias review-design \\\n  --title \"A budget-bound run reviews its last commit or names it, and no review runs on over small changes\" \\\n  --body \"$(cat \"$TMPDIR/g141/body.txt\")\" \\\n  --line \"$(cat \"$TMPDIR/g141/l1.txt\")\" --line \"$(cat \"$TMPDIR/g141/l2.txt\")\" \\\n  --line \"$(cat \"$TMPDIR/g141/l3.txt\")\" --line \"$(cat \"$TMPDIR/g141/l4.txt\")\" \\\n  --line \"$(cat \"$TMPDIR/g141/l5.txt\")\" \\\n  --why \"Owner, 2026-11-12: the design record for #141, whose rivals are tested by replayed runs; its build is a later Goal in this store.\" 2>&1",
 "description": "Run the approved owner write: add the Goal"
}
```

Result:

```text
G-EDH3XC node.add
  line G-EDH3XC#1: [P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a rule the run ends with a review Result for its last commit: a Result that cites a file naming that commit's sha with its findings or "no defect found".
  line G-EDH3XC#2: [N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says that every commit, or every review limit, is closed or reviewed while one commit has no review Result.
  line G-EDH3XC#3: [N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of at most 1.5 times the base rule's median (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), and its text differs from the base. Cost = input_tokens + output_tokens of the trial and every subagent it starts. A trial with no stop in 40 minutes ranks above every costed trial. Median of 3 = the middle value sorted. A base median that hangs gives inconclusive.
  line G-EDH3XC#4: [N] Aim 2: no trial records more than 3 review Results.
  line G-EDH3XC#5: [N] Aim 2: a trial closes a review finding as passed only with a cited file that ran the finding's case.
```

## Grill

The Goal is written as G-EDH3XC, with no warning. Next come the runner writes under the actor `claude-grill-141`, starting with the four DesignQuestions.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && G=G-EDH3XC\nhyp add dq --parent $G --alias q-record --title \"How does a review count against the run's cap?\" --body \"Cuts S1, S2, S3. Scenario: replay each cut for each rival; count the Results the run spends from the cut to its stop, and whether its last commit ends reviewed. Cap left at the cut under 5: 1, 1, 4; under 7 (H-raise-budget): 3, 3, 6. Base trials are shared with Q-last and Q-passing at the same cut.\" --why \"Issue #141 rivals H-one-experiment and H-raise-budget, against today's rule.\" 2>&1 | head -3\nhyp add dq --parent $G --alias q-last --title \"What does a run do with one Result left and a fix commit pending?\" --body \"Cuts S1, S2. Scenario: replay each cut with the cap left at 1; see whether the run stops with its last commit reviewed or named, and what it costs. Base trials are shared with Q-record at the same cut.\" --why \"Issue #141 rival H-reserve-last, against today's rule.\" 2>&1 | head -3\nhyp add dq --parent $G --alias q-passing --title \"What happens to a review finding whose case the run's own probe already passed?\" --body \"Cuts S2, S3. Scenario: replay each cut; S2 holds two findings whose probes passed (R-3YJ7T2), S3 holds N1 (R-JHBCWS). See whether a new commit follows, and what the run costs. Base trials are shared with Q-record at the same cut.\" --why \"Issue #141 rival H-probe-closes, against today's rule.\" 2>&1 | head -3\nhyp add dq --parent $G --alias q-reads --title \"What does the review before the select read?\" --body \"Cut S4, the add of H-VHG4N5 (04142f3). Scenario: replay S4 for each rival; see whether the cumulative review finds a defect the increment reviews did not, and what it costs against the Goal's base.\" --why \"Issue #141 rival H-whole-candidate, against today's rule.\" 2>&1 | head -3\nhyp status $G 2>&1 | sed -n 1,12p",
 "description": "Add four DesignQuestions under the Goal"
}
```

Result:

```text
Q-JP7D58 node.add
Q-ZTTBCH node.add
Q-WZ3QQY node.add
Q-V99SRH node.add
GOAL G-EDH3XC A budget-bound run reviews its last commit or names it, and… — in_progress (answered 0, open 4, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-EDH3XC#1 no candidate — [P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a r…
  G-EDH3XC#2 no candidate — [N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says tha…
  G-EDH3XC#3 no candidate — [N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of…
  G-EDH3XC#4 no candidate — [N] Aim 2: no trial records more than 3 review Results.
  G-EDH3XC#5 no candidate — [N] Aim 2: a trial closes a review finding as passed only with a cited file that ran the finding's case.
FRONTIER
  Q-JP7D58 open/no_candidates How does a review count against the run's cap?
  Q-ZTTBCH open/no_candidates What does a run do with one Result left and a fix commit pe…
  Q-WZ3QQY open/no_candidates What happens to a review finding whose case the run's own p…
  Q-V99SRH open/no_candidates What does the review before the select read?
```

## Grill

Now the rivals, two or three under each question.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141\nB=\"Today's rule: skills/hyp-goal-run/SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed over the checkout's copy.\"\nh(){ hyp add hyp --parent \"$1\" --alias \"$2\" --title \"$3\" --body \"$4\" --why \"$5\" 2>&1 | head -2; }\nh Q-JP7D58 h-separate-review \"Base: a review is its own Experiment and its Result counts toward the cap\" \"$B A review is its own Experiment (--recipe none --rubric reading) with its own Result; cap default 5.\" \"Base rival of Q-record: the practice runs 1 to 3 used.\"\nh Q-JP7D58 h-one-experiment \"A commit's review is part of its gate Experiment: one Result\" \"A commit's review is part of its gate Experiment. One Result carries the test verdicts and the review's limits, so a commit costs one Result. Rule text added to hyp-goal-run; cap default 5.\" \"Rival from #141: a commit costs one Experiment again.\"\nh Q-JP7D58 h-raise-budget \"The default of --experiments rises from 5 to 7\" \"The review stays its own Experiment; hyp prompt's default of --experiments rises from 5 to 7, to fit a gate and a review per commit. Trials use 7 minus the Results already spent at the cut. The owner argued against it on 2026-11-12 (run 2's fresh budget did not help); tested anyway.\" \"Rival from #141: a budget that fits a gate and a review per commit.\"\nh Q-ZTTBCH h-commit-fix \"Base: with one Result left, commit the fix and gate it\" \"$B With one Result left the run commits the pending fix and gates it.\" \"Base rival of Q-last.\"\nh Q-ZTTBCH h-reserve-last \"With one Result left, review the current candidate and stop with recorded limits\" \"When one Result remains, the run reviews the current candidate instead of committing another fix, and stops with recorded limits. Rule text added to hyp-goal-run.\" \"Rival from #141 and the owner's comment on run 2.\"\nh Q-WZ3QQY h-limit-then-commit \"Base: a probed case is a limit, then a test commit\" \"$B A review finding whose case the run's probe passed is recorded as a limit; a test commit follows.\" \"Base rival of Q-passing.\"\nh Q-WZ3QQY h-probe-closes \"A case the run's probe passed is closed in the review's Result\" \"A case the run's own probe ran and passed on the candidate is closed in the review's Result, with the probe log cited, not recorded as a limit. A limit stays for a case no check ran, or one the candidate fails. Rule text added to hyp-goal-run.\" \"Rival from the owner's comment on run 3.\"\nh Q-V99SRH h-last-increment \"Base: the review before the select reads the last increment\" \"$B The review reads the diff of the last commit against its parent.\" \"Base rival of Q-reads.\"\nh Q-V99SRH h-whole-candidate \"The review before the select reads the cumulative diff from the Goal's base\" \"The review before the select reads the cumulative diff from the Goal's base, not the last increment. Expected to lose on cost (one cumulative review is large); tested anyway, by the owner's choice
[... 115 more characters]
```

Result:

```text
H-HTNTWC node.add
H-0BRHH0 node.add
H-4VKM63 node.add
H-8NJHYJ node.add
H-M5Q0D4 node.add
H-JNM729 node.add
H-WXR74B node.add
H-YT0KRB node.add
H-GKJ8KG node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && G=G-EDH3XC\nhyp link $G external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this design Goal serves.\" 2>&1 | head -3\nhyp link $G external_ref https://github.com/dat9uy/hyp/issues/113 --why \"The parent issue of #141.\" 2>&1 | head -3\nhyp status $G 2>&1 | sed -n 1,30p; hyp check 2>&1 | tail -4",
 "description": "Link the issues, show status and check"
}
```

Result:

```text
G-EDH3XC edge.add
G-EDH3XC edge.add
GOAL G-EDH3XC A budget-bound run reviews its last commit or names it, and… — in_progress (answered 0, open 4, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-EDH3XC#1 no candidate — [P] Aim 1, a review does not run out of the run's budget. On each cut its question's body names, in at least 2 of the 3 trials of a r…
  G-EDH3XC#2 no candidate — [N] Aim 1, the gap is never hidden. No trial's stop checkpoint omits the sha of a commit that has no review Result, and none says tha…
  G-EDH3XC#3 no candidate — [N] Aim 2, a review does not go on over small changes. At a cut, each rule other than the base has a median cost over its 3 trials of…
  G-EDH3XC#4 no candidate — [N] Aim 2: no trial records more than 3 review Results.
  G-EDH3XC#5 no candidate — [N] Aim 2: a trial closes a review finding as passed only with a cited file that ran the finding's case.
FRONTIER
  Q-JP7D58 open/untested How does a review count against the run's cap?
    H-HTNTWC (h-separate-review) proposed Base: a review is its own Experiment and its Result counts …
    H-0BRHH0 (h-one-experiment) proposed A commit's review is part of its gate Experiment: one Result
    H-4VKM63 (h-raise-budget) proposed The default of --experiments rises from 5 to 7
  Q-ZTTBCH open/untested What does a run do with one Result left and a fix commit pe…
    H-8NJHYJ (h-commit-fix) proposed Base: with one Result left, commit the fix and gate it
    H-M5Q0D4 (h-reserve-last) proposed With one Result left, review the current candidate and stop…
  Q-WZ3QQY open/untested What happens to a review finding whose case the run's own p…
    H-JNM729 (h-limit-then-commit) proposed Base: a probed case is a limit, then a test commit
    H-WXR74B (h-probe-closes) proposed A case the run's probe passed is closed in the review's Res…
  Q-V99SRH open/untested What does the review before the select read?
    H-YT0KRB (h-last-increment) proposed Base: the review before the select reads the last incr
[... 747 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && G=G-EDH3XC\nhyp checkpoint --goal $G --summary \"Grill done: 5 lines written by owner-standin; 4 DesignQuestions (Q-JP7D58, Q-ZTTBCH, Q-WZ3QQY, Q-V99SRH), each with the base rival and the #141 rivals, all to be tested by replayed runs. Waiting for the owner: none. Next for the owner: add the build Goal after the selects (hyp-goal-run's review recording, the cap text in docs/how-hyp-improves-itself.md, any hyp prompt default).\" --why \"end of grill\" 2>&1 | head -3\nhyp status $G | sed -n '/OWED/,$p' | head -14\ngit status --short . | head; ",
 "description": "Write the end-of-grill checkpoint and check git state"
}
```

Result:

```text
G-EDH3XC checkpoint
OWED
  runner: Q-JP7D58 test H-HTNTWC (h-separate-review)
  runner: Q-JP7D58 test H-0BRHH0 (h-one-experiment)
  runner: Q-JP7D58 test H-4VKM63 (h-raise-budget)
  runner: Q-ZTTBCH test H-8NJHYJ (h-commit-fix)
  runner: Q-ZTTBCH test H-M5Q0D4 (h-reserve-last)
  runner: Q-WZ3QQY test H-JNM729 (h-limit-then-commit)
  runner: Q-WZ3QQY test H-WXR74B (h-probe-closes)
  runner: Q-V99SRH test H-YT0KRB (h-last-increment)
  runner: Q-V99SRH test H-GKJ8KG (h-whole-candidate)
AWAITING OWNER
RECENT
  2026-11-12 14:45 claude-grill-141 edge.add G-EDH3XC — The parent issue of #141.
  2026-11-12 14:45 claude-grill-141 edge.add G-EDH3XC — The issue this design Goal serves.
?? .tree/events/claude-grill-141.jsonl
?? .tree/events/owner-standin.jsonl
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-1/wt && git add investigations/goal-prompt/.tree/events && git commit -q -m \"docs: add the review-design Goal for #141 with its four questions and their rivals\" && git log --oneline -1 && cd investigations/goal-prompt && hyp prompt G-EDH3XC --actor claude-g141-1 --branch work 2>&1",
 "description": "Commit the grill events and print the /goal prompt"
}
```

Result:

```text
120c0c1 docs: add the review-design Goal for #141 with its four questions and their rivals
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-EDH3XC: read it with hyp show G-EDH3XC, read hyp status G-EDH3XC, and name --goal G-EDH3XC on every hyp checkpoint.
Done when the latest printed hyp status G-EDH3XC shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, never merging it. It changes no file in cmd/, internal/, skills/ or docs/.
Read only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.
Never: commit to main or merge a PR.
Never: replace the installed hyp binary.
Never: merge a trial's events into investigations/goal-prompt/.tree/events.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch work.
```

## Grill

The grill is done. G-EDH3XC is in `investigations/goal-prompt`, and `hyp check` prints 0 errors. It still prints 9 warnings, which were already there from earlier Results.

## What is in the tree

- **Goal:** G-EDH3XC (`review-design`) holds the five lines. It is linked to #141 and #113. The only owner write was the `hyp add goal` you approved, under `owner-standin`.
- **Four questions**, each with today's rule as the base rival and the #141 rivals (H-whole-candidate is marked "expected to lose on cost, tested anyway"):

| Question | Rivals |
|---|---|
| Q-JP7D58 (how a review counts against the cap) | H-HTNTWC (base), H-0BRHH0 (one experiment), H-4VKM63 (cap raised to 7) |
| Q-ZTTBCH (one Result left, fix pending) | H-8NJHYJ (base), H-M5Q0D4 (reserve last) |
| Q-WZ3QQY (a finding whose probe already passed) | H-JNM729 (base), H-WXR74B (probe closes) |
| Q-V99SRH (what the pre-select review reads) | H-YT0KRB (base), H-GKJ8KG (whole candidate) |

- **Checkpoint:** I wrote it with "end of grill" under `claude-grill-141`.
- **Commit:** `120c0c1` on branch `work` holds `.tree/events/`. It uses the conventional format and has no AI trailer, as `AGENTS.md` requires.
- **OWED:** the runner owes a test of each of the 9 rivals.

## Hand-off prompt

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-EDH3XC: read it with hyp show G-EDH3XC, read hyp status G-EDH3XC, and name --goal G-EDH3XC on every hyp checkpoint.
Done when the latest printed hyp status G-EDH3XC shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: investigations/goal-prompt/.tree/ (events and evidence), in commits on one local branch; the run may push it and open or update one PR, never merging it. It changes no file in cmd/, internal/, skills/ or docs/.
Read only: skills/, docs/, cmd/, internal/, every investigations/ tree except goal-prompt, and the goal-prompt events before this Goal.
Never: commit to main or merge a PR.
Never: replace the installed hyp binary.
Never: merge a trial's events into investigations/goal-prompt/.tree/events.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch work.
```

## The default cap is too small for this run

The cap of 5 Results is the default that #141 is about. This Goal owes at least 9 Experiments, one per rival, and 45 trials. A run with 5 would stop partway and name what it left. Pass `--experiments N` to `hyp prompt`, for example `--experiments 12`, to give it room. I left the prompt as hyp printed it, as the skill requires. Which N to use is your call.

## Left for you

- **The build Goal.** Add it after the selects by running the grill again. It owns:
  - the `hyp-goal-run` text on how a review is recorded and whether it counts;
  - the cap text in `docs/how-hyp-improves-itself.md:29` (run 4's F8);
  - any change to the default in `hyp prompt`.
- **The log.** Run `hyp prompt` on the log the run starts from. G-EDH3XC's add is in commit `120c0c1` on `work`, so the run must start from that branch.
- **Not published.** Nothing was pushed, and I opened no PR.
