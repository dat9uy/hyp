# Owner guide

You are the owner of a Goal when you record it. The grill helps you write the Goal and places its examples. The runner builds commits and runs the lines on them. hyp calculates every state and stop from the events. It prints the acts that each role may take next. This guide tells you what you see, when a Goal stops for you, and what you do then.

[state-machine.md](state-machine.md) is the one spec of these rules. When this guide and the spec disagree, the spec wins. [principles.md](principles.md#6-who-does-what) states your role as rules.

## What you do

- **You record the Goal.** A Goal has words and up to ten lines. The words are one sentence of what you want. A `gap` line is false on H0, the Goal's base commit, and a build must make it true. A `guard` line is true on H0, and a build must keep it true. A Goal needs at least one gap line.
- **You can replace a Goal until it opens.** While the Goal is `written`, `hyp add goal G …` replaces it. After it opens, its words and lines are frozen. A new line is a new Goal. The new Goal can cite the old one with `--from G`.
- **You choose between candidates.** When two or more builds meet every line, only you can select one.
- **You merge.** hyp writes nothing to GitHub. The record branch holds the Goal, its examples and their place events, and it merges before the runner runs. The runner's PR holds the build and its select. Your merge accepts the build. You veto the build's mechanics there.
- **You produce no evidence.** You never place an example, run a line or give a verdict. A verdict comes only from the compare of an example on a commit ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).

hyp does not authenticate actors. An act under your actor name shows only that someone ran it with that name ([spec §2](state-machine.md#2-ids-actors-and-roles)).

## Record a Goal

The grill drafts the words and the lines with you. You record them yourself, with the role `owner`:

```bash
export HYP_ACTOR=<you> HYP_ROLE=owner
hyp add goal --base <H0 commit> --words "<what you want>" \
  --line "gap:<what must become true>" --line "guard:<what must stay true>"
```

The first `hyp add goal` in a repository creates the store, `.tree/`. Commit the store on the record branch. The grill then places examples on each line. The Goal opens when every line holds an example that gives its line's verdict on H0. That verdict is `failed` on a gap line and `met` on a guard line.

## What `hyp status` shows

Every command needs an actor, from `--actor` or `$HYP_ACTOR`. `hyp status` prints every Goal, and `hyp status G` prints one Goal. `--role owner` shows only your allowed acts. Each Goal prints one block:

```text
G-NXHHVA  open  runs: 1 of 3 used
stop: none
waits: none
owner may:
  (none)
grill may:
  hyp place G-NXHHVA line-1/<example>
probe-writer may:
  (none)
runner may:
  hyp run <commit> G-NXHHVA/line-1
```

- **The state** is `written`, `open` or `selected` ([spec §6](state-machine.md#6-states)).
- **`runs: N of 3 used`** counts the Hypotheses that the runner ran, H1 to H3. A Hypothesis is one build, a row tree other than H0's. H0 is not counted.
- **`stop:`** is `none`, `selected`, `two candidates` or `budget used`.
- **`waits:`** names who the Goal waits for. It is `grill` while the Goal is written, `owner` at a stop for you, and `none` otherwise.
- **`<role> may:`** lists the acts that each role may take now, as ready command lines. A role with no allowed act prints `(none)`.

The block can hold more lines:

| Line | Means |
|---|---|
| `last run: after this, G stops for the owner.` | The runner used two of three runs. |
| `split: <line> on row <tree>` | The examples of one line give different verdicts on one build. |
| `unstable:` | Two runs on the same bytes gave two verdicts. The first verdict stands. The example needs your attention. |
| `refused at replay:` | A merged log broke a guard, for example with a fourth Hypothesis in one budget. hyp skips that event and keeps it in the log. |
| `outputs:` | At a stop for you: the verdict, exit code and output of each example on each build. |
| `PR body:` | When the Goal is selected: the report of the select. |

`hyp show <actor>` prints what hyp printed to that actor, byte for byte. Use it to read what the runner saw.

## When a Goal stops

hyp calculates three stops ([spec §11](state-machine.md#11-budget-and-stops)):

| Stop | When | What you see | What you do |
|---|---|---|---|
| Selected | The Goal has a select. | The PR body. | Read the PR, then merge it or not. |
| Two candidates | Two or more builds meet every line, and none is selected. | The `outputs:` table. | Select one, or have the grill add an example. |
| Budget used | Three Hypotheses ran since the boundary, and no build meets every line. | The `outputs:` table. | Have the grill change an example, or record a new Goal. |

hyp prints the table, not a reading of it. It gives no score and no summary line.

### Selected

The runner selects when exactly one build meets every line. The report of the select is the PR body. It holds each line, the words, output, stderr, exit code and verdict of each example, and the selected commit and row. `hyp status G` prints it again while the Goal is selected. The runner pastes it into the PR.

A selected Goal refuses every further place, run and select. If your review finds a gap, the gap is a new Goal that cites this one. The grill grills it from the start.

### Two candidates

Only you can select now, and only under the actor name that added the Goal. `hyp status G --role owner` prints one ready line for each candidate:

```text
owner may:
  hyp select G-NXHHVA <tree>
```

Run the line of the build you choose. The selected commit must hold a spec under `goals/<G>/specs/`, or hyp refuses the select.

A candidate can meet the lines in a way you did not mean. Then the fix is an example. The grill can add an example that fails that build. The grill can also add a floor or a ceiling to an example.

### Budget used

The runner ran three Hypotheses, and none meets every line. Read the `outputs:` table to see where each build failed.

- If an example is wrong, the grill changes it. A new version of an example can only add a floor or a ceiling. A changed run, input or Probe is a new example key. A placement that changes an example moves the boundary of the budget. Then the runner gets three new runs.
- If the Goal itself is wrong, record a new Goal with `--from G`.

## What hyp does not have yet

The spec names acts that later Goals will build: `land`, `reset` and `retract` ([spec §8](state-machine.md#8-acts)). So no command gives the budget back without an example change. A Goal also has no end state. The end waits for a real event that needs it ([spec §14](state-machine.md#14-deferred)). `hyp help` lists every command that the binary has, and `hyp status` offers only those.
