# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-b-3/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/tmp/issue-141.md"
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
 "command": "hyp status 2>&1 | head -60; ls; ls investigations",
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
checkpoint-store
checkpoint-store-design
goal-prompt
id-citation
recipe-base
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "ls -a investigations/goal-prompt; cd investigations/goal-prompt && hyp status 2>&1 | head -80; hyp --version 2>&1 | head -2; git -C ../.. log --oneline | head -5",
 "description": "Inspect goal-prompt tree status"
}
```

Result:

```text
.
..
.tree
evals
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-11 16:47 claude-grill-113
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-12 03:28 claude-g113-4
AWAITING OWNER
usage: hyp <command> [flags]

9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
04142f3 fix: test that a bad block line in a store outside every repository exits 1
d2d5d14 fix: test that an invalid block in a store outside every repository exits 1
```

### Tool Read (cwd: investigations/goal-prompt)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/skills/hyp-goal-run/SKILL.md"
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
 "pattern": "Results recorded|experiments",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal",
 "glob": "!*_test.go",
 "output_mode": "content",
 "head_limit": 30
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/recipe/recipe.go:451:// copies hash differently are left out, so Results recorded under them read
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/render/show.go:104:		r.experiments(w, id)
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/render/show.go:551:// experiments lists every Experiment under question q, grouped by parent
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/render/show.go:556:func (r renderer) experiments(w func(string, ...any), q string) {
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/render/show.go:564:			w("experiments:")
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/render/show.go:580:				w("experiments:")
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/prompt.go:123:	if *experiments < 1 {
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/prompt.go:124:		return usageErr("--experiments must be at least 1")
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/internal/cli/cli.go:86:  prompt G-id --actor A [--branch B] [--pr N] [--experiments N]   the /goal prompt for a Goal, from its body's constraints block
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp status G-ZBM2RE 2>&1 | head -60; hyp show G-ZBM2RE 2>&1 | head -70; ls .tree .tree/events | head; grep -rn \"review\" ../../docs/glossary.md | head -20",
 "description": "Show prior Goal, its lines and glossary review mentions"
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
[... 17838 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-3/wt && sed -n 95,200p internal/cli/prompt.go; grep -n \"limit\" internal/cli/record.go | head -30; grep -n \"hyp add\\|add exp\\|--limit\" docs/replay-spec.md | head -20",
 "description": "Read prompt stop clause code and limit flag locations"
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
[... 18487 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-3/wt/investigations/goal-prompt && grep -l \"8JQPWG\" .tree/events/*; hyp show R-8JQPWG 2>&1 | head -12; hyp show R-JHBCWS 2>&1 | head -8; ls ../../recipes ../../rubrics 2>&1 | head -20",
 "description": "Verify cut-point Results exist and list recipes"
}
```

Result:

```text
.tree/events/claude-g113-1.jsonl
.tree/events/claude-g113-2.jsonl
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
  2026-11-11 17:21 claude-g113-1 result.record R-8JQPWG — Records E-7QQ2DP under its pre-registered rule.
Result R-JHBCWS
path: G-ZBM2RE > Q-Z93VJ5 > H-FQ103R > E-TZTG0G > R-JHBCWS
status: recorded
observation: The fresh review of 3cfd08d reports the new CRLF test red under three mutations of prompt.go (strip \r, normalise CRLF, refuse a trailing \r) and the new no-block test red when the block check moves after the repository lookup; the new spec sentence is true of the code. N1: a mutation that checks only that a block exists before the lookup, and its validity after, passes every TestPrompt test. The probe with the 3cfd08d binary (785dd8d6a43871e7) ru
[... 475 more characters]
```

## Grill

I read issue #141 with its three comments, the `hyp-goal-run` skill, `hyp prompt`'s stop clause, and the `goal-prompt` tree. I've written nothing yet. These are the round 1 questions: I need your answers before I can compute round 2.

## What I found

- **The store.** The repo root has no `.tree`. Stores sit per topic under `investigations/`. `investigations/goal-prompt` holds G-GZR6C6 and G-ZBM2RE (both complete), and #141 is "Part of #113". Both cut points exist: R-8JQPWG for run 1, and R-JHBCWS for run 3.
- **What the budget counts.** `hyp prompt` counts Results recorded under the actor, not Experiments (`internal/cli/prompt.go:167`). One Experiment can hold several Results; E-GE085A did. So "one Experiment per candidate" is only a saving if it also means one Result.
- **No lines yet.** I won't propose any stored field. Every rival is run practice (skill text and a prompt clause), and the replay derives everything it needs from existing logs.
- **How I'd route your rivals.** The grill rule says a testable choice isn't put to you as a choice. Your rivals fall into three groups, and each group answers one question:

| DesignQuestion | Rivals | Scenario that separates them |
|---|---|---|
| DQ-A: how does a budget-bound run end with its last candidate reviewed or named? | H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget | G-ZBM2RE's cut after R-8JQPWG (one Result left, a fix commit pending), and run 2's cut after R-3YJ7T2 on a fresh budget |
| DQ-B: what does a review finding cost? | H-probe-closes against the current rule that a finding becomes a limit, then a commit (H-keep-limit) | Run 3's cut after R-JHBCWS (N1 probed and passing) |
| DQ-C: what does the last review read? | H-whole-candidate against the last increment (H-last-increment) | Review 04142f3 from the Goal's base and see whether it finds anything the increments missed |

H-whole-candidate and H-probe-closes don't answer DQ-A's question, so they get their own DesignQuestions, as the one-question rule requires. The routing table I owe you before any write comes after round 2.

## Round 1

**1. Where does the tree live?** *(your preference)*
- **Today:** `investigations/goal-prompt` is complete, merged in PR #140, and its logs are the scenario's source data.
- **Option A:** a new tree, `investigations/run-review`, with its own `hyp init`. The cuts are copied in as inputs, and `goal-prompt` stays untouched, as it should for a Goal tied to a merged PR.
- **Option B:** a third Goal in the `goal-prompt` store. Its Goal body would then have to say that the cuts are its own history.
- **Recommendation:** A. The skill's "same store" rule assumes one store per repository, and here each issue gets its own topic tree (see `checkpoint-store-design` and `checkpoint-store`).

**2. What is your owner actor name?** (for example `owner-dat`) I'll use it only in commands you approve.

**3. Should the design and its build be two Goals?**
- **This Goal:** it decides the practice. Its candidates are practice texts (a claim), tested on replayed cuts.
- **The build Goal:** it changes the commit (`skills/hyp-goal-run/SKILL.md`, `internal/cli/prompt.go` and the text they make wrong). It is a different kind of candidate, so the skill puts it in a separate Goal. You add it after the select. Your "Done means" bullet 2 ("after the change, a run that reaches its budget…") becomes one of its lines.
- **Recommendation:** two Goals. That is the same split as G-GZR6C6 and G-ZBM2RE.

**4. May a run end with the last candidate unreviewed if it names the gap?**
- **A (recommended):** yes, and the line allows it, as your "Done means" does. The tree separately measures which rivals reach "reviewed" and which only reach "named". Run 2 shows why: with findings 10, 8, 1, 2, a fixed budget can only be met by reaching "reviewed" if a review sometimes leads to no commit.
- **B:** no, "reviewed" is required. Then H-name-at-stop, H-one-experiment and H-raise-budget could contradict themselves on run 2's cut. The line would effectively select H-reserve-last or H-probe-closes before any run.
- **Basis:** A keeps the line to the purpose. B bakes a rival into the line.

