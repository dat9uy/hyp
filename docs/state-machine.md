# hyp v2 state machine

This is the one spec of hyp v2 (events v2). It states what hyp stores, what it
calculates from what it stores, each state and what it refuses, every act with
its guards and exit codes, and what happens on a crash or a merged log. The
design behind it, with the owner's words, is on the local design page "Where
the model stands"; this file wins over it for v2 behaviour once merged.

Tags: **Owner said** marks a rule the owner stated or accepted, with the date.
Every other rule here is derived from such a rule, and names it.

## 1. Names and locations

- **The store** is one folder, `.tree/`, at the git top level
  (`git rev-parse --show-toplevel`). It holds every Goal of the repository.
- **Events** are kept one file per actor: `.tree/events/<actor>.jsonl`, one JSON
  line per event. Only hyp writes them.
- **Union merge.** `.tree/.gitattributes` holds `events/*.jsonl merge=union`, so
  git merges two branches of one events file by keeping every line of both.
  Replay orders events by id, never by their place in the file (§4).
- **Goal folders** are written by the owner and the agents, never by hyp:

  ```
  goals/<Goal>/<line>/<example>/   one example folder
      example.txtar                the example file (§9)
      probe/                       the Probe code (§10)
  goals/<Goal>/specs/              the runner's spec and chart for the Goal
  ```

- **The subject** of a commit is the commit's tree without the top-level entries
  `goals` and `.tree`. A change to an example, a Probe, a spec or the store makes
  no new subject.
- An example folder is identified by its **git tree hash**. The hash of a line
  folder or a Goal folder is never used.
- Example folders travel on the Goal's record branch, which merges before the
  runner's run, so the examples reach the runner's commits.

## 2. Ids, actors and roles

- **Event id**: a ULID, unique in the store, sorting after the newest id the
  store holds (v1 `idAfter`).
- **Goal id**: `G-` plus the last six characters of its `goal.add` event's id
  (as v1 `event.DisplayID`), for example `G-NXHHVA`. **Result id**: `R-` plus the
  last six characters of its `result.record` event's id. A short id that matches
  two events is refused as ambiguous (exit 2).
- **Line key**: `line-1`, `line-2` … in the order the Goal lists its lines.
  **Example key**: `<line>/<example>`, for example `line-3/a`. An example name
  matches `^[a-z0-9][a-z0-9-]{0,39}$`.
- **Row key**: the row tree, see §5.
- **Actor**: a name matching `^[a-z0-9][a-z0-9-]{0,39}$`. It names the events
  file the act appends to.
- **Roles**: `owner`, `grill`, `probe-writer`, `runner`. The grill also does the
  re-grill. The Probe writer writes no event.
- The actor and the role are self-declared. A guard that reads them catches an
  honest mistake, not a lie. hyp has no authorization layer.

## 3. Events

