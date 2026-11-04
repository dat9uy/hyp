# Spec: events, refusals and replay rules

This file is the single authority for the event schema, the refusals, the derived-status rules and the sandbox lifecycle (§9) of `hyp`. Plans, skills and Go comments point here instead of restating the rules; Go comments cite it by section as "spec §N".

Keeping it current:
- A change to an event type, a refusal or a derived rule updates this file in the same PR.
- Section and rule numbers are cited from code, so never renumber them: add a new refusal at the end of §5's list and a new rule under the section it belongs to.
- If code and this file disagree, treat it as a bug. Decide which one is wrong, fix it, and say which in the PR.
- Text marked **[not built: #N]** is decided but not built yet. It states the target, so it is not a disagreement under the rule above. The PR that builds it removes the mark, and issue #N tracks what is left.

This spec adds the minimum needed to make the owner-confirmed design executable; additions are marked **[added]** with the reason.

## 1. Names and locations

- Binary: `hyp` (the word `tree` is aliased in the owner's shell and collides with the common `tree` utility). "tree" stays the concept name.
- Store directory in the investigated repo: `.tree/`, found by walking up from the current directory, the way git finds `.git`.
  - `.tree/events/<actor>.jsonl`: git-tracked, append-only.
  - hyp keeps no index; every read replays the events (ADR 0006). Older hyp binaries write `.tree/index.sqlite`, so the `.gitignore` below still names it; hyp neither reads nor deletes one.
  - `.tree/evidence/<experiment-display-id>/`: evidence files referenced by pointer.
  - `.tree/.gitignore` contains exactly `index.sqlite` plus `index.sqlite-*`.
- Recipes: `recipes/<name>@<version>/`, in the directory that contains `.tree/` or in any parent up to the repository root (the first directory holding `.git`). See `skills/hyp-verification/RECIPE-FORMAT.md`. §9 says when hyp runs a recipe's commands and whether a sandbox exists then.
- Actor: the `--actor` flag, or the `HYP_ACTOR` environment variable. It must match `^[a-z0-9][a-z0-9-]{0,39}$`. If it is missing, every write command fails with exit 2. There is no default actor.

## 2. Ids

- Every event has `id`, a ULID (26 characters, Crockford base32, upper case), generated monotonic per process.
- `node.add` and `result.record` create a node whose node id equals the event id.
- Display id: `<K>-<last 6 characters of the ULID>`, where K is `G` (Goal), `Q` (DesignQuestion), `H` (Hypothesis), `E` (Experiment) or `R` (Result). Example: `H-7QK2ZD`.
- Optional alias on `node.add`: it must match `^[a-z][a-z0-9-]{0,31}$` and be unique in the log.
- Resolving a user-typed reference, in order: an exact ULID, then an alias, then a display id. Zero matches gives `unknown node <ref>`. Two or more matches gives `ambiguous node <ref>`. Both exit 2, because the fix is in the command line.

## 3. Event envelope

One JSON object per line:

```json
{"v":1,"id":"01JB...","ts":"2026-10-24T03:00:00.123456789Z","actor":"claude-a","type":"node.add","node":"01JB...","payload":{...},"why":"..."}
```

- `v` must equal `1`. `ts` must be RFC3339Nano in UTC (ends in `Z`).
- `why` is required on every event: 1–300 characters, counted in runes (`utf8.RuneCountInString`), never bytes.

## 4. Event types (closed set)

| type | `node` | payload fields |
|---|---|---|
| `node.add` | the new id (= event id) | `kind` (Goal, DesignQuestion, Hypothesis, Experiment), `parent` (id; `""` only for Goal), `title`, `body` (optional), `alias` (optional), `recipe` (Experiment only, required: `"name@version"` or `"none"`), `recipe_hash` (hex sha256, required when recipe ≠ none), `targets` (Experiment only, optional list of Hypothesis ids), `decided_by` (Experiment only, optional: `"reading"`, the verdicts are decided by reading) |
| `result.record` | the new Result id (= event id) | `parent` (Experiment id), `verdicts` (object: Hypothesis id → `supports` / `contradicts` / `inconclusive`; at least one entry), `observation`, `evidence` (list of at most 10 strings, each at most 300 runes), `evidence_hashes` (optional object: evidence pointer → hex sha256 of the recipe bundle it names; each key is an entry of `evidence` that cleans to `recipes/<name>@<version>/<file>`), `limits` (optional list of at most 10 strings, each 1–300 runes), `closes` (optional list of limit keys `<result-id>#<n>`) |
| `result.retract` | Result id | `{}` |
| `hypothesis.abandon` | Hypothesis id | `{}` |
| `hypothesis.amend` | Hypothesis id | `body` (the new body, required) |
| `dq.select` | DesignQuestion id | `hypothesis`, `prev_select` (event id or `""`), `tradeoff` (bool), `scope_change` (bool) |
| `dq.unselect` | DesignQuestion id | `prev_select` |
| `dq.defer`, `dq.withdraw`, `dq.resume` | DesignQuestion id | `{}` |
| `experiment.withdraw` | Experiment id | `{}` |
| `dq.request_verify` | DesignQuestion id | read-only history: older binaries wrote `select`; replay accepts the event, whatever it holds, and does nothing with it |
| `confirm` | DesignQuestion id | `select` (a `dq.select` event id on this DQ), `role` (`verifier` or `owner`), `verdict` (`sufficient` or `insufficient`), `accepts` (optional list of limit keys `<result-id>#<n>`) |
| `reaffirm` | Hypothesis or DesignQuestion id | `{}` |
| `checkpoint` | Goal id | `summary` |
| `edge.add`, `edge.remove` | the edge's `from` node id | `type` (`depends_on`, `supersedes` between two Hypotheses or two Experiments, `external_ref`), `to` (node id; for `external_ref` a string such as `gh:976` or a URL, at most 300 runes) |

**[added] `dq.resume`**: undoes a defer or withdraw. Without it a mistaken defer is permanent in an append-only log.
**[added] `confirm` with `role`**: §3 requires a verifier check on every select and an owner check on trade-offs. Replay can only be pure if both are stored events.
**[added, then removed from use] `dq.request_verify`**: asked the verifier to judge the current select again. An insufficient verdict now lapses when the Dossier changes (§7.4), and a new verifier can confirm at any time, so no command writes this event (ADR 0011). Old logs keep it.
**[added] `experiment.withdraw`**: closes an Experiment that will not run and that no other Experiment replaces; one that is replaced gets a `supersedes` edge from its replacement instead (ADR 0008). Without either, an Experiment that will never run stays `planned`.
**[added] `hypothesis.amend`**: replaces a Hypothesis body. Without it a wrong body can only be fixed by superseding the Hypothesis.
**[added] limits**: a Result names the limits of the design it found but did not test or fix (`limits`), and a later Result closes them (`closes`). A limit's key is `<result-id>#<n>`: the Result's full ULID, then `n`, which counts the Result's `limits` from 1. The CLI accepts any node reference before the `#` (§2) and stores the full id. An owner confirm accepts limits (`accepts`). §8 derives each limit's state.

Content limits, in runes: `title` ≤ 120 and non-empty; `body`, `observation` and `summary` ≤ 600 (an amend's `body` is also non-empty); `why` 1–300.

## 5. Refusals

The same function, `replay.Validate(state, event) error`, is used by the CLI before appending and by replay on every event. At replay time a refused event is **skipped and reported** as a `Rejection{EventID, Reason}`; replay never panics and never stops on a bad event. This covers logs merged from several worktrees, where each event was valid alone.

An event is refused when any of these holds:

1. `v` ≠ 1, the type is unknown, a required field is missing, an unknown payload field is present, `ts` is not UTC, or a content limit is exceeded.
2. Its `id` duplicates an earlier event id.
3. `node` or a referenced id (parent, target, hypothesis, select, `to`) does not exist yet, or has the wrong kind.
4. Wrong parent kind. Goal: parent `""`, and only one Goal per log. DesignQuestion: parent is the Goal or a Hypothesis. Hypothesis: parent is a DesignQuestion. Experiment: parent is a Hypothesis. Result: parent is an Experiment.
5. An Experiment `targets` entry is not a Hypothesis with the same parent DesignQuestion as the Experiment's parent, or equals the Experiment's parent.
6. A `result.record` verdict key is not the Experiment's parent Hypothesis or one of its `targets`.
7. `result.retract` on a Result that is already retracted. `hypothesis.abandon` on a Hypothesis that is already abandoned.
8. `dq.select` of a Hypothesis that is not a direct child of that DQ, or is superseded or abandoned.
9. `confirm` whose `select` is not a `dq.select` on this DQ. Also refused: a `verifier` confirm whose actor equals the actor of that select (separation of duties).
10. `supersedes` where `from` = `to`, the two Hypotheses have different parent DQs, or the edge would create a supersedes cycle. **[added]** It may also join two Experiments, whose parent Hypotheses must have one parent DQ (`cross-DQ supersedes`); it is refused between an Experiment and any other kind (`wrong kind`), and onto an Experiment that has a non-retracted Result (`bad supersedes: <id> has the non-retracted Result <id>`).
11. `depends_on` whose `from` is not a Hypothesis or DQ, whose `to` is not a Hypothesis, where `from` = `to`, or where the edge would create a cycle in the **combined status graph** (see §7).
12. `edge.add` duplicating a live edge; `edge.remove` of an edge that is not live.
13. An `alias` that is already used.
14. `hypothesis.amend` on a Hypothesis that is superseded or abandoned (`bad amend: <id> is superseded or abandoned`), or whose `body` equals the current body (`bad amend: body is unchanged`).
15. A limit key in `closes` or `accepts` that does not name a limit (`bad limit ref`): the key is not `<ref>#<n>` with n ≥ 1; `<ref>` is not a Result; the Result is retracted; n is greater than the Result's number of limits; or the Result is not under the question. For `closes` the question is the one the new Result belongs to (the parent DQ of its Experiment's Hypothesis); for `accepts` it is the confirm's DQ.
16. Removed (ADR 0011). Nothing refuses a `dq.request_verify` beyond the envelope checks in rule 1.
17. A `confirm` with `accepts` whose role is not `owner` or whose verdict is not `sufficient` (`bad confirm: only a sufficient owner confirm accepts limits`).
18. A `result.record` `evidence_hashes` key that is not an entry of `evidence` or does not clean to `recipes/<name>@<version>/<file>` (`bad evidence_hashes: …`), or a value that is not lowercase hex (`invalid evidence_hashes value for …`).
19. `decided_by` on a node that is not an Experiment (`unknown field`), or a value other than `reading` (`invalid decided_by`).
20. `experiment.withdraw` on a node that is not an Experiment (`wrong kind`), on one already withdrawn (`already withdrawn: <id>`) or superseded (`already superseded: <id>`), or on one that has a non-retracted Result (`bad withdraw: <id> has the non-retracted Result <id>`). `hyp withdraw REF` writes `experiment.withdraw` for an Experiment and `dq.withdraw` for any other node.

CLI error text is `refused: <reason>` on stderr, with exit code 1. A help request (`-h` or `--help` on any command, and `help` at the top level and on a group command) prints usage on stdout and exits 0 at every level, while a missing or unknown subcommand exits 2.

### 5.1 Write-time refusals

Six more refusals come from the CLI before it appends, not from `Validate`. The first two read derived status, the next two read recipe files, and the last reads the text of the command line; `Validate` checks stored events only, so that a later change to a derivation, a recipe or a text rule cannot turn an old log into rejections. Replay accepts these events, so a log merged from another worktree can hold them.

- `hyp add exp` under Hypothesis H is refused when H's question has live rivals of H (§7.1) and `targets` names none of them, unless `--target none` is passed. Such an Experiment in a merged log still shows as `one-rival` in `hyp signals`.
- `hyp confirm --role owner --verdict sufficient` is refused while a limit under the question has `Gating` set (§8) and the confirm does not accept it. `hyp confirm` on a question with no current selection is refused too, since it has no select to name. `replay.Projection.AdmitConfirm` holds both.
- `hyp add exp` is refused when the Experiment has no measure and does not say its verdicts are decided by reading: its recipe is `none`, or a pinned recipe that is not an instrument (§7.6), and `--decided-by reading` is not passed. The refusal names `RECIPE-FORMAT.md`, Verdict scripts, as the way to a recipe. A pinned instrument needs no declaration, and `--decided-by` with one exits 2. Replay does not check the declaration: an Experiment with neither in a merged log still applies and keeps its status.
- `hyp record` under an Experiment whose recipe is not `none` is refused when that recipe no longer loads or no longer hashes as the Experiment's `recipe_hash`, or when an `--evidence` pointer into that recipe hashed differently from `recipe_hash` while recording (§7.6).
- `hyp record` under an Experiment that is `superseded` or `withdrawn` (§7.6) is refused, naming its replacement or saying it is withdrawn. Replay does not repeat the check: in a merged log such a Result still applies, the Experiment reads `run`, and `hyp check` warns `<E> is superseded|withdrawn but reads run`.
- Every write command is refused when its `--why`, `--body`, `--observation`, `--summary` or any `--limit` cites an id that is no node of the store. A cite is a token that starts at a word boundary with `G-`, `Q-`, `H-`, `E-` or `R-` and goes on with six upper-case Crockford base32 characters that end the word (a display id, §2), or with a placeholder mark: `<`, `{`, `…`, `...` or three or more `X`, as in `R-<new>`. A display id that resolves (§2) passes; a placeholder never does. Lower-case prose such as `E-mail` or `R-squared` is no cite, and a token right after `:` cites another store (`containment:R-ABC123`) and is not checked. The refusal names each flag and token, `--why names R-<new>, which is no node here`, appends nothing and exits 2, as `unknown node` does (§2). Text written before this rule, or by an older binary, replays unchanged.

`hyp add exp` also warns, without refusing, when its parent or a target is not a live rival (§7.1): it writes the Experiment, exits 0 and prints one `warn: D1:` line per such Hypothesis with its status, the Results that contradict it or the Hypotheses that supersede it. It reads the same rule as the D1 signal, so the warning and D1 name the same Hypotheses.

## 6. Merge order

- Read each `.tree/events/*.jsonl` file in its own line order. Within one actor file, append order is preserved even if `ts` goes backwards.
- Merge the files by a stable k-way merge: repeatedly take the head event with the smallest `id` (ULIDs embed time, so this approximates time order), breaking ties by actor name.
- `hyp check` warns, without rejecting, when an event references an id that sorts after it.
- `hyp sandbox create --store SRC --cut EVENT` writes a store cut in this order: the merged events up to and including the first with id `EVENT` (the `asOf` stop in §7), each line as written, so the cut store replays to the same state as `asOf = EVENT` on SRC.
- A line that does not parse has no id, so it is not in the merge. `hyp check` reports it as an error naming `file:line`; `hyp status` (with or without `--as-of`) replays the other lines, exits as usual, and prints `warn: skipped file:line: …` on stderr for each such line; `hyp sandbox create --store SRC --cut EVENT` refuses the store with exit 3, naming the first such `file:line`.

## 7. Derived state (forward pass)

Replay is a pure function: `Replay(events []Event, recipeHashes map[string]string, asOf string) Projection`. It reads no files and no clock. `recipeHashes` maps each recipe ref (`name@version`) to its current hash.

For each event, in merge order:

1. `err := Validate(state, ev)`. If there is an error, append a Rejection and continue.
2. Apply the event to the stored state: nodes, live edges, results, retracted set, abandoned set, DQ selection state, confirms, amended Hypothesis bodies and the DQ hold state.
3. Run `recompute(state)`, which recomputes every derived field below from scratch.
4. If `ev` is a **basis event** (defined in §7.3), store its snapshot, then run `recompute(state)` again.
5. For every DQ that is now `answered`, set `everAnswered[dq] = true`.

Stop after the event whose id equals `asOf`, when `asOf` is not empty.

### 7.1 Hypothesis status (first matching rule wins)

1. `superseded`: a live `supersedes` edge points at it.
2. `abandoned`: it has an applied `hypothesis.abandon`.
3. Collect the verdicts for this Hypothesis from all **non-retracted** Results. No verdicts gives `proposed`. Any `contradicts` gives `contradicted`. Otherwise any `supports` gives `supported`. Otherwise it is `inconclusive`.
4. **[added]** If rule 3 gave `supported` and any sub-question that is not withdrawn is not `answered` (deferred children still block), the status is `awaiting_subquestions`. Everywhere this spec says `supported`, it means exactly `supported`, never `awaiting_subquestions`.

A Hypothesis is **live** unless it is `superseded` or `abandoned`. A DQ is live unless it is withdrawn.

A Hypothesis is a **live rival** when it is live and not `contradicted`. The untargeted-rival refusal (§5.1), the one-rival and D1 signals count live rivals. Staleness (§7.3) and the trade-off candidates (§7.4) count live Hypotheses, contradicted ones included.

### 7.2 Combined status graph (for cycle refusal)

Directed edges: Hypothesis → each of its sub-questions; DQ → each of its child Hypotheses; N → T for each live `depends_on`. A new `depends_on` that would make this graph cyclic is refused (§5 rule 11). Because of this, derived fields can always be computed by memoized recursion without looping.

### 7.3 Stale (snapshot rule, the minimal form of truth maintenance)

- Stale is computed only for **live** Hypotheses and live DQs. Every other node reports `stale = none`.
- **Basis events for node N:** a non-retracted `result.record` with a `supports` verdict for N (when N is a Hypothesis); `dq.select` on N (when N is a DQ); `reaffirm` on N; and `edge.add depends_on` whose `from` is N.
- **Snapshot:** after a basis event for N, for each live `depends_on` target T of N, store `snap[N][eventID][T] = (T.status, T.stale != none)`.
- **Effective snapshot for edge N→T:** take the latest basis event of N, in merge order, that is still in force and whose snapshot includes T. A `result.record` stops being in force once retracted. Retracting therefore falls back to the previous snapshot, which is what makes "retract restores the earlier statuses exactly" hold.
- **N's stale cause** comes from the first target T, in edge-add order, that meets one of these:
  - `T.status != supported` gives cause `{Node: T, Reason: T.status}`;
  - `T.stale != none` gives cause `{Node: T, Reason: "stale", Next: T.staleCause}`;
  - an effective snapshot exists and is not `(supported, false)` gives cause `{Node: T, Reason: "changed_since_basis"}`.
- If nothing matches, stale is `none`.
- `reaffirm` creates a new snapshot. It cannot clear staleness while T is still unsupported, because the first test is always live; `hyp check` warns "reaffirm had no effect" in that case.
- `depends_on` is never followed across `supersedes`. A superseded target is not `supported`, so its dependents go stale until the agent re-points the edge and reaffirms.
- Stale is returned as a cause chain, not a bool, so the status screen can print `H-5 stale ← H-3 contradicted ← R-9`. The chain ends at the Result that contradicted (the latest non-retracted contradicting Result), when there is one.

### 7.4 DesignQuestion state

Hold state is the latest of `dq.defer` / `dq.withdraw` / `dq.resume`; resume clears it.

One record, the question's `Decision`, owns everything below except the hold state: select history, conflicts, verifier and owner confirms in one list per select (each with the Dossier at its place in the log), the owner requirement, limit acceptances and gating (§8), the reason order and the verifier-loop chain. Render, signals, check and the CLI read it.

Selection is the current selection id (`""` when there is none) plus a conflict set:
- Applying `dq.select` or `dq.unselect` whose `prev_select` ≠ the current selection id adds both ids to the conflict set. The event still applies; the last one in merge order becomes current.
- A later select or unselect whose `prev_select` equals the current id clears the conflict set.
- The **owner requirement** is derived from the current state and stored nowhere: the question has two or more trade-off candidates now (live child Hypotheses that are `supported` and not stale, as for the suggested candidate below), or the current select's `tradeoff` or `scope_change` is true. A stale supported sibling does not count.

Dossier: the evidence a question's current select rests on (ADR 0011). It is derived per question, from stored events only, and never stored: the selected Hypothesis's body, the set of live child Hypotheses, every non-retracted Result with a verdict on one of them with that verdict (found by verdict, not by tree parent), and the live links going out from those Hypotheses and Results, not counting `external_ref`. A rival's body, its planned Experiments and links other nodes add pointing in stay out. `hyp show <Q>` lists it under `dossier:`.

A `contradicts` is **deciding** when the Hypothesis it judges is not the selected one and, by §7.1, that Hypothesis would be `supported` without this one verdict. `hyp show <Q>` marks its Result as deciding the owner requirement. It is a mark: §7.1 still stands and no status reads it.

Confirms: a confirm judged the Dossier as it stood at the confirm's place in the merged log. For each (select id, role) pair, the confirm that **counts** is the latest one whose Dossier equals the current Dossier; when there is none, no confirm counts. A confirm that no longer matches has **lapsed**, and it counts again if the Dossier returns to what it judged, which is the §7.3 rule applied to a confirm, so a retract restores the earlier state exactly. This holds for both roles and both verdicts, and it is the only way a confirm stops counting. `hyp show <Q>` prints `changed since <actor>: …` for each role whose latest confirm has lapsed, naming what the Dossier gained (`+`) and lost (`-`) and `body` for an amended selected body. A lapse adds no status word. A reopened question closes by appending: a fresh confirm when the claim is unchanged, or a new Hypothesis that `supersedes` the old one and a select of it when the claim changed.

The state is decided by the first matching rule:
1. `withdrawn`, then `deferred` (from the hold state).
2. `answered` when all of these hold: the conflict set is empty; there is a current selection; the selected Hypothesis is `supported` and not stale; the DQ itself is not stale; the verifier confirm that counts on the current select is `sufficient`; and, if the owner requirement holds, the owner confirm that counts on it is `sufficient`.
3. Otherwise the state is `reopened` if `everAnswered`, else `open`. The reason is the first that holds:
   `conflicting_selections`, `selected_not_supported`, `selected_stale` (selected Hypothesis or the DQ is stale), `verifier_insufficient` (the verifier confirm that counts on the current select is `insufficient`), `awaiting_verifier` (no verifier confirm counts on it), `awaiting_owner`, `no_candidates` (no live child Hypothesis), `untested` (every live child is `proposed`), `all_contradicted` (every live child is `contradicted`), `inconclusive` (no live child is `supported`), `supported_unselected`.

The verifier-loop chain (`Decision.Chain`), which the verifier-loop signal reads, holds the insufficient verifier verdicts in a row on the current select. A sufficient verdict, an unselect, a select that does not continue the current one (the same Hypothesis or one that supersedes it) and a `hypothesis.amend` of the selected Hypothesis clear it, so verdicts on an old body never count towards a loop. The signal fires at two or more.

**Suggested candidate:** the trade-off candidates are the live children that are `supported` and not stale. When there is exactly one, suggest it. With two or more such children, report "trade-off: owner decides" and list them.

### 7.5 Goal

- `complete` when every live DQ whose parent is the Goal is `answered`. Sub-questions are covered through rule 7.1.4, so sub-questions under rejected rival hypotheses never block the Goal. (This clarifies the advice's "nothing stale" clause: stale on the selected path already prevents `answered`.)
- Otherwise `in_progress`, with counts per DQ state over all DQs and a count of stale live nodes.

### 7.6 Experiments and Results

- An Experiment's status, first matching rule wins (ADR 0008):
  1. `run`: it has a non-retracted Result. A live Result beats a close.
  2. `superseded`: a live `supersedes` edge from another Experiment points at it.
  3. `withdrawn`: it has an applied `experiment.withdraw`.
  4. `planned`.

  `superseded` and `withdrawn` are **closed**. Removing the `supersedes` edge reopens an Experiment. A closed Experiment's round in `hyp signals` is neither pending nor wasted and keeps its D1 and one-rival marks; `hyp status` lists no line for it, and `hyp show` prints `status: superseded (replaced by E-…)` or `status: withdrawn`. Question and Goal states do not read Experiment status, so a close changes neither.
- A **pin** is a recipe ref (`name@version`) and the hash an event recorded for it at write time. A bundle is its ref and hash; a `recipes/<name>@<version>/` directory name is only the library's lookup key, and hyp opens each copy it makes by the ref it holds ([ADR 0005](adr/0005-a-bundle-is-its-ref-and-hash.md)). A Result depends on its Experiment's pin (`recipe` and `recipe_hash`, unless the recipe is `none`) and on one pin per entry of its own `evidence_hashes`, whose ref is the pointer's first segment under `recipes/` after cleaning the path. A Result is `questionable_evidence` when, for any pin it depends on, `recipeHashes[ref]` is missing or differs from the pinned hash. A new version of a recipe is a new ref, so it never flags Results recorded under an older one. `recipeHashes` leaves out a bundle that cannot be hashed (a symlink, a file or directory name outside `[A-Za-z0-9._-]`, or a `recipe.md` whose name or version does not match its directory; a name below a known answer's `known-answers/NAME/input/` may also hold `@`, so an input can hold a store with its pinned bundles) and a ref whose copies hash differently, and `hyp check` names each one. A bundle that hashes but does not open stays in `recipeHashes`, so its pins verify, and `hyp check` names it too (§9.2). This is a flag only and never changes any status.
- **[added]** An evidence pin records the bundle as it was when the Result was written, not when its files were read; an edit made in between and undone before recording is not detected, and for refs other than the Experiment's own, neither is one still in place. `hyp record` refuses a Result whose Experiment's recipe no longer hashes as the Experiment pinned, or whose evidence pin for that recipe differs from the Experiment's `recipe_hash`, so a Result never pins two hashes for one ref. Replay does not repeat that check: such an event still applies and reads as questionable.
- **[added]** A matching hash says the procedure is the same, not that it gives the right answer. An **instrument** is a recipe bundle with an Evidence or Standalone command, the checks a verdict counts. A script whose output a Result's verdict counts lives in a bundle and runs as one of its commands (§9), never outside `recipes/`. An instrument must have known answers: cases under `known-answers/NAME/`, each an `input/` and an `expect.json` (`RECIPE-FORMAT.md`, Known answers). A check whose cases print `verdict: LABEL=WORD` lines is a verdict check, and `hyp recipe check` warns when a label has no case printing `supports` or none printing `contradicts`, on a malformed verdict line and on a label printed twice in one case; nothing is refused (ADR 0004). Before every `hyp sandbox create` (for each stacked recipe, in stack order, before anything is made), `hyp recipe run` and `hyp recipe evidence`, hyp runs the instrument's known answers with the operator's hyp; a failed case refuses the run with exit 1, a timed-out case exits 3, and a refused create leaves no sandbox. `hyp add exp --recipe` refuses an instrument with no case. A **measured** Experiment is one whose harness recipe is an instrument: its pinned recipe's commands hold the measure, and `recipe.md` declares nothing else. An Experiment with no measure, `none` or a pinned recipe that is not an instrument, declares `decided_by: reading` (§5.1); an environment recipe is never pinned, so the rule never asks it. Replay never reads the declaration, and a false `--decided-by reading` that hides a measure is not seen. The sandbox record and the command record both hold the outcome of each case, as `known_answers` (§9.4). A bundle without an Evidence or Standalone command is not gated (ADR 0002).
- **[added]** A Result is **uncontained** when hyp cannot vouch for its verdict script (ADR 0002). Each reason is reported once, in this order:
  1. it cites a script outside a versioned bundle (`recipes/<name>@<version>/`): a regular file that ends in `.py` or `.sh`, has any execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`, including one under `recipes/` in no versioned bundle, which hyp neither pins nor runs, and one inside a bundle that its hash skips (a dot-entry or `__pycache__`), which no pin covers;
  2. it has a pin (its Experiment's recipe or an evidence pin) to a bundle with an Evidence or Standalone command, and cites no `command.json` that names that ref at that hash, records at least one command, every one of which exited 0, and holds `known_answers` with at least one case, every one `ok`;
  3. it cites a `command.json` that hyp cannot read, or that names no ref or hash, or whose ref and hash match none of its pins.

  A pin to a bundle without an Evidence or Standalone command, such as a known-answer fixture, is held by its hash alone (§7.3). The CLI derives the mark from files when it reads, as it does `recipeHashes`. Replay never reads files, and the mark never changes a status. `hyp record` warns, and `hyp show`, `hyp status`, `hyp render` and `hyp check` show it. Nothing is refused. Limits: a script named only in prose, whose output a Result cites, is not seen; neither is a hand-made output file cited beside a real command record, or a script with none of the marks in reason 1. A hand-written `command.json` that claims a clean run with passing known answers satisfies reason 2; a verifier checks the record's `command.log` and `known-answers.log`. A script that only a Launch, Doctor, Drive or Cleanup command reaches, cited from another Experiment, is held by its hash alone: hyp runs it in a sandbox, but no reason ties that sandbox's record to the Result. The gate proves that a bundle reproduces its own expected outputs, not that those outputs are right: a weak or wrong case needs review (ADR 0002). A case and a run get the same environment, working directory and hyp, and a case compares exit code, stdout and the tree under `$HYP_OUT` (ADR 0003), so a passed gate carries over to the run, except for what a case cannot see: a read of a fixed absolute path, a read from the network, the versions of the tools on the host `PATH` after hyp, and a write outside `$HYP_OUT`; all pass the double run.

## 8. Projection (what replay returns)

```go
type Projection struct {
    Nodes      map[string]*Node      // static fields from node.add / result.record
    Status     map[string]NodeStatus // derived; see below
    Goal       GoalStatus
    Rejections []Rejection
    Applied    []Event              // events that applied, in merge order
    LastCheckpoint *Event
    RivalsAtAdd map[string][]string  // Experiment id → the live rivals (§7.1) of its parent just before its add
    Exps       map[string]*Exp       // Experiment id → its derived record, built once per projection
    Results    map[string]*Res       // Result id → its derived record, built once per projection
    Edges      []Edge                // every edge in apply order; Live is false once removed
    Decisions  map[string]*Decision  // DQ id → its Decision record (§7.4)
}
type Decision struct {
    Current   string                 // the current select id, or ""
    Conflicts []string
    Selects   map[string]SelectRec   // select id → its record
    Confirms  map[string][]ConfirmRec // select id → its confirms in log order, each with the Dossier at that point
    Chain     []Verdict              // insufficient verifier verdicts on the current select
    Limits    []Limit                // the limits recorded under the question, in Result creation order
    Dossier   *Dossier               // of the current select (§7.4); nil when there is none
}
type Dossier struct {
    Items    []DossierItem           // sorted; two Dossiers hold the same facts when their Items are equal
    Deciding map[string][]string     // rival Hypothesis id → Results whose contradicts is deciding (§7.4)
}
type DossierItem struct { Kind, Node, To, Word, Body string } // body, hypothesis, result or link
type Exp struct {
    Status     string            // planned, run, superseded or withdrawn (§7.6); equal to Status[id].Status
    Closed     string            // superseded or withdrawn when closed, even if a Result makes it run; "" otherwise
    Live       []string          // non-retracted Results, in creation order
    Judged     map[string]string // Hypothesis id → verdict over Live; a later Result wins
    OneRival   bool              // its parent had live rivals at its add and its targets name none of them
    Unjudged   []string          // targets live at its add that no Live Result judged; nil while Live is empty
    ReplacedBy []string          // Experiments with a live supersedes edge to it, in edge order
}
type Res struct {
    Status       string   // recorded or retracted; equal to Status[id].Status
    Questionable bool     // questionable_evidence (§7.6)
    Pins         []Pin    // its Experiment's pin, then one per evidence_hashes entry, in the order cited (§7.6)
    Uncontained  []string // reasons (§7.6); set by the CLI from files, not by replay
}
type Limit struct {
    Key, Result, Text string   // Key is "<result-id>#<n>", full ULID
    Hyps  []string             // Hypotheses the Result has a verdict for, sorted
    State string               // "open", "closed" or "accepted"
    By    string               // the closing Result id, or the accepting owner's actor
    Gating bool                // open, and Hyps contains the Hypothesis of the question's current select
}
type NodeStatus struct {
    Status     string      // hypothesis status, DQ state, experiment status (§7.6), result recorded/retracted, goal state
    Reason     string      // DQ open/reopened reason; "" otherwise
    Stale      *StaleCause // nil when not stale
    Suggested  []string    // DQ only
}
type StaleCause struct { Node, Reason string; Next *StaleCause }
type Edge struct { ID, From, Type, To string; Live bool } // ID is the edge.add event's id
```

`Confirms` keeps every applied confirm, including one that has lapsed, so `hyp show` still prints the latest. It prints the verifier confirm that counts, or the latest when none does.

Limits **[added]**: a Result's limits are filed under its question, the parent DQ of its Experiment's Hypothesis, whichever Hypotheses it judged. Limits of a retracted Result are left out. A limit is:
- `closed` when a non-retracted Result lists its key in `closes`; `By` is the first such Result in creation order. Retracting that Result reopens the limit unless another one closes it.
- otherwise `accepted` when a sufficient owner confirm on the question's **current** select, with the current Dossier, accepted it, and no later owner `insufficient` has that Dossier; `By` is that owner's actor. A new select, a Dossier change or an insufficient owner confirm at this Dossier makes the limit open again, and a Dossier that returns restores it.
- otherwise `open`.

Limits never change a derived status. Only the CLI reads them: `hyp show` and `hyp status` list them, and `hyp confirm --role owner --verdict sufficient` is refused while a limit has `Gating` set (it is open and its Result has a verdict for the Hypothesis of the current select) and the confirm does not accept it (§5.1). An open limit whose Result judged only other Hypotheses does not gate; it stays listed, and `hyp show <Q>` marks it `open, does not gate (judges …)`.

Rivals at add **[added]**: when an Experiment is applied, replay stores the live rivals of its parent Hypothesis under the same question, read from the derived state just before that event. A target is **unjudged** when it is in that list and no non-retracted Result of the Experiment has a verdict for it; an Experiment with no non-retracted Result has no unjudged target. An Experiment is **one-rival** when that list is not empty and its `targets` name none of it. Both are read at the Experiment's add, so a rival that dies later does not clear a mark, and a rival added later does not create one. `hyp signals` reports an unjudged target, and `hyp status` and `hyp show <Q>` mark the Experiment. They never change a derived status.

The Projection holds every fact hyp derives; render, signals and the CLI read its fields, and nothing stores them a second time (ADR 0006). Scripts read the applied events and rejections through `hyp events --json` and derived signals through `hyp signals --json`, never the files under `.tree/events`.

Equality test used by acceptance: `reflect.DeepEqual(p1.Status, p2.Status) && reflect.DeepEqual(p1.Goal, p2.Goal)`.

## 9. Sandbox lifecycle and steps

This section is the authority for when hyp runs a recipe's commands and whether a sandbox exists when it does. The words are defined in [glossary.md](glossary.md), and the decision behind the section is [ADR 0001](adr/0001-hyp-runs-every-recipe-command.md). `skills/hyp-verification/RECIPE-FORMAT.md` says how to write a recipe that follows it.

### 9.1 Steps and sections

A bundle's `recipe.md` has one level-2 section per step (Launch, Doctor, Drive, Evidence, Cleanup) and may have a Standalone section, which is not a step. A line in one of these sections that starts with `Run:`, outside a code fence, is a **command**: one script inside the bundle followed by fixed words. The section a command sits under decides when hyp runs it and whether a sandbox exists then. A command carries no other label, except that a Standalone command is named: `Run NAME: ./script WORDS`, with NAME matching `[a-z][a-z0-9-]*` and used once in the file. A plain `Run:` line under Standalone, a named line under any other heading, and a `Run:` line under a heading that is not one of the six are refused.

| Section | Does a sandbox exist? | When hyp runs it |
|---|---|---|
| Launch | yes | `hyp sandbox create`, for each recipe in stack order |
| Doctor | yes | `hyp sandbox create`, right after the same recipe's Launch |
| Drive | yes | `hyp sandbox drive ID`, between create and remove, for each recipe in stack order; it can be run again |
| Cleanup | yes | `hyp sandbox remove`, the last recipe first; also `hyp sandbox create`, after a failed Launch or Doctor |
| Evidence | no: it reads the kept files after remove | `hyp recipe evidence --record DIR REF OUT_DIR`, after remove, the first time and again on demand, also from a later version of the recipe; `hyp sandbox remove` never runs it. An OUT_DIR without `sandbox.json` exits 2, and the refusal names a Standalone command (`hyp recipe run`) for files no sandbox wrote |
| Standalone | no sandbox at all | `hyp recipe run --record DIR REF NAME ARGS` |

Evidence that the operator gathers by hand while the sandbox exists is not a command, and is unchanged: it is the files saved under `.tree/evidence/<E-id>/`. Scripted Evidence is the same step done by a command, which is why it runs after Cleanup: it reads only the kept files, so it can be run again.

A sandbox is a directory plus its record, not process isolation. Each command hyp starts leads its own process group, which its row records as `pgid` (ADR 0002). After Cleanup, `hyp sandbox remove` stops the groups of the Launch, Doctor, Drive and Cleanup commands that still have members: SIGTERM, up to 5 seconds, then SIGKILL. A timeout or Ctrl-C kills the running command's group (its row says `interrupted` for Ctrl-C), and an Evidence or Standalone group ends with its command; create still runs Cleanup after a Ctrl-C. Limits: a daemon that leaves its group, as Postgres's postmaster does by calling `setsid`, still needs its recipe's Cleanup, and the leftover check still applies; a group id reused after its group ended is not told apart; what a check cannot see is listed under §7.6 and ADR 0003. "Runs in a sandbox" means only that the command runs while the sandbox exists and gets that sandbox's `HYP_SANDBOX_*` variables (§9.3).

### 9.2 hyp runs every command

hyp runs every command in a bundle, and the operator runs none of a bundle's scripts directly. A command is what a `Run:` line names; every other file in the bundle, a helper script included, is hashed content (ADR 0002). Hashing never parses: a bundle whose `recipe.md` breaks a command rule, such as a `Run:` line outside a section or shell syntax, still hashes, so every pin to it still verifies, and `hyp check` names it as a bundle that does not open. `hyp recipe new` writes `format: 1` into every bundle it makes, a first version from its embedded template (`hyp recipe new NAME@1`) as well as a copy (`--from`), the one recipe format, and the line is a required field. Every use of a bundle without it is refused with exit 1, reading only its frontmatter: `hyp add exp`, `hyp sandbox create`, `hyp recipe run`, `hyp recipe evidence`, `hyp recipe known-answers`, and `hyp record --evidence` into such a bundle. The refusal names `hyp recipe new` as the way to copy it. `hyp recipe hash` still hashes it (ADR 0002, ADR 0007). `hyp recipe check` names each versioned bundle `frozen` when a tree pins its current hash, so its bytes never change, and `draft` when none does. When the repository root holds `go.mod`, `hyp recipe check` warns about each `.go` file in a versioned bundle, and `hyp add exp --recipe` about each one in the bundle it pins, because `go vet ./...` reads a `recipes/<name>@<N>/` directory holding one as a package outside the main module; the warning suggests a non-`.go` name, such as `x.go.txt`, and both commands still exit 0.

A recipe is self-contained when it relies on nothing but its own files, hyp and the subject, wherever its commands run. Self-contained does not mean "runs in the sandbox": a Standalone or scripted Evidence command is self-contained and runs with no sandbox. A fixed input that is not the subject, such as a fixture or a cases file, is a file of the bundle, so the recipe pin hashes it, and the code that reads it is in the bundle too. A file outside the bundle is the subject, or a read or write that the command's `External:` line names (ADR 0008, owner decision on 2026-11-01). hyp does not detect a fixture outside the bundle.

Each Evidence and Standalone command states its reach on one line `External: TEXT` directly under its `Run:` line in `recipe.md`: the subject arguments and any outside reads or writes, or `none`. The line is no field `recipe.Load` requires, so a frozen bundle without it still loads and records. `hyp recipe check` warns `<command>: no External: line` for a command without one and `<command>: External: not stated (TODO)` for one whose text is empty or `TODO`, and exits as before. `hyp recipe new` writes `External: TODO` under each such command of a copy that has no line, and under the Standalone command of a first version. Every line of a first version's template that its author must replace holds the marker `hyp:placeholder`, and `hyp add exp --recipe` refuses (exit 1) a bundle that still holds it in any file outside a known answer's `input/`, naming the files (ADR 0009).

hyp's Go learns no replay concepts. It runs a script with its arguments and records the exit code and the output. hyp is not a job runner: no retries, no background jobs and no stdin streaming.

### 9.3 How a command runs

- No shell runs a command. Its first word is a relative path containing `/` to an executable script inside the bundle, and the rest are plain words; quotes, `$`, pipes, redirects and globs are refused.
- Launch, Doctor and Cleanup run with the recipe's copy in the sandbox as their working directory, and get `HYP_SANDBOX_ID`, `HYP_SANDBOX_WT`, `HYP_SANDBOX_TMP`, `HYP_SANDBOX_RECIPE_TMP`, `HYP_SANDBOX_INPUTS` (the sandbox's copy of `--inputs FILE`, when given), `HYP_SANDBOX_PATCHES` (the directory of the sandbox's copies of each `--patch ARM=FILE`, `<ARM>.patch` each, when any is given) and a `PATH` that starts with the sandbox hyp. They also inherit the environment of the process that runs hyp.
- Drive commands run like Launch: in the recipe's copy, with the sandbox's variables. They take no arguments from the operator; a Drive reads per-run values from the inputs file or from the state its Launch wrote.
- Evidence and Standalone commands are checks, and run exactly as their known-answer cases do (ADR 0003). They get no `HYP_SANDBOX_*` variables, because no sandbox exists when they run, and inherit nothing else of the caller's environment: only `PATH` (the copy of the running hyp the gate used, first, then the caller's `PATH`), a scratch `HOME` and `TMPDIR`, `GIT_CONFIG_NOSYSTEM=1` and `HYP_OUT`, an empty directory for the files the check leaves; stdin is `/dev/null`. They run from a temporary copy of the bundle whose hash equals the library's, deleted afterwards, in an empty working directory of their own. A hyp killed while it runs leaves that copy (`hyp-run-*` under the system temp directory): hyp never reads it again and does not clean it up, the system's temp cleaning removes it, and the operator may delete it. An Evidence command gets the absolute OUT_DIR appended as its last argument; a Standalone command gets the caller's ARGS appended after its fixed words, each file argument replaced by a copy and each directory argument made absolute (see Kept inputs in §9.4).
- Launch and Doctor together have 10 minutes for each recipe, and each Cleanup command has 2 minutes. Each Drive command has 60 minutes, each Evidence command 10 minutes, and a Standalone command 60 minutes (§9.5).
- When hyp runs a Drive, Evidence or Standalone command for the operator and the script exits non-zero, hyp stops there and exits with that code unchanged, so a check that exits 30 can be told apart from hyp's own 1 (refusal), 2 (bad arguments) and 3 (environment). A command that times out or is killed by a signal has no exit code, and hyp exits 3. hyp's own refusals come before any command starts and leave no row, so the record always tells a script's exit from hyp's; a script whose code alone must tell it apart exits 10 or above.

### 9.4 Records

- **Sandbox record.** `hyp sandbox create` writes `sandbox.json`, and only hyp writes it; `hyp sandbox remove` copies it to OUT_DIR. Each command's output goes to `recipe.log` after a `== REF step: command` line. The fields are listed under "The sandbox record" in `RECIPE-FORMAT.md`. It names no host path for anything a replay needs: `host` is the host's directory name, a store inside the host is `<host>/REL`, and the sandbox's clone, scratch, `bin/`, inputs copy and recipe copies are derived from its id, not recorded (format 3). `patches` names each Arm's patch by its sha256 (§9.6). Its `commands` list holds each command hyp ran in the sandbox (Launch, Doctor, Drive, Cleanup), in order, with its step, its arguments and its exit code; `hyp sandbox drive` adds to it.
- **Command record.** For each run of Evidence or Standalone commands, hyp writes a command record in the new directory that `--record DIR` names (refused with exit 2 if it exists, is or resolves inside a `recipes/` directory the recipe lookup searches, or, for Evidence, inside OUT_DIR, comparing directories by identity so symlinks and other spellings do not get past it; with exit 3 if its path cannot be resolved): `command.json`, naming the recipe ref, its hash, the library path and the working directory (each `<repo>/REL`, see Kept inputs), the step, with one row per command (its arguments as shown, its exit code or why it has none, and when it ran) and `inputs`, each file an operator argument names; `command.log`, the output after `== REF step: command` lines; and each check's product, its stdout in `stdout` and its `$HYP_OUT` tree in `out/` (Standalone), or `stdout/<n>` and `out/<n>/` (Evidence command `n`), kept even when the check fails (ADR 0003). The fields are listed under "The command record" in `RECIPE-FORMAT.md`.
- **Kept inputs.** An operator argument of `hyp recipe run` or `hyp recipe evidence`, alone or as the VALUE of a `KEY=VALUE` word, is a file argument when it is written as a path (absolute, or relative with a `/`; a relative one is read against the caller's directory) and names a regular file, symlinks followed. A bare word is never one, even when a file of that name exists. Before the first command runs, hyp copies each file argument N (ARGS counted from 1) under the run's temporary directory, which the script gets in its place, hashes that copy, and keeps a second copy beside the record at `DIR/inputs/N/<base name>`, so a script that writes to its argument changes neither the original nor the kept bytes. A file over 1 MB is not kept; its sha256 and size are. `command.json` (format 4) lists each in `inputs` (`arg`, `path`, `sha256`, `size`, `kept`), and the row shows `inputs/N/<base name>` in its place. A file inside the tree is copied too. A directory argument (such as Evidence's OUT_DIR) is passed as an absolute path and shown as `<repo>/REL`, `<repo>` being the directory name of the caller's git work tree (or of the caller's directory outside one); one outside it is shown as given, and hyp warns that a replay cannot reach it. `path` and `cwd` are written the same way. So neither the record nor the `==` lines name a host path for anything a replay needs. The inputs are bound at run time, not when the Experiment is added (ADR 0008). Limits: a file the command reads without being given it is not kept; its `External:` line names it. A format-3 record's `arg_sha256` (each absolute file argument's sha256) is still read.
- **Known answers.** Each recipe entry of `sandbox.json`, and `command.json`, holds `known_answers`: `hyp_sha256`, the hyp that ran the cases, and `cases`, one `name` and `outcome` per case (§7.6). Each case's stderr goes to `known-answers.log` after a `== REF known answer NAME` line: in the sandbox, kept in OUT_DIR on remove, and in the command record's directory (ADR 0002).
- **Format.** `sandbox.json` and `command.json` each carry `"format"`, the shape number of that file (ADR 0002). A golden test per file fails `go test` when its shape changes without a new number, so a reader can compare the number before reading the rest. A record without the key was written before it.
- **Changes.** When create finishes, hyp writes `manifest.json` in the sandbox, a list of every path in `wt/` with its size, mode and hash. `hyp sandbox remove` diffs `wt/` against it and writes `changes.json` (format 1) to OUT_DIR, with the added and modified files up to 10 MB under `changes/`. The fields are listed under "Changes" in `RECIPE-FORMAT.md`.
- **Writes.** hyp writes `sandbox.json` and `command.json` to a temporary file in the same directory and renames it over the record, so a reader sees the old record or the new one, never part of one. A hyp killed between the two leaves a `.sandbox.json-*` or `.command.json-*` file that nothing reads; a sandbox's goes with it at remove.
- **Rows.** hyp writes each command's row before the command starts and rewrites it when the command ends. A row with no `ended` (and so no `exit` or `error`) is a command hyp stopped during: hyp was killed, or the machine stopped. Its output so far is in the log. When the write before a command fails, the command does not run and hyp exits 3, except under Cleanup, which runs every command anyway and reports the failed write as a failed Cleanup. A remove refused for a leftover process keeps the Cleanup rows it ran, and each recipe's `cleanup`, in the sandbox's record.
- **One hyp command per sandbox.** `hyp sandbox create`, `drive` and `remove` hold a lock on `sandbox.lock` in the sandbox while they run. A second one on the same sandbox is refused with exit 3, naming the hyp command that holds it, and changes nothing. The lock goes away when its hyp exits, however it exits. `hyp recipe evidence` and `hyp recipe run` take no lock, because each writes only the new directory it made.

### 9.5 Choices

The open choices were decided in a hyp tree, whose records stay in the private development history, and while building.

1. How a Standalone command is named: an inline label, `Run NAME: ./script WORDS`. The owner picked it over a heading; both resolved every real use.
2. How the operator hyp is checked: by a Doctor command in the harness recipe, so hyp records no operator fields. Accepted limits: an Experiment under `--recipe none` gets no check, and each recipe that needs it carries the check.
3. One recipe format (ADR 0002). The reach test, which froze a bundle holding a script no command named, was deleted: the known-answer gate proves what a verdict script gives, while a reach test only proved that a script was named. Every file other than a command's script is hashed content, hashing never parses, and a bundle is `frozen` once a tree pins it and `draft` before.
4. The Drive command is `hyp sandbox drive ID`. The name is the step, as create runs Launch and Doctor and remove runs Cleanup. With no operator arguments, the recipe's fixed words and its inputs decide a drive.
5. `hyp sandbox remove` does not run scripted Evidence; `hyp recipe evidence` runs it, the first time and every time. One path serves the first run and every re-run, a later version is only another REF, and remove's own exit codes never mix with a check's.
6. Timeouts: Drive and Standalone 60 minutes per command, because the longest kept create-to-remove was 11.6 minutes and a re-test script drives sandboxes itself; Evidence 10 minutes per command, because it only reads kept files.
7. The command record is a directory the operator names with `--record`, so it can sit under `.tree/evidence/<E-id>/` and be cited, and large output stays in its log. The commands run from a temporary copy, so a script cannot change the library version it ran from.
8. (replaced by choice 3)
9. A busy sandbox is refused, not waited for. A Drive can run 60 minutes, a waiting hyp would hang its caller silently, and queueing is job-runner work §9.2 rules out. The lock is an flock inside the sandbox, so the kernel frees it when hyp dies and it is deleted with the sandbox.
10. An interrupted command leaves a row with no `ended`. Every command hyp runs is a row (§9.2), and only a row written before the command starts survives a kill. A finished row is unchanged.
11. hyp does not clean up a killed run's temporary copy. It holds nothing the library does not, and a sweep of another hyp's directories would be a reaper.
12. A record DIR inside OUT_DIR is refused. The check reads OUT_DIR, its record is written before each command starts, and a re-run must read the same files as the first run. A record beside OUT_DIR, the documented place, is unaffected.
13. A record DIR inside any `recipes/` directory the lookup searches is refused, not only one inside the ref's own bundle. A record changes the hash of whatever bundle it lands in, an outer copy of the ref must hash the same as the nearest, and a record in a `recipes/` directory that does not exist yet would start a library the lookup finds.
14. (replaced by choice 3)

### 9.6 The subject

A sandbox is the only place hyp assembles a subject. `hyp sandbox create` builds it from the host at `--base`, the store cut when one is given, and the per-run inputs. It also builds it from one patch for each Arm (see Arm in [glossary.md](glossary.md)), given as `--patch ARM=FILE`.

- **A change that lands in code is a commit.** An Experiment about it uses that commit as `--base` and takes no patch.
- **A patch is a throwaway Arm: a change over the base that never lands in code.** An Arm has at most one patch. A patch is never versioned, never stacked on another patch and never rebased. An Arm worth keeping is merged and becomes a commit, and the next Experiment uses that commit as its base. hyp refuses a patch that does not apply to the base (`git apply --check` in the clone) before any command runs, with exit 2, and leaves no sandbox. It applies nothing: Launch builds each Arm. It records each patch's Arm and sha256 under `patches` in `sandbox.json`, copies it to `patches/<ARM>.patch` in the sandbox, passes that directory as `HYP_SANDBOX_PATCHES`, and `hyp sandbox remove` keeps it as `OUT_DIR/patches/<ARM>.patch`, so where the operator wrote it does not matter. A patch is the subject, so it is not a file of the recipe bundle (see Self-contained in [glossary.md](glossary.md)).
- **Code under test is built in Launch, from the subject the sandbox assembled.** Each binary it builds is named by its sha256, and only the sha256 is kept. A check does not take a binary under test from outside a sandbox, and a known-answer case cannot build one: it has 60 seconds and a scratch `HOME` by design.

The rule exists so a record names its subject only by what the repository holds (the base commit) or the tree keeps (each patch and its sha256), and a replay can rebuild the subject from them. A base commit holds only while the repository's history keeps it.
