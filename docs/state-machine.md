# hyp v2 state machine

This is the one spec of hyp v2 (events v2). It states what hyp stores, what it
calculates from what it stores, each state and what it refuses, every act with
its guards and exit codes, and what happens on a crash or a merged log. Where
an earlier design text differs from it, this file wins.

Tags: **Owner said** marks a rule the owner stated or accepted, with the date.
Every other rule here is derived from such a rule, and names it. A rule cites
only what a reader of the repository can check: an issue, a PR, a commit or a
Goal.

## 1. Names and locations

- **The store** is one folder, `.tree/`, at the git top level
  (`git rev-parse --show-toplevel`). It holds every Goal of the repository.
- **Events** are kept one file per actor: `.tree/events/<actor>.jsonl`, one JSON
  line per event. Only hyp writes them.
- **Union merge.** `.tree/.gitattributes` holds `events/*.jsonl merge=union`, so
  git merges two branches of one events file by keeping every line of both.
  Replay orders events by id, never by their place in the file (§4).
- **The trace** of a clone is one folder outside every worktree and outside
  the git directory, in hyp's own folder: `$HOME/.hyp/traces/<key>/`. hyp
  does not read `$XDG_STATE_HOME` and writes nothing under `~/.local`. Owner
  said, 2026-11-19 (record PR #238): "I don't want to use $XDG_STATE_HOME at
  all". `<key>` is the sha256 of the clone's git common directory
  (`git rev-parse --path-format=absolute --git-common-dir`, symlinks
  resolved), which every worktree of the clone shares. The folder holds
  `events/<actor>.jsonl`, `outputs/<sha256>` and a file `clone` that names the
  common directory. It is not the log: no git command sees it, and no commit
  carries it (§8, "The trace is not the log"). It also keeps the evidence of
  the clone's acts: every cell, Result and output (§8, "Git keeps the acts;
  the trace keeps the evidence"). Reads that an earlier hyp
  wrote under `~/.local/state/hyp` are let go: `show` does not print them, and
  hyp never reads, moves or removes that folder; the owner removes it by hand
  (record PR #238, decision 8).
- **Goal folders** are written by the owner and the agents, never by hyp:

  ```
  goals/<Goal>/<line>/<example>/   one example folder
      example.txtar                the example file (§9)
      probe/                       the Probe code (§10)
  goals/<Goal>/specs/              the Goal's spec; a model: a .tla module, its
                                   check configs, variants/<variant>.cfg and chart.mmd
  ```

- **The subject** of a commit is the commit's git tree without the top-level
  entries `goals` and `.tree`. A change to an example, a Probe, a spec or the
  store makes no new subject. Its **subject hash** is the git tree hash of that
  tree, calculated as §13 says. The events store it in the field `row_tree`
  (`base_tree` for H0 in `goal.add`).
- An example folder is identified by its **example hash**: the git tree hash of
  the folder (§13). The hash of a line folder or a Goal folder is never used.
- Example folders travel on the Goal's record branch, which merges before the
  runner's run, so the examples reach the runner's commits.

## 2. Ids, actors and roles

- **Event id**: a ULID, unique in the store, sorting after the newest id the
  store holds.
- **Goal id**: `G-` plus the last six characters of its `goal.add` event's id,
  for example `G-NXHHVA`. **Result id**: `R-` plus the
  last six characters of its `result.record` event's id. A short id that matches
  two events is refused as ambiguous (exit 2).
- **Line key**: `line-1`, `line-2` … in the order the Goal lists its lines.
  **Example key**: `<line>/<example>`, for example `line-3/a`. An example name
  matches `^[a-z0-9][a-z0-9-]{0,39}$`.
- **Row key**: the subject hash, see §5.
- **Actor**: a name matching `^[a-z0-9][a-z0-9-]{0,39}$`. It names the events
  file the act appends to.
- **Roles**: `owner`, `grill`, `probe-writer`, `runner`. The grill also does the
  re-grill. The Probe writer writes no event.
- The actor and the role are self-declared. A guard that reads them catches an
  honest mistake, not a lie. hyp has no authorization layer.

## 3. Events

Every event is one JSON line with this envelope (with `v` = 2):

| Field | Holds |
|---|---|
| `v` | `2`. A line with another `v` does not parse. |
| `id` | The event's ULID. |
| `ts` | hyp's time, RFC 3339 UTC. |
| `actor` | The actor; equal to the events file's base name. |
| `type` | One of the eleven types below. |
| `node` | The id the event is about: the Goal's `goal.add` id, the Result's `result.record` id for `result.retract`, or the event's own id for `read`. |
| `payload` | The type's payload, parsed strictly: an unknown field does not parse. |
| `why` | The actor's short reason, optional. |

The eleven events. Git keeps `goal.add`, `examples.place`, `select`, `land`
and `unselect`; the clone's trace keeps `result.record`, `cells.save` and
`read` (§8, "Git keeps the acts; the trace keeps the evidence"). An older log
keeps `result.record` and the old forms below in git, and replay reads them
there.

| Event | Written by | `node` | Payload |
|---|---|---|---|
| `goal.add` | `add goal` | its own id; a Draft replace names the replaced Goal's id | role, words, H0 commit, H0 subject hash, budget (3), `from` Goal or empty, lines: key, text, kind |
| `examples.place` | `place`, into git | the Goal | role, H0 commit, H0 subject hash, and for each example the place changed: line key, example name, the example hash. An older `examples.place` holds H0's cell for each example instead of the H0 commit and subject hash: H0 commit, H0 subject hash, Probe hash, output hash, stderr hash, exit code, verdict |
| `example.place` | `place` before G-SBFXAY; no act writes it now | the Goal | role, line key, example name, the example hash, and H0's cell, as `examples.place` holds it for one example |
| `result.record` | `run`, into the trace; in git in an older log | the Goal | role, line key, commit, subject hash, `refused` (empty, or `Stable` with the differing example keys), and per example: example key, example hash, Probe hash, output hash, stderr hash, exit code, verdict |
| `result.retract` | `retract` | the Result | role |
| `select` | `select`, into git | the Goal | role, commit, subject hash, and the examples the row was judged by: each current example key with its hash. An older `select` names no example |
| `budget.reset` | `reset` | the Goal | role |
| `land` | `land` | the Goal | role, the stop it ends (`selected`, `two candidates` or `budget used`), and `cause`: the id of the event of the act that caused that stop |
| `unselect` | `unselect` | the Goal | role |
| `read` | every command that names an actor, in a repository with a store, into the trace, never into the log (§8) | its own id | role (empty when the command declared none of the four), the command's arguments, exit code, stdout hash, stderr hash |
| `cells.save` | `place` and `try`, into the trace only | the Goal | role, commit, subject hash, `again` (false for the cells of a place, true for a cell `try` ran again), and per cell: line key, example name, example hash, Probe hash, output hash, stderr hash, exit code, verdict |

The **Probe hash** is the git tree hash of the example's `probe/` folder. The
example hash already covers it; it is kept on the cell so hyp can show which
examples share one Probe: a Probe is copied, and hyp shows the copies.

The first six events were decided with the design. Owner said, 2026-11-19:
"Why ask me, event list is already decided in report".

`land` is the owner's event that ends a Goal (§8, `land`). It was added by
Goal G-M66KKY, after the six. A stop itself is not stored: it follows from the
stored events (§11), so the log cannot disagree with itself about it. The land
names the stop it ends, so a landed Goal keeps that stop. Owner said,
2026-11-20 (#245): "Stop and land are two different events", and for the stop,
"1 it is": shown as an event, calculated from the log.

`unselect` is the owner's event that puts a selected Goal back to open (§8,
`unselect`). It was added by Goal G-VDG87A (#246), after `land`. It names no
select: a Goal has at most one select at a time, and the unselect undoes the
one that stands at its place in the merge order.

The ninth, `read`, is the trace of what hyp printed to an actor (§8, "The
trace is not the log"). hyp writes it into the clone's trace folder (§1), never
into `.tree/events/`. Its two hashes name output files in the trace's
`outputs/` that hold the bytes hyp wrote to stdout and to stderr. It changes no
Goal, Result or row, and no guard reads it. A `read` line that an earlier hyp
wrote into the log stays there: replay parses it as before, and `show` prints
it with its outputs from `.tree/outputs/`.

`examples.place` names every example it places by its hash, with the H0 commit
and subject hash the set ran on. So one place of the whole set is one line,
and a crash leaves the whole set placed or none of it: a new example runs on
H0 as part of the act that adds it. Built by Goal G-SBFXAY (#312), which
replaced `example.place`, one event per example. H0's cell of each example
goes to the trace as a `cells.save` (Goal G-G4NNT7, #313).

`cells.save` holds cells measured outside a run, and only the trace holds it.
It uses no run of the budget, changes no state and has no guard: a saved cell
whose example hash is no longer current is history, as a Result's cell is.
A log that holds `example.place` events replays them as before: each places
its one example and starts a new budget, so every Goal recorded before keeps
its state, stop, runs and allowed acts. A placed example's H0 cell is live on
H0's row like a Result's cell; it is not a Result and cannot be retracted. It
is live only while its subject hash equals the Goal's H0 subject hash: after a
Draft replace that changes the base it reads `not run`, as every cell of a
changed H0 subject does, and `run <H0 commit> LINE` measures it.

## 4. Merge order, crash and merged log

- **Lock.** Every write act takes the store's directory lock, checks its guards,
  appends one line and releases the lock.
- **Merge order.** Replay reads every events file and orders the events by id,
  then by actor, then by line bytes, so the order of lines in a
  file decides nothing. Two lines with the same actor and the same bytes are one
  event: a union merge can keep a line twice, and the copy is read once.

- **Crash.** An act writes its output files first (§5, Outputs), then its
  event as one `write` of one line and an `fsync`. A crash before the event
  leaves output files no event names; they are harmless. A torn last line does
  not parse: replay skips it, and every state block lists it under "not read
  at replay" with its file and line number. A line whose id is not a ULID or
  whose actor is not the file's actor is listed the same way.
  Before it appends, hyp writes a newline if the file does not end in one, so a
  torn line stays alone and the next event is not glued to it.
- **Trace.** A command that names an actor writes its trace after its act and
  before it prints (§12): the two output files, then the `read` event, in the
  clone's trace folder (§1), under that folder's own directory lock. A write
  act's own event comes first, under the store's lock. A crash after the trace
  and before the print leaves the trace of bytes the actor never got; a crash
  before the trace leaves bytes that were neither traced nor printed. So
  `show` holds every result and refusal the actor got, in order, byte for
  byte, and one more read for each command that crashed after its trace. The
  model of these rules, with both ways `show` can differ from what the actor
  got, is `goals/G-9XX4KB/specs/channel.tla`; where the trace lives, and what
  a command leaves in the repository, is `goals/G-3FDQC0/specs/trace.tla`;
  that the trace is under `~/.hyp` and how `prune` cleans it is
  `goals/G-RK4VYD/specs/home.tla`; that `try` shows the cell `place` and
  `run` record and writes nothing to the log is
  `goals/G-CJWT9T/specs/try.tla`.
- **Outputs of an act.** `place` and `run` hash their Probes' outputs when
  the Probes end, and write the output files into the clone's trace only under
  the store's lock, after every guard held again and just before the event
  that names them. So a refused act leaves no output file, and a crash before
  the event leaves output files no event names, which are harmless. `try`
  (§8) hashes its outputs the same way, and writes them only when it saves a
  missing cell.
- **Place across git and the trace.** `place` appends its `examples.place` to
  git first, then its `cells.save` to the trace. A crash between them leaves
  the set placed with H0's cells not in this clone's trace: `status G` prints
  each such cell as not in this clone's trace, with the `try` line that saves
  it (§8).
- **Ids across git and the trace.** An act checks its guards against the
  events of git and of the clone's trace in one merge order, under the store's
  lock, and its event takes an id after every event it read, in git or in the
  trace. So a `select` in git sorts after the Results it was checked against,
  and a Result in the trace after the place it ran.
- **Merged log.** Two clones that each passed their guards can merge into a log
  that breaks one, for example a fourth Hypothesis in one budget. Replay checks
  every event at its place in the merge order. An event whose guards fail there
  is skipped, listed under "refused at replay" in `status`, and never deleted: a
  guard that a merged log breaks shows when hyp replays it. Output files are named by their hash, so they never conflict in a
  merge. Replay reads git's events and the clone's trace in one merge order,
  so the Results and cells a clone sees are its own: another clone's runs are
  in that clone's trace. A `read` has no guard, so replay keeps every `read` whose payload
  parses, an old one in the log as well as one in the trace. The trace is
  never merged: a clone's agents share one trace folder, and git carries none
  of it.
- **Which guards replay checks again.** Only the guards that read the tables:
  Draft, Whole lines, Not landed, Grill places, Open, Goal's actor, Budget, Stable,
  Selected, Owner lands, Owner unselects, New cell and Own Result, the table
  half of Adds only (each hash
  is new on its key; an `examples.place` names at least one example and each
  key once, and when one of its examples fails, the whole event is refused at
  replay and places none of the set), the table half of Candidate (the row is met on every
  line, and the role is runner or owner; for a `select` that names its
  examples, the examples it names are the current placements, each key at its
  current hash, and the met rows are trusted as written, since the Results are
  in the runner's trace; the leaf mark and the one-or-two count
  read git and are trusted as written), and the table half of Stop (the stop a
  land names is one of the three; `selected` needs a select and the other two
  need none; that the candidates or the budget held reads git and is trusted
  as written). The guards that read disk, git objects, the moving git state
  or a Probe (Whole set, Committed, Kind, Bytes, Leaf only, Spec, Own path, and
  the content half of Adds only) are trusted as written. From landed is not checked again:
  a log written before land existed holds Goals that cite a Goal that is not
  landed, and it replays as it was. Cost: two clones that
  each adjust one example key merge to @2 and @3, and @3 was checked against @1
  only. So when
  two clones merge and one Goal holds a fourth counted Hypothesis, the fourth by
  merge order is refused at replay: its Result counts for nothing, even when it
  was a candidate. Owner said, 2026-11-19: "Refused at replay".
- **An unselect in a merged log.** Replay applies an `unselect` at its place
  in the merge order, with all its guards, which read only the tables. A
  `place`, `run` or `select` that another clone wrote while the Goal was
  still selected there, and that comes before the unselect in merge order, is
  refused at replay by Selected; one that comes after it is checked against
  the open Goal. Two unselects of one select, one from each clone: the second
  by merge order is refused at replay by Selected. A runner's select after
  the unselect in merge order, with no new cell between them, is refused at
  replay by New cell, even when the clone that wrote it had not seen the
  unselect.
- **A write after a land in another clone.** A clone that has not merged a
  land can still place, run or select on that Goal. Once the logs merge, each
  such event after the land in merge order is refused at replay by Not landed,
  and `status G` lists it under "refused at replay".
- These rules follow from rules already decided: the store of ADR 0020 and
  ADR 0022. Owner said, 2026-11-19: "Why ask me, event
  list is already decided in report".

## 5. Calculated tables

hyp stores no table. On every command it reads the log and calculates these
tables in memory. Each guard (§7) is a query over them, or over git.

| Table | Key | Columns |
|---|---|---|
| Goals | Goal id | words, H0 commit, H0 subject hash, budget, actor (of the first `goal.add`; a Draft replace keeps it), `from`, state (§6) |
| Lines | Goal, line key | text, kind (`gap` or `guard`) |
| Placements | Goal, example key | example hash, version `@n` (count of distinct hashes placed on that key, in order), time, actor, the H0 subject hash it was placed on; the latest is current |
| Saved | Goal, `cells.save` id, example key | example hash, commit, subject hash, Probe hash, output hash, stderr hash, exit code, verdict, `again`; from the clone's trace |
| Results | Result id | Goal, line key, commit, subject hash, actor, role, time, retracted, refused by Stable |
| Cells | Result, example hash | example key, Probe hash, output hash, stderr hash, exit code, verdict (`met` or `failed`) |
| Rows | Goal, row key | the commits it was seen at; leaf (from git); per example key, the verdict of its live cell under the current placement, or `not run` |
| Selects | Goal | select id, commit, subject hash, time, actor; the current select only: an unselect removes it |
| Unselects | Goal, unselect id | the select it undid, time, actor, and what had run before it: the Results recorded up to it and the example versions (key and hash) placed up to it |
| Lands | Goal | land id, the stop it ends, its cause, time, actor |
| Reads | actor, read id | time, role, command, exit code, stdout hash, stderr hash; from the clone's trace and from the old `read` lines of the log; an actor's reads are in id order, oldest first |

- A cell is **live** when its Result is not retracted, not refused by Stable,
  and its example hash is the current placement's hash for its example key. A
  saved cell is live under the same rule on its row; on one row the place's
  own H0 cell comes first, then the Results' cells, then the cells `try` ran
  again.
- **Outputs.** An output is a file named by its sha256 in the clone's trace,
  `outputs/<sha256>`; an older log keeps its outputs under
  `.tree/outputs/<sha256>`, and hyp reads both. The cell holds the hash. Equal outputs are stored
  once, and "same bytes" is a compare of two hashes. hyp writes the file before
  the event that names it. Owner said, 2026-11-19: "File by sha256".
- **Cell key.** A cell is keyed by its Result and the example's hash, not by
  the example key and version. An adjust makes new cells; the old cells stay as
  history and are not live under the new placement. Owner said, 2026-11-19:
  "Example hash".
- **Row key.** A row is keyed by its Goal and its subject hash (§1, §13). A
  row records one **Hypothesis**: one build the runner offers, the subject of a
  commit. Commits with the same subject hash are one Hypothesis, so the commits
  a row was seen at are a column: a leaf comes from commit ancestry, and one
  subject can sit at several commits. **H0** is the Hypothesis of the Goal's
  base. H1, H2, H3 are the others, in the order they first ran, and the budget
  counts them (§11).
- **Hypothesis tree.** The Hypotheses of one Goal, arranged by the git ancestry
  of their commits, with H0 as its root: a Hypothesis is below another when one
  of its commits descends from one of the other's. hyp reads the arrangement
  from git (§13) and stores no edge. A row is a **leaf**, a Hypothesis with
  nothing below it, when none of its commits has a descendant commit among the
  commits of the Goal's other rows. Owner said, 2026-11-19: "Tree; leaf = no
  commit has a descendant". Owner decided, 2026-11-20 (#253): a Hypothesis is
  one build, keyed by its subject hash, and "tree" alone names only the
  Hypothesis tree.

## 6. States

Each state is calculated from the tables; none is stored. `landed` wins over
`selected`, and `selected` wins over `open`: a selected Goal is not open.

### Goal: `written → open → selected → landed`

A Goal with no select lands from `open` at the stops Two candidates and Budget
used (§11), so `open → landed` is a path too. The owner's `unselect` puts a
selected Goal back to `open` (§8), so `selected → open` is a path too, and a
Goal can be selected again after it.

| State | Holds when | Refuses |
|---|---|---|
| `written` | A `goal.add` exists, and the Goal is not open. | `run` on a Hypothesis other than H0 (**Open**), `select` (**Candidate**: no row is met yet), `land` (**Stop**: a written Goal has no stop), `unselect` (**Selected**: no select) |
| `open` | Every line has at least one placed example, and every current placement has a live cell on H0's row with the verdict its line's kind needs: `failed` on a gap line, `met` on a guard line. | `add goal G` replace (**Draft**), `land` while the Goal has no stop (**Stop**), `unselect` (**Selected**: no select), a runner's `select` after an unselect with no new cell since (**New cell**) |
| `selected` | The Goal has a select and no land. | `add goal G` replace (**Draft**); `place`, `run` and a second `select` (**Selected**: the Goal already has a select); `add goal --from G` (**From landed**) |
| `landed` | The Goal has a land. | every write: `add goal G` replace (**Draft**), `place`, `run`, `select`, `unselect` and a second `land` (**Not landed**) |

All refusals are exit 1. A Draft replace that changes the lines leaves existing
placements on their keys; a placement whose line key is gone is history, and the
Goal is open again only when the rule above holds.

`selected` refuses `place`, a second `select` and `run` by a
guard named for the state, **Selected**, so that no refusal reads "Open" by
accident: a run after the select counts nothing. The
owner's `land` and `unselect` are the only writes it allows.
Owner accepted, 2026-11-19. The re-grill of a selected Goal goes through
`unselect`: it puts the Goal back to `open`, and then the grill places and
the runner runs as on any open Goal. Owner said, 2026-11-20 (#245): "that's
why #246 is needed (unselect)". Built by Goal G-VDG87A (#246).

`landed` ends the Goal. Only the owner lands it, and only at a stop: the Goal
is done when the owner says so, since a re-grill can still add an example
inside the Goal. Owner said, 2026-11-20 (#245): "That's why owner is a one who
decide the `land`". The land is not coupled to a merge: a Goal can land before
the merge of its change, after it, or with no merge, and a merge never lands a
Goal. Owner said: "I find coupling `land` to the merge action is the bad idea".

### Example: placed `@1 → @2 → …`

An example key is placed by the first place whose event names it. Each later
place that names a different hash on the same key places the next version. A
place names only the examples it changes; a place that changes no example is
refused (Adds only: no new version). A placed example stays in the set: its
folder may gain floors and ceilings, and is never removed (Whole set). A version changes only floors and ceilings; a changed input, run,
`external` section or Probe is a new example key (§7). The old versions are history; their cells are no longer live.

### Result: `recorded → retracted`

`retract R` moves a Result to retracted. Its cells stop being live, so their
examples read `not run` on that row again. A retracted Result is never deleted.

### Row

A row has no stored state. H0's row always exists. Any other row exists when a
Result names its subject hash. Its leaf mark and its verdicts are calculated (§5). When
one of a row's commits is not in this clone, reads print `leaf unknown` for it,
and `run` and `select`, which need the mark, exit 3 naming the commit.

## 7. Guards

An act is allowed when every guard it names holds. A refusal is exit 1 and names
the first guard that fails, in the act's order (§8).

- **Draft** — Before the Goal opens, `add goal G` may replace it. After, its words
  and lines are frozen. Query: `Goals[G].state == written`. A new Goal (no `G`)
  always passes.
- **Whole lines** — Every line has a kind, `gap` or `guard`, and the Goal has at
  least one gap line. Query: over the act's own input.
- **Not landed** — No write after the Goal is landed: the owner's `land` ends it.
  Checked first by `place`, `run`, `select`, `land` and `unselect`, and by
  replay. Query:
  `Lands[G]` does not exist.
- **Open** — The Goal is open. Checked by `run` unless the commit's subject
  hash equals H0's. Query: `Goals[G].state == open`.
- **Goal's actor** — Only the actor that added the Goal selects for the owner,
  lands or unselects the Goal, or resets the budget. Query: the act's actor `== Goals[G].actor`.
- **Grill places** — Only the grill places an example; the runner never places.
  Query: the act's role `== grill`.
- **Kind** — H0 gives the verdict the line's kind needs under the example's
  Probe: a gap example fails on H0, a guard example is met on H0. `place`
  checks it on every example the place changes. Query: run the Probe on H0's
  commit (§10), compare (§9), read the verdict.
- **Whole set** — The example set of a place is every folder under
  `goals/<G>/<line>/`, on disk or committed at HEAD, for each line of the
  Goal; `goals/<G>/specs/` and every other folder are not in it. Every line
  holds at least one example folder, every example key placed before still
  has its folder, and every folder's name is an example name
  (`^[a-z0-9][a-z0-9-]{0,39}$`). Query: read the line folders on disk and at
  HEAD, and the Placements table. Built by Goal G-SBFXAY (#312).
- **Committed** — Each example folder's bytes on disk equal the committed bytes.
  Query: the folder's example hash computed from disk (§13) equals
  `git rev-parse HEAD:goals/<G>/<line>/<example>`, and the committed folder
  holds an `example.txtar` that parses by §9; the refusal names the file and
  line.
- **Adds only** — A new version of an example only adds floor or ceiling
  sections. Everything else is byte-equal to the current version: the `words`,
  `run`, every `input/…` and the `external` section, and every other entry of
  the folder (`probe/` and any file beside it) by mode and hash.
  Any other change is a new example key, not a new version. Query: read the
  current placement's folder by its hash (§13); every root entry but
  `example.txtar` has an equal mode and hash; the non-floor, non-ceiling sections are equal by name and bytes; every
  floor and ceiling of the current version (kind and literal) is in the new
  file; and the new file holds at least one more. This follows the owner's
  "re grilling will have two choice, adjust (floor and ceiling) and create new
  Goal if need new line", and narrows it. The design let a re-grill change a
  Probe or add an input under the same key. v2 makes each of those a new
  example key, so the verdicts of the old check never mix with the verdicts of
  the new one. Owner decided, 2026-11-20 (#253): keep this narrower rule. The
  refusal names the first section or
  path that differs. A hash the key held before, current or earlier, is refused:
  an earlier version never comes back, also not from a merged log, which lists
  it under "refused at replay". Owner said, 2026-11-19: "Refused". When the
  current version's object is not in this clone, the act is exit 3.
- **Leaf only** — A run is on H0, or on a commit that no row of the Goal descends
  from. Query: for each commit `c` of the Goal's rows other than COMMIT,
  `git merge-base --is-ancestor COMMIT c` fails.
- **Nothing new** — Removed. It refused a second run of an example on the same
  Hypothesis under the same example hash. Stable now judges every repeat instead.
  Owner said, 2026-11-19: "Drop Nothing new". Leaf only alone keeps runs on H0
  and leaves.
- **Bytes** — The commit exists. Query: `git cat-file -e COMMIT^{commit}`.
- **Stable** — Two runs on the same bytes give one verdict. Checked by `run`
  after the Probes run: for each example run, no live cell on the same subject
  hash under the same example hash has another verdict. A Stable refusal is exit 1,
  and its cells never become live, so the first verdict stands. It still keeps
  what it saw: the new output files stay, and `run` appends its `result.record`
  marked `refused: "Stable"` with the examples whose verdict differed. That
  Result counts nothing: no live cell, no budget, no retract. `status` lists it
  under `unstable:` with both outputs' hashes, so the owner sees the unstable
  example. The report prints both outputs. A
  retracted cell is not live, so a retract followed by a rerun is never refused
  by Stable. A repeat on a Hypothesis already counted costs no budget (§11). Owner
  said, 2026-11-19: "Drop Nothing new" (C5: keep Stable alone).
- **Budget** — No fourth Hypothesis since the boundary, H0 never counted.
  Query: the distinct subject hashes, other than H0's, of the Goal's Results
  recorded after the boundary (§11), without Results refused by Stable; refuse
  when there are three and the commit's subject hash is not among them. A
  Hypothesis run before the boundary and run again after it counts: it is run
  against the new evidence. A run on a subject hash that was H0's when it ran
  never counts, even after a Draft replace moved
  the base. Owner said,
  2026-11-19: "It counts".
  A retracted Result still counts: a retract fixes a wrong
  Result, and does not give a run back: a runner that hides failed
  paths is what the budget cannot see.
- **Candidate** — The Hypothesis is a candidate: a leaf row met on every line under the
  current placements. With exactly one candidate the runner may select it; with
  two or more, only the owner. Query: Rows of G with leaf = true and every
  current example key `met`.
- **New cell** — After an unselect, the runner selects only once a run since
  the latest unselect recorded a new cell: a live cell whose example version,
  or whose row, had not run before that unselect. A version is an example key
  at one hash; one placed before the unselect ran on H0 when it was placed.
  A row ran before when a Result before the unselect, not refused by Stable,
  named its subject hash; H0's row always ran before. So a rerun of the same version
  on the old row is no new cell, and neither is an old version run on an old
  row that never had it. A run on a new commit whose subject hash is new, or a
  run of a version placed after the unselect, is one. The owner selects at
  any time. A Goal never unselected always passes. Query: over the Unselects,
  Placements and Results tables, for the runner's role only. `status` prints
  the runner's select line only while it holds (§12). Owner said in the Goal's
  line 4: "the runner selects only once a run records a cell whose version,
  or whose row, was not run before it".
- **Spec** — The selected commit holds a spec under `goals/<G>/specs/`. When
  that spec is a model, every rule its check configs check is named by a
  variant config, and the folder holds a chart. Query: `git ls-tree -r COMMIT
  goals/<G>/specs/` lists at least one file. When one of them ends in `.tla`,
  the folder is a model: every `.cfg` under `specs/variants/` is a variant
  config, and so is every `.cfg` under `specs/switches/`, the folder that held
  the variant configs of a spec written before the rename; every other `.cfg`
  is a check config, and the rules of a config are the names it lists under
  `INVARIANT`, `INVARIANTS`, `PROPERTY` or `PROPERTIES` (`git cat-file blob`).
  A rule of a check config that no variant config lists is refused
  (`rule-no-variant`), and the refusal names every such rule and the folder
  `goals/<G>/specs/variants/`; then a model with no `.mmd` file outside the
  variant config folders is refused, and the refusal names the missing chart.
  hyp runs no model checker, reads no model report and records no model
  result: the statechart tool writes the variant configs, and whether a variant
  really breaks the rules it names is the tool's report, not a guard. A spec
  folder with no `.tla` file is checked for one file only, as before. Owner
  said, 2026-11-20 (#247): "Grill models, option 2"; "A model but no chart:
  refuse the select. since we have grill doing it".
- **Own path** — A commit writes no path keyed by another Goal. Only the
  runner's own change of the commit is read (§13), never the diff from H0: when
  the runner takes in a main that merged another Goal's record branch, the diff
  from H0 holds `goals/<X>/` that the runner did not write. Query: the paths of
  the runner's own change of COMMIT hold none under `goals/<X>/` with `X` ≠ G.
  The refusal names the path and the commit that wrote it.
- **Selected** — The Goal has a select. Query: `Selects[G]` exists. `place`,
  `run` and `select` need it not to hold (§6); replay also checks it on
  `result.record`. `unselect` needs it to hold: an open Goal, never selected
  or already unselected, has no select to undo.
- **Owner lands** — Only the owner lands a Goal. Query: the act's role
  `== owner`.
- **Owner unselects** — Only the owner unselects a Goal. Query: the act's
  role `== owner`.
- **Stop** — The Goal is at a stop (§11): Selected, Two candidates or Budget
  used. Query: the stop of §11, calculated from the tables and git. A land on a
  Goal with no stop is refused, and checks no commit: land is not a recheck of
  a commit on a main that moved. That recheck is `unselect` (#246), then a run
  and a select again. Owner said, 2026-11-20 (#245), decision 9: "the problem
  is by design, it affect only H0, right? so why this is complicated?"
- **From landed** — `add goal --from G` cites only a landed Goal. Query:
  `Goals[G].state == landed`. Checked by the act only, never at replay (§4).
  Owner said, 2026-11-20 (#245): "The --from should be the Goal that landed
  already".
- **Own Result** — Only the actor that ran a Result retracts it. Query: the act's
  actor `== Results[R].actor`.
  Owner accepted, 2026-11-19.

## 8. Acts

The act comes first, then its keys. Every write act takes the lock, checks its
guards in the order listed, appends one event, and prints the return (§12).
Common to every act: exit 2 for bad arguments; exit 3 for an environment problem
(§12).

| Act | Role | Guards, in order | Writes |
|---|---|---|---|
| `add goal [G] [--from G] --base H --words TEXT --line KIND:TEXT …` | owner | Draft, Whole lines, From landed (with `--from`) | `goal.add` |
| `place G` | grill | Not landed, not Selected, Grill places, then for the whole set: Whole set, Committed, Adds only, Kind | `examples.place`, in git; `cells.save` of H0's cells, in the trace |
| `run COMMIT LINE` | runner | Not landed, not Selected, Bytes, Open (not for H0), Own path, Leaf only, Budget, Stable | `result.record`, in the trace |
| `select G HASH` | runner, or owner | Not landed, not Selected, Candidate, New cell (runner only), Goal's actor (owner only), Spec, Own path | `select` |
| `land G` | owner | Not landed, Owner lands, Goal's actor, Stop | `land` |
| `unselect G` | owner | Not landed, Owner unselects, Goal's actor, Selected | `unselect` |
| `reset G` | owner | Not landed, Goal's actor | `budget.reset` |
| `retract R` | the actor that ran R | Own Result | `result.retract` |
| `status [G] [--role ROLE]` | any | none | nothing in the repository; a `read` in the trace |
| `try G LINE/EXAMPLE [--worktree \| --commit COMMIT] [--role ROLE]` | any | none | nothing in the repository; a `read` in the trace, and a `cells.save` when it saves a missing cell |
| `show ACTOR` | any | none | nothing in the repository; a `read` in the trace |
| `prune [--remove]` | any | none | nothing in the repository; a `read` in the trace; with `--remove`, removes the trace folders it lists |

- In `select G HASH`, HASH is the selected row's subject hash, or a prefix of
  it, as hyp prints it. The select records the commit of the row's latest
  Result in merge order.
  Spec and Own path read that commit, and the PR holds it. Every commit of a leaf
  row has no descendant among the Goal's rows (§5), so the commit is a leaf. A
  spec committed after the last run is not in that commit: the runner runs one
  line on the spec commit first. Its subject is the same Hypothesis, so it costs no
  budget, and Stable checks the repeat.
- `place G` places the Goal's whole example set (Whole set, §7) in one act.
  It checks Whole set, Committed and Adds only on every example folder before
  it runs any Probe, so a refusal that needs no Probe costs no Probe run. Then
  it runs each example that the place changes, a new example key or a new
  version, on H0 for Kind. It places the set only when every guard holds on
  every example: one `examples.place` event names every changed example with
  its hash, and the trace saves each one's H0 cell. Otherwise it places none of the set, exit 1, and its refusal
  names every cause it found, each under its example key (or its line key for
  a line that holds no example) with its guard: "none of the example set of G
  is placed; N causes:". The causes that need no Probe are reported without
  running a Probe; Kind causes are reported once every Probe ran. A place
  that changes no example is refused by Adds only ("no example of G changed
  since its last place"), and so it restarts no budget; a change under
  `goals/<G>/specs/` alone is no change. Its report is `placed G: N examples`,
  then one line per changed example with its version, hash and H0's verdict
  and exit code, then the Goal's state. `place G LINE/EXAMPLE`, the old form,
  is a bad argument, exit 2. An example placed before keeps its H0 cell: a
  place does not run it again. Built by Goal G-SBFXAY (#312).
- `run` checks every guard that needs no Probe before it runs a Probe; Stable is
  checked on the outputs.
- `land G` ends a Goal at its stop. It records one `land` event that names
  the stop and the event of the act that caused it, whether or not the change
  was merged. Its report is `landed G` and the stop as an event (§11), then the
  state block of the landed Goal, which prints the record of the decision
  (§12). `status` offers it to the owner at every stop, as `hyp land G`. Built
  by Goal G-M66KKY. The model of its rules, with the cheaper builds each rule
  refuses, is `goals/G-M66KKY/specs/land.tla`.
- `unselect G` puts a selected Goal back to `open`. It records one
  `unselect` event; its report is `unselected G` and the select it undid
  (its id, time, commit and row), then the state block of the open Goal.
  `status` offers it to the owner on every selected Goal, as
  `hyp unselect G`, beside `hyp land G`. What it changes, and what it keeps:
  - The Goal is `open` again: the grill may place, the runner may run.
  - Every Result and cell stays as it was, on every row. The old row keeps
    its live cells, so it is still a candidate while it is a leaf, and no new
    run on H0 is needed.
  - The budget's boundary moves to the unselect (§11): `runs: 0 of 3 used`.
  - The owner may select a candidate at once. The runner's select needs a
    new cell since the unselect (New cell, §7), and `status` prints the
    runner no select line until then.
  - A Goal never unselected selects as before.
  It checks no commit, as `land` checks none. The recheck of a selected
  commit on a main that moved is `unselect`, then a run on the merge commit,
  whose subject hash is new, then a select. Built by Goal G-VDG87A (#246). The
  model of its rules, with the cheaper builds each rule refuses, is
  `goals/G-VDG87A/specs/unselect.tla`.
- **The snapshot** of a landed Goal is not a separate thing: it is the Goal's
  events in the `.tree/` of the commit or checkout being read, up to its land,
  and the outputs they name. hyp stores nothing more for it, and it holds no
  code of a Hypothesis, only Results. So each commit shows what it holds: hyp
  reads a Goal as landed in a checkout only once it holds the Goal's stop and
  its land, as git shows
  any branch before a merge. Owner said, 2026-11-20 (#245): "Result only then",
  and "If you want to replay, why not read the tree itself?" It is the owner's
  decision, fixed: what the owner accepted (the words, lines, examples, the
  selected row, the verdicts and outputs), or with no select what failed.
  Later Goals that cite it with `--from` start from it. It is read:

  | When | Who reads | What, and why |
  |---|---|---|
  | Every hyp command | hyp, in replay | the land event, to calculate `landed` and refuse every write to the Goal (Not landed) |
  | The return of `land`, and `status G` | any role | the record of the decision: the PR body, or the table of outputs with no select (§12) |
  | `status` for every Goal | any role | only the state block of a landed Goal, as of a selected one (§12) |
  | `add goal --from G` | hyp | whether G is landed (From landed) |

- `reset` and `retract` are specified here and built as later v2 Goals.

**The trace is not the log.** `status` and `show ACTOR` are reads: they
check no guard and are always allowed. Every command that names a valid actor,
a read or a write act, records what hyp printed to that actor as one `read`
event (§3), before it prints (§12), in the clone's trace folder (§1). So
`hyp show <actor>` prints exactly what that actor got, however the log changed
after. Owner said, 2026-11-19 (record PR #231):

1. "Separate the trace from the log. The log holds only the acts that change a
   Goal. What hyp printed to an actor is the trace, and `hyp show` keeps
   reading it."
2. "Every command. No command leaves a change in tracked files beyond its own
   act's event and the outputs that event names. A read, a refusal (exit 1)
   and a bad argument (exit 2) leave nothing."
3. "Any worktree of one clone. The owner can read an agent's reads from any git
   worktree of the clone the agent used. Reads in a separate clone are not
   covered."

This reverses the part of "Reads leave a trace" (record PR #226) that put the
trace in the log; the rest of it, that every command's output is traced
before it prints and `show` prints the trace, stands.

- **What a command leaves in the repository.** A read, a refusal and a bad
  argument leave no file of the repository changed, added or removed. A
  recorded act in git leaves its own event, and no output file: `run` leaves
  nothing in the repository, and `place` leaves one line.
  The first `add goal` also creates the store (§8, no `init`). A run that
  Stable refuses is exit 1 and still records its `result.record` marked
  refused, with the outputs it names (§7): it is the one refusal that is also a
  recorded act. hyp never hides a file from git to meet this rule: it writes
  no `.gitignore`, no `info/exclude` and no index flag.
**Git keeps the acts; the trace keeps the evidence.** Git keeps only the
owner's acts, the place of each example set and the select. Every cell,
Result and output stays in the clone's trace, as a cache that a reader fills
by running an example again. This is Goal G-G4NNT7 (#313):

- `place G` writes to `.tree/` one line, its `examples.place`, which names the
  example set's hashes and the H0 it ran on. H0's cells and outputs go to the
  trace as a `cells.save`.
- `run` writes nothing under `.tree/`: its `result.record` and outputs go to
  the trace. `select` writes one line that names the commit, its subject hash
  and the hash of each example it was judged by.
- **State from git.** A place in git checked Kind on the H0 subject hash it
  names, so while that is the Goal's H0 the placement gives the verdict its
  line needs in every clone, and the Goal opens from git alone. A select that
  names its examples selects the Goal in every clone whose current placements
  are those examples. A placement on another H0, after a Draft replace, needs
  a live cell on H0's row, as before.
- **Budget per clone.** The budget counts the Hypotheses in the clone's own
  trace (and the Results an older log keeps in git). A fresh clone of an open
  Goal prints `runs: 0 of 3 used`, whatever other clones ran.
- **A missing cell.** In a clone whose trace lacks a cell that `status G`
  would print, it prints that the cell is not in this clone's trace and the
  ready `try` line that runs it again: `hyp try G LINE/EXAMPLE` for H0's cell
  of an open Goal (under `The Goal asks:`, as `verdict on H0 at place: not in
  this clone's trace; run it again with: …`), and `hyp try G LINE/EXAMPLE
  --commit C` with the selected commit for a cell of the PR body (as
  `verdict: not in this clone's trace; run it again with: …`).
- **Run again.** A `try` of a missing cell saves that cell in the clone's
  trace as a `cells.save` with `again`, and `status G` then prints it marked
  `run again in this clone`. Other clones show no mark: the mark is in this
  clone's trace only.
- **The selected commit on no ref.** Every state block of a Goal with a
  select (`status`, `status G`, and the return of `select` and `land`) prints
  `warning: the selected commit C is on no branch or tag of this clone; …`
  when no branch, remote branch or tag holds C (`git for-each-ref --count=1
  --format=%(refname) --contains C refs/heads refs/remotes refs/tags`), also
  when another commit with the same subject hash is on one, as a squash merge
  leaves. A commit the clone lacks is on no ref.
- **An older log stays as it was.** A landed Goal whose place, run, select and
  land are lines of git prints its PR body with each verdict and output, as
  before: replay reads its cells from git and its outputs from
  `.tree/outputs/`.

- **Where the trace lives** is outside every worktree and the git directory
  (§1), so it needs no git state (§13). Its key is the clone's git common
  directory, so every worktree of one clone reads one trace, and `show`
  prints a read in another worktree before any commit or merge. A separate
  clone, or the same clone moved to another path, has a trace of its own. A
  new repository at a deleted one's path has the deleted one's key, and shows
  its reads. Owner accepted, 2026-11-19 (record PR #238, decision 2).
- A read needs an actor, from `--actor` or `$HYP_ACTOR`. With none, or one that
  does not match §2, it is exit 2 and leaves no trace: a read with no actor
  would print bytes no `show` can find. The role is not needed for a read. A
  role from `--role` must be one of the four (exit 2), and filters `status`
  (§12). The trace records the declared role, from `--role` or `$HYP_ROLE`,
  when it is one of the four, else none.
- A command records its trace once it holds a valid actor, whatever its exit
  code: a refusal, or a bad argument found after the actor, is traced with its
  exit code and its stderr bytes.
- No trace is written outside a git repository or in a repository with no
  store: a trace never creates the store. The `add goal` that creates the store
  is traced. With no absolute `$HOME`, the trace has no folder: the command
  prints only the notice of §12 and exits 3.
- **`show ACTOR`** prints every read of ACTOR in id order, oldest first: from
  the trace, and from the old `read` lines of the log. For
  each, its id, time, role, exit code and command, then the bytes hyp wrote to
  stderr and to stdout, each as a byte block (§12) of the stored bytes. It never
  prints a fresh read. An actor with no read prints `no read of ACTOR is
  recorded`, exit 0. ACTOR must match §2 (exit 2). `show` is itself a read of
  the actor that runs it, never of the actor it shows.
- **`prune`** lists the trace folders under `$HOME/.hyp/traces/` whose
  clone no longer exists: the path that the folder's file `clone` names is not
  on the disk. A deleted clone and a clone moved to another path are both
  gone. For each it prints the folder's key and that path, and it removes
  nothing. A live clone's trace is never listed or removed. Owner said,
  2026-11-19 (record PR #238): "`hyp prune` lists, `hyp prune --remove`
  removes"; the owner sees the list before anything is removed.
  - `prune --remove` removes the folders of those clones and prints what it
    removed. Before it removes a folder it takes the folder's lock and checks
    its clone again, so a clone that came back keeps its trace.
  - A folder whose file `clone` is missing, empty or not an absolute path is
    never listed: hyp cannot tell whose it is. A clone path hyp cannot check
    for another cause than that it does not exist counts as live.
  - With nothing gone it prints `no trace under DIR is of a repository that no
    longer exists`, exit 0.
  - `prune` never reads or writes `~/.local` (§1).
  - Like a read, it needs an actor (exit 2), records its own `read` in the
    trace of the clone it runs in, which is live, and prints no state block.
    It takes no argument but `--remove` (exit 2).
  - It keeps the store rule below: with no store it is exit 1, "no store; run
    hyp add goal". It needs the store only as a place to run from; what it
    prunes is read from `$HOME/.hyp`. Owner said, 2026-11-19: "for
    consistency". Outside a git repository it is exit 3, as every command.
- **`try G LINE/EXAMPLE`** shows an example's output, stderr, exit code and
  verdict as `place` and `run` would record them, and records nothing in git.
  It saves one thing in the clone's trace: the cell it measured, when the
  folder on disk is the current placement, the subject's hash is H0's or the
  selected row's, and the trace holds no live cell of that example there
  (Goal G-G4NNT7, "Run again"). It is
  line 1 of Goal G-CJWT9T: "An agent can see an example's output, stderr,
  exit code and verdict, as hyp would record them, on H0 or on an uncommitted
  change, without a commit, without recording anything in the log, and
  without using a Hypothesis." Owner said, 2026-11-19 (record PR #241,
  decision 3): "The command is `hyp try G LINE/EXAMPLE`. It runs the example
  folder as it is on disk, on H0. With `--worktree` it runs on the files of
  this checkout instead. It is one new read; `place` and `run` get no
  `--dry-run`."
  - The example is the folder `goals/<G>/<LINE>/<EXAMPLE>` as it is on disk,
    committed or not: its `example.txtar` and its `probe/` are read from
    disk, never from git, so Committed is not checked.
  - Without `--worktree` the subject is H0's commit, extracted by hash as
    `place` extracts it (§10): an edit in the checkout does not reach it.
    With `--commit C` it is the commit C of this clone, extracted the same
    way; a C this clone lacks is exit 2, and `--commit` with `--worktree` is
    exit 2.
    With `--worktree` the subject is the files of this checkout as a commit
    of every change would hold them: each file git tracks or would add
    (`git ls-files -z --cached --others --exclude-standard`, §13), with its
    bytes and execute bit on disk, without `goals` and `.tree`. A tracked
    file deleted on disk is left out, and a file git ignores is not read. A
    folder in that list, a submodule or a nested repository, is refused,
    exit 1, as an extracted submodule is (§13).
  - The Probe runs as §10 says, and the compare is §9's. hyp hashes the
    output and the stderr as `place` and `run` do, and writes the output files
    into the trace only when it saves the cell.
  - It prints where it ran (H0's commit, the commit named, or HEAD's commit
    and every change not committed), the line `nothing is recorded: no event,
    no output file, no run of the budget`, or when it saved a missing cell
    `saved in this clone's trace: the cell of K on row R, which the trace
    lacked, marked as run again in this clone; nothing in git, no run of the
    budget`, the folder's example hash on disk and its Probe hash,
    the example's words, `verdict: V, exit N`, and the output and the stderr
    as byte blocks (§12) labelled with the sha256 each output file has. Then the Goal's state block, which `--role` filters as
    in `status`.
  - It is a read: it checks no guard, so it runs in every state of the Goal,
    `selected` included, and no state block lists it as an allowed act (§12
    lists the acts that write). It needs an actor (exit 2) and leaves its
    `read` in the trace, so `show` holds what it printed.
  - Exit 2 for a key that is not `LINE/EXAMPLE`, an unknown Goal or line, a
    folder that is not on disk, or one with no `example.txtar`; exit 1 for an
    `example.txtar` that does not parse (§9), the refusal `place` prints; exit
    3 for a Probe that fails, times out or prints no `exit N` line (§10).
  - What it shows equals what `place` records for H0's cell once the folder
    is committed as it was, and what `run` records once the checkout's change
    is committed as it was, because both run one Probe on the same bytes. A
    Probe whose output changes from run to run shows a different output;
    Stable (§7) judges that only in `run`.
- `show KEY` for a Goal or a Result, and `prompt ROLE`, are later v2 Goals.

No `init` act: `add goal` creates `.tree/` at
`git rev-parse --show-toplevel` when it is absent. Every other act, and every
read, with no store
is exit 1 naming "no store; run hyp add goal". Outside a git repository is exit 3.
Owner accepted, 2026-11-19.

The actor comes from `--actor` or `$HYP_ACTOR`, the role from
`--role` or `$HYP_ROLE`. Each event stores its role in the payload. A write act
missing either, or naming a role outside the four, is exit 2. Only Grill places and
Candidate read the role; elsewhere it is recorded, not enforced.
Owner accepted, 2026-11-19.

Own path is checked by `run` and `select`: they refuse a commit whose own
change (§13) touches `goals/<other Goal>/`.

`run` names its Goal in LINE, written `<Goal>/<line>`, for
example `hyp run 4df2b9e G-NXHHVA/line-3`, so the act stays first and then two
keys.
Owner accepted, 2026-11-19.

`add goal` takes the Goal's words as `--words TEXT` (required; owner said,
2026-11-19: "Accept --words TEXT"), and repeated
`--line gap:TEXT` and `--line guard:TEXT`, in order; hyp gives the keys `line-1`, `line-2` …. A Draft
replace is `hyp add goal G …`. `--from G` needs G to be landed (From landed,
§7), stores only the cited Goal's full `goal.add` id, which no other Goal
shares, and copies nothing. Whole lines also refuses more than ten lines.
Owner accepted, 2026-11-19.

A line with no placed example has nothing to run: `run` refuses it by Open,
even on H0. `run` runs every placed example of the line on the commit, whether or not a
cell already exists on that Hypothesis; a repeat is judged by Stable. Owner said,
2026-11-19: "'live' is depend on the version of probe. So when new probe
created, Why not [j]ust run H0 and leaf without caring about the live ?"

## 9. The example file and the compare

An example is one txtar file, `example.txtar`. Each section starts with a line
`-- name --`. The sections:

| Section | Holds | Count |
|---|---|---|
| `words` | The reading of the line the owner chose for this example: the part or meaning of the line it checks, in plain words. hyp prints it. A new reading is a new example key. | exactly one |
| `input/<path>` | A file the run uses, written at `<path>`. | zero or more |
| `run` | The command the Probe runs on the subject. | exactly one |
| `floor: contains` | A literal the output must contain. | one or more floors in total |
| `floor: equals` | A literal the output must equal. | |
| `ceiling` | A literal the output must never contain. | zero or more |
| `external` | What the Probe reads outside itself and the subject: one environment variable name per line. | zero or one |

- One literal per section, byte for byte, with no escaping. A section body's
  final newline does not count.
- Floors are literal only. hyp has no pattern language. A filter, such as a
  pattern list, is an input file.
- **The compare.** An example is **met** when the output holds every floor (it
  contains a `contains` floor; it equals an `equals` floor) and contains no
  ceiling. Otherwise it is **failed**.
- A **line is met** on a row when every placed example of the line is met there.
- **The output** the compare reads is the stdout of `run`, with one final
  newline dropped, the same rule as a section body: `floor: equals` `abc` meets
  the stdout `abc\n`. Owner said, 2026-11-19: "Drop one, like a section". Its exit code and
  its stderr are recorded and shown beside it, and never compared. Owner said,
  2026-11-19: "stdout only" for the compare; stderr "Kept beside, not compared".
  The decided example of line 1 (an empty `floor: equals` over a `grep`
  pipeline that exits 1 when nothing matches) works as written.

## 10. The Probe contract

- An example's Probe is its folder's `probe/`. The Probe runs `run` on the
  subject and prints what `run` printed. It gives no verdict.
- A Probe is self-contained. It reads nothing outside itself except the subject
  and what its example's `external` section names. Owner said, 2026-11-19:
  "-- external -- section".
- A Probe failure that is not the run's own failure records nothing (exit 3).
- hyp does not sandbox a Probe. It is the grill's own committed code, run at the
  grill's request, with `PATH` and `HOME`; it can write outside its work
  directories. hyp's own writes stay inside the store, its temp directories
  and the trace folders under `$HOME/.hyp` (§1). hyp writes nothing under
  `~/.local`.

**How hyp runs a Probe.**

- hyp runs `probe/run`, an executable file of the example folder, with empty
  stdin and a timeout of 10 minutes.
- Its **working directory** is a fresh copy of the commit's subject, extracted
  by hash (§13); for `try --worktree`, a copy of the checkout's files (§8,
  `try`). `HYP_SUBJECT` names it too.
- `HYP_EXAMPLE` names a fresh working copy of the example folder, extracted by
  its example hash; for `try`, copied from the folder on disk. Each `input/<path>` section is written there at
  `input/<path>`. Inputs belong to the example; hyp never writes them into the
  subject. Owner said, 2026-11-19: "if input is example scope, they should
  belong to example folder", and "Start in the subject".
- `HYP_RUN` holds the `run` section's text.
- Of the parent environment only `PATH`, `HOME` and the names in the example's
  `external` section pass.
- The Probe prints the stdout of `run`, then a last line `exit N`, N being the
  exit code of `run`. When that stdout is not empty and does not end in a
  newline, the Probe prints one before `exit N`. hyp splits the last line off
  and drops one final newline from the rest: that is the output (§9).
  The Probe's own stderr is the stderr kept beside the output.
- A Probe that exits non-zero itself, times out, or whose stdout does not end
  with a line `exit N`, records nothing (exit 3).

So the decided example of line 1 reads its pattern list from the example:

```
-- run --
grep -rlE -f "$HYP_EXAMPLE/input/patterns.txt" --include=*.go . | grep -v _test.go
```

Committed compares the folder's example hash from disk with
`git rev-parse HEAD:goals/<G>/<line>/<example>`. hyp extracts the example for a
run by its example hash (`ls-tree -r -z` and `cat-file --batch`), so a run does
not depend on the branch checked out: hyp places only a folder whose on-disk
hash equals its committed hash.

## 11. Budget and stops

- **Budget**: at most three Hypotheses run, H0 never counted, since the
  boundary. The unit is the Hypothesis, one subject hash, not the commit. Three
  is fixed for every Goal.
- **The boundary** is the latest of: the latest place that changed an example
  (an `examples.place`, or an old `example.place`; defined below), the latest
  `budget.reset`, and the latest `unselect`. Before any, it is the Goal's start.
- Every return prints `runs: N of 3 used`, where N counts Hypotheses. The last
  run warns:
  `last run: after this, G stops for the owner.`
- A refused run costs nothing.

**Stops**, calculated:

| Stop | Holds when | Caused by | What hyp prints |
|---|---|---|---|
| Selected | The Goal has a select: one that no `unselect` undid. | the `select` | The PR body (§12) |
| Two candidates | Two or more candidates and no select. | the latest run | The table of outputs, rows × examples |
| Budget used | Three Hypotheses counted since the boundary and no candidate. | the latest run | The table of outputs, rows × examples |

**A stop is an event**, calculated from the log and never stored (§3): which
stop, the act that caused it, and when. `status` prints it on the `stop:`
line as `stop: STOP, by ACT ID at TIME`, for example
`stop: selected, by select 01M7W4E3Y3MB9PZB2X2YHY4K7H at 2026-11-20T09:26:09.091171Z`.
For Selected the cause is the `select` event. For the two stops of the runs it
is the latest `result.record` of the Goal in merge order that counts, neither
retracted nor refused by Stable: only a run adds a row or a cell, so the stop
held after it. A Goal with no stop prints `stop: none`. A landed Goal prints
the stop its land names, with that cause, and a line
`landed: by ACTOR, land ID at TIME`.

**Each stop waits for the owner**: `waits: owner`, with `hyp land G` among the
owner's acts. At Selected the owner's acts are `hyp land G` and
`hyp unselect G`. After a select, the owner reads the PR and lands the Goal, before
or after the merge. Owner said, 2026-11-20 (#245): "C". A landed Goal waits
for no one.

Two candidates can happen with one runner: "selects at once" is runner
behaviour, not a guard, so a runner can meet every line on a second leaf before
it selects the first. No guard refuses that run. Candidate then lets only the
owner select. Owner said, 2026-11-19: "Keep the stop as designed".

At "Two candidates" and "Budget used" the state block adds an `outputs:` table:
rows by examples. Each cell shows the verdict, the exit code, the output's
hash, its size in bytes and at most 60 characters of its first line, as hyp
holds them: no quotes and no escapes are added (§12, Values as held). Its full
bytes are in the output file of that hash: in the clone's trace, or under
`.tree/outputs/` for an older log.

hyp prints the table, not a reading of it: no distance, no summary line. The
leaves' statecharts at a stop are a later v2 Goal.

The stop "Budget used" shows once the third Hypothesis has a live
cell for every placed example and no candidate exists; until then the third
Hypothesis's remaining lines stay among the runner's allowed acts. The last-run
warning prints in the state block while `runs: 2 of 3 used`, before the third
Hypothesis's first run.
Owner accepted, 2026-11-19.

An example-changing place is one that writes a new example key or a new hash
on an existing key; every `examples.place` is one, since a place that changes
no example is refused (Adds only). One place starts one budget, however many
examples it names.
Owner accepted, 2026-11-19.

## 12. The return shape and exit codes

Every command, read or write, but `show` and `prune`, returns in one shape
(`try` too):

1. the exit code,
2. a report: what was recorded, or the guard that refused and why,
3. the state block of the Goal: its stop, what it waits for, `runs: N of 3
   used`, and the allowed acts of every role as ready command lines with their
   keys filled in.

```
G-NXHHVA  open  runs: 1 of 3 used
stop: none
waits: none
owner may:
  (none)
grill may:
  hyp place G-NXHHVA
probe-writer may:
  (none)
runner may:
  hyp run <commit> G-NXHHVA/line-1
  hyp run <commit> G-NXHHVA/line-2
```

- The roles print in the order of §2: owner, grill, probe-writer, runner. A
  role with no allowed act prints `(none)`. Only acts the binary has are
  printed.
- **One place line.** While the Goal is written or open, at any stop but
  `selected`, the grill's block holds one ready place line, `hyp place G`,
  which places the whole example set (§8). Once the Goal is selected the grill
  may place nothing, and prints `(none)`. This is line 5 of Goal G-SBFXAY
  (#312). It replaces the one place line per line, `hyp place G
  <line>/<example>`, of Goal G-1908D1.
- **The run on the spec commit.** An open Goal with one candidate that fails
  Spec or Own path on its commit (§7) offers the runner
  `hyp run <commit> G/<line>` for every line that holds an example, also when
  three Hypotheses are counted. Both guards read only `goals/`, which is not in
  the subject, so the fix is a new commit of the same Hypothesis: its run costs
  no budget (§8, `select`). hyp cannot know that commit, so it prints
  `<commit>`. Without this line a Goal at `runs: 3 of 3 used` with such a
  candidate had no stop, no wait and no allowed act (#229).
- A refusal prints to stderr, then the state block to stdout. A report prints
  only on exit 0. Exit 2 and exit 3 print no state block.
- **The selected commit on no ref.** The state block of a Goal with a select
  prints, below its `stop:` (and `landed:`) line, the warning of §8 when the
  selected commit is on no branch or tag of this clone.
- **The record of a landed Goal.** `status G` and the return of `land` print
  the state block of a landed Goal, then the record of its decision: the PR
  body when the Goal has a select, or the table of outputs (§11) when it has
  none. A selected Goal that is not landed prints its PR body in `status G`
  too.
- **The list of every Goal holds no PR body.** `status` for every Goal prints
  only the state block of a selected or a landed Goal. A PR body quotes each
  output byte for byte at column 0, and an output can hold any text: an
  example that runs `hyp status` in a store of its own prints lines such as
  `G-J3MA1E  written`, which a reader that scans lines takes for Goals of this
  store (#255). So in the list each line that starts with a Goal id is a Goal
  of this store, and the PR body stays one `status G` away. This replaces the
  owner's decision of 2026-11-20 (#245) that the list prints a selected
  Goal's PR body; owner decided, 2026-11-21 (#255).
- **What a written or open Goal asks.** `status G`, with or without `--role`,
  prints below the state block of a written or open Goal what it asks, under
  `The Goal asks:`: its words, each line's key, kind and text, and under each
  line every placed example in its current version, with its version, hash,
  who placed it and when, its verdict and exit code on H0 at place, and its
  `example.txtar` byte for byte in a fence, which holds its words, run, inputs,
  floors and ceilings. A line with no placed example prints
  `no example placed`. The example is read from git by its hash, never from
  the files on disk. Only `status G` prints it: `status` for every Goal, the
  return of every act and the return of `try` print the state block alone, and
  a selected or landed Goal prints its record as above. This is line 1 of Goal
  G-X0HY3A: "While a Goal is written or open, `hyp status G`, with or without
  `--role`, prints the Goal's words, each line's key and text, and for each
  placed example its key, words, run, floors, ceilings and its verdict on H0
  at place".
- **The Probe writer's contract.** The Probe writer writes no event, so its
  block holds no act. While a Goal is written or open, `status G`, with no role
  or with `--role probe-writer`, prints in the probe-writer block, every line
  indented: the Probe writer's contract, for the Probe of one example (read the
  line, its example and H0; write and commit only that example's probe folder;
  start from the generic Probe and write new code only when it cannot run the
  example; run the example on H0 with `hyp try` and return its output; write no
  event), the generic Probe, which is `skills/hyp-grill/probe/run` byte for
  byte, built into the binary, and one task for each line of the Goal, in its
  order, with the `hyp try G <line>/<example>` line that checks it. Everywhere
  else the block prints `(none)`: `status` for every Goal, the return of every
  act and of `try`, and a selected or landed Goal. This is line 1 of Goal
  G-GRXA9K.
- `show ACTOR` prints the trace (§8) and no state block: it shows what the
  actor got, not the state now. `prune` prints its list and no state block:
  it is about the trace folders, not a Goal.

**The channel.** One block of code, `internal/channel`, writes every byte hyp
prints and turns every failure into its words and its exit code. Every other
part of the binary reports a failure as data: its kind, the guard that
refused, a code, and the values the failure is about. A test of the channel
holds the boundary: outside it no Go file the binary is built from imports
`fmt`, calls `errors.New` or passes a sentence to a failure, and every code has
words in the channel that name as many values as each failure passes. This is
line 1 of Goal G-9XX4KB: "The communication channel, which is used for `hyp`
to communicate with agent, is one block of code".

- **Values as held.** hyp prints a value as its state holds it: an example's
  words, an output, a path or a command print raw, never through a quote or an
  escape. No marker is put inside a value's bytes. This is line 3 of Goal
  G-9XX4KB, whose gap is that "hyp tells the owner and the agent something its
  state doesn't hold".
- **Byte blocks.** Stored bytes print as a block: `<label>, N bytes:`, a
  newline, the bytes, a newline. A block cut to a cap is `<label>, first N of M
  bytes:`; an empty one is `<label>, 0 bytes`. The cut is said outside the
  bytes, in the label. The PR body puts the bytes of a block in a fence below
  its label.
- **Trace before print.** hyp writes nothing to stdout or stderr until the
  command ends. It then records the trace (§8) in the clone's trace folder,
  then prints stderr, then
  stdout. When the trace cannot be written, hyp prints none of the command's
  bytes, only a notice that it could not record them, with the cause, and
  exits 3. So the one thing an actor can get that its trace lacks is that
  notice, and the notice holds nothing of the state: no Goal, verdict, word or
  output.

- `--role ROLE` filters the allowed acts to one role. It is a convenience of
  the read, not a boundary.
- An agent runs a line hyp printed. A skill names no write command.
- **Split line.** A line whose examples give different verdicts on one
  Hypothesis is marked every time it prints, in the state block and in the PR
  body. It is never a refusal.
- `unstable:` lists every Result refused by Stable: the example, the subject hash, and
  the two output hashes (§7).
- A **Kind** refusal prints H0's output beside the example's floors and
  ceilings.

| Code | Means |
|---|---|
| 0 | Recorded, or the read printed. |
| 1 | Refused by a named guard. |
| 2 | Bad arguments. |
| 3 | Environment: not in a git repository, an object format other than sha1, a missing tool, a Probe that failed or timed out, no absolute `$HOME`, a trace hyp could not write, a trace folder `prune` could not read or remove. A check's own exit code never comes back as hyp's. |

`select`'s report is the PR body: each line, each example's words, output,
stderr, exit code and verdict, and the selected commit and row. `status` prints
it again while the Goal is selected, and `status G` after it is landed. The runner pastes it into the PR, so hyp
writes nothing to GitHub. It is the one read hyp prints per Goal: each line,
each example, its output and its verdict. After the
lines, the PR body prints each chart of the selected commit's spec folder (each
`.mmd` file outside `specs/variants/` and `specs/switches/`) under `## Chart: <path>`, byte for byte,
in a `mermaid` block; a folder with no chart prints none. So the select return,
`status` and `status G` of the selected Goal, and the land return and `status G`
of the landed Goal hold the chart (#247).

`place` and `run` check their guards under the lock, release
it, run the Probes, take the lock again, check every guard again and append; a
second check that fails refuses with exit 1. When the example set or H0's base
changed while the Probes ran, the refusal says what changed: Committed for the
set, Kind for H0's base, Open for a run. Allowed acts are calculated from
the guards that read only the tables and git; Kind and Stable are checked when
the act runs.
Owner accepted, 2026-11-19.

## 13. What hyp reads from git

hyp writes no git state in the host: no refs, objects, config, index or hooks.
Hiding a file from git is git state too: hyp writes no `.gitignore`, no
`info/exclude`, and sets no assume-unchanged or skip-worktree flag. The trace
is outside the worktree and the git directory (§1) for this reason, in
`$HOME/.hyp`. `try --worktree` reads the checkout's uncommitted change
from disk, never by `git stash create` or `git add`, which write objects
and the index. `prune` runs no git command on the clones it lists: whether a
clone is gone is read from the disk, as whether the path its trace names
exists, so a gone clone needs no git state and `prune` writes none. The owner
and the agents commit. hyp runs only these git commands, each checked
by its full argument shape in `internal/gitread`:

| Command | Used for |
|---|---|
| `git rev-parse` | the top level, the object format, a commit, `HEAD:<path>`, `--path-format=absolute --git-common-dir` (the trace's key, §1) |
| `git cat-file` | `-e` (Bytes), `--batch` (extraction by hash), `-p <commit>` (a commit's parents, for the own change), `blob <hash>` (a model's configs for Spec, and the charts of the PR body) |
| `git ls-tree` | `-r -z` (extraction), the root of a commit (subject hash), `-r -z` of `goals/<G>/specs/` (Spec, and the charts of the PR body) |
| `git merge-base --is-ancestor` | Leaf only |
| `git diff-tree` | `-r --name-only` and `-r -c --name-only` (own change paths) |
| `git ls-files -z --cached --others --exclude-standard` | the files of the checkout, for `try --worktree` (§8); this one form only. It reads the index and takes no optional lock; the files' bytes are read from disk |

- Every git command runs with `GIT_OPTIONAL_LOCKS=0` and `GIT_NO_LAZY_FETCH=1`:
  a read takes no optional lock, and in a partial clone a missing object is an
  environment problem (exit 3), never a fetch that writes a pack.
- hyp never runs `git archive`: `export-ignore` and `export-subst` attributes
  change what it extracts, so the bytes could differ from the hash that names
  them. hyp extracts a tree with `git ls-tree -r -z` and `git cat-file --batch`.
  It writes only into real directories under its destination, never through a
  symlink, and never over an existing path; two paths that collide on a
  case-insensitive filesystem are exit 3, naming the path.
- **The runner's own change** of a commit C is what the runner wrote, not what
  it took in from main. Its **base** is found by walking first parents from C
  (C excluded) to the first commit that is H0 or a merge commit. The own
  change is the diff from the base to C. Its paths are the union of each
  non-merge commit's `git diff-tree -r --name-only` against its parent on that
  walk, plus, when C itself is a merge commit, the paths of
  `git diff-tree -r -c --name-only C` (only what differs from every parent: the
  runner's own conflict resolution). hyp reads parents with `git cat-file -p`.
  Limit: when main took another Goal's record branch by a squash or a rebase,
  not a merge commit, the walk does not stop there, and Own path can refuse a
  path the runner did not write; the refusal names the commit, so the runner
  takes main in by a merge instead.
- A git tree entry with mode `160000` (a submodule) is refused, exit 1, naming the
  path.
- **The subject hash** of a commit is calculated in memory from `git ls-tree`
  of the commit's root without the entries `goals` and `.tree`: the git tree
  hash that tree would have. hyp writes no object.
- **The disk hash** of a folder, the example hash that Committed and `try`
  read, is its git tree hash calculated from its raw bytes, as git would hash
  them with no filter. A file is mode `100755` when its owner
  execute bit is set, as git reads it, else `100644`. A clean filter (`autocrlf`, `eol`, LFS),
  `core.filemode=false` or a stray file makes Committed refuse; the refusal
  names the first differing path.

Sha1 repositories only: `git rev-parse --show-object-format`
other than `sha1` is exit 3.
Owner accepted, 2026-11-19.

`why` is optional on every event, at most 600 runes.
Owner accepted, 2026-11-19.


## 14. Deferred

Deferred by the owner until a real event shows the need:

- **o7-actors** — whether two rival builds need two actors that cannot see each
  other. Owner said, 2026-11-18: deferred, "the two items below wait for an
  event".
- **o11-blind** — whether the Probe writer can be kept from reading the
  candidate. Owner said, 2026-11-18: deferred with o7-actors.
- **o8-end** — closed by Goal G-M66KKY: the owner's `land` is the act that
  ends a Goal, at a stop (§8). Not landed refuses every write after it. A Goal
  that never reaches a stop has no end yet; that waits for a real event, as
  the owner said: "whether we need an end state + end action for owner will
  leave open until we see 'dangling open' goal is the problem later on".
