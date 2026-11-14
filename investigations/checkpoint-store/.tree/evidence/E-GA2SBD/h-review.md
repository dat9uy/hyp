## Review of `git diff eed0938 3bbf187` (G-AKJJF8)

I built the candidate from a clean detached worktree at 3bbf187 into my scratchpad. `go vet ./...` is clean and `go test ./... -count=1` passes. I probed the binary in throwaway git repos under mktemp folders. `hyp check` and `hyp status --all` ran here read-only. I removed the worktree and every probe sandbox, and the repository's `git status` is the same as at the start.

### Findings, most severe first

**1. A store cut (`--store SRC --cut EVENT`) now gives a clone whose history rejects the cut's own events.**
- **Where:** `internal/sandbox/sandbox.go` in `build()` (the `cuts["."] … o.Store != ""` branch), together with `internal/store/history.go` `openHistory`, the case where an untracked store at the clone's root joins the history.
- **Scenario:**
  - The host has `investigations/a` tracked at `--base`.
  - `hyp sandbox create … --store <host>/investigations/a --cut <Q's id>` writes the cut to `wt/.tree`.
  - The base's own `wt/investigations/a/.tree` stays in the clone.
  - In `wt/`, the new binary reads `.` (untracked, read from its own folder) and `investigations/a` as one history, so every event of the cut has a duplicate id.
- **Result:** `hyp check` in `wt/` printed:
  - `rejected G-…: duplicate id`
  - `rejected Q-…: duplicate id`
  - `event crosses stores: its parent Q-… is in store . and the event in store investigations/a`
  - then `check: 3 errors`, exit 1.

  The installed pre-change hyp on the same clone printed `0 errors`.
- **Further effects:**
  - The base copy's events after the cut also join the clone's history, so the future leaks into `--all`, into cites, and into which id an append sorts after.
  - sandbox-spec says the clone's `.tree` "is the cut and nothing else" and that status in `wt/` equals `--as-of` in SRC. That is now true only for the view, not for check or rejections.
- **Fix direction:** exclude a root store written by a store cut from the base's stores, or remove the base's copy of SRC.
- **Sure:** high (reproduced).
- **Goal line:** none. It is a regression of the `--store` path that ADR 0023 says is unchanged. It sits next to line 10.

**2. `hyp status G-… --as-of REF` fails when run from a folder whose store adds nothing to the history.**
- **Where:** `internal/cli/read.go`, `cmdStatus`. `r.at(*asOf)` resolves REF against the local history before the fallback to the repository's history.
- **Scenario:** from `x/testdata/s` (its own store), Goal B and its question Q are held by `investigations/b`.
  - `hyp status G-B` works.
  - `hyp status G-B --as-of Q-…` prints `unknown node Q-…` and exits 2.
  - The same command at the repository root prints Goal B.
- **Sure:** high (reproduced).
- **Goal line:** 2 ("run anywhere in the repository, prints what it prints in the folder of the store that holds that Goal").
- **Related:** `hyp show G-B` from that folder is also `unknown node`. Line 4 does not say "anywhere", so this is only a note.

**3. The same-ask warning after `hyp answer` tells the owner to add a key from another store.**
- **Where:** `sameAskWarnings` in `internal/cli/write.go` reads the whole history.
- **Scenario:** two stores hold readings with the same text. `hyp answer --answer R-B#r1=…` prints `warn: R-A#r1 asks the same as R-B#r1 and stays open; add --answer R-A#r1`. Adding that key to the same answer is refused: `an answer's keys lie in two stores`, exit 1.
- **Sure:** high (reproduced).
- **Goal line:** none. The advice contradicts line 8's rule; the warning should say to answer it in its own `hyp answer`.

**4. Two different write locks can serve one history.**
- **Where:** `History.lock` in `internal/store/history.go`. A history of exactly `["."]` locks `<root>/.tree/events`; any other history locks the repository root.
- **Scenario:** a repository whose only tracked store is at its root, plus an untracked store under it that a writer reads from its own folder. Writers in the two stores then take different locks, so their appends are not serialized, which line 5 asks for.
- **Sure:** medium on the code path, low on reach. `hyp init` refuses below a root store, so the state needs a hand-made or checked-out-but-untracked store. Not reproduced.
- **Goal line:** 5.

**5. Minor message and comment slips.**
- From a folder with no store, `hyp checkpoint` without `--goal` says `missing --goal: the store holds 2 Goals` while counting every Goal of the history (`checkpointGoal`, write.go).
- `hyp check` names a rejected event (`error: rejected G-000002: duplicate id`) without its store's path, so the owner cannot tell which file holds the later line.
- The `CreateOptions` comment in `internal/sandbox/sandbox.go` (around line 355) still says "Store and Cut, both or neither".
- `skills/hyp-verification/RECIPE-FORMAT.md` describes only the `--store` cut, not the history cut.
- **Sure:** high.
- **Goal line:** none.

### Verdict per line
1. Met. Status in a store's folder shows only that store's Goals and rejections. A Goal belongs to the store holding its add, including a Goal whose line cites another store.
2. Met for plain `hyp status`, `--all` and `hyp status G-…` (also from a `testdata/` store). Broken with `--as-of` from a left-out folder (finding 2).
3. Met. Untracked stores, stores under `.tree/`, under `testdata/`, under `known-answers/` beside `recipe.md`, and under a nested `.git` are all left out. With no git, the history is the one store.
4. Met. Cross-store cites of open readings, open limits and unselected Hypotheses are refused with exit 1; settled ones pass. `hyp show` prints `cited by: G-… (store …)`.
5. Met. 30 concurrent appends across two stores kept id order equal to ts order and `check` reported 0 errors. One narrow lock gap remains (finding 4).
6. Met. A duplicate id across stores is rejected in the later store by path. The next append sorts after it.
7. Met. `show h-null` from the root is `ambiguous node` (exit 2) and resolves locally from a store's folder. Verdict labels resolve in the Result's store.
8. Met. A crossing node add is refused at write (exit 1) and rejected on replay. A two-store answer and `add goal` with no store are refused with exit 1, naming the rule.
9. Met. `hyp check` at the repository root reads the five `investigations/` stores: `check: 0 errors, 44 warnings`, with no rejections.
10. Met. A history cut writes each store that holds an event at or before EVENT in its own folder. A store added later is absent. `status --all` in the clone equals the host's `--as-of`. The `--store` cut regressed (finding 1).

Status: DONE_WITH_CONCERNS
Summary: All ten Goal lines hold on the probes, with one gap on line 2 (`--as-of` from a left-out folder). The serious defect is outside the lines: the unchanged `--store SRC --cut` sandbox now gives a clone where `hyp check` rejects the cut's events as duplicates and crossings, because the base's copy of SRC joins the clone's history.
