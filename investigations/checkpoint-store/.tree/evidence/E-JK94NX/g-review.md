Review of commit a9963fc, "feat: read a repository's stores as one history" (parent eed0938), checked against G-AKJJF8 lines 1–10.

How I checked: I read the full diff, ran `go vet ./...` and `go test ./... -count=1` (both pass), and built the candidate into the scratchpad. I ran it in throwaway repositories and read-only in this repository. `hyp check` at the repository root reads the five `investigations/*` stores and prints `check: 0 errors, 41 warnings`. It leaves out the testdata stores and the trial stores under `.tree/evidence`.

## Findings (most severe first)

1. **line 4 (also the hyp-owner-grill skill)** — `internal/store/history.go:58-86` (`OpenHistory`) and `:98-127` (`repoStores`). **Sure: confirmed by a probe.**
   - **Cause:** a store counts in the history only when git tracks one of its events files. A store just made by `hyp init` has no events file, so git tracks none, and every write in it reads that store alone. The store's first event therefore can never cite another store.
   - **Scenario:** repository with `inv/a` (Goal G-TAYPDG, tracked) and a fresh `inv/b` after `hyp init`. In `inv/b`:
     - `hyp add goal --title B --line x --cite 1=G-TAYPDG#1 --why w` gives `unknown node G-TAYPDG`, exit 2.
     - `hyp add goal --body "see G-TAYPDG" ...` gives `--body names G-TAYPDG, which is no node here; ... write <name>:<id> for another store's`, exit 2.
     - Running `git add -A` first does not help, because nothing is written yet. It works only after a Goal with no cite is appended and then `git add`ed.
   - **Who hits it:** hyp-owner-grill's new-store path (SKILL.md row 105, `hyp init`). The new text at `skills/hyp-owner-grill/SKILL.md:122` says a body "may name the id of a node in another store of this repository as it is". That fails on exactly this path.
   - **Line 5 too:** until the first `git add`, such a store's appends lock only its own `.tree/events` and take ids after its own events only.

2. **line 2** — `internal/store/history.go:76-84` (`Single(cur)` when the store found is not in the history), used by `cmdStatus` in `internal/cli/read.go`. **Sure: confirmed in this repository.**
   - **Problem:** line 2 says `hyp status --all` and `hyp status G-id` work "run anywhere in the repository". In any folder whose discovered store adds nothing to the history, they read only that store instead. Such folders are a `testdata/` store, a trial store under `.tree/evidence/...`, or a fresh untracked store.
   - **Scenario:** in `internal/testutil/testdata/synth`, `hyp status G-AKJJF8` gives `unknown node G-AKJJF8`, exit 2. `hyp status --all` there prints only the synthetic store's Goal G-KMMKHH.
   - **Note:** replay-spec §1 and ADR 0023 state this fallback on purpose ("so it is when the store the walk finds adds nothing…"). So the spec disagrees with line 2's "anywhere", not only the code. No test covers status run from such a folder.

3. **line 8 (tests)** — `internal/cli/history_test.go:395-454`. **Sure.**
   - **No select case:** line 8 names "Result, select, edge, refer or checkpoint". The write-time loop has no `hyp select` case. The replay loop has no `dq.select` line and no checkpoint with `refers` into another store; it has node.add, result.record, edge.add, a plain checkpoint, abandon and answer.
   - **Weak checks:** every write-time case in the loop checks only exit 1 with an empty substring (`want(..., 1, "")`). The crossing rule is checked by message only for the first `add dq`. A refusal for an unrelated reason would also pass.
   - The code looks right: `checkStore` checks `ev.Node` and the checkpoint refers keys. The gap is in the tests.

4. **line 4 (tests)** — `internal/cli/history_test.go:278-280`. **Sure.**
   - The "not settled" refusals for `--cite R#1` and `--reading-of R#r1` check only exit 1 with an empty message.
   - Behaviour is correct: a probe shows `refused: bad cite: line 1 cites R-…#1, which is not settled: the limit is neither closed nor accepted`, exit 1. The test would also pass on any other exit-1 refusal.

5. **none** — `internal/cli/idcite.go:99`. **Sure.** The refusal hint still says `cite an id this store holds, or write <name>:<id> for another store's`. That contradicts replay-spec §5.1, the glossary's Cite entry and the new skill text, which all say a `<name>:` token is only for a tree outside the repository and that other stores' ids are checked as they are.

6. **none (spec text vs code)** — `docs/replay-spec.md` §1 versus `internal/store/history.go:262-275` (`lock`). **Fairly sure.**
   - **Spec:** the lock is on the store's `.tree/events` "when the history is one store alone".
   - **Code:** it uses the events dir only when `Stores == ["."]`. A repository whose only tracked store is `investigations/a` locks the repository root instead.
   - Harmless as long as every writer sees the same history; the spec wording just does not match.

7. **line 8 (ambiguity, low)** — `internal/store/history.go:290-308` (`target`). **Sure about the behaviour; unsure it is a defect.**
   - **Scenario:** in store a's folder, `hyp answer --answer <R of store b>#r1=…` (or `add dq --parent <G of b>`) is refused as `event crosses stores`. It is not appended to b.
   - Line 8's first sentence ("it is appended to the store that holds its node") could be read to route it to b. replay-spec §5.1 and the tests choose "the current store, else refuse". The owner may want to confirm that reading.

8. **line 1 (low)** — `internal/render/status.go` StatusOf/goalsTable (`REJECTED %d events`) and `cmdEvents` in `internal/cli/read.go`. **Sure.**
   - `hyp status` in a store's folder shows the Goals of that store only, but its `REJECTED N events` counts rejections in the whole history.
   - `hyp events --json` filters rejections by event id, so when two stores hold the same id, a rejection from the other store shows up in this one.

## Per line
- line 1: no defect against the line; a minor side note (finding 8).
- line 2: defect — `--all` and `status G-id` do not work from a folder whose store is outside the history (finding 2).
- line 3: no defect found; each exclusion is implemented and tested.
- line 4: defect — a fresh store's first event cannot cite another store (finding 1); the not-settled tests only check exit 1 (finding 4).
- line 5: no defect in one clone with tracked stores (concurrency test present); a fresh untracked store is outside the lock and the id rule until tracked (finding 1).
- line 6: no defect found; tests cover the store-path and actor ordering.
- line 7: no defect found; alias scoping and the verdict-label binding are tested, including 20 repeated runs against random map order.
- line 8: tests incomplete for select and refer (finding 3); possible routing ambiguity (finding 7).
- line 9: no defect — the candidate's `hyp check` at the root reads the five stores and reports 0 errors (0 rejected).
- line 10: no defect found; the history-cut test checks per-store events, evidence pruning, a store the base lacks, and status equal to the host's `--as-of` view.

Status: DONE
Summary: Two real defects: a brand-new store can never cite another store with its first event (line 4 and the hyp-owner-grill skill), and `status --all` / `status G-id` from a testdata, trial or untracked store folder ignore the history (line 2). The rest are test gaps for line 8's select/refer and line 4's refusal messages, plus a stale hint and spec text that no longer match the code.
