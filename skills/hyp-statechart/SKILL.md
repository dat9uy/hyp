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
- `<name>.cfg`: the check config, with every variant FALSE and every rule
  under `INVARIANTS` (or `PROPERTIES`), except a type rule (step 2).
- `<name>-view.mjs`: the view module, with the `print()` the examples are
  matched against.
- `variants/<variant>.cfg`: one variant config per variant. The tool writes
  them; do not write them by hand. hyp also reads the variant configs of an
  older spec folder from `specs/switches/`.
- `tried/<variant>/<line>/<example>.txt`: the verdict line `hyp try --worktree`
  printed for that example on the variant's patch.
- `chart.mmd`: the chart the tool wrote from the check config.

## The kept model

A repository keeps one model of its machine in `goals/model/`: a TLA+ module
and its check config. It is no Goal's folder, so a runner's own commit may
change it. A Goal whose change moves the machine changes the kept model
instead of writing the whole machine again: the commit edits `goals/model/`,
and the Goal's spec folder holds the variant configs and the chart. Run the
tool on the kept model with `--out` in the Goal's spec folder, such as
`--out goals/<G>/specs/out`: it then writes the variant configs under
`goals/<G>/specs/variants/` and none under `goals/model/`.

Name the source of a kept rule with its Goal: `\* source: G-XXXXXX:line-2`
for a line of a Goal, or `\* source: model only` for a rule no line holds.
At the selected stop, `hyp status G` lists each step and rule that the
selected commit adds or changes in the kept model: a step by its command, a
rule with the words of the line its source names, or `no line`.

At select, hyp compares the kept model of the commit with H0's, whatever
commit the clone has checked out. Each rule that is new, or whose
definition text differs, needs a variant config under
`goals/<G>/specs/variants/` that names it; a rule the commit leaves as it was
needs none. A change to the kept model also needs a chart in the Goal's spec
folder. A Goal whose spec folder holds its own whole model is checked as the
start of this skill says.

## Steps, prints and outside reads

A step is an action that changes a variable: a side effect of the system, such
as a write to its store. A read or a refusal is a print, not a step. A variable
changes only inside an action that a command takes.

What a step reads from outside the system is that step's input. Examples are
git history, a file another program wrote, a clock or a web service. When a
rule or a later step needs the input, the step stores it in a variable, as the
value or as the hash of the bytes. No rule or print reads the outside system
again.

Test the model: rewrite or remove the outside system's history. No variable
may change.

## How to write the model

1. **Write each rule the owner agrees to as an INVARIANT.** Take the rules from
   the Goal's words, its lines and the owner's decisions. A rule can be about a
   screen, a result, a stored value or a message. If the model has no variable
   for a rule, add the variable. Never weaken a rule to make it hold.
2. **Name the source of each rule.** Put a line comment right above each rule:
   `\* source: line-2` for a line of the Goal, `\* source: owner decision 3`
   for a decision of the owner, or `\* source: model only` for another rule the
   model needs that a variant can break. The tool lists each rule with no
   source and each line of the Goal that no rule names.

   A type rule, such as `TypeOK`, says only that each variable holds a value of
   its type. A variant changes what the system does, not the types of its
   variables, so no variant breaks a type rule. At select, hyp refuses a commit
   whose check config (any `.cfg` outside `variants/`) checks a rule that no
   variant config names. So define a type rule in the module if you want it,
   but do not list it in any check config. Leaving it out is not weakening a
   rule: it describes the model, not the system.
3. **One variant per potential Hypothesis.** Model each other behaviour that a
   cheap build could ship, such as the old behaviour, behind a boolean CONSTANT,
   such as `WriteOutputsEarly`, that is FALSE in the check config. A variant is
   a build a runner could make in a few lines, not a wish list.
