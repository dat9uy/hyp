#!/bin/sh
# launch.sh: make the sandbox's working tree the one a trial's runner starts in.
# Reads $HYP_SANDBOX_INPUTS (key=value lines: arm, trial, variant, brief).
set -eu
inp=${HYP_SANDBOX_INPUTS:-}
[ -n "$inp" ] && [ -f "$inp" ] || { echo "launch: no inputs file (hyp sandbox create --inputs FILE)" >&2; exit 1; }
wt=${HYP_SANDBOX_WT:?}
tmp=${HYP_SANDBOX_RECIPE_TMP:?}

get() { sed -n "s/^$1=//p" "$inp" | tail -n 1; }
for key in $(sed -n 's/^\([^=]*\)=.*/\1/p' "$inp" | sort | uniq -d); do
	echo "launch: $key given twice in $inp" >&2; exit 1
done
for key in $(sed -n 's/^\([^=]*\)=.*/\1/p' "$inp"); do
	case $key in arm | trial | variant | brief) ;; *) echo "launch: unknown input $key" >&2; exit 1 ;; esac
done
arm=$(get arm); trial=$(get trial); variant=$(get variant); brief=$(get brief)

case $arm in
'' | *[!a-z0-9-]* | -* | [0-9]*) echo "launch: arm '$arm' is not [a-z][a-z0-9-]*" >&2; exit 1 ;;
esac
[ ${#arm} -le 32 ] && [ "$arm" != repro ] || { echo "launch: arm '$arm' is too long or reserved" >&2; exit 1; }
case $trial in
'' | *[!0-9]* | 0*) echo "launch: trial '$trial' is not a positive whole number" >&2; exit 1 ;;
esac
[ -f "$brief" ] && [ -s "$brief" ] || { echo "launch: brief '$brief' is not a non-empty file" >&2; exit 1; }
if [ "$arm" = base ]; then
	[ -z "$variant" ] || { echo "launch: the base arm takes no variant" >&2; exit 1; }
else
	[ -d "$variant" ] || { echo "launch: variant '$variant' is not a directory" >&2; exit 1; }
fi

cd "$wt"
# The two hyp skills, as the sandbox's commit has them.
skills=$(ls skills)
[ "$skills" = "hyp-prototype
hyp-verification" ] || { echo "launch: skills/ should hold exactly hyp-prototype and hyp-verification, got: $skills" >&2; exit 1; }
for s in $skills; do [ -f "skills/$s/SKILL.md" ] || { echo "launch: skills/$s/SKILL.md missing" >&2; exit 1; }; done

# A variant arm: variant/<skill>/SKILL.md over skills/<skill>/SKILL.md.
vhash=""
if [ "$arm" != base ]; then
	n=0
	for d in "$variant"/*; do
		[ -e "$d" ] || [ -L "$d" ] || break
		s=$(basename "$d")
		[ -d "$d" ] && [ ! -L "$d" ] && [ -d "skills/$s" ] || { echo "launch: $d is not a directory named after a skill" >&2; exit 1; }
		[ "$(ls -A "$d")" = SKILL.md ] && [ -f "$d/SKILL.md" ] && [ ! -L "$d/SKILL.md" ] || { echo "launch: $d should hold SKILL.md and nothing else" >&2; exit 1; }
		n=$((n + 1))
	done
	[ "$n" -ge 1 ] || { echo "launch: variant $variant holds no skill" >&2; exit 1; }
	[ "$(ls -A "$variant" | wc -l | tr -d ' ')" = "$n" ] || { echo "launch: variant $variant holds entries that are no skill directory" >&2; exit 1; }
	for d in "$variant"/*; do
		s=$(basename "$d")
		cp "$d/SKILL.md" "skills/$s/SKILL.md"
		h=$(shasum -a 256 "$d/SKILL.md" | cut -d' ' -f1)
		vhash="$vhash\"$s\": \"$h\", "
	done
fi
vhash="{${vhash%, }}"

# No agent files and no plans; .claude holds only the skills as the runner loads them.
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations
mkdir -p .claude/skills
for s in $skills; do cp -R "skills/$s" ".claude/skills/$s"; done

mkdir -p "$tmp"
cp "$brief" "$tmp/brief.md"
commit=$(git rev-parse HEAD)
bh=$(shasum -a 256 "$tmp/brief.md" | cut -d' ' -f1)
sh_=""
for s in $skills; do
	h=$(shasum -a 256 "skills/$s/SKILL.md" | cut -d' ' -f1)
	sh_="$sh_\"$s\": \"$h\", "
done
sh_="{${sh_%, }}"
printf '{"arm": "%s", "trial": %s, "commit": "%s", "brief_sha256": "%s", "variant_skills": %s, "skill_sha256": %s}\n' \
	"$arm" "$trial" "$commit" "$bh" "$vhash" "$sh_" >"$tmp/trial.json"
echo "launch: arm $arm trial $trial at $commit"
