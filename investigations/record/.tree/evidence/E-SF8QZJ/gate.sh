#!/bin/bash
# Gate of E-SF8QZJ in the detached checkout CAND of 32d707c23b7c; logs go to OUT.
set -u
CAND=$1; OUT=$2
cd "$CAND" || exit 3
echo "$(git rev-parse HEAD)" > "$OUT/candidate-commit.txt"
git status --short > "$OUT/candidate-clean.txt"
NEW='^(TestRecordRefDeletedWithNoGitUsesTheWorkTree|TestSyncedReftableRecordNeedsGit|TestReftableRepositoryWithNoGit|TestRecordNeedsGitToReadTheRecord|TestNoGitInARepositoryWithNoRecordUsesTheWorkTree|TestImportKeepsCitedFileTrackedBesideAGlobName|TestShowPickPrintsOnlyBytesOfThatSha|TestShowBundleEvidenceAfterSync|TestPinnedBundleAtTwoLibraryLevels|TestSyncKeepsEachSidesPinnedBundle|TestRecordWriteHardensObjectsAndRef|TestBundleWithStoreInputKeepsItsHash|TestSyncKeepsNamedFileShapedLikeAnEventsFile|TestSyncKeepsEachSidesNamedBytes|TestRecordRefusesEvidenceChangedDuringTheWrite|TestBoundaryFindsEveryWayToRunGit|TestOnlyRecordNamesTheFoldersAndRunsGit|TestRecordOutlivesEveryWorktreeAct|TestSubjectCommitStaysReachable|TestTwoWorktreesWriteAtOnce|TestSyncJoinsTwoRecords|TestRecordHoldsOnlyNamedFiles|TestFlagsSameFromEveryCheckout|TestImportAddsEachMissingEvent|TestUnimportedWorkTreeStore)$'
{ go test ./internal/cli -run "$NEW" -count=1 -v 2>&1; echo "go test cli exit $?";
  go test ./internal/store -run '^TestAppendRefusesANamedFileNoOneHolds$' -count=1 -v 2>&1; echo "go test store exit $?";
  go test ./internal/record -run '^(TestStoppedWriteLeavesAllOrNothing|TestReftableHoldsReadsAsGitDoes|TestReftableHoldsRefusesAMalformedTable|TestRefOnDiskCountsAnUnreadableReftable|TestReftableHoldsRefusesLengthsPastTheBlock|TestReftableHoldsNeverPanicsOnACorruptTable)$' -count=1 -v 2>&1; echo "go test record exit $?"; } > "$OUT/candidate-new-tests.txt"
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
run_removal r2b-record-copy-differs TestRecordOutlivesEveryWorktreeAct ./internal/cli
run_removal r3-named-file-no-one-holds TestAppendRefusesANamedFileNoOneHolds ./internal/store
run_removal r4-import-in-sandbox TestSubjectCommitStaysReachable ./internal/cli
run_removal r5-sync-in-sandbox TestSubjectCommitStaysReachable ./internal/cli
run_removal r6-changed-while-written TestRecordRefusesEvidenceChangedDuringTheWrite ./internal/cli
run_removal r7-show-several-contents TestSyncKeepsEachSidesNamedBytes ./internal/cli
run_removal r8-boundary-exec-import TestBoundaryFindsEveryWayToRunGit ./internal/cli
run_removal r9b-boundary-start-process TestBoundaryFindsEveryWayToRunGit ./internal/cli
run_removal r10-sync-keeps-other-copy TestSyncKeepsEachSidesNamedBytes ./internal/cli
run_removal r11b-bundle-file-belongs-to-bundle TestBundleWithStoreInputKeepsItsHash ./internal/cli
run_removal r12-only-a-stores-events-file TestSyncKeepsNamedFileShapedLikeAnEventsFile ./internal/cli
run_removal r13-sync-keeps-bundle-whole TestSyncKeepsEachSidesPinnedBundle ./internal/cli
run_removal r14-pin-reads-kept-copy TestSyncKeepsEachSidesPinnedBundle ./internal/cli
run_removal r15-write-keeps-other-copy TestSyncKeepsEachSidesPinnedBundle ./internal/cli
run_removal r16b-git-fsync TestRecordWriteHardensObjectsAndRef ./internal/cli
run_removal r17-snapshot-holds-every-level TestPinnedBundleAtTwoLibraryLevels ./internal/cli
run_removal r18-show-reads-each-pinned-copy TestShowBundleEvidenceAfterSync ./internal/cli
run_removal r19b-unreadable-record-exits-3 TestRecordNeedsGitToReadTheRecord ./internal/cli
run_removal r20-literal-pathspecs TestImportKeepsCitedFileTrackedBesideAGlobName ./internal/cli
run_removal r21-pick-prints-only-its-sha TestShowPickPrintsOnlyBytesOfThatSha ./internal/cli
run_removal r22b-no-ref-on-disk-reads-work-tree TestNoGitInARepositoryWithNoRecordUsesTheWorkTree ./internal/cli
run_removal r24-read-the-reftable TestSyncedReftableRecordNeedsGit ./internal/cli
run_removal r25-reftable-deletion TestReftableHoldsReadsAsGitDoes ./internal/record
run_removal r26-unreadable-reftable-holds TestRefOnDiskCountsAnUnreadableReftable ./internal/record
run_removal r27-reftable-lengths-fit-the-block TestReftableHoldsRefusesLengthsPastTheBlock ./internal/record
cp "$OUT/strict-review_bundle_evidence_show_test.go" internal/cli/zz_strict_test.go
{ echo "== strict check TestReviewShowBundleEvidenceAfterSync of E-AK3VNM (the other meaning of R-CYBPRP#1's case), run at the candidate; it sets no verdict";
  go test ./internal/cli -run '^TestReviewShowBundleEvidenceAfterSync$' -count=1 -v 2>&1; echo "go test exit $?"; } > "$OUT/strict-check.txt"
rm internal/cli/zz_strict_test.go
cp "$OUT/oog-review_findings_test.go" internal/cli/zz_oog_review_test.go
{ echo "== out-of-Goal by answer R-EK3R0H#r1: the review tests of E-FWBCRZ for a path named with other bytes, run at the candidate; they set no verdict";
  go test ./internal/cli -run '^(TestReviewImportFromRefThenCheckoutAddsTheLaterEvent|TestReviewShowNeverPrintsOtherBytesForAPickedSha)$' -count=1 -v 2>&1; echo "go test exit $?"; } > "$OUT/oog-review-tests.txt"
rm internal/cli/zz_oog_review_test.go
git status --short >> "$OUT/candidate-clean.txt"
{ echo "== TestAddGoalRefusesATruncatedHash, 20 runs"; go test ./internal/cli -run '^TestAddGoalRefusesATruncatedHash$' -count=20 -v 2>&1 | grep -E '^(--- |ok|FAIL)'; echo "go test exit $?"; } > "$OUT/candidate-truncated-hash-20.txt"
{ go vet ./... 2>&1; echo "go vet exit $?"; } > "$OUT/candidate-vet.txt"
{ go test ./... -count=1 2>&1; echo "go test exit $?"; } > "$OUT/candidate-go-test.txt"
