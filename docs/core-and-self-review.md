# Map: core and self-review

This repository holds hyp, and a few recipes and trees that test it. Records made before the public release stay in the private development history. This page says which concepts and files belong to which scope, so that an issue, a plan or a change can say which scope it serves. The two words are defined in [glossary.md](glossary.md), "Scopes of this repository".

## The test

Core and self-review are scopes. An item's scope is the concept it belongs to, not who uses it or what tested it.

- **Core** is every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A user writes recipes, runs sandboxes and records trees for their own code, so recipe, sandbox, store cut, runner and instrument are core concepts, even where only self-review uses them today.
- **Self-review** is the recipes this repository keeps to test hyp, whatever repository the runner works on, plus the inputs only they use. Nothing else is self-review.

The repository a runner works on is its investigated repo, a core concept (spec §1). Here it was often hyp itself, and often a real outside project: session-replay's snapshots were sessions of a runner working on another application. Which repository it is does not change the scope.

Investigating hyp is an activity, not a scope. A tree takes its place by its question:

- A tree whose question is about hyp takes that question's scope. Most trees here ask whether a core concept works, so they are core-scope work, carried out with self-review recipes.
- A tree whose question is about another project is a self-review input: a run this repository keeps as evidence about hyp, not for its answer.

Three rules keep the boundary honest:

1. **A concept stays core when self-review is what tests it.** The runner was measured by session-replay, and it is still a core concept.
2. **A self-review recipe does not make its kind self-review.** session-replay was a recipe; recipes are core.
3. **Self-review invents no concept of its own.** Every word a self-review recipe uses maps onto core concepts (see "Self-review words" below). Work whose question is about a self-review recipe itself is a special case, and it needs a reason. Today there is one: an instrument must give its known answers before its verdicts count, which is a core rule applied to a self-review recipe.

There are exactly two scopes. An item that fits neither is a problem that has not been scoped well enough: split it until each part has one scope. `.agentkit/` is not an item of this project at all: it is the local development setup for working on the repository, like `.gitignore`.

## Visibility

Visibility is a separate axis from scope. Every item in this repository is public. A self-review row below marks its own visibility.

## Self-review

The self-review recipes kept before the public release (session-replay, scratch-pg, readout-read, containment-arms, declaration-shape and signals-replay-proto), their inputs, and the trees that pinned them are not published. Records made before the public release stay in the private development history.

When a question needs an instrument, it is a recipe with known answers that measures, and a rubric in `rubrics/` that turns its measures into verdicts, both pinned by the Experiments that use them (ADR 0018). A new self-review recipe gets a row here, with its visibility.

The self-review recipes and trees of hyp 0.1.0 are kept on the `v1` branch, not on `main`: `recipes/id-citation-arms@1` and `@2` with `investigations/id-citation`, `recipes/authoring-stops@2`, `recipes/recipe-authoring-arms@4`, `recipes/recipe-base-arms@1` with `investigations/recipe-base`, and `investigations/goal-prompt` with its `hyp-owner-grill-eval@1` and `@2`. This file on that branch maps each one to its scope and visibility.

## Core

Every item is core: `cmd/`, `internal/`, `go.mod`, `go.sum`, the skills, and the docs in this directory, the specs [replay-spec.md](replay-spec.md) and [sandbox-spec.md](sandbox-spec.md) among them. The test data under `internal/*/testdata` is core too: synthetic stores and bundles written by hyp for its own tests, such as the store in `internal/testutil/testdata/synth` and its `generate.sh` (that store is history from before ADR 0015, which a current hyp still replays).

A new tree under `investigations/` that asks about hyp takes its question's scope, as the test above says. A tree that asks about a self-review recipe itself is the special case under rule 3, and its reason is instrument trust.

`investigations/recipe-base`, on the `v1` branch, asks whether hyp should embed recipe bases, so it is core-scope work. Its Result R-H5H95P selected templates only (ADR 0009), so the base candidates it tested were not shipped. `investigations/recipe-base/arms.txt` names the prototype branch they were built on, which stays in the private development history.

`investigations/checkpoint-store-design` and `investigations/checkpoint-store`, on the `v1` branch, share one question, hyp's shared store, so both are core-scope work. The first is a discovery Goal, G-73R704, that records the design (ADR 0020); the second is the build Goal that implements it, G-D1CH1R, whose Hypotheses are hyp commits. Neither pins a recipe.

## Self-review words

A word a self-review recipe uses is a name for core concepts. A new self-review word gets an entry here, and a word that maps onto no core concept is a sign that a core concept is missing.

| Self-review word | In core terms |
|---|---|
| arm (id-citation-arms) | an Arm: a hyp binary built in Launch from main or main plus one rival's patch, named by sha256, the subject the cases run (version 1 read binaries built by hand from absolute paths) |
| arm (recipe-authoring-arms) | an Arm: a hyp binary, main or main plus one rival's prototype patch, pinned as the sandbox's `--hyp` |
| arm (recipe-base-arms) | an Arm: a hyp binary built from main, or from main plus the embed prototype and one rival's base candidates, pinned as the sandbox's `--hyp`; `null` embeds no base; the candidates were not shipped (ADR 0009) |
| stop (authoring-stops, recipe-authoring-arms) | a point where an author turning a method into an instrument is held up; counted as a failing hyp command or a fingerprint in the bundle or transcript |

## Where the scopes touch

**Core tests read only core test data.** Core tests read the synthetic store in `internal/testutil/testdata/synth`: `TestRunTreesReplayWithoutRejections` and `TestRepositoryRecipePinsMatch` (check), `TestCutMatchesExpected` (store), `TestCreateStoreCut` (sandbox), the signals goldens, `TestAmendShowsTheNewDesign` and `TestSignalsReplaysOnce` (cli), and `TestReadFixture` (cmdrecord). `TestStackedScratchPg` stacks `internal/sandbox/testdata/recipes/scratch-pg@1`. A copy of the repository that holds only core passes `go test`.

**Self-review text inside core docs.** The core part of each passage is the general rule; the self-review part is session-replay's versions and history.

- `skills/hyp-verification/RECIPE-FORMAT.md`: the examples that name scratch-pg teach core rules with a recipe this repository no longer keeps, which is a choice of example rather than a split. The stacking example names a generic `harness@1`.
- [sandbox-spec.md](sandbox-spec.md) §5.2 records choices made in design trees that are core-scope, so that is ordinary provenance.
- [how-hyp-improves-itself.md](how-hyp-improves-itself.md), "Guard the loop while it runs, not before", describes the tripwire that stops a run; the session-replay recipe whose `tripwire.py` built it stays in the private development history.

Skill text reaches replayed runs, so moving a passage out of the skills is its own change, a build Goal like any change to `skills/`. This page is the owner decision such a change can cite.
