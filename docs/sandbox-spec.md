# Spec: the sandbox

This file is the single authority for the sandbox: its steps and commands, its state machine, how the levels relate, what a crash leaves and its records. Plans, skills and Go comments point here instead of restating the rules; Go comments cite it by section as "sandbox spec §N". [replay-spec.md](replay-spec.md) owns the event schema, the refusals and the derived-status rules of the tree.

Keeping it current:
- A change to a step, a state or a record updates this file in the same PR.
- Section and rule numbers are cited from code, so never renumber them: add a new rule under the section it belongs to.
- If code and this file disagree, treat it as a bug. Decide which one is wrong, fix it, and say which in the PR.
- Text marked **[not built: #N]** is decided but not built yet. It states the target, so it is not a disagreement under the rule above. The PR that builds it removes the mark, and issue #N tracks what is left.

## 1. Steps and commands

This section is the authority for when hyp runs a recipe's commands and whether a sandbox exists when it does. The words are defined in [glossary.md](glossary.md), and the decision behind the section is [ADR 0001](adr/0001-hyp-runs-every-recipe-command.md). `skills/hyp-verification/RECIPE-FORMAT.md` says how to write a recipe that follows it.

### 1.1 Steps and sections

A bundle's `recipe.md` has one level-2 section per step (Launch, Doctor, Drive, Evidence, Cleanup) and may have a Standalone section, which is not a step. A line in one of these sections that starts with `Run:`, outside a code fence, is a **command**: one script inside the bundle followed by fixed words. The section a command sits under decides when hyp runs it and whether a sandbox exists then. A command carries no other label, except that a Standalone command is named: `Run NAME: ./script WORDS`, with NAME matching `[a-z][a-z0-9-]*` and used once in the file. A plain `Run:` line under Standalone, a named line under any other heading, and a `Run:` line under a heading that is not one of the six are refused.

| Section | Does a sandbox exist? | When hyp runs it |
|---|---|---|
| Launch | yes | `hyp sandbox create`, for each recipe in stack order |
| Doctor | yes | `hyp sandbox create`, right after the same recipe's Launch |
| Drive | yes | `hyp sandbox drive ID`, between create and remove, for each recipe in stack order; it can be run again |
| Cleanup | yes | `hyp sandbox remove`, the last recipe first; also `hyp sandbox create`, after a failed Launch or Doctor |
| Evidence | no: it reads the kept files after remove | `hyp bundle evidence --record DIR REF OUT_DIR`, after remove, the first time and again on demand, also from a later version of the recipe; `hyp sandbox remove` never runs it. An OUT_DIR without `sandbox.json` exits 2, and the refusal names a Standalone command (`hyp bundle run`) for files no sandbox wrote |
| Standalone | no sandbox at all | `hyp bundle run --record DIR REF NAME ARGS` |

`hyp bundle run` and `hyp bundle known-answers` find REF in `recipes/` or `rubrics/`; a rubric is a bundle with one Standalone command and no step (replay spec §7.6), so `hyp bundle evidence` refuses it with exit 2, and a ref found in both libraries is refused with exit 1. `hyp recipe` runs as a hidden alias of `hyp bundle`, because frozen bundles call it from their scripts (ADR 0018).

Evidence that the operator gathers by hand while the sandbox exists is not a command, and is unchanged: it is the files saved under `.tree/evidence/<E-id>/`. Scripted Evidence is the same step done by a command, which is why it runs after Cleanup: it reads only the kept files, so it can be run again.

A sandbox is a directory plus its record, not process isolation. Each command hyp starts leads its own process group, which its row records as `pgid` (ADR 0002). Before Cleanup, `hyp sandbox remove` stops the groups of commands a killed hyp left unfinished (§4); after Cleanup, it stops the groups of the Launch, Doctor, Drive and Cleanup commands that still have members: SIGTERM, up to 5 seconds, then SIGKILL. A timeout or Ctrl-C kills the running command's group (its row says `interrupted` for Ctrl-C), and an Evidence or Standalone group ends with its command; create still runs Cleanup after a Ctrl-C. Limits: a daemon that leaves its group, as Postgres's postmaster does by calling `setsid`, still needs its recipe's Cleanup, and the leftover check still applies; a group id reused after its group ended is not told apart; what a check cannot see is listed under [replay-spec.md](replay-spec.md) §7.6 and ADR 0003. "Runs in a sandbox" means only that the command runs while the sandbox exists and gets that sandbox's `HYP_SANDBOX_*` variables (§1.3).

### 1.2 hyp runs every command

hyp runs every command in a bundle, and the operator runs none of a bundle's scripts directly. A command is what a `Run:` line names; every other file in the bundle, a helper script included, is hashed content (ADR 0002). Hashing never parses: a bundle whose `recipe.md` breaks a command rule, such as a `Run:` line outside a section or shell syntax, still hashes, so every pin to it still verifies, and `hyp check` names it as a bundle that does not open. `hyp bundle new` writes `format: 1` into every bundle it makes, a first version from its embedded template (`hyp bundle new NAME@1`) as well as a copy (`--from`), the one recipe format, and the line is a required field. Every use of a bundle without it is refused with exit 1, reading only its frontmatter: `hyp add exp`, `hyp sandbox create`, `hyp bundle run`, `hyp bundle evidence`, `hyp bundle known-answers`, and `hyp record --evidence` into such a bundle. The refusal names `hyp bundle new` as the way to copy it. `hyp bundle hash` still hashes it (ADR 0002, ADR 0007). `hyp bundle check` names each versioned bundle `frozen` when a tree pins its current hash, so its bytes never change, and `draft` when none does. When the repository root holds `go.mod`, `hyp bundle check` warns about each `.go` file in a versioned bundle, and `hyp add exp --recipe` about each one in the bundle it pins, because `go vet ./...` reads a `recipes/<name>@<N>/` directory holding one as a package outside the main module; the warning suggests a non-`.go` name, such as `x.go.txt`, and both commands still exit 0.

A recipe is self-contained when it relies on nothing but its own files, hyp and the subject, wherever its commands run. Self-contained does not mean "runs in the sandbox": a Standalone or scripted Evidence command is self-contained and runs with no sandbox. A fixed input that is not the subject, such as a fixture or a cases file, is a file of the bundle, so the recipe pin hashes it, and the code that reads it is in the bundle too. A file outside the bundle is the subject, or a read or write that the command's `External:` line names (ADR 0008, owner decision on 2026-11-01). hyp does not detect a fixture outside the bundle.

Each Evidence and Standalone command states its reach on one line `External: TEXT` directly under its `Run:` line in `recipe.md`: the subject arguments and any outside reads or writes, or `none`. The line is no field `recipe.Load` requires, so a frozen bundle without it still loads and records. `hyp bundle check` warns `<command>: no External: line` for a command without one and `<command>: External: not stated (TODO)` for one whose text is empty or `TODO`, and exits as before. `hyp bundle new` writes `External: TODO` under each such command of a copy that has no line, and under the Standalone command of a first version. Every line of a first version's template that its author must replace holds the marker `hyp:placeholder`, and `hyp add exp --recipe` refuses (exit 1) a bundle that still holds it in any file outside a known answer's `input/`, naming the files (ADR 0009).

hyp's Go learns no replay concepts. It runs a script with its arguments and records the exit code and the output. hyp is not a job runner: no retries, no background jobs and no stdin streaming.

### 1.3 How a command runs

- No shell runs a command. Its first word is a relative path containing `/` to an executable script inside the bundle, and the rest are plain words; quotes, `$`, pipes, redirects and globs are refused.
- Launch, Doctor and Cleanup run with the recipe's copy in the sandbox as their working directory, and get `HYP_SANDBOX_ID`, `HYP_SANDBOX_WT`, `HYP_SANDBOX_TMP`, `HYP_SANDBOX_RECIPE_TMP`, `HYP_SANDBOX_INPUTS` (the sandbox's copy of `--inputs FILE`, when given), `HYP_SANDBOX_PATCHES` (the directory of the sandbox's copies of each `--patch ARM=FILE`, `<ARM>.patch` each, when any is given) and a `PATH` that starts with the sandbox hyp. They also inherit the environment of the process that runs hyp.
- Drive commands run like Launch: in the recipe's copy, with the sandbox's variables. They take no arguments from the operator; a Drive reads per-run values from the inputs file or from the state its Launch wrote.
- Evidence and Standalone commands are checks, and run exactly as their known-answer cases do (ADR 0003). They get no `HYP_SANDBOX_*` variables, because no sandbox exists when they run, and inherit nothing else of the caller's environment: only `PATH` (the copy of the running hyp the gate used, first, then the caller's `PATH`), a scratch `HOME` and `TMPDIR`, `GIT_CONFIG_NOSYSTEM=1` and `HYP_OUT`, an empty directory for the files the check leaves; stdin is `/dev/null`. They run from a temporary copy of the bundle whose hash equals the library's, deleted afterwards, in an empty working directory of their own. One function makes the copy and one makes the check's directory, for a known-answer case and for a run alike, so their paths differ only in the random digits of their names and a check cannot tell the gate from the run. The check's `TMPDIR`, `HOME` and working directory are deleted when it ends, also when it fails; what explains a failure goes to stderr, and a file to keep goes under `$HYP_OUT`. A hyp killed while it runs leaves that copy (`hyp-bundle-*` under the system temp directory): hyp never reads it again and does not clean it up, the system's temp cleaning removes it, and the operator may delete it. An Evidence command gets the absolute OUT_DIR appended as its last argument; a Standalone command gets the caller's ARGS appended after its fixed words, each file argument replaced by a copy and each directory argument made absolute (see Kept inputs in §5.1).
- Launch and Doctor together have 10 minutes for each recipe, and each Cleanup command has 2 minutes. Each Drive command has 60 minutes, each Evidence command 10 minutes, and a Standalone command 60 minutes (§5.2).
- When hyp runs a Drive, Evidence or Standalone command for the operator and the script exits non-zero, hyp stops there, prints the last 30 lines of the command's log with the error, and exits with that code unchanged, so a check that exits 30 can be told apart from hyp's own 1 (refusal), 2 (bad arguments) and 3 (environment). A command that times out or is killed by a signal has no exit code, and hyp exits 3. hyp's own refusals come before any command starts and leave no row, so the record always tells a script's exit from hyp's; a script whose code alone must tell it apart exits 10 or above.

