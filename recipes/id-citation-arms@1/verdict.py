#!/usr/bin/env python3
"""Apply e-arms' pre-registered verdict rule to a `cases` table.

usage: verdict.py TABLE CASES

TABLE is the stdout of `cases` (arm, case, exit, appended, warned); CASES is
the cases file it ran, read for each case's class: slip (K1-K6), prose-slip
(K7, never scored), guard (G1-G5), cite-slip and cite-guard (C1). The arms
are today (h-null), text-warn (h-text-warn), text-refuse (h-text-refuse)
and cite (h-cite); TABLE must hold a row for each arm and case.

- h-text-warn supports when its arm, on every slip, exits 0, appends and
  warns, and on every guard exits 0, appends and does not warn.
- h-text-refuse supports when its arm, on every slip, exits non-zero and
  appends nothing, and on every guard exits 0, appends and does not warn.
- h-cite supports when its arm, on every cite-slip, exits non-zero and
  appends nothing, on every cite-guard exits 0 and appends, and on every
  guard gives today's row (exit, appended, warned).
- h-null contradicts when the today arm, on any slip, exits 0, appends and
  does not warn; otherwise it supports.
- An arm with a scored row that exited 3 (cases already re-ran it once) is
  inconclusive, and so is h-cite when today is.

Prints one summary line per arm, then `verdict: LABEL=WORD` per rival.
Exit 0, or 2 for bad arguments or a malformed or incomplete table.
"""
import sys

ARMS = [("text-warn", "h-text-warn"), ("text-refuse", "h-text-refuse"),
        ("cite", "h-cite"), ("today", "h-null")]


def die(msg):
    print(f"verdict.py: {msg}", file=sys.stderr)
    sys.exit(2)


def read_tsv(path, header):
    try:
        with open(path) as f:
            lines = f.read().splitlines()
    except OSError as e:
        die(str(e))
    if not lines or lines[0].split("\t") != header:
        die(f"{path}: want header {', '.join(header)}")
    rows = [l.split("\t") for l in lines[1:]]
    for r in rows:
        if len(r) != len(header):
            die(f"{path}: bad row {r!r}")
    return rows


def main(argv):
    if len(argv) != 2:
        die("usage: verdict.py TABLE CASES")
    table = read_tsv(argv[0], ["arm", "case", "exit", "appended", "warned"])
    classes = {r[0]: r[1] for r in read_tsv(argv[1], ["case", "class", "steps"])}
    rows = {}
    for arm, case, code, app, warn in table:
        if not code.lstrip("-").isdigit() or app not in "yn" or warn not in "yn" or len(app) != 1 or len(warn) != 1:
            die(f"bad row {arm} {case}")
        if case not in classes:
            die(f"case {case} is not in {argv[1]}")
        rows[(arm, case)] = (int(code), app, warn)
    for arm, _ in ARMS:
        for case in classes:
            if (arm, case) not in rows:
                die(f"no row for {arm} {case}")

    def of(cls):
        return [c for c, k in classes.items() if k == cls]

    slips, guards = of("slip"), of("guard")
    cslips, cguards = of("cite-slip"), of("cite-guard")
    clean_guard = lambda r: r == (0, "y", "n")
    refused = lambda r: r[0] != 0 and r[1] == "n"
    env = lambda arm, cases: any(rows[(arm, c)][0] == 3 for c in cases)

    def word(ok):
        return "supports" if ok else "contradicts"

    out = {}
    fails = {}

    def judge(arm, scored, checks):
        bad = [c for c, ok in checks if not ok]
        fails[arm] = bad
        if env(arm, scored):
            return "inconclusive"
        return word(not bad)

    r = lambda arm, c: rows[(arm, c)]
    out["h-text-warn"] = judge("text-warn", slips + guards,
        [(c, r("text-warn", c) == (0, "y", "y")) for c in slips] +
        [(c, clean_guard(r("text-warn", c))) for c in guards])
    out["h-text-refuse"] = judge("text-refuse", slips + guards,
        [(c, refused(r("text-refuse", c))) for c in slips] +
        [(c, clean_guard(r("text-refuse", c))) for c in guards])
    out["h-cite"] = judge("cite", cslips + cguards + guards,
        [(c, refused(r("cite", c))) for c in cslips] +
        [(c, r("cite", c)[:2] == (0, "y")) for c in cguards] +
        [(c, r("cite", c) == r("today", c)) for c in guards])
    if out["h-cite"] != "inconclusive" and env("today", guards):
        out["h-cite"] = "inconclusive"
    slipped = [c for c in slips if r("today", c) == (0, "y", "n")]
    fails["today"] = slipped
    out["h-null"] = "inconclusive" if env("today", slips) else word(not slipped)

    for arm, label in ARMS:
        print(f"{arm}: {label} {out[label]}; failing cases: {', '.join(fails[arm]) or 'none'}")
    for _, label in ARMS:
        print(f"verdict: {label}={out[label]}")


if __name__ == "__main__":
    main(sys.argv[1:])
