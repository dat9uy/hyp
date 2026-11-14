#!/bin/bash
# Gate of E-KCSN9Y in the detached checkout CAND of f296bc11b479; logs go to OUT.
set -u
CAND=$1; OUT=$2
cd "$CAND" || exit 3
echo "$(git rev-parse HEAD)" > "$OUT/candidate-commit.txt"
git status --short > "$OUT/candidate-clean.txt"
NEW='^(TestBundleWithStoreInputKeepsItsHash|TestSyncKeepsNamedFileShapedLikeAnEventsFile|TestSyncKeepsEachSidesNamedBytes|TestRecordRefusesEvidenceChangedDuringTheWrite|TestBoundaryFindsEveryWayToRunGit|TestOnlyRecordNamesTheFoldersAndRunsGit|TestRecordOutlivesEveryWorktreeAct|TestSubjectCommitStaysReachable|TestTwoWorktreesWriteAtOnce|TestSyncJoinsTwoRecords|TestRecordHoldsOnlyNamedFiles|TestFlagsSameFromEveryCheckout|TestImportAddsEachMissingEvent|TestUnimportedWorkTreeStore)$'
{ go test ./internal/cli -run "$NEW" -count=1 -v 2>&1; echo "go test cli exit $?";
  go test ./internal/store -run '^TestAppendRefusesANamedFileNoOneHolds$' -count=1 -v 2>&1; echo "go test store exit $?";
  go test ./internal/record -run '^TestStoppedWriteLeavesAllOrNothing$' -count=1 -v 2>&1; echo "go test record exit $?"; } > "$OUT/candidate-new-tests.txt"
cp "$OUT/zz_boundary_probe.go" internal/render/zz_boundary_probe.go
{ echo "== TestOnlyRecordNamesTheFoldersAndRunsGit with the probe file internal/render/zz_boundary_probe.go; the line asks it to fail";
  go test ./internal/cli -run '^TestOnlyRecordNamesTheFoldersAndRunsGit$' -count=1 -v 2>&1; echo "go test exit $?"; } > "$OUT/candidate-boundary-probe.txt"
rm internal/render/zz_boundary_probe.go
run_removal() { # name test pkg
	local n=$1 t=$2 pkg=$3
	{ echo "===== removal $n (patch removal-$n.patch), named test $t"
	  if git apply "$OUT/removal-$n.patch"; then echo "patch applied whole"; else echo "patch did NOT apply whole: failed removal"; fi
	  go test "$pkg" -run "^$t\$" -count=1 2>&1; echo "go test exit $?"
	  git checkout -q -- . ; } >> "$OUT/removal-checks.txt"
}
: > "$OUT/removal-checks.txt"
run_removal r1-refuse-unimported TestUnimportedWorkTreeStore ./internal/cli
run_removal r2-record-copy-differs TestRecordOutlivesEveryWorktreeAct ./internal/cli
run_removal r3-named-file-no-one-holds TestAppendRefusesANamedFileNoOneHolds ./internal/store
run_removal r4-import-in-sandbox TestSubjectCommitStaysReachable ./internal/cli
run_removal r5-sync-in-sandbox TestSubjectCommitStaysReachable ./internal/cli
run_removal r6-changed-while-written TestRecordRefusesEvidenceChangedDuringTheWrite ./internal/cli
run_removal r7-show-several-contents TestSyncKeepsEachSidesNamedBytes ./internal/cli
run_removal r8-boundary-exec-import TestBoundaryFindsEveryWayToRunGit ./internal/cli
run_removal r9b-boundary-start-process TestBoundaryFindsEveryWayToRunGit ./internal/cli
run_removal r10-sync-keeps-other-copy TestSyncKeepsEachSidesNamedBytes ./internal/cli
run_removal r11-bundle-file-belongs-to-bundle TestBundleWithStoreInputKeepsItsHash ./internal/cli
run_removal r12-only-a-stores-events-file TestSyncKeepsNamedFileShapedLikeAnEventsFile ./internal/cli
git status --short >> "$OUT/candidate-clean.txt"
{ go vet ./... 2>&1; echo "go vet exit $?"; } > "$OUT/candidate-vet.txt"
{ go test ./... -count=1 2>&1; echo "go test exit $?"; } > "$OUT/candidate-go-test.txt"
