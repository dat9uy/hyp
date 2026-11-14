# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them. For the owner of a Goal, [owner-guide.md](owner-guide.md) introduces these words in the order the owner meets them.

Each entry names a larger set and the one property that makes the term a subset of it. Where a term has subsets, the term comes first and its subsets follow it directly.

## Language

### Trees

**Tree**:
The view of one Goal and of every node whose parents lead to it, kept as events in a store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)). A store holds many Goals, so it holds many trees (ADR 0020).
_Avoid_: graph, log (the log is its events), project

**Store**:
The `.tree/` directory of a repository: one events file per actor and the evidence files that Results point at ([spec §1](replay-spec.md#1-names-and-locations)). hyp finds it by a walk up from the current directory that stops at the repository root (ADR 0022). In code: package `store`.
_Avoid_: database, index (hyp keeps none, ADR 0006)

**Replay**:
The derivation of every status of a store from its events, in merge order ([spec §7](replay-spec.md#7-derived-state-forward-pass)).
_Avoid_: replay for a re-run of an agent session (that is an Experiment run with a session-replay recipe)

**Projection**:
Everything that replay derives from a store's events: each node's derived record, each Goal's state, the rejections, limits and selects ([spec §8](replay-spec.md#8-projection-what-replay-returns)). Nothing stores it, and every read replays (ADR 0006). In code: `replay.Projection`.
_Avoid_: state (alone), index, cache

**Event**:
One line of `.tree/events/<actor>.jsonl`: an id, a time, an actor, a type from a closed set, and its payload ([spec §3](replay-spec.md#3-event-envelope), [spec §4](replay-spec.md#4-event-types-closed-set)). Events are only appended, and every status derives from them. In code: `event.Event`.
_Avoid_: entry, trace (a trace is what hyp writes about a run of commands)

**Answer**:
An event that records the owner's act: it answers readings, or accepts or returns referred limits ([spec §4](replay-spec.md#4-event-types-closed-set), ADR 0020). Each key carries the owner's words. It has no node and no why, and owner fields and runner fields never mix in one event ([spec §5](replay-spec.md#5-refusals) rule 30). `hyp answer` writes it. In code: `event.Answer`.
_Avoid_: owner checkpoint (history), approval, confirm (retired)

**Checkpoint**:
An event in which the runner reports on one Goal: a summary, refers that hand limits to the owner, or both ([spec §4](replay-spec.md#4-event-types-closed-set)). `hyp status` prints the Goal's last one. An old owner checkpoint is read-only history that replay still applies. In code: `event.Checkpoint`.
_Avoid_: owner checkpoint (history; see Answer), report

**Select**:
An event, `dq.select`, that records one Hypothesis as the answer of a DesignQuestion. The select is the gate: `hyp select` admits it only when the ledger owes it ([spec §5.1](replay-spec.md#51-write-time-refusals), ADR 0015). The ledger owes a `select` of that Hypothesis, or the owner's `choose` on the question. Whoever runs the tree writes it, never in the owner's role. In code: the question's `replay.Decision` and `replay.Projection.AdmitSelect`.
_Avoid_: pick, choose (the owner chooses; the select records it), accept, decide (alone)

**Goal**:
A node, `G-`, with no parent, that states what an investigation is for. Its lines come with its add, and no one changes it after that. A new purpose is a new Goal, which can be in the same store (ADR 0020). A Goal is complete when every live top-level question is answered ([spec §7.5](replay-spec.md#75-goal)). The owner reads "met" as a complete Goal with no line under OWED, and hyp stores no word for it.
_Avoid_: objective, epic, done, met (as a status word)

**Discovery Goal**:
A Goal whose Hypotheses are claims: what to build, or why something behaves as it does. Rivals answer each DesignQuestion, and Experiments separate them. hyp stores no Goal kind: the same events, rules and ledger serve each kind (ADR 0020).
_Avoid_: investigation Goal, research Goal; discovery kind (hyp stores none)

**Design record**:
A discovery Goal whose questions the owner settles by argument in a session, with no lines. The session is its evidence. No run checks its candidates, so its cases go to the lines of the build Goal that follows it ([hyp-owner-grill](../skills/hyp-owner-grill/SKILL.md#derive-the-lines)).
_Avoid_: design Goal, decision log

**Build Goal**:
A Goal whose Hypotheses are commits. Each commit is a Hypothesis whose subject is that commit, and a fix is a new Hypothesis that supersedes it. A change to a skill is a build Goal, except a prose-only change, which ships as a plain PR (ADR 0021). A run works it as [hyp-goal-run, "Build Goals"](../skills/hyp-goal-run/SKILL.md#build-goals) says.
_Avoid_: delivery Goal, implementation Goal, building Goal; delivery kind (hyp stores none)

**Constraints block**:
The end of a Goal's body that states what a run may change, may only read and never does. Its headings are `May change:`, `Read only:` and `Never:`. `hyp prompt` prints the `/goal` prompt from it, and refuses a Goal without a valid block ([spec §7.5](replay-spec.md#75-goal)).
_Avoid_: scope section, run rules

**DesignQuestion**:
A node, `Q-`, under the Goal or a Hypothesis, that asks one design question. Its children are Hypotheses, and any Experiment with no targets that measures for it. Its state is open, reopened, answered, deferred or withdrawn ([spec §7.4](replay-spec.md#74-designquestion-state)). Withdrawn here is a hold that `dq.resume` lifts, not the close of an Experiment.
_Avoid_: issue, decision; withdrawn for an Experiment (see Experiment status)

**Top-level question**:
A DesignQuestion whose parent is the Goal. The Goal is complete when every live top-level question is answered ([spec §7.5](replay-spec.md#75-goal)). Each line of the Goal is owed on each live top-level question where it is uncovered ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
_Avoid_: root question, main question

**Sub-question**:
A DesignQuestion whose parent is a Hypothesis, asked because it matters only if that Hypothesis holds. It is an obligation: while it is not withdrawn and not answered, the question that selected its parent cannot be answered ([spec §7.4](replay-spec.md#74-designquestion-state), ADR 0014). It changes no status of its parent.
_Avoid_: premise (a premise is a `depends_on` edge; see Ledger item)

**Hypothesis**:
A node, `H-`, under a DesignQuestion, that states one possible answer as a claim an Experiment can test. It can name its subject, the bytes that embody the claim, with `hyp subject` ([spec §4](replay-spec.md#4-event-types-closed-set), subject). Its status comes from the verdicts of non-retracted Results, from a `supersedes` edge or from an abandon ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)).
_Avoid_: option, idea, solution; stand-in for the Hypothesis (a stand-in is a file of its subject written for a run)

**Live**:
A Hypothesis that is not superseded and not abandoned. A contradicted Hypothesis is live, but it is not a rival. Staleness and the candidates read live Hypotheses ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)). A DesignQuestion is live when it is not withdrawn, and an edge is live until `edge.remove`. In code: `replay.LiveHypStatus`.
_Avoid_: active, open (open is a question's state)

**Rival**:
A live Hypothesis that is not contradicted: proposed, inconclusive or supported. A rival can still win. Two rivals are **separated** when a non-retracted Result gives them different verdicts, and `hyp show Q` lists the pairs that are not ([spec §5.1](replay-spec.md#51-write-time-refusals)). The one-rival refusal and the one-rival and D1 signals count rivals. In code: `replay.RivalStatus`.
_Avoid_: alternative, competitor; rival for any other Hypothesis of the question (say sibling)

**Candidate**:
A rival that is supported. Neither stale nor an open sub-question removes it ([spec §7.4](replay-spec.md#74-designquestion-state)). One candidate makes the runner's `select` owed, and two or more make a trade-off, where the owner's `choose` is owed. Replay stores the candidates at each select, so a new one reopens the question (`new_candidate`). In code: `replay.Question.Candidates`.
_Avoid_: trade-off candidate, suggested (what `hyp select Q` prints for an owed select), rival (a rival need not be supported); candidate for any Hypothesis of a Goal kind (say Hypothesis)

**Sibling**:
A Hypothesis of the same DesignQuestion as another one, with any status. An Experiment's targets are its parent and siblings of its parent ([spec §5](replay-spec.md#5-refusals) rule 5).
_Avoid_: rival (a rival is live and not contradicted)

**Trade-off**:
A state of a DesignQuestion that has two or more candidates. It is derived, never recorded ([spec §7.4](replay-spec.md#74-designquestion-state)). It ends when Results leave one candidate, or when the owner chooses one, which is recorded as the select (ADR 0015). A trade-off asks which candidate; a reading asks what the owner's words mean.
_Avoid_: `--tradeoff` (removed: a flag cannot record a derived state), tie

**Decision**:
The record of a DesignQuestion's selects: the current select, the select history with the candidates at each select, and the conflicts ([spec §7.4](replay-spec.md#74-designquestion-state)). In code: `replay.Decision`.
_Avoid_: answer (`answered` is a status)

**new_candidate**:
A reason of an open or reopened DesignQuestion: a candidate now was not a candidate when the current select applied. An example is a rival whose deciding `contradicts` was retracted. The owner's choice is owed again ([spec §7.4](replay-spec.md#74-designquestion-state)).
_Avoid_: lapse (a select does not lapse)

**Turn**:
A derived value that says where a DesignQuestion's decision stands: closed, waits, no candidate or owed. One rule decides it, first match wins, and nothing stores it ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes), Turn). The ledger, `hyp select` and `hyp status` read only it. In code: `replay.Turn`, from `Projection.Turn`.
_Avoid_: hold (a DesignQuestion's defer or withdraw), pending (retired, ADR 0017), stands (a select that stands is an answered question)

**Experiment**:
A node, `E-`, that pre-registers one bounded run and what its Result can carry. Before the run it fixes its recipe, rubric, targets, aims and the records it reads. None of them changes after the add ([spec §4](replay-spec.md#4-event-types-closed-set), pre-registration; ADR 0018). It sits under a Hypothesis when it has targets, and under a DesignQuestion when it only measures.
_Avoid_: test, trial (a trial is one scored repeat inside it), run (that is a status)

**Experiment status**:
The derived state of an Experiment: planned, run, superseded or withdrawn. Superseded and withdrawn are closed ([spec §7.6](replay-spec.md#76-experiments-and-results), ADR 0008). In code: `replay.Exp`, in `Projection.Exps`.
_Avoid_: abandoned for an Experiment (abandon is about a Hypothesis); withdrawn for a DesignQuestion's hold

**Target**:
A Hypothesis that each Result of an Experiment must give a verdict on: the Experiment's parent, or a sibling of it named with `--target`. The targets are fixed at the add ([spec §4](replay-spec.md#4-event-types-closed-set); [spec §5](replay-spec.md#5-refusals) rules 5 and 6). An Experiment under a DesignQuestion has no targets.
_Avoid_: subject (the thing under test is the Subject); judges (retired: say targets)

**Trial**:
A scored repeat of an attempt whose outcome can vary, such as one fresh agent on one task or one session replay from a cut. A rubric counts trials ("T correct of 6"). hyp stores no trial: the Experiment's body says how many run and how each is scored.
_Avoid_: run when the count is meant (say "3 of 5 trials"); trial for a fixed case a script checks, or for an instrument check

**Result**:
A node, `R-`, under an Experiment, that records what a run observed: a verdict per target, an observation, evidence pointers, limits and readings ([spec §4](replay-spec.md#4-event-types-closed-set)). A retracted Result no longer counts. In code: `replay.Res`, in `Projection.Results`.
_Avoid_: finding, outcome, run

**Verdict**:
A Result's label for one target: supports, contradicts or inconclusive. Whoever runs `hyp record` types it. Statuses derive from these labels alone, and hyp never reads prose ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)).
_Avoid_: judge, judged (retired: a Result gives a verdict on a target); verdict for the rubric that led to the label

**Evidence pointer**:
One entry of a Result's `evidence`: a URL, a path under `.tree/evidence/`, or `recipes/NAME@N/FILE`. hyp pins a pointer of the last kind by its bundle hash ([spec §7.6](replay-spec.md#76-experiments-and-results)). In code: `event.EvidenceRef` reads the ref of the last kind.
_Avoid_: attachment, link (alone)

**Pin**:
A bundle ref and the hash that an event recorded for it at write time ([spec §7.6](replay-spec.md#76-experiments-and-results)). An Experiment pins its recipe and its rubric, and a Result pins each evidence pointer into a bundle. In code: `replay.Pin`, in `replay.Res.Pins`.
_Avoid_: lock, version (alone)

**Questionable**:
A Result that has a pin that no longer matches its bundle's current hash. It is a flag and changes no status ([spec §7.6](replay-spec.md#76-experiments-and-results), `questionable_evidence`). In code: `replay.Res.Questionable`.
_Avoid_: stale (that is the dependency rule), invalid

**Stale**:
A live Hypothesis or DesignQuestion whose basis may no longer hold. A `depends_on` target is not supported, is stale, or changed since the basis event ([spec §7.3](replay-spec.md#73-stale-snapshot-rule-the-minimal-form-of-truth-maintenance)). hyp shows it as a cause chain, such as `H-5 stale ← H-3 contradicted ← R-9`. In code: `replay.StaleCause`.
_Avoid_: outdated, questionable (that is about bundle hashes)

**Keyed item**:
An item that is one entry inside its parent's event, addressed as `<parent>#<n>` or `<parent>#r<n>`. It is not a node: it has no add, no title and no status of its own. Its text is fixed, and only other events move its state ([spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings)). In code: `internal/replay/keyed.go`.
_Avoid_: sub-node, item node

**Limit**:
A keyed item that names a case the Goal covers and that no check has run on a Hypothesis. A Result records it as `<result-id>#<n>`, and a later Result closes it when it runs the case (ADR 0016). The runner refers a limit that no test within the Goal can close, and the owner accepts or returns it ([spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings)). An **open** limit is one that is neither closed nor accepted ([spec §8](replay-spec.md#8-projection-what-replay-returns), Limits). In code: `replay.Limit`.
_Avoid_: caveat, TODO, known issue, scope, assumption, provenance, criterion

**Line**:
A limit that the owner records on the Goal: a case that every candidate must meet, keyed `G-<goal-id>#<n>`. It comes with the Goal's add and is never edited ([spec §4](replay-spec.md#4-event-types-closed-set), lines). It is neither open nor closed: on each top-level question with a candidate it is covered or uncovered ([spec §8](replay-spec.md#8-projection-what-replay-returns), Lines). In code: a `replay.Limit` whose `Question` is empty, and its `Cites`.
_Avoid_: met line, acceptance criterion, requirement, rule change (a change is a reading, an accept, a return or a new Goal)

**Reading**:
A keyed item, `<result-id>#r<n>`, that records how a run read owner words that a verdict rests on. The words are a line of the Goal, or an owner's answer to an earlier reading. It stays open until an owner answer closes it, and a closed one is owed until a Result applies it ([spec §4](replay-spec.md#4-event-types-closed-set), readings). A reading asks what the owner's words mean; a trade-off asks which candidate to choose. An open reading **holds the decision, not the work**: its question owes no select, choice or stale item, and refuses a select or a reaffirm, whatever the answer can move. It also holds the question's next Hypothesis while its answer can move a Hypothesis there, or while no open limit on the question is unreferred, so the runner never builds on a guess ([spec §4](replay-spec.md#4-event-types-closed-set), readings).
_Avoid_: reopen, re-answer, amend (for a change of an answer)

**Reach**:
The live Hypotheses under a reading's question that the owner's answer can move, read from the reading's own Result: one it judged that is supported now, one it judged inconclusive that is still inconclusive, and a contradicted one whose only contradiction is that Result. A reading with a reach holds the question's next Hypothesis; one with an empty reach still holds the decision. In code: `replay.Reading.Reach`.
_Avoid_: scope (a Goal's), impact

**Reaffirm**:
The runner's act that renews a stale node's basis: `hyp reaffirm` of a Hypothesis or a DesignQuestion takes a new snapshot of its `depends_on` targets ([spec §7.3](replay-spec.md#73-stale-snapshot-rule-the-minimal-form-of-truth-maintenance)). It settles a `stale` item. It is refused while an open reading on the question holds that question's decision ([spec §5.1](replay-spec.md#51-write-time-refusals)), because the answer may change what the select rests on.
_Avoid_: confirm (retired), reselect

**Aim**:
A limit or an answered reading that an Experiment is planned to settle, named with `hyp add exp --aims`. An aim is intent and settles nothing: only a Result's `--closes` or `--applies` settles the item ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
_Avoid_: target (that names a Hypothesis an Experiment judges), plan

**Reassess**:
The step that applies an owner's answer to the tree. An Experiment aims at the answered reading, and its Result applies the answer and gives new verdicts. Earlier Results stay as they are. The steps are in [hyp-goal-run, "Reassess"](../skills/hyp-goal-run/SKILL.md#reassess).
_Avoid_: re-judge (retired), reapply, update the tree, recalculate (replay derives; a reassess adds Results), re-confirm

**Coverage**:
The set of Hypotheses that a Result ran on. These are the Hypotheses it judged, and those whose whole subject one of its run subjects holds. hyp derives it from the log, and the runner never types it ([spec §4](replay-spec.md#4-event-types-closed-set), subject; [spec §8](replay-spec.md#8-projection-what-replay-returns), Lines). Coverage says the bytes ran, not that the claim held. In code: `replay.Projection.Covers`.
_Avoid_: tested (that is a verdict), ran as (a per-Result claim hyp does not store)

**Out-of-Goal finding**:
A finding that is a case of none of the Goal's lines and asks no question about what a line means. It is not a limit: it gets no tree event and never holds the select. A run sorts it as [hyp-goal-run, "Out-of-Goal findings"](../skills/hyp-goal-run/SKILL.md#out-of-goal-findings) says.
_Avoid_: limit, refer (for a case no line covers), side note

**Ledger item**:
An item of open work that the ledger owes, to the runner or to the owner. The ledger is every such item of a Goal: it is derived, never stored, and owing an item changes no status ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)). `hyp select` reads it to admit a select. In code: `replay.Owed`, from `replay.Projection.Owed`, with its kind in `replay.OwedKind`.
_Avoid_: task, ticket (an item is derived, never written)

| Item | Who owes it | Owed while | _Avoid_ |
|---|---|---|---|
| `rivals` | runner | an open question has no rival, and no open reading on it holds its next Hypothesis | |
| `untested` | runner | a proposed Hypothesis has no planned or run Experiment as parent or target | undecided (an inconclusive Hypothesis) |
| `undecided` | runner | a live Hypothesis is inconclusive, no planned Experiment has it as parent or target, and no open reading's answer can move it because the reading's Result judged it inconclusive | untested (a proposed Hypothesis with no Experiment) |
| `planned` | runner | a planned Experiment has no Result | |
| `limit` | runner | an open limit that is not referred sits under a question with a candidate, or a line is uncovered on a live top-level question | |
| `referred` | owner | an open limit's latest hand is a refer; `hyp status` prints `limit referred, accept or return` | |
| `reading` | runner | no Result applies the answer in force of a closed reading; `hyp status` prints `apply answer` | |
| `premise` | runner | a candidate's `depends_on` target is under another question, whose select names another Hypothesis | assumption (prose), sub-question (an obligation, not a premise) |
| `stale` | runner | a question reopened `selected_stale`, owes no other runner item, does not wait on a premise, and has no open reading | |
| `select` | runner | the question's turn is owed, and it has one candidate | |
| `choose` | owner | the question's turn is owed, and it has two or more candidates | pick |

A **premise** is a live `depends_on` target of a candidate under another question. While that question is open or reopened, the candidate's question waits on it; when it selects another Hypothesis, the runner owes the `premise` item ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).

**Status section**:
A part of the `hyp status` output, under one heading. `hyp status G-…` prints the sections of one Goal, in the order [spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes) gives.

**Frontier**:
A status section that shows the open work: the tree, pruned to its open and reopened questions. One walk prints each such question beneath its parent, with its candidates, the Hypotheses they assume, and what its turn waits on ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)). `hyp render --markdown` lists every question.
_Avoid_: hidden, for a question that does not print (say out of play)

**Owed**:
A status section that lists the ledger items the runner owes on one Goal. When the runner owes nothing, it ends with `runner owes nothing: stop and report to the owner`, the stop that a run reads ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)). In code: `replay.Projection.Owed`.
_Avoid_: todo list, backlog, next steps (those are prose; the ledger is derived)

**Awaiting owner**:
A status section that lists the ledger items the owner owes, and every open reading. A `choose` item prints only for a question that the Frontier prints. In a store with many Goals, `hyp status` prints one Awaiting owner list, grouped by Goal ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).

**Open limits**:
A status section that lists every open limit of a Goal's Results, whatever the state of its question. It is a list, not open work: Owed and Awaiting owner name the open work ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).

**Edge**:
A typed link from one node to another node, or to something outside the tree. `edge.add` adds it, and it is live until `edge.remove` ([spec §4](replay-spec.md#4-event-types-closed-set)). In code: `replay.Edge`.

**depends_on**:
An edge from a Hypothesis or a DesignQuestion to a Hypothesis whose support it relies on. Stale follows it ([spec §7.3](replay-spec.md#73-stale-snapshot-rule-the-minimal-form-of-truth-maintenance)), and it cannot close a cycle ([spec §7.2](replay-spec.md#72-combined-status-graph-for-cycle-refusal)).
_Avoid_: requires, blocks

**supersedes**:
An edge that replaces its target: a Hypothesis with a sibling, or an Experiment with another Experiment of the same question. The target reads superseded ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins), [spec §7.6](replay-spec.md#76-experiments-and-results), ADR 0008).
_Avoid_: replaces, obsoletes; superseded for an Experiment closed with no replacement (that is withdrawn)

**external_ref**:
An edge from a node to something outside the tree, such as `gh:976` or a URL ([spec §4](replay-spec.md#4-event-types-closed-set)). It changes no status.
_Avoid_: link (alone), reference

**Admission**:
The write-time rules that the CLI checks before it appends an Experiment, a Result or a select ([spec §5.1](replay-spec.md#51-write-time-refusals)). Each finding is a refusal (exit 1), bad arguments (exit 2) or a warning. Replay does not repeat them, so a merged log can hold an event they would refuse. In code: `replay.State.AdmitExp`, `replay.State.AdmitResult`, `replay.Projection.AdmitSelect` and `replay.Finding`.
_Avoid_: validation (that is `replay.Validate`, which replay also runs on every event)

**Refusal**:
hyp's no to a write: exit 1, and nothing is appended. It comes from `replay.Validate`, which replay also runs, or from Admission, which runs only at write time ([spec §5](replay-spec.md#5-refusals)).
_Avoid_: rejection (that is what replay reports), error

**Cite**:
A token in the text of a write command that names a node. It is a display id such as `R-BV4CWV`, or a placeholder such as `R-<new>`. hyp refuses a write whose cite is no node of the store, with exit 2 ([spec §5.1](replay-spec.md#51-write-time-refusals)). A line's cite is a different thing: a settled key that the line names with `--cite` ([spec §4](replay-spec.md#4-event-types-closed-set), lines).
_Avoid_: reference for this (a reference is what a flag such as `--parent` resolves)

**Round**:
One Experiment as `hyp signals` lists it: its statuses at add, its design marks, and its non-retracted Results with their status moves. In code: `signals.Round`.
_Avoid_: iteration, cycle, step

**Signal**:
A design mark that `hyp signals` derives from the events alone. Signals change no status, and none tracks progress: the ledger does that (ADR 0017). In code: `signals.Signal`, from `signals.Compute`.
_Avoid_: alert, error

**one-rival**:
A signal on a round whose parent had sibling rivals at the add, none of which its targets name ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add). `hyp add exp` refuses such an Experiment unless `--target none` is given ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: no-rival, single-rival

**D1**:
A signal on a round whose parent or a target was not a rival at the add. `hyp add exp` warns `D1:` for the same Hypotheses, because both read one rule ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: dead rival, invalid target

### Recipes

**Bundle**:
A versioned directory that holds a description and the scripts it runs. Its ref and hash identify it, not the directory it sits in (ADR 0005). Its two kinds, Recipe and Rubric, share one format, are gated by their known answers and are handled by `hyp bundle`. An Experiment pins at most one of each.
_Avoid_: recipe folder, package; recipe for a bundle of either kind

**Frozen**:
A bundle version whose current hash a tree pins. Its bytes never change; a change is a new version ([sandbox spec §1.2](sandbox-spec.md#12-hyp-runs-every-command)).
_Avoid_: locked, released

**Draft**:
A bundle version that no tree pins yet. It can be edited in place.
_Avoid_: unpinned, work in progress

**Recipe**:
A bundle in `recipes/` that says how to run one kind of subject and what to measure: its steps and its checks. It names no tree id and never prints a verdict, so one recipe version serves any Experiment and any tree.
_Avoid_: harness script, driver, wrapper; recipe for the rubric (a rubric is a separate bundle)

**Harness recipe**:
The recipe of a stack that an Experiment pins, because it defines what is measured.
_Avoid_: main recipe; harness (alone, that is the agent runtime)

**Environment recipe**:
A recipe of a stack that only sets up something the subject needs, and that no Experiment pins. A sandbox trace does not tell it apart: it is the stack entry the Experiment does not pin.
_Avoid_: dependency recipe, helper recipe

**Stack**:
A set of whole recipes that one sandbox sets up together, in the order given ([RECIPE-FORMAT.md, "Stacking recipes"](../skills/hyp-verification/RECIPE-FORMAT.md#stacking-recipes)).
_Avoid_: recipe chain, composition

**Rubric**:
A bundle in `rubrics/` that turns an Experiment's measures into verdicts, or the same rule as prose in the Experiment's body (`--rubric reading`). A rubric bundle has no step section and one Standalone command, which prints one `verdict: LABEL=WORD` line per label ([spec §7.6](replay-spec.md#76-experiments-and-results)). It names no tree id: the runner gives each label's verdict to its target.
_Avoid_: verdict rule, verdict check (both retired), rule (a numbered refusal rule of the spec), verdict script, criterion (a line is a case), conclusion method

**Instrument**:
A bundle with an Evidence or Standalone command. Its known answers must exist and pass before each run (ADR 0002, [spec §7.6](replay-spec.md#76-experiments-and-results)). Each rubric bundle is one, because its one command is Standalone. hyp cannot vouch for a script outside a bundle (see Uncontained). In code: `recipe.Recipe.IsInstrument`.
_Avoid_: checker

**Recipe ref**:
A name that identifies one bundle version: its name and version, `name@N`, the form an Experiment cites. hyp opens every copy of a bundle by the ref it holds (ADR 0005). In code: `event.Ref`, parsed by `event.ParseRef`; its JSON is the string `name@N`.
_Avoid_: recipe id, recipe tag

**Recipe library**:
The set of recipe versions that a repository keeps in `recipes/`, which Experiments cite instead of copying. Rubrics have their own library, `rubrics/`, and a name is used in one library only ([spec §1](replay-spec.md#1-names-and-locations)).
_Avoid_: recipe registry, recipe cache

**Command**:
A `Run:` line in a recipe section: one script inside the bundle plus fixed words, which hyp runs. The operator runs none of a bundle's scripts directly ([sandbox spec §1.2](sandbox-spec.md#12-hyp-runs-every-command)).
_Avoid_: hook, task, action

**Check**:
A command under Evidence or Standalone. hyp runs it the same way in the known-answer gate and in a run (ADR 0003, [sandbox spec §1.3](sandbox-spec.md#13-how-a-command-runs)). A recipe's check prints measures and never a verdict; a rubric's one check prints `verdict:` lines ([spec §7.6](replay-spec.md#76-experiments-and-results)).
_Avoid_: check for one measure ("adds two checks" when one command gained two measures); checker (see Instrument)

**Measure**:
A named line that a check prints about one subject: a pass or a fail, such as `S1 fail`, or a value, such as `stored rows 4`. A measure is not a verdict: only the Experiment's rubric turns measures into verdicts. hyp gives a measure no identity, no known answer and no pin of its own.
_Avoid_: verdict (a measure judges no Hypothesis); measure for the script that turns measures into verdicts (that is the rubric); measurement script

**Standalone**:
A recipe section that is not a step: its commands run when no sandbox exists ([sandbox spec §1.1](sandbox-spec.md#11-steps-and-sections)). An example is a command that captures an Experiment's inputs or classifies kept files.
_Avoid_: tools, prepare, utilities

**Known answer**:
A fixed input of an instrument, with its expected output, kept under `known-answers/NAME/`. hyp runs it before every run of the instrument, exactly as the run runs the check (ADR 0002, ADR 0003). It compares the exit code, stdout and the files under `$HYP_OUT`. The recipe format calls each one a known-answer case ([RECIPE-FORMAT.md, "Known answers"](../skills/hyp-verification/RECIPE-FORMAT.md#known-answers)).
_Avoid_: fixture (alone), golden test

**Known-answer run**:
A run of one known answer by the gate, as a record lists it: its name and its outcome, `ok` or why it failed. The sandbox trace and the command trace list one per known answer under `known_answers` ([sandbox spec §5.1](sandbox-spec.md#51-records)). In code: `cmdrecord.Case`.
_Avoid_: case (a case is a scenario that a line or a limit covers), test case (alone), result (a Result is a tree node)

**External line**:
The `External:` line under an Evidence or Standalone command, which names what the command reads or writes outside its bundle, or `none`. `hyp bundle check` warns when it is missing or reads `TODO` ([sandbox spec §1.2](sandbox-spec.md#12-hyp-runs-every-command), ADR 0008).
_Avoid_: dependencies line, inputs line

**Kept input**:
A file that an operator argument names, which hyp copies beside the command trace and hashes before the first command runs ([sandbox spec §5.1](sandbox-spec.md#51-records)). It binds the subject when the command runs, not when the Experiment is added (ADR 0008). In code: `cmdrecord.Input` and `cmdrecord.Dir`.
_Avoid_: input binding, arg hashes, inputs field (ADR 0008 rejected one on the Experiment)

**Trace**:
hyp's file of one run of a bundle's commands: what ran, on which bytes, and how each command ended. Only hyp writes it. A Result cites it as evidence, and `hyp record` copies its run subjects into the Result ([sandbox spec §5.1](sandbox-spec.md#51-records)). Its two kinds are the Sandbox trace and the Command trace.
_Avoid_: record (kept for the history on `refs/hyp/record`, #181), receipt, run file, manifest (`manifest.json` is the sandbox's file list)

**Command trace**:
A trace of one run of Evidence or Standalone commands outside a sandbox: `command.json`, in the directory `--record DIR` names ([sandbox spec §5.1](sandbox-spec.md#51-records)). It keeps each file argument, and containment reads it for a clean run. In code: `cmdrecord.Record`, read by `cmdrecord.Read`; `Record.CleanRun` is the clean run.
_Avoid_: command record (retired), command log (`command.log` is a file beside it)

**Self-contained**:
A recipe that relies on nothing but its own files, hyp and the subject, wherever its commands run. A fixed input that is not the subject is a file of the bundle ([sandbox spec §1.2](sandbox-spec.md#12-hyp-runs-every-command), ADR 0008).
_Avoid_: sandboxed (self-contained does not mean "runs in the sandbox")

**Containment**:
What hyp derives from a Result's files about whether it can vouch for the script that decided the verdict ([spec §7.6](replay-spec.md#76-experiments-and-results)). The mark is read from files each time and changes no status. It is not self-contained, which is about a recipe, and not sandbox isolation, which is about where commands run.
_Avoid_: isolation, contained (for a sandbox)

**Uncontained**:
A Result whose rubric or check hyp cannot vouch for, for one of the three reasons of [spec §7.6](replay-spec.md#76-experiments-and-results).
_Avoid_: unpinned (alone)

### Sandboxes

**Sandbox**:
A disposable copy of the subject at a recorded base commit, with its own scratch space, pinned hyp and record. It is the only place hyp assembles a subject, and code under test is built in its Launch ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). It is not process isolation: commands still run on the host.
_Avoid_: container, VM, environment

**Sandbox trace**:
A trace of one sandbox from create to remove: `sandbox.json`. It names no host path, since the sandbox's own paths derive from its id ([sandbox spec §5.1](sandbox-spec.md#51-records)).
_Avoid_: sandbox record (retired), sandbox manifest, metadata

**Sandbox state**:
The derived state of a sandbox: `building`, `setting_up`, `interrupted`, `ready`, `setup_failed` or `removing` ([sandbox spec §2](sandbox-spec.md#2-the-sandbox-machine)). It is never stored. In code: `sandbox.boxState`, from `deriveState`.
_Avoid_: stored state, status of a sandbox

**Step**:
A stage of running an Experiment: Launch, Doctor, Drive, Evidence or Cleanup. In a recipe, each step is a section of that name ([sandbox spec §1.1](sandbox-spec.md#11-steps-and-sections)). In code: `recipe.Step`, whose constants also name Standalone.
_Avoid_: phase, stage, hook

**Launch**:
A step that sets up the subject and the recipe's resources inside a new sandbox.

**Doctor**:
A step that checks, before any drive, that Launch left the sandbox ready.
_Avoid_: preflight, health check

**Drive**:
A step that exercises the subject while the sandbox exists.
_Avoid_: run, execute

**Evidence**:
A step that turns what happened into findings. When it is scripted, it reads only the kept files after the sandbox is gone, so it can run again.
_Avoid_: check (as a step name), verify

**Cleanup**:
A step that stops what a recipe's Launch started, before hyp deletes the sandbox.
_Avoid_: teardown

**Kept files**:
What remove copies out of a sandbox before it deletes it, which scripted Evidence reads. In code: `sandbox.KeptFiles`, every name that remove writes into OUT_DIR, `patches/` among them.
_Avoid_: artifacts, output (alone)

**Store cut**:
A copy of a store as it stood at one event, given to a sandbox ([spec §6](replay-spec.md#6-merge-order)).
_Avoid_: snapshot (a capture of a live session is a different thing)

**Patch**:
A throwaway change over the subject's base, given to `hyp sandbox create` as `--patch ARM=FILE` for one Arm. An Arm has at most one, and a change that lands is a commit used as the base instead ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)). A patch is the subject, not a file of the recipe bundle.
_Avoid_: variant (alone), diff, fork

### People and binaries

**Owner**:
The person whom hyp asks on real choices, and the role they act in. The owner adds the Goal with its lines, and answers what Awaiting owner lists: a reading, a referred limit or a choice between candidates. The runner records a choice as the select, with the owner's words in `--why` (ADR 0015). The rules of the role are in [principles §7](principles.md#7-who-does-what), and the moves in [spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings). In code: `event.Answer`; `event.RoleOwner` on old owner checkpoints.
_Avoid_: user, stakeholder, approver, pick, decide (alone)

**Runner**:
The agent that runs the loop with hyp's skills and settles what the ledger owes it. In a replay it runs inside the sandbox and writes that tree's verdicts there. It is not the operator of the Experiment that started it.
_Avoid_: drive agent, subject

**Operator**:
The person or agent that runs an Experiment from outside the sandbox and types its verdicts into `hyp record`.
_Avoid_: runner, driver, user

**Actor**:
The name that a write records as its author, from `--actor` or `HYP_ACTOR` ([spec §1](replay-spec.md#1-names-and-locations)). Each actor appends to its own events file, so logs from several worktrees merge ([spec §6](replay-spec.md#6-merge-order)). An actor is not a role: a role is what a person does, and separation of duties compares only actor names.
_Avoid_: user, author (alone)

**Harness**:
The agent runtime that loads a skill and hosts the runner, such as Claude Code or Codex.
_Avoid_: harness for a recipe or a binary (say harness recipe, operator hyp)

**Sandbox hyp**:
The hyp binary that a sandbox pins and a runner inside it calls: a copy of the operator hyp. In code: `sandbox.SandboxHyp`.
_Avoid_: subject hyp (true only when hyp itself is under test), pinned hyp (alone)

**Operator hyp**:
The hyp binary that the operator runs, which sets up the sandbox and runs the recipe's commands.
_Avoid_: harness hyp, harness (for a binary)

**Subject**:
The thing under test, in one shape: a base commit, at most one patch over it, and input files (ADR 0020). A Hypothesis's subject is what embodies its claim. A run's subject is what its record names, and a Result keeps one per cited run ([spec §4](replay-spec.md#4-event-types-closed-set), subject; [sandbox spec §5.3](sandbox-spec.md#53-the-subject)). A run's subject proves which bytes ran, but hyp cannot check which claim they stand for.
_Avoid_: target, system under test; fingerprints (only an old Result's hashed files)

**Arm**:
A variant of the subject that an Experiment compares: the base, or the base plus one Patch. A lowercase word names it, the ARM of `--patch ARM=FILE` ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)).
_Avoid_: variant (alone), branch

### Scopes of this repository

**Core**:
Every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A concept stays core when self-review is what tests it.
_Avoid_: product, the tool, hyp proper

**Self-review**:
The recipes this repository keeps to test hyp, whatever repository the runner works on, with the inputs only they use. Investigating hyp is an activity, not a scope. A tree about hyp takes the scope of its question, and a tree about another project is a self-review input.
_Avoid_: dogfood (alone), meta, hyp-self (that is one tree)

### Retired words

A retired word stays only in old logs, ADRs and history. Do not use it for a current concept.

| Retired word | Retired by | Say instead |
|---|---|---|
| Confirm | ADR 0019 | the select, which records the owner's choice; for a scope change, a line that the runner refers and the owner accepts. In plain speech the owner chooses or accepts. Not approve or review. |
| Dossier | ADR 0019 | nothing: a select stands until a new candidate reopens its question |
| Lapse | ADR 0019 | `new_candidate`: a select does not lapse |
| Verifier | ADR 0015 | the select is the gate. Any actor can review, with a Result, a retract or a limit; there is no reviewer or checker role. In code: `event.RoleVerifier`, read only from old logs. |
| verifier-loop | ADR 0015 | nothing; not spin, which is also retired |
| unjudged | ADR 0018 | a refusal: `hyp record` refuses a Result that leaves a target without a verdict (see Target). Not untested, which is a DesignQuestion reason. |
| wasted, pending, no-movement, spin | ADR 0017 | the ledger, the one account of progress; a stall is a ledger item that does not settle |
| verdict check, verdict rule | ADR 0018 | Rubric for the rule; Measure for what a recipe's check prints |
| command record, sandbox record | owner decision of 2026-11-14 | Command trace, Sandbox trace: "record" is kept for the history on `refs/hyp/record` (#181) |
