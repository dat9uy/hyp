#!/bin/sh
# (a) eed0938 with c770b50's internal/cli/history_test.go: the eleven cli tests.
# (b) 3bbf187 with c770b50's cli and store history_test.go files.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
for c in eed0938 3bbf187; do
  WT=$OUT/wt-red-$c
  git -C $REPO worktree add --detach $WT $c >/dev/null 2>&1 || exit 3
  git -C $REPO show c770b50:internal/cli/history_test.go > $WT/internal/cli/history_test.go
  [ $c = 3bbf187 ] && git -C $REPO show c770b50:internal/store/history_test.go > $WT/internal/store/history_test.go
  cd $WT
  echo "===== $c with c770b50's history tests: $(git rev-parse HEAD)"
  go vet ./internal/cli && echo "vet: compiles"
  for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut TestUntrackedStoreAllAndRoot TestSameAskAcrossStores; do
    go test ./internal/cli -count=1 -run "^$t\$" > $OUT/t.out 2>&1; e=$?
    echo "== $t (internal/cli): go test exit $e"; grep -E "history_test.go|^(--- |ok|FAIL)" $OUT/t.out | cut -c1-400
  done
  if [ $c = 3bbf187 ]; then
    go test ./internal/store -count=1 -run '^TestOneLockRoot$' > $OUT/t.out 2>&1; e=$?
    echo "== TestOneLockRoot (internal/store): go test exit $e"; grep -E "history_test.go|^(--- |ok|FAIL)" $OUT/t.out | cut -c1-400
  fi
  cd $REPO && git worktree remove --force $WT
done
