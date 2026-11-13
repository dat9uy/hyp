#!/bin/sh
# (a) Red on base: a7dbc00 with only the two new test files copied in.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT a7dbc00 >/dev/null 2>&1 || exit 3
git -C $REPO show 65a1cc5:internal/cli/store_layer_test.go > $WT/internal/cli/store_layer_test.go
git -C $REPO show 65a1cc5:internal/store/store_layer_test.go > $WT/internal/store/store_layer_test.go
cd $WT
echo "== base $(git rev-parse HEAD)"
go vet ./internal/cli ./internal/store && echo "vet: builds"
for t in TestSandboxCutOverBaseTree TestInitRefusedInsideStore TestDiscoveryStopsAtRepositoryRoot TestReplayIgnoresLineOrder; do
  echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -run "^$t\$" 2>&1; echo "exit $?"
done
for t in TestAppendSortsAfterNewest TestUnionMergeKeepsBothBranches; do
  echo "== $t (internal/store)"; go test ./internal/store -count=1 -run "^$t\$" 2>&1; echo "exit $?"
done
