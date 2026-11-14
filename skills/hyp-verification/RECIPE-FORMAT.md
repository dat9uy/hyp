---
name: hyp-recipe-format
description: The format of bundles used by hyp Experiments, recipes (recipes/<name>@<version>/) and rubrics (rubrics/<name>@<version>/), and where hyp looks for them.
---

# Recipe format

A recipe is a bundle directory `recipes/<name>@<version>/` holding
`recipe.md` and the scripts it runs. Each version is its own full
directory, so a bump adds one directory and never edits an old one. A
single Markdown file `recipes/<name>.md` is not a recipe, and neither is a
directory `recipes/<name>/` without a version: no lookup loads it or pins it. Only
`hyp bundle check` names it, in a rename hint when it hashes as a pin
whose bundle is missing.
Start a new recipe with `hyp bundle new NAME@1` (see To make a first
version, under Changing a recipe).

A recipe measures and never decides: its checks print measures, one named
line about one subject such as `S1 fail` or `stored rows 4`. How measures
become verdicts is a **rubric**, a bundle in this same format kept in
`rubrics/<name>@<version>/` (see Rubrics). `hyp bundle` handles both kinds,
and a name is used in one library only.

This file says when hyp runs a recipe's commands, and whether a
sandbox exists then.

## Frontmatter

```text
---
name: <name>        # equals the bundle directory name before the @
version: <integer>  # a new version for any byte change to a frozen version; a versioned bundle's directory name ends in @<version>
from: <name>@<N> <sha256>  # written by hyp bundle new: the version this one was copied from and its hash; absent in a first version
format: 1  # the one recipe format; every use of a bundle without it is refused
subject: <what is under test, e.g. "Rails app with seed data">
---
```

`hyp add exp --recipe <name>@<version>` refuses the experiment when the
recipe's `name` or `version` does not match, and stores the recipe's hash (see
Hash below). When the recipe later changes, every Result recorded under the old
hash is flagged questionable. Statuses are not changed; the flag tells the reader the
evidence came from a different procedure. `hyp record` refuses a new Result for an
Experiment whose recipe no longer hashes as recorded, or no longer loads; restore
the recipe, or add a new Experiment or a new version.

## Versions and lookup

`hyp add exp --recipe <name>@<version>` looks for the
ref in `recipes/` beside `.tree/`, then in `recipes/` of each parent directory
up to the repository root (the first directory holding `.git`). Outside a git
repository only `recipes/` beside `.tree/` is searched. So a repo-root
library serves every tree in the repository, and a tree needs no copy.
`hyp sandbox create` takes the library from `--recipes DIR` (the directory
that holds `recipes/`), which it requires unless `--recipe none`.

- In each directory, `recipes/<name>@<version>/` is the copy of the ref.
- Every copy found must hash the same. Otherwise hyp refuses the ref and
  names both paths: `recipe x@1: A and B differ; one ref must hash the same
  everywhere`. The same hash at two levels is fine; the nearest copy is used.
- A missing ref is refused naming every directory searched.
- The questionable flag is keyed by ref. A Result recorded under `x@1` stays
  clean when `x@2` is added, and is flagged only when the `x@1` it cites no
  longer hashes the same.
- `hyp check` prints one `warn:` line for each recipe left out of the
  hashes: a bundle that cannot be hashed (a symlink, a file name outside
  `[A-Za-z0-9._-]`, or a name or version that does not match its
  directory) and a ref whose copies differ. Results
  under it read as questionable until it is fixed. It also warns about a
  bundle that hashes but does not open (with the reason, for example a
  script that lost its execute bit); its pins still verify.
- A Result may cite a file inside a recipe as
  `hyp record --evidence recipes/<name>@<version>/<file>`, for example a
  known-answer fixture. hyp looks the ref up the same way, so a tree with no
  `recipes/` of its own can cite the repo-root library. The pointer is
  stored as written, and the Result records the recipe's hash beside it
  (`evidence_hashes`), so an in-place edit of that recipe flags the Result
  questionable like a change to its Experiment's recipe; `hyp check` names
  the pointer and both hashes. A ref that resolves nowhere is refused naming every
  directory searched, copies that differ are refused as above, and a file
  missing from the recipe is refused naming the recipe's directory. A
  directory is refused in every evidence form; cite a file in it. A path
  `recipes/<name>/...` without a version still means the store's own
  `recipes/`, and is not pinned by any recipe hash.

## Changing a recipe

Where the changed bytes live decides what a change costs, never a
judgment made after a gap appears.

- Inside a frozen bundle, any byte change in a file the hash covers is a
  new version, whose known answers hyp runs before every run. There is no
  diff-based exemption: a one-line fix in a helper costs what a rewrite costs.
  A Draft, a version no tree pins yet, may be edited in place, after which
  `hyp bundle known-answers` must pass again.
- Outside a bundle, inputs (snapshots, a live run, a store, the operator
  hyp) are pinned by hash in kept files, and setup (building the
  operator hyp) is never cited. Changing them costs no version.
- A recorded limit discloses a gap. It never replaces a version bump.

To make a first version:

1. Run `hyp bundle new <name>@1` in the tree. It writes `recipes/<name>@1/`
   beside the nearest `.tree/` from a template inside hyp:
   `recipe.md` with `name:`, `version: 1` and `format: 1` filled in and no
   `from:` line, guidance under Launch, Doctor, Drive, Evidence, Standalone
   and Cleanup, and one Standalone command `Run check: ./check.sh` with
   `External: TODO` under it; a stub `check.sh` that always fails; and
   `known-answers/must-pass/` and `known-answers/must-fail/`, each with a
   file in `input/`. It prints the new directory. Every line to replace
   holds the marker `hyp:placeholder`.
2. Fill it: replace each `hyp:placeholder` line and the subject, write the
   commands and their scripts, state each `External:` line, and give every
   command a must-pass and a must-fail case (see Known answers). Delete
   the guidance of a section you leave without commands.
3. Run `hyp bundle known-answers <name>@1` until it passes. Unedited, it
   fails: the stub fails the must-pass case.
4. Run `hyp add exp --recipe <name>@1`. It pins the bundle's hash, so the
   version is frozen from then on.

`hyp bundle new <name>@1` refuses (exit 1) a name that has any version in
a library the lookup searches, naming it: make the next version from it
with `--from` instead. `<name>@<N>` above 1 without `--from` is a bad
argument (exit 2), and so is a bad ref; outside a store it exits 3, as
other store commands do. `hyp add exp --recipe` refuses (exit 1) a bundle
that still holds `hyp:placeholder` in any file outside a case's `input/`,
naming the files, since a case input may be a store that holds the marker.

