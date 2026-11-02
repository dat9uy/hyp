#!/usr/bin/env python3
import json
import os
import sys

import gatelib as g

if len(sys.argv) != 2:
    g.die("usage: score.py OUT_DIR")
out = sys.argv[1]


def need(name):
    p = os.path.join(out, name)
    if not os.path.isfile(p):
        g.die("score: %s is not in OUT_DIR (keep tmp/skill-gate/%s)" % (name, name))
    return p


with open(need("state.json")) as f:
    state = json.load(f)
msgs = g.read_jsonl(need("transcript.jsonl"))
with open(need("events.json")) as f:
    events = json.load(f)
need("summary.json")
loaded, failing, f_pin, f_body = g.score_transcript(msgs)
rows = [
    ("arm", state["arm"]),
    ("trial", state["trial"]),
    ("skill_loaded", loaded),
    ("failing_hyp", failing),
    ("f_reading_pin", f_pin),
    ("f_body_limit", f_body),
    ("end_state", g.end_state(events, state["cut_events"])),
]
for k, v in rows:
    print("%s\t%s" % (k, v))
