# In-process retry under a restart of the web process

## Setup

The load generator posted one order change per 100 ms, and each change queued one webhook to a local receiver that records every delivery id it answers.

The receiver answered 503 on about one call in six, to keep retries running through the event under test. The event was: restart of the web process at delivery 275.

## Run

Each delivery had its own retry timer in the web process, and each attempt was appended to the attempt log, which the process kept in memory as an in-process array and read back at start-up. The web process was restarted at delivery 275.

The run used one process of each kind on a single 4-core host, with local disks, so timings in the log say nothing about capacity.

## Reading the result

Reconciliation at the end compared the ids the service had accepted with the ids the receiver had recorded. A delivery counts as lost when it was accepted and the receiver never recorded it.

Delivery twice is not counted as a failure, because the receiver removes repeats by idempotency key.
