---
status: accepted
---

# An Experiment is a pre-registered Result

After run 5 of the multi-tenant tree, the owner wanted to know what each of two candidates costs before choosing between them. `E-9FSCC5` counted the rows each shape stores and the files each shape edits for one change. The counts settled no Hypothesis: both candidates met every Goal line, and the difference was a cost for the owner to weigh. hyp had no place for that. An Experiment sat under a Hypothesis, and its Result had to give a verdict, so `R-8DYPS6` said `inconclusive` for both and put the counts in its observation (#106).

The owner also found it strange that a verdict rule is written before the run. That is the point of pre-registration, but the model had never said what an Experiment is. It treated an Experiment as something that proves a Hypothesis, and the measuring shape, the rule by reading, and a reassessment from records already kept each had to look like that.

Two facts from the trees showed where verdict logic lives today. `tenant-lines` has 15 versions: across its 14 steps, `known-answers/` changed in 13 and `check.py` in 9, and at `@15` it has 16 run modes, each described as "the previous checks plus X, every one counts toward the verdict". In the 6 trees that hold events, 15 of the 20 pinned recipe versions serve one Experiment. The procedure stays stable, and the verdict logic changes with the question, so each change of the rule forks the whole recipe.

Owner decision on 2026-11-10: the whole model is remodelled at once, without a separate real event for each shape, because remodelling only the measuring shape would leave a gap.

## Decision

**An Experiment is a pre-registered Result.** It binds a Recipe and a Rubric to the tree. The recipe says how to run and what to measure; the rubric says how the measures become verdicts; the Experiment says on what, about which Hypotheses, and to settle which items. Before the run it fixes five facts, and none of them changes after the add. A different recipe or rubric is a new Experiment.

| Fact | Stored as | What it fixes |
|---|---|---|
| Targets | the parent Hypothesis plus `targets`; none under a DesignQuestion | the Hypotheses its Results give a verdict on, every one and no other |
| Aims | `aims` | the limit and reading keys it plans to settle (unchanged) |
| Recipe | `recipe` and `recipe_hash`, or `none` | the procedure and its checks |
| Rubric | `rubric` and `rubric_hash`, `reading`, or absent | how measures become verdicts |
| Records | `records`: files by path and sha256 | the earlier records it reads beyond its own run |

- **Targets are the one set.** Under a Hypothesis, the targets are the parent and each `--target`. Every Result gives a verdict on each target and on nothing else, and replay refuses one that leaves a target out. `inconclusive` is the verdict for a run that could not decide. The `unjudged` mark is retired: a missing verdict is now a refusal, not a mark. "Judge" is retired as a verb, and the owner confirms.
- **An Experiment with no targets sits under the DesignQuestion it informs.** It pins a recipe, has no rubric, and its Results record measures in the observation and the cited records, with no verdicts. Nothing stores "gives no verdict": the parent kind and the empty targets declare it.
- **The shape is derived, never stored.** It is read from the parent kind, the targets and the rubric:

  | Shape | Parent | Recipe | Rubric | Records |
  |---|---|---|---|---|
  | Run and rule | Hypothesis | pinned | bundle | its own run |
  | Rule by reading | Hypothesis | pinned or `none` | `reading` | own run or pinned |
  | Reassess from kept records | Hypothesis | `none` | bundle or `reading` | pinned |
  | Measure only | DesignQuestion | pinned | absent | its own run |

- **A Recipe measures and never decides.** A recipe is a procedure plus checks. It names no tree id, and each of its checks prints **measures**: one named line about one subject, such as `S1 fail` or `stored rows 4`. `hyp add exp --recipe` refuses a recipe whose known answers expect a `verdict:` line.
- **A Rubric decides.** A rubric is a bundle in the recipe format with no step section and one Standalone command. It lives in its own library, `rubrics/<name>@<N>/`, so two Experiments can pin the same one. It reads measures from records and prints one `verdict: LABEL=WORD` line per label. It is versioned, pinned by its ref and hash (ADR 0005), and gated by its known answers before every run (ADR 0002), and each label it prints needs a case printing `supports` and a case printing `contradicts`. The alternative is prose in the Experiment's body (`rubric: reading`), for verdicts that are decided by reading. `decided_by: reading` in an old log is read as `rubric: reading`.
- **The runner binds labels to targets.** hyp never reads a verdict line at run time. The runner types each target's verdict into `hyp record`, as ADR 0004 already requires; the verdict-mismatch warning compares them afterwards. A rubric names no tree id.
- **Records are pinned at add.** An Experiment that reads records from an earlier run names them at add, and hyp stores each file's path and sha256, the same way a Hypothesis's subject pins files. `hyp record` refuses a Result that cites a file under `.tree/evidence/` outside its Experiment's own run and its pinned records, and one whose pinned record no longer hashes as pinned. The reference is never a Result id, because a ruling depends on the bytes, not on another Result's claim. Containment counts a pinned record as cited.

## Consequences

- The event grammar grows `rubric`, `rubric_hash` and `records` on an Experiment, an Experiment parent may be a DesignQuestion, and rule 6 of the spec requires a verdict on every target. Spec §4, §5, §5.1, §7.6, §8 and §8.1 change with this ADR.
- **ADR 0004 is amended, not reversed.** A verdict check's known answers specify its verdict rule; the verdict check is now the rubric's command, and a recipe has none. The label-to-target binding stays with the runner.
- **ADR 0002, ADR 0003 and ADR 0005 extend to rubrics.** A rubric runs on a copy of itself, reads only the records it is given, and is gated by its known answers.
- **ADR 0008 is extended:** an Experiment's parent may be a DesignQuestion. Its lifecycle is unchanged.
- **ADR 0017 is amended:** `unjudged` leaves `hyp signals`, which keeps one-rival and D1.
- A frozen recipe whose checks print verdicts, such as every `tenant-lines` version, stays valid as history, and an Experiment that already pins one records as before. A new Experiment cannot pin it: the next version splits it into a recipe and a rubric.
- In the 300 trees on the maintainer's machine, most of them known-answer fixtures, only the synthetic store in `internal/testutil/testdata` holds a Result that leaves a target without a verdict. Its fixture changes with the code.
- `E-9FSCC5` counted by reading, with no recipe. Under this model a count by reading is an observation, not a measure: a measure-only Experiment pins a recipe whose check prints the count.
- `skills/hyp-verification/RECIPE-FORMAT.md`, its Verdict scripts section above all, and the glossary change to match. The skill change agrees with this owner decision, and its commit names it.
- The work runs in three slices behind one pre-registration module in `internal/replay`: targets and the missing-verdict refusal; the derived shape with the DesignQuestion parent; pinned records with containment. The Recipe and Rubric split follows them.

## When this stops holding

- If a real Experiment needs to give verdicts and measure at once in a way no shape holds, such as a rubric that must read a measure no recipe can print, add the shape by its facts, not with a stored shape word.
- If rubrics turn out to be single-use in practice, with each one pinned by one Experiment as recipes were, reconsider keeping them in their own library.

## Considered Options

- **A stored `decided_by: none` for "gives no verdict".** Rejected: the parent kind and the empty targets already say it, and a second word could disagree with them.
- **Optional verdicts on a Result.** Rejected: a Result then no longer answers its pre-registration, and a target with no verdict comes back as a silent gap, which the `unjudged` mark existed to show.
- **A rule that cites earlier Result ids as its source.** Rejected: an Experiment is itself the pre-registered Result, so naming another Result as the source is circular, and a ruling depends on bytes that a record pins, not on another Result's claim.
- **The rubric inside the recipe.** Rejected by the `tenant-lines` history: the verdict logic changed in 13 of 14 versions while the procedure stayed stable, so each rule change forked the procedure too.
- **The rubric as prose only.** Rejected for ADR 0004's reason: prose cannot close thresholds that overlap or leave gaps, and known answers can. Prose stays for verdicts decided by reading.
- **Remodel only the measuring shape now.** Rejected by the owner: it would leave the other shapes on the old model and a gap between them.
