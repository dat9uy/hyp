#!/usr/bin/env python3
"""verdict --row ROW [--row ROW]... SCORE...: the skill-text gate over score outputs.

ROW is a fingerprint name, f_reading_pin or f_body_limit (any f_ field the
scores hold). Each SCORE is a file holding one trial's score lines. Every
arm, base and each variant, must have exactly 5 trials with distinct trial
numbers, and base must be there.

For each row and arm, k is the number of trials whose fingerprint is 1. A
trial with skill_loaded other than 1 counts against its arm:

  base         a trial that loaded no skill is never a reproduction
               (k counts skill_loaded 1 and fingerprint 1)
  variant arm  a trial that loaded no skill always counts as a failure
               (k counts fingerprint 1 or skill_loaded 0)

A row reproduces when base has k >= 3. A variant arm passes a row at k <= 1.

  verdict: repro=supports       when any given row reproduces, else contradicts
  verdict: ARM=supports         a variant arm that passes every reproducing row
  verdict: ARM=contradicts      one that fails a reproducing row
  verdict: ARM=inconclusive     every variant arm when no row reproduces

Prints one line per row and arm, then the verdict lines, variant arms sorted.
Exits 2 on a usage error, an unknown row, an unreadable file, a missing
field, an arm that is not a label or is named repro, no base arm, or an arm
without exactly 5 distinct trials.
"""
import re
import sys

TRIALS = 5
LABEL = re.compile(r"^[a-z][a-z0-9-]{0,31}$")


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    args, rows = sys.argv[1:], []
    while len(args) >= 2 and args[0] == "--row":
        if not re.match(r"^f_[a-z0-9_]+$", args[1]):
            fail(f"unknown row {args[1]} (want a fingerprint name such as f_body_limit)")
        if args[1] not in rows:
            rows.append(args[1])
        args = args[2:]
    if not rows or not args:
        fail("usage: verdict --row ROW [--row ROW]... SCORE...")
    need = {"arm", "trial", "skill_loaded"} | set(rows)
    arms = {}
    for p in args:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        if not need <= kv.keys():
            fail(f"{p} lacks " + ", ".join(sorted(need - kv.keys())))
        a = kv["arm"]
        if not LABEL.match(a) or a == "repro":
            fail(f"{p} names arm {a!r}")
        arms.setdefault(a, {})
        if kv["trial"] in arms[a]:
            fail(f"{p} repeats {a} trial {kv['trial']}")
        arms[a][kv["trial"]] = kv
    if "base" not in arms:
        fail("no base trial")
    for a, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {a} has {len(t)} trials, want {TRIALS}")

    def count(a, row):
        if a == "base":
            return sum(1 for kv in arms[a].values() if kv["skill_loaded"] == "1" and kv[row] == "1")
        return sum(1 for kv in arms[a].values() if kv[row] == "1" or kv["skill_loaded"] != "1")

    variants = sorted(a for a in arms if a != "base")
    repro, passes = [], {a: True for a in variants}
    for r in rows:
        k = count("base", r)
        ok = k >= 3
        print(f"{r} base: {k}/{TRIALS} " + ("reproduces" if ok else "does not reproduce"))
        if ok:
            repro.append(r)
        for a in variants:
            ka = count(a, r)
            print(f"{r} {a}: {ka}/{TRIALS} " + ("passes" if ka <= 1 else "fails"))
            if ok and ka > 1:
                passes[a] = False
    print("verdict: repro=" + ("supports" if repro else "contradicts"))
    for a in variants:
        word = "inconclusive" if not repro else ("supports" if passes[a] else "contradicts")
        print(f"verdict: {a}={word}")


if __name__ == "__main__":
    main()
