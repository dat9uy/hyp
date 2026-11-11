# hyp

hyp keeps an agent's investigation as a hypothesis tree: append-only events, statuses derived from them, and skills that any agent harness able to load a `SKILL.md` can use.

## The problem

An agent that investigates a design question usually keeps its account in prose: notes, plans, reports. The account grows until nobody reads it. It loses which hypothesis started the work, which experiment ruled out which answer, and why the answer changed. A new session cannot tell what is decided and what is still open, so it starts again or trusts a summary it cannot check.

## How hyp works

- **Rival answers, side by side.** Each design question gets at least two rival Hypotheses as siblings. A refinement is a new sibling that supersedes the old one, so the lineage stays visible.
- **Experiments with rules written first.** An Experiment states its verdict rule before it runs, and can run from a versioned, hash-pinned recipe, so the same procedure can be run again and checked.
- **Statuses are derived, never typed.** Every write is an event in an append-only log under `.tree/`, committed with the work. `supported`, `contradicted`, `stale` and `answered` are computed from those events; no command can set one, and a changed mind is a new event, never an edit.
- **The Goal's lines are the owner's.** The owner writes the cases every candidate must meet once, as the Goal's lines. A line is owed on each question until a Result that ran its case covers every candidate there, and `hyp status` shows each line as covered or uncovered.
- **A limit is the next case to run.** A Result records what its run did not cover. A later Result closes it by running the case, and a case no test within the Goal can run goes to the owner. A requirement the Goal lacks is a new Goal, so a new tree that cites this one.
- **Every decision is owed first.** hyp admits a select only when its ledger owes it, so a question is answered only once the open work on it is done.

### Who does what

The ledger asks, the owner answers, the runner runs, and hyp derives.

| | Does | Never does |
|---|---|---|
| **Owner** | Writes the Goal's lines once. Answers what `hyp status` lists under AWAITING OWNER: what a line means, whether a referred case runs, which of the candidates that meet every line to choose. | Runs a case, types a verdict or edits what was recorded |
| **Runner** (the agent) | Settles what OWED lists for it, runs Experiments, records verdicts, and stops when it owes nothing. | Answers for the owner or writes a line |
| **hyp** | Refuses a write that would corrupt the tree or skip a step, and derives every status and what is owed. | Judges whether an Experiment was well designed |

[principles.md](docs/principles.md) states the principles hyp holds to, among them the owner's role. [how-hyp-improves-itself.md](docs/how-hyp-improves-itself.md) explains why the guards run during an investigation instead of before it, and how a change to hyp earns its way in.

## Where hyp comes from

