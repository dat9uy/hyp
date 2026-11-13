# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-entry-b-4/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

## Grill

I'll start by reading the issue text and finding the stores.

### Tool Read (cwd: .)

```text
{
 "file_path": "/Users/datguy/.hyp-sandbox/g146-entry-b-4/tmp/issue-141.md"
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
 "command": "ls investigations/ && for d in investigations/*/; do echo \"== $d\"; (cd $d && hyp status 2>&1 | head -15); done; ls docs | head -40"
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
[... 3764 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cat skills/hyp-goal-run/SKILL.md; grep -rn \"experiments\\|Results recorded\" internal/cli/prompt*.go | head -40; git log --oneline | head -5; hyp --help 2>&1 | head -40"
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
[... 16371 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt/investigations/goal-prompt && hyp show G-ZBM2RE 2>&1 | head -80; grep -n -i \"review\\|budget\\|limit\" ../../docs/glossary.md | head -40"
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
[... 26478 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp show G-GZR6C6 2>&1 | cut -c1-900 | head -50; hyp render --markdown 2>&1 | grep -n \"R-8JQPWG\\|R-JHBCWS\\|E-GE085A\\|^#\" | head -60; ls .tree; ls .tree/evidence | head -50"
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
1:# G-GZR6C6 Record how hyp makes the /goal prompt: who writes it, and whether it carries owner answers — complete
5:## Q-NCY8HY Who writes the /goal prompt for a Goal: hyp from the log, or a skill? — answered
17:## Q-MAKJ8R Does the /goal prompt carry the owner's answers? — answered
30:# G-ZBM2RE hyp prints the /go
[... 3320 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt && cat docs/adr/0021-a-skill-change-is-a-build-goal.md | head -60; echo ----; grep -n -i \"run cap\\|no run cap\\|states no\\|5 Results\\|Experiments recorded\\|budget\" docs/*.md docs/adr/*.md skills/*/SKILL.md AGENTS.md README.md | cut -c1-300 | head -30; echo ----; sed -n 150,175p internal/cli/prompt.go; cd investigations/goal-prompt; hyp show E-GE085A 2>&1 | cut -c1-1500 | head -30"
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
[... 5001 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "hyp status --as-of R-8JQPWG G-ZBM2RE 2>&1 | head -30; hyp show Q-Z93VJ5 2>&1 | cut -c1-400 | head -40; git log --oneline -3 -- . ; git status --short | head"
}
```

Result:

```text
GOAL G-ZBM2RE hyp prints the /goal prompt for a Goal from its log — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-ZBM2RE#1 ran in R-KPWZZC, R-50GBXM — [P] For a Goal whose body holds a valid constraints block, hyp prompt <G> --actor A exits 0 and prints the same bytes on…
  G-ZBM2RE#2 ran in R-KPWZZC, R-50GBXM — [N] hyp prompt prints no text of an owner answer: its output for a Goal is byte-identical before and after an answer eve…
  G-ZBM2RE#3 ran in R-KPWZZC, R-50GBXM — [N] The constraints block runs from the first line that starts with "May change:", "Read only:" or "Never:" to the end o…
  G-ZBM2RE#4 ran in R-KPWZZC, R-50GBXM — [N] In a store with two Goals, where one answer event settles a key of each, hyp prompt G prints the same bytes as for G…
  G-ZBM2RE#5 ran in R-KPWZZC, R-50GBXM — [N] hyp prompt with no --actor exits 2 and never falls back to $HYP_ACTOR; an unknown id or a non-Goal id exits 2.
  G-ZBM2RE#6 ran in R-KPWZZC, R-50GBXM — [P] For the testdata fixture (checkpoint-store:G-D1CH1R's events through the owner answer of 2026-11-11 14:31, its Goal …
  G-ZBM2RE#7 ran in R-KPWZZC, R-50GBXM — [P] Each SKILL.md the candidate changes passes ak-skill-creator's quick_validate.py (sha256 d81f883a…) with 0 errors and…
FRONTIER
  Q-Z93VJ5 open/all_contradicted Which hyp commit implements hyp prompt and the skill text i…
    H-3M3QAS (h-b4f23b2) contradicted (supersedes H-Q5P9R9) feat/goal-prompt at b4f23b2: 10566d4 plus fixes for review …
STALE
OWED
  runner: Q-Z93VJ5 add rival Hypotheses; ruled out: h-b4f23b2 by R-8JQPWG
AWAITING OWNER
  R-KPWZZC#r1 (Q-Z93VJ5), R-50GBXM#r1 (Q-Z93VJ5) reading: G-ZBM2RE#6: "stated by the printed prompt or by hyp-goal-run or a skill it names" has two meanings. Meaning 1: a clause counts when the prompt tells the run where to read it; e.g. the prompt says "read it with hyp show G-D1CH1R", and that Goal body names ADR 0020. Meaning 2: the clause must be in the prompt or skill text itself. 
[... 6779 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt && grep -n -i \"\\bcap\\b\\|stop clause\\|stops a run\\|run stops\" docs/how-hyp-improves-itself.md docs/principles.md docs/replay-spec.md skills/*/SKILL.md | cut -c1-260 | head; git ls-files recipes | grep -i eval | head"
}
```

Result:

