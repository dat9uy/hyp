# Replay names no guard for an event whose payload it cannot read

The spec of Goal G-H32WNG. It has no model, and the kept model in
`goals/model/` does not change: the change names no guard for refusals replay
already gives, and removes one code from `hyp spec`. No Goal moves between
states on another event than before, and the kept model has no step and no
rule for the guard a refusal names.

## Line 1: an undecodable payload is a form fault of every event kind

- Replay of a goal.add, example.place, examples.place, select, land, close or
  unselect whose payload does not decode refuses it with the form-fault code
  `payload-decode` and names no guard. `hyp status` lists it under
  `refused at replay:` as `ULID: KIND payload: CAUSE`.

## Line 2: hyp spec lists payload-decode only under (no guard)

- The code `payload-invalid` is gone. No guard lists a payload code.

## Line 3: a payload that decodes and breaks a guard is refused under it

- The guards are unchanged: a decoded event that breaks one is listed with
  that guard and its code.
