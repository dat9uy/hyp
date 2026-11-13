# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-b-1/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

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
 "command": "hyp status 2>&1 | head -50",
 "description": "Show hyp store status"
}
```

Result:

```text
no .tree store found; run 'hyp init'
```

## Grill

There's no store at the repo root, so I'll check where the trees live and read the practice text the issue cites.

### Tool Bash (cwd: .)

```text
{
 "command": "ls; ls investigations; ls -d investigations/*/.tree; git branch -a | head; which hyp; hyp version 2>&1 | head -2",
 "description": "Locate stores and hyp binary"
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
investigations/checkpoint-store-design/.tree
investigations/checkpoint-store/.tree
investigations/goal-prompt/.tree
investigations/id-citation/.tree
investigations/recipe-base/.tree
* work
/Users/datguy/.hyp-sandbox/g146-base-b-1/bin/hyp
hyp 0.1.0 (events v1)
```

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/wt/skills/hyp-goal-run/SKILL.md"
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

### Tool Grep (cwd: .)

```text
{
 "pattern": "experiments|Results recorded",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/wt/internal",
 "glob": "*prompt*.go",
 "output_mode": "content",
 "head_limit": 30
}
```

Result:

```text
internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
internal/cli/prompt.go:109:	experiments := fs.Int("experiments", 5, "stop after this many Results recorded under the actor")
internal/cli/prompt.go:123:	if *experiments < 1 {
internal/cli/prompt.go:124:		return usageErr("--experiments must be at least 1")
internal/cli/prompt.go:167:	fmt.Fprintf(&b, "Done when the latest printed hyp status %s shows \"runner owes nothing: stop and report to the owner\" under OWED, or %d Results have been recorded under %s, or you cannot take the next step honestly and say why.\n", g, *experiments, *actor)
internal/cli/prompt_test.go:63:	want(t, hyp(t, "prompt", g, "--actor", "claude-run-1", "--experiments", "3"), 0, "or 3 Results have been recorded under claude-run-1,")
internal/cli/prompt_test.go:331:	want(t, hyp(t, "prompt", g, "--actor", "r", "--experiments", "0"), 2, "--experiments must be at least 1")
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && export HYP_ACTOR=grill-141 && hyp status 2>&1 | head -60; ls ../../recipes; grep -n -i \"review\\|budget\" ../../docs/glossary.md | head -20",
 "description": "Show goal-prompt tree status and glossary mentions"
}
```

Result:

```text
GOALS (hyp status G-… prints one Goal)
  G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and w… — complete (answered 2, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-11 16:47 claude-grill-113
  G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0); runner owes nothing; checkpoint 2026-11-12 03:28 claude-g113-4
AWAITING OWNER
authoring-stops@2
id-citation-arms@1
id-citation-arms@2
recipe-authoring-arms@4
recipe-base-arms@1
3:The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.
46:A Goal whose candidate is a commit: building is a Goal like any other. A line asserts one property of the candidate, decided by its own tests or one pinned instrument: `[P]` pins what a valid input produces, and `[N]` pins that a forbidden state is refused. Each commit is a Hypothesis whose subject is the base plus that commit's change, and a fix is a new Hypothesis that supersedes it. Each Experiment's verdict rule asks for red on base and green on the candidate; the base of the red run's subject is where the build starts, and the Goal records no start commit. A finding of a run's review that is a case of a line is a limit; one that is not is an out-of-Goal finding. A change to a skill is a build Goal too (ADR 0021).
145:_Avoid_: approve, review; verdict (alone: a confirm's answer was a confirm verdict)
155:Retired by ADR 0015. It was the role that confirmed every select, 
[... 1321 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "grep -l \"R-8JQPWG\" -r .tree/events | head -3; ls .tree; hyp --help 2>&1 | head -40; ls ../../recipes/*; grep -rn \"replay\\|--at\\|cut\" ../../docs/replay-spec.md | grep -i \"cut\" | head -8",
 "description": "Check cut availability and hyp commands"
}
```

Result:

```text
.tree/events/claude-g113-1.jsonl
.tree/events/claude-g113-2.jsonl
events
evidence
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
  render --markdown [--as-of REF|TIME]        the w
[... 3768 more characters]
```

## Grill

The ledger I need is in place: `investigations/goal-prompt` already holds R-8JQPWG and run 3's events, so `hyp sandbox create --store … --cut <event>` can build both replay scenarios. That is a fact I looked up, so I'm not asking about it. Nothing is written yet. I need your answers to the five questions below, which only you can settle.

## Round 1

**Q1. Store and actor.**
- Today the goal-prompt store holds two complete Goals, G-GZR6C6 and G-ZBM2RE. #141 is "Part of #113", which is that store's issue.
- The problem is where the new Goal lives.
  - (a) A third Goal in the same store, for example `hyp add goal` in `investigations/goal-prompt`. Its Results can cite the run-1 to run-3 Results directly.
  - (b) A new tree under `investigations/run-review/`. The old Results would then be cited as `goal-prompt:<id>`, and the new tree starts clean.
- My recommendation is (a). One store holds many Goals, and the scenarios are cuts of this very store.
- Please also give me your actor name, for example `owner-dat`. My runner actor for this session is `grill-141`.

**Q2. One Goal or two.**
- Today #141 mixes a decision (which rival) with its effect ("after the change, a run that reaches its budget…").
- The decision is tested by replaying cuts, so its candidates are rule texts and the Goal is a discovery Goal. The change is a commit to `skills/hyp-goal-run/SKILL.md` and `internal/cli/prompt.go`, so it is a build Goal. Their candidates differ, so by the grid rule they are separate Goals.
- (a) Two Goals. This one, "Decide the review and stop practice of a build Goal's run", ends at the select. You add the build Goal afterwards.
  - Example: this Goal's run replays the cut after R-8JQPWG under rival rule texts and records which one ends with every candidate reviewed or named.
  - The build Goal then carries lines like "`hyp prompt` prints the clause X", and its tests check them.
- (b) One Goal that decides and builds.
  - A line about the printed clause could not be filled by a replay-only candidate, so some cells in the grid would never be filled.
- My recommendation is (a). #141's "Done means" first bullet is this Goal. The second bullet is the build Goal's.

**Q3. Lines.** These are the proposed invariants, in this Goal's terms (a check can run each on every candidate rule). Each [N] line comes from something the grill or the issue rejected.
1. [N] On a replay of the G-ZBM2RE cut after R-8JQPWG (one Experiment left, a fix commit pending), a run that stops with a candidate commit that has no review Result does not name that candidate as unreviewed in its stop checkpoint.
2. [P] The rule states how a review is recorded (Experiment kind, Result, verdict) and whether it counts against the stop clause. This is your 03:19 comment: "whichever rival, `hyp-goal-run` should state how a review is recorded…".
3. [N] A case that no check ran, or that the candidate fails, is never closed as passing. This guards H-probe-closes against closing on a claim.
4. [N] A finding of a review is never dropped. Each is a limit, a close by a check, a fix commit, a reading, or a named out-of-Goal finding.
5. [P] On run 3's cut after R-JHBCWS, a review that finds only a case the candidate already passes under a probe leads to no further commit.
- Line 5 is the one I'm least sure of. It fits H-probe-closes and H-reserve-last but might favour them. Say if it should instead be a plain rival property and not a line.
- Do you want these lines, fewer, or others?
- The grill adds no stored field. "Unreviewed" is derivable (no review Result cites the candidate), but "review" has no event type today, so only the text can name it.

**Q4. What a run may change.**
- I recommend this block.
  - `May change:` this tree's events and evidence, and scratch sandboxes.
  - `Read only:` `skills/`, `internal/`, `docs/` and the other trees.
  - `Never:` edit `skills/` or `docs/`, replace the installed binary, push or merge.
- The rule texts that win belong to the build Goal.
- How many Experiments should the stop clause allow? Four questions with two or three rivals each plus two scenarios need more than the default 5. I recommend `--experiments 12`.

**Q5. Which rivals the run tests.**
- I would put all of them under the tree and let the run test each:
  - H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate
  - the status quo as a baseline rival (separate Experiment, counted)
  - the combined rivals, as the tree's call
- Say if any rival is out.
- I'd split them into four DesignQuestions so combinations are one pick per question, and so each rival answers its own question:
  1. How a review is recorded and counted.
  2. How a budget-bound run ends with its final candidate accounted for.
  3. How a case the run's own probe passes is closed.
  4. What the review reads.
- You confirm that structure in the routing table before anything is written.

After your answers I'll compute round 2. It will cover the second-reader pass and the preflight scan, then the routing table.

## Owner, message 2

1 a, same store as a third Goal. My actor name is owner-standin.

2 a, agree to two Goals.

3 I can't judge the lines until I see how each one serves my two aims: a review must not easily run out of the run's budget, and a review must not go on and on because of small changes. Please say which aim each line serves. Until then I won't add or cut lines, and your call on line 5 stands.

4 Agree with the May change / Read only / Never block. Agree with `--experiments 12`.

5 I want every candidate fix under one question, tested against each other, not split into four. Keep the status quo as the baseline. H-whole-candidate stays in, but I expect it to lose because one cumulative review costs a lot, so treat it as the expected loser. The combined rivals are fine as the tree's call.

## Grill

Here is each line against your two aims. Aim A is that a review must not easily run out of the run's budget. Aim B is that a review must not go on and on because of small changes.

| Line | Aim A (budget) | Aim B (no endless review) | What it does |
|---|---|---|---|
| 1 [N] stop with an unreviewed final candidate, not named | Fallback only | No | It does not stop the budget running out. It makes the gap visible when it does. |
| 2 [P] the rule states how a review is recorded and whether it counts | Yes, directly | Weakly | Run 1 spent 2 Experiments per candidate because nothing said whether a review counts. |
| 3 [N] a case no check ran, or one the candidate fails, never closes as passing | No | Guard | It stops a rule from winning on B by closing cases it never checked, which is what H-probe-closes could do. |
| 4 [N] a finding is never dropped | No | Guard | It stops a rule from winning on B by ignoring findings. |
| 5 [P] run 3's cut: a review that finds only a case already passed by a probe leads to no commit | Yes, indirectly | Yes, directly | Without that commit the review chain ends on a reviewed candidate. |

Two gaps follow from the table:
- **Aim A has no line that asks for a reviewed final candidate.** Line 1 only asks that the gap be named. A rule that always names the gap and never reviews would meet line 1, though it fails your aim. Line 5 reaches A only on run 3's cut.
- **Aim B has one line that asks for the result (5), and two guards (3, 4) that stop it being met by cheating.**

I suggest one new line for A:
- **6. [P]** On the G-ZBM2RE cut after R-8JQPWG, the run's stop leaves its final candidate with a review Result.

Line 1 stays as the second best. Line 6 means H-name-at-stop alone is contradicted on the first scenario and survives only combined with a rival that reviews. That matches the issue, which says H-name-at-stop should combine with another rival.

You said you won't add or cut lines until you saw this, so line 6 is your call. If you decline it, lines 1 to 5 stand as before, with line 5 as I had it.

**Q5, restructured.** I'll make it one DesignQuestion with all the rivals under it, tested against each other. I have worded it around your two aims:
- *How does a build Goal's run keep every candidate reviewed within its budget, without reviews that continue because of small changes?*

The rivals are the six already listed plus the status quo as baseline. H-whole-candidate goes in as the expected loser, so a run still tests it. A combination is a rival if the tree adds it.

**One caution.** One question means one select. The select can pick a combination as a single Hypothesis, but it cannot pick one rule for "how a review is recorded" and another for "what the review reads" separately. The tree would then need a combination rival for each pair that should hold together. If you want finer picks later, the owner adds a new Goal for each.

**What's left.** I will run the second reader and the preflight scan against these rules once you answer line 6. After that you get the routing table to confirm, and nothing is written before you do.

## Owner, message 3

Line 6: agree, add it. It serves aim A directly, and I'm fine with line 1 staying as the fallback.

On the one-question caution: keep one question. If the select needs combination rivals for pairs that should hold together, let the tree add them.

Before the routing table, I have two questions:
1. How is cost computed for a candidate against the base at the same starting point? My limit is 1.5 times the base's cost.
2. Could the ak-context-engineering skill help with how we evaluate these rivals?

### Tool Grep (cwd: .)

```text
{
 "pattern": "\\bcost\\b|1\\.5",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/wt",
 "glob": "*.{md,go}",
 "output_mode": "content",
 "head_limit": 25
}
```

Result:

```text
skills/hyp-prototype/SKILL.md:83:- A measure that settles no Hypothesis, such as a cost the owner weighs
recipes/recipe-authoring-arms@4/recipe.md:54:cost, denials). It exits 1 when no transcript was written.
docs/how-hyp-improves-itself.md:19:So the loop is guarded during the run, by mechanisms that cost nothing when
docs/adr/0017-the-ledger-is-the-stall-guard.md:11:Run 3 of the multi-tenant tree showed the cost. `hyp signals` printed five `wasted` rounds, three `no-movement` pairs and one `spin`. Four of the five wasted rounds closed limits or applied an answer: E-FWTH64, E-YS13MW, E-BGHDGF and E-EKCW21. The one round that settled nothing, E-TT1HSQ, the ledger already showed while it mattered: `hyp status` printed `limit R-XFSQEE#1 [aimed_not_settled by E-TT1HSQ]`, naming the item and the Experiment.
docs/adr/0018-an-experiment-is-a-pre-registered-result.md:7:[Omitted long matching line]
docs/adr/0015-the-select-is-the-gate.md:83:[Omitted long matching line]
docs/adr/0016-a-limit-is-an-untested-case.md:25:- A limit that is not a case, such as a scope statement, now holds the select until it is referred. This is the cost of one meaning: the runner says once why no test closes it, and the owner accepts it.
docs/adr/0013-the-tree-machine-owns-its-words-and-its-rules.md:46:Owner decision on 2026-11-09: only a Result that names the key in `applies` applies an answer. The fallback is removed, because which Result read an answer is something hyp cannot see unless the Result says so, and a guess that is wrong once settles an owner answer silently. The cost: the multi-tenant tree owes `reading` again for six answers that were applied without being named. Clearing them takes a Result under an Experiment that aims at each key and applies it.
recipes/recipe-base-arms@1/recipe.md:73:writes `drive.json` (exit, turns, cost, denials). It exits 1 when no
recipes/recipe-base-arms@1/recipe.md:114:  1.5] it prints `extend: yes` and every label inconclusive. It exits 2
recipes/recipe-
[... 2390 more characters]
```

### Tool Glob (cwd: .)

```text
{
 "pattern": "**/ak-context-engineering*",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/wt"
}
```

Result:

```text
No files found
```

## Grill

**1. Cost against the base.** I can't compute it yet, because nothing we've settled defines it, and I won't make up a definition for a limit you care about.

- **What I found.**
  - The repository has a run cost only inside recipe summaries: `drive.json` carries `exit, turns, cost` per run (`recipes/recipe-base-arms@1/recipe.md:73`).
  - `skills/hyp-prototype/SKILL.md:83` treats a cost as a measure that settles no Hypothesis, one the owner weighs.
  - Your 1.5 times figure is not in the tree. `recipe-base-arms@1/recipe.md:114` has a `1.5` in a verdict rule, but I haven't read it, so I can't say it means the same thing.
- **Choices I need from you.**
  - **What counts as cost.** I recommend Results recorded under the run's actor from the cut to the stop. That is the unit of the stop clause, so it is the unit of aim A. It is also what H-whole-candidate would cost more of. The `drive.json` token and dollar cost would be reported next to it but would not decide anything.
  - **What the base is.** I recommend the status-quo rival, replayed on the same cut with the same starting events. Each candidate rival is then measured against that baseline run.
  - **Whether 1.5 times is a line or only a measure.**
    - As a line it would be `[N]` and would be line 7: a rule whose replay on the cut costs more than 1.5 times the status-quo baseline's is not admitted. By the line guide, a line that compares with a baseline names it by sha256, so it would carry the sha256 of the status-quo rule text, and it asserts that the candidate differs from that baseline.
    - As a measure it would go into each Experiment's verdict rule and would decide nothing alone.
  - I lean toward the line, since it serves aim A and is how H-whole-candidate loses on a stated rule rather than on your expectation. It is your call.
- **One issue to know about.** A replayed run is an LLM run, so its cost varies from one replay to the next. With one replay per rival, a 1.5 times limit sits within that noise. The rule needs to say how many replays set the baseline, or accept that it is only a coarse bound. Please tell me which you want.

**2. The ak-context-engineering skill.** I can't use it. The skills available to me in this session are the ones listed to me, and `ak-context-engineering` isn't among them (there is `ak-skill-creator`, which is for writing skills, not for evaluating rivals). A search of the repository also finds no such file. I won't guess what it says. If it is installed somewhere on your machine, give me its path or paste its text and I'll say what it adds. If it applies, it would be to how a replay run's context is built, which is a question for the Experiment's verdict rule and not for a line.

Nothing has been written. Once you answer the three cost choices and the replay-count question, I'll run the second reader and the preflight scan, then show you the routing table.

## Owner, message 4

Cost answers:

- **Unit:** I accept your recommendation. Cost is the Results recorded under the run's actor from the cut to the stop. The `drive.json` token and dollar cost is reported alongside and decides nothing.
- **Base:** I accept your recommendation. The base is the status-quo rule replayed on the same cut.
- **1.5 times:** I accept your lean. It becomes line 7, an `[N]` line.
- **Replay count:** 3 trials for each rival at each starting point, and 3 for the status-quo baseline too. The replays run on Sonnet. A trial that hangs counts as the most expensive. Don't replay a whole build from its start for any trial.
- **Skill:** drop the ak-context-engineering question. I have no path for it.

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader scenario walk",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer. Read-only: change no file, run no write command. Repo: /Users/datguy/.hyp-sandbox/g146-base-b-1/wt (Go program `hyp`; specs in docs/replay-spec.md, docs/sandbox-spec.md, docs/glossary.md; /goal prompt in internal/cli/prompt.go; skill text in skills/hyp-goal-run/SKILL.md; store with real runs in investigations/goal-prompt/.tree). Issue text: /Users/datguy/.hyp-sandbox/g146-base-b-1/tmp/issue-141.md.\n\nRules being designed (a new Goal G-NEW, a discovery Goal, third Goal in the store investigations/goal-prompt that already holds G-GZR6C6 and G-ZBM2RE):\n- ONE DesignQuestion with rival Hypotheses (rule texts for how a build Goal's run records/counts reviews and ends): status quo (baseline), one-experiment, name-at-stop, reserve-last, raise-budget, probe-closes, whole-candidate, plus combinations the tree may add.\n- Rivals are separated by replaying cuts of the store (`hyp sandbox create --store SRC --cut EVENT`): the cut after R-8JQPWG (one Experiment left, a fix commit pending) and run 3's cut after R-JHBCWS. 3 trials per rival per cut, 3 for the status-quo baseline, on Sonnet; a hung trial counts as most expensive; no trial replays a build from its start.\n- Cost of a candidate = number of Results recorded under the run's actor from the cut to the stop. Line 7: a rule whose replay on the cut costs more than 1.5x the status-quo baseline's is not admitted (line names the baseline text by sha256).\n- Lines (at most 10, each 1-600 runes): 1[N] stop with an unreviewed final candidate not named in the stop checkpoint; 2[P] the rule states how a review is recorded and whether it counts against the stop clause; 3[N] a case no check ran or the candidate fails never closes as passing; 4[N] a review finding is never dropped; 5[P] on run 3's cut a review finding only cases a probe already passes leads to no further commit; 6[P] on the cut after R-8JQPWG the stop leaves the final candidate with a review Result; 7[N] the 1.5x cost line.\n\nTask: walk each rule through scenarios the designers did not draw and report what breaks or is undefined, each with file:line evidence from the code/specs. Scenarios to cover at least: two or three Goals in one store (does the stop clause count, `hyp status`, OWED, or `limit` closing cross Goals? does the new Goal's runner Result count toward other Goals' budgets?); a merged log from two actors; the owner-standin actor and runner actor writing at once; a base commit that moves mid-run; whether a cut store can be replayed with a fresh runner and a different actor whose Result count starts at 0; what \"Results recorded under the actor from the cut to the stop\" means when the prompt's clause counts all Results under the actor; whether a line can name a baseline by sha256 and compare cost (check the line rules in the docs/code); whether replaying a cut that contains the Goal's own 
[... 191 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a4f4f7bb89e1dcbdf (internal ID - do not mention to user. Use SendMessage with to: 'a4f4f7bb89e1dcbdf', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-1-wt/6a7d4c53-fb15-4e5d-9692-5a1a9833b84c/tasks/a4f4f7bb89e1dcbdf.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

### Tool Grep (cwd: .)

```text
{
 "pattern": "sha256|baseline",
 "path": "/Users/datguy/.hyp-sandbox/g146-base-b-1/wt/docs/glossary.md",
 "output_mode": "content",
 "head_limit": 10
}
```

Result:

```text
57:[Omitted long matching line]
402:A throwaway change over the subject's base that never lands in code, given to `hyp sandbox create` as `--patch ARM=FILE` for one Arm. An Arm has at most one; a patch is never versioned, stacked or rebased, and a change that lands is a commit used as the base instead. hyp checks that it applies to the base, records its sha256 and keeps it with the kept files. It is the subject, not a file of the recipe bundle ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)).
408:[Omitted long matching line]
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 100,200p internal/cli/prompt.go; grep -n \"sandbox\" cmd/hyp/*.go | head -5; git log --oneline -1; git status --short | head",
 "description": "Read prompt stop clause code"
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
[... 1250 more characters]
```

## Grill

The second reader is running, and I'm waiting on its report. I've checked two facts myself so far.

- `hyp prompt` counts "Results recorded under the actor", with a default of 5 (`internal/cli/prompt.go:109` and `:167`).
- A replayed trial that uses a fresh actor on a cut therefore starts its count at 0. That is the same as your cost unit, "Results from the cut to the stop".

Once the reader reports, I'll check each fact it gives against the source. Then I'll show you the routing table.

## Owner, message 5

<task-notification>
<task-id>a4f4f7bb89e1dcbdf</task-id>
<tool-use-id>toolu_01FexNLofRxuTdwmz9VsBVin</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-1-wt/6a7d4c53-fb15-4e5d-9692-5a1a9833b84c/tasks/a4f4f7bb89e1dcbdf.output</output-file>
<status>completed</status>
<summary>Agent "Second-reader scenario walk" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>Findings, each checked in the source unless marked unverified.

**Line 7: the sha256 baseline and the 1.5x test**
- A line is free text of 1–600 runes plus up to 10 `cites`. Each cite must be a settled key: a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis (docs/replay-spec.md:86, rule 31 at :133). A sha256 cannot be a cite.
- A sha256 inside the text is just characters. I found no code that resolves a hash in a line to a rule text, and none that compares costs.
- `hyp status` prints no cost, and nothing derives "Results recorded under actor X between event A and event B". The 1.5x test is a manual count, not a gate.
- Line 7 is a judgement by the reader. The same applies to line 2's "counts against the stop clause".

**Counting is store-wide, not per Goal**
- The stop clause counts "%d Results have been recorded under %s" (internal/cli/prompt.go:167). The count is by actor, and the clause is only text for the runner. Nothing in `hyp prompt` or `status` scopes it to a Goal.
- Run 2's clause quotes the same 5 (issue-141.md).
- If a runner shares an actor across G-NEW and G-ZBM2RE, every Result for either Goal spends one budget. Replay-cost Results and the real run's would mix under that actor. This is the case when the cut holds claude-g113-N and the replay reuses its name.
- The store has per-actor files `claude-g113-1` to `-4`, `claude-grill-113` and `owner-dat` (investigations/goal-prompt/.tree/events). So "the actor" is ambiguous unless each trial gets its own.

**Cuts and a fresh actor**
- `CutAt` writes one file per actor with kept events and merges by id, ties broken by file name (internal/store/cut.go:35-87, internal/replay/replay.go:13-37). A fresh `--actor` has no file, so its count starts at 0 by construction. `ValidActor` is the only check on the actor (prompt.go:120).
- A cut at R-8JQPWG keeps the old actor's earlier Results. If the replay reuses that actor, the model sees a spent budget. The issue's "cost = Results from the cut to the stop" matches only a fresh actor, or a subtraction of the cut's count.
- Whether a fresh actor can drive a Goal another actor started: unverified.

**Cut contents and the new Goal**
- Lines are written only with the Goal's `node.add`, and a new line is a new Goal (replay-spec.md:86, rule 24 at :125). G-NEW's lines therefore cannot be added to a cut after the fact.
- A cut at R-8JQPWG or run 3's R-JHBCWS predates G-NEW, so G-NEW is not in it. A trial runs under G-ZBM2RE's prompt and its lines.
- Lines 5 and 6 are about the cut's behaviour, not a state hyp can check. Line 6 asks for a review Result on the final candidate; nothing derives that.
- Coverage is per Hypothesis and per live top-level question (spec :345). For a discovery Goal, "every candidate" means the rule texts, but the replay evidence belongs to a different store. A line closer must be a Result under G-NEW's tree.

**Two or three Goals in one store**
- Lines, limits and `closes` are keyed by Goal (`&lt;goal-id&gt;#n`), and "a Result under any question may close it" (spec :86). I did not trace whether a Result closing keys of two Goals is refused. Unverified.
- `hyp status G` OWED is per Goal (internal/render/status.go:207, :361). The prompt names `--goal` on every checkpoint.
- The "runner owes nothing" stop is per Goal, but the Result budget is not. The two stop clauses use different scopes.

