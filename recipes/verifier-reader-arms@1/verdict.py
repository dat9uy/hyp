#!/usr/bin/env python3
"""verdict SCORE...: the pre-registered rule over one score file per trial.

The rule: 5 sonnet trials per arm, void re-run. Caught: every known wrong verdict,
including a new #26 kind, answered wrong in a tool-free follow-up, and an
insufficient confirm; each scored alone. A reader with no answer is void. Base
caught <=2 else cut invalid, stop; h-null contradicts at <=2. Others: supports at
>=4 with false alarms in <=1 trial, contradicts at <=2, else inconclusive. Two
supported: owner on cost. Winner must pass the real #26 cut.

Void trials are dropped; the rest must give each arm present exactly 5 distinct
trials. Prints one line per arm:
arm<TAB>caught<TAB>wrong on deciding<TAB>wrong on lone<TAB>wrong on loader<TAB>false-alarm trials<TAB>opened all three.

  cut: valid    base caught <= 2, else cut: invalid.
  h-null        contradicts at base caught <= 2, supports at >= 4, else
                inconclusive.
  h-compare, h-session, h-reader
                from arms compare, session and reader: supports at caught >= 4
                with false-alarm trials <= 1; contradicts at caught <= 2; else
                inconclusive. An arm with caught >= 4 and false-alarm trials >= 2
                is inconclusive.
  owner: trade-off
                printed when two or more of those three are supported.

h-null can read supports here although the rule stops at base caught >= 3,
because the stop fires first. No verdict reads opened_*: the last column is
reported only. Base alone is always accepted: it prints base's line, cut and
h-null, and no other line. Any other arm needs all three and base, and a valid
cut. Exits 2 on a usage error, an unreadable file, a missing field, an unknown
arm, a repeated trial, an arm without 5 non-void trials, no base, a partial set
of arms, or an arm other than base on an invalid cut.
"""
import sys

FIELDS = {"arm", "trial", "void", "caught", "answer_deciding", "answer_lone", "answer_loader",
          "opened_deciding", "opened_lone", "opened_loader", "false_alarm"}
TRIALS = 5
RIVALS = {"compare": "h-compare", "session": "h-session", "reader": "h-reader"}


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    if len(sys.argv) < 2:
        fail("usage: verdict SCORE...")
    arms = {}
    for p in sys.argv[1:]:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        if not FIELDS <= kv.keys():
            fail(f"{p} lacks " + ", ".join(sorted(FIELDS - kv.keys())))
        a = kv["arm"]
        if a != "base" and a not in RIVALS:
            fail(f"{p} names arm {a!r}")
        if kv["void"] == "1":
            continue
        trials = arms.setdefault(a, {})
        if kv["trial"] in trials:
            fail(f"{p} repeats {a} trial {kv['trial']}")
        trials[kv["trial"]] = kv
    if "base" not in arms:
        fail("no base trial")
    for a, t in arms.items():
        if len(t) != TRIALS:
            fail(f"arm {a} has {len(t)} non-void trials, want {TRIALS}")

    def count(a, pred):
        return sum(1 for kv in arms[a].values() if pred(kv))

    def caught(a):
        return count(a, lambda kv: kv["caught"] == "1")

    def alarms(a):
        return count(a, lambda kv: kv["false_alarm"] == "1")

    base = caught("base")
    valid = base <= 2
    given = set(arms) - {"base"}
    if given and given != set(RIVALS):
        fail("an arm needs the other two: got " + " ".join(sorted(given)))
    if given and not valid:
        fail(f"base caught is {base}, an invalid cut, but got: " + " ".join(sorted(given)))

    for a in sorted(arms):
        wrong = [count(a, lambda kv: kv[f"answer_{w}"] == "wrong") for w in ("deciding", "lone", "loader")]
        opened = count(a, lambda kv: all(kv[f"opened_{w}"] == "1" for w in ("deciding", "lone", "loader")))
        print("\t".join(str(x) for x in (a, caught(a), *wrong, alarms(a), opened)))
    print("cut: valid" if valid else "cut: invalid")
    print("verdict: h-null=" + ("contradicts" if base <= 2 else "supports" if base >= 4 else "inconclusive"))
    if not given:
        return

    def word(a):
        if caught(a) >= 4 and alarms(a) <= 1:
            return "supports"
        return "contradicts" if caught(a) <= 2 else "inconclusive"

    words = {a: word(a) for a in RIVALS}
    for a, label in RIVALS.items():
        print(f"verdict: {label}={words[a]}")
    if sum(1 for w in words.values() if w == "supports") >= 2:
        print("owner: trade-off")


if __name__ == "__main__":
    main()
