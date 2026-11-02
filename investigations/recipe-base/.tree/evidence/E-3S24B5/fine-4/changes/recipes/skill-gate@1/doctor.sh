#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
command -v claude >/dev/null || fail "claude is not on PATH (this runner drives Claude Code)"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans investigations; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -z "$(ls -A .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
for s in hyp-prototype hyp-verification; do
  cmp -s "skills/$s/SKILL.md" ".claude/skills/$s/SKILL.md" || fail "$s: .claude/skills and skills/ differ"
done
[ -f "$HYP_SANDBOX_RECIPE_TMP/brief.txt" ] || fail "Launch wrote no brief"
[ -f "$HYP_SANDBOX_RECIPE_TMP/variant.sha256" ] || fail "Launch wrote no variant.sha256"
[ -d .tree/events ] || fail "wt holds no tree; create the sandbox with --store and --cut"
hyp status >/dev/null 2>&1 || fail "hyp status fails on the tree at the cut"
echo "doctor: ok ($(hyp version))"
