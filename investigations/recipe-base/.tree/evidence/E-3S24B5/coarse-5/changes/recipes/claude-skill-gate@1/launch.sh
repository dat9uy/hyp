#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required, and
# variant, the absolute path of the arm's directory of SKILL.md files, is
# required unless arm is base. A variant holds <skill>/SKILL.md, one per
# skill it changes, laid over the skills of wt's commit.
# Everything Launch writes into wt before the manifest is not a change.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial model; do
  [ -n "$(val "$k")" ] || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
arm=$(val arm); variant=$(val variant)
if [ "$arm" != base ] && [ -z "$variant" ]; then
  echo "launch: arm $arm has no variant= line" >&2; exit 2
fi
bundle=$PWD
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in hyp-prototype hyp-verification; do
  [ -d "skills/$s" ] || { echo "launch: wt has no skills/$s" >&2; exit 2; }
  cp -R "skills/$s" .claude/skills/
done
# The runner loads only .claude/skills, so the variant cannot be bypassed
# by reading the base text in skills/.
rm -rf skills
if [ "$arm" != base ] && [ -n "$variant" ]; then
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  n=0
  for d in "$variant"/*; do
    [ -e "$d" ] || continue
    s=$(basename "$d")
    [ -d "$d" ] && [ -f "$d/SKILL.md" ] && [ -d ".claude/skills/$s" ] && [ ! -L "$d/SKILL.md" ] ||
      { echo "launch: variant entry $s is not <skill>/SKILL.md of a hyp skill" >&2; exit 2; }
    cp "$d/SKILL.md" ".claude/skills/$s/SKILL.md"
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no SKILL.md" >&2; exit 2; }
fi
# The runner's brief: brief.txt, then method.md, on stdin only, so the
# method has no file in wt.
cat "$bundle/brief.txt" "$bundle/method.md" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
