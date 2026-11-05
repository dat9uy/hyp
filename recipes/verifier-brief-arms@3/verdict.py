#!/usr/bin/env python3
"""verdict SCORE...: the skill-text gate over one score file per trial.

Void trials are dropped; the rest must give each arm exactly 5 distinct trials.
For an arm, k is its failed trials of 5, opened counts trials with opened 1,
and control_flagged counts trials answering wrong on the control.

  base k <= 2   reproduced: no. h-null=supports, h-reader=inconclusive, ships:
                none. Only base may be given: the variant is not run.
  base k >= 3   reproduced: yes. h-null=contradicts, and reader must be given:
                h-reader=supports at reader k <= 1, else contradicts. ships:
                reader when it supports, else none.

Prints one line per arm, the reproduced line, the two verdict lines and a ships
line. Exits 2 on a usage error, an unreadable file, a missing field, an unknown
arm, a repeated trial, an arm without 5 non-void trials, or a set of arms the
rule above does not accept.
"""
import sys

FIELDS = {"arm", "trial", "void", "opened", "answer_control", "failed"}
TRIALS = 5
VARIANT = "reader"


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
        if a not in ("base", VARIANT):
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
    if base_k <= 2 and VARIANT in arms:
        fail(f"base k is {base_k}, so the variant is not run, but it was given")
    if base_k >= 3 and VARIANT not in arms:
        fail(f"base k is {base_k}, so the variant is needed")

    for a in sorted(arms):
        opened = sum(1 for kv in arms[a].values() if kv["opened"] == "1")
        flagged = sum(1 for kv in arms[a].values() if kv["answer_control"] == "wrong")
        print(f"{a}: k={k(a)}/{TRIALS} opened={opened}/{TRIALS} control_flagged={flagged}/{TRIALS}")

    if base_k <= 2:
        print("reproduced: no")
        print("verdict: h-null=supports")
        print("verdict: h-reader=inconclusive")
        print("ships: none")
        return
    passed = k(VARIANT) <= 1
    print("reproduced: yes")
    print("verdict: h-null=contradicts")
    print("verdict: h-reader=" + ("supports" if passed else "contradicts"))
    print("ships: " + (VARIANT if passed else "none"))


if __name__ == "__main__":
    main()
