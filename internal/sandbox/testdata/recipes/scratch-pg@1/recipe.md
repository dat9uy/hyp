---
name: scratch-pg
version: 1
format: 1
subject: Scratch PostgreSQL started inside a hyp sandbox on a Unix socket only (test data for stacked bundle recipes)
---

# scratch-pg

## Launch

Initialise a scratch cluster under `$HYP_SANDBOX_TMP/db` and start it with
`listen_addresses=''`, so it opens no TCP port, and its Unix socket in
`$HYP_SANDBOX_TMP`. The socket path must stay under 104 bytes, the macOS
limit, which a sandbox under `~/.hyp-sandbox` does.

Run: ./launch.sh

Worked when `recipe.log` shows `server started`.

## Doctor

Run: ./doctor.sh

Healthy when it prints `accepting connections` and `listen_addresses ''`.

## Drive

Connect with `psql -h $HYP_SANDBOX_TMP -U postgres -d postgres`; this bundle
drives nothing.

## Evidence

`recipe.log`, kept by `hyp sandbox remove`.

## Cleanup

Stops the server Launch started. If Postgres were left running, its working
directory (`$HYP_SANDBOX_TMP/db`) would make `hyp sandbox remove` refuse.

Run: ./cleanup.sh