## 2. The sandbox machine

A sandbox's state is derived from stored facts, as a tree's statuses are, and is never stored as a word: each recipe's `setup`, the lock and the record's `removing` field. Each command refuses by state, as this table says; a cell that says "runs" runs as §1 says. In code, `deriveState` in `internal/sandbox/state.go` is the rule, and `TestStateTable` mirrors the table by hand, row for row.

| State | Derived from | create | drive | remove OUT_DIR | remove --discard |
|---|---|---|---|---|---|
| absent | no directory | runs | 2 | 2 | 2 |
| `building` | directory, no record, lock held | 2 | 3 busy | 3 busy | 3 busy |
| `setting_up` | record, a recipe's `setup` unfinished, lock held | 2 | 3 busy | 3 busy | 3 busy |
| `interrupted`, no record | directory, no record, lock free | 2 | 3 | 2, names `--discard` | runs |
| `interrupted` | record, a recipe's `setup` unfinished, lock free | 2 | 3 | runs | runs |
| `ready` | every recipe's `setup` is `passed` or `none`, or the ref is `none` | 2 | runs, or 3 while an unfinished row's group lives, or 2 with no Drive command | runs | runs |
| `setup_failed` | any recipe's `setup` is `failed` | 2 | 3 | runs | runs |
| `removing` | `removing` is set | 2 | 3 | runs | runs |

