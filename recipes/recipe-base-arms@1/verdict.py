#!/usr/bin/env python3
"""verdict SCORE...: e-base's pre-registered rule over score outputs (#270).

Each SCORE is a file holding one trial's score lines (arm, trial, model,
stops, end_state, edited_lines). The arms are null, coarse and fine. Every
arm must have the same number of trials, 3 or 5, with distinct trial
values, and every trial's model must be sonnet.

For each arm a: pass_a is its count of end_state pass, s_a its mean stops
and e_a its mean edited_lines. A base arm b (coarse or fine) beats null
when pass_b >= pass_null and s_b <= s_null - 1 (the bar of 1 stop, set by
the owner).

  extension: with 3 trials per arm, when some base arm's gap s_null - s_b
      lies in [0.5, 1.5], print "extend: yes" and every label
      inconclusive: run 2 more trials per arm and run verdict again on
      all 5. With 5 trials, or no gap in the band, print "extend: no".
  neither base arm beats null:
      h-null=supports h-coarse=contradicts h-fine=contradicts
  one base arm beats null:
      it supports, the other contradicts, h-null=contradicts
  both beat null: h-null=contradicts, and the one with the lower e_a
      supports and the other contradicts, unless |e_coarse - e_fine| <=
      0.10 * max(e_coarse, e_fine), when both are inconclusive

Prints one summary line per arm, the extend line, then the verdict lines.
Exits 2 when a file is unreadable or lacks a field, names another arm or a
model other than sonnet, when an arm has no trial, when the arms' trial
counts differ or are not 3 or 5, or when an arm repeats a trial.
"""
import sys

ARMS = ("null", "coarse", "fine")
FIELDS = ("arm", "trial", "model", "stops", "end_state", "edited_lines")


def fail(msg):
    print("verdict: " + msg, file=sys.stderr)
    sys.exit(2)


def main():
    if len(sys.argv) < 2:
        fail("usage: verdict SCORE...")
    arms = {a: {} for a in ARMS}
    for p in sys.argv[1:]:
        try:
            kv = dict(l.rstrip("\n").split("\t", 1) for l in open(p) if "\t" in l)
        except OSError as e:
            fail(f"cannot read {p}: {e}")
        for f in FIELDS:
            if f not in kv:
                fail(f"{p} lacks {f}")
        if kv["arm"] not in arms:
            fail(f"{p} names arm {kv['arm']}")
        if kv["model"] != "sonnet":
            fail(f"{p}: the runner's model is {kv['model']}, not sonnet")
        if kv["trial"] in arms[kv["arm"]]:
            fail(f"{p}: arm {kv['arm']} has trial {kv['trial']} twice")
        arms[kv["arm"]][kv["trial"]] = (int(kv["stops"]), kv["end_state"] == "pass", int(kv["edited_lines"]))
    for a, t in arms.items():
        if not t:
            fail(f"no {a} trial")
    n = {len(t) for t in arms.values()}
    if len(n) != 1 or n.pop() not in (3, 5):
        fail("every arm needs the same number of trials, 3 or 5: " +
             ", ".join(f"{a} {len(t)}" for a, t in arms.items()))
    stat = {}
    for a in ARMS:
        t = list(arms[a].values())
        s = sum(x[0] for x in t) / len(t)
        p = sum(1 for x in t if x[1])
        e = sum(x[2] for x in t) / len(t)
        stat[a] = (s, p, e, len(t))
        print(f"{a}: trials {len(t)}, mean stops {s:.2f}, end_state pass {p}, mean edited_lines {e:.1f}")
    s_null, p_null, _, trials = stat["null"]
    gap = {b: s_null - stat[b][0] for b in ("coarse", "fine")}
    if trials == 3 and any(0.5 <= g <= 1.5 for g in gap.values()):
        print("extend: yes")
        for label in ("h-null", "h-coarse", "h-fine"):
            print(f"verdict: {label}=inconclusive")
        return
    print("extend: no")
    beats = {b: stat[b][1] >= p_null and stat[b][0] <= s_null - 1 for b in ("coarse", "fine")}
    v = {}
    if not any(beats.values()):
        v = {"h-null": "supports", "h-coarse": "contradicts", "h-fine": "contradicts"}
    elif beats["coarse"] != beats["fine"]:
        win = "coarse" if beats["coarse"] else "fine"
        lose = "fine" if win == "coarse" else "coarse"
        v = {"h-null": "contradicts", "h-" + win: "supports", "h-" + lose: "contradicts"}
    else:
        ec, ef = stat["coarse"][2], stat["fine"][2]
        v["h-null"] = "contradicts"
        if abs(ec - ef) <= 0.10 * max(ec, ef):
            v["h-coarse"] = v["h-fine"] = "inconclusive"
        else:
            win = "coarse" if ec < ef else "fine"
            lose = "fine" if win == "coarse" else "coarse"
            v["h-" + win], v["h-" + lose] = "supports", "contradicts"
    for label in ("h-null", "h-coarse", "h-fine"):
        print(f"verdict: {label}={v[label]}")


if __name__ == "__main__":
    main()
