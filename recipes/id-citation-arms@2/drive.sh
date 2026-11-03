#!/usr/bin/env bash
# Drive: run ./cases.py, the script the Standalone label `cases` runs and
# the known-answer gate checks, over the bundle's fixture/cases.tsv and
# fixture/fixture.tar with every Arm in arms.sha256, in that file's order.
# The table goes to table.tsv and each step's stderr to stderr.tar, both
# under this recipe's tmp, for remove's --keep.
set -eu
rt=${HYP_SANDBOX_RECIPE_TMP:?}
bundle=$PWD
[ -s "$rt/arms.sha256" ] || { echo "drive: arms.sha256 is missing: Launch did not run" >&2; exit 1; }
(cd "$rt/arms" && shasum -a 256 -c ../arms.sha256 >/dev/null) || { echo "drive: an Arm binary does not match arms.sha256" >&2; exit 1; }
set --
while read -r _ f; do
  set -- "$@" "${f#arm-}=$rt/arms/$f"
done < "$rt/arms.sha256"
rm -rf "$rt/stderr"
mkdir "$rt/stderr"
HYP_OUT="$rt/stderr" "$bundle/cases.py" "$bundle/fixture/cases.tsv" "$bundle/fixture/fixture.tar" "$@" > "$rt/table.tsv"
tar -cf "$rt/stderr.tar" -C "$rt" stderr
echo "drive: $(($(wc -l < "$rt/table.tsv") - 1)) rows in table.tsv"
