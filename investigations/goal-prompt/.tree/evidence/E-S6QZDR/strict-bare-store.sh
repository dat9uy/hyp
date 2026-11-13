#!/bin/bash
# Strict check of line 1, meaning 2: each CASES draft written by hyp into its field
# in a bare store (one Goal, one question, one Hypothesis; none of the ids the draft
# cites) exits 0. Sets no verdict.
set -u
BIN=$1; WORK=$2; CASES=/Users/datguy/codingProjects/tbam-text-cap/internal/cli/testdata/text-cap
echo "binary $BIN sha256 $(shasum -a 256 $BIN | cut -d' ' -f1)"
grep -v '^#' $CASES/CASES | while read sha field runes file rest; do
  d=$WORK/$file; rm -rf $d; mkdir -p $d; cd $d; mkdir .git
  export HYP_ACTOR=strict
  $BIN init >/dev/null; G=$($BIN add goal --title G --why w | cut -d' ' -f1); Q=$($BIN add dq --parent $G --title Q --why w | cut -d' ' -f1); H=$($BIN add hyp --parent $Q --title H --why w | cut -d' ' -f1)
  text=$(cat $CASES/$file; printf x); text=${text%x}
  case $field in
    experiment-body) out=$($BIN add exp --parent $H --title E --recipe none --rubric reading --body "$text" --why w 2>&1);;
    checkpoint-summary) out=$($BIN checkpoint --goal $G --summary "$text" --why w 2>&1);;
    goal-body) out=$($BIN add goal --title G2 --body "$text" --why w 2>&1);;
    question-title) out=$($BIN add dq --parent $G --title "$text" --why w 2>&1);;
  esac
  code=$?
  echo "$file $field $runes: exit $code: $(echo "$out" | head -c 160 | tr '\n' ' ')"
done