To make a next version:

1. Run `hyp bundle new <name>@<N+1> --from <name>@<N>`. It copies the
   bundle `recipes/<name>@<N>/` byte for byte, without dot-entries and
   `__pycache__`, to `recipes/<name>@<N+1>/` in the same library, sets
   `version:` and writes `from: <name>@<N> <sha256>` and `format: 1` into
   the copy's frontmatter. It prints the new directory and the `from:` line. It
   refuses (exit 1) a target that exists, a target that is not one
   above `--from`, and a higher version in any library the lookup
   searches, so a new version always starts from the newest. A bad ref and
   a `--from` that does not resolve exit 2.
2. Edit the copy. `diff -r recipes/<name>@<N> recipes/<name>@<N+1>`
   shows what changed since `from:`.
3. Add or update its known answers; `hyp bundle known-answers REF` runs
   them, and hyp runs them again before every run (see Known answers).

The rule is held by a refusal and a check:

- `hyp add exp --recipe <name>@<N>` refuses (exit 1) an instrument, a
  bundle with an Evidence or Standalone command, that has no
  `known-answers/NAME/` directory (the gate reads each case's contract
  when it runs). Every use of a bundle whose `recipe.md` has no
  `format: 1` line is refused too (see Self-containment).
- `hyp bundle check`, run from anywhere in the repository, finds every
  `.tree` under the repository root (the root the lookup stops at),
  skipping nested repositories and worktrees. It compares every pin to
  a versioned bundle, the Experiment's recipe and each Result's
  `evidence_hashes`, with the bundle on disk. Each mismatch prints
  `drift: <tree> <id> <file>: recorded <hash>, now <hash|missing>` and
  the command exits 1; a store that does not load prints an `error:`
  line and also exits 1; outside a git repository it exits 3. The last
  line is `recipe check: <T> trees, <P> pins, <D> drift`. Before it,
  each versioned bundle gets a `frozen` line when a tree pins its
  current hash and a `draft` line when none does. Then, for each
  instrument, it prints
  `warn: <dir>: instrument with no known answers`, or
  `warn: <dir> <command>: no must-pass case` and `no must-fail case`
  for each Evidence command (as a row shows it, such as `./check.sh`) or
  Standalone label that has no case expecting exit 0, or none expecting
  a non-zero exit. For a rubric (see Rubrics) it instead prints what
  `hyp add exp --rubric` would refuse; for a frozen recipe whose checks
  print verdicts it also prints `warn: <dir> <command>: verdict <label>: no supports case` or
  `no contradicts case` for each label no case prints so,
  `warn: <dir> <command>: case <name>: malformed verdict line "<line>"`
  and `warn: <dir> <command>: case <name>: label <label> printed twice`.
  Warnings never change the exit code. When a pin to
  `x@N` is `now missing` and a leftover `recipes/x/` hashes as the pin,
  the first drift line it explains is followed by
  `hint: <dir>/ holds x@N (same hash) in a folder whose name has no version; rename it: git mv <dir> <dir>@N`,
  once per folder; when every failure has a hint, the closing refusal
  asks for the renames. A stray file in a pinned bundle shows
  as drift (dot-entries and `__pycache__` are not hashed). Run it before
  opening any pull
  request that touches `recipes/`.
  Drift means a cited version was edited in place: restore it and
  make the change as a new version.
- A `.go` file in a bundle breaks `go vet ./...` in a Go repository,
  because the go tool reads `recipes/<name>@<N>/` as a package outside
  the main module. When the repository root holds `go.mod`,
  `hyp bundle check` prints
  `warn: <dir>/<file> is a .go file, and go.mod at the repository root makes go vet ./... read this bundle as a package outside the main module; give it a non-.go name, such as <file name>.txt`
  for each `.go` file in each versioned bundle, and `hyp add exp --recipe`
  prints the same line after `warn: recipe <ref>: ` for the bundle it
  pins. Both still exit 0. Keep Go source in a bundle under a non-`.go`
  name, such as `probe_test.go.txt`.

Per-store `hyp check` warns about questionable
Results in its own store and exits 0.

## Sections

Five level-2 headings, in this order, one for each step,
each saying what to do and how to tell it
worked:

1. `## Launch`: how to start the subject, with exact commands and ports.
2. `## Doctor`: the health checks that must pass before driving.
3. `## Drive`: the steps of the experiment.
4. `## Evidence`: which files to save under `.tree/evidence/<E-id>/`, by
   hand while the sandbox exists, or by a command that reads the kept
   files after remove (scripted Evidence).
5. `## Cleanup`: how to stop what Launch started.

The order of the headings is the order to read them in, not the order hyp
runs them: scripted Evidence runs after Cleanup, because it reads only the
files that remove keeps.

A sixth, optional heading, `## Standalone`, holds commands that run with no
sandbox, for example to capture an Experiment's inputs or to classify
kept files. It is not a step. Each of its commands is named:
`Run NAME: ./script WORDS`, with NAME matching `[a-z][a-z0-9-]*` and used
once.

Each Evidence and Standalone command states what it reaches outside the
bundle on the line directly under its `Run:` line:

```markdown
Run score: ./score.py
External: OUT_DIR of hyp sandbox remove (the subject); writes only $HYP_OUT
```

Name the subject arguments and any outside reads or writes, or write
`External: none`. `hyp bundle check` warns about a command with no such
line, or with `External: TODO`, which `hyp bundle new` writes under each
command of a copy that lacks the line and under a first version's check; the exit code does not change, and
a bundle without the line still loads and records.

## Hash

A bundle's hash covers every file in the
directory except entries whose name starts with `.` and `__pycache__`
directories: one line `<sha256 of the file>  <path>` per file, sorted by path,
hashed again. The directory name is not hashed, so a copy of recipes/x@1/ at recipes/x/ hashes the same.
hyp opens each copy it makes by the ref it was copied for: the copy's
`recipe.md` must give that name and version, wherever the copy sits. The shell equivalent:

```bash
(cd recipes/NAME@VERSION && find . \( -name '.*' ! -name . -o -name __pycache__ \) -prune -o -type f -print \
  | sed 's|^\./||' | LC_ALL=C sort \
  | while read -r f; do printf '%s  %s\n' "$(shasum -a 256 "$f" | cut -d' ' -f1)" "$f"; done \
  | shasum -a 256 | cut -d' ' -f1)
```

File and directory names use only `A-Z a-z 0-9 . _ -`. Symlinks are refused.
The executable bit is not hashed. A bundle holds scripts and small
inputs, not data dumps, because hyp hashes it on every read.

