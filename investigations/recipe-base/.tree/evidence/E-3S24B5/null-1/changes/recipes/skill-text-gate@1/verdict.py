#!/usr/bin/env python3
"""The gate: verdict --row ROW... SCORE_FILE...  (see recipe.md)."""
import re, sys

TRIALS = 5
REPRO_AT = 3  # base: k >= 3 reproduces
PASS_AT = 1   # variant: k <= 1 passes


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def read_score(path):
    s = {}
    try:
        with open(path, encoding="utf-8") as f:
            for line in f:
                line = line.rstrip("\n")
                if not line:
                    continue
                if "\t" not in line:
                    fail("%s: bad line %r" % (path, line))
                k, v = line.split("\t", 1)
                if k in s:
                    fail("%s: %s twice" % (path, k))
                s[k] = v
    except OSError as e:
        fail("cannot read %s: %s" % (path, e))
    return s


def main():
    args = sys.argv[1:]
    rows, files = [], []
    i = 0
    while i < len(args):
        if args[i] == "--row":
            if i + 1 >= len(args):
                fail("--row wants a name")
            rows.append(args[i + 1])
            i += 2
        else:
            files.append(args[i])
            i += 1
    if not rows:
        fail("no --row given")
    if len(set(rows)) != len(rows):
        fail("a row given twice")
    for r in rows:
        if not re.fullmatch(r"f_[a-z0-9_]+", r):
            fail("row %r is not a fingerprint (f_...)" % r)
    if not files:
        fail("no score file given")

    arms = {}  # arm -> {trial: score}
    for path in files:
        s = read_score(path)
        for need in ["arm", "trial", "skill_loaded"] + rows:
            if need not in s:
                fail("%s: no %s" % (path, need))
        arm = s["arm"]
        if not re.fullmatch(r"[a-z][a-z0-9-]{0,31}", arm) or arm == "repro":
            fail("%s: arm %r is not a usable label" % (path, arm))
        if not re.fullmatch(r"[1-9][0-9]*", s["trial"]):
            fail("%s: bad trial %r" % (path, s["trial"]))
        for key in ["skill_loaded"] + rows:
            if s[key] not in ("0", "1"):
                fail("%s: %s is %r, want 0 or 1" % (path, key, s[key]))
        trial = int(s["trial"])
        if trial in arms.setdefault(arm, {}):
            fail("arm %s: trial %d twice" % (arm, trial))
        arms[arm][trial] = s
    if "base" not in arms:
        fail("no base arm")
    for arm in sorted(arms):
        if len(arms[arm]) != TRIALS:
            fail("arm %s has %d trials, want %d" % (arm, len(arms[arm]), TRIALS))

    def k(arm, row):
        n = 0
        for s in arms[arm].values():
            loaded = s["skill_loaded"] == "1"
            if arm == "base":
                n += 1 if (loaded and s[row] == "1") else 0
            else:
                n += 1 if (not loaded or s[row] == "1") else 0
        return n

    variants = [a for a in sorted(arms) if a != "base"]
    ks = {(r, a): k(a, r) for r in rows for a in ["base"] + variants}
    reproducing = [r for r in rows if ks[(r, "base")] >= REPRO_AT]
    for r in rows:
        for a in ["base"] + variants:
            print("k: %s %s %d" % (r, a, ks[(r, a)]))
    print("verdict: repro=%s" % ("supports" if reproducing else "contradicts"))
    for a in variants:
        if not reproducing:
            word = "inconclusive"
        elif all(ks[(r, a)] <= PASS_AT for r in reproducing):
            word = "supports"
        else:
            word = "contradicts"
        print("verdict: %s=%s" % (a, word))


main()
