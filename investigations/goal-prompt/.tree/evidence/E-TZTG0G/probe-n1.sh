#!/bin/bash
# N1: a store outside every repository whose Goal body has a heading but no valid block (no "Read only:" line); line G-ZBM2RE#3 wants exit 1 and nothing on stdout, naming the missing heading.
H=$1; echo "hyp $(shasum -a 256 $H | cut -c1-16)"
U=$(mktemp -d); cd $U; $H init >/dev/null
G=$($H add goal --title G --body $'x\nMay change: a.\nNever: c.' --actor o --why w 2>/dev/null | awk '{print $1}')
so=$($H prompt $G --actor r 2>$U/err); c=$?; echo "N1: exit $c, stdout bytes ${#so}, stderr: $(cat $U/err)"
[ $c = 1 ] && [ -z "$so" ] && grep -q 'no "Read only:" line' $U/err && echo "N1: pass" || echo "N1: fail"
rm -rf $U
