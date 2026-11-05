---
status: accepted
---

# The tree machine owns its words and its rules

The tree machine (`internal/replay`, `State.Step`) existed, but its status words were one untyped string and its rules were spread across switches. `Projection.Status` held the status of every node kind, so a Hypothesis word could be compared with an Experiment word and it compiled: `superseded` and `withdrawn` are each words of two kinds. Nothing stopped a replay rule from reading derived status, which spec §5.1 forbids. The owner accepted this in the architecture review of 2026-11-05 (#45). There is no tree, because no choice below is measurable, and no output changes.

**Each node kind has its own typed status word and derived record.** `HypStatus`, `QuestionState`, `ExpStatus`, `ResStatus` and `GoalState` are separate string types, so comparing two kinds does not compile. `Projection.Hyps` and `Projection.Questions` join `Exps` and `Results`, and `Projection.Status` is gone. A Question reaches its `Decision` by the shared id: `Decision` holds what was stored and `Question` what was derived, so the hold state stays out of `Decision`. One display-only method, `Projection.Display`, serves code that prints any node without knowing its kind.

## Consequences

- Reasons, stale reasons and limit states stay plain words. `StaleCause.Reason` holds either a stale reason or a Hypothesis status, and `GoalStatus.Counts` is keyed by question states plus `stale`, so both stay strings and their readers convert.
- Tests changed only in how they read or type a word. No expected value changed.
- These words do not make statuses machines. A Hypothesis or Question status is computed afresh from stored facts on every replay and has no transitions of its own to guard, because a merged log can hold any order of events (ADR 0011). Only the tree as a whole, replaying events, and the sandbox (ADR 0012) are machines.
- What may still happen to a closed node is a separate question (#43), and this decision does not touch it.

## Rejected

- **Typed constants on one field.** One field cannot hold two word types without an interface or casts at every read, and the kinds would still share a record that a Result has no use for.
- **A status interface.** It gives nothing `Display` does not. A method per kind on a shared interface moves the kind check into a type switch at each reader, where it can be forgotten.
