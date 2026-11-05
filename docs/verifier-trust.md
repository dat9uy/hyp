# What a verifier's confirm proves

The threat is a lazy or skimming agent, not a malicious one.

| Claim | Provable per run? | Today |
|---|---|---|
| Different actor from the selector | Yes, by name (spec §5 rule 9) | Provided, weakly: hyp does not authenticate actors |
| Separate session | No | A brief requirement in `skills/hyp-prototype/VERIFIER.md`; not proved |
| Exact set to judge | Yes (Dossier, spec §7.4) | Provided |
| Evidence was touched | Not today | Missing: #26 measured no verifier opening the target's evidence in 15 trials. A recorded audit with a checked quote did not make verifiers catch wrong verdicts: 1 of 5, with `holds` on a known wrong verdict in 4 of 5 (`R-WBHRJ4`). On a synthetic store, a written comparison caught 4 of 5 and a fresh reader per Result 5 of 5, against 2 of 5 for today's brief (`R-GT9E0K`). At the real cut the comparison failed 5 of 5, with no verifier opening either Result's evidence (`R-D023Q4`), so it did not ship. A fresh reader per Result then failed 4 of 5 there (`R-GDBXXE`): every trial's reader opened the files, its report read `holds` in 4 and `unclear` in 1, and no verifier opened the evidence itself, so it did not ship either. A reader changed who opens the files, not the judgement; the `session` variant has not been run at a real cut. Each of these is a replay, and none is proved per run |
| Judgement is right | Never | Statistical only, by replay with known wrong verdicts |
| Unchecked work blocks `answered` | Yes (spec §7.4) | Provided for lapsed confirms only |
| Later re-check | Partly | `.tree/evidence` files are not hashed |

#63 tracks how to close the evidence-was-touched gap. A recorded check was tested and abandoned, so the fix moves to agent procedure.
