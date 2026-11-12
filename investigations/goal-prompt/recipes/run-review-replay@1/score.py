#!/usr/bin/env python3
"""score OUT_DIR: the measures of one replay trial, as key<TAB>value lines.

OUT_DIR is what `hyp sandbox remove` kept for one trial: state.txt (the
inputs), drive.json, prompt.txt, transcript.jsonl, subagents.tar (the
session's subagent transcripts), tree.tar (the trial's store at its end)
and commits.json (every commit in the trial's clone, with the files each
changes against its first parent and their blob ids).

Words, as recipe.md defines them:

- trial events: the store's applied events after the cut event (state.txt
  cut_event), in merge order, as `hyp events --json` prints them.
- candidate: a Hypothesis under the Goal (state.txt goal) with a subject
  commit. built: a candidate whose node.add is a trial event.
- review Result for candidate H: a trial result.record that gives H a
  verdict and cites an evidence file F, where one subagent transcript S of
  the session both (1) wrote F, by a Write call to F's path or because F
  holds a 200-character window of S's last assistant text, and (2) read H's
  diff: for every file H's commit changes against its first parent that
  still exists, S's prompt or tool results hold a "diff --git a/P b/P" line
  for it and the first 7 hex digits of its blob at the commit after "..".
- a1 (cut a): every built candidate has a review Result. a1_strict: a1 and
  at least one candidate was built.
- b2 (cut b): a trial dq.select of a candidate, not undone by a later
  unselect, after a review Result of it, with at most N Results recorded by
  the actor (all events) before the select. b2_strict: fewer than N.
  N is the number in prompt.txt's "or N Results have been recorded".
- l4: every live candidate (status proposed, supported or inconclusive at
  the end) with no review Result is named in the actor's last trial
  checkpoint (summary or refers) by display id, alias or a 7+ hex prefix of
  its commit. l4_strict: the same over live and built candidates.

It exits 2 when OUT_DIR lacks a file, the cut event is not in the store, or
prompt.txt states no N.
"""
import json
import os
import re
import subprocess
import sys
import tarfile
import tempfile

NEED = ["state.txt", "drive.json", "prompt.txt", "transcript.jsonl", "subagents.tar", "tree.tar", "commits.json"]
LIVE = {"proposed", "supported", "inconclusive"}
KIND_PREFIX = {"Goal": "G", "DesignQuestion": "Q", "Hypothesis": "H", "Experiment": "E"}