`hyp bundle hash DIR` prints the hash of the bundle in `DIR`, the same value
`hyp add exp --recipe` records for it. It exits 2 for a directory with no
`recipe.md`, and 1 for a `recipe.md` that does not match the directory. A `DIR` whose name has no `@<N>` hashes too: its
`recipe.md` name must match the directory and its version is not checked. `--json` prints `ref`, `name`, `version`, `path` and `sha256`. A recipe script that needs its own hash, for example to tell
whether its copy changed, runs this instead of hashing the files itself.

## Runnable commands

In a bundle, a line under `## Launch`, `## Doctor`, `## Drive`,
`## Evidence` or `## Cleanup` that starts with `Run:` is a command, and so
is a `Run NAME:` line under `## Standalone`.
`hyp sandbox create --recipe NAME@VERSION` runs Launch, then Doctor;
`hyp sandbox drive ID` runs Drive; `hyp sandbox remove` runs Cleanup; after
remove, `hyp bundle evidence --record DIR REF OUT_DIR` runs Evidence; and
`hyp bundle run --record DIR REF NAME ARGS` runs one Standalone command.
`--recipe` is required on `hyp sandbox create`: it chooses which recipe,
and the recipe's `Run:` lines decide what runs. `--recipe none` sets
nothing up. `--recipe` repeats to stack recipes (see Stacking recipes
below). The rest of each section stays prose for the agent.

Rules:

- A command is one script inside the bundle, as a relative path containing
  `/` (for example `./launch.sh`), then plain words.
- No shell runs the line, so quotes, `$VARS`, pipes, redirects and globs are
  refused. Put them in the script.
- Interpreters go in the script's shebang.
- `Run:` lines inside code fences are ignored. A `Run:` line under any
  other heading is refused, and so is a plain `Run:` line under Standalone
  and a named one anywhere else.
- `hyp sandbox create` copies the recipe to `~/.hyp-sandbox/ID/recipes/<name>/`
  outside `wt/`, so the agent working in `wt/`
  does not find it there. The copy is opened by its ref. Launch, Doctor, Drive and Cleanup commands run
  with that copy as the working directory and get
  `HYP_SANDBOX_ID`, `HYP_SANDBOX_WT`, `HYP_SANDBOX_TMP`,
  `HYP_SANDBOX_RECIPE_TMP` (the recipe's own `$HYP_SANDBOX_TMP/<name>`,
  made empty at create), `HYP_SANDBOX_INPUTS` and `HYP_SANDBOX_PATCHES`
  when given (see Per-run inputs below), and a `PATH` that starts with the sandbox's pinned
  `hyp`. They also inherit the environment of the process that runs hyp,
  which is how a run passes its inputs (see Per-run inputs below).
- An Experiment recorded inside `wt/` resolves recipes only against what the
  host has at the base commit; the sandbox's copy is not on its lookup path.
- Doctor passes when every command exits 0. Launch and Doctor stop at the
  first command that fails, and together have 10 minutes for each recipe.
- Every Cleanup command runs, even after an earlier one fails, and each has
  its own 2 minutes. Remove prints one warning for each recipe with a failed
  Cleanup command, naming each command's outcome, for example
  `sandbox: scratch-pg@1 cleanup: ./stop1.sh exited 7; ./stop2.sh ok; see recipe.log`.
- A Cleanup must wait, within its 2 minutes, for the processes it stops to
  exit (`pg_ctl stop` waits by default). A Cleanup that sends a kill and
  returns at once can still be running when the leftover check runs, so
  remove refuses and has to be run again.
- A Cleanup must also succeed when run twice: a refused remove runs it again
  on the next attempt, after it has already stopped what it could.
- A Cleanup must also succeed when its Launch never ran, failed before it
  wrote its state, or was stopped partway. When hyp was killed during a
  command, remove stops that command's process group before any Cleanup runs,
  so a Cleanup never races a Launch that outlived its hyp. create runs the Cleanup of the recipe whose Launch or
  Doctor failed, and removing a sandbox kept after a failed setup runs every
  recipe's Cleanup, including recipes after the failed one, which never
  launched. Such a Cleanup finds no state of its own under
  `$HYP_SANDBOX_RECIPE_TMP` and exits 0.
- Output goes to `recipe.log` beside `sandbox.json`, each command's after a
  `== REF step: command` line (for example `== scratch-pg@1 launch: ./launch.sh`),
  and remove copies it to OUT_DIR.
- A recipe never writes `sandbox.json`. It keeps its own state in its own
  file under `$HYP_SANDBOX_RECIPE_TMP`, and the caller keeps that file with
  `hyp sandbox remove ID OUT_DIR --keep tmp/<name>/<file>`. Kept files
  land in OUT_DIR by basename, so stacked recipes give their kept files
  distinct names.
- The sandbox owns no port or socket. A recipe that needs one sets it up in
  Launch, checks it in Doctor and frees it in Cleanup. Prefer a Unix socket
  under `$HYP_SANDBOX_TMP` to a TCP port, as the Example below does.
- Launch must start servers in the background (daemonised, for example
  `pg_ctl -w start -l LOG`), and Cleanup must stop everything Launch started:
  remove (and create, after a failed Launch or Doctor) deletes nothing while
  any process other than hyp has its working directory in the sandbox,
  whatever it listens on. It exits 3 and lists each such process. Your own
  shell or editor counts when its working directory is in `wt/`: leave the
  sandbox before you remove it. When create cannot run that check after a
  failed setup, it keeps the sandbox and says it could not check.

### Drive, Evidence and Standalone

- `hyp sandbox drive ID` runs every recipe's Drive commands in stack order,
  like Launch (the recipe's copy, the sandbox's variables), and stops at
  the first that fails. It takes no arguments from the operator; a Drive
  reads its per-run values from the inputs file or from the state its
  Launch wrote under `$HYP_SANDBOX_RECIPE_TMP`. Run it again to drive
  again. It exits 2 for a sandbox with no Drive command and 3 for one that
  cannot be driven, each refusal naming `hyp sandbox remove`: hyp was killed
  before setup finished, setup failed, a remove already started, or a command
  whose hyp was killed still runs (remove stops it; once it has ended, drive
  works again). A directory with no record exits 3 and names
  `hyp sandbox remove ID --discard`; `hyp sandbox remove ID OUT_DIR` on it
  exits 2, since there is nothing to keep. A drive, a remove or a create of a sandbox another hyp
  command is changing exits 3, naming that command, and changes nothing.
