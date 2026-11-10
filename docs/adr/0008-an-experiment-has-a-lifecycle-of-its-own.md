---
status: accepted; extended by ADR 0018: an Experiment may sit under a DesignQuestion
---

# An Experiment has a lifecycle of its own

hyp modelled an Experiment's outcome but not its plan. `node.add` pins the procedure (`recipe` and `recipe_hash`), the `targets` and `decided_by`, and the status (`planned` or `run`, spec §7.6) came only from its Results. Two parts of the plan were recorded nowhere hyp could see. The first is intent: whether the Experiment will still run, and what replaced it. Every other node kind has a closing event (`hypothesis.abandon`, `dq.withdraw` and `dq.defer`, `result.retract`, `supersedes`), but an Experiment had none, so it could never leave `planned`. The second is inputs: files passed to a command as arguments were recorded in `command.json` by path only. The rules about an Experiment also had no owner: "planned" was derived in five places, and `hyp check` disagreed with replay about an Experiment whose only Result is retracted. Owner decision on 2026-11-01, after two recorded events in trees whose records stay in the private development history: two Experiments stayed `planned` with no way to close, and in another tree a fixture and a cases file sat outside the pinned bundle.

## Decision

- **The lifecycle mirrors the Hypothesis's.** An Experiment is `planned`, `run`, `superseded` or `withdrawn`. Replacement reuses the `supersedes` edge, extended to two Experiments under one DesignQuestion, with no self edge and no cycle. `experiment.withdraw`, with an empty payload, closes one that nothing replaces. Both are refused while the Experiment has a non-retracted Result. A live Result beats a close: in a merged log a Result under a closed Experiment still applies, the Experiment reads `run`, and `hyp check` names it. A closed round is neither pending nor wasted in `hyp signals`, and keeps its D1 and one-rival marks. Question and Goal states do not change.
- **One record owns an Experiment's derived state.** `Projection.Exps[id]` holds `Status`, `Live` (the non-retracted Results), `Judged`, `OneRival`, `Unjudged` and `ReplacedBy`. Replay computes it once (ADR 0006), and render, signals and check read it instead of keeping their own copies.
- **Admission is pure.** `AdmitExp` and `AdmitResult` in `internal/replay` return the refusals and warnings for a new Experiment or Result from the state, the payload and the recipe facts the CLI loads. The CLI only loads those facts and maps findings to exit codes. `Validate` gains no write-time rule (spec §5.1).
- **Inputs are bound by the bundle and by hash.** A fixed input that is not the subject, such as a fixture or a cases file, is a file of the bundle, so the recipe pin hashes it. A file outside the bundle is the subject, or a read or write the command names. Each Evidence and Standalone command in `recipe.md` states its outside reach on an `External:` line, and `hyp recipe check` warns when the line is missing or still reads `TODO`. `hyp recipe run` and `hyp recipe evidence` record an `arg_sha256` for every file argument in `command.json`, bound at run time, not when the Experiment is added. hyp does not try to detect a fixture outside the bundle.

## Considered Options

- **Extend `hypothesis.abandon` to Experiments.** Rejected: abandoning is a judgement about a claim, and an Experiment makes no claim. A separate event keeps each kind's closing word its own, as `dq.withdraw` does for questions.
- **A `replaced_by` field on the old Experiment.** Rejected: it would be a second way to say what an edge already says, written on the wrong node. The `supersedes` edge is written by the replacement, can be removed, and is already checked for cycles.
- **An `inputs` field on `node.add`.** Rejected: an input declared at add time can still differ from what the command was given. A fixed input belongs in the bundle, where the recipe pin hashes it, and the subject is bound when it is run, by `arg_sha256`.
- **A required Experiment body.** Rejected: a body is prose that hyp cannot check, and an Experiment's procedure is already pinned by its recipe.
- **A structured argument declaration in `recipe.md`.** Rejected for the reason ADR 0004 gives: a declaration beside the code is a second copy of the rule that can disagree with it. The `External:` line is a warn-only statement for readers, never a `recipe.Load` field.

## Consequences

- This decision is delivered as four stacked PRs. The first adds `Projection.Exps` and this record, and changes no output. The second adds `AdmitExp` and `AdmitResult` and deletes the CLI's copy of the one-rival rule, and changes no output. The third adds the `supersedes` edge between Experiments, `experiment.withdraw` and `hyp withdraw` on an Experiment, and records the replacements in the tree where the stale Experiments sat. The fourth adds `arg_sha256`, the `External:` line, and the self-containment correction in RECIPE-FORMAT, the glossary, spec §9 and AGENTS.md.
- `id-citation-arms@1` stays frozen. Its fixture outside the bundle is a recorded limit, not something to edit.
- Frozen bundles without an `External:` line still load and still record, and `hyp recipe check` warns on each.
- Out of scope: detecting fixtures outside the bundle, and any change to Question or Goal states. ADR 0001 to 0007 stand.
