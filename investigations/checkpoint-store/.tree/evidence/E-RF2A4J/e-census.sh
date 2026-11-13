#!/bin/sh
# (e) Census: status, status per Goal and check in each store under
# investigations/, with a7dbc00's binary and 10e1925's, built from clean
# detached worktrees to $1/bin-<commit>. Prints each binary's sha256 and,
# per store and command, whether output bytes and exit code are the same.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
OUT=$1
for c in a7dbc00 10e1925; do
  git -C $REPO worktree add --detach $OUT/wt-build-$c $c >/dev/null 2>&1 || exit 3
  (cd $OUT/wt-build-$c && CGO_ENABLED=0 go build -o $OUT/bin-$c ./cmd/hyp) || exit 3
  echo "binary $c sha256 $(shasum -a 256 $OUT/bin-$c | cut -d' ' -f1) version $($OUT/bin-$c version)"
done
same=0; diff=0
for store in $REPO/investigations/*/; do
  cd $store
  goals=$($OUT/bin-10e1925 events --json | python3 -c 'import json,sys; d=json.load(sys.stdin); evs=d["events"]; print(" ".join(sorted({"G-"+e["id"][-6:] for e in evs if e["type"]=="node.add" and e["payload"].get("kind")=="Goal"})))')
  for args in "status" "check" $(for g in $goals; do echo "status:$g"; done); do
    a=$(echo $args | tr ':' ' ')
    $OUT/bin-a7dbc00 $a > $OUT/o1 2>&1; e1=$?
    $OUT/bin-10e1925 $a > $OUT/o2 2>&1; e2=$?
    if cmp -s $OUT/o1 $OUT/o2 && [ $e1 = $e2 ]; then r=same; same=$((same+1)); else r=DIFFERS; diff=$((diff+1)); fi
    echo "$(basename $store) hyp $a: exit $e1/$e2 sha256 $(shasum -a 256 < $OUT/o1 | cut -c1-12)/$(shasum -a 256 < $OUT/o2 | cut -c1-12) $r"
  done
done
echo "census: $same same, $diff differ"
