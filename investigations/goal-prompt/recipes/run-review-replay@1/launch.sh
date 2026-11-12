#!/usr/bin/env bash
# Launch: make wt the repository a replayed run saw at the cut, with the
# arm's skills loaded, no later commit reachable and no way to a remote.
# Reads the inputs file (KEY=VALUE lines; recipe.md lists them).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
RT=$HYP_SANDBOX_RECIPE_TMP
for k in arm cut cut_event trial model actor goal branch store skills_commit skill_sha skill_creator limit_min; do
  grep -q "^$k=" "$HYP_SANDBOX_INPUTS" || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
cp "$HYP_SANDBOX_INPUTS" "$RT/state.txt"
get() { grep "^$1=" "$RT/state.txt" | head -1 | cut -d= -f2-; }
store=$(get store); branch=$(get branch); goal=$(get goal); actor=$(get actor)
cd "$HYP_SANDBOX_WT"

# Every ref tip that is not in the base's history is a later commit.
: > "$RT/later.txt"
git for-each-ref --format='%(objectname)' | sort -u | while read -r sha; do
  git merge-base --is-ancestor "$sha" HEAD || echo "$sha" >> "$RT/later.txt"
done

# The arm's skills, from its commit, and ak-skill-creator from outside.
mkdir -p "$RT/cand" .claude/skills
git archive "$(get skills_commit)" skills | tar -x -C "$RT/cand"
cp -R "$RT/cand/skills/." .claude/skills/
cp -R "$(get skill_creator)" .claude/skills/ak-skill-creator
find .claude/skills -name __pycache__ -type d -prune -exec rm -rf {} +

# The store cut goes where the tree lives at the base.
[ -d .tree ] || { echo "launch: no store cut at wt/.tree" >&2; exit 2; }
rm -rf "$store/.tree"
mv .tree "$store/.tree"
if git cat-file -e "HEAD:$store/.tree/.gitignore" 2>/dev/null; then git checkout -q HEAD -- "$store/.tree/.gitignore"; fi

# One branch at the base; every other ref, reflog and unreachable object goes.
git checkout -q -B "$branch"
git for-each-ref --format='%(refname)' | grep -vx "refs/heads/$branch" | while read -r r; do git update-ref -d "$r"; done
git stash clear
git reflog expire --expire=now --all
git gc --prune=now --quiet
printf '#!/bin/sh\necho "pre-push: a replay trial never pushes" >&2\nexit 1\n' > .git/hooks/pre-push
chmod +x .git/hooks/pre-push

# Git and gh settings for the runner: no credential, no ssh, every push URL dead.
cat > "$RT/gitconfig" <<'EOF'
[user]
	name = hyp replay trial
	email = trial@localhost
[credential]
	helper =
[core]
	sshCommand = false
[url "file:///nonexistent/"]
	pushInsteadOf = https://
	pushInsteadOf = http://
	pushInsteadOf = git@
	pushInsteadOf = ssh://
EOF
mkdir -p "$RT/gh"

# The prompt, as the arm's hyp prints it at the cut.
(cd "$store" && hyp prompt "$goal" --actor "$actor" --branch "$branch") > "$RT/prompt.txt"
if [ -f "$HOME/.local/bin/hyp" ]; then shasum -a 256 "$HOME/.local/bin/hyp" | cut -d' ' -f1 > "$RT/home-hyp.before"; else echo none > "$RT/home-hyp.before"; fi
echo "launch: wt at $(git rev-parse --short HEAD) on $branch; skills from $(get skills_commit); store at $store"
