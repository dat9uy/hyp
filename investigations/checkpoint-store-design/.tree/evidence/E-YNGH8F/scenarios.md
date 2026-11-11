# Q: how many event types (session e6a35a24, 2026-11-11)

The owner's act has no node and no why; the runner's has the Goal as node and a required why. The envelope differs, not only the payload.
- Two types: each type has one envelope rule, as every type has today (spec 3, 4). Strict decoders refuse each other's fields, so "owner and runner fields never mix" is a schema fact; five "needs --role owner" usage errors in cmdCheckpoint go. Holds.
- One type: the node and why rules depend on which payload fields are present, which no type does today, and the mixing rule needs its own refusal. Breaks the per-type envelope.

The agent first recommended one type; kongming's evidence, checked against internal/event/event.go:306 and internal/replay/keyed.go, reversed it.
Owner's answer: "record them as selected, with this session as evidence".
