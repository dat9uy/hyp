---
name: __NAME__
version: 1
format: 1
subject: hyp:placeholder the measures this rubric reads, such as "the stdout of tenant-lines@3 check"
---

# __NAME__

hyp:placeholder This rubric came from hyp bundle new --rubric. Replace
every hyp:placeholder line, the stub rule.sh and the known-answers cases,
then run hyp bundle known-answers __NAME__@1 until it passes. hyp add exp
refuses the bundle while any file in it still holds hyp:placeholder.

A rubric turns measures into verdicts. It has no Launch, Doctor, Drive,
Evidence or Cleanup command and exactly one Standalone command, which reads
the records it is given and prints one "verdict: LABEL=WORD" line per
label, WORD being supports, contradicts or inconclusive. LABEL is the
rubric's own name for a rival; it names no tree id, and the runner types
each target's verdict into hyp record. Every label needs a case printing
supports and a case printing contradicts.

## Standalone

hyp:placeholder rule.sh is the rule the known answers test: the supports
case must print "verdict: a=supports", the contradicts case
"verdict: a=contradicts". Rename the label a and add a pair of cases per
label.

Run rule: ./rule.sh
External: TODO
