#!/usr/bin/env bash
# Launch: make wt look like a repository holding only hyp, its two skills
# (with the arm's variant over them), the tree at the cut, the recipe the
# method runs, the fixture and the arm binaries; the method goes in the brief. Reads the
# inputs file: KEY=VALUE lines arm, trial, model, and for each of today,
# text-warn, text-refuse and cite, bin_<arm> (an absolute path) and
# sha_<arm> (its sha256).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
in="$HYP_SANDBOX_INPUTS"
val() { sed -n "s/^$1=//p" "$in" | head -n 1; }
need() { [ -n "$(val "$1")" ] || { echo "launch: inputs file has no $1= line" >&2; exit 2; }; }
for k in arm trial model; do need "$k"; done
arm=$(val arm)
bundle=$PWD
if [ "$arm" != base ] && [ ! -d "$bundle/variants/$arm" ]; then
  echo "launch: arm $arm is neither base nor a directory under variants/" >&2; exit 2
fi
cd "$HYP_SANDBOX_WT"
[ -d .tree/events ] || { echo "launch: wt has no .tree: create the sandbox with --store and --cut" >&2; exit 2; }
mkdir -p arms
: > arms/arms.sha256
for a in today text-warn text-refuse cite; do
  need "bin_$a"; need "sha_$a"
  bin=$(val "bin_$a"); want=$(val "sha_$a")
  [ -f "$bin" ] || { echo "launch: bin_$a $bin is not a file" >&2; exit 2; }
  got=$(shasum -a 256 "$bin" | cut -d' ' -f1)
  [ "$got" = "$want" ] || { echo "launch: bin_$a has sha256 $got, want $want" >&2; exit 2; }
  cp "$bin" "arms/arm-$a"; chmod 755 "arms/arm-$a"
  echo "$want  arm-$a" >> arms/arms.sha256
done
cp "$in" "$HYP_SANDBOX_RECIPE_TMP/state.env"
mv recipes/id-citation-arms@1 "$HYP_SANDBOX_RECIPE_TMP/keep-recipe"
rm -rf AGENTS.md CLAUDE.md .claude .agentkit plans investigations recipes
mkdir -p recipes .claude/skills
mv "$HYP_SANDBOX_RECIPE_TMP/keep-recipe" recipes/id-citation-arms@1
cp -R skills/hyp-prototype skills/hyp-verification .claude/skills/
if [ "$arm" != base ]; then
  for f in "$bundle/variants/$arm"/*/SKILL.md; do
    s=$(basename "$(dirname "$f")")
    [ -d ".claude/skills/$s" ] || mkdir -p ".claude/skills/$s"
    cp "$f" ".claude/skills/$s/SKILL.md"
  done
fi
mkdir -p fixture
cp "$bundle/fixture/fixture.tar" "$bundle/fixture/cases.tsv" "$bundle/fixture/ids.txt" fixture/
cat "$bundle/brief.txt" "$bundle/method.md" > "$HYP_SANDBOX_RECIPE_TMP/brief.txt"
echo "launch: arm $arm; wt holds hyp, its skills, .tree at the cut, id-citation-arms@1, fixture/ and arms/; the method is in the brief only"