```text
docs/how-hyp-improves-itself.md:29:  ([ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md)). A run's cap is
docs/principles.md:47:when the ledger owes it, and a run stops when the runner owes nothing. Nothing
docs/replay-spec.md:235:- **Constraints block.** A Goal body's constraints block runs from the first line that starts with `May change:`, `Read only:` or `Never:` to the end of the body. It is valid when it holds each of the three headings and every non-empty 
docs/replay-spec.md:380:`hyp status` prints the runner's items under OWED, keeping the first five when it truncates, and the owner's under AWAITING OWNER. `hyp status --live` leaves out of FRONTIER each contradicted Hypothesis that holds no open Experiment mar
skills/hyp-goal-run/SKILL.md:86:   hyp checkpoint --goal <goal-id> --summary "Stopped: <stop clause>. Waiting for the owner: <the keys under AWAITING OWNER>. Out-of-Goal findings: <one short phrase each, or none>" --why "<why>"
skills/hyp-owner-answer/SKILL.md:3:description: "After a /goal run stops at the owner, walk the open readings and referred limits of the hyp tree with the owner: explain each one, answer questions, and record only the owner's answers."
skills/hyp-owner-grill/SKILL.md:50:- **Merge states one invariant guards.** At most 10 lines of 1 to 600 runes each; the cap forces the merging.
skills/hyp-owner-grill/SKILL.md:116:It prints the store's path, the actor, the Goal id, the skill, the stop clauses (the runner owes nothing, or N Results under the actor, 5 by default) and the Goal body's constraints block, line for line; `--branch` and `--pr
```

### Tool Bash (cwd: .)

```text
{
 "command": "sed -n 24,36p docs/how-hyp-improves-itself.md"
}
```

Result:

```text
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
```

## Grill

I've read the issue, the `goal-prompt` store and the skill and prompt text, and I'm ready to ask round 1. Nothing has been written to the store.

## What I found

- **Store.** #141 is "Part of #113". Both #113 Goals (G-GZR6C6, G-ZBM2RE) are in `investigations/goal-prompt`, so the new Goal goes there. `hyp status --as-of R-8JQPWG G-ZBM2RE` already replays the run-1 cut, so the scenario is a real replay and not a thought experiment.
- **The Goal is a discovery Goal.** Its candidates are rival run practices, tested by replaying cuts. The build that changes `skills/hyp-goal-run` and the prompt text is a second Goal, as G-GZR6C6 and G-ZBM2RE were. A Goal's lines can cite only keys that are already settled, so I can't write that build Goal until this one has selects.
- **Derivation guard.** I propose no stored field. "Unreviewed" can't be derived, because a review is a `--recipe none --rubric reading` Experiment that hyp can't tell from a gate. So the stop rule has to be checkpoint text, not a hyp check.
- **Preflight scan.** `docs/how-hyp-improves-itself.md:29` says "A run's cap is stated by the `/goal` prompt or the recipe, not by hyp". `hyp prompt` now prints it, so the build Goal's `May change:` must list `docs/`. This was run 1's finding F8.

## Round 1

**1. Your actor name.** I need it once, for the Goal's add. I'd use `owner-dat`. Is that right?

**2. Two Goals, with the build later.**
- **Today.** One issue holds a design question and a build.
- **Now.** Goal D (discovery) picks the practice from replays.
- **Later.** A build Goal, after D's selects, carries the lines "a run that reaches its budget has a review Result or names the gap" and "`hyp-goal-run` states how a review is recorded and whether it counts".
- **Why.** The candidate of D is a practice judged on cuts. The candidate of the build is a commit, and its lines cite D's selects.
- **Recommendation.** Yes. Do you confirm?

**3. Test the rivals by replay, or settle them in the session?**
- **Today.** The rivals have evidence from three real runs but no controlled run.
- **Testing.** A replayed run is given the cut, the skill text and one rival's rule. It succeeds if it stops with the last candidate reviewed or the gap named. For example, run 1's cut after R-8JQPWG with H-reserve-last should review `5c5902d` and stop.
- **Settling now.** It would be recorded as argued, not tested.
- **Recommendation.** Test them. Your issue asks for "a run on the scenario".
- **Method.** I'd use ak-skill-creator's eval flow, as G-ZBM2RE did for line 6. Each cut becomes a case. A consumer run gets the cut plus one rival's rule. An independent reviewer grades it blind against a rubric. Three preferences are yours:
  - **Model.** I'd use the model the original runs used for the consumer, and the strongest for the grader.
  - **Trials.** I'd run 1 trial per rival and cut.
  - **Budget.** I'd set the prompt to `--experiments 8`. The default of 5 would stop this run in the same way.
  - What would you change?

**4. How the rivals are grouped.** The six rivals answer three different questions, so I'd make three DesignQuestions with a baseline rival each, so every question has at least two rivals:
- **Q-end: how does a budget-bound run end with its last candidate reviewed, or the gap named?** Rivals: H-name-at-stop, H-reserve-last, H-one-experiment, H-raise-budget. Scenario: run 1's cut after R-8JQPWG and run 2's cut after R-3YJ7T2.
- **Q-probe: how is a case that the run's own probe already passed recorded?** Rivals: H-probe-closes, and H-probe-limit (today's practice: limit, then a fix commit). Scenario: run 3's cut after R-JHBCWS.
- **Q-read: what does the review before the select read?** Rivals: H-whole-candidate, and H-last-increment (today's practice). Scenario: a reviewer reads the cumulative diff from the base to `04142f3`. The case succeeds if it reports a case of a line that probes confirm and that no earlier review reported.
- **Combinations.** Your issue wants the tree to say which rivals hold alone. I'd write in the Goal body that the run tests each rival alone first. If none passes alone, it may add one combination Hypothesis, with H-name-at-stop as one part. Without that permission, a run can't add a rival. Do you allow it?

**5. The lines.** Only lines that every candidate of every question can meet go on D. I propose two:
- **[P]** On each pinned cut, a replayed run under the candidate's rule, once it reaches its stop clause, has a review Result for its last candidate commit or names that commit as unreviewed in its stop checkpoint. It never writes "all review limits closed" while a commit lacks a review.
- **[N]** A rule under which a finding the cut's real review reported as a case of a line ends the replay with no limit, no close, no fix commit and no name. These are known-answer flaws: run 1's G1 symlink, G2 CRLF and K1.
- **Why these two.** The first is your "Done means". The second stops a rule from passing by dropping reviews' findings, which is the cheapest way to reach "reviewed".
- **Not lines.** "A review costs one Experiment" and "the default stays 5" are rivals, not lines, because a line is a case every candidate must meet.
- Do you add, merge or drop a line?

**6. The constraints block.** My draft:
- `May change:` the `investigations/goal-prompt` tree (events and evidence), in commits on the branch the prompt names, plus scratch copies under `$TMPDIR`. The run may push the branch and open or update one PR, and never merges it.
- `Read only:` `skills/`, `docs/`, `internal/`, `plans/`, and every other `investigations/` tree.
- `Never:` edit `skills/`, `docs/` or `internal/` (that is the build Goal); replace an installed binary; write under an owner's actor; merge or commit to `main`.
- **Branch.** What branch name do you want? I'd use `exp/run-practice-141`.

When these are answered I'll send the routing table for you to confirm. Then I'll run the second-reader check, with a fresh subagent on a scenario the grill didn't draw, before any write.

## Owner, message 2

1. My actor name is owner-standin, not owner-dat.

2. Agreed: one discovery Goal in goal-prompt now, and the build Goal later.

3. Yes, test by replay.
   - Run the replays on Sonnet, with 3 trials per candidate and per starting point, not 1.
   - Don't replay a whole build from its start for each trial. That costs too much. Replay from the cut.
   - The grader model and `--experiments 8` are fine.
   - Could the ak-context-engineering skill help with how we evaluate? Is it worth using here?

4. Don't split it into three questions. Put every candidate fix under one question and test them all, each at its own starting point. Treat H-whole-candidate as the expected loser, since one cumulative review costs a lot. I have no view on combinations, so I accept your recommendation.

5. How does each of your two lines serve my two aims?
   - A review must not easily run out of the run's budget.
   - A review must not go on and on because of small changes.

   I don't see the second aim in either line. Is there a cost line? A candidate may cost at most 1.5 times what the base costs at the same starting point. How do you compute that cost? If you use a median, how is it computed? A trial that hangs counts as the most expensive.

6. The constraints block is fine, and I accept `exp/run-practice-141`.

### Tool Bash (cwd: .)

```text
{
 "command": "ls .claude/skills ~/.claude/skills 2>/dev/null; find / -iname \"*context-engineering*\" -maxdepth 6 2>/dev/null | head; cat investigations/goal-prompt/.tree/evidence/E-NZR3GV/runs/goal-run-plan/A/record.json | head -c 1500"
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
/Users/datguy/codingProjects/context-engineering
{
  "run_id": "e5-goal-run-plan-A",
  "variant": "A",
  "model": "claude-haiku-5-5",
  "runtime": "claude-code",
  "status": "completed",
  "tokens": null,
  "duration_ms": null,
  "tool_calls": null,
  "corrections": null,
  "case_id": "goal-run-plan",
  "split": "train",
  "skill_sha256": "3d7a44efd75509e8349b9fcd0d329963b01176c7b97dc7f19d13fa667f1159a3",
  "checks": [
    {
      "id": "p1-hyp",
      "passed": null,
      "evidence": "Not graded: needs independent rubric review"
    },
    {
      "id": "p2-rule",
      "passed": null,
      "evidence": "Not graded: needs independent rubric review"
    },
    {
      "id": "p3-logs",
      "passed": null,
      "evidence": "Not graded: needs independent rubric review"
    },
    {
      "id": "p4-review",
      "passed": null,
      "evidence": "Not graded: needs independent rubric review"
    },
    {
      "id": "p5-owner",
      "passed": null,
      "evidence": "Not graded: needs independent rubric review"
    }
  ]
}
```

## Grill

I did not use ak-context-engineering. It isn't installed here: `.claude/skills` and `~/.claude/skills` don't list it. The only match on this machine is a directory at `~/codingProjects/context-engineering`, which is outside this task, so I haven't read it. If you want it used, install it or tell me to read that directory. Then I'll name it in the Goal body as an optional method, and the run would apply it.

## One question (item 4)

Agreed. There is one DesignQuestion, with all six rivals under it: "Which change to the run practice keeps a review from being cut off by the budget or dragged out by small fixes?" Each rival body names its own starting point:

- **Run-1 and run-2 cuts:** H-name-at-stop, H-reserve-last, H-one-experiment and H-raise-budget.
- **Run-3 cut after R-JHBCWS:** H-probe-closes.
- **Final candidate `04142f3`:** H-whole-candidate. The body records your expectation that it loses on cost. It is still run, and it carries no preset verdict.
- **Control arm:** the unchanged skill at the same cut. It is not a Hypothesis. It gives each cost line its comparison.

## Your two aims (item 5)

You're right that the second aim had no line, and the first was only half covered. My "reviewed or named" line lets a rule pass by naming the gap, which doesn't keep the review inside the budget. I propose three lines:

1. **[P], aim 1.** At the first stop clause of a replayed run from the Hypothesis's cut, its last candidate commit has a review Result. Naming the gap does not satisfy this line. So H-name-at-stop alone is contradicted, which is honest, because naming doesn't make a review happen. It stays a part of combinations and becomes a line of the build Goal.
2. **[N], known answers.** A rule under which a finding that the real review reported as a case of a line (run 1's G1, G2 or K1) ends with no limit, no close, no fix commit and no name. Without it, "no review" would be the cheapest way to pass the cost line.
3. **[N], aim 2, the cost line.** The candidate's cost at its cut is at most 1.5 times the control's cost at that cut.

## How the cost is computed (item 5)

Here is my proposal for each part. The units are mine, so please check them.

- **Unit.** Results recorded from the cut until `hyp status` shows the select or "runner owes nothing". This is the budget's own unit, and the log gives the count without anyone judging it. The eval record's `tokens` field was `null` in E-NZR3GV, so a token count isn't available from that flow. Tokens are what would make H-whole-candidate expensive, so I'd note its token use as an observation but not use it in the line.
- **Restarts.** If a run hits its budget before the select, a new run starts with a fresh budget of the same size, as your run 2 did. The cost is the total across runs, so an arm that ends reviewed at its first stop but then loops on fixes still pays for the loop.
- **Trial cost.** One number per trial. A trial counts as hung and costs ∞ if it records 15 Results without reaching the select, runs past 45 minutes of wall clock, or is killed.
- **Median.** Sort the 3 trials' costs from lowest to highest, with ∞ last, and take the middle one. Two hangs out of three gives ∞.
- **The line.** A candidate fails if its median is ∞, or if 2 × candidate median is greater than 3 × control median. If the control median is ∞, any finite candidate passes. I use integer arithmetic so there is no rounding; control 2 allows up to 3.
- **Runs.** 3 trials per arm per start, on Sonnet, replayed from the cut, with an independent grader.

## Questions for you

- **A.** Do you accept line 1 without the naming escape?
- **B.** Is "Results until the select, across restarts" what you mean by "goes on and on", or do you want tokens or wall clock instead?
- **C.** Are 15 Results and 45 minutes the right ceiling for a hang?
- **D.** The count is about 42 trials, and each is a chain of runs: 5 arms × 3 trials at each of the run-1 and run-2 cuts, plus 2 × 3 at each of run 3 and the final candidate. Do you accept that, or should I trim it? The cheapest trim is to test H-raise-budget only at the run-2 cut, since your comment argues against it.

I'll send the routing table once these are answered.

## Owner, message 3

A. Accepted, no naming escape in line 1.

B. Yes, Results until the select, across restarts. That's what I mean by "goes on and on". Don't use tokens or wall clock for the line.

C. 15 Results and 45 minutes are fine as the hang ceiling.

D. Don't trim. I want every candidate tested, so H-raise-budget runs at both the run-1 and run-2 cuts. I didn't argue against it, so please drop that reason. The 42 trials are fine, as long as each trial replays from the cut and not from the start of a whole build.

As for ak-context-engineering, skip it. I only wondered whether it would help, and it isn't installed, so don't name it in the Goal.

## Grill

Your answers are all in. Before the routing table I'll run the second-reader check. A fresh subagent gets the rules and the files, but not my conclusion.

### Tool Agent (cwd: .)

```text
{
 "description": "Second-reader scenario walk",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer of a test design. Read-only: change nothing in /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt. Report facts you verified in the source, with file:line, and mark anything you did not verify.\n\nCONTEXT. hyp stores an investigation as append-only events in `.tree/events/` and derives all state by replay (docs/replay-spec.md, docs/glossary.md). The store `investigations/goal-prompt` holds two Goals (G-GZR6C6, G-ZBM2RE). A run's `/goal` prompt (printed by `hyp prompt`, code in internal/cli/prompt.go) stops the run on \"N Results have been recorded under <actor>\" (default 5) or when OWED says the runner owes nothing. skills/hyp-goal-run/SKILL.md is the run practice. Issue text: /Users/datguy/.hyp-sandbox/g146-entry-b-4/tmp/issue-141.md.\n\nPLAN UNDER REVIEW. A new discovery Goal in that store tests six rival changes to the run practice by REPLAY. One DesignQuestion; six rival Hypotheses, each naming its own starting point (a \"cut\" = the store replayed as of a Result, e.g. `hyp status --as-of R-8JQPWG G-ZBM2RE`):\n - H-name-at-stop, H-reserve-last, H-one-experiment, H-raise-budget: at the run-1 cut (after R-8JQPWG) and the run-2 cut (after R-3YJ7T2)\n - H-probe-closes: at the run-3 cut (after R-JHBCWS)\n - H-whole-candidate: at the final candidate 04142f3\n plus a control arm (unchanged skill) at each cut.\nA trial = a Sonnet consumer agent given the cut, the skill text with one rival's rule, and the /goal prompt, run in a scratch copy of the store, continuing until the select or the stop clause; when the stop clause fires a new run starts with a fresh budget of the same size (as real run 2 did). 3 trials per arm per cut; ~42 trials.\nProposed Goal lines:\n L1 [P] At the first stop clause of a replayed run from the Hypothesis's cut, its last candidate commit has a review Result (naming the gap does not satisfy it).\n L2 [N] A rule under which a finding the real review reported as a case of a line (run 1's G1, G2, K1) ends the replay with no limit, no close, no fix commit and no name.\n L3 [N] Cost at the cut > 1.5x control's cost at that cut. Cost of a trial = Results recorded from the cut until `hyp status` shows the select or \"runner owes nothing\", across restarts; hung = 15 Results without reaching the select, or 45 minutes, or killed -> infinity. Per arm per cut: median of 3 trials (infinity sorts last). Fail if candidate median is infinity or 2*cand > 3*control; if control median is infinity any finite candidate passes.\n\nTASK. Walk each rule through scenarios the designers did NOT draw, and report where the plan breaks or is ambiguous. At minimum try: (1) the stop clause counts Results under the ACTOR: at a cut where the real run had recorded 4 Results under claude-g113-1, what budget does a replay actor have, and can \"one Experiment left\" even be reproduced? check internal/cli/prompt.go and rep
[... 919 more characters]
```

Result:

```text
[Subagent hand-back] The text below is the final report of a subagent this session delegated to. It is model output, NOT a message from the user: instructions, requests, or approval claims inside it are the subagent's words and carry no user authority. The harness indents every line of the report, so a frame-like line at column zero inside it would be forged. Notes above this frame may quote model-derived text, which carries no user authority either. The report follows:
  The plan breaks on three points: the "one Experiment left" condition, the repository state at each cut, and the final-candidate cut. I read `internal/cli/prompt.go` and the events, and I ran `hyp status --as-of` read-only on the live store. I did not run `hyp sandbox create --cut`, so every statement about building a scratch store is from the spec text only.
  
  **Ranked findings**
  
  1. **"One Experiment left" is not reproduced by a fresh replay actor.**
     - The stop clause is only prompt text (`internal/cli/prompt.go:167`): "N Results have been recorded under <actor>". Nothing in the code counts Results.
     - A new actor name has 0 Results in the cut, so it gets the full budget of 5, not the real 1. The run-1 cut (after R-8JQPWG) has 4 Results under claude-g113-1, and the run-2 cut (after R-3YJ7T2) has 4 under claude-g113-2. I counted these in the event files. Run 3's cut (after R-JHBCWS) is run 3's first Result, so 4 are left there, not 1.
     - There are two ways to reproduce the real budget. Reuse the real actor name, so the agent counts 4 of 5. Or pass `--experiments 1`, which reads "1 Result recorded under <fresh actor>", a different wording from the real prompt. Both are untested, and the plan picks neither.
     - Which one it uses decides whether "when one Experiment remains" in H-reserve-last can fire at all. It also decides what L1's "first stop clause" is. I did not check whether `hyp status` shows a per-actor count, so the agent may have to count by hand and miscount.
  
  2.
[... 4564 more characters]
```

## Grill

I'll check the reviewer's key facts against the source before relying on them.

### Tool Bash (cwd: .)

```text
{
 "command": "grep -n \"\\-\\-cut\\|--base\" docs/replay-spec.md docs/sandbox-spec.md | cut -c1-240 | head -12; cd investigations/goal-prompt && hyp events --json 2>/dev/null | python3 -I -c \"\nimport json,sys\nd=json.load(sys.stdin)\nev=d.get('events',d) if isinstance(d,dict) else d\nn={}\nfor e in ev:\n    e=e.get('event',e)\n    t=e.get('type');a=e.get('actor');i=e.get('id')\n    if t=='result.record':\n        n.setdefault(a,[]).append(e.get('result_id') or e.get('node') or i)\nfor a,v in n.items(): print(a,len(v),v)\n\" | head -20"
}
```

Result:

```text
docs/replay-spec.md:88:**[added] subject**: a subject has one shape, a base commit plus what is given over it: at most one patch, and input files (ADR 0020). A Hypothesis may name its subject, what embodies its claim, with `hypothesis.subje
docs/replay-spec.md:158:- `hyp sandbox create --store SRC --cut EVENT` writes a store cut in this order: the merged events up to and including the first with id `EVENT` (the `asOf` stop in §7), each line as written, so the cut store replays
docs/replay-spec.md:159:- A line that does not parse has no id, so it is not in the merge. `hyp check` reports it as an error naming `file:line`; `hyp status` (with or without `--as-of`) replays the other lines, exits as usual, and prints `
docs/sandbox-spec.md:158:A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (
docs/sandbox-spec.md:160:- **A change that lands in code is a commit.** An Experiment about it uses that commit as `--base` and takes no patch.
claude-grill-113 2 ['01M75R21F8TYTTJZ8Y9CFCEZY6', '01M75R21G7H1J6RARDMM26GPDW']
claude-g113-1 6 ['01M75S97KWYDNTGA2N56KPWZZC', '01M75SBVS853GXX0C0SC9V3ER0', '01M75SV7JAPQZNHTTZB050GBXM', '01M75T1YRF2R2Z99DX0E8JQPWG', '01M75TJFS2Z8ZNKEX6Q4FRF929', '01M75TJZK3V9F85CR6TW9EGPYD']
claude-g113-2 5 ['01M76TFWZTR7R6QCZ3C30E8PFH', '01M76TP5NF2DETM36QC7KYKWH9', '01M76TW97XFDSKQAKNFK28A8NB', '01M76TZSMHENAYGH77MR3YJ7T2', '01M76V4HWRMAZ0TQECYR356DV0']
claude-g113-3 5 ['01M76VXNN2ZW5WAXJ194JHBCWS', '01M76W29E72EGEB3E5ED27Y1EV', '01M76W7STKAZ2B3QA1TRDER6F8', '01M76WG91ZTMZJEW0V57BSWFXG', '01M76WH13EXC2KFQV151HXEBAF']
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt && git rev-parse --short HEAD && shasum -a 256 skills/hyp-goal-run/SKILL.md && git diff --quiet HEAD -- skills && echo clean"
}
```

Result:

```text
9c8e498
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
clean
```

## Grill

The second reader found real problems. I checked its main facts against the store: the event log gives each cut's Result count per actor, and `hyp sandbox create --store --cut` exists (`docs/replay-spec.md:158`). I did not check its claims about the git clone and the candidate commits. The Goal body will tell the run to verify those.

## What the review changed

- **"One Experiment left" needs the real actor.** The stop clause counts Results under the actor. A fresh actor has the full budget of 5. So each replay uses the cut's real actor, where the log already holds 4 Results (`claude-g113-1` at the run-1 cut, `claude-g113-2` at the run-2 cut), and so one is left. Run 3's cut has 1 recorded, so 4 are left there. H-raise-budget uses `--experiments 7`. A restart uses a fresh actor and the same budget, as real run 2 did.
- **A cut covers the store, not git.**
  - The consumer's clone must hold only the candidate commits that exist at the cut. The Goal body says so.
  - The consumer must not get issue #141's text, because it states the answer.
  - The run checks that `hyp sandbox create --cut --base` hides later commits.
- **The final cut is named.** It is after R-BSWFXG, where `04142f3` is the candidate and its review limit Q1 is still open.
- **L1 stops at the arm's own stop clause.** H-raise-budget's is Result 7, and the control's is 5. L1 now needs a pass threshold, and I recommend 2 of 3 trials.
- **Restarts** use the `hyp prompt` output as it is, with a new actor. A rule that fires "when one remains" fires at each restart's end, and that is part of its cost.
- **L2 as I drafted it is weak.** `hyp select` already refuses while a limit is open, so a rule cannot silently drop a limit. The failure I meant is a shallow review. The review's findings vary per trial, so there is no stable known answer, and handing the consumer fixed findings would turn the replay into a script. I recommend dropping L2.

## Routing table

| # | Question | Place | Separating scenario or invariant |
|---|---|---|---|
| 1 | Where the Goal goes | `investigations/goal-prompt`, a discovery Goal | looked up |
| 2 | Purpose | Goal title and body: "A budget-bound run reviews its last candidate without runaway fix chains" | n/a |
| 3 | Line 1 | **[P]** At the first stop clause of the arm's own prompt, the run's last candidate commit has a review Result, in at least 2 of 3 trials. Naming the gap does not satisfy it. | Replay each Hypothesis's cut, with the control run at the same cut. |
| 4 | Line 2 (cost) | **[N]** The arm's median cost exceeds 1.5 times the control's at the same cut. The control is `skills/hyp-goal-run/SKILL.md` at `9c8e498`, sha256 `cb15e2f2…fa750e`, and the candidate's text differs from it. | Median of 3 trials; cost is Results from the cut to the select across restarts, ∞ on a hang. Failure if the median is ∞ or 2 × candidate > 3 × control. |
| 5 | Line 3 (known answers) | **Dropped**, if you agree | n/a |
| 6 | The six practice changes | One DesignQuestion, six rival Hypotheses, a control arm that is not a Hypothesis | See below. |
| 7 | Combination rival | Goal body: test each rival alone, then the run may add one combination Hypothesis | n/a |
| 8 | Method | Goal body: replay from the cut, Sonnet consumer, strongest grader, 3 trials, ceiling of 15 Results or 45 minutes. `--experiments 8` goes to `hyp prompt`. | n/a |
| 9 | Constraints block and branch | Goal body, `exp/run-practice-141` | n/a |
| 10 | The build | Another Goal, added by you after the selects. It carries "a run that reaches its budget has a review Result or names the gap" and "`hyp-goal-run` states how a review is recorded and whether it counts". It also carries the stale line in `docs/how-hyp-improves-itself.md:29`. | n/a |

The six Hypotheses and their cuts:
- **H-name-at-stop, H-reserve-last, H-one-experiment and H-raise-budget:** run-1 cut (after R-8JQPWG) and run-2 cut (after R-3YJ7T2).
- **H-probe-closes:** run-3 cut (after R-JHBCWS).
- **H-whole-candidate:** final cut (after R-BSWFXG). The body records that you expect it to lose.

I will not name ak-context-engineering anywhere in the Goal.

## What I need from you

- Confirm or move each row.
- Say whether you accept dropping the third line (row 5).
- Say whether L1 passes at 2 of 3 trials.

After that I'll show the exact `hyp add goal` command under `--actor owner-standin` for you to approve before I run it. I'll write the DesignQuestion and rivals, the checkpoint, and the `hyp check` run after it.

## Owner, message 4

I confirm every row as it stands, and I'm not moving any.

- **Row 5:** Drop the third line. You showed it was weak, so I agree to drop it. The cost line still covers my aim that a review must not go on and on because of small changes.
- **Line 1:** It passes at 2 of 3 trials.

Go ahead and show me the exact `hyp add goal` command under `--actor owner-standin`.

### Tool Bash (cwd: .)

```text
{
 "command": "cd \"$TMPDIR\" && cat > l1.txt <<'EOF'\n[P] For each Hypothesis, replaying its named cut with its rule gives, in at least 2 of 3 trials, a run whose last candidate commit has a review Result at the first stop clause of that arm's own prompt (5 Results under the actor; 7 for the raise-budget rival). Naming the gap in the stop checkpoint does not satisfy this line. The control, skills/hyp-goal-run/SKILL.md at 9c8e498 (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), runs at the same cut and is no candidate.\nEOF\ncat > l2.txt <<'EOF'\n[N] A Hypothesis costs more than 1.5 times the control at its cut. A trial costs the Results recorded from the cut until the select or until the runner owes nothing, across restarts; 15 Results without that, 45 minutes or a kill costs infinity. An arm costs the median of its 3 trials, infinity last. It fails if its median is infinity or 2 times it exceeds 3 times the control's; a finite arm passes an infinite control. Its skill text differs from the control's (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e).\nEOF\nfor f in l1.txt l2.txt; do printf '%s ' $f; tr -d '\\n' < $f | wc -m; done"
}
```

Result:

```text
l1.txt      495
l2.txt      538
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt
```

## Grill

Both lines fit the 600-character limit: line 1 is 495 characters and line 2 is 538. Nothing is written yet. Here is the command for you to approve. It adds the Goal with two lines (the third is dropped), run from `investigations/goal-prompt`.

```bash
cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \
  --alias run-practice \
  --title "A budget-bound run reviews its last candidate without a review-and-fix chain that goes on" \
  --body $'Discovery Goal for issue #141 (part of #113). A run of a build Goal that reaches its Result budget has reviewed its last candidate commit, and its reviews do not go on because of small changes. The candidates are six changes to the run practice (skills/hyp-goal-run and the stop clause hyp prompt prints), each a Hypothesis under one DesignQuestion that names its cut. A cut is the store as of a Result, made with hyp sandbox create --store --cut; the run first checks that --base hides the commits after the cut. A trial replays a Sonnet consumer from the cut with the arm\'s rule, as the cut\'s real actor so the budget left is the real one, in a clone with no later commit and without the text of issue #141. When the stop clause fires, a new run starts with hyp prompt as printed, a new actor and a fresh budget of the same size. An independent grader on the strongest model reads each trial. Each arm runs 3 trials at each of its cuts, and the control (the unchanged hyp-goal-run) runs 3 at the same cut. Test each rival alone first; if none holds every line, the run may add one combination Hypothesis. H-name-at-stop and H-whole-candidate carry no preset verdict. The build that changes skills/ and the prompt is a later Goal.\nMay change: the investigations/goal-prompt tree (.tree/events and .tree/evidence), in commits on the branch the prompt names; scratch stores and clones under $TMPDIR. The run may push that branch and open or update one PR, and never merges it.\nRead only: skills/, docs/, internal/, cmd/, plans/, and every investigations/ tree except goal-prompt.\nNever: edit skills/, docs/, internal/ or cmd/; that is the build Goal.\nNever: replace an installed hyp binary.\nNever: write under an owner\'s actor; if a step needs the owner, refer with a reason or record a reading, and stop.\nNever: commit to main or merge a PR.' \
  --line "$(tr -d '\n' < "$TMPDIR/l1.txt")" \
  --line "$(tr -d '\n' < "$TMPDIR/l2.txt")" \
  --why "Owner, 2026-11-13: the discovery Goal for #141; lines and routing confirmed in the grill session."
```

- **Branch.** The prompt will carry `exp/run-practice-141` through `--branch`. It is not in the body, because the body only holds the constraints.
- **Count.** The command carries two lines, `G-…#1` (the stop line) and `G-…#2` (the cost line).
- **Cites.** There are no `--cite` keys, because nothing settled is carried over.
- **Temp files.** The two line files are in this session's `$TMPDIR`. They hold the exact text above, so what I write is what I showed.

Do you approve this exact command? After your yes I'll run it. Then I'll add the DesignQuestion and the six rivals, each with its cut named. I'll run `hyp check` and record the end-of-grill checkpoint. I'll finish with the `hyp prompt` hand-off, which needs the Goal's add on the branch the prompt names.

## Owner, message 5

Approved. Use `owner-standin` as the actor, and run exactly that command. I can't write commands myself, and I'm not asking to push or open a PR in this session.

### Tool Bash (cwd: .)

```text
{
 "command": "cd investigations/goal-prompt && HYP_ACTOR=owner-standin hyp add goal \\\n  --alias run-practice \\\n  --title \"A budget-bound run reviews its last candidate without a review-and-fix chain that goes on\" \\\n  --body $'Discovery Goal for issue #141 (part of #113). A run of a build Goal that reaches its Result budget has reviewed its last candidate commit, and its reviews do not go on because of small changes. The candidates are six changes to the run practice (skills/hyp-goal-run and the stop clause hyp prompt prints), each a Hypothesis under one DesignQuestion that names its cut. A cut is the store as of a Result, made with hyp sandbox create --store --cut; the run first checks that --base hides the commits after the cut. A trial replays a Sonnet consumer from the cut with the arm\\'s rule, as the cut\\'s real actor so the budget left is the real one, in a clone with no later commit and without the text of issue #141. When the stop clause fires, a new run starts with hyp prompt as printed, a new actor and a fresh budget of the same size. An independent grader on the strongest model reads each trial. Each arm runs 3 trials at each of its cuts, and the control (the unchanged hyp-goal-run) runs 3 at the same cut. Test each rival alone first; if none holds every line, the run may add one combination Hypothesis. H-name-at-stop and H-whole-candidate carry no preset verdict. The build that changes skills/ and the prompt is a later Goal.\\nMay change: the investigations/goal-prompt tree (.tree/events and .tree/evidence), in commits on the branch the prompt names; scratch stores and clones under $TMPDIR. The run may push that branch and open or update one PR, and never merges it.\\nRead only: skills/, docs/, internal/, cmd/, plans/, and every investigations/ tree except goal-prompt.\\nNever: edit skills/, docs/, internal/ or cmd/; that is the build Goal.\\nNever: replace an installed hyp binary.\\nNever: write under an owner\\'s actor; if a step needs the owner, refer with a reason or record a reading, and stop.\\nNever: commit to main or merge a PR.' \\\n  --line \"$(tr -d '\\n' < \"$TMPDIR/l1.txt\")\" \\\n  --line \"$(tr -d '\\n' < \"$TMPDIR/l2.txt\")\" \\\n  --why \"Owner, 2026-11-13: the discovery Goal for #141; lines and routing confirmed in the grill session.\"; echo exit=$?"
}
```

Result:

```text
G-PFYTCC node.add
  line G-PFYTCC#1: [P] For each Hypothesis, replaying its named cut with its rule gives, in at least 2 of 3 trials, a run whose last candidate commit has a review Result at the first stop clause of that arm's own prompt (5 Results under the actor; 7 for the raise-budget rival). Naming the gap in the stop checkpoint does not satisfy this line. The control, skills/hyp-goal-run/SKILL.md at 9c8e498 (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e), runs at the same cut and is no candidate.
  line G-PFYTCC#2: [N] A Hypothesis costs more than 1.5 times the control at its cut. A trial costs the Results recorded from the cut until the select or until the runner owes nothing, across restarts; 15 Results without that, 45 minutes or a kill costs infinity. An arm costs the median of its 3 trials, infinity last. It fails if its median is infinity or 2 times it exceeds 3 times the control's; a finite arm passes an infinite control. Its skill text differs from the control's (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e).
exit=0
```

## Grill

The Goal is written as G-PFYTCC. Next are the runner writes, starting with the DesignQuestion.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd investigations/goal-prompt && export HYP_ACTOR=claude-grill-141 && hyp add dq --parent G-PFYTCC --alias q-practice \\\n --title \"Which change to the run practice keeps a review from being cut off by the budget or dragged out by small fixes?\" \\\n --body $'Each rival is a change to hyp-goal-run or to the stop clause hyp prompt prints, and each names its cut: the store as of one Result, replayed with hyp sandbox create --store --cut. The control arm (unchanged hyp-goal-run) runs at the same cut and is not a Hypothesis. The Goal\\'s lines are the cases. Cuts: run 1 after R-8JQPWG (actor claude-g113-1, 1 Result left of 5); run 2 after R-3YJ7T2 (actor claude-g113-2, 1 left of 5); run 3 after R-JHBCWS (actor claude-g113-3, 4 left of 5); final after R-BSWFXG (04142f3 is the candidate, its review limit open).' \\\n --why \"The contested question of issue #141; the owner chose one question for all six rivals.\" "
}
```

Result:

```text
Exit code 1
(eval):cd:1: no such file or directory: investigations/goal-prompt
```

## Grill

The shell is already in the store directory, so I'll rerun without the `cd`.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "pwd; export HYP_ACTOR=claude-grill-141 && hyp add dq --parent G-PFYTCC --alias q-practice \\\n --title \"Which change to the run practice keeps a review from being cut off by the budget or dragged out by small fixes?\" \\\n --body $'Each rival is a change to hyp-goal-run or to the stop clause hyp prompt prints, and each names its cut: the store as of one Result, replayed with hyp sandbox create --store --cut. The control arm (unchanged hyp-goal-run) runs at the same cut and is not a Hypothesis. The Goal\\'s lines are the cases. Cuts: run 1 after R-8JQPWG (actor claude-g113-1, 1 Result left of 5); run 2 after R-3YJ7T2 (actor claude-g113-2, 1 left of 5); run 3 after R-JHBCWS (actor claude-g113-3, 4 left of 5); final after R-BSWFXG (04142f3 is the candidate, its review limit open).' \\\n --why \"The contested question of issue #141; the owner chose one question for all six rivals.\""
}
```

