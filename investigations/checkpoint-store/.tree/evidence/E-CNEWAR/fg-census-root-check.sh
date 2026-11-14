#!/bin/sh
# (f) Census: hyp status and hyp status G for each Goal, in each store under
# investigations/, with eed0938's binary and c770b50's, built from clean
# detached worktrees to $1/bin-<commit>; same bytes and exit code each.
# (g) Line 9: c770b50's hyp check and hyp status --all at the repository root.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
for c in eed0938 c770b50; do
  git -C $REPO worktree add --detach $OUT/wt-build-$c $c >/dev/null 2>&1 || exit 3
  (cd $OUT/wt-build-$c && CGO_ENABLED=0 go build -o $OUT/bin-$c ./cmd/hyp) || exit 3
  echo "binary $c sha256 $(shasum -a 256 $OUT/bin-$c | cut -d' ' -f1) version $($OUT/bin-$c version)"
  git -C $REPO worktree remove --force $OUT/wt-build-$c
done
same=0; diff=0; allgoals=""
for store in $REPO/investigations/*/; do
  cd $store
  goals=$($OUT/bin-eed0938 events --json | python3 -c 'import json,sys; d=json.load(sys.stdin); print(" ".join("G-"+e["id"][-6:] for e in d["events"] if e["type"]=="node.add" and e["payload"].get("kind")=="Goal"))')
  echo "store $(basename $store): goals $goals"
  allgoals="$allgoals $goals"
  for args in "status" $(for g in $goals; do echo "status:$g"; done); do
    a=$(echo $args | tr ':' ' ')
    $OUT/bin-eed0938 $a > $OUT/o1 2>&1; e1=$?
    $OUT/bin-c770b50 $a > $OUT/o2 2>&1; e2=$?
    if cmp -s $OUT/o1 $OUT/o2 && [ $e1 = $e2 ]; then r=same; same=$((same+1)); else r=DIFFERS; diff=$((diff+1)); fi
    echo "$(basename $store) hyp $a: exit $e1/$e2 sha256 $(shasum -a 256 < $OUT/o1 | cut -c1-12)/$(shasum -a 256 < $OUT/o2 | cut -c1-12) $r"
  done
done
echo "census: $same same, $diff differ"
cd $REPO
echo "== (g) hyp check at $REPO with c770b50's binary"
$OUT/bin-c770b50 check; echo "check exit $?"
echo "== (g) hyp status --all --max-lines 0 at $REPO"
$OUT/bin-c770b50 status --all --max-lines 0 > $OUT/all.txt; echo "status exit $?"
cat $OUT/all.txt
missing=0
for g in $allgoals; do grep -q "^  $g " $OUT/all.txt || { echo "MISSING $g"; missing=$((missing+1)); }; done
echo "goals of the five stores missing from status --all: $missing"