A `setup` is unfinished when it is absent on a recipe other than `none`. The first match wins, in this order: lock held, `removing` set, `failed`, unfinished, otherwise `ready`. A lock held by another hyp is the busy refusal, whichever command holds it and whatever the record says, because a command takes the lock before it reads the state: a drive running on a `ready` sandbox refuses a second drive the same way. Create refuses any directory that exists with 2, before it looks at the lock.

A record that does not parse is not a missing one: every command but create refuses it with 3, as a corrupt record, and hyp changes nothing. A crash cannot produce one, because records are written through a temporary file and a rename (§5.1).

## 3. How the levels relate

Three levels write the facts, and a higher level's state is read from the level below it, never kept in step with it.

| Level | What it writes | Who writes it |
|---|---|---|
| The sandbox | the lock (`sandbox.lock`), `removing`, `removed` | create, drive and remove hold the lock; remove writes `removing`, and `removed` in OUT_DIR |
| Each recipe | `setup`, `cleanup` | create writes `setup` as each recipe's setup ends; create or remove writes `cleanup` after its Cleanup |
| Each command | its row in `commands`, with `pgid`, `ended`, and `exit` or `error` | hyp, before the command starts, once it has a group, and when it ends |

The order within create is fixed so a reader of the record is never told more than happened. Create writes the record with every `setup` absent. It runs each recipe in turn and writes the record once that recipe's `setup` is known. It writes `manifest.json` before the last recipe's `setup`, so a sandbox that reads as `ready` has its baseline. `--recipe none` has no recipe to wait for and reads as `ready` once its record exists, so its manifest comes before its record.

