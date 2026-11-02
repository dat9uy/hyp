#!/usr/bin/env bash
# Launch: make wt hold only hyp's code, the two skills (with the arm's variant over
# them) and the tree at the cut. Reads the inputs file (--inputs FILE), KEY=VALUE
# lines: arm, trial, model (sonnet) and, for an arm other than base, variant (the
# absolute path of a directory holding <skill>/SKILL.md files).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
STRIP="AGENTS.md CLAUDE.md .claude .agentkit plans investigations"
SKILLS="hyp-prototype hyp-verification"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
die() { echo "launch: $*" >&2; exit 2; }
arm=$(val arm); trial=$(val trial); model=$(val model); variant=$(val variant)
[ -n "$arm" ] && [ -n "$trial" ] || die "inputs file needs arm= and trial= lines"
[ "$model" = sonnet ] || die "model= must be sonnet, got '$model'"
printf '%s' "$arm" | grep -Eq '^[a-z][a-z0-9-]{0,31}$' || die "arm '$arm' is not a label"
[ "$arm" != repro ] || die "arm repro collides with the verdict label repro"
printf '%s' "$trial" | grep -Eq '^[0-9]+$' || die "trial '$trial' is not a number"
bundle=$PWD
tmp=$HYP_SANDBOX_RECIPE_TMP

over=""
if [ "$arm" = base ]; then
  [ -z "$variant" ] || die "the base arm takes no variant="
  echo none > "$tmp/variant.sha256"
else
  [ -n "$variant" ] || die "arm $arm needs a variant= line"
  case $variant in /*) ;; *) die "variant= must be an absolute path" ;; esac
  [ -d "$variant" ] || die "variant $variant is not a directory"
  files=$(find "$variant" -type f -name SKILL.md | LC_ALL=C sort)
  [ -n "$files" ] || die "variant $variant holds no SKILL.md"
  seen=""
  for f in $files; do
    s=$(basename "$(dirname "$f")")
    case " $SKILLS " in *" $s "*) ;; *) die "$f: '$s' is not one of: $SKILLS" ;; esac
    case " $seen " in *" $s "*) die "variant has two SKILL.md for $s" ;; esac
    seen="$seen $s"; over="$over $s=$f"
  done
  : > "$tmp/variant.sha256"
  for f in $files; do
    printf '%s  %s/SKILL.md\n' "$(shasum -a 256 "$f" | cut -d' ' -f1)" "$(basename "$(dirname "$f")")" >> "$tmp/variant.sha256"
  done
fi

cp "$in" "$tmp/state.env"
cd "$HYP_SANDBOX_WT"
rm -rf $STRIP
mkdir -p .claude/skills
for s in $SKILLS; do
  [ -d "skills/$s" ] || die "wt has no skills/$s"
  cp -R "skills/$s" .claude/skills/
done
# The arm's variant goes over both copies, so no unchanged copy is left in wt.
for o in $over; do
  s=${o%%=*}; f=${o#*=}
  cp "$f" ".claude/skills/$s/SKILL.md"
  cp "$f" "skills/$s/SKILL.md"
done
# The brief goes to the runner on stdin only, so it has no file in wt.
cp "$bundle/brief.txt" "$tmp/brief.txt"
echo "launch: arm $arm trial $trial; stripped $STRIP; variant ${variant:-none}"
