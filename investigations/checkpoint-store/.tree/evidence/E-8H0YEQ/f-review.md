Review of 9643341 (fix on ded3e4b) against G-1391DM

**What I ran.** I ran `go vet ./...` and `go test ./... -count=1` on the working tree at 9643341, and both pass. I also ran the new TestAppendAfterRejectedID against ded3e4b's code, in a `git archive` copy under the scratchpad, which I deleted afterwards. Nothing in the repository was changed.

**How the targeted cases are closed**
- **Old 1(b) and 1(c): closed.** The junk id "zz" and the 25-character "7ZZ…Z" no longer block appends. On ded3e4b the new test fails for both ids with "refused: no ULID sorts after …". On 9643341 it passes.
- **Old 1(a): closed under the run's reading.** A lower-case ULID no longer sets the newest id. That the appended id may still sort before such a line is the reading that goes to the owner, so I do not re-report it.
- **How the fix works.** `AppendWith` (internal/store/store.go:275-283) walks the merged list from the end and takes the first id that matches `ulidRe`. The list from `Events` is in id order, so that id is the greatest id of the valid form, which is what the new comment and spec §2 say. Now that `last` is always "" or a valid-form id, the `""` return of `ParseStrict` in idAfter fires only for an id whose first character is above 7. That is exactly "an id of that form above" the greatest ULID, as spec §2 and the comment at store.go:208-210 say.
- **Stale comment: closed.** replay.go:11 now names MergeOrder, and the MergeOrder doc says `tie` must not be nil.
- **Shared id in the greatest millisecond: closed.** Spec §2 now says that two clones may share the +1 id there.

**Findings**

1. **G-1391DM#5: one case of the new test passes by construction.** internal/store/store_layer_test.go:284-293. The test only checks `ev.ID > g.ID`. For the lower-case case "01mzzzzzzzzzzzzzzzzzzzzzzz", ded3e4b's code also passes it. There, ParseStrict decoded the lower-case id, and the new id became `01N0000…`, which is still above g.ID (`01M79…`). I confirmed this: on ded3e4b only "zz" and "7ZZ…Z" fail. So nothing pins the actual change for lower-case lines, which is that they no longer push the new id into the millisecond after them. To pin it, the test could also assert that the new id's time is the clock's millisecond, or that `ev.ID < "01N"`.

2. **none: minor spec wording.** docs/replay-spec.md:29 says a line whose id has another form "is rejected by replay (§5)". But §5's list (docs/replay-spec.md:104-105) does not name a wrongly formed id. Rule 1 says "a required field is missing", and rule 2 covers duplicates. The code rejects it as "missing field: id must be a 26-character ULID" (validate.go:49-50), so rule 1 covers it only loosely. One clause in rule 1, such as "or `id` is not a 26-character upper-case ULID", would make the cross-reference exact.

3. **none: very minor.** ADR 0022:24 still says "an appended id sorts after every id the store holds" and does not say the newest id counts only ids of the valid form. The spec owns the detail and the ADR gives the reason, so this is optional. I note it only because it is the same wording behind the reading that goes to the owner.

**Where I found nothing**
- **No regressions.** A line of valid form that replay rejects for another reason (a duplicate, a bad ts) still counts as the newest id. That is conservative and matches "the greatest id of the form above".
- **Exit codes.** The refusal still goes through herr.Refused (exit 1).
- **Other uses of the id check.** `ValidID` reuses the same `ulidRe` that Validate uses, so the two checks cannot drift apart. No other code reads the newest id or calls ParseStrict.
- **Other text.** Nothing in skills/, README.md, AGENTS.md or the rest of docs/ mentions the newest-id rule, idAfter or replay.Merge. No skill cites an ADR.
- **Not re-reported.** I did not re-report the owner reading for line 5, the file:line naming, or the lsof skip.

Status: DONE
Summary: 9643341 closes the junk-id and wrong-length blocking cases (its test fails on ded3e4b for both), fixes the stale replay.At comment, and states the shared top-millisecond id in spec §2. The only gaps left are small: the lower-case case of the new test passes by construction, and §5 does not explicitly list a wrongly formed id, which §2 now cites.
