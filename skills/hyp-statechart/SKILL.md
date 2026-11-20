---
name: hyp-statechart
description: "Write the model of a hyp Goal: a TLA+ spec of the states the change moves through, one switch per cheap wrong build, and a Mermaid chart, checked with the statechart tool this skill ships. Use when the grill of a hyp Goal writes the Goal's model in its spec folder, or when a wrong build of a Goal needs to be found before the runner builds. Not for running a Goal's lines on commits: the runner does that."
---

# hyp statechart

The grill uses this skill to write a Goal's model. The model is a tool that
finds missing examples: each switch in it is a cheap wrong build, and a wrong
build that passes every example shows an example the Goal still needs. The
chart shows the owner, in the owner's words, how the system moves between
states.

hyp reads the model's files and runs no model checker. At select, when the
spec folder holds a model, hyp refuses the commit while a rule the check config
checks has no switch config that names it, or while the folder holds no chart.
A model result never counts at select: only the examples on the commit count.

## When to use

- The grill writes the model of a Goal on its record branch, after the owner
  consented to the Goal's words and lines.
- A Goal needs its wrong builds found before a runner builds it.
- Not for: a Goal whose change moves through no states worth a chart. Its spec
  folder can hold a plain text spec instead; hyp then checks no model.

## Where the files go

The model lives in the Goal's spec folder, `goals/<G>/specs/`, committed on the
record branch so that every runner commit holds it:

- `<name>.tla`: the TLA+ module. A `.tla` file is what makes the folder a model.
- `<name>.cfg`: the check config, with every switch FALSE and every rule under
  `INVARIANTS` (or `PROPERTIES`).
- `switches/<switch>.cfg`: one switch config per switch.
- `chart.mmd`: the chart the tool wrote from the check config.
- `report.md`: the tool's report for the check config, if you keep it.

## How to write the model

1. **Write each rule the owner agrees to as an INVARIANT.** Take the rules from
   the Goal's words, its lines and the owner's decisions. A rule can be about a
   screen, a result, a stored value or a message. If the model has no variable
   for a rule, add the variable. Never weaken a rule to make it hold.
2. **One switch per cheap wrong build.** Model each wrong or old behaviour that
   a cheap build could ship behind a boolean CONSTANT switch, such as
   `WriteOutputsEarly`, that is FALSE in the check config. A switch is a wrong
   build a runner could make in a few lines, not a wish list.
3. **Every rule is named by a switch config.** For each switch, write
   `switches/<switch>.cfg`: the check config with only that switch TRUE, whose
   `INVARIANTS` section lists only the rules that switch breaks. Run the tool
   with each switch config, and keep a rule in the list only when the report
   says it breaks. Each rule must be named by at least one switch config. A rule
   that no switch breaks can hold whatever the system does: fix the model, or
   take the rule out of the check config and name it in the hand-off as not
   shown to fail. A type rule such as `TypeOK` counts as a rule too.
4. **Write `Next` as a disjunction of the actions**, such as
   `YouPressPlay \/ YouPressStop`, so that each step has its own name.
5. **Run the tool** on the check config, then on each switch config, each into
   a new `--out` folder outside the repository:

   ```sh
   <skill dir>/scripts/statechart SPEC.tla --config SPEC.cfg [--view VIEW.ts] --out DIR
   ```

   The tool needs Node 22.18 or later (or Bun 1.1.5, or Deno 2.0). Its first
   run installs TLC and Java with mise and needs the network.
   [references/tool.md](references/tool.md) says how to label the chart with a
   view module and how to read the report.
6. **Commit the chart in the spec folder.** Copy `chart.mmd` that the tool wrote
   for the check config to `goals/<G>/specs/chart.mmd`, beside the model and its
   configs. The select report prints every chart of the spec folder, so the
   owner reads it in the PR.

## Try each switch as a patch

A switch in the model is only a claim about the code. Make each one real:

1. Write the switch's wrong build as a patch on the code, in a scratch
   worktree of the record branch.
2. Run each placed example of the Goal on that patch with
   `hyp try G LINE/EXAMPLE --worktree`. It records nothing and uses no run.
3. When some example fails on the patch, the examples catch that wrong build.
4. When a wrong build passes every example, the examples miss it. Report it to
   the owner with the switch, the patch and the outputs, and propose a
   separating example: one that fails on that wrong build and is met on a
   right one. The owner decides whether it is placed or listed under
   "Not covered".

Remove each patch when you are done; the record branch holds only the model and
the examples.

## Hand-off

Add to the grill's hand-off:

- the chart, as a `mermaid` block;
- for each switch, the rules its config names, and whether its patch failed an
  example or passed every example;
- each rule left out of the check config, and why;
- what the model leaves out;
- the command that makes the chart again.

## Done when

- `goals/<G>/specs/` holds the model, its check config, a switch config for
  every rule the check config checks, and `chart.mmd`.
- Each switch was tried as a patch with `hyp try --worktree`, and each wrong
  build that passed every example is in the hand-off for the owner.

## Resources

- [scripts/statechart](scripts/statechart): the statechart tool, run as above.
- [references/tool.md](references/tool.md): the view module, the report and
  the chart's limits.
- [LICENSE](LICENSE) and [SOURCE](SOURCE): the tool's MIT license and where it
  comes from.
