---
name: hyp-verification
description: Run one hyp Experiment honestly through Launch, Doctor, Drive, Evidence and Cleanup, using a versioned recipe when one exists. Use whenever a hyp-prototype Experiment has to be executed.
---

# hyp-verification

This skill runs one Experiment (`E-` id) created by the hyp-prototype skill and
leaves evidence that anyone can check without you. It does not decide what
the result means; that is the `hyp record` verdict.

## Find or write a recipe

- A recipe is a repeatable procedure, a bundle `recipes/<name>@<version>/`.
  hyp looks
  in `recipes/` beside `.tree/`, then in each parent directory up to the
  repository root, so list `recipes/` at each of those levels, read the one
  that fits, and cite it as `--recipe <name>@<version>`.
- When none fits, run `hyp bundle new <name>@1` if any script's output
  will count toward a verdict, or if the procedure will be repeated.
  Otherwise use `--recipe none` and state the procedure in the Experiment
  body. Fill the bundle it writes in `recipes/`
  beside the nearest `.tree/`, following
  [RECIPE-FORMAT.md](RECIPE-FORMAT.md), To make a first version. To change a recipe, run
  `hyp bundle new <name>@<N+1> --from <name>@<N>` and edit the copy;
  never edit a version a tree pins (`hyp bundle check` prints it
  `frozen`). A Draft, which no tree pins yet, may be edited in place; then
  `hyp bundle known-answers` must pass again. A new version needs its known
  answers (RECIPE-FORMAT.md,
  Changing a recipe and Known answers). A recipe measures and never
  prints a verdict. How measures become verdicts is a rubric in
  `rubrics/`, pinned with `--rubric <name>@<version>` (RECIPE-FORMAT.md,
  Rubrics); `--rubric reading` is for a verdict decided by reading,
  and then the rubric is prose in the Experiment's body.
- `hyp add exp ... --recipe <name>@<version>` stores a hash of every file
  of the bundle directory. A Result stays clean when a new version is
  added. If a pinned version is edited in place, every Result under it is
  flagged questionable in `hyp status` and `hyp check`, and
  `hyp bundle check` reports drift, so make every change as a new version.
  A bundle's commands are run by hyp: Launch and Doctor by
  `hyp sandbox create --recipe`, Drive by `hyp sandbox drive ID`, and
  Cleanup by `hyp sandbox remove`. After remove,
  `hyp bundle evidence --record DIR REF OUT_DIR` runs scripted Evidence,
  and `hyp bundle run --record DIR REF NAME ARGS` runs a Standalone
  command. When the sandbox ran them, Launch and Doctor are already done.
  Never run a bundle's script yourself, so the known-answer gate and its
  command record cover every run, and never run a check on a candidate
  before the Experiment that tests it pins its rubric.
  When a candidate shows a check is wrong, write the case that shows it in
  a new version first, and name the case in the Experiment body.
- A bundle with an Evidence or Standalone command is an instrument and
  must have known answers. hyp runs them before every
  `hyp sandbox create`, `hyp bundle run` and `hyp bundle evidence`, and
  refuses the run when a case fails; that gate is the re-test of every
  version.
  `hyp add exp --recipe <name>@<version>` refuses an instrument with no
  case, and any bundle whose `recipe.md` has no `format: 1` line.
- Before opening a pull request that touches `recipes/`, run
  `hyp bundle check` from anywhere in the repository. Exit 1 names the
  tree, the Result, the file and both hashes of a cited bundle that was
  edited in place: restore it and make the change as a new version.

## Launch

With a bundle, hyp runs Launch, Doctor, Drive and Cleanup:
`hyp sandbox create <id> --host <repo root> --base <commit> --recipe <ref> --recipes <dir holding recipes/>`,
then `hyp sandbox drive <id>`, then
`hyp sandbox remove <id> .tree/evidence/<E-id>/run`. A sandbox id uses
`[a-z0-9-]` only, so an E-id is refused (exit 2). Put each `--record DIR`
beside OUT_DIR, never inside it. The sections below are the manual steps
for `--recipe none`.

Start the subject under test: the app, a script, a database. Work from the
repository root and use seed or test data only, never real customer or
organization data. Note every process you start (command, PID, port) so you
can stop exactly those later.

## Doctor

Prove the subject is healthy before testing anything: it answers on its port,
the database is the test one, migrations are applied, the build is the
revision you meant. An experiment run against a broken subject says nothing
about the hypothesis; fix the subject or record the result as inconclusive.

## Drive

Perform the steps the Experiment's title and recipe describe, exactly. When
reality forces a deviation, write it down; a deviation that matters makes the
result inconclusive rather than supporting.

## Evidence

Evidence comes in two kinds. Evidence gathered by hand is what you save
while the sandbox still exists, before Cleanup, as below. Scripted Evidence
is a recipe's `## Evidence` command. Run it after remove with
`hyp bundle evidence --record .tree/evidence/<E-id>/<name> REF OUT_DIR`,
and again into a new directory to re-run it. hyp exits with the check's own
exit code. Cite its `command.json` and output, and never run its script
yourself.

A script whose output your verdict counts is a bundle command too: a
recipe check or a rubric (RECIPE-FORMAT.md, Rubrics). Cite its
`recipes/<name>@<version>/<script>` or `rubrics/<name>@<version>/<script>`
pointer and its `command.json`, whose `known_answers` show the gate
passed. If `hyp record` prints `warn: uncontained`, record a new Result
under the same Experiment that cites the missing pointer or command
record, before you select; a Result cannot change after it is recorded.

Store what you observed by hand under `.tree/evidence/<E-id>/`, with short file names:
command output, logs, timings, screenshots, the revision (`git rev-parse HEAD`).
Then record it with the hyp-prototype skill:
`hyp record --exp <E-id> --verdict <H-id>=supports|contradicts|inconclusive --observation "..." --evidence .tree/evidence/<E-id>/<file> --why "..."`.
Give a verdict on every target of the Experiment, `inconclusive` when the
run could not decide; a Result under a measuring Experiment gives none.
Of the files under `.tree/evidence/`, cite only those of this
Experiment's own run, `.tree/evidence/<E-id>/`, and the records it pinned with `--records`.
Evidence must let someone who was not there judge the claim.

## Cleanup

Stop every process you started, by PID, and remove temporary data you created.
Never kill a process you did not start. Note anything you could not clean up in
an evidence file.
