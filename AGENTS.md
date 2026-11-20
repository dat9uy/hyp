# Working on hyp

This file is for any agent or person who changes hyp. It says how the repository is divided, how a change earns its way in, and how to build and test. Each section states a rule and links to the document that owns it; read that document for the detail.

## What this repository holds

- **hyp**, the binary: `cmd/hyp` and `internal/`. It stores each Goal as append-only events under `.tree/events/`, one file per actor, and calculates every state from them.
- **Two skills on `main`**, the first of events v2: `skills/hyp-grill`, which grills and re-grills a Goal and ships the runner's run rules as its part `run-rules.md`, and `skills/hyp-statechart`, with which the grill writes a Goal's model, its switches and its chart, and which ships the statechart tool (MIT) with its license and source. The five skills of hyp 0.1.0 are kept on the `v1` branch: `skills/hyp-prototype` (run an investigation as a tree of rival Hypotheses), `skills/hyp-verification` (run one Experiment from a recipe), and the owner flow: `skills/hyp-owner-grill` (grill the owner into a Goal), `skills/hyp-goal-run` (run a Goal as an unattended `/goal`, whose prompt the hyp 0.1.0 binary prints) and `skills/hyp-owner-answer` (answer what a stopped run left for the owner).
- **Docs** in `docs/`: the v2 rules in [state-machine.md](docs/state-machine.md); the rules of hyp 0.1.0 in [replay-spec.md](docs/replay-spec.md) and [sandbox-spec.md](docs/sandbox-spec.md); the words in [glossary.md](docs/glossary.md), what the owner of a Goal sees and answers in [owner-guide.md](docs/owner-guide.md), the principles and the owner's role in [principles.md](docs/principles.md), the loop's guards in [how-hyp-improves-itself.md](docs/how-hyp-improves-itself.md), why the select is the gate in [ADR 0015](docs/adr/0015-the-select-is-the-gate.md), why a skill change is a build Goal in [ADR 0021](docs/adr/0021-a-skill-change-is-a-build-goal.md), and the scope map in [core-and-self-review.md](docs/core-and-self-review.md).
- **Test data** under `internal/*/testdata`, such as the synthetic store in `internal/testutil/testdata/synth`. The recipes and trees that tested hyp 0.1.0, and the trees that decided and built its shared store (ADR 0020), are kept on the `v1` branch, not on `main`. Records made before the public release stay in the private development history.

Install the binary as [install.md](docs/install.md) describes. To use the skills of hyp 0.1.0, link them from a checkout of the `v1` branch.

## Two scopes, and why

Every item in this repository belongs to one of two scopes:

- **Core** is every concept hyp has, such as Goal, line, example, Probe, runner and select, and everything a user who installs hyp into another repository runs, loads or reads there.
- **Self-review** is a Goal of this repository whose lines are about how hyp is tested, not about a concept hyp has. An example is a Goal that checks a Probe or a test fixture of hyp itself. `main` keeps none today. In hyp 0.1.0, self-review was the recipes on the `v1` branch, and the scope map lists them.

The split exists for four reasons:

1. **The skills and docs ship to other repositories.** Text inside a skill about how hyp tests itself reaches users who will never run that test. Naming the scope of each passage is how such text gets noticed and moved out.
2. **Every change says what it serves.** An issue, a plan or a PR names the concept it changes, so work on how hyp is tested is not mistaken for a feature, and a feature is not judged by how one test happens to use it.
3. **Self-review has no rule of its own.** Its examples give their line's verdict on H0 before they count, like the examples of every Goal.
4. **Visibility is a separate question.** Every item in this repository is public, so "private" never needs to mean "self-review".

Three rules keep the boundary honest: a concept stays core when self-review is what tests it; a self-review Goal does not make the concepts it uses self-review; and self-review invents no concept of its own. The test, the rules, the map of every item and the table of self-review words are in [core-and-self-review.md](docs/core-and-self-review.md), section "The test".

## Before you change anything

- Name the scope of the change by the concept it changes, not by the Goal or test that checks it. An item that fits neither scope has not been split far enough.
- If the issue names a Goal, run `hyp status G` first, with `HYP_ACTOR` set. It shows the Goal's state, its stop, who it waits for, the runs used and the allowed acts of each role.
- Use the words in [glossary.md](docs/glossary.md). A new self-review word needs an entry in the scope map's word table.

## Improving hyp with hyp

