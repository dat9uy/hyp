#!/bin/sh
set -eu
db="$HYP_SANDBOX_TMP/db"
initdb -D "$db" -U postgres --auth=trust -E UTF8 >/dev/null
pg_ctl -D "$db" -l "$HYP_SANDBOX_TMP/pg.log" -w \
  -o "-c listen_addresses='' -k $HYP_SANDBOX_TMP" start
