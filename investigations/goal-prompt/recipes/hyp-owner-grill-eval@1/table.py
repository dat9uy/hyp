#!/usr/bin/env python3
"""table REVIEWS MAP SCORE...: set the scored trials side by side; no verdict.

SCORE is the record directory of one score run (its stdout and out/).
MAP is a JSON object from each blind label a reviewer saw to the run id
in a record.json (<arm>-<fixture>-<trial>). REVIEWS is a JSON object from
each label to {assertion id: "pass" or "fail"}, one entry per rubric
assertion of that trial's case.

Per trial: a check passes when eval_skill.py passed it (json-equals) or
the review says pass (rubric). Lines 1 to 6 need has-goal; line 1 is
l1-store and l1-init, line 2 l2 and line 3 l3 (fixture a only), line 4
l4-hash and l4-grader, line 5 l5, line 6 l6 (fixture b only).

Prints a header and one row per arm and fixture, in sorted order: trials,
the passing count of each line ("-" where the fixture has none), the sorted
costs, median_cost (the middle of the sorted costs, a trial with no cost
counting highest) and its ratio over the base arm's median at the same
fixture, and the same two for cost_usd_literal. Exits 2 when a score
lacks a field, a run id is missing from MAP or repeated, a review lacks a
rubric assertion, or a fixture has trials and no base row.
"""
import json
import os
import sys

LINES = {1: ["l1-store", "l1-init"], 2: ["l2"], 3: ["l3"], 4: ["l4-hash", "l4-grader"], 5: ["l5"], 6: ["l6"]}
ONLY = {2: "a", 3: "a", 6: "b"}


def die(msg):
    print("table: " + msg, file=sys.stderr)
    sys.exit(2)


def kv(path):
    out = {}
    for line in open(path):
        if "\t" in line:
            k, v = line.rstrip("\n").split("\t", 1)
            out[k] = v
    return out


def cost(v):
    return None if v == "none" else float(v)


def median(xs):
    s = sorted(xs, key=lambda x: (x is None, x or 0))
    return s[len(s) // 2] if s else None


def fmt(x):
    return "none" if x is None else ("%.4f" % x)


def main():
    if len(sys.argv) < 4:
        die("usage: table REVIEWS MAP SCORE...")
    reviews, labels = json.load(open(sys.argv[1])), json.load(open(sys.argv[2]))
    label_of = {}
    for lab, rid in labels.items():
        if rid in label_of:
            die("run id %s has two labels" % rid)
        label_of[rid] = lab
    rows, seen = {}, set()
    for d in sys.argv[3:]:
        try:
            m, rec = kv(os.path.join(d, "stdout")), json.load(open(os.path.join(d, "out", "record.json")))
        except (OSError, ValueError) as e:
            die("%s: %s" % (d, e))
        for k in ("arm", "fixture", "trial", "cost_usd", "cost_usd_literal"):
            if k not in m:
                die("%s: score has no %s" % (d, k))
        rid = rec.get("run_id")
        if rid in seen:
            die("run id %s is scored twice" % rid)
        seen.add(rid)
        if rid not in label_of:
            die("run id %s has no label in MAP" % rid)
        rev = reviews.get(label_of[rid]) or {}
        passed = {}
        for c in rec.get("checks") or []:
            if c.get("passed") is None:
                v = rev.get(c["id"])
                if v not in ("pass", "fail"):
                    die("review %s has no pass or fail for %s" % (label_of[rid], c["id"]))
                passed[c["id"]] = v == "pass"
            else:
                passed[c["id"]] = bool(c["passed"])
        fx = m["fixture"]
        line = {}
        for n, ids in LINES.items():
            if ONLY.get(n, fx) != fx:
                continue
            line[n] = passed.get("has-goal", False) and all(passed.get(i, False) for i in ids)
        rows.setdefault((m["arm"], fx), []).append((line, cost(m["cost_usd"]), cost(m["cost_usd_literal"])))
    base = {}
    for (arm, fx), trials in rows.items():
        if arm == "base":
            base[fx] = (median([t[1] for t in trials]), median([t[2] for t in trials]))
    for _, fx in rows:
        if fx not in base:
            die("fixture %s has trials and no base row" % fx)
    print("arm\tfixture\ttrials\tl1\tl2\tl3\tl4\tl5\tl6\tcosts\tmedian_cost\tratio\tmedian_cost_literal\tratio_literal")
    for (arm, fx) in sorted(rows):
        trials = rows[(arm, fx)]
        counts = []
        for n in range(1, 7):
            if ONLY.get(n, fx) != fx:
                counts.append("-")
            else:
                counts.append("%d/%d" % (sum(t[0][n] for t in trials), len(trials)))
        med, medl = median([t[1] for t in trials]), median([t[2] for t in trials])
        bm, bml = base[fx]
        ratio = "none" if med is None or not bm else "%.3f" % (med / bm)
        ratiol = "none" if medl is None or not bml else "%.3f" % (medl / bml)
        costs = ",".join(fmt(c) for c in sorted((t[1] for t in trials), key=lambda x: (x is None, x or 0)))
        print("\t".join([arm, fx, str(len(trials))] + counts + [costs, fmt(med), ratio, fmt(medl), ratiol]))


if __name__ == "__main__":
    main()
