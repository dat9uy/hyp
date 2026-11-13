# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-pre-b-2/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue and checking the store.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-b-2/tmp/issue-141.md"
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
 "command": "which hyp; ls; ls investigations; hyp status 2>&1 | head -40"
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-pre-b-2/bin/hyp
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
no .tree store found; run 'hyp init'
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt/investigations/goal-prompt && ls -a; hyp status 2>&1 | head -80; git -C ../.. log --oneline | head -5; hyp version 2>&1 | head -2"
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
9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
04142f3 fix: test that a bad block line in a store outside every repository exits 1
d2d5d14 fix: test that an invalid block in a store outside every repository exits 1
hyp 0.1.0 (events v1)
```

### Tool Read (cwd: investigations/goal-prompt)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-pre-b-2/wt/skills/hyp-goal-run/SKILL.md"
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

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp show G-ZBM2RE 2>&1 | head -70; ls evals; cd ../..; grep -rn \"experiments\" internal/*/prompt*.go cmd/hyp/*.go | grep -v _test | head -20; ls docs docs/adr | head -60"
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
[... 13607 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rniE \"review\" docs/glossary.md docs/principles.md docs/how-hyp-improves-itself.md | head -30; echo ---; grep -rnE \"R-8JQPWG|E-HQWTA0\" investigations/goal-prompt/.tree --include=*.json* -l | head; ls investigations/goal-prompt/.tree investigations/goal-prompt/.tree/events | head -30; hyp --help 2>&1 | head -60"
}
```

Result:

```text
docs/glossary.md:3:The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.
docs/glossary.md:46:A Goal whose candidate is a commit: building is a Goal like any other. A line asserts one property of the candidate, decided by its own tests or one pinned instrument: `[P]` pins what a valid input produces, and `[N]` pins that a forbidden state is refused. Each commit is a Hypothesis whose subject is the base plus that commit's change, and a fix is a new Hypothesis that supersedes it. Each Experiment's verdict rule asks for red on base and green on the candidate; the base of the red run's subject is where the build starts, and the Goal records no start commit. A finding of a run's review that is a case of a line is a limit; one that is not is an out-of-Goal finding. A change to a skill is a build Goal too (ADR 0021).
docs/glossary.md:145:_Avoid_: approve, review; verdict (alone: a confirm's answer was a confirm verdict)
docs/glossary.md:155:Retired by ADR 0015. It was the role that confirmed every select, independently of the actor who made it. Old verifier confirms replay as history and apply nothing, and `hyp confirm` is retired (ADR 0019). Any actor may still review: a defect is recorded as a Result, a retract or a limit. In code: `event.RoleVerifier`, read only from old logs.
docs/glossary.md:156:_Avoid_: reviewer, checker (as a role)
docs/glossary.md:179:A finding, often from a run's review, that is a case of none of the Goal's lines and asks no question about what a line means. It is not a limit, 
[... 4959 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rnE \"budget|--experiments|Results recorded|cap\\b|5 Results|5 Experiments\" skills docs AGENTS.md README.md internal/cli/prompt.go | grep -v \"^docs/adr\" | head -30; git branch -a | head; git log --oneline -3 -- skills/hyp-goal-run"
}
```

Result:

```text
skills/hyp-owner-grill/SKILL.md:50:- **Merge states one invariant guards.** At most 10 lines of 1 to 600 runes each; the cap forces the merging.
skills/hyp-owner-grill/SKILL.md:113:hyp prompt <G-id> --actor <runner-actor> [--branch <branch>] [--pr <N>] [--experiments <N>]
docs/replay-spec.md:246:- A **pin** is a recipe ref (`name@version`) and the hash an event recorded for it at write time. A bundle is its ref and hash; a `recipes/<name>@<version>/` directory name is only the library's lookup key, and hyp opens each copy it makes by the ref it holds ([ADR 0005](adr/0005-a-bundle-is-its-ref-and-hash.md)). A Result depends on its Experiment's pin (`recipe` and `recipe_hash`, unless the recipe is `none`), on its Experiment's rubric pin (`rubric` and `rubric_hash`, when it names a bundle), and on one pin per entry of its own `evidence_hashes`, whose ref is the pointer's first segment under `recipes/` after cleaning the path. A Result is `questionable_evidence` when, for any pin it depends on, `recipeHashes[ref]` is missing or differs from the pinned hash. A new version of a recipe is a new ref, so it never flags Results recorded under an older one. `recipeHashes` leaves out a bundle that cannot be hashed (a symlink, a file or directory name outside `[A-Za-z0-9._-]`, or a `recipe.md` whose name or version does not match its directory; a name below a known answer's `known-answers/NAME/input/` may also hold `@`, so an input can hold a store with its pinned bundles) and a ref whose copies hash differently, and `hyp check` names each one. A bundle that hashes but does not open stays in `recipeHashes`, so its pins verify, and `hyp check` names it too (sandbox spec §1.2). This is a flag only and never changes any status.
docs/how-hyp-improves-itself.md:29:  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
internal/cli/prompt.go:20:const promptUse = "usage: hyp prompt G-id --actor A [--branch B] [--pr N] [--experiments N]"
internal/cli/prompt.go:109:	exp
[... 436 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt/investigations/goal-prompt && grep -l \"8JQPWG\" .tree/events/* ; hyp show R-8JQPWG 2>&1 | head -15; hyp show R-JHBCWS 2>&1 | head -8; grep -n \"cut\" ../../docs/sandbox-spec.md | head -12; cat ../../docs/adr/0020-one-store-holds-many-goals.md | head -30; ls evals ../../.claude/skills"
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
[... 3399 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt && grep -n -iE \"\\-\\-store|\\-\\-cut|holds a .tree|\\.tree\" docs/sandbox-spec.md | head -30; grep -rn \"runner\" docs/glossary.md | head -12"
}
```

Result:

```text
30:Evidence that the operator gathers by hand while the sandbox exists is not a command, and is unchanged: it is the files saved under `.tree/evidence/<E-id>/`. Scripted Evidence is the same step done by a command, which is why it runs after Cleanup: it reads only the kept files, so it can be run again.
147:7. The command record is a directory the operator names with `--record`, so it can sit under `.tree/evidence/<E-id>/` and be cited, and large output stays in its log. The commands run from a temporary copy, so a script cannot change the library version it ran from.
docs/glossary.md:65:The section of `hyp status` that names the one step due from the owner on each question the Frontier prints: `choose: H…, H…` when they should choose between the candidates and the runner owes nothing on the question. A question the Frontier does not print asks nothing, and the choice is owed again once it prints. It also lists every open reading, which blocks nothing, and every referred limit on a candidate. `hyp status G-…` lists one Goal's items; in a store with many Goals, `hyp status` prints one Awaiting owner list, grouped by Goal.
docs/glossary.md:69:A reading asks what the owner's words mean, and a trade-off asks which candidate to choose. Closing a reading changes no status, because statuses come only from verdicts. An answer that rules out a candidate takes effect when a later Result records `contradicts` on it. A closed reading is owed by the runner until a Result applies it (`hyp record --applies`, from an Experiment that aims at it), and the owner is not asked to choose on its question until then. An owner who changes an answer answers the reading again (`hyp answer --answer R-…#rN=TEXT`): the new words are the answer in force from then on, the earlier answer stays in history, and the reading is owed by the runner again until a later Result applies it. A revision changes no earlier event or verdict. One verb serves both: whether an answer is a first answer or a revision
[... 9626 more characters]
```

## Grill

Checking whether a store cut works for the two scenarios before I send round 1.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n -E -- \"--store|--cut|store cut\" docs/sandbox-spec.md docs/glossary.md | cut -c1-400 | head -14; git cat-file -t 10566d4 b4f23b2 5c5902d 2cd 2>&1 | head; cd investigations/goal-prompt; grep -h \"R-8JQPWG\\|R-JHBCWS\" .tree/events/*.jsonl | head -c 600; echo; hyp events --json 2>/dev/null | head -c 400"
}
```