Every event is one JSON line with this envelope (v1's, with `v` = 2):

| Field | Holds |
|---|---|
| `v` | `2`. A line with another `v` does not parse. |
| `id` | The event's ULID. |
| `ts` | hyp's time, RFC 3339 UTC. |
| `actor` | The actor; equal to the events file's base name. |
| `type` | One of the six types below. |
| `node` | The id the event is about: the Goal's `goal.add` id, or the Result's `result.record` id for `result.retract`. |
| `payload` | The type's payload, parsed strictly: an unknown field does not parse. |
| `why` | The actor's short reason, optional. |

The six events:

| Event | Written by | `node` | Payload |
|---|---|---|---|
| `goal.add` | `add goal` | its own id; a Draft replace names the replaced Goal's id | role, words, H0 commit, H0 row tree, budget (3), `from` Goal or empty, lines: key, text, kind |
| `example.place` | `place` | the Goal | role, line key, example name, the folder's tree hash, and H0's cell: H0 commit, H0 row tree, Probe hash, output hash, stderr hash, exit code, verdict |
| `result.record` | `run`, `land` | the Goal | role, line key, commit, row tree, `refused` (empty, or `Stable` with the differing example keys), and per example: example key, example hash, Probe hash, output hash, stderr hash, exit code, verdict |
| `result.retract` | `retract` | the Result | role |
| `select` | `select`, `land` | the Goal | role, commit, row tree |
| `budget.reset` | `reset` | the Goal | role |

The Probe hash is the tree hash of the example's `probe/` folder. The example
hash already covers it; it is kept on the cell so hyp can show which examples
share one Probe (main page §4, "Copy, and let hyp show the copies").

The six events are the list of `v2-architecture.html` §8. Owner said,
2026-11-19: "Why ask me, event list is already decided in report".

`example.place` carries H0's cell for the example it places, with the fields of
a Result's cell. So one placement is one line, and a crash
cannot leave a placement without its H0 verdict. This follows the main page §6:
"A new example runs on H0 as part of the act that adds it". That cell is live on
H0's row like a Result's cell; it is not a Result and cannot be retracted. It
is live only while its row tree equals the Goal's H0 row tree: after a Draft
replace that changes the base it reads `not run` (main page §10: "A changed H0
subject shows its cells as not run"), and `run <H0 commit> LINE` measures it.

## 4. Merge order, crash and merged log

- **Lock.** Every write act takes the store's directory lock, checks its guards,
  appends one line and releases the lock (v1's lock).
- **Merge order.** Replay reads every events file and orders the events by id,
  then by actor, then by line bytes (v1 `MergeOrder`), so the order of lines in a
  file decides nothing.

- **Crash.** An act writes its output files first (§5, Outputs), then its
  event as one `write` of one line and an `fsync`. A crash before the event
  leaves output files no event names; they are harmless. A torn last line does
  not parse: replay reports it with its file and line number and skips it.
  Before it appends, hyp writes a newline if the file does not end in one, so a
  torn line stays alone and the next event is not glued to it.
- **Merged log.** Two clones that each passed their guards can merge into a log
  that breaks one, for example a fourth tree in one budget. Replay checks every
  event at its place in the merge order. An event whose guards fail there is
  skipped, listed under "refused at replay" in `status`, and never deleted (main
  page §11: "A guard that a merged log breaks shows when hyp replays it, as v1
  does"). Output files are named by their hash, so they never conflict in a
  merge.
- **Which guards replay checks again.** Only the guards that read the tables:
  Draft, Whole lines, Own case, Open, Goal's actor, Budget, Stable, Candidate,
  Selected and Own Result, and the table half of Adds only (the new hash is not
  the current one). The guards that read disk, git objects, the moving git state
  or a Probe (Committed, Kind, Bytes, Leaf only, Spec, Own path, Same change, and
  the content half of Adds only) are trusted as written. Cost: two clones that
  each adjust one example key merge to @2 and @3, and @3 was checked against @1
  only. So when
  two clones merge and one Goal holds a fourth counted tree, the fourth by
  merge order is refused at replay: its Result counts for nothing, even when it
  was a candidate. Owner said, 2026-11-19: "Refused at replay".
- These rules follow from rules already decided: the store of ADR 0020 and
  ADR 0022, and the main page §11. Owner said, 2026-11-19: "Why ask me, event
  list is already decided in report".

## 5. Calculated tables

hyp stores no table. On every command it reads the log and calculates these
tables in memory. Each guard (§7) is a query over them, or over git.

| Table | Key | Columns |
|---|---|---|
| Goals | Goal id | words, H0 commit, H0 row tree, budget, actor (of the first `goal.add`; a Draft replace keeps it), `from`, state (§6) |
| Lines | Goal, line key | text, kind (`gap` or `guard`) |
| Placements | Goal, example key | example hash, version `@n` (count of distinct hashes placed on that key, in order), time, actor; the latest is current |
| Results | Result id | Goal, line key, commit, row tree, actor, role, time, retracted, refused by Stable |
| Cells | Result, example hash | example key, Probe hash, output hash, stderr hash, exit code, verdict (`met` or `failed`) |
| Rows | Goal, row key | the commits it was seen at; leaf (from git); per example key, the verdict of its live cell under the current placement, or `not run` |
| Selects | Goal | commit, row tree, time, actor |

- A cell is **live** when its Result is not retracted, not refused by Stable,
  and its example hash is the current placement's hash for its example key.
- **Outputs.** An output is a file named by its sha256 under
  `.tree/outputs/<sha256>`; the cell holds the hash. Equal outputs are stored
  once, and "same bytes" is a compare of two hashes. hyp writes the file before
  the event that names it. Owner said, 2026-11-19: "File by sha256" (the lean of
  `v2-architecture.html` §8).
- **Cell key.** A cell is keyed by its Result and the example's hash, not by
  the example key and version. An adjust makes new cells; the old cells stay as
  history and are not live under the new placement. Owner said, 2026-11-19:
  "Example hash" (the lean of `v2-architecture.html` §8).
- **Row key.** A row is keyed by its Goal and its row tree: the subject tree
  hash hyp calculates (§13). The commits it was seen at are a column, because a
  leaf comes from commit ancestry and one tree can sit at several commits. A row
  is a **leaf** when none of its commits has a descendant commit among the
  commits of the Goal's other rows. Owner said, 2026-11-19: "Tree; leaf = no
  commit has a descendant".

## 6. States

Each state is calculated from the tables; none is stored. `selected` wins over
`open`: a selected Goal is not open.

### Goal: `written → open → selected`

| State | Holds when | Refuses |
|---|---|---|
| `written` | A `goal.add` exists, and the Goal is not open. | `run` on a tree other than H0's (**Open**), `select` (**Candidate**: no row is met yet), `land` (**Selected**) |
| `open` | Every line has at least one placed example, and every current placement has a live cell on H0's row with the verdict its line's kind needs: `failed` on a gap line, `met` on a guard line. | `add goal G` replace (**Draft**), `land` (**Selected**) |
| `selected` | The Goal has a select. | `add goal G` replace (**Draft**); `place`, `run` and a second `select` (**Selected**: the Goal already has a select) |

All refusals are exit 1. A Draft replace that changes the lines leaves existing
placements on their keys; a placement whose line key is gone is history, and the
Goal is open again only when the rule above holds.

`selected` refuses `place`, a second `select` and `run` by a
guard named for the state, **Selected**, so that no refusal reads "Open" by
accident ("a run after the select counts nothing", main page §10). `land` is
the only write it allows.
Owner accepted, 2026-11-19 (S13).

### Example: placed `@1 → @2 → …`

An example key is placed when its first `example.place` is written. Each later
placement of a different hash on the same key is the next version. A placement
of the hash already current changes nothing and is refused (Adds only: no
new version). A version changes only floors and ceilings; a changed input, run,
`external` section or Probe is a new example key (§7). The old versions are history; their cells are no longer live.

### Result: `recorded → retracted`

`retract R` moves a Result to retracted. Its cells stop being live, so their
examples read `not run` on that row again. A retracted Result is never deleted.

### Row

A row has no stored state. H0's row always exists. Any other row exists when a
Result names its tree. Its leaf mark and its verdicts are calculated (§5). When
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
- **Not ended** — No write after the Goal ends. No act ends a Goal yet (§14), so
  it always holds. Query: none.
- **Open** — The Goal is open. Checked by `run` unless the commit's row tree
  equals H0's row tree. Query: `Goals[G].state == open`.
- **Goal's actor** — Only the actor that added the Goal selects for the owner or
  resets the budget. Query: the act's actor `== Goals[G].actor`.
- **Own case** — Only the grill places an example; the runner never places.
  Query: the act's role `== grill`.
- **Kind** — H0 gives the verdict the line's kind needs under the example's
  Probe: a gap example fails on H0, a guard example is met on H0. Query: run the
  Probe on H0's commit (§10), compare (§9), read the verdict.
- **Committed** — The example folder's bytes on disk equal the committed bytes.
  Query: the folder's tree hash computed from disk (§13) equals
  `git rev-parse HEAD:goals/<G>/<line>/<example>`.
- **Adds only** — A new version of an example only adds floor or ceiling
  sections. Everything else is byte-equal to the current version: the `words`,
  `run`, every `input/…` and the `external` section, and the `probe/` folder.
  Any other change is a new example key, not a new version. Query: read the
  current placement's folder by its hash (§13); the `probe/` tree hashes are
  equal; the non-floor, non-ceiling sections are equal by name and bytes; every
  floor and ceiling of the current version (kind and literal) is in the new
  file; and the new file holds at least one more. This follows the owner's
  "re grilling will have two choice, adjust (floor and ceiling) and create new
  Goal if need new line" (main page §8). The refusal names the first section or
  path that differs.
- **Leaf only** — A run is on H0, or on a commit that no row of the Goal descends
  from. Query: for each commit `c` of the Goal's rows other than COMMIT,
  `git merge-base --is-ancestor COMMIT c` fails.
- **Nothing new** — Removed. It refused a second run of an example on the same
  tree under the same example hash. Stable now judges every repeat instead.
  Owner said, 2026-11-19: "Drop Nothing new". This replaces the main page's §11
  guard row and its §6 sentence "The guards Leaf only and Nothing new enforce
  it"; Leaf only alone keeps runs on H0 and leaves.
- **Bytes** — The commit exists. Query: `git cat-file -e COMMIT^{commit}`.
- **Stable** — Two runs on the same bytes give one verdict. Checked by `run`
  after the Probes run: for each example run, no live cell on the same row tree
  under the same example hash has another verdict. A Stable refusal is exit 1,
  and its cells never become live, so the first verdict stands. It still keeps
  what it saw: the new output files stay, and `run` appends its `result.record`
  marked `refused: "Stable"` with the examples whose verdict differed. That
  Result counts nothing: no live cell, no budget, no retract. `status` lists it
  under `unstable:` with both outputs' hashes, so the owner sees the unstable
  example. The report prints both outputs. A
  retracted cell is not live, so a retract followed by a rerun is never refused
  by Stable. A repeat on a tree already counted costs no budget (§11). Owner
  said, 2026-11-19: "Drop Nothing new" (C5: keep Stable alone).
- **Budget** — No fourth distinct tree since the boundary, H0 never counted.
  Query: the distinct row trees, other than H0's, of the Goal's Results recorded
  after the boundary (§11), without Results refused by Stable; refuse when there are three
  and the commit's row tree is not among them. A tree run before the boundary
  and run again after it counts: it is run against the new evidence. Owner said,
  2026-11-19: "It counts".
  A retracted Result still counts: a retract fixes a wrong
  Result, and does not give a run back (main page §8: a runner that hides failed
  paths is what the budget cannot see).
- **Candidate** — The tree is a candidate: a leaf row met on every line under the
  current placements. With exactly one candidate the runner may select it; with
  two or more, only the owner. Query: Rows of G with leaf = true and every
  current example key `met`.
- **Spec** — The selected commit holds a spec under `goals/<G>/specs/`. Query:
  `git ls-tree COMMIT goals/<G>/specs/` lists at least one file.
- **Own path** — A commit writes no path keyed by another Goal. Only the
  runner's own change of the commit is read (§13), never the diff from H0: when
  the runner takes in a main that merged another Goal's record branch, the diff
  from H0 holds `goals/<X>/` that the runner did not write. Query: the paths of
  the runner's own change of COMMIT hold none under `goals/<X>/` with `X` ≠ G.
  The refusal names the path and the commit that wrote it.
- **Selected** — The Goal has a select. Query: `Selects[G]` exists. `land`
  needs it to hold; `place`, `run` and `select` need it not to hold (§6).
- **Same change** — The landed commit carries the same whole change as the
  selected commit. Query: the patch-id of the selected commit's own change (its
  diff from its base, §13) equals the patch-id of the landed commit's own change
  (its diff from its new base). Each is `git diff-tree -p <base> <commit>`
  piped to `git patch-id --stable`. Comparing each commit against its first
  parent would compare only the last commit of a change made of several.
  Limit: when the runner took main in by a merge commit before the select, the
  selected commit's base is that merge, so its own change holds only the commits
  after it; a rebase that flattens that merge then gives a different patch-id and
  is refused. `land` is built as a later v2 Goal, which tests this against a real
  landing.
- **Own Result** — Only the actor that ran a Result retracts it. Query: the act's
  actor `== Results[R].actor`.
  Owner accepted, 2026-11-19 (S17).

## 8. Acts

The act comes first, then its keys. Every write act takes the lock, checks its
guards in the order listed, appends one event, and prints the return (§12).
`land` is the one exception: it appends a `result.record` per line and, when met
on every line, a `select`.
Common to every act: exit 2 for bad arguments; exit 3 for an environment problem
(§12).

| Act | Role | Guards, in order | Writes |
|---|---|---|---|
| `add goal [G] [--from G] --base H` | owner | Draft, Whole lines | `goal.add` |
| `place G LINE/EXAMPLE` | grill | Not ended, not Selected, Own case, Committed, Adds only, Kind | `example.place` |
| `run COMMIT LINE` | runner | Not ended, not Selected, Bytes, Open (not for H0), Own path, Leaf only, Budget, Stable | `result.record` |
| `select G TREE` | runner, or owner | Not ended, not Selected, Candidate, Goal's actor (owner only), Spec, Own path | `select` |
| `land G COMMIT` | runner | Selected, Same change | `result.record` per line; `select` when met on every line |
| `reset G` | owner | Not ended, Goal's actor | `budget.reset` |
| `retract R` | the actor that ran R | Own Result | `result.retract` |

- `select G TREE` records the commit of the row's latest Result in merge order.
  Spec and Own path read that commit, and the PR holds it. Every commit of a leaf
  row has no descendant among the Goal's rows (§5), so the commit is a leaf. A
  spec committed after the last run is not in that commit: the runner runs one
  line on the spec commit first. Its tree is the same row, so it costs no
  budget, and Stable checks the repeat.
- `place` checks Committed and Adds only before it runs the Probe for Kind, so a
  refusal that needs no Probe costs no Probe run.
- `run` checks every guard that needs no Probe before it runs a Probe; Stable is
  checked on the outputs.
- `land`, `reset` and `retract` are specified here and built as later v2 Goals.
- Reads write nothing and are always allowed: `status`, `show KEY`,
  `prompt ROLE`. `show` and `prompt` are later v2 Goals.

No `init` act: `add goal` creates `.tree/` at
`git rev-parse --show-toplevel` when it is absent. Every other act with no store
is exit 1 naming "no store; run hyp add goal". Outside a git repository is exit 3.
Owner accepted, 2026-11-19 (S3).

The actor comes from `--actor` or `$HYP_ACTOR`, the role from
`--role` or `$HYP_ROLE`. Each event stores its role in the payload. A write act
missing either, or naming a role outside the four, is exit 2. Only Own case and
Candidate read the role; elsewhere it is recorded, not enforced.
Owner accepted, 2026-11-19 (S4).

Own path is checked by `run` and `select`: they refuse a commit whose own
change (§13) touches `goals/<other Goal>/`. The main page §10 names it on both acts;
`v2-architecture.html` §7 lists it in no act, and the main page wins.

`run` names its Goal in LINE, written `<Goal>/<line>`, for
example `hyp run 4df2b9e G-NXHHVA/line-3`, so the act stays first and then two
keys.
Owner accepted, 2026-11-19 (S10).

`add goal` takes repeated `--line gap:TEXT` and
`--line guard:TEXT`, in order; hyp gives the keys `line-1`, `line-2` …. A Draft
replace is `hyp add goal G …`. `--from G` stores only the cited Goal's id and
copies nothing. Whole lines also refuses more than ten lines
(`v2-architecture.html` §2: "up to ten lines").
Owner accepted, 2026-11-19 (S14).

A line with no placed example has nothing to run: `run` refuses it by Open,
even on H0. `run` runs every placed example of the line on the commit, whether or not a
cell already exists on that tree; a repeat is judged by Stable. Owner said,
2026-11-19: "'live' is depend on the version of probe. So when new probe
created, Why not [j]ust run H0 and leaf without caring about the live ?"

## 9. The example file and the compare

An example is one txtar file, `example.txtar`. Each section starts with a line
`-- name --`. The sections:

| Section | Holds | Count |
|---|---|---|
| `words` | The owner's sentence for this example. hyp prints it. | exactly one |
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
  and what its example's `external` section names (main page §4: "what its
  External: line names"). Owner said, 2026-11-19: "-- external -- section".
- A Probe failure that is not the run's own failure records nothing (exit 3).

**How hyp runs a Probe.**

- hyp runs `probe/run`, an executable file of the example folder, with empty
  stdin and a timeout of 10 minutes.
- Its **working directory** is a fresh copy of the commit's subject, extracted
  by hash (§13). `HYP_SUBJECT` names it too.
- `HYP_EXAMPLE` names a fresh working copy of the example folder, extracted by
  its tree hash. Each `input/<path>` section is written there at
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
grep -rlE -f "$HYP_EXAMPLE/input/patterns.txt" --include=*.go . | grep -v _test.go | grep -v internal/render/
```

Committed compares the folder's tree hash from disk with
`git rev-parse HEAD:goals/<G>/<line>/<example>`. hyp extracts the example for a
run by its tree hash (`ls-tree -r -z` and `cat-file --batch`), so a run does not
depend on the branch checked out. This follows the main page §5: "hyp places
only a folder whose on-disk hash equals its committed hash".

## 11. Budget and stops

- **Budget**: at most three distinct trees run, H0 never counted, since the
  boundary. The unit is the row tree (main page §8: "three distinct trees run";
  `v2-architecture.html` §7 says "commits", and the main page wins). Three is
  fixed for every Goal.
- **The boundary** is the latest of: the latest `example.place` that changed an
  example (defined below), and the latest `budget.reset`. Before either, it
  is the Goal's start.
- Every return prints `runs: N of 3 used`. The last run warns:
  `last run: after this, G stops for the owner.`
- A refused run costs nothing.

**Stops**, calculated:

| Stop | Holds when | What hyp prints |
|---|---|---|
| Selected | The Goal has a select. | The PR body (§12) |
| Two candidates | Two or more candidates and no select. | The table of outputs, rows × examples |
| Budget used | Three trees counted since the boundary and no candidate. | The table of outputs, rows × examples |

Two candidates can happen with one runner: "selects at once" is runner
behaviour, not a guard, so a runner can meet every line on a second leaf before
it selects the first. No guard refuses that run. Candidate then lets only the
owner select. Owner said, 2026-11-19: "Keep the stop as designed".

hyp prints the table, not a reading of it: no distance, no summary line. The
leaves' statecharts at a stop are a later v2 Goal.

The stop "Budget used" shows once the third tree has a live
cell for every placed example and no candidate exists; until then the third
tree's remaining lines stay among the runner's allowed acts. The last-run
warning prints in the state block while `runs: 2 of 3 used`, before the third
tree's first run.
Owner accepted, 2026-11-19 (S15).

An example-changing placement is one that writes a new example
key or a new hash on an existing key. A placement of the current hash changes
nothing (refused by Adds only).
Owner accepted, 2026-11-19 (S16).

## 12. The return shape and exit codes

Every command, read or write, returns in one shape:

1. the exit code,
2. a report: what was recorded, or the guard that refused and why,
3. the state block of the Goal: its stop, what it waits for, `runs: N of 3
   used`, and the allowed acts of every role as ready command lines with their
   keys filled in.

```
G-NXHHVA  open  runs: 1 of 3 used
stop: none
waits: none
runner may:
  hyp run <commit> G-NXHHVA/line-1
owner may:
  hyp reset G-NXHHVA
```

- `--role ROLE` filters the allowed acts to one role. It is a convenience of
  the read, not a boundary.
- An agent runs a line hyp printed. A skill names no write command.
- **Split line.** A line whose examples give different verdicts on one tree is
  marked every time it prints, in the state block and in the PR body (main page
  §6). It is never a refusal.
- `unstable:` lists every Result refused by Stable: the example, the tree, and
  the two output hashes (§7).
- A **Kind** refusal prints H0's output beside the example's floors and ceilings
  (main page §4).

| Code | Means |
|---|---|
| 0 | Recorded, or the read printed. |
| 1 | Refused by a named guard. |
| 2 | Bad arguments. |
| 3 | Environment: not in a git repository, an object format other than sha1, a missing tool, a Probe that failed or timed out. A check's own exit code never comes back as hyp's. |

`select`'s report is the PR body: each line, each example's words, output,
stderr, exit code and verdict, and the selected commit and row. `status` prints
it again while the Goal is selected. The runner pastes it into the PR, so hyp
writes nothing to GitHub. This follows the main page §8: "hyp prints one read
per Goal: each line, each example, its output and its verdict". Charts are a
later v2 Goal.

`place` and `run` check their guards under the lock, release
it, run the Probes, take the lock again, check every guard again and append; a
second check that fails refuses with exit 1. Allowed acts are calculated from
the guards that read only the tables and git; Kind and Stable are checked when
the act runs.
Owner accepted, 2026-11-19 (S11).

## 13. What hyp reads from git

hyp writes no git state in the host: no refs, objects, config, index or hooks.
The owner and the agents commit. hyp runs only these git commands, each checked
by its full argument shape in `internal/gitread`:

| Command | Used for |
|---|---|
| `git rev-parse` | the top level, the object format, a commit, `HEAD:<path>` |
| `git cat-file` | `-e` (Bytes), `--batch` (extraction by hash), `-p <commit>` (a commit's parents, for the own change) |
| `git ls-tree` | `-r -z` (extraction), the root of a commit (row tree), `goals/<G>/specs/` (Spec) |
| `git merge-base --is-ancestor` | Leaf only |
| `git diff-tree` | `-r --name-only` and `-r -c --name-only` (own change paths), `-p` (Same change) |
| `git patch-id` | Same change |

- hyp never runs `git archive`: `export-ignore` and `export-subst` attributes
  change what it extracts, so the bytes could differ from the hash that names
  them. hyp extracts a tree with `git ls-tree -r -z` and `git cat-file --batch`.
- **The runner's own change** of a commit C is what the runner wrote, not what
  it took in from main. Its **base** is found by walking first parents from C
  (C excluded) to the first commit that is H0 or a merge commit; for a landed
  commit the walk stops at the first merge commit, which is its new base. The own
  change is the diff from the base to C. Its paths are the union of each
  non-merge commit's `git diff-tree -r --name-only` against its parent on that
  walk, plus, when C itself is a merge commit, the paths of
  `git diff-tree -r -c --name-only C` (only what differs from every parent: the
  runner's own conflict resolution). hyp reads parents with `git cat-file -p`.
  Limit: when main took another Goal's record branch by a squash or a rebase,
  not a merge commit, the walk does not stop there, and Own path can refuse a
  path the runner did not write; the refusal names the commit, so the runner
  takes main in by a merge instead.
- A tree entry with mode `160000` (a submodule) is refused, exit 1, naming the
  path.
- **The row tree** of a commit is calculated in memory from `git ls-tree` of the
  commit's root without the entries `goals` and `.tree`; hyp writes no object.
- **The disk tree hash** of a folder is calculated from its raw bytes, as git
  would hash them with no filter. A clean filter (`autocrlf`, `eol`, LFS),
  `core.filemode=false` or a stray file makes Committed refuse; the refusal
  names the first differing path.

Sha1 repositories only: `git rev-parse --show-object-format`
other than `sha1` is exit 3.
Owner accepted, 2026-11-19 (S6).

`why` is optional on every event, at most 600 runes.
Owner accepted, 2026-11-19 (S8).


## 14. Deferred

Deferred by the owner until a real event shows the need:

- **o7-actors** — whether two rival builds need two actors that cannot see each
  other. Owner said, 2026-11-18: deferred, "the two items below wait for an
  event" (main page §13, Batch 4).
- **o11-blind** — whether the Probe writer can be kept from reading the
  candidate. Owner said, 2026-11-18: deferred with o7-actors (main page §13,
  Batch 4).
- **o8-end** — an ended state and an `end G` act. Owner said: "whether we need
  an end state + end action for owner will leave open until we see 'dangling
  open' goal is the problem later on" (main page §8). Until then **Not ended**
  always holds.
