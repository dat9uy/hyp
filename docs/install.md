# Installing hyp and its skills

`hyp` is a single static binary; the two skills are plain Markdown directories
that any harness able to load a `SKILL.md` can use. Claude Code is the only
harness hyp has been verified on.

There are two ways to set hyp up: to use it in your own project, or to develop
hyp itself in this repository.

## Use hyp in your project

### Install the binary

```bash
go install github.com/dat9uy/hyp/cmd/hyp@latest
hyp version   # hyp 0.1.0 (events v1)
```

Make sure Go's install directory (`$(go env GOPATH)/bin` unless `GOBIN` is set)
is on your `PATH`. To build from a clone instead, use the command under
[Develop hyp](#develop-hyp). The binary is named `hyp` because `tree` collides
with the common `tree` utility and shell aliases.

### Copy the skills into the project

From your project's root, with a clone (or download) of hyp at `<hyp>`:

```bash
mkdir -p .claude/skills
cp -R <hyp>/skills/hyp-prototype <hyp>/skills/hyp-verification .claude/skills/
git add .claude/skills
```

Committing the copies means every session in the project, and every person who
clones it, loads the same skills. To update, re-run `go install ...@latest` and
copy the skills again, so the binary and the skills stay the same version.

### Per-repo setup

In the repository you investigate:

```bash
hyp init                         # creates .tree/events, .tree/evidence, .tree/.gitignore
git add .tree/.gitignore .tree/events
hyp bundle new NAME@1            # optional: a first versioned procedure, in recipes/ beside .tree/
```

Commit `.tree/events/` and `.tree/evidence/` with the work they describe.
hyp keeps no index (ADR 0006); the `.tree/.gitignore` still ignores the
`index.sqlite` that older hyp binaries write.
Each agent session sets its own actor, for example
`export HYP_ACTOR=claude-<topic>-1`.

## Develop hyp

### Build the binary

From this repository:

```bash
mkdir -p ~/.local/bin
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
hyp version   # hyp 0.1.0 (events v1) <commit>, plus -dirty for a modified checkout
```

Make sure `~/.local/bin` is on your `PATH`. A build from a clone prints the
commit it was built from; a `go install` build does not.

### Link the skills globally

Run from this repository's root. Claude Code:

```bash
mkdir -p ~/.claude/skills
ln -s "$PWD/skills/hyp-prototype" ~/.claude/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.claude/skills/hyp-verification
```

Symlinks make your sessions load the working text under `skills/`, so an edit is
live at once, and hyp's repository carries no second copy of the skills under
`.claude/`. Do not commit a `.claude/skills` directory here. The skill-text replay
recipe (`recipes/skill-text-arms@4`) does not depend on either: it deletes the
clone's `.claude` and installs the skills from `skills/` into its runner itself.

### Other harnesses (unverified)

Codex:

```bash
mkdir -p ~/.codex/skills
ln -s "$PWD/skills/hyp-prototype" ~/.codex/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.codex/skills/hyp-verification
```

pi:

```bash
mkdir -p ~/.pi/agent/skills
ln -s "$PWD/skills/hyp-prototype" ~/.pi/agent/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.pi/agent/skills/hyp-verification
```
