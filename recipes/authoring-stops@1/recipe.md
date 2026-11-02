---
name: authoring-stops
version: 1
format: 1
subject: Liveness of each method-to-instrument stop on gh:235, probed against the hyp first on PATH and a repository under study (self-review, #235)
---

# authoring-stops

e-classify's instrument for #235 (`investigations/recipe-authoring`).
`probes` runs one probe per stop: each probe is the smallest hyp, git or go
command that hits the stop, run in a scratch directory under `$TMPDIR`, and
it prints what happened. Three probes read the repository under study
instead. The classing itself (binary, skill text, no change) is a reading of
this table by the AGENTS.md rule, kept as evidence; the probes say whether
each stop is still live, and `verdict` decides the rivals from the two
refusal stops they name.

The probes run the hyp that the gate copies first on `PATH`, so the known
answers pin that hyp's refusal messages: a hyp whose messages differ fails
the gate, and a new version records the new messages.

## Launch

None: the command runs with no sandbox.

## Doctor

None.

## Drive

None.

## Evidence

Gathered by hand: the command record of the run, kept under
`.tree/evidence/<E-id>/`.

## Cleanup

None.

## Standalone

Run probes: ./probes.sh

Run verdict: ./verdict.sh

- `probes REPO` prints a header and one line per probe:
  `stop<TAB>exit<TAB>output`, where output is the first line the probe
  printed, with its scratch directory written `W`, `$HOME` written `HOME`
  and a trailing commit id written `BASE`. It exits 2 when REPO is missing
  or is not a git work tree, and 0 otherwise.
- Probes on hyp: `measured-none` (add exp, recipe none, no declaration),
  `no-known-answers` (add exp on an instrument with no case),
  `evidence-plain-dir` (recipe evidence over a directory no sandbox wrote),
  `at-in-input` (recipe hash of a bundle whose known-answer input holds an
  `@` path), `placeholder-id` (`R-<new>` in a --why), `alias-held` (a
  held alias), `first-bundle` (recipe new with no --from),
  `unreachable-base` (sandbox create on a commit no ref reaches, local
  host).
- Probe on go: `go-file-in-bundle` (go vet ./... in a module holding
  `recipes/x@1/a.go`; needs go on PATH).
- Probes on REPO: `tracked-trees` (directories under `investigations/`
  and `fixtures/`), `bundles-without-format` (recipe.md files with no
  `format: 1` line), `template-sections` (lines of the recipe template
  naming a Standalone section or `known-answers`).

- `verdict TABLE` reads a probes table and prints `live refusal stops: N of 2`
  and one `verdict:` line each for `h-binary`, `h-partial` and `h-null`
  (the rule is in the script's header): a refusal stop is live when its
  probe exits non-zero, and the rivals name two, `evidence-plain-dir` and
  `at-in-input`. It exits 2 when TABLE is missing or lacks either row.

For `probes`, `known-answers/` holds a clean repository, a legacy one (an old fixture
directory, a bundle with no format line, a template with both sections)
and a usage failure. Each input repository is a git bundle of one commit, holding HEAD and main,
made by hand with fixed author dates; the expected tables were written
from the probes' rule and hyp's documented refusals, not captured.
For `verdict`, it holds a table with both, neither and one refusal stop
live, and one missing a row.
