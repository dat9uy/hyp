# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The spec ([replay-spec.md](replay-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.

## Language

### Trees

**Tree**:
One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
_Avoid_: graph, log (the log is its events), project

**Store**:
The `.tree/` directory of a repository: one events file per actor and the evidence files Results point at ([spec §1](replay-spec.md#1-names-and-locations)). hyp finds it by walking up from the current directory, as git finds `.git`. In code: package `store`.
_Avoid_: database, index (hyp keeps none, ADR 0006)

**Event**:
One line of `.tree/events/<actor>.jsonl`: an id, a time, an actor, a type from a closed set and its payload ([spec §3](replay-spec.md#3-event-envelope), [spec §4](replay-spec.md#4-event-types-closed-set)). Events are only appended, and every status is derived from them. In code: `event.Event`.
_Avoid_: entry, record (a record is what hyp writes for a sandbox or a command)

**Actor**:
The name a write records as its author, from `--actor` or `HYP_ACTOR` ([spec §1](replay-spec.md#1-names-and-locations)). Each actor appends to its own events file, so logs from several worktrees merge ([spec §6](replay-spec.md#6-merge-order)). An actor is not a role: verifier and owner are the roles a confirm names.
_Avoid_: user, author (alone)

**Replay**:
Deriving every status of a tree from its ordered events.
_Avoid_: replay for re-running an agent session (that is an Experiment run with a session-replay recipe)

**Projection**:
Everything replay derives from a store's events: each node and its status, the Goal's state, rejections, limits, selections, and the derived records of Experiments and Results ([spec §8](replay-spec.md#8-projection-what-replay-returns)). Nothing stores it; every read replays (ADR 0006). In code: `replay.Projection`.
_Avoid_: state (alone), index, cache

**Goal**:
The root node of a tree, `G-`: what the investigation is for. A tree has one, and it is complete when every live DesignQuestion under it is answered ([spec §7.5](replay-spec.md#75-goal)).
_Avoid_: objective, epic

**DesignQuestion**:
A node, `Q-`, under the Goal or a Hypothesis, that asks one design question; its children are rival Hypotheses. Its state is open, reopened, answered, deferred or withdrawn ([spec §7.4](replay-spec.md#74-designquestion-state)). Withdrawn here is a hold that `dq.resume` lifts; it is not an Experiment's close.
_Avoid_: issue, decision; withdrawn for an Experiment (see Experiment status)

**Hypothesis**:
A node, `H-`, under a DesignQuestion, that states one candidate answer as a claim an Experiment can test. Its status comes from the verdicts of non-retracted Results, or from a supersedes edge or an abandon ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)).
_Avoid_: option, idea, solution

**Rival**:
Another Hypothesis under the same DesignQuestion. A live rival is live and not contradicted ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)); the one-rival refusal and the one-rival, D1 and unjudged signals read the live rivals of an Experiment's parent at its add ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add).
_Avoid_: alternative, competitor

**Experiment**:
A node, `E-`, under a Hypothesis, that plans one bounded test. Its body pre-registers the verdict rule, and it pins the recipe it runs from (a recipe ref and its hash) or `none` ([spec §4](replay-spec.md#4-event-types-closed-set)). Its derived state is its Experiment status.
_Avoid_: test, trial (a trial is one scored repeat inside it), run (that is a status)

**Experiment status**:
An Experiment is planned, run, superseded or withdrawn (spec §7.6). It is run while it has a non-retracted Result, which wins over a close. Otherwise it is superseded when another Experiment under the same question has a live `supersedes` edge to it (`hyp supersede NEW --old OLD`), withdrawn when `hyp withdraw` closed it with nothing to replace it, and planned otherwise. A superseded or withdrawn Experiment is closed: `hyp status` drops it, `hyp record` refuses a Result under it, and `hyp signals` counts its round as neither pending nor wasted (ADR 0008). In code: `replay.Exp`, built once per projection in `Projection.Exps`.
_Avoid_: abandoned for an Experiment (abandoning is a judgement about a Hypothesis's claim); withdrawn for a DesignQuestion's hold (see DesignQuestion)

**Target**:
A sibling Hypothesis an Experiment also tests besides its parent, given with `--target` ([spec §4](replay-spec.md#4-event-types-closed-set), [spec §5](replay-spec.md#5-refusals) rule 5). A Result may judge only the parent and the targets.
_Avoid_: subject (the thing under test is the Subject; a target is a Hypothesis)

**Result**:
A node, `R-`, under an Experiment, that records what a run observed: a verdict per Hypothesis it judges, an observation, evidence pointers and any limits ([spec §4](replay-spec.md#4-event-types-closed-set)). It is recorded or retracted, and a retracted Result no longer counts. In code: `replay.Res`, which holds its status, pins, questionable flag and uncontained reasons, in `Projection.Results`.
_Avoid_: finding, outcome, run

**Verdict**:
A Result's label for one Hypothesis it judges: supports, contradicts or inconclusive. hyp derives status from these labels alone (spec §7.1) and never reads prose. A confirm's sufficient or insufficient is a confirm verdict, a separate thing.
_Avoid_: verdict for the rule or the script that led to the label

**Verdict rule**:
The author's rule, written in a body before the run, that says which outcome gives which verdict. It is prose, or a verdict check when the rule can be computed. hyp derives status from recorded verdicts only and reads neither, so only its author and the verifier can make sure prose has one reading; a verdict check's cases pin it down (ADR 0004). A script that applies the rule is an instrument.
_Avoid_: conclusion method

**Verdict check**:
A check (an Evidence or Standalone command) whose known answers print `verdict: LABEL=WORD` lines. Its cases are the verdict rule's specification, and `hyp recipe check` warns when a label lacks a supports or a contradicts case (ADR 0004).

**Evidence pointer**:
One `evidence` entry of a Result: a URL, a path under `.tree/evidence/`, or `recipes/NAME@N/FILE`, which hyp looks up like a recipe ref and pins by its bundle hash ([spec §4](replay-spec.md#4-event-types-closed-set), [spec §7.6](replay-spec.md#76-experiments-and-results)). `event.EvidenceRef` reads the ref of the last kind.
_Avoid_: attachment, link (alone)

**Pin**:
A recipe ref and the bundle hash an event recorded for it when it was written: an Experiment's recipe, or an evidence pointer into a versioned bundle ([spec §7.6](replay-spec.md#76-experiments-and-results)). In code: `replay.Pin`, listed per Result in `replay.Res.Pins`.
_Avoid_: lock, version (alone)

**Questionable**:
A Result one of whose pins no longer matches the current recipe hashes, because its bundle changed or is missing ([spec §7.6](replay-spec.md#76-experiments-and-results), `questionable_evidence`). It is a flag only and never changes a status. In code: `replay.Res.Questionable`.
_Avoid_: stale (that is the dependency rule), invalid

**Stale**:
A live Hypothesis or DesignQuestion whose basis may no longer hold: a `depends_on` target is not supported, is stale itself, or changed since the basis event ([spec §7.3](replay-spec.md#73-stale-snapshot-rule-the-minimal-form-of-truth-maintenance)). It is shown as a cause chain, such as `H-5 stale ← H-3 contradicted ← R-9`. In code: `replay.StaleCause`.
_Avoid_: outdated, questionable (that is about recipe hashes)

**Select**:
The `dq.select` event that picks one Hypothesis as a DesignQuestion's answer; the verifier, and on a trade-off the owner, then confirm it ([spec §7.4](replay-spec.md#74-designquestion-state)). In code: the question's `replay.Sel`.
_Avoid_: choose, accept, decide (alone)

**Confirm**:
The `confirm` event in which the verifier or the owner judges a select `sufficient` or `insufficient` ([spec §4](replay-spec.md#4-event-types-closed-set), [spec §7.4](replay-spec.md#74-designquestion-state)). An owner confirm may also accept limits.
_Avoid_: approve, review; verdict (alone: a confirm's answer is a confirm verdict)

**Verifier**:
The role that confirms every select, independently of the actor who made it: a verifier confirm by the select's own actor is refused ([spec §5](replay-spec.md#5-refusals) rule 9). In code: `event.RoleVerifier`.
_Avoid_: reviewer, checker

**Owner**:
The person who decides trade-offs, and the confirm role they act in. An owner confirm is required when a select is a trade-off or a scope change, or when the question has two or more trade-off candidates ([spec §7.4](replay-spec.md#74-designquestion-state)), and only an owner confirm accepts limits. In code: `event.RoleOwner`.
_Avoid_: user, stakeholder, approver

**Limit**:
A limit of the tested design that a Result found but did not test or fix, keyed `<result-id>#<n>`. It is open until a later Result closes it or an owner confirm on the current select accepts it, and an open limit that judges the selected Hypothesis gates the owner's sufficient confirm ([spec §8](replay-spec.md#8-projection-what-replay-returns), Limits). In code: `replay.Limit`.
_Avoid_: caveat, TODO, known issue

**depends_on**:
An edge from a Hypothesis or DesignQuestion to a Hypothesis whose support it relies on. Stale follows it ([spec §7.3](replay-spec.md#73-stale-snapshot-rule-the-minimal-form-of-truth-maintenance)), and it may not close a cycle in the combined status graph ([spec §7.2](replay-spec.md#72-combined-status-graph-for-cycle-refusal)).
_Avoid_: requires, blocks

**supersedes**:
An edge that replaces its target. Between two Hypotheses under one question, a rival claim replaces the old one, which reads superseded ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)). Between two Experiments, a new procedure replaces one with no live Result, which is then closed ([spec §7.6](replay-spec.md#76-experiments-and-results), ADR 0008).
_Avoid_: replaces, obsoletes; superseded for an Experiment closed with no replacement (that is withdrawn)

**external_ref**:
An edge from a node to something outside the tree, such as `gh:976` or a URL ([spec §4](replay-spec.md#4-event-types-closed-set)). It changes no status.
_Avoid_: link (alone), reference

**Admission**:
The write-time rules the CLI checks before it appends an Experiment or a Result. Each finding is a refusal (exit 1), a bad-arguments finding (exit 2) or a warning that changes no exit code ([spec §5.1](replay-spec.md#51-write-time-refusals)). Replay does not repeat them, so a merged log can hold an event they would refuse. In code: `replay.State.AdmitExp`, `replay.State.AdmitResult` and `replay.Finding`.
_Avoid_: validation (that is `replay.Validate`, which replay also runs on every event)

**Cite**:
A token in a write command's text that names a node: a display id such as `R-BV4CWV`, or a placeholder such as `R-<new>` that stands for one. hyp refuses a write whose cite is no node of the store, with exit 2 (spec §5.1). A token after `<name>:` cites another store and is not checked.
_Avoid_: reference for this (a reference is what a flag such as `--parent` resolves)

**Round**:
One Experiment as `hyp signals` judges it by its non-retracted Results: wasted, pending, ok, or closed (superseded or withdrawn), with the signals it fired at its add. In code: `signals.Round`.
_Avoid_: iteration, cycle, step

**Signal**:
A mark of possible waste that `hyp signals` reports from the events alone: wasted, no-movement, one-rival, D1, spin, verifier-loop or unjudged. Signals never change a status. In code: `signals.Signal`, computed by `signals.Compute`.
_Avoid_: alert, error

**wasted**:
A round with Results and no close, none of whose Results changed the status of its parent, a target or a live rival at its add.
_Avoid_: failed, useless (a wasted round can still record a correct Result)

**pending**:
A round whose Experiment is not closed and has no non-retracted Result yet. It is neither wasted nor ok.
_Avoid_: planned (that is an Experiment status), open

**no-movement**:
Two wasted rounds in a row, in the order of their last Results. It names the question's current selection when there is one.
_Avoid_: stall (alone), stuck

**one-rival**:
A round whose parent had live rivals at its add and whose targets name none of them ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add). `hyp add exp` refuses such an Experiment unless `--target none` is passed ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: no-rival, single-rival

**D1**:
A round whose parent or a target was not a live rival at its add. `hyp add exp` warns `D1:` for the same Hypotheses, because both read one rule ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: dead rival, invalid target

**spin**:
Three Results in a row that judge a Hypothesis as their Experiment's parent and each leave its status unchanged. Rivals an Experiment only targets do not count.
_Avoid_: loop (alone: the verifier-loop is a different signal)

**verifier-loop**:
A select that collected two or more insufficient verifier verdicts in one chain, or a re-verify request with no new Result, retract or link under the question since the last verdict ([spec §7.4](replay-spec.md#74-designquestion-state)).
_Avoid_: spin (that is about a Hypothesis's status)

**unjudged**:
A target, live at its Experiment's add, that no non-retracted Result of the Experiment judged ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add).
_Avoid_: untested (that is a DesignQuestion reason: every live child is proposed)

**Trial**:
One independent, scored repeat of an attempt whose outcome can vary, such as a fresh agent given one task or one session replay from a cut. A verdict rule counts trials ("T correct of 6"). A trial can take a whole Experiment run, as a session replay does, or share one with other trials. hyp does not record trials; the Experiment's body or method says how many and how each is scored.
_Avoid_: run when the count is meant (say "3 of 5 trials"); trial for a fixed case a script checks, a probe that fixes the trial command, or an instrument check

### Recipes

**Recipe**:
A versioned, hash-pinned description of how to run one kind of Experiment, together with the scripts it runs.
_Avoid_: harness script, driver, wrapper

**Bundle**:
A recipe stored as a directory, so it can hold scripts beside its description. A bundle is identified by its ref and hash, not by the directory it sits in (ADR 0005).
_Avoid_: recipe folder, package

**Recipe ref**:
A recipe's name and version written together, the form an Experiment cites. hyp opens every copy of a bundle by the ref it holds (ADR 0005). In code: `event.Ref`, parsed once by `event.ParseRef`; its JSON is the string `name@N`.
_Avoid_: recipe id, recipe tag

**Recipe library**:
The set of recipe versions a repository keeps, which Experiments cite instead of copying.
_Avoid_: recipe registry, recipe cache

**Frozen**:
A recipe version a tree pins by hash; its bytes never change.
_Avoid_: locked, released

**Draft**:
A recipe version no tree pins yet; it may be edited in place.
_Avoid_: unpinned, work in progress

**Harness recipe**:
The one recipe in a stack that an Experiment pins, because it defines what is measured.
_Avoid_: main recipe; harness (alone, that is the agent runtime)

**Environment recipe**:
A recipe stacked before the harness recipe that only sets up something the subject needs, and is never pinned. A sandbox record does not tell it apart: it is the stack entry the Experiment does not pin.
_Avoid_: dependency recipe, helper recipe

**Stack**:
The ordered set of whole recipes that one sandbox sets up together.
_Avoid_: recipe chain, composition

**Command**:
One `Run:` line in a recipe section (a step or Standalone): a script inside the bundle plus fixed words, which hyp runs.
_Avoid_: hook, task, action

**Standalone**:
The recipe section whose commands run when no sandbox exists, for example to capture an Experiment's inputs or to classify kept files.
_Avoid_: tools, prepare, utilities

**Command record**:
hyp's record of one command it ran outside a sandbox: a standalone command, or a scripted Evidence command run again later ([spec §9.4](replay-spec.md#94-records)). In code: `cmdrecord.Record`, read by `cmdrecord.Read`; `Record.CleanRun` is the clean run that containment needs.
_Avoid_: run record (collides with live run)

**Self-contained**:
A recipe that relies on nothing but its own files, hyp and the subject, wherever its commands run. A fixed input that is not the subject, such as a fixture or a cases file, is a file of the bundle, and so is the code that reads it. A file outside the bundle is the subject, or a read or write that the command's `External:` line names (ADR 0008).
_Avoid_: sandboxed (self-contained does not mean "runs in the sandbox")

**Instrument**:
A recipe bundle with an Evidence or Standalone command; it must have known answers (ADR 0002). A script outside a bundle is not one hyp can vouch for (see Uncontained).
_Avoid_: checker

**Measure**:
The script or program whose output a Result's verdict counts. It lives in the Experiment's pinned instrument; an Experiment with no measure declares that its verdicts are decided by reading (`--decided-by reading`). An agent arm is a Trial, not a measure. Code has no other word for it: an Experiment is measured when its pinned recipe is an instrument (`recipe.Recipe.IsInstrument`), and otherwise declares `decided_by: reading`.
_Avoid_: measurement script

**Known answer**:
A case under `known-answers/NAME/` that hyp runs before every run of its instrument (ADR 0002). It runs the check exactly as the run does and compares its exit code, stdout and the files it leaves under `$HYP_OUT` (ADR 0003).
_Avoid_: fixture (alone), golden test

**Case**:
One known answer as the gate ran it: its name and its outcome, `ok` or why it failed. The sandbox record and the command record list one per case under `known_answers` ([spec §9.4](replay-spec.md#94-records)). In code: `cmdrecord.Case`.
_Avoid_: test case (alone), result (a Result is a tree node)

**External line**:
The `External:` line under an Evidence or Standalone command that names what the command reads or writes outside its bundle, or `none`. `hyp recipe check` warns when it is missing or still reads `TODO` (ADR 0008).
_Avoid_: dependencies line, inputs line

**Input binding**:
The sha256 of each file an operator argument names, which hyp records as `arg_sha256` in `command.json` before the first command runs ([spec §9.4](replay-spec.md#94-records)). It binds the subject when the command runs, not when the Experiment is added (ADR 0008). In code: `cmdrecord.Record.ArgSHA256`.
_Avoid_: inputs field (ADR 0008 rejected one on the Experiment)

**Containment**:
What hyp derives, from a Result's files, about whether it can vouch for the script that decided the verdict (spec §7.6). A Result is **uncontained** when one of §7.6's three reasons holds. The mark is read from files each time and never changes a status. Containment is about one Result's evidence. It is not self-contained, which is about a recipe relying only on its own files, and it is not sandbox isolation, which is about where commands run.
_Avoid_: isolation, contained (for a sandbox)

**Uncontained**:
A Result whose verdict script hyp cannot vouch for: a script outside what a versioned bundle hashes (one that ends in `.py` or `.sh`, has an execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`), a pinned bundle without a clean command record whose known answers passed, or a command record that matches no pin (spec §7.6).
_Avoid_: unpinned (alone)

### Sandboxes

**Sandbox**:
A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. The sandbox is the only place hyp assembles a subject: from the base, a store cut, per-run inputs and, **[not built: #8]**, one patch for each Arm. Code under test is built in its Launch ([spec §9.6](replay-spec.md#96-the-subject)). It is not process isolation: commands still run on the host.
_Avoid_: container, VM, environment

**Sandbox record**:
The record of one sandbox that only hyp writes.
_Avoid_: sandbox manifest, metadata

**Step**:
One stage of running an Experiment: Launch, Doctor, Drive, Evidence or Cleanup. In a recipe, each step is a section of that name. In code: `recipe.Step`, whose constants also name Standalone.
_Avoid_: phase, stage, hook

**Launch**:
The step that sets up the subject and the recipe's resources inside a new sandbox.

**Doctor**:
The step that checks, before any drive, that Launch left the sandbox ready.
_Avoid_: preflight, health check

**Drive**:
The step that exercises the subject while the sandbox exists.
_Avoid_: run, execute

**Cleanup**:
The step that stops what a recipe's Launch started, before hyp deletes the sandbox.
_Avoid_: teardown

**Evidence**:
The step that turns what happened into findings. When it is scripted, it reads only the kept files after the sandbox is gone, and can be re-run.
_Avoid_: check (as a step name), verify

**Kept files**:
What remove copies out of a sandbox before deleting it, which scripted Evidence reads. In code: `sandbox.KeptFiles`, every name remove writes into OUT_DIR.
_Avoid_: artifacts, output (alone)

**Store cut**:
A copy of a tree's store as it stood at one event, given to a sandbox.
_Avoid_: snapshot (that is a different thing: a capture of a live session)

**Patch**:
A throwaway change over the subject's base that never lands in code, given to `hyp sandbox create` as `--patch ARM=FILE` for one Arm. An Arm has at most one; a patch is never versioned, stacked or rebased, and a change that lands is a commit used as the base instead. hyp checks that it applies to the base, records its sha256 and keeps it with the kept files. It is the subject, not a file of the recipe bundle ([spec §9.6](replay-spec.md#96-the-subject)). **[not built: #8]**
_Avoid_: variant (alone), diff, fork

### People and binaries

**Subject**:
The thing under test in an Experiment. A record names it by what the repository holds or the tree keeps: a commit, and for a throwaway arm one patch over it, both assembled in a sandbox ([spec §9.6](replay-spec.md#96-the-subject)). A binary built from it is named by its sha256. **[not built: #8]**
_Avoid_: target, system under test

**Arm**:
One variant of the subject that an Experiment compares: either the base, or the base plus one Patch. It is named by a lowercase word, the ARM of `--patch ARM=FILE` ([spec §9.6](replay-spec.md#96-the-subject)).
_Avoid_: variant (alone), branch

**Operator**:
The person or agent who runs an Experiment from outside the sandbox.
_Avoid_: runner, driver, user

**Runner**:
The agent that runs the loop with hyp's skills and writes a tree's verdicts. It is not the operator.
_Avoid_: drive agent, subject

**Harness**:
The agent runtime that loads a skill and hosts the runner, such as Claude Code or Codex.
_Avoid_: harness for a recipe or a binary (say harness recipe, operator hyp)

**Sandbox hyp**:
The hyp binary a sandbox pins, which a runner inside the sandbox calls: the operator hyp, copied for the sandbox. In code: `sandbox.SandboxHyp`.
_Avoid_: subject hyp (true only when hyp itself is under test), pinned hyp (alone)

**Operator hyp**:
The hyp binary the operator runs, which sets up the sandbox and runs the recipe's commands.
_Avoid_: harness hyp, harness (for a binary)

### Scopes of this repository

**Core**:
Every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A concept stays core when self-review is what tests it.
_Avoid_: product, the tool, hyp proper

**Self-review**:
The recipes this repository keeps to test hyp, whatever repository the runner works on, with the inputs only they use. Investigating hyp is an activity, not a scope: a tree about hyp takes the scope of its question, and a tree about another project is a self-review input.
_Avoid_: dogfood (alone), meta, hyp-self (that is one tree)
