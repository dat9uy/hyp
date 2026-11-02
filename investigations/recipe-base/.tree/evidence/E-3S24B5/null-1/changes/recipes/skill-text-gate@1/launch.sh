#!/usr/bin/env bash
set -euo pipefail
. "$(dirname "$0")/lib.sh"
[ -n "${HYP_SANDBOX_INPUTS:-}" ] || die "no --inputs file"
wt=$HYP_SANDBOX_WT
tmp=$HYP_SANDBOX_RECIPE_TMP
arm=$(inputs_get arm)
trial=$(inputs_get trial)
variant=$(inputs_get variant)
[[ $arm =~ ^[a-z][a-z0-9-]{0,31}$ ]] || die "bad arm '$arm'"
[ "$arm" != repro ] || die "arm 'repro' is the verdict's own label"
[[ $trial =~ ^[1-9][0-9]*$ ]] || die "bad trial '$trial'"
if [ "$arm" = base ]; then
	[ -z "$variant" ] || die "arm base takes no variant"
else
	[ -d "$variant" ] || die "variant directory '$variant' not found"
fi
# The brief: everything after the first blank line.
awk 'f {print} NF==0 && !f {f=1}' "$HYP_SANDBOX_INPUTS" > "$tmp/brief.txt"
grep -q '[^[:space:]]' "$tmp/brief.txt" || die "the inputs file holds no brief"

# No agent files, no plans, and not the pre-registered method itself.
rm -rf "$wt/.claude" "$wt/plans" "$wt/docs/plans" "$wt/.plans" "$wt/METHOD.md"
find "$wt" \( -name .git -o -name .tree \) -prune -o \( -name AGENTS.md -o -name CLAUDE.md \) -type f -exec rm -f {} +

# The arm's variant laid over the skills.
: > "$tmp/variant-files.txt"
if [ "$arm" != base ]; then
	( cd "$variant" && find . -type f ! -name '.*' | sed 's|^\./||' | LC_ALL=C sort ) > "$tmp/variant-files.txt"
	[ -s "$tmp/variant-files.txt" ] || die "variant '$variant' holds no file"
	while read -r rel; do
		[[ $rel =~ ^[A-Za-z0-9._-]+/SKILL\.md$ ]] || die "variant file '$rel' is not <skill>/SKILL.md"
		[ -f "$wt/skills/$rel" ] || die "variant file '$rel' names no skill of the sandbox"
		cp "$variant/$rel" "$wt/skills/$rel"
	done < "$tmp/variant-files.txt"
fi

# The skills the runner loads: .claude/skills/<skill> beside nothing else.
mkdir -p "$wt/.claude/skills"
for d in "$wt"/skills/*/; do
	cp -R "${d%/}" "$wt/.claude/skills/"
done

# What each skill file is, and who supplied it.
: > "$tmp/arm.tsv"
( cd "$wt" && find skills -type f | LC_ALL=C sort ) | while read -r f; do
	rel=${f#skills/}
	src=commit
	grep -qxF "$rel" "$tmp/variant-files.txt" && src=variant
	printf '%s\t%s\t%s\n' "$f" "$(shasum -a 256 "$wt/$f" | cut -d' ' -f1)" "$src"
done > "$tmp/arm.tsv.new"
mv "$tmp/arm.tsv.new" "$tmp/arm.tsv"

# The events of the cut.
: > "$tmp/cut-events.txt"
if [ -d "$wt/.tree" ]; then
	( cd "$wt" && hyp events --json ) | python3 -I -c '
import json, sys
for e in json.load(sys.stdin)["events"]:
    print(e["id"])
' > "$tmp/cut-events.txt"
fi
