---
status: accepted
---

# A limit is an untested case

A Result's `limits` field carried several jobs. The 34 trees kept on 2026-11-08 hold 173 limits. About two in five name a case nobody ran. The rest state the scope of a run, a procedure note, a provenance claim, a confound, or a premise. hyp gives a limit four transitions: a Result `closes` it, an Experiment `aims` at it, a runner `refers` it, and the owner `accepts` it. All four are about a run that has not happened yet, so only the untested case fits all of them.

The ledger (spec §8.1) owed a limit only when its Result judged a candidate. A limit found on a rival's Result was owed by nobody once that rival fell, even when its case applied to the candidate. ADR 0015 named this as loss 2 and mitigated it with a select warning.

Owner decision on 2026-11-08: a limit is a case the Goal covers that no check ran. A rival's limit is answered by a check that runs the case on every candidate. The rivals that survive are what the verdicts leave. The multi-tenant tree did this once already: R-FAMJDB named two cases, E-236WKJ aimed at both, and R-S7MX6V closed them while contradicting both of its candidates.

## Decision

- **A limit is a case the Goal covers that no check ran on the Hypotheses its Result judged.** Its one job is to name the next case. The field name, the key and the four transitions stay. A run's other notes have homes of their own. The scope of a run goes in the Experiment body. A procedure note or a commit goes in the observation, with the commit in a cited file. A premise is a `depends_on` edge. The glossary's Avoid list gains scope, assumption and provenance.
- **An open limit is owed while its question has a candidate, whichever Hypotheses its Result judged.** The runner owes an unreferred one, and the owner owes a referred one. The limit therefore holds the select, and `hyp select` refuses with `Q owes the runner first: limit R-…#1`. This computes ADR 0015's loss 2. The select warning `N open limits judged other Hypotheses` is gone, because an admitted select can no longer have such a limit unreferred.
- **`refers` also covers a case that is void for every candidate.** A runner refers a limit that no test within the Goal can close, or whose case does not arise for any candidate. It is still one transition: the runner hands the limit to the owner.
- **A close covers every case the limit names, and one limit names one case (2026-11-09).** A Result that closes a limit and did not run every case its text names records each case it did not run as a `--limit` of its own. A runner refers that new key when its case is void for every candidate or outside the Goal's lines. A claim that a case is void, out of scope or covered elsewhere is such a limit plus a refer, never a sentence in a body or an observation, because a close leaves the owner nothing to answer. A new limit names one case: "not X or Y" is two limits. `hyp record --closes` prints the whole text of each limit it closes, so the runner rereads every case it names at the moment it claims them.

  Owner decision on 2026-11-09, from run 3 of the multi-tenant tree. Two of its ten closes ended a two-case limit after one case ran. `R-BZ7C5H` closed `R-XFSQEE#1` (Postgres and `.xlsx`) after Postgres alone, and called the `.xlsx` case void only in its observation. `R-M4ABBF` closed `R-AMG1CR#2` (a crash mid-revision and a deploy that misses one) after the crash alone, and named the deploy case nowhere. The same runner filed correctly twice in the same run: it referred `R-GVPBZZ#2`, and recorded K2's untested remainder as `R-M4ABBF#1` and referred it. No new field, verb or kind is added (ADR 0013); the existing `--limit` and `--refers` carry it.

## Consequences

- Old trees may owe more. On a scratch build over the 34 trees, limit items went from 79 to 95. Five trees changed, the live multi-tenant tree among them (2 to 6), and the other 29, both public trees among them, did not. A complete Goal can owe limits that were left on a dead round. Owing an item never changes a status.
- A limit that is not a case, such as a scope statement, now holds the select until it is referred. This is the cost of one meaning: the runner says once why no test closes it, and the owner accepts it.
- Not decided: a candidate added after a limit was closed was never run through that case. A rule that owes such a limit could be derived (§8 allows several closers). It waits for a real event (AGENTS.md, step 1).

## Considered Options

- **A `kind` on each limit.** Rejected: one payload field means one item kind and one transition (ADR 0013), and a kind field would split the transitions without making them computable.
- **Keep the select warning.** Rejected: it warned of work the ledger can owe.

## Amended 2026-11-09: a limit's category is derived, and the owner can return a referred limit

