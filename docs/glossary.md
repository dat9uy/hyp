# Glossary

The words of hyp v2 (events v2). [state-machine.md](state-machine.md) is the one spec of v2, and its rules win over this file. This file fixes which word means what, and links the section of the spec that holds the rule. `hyp help` lists the commands the binary has. It also names the two scopes of this repository, which [AGENTS.md](../AGENTS.md) applies.

The words of hyp 0.1.0 are in the last section, "Words of hyp 0.1.0 (v1 branch)". Do not use them for a v2 concept.

## Store and events

**Store**:
The folder `.tree/` at the git top level. It holds every Goal of the repository ([spec §1](state-machine.md#1-names-and-locations)). The first `add goal` creates it; hyp has no `init` act ([spec §8](state-machine.md#8-acts)).
_Avoid_: database, index (hyp keeps no table, [spec §5](state-machine.md#5-calculated-tables))

**Event**:
One JSON line in `.tree/events/<actor>.jsonl`. Only hyp writes it, and it is only appended ([spec §1](state-machine.md#1-names-and-locations), [spec §3](state-machine.md#3-events)). Every table and state comes from the events.
_Avoid_: entry, record (alone)

The nine event types ([spec §3](state-machine.md#3-events)):

| Event | Records |
|---|---|
| `goal.add` | a Goal, or a Draft replace of a written Goal |
| `example.place` | a placement of an example, with H0's cell |
| `result.record` | a Result: the cells of one line on one commit |
| `result.retract` | a retract of a Result |
| `select` | the select of a candidate row |
| `budget.reset` | a reset of the budget's boundary |
| `land` | the owner's land of a Goal at a stop; it names the stop and the act that caused it |
| `unselect` | the owner's unselect of a selected Goal, which puts it back to open |
| `read` | what hyp printed to an actor; it goes into the trace, never into the log |

**Log**:
The events of `.tree/events/`: every act that changes a Goal. The log is not the trace ([spec §8](state-machine.md#8-acts), "The trace is not the log").
_Avoid_: trace (for the log)

**Union merge**:
The git merge that `.tree/.gitattributes` names for the events files. Git keeps every line of both branches ([spec §1](state-machine.md#1-names-and-locations)). GitHub does not run it, so [AGENTS.md](../AGENTS.md) says how to merge such a PR.

**Merge order**:
The order replay reads events in: by id, then by actor, then by line bytes. The place of a line in its file decides nothing ([spec §4](state-machine.md#4-merge-order-crash-and-merged-log)).

**Replay**:
The calculation of every table and state from the log, in merge order, on every command. Replay checks each event again at its place ([spec §4](state-machine.md#4-merge-order-crash-and-merged-log), [spec §5](state-machine.md#5-calculated-tables)).
_Avoid_: re-run (a run of a line again is a repeat, which Stable judges)

**Refused at replay**:
An event whose guards fail at its place in the merge order, for example a fourth counted Hypothesis after a merge. Replay skips it and `status` lists it. hyp never deletes it ([spec §4](state-machine.md#4-merge-order-crash-and-merged-log)).

**Calculated table**:
A table that replay calculates in memory: Goals, Lines, Placements, Results, Cells, Rows, Selects, Unselects, Lands and Reads. hyp stores no table, and each guard is a query over the tables or over git ([spec §5](state-machine.md#5-calculated-tables)).
_Avoid_: projection (a v1 word), cache

**Output**:
A file `.tree/outputs/<sha256>` that holds the bytes a Probe printed. A cell holds its hash. Equal outputs are stored once ([spec §5](state-machine.md#5-calculated-tables)). For the compare, the output is the stdout of `run` with one final newline dropped ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).

**Trace**:
What hyp printed to each actor, kept as `read` events and output files in `$HOME/.hyp/traces/<key>/` ([spec §1](state-machine.md#1-names-and-locations)). Every worktree of one clone shares one trace. No commit carries it. `show ACTOR` prints it ([spec §8](state-machine.md#8-acts)).
_Avoid_: log (the log holds only the acts that change a Goal); trace for a run of commands (a v1 meaning)

**Read** (event):
The trace of one command: its role, arguments, exit code and the hashes of its stdout and stderr. hyp writes it before it prints. No guard reads it ([spec §3](state-machine.md#3-events), [spec §8](state-machine.md#8-acts)).

## Ids, actors and roles

**Goal id**, **Result id**:
`G-` or `R-` plus the last six characters of the event's ULID, for example `G-NXHHVA`. A short id that matches two events is exit 2 ([spec §2](state-machine.md#2-ids-actors-and-roles)).

**Line key**:
`line-1`, `line-2` …, in the order the Goal lists its lines ([spec §2](state-machine.md#2-ids-actors-and-roles)).

**Example key**:
The example's name within its Goal, written `<line>/<example>`, for example `line-4/e`. The example name matches `^[a-z0-9][a-z0-9-]{0,39}$`. Each version of an example keeps its key; a change that is not a new version is a new key ([spec §2](state-machine.md#2-ids-actors-and-roles), [spec §6](state-machine.md#6-states)).

**Actor**:
The name an act records as its author, from `--actor` or `$HYP_ACTOR`. It names the events file the act appends to. Every act and every read needs one ([spec §2](state-machine.md#2-ids-actors-and-roles), [spec §8](state-machine.md#8-acts)).
_Avoid_: user, author (alone); actor for a role

**Role**:
What an actor does in an act: `owner`, `grill`, `probe-writer` or `runner`. It comes from `--role` or `$HYP_ROLE`, and a write act needs one. Roles are self-declared: hyp has no authorization layer ([spec §2](state-machine.md#2-ids-actors-and-roles)).

**Owner**:
The role that adds the Goal, selects a candidate, unselects and lands a Goal, and resets the budget. The owner may select at any time, and is the only role that selects when two or more candidates exist. Only the actor that added the Goal selects for the owner ([spec §7](state-machine.md#7-guards), Candidate, New cell and Goal's actor; [spec §8](state-machine.md#8-acts)). The owner answers product questions; the builder chooses the mechanics.
_Avoid_: user, stakeholder, approver

**Grill**:
The role that writes a Goal's examples and places them, and that does the re-grill. Only the grill places ([spec §2](state-machine.md#2-ids-actors-and-roles), [spec §7](state-machine.md#7-guards), Own case). The skill `skills/hyp-grill` teaches it. In that skill the grill also drafts the words and lines with the owner, and the owner records them.
_Avoid_: interviewer, reviewer

**Re-grill**:
The grill's review of a Goal at a stop for the owner. At Two candidates or Budget used, it adjusts an example or places a new one. A selected Goal refuses every place, so its re-grill goes through the owner's `unselect`: the Goal is open again, and then the grill places. A new line is a new Goal ([spec §2](state-machine.md#2-ids-actors-and-roles), [spec §6](state-machine.md#6-states)).

**Probe writer**:
The role that writes a Probe. It writes no event ([spec §2](state-machine.md#2-ids-actors-and-roles)). Whether it can be kept from reading the candidate is deferred ([spec §14](state-machine.md#14-deferred), o11-blind).

**Runner**:
The role that builds commits for a Goal, runs lines on them and selects a single candidate. The runner never places an example ([spec §7](state-machine.md#7-guards), Own case). `skills/hyp-grill/run-rules.md` lists the rules a runner follows that hyp cannot check.
_Avoid_: operator, driver

## Goal and lines

**Goal**:
What a change must make true, held as words and lines and recorded by `add goal`. Its id is `G-…`. A new purpose or a new line is a new Goal, which can cite an earlier, landed Goal with `--from` ([spec §7](state-machine.md#7-guards), From landed; [spec §8](state-machine.md#8-acts)).
_Avoid_: objective, epic, issue

**Words** (of a Goal):
The owner's one sentence that states the Goal, given as `--words TEXT` ([spec §8](state-machine.md#8-acts)).

**Line**:
One claim of a Goal, with a key `line-N`, a text and a kind: gap or guard. A Goal has one to ten lines and at least one gap line ([spec §7](state-machine.md#7-guards), Whole lines; [spec §8](state-machine.md#8-acts)). A line is **met** on a row when every placed example of the line is met there ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).
_Avoid_: acceptance criterion, requirement; limit (a v1 word)

**Gap line**:
A line that is false on H0 and that the build must make true. Each of its examples must fail on H0 ([spec §6](state-machine.md#6-states), [spec §7](state-machine.md#7-guards), Kind).

**Guard line**:
A line that is true on H0 and that the build must keep true. Each of its examples must be met on H0 ([spec §6](state-machine.md#6-states), [spec §7](state-machine.md#7-guards), Kind).
_Avoid_: guard (alone) for a guard line; a guard is a rule of an act

**Draft replace**:
`add goal G …` on a Goal that is still written. It replaces the Goal's words, lines or base and keeps its actor. After the Goal opens, the words and lines are frozen ([spec §7](state-machine.md#7-guards), Draft).
_Avoid_: edit, amend

**Goal state**:
`written`, `open`, `selected` or `landed`, calculated and never stored ([spec §6](state-machine.md#6-states)).
- **written**: a `goal.add` exists, and the Goal is not open.
- **open**: every line holds a placed example. Each current placement has a live cell on H0's row with the verdict its line's kind needs.
- **selected**: the Goal has a select. It allows only the owner's `land` and `unselect`. An unselect puts it back to open.
- **landed**: the owner landed the Goal. It takes no write.

_Avoid_: done, met (as a state word); ended for a Goal that is not landed

## Examples and Probes

**Example**:
A folder `goals/<G>/<line>/<example>/` that holds an `example.txtar` and a `probe/` folder. It checks one line. A line can hold more than one example ([spec §1](state-machine.md#1-names-and-locations), [spec §9](state-machine.md#9-the-example-file-and-the-compare)). hyp identifies it by its example hash.
_Avoid_: test case, fixture, known answer (a v1 word)

**Example file**:
The `example.txtar` of an example. Its sections are `words`, `input/<path>`, `run`, `floor: contains`, `floor: equals`, `ceiling` and `external` ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).

**Floor**:
A literal the output must contain (`floor: contains`) or equal (`floor: equals`). An example has at least one floor. Floors are literal only: hyp has no pattern language ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).
_Avoid_: pattern, regex, assertion

**Ceiling**:
A literal the output must never contain ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).

**External section**:
The `external` section of an example file. It names, one per line, each environment variable the Probe reads from outside itself and the subject ([spec §9](state-machine.md#9-the-example-file-and-the-compare), [spec §10](state-machine.md#10-the-probe-contract)).

**Compare**:
The check that gives an example's verdict from its output. The example is **met** when the output holds every floor and contains no ceiling. Otherwise it is **failed**. The exit code and stderr are kept beside it and never compared ([spec §9](state-machine.md#9-the-example-file-and-the-compare)).

**Verdict**:
`met` or `failed`: the result of the compare for one example on one row. A row with no live cell for an example reads `not run` there ([spec §5](state-machine.md#5-calculated-tables)).
_Avoid_: supports, contradicts, inconclusive (v1 verdicts)

**Probe**:
The `probe/` folder of an example. hyp runs its `probe/run` on a fresh copy of the subject, and the Probe prints what `run` printed and a last line `exit N`. It gives no verdict ([spec §10](state-machine.md#10-the-probe-contract)). hyp does not sandbox it.
_Avoid_: checker, instrument, test script

**Example hash**:
The git tree hash of an example folder, which identifies the example. hyp calculates it from the committed folder, and from the bytes on disk for Committed and `try`. A Probe change, an input change or a floor change gives a new example hash ([spec §1](state-machine.md#1-names-and-locations), [spec §13](state-machine.md#13-what-hyp-reads-from-git)).

**Probe hash**:
The git tree hash of an example's `probe/` folder. The example hash already covers it. Each cell keeps it, so hyp can show which examples share one Probe ([spec §3](state-machine.md#3-events)).

**Generic Probe**:
The Probe that `skills/hyp-grill/probe/run` ships. It runs `$HYP_RUN` in the subject and prints its stdout, then `exit N`. The grill copies it into an example and writes a new Probe only when the example needs one.

**Placement**:
The act `place G LINE/EXAMPLE` and its `example.place` event. It records the example hash and H0's cell for the example. The latest placement of an example key is the current one ([spec §3](state-machine.md#3-events), [spec §5](state-machine.md#5-calculated-tables)).

**Version** (of an example):
`@1`, `@2` …: the count of distinct hashes placed on one example key. A new version only adds floors or ceilings. A changed input, `run`, `external` section or Probe is a new example key ([spec §6](state-machine.md#6-states), [spec §7](state-machine.md#7-guards), Adds only).
_Avoid_: edit, revision

**Adjust**:
The re-grill's placement of a new version of an example, which adds a floor or a ceiling. It moves the budget's boundary ([spec §7](state-machine.md#7-guards), Adds only; [spec §11](state-machine.md#11-budget-and-stops)).

**Example-changing placement**:
A placement that writes a new example key, or a new hash on a key. The latest one sets the boundary of the budget ([spec §11](state-machine.md#11-budget-and-stops)).

## Rows, Results and candidates

**Subject**:
A commit's tree without the top-level entries `goals` and `.tree`. A change to an example, a Probe, a spec or the store makes no new subject ([spec §1](state-machine.md#1-names-and-locations)).
_Avoid_: base plus patch (a v1 subject)

**Subject hash**:
The git tree hash of a commit's subject. hyp calculates it in memory and writes no object ([spec §1](state-machine.md#1-names-and-locations), [spec §13](state-machine.md#13-what-hyp-reads-from-git)). It keys a row and names a Hypothesis. The events store it in the field `row_tree`, and hyp's output still prints it as a tree, for example `select G TREE`.
_Avoid_: row tree, tree (alone): "tree" means only the Hypothesis tree

**Row**:
hyp's record of one Hypothesis: the verdicts of one Goal on one subject hash, keyed by the Goal and the subject hash. It lists the commits it was seen at. H0's row always exists, and any other row exists when a Result names its subject hash ([spec §5](state-machine.md#5-calculated-tables), [spec §6](state-machine.md#6-states)).

**Hypothesis**:
One build the runner offers against a Goal's examples: the subject of a commit. hyp records it as a row, so commits with the same subject hash are one Hypothesis. H0 is the Hypothesis of the Goal's base. H1, H2, H3 are the others, in the order they first ran. The budget counts Hypotheses other than H0 ([spec §5](state-machine.md#5-calculated-tables), Row key; [spec §11](state-machine.md#11-budget-and-stops)).
_Avoid_: tree, first tree (say H1, or the 1st Hypothesis); attempt; candidate (a candidate is a Hypothesis met on every line)

**H0**:
The Hypothesis of the Goal's base: the commit `--base` names, and its subject hash. A gap example fails on H0 and a guard example is met on H0. The budget never counts H0 ([spec §6](state-machine.md#6-states), [spec §11](state-machine.md#11-budget-and-stops)).
_Avoid_: baseline, main

**Rival**:
A Hypothesis the runner builds beside another, from the same parent in the Hypothesis tree, to test a different design. It costs the budget like any Hypothesis ([spec §11](state-machine.md#11-budget-and-stops)).
_Avoid_: rival for a v1 Hypothesis with a status (see the v1 table)

**Potential Hypothesis**:
A build that hyp has not recorded, such as a switch of a Goal's model written as a patch and tried with `hyp try --worktree`. When it meets every example, the examples do not separate it from the build under test, and it goes to the owner with its outputs. The owner chooses which output is wanted; no party calls a build right or wrong. Run with `hyp run`, it is a Hypothesis.
_Avoid_: wrong build, broken variant, right build

**Hypothesis tree**:
The Hypotheses of one Goal, arranged by the git ancestry of their commits. H0 is its root. A Hypothesis is below another when one of its commits descends from one of the other's. hyp reads the arrangement from git and stores no edge ([spec §5](state-machine.md#5-calculated-tables), Hypothesis tree). It is the only tree in hyp's words.
_Avoid_: run tree, tree for a hash

**Leaf**:
A Hypothesis with nothing below it in the Hypothesis tree: none of its row's commits has a descendant among the commits of the Goal's other rows. hyp reads it from git. A missing commit prints `leaf unknown` ([spec §5](state-machine.md#5-calculated-tables), Hypothesis tree; [spec §6](state-machine.md#6-states)).

**Result**:
The record of one `run`: one line of one Goal on one commit, with one cell per placed example. Its id is `R-…`. A retracted Result is never deleted ([spec §3](state-machine.md#3-events), [spec §6](state-machine.md#6-states)).
_Avoid_: finding, outcome

**Cell**:
One example's record on one Result: the example key and hash, the Probe hash, the output and stderr hashes, the exit code and the verdict ([spec §5](state-machine.md#5-calculated-tables)).

**Cell key**:
A cell is keyed by its Result and its example hash, not by the example key and version. So an adjust makes new cells, and the old cells stay as history that is not live under the new placement ([spec §5](state-machine.md#5-calculated-tables), Cell key).

**Live** (cell):
A cell whose Result is not retracted and not refused by Stable, and whose example hash is the current placement's hash ([spec §5](state-machine.md#5-calculated-tables)).
_Avoid_: live for a Hypothesis (a v1 meaning)

**Candidate**:
A leaf row met on every line under the current placements. With one candidate the runner may select it; with two or more, only the owner selects ([spec §7](state-machine.md#7-guards), Candidate).
_Avoid_: winner; supported rival (a v1 meaning)

**Split line**:
A line whose examples give different verdicts on one row. hyp marks it every time it prints. It is never a refusal ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**Unstable**:
A Result that Stable refused, because a repeat on the same bytes gave another verdict. `status` lists it under `unstable:` with both output hashes ([spec §7](state-machine.md#7-guards), Stable; [spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).
_Avoid_: flaky (say unstable, and name the example)

**Select**:
The act `select G TREE` and its `select` event. TREE is the row's subject hash, as hyp prints it. The act records one candidate row as the Goal's answer, with the commit of the row's latest Result. The select is the gate, and its report is the PR body ([spec §8](state-machine.md#8-acts), [spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).
_Avoid_: pick, choose, accept, merge

**Spec** (of a Goal):
The runner's files under `goals/<G>/specs/`, for example a model and its chart. A select needs the selected commit to hold one. When the spec is a model, every rule of its check configs must be named by a switch config, and the folder must hold a chart ([spec §1](state-machine.md#1-names-and-locations), [spec §7](state-machine.md#7-guards), Spec).

**Model**:
A spec folder that holds a `.tla` file. hyp reads its configs and its chart at select, and runs no model checker ([spec §7](state-machine.md#7-guards), Spec). The skill `skills/hyp-statechart` teaches how to write one.

**Check config**:
A `.cfg` file of a model that is not under `specs/switches/`. Its rules are the names it lists under `INVARIANT`, `INVARIANTS`, `PROPERTY` or `PROPERTIES` ([spec §7](state-machine.md#7-guards), Spec).

**Switch config**:
A `.cfg` file under `specs/switches/`. It names the rules that one switch of the model breaks. A rule of a check config that no switch config names makes hyp refuse the select ([spec §7](state-machine.md#7-guards), Spec).

**Chart**:
A `.mmd` file of a spec folder outside `specs/switches/`. A model with no chart is refused at select, and the PR body prints each chart in a `mermaid` block ([spec §7](state-machine.md#7-guards), Spec; [spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**Record branch**:
The branch that holds a Goal's examples and their place events. It merges before the runner's run, so the examples reach the runner's commits ([spec §1](state-machine.md#1-names-and-locations)).

**Own change**:
What the runner wrote in a commit, not what it took in from main: the diff from its base to the commit. The **base** is the first commit on the first-parent walk that is H0 or a merge commit ([spec §13](state-machine.md#13-what-hyp-reads-from-git)).

## Budget and stops

**Budget**:
At most three Hypotheses run since the boundary, H0 never counted. Three is fixed for every Goal. A refused run costs nothing, and a retracted Result still counts ([spec §7](state-machine.md#7-guards), Budget; [spec §11](state-machine.md#11-budget-and-stops)).
_Avoid_: three commits (the unit is the Hypothesis: one subject hash), retries

**Boundary**:
The start of the budget's count: the latest of the latest example-changing placement, the latest `budget.reset` and the latest `unselect`, else the Goal's start ([spec §11](state-machine.md#11-budget-and-stops)).

**Stop**:
A calculated point that ends the runner's runs on a Goal. A stop is not stored: hyp shows it as an event, calculated from the log, on the `stop:` line of the state block: which stop, the act that caused it, its event id and when. Each stop waits for the owner, who may land the Goal ([spec §11](state-machine.md#11-budget-and-stops)):

| Stop | Holds when | Caused by | hyp prints |
|---|---|---|---|
| Selected | the Goal has a select | the `select` | the PR body |
| Two candidates | two or more candidates and no select | the latest run | the table of outputs |
| Budget used | three Hypotheses counted since the boundary and no candidate | the latest run | the table of outputs |

_Avoid_: land for a stop (a stop ends the runs; the land ends the Goal)

**Table of outputs**:
The `outputs:` table of the state block at "Two candidates" and "Budget used": rows by examples. Each cell shows the verdict, the exit code, the output's hash, its size and at most 60 characters of its first line. hyp prints the table and no reading of it ([spec §11](state-machine.md#11-budget-and-stops)).

**Last-run warning**:
The line `last run: after this, G stops for the owner.` hyp prints it while `runs: 2 of 3 used` ([spec §11](state-machine.md#11-budget-and-stops)).

## Acts

**Act**:
A command that writes one event to the log under the store's lock, after its guards hold ([spec §8](state-machine.md#8-acts)).

**Read** (command):
A command that writes nothing in the repository and only a `read` in the trace: `status`, `try`, `show` and `prune`. A read checks no guard ([spec §8](state-machine.md#8-acts)).

| Act | Role | Built |
|---|---|---|
| `add goal` | owner | yes |
| `place` | grill | yes |
| `run` | runner | yes |
| `select` | runner, or owner | yes |
| `land` | owner | yes |
| `unselect` | owner | yes |
| `reset` | owner | specified, not built |
| `retract` | the actor that ran the Result | specified, not built |

"Built" means `hyp help` lists the act. [spec §8](state-machine.md#8-acts) gives each act's guards, in order.

**Run**:
The act `run COMMIT G/LINE`. It runs every placed example of the line on the commit and records one Result ([spec §8](state-machine.md#8-acts)).
_Avoid_: run for the example file's `run` section (say the `run` section)

**Try**:
The read `try G LINE/EXAMPLE`. It runs an example folder as it is on disk, on H0, or with `--worktree` on this checkout's files. It prints the cell that `place` and `run` would record, and records nothing ([spec §8](state-machine.md#8-acts)).
_Avoid_: dry run

**Land**:
The owner's act `land G` and its `land` event, which end a Goal. The owner lands a Goal at a stop: after a select, or at Two candidates or Budget used with no select. The event names the stop it ends and the act that caused it. Land is the owner's decision, not a check of a commit, and it is not coupled to a merge: a Goal can land before the merge, after it, or with no merge. The landed Goal is a fixed snapshot that later Goals cite with `--from` ([spec §6](state-machine.md#6-states), [spec §8](state-machine.md#8-acts)).
_Avoid_: merge, close, recheck (the recheck of a commit on a main that moved is `unselect`, then a run)

**Unselect**:
The owner's act `unselect G` and its `unselect` event. It puts a selected Goal back to open, so the grill can place a new example and the runner can run again, for example on a merge of a main that moved. The rows and cells stay as they were, and the budget starts again at the unselect. The owner may select a candidate again at once; the runner selects again only after a run that records a new cell: a cell of an example version, or of a row, that had not run before the unselect. A landed Goal takes no unselect ([spec §6](state-machine.md#6-states), [spec §7](state-machine.md#7-guards), New cell; [spec §8](state-machine.md#8-acts); [spec §11](state-machine.md#11-budget-and-stops)).
_Avoid_: deselect, reopen, undo

**Snapshot**:
What a landed Goal leaves: its events in the `.tree/` of the tree being read, up to its land, and the outputs they name. hyp stores nothing more for it, and it holds no code, only Results. It is the owner's decision, fixed. `status G` and the return of `land` print its record: the PR body, or the table of outputs when the Goal has no select. A later Goal starts from it with `add goal --from G` ([spec §8](state-machine.md#8-acts), `land`).
_Avoid_: archive, report file

**Reset** (specified, not built):
The owner's act `reset G`. It writes `budget.reset`, which moves the budget's boundary ([spec §8](state-machine.md#8-acts), [spec §11](state-machine.md#11-budget-and-stops)).

**Retract** (specified, not built):
The act `retract R`. It moves a Result to retracted, so its cells stop being live. Only the actor that ran it retracts it ([spec §6](state-machine.md#6-states), [spec §7](state-machine.md#7-guards), Own Result).

**Show**:
The read `show ACTOR`. It prints every read of ACTOR from the trace, oldest first, byte for byte ([spec §8](state-machine.md#8-acts)).

**Prune**:
The read `prune`. It lists the trace folders whose clone no longer exists. With `--remove`, it removes them ([spec §8](state-machine.md#8-acts)).

**Allowed act**:
An act whose table and git guards hold now. hyp prints each one as a ready command line with its keys filled in. An agent runs a line hyp printed, and a skill names no write command ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

## Guards and refusals

**Guard**:
A named rule an act must meet. A refusal names the first guard that fails, in the act's order, with exit 1 ([spec §7](state-machine.md#7-guards)).
_Avoid_: check, validation, rule (alone)

| Guard | Holds when |
|---|---|
| Draft | a replace names a Goal that is still written |
| Whole lines | every line has a kind, one line or more is a gap line, and there are at most ten lines |
| Not ended | the Goal is not landed |
| Open | the Goal is open; `run` on H0 does not check it |
| Goal's actor | the actor added the Goal, for an owner's select, a land, an unselect and a reset |
| Own case | the act's role is grill, for a place |
| Kind | H0 gives the verdict the line's kind needs |
| Committed | the example folder on disk equals its committed bytes |
| Adds only | a new version only adds floors or ceilings, and its hash is new on its key |
| Leaf only | the commit is H0, or no row of the Goal descends from it |
| Bytes | the commit exists |
| Stable | a repeat on the same bytes gives the same verdict |
| Budget | the run adds no fourth counted Hypothesis since the boundary |
| Candidate | the Hypothesis is a candidate; with two or more, the role is owner |
| New cell | after an unselect, a runner's select follows a run that recorded a new cell; the owner's select does not need one |
| Spec | the selected commit holds a file under `goals/<G>/specs/`; a model also names every check-config rule in a switch config, and holds a chart |
| Own path | the commit's own change writes no path of another Goal |
| Selected | the Goal has no select, for `place`, `run` and `select`; it has one, for `unselect` |
| Owner lands | the act's role is owner, for a land |
| Owner unselects | the act's role is owner, for an unselect |
| Stop | the Goal is at a stop, for a land |
| From landed | the Goal that `add goal --from` cites is landed |
| Own Result | the actor ran the Result it retracts |

The guard **Nothing new** is removed: Stable judges every repeat ([spec §7](state-machine.md#7-guards)).

**Refusal**:
hyp's no to an act: exit 1, naming a guard. A refused act appends nothing, except a run that Stable refuses: it records its Result marked refused, and that Result counts nothing ([spec §7](state-machine.md#7-guards), [spec §8](state-machine.md#8-acts)).
_Avoid_: error (exit 1 is a rule, not a fault)

**Exit code**:
`0` recorded or printed, `1` refused by a guard, `2` bad arguments, `3` an environment problem ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

## What hyp prints

**State block**:
The part of a return that shows a Goal. It holds the state, `runs: N of 3 used`, the stop, the waits and each role's allowed acts ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)). `--role` filters it to one role.

**Report**:
The part of a return that says what was recorded, or which guard refused and why ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**PR body**:
The select's report: each line, each example's words, output, stderr, exit code and verdict, and the selected commit and row. The runner pastes it into the PR; hyp writes nothing to GitHub. `status G` prints it again after the Goal is landed, as the record of its decision ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**Byte block**:
Stored bytes printed under a label `<label>, N bytes:`. A cut is said in the label, never inside the bytes ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**Values as held**:
The rule that hyp prints a value as its state holds it, with no quote, escape or marker added ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

**Channel**:
The one block of code, `internal/channel`, that writes every byte hyp prints and turns every failure into its words and exit code ([spec §12](state-machine.md#12-the-return-shape-and-exit-codes)).

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

The spec defers two items until a real event shows the need: o7-actors and o11-blind. A third, o8-end, is closed: the owner's `land` ends a Goal at a stop, and Not ended refuses every write after it. A Goal that never reaches a stop has no end yet ([spec §14](state-machine.md#14-deferred)).

## Scopes of this repository

**Core**:
Every concept hyp has, and everything a user who installs hyp into another repository runs, loads or reads there. A concept stays core when self-review is what tests it.
_Avoid_: product, the tool, hyp proper

**Self-review**:
A Goal of this repository whose lines are about how hyp is tested, not about a concept hyp has, for example a Goal that checks a Probe or a test fixture of hyp itself. Its examples give their line's verdict on H0 before they count, as every Goal's do ([spec §7](state-machine.md#7-guards), Kind), so self-review has no rule of its own. A Goal that changes a concept of hyp takes the scope of that concept, which is core. In hyp 0.1.0, self-review was the recipes on the `v1` branch.
_Avoid_: dogfood (alone), meta, hyp-self

## Words of hyp 0.1.0 (v1 branch)

These words belong to hyp 0.1.0. Its rules are in [replay-spec.md](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md) and [sandbox-spec.md](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md), and they and its skills are on the `v1` branch. The v1 specs and old logs still use them. Do not use them for a v2 concept. Where a v2 word has the same spelling, the v2 meaning above is the current one.

### Tree words

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Tree | The view of one Goal and every node under it, kept as events in a store. | No node tree. "Tree" means only the Hypothesis tree: a Goal's Hypotheses by git ancestry. |
| Store | `.tree/`, found by a walk up from the current directory to the repository root ([spec §1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#1-names-and-locations)). | Store, at the git top level. |
| Event | A v1 line (`v` = 1) with a type from the v1 set ([spec §4](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#4-event-types-closed-set)). | Event, with `v` = 2 and nine types. |
| Replay, Projection | The derivation of every status, and what it returns ([spec §7](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#7-derived-state-forward-pass), [§8](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#8-projection-what-replay-returns)). | Replay and the calculated tables. |
| Goal | A node `G-` with no parent; its lines never change after its add; complete when every live top-level question is answered. | Goal, with states written, open, selected and landed; a Draft replace may change a written Goal. |
| End, Ended | The owner's terminal act on a Goal, and its state ([spec §7.5](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#75-goal)). | Land: the owner's act `land`, and the state `landed`. |
| Discovery Goal, Build Goal | A Goal whose Hypotheses are claims, or commits. hyp stored no kind. | Every v2 Goal runs commits. |
| Constraints block | The end of a Goal's body that `hyp prompt` read. | None; `prompt ROLE` is a later v2 Goal. |
| DesignQuestion, Top-level question, Sub-question | A node `Q-` that asks one design question, and where it sits ([spec §7.4](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#74-designquestion-state)). | None. A new question is a new Goal. |
| Hypothesis | A node `H-` that states one answer as a claim, with a status from verdicts ([spec §7.1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#71-hypothesis-status-first-matching-rule-wins)). | Hypothesis: one build, the subject of a commit, recorded as a row. |
| Live, Rival, Sibling | Statuses and relations of v1 Hypotheses. | Live is a property of a cell. No rival or sibling. |
| Candidate | A supported rival. | Candidate: a leaf row met on every line. |
| Trade-off | A question with two or more candidates. | The stop "Two candidates". |
| Select | `dq.select`: one Hypothesis as the answer of a DesignQuestion ([spec §5.1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#51-write-time-refusals)). | Select: one candidate row of a Goal. |
| Decision, Turn, new_candidate | The record of a question's selects, where it stands, and why it reopened. | The Goal's state and its stop. |
| Experiment, Experiment status, Target, Trial, Round | A node `E-` that pre-registers one run, its state, what it judges, a scored repeat, and a `hyp signals` row. | An example on a line; a `run` on a commit. |
| Result | A node `R-` under an Experiment, with a verdict per target. | Result: one line on one commit. |
| Verdict | supports, contradicts or inconclusive, typed by the operator. | Verdict: met or failed, from the compare. |
| Evidence pointer, Pin, Questionable | Where a Result points, the hash it recorded for a bundle, and the flag when that hash moved. | Outputs by hash, and the example hash on each cell. |
| Stale, depends_on, Reaffirm | A node whose basis may no longer hold, the edge it follows, and the act that renews it. | None. |
| Keyed item, Limit, Line, Reading, Reach | Items inside an event: a case no check ran, a limit on the Goal (`G-…#n`), how a run read owner words, and what an answer could move ([spec §8.2](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#82-keyed-items-limits-lines-and-readings)). | Line: `line-N`, a gap or guard line with examples. A gap a review finds is an example or a new Goal. |
| Answer, Checkpoint | The owner's event, and the runner's report event. | The owner's acts `add goal`, `select`, `land`, `unselect` and `reset`. |
| Aim, Reassess, Coverage | What an Experiment planned to settle, how an answer reached the tree, and which Hypotheses a Result ran on. | None. |
| Out-of-Goal finding | A finding that is a case of no line. | A defect outside the Goal, which the runner lists in the PR. |
| Ledger item, Owed, Awaiting owner, Frontier, Open limits, Status section | The open work and the `hyp status` sections that showed it ([spec §8.1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#81-ledger-what-the-tree-owes)). | The state block: the stop, the waits and the allowed acts. |
| Edge, supersedes, external_ref | Typed links between nodes. | None. |
| Admission, Cite | Write-time rules, and a token that names a node ([spec §5.1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#51-write-time-refusals)). | Guards. |
| Refusal | Exit 1, and nothing is appended ([spec §5](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#5-refusals)). | Refusal; a Stable refusal appends its Result marked refused. |
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
| Self-contained | A recipe that relies on nothing but its files, hyp and the subject ([sandbox spec §1.2](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md#12-hyp-runs-every-command)). | The same rule for a Probe ([spec §10](state-machine.md#10-the-probe-contract)). |
| Containment, Uncontained | Whether hyp could vouch for the script that decided a verdict. | None. |
| Trace, Command trace | hyp's file of one run of a bundle's commands, `command.json`. | Trace: what hyp printed to an actor. |

### Sandbox words

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Sandbox, Sandbox trace, Sandbox state | A disposable copy of the subject at a base commit, its file `sandbox.json`, and its derived state ([sandbox spec §2](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md#2-the-sandbox-machine)). | A fresh copy of the subject, extracted by hash for each Probe run. hyp does not sandbox a Probe. |
| Step: Launch, Doctor, Drive, Evidence, Cleanup | The stages of an Experiment, each a recipe section ([sandbox spec §1.1](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md#11-steps-and-sections)). | None. |
| Kept files | What remove copied out of a sandbox. | Outputs by hash. |
| Store cut | A copy of a store as it stood at one event. | None. |
| Patch, Arm | A throwaway change over the base, and a variant of the subject an Experiment compares ([sandbox spec §5.3](https://github.com/dat9uy/hyp/blob/v1/docs/sandbox-spec.md#53-the-subject)). | A commit; each build is a Hypothesis. |
| Subject | A base commit, at most one patch and input files. | Subject: a commit's tree without `goals` and `.tree`. |

### People and binaries

| Word | In hyp 0.1.0 | In v2 |
|---|---|---|
| Owner | The person hyp asks on real choices, who answers readings, limits and choices. | Owner, a role. |
| Runner | The agent that settles what the ledger owes it, inside a sandbox in a replay. | Runner, a role. |
| Operator | The person or agent that ran an Experiment from outside the sandbox. | None. |
| Actor | The author of a write ([spec §1](https://github.com/dat9uy/hyp/blob/v1/docs/replay-spec.md#1-names-and-locations)). | Actor, unchanged in meaning. |
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
