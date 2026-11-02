#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
in="${HYP_SANDBOX_INPUTS:?}"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
arm=$(val arm); cut=$(val cut)
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
[ -s "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "the brief is missing or empty"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md METHOD.md .agentkit plans investigations recipes; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
[ "$(ls .claude/skills | tr '\n' ' ')" = "hyp-prototype hyp-verification " ] || fail ".claude/skills does not hold exactly the two hyp skills"
for s in hyp-prototype hyp-verification; do
  want=$(git show "HEAD:skills/$s/SKILL.md" 2>/dev/null) || fail "HEAD has no skills/$s/SKILL.md"
  if [ "$arm" != base ] && [ -f "$(val variant)/$s/SKILL.md" ]; then want=$(cat "$(val variant)/$s/SKILL.md"); fi
  for p in ".claude/skills/$s/SKILL.md" "skills/$s/SKILL.md"; do
    [ "$(cat "$p")" = "$want" ] || fail "$p is not the text arm $arm promises"
  done
done
last=$(hyp events --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["events"][-1]["id"])') || fail "hyp events does not read wt/.tree"
[ "$last" = "$cut" ] || fail "wt/.tree ends at $last, not the cut $cut"
echo "doctor: ok ($(hyp version)); arm $arm; cut $cut"
