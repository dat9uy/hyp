---
status: accepted
---

# Scope change folds into Goal lines

Since ADR 0015 the select is the gate, and the owner confirm survived for one job: a select marked `--scope-change` waited for the owner's `sufficient`, and that confirm lapsed when its Dossier changed (ADR 0011). That lapse rule was the Dossier's only effect on any status. Spec §7.4 already said that for a plain select, a Dossier change that adds no candidate leaves the question answered.

Since #102 a Goal's lines say what the Goal requires, and they move by typed events: a runner refers a line the answer will not meet and the owner accepts or returns it, and a Goal that needs more is a new Goal, so a new tree (ADR 0016 amendment). A scope-change select said the same thing in a second way that no line could read.

In the 6 trees whose events `hyp events --json` could read on 2026-11-10 there were 6 selects, none with `--scope-change`, and 4 owner confirms, all written before ADR 0015. Owner decision on 2026-11-10 (#107): fold scope change into the moves the lines already have.

## Decision

| A scope change that… | Was | Is |
|---|---|---|
| narrows the Goal: the answer will not meet a line | `hyp select --scope-change`, then `hyp confirm --role owner --verdict sufficient` | the runner refers the line (`hyp checkpoint --refers G-…#n`), and the owner accepts it |
| widens the Goal: the answer needs what no line holds | the same | a new Goal, so a new tree that cites this one |

- **`hyp confirm` and `hyp select --scope-change` are retired.** Both exit 2 before anything else and name the refer and accept of a line, or a new tree. No command writes a `confirm` event or a `scope_change` flag.
- **Old events are history.** An old `confirm` applies only the `accepts` it made, at its place in the log, as before, so every limit keeps its state. Its verdict and role apply nothing, and a verifier confirm applies nothing at all. An old `scope_change` flag on a select changes nothing.
- **The Dossier, Lapse and `awaiting_owner` are removed.** Without an owner confirm, nothing reads the Dossier, so it is no longer derived. A question with a select that stands is answered. `hyp show` prints no `dossier:`, `owner:` or `changed since` lines, nor the Dossier's mark of a `contradicts` that decides a trade-off, and `hyp status` prints no `awaiting_owner` line.
- **A limit no longer has a gating flag.** It only marked which open limits held the owner's confirm. An open limit holds the select through the ledger whichever Hypotheses its Result judged (ADR 0016), and `hyp show` drops the `does not gate` mark.
- This supersedes what remained of ADR 0011: the Dossier, its lapse rule and the owner requirement. The Decision record stays: the select history, the conflict set and the reason order.

## Consequences

- A tree written before #102 has no lines, so its scope-change confirm was its only owner gate on scope. In the kept trees no select used it. A question answered by an old scope-change select now reads answered without a confirm. The owner accepted that old trees may change.
- Before and after this change, `hyp status`, `hyp check` and `hyp render --markdown` print the same in all 300 trees on the maintainer's machine, most of them known-answer fixtures.
- Widening a Goal costs a new tree. That is intended: the Goal is written once.
- A change of a question's scope, rather than the Goal's, belongs to #105. A withdraw would then also be a refer plus an owner accept.
- Spec §4, §5, §5.1, §7.4, §8 and §8.1, the glossary, `docs/principles.md` §7 rule 1, and the hyp-prototype skill change to match. The skill change agrees with this owner decision, and its commit names it.

## Considered Options

- **Keep the owner confirm for a scope change that widens the Goal.** Rejected: a widening needs a line no tree holds, and lines are written once, so it is a new Goal.
- **Keep the Dossier as a view in `hyp show`.** Rejected: with no confirm to lapse, it only restated the experiments, limits and edges that `hyp show` already prints.
- **Refuse old confirm events at replay.** Rejected: their accepts settle limits in kept trees, and rejecting them would reopen those limits.