- `hyp bundle evidence --record DIR REF OUT_DIR` runs REF's Evidence
  commands after remove, with OUT_DIR's absolute path appended as the last
  argument, and stops at the first that fails. Run it again, into a new
  DIR, to re-run the check, also from a later version of the recipe.
  `hyp sandbox remove` never runs Evidence. OUT_DIR is always the OUT_DIR
  of `hyp sandbox remove`, holding its `sandbox.json`. For files no
  sandbox wrote, write the check as a Standalone command and give it their
  path with `hyp bundle run`.
- `hyp bundle run --record DIR REF NAME ARGS` runs the Standalone command
  NAME with ARGS appended after its fixed words. Flags go before REF;
  everything after NAME is the script's. A word, or the VALUE of a
  `KEY=VALUE` word, written as a path (absolute, or relative with a `/`,
  read against your working directory) that names a regular file is a file
  argument: hyp keeps a copy beside the record as `inputs/N/<base name>`
  (N counts ARGS from 1; not kept over 1 MB, but still hashed), and the
  script gets another copy, never your file. One that names a directory is
  passed as an absolute path. A bare word, such as `today`, is passed as
  given even when a file of that name exists, so write `./today` to pass
  that file.
- Evidence and Standalone commands are checks, and run exactly as their
  known-answer cases do: from a temporary copy of the bundle,
  deleted afterwards, in an empty working directory of their own, with
  stdin `/dev/null` and only this environment: `PATH` (a directory holding
  the copy of the running hyp the gate used, named `hyp`, then the
  caller's `PATH`), a scratch `HOME` and `TMPDIR`, `GIT_CONFIG_NOSYSTEM=1`
  and `HYP_OUT`, an empty directory for the check's files. The copy and
  the check's directory have the same paths in a known-answer case and in
  a run, apart from the random digits in their names, so a script cannot
  tell the gate from the run. A hyp killed while it runs leaves the copy
  (`hyp-bundle-*` in the system temp directory); hyp does not clean it up,
  and you may delete it. A script finds its own files from its own path
  and never relies on where the library is.
- The check's `TMPDIR`, `HOME` and working directory are deleted when it
  ends, also when it fails. Send what explains a failure, such as a
  server's log, to stderr: hyp keeps it in `command.log` (or
  `known-answers.log` for a case) and prints its last lines with the
  error. A file the record must keep goes under `$HYP_OUT`.
- Timeouts: each Drive command 60 minutes, each Evidence command 10
  minutes, a Standalone command 60 minutes.
- When the script exits non-zero, hyp exits with its code unchanged; a
  timeout or a signal exits 3. Scripted verdicts exit 10 or above (30 by
  convention), so the code alone tells them from hyp's 1, 2 and 3.
- A ref that does not load, an unknown NAME, a recipe with no Evidence
  command, an OUT_DIR without `sandbox.json`, an existing DIR, a DIR that is
  or is inside a `recipes/` directory the lookup searches and, for Evidence,
  a DIR inside OUT_DIR exit 2, before anything runs. Both location checks see through symlinks and other
  spellings of the same directory. A DIR whose path cannot be resolved
  exits 3, before anything runs.

## Per-run inputs

A `Run:` line under Launch, Doctor, Drive or Cleanup takes no arguments
from the run, so a run passes its inputs (a snapshot, a live-run directory)
in a file. A Standalone command takes its arguments from
`hyp bundle run`.

1. The caller writes the inputs to a file and runs
   `hyp sandbox create ... --inputs FILE`. hyp copies the file to
   `inputs/<base name>` in the sandbox and records the base name and the
   copy's sha256 under `inputs` in `sandbox.json`. The recipe's `recipe.md`
   names the fields it reads under Launch.
2. Every sandbox command, Launch, Doctor, Drive and Cleanup, gets the copy's
   path in `HYP_SANDBOX_INPUTS`, including Drive and Cleanup, which run in
   later processes.
3. `hyp sandbox remove` copies it to `OUT_DIR/inputs/<base name>`, so a
   Result that depends on the inputs cites the kept copy. `inputs`,
   `patches`, `changes`, `changes.json`, `tree`, `sandbox.json`,
   `recipe.log` and `known-answers.log` are reserved names in OUT_DIR for
   `--keep`; remove keeps `wt/.tree` as `OUT_DIR/tree`.
4. A throwaway Arm, a change over the base that never lands in code, is
   one patch: `hyp sandbox create ... --patch ARM=FILE`, repeatable, at
   most one per Arm (ARM is `[a-z][a-z0-9-]{0,39}`). A
   change that lands is a commit, given as `--base` instead. hyp refuses a
   patch that does not apply to the base with exit 2 before any command
   runs, and leaves no sandbox. It applies nothing to `wt/`: it copies
   each patch to `patches/<ARM>.patch`, passes that directory to every
   command as `HYP_SANDBOX_PATCHES`, and Launch builds each Arm from it.
   `sandbox.json` names each by `arm` and `sha256` under `patches`, and
   remove keeps it as `OUT_DIR/patches/<ARM>.patch`.

A recipe reads its inputs only from `HYP_SANDBOX_INPUTS`.

## A store cut

A replay starts from a hyp store as it was at one event.
`hyp sandbox create ID ... --store SRC --cut EVENT` writes that store to
`wt/.tree` before any recipe's Launch runs, so no recipe copies events.

- The cut is the replay's own: hyp orders SRC's events as replay does, by
  id and then by actor, whatever the line order inside a file, and keeps
  every event up to and including the first with id `EVENT`, each line as it
  was written. Each events file of the cut holds its kept lines in id order,
  so `hyp status` in `wt/` equals `hyp status --as-of EVENT` in SRC.
- `wt/.tree/evidence` is SRC's evidence without the `E-*` folders of
  Experiments added after the cut. No `index.sqlite` is copied: hyp keeps
  no index.
  An actor with no event before the cut gets no events file. A symlink
  under SRC's evidence is refused (exit 3, and the sandbox is removed).
- The cut replaces any `.tree` the base commit holds at its root: `wt/.tree`
  holds the cut and no other file of the base's `.tree`.
- `--store` and `--cut` go together. An `EVENT` that SRC does not hold and a
  SRC with no `.tree/events` are refused with exit 2 before anything is
  made. An events line that does
  not parse exits 3; run `hyp check` in SRC.
- The record names the cut: `store` (SRC as `<host>/REL` when it is inside
  the host, an absolute path otherwise) and `cut`.

## Stacking recipes

`hyp sandbox create ID ... --recipe scratch-pg@3 --recipe harness@1`
stacks independent recipes in one sandbox: environment recipes first, the
harness recipe last. A stack is whole bundles, not one-job recipes.

