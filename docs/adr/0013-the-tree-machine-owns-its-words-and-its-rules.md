---
status: accepted
---

# The tree machine owns its words and its rules

The tree machine (`internal/replay`, `State.Step`) existed, but its status words were one untyped string and its rules were spread across switches. `Projection.Status` held the status of every node kind, so a Hypothesis word could be compared with an Experiment word and it compiled: `superseded` and `withdrawn` are each words of two kinds. Nothing stopped a replay rule from reading derived status, which spec §5.1 forbids. The owner accepted this in the architecture review of 2026-11-05 (#45). There is no tree, because no choice below is measurable, and no output changes.

**Each node kind has its own typed status word and derived record.** `HypStatus`, `QuestionState`, `ExpStatus`, `ResStatus` and `GoalState` are separate string types, so comparing two kinds does not compile. `Projection.Hyps` and `Projection.Questions` join `Exps` and `Results`, and `Projection.Status` is gone. A Question reaches its `Decision` by the shared id: `Decision` holds what was stored and `Question` what was derived, so the hold state stays out of `Decision`. One display-only method, `Projection.Display`, serves code that prints any node without knowing its kind.

**Each event type has one definition, and a replay rule can read stored facts only.** `internal/replay/rules.go` holds a table from event type to its replay rule, its apply and its basis nodes. `State` embeds `Facts`, which holds what events stored and the lookup indexes over it, and every replay rule takes `*Facts`, so reading a derived status from one does not compile. That is the permanent reason for the split: spec §5.1 says a merged log may hold any order of events, so a rule that read derived status could refuse a log depending on how it merged.

## Consequences

- Admission stays out of the table. Its rules read derived status, and that is why they can only warn or refuse at write time. The split between the two is permanent.
- `Snaps`, `EverAnswered` and `RivalsAtAdd` are stored, but they record derived status at an event, so they stay out of `Facts`. A test fails when `Facts` gains a field outside its allowed set.
- `apply` takes `*State`, because `node.add` stores the live rivals at the add.

- Reasons, stale reasons and limit states stay plain words. `StaleCause.Reason` holds either a stale reason or a Hypothesis status, and `GoalStatus.Counts` is keyed by question states plus `stale`, so both stay strings and their readers convert.
- Tests changed only in how they read or type a word. No expected value changed.
- These words do not make statuses machines. A Hypothesis or Question status is computed afresh from stored facts on every replay and has no transitions of its own to guard, because a merged log can hold any order of events (ADR 0011). Only the tree as a whole, replaying events, and the sandbox (ADR 0012) are machines.
- What may still happen to a closed node is a separate question (#43), and this decision does not touch it.

## Rejected

- **Moving Admission into the table.** It would put rules that read derived status beside rules that must not, and the compile guard would have to go.
- **A generic state-machine engine.** The table is a Go map of three functions, and no engine would make it shorter. ADR 0012 records why the sandbox shares none.

- **Typed constants on one field.** One field cannot hold two word types without an interface or casts at every read, and the kinds would still share a record that a Result has no use for.
- **A status interface.** It gives nothing `Display` does not. A method per kind on a shared interface moves the kind check into a type switch at each reader, where it can be forgotten.

## Amended 2026-11-08: a payload field is a word too

PR #76 let one checkpoint field, `closes`, name a reading key, which closed the reading, or a limit key, which accepted the limit. Replay stayed total, but one word meant two transitions, and every reader repeated the kind check that this decision rejected for statuses. So a payload field that names keyed items names one kind of item and one transition: a Result's `closes` closes limits, a checkpoint's `closes` closes readings and its `accepts` accepts limits, as a confirm's `accepts` does, and a reading's `of` names the answered reading it reads. A generic `answers` field whose effect depends on the key was rejected for the same reason, and a stored flag that copies a derived state goes: the select's `tradeoff` is ignored, because a trade-off is two or more candidates, and the owner's turn to choose is the ledger item `choose` (replay spec §4, §8.1). A checkpoint's `revises` replaces the answer of an answered reading, and is the only way to: a second `closes` on an answered reading is refused rather than read as a replacement, which would give `closes` two transitions. A revision never reopens a reading and does not change a verdict already recorded; the runner owes the reading again until a Result applies the answer in force.
