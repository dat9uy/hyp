# e-arms: method (pre-registration)

Experiment `e-arms` in this repository's tree (`.tree/`), under DQ `q-rule`
(Q-DVZZHY). Parent h-text-refuse (H-88ZF3E); targets h-text-warn
(H-RAJAFK), h-cite (H-C5BA8J) and h-null (H-BMN916). This method, the
fixture and the cases file are given before the Experiment is added.

## Instrument

`id-citation-arms@1`, two Standalone commands, run with
`hyp recipe run --record`:

- `cases CASES FIXTURE today=… text-warn=… text-refuse=… cite=…` copies
  the fixture store for every arm and case, runs the case's steps with
  that arm (as `step1 && step2`), and prints
  `arm case exit appended warned` for the last step that ran. It runs only
  the arm binaries it is given, by absolute path, never the `hyp` on PATH.
- `verdict TABLE CASES` reads that table and prints one verdict per rival.

## Inputs

- `fixture/fixture.tar` (sha256
  `2ff207cbe77bc6c835b1672808ff3ad3ccfc0db69467bc189e55e06885491dab`): `fixture/`, a store
  with a Goal, a DQ, two Hypotheses (aliases h-alpha, h-beta), an
  Experiment, Result R1 (R-61SFBW), a select and an insufficient verifier
  confirm; and `other/`, a throwaway store whose Result R-64ZZ43 is K6's
  cite. Built once with the operator hyp; ids are in `fixture/ids.txt`.
- `fixture/cases.tsv` (sha256
  `c7a018ce6a199610ab6a420d186fde5cead9b376fe1f3b0704f72ba6ddb2a3ee`):
  - K1 (slip): the rebuilt chain `hyp record … && hyp request-verify Q --why "…"`,
    the why copied verbatim from the `dq.request_verify` event of
    `investigations/containment/.tree/events/claude-234-r2-run.jsonl` at
    97219b1 ("… R-<new> shows declared honest reading Results clear on both
    columns").
  - K2 to K4 (slip): placeholders `E-…` in `--why`, `R-XXXXXX` in
    `--observation`, `H-<id>` in `--body`.
  - K5 (slip): R1 with its last character changed (R-61SFBZ).
  - K6 (slip): R-64ZZ43, a real Result id of the other store, no prefix.
  - K7 (prose-slip): "the new Result shows …", no id. Reported, never scored.
  - G1 to G5 (guard): R1's display id; R1's full ULID;
    `other-tree:R-ABC123`; "E-mail, R-squared and H-index"; the alias
    h-alpha.
  - C1a, C1b (cite-slip) and C1c (cite-guard): `request-verify --result`
    with R-<new> (in the K1 chain, verbatim why), the typo, and R1.
- Arms, named by sha256 in `arms/arms.sha256`: `today` (main ff47d32, no
  patch), `text-warn`, `text-refuse` and `cite` (ff47d32 plus a prototype
  patch), built already as `arms/arm-<arm>`.

## Verdict rules (zero tolerance)

- **h-text-warn** supports if, on every slip case K1 to K6, its arm exits
  0, appends and warns, and on every guard it exits 0, appends and does
  not warn. Otherwise it contradicts.
- **h-text-refuse** supports if, on every slip case K1 to K6, its arm exits
  non-zero and appends nothing, and on every guard it exits 0, appends and
  does not warn. Otherwise it contradicts.
- **h-cite** supports if, on C1, its arm refuses both slips with nothing
  appended and accepts the guard, and its rows on G1 to G5 equal today's.
  Otherwise it contradicts. The leftover `R-<new>` in C1a's `--why` is a
  recorded limit, not a miss.
- **h-null** contradicts if the today arm appends any slip case K1 to K6
  with exit 0 and no warning. Otherwise it supports.
- An arm whose scored case exits 3 after one re-run is inconclusive. If two
  or more rivals support, the result is a trade-off for the owner, with no
  default.

`verdict.py`'s docstring states the same rule; its known answers hold a
supports and a contradicts case for every label.

## Limits to record on the Result

1. The fixture is hand-made apart from K1, and small: one store, one
   command (request-verify) for most cases.
2. K7 is unscored for the scan arms: a scan cannot see an id that is not
   written.
3. The arms are prototypes; the selected arm must later be reproduced in
   Go and give its rows on every case.
4. h-cite's fixture covers request-verify only; select and confirm take
   `--result` in the arm but no case runs them. Under h-cite the why still
   holds `R-<new>` (C1a).
5. The arm binaries are read from absolute paths outside the bundle, which
   the gate cannot see (RECIPE-FORMAT, The gate).
