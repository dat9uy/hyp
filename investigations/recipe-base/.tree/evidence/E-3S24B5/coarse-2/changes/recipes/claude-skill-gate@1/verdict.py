#!/usr/bin/env python3
"""verdict --row ROW [--row ROW]... SCORE...: the skill-text gate over every trial's score.

Each SCORE is a file of score's key<TAB>value lines for one trial. A ROW is
a fingerprint key of those files (f_reading_pin, f_body_limit, ...); every
file must hold arm, trial, skill_loaded and each ROW, each 0 or 1 (skill_loaded
0 or 1, arm and trial non-empty). The arm named base is required, and every
arm must have exactly 5 trials with distinct trial values.

For each row and arm, k is the number of trials that fail the row: a trial
whose ROW is 1 and, in base, whose skill_loaded is 1 (a base trial that
never loaded a skill is never a reproduction); in an arm other than base,
also a trial whose skill_loaded is 0, which always counts as a failure.

  A row reproduces when base has k >= 3.
  A variant arm passes a row when its k <= 1.

Prints one "ROW ARM: failing k of 5" line per row and arm (rows in the order
given, base first, then the other arms sorted), then

  verdict: repro=supports       when any row reproduces, else
  verdict: repro=contradicts
  verdict: ARM=supports         for each non-base arm that passes every reproducing row
  verdict: ARM=contradicts      when it fails one
  verdict: ARM=inconclusive     when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing or non-0/1 field, no
base arm, or an arm without exactly 5 distinct trials. The thresholds are the
pre-registered ones and are not arguments.
"""
import argparse
import sys

TRIALS, REPRO, PASS = 5, 3, 1


class Parser(argparse.ArgumentParser):
    def error(self, message):
        fail(message)


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def fails(f, row, arm):
    if arm == "base":
        return f[row] and f["skill_loaded"]
    return f[row] or not f["skill_loaded"]


def main():
    ap = Parser(add_help=False)
    ap.add_argument("--row", action="append", required=True)
    ap.add_argument("scores", nargs="+")
    a = ap.parse_args()
    rows = list(dict.fromkeys(a.row))
    arms = {}
    for p in a.scores:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ["arm", "trial", "skill_loaded"] + rows:
            if k not in kv:
                fail(f"{p} has no {k}")
        if not kv["arm"] or not kv["trial"]:
            fail(f"{p}: arm or trial is empty")
        flags = {}
        for k in ["skill_loaded"] + rows:
            if kv[k] not in ("0", "1"):
                fail(f"{p}: {k} is {kv[k]!r}, not 0 or 1")
            flags[k] = kv[k] == "1"
        t = arms.setdefault(kv["arm"], {})
        if kv["trial"] in t:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        t[kv["trial"]] = flags
    if "base" not in arms:
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = ["base"] + sorted(x for x in arms if x != "base")
    k = {}
    for r in rows:
        for arm in order:
            k[r, arm] = sum(1 for f in arms[arm].values() if fails(f, r, arm))
            print(f"{r} {arm}: failing {k[r, arm]} of {TRIALS}")
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
