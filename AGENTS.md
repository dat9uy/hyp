# Working on hyp

This file is for any agent or person who changes hyp. It says how the repository is divided, how a change earns its way in, and how to build and test. Each section states a rule and links to the document that owns it; read that document for the detail.

## What this repository holds

- **hyp**, the binary: `cmd/hyp` and `internal/`. It stores each Goal as append-only events under `.tree/events/`, one file per actor, and calculates every state from them.
- **Four skills**: `skills/hyp-grill`, which grills and re-grills a Goal and ships the runner's run rules as its part `run-rules.md`; `skills/hyp-statechart`, with which the grill writes a Goal's model, its switches and its chart, and which ships the statechart tool (MIT) with its license and source; `skills/hyp-rethinking`, which rethinks one concept on a page drawn from checked item data, the source of a Goal's lines; and `skills/hyp-bayes-thinking`, the Bayes lens that a rethink uses to find breaks and requirements.
- **The spec** is `hyp spec`: it prints every event kind, state, act, guard and exit code, from the entries in `internal/spec` and `internal/events` that the binary acts on.
- **Docs** in `docs/`: the words in [glossary.md](docs/glossary.md), what the owner of a Goal sees and answers in [owner-guide.md](docs/owner-guide.md), the principles and the owner's role in [principles.md](docs/principles.md), why the select is the gate in [ADR 0015](docs/adr/0015-the-select-is-the-gate.md), why a skill change is a build Goal in [ADR 0021](docs/adr/0021-a-skill-change-is-a-build-goal.md).
- **The kept model** in `goals/model/`: the one model of hyp's machine. A Goal whose change moves the machine changes it in the selected commit.
- **Test data** under `internal/*/testdata`. Records made before the public release stay in the private development history.

Install the binary as [install.md](docs/install.md) describes.

## Two scopes, and why

Every item in this repository belongs to one of two scopes:

- **Core** is every concept hyp has, such as Goal, line, example, Probe, runner and select, and everything a user who installs hyp into another repository runs, loads or reads there.
- **Self-review** is a Goal of this repository whose lines are about how hyp is tested, not about a concept hyp has. An example is a Goal that checks a Probe or a test fixture of hyp itself. `main` keeps none today.

The split exists for four reasons:

1. **The skills and docs ship to other repositories.** Text inside a skill about how hyp tests itself reaches users who will never run that test. Naming the scope of each passage is how such text gets noticed and moved out.
2. **Every change says what it serves.** An issue, a plan or a PR names the concept it changes, so work on how hyp is tested is not mistaken for a feature, and a feature is not judged by how one test happens to use it.
3. **Self-review has no rule of its own.** Its examples give their line's verdict on H0 before they count, like the examples of every Goal.
4. **Visibility is a separate question.** Every item in this repository is public, so "private" never needs to mean "self-review".

Three rules keep the boundary honest: a concept stays core when self-review is what tests it; a self-review Goal does not make the concepts it uses self-review; and self-review invents no concept and no rule of its own: its examples give their line's verdict on H0 before they count, like every Goal's (the guard Kind in `hyp spec`). There are exactly two scopes. An item that fits neither has not been split far enough. Every item on `main` today is core.

## Before you change anything

- Name the scope of the change by the concept it changes, not by the Goal or test that checks it. An item that fits neither scope has not been split far enough.
- If the issue names a Goal, run `hyp status G` first, with `HYP_ACTOR` set. It shows the Goal's state, its stop, who it waits for, the runs used and the allowed acts of each role.
- Use the words in [glossary.md](docs/glossary.md). A self-review Goal uses only core words; a new word it needs gets a glossary entry that maps it onto core concepts, and a word that maps onto none is a sign that a core concept is missing.

## Improving hyp with hyp

hyp is changed the way it asks its users to decide things: with a Goal whose gap and guard lines each hold examples, runs of rival Hypotheses against those examples, and a select hyp records only for the one candidate met on every line. The loop:

