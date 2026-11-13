Review of 7a05e9e (fixes on 10e1925) against G-1391DM. On 7a05e9e, `go vet ./...` and `go test ./... -count=1` pass. To check the new tests are red on base, I ran them against a `git archive` copy of 10e1925 in the scratchpad. Both fail there: TestSameIDLinesAnyOrder fails on the title and on the bytes, and TestAppendInGreatestMillisecond fails on the same id and on the missing refusal. So neither passes by construction. I also ran one probe test in a scratchpad copy of 7a05e9e. Nothing in the repository was changed.

**How the targeted findings are closed**
- **Old finding 1 (same-id tie): closed.** `store.Events` (store.go:161-173) breaks the tie on the trimmed line bytes through `byLine`. `CutAt` (cut.go:75-82) breaks it on the trimmed line first, then on the raw line. The first key is the same in both paths, and the parsed `Actor` no longer takes part. Every production reader goes through `store.Events` (check.go:20, cli/read.go:54, cli/write.go:566, AppendWith, LoadAsOf), so status, `--as-of`, check and append all agree with the cut. When two lines are equal after trimming, they parse to the same event, so an unstable sort changes nothing that anyone can see. The raw-byte key makes the cut's bytes deterministic.
- **Old finding 3 (AGENTS.md): closed.** AGENTS.md:64 now says "commit it if it changed".
- **Old finding 5 (greatest millisecond): partly closed.** Two clones now differ, and the all-Z id is refused with exit 1 (herr.Refused). Findings 1 and 2 below are what remains.

**Findings**

1. **G-1391DM#5: an id above the greatest ULID still gets an append that sorts before it, with no refusal.** internal/store/store.go:214-216.
   - `idAfter` ignores a `last` that `ulid.ParseStrict` rejects and returns the clock id.
   - Replay's id check (replay/validate.go:17) is only the regex `^[0-9A-HJKMNP-TV-Z]{26}$`. It accepts a first character from 8 to Z, which overflows 128 bits and which ParseStrict rejects.
   - Probe: a store holding a Goal with id `80000000000000000000000000` or `ZZZZZZZZZZZZZZZZZZZZZZZZZZ` replays with that node applied. `Append` then succeeds with id `01M79Y…`, which sorts before it.
   - This breaks line #5 ("sorts after every event id the store holds") and spec §2. It is the same far-future hand-edit case this commit set out to close.
   - Fix: refuse when `last` does not parse but `newID(t) <= last`, or tighten the regex so the first character is 0-7.

2. **none (spec and code text): "no id sorts after" is not what the refusal tests.** store.go:221-232; docs/replay-spec.md:29; the error text at store.go:282.
   - The refusal fires when last + 1 + rand32 overflows 80 bits, not only when last is the greatest ULID.
   - Probe: with `last = 7ZZZZZZZZZZZZZZZZZZZZZZZZY`, the append is refused with "no event id sorts after …", although `…ZZZ` sorts after it.
   - The result is also not deterministic: a retry can succeed when the random step is small.
   - The spec sentence "when they would overflow, no id sorts after the newest" is therefore false, and so is the comment "returns "" when no id sorts after last".
   - Severity: negligible in practice (year 10889). The fix is to fall back to last+1 when last+1 does not overflow, or to reword the spec and the message.

3. **none (exit code, negligible).** store.go:225-227. A `crypto/rand.Read` failure also returns "", so it is reported as a rule refusal (exit 1) with a false message. By the repository rule it would be exit 3. Under go 1.27, `crypto/rand.Read` does not return an error, so this path cannot happen. It only makes the comment's "or when the random bits cannot be read" misleading.

4. **G-1391DM#4 (test gap).** internal/replay/replay.go:19-46.
   - `replay.Merge` and `MergeOrder(files, nil)` still break ties by JSON encoding. No production code calls them now: every reader uses `store.Events` or `CutAt`.
   - TestMergeSameIDSameActor (replay_test.go:534-539) therefore tests a tie rule that no command uses.
   - The new store test does cover the production paths, so nothing is untested. But the exported nil-tie default is a second, different tie rule that disagrees with spec §6 ("ordered by their line bytes").
   - Fix: drop the nil default, or make the replay-level test exercise `byLine`.

**Where I found nothing**
- I found no regressions in `readLines`/`parseFile`: the lines are trimmed and cloned per index, and the actor key matches the cut's.
- The cut still stops at the first event with id `at`, as `asOf` does.
- No other code generates ids, so no path skips `idAfter`.
- ADR 0022:23-24 and replay-spec §6 match the code. In §2, only the overflow sentence is wrong (finding 2).
- No text in skills/ or README.md, and no other docs/ text, mentions the tie rule or the id step. No skill cites an ADR.
- I did not re-report the unparseable-line file:line case (R-ZN1VJJ#2). The earlier note that the line #1 test skips without `lsof` is not touched by this diff.
- Untracked, not created by me: `investigations/checkpoint-store/.tree/evidence/E-HJ8SQ9/` (new since the snapshot), plus the E-HSEK8J/E-RF2A4J folders and `claude-g138-1.jsonl` that the snapshot already showed.
- Probe copies are under /private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/a4aec036-c01b-4976-9dc8-6902778407b3/scratchpad/ (`old/` and `new/`).

Status: DONE_WITH_CONCERNS
Summary: 7a05e9e closes the same-id tie (replay and cut now agree on line bytes, and both new tests are red on 10e1925) and the AGENTS.md commit step. One line #5 case is still open: a hand-written id above the greatest ULID (first character 8-Z, which replay accepts) still gets an appended id that sorts before it with no refusal. The overflow refusal can also fire when an id would still fit, which makes the §2 wording and the error message wrong.
