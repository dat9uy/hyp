#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required; variant
# (the directory of SKILL.md files laid over the skills) is required when arm
# is not base; brief (a file replacing brief.txt) is optional.
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
arm=$(val arm); variant=$(val variant); brief=$(val brief)
if [ "$arm" != base ]; then
  [ -n "$variant" ] || { echo "launch: arm $arm has no variant= line" >&2; exit 2; }
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
fi
bundle=$PWD
[ -n "$brief" ] || brief="$bundle/brief.txt"
[ -f "$brief" ] || { echo "launch: no brief at $brief" >&2; exit 2; }
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] && cp -R "skills/$s" .claude/skills/
done
if [ "$arm" != base ]; then
  # A variant holds <skill>/SKILL.md, one per skill it changes.
  n=0
  for f in "$variant"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    case " $SKILLS " in *" $s "*) ;; *) echo "launch: variant has SKILL.md for $s, not a hyp skill" >&2; exit 2;; esac
    cp "$f" ".claude/skills/$s/SKILL.md"
    cp "$f" "skills/$s/SKILL.md"
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no <skill>/SKILL.md" >&2; exit 2; }
fi
# The runner's brief: brief, then method.md, on stdin only, so the method has
# no file in wt.
cat "$brief" "$bundle/method.md" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
