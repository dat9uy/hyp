#!/bin/sh
set -eu
pg_isready -h "$HYP_SANDBOX_TMP"
la=$(psql -h "$HYP_SANDBOX_TMP" -U postgres -d postgres -tAc 'SHOW listen_addresses')
echo "listen_addresses '$la'"
[ -z "$la" ]
