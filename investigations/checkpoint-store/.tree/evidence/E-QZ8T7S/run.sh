#!/bin/sh
# E-QZ8T7S (a) to (d). Usage: run.sh <scratch dir>
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
CAND=d966e85c2dd802ee9823689498300e241dd87601
T=TestAppendAfterRejectedIDKeepsClock
git -C $REPO worktree add --detach $OUT/wt-ded $(git -C $REPO rev-parse ded3e4b) >/dev/null 2>&1 || exit 3
git -C $REPO worktree add --detach $OUT/wt-cand $CAND >/dev/null 2>&1 || exit 3
git -C $REPO worktree add --detach $OUT/wt-rm $CAND >/dev/null 2>&1 || exit 3
echo "== (a) red: ded3e4b $(git -C $OUT/wt-ded rev-parse HEAD) with $CAND's internal/store/store_layer_test.go"
git -C $REPO show $CAND:internal/store/store_layer_test.go > $OUT/wt-ded/internal/store/store_layer_test.go
(cd $OUT/wt-ded && go vet ./internal/store && echo "vet: builds"; go test ./internal/store -count=1 -run "^$T\$" 2>&1; echo "exit $?")
echo "== (b) green: $(git -C $OUT/wt-cand rev-parse HEAD)"
(cd $OUT/wt-cand && go test ./internal/store -count=1 -v -run "^$T\$" 2>&1; echo "exit $?")
echo "== (c) removal: ValidID loop replaced by last = evs[len(evs)-1].ID"
python3 - $OUT/wt-rm/internal/store/store.go <<'PY'
import sys
p=sys.argv[1]; s=open(p).read()
old="""	for i := len(evs) - 1; i >= 0; i-- {
		if replay.ValidID(evs[i].ID) {
			last = evs[i].ID
			break
		}
	}
"""
new="""	if len(evs) > 0 {
		last = evs[len(evs)-1].ID
	}
"""
assert old in s; open(p,'w').write(s.replace(old,new,1))
PY
(cd $OUT/wt-rm && git diff --stat && go test ./internal/store -count=1 -run "^$T\$" 2>&1; echo "exit $?")
echo "== (d) vet and test on $CAND"
(cd $OUT/wt-cand && go vet ./... && echo "vet: ok"; go test ./... -count=1 2>&1; echo "exit $?")
echo "== (d) line tests, verbose"
(cd $OUT/wt-cand && go test ./internal/cli -count=1 -v -run '^(TestSandboxCutOverBaseTree|TestInitRefusedInsideStore|TestDiscoveryStopsAtRepositoryRoot|TestReplayIgnoresLineOrder)$' 2>&1; echo "exit $?"
 go test ./internal/store -count=1 -v -run '^(TestAppendSortsAfterNewest|TestClockBehindOnTwoClones|TestAppendAfterRejectedID|TestAppendAfterRejectedIDKeepsClock|TestUnionMergeKeepsBothBranches|TestSameIDLinesAnyOrder)$' 2>&1; echo "exit $?")
for w in wt-ded wt-cand wt-rm; do git -C $REPO worktree remove --force $OUT/$w; done
