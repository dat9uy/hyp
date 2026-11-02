#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines: arm, trial and model are required, model
# must be sonnet, and variant=DIR is required for every arm but base and
# refused for base. DIR holds <skill>/SKILL.md for each skill the arm changes.
# Everything Launch writes into wt before the manifest is not a change.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans"
SKILLS="hyp-prototype hyp-verification"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
die() { echo "launch: $*" >&2; exit 2; }
for k in arm trial model; do
  [ -n "$(val "$k")" ] || die "inputs file has no $k= line"
done
[ "$(val model)" = sonnet ] || die "model is $(val model), the method pins sonnet"
arm=$(val arm); variant=$(val variant)
if [ "$arm" = base ]; then
  [ -z "$variant" ] || die "the base arm takes no variant="
else
  [ -n "$variant" ] || die "arm $arm needs a variant=DIR line"
  [ -d "$variant" ] || die "variant $variant is not a directory"
  variant=$(cd "$variant" && pwd -P)
  n=0
  for e in "$variant"/*; do
    [ -e "$e" ] || continue
    s=$(basename "$e")
    case " $SKILLS " in *" $s "*) ;; *) die "variant holds $s, which is not a hyp skill";; esac
    [ -f "$e/SKILL.md" ] || die "variant $s has no SKILL.md"
    n=$((n+1))
  done
  [ "$n" -gt 0 ] || die "variant $variant changes no skill"
fi
bundle=$PWD
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] || die "the commit has no skills/$s"
  cp -R "skills/$s" .claude/skills/
  if [ -n "$variant" ] && [ -f "$variant/$s/SKILL.md" ]; then
    # laid over both the loaded copy and the commit's own skills/ copy, so
    # the runner finds one text of the skill wherever it looks
    cp "$variant/$s/SKILL.md" ".claude/skills/$s/SKILL.md"
    cp "$variant/$s/SKILL.md" "skills/$s/SKILL.md"
  fi
done
# The runner's brief goes on stdin only, so it has no file in wt.
cat "$bundle/brief.txt" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $(val trial); stripped $STRIP; variant ${variant:-none}"
