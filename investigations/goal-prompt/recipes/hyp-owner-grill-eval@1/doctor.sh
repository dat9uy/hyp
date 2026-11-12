#!/usr/bin/env bash
# Doctor: the grill's environment is the one Launch promised.
set -u
fail() { echo "doctor: $*" >&2; exit 1; }
RT=$HYP_SANDBOX_RECIPE_TMP
get() { grep "^$1=" "$RT/state.txt" | head -1 | cut -d= -f2-; }
for c in claude python3 git; do command -v "$c" >/dev/null || fail "$c is not on PATH"; done
hyp version >/dev/null 2>&1 || fail "hyp on PATH does not run"
cd "$HYP_SANDBOX_WT" || fail "no wt"
case "$(get fixture)" in
  a) commit=a51f1408eb396f4b2aab2f2418d5f5bc4be9396c ;;
  b) commit=9c8e498b87683e50a94a410b6a3cfecdd7a0d441 ;;
  *) fail "fixture must be a or b" ;;
esac
[ "$(git rev-parse HEAD)" = "$commit" ] || fail "HEAD is not the fixture commit $commit"
case "$(get model)" in *sonnet*) ;; *) fail "model $(get model) is not a Sonnet model" ;; esac

got=$(shasum -a 256 .claude/skills/hyp-owner-grill/SKILL.md | cut -d' ' -f1)
[ "$got" = "$(get skill_sha)" ] || fail "loaded hyp-owner-grill is $got, want $(get skill_sha)"
[ "$(shasum -a 256 .claude/skills/ak-skill-creator/scripts/quick_validate.py | cut -d' ' -f1)" = d81f883a9ddbb87fa2b015781e42041811cb14750dd3a703cfaeca409c1d7247 ] || fail "quick_validate.py is not the pinned one"
[ "$(shasum -a 256 .claude/skills/ak-skill-creator/scripts/eval_skill.py | cut -d' ' -f1)" = c47a38dee2456cc6172dff28472052c0f7b28d7baf310af6e83b5ca655d10a24 ] || fail "eval_skill.py is not the pinned one"
[ -z "$(ls .claude | grep -v '^skills$')" ] || fail ".claude holds more than skills/"
[ ! -e .tree ] || fail "a store is at the clone's top directory before the grill"

[ "$(git for-each-ref --format='%(refname)')" = "refs/heads/work" ] || fail "refs other than work remain"
[ -z "$(git remote)" ] || fail "a remote remains"
while read -r sha; do
  [ -n "$sha" ] || continue
  git cat-file -e "$sha^{commit}" 2>/dev/null && fail "later commit $sha is still in the clone"
done < "$RT/later.txt"
[ "$(git rev-list --all | wc -l)" = "$(git rev-list HEAD | wc -l)" ] || fail "a commit outside HEAD's history is reachable"

[ -s "$RT/prompt.txt" ] && head -c 17 "$RT/prompt.txt" | grep -q '^/hyp-owner-grill ' || fail "prompt.txt does not start with /hyp-owner-grill"
grep -q '^You stand in for the owner' "$RT/standin-system.txt" || fail "standin-system.txt is not the stand-in's instructions"
grep -q '"enabled":true' "$RT/claude-settings.json" || fail "Claude Code's sandbox is not enabled"
[ -z "$(GIT_CONFIG_GLOBAL=$RT/gitconfig GIT_CONFIG_NOSYSTEM=1 git config credential.helper)" ] || fail "a credential helper is set"
if command -v gh >/dev/null; then
  env -u GH_TOKEN -u GITHUB_TOKEN GH_CONFIG_DIR="$RT/gh" gh auth status >/dev/null 2>&1 && fail "gh is logged in under the trial's config"
fi
echo "doctor: ok ($(hyp version); skill $got; $(git rev-parse --short HEAD) on work)"