Result:

```text
docs/sandbox-spec.md:158:A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (see Arm in [glossary.md](glossary.md)), given as `--patch ARM=FILE`.
docs/glossary.md:359:A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. The sandbox is the only place hyp assembles a subject: from the base, a store cut, per-run inputs and one patch for each Arm. Code under test is built in its Launch ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). It is not process isolation: commands still run 
fatal: too many arguments

usage: git cat-file <type> <object>
   or: git cat-file (-e | -p | -t | -s) <object>
   or: git cat-file (--textconv | --filters)
                    [<rev>:<path|tree-ish> | --path=<path|tree-ish> <rev>]
   or: git cat-file (--batch | --batch-check | --batch-command) [--batch-all-objects]
                    [--buffer] [--follow-symlinks] [--unordered]
                    [--textconv | --filters] [-Z]

{"v":1,"id":"01M75T6JV8AC7H99GHHPYXSJEM","ts":"2026-11-11T17:23:58.952283Z","actor":"claude-g113-1","type":"node.add","node":"01M75T6JV8AC7H99GHHPYXSJEM","payload":{"kind":"Hypothesis","parent":"01M75R48JNGWS8ENJHN7Z93VJ5","title":"feat/goal-prompt at 5c5902d: b4f23b2 plus fixes for second-review findings G1-G8","body":"5c5902d holds b4f23b2, reads the store's real path through symlinks, pins a tab and a CRLF blank line as bad block lines, checks every Goal id in the prompt, guards the no-repository test, and fixes the --branch message and the grill, answer and goal-run text.","alias":"h-5c590
{"format":1,"events":[{"v":1,"id":"01M75R06QEHM6EG2PSEEGZR6C6","ts":"2026-11-11T16:45:32.782596Z","actor":"owner-dat","type":"node.add","node":"01M75R06QEHM6EG2PSEEGZR6C6","payload":{"kind":"Goal","parent":"","title":"Record how hyp makes the /goal prompt
[... 145 more characters]
```

## Grill

Round 1. The store is `investigations/goal-prompt`, which holds #141's parent Goals G-GZR6C6 (the design record) and G-ZBM2RE (the build). I looked up what I could:

- **What the stop clause counts.** It counts Results recorded under the actor (`internal/cli/prompt.go:167`), not Experiments. The default of 5 comes from `--experiments`.
- **Both cuts can be replayed.** `hyp sandbox create` takes a store cut (`--store SRC --cut EVENT`). The R-8JQPWG record event is in the claude-g113-1 and claude-g113-2 logs, and the three candidate commits are in git history.
- **Why every review costs a Result.** A limit needs a Result and a Result needs an Experiment (glossary, "Limit"; ADR 0018). I won't ask about that.

**1. Is this Goal a design Goal, with the build as a separate Goal later?**
- Today: the issue's "Done means" has two parts. One is to select a rival from a run on the scenario. The other is the changed behaviour after the change, which is edits to `hyp-goal-run` and `hyp prompt`.
- Option A, one design Goal now:
  - Its candidates are rival run rules. Each rule is given to a runner as stand-in text on the cut, and the runner is not a commit.
  - Example: the run replays cut A under H-reserve-last's text and checks where the replay stops.
  - After the select, a build Goal gets its own grill, with lines that quote the selected rule. This is the same split as G-GZR6C6 and G-ZBM2RE.
- Option B, one Goal holding both:
  - Its lines would be about a commit that doesn't exist yet.
  - Every rival would be owed lines about skill text that only the winner gets built into.
- **Recommend A.** A candidate that is a claim and a candidate that is a commit can't share lines.

**2. Do the rivals belong to one question, and is H-whole-candidate in this Goal?**
- Today: the issue lists H-one-experiment, H-name-at-stop, H-reserve-last and H-raise-budget. Your comments add H-probe-closes and H-whole-candidate.
- The first five answer one question: what makes a run reach its stop with its last candidate reviewed or named? Their scenario is the end of a budget-bound run.
- H-whole-candidate answers a different question: what does the review read? On cut A or B it changes no stop, so no line about the stop could separate it. hyp owes every line on every top-level question, so that line would be unfillable.
- Option A:
  - One DesignQuestion with five rivals, and H-whole-candidate as its own Goal, to be added by you.
  - The run adds a combination rival only if no single rival holds on both cuts. This is hyp-goal-run's own "before a trade-off" practice, and it answers the issue's "which hold alone".
- Option B: H-whole-candidate as a second DesignQuestion in this Goal, with a second line "the final candidate's cumulative diff from the base is reviewed". Every rival for the first question would then fail that line.
- **Recommend A.** Do you want it as a separate Goal?

**3. What must a line require at the stop?**
- Today: the issue accepts either end. The final candidate has a review Result, or the stop checkpoint names it as unreviewed.
- Option A, floor only:
  - Line: no candidate is left both unreviewed and unnamed at the stop, on either cut.
  - Every rival meets the floor. They differ on whether a review happened, and several can survive.
  - You choose among the survivors, which is the owner's turn to choose (ADR 0015).
- Option B, "reviewed" as a positive line:
  - H-name-at-stop alone is then contradicted. H-reserve-last is the only way to meet it on cut A, and the lines pick the winner.
- **Recommend A.** It is faithful to "either", and B decides a trade-off for you.

**4. Which lines (all [N] unless marked)?** All three are checked by replay on the two cuts, cut A and cut B, which are listed in question 5.
1. [N] On either cut, the run stops with a candidate commit that has no review Result and is not named as unreviewed in the stop checkpoint.
2. [N] On either cut, the run records more Results under its actor than the stop clause printed for it allows.
3. [P] On either cut, the stop checkpoint and the run's last message state, for each candidate commit, either its review Result or that it has none.

Line 3 is the stop-report side of line 1. Do you want 1 and 3 merged? That would be one invariant with an exception, but it would run to about 400 characters. I'd keep line 3 apart, because line 1 forbids a state and line 3 pins the report. "Say how a review is recorded and whether it counts against the budget" is a line of the build Goal. I'll note it in the Goal body and add it to that Goal.

**5. How is the scenario run? Use ak-skill-creator's eval flow, and with what model, trials and budget?**
- Today: ak-skill-creator owns the test method for skill text: matched original and variant runs on cases.
- The cases are:
  - Cut A: G-ZBM2RE after R-8JQPWG. One Result is left and a fix commit is pending.
  - Cut B: run 3's cut after R-JHBCWS. A probe passed.
- Each rival is a stand-in text for hyp-goal-run (or the printed clause), given to the runner. Nothing under `skills/` is edited.
- **Recommend yes** to the flow. For model, trials and budget: the model that ran your `/goal` runs, 3 trials per rival per cut, and `--experiments 8`. That is one Experiment per rival, one for any combination, and spare room. The default of 5 would stop this run before it finished, which is the issue's own failure.
- Do you accept these, or give your own?

**6. Constraints block (for approval).**
```text
May change: this repository, in commits on the branch the prompt names that follow AGENTS.md; the run may push that branch and open or update one PR, and never merges it. It changes investigations/goal-prompt (.tree/ and its evidence) and may add a new version of a recipe or rubric under recipes/ or rubrics/. It edits no skill, doc or code.
Read only: skills/, docs/, internal/, cmd/, plans/, every other investigations/ tree, and ak-skill-creator under .claude/skills.
Never: commit to main or merge a PR.
Never: replace an installed binary; compile any candidate to a separate path.
Never: edit a frozen recipe or rubric version, another investigations/ tree or another repository.
Never: write under an owner's actor.
```
The change belongs to the build Goal, so `May change:` here lists no skill or doc. Any edit to the texts this rule makes wrong waits for that Goal.

