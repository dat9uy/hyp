#!/usr/bin/env bash
# probes REPO: one liveness probe per method-to-instrument stop, run with
# the hyp first on PATH. Prints a header, then "stop<TAB>exit<TAB>output",
# where output is the probe's first line with its scratch directory written
# W, $HOME written HOME and a trailing commit id written BASE. Exits 2 on a bad argument; otherwise 0, whatever the probes found.
set -u
if [ $# -ne 1 ] || [ ! -d "$1/.git" ]; then
  echo "usage: probes REPO (a git work tree)" >&2; exit 2
fi
REPO=$(cd "$1" && pwd)
W="${TMPDIR:-/tmp}/authoring-stops"
rm -rf "$W"; mkdir -p "$W"
export HYP_ACTOR=probe GOTOOLCHAIN=local GOFLAGS=-mod=mod
export GIT_AUTHOR_NAME=probe GIT_AUTHOR_EMAIL=probe@example.invalid GIT_COMMITTER_NAME=probe GIT_COMMITTER_EMAIL=probe@example.invalid
export GIT_AUTHOR_DATE=2026-01-01T00:00:00Z GIT_COMMITTER_DATE=2026-01-01T00:00:00Z
p() {
  local id=$1; shift; local out rc
  out=$("$@" 2>&1); rc=$?
  printf '%s\t%s\t%s\n' "$id" "$rc" "$(printf '%s\n' "$out" | head -1 | sed -e "s#$W#W#g" -e "s#$HOME#HOME#g" -e 's/ @ [0-9a-f]*$/ @ BASE/')"
}
printf 'stop\texit\toutput\n'

# A scratch repository with one tree and one instrument bundle with no cases.
mkdir -p "$W/r" && cd "$W/r" && git init -q . && hyp init >/dev/null
G=$(hyp add goal --title g --why w | awk '{print $1}')
Q=$(hyp add dq --parent "$G" --title q --why w | awk '{print $1}')
H=$(hyp add hyp --parent "$Q" --title h --why w | awk '{print $1}')
mkdir -p recipes/inst@1
printf -- '---\nname: inst\nversion: 1\nformat: 1\nsubject: probe\n---\n\n# inst\n\n## Launch\n\nNone.\n\n## Doctor\n\nNone.\n\n## Drive\n\nNone.\n\n## Evidence\n\nNone.\n\n## Cleanup\n\nNone.\n\n## Standalone\n\nRun m: ./m.sh\n' > recipes/inst@1/recipe.md
printf '#!/bin/sh\necho 1\n' > recipes/inst@1/m.sh && chmod +x recipes/inst@1/m.sh

# A measured Experiment with recipe none and no declaration.
p measured-none hyp add exp --parent "$H" --title e --recipe none --target none --why w
# An instrument with no known-answers case.
p no-known-answers hyp add exp --parent "$H" --title e --recipe inst@1 --target none --why w
# Evidence over a directory that no sandbox wrote.
mkdir -p "$W/plain"
p evidence-plain-dir hyp recipe evidence --record "$W/rec" inst@1 "$W/plain"
# A known-answer input holding a bundle path with '@'.
mkdir -p "$W/b/inst@1/known-answers/c/input/recipes/x@1"
cp recipes/inst@1/recipe.md recipes/inst@1/m.sh "$W/b/inst@1/" && touch "$W/b/inst@1/known-answers/c/input/recipes/x@1/recipe.md"
p at-in-input hyp recipe hash "$W/b/inst@1"
# A placeholder id in --why.
p placeholder-id hyp add hyp --parent "$Q" --title h2 --why "see R-<new>"
# An alias a node already holds.
hyp add hyp --parent "$Q" --title h3 --alias h-held --why w >/dev/null
p alias-held hyp add hyp --parent "$Q" --title h4 --alias h-held --why w
# A first bundle with no --from.
p first-bundle hyp recipe new fresh@1
# A sandbox base commit that no ref reaches.
git add -A >/dev/null && git commit -qm base
D=$(git commit-tree 'HEAD^{tree}' -p HEAD -m dangling)
p unreachable-base hyp sandbox create probe --host "$W/r" --base "$D" --recipe none
hyp sandbox remove probe --discard >/dev/null 2>&1
# A .go file inside a bundle of a Go module.
mkdir -p "$W/g/recipes/x@1" && cd "$W/g"
printf 'module m\n\ngo 1.21\n' > go.mod && printf 'package main\n\nfunc main() {}\n' > main.go && printf 'package x\n' > 'recipes/x@1/a.go'
p go-file-in-bundle go vet ./...

# The repository under study: what it still tracks.
cd "$REPO"
p tracked-trees sh -c "git ls-tree -d --name-only HEAD investigations/ fixtures/ | tr '\n' ' '"
p bundles-without-format sh -c "for f in recipes/*/recipe.md; do [ -f \"\$f\" ] && ! grep -q '^format: 1' \"\$f\" && printf '%s ' \"\$f\"; done; true"
p template-sections sh -c "grep -c -e '^## Standalone' -e 'known-answers' skills/hyp-verification/recipe-template.md"
