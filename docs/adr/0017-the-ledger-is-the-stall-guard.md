---
status: accepted; amended by ADR 0018: unjudged is a refusal, not a mark
---

# The ledger is the stall guard

`hyp signals` had two halves. The design half (`one-rival`, `D1`, `unjudged`) marks how an Experiment was set up. The progress half (`wasted`, `pending`, `no-movement`, `spin`) judged a round by whether a Result changed a Hypothesis's status. On 2026-10-24 the owner defined wasted as a round that moved no live rival's status. The wasted-definition tree of 2026-10-27 tried to calibrate that label against the owner's sense of a stall, and its supported rival was never selected. The owner's reasons in that tree were structural: a test of a fix is a new Hypothesis, and an untested part should be named.

Each of those reasons has since become a typed ledger item (spec §8.1): `untested`, `limit`, `reading` and `undecided`, each settled by an explicit event. ADR 0016 makes the intended loop "run every candidate through the next case", and statuses move only forward, so a candidate that holds every case reads `supported→supported` at every round. The progress signals read only that status change, never `aims`, `closes` or `applies`.

Run 3 of the multi-tenant tree showed the cost. `hyp signals` printed five `wasted` rounds, three `no-movement` pairs and one `spin`. Four of the five wasted rounds closed limits or applied an answer: E-FWTH64, E-YS13MW, E-BGHDGF and E-EKCW21. The one round that settled nothing, E-TT1HSQ, the ledger already showed while it mattered: `hyp status` printed `limit R-XFSQEE#1 [aimed_not_settled by E-TT1HSQ]`, naming the item and the Experiment.

Owner decision on 2026-11-09: progress on a question has one account, the ledger and the statuses. Nothing scores a round.

## Decision

- **`wasted`, `pending`, `no-movement` and `spin` are removed.** A second account of progress either agrees with the ledger, and adds nothing, or disagrees with it, and is a false alarm.
- **`one-rival`, `D1` and `unjudged` stay.** Each is already a write-time refusal or warning, or a mark that `hyp status` and `hyp show` print. `hyp signals` lists them for the whole tree.
- **`hyp signals --json` keeps the facts of each round:** its Experiment, cut, parent, targets, question, statuses at add, its marks and its Results with their status moves. It drops `pending`, `wasted`, `closed`, `line` and `moved`.
- **A run's cap stays outside hyp.** A `/goal` prompt or a recipe states how many Experiments a run may take.

## Consequences

- Old trees print fewer signal lines. No event, payload, status or ledger item changes.
- A stall the ledger cannot see stays unseen: an Experiment that aims at nothing owed, moves nothing and settles nothing. No real run has been called a stall for that reason yet.
- AGENTS.md step 1 now names the ledger, a refusal or warning, a design mark, or the owner as the event that points at hyp.

## When this stops holding

- If a real run has an unaimed Result that moves nothing and settles nothing, and the owner calls it a stall, add a `hyp record` warning that the Result left no trace, not a signal.
- If one ledger item is `aimed_not_settled` by two Experiments in a real tree, give the item one derived state word.

## Considered Options

- **Redefine wasted as a round that settled no ledger item and moved no status.** Rejected: it keeps two accounts of progress, and `spin` would still fire on a candidate that holds every case.
- **Refuse an Experiment that aims at nothing owed.** Rejected: E-S2G7KA, an overlap check the owner asked for, aimed at nothing owed and contradicted a supported candidate. A new case has no key until a Result names it, and `untested` and `undecided` Experiments carry no aims.
- **A ledger `stalled` state now.** Rejected until a real event: no item in a kept tree was missed by two Experiments.

## Amended 2026-11-10: a question's turn, and a stale select is owed

An architecture review found the decision gate of a question computed three times from conditions joined with `&&`: the ledger, the refusal of `hyp select`, and `hyp status`, each in its own order. A replay probe then showed a stall this ADR's guard missed. A question answered with H reopened as `selected_stale` after H's basis changed while its premise's question stayed answered, and nothing was owed to anyone: the ledger's own test of a standing select ignored stale, so `hyp select` said only "owes no select".

The owner decided (2026-11-10, #115):

- A question's **turn** is one derived value, read by the ledger, `hyp select` and `hyp status`, first match wins: `closed` (answered, deferred or withdrawn), `waits` (on runner work, an open reading with a reach, a referred limit, a premise, a sub-question, all listed in that order), `no_candidate`, `owed`. Spec §8.1 owns the rule.
- A standing select is an answered question; the ledger has no second definition. A reopened question that keeps its select is never owed a select of the same Hypothesis again: it waits.
- A stale select owes the runner a `stale` item, keyed on the stale node, when no other runner item and no open premise accounts for it: the runner reaffirms it or renews its basis.
- The word is "waits". "Pending" is a retired signal word, and "hold" stays the DesignQuestion's defer or withdraw, so "holds the select" is reworded in the spec, the glossary and the skill.

No event or exit code changes. `hyp status` prints `waits on …` where it printed `held: …`, and prints a sub-question the turn waits on. `hyp select` on an answered question names its state instead of a runner item.
