#!/usr/bin/env bash
# Launch: build every Arm from the sandbox base. The base itself, unpatched,
# is the Arm the inputs file names in base_arm=NAME; each
# $HYP_SANDBOX_PATCHES/<ARM>.patch is one more Arm, applied alone to the
# base. Each Arm is built twice, in two copies of wt at different paths, with
# CGO_ENABLED=0 go build -trimpath, and the two sha256 values must agree.
# The binaries go to $HYP_SANDBOX_RECIPE_TMP/arms/arm-<arm>, named by sha256
# in arms.sha256 there. wt is left as the base checked it out.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}"
die() { echo "launch: $*" >&2; exit 2; }
[ -n "${HYP_SANDBOX_INPUTS:-}" ] || die "no inputs file: pass --inputs FILE holding base_arm=NAME"
base=$(sed -n 's/^base_arm=//p' "$HYP_SANDBOX_INPUTS" | head -n 1)
[ -n "$base" ] || die "inputs file has no base_arm= line"
valid() { printf '%s\n' "$1" | grep -Eqx '[a-z][a-z0-9-]{0,39}'; }
valid "$base" || die "base_arm $base is not an Arm name ([a-z][a-z0-9-], at most 40)"

rt=$HYP_SANDBOX_RECIPE_TMP
mkdir -p "$rt/arms" "$rt/src"
: > "$rt/arms.sha256"

names="$base"
if [ -n "${HYP_SANDBOX_PATCHES:-}" ] && [ -d "$HYP_SANDBOX_PATCHES" ]; then
  for p in "$HYP_SANDBOX_PATCHES"/*.patch; do
    [ -e "$p" ] || continue
    a=$(basename "$p" .patch)
    [ "$a" != "$base" ] || die "Arm $a is both base_arm and a --patch"
    names="$names $a"
  done
fi

one() { # arm, patch file or empty
  local a=$1 patch=$2 n sums=""
  for n in 1 2; do
    local src="$rt/src/$a.$n"
    rm -rf "$src"
    cp -R "$HYP_SANDBOX_WT" "$src"
    if [ -n "$patch" ]; then
      git -C "$src" apply "$patch" || die "patch $a does not apply to the base"
    fi
    (cd "$src" && CGO_ENABLED=0 go build -trimpath -o "$rt/arms/arm-$a.$n" ./cmd/hyp) \
      || die "Arm $a does not build"
    sums="$sums $(shasum -a 256 "$rt/arms/arm-$a.$n" | cut -d' ' -f1)"
    rm -rf "$src"
  done
  set -- $sums
  [ "$1" = "$2" ] || die "Arm $a built twice gave two sha256 values: $1 and $2"
  mv "$rt/arms/arm-$a.1" "$rt/arms/arm-$a"
  rm "$rt/arms/arm-$a.2"
  echo "$1  arm-$a" >> "$rt/arms.sha256"
  echo "launch: arm-$a $1"
}

for a in $(printf '%s\n' $names | LC_ALL=C sort); do
  if [ "$a" = "$base" ]; then
    one "$a" ""
  else
    one "$a" "$HYP_SANDBOX_PATCHES/$a.patch"
  fi
done
rmdir "$rt/src"
echo "launch: $(wc -l < "$rt/arms.sha256" | tr -d ' ') Arms built; base is $base"
