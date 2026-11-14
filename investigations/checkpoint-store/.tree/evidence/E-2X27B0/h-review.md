## Review of `git diff eed0938 3a886de` (feat/one-history, G-AKJJF8)

How I checked it: I built the candidate from a clean detached worktree at 3a886de into my temp folder. `go vet ./...` and `go test ./... -count=1` both pass. `hyp check` and `hyp status --all` at this repository's root return `check: 0 errors, 50 warnings` (exit 0) and list the Goals of all five stores. I then ran throwaway repositories with HOME set to a temp folder, covering: two stores sharing an alias, a cross-store parent, `add goal` at the root, a testdata store, a fresh untracked store, symbolic links, two actors appending in parallel, and two history cuts (base at HEAD, and base at HEAD~1 where the store did not exist yet). The worktree, sandboxes and temp repositories are removed. Nothing in this repository was written.

### Findings, most severe first

**1. A store reached through a symbolic link from outside the repository falls back to a one-store history, and its appends can sort before the other stores' ids** — internal/store/history.go:124-139 (`openHistory`, using `recipe.RepoRoot(abs)` on the path as typed). Line 5, partly 2 and 4. Sure: high on the behaviour; medium on whether a path outside the repository is "in one clone".
- Scenario: a repo with tracked stores `a` and `b`, where `b` holds an id later than now (01M9000000…, as after a clone whose clock ran ahead). I ran `ln -s <repo>/a /elsewhere/alink`, then `cd /elsewhere/alink && hyp add dq --parent ga …`.
  - RepoRoot finds no `.git` above the link, so the history is `Single(a)`.
  - The new id 01M7ABSSCJ… is written into a's events file and sorts before b's 01M9000….
  - The same command run from `<repo>/a` wrote 01M9000001…, correctly after b's id.
- Two more effects of the same fallback:
  - The lock taken is on the store's folder, not the repository root. A writer using the link and one using the repo path do not exclude each other.
  - From the link, `hyp status --all` lists only store a's Goals, and `hyp add goal --cite 1=G-<b's goal>#1` fails with `unknown node` (exit 2).
- The fix commit 3a886de handles "another path" only for links inside the repository (`sameStore`).
- Possible fix: resolve the current directory with `filepath.EvalSymlinks` before RepoRoot and FindRoot.
- Note: line 3's literal wording ("no .git at or above the current directory") permits this, so it is partly a gap in the line itself.

**2. A legacy `confirm` that accepts limits is not checked by the store rule** — internal/replay/stores.go:41-133 (`checkStore`, which for a confirm checks only `ev.Node`). Line 8. Sure: high from reading the code, not probed.
- `validateConfirm` (validate.go:739) checks each accepted key with `checkKey(k, KeyLimit, ev.Node)`. That accepts a Goal-line key (`G-…#n`) from any Goal, because a Goal line counts as "under every question".
- Scenario: a merged or hand-written owner confirm in store a, on a DQ of a, accepts `G-<b's goal>#1`. Replay applies the accept, so the event moves a key of store b and is not rejected.
- Spec rule 33 lists a checkpoint's and an answer's keys but not a confirm's `accepts`.
- No command writes a confirm, so this can only arrive through history or a merge. Severity is low.

**3. `hyp status G-… --as-of ALIAS` gives different output depending on the folder** — internal/cli/read.go:107 and 147-169. Line 2 (against line 7). Sure: high, probed.
- Scenario: alias `q` is held only by store a. `hyp status G-<a's goal> --as-of q` prints the Goal from `a/` and from the repository root, but from `b/` it prints `unknown node q` (exit 2).
- Cause: the `--as-of` alias resolves only in the current store. The fallback to the repository history runs only when the two histories differ, and from `b/` they do not.
- This follows the alias rule as written, but line 2 says `hyp status G-…` prints the same from anywhere. A possible fix: when `G-…` is given, resolve `--as-of` in the Goal's store.

