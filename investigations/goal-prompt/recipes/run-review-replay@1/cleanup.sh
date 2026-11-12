#!/usr/bin/env bash
# Cleanup: stop the runner's process group if it still runs; succeeds when
# Launch or Drive never ran, and when run twice.
RUN="${HYP_SANDBOX_RECIPE_TMP:-}/run.json"
[ -f "$RUN" ] || exit 0
pid=$(python3 -c 'import json,sys; print(json.load(open(sys.argv[1]))["pid"])' "$RUN" 2>/dev/null) || exit 0
kill -0 "$pid" 2>/dev/null || exit 0
kill -TERM -- "-$pid" 2>/dev/null
for _ in $(seq 1 60); do kill -0 "$pid" 2>/dev/null || exit 0; sleep 1; done
kill -KILL -- "-$pid" 2>/dev/null
sleep 2
exit 0
