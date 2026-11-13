Review of 10e1925 (fixes on 65a1cc5) against G-1391DM. Base is a7dbc00. On 10e1925, `go vet ./...` and `go test ./... -count=1` pass. I ran two probes: one Go test in a throwaway copy of the tree, and one scratch git repository. I deleted both afterwards and changed nothing in the repository.

**How the three targeted findings are closed**
- **Old finding 1 (#5, same id on two clones): closed.** When the clock is behind, `idAfter` now uses the millisecond after `last` with fresh random bits from crypto/rand (internal/store/store.go:193-195). Two clones whose clocks are behind get different ids, and both sort after `last`. TestClockBehindOnTwoClones fails on 65a1cc5, because the duplicate id is rejected. It does not pass by construction: each clone draws fresh entropy, because the monotonic reader is reset when the millisecond changes.
- **Old finding 2 (spec said "file order"): closed in the text.** The spec and ADR now describe the new tie-break. Finding 1 below shows the tie-break is not fully order-free.
- **Old finding 4 (AGENTS.md merge procedure): closed.** In a scratch repository, a PR branch had no `.gitattributes`, and main had added it and appended to the same events file. Running checkout, then `git commit`, then `git merge origin/main` exits 0 and leaves `l0, pr1, main1`.

**Findings**

1. **G-1391DM#4 (and #1): the new tie-break re-encodes the parsed event, and the cut and replay encode it differently.**
   - Where: internal/replay/replay.go:49-51 and internal/store/cut.go:65-66.
   - What is wrong: `MergeOrder` breaks a same-id, same-actor tie by `json.Marshal` of the parsed event. `store.ReadAll` keeps the `actor` field from the line. `CutAt` overwrites `ev.Actor` with the file's actor. The comment at cut.go:65 ("Merge never reads Actor") became false with this commit.
   - Scenario (reproduced): `a.jsonl` holds two lines with id X. Line 1 has `"actor":"b"` and title x. Line 2 has `"actor":"a"` and title y.
     - `hyp status --as-of X` applies title y.
     - The cut at X keeps line 1, and its replay shows title x.
     - So the clone's status differs from `--as-of`, which breaks line #1.
   - A second case: if two such lines have equal encodings but different bytes (key order, whitespace), the comparator returns 0, and the cut's bytes depend on the line order. I reproduced this: lines `{"v":1,"id":…}` and `{"id":…,"v":1}` cut to different bytes depending on which comes first. That breaks #4 ("writes the same bytes").
   - Severity: low. After this commit, same-id lines no longer come from clones and only come from hand edits. But spec §6 (docs/replay-spec.md:158) and ADR 0022:23 say "file order decides nothing", which is not true in these cases.
   - Fix: break ties on the raw line bytes, and use the same actor in both paths.

2. **G-1391DM#4 (test gap).** internal/replay/replay_test.go:534 (TestMergeSameIDSameActor) tests only `Merge` on events built in Go. No test covers the same-id tie through `CutAt`/`Write` or through raw lines, so finding 1 passes the suite.

3. **none (AGENTS.md:64 wording).** "then `git commit`" fails when the PR branch already holds main's `.tree/.gitattributes`. That is the common case once every store has the file, and GitHub still conflicts then because it never runs the union driver. In the scratch repository, `git commit` printed "nothing to commit" and exited 1. A reader who chains the steps with `&&` stops there.
   - Fix: add "if it changed" to the commit step.

4. **G-1391DM#4 (still open from the earlier review, finding 3).** A line that does not parse is still reported by line number by `hyp check` (internal/check/check.go:36) and in `hyp status`'s skip warning. The same lines in another order therefore print differently. 10e1925 does not address it, and it says nothing about it.

5. **G-1391DM#5 (residual, negligible).** When `last` is in the greatest ULID millisecond, the old increment loop still runs. It gives the same id on two clones. If `last` is the greatest ULID, it wraps to all zeros, which sorts first. Only a hand-written id far in the future reaches this.

**Where I found nothing**
- The new id rule always sorts after `last`, keeps append order equal to id order within one store, and keeps `ts` as the clock time.
- I found no regression in the line #6 union-merge test.
- The ADR 0022, replay-spec §2 and §6, and the AGENTS.md text match the code, apart from findings 1 and 3.
- No other text in skills/, README.md or docs/ mentions "plus one" or file-order ties.
- No skill cites an ADR, and no exit code changed.
- The earlier review's finding 5 (the line #1 test skips without `lsof`) still stands.

An untracked folder `investigations/checkpoint-store/.tree/evidence/E-RF2A4J/` appeared during my run. I did not create it.

Status: DONE_WITH_CONCERNS
Summary: 10e1925 closes the three targeted findings: two clones whose clocks are behind now get different ids, the spec text is corrected, and the AGENTS.md merge procedure works. The new same-id tie-break re-encodes the parsed event, and the cut sets the actor differently from replay, so a hand-made store can still cut differently from `--as-of` or by line order. AGENTS.md's `git commit` step also fails when there is nothing to commit.
