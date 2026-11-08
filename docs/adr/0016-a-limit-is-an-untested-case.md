---
status: accepted
---

# A limit is an untested case

A Result's `limits` field carried several jobs. The 34 trees kept on 2026-11-08 hold 173 limits. About two in five name a case nobody ran. The rest state the scope of a run, a procedure note, a provenance claim, a confound, or a premise. hyp gives a limit four transitions: a Result `closes` it, an Experiment `aims` at it, a runner `refers` it, and the owner `accepts` it. All four are about a run that has not happened yet, so only the untested case fits all of them.

The ledger (spec §8.1) owed a limit only when its Result judged a candidate. A limit found on a rival's Result was owed by nobody once that rival fell, even when its case applied to the candidate. ADR 0015 named this as loss 2 and mitigated it with a select warning.

Owner decision on 2026-11-08: a limit is a case the Goal covers that no check ran. A rival's limit is answered by a check that runs the case on every candidate. The rivals that survive are what the verdicts leave. The multi-tenant tree did this once already: R-FAMJDB named two cases, E-236WKJ aimed at both, and R-S7MX6V closed them while contradicting both of its candidates.

## Decision

- **A limit is a case the Goal covers that no check ran on the Hypotheses its Result judged.** Its one job is to name the next case. The field name, the key and the four transitions stay. A run's other notes have homes of their own. The scope of a run goes in the Experiment body. A procedure note or a commit goes in the observation, with the commit in a cited file. A premise is a `depends_on` edge. The glossary's Avoid list gains scope, assumption and provenance.
- **An open limit is owed while its question has a candidate, whichever Hypotheses its Result judged.** The runner owes an unreferred one, and the owner owes a referred one. The limit therefore holds the select, and `hyp select` refuses with `Q owes the runner first: limit R-…#1`. This computes ADR 0015's loss 2. The select warning `N open limits judged other Hypotheses` is gone, because an admitted select can no longer have such a limit unreferred.
- **`refers` also covers a case that is void for every candidate.** A runner refers a limit that no test within the Goal can close, or whose case does not arise for any candidate. It is still one transition: the runner hands the limit to the owner.

## Consequences

- Old trees may owe more. On a scratch build over the 34 trees, limit items went from 79 to 95. Five trees changed, the live multi-tenant tree among them (2 to 6), and the other 29, both public trees among them, did not. A complete Goal can owe limits that were left on a dead round. Owing an item never changes a status.
- A limit that is not a case, such as a scope statement, now holds the select until it is referred. This is the cost of one meaning: the runner says once why no test closes it, and the owner accepts it.
- Not decided: a candidate added after a limit was closed was never run through that case. A rule that owes such a limit could be derived (§8 allows several closers). It waits for a real event (AGENTS.md, step 1).

## Considered Options

- **A `kind` on each limit.** Rejected: one payload field means one item kind and one transition (ADR 0013), and a kind field would split the transitions without making them computable.
- **Keep the select warning.** Rejected: it warned of work the ledger can owe.
