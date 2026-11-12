#!/bin/bash
# M2: a CRLF body with no blank line in its block; line G-ZBM2RE#1 wants exit 0 and each block line byte for byte, so each line keeps its \r.
# M3: a store outside every repository whose Goal body has no constraints block; line G-ZBM2RE#3 wants exit 1 and nothing on stdout.
H=$1; echo "hyp $(shasum -a 256 $H | cut -c1-16)"
T=$(mktemp -d); cd $T; python3 -c "import os; os.mkdir('.'+'git')"; $H init >/dev/null
G=$($H add goal --title G --body $'Builds.\r\nMay change: a.\r\nRead only: b.\r\nNever: c.\r\n' --actor o --why w 2>/dev/null | awk '{print $1}')
out=$($H prompt $G --actor r 2>/dev/null; echo "exit $?")
n=$(printf '%s' "$out" | grep -c $'\r$'); echo "M2: $(printf '%s' "$out" | tail -1); block lines ending in \\r: $n (want 3)"
printf '%s' "$out" | grep -q 'May change: a.'$'\r' && [ "$n" = 3 ] && echo "M2: pass" || echo "M2: fail"
U=$(mktemp -d); cd $U; $H init >/dev/null
G2=$($H add goal --title G --body 'Builds, with no block.' --actor o --why w 2>/dev/null | awk '{print $1}')
so=$($H prompt $G2 --actor r 2>$U/err); c=$?; echo "M3: exit $c, stdout bytes ${#so}, stderr: $(cat $U/err)"
[ $c = 1 ] && [ -z "$so" ] && echo "M3: pass" || echo "M3: fail"
rm -rf $T $U
