# How hyp improves itself

hyp is changed the way it asks its users to decide things: from a real event,
through a gate. This page explains where a fix goes: the binary or the skill
text. It explains why the loop is guarded while it runs, not proven in advance.
It also says when a change earns its way in. The loop is a Goal with gap and guard lines,
examples placed on each line, rival Hypotheses run against the examples, and
a select. Its main risk is runs that teach nothing. It is the change-control part
of [principles.md](principles.md), which states hyp's other principles.

## Guard the loop while it runs, not before

Studies run before a real investigation tried to show that agents choose good
experiments. With the skill loaded, agents rarely made the mechanical mistakes
those studies could count. So a skill change had nothing left to remove, and the
studies could not decide anything.

So the loop is guarded during the run, by mechanisms that cost nothing when the
run goes well. [state-machine.md](state-machine.md) owns each rule:

- **Write-time refusals.** Every act checks its guards and refuses with exit 1,
  naming the first guard that fails ([spec §7](state-machine.md#7-guards)). For
  example, Kind refuses an example that gives the wrong verdict on H0, and Leaf
  only refuses a run on a commit that another build descends from.
- **The budget.** A Goal gets three Hypotheses beside H0, one per subject hash.
  Then it stops for the owner. A refused run costs nothing, and a rerun of
  a counted Hypothesis costs nothing ([spec §11](state-machine.md#11-budget-and-stops)).
- **Stable.** Two runs on the same bytes give one verdict. A rerun that gives
  another verdict is refused, and `hyp status` lists it under `unstable:`. So a
  rerun cannot buy a different verdict.
- **Replay.** An event whose guards fail in a merged log is skipped and listed
  under `refused at replay:`. It is never deleted
  ([spec §4](state-machine.md#4-merge-order-crash-and-merged-log)).
- **The select is the gate.** `hyp select` records only a candidate, a leaf
  build met on every line, and the selected commit must hold a spec. With two
  candidates, only the owner selects
  ([ADR 0015](adr/0015-the-select-is-the-gate.md)). What the owner decides is
  in [principles.md](principles.md#6-who-does-what).
- **The trace.** `hyp show <actor>` prints what hyp printed to an actor, byte
  for byte. So the owner can read what the runner saw.

A run that reaches a select with no refusal it had to work around is a success.
It needs no further study.

## Mechanical failures belong in the binary

When a failure can be detected from the log or from git, fix it in the binary
as a refusal or a warning, not as a sentence in a `SKILL.md`. A refusal holds
for every runner and model. Skill text works only if it was loaded and the agent
follows it. Once the binary blocks a failure, a skill-text change aimed at the
same failure has no effect left to show.

Skill text teaches what the binary cannot check. The
[hyp-grill skill](../skills/hyp-grill/SKILL.md) teaches how to ask the owner
product questions and how to write examples that can fail. Its
[run rules](../skills/hyp-grill/run-rules.md) list what a runner must do that
hyp cannot see, such as keeping every commit it ran.

## Change hyp only when a real event points at it

Do not change hyp or a skill on speculation. A change starts from a real run.
In that run a refusal or a warning fires, a Goal reaches a stop it cannot leave,
`hyp status` lists an event under `refused at replay:`, or the owner flags it.

A change to a skill is a build Goal like any change to code: its lines,
examples whose Probes the Goal writes, and the select. Each example gives its
line's verdict on H0 before it is placed. How a skill is tested is that Goal's
choice. hyp keeps no example or Probe for skill text as such
([ADR 0021](adr/0021-a-skill-change-is-a-build-goal.md)).

## Nothing scores a round

At a stop for the owner, hyp prints the outputs table: the verdict, exit code
and output of each example on each build. It prints no distance and no summary
line ([spec §11](state-machine.md#11-budget-and-stops)). The owner reads the
table and decides.

Nothing in hyp judges whether an example was well chosen. The grill reports
each decision of the owner that no example checks. Measuring how good an
example is needs the owner's judgment, and hyp has no check for it.
