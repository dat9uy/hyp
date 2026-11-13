#!/bin/sh
# (b) line tests, F1, F2, G1 and G2 green and (d) go vet, go test on 7a05e9e.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT 7a05e9e >/dev/null 2>&1 || exit 3
cd $WT
echo "== candidate $(git rev-parse HEAD)"
echo "== (b) line tests"
go test ./internal/cli ./internal/store ./internal/replay -count=1 -v -run '^(TestSandboxCutOverBaseTree|TestInitRefusedInsideStore|TestDiscoveryStopsAtRepositoryRoot|TestReplayIgnoresLineOrder|TestAppendSortsAfterNewest|TestUnionMergeKeepsBothBranches|TestClockBehindOnTwoClones|TestMergeSameIDSameActor|TestSameIDLinesAnyOrder|TestAppendInGreatestMillisecond)$' 2>&1 | grep -E '^(--- |ok|FAIL|PASS)'
echo "== (d) go vet ./..."
go vet ./... 2>&1; echo "exit $?"
echo "== (d) go test ./... -count=1"
go test ./... -count=1 2>&1; echo "exit $?"
