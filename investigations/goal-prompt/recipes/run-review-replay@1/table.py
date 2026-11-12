#!/usr/bin/env python3
"""table SCORE...: one row per (arm, cut) from score outputs; measures only.

Columns: arm, cut, trials, costs (sorted, "none" for a trial with no result
record), median_cost (the middle of the sorted costs, a trial with no cost
counting highest; for an even count the higher of the two middles), ratio
(median_cost over the base arm's median at the same cut, "inf" when the
median is a missing cost), a1, a1_strict, b2, b2_strict (passing trials,
cut a for a1 and cut b for b2, "-" at the other cut), l4_fail and
l4_strict_fail (failing trials).

The base arm is the one named "base". It exits 2 when a score file lacks a
field, names a cut other than a or b, or a cut has trials but no base row.
"""
import sys

FIELDS = ["arm", "cut", "trial", "cost_usd", "a1", "a1_strict", "b2", "b2_strict", "l4", "l4_strict"]
INF = float("inf")


def die(msg):
    print("table: " + msg, file=sys.stderr)
    sys.exit(2)


def read(path):
    d = {}
    with open(path) as f:
        for line in f:
            if "\t" in line:
                k, v = line.rstrip("\n").split("\t", 1)
                d[k] = v
    for k in FIELDS:
        if k not in d:
            die("%s has no %s" % (path, k))
    if d["cut"] not in ("a", "b"):
        die("%s names cut %s" % (path, d["cut"]))
    return d


def median(costs):
    s = sorted(costs)
    return s[len(s) // 2]


def fmt(x):
    return "inf" if x == INF else "%.4f" % x


def main():
    if len(sys.argv) < 2:
        die("usage: table SCORE...")
    rows = [read(p) for p in sys.argv[1:]]
    groups = {}
    for r in rows:
        groups.setdefault((r["arm"], r["cut"]), []).append(r)
    med = {}
    for (arm, cut), rs in groups.items():
        med[(arm, cut)] = median([INF if r["cost_usd"] == "none" else float(r["cost_usd"]) for r in rs])
    for cut in ("a", "b"):
        if any(k[1] == cut for k in groups) and ("base", cut) not in groups:
            die("cut %s has no base trials" % cut)
    print("\t".join(["arm", "cut", "trials", "costs", "median_cost", "ratio", "a1", "a1_strict",
                     "b2", "b2_strict", "l4_fail", "l4_strict_fail"]))
    for arm, cut in sorted(groups, key=lambda k: (k[0] != "base", k[0], k[1])):
        rs = groups[(arm, cut)]
        costs = sorted(INF if r["cost_usd"] == "none" else float(r["cost_usd"]) for r in rs)
        m, b = med[(arm, cut)], med[("base", cut)]
        ratio = INF if m == INF else (0.0 if b == INF else m / b)
        cnt = lambda k: str(sum(int(r[k]) for r in rs))
        print("\t".join([
            arm, cut, str(len(rs)), ",".join("none" if c == INF else "%.4f" % c for c in costs), fmt(m),
            fmt(ratio) if ratio != INF else "inf",
            cnt("a1") if cut == "a" else "-", cnt("a1_strict") if cut == "a" else "-",
            cnt("b2") if cut == "b" else "-", cnt("b2_strict") if cut == "b" else "-",
            str(sum(1 - int(r["l4"]) for r in rs)), str(sum(1 - int(r["l4_strict"]) for r in rs)),
        ]))


if __name__ == "__main__":
    main()
