#!/usr/bin/env python3
"""verdict --row F [--row G ...] SCORE...: the skill-text gate.

Each SCORE is a file of key<TAB>value lines for one trial (score's output),
with at least arm, trial, skill_loaded and every row F, whose value is 0 or 1.
Exactly 5 trials per arm, with distinct trial numbers; one arm is named base.
An arm name is a label, [a-z][a-z0-9-]{0,31}, and is not "repro".

For each row and arm, k is the number of trials whose F is 1, where a trial
with skill_loaded 0 counts against its arm: in base it is never a
reproduction (it adds nothing to k), in any other arm it always counts as a
failure (it adds 1 to k, whatever F is).

  A row reproduces when base has k >= 3.
  A variant arm passes a row when its k <= 1.

Prints one "ROW ARM: k of 5" line per row and arm, then
  verdict: repro=supports      when any row given reproduces, else contradicts
and, for each other arm in name order,
  verdict: ARM=supports        when it passes every reproducing row
  verdict: ARM=contradicts     when it fails a reproducing row
  verdict: ARM=inconclusive    when no row reproduces

Exits 2 on a usage error, an unreadable file, a missing key, a value that is
not as above, an arm without exactly 5 distinct trials, or no base arm.
"""
import sys

TRIALS, REPRO, PASS = 5, 3, 1


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    args, rows, files, i = sys.argv[1:], [], [], 0
    while i < len(args):
        if args[i] == "--row":
            if i + 1 >= len(args):
                fail("--row needs a name")
            rows.append(args[i + 1])
            i += 2
        elif args[i].startswith("--"):
            fail(f"unknown flag {args[i]}")
        else:
            files.append(args[i])
            i += 1
    if not rows or not files:
        fail("usage: verdict --row F [--row G ...] SCORE...")
    if len(set(rows)) != len(rows):
        fail("a row is given twice")
    import re
    arms = {}
    for p in files:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for k in ("arm", "trial", "skill_loaded", *rows):
            if k not in kv:
                fail(f"{p} has no {k}")
        for k in ("skill_loaded", *rows):
            if kv[k] not in ("0", "1"):
                fail(f"{p}: {k} is {kv[k]!r}, not 0 or 1")
        arm = kv["arm"]
        if not re.fullmatch(r"[a-z][a-z0-9-]{0,31}", arm) or arm == "repro":
            fail(f"{p}: arm {arm!r} is not a usable label")
        t = arms.setdefault(arm, {})
        if kv["trial"] in t:
            fail(f"arm {arm} has trial {kv['trial']} twice")
        t[kv["trial"]] = kv
    if "base" not in arms:
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")
    order = ["base"] + sorted(a for a in arms if a != "base")

    def k(arm, row):
        n = 0
        for kv in arms[arm].values():
            if kv["skill_loaded"] == "0":
                n += 0 if arm == "base" else 1
            else:
                n += int(kv[row])
        return n

    reproducing = []
    for row in rows:
        for arm in order:
            print(f"{row} {arm}: {k(arm, row)} of {TRIALS}")
        if k("base", row) >= REPRO:
            reproducing.append(row)
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for arm in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(k(arm, r) <= PASS for r in reproducing) else "contradicts"
        print(f"verdict: {arm}={word}")


if __name__ == "__main__":
    main()