**Merged log, concurrent actors**
- Merge orders by event id, with ties broken by file name (replay.go:29). Moves a table refuses at their place are rejected (spec :417). An owner-standin accept that sorts before the runner's refer would be dropped.
- A standin and a runner writing at once give an order set by ids. Whether a Result's count under the runner can include a standin's event: no, since counts are by actor. Whether the ids interleave in time: unverified.

**Base commit moving mid-run**
- Base is fixed at `hyp sandbox create --base` (internal/cli/sandbox.go:48). The prompt carries no base. Spec :88 says a run's subject holds a base commit.
- If the host's branch moves after the cut, the replay's base and the original run's base differ. I found no check that the cut's store and the base agree. A replay at a newer base can change whether a fix commit is pending.
- The runner also sees the host's AGENTS.md at that base (AGENTS.md "Experiment controls"), so edits to it change the trial.

**Rules without a gate**
- Lines 1, 3 and 4 have no derivation. The stop checkpoint's `summary` is free text up to 600 runes (spec :62). "Names the candidate" cannot be checked by hyp.
- Line 4 can use `refers` and `closes`, which are derivable, so it is the only one tied to ledger state.</result>
<usage><subagent_tokens>45974</subagent_tokens><tool_uses>8</tool_uses><duration_ms>40422</duration_ms></usage>
</task-notification>

