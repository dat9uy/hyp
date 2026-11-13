# Issue #113: Owner flow: a /goal prompt generated from the tree, and the run's constraints

Opened by dat9uy at 2026-11-10T10:27:52Z. Text as of 2026-11-11T16:20:00Z.

## Event

On 2026-11-10, after the multi-tenant tree completed, the owner asked to replace planning in prose (`ak-plan`) with a `/goal` prompt generated from the tree. The examples are the hand-written prompts of runs 4 and 5 on that tree. From run 3 to run 5 they changed in only three places: the actor, one sentence of context, and the stop clause. Everything else had moved into hyp (OWED and the stop line `runner owes nothing: stop and report to the owner`) and into the runner's skill.

On 2026-11-11 the owner settled that **building is a Goal like any other**: hyp has no discovery kind and no delivery kind, so there is no separate delivery `/goal` (#114 is restated to match). The generated prompt therefore serves any Goal, including one whose candidate is a commit. The same session wrote two Goals with a grilling session, which is the owner flow's first step and has no skill yet.

The run-5 prompt, with its paths shortened:

```text
/goal Run the hyp tree in <tree> with the hyp-goal-run skill, as HYP_ACTOR=<actor>. Its Goal is <G-id>.
Done when the latest printed `hyp status` shows `runner owes nothing: stop and report to the owner` under OWED, or every runner item under OWED is `add rival Hypotheses` on a Question for which no owner checkpoint since the latest stop checkpoint asks for a new Hypothesis, or 5 Experiments have been recorded since this goal began, or you cannot take the next step honestly and say why. At each stop, record the stop checkpoint the skill names, and list the lines under LINES with the Results that ran them.
Never change <the investigated repository> or its issues; read it only, and work in throwaway copies.
Never put application code or real source data in the public hyp repository.
Treat <legacy code> as legacy. <...> may be test cases.
Never open or merge a PR.
```

## Gaps

1. **A run's constraints have no typed place.** They have a read side and a write side.
   - Read side: which repositories a run may only read, what data it may not publish, and what counts as legacy.
   - Write side (moved here from #114): which repository a run may change and how, for example commits on a local branch of one repository, merged only on the owner's word.

   Both live only in the prompt today, partly as prose in the Goal body, so a generator cannot read them from the tree.

Gaps 2 to 4 of the first version (publishing the owner-flow skills, the gate for a new skill, and the Goal-writing skill) moved to #134 on 2026-11-11, on the owner's word.

## Scope

Core. It covers:

- the `/goal` prompt for any Goal, as something hyp prints or a skill writes from the tree;
- the typed place of a run's constraints, read side and write side.

It does not cover what a Goal whose candidate is a commit needs from hyp; that is #114.

## Done means

Restated 2026-11-11 on the owner's word: the multi-tenant tree is deprecated, so its run-5 prompt can no longer be generated.

- The generated prompt, given G-D1CH1R as it stood before run 6, matches run 6's prompt (`goal-prompt-261111-2129-g1b-run-6`) clause for clause, except the run-specific inputs: the actor, the owner answer it starts from, and the branch and commit. The only other allowed difference is wording that no step depends on.
- For a Goal whose candidate is a commit, the generated prompt names the target repository and what the run may change, read from the recorded place.
- The "Never" lines come from a recorded place, not from a hand edit.

---

Comment by the owner at 2026-11-10T10:28:07Z:

Delivery, the next stage after a discovery run, is #114.

---

Comment by the owner at 2026-11-11T15:25:34Z:

Owner decision, 2026-11-11, gap 1: a run's constraints live in the Goal body under the headings `May change:`, `Read only:` and `Never:`. The Goal-writing skill writes them and the run obeys them; a `/goal` prompt may narrow them, never widen them. A typed field on `hyp add goal` waits until a real run misreads or loses one. The skill text is in PR #135; the prompt generator is still open.

---

Comment by the owner at 2026-11-11T16:10:04Z:

Folded in on the owner's word, 2026-11-11: **the skill descriptions predate the words discovery Goal and build Goal** (glossary, PR #137).

- `skills/hyp-goal-run/SKILL.md` describes itself as "an unattended discovery /goal", though it ran G-D1CH1R, a build Goal (`investigations/checkpoint-store`). The generated prompt this issue owes serves any Goal, so the skill that runs it should say so.
- `skills/hyp-prototype/SKILL.md` describes itself as "Investigate an uncertain design", while the same tree machinery now serves build Goals too.

Changing skill text is a build Goal of its own (ADR 0021), so this is done as part of that work, not as a separate text fix.
