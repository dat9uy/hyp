# The principles hyp holds to

This page states why hyp is shaped the way it is. Each section is one claim
and a few sentences of why. No rule is stated only here: the rules are in the
binary, and `hyp spec` prints each one. When this page and `hyp spec` differ,
`hyp spec` wins.

## 1. What hyp is for

**A change is admitted only when evidence shows that it meets the owner's
Goal.** The owner states the Goal as words and lines. The grill places examples
that can fail. The runner builds Hypotheses, the subjects of its commits, and
runs the lines on them. A
later reader, such as the owner, a new session or another agent, reads the log
and the PR body, not a chat. They find in the log what each example printed on
each build, and in the PR body which build was selected.

**Learning is why anything is kept.** Three things are learned, by three
parties. An agent finds a fact: a finding. A Goal names the findings that bear
on its want. The owner learns a want from a consequence and states it as a
line, a decline or a "not yet". A record that serves none of these is not kept.

## 2. The log is the truth; everything else is calculated

**Events are appended and never edited, and every state is calculated from
them.** Each event is a file of its own under `.tree/events/<actor>/`. hyp stores no
table: the Goal states, the rows, the candidates and the stops are calculated
on every command. No command sets a state. A changed mind is a
new event, for example a new version of an example, so the earlier event stays
in the history. Two clones merge their logs with a plain git merge, since no act
changes a file another wrote, and replay orders the events by id.

The log keeps what the owner learned or approved: the acts that change a
Goal, such as `add goal`, `select`, `land` and `close`, and one place of each
example set. What agents learned is the trace, outside the repository: the
Results, their outputs, the findings and what hyp printed to each actor. The
trace is a cache. It may be lost with its clone, and a reader that lacks a cell
runs the example again. `hyp show` prints what hyp printed, byte for byte.

**A Goal's state never comes from another system's history.** An act may read
git, the file system or the clock. What it reads is the input of that one act,
and the event it writes carries what a later state needs, as the value or the
hash of the bytes. Replay calculates a Goal's state from events alone, so a
rebase, a squash or a new clone changes none. A write is a step; a read of the
state is a print.

## 3. H0 and the lines are the prior; an example is evidence only if it can fail

**An example counts only when H0 gives the verdict its line needs.** H0 is the
Goal's base commit. An example agrees with its line on H0 when a gap example
fails there and a guard example is met there. `place` runs the example on H0 and refuses it otherwise (the guard
Kind). So no check that cannot fail
enters the Goal. The compare is literal: floors the output must hold and
ceilings it must not hold. No reader judges a verdict.

## 4. The select is the gate

**`select` records only a candidate: a leaf build that meets every line under
the current examples.** With one candidate the runner may select it. With two or
more, only the owner may select. The selected commit must hold a spec
(the guards Candidate and Spec). The report of the
select is the PR body, and the owner's merge accepts it
([ADR 0015](adr/0015-the-select-is-the-gate.md) records why the select is the
gate).

## 5. A budget stops the loop, and nothing scores a round

**A Goal gets three Hypotheses beside H0, and then it stops for the owner.**
A Hypothesis is one build, counted by its subject hash. A refused run costs nothing.
A rerun of a Hypothesis already counted costs nothing. Stable refuses a run whose verdict differs from an earlier run on
the same bytes, so a rerun cannot change a verdict. A placement that changes an
example moves the boundary, because the runner then runs against new evidence.
At a stop, hyp prints the outputs table, not a reading of it: no distance and no
summary.

## 6. Who does what

**The owner decides, the grill places, the runner runs, and hyp calculates.**
The four roles are `owner`, `grill`, `probe-writer` and `runner`.

1. **The owner records the Goal.** Its words and lines are frozen once it opens.
   A new line or a new purpose is a new Goal, which can cite the old one with
   `--from`.
2. **The owner chooses between candidates.** The owner may select at any time,
   and with two or more candidates only the owner selects. Only the actor that
   added the Goal selects for the owner (the guard Goal's actor).
3. **The owner ends a Goal.** At a stop the owner lands the Goal, at any stop
   or none the owner may close it, and the owner may unselect a selected Goal
   so that the grill can place again.
4. **The owner merges.** hyp writes nothing to GitHub. The merge is where the
   owner accepts a build and vetoes its mechanics.
5. **Only the grill places an example** (the guard Grill places). The grill asks
   the owner product questions only, and reports each decision that no example
   checks.
6. **The runner runs lines and selects one candidate.** It never places. It
   runs only the lines that hyp prints for it.
7. **The probe writer writes no event.**
8. **hyp does not authenticate a role.** The actor and the role are
   self-declared. A guard that reads them catches an honest mistake, not a lie.

The owner's guide is [owner-guide.md](owner-guide.md).

## 7. Cheap errors come from hyp, not from the runner's judgment

**A failure that hyp can see goes into the binary as a refusal or a warning.**
Skill text teaches only what hyp cannot check. Every refusal names its guard
and exits 1.
hyp prints the allowed acts of each role as ready command lines. An agent runs
a line that hyp printed, and a skill names no write command. When a fix would
ask the runner to judge in prose what hyp could see, the gap is in hyp.
A refusal holds for every runner and every model. Skill text works only when
an agent loads it and follows it.

## 8. hyp writes no git state in the host

**hyp reads git, and the owner and the agents commit.** hyp writes no refs,
objects, config, index or hooks, and it hides no file from git. It runs only a
fixed set of read commands.

## 9. hyp changes only on a real event

**A change is proposed when a real event points at hyp, and it earns its way in
through a gate.** [AGENTS.md](../AGENTS.md) states the gate.

**hyp guards the loop while it runs, and does not prove it in advance.** A
study run before a real Goal cannot show that agents behave well: an agent
with the skill loaded seldom makes the errors that such a study counts. The
guards cost nothing when a run goes well, and they stop a run that goes
wrong: the refusals and Stable, the budget and its stops, and the replay of a
merged log. A run that reaches a select with no refusal
needs no further study.

## 10. What hyp does not claim

- It does not authenticate actors.
- It does not judge whether an example is well chosen. The grill reports each
  decision of the owner that no example checks.
- A select is not proof that the design is right. It shows that one build met
  every example. A gap found after a select goes back to the grill after the
  owner's `unselect`; a gap that needs a new line is a new Goal.
- It does not sandbox a Probe. A Probe is the grill's own committed code.
- `hyp try` shows a verdict without a record, and a run from a shell outside
  hyp leaves no record. Neither is evidence in the log.
