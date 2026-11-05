#!/usr/bin/env bash
# Launch: build the arm's hyp, write the arm's briefs, then leave wt holding only
# the fixture store. Reads the inputs file: KEY=VALUE lines arm, trial and model.
# An arm other than base is the trial's one --patch, $HYP_SANDBOX_PATCHES/<arm>.patch;
# base takes no patch. The binary is built twice from copies of wt at different
# paths, with CGO_ENABLED=0 go build -trimpath, and the two sha256 values must agree.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
in="$HYP_SANDBOX_INPUTS"
FILE=skills/hyp-prototype/VERIFIER.md
die() { echo "launch: $*" >&2; exit 2; }
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial model; do [ -n "$(val "$k")" ] || die "inputs file has no $k= line"; done
arm=$(val arm)
trial=$(val trial)
rt=$HYP_SANDBOX_RECIPE_TMP
bundle=$PWD
. "$bundle/fixture/ids.env"

patches=""
if [ -n "${HYP_SANDBOX_PATCHES:-}" ] && [ -d "$HYP_SANDBOX_PATCHES" ]; then
  patches=$(cd "$HYP_SANDBOX_PATCHES" && ls)
fi
if [ "$arm" = base ]; then
  [ -z "$patches" ] || die "arm base takes no patch, but got: $(echo $patches)"
else
  printf '%s\n' "$arm" | grep -Eqx '[a-z][a-z0-9-]{0,39}' || die "arm $arm is not an Arm name"
  [ "$patches" = "$arm.patch" ] || die "arm $arm needs exactly one patch, $arm.patch (--patch $arm=FILE); got: $(echo $patches)"
  git -C "$HYP_SANDBOX_WT" apply "$HYP_SANDBOX_PATCHES/$arm.patch" || die "patch $arm does not apply to wt"
fi

mkdir -p "$rt/bin" "$rt/src"
sums=""
for n in 1 2; do
  cp -R "$HYP_SANDBOX_WT" "$rt/src/$n"
  (cd "$rt/src/$n" && CGO_ENABLED=0 go build -trimpath -o "$rt/bin/hyp.$n" ./cmd/hyp) || die "arm $arm does not build"
  sums="$sums $(shasum -a 256 "$rt/bin/hyp.$n" | cut -d' ' -f1)"
done
rm -rf "$rt/src"
set -- $sums
[ "$1" = "$2" ] || die "arm $arm built twice gave two sha256 values: $1 and $2"
mv "$rt/bin/hyp.1" "$rt/bin/hyp"
rm "$rt/bin/hyp.2"
echo "$1" > "$rt/arm.sha256"

# The named section of the brief file, up to the next heading, with the names a
# parent fills in.
section() {
  awk -v head="## $1" 'found && /^## /{exit} found{print} $0==head{found=1}' "$HYP_SANDBOX_WT/$FILE" \
    | sed -e "s/<Q-id>/$QUESTION/g" -e "s/verifier-<topic>-1/verifier-$arm-$trial/g"
}
section "Brief for the verifier" > "$rt/brief.txt"
rm -f "$rt/reader-brief.txt"
if grep -qx '## Brief for a reader' "$HYP_SANDBOX_WT/$FILE"; then
  section "Brief for a reader" > "$rt/reader-brief.txt"
fi
cp "$in" "$rt/state.env"
cp "$bundle/fixture/items.tsv" "$rt/items.tsv"

# Only the fixture stays. The repository's files name this recipe and git history
# can show them, so .git goes too.
find "$HYP_SANDBOX_WT" -mindepth 1 -maxdepth 1 -exec rm -rf {} +
cp -R "$bundle/fixture/tree" "$HYP_SANDBOX_WT/.tree"
echo "launch: arm $arm, hyp $1; wt holds only the fixture store"
