# Broker and outbox under a short receiver outage

## Setup

The load generator posted one order change per 100 ms, and each change queued one webhook to a local receiver that records every delivery id it answers.

The receiver answered 503 on about one call in six, to keep retries running through the event under test. The event was: receiver unavailable for 90 seconds at delivery 80.

## Run

No process was restarted in this run: only the receiver went away for 90 seconds, so it cannot speak to a restart on either path. Both paths retried and delivered everything.

The run used one process of each kind on a single 4-core host, with local disks, so timings in the log say nothing about capacity.

## Reading the result

Reconciliation at the end compared the ids the service had accepted with the ids the receiver had recorded. A delivery counts as lost when it was accepted and the receiver never recorded it.

Delivery twice is not counted as a failure, because the receiver removes repeats by idempotency key.
