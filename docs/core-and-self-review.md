# Map: core and self-review

This repository holds hyp, its skill, and the Goals that built hyp v2. Records made before the public release stay in the private development history. This page says which concepts and files belong to which scope, so that an issue, a plan or a change can say which scope it serves. The two words are defined in [glossary.md](glossary.md), "Scopes of this repository".

## The test

Core and self-review are scopes. An item's scope is the concept it belongs to, not who uses it or what tested it.

- **Core** is every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A user records Goals, places examples and runs lines on their own code. So Goal, line, example, Probe, runner and select are core concepts.
- **Self-review** is the recipes this repository keeps to test hyp, whatever repository the runner works on, plus the inputs only they use. Nothing else is self-review.

The repository a runner works on is its investigated repository, a core concept. Here it is often hyp itself. Which repository it is does not change the scope.

Investigating hyp is an activity, not a scope. A Goal takes its place by its question:

- A Goal about hyp takes the scope of the concept it changes. Every Goal here changes a core concept, so each is core-scope work.
- A Goal whose question is about another project is a self-review input: a run this repository keeps as evidence about hyp, not for its answer.

Three rules keep the boundary honest:

1. **A concept stays core when self-review is what tests it.**
2. **A self-review recipe does not make its kind self-review.** Recipes are core concepts of hyp 0.1.0.
3. **Self-review invents no concept of its own.** Every word a self-review recipe uses maps onto core concepts (see "Self-review words" below). Work whose question is about a self-review recipe itself is a special case, and it needs a reason. Today there is one: an instrument must give its known answers before its verdicts count, which is a core rule applied to a self-review recipe.

There are exactly two scopes. An item that fits neither is a problem that has not been scoped well enough: split it until each part has one scope. `.agentkit/` is not an item of this project at all: it is the local development setup for working on the repository, like `.gitignore`.

## Visibility

Visibility is a separate axis from scope. Every item in this repository is public. A self-review row below marks its own visibility.

## Self-review

`main` holds no self-review item. A new self-review recipe gets a row here, with its visibility.

The self-review recipes kept before the public release (session-replay, scratch-pg, readout-read, containment-arms, declaration-shape and signals-replay-proto), their inputs, and the trees that pinned them are not published.

The self-review recipes and trees of hyp 0.1.0 are kept on the `v1` branch, not on `main`: `recipes/id-citation-arms@1` and `@2` with `investigations/id-citation`, `recipes/authoring-stops@2`, `recipes/recipe-authoring-arms@4`, `recipes/recipe-base-arms@1` with `investigations/recipe-base`, and `investigations/goal-prompt` with its `hyp-owner-grill-eval@1` and `@2`. This file on that branch maps each one to its scope and visibility, together with the trees that decided and built the shared store of hyp 0.1.0.

## Core

Every item on `main` is core:

- **The binary:** `cmd/`, `internal/`, `go.mod` and `go.sum`. The binary builds only the v2 packages that `cmd/hyp` imports. The packages of hyp 0.1.0, such as `internal/cli`, `internal/replay`, `internal/sandbox` and `internal/recipe`, are still under `internal/`. The binary does not import them, and `go test` still runs their tests.
- **The skills:** `skills/hyp-grill`, with its run rules and its generic Probe, and `skills/hyp-statechart`, with the statechart tool. The skills of hyp 0.1.0 are kept on the `v1` branch.
- **The Goals:** `goals/`, with each Goal's examples, Probes and specs, and the store `.tree/` that records them. Each Goal changes a core concept of hyp, so each is core-scope work.
- **The docs** in this directory. [state-machine.md](state-machine.md) is the v2 spec. [replay-spec.md](replay-spec.md) and [sandbox-spec.md](sandbox-spec.md) are the specs of hyp 0.1.0.
- **The test data** under `internal/*/testdata`: synthetic stores, goldens and fixtures written for hyp's own tests. The store in `internal/testutil/testdata/synth` and its `generate.sh` are history from hyp 0.1.0, which the packages of hyp 0.1.0 still replay.

## Self-review words

A word a self-review recipe uses is a name for core concepts. A new self-review word gets an entry here, and a word that maps onto no core concept is a sign that a core concept is missing. The words below are those of the hyp 0.1.0 recipes on the `v1` branch.

| Self-review word | In core terms |
|---|---|
| arm (id-citation-arms) | an Arm: a hyp binary built in Launch from main or main plus one rival's patch, named by sha256, the subject the cases run (version 1 read binaries built by hand from absolute paths) |
| arm (recipe-authoring-arms) | an Arm: a hyp binary, main or main plus one rival's prototype patch, pinned as the sandbox's `--hyp` |
| arm (recipe-base-arms) | an Arm: a hyp binary built from main, or from main plus the embed prototype and one rival's base candidates, pinned as the sandbox's `--hyp`; `null` embeds no base; the candidates were not shipped (ADR 0009) |
| stop (authoring-stops, recipe-authoring-arms) | a point where an author turning a method into an instrument is held up; counted as a failing hyp command or a fingerprint in the bundle or transcript |

## Where the scopes touch

**Core tests read only core test data.** The v2 tests build their stores in temporary git repositories and compare against goldens under `internal/channel/testdata` and fixtures under `internal/example/testdata`. The tests of the hyp 0.1.0 packages `check`, `store`, `sandbox`, `signals`, `cli`, `cmdrecord` and `recipe` read the synthetic store in `internal/testutil/testdata/synth`. The sandbox tests also stack `internal/sandbox/testdata/recipes/scratch-pg@1`. A copy of the repository that holds only core passes `go test`.

**Self-review text inside core docs.**

- [sandbox-spec.md](sandbox-spec.md) §5.2 records choices made in design trees that are core-scope, so that is ordinary provenance.
- On the `v1` branch, `skills/hyp-verification/RECIPE-FORMAT.md` names scratch-pg in its examples. That is a choice of example rather than a split.

Skill text reaches every run that loads the skill, so moving a passage out of a skill is its own change, a build Goal like any change to `skills/`. This page is the owner decision such a change can cite.
