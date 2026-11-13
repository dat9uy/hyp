# Nodes the grill added

Stores at the end: ., internal/testutil/testdata/synth, investigations/checkpoint-store, investigations/checkpoint-store-design, investigations/id-citation, investigations/recipe-base

## Goal in ., by owner-standin: hyp prints the /goal prompt for any Goal from the log, reading a run's constraints from the Goal body

Body:

```text
Builds #113: hyp prints the /goal prompt for any Goal from the log (hyp prompt), and reads a run's constraints from the Goal body's headings. Owner decision 2026-11-11: no typed field until a real run misreads or loses a constraint. The candidate is a hyp commit; go test is the instrument. The change also fixes the texts it makes wrong: the hyp-owner-grill hand-off, the word discovery in the descriptions of hyp-goal-run and hyp-prototype, the usage text, the specs, the glossary and the README.
Out by design: a Goal written before the headings, such as checkpoint-store:G-D1CH1R, has no block. hyp prompt refuses it with exit 1 and names what is missing; adding such a Goal stays allowed.
Not here: matching run 6's prompt (goal-prompt-261111-2129-g1b-run-6) clause for clause. The owner must supply its text first; a later Goal, added by the owner, carries that comparison by sha256. Differences it allows: the actor, the owner's answer the run starts from, the branch, the PR, the commit and the run's own plan. Branch and PR are flags. What a commit-candidate Goal needs from hyp is #114.
May change: this repository, as commits on one local branch, in cmd/, internal/, skills/ (the hyp-owner-grill hand-off, the descriptions of hyp-goal-run and hyp-prototype), docs/, README.md and AGENTS.md step 3, where it says where trees live; a PR may be opened and is never merged
Read only: investigations/ and recipes/, and every store under them
Never: replace the installed hyp binary
Never: edit another tree, a recorded event or this Goal
Never: add a Goal or a line
Never: write under the owner's actor
```

Lines:

1. [P] hyp prompt G --actor A on a Goal with a valid constraints block exits 0 and prints, in this order: the absolute store path, A, the resolved G- id, hyp-goal-run, three stop clauses (hyp status shows the runner owes nothing; N Experiments recorded under A; the run cannot take the next step honestly and says why), then the block's lines byte for byte. Two runs on one log print identical bytes. In a store with many Goals it names hyp status <G-id> and hyp checkpoint --goal <G-id>.
2. [P] --experiments N (default 5) changes only the number in the count clause. --branch B and --pr N each add exactly one clause after the block and change nothing before it. The PR clause says the PR applies only if May change allows it.
3. [N] A Goal body without a valid block makes hyp prompt exit 1 and name the first failing line. The block runs from the first column-0, case-sensitive line starting May change:, Read only: or Never: to the end of the body; it holds each heading at least once; every non-empty line starts with one. A line of spaces, a tab or a lone CR is not empty. Lines split on \n only, one trailing \r is ignored for the test, and one trailing blank at the very end is allowed.
4. [N] The Experiments cap counts only non-withdrawn Experiments added under the named Goal by actor A. Another Goal's Experiments and another actor's are never counted. An actor that wrote the Goal's add is refused with exit 1.
5. [N] A non-Goal id, an unknown id or a complete Goal makes hyp prompt exit 1 and name the cause. An unknown id names the store searched and says the Goal's add must be in the log. An alias never prints.
6. [N] A bad actor name, a --branch that is not a valid git ref name, or a non-positive --pr or --experiments exits 2. No store above the current directory exits 3. A prompt over 4000 characters exits 1 and says by how many.
7. [N] hyp add goal with a body that has a heading but no valid block exits 0 and warns. A valid block, or a body with no heading, adds silently. Replay accepts both logs.
8. [N] hyp prompt appends no event, writes no file and stores no copy of the block: the store is byte-identical before and after.

## DesignQuestion in ., by claude-grill-113: How does hyp prompt find a Goal body's constraints block?

Scenario that separates the rivals: a body with a prose line that starts with Never:, more prose, then a real block of May change:, Read only: and Never: lines. Decided by argument in the grill session; the run builds the winner.

## Hypothesis in ., by claude-grill-113: The block runs from the first heading line to the end of the body

From the first column-0 line starting May change:, Read only: or Never: to the end of the body. Every non-empty line in it starts with a heading and all three headings appear. The scenario body is refused with exit 1 and the first failing line is named.

## Hypothesis in ., by claude-grill-113: The block is the trailing run of heading lines

The block is the last run of consecutive heading lines at the end of the body; prose before it is ignored. The scenario body is accepted and the real block is printed.

## Experiment in ., by claude-grill-113: Scenarios that separate the two block rules, argued in the grill

Rule, fixed before the Result: S1 = a body with a prose line starting Never:, a line of prose, then a block of May change:, Read only: and Never: lines. A refuses S1 (exit 1) and names the first failing line, which is the prose line after the first Never:. B accepts S1 and prints the trailing block. S2 = a body that is only a valid block: both accept. S3 = a block followed by a prose line: A refuses, B accepts. The owner chose A in the session because a heading-like line in prose must not pass unnoticed. Verdicts: A supports on S1 and S3; B contradicts. Evidence: the scenarios file.

## Goal in ., by owner-standin: How hyp prompt finds a Goal body's constraints block (design record for G-NSXQ3E)

Design record for the build Goal G-NSXQ3E (#113): how hyp prompt finds a Goal body's constraints block. The owner settled it by argument in the grill session. The DesignQuestion, both rival Hypotheses, the argued Experiment and its Result sit under this Goal, and the select closes it. The build Goal carries the lines that check the chosen rule; this Goal has none, because no run checks its candidates.
May change: this store, under this Goal only: the DesignQuestion, the rival Hypotheses, one Experiment, its Result, the select and the end checkpoint, written by the runner as events and one evidence file under .tree/evidence/
Read only: everything else, G-NSXQ3E included
Never: edit G-NSXQ3E, add a Goal or a line
Never: write under the owner's actor

## DesignQuestion in ., by claude-grill-113: How does hyp prompt find a Goal body's constraints block?

Scenario that separates the rivals: a body with a prose line that starts with Never:, more prose, then a real block of May change:, Read only: and Never: lines. Decided by argument in the grill session; the build Goal G-NSXQ3E builds the winner.

## Hypothesis in ., by claude-grill-113: The block runs from the first heading line to the end of the body

From the first column-0 line starting May change:, Read only: or Never: to the end of the body. Every non-empty line in it starts with a heading and all three headings appear. The scenario body is refused with exit 1 and the first failing line is named.

## Hypothesis in ., by claude-grill-113: The block is the trailing run of heading lines

The block is the last run of consecutive heading lines at the end of the body; prose before it is ignored. The scenario body is accepted and the real block is printed.

## Experiment in ., by claude-grill-113: Scenarios that separate the two block rules, argued in the grill

Rule, fixed before the Result: S1 = a body with a prose line starting Never:, a line of prose, then a block of May change:, Read only: and Never: lines. A refuses S1 (exit 1) and names the first failing line, which is the prose line after the first Never:. B accepts S1 and prints the trailing block. S2 = a body that is only a valid block: both accept. S3 = a block followed by a prose line: A refuses, B accepts. The owner chose A in the session because a heading-like line in prose must not pass unnoticed. Verdicts: A supports on S1 and S3; B contradicts. Evidence: the scenarios file.
