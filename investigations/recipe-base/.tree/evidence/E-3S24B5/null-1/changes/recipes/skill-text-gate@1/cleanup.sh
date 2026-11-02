#!/usr/bin/env bash
set -u
f="${HYP_SANDBOX_RECIPE_TMP:-}/runner.pgid"
[ -f "$f" ] || exit 0
pg=$(cat "$f")
if [[ $pg =~ ^[0-9]+$ ]]; then
	kill -KILL -- "-$pg" 2>/dev/null
	for _ in $(seq 1 100); do
		kill -0 -- "-$pg" 2>/dev/null || break
		sleep 0.1
	done
fi
rm -f "$f"
exit 0
