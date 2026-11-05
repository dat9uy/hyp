---
status: accepted
---

# A Confirm counts while its Dossier stands

A Confirm judges a body of evidence, but hyp never modelled that evidence. So hyp could not say what a verifier must read, or whether a Confirm still stood once the evidence changed. Two events showed it. In #24, a question stayed answered after a Result its verifier had judged was retracted. In #26, a wrong `contradicts` (R-KA7Y6Z) sat outside the walk the verifier's brief describes, because that walk follows tree parents while a Hypothesis's status follows verdict edges.

The rules also had no single owner. Replay dropped Confirms per event (an amend, an owner's `insufficient`), froze the owner requirement at the select, and kept a request flag for re-verification, while the limit gate and other select and confirm refusals sat in the CLI. Owner decision on 2026-11-04, from the audit in #23.

The log is append-only and merged across worktrees, so a refusal cannot stop an event from existing. "Unrepresentable" therefore means that the derived `answered` status cannot be computed unless its Confirms still hold.

## Decision

- **The Dossier is derived, per question.** It holds the selected Hypothesis's body, the set of live Hypotheses, every live Result with a verdict on one of them, and the live links going out from those Hypotheses and Results, not counting `external_ref`. Results are found by verdict edge, not by tree parent. A rival's body and planned Experiments stay out, and so do links other nodes add pointing in.
- **A Confirm counts while the Dossier is what it judged.** A Confirm judged the Dossier as it stood at the Confirm's place in the merged log. hyp compares that with the Dossier now. When they differ, the Confirm lapses, and when they match again, for example after a retract, it counts again. This is the §7.3 snapshot rule applied to a Confirm, so "retract restores the earlier statuses exactly" holds here too. For each select and role, the Confirm that counts is the latest one whose Dossier equals the current one, so a verdict about evidence that no longer exists gives way to the earlier verdict about the evidence that does. It applies to verifier and owner Confirms, sufficient or insufficient, and limit acceptances follow the owner Confirm that made them. It replaces every per-event drop rule.
- **A reopened question closes by appending, never by touching old events.** When the claim is unchanged, a fresh Confirm on the same select is enough. When the evidence changes the claim, the agent adds a new Hypothesis that `supersedes` the old one and selects that.
- **One record owns the Decision.** Like `Exps` (ADR 0008), one replay record holds a question's select history, conflicts, Confirms, owner requirement, limits and gating, reason order and the verifier-loop chain. Render, signals and check read it.
- **The owner requirement is derived from the current state.** It reads the current trade-off candidates and the select's `--tradeoff` and `--scope-change` flags, and it is no longer frozen at the select. Candidates read verdicts only: a live child Hypothesis whose verdicts alone give it `supported`, whether or not it is stale or has an open sub-question. They therefore come only from facts the Dossier lists, so the requirement changes only with the Dossier, and every Dossier change lapses the Confirms (amended 2026-11-04, #39).
- **A lapse adds no status word.** The question reads `awaiting_verifier` or `awaiting_owner`, and `hyp show` prints what changed since the Confirm.
- **A contested rival is marked, not escalated.** A `contradicts` that alone keeps a rival from being a candidate is marked as deciding the owner requirement. §7.1 stands: any `contradicts` wins.
- **`hyp request-verify` is removed.** An insufficient verdict lapses when the Dossier changes, so a request has nothing left to do, and a new verifier can already confirm at any time. Spec rule 16 goes. Old `dq.request_verify` events stay readable and have no effect. The verifier-loop signal keeps only its trigger of two or more insufficient verdicts in a row.
- **The confirm refusals that read derived state join Admission.** The owner limit gate in `internal/cli/write.go` and its neighbours move into a pure function in `internal/replay`, with the same exit codes and messages and no new refusal. Refusals that read only stored events stay in `Validate` (spec §5.1), and that covers every select refusal, so no select rule moves.

## Considered Options

- **Narrow the Dossier to removals and candidate changes.** Rejected: it needs a list of event types, which is the habit that caused #24. In the census below, the wide rule costs one reopened question.
- **A digest field on the Confirm (`--dossier X`).** Not now: log order wrongly credits a Confirm with a Result that lands or merges after the verifier read but before it confirmed, and a field would catch that. It also adds a required flag and a fallback for old logs, and no recorded event shows this race. It can be added later.
- **Require the owner on a contested rival.** Rejected: the #23 audit found 13 of 14 sampled `contradicts` hold, so this would mostly ask the owner about verdicts that are right, on 4 of 25 answered questions.
- **Keep `request-verify` as the way to ask for a new look.** Rejected: it never starts a verifier, because hyp starts no sessions, and a skill's text does that. Only 2 request events exist against 9 insufficient verdicts in the census.
- **Refuse a Confirm or a change at write time instead.** Rejected: a merged log can still hold what a refusal stops, which is why the rule lives in the derived status.
- **Put the new refusals in `Validate`.** Rejected: `Validate` reads stored events only (spec §5.1).
- **List the candidates in the Dossier (#39).** Rejected: it puts derived status into a Dossier built from stored events.
- **A Confirm that remembers its candidates (#39).** Rejected: it adds a second rule beside the Dossier comparison.
- **A one-step verifier release (#39).** Rejected: a re-confirm could then drop the owner without saying so.

## Consequences

- A census replayed the rules over the final state of 25 answered questions, from 26 pre-release trees and the 2 public ones. One reopens: `recipe-trust` Q-EGFRZH, where a `supports` Result landed after its verifier confirmed. No owner Confirm lapses, and no question's owner requirement changes. Old trees may change, as the owner has accepted. The census covers final states, not every question at every point in its history.
- The verifier may still not read what the Dossier lists. `hyp show` makes the list exact, and `skills/hyp-prototype/VERIFIER.md` and `skills/hyp-prototype/SKILL.md` changed as corrections to this decision, without a replay. #26 has since measured the reading once, with `recipes/verifier-brief-arms@1` at the select's cut of `R-KA7Y6Z`'s question: no sonnet verifier opened that Result's evidence before its confirm, in 15 trials across `base` and two one-sentence variants of the brief, so each failed 5 of 5 and no sentence shipped (`R-CSRP4F`, in a private tree). That is one cut, one wrong Result and one model. The limit stands, measured.
- Asking the owner again when evidence changes costs their time on any owner-required question, which is the case where their sign-off is out of date.
- This decision is delivered as four stacked PRs. The first adds the Dossier and changes no status. The second adds the Decision record behind the same rules and changes no output. The third flips the rule, derives the owner requirement, removes `hyp request-verify`, and updates the spec, skills and glossary marks. The fourth moves the refusals into Admission and changes no output.
- Owner decision on 2026-11-04, in #39: a stale rival or one with an open sub-question could remove the owner with no Confirm lapsing, so the candidates stopped reading those states. A rival whose premise fell still needs the owner, or must be abandoned, superseded or contradicted, and abandon cannot be undone. The cuts scanned for #39's sub-question path showed no question losing its owner that way; the two paths through a `depends_on` target were not scanned in real trees.
- Amended 2026-11-05, #51: candidates no longer need the exception, because a Hypothesis's status now reads verdicts only (ADR 0014).
- Out of scope: comparing verdicts with checks (ADR 0004), a judge or modelled field on Results, and limits blocking a plain select. Evidence file marks such as questionable and uncontained stay out of the Dossier, because replay reads no files.