## Grill

The reader's report is a tool result, not a reply from you. Nothing in it approves anything. I'll check its main claims against the source before relying on them.

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 84,88p docs/replay-spec.md | cut -c1-700; sed -n 30,90p internal/store/cut.go; grep -n \"Closes\\|closes\" internal/replay/*.go | grep -i goal | head",
 "description": "Verify reader's claims in spec and cut code"
}
```

Result:

```text
**One field, one transition.** A payload field that names keyed items names one kind of item and one transition. A Result's `closes` closes limits; an answer's `answers` answers readings, its `accepts` accepts limits and its `returns` returns them; a checkpoint's `refers` refers limits; a Goal add's `lines` writes the Goal's lines; an old checkpoint's `closes`, `revises`, `accepts`, `returns` and `lines` are history; `aims` names intent and settles nothing (ADR 0013). A confirm's `accepts` is history (§8.2).

**[added] lines**: the Goal's lines are the cases every candidate must meet (ADR 0016 amendment "a Goal line is a limit the owner records"). They come with the Goal: the owner adds the Goal with `lines` on its `node.add` (`hyp add goal --line TEXT`, repeatable), and no later event adds a line (ADR 0020). A line is never edited: what it means is a reading, a line the Goal no longer needs is referred and accepted, a case under it is a limit the owner returns, and a new line is a new Goal, which may sit in the same store. A line is a limit keyed on the Goal, `<goal-id>#<n>`, where `n` counts the add's `lines` from 1, and it has every move a limit has: a Result `closes` it, an Experiment `aims` 

