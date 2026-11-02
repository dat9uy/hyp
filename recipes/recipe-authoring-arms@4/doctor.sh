#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
for f in AGENTS.md CLAUDE.md .agentkit plans investigations recipes; do [ -e "$f" ] && fail "$f is still in wt"; done
[ -f METHOD.md ] || fail "METHOD.md is missing"
[ "$(ls .claude/skills | tr '\n' ' ')" = "hyp-prototype hyp-verification " ] || fail ".claude/skills holds more than the two hyp skills"
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
echo "doctor: ok ($(hyp version))"