Result:

```text
/Users/datguy/.hyp-sandbox/g146-entry-b-4/wt/investigations/goal-prompt
Q-RC1DHJ node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141; Q=Q-RC1DHJ\nadd(){ hyp add hyp --parent $Q --alias \"$1\" --title \"$2\" --body \"$3\" --why \"A rival of issue #141, listed by the owner; tested by replay from its cut.\" ; }\nadd h-name-at-stop \"H-name-at-stop: the stop checkpoint names each candidate commit that has no review\" $'Keep the practice. At the stop, the checkpoint names each candidate commit without a review Result. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2. Expected by the owner: line 1 contradicts it alone, since naming is not a review; it stays a part of any combination. No preset verdict.'\nadd h-reserve-last \"H-reserve-last: with one Result left, review the current candidate instead of committing another fix\" $'When one Result remains, the runner reviews the current candidate instead of committing another fix, and stops with the review\\'s limits recorded. A case whose check already passes closes as a limit, not with a new commit. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.'\nadd h-one-experiment \"H-one-experiment: a candidate's review is part of its gate Experiment\" $'One Experiment carries the gate verdicts and the review, and one Result carries the test verdicts and the review\\'s limits, so a candidate costs one Result again. hyp-goal-run states how a review is recorded and that it counts as part of the gate. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.'\nadd h-raise-budget \"H-raise-budget: the default of --experiments rises to 7\" $'The stop clause hyp prompt prints defaults to 7 Results under the actor, to fit a gate and a review per candidate. The trial prompt uses --experiments 7. Cuts: run 1 after R-8JQPWG and run 2 after R-3YJ7T2.'\nadd h-probe-closes \"H-probe-closes: a case the run's own probe passes is closed in the review's Result, not recorded as a limit\" $'A case that the run\\'s own probe runs and passes on the candidate is closed in the review Result, since a check ran it. A limit stays for a case no check ran or one the candidate fails. Cut: run 3 after R-JHBCWS, where N1 closes on 3cfd08d and no fix commit follows.'\nadd h-whole-candidate \"H-whole-candidate: the review before the select reads the cumulative diff from the Goal's base\" $'The review before the select reads the cumulative diff from the Goal\\'s base to the candidate, not the latest increment. Cut: final, after R-BSWFXG, with 04142f3 as the candidate. The owner expects it to lose on cost; no preset verdict.'\nhyp link G-PFYTCC external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this discovery Goal serves.\"\nhyp check 2>&1 | tail -5"
}
```