## 4. Crashes

A crash here means hyp was killed or the machine stopped. A Ctrl-C or SIGTERM is not one: hyp ends the running command's group itself and runs Cleanup, as §1.1 says.

**What survives a killed hyp.** The record survives as of its last write. The rows hold every command that started, and a row with no `ended` is one hyp stopped during. The lock goes, because the kernel drops an flock when its process dies. The command hyp was running does not stop: it leads its own group, so it keeps running with no hyp watching it.

**What a kill leaves, by where it lands.**

| Killed during | The sandbox reads as | Then |
|---|---|---|
| making the sandbox, before its record | `interrupted`, no record | `remove --discard` clears it; there is nothing to keep |
| a recipe's Launch or Doctor | `interrupted` | remove runs; drive is refused |
| after the last Doctor, before the last `setup` is written | `interrupted`, with a manifest | remove runs; drive is refused, so the operator makes a new sandbox |
| a Cleanup after a failed setup | `setup_failed` | remove runs and Cleanup runs again |

**A create whose setup fails** runs every Cleanup, the last recipe first, and stops every group. Then it reads the leftovers. When no process still runs in the sandbox, it deletes the whole sandbox, record and `recipe.log` included, and exits 3 with the last 30 lines of `recipe.log` in its message, so nothing is left to remove. When a process still runs, or hyp could not check, it writes the record with the failed `setup`, keeps the sandbox, which reads `setup_failed`, and names `remove ID OUT_DIR`, which keeps `recipe.log`, and `remove ID --discard`.
| a Drive | `ready`, with an unfinished row | while the row's group lives, drive exits 3 and remove stops it; once the group is gone, drive runs again |
| a remove, before its save | `removing` | remove runs again to the same OUT_DIR, Cleanup included |
| a remove's save to OUT_DIR | `removing`, with `OUT_DIR.partial-ID` and no OUT_DIR | remove runs again to the same OUT_DIR: it clears the staging directory and saves again |
| a remove's delete, after its save | `removing` with the whole record in OUT_DIR, or `interrupted`, no record, if the delete reached `sandbox.json` | remove to the same OUT_DIR exits 2 and names `--discard`, which finishes; the record is already saved |

**Remove always gets through what it can.** Refusals that change nothing come first and leave the state as it was: bad arguments, busy, a missing `lsof`, a bad `--keep`, an existing `OUT_DIR`, and a directory with no record given an OUT_DIR. An existing OUT_DIR that holds this sandbox's saved record, with `removed` set, is refused naming `remove ID --discard`, because the remove that saved it stopped before the delete. Then, in order, remove:

1. writes `removing`;
2. stops the groups of unfinished rows;
3. runs every Cleanup, the last recipe first;
4. stops every remaining group;
5. checks for leftovers, and exits 3 if a process still runs in the sandbox;
6. saves the record to OUT_DIR, unless `--discard`: it builds the save in `OUT_DIR.partial-ID` beside OUT_DIR, ID being the sandbox's id, and renames it to OUT_DIR, so OUT_DIR holds a whole record or does not exist. A staging directory of that name, which only a killed remove of this sandbox leaves, is cleared first;
7. deletes the sandbox.

A command a killed hyp left running is stopped before Cleanup, so it cannot change what a Cleanup has already looked at. A failed write of `removing` only warns, as a failed Cleanup row write does, because stopping processes matters more than the trace. Once `removing` is written it stays: a remove refused at step 5 leaves a sandbox that drive refuses and remove can run again, so a Cleanup that already ran is never followed by a Drive. A directory with no record has no rows and no Cleanup, so `--discard` takes the lock, runs the leftover check and deletes.

"Can be removed" means once any process that left its group has stopped. Such a process still runs in the sandbox, and remove keeps exiting 3 and listing it until the operator stops it.

**Limits.** A group id reused after its group ended is not told apart from it. So an unrelated group holding a recorded unfinished id refuses a drive, whose refusal says remove still works, and remove can signal it (§1.1). hyp does not check for a killed run's scratch outside the sandbox (§1.3). `hyp bundle evidence` does not read how far setup got; the kept record's `setup` and rows show it, and #46 holds the open question.

**Older records.** A record's format decides nothing here: every format is read by the same rule. An older record that says `none` or `passed` reads as `ready`. One whose recipe has no `setup` at all reads as `interrupted`, so it cannot be driven and can be removed.

## 5. Records, choices and the subject

### 5.1 Records

