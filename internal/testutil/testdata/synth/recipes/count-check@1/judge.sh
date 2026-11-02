#!/bin/sh
# judge.sh COUNT: print the count and the verdict for the alias "short".
case "$1" in
'' | *[!0-9]*)
	echo "judge.sh: want one whole number, got '$1'" >&2
	exit 2
	;;
esac
echo "count $1"
if [ "$1" -le 2 ]; then
	echo "verdict: short=supports"
else
	echo "verdict: short=contradicts"
fi
