#!/usr/bin/env python3
"""score OUT_DIR SCRIPTS_DIR: measure one grill trial; print no verdict.

OUT_DIR is the OUT_DIR of hyp sandbox remove, holding the kept files.
SCRIPTS_DIR is ak-skill-creator's scripts/ directory; its eval_skill.py
must have the pinned sha256.

Writes under $HYP_OUT:
- outputs/: what a reviewer who does not know the arm reads, and what
  eval_skill.py grades: conversation.md (the grill's session, with each
  tool call's working directory and the loaded grill skill's text
  redacted), goals.md (each Goal with lines and each other node the grill
  added), and has_goal.json, l1.json and l4.json ({"pass": true|false}).
- record.json: eval_skill.py grade's record for the fixture's case.
- facts.json: the measures below, with each trial Goal.
Prints key<TAB>value lines: arm, fixture, trial, end, exchanges,
timed_out, cost_usd, cost_usd_literal, models, subagents, goals, stores,
has_goal, l1, l4, eval_exit. Exits 2 when OUT_DIR lacks a kept file or
eval_skill.py is not the pinned one.
"""
import hashlib
import json
import os
import re
import subprocess
import sys
import tarfile
import tempfile
from datetime import datetime

EVAL_SHA = "c47a38dee2456cc6172dff28472052c0f7b28d7baf310af6e83b5ca655d10a24"
KEPT = ["state.txt", "drive.json", "calls.json", "transcript.jsonl", "subagents.tar", "standin.jsonl", "stores.tar", "skill.tar"]
PARENT_STORE = {"a": None, "b": "investigations/goal-prompt"}
TREES = "investigations"
CUT_HEX = re.compile(r"(?:^|[^0-9A-Fa-f])([0-9A-Fa-f]{7,63})(?:…|\.\.\.)")
SHA_SHORT = re.compile(r"sha-?256\W{0,12}([0-9A-Fa-f]{7,63})(?![0-9A-Fa-f])", re.I)
GRILL_SKILL = ".claude/skills/hyp-owner-grill"
BUNDLE = os.path.dirname(os.path.abspath(__file__))