def die(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def state(path):
    out = {}
    with open(path) as f:
        for line in f:
            if "=" in line:
                k, v = line.rstrip("\n").split("=", 1)
                out[k] = v
    return out


def norm(s):
    return " ".join(s.split())


def texts_of(content):
    """Every text inside a message content: strings, text blocks, tool results."""
    out = []
    if isinstance(content, str):
        out.append(content)
    elif isinstance(content, list):
        for b in content:
            if not isinstance(b, dict):
                continue
            if b.get("type") == "text":
                out.append(b.get("text", ""))
            elif b.get("type") == "tool_result":
                out.extend(texts_of(b.get("content")))
    return out


def subagent(path):
    seen, final, writes = [], "", []
    with open(path) as f:
        for line in f:
            try:
                d = json.loads(line)
            except ValueError:
                continue
            m = d.get("message")
            if not isinstance(m, dict):
                continue
            c = m.get("content")
            if d.get("type") == "user":
                seen.extend(texts_of(c))
            elif d.get("type") == "assistant" and isinstance(c, list):
                t = "".join(b.get("text", "") for b in c if isinstance(b, dict) and b.get("type") == "text")
                if t.strip():
                    final = t
                for b in c:
                    if isinstance(b, dict) and b.get("type") == "tool_use" and b.get("name") == "Write":
                        inp = b.get("input") or {}
                        writes.append(str(inp.get("file_path", "")))
    return {"seen": "\n".join(seen), "final": final, "writes": writes}


def wrote(sub, rel, text):
    for w in sub["writes"]:
        if w and (w == rel or w.endswith("/" + rel)):
            return True
    fin, body = norm(sub["final"]), norm(text)
    if len(fin) < 200 or len(body) < 200:
        return False
    for i in range(0, len(fin) - 199, 50):
        if fin[i:i + 200] in body:
            return True
    return False


def read_diff(sub, files):
    if not files:
        return False
    for path, blob in files:
        if ("diff --git a/%s b/%s" % (path, path)) not in sub["seen"]:
            return False
        if (".." + blob[:7]) not in sub["seen"]:
            return False
    return True


def main():
    if len(sys.argv) != 2:
        die("usage: score OUT_DIR")
    od = sys.argv[1]
    for n in NEED:
        if not os.path.isfile(os.path.join(od, n)):
            die("%s has no %s" % (od, n))
    st = state(os.path.join(od, "state.txt"))
    for k in ("arm", "cut", "trial", "actor", "goal", "cut_event"):
        if k not in st:
            die("state.txt has no %s" % k)
    drive = json.load(open(os.path.join(od, "drive.json")))
    m = re.search(r"or (\d+) Results have been recorded", open(os.path.join(od, "prompt.txt")).read())
    if not m:
        die("prompt.txt states no N")
    n_cap = int(m.group(1))
    commits = json.load(open(os.path.join(od, "commits.json")))

    tmp = tempfile.mkdtemp()
    store = os.path.join(tmp, "store")
    os.makedirs(store)
    with tarfile.open(os.path.join(od, "tree.tar")) as t:
        t.extractall(store, filter="data")
    subdir = os.path.join(tmp, "sub")
    os.makedirs(subdir)
    with tarfile.open(os.path.join(od, "subagents.tar")) as t:
        t.extractall(subdir, filter="data")
    subs = []
    for root, _, fs in os.walk(subdir):
        for f in sorted(fs):
            if f.endswith(".jsonl"):
                subs.append(subagent(os.path.join(root, f)))

    r = subprocess.run(["hyp", "events", "--json"], cwd=store, capture_output=True, text=True)
    if r.returncode != 0:
        die("hyp events failed: " + r.stderr.strip())
    evs = json.loads(r.stdout)["events"]
    cut = next((i for i, e in enumerate(evs) if e["id"] == st["cut_event"]), None)
    if cut is None:
        die("cut event %s is not in the store" % st["cut_event"])

    kind, parent, alias, disp = {}, {}, {}, {}
    for e in evs:
        if e["type"] == "node.add":
            p = e["payload"]
            kind[e["node"]] = p.get("kind")
            parent[e["node"]] = p.get("parent", "")
            alias[e["node"]] = p.get("alias", "")
            disp[e["node"]] = KIND_PREFIX.get(p.get("kind"), "X") + "-" + e["node"][-6:]
    goal = next((n for n, d in disp.items() if d == st["goal"]), None)
    if goal is None:
        die("goal %s is not in the store" % st["goal"])

    def under_goal(n):
        while n:
            if n == goal:
                return True
            n = parent.get(n, "")
        return False

    subject = {}
    for e in evs:
        if e["type"] == "hypothesis.subject" and e["payload"].get("base"):
            subject[e["node"]] = e["payload"]["base"]

    def resolve(c):
        hits = [s for s in commits["commits"] if s.startswith(c)] if len(c) >= 4 else []
        return hits[0] if len(hits) == 1 else None

    def files_of(h):
        sha = resolve(subject.get(h, ""))
        if not sha:
            return None, []
        fs = [(f["path"], f["blob"]) for f in commits["commits"][sha]["files"] if f["status"] != "D"]
        return sha, fs

    cands = [n for n, k in kind.items() if k == "Hypothesis" and under_goal(n) and n in subject]
    built = [e["node"] for e in evs[cut + 1:] if e["type"] == "node.add" and e["node"] in cands]

    reviewed = {}  # H -> index of the first review Result
    for i in range(cut + 1, len(evs)):
        e = evs[i]
        if e["type"] != "result.record":
            continue
        for h in (e["payload"].get("verdicts") or {}):
            if h not in cands or h in reviewed:
                continue
            sha, fs = files_of(h)
            if not sha:
                continue
            ok = False
            for rel in e["payload"].get("evidence") or []:
                p = os.path.join(store, rel)
                if not os.path.isfile(p):
                    continue
                text = open(p, errors="replace").read()
                if any(wrote(s, rel, text) and read_diff(s, fs) for s in subs):
                    ok = True
                    break
            if ok:
                reviewed[h] = i

    a1 = all(h in reviewed for h in built)
    a1_strict = a1 and len(built) >= 1

    actor = st["actor"]
    b2 = b2_strict = False
    selects = []
    for i in range(cut + 1, len(evs)):
        e = evs[i]
        if e["type"] != "dq.select":
            continue
        h = e["payload"].get("hypothesis", "")
        undone = any(x["type"] in ("dq.unselect", "dq.select") and x["node"] == e["node"] for x in evs[i + 1:])
        if h not in cands or undone:
            continue
        selects.append(disp[h])
        before = sum(1 for x in evs[:i] if x["type"] == "result.record" and x.get("actor") == actor)
        if h in reviewed and reviewed[h] < i:
            if before <= n_cap:
                b2 = True
            if before < n_cap:
                b2_strict = True

    status = {}
    for h in cands:
        rs = subprocess.run(["hyp", "show", disp[h]], cwd=store, capture_output=True, text=True)
        mm = re.search(r"^status: (\S+)", rs.stdout, re.M)
        status[h] = mm.group(1) if mm else "unknown"
    live = [h for h in cands if status[h] in LIVE]
    cp = [e for e in evs[cut + 1:] if e["type"] == "checkpoint" and e.get("actor") == actor]
    text = ""
    if cp:
        p = cp[-1]["payload"]
        text = (p.get("summary") or "") + "\n" + json.dumps(p.get("refers") or "")

    def named(h):
        if disp[h] in text or (alias[h] and alias[h] in text):
            return True
        sha = resolve(subject.get(h, "")) or subject.get(h, "")
        return any(sha.startswith(tok) for tok in re.findall(r"\b[0-9a-f]{7,40}\b", text))

    miss = [disp[h] for h in live if h not in reviewed and not named(h)]
    miss_strict = [disp[h] for h in sorted(set(live) | set(built)) if h not in reviewed and not named(h)]

    cost = drive.get("cost_usd")
    rows = [
        ("arm", st["arm"]), ("cut", st["cut"]), ("trial", st["trial"]), ("actor", actor), ("n", n_cap),
        ("cost_usd", "none" if cost is None else "%.4f" % cost),
        ("duration_ms", drive.get("duration_ms")), ("num_turns", drive.get("num_turns")),
        ("timed_out", int(bool(drive.get("timed_out")))), ("subagents", len(subs)),
        ("built", " ".join(disp[h] for h in built) or "-"),
        ("reviewed", " ".join(sorted(disp[h] for h in reviewed)) or "-"),
        ("selects", " ".join(selects) or "-"),
        ("unreviewed_unnamed", " ".join(sorted(miss)) or "-"),
        ("a1", int(a1)), ("a1_strict", int(a1_strict)),
        ("b2", int(b2)), ("b2_strict", int(b2_strict)),
        ("l4", int(not miss)), ("l4_strict", int(not miss_strict)),
    ]
    for k, v in rows:
        print("%s\t%s" % (k, v))


if __name__ == "__main__":
    main()
