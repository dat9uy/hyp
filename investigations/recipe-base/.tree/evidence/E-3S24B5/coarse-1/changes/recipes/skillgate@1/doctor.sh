#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
command -v claude >/dev/null || fail "claude is not on PATH (this runner drives Claude Code)"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
[ -f "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "Launch wrote no brief"
st="$HYP_SANDBOX_RECIPE_TMP/state.env"
[ -f "$st" ] || fail "Launch wrote no state.env"
v=$(sed -n 's/^variant=//p' "$st" | head -n 1)
if [ -n "$v" ]; then
  for f in "$v"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    cmp -s "$f" ".claude/skills/$s/SKILL.md" || fail "variant $s/SKILL.md is not the loaded skill"
  done
fi
echo "doctor: ok ($(hyp version))"