- **Order.** Recipes run in the order given. Each recipe runs its Launch
  and then its Doctor, under its own 10 minutes, before the next recipe's
  Launch starts.
- **Refusals.** A name given twice (at any version) and `none` together
  with any other ref are refused with exit 2 before anything is made.
- **A failed setup.** When a recipe's Launch or Doctor fails or times out,
  create runs the Cleanup of that recipe and then of every recipe before it,
  last first. Recipes after it are never launched. Create then deletes the
  sandbox only when no process runs in it, as for one recipe.
- **Remove.** Remove runs every recipe's Cleanup, the last recipe first.
  Each failure is reported and never skips another recipe's Cleanup. The
  leftover-process check then decides whether the sandbox is deleted.
- **Per-recipe tmp.** Each recipe gets its own empty
  `$HYP_SANDBOX_TMP/<name>/`, exported as `HYP_SANDBOX_RECIPE_TMP`. Write
  state there. A file a recipe writes directly in `$HYP_SANDBOX_TMP` must
  not use another recipe's name.
- **No dependencies.** Recipes declare none, and hyp runs no combined
  Doctor. A recipe that reads another recipe's files (a socket, a state
  file) names them in its own `recipe.md`, and its own Doctor checks them.
  hyp does not verify data passed between recipes.
- **One pinned ref.** An Experiment cites one recipe, the harness recipe:
  `hyp add exp --recipe harness@1`. The environment recipes' hashes
  are provenance: `sandbox.json` lists every recipe with its hash in stack
  order, and the Result's evidence points at `OUT_DIR/sandbox.json`. A new
  version of an environment recipe never flags a Result questionable.

## The sandbox record

`hyp sandbox create` writes `~/.hyp-sandbox/ID/sandbox.json`, and
`hyp sandbox remove ID OUT_DIR` copies it to `OUT_DIR/sandbox.json` with
`removed` added. Only hyp writes it; recipe scripts read it and never write
it. A recipe keeps its own state under `$HYP_SANDBOX_RECIPE_TMP`. Top-level
keys are written in sorted order. hyp writes the record through a temporary
file and a rename, so a reader never sees part of one, and writes each row
of `commands` before its command starts. A hyp killed between the two
leaves a `.sandbox.json-*` file that nothing reads. The sandbox also holds
`sandbox.lock`, held by the one hyp command (create, drive or remove) that
is changing the sandbox; the lock goes away when that hyp exits.