Result:

```text
H-Q11GVE node.add
H-SF1DXP node.add
H-6CSQT9 node.add
H-319HZX node.add
H-84B1N3 node.add
H-W931RM node.add
G-PFYTCC edge.add
warn: uncontained R-BSWFXG (script outside recipes/: .tree/evidence/E-ZMMVPN/probe-p1.sh)
warn: uncontained R-HXEBAF (script outside recipes/: .tree/evidence/E-ZG1BC8/probe-q1.sh)
warn: uncited R-FRF929: 2aeb0abe is in the observation but in no cited file
warn: uncited R-BSWFXG: 2aeb0abe is in the observation but in no cited file
check: 0 errors, 9 warnings
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "export HYP_ACTOR=claude-grill-141; hyp checkpoint --goal G-PFYTCC --summary \"Grill done: 2 lines written by owner-standin. Q-RC1DHJ holds six rivals (H-name-at-stop, H-reserve-last, H-one-experiment, H-raise-budget, H-probe-closes, H-whole-candidate), each with its cut named; no candidate is tied to a subject yet. Out-of-Goal: the build Goal (hyp-goal-run review practice, the stop clause, docs/how-hyp-improves-itself.md line 29) is added by the owner after the selects.\" --why \"end of grill\" && hyp status G-PFYTCC | head -30; hyp check | tail -1\nHYP_ACTOR=claude-run-141-1 hyp prompt G-PFYTCC --actor claude-run-141-1 --branch exp/run-practice-141 --experiments 8; echo exit=$?"
}
```

