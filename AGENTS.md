# Working on hyp

This file is for any agent or person who changes hyp. It says how the repository is divided, how a change earns its way in, and how to build and test. Each section states a rule and links to the document that owns it; read that document for the detail.

## What this repository holds

- **hyp**, the binary: `cmd/hyp` and `internal/`. It stores an investigation as append-only events, in a git repository on the hidden ref `refs/hyp/record` (the record, which `hyp sync` shares), and derives every status from them.
- **Five skills**: `skills/hyp-prototype` (run an investigation as a tree of rival Hypotheses), `skills/hyp-verification` (run one Experiment from a recipe), and the owner flow: `skills/hyp-owner-grill` (grill the owner into a Goal), `skills/hyp-goal-run` (run a Goal as an unattended `/goal`, whose prompt `hyp prompt` prints) and `skills/hyp-owner-answer` (answer what a stopped run left for the owner).
- **Docs** in `docs/`: the rules in [replay-spec.md](docs/replay-spec.md) and [sandbox-spec.md](docs/sandbox-spec.md), the words in [glossary.md](docs/glossary.md), what the owner of a Goal sees and answers in [owner-guide.md](docs/owner-guide.md), the principles and the owner's role in [principles.md](docs/principles.md), the loop's guards in [how-hyp-improves-itself.md](docs/how-hyp-improves-itself.md), why the select is the gate in [ADR 0015](docs/adr/0015-the-select-is-the-gate.md), why a skill change is a build Goal in [ADR 0021](docs/adr/0021-a-skill-change-is-a-build-goal.md), and the scope map in [core-and-self-review.md](docs/core-and-self-review.md).
- **Test data** under `internal/*/testdata`, such as the synthetic store in `internal/testutil/testdata/synth`. The recipes and trees kept to test hyp are `recipes/id-citation-arms@1` and `@2` with `investigations/id-citation` (whose Arm patches are in its `arms/`), `recipes/recipe-base-arms@1` with `investigations/recipe-base`, `investigations/goal-prompt/recipes/hyp-owner-grill-eval@1` and `@2` with `investigations/goal-prompt`, and the recipes `recipes/authoring-stops@2` and `recipes/recipe-authoring-arms@4`, whose trees are not published. The core-scope trees `investigations/checkpoint-store-design` and `investigations/checkpoint-store` decided and built hyp's shared store (ADR 0020) and pin no recipe. Records made before the public release stay in the private development history.

Install the binary and link the skills as [install.md](docs/install.md) describes.

## Two scopes, and why

Every item in this repository belongs to one of two scopes:

- **Core** is every concept hyp has, such as recipe, sandbox, store cut, runner and instrument, and everything a user who installs hyp into another repository runs, loads or reads there.
- **Self-review** is only the recipes this repository keeps to test hyp, plus the inputs only they use. Those kept today are listed in the scope map.

The split exists for four reasons:

1. **The skills and docs ship to other repositories.** Text about a test recipe's versions inside a skill reaches users who will never run that recipe. Naming the scope of each passage is how such text gets noticed and moved out.
2. **Every change says what it serves.** An issue, a plan or a PR names the concept it changes, so work on a test recipe is not mistaken for a feature, and a feature is not judged by how one test recipe happens to use it.
3. **A self-review recipe is an instrument.** It must give its known answers before its verdicts count. That is the one reason a tree may ask about a self-review recipe itself.
4. **Visibility is a separate question.** Every item in this repository is public, so "private" never needs to mean "self-review".

Three rules keep the boundary honest: a concept stays core when self-review is what tests it; a self-review recipe does not make its kind self-review; and self-review invents no concept of its own. The test, the rules, the map of every item and the table of self-review words are in [core-and-self-review.md](docs/core-and-self-review.md), section "The test".

## Before you change anything

- Name the scope of the change by the concept it changes, not by the recipe that tests it. An item that fits neither scope has not been split far enough.
- If the issue names a tree, run `hyp status` in it first. It shows the goal, what is open, what awaits the owner and the last checkpoint.
- Use the words in [glossary.md](docs/glossary.md). A new self-review word needs an entry in the scope map's word table.

## Improving hyp with hyp

hyp is changed the way it asks its users to decide things: with a tree of rival Hypotheses, Experiments run from recipes, and a select hyp admits only when the ledger owes it. The loop:

