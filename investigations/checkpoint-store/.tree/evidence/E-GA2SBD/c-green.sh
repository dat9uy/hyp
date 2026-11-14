#!/bin/sh
# (c) Green: each line test on 3bbf187, in a clean detached worktree.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT 3bbf187 >/dev/null 2>&1 || exit 3
cd $WT
echo "== candidate $(git rev-parse HEAD)"
for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut; do
  echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -v -run "^$t\$" 2>&1 | grep -E "^(--- |ok|FAIL|PASS)|SKIP"; echo "exit $?"
done
echo "== TestAppendSortsAfterHistory (internal/store)"; go test ./internal/store -count=1 -v -run '^TestAppendSortsAfterHistory$' 2>&1 | grep -E "^(--- |ok|FAIL|PASS)|SKIP"
cd $REPO && git worktree remove --force $WT
