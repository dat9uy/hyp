---
status: accepted
---

Amends [ADR 0020](0020-one-store-holds-many-goals.md) (one store holds many Goals) and [ADR 0022](0022-a-store-merges-and-is-found-within-its-repository.md) (a store merges by union and is found within its repository).

# A repository has one history

ADR 0020 let one store hold many Goals, and ADR 0022 made a store safe to share between branches. A repository still held several stores, one per folder under `investigations/`, and each was its own log. Issue #168, which continues #138, named what that cost:

- The owner had no overview: `hyp status` showed one store, so the owner had to know which folder held which Goal.
- A Goal could not cite a key of a Goal in another store. The only link between stores was a `<tree>:ID` token in prose, which hyp does not check.
- Two stores could hold ids in any order, because each append counted only the ids of its own store.

The owner settled the model in the grill of 2026-11-13, recorded as G-AKJJF8 in `investigations/checkpoint-store`, whose ten lines are this ADR's acceptance.

## Decision

- **The repository has one history:** every event of every store, ordered by id, then actor, then the store's path, then the line's bytes. Replay reads the history, so every rule that reads the log reads all stores.
- **Stores stay in their folders, and no event moves.** The store an event belongs to is the folder of its events file. No event stores it: the reader sets it from where it read the line.
- **A store adds nothing to the history** when git tracks none of its events files, or when it sits under another store's `.tree/`, under a `testdata/` directory, under a bundle's `known-answers/`, or under a directory below the repository root that holds its own `.git`. With no `.git` at or above the current directory, the history is the one store discovery finds (line 3). A store `hyp init` just made, which git does not track yet but would add to the history once tracked, joins the history read from its own folder, so its first Goal can cite another store; from any other folder it adds nothing. A store that discovery finds and that would add nothing even when tracked is read alone, as its own history, and so is the current store when git cannot list the repository's files.
- **An event never crosses stores** (line 8). A write goes to the store found from the current folder, or, where discovery finds none, to the store that holds its node. Every node and key an event moves lies in its store; replay rejects any other event (spec §5 rule 33). `hyp add goal` where discovery finds no store is refused, since a Goal has no parent whose store holds it, and so is an answer whose keys lie in two stores.
- **A cite may cross stores.** A Goal line's `--cite` and a reading's `of` may name a key of a Goal in another store. Each is checked settled at its place in the history (line 4). `hyp show G-…` lists the Goals it cites and the Goals that cite it.
- **An alias is a name local to its store** (line 7). A rubric's verdict label binds to a Hypothesis by alias, and arms reuse labels such as `h-null`, so two stores may hold the same alias. An alias resolves in the store found from the current folder, and a cited verdict label in its Result's store. Where an alias names nodes of two stores, it is `ambiguous node` (exit 2).
- **One lock for the history.** Every write locks the repository root's directory, replays the whole history and takes an id after every id the history holds (line 5). A history whose stores hold one id twice rejects the later line (line 6).
- **The owner's overview is `hyp status --all`** (line 2). `hyp status` in a store's folder shows that store's Goals (line 1), and `hyp status G-…` prints the same view from anywhere in the repository. Both read the repository's history also from a folder whose store adds nothing to it.
- **A history cut.** `hyp sandbox create --cut EVENT` without `--store` writes each store of the host's history as it was at EVENT, in its own folder (line 10). `--store SRC --cut EVENT` still cuts one store to the clone's root.

## Consequences

- **Accepted: an older hyp binary rejects a cite into another store.** A hyp built before this change reads one store, so a Goal line that cites another store's key is rejected there. The owner accepted this in the grill (G-AKJJF8's body).
- **Accepted: ids keep only the order ADR 0022 gives across branches.** The lock orders appends in one clone. Two clones still order their appends by clock, as ADR 0022 says.
- **A `<name>:ID` token stays unchecked prose** for a tree outside the repository.
- **Changed with this ADR:** replay-spec §1, §2, §5 (rules 2, 13 and 33, and §5.1), §6, §7.6 and §8; sandbox-spec §5.3; the glossary's Tree, Store, History, Cite and Store cut.
- ADR 0020's "Awaiting owner is store-level, grouped by Goal" now reads: Awaiting owner is per view, grouped by Goal, and `--all` is the whole history. ADR 0022's "An appended id sorts after every id the store holds" now reads: after every id the history holds.

## Considered Options

The Goal's body records each of these as decided in the grill:

- **Events moved into one store per repository.** Rejected: stores stay in their folders and no event moves.
- **A store recorded in each event.** Rejected: the store an event belongs to is the folder of its events file, and no event stores it.
- **Aliases unique in the whole history.** Rejected: a rubric's verdict label binds to a Hypothesis by alias, and arms reuse labels such as `h-null`.
