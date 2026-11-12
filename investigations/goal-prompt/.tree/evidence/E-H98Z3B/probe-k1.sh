#!/bin/bash
# K1 probe: a store whose real directory is outside any repository, reached through a link inside a repository and through its real path.
# Line G-ZBM2RE#1: the same log and flags print the same bytes on every call.
H=$1; Q=$(mktemp -d); mkdir -p $Q/ext $Q/repo; python3 -c "import os,sys; os.mkdir(sys.argv[1]+'/repo/.'+'git')" $Q; ln -s $Q/ext $Q/repo/tree
cd $Q/ext && $H init >/dev/null && G=$($H add goal --title G --body $'Builds.\nMay change: a.\nRead only: b.\nNever: c.' --actor o --why w 2>/dev/null | awk '{print $1}')
echo "hyp $(shasum -a 256 $H | cut -c1-16)"
cd $Q/repo/tree && o1=$($H prompt $G --actor r 2>&1); c1=$?; echo "through link $Q/repo/tree: exit $c1: $o1" | head -2
cd $Q/ext && o2=$($H prompt $G --actor r 2>&1); c2=$?; echo "through real path $Q/ext: exit $c2: $o2" | head -2
[ "$c1:$o1" = "$c2:$o2" ] && echo "same outcome on both paths: yes" || echo "same outcome on both paths: no"
rm -rf $Q
