#!/usr/bin/env bash
# Doctor: the trial's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
rt=$HYP_SANDBOX_RECIPE_TMP
bundle=$PWD
. "$bundle/fixture/ids.txt"
arm=$(sed -n 's/^arm=//p' "$HYP_SANDBOX_INPUTS" | head -n 1)
hyp="$rt/bin/hyp"
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
"$hyp" version >/dev/null 2>&1 || fail "the arm's hyp does not run"
[ "$(shasum -a 256 "$hyp" | cut -d' ' -f1)" = "$(cat "$rt/arm.sha256")" ] || fail "the arm's hyp does not match arm.sha256"
"$hyp" audit --help >/dev/null 2>&1 && fail "arm $arm has hyp audit"
cd "$HYP_SANDBOX_WT" || fail "no wt"
[ "$(ls -A)" = ".tree" ] || fail "wt holds more than .tree: $(ls -A | tr '\n' ' ')"
show=$("$hyp" show "$QUESTION") || fail "$QUESTION does not show"
printf '%s\n' "$show" | grep -qx 'status: open/awaiting_verifier' || fail "$QUESTION does not read open/awaiting_verifier"
printf '%s\n' "$show" | grep -qE 'uncontained|questionable' && fail "$QUESTION shows a mark"
"$hyp" events --json | grep -qiE 'seed|control|plant|harness' && fail "the events hold a telling word"
[ -s "$rt/brief.txt" ] || fail "brief.txt is missing or empty"
grep -q '<Q-id>' "$rt/brief.txt" && fail "brief.txt still holds <Q-id>"
grep -q '<topic>' "$rt/brief.txt" && fail "brief.txt still holds <topic>"
grep -qF "$QUESTION" "$rt/brief.txt" || fail "brief.txt does not name $QUESTION"
for id in "$SELECTED" "$SELECT_EVENT" "$TARGET_DECIDING" "$TARGET_LONE" "$TARGET_LOADER" "$CONTROL" "$EXP_DECIDING" "$EXP_LONE" "$EXP_LOADER" "$EXP_CONTROL"; do
  grep -qF "$id" "$rt/brief.txt" && fail "brief.txt names $id"
done
case "$arm" in
  session|reader)
    [ -s "$rt/reader-brief.txt" ] || fail "arm $arm has no reader brief"
    grep -q '<R-id>' "$rt/reader-brief.txt" || fail "the reader brief holds no <R-id>"
    if [ "$arm" = reader ]; then
      grep -qF 'on each Hypothesis' "$rt/reader-brief.txt" || fail "arm reader's brief does not send the reader to the Hypothesis"
    else
      grep -qF 'on each Hypothesis' "$rt/reader-brief.txt" && fail "arm session's brief sends the reader to the Hypothesis"
    fi
    ;;
  *)
    [ ! -e "$rt/reader-brief.txt" ] || fail "arm $arm has a reader brief"
    ;;
esac
echo "doctor: ok; arm $arm; hyp $(cat "$rt/arm.sha256")"