1. **A real event points at hyp.** In a real run a refusal or a warning fires, a Goal reaches a stop it cannot leave, `hyp status` lists an event under `refused at replay`, or the owner flags a run. A change is never proposed on speculation.
2. **The event is recorded where it fits.** A fact an agent found is a finding from the moment it is found, with `hyp finding add`, and names no Goal. An issue tracks the process toward a want: it names the event and its scope, and holds the owner's decisions and the split into Goals. A Goal is the snapshot of the want. The issue closes when each want it tracks is held by an added Goal.
3. **A Goal holds the question.** The owner adds it with `hyp add goal`: words, gap and guard lines, and a `--finding F` for each finding the Goal is about. The grill drafts one or more examples on each line in a named draft in the clone's trace and tries each on H0 with `hyp try --draft`, to show it gives the verdict its line needs. The owner adds the Goal with `--draft` and that name, and the Goal is open at once; a later example is the owner's `hyp place`. The rules are in `hyp spec`.
4. **The runner runs, and hyp shows the select.** `hyp run` runs a line on a commit, within a budget of three Hypotheses, and writes only to the clone's trace. When those Results show one candidate that is met on every line and holds a spec, `hyp status` shows the Goal selected on it, with no event, and prints the PR body. `hyp status` prints the allowed acts of each role as ready command lines. Every change is a Goal.
5. **The change passes its gate.**
   - A failure that can be detected from the log or from git goes into the binary as a refusal or a warning, not into skill text. Its gate is `go test`.
   - A change to `skills/`, a new skill included, is a build Goal like any code change: its lines, examples whose Probes the Goal writes and that give their line's verdict on H0 before they are placed, and the select. How a skill is tested is that Goal's choice.
   - A prose-only skill change, one that changes no command, flag, default, cap or exit code a skill names, ships as a plain PR that names its event and the owner's decision. Its gate is `go test`.
   - A refactor that changes no output, no golden and no line `hyp spec` prints ships as a plain PR that names its event and the owner's decision. Its gate is `go test` with every golden unchanged.

   [principles.md](docs/principles.md) §9 explains why the loop is guarded while it runs instead of proven in advance.
6. **The PR body is what `hyp status G` prints under `PR body:`** once the Goal is selected: each line's examples, and the selected commit and row. The owner's `hyp land G COMMIT` then carries the evidence into git: the commit, its tree, its row and each example's verdict.

## Run controls

- Skill text reaches every run that loads it. Editing `skills/` changes what later runs that install skills from this repository measure, even when the edit looks harmless.
- A runner also sees the investigated repository's own agent files. The runner works in a checkout of that repository, and a Probe runs in a copy of the commit's subject, so any `AGENTS.md` or `CLAUDE.md` in the commit is part of the runner's environment. This file is one of them whenever hyp is the host.

## Rules

- An example is self-contained: its Probe reads nothing outside itself except the subject and the environment variables its `external` section names. A fixed input that is not the subject is an `input/<path>` section of the example. A new version of an example only adds a floor or a ceiling; any other change is a new example key (the guard Adds only and the Probe rule in `hyp spec`).
- Name a binary by its sha256, not by the commit it was built from.
- hyp exits 1 for a rule refusal, 2 for bad arguments and 3 for an environment problem, such as a missing tool or a Probe that failed or timed out (the exit codes in `hyp spec`).
- Skill text cites no ADR and no file under `docs/`: the skills are copied into other repositories without them. A skill states its rule inline, and the ADR behind it names the skill file. Removing such a citation is a product fix.
- hyp has one spec: its entries in `internal/spec` and `internal/events`, which the binary acts on and `hyp spec` prints. A change to a rule changes its entry in the same PR, and no doc restates the rule. The entries state each state and what it allows, each event kind with where it is stored and whether an end freezes it, each act with its roles and its guards in order, each guard with its codes, each exit code, and the rules of the store, the merge order, the budget, the example file and the Probe.
- Write "Part of #N", never a closing keyword, next to an issue that stays open.
- An issue that closes with its goal not met first gets a record that holds the goal: a written Goal with the goal as a line when its lines are known, or else a new issue that states it. Its closing comment names that record. A goal is never held only in a closed issue.
- Commits use the conventional format (`feat:`, `fix:`, `docs:` and so on) and carry no AI trailers.

## Build and test

```bash
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
go vet ./...
go test ./... -count=1
```

There is no CI; these commands are the only gate. `go test` includes skillcheck, which fails when a file under `skills/` names a `hyp` command that does not exist. It reads every file under `skills/`, at any depth, so it runs on all four skills. Core tests read only core test data: the tests of the binary build their stores in temporary git repositories and compare against goldens under `internal/channel/testdata` and fixtures under `internal/example/testdata`; other tests read only data under `internal/*/testdata`. A copy of the repository that holds only core passes `go test`.
