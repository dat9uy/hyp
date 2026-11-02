#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm and trial are required; model, when
# given, must be sonnet (the method pins it).
#
# The arm `base` leaves the two hyp skills as the sandbox's commit has them.
# Any other arm is the directory variants/ARM/ of this bundle: one
# SKILL-NAME/SKILL.md per skill it changes, laid over the skill.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial; do
  [ -n "$(val "$k")" ] || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
model=$(val model)
[ -z "$model" ] || [ "$model" = sonnet ] || { echo "launch: model is $model, the method pins sonnet" >&2; exit 2; }
arm=$(val arm)
bundle=$PWD
variant=""
if [ "$arm" != base ]; then
  variant="$bundle/variants/$arm"
  [ -d "$variant" ] || { echo "launch: arm $arm has no directory variants/$arm" >&2; exit 2; }
  n=0
  for f in "$variant"/*/SKILL.md; do
    [ -f "$f" ] || continue
    n=$((n+1))
    case "$(basename "$(dirname "$f")")" in
      hyp-prototype|hyp-verification) ;;
      *) echo "launch: variants/$arm holds a file for an unknown skill: $f" >&2; exit 2 ;;
    esac
  done
  [ "$n" -gt 0 ] || { echo "launch: variants/$arm holds no SKILL.md" >&2; exit 2; }
fi
{ grep -v '^model=' "$in" || true; echo "model=sonnet"; } > "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in hyp-prototype hyp-verification; do
  [ -d "skills/$s" ] && cp -R "skills/$s" .claude/skills/
done
# The arm's variant lies over the skills, in the copy the runner loads and
# in the repository's own skills/ directory, which it can also read.
if [ -n "$variant" ]; then
  for f in "$variant"/*/SKILL.md; do
    [ -f "$f" ] || continue
    s=$(basename "$(dirname "$f")")
    mkdir -p ".claude/skills/$s"
    cp "$f" ".claude/skills/$s/SKILL.md"
    [ -d "skills/$s" ] && cp "$f" "skills/$s/SKILL.md"
  done
fi
# The runner's brief: brief.txt, then method.md, on stdin only, so the
# method has no file in wt.
cat "$bundle/brief.txt" "$bundle/method.md" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP"
