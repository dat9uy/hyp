---
status: superseded by the owner's `land`, see the note below
---

Superseded by `land`: the owner's act that ends a Goal at a stop, with the state `landed` ([state-machine.md §6](../state-machine.md#6-states), [§8](../state-machine.md#8-acts), `land`; #245). The `answer` event, the `end` item and the state `ended` are v1 only.

Amends [ADR 0020](0020-one-store-holds-many-goals.md) (one store holds many Goals) and [principles.md](../principles.md) §2 and §7.

# The end is the owner's one terminal act on a Goal

Issue #177. A Goal the owner no longer wants stayed in progress forever: its readings waited for the owner, its items stayed owed, and a later Goal could not point at the fact that the owner gave it up. A run can teach the owner that a Goal is wrong or outgrown, and hyp had no move that says so. The owner settled the end in the grill of #177, recorded as G-K0AXM0 in `investigations/goal-end`.

## Decision

- **An end is an item of an answer.** `hyp answer --end G=TEXT` writes an `answer` whose `ends` holds `{key: <Goal id>, text}`. TEXT is the owner's words, 1 to 600 runes. Only the actor that added G ends it; a Goal a runner actor added is ended only under that runner's name. Any other actor is refused with exit 1 at write, and replay rejects such an end (`bad end: only the actor that added a Goal ends it`).
- **A Goal ends once.** A second end of G is refused at write and rejected on replay (`bad end: G-… is already ended`). In a merged log the first end of G by id order applies. One answer that ends G and also moves a key of G (an answer, accept or return of one of G's keys) is refused and rejected: an end is an answer of its own. Keys of other Goals may share the answer.
- **`ended` is a Goal state, derived.** A Goal with an applied end reads `ended` in place of `in_progress` or `complete`, whatever its questions. `hyp status` prints the end's words under the GOAL line, and `hyp show` prints them under `status:`. No stored field holds the word.
- **Nothing of an ended Goal is owed or waits.** The ledger owes no item on its questions, AWAITING OWNER lists none of its readings, referred limits or choices, its questions' turns are closed, and FRONTIER, STALE and OPEN LIMITS print nothing of it. `hyp show` still prints its lines, readings, limits and Results.
- **Nothing more is written under an ended Goal.** A write whose node leads to an ended Goal is refused at write with exit 1, naming the end: adds under it, Results, retracts, selects, defers, withdraws, resumes, subjects, edges from its nodes, checkpoints, and answer items on its keys. `hyp add goal` and cites from other Goals stay admitted. This is the binary's refusal, not a replay rule: replay applies such an event when a merged log holds it after the end, and `hyp check` warns about each one.
- **`hyp prompt` for an ended Goal exits 1**, names the end and prints nothing on stdout. It reads this one fact beyond the Goal's node.
- **An end settles every key of its Goal**, so a later line may cite it. The table of ADR 0020 gains one row:

  | The cited key, when the line was written | The cite means |
  |---|---|
  | an ended Goal's id, or a key or Hypothesis of an ended Goal not settled otherwise | ended |

  A key settled before the end keeps its meaning: an answered reading stays applied. `hyp show` of a line prints the end's words beside each cite of an ended Goal, and a line written before the end prints that the Goal has since ended. The id of a Goal not ended is refused as a cite with exit 1.

## Exceptions to the principles

The end is the one owner move that breaks three rules, and only for the Goal it ends:

- principles.md §2, "a later event can move the status again": no event moves an ended Goal again.
- principles.md §7 rule 3, every owner move but the lines answers what the ledger hands them: the end answers nothing the ledger hands the owner. It is the owner's own word on the Goal.
- principles.md §7 rule 5, an owner move changes no status: the end sets the Goal's state to `ended` with no Result.

## Consequences

- Text changed with this ADR: principles.md §2 and §7, the glossary's End and Ended, ADR 0020's cite table, replay-spec §4 (`answer`, the end), §5 rules 31 and 33, §5.1, §7.5, §7.6 and §8.1, owner-guide.md, skills/hyp-owner-answer, skills/hyp-owner-grill and skills/hyp-prototype.
- **Out of this ADR, as G-K0AXM0's body states:** a tool that lands a record (#203); the grill's Start entry for an ended Goal and its base rule (#205); a run rule against sub-questions under a held candidate; readings on withdrawn questions of live Goals; and `hyp status --all`. When `--all` lands, it prints nothing of an ended Goal under OWED or AWAITING OWNER.
- **The first use of the end waits for the new binary.** An older hyp rejects the `ends` field, so the ends of record:G-SZCBXZ, record:G-BPWZ24 and record:G-4N85GM wait until it is installed.
