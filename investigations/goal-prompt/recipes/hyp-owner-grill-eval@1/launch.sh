#!/usr/bin/env bash
# Launch: make wt the repository a grill saw at its fixture, with the arm's
# hyp-owner-grill loaded, no later commit reachable and no way to a remote.
# Reads the inputs file (KEY=VALUE lines; recipe.md lists them).
set -eu
: "${HYP_SANDBOX_WT:?}" "${HYP_SANDBOX_TMP:?}" "${HYP_SANDBOX_RECIPE_TMP:?}" "${HYP_SANDBOX_INPUTS:?}"
RT=$HYP_SANDBOX_RECIPE_TMP
BUNDLE=$(cd "$(dirname "$0")" && pwd)
for k in arm fixture trial model skills_commit skill_sha skill_creator limit_min max_exchanges; do
  grep -q "^$k=" "$HYP_SANDBOX_INPUTS" || { echo "launch: inputs file has no $k= line" >&2; exit 2; }
done
cp "$HYP_SANDBOX_INPUTS" "$RT/state.txt"
get() { grep "^$1=" "$RT/state.txt" | head -1 | cut -d= -f2-; }
case "$(get fixture)" in
  a) commit=a51f1408eb396f4b2aab2f2418d5f5bc4be9396c; issue=113 ;;
  b) commit=9c8e498b87683e50a94a410b6a3cfecdd7a0d441; issue=141 ;;
  *) echo "launch: fixture must be a or b" >&2; exit 2 ;;
esac
cd "$HYP_SANDBOX_WT"
[ "$(git rev-parse HEAD)" = "$commit" ] || { echo "launch: wt is at $(git rev-parse HEAD), fixture $(get fixture) is $commit; create the sandbox with --base $commit" >&2; exit 2; }

# Every ref tip that is not in the fixture's history is a later commit.
: > "$RT/later.txt"
git for-each-ref --format='%(objectname)' | sort -u | while read -r sha; do
  git merge-base --is-ancestor "$sha" HEAD || echo "$sha" >> "$RT/later.txt"
done

# The fixture's own skills, the arm's hyp-owner-grill over them, and
# ak-skill-creator from outside.
mkdir -p "$RT/arm" .claude/skills
git archive "$(get skills_commit)" skills/hyp-owner-grill | tar -x -C "$RT/arm"
cp -R skills/. .claude/skills/
rm -rf .claude/skills/hyp-owner-grill
cp -R "$RT/arm/skills/hyp-owner-grill" .claude/skills/hyp-owner-grill
cp -R "$(get skill_creator)" .claude/skills/ak-skill-creator
find .claude/skills -name __pycache__ -type d -prune -exec rm -rf {} +
(cd "$RT/arm/skills" && tar -cf "$RT/skill.tar" hyp-owner-grill)

# One branch at the fixture; every other ref, reflog and unreachable object goes.
git checkout -q -B work
git for-each-ref --format='%(refname)' | grep -vx "refs/heads/work" | while read -r r; do git update-ref -d "$r"; done
git remote | while read -r r; do git remote remove "$r"; done
git stash clear
git reflog expire --expire=now --all
git gc --prune=now --quiet
printf '#!/bin/sh\necho "pre-push: a grill trial never pushes" >&2\nexit 1\n' > .git/hooks/pre-push
chmod +x .git/hooks/pre-push

# Git and gh settings for the grill: no credential, no ssh, every push URL dead.
cat > "$RT/gitconfig" <<'EOF'
[user]
	name = hyp grill trial
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
mkdir -p "$RT/gh" "$RT/standin"

# Claude Code's sandbox: Bash writes only inside wt and reaches no network.
printf '{"sandbox":{"enabled":true,"allowUnsandboxedCommands":false,"autoAllowBashIfSandboxed":true}}\n' > "$RT/claude-settings.json"

# The issue as of the fixture, where the grill's first message points.
cp "$BUNDLE/issue-$(get fixture).md" "$HYP_SANDBOX_TMP/issue-$issue.md"
cat "$BUNDLE/standin-system.md" "$BUNDLE/standin-$(get fixture).md" > "$RT/standin-system.txt"
printf '/hyp-owner-grill #%s (the text of issue #%s as of this session is in %s; this machine has no network, so gh cannot fetch it)\n' \
  "$issue" "$issue" "$HYP_SANDBOX_TMP/issue-$issue.md" > "$RT/prompt.txt"
if [ -f "$HOME/.local/bin/hyp" ]; then shasum -a 256 "$HOME/.local/bin/hyp" | cut -d' ' -f1 > "$RT/home-hyp.before"; else echo none > "$RT/home-hyp.before"; fi
echo "launch: wt at $(git rev-parse --short HEAD) on work; fixture $(get fixture); arm $(get arm) from $(get skills_commit)"
