#!/bin/sh
# Re-tie a Hypothesis subject to an edited patch at the same path, with the
# base hyp and with the candidate hyp, each in a new temporary git repository.
# usage: retie-probe.sh HYP_BINARY
set -u
HYP=$1
T=$(mktemp -d)
export GIT_CONFIG_GLOBAL=$T/gitconfig GIT_CONFIG_NOSYSTEM=1 HYP_ACTOR=a
printf '[user]\n\tname = t\n\temail = t@t\n' > $GIT_CONFIG_GLOBAL
cd $T && git init -q repo && cd repo && mkdir s && echo x > s/README.md && git add -A && git commit -qm base
cd s && $HYP init >/dev/null
g=$($HYP add goal --title G --why w | cut -d' ' -f1)
q=$($HYP add dq --parent $g --title Q --why w | cut -d' ' -f1)
h=$($HYP add hyp --parent $q --title H --why w | cut -d' ' -f1)
mkdir arms && echo first > arms/a.patch
$HYP subject $h --base HEAD --patch arms/a.patch --why w >/dev/null; echo "first tie: exit $?"
echo second > arms/a.patch
$HYP subject $h --base HEAD --patch arms/a.patch --why w; echo "re-tie after edit: exit $?"
git -C .. for-each-ref refs/hyp
rm -rf $T
