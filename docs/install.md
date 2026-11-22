# Installing hyp and its skill

`hyp` is a single static binary. Its skill is a plain Markdown directory that any
harness able to load a `SKILL.md` can use. Claude Code is the only harness hyp
has been verified on.

There are two ways to set hyp up: to use it in your own project, or to develop
hyp itself in this repository.

## Use hyp in your project

### Install the binary

```bash
go install github.com/dat9uy/hyp/cmd/hyp@latest
hyp version   # hyp 0.2.0-dev (events v2)
```

Make sure Go's install directory (`$(go env GOPATH)/bin` unless `GOBIN` is set)
is on your `PATH`. To build from a clone instead, use the command under
[Develop hyp](#develop-hyp). The binary is named `hyp` because `tree` collides
with the common `tree` utility and shell aliases.

hyp needs `git`, and it works only in a repository whose object format is
`sha1`. A Probe that the `hyp-grill` skill ships is a POSIX `sh` script.

### Copy the skill into the project

`main` holds two skills. `skills/hyp-grill` grills the owner into a Goal and
re-grills a Goal that stopped or was selected. Its part `run-rules.md` holds the
rules a runner follows. `skills/hyp-statechart` is what the grill writes a
Goal's model with: a TLA+ spec, one switch per potential Hypothesis, and a chart. Its
statechart tool needs Node 22.18 or later (or Bun or Deno); its first run
installs Java and TLC with mise. From your project's root, with a checkout of hyp's
`main` at `<hyp>`:

```bash
mkdir -p .claude/skills
cp -R <hyp>/skills/hyp-grill <hyp>/skills/hyp-statechart .claude/skills/
git add .claude/skills
```

Committing the copy means every session in the project, and every person who
clones it, loads the same skill. To update, re-run `go install ...@latest` and
copy the skill again, so the binary and the skill stay the same version.

The skills of hyp 0.1.0 are kept on the `v1` branch. They call commands that
the binary on `main` does not have. To use them, build the binary from a
checkout of the `v1` branch as well.

### Per-repo setup

hyp has no `init` command. The first `hyp add goal` creates the store, `.tree/`,
at the top level of the repository. The store holds `events/` and a `.gitattributes`
that sets a union merge for the events files. Commit `.tree/` with the work it
describes. Every cell, Result and output stays in the clone's trace under
`~/.hyp/traces/`, not in the store.

Every command needs an actor, and every act that writes also needs a role:

```bash
export HYP_ACTOR=<your-name> HYP_ROLE=owner   # roles: owner, grill, probe-writer, runner
hyp status
```

Each agent session sets its own actor. hyp records what it prints to each actor
in a trace outside the repository, one folder per clone under `~/.hyp/traces`.
`hyp show <actor>` prints that trace. `hyp prune` lists the traces of clones
that no longer exist, and `hyp prune --remove` removes them.

[owner-guide.md](owner-guide.md) tells the owner what to do next.

## Develop hyp

### Build the binary

From this repository:

```bash
mkdir -p ~/.local/bin
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
hyp version   # hyp 0.2.0-dev (events v2) <commit>, plus -dirty for a modified checkout
```

Make sure `~/.local/bin` is on your `PATH`. A build from a clone prints the
commit it was built from; a `go install` build does not.

### Link the skill globally

Run from the root of a checkout of `main`. Claude Code:

```bash
mkdir -p ~/.claude/skills
for s in skills/hyp-*; do ln -s "$PWD/$s" ~/.claude/skills/"$(basename "$s")"; done
```

Symlinks make your sessions load the working text under `skills/`, so an edit is
live at once, and hyp's repository carries no second copy of the skill under
`.claude/`. Do not commit a `.claude/skills` directory here.

### Other harnesses (unverified)

Codex:

```bash
mkdir -p ~/.codex/skills
for s in skills/hyp-*; do ln -s "$PWD/$s" ~/.codex/skills/"$(basename "$s")"; done
```

pi:

```bash
mkdir -p ~/.pi/agent/skills
for s in skills/hyp-*; do ln -s "$PWD/$s" ~/.pi/agent/skills/"$(basename "$s")"; done
```
