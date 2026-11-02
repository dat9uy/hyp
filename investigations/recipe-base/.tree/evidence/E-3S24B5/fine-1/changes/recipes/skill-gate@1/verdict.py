#!/usr/bin/env python3
"""verdict ROW... -- SCORE...: the skill-text gate over the score of every trial.

Each SCORE is a file of key<TAB>value lines for one trial (score's output),
with at least arm, trial, skill_loaded and every ROW. A ROW is the name of a
0/1 fingerprint, such as f_reading_pin; skill_loaded is 0 or 1 too. The
arm named base is the skills as they are; any other arm is a variant, and its
name must be alias-shaped ([a-z][a-z0-9-]{0,31}) and not repro. Every arm
must have exactly 5 trials with distinct trial values, and base must exist.

For a ROW and an arm, k is the number of the arm's trials that fail:
  base       fingerprint is 1 and skill_loaded is 1 (a trial that loaded no
             skill is never a reproduction)
  a variant  fingerprint is 1 or skill_loaded is 0 (a trial that loaded no
             skill always counts as a failure)

A row reproduces when base has k >= 3. A variant passes a row at k <= 1.

  verdict: repro=supports        when any given row reproduces
  verdict: repro=contradicts     when none does
  verdict: ARM=supports          a variant that passes every reproducing row
  verdict: ARM=contradicts       a variant that fails one reproducing row
  verdict: ARM=inconclusive      every variant, when no row reproduces

Before the verdict lines it prints one line per row and arm, base first and
the other arms sorted: "ROW ARM: k of 5". Exits 2 on a usage error (no row,
no score, a row given twice, no "--"), an unreadable file, a missing key, a
value that is not 0 or 1, a bad arm name, no base arm, or an arm without
exactly 5 distinct trials.
"""
import re
import sys

TRIALS, REPRO, PASS = 5, 3, 1
ALIAS = re.compile(r"[a-z][a-z0-9-]{0,31}")


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def bit(p, kv, key):
    if key not in kv:
        fail(f"{p} has no {key}")
    if kv[key] not in ("0", "1"):
        fail(f"{p}: {key} is {kv[key]!r}, not 0 or 1")
    return kv[key] == "1"


def main():
    args = sys.argv[1:]
    if "--" not in args:
        fail("usage: verdict ROW... -- SCORE...")
    i = args.index("--")
    rows, scores = args[:i], args[i + 1:]
    if not rows or not scores:
        fail("usage: verdict ROW... -- SCORE...")
    if len(set(rows)) != len(rows):
        fail("a row is given twice")
    arms = {}
    for p in scores:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ("arm", "trial"):
            if not kv.get(k):
                fail(f"{p} has no {k}")
        arm, trial = kv["arm"], kv["trial"]
        if arm != "base" and (not ALIAS.fullmatch(arm) or arm == "repro"):
            fail(f"{p}: arm {arm!r} cannot be a verdict label")
        loaded = bit(p, kv, "skill_loaded")
        fails = {}
        for r in rows:
            f = bit(p, kv, r)
            fails[r] = (f and loaded) if arm == "base" else (f or not loaded)
        t = arms.setdefault(arm, {})
        if trial in t:
            fail(f"arm {arm} has trial {trial} twice")
        t[trial] = fails
    if "base" not in arms:
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = ["base"] + sorted(a for a in arms if a != "base")
    k = {(r, a): sum(1 for f in arms[a].values() if f[r]) for r in rows for a in order}
    for r in rows:
        for a in order:
            print(f"{r} {a}: {k[r, a]} of {TRIALS}")
    reproducing = [r for r in rows if k[r, "base"] >= REPRO]
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for a in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(k[r, a] <= PASS for r in reproducing) else "contradicts"
        print(f"verdict: {a}={word}")


if __name__ == "__main__":
    main()