def die(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def state(path):
    out = {}
    for line in open(path):
        if "=" in line:
            k, v = line.rstrip("\n").split("=", 1)
            out[k] = v
    return out


def ts(s):
    return datetime.fromisoformat(s.replace("Z", "+00:00")).timestamp()


def clip(s, n):
    s = s if isinstance(s, str) else json.dumps(s)
    return s if len(s) <= n else s[:n] + "\n[... %d more characters]" % (len(s) - n)


def text_of(content):
    if isinstance(content, str):
        return content
    out = []
    for b in content or []:
        if isinstance(b, dict) and b.get("type") == "text":
            out.append(b.get("text", ""))
    return "\n".join(out)


def conversation(path):
    """The grill's session as Markdown, in order, with tool calls."""
    lines, root, redact, n = [], None, set(), 0
    for raw in open(path):
        try:
            e = json.loads(raw)
        except ValueError:
            continue
        if e.get("isSidechain") or e.get("type") not in ("user", "assistant"):
            continue
        if root is None and e.get("cwd"):
            root = e["cwd"]
        m = e.get("message") or {}
        c = m.get("content")
        if e["type"] == "user":
            if e.get("isMeta"):
                continue
            if isinstance(c, list) and any(isinstance(b, dict) and b.get("type") == "tool_result" for b in c):
                for b in c:
                    if not (isinstance(b, dict) and b.get("type") == "tool_result"):
                        continue
                    body = "[redacted: the loaded grill skill]" if b.get("tool_use_id") in redact else clip(text_of(b.get("content")) if not isinstance(b.get("content"), str) else b["content"], 2000)
                    lines.append("Result:\n\n```text\n%s\n```\n" % body.replace("```", "'''"))
                continue
            t = text_of(c).strip()
            if t:
                n += 1
                lines.append("## Owner, message %d\n\n%s\n" % (n, t))
            continue
        for b in c or []:
            if not isinstance(b, dict):
                continue
            if b.get("type") == "text" and b.get("text", "").strip():
                lines.append("## Grill\n\n%s\n" % b["text"].strip())
            elif b.get("type") == "tool_use":
                inp = b.get("input") or {}
                if GRILL_SKILL in json.dumps(inp):
                    redact.add(b.get("id"))
                cwd = os.path.relpath(e.get("cwd") or root, root) if root else "?"
                lines.append("### Tool %s (cwd: %s)\n\n```text\n%s\n```\n" % (b.get("name"), cwd, clip(json.dumps(inp, indent=1, ensure_ascii=False), 3000).replace("```", "'''")))
    return "\n".join(lines)


def models(path):
    out = set()
    for raw in open(path):
        try:
            e = json.loads(raw)
        except ValueError:
            continue
        if e.get("type") == "assistant":
            mod = (e.get("message") or {}).get("model")
            if mod and not mod.startswith("<"):
                out.add(mod)
    return out


def main():
    if len(sys.argv) != 3:
        die("usage: score OUT_DIR SCRIPTS_DIR")
    od, scripts = sys.argv[1], sys.argv[2]
    for k in KEPT:
        if not os.path.isfile(os.path.join(od, k)):
            die("OUT_DIR has no " + k)
    ev = os.path.join(scripts, "eval_skill.py")
    if not os.path.isfile(ev) or hashlib.sha256(open(ev, "rb").read()).hexdigest() != EVAL_SHA:
        die("eval_skill.py in SCRIPTS_DIR is not sha256 " + EVAL_SHA)
    st = state(os.path.join(od, "state.txt"))
    drive = json.load(open(os.path.join(od, "drive.json")))
    fixture = st["fixture"]
    out = os.environ["HYP_OUT"]
    outputs = os.path.join(out, "outputs")
    os.makedirs(outputs)
    tmp = tempfile.mkdtemp()

    # Every store at the end, and the nodes the trial added to it.
    with tarfile.open(os.path.join(od, "stores.tar")) as t:
        t.extractall(os.path.join(tmp, "s"), filter="data")
    stores, goals, others = [], [], []
    for root, dirs, _ in os.walk(os.path.join(tmp, "s")):
        dirs.sort()
        if ".tree" in dirs:
            rel = os.path.relpath(root, os.path.join(tmp, "s"))
            stores.append(rel)
            dirs.remove(".tree")
            r = subprocess.run(["hyp", "events", "--json"], cwd=root, capture_output=True, text=True)
            if r.returncode != 0:
                continue
            for e in json.loads(r.stdout).get("events") or []:
                if e.get("type") != "node.add" or ts(e["ts"]) < drive["start"]:
                    continue
                p = e.get("payload") or {}
                node = {"store": rel, "id": e["id"], "actor": e.get("actor"), "kind": p.get("kind"), "title": p.get("title", ""),
                        "body": p.get("body", ""), "lines": [l.get("text", "") for l in p.get("lines") or []]}
                (goals if str(p.get("kind")).lower() == "goal" and node["lines"] else others).append(node)
    has_goal = bool(goals)
    want = PARENT_STORE[fixture]
    ok_store = all((g["store"] == want) if want else re.fullmatch(TREES + r"/[^/]+", g["store"]) for g in goals)
    l1 = has_goal and ok_store and "." not in stores
    bad_hash = [l for g in goals for l in g["lines"] if CUT_HEX.search(l) or SHA_SHORT.search(l)]
    l4 = has_goal and not bad_hash

    for name, v in (("has_goal", has_goal), ("l1", l1), ("l4", l4)):
        with open(os.path.join(outputs, name + ".json"), "w") as f:
            json.dump({"pass": v}, f)
            f.write("\n")
    with open(os.path.join(outputs, "conversation.md"), "w") as f:
        f.write("# The grill's session\n\n" + conversation(os.path.join(od, "transcript.jsonl")))
    with open(os.path.join(outputs, "goals.md"), "w") as f:
        f.write("# Nodes the grill added\n\nStores at the end: %s\n" % (", ".join(stores) or "none"))
        for g in goals:
            f.write("\n## Goal in %s, by %s: %s\n\nBody:\n\n```text\n%s\n```\n\nLines:\n\n" % (g["store"], g["actor"], g["title"], g["body"]))
            f.write("".join("%d. %s\n" % (i + 1, l) for i, l in enumerate(g["lines"])))
        for o in others:
            f.write("\n## %s in %s, by %s: %s\n\n%s\n" % (o["kind"], o["store"], o["actor"], o["title"], o["body"]))

    # eval_skill.py grades the outputs against the fixture's case.
    skill = os.path.join(tmp, "skill")
    with tarfile.open(os.path.join(od, "skill.tar")) as t:
        t.extractall(skill, filter="data")
    env = dict(os.environ, PYTHONDONTWRITEBYTECODE="1")
    r = subprocess.run([sys.executable, ev, "grade", os.path.join(BUNDLE, "cases.json"), "--case-id", "fixture-" + fixture,
                        "--outputs", outputs, "--run-id", "%s-%s-%s" % (st["arm"], fixture, st["trial"]),
                        "--variant", st["arm"], "--model", drive["model"], "--runtime", drive["claude_version"] or "unknown",
                        "--skill", os.path.join(skill, "hyp-owner-grill")], capture_output=True, text=True, env=env)
    if r.returncode not in (0, 1):
        die("eval_skill.py grade exited %d: %s" % (r.returncode, r.stderr.strip()[-500:]))
    with open(os.path.join(out, "record.json"), "w") as f:
        f.write(r.stdout)

    seen = models(os.path.join(od, "transcript.jsonl")) | models(os.path.join(od, "standin.jsonl"))
    with tarfile.open(os.path.join(od, "subagents.tar")) as t:
        t.extractall(os.path.join(tmp, "sub"), filter="data")
    for root, _, files in os.walk(os.path.join(tmp, "sub")):
        for name in files:
            if name.endswith(".jsonl"):
                seen |= models(os.path.join(root, name))
    facts = [("arm", st["arm"]), ("fixture", fixture), ("trial", st["trial"]), ("end", drive["end"]),
             ("exchanges", drive["exchanges"]), ("timed_out", int(bool(drive["timed_out"]))),
             ("cost_usd", "none" if drive["cost_usd"] is None else drive["cost_usd"]),
             ("cost_usd_literal", "none" if drive["cost_usd_literal"] is None else drive["cost_usd_literal"]),
             ("models", ",".join(sorted(seen)) or "none"), ("subagents", drive["subagents"]),
             ("goals", len(goals)), ("stores", ",".join(stores) or "none"),
             ("has_goal", int(has_goal)), ("l1", int(l1)), ("l4", int(l4)), ("eval_exit", r.returncode)]
    with open(os.path.join(out, "facts.json"), "w") as f:
        json.dump({"measures": dict(facts), "goals": goals, "bad_hash_lines": bad_hash}, f, indent=1)
        f.write("\n")
    for k, v in facts:
        print("%s\t%s" % (k, v))


if __name__ == "__main__":
    main()