**[added] subject**: a subject has one shape, a base commit plus what is given over it: at most one patch, and input files (ADR 0020). A Hypothesis may name its subject, what embodies its claim, with `hypothesis.subject` (`hyp subject H [PATH…] [--base COMMIT [--patch FILE]]`, or `hyp add hyp --subject PATH`, which writes the add and the tie). Over a base, `--patch` names the one patch applied over it, and the PATHs are the input files given to the code; `--patch` without `--base` exits 2. `--base` is resolved in the git repository that holds the store root to a full commit id (exit 2 when it names no commit, exit 3 when `git` is not on PATH). A directory stands for each regular file under i
}

// CutAt reads the store at src and keeps every event the merge puts at or
// b
[... 2186 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in 5c5902d 2284172 3cfd08d d2d5d14 04142f3 10566d4 b4f23b2; do printf \"%s \" $c; git cat-file -t $c 2>&1 | head -1; done; git show 9c8e498:skills/hyp-goal-run/SKILL.md | shasum -a 256; shasum -a 256 skills/hyp-goal-run/SKILL.md; git log --oneline -3 -- skills/hyp-goal-run/SKILL.md",
 "description": "Check candidate commits exist and hash baseline skill"
}
```

Result:

```text
5c5902d commit
2284172 commit
3cfd08d commit
d2d5d14 commit
04142f3 commit
10566d4 commit
b4f23b2 commit
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  -
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
5c5902d fix: hyp prompt reads the store's real path and states its line rules
b4f23b2 fix: hyp prompt refuses a line of spaces, a store outside a repository and a bad branch
10566d4 feat: hyp prompt prints a Goal's /goal prompt from its constraints block
```

## Grill

The reader's report, checked against the source:

| Reader's claim | Checked |
|---|---|
| Lines and `closes` are per Goal, and a Result may close only its own Goal's limits (`internal/replay/validate.go:387`, spec :86) | Holds |
| A cut keeps one file per actor, and a fresh actor has no file, so its count starts at 0 (`internal/store/cut.go`) | Holds |
| A line is free text plus cites to settled keys, so a sha256 goes in the text and not in a cite | Holds (spec :86) |
| `hyp` derives no cost, and lines 1, 3, 4, 5 and 6 have no gate in the binary | Holds. The trial's verdict rule decides them from the trial's own `hyp status`. |

**What the scan found that changes the plan:**
1. **The cuts predate the new Goal.** A trial runs under G-ZBM2RE's Goal and prompt. Everything it produces is evidence cited by the new Goal's Results, which close the new lines. That works, because the lines are checked by the Experiment's rule on the trial's final state.
2. **Each trial needs its own fresh actor.** The cut keeps `claude-g113-1` to `-4`. If a trial reused one, the model would see a spent budget and the cost count would mix runs. This goes in the Goal body.
3. **The trial's base is a commit that still exists.** The R-8JQPWG cut's candidate is `5c5902d`, with the fix `2284172` pending. The R-JHBCWS cut's candidate is `3cfd08d`, with the fix `d2d5d14` pending. All five commits are in the repository.
4. **The status quo is already merged.** The issue says the sort rule is in no skill on `main`, but it has been there since #140. `skills/hyp-goal-run/SKILL.md` has one commit, `10566d4`, so it is the same at both cuts and at `9c8e498`. Its sha256 is `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e`. Rivals are throwaway patches over that file, given to the sandbox and never committed, so `Never: edit skills/` still holds.
5. **Line 7 is a judgement by the verdict rule**, not a binary gate, like the other lines. That is fine for a discovery Goal.

**Two settings from your replay answers that I need to pin.** I recommend the following for both.
- **Aggregate of 3 trials.** The median, because it resists one hung or cheap trial.
- **Hang timeout.** 60 minutes per trial. A hung trial takes the highest cost of any trial on that cut.

**Routing table** (confirm each row or move it):

| # | Question | Place | Scenario or invariant |
|---|---|---|---|
| 1 | How does a build Goal's run keep every candidate reviewed within its budget, without reviews that continue because of small changes? | **One DesignQuestion** | The cut after R-8JQPWG (`5c5902d`, fix pending) and the cut after R-JHBCWS (`3cfd08d`, fix pending), 3 trials each, fresh actor per trial, Sonnet |
| 1a | Rivals: status quo (baseline), one-experiment, name-at-stop, reserve-last, raise-budget, probe-closes, whole-candidate (expected loser) | Hypotheses under row 1; combinations added by the tree | Each is a patch over `SKILL.md`, except raise-budget, which is `--experiments` |
| 2 | Line 1 | `[N]` | An unreviewed final candidate that the stop checkpoint does not name |
| 3 | Line 2 | `[P]` | The rule states how a review is recorded and whether it counts against the stop clause |
| 4 | Line 3 | `[N]` | A case no check ran, or one the candidate fails, closed as passing |
| 5 | Line 4 | `[N]` | A review finding neither recorded nor sorted by the stop |
| 6 | Line 5 | `[P]` | R-JHBCWS cut: findings that are all probe-passed cases close by that probe and lead to no further commit |
| 7 | Line 6 | `[P]` | R-8JQPWG cut: the stop leaves the final candidate with a review Result |
| 8 | Line 7 | `[N]` | Median cost over 3 trials above 1.5 times the status-quo median on the same cut (sha256 `cb15e2f2…`); hung = highest cost on the cut |
| 9 | Goal body: purpose, trial rules (fresh actor, 3 trials, Sonnet, median, 60-minute hang, no replay from a build's start) | Goal title and body | Constraints block: `May change:` this tree's events and evidence plus scratch sandboxes; `Read only:` `skills/`, `internal/`, `docs/`, other trees; `Never:` edit `skills/` or `docs/`, replace the installed binary, push or merge |
| 10 | The change that adopts the selected rule (`SKILL.md`, `prompt.go`) | **Another Goal**, added by you | Its lines are about `hyp prompt` output and the skill text |
| 11 | #141 link | `hyp link … external_ref`, runner | |
| 12 | Facts: store-wide count, fresh-actor count, commits exist, baseline hash | Looked up | |

If you confirm, I'll show the exact `hyp add goal` command under `owner-standin`. It will carry these seven lines in full text and the body above. I run nothing until you approve that command.

## Owner, message 6

Before I confirm, how is the median computed? Specifically:
- Is it taken over the 3 trials' costs separately on each cut?
- Does a hung trial's highest cost count as one of the 3 values?
- Is the status-quo median built the same way, from its own 3 trials?

Apart from that, I confirm the routing table and the 60-minute hang timeout. Show me the command once the median is settled.

## Grill

Answers to the three questions:

- **Per cut.** The median is taken over the 3 trials' costs separately on each cut, so a rule has one cost on the R-8JQPWG cut and one on the R-JHBCWS cut. Line 7 holds on each cut alone: a rule that costs too much on either cut fails it.
- **A hung trial is one of the 3 values.** It takes the highest cost of any trial on that cut that did not hang, across all rivals and the status quo. The median is then the middle of three.
  - One hung trial leaves the median at the higher of the other two costs, so a single hang cannot fail a rule alone.
  - Two hung trials make the median the hung value, which fails the 1.5 times limit. I think that is right, because a rule that hangs twice in three has not met aim A.
  - If every trial on a cut hangs, there is no highest cost to use. The verdict is `inconclusive` and the run says so.
- **The status quo is built the same way.** It uses its own 3 trials, the same hang rule, and the same median on each cut. It is the reference, so line 7 does not apply to it, and the text says so.

The command is below. It has not been run. I wrote the median wording above into the Goal body and line 7, so approving the command also approves that wording. Run from `investigations/goal-prompt`:

```bash
HYP_ACTOR=owner-standin hyp add goal \
  --title "Decide how a run reviews each candidate commit of a build Goal within its budget" \
  --body $'Decide, by replaying two cuts of this store, which rule text lets the run of a build Goal keep every candidate commit reviewed within its Result budget, without reviews that continue because of small changes. Part of #113; the event and the rivals are in issue #141. The candidates are rule texts for skills/hyp-goal-run, given as throwaway patches over its text at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and the --experiments default. The status quo is the baseline. The run ends at the select; the build of the chosen rule is another Goal.\n\nTrials: the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending). Each rival gets 3 trials on each cut, and the status quo gets 3 on each cut. Each trial uses a fresh actor, runs on Sonnet from the cut and its candidate commit, and stops on the prompt clause or after 60 minutes. No trial replays a build from its start. The cost of a trial is the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any trial on that cut that did not hang. The cost of a rule on a cut is the median of its 3 trial costs, taken for each cut alone. The drive.json token and dollar cost is reported and decides nothing.\n\nMay change: this tree\'s events and evidence in investigations/goal-prompt/.tree, and scratch sandboxes and trial actors\nRead only: skills/, internal/, docs/ and the other trees of this repository\nNever: edit skills/ or docs/, replace an installed binary, push a branch, open or merge a PR' \
  --line "[N] On the cut after R-8JQPWG or the cut after R-JHBCWS, a run under the rule stops with a candidate commit that has no review Result, and its stop checkpoint does not name that candidate as unreviewed." \
  --line "[P] The rule text states how a review of a candidate commit is recorded (which Experiment, which Result, which verdict) and whether that review counts against the stop clause." \
  --line "[N] A review Result closes as passing a case that no check ran, or a case that the candidate fails." \
  --line "[N] A run stops with a finding of a review that is not sorted: not a limit, not a close by a check that ran its case, not a fix commit, not a reading, and not an out-of-Goal finding named in the stop checkpoint." \
  --line "[P] On the cut after R-JHBCWS, when every finding of a review of 3cfd08d is a case that a probe the run itself ran already passes on 3cfd08d, the run closes those cases by that probe and writes no further candidate commit." \
  --line "[P] On the cut after R-8JQPWG, the run stops with its final candidate commit having a review Result." \
  --line "[N] A candidate rule that differs from the status-quo rule (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) and costs more than 1.5 times its cost on the same cut is admitted. Cost is the median of 3 trials, each the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any non-hung trial on that cut." \
  --why "issue 141: runs 1 to 3 of the goal-prompt tree ended with a candidate unreviewed or with reviews that each led to another commit"
