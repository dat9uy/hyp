#!/bin/sh
# Run the review's returned test file $2 at candidate commit $1 in a detached checkout; output to $3.
S=/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/55d8dd85-ee7f-47ed-9ca9-1d64f08971d0/scratchpad
MAIN=/Users/datguy/codingProjects/tree-based-agent-memory-feat-record
WT=$S/returned-wt
rm -rf "$WT"; git -C "$MAIN" worktree prune
git -C "$MAIN" worktree add -q --detach "$WT" "$1" || exit 3
cp "$2" "$WT/internal/cli/record_review_test.go"
cd "$WT" || exit 3
{ echo "returned tests at $(git rev-parse HEAD)"; go test ./internal/cli/ -count=1 -v -run 'TestReview' 2>&1; echo "exit $?"; } > "$3"
cd "$S"; git -C "$MAIN" worktree remove --force "$WT"
grep -E "^(--- |exit)|_test.go:" "$3" | cut -c1-220
