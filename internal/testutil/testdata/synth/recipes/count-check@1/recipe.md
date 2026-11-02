---
name: count-check
version: 1
format: 1
subject: A synthetic instrument for hyp's own tests; it judges a count against a limit of 2
---

# count-check

A verdict check for the synthetic tree in `internal/testutil/testdata/synth`.
It takes one count and prints `verdict: short=supports` when the count is 2
or less, and `verdict: short=contradicts` when it is more. Anything that is
not a whole number exits 2.

## Standalone

Run judge: ./judge.sh