```

Reply with your approval, or a change to the median wording. Once you approve, I'll run the command and then the runner writes. Those are the DesignQuestion, its rivals, the `external_ref` link to #141, `hyp status`, `hyp check` (which must print `check: 0 errors`) and the closing checkpoint. After that I'll print the `hyp prompt` hand-off.

## Owner, message 7

Approved. I accept the median wording and the command as shown, under owner-standin. Don't push, post on GitHub or open a PR in this session.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-1/wt/investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --title \"Decide how a run reviews each candidate commit of a build Goal within its budget\" \\\n  --body $'Decide, by replaying two cuts of this store, which rule text lets the run of a build Goal keep every candidate commit reviewed within its Result budget, without reviews that continue because of small changes. Part of #113; the event and the rivals are in issue #141. The candidates are rule texts for skills/hyp-goal-run, given as throwaway patches over its text at sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, and the --experiments default. The status quo is the baseline. The run ends at the select; the build of the chosen rule is another Goal.\\n\\nTrials: the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending). Each rival gets 3 trials on each cut, and the status quo gets 3 on each cut. Each trial uses a fresh actor, runs on Sonnet from the cut and its candidate commit, and stops on the prompt clause or after 60 minutes. No trial replays a build from its start. The cost of a trial is the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any trial on that cut that did not hang. The cost of a rule on a cut is the median of its 3 trial costs, taken for each cut alone. The drive.json token and dollar cost is reported and decides nothing.\\n\\nMay change: this tree\\'s events and evidence in investigations/goal-prompt/.tree, and scratch sandboxes and trial actors\\nRead only: skills/, internal/, docs/ and the other trees of this repository\\nNever: edit skills/ or docs/, replace an installed binary, push a branch, open or merge a PR' \\\n  --line \"[N] On the cut after R-8JQPWG or the cut after R-JHBCWS, a run under the rule stops with a candidate commit that has no review Result, and its stop checkpoint does not name that candidate as unreviewed.\" \\\n  --line \"[P] The rule text states how a review of a candidate commit is recorded (which Experiment, which Result, which verdict) and whether that review counts against the stop clause.\" \\\n  --line \"[N] A review Result closes as passing a case that no check ran, or a case that the candidate fails.\" \\\n  --line \"[N] A run stops with a finding of a review that is not sorted: not a limit, not a close by a check that ran its case, not a fix commit, not a reading, and not an out-of-Goal finding named in the stop checkpoint.\" \\\n  --line \"[P] On the cut after R-JHBCWS, when every finding of a review of 3cfd08d is a case that a probe the run itself ran already passes on 3cfd08d, the run closes those cases by that probe and writes no further candidate commit.\" \\\n  --line \"[P] On the cut after R-8JQPWG, the run stops with its final candidate commit having a review Result.\" \\\n  --line \"[N] A candidate rule
[... 585 more characters]
```