The owner asked whether a limit belongs to the Result, which is provable because the run already happened, or to the Hypothesis or the Goal, which cannot be known in advance, and how a limit is promoted up the tree. A new Goal is a new tree and a new Hypothesis is a new branch, and turning a Result's limit into a rethink of the design or of the Goal is how thinking works, so it must be a visible, cheap path made of ordinary events.

The run that prompted this: `R-P424P1#1` (the missed-deploy case) was referred. `E-ZVN0EM` aimed at it, and its Result `R-5JYRA0` ran the case and contradicted h-module on D2 and D3. But `R-5JYRA0` did not close the limit, following the Experiment's own rule ("Closes R-P424P1#1 if D1-D3 hold"), so `hyp status` still asked the owner to accept or refuse a case that had been run.

Owner decision on 2026-11-09:

- **A limit has one stored scope, its Result's.** At record time, the one provable fact is "this run did not cover case C". No category, kind or scope is stored, because choosing one at record time would ask the runner to judge what it cannot yet know.
- **Its category is derived by rules over typed fields, never from prose.** A Result's `closes` and its `verdicts` say whether the case ran and what it did to each Hypothesis judged: a contradicted Hypothesis means the design failed the case, and a supported one means it passed. An owner `accepts` means this tree does not run the case. The latest `refers` or `returns` says whose turn an open limit is. Summaries and observations explain the reason, and no rule reads them.
- **A close records that the case ran, whatever the verdicts.** A Result that runs a limit's case closes it whether its verdicts support or contradict. Closing only on a pass is the misreading that left `R-P424P1#1` open. Displays show the closer's verdicts next to `closed by`, so the promotion into the design is visible.
- **The owner can return a referred limit.** "The Goal covers this case, run it" is the owner's answer to a refer, and it is a typed transition: `hyp checkpoint --role owner --returns <R-id>#<n>`, the mirror of `--refers`. A returned limit is an ordinary open limit again, owed to the runner while its question has a candidate. Whose turn an open limit is follows the latest `refers` or `returns` in merge order. A RULE CHANGE checkpoint that widens the Goal stays allowed, and names the keys it brings into the Goal with `--returns`.
- **A new Goal is a new tree.** The new tree cites the old tree's keys in text and with an `external_ref` edge. No typed cross-tree cite is added, because nothing would be computed from it.
- **A candidate that no closer judged is owed the case again.** This settles the "Not decided" item above in principle: once a closed limit's question has a supported candidate that no closing Result judged, the limit is owed to the runner for that candidate. It is built when the first successor of a Hypothesis that a closer contradicted is added, because that is its real event.

A return also reopens an accepted limit (2026-11-09, owner: "fix gap 11"). An accept was final, and a `--returns` on an accepted limit was admitted and changed nothing. Now the latest of accept and return stands, so an owner who changes their mind writes one checkpoint, and a runner's refer of an accepted limit is refused, since only the owner reopens one. An accept of a closed limit is refused too (gap 12): the case ran, so the accept would change nothing.

No kind field, no edge to a limit key, no `--answers` on `add hyp`, and no change of turn on `--aims` are added (ADR 0013). `aims` names intent and settles nothing, so it cannot carry the owner's answer.

Displays and checks that follow, in order:
- a limit's line shows its closers' verdicts;
- `hyp record` warns when the Experiment aims at a limit that the Result neither closes nor replaces with a new `--limit`, and prints the admit warnings it drops today;
- `hyp show <R-id>#<n>` prints the limit, the Hypotheses judged and their status now, whose turn it is with the refer and return summaries, the aimers, the closers with their verdicts, and the acceptor;
- `--returns`;
- the re-owe rule, when a successor is added.

Kongming's report `plans/reports/kongming-261109-1529-typed-limit-promotion.md` holds the options weighed.

## Amended 2026-11-09: a referred limit holds the select

The run that prompted this: `R-KHDA9E#1` on Q-QSDM8H was referred to the owner, and `hyp status` still owed the runner `select H-47B370 (h-several)`. A runner could select before the owner answered. If the owner then returned the limit, its case would run against a question already answered.

Owner decision on 2026-11-09: a referred limit holds its question's `select` and `choose` until the owner accepts or returns it, the way an open reading with a reach already holds them. An accept releases the select. A return makes the limit the runner's `limit` item, which holds the select until a Result closes it. The hold does not reach new rivals, because a refer says nothing about the rivals. `hyp select` names the limit in its refusal, and `hyp status` prints a `held:` line under the question.

A limit referred after the select still waits for the owner, and a scope-change owner confirm still accepts it with `--accept`.

