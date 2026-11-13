#!/bin/sh
# (a) Red on a7dbc00 (line tests, F1, G1) and on 7a05e9e (G2), with only
# ded3e4b's two store_layer_test.go files copied in.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
git -C $REPO worktree add --detach $OUT/wt-red-base a7dbc00 >/dev/null 2>&1 || exit 3
git -C $REPO worktree add --detach $OUT/wt-red-7a0 $(git -C $REPO rev-parse 7a05e9e) >/dev/null 2>&1 || exit 3
for f in internal/cli/store_layer_test.go internal/store/store_layer_test.go; do
  git -C $REPO show ded3e4b:$f > $OUT/wt-red-base/$f
  git -C $REPO show ded3e4b:$f > $OUT/wt-red-7a0/$f
done
cd $OUT/wt-red-base; echo "== a7dbc00 $(git rev-parse HEAD)"
go vet ./internal/cli ./internal/store && echo "vet: builds"
for t in TestSandboxCutOverBaseTree TestInitRefusedInsideStore TestDiscoveryStopsAtRepositoryRoot TestReplayIgnoresLineOrder; do
  echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -run "^$t\$" 2>&1; echo "exit $?"; done
for t in TestAppendSortsAfterNewest TestUnionMergeKeepsBothBranches TestClockBehindOnTwoClones TestSameIDLinesAnyOrder; do
  echo "== $t (internal/store)"; go test ./internal/store -count=1 -run "^$t\$" 2>&1; echo "exit $?"; done
cd $OUT/wt-red-7a0; echo "== 7a05e9e $(git rev-parse HEAD)"
go vet ./internal/store && echo "vet: builds"
echo "== TestAppendInGreatestMillisecond (internal/store)"; go test ./internal/store -count=1 -run '^TestAppendInGreatestMillisecond$' 2>&1; echo "exit $?"