hyp is changed the way it asks its users to decide things: with a Goal whose gap and guard lines each hold examples, runs of rival Hypotheses against those examples, and a select hyp records only for the one candidate met on every line. The loop:

1. **A real event points at hyp.** In a real run a refusal or a warning fires, a Goal reaches a stop it cannot leave, `hyp status` lists an event under `refused at replay`, or the owner flags a run. A change is never proposed on speculation.
2. **An issue names the event and its scope.**
3. **A Goal holds the question.** The owner adds it with `hyp add goal`: words, and gap and guard lines. The grill places one or more examples on each line with `hyp place`, each run on H0 to show it gives the verdict its line needs. The rules are in [state-machine.md](docs/state-machine.md).
4. **The runner runs and selects.** `hyp run` runs a line on a commit, within a budget of three Hypotheses, and `hyp select` records the one candidate that is met on every line and holds a spec. The select's report is the PR body, and `hyp status` prints the allowed acts of each role as ready command lines. hyp v2 was built without hyp up to this point, because v1 could not measure v2; from here every change is a v2 Goal.
5. **The change passes its gate.**
   - A failure that can be detected from the log or from git goes into the binary as a refusal or a warning, not into skill text. Its gate is `go test`.
   - A change to `skills/`, a new skill included, is a build Goal like any code change: its lines, examples whose Probes the Goal writes and that give their line's verdict on H0 before they are placed, and the select. How a skill is tested is that Goal's choice.
   - A prose-only skill change, one that changes no command, flag, default, cap or exit code a skill names, ships as a plain PR that names its event and the owner's decision. Its gate is `go test`.

   [how-hyp-improves-itself.md](docs/how-hyp-improves-itself.md) explains why the loop is guarded while it runs instead of proven in advance.
6. **The PR body is the report `hyp select` printed:** each line's examples, outputs and verdicts, and the selected commit and row.

## Run controls

- Skill text reaches every run that loads it. Editing `skills/` changes what later runs that install skills from this repository measure, even when the edit looks harmless.
- A runner also sees the investigated repository's own agent files. The runner works in a checkout of that repository, and a Probe runs in a copy of the commit's subject, so any `AGENTS.md` or `CLAUDE.md` in the commit is part of the runner's environment. This file is one of them whenever hyp is the host.

## Rules

- An example is self-contained: its Probe reads nothing outside itself except the subject and the environment variables its `external` section names. A fixed input that is not the subject is an `input/<path>` section of the example. A new version of an example only adds a floor or a ceiling; any other change is a new example key ([state-machine.md](docs/state-machine.md) §7 and §10).
- Name a binary by its sha256, not by the commit it was built from.
- hyp exits 1 for a rule refusal, 2 for bad arguments and 3 for an environment problem, such as a missing tool or a Probe that failed or timed out ([state-machine.md](docs/state-machine.md) §12).
- Skill text cites no ADR and no file under `docs/`: the skills are copied into other repositories without them. A skill states its rule inline, and the ADR behind it names the skill file. Removing such a citation is a product fix.
- A state machine has one spec, and a change to the machine updates that spec in the same PR. Every spec states the states or statuses, what is stored and what is derived, the events or steps with what each state refuses and its exit code, and what happens on a crash or a merged log.
- A PR whose events file conflicts on GitHub is merged locally first: GitHub does not run the union merge that the store's `.tree/.gitattributes` names, so take main's `.tree/.gitattributes` (`git checkout origin/main -- .tree/.gitattributes`), commit it if it changed, then run `git merge origin/main`, and push. Without that commit, `git merge` stops because the checked-out file would be overwritten.
- Write "Part of #N", never a closing keyword, next to an issue that stays open.
- An issue that closes with its goal not met gets a new issue first, as a placeholder that states the unmet goal, and its closing comment names that issue. A goal is never held only in a closed issue.
- Commits use the conventional format (`feat:`, `fix:`, `docs:` and so on) and carry no AI trailers.

## Build and test

```bash
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
go vet ./...
go test ./... -count=1
```

There is no CI; these commands are the only gate. `go test` includes skillcheck, which fails when a file under `skills/` names a `hyp` command that does not exist. It reads every file under `skills/`, at any depth, and runs on `skills/hyp-grill` and `skills/hyp-statechart`. Core tests read only the synthetic data under `internal/*/testdata`; [core-and-self-review.md](docs/core-and-self-review.md), section "Where the scopes touch", lists them.