**5. May a run select a candidate that has no review Result of its own?**
- **Today:** the skill says "before the stop", not "before the select". Run 4 selected 04142f3 after the review inside E-GE085A.
- **Option A (recommended):** never. The run stops and names the candidate, and the next run reviews it and then selects. This would be an `[N]` line.
- **Option B:** it may select and name the gap. A selected candidate then carries a known unreviewed diff.
- **Basis:** the select is the gate (ADR 0015). That gate is only worth having if the reader of the select can trust that a review ran first.

**6. A line about how a review is recorded.** Your comment says that whichever rival wins, `hyp-goal-run` must state how a review is recorded (which command, whether it has its own Experiment) and whether it counts against the budget.
- **Proposal:** one `[P]` line says every candidate's skill text states both, and states them inline with no ADR or `docs/` citation, per AGENTS.md. A check can read the candidate's text for this.
- Do you want this line as one `[P]`, or split into two (recording, budget)? I recommend one: it is one invariant, that the run need not invent the practice.

## What I'll do unless you object

- **Combinations.** The four DQ-A rivals are tested alone. A combination is added as its own Hypothesis only if no single rival holds, the same way `hyp-goal-run`'s "Before a trade-off" works. The Result states which rivals hold alone, as the issue asks.
- **Lines.** I expect three or four. Draft: `[P]` on the cut, the run ends with its last candidate reviewed or named in the stop checkpoint. `[N]` for a select on a candidate with no review Result (if you answer 5A). `[P]` for the skill text stating how a review is recorded and counted. I'll finish these after your answers. I'll also check whether a discovery Goal owes its lines on a DesignQuestion with rivals, or whether the lines apply only to candidates under a top-level question. If they don't, the lines move to the build Goal, as for a design record.
- **Constraints block, draft.**
  - *May change:* the new tree and its evidence.
  - *Read only:* `skills/`, `internal/`, `investigations/goal-prompt`.
  - *Never:* edit `skills/` or `internal/`, replace the installed binary, or write under an owner's actor.
- **Second reader and preflight.** Before the round closes, I'll have a fresh subagent walk the rules through a scenario we didn't draw (two Goals in one store, a merged log, a base that moves mid-run). I'll also scan the three texts for claims a run would contradict, for example `docs/how-hyp-improves-itself.md` on run caps, which the stop checkpoint already flagged as F8.

## Owner, message 2

1. Option B. #141 is part of #113, so the new Goal goes in the existing `goal-prompt` store, which already holds the history it is about. Its Goal body can say the cuts are its own history.
2. owner-standin.
3. Two Goals, agreed.
4. A, agreed.
5. A, agreed.
6. One `[P]` line, agreed.

On the routing table, move H-whole-candidate and H-probe-closes into DQ-A. I want every candidate fix under one question and all of them tested. I expect H-whole-candidate to lose, because one cumulative review costs a lot, so treat it as the expected loser.

Replay on Sonnet with 3 trials for each candidate and each starting point. Don't replay a whole build from its start for every trial; that costs too much. Can the ak-context-engineering skill help with how we evaluate?

Please say how each proposed line serves these two aims:
- A review must not easily run out of the run's budget.
- A review must not go on and on because of small changes.

