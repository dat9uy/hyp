---
name: hyp-statechart
description: "Write the model of a hyp Goal: a TLA+ spec of the states the change moves through, one variant per potential Hypothesis, and a Mermaid chart, checked with the statechart tool this skill ships, which says which rules each variant breaks and which examples it would meet. Use when the grill of a hyp Goal writes the Goal's model in its spec folder, or when the builds the examples do not separate need to be found before the runner builds. Not for running a Goal's lines on commits: the runner does that."
---

# hyp statechart

The grill uses this skill to write a Goal's model. When the record branch
holds no spec and the runner writes the model, the runner uses this skill the
same way, variant tries included. The model is a tool that finds potential
Hypotheses: each variant in it is a different build a runner could make in a
few lines. The statechart tool, not the writer of the model, says which rules
each variant breaks and which examples each variant would meet. A tried
variant that meets every example shows two builds the examples do not
separate. Which output is wanted is the owner's choice. The chart shows the
owner, in the owner's words, how the system moves between states.

hyp reads the model's files and runs no model checker. At select, when the
spec folder holds a model, hyp refuses the commit while a rule the check config
checks has no variant config that names it, or while the folder holds no chart.
A model result never counts at select: only the examples on the commit count.

## When to use

- The grill drafts the model of a Goal on its record branch, after the owner
  consented to the Goal's words and lines and before the examples are final.
- A Goal needs its potential Hypotheses found before a runner builds it.
- Not for: a Goal whose change moves through no states worth a chart. Its spec
  folder can hold a plain text spec instead; hyp then checks no model.

## Where the files go

The model lives in the Goal's spec folder, `goals/<G>/specs/`, committed on the
record branch so that every runner commit holds it:

- `<name>.tla`: the TLA+ module. A `.tla` file is what makes the folder a model.
- `<name>.cfg`: the check config, with every variant FALSE and every rule under
  `INVARIANTS` (or `PROPERTIES`).
- `<name>-view.mjs`: the view module, with the `print()` the examples are
  matched against.
- `variants/<variant>.cfg`: one variant config per variant. The tool writes
  them; do not write them by hand. hyp also reads the variant configs of an
  older spec folder from `specs/switches/`.
- `tried/<variant>/<line>/<example>.txt`: the verdict line `hyp try --worktree`
  printed for that example on the variant's patch.
- `chart.mmd`: the chart the tool wrote from the check config.

## How to write the model

1. **Write each rule the owner agrees to as an INVARIANT.** Take the rules from
   the Goal's words, its lines and the owner's decisions. A rule can be about a
   screen, a result, a stored value or a message. If the model has no variable
   for a rule, add the variable. Never weaken a rule to make it hold.
2. **Name the source of each rule.** Put a line comment right above each rule:
   `\* source: line-2` for a line of the Goal, `\* source: owner decision 3`
   for a decision of the owner, or `\* source: model only` for a rule the model
   needs, such as `TypeOK`. The tool lists each rule with no source and each
   line of the Goal that no rule names.
3. **One variant per potential Hypothesis.** Model each other behaviour that a
   cheap build could ship, such as the old behaviour, behind a boolean CONSTANT,
   such as `WriteOutputsEarly`, that is FALSE in the check config. A variant is
   a build a runner could make in a few lines, not a wish list.
4. **Write `Next` as a disjunction of the actions**, such as
   `YouPressPlay \/ YouPressStop`, so that each step has its own name.
5. **Write the view's `print()`.** `print(action, from, to)` returns what the
   system prints on that step, in the form the examples' floors use. The output
   of a walk is the prints of its steps, joined by newlines, and the tool
   matches it against each example's floors and ceilings as hyp does.
6. **Run the tool** on the check config, into a new `--out` folder outside the
   repository:

   ```sh
   <skill dir>/scripts/statechart SPEC.tla --config SPEC.cfg --view VIEW.mjs --out DIR
   ```

   The tool needs Node 22.18 or later (or Bun 1.1.5, or Deno 2.0). Its first
   run installs TLC and Java with mise and needs the network. When the spec is
   in `goals/<G>/specs/`, the tool runs the model once for each variant and
   writes `goal.md` beside `report.md`:
   - one variant config per variant under `specs/variants/`, which lists
     exactly the rules that variant breaks;
   - each rule no variant breaks, printed as not shown to fail: fix the model,
     or take the rule out of the check config and name it in the hand-off;
   - for each example, a walk of the model whose prints meet the example's
     floors and ceilings, or that no walk of the model does;
   - for each example and each variant, whether the example is met under that
     variant, predicted from the model alone by following the example's walk,
     and for each variant the examples that separate it;
   - one part for each variant that no example separates, naming its rules,
     the step no example takes and what the model and the variant print there.

   [references/tool.md](references/tool.md) says how to label the chart with a
   view module and how to read the report.
7. **Commit the chart in the spec folder.** Copy `chart.mmd` that the tool wrote
   for the check config to `goals/<G>/specs/chart.mmd`, beside the model and its
   configs. The select report prints every chart of the spec folder, so the
   owner reads it in the PR.

## Try what the tool predicts

A prediction is only a claim about the code. Make it real. Trying a variant
is not running a rival: a rival is a second commit run with `hyp run`, which
uses runs of the budget and can stop the Goal at `Two candidates`.

1. Write the variant's build as a patch on the code, in a scratch worktree of
   the record branch. The patch is a potential Hypothesis: hyp has no record
   of it.
2. For each variant the tool predicts to meet every example, run each placed
   example on its patch with `hyp try G LINE/EXAMPLE --worktree`. It records
   nothing and uses no run.
3. For each other variant, try one example the tool predicts to fail on its
   patch, as a spot check of the model.
4. Write the verdict line each try printed, such as `verdict: met, exit 0`, in
   `specs/tried/<variant>/<line>/<example>.txt`, and run the tool again. The
   report sets each tried cell beside its prediction and names each cell where
   they differ. Where they differ, fix the model, not an example, and say so in
   one line of the hand-off.
5. When a tried patch meets every example, the examples do not separate the two
   builds. The grill asks the owner one plain question about it, at the example
   level: the case, what each build prints, and which output is wanted. A
   prediction alone, never tried, never reaches the owner.

Remove each patch when you are done; the record branch holds the model, its
configs, the tried verdicts and the examples.

## Hand-off

Add to the grill's hand-off:

- the chart, as a `mermaid` block;
- for each variant, the rules its config names, and whether its patch failed an
  example or met every example, as `hyp try` printed it;
- one line for each tried verdict that differed from its prediction, and what
  was fixed in the model;
- each rule left out of the check config, and why;
- what the model leaves out;
- the command that makes the chart again.

## Done when

- `goals/<G>/specs/` holds the model, its check config, the variant configs the
  tool wrote, and `chart.mmd`.
- Each variant the tool predicts to meet every example was tried as a patch
  with `hyp try --worktree`, and one predicted-fail example of each other
  variant was tried; no tried verdict differs from its prediction.

## Resources

- [scripts/statechart](scripts/statechart): the statechart tool, run as above.
- [references/tool.md](references/tool.md): the view module, the report and
  the chart's limits.
- [LICENSE](LICENSE) and [SOURCE](SOURCE): the tool's MIT license and where it
  comes from.