4. **Model every act a line allows, in every state it allows it.** A line
   that lets an act happen again (for example "an unselected Goal may be
   selected again") needs that step from every state the line names, ended states
   included. The view prints what the system prints on that step, ready lines
   too, so an example can see it. A step the model leaves out is a behaviour
   that no variant and no example reaches, and the chart does not show it.
5. **Write `Next` as a disjunction of the actions**, such as
   `YouPressPlay \/ YouPressStop`, so that each step has its own name.
   Put a line comment right above each action that a command of the system
   takes, with that command: `\* command: hyp select`. The tool reads the
   commands each example's Probe and run section call, and takes the action
   of each command in the order they call them.
6. **Write the view's `print()`.** `print(action, from, to)` returns what the
   system prints on that step, in the form the examples' floors use. The output
   of an example's walk is the prints of its run section's steps, joined by
   newlines, and the tool matches it against the example's floors and ceilings
   as hyp does. The prints of the steps the Probe takes before the run section
   are not matched, as hyp matches only the run section's output.
7. **Run the tool** on the check config, into a new `--out` folder outside the
   repository, or, for the kept model, a new folder in the Goal's spec folder
   that you delete before you commit:

   ```sh
   <skill dir>/scripts/statechart SPEC.tla --config SPEC.cfg --view VIEW.mjs --out DIR
   ```

   The tool needs Node 22.18 or later (or Bun 1.1.5, or Deno 2.0). Its first
   run installs TLC and Java with mise and needs the network. When the spec is
   in `goals/<G>/specs/`, the tool runs the model once for each variant and
   writes `goal.md` beside `report.md`:
   - one variant config per variant under `specs/variants/`, which lists
     exactly the rules that variant breaks;
   - each rule no variant breaks, printed as not shown to fail. hyp refuses
     this check config at select, so the tool names each such rule, says that
     hyp select refuses the check config, and exits 3 where it would exit 0.
     Before you commit the model, fix the model or add the variant that
     breaks the rule.
     If no variant can break a rule of a line or of an owner decision, take
     the rule out of the check config and name it in the hand-off;
   - for each example, its walk: the actions of the commands its Probe and
     its run section call, in the order they call them, and what the run
     section prints. A command of another program, such as `echo` or
     `mkdir`, takes no step. With no `command:` comment in the module, the
     walk is the shortest walk whose prints meet the example, which can set
     up another case than the example's Probe;
   - each example the tool cannot predict, and why: its Probe or run section
     calls a command of the system that the model has no step for; no
     print of its walk gives its floors, under the check config or any
     variant, so it is outside the model's prints; or the model has more than
     one start case and the Probe writes files before its run section, so
     the tool cannot see the example's start case. The tool lists each such
     example to try on every variant;
   - for each example and each variant, whether the example is met under that
     variant, predicted from the model alone by following the example's walk,
     and for each variant the examples that separate it: each example whose
     tried verdict on the variant's patch is a fail, and each untried example
     the tool predicts to fail. A tried met stands over a predicted fail;
   - one part for each variant that no example separates, naming its rules,
     the step no example takes and what the model and the variant print there.

   The chart has a budget of 500 states, counted by TLC on the config as
   written; above it the tool draws only the bug paths, or no chart when no
   rule breaks. Never make the checked model smaller to fit the budget, such
   as by dropping a cause, an action or a rule: the rules would then hold on
   fewer states than the hand-off claims. Write a second config, such as
   `chart.cfg`, with smaller constants for the chart, and keep both configs.
   hyp reads every `.cfg` outside `variants/` as a check config, so the chart
   config lists only rules of the check config. A run on any config of the
   spec folder rewrites `specs/variants/` from that config, so run the check
   config last, or run it again after the chart config, so that the variant
   configs come from the check config.

   [references/tool.md](references/tool.md) says how to label the chart with a
   view module and how to read the report.
8. **Commit the chart in the spec folder.** Copy the `chart.mmd` the tool wrote
   to `goals/<G>/specs/chart.mmd`, beside the model and its configs: the check
   config's chart, or the chart config's when the check config is over the
   budget. The select report prints every chart of the spec folder, so the
   owner reads it in the PR.

## Try what the tool predicts

A prediction is only a claim about the code. Make it real. Trying a variant
is not running a rival: a rival is a second commit run with `hyp run`, which
uses runs of the budget and can stop the Goal at `Two candidates`.

1. Write the variant's build as a patch on the code, in a scratch worktree of
   the record branch. The patch is a potential Hypothesis: hyp has no record
   of it.
2. Try the placed examples on the variant's patch, one at a time, with
   `hyp try G LINE/EXAMPLE --worktree`. It records nothing and uses no run.
   Try first the example the tool predicts the variant to fail. With no
   prediction, try first the examples on the source lines of the rules the
   variant breaks. Try each example the tool lists to try on every variant
   on each variant's patch.
3. Stop trying a variant at its first failed example. That failed try
   separates the variant, and the report counts it so. A variant whose every
   try is met is tried on every placed example.
4. Write the verdict line each try printed, such as `verdict: met, exit 0`, in
   `specs/tried/<variant>/<line>/<example>.txt`, and run the tool again. The
   report sets each tried cell beside its prediction and names each cell where
   they differ. Where they differ, find the cause. When the model is wrong, fix
   the model, not an example. When the cause is the shortest walk of a model
   with no `command:` comment (step 7), which set up another case than the
   example's Probe, the model is right and the tried verdict stands over the
   prediction: tag the actions with their commands. Say which cause it was in
   one line of the hand-off.
5. When a tried patch meets every example, the examples do not separate the two
   builds. The grill asks the owner one plain question about it, at the example
   level: the case, what each build prints, and which output is wanted.

Before the examples exist, a prediction may reach the owner only as a
direction: the grill shows the step no example takes, with what the model and
the variant print there, each marked "model print, not built", and the owner
chooses the wanted output. It is never a verdict: the example written from
that choice is tried on H0 with `hyp try` before it is placed and must agree
with its line, and a model result never counts at select.

Remove each patch when you are done; the record branch holds the model, its
configs, the tried verdicts and the examples.

## Hand-off

Add to the grill's hand-off:

- the chart, as a `mermaid` block;
- for each variant, the rules its config names, and whether its patch failed an
  example or met every example, as `hyp try` printed it;
- one line for each tried verdict that differed from its prediction: what was
  fixed in the model, or which case the tool's walk set up instead of the
  example's;
- each rule other than a type rule left out of the check config, and why;
- when a chart config exists, the config and the state count behind each
  result;
- what the model leaves out;
- the command that makes the chart again.

## Done when

- `goals/<G>/specs/` holds the model, its check config, the variant configs the
  tool wrote, and `chart.mmd`.
- Each variant was tried as a patch with `hyp try --worktree` up to its first
  failed example, or on every placed example when each try was met.
- Each tried verdict that differs from its prediction was traced to its cause
  (step 4 of "Try what the tool predicts") and is named in the hand-off.

## Resources

- [scripts/statechart](scripts/statechart): the statechart tool, run as above.
- [references/tool.md](references/tool.md): the view module, the report and
  the chart's limits.
- [LICENSE](LICENSE) and [SOURCE](SOURCE): the tool's MIT license and where it
  comes from.
