# The principles hyp holds to

This page states why hyp is shaped the way it is. Each section is one claim,
a few sentences of why, and links to the spec, glossary entry or ADR that owns
the rule. No rule is stated only here: when this page and its link disagree,
the link wins. The words are the [glossary](glossary.md)'s.

## 1. What hyp is for

**An investigation stays legible to a later reader, and a decision is admitted
only when the work it rests on is done.** The later reader may be the owner, a
new session or another agent. They read the tree, not a chat, and they find in
it what was claimed, what ran, what failed and what is still owed. Deciding
a design and building it are both Goals, a discovery Goal and a build Goal,
kept and gated the same way; hyp stores no kind
([ADR 0020](adr/0020-one-store-holds-many-goals.md)).

## 2. The log is the truth; everything else is derived

**Events are appended and never edited, and every status is computed from
them.** `supported`, `contradicted`, `stale`, `answered`, the ledger, coverage
and a limit's category are all derived; no command can set one, and no stored
field copies a derived state. A command appends an event: `hypothesis.abandon`
or `dq.withdraw` is a fact a status rule reads, not the status word, and a
later event can move the status again. While it derives, replay keeps working
facts such as whether a question was ever answered; it rebuilds them from the
log every time and never writes them to it. A changed mind is a new event, so the earlier one
stays in history. See [replay-spec.md](replay-spec.md) §7 and §8,
[ADR 0006](adr/0006-scripts-read-the-tree-through-hyp-events.md) and
[ADR 0013](adr/0013-the-tree-machine-owns-its-words-and-its-rules.md).

## 3. A claim is tested by rivals and cases, not by a reader

**A question has rival Hypotheses, and an Experiment fixes its rubric before
it runs.** A recipe's known answers run before every use, so an
instrument shows it can fail before its verdicts count. A reader who checks a
verdict after the fact catches less than a rival that could have won: the
retired verifier passed a wrong verdict, and so did the owner. See
[ADR 0002](adr/0002-known-answers-gate-every-recipe-run.md),
[ADR 0004](adr/0004-a-verdict-checks-known-answers-specify-its-verdict-rule.md)
and [ADR 0015](adr/0015-the-select-is-the-gate.md).

## 4. The ledger owes; it never judges

**`hyp status` names what the tree owes and who owes it, and owing an item
changes no status.** The gate is at the write: `hyp select` is admitted only
when the ledger owes it, and a run stops when the runner owes nothing. Nothing
scores a round. See [replay-spec.md](replay-spec.md#81-ledger-what-the-tree-owes)
§8.1, [ADR 0015](adr/0015-the-select-is-the-gate.md) and
[ADR 0017](adr/0017-the-ledger-is-the-stall-guard.md).

## 5. A limit is the next case, and moving it up the tree is how thinking works

**A limit names a case no check has run, and its life is a few typed moves.**
A Result records it, a later Result closes it by running the case, the runner
refers what no test within the Goal can run, and the owner accepts or returns a
referred one. A case can also change what the work is for: a new Hypothesis is
a new branch, and a new purpose is a new Goal. One store holds many Goals, and
a tree is the view of one Goal ([ADR 0020](adr/0020-one-store-holds-many-goals.md)). A move never crosses Goals: a later Goal takes a case over by citing it. None of these edits what was recorded. See
[ADR 0016](adr/0016-a-limit-is-an-untested-case.md) and
[replay-spec.md](replay-spec.md#82-keyed-items-limits-lines-and-readings) §8.2.

## 6. Coverage says the bytes ran, not that the claim held

**A Result covers a Hypothesis it judged, or one whose whole subject, its base,
patch and files, one run it cites holds.** hyp copies each run's subject from
those records; the runner never types it. A matching hash proves which bytes ran, not which
claim they stand for, and only a verdict says whether a claim held. See the
glossary's Coverage and Subject.

## 7. Who does what

**The ledger asks, the owner answers, the runner runs, and hyp derives.**

The owner is the person hyp asks on real choices. Their role holds to these
rules:

1. **The Goal is written once.** Its lines come with its add and are never
   edited; a line may cite settled keys of other Goals, and no move crosses Goals. A new line or a new
   purpose is a new Goal, which may sit in the same store. A narrower scope is a line the runner refers and the owner
   accepts; there is no other change of scope inside a tree.
2. **The owner is asked only on real choices, and only when the runner owes
   nothing on the question.** The ledger routes each one to AWAITING OWNER: a
   reading, a referred limit, a choice between candidates.
3. **Every owner move but the lines answers what the ledger hands them.** A
   reading answer sets what a line or an answer means. An accept says this tree
   does not run a case, and a return sends it back to run. A choice selects one
   of the candidates that meet every line, for the owner's own reasons. Each
   move turns one item toward the Goal, or the owner starts a new Goal.
4. **An owner move is one typed event, an answer, and hyp reads the move,
   never the prose.** An answer may move items of several Goals, and it never
   holds a runner field. The owner's words are kept and shown back, and no status
   reads them.
5. **An owner move changes no status.** Statuses move by verdicts: an answer
   takes effect when a Result applies it, a return when a Result runs the case,
   a choice when the select the ledger owes is recorded.
6. **The owner produces no evidence.** They never run a case, type a verdict,
   record or close a limit, tie a subject or write a fingerprint.
7. **hyp does not authenticate the owner.** Separation of duties compares actor
   names; that the owner stood behind a move rests on the procedure.

A check the Goal never asked for cannot decide inside the tree, because a
verdict on it would drop a candidate for a case the Goal does not hold. It can
inform a choice. If the owner would reject a design that fails it, it is a
requirement the Goal lacks, and so a new Goal.

The runner settles every item the ledger owes it, maps a recipe's checks onto a
line when it closes the line, refers what no test within the Goal can run, and
stops when it owes nothing. It never answers a reading, accepts, returns or
writes a line, and it records the owner's choice but never makes it. See the
glossary's Owner, Runner and Operator, and
[replay-spec.md](replay-spec.md#82-keyed-items-limits-lines-and-readings) §8.2.

## 8. Cheap errors come from hyp, not from the runner's judgment

**A failure the tree can show goes into the binary as a refusal or a warning.**
Skill text teaches only what hyp cannot check. When a fix would ask the runner
to judge in prose what hyp could see, the gap is in hyp. See
[how-hyp-improves-itself.md](how-hyp-improves-itself.md#mechanical-failures-belong-in-the-binary)
and [sandbox-spec.md](sandbox-spec.md) §1.3 for the exit codes.

## 9. hyp changes only on a real event

**A change is proposed when a real investigation points at hyp, and it earns
its way in through a gate.** [how-hyp-improves-itself.md](how-hyp-improves-itself.md) explains
why the loop is guarded while it runs, and [AGENTS.md](../AGENTS.md) states the
gate.

## 10. What hyp does not claim

- It does not authenticate actors.
- It does not judge whether an Experiment was well designed
  ([how-hyp-improves-itself.md](how-hyp-improves-itself.md#the-ledger-shows-a-stall-nothing-scores-a-round)).
- A select is not proof that the design is right. A later Result can
  contradict the selected Hypothesis or make it stale, and the question
  reopens ([replay-spec.md](replay-spec.md#74-designquestion-state) §7.4).
- A trial run from a shell, outside hyp, leaves no record, so the tree cannot
  see it ([ADR 0015](adr/0015-the-select-is-the-gate.md), Losses).
- A complete Goal is not a met one: the owner reads met as a complete Goal with
  no line owed, and hyp stores no word for it. See the glossary's Goal.