**4. `hyp diff` and `hyp signals` still read the whole history from a store's folder** — internal/cli/read.go (cmdDiff and cmdSignals are unchanged and use `r.full` / `r.evs`). Line: none. Sure: high, probed.
- In store a's folder, `hyp diff --since 2020-01-01T00:00:00Z` listed b's Goal, question and Hypothesis.
- `status`, `events`, `render` and `checkpoint` were all narrowed to the current store. The view rule in spec §6 covers only `status`, so these two are now inconsistent. Low.

**5. `sandbox remove` keeps only `wt/.tree`; a history-cut sandbox has its stores under `wt/<store>/.tree`** — internal/sandbox/sandbox.go:1621 (`saveRecord`). Line: none. Sure: medium.
- After a history cut there is usually no `wt/.tree`, so OUT_DIR gets no tree copy.
- The runner's new events reach OUT_DIR only through `changes/`.
- The sandbox spec's description of `remove` was not updated for history cuts. Low.

**6. A store at the repository root that git does not track is skipped when the next id is chosen** — internal/store/history.go:81 (`dir == "."` skips the root in `untrackedLast`). Line 5, edge case. Sure: medium, not probed.
- Writers in such a root store use `Single(root)`, so their ids count only the root store's own events.
- If the root store is tracked later, it joins the history, and its ids may not follow append order.
- The spec says an appended id sorts after "every id of a store git does not track yet that would join the history once tracked", and a root store would join. Rare; a `--store` cut clone is the usual case of an untracked root store, and is meant to stay separate.

**7. A git-ignored store joins the history when read from its own folder, but other writers ignore its ids** — history.go:67 (`--exclude-standard`) versus 164 (`storeOfHistory` does not check whether a store is ignored). Line 3, edge case. Sure: medium, not probed. It is the same question as the open reading R-0C4WHP#r1 (meaning 2). Low.

### Verdict per line
1. **Pass.** In a store's folder, `status` shows only that store's Goals and rejections (probed with one Goal and with two).
2. **Pass, with finding 3.** `status --all` and plain `status` at the root list every Goal and every item awaiting the owner. `status G-…` gave byte-identical output (same md5) from a, b, the root, a testdata store and a fresh store.
3. **Pass for the stated cases.** testdata, `.tree/` evidence trials, `known-answers/` and nested `.git` are excluded; here the history is exactly the five `investigations/` stores. Findings 1 and 7 are edges.
4. **Pass on the code paths and tests; I did not probe settled-ness end to end.** Cites resolve across stores, and `show` lists "cites" and "cited by". A cite from a store reached through an outside link fails (finding 1).
5. **Pass inside the repository.** Two actors appending 15 events each to two stores at once: id order equals append order and check shows 0 errors. Fails through an outside symbolic link (finding 1); finding 6 is an edge.
6. **Pass.** The merge order is id, actor, store path, trimmed bytes, raw bytes; the duplicate rule rejects the later line. The open reading R-8DHPFE#r1 covers its remaining meaning.
7. **Pass.** The same alias in two stores is allowed. `show ga` at the root returns `ambiguous node ga` (exit 2). A reused alias within one store is still refused. Verdict labels are scoped to the Result's store in claim.go.
8. **Pass for writes** (probed: cross-store parent refused with exit 1; `add goal` at the root refused with exit 1; an answer with keys in two stores is refused with exit 1). **Replay misses a legacy confirm's accepts** (finding 2).
9. **Pass.** `check: 0 errors, 50 warnings`, exit 0, at this repository's root.
10. **Pass.** Each cut store holds exactly its events at or before EVENT. A base store with no such event loses its `.tree`. The evidence folder of an Experiment added after EVENT is left out. The testdata store is untouched. A store missing from the base is created and registered with intent-to-add. `status --all` in the clone equals the host's `status --all --as-of EVENT`, also with the base at HEAD~1.

Status: DONE_WITH_CONCERNS
Summary: The change meets all ten lines in the probed in-repository scenarios, and vet and tests pass. The main defect is that a store reached through a symbolic link from outside the repository is read as a one-store history: its appends can sort before other stores' ids and take a separate lock (line 5). A smaller gap is that replay does not check a legacy confirm's accepted keys against the store rule (line 8).
