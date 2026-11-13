#!/usr/bin/env python3
"""S1, a strict check that sets no verdict: line 5 read as 'every id string in
any events line, rejected lines included'. With the hyp binary given, in a
new store: add a Goal, write one events line whose id replay rejects for its
form, add a DesignQuestion, and report whether the new id sorts after the
rejected id. Known answers: a rejected id that sorts low passes; the
lower-case 01mzzzzzzzzzzzzzzzzzzzzzzz fails."""
import json, os, subprocess, sys, tempfile
hyp = sys.argv[1]
CASES = [("known answer: low rejected id", "0000000000000000000000000x", "pass"),
         ("R-53N6JZ#1 case: lower-case high id", "01mzzzzzzzzzzzzzzzzzzzzzzz", "fail")]
env = dict(os.environ, HYP_ACTOR="a")
for name, bad, expect in CASES:
    d = tempfile.mkdtemp()
    run = lambda *a: subprocess.run([hyp, *a], cwd=d, env=env, capture_output=True, text=True)
    run("init")
    g = run("add", "goal", "--title", "g", "--why", "w").stdout.split()[0]
    with open(os.path.join(d, ".tree", "events", "z.jsonl"), "w") as f:
        f.write(json.dumps({"v": 1, "id": bad, "ts": "2026-10-27T00:00:00Z", "actor": "z", "type": "node.add", "node": bad, "payload": {"kind": "Goal", "parent": "", "title": "bad"}, "why": "w"}) + "\n")
    r = run("add", "dq", "--parent", g, "--title", "q", "--why", "w")
    ids = [json.loads(l)["id"] for l in open(os.path.join(d, ".tree", "events", "a.jsonl"))]
    new = ids[-1]
    got = "pass" if r.returncode == 0 and new > bad else "fail"
    print(f"{name}: rejected id {bad}, add dq exit {r.returncode}, new id {new}: {got} (known answer {expect})")
