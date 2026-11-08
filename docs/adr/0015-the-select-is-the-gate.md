---
status: accepted
---

# The select is the gate

Since ADR 0011 a question read `answered` only once a verifier typed `sufficient` on its select, and the owner confirmed too on a trade-off. Since 2026-11-08 the ledger (spec §8.1) derives who owes what, and every owner word has one field and one transition (ADR 0013). So the verifier competed with a computed rule, not with nothing, and the question became what a verifier confirm still said that the ledger did not. At the first select of the multi-tenant tree (2026-11-08) the answer was: almost nothing, and what it did say was not computed. Owner decision on 2026-11-08: the owner accepted this design with "choose only", so the owner's choice, recorded as the select, ends a trade-off, and the owner confirm survives only for `--scope-change`.

This supersedes the verifier half of ADR 0011. The Dossier, its lapse rule for the owner confirm, and the Decision record stay.

## Decision

- **`hyp select` is admitted only when the ledger owes it.** `Projection.AdmitSelect` admits a select of H on Q when the ledger owes a `select` of H, or the owner's `choose` on Q and H is a candidate. Otherwise it refuses at exit 1, naming what is owed: the first runner item on Q, an open reading that waits for the owner, the premise hold, the lone candidate's open sub-question, or Q's state. Replay still applies a select in a merged log, so the rule is write-time (spec §5.1). To change a choice, `hyp unselect`, then select.
- **`answered` is derived from the tree and one stored fact.** The conflict set is empty; the current select names a live child Hypothesis that is `supported` and not stale; the DQ is not stale; the selected Hypothesis has no open sub-question (ADR 0014, unchanged); every candidate now was a candidate at the select; and, only when the select is a scope change, the owner confirm that counts is `sufficient`. Replay stores the candidates when a select applies, as it stores the live rivals at an Experiment's add, so old logs get it too. A new candidate reopens the question with the reason `new_candidate`, after `awaiting_subquestions`; a candidate that falls does not, because the chosen one still stands. The reasons `awaiting_verifier` and `verifier_insufficient` go, and `awaiting_owner` is for a scope change only.
- **The verifier role is retired.** `hyp confirm --role verifier` is refused at exit 1: "the verifier role was retired; a review that finds a defect records a Result, a retract or a limit". An old verifier confirm replays as history and applies nothing, as `dq.request_verify` does. The separation-of-duties half of spec rule 9, `Decision.Chain` and the verifier-loop signal go. Any actor may still review; an objection is a Result with a `contradicts` or a limit, or a retract. No optional review role is kept, because nothing would compute from it.
- **The ledger grows by what the verifier used to catch, where it can be computed.** `undecided` (runner): a live `inconclusive` Hypothesis with no planned Experiment, so an untested rival in all but name holds the select. `premise` (runner): a candidate's `depends_on` target sits under a question whose current select names another Hypothesis. The **premise hold**: neither `select` nor `choose` is owed while a candidate's premise sits under an open or reopened question, so premise questions are settled first; `hyp check` warns of a premise cycle between two questions. A `referred` limit is owed by the owner whichever Hypothesis its Result judged. `choose` is owed again when a candidate is new since the select. The `limit`, `reading`, `referred` and `premise` items are computed on answered questions too, so evidence that lands after the answer is owed and not only listed. Owing an item still never changes a status: the gate is at the write.
- **An admitted select warns of rivals' limits.** `warn: N open limits judged other Hypotheses: R-…#1, …` lists the open limits under the question whose Result did not judge the selected Hypothesis, since a limit on a rival's Result may apply to the chosen one too. The runner refers one that applies.

## What was tried

The verifier's confirm was meant to prove that a separate reader judged the evidence a select rests on. What it proved, as recorded until this decision:

