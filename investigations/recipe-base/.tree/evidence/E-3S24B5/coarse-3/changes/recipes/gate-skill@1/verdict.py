#!/usr/bin/env python3
"""verdict --row F [--row F]... SCORE...: the skill-text gate over score outputs.

Each SCORE is a file of key<TAB>value lines for one trial (score's output),
with at least arm, trial, skill_loaded and every row F. A row is a
fingerprint field such as f_reading_pin; its value is 0 or 1. The rows to
judge are the --row arguments (at least one). The rule is the method's and
is fixed here: the base arm is `base`, every arm needs exactly 5 trials with
distinct trial numbers, a row reproduces at k >= 3 and a variant arm passes
a row at k <= 1.

For a row and an arm, k is the number of trials that count as a failure:
  base arm      a trial whose fingerprint is 1 and whose skill_loaded is 1
                (a trial that never loaded a skill is never a reproduction)
  variant arm   a trial whose fingerprint is 1 or whose skill_loaded is 0
                (a trial that never loaded a skill always counts as failure)

Prints, per row and arm, "ROW ARM: k of 5" (rows in argument order, base
first, then the other arms sorted), then:

  verdict: repro=supports        when any row given reproduces
  verdict: repro=contradicts     when none does
  verdict: ARM=supports          for each variant arm (sorted), when some row
                                 reproduces and the arm passes every
                                 reproducing row
  verdict: ARM=contradicts       when some row reproduces and the arm fails one
  verdict: ARM=inconclusive      when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing field, a value that
is not 0 or 1, no base arm, or an arm without exactly 5 distinct trials.
"""
import argparse
import sys

BASE, TRIALS, REPRO, PASS = "base", 5, 3, 1


class Parser(argparse.ArgumentParser):
    def error(self, message):
        fail(message)


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def flag(kv, p, key):
    if key not in kv:
        fail(f"{p} has no {key}")
    if kv[key] not in ("0", "1"):
        fail(f"{p}: {key} is {kv[key]!r}, not 0 or 1")
    return kv[key] == "1"


def main():
    ap = Parser(add_help=False)
    ap.add_argument("--row", action="append", required=True)
    ap.add_argument("scores", nargs="+")
    a = ap.parse_args()
    arms = {}
    for p in a.scores:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ("arm", "trial"):
            if k not in kv:
                fail(f"{p} has no {k}")
        loaded = flag(kv, p, "skill_loaded")
        hits = {r: flag(kv, p, r) for r in a.row}
        trials = arms.setdefault(kv["arm"], {})
        if kv["trial"] in trials:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        trials[kv["trial"]] = (loaded, hits)
    if BASE not in arms:
        fail(f"no trial of the base arm {BASE}")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = [BASE] + sorted(x for x in arms if x != BASE)
    k = {}
    for r in a.row:
        for arm in order:
            ts = arms[arm].values()
            if arm == BASE:
                k[r, arm] = sum(1 for loaded, h in ts if loaded and h[r])
            else:
                k[r, arm] = sum(1 for loaded, h in ts if h[r] or not loaded)
            print(f"{r} {arm}: {k[r, arm]} of {TRIALS}")
    reproducing = [r for r in a.row if k[r, BASE] >= REPRO]
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for arm in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(k[r, arm] <= PASS for r in reproducing) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
