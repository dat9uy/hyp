#!/bin/sh
# Import check of E-FWBCRZ: in the sandbox clone WT, the candidate BIN prints
# hyp render --markdown for each store, imports all five
# stores, and prints them again. OUT receives every file.
set -u
BIN=$1
WT=$2
OUT=$3
STORES="checkpoint-store-design checkpoint-store goal-prompt id-citation recipe-base"
mkdir -p "$OUT"
echo "binary: $("$BIN" version)"
echo "record ref before: $(git -C "$WT" for-each-ref --format='%(objectname)' refs/hyp/record)"
for s in $STORES; do
	(cd "$WT/investigations/$s" && "$BIN" render --markdown > "$OUT/before-$s.md" 2>&1; echo "render $s before: exit $?")
done
cd "$WT" || exit 3
start=$(date +%s)
"$BIN" import $(for s in $STORES; do printf 'investigations/%s ' "$s"; done) > "$OUT/import-stdout.txt" 2> "$OUT/import-stderr.txt"
echo "import exit $?, $(( $(date +%s) - start ))s"
cat "$OUT/import-stdout.txt"
echo "import warnings: $(wc -l < "$OUT/import-stderr.txt") lines"
echo "record ref after: $(git -C "$WT" for-each-ref --format='%(objectname)' refs/hyp/record)"
"$BIN" import $(for s in $STORES; do printf 'investigations/%s ' "$s"; done) > "$OUT/import-again-stdout.txt" 2>&1
echo "second import exit $?"
cat "$OUT/import-again-stdout.txt"
echo "record ref after the second import: $(git -C "$WT" for-each-ref --format='%(objectname)' refs/hyp/record)"
fail=0
for s in $STORES; do
	(cd "$WT/investigations/$s" && "$BIN" render --markdown > "$OUT/after-$s.md" 2>&1; echo "render $s after: exit $?")
	if diff "$OUT/before-$s.md" "$OUT/after-$s.md" > "$OUT/diff-$s.txt"; then
		echo "$s: render equal"
	else
		echo "$s: render DIFFERS ($(wc -l < "$OUT/diff-$s.txt") diff lines)"
		fail=1
	fi
done
echo "import check: $( [ $fail = 0 ] && echo every render equal || echo a render differs )"
