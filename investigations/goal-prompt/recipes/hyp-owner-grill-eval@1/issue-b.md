# Issue #141: Run practice: the last candidate of a budget-bound run goes unreviewed

Opened by dat9uy at 2026-11-12T02:21:06Z. Text as of 2026-11-12T03:45:00Z.

## Event

On 2026-11-12, run 1 of `investigations/goal-prompt` (G-ZBM2RE, PR #140) stopped with its last candidate, H-YXSJEM at `5c5902d`, never reviewed. The two earlier candidates were reviewed, and each review found real defects (10 findings on `10566d4`, 8 on `b4f23b2`). The stop checkpoint says "all review limits closed" and does not name the missing review.

The chain, from the run's transcript and the two trees:

1. A review finding that is a case of a line becomes a limit, and `--limit` exists only on `hyp record`. So the runner recorded each review as its own Experiment: E-C3GGGJ reviewed `10566d4` and E-7QQ2DP reviewed `b4f23b2`.
2. Each candidate therefore cost two Experiments, a gate and a review. The run spent E-HQWTA0, E-C3GGGJ, E-Z5BTZR, E-7QQ2DP and E-NZR3GV (the gate of `5c5902d`), and the stop clause "5 Experiments have been recorded" ended it.
3. The default of 5 was set from G-D1CH1R's runs, where every Experiment gates a commit and review findings went straight into fix commits (one Experiment per candidate; run 5 used 4).
4. At 17:21:37 the runner wrote "E5 is the last Experiment the budget allows, so this fix commit must close everything." `skills/hyp-goal-run/SKILL.md` asks for "one review of each candidate commit's diff" before the stop. The prompt's stop clause and the skill conflict, and neither says which yields or that the gap is named at the stop.

## Scope

Core: the run practice of a Goal whose candidate is a commit (review, budget and stop report) in `hyp-goal-run`, and the stop clause `hyp prompt` prints. No self-review recipe is involved.

## Rivals

A tree holds this question; each option below is a rival Hypothesis, not a decision. The scenario that separates them: replay the G-ZBM2RE cut after R-8JQPWG (one Experiment left, a fix commit pending) and see whether the run stops with every candidate reviewed or with the gap named.

- **H-one-experiment:** a candidate's review is part of its gate Experiment. One Result carries the test verdicts and the review's limits, so a candidate costs one Experiment again.
- **H-name-at-stop:** keep the practice; the stop checkpoint names each candidate without a review, so the gap is visible to the owner.
- **H-reserve-last:** when one Experiment remains, the runner reviews the current candidate instead of committing another fix, and stops with recorded limits.
- **H-raise-budget:** the default of `--experiments` rises (for example to 7) to fit a gate and a review per candidate.

These can combine (H-name-at-stop with any other); the tree should say which hold alone.

## Done means

- A tree under this issue selects a rival, or a combination, from a run on the scenario above.
- After the change, a run that reaches its budget either has a review Result for its final candidate or names that candidate as unreviewed in the stop checkpoint.

Part of #113. Observation log: run 1, entry R10.

---

Comment by the owner at 2026-11-12T03:04:10Z:

Second event, 2026-11-12: run 2 of `investigations/goal-prompt` (actor `claude-g113-2`, session `d3fe6332`) ended the same way, with its last candidate unreviewed.

It started with a fresh budget of 5 (`hyp prompt`'s clause "5 Results recorded under the actor") and spent it as:

1. R-0E8PFH applies the owner's six answers.
2. R-KYKWH9 reviews `5c5902d`, closing run 1's gap; 1 limit.
3. R-28A8NB gates the fix commit `2284172`.
4. R-3YJ7T2 reviews `2284172`; 2 limits, and the run's own probes of both cases pass.
5. R-356DV0 gates `3cfd08d`, a commit that adds tests and spec text for those two passing cases.

`3cfd08d` (H-FQ103R) is unreviewed. Unlike run 1, the stop names the gap: "3cfd08d needs a fresh review of its diff, and the cap left no room to run one."

What this adds to the rivals:

- **Against H-raise-budget.** A fresh budget did not help. Each review so far has led to a new commit (findings 10, 8, 1, 2), and each commit owes its own review, so any fixed budget ends on an unreviewed commit unless a review leads to no commit.
- **For H-reserve-last.** The last commit was not needed: both limits' probes passed on `2284172`, so a Result closing them on H-YTD2JG followed by the select would have ended on a reviewed candidate. A related rule to test with it: a case whose check already passes closes as a limit, not with a new commit.
- **H-name-at-stop** happened without a rule this time; the scenario should still check whether it holds when the runner is not prompted by an earlier gap.

Observation log entry R14.

---

Comment by the owner at 2026-11-12T03:19:12Z:

Premise behind the rivals: no text says a review is its own Experiment.

What is written:

- The glossary's Verifier entry (ADR 0015): "Any actor may still review: a defect is recorded as a Result, a retract or a limit."
- replay-spec §2: "a Result names the limits it found", and an Experiment is "a pre-registered Result" (ADR 0018). So a finding recorded as a limit needs a Result, and a Result needs an Experiment.
- The sort rule "a case of a line is a `--limit`". On `main` it is in no skill: it came from `hyp-owner-grill`'s old hand-off, which told the grill to write it into the `/goal` prompt. On `feat/goal-prompt`, run 1 moved it into `skills/hyp-goal-run/SKILL.md` (lines 34 and 35) in its own candidate commit `10566d4`.

What no text says, and run 1 chose (run 2 copied):

- A review is a separate Experiment (E-C3GGGJ, E-7QQ2DP, E-H98Z3B, E-2A0909: `--recipe none --rubric reading`, the candidate as target, probes as evidence), not part of the candidate's gate Experiment.
- That Experiment, and its Result, counts against the run's budget.
- Its Result gives a verdict on the candidate.

So the only written review practice is text the candidate under test wrote, unreviewed and unmerged. Whichever rival this issue selects, `hyp-goal-run` should state how a review is recorded and whether it counts against the budget, so the next run does not invent it again.

---

Comment by the owner at 2026-11-12T03:25:36Z:

Run 3 (actor `claude-g113-3`, session `e2fd7562`, still running): two more rivals, from how its reviews turned into commits.

What happened:

- Review of `3cfd08d` (R-JHBCWS): "no defect"; N1, a case of line 3 that no test runs. The runner probed it, the code passed, it was recorded as a limit, and the fix commit `d2d5d14` added a test.
- Review of `d2d5d14` (E-ZMMVPN): "no defect"; P1, rated low by the reviewer ("Mutation D needs a deliberate split"), a case the real code passes. Probed (`probe-p1.sh`), recorded as a limit, fix commit `04142f3`.
- Each brief asked the reviewer to review "the diff of commit X against its parent" (one test assertion) and "does the new assertion go red when the behaviour … is broken?". A test-strength question about a one-assertion diff almost always finds one more mutation that slips past, so every review so far has yielded a commit.
- For `04142f3` the runner put the gate and the review into one Experiment (E-GE085A), which is H-one-experiment arising unprompted.

The reviewer itself was rigorous: mutations in a scratch copy, each confirmed by running, severities stated, "no other defect found" said plainly.

Two rivals to add:

- **H-probe-closes:** a case the run's own probe runs and passes on the candidate is closed in the review's Result (a check ran it), not recorded as a limit. A limit stays for a case no check ran, or one the candidate fails. Under this rival N1 and P1 close on `3cfd08d`, and no commit follows.
- **H-whole-candidate:** the review before the select reads the cumulative diff from the Goal's base, not the latest increment. Since `10566d4` every review has read only the last fix, so the code at the final candidate has not been read in one piece against the lines.

Both fit the scenario already in the issue body; a second scenario for H-probe-closes is run 3's cut after R-JHBCWS.
