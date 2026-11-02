#!/usr/bin/env bash
# Launch: make wt look like a repository with only hyp, its two skills and
# the method. Reads the inputs file (KEY=VALUE lines: arm, trial, model).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
for k in arm trial model; do
  grep -q "^$k=" "$HYP_SANDBOX_INPUTS" || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
cp "$HYP_SANDBOX_INPUTS" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations recipes
mkdir -p .claude/skills
cp -R skills/hyp-prototype skills/hyp-verification .claude/skills/
cp "$OLDPWD/method.md" METHOD.md
cp "$OLDPWD/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: wt holds hyp, its two skills and METHOD.md"