A candidate may cost at most 1.5 times the base at the same starting point. How do you compute cost? A trial that hangs counts as the most expensive.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills ~/.claude/skills 2>&1 | head -30; ls investigations/goal-prompt/evals | head",
 "description": "Look for ak-context-engineering skill and existing evals"
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
goal-prompt-evals-2.json
goal-prompt-evals.json
```

### Tool Bash (cwd: .)

```text
{
 "command": "git show HEAD:skills/hyp-goal-run/SKILL.md | shasum -a 256; ls .claude/skills/ak-skill-creator .claude/skills/ak-skill-creator/scripts 2>&1 | head -30; grep -n \"owed\\|owes\" docs/replay-spec.md | grep -i \"design\\|top-level\\|question\" | head -8; grep -n \"trial\\|token\\|cost\" .claude/skills/ak-skill-creator/scripts/eval_skill.py | head -15",
 "description": "Hash base skill text; inspect eval tooling and owed-lines rule"
}
```

Result:

```text
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  -
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
90:**[added] readings**: a Result names a line of the Goal that a verdict rests on and how the run read it (`readings`). A reading's key is `<result-id>#r<n>`, and `n` counts the Result's `readings` from 1. Its states are open and closed, and both are derived. An owner answer closes it: its `answers` names the key, and the item's text is the owner's answer, which `hyp show` prints beside the closed reading (an old owner checkpoint closed it with `closes` and its `summary`). An owner changes an answer they gave with a later answer that names the key again: whether an item is a first answer or a revision is derived from the reading's state, so the old `closes` and `revises` merge (ADR 0020). `hyp answer` warns, at exit 0, when it revises a key that a line of another Goal cites. The answers are kept in merge order, the close first, and the latest is the answer in force; `hyp show` prints it beside the reading and each earlier one beneath, with the Result that applied it. A revision never reopens the reading and changes no event before it: an earlier Result keeps its verdicts, and a Hypothesis it contradicted stays contradicted until a new Result judges a successor. In an old log, a `revises` that sorts before its key's close is refused, and a second `closes` of an answered reading is rejected; two revisions both apply, the later
[... 10470 more characters]
```

## Grill

I haven't written anything yet. `ak-context-engineering` is not installed here, so I can't say whether it helps. Only `ak-skill-creator` is under `.claude/skills`, and `~/.claude/skills` holds hyp skills and `public-doc-voice`. If you give me its path, I'll read it before the routing table. What `ak-skill-creator` offers already fits the evaluation: `scripts/eval_skill.py` records tokens, duration, tool calls and cost for each run and each variant. G-ZBM2RE#6 already used it for matched original and candidate runs.

## Routing change

DQ-A now holds all six rivals, worded so each one answers it: "What practice makes a budget-bound run end with its last candidate reviewed or named, without the review going on?" The rivals are H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes and H-whole-candidate.

- With one top-level DesignQuestion, hyp owes every line on these six candidates (§8.1 lines), so the grid is one column.
- H-whole-candidate as the expected loser goes in its Experiment's body as a pre-registered prediction, not in a line. A line saying a rival must lose would be circular. The cost line can make it lose by rule.
- The base is the current `hyp-goal-run` text. Its sha256 is `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e` (`git show HEAD:skills/hyp-goal-run/SKILL.md` at 9c8e498). It is not a candidate. It is the baseline for cost.

## Replay design

- **Starting points.**
  1. Run 1 after R-8JQPWG (one Result left, a fix pending).
  2. Run 2 after R-3YJ7T2 (fresh budget of 5).
  3. Run 3 after R-JHBCWS (the probed case N1 passes).
- **Trials.** That is 3 starts × (6 candidates + base) × 3 trials = 63 Sonnet trials (`claude-sonnet-5-5`).
- **No full rebuild.** Each trial copies the log through the starting Result, checks out the commit under review in a scratch worktree, and runs only to the trial's stop. The candidate's skill text is installed, and `hyp prompt` prints the stop clause with that candidate's budget.
- **The real gate runs.** `go test` is cheap, so it runs for real.

## Cost

- **Unit: tokens, not Results.** The budget is counted in Results. Under the base at start 1, only 1 Result is left, so a 1.5× limit measured in Results would allow 1.5 Results and fail every rival that spends a second Result. Tokens measure the work instead.
- **Per trial.** The cost is the total tokens `eval_skill.py` reports for the trial's agent plus every subagent it spawns, from the start to the stop. Every variant is measured the same way.
- **Per candidate and start.** Take the mean of its 3 trials, compare it with the base's mean at the same start, and require it to be at most 1.5×. All three starts must pass.
- **A hang.** A trial that has not reached its stop at the cap is stopped and scored at the cap. The cap is 3× the largest of the base's three trials at that start, fixed in the Experiment's body after the base trials run and before any candidate runs. A hung trial has no stop, so it also fails the first line below.
- **Noise.** A 1.5× limit on a 3-trial mean is noisy. I'll state this in the Experiment's rule rather than hide it.

## Lines and the two aims

| # | Line | Aim 1: a review must not easily run out of budget | Aim 2: a review must not go on and on over small changes |
|---|---|---|---|
| 1 | `[P]` At each start, all 3 trials end with the final candidate either holding a review Result or named unreviewed in the stop checkpoint. | The floor: a gap is never silent, as in run 1. | A hang fails it. |
| 2 | `[P]` At starts 2 and 3, at least 2 of 3 trials end with the final candidate reviewed. | The real test. At those starts the last commit was not needed. | It rewards ending without more commits. |
| 3 | `[N]` No trial has a third consecutive cycle of review followed by a fix commit. | None. | The bound on cycles. Your findings of 10, 8, 1 and 2 were cycles. |
| 4 | `[N]` No trial selects a candidate with no review Result (your answer 5A). | A review can't be skipped to finish. | None. |
| 5 | `[P]` Every candidate's skill text states how a review is recorded and whether it counts against the budget, with no ADR or `docs/` cite (your answer 6). | The run needn't invent the practice again. | The same. |
| 6 | `[P]` Each candidate's mean cost at each start is at most 1.5× the base's (baseline sha256 above, which differs from every candidate). A hang counts as the cap. | A cheaper path leaves budget. | It bounds the cost of a loop. |

## Questions for round 2

1. **Line 2.** It narrows your answer 4A: "named" stays enough on line 1, but at starts 2 and 3 "reviewed" is required in 2 of 3 trials. Do you want it? Without it, no line says a rival can fail aim 1 by only naming the gap. I recommend yes.
2. **Line 3 bound.** Is "no third consecutive cycle" the right number? A bound is your call. I recommend 2 cycles allowed, which makes the third a violation.
3. **Cap.** Is 3× the base's largest trial acceptable?
4. **Reviewer model.** Should the reviewer subagent inside each trial also be Sonnet, so one model is tested? The real runs used a stronger reviewer. I recommend Sonnet.

After your answers I'll run the second reader and the preflight scan (including the `how-hyp-improves-itself.md` F8 claim about run caps). Then I'll show the routing table for you to confirm. No write happens before that.

## Owner, message 3

I can't give you a path to ak-context-engineering. I only wanted to know once whether it could help, so skip it if it isn't installed and use `eval_skill.py`.

The one-question routing, the baseline and the cost rule are fine. Counting tokens, taking the mean of 3 trials, and scoring a hang at the cap all work for me.

1. Yes, keep line 2.
2. Agreed: 2 cycles allowed, the third is a violation.
3. Agreed, the cap is 3× the base's largest trial.
4. Yes, Sonnet for the reviewer too.

Go ahead with the second reader and the preflight scan, then show me the routing table.

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader on grill rules",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer of a design for a hyp \"Goal\" (hyp is a Go CLI in /Users/datguy/.hyp-sandbox/g146-base-b-3/wt; read-only for you: change nothing). Do not trust my wording; check each fact against the source and report only what you verified, with file:line.\n\n## The design to review\nA new discovery-style Goal G-new will be added to the EXISTING store investigations/goal-prompt/.tree (which already holds Goals G-GZR6C6 and G-ZBM2RE and 5 actor logs under .tree/events/). Its single top-level DesignQuestion DQ-A asks: what practice makes a budget-bound unattended run (driven by skills/hyp-goal-run/SKILL.md and the stop clause printed by `hyp prompt`, internal/cli/prompt.go) end with its last candidate commit reviewed or named unreviewed, without the review going on and on. Six rival Hypotheses (practice texts): H-one-experiment (gate and review in one Experiment/Result), H-name-at-stop, H-reserve-last (with one Result left, review instead of another fix), H-raise-budget (default --experiments up), H-probe-closes (a case whose probe passes on the candidate is closed in the review's Result, not recorded as a limit then fixed by a commit), H-whole-candidate (the last review reads the cumulative diff from the Goal's base). The base (not a candidate) is the current skill text, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e.\nEvidence plan: replay from three starting points of G-ZBM2RE's history (after Result R-8JQPWG; after R-3YJ7T2; after R-JHBCWS): copy the log through that Result, check out the commit under review in a scratch worktree, install the candidate's skill text, print `hyp prompt`, run a Sonnet agent to its stop; 3 trials per candidate and per start; the reviewer subagent is Sonnet too. Cost per trial = total tokens (agent plus subagents), mean of 3, candidate must be <= 1.5x base mean at the same start; a hung trial scores the cap = 3x the largest base trial at that start.\nDraft lines (a Goal line is written once, never edited; each is a limit owed on every candidate of DQ-A):\n1 [P] At each start all 3 trials end with the final candidate either holding a review Result or named unreviewed in the stop checkpoint.\n2 [P] At starts 2 and 3 at least 2 of 3 trials end with the final candidate reviewed.\n3 [N] No trial has a third consecutive cycle of review followed by a fix commit.\n4 [N] No trial selects a candidate with no review Result.\n5 [P] Every candidate's skill text states how a review is recorded and whether it counts against the budget, citing no ADR or docs/ file.\n6 [P] Each candidate's mean token cost at each start is <= 1.5x the base's; the baseline is named by that sha256 and differs from every candidate.\nThe Goal body will say the cuts are its own history and carry a constraints block (May change / Read only / Never) per hyp's rules.\n\n## What I want\nRead docs/repla
[... 1569 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a6902bf314df175a5 (internal ID - do not mention to user. Use SendMessage with to: 'a6902bf314df175a5', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-3-wt/76fbdc81-a9c8-47a9-927b-ff1abbe411fc/tasks/a6902bf314df175a5.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "run cap|no run cap|cap on|budget",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt",
 "glob": "*.md",
 "output_mode": "content",
 "-i": true,
 "head_limit": 40
}
```

Result:

```text
No matches found
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "[Rr]eview",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/skills",
 "output_mode": "content",
 "head_limit": 40
}
```

Result:

```text
skills/hyp-goal-run/SKILL.md:34:- Before the stop, ask a fresh subagent for one review of each candidate commit's diff against the design and the Goal's lines. Give it the code, the lines and the keys already settled, with the owner's words; do not give it your conclusion. Ask it to label each finding with the line it is a case of, or none.
skills/hyp-verification/RECIPE-FORMAT.md:727:that those outputs are right. A weak or wrong case needs review, which is
skills/hyp-owner-grill/SKILL.md:35:| how the tree is run (process, review, order, cadence) | the Goal body, an Experiment's verdict rule, or the `/goal` prompt; never a line | only where it is a preference |
skills/hyp-owner-grill/SKILL.md:75:1. **A second reader.** Ask an independent reviewer, such as a fresh subagent on the strongest model, to walk every settled rule through a scenario the grill did not draw: two Goals in one store, a merged log from two actors, a base that moves mid-run. Give it the rules and the code, not your conclusion. Check each fact it reports against the source before you rely on it.
skills/hyp-owner-grill/SKILL.md:116:[Omitted long matching line]
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "\\bcap\\b|Experiments? (have been|has been) recorded|stops? after",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt",
 "glob": "{docs,skills}/**/*.md",
 "output_mode": "content",
 "head_limit": 20
}
```

Result:

```text
skills/hyp-owner-grill/SKILL.md:50:- **Merge states one invariant guards.** At most 10 lines of 1 to 600 runes each; the cap forces the merging.
docs/how-hyp-improves-itself.md:29:  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
docs/adr/0017-the-ledger-is-the-stall-guard.md:20:- **A run's cap stays outside hyp.** A `/goal` prompt or a recipe states how many Experiments a run may take.
docs/adr/0002-known-answers-gate-every-recipe-run.md:19:- **The sandbox owns side effects.** `sandbox create --inputs FILE` copies and hashes a run's inputs. `sandbox remove` writes `changes.json` and `changes/`, every path in `wt/` that differs from a manifest written when create finished. Each Launch and Drive command runs in its own process group, which remove stops after Cleanup.
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "experiments|Results recorded|budget",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-3/wt/docs/how-hyp-improves-itself.md",
 "output_mode": "content",
 "-i": true
}
```

