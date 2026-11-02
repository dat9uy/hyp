#!/usr/bin/env python3
import json
import os
import signal
import subprocess
import time

import gatelib as g

TIMEOUT = 50 * 60

wt = os.environ["HYP_SANDBOX_WT"]
tmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
inputs = g.read_inputs(os.environ["HYP_SANDBOX_INPUTS"])
if os.path.exists(os.path.join(tmp, "summary.json")):
    g.die("drive: this sandbox was driven already; one sandbox is one trial", 1)

env = dict(os.environ)
env["HYP_ACTOR"] = "runner-%s" % inputs["trial"]
for k in ("HYP_SANDBOX_INPUTS", "HYP_SANDBOX_RECIPE_TMP"):
    env.pop(k, None)
cmd = ["claude", "-p", "--model", "sonnet", "--setting-sources", "project",
       "--permission-mode", "bypassPermissions", "--output-format", "stream-json",
       "--verbose", "--no-session-persistence"]
t0 = time.time()
timed_out = False
with open(inputs["brief"], "rb") as brief, \
        open(os.path.join(tmp, "transcript.jsonl"), "wb") as out:
    try:
        p = subprocess.Popen(cmd, cwd=wt, env=env, stdin=brief, stdout=out,
                             start_new_session=True)
    except OSError as e:
        g.die("drive: cannot start claude: %s" % e, 1)
    try:
        code = p.wait(timeout=TIMEOUT)
    except subprocess.TimeoutExpired:
        timed_out = True
        code = None
    try:
        os.killpg(p.pid, signal.SIGKILL)
    except OSError:
        pass
    if timed_out:
        p.wait()

turns = cost = None
for m in g.read_jsonl(os.path.join(tmp, "transcript.jsonl")):
    if m.get("type") == "result":
        turns, cost = m.get("num_turns"), m.get("total_cost_usd")
summary = {"format": 1, "exit": code, "timed_out": timed_out, "turns": turns,
           "cost_usd": cost, "seconds": round(time.time() - t0, 1)}
with open(os.path.join(tmp, "summary.json"), "w") as f:
    json.dump(summary, f, indent=1, sort_keys=True)
    f.write("\n")

ev = subprocess.run(["hyp", "events", "--json"], cwd=wt, capture_output=True, text=True)
with open(os.path.join(tmp, "events.json"), "w") as f:
    f.write(ev.stdout if ev.returncode == 0 else json.dumps({"format": 1, "events": [], "error": ev.stderr}))
print("drive: %s" % json.dumps(summary, sort_keys=True))
