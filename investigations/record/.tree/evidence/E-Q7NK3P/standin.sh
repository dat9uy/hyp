#!/bin/sh
# Run each new test of candidate commit $1 alone on stand-in commit $2; logs go to $3.
# CS_TESTS names tests that also run with TMPDIR on the case-sensitive volume.
S=/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/55d8dd85-ee7f-47ed-9ca9-1d64f08971d0/scratchpad
CAND=$1; BASE=$2; OUT=$3; shift 3
MAIN=/Users/datguy/codingProjects/tree-based-agent-memory-feat-record
WT=$S/standin-$BASE
mkdir -p "$OUT"
rm -rf "$WT"; git -C "$MAIN" worktree prune
git -C "$MAIN" worktree add -q --detach "$WT" "$BASE" || exit 3
# The new and changed test files, from the candidate.
for f in internal/cli/record_import_names_test.go internal/cli/record_git_env_test.go internal/cli/prompt_test.go internal/record/record_test.go; do
  git -C "$MAIN" show "$CAND:$f" > "$WT/$f"
done
cd "$WT" || exit 3
for spec in "$@"; do
  pkg=${spec%%:*}; t=${spec#*:}
  log="$OUT/standin-$t.txt"
  echo "stand-in $BASE, test $t in $pkg, test files from $CAND" > "$log"
  go test "$pkg" -count=1 -v -run "^$t\$" >> "$log" 2>&1
  echo "exit $?" >> "$log"
  if [ "$t" = TestStoreFolderNameIsPlain ]; then
    log="$OUT/standin-$t-casesensitive.txt"
    echo "stand-in $BASE, test $t in $pkg, TMPDIR on a case-sensitive APFS volume" > "$log"
    TMPDIR=$S/cs go test "$pkg" -count=1 -v -run "^$t\$" >> "$log" 2>&1
    echo "exit $?" >> "$log"
  fi
  tail -1 "$OUT/standin-$t.txt" | sed "s/^/$t: /"
done
cd "$S"; git -C "$MAIN" worktree remove --force "$WT"
