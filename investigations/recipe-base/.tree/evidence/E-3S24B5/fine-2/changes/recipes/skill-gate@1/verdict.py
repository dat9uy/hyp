#!/usr/bin/env python3
"""verdict --row ROW... SCORE...: the skill-text gate over the scores of all trials.

Each SCORE is one trial's score output (key<TAB>value lines), with at least
arm, trial, skill_loaded and every ROW. A ROW is a fingerprint key such as
f_body_limit, 0 or 1 in every score; --row repeats for the rows to judge.
The rule is fixed: 5 trials per arm, a row reproduces at k >= 3, a variant
arm passes a row at k <= 1.

For each row and arm, k is the number of the arm's trials whose ROW is 1,
where a trial with skill_loaded 0 counts as follows: in the arm "base"
(the arm with no change) never (its ROW is read as 0); in any other arm
always (its ROW is read as 1).

  verdict: repro=supports        when base has k >= 3 on any row given
  verdict: repro=contradicts     when it has none
  verdict: ARM=supports          for every non-base arm: when it has k <= 1
                                 on every row base reproduces
  verdict: ARM=contradicts       ... when it has k >= 2 on one of them
  verdict: ARM=inconclusive      ... when no row reproduces

Before the verdict lines it prints one `ARM ROW: k of 5` line per arm
(base first, then sorted) and row (in the order given). Exits 2 on a usage
error, an unreadable file, a missing key, a ROW or skill_loaded that is not 0
or 1, an arm name that is not [a-z][a-z0-9-]{0,31}, no arm "base", two
scores with the same arm and trial, or an arm without exactly 5 trials.
"""
import re
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
            if i + 1 >= len(argv):
                fail("--row needs a value")
            rows.append(argv[i + 1])
            i += 2
        elif argv[i].startswith("--"):
            fail(f"unknown flag {argv[i]}")
        else:
            files.append(argv[i])
            i += 1
    if not rows or not files:
        fail("usage: verdict --row ROW... SCORE...")
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
        arm = kv["arm"]
        if not re.fullmatch(r"[a-z][a-z0-9-]{0,31}", arm):
            fail(f"{p}: arm {arm!r} is not a label")
        for k in ["skill_loaded"] + rows:
            if kv[k] not in ("0", "1"):
                fail(f"{p}: {k} is {kv[k]!r}, not 0 or 1")
        if kv["trial"] in arms.setdefault(arm, {}):
            fail(f"arm {arm} has trial {kv['trial']} twice")
        arms[arm][kv["trial"]] = kv
    if "base" not in arms:
        fail("no trial of the base arm")
    for arm, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {arm} has {len(t)} trials, want {TRIALS}")

    def k(arm, row):
        n = 0
        for kv in arms[arm].values():
            if kv["skill_loaded"] == "0":
                n += 0 if arm == "base" else 1
            else:
                n += int(kv[row])
        return n

    order = ["base"] + sorted(a for a in arms if a != "base")
    ks = {(a, r): k(a, r) for a in order for r in rows}
    for a in order:
        for r in rows:
            print(f"{a} {r}: {ks[a, r]} of {TRIALS}")
    reproducing = [r for r in rows if ks["base", r] >= REPRO]
    print(f"verdict: repro={'supports' if reproducing else 'contradicts'}")
    for a in order[1:]:
        if not reproducing:
            word = "inconclusive"
        else:
            word = "supports" if all(ks[a, r] <= PASS for r in reproducing) else "contradicts"
        print(f"verdict: {a}={word}")


if __name__ == "__main__":
    main()
