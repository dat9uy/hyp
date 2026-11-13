---
status: accepted; amended 2026-11-13: a prose-only change ships as a plain PR
---

Supersedes in part [ADR 0007](0007-records-start-at-230.md) (the skill-text gate's replay instrument), [ADR 0010](0010-a-record-keeps-or-identifies-what-a-replay-needs.md) (the follow-up that rebuilds `skill-text-arms` Arms from patches) and [ADR 0015](0015-the-select-is-the-gate.md) (the frozen verifier recipes).

# A skill change is a build Goal

AGENTS.md gated skill text apart from code. A skill-text change shipped only when a session replay from the event's store cut showed the failure in 3 or more of 5 trials without the change and in 1 or fewer of 5 with it, and nothing shipped when no cut reproduced the failure. Its one instrument, `recipes/skill-text-arms@5` with the rubric `rubrics/skill-text-gate@1`, replays the id-citation tree's e-arms with `hyp-prototype` and `hyp-verification` loaded. It cannot reach a skill that runs before any tree exists, such as `hyp-owner-grill`, or one that spans a whole `/goal` run, such as `hyp-goal-run` and `hyp-owner-answer`. For those skills the rule was a permanent freeze, and a new skill needed its own exception. [how-hyp-improves-itself.md](../how-hyp-improves-itself.md) had already found that replay studies of skill text decided nothing, because agents with the skill loaded rarely made the mistakes they counted.

Owner decision on 2026-11-11 (#134): building is a Goal like any other (ADR 0020), and a change to `skills/` is building. The special gate was context that led agents to design gate policy instead of applying that model.

## Decision

- **A change to `skills/`, a new skill included, is a build Goal like any code change.** Its candidate is the commit; its lines state the skill's behaviour; each line is checked by its own test or by an instrument the Goal builds, red on base, and the select ends it. For a new skill, base is no skill.
- **A prose-only change ships as a plain PR** (owner amendment, 2026-11-13). It adds, removes or rewords skill text and changes nothing the binary holds: no command, flag, default, cap or exit code a skill names. The PR names the real event and the owner's decision, and its gate is `go test`, which includes skillcheck. A new skill, a change to what a skill says about the binary, and any change the owner asks to have measured stay build Goals. The first such PRs added the smoke rule to `hyp-goal-run` and replaced the retired word "verdict rule" with rubric; a Goal of trials for either would have cost more than the trials it guards.
- **How a skill is tested is that Goal's choice**, settled in its grill. One choice is ak-skill-creator's eval: cases in its `evals.json` format kept as a file of a `recipes/<skill>-eval@N` bundle, graded by `eval_skill.py`, which is named on the command's `External:` line and never copied. hyp keeps no instrument and no threshold for skill text as such.
- **A skill error seen during another run** is an out-of-Goal finding, or a new Goal by promotion. No audit cadence is kept.
- **Unchanged:** a change still needs a real event, and a failure detectable from the tree goes into the binary as a refusal or a warning, gated by `go test`.
- **Removed:** `recipes/skill-text-arms@3`, `@4` and `@5`, `rubrics/skill-text-gate@1`, and the retired `recipes/verifier-brief-arms@1` and `recipes/verifier-audit-arms@1`.

## Alternatives

- **Replay from eval cases instead of a store cut, keeping 3 of 5 against 1 of 5.** Rejected: it keeps skill text apart from code, and the count is a test design that belongs to the Goal.
- **A standing audit after each real use of a skill.** Rejected: an error seen in a run already has a home, as an out-of-Goal finding or a new Goal.

## Consequences

- No published tree pins a removed recipe. `investigations/recipe-base` names `skill-text-arms` only as the held-out subject its runners authored, and `recipes/recipe-base-arms@1` carries its own copy of that method.
- The removed bundles stay in git history; an old Result that names one can still be read from there.
- `id-citation-arms`, `recipe-base-arms`, `recipe-authoring-arms` and `authoring-stops` are kept.
