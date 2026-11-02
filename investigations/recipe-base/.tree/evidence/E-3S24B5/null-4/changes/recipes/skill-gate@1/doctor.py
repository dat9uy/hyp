#!/usr/bin/env python3
import json
import os
import shutil
import subprocess

import gatelib as g

wt = os.environ["HYP_SANDBOX_WT"]
tmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
with open(os.path.join(tmp, "state.json")) as f:
    state = json.load(f)
bad = []

for name in ("AGENTS.md", "CLAUDE.md", "plans", "PLAN.md", "PLANS.md", "METHOD.md", ".agentkit",
             "investigations", "skills"):
    if os.path.lexists(os.path.join(wt, name)):
        bad.append("%s is in the sandbox" % name)
cl = os.path.join(wt, ".claude")
if not os.path.isdir(cl) or sorted(os.listdir(cl)) != ["skills"]:
    bad.append(".claude holds more than skills")
elif sorted(os.listdir(os.path.join(cl, "skills"))) != sorted(g.SKILLS):
    bad.append(".claude/skills is not exactly the two hyp skills")
now = g.skill_hashes(wt)
if now != state["skill_files"]:
    bad.append("a skill file differs from what Launch recorded")
for p, h in state["variant_files"].items():
    if now.get(p) != h:
        bad.append("variant file %s is not laid over its skill" % p)
if state["arm"] == "base" and state["variant_files"]:
    bad.append("base arm with variant files")
if state["arm"] != "base" and not state["variant_files"]:
    bad.append("variant arm with no variant file")

ev = subprocess.run(["hyp", "events", "--json"], cwd=wt, capture_output=True, text=True)
if ev.returncode != 0:
    bad.append("hyp events --json failed: " + ev.stderr.strip())
else:
    doc = json.loads(ev.stdout)
    if [e["id"] for e in doc["events"]] != state["cut_events"]:
        bad.append("the tree's events are not the cut's")
    if doc.get("rejections"):
        bad.append("the tree at the cut has rejected events")
v = subprocess.run(["hyp", "version"], capture_output=True, text=True)
if v.returncode != 0:
    bad.append("hyp version failed")
else:
    print("doctor: hyp %s" % v.stdout.strip())
if not shutil.which("claude"):
    bad.append("claude is not on the PATH")

if bad:
    for b in bad:
        print("doctor: FAIL " + b)
    raise SystemExit(1)
print("doctor: ok")
