# Why the hyp loop is guarded at run time

The loop is: question, rivals, experiment, record, revise, select, verify. Its
main risk is spending tokens on rounds that teach nothing. This page explains
why hyp handles that risk while the loop runs instead of proving beforehand
that it can't happen, and where each guard lives.

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
  without saying so, or a verifier confirming its own select. The list is
  owned by [replay-spec.md](replay-spec.md) §5.
- **Waste signals.** `hyp signals` computes them from the event log
  (`internal/signals/signals.go`). The tripwire policy, which warns first and
  then stops, and the round cap lived in the session-replay recipe's
  `tripwire.py`. Records made before the public release stay in the private
  development history.
- **An independent verifier on every select**, and the owner on trade-offs
  ([VERIFIER.md](../skills/hyp-prototype/VERIFIER.md)).

A run that reaches a verified decision with no tripwire events is a success.
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
investigation trips a signal or the owner flags a round. The change ships only
when a session replay from that event's cut shows the failure in 3 or more of 5 trials
without the change and in 1 or fewer of 5 with it. If no current cut
reproduces a failure, nothing ships for it.

One exception: a correction that makes skill text agree with an owner
decision already recorded (in the tree, an issue or a decision record) is
not a behaviour change and ships without a session replay. Its commit names the
decision. Text that adds a new behaviour is still gated.

## Signals detect stalls, not quality

Judge-free signals count over the verdicts the runner wrote. They are good at
spotting a loop that has stopped moving, such as two rounds in a row that
moved no rival. They can't tell whether an experiment was well designed:

- Status moves only forward (`proposed`, then `inconclusive` or `supported`,
  then `contradicted`; see [replay-spec.md](replay-spec.md) §7.1). So whether
  a round "moved" anything depends on earlier rounds. A replication that
  supports an already-supported rival changes nothing.
- A cautious runner and a bold one can record the same evidence as
  `inconclusive` and `supports`. The count follows the label, not the
  evidence.

Read signals as stop-and-look prompts for the loop, never as a quality score
for single rounds. Measuring experiment quality needs judgment of the
evidence, from the owner or a judge calibrated against the owner's labels. No
such instrument is part of hyp.
