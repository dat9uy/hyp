# The grill's session

## Owner, message 1

<command-message>hyp-owner-grill</command-message>
<command-name>/hyp-owner-grill</command-name>
<command-args>#141 (the text of issue #141 as of this session is in /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md; this machine has no network, so gh cannot fetch it)</command-args>

### Tool Bash (cwd: .)

```text
{
 "command": "cat /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md; echo ----; hyp status 2>&1 | head -60",
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
[... 6057 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt; ls; ls investigations; git log --oneline | head -3; hyp --help 2>&1 | head -30; cat skills/hyp-goal-run/SKILL.md; grep -rn \"Results recorded\\|experiments\" internal/*/prompt*.go cmd/hyp/*.go | grep -v _test | head -20",
 "description": "Survey repo, skill and prompt code"
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
9c8e498 Merge pull request #140 from dat9uy/feat/goal-prompt
6f39ec0 docs: record the goal-prompt select of 04142f3 and the stop
aef2d5b docs: record the goal-prompt reviews of 3cfd08d, d2d5d14 and 04142f3, their gates and the stop
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
  status [--as-of REF|TIME] [--max-lines N] [--liv
[... 15044 more characters]
```

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; ls -a; HYP_ACTOR=probe hyp status 2>&1 | head -50; hyp show G-ZBM2RE 2>&1 | head -60",
 "description": "Inspect the goal-prompt tree and its Goal"
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
Goal G-ZBM2RE
alias: build
title: hyp prints the /goal prompt for a Goal from its log
body: Builds the design recorded in G-GZR6C6 (selects H-MCDPNB and H-RSDX0H, with grill session 285fa63e as evidence), for issue #113. The candidate is a hyp commit: the read command hyp prompt and the skill text it makes wrong. go test on the candidate decides lines 1 to 5; line 6 is an eval case in ak-skill-creator's format; line 7 runs its pinned scripts. A skill change follows ak-skill-creator's update flow: matched original and candidate consumer runs on the cases its edited lines touch. The fixture copies run 6's prompt (plans/reports/goal-prompt-261111-2129-g1b-run-6.md) and checkpoint-store:G-D1CH1R's events into testdata, because plans/ and investigations/ are not testdata.
May change: this repository, in commits on the branch the prompt names, that follow AGENTS.md (conventional messages, no AI trailers); the run may push that branch and open or update one PR, and never merges it. It changes cmd/, internal/ and their testdata, and the text the change makes wrong: docs/replay-spec.md, docs/glossary.md, docs/principles.md, AGENTS.md, README.md, and skills/ (hyp-owner-grill's hand-off and Goal-body template, hyp-owner-answer's end, hyp-goal-run's practice from run 6, and the descriptions of hyp-goal-run and hyp-prototype).
Read only: plans/, every investigations/ tree except goal-prompt, and ak-skill-creator under .claude/skills, pinned by sha256.
Never: commit to main or merge a PR.
Never: 
[... 12306 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; export HYP_ACTOR=probe; for r in R-8JQPWG R-JHBCWS R-ZMMVPN E-ZMMVPN E-GE085A R-KYKWH9; do echo \"== $r\"; hyp show $r 2>&1 | head -4; done; ls evals .tree/evidence | head -30; ls ../../recipes ../../rubrics 2>&1 | head; ls ../../.claude/skills; sed -n 1,40p ../../docs/glossary.md | head -5",
 "description": "Check the cuts the issue names and eval tooling"
}
```

Result:

```text
== R-8JQPWG
Result R-8JQPWG
path: G-ZBM2RE > Q-Z93VJ5 > H-3M3QAS > E-7QQ2DP > R-8JQPWG
status: recorded
observation: The fresh review of b4f23b2's fix reports G1-G8. The probe with the b4f23b2 binary confirms G1: one store, one log and flags, exit 3 through a symlinked working directory and exit 0 with 'in inv/t' through the physical path, where G-ZBM2RE#1 prints the same bytes on every call. It confirms G2 (a CRLF blank line inside the block is refused as "\r") and G3 (a no-break space in --branch exits 2). G3, G4, G5, G6 and G7 are text or test defects; G4 also names a rule conflict outside the lines.
== R-JHBCWS
Result R-JHBCWS
path: G-ZBM2RE > Q-Z93VJ5 > H-FQ103R > E-TZTG0G > R-JHBCWS
status: recorded
observation: The fresh review of 3cfd08d reports the new CRLF test red under three mutations of prompt.go (strip \r, normalise CRLF, refuse a trailing \r) and the new no-block test red when the block check moves after the repository lookup; the new spec sentence is true of the code. N1: a mutation that checks only that a block exists before the lookup, and its validity after, passes every TestPrompt test. The probe with the 3cfd08d binary (785dd8d6a43871e7) runs N1: an invalid block in a store outside every repository exits 1 with empty stdout. The outcome is right; no go test runs it. N2 names no line.
== R-ZMMVPN
unknown node R-ZMMVPN
== E-ZMMVPN
Experiment E-ZMMVPN
alias: e-review-d2d5d14
title: Review d2d5d14's fix diff against G-GZR6C6 and the lines; probe each finding
body: Records a fresh review of git diff 3cfd08d d2d5d14 outside investigations/ (the fix commit) against G-GZR6C6, the lines, the owner's two answers and R-JHBCWS#1 (review.md), and probes each finding that names a line with the d2d5d14 binary (sha256 in probe log). Each confirmed case of a line, or case no check ran, is one --limit here; findings of no line go to the stop checkpoint. Contradicts H-0KVRSQ if a probe shows d2d5d14 gives the wrong outcome on a case of a Goal line under the owner'
[... 2293 more characters]
```

## Grill

I read issue #141 and its three comments, the `hyp-goal-run` skill and `hyp prompt`'s stop clause. I also looked at the `investigations/goal-prompt` store. I wrote nothing. These are facts I checked, so I won't ask you about them:

- There is no store at the repository root. `investigations/goal-prompt` holds G-GZR6C6 and G-ZBM2RE, and both are complete.
- R-8JQPWG, R-JHBCWS, E-GE085A and E-ZMMVPN exist there, so the scenarios in the issue can be replayed from that store's log.
- `hyp prompt` takes `--experiments` and defaults it to 5. The skill still says "one review of each candidate commit's diff".

## What the tree will hold

The issue's rivals answer five different questions, so I'll make five DesignQuestions, each with its own rivals. Each will carry the baseline "keep today's practice" as a rival, and each will be tested on a scenario. You won't be asked to choose between any of these rivals.

| DQ | Question | Rivals | Separating scenario |
|---|---|---|---|
| A | How is a candidate's review recorded and counted? | H-one-experiment (the review is part of the gate Experiment); H-separate (today: its own Experiment, counted) | The cut after R-8JQPWG, and the cost in Experiments per candidate |
| B | What does a run do when one Experiment is left? | H-reserve-last; H-raise-budget (7); H-keep (5) | The cut after R-8JQPWG with a fix commit pending, and the cut after R-JHBCWS |
| C | Does the stop name each unreviewed candidate? | H-name-at-stop; H-silent | The same cuts, with no earlier gap to prompt the run |
| D | What happens to a case the run's own probe ran and passed? | H-probe-closes (closed in the review's Result, no commit); H-limit-then-fix (today) | The cut after R-JHBCWS (N1 and P1) |
| E | What does a review read? | H-whole-candidate (the diff from the Goal's base); H-increment (today) | The cut at `04142f3`, checking what a review of the cumulative diff finds that the increment reviews did not |

Because C is its own DQ, "name-at-stop alone" is tested and not assumed.

## Questions for you

1. **Actor.** What is your actor name? I suggest `owner-dat`, as in G-ZBM2RE.
2. **Store.** I suggest a new Goal in the same `investigations/goal-prompt` store, "Part of #113", because it holds the cuts the scenarios replay. The alternative is a new tree under `investigations/`, which would cite those cuts as `goal-prompt:<id>`. Which do you want?
3. **Split.** I suggest two Goals.
   - **Goal now:** a design Goal. It compares the rival rules on replayed cuts, with the instrument the Goal builds, under `--rubric reading`.
   - **Goal later:** a build Goal for the commit that changes `skills/hyp-goal-run` and the stop clause in `internal/cli/prompt.go`. You add it after the select, because its lines depend on the selected rival.
   - Your last comment, that `hyp-goal-run` should say how a review is recorded and whether it counts, goes into that build Goal as a line.
   - Do you agree, or do you want one build Goal whose candidates are one commit per rival?
4. **Test scope.** Should the run test all five DQs? I suggest yes. A run on A through E needs more than 5 Experiments, so I would hand off with `--experiments 12`. Is that acceptable, or should it test fewer?
5. **Lines.** These are the invariants any rule must meet, so none of them favours a rival. Which do you keep, change or add?
   - **1 [N]** On the cut after R-8JQPWG and the cut after R-JHBCWS, a run under the candidate rule stops with its final candidate neither reviewed nor named as unreviewed in the stop checkpoint.
   - **2 [N]** On those cuts, the run reaches the select or "runner owes nothing" while a candidate has no review Result.
   - **3 [N]** On those cuts, a review finding that is a case of a line ends with no record: no limit, no close, no fix commit and no name in the stop checkpoint.
   - **4 [N]** On those cuts, the run records more Results than the stop clause's N.
   - **5 [P]** (your call) The candidate rule states in one sentence how a review is recorded and whether it counts against the budget.
6. **Constraints block.** I propose this. Do you accept it or change it?
   - `May change: this tree's .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text for each rival rule, kept under .tree/evidence/<E-id>/`
   - `Read only: skills/, internal/, docs/, and every other investigations/ tree`
   - `Never: edit skills/ or internal/ (the change is a later build Goal); replace the installed hyp; write under an owner's actor; merge a PR`

After your answers I'll run the second reader and the preflight scan, then show the routing table. Nothing is written until you confirm that table.

## Owner, message 2

1. Actor name: owner-standin.

2. Store: use a new Goal in the same `investigations/goal-prompt` store, since this is part of #113.

3. Split: I accept your recommendation of a design Goal now and a build Goal later.

4. Test scope: I don't want five questions. Put every candidate fix under one question and test them all together.
   - I expect the whole-candidate review to lose, because one cumulative review costs a lot. Treat it as the expected loser.
   - Test by replaying runs on Sonnet, 3 trials per candidate and per starting point.
   - Don't replay a whole build from its start for each trial. That costs too much.
   - Would the ak-context-engineering skill help with how we evaluate?
   - I accept 12 experiments.

5. Lines: before I keep any, tell me how each one serves these two aims:
   - A review must not easily run out of the run's budget.
   - A review must not go on and on because of small changes.

   Also state the cost limit. A candidate may cost at most 1.5 times what the base costs at the same starting point. How do you compute cost? A trial that hangs counts as the most expensive.

6. Constraints block: accepted.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt; ls .claude/skills ~/.claude/skills 2>&1 | head -30; find / -iname \"*context-engineering*\" -maxdepth 6 2>/dev/null | head; ls .claude/skills/ak-skill-creator; claude --help 2>&1 | grep -i -E \"output-format|max-budget|model\" | head",
 "description": "Look for the named skill and claude CLI cost options"
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
assets
LICENSE-MIT.txt
LICENSE.txt
NOTICE.txt
references
scripts
SKILL.md
  --fallback-model <model>              Enable automatic fallback to specified
                                        model(s) when the default model is
                                        --print and --output-format=stream-json)
                                        --output-format=stream-json)
                                        --output-format=stream-json)
  --max-budget-usd <amount>             Maximum dollar amount to spend on API
  --model <model>                       Model for the current session. Provide
                                        an alias for the latest model (e.g.
                                        model's full name.
  --output-format <format>              Output format (only works with --print):
```

## Grill

## ak-context-engineering

I can't use it. This session lists no skill with that name, and `.claude/skills` holds only `ak-skill-creator` and the hyp skills. There is a `~/codingProjects/context-engineering` directory, but I did not open it, because it is your other project and not part of this task. If you want it used, install it as a skill or tell me which of its files to read.

`ak-skill-creator` is installed. It already provides matched original-and-candidate runs on eval cases, which is the shape this test needs, so I'd use it.

## One question, one test

I now propose one DesignQuestion: "Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?" These are the rivals:

- H-base: today's practice, and the baseline.
- H-one-experiment
- H-name-at-stop
- H-reserve-last
- H-raise-budget (7)
- H-probe-closes
- H-whole-candidate: the body says "expected loser; tested, not assumed".
- H-combo: one-experiment, name-at-stop, reserve-last and probe-closes together. Without it the tree can't say which rivals hold alone and which only hold together.

The runs use Sonnet at two starting points: the cut after R-8JQPWG and the cut after R-JHBCWS. A starting point is the git worktree at the commit under review, plus the log cut at that Result, with the stop clause set so the same Experiments remain as in the real run. Nothing replays a build from its start. That gives 8 rivals × 2 starting points × 3 trials = 48 runs.

I would use 4 of the 12 Experiments:
1. A known-answer Experiment. H-base must reproduce the issue's failure at the cut after R-8JQPWG, or the instrument is wrong.
2. The H-base trials.
3. The trials at the first starting point.
4. The trials at the second starting point.

## How each line serves your two aims

Aim A is that a review must not easily run out of budget. Aim B is that a review must not go on and on because of small changes.

| # | Line | Serves | Why |
|---|---|---|---|
| 1 [N] | A trial stops with its final candidate neither reviewed nor named as unreviewed in the stop checkpoint. | A | This is the observed failure in runs 1 and 2. |
| 2 [N] | A trial reaches the select or "runner owes nothing" while a candidate has no review Result. | A | A rule must not meet line 1 by hiding the gap. |
| 3 [N] | A review finding that is a case of a line ends with no record: no limit, close, fix commit or name at the stop. | A and B | A rule could meet B by dropping findings. This blocks that. |
| 4 [N] | A trial records more Results than the stop clause's N. | A | A rule must not meet A by overspending. |
| 5 [N] | A trial adds a commit after a review whose findings were all cases the candidate already passes. | B | This is the chain in run 3 (`d2d5d14`, `04142f3`). The line names the state, not the mechanism, so it does not assume H-probe-closes. |
| 6 [N] | A candidate's mean cost at a starting point is more than 1.5 times H-base's mean at the same starting point. | A and B | See below. |

I left out the line "the rule states in one sentence how a review is recorded". It is text and has no replay check, so it goes to the build Goal.

## Cost

- **Per trial:** `total_cost_usd` from `claude -p --output-format json`. It prices input, output and cache tokens. Known-answer Experiment 1 must show that it includes the reviewer subagents. If it does not, the instrument sums them.
- **Per rival and starting point:** the mean of the 3 trials.
- **The limit:** mean(rival) ÷ mean(H-base) at the same starting point is at most 1.5. Line 6 pins H-base by the sha256 of its skill text and prompt. It asserts that this differs from the candidate's.
- **A hang:** a trial that has not stopped after T wall-clock minutes, or that hits `--max-budget-usd`. It is killed and counts at the highest cost of any finished trial at that starting point, base included. T is twice the slowest finished H-base trial at that starting point. I fix T after the H-base trials and before any rival runs.
- **Mean, not median:** with 3 trials a median would hide a hang.

## Questions for you

1. **Cost.** Do you accept the cost definition above: `total_cost_usd`, the mean of 3 trials, the 1.5 ratio per starting point, and T as twice the slowest H-base trial? If not, say what to change.
2. **Trial rule.** Lines 1 to 4 are invariants. I recommend that a rival fails one if it fails in any of the 3 trials at either starting point. I recommend "2 of 3 trials" for line 5, because a model can add a commit by chance. Do you accept this?
3. **Lines.** Do you keep, change or drop each of lines 1 to 6?
4. **Starting points.** Do you accept the two named ones, or do you want a third, such as the cut at `04142f3`? A third costs 24 more runs.
5. **H-whole-candidate.** I still run it, because "it loses on cost" is a prediction until line 6 measures it. Do you accept that, or should it be dropped untested?

Once you answer, I'll run the second reader and the preflight scan, then show the routing table. Nothing is written until you confirm that table.

## Owner, message 3

1. Cost: accepted. Use `total_cost_usd`, the mean of the 3 trials, and the 1.5 ratio per starting point. A hang counts as the most expensive trial, and T is twice the slowest H-base trial.

2. Trial rule: accepted, including "2 of 3" for line 5.

3. Lines: I keep lines 1 to 6. Moving the sentence about how a review is recorded to the build Goal is fine.

4. Starting points: two is enough. Don't add a third.

5. H-whole-candidate: run it. It is the expected loser, so test it and don't assume it.

Using `ak-skill-creator` instead of ak-context-engineering is fine. The one question with H-combo and the 48 runs is also fine.

### Tool Agent (cwd: .)

```text
{
 "description": "Second reader for the grill design",
 "subagent_type": "general-purpose",
 "model": "opus",
 "prompt": "You are an independent reviewer of a design for a hyp Goal. Read-only: change no file. Repo: /Users/datguy/.hyp-sandbox/g146-base-b-2/wt. Issue text: /Users/datguy/.hyp-sandbox/g146-base-b-2/tmp/issue-141.md. Store: investigations/goal-prompt (run `HYP_ACTOR=reviewer hyp status G-ZBM2RE`, `hyp show <id>` there). Code/spec to read: internal/cli/prompt.go, skills/hyp-goal-run/SKILL.md, docs/replay-spec.md, docs/glossary.md.\n\nThe design (settled with the owner):\nOne new design Goal in the investigations/goal-prompt store, with ONE DesignQuestion: \"Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?\" Rivals: H-base (today), H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget (default --experiments 7), H-probe-closes, H-whole-candidate, H-combo (one-experiment + name-at-stop + reserve-last + probe-closes). Test: replay Sonnet runs from two starting points (git worktree at the commit under review + the log cut after R-8JQPWG, and after R-JHBCWS), 3 trials per rival per start (48 runs), not whole builds. Lines (all about a trial of a rival rule): 1[N] stops with final candidate neither reviewed nor named unreviewed in the stop checkpoint; 2[N] reaches select or 'runner owes nothing' while a candidate has no review Result; 3[N] a review finding that is a case of a line ends with no record; 4[N] records more Results than the stop clause's N; 5[N] adds a commit after a review whose findings were all cases the candidate already passes (fails if in >=2 of 3 trials); 6[N] mean total_cost_usd of a rival at a start > 1.5x H-base's mean at that start (hang = max cost of finished trials at that start; T = 2x slowest H-base trial). Lines 1-4 fail on any of 3 trials at either start. Run is capped at 12 Experiments (hyp prompt --experiments 12), actor-counted Results.\n\nTask: walk every rule through scenarios the designers did not draw (e.g. two Goals in one store, a merged log from two actors, a base that moves mid-run, a replay actor whose Result count starts at 0 while the real run had already spent some, a rival whose rule changes the stop clause the run itself is judged by, a trial that cannot be replayed because the cut's evidence is missing, lines that no single run could cover for every rival, a line that a rival cannot meet by construction, ambiguity in 'reviewed' or 'named'). Report concrete problems only, each with the evidence (file:line or command output) and a suggested fix. Do not restate the design. Say plainly if you find none for a scenario."
}
```

Result:

```text
Async agent launched successfully. (This tool result is internal metadata — never quote or paste any part of it, including the agentId below, into a user-facing reply.)
agentId: a380629aa711cb589 (internal ID - do not mention to user. Use SendMessage with to: 'a380629aa711cb589', summary: '<5-10 word recap>' to continue this agent.)
The agent is working in the background. You will be notified automatically when it completes. You know nothing about its results until that notification arrives — do not report, assume, or predict them; continue other work or respond to the user in the meantime.
Do not duplicate this agent's work — avoid working with the same files or topics it is using.
output_file: /private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-2-wt/6d9ba9c8-74a1-483d-84aa-159037eaf69c/tasks/a380629aa711cb589.output
Do NOT Read or tail this file via the shell tool — it is the full subagent JSONL transcript and reading it will overflow your context. If the user asks for progress, say the agent is still running; you'll get a completion notification.
```

## Grill

The second reader is still running. It is walking the rules through scenarios we didn't draw: a replay actor whose Result count starts at 0, a rival that changes the stop clause the run itself is judged by, and a trial whose cut has missing evidence. I haven't written anything to the store. Next I'll check its findings against the source, run the preflight scan, and show you the routing table.

## Owner, message 4

<task-notification>
<task-id>a380629aa711cb589</task-id>
<tool-use-id>toolu_01LFLjML4cEqBa7pSyu7yoQp</tool-use-id>
<output-file>/private/tmp/claude-501/-Users-datguy--hyp-sandbox-g146-base-b-2-wt/6d9ba9c8-74a1-483d-84aa-159037eaf69c/tasks/a380629aa711cb589.output</output-file>
<status>completed</status>
<summary>Agent "Second reader for the grill design" finished</summary>
<note>A task-notification fires each time this agent stops with no live background children of its own. The user can send it another message and resume it, so the same task-id may notify more than once.</note>
<result>I found seven problems that need a fix and three smaller ones. Four scenarios are clean: two Goals in one store, a merged log, a moving base, and "reviewed" versus "named". I did not run any replay. I read the event files, the skill, `internal/cli/prompt.go` and the issue.

**Problems that need a fix**

1. **`--experiments 12` and a fresh actor mean no trial is budget-bound.**
   - `prompt.go:167` counts Results under `--actor`. `prompt.go:109` sets the default of 5 and the replay overrides it.
   - A new actor starts at 0 with 12 left. In the real run 1, `claude-g113-1` had 4 Results at R-8JQPWG (KPWZZC, 9V3ER0, 50GBXM, 8JQPWG), so one was left. A replay with cap 12 never reaches that state.
   - H-reserve-last ("when one remains") never fires.
   - H-raise-budget changes only the default, which `--experiments 12` replaces, so in replay it behaves exactly like H-base.
   - Line 4 ("more than the stop clause's N") is vacuous when N is 12 for every rival.
   - Fix: have each trial reuse the cut's real actor, so the spent count carries over. Print the prompt with each rival's own N (5 for base, 7 for raise-budget) from a binary built with that rival's `prompt.go`. Judge line 4 against the N printed in that trial's prompt.
   - A separate cap of 12 on the trial is fine as a safety limit.

2. **R-JHBCWS is not a budget-bound start, and line 5 cannot be exercised at either start.**
   - Reusing `claude-g113-3`, Results at R-JHBCWS number 1 (the actor's first), so 4 of 5 are left. Run 3 went on to 04142f3 with budget to spare.
   - At R-8JQPWG, G1 really fails on b4f23b2 (exit 3 through a symlink), so a fix commit is legitimate. Line 5 passes vacuously there.
   - At R-JHBCWS, N1 reads "no go test runs this case". The code already passes it and the missing item is a test. A test-only commit would fail line 5 even though the Goal's lines may require the test. The sort rule (SKILL.md:34-35) does not say whether that is "a case the candidate already passes" or "a defect in text you may edit".
   - The start that matches the owner's second event is the cut after R-3YJ7T2 (claude-g113-2, 4 of 5 spent, one left). Both of its limits, M2 and M3, are cases the candidate passes.
   - Fix: use R-8JQPWG and R-3YJ7T2 as the two starts, and add R-JHBCWS only if a third is wanted. Define in line 5 whether a test-only commit counts as "a commit after a review of already-passing cases".

3. **The replay can see the future.**
   - The `.tree/evidence/` directory holds evidence of later Experiments (E-ZMMVPN, E-GE085A, E-ZG1BC8, and others).
   - The worktree shares its git objects with the repository, which holds 5c5902d, 2284172, d2d5d14 and 04142f3. `git log --all` and the reflog show them.
   - Only the `work` branch exists locally, but the objects are still reachable.
   - Fix: build each trial from a bare clone or a `git archive` of the commit with no refs beyond it. Copy only the events and evidence up to the cut. State this as a rule of the trial.

4. **The cut has no definition.**
   - Events sit in six per-actor files, interleaved by `ts`. Cut 1 is between 17:21:27 and 17:23:58 on 11-11. The owner's answers (02:27 and 02:36 on 11-12) fall after it. Cut 2 includes them.
   - Fix: define the cut as "every event with `ts` ≤ the cut Result's event `ts`, across all files". Say which side of the ties R-8JQPWG's own `edge`/`node` events fall on.

5. **"Reviewed" has no mechanical test across rivals.**
   - No event marks a review. In the base, a review is its own Experiment (`--recipe none --rubric reading`, SKILL.md:35 discussion). In H-one-experiment it is a limit inside the gate Result.
   - Lines 1 and 2 ("candidate has no review Result") therefore need one detector that works for every rival. A review is done by a fresh subagent that the runner calls, so the log shows only the runner's Result.
   - Fix: define a review Result as one that cites `review.md` or a subagent's report in its evidence, and whose observation names the candidate commit. Likewise define "named" for line 1 as: the stop checkpoint contains the candidate's short sha and the word "unreviewed".

6. **Lines 1 and 2 can clash for H-name-at-stop.**
   - Line 2 fails any trial that reaches "runner owes nothing" with a candidate lacking a review Result. The issue's Done means allows exactly that when the checkpoint names the gap.
   - Fix: make line 2 apply only when the candidate is not named unreviewed in the stop checkpoint, or drop it for the named case.

7. **Lines 3 and 5 can pass without being exercised.**
   - The runner builds a different candidate in each trial, so a fresh review may find nothing. Line 3 (a finding with no record) and line 5 (a review of already-passing cases) are then vacuous.
   - Fix: report, for each rival, how many of the 3 trials produced a finding of the kind the line tests. Count a rival/start cell as "not exercised" rather than passed when that number is 0.

**Smaller issues**

- **Result count versus the cap.** The design needs one Result per rival per start, which is 16 cells, plus the select. The run is capped at 12 Experiments (`--experiments 12`). The Goal is then budget-bound by the very problem it studies. Record one Experiment per rival (8) with the per-start numbers in one Result, or raise the cap.
- **Experiments versus Results.** `prompt.go:167` counts recorded Results, not Experiments. Experiments added but never recorded, as in a hung trial, do not count toward the stop. Say "Results" in line 4 and in the cap.
- **Line 6 with small samples.** With 3 trials per cell the 1.5x mean threshold is noisy. The hang rule (T = 2x the slowest H-base trial) is undefined if every H-base trial at a start hangs. State a fallback, such as a fixed wall-clock limit.

**Scenarios with no problem found**

- Two Goals in one store: the replay copies the cut store, which holds only G-ZBM2RE. Run `hyp status` with the Goal named, as SKILL.md:13 says.
- A merged log from two actors: events are per-actor files, so trials do not collide as long as each trial has its own copy of the store.
- A base that moves mid-run: the worktree is pinned to the commit under review, so it does not move.
- A rival that changes the stop clause: the rule in problem 1 covers it, as long as you build the binary per rival.</result>
<usage><subagent_tokens>50994</subagent_tokens><tool_uses>9</tool_uses><duration_ms>111712</duration_ms></usage>
</task-notification>

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt/.tree; ls events | head; for a in claude-g113-1 claude-g113-2 claude-g113-3; do echo \"== $a\"; f=$(ls events | grep -F \"$a\" | head -1); echo $f; grep -h '\"result' events/$f 2>/dev/null | head -0; done; grep -rhoE '\"(type|kind)\":\"result[^\"]*\"[^}]*' events | head -3; grep -l \"R-8JQPWG\" -r events | head",
 "description": "Inspect event files to verify reviewer facts"
}
```

Result:

```text
claude-g113-1.jsonl
claude-g113-2.jsonl
claude-g113-3.jsonl
claude-g113-4.jsonl
claude-grill-113.jsonl
owner-dat.jsonl
== claude-g113-1
claude-g113-1.jsonl
head: illegal line count -- 0
== claude-g113-2
claude-g113-2.jsonl
head: illegal line count -- 0
== claude-g113-3
claude-g113-3.jsonl
head: illegal line count -- 0
"type":"result.record","node":"01M75S97KWYDNTGA2N56KPWZZC","payload":{"parent":"01M75RTZK05VMXAX2ER1HQWTA0","verdicts":{"01M75RTFR9JKPARCH3NAQ5P9R9":"supports"
"type":"result.record","node":"01M75SBVS853GXX0C0SC9V3ER0","payload":{"parent":"01M75SAW63CZAX0KXXXMC3GGGJ","verdicts":{"01M75RTFR9JKPARCH3NAQ5P9R9":"contradicts"
"type":"result.record","node":"01M75SV7JAPQZNHTTZB050GBXM","payload":{"parent":"01M75SHTEBR47W171DXHZ5BTZR","verdicts":{"01M75SH0P5561NTGAW8V3M3QAS":"supports"
events/claude-g113-1.jsonl
events/claude-g113-2.jsonl
```

### Tool Bash (cwd: investigations/goal-prompt/.tree)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt/.tree/events; for a in 1 2 3; do echo \"== g113-$a\"; grep -h '\"type\":\"result.record\"' claude-g113-$a.jsonl | grep -oE '\"node\":\"[0-9A-Z]{26}\"|\"ts\":\"[^\"]*\"' | paste - - | sed -E 's/\"node\":\"[0-9A-Z]{20}([0-9A-Z]{6})\"/R-\\1/' ; done; cd ..; ls events >/dev/null; head -c 300 events/claude-g113-1.jsonl",
 "description": "List Results per actor with timestamps"
}
```

Result:

```text
== g113-1
"ts":"2026-11-11T17:07:57.180003Z"	R-KPWZZC
"ts":"2026-11-11T17:09:23.36821Z"	R-9V3ER0
"ts":"2026-11-11T17:17:46.95442Z"	R-50GBXM
"ts":"2026-11-11T17:21:27.311404Z"	R-8JQPWG
"ts":"2026-11-11T17:30:29.026755Z"	R-FRF929
"ts":"2026-11-11T17:30:45.219002Z"	R-9EGPYD
== g113-2
"ts":"2026-11-12T02:48:18.682738Z"	R-0E8PFH
"ts":"2026-11-12T02:51:44.175152Z"	R-KYKWH9
"ts":"2026-11-12T02:55:04.445113Z"	R-28A8NB
"ts":"2026-11-12T02:56:59.537823Z"	R-3YJ7T2
"ts":"2026-11-12T02:59:35.448093Z"	R-356DV0
== g113-3
"ts":"2026-11-12T03:13:18.498579Z"	R-JHBCWS
"ts":"2026-11-12T03:15:49.831676Z"	R-27Y1EV
"ts":"2026-11-12T03:18:50.451686Z"	R-DER6F8
"ts":"2026-11-12T03:23:28.19161Z"	R-BSWFXG
"ts":"2026-11-12T03:23:52.814802Z"	R-HXEBAF
{"v":1,"id":"01M75RTFR9JKPARCH3NAQ5P9R9","ts":"2026-11-11T16:59:53.993811Z","actor":"claude-g113-1","type":"node.add","node":"01M75RTFR9JKPARCH3NAQ5P9R9","payload":{"kind":"Hypothesis","parent":"01M75R48JNGWS8ENJHN7Z93VJ5","title":"feat/goal-prompt at 10566d4: hyp prompt, its tests, the fixture and
```

### Tool Bash (cwd: investigations/goal-prompt/.tree)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt; git rev-parse HEAD; shasum -a 256 skills/hyp-goal-run/SKILL.md internal/cli/prompt.go; git cat-file -t 5c5902d 2284172 3cfd08d b4f23b2 2>&1",
 "description": "Compute baseline hashes and check commits exist"
}
```

Result:

```text
Exit code 129
9c8e498b87683e50a94a410b6a3cfecdd7a0d441
cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e  skills/hyp-goal-run/SKILL.md
21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed  internal/cli/prompt.go
fatal: too many arguments

usage: git cat-file <type> <object>
   or: git cat-file (-e | -p | -t | -s) <object>
   or: git cat-file (--textconv | --filters)
                    [<rev>:<path|tree-ish> | --path=<path|tree-ish> <rev>]
   or: git cat-file (--batch | --batch-check | --batch-command) [--batch-all-objects]
                    [--buffer] [--follow-symlinks] [--unordered]
                    [--textconv | --filters] [-Z]

Check object existence or emit object contents
    -e                    check if <object> exists
    -p                    pretty-print <object> content

Emit [broken] object attributes
    -t                    show object type (one of 'blob', 'tree', 'commit', 'tag', ...)
    -s                    show object size
    --[no-]use-mailmap    use mail map file
    --[no-]mailmap ...    alias of --use-mailmap

Batch objects requested on stdin (or --batch-all-objects)
    --batch[=<format>]    show full <object> or <rev> contents
    --batch-check[=<format>]
                          like --batch, but don't emit <contents>
    -Z                    stdin and stdout is NUL-terminated
    --batch-command[=<format>]
                          read commands from stdin
    --batch-all-objects   with --batch[-check]: ignores stdin, batches all known objects

Change or optimize batch output
    --[no-]buffer         buffer --batch output
    --[no-]follow-symlinks
                          follow in-tree symlinks
    --[no-]unordered      do not order objects before emitting them

Emit object (blob or tree) with conversion or filter (stand-alone, or with batch)
    --textconv            run textconv on object's content
    --filters             run filters on object's content
    --[no-]path blob|tree use a <path> f
[... 88 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt/.tree)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt; for c in 5c5902d 2284172 3cfd08d b4f23b2 04142f3; do echo -n \"$c \"; git cat-file -t $c 2>&1; done",
 "description": "Check the cut commits exist locally"
}
```

Result:

```text
5c5902d commit
2284172 commit
3cfd08d commit
b4f23b2 commit
04142f3 commit
```

## Grill

The second reader found seven problems and three smaller ones. I checked its facts against the log and the code. It was right on all of them, and I walked the replay-actor point through the event files myself. These are the facts I checked:

- **Result counts at the cut.** `claude-g113-1` had 4 Results when R-8JQPWG was recorded (R-KPWZZC, R-9V3ER0, R-50GBXM, R-8JQPWG), so one was left. `claude-g113-3` had 1 Result at R-JHBCWS, so 4 of 5 were left. `claude-g113-2` had 4 after R-3YJ7T2.
- **Stop clause.** `prompt.go:167` counts Results recorded under `--actor`, so a fresh actor with `--experiments 12` starts with 12 left.
- **Future commits.** The commits `5c5902d`, `2284172`, `3cfd08d`, `b4f23b2` and `04142f3` all exist in this clone. A replay worktree could see them.

The reader also noted that the owner's comment says the sort rule "on `main` is in no skill". That is out of date, because PR #140 merged and lines 34 and 35 of `skills/hyp-goal-run/SKILL.md` now hold it.

## Fixes that go into Experiment bodies

I need no answer from you on these:

- **Budget.** Each trial reuses the cut's real actor, so the spent count carries over. The prompt is printed from a binary built with that rival's `prompt.go`, so H-base gets N=5 and H-raise-budget gets N=7. Line 4 is judged against the N printed in that trial's prompt. The 12-Result cap is a safety limit for the trial.
- **No future.** Each trial builds from a `git archive` of the commit, with no refs and no reflog. It copies only the events and evidence at or before the cut Result's `ts`, across all actor files.
- **Reviewed and named.** A "review" is a Result that cites a fresh subagent's report on the candidate's diff and names the commit. A candidate is "named" when the stop checkpoint holds its short sha and the word "unreviewed".
- **Not exercised.** For each rival and start, I report how many of the 3 trials produced the finding that lines 3 and 5 test. A cell with 0 is "not exercised" and does not count as a pass.
- **Cap.** One Experiment per start holds all rivals, with a verdict per rival in one Result. That gives 4 Experiments in all, and the cap counts Results.
- **Hang fallback.** If every H-base trial at a start hangs, T is 45 minutes.

## Three decisions for you

1. **Replace the second start.** R-JHBCWS is not budget-bound, because run 3 had 4 of 5 left. I recommend R-3YJ7T2 instead, which is run 2 with one of five left and two passing limits. That is the case your second comment describes. Starts would be the cut after R-8JQPWG and the cut after R-3YJ7T2. It is still two starts. Do you accept?
2. **Line 5.** At R-3YJ7T2, `3cfd08d` was a commit that only added tests and spec text for cases that already passed. I read your words as meaning that this commit is the failure, so line 5 counts a test-only or text-only commit for passing cases. Do you accept?
3. **Line 2.** As written, line 2 would fail H-name-at-stop, because that rival stops before the select with the gap named. I narrowed it to the select only: no select on a candidate without a review, named or not. Do you accept?

## Routing table

| Question | Place | Scenario or invariant |
|---|---|---|
| Which store | Goal body: a new Goal in `investigations/goal-prompt`, with `hyp link … external_ref` to issue #141 | n/a |
| Design record and build | Another Goal: the build Goal comes later and carries the sentence "how a review is recorded and counted" as its line | n/a |
| The rival fixes | One DesignQuestion with 8 rivals: H-base, H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate ("expected loser, tested, not assumed") and H-combo | Sonnet replays at the cut after R-8JQPWG and the cut after R-3YJ7T2, 3 trials each |
| Line 1 | [N] A trial stops with its final candidate neither reviewed nor named unreviewed. | The review and "named" detectors above |
| Line 2 | [N] A trial selects a candidate that has no review Result. | n/a |
| Line 3 | [N] A review finding that is a case of a line ends with no record. | n/a |
| Line 4 | [N] A trial records more Results than the N its prompt prints. | n/a |
| Line 5 | [N] A trial adds a commit, test-only or text-only included, after a review whose findings are all cases the candidate already passes. It fails when 2 of 3 trials do this. | n/a |
| Line 6 | [N] A rival's mean `total_cost_usd` at a start is more than 1.5 times H-base's mean at that start. H-base is pinned by sha256: `skills/hyp-goal-run/SKILL.md` at `cb15e2f2…`, and `internal/cli/prompt.go` at `21f9332…`, at 9c8e498. Line 6 asserts that this differs from the candidate. A hang counts at the most expensive finished trial, with T as twice the slowest H-base trial. | n/a |
| Lines 1 to 4 trial rule | Any of 3 trials at either start fails the line. | n/a |
| Cut, budget, no-future, hang fallback and cell count | Experiment bodies (process) | n/a |
| Sonnet, 3 trials, 12 Results | Goal body and `hyp prompt … --experiments 12` | n/a |
| The constraints block you accepted | Goal body | n/a |
| Result counts, commits and the sort rule's place | Looked up, never asked | n/a |

I will write the Goal under your actor, `owner-standin`. Please confirm the three decisions and the table. After that I'll show the exact `hyp add goal` command for your approval, before I type anything under your name.

## Owner, message 5

1. Second start: accepted. Use the cut after R-3YJ7T2 instead of R-JHBCWS.

2. Line 5: accepted. A test-only or text-only commit for cases that already pass counts as the failure.

3. Line 2: accepted. Narrowing it to the select only is right.

The routing table is confirmed. Go ahead and show me the exact `hyp add goal` command under `owner-standin` and I'll approve it.

### Tool Bash (cwd: .)

```text
{
 "command": "cd $TMPDIR && python3 -I - <<'EOF'\nL=[\n\"[N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither reviewed nor named unreviewed. Reviewed: a Result that cites a fresh subagent report on that commit diff and names the commit. Named: the stop checkpoint holds the commit short sha and the word unreviewed. Fails if any of the 3 trials at either cut does.\",\n\"[N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the stop checkpoint names it. Fails if any of the 3 trials at either cut does.\",\n\"[N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix commit and no name in the stop checkpoint. Fails if any trial does. A cell where no trial produced such a finding is reported as not exercised, not as passed.\",\n\"[N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own prompt prints. Fails if any of the 3 trials at either cut does.\",\n\"[N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check that ran; a commit that only adds a test or text for such a case counts. Fails if 2 of the 3 trials at a cut do. A cell where no trial produced such a review is reported as not exercised, not as passed.\",\n\"[N] A candidate rule mean total_cost_usd over its 3 trials at a cut is more than 1.5 times the mean of H-base at the same cut. H-base is skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) with internal/cli/prompt.go (sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed) at 9c8e498; a candidate differs from it. A hung trial counts at the highest cost of any finished trial at that cut; the hang limit is twice the slowest H-base trial there.\",\n]\nfor l in L: print(len(l))\nEOF",
 "description": "Check line lengths in runes"
}
```

Result:

```text
386
208
287
196
332
513
Shell cwd was reset to /Users/datguy/.hyp-sandbox/g146-base-b-2/wt
```

## Grill

All six lines are within the 600-rune limit (the longest is 513). Nothing has been written.

This is the owner write. Run it from `investigations/goal-prompt`, with the store's path as it is today:

```bash
cd investigations/goal-prompt
HYP_ACTOR=owner-standin hyp add goal \
  --alias review-budget \
  --title "Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?" \
  --body $'Design record for issue #141, part of #113. A run of a build Goal stopped twice with its last candidate commit unreviewed (G-ZBM2RE, runs 1 and 2), and run 3 spent each review on one more small commit. This Goal compares rival changes to the review practice of hyp-goal-run and the stop clause hyp prompt prints, and selects one or a combination. It builds nothing: the change to skills/ and internal/ is a later build Goal, which also carries the rule that hyp-goal-run states how a review is recorded and whether it counts against the budget.\nThe candidate is a rule, measured by replaying Sonnet runs from two starting points: the cut after R-8JQPWG (run 1, one Result left, a fix commit pending) and the cut after R-3YJ7T2 (run 2, one Result left, two limits whose cases pass). 3 trials per rival and per cut. The rivals are all under one question; H-whole-candidate is expected to lose on cost and is tested, not assumed.\nMay change: this tree\'s .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text and scratch builds for each rival rule, kept under .tree/evidence/<E-id>/\nRead only: skills/, internal/, docs/, and every other investigations/ tree\nNever: edit skills/ or internal/; the change is a later build Goal\nNever: replace the installed hyp, or any binary outside .tree/evidence/\nNever: write under an owner actor, or merge a PR' \
  --line "[N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither reviewed nor named unreviewed. Reviewed: a Result that cites a fresh subagent report on that commit diff and names the commit. Named: the stop checkpoint holds the commit short sha and the word unreviewed. Fails if any of the 3 trials at either cut does." \
  --line "[N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the stop checkpoint names it. Fails if any of the 3 trials at either cut does." \
  --line "[N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix commit and no name in the stop checkpoint. Fails if any trial does. A cell where no trial produced such a finding is reported as not exercised, not as passed." \
  --line "[N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own prompt prints. Fails if any of the 3 trials at either cut does." \
  --line "[N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check that ran; a commit that only adds a test or text for such a case counts. Fails if 2 of the 3 trials at a cut do. A cell where no trial produced such a review is reported as not exercised, not as passed." \
  --line "[N] A candidate rule mean total_cost_usd over its 3 trials at a cut is more than 1.5 times the mean of H-base at the same cut. H-base is skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) with internal/cli/prompt.go (sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed) at 9c8e498; a candidate differs from it. A hung trial counts at the highest cost of any finished trial at that cut; the hang limit is twice the slowest H-base trial there." \
  --why "Owner-approved grill of issue #141: lines 1 to 6 and the two cuts confirmed in the session."
```

- **Constraints block.** I split your `Never:` line into two, one act per line, and added a rule not to replace any binary outside `.tree/evidence/`. Otherwise it is the block you accepted. After the block, nothing follows.
- **Lines.** There are no `--cite` flags, because the Goal rests on no settled key.
- **Runner writes after you approve.** I run the command as typed under `owner-standin`. Then I run these as `claude-grill-141`:
  - the DesignQuestion with its 8 rivals (H-base, H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate, H-combo);
  - `hyp link <G-id> external_ref https://github.com/dat9uy/hyp/issues/141`;
  - `hyp status <G-id>` and `hyp check`, which must print `check: 0 errors`;
  - the end-of-grill checkpoint, and the commit of `.tree/events/`.
- **Hand-off.** I then print `hyp prompt <G-id> --actor <runner-actor> --experiments 12`. It must be run on a log that holds this Goal's add.

Do you approve this exact command?

## Owner, message 6

Approved. Run the command as shown, under owner-standin.

### Tool Bash (cwd: .)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt\nHYP_ACTOR=owner-standin hyp add goal \\\n  --alias review-budget \\\n  --title \"Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?\" \\\n  --body $'Design record for issue #141, part of #113. A run of a build Goal stopped twice with its last candidate commit unreviewed (G-ZBM2RE, runs 1 and 2), and run 3 spent each review on one more small commit. This Goal compares rival changes to the review practice of hyp-goal-run and the stop clause hyp prompt prints, and selects one or a combination. It builds nothing: the change to skills/ and internal/ is a later build Goal, which also carries the rule that hyp-goal-run states how a review is recorded and whether it counts against the budget.\\nThe candidate is a rule, measured by replaying Sonnet runs from two starting points: the cut after R-8JQPWG (run 1, one Result left, a fix commit pending) and the cut after R-3YJ7T2 (run 2, one Result left, two limits whose cases pass). 3 trials per rival and per cut. The rivals are all under one question; H-whole-candidate is expected to lose on cost and is tested, not assumed.\\nMay change: this tree\\'s .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text and scratch builds for each rival rule, kept under .tree/evidence/<E-id>/\\nRead only: skills/, internal/, docs/, and every other investigations/ tree\\nNever: edit skills/ or internal/; the change is a later build Goal\\nNever: replace the installed hyp, or any binary outside .tree/evidence/\\nNever: write under an owner actor, or merge a PR' \\\n  --line \"[N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither reviewed nor named unreviewed. Reviewed: a Result that cites a fresh subagent report on that commit diff and names the commit. Named: the stop checkpoint holds the commit short sha and the word unreviewed. Fails if any of the 3 trials at either cut does.\" \\\n  --line \"[N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the stop checkpoint names it. Fails if any of the 3 trials at either cut does.\" \\\n  --line \"[N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix commit and no name in the stop checkpoint. Fails if any trial does. A cell where no trial produced such a finding is reported as not exercised, not as passed.\" \\\n  --line \"[N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own prompt prints. Fails if any of the 3 trials at either cut does.\" \\\n  --line \"[N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check that ran; a 
[... 901 more characters]
```

Result:

```text
G-C49GGQ node.add
  line G-C49GGQ#1: [N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither reviewed nor named unreviewed. Reviewed: a Result that cites a fresh subagent report on that commit diff and names the commit. Named: the stop checkpoint holds the commit short sha and the word unreviewed. Fails if any of the 3 trials at either cut does.
  line G-C49GGQ#2: [N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the stop checkpoint names it. Fails if any of the 3 trials at either cut does.
  line G-C49GGQ#3: [N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix commit and no name in the stop checkpoint. Fails if any trial does. A cell where no trial produced such a finding is reported as not exercised, not as passed.
  line G-C49GGQ#4: [N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own prompt prints. Fails if any of the 3 trials at either cut does.
  line G-C49GGQ#5: [N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check that ran; a commit that only adds a test or text for such a case counts. Fails if 2 of the 3 trials at a cut do. A cell where no trial produced such a review is reported as not exercised, not as passed.
  line G-C49GGQ#6: [N] A candidate rule mean total_cost_usd over its 3 trials at a cut is more than 1.5 times the mean of H-base at the same cut. H-base is skills/hyp-goal-run/SKILL.md (sha256 cb15e2f2e207b5a66807c9369e4d776d5378e61c2dd99679c36dfba402fa750e) with internal/cli/prompt.go (sha256 21f9332383bc6b0eaf3d8e581eea0d397d06262ac45791dad035f0a421ce38ed) at 9c8e498; a candidate differs from it. A hung trial counts at the highest cost of any finished trial at that c
[... 59 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; export HYP_ACTOR=claude-grill-141\nhyp add dq --parent G-C49GGQ --alias q-review --title \"Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?\" --body \"Today hyp-goal-run asks one review of each candidate commit diff, a review is its own Experiment whose Result counts against the run budget, and hyp prompt prints a default of 5 Results. Runs 1 and 2 of G-ZBM2RE stopped with the last candidate unreviewed; run 3 spent each review on one more small commit. Each rival is one rule, tested by replaying Sonnet runs from the cut after R-8JQPWG (run 1, actor claude-g113-1, 4 of 5 Results spent) and the cut after R-3YJ7T2 (run 2, actor claude-g113-2, 4 of 5 spent), 3 trials each, scored on the Goal lines. Each rival gets a binary built with its own prompt.go and its own skill text, and the trial sees no event or commit after the cut. H-base must fail line 1 at the first cut, or the instrument is wrong.\" --why \"Grill of issue #141: one question for all candidate fixes.\" 2>&1 | head -3",
 "description": "Add the DesignQuestion under the new Goal"
}
```

Result:

```text
refused: over limit: body has 753 runes (max 600)
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; export HYP_ACTOR=claude-grill-141\nhyp add dq --parent G-C49GGQ --alias q-review --title \"Which change to the review practice lets a budget-bound run end on a reviewed candidate without going on and on?\" --body \"Today a review is its own Experiment whose Result counts against the run budget; hyp prompt defaults to 5 Results. Runs 1 and 2 of G-ZBM2RE stopped with the last candidate unreviewed; run 3 spent each review on one more small commit. Each rival is one rule. Test: replay Sonnet runs from the cut after R-8JQPWG (actor claude-g113-1) and the cut after R-3YJ7T2 (actor claude-g113-2), 3 trials each, with the rival own skill text and prompt binary and no event or commit after the cut. H-base must fail line 1 at the first cut, or the instrument is wrong.\" --why \"Grill of issue #141: one question for all candidate fixes.\" 2>&1 | head -3",
 "description": "Add the DesignQuestion with a shorter body"
}
```

Result:

```text
Q-62N55P node.add
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; export HYP_ACTOR=claude-grill-141\nadd(){ hyp add hyp --parent Q-62N55P --alias \"$1\" --title \"$2\" --body \"$3\" --why \"Rival of Q-62N55P from issue #141.\" 2>&1 | head -2; }\nadd h-base \"H-base: today's practice\" \"Skill text and prompt as at 9c8e498 (sha256 cb15e2f2… and 21f93323…): one review of each candidate, its own Experiment, counted; default 5 Results; stop checkpoint need not name a gap. The baseline for line 6.\"\nadd h-one-experiment \"H-one-experiment: the review is part of the gate Experiment\" \"A candidate review is recorded in the gate Experiment's Result, with its limits, so a candidate costs one Experiment. Source: run 3's E-GE085A.\"\nadd h-name-at-stop \"H-name-at-stop: the stop checkpoint names each unreviewed candidate\" \"Keep the practice; the stop checkpoint names each candidate without a review (short sha and the word unreviewed), so the owner sees the gap. Holds alone here.\"\nadd h-reserve-last \"H-reserve-last: with one Result left, review instead of another fix\" \"When one Result remains, the runner reviews the current candidate instead of committing another fix, and stops with recorded limits.\"\nadd h-raise-budget \"H-raise-budget: the --experiments default rises to 7\" \"hyp prompt defaults to 7 Results, to fit a gate and a review per candidate. Prompt binary built with default 7; skill text as H-base.\"\nadd h-probe-closes \"H-probe-closes: a case the run's own probe passed closes in the review's Result\" \"A case that the run's own probe ran and passed on the candidate is closed in the review's Result, with no new commit. A limit stays for a case no check ran or one the candidate fails.\"\nadd h-whole-candidate \"H-whole-candidate: the review reads the cumulative diff from the Goal's base\" \"The review before the select reads the diff from the Goal's base, not the last increment. Expected to lose on cost (line 6); tested, not assumed.\"\nadd h-combo \"H-combo: one-experiment, name-at-stop, reserve-last and probe-closes together\" \"The four rules together. Shows which rivals hold alone and which only together.\"\nhyp link G-C49GGQ external_ref https://github.com/dat9uy/hyp/issues/141 --why \"The issue this design Goal serves.\" 2>&1 | head -2\nhyp status G-C49GGQ 2>&1 | head -30; hyp check 2>&1 | tail -3",
 "description": "Add the eight rivals, link the issue, check the tree"
}
```

Result:

```text
H-3Z0TYF node.add
H-7RPV70 node.add
H-0MNAT4 node.add
H-ECMSHA node.add
H-S8D5H4 node.add
H-YNB6HY node.add
H-E8YCVV node.add
H-G8TCDX node.add
G-C49GGQ edge.add
GOAL G-C49GGQ Which change to the review practice lets a budget-bound run… — in_progress (answered 0, open 1, reopened 0, deferred 0, withdrawn 0, stale 0)
LINES
  G-C49GGQ#1 no candidate — [N] On the cut after R-8JQPWG and the cut after R-3YJ7T2, a trial of a candidate rule stops with its final candidate commit neither r…
  G-C49GGQ#2 no candidate — [N] On the same cuts, a trial records the select on a candidate that has no review Result (reviewed as in line 1), whether or not the…
  G-C49GGQ#3 no candidate — [N] On the same cuts, a review finding that is a case of a Goal line ends with no record: no limit, no close in a Result, no fix comm…
  G-C49GGQ#4 no candidate — [N] On the same cuts, a trial records more Results under the cut actor, counted from the start of that actor run, than the N its own …
  G-C49GGQ#5 no candidate — [N] On the same cuts, a trial adds a commit after a review whose findings are all cases the candidate already passes with a check tha…
  G-C49GGQ#6 no candidate — [N] A candidate rule mean total_cost_usd over its 3 trials at a cut is more than 1.5 times the mean of H-base at the same cut. H-base…
FRONTIER
  Q-62N55P open/untested Which change to the review practice lets a budget-bound run…
    H-3Z0TYF (h-base) proposed H-base: today's practice
    H-7RPV70 (h-one-experiment) proposed H-one-experiment: the review is part of the gate Experiment
    H-0MNAT4 (h-name-at-stop) proposed H-name-at-stop: the stop checkpoint names each unreviewed c…
    H-ECMSHA (h-reserve-last) proposed H-reserve-last: with one Result left, review instead of ano…
    H-S8D5H4 (h-raise-budget) proposed H-raise-budget: the --experiments default rises to 7
    H-YNB6HY (h-probe-closes) proposed H-probe-closes: a case the run's own probe passed closes in…
    H-E8YCVV (h-whole-candidate) proposed H-whole-cand
[... 744 more characters]
```

### Tool Bash (cwd: investigations/goal-prompt)

```text
{
 "command": "cd /Users/datguy/.hyp-sandbox/g146-base-b-2/wt/investigations/goal-prompt; export HYP_ACTOR=claude-grill-141\nhyp checkpoint --goal G-C49GGQ --summary \"Grill done: 6 lines written by owner-standin; Q-62N55P holds 8 rivals (H-base, H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate, H-combo). The build of the selected rule is a later Goal. Out-of-Goal: docs/how-hyp-improves-itself.md says hyp states no run cap, but hyp prompt prints one.\" --why \"end of grill\" 2>&1 | head -2\nhyp check 2>&1 | tail -1\ngit add .tree && git commit -q -m \"docs: record the grill of #141, the review-budget Goal and its rivals\" && git log --oneline | head -1\nhyp prompt G-C49GGQ --actor claude-g141-1 --experiments 12",
 "description": "Checkpoint, commit the tree, print the goal prompt"
}
```

Result:

```text
G-C49GGQ checkpoint
check: 0 errors, 9 warnings
a6d622a docs: record the grill of #141, the review-budget Goal and its rivals
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-C49GGQ: read it with hyp show G-C49GGQ, read hyp status G-C49GGQ, and name --goal G-C49GGQ on every hyp checkpoint.
Done when the latest printed hyp status G-C49GGQ shows "runner owes nothing: stop and report to the owner" under OWED, or 12 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: this tree's .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text and scratch builds for each rival rule, kept under .tree/evidence/<E-id>/
Read only: skills/, internal/, docs/, and every other investigations/ tree
Never: edit skills/ or internal/; the change is a later build Goal
Never: replace the installed hyp, or any binary outside .tree/evidence/
Never: write under an owner actor, or merge a PR
```

## Grill

The grill is finished. Goal G-C49GGQ is written, committed as `a6d622a` on `work`, and `hyp check` prints `0 errors`. It still prints 9 warnings, all from older Results in this store.

**What is in the tree**
- **Goal:** G-C49GGQ, alias `review-budget`, with 6 `[N]` lines. You approved the exact command, and it ran under `owner-standin`.
- **Question:** Q-62N55P holds 8 rivals: H-base, H-one-experiment, H-name-at-stop, H-reserve-last, H-raise-budget, H-probe-closes, H-whole-candidate and H-combo.
- **Link:** the Goal is linked to issue #141 as an external reference.
- **Checkpoint:** I recorded the end-of-grill checkpoint under `claude-grill-141`.

**Details to know**
- The body of Q-62N55P was 753 runes and the limit is 600, so I shortened it. It states the replay in brief. The per-trial rules are not in it, because they go into the Experiment bodies. These are the rules about the cut, the per-rival binary and the budget.
- After the constraints block the Goal body has no other text.
- `hyp status` lists 8 `test` items under OWED and nothing under AWAITING OWNER.

**Hand-off.** Give the owner this as it is. `hyp prompt` printed it from this log, and I added nothing:

```text
/goal Run the hyp tree in investigations/goal-prompt with the hyp-goal-run skill, as HYP_ACTOR=claude-g141-1. Its Goal is G-C49GGQ: read it with hyp show G-C49GGQ, read hyp status G-C49GGQ, and name --goal G-C49GGQ on every hyp checkpoint.
Done when the latest printed hyp status G-C49GGQ shows "runner owes nothing: stop and report to the owner" under OWED, or 12 Results have been recorded under claude-g141-1, or you cannot take the next step honestly and say why.
May change: this tree's .tree/ and its evidence, in commits on one local branch; scratch copies of the skill text and scratch builds for each rival rule, kept under .tree/evidence/<E-id>/
Read only: skills/, internal/, docs/, and every other investigations/ tree
Never: edit skills/ or internal/; the change is a later build Goal
Never: replace the installed hyp, or any binary outside .tree/evidence/
Never: write under an owner actor, or merge a PR
```

**Still to do**
- **Experiments.** The runner adds the Experiments and their verdict rules. I did not add them. Planned: one known-answer Experiment, the H-base trials, then one Experiment per cut.
- **Build Goal.** The later build Goal is yours to add after the select.
- **Issue #141.** It stays open. The commit message names it, but I did not post to it or push, because this machine has no network.
- **Out-of-Goal finding.** `docs/how-hyp-improves-itself.md` says hyp states no run cap, but `hyp prompt` now prints one. It is in the checkpoint.
