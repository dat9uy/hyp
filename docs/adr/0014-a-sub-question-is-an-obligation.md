---
status: accepted
---

# A sub-question is an obligation

Rule 4 of spec §7.1 stored pending work in a word about a claim: a supported Hypothesis with an open sub-question read `awaiting_subquestions`. Readers of "supported" misread it. Candidates needed an exception (#39). A question's reason said `inconclusive` or `selected_not_supported` when the work waiting was the sub-question (#35). A dependent went stale when its target gained a sub-question (#36's comment). The parent edge carried relevance down and a block up, and the spec never named the block as a relation. A Hypothesis's status also read a question's state, which is why `h3 depends_on h1` was refused as a cycle when h3 answers h1's sub-question.

Owner decision on 2026-11-05, in #51. There is no tree, because the owner decided the meaning directly.

**A sub-question must be settled, whichever way, before its parent can be the answer.** It is neither a refinement, which would not block, nor a premise, which would care which answer won. A premise is `depends_on`.

**The block moves from the Hypothesis to the question that selected it.** `awaiting_subquestions` stops being a Hypothesis status and becomes a question reason, right after `selected_stale`. `answered` also requires that the selected Hypothesis has no sub-question that is not withdrawn and not answered, so a deferred one still blocks. Nothing new is stored: the block is derived from the facts already there.

**What reads what.** A Hypothesis reads only other Hypotheses: its status from verdicts, its stale through `depends_on`. A question reads its child Hypotheses, the sub-questions of its selected one and its own `depends_on` targets. No edge runs from a Hypothesis to a question, so a status cycle can only run through `depends_on` between Hypotheses, and spec §7.2 drops the Hypothesis → sub-question edge.

## Consequences

- #35 is the reason word: a question with an open sub-question now names that work, as `awaiting_subquestions` when the parent is selected and `supported_unselected` when it is not.
- #36's Confirm behaviour is intended. Any answer discharges the obligation, so a Confirm on the parent counts again once the sub-question is settled, and a flip of the answer gives no `changed since` line. A claim that cares which answer won says so with `depends_on`: its question goes stale when that answer falls, and re-pointing the link lapses the Confirm.
- A Result recorded on a dependent while its target has an open sub-question is no longer stale once the sub-question is answered, because the target's status never changed.
- A lone candidate with an open sub-question stays `suggested: none`, now printing `awaiting_subquestions`.
- `h3 depends_on h1` and `s1 depends_on h1`, where h3 and s1 sit under h1, are accepted. Relaxing the rule can only accept more, so it cannot turn an old log into rejections.
- Old trees change at one place. In the 26 pre-release trees, `wasted-definition` at cuts 27 to 37 reads `q-def` as `open/awaiting_subquestions` rather than `open/selected_not_supported`, and `h-split` as `supported` rather than `awaiting_subquestions`. Goal, rejections, staleness, candidates and signals do not change at any cut.
- Withdrawing a sub-question still releases the block with no review. What may happen to a closed node is #43, and this decision does not touch it.

## Rejected

- **A premise.** The block would care which answer won, and `depends_on` already says that, so two relations would carry one meaning.
- **A refinement.** It would not block, so open work would sit unseen under an answered question.
- **An author-declared kind of sub-question.** It adds a stored field and a branch in every reader, on 2 real examples out of 52 questions.
- **Listing the sub-question's answer in the Dossier.** Any answer discharges the obligation, so what the answer says is not evidence for the parent. A claim that rests on it uses `depends_on`.
