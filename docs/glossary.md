# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.

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
The name a write records as its author, from `--actor` or `HYP_ACTOR` ([spec §1](replay-spec.md#1-names-and-locations)). Each actor appends to its own events file, so logs from several worktrees merge ([spec §6](replay-spec.md#6-merge-order)). An actor is not a role: owner is the role a confirm names (the verifier role is retired, ADR 0015).
_Avoid_: user, author (alone)

**Replay**:
Deriving every status of a tree from its ordered events.
_Avoid_: replay for re-running an agent session (that is an Experiment run with a session-replay recipe)

**Projection**:
Everything replay derives from a store's events: each node, a derived record per node kind (Hypothesis, Question, Experiment and Result) with its own status word, the Goal's state, rejections, limits and selections ([spec §8](replay-spec.md#8-projection-what-replay-returns)). Nothing stores it; every read replays (ADR 0006). In code: `replay.Projection`.
_Avoid_: state (alone), index, cache

**Goal**:
The root node of a tree, `G-`: what the investigation is for. A tree has one, and it is complete when every live DesignQuestion under it is answered ([spec §7.5](replay-spec.md#75-goal)).
_Avoid_: objective, epic

**DesignQuestion**:
A node, `Q-`, under the Goal or a Hypothesis, that asks one design question; its children are rival Hypotheses. Its state is open, reopened, answered, deferred or withdrawn ([spec §7.4](replay-spec.md#74-designquestion-state)). Withdrawn here is a hold that `dq.resume` lifts; it is not an Experiment's close.
_Avoid_: issue, decision; withdrawn for an Experiment (see Experiment status)

**Sub-question**:
A DesignQuestion whose parent is a Hypothesis, asked because it only matters if that Hypothesis holds. It is an obligation: while it is not withdrawn and not answered, deferred included, the question that selected its parent cannot be answered, and reads `awaiting_subquestions` ([spec §7.4](replay-spec.md#74-designquestion-state)). Any answer discharges it and withdrawing it releases it, whichever way it was going. That is its only effect on the parent: the parent's status does not change, a Result under it judges its own Hypotheses, its answer gives the parent no support, and the Dossier does not list it. It blocks the Goal only through this rule ([spec §7.5](replay-spec.md#75-goal)). A claim that rests on an answer is a premise, and a premise is `depends_on`.

**Hypothesis**:
A node, `H-`, under a DesignQuestion, that states one possible answer as a claim an Experiment can test. Its status comes from the verdicts of non-retracted Results, or from a supersedes edge or an abandon ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)).
_Avoid_: option, idea, solution

**Frontier**:
The section of `hyp status` that shows the open work: the tree pruned to it, from one walk over every question. A node prints only beneath its parent. A sub-question prints only from an open or reopened question and through a live rival; a deferred sub-question that a printed supported Hypothesis still owes also prints. A printed question with candidates has a `candidates (N)` line naming them, so after a re-judge it shows whether one or several are left; with none, the question's reason says so. Beneath it, an `assumes` line names each Hypothesis a candidate rests on through `depends_on`, with that Hypothesis's status, the question it answers when that is another one, and which candidates assume it, so a choice that rests on another question's answer says so. Open questions that do not print are counted in one line, so the Goal's counts stay accounted for; `hyp render --markdown` lists every question.
_Avoid_: hidden, for a question that does not print

**Awaiting owner**:
The section of `hyp status` that names the one step due from the owner on each question the Frontier prints: `choose: H…, H…` when they should choose between the candidates and the runner owes nothing on the question, or `awaiting_owner: H (scope change)` when they should confirm a scope-change select. A question the Frontier does not print asks nothing, and the requirement holds again once it prints. It also lists every open reading, which blocks nothing, and every referred limit on a candidate.

**Reading**:
How a run read owner words that a verdict rests on: a line of the Goal, recorded with `hyp record --reading`, or the owner's answer to an earlier reading, recorded with `hyp record --reading-of R-…#rN=TEXT`, which stores that reading's key as the new reading's `of`. A reading of an answer narrows it and never reopens it: the earlier reading stays closed and its answer stays in force. A reading stays open until an owner checkpoint closes it with their answer.
A reading asks what the owner's words mean, and a trade-off asks which candidate to choose. Closing a reading changes no status, because statuses come only from verdicts. An answer that rules out a candidate takes effect when a later Result records `contradicts` on it. A closed reading is owed by the runner until a Result applies it (`hyp record --applies`, from an Experiment that aims at it), and the owner is not asked to choose on its question until then. An owner who changes an answer revises it (`hyp checkpoint --role owner --revises R-…#rN`): the new words are the answer in force from then on, the earlier answer stays in history, and the reading is owed by the runner again until a later Result applies it. A revision changes no earlier event or verdict. A reading waiting under AWAITING OWNER is answered with `--closes`; one already answered is revised with `--revises`, and hyp refuses the other verb.
_Avoid_: reopen, re-answer, amend (for changing an answer)

**Re-judge**:
The step after the owner answers readings: new Results judge the live Hypotheses again under those answers. Each answer is written into a check of a new recipe version, the verdict rule is recorded on a new Experiment before the run, and old Results stay as they are. Closing a reading changes no status; a re-judge does, through its verdicts. Until a Result applies each answer, `hyp status` lists it under Owed.
_Avoid_: update the tree, recalculate (replay already derives every status; a re-judge adds Results), re-confirm (a confirm judges a select, not a Hypothesis)

**Open limits**:
The section of `hyp status` that lists every open limit, whatever the state of its question and whether the Frontier prints it. It is a list, not open work: the open work in it is what Owed and Awaiting owner name.

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
A Result's label for one Hypothesis it judges: supports, contradicts or inconclusive. Whoever runs `hyp record` types it, a person or an agent, and hyp stores only an actor name. hyp derives status from these labels alone (spec §7.1) and never reads prose. A confirm's sufficient or insufficient is a confirm verdict, a separate thing.
_Avoid_: verdict for the rule or the script that led to the label

**Verdict rule**:
The author's rule, written in a body before the run, that says which outcome gives which verdict. It is prose, or a verdict check when the rule can be computed. hyp derives status from recorded verdicts only and reads neither, so only its author and a reader of the Result can make sure prose has one reading; a verdict check's cases pin it down (ADR 0004). A script that applies the rule is an instrument.
_Avoid_: conclusion method

**Check**:
An Evidence or Standalone command of a recipe. hyp runs it the same way in the gate and in a run (ADR 0003), its known-answer cases pin its exit code and its whole stdout, and `hyp recipe check` warns per command and per verdict label (ADR 0004). What a check tests inside, such as one line of its stdout like `A1 hold`, is the author's rule: hyp gives that line no identity, no case of its own and no pin of its own.
_Avoid_: check for one such line ("adds two checks" when one command gained two lines); checker (see Instrument)

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

**Candidate**:
A live child Hypothesis of a DesignQuestion whose status is `supported`, whether or not it is stale or has an open sub-question ([spec §7.4](replay-spec.md#74-designquestion-state)). `hyp select Q` suggests a lone candidate when the ledger owes its select; two or more make the question a trade-off. Replay stores the candidates when a select applies, and a candidate that was not one then reopens the question (`new_candidate`); one that falls does not. Every candidate is a live rival (see Rival); a live rival with no Result yet, or only inconclusive ones, is not a candidate. In code: `replay.Question.Candidates`.
_Avoid_: trade-off candidate, suggested (that is what `hyp select Q` prints for a lone ready candidate), rival (a rival need not be supported)

**Trade-off**:
The state of a DesignQuestion with two or more candidates now. It is derived, never recorded ([spec §7.4](replay-spec.md#74-designquestion-state)); `hyp status` marks it `(trade-off)` on the question's candidates line. A select of one of those candidates is the owner's choice, and `hyp show` prints it `(trade-off)`.
A trade-off is between candidates, not about what a Goal line means, which is a Reading. It ends when Results leave one candidate, or when the owner chooses one, recorded as the select (ADR 0015); no confirm follows the choice. The owner's turn to choose is the ledger item `choose`, owed only when the runner owes nothing on the question and no reading on it is open ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
_Avoid_: `--tradeoff` (removed: a flag cannot record a state hyp derives), tie

**Select**:
The `dq.select` event that records one Hypothesis as a DesignQuestion's answer. It is written by whoever runs the tree, never in the owner's role. The select is the gate: `hyp select` admits it only when the ledger owes a `select` of that Hypothesis or the owner's `choose` on the question, and otherwise refuses, naming what is owed ([spec §5.1](replay-spec.md#51-write-time-refusals)). An admitted select answers the question unless it is a scope change, which an owner confirm must then count on ([spec §7.4](replay-spec.md#74-designquestion-state)). On a trade-off the owner chooses first (see Owner). In code: the question's `replay.Decision` and `replay.Projection.AdmitSelect`.
_Avoid_: pick, choose (the owner chooses; the select records it), accept, decide (alone)

**Confirm**:
The `confirm` event in which the owner judges a scope-change select `sufficient` or `insufficient` ([spec §4](replay-spec.md#4-event-types-closed-set), [spec §7.4](replay-spec.md#74-designquestion-state)). An owner confirm may also accept limits. A verifier confirm in an old log is history and applies nothing; `hyp confirm --role verifier` is refused.
_Avoid_: approve, review; verdict (alone: a confirm's answer is a confirm verdict)

**Decision**:
A DesignQuestion's current Select, the candidates when it applied, and the owner Confirms that still count.
_Avoid_: answer (`answered` is a status)

**Dossier**:
The evidence a DesignQuestion's Decision rests on: the selected Hypothesis's body, the live Hypotheses, every live Result with a verdict on one of them, and the links going out from those Hypotheses and Results, not counting `external_ref`. It is derived, never stored.
_Avoid_: basis (a basis event belongs to the Stale rule), case (a Case belongs to a known answer)

**Lapse**:
What happens to an owner Confirm when the Dossier differs from the one it judged: it stops counting, and counts again if the Dossier returns to what it judged. A select does not lapse: after it, only a new candidate reopens the question.
_Avoid_: drop, invalidate

**Verifier**:
Retired by ADR 0015. It was the role that confirmed every select, independently of the actor who made it. Old verifier confirms replay as history and apply nothing, and `hyp confirm --role verifier` is refused. Any actor may still review: a defect is recorded as a Result, a retract or a limit. In code: `event.RoleVerifier`, read only from old logs.
_Avoid_: reviewer, checker (as a role)

**Owner**:
The person hyp asks to act on real choices, and the confirm role they act in. On a trade-off they act once:

1. **Choose.** The owner chooses between the candidates once `hyp status` lists `choose:` for the question. Whoever runs the tree records the choice as the select, with the owner's words in `--why`, and `hyp select` admits it only while the ledger owes the choice. The tree holds what was chosen, but not who chose it: the select's actor is whoever wrote it, and no status reads `--why`. The choice ends the trade-off (ADR 0015).

They also answer readings, accept referred limits with an owner checkpoint, and confirm a scope-change select: an owner confirm is required only when the current select is a scope change ([spec §7.4](replay-spec.md#74-designquestion-state)). That confirm lapses when the Dossier changes, and with an owner checkpoint it is what accepts limits. hyp does not authenticate actors, so an owner act records that someone wrote it in the owner's role; that the owner stood behind it rests on the procedure. In code: `event.RoleOwner`; the requirement is `SelectRec.ScopeChange` in `replay`.
_Avoid_: user, stakeholder, approver, pick, decide (alone)

**Limit**:
A limit of the tested design that a Result found but did not test or fix, keyed `<result-id>#<n>`. It is open until a later Result closes it, or the owner accepts it with `--accept` on an owner checkpoint or on an owner confirm of the current select. A runner refers a limit no test within the Goal can close (`hyp checkpoint --refers`), which moves it from the runner's Owed items to Awaiting owner, whichever Hypothesis its Result judged. An open limit on a candidate is owed until then, so it holds the select; an admitted select warns of open limits whose Results judged other Hypotheses; and an open limit that judges the selected Hypothesis gates the owner's sufficient confirm ([spec §8](replay-spec.md#8-projection-what-replay-returns), Limits). In code: `replay.Limit`.
_Avoid_: caveat, TODO, known issue

**Owed**:
The ledger `hyp status` prints under OWED: each item of open work on a question and who owes it, the runner; the owner's items print under Awaiting owner ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)). An answered question owes only its limit, reading, referred and premise items. It is derived, never stored, and owing an item changes no status; `hyp select` reads it to admit a select. When the runner owes nothing, OWED ends with `runner owes nothing: stop and report to the owner`, the stop a run reads. In code: `replay.Projection.Owed`.
_Avoid_: todo list, backlog, next steps (those are prose; the ledger is derived)

**undecided**:
A ledger item the runner owes: a live Hypothesis that is `inconclusive`, neither supported nor contradicted, with no planned Experiment as parent or target. It holds the select of a lone candidate, because the rival may yet be one.
_Avoid_: untested (that is a proposed Hypothesis with no Experiment)

**premise**:
A live `depends_on` target of a candidate, under another question. While that question is open or reopened, neither a select nor a choice is owed on the candidate's question (the premise hold). Once it selects a Hypothesis other than the target, the runner owes a `premise` item until the edge is re-pointed, the dependent superseded, or the target selected. `hyp check` warns of two questions whose premises hold each other ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
_Avoid_: assumption (prose), sub-question (that is an obligation, not a premise)

**new_candidate**:
The reason a question with a current select reads open or reopened when a candidate now was not a candidate when the select applied, such as a rival whose deciding `contradicts` was retracted. The owner's choice is owed again, and a new select of a candidate answers it ([spec §7.4](replay-spec.md#74-designquestion-state)).
_Avoid_: lapse (a select does not lapse)

**Aim**:
A limit or answered reading an Experiment is planned to settle, named with `hyp add exp --aims`. An aim is intent: it moves the item to `in_progress`, and to `aimed_not_settled` when every aiming Experiment has a Result and none closed or applied it. Only a Result's `--closes` or `--applies` settles the item, and settling it never answers a question.
_Avoid_: target (that names a rival Hypothesis an Experiment judges), plan

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
The write-time rules the CLI checks before it appends an Experiment, a Result, a Select or a Confirm. Each finding is a refusal (exit 1), a bad-arguments finding (exit 2) or a warning that changes no exit code ([spec §5.1](replay-spec.md#51-write-time-refusals)). Replay does not repeat them, so a merged log can hold an event they would refuse. In code: `replay.State.AdmitExp`, `replay.State.AdmitResult`, `replay.Projection.AdmitSelect`, `replay.Projection.AdmitConfirm` and `replay.Finding`.
_Avoid_: validation (that is `replay.Validate`, which replay also runs on every event)

**Refusal**:
hyp's no to a write: exit 1, and nothing is appended. It comes from `replay.Validate`, which reads stored events and also runs in replay, or from Admission, which runs only at write time and may read derived state.
_Avoid_: rejection (that is what replay reports), error

**Cite**:
A token in a write command's text that names a node: a display id such as `R-BV4CWV`, or a placeholder such as `R-<new>` that stands for one. hyp refuses a write whose cite is no node of the store, with exit 2 (spec §5.1). A token after `<name>:` cites another store and is not checked.
_Avoid_: reference for this (a reference is what a flag such as `--parent` resolves)

**Round**:
One Experiment as `hyp signals` judges it by its non-retracted Results: wasted, pending, ok, or closed (superseded or withdrawn), with the signals it fired at its add. In code: `signals.Round`.
_Avoid_: iteration, cycle, step

**Signal**:
A mark of possible waste that `hyp signals` reports from the events alone: wasted, no-movement, one-rival, D1, spin or unjudged. Signals never change a status. In code: `signals.Signal`, computed by `signals.Compute`.
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
_Avoid_: loop (alone: the verifier-loop was a different signal)

**verifier-loop**:
Retired with the verifier role (ADR 0015). It was a select that collected two or more insufficient verifier verdicts in one chain.
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
hyp's record of one command it ran outside a sandbox: a standalone command, or a scripted Evidence command run again later ([sandbox spec §5.1](sandbox-spec.md#51-records)). It keeps each file argument and names no host path for anything a replay needs. In code: `cmdrecord.Record`, read by `cmdrecord.Read`; `Record.CleanRun` is the clean run that containment needs.
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
One known answer as the gate ran it: its name and its outcome, `ok` or why it failed. The sandbox record and the command record list one per case under `known_answers` ([sandbox spec §5.1](sandbox-spec.md#51-records)). In code: `cmdrecord.Case`.
_Avoid_: test case (alone), result (a Result is a tree node)

**External line**:
The `External:` line under an Evidence or Standalone command that names what the command reads or writes outside its bundle, or `none`. `hyp recipe check` warns when it is missing or still reads `TODO` (ADR 0008).
_Avoid_: dependencies line, inputs line

**Kept input**:
A file an operator argument names, which hyp copies beside the command record as `inputs/N/<base name>`, hashes and lists under `inputs` in `command.json` before the first command runs; the script gets another copy. A file over 1 MB is hashed, not kept ([sandbox spec §5.1](sandbox-spec.md#51-records)). It binds the subject when the command runs, not when the Experiment is added (ADR 0008). In code: `cmdrecord.Input`.
_Avoid_: input binding, arg hashes, inputs field (ADR 0008 rejected one on the Experiment)

**Containment**:
What hyp derives, from a Result's files, about whether it can vouch for the script that decided the verdict (spec §7.6). A Result is **uncontained** when one of §7.6's three reasons holds. The mark is read from files each time and never changes a status. Containment is about one Result's evidence. It is not self-contained, which is about a recipe relying only on its own files, and it is not sandbox isolation, which is about where commands run.
_Avoid_: isolation, contained (for a sandbox)

**Uncontained**:
A Result whose verdict script hyp cannot vouch for: a script outside what a versioned bundle hashes (one that ends in `.py` or `.sh`, has an execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`), a pinned bundle without a clean command record whose known answers passed, or a command record that matches no pin (spec §7.6).
_Avoid_: unpinned (alone)

### Sandboxes

**Sandbox**:
A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. The sandbox is the only place hyp assembles a subject: from the base, a store cut, per-run inputs and one patch for each Arm. Code under test is built in its Launch ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). It is not process isolation: commands still run on the host.
_Avoid_: container, VM, environment

**Sandbox record**:
The record of one sandbox that only hyp writes. It names no host path: the sandbox's own paths are derived from its id.
_Avoid_: sandbox manifest, metadata

**Sandbox state**:
What a sandbox can do now, derived from each recipe's setup, its lock and its record's `removing` field, and never stored: `building`, `setting_up`, `interrupted`, `ready`, `setup_failed` or `removing` ([sandbox spec §2](sandbox-spec.md#2-the-sandbox-machine)). `interrupted` means hyp stopped while making the sandbox or during its setup: it cannot be driven and can be removed. In code: `sandbox.boxState`, from `deriveState`.
_Avoid_: stored state, status of a sandbox

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
What remove copies out of a sandbox before deleting it, which scripted Evidence reads. In code: `sandbox.KeptFiles`, every name remove writes into OUT_DIR, `patches/` among them.
_Avoid_: artifacts, output (alone)

**Store cut**:
A copy of a tree's store as it stood at one event, given to a sandbox.
_Avoid_: snapshot (that is a different thing: a capture of a live session)

**Patch**:
A throwaway change over the subject's base that never lands in code, given to `hyp sandbox create` as `--patch ARM=FILE` for one Arm. An Arm has at most one; a patch is never versioned, stacked or rebased, and a change that lands is a commit used as the base instead. hyp checks that it applies to the base, records its sha256 and keeps it with the kept files. It is the subject, not a file of the recipe bundle ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)).
_Avoid_: variant (alone), diff, fork

### People and binaries

**Subject**:
The thing under test in an Experiment. A record names it by what the repository holds or the tree keeps: a commit, and for a throwaway arm one patch over it, both assembled in a sandbox ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). A binary built from it is named by its sha256.
_Avoid_: target, system under test

**Arm**:
One variant of the subject that an Experiment compares: either the base, or the base plus one Patch. It is named by a lowercase word, the ARM of `--patch ARM=FILE` ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)).
_Avoid_: variant (alone), branch

**Operator**:
The person or agent who runs an Experiment from outside the sandbox, and types its verdicts into `hyp record`.
_Avoid_: runner, driver, user

**Runner**:
The agent that runs the loop with hyp's skills. In a replay it runs inside the sandbox and writes the verdicts of the tree there. It is not the operator of the Experiment that started it.
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
