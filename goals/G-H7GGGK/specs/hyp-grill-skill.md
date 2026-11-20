# Spec: the skill hyp-grill

The build of Goal G-H7GGGK, line-1 (gap): "A user of hyp in another repository
can grill and re-grill a Goal with no prompt written by hand."

No state machine is added, so this spec has no TLA+ model. The skill is text and
one script; hyp's state machine stays the one in the repository's spec.

## Parts

| File | Holds |
|---|---|
| `skills/hyp-grill/SKILL.md` | The grill and the re-grill: when to use; what hyp checks and the rule "An agent runs a line hyp printed. A skill names no write command"; decisions; how to grill; how to write examples; how to re-grill; the hand-off; done when; resources. |
| `skills/hyp-grill/run-rules.md` | The run part: the run rules hyp cannot check, for the runner. SKILL.md names it twice. |
| `skills/hyp-grill/probe/run` | The generic Probe, POSIX `sh`, mode 755: runs `$HYP_RUN` in its working directory, prints its stdout (adds a newline when it does not end in one), then `exit N`. A byte copy of the Probe the examples of this Goal use. |

## Which rule lives where

| Rule | File, section |
|---|---|
| A decision is the words, a line or an example | SKILL.md, Decisions |
| A decision no example checks is a gap the grill reports under "Not covered" | SKILL.md, Decisions |
| Product questions only; mechanics are the runner's, vetoed at merge | SKILL.md, How to grill, step 2 |
| An option's description says the build and the cheapest build | SKILL.md, How to grill, step 3 |
| The picked label is the recorded text, exact words | SKILL.md, How to grill, step 4 |
| A correction is not consent | SKILL.md, How to grill, step 5 |
| A missing name is proposed for the owner's veto | SKILL.md, How to grill, step 6 |
| A form question names every state and every command | SKILL.md, How to grill, step 7 |
| Surfaces before examples | SKILL.md, How to write examples, step 1 |
| Copy the shipped generic Probe | SKILL.md, How to write examples, step 3 |
| A check can fail: gap fails on H0, stand-in is met | SKILL.md, How to write examples, step 4 |
| `hyp try` before place | SKILL.md, How to write examples, step 5 |
| Re-grill: adjust adds only floors or ceilings; a new line or a gap after a select is a new Goal | SKILL.md, How to re-grill |
| Run only printed lines; `hyp try --worktree` before a run; a refusal names its guard; keep every tree run; leave examples; commit by name; select at once; stop when hyp stops; PR body is the select's report; outside defects listed | run-rules.md, rules 1 to 10 |

The skill shows only the reads `hyp help`, `hyp status`, `hyp try` and
`hyp show` in code, and names no write command.

## What each example checks

All three run the same `check.pl` over the `.md` files under `skills/`.

- **line-1/a**: folder `skills/hyp-grill` with a SKILL.md; the seven rules of
  its `input/rules.txt` each in one paragraph, list item, heading or table row;
  no citation of a file under the repository's docs or plans folders, an
  architecture decision record, an issue number or a Goal of this repository;
  relative links resolve inside the skill folder; some hyp command shown in
  code, each one listed by the built `hyp help`; an executable file of the skill
  works as the generic Probe.
- **line-1/b**: a `.md` file beside SKILL.md that SKILL.md names (run-rules.md);
  it shows a hyp command in code; no other skill folder repeats a paragraph of it.
- **line-1/c**: folder `skills/hyp-grill`; the two decision rules.

## Tests changed

`internal/skillcheck`: both tests read every file under `skills/` at any depth,
not only `skills/*/*`; the floor of 20 commands, written for the five v1 skills,
becomes "at least one". An unknown hyp command still fails the test.
