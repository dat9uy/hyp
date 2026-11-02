#!/usr/bin/env bash
# Launch: make wt hold only what the runner may see. Reads the inputs file
# (--inputs FILE), KEY=VALUE lines:
#   arm      required; "base", or the name of a variant arm ([a-z][a-z0-9-]{0,31})
#   trial    required; the trial's number
#   variant  required for a non-base arm, absent for base: a directory of
#            <skill>/SKILL.md files, one per hyp skill the arm changes
# The model is always sonnet. Exits 2 on a bad inputs file or variant.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans METHOD.md"
SKILLS="hyp-prototype hyp-verification"
in="$HYP_SANDBOX_INPUTS"
die() { echo "launch: $*" >&2; exit 2; }
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
arm=$(val arm); trial=$(val trial); variant=$(val variant)
[ -n "$arm" ] || die "inputs file has no arm= line"
[ -n "$trial" ] || die "inputs file has no trial= line"
printf '%s' "$trial" | grep -Eq '^[0-9]+$' || die "trial $trial is not a number"
bundle=$PWD
state="$HYP_SANDBOX_RECIPE_TMP/state.env"
{ echo "arm=$arm"; echo "trial=$trial"; echo "model=sonnet"; } > "$state"
if [ "$arm" = base ]; then
  [ -z "$variant" ] || die "arm base takes no variant="
else
  printf '%s' "$arm" | grep -Eq '^[a-z][a-z0-9-]{0,31}$' || die "arm $arm is not an alias-shaped name"
  [ "$arm" != repro ] || die "arm repro is the verdict's own label"
  [ -n "$variant" ] || die "arm $arm has no variant= line"
  [ -d "$variant" ] || die "variant $variant is not a directory"
  [ ! -L "$variant" ] || die "variant $variant is a symlink"
  [ -n "$(ls -A "$variant")" ] || die "variant $variant changes no skill"
  for e in "$variant"/* "$variant"/.[!.]*; do
    [ -e "$e" ] || [ -L "$e" ] || continue
    n=$(basename "$e")
    case " $SKILLS " in *" $n "*) ;; *) die "variant holds $n, which is not a hyp skill";; esac
    [ -d "$e" ] && [ ! -L "$e" ] || die "variant/$n is not a directory"
    [ -f "$e/SKILL.md" ] && [ ! -L "$e/SKILL.md" ] || die "variant/$n has no SKILL.md"
    [ "$(ls -A "$e")" = SKILL.md ] || die "variant/$n holds more than SKILL.md"
  done
fi
cd "$HYP_SANDBOX_WT"
[ -d .tree/events ] || die "wt has no .tree/events; create the sandbox with --store and --cut"
rm -rf $STRIP
rm -rf recipes/skill-gate@*
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] && cp -R "skills/$s" .claude/skills/
  if [ -n "$variant" ] && [ -f "$variant/$s/SKILL.md" ]; then
    cp "$variant/$s/SKILL.md" ".claude/skills/$s/SKILL.md"
    cp "$variant/$s/SKILL.md" "skills/$s/SKILL.md"
    echo "variant_sha256_$s=$(shasum -a 256 "$variant/$s/SKILL.md" | cut -d' ' -f1)" >> "$state"
  fi
done
# The runner's brief goes on stdin only, so it has no file in wt.
cp "$bundle/brief.txt" "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm trial $trial; stripped $STRIP"
