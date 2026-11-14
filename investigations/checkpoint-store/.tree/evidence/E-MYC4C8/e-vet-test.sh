#!/bin/sh
# (e) go vet ./... and go test ./... -count=1 on 8a12324, in a clean detached worktree.
set -u
REPO=/Users/datguy/codingProjects/tree-based-agent-memory
WT=$1
git -C $REPO worktree add --detach $WT 8a12324 >/dev/null 2>&1 || exit 3
cd $WT
echo "== candidate $(git rev-parse HEAD)"
go vet ./... ; echo "vet exit $?"
go test ./... -count=1 2>&1 | grep -v "no test files"; echo "test exit $?"
cd $REPO && git worktree remove --force $WT
