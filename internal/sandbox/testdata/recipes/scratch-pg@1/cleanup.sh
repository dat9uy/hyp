#!/bin/sh
db="$HYP_SANDBOX_TMP/db"
[ -f "$db/postmaster.pid" ] || exit 0
pg_ctl -D "$db" stop -m fast
