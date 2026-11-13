Review of 65a1cc5 against G-1391DM (base a7dbc00). On the candidate, `go vet ./...` and `go test ./... -count=1` pass. All 7 new or changed store-layer tests ran with no skips on this machine. I built the base and candidate binaries in the scratchpad and ran `hyp status` and `hyp check` in all five stores under investigations/. Output and exit codes are byte-identical, so the new merge order changes no existing store.

Findings:

1. **G-1391DM#5, with a knock-on to #4.** internal/store/store.go:182-198 (`idAfter`). When the clock reads earlier than the newest id, the new id is always `last + 1`. That makes it the same on every clone.
   - Scenario: a store's newest event has an id from a fast clock. Branches A and B both start there and each append one event. Both get the identical id `last+1`, which is also the same node id, since `ev.Node = ev.ID`.
   - After the union merge, the store holds two different events with one id. Replay refuses the second as "duplicate id" (internal/replay/validate.go:49), so one branch's node is silently lost.
   - If both appends used the same actor, the two lines sit in one file. Which one survives depends on the line-order tie-break `a.Index - b.Index` (internal/replay/replay.go:47). Then line #4 ("whatever the line order") does not hold.
   - The line's literal case (one clone, clock behind) is met and tested. But the fix breaks the merge that line #6 exists for.
   - A fix would keep the random part: for example, a fresh ULID with the timestamp of `last`, or a random increment above `last`.

2. **G-1391DM#4 (doc).** replay-spec.md:158 says "two lines with the same id in one file keep their file order". ADR 0022:23 says the same. Both claims sit next to "whatever the line order". The two contradict each other, and finding 1 shows the exception can be reached.

3. **G-1391DM#4 (edge case).** An events file with a line that does not parse is not order-independent. `hyp check` prints `error: file:LINE` (internal/check/check.go:36) and `hyp status` prints `warn: skipped file:LINE` (internal/cli/read.go:155). Both use the line number. The same lines in another order therefore print different `hyp check` / `hyp status` output. This is low severity. TestReplayIgnoresLineOrder only uses stores whose lines all parse.

4. **none (AGENTS.md text the body requires).** AGENTS.md:64. The documented procedure fails in exactly the case it is meant for. I reproduced it in a scratch repo: the PR branch has no `.tree/.gitattributes`, and main has added it and appended to the same events file.
   - `git checkout origin/main -- s/.tree/.gitattributes` then `git merge --no-edit origin/main` aborts with "Your local changes to the following files would be overwritten by merge: s/.tree/.gitattributes" (exit 2).
   - Committing the checked-out file first, then merging, works: exit 0, union applied, lines `l0, pr1, main1`.
   - The text needs a commit step between the two commands, or another method such as `.git/info/attributes`.

5. **G-1391DM#1 (test coverage, minor).** internal/cli/store_layer_test.go:83 skips when `lsof` is not on PATH. On such a machine nothing tests line #1. It ran here, on macOS.

No defects found in these areas:
- Store discovery (`.git` dir or file; a store at the repository root is still found because it is checked before `.git`).
- Init refusal: exit 1, names the store, creates nothing; also covered when the store is at cwd.
- `.gitattributes` content and the union-merge test.
- Removing the base's `.tree` before `cut.Write` in sandbox create.
- The cut writing each file in id order.

Spec and doc text: replay-spec §1, §2 and §6, sandbox-spec §5.3, the glossary, install.md, RECIPE-FORMAT.md, the hyp-prototype step 2, README.md and ADR 0022 plus the amendment to 0020 all match the code. The exceptions are findings 2 and 4. No remaining text mentions the old "base already holds .tree" refusal or the k-way file-order merge. The hyp-owner-grill placement text already agrees. No skill cites an ADR.

Status: DONE_WITH_CONCERNS
Summary: All six lines are met as written and tested, and existing stores replay unchanged. But the clock-behind id rule (`last+1`) gives the same ids on two branches, so a union merge loses one branch's event through a duplicate-id refusal. The AGENTS.md conflict-merge recipe fails as written (exit 2) unless the checked-out `.gitattributes` is committed before the merge.
