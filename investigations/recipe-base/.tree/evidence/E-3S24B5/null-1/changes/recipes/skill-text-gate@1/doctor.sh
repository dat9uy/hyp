#!/usr/bin/env bash
set -euo pipefail
. "$(dirname "$0")/lib.sh"
wt=$HYP_SANDBOX_WT
tmp=$HYP_SANDBOX_RECIPE_TMP
arm=$(inputs_get arm)
variant=$(inputs_get variant)
for s in hyp-prototype hyp-verification; do
	[ -f "$wt/skills/$s/SKILL.md" ] || die "skills/$s/SKILL.md missing"
	diff -r "$wt/skills/$s" "$wt/.claude/skills/$s" >/dev/null || die ".claude/skills/$s differs from skills/$s"
done
[ "$(ls -A "$wt/.claude" | tr '\n' ' ')" = "skills " ] || die ".claude holds more than skills"
n=$(ls -A "$wt/.claude/skills" | wc -l | tr -d ' ')
[ "$n" = "$(ls -A "$wt/skills" | wc -l | tr -d ' ')" ] || die ".claude/skills lists other skills than skills/"
if [ -n "$(find "$wt" \( -name .git -o -name .tree \) -prune -o \( -name AGENTS.md -o -name CLAUDE.md \) -print)" ]; then
	die "an agent file is in the sandbox"
fi
for p in plans docs/plans .plans METHOD.md; do
	[ ! -e "$wt/$p" ] || die "$p is in the sandbox"
done
grep -q '[^[:space:]]' "$tmp/brief.txt" || die "empty brief"
[ -f "$tmp/cut-events.txt" ] || die "no cut-events.txt"
if [ -d "$wt/.tree" ]; then
	[ "$(wc -l < "$tmp/cut-events.txt" | tr -d ' ')" -gt 0 ] || die "a store with no event"
	( cd "$wt" && hyp status >/dev/null ) || die "hyp status fails on the cut"
fi
# Each skill file is the commit's, except the variant's.
base=$(git -C "$wt" rev-parse HEAD)
git -C "$wt" ls-tree -r --name-only "$base" -- skills | while read -r f; do
	rel=${f#skills/}
	if [ "$arm" != base ] && grep -qxF "$rel" "$tmp/variant-files.txt"; then
		cmp -s "$variant/$rel" "$wt/$f" || die "$f differs from the variant's"
	else
		git -C "$wt" diff --quiet HEAD -- "$f" || die "$f differs from the commit's"
	fi
done
if [ "$arm" != base ]; then
	while read -r rel; do
		git -C "$wt" cat-file -e "HEAD:skills/$rel" 2>/dev/null || die "variant file $rel is not a skill of the commit"
	done < "$tmp/variant-files.txt"
fi
command -v claude >/dev/null || die "claude not on PATH"
hyp version >/dev/null || die "hyp does not run"
