#!/usr/bin/env bash
# Launch: make wt hold only hyp's code, its two skills (the arm's variant over
# them) and the tree at the cut. Reads the inputs file: KEY=VALUE lines arm,
# trial, brief, cut, and variant for an arm other than base.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
need() { [ -n "$(val "$1")" ] || { echo "launch: inputs file has no $1= line" >&2; exit 2; }; }
for k in arm trial brief cut; do need "$k"; done
arm=$(val arm)
case "$(val trial)" in *[!0-9]*) echo "launch: trial is not a whole number" >&2; exit 2 ;; esac
printf '%s' "$arm" | grep -Eq '^[a-z][a-z0-9-]{0,31}$' || { echo "launch: arm $arm is not a label" >&2; exit 2; }
[ "$arm" != repro ] || { echo "launch: arm repro is reserved" >&2; exit 2; }
[ -f "$(val brief)" ] || { echo "launch: brief $(val brief) is not a file" >&2; exit 2; }
variant=
if [ "$arm" != base ]; then
  need variant
  variant=$(val variant)
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  n=0
  for f in "$variant"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    case "$s" in hyp-prototype | hyp-verification) ;; *) echo "launch: variant holds skill $s" >&2; exit 2 ;; esac
    n=$((n + 1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no <skill>/SKILL.md" >&2; exit 2; }
fi
cd "$HYP_SANDBOX_WT"
[ -d .tree/events ] || { echo "launch: wt has no .tree: create the sandbox with --store and --cut" >&2; exit 2; }
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cp "$(val brief)" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations recipes METHOD.md
mkdir -p .claude/skills
cp -R skills/hyp-prototype skills/hyp-verification .claude/skills/
if [ "$arm" != base ]; then
  for f in "$variant"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    cp "$f" ".claude/skills/$s/SKILL.md"
    cp "$f" "skills/$s/SKILL.md"
  done
fi
echo "launch: arm $arm trial $(val trial); wt holds hyp, its two skills and .tree at the cut"
