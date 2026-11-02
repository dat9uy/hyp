---
status: accepted
---

Superseded in part by [ADR 0007](0007-records-start-at-230.md): single-file recipes and `SingleFile` are gone.

# A bundle is its ref and its hash

hyp pins a recipe by its ref and the hash an event recorded for it (spec §7.6), and `RECIPE-FORMAT.md` already said the directory name is not hashed. The code disagreed in three places. `recipe.Open` took a bundle's name and version from its folder name and skipped the version check for a folder without `@`, and every copy hyp makes went through that skip: the sandbox's copy at `recipes/<name>/`, the known-answer gate's copy and the `hyp recipe run` or `evidence` copy at `<tmp>/<name>/`. `hyp sandbox remove` reopened the copy named in `sandbox.json` and ran its Cleanup without comparing it with the record's `sha256`, so Cleanup could run from a copy that changed after create checked it. And after recipes moved to versioned folders, a pin to `x@N` whose bundle sat in a leftover `recipes/x/` reported `now missing`, and `hyp recipe check` advised `hyp recipe new` where `git mv recipes/x recipes/x@N` keeps the hash. Owner decision on 2026-10-31.

## Decision

A bundle is identified by its ref and its content hash. A `recipes/<name>@<N>/` directory is how a library finds a version; a copy hyp makes is checked against the ref and hash the caller holds, wherever it sits, so a record's `path` says where a copy was, never what it is.

- **Opened by ref.** `OpenAs(dir, ref)` opens the bundle in `dir` and requires its `recipe.md` name and version to equal `ref`'s, whatever the folder is called. `Open(dir)` requires a `name@N` folder, refusing (exit 1) any other, and checks it as `OpenAs(dir, base(dir))`. Every copy is opened with `OpenAs` and the ref the caller holds.
- **A drifted copy runs nothing.** When a sandbox record's entry has a `sha256` and the copy now hashes differently, `hyp sandbox remove` treats the copy as unreadable: its `cleanup` is `failed: no Cleanup ran`, a `sandbox:` warning names the ref, the path and both hashes, and remove goes on to its leftover check. `hyp sandbox drive` refuses the same copy with exit 1 and drives nothing. A record without `sha256` (ref `none`) has no copy to compare.
- **`hyp recipe hash DIR` still hashes any folder.** For a folder without `@`, `recipe.md`'s name must equal the folder's and its version is not checked. `HashOnly` is unchanged.
- **Rename hint.** For a pin to `x@N` that reports `now missing`, `hyp recipe check` looks for `x/` in each recipes directory the store searches. When that folder hashes as the pin, it prints, once per folder, after the first drift line the folder explains: `hint: <rel>/ holds x@N (same hash) in a folder whose name has no version; rename it: git mv <rel> <rel>@N`. When every failure has a hint, the closing refusal asks for the renames. `Load`, `Scan` and `hyp check` do not resolve such a folder.

## Considered Options

- **Move the copy to `recipes/<name>@<N>/`.** Rejected: records written before the move still name `recipes/<name>/`, so the fallback it was meant to remove would stay, and it buys only a visible name.
- **Move the copy and open it by ref.** Rejected: the same gain as opening by ref, without the move.
- **Warn and still run Cleanup from a drifted copy.** Rejected: it runs code the record does not vouch for.

## Consequences

- `hyp sandbox drive` and `hyp sandbox remove` both refuse to run recipe code from a copy whose hash differs from the record: drive exits 1, and remove runs no Cleanup and continues.
- Nothing moves. The copy stays at `recipes/<name>/` in a sandbox and `<tmp>/<name>/` elsewhere, so the text of frozen bundles about the copy path, and old sandbox records, stay true.
- The library lookup (`Scan`, `Dirs`, `Hashes`, `SingleFile`) is unchanged, and `hyp check` gives no warning for a leftover folder.
- A bundle kept outside the repository would be verifiable by its ref and hash, but nothing is built for it.
- `RECIPE-FORMAT.md`, the spec §7.6, the glossary and the `sandbox` package comment cite this decision. ADR 0001 to 0004 stand.
