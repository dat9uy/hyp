#!/usr/bin/env bash
# A complete hyp-prototype walkthrough in a throwaway directory, using only
# commands from SKILL.md. Run with HYP=/path/to/hyp to use a specific binary.
set -euo pipefail

HYP=${HYP:-hyp}
work=$(mktemp -d)
trap 'rm -rf "$work"' EXIT
cd "$work"

# First word of a write command's output is the new node's display id.
id() { "$HYP" "$@" | awk 'NR==1{print $1}'; }

export HYP_ACTOR=investigator-1
"$HYP" init
goal=$(id add goal --title "Dashboard loads in under 1 s on seed data" --line "[P] The dashboard renders in under 1 s on the seed data" --why "owner request")
q=$(id add dq --parent "$goal" --title "Where should totals be computed?" --body "On the seed data, nightly totals meet line 1 and per-request sums do not, because the page then sums every order row on each load." --why "slow page")
h1=$(id add hyp --parent "$q" --title "Precompute totals in a nightly job" --why "rival A")
h2=$(id add hyp --parent "$q" --title "Compute totals per request with an index" --why "rival B")
e=$(id add exp --parent "$h1" --title "Read both designs against the page's query" --recipe none --rubric reading --target "$h2" --aims "$goal#1" --why "excludes one rival")

mkdir -p ".tree/evidence/$e"
echo "nightly: the page reads one totals row; per-request: the page sums every order row on each load" > ".tree/evidence/$e/notes.txt"
"$HYP" record --exp "$e" --verdict "$h1=supports" --verdict "$h2=contradicts" \
  --observation "Precomputed totals are one row read; per-request sums every order row on each load" \
  --evidence ".tree/evidence/$e/notes.txt" --closes "$goal#1" --why "read both designs"

"$HYP" select "$q"
"$HYP" select "$q" "$h1" --why "only supported candidate"
"$HYP" checkpoint --summary "Totals question answered: nightly precompute. Next: schedule the job." --why "end of session"
"$HYP" status
"$HYP" check
