#!/bin/sh
# cleanup.sh: stop what drive left running. Succeeds twice, and when launch never ran.
d=${HYP_SANDBOX_RECIPE_TMP:-}
if [ -n "$d" ] && [ -f "$d/pgid" ]; then
	pgid=$(cat "$d/pgid")
	case $pgid in
	'' | *[!0-9]*) ;;
	*)
		kill -TERM -- "-$pgid" 2>/dev/null || true
		i=0
		while kill -0 -- "-$pgid" 2>/dev/null && [ $i -lt 50 ]; do sleep 0.2; i=$((i + 1)); done
		kill -KILL -- "-$pgid" 2>/dev/null || true
		;;
	esac
fi
exit 0
