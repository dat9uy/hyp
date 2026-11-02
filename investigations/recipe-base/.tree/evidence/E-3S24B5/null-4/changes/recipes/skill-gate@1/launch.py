#!/usr/bin/env python3
import json
import os
import shutil
import subprocess
import sys

import gatelib as g

wt = os.environ["HYP_SANDBOX_WT"]
tmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
inputs = g.read_inputs(os.environ["HYP_SANDBOX_INPUTS"])

for name in ("AGENTS.md", "CLAUDE.md", ".claude", "plans", "PLAN.md", "PLANS.md", "METHOD.md",
             ".agentkit", "investigations"):
    p = os.path.join(wt, name)
    if os.path.islink(p) or os.path.isfile(p):
        os.remove(p)
    elif os.path.isdir(p):
        shutil.rmtree(p)

os.makedirs(os.path.join(wt, ".claude", "skills"))
for sk in g.SKILLS:
    src = os.path.join(wt, "skills", sk)
    if not os.path.isfile(os.path.join(src, "SKILL.md")):
        g.die("launch: the commit has no skills/%s/SKILL.md" % sk, 1)
    shutil.move(src, os.path.join(wt, ".claude", "skills", sk))
if os.path.isdir(os.path.join(wt, "skills")) and not os.listdir(os.path.join(wt, "skills")):
    os.rmdir(os.path.join(wt, "skills"))

variant_files = {}
if inputs["arm"] != "base":
    for sk in sorted(os.listdir(inputs["variant"])):
        dst = os.path.join(wt, ".claude", "skills", sk, "SKILL.md")
        shutil.copyfile(os.path.join(inputs["variant"], sk, "SKILL.md"), dst)
        variant_files[os.path.relpath(dst, wt)] = g.sha256_file(dst)

ev = subprocess.run(["hyp", "events", "--json"], cwd=wt, capture_output=True, text=True)
if ev.returncode != 0:
    g.die("launch: hyp events --json failed (is the tree at the cut there?): " + ev.stderr, 1)
cut_ids = [e["id"] for e in json.loads(ev.stdout)["events"]]

state = {
    "format": 1,
    "arm": inputs["arm"],
    "trial": int(inputs["trial"]),
    "brief_sha256": g.sha256_file(inputs["brief"]),
    "variant_files": variant_files,
    "skill_files": g.skill_hashes(wt),
    "cut_events": cut_ids,
}
with open(os.path.join(tmp, "state.json"), "w") as f:
    json.dump(state, f, indent=1, sort_keys=True)
    f.write("\n")
print("launch: arm %s trial %s, %d skill files, %d events at the cut" % (
    state["arm"], state["trial"], len(state["skill_files"]), len(cut_ids)))
