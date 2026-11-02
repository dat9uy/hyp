#!/usr/bin/env python3
"""verdict --base ARM --trials N --repro R --pass P --row F... SCORE...: the skill-text gate.

Each SCORE is a file of key<TAB>value lines for one trial, with at least
arm, trial, skill_loaded and every F named by --row (repeatable; the rows
to judge). Every arm must have exactly N trials with distinct trial values.

For a row and an arm, k is the number of trials whose F is 1 (a value
other than 0 or 1 is an error). A trial whose skill_loaded is 0 counts
against its arm: in the base arm it is never a reproduction (it adds
nothing to k), in any other arm it always counts as a failure (it adds one
to k, whatever F says).

A row reproduces when the base arm's k >= R. A variant arm passes a row
when its k <= P.

  verdict: repro=supports        when any row given reproduces
  verdict: repro=contradicts     when none does
  verdict: ARM=supports          for a variant arm that passes every reproducing row
  verdict: ARM=contradicts       for a variant arm that fails one
  verdict: ARM=inconclusive      for every variant arm, when no row reproduces

Prints one k line per row and arm (rows in the order given, base first),
then the verdict lines. Exits 2 on a usage error, an unreadable file, a
missing field, a non-0/1 value, no base arm, or an arm without exactly N
distinct trials.
"""
import argparse
import sys


class Parser(argparse.ArgumentParser):
    def error(self, message):
        fail(message)


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def flag(kv, p, name):
    if name not in kv:
        fail(f"{p} has no {name}")
    if kv[name] not in ("0", "1"):
        fail(f"{p}: {name} is {kv[name]!r}, not 0 or 1")
    return kv[name] == "1"


def main():
    ap = Parser(add_help=False)
    ap.add_argument("--base", required=True)
    ap.add_argument("--row", action="append", required=True)
    for f in ("--trials", "--repro", "--pass"):
        ap.add_argument(f, required=True, type=int)
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
        t = {}
        for row in a.row:
            hit = flag(kv, p, row)
            # an unloaded trial: never a reproduction in base, always a failure elsewhere
            t[row] = hit and loaded if kv["arm"] == a.base else hit or not loaded
        arms.setdefault(kv["arm"], {})
        if kv["trial"] in arms[kv["arm"]]:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        arms[kv["arm"]][kv["trial"]] = t
    if a.base not in arms:
        fail(f"no trial of the base arm {a.base}")
    for arm, t in arms.items():
        if len(t) != a.trials:
            fail(f"arm {arm} has {len(t)} trials, want {a.trials}")
    order = [a.base] + sorted(x for x in arms if x != a.base)
    k = {(arm, r): sum(t[r] for t in arms[arm].values()) for arm in order for r in a.row}
    for r in a.row:
        for arm in order:
            print(f"{arm}: {r} failing {k[arm, r]} of {a.trials}")
    reproducing = [r for r in a.row if k[a.base, r] >= a.repro]
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for arm in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(k[arm, r] <= getattr(a, "pass") for r in reproducing) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
