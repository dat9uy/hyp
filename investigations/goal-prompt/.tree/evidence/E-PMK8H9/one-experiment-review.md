F1. Label: G-908PRZ#4
Text: the changed bullet in Build Goals, together with "## Stop" (unchanged), which has no clause for an unreviewed candidate.
Finding: The review now comes before the gate's Result, so any fix commit made after it is a new candidate that needs its own gate Experiment and review. Stop step 1 says "A defect in text you changed and may edit ... fix it with a commit before the stop". The checkpoint template has no slot for a candidate without a review, so a run that follows the text has no way to make the stop checkpoint "name that candidate".
Case: N is 5. Result 4 is the gate-plus-review of C1. The review finds a defect in the SKILL text. The run commits C2 as the fix, and Result 5 ends the budget before C2 is gated. The run stops with C2 unreviewed. Its checkpoint says "Stopped: 5 Results. Waiting for the owner: ... Out-of-Goal findings: none", which does not name C2.

F2. Label: G-908PRZ#1
Text: "Review each candidate in its gate Experiment" and "Add no separate review Experiment".
Finding: Lines #1, #2 and #4 count a "review Result". The text records one mixed Result and never says that it is the review Result. Whether a gate Result that cites a review file counts is left to whoever reads the trial. "Its gate Experiment" is also not fixed: a candidate can have several Experiments (gate, then one per limit case through "add an Experiment with --aims"), and the text does not say which one holds the review.
Case: The run reviews C1 in the Experiment that closes a limit and not in the first gate Experiment. A reader who counts only gate Results finds no review Result for C1 and fails line #1 although a review exists.

F3. Label: G-908PRZ#1
Text: the claim "a candidate costs one Result", and "the review's findings sorted as below".
Finding: The sort rule makes each finding that is a case of a line a `--limit`, and Reassess step 4 closes a limit only with a new Experiment and a new Result (`--closes`). A "fix commit" is a new Hypothesis with its own gate Result. A candidate whose review finds anything therefore costs more than one Result. The claim holds only for a clean review. The text does not warn that N=5 can end before the limits close, and it does not bound the chain of fix, gate and review. The Goal asks for no chain of reviews on small changes.
Case: C1 gets 2 case-of-line findings. Result 1 is the gate and the review. Results 2 and 3 close the limits. Result 4 gates fix commit C2, which carries a review of its own. Result 5 closes C2's limits. The budget is gone and the run has not yet selected.

F4. Label: G-908PRZ#3
Text: "Before you record the gate's Result, ask a fresh subagent for one review ...".
Finding: The review is required for every candidate, even one whose gate has already failed (red not red, a failing test). A candidate the gate contradicts and a fix supersedes still gets a full subagent review before its Result. No text lets the run skip the review, or run it only for a candidate that passed the gate, so cost is not reduced. The old "before the stop" wording at least allowed one review of the final candidate. The new wording makes the review per candidate with a gate.
Case: C1 fails its removal check. The run still spawns a review subagent on C1, which costs tokens, and then supersedes C1 with C2 and reviews C2 too. Two reviews are paid for, and the median cost is more than 1.5 times the base's on a cut where the base reviews once.

F5. Label: none
Text: "Before you run anything, write each Experiment's verdict rule in its body: ..." (unchanged) against the new bullet.
Finding: The verdict rule list names the gate checks but not the review. The new bullet says the Result carries "the gate's verdicts" plus the review's findings, but the Experiment body does not have to say how a finding changes a verdict. A run that follows the text has no pre-registered rule for whether a case-of-line finding turns a passing gate into `contradicts`. Under "A recipe measures and never prints a verdict" and "a script decides the verdict", the review is a manual input in a scripted Experiment.
Case: The recipe gate passes (`supports`) and the review finds a case of #2. The run records `supports` and a limit. A second run, on the same facts, records `contradicts`. The rule does not tell which is right.

F6. Label: none
Text: "--limit" use inside "record one Result that cites the gate's logs and the review".
Finding: `hyp record --help` shows `--limit` at most 10 per Result, each at most 300 runes, and `--evidence` at most 10 pointers. Gate logs plus a review file plus more than 10 findings do not fit in one Result. A review that returns 11 case-of-line findings, or one finding longer than 300 runes, cannot be recorded as the text says, and the text does not say what to do (merge them, drop them, or record a second Result, which breaks "one Result").
Case: A review labels 12 findings as cases of lines. The 11th `--limit` is refused. The run needs a second Result for the candidate.

F7. Label: none
Text: docs/glossary.md "A finding of a run's review that is a case of a line is a limit ..." and README.md line 10 / line 241 ("A self-review in the run recorded its findings as limits of a Result").
Finding: These texts still read correctly (a review's finding is a limit of a Result). No doc states "a separate review Experiment", so the change makes nothing in docs/, AGENTS.md or README.md wrong. hyp-owner-grill line 116 says the practice of a run, "a build Goal's review round included, is in hyp-goal-run", which also stays true. Every command and flag named in the changed line (`--limit`) exists in `hyp record --help`. No new `hyp` command was added, so skillcheck is not affected. The changed bullet is long (about 900 characters) but the front matter is unchanged, so G-908PRZ#5 is expected to hold; the quick_validate.py and lint_cruft.py runs were not done in this review.

Summary: the diff cuts the review Experiment but leaves a run free to end on an unreviewed fix commit with no checkpoint slot (F1), leaves "review Result" undefined (F2), and "a candidate costs one Result" is false whenever the review finds a case of a line (F3).
