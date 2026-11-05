# Redis queue survives a graceful restart

## Setup

The load generator posted one order change per 100 ms, and each change queued one webhook to a local receiver that records every delivery id it answers.

The receiver answered 503 on about one call in six, to keep retries running through the event under test. The event was: graceful restart of redis-server at delivery 210.

## Run

Redis ran from `redis.conf` with `appendonly yes` and `appendfsync everysec`, and was restarted with SHUTDOWN at delivery 210, after which the AOF was replayed. The outbox path ran alongside on only 20 deliveries, which is too few to judge it.

The run used one process of each kind on a single 4-core host, with local disks, so timings in the log say nothing about capacity.

## Reading the result

Reconciliation at the end compared the ids the service had accepted with the ids the receiver had recorded. A delivery counts as lost when it was accepted and the receiver never recorded it.

Delivery twice is not counted as a failure, because the receiver removes repeats by idempotency key.