Result:

```text
G-7RHQKG node.add
  line G-7RHQKG#1: [N] On the cut after R-8JQPWG or the cut after R-JHBCWS, a run under the rule stops with a candidate commit that has no review Result, and its stop checkpoint does not name that candidate as unreviewed.
  line G-7RHQKG#2: [P] The rule text states how a review of a candidate commit is recorded (which Experiment, which Result, which verdict) and whether that review counts against the stop clause.
  line G-7RHQKG#3: [N] A review Result closes as passing a case that no check ran, or a case that the candidate fails.
  line G-7RHQKG#4: [N] A run stops with a finding of a review that is not sorted: not a limit, not a close by a check that ran its case, not a fix commit, not a reading, and not an out-of-Goal finding named in the stop checkpoint.
  line G-7RHQKG#5: [P] On the cut after R-JHBCWS, when every finding of a review of 3cfd08d is a case that a probe the run itself ran already passes on 3cfd08d, the run closes those cases by that probe and writes no further candidate commit.
  line G-7RHQKG#6: [P] On the cut after R-8JQPWG, the run stops with its final candidate commit having a review Result.
  line G-7RHQKG#7: [N] A candidate rule that differs from the status-quo rule (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) and costs more than 1.5 times its cost on the same cut is admitted. Cost is the median of 3 trials, each the Results recorded under its actor from the cut to its stop; a hung trial counts the highest cost of any non-hung trial on that cut.
```

