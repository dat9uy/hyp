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
- When none fits and the procedure will be repeated, run
  `hyp recipe new <name>@1` and fill the bundle it writes in `recipes/`
  beside the nearest `.tree/`, following
  [RECIPE-FORMAT.md](RECIPE-FORMAT.md), To make a first version. To change a recipe, run
  `hyp recipe new <name>@<N+1> --from <name>@<N>` and edit the copy;
  never edit an existing version. Any byte change in a bundle is a new
  version, and a new version needs its known answers (RECIPE-FORMAT.md,
  Changing a recipe and Known answers). A verdict a measure decides
  needs a recipe that holds it (RECIPE-FORMAT.md, Verdict scripts);
  `--recipe none` with `--decided-by reading` is for a verdict decided
  by reading.
- `hyp add exp ... --recipe <name>@<version>` stores a hash of the recipe (the
  file, or every file of a bundle directory).
  A changed recipe is a new version. Old Results that followed the
  earlier text are then flagged questionable in `hyp status` and `hyp check`,
  which is the point: they were obtained with a different procedure.
  A bundle's commands are run by hyp: Launch and Doctor by
  `hyp sandbox create --recipe`, Drive by `hyp sandbox drive ID`, and
  Cleanup by `hyp sandbox remove`. After remove,
  `hyp recipe evidence --record DIR REF OUT_DIR` runs scripted Evidence,
  and `hyp recipe run --record DIR REF NAME ARGS` runs a Standalone
  command. When the sandbox ran them, Launch and Doctor are already done.
  Never run a bundle's script yourself.
- A bundle with an Evidence or Standalone command is an instrument and
  must have known answers. hyp runs them before every
  `hyp sandbox create`, `hyp recipe run` and `hyp recipe evidence`, and
  refuses the run when a case fails; that gate is the re-test of every
  version.
  `hyp add exp --recipe <name>@<version>` refuses an instrument with no
  case, and any bundle whose `recipe.md` has no `format: 1` line.
- Before opening a pull request that touches `recipes/`, run
  `hyp recipe check` from anywhere in the repository. Exit 1 names the
  tree, the Result, the file and both hashes of a cited bundle that was
  edited in place: restore it and make the change as a new version.

## Launch

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
`hyp recipe evidence --record .tree/evidence/<E-id>/<name> REF OUT_DIR`,
and again into a new directory to re-run it. hyp exits with the check's own
exit code. Cite its `command.json` and output, and never run its script
yourself.

A script whose output your verdict counts is a recipe command too
(RECIPE-FORMAT.md, Verdict scripts). Cite its `recipes/<name>@<version>/<script>`
pointer and its `command.json`, whose `known_answers` show the gate
passed. If `hyp record` prints `warn: uncontained`, fix what it names
before you select.

Store what you observed by hand under `.tree/evidence/<E-id>/`, with short file names:
command output, logs, timings, screenshots, the revision (`git rev-parse HEAD`).
Then record it with the hyp-prototype skill:
`hyp record --exp <E-id> --verdict <H-id>=supports|contradicts|inconclusive --observation "..." --evidence .tree/evidence/<E-id>/<file> --why "..."`.
Evidence must let someone who was not there judge the claim.

## Cleanup

Stop every process you started, by PID, and remove temporary data you created.
Never kill a process you did not start. Note anything you could not clean up in
an evidence file.
