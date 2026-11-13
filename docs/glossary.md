# Glossary

The words hyp uses for its tree, its recipes and its sandboxes. The specs ([replay-spec.md](replay-spec.md), [sandbox-spec.md](sandbox-spec.md)) and the recipe format define the rules; this file only fixes which word means what, links the rule's section instead of restating it, and names the Go type that owns each concept where one exists. It also names the two scopes of this repository, core and self-review, and the test that puts an item in one; [core-and-self-review.md](core-and-self-review.md) maps the repository onto them.

## Language

### Trees

**Tree**:
One investigation: a Goal and the DesignQuestions, Hypotheses, Experiments and Results under it, kept as events in one store. The binary is `hyp`; "tree" stays the concept name ([spec §1](replay-spec.md#1-names-and-locations)).
_Avoid_: graph, log (the log is its events), project

**Store**:
The `.tree/` directory of a repository: one events file per actor and the evidence files Results point at ([spec §1](replay-spec.md#1-names-and-locations)). hyp finds it by walking up from the current directory to the repository root, the first directory that holds `.git`, and never above it; `hyp init` is refused where it finds one. In code: package `store`.
_Avoid_: database, index (hyp keeps none, ADR 0006)

**Event**:
One line of `.tree/events/<actor>.jsonl`: an id, a time, an actor, a type from a closed set and its payload ([spec §3](replay-spec.md#3-event-envelope), [spec §4](replay-spec.md#4-event-types-closed-set)). Events are only appended, and every status is derived from them. In code: `event.Event`.
_Avoid_: entry, record (a record is what hyp writes for a sandbox or a command)

**Actor**:
The name a write records as its author, from `--actor` or `HYP_ACTOR` ([spec §1](replay-spec.md#1-names-and-locations)). Each actor appends to its own events file, so logs from several worktrees merge ([spec §6](replay-spec.md#6-merge-order)). An actor is not a role: owner is the role an owner checkpoint names (the verifier role is retired, ADR 0015).
_Avoid_: user, author (alone)

**Replay**:
Deriving every status of a tree from its ordered events.
_Avoid_: replay for re-running an agent session (that is an Experiment run with a session-replay recipe)

**Projection**:
Everything replay derives from a store's events: each node, a derived record per node kind (Hypothesis, Question, Experiment and Result) with its own status word, the Goal's state, rejections, limits and selections ([spec §8](replay-spec.md#8-projection-what-replay-returns)). Nothing stores it; every read replays (ADR 0006). In code: `replay.Projection`.
_Avoid_: state (alone), index, cache

**Goal**:
The root node of a tree, `G-`: what the investigation is for. A store holds many Goals, and a tree is the view of one Goal and every node whose parents lead to it (ADR 0020); each Goal has its own lines, OWED and coverage. A Goal is complete when every live DesignQuestion under it is answered ([spec §7.5](replay-spec.md#75-goal)). Its lines are the limits the owner records on it (see Line); the owner's "met" is a complete Goal with no line under OWED, read from `hyp status`, and hyp stores no word for it. No one changes a Goal once it is written: its lines come with its add, and a new purpose is a new Goal, which may sit in the same store and cite the keys it carries over (see Line). By what its candidate is, a Goal is a discovery Goal or a build Goal; hyp stores no kind, and the same events, rules and ledger serve both (ADR 0020).
_Avoid_: objective, epic, done, met (as a status word)

**Constraints block**:
The end of a Goal's body that states what a run may change, may only read and never does: the lines from the first one that starts with `May change:`, `Read only:` or `Never:` to the end of the body, each non-empty line starting with one of the three headings; a line of spaces, a tab or a lone carriage return is not empty ([spec §7.5](replay-spec.md#75-goal)). `hyp prompt` prints the `/goal` prompt from it, with the store, the actor, the Goal and the stop clauses, and refuses a Goal without a valid block. The prompt holds no owner answer; OWED lists what a run applies.
_Avoid_: scope section, run rules

**Discovery Goal**:
A Goal whose candidate is a claim: what to build, or why something behaves as it does. Rival Hypotheses answer each DesignQuestion, and Experiments separate them, often from a pinned recipe whose rubric turns its measures into verdicts. A line names a case every candidate must meet that a check can run on every candidate. A discovery Goal whose questions the owner settles by argument in a session, with the session as evidence, is a design record: no run checks its candidates, so it has no lines, and its cases go to the lines of the build Goal that follows it.
_Avoid_: investigation Goal, research Goal; discovery kind (hyp stores none)

**Build Goal**:
A Goal whose candidate is a commit: building is a Goal like any other. A line asserts one property of the candidate, decided by its own tests or one pinned instrument: `[P]` pins what a valid input produces, and `[N]` pins that a forbidden state is refused. Each commit is a Hypothesis whose subject is the base plus that commit's change, and a fix is a new Hypothesis that supersedes it. Each Experiment's rubric asks for red on base and green on the candidate; the base of the red run's subject is where the build starts, and the Goal records no start commit. A finding of a run's review that is a case of a line is a limit; one that is not is an out-of-Goal finding. A change to a skill is a build Goal too, except a prose-only change, which ships as a plain PR (ADR 0021).
_Avoid_: delivery Goal, implementation Goal, building Goal; delivery kind (hyp stores none)

**DesignQuestion**:
A node, `Q-`, under the Goal or a Hypothesis, that asks one design question; its children are rival Hypotheses, and any Experiment with no targets that informs the question. Its state is open, reopened, answered, deferred or withdrawn ([spec §7.4](replay-spec.md#74-designquestion-state)). Withdrawn here is a hold that `dq.resume` lifts; it is not an Experiment's close.
_Avoid_: issue, decision; withdrawn for an Experiment (see Experiment status)

**Sub-question**:
A DesignQuestion whose parent is a Hypothesis, asked because it only matters if that Hypothesis holds. It is an obligation: while it is not withdrawn and not answered, deferred included, the question that selected its parent cannot be answered, and reads `awaiting_subquestions` ([spec §7.4](replay-spec.md#74-designquestion-state)). Any answer discharges it and withdrawing it releases it, whichever way it was going. That is its only effect on the parent: the parent's status does not change, a Result under it judges its own Hypotheses, its answer gives the parent no support, and the Dossier does not list it. It blocks the Goal only through this rule ([spec §7.5](replay-spec.md#75-goal)). A claim that rests on an answer is a premise, and a premise is `depends_on`.

**Hypothesis**:
A node, `H-`, under a DesignQuestion, that states one possible answer as a claim an Experiment can test. It may name its subject, what embodies the claim, tied with `hyp subject`: a base commit (`--base`), at most one patch over it (`--patch`), and files kept by each file's sha256, which over a base are its input files; two Hypotheses may share a subject, since the claim is what differs, and a Hypothesis with no subject, such as a diagnosis, is covered only when a Result judges it. Its status comes from the verdicts of non-retracted Results, or from a supersedes edge or an abandon ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)); its subject changes no status.
_Avoid_: option, idea, solution; stand-in for the Hypothesis (a stand-in is a file of its subject written for a run)

**Frontier**:
The section of `hyp status` that shows the open work: the tree pruned to it, from one walk over every question. A node prints only beneath its parent. A sub-question prints only from an open or reopened question and through a live rival; a deferred sub-question that a printed supported Hypothesis still owes also prints. A printed question with candidates has a `candidates (N)` line naming them, so after a reassess it shows whether one or several are left; with none, the question's reason says so. Beneath it, an `assumes` line names each Hypothesis a candidate rests on through `depends_on`, with that Hypothesis's status, the question it answers when that is another one, and which candidates assume it, so a choice that rests on another question's answer says so. Open questions that do not print are counted in one line, so the Goal's counts stay accounted for; `hyp render --markdown` lists every question.
_Avoid_: hidden, for a question that does not print

**Awaiting owner**:
The section of `hyp status` that names the one step due from the owner on each question the Frontier prints: `choose: H…, H…` when they should choose between the candidates and the runner owes nothing on the question. A question the Frontier does not print asks nothing, and the choice is owed again once it prints. It also lists every open reading, which blocks nothing, and every referred limit on a candidate. `hyp status G-…` lists one Goal's items; in a store with many Goals, `hyp status` prints one Awaiting owner list, grouped by Goal.

**Reading**:
How a run read owner words that a verdict rests on: a line of the Goal, recorded with `hyp record --reading`, or the owner's answer to an earlier reading, recorded with `hyp record --reading-of R-…#rN=TEXT`, which stores that reading's key as the new reading's `of`. A reading of an answer narrows it and never reopens it: the earlier reading stays closed and its answer stays in force. A reading stays open until an owner Answer closes it with their words. Two readings with the same `of` and the same text ask the owner the same question: each makes its own Result's question wait, `hyp status` prints them as one line, and the owner closes them in one answer.
A reading asks what the owner's words mean, and a trade-off asks which candidate to choose. Closing a reading changes no status, because statuses come only from verdicts. An answer that rules out a candidate takes effect when a later Result records `contradicts` on it. A closed reading is owed by the runner until a Result applies it (`hyp record --applies`, from an Experiment that aims at it), and the owner is not asked to choose on its question until then. An owner who changes an answer answers the reading again (`hyp answer --answer R-…#rN=TEXT`): the new words are the answer in force from then on, the earlier answer stays in history, and the reading is owed by the runner again until a later Result applies it. A revision changes no earlier event or verdict. One verb serves both: whether an answer is a first answer or a revision is derived from the reading's state.
_Avoid_: reopen, re-answer, amend (for changing an answer)

**Reassess**:
The step after the owner answers a reading: an Experiment aims at the answered reading, and its Result applies the answer and gives new verdicts on its targets. The answer goes into a check of a new recipe version, or into the rule in the Experiment's body when its verdicts are decided by reading; either way the rule is fixed before the run. Earlier Results stay as they are. Closing a reading changes no status; a reassess does, through its verdicts. Until a Result applies the answer, `hyp status` lists it under Owed as `apply answer`.
_Avoid_: re-judge (retired verb), reapply (an answer is applied once; replay also applies events), update the tree, recalculate (replay already derives every status; a reassess adds Results), re-confirm (the owner confirms a select, not a Hypothesis)

**Open limits**:
The section of `hyp status` that lists every open limit, whatever the state of its question and whether the Frontier prints it. It is a list, not open work: the open work in it is what Owed and Awaiting owner name.

**Rival**:
Another Hypothesis under the same DesignQuestion. Two live rivals are **separated** when a non-retracted Result gives them different verdicts, and `hyp show Q` lists the pairs that are not; exhaustiveness is a property of the cases run on the candidates (see Limit), never of the set of rivals, which the owner may reserve. A live rival is live and not contradicted ([spec §7.1](replay-spec.md#71-hypothesis-status-first-matching-rule-wins)); the one-rival refusal and the one-rival and D1 signals read the live rivals of an Experiment's parent at its add ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add).
_Avoid_: alternative, competitor

**Experiment**:
A node, `E-`, that pre-registers one bounded run and what its Result may carry. It binds a Recipe and a Rubric to the tree: the recipe says how to run and what to measure, the rubric says how the measures become verdicts, and the Experiment says on what, about which Hypotheses and to settle which items. Before the run it fixes its targets, the items it aims to settle, the recipe it pins (a ref and its hash) or `none` when it rules from records already kept, the records it reads beyond its own run (a Result that cites any other record is refused), and its rubric: a pinned rubric bundle, prose in its body when its verdicts are decided by reading, or none when it has no targets ([spec §4](replay-spec.md#4-event-types-closed-set)). None of these changes after the add: a different recipe or rubric is a new Experiment. It sits under a Hypothesis when it has targets, and under the DesignQuestion it informs when it has none. Its derived state is its Experiment status.
_Avoid_: test, trial (a trial is one scored repeat inside it), run (that is a status)

**Experiment status**:
An Experiment is planned, run, superseded or withdrawn (spec §7.6). It is run while it has a non-retracted Result, which wins over a close. Otherwise it is superseded when another Experiment under the same question has a live `supersedes` edge to it (`hyp supersede NEW --old OLD`), withdrawn when `hyp withdraw` closed it with nothing to replace it, and planned otherwise. A superseded or withdrawn Experiment is closed: `hyp status` drops it, `hyp record` refuses a Result under it, and `hyp signals` keeps its round's D1 and one-rival marks (ADR 0008). In code: `replay.Exp`, built once per projection in `Projection.Exps`.
_Avoid_: abandoned for an Experiment (abandoning is a judgement about a Hypothesis's claim); withdrawn for a DesignQuestion's hold (see DesignQuestion)

**Target**:
A Hypothesis an Experiment's Result may give a verdict on, fixed before the run when the Experiment is added: its parent Hypothesis and each sibling given with `--target` ([spec §4](replay-spec.md#4-event-types-closed-set), [spec §5](replay-spec.md#5-refusals) rule 5). An Experiment that gives no verdict has no targets and sits under the DesignQuestion it informs. A Result gives a verdict on each of its Experiment's targets and on nothing else: `hyp record` refuses a Result that leaves a target without one, and `inconclusive` is the verdict for a run that could not decide.
_Avoid_: subject (the thing under test is the Subject; a target is a Hypothesis); judges (retired: say targets)

**Result**:
A node, `R-`, under an Experiment, that records what a run observed: a verdict per target of its Experiment, an observation, evidence pointers and any limits ([spec §4](replay-spec.md#4-event-types-closed-set)). It is recorded or retracted, and a retracted Result no longer counts. In code: `replay.Res`, which holds its status, pins, questionable flag and uncontained reasons, in `Projection.Results`.
_Avoid_: finding, outcome, run

**Verdict**:
A Result's label for one target of its Experiment: supports, contradicts or inconclusive. Whoever runs `hyp record` types it, a person or an agent, and hyp stores only an actor name. hyp derives status from these labels alone (spec §7.1) and never reads prose. An owner confirm's sufficient or insufficient is the owner confirming a select, a separate thing.
_Avoid_: judge, judged (retired: a Result gives a verdict on a target); verdict for the rubric that led to the label

**Rubric**:
How an Experiment's measures become verdicts, fixed before the run. It is the kind of Bundle kept in `rubrics/`, apart from recipes, so two Experiments can pin the same one; it has no step section and its one command reads the measures in the Experiment's records and prints a `verdict: LABEL=WORD` line per label, with a supports case and a contradicts case for each label among its known answers; or, when the verdicts are decided by reading, prose in the Experiment's body, which only its author and a reader of the Result can hold to one reading. A rubric names no tree id: the runner gives each label's verdict to its target in `hyp record`. An Experiment with no targets has no rubric.
_Avoid_: verdict rule, rule (the spec's numbered refusal rules), verdict check, verdict script, criterion (a Goal line is the case the Goal requires), conclusion method

**Check**:
An Evidence or Standalone command of a recipe. hyp runs it the same way in the gate and in a run (ADR 0003), its known-answer cases pin its exit code and its whole stdout, and `hyp bundle check` warns per command and per verdict label (ADR 0004). Each line it prints about one subject, such as `A1 hold`, is a Measure: hyp gives a measure no identity, no case of its own and no pin of its own. A check never prints a verdict.
_Avoid_: check for one measure ("adds two checks" when one command gained two measures); checker (see Instrument)

**Verdict check**, **verdict rule**:
Retired: a verdict comes from the Experiment's Rubric, and a recipe's check prints measures only.

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
A trade-off is between candidates, not about what a Goal line means, which is a Reading. It ends when Results leave one candidate, or when the owner chooses one, recorded as the select (ADR 0015); no confirm follows the choice. The owner's turn to choose is the ledger item `choose`, owed only when the runner owes nothing on the question and no open reading on it can move a Hypothesis there ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
_Avoid_: `--tradeoff` (removed: a flag cannot record a state hyp derives), tie

**Select**:
The `dq.select` event that records one Hypothesis as a DesignQuestion's answer. It is written by whoever runs the tree, never in the owner's role. The select is the gate: `hyp select` admits it only when the ledger owes a `select` of that Hypothesis or the owner's `choose` on the question, and otherwise refuses, naming what is owed ([spec §5.1](replay-spec.md#51-write-time-refusals)). An admitted select answers the question while it stands ([spec §7.4](replay-spec.md#74-designquestion-state)); a scope change is not a select but a Goal line the owner accepts, or a new tree (ADR 0019). On a trade-off the owner chooses first (see Owner). In code: the question's `replay.Decision` and `replay.Projection.AdmitSelect`.
_Avoid_: pick, choose (the owner chooses; the select records it), accept, decide (alone)

**Confirm**:
Retired by ADR 0019. It was the `confirm` event in which a role judged a select `sufficient` or `insufficient`; the last use was the owner's confirm of a scope-change select. `hyp confirm` exits 2. An old owner confirm replays only the limits it accepted ([spec §4](replay-spec.md#4-event-types-closed-set)), and an old verifier confirm applies nothing. The owner still "confirms" in plain speech when they choose or accept.
_Avoid_: approve, review; verdict (alone: a confirm's answer was a confirm verdict)

**Decision**:
A DesignQuestion's current Select, its select history and conflicts, and the candidates when the select applied.
_Avoid_: answer (`answered` is a status)

**Dossier**, **Lapse**:
Retired by ADR 0019. The Dossier was the evidence an owner Confirm judged, and a Confirm lapsed when it changed. With no confirm left, nothing reads it: a select stands until a new candidate reopens its question.

**Verifier**:
Retired by ADR 0015. It was the role that confirmed every select, independently of the actor who made it. Old verifier confirms replay as history and apply nothing, and `hyp confirm` is retired (ADR 0019). Any actor may still review: a defect is recorded as a Result, a retract or a limit. In code: `event.RoleVerifier`, read only from old logs.
_Avoid_: reviewer, checker (as a role)

**Owner**:
The person hyp asks to act on real choices, and the role they act in. The owner adds the Goal with its lines, and then answers what the ledger lists under Awaiting owner: a reading or a referred limit with an Answer, and a choice between candidates (`choose:`). Whoever runs the tree records a choice as the select, with the owner's words in `--why`; the tree holds what was chosen, but not who chose it ([ADR 0015](adr/0015-the-select-is-the-gate.md)). hyp does not authenticate actors, so an owner act records that someone wrote an answer; that the owner stood behind it rests on the procedure. The owner's role and what they never do are stated in [principles.md](principles.md#7-who-does-what), and the moves in [spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings). In code: `event.Answer`; `event.RoleOwner` on old owner checkpoints.
_Avoid_: user, stakeholder, approver, pick, decide (alone)

**Answer**:
The owner's act, its own event type (ADR 0020): `answers` for readings, `accepts` and `returns` for referred limits, each item a key with the owner's words for it, and at least one key. It has no node, no summary, no why and no role, and owner and runner fields never mix in one event. One answer may settle keys of several Goals. An item in `answers` is a first answer of an open reading or a revision of an answered one, derived from the reading's state. An accept means only that this Goal does not run the case. `hyp answer` writes it; an old owner checkpoint's `closes`, `revises`, `accepts` and `returns` are read-only history that replay still applies ([spec §4](replay-spec.md#4-event-types-closed-set)). In code: `event.Answer`.
_Avoid_: owner checkpoint (history), approval, confirm (retired)

**Checkpoint**:
The runner's report on one Goal: a `summary` of where the work stands, `refers` that hand limits to the owner with why, or both, and a required `why`; an empty one is refused. `hyp status` prints its Goal's last one, never an old owner checkpoint, and a run records one at each stop. It holds no owner act: an old owner checkpoint's `role`, `closes`, `revises`, `accepts`, `returns` and `lines` are read-only history that replay applies and no command writes (owner decision of 2026-11-11). In code: `event.Checkpoint`.
_Avoid_: owner checkpoint (history; see Answer), report

**Limit**:
A case the Goal covers that no check has yet run on a Hypothesis. It is a keyed item with two births and one life: a Result records a limit its run left uncovered, keyed `<result-id>#<n>` (ADR 0016), and the owner records a line of the Goal, keyed `G-<goal-id>#<n>` (see Line). Its one job is to name the next case to run: an Experiment aims at it (`--aims`), and a later Result that runs the case on the candidates closes it (`--closes`), so its verdicts separate the rivals. A runner refers a limit that no test within the Goal can close, or whose case is void for every candidate (`hyp checkpoint --refers K=WHY`), which moves it from the runner's Owed items to Awaiting owner. The owner answers the refer with an Answer: accepts it (`hyp answer --accept K=TEXT`), so this Goal does not run the case and a later Goal may carry it by citing the key, or returns it (`--returns K=TEXT`) when the Goal covers its case, which makes it the runner's again; a return also reopens an accepted limit, and an accept needs a refer first. One move table states each move, from which state and by whom ([spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings)). A close records that the case ran, whatever the verdicts, and a limit's category is never stored: it is what its closers' verdicts, an accept, and the latest refer or return say. A limit's hand (the latest refer, return or accept) and its closers are two facts, and an **open** limit is one neither closed nor accepted; open is said of a Result's limit only. An open limit is owed while its question has a candidate, whichever Hypotheses its Result judged, so the select waits on it ([spec §8](replay-spec.md#8-projection-what-replay-returns), Limits). Other notes a run makes have homes of their own: the scope of a run goes in the Experiment body, a procedure note or a commit goes in the observation with the commit in a cited file, and a premise is a `depends_on` edge. In code: `replay.Limit`.
_Avoid_: caveat, TODO, known issue, scope, assumption, provenance, criterion

**Line**:
A limit the owner records on the Goal: a case every candidate must meet, keyed `G-<goal-id>#<n>`. A Goal's lines come with its add (`hyp add goal --line TEXT`), and no later event adds one (ADR 0020); an old Goal's lines, written on its first owner checkpoint, keep their keys. A line may cite 0 to 10 keys settled when it is written (`--cite N=KEY`): a line, an accepted or closed limit, an answered reading, or a selected or contradicted Hypothesis. What a cite means (carried, forbidden, applied or realised) is derived from the key's state at the line's place in the log, and a later change is only a `hyp check` warning. A line names a hash whole: `hyp add goal` refuses a line with a run of 8 to 63 hex digits followed by an ellipsis. A line is never edited: what it means is a reading and the owner's answer; a line the Goal no longer needs is referred by the runner and accepted by the owner; a case under a line is a limit, returned by the owner when the Goal covers it; a new line or a new purpose is a new Goal, which may sit in the same store. A line is neither open nor closed: `hyp status` prints its hand when it has one, `covered on Q-…` or `uncovered on Q-…: H-…` for each top-level question with a candidate (see Coverage), and the Results that ran it. A line is owed as a `limit` item on each live top-level question of its own Goal where it is uncovered, and holds that question's select. A verdict is per Hypothesis, never per line: a Result that closes several lines and contradicts a Hypothesis says the cases ran and the Hypothesis failed, not which line, and its observation says which check failed. `hyp status` prints the lines under LINES, and `hyp show G-…#n` prints one ([spec §8](replay-spec.md#8-projection-what-replay-returns), Lines). In code: a `replay.Limit` whose `Question` is empty, and its `Cites`.
_Avoid_: met line, acceptance criterion, requirement, rule change (a change is a reading, an accept, a return or a new tree)

**Out-of-Goal finding**:
A finding, often from a run's review, that is a case of none of the Goal's lines and asks no question about what a line means. It is not a limit, since a limit is a case the Goal covers, so it gets no tree event, no refer and no accept, and it never holds the select. The runner names it in its stop checkpoint's summary and its last message as a possible new Goal or issue, and the owner decides at global scope whether to add one. A defect in text the run changed and may edit is not one: the run fixes it with a commit.
_Avoid_: limit, refer (for a case no line covers), side note

**Keyed item**:
An item created as one entry inside its parent's event and addressed as `<parent>#<n>` or `<parent>#r<n>`: a limit or a reading inside a Result, or a line of the Goal. Its text is fixed, it takes no edge, and it is gone when its Result is retracted. Only other events move its state, and one move table per kind says which move applies from which state and by whom ([spec §8.2](replay-spec.md#82-keyed-items-limits-lines-and-readings)). It is not a node: it has no add event, no title and no status of its own. In code: `internal/replay/keyed.go`.
_Avoid_: sub-node, item node

**Coverage**:
The Hypotheses a Result ran on: those it judged, and those whose whole subject one of its run subjects holds: the subjects, one per cited run, that `hyp record` copied from its cited command and sandbox records. A Hypothesis that names a base is held only by a run at that base with the same patch, none when it names none, that was given each of its files as input. A patch changes the code and an input file is data given to it, so a run at B plus a patch never covers bare B, a run at bare B never covers B plus a patch, and a run at B given input files covers bare B. It is derived from the log and never typed by the runner. A line is covered on a question when a closer covers each candidate there, and uncovered there otherwise; coverage is derived, never a state of the move table. Coverage says the bytes ran, not that the claim held: a claim that depends on its host is covered only together with that host, so `hyp status` marks a line whose selected Hypotheses no one closer covered together, and `hyp show` prints `ran on H-…` for a covered Hypothesis the closer did not judge. In code: `replay.Projection.Covers`.
_Avoid_: tested (that is a verdict), ran as (a per-Result claim hyp does not store)

**Owed**:
The ledger `hyp status` prints under OWED: each item of open work on a question and who owes it, the runner; the owner's items print under Awaiting owner ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)). OWED is per Goal: `hyp status G-…` prints one Goal's, and in a store with many Goals `hyp status` prints a GOALS table with what each Goal's runner owes. A line of the Goal is owed on each live top-level question that has a candidate no closer covers, as a `limit` item keyed `G-…#n`. A move never crosses Goals: only an item's own Goal refers, aims at, closes or applies it. An answered question owes only its limit, reading, referred and premise items. It is derived, never stored, and owing an item changes no status; `hyp select` reads it to admit a select. When the runner owes nothing, OWED ends with `runner owes nothing: stop and report to the owner`, the stop a run reads. In code: `replay.Projection.Owed`.
_Avoid_: todo list, backlog, next steps (those are prose; the ledger is derived)

**Turn**:
Where a question's decision stands, derived by one rule, first match wins, and never stored ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)): **closed** (answered, deferred or withdrawn; a select that stands is an answered question), **waits** (on runner work, an open reading with a reach, a referred limit, a premise, or a sub-question, in that order), **no candidate**, or **owed** (the runner's select of a lone candidate, or the owner's choice). The ledger owes a select or a choice only on an owed turn; `hyp select` refuses with the first thing the turn waits on; `hyp status` prints a `waits on …` line for each thing other than runner work. In code: `replay.Turn`, from `Projection.Turn`.
_Avoid_: hold (a DesignQuestion's defer or withdraw), pending (retired, ADR 0017), stands (a standing select is an answered question)

**undecided**:
A ledger item the runner owes: a live Hypothesis that is `inconclusive`, neither supported nor contradicted, with no planned Experiment as parent or target. The lone candidate's select waits on it, because the rival may yet be one.
_Avoid_: untested (that is a proposed Hypothesis with no Experiment)

**premise**:
A live `depends_on` target of a candidate, under another question. While that question is open or reopened, the candidate's question's turn waits on it, so neither a select nor a choice is owed there. Once it selects a Hypothesis other than the target, the runner owes a `premise` item until the edge is re-pointed, the dependent superseded, or the target selected. `hyp check` warns of two questions whose premises make each other wait ([spec §8.1](replay-spec.md#81-ledger-what-the-tree-owes)).
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
The write-time rules the CLI checks before it appends an Experiment, a Result or a Select. Each finding is a refusal (exit 1), a bad-arguments finding (exit 2) or a warning that changes no exit code ([spec §5.1](replay-spec.md#51-write-time-refusals)). Replay does not repeat them, so a merged log can hold an event they would refuse. In code: `replay.State.AdmitExp`, `replay.State.AdmitResult`, `replay.Projection.AdmitSelect` and `replay.Finding`.
_Avoid_: validation (that is `replay.Validate`, which replay also runs on every event)

**Refusal**:
hyp's no to a write: exit 1, and nothing is appended. It comes from `replay.Validate`, which reads stored events and also runs in replay, or from Admission, which runs only at write time and may read derived state.
_Avoid_: rejection (that is what replay reports), error

**Cite**:
A token in a write command's text that names a node: a display id such as `R-BV4CWV`, or a placeholder such as `R-<new>` that stands for one. hyp refuses a write whose cite is no node of the store, with exit 2 (spec §5.1). A token after `<name>:` cites another store and is not checked: a new tree cites the tree it came from as `<tree>:R-…`, and copies the evidence and recipe versions it reads into itself, since an evidence pointer stays inside its store.
_Avoid_: reference for this (a reference is what a flag such as `--parent` resolves)

**Round**:
One Experiment as `hyp signals` lists it: its statuses at add, its design marks and its non-retracted Results with their status moves. In code: `signals.Round`.
_Avoid_: iteration, cycle, step

**Signal**:
A design mark that `hyp signals` lists from the events alone: one-rival or D1. Signals never change a status, and none tracks progress: that is the ledger's (ADR 0017). In code: `signals.Signal`, computed by `signals.Compute`.
_Avoid_: alert, error

**wasted**, **pending**, **no-movement**, **spin**:
Retired (ADR 0017). They judged a round by whether a Result changed a status. The ledger (spec §8.1) is the one account of progress.
_Avoid_: stall (alone: a stall is a ledger item that does not settle)

**one-rival**:
A round whose parent had live rivals at its add and whose targets name none of them ([spec §8](replay-spec.md#8-projection-what-replay-returns), Rivals at add). `hyp add exp` refuses such an Experiment unless `--target none` is passed ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: no-rival, single-rival

**D1**:
A round whose parent or a target was not a live rival at its add. `hyp add exp` warns `D1:` for the same Hypotheses, because both read one rule ([spec §5.1](replay-spec.md#51-write-time-refusals)).
_Avoid_: dead rival, invalid target

**verifier-loop**:
Retired with the verifier role (ADR 0015). It was a select that collected two or more insufficient verifier verdicts in one chain.
_Avoid_: spin (retired; it was about a Hypothesis's status)

**unjudged**:
Retired. It marked a target that no Result of its Experiment gave a verdict on. A target is a promise made before the run, so `hyp record` now refuses a Result that leaves one without a verdict (see Target).
_Avoid_: untested (that is a DesignQuestion reason: every live child is proposed)

**Trial**:
One independent, scored repeat of an attempt whose outcome can vary, such as a fresh agent given one task or one session replay from a cut. A rubric counts trials ("T correct of 6"). A trial can take a whole Experiment run, as a session replay does, or share one with other trials. hyp does not record trials; the Experiment's body or method says how many and how each is scored.
_Avoid_: run when the count is meant (say "3 of 5 trials"); trial for a fixed case a script checks, a probe that fixes the trial command, or an instrument check

### Recipes

**Recipe**:
The kind of Bundle that says how to run one kind of subject and what to measure: its procedure (Launch, Doctor, Drive, Cleanup) and its checks. It names no tree id and never prints a verdict, so the same recipe version serves any Experiment and any tree.
_Avoid_: harness script, driver, wrapper; recipe for the rubric (a rubric is a separate bundle the Experiment pins)

**Bundle**:
A versioned directory that holds a description and the scripts it runs, identified by its ref and hash, not by the directory it sits in (ADR 0005). There are two kinds: a **Recipe**, kept in `recipes/`, says how to run a subject and what to measure; a **Rubric**, kept in `rubrics/`, says how measures become verdicts. Both share one format, are gated by their known answers and are handled by `hyp bundle`. An Experiment pins at most one of each.
_Avoid_: recipe folder, package; recipe for a bundle of either kind

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
One named line a check prints about one subject: a pass or a fail, such as `S1 fail`, or a value, such as `stored rows 4`. A recipe says which measures its checks print. A measure is not a verdict: only the Experiment's rubric turns measures into verdicts, and an Experiment with no targets records measures and nothing else.
_Avoid_: verdict (a measure judges no Hypothesis); measure for the script that turns measures into verdicts (that is the rubric); measurement script

**Known answer**:
A case under `known-answers/NAME/` that hyp runs before every run of its instrument (ADR 0002). It runs the check exactly as the run does and compares its exit code, stdout and the files it leaves under `$HYP_OUT` (ADR 0003).
_Avoid_: fixture (alone), golden test

**Case**:
One known answer as the gate ran it: its name and its outcome, `ok` or why it failed. The sandbox record and the command record list one per case under `known_answers` ([sandbox spec §5.1](sandbox-spec.md#51-records)). In code: `cmdrecord.Case`.
_Avoid_: test case (alone), result (a Result is a tree node)

**External line**:
The `External:` line under an Evidence or Standalone command that names what the command reads or writes outside its bundle, or `none`. `hyp bundle check` warns when it is missing or still reads `TODO` (ADR 0008).
_Avoid_: dependencies line, inputs line

**Kept input**:
A file an operator argument names, which hyp copies beside the command record as `inputs/N/<base name>`, hashes and lists under `inputs` in `command.json` before the first command runs; the script gets another copy. A file over 1 MB is hashed, not kept ([sandbox spec §5.1](sandbox-spec.md#51-records)). It binds the subject when the command runs, not when the Experiment is added (ADR 0008). A directory argument is not copied; its tree's hash is listed under `dirs`. In code: `cmdrecord.Input` and `cmdrecord.Dir`.
_Avoid_: input binding, arg hashes, inputs field (ADR 0008 rejected one on the Experiment)

**Containment**:
What hyp derives, from a Result's files, about whether it can vouch for the script that decided the verdict (spec §7.6). A Result is **uncontained** when one of §7.6's three reasons holds. The mark is read from files each time and never changes a status. Containment is about one Result's evidence. It is not self-contained, which is about a recipe relying only on its own files, and it is not sandbox isolation, which is about where commands run.
_Avoid_: isolation, contained (for a sandbox)

**Uncontained**:
A Result whose rubric or check hyp cannot vouch for: a script outside what a versioned bundle hashes (one that ends in `.py` or `.sh`, has an execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`), a pinned bundle without a clean command record whose known answers passed, or a command record that matches no pin (spec §7.6).
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
The thing under test. Of an Experiment, what the repository holds or the tree keeps: a commit, and for a throwaway arm one patch over it, both assembled in a sandbox ([sandbox spec §5.3](sandbox-spec.md#53-the-subject)); a binary built from it is named by its sha256. It has one shape: a base commit, plus what is given over it, at most one patch and input files (ADR 0020). Of a Hypothesis, what embodies its claim, in the same shape: a base, at most one patch over it, and files each named by its sha256, in one `hypothesis.subject` event; the latest tie stands until a Result's run subject holds it, and then a different subject is a new Hypothesis. Of a run, the base, the one patch and the input files its record names, `{base, patch, files}`; a Result keeps one per cited run, which `hyp record` copies, and an old Result's merged `fingerprints` read as one with no base. The word fingerprints stays for the hashed files inside a subject. A run's subject proves which bytes ran; the tie says which claim they stood for, and hyp cannot check that.
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