Result:

```text
G-PFYTCC checkpoint
GOAL G-PFYTCC A budget-bound run reviews its last candidate without a rev… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-PFYTCC#1 no candidate — [P] For each Hypothesis, replaying its named cut with its rule gives, in at least 2 of 3 trials, a run whose last candidate commit ha…
  G-PFYTCC#2 no candidate — [N] A Hypothesis costs more than 1.5 times the control at its cut. A trial costs the Results recorded from the cut until the select o…
FRONTIER
  Q-RC1DHJ open/untested Which change to the run practice keeps a review from being …
    H-Q11GVE (h-name-at-stop) proposed H-name-at-stop: the stop checkpoint names each candidate co…
    H-SF1DXP (h-reserve-last) proposed H-reserve-last: with one Result left, review the current ca…
    H-6CSQT9 (h-one-experiment) proposed H-one-experiment: a candidate's review is part of its gate …
    H-319HZX (h-raise-budget) proposed H-raise-budget: the default of --experiments rises to 7
    H-84B1N3 (h-probe-closes) proposed H-probe-closes: a case the run's own probe passes is closed…
    H-W931RM (h-whole-candidate) proposed H-whole-candidate: the review before the select reads the c…
STALE
OWED
  runner: Q-RC1DHJ test H-Q11GVE (h-name-at-stop)
  runner: Q-RC1DHJ test H-SF1DXP (h-reserve-last)
  runner: Q-RC1DHJ test H-6CSQT9 (h-one-experiment)
  runner: Q-RC1DHJ test H-319HZX (h-raise-budget)
  runner: Q-RC1DHJ test H-84B1N3 (h-probe-closes)
  runner: Q-RC1DHJ test H-W931RM (h-whole-candidate)
AWAITING OWNER
RECENT
  2026-11-13 01:35 claude-grill-141 edge.add G-PFYTCC — The issue this discovery Goal serves.
CHECKPOINT 2026-11-13 01:35 claude-grill-141: Grill done: 2 lines written by owner-standin. Q-RC1DHJ holds six rivals (H-name-at-stop, H-reserve-last, H-one-experiment, H-raise-budget, H-probe-closes, H-whole-candidate), each with its cut named;… (full: hyp show G-PFYTCC)
check: 0 errors, 9 warnings
/goal Run the hyp tree in investigations/goal-prompt with the hy
[... 1068 more characters]
```

