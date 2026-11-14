## Review of `git diff eed0938 c770b50` (feat/one-history) against G-AKJJF8

I built the candidate from a detached worktree of c770b50 into my scratchpad. `go vet ./...` is clean and `go test ./... -count=1` passes. I ran the binary in throwaway repositories, including `hyp sandbox create` with HOME pointed at a temp folder. Afterwards I removed the worktree and all sandboxes. I wrote nothing to this repository.

### Findings, most severe first

**1. A store reached through a symlink inside the repository, or through a path in different letter case, is read twice. Every event in it then gets a false duplicate-id rejection.** Labels: line 6 (false rejection), and line 3 (the history should be the tracked stores only). Sureness: high, reproduced.
- **Where:** `internal/store/history.go:94-109` (`openHistory`). It takes `rel := filepath.Rel(repo, cur)` from the logical cwd, which `filepath.Abs` does not resolve.
  - If `rel` is not in the `git ls-files` store list, the "store hyp init just made" branch runs. `storeOfHistory` follows the symlink and returns true.
  - The same physical store then joins a second time under the alias path.
- **Symlink scenario:** repo with tracked store `investigations/a` holding 3 events, then `ln -s investigations/a alink; cd alink; hyp check`. Output: `error: rejected Q-… in store investigations/a: duplicate id …` (3 errors, exit 1).
  - The alias path `alink` sorts before `investigations/a`, so the canonical copies are the ones rejected.
- **Letter-case scenario:** same result from `cd INVESTIGATIONS/a` on default macOS APFS, which ignores case.
- **Effect on other commands:** `hyp status --all` from there counts the duplicates under REJECTED. From the canonical path, `hyp check` reports 0 errors.
- `internal/cli/prompt.go` already notes that "the working directory may reach the store through a symbolic link", so this cwd is expected.
- **Fix:** resolve both `repo` and `cur` with `filepath.EvalSymlinks` before `Rel`. Or refuse the untracked-join branch when the folder is the same file (`os.SameFile`) as a store already listed.

**2. When a cross-store copy of an event sorts first, the duplicate rule rejects the wrong line.** Label: line 6. Sureness: high, reproduced. Severity: low.
- **Where:** `internal/replay/validate.go:58` with `replay.go:150`. `SeenIDs` is set only when an event is applied.
- **Scenario:** copy store a's DQ line into store b as actor `a0`, which sorts before `t1`. The b copy comes first and is rejected for crossing stores. The later line in a, with the same id, is then applied, not rejected.
- Line 6 says the later line is rejected. Rule 2 reads "duplicates an earlier event id".
- This semantic existed before inside one store, but cross-store copies make it newly reachable. The outcome is benign (the copy in the right store survives), but it does not match the line as written.

**3. A history cut deletes the base's `.tree/.gitattributes` and `.tree/.gitignore` of every store.** Label: none. Sureness: high, reproduced. Severity: low.
- **Where:** `internal/sandbox/sandbox.go:1135-1158` (`writeHistoryCut`).
  - It runs `os.RemoveAll(<store>/.tree)` for every store of the base.
  - `Cut.Write` then writes only `events/` and `evidence/`.
- **Effect:** in the clone, `git status` shows `D …/.tree/.gitattributes` and `D …/.tree/.gitignore` for each store.
  - The `merge=union` driver is lost.
  - `index.sqlite` is no longer ignored.
- The `--store` cut had the same gap, but only at the clone root, where the base usually holds no `.tree`. The history cut now strips these files from every tracked store.

**4. A `--store SRC --cut EVENT` clone leaks the base's later events through `--all`.** Label: none (behaves as the spec text says). Sureness: high, observed. Severity: low.
- **Scenario:** `--store investigations/a --cut QA` on a base that holds a's later events (and evidence `E-…` added after the cut).
  - In the clone root, `hyp status --all` reads the base's `investigations/*` stores, not the root cut.
  - So it shows Goal C, Goal B and the later Experiment, all after the cut.
- Before this change, only a runner who `cd`'d into the base copy saw them. Now `--all`, and `hyp status G-…` for a Goal the cut does not hold, reach them from the clone root.

**5. A new untracked store and a tracked store can append concurrently with interleaved ids.** Label: line 5, borderline. Sureness: medium, by reading. Severity: low.
- A writer in a tracked store does not see a just-initialised untracked store; it joins only from its own folder.
- Both writers lock the repo root, but the tracked writer's `last` id leaves out the new store's ids. Two appends in the same millisecond can therefore sort before the other store's earlier append.
- Once the new store is tracked, the order no longer matches the order of appends. By the definition of "history" from that folder this is allowed, so I note it only.

**Minor:** in `History.AppendWith`, when `build` fails and `target()` also fails, the build error wins. Example: an answer drafted with keys in two stores that also breaks another CLI rule is refused for the other rule, not the store rule.

### Probed and correct
- **Cross-store refusals (exit 1, naming the rule), refused at write and rejected on replay (hand-written lines):** node add with a parent in another store; checkpoint on another store's Goal; `hyp add goal` where discovery finds no store.
- **From the repo root:** a write goes to its node's store; an alias two stores hold gives `ambiguous node` (exit 2).
- **Cross-store cites:** a settled cite passes. An unsettled Hypothesis cite is refused with exit 1. `hyp show` prints `cited by: G-… (store …)`.
- **Line 5 concurrency:** 30 parallel appends across two stores, with id order equal to ts order and `hyp check` at 0 errors.
- **Duplicate ids across stores:** the later store-path line is rejected, and `hyp check` names the store.
- **Line 9:** at this repo root, `hyp check` reads the five stores under `investigations/` and reports `0 errors` (47 warnings). It leaves out the testdata store and the trial stores under `.tree/`.
- **History cut:**
  - At a DQ (base C2): only a's store remains, holding 2 lines; b and c are gone; the post-cut evidence is gone.
  - At H2 (base C1): b is created in its own folder and marked intent-to-add. The clone's `hyp status --all` equals the host's `hyp status --all --as-of H2`.
  - At E1: the evidence is kept.
  - An unknown cut event exits 2.

### Verdict per line
1. **Met.** Only the folder's store's Goals are shown, and a Goal's store comes from its add.
2. **Met.** `--all` works from anywhere, including the testdata and trial folders, and `hyp status G-…` falls back to the repo history.
3. **Met for the listed exclusions.** But a symlinked or wrong-case cwd adds a store a second time (finding 1).
4. **Met.**
5. **Met in one clone.** Caveat: the untracked-store window (finding 5).
6. **Met for the plain case.** Not met when a symlink or case alias creates false duplicates (finding 1). The wrong line is rejected when the earlier copy is invalid (finding 2).
7. **Met.**
8. **Met.**
9. **Met** (0 rejected events).
10. **Met for the events and evidence.** The clone loses each store's `.gitattributes` and `.gitignore` (finding 3).

Status: DONE_WITH_CONCERNS
Summary: The candidate meets all ten lines in my probes, and `go vet`, `go test` and a repo-root `hyp check` are all clean. One real defect remains: a symlinked or wrong-case cwd reads a store twice and rejects all its events as duplicates (`internal/store/history.go:94-109`). The other findings are low-severity sandbox and edge-ordering issues.
