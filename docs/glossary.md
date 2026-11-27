# Glossary

The words of hyp v2 (events v2). This file fixes which word means what. The rules are in the binary: `hyp spec` prints every event kind, state, act, guard and exit code, and when this file and `hyp spec` differ, `hyp spec` wins. `hyp help` lists the commands the binary has. It also names the two scopes of this repository, which [AGENTS.md](../AGENTS.md) applies.

The words of hyp 0.1.0 are in the last section, "Words of hyp 0.1.0 (v1 branch)". Do not use them for a v2 concept.

## Store and events

**Store**:
The folder `.tree/` at the git top level. It holds every Goal of the repository. The first `add goal` creates it; hyp has no `init` act.
_Avoid_: database, index (hyp keeps no table)

**Event**:
One JSON line in a file of its own, `.tree/events/<actor>/<id>.json`, or in an older store a line of `.tree/events/<actor>.jsonl`. Only hyp writes it, and it is never changed. Every table and state comes from the events.
_Avoid_: entry, record (alone)

The fourteen event types, each with where it is stored in `hyp spec`:

| Event | Records |
|---|---|
| `goal.add` | a Goal, or a Draft replace of a written Goal |
| `examples.place` | a place of the example set: each example it changed, with H0's cell |
| `example.place` | the old form, one example per event; replay reads it, no act writes it |
| `result.record` | a Result: the cells of one line on one commit |
| `result.retract` | a retract of a Result |
| `select` | the select of a candidate row |
| `budget.reset` | a reset of the budget's boundary |
| `land` | the owner's land of a Goal at a stop; it names the stop, the act that caused it and the snapshot |
| `close` | the owner's close of a Goal at any stop or none; it names the snapshot |
| `unselect` | the owner's unselect of a selected Goal, which puts it back to open |
| `read` | what hyp printed to an actor; it goes into the trace, never into the log |
| `finding.add` | a finding, which names no Goal; it goes into the trace, never into the log |
| `finding.fold` | the old form of a name: a fold of one or more findings into one Goal in work, which counts as a name by that Goal; replayed and no longer written; it goes into the trace |
| `finding.refute` | a refute of a finding; it goes into the trace |
| `finding.decline` | the owner's decline of a finding, with a why; it goes into the trace |

**Log**:
The events of `.tree/events/`: every act that changes a Goal. The log is not the trace.
_Avoid_: trace (for the log)

**Union merge**:
The git merge that `.tree/.gitattributes` names for the events files. Git keeps every line of both branches. GitHub does not run it, so [AGENTS.md](../AGENTS.md) says how to merge such a PR.

**Merge order**:
The order replay reads events in: by id, then by actor, then by line bytes. The place of a line in its file decides nothing.

**Replay**:
The calculation of every table and state from the log, in merge order, on every command. Replay checks each event again at its place.
_Avoid_: re-run (a run of a line again is a repeat, which Stable judges)

**Refused at replay**:
An event whose guards fail at its place in the merge order, for example a fourth counted Hypothesis after a merge. Replay skips it and `status` lists it. hyp never deletes it.

**Calculated table**:
A table that replay calculates in memory: Goals, Lines, Placements, Results, Cells, Rows, Selects, Unselects, Lands, Closes and Reads. hyp stores no table, and each guard is a query over the tables or over git.
_Avoid_: projection (a v1 word), cache

**Output**:
A file `outputs/<sha256>` in the clone's trace that holds the bytes a Probe printed. A cell holds its hash. Equal outputs are stored once. A clone that lacks an output runs its example again with `hyp try`. For the compare, the output is the stdout of `run` with one final newline dropped.

