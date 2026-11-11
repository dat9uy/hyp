# Review: 3ec4979..cbd734a (internal, docs): one subject per run, patch kept apart from inputs

Judged against: ADR 0020 "A subject has one shape", Goal line 3, and the owner's reading of line 3 (inputs are data, so a run at B given inputs covers bare B).
`go vet ./...` and `go test ./... -count=1` pass at cbd734a. I ran them in a detached temp worktree, which I then removed.
The three points already known were not re-reported.

## Findings

1. If a sandbox.json Arm's `sha256` is not lowercase hex (for example uppercase or empty), `runSubjects` turns `patch` into "" and records `{base:B}`. That patched run then covers a bare-B Hypothesis (`hyp status` prints `covered`), which line 3 forbids. Confirmed with a temporary CLI test.
2. `hyp subject H arm.patch --base B --patch arm.patch` is accepted with no rule 26 "listed twice" refusal, and it stores the one path as both patch and input. A normal sandbox run at B plus that Arm then never covers H (`uncovered on Q2: H`). Confirmed with a temporary CLI test.
3. replay-spec rule 28 now says a `records` path named twice is refused with `bad subject: <path> listed twice`. The code (validate.go:225) emits `bad records: <path> named twice`, so the commit pasted rule 26's text into rule 28. Confirmed by grep.
4. replay-spec §6 still writes the projection as `Subjects map[string][]SubjectFile`. The code holds `map[string]event.HypothesisSubject`, which includes base and patch, so the spec's state shape cannot express a base or patch subject. This predates the range but the range widens the gap. Confirmed by grep.

No other disagreement was found:
- HoldsSubject matches line 3 and the owner's reading for every combination of bare B, B+P, B+inputs and B+P+inputs, and for base-less (old) subjects.
- `--patch` without `--base` exits 2.
- Rules 26 and 27 in the code match the spec, except for item 3.
