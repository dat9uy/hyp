#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
command -v claude >/dev/null || fail "claude is not on PATH (this runner drives Claude Code)"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans skills; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
for s in hyp-prototype hyp-verification; do
  [ -f ".claude/skills/$s/SKILL.md" ] || fail "no skill $s"
done
[ -d .tree/events ] || fail "wt holds no tree at the cut (.tree/events)"
[ -f "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "Launch wrote no brief"
val() { sed -n "s/^$1=//p" "$HYP_SANDBOX_RECIPE_TMP/state.env" | head -n 1; }
variant=$(val variant)
if [ "$(val arm)" != base ]; then
  for d in "$variant"/*; do
    cmp -s "$d/SKILL.md" ".claude/skills/$(basename "$d")/SKILL.md" || fail "variant $(basename "$d") is not laid over the skill"
  done
fi
echo "doctor: ok ($(hyp version))"
