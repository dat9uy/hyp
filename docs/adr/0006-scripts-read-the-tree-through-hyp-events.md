---
status: accepted
---

# hyp has no index, and scripts read the tree through `hyp events`

The MVP stored events as JSONL and kept a rebuildable SQLite index beside them (`.tree/index.sqlite`, spec §1). hyp never answered a question from it. Every read command replays the whole log, then refreshes the index from that replay, and the only query the CLI runs is `Ancestors`, which has an in-memory fallback. An open question collected the problems an index might solve. After ADR 0002 to 0005, none of them needs one, and one of them needs something else: recipe scripts read tree facts by globbing `.tree/events/*.jsonl` (the session-replay recipe's `common.events()`, used by six scripts). That ties every such script to the store's file layout. It also gives them a different answer from hyp's, because a glob also reads the events replay rejected. Owner decision on 2026-10-31. The shape of the read interface was then tested as rival Hypotheses in a tree whose records stay in the private development history.

## Decision

- **No index.** hyp drops `internal/index` and stops writing `.tree/index.sqlite`. Every read replays the events, as it already does. The events stay the only source of truth.
- **Derived facts live in the Projection.** A fact hyp derives (a status, a selection, a live edge, rivals at add, a verdict's order) is a Projection field that `internal/replay` computes once. Render, signals and the CLI read those fields; nothing stores them a second time. A later change moves what render and signals still rebuild into the Projection.
- **`hyp events --json` is the scripts' read.** Run in a store, it prints one JSON object: `format` (1), `events`, every applied event in merge order (spec §6) exactly as stored, and `rejections`, each refused event's id and reason. It reads the store as `hyp status` does and has no text form: without `--json` it is a usage error (exit 2). A shape change bumps `format`, and a `go test` golden forces the bump, as ADR 0002 requires for hyp's files.
- **Derived facts for scripts come from `hyp signals --json`**, which already exists. hyp adds no `--json` to `status` or `show`. A script reads hyp's text only to measure that text itself, as count-wasted counts the lines of `hyp status`.
- **A recipe script reads no file under `.tree/events`.** It calls `hyp events --json` or `hyp signals --json`. In a check, that is the copy of the running hyp that ADR 0003 puts first on `PATH`, so the gate's `hyp_sha256` names the binary that read the tree, and a check's known answers pin how hyp reads its input.

## Considered Options

- **Answer reads from the index, replaying only when it is stale.** Rejected: speed is not a problem (the largest tree has 52 events and `hyp status` takes about 5 ms), and the freshness check hashes every event file, which costs about as much as the replay. Tables of derived facts would be a second copy of the rules that can disagree with the replay, the pattern ADR 0004 rejected for a `recipe.md` declaration.
- **Keep the index unused.** Rejected: it is dead code that every architecture review raises again. This record is what that question asked for, so that it is not suggested again.
- **`--json` on `status` and `show`, beside `signals`.** Contradicted by a Result of that tree: 12 of the 23 places the session-replay recipe's scripts read a tree fact need event ids, event timestamps, every checkpoint, or a walk over nodes that neither command prints.
- **Both an events command and `--json` on `status` and `show`.** Contradicted by another Result of that tree: with a prototype of `hyp events --json`, a port of `common.events()` and one title read kept 111 of the bundle's 112 known answers. The remaining case changed only because replay rejects its trial's Experiment, and no read needed `status` or `show` JSON.
- **A narrow query command for agents** (the question's candidate 4) **or a cross-tree owner view in hyp** (candidate 3). Not built: no recorded event shows the need. A cross-tree view, when one is needed, can call `hyp status` or `hyp events --json` in each tree; it needs no index.

## Consequences

- The question's candidates are answered elsewhere or not needed. Shared derived facts (1) and signals history (5) go to the Projection. The Experiment-to-rival view (2) and dangling rivals were answered from the replay by earlier changes. The owner view (3) and smaller reads (4) wait for an event. Scale (6) is not needed at current sizes, and a cache could come back as a speed fix written from the Projection, never as a home for rules.
- A script that reads events through hyp sees applied events only: every caller of the session-replay recipe's `common.events()` (count-wasted, extract-exp, replay-check, resupport-check, session-snapshot, tripwire, verdict-moves) skips rejected events. The session-replay recipe's old `extract-exp-hypfail` case, whose trial Experiment names a parent that does not exist, now reports no Experiment instead of failing on `hyp show`; it is renamed `extract-exp-rejected`. A new `extract-exp-hypfail` keeps extract-exp's exit 3 covered: its trial tree holds `events` as a file, so the tree has no events directory to read.
- A script that calls `hyp events` needs a hyp that has the command: in a check, the gate's copy; in a sandbox step, whichever hyp the recipe puts on `PATH`. A pinned subject hyp older than this decision does not have it.
- `.tree/.gitignore` keeps its two lines, and `hyp sandbox create` and `remove` keep skipping `index.sqlite*`. Older hyp binaries, such as the pinned subjects of session replays, still write the file, and an old store may still hold one. hyp ignores it.
- Out of scope: what that later change moves into the Projection, `--as-of` for `hyp events`, and any change to the event schema or a derived status.
- The spec §1 and §8, `docs/install.md`, `RECIPE-FORMAT.md`, the hyp-prototype skill and the CLI help change to match. Each change is a correction that cites this decision. ADR 0001 to 0005 stand.
