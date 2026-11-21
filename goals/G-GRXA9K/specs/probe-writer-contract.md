# The Probe writer's contract

The spec of Goal G-GRXA9K. It has no model: the change adds one output to one
read and moves no Goal between states. The Probe writer writes no event, so
no act and no guard changes.

## Line 1: `hyp status G` prints the Probe writer's contract

- While a Goal is `written` or `open`, `hyp status G`, with no role or with
  `--role probe-writer`, prints in the probe-writer block, instead of
  `(none)`:
  - the Probe writer's contract, for the Probe of one example, in five
    numbered clauses: read the line, its example and H0; write and commit only
    that example's probe folder, `goals/G/<line>/<example>/probe`; start from
    the generic Probe and write new code only when it cannot run the example;
    run the example on H0 with `hyp try` and return its output; write no event.
  - the generic Probe, under `the generic Probe, probe/run:`. It is
    `skills/hyp-grill/probe/run` byte for byte, built into the binary with
    `go:embed` (package `hyp` at the module root), so the binary and the skill
    never print two Probes.
  - one task for each line of the Goal, in its order:
    `<line> (<kind>): the Probe of each example of <line>; check it with: hyp
    try G <line>/<example>`.
- Every line of the block is indented, so the block ends where the runner's
  block starts. The block holds no `hyp add`, `place`, `run`, `select`, `land`
  or `unselect` line.

## Line 2: every other output stays as on H0

- `hyp status G` with `--role owner`, `grill` or `runner` prints no
  probe-writer block, as before.
- `hyp status` for every Goal prints `(none)` in the probe-writer block, as
  before. So do the return of every act and the return of `try`.
- A `selected` or `landed` Goal prints `(none)` in the probe-writer block of
  `hyp status G`, and its record as before.
