#!/usr/bin/env bash
# Doctor: the runner's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
RT=$HYP_SANDBOX_RECIPE_TMP
get() { grep "^$1=" "$RT/state.txt" | head -1 | cut -d= -f2-; }
for c in claude python3 go git; do command -v "$c" >/dev/null || fail "$c is not on PATH"; done
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
branch=$(get branch); store=$(get store); goal=$(get goal); actor=$(get actor)

got=$(shasum -a 256 .claude/skills/hyp-goal-run/SKILL.md | cut -d' ' -f1)
[ "$got" = "$(get skill_sha)" ] || fail "loaded hyp-goal-run is $got, want $(get skill_sha)"
[ "$(shasum -a 256 .claude/skills/ak-skill-creator/scripts/quick_validate.py | cut -d' ' -f1)" = d81f883a9ddbb87fa2b015781e42041811cb14750dd3a703cfaeca409c1d7247 ] || fail "quick_validate.py is not the pinned one"
[ "$(shasum -a 256 .claude/skills/ak-skill-creator/scripts/lint_cruft.py | cut -d' ' -f1)" = ff36baefb5aa720b3e74a4c3f71df3cdb1c88d320e8296f25ce1b961660a0946 ] || fail "lint_cruft.py is not the pinned one"
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"

[ "$(git for-each-ref --format='%(refname)')" = "refs/heads/$branch" ] || fail "refs other than $branch remain"
[ -z "$(git remote)" ] || fail "a remote remains"
while read -r sha; do
  [ -n "$sha" ] || continue
  git cat-file -e "$sha^{commit}" 2>/dev/null && fail "later commit $sha is still in the clone"
done < "$RT/later.txt"
[ "$(git rev-list --all | wc -l)" = "$(git rev-list HEAD | wc -l)" ] || fail "a commit outside HEAD's history is reachable"
[ -z "$(git fsck --unreachable --no-reflogs 2>/dev/null | grep ' commit ')" ] || fail "an unreachable commit is still in the clone"

last=$(cd "$store" && hyp events --json | python3 -c 'import json,sys; print(json.load(sys.stdin)["events"][-1]["id"])') || fail "hyp events fails in $store"
[ "$last" = "$(get cut_event)" ] || fail "the store's last event is $last, want the cut $(get cut_event)"
(cd "$store" && hyp status "$goal" >/dev/null) || fail "hyp status $goal fails"
head -c 6 "$RT/prompt.txt" | grep -q '^/goal ' || fail "prompt.txt does not start with /goal"
grep -q "as HYP_ACTOR=$actor\." "$RT/prompt.txt" || fail "prompt.txt does not name $actor"

[ -z "$(GIT_CONFIG_GLOBAL=$RT/gitconfig GIT_CONFIG_NOSYSTEM=1 git config credential.helper)" ] || fail "a credential helper is set"
if command -v gh >/dev/null; then
  env -u GH_TOKEN -u GITHUB_TOKEN GH_CONFIG_DIR="$RT/gh" gh auth status >/dev/null 2>&1 && fail "gh is logged in under the trial's config"
fi
echo "doctor: ok ($(hyp version); skill $got; cut $last; $(git rev-parse --short HEAD) on $branch)"
