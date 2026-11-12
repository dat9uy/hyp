#!/bin/bash
# Q1: a store outside every repository whose Goal body has all three headings and a block line of text that starts with none of them ("The end."); line G-ZBM2RE#3 wants exit 1 and nothing on stdout, naming that line.
H=$1; echo "hyp $(shasum -a 256 $H | cut -c1-16)"
U=$(mktemp -d); cd $U; $H init >/dev/null
G=$($H add goal --title G --body $'Builds.\nMay change: a.\nRead only: b.\nNever: c.\nThe end.' --actor o --why w 2>/dev/null | awk '{print $1}')
so=$($H prompt $G --actor r 2>$U/err); c=$?; echo "Q1: exit $c, stdout bytes ${#so}, stderr: $(cat $U/err)"
[ $c = 1 ] && [ -z "$so" ] && grep -q 'starts with no heading: "The end."' $U/err && echo "Q1: pass" || echo "Q1: fail"
rm -rf $U
