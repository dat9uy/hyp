#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see: hyp's code, the two hyp
# skills (the arm's variant laid over them) and the tree at the cut.
# Reads the inputs file (--inputs FILE), KEY=VALUE lines: arm, trial and model
# are required; variant (an absolute directory) is required unless arm=base.
# Everything Launch writes into wt before the manifest is not a change.
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
if [ "$arm" = base ]; then
  [ -z "$variant" ] || { echo "launch: arm base takes no variant= line" >&2; exit 2; }
else
  case "$variant" in /*) ;; *) echo "launch: arm $arm needs an absolute variant= directory" >&2; exit 2 ;; esac
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  # A variant holds <skill>/SKILL.md, one per skill it changes, and nothing else.
  n=0
  for p in "$variant"/* ; do
    [ -e "$p" ] || break
    s=$(basename "$p")
    case " $SKILLS " in *" $s "*) ;; *) echo "launch: variant holds $s, not one of: $SKILLS" >&2; exit 2 ;; esac
    [ -d "$p" ] && [ ! -L "$p" ] && [ -f "$p/SKILL.md" ] && [ -z "$(ls -A "$p" | grep -vx SKILL.md)" ] ||
      { echo "launch: variant/$s must hold exactly SKILL.md" >&2; exit 2; }
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no <skill>/SKILL.md" >&2; exit 2; }
fi
bundle=$PWD
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] || { echo "launch: the commit has no skills/$s" >&2; exit 2; }
  cp -R "skills/$s" .claude/skills/
done
: > "$HYP_SANDBOX_RECIPE_TMP/variant.sha256"
if [ "$arm" != base ]; then
  for p in "$variant"/*; do
    s=$(basename "$p")
    cp "$p/SKILL.md" ".claude/skills/$s/SKILL.md"
    printf '%s  %s/SKILL.md\n' "$(shasum -a 256 "$p/SKILL.md" | cut -d' ' -f1)" "$s" >> "$HYP_SANDBOX_RECIPE_TMP/variant.sha256"
  done
fi
# The runner's brief goes on stdin only, so it has no file in wt.
cp "$bundle/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
