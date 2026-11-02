#!/usr/bin/env python3
"""verdict --row F [--row F]... SCORE...: the skill-text gate.

Each SCORE is a file of key<TAB>value lines for one trial: arm, trial,
skill_loaded and every row F. The arm named base is the reference; every
other arm is a variant. Every arm must have exactly 5 trials with distinct
trial numbers, and the base arm must exist.

For a row F and an arm, k is the number of trials whose F is 1. A trial
whose skill_loaded is 0 counts against its arm: in base it is never a
reproduction (it adds nothing to k), in a variant arm it always adds 1.

  A row reproduces when base has k >= 3.
  A variant passes a row at k <= 1.

Prints one line `ARM: ROW failing K of 5` per arm (base first, then sorted)
and row (in the order given), then:

  verdict: repro=supports        when any row given reproduces
  verdict: repro=contradicts     otherwise
  verdict: ARM=supports          a variant that passes every reproducing row
  verdict: ARM=contradicts       a variant that fails one reproducing row
  verdict: ARM=inconclusive      every variant, when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing key, a value that is
not 0 or 1, no base arm, or an arm without exactly 5 distinct trials.
"""
import sys

TRIALS, REPRO, PASS = 5, 3, 1


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    argv = sys.argv[1:]
    rows, files = [], []
    i = 0
    while i < len(argv):
        if argv[i] == "--row":
            if i + 1 >= len(argv) or argv[i + 1].startswith("-"):
                fail("--row needs a field name")
            rows.append(argv[i + 1])
            i += 2
        else:
            files.append(argv[i])
            i += 1
    if not rows or not files:
        fail("usage: verdict --row F [--row F]... SCORE...")
    if len(set(rows)) != len(rows):
        fail("a row is given twice")
    arms = {}
    for p in files:
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
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")

    def k(arm, row):
        n = 0
        for kv in arms[arm].values():
            loaded = kv["skill_loaded"] == "1"
            if arm == "base":
                n += 1 if loaded and kv[row] == "1" else 0
            else:
                n += 1 if (not loaded) or kv[row] == "1" else 0
        return n

    order = ["base"] + sorted(a for a in arms if a != "base")
    for arm in order:
        for row in rows:
            print(f"{arm}: {row} failing {k(arm, row)} of {TRIALS}")
    reproducing = [r for r in rows if k("base", r) >= REPRO]
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for arm in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(k(arm, r) <= PASS for r in reproducing) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