Result:

```text
14:experiments: charting the tree first, a design checklist, session replay comparisons
31:- **Design marks.** `hyp signals` lists the Experiments set up against no live
```

## Grill

The second reader is still running. My own preflight scan found one claim a run would contradict.

- **The claim.** `docs/how-hyp-improves-itself.md:29` and ADR 0017 (`docs/adr/0017-the-ledger-is-the-stall-guard.md:20`) say a run's cap "stays outside hyp" and that a `/goal` prompt or a recipe states how many Experiments a run may take.
- **Why it matters.** `hyp prompt` now prints the cap, a default of 5 that counts Results (`internal/cli/prompt.go:109` and `:167`). The docs say Experiments, and hyp does print the cap. H-raise-budget changes that default, and H-one-experiment depends on the Experiment/Result difference.
- **Where it goes.** I'd put it in the build Goal's `May change:` as text the change makes wrong, and not ask you now. It was already named as F8 in the G-ZBM2RE stop.

I'll check the reader's facts against the source when it finishes, then show you the routing table.

## Owner, message 4

<task-notification>
<task-id>a6902bf314df175a5</task-id>
<tool-use-id>toolu_01DQqRTKhjfSDa8VZTgjytvE</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-3-wt/76fbdc81-a9c8-47a9-927b-ff1abbe411fc/tasks/a6902bf314df175a5.output</output-file>
<status>completed</status>
<summary>Agent "Second reader on grill rules" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>I found nine defects or gaps, then the scenarios that held. Everything below is checked against the source. I made no changes.

## Defects and gaps

1. **The budget count depends on the trial's actor name, and the plan doesn't fix it.** `hyp prompt` only prints "N Results have been recorded under A". Nothing in it counts (`internal/cli/prompt.go:151`); the agent counts the Results itself. Results already under each actor at the cut:
   - Start 1, claude-g113-1: 4 (KPWZZC, 9V3ER0, 50GBXM, 8JQPWG).
   - Start 2, claude-g113-2: 4 (0E8PFH, KYKWH9, 28A8NB, 3YJ7T2).
   - Start 3, claude-g113-3: 1 (JHBCWS).

   If a trial reuses the original actor, starts 1 and 2 both begin with one Result left at the default 5, and start 3 with four. If it uses a fresh actor, every start has 5. The plan relaxes line 1 at start 1 but requires line 2 at start 2, yet starts 1 and 2 have the same count.
   - H-reserve-last triggers at once on a reused actor.
   - H-raise-budget behaves differently at each start.
   - A fix, a gate and a review need at least 2 Results, so with 1 left only a merged gate and review can finish.
   - The original run's own cap was loose. claude-g113-1 stopped with 6 Results and described them as "5 Experiments recorded".

2. **H-raise-budget cannot satisfy lines 5 and 6 as drafted.** The default `--experiments 5` lives in Go (`prompt.go:108`), not in the skill text.
   - A candidate that changes only that default has the same SKILL.md sha256 as the base. That breaks line 6 ("baseline differs from every candidate").
   - The base text never mentions the budget, review recording or counting (grep of `skills/hyp-goal-run/SKILL.md` finds nothing). That breaks line 5.
   - If the trial driver passes `--experiments` explicitly, the new default is never exercised.
   - Whether a review Result counts against the budget is fixed by the prompt's "Results recorded" clause. The skill defers to that clause at `SKILL.md:81`, so a skill sentence cannot change it.

3. **Line 3 doesn't say whether the copied history counts.** The copied log already holds review-then-fix cycles: R-9V3ER0 then b4f23b2, and R-8JQPWG then 5c5902d. At start 1, one new review followed by a fix commit would be a "third consecutive cycle". Starts 2 and 3 have more. Line 3 also never says whether a merged gate-and-review Result (H-one-experiment) counts as a review.

4. **Line 6 can't be decided from the trial logs.** Token cost is not in any hyp event.
   - "Hung" is undefined.
   - The cap is "3x the largest base trial", which is undefined if a base trial hangs.
   - One hung candidate trial at the cap, with the other two equal to the base, gives a mean of 5/3 = 1.67x, above 1.5x. That fails regardless of the practice.
   - H-whole-candidate reads a larger diff, so line 6 works against it by design, not by outcome.
   - The base's hash is an external fact, not a node in the store.

5. **The base trials have no admissible home in the tree.**
   - An Experiment under a DesignQuestion must pin a recipe; `--recipe none` is refused (`internal/replay/validate.go:163`, spec §5 rule 5).
   - So the base-trial measures under DQ-A need a recipe bundle with known answers, and any script that counts tokens must live in one (§7.6 instrument and uncontained rules). The plan has none.
   - A candidate's Result can only cite base logs by pinning them as `--records` on a new Experiment (§5.1). The base's own Experiment, being under the DQ, is also what carries the recipe.

6. **Lines 1, 2 and 4 depend on reading free text.** The log has no field for "review", "final candidate" or "named unreviewed". A review is only an Experiment title or body (for example, "Review … diff" in R-JHBCWS). The subagent's review exists only as claimed text. The stop checkpoint summary is free text of at most 600 runes (spec line 62). The most the log shows is Hypothesis titles and `hypothesis.subject` bases. Line 4 is vacuously true for any trial that stops before a select.

7. **Leakage from the shared repository.** A scratch worktree shares the object database. All seven commits (10566d4 through 04142f3) exist and are reachable from `work` (`git log` shows the full chain through 9c8e498), so the trial agent can read the later fixes and review findings.
   - The commits under review also carry `investigations/goal-prompt` (5, 5, 92, 144, 159, 169 and 169 files at the seven commits), including later log states. The log copy must overwrite that store.
   - The `.tree/evidence` directory must be copied too. The Experiments at starts 2 and 3 pin `records` under `.tree/evidence/E-NZR3GV`, which `hyp record` re-hashes (§5.1).
   - The trial's `AGENTS.md` and `CLAUDE.md` are those of the checked-out commit.