**7. Your actor name.** For example `owner-dat`. I'll use `HYP_ACTOR=claude-grill-141` for my runner writes.

I'll hold the preflight scan and the independent review until you answer. After that I'll send the routing table.

## Owner, message 2

1 A, agree. It's a new Goal in goal-prompt, as part of #113.

2. I wanted all the rivals under one question and all tested. If a line can't separate H-whole-candidate from the others, I accept your reason and take it as a separate Goal that I'll add. Treat it as the expected loser, because one cumulative review costs a lot. Is that what you mean by A? I also agree to a combination rival if no single rival holds.

3. A, agree.

4. Keep 3 apart from 1. Two things are missing:
- How do the lines serve my two aims? A review must not easily run out of the run's budget, and it must not go on and on over small changes.
- Where is the cost limit? A candidate may cost at most 1.5 times the base at the same cut. How do you compute cost? If you use a median, how is it computed? A trial that hangs counts as the most expensive.

5. Yes to the eval flow, but run it on Sonnet, not "the model that ran /goal". Use 3 trials per rival per cut, and `--experiments 8` is fine. Don't replay a whole build from its start; the cuts are enough. Could the ak-context-engineering skill help with how to evaluate? Look at it once.

6. Fine, except that the grill pushes nothing and opens no PR in this session.

7. Use `owner-standin` as my actor name.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills .claude/skills/ak-skill-creator ~/.claude/skills 2>&1 | head -40; find / -type d -name \"ak-context-engineering*\" 2>/dev/null | head"
}
```

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rniE \"token|duration|timing|timeout|median|trials|runs_per\" .claude/skills/ak-skill-creator/SKILL.md .claude/skills/ak-skill-creator/references 2>/dev/null | cut -c1-220 | head -30"
}
```

Result:

```text
.claude/skills/ak-skill-creator/SKILL.md:123:| Layout, metadata fields and context loading | `references/skill-anatomy-and-requirements.md`, `references/yaml-frontmatter-reference.md`, `references/token-efficiency-criter
.claude/skills/ak-skill-creator/references/benchmark-optimization-guide.md:38:user decisions. Record quality/time/token/correction tradeoffs. A shorter prompt,
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:24:Runs at startup without credential helpers. Set auth tokens in environment:
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:28:| GitHub | `GITHUB_TOKEN` or `GH_TOKEN` | PAT or GitHub App token |
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:29:| GitLab | `GITLAB_TOKEN` or `GL_TOKEN` | PAT or project token |
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:30:| Bitbucket | `BITBUCKET_TOKEN` | App password or repo token |
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:33:export GITHUB_TOKEN=ghp_xxxxxxxxxxxxxxxxxxxx
.claude/skills/ak-skill-creator/references/plugin-marketplace-hosting.md:36:CI/CD: configure as secret env variable. GitHub Actions auto-provides `GITHUB_TOKEN`.
.claude/skills/ak-skill-creator/references/skillmark-benchmark-criteria.md:9:| Type | Purpose | Default Timeout | Scorer |
.claude/skills/ak-skill-creator/references/skillmark-benchmark-criteria.md:71:`accuracy`, `tokensTotal`, `tokensInput`, `tokensOutput`, `durationMs`, `toolCount`, `costUsd`
.claude/skills/ak-skill-creator/references/skillmark-benchmark-criteria.md:86:timeout: 600
.claude/skills/ak-skill-creator/references/creator-consumer-evaluation.md:49:success/invariants, then corrections, tokens, latency and tool calls. Do not hide
.claude/skills/ak-skill-creator/references/validation-checklist.md:49:- [ ] No "sacrifice grammar", "ensure token efficiency", or other copied boilerplate
.claude/skills/ak-skill-creator/references/script-quality-c
[... 2432 more characters]
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

.claude/skills/ak-skill-creator:
assets
LICENSE-MIT.txt
LICENSE.txt
NOTICE.txt
references
scripts
SKILL.md

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
 "command": "D=/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering; ls $D $D/references 2>/dev/null; head -c 5000 $D/SKILL.md; echo ======; sed -n 35,90p /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt/.claude/skills/ak-skill-creator/references/evaluation-tools.md"
}
```

Result:

```text
Exit code 1
/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering:
references
scripts
SKILL.md

/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering/references:
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
4. **Execute:** search for the owner, read a sufficient logical unit, bound tool
   output while preserving errors, and record decisions/evidence worth retaining.
5. **Reassess:** expand only when evidence is mis
[... 3563 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "D=/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering; tail -c +5000 $D/SKILL.md | head -c 2500; echo =====; sed -n 1,80p $D/references/evaluation.md | cut -c1-240; echo =====; sed -n 30,90p /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt/.claude/skills/ak-skill-creator/references/evaluation-tools.md"
}
```

Result:

```text
Exit code 1
m.
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

- `scripts/context_analyzer.py estimate <paths> [--limit N]`: approximate file sizes.
- `scripts/context_analyzer.py analyze <messages.json> [--limit N] [--used-tokens N]`:
  capacity and lexical diagnostics; no measured health or poisoning score.
- `scripts/context_analyzer.py budget`: separate window and cumulative task accounting.
- `scripts/compression_evaluator.py`: lexical diagnostics or explicitly supplied grading.
- `scripts/benchmark_metrics.py`: aggregate supplied run records, not 
[... 536 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "D=/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering; python3 -I $D/scripts/benchmark_metrics.py --help 2>&1 | head -30; grep -nE \"median|timeout|timed|censor|p50|mean|average\" $D/scripts/benchmark_metrics.py | head -15; grep -nEi \"median|timeout|hang|censor|trial|variance\" $D/references/evaluation.md | cut -c1-200 | head"
}
```

Result:

```text
Traceback (most recent call last):
  File "/System/Volumes/Data/Users/datguy/.agentkit/cache/kits/engineer/claude-code/3.0.0/engineer/skills/ak-context-engineering/scripts/benchmark_metrics.py", line 13, in <module>
    from benchmark_statistics import summarize
