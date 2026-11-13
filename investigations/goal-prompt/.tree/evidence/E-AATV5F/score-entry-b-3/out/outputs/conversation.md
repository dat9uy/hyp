# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-entry-b-3/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and finding the stores.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-entry-b-3/tmp/issue-141.md"
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
 "command": "ls investigations/ && which hyp; for d in investigations/*/; do echo \"== $d\"; (cd $d && hyp status 2>&1 | head -15); done",
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
/Users/datguy/.hyp-sandbox/g146-entry-b-3/bin/hyp
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
  G-D1CH1R#3 covered on Q-PC04ED; ran in R-A4FZG7, R-CX1G1C, R-7XVMDC, R-Q9WTZB, R-KBW1
[... 3691 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "git log --oneline | head -5; git branch -a | head; cd investigations/goal-prompt && hyp status G-ZBM2RE 2>&1 | head -60; hyp status G-GZR6C6 2>&1 | head -30; grep -rn \"113\" .tree/events 2>/dev/null | head -5 | cut -c1-300",
 "description": "Inspect goal-prompt store Goals and #113 link"
}
```

Result:

```text
9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
04142f3 fix: test that a bad block line in a store outside every repository exits 1
d2d5d14 fix: test that an invalid block in a store outside every repository exits 1
* work
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
  2026-11-12 03:23 claude-g113-3 result.record R-HXEBAF — Records E-ZG1BC
[... 4154 more characters]
```

### Tool Read (cwd: investigations/goal-prompt)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/skills/hyp-goal-run/SKILL.md"
}
```

Result:

```text
1	---
2	name: hyp-goal-run
3	description: "Run a hyp tree under an owner's Goal and met rule as an unattended /goal, for a discovery Goal or a build Goal: act only on what OWED says the runner owes, test before asking, write readings the owner can answer, reassess the targets under the owner's answers, and stop with a report the owner can read from hyp status. Use when a /goal prompt names this skill, such as the one hyp prompt prints."
4	---
5	
6	The `/goal` prompt names the tree, your actor, the Goal, the stop and the Goal body's constraints; `hyp prompt <goal-id> --actor <actor>` prints it from the log. This skill says how to work until that stop. Use it with the hyp-prototype skill, which says how to add Questions, Hypotheses, Experiments and Results, and the hyp-verification skill, which says how to run an Experiment.
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
20	- Use the Goal's must-fail cases as known-answer flaw cases, in the o
[... 12317 more characters]
```

### Tool Grep (cwd: investigations/goal-prompt)

```text
{
 "pattern": "Results recorded|Experiments have been|experiments",
 "path": "/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal",
 "glob": "*.go",
 "output_mode": "content",
 "head_limit": 30
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/render/show.go:104:		r.experiments(w, id)
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/render/show.go:551:// experiments lists every Experiment under question q, grouped by parent
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/render/show.go:556:func (r renderer) experiments(w func(string, ...any), q string) {
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/render/show.go:564:			w("experiments:")
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/render/show.go:580:				w("experiments:")
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/recipe/recipe.go:451:// copies hash differently are left out, so Results recorded under them read
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt.go:123:	if *experiments < 1 {
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt.go:124:		return usageErr("--experiments must be at least 1")
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt_test.go:63:	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 Results have been recorded under claude-run-1,")
/Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/internal/cli/prompt_test.go:331:	want(t, hyp(t, "prompt", g, "--actor", "r", "--experiments", "0"), 2, "--experiments must be at lea
[... 203 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt && sed -n 95,200p internal/cli/prompt.go; grep -n -i \"review\\|verifier\\|limit\\b\" docs/glossary.md | head -40; ls docs/adr | tail -8; grep -rn \"goal-run\\|prompt.go\\|hyp prompt\" docs/*.md | cut -c1-200 | head -20",
 "description": "Read prompt stop clause, glossary review terms, docs referencing prompt"
}
```

Result:

```text
		}
		d = up
	}
}

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
	// the store's place in its reposit
[... 14220 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat docs/adr/0021-a-skill-change-is-a-build-goal.md | head -60; grep -n -i \"^\\*\\*.*cut\\|store cut\\|\\*\\*Runner\\|\\*\\*Instrument\" docs/glossary.md | cut -c1-600; cd investigations/goal-prompt && hyp show G-ZBM2RE 2>&1 | sed -n 1,60p | cut -c1-900",
 "description": "Read ADR 0021, cut/runner glossary, and prior Goal body and lines"
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
[... 15712 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt && shasum -a 256 skills/hyp-goal-run/SKILL.md; git hash-object skills/hyp-goal-run/SKILL.md >/dev/null; cd investigations/goal-prompt && for k in 8JQPWG JHBCWS HQWTA0 C3GGGJ 7QQ2DP; do echo \"-- $k\"; grep -l \"$k\" .tree/events/*.jsonl | head -2; grep -h \"\\\"node\\\":\\\"[0-9A-Z]*$k\\\"\" .tree/events/*.jsonl | cut -c1-400 | head -1; done; ls /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/.claude/skills/ak-skill-creator/ 2>&1 | head -20; ls /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/.claude/skills/ak-skill-creator/scripts 2>&1 | head -20",
 "description": "Verify cited Result ids, hash SKILL.md baseline, list eval tooling"
}
```

Result:

```text
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
-- 8JQPWG
.tree/events/claude-g113-1.jsonl
.tree/events/claude-g113-2.jsonl
{"v":1,"id":"01M75T1YRF2R2Z99DX0E8JQPWG","ts":"2026-11-11T17:21:27.311404Z","actor":"claude-g113-1","type":"result.record","node":"01M75T1YRF2R2Z99DX0E8JQPWG","payload":{"parent":"01M75SVF189290RGG0GQ7QQ2DP","verdicts":{"01M75SH0P5561NTGAW8V3M3QAS":"contradicts"},"observation":"The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and fl
-- JHBCWS
.tree/events/claude-g113-3.jsonl
{"v":1,"id":"01M76VXNN2ZW5WAXJ194JHBCWS","ts":"2026-11-12T03:13:18.498579Z","actor":"claude-g113-3","type":"result.record","node":"01M76VXNN2ZW5WAXJ194JHBCWS","payload":{"parent":"01M76VPHDJDBC3E22WR3TZTG0G","verdicts":{"01M76V0YRS94HE0SRV9TFQ103R":"supports"},"observation":"The fresh review of 3cfd08d reports the new CRLF test red under three mutations of prompt.go (strip \\r, normalise CRLF, ref
-- HQWTA0
.tree/events/claude-g113-1.jsonl
{"v":1,"id":"01M75RTZK05VMXAX2ER1HQWTA0","ts":"2026-11-11T17:00:10.208014Z","actor":"claude-g113-1","type":"node.add","node":"01M75RTZK05VMXAX2ER1HQWTA0","payload":{"kind":"Experiment","parent":"01M75RTFR9JKPARCH3NAQ5P9R9","title":"Gate 10566d4 against G-ZBM2RE's seven lines","body":"Supports H-Q5P9R9 only if all hold on 10566d4, else contradicts. #1-#5: go vet ./... and go test ./... -count=1 pas
-- C3GGGJ
.tree/events/claude-g113-1.jsonl
{"v":1,"id":"01M75SAW63CZAX0KXXXMC3GGGJ","ts":"2026-11-11T17:08:51.011007Z","actor":"claude-g113-1","type":"node.add","node":"01M75SAW63CZAX0KXXXMC3GGGJ","payload":{"kind":"Experiment","parent":"01M75RTFR9JKPARCH3NAQ5P9R9","title":"Review 10566d4's diff against G-GZR6C6 and the lines; probe each finding","body":"Records the fresh review of git diff 2881e7d 10566d4 (findings F1-F10, review.md) and 
-- 7QQ2DP
.tree/events/claude-g113-1.jsonl
{"v":1,"id":"01M75SVF189290RGG0GQ7QQ2DP","ts":"2026-1
[... 857 more characters]
```

## Grill

I read the issue, the skill, `hyp prompt`, the glossary, ADR 0021, G-ZBM2RE and the two runs' logs. Nothing is written yet. These are Round 1's questions. Round 2 is the routing table, once you've answered.

## What I checked

