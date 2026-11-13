## Review: diff 9643341..07d44c0 (commits e7207d6 and 07d44c0)

I found no blocking defect. The new test is correct and catches the mutation it targets. The doc changes match the code, with one wording imprecision.

### What I checked
- `go test ./internal/store/ -run TestAppend -count=20` passes. `go vet ./internal/store/` is clean, and the gofmt fix in `store_layer_test.go` is correct.
- **Mutation check:** I copied the repo to `/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/7ee8bf30-413d-4dfc-963d-ea43faef812f/scratchpad/mut`, leaving the real repo untouched. There I changed the newest-id loop in `AppendWith` (`internal/store/store.go:279`) from `if replay.ValidID(evs[i].ID)` to `if true`.
  - The new test fails on the mutation: "appended id 01N0000000NFSZ6KV08BWT2T9C has millisecond 1821066133504, want the clock's, 1794558536112 to 1794558536116".
  - The older `TestAppendAfterRejectedID` does not catch the lower-case case. Under the mutation it fails only for `"zz"` and `"7ZZZ…"` (25 chars), because for `01mzzz…` the id still sorts after `g.ID`.
  - So the new test fills a real gap.

### Findings

1. **[line 5] Test logic: correct, not flaky in practice.**
   - `01mzzz…` is the last element of `evs` because `MergeOrder` compares bytes and `'m' > 'M'`. Without the `ValidID` filter it becomes `last`.
   - `ulid.ParseStrict` accepts lower case, so the mutated path gives ms+1 of `01MZZZZZZZ`, about 2027-09-15. That falls outside `[before, after]`, and stays outside after that date, because ms+1 is a fixed past millisecond.
   - `before` and `after` bracket the `time.Now()` that `AppendWith` reads. The only flake source is a backward wall-clock step (NTP) between the reads, which is negligible.
   - The `ulid.Monotonic` entropy increments within a millisecond and does not change the millisecond.

2. **[line 5] Low: the test has a shelf life.** The comment says the lower-case id sorts after "every upper-case id of today". That stays true until ids start with `02` (about 2039). After that, the valid id sorts after `01m…`, the mutation no longer changes `last`, and the test passes without testing anything. No action needed now; you could say so in the comment, or build the bad id from the current time's prefix in lower case.

3. **[line 5] Low: doc wording.**
   - `docs/replay-spec.md` rule 1 now says "`id` is not a 26-character upper-case ULID". The check is `ulidRe = ^[0-9A-HJKMNP-TV-Z]{26}$` (`internal/replay/validate.go:17`), which does not enforce ULID's first-character bound (≤ `7`).
   - So a 26-character id starting `8`–`Z` passes replay, though it is not a ULID.
   - `AppendWith` counts it as the newest id. `idAfter` then gets `ErrOverflow` from `ParseStrict` and the append is refused ("no ULID sorts after …"). One such line therefore blocks every later append to the store.
   - The `idAfter` comment documents this ("an id of the id form above it"), and it behaves as written, so it is not a new defect from this diff. But the spec sentence overstates what replay rejects. Either say "26 characters of Crockford base32 in upper case", or reject ids above the maximum ULID in `validateEnvelope`. The second fits the owner's "validate bad data in the first place".

4. **[line 5] The ADR text is accurate.**
   - "counting the ids of the event-id form; replay rejects a line whose id has another form" matches `AppendWith`: the greatest `ValidID` id in merge order.
   - Valid-form ids that replay rejects for other reasons (a duplicate, a missing parent) still count. That is conservative: it only pushes the new id later. It is consistent with the owner's answer, which makes counting rejected lines optional.
   - Lines that fail JSON parsing go to the `LineError`s and never reach `evs`, so they do not count either.

5. **[none] The `replay-spec.md` addition matches the code.** `validateEnvelope` checks the id form after `v` and type and before the duplicate check, so placing it in rule 1 is right. The actor-form refusal is still not listed in rule 1; that gap predates this diff.

6. **[none] Outside the diff:** `gofmt -l internal/` still lists `internal/recipe/first_test.go` as unformatted. It was last touched by the initial release, not by this diff.

Status: DONE
Summary: TestAppendAfterRejectedIDKeepsClock is correct and not flaky in practice: it fails when the ValidID filter is removed, which the older test does not detect for lower-case ids. The only issues are low: the spec says "upper-case ULID" where the check accepts 26-character ids above the maximum ULID (one of which blocks all appends), and the test stops testing anything around 2039.
