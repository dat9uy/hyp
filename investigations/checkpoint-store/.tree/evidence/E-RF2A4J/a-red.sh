#!/bin/sh
# (a) Red on a7dbc00 (line tests and F1) and on 65a1cc5 (F1, F2), with only
# 10e1925's new or changed test files copied in.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
git -C $REPO worktree add --detach $OUT/wt-red-base a7dbc00 >/dev/null 2>&1 || exit 3
git -C $REPO worktree add --detach $OUT/wt-red-65a $(git -C $REPO rev-parse 65a1cc5) >/dev/null 2>&1 || exit 3
for f in internal/cli/store_layer_test.go internal/store/store_layer_test.go; do
  git -C $REPO show 10e1925:$f > $OUT/wt-red-base/$f
  git -C $REPO show 10e1925:$f > $OUT/wt-red-65a/$f
done
git -C $REPO show 10e1925:internal/replay/replay_test.go > $OUT/wt-red-65a/internal/replay/replay_test.go
cd $OUT/wt-red-base; echo "== a7dbc00 $(git rev-parse HEAD)"
go vet ./internal/cli ./internal/store && echo "vet: builds"
for t in TestSandboxCutOverBaseTree TestInitRefusedInsideStore TestDiscoveryStopsAtRepositoryRoot TestReplayIgnoresLineOrder; do
  echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -run "^$t\$" 2>&1; echo "exit $?"; done
for t in TestAppendSortsAfterNewest TestUnionMergeKeepsBothBranches TestClockBehindOnTwoClones; do
  echo "== $t (internal/store)"; go test ./internal/store -count=1 -run "^$t\$" 2>&1; echo "exit $?"; done
cd $OUT/wt-red-65a; echo "== 65a1cc5 $(git rev-parse HEAD)"
go vet ./internal/store ./internal/replay && echo "vet: builds"
echo "== TestClockBehindOnTwoClones (internal/store)"; go test ./internal/store -count=1 -run '^TestClockBehindOnTwoClones$' 2>&1; echo "exit $?"
echo "== TestMergeSameIDSameActor (internal/replay)"; go test ./internal/replay -count=1 -run '^TestMergeSameIDSameActor$' 2>&1; echo "exit $?"
