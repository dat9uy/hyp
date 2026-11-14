#!/bin/sh
# Clone this repository into a temp dir, import every store with the
# candidate's hyp, and compare hyp status and hyp check before and after.
set -u
HYP=$1; SRC=$2; W=$(mktemp -d)
export GIT_CONFIG_GLOBAL=$W/gitconfig GIT_CONFIG_NOSYSTEM=1 HYP_ACTOR=review
printf '[user]\n\tname = t\n\temail = t@t\n' > $GIT_CONFIG_GLOBAL
git clone -q "$SRC" $W/c
cd $W/c
stores=$(ls -d investigations/*/.tree | sed 's#/.tree##')
for s in $stores; do (cd $s && $HYP status > $W/before.$(basename $s).status 2>&1; $HYP check > $W/before.$(basename $s).check 2>&1); done
$HYP import $stores; echo "import exit $?"
$HYP import $stores; echo "second import exit $?"
for s in $stores; do b=$(basename $s); (cd $s && $HYP status > $W/after.$b.status 2>&1; $HYP check > $W/after.$b.check 2>&1)
  for k in status check; do if cmp -s $W/before.$b.$k $W/after.$b.$k; then echo "$b $k same"; else echo "$b $k DIFFERS"; diff $W/before.$b.$k $W/after.$b.$k | head -20; fi; done
done
git status --short | head
echo W=$W
