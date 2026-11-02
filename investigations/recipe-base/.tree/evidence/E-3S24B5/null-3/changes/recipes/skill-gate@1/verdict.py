#!/usr/bin/env -S python3 -I
"""verdict.py ROW... SCORE_FILE...: the skill-text gate's verdict.

ROW is a fingerprint name (f_...); SCORE_FILE is the absolute path of one
trial's score (the output of score.py). Needs exactly 5 trials per arm with
distinct trial numbers. See recipe.md, The gate.
"""
import re
import sys

ROW = re.compile(r"f_[a-z0-9_]+")
LABEL = re.compile(r"[a-z][a-z0-9-]{0,31}")
TRIALS = 5
REPRODUCES = 3   # base: k >= 3
PASSES = 1       # variant arm: k <= 1


def die(msg):
    sys.stderr.write("verdict.py: %s\n" % msg)
    sys.exit(2)


def read_score(path):
    try:
        with open(path, encoding="utf-8") as f:
            lines = f.read().split("\n")
    except OSError as e:
        die("cannot read %s: %s" % (path, e))
    kv = {}
    for line in lines:
        if not line:
            continue
        if line.count("\t") != 1:
            die("%s: not a key<TAB>value line: %r" % (path, line))
        k, v = line.split("\t")
        if k in kv:
            die("%s: %s twice" % (path, k))
        kv[k] = v
    return kv


def flag(path, kv, key):
    if kv.get(key) not in ("0", "1"):
        die("%s: %s is %r, want 0 or 1" % (path, key, kv.get(key)))
    return kv[key] == "1"


def main():
    rows = [a for a in sys.argv[1:] if ROW.fullmatch(a)]
    files = [a for a in sys.argv[1:] if not ROW.fullmatch(a)]
    if not rows:
        die("give at least one row (f_...) to judge")
    if len(set(rows)) != len(rows):
        die("a row is given twice")
    if not files:
        die("give the score files")

    arms = {}   # arm -> {trial -> (skill_loaded, {row: fingerprint})}
    for path in files:
        kv = read_score(path)
        arm = kv.get("arm", "")
        if not LABEL.fullmatch(arm) or arm == "repro":
            die("%s: arm %r is not a usable arm name" % (path, arm))
        if not re.fullmatch(r"[1-9][0-9]*", kv.get("trial", "")):
            die("%s: trial %r is not a positive whole number" % (path, kv.get("trial")))
        trial = int(kv["trial"])
        if trial in arms.setdefault(arm, {}):
            die("arm %s: trial %d twice" % (arm, trial))
        loaded = flag(path, kv, "skill_loaded")
        arms[arm][trial] = (loaded, {r: flag(path, kv, r) for r in rows})

    if "base" not in arms:
        die("no base arm")
    for arm in sorted(arms):
        if len(arms[arm]) != TRIALS:
            die("arm %s has %d trials, want exactly %d" % (arm, len(arms[arm]), TRIALS))

    def k(arm, row):
        n = 0
        for loaded, fp in arms[arm].values():
            if arm == "base":
                n += 1 if (loaded and fp[row]) else 0   # no skill: never a reproduction
            else:
                n += 1 if (not loaded or fp[row]) else 0  # no skill: always a failure
        return n

    reproducing = [r for r in rows if k("base", r) >= REPRODUCES]
    for r in rows:
        for arm in sorted(arms):
            print("k %s %s %d" % (r, arm, k(arm, r)))
    print("verdict: repro=%s" % ("supports" if reproducing else "contradicts"))
    for arm in sorted(arms):
        if arm == "base":
            continue
        if not reproducing:
            word = "inconclusive"
        elif all(k(arm, r) <= PASSES for r in reproducing):
            word = "supports"
        else:
            word = "contradicts"
        print("verdict: %s=%s" % (arm, word))


main()
