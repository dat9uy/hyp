#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required; for an
# arm other than base, variant=DIR (absolute) is required: a directory of
# <skill>/SKILL.md files, one per skill the arm changes, laid over the two
# hyp skills.
# Edit STRIP and the copies below for your method; everything Launch writes
# into wt before the manifest is not a change.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
for k in arm trial model; do
  [ -n "$(val "$k")" ] || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
bundle=$PWD
arm=$(val arm)
variant=$(val variant)
if [ "$arm" != base ]; then
  [ -n "$variant" ] || { echo "launch: arm $arm needs a variant= line" >&2; exit 2; }
  [ -d "$variant" ] || { echo "launch: variant $variant is not a directory" >&2; exit 2; }
  n=0
  for f in "$variant"/*; do
    [ -f "$f/SKILL.md" ] && [ ! -L "$f/SKILL.md" ] || { echo "launch: $f is not <skill>/SKILL.md" >&2; exit 2; }
    case $(basename "$f") in hyp-prototype|hyp-verification) ;; *) echo "launch: $f names no hyp skill" >&2; exit 2;; esac
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || { echo "launch: variant $variant holds no skill" >&2; exit 2; }
fi
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in hyp-prototype hyp-verification; do
  [ -d "skills/$s" ] && cp -R "skills/$s" .claude/skills/
done
# An arm other than base: its SKILL.md files over the skills, in the loaded
# copy and in the repository's own skills/ so the runner reads one text.
if [ "$arm" != base ]; then
  for f in "$variant"/*; do
    s=$(basename "$f")
    cp "$f/SKILL.md" ".claude/skills/$s/SKILL.md"
    cp "$f/SKILL.md" "skills/$s/SKILL.md"
  done
fi
# The runner's brief: brief.txt, then method.md, on stdin only, so the
# method has no file in wt.
cat "$bundle/brief.txt" "$bundle/method.md" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $(val arm) trial $(val trial); stripped $STRIP"
