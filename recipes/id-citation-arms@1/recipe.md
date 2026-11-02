---
name: id-citation-arms
version: 1
format: 1
subject: Prototypes of #240's rules for write-command text that cites an id that is no node (today, text-warn, text-refuse, cite), run on a frozen hyp store, and e-arms' verdict rule (self-review, #240)
---

# id-citation-arms

e-arms' instrument for #240 (`investigations/id-citation`). `cases` runs
every case of a cases file with every arm binary, each on a fresh copy of a
fixture store, and prints what each arm did; `verdict` applies e-arms'
pre-registered rule to that table. Both run as checks (ADR 0003), from a
temporary copy, with every path an absolute argument. The arms are the
subject under test: `cases` runs only the binaries it is given and never
the `hyp` on PATH.

## Launch

None: both commands run with no sandbox.

## Doctor

None.

## Drive

None.

## Evidence

Gathered by hand: the cases file, the fixture archive and the arm
binaries' sha256, under `.tree/evidence/<E-id>/`.

## Cleanup

None.

## Standalone

Run cases: ./cases.py

Run verdict: ./verdict.py

- `cases CASES FIXTURE NAME=/abs/path...` prints a header and one row per
  arm and case: `arm case exit appended warned`, for the last step that
  ran. `appended` is `y` when the bytes under `fixture/.tree/events`
  changed during that step, and `warned` is `y` when its stderr holds a
  line starting with `warn:`. A case whose last step exits 3 runs once
  more. Each step's stderr goes to `$HYP_OUT/<arm>/<case>.<n>.stderr`.
- `verdict TABLE CASES` reads that table and the cases' classes (`slip`,
  `prose-slip`, `guard`, `cite-slip`, `cite-guard`), for the arms
  `today`, `text-warn`, `text-refuse` and `cite`, and prints a summary
  line per arm and `verdict: h-text-warn=…`, `verdict: h-text-refuse=…`,
  `verdict: h-cite=…` and `verdict: h-null=…` (the rule is in the
  script's docstring). `known-answers/` holds a supports and a
  contradicts case for each label.
