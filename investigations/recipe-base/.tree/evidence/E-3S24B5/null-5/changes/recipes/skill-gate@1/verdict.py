#!/usr/bin/env python3
"""verdict.py SCOREDIR ROW...: the gate's verdict from every trial's score."""
import glob
import os
import re
import sys

ARM = re.compile(r"^[a-z][a-z0-9-]{0,31}$")
FIELDS = ("f_reading_pin", "f_body_limit")


def die(msg):
    print("verdict.py: " + msg, file=sys.stderr)
    sys.exit(2)


def read(path):
    kv = {}
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")
            if not line:
                continue
            if "\t" not in line:
                die("%s: not key<TAB>value: %r" % (path, line))
            k, v = line.split("\t", 1)
            kv[k] = v
    return kv


def main():
    if len(sys.argv) < 3:
        die("want SCOREDIR and at least one row")
    d, rows = sys.argv[1], []
    for r in sys.argv[2:]:
        f = r if r.startswith("f_") else "f_" + r
        if f not in FIELDS:
            die("unknown row %r (want one of %s)" % (r, ", ".join(x[2:] for x in FIELDS)))
        if f not in rows:
            rows.append(f)
    files = sorted(glob.glob(os.path.join(d, "*.tsv")))
    if not files:
        die("no *.tsv score files in %s" % d)
    arms = {}
    for p in files:
        kv = read(p)
        for k in ("arm", "trial", "skill_loaded") + FIELDS:
            if k not in kv:
                die("%s: no %s" % (p, k))
        arm = kv["arm"]
        if not ARM.match(arm) or arm == "repro":
            die("%s: bad arm %r" % (p, arm))
        if not kv["trial"].isdigit():
            die("%s: bad trial %r" % (p, kv["trial"]))
        for k in ("skill_loaded",) + FIELDS:
            if kv[k] not in ("0", "1"):
                die("%s: %s must be 0 or 1, got %r" % (p, k, kv[k]))
        arms.setdefault(arm, []).append(kv)
    if "base" not in arms:
        die("no trials of the arm base")
    for arm, trials in sorted(arms.items()):
        nums = [int(t["trial"]) for t in trials]
        if len(trials) != 5:
            die("arm %s has %d trials, want exactly 5" % (arm, len(trials)))
        if len(set(nums)) != 5:
            die("arm %s has repeated trial numbers" % arm)

    def k(arm, row):
        n = 0
        for t in arms[arm]:
            loaded = t["skill_loaded"] == "1"
            hit = t[row] == "1"
            if arm == "base":
                n += hit and loaded
            else:
                n += hit or not loaded
        return n

    base = {r: k("base", r) for r in rows}
    repro = [r for r in rows if base[r] >= 3]
    for r in rows:
        print("k: base %s=%d" % (r[2:], base[r]))
    variants = [a for a in sorted(arms) if a != "base"]
    ks = {}
    for a in variants:
        for r in rows:
            ks[a, r] = k(a, r)
            print("k: %s %s=%d" % (a, r[2:], ks[a, r]))
    print("verdict: repro=%s" % ("supports" if repro else "contradicts"))
    for a in variants:
        if not repro:
            w = "inconclusive"
        elif all(ks[a, r] <= 1 for r in repro):
            w = "supports"
        else:
            w = "contradicts"
        print("verdict: %s=%s" % (a, w))


main()
