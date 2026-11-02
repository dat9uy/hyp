---
status: accepted
---

Supersedes in part [ADR 0002](0002-known-answers-gate-every-recipe-run.md) (an older recipe format) and [ADR 0005](0005-a-bundle-is-its-ref-and-hash.md) (`SingleFile` in the library lookup).

# Records start at the decided_by cut

One change made every Experiment say how its verdicts are decided: it pins an instrument, or it declares `decided_by: reading`. Every tree written before it lacked the declaration, so hyp kept code that only those trees needed: the `undeclared` flag, a branch for recipe bundles with no `format: 1` line (ADR 0002), and single-file recipes `recipes/<name>.md`, which could be dropped only once no tree pinned one. Core tests read those trees, and some of them held content that could not be published. Owner decision on 2026-11-01: every record written before that change is prototype history and is dropped.

## Decision

- **The cut is the change that made every Experiment declare how its verdicts are decided.** All 26 trees under `investigations/`, `fixtures/`, all 13 recipe versions under `recipes/` and one older recipe copy kept with the plans are deleted. None of their Experiments has a `decided_by` field. Records made before the public release stay in the private development history. The open questions that no closed issue or merged PR answered moved to open issues.
- **Core tests read synthetic data.** A store written by the current hyp, `internal/testutil/testdata/synth`, with its `generate.sh` and a synthetic instrument `count-check@1`, replaces the real trees in the replay, cut, signals, amend and recipe-check tests. The cut tests assert explicit expected cuts instead of comparing with a Python cutter.
- **No undeclared flag.** It could only mark an Experiment written before the cut. The one case a new tree reaches, a pin to an instrument whose bundle stops loading, is a `drift ... now missing` line from `hyp recipe check`.
- **`format: 1` is a required field.** `Load` refuses a bundle without it before parsing, with exit 1, like any other missing field. `LoadForUse`, `OlderFormat` and the allowance for a Result that cites its own Experiment's older recipe are gone. `hyp recipe hash` still hashes such a bundle, exit 0, now without a warning.
- **A single file is not a recipe.** `Load`, `Scan`, `hyp recipe check`, `hyp record` and the sandbox know only `recipes/<name>@<version>/`. A ref only `recipes/<name>.md` could answer reads as not found, exit 2.

## Considered options

- **Keep the trees and mark them legacy.** Rejected: the code paths would stay for records hyp no longer writes, and content that could not be published would stay in the tests.
- **Rebuild the trees in the current shape.** Rejected: a rebuilt tree would claim Results its Experiments never declared. The questions that are still live are issues instead.

## Consequences

- hyp's code, spec and docs read only the shapes hyp writes now. Exit codes are unchanged except that a single-file ref now exits 2, the code a missing recipe has always had.
- No replay instrument was left in the repository. The skill-text gate in AGENTS.md waited for one built from a tree written after the cut; `recipes/skill-text-arms@3` is that instrument now.
- The `frozen` and `draft` labels, `HashOnly` and the rename hint for a leftover `recipes/<name>/` folder stay: each describes a bundle a new tree can make.
- A sandbox record or store that an older hyp binary left on disk is still read as before; it is runtime state, not a record in this repository.