**Trace**:
The events of every kind that changes no Goal's state, the Results and the `read` events, with the output files and the cells of the clone, in `$HOME/.hyp/traces/<key>/`. Every worktree of one clone shares one trace. No commit carries it: it is a cache, which a reader fills by running an example again. `show ACTOR` prints the reads.
_Avoid_: log (the log holds only the acts that change a Goal's state); trace for a run of commands (a v1 meaning)

**Read** (event):
The trace of one command: its role, arguments, exit code and the hashes of its stdout and stderr. hyp writes it before it prints. No guard reads it.

## Ids, actors and roles

**Goal id**, **Result id**:
`G-` or `R-` plus the last six characters of the event's ULID, for example `G-` and six characters. A short id that matches two events is exit 2.

**Line key**:
`line-1`, `line-2` …, in the order the Goal lists its lines.

**Example key**:
The example's name within its Goal, written `<line>/<example>`, for example `line-4/e`. The example name matches `^[a-z0-9][a-z0-9-]{0,39}$`. Each version of an example keeps its key; a change that is not a new version is a new key.

**Actor**:
The name an act records as its author, from `--actor` or `$HYP_ACTOR`. It names the events file the act appends to. Every act and every read needs one.
_Avoid_: user, author (alone); actor for a role

**Role**:
What an actor does in an act: `owner`, `grill`, `probe-writer` or `runner`. It comes from `--role` or `$HYP_ROLE`, and a write act needs one. Roles are self-declared: hyp has no authorization layer.

**Owner**:
The role that adds the Goal, selects a candidate, unselects and lands a Goal, and resets the budget. The owner may select at any time, and is the only role that selects when two or more candidates exist. Only the actor that added the Goal selects for the owner (`hyp spec`, Candidate, New cell and Goal's actor). The owner answers product questions; the builder chooses the mechanics.
_Avoid_: user, stakeholder, approver

**Grill**:
The role that writes a Goal's examples in a draft folder and tries them, and that does the re-grill. The grill records no event: the owner adds the Goal with the examples, and only the owner places (`hyp spec`, Owner places). The skill `skills/hyp-grill` teaches it. In that skill the grill also drafts the words and lines with the owner, and the owner records them.
_Avoid_: interviewer, reviewer

**Re-grill**:
The grill's review of a Goal at a stop for the owner. At Two candidates or Budget used, it adjusts an example or places a new one. A Goal that this clone's Results show selected takes the owner's place of a new example, which shows the Goal open again. A Goal that a select event holds refuses every place, so its re-grill goes through the owner's `unselect`: the Goal is open again, and then the owner places. A new line is a new Goal.

**Probe writer**:
The role that writes a Probe. It writes no event. Whether it can be kept from reading the candidate is deferred (o11-blind).

**Runner**:
The role that builds commits for a Goal and runs lines on them. Its Results stay in the clone's trace, and hyp shows the Goal selected when they show a single candidate. The runner never places an example and never selects (`hyp spec`, Owner places). `skills/hyp-grill/run-rules.md` lists the rules a runner follows that hyp cannot check.
_Avoid_: operator, driver

## Goal and lines

**Goal**:
What a change must make true, held as words and lines and recorded by `add goal`. A Goal is a snapshot of the want, or the problem statement, at the moment the owner adds it. Its id is `G-…`. A new purpose or a new line is a new Goal, which can cite an earlier, ended Goal with `--from` (`hyp spec`, From ended).
_Avoid_: objective, epic, issue

**Issue**:
A GitHub issue that tracks the process toward a want, not the want itself: the event that started it, the owner's decisions in the owner's words, the split into Goals, and where each part stands. A Goal holds the want; the issue holds what happens before and around that snapshot. One issue may track several Goals, each named "Part of #N". It closes when each want it tracks is held by an added Goal, or by a new issue that its closing comment names. hyp neither reads nor writes issues.
_Avoid_: ticket, backlog item

**Words** (of a Goal):
The owner's one sentence that states the Goal, given as `--words TEXT`.

**Line**:
One claim of a Goal, with a key `line-N`, a text and a kind: gap or guard. A Goal has one to ten lines and at least one gap line (`hyp spec`, Whole lines). A line is **met** on a row when every placed example of the line is met there.
_Avoid_: acceptance criterion, requirement; limit (a v1 word)

**Reading** (of a line):
One part or one meaning of a line's words that an example checks. The grill shows the owner each reading as a case with the output it expects; where a model exists, each variant that no example separates is one more reading, shown as the model's print and the variant's print. The owner chooses which readings the line means, and each chosen reading is one example, whose `words` section holds it. A new reading is a new example key.
_Avoid_: interpretation, test idea, direction (alone)

**Gap line**:
A line that is false on H0 and that the build must make true. Each of its examples agrees with it on H0: the example fails there (`hyp spec`, Kind).

**Guard line**:
A line that is true on H0 and that the build must keep true. Each of its examples agrees with it on H0: the example is met there (`hyp spec`, Kind).
_Avoid_: guard (alone) for a guard line; a guard is a rule of an act

**Agree** (of an example with its line):
An example agrees with its line when H0 gives it the verdict its line's kind needs: a gap example fails on H0, and a guard example is met on H0. An example that does not agree disagrees, and `place` then refuses the whole set (`hyp spec`, Kind). Agree and disagree are said only of H0; `met` and `failed` stay the verdicts of a build.
_Avoid_: pass on H0, red on base, valid

**Draft replace**:
`add goal G …` on a Goal that is still written. It replaces the Goal's words, lines or base and keeps its actor. After the Goal opens, the words and lines are frozen (`hyp spec`, Draft).
_Avoid_: edit, amend

**Goal state**:
`written`, `open`, `selected` or `ended`, calculated and never stored.
- **written**: a `goal.add` exists, and the Goal is not open. It has an H0 but no verdict on it, so its gap is claimed, not shown. A written Goal may wait; only its H0 ages, and an outdated Goal is replaced by a new Goal.
- **open**: every line holds a placed example. Each current placement has a live cell on H0's row with the verdict its line's kind needs.
- **selected**: the Goal has a select. It allows only the owner's `land`, `unselect` and `close`. An unselect puts it back to open.
- **ended**: the owner landed or closed the Goal. It takes no write. A line under `stop:` names the end: `landed:` or `closed:`.

_Avoid_: done, met (as a state word); unmet, placeholder (say written or open); landed as the state word (say ended, and the line names the land)

## Examples and Probes

**Example**:
A folder `goals/<G>/<line>/<example>/` that holds an `example.txtar` and a `probe/` folder. It checks one line. A line can hold more than one example. hyp identifies it by its example hash.
_Avoid_: test case, fixture, known answer (a v1 word)

**Example file**:
The `example.txtar` of an example. Its sections are `words`, `input/<path>`, `run`, `floor: contains`, `floor: equals`, `ceiling` and `external`.

**Floor**:
A literal the output must contain (`floor: contains`) or equal (`floor: equals`). An example has at least one floor. Floors are literal only: hyp has no pattern language.
_Avoid_: pattern, regex, assertion

**Ceiling**:
A literal the output must never contain.

**External section**:
The `external` section of an example file. It names, one per line, each environment variable the Probe reads from outside itself and the subject.

**Compare**:
The check that gives an example's verdict from its output. The example is **met** when the output holds every floor and contains no ceiling. Otherwise it is **failed**. The exit code and stderr are kept beside it and never compared.

**Verdict**:
`met` or `failed`: the result of the compare for one example on one row. A row with no live cell for an example reads `not run` there.
_Avoid_: supports, contradicts, inconclusive (v1 verdicts)

**Probe**:
The `probe/` folder of an example. hyp runs its `probe/run` on a fresh copy of the subject, and the Probe prints what `run` printed and a last line `exit N`. It gives no verdict. hyp does not sandbox it.
_Avoid_: checker, instrument, test script

**Example hash**:
The git tree hash of an example folder, which identifies the example. hyp calculates it from the committed folder, and from the bytes on disk for Committed and `try`. A Probe change, an input change or a floor change gives a new example hash.

**Probe hash**:
The git tree hash of an example's `probe/` folder. The example hash already covers it. Each cell keeps it, so hyp can show which examples share one Probe.

**Generic Probe**:
The Probe that `skills/hyp-grill/probe/run` ships. It runs `$HYP_RUN` in the subject and prints its stdout, then `exit N`. The grill copies it into an example and writes a new Probe only when the example needs one.

**Placement**:
One example of a place: the act `place G`, which places the Goal's whole example set, and its `examples.place` event. The event records the example hash and H0's cell for each example the place changed. An older log holds `example.place` events, one placement each. The latest placement of an example key is the current one.

**Version** (of an example):
`@1`, `@2` …: the count of distinct hashes placed on one example key. A new version only adds floors or ceilings. A changed input, `run`, `external` section or Probe is a new example key (`hyp spec`, Adds only).
_Avoid_: edit, revision

**Adjust**:
The re-grill's placement of a new version of an example, which adds a floor or a ceiling. It moves the budget's boundary (`hyp spec`, Adds only).

**Example-changing placement**:
A place that writes a new example key, or a new hash on a key; a place that changes no example is refused. The latest one sets the boundary of the budget, one boundary for the whole set.

## Rows, Results and candidates

**Subject**:
A commit's tree without the top-level entries `goals` and `.tree`. A change to an example, a Probe, a spec or the store makes no new subject.
_Avoid_: base plus patch (a v1 subject)

**Subject hash**:
The git tree hash of a commit's subject. hyp calculates it in memory and writes no object. It keys a row and names a Hypothesis. The events store it in the field `row_tree`; hyp prints it as a subject hash, for example `select G HASH`.
_Avoid_: row tree, tree (alone): "tree" means only the Hypothesis tree

**Row**:
hyp's record of one Hypothesis: the verdicts of one Goal on one subject hash, keyed by the Goal and the subject hash. It lists the commits it was seen at. H0's row always exists, and any other row exists when a Result names its subject hash.

**Hypothesis**:
One build the runner offers against a Goal's examples: the subject of a commit. hyp records it as a row, so commits with the same subject hash are one Hypothesis. H0 is the Hypothesis of the Goal's base. H1, H2, H3 are the others, in the order they first ran. The budget counts Hypotheses other than H0 (`hyp spec`, Row key).
_Avoid_: tree, first tree (say H1, or the 1st Hypothesis); attempt; candidate (a candidate is a Hypothesis met on every line)

**H0**:
The Hypothesis of the Goal's base: the commit `--base` names, and its subject hash. A gap example fails on H0 and a guard example is met on H0. The budget never counts H0.
_Avoid_: baseline, main

**Rival**:
A Hypothesis the runner builds beside another, from the same parent in the Hypothesis tree, to test a different design. It costs the budget like any Hypothesis.
_Avoid_: rival for a v1 Hypothesis with a status (see the v1 table)

**Potential Hypothesis**:
A build that hyp has not recorded, such as a variant of a Goal's model written as a patch and tried with `hyp try --worktree`. When it meets every example, the examples do not separate it from the build under test, and it goes to the owner with its outputs. The owner chooses which output is wanted; no party calls a build right or wrong. Run with `hyp run`, it is a Hypothesis.
_Avoid_: wrong build, broken variant, right build

**Hypothesis tree**:
The Hypotheses of one Goal, arranged by the git ancestry of their commits. H0 is its root. A Hypothesis is below another when one of its commits descends from one of the other's. A run reads that from git and its Result carries it; no state or print reads git for it (`hyp spec`, the subject). It is the only tree in hyp's words.
_Avoid_: run tree, tree for a hash

**Leaf**:
A Hypothesis with nothing below it in the Hypothesis tree: no Result on another row carries one of its row's commits among the commits its run's commit descends from. hyp calculates it from the Results, so a rebase, a squash or a new clone changes no leaf (`hyp spec`, the subject).

**Result**:
The record of one `run`: one line of one Goal on one commit, with one cell per placed example. Its id is `R-…`. A retracted Result is never deleted.
_Avoid_: finding, outcome

**Finding**:
A fact an agent found, with a pointer to its evidence, such as a Result id. Its id is `F-…`. Any agent adds one with `hyp finding add`, which names no Goal, or refutes it when its fact is false; it may carry an `add goal` the agent drafted for the owner. A Goal names the findings it is about in its own record: the owner with `hyp add goal --finding F`, the grill in the `findings` section of an example. The owner declines, with `hyp finding decline F --why TEXT`, a finding that no Goal named whose fact is true and that the owner does not want fixed. `hyp finding list` prints each finding with its state: `not folded`, `folded into G`, `settled in G`, `ended in G, not settled`, `refuted` or `declined`, and `hyp status G` prints the findings G names. Its events are in the trace, so a finding changes no Goal's state (`hyp spec`, findings).
_Avoid_: Result (a finding is not the record of a run); out-of-Goal finding (every finding names no Goal)

**Fold** (of a finding):
The state of a finding that a Goal's record names. It reads `folded into G` while a Goal in work names it, `settled in G` once a Goal that names it lands with a select, and `ended in G, not settled` when every Goal that named it ended with no select. Fold is a state, not an act: no command folds. An old `finding.fold` event counts as a name by its Goal (`hyp spec`, finding states).
_Avoid_: link, ref, attach

**Learn**:
What every record is kept for. An agent learns a fact, which is a finding. A Goal names the findings that bear on its want. The owner learns a want from a consequence and states it as a line, a decline or a "not yet". The log keeps what the owner learned or approved. The trace keeps what agents learned, as a cache that may be lost with its clone.
_Avoid_: knowledge (alone), memory

**Cell**:
One example's record on one Result: the example key and hash, the Probe hash, the output and stderr hashes, the exit code and the verdict.

**Cell key**:
A cell is keyed by its Result and its example hash, not by the example key and version. So an adjust makes new cells, and the old cells stay as history that is not live under the new placement (`hyp spec`, Cell key).

**Live** (cell):
A cell whose Result is not retracted and not refused by Stable, and whose example hash is the current placement's hash.
_Avoid_: live for a Hypothesis (a v1 meaning)

**Candidate**:
A leaf row met on every line under the current placements. With one candidate the runner may select it; with two or more, only the owner selects (`hyp spec`, Candidate).
_Avoid_: winner; supported rival (a v1 meaning)

**Split line**:
A line whose examples give different verdicts on one row. hyp marks it every time it prints. It is never a refusal.

**Unstable**:
A Result that Stable refused, because a repeat on the same bytes gave another verdict. `status` lists it under `unstable:` with both output hashes (`hyp spec`, Stable).
_Avoid_: flaky (say unstable, and name the example)

**Select**:
The one candidate row that is the Goal's answer, with the commit of the row's latest Result. With one candidate, hyp calculates the select from the Results of the clone's trace, and no event holds it. With two, the owner's act `select G HASH` and its `select` event hold it; HASH is the row's subject hash, as hyp prints it. The select is the gate, and `hyp status G` prints its PR body.
_Avoid_: pick, choose, accept, merge

**Spec** (of a Goal):
The runner's files under `goals/<G>/specs/`: a Goal's own whole model and its chart, or, when the commit changes the kept model, the variant configs and the chart of that change. A select needs the selected commit to hold one. When the spec is a model, every rule of its check configs must be named by a variant config, and the folder must hold a chart (`hyp spec`, Spec).

**Model**:
A TLA+ module of the machine, with its check config. It is either the kept model of the repository or a Goal's own whole model in its spec folder. hyp reads its configs and its chart at select, and runs no model checker (`hyp spec`, Spec). The skill `skills/hyp-statechart` teaches how to write one.

**Kept model**:
The one model of a repository's machine, in `goals/model/`. It is no Goal's folder, so a runner's own commit may change it. A Goal's model is a change to it, carried only by the Goal's selected commit. At select, each rule that the commit adds or changes against H0 needs a variant config in the Goal's spec folder; a rule left as H0 holds it needs none. At `Selected`, `hyp status G` lists what the commit adds or changes in it (`hyp spec`, Spec).
_Avoid_: shared model, global model

**Step**:
An action of a model that changes state: a side effect, such as an act that writes an event. A comment `\* command: <command>` above the action names the command that takes it. A read or a refusal is a print, not a step.

**Walk**:
The steps that an example's Probe and run section take, in the order they call their commands. An example is a requirement on a walk: its floors and ceilings say what the run section's steps must print. The statechart tool predicts an example from its walk; when it cannot see the walk, it predicts no verdict and lists the example to try on every variant.

**Step log**:
The steps a program really took, as a file of step lines, one per line: `step <Action> <json>`. Any program may append to it. hyp appends one step line for each event it writes while `HYP_STEP_LOG` names a file, with the step its event kind names and the event as stored. For each Probe run hyp sets `HYP_STEP_LOG` to a fresh file; `run` keeps it in the clone's trace when the subject wrote to it, the Result names it by hash, and `hyp show` prints it (`hyp spec`, the step log). A walk is what the kept model says the steps are; a step log is what they were.

**Job**:
What a Goal line names in the kept model: a step, or a rule over several steps. A rule names its line with the source comment `\* source: G:line-N` above it, and only such a rule is a want. Agents also write steps, prints and `model only` rules; these describe the build. A later Goal changes a job by changing the step or rule that a line names.
_Avoid_: feature, capability

**Check config**:
A `.cfg` file of a model that is not under `specs/variants/` (or the older `specs/switches/`). Its rules are the names it lists under `INVARIANT`, `INVARIANTS`, `PROPERTY` or `PROPERTIES` (`hyp spec`, Spec).

**Variant config**:
A `.cfg` file under `specs/variants/`. It names exactly the rules that one variant of the model breaks: a boolean constant that the check config sets FALSE, which stands for one potential Hypothesis. The statechart tool writes it. A spec folder written before the rename holds its variant configs under `specs/switches/`, and hyp reads them there too. A rule of a check config that no variant config names makes hyp refuse the select (`hyp spec`, Spec).

**Chart**:
A `.mmd` file of a spec folder outside its variant config folders. A model with no chart is refused at select, and the PR body prints each chart in a `mermaid` block (`hyp spec`, Spec).

**Record branch**:
The branch that holds a Goal's examples and their place events. It merges before the runner's run, so the examples reach the runner's commits.

**Own change**:
What the runner wrote in a commit, not what it took in from main: the diff from its base to the commit. It may write under its own Goal's folder and under `goals/model/`, the kept model, and under no other Goal's folder. The **base** is the first commit on the first-parent walk that is H0 or a merge commit.

## Budget and stops

**Budget**:
At most three Hypotheses run since the boundary, H0 never counted. Three is fixed for every Goal. A refused run costs nothing, and a retracted Result still counts (`hyp spec`, Budget).
_Avoid_: three commits (the unit is the Hypothesis: one subject hash), retries

**Boundary**:
The start of the budget's count: the latest of the latest example-changing placement, the latest `budget.reset` and the latest `unselect`, else the Goal's start.

**Stop**:
A calculated point that ends the runner's runs on a Goal. A stop is not stored: hyp shows it as an event, calculated from the log, on the `stop:` line of the state block: which stop, the act that caused it, its event id and when. Each stop waits for the owner, who may land the Goal:

| Stop | Holds when | Caused by | hyp prints |
|---|---|---|---|
| Selected | the Goal has a select | the `select` | the PR body |
| Two candidates | two or more candidates and no select | the latest run | the table of outputs |
| Budget used | three Hypotheses counted since the boundary and no candidate | the latest run | the table of outputs |

_Avoid_: land for a stop (a stop ends the runs; the land ends the Goal)

**Table of outputs**:
The `outputs:` table of the state block at "Two candidates" and "Budget used": rows by examples. Each cell shows the verdict, the exit code, the output's hash, its size and at most 60 characters of its first line. hyp prints the table and no reading of it.

**Last-run warning**:
The line `last run: after this, G stops for the owner.` hyp prints it while `runs: 2 of 3 used`.

## Acts

**Act**:
A command that writes one event to the log under the store's lock, after its guards hold.

**Read** (command):
A command that writes nothing in the repository and only a `read` in the trace: `status`, `try`, `show` and `prune`. A read checks no guard.

| Act | Role | Built |
|---|---|---|
| `add goal` | owner | yes |
| `place` | grill | yes |
| `run` | runner | yes |
| `select` | runner, or owner | yes |
| `land` | owner | yes |
| `close` | owner | yes |
| `unselect` | owner | yes |
| `reset` | owner | specified, not built |
| `retract` | the actor that ran the Result | specified, not built |

"Built" means `hyp help` lists the act. `hyp spec` gives each act's guards, in order.

**Run**:
The act `run COMMIT G/LINE`. It runs every placed example of the line on the commit and records one Result.
_Avoid_: run for the example file's `run` section (say the `run` section)

**Try**:
The read `try G LINE/EXAMPLE`. It runs an example folder as it is on disk, on H0, or with `--worktree` on this checkout's files. It prints the cell that `place` and `run` would record, and records nothing.
_Avoid_: dry run

**Land**:
The owner's act `land G` and its `land` event, which end a Goal. The owner lands a Goal at a stop: on the selected commit, or at Two candidates or Budget used with no commit. The event names the stop it ends, the act that caused it and the snapshot. A land on a commit carries the evidence: the commit, its tree, its row and each example's verdict there, run first when the clone holds no Result. It is not coupled to a merge: a Goal can land before the merge, after it, or with no merge. The landed Goal is a fixed snapshot that later Goals cite with `--from`.
_Avoid_: merge, recheck (the recheck of a commit on a main that moved is `unselect`, then a run)

**Close**:
The owner's act `close G` and its `close` event, which end a Goal the owner will not go on with. The owner closes a Goal that is not ended, at any stop or none. The event names the snapshot. A closed Goal keeps its rows and verdicts as they were, and later Goals cite it with `--from` like a landed one.
_Avoid_: abandon, delete, land (a land needs a stop; a close does not)

**Unselect**:
The owner's act `unselect G COMMIT` and its `unselect` event, which names the commit it rejects and its row. It puts a selected Goal back to open, so the owner can place a new example and the runner can run again, for example on a merge of a main that moved. No clone shows the Goal selected on the rejected row again. The rows and cells stay as they were, and the budget starts again at the unselect. hyp shows a select again only after a run that records a new cell: a cell of an example version, or of a row, that had not run before the unselect. An ended Goal takes no unselect (`hyp spec`, New cell; `hyp spec`; `hyp spec`).
_Avoid_: deselect, reopen, undo

**Snapshot**:
What an ended Goal leaves: the Goal's events that its land or close names, in the `.tree/` of the tree being read, and the outputs they name. The end stores only their ids, and the snapshot holds no code, only Results. It is the owner's decision, fixed: replay refuses an event of the Goal that another clone wrote before the end and that the end does not name, so the Goal holds what the owner saw in every clone. `status G` prints its record: the Goal's words, the why it was added with, each line with its kind, the why of its land or close, and the PR body, or the table of outputs when the Goal has no select. The return of `land` or `close` does not print that record: it prints the state block, without the words, the whys or the lines. A later Goal starts from it with `add goal --from G` (`hyp spec`, `land`).
_Avoid_: archive, report file

**Reset** (specified, not built):
The owner's act `reset G`. It writes `budget.reset`, which moves the budget's boundary.

**Retract** (specified, not built):
The act `retract R`. It moves a Result to retracted, so its cells stop being live. Only the actor that ran it retracts it (`hyp spec`, Own Result).

**Show**:
The read `show ACTOR`. It prints every read of ACTOR from the trace, oldest first, byte for byte.

**Prune**:
The read `prune`. It lists the trace folders whose clone no longer exists. With `--remove`, it removes them.

**Allowed act**:
An act whose table and git guards hold now. hyp prints each one as a ready command line with its keys filled in. An agent runs a line hyp printed, and a skill names no write command.

## Guards and refusals

**Guard**:
A named rule an act must meet. A refusal names the first guard that fails, in the act's order, with exit 1; a place that places none of the set names every cause, each with its guard.
_Avoid_: check, validation, rule (alone)

| Guard | Holds when |
|---|---|
| Draft | a replace names a Goal that is still written |
| Whole lines | every line has a kind, one line or more is a gap line, and there are at most ten lines |
| Not ended | the Goal is not landed or closed, and at replay the end's snapshot names the event |
| Open | the Goal is open; `run` on H0 does not check it |
| Goal's actor | the actor added the Goal, for an owner's select, a land, a close, an unselect and a reset |
| Grill places | the act's role is grill, for a place |
| Kind | H0 gives the verdict the line's kind needs |
| Whole set | for a place: every line holds an example folder, every example placed before still has its folder, and every folder name is an example name |
| Committed | each example folder on disk equals its committed bytes |
| Adds only | a new version only adds floors or ceilings, its hash is new on its key, and a place changes at least one example |
| Leaf only | the commit is H0, or no row of the Goal descends from it |
| Bytes | the commit exists |
| Stable | a repeat on the same bytes gives the same verdict |
| Budget | the run adds no fourth counted Hypothesis since the boundary |
| Candidate | the Hypothesis is a candidate; with two or more, the role is owner |
| New cell | after an unselect, a runner's select follows a run that recorded a new cell; the owner's select does not need one |
| Spec | the selected commit holds a file under `goals/<G>/specs/`; a model also names every check-config rule in a variant config, and holds a chart |
| Own path | the commit's own change writes no path of another Goal |
| Selected | the Goal has no select, for `run` and `select`; no select event holds it, for `place`; it has one, for `unselect` |
| Owner lands | the act's role is owner, for a land |
| Owner closes | the act's role is owner, for a close |
| Owner unselects | the act's role is owner, for an unselect |
| Stop | the Goal is at a stop, for a land |
| From ended | the Goal that `add goal --from` cites is landed or closed |
| Own Result | the actor ran the Result it retracts |

The guard **Nothing new** is removed: Stable judges every repeat.

**Refusal**:
hyp's no to an act: exit 1, naming a guard. A refused act appends nothing, except a run that Stable refuses: it records its Result marked refused, and that Result counts nothing.
_Avoid_: error (exit 1 is a rule, not a fault)

**Exit code**:
`0` recorded or printed, `1` refused by a guard, `2` bad arguments, `3` an environment problem.

## What hyp prints

**State block**:
The part of a return that shows a Goal. It holds the state, `runs: N of 3 used` while the Goal has not ended, the stop, the waits and each role's allowed acts. An ended Goal prints no runs count: the runs are in each clone's trace, and an ended Goal reads the same in every clone. `--role` filters it to one role.

**Report**:
The part of a return that says what was recorded, or which guard refused and why.

**PR body**:
The select's report: the Goal's words, each line with its kind, each placed example's key and words, each chart of the model, and the selected commit and row. It holds no result of an example: every placed example is met by the select, so a result teaches the owner nothing at the merge. `hyp try G LINE/EXAMPLE --commit C` prints an example's result on the selected commit C. The runner pastes it into the PR; hyp writes nothing to GitHub. `status G` prints it again after the Goal is landed, as the record of its decision.

**Byte block**:
Stored bytes printed under a label `<label>, N bytes:`. A cut is said in the label, never inside the bytes.

**Values as held**:
The rule that hyp prints a value as its state holds it, with no quote, escape or marker added.

**Channel**:
The one block of code, `internal/channel`, that writes every byte hyp prints and turns every failure into its words and exit code.

## Grill words

These words come from `skills/hyp-grill`. hyp stores none of them.

**Decision** (of the owner):
A choice the owner made in the grill. It is held by the Goal's words, a line or an example, and nothing else carries it to the runner.

**Not covered**:
The part of the grill's hand-off that lists each decision no example checks, with where the decision came from. It is never passed to the runner as an extra rule.

**Surface**:
A file, output or command where a line's words can be true or false. The grill lists every surface of a line before it writes examples.

**Hand-off**:
What the grill leaves for the owner: the record branch, the Goal's lines, the owner's decisions with their words, and "Not covered". A runner works from the Goal and the run rules; if it reads the hand-off, it is background, and no "Not covered" item is a task.

**Run rules**:
The rules a runner follows that hyp cannot check, in `skills/hyp-grill/run-rules.md`.

## Deferred

hyp defers two items until a real event shows the need: o7-actors and o11-blind. A third, o8-end, is closed: the owner's `land` ends a Goal at a stop, the owner's `close` ends it at any stop or none, and Not ended refuses every write after either.

## Scopes of this repository

**Core**:
Every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A concept stays core when self-review is what tests it.
_Avoid_: product, the tool, hyp proper

**Self-review**:
A Goal of this repository whose lines are about how hyp is tested, not about a concept hyp has, for example a Goal that checks a Probe or a test fixture of hyp itself. Its examples give their line's verdict on H0 before they count, as every Goal's do (`hyp spec`, Kind), so self-review has no rule of its own. A Goal that changes a concept of hyp takes the scope of that concept, which is core. In hyp 0.1.0, self-review was the recipes on the `v1` branch.
_Avoid_: dogfood (alone), meta, hyp-self

## Words of hyp 0.1.0 (v1 branch)

These words belong to hyp 0.1.0. Its rules are in [replay-spec.md](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md) and [sandbox-spec.md](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md), and they and its skills are on the `v1` branch. The v1 specs and old logs still use them. Do not use them for a v2 concept. Where a v2 word has the same spelling, the v2 meaning above is the current one.

### Tree words

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Tree | The view of one Goal and every node under it, kept as events in a store. | No node tree. "Tree" means only the Hypothesis tree: a Goal's Hypotheses by git ancestry. |
| Store | `.tree/`, found by a walk up from the current directory to the repository root. | Store, at the git top level. |
| Event | A v1 line (`v` = 1) with a type from the v1 set. | Event, with `v` = 2 and fourteen types. |
| Replay, Projection | The derivation of every status, and what it returns. | Replay and the calculated tables. |
| Goal | A node `G-` with no parent; its lines never change after its add; complete when every live top-level question is answered. | Goal, with states written, open, selected and ended; a Draft replace may change a written Goal. |
| End, Ended | The owner's terminal act on a Goal, and its state. | Land and close: the owner's acts `land` and `close`, and the state `ended`. |
| Discovery Goal, Build Goal | A Goal whose Hypotheses are claims, or commits. hyp stored no kind. | Every v2 Goal runs commits. |
| Constraints block | The end of a Goal's body that `hyp prompt` read. | None; `prompt ROLE` is a later v2 Goal. |
| DesignQuestion, Top-level question, Sub-question | A node `Q-` that asks one design question, and where it sits. | None. A new question is a new Goal. |
| Hypothesis | A node `H-` that states one answer as a claim, with a status from verdicts. | Hypothesis: one build, the subject of a commit, recorded as a row. |
| Live, Rival, Sibling | Statuses and relations of v1 Hypotheses. | Live is a property of a cell. No rival or sibling. |
| Candidate | A supported rival. | Candidate: a leaf row met on every line. |
| Trade-off | A question with two or more candidates. | The stop "Two candidates". |
| Select | `dq.select`: one Hypothesis as the answer of a DesignQuestion. | Select: one candidate row of a Goal. |
| Decision, Turn, new_candidate | The record of a question's selects, where it stands, and why it reopened. | The Goal's state and its stop. |
| Experiment, Experiment status, Target, Trial, Round | A node `E-` that pre-registers one run, its state, what it judges, a scored repeat, and a `hyp signals` row. | An example on a line; a `run` on a commit. |
| Result | A node `R-` under an Experiment, with a verdict per target. | Result: one line on one commit. |
| Verdict | supports, contradicts or inconclusive, typed by the operator. | Verdict: met or failed, from the compare. |
| Evidence pointer, Pin, Questionable | Where a Result points, the hash it recorded for a bundle, and the flag when that hash moved. | Outputs by hash, and the example hash on each cell. |
| Stale, depends_on, Reaffirm | A node whose basis may no longer hold, the edge it follows, and the act that renews it. | None. |
| Keyed item, Limit, Line, Reading, Reach | Items inside an event: a case no check ran, a limit on the Goal (`G-…#n`), how a run read owner words, and what an answer could move. | Line: `line-N`, a gap or guard line with examples. A gap a review finds is an example or a new Goal. |
| Answer, Checkpoint | The owner's event, and the runner's report event. | The owner's acts `add goal`, `select`, `land`, `unselect` and `reset`. |
| Aim, Reassess, Coverage | What an Experiment planned to settle, how an answer reached the tree, and which Hypotheses a Result ran on. | None. |
| Out-of-Goal finding | A finding that is a case of no line. | A finding (`hyp finding add`), which names no Goal; a Goal names it. |
| Ledger item, Owed, Awaiting owner, Frontier, Open limits, Status section | The open work and the `hyp status` sections that showed it. | The state block: the stop, the waits and the allowed acts. |
| Edge, supersedes, external_ref | Typed links between nodes. | None. |
| Admission, Cite | Write-time rules, and a token that names a node. | Guards. |
| Refusal | Exit 1, and nothing is appended. | Refusal; a Stable refusal appends its Result marked refused. |
| Signal, one-rival, D1 | Design marks that `hyp signals` printed. | None. |

### Recipe words

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Bundle, Frozen, Draft | A versioned folder of a description and scripts; a version a tree pins; a version no tree pins. | An example folder, by its example hash. Draft is now a guard. |
| Recipe, Harness recipe, Environment recipe, Stack, Recipe ref, Recipe library | How to run one kind of subject, its roles in a stack, and how it is named and kept. | An example's `run` section and its Probe. |
| Rubric | The rule that turned measures into verdicts. | The compare: floors and ceilings. |
| Instrument | A bundle with an Evidence or Standalone command, gated by its known answers. | None: say example and Probe. An example gives its line's verdict on H0 before it is placed. |
| Command, Check, Measure, Standalone | A `Run:` line, a command under Evidence or Standalone, a line a check printed, and a section that is not a step. | The `run` section and the Probe. |
| Known answer, Known-answer run | A fixed input of an instrument with its expected output, and one run of it by the gate. | An example placed on H0, which Kind checks. |
| External line | The `External:` line under a command. | The `external` section of an example file. |
| Kept input | A file an operator argument named, copied and hashed before a run. | An `input/<path>` section of an example file. |
| Self-contained | A recipe that relies on nothing but its files, hyp and the subject. | The same rule for a Probe. |
| Containment, Uncontained | Whether hyp could vouch for the script that decided a verdict. | None. |
| Trace, Command trace | hyp's file of one run of a bundle's commands, `command.json`. | Trace: what hyp printed to an actor. |

### Sandbox words

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Sandbox, Sandbox trace, Sandbox state | A disposable copy of the subject at a base commit, its file `sandbox.json`, and its derived state. | A fresh copy of the subject, extracted by hash for each Probe run. hyp does not sandbox a Probe. |
| Step: Launch, Doctor, Drive, Evidence, Cleanup | The stages of an Experiment, each a recipe section. | None. |
| Kept files | What remove copied out of a sandbox. | Outputs by hash. |
| Store cut | A copy of a store as it stood at one event. | None. |
| Patch, Arm | A throwaway change over the base, and a variant of the subject an Experiment compares. | A commit; each build is a Hypothesis. |
| Subject | A base commit, at most one patch and input files. | Subject: a commit's tree without `goals` and `.tree`. |

### People and binaries

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Owner | The person hyp asks on real choices, who answers readings, limits and choices. | Owner, a role. |
| Runner | The agent that settles what the ledger owes it, inside a sandbox in a replay. | Runner, a role. |
| Operator | The person or agent that ran an Experiment from outside the sandbox. | None. |
| Actor | The author of a write. | Actor, unchanged in meaning. |
| Harness | The agent runtime that loads a skill and hosts the runner. | Unchanged. |
| Sandbox hyp, Operator hyp | The binary inside a sandbox, and the binary outside it. | One binary. |

### Retired words

A retired word stays only in old logs, ADRs and history.

| Retired word | Retired by | Say instead |
|---|---|---|
| Confirm | ADR 0019 | the select |
| Dossier | ADR 0019 | nothing |
| Lapse | ADR 0019 | `new_candidate` (v1) |
| Verifier, verifier-loop | ADR 0015 | the select is the gate |
| unjudged | [ADR 0018](https://github.com/dat9uy/hyp/blob/v1/docs/adr/0018-an-experiment-is-a-pre-registered-result.md) | a refusal (v1) |
| wasted, pending, no-movement, spin | ADR 0017 | the ledger (v1) |
| verdict check, verdict rule | [ADR 0018](https://github.com/dat9uy/hyp/blob/v1/docs/adr/0018-an-experiment-is-a-pre-registered-result.md) | Rubric (v1); in v2, the compare |
| command record, sandbox record | owner decision of 2026-11-14 | Command trace, Sandbox trace (v1) |
