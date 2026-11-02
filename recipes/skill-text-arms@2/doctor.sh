#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
CUT=01M6A7K0994S8115J0ZHBMN916
RECIPE_SHA=3fb1e1395f30d6201091602256a4e5d5261e201ec9ca27cc7178611a60210d7b
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans investigations; do [ -e "$f" ] && fail "$f is still in wt"; done
[ "$(ls recipes)" = "id-citation-arms@1" ] || fail "recipes/ holds more than id-citation-arms@1"
[ "$(hyp recipe hash recipes/id-citation-arms@1)" = "$RECIPE_SHA" ] || fail "id-citation-arms@1 is not the bundle e-arms pinned"
last=$(hyp events --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["events"][-1]["id"])') || fail "hyp events does not read wt/.tree"
[ "$last" = "$CUT" ] || fail "wt/.tree ends at $last, not the cut $CUT"
[ -e METHOD.md ] && fail "METHOD.md is in wt; the method belongs to the brief only"
for f in fixture.tar cases.tsv ids.txt; do [ -f "fixture/$f" ] || fail "fixture/$f is missing"; done
(cd arms && shasum -a 256 -c arms.sha256 >/dev/null) || fail "an arm binary does not match arms/arms.sha256"
[ "$(ls .claude/skills | tr '\n' ' ')" = "hyp-prototype hyp-verification " ] || [ -n "$(sed -n 's/^arm=//p' "$HYP_SANDBOX_INPUTS" | grep -v '^base$')" ] || fail ".claude/skills holds more than the two hyp skills"
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
echo "doctor: ok ($(hyp version)); cut $CUT"
