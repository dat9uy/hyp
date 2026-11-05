#!/usr/bin/env python3
"""verdict SCORE...: the pre-registered rule over one score file per trial.

Void trials are dropped; the rest must give each arm present exactly 5 distinct
trials. For an arm, caught counts trials with caught 1, opened those with both
wrong Results' files opened, control_flagged those with control_flagged 1, and
holds_on_seed those with holds_on_seed 1. Prints one line per arm:
arm<TAB>caught<TAB>opened<TAB>control_flagged<TAB>holds_on_seed.

  cut: valid    base caught <= 1, else cut: invalid.
  h-null        contradicts at base caught <= 1, supports at >= 4, else
                inconclusive.
  h-audit, h-readers
                from arms audit and readers: supports at caught >= 4 with
                control_flagged <= 1; contradicts at caught <= 2; else
                inconclusive. An arm with caught >= 4 and control_flagged >= 2
                is inconclusive.
  abandon: yes  when both audit arms contradict, or either reads holds on a
                wrong Result in >= 3 trials; else abandon: no.

Base alone is always accepted: it prints base's line, cut and h-null, and no
other verdict or abandon line. An audit arm needs the other audit arm and base,
and a valid cut. Exits 2 on a usage error, an unreadable file, a missing field,
an unknown arm, a repeated trial, an arm without 5 non-void trials, no base, one
audit arm alone, or an audit arm on an invalid cut.
"""
import sys

FIELDS = {"arm", "trial", "void", "caught", "opened_deciding", "opened_lone", "control_flagged", "holds_on_seed"}
TRIALS = 5
AUDIT_ARMS = {"audit": "h-audit", "readers": "h-readers"}


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
        if a != "base" and a not in AUDIT_ARMS:
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

    def flagged(a):
        return count(a, lambda kv: kv["control_flagged"] == "1")

    def holds(a):
        return count(a, lambda kv: kv["holds_on_seed"] == "1")

    base = caught("base")
    valid = base <= 1
    given = set(arms) - {"base"}
    if given and given != set(AUDIT_ARMS):
        fail("an audit arm needs the other: got " + " ".join(sorted(given)))
    if given and not valid:
        fail(f"base caught is {base}, an invalid cut, but got: " + " ".join(sorted(given)))

    for a in sorted(arms):
        opened = count(a, lambda kv: kv["opened_deciding"] == "1" and kv["opened_lone"] == "1")
        print(f"{a}\t{caught(a)}\t{opened}\t{flagged(a)}\t{holds(a)}")
    print("cut: valid" if valid else "cut: invalid")
    print("verdict: h-null=" + ("contradicts" if base <= 1 else "supports" if base >= 4 else "inconclusive"))
    if not given:
        return

    def word(a):
        if caught(a) >= 4 and flagged(a) <= 1:
            return "supports"
        return "contradicts" if caught(a) <= 2 else "inconclusive"

    words = {a: word(a) for a in AUDIT_ARMS}
    for a, label in AUDIT_ARMS.items():
        print(f"verdict: {label}={words[a]}")
    abandon = all(w == "contradicts" for w in words.values()) or any(holds(a) >= 3 for a in AUDIT_ARMS)
    print("abandon: yes" if abandon else "abandon: no")


if __name__ == "__main__":
    main()