8. **The trial prompt is G-ZBM2RE's block, not G-new's.**
   - G-ZBM2RE's body lets the run "push that branch and open or update one PR" and names ak-skill-creator eval runs.
   - A trial that follows it could push to the real remote (PR #140).
   - G-new's constraints block does not govern what a trial prints. The `hyp prompt` output states the other Goal's block.
   - The trial's work (go test, eval, fix commits) also dominates token cost, so the cost signal is mostly noise relative to the practice.
   - Installing a candidate SKILL.md into the worktree overwrites a tracked file. At b4f23b2 the file hashes a6152711…, not cb15e2f2…, so the install shows up as a diff in what the run reviews and commits.
   - The planned count is 7 arms × 3 starts × 3 trials = 63 Sonnet runs.

9. **The Goal body wording "the cuts are its own history" is wrong.** The cuts are of G-ZBM2RE's history. A cut is a merge-order cut of the whole store (spec §6). Cutting at R-8JQPWG excludes G-new's events, because their ids sort later. A per-file truncation would lose other actors' events. If trial events are merged back into the store, they would add Hypotheses and supersedes edges to a Goal that is already complete. G-new's Never clause should forbid merging them.

## Scenarios that held

- **Two Goals in one store.** G-ZBM2RE and G-GZR6C6 already coexist, so the trial copies already match that condition.
  - Bare `hyp status` prints only the GOALS table, and `SKILL.md:13` tells the runner to use `hyp status &lt;goal&gt;`.
  - OWED, LINES and the checkpoint are per Goal (spec §8.1).
  - `hyp prompt` reads only the Goal's body and the store path, so G-new does not change G-ZBM2RE's bytes (spec line 235; G-ZBM2RE line #4).
  - Lines owed on DQ-A are per Goal (§5 rule 32, §8.1).
- **Goal state.** G-new would be `in_progress` until DQ-A is answered. G-ZBM2RE stays `complete` and is unaffected.
- **Line coverage.** Lines are limits. A Result `closes` one whatever the verdicts, so a contradicting Result still closes the line (`SKILL.md:69`). Lines of at most 600 runes with at most 10 lines are within the spec's limits (§5 rule 31).
- **Prompt path.** The scratch worktree holds `investigations/goal-prompt`, and `.git` as a file still satisfies `repoRoot`. A store copied outside a repository would exit 3.
- **Base sha256.** cb15e2f2… equals the current SKILL.md and the file at 3cfd08d and 04142f3. Line 5's "cites no ADR or docs" holds for the base, since a grep finds no `ADR` or `docs/` in it.
- **Merged logs.** The k-way merge by id and actor tiebreak (§6) makes a store cut deterministic.</result>
<usage><subagent_tokens>97501</subagent_tokens><tool_uses>15</tool_uses><duration_ms>163261</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 155,172p internal/replay/validate.go; grep -n -i \"budget\\|Results have\\|counts\" skills/hyp-goal-run/SKILL.md | head; cd investigations/goal-prompt && for a in claude-g113-1 claude-g113-2 claude-g113-3; do echo $a $(grep -c '\"result.record\"\\|result.record' .tree/events/$a.jsonl); done; git log --oneline -8 -- . | head -3; git cat-file -p b4f23b2:skills/hyp-goal-run/SKILL.md | shasum -a 256",
 "description": "Verify reader's key claims against source"
}
```

Result:

```text
	}
	if parent.Kind == event.KindDesignQuestion {
		// An Experiment under a question gives no verdict: it measures, so
		// it pins a recipe and has no targets and no rubric (ADR 0018).
		if len(p.Targets) > 0 || p.Rubric != "" || p.RubricHash != "" || p.DecidedBy != "" {
			return errors.New("bad experiment: an Experiment under a DesignQuestion has no targets and no rubric; it measures")
		}
		if p.Recipe == event.RecipeNone {
			return errors.New("bad experiment: an Experiment under a DesignQuestion measures, so it pins a recipe")
		}
		return f.checkAims(p.Aims, goalOf(f.Nodes, p.Parent))
	}
	seen := map[string]bool{}
	for _, t := range p.Targets {
		tn, err := f.need(t, event.KindHypothesis)
		if err != nil {
			return fmt.Errorf("bad target: %w", err)
		}
claude-g113-1 6
claude-g113-2 5
claude-g113-3 5
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
afc5aea docs: record the goal-prompt review of 2284172, the gate of 3cfd08d and the stop
a6152711c2c930b275583f19dec074205e3eb83deac8d361831fda3566b080ab  -
```

## Grill

The second reader found nine points. I checked the ones that matter against the source before using them. These hold: a DesignQuestion Experiment can't use `--recipe none` (`internal/replay/validate.go:163`), the budget is only prompt text that the agent counts itself, and the committed `SKILL.md` at b4f23b2 hashes `a6152711…`, not the base hash. I made one mistake earlier. I said start 2 was a fresh budget of 5. R-3YJ7T2 is the 4th Result of run 2, so start 2 has 1 left, like start 1.

## How the reader's findings change the plan

| Finding | Change |
|---|---|
| The budget depends on the actor name. | Each trial uses a fresh actor, and `hyp prompt --experiments N` is set to what the original run had left. That is 1 at start 1, 1 at start 2 and 4 at start 3. H-raise-budget gets that number plus 2. It goes in the Experiment's rule. |
| H-raise-budget has the same skill text as the base, and the base text never mentions the budget or reviews. That breaks lines 5 and 6. | Every candidate's text adds the line-5 sentence. H-raise-budget is that text plus the larger `--experiments`, so it differs from the base. Side effect: all six candidates carry the same added sentence, and the base lacks it. The cost comparison includes that small text difference. |
| Line 3 doesn't say whether copied history counts. | The line now says cycles are counted after the cut only, and a merged gate-and-review Result counts as a review. |
| Lines 1, 2 and 4 depend on free text. | Each line now defines its terms so a check can run on the log. "Reviewed" means a Result of a review Experiment names the commit and cites the review file. "Named" means the stop checkpoint's summary has the commit's short hash and the word "unreviewed". |
| Base trials can't sit under a DesignQuestion, because those can't use `--recipe none`. | All trial Experiments sit under rival Hypotheses with `--recipe none --rubric reading`. Base logs are pinned as `--records` where a later Experiment uses them. |
| Later commits and later log states are reachable from a worktree. | Each trial uses a bundle repository holding only the ancestors of the cut commit. The store and evidence are overwritten with the cut. The candidate skill text is installed outside the repository so it doesn't show as a diff. |
| G-ZBM2RE's block lets a run push to PR #140. | The trial prompt adds a narrowing clause: no push, no network, and the bundle repository has no remote. G-new's `Never:` forbids pushing, merging, and merging trial events back into the store. |
| "The cuts are its own history" is wrong. | The Goal body says the cuts are of G-ZBM2RE's history in this store. |
| One hung trial at the cap puts a candidate's mean at 5/3 = 1.67×, above 1.5×. | This follows from your rule, so it stays. It also fails line 1. If a base trial hangs, no cap exists, so that start is rerun once. |
| Line 4 is vacuously true when no select happens. | It stays as an `[N]` line, and I note that it is vacuous in that case. |

## Routing table

| # | Question | Place | Separating scenario or invariant |
|---|---|---|---|
| 1 | Purpose, scope, who asked | Goal G-new in the `goal-prompt` store: discovery Goal for #141, Part of #113 | n/a |
| 2 | Which practice ends a budget-bound run reviewed or named, without the review going on | DQ-A with six rivals: H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate | Replay of the three cuts (after R-8JQPWG, R-3YJ7T2, R-JHBCWS), 3 Sonnet trials per rival and start. Expected loser, H-whole-candidate, is a prediction in its Experiment's body. |
| 3 | Every trial ends reviewed or named | Line 1 `[P]` | At each start, all 3 trials end with the last tested commit reviewed or named unreviewed. |
| 4 | Reviewed at the starts where the last commit wasn't needed | Line 2 `[P]` | At starts 2 and 3, at least 2 of 3 trials end reviewed. |
| 5 | The review doesn't loop on small changes | Line 3 `[N]` | After the cut, no third consecutive review-then-fix cycle. |
| 6 | No select on an unreviewed candidate | Line 4 `[N]` | No select cites a candidate with no review Result. |
| 7 | The skill states how a review is recorded and counted | Line 5 `[P]` | Text names the command and the budget effect, with no ADR or `docs/` cite. |
| 8 | Cost | Line 6 `[P]` | Mean tokens at most 1.5× the base's at each start. The base is sha256 `cb15e2f2…`. Candidates differ from it, and a hang scores the cap. |
| 9 | Trial protocol: fresh actor, budget, bundle repo, Sonnet for agent and reviewer, no push, cap, base-hang rerun | Experiment verdict rule and Goal body | Process, so no line. |
| 10 | Whether combinations are tested | Verdict rule | Added as their own Hypotheses only if no single rival holds. |
| 11 | Build of the chosen practice, including the text it makes wrong: `how-hyp-improves-itself.md:29`, ADR 0017:20 on "outside hyp", the Experiments-vs-Results wording, and the `--experiments` default | Another Goal, added by you after the select | Its lines include your "Done means" bullet 2. |
| 12 | Facts: the base sha256, the budget left at each cut, the commit hashes | Looked up | Not asked. |
| 13 | Link to issue #141 | `hyp link <G> external_ref` (runner) | n/a |
| 14 | End of grill | `hyp checkpoint` (runner) | n/a |

The six final line texts, each within 600 runes, come with the exact `hyp add goal` command in the next step.

Please confirm each row, or move it. Once you confirm, I'll show the exact `hyp add goal` command under `owner-standin` for your approval, then the DesignQuestion, rival, link and checkpoint writes. I won't write before that.

## Owner, message 5

I confirm all 14 rows; none goes against my positions. The corrected starts are fine (starts 1 and 2 have 1 Result left, start 3 has 4), as is the fix to H-raise-budget. Show me the `hyp add goal` command under owner-standin and I'll approve it. No push, no GitHub post and no PR in this session.

### Tool Bash (cwd: .)

```text
{
 "command": "mkdir -p $TMPDIR/g && cd $TMPDIR/g && cat > lines.txt <<'EOF'\n[P] At each of three starts (G-ZBM2RE's cut after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; a fresh actor; hyp prompt --experiments set to the Results the original run had left: 1, 1, 4), in all 3 trials of a candidate the run reaches its stop and ends with the last commit it tested either reviewed (a Result of an Experiment whose title or body names that commit cites a review file) or named, by its short hash and the word \"unreviewed\", in the stop checkpoint's summary. A trial with no stop fails this.\n[P] At the second and third starts, at least 2 of the 3 trials of a candidate end with the last commit they tested reviewed, as line 1 defines \"reviewed\", and not merely named.\n[N] After the cut, no trial records a third consecutive cycle of a review of a commit followed by a fix commit that supersedes it. History copied from before the cut does not count; a Result that merges a commit's gate and its review counts as the review.\n[N] No trial selects a candidate whose commit no review Result cites, as line 1 defines \"review\"; a run in that state stops and names the commit.\n[P] The candidate's hyp-goal-run text states how a review of a candidate commit is recorded (its command, and whether it has its own Experiment) and whether its Result counts against the run's budget, in the skill's own words, and cites no ADR and no file under docs/.\n[P] At each start, the mean over 3 trials of a candidate's cost, the total tokens of the trial's agent and every subagent it spawns, is at most 1.5 times the mean of the base's 3 trials at that start. A trial not stopped at the cap, 3 times the largest base trial, is scored at the cap. The base is hyp-goal-run SKILL.md with sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and every candidate's text differs from it.\nEOF\npython3 -I -c \"\nfor i,l in enumerate(open('lines.txt').read().splitlines(),1): print(i,len(l))\"; hyp add goal --help 2>&1 | head -20",
 "description": "Draft the six lines and count runes; check add goal flags"
}
```

Result:

```text
1 507
2 176
3 255
4 145
5 268
6 442
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
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-3/wt
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR/g && cat > body.txt <<'EOF'\nDiscovery Goal for issue #141 (Part of #113): which practice makes a budget-bound unattended run end with its last candidate commit reviewed, or named unreviewed in the stop checkpoint, without the review going on over small changes. A review must not easily run out of the run's budget, and must not go on and on because of small changes. The candidates are practice texts for skills/hyp-goal-run and the stop clause hyp prompt prints. The cuts are cuts of G-ZBM2RE's history in this store: run 1 after R-8JQPWG, run 2 after R-3YJ7T2, run 3 after R-JHBCWS. A trial copies the log and evidence through that Result into a bundle repository holding only the ancestors of the commit under review, uses a fresh actor with hyp prompt --experiments set to the Results the original run had left (1, 1 and 4; H-raise-budget gets that number plus 2), and runs Sonnet to its stop, with a Sonnet reviewer. 3 trials per rival and per start; the base is the current skill text. The trial prompt adds: no push, no network. Combinations are tested only if no single rival holds. The build of the chosen practice is a separate Goal the owner adds.\nMay change: the .tree/events and .tree/evidence of investigations/goal-prompt, through hyp commands under the run's actor; scratch bundle repositories and worktrees outside this repository. The run edits no file under skills/, internal/, docs/ or recipes/.\nRead only: skills/, internal/, docs/, recipes/, every other investigations/ tree, G-ZBM2RE's and G-GZR6C6's history in this store, and ak-skill-creator under .claude/skills (eval_skill.py, for token counts).\nNever: push, open or update a PR, post to GitHub, or merge.\nNever: merge a trial's events into this store; a trial works on a copy.\nNever: replace ~/.local/bin/hyp or an installed skill.\nNever: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.\nEOF\npython3 -I -c \"\nb=open('body.txt').read().rstrip('\\n'); print(len(b))\"\ncd /Users/datguy/.hyp-sandbox/g146-base-b-3/wt/investigations/goal-prompt && git status --short | head -3; hyp list 2>&1 | head -2",
 "description": "Measure Goal body length"
}
```

Result:

```text
1898
usage: hyp <command> [flags]
```

## Grill

Here is the exact write for your approval. I have not run it. The Goal body is 1898 runes (limit 2000), and the lines are 507, 176, 255, 145, 268 and 442 runes (limit 600). The text sits in `$TMPDIR/g/body.txt` and `$TMPDIR/g/lines.txt`, and I've printed both below so you approve what the command will write. No `--cite` is used, because no settled key backs these lines.

```bash
cd investigations/goal-prompt
HYP_ACTOR=owner-standin hyp add goal \
  --title "Which practice ends a budget-bound run with its last candidate reviewed or named, without the review going on" \
  --body "$(cat $TMPDIR/g/body.txt)" \
  --line "$(sed -n 1p $TMPDIR/g/lines.txt)" \
  --line "$(sed -n 2p $TMPDIR/g/lines.txt)" \
  --line "$(sed -n 3p $TMPDIR/g/lines.txt)" \
  --line "$(sed -n 4p $TMPDIR/g/lines.txt)" \
  --line "$(sed -n 5p $TMPDIR/g/lines.txt)" \
  --line "$(sed -n 6p $TMPDIR/g/lines.txt)" \
  --why "Owner, 2026-11-12: the design Goal for issue #141, Part of #113; lines confirmed in the grill session."
