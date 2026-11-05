#!/usr/bin/env bash
# Launch: leave wt holding only the tree at the cut, and write the brief a real
# parent would give. Reads the inputs file: KEY=VALUE lines arm, trial and model.
# An arm other than base is the trial's one --patch,
# $HYP_SANDBOX_PATCHES/<arm>.patch, touching only skills/hyp-prototype/VERIFIER.md;
# base takes no patch.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
in="$HYP_SANDBOX_INPUTS"
FILE=skills/hyp-prototype/VERIFIER.md
QID=Q-KW0PWZ
die() { echo "launch: $*" >&2; exit 2; }
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial model; do [ -n "$(val "$k")" ] || die "inputs file has no $k= line"; done
arm=$(val arm)
trial=$(val trial)
rt=$HYP_SANDBOX_RECIPE_TMP

patches=""
if [ -n "${HYP_SANDBOX_PATCHES:-}" ] && [ -d "$HYP_SANDBOX_PATCHES" ]; then
  patches=$(cd "$HYP_SANDBOX_PATCHES" && ls)
fi
variant=""
if [ "$arm" = base ]; then
  [ -z "$patches" ] || die "arm base takes no patch, but got: $(echo $patches)"
else
  printf '%s\n' "$arm" | grep -Eqx '[a-z][a-z0-9-]{0,39}' || die "arm $arm is not an Arm name"
  [ "$patches" = "$arm.patch" ] || die "arm $arm needs exactly one patch, $arm.patch (--patch $arm=FILE); got: $(echo $patches)"
  variant="$HYP_SANDBOX_PATCHES/$arm.patch"
  touched=$(git -C "$HYP_SANDBOX_WT" apply --numstat "$variant" | cut -f3-) || die "patch $arm does not parse"
  [ "$touched" = "$FILE" ] || die "patch $arm must touch only $FILE, not: $(echo $touched)"
fi

cd "$HYP_SANDBOX_WT"
[ -d .tree/events ] || die "wt has no .tree: create the sandbox with --store and --cut"

# The section after the brief's heading, with the two names a parent fills in.
section() {
  sed -n '/^## Brief for the verifier$/,$p' | tail -n +2 \
    | sed -e "s/<Q-id>/$QID/g" -e "s/verifier-<topic>-1/verifier-$arm-$trial/g"
}
git show "HEAD:$FILE" | section > "$rt/base-brief.txt"
if [ -n "$variant" ]; then
  git apply "$variant" || die "patch $arm does not apply to wt"
fi
section < "$FILE" > "$rt/brief.txt"

cp "$in" "$rt/state.env"
# Only .tree stays. A file in the repository names the wrong Result and git
# history can show that file, so .git goes too. This changes the clone, not the cut.
find . -mindepth 1 -maxdepth 1 ! -name .tree -exec rm -rf {} +
echo "launch: arm $arm; wt holds only .tree at the cut; the brief is in this recipe's tmp"
