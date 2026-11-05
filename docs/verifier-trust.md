# What a verifier's confirm proves

The threat is a lazy or skimming agent, not a malicious one.

| Claim | Provable per run? | Today |
|---|---|---|
| Different actor from the selector | Yes, by name (spec §5 rule 9) | Provided, weakly: hyp does not authenticate actors |
| Separate session | No | A brief requirement in `skills/hyp-prototype/VERIFIER.md`; not proved |
| Exact set to judge | Yes (Dossier, spec §7.4) | Provided |
| Evidence was touched | Not today | Missing: #26 measured no verifier opening the target's evidence in 15 trials |
| Judgement is right | Never | Statistical only, by replay with known wrong verdicts |
| Unchecked work blocks `answered` | Yes (spec §7.4) | Provided for lapsed confirms only |
| Later re-check | Partly | `.tree/evidence` files are not hashed |

#26 tracks whether a recorded check closes the evidence-was-touched gap.