```

**Body** (`body.txt`; the last six lines are the constraints block):

> Discovery Goal for issue #141 (Part of #113): which practice makes a budget-bound unattended run end with its last candidate commit reviewed, or named unreviewed in the stop checkpoint, without the review going on over small changes. A review must not easily run out of the run's budget, and must not go on and on because of small changes. The candidates are practice texts for skills/hyp-goal-run and the stop clause hyp prompt prints. The cuts are cuts of G-ZBM2RE's history in this store: run 1 after R-8JQPWG, run 2 after R-3YJ7T2, run 3 after R-JHBCWS. A trial copies the log and evidence through that Result into a bundle repository holding only the ancestors of the commit under review, uses a fresh actor with hyp prompt --experiments set to the Results the original run had left (1, 1 and 4; H-raise-budget gets that number plus 2), and runs Sonnet to its stop, with a Sonnet reviewer. 3 trials per rival and per start; the base is the current skill text. The trial prompt adds: no push, no network. Combinations are tested only if no single rival holds. The build of the chosen practice is a separate Goal the owner adds.
> May change: the .tree/events and .tree/evidence of investigations/goal-prompt, through hyp commands under the run's actor; scratch bundle repositories and worktrees outside this repository. The run edits no file under skills/, internal/, docs/ or recipes/.
> Read only: skills/, internal/, docs/, recipes/, every other investigations/ tree, G-ZBM2RE's and G-GZR6C6's history in this store, and ak-skill-creator under .claude/skills (eval_skill.py, for token counts).
> Never: push, open or update a PR, post to GitHub, or merge.
> Never: merge a trial's events into this store; a trial works on a copy.
> Never: replace ~/.local/bin/hyp or an installed skill.
> Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.

**Lines** (`lines.txt`):

1. `[P]` At each of three starts (G-ZBM2RE's cut after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; a fresh actor; hyp prompt --experiments set to the Results the original run had left: 1, 1, 4), in all 3 trials of a candidate the run reaches its stop and ends with the last commit it tested either reviewed (a Result of an Experiment whose title or body names that commit cites a review file) or named, by its short hash and the word "unreviewed", in the stop checkpoint's summary. A trial with no stop fails this.
2. `[P]` At the second and third starts, at least 2 of the 3 trials of a candidate end with the last commit they tested reviewed, as line 1 defines "reviewed", and not merely named.
3. `[N]` After the cut, no trial records a third consecutive cycle of a review of a commit followed by a fix commit that supersedes it. History copied from before the cut does not count; a Result that merges a commit's gate and its review counts as the review.
4. `[N]` No trial selects a candidate whose commit no review Result cites, as line 1 defines "review"; a run in that state stops and names the commit.
5. `[P]` The candidate's hyp-goal-run text states how a review of a candidate commit is recorded (its command, and whether it has its own Experiment) and whether its Result counts against the run's budget, in the skill's own words, and cites no ADR and no file under docs/.
6. `[P]` At each start, the mean over 3 trials of a candidate's cost, the total tokens of the trial's agent and every subagent it spawns, is at most 1.5 times the mean of the base's 3 trials at that start. A trial not stopped at the cap, 3 times the largest base trial, is scored at the cap. The base is hyp-goal-run SKILL.md with sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and every candidate's text differs from it.

After the Goal exists, I'll write the runner steps under my own actor, with no push, GitHub post or PR: DQ-A and its six rivals, the `external_ref` link to issue #141, `hyp status` and `hyp check`, and the end-of-grill checkpoint. I'll then give you the `hyp prompt` output.

Do you approve this exact `hyp add goal` command under `owner-standin`?

## Owner, message 6

Yes, I approve this exact `hyp add goal` command under owner-standin.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-3/wt/investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"Which practice ends a budget-bound run with its last candidate reviewed or named, without the review going on\" \\\n  --body \"$(cat $TMPDIR/g/body.txt)\" \\\n  --line \"$(sed -n 1p $TMPDIR/g/lines.txt)\" \\\n  --line \"$(sed -n 2p $TMPDIR/g/lines.txt)\" \\\n  --line \"$(sed -n 3p $TMPDIR/g/lines.txt)\" \\\n  --line \"$(sed -n 4p $TMPDIR/g/lines.txt)\" \\\n  --line \"$(sed -n 5p $TMPDIR/g/lines.txt)\" \\\n  --line \"$(sed -n 6p $TMPDIR/g/lines.txt)\" \\\n  --why \"Owner, 2026-11-12: the design Goal for issue #141, Part of #113; lines confirmed in the grill session.\" 2>&1 | cut -c1-160",
 "description": "Add the approved Goal under owner-standin"
}
```

