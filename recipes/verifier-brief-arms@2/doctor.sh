#!/usr/bin/env bash
# Doctor: the verifier's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
CUT=01M5Q03CGKKXTZRYKYZW4PYKPJ
QID=Q-KW0PWZ
TARGET=R-KA7Y6Z
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
arm=$(sed -n 's/^arm=//p' "$HYP_SANDBOX_INPUTS" | head -n 1)
rt=$HYP_SANDBOX_RECIPE_TMP
cd "$HYP_SANDBOX_WT" || fail "no wt"
[ "$(ls -A)" = ".tree" ] || fail "wt holds more than .tree: $(ls -A | tr '\n' ' ')"
last=$(hyp events --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["events"][-1]["id"])') || fail "hyp events does not read wt/.tree"
[ "$last" = "$CUT" ] || fail "wt/.tree ends at $last, not the cut $CUT"
hyp show "$QID" | grep -qx 'status: open/awaiting_verifier' || fail "$QID does not read open/awaiting_verifier"
check() { # sha256, file under .tree/evidence
  [ "$(shasum -a 256 ".tree/evidence/$2" | cut -d' ' -f1)" = "$1" ] || fail "$2 is not the cited evidence"
}
check 99a06a038a37d715bdd69eb67281002b51ab65fccbd1734ba466fbcede9bcc91 E-171ACT/e3.rb
check 3ce455e1d56275052d1afac07779f69538aeba2d0f6512e0f19d94b4e2aa7147 E-171ACT/run.log
check 149f8ea78d6e07bc05f82717f4ff59bfdee86773b9118106ac09f567b58f0804 E-X94VRM/e6.rb
check 789fa7865fb7531c94d8ce6ceafd1f0f68da8aea4a342587fcfd83705a9a44c9 E-X94VRM/run.log
[ -s "$rt/brief.txt" ] || fail "brief.txt is missing or empty"
grep -q '<Q-id>' "$rt/brief.txt" && fail "brief.txt still holds <Q-id>"
grep -q '<topic>' "$rt/brief.txt" && fail "brief.txt still holds <topic>"
grep -q "$TARGET" "$rt/brief.txt" && fail "brief.txt names $TARGET"
grep -q '^## ' "$rt/brief.txt" && fail "brief.txt holds a heading: it runs past its section"
[ ! -e "$rt/reader-brief.txt" ] || fail "arm $arm has a reader brief, and no arm of this version has a reader"
[ -z "$(grep -rl --exclude-dir=.tree "$TARGET" . 2>/dev/null)" ] || fail "$TARGET is named outside .tree"
if [ "$arm" = base ]; then
  cmp -s "$rt/brief.txt" "$rt/base-brief.txt" || fail "arm base has a brief that differs from the base commit's"
else
  cmp -s "$rt/brief.txt" "$rt/base-brief.txt" && fail "arm $arm has the base commit's brief"
fi
echo "doctor: ok ($(hyp version)); cut $CUT; arm $arm"
