#!/bin/sh
# doctor.sh: the runner's starting point is what the method says.
set -eu
wt=${HYP_SANDBOX_WT:?}
tmp=${HYP_SANDBOX_RECIPE_TMP:?}
cd "$wt"
fail() { echo "doctor: $*" >&2; exit 1; }
for f in AGENTS.md CLAUDE.md plans .agentkit; do
	[ ! -e "$f" ] || fail "$f is still in the sandbox"
done
[ "$(ls -A .claude)" = skills ] || fail ".claude holds more than skills"
[ "$(ls .claude/skills)" = "hyp-prototype
hyp-verification" ] || fail ".claude/skills is not the two hyp skills"
for s in hyp-prototype hyp-verification; do
	diff -r "skills/$s" ".claude/skills/$s" >/dev/null || fail "skills/$s differs from .claude/skills/$s"
done
[ -d .tree/events ] || fail "no .tree/events: create with --store SRC --cut EVENT"
[ -s "$tmp/trial.json" ] && [ -s "$tmp/brief.md" ] || fail "launch left no trial.json and brief.md"
command -v claude >/dev/null || fail "claude is not on PATH"
command -v python3 >/dev/null || fail "python3 is not on PATH"
hyp version >/dev/null || fail "hyp does not run"
echo "doctor: ok"
