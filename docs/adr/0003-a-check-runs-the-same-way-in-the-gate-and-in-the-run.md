---
status: accepted
---

# A check runs the same way in the gate and in the run

ADR 0002 says a bundle's checks, its Evidence and Standalone commands, are pure functions of frozen input files, and that the known-answer gate proves it before every run. A re-check of the session-replay recipe against ADR 0002 (findings N1 to N5) showed that hyp ran a check in two different ways, plus a third way for a case's git clone:

| | environment | `hyp` on PATH | working directory | product compared or kept |
|---|---|---|---|---|
| known-answer case | `PATH`, `HOME`, `TMPDIR` only | a copy of the running hyp | an empty `$CASE/cwd` | exit + stdout |
| the case's `git clone` | the host's, git config included | the host's | the caller's | none |
| `hyp recipe run` and `evidence` | the host's | the host's | the caller's | exit + files written anywhere |

Each finding sits where two rows differ. With a stub `hyp` first on the caller's `PATH`, a run passed the gate, recorded the gate's `hyp_sha256`, and then ran the stub (N1). A case compared stdout only, so extract-exp lost the content of `item.md` while its case still passed (N2). classify-pg's verdict changed with `HOME`, which the gate held fixed and the run did not (N3). The double run proves that a check is deterministic within one environment, never that it reads only its declared inputs. The one honest claim hyp can make is parity: what the gate saw is what the run does. We decided to make that hold by construction. Owner decision on 2026-10-30.

## Decision

- **Two ways to run a command.** A check (an Evidence or Standalone command) runs as described here, both in its known-answer cases and in `hyp recipe run` and `hyp recipe evidence`. Sandbox steps (Launch, Doctor, Drive and Cleanup) run as before, with the caller's environment and the sandbox's variables.
- **One environment.** One function builds a check's whole environment: `PATH` (a directory holding a copy of the running hyp named `hyp`, then the host `PATH`), a scratch `HOME` and `TMPDIR`, `GIT_CONFIG_NOSYSTEM=1` and `HYP_OUT`. Nothing else is inherited, and stdin is `/dev/null`. A case's `git clone` of its `repo` gets the same environment.
- **One working directory.** A check runs in an empty directory of its own, in the gate and in the run. A Standalone command's ARGS are passed as given, so a path in them must be absolute. hyp still resolves its own paths, `--record DIR` and Evidence's OUT_DIR, against the caller's directory.
- **One product.** A check's product is its exit code, its stdout and the tree it leaves under `$HYP_OUT`, an empty directory hyp gives each check. A case may hold `expect/` beside `input/`. Its out tree must equal `expect/` (paths and bytes, with the case directory's path in file contents written `$CASE`), and a case without `expect/` must leave the out tree empty. The double run compares the out trees too. The run keeps the tree in `DIR/out/` and stdout in `DIR/stdout` (Standalone), or `DIR/out/<n>/` and `DIR/stdout/<n>` (Evidence command `n`, from 0), even when the check fails.
- **One read of the executable.** `hyp recipe run` and `evidence` read the running hyp once. The gate's cases and the run's commands find that one copy first on `PATH`, so `known_answers.hyp_sha256` names the bytes the commands ran.

## Considered Options

- **Keep comparing stdout only, and say so.** Rejected because a check whose product is a file, such as extract-exp's `item.md` read by a judge, would stay unchecked, and the run would keep files the gate never saw.
- **Document the differences as limits.** Rejected because a passed gate would then say nothing about the run that follows it, which is the gate's whole purpose.
- **Pass the host environment with only the hyp copy put first on `PATH`.** Rejected because `HOME`, git configuration and every other variable would still reach the run and not the case: N3 would stay.
- **Run checks in the caller's working directory, in the case too.** Rejected because a case's cwd would then depend on where the operator stands, so a case could pass in one place and fail in another.
- **An OS sandbox.** Rejected: `sandbox-exec` is deprecated on macOS, `bwrap` needs user namespaces, and it would add a third isolation concept beside the spec's "a sandbox is a directory plus its record".

## Consequences

- A passed gate now carries over to the run, except for the limits: reads of fixed absolute paths, the network, the versions of tools on the host `PATH` after hyp (the Python that runs a `#!/usr/bin/env python3` script among them), and writes outside `$HYP_OUT`. Environment variables leave the list of limits.
- A Standalone command given a relative path as an ARG no longer finds it. Recipes take absolute paths.
- A check writes its product to `$HYP_OUT`, not to a path passed as an argument, when it wants the gate to compare it and the record to keep it.
- `command.json` keeps its shape. The record directory gains `out/` and `stdout`.
- Out of scope: `hyp sandbox create`'s clone of the host repository, which is a sandbox step, and a recipe's own executor that runs commands inside a check, such as session-replay's `unit.py`, which builds its own environment.
- `RECIPE-FORMAT.md`, the spec and the glossary change to match. Each change is a correction that cites this decision, made in the same change as the binary behaviour it describes. ADR 0002 stands, refined by this decision for how a check runs.
