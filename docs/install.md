# Installing hyp and its skills

`hyp` is a single static binary; the two skills are plain Markdown directories
that any harness able to load a `SKILL.md` can use.

## Build the binary

From this repository:

```bash
mkdir -p ~/.local/bin
CGO_ENABLED=0 go build -o ~/.local/bin/hyp ./cmd/hyp
hyp version   # hyp 0.1.0 (events v1) <commit>, plus -dirty for a modified checkout
```

Make sure `~/.local/bin` is on your `PATH`. The binary is named `hyp` because
`tree` collides with the common `tree` utility and shell aliases.

## Link the skills

Run from this repository's root. Symlinks keep the skills in sync with this
repository.

Claude Code:

```bash
mkdir -p ~/.claude/skills
ln -s "$PWD/skills/hyp-prototype" ~/.claude/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.claude/skills/hyp-verification
```

Codex (unverified until the Phase 8 dogfood):

```bash
mkdir -p ~/.codex/skills
ln -s "$PWD/skills/hyp-prototype" ~/.codex/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.codex/skills/hyp-verification
```

pi (unverified until the Phase 8 dogfood):

```bash
mkdir -p ~/.pi/agent/skills
ln -s "$PWD/skills/hyp-prototype" ~/.pi/agent/skills/hyp-prototype
ln -s "$PWD/skills/hyp-verification" ~/.pi/agent/skills/hyp-verification
```

## Per-repo setup

In the repository you investigate:

```bash
hyp init                         # creates .tree/events, .tree/evidence, .tree/.gitignore
git add .tree/.gitignore .tree/events
hyp recipe new NAME@1            # optional: a first versioned procedure, in recipes/ beside .tree/
```

Commit `.tree/events/` and `.tree/evidence/` with the work they describe.
hyp keeps no index (ADR 0006); the `.tree/.gitignore` still ignores the
`index.sqlite` that older hyp binaries write.
Each agent session sets its own actor, for example
`export HYP_ACTOR=claude-<topic>-1`.
