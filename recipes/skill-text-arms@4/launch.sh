#!/usr/bin/env bash
# Launch: make wt look like a repository holding only hyp, its two skills
# (with the arm's SKILL.md patch applied), the tree at the cut, the recipe
# the method runs, the fixture and the four id-citation Arm binaries; the
# method goes in the brief. Reads the inputs file: KEY=VALUE lines arm,
# trial and model. An arm other than base is the trial's one --patch,
# $HYP_SANDBOX_PATCHES/<arm>.patch, touching only skills/*/SKILL.md; base
# takes no patch. The four binaries are built here from fixture/arms/: the
# commit in BASE unpatched as the Arm named in base_arm, and each patch
# applied alone to it.
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
in="$HYP_SANDBOX_INPUTS"
die() { echo "launch: $*" >&2; exit 2; }
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
need() { [ -n "$(val "$1")" ] || die "inputs file has no $1= line"; }
for k in arm trial model; do need "$k"; done
arm=$(val arm)
bundle=$PWD
rt=$HYP_SANDBOX_RECIPE_TMP

# The trial's patches: none for base, exactly <arm>.patch otherwise.
patches=""
if [ -n "${HYP_SANDBOX_PATCHES:-}" ] && [ -d "$HYP_SANDBOX_PATCHES" ]; then
  patches=$(cd "$HYP_SANDBOX_PATCHES" && ls)
fi
variant=""
if [ "$arm" = base ]; then
  [ -z "$patches" ] || die "arm base takes no patch, but got: $(echo $patches)"
else
  printf '%s\n' "$arm" | grep -Eqx '[a-z][a-z0-9-]{0,39}' || die "arm $arm is not an Arm name"
  [ "$patches" = "$arm.patch" ] || die "arm $arm needs exactly one patch, $arm.patch (--patch $arm=FILE); got: $(echo $patches)"
  variant="$HYP_SANDBOX_PATCHES/$arm.patch"
  touched=$(git -C "$HYP_SANDBOX_WT" apply --numstat "$variant" | cut -f3-) || die "patch $arm does not parse"
  [ -n "$touched" ] || die "patch $arm changes no file"
  bad=$(printf '%s\n' "$touched" | grep -Evx 'skills/[A-Za-z0-9._-]+/SKILL\.md' || true)
  [ -z "$bad" ] || die "patch $arm touches more than skills/*/SKILL.md: $(echo $bad)"
fi

cd "$HYP_SANDBOX_WT"
[ -d .tree/events ] || die "wt has no .tree: create the sandbox with --store and --cut"

# The four id-citation Arms, each built twice in two clean copies of wt at
# BASE, with CGO_ENABLED=0 go build -trimpath; the two sha256 values must agree.
fbase=$(cat "$bundle/fixture/arms/BASE")
fname=$(cat "$bundle/fixture/arms/base_arm")
git cat-file -e "$fbase^{commit}" 2>/dev/null || die "fixture base $fbase is not in the host's history"
mkdir -p "$rt/arms" "$rt/src"
: > "$rt/arms.sha256"
one() { # arm, patch file or empty
  local a=$1 patch=$2 n sums=""
  for n in 1 2; do
    local src="$rt/src/$a.$n"
    rm -rf "$src"
    cp -R "$HYP_SANDBOX_WT" "$src"
    git -C "$src" checkout -q -f --detach "$fbase"
    git -C "$src" clean -q -fdx
    if [ -n "$patch" ]; then
      git -C "$src" apply "$patch" || die "fixture patch $a does not apply to $fbase"
    fi
    (cd "$src" && CGO_ENABLED=0 go build -trimpath -o "$rt/arms/arm-$a.$n" ./cmd/hyp) \
      || die "fixture Arm $a does not build"
    sums="$sums $(shasum -a 256 "$rt/arms/arm-$a.$n" | cut -d' ' -f1)"
    rm -rf "$src"
  done
  set -- $sums
  [ "$1" = "$2" ] || die "fixture Arm $a built twice gave two sha256 values: $1 and $2"
  mv "$rt/arms/arm-$a.1" "$rt/arms/arm-$a"
  rm "$rt/arms/arm-$a.2"
  echo "$1  arm-$a" >> "$rt/arms.sha256"
}
one "$fname" ""
for p in "$bundle"/fixture/arms/*.patch; do
  one "$(basename "$p" .patch)" "$p"
done
rmdir "$rt/src"
mkdir -p arms
mv "$rt"/arms/arm-* arms/
rmdir "$rt/arms"
chmod 755 arms/arm-*
cp "$rt/arms.sha256" arms/arms.sha256

# The arm's SKILL.md patch, applied to wt's skills before they are copied.
if [ -n "$variant" ]; then
  git apply "$variant" || die "patch $arm does not apply to wt"
fi

cp "$in" "$rt/state.env"
mv recipes/id-citation-arms@1 "$rt/keep-recipe"
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations recipes
mkdir -p recipes .claude/skills
mv "$rt/keep-recipe" recipes/id-citation-arms@1
cp -R skills/hyp-prototype skills/hyp-verification .claude/skills/
if [ -n "$variant" ]; then
  for f in $touched; do
    s=$(basename "$(dirname "$f")")
    mkdir -p ".claude/skills/$s"
    cp "$f" ".claude/skills/$s/SKILL.md"
  done
fi
mkdir -p fixture
cp "$bundle/fixture/fixture.tar" "$bundle/fixture/cases.tsv" "$bundle/fixture/ids.txt" fixture/
cat "$bundle/brief.txt" "$bundle/method.md" > "$rt/brief.txt"
echo "launch: arm $arm; wt holds hyp, its skills, .tree at the cut, id-citation-arms@1, fixture/ and arms/ (built from $fbase); the method is in the brief only"