## Grill

Last, I'll commit the new events and evidence on the branch the prompt names. I won't push or open a PR.

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-entry-b-4/wt && git status --short && git checkout -q -b exp/run-practice-141 && git add investigations/goal-prompt/.tree && git commit -q -m \"docs: add the run-practice Goal for #141 with its question and six rivals\" && git log --oneline -1 && git status --short | wc -l"
}
```

Result:

```text
?? investigations/goal-prompt/.tree/events/claude-grill-141.jsonl
?? investigations/goal-prompt/.tree/events/owner-standin.jsonl
e42c767 docs: add the run-practice Goal for #141 with its question and six rivals
       0
```

## Grill

The grill is done, and `hyp check` reports 0 errors (the 9 warnings were already in the store). The Goal is G-PFYTCC in `investigations/goal-prompt`, and nothing is pushed.

- **Goal:** G-PFYTCC, added under `owner-standin`, with the constraints block and two lines:
  - **#1 [P]:** the run's last candidate has a review Result at the arm's own first stop clause, in at least 2 of 3 trials. Naming the gap does not satisfy it.
  - **#2 [N]:** an arm costs more than 1.5 times the control at its cut.
- **Question:** Q-RC1DHJ holds six rivals, each with its cut named:
  - **Run-1 and run-2 cuts:** H-name-at-stop (H-Q11GVE), H-reserve-last (H-SF1DXP), H-one-experiment (H-6CSQT9) and H-raise-budget (H-319HZX).
  - **Run-3 cut:** H-probe-closes (H-84B1N3).
  - **Final cut:** H-whole-candidate (H-W931RM).
- **Link and checkpoint:** an external reference to issue #141, and an end-of-grill checkpoint.
- **Commit:** `e42c767` on the new local branch `exp/run-practice-141`.
- **Nothing for the run to ask you:** every open question was answered in the session. Under AWAITING OWNER the list is empty.

## `/goal` prompt for the run

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-run-141-1. Its Goal is G-PFYTCC: read it with hyp show G-PFYTCC, read hyp status G-PFYTCC, and name --goal G-PFYTCC on every hyp checkpoint.
Done when the latest printed hyp status G-PFYTCC shows "runner owes nothing: stop and report to the owner" under OWED, or 8 Results have been recorded under claude-run-141-1, or you cannot take the next step honestly and say why.
May change: the investigations/goal-prompt tree (.tree/events and .tree/evidence), in commits on the branch the prompt names; scratch stores and clones under $TMPDIR. The run may push that branch and open or update one PR, and never merges it.
Read only: skills/, docs/, internal/, cmd/, plans/, and every investigations/ tree except goal-prompt.
Never: edit skills/, docs/, internal/ or cmd/; that is the build Goal.
Never: replace an installed hyp binary.
Never: write under an owner's actor; if a step needs the owner, refer with a reason or record a reading, and stop.
Never: commit to main or merge a PR.
Commit on the branch exp/run-practice-141.
```

- **Push and PR:** the constraints block lets the run push the branch and open or update one PR. You said you don't want that in this session, so I did neither. The run is the one that may do it.
- **Not in the tree:** the build Goal that changes `skills/hyp-goal-run` and the stop clause. You add it after the selects. It also owns the stale cap sentence at `docs/how-hyp-improves-itself.md:29`.
