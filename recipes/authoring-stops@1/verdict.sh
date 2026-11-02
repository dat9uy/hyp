#!/usr/bin/env bash
# verdict TABLE: e-classify's rule over a probes table.
# A refusal stop is live when its probe exits non-zero: hyp itself refuses,
# so hyp can carry the stop. The rivals name two refusal stops,
# evidence-plain-dir and at-in-input.
#   both live:    h-binary=supports    h-partial=supports    h-null=contradicts
#   neither live: h-binary=contradicts h-partial=contradicts h-null=supports
#   one live:     h-binary=contradicts h-partial=contradicts h-null=contradicts
# Exits 2 when TABLE is missing or lacks either row.
set -u
[ $# -eq 1 ] && [ -f "$1" ] || { echo "usage: verdict TABLE" >&2; exit 2; }
ex() { awk -F'\t' -v s="$1" '$1==s {print $2; f=1} END {if (!f) exit 1}' "$2"; }
a=$(ex evidence-plain-dir "$1") || { echo "verdict: no evidence-plain-dir row" >&2; exit 2; }
b=$(ex at-in-input "$1") || { echo "verdict: no at-in-input row" >&2; exit 2; }
live=0; [ "$a" != 0 ] && live=$((live+1)); [ "$b" != 0 ] && live=$((live+1))
echo "live refusal stops: $live of 2"
case $live in
  2) echo "verdict: h-binary=supports"; echo "verdict: h-partial=supports"; echo "verdict: h-null=contradicts" ;;
  0) echo "verdict: h-binary=contradicts"; echo "verdict: h-partial=contradicts"; echo "verdict: h-null=supports" ;;
  *) echo "verdict: h-binary=contradicts"; echo "verdict: h-partial=contradicts"; echo "verdict: h-null=contradicts" ;;
esac
