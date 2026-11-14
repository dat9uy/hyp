#!/bin/sh
# (a) Red on base: eed0938 with only 3bbf187's internal/cli/history_test.go.
# (b) Red on the old candidate: a9963fc with only 3bbf187's history_test.go.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
for c in eed0938 a9963fc; do
  WT=$OUT/wt-red-$c
  git -C $REPO worktree add --detach $WT $c >/dev/null 2>&1 || exit 3
  git -C $REPO show 3bbf187:internal/cli/history_test.go > $WT/internal/cli/history_test.go
  cd $WT
  echo "===== $c with 3bbf187's history_test.go: $(git rev-parse HEAD)"
  go vet ./internal/cli && echo "vet: compiles"
  for t in TestStatusShowsTheFolderStoresGoals TestStatusAllListsTheHistory TestHistoryLeavesOutStores TestCiteAcrossStores TestAppendSortsAfterOtherStores TestDuplicateIDAcrossStores TestAliasLocalToStore TestEventNeverCrossesStores TestSandboxHistoryCut; do
    echo "== $t (internal/cli)"; go test ./internal/cli -count=1 -run "^$t\$" 2>&1 | grep -E "history_test.go|^(--- |ok|FAIL)" | cut -c1-400; echo "exit $?"
  done
  cd $REPO && git worktree remove --force $WT
done