## Amended 2026-11-09: a Goal line is a limit the owner records

The runs that prompted this: run 3 and run 4 of the multi-tenant tree both stopped with "Lines A, B, C not checked" in the stop checkpoint. Every scoring Experiment from `E-PKXXT4` to `E-SYV141` had run the line checks A1–C2 and FA–FC2 on every candidate, so the facts were in 27 Results. hyp could not see them, because a Goal line lived only in the Goal's prose and the runner's CHECK Line proposals lived only in chat.

Owner decisions on 2026-11-09:

- **A Goal line is a limit.** It is a case every candidate must meet, and it has every move a limit has. A Result that runs it closes it, whatever its verdicts. An Experiment aims at it. The runner refers it, and the owner accepts or returns it. It is owed while a candidate has not met it, and it holds the select. One kind, two births: a Result records a limit its run left uncovered, keyed `<result-id>#<n>`, and the owner records a line on the Goal, keyed `G-<goal-id>#<n>`. "Line" is the display word for a limit born on the Goal. No `Line` node kind is added.
- **Limits, lines and readings are keyed items, moved by one state machine.** A keyed item is created as one entry inside its parent's event, a Result or the Goal, and is addressed as `<parent>#<n>`. Its text is fixed, it takes no edge, and it is gone when its parent is retracted. Only other events move its state: a Result's `closes`, a checkpoint's `refers`, `returns` and `accepts`, an Experiment's `aims`. One move table states each move, from which state, by whom, and the refusal and exit code. Replay applies the moves in merge order, `Validate` asks the same table, and status, show and the ledger read the same states. Today the state is derived twice, once in `LimitHand` for refusals and once in `limits()` for status, and the two already disagree: a limit accepted on an owner confirm reads `accepted` in status, but `--returns` refuses it because `LimitHand` never saw the accept. The spec gains one section for keyed items, as AGENTS.md asks of a state machine.
- **A line is written once and never edited.** The owner writes the lines on one owner checkpoint, and a later checkpoint with lines is refused. A change to a line goes by one of four existing routes, and none of them edits it:
  - what a line means is a reading, answered by the owner;
  - a line the Goal no longer needs is referred by the runner and accepted by the owner;
  - a case under a line is a limit, returned by the owner when the Goal covers it;
  - a new line or a new purpose is a new Goal, and so a new tree.

  The RULE CHANGE checkpoint convention and the runner's CHECK Line step are removed. This replaces the earlier amendment's sentence "A RULE CHANGE checkpoint that widens the Goal stays allowed": its `--returns` half stays, the convention goes. hyp read neither, and the live tree holds one RULE CHANGE summary and no CHECK Line one.