- **Store.** The issue is "Part of #113". G-GZR6C6 and G-ZBM2RE, both for #113, are in `investigations/goal-prompt`. The new Goal goes in that store. That is not a choice, so there is no question on it.
- **The cuts exist.** R-8JQPWG is the review Result of `b4f23b2` (run 1's fourth Result, one left). R-JHBCWS is run 3's review of `3cfd08d`. Both are in the log, so S1 and S2 are real cuts.
- **The unit of cost is a Result.** `hyp prompt` prints "N Results have been recorded under A", and the flag `--experiments` counts Results. An Experiment can hold several Results. So H-one-experiment saves budget only if the gate and the review go into **one Result**. I'd reword that rival as "one Result".
- **Text now on main.** The review and sort text (SKILL.md lines 34–35) was unmerged when the issue was written. PR #140 has merged it, at `9c8e498`. The baseline is `skills/hyp-goal-run/SKILL.md` with sha256 `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e`.
- **Derivation guard.** I propose no stored field. replay cannot tell which Result is a review: it has no marker, and a review is `--recipe none --rubric reading` like any gate. So "a candidate lacks a review" cannot be a refusal or warning in the binary. Every rival is skill text, except H-raise-budget, which changes a default in `prompt.go`.

## Round 1

**1. Your actor name.** The grill's owner writes go under it, for example `owner-dat`, which the goal-prompt store already uses. I run as `claude-grill-141`. *Recommend: `owner-dat`.*

**2. One Goal, one DesignQuestion, and what is in it.** The issue's rivals, plus H-probe-closes from your third comment, all answer one question: *what change makes a budget-bound run end with its final candidate reviewed or named?* Each is measured on the same two cuts, so one run can cover every candidate for every line. That makes one build Goal and one DesignQuestion with these rivals:
- H-one-result (your H-one-experiment, reworded)
- H-name-at-stop
- H-reserve-last
- H-raise-budget
- H-probe-closes

H-whole-candidate answers a different question: what the review reads. No S1 or S2 outcome separates it from the status quo, so it fails the grid test. I'd make it a separate Goal that you add, and I name it in the checkpoint.

The tree should say which rivals hold alone, so each is run alone. If none holds on both cuts, the Goal body tells the run to add one combination of the rivals that each held on one cut. Name-at-stop is always a candidate for that combination.
- *Recommend:* the five rivals above, H-whole-candidate as its own Goal, and the combination rule in the body.
- *Ask:* do you agree? Should the run test H-raise-budget, or has run 2 settled it? Run 2 had a fresh budget of 5 and still ended unreviewed. If you say it's settled, it is recorded as a decided rival with your session as evidence. If not, it is a `prompt.go` commit measured by S1 and S2 with the larger budget. *Recommend: test it, because it is a one-line change.*

**3. What the Goal's lines demand.** I draft four, in the build-Goal form:
- **[P] L1.** On S1 (the log through R-8JQPWG, one Result left, a fix commit pending) and on S2 (through R-JHBCWS), a consumer run under the candidate skill ends with a review Result of its final candidate, **or** a stop checkpoint that names that candidate as unreviewed. A checkpoint that says "all review limits closed" without naming it fails. The baseline skill, by the sha256 above, ends with neither on S1, and the candidate's runs differ from it.
- **[P] L2.** `hyp-goal-run` states where a review is recorded, whether it counts against the stop clause, and which of the review and the stop clause yields at the last Result. This is your third comment.
- **[P] L3.** If a candidate changes what `hyp prompt` prints, G-ZBM2RE's lines 1–5 hold with only that clause changed. The prompt tests are updated for it, and `go test` passes.
- **[P] L4.** Each changed SKILL.md passes `quick_validate.py` and `lint_cruft.py --routing` (same pins as G-ZBM2RE#7), and skillcheck passes.

L1 has a catch. Name-at-stop passes it by itself, so it cannot tell "reviewed" from "named".
- *Option A:* L1 reads "reviewed or named" (Done-means verbatim). Each Experiment's rule records which of the two happened per rival and cut, and the select is your trade-off between a gap that is visible and a gap that is closed.
- *Option B:* L1 demands "reviewed". Name-at-stop alone then fails it and only survives in a combination.
- *Recommend A.* The rivals are separated by which outcome they produce, and the choice between the outcomes stays yours.

**4. How the run tests the candidates.** Per ADR 0021 that is the Goal's choice. I propose ak-skill-creator's eval flow: S1 and S2 become cases in `evals.json` format. Each case holds the log cut as a fixture and the pending state. Matched runs of the original skill and the candidate are graded by assertions "ends reviewed or named". L3 and L4 use `go test` and the pinned scripts.
- *Ask:* use that flow? Which model for the consumer runs? How many trials per arm per case (*recommend 3*)? What budget?
- *Ask:* where do the cases live? *Recommend:* as evidence files under `.tree/evidence/<E-id>/`, with no `recipes/` bundle. That keeps `Never: edit recipes/` as in G-ZBM2RE, and the run adds no self-review recipe. The issue says none is involved. ADR 0021's named place, `recipes/goal-run-eval@1`, would be a self-review recipe and would need `recipes/` in `May change:`.

**5. The constraints block.** My draft:
- **May change:** this repository, in commits on the branch the prompt names, conventional messages, no AI trailers. The run may push the branch and open or update one PR, and never merges it. It changes `skills/` (hyp-goal-run, and any hand-off text in hyp-owner-grill or hyp-owner-answer that the change makes wrong), `docs/`, `AGENTS.md` and `README.md`, plus `internal/cli/prompt.go` and its tests **only** for H-raise-budget or a reworded stop clause.
- **Read only:** `plans/`, every `investigations/` tree except goal-prompt, and ak-skill-creator under `.claude/skills`, pinned by sha256.
- **Never:** commit to main or merge a PR; replace `~/.local/bin/hyp`; edit `recipes/`, `rubrics/`, another tree or another repository; write under an owner's actor.

The conflict to settle is that G-ZBM2RE#1 and `prompt_test.go` pin "N Results … default 5" byte for byte. A candidate that changes the default or the clause turns those tests red. *Ask:* may a candidate change `hyp prompt`'s printed stop clause and default, and update those tests? *Recommend: yes.* The issue's scope names the stop clause, and L3 holds the rest of the output fixed.

**6. Out-of-Goal items.** I will name these in the checkpoint, not hold them:
- H-whole-candidate as its own Goal (see 2).
- `--experiments` counts Results, so its name is wrong.
- `how-hyp-improves-itself.md` says hyp states no run cap, which was run 1's F8.

*Ask:* is that right, or should the flag name be a line here? *Recommend:* leave it, because it serves a different purpose.

Before Round 2's routing table I'll also run the second-reader check and the preflight scan. The reader gets the rules and code, not my conclusion, on three scenarios: a merged log from two actors, a base that moves mid-run, and S2 with budget 2. Any owner decision they raise comes to you first.

## Owner, message 2

1. Use `owner-standin`, not `owner-dat`.

2. One Goal and one question is fine. Test H-raise-budget, as you recommend. What does H-probe-closes mean? You haven't explained the label. I expect H-whole-candidate to lose, because one cumulative review costs a lot, so record it as the expected loser. If S1 and S2 can't separate it from the status quo, a separate Goal is fine. The combination rule in the body is fine.

3. I agree with A. Two questions for each of L1–L4:
- How does the line serve the aim that a review must not easily run out of the run's budget?
- How does it serve the aim that a review must not go on and on over small changes?

4. Replay the S1 and S2 cuts on Sonnet with 3 trials per candidate per cut. I don't want a full replay of a whole build from its start for every trial. Take your recommendations on the eval flow, the budget, and where the cases live.
- Could the ak-context-engineering skill help with how we evaluate?
- Each candidate may cost at most 1.5 times what the base costs at the same starting point. How is that cost computed?
- If you use a median, how is it computed? A trial that hangs counts as the most expensive.

5. Yes, a candidate may change the printed stop clause and default and update those tests. The rest of the block is fine.

6. Agreed.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/.claude/skills && ls; ls ~/.claude/skills 2>/dev/null | head -30; cd ak-skill-creator && sed -n 1,60p scripts/session_usage.py; grep -n \"add_argument\\|usage\\|tokens\\|timeout\\|median\\|trials\\|runs\" scripts/eval_skill.py | head -60; ls references assets | head -30",
 "description": "Find eval tooling options, cost/usage fields, and context-engineering skill"
}
```

Result:

```text
ak-skill-creator
hyp-goal-run
hyp-owner-answer
hyp-owner-grill
hyp-prototype
hyp-verification
hyp-goal-run
hyp-owner-answer
hyp-owner-grill
hyp-prototype
hyp-verification
public-doc-voice
synced
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
        path = Path(raw)
      
[... 4565 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "grep -n -i \"median\\|timeout\\|hang\\|tokens\\|paired\\|matched\\|trials\\|fixture\\|cwd\\|context\" references/creator-consumer-evaluation.md references/evaluation-tools.md references/benchmark-optimization-guide.md | cut -c1-260 | head -50; grep -rli \"context-engineering\\|context engineering\" /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt/.claude /Users/datguy/.claude/skills 2>/dev/null | head",
 "description": "Check eval docs for cost, median, hang handling, and context-engineering skill"
}
```

Result:

```text
references/creator-consumer-evaluation.md:6:## Matched briefs
references/creator-consumer-evaluation.md:22:2. Run each in fresh context with equal brief, tools and budget. Save generated
references/creator-consumer-evaluation.md:32:no-skill conditions with identical fixtures and tool permissions. Exercise the
references/creator-consumer-evaluation.md:34:context from consumer context; include another model family when available.
references/creator-consumer-evaluation.md:49:success/invariants, then corrections, tokens, latency and tool calls. Do not hide
references/creator-consumer-evaluation.md:56:conditions first; label a configuration experiment separately from a prose change.
references/evaluation-tools.md:5:commands. Run from this skill's root or use the script's absolute path from any cwd.
references/evaluation-tools.md:31:context with explicit output location. Do not assume provider credentials, model
references/evaluation-tools.md:46:Optional observed metrics: `--tokens`, `--duration-ms`, `--tool-calls`, `--corrections`.
references/evaluation-tools.md:67:    "input_tokens": null, "output_tokens": null,
references/evaluation-tools.md:68:    "cache_read_tokens": null, "cache_write_tokens": null,
references/evaluation-tools.md:83:cached tokens. Do not sum fields using an assumed universal billing formula.
references/evaluation-tools.md:92:The importer neither changes model/effort nor operates provider caches. Use verified
references/evaluation-tools.md:113:and cost provenance. They expose `comparison_context`, `missing_evidence`,
references/evaluation-tools.md:119:matched repetitions. Inspect those outcomes and quality before interpreting means;
references/evaluation-tools.md:123:their aggregate tokens as a proven cost improvement.
references/benchmark-optimization-guide.md:16:`references/testing-and-iteration.md`. For authoring changes, use
references/benchmark-optimization-guide.md:30:Use fabricated leakage fixtures and verify actual outputs/side effects, not r
[... 142 more characters]
```

### Tool Bash (cwd: .claude/skills/ak-skill-creator)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt && for c in b4f23b2 3cfd08d 10566d4 5c5902d 9c8e498; do printf \"%s \" $c; git cat-file -t $c 2>&1; done; sed -n 397,399p docs/glossary.md | cut -c1-900; grep -n \"store-cut\\|--store\\|cut\" docs/sandbox-spec.md | head -15 | cut -c1-220; hyp sandbox create 2>&1 | head -5",
 "description": "Verify cut commits exist and store-cut facilities in sandbox"
}
```

Result:

```text
b4f23b2 commit
3cfd08d commit
10566d4 commit
5c5902d commit
9c8e498 commit
**Store cut**:
A copy of a tree's store as it stood at one event, given to a sandbox.
_Avoid_: snapshot (that is a different thing: a capture of a live session)
46:- No shell runs a command. Its first word is a relative path containing `/` to an executable script inside the bundle, and the rest are plain words; quotes, `$`, pipes, redirects and globs are refused.
158:A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (s
usage: hyp sandbox create ID --host H --base REV --recipe NAME@VERSION|none [--recipe NAME@VERSION]... [--recipes DIR] [--store SRC --cut EVENT] [--inputs FILE] [--patch ARM=FILE]... [--hyp BIN]
```

## Grill

Your answers are applied: owner actor `owner-standin`, one Goal and one DesignQuestion, H-raise-budget is tested, the combination rule is in the body, option A for L1, a candidate may change the stop clause and update its tests, and the constraints block and out-of-Goal items are as drafted. Nothing is written yet. Three decisions remain.

## What H-probe-closes means

That label is from your third comment, and I used it without defining it.
- **Today:** a review finds a case, for example N1 on `3cfd08d` ("a case of line 3 that no test runs"). The runner probes it and the code passes. The runner then records it as a `--limit`, because a limit is an untested case. A fix commit adds the test, and that commit owes its own review. This is the chain that never ends.
- **H-probe-closes:** because the probe ran the case and the case passed, the review's Result closes it (`--closes`), and no commit follows. A limit stays only for a case no check ran, or one the candidate fails.
- **Principle:** ADR 0016 says a limit names the next case to run. A case a check already ran has no next run.

## H-whole-candidate

The two cuts can separate it once there is a cost line (L5 below). It reads the cumulative diff, so it should cost more tokens than the base. That is your expected loss, and a number now decides it. So I'd keep it in this DesignQuestion as a sixth rival, with "expected to lose on cost: the owner's expectation, 2026-11-12" in its body. No separate Goal is needed. L1 will probably also fail it, because its ending matches the base.

## How L1–L4 serve your two aims

| Line | Review not running out of budget | Review not going on over small changes |
|---|---|---|
| L1 (ends reviewed or named) | Yes, directly, on S1 and S2. Naming is visibility, not protection. | No. A review, commit, review loop that ends by naming passes. |
| L2 (skill states where a review is recorded, whether it counts, who yields) | Indirectly: the rule is written once, so the run does not invent it. | Indirectly, the same way. It sets no limit on rounds. |
| L3 (printed clause changes only as the candidate says) | Only for H-raise-budget, which changes the clause. | No. |
| L4 (validators pass) | No. | No. |

Aim 2 is covered by none of them. I propose two more lines.
- **L5 [N], cost.** On S1 or S2, a candidate whose median trial cost exceeds 1.5 times the base's median on the same cut is contradicted. This bounds a runaway loop, and it is what separates H-whole-candidate.
- **L6 [N], small changes.** On S2, no commit is added whose only purpose is a test or text for a case the run's own probe already passed on the candidate. This is aim 2 as a must-not. Consequence: name-at-stop alone and raise-budget would add that same commit, so they fail L6 on their own. They survive only inside a combination, which the body's combination rule already allows. Probe-closes and reserve-last can hold alone.

## How the eval is built

- **Replay.** `hyp sandbox create … --base b4f23b2 --store <goal-prompt> --cut <R-8JQPWG's event>` gives S1. S2 uses `--base 3cfd08d` and R-JHBCWS's event. Both commits exist. Each trial starts at the cut under the cut's actor, so the printed clause still counts the Results already recorded. Nothing before the cut is replayed.
- **`ak-context-engineering` is not installed here.** It appears only as a name in one routing-case file. I can't use it, and I won't cite a method from memory. If you meant another skill, tell me its name.
- **What ak-skill-creator provides.** It runs matched original and candidate trials and records the five token fields separately. It says not to add them with a billing formula, and it summarises with means. It has no median and no hang rule, so the Experiment's verdict rule has to define both.
- **Cost of one trial.** The trial's fresh tokens from the cut to its stop: `input + cache_write + output + reasoning`. `cache_read` is reported but excluded, because it depends on timing and not on the work done.
- **Compared with the base.** The base is the original skill (sha256 `cb15e2f2…`) run from the same cut, with the same model and the same three trials. For each cut, the ratio is the candidate's median divided by the base's median. A candidate passes L5 only if the ratio is at most 1.5 on both cuts. Cuts are never pooled.
- **Median of 3.** Sort the three costs and take the second. There is no averaging.
- **A hang.** A trial that reaches no stop clause within the timeout counts as cost ∞, which sorts last. Example: the base costs 400k, 520k and a hang, so its median is 520k. A candidate at 600k, 700k and 790k has median 700k, a ratio of 1.35, and passes. A candidate at 600k and two hangs has median ∞ and fails. If the base itself has two hangs, the ratio is undefined and the Result is `inconclusive`, not `supports`.

## Decisions for you

1. **L5 (cost).** Add it? Confirm the cost metric above and a per-trial timeout of 30 minutes. Run 1 took about 10 minutes from S1 to its stop. *Recommend: yes, with fresh tokens and a 30-minute timeout.*
2. **L6 (no commit for a case the probe passed).** Add it as a line, with the consequence above? The alternative is to record each rival's commit count per cut in the verdict rule and let the select weigh it. *Recommend: yes, as a line, since you wrote it as a must-not.*
3. **H-whole-candidate as a sixth rival** in this DesignQuestion, with the expected loss recorded, instead of a separate Goal. *Recommend: yes.*

When you have answered, I'll run the second-reader check. It covers the merged-log, moving-base and S2 budget-2 scenarios, plus the preflight scan. Then I'll send the routing table for you to confirm before any write.

## Owner, message 3

1. Yes to L5. The cost metric and the 30-minute timeout are fine. The median of 3 and the hang rule are fine too.

2. Yes to L6 as a line, with the consequence you describe.

3. Yes, H-whole-candidate goes in as the sixth rival with the expected loss recorded.

I only wondered whether ak-context-engineering could help with the evaluation. If it isn't installed, drop it. Go ahead with the second-reader check.

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader check of grill rules",
 "subagent_type": "general-purpose",
 "model": "opus",
 "run_in_background": false,
 "prompt": "You are an independent reviewer. Read-only: change no file, write no hyp event, run no `hyp add`/`record`/`checkpoint`. Repo: /Users/datguy/.hyp-sandbox/g146-entry-b-3/wt (Go; binary `hyp` on PATH). Read docs/glossary.md, docs/replay-spec.md (esp. §7, §8, §8.1, §8.2), docs/sandbox-spec.md, skills/hyp-goal-run/SKILL.md, internal/cli/prompt.go, and the store investigations/goal-prompt (`hyp status`, `hyp show G-ZBM2RE`, events in .tree/events/).\n\nRULES SETTLED (a planned new Goal in investigations/goal-prompt; nothing is written yet). Do NOT assume these are right; try to break them.\nGoal: build Goal for issue #141. Candidate = a commit on a branch that changes skills/hyp-goal-run/SKILL.md (and, for one rival, internal/cli/prompt.go's `--experiments` default / printed stop clause, plus tests). One top-level DesignQuestion: \"what change makes a budget-bound run end with its final candidate reviewed or named?\" Six rival Hypotheses, each a commit tested alone: one-result (a candidate's review is recorded in the gate's single Result), name-at-stop (stop checkpoint names each unreviewed candidate), reserve-last (with one Result left, review the current candidate instead of committing another fix), raise-budget (default of --experiments rises, e.g. 7), probe-closes (a case the run's own probe ran and passed on the candidate is closed with --closes in the review's Result, not recorded as a limit, so no fix commit follows), whole-candidate (the review reads the cumulative diff from the Goal's base; recorded as the owner's expected loser). If no rival holds alone on both cuts, the run adds one combination of rivals.\nTwo scenario cuts (store cuts via `hyp sandbox create --base REV --store SRC --cut EVENT`): S1 = the goal-prompt store through result R-8JQPWG (base commit b4f23b2, actor claude-g113-1, 4 of 5 Results used, a fix commit pending); S2 = through R-JHBCWS (base 3cfd08d, actor claude-g113-3). 3 Sonnet trials per candidate per cut, each starting at the cut, same actor, with the prompt printed by the candidate's own `hyp prompt`.\nProposed lines (owner writes them; never edited):\nL1 [P] On S1 and on S2 a consumer run under the candidate skill ends with a review Result of its final candidate, or a stop checkpoint that names that candidate as unreviewed; a checkpoint saying \"all review limits closed\" without naming it fails; the baseline skill (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) ends with neither on S1 and the candidate's runs differ from it.\nL2 [P] hyp-goal-run states where a review is recorded, whether it counts against the stop clause, and which of the review and the stop clause yields at the last Result.\nL3 [P] If a candidate changes what `hyp prompt` prints, G-ZBM2RE lines 1-5 hold with only that clause changed; prompt tests are updated; go test passes
[... 2807 more characters]
```

Result:

```text
Agent terminated early due to an API error: You've hit your 5-hour limit · resets 12:50am (Asia/Saigon) (error type rate_limit, HTTP 429, request id req_011CgyE8iwewLEkJZyFmTZzx, model sent to the API: claude-sonnet-5-5)
```

## Grill

You've hit your 5-hour limit · resets 12:50am (Asia/Saigon)
