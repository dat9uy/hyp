---
status: accepted
---

Superseded in part by [ADR 0007](0007-records-start-at-230.md): a bundle with no `format: 1` line is refused as a missing field; no older format is read.

Refined by [ADR 0003](0003-a-check-runs-the-same-way-in-the-gate-and-in-the-run.md) for how a check runs.

# Known answers gate every recipe run, and the sandbox owns every side effect

Before this decision, hyp vouched for a verdict script after the fact. A bundle's known answers were a script the operator ran as a re-test, its Result was recorded under an instrument Hypothesis, every measurement linked to that Hypothesis with `depends_on`, and containment checked those links (spec §7.6, reasons 4 and 5). Each guard was added for one recipe, and the recipe grew to match: the session-replay recipe was 3,620 lines of Python, 1,339 of them its own re-test, and its known answers were code that read live git history, host ports and real sandboxes. Ten recorded events showed a verdict that depended on something no recipe declared. We decided to split a recipe into a functional core and an imperative shell. The checks a verdict counts, a bundle's Evidence and Standalone commands, are pure functions of frozen input files, and hyp runs their known answers before every run. The steps that change things, Launch, Doctor, Drive and Cleanup, run inside a sandbox that owns their inputs, outputs and processes. Owner decision on 2026-10-30.

## Decision

- **Instrument.** A bundle with an Evidence or Standalone command is an instrument and must have known answers. A bundle without one, such as an environment recipe, is never gated.
- **Known answers are data in the bundle.** Each case is `known-answers/NAME/` with `input/` and `expect.json` (command, arguments, exit code, stdout, and an optional `repo.bundle`). The directory names are the case list. A case addresses only an Evidence or Standalone command.
- **hyp runs each case statelessly.** The case runs on a fresh copy of its input, in an empty working directory, with only `PATH`, a scratch `HOME` and `TMPDIR`. hyp clones a named git bundle into `repo/`, replaces the case directory's path with a token, runs the case twice and requires the same output both times, then compares it exactly with `expect.json`.
- **The gate.** Known answers run before every `hyp sandbox create` (for each recipe in stack order, before Launch), `hyp recipe evidence` and `hyp recipe run`, with the operator's hyp and with no cache. Any failure refuses the run. The sandbox record and the command record hold the outcome of each case.
- **The sandbox owns side effects.** `sandbox create --inputs FILE` copies and hashes a run's inputs. `sandbox remove` writes `changes.json` and `changes/`, every path in `wt/` that differs from a manifest written when create finished. Each Launch and Drive command runs in its own process group, which remove stops after Cleanup.
- **hyp's files carry a format number.** `sandbox.json`, `command.json` and `changes.json` get `"format": N`, a `go test` golden per file forces a bump when a shape changes, and a known-answer input that holds a hyp file of another format is refused.
- **One recipe format.** A command is whatever a `Run:` line names; every other file in a bundle is hashed content. A bundle is in the current format when its `recipe.md` frontmatter says `format: 1`, which `hyp recipe new` writes into every copy. hyp hashes a bundle in an older format without parsing it, so its pins still verify, and refuses every use of it before parsing, naming `hyp recipe new` as the way to copy it. Only containment, which judges Results already recorded, still reads an older bundle's commands, to know whether a pin is to an instrument. The line only refuses: no use path reads an older format (owner decision on 2026-10-30, at planning).
- **Frozen and Draft.** A version is frozen once a tree pins its hash, and its bytes never change. Until then it is a draft and may be edited in place.
- **Transcripts stay out of hyp.** Reading what an agent did matters only when the agent is the subject, in trials. A harness recipe holds that reader, so ADR 0001's rule that hyp's Go reads no agent transcripts stands.

## Considered Options

- **Answer each of those events' questions with a declaration** (declared outputs, a scorer recipe, hashed host agent files, a recorded version state). Rejected because each adds a rule on both sides of the Python and Go seam, the pattern that produced the events.
- **Keep instrument Hypotheses and add the gate beside them.** Rejected because the gate proves the same thing on every run, and the links it would keep are the source of event 10: a measurement in one tree cannot cite an instrument supported in another.
- **Run the gate with the sandbox's pinned hyp.** Rejected because a session replay pins an old hyp on purpose, and checks that call hyp would then fail on every replay of an old subject.
- **Pass a bundle with no known answers, recording `none`.** Rejected because only scratch-pg needed it, and a core rule should follow from the instrument concept, not from one self-review fixture.
- **Keep reading older formats behind a `format: 2` marker.** Rejected because the owner does not require old recipes to run. The marker would keep the reach test, the freeze of unreached-script bundles and a second set of refusals alive for bundles nobody runs. The `format: 1` line kept above is the opposite: it marks the one format hyp reads, and everything without it is refused unread.
- **No marker: any bundle the grammar accepts is usable.** Rejected because the older session-replay and scratch-pg versions parse under the one grammar, so hyp would run old bundles as environment recipes, which nobody needs.
- **Put a transcript reader, or a log of hyp calls, in hyp.** Rejected because both serve trials only, and every recipe that drives an agent as its subject is self-review today. A second such recipe would reopen this, as the second adapter at that seam.

## Consequences

- Containment drops reasons 4 and 5 and limit 4. Reason 2 also requires the record's known answers to have passed. Reasons 1 and 3 stay, and `recipe.IsScript` stays for reason 1. None of the 21 trees under `investigations/` has a reason-4 or reason-5 mark on 2026-10-30, so no existing status changes. Instrument Hypotheses, re-test Experiments, `hyp amend` on each version bump, and `depends_on` links to an instrument are no longer used for new bundles.
- ADR 0001's consequence about freezing is replaced. The unreached-script rule, the freeze it caused and `OlderForm` are deleted, and spec §9.5 choice 14 is replaced. "Frozen" now means pinned. Old bundles stay in the repository, unedited and hashed. Old trees' output may change, and the delivery diffs `hyp status` and `hyp check` across every tree before and after.
- The gate proves that a bundle reproduces its own expected outputs, not that those outputs are right. A weak or wrong case, such as the session-replay recipe's L-export case, which checked only the shape of its output, needs review. `hyp recipe check` warns when a command has no case that must pass or none that must fail, and the verifier reads whether a run's mode has cases. A weak case found after a version is frozen leads to a new version, and no Result already recorded goes stale.
- The checks cannot see two kinds of input: a read of a fixed absolute path and a read from the network both pass the double run. A daemon that leaves its process group, as Postgres does, still needs its recipe's Cleanup. Both are documented limits.
- A new version is cheap to make. Its re-test is the gate, which runs in seconds, not a set of live agent drives. In the session-replay recipe's next version, those drives become frozen OUT_DIRs, trimmed to the files each check reads, as known-answer inputs, and its lifecycle cases move into hyp's `go test`.
- `RECIPE-FORMAT.md`, the hyp-verification skill, `VERIFIER.md`, the spec, the glossary and `core-and-self-review.md` change to match. Each change is a correction that cites this decision, made in the same change as the binary behaviour it describes.
