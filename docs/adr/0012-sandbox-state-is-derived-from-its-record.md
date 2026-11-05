---
status: accepted
---

# A sandbox's state is derived from its record and its lock

A `kill -9` during Launch left `setup` absent. Drive refused only `setup == "failed"`, so `hyp sandbox drive` ran and exited 0 on a setup that never finished. A directory with no record could not be removed at all: create said to remove it and remove said no such sandbox. State was being inferred in several places, from a file, a per-recipe string, a lock and process checks. The sandbox had no machine. The owner accepted this in the architecture review of 2026-11-05 (#44). There is no tree: the choices below were design decisions, not measurable rivals, and the gate is `go test`.

**The sandbox has a state machine, whose state is derived.** One pure function, `deriveState`, reads each recipe's `setup`, whether the lock is held and one stored field, `removing`, and returns `building`, `setting_up`, `interrupted`, `ready`, `setup_failed` or `removing`. Each command refuses by that state ([sandbox-spec.md](../sandbox-spec.md) §2). `interrupted` is never written: it is what an unfinished setup and a free lock mean, so no crash can leave it wrong.

**`setup` is written per recipe, and the manifest before the last one.** A sandbox that reads as `ready` has finished every setup and has its baseline, whichever write a kill lands between.

**Remove can clear what a crash leaves.** It writes `removing` only after the refusals that change nothing, stops the groups of unfinished rows before Cleanup so a Launch that outlived its hyp cannot change what a Cleanup looked at, and then runs as before. `remove --discard` clears a directory with no record; `remove OUT_DIR` refuses it, since there is nothing to keep.

## Consequences

- `removing` is a deliberate behaviour change. A remove refused for a leftover process used to leave a sandbox that could still be driven, though its Cleanup had run. Now drive refuses it, and remove can run again, which the Cleanup contract already allowed.
- Drive exits 3 while an unfinished row's group lives, and runs again once the group is gone, so "run it again" still works after a killed drive whose command died.
- `sandbox.json` is format 4. An older record is read by the same rule: `none` and `passed` read as `ready`, and a recipe with no `setup` reads as `interrupted`, so it can be removed and not driven.
- `hyp recipe evidence` is unchanged. A kept record can come from a setup that never finished, as one from a failed setup already could. The record shows it, and #46 holds the question.
- `skills/hyp-verification/RECIPE-FORMAT.md` states Drive's refusals, the `removing` field and that remove stops unfinished commands before Cleanup. That text ships as a correction agreeing with this decision, with no replay.
- A reused process group id can refuse a Drive or be signalled by remove. This limit already existed on `stopGroups`, and Drive's refusal says remove still works.

## Rejected

- **A stored state word.** It must be kept in step with what it summarises, and a crash lands between the two. Storing the facts and deriving the status is what the tree already does.
- **A sandbox journal.** An event log would give history no reader needs: the record's rows already say what ran, and one field says a remove started.
- **A generic state-machine engine shared with the tree.** The two machines have nothing to share: one replays a merged log and must be total, the other refuses commands on one host. A table test mirrors the spec's table by hand instead of parsing it.
- **Having Evidence refuse interrupted records now.** No event has asked for it, and whether a verdict may read such a record is a separate question (#46).
