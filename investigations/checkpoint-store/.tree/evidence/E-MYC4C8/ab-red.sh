#!/bin/sh
# (a) eed0938 with 8a12324's internal/cli/history_test.go: the thirteen cli tests.
# (b) 3a886de with 8a12324's cli and store history_test.go files.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
for c in eed0938 3a886de; do
  WT=$OUT/wt-red-$c
  git -C $REPO worktree add --detach $WT $c >/dev/null 2>&1 || exit 3
  git -C $REPO show 8a12324:internal/cli/history_test.go > $WT/internal/cli/history_test.go
  [ $c = 3a886de ] && git -C $REPO show 8a12324:internal/store/history_test.go > $WT/internal/store/history_test.go
  cd $WT
  echo "===== $c with 8a12324's history tests: $(git rev-parse HEAD)"
  go vet ./internal/cli && echo "vet: compiles"
  for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut TestUntrackedStoreAllAndRoot TestSameAskAcrossStores TestStoreByAnotherPath TestAppendSortsAfterUntracked; do
    go test ./internal/cli -count=1 -run "^$t\$" > $OUT/t.out 2>&1; e=$?
    echo "== $t (internal/cli): go test exit $e"; grep -E "history_test.go|^(--- |ok|FAIL)" $OUT/t.out | cut -c1-400
  done
  if [ $c = 3a886de ]; then
    go test ./internal/store -count=1 -run '^TestAppendSortsAfterUnreadRoot$' > $OUT/t.out 2>&1; e=$?
    echo "== TestAppendSortsAfterUnreadRoot (internal/store): go test exit $e"; grep -E "history_test.go|^(--- |ok|FAIL)" $OUT/t.out | cut -c1-400
  fi
  cd $REPO && git worktree remove --force $WT
done
