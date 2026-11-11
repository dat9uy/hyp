# How hyp improves itself

hyp is changed the way it asks its users to decide things: from a real event,
through a gate. This page explains where a fix goes, the binary or the skill
text, why the loop is guarded while it runs instead of proven in advance, and
when a change earns its way in. The loop is: question, rivals, experiment,
record, revise, select, and its main risk is spending tokens on rounds that
teach nothing. It is the change-control part of [principles.md](principles.md),
which states hyp's other principles.

## Guard the loop while it runs, not before

Studies run before a real investigation tried to show that agents choose good
experiments: charting the tree first, a design checklist, session replay comparisons
of skill text. With the skill loaded, agents rarely made the mechanical
mistakes those studies could count, so a skill change had nothing left to
remove and the studies could not decide anything.

So the loop is guarded during the run, by mechanisms that cost nothing when
the run goes well:

- **Write-time refusals.** The binary refuses events that would corrupt the
  tree or skip a step, for example an experiment that ignores live rivals
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
  and what they never do, is in [principles.md](principles.md#7-who-does-what).

A run that reaches a select the ledger owes, with no refusal it had to work around, is a success.
It needs no further study.

## Mechanical failures belong in the binary

When a failure can be detected from the tree, fix it in the binary as a
refusal or a warning, not as a sentence in `SKILL.md`. A refusal holds for
every runner and model, while skill text only works if it was loaded and the
agent follows it. Once the binary blocks a failure, a skill-text change aimed
at the same failure has no effect left to show.

Skill text ([hyp-prototype SKILL.md](../skills/hyp-prototype/SKILL.md)) teaches
what the binary can't check: intent, such as choosing the experiment that
would exclude a rival, and honest recording.

## Change hyp only when a real event points at it

Don't change hyp or a skill on speculation. A change is proposed when a real
investigation trips a refusal, a warning, a design mark or an item the ledger cannot settle, or the owner flags a round.

A change to a skill is a build Goal like any change to code: its lines, its
own test or an instrument the Goal builds, red on base, and the select. How a
skill is tested is that Goal's choice; no instrument is kept for skill text as
such.

## The ledger shows a stall; nothing scores a round

A stall is work the tree owes that does not get settled. The ledger shows it
by name: an item, who owes it, and the Experiment that aimed at it and did not
settle it. hyp once also judged rounds by whether a Result changed a status.
That account disagreed with the ledger: statuses move only forward, so a
candidate that holds every new case reads the same after each round, though
each round closed a case. ADR 0017 removed it.

Nothing in hyp judges whether an experiment was well designed:

- A cautious runner and a bold one can record the same evidence as
  `inconclusive` and `supports`. A count follows the label, not the
  evidence.

Measuring experiment quality needs judgment of the evidence, from the owner or
a judge calibrated against the owner's labels. No such instrument is part of
hyp.
