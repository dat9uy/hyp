#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required; any
# arm but "base" also needs variant=DIR, a directory of <skill>/SKILL.md
# files laid over the skills of the sandbox's commit. Everything Launch
# writes into wt before the manifest is not a change.
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
variant=
if [ "$arm" != base ]; then
  variant=$(val variant)
  [ -n "$variant" ] || { echo "launch: arm $arm needs a variant= line" >&2; exit 2; }
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  n=0
  for f in "$variant"/*; do
    [ -e "$f" ] || continue
    s=$(basename "$f")
    case " $SKILLS " in *" $s "*) ;; *) echo "launch: variant holds $s, not a skill" >&2; exit 2;; esac
    [ -f "$f/SKILL.md" ] || { echo "launch: variant $s holds no SKILL.md" >&2; exit 2; }
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
done
: > "$HYP_SANDBOX_RECIPE_TMP/variant.sha256"
if [ -n "$variant" ]; then
  for f in "$variant"/*; do
    [ -e "$f" ] || continue
    s=$(basename "$f")
    cp "$f/SKILL.md" ".claude/skills/$s/SKILL.md"
    [ -d "skills/$s" ] && cp "$f/SKILL.md" "skills/$s/SKILL.md"
    printf '%s  %s/SKILL.md\n' "$(shasum -a 256 "$f/SKILL.md" | cut -d' ' -f1)" "$s" >> "$HYP_SANDBOX_RECIPE_TMP/variant.sha256"
  done
fi
# The runner's brief goes on stdin only, so it has no file in wt.
cp "$bundle/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