- **A line is a ledger item only; `complete` does not change.** A line is owed as a `limit` item on each live top-level question that has a candidate no closer covers, and it holds that question's select as any limit does. Spec §7.5 stays "every live top-level question answered". A Goal word that reads lines would be the "status reads limits" option that ADR 0015 rejected, one level up, and the gate already does the work: no question is answered while a line is unrun on its candidate. The owner's "met" is a complete Goal with no line under OWED. hyp stores no word for it.
- **A verdict is per Hypothesis, never per line.** A Result that closes several lines and contradicts H says the cases ran and H failed, not which line H failed. Displays print `ran on H-…: <verdict>` beside `closed by`, and never `fails <line>`, unless the Result closes exactly one line. The observation says which check failed.
- **A Hypothesis is a claim, and it may name its subject.** The subject is the files that embody the claim, tied once and kept by sha256: one `hypothesis.subject` event, written by `hyp subject H PATH…` or by `hyp add hyp --subject PATH`. The latest tie stands until a non-retracted Result's fingerprints hold it. After that a different subject is refused and the refusal names supersede, the rule an Experiment's recipe already follows (ADR 0008). A slip before the first run costs one event. Two Hypotheses may share a subject, because the claim is what differs: in the live tree `h-reader.py`, `h-reader-file.py` and h-reader-record are one file, and `h-name-copy.py` and `h-module.py` are one file. A Hypothesis with no subject, such as a diagnosis or a base arm, is covered by verdicts only, which is the rule today. A prose subject is not added, because it computes nothing.
- **Coverage is derived from the log.** `hyp record` copies the sha256s that the Result's cited `command.json` and `sandbox.json` records fingerprinted (`inputs`, each `dirs` entry, each patch) into the Result as `fingerprints`, as it already copies `evidence_hashes`. Replay reads no files, so the fingerprints must be in the log to reach the ledger. A Result covers a Hypothesis when it judges it or when the Hypothesis's subject is among its fingerprints. Coverage says these bytes ran, not that the claim held. A claim that depends on its host is covered only together with that host, so the display prints `ran on H-…` for a covered Hypothesis the Result did not judge, and a mark `(no one closer covers H-… and H-… together)` under a line when the selected Hypotheses were never run in one record. The command record moves to format 6, with one entry per file under `dirs`, so a subject can be matched by file.
- **An accept answers a refer.** An owner `--accept` is admitted only while the key's latest move is a refer. `--returns` is admitted while it is a refer or an accept, as gap 11 set. The actor who answers differs from the actor of the latest refer. This replaces the rule that compared the answerer with the actor that recorded the key's Result, which would refuse the owner's own moves on a line, since the Goal's actor is the owner. `hyp confirm --accept` is retired as a write, and an old confirm's `accepts` replays as history, so the trees id-citation and recipe-base stay `complete`. Since a limit holds the select, a confirm only meets a limit found after the select, and that limit is the runner's to close or refer. A scope-change confirm is now a checkpoint accept, then `hyp confirm Q --role owner --verdict sufficient`. Readings keep their rule: a reading has no refer, so the author of its Result still cannot answer it. Two costs follow. A waive is always two steps, the runner's refer and the owner's accept. After a return, an accept needs a fresh refer. Excusing a line that no one referred is a change of the Goal, so a new tree. This replaces the last sentence of the amendment above, "a scope-change owner confirm still accepts it with `--accept`".
- **A new tree cites the old one and re-runs.** A cross-tree cite is written `<tree>:ID`, which the id-cite check already skips. Evidence does not cross trees, because an evidence pointer must sit under the citing store's `.tree/evidence/` or `recipes/`: the new tree copies the evidence directories it reads into its own `.tree/evidence/`, and copies the recipe versions it reuses beside its own `.tree/`. The same-hash rule across levels keeps the copies honest. There is no fork command and no carried-over select, because a fork copies decided things. The promotion link is the old tree's owner accept of the line, with a summary that names the new tree as `<tree>:G-…`, and an `external_ref` edge each way.

No new verb or flag is added for the accept rule. The new stored facts are the line items on an owner checkpoint, the `hypothesis.subject` event and a Result's `fingerprints`. hyp writes the last one, and no runner types it.

What waits for a real event:
- a line added after questions are answered, keyed on its own checkpoint as `<checkpoint-id>#<n>`;
- a typed "ended" state for a Goal that a new tree replaced;
- `complete` reading lines, if an owner reads `complete` as done while a line is owed under it;
- a commit as a subject, for a base arm;
- whether one candidate's line item is settled by byte coverage alone, or only by a closer that covers the whole selection. Run 4's `R-KHDA9E` ran h-reader-record's bytes on h-module, where the claim failed under another name, so byte coverage of one Hypothesis can mislead. The first line typed on G-WH6GK5 is the event that decides it.

Considered options:
- **A `Line` node kind,** with a per-line, per-Hypothesis outcome. Rejected: the outcome is a second verdict vocabulary bound to the owner's labels, which ADR 0004 and ADR 0013 refuse, and every move a line needed in four runs is a limit's. It becomes right if owners need per-line failures typed.
- **Owner-declared met** (`--met` on a checkpoint). Rejected: it copies a judgment hyp cannot check.
- **A limit as a node kind.** Rejected: a limit is a gap a Result reported, so it needs no edges and goes with its Result.
- **A per-Result `--subject H=PATH`.** Rejected by the owner as a per-Result claim. It is stronger in one place: it names the claim at the moment the bytes ran.
- **Keep `hyp confirm --accept` and require a referred key.** Rejected: two paths for one meaning of accept.

Kongming's reports `plans/reports/kongming-261109-2052-met-and-limit-lifecycle.md`, `kongming-261109-2135-goal-has-limits-review.md`, `kongming-261109-2205-hypothesis-claim-plus-subject.md` and `kongming-261109-2237-accept-answers-a-refer.md` hold the options weighed. The build follows the owner's choose on Q-QSDM8H, in one PR with the spec, the glossary entries Limit, Line, Coverage, Subject, Hypothesis and Goal, and a census over the trees on disk.
