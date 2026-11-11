---
status: proposed
---

# One store holds many Goals

Until now a tree had one Goal (spec §5 rule 4: "only one Goal per log"). A new purpose was a new tree, and that tree cited the old one in prose as `<tree>:ID`, which hyp does not check (ADR 0016). After the multi-tenant tree completed on 2026-11-10, the owner wanted the build that follows a tree to run from the tree instead of from a plan in prose. Every problem with that came from the one-Goal rule:

- A build had no checked way to cite what the tree had decided.
- The owner's checkpoints were split across trees.
- A commit could not be a subject, because `hyp record` copied file hashes and dropped the sandbox's `base` (`internal/cli/subject.go:118`).

In session `e6a35a24` (2026-11-10 and 11) the owner settled the model: **building is a Goal like any other.** hyp keeps no discovery kind and no delivery kind. The design questions it raised are recorded and selected in `investigations/checkpoint-store-design` (G-73R704), with that session as evidence:

- Q-D906CD (R-HW9B05): what a Result keeps about its runs.
- Q-HC4V9T (R-V8032Z): where a build Goal's start commit lives.
- Q-H8PJ20 (R-8CJCV6): what the owner's act is attached to.
- Q-R21S8R (R-AWPQ4X): how many event types there are.

`investigations/checkpoint-store` (G-D1CH1R) builds the design, and its nine owner lines are this ADR's acceptance. Tracked in #126, part of #114.

## Decision

- **One store holds many Goals.** The store is still one `.tree/` with one events file per actor. A tree is a view, not a container: a Goal and every node whose parents lead to it. Rule 4's "only one Goal per log" goes.
- **A subject has one shape: a base commit, plus what is given over it.** What is given over it is at most one patch, and input files. This holds for every sense of the word:
  - an Experiment's subject is what its sandbox assembles;
  - a run's subject is the base and files its records name;
  - a Hypothesis's subject may name a base, and names files by sha256.

  A Result keeps **one subject per cited run**, `{base, files}`, instead of one merged `fingerprints` list. A Hypothesis is covered when one entry holds its whole subject, so a run at B plus a patch never covers bare B. An old Result reads as one entry with no base. The word `fingerprints` stays for the hashed files inside a subject (Q-D906CD).
- **A Goal records no start commit.** A build Goal's start is the base of its red-on-base Result's subject (Q-HC4V9T).
- **A Goal's lines come with the Goal.** The owner adds the Goal with `lines: [{text, cites}]` in its `node.add`, at most 10 lines of 1 to 600 runes. No later event adds a line.
- **A cite is a plain key, 0 to 10 per line.** Each must be settled when the line is written:
  - a line;
  - an accepted or closed limit;
  - an answered reading;
  - a selected or contradicted Hypothesis.

  Anything else is refused with exit 1. A cite carries no relation label. Its meaning is derived from the key's state at the line's position in the log:

  | The cited key, when the line was written | The cite means |
  |---|---|
  | a line, or an accepted limit | carried |
  | a closed limit, or a contradicted Hypothesis | forbidden |
  | an answered reading | applied |
  | a selected Hypothesis | realised |

  A Goal's origin is the Goals its cited keys belong to. When a cited key's state later differs, `hyp check` warns and `hyp show` prints both states; no status changes.
- **The owner's acts are their own event type, `answer`.** It holds `answers`, `accepts` and `returns`, each `[{key, text}]`, with at least one key. It has no `node`, no `summary`, no envelope `why` and no `role`. The owner's reason is in each item's text (Q-R21S8R, Q-H8PJ20).
  - An answer settles an item Awaiting owner lists, or replaces the owner's own earlier answer.
  - Whether an item in `answers` is a first answer or a revision is derived from the reading's state, so `closes` and `revises` merge.
  - `hyp answer` warns, at exit 0, when it revises a key another Goal's line cites.
  - An accept means only "this Goal does not run this case". A later Goal carries the case by citing the accepted key, and there is no deferred state.