hyp's way of running an Experiment starts from pstack's [`create-verification-skill`](https://github.com/cursor/plugins/blob/ae5650febf4320793844c12c4adaa937f417f8da/harness/pstack/claude-code/skills/create-verification-skill/SKILL.md). That skill writes a project-local verify skill whose sections include Launch, Doctor, Drive, Evidence and Cleanup, with honesty rules: evidence survives cleanup, a step proven only in a stand-in for the real target stays unverified, a journey is never weakened to make it pass, and a skill that was never run is a draft. hyp keeps the five names as the sections of a recipe (see "Step" in [glossary.md](docs/glossary.md) and [sandbox-spec.md §1](docs/sandbox-spec.md#1-steps-and-commands)) and adds:

- **Versioned recipes.** A recipe is a bundle of files with a version, pinned by its hash and never edited once a tree pins it ([ADR 0005](docs/adr/0005-a-bundle-is-its-ref-and-hash.md), [RECIPE-FORMAT.md](skills/hyp-verification/RECIPE-FORMAT.md)).
- **hyp runs the commands.** hyp, not the agent, runs every recipe command and records what it ran ([ADR 0001](docs/adr/0001-hyp-runs-every-recipe-command.md)).
- **Known answers first.** Before a recipe runs, its checks must give their known answers, and they run the same way in that gate as in the run ([ADR 0002](docs/adr/0002-known-answers-gate-every-recipe-run.md), [ADR 0003](docs/adr/0003-a-check-runs-the-same-way-in-the-gate-and-in-the-run.md)).
- **A verdict rule written first.** Each Experiment states, before it runs, which outcome gives which verdict ("Verdict rule" in [glossary.md](docs/glossary.md)).
- **Rivals and a gate.** Each Experiment tests rival Hypotheses in a tree, and a select is admitted only when the ledger owes it ("Rival", "Select" and "Owed" in [glossary.md](docs/glossary.md)).

## An example

hyp's own design is decided this way. The tree `investigations/id-citation` settled which rule stops a write command from appending text that cites an id that is no node. Its one Result, R-BV4CWV, supported three rivals and contradicted the fourth, and the owner selected among the three as a trade-off. `hyp status` in that directory:

```text
GOAL G-4S36M5 A write command never appends text that cites a node id thi… — complete (answered 1, open 0, reopened 0, deferred 0, withdrawn 0, stale 0)
FRONTIER
STALE
OWED
  runner owes nothing: stop and report to the owner
AWAITING OWNER
RECENT
  2026-11-01 01:51 owner-dat confirm Q-DVZZHY — Owner: sufficient, accept limits #1-#5 of R-BV4CWV
  2026-11-01 01:16 verifier-240-1 confirm Q-DVZZHY — R-BV4CWV: text-refuse exits 2, appends nothing on K1-K6, passes G1-G5; warn and…
  2026-11-01 01:15 claude-240-1 dq.select Q-DVZZHY — Owner chose h-text-refuse over h-text-warn and h-cite after kongming's counsel:…
  2026-11-01 00:26 claude-240-1 result.record R-BV4CWV — e-arms' pre-registered rule (8e5bc68), read from id-citation-arms@1 verdict
CHECKPOINT 2026-11-01 01:51 claude-240-1: Q-DVZZHY answered: h-text-refuse (H-88ZF3E) selected as the owner's trade-off on R-BV4CWV, verifier and owner sufficient. Next: port the rule into internal/cli (exit 2, write-time only), implementing… (full: hyp show G-4S36M5)
```

The goal is complete and nothing is open. `hyp show Q-DVZZHY` shows how the question was answered (lines selected from its output):

```text
DesignQuestion Q-DVZZHY
title: Which rule stops a write command from appending text that cites an id that is no node?
status: answered
selected: H-88ZF3E (trade-off) — Owner chose h-text-refuse over h-text-warn and h-cite after kongming's counsel: R-BV4CWV supports all three; only refuse keeps the slip out of the log. Check stays write-time, exit 2.
owner: sufficient — Owner: sufficient, accept limits #1-#5 of R-BV4CWV
experiments:
  E-YKJ70D tests H-88ZF3E, H-RAJAFK, H-C5BA8J, H-BMN916
    judged: H-88ZF3E=supports, H-RAJAFK=supports, H-C5BA8J=supports, H-BMN916=contradicts
children:
  H-RAJAFK supported Scan text flags for id-shaped tokens; warn on an unresolved…
  H-88ZF3E supported Scan text flags for id-shaped tokens; refuse an unresolved …
  H-C5BA8J supported request-verify, select and confirm take --result REF, resol…
  H-BMN916 contradicted No change: today's hyp is enough
```

Four rivals were written down before anything ran, including a null rival that said no change was needed. One Experiment tested all four and contradicted only the null. Because the evidence could not choose among the other three, the select is marked as a trade-off the owner made, and the owner's choice ended it.

The tree was recorded with an earlier hyp, and its log keeps what that hyp wrote: a verifier confirmed every select, and the owner's confirm accepted the Result's limits. Both are history. Today the select is the gate, with no verifier ([ADR 0015](docs/adr/0015-the-select-is-the-gate.md)), and the owner accepts a limit only after the runner refers it. A log is never rewritten, so the same events replay the same way.

## Get started

Claude Code is the only harness hyp has been verified on.

**1. Install the binary** (needs Go):

```bash
go install github.com/dat9uy/hyp/cmd/hyp@latest
hyp version
```

**2. Copy the two skills into your project** and commit them, so every session in the project loads them. From your project's root, with a clone (or download) of hyp at `<hyp>`:

```bash
mkdir -p .claude/skills
cp -R <hyp>/skills/hyp-prototype <hyp>/skills/hyp-verification .claude/skills/
git add .claude/skills
```

To update, re-run the `go install` line and copy the skills again, so the binary and the skills stay the same version.

**3. Start a session** in your project and paste this prompt, with the goal and its lines filled in:

```text
Use the hyp-prototype skill. Set HYP_ACTOR=claude-<topic>-1 for this session.
If `hyp status` says there is no .tree store, run `hyp init` at the repository root.
Record this goal with its lines under my actor name <your-name>, showing me
the command first: <the outcome I want, and who needs it>.
Lines: <each case every candidate must meet>.
Add the first design question under it, with at least two rival hypotheses,
one of them "no change needed". Do not add or run experiments yet.
Show me `hyp status` and stop.
```

**4. The lines are yours.** They are the cases every candidate must meet. They come with the Goal and are never added to later, so take your time over them. The session shows you the `hyp add goal` command with your lines and runs it under your actor name once you approve:

```bash
hyp add goal --title "<the outcome>" --actor <your-name> --why "<who needs it>" \
  --line "<a case every candidate must meet>" --line "<another>"
```

Your answers later are `hyp answer` events under the same name, for example `hyp answer --answer "R-…#r1=<your words>"`.

From there the [hyp-prototype skill](skills/hyp-prototype/SKILL.md) guides the next rounds: experiments, verdicts, selects and checks. When `hyp status` lists something under AWAITING OWNER, it is yours to answer. [example.sh](skills/hyp-prototype/example.sh) runs a complete investigation in a throwaway directory.

To work on hyp itself, see [install.md, "Develop hyp"](docs/install.md#develop-hyp).

## What is in this repository

What you install is **core**: the binary (`cmd/`, `internal/`), the two skills (`skills/`) and the docs (`docs/`). The **self-review** recipes this repository keeps to test hyp, and the trees in which parts of hyp's own design were decided, are kept under `recipes/` and `investigations/`. You need none of them to use hyp. [core-and-self-review.md](docs/core-and-self-review.md) maps every item to its scope.

## Status

hyp is experimental (version 0.1.0, event format v1). One author builds and tests it. Claude Code is the only harness it has been verified on; Codex and pi are untested. The evidence that it works is the trees under `investigations/`, each a recorded investigation with its verdicts and owner choices. Records made before the public release stay in the private development history. hyp itself changes only when a real investigation shows a failure, and a change to the skills must pass a replay of the session where the failure happened.

## Contributing

Read [AGENTS.md](AGENTS.md). It explains the two scopes, how hyp is improved with hyp, and how to build and test.
