#!/bin/sh
# (c) Green: each history test on 8a12324, in a clean detached worktree.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT 8a12324 >/dev/null 2>&1 || exit 3
cd $WT
echo "== candidate $(git rev-parse HEAD)"
for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut TestUntrackedStoreAllAndRoot TestSameAskAcrossStores TestStoreByAnotherPath TestAppendSortsAfterUntracked; do
  go test ./internal/cli -count=1 -v -run "^$t\$" > $WT/t.out 2>&1; echo "== $t (internal/cli): go test exit $?"; grep -E "^(--- |ok|FAIL)" $WT/t.out
done
for t in TestAppendSortsAfterHistory TestOneLockRoot TestAppendSortsAfterUnreadRoot; do
  go test ./internal/store -count=1 -v -run "^$t\$" > $WT/t.out 2>&1; echo "== $t (internal/store): go test exit $?"; grep -E "^(--- |ok|FAIL)" $WT/t.out
done
cd $REPO && git worktree remove --force $WT