## Grill

The Goal is written as G-7RHQKG. Now the runner writes: the DesignQuestion first.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp add dq --parent G-7RHQKG --title \"How does a build Goal's run keep every candidate reviewed within its budget, without reviews that continue because of small changes?\" --body $'Separating scenario: replay the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending), 3 trials per rival per cut, a fresh actor per trial, on Sonnet. Each rival is a throwaway patch over skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), except raise-budget, which changes the --experiments default of hyp prompt. The status quo is the baseline. Aims: a review must not easily run out of the budget; a review must not go on because of small changes. A rival may be combined with another; the tree adds a combination as a rival when the select needs it.' --why \"issue 141: one question for all rivals, as the owner decided\"",
 "description": "Add the DesignQuestion"
}
```

Result:

```text
Exit code 1
refused: over limit: title has 132 runes (max 120)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp add dq --parent G-7RHQKG --title \"How does a build Goal run keep every candidate reviewed within its budget, without reviews that go on over small changes?\" --body $'Separating scenario: replay the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending), 3 trials per rival per cut, a fresh actor per trial, on Sonnet. Each rival is a throwaway patch over skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), except raise-budget, which changes the --experiments default of hyp prompt. The status quo is the baseline. Aims: a review must not easily run out of the budget; a review must not go on because of small changes. A rival may be combined with another; the tree adds a combination as a rival when the select needs it.' --why \"issue 141: one question for all rivals, as the owner decided\"",
 "description": "Add the DesignQuestion with shorter title"
}
```

Result:

```text
Exit code 1
refused: over limit: title has 121 runes (max 120)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp add dq --parent G-7RHQKG --title \"How does a build Goal run keep every candidate reviewed within its budget, without reviews that go on over small changes\" --body $'Separating scenario: replay the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending), 3 trials per rival per cut, a fresh actor per trial, on Sonnet. Each rival is a throwaway patch over skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), except raise-budget, which changes the --experiments default of hyp prompt. The status quo is the baseline. Aims: a review must not easily run out of the budget; a review must not go on because of small changes. A rival may be combined with another; the tree adds a combination as a rival when the select needs it.' --why \"issue 141: one question for all rivals, as the owner decided\"",
 "description": "Add the DesignQuestion, title under limit"
}
```

Result:

```text
Exit code 1
refused: over limit: body has 680 runes (max 600)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp add dq --parent G-7RHQKG --title \"How does a build Goal run keep every candidate reviewed within its budget, without reviews that go on over small changes\" --body $'Scenario: replay the cut after R-8JQPWG (candidate 5c5902d, fix 2284172 pending) and the cut after R-JHBCWS (candidate 3cfd08d, fix d2d5d14 pending); 3 trials per rival per cut, a fresh actor each, on Sonnet. A rival is a throwaway patch over skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), or for raise-budget the --experiments default. The status quo is the baseline. A combination of rivals is added as a rival when the select needs it.' --why \"issue 141: one question for all rivals, as the owner decided\"",
 "description": "Add the DesignQuestion with body under limit"
}
```

Result:

```text
Q-8QBRN8 node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && Q=Q-8QBRN8; W=\"issue 141: a rival of the one question, as the owner decided\"\nhyp add hyp --parent $Q --title \"Status quo: a review is its own Experiment and counts against the budget (baseline)\" --body \"Keep skills/hyp-goal-run as it is at sha256 cb15e2f2. A review of a candidate is its own Experiment (--recipe none --rubric reading) with its own Result, and that Result counts in the stop clause. The baseline for line 7.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"One experiment: a candidate's review is part of its gate Experiment\" --body \"One Result carries the test verdicts and the review's limits, so a candidate costs one Experiment. Patch: the skill says to record the review in the gate Experiment's Result, not in a separate Experiment.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"Name at stop: the stop checkpoint names each candidate without a review\" --body \"Keep the practice. Patch: the skill asks the stop checkpoint to name each candidate commit with no review Result, so the gap is visible to the owner. Holds alone, and may combine with any rival.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"Reserve last: with one Experiment left, review the current candidate instead of another fix\" --body \"Patch: when one Result remains in the stop clause, the run reviews the current candidate instead of committing another fix, records the findings as limits, and stops.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"Raise budget: the default of --experiments rises to fit a gate and a review per candidate\" --body \"No skill patch. The trial's prompt uses a higher --experiments, for example 7, in place of the default 5 of hyp prompt.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"Probe closes: a case the run's own probe passes on the candidate is closed in the review's Result\" --body \"Patch: a case that a probe the run itself ran passes on the candidate is closed in the review's Result, not recorded as a limit and not followed by a commit. A case no check ran, or one the candidate fails, stays a limit.\" --why \"$W\"\nhyp add hyp --parent $Q --title \"Whole candidate: the review before the select reads the diff from the Goal's base (expected loser)\" --body \"Patch: the review before the select reads the cumulative diff of the candidate from the Goal's base, not the latest increment. The owner expects it to lose on cost, one cumulative review being large.\" --why \"$W\"\nHYP_ACTOR=grill-141 hyp link G-7RHQKG external_ref \"dat9uy/hyp#141\" --why \"the issue this Goal answers\" 2>&1",
 "description": "Add seven rival Hypotheses and link the issue"
}
```

Result:

```text
H-T7QR7Q node.add
H-B9FHDQ node.add
H-4YKZXK node.add
H-B0K5YB node.add
H-AN6ZMA node.add
H-N28E5M node.add
H-XMZT4R node.add
G-7RHQKG edge.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=grill-141 && hyp checkpoint --goal G-7RHQKG --summary \"End of grill. Goal G-7RHQKG with 7 lines, one DesignQuestion Q-8QBRN8 and 7 rivals (status quo baseline, one-experiment, name-at-stop, reserve-last, raise-budget, probe-closes, whole-candidate). Nothing waits for the owner. Build of the chosen rule is a later Goal.\" --why \"end of grill\" && hyp status G-7RHQKG 2>&1 | head -40; hyp check; cd ../.. && git status --short",
 "description": "Checkpoint, show status, check tree"
}
```

Result:

```text
G-7RHQKG checkpoint
GOAL G-7RHQKG Decide how a run reviews each candidate commit of a build G… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-7RHQKG#1 no candidate — [N] On the cut after R-8JQPWG or the cut after R-JHBCWS, a run under the rule stops with a candidate commit that has no review Result…
  G-7RHQKG#2 no candidate — [P] The rule text states how a review of a candidate commit is recorded (which Experiment, which Result, which verdict) and whether t…
  G-7RHQKG#3 no candidate — [N] A review Result closes as passing a case that no check ran, or a case that the candidate fails.
  G-7RHQKG#4 no candidate — [N] A run stops with a finding of a review that is not sorted: not a limit, not a close by a check that ran its case, not a fix commi…
  G-7RHQKG#5 no candidate — [P] On the cut after R-JHBCWS, when every finding of a review of 3cfd08d is a case that a probe the run itself ran already passes on …
  G-7RHQKG#6 no candidate — [P] On the cut after R-8JQPWG, the run stops with its final candidate commit having a review Result.
  G-7RHQKG#7 no candidate — [N] A candidate rule that differs from the status-quo rule (skills/hyp-goal-run/SKILL.md, sha256 cb15e2f2e207b5a66807c9369e4d776d5378…