- **The runner's `checkpoint` keeps its Goal as `node` and a required `why`.** It holds `summary` and `refers: [{key, text}]`, and at least one of the two. An empty `checkpoint` is refused.
- **Whose hand an item is in is the contract between the two event types, and OWED is a view of it.** The move tables (`internal/replay/keyed.go`) enforce the hand on the stored log and in a merged log. No write is checked against OWED.
- **A move never crosses Goals; a cite or a reading's `of` may.**
  - A refer must lead to its checkpoint's Goal.
  - `aims`, `closes` and `applies` must lead to their Experiment's Goal.

  A later Goal takes a case over by citing it, never by moving it.
- **Status across Goals:**
  - OWED is per Goal.
  - Awaiting owner is store-level, grouped by Goal.
  - `hyp status G-…` shows one Goal.
  - `hyp status` prints today's output in a one-Goal store, and in a store with many Goals a GOALS table and one Awaiting owner list.
  - `LastCheckpoint` is per Goal and counts runner checkpoints only.
- **A free-text owner approval has no event.** A runner that needs the owner's word records a reading, and the owner answers it.

## Consequences

- **Breaking, amended by the owner decision of 2026-11-11.** The break is in what commands write, not in what replay reads. An old checkpoint's `lines`, `role`, `closes`, `revises`, `accepts` and `returns` stay read-only history that replay still applies and no command writes, as `confirm` is in spec §4; its `refers`, stored as bare keys, decode as refers with no text. G-D1CH1R's own lines are stored that way, on its first owner checkpoint, and keep their keys. `hyp checkpoint` no longer takes `--role`, `--closes`, `--revises`, `--accept`, `--returns` or `--line` (exit 2, naming `hyp answer` or `hyp add goal --line`). In the public trees, id-citation and recipe-base, each checkpoint holds only a `summary`, so they replay unchanged. G-D1CH1R line 1 is that census.
- **Order of the build.** Two-Goal stores land first. Before that, a test of a cross-Goal rule fails while building its two-Goal fixture, which is red for the wrong reason. Each negative-space line gets a test pair: a refused write, and a rejected replay of a log holding the forbidden event. Each half fails on the commit before its own change, with a message naming its rule (G-D1CH1R line 9).
- **Superseded text, changed in the PRs that build it:**
  - principles §5 ("a new purpose is a new Goal, so a new tree") and §7 rule 1 ("its lines are written on one owner checkpoint … so a new tree");
  - the glossary's Goal, Line, Subject, Coverage and Checkpoint, plus a new Answer entry;
  - ADR 0016's "A new Goal is a new tree" and its cross-tree cite rules;
  - ADR 0019's table row "a new Goal, so a new tree";
  - spec §4 (the `checkpoint` row, a new `answer` row, `lines` on `node.add`), §5 rules 4, 21 to 24, and §8, §8.1 and §8.2.

  The skills follow under the recorded-decision exception, and each commit names this ADR.
- **Unchanged.** `choose` stays settled by the runner's `dq.select` (ADR 0015). The owner is still not authenticated (principles §7 rule 7): when the owner approves a write in a session, the runner may type it under the owner's actor name after showing the command, and that approval is the procedure.

## Considered Options

- **A relation enum on each cite** (`carries`, `applies`, `realises`, `forbids`). Rejected: every label is a function of the key's state at the line's position, and a stored label could only agree with replay or contradict it (principles §2).
- **A separate `bases` or `commits` field beside `fingerprints`.** Rejected: it split one concept into several names. A merged list of bases also lets a patched run cover a bare commit (R-HW9B05).
- **One base per Result.** Rejected: no single Result could show "red on B, green on C" with both halves checked (R-HW9B05).
- **A start commit on the Goal.** Rejected: it goes wrong when the base moves mid-build, and only builds would fill it (R-V8032Z).
- **One owner act per Goal.** Rejected: one sitting across two Goals would need two events, and the owner wanted one checkpoint source per project (R-8CJCV6).
- **One `checkpoint` type with its kind derived from its fields.** Rejected: the envelope (`node`, `why`) differs by kind, and the spec defines the envelope per type (R-AWPQ4X).
- **Lines on the first owner checkpoint, as today.** Rejected: lines have the Goal's life. They are written once, by the owner, and keyed on the Goal.
- **Migrating old `<tree>:ID` cites into one store.** Rejected: the log is append-only, and the one real cite between trees names a tree no longer kept. Old trees stay as they are, as history.
