#!/bin/sh
# (a) Red on base: eed0938 with only internal/cli/history_test.go copied in.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT eed0938 >/dev/null 2>&1 || exit 3
git -C $REPO show a9963fc:internal/cli/history_test.go > $WT/internal/cli/history_test.go
cd $WT
echo "== base $(git rev-parse HEAD)"
go vet ./internal/cli && echo "vet: builds"
for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut; do
  echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -run "^$t\$" 2>&1; echo "exit $?"
done
cd $REPO && git worktree remove --force $WT
