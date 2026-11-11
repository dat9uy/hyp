---
status: accepted
---

Superseded in part by [ADR 0021](0021-a-skill-change-is-a-build-goal.md): the skill-text recipes it names as a follow-up are removed.

Extends [ADR 0005](0005-a-bundle-is-its-ref-and-hash.md) (a path says where a thing was, a hash says what it is) and [ADR 0008](0008-an-experiment-has-a-lifecycle-of-its-own.md) (inputs are bound by the bundle and by hash).

# A record keeps or identifies what a replay needs

Before the public release, hyp's records were measured for host paths, and both records named many: `sandbox.json` held the absolute `host`, `wt`, `tmp`, `bin`, the store, the inputs source and each recipe copy, and `command.json` held the absolute bundle path, the caller's directory and `arg_sha256` keyed by each file argument's host path. Experiment E-YKJ70D, in a tree whose records stay in the private development history, then showed the larger gap. It named four arm binaries, built from patches, by sha256 only: the patches and the files its Standalone check was given were in a scratch directory that no longer exists, so no one could replay it. The owner reframed #8 from "remove host paths" to "a replay needs only the pinned recipe, the base commit and the files the tree keeps" (dat9uy/hyp#8 issuecomment-6600201823), and decided that the sandbox, not each Launch script, assembles the subject (issuecomment-6600308366).

## Decision

- **A record keeps or identifies what a replay needs, and names no host path for it.** Every value in `sandbox.json` and `command.json` is an identity (a hash, a commit, a ref, an event id), a path relative to the record's own directory naming a file hyp copied there, or an anchored pointer `<name>/REL` to something hyp does not keep.
- **The sandbox assembles the subject from the base plus at most one patch per Arm.** A change that lands is a commit used as `--base`. A throwaway Arm is `hyp sandbox create --patch ARM=FILE`: hyp checks with `git apply --check` that it applies to the base before any command runs (exit 2 and no sandbox when it does not), copies it to `patches/<ARM>.patch`, records its sha256 under `patches`, passes the directory as `HYP_SANDBOX_PATCHES`, and remove keeps it as `OUT_DIR/patches/`. hyp applies nothing; code under test is built in Launch, and a binary it builds is named by its sha256 (spec §9.6).
- **A check's file arguments are kept beside its record.** `hyp recipe run` and `hyp recipe evidence` copy each file argument, written as a path (absolute, or relative with a `/`), to `DIR/inputs/N/<base name>`, hash it, and give the script another copy. A file over 1 MB is hashed, not kept. `inputs` in `command.json` replaces `arg_sha256`. This is the sandbox's `--inputs` and the gate's `$CASE/input/` copy, applied to a run (ADR 0003: a check runs the same way in the gate and in the run).
- **The anchors that remain** are the pointers to what hyp does not keep: `sandbox.json` `host` (its directory name) and `store` (`<host>/REL`), and `command.json` `path`, `cwd` and each directory argument (`<repo>/REL`). A directory outside the repository is written as given, and hyp warns that a replay cannot reach it.
- **Live paths are derived, not recorded.** A sandbox's `wt/`, `tmp/`, `bin/`, inputs copy and recipe copies are found from `~/.hyp-sandbox/<id>`.

The rule is stated inline in the hyp-verification skill's `RECIPE-FORMAT.md` (Per-run inputs, Runnable commands, the record tables and Self-containment), which ships without this file.

## Considered Options

- **Round-1 anchors only: rewrite every path as `<sandbox>/REL` or `<host>/REL`.** Rejected: nothing reads a kept record's locators, so anchoring them makes the record tidier without making a replay possible. The files a Standalone command is given would still be hashed and lost.
- **A leak scan that refuses a record holding a host path.** Rejected: it detects the symptom after the record is written and names no home for the bytes a replay needs.
- **Patches as files of the recipe bundle.** Rejected: a patch is the subject, not a fixed input of the procedure. In the bundle it would change the recipe's hash for every Arm, and a frozen recipe could never take a new Arm.
- **hyp applies the patch to `wt/`.** Rejected: an Experiment usually compares several Arms in one sandbox, and how an Arm is built (a worktree, a binary) belongs to the recipe. hyp checks that each applies and keeps it.

## Consequences

- `sandbox.json` is format 3 and `command.json` format 4. A format-2 sandbox still drives and removes, its old fields copied out as read. A format-3 command record's `arg_sha256` is still read. A known-answer case holding an older-format `input/sandbox.json` or `input/command.json` is refused as before, and none under `recipes/` holds one.
- `patches` is a reserved `--keep` name. A check's relative path arguments now work; a bare word is never a path.
- A frozen script that uses its argument's location, such as reading sibling files, gets a copy and may break. Every kept recipe's known answers already pass copies, and all five pass.
- Out of scope: rewriting recorded evidence such as E-YKJ70D, recipe output such as transcripts and logs, and building inside a known-answer case. The self-review recipes that will build Arms from patches (`id-citation-arms@2`, `skill-text-arms@4` with its variants as patches) are a follow-up issue. *Superseded in part by ADR 0021: the skill-text recipes are removed.*
