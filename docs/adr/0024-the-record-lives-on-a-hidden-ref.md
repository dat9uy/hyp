---
status: accepted
---

Amends [ADR 0022](0022-a-store-merges-and-is-found-within-its-repository.md) (a store merges by union and is found within its repository). Reopens the unmerged ADR 0023.

# The record lives on a hidden ref, owned by one module

Issue #181, order 1 of #180, asked where hyp keeps what an investigation recorded, so that no code-branch act loses it. Before this change a store's events and evidence were work-tree files under `.tree/`, committed on the branch the work happened on:

- Deleting, squash-merging or rebasing that branch, a `git clean -fdx`, or removing the worktree lost every event not yet merged to another branch, and every untracked evidence file (#179).
- Two worktrees of one clone each wrote their own copy of a store, so a write in one was not seen in the other.
- Evidence a Result cited lived beside thousands of files no event named, so the tree's cited files and its noise were tracked together (#152).
- A Result's `questionable` and `uncontained` flags read the checkout's `recipes/` and `.tree/evidence/`, so two checkouts of one store could print different flags for one Result.

The owner settled the place in the grill of 2026-11-14, recorded as G-4N85GM (`record-design`) in `investigations/record`: a hidden ref written by git plumbing (h-ref) over a record worktree (h-record-worktree), on the scenario of a crash between a write's file line and its commit. The build Goal G-SZCBXZ in the same store holds the lines this ADR's code meets.

## Decision

- **The record is commits on `refs/hyp/record`**, which no branch holds, so no GitHub button deletes, merges or checks it out, and a plain `git fetch` does not bring it. A record repository is a git repository that holds that ref. One record per repository; a store keeps its name, its folder's path from the repository root.
- **The record holds only what events name**: every event, the bytes of each file an event names (evidence, an Experiment's records, a subject's patch and files, also of retracted Results), each bundle an event pins, and every commit a Hypothesis subject names, as extra parents of the record commit that named it. `hyp init` and `hyp import` make git ignore every file under the store's `.tree/evidence/`; `hyp import` untracks each file there that no event names.
- **A write is durable once it exits 0.** A write takes one lock for the whole clone, writes its objects, tree and commit, and moves the ref last with a compare-and-swap. A write stopped before that leaves nothing a reader sees.
- **A named file never changes.** A Result stores each evidence pointer's sha256, and a write naming a path the record holds with other bytes is refused.
- **Every read reads the record** when it holds the store: a snapshot of the record commit under the common git directory, which no worktree act touches. So the flags and `hyp check`'s warnings for a Result are the same from every checkout. `hyp check` also reports a work-tree store holding events the record lacks, and a checked-out bundle whose hash differs from a pin.
- **`hyp sync [REMOTE]`** fetches the remote's record, joins the two logs as a set by id and pushes; the hyp-goal-run skill runs it at every stop. **`hyp import [REF:]PATH...`** adds a work-tree store's events the record lacks, from the checkout or a commit, and may run again.
- **A write where a work-tree store holds events the record lacks is refused**, naming `hyp import`, so no event is lost to the readers of the record.
- **Elsewhere, nothing changes.** A sandbox clone (marked by `hyp sandbox create`), a git repository with no record ref and a directory with no git read and write the work-tree `.tree/` as before. Discovery still stops at the repository root (ADR 0022).
- **One module owns it.** Package `record` is the only non-test Go code that names `.tree` or `.git`, and with package `sandbox` the only code that runs git; a go test fails on any other.

## Consequences

- After this change merges, the owner runs `hyp import` on this repository's five stores and on the tag `g-akjjf8/record`; the old `.tree/` folders stay as frozen copies, cited files tracked.
- **Accepted: an older hyp sees only the frozen copies.** A hyp built before this change reads the work-tree copies, sees no later event, and rejects an event that stores an evidence sha256.
- The union merge of ADR 0022 now matters only for work-tree stores; the record joins through `hyp sync`.
- A record PR no longer carries a run's tree: `hyp sync` shares it.
- Text changed with this ADR: replay-spec §1 and §1.1, sandbox-spec §5.3, the glossary's Store, Event, Record and Record repository, the hyp-goal-run, hyp-prototype and hyp-owner-grill skills, RECIPE-FORMAT.md's "A store cut", README.md and docs/install.md.
- **Out of this ADR, as G-SZCBXZ's body states:** the filters of #168 (status across every store, cites across stores, alias scope), a later Goal.
