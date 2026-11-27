---
status: accepted
---

Amends [ADR 0022](0022-a-store-merges-and-is-found-within-its-repository.md) (a store merges by union).

# Each event of the store is a file of its own

ADR 0022 made two branches that append to one actor's events file merge by the union driver that `.tree/.gitattributes` names. GitHub does not run that driver, so a record PR and a runner PR that both appended to `.tree/events/owner.jsonl` conflicted there, and AGENTS.md told every maintainer to merge such a PR locally first (#397). The owner chose to remove the conflict, not to explain it: Goal G-TWJ4PF holds the want.

## Decision

- **An act adds one new file and changes none.** hyp writes each event of the store to `.tree/events/<actor>/<id>.json`, one JSON line, created once and never changed. Two branches never change the same file, so a plain git merge, with no merge driver, takes the acts of both, on GitHub too.
- **The trace keeps one file per actor.** The clone's trace is never in git and is never merged, so its events stay appended to `events/<actor>.jsonl`.
- **An older store still replays.** hyp reads `.tree/events/<actor>.jsonl` as before, and both forms in one merge order: by id, then actor, then line bytes. `.tree/.gitattributes` keeps its union line for those files, which an older hyp can still append to.
- **Acts that cannot both stand are handled as before.** Two branches that each end one Goal now merge with no conflict, and replay refuses the later act at its place, as it did after a union merge.

## Consequences

- AGENTS.md drops the rule to merge a PR locally because of an events file.
- The hyp-grill skill (`skills/hyp-grill/SKILL.md`) names the actor of a Goal by its folder under `.tree/events/`.