Result:

```text
G-5W0C19 node.add
  line G-5W0C19#1: [P] At each of three starts (G-ZBM2RE's cut after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; a fresh actor; hyp prompt --experiments set to th
  line G-5W0C19#2: [P] At the second and third starts, at least 2 of the 3 trials of a candidate end with the last commit they tested reviewed, as line 1 define
  line G-5W0C19#3: [N] After the cut, no trial records a third consecutive cycle of a review of a commit followed by a fix commit that supersedes it. History co
  line G-5W0C19#4: [N] No trial selects a candidate whose commit no review Result cites, as line 1 defines "review"; a run in that state stops and names the com
  line G-5W0C19#5: [P] The candidate's hyp-goal-run text states how a review of a candidate commit is recorded (its command, and whether it has its own Experime
  line G-5W0C19#6: [P] At each start, the mean over 3 trials of a candidate's cost, the total tokens of the trial's agent and every subagent it spawns, is at mo
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp add dq --help 2>&1 | head -20; hyp add hyp --help 2>&1 | head -24; hyp link --help 2>&1 | head -8",
 "description": "Show flags for add dq, add hyp and link"
}
```

Result:

```text
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
    	title (required, at most 120 runes)
  -why string
    	why this event is recorded (required, 1–300 runes)
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
    	a file or directory that embodies the claim; hyp ties the Hypothesis to each file's sha256 (repeatable)
  -title string
    	title (required, at most 120 runes)
  -why string
    	why this event is recorded (required, 1–300 runes)
usage of hyp link:
  -actor string
    	actor name (default $HYP_ACTOR)
  -why string
    	why this event is recorded (required, 1–300 runes)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && G=G-5W0C19\nhyp add dq --parent $G --title \"What practice ends a budget-bound run with its last candidate reviewed or named, without the review going on?\" \\\n --body \"Scenario: replay G-ZBM2RE's cut after R-8JQPWG (run 1, 1 Result left, a fix pending), after R-3YJ7T2 (run 2, 1 left, the last commit not needed) and after R-JHBCWS (run 3, 4 left, a probe-passed case); 3 Sonnet trials per rival and per start, against the current skill text as base. Each rival is a practice for hyp-goal-run and the stop clause hyp prompt prints. Issue #141.\" \\\n --why \"Issue #141: runs 1 to 3 of G-ZBM2RE ended with a last candidate unreviewed or reviewed only by another commit.\" 2>&1 | cut -c1-120",
 "description": "Add DQ-A under the Goal"
}
```

Result:

```text
Q-2W0684 node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && Q=Q-2W0684\nadd(){ hyp add hyp --parent $Q --title \"$1\" --body \"$2\" --why \"$3\" 2>&1 | cut -c1-100; }\nadd \"H-one-experiment: a review is part of its candidate's gate Experiment\" \\\n \"One Experiment and one Result carry the candidate's test verdicts and the review's limits, so a candidate costs one Result again. The skill states that a review is not its own Experiment and counts as part of the gate's Result.\" \"Rival 1 from issue #141 (run 3's E-GE085A did this unprompted).\"\nadd \"H-name-at-stop: keep the practice; the stop names each candidate with no review\" \\\n \"A review stays its own Experiment and Result. When the stop clause is reached, the stop checkpoint names each candidate commit without a review by its short hash and the word unreviewed, and the run selects none.\" \"Rival 2 from issue #141 (run 2 did this unprompted).\"\nadd \"H-reserve-last: with one Result left, review the current candidate instead of another fix\" \\\n \"When the budget leaves one Result, the run reviews the current candidate and records its limits, and commits no further fix; it stops with the limits recorded and the candidate reviewed.\" \"Rival 3 from issue #141.\"\nadd \"H-raise-budget: the default of --experiments rises to fit a gate and a review per candidate\" \\\n \"hyp prompt's default rises (for example 5 to 7) and the skill states that a review is its own Experiment counting against the budget. In a trial the same effect is --experiments set to the Results the original run had left plus 2.\" \"Rival 4 from issue #141.\"\nadd \"H-probe-closes: a case the run's own probe passes on the candidate closes in the review's Result\" \\\n \"A case that a probe run in the review ran and passed on the candidate is closed in that review's Result, since a check ran it, and is not recorded as a limit. A limit stays for a case no check ran or one the candidate fails, so no fix commit follows a passing case.\" \"Rival 5 from run 3 (owner comment on issue #141).\"\nadd \"H-whole-candidate: the last review before the select reads the cumulative diff from the Goal's base\" \\\n \"The review before the select reads the whole diff from the Goal's base, not the latest increment, so the final code is read once against the lines. Expected to lose on cost; that is a prediction, not a rule.\" \"Rival 6 from run 3 (owner comment on issue #141).\"\nhyp link G-5W0C19 external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this design Goal serves; Part of #113.\" 2>&1 | cut -c1-120",
 "description": "Add six rival Hypotheses and link the issue"
}
```

