Review of ded3e4b (fixes on 7a05e9e) against G-1391DM

I ran `go vet ./...` and `go test ./... -count=1` on a `git archive` copy of ded3e4b, and both pass. I also ran ded3e4b's TestAppendInGreatestMillisecond against 7a05e9e's code, where it fails: `7ZZZZZZZZZZZZZZZZZZZZZZZZY` is refused there. So the new case does not pass by construction. I used one probe test in a scratchpad copy and deleted both copies afterwards. Nothing in the repository was changed.

**How the four targeted findings are closed**
- **Old finding 1 (id above the greatest ULID): closed for ids the id pattern admits.** `80000…` and `ZZZ…` are now refused with exit 1 and nothing is written (test case `80000000000000000000000000`). Finding 1 below covers what is still open.
- **Old finding 2 (refusal while an id would still fit): closed.** When the random step would overflow, the id falls back to +1. `…ZY` now appends `…ZZ`. The texts in spec §2, ADR 0022:24, the idAfter comment and the error message now match that behaviour.
- **Old finding 3 (crypto/rand failure): closed.** That error path is gone.
- **Old finding 4 (second tie rule in replay.Merge): closed.** `Merge` and the nil JSON default are removed. The replay test helper panics if the tie is ever called, and every remaining `MergeOrder` caller passes the line-bytes tie (store.go:166, cut.go:82). This agrees with spec §6.

**Findings**

1. **G-1391DM#5: `idAfter` trusts `ulid.ParseStrict` to agree with string order, and it does not for ids that replay rejects.** internal/store/store.go:215-218; the comment is at store.go:208-209 and the spec at docs/replay-spec.md:29.
   - **Cause.** ParseStrict decodes lower case and refuses a wrong length. Replay's own check (validate.go:17,45) rejects both kinds of line, but `store.Events` still returns them, so one can be `last`. The function never checks that its result is greater than `last`.
   - **Probe on ded3e4b, one rejected Goal line, then `Append`:**
     - (a) `last = "01mzzzzzzzzzzzzzzzzzzzzzzz"` (lower-case ULID): the append succeeds with `01N0000000…`, which sorts before `last`. This is the same failure as old finding 1, reached through a lower-case id.
     - (b) `last = "zz"`: the append is refused with "no ULID sorts after zz". This is a regression from 7a05e9e, which ignored a non-ULID `last`. One hand-edited junk line that `hyp check` already reports now blocks every write to the store until someone deletes the line by hand.
     - (c) `last = "7ZZZZZZZZZZZZZZZZZZZZZZZZ"` (25 characters) or a 26-character id with a trailing space: refused, although real ULIDs sort after it. The message is then false.
   - **Spec disagreement.** §2 and the comment say the refusal happens only for "the greatest ULID, or an id above it that the id pattern admits". The code refuses more than that.
   - **Fix.** Compute `last` only over ids that match the id pattern, the way replay does. Or make idAfter end with a guard that returns its id only when the id is greater than `last`, and decide what a rejected id should do.
   - **Severity.** Low, because it needs a hand-edited line. Case (b) is the only change in behaviour.

2. **none (stale comment).** internal/replay/replay.go:11 still says "At names one event of the files Merge reads". `Merge` no longer exists. `MergeOrder` (replay.go:18-24) is exported, and its doc does not say that `tie` must be non-nil. With nil it would panic on the first same-id, same-actor pair. Today no caller passes nil.

3. **none (minor, spec wording).** At the very top of the ULID range, where only the +1 fallback fits (for example after `…ZY`), two clones append the same id. That contradicts §2's "two clones … append different ids", which is stated without that exception. The test case accepts the collision (`differ: false`). It matters only around the year 10889 and is harmless, because a union merge keeps both lines and the line-bytes tie orders them. I mention it only so the spec sentence is not read as unconditional.

**Where I found nothing**
- No regression in the cut, the replay tie, or `Events`/`LoadAsOf`.
- Exit codes are correct: the refusal goes through herr.Refused (exit 1).
- ADR 0022:24 matches the code, except for the scope point in finding 1.
- Nothing in skills/, README.md, AGENTS.md or the rest of docs/ mentions `replay.Merge`, the id step or the overflow rule, so the diff makes no other text wrong. No skill cites an ADR.
- I did not re-report R-ZN1VJJ#2 or the lsof skip.
- Untracked folders I did not create: `E-FB7QAF/` (new since the snapshot), `E-HJ8SQ9/`, `E-HSEK8J/`, `E-RF2A4J/` and `claude-g138-1.jsonl`.

Status: DONE_WITH_CONCERNS
Summary: ded3e4b closes the four targeted cases for ids the id pattern admits, and its new test cases fail on 7a05e9e. idAfter still relies on ParseStrict instead of string order: a rejected lower-case id gets an appended id that sorts before it (line #5), and any non-ULID id that sorts high now blocks every append. The second is a regression from 7a05e9e, and the code refuses in more cases than spec §2 states.
