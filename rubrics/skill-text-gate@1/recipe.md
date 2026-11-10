---
name: skill-text-gate
version: 1
format: 1
subject: The score files of one skill-text replay, five trials per arm, as `score` of recipes/skill-text-arms@5 prints them (self-review)
---

# skill-text-gate

The skill-text gate (AGENTS.md, "Improving hyp with hyp") as a rubric. It
turns the measures of `recipes/skill-text-arms@5`, one score file per
trial, into the gate's verdicts. Its script and cases came unchanged from
the `verdict` command of `recipes/skill-text-arms@4`, which printed
verdicts from a recipe; a recipe now measures and never decides
(ADR 0018).

## Standalone

Run verdict: ./verdict.py
External: none; reads only the score files given as arguments

- `verdict --row ROW [--row ROW]... SCORE...` applies the gate to every
  trial: per row, base reproduces at 3 or more of 5 and a text arm passes
  at 1 or fewer of 5; a trial with `skill_loaded` 0 counts against its arm
  (never a reproduction for base, always a failure for a text arm). It
  prints `verdict: repro=…` and `verdict: none=…` (any row reproduces, or
  none does) and `verdict: ARM=…` for each text arm. It exits 2 when a
  field or the base arm is missing, or an arm has other than 5 trials. The
  script's docstring states the rule.

The labels are the rubric's own: `repro`, `none`, and each text arm's
name. The runner binds each to a target when it records the Result.

`known-answers/` holds a reproducing base (`repro=supports`), a base
whose third failure loaded no skill (`repro=contradicts`), a passing and
a failing text arm, an arm with four trials and a missing field. Every
expected output was written by hand from the rules above.
