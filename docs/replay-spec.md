# Spec: events, refusals and replay rules

This file is the single authority for the event schema, the refusals and the derived-status rules of `hyp`; [sandbox-spec.md](sandbox-spec.md) owns the sandbox. Plans, skills and Go comments point here instead of restating the rules; Go comments cite it by section as "spec §N".

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
- Recipes: `recipes/<name>@<version>/`, in the directory that contains `.tree/` or in any parent up to the repository root (the first directory holding `.git`). See `skills/hyp-verification/RECIPE-FORMAT.md`. [sandbox-spec.md](sandbox-spec.md) §1 says when hyp runs a recipe's commands and whether a sandbox exists then.
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
| `node.add` | the new id (= event id) | `kind` (Goal, DesignQuestion, Hypothesis, Experiment), `parent` (id; `""` only for Goal), `title`, `body` (optional), `alias` (optional), `recipe` (Experiment only, required: `"name@version"` or `"none"`), `recipe_hash` (hex sha256, required when recipe ≠ none), `targets` (Experiment only, optional list of Hypothesis ids), `decided_by` (Experiment only, optional: `"reading"`, the verdicts are decided by reading), `aims` (Experiment only, optional list of at most 10 limit or reading keys the Experiment is planned to settle) |
| `result.record` | the new Result id (= event id) | `parent` (Experiment id), `verdicts` (object: Hypothesis id → `supports` / `contradicts` / `inconclusive`; at least one entry), `observation`, `evidence` (list of at most 10 strings, each at most 300 runes), `evidence_hashes` (optional object: evidence pointer → hex sha256 of the recipe bundle it names; each key is an entry of `evidence` that cleans to `recipes/<name>@<version>/<file>`), `limits` (optional list of at most 10 strings, each 1–300 runes), `closes` (optional list of limit keys `<result-id>#<n>`), `readings` (optional list of at most 10 readings: each a string of 1–600 runes, or an object `{"of": <reading key>, "text": <1–600 runes>}` when it reads an owner's answer; a reading with no `of` is always stored as the string), `applies` (optional list of reading keys `<result-id>#r<n>` whose owner answer this Result applies) |
| `result.retract` | Result id | `{}` |
| `hypothesis.abandon` | Hypothesis id | `{}` |
| `hypothesis.amend` | Hypothesis id | `body` (the new body, required) |
| `dq.select` | DesignQuestion id | `hypothesis`, `prev_select` (event id or `""`), `tradeoff` (optional bool, written by older binaries; replay ignores it), `scope_change` (bool) |
| `dq.unselect` | DesignQuestion id | `prev_select` |
| `dq.defer`, `dq.withdraw`, `dq.resume` | DesignQuestion id | `{}` |
| `experiment.withdraw` | Experiment id | `{}` |
| `dq.request_verify` | DesignQuestion id | read-only history: older binaries wrote `select`; replay accepts the event, whatever it holds, and does nothing with it |
| `confirm` | DesignQuestion id | `select` (a `dq.select` event id on this DQ), `role` (`owner`; a `verifier` confirm, written by older binaries, is read-only history: replay accepts it, whatever it holds, and does nothing with it), `verdict` (`sufficient` or `insufficient`), `accepts` (optional list of limit keys `<result-id>#<n>`) |
| `reaffirm` | Hypothesis or DesignQuestion id | `{}` |
| `checkpoint` | Goal id | `summary`, `role` (optional, `owner`), `closes` (optional list of open reading keys `<result-id>#r<n>` the owner answers), `revises` (optional list of answered reading keys whose answer the owner's `summary` replaces), `accepts` (optional list of at most 10 limit keys `<result-id>#<n>` the owner accepts untested), `refers` (optional list of at most 10 limit keys referred to the owner), `returns` (optional list of at most 10 referred limit keys the owner hands back to the runner) |
| `edge.add`, `edge.remove` | the edge's `from` node id | `type` (`depends_on`, `supersedes` between two Hypotheses or two Experiments, `external_ref`), `to` (node id; for `external_ref` a string such as `gh:976` or a URL, at most 300 runes) |

**[added] `dq.resume`**: undoes a defer or withdraw. Without it a mistaken defer is permanent in an append-only log.
**[added] `confirm` with `role`**: §3 required a verifier check on every select and an owner check on trade-offs, and replay can only be pure if both are stored events. The verifier role is retired (ADR 0015): the select is the gate, and an owner confirm is needed only on a scope-change select.
**[added, then removed from use] `dq.request_verify`**: asked the verifier to judge the current select again. An insufficient verdict now lapses when the Dossier changes (§7.4), and a new verifier can confirm at any time, so no command writes this event (ADR 0011). Old logs keep it.
**[added] `experiment.withdraw`**: closes an Experiment that will not run and that no other Experiment replaces; one that is replaced gets a `supersedes` edge from its replacement instead (ADR 0008). Without either, an Experiment that will never run stays `planned`.
**[added] `hypothesis.amend`**: replaces a Hypothesis body. Without it a wrong body can only be fixed by superseding the Hypothesis.
**[added] limits**: a Result names the limits it found (`limits`): each is a case the Goal covers that no check ran on the Hypotheses the Result judged (ADR 0016). A later Result that runs the case closes it (`closes`). A limit's key is `<result-id>#<n>`: the Result's full ULID, then `n`, which counts the Result's `limits` from 1. The CLI accepts any node reference before the `#` (§2) and stores the full id. An owner confirm or an owner checkpoint accepts limits (`accepts`). §8 derives each limit's state. `hyp record` prints the key of each limit and reading it records, under the Result's line, so a later command names them without counting `--limit` or `--reading` flags, and the whole text of each limit it closes (`  closes R-…#n: <text>`). A close covers every case its limit names; a case the Result did not run is a new `--limit` on it, referred when void (ADR 0016). A close records that the case ran, whatever the verdicts: a Result that runs the case and contradicts a Hypothesis closes it too. `hyp record` warns, at exit 0, when its Experiment aims at an open limit that the Result neither closes nor follows with a new `--limit` (`warn: E-… aims at R-…#n and this Result does not close it; …`), since an instrument failure may leave an aimed limit open on purpose.

**One field, one transition.** A payload field that names keyed items names one kind of item and one transition. A Result's `closes` closes limits; a checkpoint's `closes` closes readings and its `accepts` accepts limits, as a confirm's `accepts` does; `refers` refers limits and `returns` returns them; `aims` names intent and settles nothing (ADR 0013).

**[added] readings**: a Result names a line of the Goal that a verdict rests on and how the run read it (`readings`). A reading's key is `<result-id>#r<n>`, and `n` counts the Result's `readings` from 1. Its states are open and closed, and both are derived. An owner checkpoint closes it: `role` is `owner`, `closes` names the key, and `summary` holds the owner's answer, which `hyp show` prints beside the closed reading. A reading is closed once: a second close of an answered reading is refused, so a merged log with two closes keeps the first and rejects the other. An owner changes an answer they gave with `revises`: a later owner checkpoint whose `revises` names the key and whose `summary` holds the new answer. The answers are kept in merge order, the close first, and the latest is the answer in force; `hyp show` prints it beside the reading and each earlier one beneath, with the Result that applied it. A revision never reopens the reading and changes no event before it: an earlier Result keeps its verdicts, and a Hypothesis it contradicted stays contradicted until a new Result judges a successor. A revision that sorts before its key's close is refused, and two revisions both apply, the later in force. A reading of a retracted Result is left out. It refuses, at exit 1, a close without the owner role, a key that names no reading, a retracted Result, and a close under the actor that recorded the Result. An open reading prints under AWAITING OWNER and changes no status: it never changes a question's state or the Goal's. It holds back ledger items on its question only while its **reach** is not empty: the live Hypotheses under that question whose standing the answer can move, read from the Result's own verdicts, which are a candidate the Result judged, and a contradicted rival whose only `contradicts` is that Result's. While the reach is not empty it holds `rivals`, `select` and `choose`, because the answer may revive a rival, drop the lone candidate or change the candidates (§8.1). A reading whose reach is empty, such as one whose Result's Hypotheses another Result has since contradicted, holds nothing. A closed reading is **applied** by the first non-retracted Result that lists its key in `applies`. Each answer is applied by the first non-retracted Result after its checkpoint, and before the next revision, that lists the key in `applies`, and by no other Result: a later Result under the question may have run for another reason, such as a repair, and never read the answer (ADR 0013 amendment of 2026-11-09). A tree written before `applies` existed therefore owes `reading` for answers its Results applied without naming them. The reading is applied when the answer in force is. Until a Result applies it, an answered reading is the runner's `reading` item (§8.1), which holds back the owner's `choose` on its question, because the owner's answer cannot yet have changed the candidates. Closing one does not change the Goal's text. A reading reads either a line of the Goal or an owner's answer to an earlier reading; a reading of an answer names that reading's key in `of`. It narrows the answer and never reopens it: the earlier reading stays closed and its answer stays in force, and the owner changes their words only by a rule change. `of` may name only a closed reading of a non-retracted Result, so a chain always points back in merge order and holds no cycle; a merged log in which the reading of an answer comes before its close rejects that Result. Without this, a verdict's reliance on an unclear Goal line stays in prose the owner has to go looking for. Two readings with the same `of` and the same text, spaces trimmed, **ask the same question**: each holds its own Result's question, and nothing stored ties them. `hyp status` prints the open ones as one AWAITING OWNER line, each key with its question, and `hyp checkpoint` warns (`warn: R-…#r1 asks the same as R-…#r1 and stays open; add --closes R-…#r1`, or `… keeps its earlier answer; add --revises …`) when its `closes` or `revises` leaves a copy behind (ADR 0013).

Content limits, in runes: `title` ≤ 120 and non-empty; `body`, `observation` and `summary` ≤ 600 (an amend's `body` is also non-empty), except a Goal's `body` ≤ 2000, because it holds the owner's met rule; a reading ≤ 600, because it holds both meanings of a Goal line and the question for the owner; `why` 1–300.

## 5. Refusals

The same function, `replay.Validate(state, event) error`, is used by the CLI before appending and by replay on every event. At replay time a refused event is **skipped and reported** as a `Rejection{EventID, Reason}`; replay never panics and never stops on a bad event. This covers logs merged from several worktrees, where each event was valid alone.

Each event type's rule is its entry in the table in `internal/replay/rules.go`, with its apply and basis nodes (ADR 0013). A rule takes `Facts`, so it cannot read a derived status.

An event is refused when any of these holds:

1. `v` ≠ 1, the type is unknown, a required field is missing, an unknown payload field is present, `ts` is not UTC, or a content limit is exceeded.
2. Its `id` duplicates an earlier event id.
3. `node` or a referenced id (parent, target, hypothesis, select, `to`) does not exist yet, or has the wrong kind.
4. Wrong parent kind. Goal: parent `""`, and only one Goal per log. DesignQuestion: parent is the Goal or a Hypothesis. Hypothesis: parent is a DesignQuestion. Experiment: parent is a Hypothesis. Result: parent is an Experiment.
5. An Experiment `targets` entry is not a Hypothesis with the same parent DesignQuestion as the Experiment's parent, or equals the Experiment's parent.
6. A `result.record` verdict key is not the Experiment's parent Hypothesis or one of its `targets`.
7. `result.retract` on a Result that is already retracted. `hypothesis.abandon` on a Hypothesis that is already abandoned.
8. `dq.select` of a Hypothesis that is not a direct child of that DQ, or is superseded or abandoned.
9. An owner `confirm` whose `select` is not a `dq.select` on this DQ. A `verifier` confirm is history and is not refused (§4).
10. `supersedes` where `from` = `to`, the two Hypotheses have different parent DQs, or the edge would create a supersedes cycle. **[added]** It may also join two Experiments, whose parent Hypotheses must have one parent DQ (`cross-DQ supersedes`); it is refused between an Experiment and any other kind (`wrong kind`), and onto an Experiment that has a non-retracted Result (`bad supersedes: <id> has the non-retracted Result <id>`).
11. `depends_on` whose `from` is not a Hypothesis or DQ, whose `to` is not a Hypothesis, where `from` = `to`, or where the edge would create a cycle among `depends_on` edges, which is the only way the **combined status graph** can hold one (see §7.2).
12. `edge.add` duplicating a live edge; `edge.remove` of an edge that is not live.
13. An `alias` that is already used.
14. `hypothesis.amend` on a Hypothesis that is superseded or abandoned (`bad amend: <id> is superseded or abandoned`), or whose `body` equals the current body (`bad amend: body is unchanged`).
15. A limit key in `closes` or `accepts` that does not name a limit (`bad limit ref`): the key is not `<ref>#<n>` with n ≥ 1; `<ref>` is not a Result; the Result is retracted; n is greater than the Result's number of limits; or the Result is not under the question. For `closes` the question is the one the new Result belongs to (the parent DQ of its Experiment's Hypothesis); for `accepts` it is the confirm's DQ.
16. Removed (ADR 0011). Nothing refuses a `dq.request_verify` beyond the envelope checks in rule 1.
17. An owner `confirm` with `accepts` whose verdict is not `sufficient` (`bad confirm: only a sufficient owner confirm accepts limits`).
18. A `result.record` `evidence_hashes` key that is not an entry of `evidence` or does not clean to `recipes/<name>@<version>/<file>` (`bad evidence_hashes: …`), or a value that is not lowercase hex (`invalid evidence_hashes value for …`).
19. `decided_by` on a node that is not an Experiment (`unknown field`), or a value other than `reading` (`invalid decided_by`).
20. `experiment.withdraw` on a node that is not an Experiment (`wrong kind`), on one already withdrawn (`already withdrawn: <id>`) or superseded (`already superseded: <id>`), or on one that has a non-retracted Result (`bad withdraw: <id> has the non-retracted Result <id>`). `hyp withdraw REF` writes `experiment.withdraw` for an Experiment and `dq.withdraw` for any other node.
21. `aims` on a node that is not an Experiment (`unknown field`); more than 10 (`over limit`); a key named twice, a reading key that names no reading or a retracted Result's reading, a reading still open (`bad aim: reading <key> is open; only an answered reading can be applied`), or a limit key that names no limit under any question (all `bad aim: …`).
22. A `result.record` `closes` key under another question than the Result's, unless its Experiment's `aims` names the key (`bad limit ref`). An `applies` key that is not among its Experiment's `aims` (`bad applies: <key> is not among <exp>'s aims`).
23. A `checkpoint` `closes` without the owner role (`bad checkpoint: only an owner closes a reading`), a limit key (`bad closes: <key> is a limit; an owner accepts a limit with accepts`), a key that is bad as in §4 readings (`bad reading ref: …`), a close of a reading already answered (`bad closes: <key> is answered; an owner revises an answer with revises`), or a close under the actor that recorded the key's Result (separation of duties). A `revises` without the owner role (`bad checkpoint: only an owner revises an answer`), a limit key (`bad revises: <key> is a limit; a limit is closed, accepted or referred, never revised`), a bad key as in §4 readings (`bad reading ref: …`), a key also in the same checkpoint's `closes` (`bad revises: <key> is also in closes`), a reading still open (`bad revises: <key> is open; an owner answers an open reading with closes`), or a revision under the actor that recorded the key's Result (separation of duties). A `refers` with more than 10 keys (`over limit`), a key that names no limit (`bad refers: …`), or a limit an owner checkpoint accepted (`bad refers: <key> is accepted; only the owner reopens it, with returns`). A `returns` without the owner role (`bad checkpoint: only an owner returns a limit`), with more than 10 keys (`over limit`), a key that names no limit (`bad returns: …`), a key also in the same checkpoint's `refers` or `accepts`, a limit whose latest refer, return or accept is a return or none (`bad returns: <key> is not referred or accepted; only a referred or accepted limit is returned`), a closed limit (`bad returns: <key> is closed by R-…`), or a return under the actor that recorded the key's Result (separation of duties). The CLI refuses `--returns` without `--role owner` at exit 2. In a merged log, a return that sorts before its limit's refer is rejected and reported.
24. A `checkpoint` `accepts` without the owner role (`bad checkpoint: only an owner accepts a limit`), with more than 10 keys (`over limit`), with a key that names no limit (`bad accepts: …`), or under the actor that recorded the key's Result (separation of duties).
25. A reading whose `of` names no reading or a retracted Result's (`bad reading of: …`), or a reading still open (`bad reading of: <key> is open; the owner has not answered it`).

CLI error text is `refused: <reason>` on stderr, with exit code 1. A help request (`-h` or `--help` on any command, and `help` at the top level and on a group command) prints usage on stdout and exits 0 at every level, while a missing or unknown subcommand exits 2.

### 5.1 Write-time refusals

Eight more refusals come from the CLI before it appends, not from `Validate`. The first three read derived status, the next two read recipe files, and the last two read the command line; `Validate` checks stored events only, so that a later change to a derivation, a recipe or a text rule cannot turn an old log into rejections. Replay accepts these events, so a log merged from another worktree can hold them.

- `hyp add exp` under Hypothesis H is refused when H's question has live rivals of H (§7.1) and `targets` names none of them, unless `--target none` is passed. Such an Experiment in a merged log still shows as `one-rival` in `hyp signals`.
- `hyp select Q H` is admitted only when the ledger (§8.1) owes it: a `select` item for H, or the owner's `choose` on Q with H a candidate. Otherwise it is refused, naming what is owed instead, in this order: the first item the runner owes on Q (`Q owes the runner first: limit R-…#1`), an open reading on Q that waits for the owner (`open reading R-…#r1 waits for the owner`), the premise hold (`H-… assumes H-…, in Q-… open; decide Q-… first`), the lone candidate's open sub-question (`awaiting_subquestions: H-… has a sub-question not answered`), or Q's state. When a select is owed for another Hypothesis or H is not a candidate, the refusal says so. An answered Q owes no select: `hyp unselect` first, then select. An open limit under Q holds the select whichever Hypotheses its Result judged, since its case may apply to every candidate (§8.1, ADR 0016). `hyp show Q` lists, under `not separated:`, each pair of live rivals no non-retracted Result gives different verdicts (`replay.Projection.Unseparated`); it is a derived list, not a signal, and changes nothing the ledger owes. The open-reading refusal names the reach: `open reading R-…#r1 waits for the owner; its answer can move H-…`. `hyp select Q` with no Hypothesis prints what the same ledger reads: `suggested: H-…` for an owed select, `trade-off: owner decides between …` for an owed choice, or `suggested: none (…)` with the refusal's reason. `replay.Projection.AdmitSelect` holds the rule. Replay still applies a select in a merged log.
- `hyp confirm --role owner --verdict sufficient` is refused while a limit under the question has `Gating` set (§8) and the confirm does not accept it. `hyp confirm` on a question with no current selection is refused too, since it has no select to name. `replay.Projection.AdmitConfirm` holds both.
- `hyp add exp` is refused when the Experiment has no measure and does not say its verdicts are decided by reading: its recipe is `none`, or a pinned recipe that is not an instrument (§7.6), and `--decided-by reading` is not passed. The refusal names `RECIPE-FORMAT.md`, Verdict scripts, as the way to a recipe. A pinned instrument needs no declaration, and `--decided-by` with one exits 2. Replay does not check the declaration: an Experiment with neither in a merged log still applies and keeps its status.
- `hyp record` under an Experiment whose recipe is not `none` is refused when that recipe no longer loads or no longer hashes as the Experiment's `recipe_hash`, or when an `--evidence` pointer into that recipe hashed differently from `recipe_hash` while recording (§7.6).
- `hyp record` under an Experiment that is `superseded` or `withdrawn` (§7.6) is refused, naming its replacement or saying it is withdrawn. Replay does not repeat the check: in a merged log such a Result still applies, the Experiment reads `run`, and `hyp check` warns `<E> is superseded|withdrawn but reads run`.
- Every write command is refused when its `--why`, `--body`, `--observation`, `--summary` or any `--limit` cites an id that is no node of the store. A cite is a token that starts at a word boundary with `G-`, `Q-`, `H-`, `E-` or `R-` and goes on with six upper-case Crockford base32 characters that end the word (a display id, §2), or with a placeholder mark: `<`, `{`, `…`, `...` or three or more `X`, as in `R-<new>`. A display id that resolves (§2) passes; a placeholder never does. Lower-case prose such as `E-mail` or `R-squared` is no cite, and a token right after `:` cites another store (`containment:R-ABC123`) and is not checked. The refusal names each flag and token, `--why names R-<new>, which is no node here`, appends nothing and exits 2, as `unknown node` does (§2). Text written before this rule, or by an older binary, replays unchanged.
- `hyp confirm --role verifier` is refused at exit 1: `the verifier role was retired; a review that finds a defect records a Result, a retract or a limit` (ADR 0015). An old verifier confirm in a merged log is history (§4).

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

A Hypothesis's status reads verdicts only. A sub-question does not change it: the block it puts on its parent belongs to the question that selected the parent (§7.4).

A Hypothesis is **live** unless it is `superseded` or `abandoned`. A DQ is live unless it is withdrawn.

A Hypothesis is a **live rival** when it is live and not `contradicted`. The untargeted-rival refusal (§5.1), the one-rival and D1 signals count live rivals. Staleness (§7.3) and the candidates (§7.4) count live Hypotheses, contradicted ones included.

### 7.2 Combined status graph (for cycle refusal)

What reads what: a Hypothesis reads only other Hypotheses, its status from verdicts and its stale through `depends_on`. A DQ reads its child Hypotheses, the sub-questions of its selected one and its own `depends_on` targets. No edge runs from a Hypothesis to a DQ, and a DQ reads other DQs only down the tree, so a cycle can only run through `depends_on` between Hypotheses. A new `depends_on` that would make such a cycle is refused (§5 rule 11). Because of this, derived fields can always be computed by memoized recursion without looping.

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
- A sub-question on a target does not make its dependents stale, because it does not change the target's status. A claim that rests on an answer says so with `depends_on`.
- Stale is returned as a cause chain, not a bool, so the status screen can print `H-5 stale ← H-3 contradicted ← R-9`. The chain ends at the Result that contradicted (the latest non-retracted contradicting Result), when there is one.

### 7.4 DesignQuestion state

Hold state is the latest of `dq.defer` / `dq.withdraw` / `dq.resume`; resume clears it.

One record, the question's `Decision`, owns everything below except the hold state: select history (each select with the candidates when it applied), conflicts, owner confirms in one list per select (each with the Dossier at its place in the log), the owner requirement, limit acceptances and gating (§8) and the reason order. Render, signals, check and the CLI read it.

Selection is the current selection id (`""` when there is none) plus a conflict set:
- Applying `dq.select` or `dq.unselect` whose `prev_select` ≠ the current selection id adds both ids to the conflict set. The event still applies; the last one in merge order becomes current.
- A later select or unselect whose `prev_select` equals the current id clears the conflict set.
- **Candidates at the select**: when a `dq.select` applies, replay stores the question's candidates just before it, as it stores the live rivals at an Experiment's add (§8). It is a fact derived at apply time, not a payload field, so old logs get it too.
- A **candidate** is a live child Hypothesis whose status is `supported`; neither stale nor an open sub-question removes it. Two or more candidates are a **trade-off**, a state derived from the candidates and never recorded; an older select's stored `tradeoff` is ignored. The owner ends a trade-off with their choice, recorded as the select (ADR 0015): hyp admits that select only when the ledger owes the choice (§5.1, §8.1). The select holds what was chosen, not who chose it: its actor is whoever wrote it, and no status reads its `why`.
- The **owner requirement** is derived and stored nowhere: the current select's `scope_change` is true. The owner confirm is the only owner act replay reads, and only a scope-change select needs one.

Dossier: the evidence a question's current select rests on (ADR 0011). It is derived per question, from stored events only, and never stored: the selected Hypothesis's body, the set of live child Hypotheses, every non-retracted Result with a verdict on one of them with that verdict (found by verdict, not by tree parent), and the live links going out from those Hypotheses and Results, not counting `external_ref`. A rival's body, its planned Experiments and links other nodes add pointing in stay out. `hyp show <Q>` lists it under `dossier:`, as the readout of what the select rests on, and it is the lapse rule of the owner confirm. A sub-question's answer never reaches the parent's Dossier: any answer discharges it, so an owner confirm on the parent counts again once the sub-question is settled, and a flip of its answer gives no `changed since` line.

A `contradicts` is **deciding** when the Hypothesis it judges is not the selected one and, by rule 3 of §7.1, that Hypothesis would be `supported` without this one verdict. `hyp show <Q>` marks its Result as deciding the trade-off. It is a mark: §7.1 still stands and no status reads it.

Owner confirms: an owner confirm judged the Dossier as it stood at the confirm's place in the merged log. For each select, the owner confirm that **counts** is the latest one whose Dossier equals the current Dossier; when there is none, none counts. A confirm that no longer matches has **lapsed**, and it counts again if the Dossier returns to what it judged, which is the §7.3 rule applied to a confirm, so a retract restores the earlier state exactly. This holds for both verdicts, and it is the only way a confirm stops counting. `hyp show <Q>` prints `owner: <verdict> — <why>` for the owner confirm that counts, or the latest, and `changed since <actor>: …` when the latest has lapsed, naming what the Dossier gained (`+`) and lost (`-`) and `body` for an amended selected body. A lapse adds no status word. A verifier confirm in an old log applies nothing and is not listed (ADR 0015).

The state is decided by the first matching rule:
1. `withdrawn`, then `deferred` (from the hold state).
2. `answered` when all of these hold: the conflict set is empty; there is a current selection, naming a live child Hypothesis that is `supported` and not stale; the selected Hypothesis has no **open sub-question**, meaning one that is not withdrawn and not `answered`, so a deferred one still blocks; the DQ itself is not stale; every candidate now was a candidate when the current select applied (a candidate that fell since does not count); and, if the owner requirement holds, the owner confirm that counts on the select is `sufficient`.
3. Otherwise the state is `reopened` if `everAnswered`, else `open`. The reason is the first that holds:
   `conflicting_selections`, `selected_not_supported`, `selected_stale` (selected Hypothesis or the DQ is stale), `awaiting_subquestions` (the selected Hypothesis has an open sub-question), `new_candidate` (a candidate now was not one when the current select applied, such as a rival whose deciding `contradicts` was retracted), `awaiting_owner` (the select is a scope change and the owner confirm that counts, if any, is not `sufficient`), `no_candidates` (no live child Hypothesis), `untested` (every live child is `proposed`), `all_contradicted` (every live child is `contradicted`), `inconclusive` (no live child is `supported`), `supported_unselected`.

A Dossier change that adds no candidate, such as a retract that leaves the selected Hypothesis supported, a new rival or an amend, leaves the question answered: the select is the gate, and the ledger owes what such a change brings (§8.1). The verifier-loop signal and its chain were removed with the verifier role (ADR 0015).

**Status mark:** the owner's `choose:` line under AWAITING OWNER (§8.1) ends `(no Result separates them)` unless a non-retracted Result supports one candidate and gives another candidate a different verdict. The mark is derived at render and changes no state.

**Suggestion:** `hyp select Q` with no Hypothesis reads the ledger, as admission does (§5.1): it suggests the Hypothesis of an owed `select`, prints "trade-off: owner decides between …" for an owed `choose`, and otherwise prints `suggested: none (…)` with the reason a select would be refused, such as `awaiting_subquestions: H-… has a sub-question not answered`.

### 7.5 Goal

- `complete` when every live DQ whose parent is the Goal is `answered`. Sub-questions are covered through the `answered` rule of §7.4, so sub-questions under rejected rival hypotheses never block the Goal. (This clarifies the advice's "nothing stale" clause: stale on the selected path already prevents `answered`.)
- Otherwise `in_progress`, with counts per DQ state over all DQs and a count of stale live nodes.

### 7.6 Experiments and Results

- An Experiment's status, first matching rule wins (ADR 0008):
  1. `run`: it has a non-retracted Result. A live Result beats a close.
  2. `superseded`: a live `supersedes` edge from another Experiment points at it.
  3. `withdrawn`: it has an applied `experiment.withdraw`.
  4. `planned`.

  `superseded` and `withdrawn` are **closed**. Removing the `supersedes` edge reopens an Experiment. A closed Experiment keeps its D1 and one-rival marks in `hyp signals`; `hyp status` lists no line for it, and `hyp show` prints `status: superseded (replaced by E-…)` or `status: withdrawn`. Question and Goal states do not read Experiment status, so a close changes neither.
- A **pin** is a recipe ref (`name@version`) and the hash an event recorded for it at write time. A bundle is its ref and hash; a `recipes/<name>@<version>/` directory name is only the library's lookup key, and hyp opens each copy it makes by the ref it holds ([ADR 0005](adr/0005-a-bundle-is-its-ref-and-hash.md)). A Result depends on its Experiment's pin (`recipe` and `recipe_hash`, unless the recipe is `none`) and on one pin per entry of its own `evidence_hashes`, whose ref is the pointer's first segment under `recipes/` after cleaning the path. A Result is `questionable_evidence` when, for any pin it depends on, `recipeHashes[ref]` is missing or differs from the pinned hash. A new version of a recipe is a new ref, so it never flags Results recorded under an older one. `recipeHashes` leaves out a bundle that cannot be hashed (a symlink, a file or directory name outside `[A-Za-z0-9._-]`, or a `recipe.md` whose name or version does not match its directory; a name below a known answer's `known-answers/NAME/input/` may also hold `@`, so an input can hold a store with its pinned bundles) and a ref whose copies hash differently, and `hyp check` names each one. A bundle that hashes but does not open stays in `recipeHashes`, so its pins verify, and `hyp check` names it too (sandbox spec §1.2). This is a flag only and never changes any status.
- **[added]** An evidence pin records the bundle as it was when the Result was written, not when its files were read; an edit made in between and undone before recording is not detected, and for refs other than the Experiment's own, neither is one still in place. `hyp record` refuses a Result whose Experiment's recipe no longer hashes as the Experiment pinned, or whose evidence pin for that recipe differs from the Experiment's `recipe_hash`, so a Result never pins two hashes for one ref. Replay does not repeat that check: such an event still applies and reads as questionable.
- **[added]** A matching hash says the procedure is the same, not that it gives the right answer. An **instrument** is a recipe bundle with an Evidence or Standalone command, the checks a verdict counts. A script whose output a Result's verdict counts lives in a bundle and runs as one of its commands (sandbox spec §1), never outside `recipes/`. An instrument must have known answers: cases under `known-answers/NAME/`, each an `input/` and an `expect.json` (`RECIPE-FORMAT.md`, Known answers). A check whose cases print `verdict: LABEL=WORD` lines is a verdict check, and `hyp recipe check` warns when a label has no case printing `supports` or none printing `contradicts`, on a malformed verdict line and on a label printed twice in one case; nothing is refused (ADR 0004). Before every `hyp sandbox create` (for each stacked recipe, in stack order, before anything is made), `hyp recipe run` and `hyp recipe evidence`, hyp runs the instrument's known answers with the operator's hyp; a failed case refuses the run with exit 1, a timed-out case exits 3, and a refused create leaves no sandbox. `hyp add exp --recipe` refuses an instrument with no case. A **measured** Experiment is one whose harness recipe is an instrument: its pinned recipe's commands hold the measure, and `recipe.md` declares nothing else. An Experiment with no measure, `none` or a pinned recipe that is not an instrument, declares `decided_by: reading` (§5.1); an environment recipe is never pinned, so the rule never asks it. Replay never reads the declaration, and a false `--decided-by reading` that hides a measure is not seen. The sandbox record and the command record both hold the outcome of each case, as `known_answers` (sandbox spec §5.1). A bundle without an Evidence or Standalone command is not gated (ADR 0002).
- **[added]** A Result is **uncontained** when hyp cannot vouch for its verdict script (ADR 0002). Each reason is reported once, in this order:
  1. it cites a script outside a versioned bundle (`recipes/<name>@<version>/`): a regular file that ends in `.py` or `.sh`, has any execute bit, starts with `#!`, or ends in `.go`, `.js`, `.ts`, `.rb`, `.pl` or `.rs`, including one under `recipes/` in no versioned bundle, which hyp neither pins nor runs, and one inside a bundle that its hash skips (a dot-entry or `__pycache__`), which no pin covers;
  2. it has a pin (its Experiment's recipe or an evidence pin) to a bundle with an Evidence or Standalone command, and cites no `command.json` that names that ref at that hash, records at least one command, every one of which exited 0, and holds `known_answers` with at least one case, every one `ok`;
  3. it cites a `command.json` that hyp cannot read, or that names no ref or hash, or whose ref and hash match none of its pins.

  A pin to a bundle without an Evidence or Standalone command, such as a known-answer fixture, is held by its hash alone (§7.3). The CLI derives the mark from files when it reads, as it does `recipeHashes`. Replay never reads files, and the mark never changes a status. `hyp record` warns, and `hyp show`, `hyp status`, `hyp render` and `hyp check` show it. Nothing is refused. Limits: a script named only in prose, whose output a Result cites, is not seen; neither is a hand-made output file cited beside a real command record, or a script with none of the marks in reason 1. A hand-written `command.json` that claims a clean run with passing known answers satisfies reason 2; a reviewer checks the record's `command.log` and `known-answers.log`. A script that only a Launch, Doctor, Drive or Cleanup command reaches, cited from another Experiment, is held by its hash alone: hyp runs it in a sandbox, but no reason ties that sandbox's record to the Result. The gate proves that a bundle reproduces its own expected outputs, not that those outputs are right: a weak or wrong case needs review (ADR 0002). A case and a run get the same environment, working directory and hyp, and a case compares exit code, stdout and the tree under `$HYP_OUT` (ADR 0003), so a passed gate carries over to the run, except for what a case cannot see: a read of a fixed absolute path, a read from the network, the versions of the tools on the host `PATH` after hyp, and a write outside `$HYP_OUT`; all pass the double run. The paths of a run's ARGS are the subject's, and a limit of the host, such as the length of a Unix socket path, can differ between machines. Reason 2 names why a cited record of the pin is not clean: the first failed command, or known answers that did not pass.
- **[added]** A Result's text can disagree with the files it cites. The CLI reads each local file the Result cites (a URL is skipped), up to 16 MiB each, and warns in two cases:
  1. **verdict mismatch**: a cited file holds `verdict: LABEL=WORD` lines (WORD is `supports`, `contradicts` or `inconclusive`) for a Hypothesis the Result judges, LABEL being its alias, display id or id, and none of them gives the verdict the Result typed;
  2. **uncited**: the observation or a limit holds a lowercase hex token of 7 to 64 characters, with at least one digit and one letter (a commit or a hash), that no cited file holds; the warning names the field. A reading quotes the owner and is not read. A Result that cites no readable local file gets no such warning.

  `hyp record` warns after the event is appended and `hyp check` warns per Result. Nothing is refused and no status changes; hyp never reads a verdict rule (ADR 0004), and replay never reads files. Limits: a verdict line that names no Hypothesis is ignored, a cited file that gives two words for one Hypothesis agrees with either, an id-like word that is not hex is not seen, and a token that a cited file holds by chance passes.

## 8. Projection (what replay returns)

```go
type Projection struct {
    Nodes      map[string]*Node      // static fields from node.add / result.record
    Hyps       map[string]*Hyp       // Hypothesis id → its derived record
    Questions  map[string]*Question  // DQ id → its derived record
    Goal       GoalStatus
    Rejections []Rejection
    Applied    []Event              // events that applied, in merge order
    Readings []Reading  // readings of non-retracted Results, in creation order
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
    Confirms  map[string][]ConfirmRec // select id → its owner confirms in log order, each with the Dossier at that point
    Limits    []Limit                // the limits recorded under the question, in Result creation order
    Dossier   *Dossier               // of the current select (§7.4); nil when there is none
}
type SelectRec struct {
    ID, Hyp, Actor, Why string
    ScopeChange         bool
    Candidates          []string     // the question's candidates just before the select applied (§7.4)
}
type Dossier struct {
    Items    []DossierItem           // sorted; two Dossiers hold the same facts when their Items are equal
    Deciding map[string][]string     // rival Hypothesis id → Results whose contradicts is deciding (§7.4)
}
type DossierItem struct { Kind, Node, To, Word, Body string } // body, hypothesis, result or link
type Exp struct {
    Status     ExpStatus         // planned, run, superseded or withdrawn (§7.6)
    Closed     ExpStatus         // superseded or withdrawn when closed, even if a Result makes it run; "" otherwise
    Live       []string          // non-retracted Results, in creation order
    Judged     map[string]string // Hypothesis id → verdict over Live; a later Result wins
    OneRival   bool              // its parent had live rivals at its add and its targets name none of them
    Unjudged   []string          // targets live at its add that no Live Result judged; nil while Live is empty
    ReplacedBy []string          // Experiments with a live supersedes edge to it, in edge order
}
type Res struct {
    Status       ResStatus // recorded or retracted
    Questionable bool     // questionable_evidence (§7.6)
    Pins         []Pin    // its Experiment's pin, then one per evidence_hashes entry, in the order cited (§7.6)
    Uncontained  []string // reasons (§7.6); set by the CLI from files, not by replay
}
type Limit struct {
    Key, Result, Text string   // Key is "<result-id>#<n>", full ULID
    Hyps  []string             // Hypotheses the Result has a verdict for, sorted
    State string               // "open", "closed" or "accepted"
    By    string               // the first closing Result id, or the accepting owner's actor
    Closers []string           // every non-retracted Result that closes it, in creation order
    Gating bool                // open, and Hyps contains the Hypothesis of the question's current select
    Referred string            // the latest refer, while no return follows it, or ""
    Returned string            // the latest return, while no refer follows it, or ""
}
type Hyp struct {
    Status HypStatus   // §7.1
    Stale  *StaleCause // nil when not stale
}
type Question struct {
    State      QuestionState // §7.4
    Reason     string        // open/reopened reason; "" otherwise
    Stale      *StaleCause   // nil when not stale
    Candidates []string      // its candidates (§7.4)
}
type StaleCause struct { Node, Reason string; Next *StaleCause }
type Edge struct { ID, From, Type, To string; Live bool } // ID is the edge.add event's id
```

`Confirms` keeps every applied owner confirm, including one that has lapsed, so `hyp show` still prints the latest. It prints the owner confirm that counts, or the latest when none does. A verifier confirm in an old log is in `Applied`, so `hyp status` RECENT and `hyp show` history still list it, and it is in no `Confirms` list.

Limits **[added]**: a Result's limits are filed under its question, the parent DQ of its Experiment's Hypothesis, whichever Hypotheses it judged. Limits of a retracted Result are left out. A limit is:
- `closed` when a non-retracted Result lists its key in `closes`; `By` is the first such Result in creation order, and `Closers` lists them all. Retracting that Result reopens the limit unless another one closes it. `hyp show` prints each closer with its verdicts (`closed by R-… (H-… (alias) contradicts)`), so the line says what the case did to each Hypothesis it judged: a contradicted one failed the case. `hyp show <R-id>#<n>` prints one limit: its text, Result and question, its state and, while open, whose turn it is; each Hypothesis its Result judged with its status now; the question's candidates; for a closed limit, each candidate no closer judged (`not judged by a closer:`); the Experiments that aim at it; and a history of the Result that records it, the Results that close it with their verdicts, and the checkpoints and confirms that refer, return or accept it with their summaries. A key that names no limit of a live Result is exit 2.
- otherwise `accepted` when an owner checkpoint lists its key in `accepts` and no later owner checkpoint lists it in `returns`; `By` is the latest accepting checkpoint's actor. A Result that closes it comes first.
- otherwise `accepted` when a sufficient owner confirm on the question's **current** select, with the current Dossier, accepted it, and no later owner `insufficient` has that Dossier; `By` is that owner's actor. A new select, a Dossier change or an insufficient owner confirm at this Dossier makes the limit open again, and a Dossier that returns restores it.
- otherwise `open`. An open limit whose key a checkpoint lists in `refers` is **referred**: the runner says no test within the Goal can close it, or that its case is void for every candidate, and only the owner can settle it: by accepting it, so this tree does not run the case, or by returning it (`returns`), because the Goal covers its case. The latest `refers` or `returns` in merge order names whose turn an open limit is: a returned limit is the runner's again, owed and holding the select like any open limit, until a Result closes it or a later checkpoint refers it again. A return also reopens a limit an owner checkpoint accepted, so an owner who changes their mind writes one checkpoint: the latest of accept and return stands. A limit's category is never stored (ADR 0016): a close with the closer's verdicts, an accept, and the latest refer or return are what it is.

Limits never change a derived status. Only the CLI and the ledger read them: `hyp show` and `hyp status` list them, an open limit under a question with a candidate holds the select through the ledger (§8.1), and `hyp confirm --role owner --verdict sufficient` is refused while a limit has `Gating` set (it is open and its Result has a verdict for the Hypothesis of the current select) and the confirm does not accept it (§5.1). An open limit whose Result judged only other Hypotheses does not gate; it stays listed, and `hyp show <Q>` marks it `open, does not gate (judges …)`.

### 8.1 Ledger: what the tree owes

`Projection.Owed()` derives the ledger: each item of open work on a question, with who owes it. An open or reopened question owes every kind of item; an answered one owes only the `limit`, `reading`, `referred` and `premise` items, so evidence that lands after the answer is owed and not only listed. Nothing stores the ledger, and owing an item never changes a status: the gate is at the write (§5.1), not in the derivation. Settling one is always an explicit event; a planned test is intent, not an outcome, so an Experiment's `aims` never settles a limit or a reading, and settling an item never selects, answers a question or meets the Goal. Items, per question in creation order:

| Item | Who | Owed while | Settled by |
|---|---|---|---|
| `rivals` | runner | the question has no live, non-contradicted Hypothesis and no open reading on its Results with a reach (§4 readings); `hyp status` names each contradicted Hypothesis and the Results that contradict it (`ruled out: h-one by R-…`) | a new Hypothesis |
| `untested` | runner | a proposed Hypothesis has no planned or run Experiment as parent or target | an Experiment |
| `undecided` | runner | a live Hypothesis is `inconclusive` (neither supported nor contradicted) and no planned Experiment has it as parent or target | a planned Experiment, then its Result |
| `planned` | runner | an Experiment is planned | its Result |
| `limit` | runner | an open limit under a question that has a candidate, whichever Hypotheses its Result judged, unless its latest refer or return is a refer | a Result that `closes` it, or a `refers` |
| `referred` | owner | an open limit's latest refer has no return, whichever Hypotheses its Result judged; `hyp status` prints `limit referred, accept or return` | an owner checkpoint that `accepts` or `returns` it |
| `reading` | runner | the answer in force of a closed reading is not applied, including after a revision | a Result that `applies` it |
| `premise` | runner | a candidate's live `depends_on` target T sits under another question whose current select names a Hypothesis other than T | re-pointing the edge (`hyp unlink`, `hyp link`), superseding the dependent, or that question selecting T |
| `select` | runner | the question has one candidate, its current select does not stand, no other runner item, no premise hold, no open reading on its Results with a reach, and the candidate has no open sub-question | `dq.select` |
| `choose` | owner | the question has two or more candidates, its current select does not stand, no runner item, no premise hold and no open reading on its Results with a reach | `dq.select` of a candidate, with the owner's words in `why` |

A current select **stands** when the conflict set is empty, it names a candidate, and no candidate is new since it applied (§7.4); a candidate that fell since does not undo it. So `choose` is owed again when a new candidate appears, and the new select stores the new set. The **premise hold**: neither `select` nor `choose` is owed on Q while a candidate has a live `depends_on` target under another question that is `open` or `reopened`; the `assumes … in Q-… open` line under the candidates in FRONTIER is the hold's reason, and `hyp select` refuses with it. The runner settles the premise's question first. A `depends_on` target under Q itself is no premise, because the choice between them is Q's own. Two questions whose candidates assume each other's Hypotheses hold each other; `hyp check` warns `premise cycle between Q-… and Q-…`, and removing an edge breaks it. An answered Q whose premise's question later reopens owes nothing until that question selects against the premise.

A `limit` or `reading` item also has an aim state: `owed` when no live Experiment aims at its key, `in_progress` while an aiming Experiment is planned, and `aimed_not_settled` once every aiming Experiment has a Result and none settled it. An open reading and a scope-change owner confirm are owed by the owner too; `hyp status` prints those under AWAITING OWNER from their own records. A trade-off is a state, printed as `(trade-off)` on the question's candidates line in FRONTIER; only `choose` is the owner's turn, and it waits for the runner because every runner item can change the candidates.

`hyp status` prints the runner's items under OWED, keeping the first five when it truncates, and the owner's under AWAITING OWNER. `hyp status --live` leaves out of FRONTIER each contradicted Hypothesis that holds no open Experiment marked one-rival or unjudged, and prints `(N contradicted hidden)` under its question; it is the owner's view, and the other sections are the same, since a runner needs the contradicted rivals to supersede them. An open question that owes no runner item and no `choose` while an open reading with a reach waits prints `held: open reading R-…#r1 waits for the owner; its answer can move H-…` under it, so `runner owes nothing` names its reason; a premise hold has its `assumes` line instead. Under a grouped reading line, each key prints what the answer can touch on its question: `R-…#r1: Q-… candidates …` (or `no candidate`) `; judged H-… supports, now contradicted`, from the Result's verdicts and each Hypothesis's status now. When no item is the runner's, OWED ends with `runner owes nothing: stop and report to the owner`, the stop clause a run reads; an item on an answered question counts, so a run does not stop with a limit or a premise it owes there.

Rivals at add **[added]**: when an Experiment is applied, replay stores the live rivals of its parent Hypothesis under the same question, read from the derived state just before that event. A target is **unjudged** when it is in that list and no non-retracted Result of the Experiment has a verdict for it; an Experiment with no non-retracted Result has no unjudged target. An Experiment is **one-rival** when that list is not empty and its `targets` name none of it. Both are read at the Experiment's add, so a rival that dies later does not clear a mark, and a rival added later does not create one. `hyp signals` reports an unjudged target, and `hyp status` and `hyp show <Q>` mark the Experiment. They never change a derived status.

The Projection holds every fact hyp derives; render, signals and the CLI read its fields, and nothing stores them a second time (ADR 0006). Scripts read the applied events and rejections through `hyp events --json` and each round's facts and design marks through `hyp signals --json`, never the files under `.tree/events`.

Each node kind has its own status word: `HypStatus`, `QuestionState`, `ExpStatus`, `ResStatus` and `GoalState` are separate string types, so comparing two kinds does not compile (ADR 0013). A `Question` reaches its `Decision` by the shared id. `Projection.Display(id)` returns the word, reason and stale cause of any node, for code that prints a node without knowing its kind.

Equality test used by acceptance: every node's derived record and the Goal are deeply equal.

## 9. Sandbox lifecycle and steps

The sandbox has its own spec, [sandbox-spec.md](sandbox-spec.md), and these sections moved there. Older text, ADRs and historical records still cite the numbers below.

### 9.1 Steps and sections

Now [sandbox spec §1.1](sandbox-spec.md#11-steps-and-sections).

### 9.2 hyp runs every command

Now [sandbox spec §1.2](sandbox-spec.md#12-hyp-runs-every-command).

### 9.3 How a command runs

Now [sandbox spec §1.3](sandbox-spec.md#13-how-a-command-runs).

### 9.4 Records

Now [sandbox spec §5.1](sandbox-spec.md#51-records).

### 9.5 Choices

Now [sandbox spec §5.2](sandbox-spec.md#52-choices).

### 9.6 The subject

Now [sandbox spec §5.3](sandbox-spec.md#53-the-subject).
