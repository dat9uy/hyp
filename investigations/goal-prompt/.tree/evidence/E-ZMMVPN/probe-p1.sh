#!/bin/bash
# P1: a store outside every repository whose Goal body has all three headings and a block line that starts with none of them (a line of two spaces); line G-ZBM2RE#3 wants exit 1 and nothing on stdout, naming the bad line.
H=$1; echo "hyp $(shasum -a 256 $H | cut -c1-16)"
U=$(mktemp -d); cd $U; $H init >/dev/null
G=$($H add goal --title G --body $'Builds.\nMay change: a.\nRead only: b.\nNever: c.\n  ' --actor o --why w 2>/dev/null | awk '{print $1}')
so=$($H prompt $G --actor r 2>$U/err); c=$?; echo "P1: exit $c, stdout bytes ${#so}, stderr: $(cat $U/err)"
[ $c = 1 ] && [ -z "$so" ] && grep -q 'starts with no heading' $U/err && echo "P1: pass" || echo "P1: fail"
rm -rf $U
