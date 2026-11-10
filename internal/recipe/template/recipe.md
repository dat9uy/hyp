---
name: __NAME__
version: 1
format: 1
subject: hyp:placeholder what is under test, such as "the hyp binary on PATH" or "a Rails app with seed data"
---

# __NAME__

hyp:placeholder This bundle came from hyp bundle new. Replace every
hyp:placeholder line, the stub check.sh and the known-answers cases, then
run hyp bundle known-answers __NAME__@1 until it passes. hyp add exp
refuses the bundle while any file in it still holds hyp:placeholder.
RECIPE-FORMAT.md in the hyp-verification skill describes each section.

A command is one line, "Run: ./script WORDS" in Launch, Doctor, Drive,
Evidence and Cleanup, or "Run NAME: ./script WORDS" in Standalone. The
script lives in this bundle and is executable. hyp runs every command;
nothing here is run by hand. Delete the guidance of a section you leave
without commands, but keep its heading.

## Launch

hyp:placeholder How to start the subject in a sandbox, with exact
commands and the ports or files they use. Leave it without commands when
the subject needs no start.

## Doctor

hyp:placeholder The checks that must pass before driving, and what each
prints when healthy.

## Drive

hyp:placeholder The steps of the Experiment, in order.

## Evidence

hyp:placeholder Which files to save under .tree/evidence/<E-id>/. A
command here reads the files hyp sandbox remove kept, after Cleanup, and
gets their directory as its last argument.

Each Evidence and Standalone command states what it reaches outside the
bundle on the line directly under it: its subject arguments and any
outside reads or writes, or "External: none". hyp bundle check warns
about "External: TODO" until it is replaced.

## Standalone

hyp:placeholder Commands that run with no sandbox. check.sh is the check
the known answers test: known-answers/must-pass/ must give exit 0 and
stdout "pass", and known-answers/must-fail/ exit 1 and stdout "fail".
Rename it, add commands, and give each command a must-pass and a
must-fail case.

Run check: ./check.sh
External: TODO

## Cleanup

hyp:placeholder How to stop everything Launch started.
