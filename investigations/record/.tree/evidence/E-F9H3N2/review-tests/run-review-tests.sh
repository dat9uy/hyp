#!/bin/sh
# Run each test the review of E-F9H3N2 returned, in the detached candidate
# checkout CAND (20bdb699b090), and remove the probe files afterwards.
# usage: run-review-tests.sh CAND TESTS OUT BASEHYP CANDHYP
set -u
CAND=$1
TESTS=$2
OUT=$3
cd "$CAND" || exit 3
echo "checkout: $(git rev-parse HEAD)"
cp "$TESTS/review_probe_test.go" internal/cli/review_probe_test.go
go test ./internal/cli -run TestProbe -count=1 -v > "$OUT/run-probes.txt" 2>&1
echo "probe tests: exit $?"
rm internal/cli/review_probe_test.go
cp "$TESTS/zz_boundary_probe.go" internal/render/zz_boundary_probe.go
{ go vet ./internal/render && go test ./internal/cli -run TestOnlyRecordNamesTheFoldersAndRunsGit -count=1 -v; } > "$OUT/run-boundary-probe.txt" 2>&1
echo "boundary probe: exit $? (0 means the boundary test passed with the probe file present)"
rm internal/render/zz_boundary_probe.go
echo "checkout after: $(git status --short | wc -l) changed files"
{ echo "--- base $4"; sh "$TESTS/retie-probe.sh" "$4"; echo "--- candidate $5"; sh "$TESTS/retie-probe.sh" "$5"; } > "$OUT/run-retie-probe.txt" 2>&1
echo "retie script ran"
