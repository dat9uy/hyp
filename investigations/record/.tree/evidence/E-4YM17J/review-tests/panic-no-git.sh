#!/bin/sh
# Run a hyp binary with no git on PATH in a temp repository whose reftable
# stack holds one malformed table: a symref record whose target-length varint
# wraps past int (2^64-1000). replay-spec: a reftable stack hyp cannot read
# counts as holding the record, so the command exits 3.
set -u
HYP=$1
T=$(mktemp -d)
mkdir -p "$T/.git/reftable"
printf 'a.ref\n' > "$T/.git/reftable/tables.list"
printf '\122\105\106\124\001\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\000\162\000\000\057\000\013\141\000\200\376\376\376\376\376\376\376\367\030\000\000\030\000\001' > "$T/.git/reftable/a.ref"
EMPTY=$(mktemp -d)
cd "$T" || exit 1
PATH=$EMPTY "$HYP" status > "$EMPTY.out" 2>&1
echo "exit: $?"
head -6 "$EMPTY.out"
cd / && rm -rf "$T" "$EMPTY" "$EMPTY.out"