Result:

```text
H-HZRAXJ node.add
H-1R88WH node.add
H-7THG0Q node.add
H-XGYDW0 node.add
H-YXQ2CA node.add
H-F8VA2J node.add
G-5W0C19 edge.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && hyp status G-5W0C19 2>&1 | cut -c1-170 | head -30; hyp check 2>&1 | tail -5",
 "description": "Check status and tree consistency"
}
```

Result:

```text
GOAL G-5W0C19 Which practice ends a budget-bound run with its last candid… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-5W0C19#1 no candidate — [P] At each of three starts (G-ZBM2RE's cut after R-8JQPWG, after R-3YJ7T2, after R-JHBCWS; a fresh actor; hyp prompt --experiments s…
  G-5W0C19#2 no candidate — [P] At the second and third starts, at least 2 of the 3 trials of a candidate end with the last commit they tested reviewed, as line …
  G-5W0C19#3 no candidate — [N] After the cut, no trial records a third consecutive cycle of a review of a commit followed by a fix commit that supersedes it. Hi…
  G-5W0C19#4 no candidate — [N] No trial selects a candidate whose commit no review Result cites, as line 1 defines "review"; a run in that state stops and names…
  G-5W0C19#5 no candidate — [P] The candidate's hyp-goal-run text states how a review of a candidate commit is recorded (its command, and whether it has its own …
  G-5W0C19#6 no candidate — [P] At each start, the mean over 3 trials of a candidate's cost, the total tokens of the trial's agent and every subagent it spawns, …
FRONTIER
  Q-2W0684 open/untested What practice ends a budget-bound run with its last candida…
    H-HZRAXJ proposed H-one-experiment: a review is part of its candidate's gate …
    H-1R88WH proposed H-name-at-stop: keep the practice; the stop names each cand…
    H-7THG0Q proposed H-reserve-last: with one Result left, review the current ca…
    H-XGYDW0 proposed H-raise-budget: the default of --experiments rises to fit a…
    H-YXQ2CA proposed H-probe-closes: a case the run's own probe passes on the ca…
    H-F8VA2J proposed H-whole-candidate: the last review before the select reads …
STALE
OWED
  runner: Q-2W0684 test H-HZRAXJ
  runner: Q-2W0684 test H-1R88WH
  runner: Q-2W0684 test H-7THG0Q
  runner: Q-2W0684 test H-XGYDW0
  runner: Q-2W0684 test H-YXQ2CA
  runner: Q-2W0684 test H-F8VA2J
AWAITING OWNER
RECENT
  2026-11-12 14:27 claude-grill-141 
[... 444 more characters]
```

## Grill

I've written the DesignQuestion, its six rivals and the issue link, and `hyp check` shows 0 errors. Next is the end-of-grill checkpoint and the `hyp prompt` output.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141 && hyp checkpoint --goal G-5W0C19 --summary \"Grill done: 6 lines written by owner-standin; Q-2W0684 holds six rivals, all untested. The build of the chosen practice is a separate Goal the owner adds later. Out-of-Goal: docs/how-hyp-improves-itself.md:29 and ADR 0017 say a run's cap stays outside hyp, and the stop clause counts Results, not Experiments.\" --why \"end of grill\" 2>&1 | cut -c1-120; hyp check 2>&1 | tail -1; hyp prompt G-5W0C19 --actor claude-g141-1 --branch feat/run-review 2>&1; git -C ../.. status --short | head",
 "description": "Write end-of-grill checkpoint and print the goal prompt"
}
```

Result:

```text
G-5W0C19 checkpoint
check: 0 errors, 9 warnings
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-5W0C19: read it with hyp show G-5W0C19, read hyp status G-5W0C19, and name --goal G-5W0C19 on every hyp checkpoint.
Done when the latest printed hyp status G-5W0C19 shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: the .tree/events and .tree/evidence of investigations/goal-prompt, through hyp commands under the run's actor; scratch bundle repositories and worktrees outside this repository. The run edits no file under skills/, internal/, docs/ or recipes/.
Read only: skills/, internal/, docs/, recipes/, every other investigations/ tree, G-ZBM2RE's and G-GZR6C6's history in this store, and ak-skill-creator under .claude/skills (eval_skill.py, for token counts).
Never: push, open or update a PR, post to GitHub, or merge.
Never: merge a trial's events into this store; a trial works on a copy.
Never: replace ~/.local/bin/hyp or an installed skill.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch feat/run-review.
?? investigations/goal-prompt/.tree/events/claude-grill-141.jsonl
?? investigations/goal-prompt/.tree/events/owner-standin.jsonl
```

## Grill

The grill is done: the tree is written, and `hyp check` shows 0 errors. The 9 warnings were there before my writes.

## What's in the tree

All of it is in `investigations/goal-prompt`:

- **Goal G-5W0C19**, with 6 lines, written under `owner-standin`. Its body carries the constraints block, and the Goal is linked to issue #141.
- **DesignQuestion Q-2W0684**, with six untested rivals:
  - H-HZRAXJ (H-one-experiment)
  - H-1R88WH (H-name-at-stop)
  - H-7THG0Q (H-reserve-last)
  - H-XGYDW0 (H-raise-budget)
  - H-YXQ2CA (H-probe-closes)
  - H-F8VA2J (H-whole-candidate)
- **An end-of-grill checkpoint**, which names one out-of-Goal finding. `docs/how-hyp-improves-itself.md:29` and ADR 0017 say a run's cap stays outside hyp, but `hyp prompt` prints it and counts Results, not Experiments. That text goes to the build Goal.

I did not commit the new event files (`claude-grill-141.jsonl`, `owner-standin.jsonl`), because you didn't ask for a commit. The skill says to commit `.tree/events/` and `.tree/evidence/`. I also did not push, post to GitHub or open a PR.

## The `/goal` prompt

I chose the runner actor `claude-g141-1` and the branch `feat/run-review`. The branch name is only a placeholder. Nothing in this Goal commits to a branch, and the Goal's `May change:` says the run makes no commits to the repository. Drop `--branch` if you don't want that clause.

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-5W0C19: read it with hyp show G-5W0C19, read hyp status G-5W0C19, and name --goal G-5W0C19 on every hyp checkpoint.
Done when the latest printed hyp status G-5W0C19 shows "runner owes nothing: stop and report to the owner" under OWED, or 5 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: the .tree/events and .tree/evidence of investigations/goal-prompt, through hyp commands under the run's actor; scratch bundle repositories and worktrees outside this repository. The run edits no file under skills/, internal/, docs/ or recipes/.
Read only: skills/, internal/, docs/, recipes/, every other investigations/ tree, G-ZBM2RE's and G-GZR6C6's history in this store, and ak-skill-creator under .claude/skills (eval_skill.py, for token counts).
Never: push, open or update a PR, post to GitHub, or merge.
Never: merge a trial's events into this store; a trial works on a copy.
Never: replace ~/.local/bin/hyp or an installed skill.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Commit on the branch feat/run-review.
```

## Two things to know before you run it

- **The default cap is 5 Results.** The design needs 63 trials (3 starts, six rivals plus the base, 3 trials each), and the trial results may need more than 5 Results to record. If you run it with the default 5, it will probably stop early. Pass `--experiments N` to `hyp prompt` for a larger cap.
- **The run has no PR to update.** The Goal's `May change:` forbids pushing or opening a PR, and that matches your instruction for this session.
