## Review of eed0938..8a12324 (feat/one-history) against G-AKJJF8

I built the candidate from a detached worktree at 8a12324. The binary went to my scratchpad, and I removed the worktree afterwards. `go vet ./...` and `go test ./... -count=1` pass. I probed the binary in throwaway repositories under the scratchpad with HOME set to a temp folder. I wrote nothing to this repository; here I ran only `hyp check` and `hyp status --all`.

### Findings, most severe first

**1. A store reached through a symbolic link that sits inside another repository joins that other repository's history.**
- **Where:** `internal/store/history.go:142-143`, with the join case at `:184` and the lock at `:395`.
- **Goal lines:** 1 and 5 (also the ADR's "read where the link leads").
- **Confidence:** high, reproduced.
- **Cause:** the symlink is resolved only when the lexical path is in no repository. When the link lives inside a second git repository, the code does not resolve it. `storeOfHistory(r2, "linka")` then accepts the folder as an untracked store of r2.
- **Scenario:**
  - Repository r1 has tracked stores `inv/a` (Goals GA and GA2) and `inv/b` (GB). GA2's line cites `G-…(GB)#1`.
  - A second repository r2 holds a link `linka -> r1/inv/a`.
  - In `r2/linka`, `hyp status` shows only GA. GA2 is missing because its add is rejected: GB is not in the history read from there.
  - `hyp status --all` there lists only store a.
  - `hyp check` there reports `error: rejected …GZSG1F: bad cite … R-BY40Q6#1 …` and exits 1. In r1 the same store checks clean.
  - `hyp add dq --parent G-GZSG1F` there gives `unknown node` (exit 2).
  - A write there succeeds, but it is validated against a history without store b. It locks r2's root, not r1's.
- **Impact on line 5:** two writers, one in r1 and one through the link, hold different locks. They can append to the same store without serializing, so a duplicate alias or out-of-order ids become possible.
- **Fix direction:** also resolve the link when the real path's repository root differs from the lexical one.

**2. `hyp sandbox create --cut EVENT` without `--store` cuts only an untracked root `.tree`, not the history.**
- **Where:** `internal/sandbox/sandbox.go:726` (`store.OpenHistory(o.Host)`), with `history.go:191`.
- **Goal line:** 10.
- **Confidence:** high, reproduced.
- **Cause:** `OpenHistory` returns `Single(root)` for an untracked store at the repository root. That is exactly the "clone a store cut wrote" layout the spec names. `h.Root == host`, so the cut is written as the root store only, and `writeHistoryCut` deletes every base store's `.tree`.
- **Scenario A (root `.tree` copied from `inv/a`):**
  - Host r4 has tracked `inv/a` and `inv/b`, plus an untracked `.tree` at the root holding a copy of `inv/a`'s events.
  - Run `--cut <first event of a>`. It succeeds.
  - The clone has `wt/.tree` and no `inv/a/.tree`; `inv/a`'s events are in the root folder.
  - Line 10 requires each store in its own folder.
- **Scenario B (root `.tree` holding nothing of the cut):**
  - The same command fails with `--cut: no such event … in r4` (exit 2), although the host's history holds the event.
- **Fix direction:** use `OpenRepoHistory` here. With it, the untracked root store is left out and the repository's history is read.

**3. A lock change that may matter: an old hyp and the new one no longer lock each other out.**
- **Where:** `internal/store/history.go:395`.
- **Goal line:** none.
- **Confidence:** medium. This is from reading the code; I did not race two binaries.
- **What changed:** the old binary takes `flock` on `<store>/.tree/events`. The new one takes it on the repository root.
- **Scenario:** an older hyp and the candidate append to the same store at the same moment. Neither waits for the other, so both validate against a stale log, and duplicate aliases or ids become possible.
- **Why it may be acceptable:** the owner accepted only that an older hyp rejects cross-store cites, not that the two run without a shared lock. This may be acceptable but is not documented.

**Open readings already in G-AKJJF8 (not counted as defects):** R-0C4WHP#r1 (line 3), R-0C4WHP#r2 (line 8) and R-8DHPFE#r1 (line 6). My probes confirm the code does what each reading describes:
- An untracked store joins the history read from its own folder.
- A write goes to the current folder's store, and an event that would cross stores is refused there.
- A line whose id only a rejected event holds is applied.

### Verdict per line
1. **Met.** In a store's folder, `hyp status` shows only that store's Goals; through an inner symlink it resolves the same store. Finding 1 breaks it for a link that sits inside another repository.
2. **Met.** `--all` works from a store folder, the root and a testdata store. Plain status with no store lists all Goals. `hyp status G-…` works from another store and from a testdata store.
3. **Met.** testdata, nested `.git`, `.tree`-nested trial stores and untracked stores are each left out. Outside git, the history is the one store. Subject to reading R-0C4WHP#r1.
4. **Met.** A cross-store `--cite` works. An unsettled cross-store Hypothesis cite is refused with exit 1. `hyp show` lists `cited by: … (store inv/a)`.
5. **Met in one clone.** 45 parallel appends by three actors to two stores gave ids in ts order, no duplicates and `check` with 0 errors. Findings 1 and 3 are the gaps.
6. **Met by code.** The merge order is id, actor, store path, line bytes, and `SeenIDs` is global. Subject to reading R-8DHPFE#r1.
7. **Met.** The same alias in two stores resolves per folder and is `ambiguous node g` (exit 2) at the root. Verdict labels resolve within the Result's store.
8. **Met.** A cross-store parent is refused with exit 1 and the store named. `hyp add goal` with no store is refused with exit 1. A two-store answer is refused (code read). Subject to reading R-0C4WHP#r2.
9. **Met.** At this repository's root, the candidate's `hyp check` gives `check: 0 errors, 53 warnings`. The history is the five `investigations/` stores; testdata and trial stores are left out.
10. **Met in the usual case.** In the clone, store a holds events up to EVENT and keeps the evidence folder of the one Experiment before the cut. Store b, which holds no event at or before EVENT, is not in the clone, and neither is base-only store d. The evidence folder of an Experiment added after EVENT was left out. `hyp status --all` in the clone equals the host's `--all --as-of EVENT`. Finding 2 breaks it when the host holds an untracked root `.tree`.

Skill text cites no ADR. Specs and glossary are updated in the same change.

Status: DONE_WITH_CONCERNS
Summary: Lines 1–10 hold in the common layouts, and vet, tests and `hyp check` (0 errors) pass. Two defects reproduced: a store reached through a symlink inside another repository joins that repository's history (lines 1 and 5), and a history cut on a host with an untracked root `.tree` cuts only that root store (line 10). One lock change may also matter: an old and a new hyp no longer lock each other out.
