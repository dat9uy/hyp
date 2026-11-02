#!/usr/bin/env python3
"""verdict --row ROW [--row ROW]... SCORE...: the skill-text gate over score outputs.

ROW is P5 (f_launch_product), P6 (f_reading_pin) or S1 (f_body_limit). Each
SCORE is a file holding one trial's score lines. Every arm, base and each
text arm, must have exactly 5 trials with distinct trial values.

For each row, an arm's count k is its number of failing trials out of 5. A
trial with skill_loaded 0 counts against its arm:

  base       k counts trials with skill_loaded 1 and the fingerprint 1; a
             trial that never loaded a skill is never a reproduction
  text arm   k counts trials with the fingerprint 1 or skill_loaded 0

A row reproduces when base's k >= 3. A text arm passes a row at k <= 1.

  verdict: repro=supports, none=contradicts   when any given row reproduces
  verdict: repro=contradicts, none=supports   when none does
  verdict: ARM=supports        for a text arm that passes every reproducing row
  verdict: ARM=contradicts     for one that fails a reproducing row
  verdict: ARM=inconclusive    for every text arm when no row reproduces

Prints one line per row and arm, then the verdict lines, text arms sorted.
Exits 2 on a usage error, an unknown row, an unreadable file, a missing
field, an arm named repro or none or not a label, no base arm, or an arm
without exactly 5 distinct trials.
"""
import re
import sys

ROWS = {"P5": "f_launch_product", "P6": "f_reading_pin", "S1": "f_body_limit"}
TRIALS = 5
LABEL = re.compile(r"^[a-z][a-z0-9-]{0,31}$")


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    args, rows = sys.argv[1:], []
    while len(args) >= 2 and args[0] == "--row":
        if args[1] not in ROWS:
            fail(f"unknown row {args[1]} (want P5, P6 or S1)")
        if args[1] not in rows:
            rows.append(args[1])
        args = args[2:]
    if not rows or not args:
        fail("usage: verdict --row ROW [--row ROW]... SCORE...")
    need = {"arm", "trial", "skill_loaded"} | {ROWS[r] for r in rows}
    arms = {}
    for p in args:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        if not need <= kv.keys():
            fail(f"{p} lacks " + ", ".join(sorted(need - kv.keys())))
        a = kv["arm"]
        if not LABEL.match(a) or a in ("repro", "none"):
            fail(f"{p} names arm {a!r}")
        arms.setdefault(a, {})
        if kv["trial"] in arms[a]:
            fail(f"{p} repeats {a} trial {kv['trial']}")
        arms[a][kv["trial"]] = kv
    if "base" not in arms:
        fail("no base trial")
    for a, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {a} has {len(t)} trials, want {TRIALS}")

    def count(a, field):
        if a == "base":
            return sum(1 for kv in arms[a].values() if kv["skill_loaded"] == "1" and kv[field] == "1")
        return sum(1 for kv in arms[a].values() if kv[field] == "1" or kv["skill_loaded"] != "1")

    texts = sorted(a for a in arms if a != "base")
    repro, passes = [], {a: True for a in texts}
    for r in rows:
        k = count("base", ROWS[r])
        ok = k >= 3
        print(f"{r} base: {k}/{TRIALS} " + ("reproduces" if ok else "does not reproduce"))
        if ok:
            repro.append(r)
        for a in texts:
            ka = count(a, ROWS[r])
            print(f"{r} {a}: {ka}/{TRIALS} " + ("passes" if ka <= 1 else "fails"))
            if ok and ka > 1:
                passes[a] = False
    print("verdict: repro=" + ("supports" if repro else "contradicts"))
    print("verdict: none=" + ("contradicts" if repro else "supports"))
    for a in texts:
        word = "inconclusive" if not repro else ("supports" if passes[a] else "contradicts")
        print(f"verdict: {a}={word}")


if __name__ == "__main__":
    main()
