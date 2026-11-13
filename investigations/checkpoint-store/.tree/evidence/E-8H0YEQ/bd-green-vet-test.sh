#!/bin/sh
# (b) line tests, F1, G1, G2 and H1 green and (d) go vet, go test on 9643341.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT 9643341 >/dev/null 2>&1 || exit 3
cd $WT
echo "== candidate $(git rev-parse HEAD)"
echo "== (b) line tests"
go test ./internal/cli ./internal/store -count=1 -v -run '^(TestSandboxCutOverBaseTree|TestInitRefusedInsideStore|TestDiscoveryStopsAtRepositoryRoot|TestReplayIgnoresLineOrder|TestAppendSortsAfterNewest|TestUnionMergeKeepsBothBranches|TestClockBehindOnTwoClones|TestSameIDLinesAnyOrder|TestAppendInGreatestMillisecond|TestAppendAfterRejectedID)$' 2>&1 | grep -E '^(--- |ok|FAIL|PASS)'
echo "== (d) go vet ./..."
go vet ./... 2>&1; echo "exit $?"
echo "== (d) go test ./... -count=1"
go test ./... -count=1 2>&1; echo "exit $?"
