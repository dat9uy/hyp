#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
bundle=$PWD
command -v claude >/dev/null || fail "claude is not on PATH (this runner drives Claude Code)"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
[ -f "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "Launch wrote no brief"
arm=$(sed -n 's/^arm=//p' "$HYP_SANDBOX_RECIPE_TMP/state.env" | head -n 1)
[ -n "$arm" ] || fail "Launch wrote no arm"
if [ "$arm" != base ]; then
  for f in "$bundle/variants/$arm"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    cmp -s "$f" ".claude/skills/$s/SKILL.md" || fail "arm $arm's $s is not laid over .claude/skills"
  done
fi
echo "doctor: ok arm $arm ($(hyp version))"
