---
status: accepted
---

Amends [ADR 0020](0020-one-store-holds-many-goals.md) (one store holds many Goals).

# A store merges by union and is found within its repository

ADR 0020 made one store hold many Goals, so two branches now append to the same store, and often to the same actor's events file. Issue #138 asked where stores live and how they merge. The owner settled the layer every layout needs in the grill of 2026-11-13, recorded as G-1391DM in `investigations/checkpoint-store`, whose six lines are this ADR's acceptance. The layout, one view over every store, is #168 and builds on this one.

Four facts of the store before this change made a shared store break:

- An events file that two branches appended to conflicted in `git merge`, and the conflict could only be solved by hand.
- Replay merged the files by a k-way merge that kept each file's line order, so the same lines in another order replayed differently.
- An appended id came from the clock alone, so an event appended after a merge could sort before events already in the store.
- Store discovery walked up to the filesystem root, so a command in a repository with no store could read a store above that repository, and `hyp init` below a store made a second store that no command above it reads.

`hyp sandbox create --store SRC --cut EVENT` also refused a base commit whose root holds `.tree`, so no replay could start from a commit of a repository that keeps its store at the root.

## Decision

- **An events file merges by union.** `hyp init` writes `.tree/.gitattributes` holding exactly `events/*.jsonl merge=union`. `git merge` of two branches that each appended to one events file exits 0 and keeps every line of both once (line 6). GitHub does not run the union driver, so a PR whose events file conflicts there is merged locally first; AGENTS.md says how.
- **Replay orders events by id, then by actor, whatever the line order inside a file.** A union merge may leave both branches' lines in any order, so the order inside a file carries no meaning. Two lines with the same id and actor are ordered by their line bytes, in replay and in the cut alike, so file order decides nothing. A store cut writes each events file's kept lines in id order, so the same lines in another order cut to the same bytes (line 4).
- **An appended id sorts after every id the store holds.** When the clock reads earlier than the newest id, or in its millisecond, the new id takes the millisecond after the newest id's, with fresh random bits, so two clones whose clocks are behind append different ids, which a union merge keeps apart. In the greatest ULID millisecond the new id adds a random amount to the newest id, and when no id sorts after the newest the append is refused. Append order and id order are then the same within one store, so the new merge order changes no store written by one clock that ran forward (line 5).
- **Discovery stops at the repository root.** It walks up from the current directory and stops at the first directory that holds `.git`. When no directory up to that one holds a store, a command reports no store (exit 3). With no `.git` at or above, it walks to the filesystem root (line 3).
- **`hyp init` is refused where discovery finds a store**, at the directory or above it, with exit 1, naming that store's directory, and it creates nothing (line 2). A store is made only where no command would read another one.
- **A store cut replaces the base's own `.tree`.** On a base whose root holds `.tree`, the clone's `.tree` holds the cut and no other file of the base's `.tree` (line 1).

## Consequences

- Every store under `investigations/` gets a `.tree/.gitattributes` in the PR that builds this.
- **Accepted: an older pinned hyp reading a live union-merged store.** A hyp binary built before this change merges by file order, so on a store whose files a union merge left out of id order it may replay differently from today's hyp. The owner accepted this in the grill (G-1391DM's body).
- Text changed with this ADR: replay-spec §1, §2 and §6, sandbox-spec §5.3, the glossary's Store, RECIPE-FORMAT.md's "A store cut", the `hyp init` placement in skills/hyp-prototype, README.md and AGENTS.md, and docs/install.md.
- **Out of this ADR, as G-1391DM's body states:** a patch that names a path under `.tree/` together with `--store`; a `hyp check` warning for a store whose `.gitattributes` lacks the union line; and a store below a base's root, which #168 covers.

