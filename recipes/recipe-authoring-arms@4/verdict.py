#!/usr/bin/env python3
"""verdict SCORE...: e-author's pre-registered rule over score outputs.

Each SCORE is a file holding one trial's score lines (arm, stops,
end_state). Arms are main and scaffold. With m and s the mean stops of
main and scaffold, and pm and ps their counts of end_state pass:

  null region: m <= 1 and pm >= 2/3 of main's trials
      h-binary=contradicts h-partial=supports h-null=supports
  bar (and not the null region): s <= m / 2 and ps >= pm
      h-binary=supports h-partial=contradicts h-null=contradicts
  otherwise: every label inconclusive

Prints one summary line per arm, then the verdict lines. Exits 2 when a
file is unreadable, lacks a field, or names another arm, or when an arm has
no trial.
"""
import sys


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    if len(sys.argv) < 3:
        fail("usage: verdict SCORE...")
    arms = {"main": [], "scaffold": []}
    for p in sys.argv[1:]:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        if not {"arm", "stops", "end_state"} <= kv.keys():
            fail(f"{p} lacks arm, stops or end_state")
        if kv["arm"] not in arms:
            fail(f"{p} names arm {kv['arm']}")
        arms[kv["arm"]].append((int(kv["stops"]), kv["end_state"] == "pass"))
    for a, t in arms.items():
        if not t:
            fail(f"no {a} trial")
    stat = {}
    for a, t in arms.items():
        mean = sum(s for s, _ in t) / len(t)
        passes = sum(1 for _, ok in t if ok)
        stat[a] = (mean, passes, len(t))
        print(f"{a}: trials {len(t)}, mean stops {mean:.2f}, end_state pass {passes}")
    m, pm, nm = stat["main"]
    s, ps, _ = stat["scaffold"]
    if m <= 1 and 3 * pm >= 2 * nm:
        v = ("contradicts", "supports", "supports")
    elif s <= m / 2 and ps >= pm:
        v = ("supports", "contradicts", "contradicts")
    else:
        v = ("inconclusive",) * 3
    for label, word in zip(("h-binary", "h-partial", "h-null"), v):
        print(f"verdict: {label}={word}")


if __name__ == "__main__":
    main()
