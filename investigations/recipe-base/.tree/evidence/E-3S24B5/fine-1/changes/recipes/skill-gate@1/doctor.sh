#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
command -v claude >/dev/null || fail "claude is not on PATH (this runner drives Claude Code)"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans METHOD.md; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
[ -d .tree/events ] || fail "wt holds no tree at a cut"
for s in hyp-prototype hyp-verification; do
  [ -f ".claude/skills/$s/SKILL.md" ] || fail "no $s skill for the runner"
  cmp -s ".claude/skills/$s/SKILL.md" "skills/$s/SKILL.md" || fail "$s differs between .claude/skills and skills"
done
st="$HYP_SANDBOX_RECIPE_TMP/state.env"
for f in arm trial model; do grep -q "^$f=" "$st" || fail "state.env has no $f"; done
grep -qx 'model=sonnet' "$st" || fail "model is not sonnet"
[ -f "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "Launch wrote no brief"
echo "doctor: ok ($(hyp version))"
