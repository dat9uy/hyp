#!/bin/sh
# Gate steps 1 and 3 at candidate commit $1, logs into $2: each new test alone, vet, the whole suite.
S=/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/55d8dd85-ee7f-47ed-9ca9-1d64f08971d0/scratchpad
MAIN=/Users/datguy/codingProjects/tree-based-agent-memory-feat-record
C=$1; OUT=$2; shift 2
WT=$S/cand-$C
mkdir -p "$OUT"
rm -rf "$WT"; git -C "$MAIN" worktree prune
git -C "$MAIN" worktree add -q --detach "$WT" "$C" || exit 3
cd "$WT" || exit 3
git rev-parse HEAD > "$OUT/candidate-commit.txt"
CGO_ENABLED=0 go build -o "$S/hyp-$C" ./cmd/hyp && shasum -a 256 "$S/hyp-$C" > "$OUT/candidate-binary-sha256.txt"
: > "$OUT/candidate-new-tests.txt"
for spec in "$@"; do
  pkg=${spec%%:*}; t=${spec#*:}
  echo "== $t ($pkg)" >> "$OUT/candidate-new-tests.txt"
  go test "$pkg" -count=1 -v -run "^$t\$" >> "$OUT/candidate-new-tests.txt" 2>&1
  echo "exit $?" >> "$OUT/candidate-new-tests.txt"
  if [ "$t" = TestStoreFolderNameIsPlain ]; then
    echo "== $t ($pkg), TMPDIR on a case-sensitive APFS volume" >> "$OUT/candidate-new-tests.txt"
    TMPDIR=$S/cs go test "$pkg" -count=1 -v -run "^$t\$" >> "$OUT/candidate-new-tests.txt" 2>&1
    echo "exit $?" >> "$OUT/candidate-new-tests.txt"
  fi
done
go vet ./... > "$OUT/candidate-vet.txt" 2>&1; echo "exit $?" >> "$OUT/candidate-vet.txt"
go test ./... -count=1 > "$OUT/candidate-go-test.txt" 2>&1; echo "exit $?" >> "$OUT/candidate-go-test.txt"
grep -E "^== |^exit" "$OUT/candidate-new-tests.txt" | paste - -
tail -1 "$OUT/candidate-vet.txt"; tail -1 "$OUT/candidate-go-test.txt"
cd "$S"; git -C "$MAIN" worktree remove --force "$WT"
