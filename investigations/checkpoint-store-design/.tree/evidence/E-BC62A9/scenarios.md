# Q: what the owner's act is attached to (session e6a35a24, 2026-11-11)

S: Goal A has a referred limit; Goal B has an open reading. The owner sits once, accepts A's refer, answers B's reading.
- No node: one event, its keys lead to their Goals (R-... through its Result's parents, G-...#n directly). Replay files each move under its Goal. Holds.
- One Goal per act: two events for one sitting, and the "one log of owner checkpoints" is split by Goal again. Breaks the owner's rule.

Fact checked in the session: no event in this repo's stores had an empty node, so the owner act is the first event that acts on the store.
Owner's words: "if we try to split the event, we need a lot of mechanism to ref two Goal at the same time while a little prose in owner checkpoint works fine."
Lines moved into the Goal's node.add later in the session, so the owner act never needs a node.
