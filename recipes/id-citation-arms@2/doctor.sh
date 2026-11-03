#!/usr/bin/env bash
# Doctor: python3 and go run, every Arm binary matches arms.sha256, and the
# Arms are exactly the base Arm plus one per patch.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
rt=${HYP_SANDBOX_RECIPE_TMP:?}
command -v python3 >/dev/null || fail "python3 is not on PATH"
go version >/dev/null 2>&1 || fail "go does not run"
[ -s "$rt/arms.sha256" ] || fail "arms.sha256 is missing or empty"
(cd "$rt/arms" && shasum -a 256 -c ../arms.sha256 >/dev/null) || fail "an Arm binary does not match arms.sha256"
base=$(sed -n 's/^base_arm=//p' "$HYP_SANDBOX_INPUTS" | head -n 1)
want=$(
  echo "$base"
  if [ -n "${HYP_SANDBOX_PATCHES:-}" ] && [ -d "$HYP_SANDBOX_PATCHES" ]; then
    for p in "$HYP_SANDBOX_PATCHES"/*.patch; do [ -e "$p" ] && basename "$p" .patch; done
  fi
)
want=$(printf '%s\n' $want | LC_ALL=C sort | tr '\n' ' ')
got=$(sed 's/^[0-9a-f]*  arm-//' "$rt/arms.sha256" | LC_ALL=C sort | tr '\n' ' ')
[ "$got" = "$want" ] || fail "Arms built: $got; want the base and one per patch: $want"
[ "$(ls "$rt/arms" | sed 's/^arm-//' | LC_ALL=C sort | tr '\n' ' ')" = "$want" ] || fail "arms/ holds a file that is no Arm"
echo "doctor: ok; Arms $got"
