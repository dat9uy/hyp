## Review: `git diff 07d44c0 d966e85` (replay-spec rule 1, ValidID comment, TestAppendAfterRejectedIDKeepsClock)

I found no defect in the diff. The test still guards line 5's owner ruling, it is not flaky, and the new doc and comment texts match the code.

**Does the test still check the ruling, and for how long? (line 5)**
- **Yes, it still checks it.** I copied the repo to `/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/7ee8bf30-413d-4dfc-963d-ea43faef812f/scratchpad/mut2` and changed the `replay.ValidID` filter in `AppendWith` (`internal/store/store.go:279`) to `if true`, so the rejected id counts as the newest. The test then fails: the appended id is `100000000059GFSMHS7R4VA1E0` at millisecond 35184372088832, while the clock window was 1794558781149–1794558781173.
- **Why it fails that way:** `ulid.ParseStrict` accepts lower case. It decodes `0zzzzzzzzzzzzzzzzzzzzzzzzz` to millisecond 2^45−1, and `idAfter` then takes millisecond 2^45. If a later `oklog/ulid` version rejected lower case, `idAfter` would return "" and Append would refuse. The test's `t.Fatal(err)` would still catch that, so it guards either way.
- **How long it keeps working:** until the clock reaches millisecond 2^45, which is 3084-12-12T12:41:28.832Z. From then on, clock ids start with `1` and sort above `0z…`, so `newID > last` holds no matter what and the test stops telling the two behaviours apart. The comment's "until about the year 3084" is correct.
- **Compared with the old fixture:** `01mzzz…` would have stopped telling them apart once clock ids reached the `02` prefix, around 2039. The change is a real extension.

**Is it flaky? (line 5)**
- No. With the store holding only the rejected line, `last` is "" and the id is `newID(time.Now().UTC())`, read between `before` and `after`.
- The only way it could fail is the wall clock stepping backward (for example an NTP correction) during Append, which takes a few milliseconds. That is negligible.
- I ran `go test -run RejectedID -count=50` and it passed every time. `go vet` is clean on `internal/store` and `internal/replay`.

**Are the doc and comment texts accurate? (line 5 / none)**
- **ValidID comment (none):** "26 characters of Crockford base32 in upper case" matches `ulidRe` `^[0-9A-HJKMNP-TV-Z]{26}$`, which leaves out I, L, O and U. "May sort above the greatest ULID" is true for ids starting 8–Z.
- **replay-spec rule 1 (line 5):** the claim "an id of that form above the greatest ULID passes, and an append after it is refused" matches the code. `validateEnvelope` checks only the regex. In `idAfter`, every clock id sorts below such an id, `ParseStrict` then fails on overflow, it returns "", and `AppendWith` returns `herr.Refused` before writing. `TestAppendAfter…`'s refused case already covers this.
- **Test comment (line 5):** "sorts after every upper-case id whose first character is 0" is correct, because `z` (0x7A) is above every upper-case character.

**Findings**
1. **none (Low, outside the diff):** the error text at `internal/replay/validate.go:50` still says "id must be a 26-character ULID". After this diff the spec and the ValidID comment say the form is 26 upper-case Crockford base32 characters, which also admits ids above the greatest ULID. The message is now slightly inaccurate for those ids, though it is harmless because it only fires on ids that fail the regex. If you want consistent wording, change it to "26 upper-case Crockford base32 characters".
2. **line 5 (Informational):** `TestAppendAfterRejectedID`'s case `01mzzzzzzzzzzzzzzzzzzzzzzz` checks only that the appended id is greater than `g.ID`, so it never depended on how long the fixture keeps working. Nothing to change.
3. **none (Informational):** rule 1 sits in the list "An event is refused when…", but its new bracket describes an append-time refusal, not a replay one. The text is accurate, just mixed into a replay list. Optional: move that clause to the line 5 / append section of the spec.

Status: DONE
Summary: The diff has no defect. The test, proven by mutation, catches a rejected lower-case id moving the appended id off the clock until 3084-12-12 and is not flaky; the doc and comment texts match the code. The only nit is that validate.go:50's error message still says "26-character ULID".
