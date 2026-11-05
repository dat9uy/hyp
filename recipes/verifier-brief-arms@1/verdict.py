#!/usr/bin/env python3
"""verdict SCORE...: the skill-text gate over one score file per trial.

Void trials are dropped; the rest must give each arm exactly 5 distinct trials.
For an arm, k is its failed trials of 5, opened counts trials with opened 1,
and control_flagged counts trials answering wrong on the control.

  base k <= 2   h-base-enough=supports, h-judge=inconclusive,
                h-open-all=inconclusive, h-neither=contradicts. Only base may
                be given: the variants are not run (early stop).
  base k >= 3   h-base-enough=contradicts; judge and open-all must both be
                given; an arm with k <= 1 supports, else contradicts;
                h-neither supports when no arm supports, else contradicts.

Prints one line per arm, the four verdict lines and a ships line: the arm that
passes; if both, the one with fewer failures; if tied, judge; if none, none.
Exits 2 on a usage error, an unreadable file, a missing field, an unknown arm,
a repeated trial, an arm without 5 non-void trials, or a set of arms the rule
above does not accept.
"""
import sys

FIELDS = {"arm", "trial", "void", "opened", "answer_control", "failed"}
TRIALS = 5
ARMS = ("judge", "open-all")


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    if len(sys.argv) < 2:
        fail("usage: verdict SCORE...")
    arms = {}
    for p in sys.argv[1:]:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        if not FIELDS <= kv.keys():
            fail(f"{p} lacks " + ", ".join(sorted(FIELDS - kv.keys())))
        a = kv["arm"]
        if a != "base" and a not in ARMS:
            fail(f"{p} names arm {a!r}")
        if kv["void"] == "1":
            continue
        trials = arms.setdefault(a, {})
        if kv["trial"] in trials:
            fail(f"{p} repeats {a} trial {kv['trial']}")
        trials[kv["trial"]] = kv
    if "base" not in arms:
        fail("no base trial")
    for a, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {a} has {len(t)} non-void trials, want {TRIALS}")

    def k(a):
        return sum(1 for kv in arms[a].values() if kv["failed"] != "0")

    base_k = k("base")
    given = set(arms) - {"base"}
    if base_k <= 2 and given:
        fail(f"base k is {base_k}, so the variants are not run, but got: {' '.join(sorted(given))}")
    if base_k >= 3 and given != set(ARMS):
        fail(f"base k is {base_k}, so both variants are needed; got: {' '.join(sorted(given)) or 'base alone'}")

    for a in sorted(arms):
        opened = sum(1 for kv in arms[a].values() if kv["opened"] == "1")
        flagged = sum(1 for kv in arms[a].values() if kv["answer_control"] == "wrong")
        print(f"{a}: k={k(a)}/{TRIALS} opened={opened}/{TRIALS} control_flagged={flagged}/{TRIALS}")

    if base_k <= 2:
        words = {"h-base-enough": "supports", "h-judge": "inconclusive", "h-open-all": "inconclusive",
                 "h-neither": "contradicts"}
        ships = "none"
    else:
        passing = [a for a in ARMS if k(a) <= 1]
        words = {"h-base-enough": "contradicts",
                 "h-judge": "supports" if "judge" in passing else "contradicts",
                 "h-open-all": "supports" if "open-all" in passing else "contradicts",
                 "h-neither": "contradicts" if passing else "supports"}
        ships = min(passing, key=lambda a: (k(a), ARMS.index(a))) if passing else "none"
    for label, word in words.items():
        print(f"verdict: {label}={word}")
    print(f"ships: {ships}")


if __name__ == "__main__":
    main()
