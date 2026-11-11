# Q: where a build Goal's start commit lives (session e6a35a24, 2026-11-11)

A subject belongs to what is under test, from the Hypothesis down; a DesignQuestion has none today, and the Goal is never under test.

S1, the base moves. main moves from B to B' before the merge.
- From the runs: a new red-on-base run at B' records the new fact; the earlier Result stays a true fact about B. Holds.
- On the Goal: the stored B is now wrong, and no event corrects it but a new Goal. Breaks.

S2, a discovery Goal. Its candidates are stand-ins; no start commit means anything.
- On the Goal: the field is empty there and filled only by builds: a delivery kind by another door, which the owner rejected. Breaks.

Gap left: hyp cannot check a line's prose commit against the Result's base; the owner checks it by eye. A failure of that check on the first build Goal is the event that would add a field.
Owner's answer: "sure, just be aware that under shared stores between the tree, we should know where the Goal come from, anyway" (origin is derived from the Goal's cited keys).
