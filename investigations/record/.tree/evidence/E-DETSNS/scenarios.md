# Scenarios for Q-9WV3D7 (grill of #181, 2026-11-14)

Settled by argument in session with the owner (owner-dat), round 2 question 6. The second reader's finding that the issue's own scenario (a stopped Goal whose code branch is deleted, an owner answer from another worktree) does not separate the rivals, since neither keeps the record on the code branch, led to this scenario.

## Scenario: a crash between the file line and its commit

1. A run appends an event (a Result) through hyp.
2. h-record-worktree: hyp writes the line into the record worktree's events file, then would run git commit. The process is killed between the two steps.
   - A reader of the record worktree now sees the Result.
   - `git clean -fdx`, `git checkout -- .` or removing that worktree loses the line; nothing in git holds it.
   - So a reader saw an event that is later lost: it fails the line "no command reads part of a write".
3. h-ref: hyp writes the event's blob, tree and commit objects, then runs `git update-ref refs/hyp/record NEW OLD`. The process is killed between the two steps.
   - The ref did not move, so no reader sees the event; the objects are unreferenced and git gc may prune them.
   - Once update-ref lands, the event is reachable from a ref that no worktree act touches.
   - So a reader only ever sees a durable event: it meets that line.

## Variant

If h-record-worktree reads only committed state, readers no longer see a half-written event, but then the worktree's files serve no reader: the rival becomes h-ref plus a checkout that adds nothing, and a branch on GitHub that one "Delete branch" click removes.

verdict: h-ref=supports
verdict: h-record-worktree=contradicts