| Claim | Provable per run? | What it was |
|---|---|---|
| Different actor from the selector | Yes, by name (spec rule 9) | Provided, weakly: hyp does not authenticate actors |
| Separate session | No | A brief requirement in `skills/hyp-prototype/VERIFIER.md`; not proved |
| Exact set to judge | Yes (the Dossier, ADR 0011) | Provided |
| Evidence was touched | No | Missing: #26 measured no verifier opening the target's evidence in 15 trials (`recipes/verifier-brief-arms@1`). A recorded audit with a checked quote did not make verifiers catch wrong verdicts: 1 of 5, with `holds` on a known wrong verdict in 4 of 5 (`R-WBHRJ4`, `recipes/verifier-audit-arms@1`) |
| Judgement is right | Never | Statistical only, by replay with known wrong verdicts |
| Unchecked work blocks `answered` | Yes | Provided for lapsed confirms only |
| Later re-check | Partly | `.tree/evidence` files are not hashed |

A census of 42 verifier confirms found 9 `insufficient`. Four were citation slips (an id or commit not in the cited files), two an untested live rival, one a misread select kind, one a coverage gap the runner already knew and wrote in the select's `why`, and one an undeclared coverage gap. None caught a wrong verdict, and one wrong verdict (`R-KA7Y6Z`) passed. Three replay programmes could not make a verifier read evidence. The first five kinds are now computed or owed: a cite of an id that is no node is refused at write time (spec §5.1), an untested or undecided rival is a runner item, a trade-off is derived, and a gap known at select time is a limit, which the ledger owes.

## Losses

Two things the verifier did are not computed, and this decision names them:

1. **An undeclared coverage gap.** Once in 42 confirms a verifier saw coverage nobody had declared, and it led to a new Result. Under this decision only the owner reading `hyp show` at `choose`, or a later Result, recovers it.
2. **A limit on a rival's Result that also applies to the candidate.** The ledger keys a `limit` item on the candidates its Result judged, so a limit written on a dead rival's Result is owed by nobody, even when the candidate was measured by the same instrument. `hyp select` warns of every such limit, and `refers` on one creates the owner's item; that is a mitigation, not a computation.

## Considered Options

- **Keep the verifier beside the ledger.** Rejected: a second gate whose item, `confirm`, the ledger can compute nothing about, at a session per select for one novel catch in 42 confirms.
- **An owner confirm on every select.** Rejected: the owner acts only on real choices (owner, 2026-11-06), and the ledger already routes every real choice to them. It is the fallback if a loss shows up (below).
- **Status reads limits.** Rejected: it re-admits typed text as a status input, and refusing the select gives the same protection at the write.
- **Keep the owner confirm on a trade-off.** Rejected by the owner's "choose only": after the choice it would be the owner typing `sufficient` on their own choice, as unauthenticated as the choice. This overturns the 2026-11-05 "choose, then confirm" wording, which was written with a verifier between the two acts.

## Consequences

- Both public trees stay `complete`: each has one trade-off select whose candidate set did not change afterwards (`investigations/id-citation` Q-DVZZHY, `investigations/recipe-base` Q-WXZRA3). Their RECENT lines keep the old confirm events. Before the change merges, the census spike runs again over the private trees for final-state flips; the 2026-11-05 counsel found no real `new_candidate` case, so flips are expected only where a verifier never confirmed. Old trees may change, as the owner has accepted.
- The synthetic store under `internal/testutil/testdata/synth` keeps its verifier confirms and `dq.request_verify` events as history. Its `generate.sh` no longer runs to the end on a current hyp.
- `recipes/verifier-brief-arms@1` and `recipes/verifier-audit-arms@1` stay frozen. They measured a role this decision removes.
- `skills/hyp-prototype/VERIFIER.md` and the Select and Confirm sections of `skills/hyp-prototype/SKILL.md` change as a correction to this decision, in a separate change.
- Record-time warnings for a verdict line that disagrees with the typed verdict, and for an id or hex token no cited file holds, are a separate change.

## Success metrics

- A question answered under this decision is later reopened by a retract or a new Result for a defect a Dossier reader would have seen and no limit declared. Two such in the first ten answered questions add the owner confirm on every select.
- `--limit` per Result does not fall, and `refers` appears in real trees.
- Owner acts per question stay at the real-choice count (`choose`, `accepts`, readings), not an extra confirm per select.
