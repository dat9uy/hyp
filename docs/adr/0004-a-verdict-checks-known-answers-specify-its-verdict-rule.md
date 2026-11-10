---
status: accepted; amended by ADR 0018: the verdict check is a rubric's command, and a recipe prints measures only
---

# A verdict check's known answers specify its verdict rule

A verdict rule is the author's prose, written before the run, that says which outcome gives which verdict. hyp stores it and never reads it. An earlier change tried to make that prose close by guidance alone: its skill-text gate failed 5 of 5 trials in both arms on the judge's material-point rule, and most material points were verdict logic, such as thresholds that overlapped or left a gap and terms a threshold relied on without defining. Prose cannot be made to close those; a check run against known answers can, because each case pins an input to the verdict it must give.

Two facts made the gap visible in hyp itself. The session-replay recipe's `score.py` printed a verdict line that nothing read, so deleting it changed nothing hyp does. And `WeakCases`, behind `hyp recipe check`'s warnings, asked each check only for one case expecting exit 0 and one expecting a non-zero exit, so `score.py`'s contradicts and inconclusive branches had no case and no warning said so. Owner decision on 2026-10-31.

## Decision

- **Recognised by its known answers.** A check (an Evidence or Standalone command) is a **verdict check** when any of its known-answer cases expects, in its `stdout`, a line starting with `verdict:`. There is no new recipe syntax. A check whose cases show no such line gets only the existing must-pass and must-fail warnings.
- **One line per label.** A verdict line is a whole line of stdout, `verdict: LABEL=WORD`. LABEL is the check's own name for a rival and has the shape of an alias (`[a-z][a-z0-9-]{0,31}`); WORD is a Result verdict, `supports`, `contradicts` or `inconclusive`. A case prints each label at most once. hyp never binds a label to an H-id: the operator still types each verdict into `hyp record`.
- **Coverage per check and label.** For each Evidence command and each Standalone label, every label its cases print needs a case printing `supports` and a case printing `contradicts`. `inconclusive` needs no case. The must-pass and must-fail warnings stay; a must-fail case means bad input, not a verdict.
- **Warnings only.** `hyp recipe check` warns, without changing its exit code: `<command>: verdict <label>: no supports case`, `<command>: verdict <label>: no contradicts case`, `<command>: case <name>: malformed verdict line "<line>"` and `<command>: case <name>: label <label> printed twice`. The gate, `hyp recipe run` and `hyp recipe evidence` are unchanged and never read a verdict line.

The verdict rule of a verdict check lives in the check and its cases. This decision and the docs describe the mechanism; they do not restate any recipe's thresholds.

## Considered Options

- **A declaration in `recipe.md`** naming the verdict checks and their labels. Rejected: a second place that can disagree with the cases, when the cases already show which checks print verdicts.
- **A recognised command name**, such as a Standalone label `score`. Rejected: a naming convention hyp would have to enforce, and one recipe may have several verdict checks.
- **Roles or H-ids in the line.** Rejected: no binding is needed while comparing a typed verdict with the check's output stays out of scope.
- **A containment reason that compares a Result's typed verdicts with the check's output.** Parked as an open issue about how `wasted` counts labels: it adds a rule after the fact, and the label-to-H-id binding it needs is its own question.
- **Refusing a run whose verdict check lacks coverage.** Rejected: the existing must-pass and must-fail rule warns, and a verdict check follows it.

## Consequences

- Deleting a case that covers a label's supports or contradicts makes `hyp recipe check` warn, so hyp now depends on the verdict lines it once ignored.
- The ADR 0002 limit carries over: the gate proves a check reproduces its cases, not that the cases are the right specification. Whether each label's cases match the Hypothesis it names was the verifier's question; since ADR 0015 retired the verifier, it is read by the owner at `choose`, and a case changed after a candidate was seen is ADR 0015's third loss.
- The session-replay recipe's Draft version prints `verdict: h-step=WORD` and `verdict: h-same=WORD` and holds a case for each of its branches. Older bundles are hashed only and do not change.
- Out of scope: comparing a Result's typed verdict with the check's output, trial counts in the check's product, and reading verdict lines from `$HYP_OUT` or at run time.
- The glossary, `RECIPE-FORMAT.md`, the spec §7.6 and the hyp-prototype `VERIFIER.md` change to match. Each change is a correction that cites this decision. ADR 0001 to 0003 stand.