| Field | Value |
|---|---|
| `format` | the record's shape number, now `4` (`3` had no `removing`; `2` also held the absolute `wt`, `tmp`, `bin`, `inputs.source`, `inputs.path` and `recipes[].path`, and an absolute `host` and `store`; `1` had no `inputs` and no `pgid`); absent (read as `0`) in a record written before it. A reader compares it before reading the rest. A format-2 sandbox still drives and removes, and its old fields are copied to OUT_DIR as read |
| `id` | the sandbox id; the sandbox is `~/.hyp-sandbox/ID`, and its clone, scratch directory, `bin/` and copies are found from it, not from the record. A script finds them in the `HYP_SANDBOX_` variables |
| `host` | the host repository's directory name, not its path |
| `base` | the full commit `wt/` is detached at |
| `patches` | with `--patch ARM=FILE`: one `{"arm", "sha256"}` per patch, in the order given, the sha256 of the sandbox's copy `patches/<ARM>.patch`; absent otherwise |
| `inputs` | with `--inputs FILE`: `name` (the base name of the sandbox's copy, `HYP_SANDBOX_INPUTS`, kept as `OUT_DIR/inputs/<name>`) and `sha256` (of the copy); absent otherwise |
| `hyp_sha256`, `hyp_version` | hash of `bin/hyp` and what its `version` printed |
| `recipes` | the sandbox's recipes in stack order, always a list (see below) |
| `commands` | each command hyp ran in the sandbox, in order (see below); absent when none ran |
| `store`, `cut` | with `--store` and `--cut`: the store cut into `wt/.tree`, written `<host>/REL` when it is inside the host and as an absolute path otherwise, and the cut event's id; absent otherwise |
| `created`, `removed` | UTC times, RFC 3339; `removed` only in OUT_DIR |
| `removing` | UTC time, RFC 3339: a remove passed its refusals and is stopping the sandbox. Once set, drive refuses and remove can run again; a sandbox kept for a leftover process holds it, and OUT_DIR keeps it |

Each entry of `recipes`:

| Field | Value |
|---|---|
| `ref` | `NAME@VERSION` as passed to `--recipe`, or `none` |
| `name`, `version`, `sha256` | the recipe's name, version and hash (see Hash). The commands run from the sandbox's copy, `~/.hyp-sandbox/ID/recipes/<name>`. Remove and drive reopen it by `ref` and check it against `sha256`; a copy that hashes differently runs nothing: remove records `failed: no Cleanup ran` and warns with both hashes, and drive refuses with exit 1 |
| `setup` | `passed` (Launch and Doctor passed), `failed` (one failed; the sandbox was kept), or `none` (nothing to run). Written as each recipe's setup ends. Absent until then, so a recipe whose setup hyp was killed during has none, and absent for the ref `none` and for a recipe after the one that failed, which was never launched |
| `cleanup` | `ok`, or `failed: ` and each command that failed (`failed: ./stop1.sh`); `failed: no Cleanup ran` when the copy or `recipe.log` did not open. Absent when no Cleanup command ran. It is the outcome of the last Cleanup run |
| `known_answers` | the outcome of the recipe's known answers, run before its commands: `hyp_sha256` (the hyp that ran them) and `cases`, one `{"name", "outcome"}` per known-answer run. Absent for a recipe with no Evidence or Standalone command, which is not gated |

`--recipe none` records exactly `"recipes": [{"ref": "none"}]`.

Each entry of `commands` (a command record's rows have the same fields):

| Field | Value |
|---|---|
| `ref` | the recipe's ref |
| `step` | `launch`, `doctor`, `drive` or `cleanup`; in a command record `evidence` or `standalone` |
| `label` | a Standalone command's name; absent otherwise |
| `args` | the command as run: its script as `./PATH` inside the bundle, then every word |
| `exit` | the script's exit code; absent when it has none |
| `error` | why it has none (`timed out after ...`, `interrupted` for Ctrl-C, `signal: killed`, or a start failure); absent otherwise |
| `pgid` | the process group the command led (its process id); absent when it did not start |
| `started`, `ended` | UTC times, RFC 3339; `ended` absent when hyp stopped during the command (it was killed, or the machine stopped), and `exit` and `error` are then absent too |

A remove refused for a leftover process keeps the Cleanup rows it ran, and
each recipe's `cleanup`, in the sandbox's record.

A record written by an older hyp holds one `recipe` object instead of
`recipes`, and may hold `port`. Remove still reads it: the object becomes the
one entry of `recipes`, its Cleanup runs from the `path` it names (inside
`wt/recipes/` for those records), and every field hyp does not know, `port`
included, is kept in OUT_DIR as it was read.

## The command record

`hyp bundle evidence` and `hyp bundle run` make the new directory that
`--record DIR` names, holding `command.json`, `command.log` (each
command's stderr, then its stdout, after a `== REF step: command` line),
and each check's product: for a Standalone command, its stdout
in `stdout` and its `$HYP_OUT` tree in `out/`; for Evidence command `n`
(from 0, in `commands` order), `stdout/<n>` and `out/<n>/`. They are kept
even when the check fails. hyp reads `$HYP_OUT` by the rules a case
follows and never follows a symlink: a check that exited 0 but left a tree
a case would fail with (a symlink, a special or unreadable file, or no
directory) is refused with exit 1, keeping the rest of the tree; a check
that failed exits with its own code. A relative DIR is under the working
directory. hyp refuses a DIR that exists. It also refuses one that is or
is inside any `recipes/` directory the recipe lookup searches, even one that
does not exist yet, because a record there changes the hash of the bundle it
lands in and every pin to it would read as drift, and, for
`hyp bundle evidence`, one inside OUT_DIR, because the check reads OUT_DIR.
Put it under `.tree/evidence/<E-id>/` and cite its `command.json` with
`--evidence`. hyp writes `command.json` through a temporary file and a
rename, before each command starts and after it ends.

| Field | Value |
|---|---|
| `format` | the record's shape number, now `6` (`6` added `entries` to `dirs`; `5` added `dirs`; `4` replaced `arg_sha256` with `inputs`, and its `path`, `cwd` and directory arguments name no host path; `3` added `arg_sha256`; `2` added `pgid` to rows); absent (read as `0`) in a record written before it |
| `ref` | the recipe ref as given |
| `name`, `version`, `sha256` | the recipe's name, version and hash, the hash of what ran |
| `path` | the library bundle the temporary copy came from, as `<repo>/REL` (`<repo>` is the directory name of the git work tree that holds `cwd`, or of `cwd` outside one), or as given outside it |
| `step` | `evidence` or `standalone` |
| `cwd` | the caller's working directory, as `<repo>/REL`, where the recipe lookup and a relative `--record` start; each command ran in an empty directory of its own |
| `created` | UTC time, RFC 3339 |
| `commands` | one entry per command, with the fields of the sandbox record's `commands` |
| `known_answers` | as in the sandbox record's `recipes` |
| `inputs` | one entry per file argument, in ARGS order: `arg` (its place in ARGS, from 1), `path` (`inputs/<arg>/<base name>`, beside `command.json`), `sha256` and `size` (of the copy the script got) and `kept` (`false` over 1 MB, when no copy is at `path`). A row shows `path` in the argument's place, and a directory argument as `<repo>/REL`, or as given outside the repository, which hyp warns about. Absent when no argument names a file |
| `dirs` | one entry per directory argument, in ARGS order: `arg` (its place in ARGS, from 1), `path` (as the row shows it), `sha256` (of its tree just before the first command ran: each entry's slash path and kind, a file's sha256 and a symlink's target, dot-entries included, in lexical order) `files` (its regular files) and `entries` (each regular file's slash path under the directory and its sha256, absent when it holds more than 200). hyp passes the directory in place and keeps no copy. Absent when no argument names a directory |

## Changes

When create finishes, Launch and Doctor included, hyp writes
`manifest.json` in the sandbox: every path in `wt/` but `wt/.git` and the
`index.sqlite*` files older hyp binaries write under `wt/.tree`,
each with its size, mode, and its
sha256 (a regular file) or link target (a symlink). `hyp sandbox remove
ID OUT_DIR` lists the paths that changed since then in
`OUT_DIR/changes.json`:

| Field | Value |
|---|---|
| `format` | the file's shape number, now `1` |
| `baseline` | `manifest.json at create` |
| `changes` | one entry per changed path, sorted: `path`, `change` (`added`, `modified` or `deleted`), `size`, `mode`, `sha256` or `link`, and, for an added or modified regular file, `copied` |

A path is `modified` when its size, sha256, mode or link target differs.
remove copies each added or modified regular file up to 10 MB to
`OUT_DIR/changes/<path>` (`copied: true`); a larger one is listed with
`copied: false`. A sandbox whose setup did not finish, or that an older hyp made, has no
manifest: remove warns and writes no `changes.json`.

Limits: commits made inside `wt/.git` are not listed, only the files they
change in the work tree; a symlink is compared by its target only; a file
that vanishes while the list is made is left out.

## Known answers

A bundle with an Evidence or Standalone command is an instrument, and must
have known answers. For each mode it offers (each Evidence
command, Standalone label, or flag that changes what its checks expect),
write a must-pass case and a must-fail case, each with the exact exit code
and output it must give. A pass condition never admits the failure output
of the check it tests. When a case gives the wrong answer, the instrument
is wrong: make the next version (see Changing a recipe).

A recipe's cases never expect a verdict: `hyp add exp --recipe` refuses a
recipe when any case's expected `stdout` holds a line starting with
`verdict:`. A frozen recipe whose checks print verdicts stays valid for the
Experiments that already pin it; its next version splits into a recipe
and a rubric. A rubric's cases are its specification (see
Rubrics).

### Cases as data

A bundle can hold its known answers as data, and
`hyp bundle known-answers REF` runs them with no sandbox. Before
every run of an instrument, hyp runs them and refuses the run on any
failure (see The gate).

Each case is a directory `known-answers/NAME/`, NAME matching
`[a-z0-9][a-z0-9-]*`, holding `input/`, `expect.json` and, optionally,
`expect/`:

```json
{"step": "evidence", "command": "./check.py", "exit": 0, "stdout": "pass\n"}
```

- `step` is `evidence` or `standalone`.
- `command` is, for `evidence`, one Evidence `Run:` line as a command row
  shows it (the script as `./PATH`, then its fixed words), and, for
  `standalone`, a Standalone label.
- `args` (optional, `standalone` only) is appended to the command.
- `exit` and `stdout` are the exact exit code and stdout the case must give.
- `repo` (optional) names a git bundle file directly under `input/`. The
  bundle must hold `HEAD`, or its clone has no checkout: make it with
  `git bundle create FILE HEAD <branch>`. A bundle with no `HEAD` breaks
  the rules below.

`expect/` holds the files the command must leave under `$HYP_OUT`:
the same paths, directories included, and the same bytes,
with the case directory's path written `$CASE`. Without `expect/`, the
command must leave `$HYP_OUT` empty. `expect/` holds only directories and
regular files.

Everything under `input/` is hashed, dot-entries and `__pycache__` included,
since an input is often a store holding `.tree/`. A name below `input/` may
also hold `@`, so an input can be a store with its pinned
`recipes/<name>@<N>/`; every other name in a bundle keeps
`[A-Za-z0-9._-]`. An `input/sandbox.json`,
`input/command.json` or `input/changes.json` must carry this hyp's `format`
for that file.

hyp runs each case in a new case directory, `$CASE`. It copies `input/` to
`$CASE/input`, clones `repo` to `$CASE/repo` when it is given, and runs the
command as a check, the way `hyp bundle run` and `evidence` run it:
from a temporary copy of the bundle, in the empty `$CASE/cwd`,
with stdin `/dev/null` and only `PATH` (first, a directory holding a copy
of the running hyp named `hyp`), `HOME=$CASE/home`, `TMPDIR=$CASE/tmp`,
`GIT_CONFIG_NOSYSTEM=1` and `HYP_OUT=$CASE/out`. The clone of `repo` gets
the same environment, so host git configuration cannot change it. An
Evidence case gets `$CASE/input` as OUT_DIR. In `args`, the literal
`$CASE` becomes the case directory, and in stdout and in the files under
`$CASE/out` the case directory becomes `$CASE`. stderr is shown, not
compared. The command runs in its own process group, killed when it ends
or after 60 seconds.

So a check cannot build the binary it runs: a case has 60 seconds and a
scratch `HOME`, by design. Build real binaries before the run
and pass their paths to a Standalone command as arguments; its cases pass
small stand-in scripts under `input/` in their place.

Name each such binary by its sha256, never by the commit it was built
from. Rebuilding the same commit gives a different sha256 unless the build
is reproducible (Go: `go build -trimpath`), so a Result that cites the
commit does not say which binary ran.

Each case runs twice, each time in a new case directory, unless its first
run times out. Its outcome is the first of: `timed out after 1m0s` when a
run timed out; `out is not a directory` when a run removed or replaced
`$HYP_OUT`, `out holds a non-regular file: PATH` when it left a symlink or
other special file there, or `out holds an unreadable entry: PATH`;
`not stateless` when the
two runs differ in exit code, stdout or out tree; `exit N, want M` or
`stdout differs` when they disagree with `expect.json`;
`out differs: PATH` (the first differing path, sorted) when the out tree
does not equal `expect/`; `out not empty: PATH` when a case without
`expect/` left a file; and otherwise `ok`. One bundle's cases get 10 minutes in all.
`hyp bundle known-answers` prints `ok NAME` or `FAIL NAME: OUTCOME` per case,
and exits 0 when every case is ok, 1 when one failed, and 3 when one timed
out. It exits 1 before running anything for a case that breaks the rules
above, and for an instrument with no cases.

### The gate

Before `hyp sandbox create` (each stacked recipe in stack order, before
anything is made), `hyp bundle run` and `hyp bundle evidence`, hyp runs
the instrument's known answers as above, with a copy of the operator's
running hyp first on `PATH`, never the sandbox's `--hyp BIN`. hyp reads
the running hyp once: `hyp bundle run` and `evidence` run their commands
with the same copy, so `known_answers.hyp_sha256` names the hyp the
commands ran. A failed case
refuses the run (exit 1,
`recipe REF failed its known answers: NAME: OUTCOME; nothing ran`), a
timed-out case exits 3, and a refused create leaves no sandbox. The
outcome of each known-answer run goes to `known_answers` in
`sandbox.json` and `command.json`, and each case's stderr to `known-answers.log`, after a
`== REF known answer NAME` line: in the sandbox, kept in OUT_DIR on
remove, and in the command record's directory. A recipe with no Evidence
or Standalone command is not gated.

The gate proves that a bundle reproduces its own expected outputs, not
that those outputs are right. A weak or wrong case needs review, which is
why `hyp bundle check` warns about a missing must-pass or must-fail case,
and, for a rubric, a label with no supports or contradicts case.
Because a case and a run get the same environment, working directory and
hyp, a passed gate carries over to the run, except for what a case cannot
see: a read of a fixed absolute path, a read from the network,
the versions of the tools on the host `PATH` after hyp (the Python that
runs a `#!/usr/bin/env python3` script among them), and a write outside
`$HYP_OUT`. The paths of the ARGS a run is given are the subject's, so a
case cannot see them either, and a limit of the host, such as the length
of a Unix socket path, can differ from one machine to another.

Use the gate to fix a recipe before it costs an Experiment:

- Give each external resource a check starts, such as a server, a socket
  or a port, a must-pass case that starts it. The gate then fails where
  the run would.
- Run `hyp bundle known-answers <name>@<N>` after every change and before
  `hyp add exp` pins the version. When a case fails, read its stderr in
  the message, fix the Draft or make the next version, and run the gate
  again. A failure the gate catches leaves nothing in the tree.
- Try a check only on its cases' stand-ins, never on a candidate, until
  the Experiment that tests the candidate pins its rubric. A
  check fixed after it saw a candidate's verdict measures the candidate
  against the check, not against the Goal.
- When a candidate shows that a check is wrong, write the case that shows
  it first: a stand-in in the Goal's terms, with the outcome the check
  must give, in a new version. Name the case in the Experiment body. A
  sentence about a trial run pins nothing; a case does.

## Rubrics

A rubric turns measures into verdicts. It is a bundle in this format kept
in `rubrics/<name>@<version>/`, with no Launch, Doctor, Drive, Evidence or
Cleanup command and exactly one Standalone command. That command reads the
records it is given and prints one `verdict: LABEL=WORD` line per label, a
whole line, where LABEL is the rubric's own name for a rival (an alias,
`[a-z][a-z0-9-]{0,31}`) and WORD is `supports`, `contradicts` or
`inconclusive`; a case prints each label at most once. Its known-answer
cases are the rubric's specification: every label needs a case printing
`supports` and a case printing `contradicts`; `inconclusive` needs none.
A rubric names no tree id. hyp never reads a verdict line at run time and
never binds a label to an H-id: the operator types each target's verdict
into `hyp record`.

Start one with `hyp bundle new NAME@1 --rubric`, which writes the template
into `rubrics/` beside the nearest `.tree/`; a next version is
`hyp bundle new NAME@N+1 --from NAME@N`, as for a recipe. Run it with
`hyp bundle run --record .tree/evidence/<E-id>/<name> REF NAME PATH`; its
known answers gate the run as a recipe's do, and `hyp bundle evidence`
refuses a rubric.

An Experiment under a Hypothesis pins a rubric with `hyp add exp --rubric`:

- `--rubric NAME@VERSION` pins a rubric bundle by its hash. `hyp add exp`
  refuses one with a step section, other than one Standalone command, no
  known-answer case, or a label without a supports case and a contradicts
  case, and refuses it with `--recipe none` unless `--records` names the
  measures it reads.
- `--rubric reading` says the Experiment's body decides the verdicts by
  reading, for a verdict no script can decide.

An Experiment under a DesignQuestion measures only: it pins a recipe, has
no rubric and no target, and its Results carry the measures in the
observation and the cited records, with no verdict.

An Experiment that reassesses records an earlier run kept pins them with
`--records PATH` (a file or a directory under `.tree/evidence/`); hyp
stores each file's path and sha256. `hyp record` refuses a Result that
cites a file under `.tree/evidence/` outside its Experiment's own run,
`.tree/evidence/<E-id>/`, and its pinned records, and one whose pinned
record no longer hashes as pinned.

Never keep a copy of a script under `.tree/evidence/` and run that copy.
Cite a command with two pointers: `recipes/<name>@<version>/<script>` or
`rubrics/<name>@<version>/<script>`, which pins it, and the record's
`command.json`, whose `known_answers` show the gate passed.

hyp marks a Result `uncontained` when it cannot vouch for the script. Each
reason is reported once, in this order:

1. `script outside recipes/: <pointer>`: the Result cites a file outside a
   versioned bundle (`recipes/<name>@<version>/`) that ends in `.py` or
   `.sh`, has any execute bit, starts with `#!`, or ends in `.go`, `.js`,
   `.ts`, `.rb`, `.pl` or `.rs`. A script under
   `recipes/` in no versioned bundle counts: hyp neither pins nor runs it.
   So does a file inside a bundle that its hash skips (a dot-entry or
   `__pycache__`): no pin covers it.
2. `no command record of <ref> at <h12> with passing known answers`: the
   Result pins a bundle with an Evidence or Standalone command (its
   Experiment's recipe, its Experiment's rubric, or an evidence pin), and
   cites no `command.json`
   that names that ref at that hash, records at least one command, every
   one of which exited 0, and holds `known_answers` with at least one
   known-answer run, every one `ok`. A record of a failed run is not a run. When a cited
   record of that ref and hash ran but is not clean, the reason goes on
   with `; in the cited record, ` and the first command that failed (its
   words and outcome) or `the known answers did not pass`. A
   `command.json` among the Experiment's pinned records counts too, so a
   reassessment is held by the runs it read.
3. `command record of <ref> at <h12> matches no pin of this Result`: a cited
   `command.json` names a ref and hash that match none of its pins. One that
   hyp cannot read, or that names no ref or hash, gives
   `command record <pointer> names no recipe run`.

A pin to a bundle without an Evidence or Standalone command, such as a
known-answer fixture, is held by its hash alone. hyp derives the mark from
files each time it reads, so it never changes a status. `hyp record` warns
(exit 0), `hyp show` prints `uncontained:` lines, `hyp status` prints
`UNCONTAINED <n> results`, `hyp render` marks the Result, and `hyp check`
prints `warn: uncontained …`. Nothing is refused, and citing the missing
record clears the mark.

Limits:

1. A script named only in prose, whose output a Result cites, is not seen.
2. A hand-made output file cited beside a real command record is not seen.
3. A script with none of the marks in reason 1 is not seen.
4. A hand-written `command.json` that claims a clean run with passing
   known answers satisfies reason 2. A reader checks the record's
   `command.log` and `known-answers.log`.
5. A script that only a Launch, Doctor, Drive or Cleanup command reaches,
   cited from another Experiment, is held by its hash alone. hyp runs it
   in a sandbox, but no reason ties that record to the Result.
6. A `--rubric reading` that hides a measure is not seen.

## Self-containment

A recipe relies on nothing but its own files, hyp and the subject under
test, wherever its commands run. Self-contained does not mean "runs in
the sandbox": a Standalone or scripted Evidence command is self-contained
and runs with no sandbox. `hyp add exp --recipe` refuses a recipe whose
`Run:` names a script outside the bundle, or a script that is missing or
not executable.

A fixed input that is not the subject, such as a fixture or a cases file,
is a file of the bundle, so the recipe pin hashes it, and the code that
reads it is in the bundle too. A file outside the bundle is the subject,
or a read or write that the command's `External:` line names. A patch
given to `hyp sandbox create --patch` is the subject, not a file of the
bundle: the sandbox keeps it and its sha256.
hyp does not detect a fixture outside the bundle.

The operator runs none of a bundle's scripts directly; hyp runs every
command. A command is what a `Run:` line names, and every other file in the
bundle, a helper script included, is hashed content. Hashing
never parses, so a bundle whose `recipe.md` breaks a command rule still
hashes and its pins still verify. `hyp bundle new` writes `format: 1`, the
one recipe format, into every bundle it makes. Every use of a bundle without it is
refused (exit 1), naming `hyp bundle new` as the way to copy it;
`hyp bundle hash` still hashes it. `hyp bundle check` names each versioned
bundle `frozen` once a tree pins its hash, so its bytes never change, and
`draft` before.

### Reading the tree

A script reads a tree through hyp, never through the files under
`.tree/events`. `hyp events --json`, run in the store, prints
`format`, every applied event in merge order as stored, and `rejections`,
each refused event's id and reason; `hyp signals --json` prints the rounds
and signals hyp derives. A file glob would also read the events replay
rejected, and would tie the script to the store's layout. In a check, the
`hyp` first on `PATH` is the gate's copy of the running hyp; in
a sandbox step it is whichever hyp the recipe puts there, and an older hyp
may have no `hyp events`.

## Example

```text
recipes/scratch-pg@3/
  recipe.md
  launch.sh    # initdb, then pg_ctl -w start with listen_addresses='' and its socket in $HYP_SANDBOX_TMP
  doctor.sh    # pg_isready on that socket, then SHOW listen_addresses is empty
  cleanup.sh   # pg_ctl stop -m fast
```

```markdown
## Launch

Start a scratch PostgreSQL under the sandbox's tmp directory, on a Unix socket only.

Run: ./launch.sh
```

Keep recipes free of real customer or organization data and of secrets.