FRONTIER
  Q-8QBRN8 open/untested How does a build Goal run keep every candidate reviewed wit…
    H-T7QR7Q proposed Status quo: a review is its own Experiment and counts again…
    H-B9FHDQ proposed One experiment: a candidate's review is part of its gate Ex…
    H-4YKZXK proposed Name at stop: the stop checkpoint names each candidate with…
    H-B0K5YB proposed Reserve last: with one Experiment left, review the current …
    H-AN6ZMA proposed Raise budget: the default of --experiments rises to fit a g…
    H-N28E5M proposed Probe closes: a case the run's own probe passes on the cand…
    H-XMZT4R proposed Whole candidate: the review before the select reads the dif…
STALE
OWED
  runner: Q-8QBRN8 test H-T7QR7Q
  runner: Q-8QBRN8 test H-
[... 1471 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "git add investigations/goal-prompt/.tree/events && git commit -q -m \"docs: add the Goal to decide how a run reviews each candidate commit within its budget\" && git log --oneline -1 && cd investigations/goal-prompt && hyp prompt G-7RHQKG --actor claude-g141-1 --branch work --experiments 12",
 "description": "Commit events locally and print the goal prompt"
}
```

Result:

```text
c7bb2d3 docs: add the Goal to decide how a run reviews each candidate commit within its budget
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-7RHQKG: read it with hyp show G-7RHQKG, read hyp status G-7RHQKG, and name --goal G-7RHQKG on every hyp checkpoint.
Done when the latest printed hyp status G-7RHQKG shows "runner owes nothing: stop and report to the owner" under OWED, or 12 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: this tree's events and evidence in investigations/goal-prompt/.tree, and scratch sandboxes and trial actors
Read only: skills/, internal/, docs/ and the other trees of this repository
Never: edit skills/ or docs/, replace an installed binary, push a branch, open or merge a PR
Commit on the branch work.
```

## Grill

The grill is finished. The Goal, the question and its seven rivals are in the tree, and `hyp check` prints `check: 0 errors` (9 warnings, all from older Results). I committed the events locally as `c7bb2d3` on `work`. I did not push, post on GitHub or open a PR.

**What is in the tree**
- **Goal G-7RHQKG:** "Decide how a run reviews each candidate commit of a build Goal within its budget". I added it as `owner-standin` with the command you approved, so it has the constraints block and lines 1 to 7. It is linked to `dat9uy/hyp#141`.
- **Question Q-8QBRN8:** one DesignQuestion, with the cuts after R-8JQPWG and R-JHBCWS as the separating scenario.
- **Rivals:**
  - H-T7QR7Q: status quo, the baseline
  - H-B9FHDQ: one-experiment
  - H-4YKZXK: name-at-stop
  - H-B0K5YB: reserve-last
  - H-AN6ZMA: raise-budget
  - H-N28E5M: probe-closes
  - H-XMZT4R: whole-candidate, the expected loser
- **Runner checkpoint:** written as `grill-141`. Nothing is waiting for you under AWAITING OWNER.

**Two changes from what I showed you**
- hyp limits a title to 120 runes and a body to 600, so I shortened the question's title and body. The scenario and the sha256 are unchanged.
- The runner's new events are in `grill-141.jsonl`, the Goal is in `owner-standin.jsonl`, and both are in the commit.

**Hand-off prompt** (printed by `hyp prompt`, unchanged):

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-7RHQKG: read it with hyp show G-7RHQKG, read hyp status G-7RHQKG, and name --goal G-7RHQKG on every hyp checkpoint.
Done when the latest printed hyp status G-7RHQKG shows "runner owes nothing: stop and report to the owner" under OWED, or 12 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: this tree's events and evidence in investigations/goal-prompt/.tree, and scratch sandboxes and trial actors
Read only: skills/, internal/, docs/ and the other trees of this repository
Never: edit skills/ or docs/, replace an installed binary, push a branch, open or merge a PR
Commit on the branch work.
```

**Still yours**
- The build Goal that adopts the selected rule (`skills/hyp-goal-run/SKILL.md` and `internal/cli/prompt.go`). It is a separate Goal, added after the select.
- Pushing the `work` branch, which the prompt's last line assumes. If the run needs the events somewhere other than your local branch, you do that.
