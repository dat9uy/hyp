#!/usr/bin/env python3
"""verdict --row F [--row F]... SCORE...: the skill-text gate over every trial's score.

Each SCORE is a file of key<TAB>value lines for one trial, as score prints
them. Needs at least one --row (a fingerprint field, e.g. f_body_limit)
and exactly 5 trials per arm, with distinct trial values. The base arm is
the one named base.

A trial fails a row when its F is 1 (0 otherwise; any other value is an
error). skill_loaded decides what a trial that never loaded a skill (0)
counts as: in base it is never a failure, so it is never part of a
reproduction; in any other arm it is always a failure, on every row.

For each row and arm, k is the number of failing trials.
  - A row reproduces when base has k >= 3.
  - A variant arm passes a row when its k <= 1.

Prints "ROW ARM: failing K of 5" for each row and arm (rows in the given
order, base first), then:

  verdict: repro=supports       when any row given reproduces
  verdict: repro=contradicts    when none does
  verdict: ARM=supports         for each variant arm that passes every
                                reproducing row
  verdict: ARM=contradicts      when it fails one reproducing row
  verdict: ARM=inconclusive     when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing key (arm, trial,
skill_loaded or a row), a value of skill_loaded or a row that is not 0 or
1, no base arm, or an arm without exactly 5 distinct trials.
"""
import sys

TRIALS, REPRO, PASS = 5, 3, 1


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def flag(kv, p, key):
    if kv.get(key) not in ("0", "1"):
        fail(f"{p}: {key} is {kv.get(key)!r}, want 0 or 1")
    return kv[key] == "1"


def main():
    args, rows, scores = sys.argv[1:], [], []
    i = 0
    while i < len(args):
        if args[i] == "--row":
            if i + 1 >= len(args):
                fail("--row needs a field")
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
        for k in ("arm", "trial", "skill_loaded", *rows):
            if k not in kv:
                fail(f"{p} has no {k}")
        loaded = flag(kv, p, "skill_loaded")
        fails = {}
        for r in rows:
            f = flag(kv, p, r)
            fails[r] = f if loaded else kv["arm"] != "base"
        t = arms.setdefault(kv["arm"], {})
        if kv["trial"] in t:
            fail(f"arm {kv['arm']} has trial {kv['trial']} twice")
        t[kv["trial"]] = fails
    if "base" not in arms:
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = ["base"] + sorted(a for a in arms if a != "base")
    k = {(r, a): sum(t[r] for t in arms[a].values()) for r in rows for a in order}
    for r in rows:
        for a in order:
            print(f"{r} {a}: failing {k[r, a]} of {TRIALS}")
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
