---
status: accepted
---

ADR 0001 to 0008 stand. This ADR leaves ADR 0002's rationale on a transcript reader unchanged. It was rewritten in place on 2026-11-02: the first version recorded only the recipe-base experiment and kept the template in the skill; the owner's decision on how a first version is made is added below.

# A first instrument starts from hyp's template, not an embedded base

A user who installs hyp into another repository and wants an instrument, such as a gate that runs an agent at a store cut and scores what it did, needs a first bundle. Before this decision that user started from a template the hyp-verification skill carried, `recipe-template.md`, and `RECIPE-FORMAT.md`. This repository did not: each of its runner instruments was copied by hand from the last one. The first skill-text gate instrument was copied from `recipe-authoring-arms@3` with no `from:` line, because `hyp recipe new --from` copies the same name only. The user path has a recorded event, in a tree whose records stay in the private development history: `hyp recipe new` refuses a first bundle without `--from`, and the template has no Standalone or known-answers layout. It also has a recorded negative: R-MHH312 (`recipe-authoring-arms@3`) found that a first-bundle skeleton gave no fewer stops than main, and the issue that proposed the skeleton closed with no change. Its first limit names the case it did not test: "a method needing Evidence or a pinned store might show stops main does not carry".

There is also a product reason, independent of stops. A user who installs only the binary has no skill and so no template, yet the binary enforces rules the template exists to help meet: a `format: 1` line, scripts inside the bundle, an `External:` line under each Evidence and Standalone command, and known answers with a must-pass and a must-fail case whose `input/` is not empty. B6 is that gap seen from the skill side.

We proposed embedding recipe bases in hyp, `base:<name>@<N>` bundles a user copies with `hyp recipe new X@N --from base:A [--from base:B]`, composed at copy time with one `from:` line per base, and retested R-MHH312 under its first limit with a method that needs a runner and a store cut. `investigations/recipe-base` decided it. Owner decision on 2026-11-02. How a first version is made was then decided by the owner on product grounds, with no tree, on 2026-11-02.

## Decision

1. **No base.** hyp embeds no recipe base, and `hyp recipe new --from` copies the same name only. The deciding Result is R-H5H95P (below).
2. **The binary makes a first version.** `hyp recipe new NAME@1`, with no `--from`, writes `recipes/NAME@1/` beside the nearest `.tree/` from a template embedded in the binary. The bundle holds `recipe.md` with its name, version and `format: 1` filled in and a placeholder subject, the sections Launch, Doctor, Drive, Evidence, Standalone and Cleanup with guidance, and `External:` guidance; a stub `check.sh` that always fails; and `known-answers/` with a must-pass and a must-fail case, each with a file in `input/`. It refuses (exit 1) a name that has any version in a library the lookup searches, and `NAME@N` above 1 without `--from` is a bad argument (exit 2).
3. **A generated bundle never passes unedited.** The stub check fails the must-pass case, so `hyp recipe known-answers` fails, and every line to replace holds the marker `hyp:placeholder`. `hyp add exp --recipe` refuses (exit 1) a bundle with the marker in any file outside a case's `input/`, naming the files.
4. **One template.** The binary's copy is the only one; `skills/hyp-verification/recipe-template.md` is deleted, and `RECIPE-FORMAT.md` describes the first-version flow.

The template is a layout, not a method: it holds no command that does anything, so it is not a base, and decision 1 stands. Decision 2 is an owner decision on product grounds, not a measured gain: R-MHH312 measured stops, and whether the binary alone can make a valid bundle is not a stop.

The base experiment, R-H5H95P (E-3S24B5, `recipe-base-arms@1`): fifteen sonnet runners built the held-out skill-text gate instrument, five on each of three hyp binaries: null (main, the template path), coarse (one base, `claude-runner-gate@1`) and fine (four composable bases, `claude-runner@1`, `transcript-score@1`, `gate-verdict@1` and `fixture-check@1`). The pre-registered bar, set by the owner, was that a base arm has no fewer end-state passes than null and at least 1 stop fewer on average.

| Arm | Mean stops | End state pass | Mean lines written past the bases |
|---|---|---|---|
| null | 1.20 | 5 of 5 | 1068 |
| coarse | 1.00 | 4 of 5 | 217 |
| fine | 0.60 | 4 of 5 | 249 |

Neither base arm cleared the bar, so h-null supports and h-coarse and h-fine contradict.

## Considered Options

- **One coarse base** (`base:claude-runner-gate@1`: Launch, Doctor, Drive, Cleanup, score and verdict). Contradicted: 0.20 stops fewer than null, one pass fewer.
- **Four fine bases composed at copy time.** Contradicted: 0.60 stops fewer than null, one pass fewer.
- **Stack a shared bundle in the sandbox.** Rejected before the tree: under ADR 0003 a check runs on a copy of its own bundle, so a check cannot import from another bundle.
- **A runner built into hyp.** Rejected by ADR 0001 and ADR 0002: hyp runs recipe commands, and its Go reads no transcripts.
- **A cross-name `--from` without bases, or bases from an installed folder or a git URL.** Not pursued: with no base shipping, they have nothing to copy that a user lacks.
- **Template only in the skill.** The state before this decision, the Null arm of `investigations/recipe-base`. Rejected: a user with only the binary has no template, and two copies of a layout drift.
- **A separate `hyp recipe template` command that prints the template.** Rejected: two commands would make a bundle, and printed text cannot write a directory with an executable script and case inputs.
- **`hyp recipe new NAME@1` from an embedded template.** Chosen.

## Consequences

- No base, `base:` ref, `hyp recipe list` or multi-`from:` line exists. The prototype stays on an unmerged branch in the private development history; `investigations/recipe-base/arms.txt` names its commits and the arm binaries by sha256.
- `hyp recipe new` has two forms: `NAME@1` for a first version and `NAME@N --from NAME@M` for a next one. A first version has no `from:` line.
- `skills/hyp-verification/recipe-template.md` is deleted. The template lives in `internal/recipe/template/` and ships inside the binary.
- The frozen `authoring-stops@1` pins hyp's old behaviour: its `first-bundle` probe expects `hyp recipe new` without `--from` to exit 2, and its `template-sections` probe reads the deleted template, so its known answers fail under this hyp. It is not edited; re-running it needs `authoring-stops@2`.
- The five recipes this repository keeps (`authoring-stops@1`, `id-citation-arms@1`, `recipe-authoring-arms@3`, `skill-text-arms@2`, `recipe-base-arms@1`) are not rebuilt, and AGENTS.md gets no rule that a new instrument starts from a base.
- R-H5H95P's limits stay on the record. Limit 1: both base-arm end-state failures came from known-answer cases whose `input/` is an empty directory, which `hyp sandbox remove` and git both drop: a base's case in both, and in fine-5 also a case its runner wrote (`E-3S24B5/empty-input-correction.txt`); fixed, the gaps are 0.40 and 0.80 stops, still under the bar. The template's cases each hold a file in `input/` for this reason. Limit 2: the bases cut the lines a runner wrote to about a fifth of null's, which the stop bar does not reward. Limit 3: one method, five sonnet trials per arm, candidates derived from three recipes this repository wrote. Writing saved is not a recorded event; a later issue that names one could reopen the question with a rule that counts it.
- ADR 0002's rationale on a transcript reader is unchanged: no recipe that drives an agent ships in core.
