#!/usr/bin/env bash
# Launch: make wt look like a repository with only hyp, its two skills and
# the method. Reads the inputs file (KEY=VALUE lines: arm, trial, model;
# arm is null, coarse or fine, the label of the sandbox's --hyp binary).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
for k in arm trial model; do
  grep -q "^$k=" "$HYP_SANDBOX_INPUTS" || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
case "$(sed -n 's/^arm=//p' "$HYP_SANDBOX_INPUTS" | head -n 1)" in
  null|coarse|fine) ;;
  *) echo "launch: arm must be null, coarse or fine" >&2; exit 2 ;;
esac
cp "$HYP_SANDBOX_INPUTS" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations recipes
mkdir -p .claude/skills
cp -R skills/hyp-prototype skills/hyp-verification .claude/skills/
cp "$OLDPWD/method.md" METHOD.md
cp "$OLDPWD/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: wt holds hyp, its two skills and METHOD.md"
