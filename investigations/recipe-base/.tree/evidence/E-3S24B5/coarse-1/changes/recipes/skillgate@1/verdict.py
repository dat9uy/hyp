#!/usr/bin/env python3
"""verdict --base ARM --trials N --repro R --pass P --row F [--row F]... SCORE...: the skill-text gate.

Each SCORE is a file of key<TAB>value lines for one trial, with at least arm,
trial, skill_loaded and every row F. Every arm must have exactly N trials with
distinct trial values. Each --row names a fingerprint field to judge.

For each row and arm, k is the number of failing trials. A trial's F is a
failure when it is not 0. A trial whose skill_loaded is 0 (the runner never
loaded a skill) counts against its arm: in the base arm it is never a
failure (never a reproduction), in any other arm it always is.

  A row reproduces when the base arm's k >= R.
  A non-base arm passes a row when its k <= P.

  verdict: repro=supports        when any row reproduces
  verdict: repro=contradicts     when none does
  verdict: ARM=supports          when ARM passes every reproducing row
  verdict: ARM=contradicts       when it fails one of them
  verdict: ARM=inconclusive      when no row reproduces

Prints, first, one line per row and arm ("ARM: ROW failing K of N", base
first, then arms sorted, rows in argument order), then the verdict lines
(arms sorted). Exits 2 on a usage error, an unreadable file, a missing key, a
value that is not a number, no base arm, or an arm without exactly N
distinct trials. The thresholds and rows are arguments, so a pre-registered
rule is the command line that runs this script.
"""
import argparse
import sys


class Parser(argparse.ArgumentParser):
    def error(self, message):
        fail(message)


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def number(p, kv, key):
    try:
        return float(kv[key])
    except ValueError:
        fail(f"{p}: {key} is {kv[key]!r}, not a number")


def main():
    ap = Parser(add_help=False)
    ap.add_argument("--base", required=True)
    for flag in ("--trials", "--repro", "--pass"):
        ap.add_argument(flag, required=True, type=int)
    ap.add_argument("--row", action="append", required=True)
    ap.add_argument("scores", nargs="+")
    a = ap.parse_args()
    arms = {}
    for p in a.scores:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ["arm", "trial", "skill_loaded"] + a.row:
            if k not in kv:
                fail(f"{p} has no {k}")
        loaded = number(p, kv, "skill_loaded") != 0
        failed = {r: number(p, kv, r) != 0 for r in a.row}
        if not loaded:
            failed = {r: kv["arm"] != a.base for r in a.row}
        arms.setdefault(kv["arm"], {})
        if kv["trial"] in arms[kv["arm"]]:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        arms[kv["arm"]][kv["trial"]] = failed
    if a.base not in arms:
        fail(f"no trial of the base arm {a.base}")
    for arm, t in arms.items():
        if len(t) != a.trials:
            fail(f"arm {arm} has {len(t)} trials, want {a.trials}")
    k = {arm: {r: sum(x[r] for x in t.values()) for r in a.row} for arm, t in arms.items()}
    others = sorted(x for x in arms if x != a.base)
    for arm in [a.base] + others:
        for r in a.row:
            print(f"{arm}: {r} failing {k[arm][r]} of {a.trials}")
    repro = [r for r in a.row if k[a.base][r] >= a.repro]
    print(f"verdict: repro={'supports' if repro else 'contradicts'}")
    for arm in others:
        if not repro:
            word = "inconclusive"
        else:
            word = "supports" if all(k[arm][r] <= getattr(a, "pass") for r in repro) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
