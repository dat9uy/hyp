#!/usr/bin/env python3
"""verdict --row F [--row F]... SCORE...: the skill-text gate over every trial's score.

Each SCORE is the output of score for one trial: key<TAB>value lines with at
least arm, trial, skill_loaded and every row F (a fingerprint, 0 or 1). The
arm named base is the skills as they are; every other arm is a change. Each
arm must have exactly 5 trials with distinct trial numbers.

For each row and arm, k is the number of trials whose F is 1. A trial whose
skill_loaded is 0 never counts as a reproduction in base (its F is not read),
and always counts as a failure in any other arm.

  A row reproduces when base has k >= 3.
  An arm other than base passes a row when its k <= 1.

Prints, for each row and arm (base first, then the others sorted), a line
"ROW ARM: k of 5"; then

  verdict: repro=supports        when any row given reproduces, else contradicts
  verdict: ARM=supports          for each other arm that passes every reproducing row
  verdict: ARM=contradicts       when it fails one
  verdict: ARM=inconclusive      when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing field, a fingerprint
that is not 0 or 1, no base arm, an arm without exactly 5 trials, or a trial
number twice in an arm.
"""
import sys

TRIALS, REPRO, PASS = 5, 3, 1


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    args, rows, scores = sys.argv[1:], [], []
    i = 0
    while i < len(args):
        if args[i] == "--row":
            if i + 1 >= len(args):
                fail("--row needs a value")
            rows.append(args[i + 1])
            i += 2
        elif args[i].startswith("--"):
            fail(f"unknown flag {args[i]}")
        else:
            scores.append(args[i])
            i += 1
    if not rows or not scores:
        fail("usage: verdict --row F [--row F]... SCORE...")
    arms = {}
    for p in scores:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ["arm", "trial", "skill_loaded"] + rows:
            if k not in kv:
                fail(f"{p} has no {k}")
        for k in ["skill_loaded"] + rows:
            if kv[k] not in ("0", "1"):
                fail(f"{p}: {k} is {kv[k]!r}, not 0 or 1")
        t = arms.setdefault(kv["arm"], {})
        if kv["trial"] in t:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        t[kv["trial"]] = kv
    if "base" not in arms:
        fail("no trial of the base arm base")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = ["base"] + sorted(a for a in arms if a != "base")
    k = {}
    for r in rows:
        for arm in order:
            n = 0
            for kv in arms[arm].values():
                if kv["skill_loaded"] == "0":
                    n += arm != "base"
                else:
                    n += kv[r] == "1"
            k[r, arm] = n
            print(f"{r} {arm}: {n} of {TRIALS}")
    repro = [r for r in rows if k[r, "base"] >= REPRO]
    print(f"verdict: repro={'supports' if repro else 'contradicts'}")
    for arm in order[1:]:
        if not repro:
            word = "inconclusive"
        else:
            word = "supports" if all(k[r, arm] <= PASS for r in repro) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