- **Sandbox record.** `hyp sandbox create` writes `sandbox.json`, and only hyp writes it; `hyp sandbox remove` copies it to OUT_DIR. Each command's output goes to `recipe.log` after a `== REF step: command` line. The fields are listed under "The sandbox record" in `RECIPE-FORMAT.md`. It names no host path for anything a replay needs: `host` is the host's directory name, a store inside the host is `<host>/REL`, and the sandbox's clone, scratch, `bin/`, inputs copy and recipe copies are derived from its id, not recorded (format 3). Format 4 adds `removing`, which a remove writes once it has passed its refusals and which the sandbox's state reads (§2). `patches` names each Arm's patch by its sha256 (§5.3). Its `commands` list holds each command hyp ran in the sandbox (Launch, Doctor, Drive, Cleanup), in order, with its step, its arguments and its exit code; `hyp sandbox drive` adds to it.
- **Command record.** For each run of Evidence or Standalone commands, hyp writes a command record in the new directory that `--record DIR` names (refused with exit 2 if it exists, is or resolves inside a `recipes/` directory the recipe lookup searches, or, for Evidence, inside OUT_DIR, comparing directories by identity so symlinks and other spellings do not get past it; with exit 3 if its path cannot be resolved): `command.json`, naming the recipe ref, its hash, the library path and the working directory (each `<repo>/REL`, see Kept inputs), the step, with one row per command (its arguments as shown, its exit code or why it has none, and when it ran) `inputs`, each file an operator argument names, and `dirs`, each directory one names, with its tree's hash and each file's (format 6, so `hyp record` can copy them into the run's subject on a Result); `command.log`, the output after `== REF step: command` lines; and each check's product, its stdout in `stdout` and its `$HYP_OUT` tree in `out/` (Standalone), or `stdout/<n>` and `out/<n>/` (Evidence command `n`), kept even when the check fails (ADR 0003). The fields are listed under "The command record" in `RECIPE-FORMAT.md`.
- **Kept inputs.** An operator argument of `hyp bundle run` or `hyp bundle evidence`, alone or as the VALUE of a `KEY=VALUE` word, is a file argument when it is written as a path (absolute, or relative with a `/`; a relative one is read against the caller's directory) and names a regular file, symlinks followed. A bare word is never one, even when a file of that name exists. Before the first command runs, hyp copies each file argument N (ARGS counted from 1) under the run's temporary directory, which the script gets in its place, hashes that copy, and keeps a second copy beside the record at `DIR/inputs/N/<base name>`, so a script that writes to its argument changes neither the original nor the kept bytes. A file over 1 MB is not kept; its sha256 and size are. `command.json` (format 5) lists each in `inputs` (`arg`, `path`, `sha256`, `size`, `kept`), and the row shows `inputs/N/<base name>` in its place. A file inside the tree is copied too. A directory argument (such as Evidence's OUT_DIR) is passed as an absolute path and shown as `<repo>/REL`, `<repo>` being the directory name of the caller's git work tree (or of the caller's directory outside one); one outside it is shown as given, and hyp warns that a replay cannot reach it. hyp keeps no copy of a directory argument, but hashes its tree just before the first command runs and lists it in `dirs` (`arg`, `path`, `sha256`, `files`), so a record names the subject it judged. `path` and `cwd` are written the same way. So neither the record nor the `==` lines name a host path for anything a replay needs. The inputs are bound at run time, not when the Experiment is added (ADR 0008). Limits: a file the command reads without being given it is not kept; its `External:` line names it. A format-3 record's `arg_sha256` (each absolute file argument's sha256) is still read.
- **Known answers.** Each recipe entry of `sandbox.json`, and `command.json`, holds `known_answers`: `hyp_sha256`, the hyp that ran the cases, and `cases`, one `name` and `outcome` per case ([replay-spec.md](replay-spec.md) §7.6). Each case's stderr goes to `known-answers.log` after a `== REF known answer NAME` line: in the sandbox, kept in OUT_DIR on remove, and in the command record's directory (ADR 0002).
- **Format.** `sandbox.json` and `command.json` each carry `"format"`, the shape number of that file (ADR 0002). A golden test per file fails `go test` when its shape changes without a new number, so a reader can compare the number before reading the rest. A record without the key was written before it.
- **Changes.** Before the last recipe's `setup` is written (§3), hyp writes `manifest.json` in the sandbox, a list of every path in `wt/` with its size, mode and hash. `hyp sandbox remove` diffs `wt/` against it and writes `changes.json` (format 1) to OUT_DIR, with the added and modified files up to 10 MB under `changes/`. The fields are listed under "Changes" in `RECIPE-FORMAT.md`.
- **Writes.** hyp writes `sandbox.json` and `command.json` to a temporary file in the same directory and renames it over the record, so a reader sees the old record or the new one, never part of one. A hyp killed between the two leaves a `.sandbox.json-*` or `.command.json-*` file that nothing reads; a sandbox's goes with it at remove.
- **Rows.** hyp writes each command's row before the command starts and rewrites it when the command ends. A row with no `ended` (and so no `exit` or `error`) is a command hyp stopped during: hyp was killed, or the machine stopped. Its output so far is in the log. When the write before a command fails, the command does not run and hyp exits 3, except under Cleanup, which runs every command anyway and reports the failed write as a failed Cleanup. A remove refused for a leftover process keeps the Cleanup rows it ran, and each recipe's `cleanup`, in the sandbox's record.
- **One hyp command per sandbox.** `hyp sandbox create`, `drive` and `remove` hold a lock on `sandbox.lock` in the sandbox while they run. A second one on the same sandbox is refused with exit 3, naming the hyp command that holds it, and changes nothing. The lock goes away when its hyp exits, however it exits. `hyp bundle evidence` and `hyp bundle run` take no lock, because each writes only the new directory it made.

### 5.2 Choices

The open choices were decided in a hyp tree, whose records stay in the private development history, and while building.

1. How a Standalone command is named: an inline label, `Run NAME: ./script WORDS`. The owner picked it over a heading; both resolved every real use.
2. How the operator hyp is checked: by a Doctor command in the harness recipe, so hyp records no operator fields. Accepted limits: an Experiment under `--recipe none` gets no check, and each recipe that needs it carries the check.
3. One recipe format (ADR 0002). The reach test, which froze a bundle holding a script no command named, was deleted: the known-answer gate proves what a verdict script gives, while a reach test only proved that a script was named. Every file other than a command's script is hashed content, hashing never parses, and a bundle is `frozen` once a tree pins it and `draft` before.
4. The Drive command is `hyp sandbox drive ID`. The name is the step, as create runs Launch and Doctor and remove runs Cleanup. With no operator arguments, the recipe's fixed words and its inputs decide a drive.
5. `hyp sandbox remove` does not run scripted Evidence; `hyp bundle evidence` runs it, the first time and every time. One path serves the first run and every re-run, a later version is only another REF, and remove's own exit codes never mix with a check's.
6. Timeouts: Drive and Standalone 60 minutes per command, because the longest kept create-to-remove was 11.6 minutes and a re-test script drives sandboxes itself; Evidence 10 minutes per command, because it only reads kept files.
7. The command record is a directory the operator names with `--record`, so it can sit under `.tree/evidence/<E-id>/` and be cited, and large output stays in its log. The commands run from a temporary copy, so a script cannot change the library version it ran from.
8. (replaced by choice 3)
9. A busy sandbox is refused, not waited for. A Drive can run 60 minutes, a waiting hyp would hang its caller silently, and queueing is job-runner work §1.2 rules out. The lock is an flock inside the sandbox, so the kernel frees it when hyp dies and it is deleted with the sandbox.
10. An interrupted command leaves a row with no `ended`. Every command hyp runs is a row (§1.2), and only a row written before the command starts survives a kill. A finished row is unchanged.
11. hyp does not clean up a killed run's temporary copy. It holds nothing the library does not, and a sweep of another hyp's directories would be a reaper.
12. A record DIR inside OUT_DIR is refused. The check reads OUT_DIR, its record is written before each command starts, and a re-run must read the same files as the first run. A record beside OUT_DIR, the documented place, is unaffected.
13. A record DIR inside any `recipes/` or `rubrics/` directory the lookup searches is refused, not only one inside the ref's own bundle. A record changes the hash of whatever bundle it lands in, an outer copy of the ref must hash the same as the nearest, and a record in a `recipes/` directory that does not exist yet would start a library the lookup finds.
14. (replaced by choice 3)

### 5.3 The subject

A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. The store cut replaces any `.tree` the base holds at its root, so the clone's `.tree` is the cut and nothing else (ADR 0022); when the record holds the store, the cut reads the record's copy. The clone is marked a sandbox clone, so hyp in it keeps no record and reads and writes its work-tree `.tree` (ADR 0024). It also builds it from one patch for each Arm (see Arm in [glossary.md](glossary.md)), given as `--patch ARM=FILE`.

- **A change that lands in code is a commit.** An Experiment about it uses that commit as `--base` and takes no patch.
- **A patch is a throwaway Arm: a change over the base that never lands in code.** An Arm has at most one patch. A patch is never versioned, never stacked on another patch and never rebased. An Arm worth keeping is merged and becomes a commit, and the next Experiment uses that commit as its base. hyp refuses a patch that does not apply to the base (`git apply --check` in the clone) before any command runs, with exit 2, and leaves no sandbox. It applies nothing: Launch builds each Arm. It records each patch's Arm and sha256 under `patches` in `sandbox.json`, copies it to `patches/<ARM>.patch` in the sandbox, passes that directory as `HYP_SANDBOX_PATCHES`, and `hyp sandbox remove` keeps it as `OUT_DIR/patches/<ARM>.patch`, so where the operator wrote it does not matter. A patch is the subject, so it is not a file of the recipe bundle (see Self-contained in [glossary.md](glossary.md)).
- **Code under test is built in Launch, from the subject the sandbox assembled.** Each binary it builds is named by its sha256, and only the sha256 is kept. A check does not take a binary under test from outside a sandbox, and a known-answer case cannot build one: it has 60 seconds and a scratch `HOME` by design.

The rule exists so a record names its subject only by what the repository holds (the base commit) or the tree keeps (each patch and its sha256), and a replay can rebuild the subject from them. A base commit holds only while the repository's history keeps it.