ModuleNotFoundError: No module named 'benchmark_statistics'
63:        required = ("average_cost_per_task", "average_duration_seconds",
64:                    "average_agent_steps", "first_attempt_success_rate")
76:    axes = ("estimated_cost_per_first_success", "average_duration_seconds")
88:        summaries[item]["average_agent_steps"]["value"], item))
122:            # fan-out bought; zero means a single pass would have found the same.
10:1. Select representative tasks across lookup, bug repair, cross-file changes and long
13:   tools and limits. Add model/effort candidates as separate configurations, not hidden changes.
16:4. Use matched task/trial IDs and multiple independent trials when feasible. Record failed,
46:    "rows":[{"task_id":"<task>","trial_id":"1","attempt":1,
56:Attempts must start at 1, be contiguous per task/trial, and stop after a success. Rows
72:- It intersects selected task/trial cohorts, reports exclusions, and **recomputes** metrics
98:`model`, `effort`, and `harness` identical when those conditions did not change;
```

### Tool Bash (cwd: .)

```text
{
 "command": "shasum -a 256 skills/hyp-goal-run/SKILL.md ~/.local/bin/hyp /Users/datguy/.hyp-sandbox/g146-pre-b-2/bin/hyp 2>&1; git rev-parse --short HEAD; cd investigations/goal-prompt && hyp events --json | python3 -c \"\nimport json,sys\nd=json.load(sys.stdin)\nfor e in d['events']:\n    if e['type']=='result.record':\n        print(e['id'],e['actor'],e['ts'][:19],e['payload'].get('id') or e['node'])\n\" | head -30"
}
```

Result:

```text
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c  /Users/datguy/.local/bin/hyp
a0e464024c2907cb275a8fc460a218b01ffaafb52897547a434bf1d5b9fc5687  /Users/datguy/.hyp-sandbox/g146-pre-b-2/bin/hyp
9c8e498
01M75R21F8TYTTJZ8Y9CFCEZY6 claude-grill-113 2026-11-11T16:46:32 01M75R21F8TYTTJZ8Y9CFCEZY6
01M75R21G7H1J6RARDMM26GPDW claude-grill-113 2026-11-11T16:46:32 01M75R21G7H1J6RARDMM26GPDW
01M75S97KWYDNTGA2N56KPWZZC claude-g113-1 2026-11-11T17:07:57 01M75S97KWYDNTGA2N56KPWZZC
01M75SBVS853GXX0C0SC9V3ER0 claude-g113-1 2026-11-11T17:09:23 01M75SBVS853GXX0C0SC9V3ER0
01M75SV7JAPQZNHTTZB050GBXM claude-g113-1 2026-11-11T17:17:46 01M75SV7JAPQZNHTTZB050GBXM
01M75T1YRF2R2Z99DX0E8JQPWG claude-g113-1 2026-11-11T17:21:27 01M75T1YRF2R2Z99DX0E8JQPWG
01M75TJFS2Z8ZNKEX6Q4FRF929 claude-g113-1 2026-11-11T17:30:29 01M75TJFS2Z8ZNKEX6Q4FRF929
01M75TJZK3V9F85CR6TW9EGPYD claude-g113-1 2026-11-11T17:30:45 01M75TJZK3V9F85CR6TW9EGPYD
01M76TFWZTR7R6QCZ3C30E8PFH claude-g113-2 2026-11-12T02:48:18 01M76TFWZTR7R6QCZ3C30E8PFH
01M76TP5NF2DETM36QC7KYKWH9 claude-g113-2 2026-11-12T02:51:44 01M76TP5NF2DETM36QC7KYKWH9
01M76TW97XFDSKQAKNFK28A8NB claude-g113-2 2026-11-12T02:55:04 01M76TW97XFDSKQAKNFK28A8NB
01M76TZSMHENAYGH77MR3YJ7T2 claude-g113-2 2026-11-12T02:56:59 01M76TZSMHENAYGH77MR3YJ7T2
01M76V4HWRMAZ0TQECYR356DV0 claude-g113-2 2026-11-12T02:59:35 01M76V4HWRMAZ0TQECYR356DV0
01M76VXNN2ZW5WAXJ194JHBCWS claude-g113-3 2026-11-12T03:13:18 01M76VXNN2ZW5WAXJ194JHBCWS
01M76W29E72EGEB3E5ED27Y1EV claude-g113-3 2026-11-12T03:15:49 01M76W29E72EGEB3E5ED27Y1EV
01M76W7STKAZ2B3QA1TRDER6F8 claude-g113-3 2026-11-12T03:18:50 01M76W7STKAZ2B3QA1TRDER6F8
01M76WG91ZTMZJEW0V57BSWFXG claude-g113-3 2026-11-12T03:23:28 01M76WG91ZTMZJEW0V57BSWFXG
01M76WH13EXC2KFQV151HXEBAF claude-g113-3 2026-11-12T03:23:52 01M76WH13EXC2KFQV151HXEBAF
```

## Grill

I read ak-context-engineering (`SKILL.md` and `references/evaluation.md`), and I've worked your two aims and the cost cap into the lines. I did not run its scripts.

**Your answers**
- **Q2:** Yes, that is option A. All five rivals go under one question and all are tested. H-whole-candidate is a separate Goal that you add, and it is the expected loser on cost. That Goal can later cite line 6 below.
- **Combination rival:** the run adds one only if no single rival holds on both cuts.
- **Q6, no push and no PR:** this applies to the grill only. I'll commit the grill's writes locally and push nothing. The constraints block stays as written, so the run itself may push its branch and open one PR.

**What ak-context-engineering gives and doesn't give**
- It gives three rules I'll use:
  1. Cost comes from observed telemetry only. A missing value stays unknown and is never estimated.
  2. Original and candidate runs are matched by task and trial id.
  3. Failed and timed-out trials are recorded as rows.
- It does not give a median or a hang rule. Its aggregator reports averages, so those two rules are the Goal's own, as you said.

**The two cuts**
- Cut A is the record event of R-8JQPWG, `01M75T1YRF2R2Z99DX0E8JQPWG`. Four Results are recorded under the runner, so one is left.
- Cut B is the record event of R-JHBCWS, `01M76VXNN2ZW5WAXJ194JHBCWS`. Four Results are left.
- Each trial runs forward from its cut to the stop. No trial replays a whole build.

**4. Lines, with your two aims**
1. [N] On cut A or cut B, a trial stops with a candidate commit that has no review Result and is not named as unreviewed in the stop checkpoint.
2. [N] On either cut, a trial records more Results under its actor than the stop clause it was given allows.
3. [P] On either cut, the stop checkpoint and the last message give, for each candidate commit, its review Result or say it has none.
4. [P] On either cut, in every one of the 3 Sonnet trials, the final candidate has a review Result before the stop clause ends. This is your first aim, that a review must not easily run out of the budget.
5. [N] On either cut, a trial commits a candidate whose diff changes only tests or spec text for a case its own probe already passed on the previous candidate. This is your second aim, no endless review over small changes. It is run 2's `3cfd08d` and run 3's `d2d5d14` and `04142f3`.
6. [N] On either cut, a candidate's median trial cost is above 1.5 times the base's median at that cut. This is the cost limit; the numbers are below.

- **Check on line 4:** it is a positive "reviewed" line, so H-name-at-stop alone is contradicted. You agreed to the floor-only option earlier. Line 4 is your new aim, and I read it as overriding that. H-name-at-stop can still survive in a combination. Is that what you want?
- **Line 5:** a trial breaks it only when the commit's diff touches no non-test Go file and no skill, and a passing probe log for that case exists on the previous candidate. A commit that fixes a real defect is not a case of line 5. Is that the right edge?

**Cost for line 6, my proposal**
- **Unit:** tokens the Sonnet runner used from the cut to the stop. That is input, output, cache read and cache write, summed, as the eval harness observes them. I prefer tokens to dollars because a price table changes. Results are already bounded by line 2, and wall time depends on the machine. A trial with no observed tokens is invalid and is re-run, never estimated.
- **Median:** each rival and the base run 3 trials per cut. Sort the 3 costs and take the middle one. One hang therefore doesn't move the median. Two hangs make the median the greatest cost.
- **Hang:** a trial with no stop checkpoint after 30 minutes. Its cost is the greatest cost seen at that cut, across the base and every rival. A hang also fails lines 1 and 4 by itself.
- **Base:** the original rule texts: `skills/hyp-goal-run/SKILL.md` with sha256 `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e`, and the installed `hyp` (sha256 `c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c`) for the clause. The line names both and says the candidate's text differs from them. The base is a control run, not a Hypothesis. It costs one Experiment but adds no node.

**Budget:** `--experiments 8` is tight. The base is 1, the five rivals are 5, a combination is 1, and 1 is spare.

**Test route, to go in the Goal body**
- The test method is ak-skill-creator's eval flow. The model is Sonnet (`claude-sonnet-5-5`), with 3 trials per rival per cut.
- The selection therefore holds for Sonnet runners only. A rule that holds on Sonnet may not hold on the model that runs your `/goal`. The body will say so.

Answer these four:
- **A.** Does line 4 override the floor-only choice?
- **B.** Is the edge of line 5 right?
- **C.** Do you accept tokens, a median of 3, a 30-minute hang and the base above?
- **D.** Do you accept the Sonnet-only caveat in the body?

After that I send the routing table, then the independent review.

## Owner, message 3

A. Yes, line 4 overrides the floor-only choice, since it serves my first aim. H-name-at-stop can survive only in a combination.

B. The edge of line 5 is right. A commit that fixes a real defect is not covered by it.

C. I accept all of it:
- tokens as the unit
- a median of 3
- a 30-minute hang counting as the greatest cost at that cut
- the base as a control run, not a Hypothesis

D. I accept the Sonnet-only caveat in the body.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -rn -iE \"\\.tree\" internal/sandbox/*.go internal/cli/sandbox*.go 2>/dev/null | grep -v _test | grep -iE \"refus|exist|already|base|holds\" | cut -c1-230 | head -12; gh issue view 138 2>&1 | head -3; git show 5c5902d --stat | head -3; git ls-tree -r --name-only 5c5902d | grep -c \"\\.tree/\" "
}
```

Result:

```text
internal/sandbox/sandbox.go:84:// that does not exist, a base that already holds .tree, a sandbox with no
internal/sandbox/sandbox.go:702:		if _, err := git("-C", o.Host, "cat-file", "-e", o.Base+":.tree"); err == nil {
internal/sandbox/sandbox.go:703:			return nil, herr.BadArgs("%s already holds .tree at %s; a store cut needs a base without one", o.Host, o.Base)
internal/sandbox/sandbox.go:707:				return nil, herr.BadArgs("--store %s holds no .tree/events", o.Store)
To get started with GitHub CLI, please run:  gh auth login
Alternatively, populate the GH_TOKEN environment variable with a GitHub API authentication token.
commit 5c5902d50fe6d5de3d338e9d466e4f68dc02428d
Author: dat9uy <46886826+dat9uy@users.noreply.github.com>
Date:   Thu Nov 12 00:23:44 2026 +0700
4183
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in b4f23b2 5c5902d 3cfd08d; do git cat-file -e $c:.tree 2>&1 && echo \"$c has root .tree\" || echo \"$c no root .tree\"; done; git ls-tree 5c5902d .claude/skills/ | head -3; git ls-files -s .claude/skills/hyp-goal-run | head -2; git show 5c5902d:skills/hyp-goal-run/SKILL.md | shasum -a 256; git show main:skills/hyp-goal-run/SKILL.md 2>/dev/null | grep -c \"fresh subagent\"; git show 3cfd08d:skills/hyp-goal-run/SKILL.md | shasum -a 256; sed -n 80,100p internal/sandbox/sandbox.go | cut -c1-200"
}
```

Result:

```text
fatal: path '.tree' does not exist in 'b4f23b2'
b4f23b2 no root .tree
fatal: path '.tree' does not exist in '5c5902d'
5c5902d no root .tree
fatal: path '.tree' does not exist in '3cfd08d'
3cfd08d no root .tree
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  -
0
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  -
// repository root, base not a commit, a hyp that does not answer version,
// OUT_DIR exists, a bad keep path, no such sandbox, a missing --recipe, a
// recipe name given twice, none stacked with a recipe, a recipe that does
// not resolve, --store without --cut or the reverse, a store or cut event
// that does not exist, a base that already holds .tree, a sandbox with no
// Drive command, a remove with OUT_DIR of a directory with no record, a record directory that exists, an OUT_DIR without
// sandbox.json, an unknown Standalone name, a recipe with no Evidence
// command, an --inputs that is not a regular file, a --patch with a bad or
// repeated Arm, a patch file that is not a regular file or does not apply
// to the base); 3 the environment is wrong or I/O failed (a
// process still running in the sandbox, or one hyp could not check for,
// a sandbox another hyp command is changing,
// missing git or lsof, a failed create step, a failed or timed-out Launch or
// Doctor, a drive of a sandbox that is not ready or whose unfinished row's
// group lives, a Drive, Evidence or Standalone
// command that timed out, was interrupted or was killed, a corrupt record, a store line that
// does not parse). A Drive, Evidence or Standalone script's own non-zero exit
// is returned as *CommandFailed. Every check runs before anything is
// created or deleted.
package sandbox
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in 10566d4 b4f23b2 d2d5d14 04142f3; do echo $c $(git show $c:skills/hyp-goal-run/SKILL.md | shasum -a 256 | cut -c1-12); done; git branch -a --contains 04142f3 | head -3"
}
```

Result:

```text
(eval):1: bad substitution
10566d4 e3b0c44298fc
(eval):1: bad substitution
b4f23b2 e3b0c44298fc
(eval):1: bad substitution
d2d5d14 e3b0c44298fc
(eval):1: bad substitution
04142f3 e3b0c44298fc
* work
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in 10566d4 b4f23b2 5c5902d d2d5d14 04142f3; do git show $c:skills/hyp-goal-run/SKILL.md > \"$TMPDIR/s.md\" 2>&1; echo $c; shasum -a 256 \"$TMPDIR/s.md\"; done"
}
```

Result:

```text
Exit code 1
(eval):1: bad substitution
```

### Tool Bash (cwd: .)

```text
{
 "command": "for c in 10566d4 b4f23b2 5c5902d d2d5d14 04142f3; do echo $c $(git show ${c}:skills/hyp-goal-run/SKILL.md | shasum -a 256 | cut -c1-12); done"
}
```

Result:

```text
10566d4 e4ed9ec86eee
b4f23b2 a6152711c2c9
5c5902d cb15e2f2e207
d2d5d14 cb15e2f2e207
04142f3 cb15e2f2e207
```

## Grill

Preflight is done. I'm now sending the lines to an independent reviewer, as the skill requires before the routing table.

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader for grill lines",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer. Read-only: change no file, run no `hyp` write command, create no sandbox. Repository: /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt (a Go tool called hyp; read AGENTS.md first, then docs/glossary.md entries Goal, Line, Limit, Coverage, Subject, Store cut, and docs/replay-spec.md sections 5, 8, 8.1, 8.2; docs/sandbox-spec.md; internal/sandbox/sandbox.go around line 680-760 for `--store/--cut`).\n\nContext. A hyp store at investigations/goal-prompt/.tree holds two Goals (G-GZR6C6 design record, G-ZBM2RE build). Issue text: /Users/datguy/.hyp-sandbox/g146-pre-b-2/tmp/issue-141.md. A new Goal (a DESIGN Goal, discovery kind: its candidates are five rival rule texts for the skill skills/hyp-goal-run/SKILL.md and the stop clause printed by `hyp prompt`, tested as stand-in text; no commit is a candidate) is about to be written into that same store. Its single top-level DesignQuestion has five rival Hypotheses: H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes (defined in the issue). The test: a Sonnet runner is replayed from a store cut (two cuts: event 01M75T1YRF2R2Z99DX0E8JQPWG, \"cut A\"; event 01M76VXNN2ZW5WAXJ194JHBCWS, \"cut B\") forward to its stop, 3 trials per rival per cut, plus a control run of the original text (SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e, installed hyp sha256 c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c). Host base for cut A is commit b4f23b2, for cut B 3cfd08d. The stop clause counts Results recorded under the runner's actor (internal/cli/prompt.go).\n\nProposed Goal lines (each is owed on every top-level question that has a candidate):\n1 [N] On cut A or cut B, a trial stops with a candidate commit that has no review Result and is not named as unreviewed in the stop checkpoint.\n2 [N] On either cut, a trial records more Results under its actor than the stop clause it was given allows.\n3 [P] On either cut, the stop checkpoint and the last message give, for each candidate commit, its review Result or say it has none.\n4 [P] On either cut, in every one of 3 Sonnet trials, the final candidate has a review Result before the stop clause ends.\n5 [N] On either cut, a trial commits a candidate whose diff changes only tests or spec text for a case its own probe already passed on the previous candidate (a commit fixing a real defect is not covered).\n6 [N] On either cut, a candidate's median trial cost (tokens: input+output+cache read+cache write, from the cut to the stop; median of 3 = middle of sorted; a trial with no stop checkpoint after 30 minutes counts as the greatest cost seen at that cut) is above 1.5 times the control's median at that cut.\n\nTask: walk each line through scenarios the author did not draw, and report, with file:line evidence for every fact you rely on: (a) a rival or a trial
[... 1065 more characters]
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: aff08f4e4bce3db32 (internal ID - do not mention to user. Use SendMessage with to: 'aff08f4e4bce3db32', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-b-2-wt/a38586e6-d198-4306-b009-2696024a7a20/tasks/aff08f4e4bce3db32.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

## Grill

The independent reviewer is still working on the six lines, so I'm waiting for its report before I send the routing table.

## Owner, message 4

<task-notification>
<task-id>aff08f4e4bce3db32</task-id>
<tool-use-id>toolu_01XrxtRvtF9NcKHxrQbePyhz</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-pre-b-2-wt/a38586e6-d198-4306-b009-2696024a7a20/tasks/aff08f4e4bce3db32.output</output-file>
<status>completed</status>
<summary>Agent "Second reader for grill lines" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>The six lines are decidable only in part, and a trial can see events after the cut. I read the files and code and ran no hyp write command, sandbox or clone. I could not run a clone to confirm the leak in finding 1, so that one rests on git's documented behaviour plus the code.

## Findings, most serious first

**1. A trial can read the future through git history (c). It breaks lines 1, 3, 4, 5 and 6.**
- `build` runs `git clone --no-checkout o.Host wt`, then `checkout --detach Base`, then `remote remove origin` (`internal/sandbox/sandbox.go:1100-1102`).
- Git's clone makes a local branch for the host's HEAD. The host has one ref, `refs/heads/work` at 9c8e498 (`git for-each-ref`). Removing `origin` deletes only the remote-tracking refs, so the local branch `work` should remain in the sandbox.
- That branch holds the fix commits 5c5902d, 2284172, 3cfd08d, d2d5d14 and 04142f3, plus aef2d5b and 6f39ec0. Those two docs commits hold the later events and reviews: aef2d5b adds `E-TZTG0G/review.md` and `git log -- E-TZTG0G` returns it.
- A cut-A trial can therefore run `git log work` and cherry-pick the real fix. Its cost, its commits and its review results are then contaminated.
- The cut only filters the store (`internal/store/cut.go:100-120`). It does not filter git.

**2. The base commit carries a second, stale store (c, f). The trial can write to the wrong store.**
- b4f23b2 holds `investigations/goal-prompt/.tree` with only the grill and owner events and evidence of two Experiments. 3cfd08d holds the events of runs 1 and 2 but not run 3's (`git ls-tree`).
- `sandbox create` checks only for a root `.tree` at the base (`sandbox.go:702`), so it passes. `cut.Write` puts the cut at `wt/.tree`, so the clone has two stores.
- `FindRoot` takes the nearest `.tree` above the working directory (`internal/store/store.go:47-62`).
- A runner in `wt/` sees the cut. A runner that does `cd investigations/goal-prompt` sees the stale store, which lacks the cut's events, so cut B misses run 3's R-JHBCWS. Writes there go to the wrong store.
- `hyp prompt` prints the store path relative to the repository root (`internal/cli/prompt.go:152-165`). In the sandbox it prints `.`. Run from the host's `investigations/goal-prompt` it prints `investigations/goal-prompt`, which sends the runner into the stale store.
- On (f): `--store investigations/goal-prompt --cut ID` does work, because `CutAt` reads `&lt;store&gt;/.tree/events` (`cut.go:36`) and the error text accepts `&lt;host&gt;/REL`. But the prompt must be generated inside the sandbox's `wt/`, and the prompt text will differ from the control's original, which said `investigations/goal-prompt`. `hyp prompt` finds its Goal by id, so `G-ZBM2RE` resolves in the cut.

**3. The budget at the cut depends on the actor, which the lines do not name (a, c). It breaks lines 2, 4 and 6.**
- Cut A is event index 37 of the merged log, R-8JQPWG. Actor `claude-g113-1` has four Results by then (indices 28, 30, 35, 37), so one is left.
- Cut B is index 67, R-JHBCWS. `claude-g113-3` has one Result, so four are left. Cut B is not budget-bound at all.
- The stop clause counts Results under the actor (`prompt.go:164`). A new actor name gives a fresh 5 at both cuts and removes the issue's "one left" scenario. Reusing the old actor keeps it.
- Line 2's "the stop clause it was given allows" is ambiguous. It could count pre-cut Results or not. Reserve-last's trigger, "when one Experiment remains", never fires with a fresh actor until four Results are spent.
- Line 6 compares with the control at the same cut. A rival given a fresh or larger budget (raise-budget, 7) spends more tokens than a control with one Result left. It can fail line 6 for the budget alone, whatever it does with reviews.

**4. "Review Result" and "candidate commit" are undefined, so two readers differ (a). It breaks lines 1, 3 and 4.**
- hyp stores no review kind. The glossary's Verifier entry says only that "any actor may still review", and it lists "reviewer" under Avoid (`docs/glossary.md:154-156`).
- Run 3's E-GE085A was "Gate and review" in one Experiment. Under the strict reading that a review is its own Result, H-one-experiment fails lines 1, 3 and 4 by definition. Under a loose reading it passes. The line decides that rival before any trial.
- "Candidate commit" is a commit tied to a Hypothesis (`skills/hyp-goal-run/SKILL.md:29`, "each commit you test"). A runner that commits a last fix and ties no Hypothesis has no candidate, so lines 1, 3 and 4 do not apply. Run 1 had H-YXSJEM for 5c5902d; the line does not require one.
- "Candidate" means a supported Hypothesis of a question (`glossary.md:130`). Lines 1 to 5 use it for a commit inside a trial. Line 6 uses it for a rival. It has two meanings in one Goal.
- "Final candidate" is undefined: last by creation order, by branch tip, or by select.
- At cut A, b4f23b2 already has R-8JQPWG, and at cut B 3cfd08d has R-JHBCWS. A trial that commits nothing passes lines 1, 3 and 4 by doing nothing, whatever the open limits.

**5. Lines 1 to 4 count only trials that stop (a). They break lines 1, 3 and 4.**
- Only line 6 handles a trial with no stop. Under lines 1, 3 and 4, a trial killed at 30 minutes with an unreviewed commit has no "stop", so the line is silently satisfied.
- Line 4's "before the stop clause ends" is undefined. The clause is a disjunction of three conditions (`prompt.go:164`).
- Line 4 asks that every one of 3 trials pass, with no rule for a trial that errors or times out. Three trials of a stochastic Sonnet run give a coin-flip result at one cut. Run 2, with a fresh budget, still ended on an unreviewed commit (issue comment).
- A stop checkpoint's `summary` is at most 600 runes (`docs/replay-spec.md:62`). Line 3 asks it to list a review for each candidate commit, so it may not fit. It is unclear whether "each candidate" means the trial's own commits or every candidate in the Goal. At cut B that is at least seven.
- Line 3 is satisfied by saying a candidate "has none", even when it has one, and by naming a wrong Result id. Line 1 is a subset of line 3 (redundant, not a mixed invariant).

**6. Line 5 is decided by reading, conflicts with a build-Goal rule, and can be avoided (a, b). It breaks line 5.**
- "A real defect", "a case" and "only tests or spec text" need a judgment call on a diff.
- G-ZBM2RE#3 and the grill rule say a build line is checked by the candidate's own tests (`skills/hyp-owner-grill/SKILL.md`). Run 3's N1 limit said "no go test runs this case" (R-JHBCWS, `limits`). A tests-only commit that fixes that gap can be read as a real defect against the build line. Another reader calls it line 5's violation.
- A diff of a test plus a one-line code change escapes the "only" test.
- "Its own probe" needs a probe to exist. A rival text that tells the runner to skip probing and write the test passes line 5 while producing exactly the chain of small commits the owner wants to stop.
- Nothing bounds the number of reviews of one candidate. A rival such as "review again until the reviewer says none" passes lines 1 to 5. Only line 6 can catch it, and only against a control that is cheap at that cut.

**7. Line 6 is circular and is not named by sha256 (a, d). It breaks line 6.**
- A timed-out trial "counts as the greatest cost seen at that cut". That is a maximum over all rivals and the control. A rival's timeouts can set the reference, or can score lower than the trial's real tokens.
- The plan says "a control run", but line 6 uses "the control's median", which needs three runs. The line does not say which.
- Whether the reviewer subagent's tokens count is not stated. A rival that moves work into subagents looks cheap.
- The baseline is not named by sha256, which the grill skill requires for a baseline line (`skills/hyp-owner-grill/SKILL.md`, "A line carries its own cases").
- The control for cut A is SKILL.md `cb15e2f2...`, which first appears at 5c5902d. The cut-A base b4f23b2 holds `a6152711...`; 10566d4 holds `e4ed9ec8...`. Cut B's base, 3cfd08d, holds `cb15e2f2...`. So at cut A the "original text" is a later one than run 1 saw.
- The installed hyp c44fe92d is from HEAD. The `prompt.go` diff from b4f23b2 to HEAD changes no stop-clause text, so the clause is the same.

**8. Passing every line while defeating the aims (b).**
- *Skips the fix.* At cut A, b4f23b2 has eight open findings. A rival that commits nothing, records limits and stops passes lines 1 to 6. The final candidate is reviewed, the Results are few and the cost is low, but the findings are not fixed.
- *Hides the spend.* The stop clause counts Results under the runner's actor only (`prompt.go:164`). A rival that has the reviewer record under its own actor spends nothing against line 2.
- *Unreviewed last commit.* An untied final commit (finding 4).
- *Long review chains.* Repeated reviews of one candidate (finding 6).

**9. Process, history and reading words in the lines (d). Forbidden in a line.**
- The grill skill sends "how the tree is run (process, review, order, cadence)" to the body or a verdict rule, not a line (`skills/hyp-owner-grill/SKILL.md`, routing table).
- Process:
  - Lines 1 and 3 name the stop checkpoint and the last message.
  - Line 3's "last message" is chat. hyp stores no last message.
  - Line 5 names "its own probe".
- History: line 5 has "previous candidate" and "already passed". Line 4 has "before the stop clause ends". Line 6 has "from the cut to the stop".
- Test design: line 4 has "3 Sonnet trials". Line 6 has the control, the 30-minute rule and the median.
- Reading: lines 3 and 5 are decided only by reading a transcript or a diff. The word "reading" is absent.
- I found no owner act in any line.

**10. Smaller replay points (c).**
- The cut copies each kept Experiment's evidence folder whole (`cut.go:100-120`). It filters by Experiment id, not by time. I found no post-cut file in the E-folders of either cut. E-TZTG0G's `probe-3cfd08d.log` is cited in R-JHBCWS, which precedes the cut.
- Cut B contains run 2's stop checkpoint at index 65, which names the gap, and the owner answers at indices 45 and 47. The "not prompted by an earlier gap" check the owner asked for cannot be run at cut B. H-name-at-stop is prompted there, which weakens lines 1 and 3.
- The sandbox has no `plans/` and no `ak-skill-creator`, because `git ls-tree b4f23b2` shows neither. G-ZBM2RE's body names both under "Read only:", and its lines 6 and 7 use them. The original run had them. The trial may spend tokens or refer lines, which adds noise to line 6 equally for all rivals.
- The body lets the run "push that branch and open or update one PR". The sandbox has no `origin` (`sandbox.go:1102`), so a push fails and costs a step.
- Two actors merge by event id. I did not check whether the trial's clock sorts after the store's ids. I expect it does, as the events are dated 2026-11-12.

**11. Line length and invariants (e).**
- Rune counts of the line text, without the `[N]` or `[P]` prefix: 137, 101, 126, 116, 199, 298. None is over 600.
- No line mixes two invariants. Line 6 carries a definition but states one. Line 1 is a subset of line 3.

## Not checked
- I did not run a clone, so the local-branch leak in finding 1 is inferred.
- I did not check how the harness counts tokens or delivers a stand-in text, or whether skills load from the clone or from the installed copy.
- I did not check that a checkpoint over 600 runes is refused in code. I read only the spec at `docs/replay-spec.md:62`.</result>
<usage><subagent_tokens>134232</subagent_tokens><tool_uses>33</tool_uses><duration_ms>334445</duration_ms></usage>
</task-notification>

## Grill

The reviewer's report is a background event, not an answer from you, so it approves nothing. Before I rely on its findings I'm checking its main facts against the source.

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 1092,1106p internal/sandbox/sandbox.go; echo ---; git ls-tree -r --name-only b4f23b2 investigations/goal-prompt/.tree | head -5; git ls-tree -r --name-only b4f23b2 investigations/goal-prompt/.tree | wc -l; echo ---; sed -n 696,712p internal/sandbox/sandbox.go; echo ---; grep -n \"600\" docs/replay-spec.md | cut -c1-200 | head -6; sed -n 160,168p internal/cli/prompt.go | cut -c1-260"
}
```

Result:

```text
	return strings.TrimSpace(string(out)), nil
}

// build makes the sandbox's contents and record, writing cut, when there is
// one, to wt/.tree. For each recipe it also returns the recipe opened from
// its copy in <box>/recipes, whose commands point there, in stack order.
func build(o CreateOptions, box, version string, rcps []*recipe.Recipe, kas []*cmdrecord.KnownAnswers, cut *store.Cut) (*Record, []*recipe.Recipe, error) {
	wt, tmp, bin := filepath.Join(box, "wt"), filepath.Join(box, "tmp"), filepath.Join(box, "bin")
	for _, args := range [][]string{
		{"clone", "--quiet", "--no-checkout", o.Host, wt},
		{"-C", wt, "checkout", "--quiet", "--detach", o.Base},
		{"-C", wt, "remote", "remove", "origin"},
	} {
		if _, err := git(args...); err != nil {
			return nil, nil, err
---
investigations/goal-prompt/.tree/.gitignore
investigations/goal-prompt/.tree/events/claude-grill-113.jsonl
investigations/goal-prompt/.tree/events/owner-dat.jsonl
investigations/goal-prompt/.tree/evidence/E-T7ZS23/scenarios.md
investigations/goal-prompt/.tree/evidence/E-VJPMHG/scenarios.md
       5
---
		if fi, err := os.Stat(p.File); err != nil || !fi.Mode().IsRegular() {
			return nil, herr.BadArgs("--patch %s=%s: %s is not a regular file", p.Arm, p.File, p.File)
		}
	}
	var cut *store.Cut
	if o.Store != "" {
		if _, err := git("-C", o.Host, "cat-file", "-e", o.Base+":.tree"); err == nil {
			return nil, herr.BadArgs("%s already holds .tree at %s; a store cut needs a base without one", o.Host, o.Base)
		}
		if cut, err = store.CutAt(o.Store, o.Cut); err != nil {
			if errors.Is(err, store.ErrNoStore) {
				return nil, herr.BadArgs("--store %s holds no .tree/events", o.Store)
			}
			if errors.Is(err, store.ErrNoEvent) {
				return nil, herr.BadArgs("--cut: %v", err)
			}
			return nil, herr.Env("--cut: %v; nothing created", err)
---
49:| `node.add` | the new id (= event id) | `kind` (Goal, DesignQuestion, Hypothesis, Experiment), `parent` (id; `""` only for Goal), `title`, `body` (optional)
[... 1662 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd \"$TMPDIR\" && cat > lines.txt <<'EOF'\n[N] On cut A (event 01M75T1YRF2R2Z99DX0E8JQPWG) or cut B (event 01M76VXNN2ZW5WAXJ194JHBCWS), a trial stops with a candidate commit that has no review Result and is not named as unreviewed in its stop checkpoint, or records no stop checkpoint within 30 minutes.\n[N] On either cut, a trial records, under any actor after the cut, more Results than its stop clause's N less the Results its actor had recorded at the cut.\n[P] On either cut, the stop checkpoint names exactly the candidate commits that have no review Result, or says none lacks one, and what it says is true of the log.\n[P] On either cut, in every one of 3 Sonnet trials, the final candidate has a review Result before the trial's stop checkpoint.\n[P] On either cut, every limit open at the cut is closed or referred when the trial stops.\n[N] On either cut, a trial commits a candidate whose diff changes only test files or spec text for a case that a probe log cited by a Result on the previous candidate already shows passing; a commit that fixes a defect is not covered.\n[N] On either cut, a candidate's median trial cost is above 1.5 times the median of the base (hyp-goal-run SKILL.md sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e with hyp sha256 c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c, which the candidate's text differs from) at that cut; cost is all tokens of the trial from the cut to its stop, subagents included; a trial with no stop counts as the greatest cost of any finished trial at that cut.\nEOF\npython3 -I -c \"\nfor i,l in enumerate(open('lines.txt'),1): print(i,len(l.rstrip('\\n')))\""
}
```

Result:

```text
1 260
2 156
3 163
4 127
5 90
6 234
7 485
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-pre-b-2/wt
```

## Grill

I checked the reviewer's main claims against the source before relying on any of them.

**Verified**
- **Stale store at the base:** `b4f23b2` holds a second, older store at `investigations/goal-prompt/.tree` (5 files). `sandbox create` only refuses a store at the root (`sandbox.go:702`). A trial could `cd` into the old store and read or write the wrong events.
- **Summary limit:** a checkpoint summary is at most 600 runes (`replay-spec.md:62`). My line 3 asked it to carry a review for every candidate.
- **Budget count:** the stop clause counts Results under the actor (`prompt.go:164`). A fresh actor at cut A would get a fresh 5, which removes the "one left" scenario.
- **History leak:** `sandbox create` clones the host, checks out the base, and removes only `origin` (`sandbox.go:1100-1102`). The reviewer infers that the host's local branch `work`, with all the later fix commits, stays in the clone. I read the code but did not run a clone, so this is likely, not confirmed. I'll make the run check it.

**Revised lines**
My earlier line 6 named no baseline hash and left "review Result" undefined. The reviewer found both. It also found that a trial with no stop silently passed lines 1, 3 and 4. All seven lines are at most 485 runes.
1. [N] On cut A (event `01M75T1YRF2R2Z99DX0E8JQPWG`) or cut B (event `01M76VXNN2ZW5WAXJ194JHBCWS`), a trial stops with a candidate commit that has no review Result and is not named as unreviewed in its stop checkpoint, or records no stop checkpoint within 30 minutes.
2. [N] On either cut, a trial records, under any actor after the cut, more Results than its stop clause's N less the Results its actor had recorded at the cut.
3. [P] On either cut, the stop checkpoint names exactly the candidate commits that have no review Result, or says none lacks one, and what it says is true of the log.
4. [P] On either cut, in every one of 3 Sonnet trials, the final candidate has a review Result before the trial's stop checkpoint.
5. [P] On either cut, every limit open at the cut is closed or referred when the trial stops.
6. [N] On either cut, a trial commits a candidate whose diff changes only test files or spec text for a case that a probe log cited by a Result on the previous candidate already shows passing; a commit that fixes a defect is not covered.
7. [N] On either cut, a candidate's median trial cost is above 1.5 times the median of the base (hyp-goal-run SKILL.md sha256 `cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e` with hyp sha256 `c44fe92d0c93a4e2d4e1797df9dfd567f0233f67b00494c441acfa6881dc160c`, which the candidate's text differs from) at that cut. Cost is all tokens of the trial from the cut to its stop, subagents included. A trial with no stop counts as the greatest cost of any finished trial at that cut.

**What changed in the lines, and why**
- **Line 3:** the last message is chat and hyp doesn't store it. The line now judges the stop checkpoint only.
- **Line 5 is new.** Without it, a rule that commits nothing and records limits passes every other line. You didn't ask for this line, so please confirm it.
- **Line 7:** "most expensive" now means the greatest cost among finished trials, so a hang can't set its own reference. This is my reading of your rule. Subagent tokens now count, so a rule can't hide spend in a reviewer.
- **Line 6, the gap I'm leaving:** a rule that tells the runner to skip the probe and write the test passes it. Widening the line would also forbid tests a build Goal's lines demand. Your edge stands. A rule that reviews again and again is caught only by line 7.

**Terms, to go in the Goal body**
- **review Result:** a Result that cites a reviewer's findings on a named candidate commit's diff, whether or not it also gates that commit. The strict reading, where a review is its own Result, would decide H-one-experiment before any trial.
- **candidate commit:** a commit the trial adds after the cut's base, plus the base's own commit.
- **final candidate:** the branch tip at the stop.

**Routing table: please confirm or move each row**

| # | Question | Place | Separating scenario |
|---|---|---|---|
| 1 | Design Goal now, build later; part of #113 | Goal title and body | none |
| 2 | Five rivals under one question: what makes a run reach its stop with its last candidate reviewed or named, within budget? | one DesignQuestion with 5 rival Hypotheses (H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes) | cuts A and B, 3 trials each |
| 3 | A combination rival, only if no single rival holds | Goal body; the run adds it | none |
| 4 | H-whole-candidate | another Goal, which you add; it can cite line 7 | none |
| 5 | The floor and the review aim | lines 1 to 4 | |
| 6 | No limit dropped | line 5 [P] | |
| 7 | Small-change loop | line 6 [N] | |
| 8 | Cost cap | line 7 [N] | |
| 9 | How a review is recorded, and whether it counts against the budget | another Goal (the build) | none |
| 10 | Test method: eval flow, Sonnet, 3 trials, `--experiments 8`, Sonnet-only caveat | Goal body | none |
| 11 | Base text: `cb15e2f2` on both cuts. At cut A the host base `b4f23b2` holds an older text, `a6152711`. | Goal body | none |
| 12 | Instrument rules: no future commits in the clone, only the cut store, the actor of the cut, no `--branch` or `--pr` clause, `plans/` and `ak-skill-creator` absent | Goal body, then the verdict rule of each Experiment | none |
| 13 | The control must reproduce a failure of lines 1, 4, 5 or 6 on a cut, or the run says that cut does not reproduce the incident | verdict rule of the control Experiment | none |
| 14 | Cut A is the only cut with no earlier gap named, so it checks whether H-name-at-stop holds unprompted | DQ body | |
| 15 | Constraints block | Goal body (as approved) | none |
| 16 | Link to issue #141 | `hyp link … external_ref` (runner) | none |

**Open points**
- The Goal body is capped at 2000 runes, so rows 10 to 12 may not all fit. I'll draft the body and count it. Whatever doesn't fit goes into the Experiment bodies, which are capped at 600 runes each.
- Do you accept lines 1 to 7, the terms, and rows 11 to 14 as written? I write nothing until you confirm.
