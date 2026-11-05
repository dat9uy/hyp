#!/usr/bin/env python3
"""Drive: run one trial's sessions in wt, and keep every one.

When the arm has a reader brief, one reader per distinct Result in items.tsv
runs first, one after another, under the verifier's actor name, each stopped
after 10 minutes or $1. A reader's failure does not void the trial: drive.json
counts it in readers_failed, and the verifier covers the gap by its brief.
Readers still waiting after 28 minutes are not started, and count as failed, so
one Drive command stays inside hyp's 60-minute limit.

The verifier is a claude -p session with brief.txt on stdin, project settings
only, the arm's hyp first on PATH, the model the inputs file names, and
HYP_ACTOR set to verifier-<arm>-<trial> after every host HYP_* variable is
dropped. It is stopped after 25 minutes or $3.

The follow-up resumes that session with followup.txt on stdin and no tools,
and is stopped after 5 minutes. It runs only when the verifier's turn ended
well. A resume may append to the session's file or write a new one, so the
follow-up's file is chosen from the session id its JSON result reports, and
only records the verifier's turn does not hold are kept.

Writes to this recipe's tmp, for remove's --keep: transcript.jsonl (the
verifier's turn), followup.jsonl (the follow-up's records, empty when it did
not run), reader-<R-id>.jsonl (each reader), verifier.json and followup.json
(each turn's stdout) and drive.json (exit, timed_out, subtype and cost of each
turn and reader, readers_failed, whether the follow-up was tool-free, and the
sum of all costs). Exits 1 when no verifier transcript was written.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import time
import uuid

VERIFIER_LIMIT = 25 * 60
FOLLOWUP_LIMIT = 5 * 60
READER_LIMIT = 10 * 60
READERS_WINDOW = 28 * 60
BUDGET = "3"
READER_BUDGET = "1"
TOOLS = ["--permission-mode", "dontAsk", "--allowedTools", "Bash", "Read", "Glob", "Grep"]


def state(path):
    out = {}
    with open(path) as f:
        for line in f:
            if "=" in line:
                k, v = line.rstrip("\n").split("=", 1)
                out[k] = v
    return out


def turn(argv, cwd, env, stdin, limit, keep):
    timed_out = False
    try:
        r = subprocess.run(argv, cwd=cwd, env=env, input=stdin, capture_output=True, text=True, timeout=limit)
        code, stdout = r.returncode, r.stdout
    except subprocess.TimeoutExpired as e:
        timed_out, code = True, None
        stdout = e.stdout.decode() if isinstance(e.stdout, bytes) else (e.stdout or "")
    if keep:
        with open(keep, "w") as f:
            f.write(stdout)
    try:
        out = json.loads(stdout)
        out = out if isinstance(out, dict) else {}
    except ValueError:
        out = {}
    return {"argv": argv, "exit": code, "timed_out": timed_out, "subtype": out.get("subtype"),
            "session_id": out.get("session_id"), "result": out.get("result"), "num_turns": out.get("num_turns"),
            "cost_usd": out.get("total_cost_usd"),
            "permission_denials": len(out.get("permission_denials") or [])}


def tool_uses(lines):
    n = 0
    for l in lines:
        try:
            d = json.loads(l)
        except ValueError:
            continue
        content = (d.get("message") or {}).get("content")
        if d.get("type") == "assistant" and isinstance(content, list):
            n += sum(1 for b in content if isinstance(b, dict) and b.get("type") == "tool_use")
    return n


def failed(t):
    return t["exit"] != 0 or t["timed_out"] or "budget" in (t["subtype"] or "")


def readers(rtmp, wt, env, model, actor, project):
    """Run one reader per distinct Result, in order, and return their rows."""
    path = os.path.join(rtmp, "reader-brief.txt")
    if not os.path.isfile(path):
        return []
    with open(path) as f:
        brief = f.read().replace("<actor>", actor)
    with open(os.path.join(rtmp, "items.tsv")) as f:
        results = list(dict.fromkeys(l.split("\t")[0] for l in f if l.strip()))
    rows, start = [], time.monotonic()
    for rid in results:
        if time.monotonic() - start > READERS_WINDOW:
            rows.append({"result": rid, "exit": None, "timed_out": False, "subtype": "not-started", "cost_usd": None, "failed": True})
            continue
        sid = str(uuid.uuid4())
        t = turn(["claude", "-p", "--model", model, "--session-id", sid, "--setting-sources", "project",
                  "--output-format", "json", *TOOLS, "--max-budget-usd", READER_BUDGET],
                 wt, env, brief.replace("<R-id>", rid), READER_LIMIT, None)
        src = os.path.join(project, sid + ".jsonl")
        if os.path.isfile(src):
            shutil.copyfile(src, os.path.join(rtmp, f"reader-{rid}.jsonl"))
        rows.append({"result": rid, "exit": t["exit"], "timed_out": t["timed_out"], "subtype": t["subtype"],
                     "cost_usd": t["cost_usd"], "failed": failed(t) or not os.path.isfile(src)})
    return rows


def main():
    wt = os.path.realpath(os.environ["HYP_SANDBOX_WT"])
    rtmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
    st = state(os.path.join(rtmp, "state.env"))
    with open(os.path.join(rtmp, "brief.txt")) as f:
        brief = f.read()
    with open("followup.txt") as f:
        followup = f.read()
    sid = str(uuid.uuid4())
    actor = f"verifier-{st['arm']}-{st['trial']}"
    env = {k: v for k, v in os.environ.items()
           if not (k.startswith("HYP_") or k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_"))}
    env["HYP_ACTOR"] = actor
    env["PATH"] = os.path.join(rtmp, "bin") + os.pathsep + env.get("PATH", "")
    project = os.path.join(os.path.expanduser("~/.claude/projects"), re.sub(r"[^A-Za-z0-9]", "-", wt))

    rows = readers(rtmp, wt, env, st["model"], actor, project)

    v = turn(["claude", "-p", "--model", st["model"], "--session-id", sid, "--setting-sources", "project",
              "--output-format", "json", *TOOLS, "--max-budget-usd", BUDGET],
             wt, env, brief, VERIFIER_LIMIT, os.path.join(rtmp, "verifier.json"))
    src = os.path.join(project, sid + ".jsonl")
    kept = os.path.isfile(src)
    first = []
    if kept:
        shutil.copyfile(src, os.path.join(rtmp, "transcript.jsonl"))
        with open(src) as f:
            first = f.readlines()

    later = []
    fu = {"exit": None, "timed_out": False, "subtype": None, "session_id": None, "result": None, "cost_usd": None}
    if kept and not failed(v):
        fu = turn(["claude", "-p", "--model", st["model"], "--resume", sid, "--setting-sources", "project",
                   "--tools", "", "--output-format", "json"],
                  wt, env, followup, FOLLOWUP_LIMIT, os.path.join(rtmp, "followup.json"))
        path = os.path.join(project, (fu["session_id"] or sid) + ".jsonl")
        if os.path.isfile(path):
            held = set(first)
            with open(path) as f:
                later = [l for l in f if l not in held]
    with open(os.path.join(rtmp, "followup.jsonl"), "w") as f:
        f.writelines(later)

    cost = sum(t["cost_usd"] or 0 for t in [v, fu, *rows])
    with open(os.path.join(rtmp, "drive.json"), "w") as f:
        json.dump({"arm": st["arm"], "trial": st["trial"], "model": st["model"], "session_id": sid,
                   "transcript_kept": kept, "readers": rows, "readers_failed": sum(1 for r in rows if r["failed"]),
                   "verifier": v, "followup": fu, "followup_tool_free": tool_uses(later) == 0,
                   "cost_usd": cost}, f, indent=1)
    print(f"drive: arm {st['arm']} trial {st['trial']} readers {len(rows)} verifier exit {v['exit']} followup exit {fu['exit']} transcript {kept} cost {cost:.2f}")
    if not kept:
        print(f"drive: no transcript at {src}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
