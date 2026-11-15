#!/bin/sh
# Import check: in a temporary clone of this repository, the candidate hyp $1
# imports each store under investigations/ alone. For each store it prints, before
# and after, hyp status of each Goal and hyp render --markdown; logs go to $2.
HYP=$1; OUT=$2
S=/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/55d8dd85-ee7f-47ed-9ca9-1d64f08971d0/scratchpad
SRC=/Users/datguy/codingProjects/tree-based-agent-memory
CL=$S/importclone
rm -rf "$CL"; mkdir -p "$OUT/import"
git clone -q --no-local "$SRC" "$CL" || exit 3
export HYP_ACTOR=import-check
cd "$CL" || exit 3
{ echo "clone of $SRC at $(git rev-parse HEAD)"; echo "hyp $HYP"; shasum -a 256 "$HYP"; } > "$OUT/import/setup.txt"
stores=$(for d in investigations/*/; do [ -d "$d.tree/events" ] && echo "${d%/}"; done)
goals() { cat "$1"/.tree/events/*.jsonl | grep '"kind":"Goal"' | sed -E 's/.*"node":"[0-9A-Z]{20}([0-9A-Z]{6})".*/G-\1/' | sort -u; }
views() {
  ( cd "$1" && for g in $(goals .); do echo "### hyp status $g"; "$HYP" status "$g" --max-lines 100000 2>&1; echo "exit $?"; done; echo "### hyp render --markdown"; "$HYP" render --markdown 2>&1; echo "exit $?" )
}
for st in $stores; do
  n=$(basename "$st")
  views "$st" > "$OUT/import/$n-before.txt"
done
for st in $stores; do
  n=$(basename "$st")
  tip0=$(git rev-parse -q --verify refs/hyp/record)
  "$HYP" import "$st" > "$OUT/import/$n-import.txt" 2>&1
  code=$?
  echo "exit $code" >> "$OUT/import/$n-import.txt"
  tip1=$(git rev-parse -q --verify refs/hyp/record)
  views "$st" > "$OUT/import/$n-after.txt"
  if [ $code = 0 ]; then
    if cmp -s "$OUT/import/$n-before.txt" "$OUT/import/$n-after.txt"; then r="exit 0, prints as before"; else r="exit 0, PRINTS OTHERWISE"; fi
    again=$("$HYP" import "$st" 2>&1); echo "$again" > "$OUT/import/$n-import-again.txt"
    tip2=$(git rev-parse -q --verify refs/hyp/record)
    [ "$tip2" = "$(git rev-parse -q --verify refs/hyp/record)" ] && [ "$(echo "$again" | grep -c 'holds every event')" = 1 ] && r="$r; a second import adds nothing"
  else
    if [ "$tip0" = "$tip1" ]; then r="exit $code, record unchanged: $(head -1 "$OUT/import/$n-import.txt" | cut -c1-200)"; else r="exit $code, RECORD MOVED"; fi
  fi
  echo "$n: $r" | tee -a "$OUT/import/summary.txt"
done
git for-each-ref refs/hyp >> "$OUT/import/setup.txt"
cd "$S"; rm -rf "$CL"
