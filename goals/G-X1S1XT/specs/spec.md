# hyp spec says checked at replay exactly when replay refuses under the guard

The spec of Goal G-X1S1XT. It has no model, and the kept model in
`goals/model/` does not change: the change names a different guard, or no
guard, for refusals replay already gives, and changes what `hyp spec` prints.
No Goal moves between states on another event than before, and the kept model
has no step and no rule for the guard a refusal names.

## Line 1: Whole set and Kind are checked at replay

- Replay of a goal.add refuses an example it carries that is on no line of
  the Goal (Whole set, `example-off-line`) and an example whose recorded
  verdict on H0 is not the one its line's kind needs (Kind,
  `added-verdict-wrong`).
- `hyp spec` prints `checked at replay: yes` for both, each with the part
  replay checks; the act still reads the example folders and runs the Probes.

## Line 2: Stable lists no payload-invalid

- A result.record whose payload does not decode is a form fault: replay
  refuses it with the code `payload-invalid` and names no guard.
- Stable's codes are `other-verdict` and `verdicts-differed`.

## Line 3: a select with a role no select names is a form fault

- Replay of a select whose role is neither owner nor runner refuses it with
  the new form-fault code `select-role`, which names no guard. `hyp status`
  lists it under `refused at replay:` with the role, not under Candidate.
- The select act is unchanged: it still refuses a role other than owner with
  Candidate, `role-cannot-select`.

## Line 4: Stable and Candidate stay checked by their acts only

- Replay checks no part of Stable or Candidate, so `hyp spec` keeps
  `checked at replay: no` for both.

## Line 5: a result.record that does not decode is still refused

- The refusal of line 2 is still listed under `refused at replay:`.