1. **A real event points at hyp.** In a real run the ledger shows an item that does not settle, a refusal or warning fires, `hyp signals` reports a design mark (one-rival or D1), or the owner flags a round. A change is never proposed on speculation.
2. **An issue names the event and its scope.**
3. **A tree holds the question.** A new Goal goes in the store that holds the Goal it continues; otherwise run `hyp init` in a new `investigations/<topic>/`. Add the goal, the design question and its Hypotheses (two rivals only when a scenario separates them on a line; otherwise one, with the base as its rival), and pre-register each Experiment's rubric before running it: a pinned rubric bundle, or prose in its body. The [hyp-prototype skill](skills/hyp-prototype/SKILL.md) describes each step.
4. **Experiments run from a pinned, versioned recipe.** The [hyp-verification skill](skills/hyp-verification/SKILL.md) and [RECIPE-FORMAT.md](skills/hyp-verification/RECIPE-FORMAT.md) describe Launch, Doctor, Drive, Evidence and Cleanup. hyp admits a select only when the ledger owes it; the owner decides trade-offs.
5. **The change passes its gate.**
   - A failure that can be detected from the tree goes into the binary as a refusal or a warning, not into skill text. Its gate is `go test`.
   - A change to `skills/`, a new skill included, is a build Goal like any code change: its lines, its own test or an instrument the Goal builds, red on base, and the select. How a skill is tested is that Goal's choice.
   - A prose-only skill change, one that changes no command, flag, default, cap or exit code a skill names, ships as a plain PR that names its event and the owner's decision. Its gate is `go test`.

   [how-hyp-improves-itself.md](docs/how-hyp-improves-itself.md) explains why the loop is guarded while it runs instead of proven in advance.
6. **The PR cites the Result ids that decided it.**

## Experiment controls

- Skill text reaches replayed runs. Editing `skills/` changes what later runs that install skills from this repository measure, even when the edit looks harmless.
- A runner also sees the investigated repository's own agent files. `hyp sandbox create` clones the host repository and checks out the pinned `--base` commit, so any `AGENTS.md` or `CLAUDE.md` in that commit is part of the runner's environment. This file is one of them whenever hyp is the host.

## Rules

- A recipe is self-contained: it relies on nothing but its own files, hyp and the subject. A fixed input that is not the subject, such as a fixture or a cases file, is a file of the bundle, and a file outside it is the subject or a read or write the command's `External:` line names. A frozen recipe version is never edited; a change is a new `@<N>`.
- An id a Result cites must appear in the files it cites. Name a binary by its sha256, not by the commit it was built from.
- hyp exits 1 for a rule refusal, 2 for bad arguments and 3 for an environment problem, such as a busy port or a missing tool ([sandbox-spec.md](docs/sandbox-spec.md) §1.3).
- Skill text cites no ADR and no file under `docs/`: the skills are copied into other repositories without them. A skill states its rule inline, and the ADR behind it names the skill file. Removing such a citation is a product fix.
- A state machine has one spec, and a change to the machine updates that spec in the same PR. Every spec states the states or statuses, what is stored and what is derived, the events or steps with what each state refuses and its exit code, and what happens on a crash or a merged log.
- A PR whose events file conflicts on GitHub is merged locally first: GitHub does not run the union merge that each store's `.tree/.gitattributes` names, so take main's `.tree/.gitattributes` (`git checkout origin/main -- <store>/.tree/.gitattributes`), commit it if it changed, then run `git merge origin/main`, and push. Without that commit, `git merge` stops because the checked-out file would be overwritten.
- Write "Part of #N", never a closing keyword, next to an issue that stays open.
- An issue that closes with its goal not met gets a new issue first, as a placeholder that states the unmet goal, and its closing comment names that issue. A goal is never held only in a closed issue.
- Commits use the conventional format (`feat:`, `fix:`, `docs:` and so on) and carry no AI trailers.

## Build and test

```bash
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
go vet ./...
go test ./... -count=1
```

There is no CI; these commands are the only gate. `go test` includes skillcheck, which fails when a file under `skills/` names a `hyp` command that does not exist. Core tests read only the synthetic data under `internal/*/testdata`; [core-and-self-review.md](docs/core-and-self-review.md), section "Where the scopes touch", lists them.
