#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required; any
# arm but base also needs variant=DIR, a directory of <skill>/SKILL.md files
# (one per skill it changes) laid over the two hyp skills.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans"
SKILLS="hyp-prototype hyp-verification"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial model; do
  [ -n "$(val "$k")" ] || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
arm=$(val arm)
variant=$(val variant)
if [ "$arm" != base ]; then
  [ -n "$variant" ] || { echo "launch: arm $arm needs a variant= line" >&2; exit 2; }
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  n=0
  for d in "$variant"/*/; do
    [ -d "$d" ] || continue
    s=$(basename "$d")
    case " $SKILLS " in *" $s "*) ;; *) echo "launch: variant holds $s, not a hyp skill" >&2; exit 2;; esac
    [ -f "$d/SKILL.md" ] || { echo "launch: variant $s has no SKILL.md" >&2; exit 2; }
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no skill" >&2; exit 2; }
fi
bundle=$PWD
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] && cp -R "skills/$s" .claude/skills/
  if [ "$arm" != base ] && [ -f "$variant/$s/SKILL.md" ]; then
    cp "$variant/$s/SKILL.md" ".claude/skills/$s/SKILL.md"
    [ -d "skills/$s" ] && cp "$variant/$s/SKILL.md" "skills/$s/SKILL.md"
  fi
done
# The runner's brief goes on stdin only, so it has no file in wt.
cp "$bundle/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
