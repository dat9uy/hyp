#!/bin/bash
# usage: vet-test-census.sh REPO COMMIT SHORT OUTDIR
# (b) go vet and go test in a clean detached worktree at COMMIT;
# (c) compile COMMIT to the scratchpad and compare hyp status and hyp check
# with the installed hyp and hyp-33896ec.
set -u
REPO=$1 COMMIT=$2 SHORT=$3 OUT=$4
S=$(cd "$(dirname "$0")" && pwd)
WT=$S/wt-clean
BIN=$S/hyp-$SHORT
INST=$HOME/.local/bin/hyp
OLD=$HOME/.local/share/hyp/hyp-33896ec
git -C "$REPO" worktree remove --force "$WT" 2>/dev/null
git -C "$REPO" worktree add --detach "$WT" "$COMMIT" >/dev/null 2>&1 || exit 3
mkdir -p "$OUT/c-census"
B=$OUT/b-vet-test-$SHORT.log
{
  echo "== worktree $(git -C "$WT" rev-parse HEAD), status: $(git -C "$WT" status --porcelain | wc -l | tr -d ' ') changed files"
  (cd "$WT" && go vet ./...); echo "== go vet exit $?"
  (cd "$WT" && go test ./... -count=1); echo "== go test exit $?"
} >"$B" 2>&1
(cd "$WT" && CGO_ENABLED=0 go build -o "$BIN" ./cmd/hyp) || exit 3
C=$OUT/c-census.log
{
  echo "== binaries"
  shasum -a 256 "$BIN" "$INST" "$OLD"
  c=$(shasum -a 256 "$BIN" | cut -d' ' -f1); i=$(shasum -a 256 "$INST" | cut -d' ' -f1); o=$(shasum -a 256 "$OLD" | cut -d' ' -f1)
  [ "$c" != "$i" ] && [ "$c" != "$o" ] && echo "candidate $SHORT sha256 differs from installed and 33896ec: yes" || echo "candidate $SHORT sha256 differs from installed and 33896ec: NO"
  for tree in id-citation recipe-base checkpoint-store-design checkpoint-store; do
    for cmd in status check; do
      d=$REPO/investigations/$tree
      (cd "$d" && "$BIN" $cmd) >"$OUT/c-census/$tree.$cmd.$SHORT" 2>&1; ec=$?
      (cd "$d" && "$INST" $cmd) >"$OUT/c-census/$tree.$cmd.installed" 2>&1; ei=$?
      cmp -s "$OUT/c-census/$tree.$cmd.$SHORT" "$OUT/c-census/$tree.$cmd.installed" && [ $ec = $ei ] && r=identical || r=DIFFERENT
      echo "$tree hyp $cmd: $SHORT vs installed: $r (exit $ec/$ei, $(wc -c <"$OUT/c-census/$tree.$cmd.$SHORT" | tr -d ' ') bytes)"
      if [ $tree != checkpoint-store ]; then
        (cd "$d" && "$OLD" $cmd) >"$OUT/c-census/$tree.$cmd.33896ec" 2>&1; eo=$?
        cmp -s "$OUT/c-census/$tree.$cmd.$SHORT" "$OUT/c-census/$tree.$cmd.33896ec" && [ $ec = $eo ] && r=identical || r=DIFFERENT
        echo "$tree hyp $cmd: $SHORT vs 33896ec: $r (exit $ec/$eo)"
      fi
    done
  done
} >"$C" 2>&1
git -C "$REPO" worktree remove --force "$WT"
tail -3 "$B"; cat "$C"
