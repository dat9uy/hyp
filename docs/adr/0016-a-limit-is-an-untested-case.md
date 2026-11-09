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
- **A close covers every case the limit names, and one limit names one case (2026-11-09).** A Result that closes a limit and did not run every case its text names records each case it did not run as a `--limit` of its own. A runner refers that new key when its case is void for every candidate or outside the Goal's lines. A claim that a case is void, out of scope or covered elsewhere is such a limit plus a refer, never a sentence in a body or an observation, because a close leaves the owner nothing to answer. A new limit names one case: "not X or Y" is two limits. `hyp record --closes` prints the whole text of each limit it closes, so the runner rereads every case it names at the moment it claims them.

  Owner decision on 2026-11-09, from run 3 of the multi-tenant tree. Two of its ten closes ended a two-case limit after one case ran. `R-BZ7C5H` closed `R-XFSQEE#1` (Postgres and `.xlsx`) after Postgres alone, and called the `.xlsx` case void only in its observation. `R-M4ABBF` closed `R-AMG1CR#2` (a crash mid-revision and a deploy that misses one) after the crash alone, and named the deploy case nowhere. The same runner filed correctly twice in the same run: it referred `R-GVPBZZ#2`, and recorded K2's untested remainder as `R-M4ABBF#1` and referred it. No new field, verb or kind is added (ADR 0013); the existing `--limit` and `--refers` carry it.

## Consequences

- Old trees may owe more. On a scratch build over the 34 trees, limit items went from 79 to 95. Five trees changed, the live multi-tenant tree among them (2 to 6), and the other 29, both public trees among them, did not. A complete Goal can owe limits that were left on a dead round. Owing an item never changes a status.
- A limit that is not a case, such as a scope statement, now holds the select until it is referred. This is the cost of one meaning: the runner says once why no test closes it, and the owner accepts it.
- Not decided: a candidate added after a limit was closed was never run through that case. A rule that owes such a limit could be derived (§8 allows several closers). It waits for a real event (AGENTS.md, step 1).

## Considered Options

- **A `kind` on each limit.** Rejected: one payload field means one item kind and one transition (ADR 0013), and a kind field would split the transitions without making them computable.
- **Keep the select warning.** Rejected: it warned of work the ledger can owe.

## Amended 2026-11-09: a limit's category is derived, and the owner can return a referred limit

The owner asked whether a limit belongs to the Result, which is provable because the run already happened, or to the Hypothesis or the Goal, which cannot be known in advance, and how a limit is promoted up the tree. A new Goal is a new tree and a new Hypothesis is a new branch, and turning a Result's limit into a rethink of the design or of the Goal is how thinking works, so it must be a visible, cheap path made of ordinary events.

The run that prompted this: `R-P424P1#1` (the missed-deploy case) was referred. `E-ZVN0EM` aimed at it, and its Result `R-5JYRA0` ran the case and contradicted h-module on D2 and D3. But `R-5JYRA0` did not close the limit, following the Experiment's own rule ("Closes R-P424P1#1 if D1-D3 hold"), so `hyp status` still asked the owner to accept or refuse a case that had been run.

Owner decision on 2026-11-09:

- **A limit has one stored scope, its Result's.** At record time, the one provable fact is "this run did not cover case C". No category, kind or scope is stored, because choosing one at record time would ask the runner to judge what it cannot yet know.
- **Its category is derived by rules over typed fields, never from prose.** A Result's `closes` and its `verdicts` say whether the case ran and what it did to each Hypothesis judged: a contradicted Hypothesis means the design failed the case, and a supported one means it passed. An owner `accepts` means this tree does not run the case. The latest `refers` or `returns` says whose turn an open limit is. Summaries and observations explain the reason, and no rule reads them.
- **A close records that the case ran, whatever the verdicts.** A Result that runs a limit's case closes it whether its verdicts support or contradict. Closing only on a pass is the misreading that left `R-P424P1#1` open. Displays show the closer's verdicts next to `closed by`, so the promotion into the design is visible.
- **The owner can return a referred limit.** "The Goal covers this case, run it" is the owner's answer to a refer, and it is a typed transition: `hyp checkpoint --role owner --returns <R-id>#<n>`, the mirror of `--refers`. A returned limit is an ordinary open limit again, owed to the runner while its question has a candidate. Whose turn an open limit is follows the latest `refers` or `returns` in merge order. A RULE CHANGE checkpoint that widens the Goal stays allowed, and names the keys it brings into the Goal with `--returns`.
- **A new Goal is a new tree.** The new tree cites the old tree's keys in text and with an `external_ref` edge. No typed cross-tree cite is added, because nothing would be computed from it.
- **A candidate that no closer judged is owed the case again.** This settles the "Not decided" item above in principle: once a closed limit's question has a supported candidate that no closing Result judged, the limit is owed to the runner for that candidate. It is built when the first successor of a Hypothesis that a closer contradicted is added, because that is its real event.

A return also reopens an accepted limit (2026-11-09, owner: "fix gap 11"). An accept was final, and a `--returns` on an accepted limit was admitted and changed nothing. Now the latest of accept and return stands, so an owner who changes their mind writes one checkpoint, and a runner's refer of an accepted limit is refused, since only the owner reopens one. An accept of a closed limit is refused too (gap 12): the case ran, so the accept would change nothing.

No kind field, no edge to a limit key, no `--answers` on `add hyp`, and no change of turn on `--aims` are added (ADR 0013). `aims` names intent and settles nothing, so it cannot carry the owner's answer.

Displays and checks that follow, in order:
- a limit's line shows its closers' verdicts;
- `hyp record` warns when the Experiment aims at a limit that the Result neither closes nor replaces with a new `--limit`, and prints the admit warnings it drops today;
- `hyp show <R-id>#<n>` prints the limit, the Hypotheses judged and their status now, whose turn it is with the refer and return summaries, the aimers, the closers with their verdicts, and the acceptor;
- `--returns`;
- the re-owe rule, when a successor is added.

Kongming's report `plans/reports/kongming-261109-1529-typed-limit-promotion.md` holds the options weighed.
