#!/usr/bin/env python3
"""Drive: run one trial's sessions in wt, and keep every one.

When the arm has a reader brief, one reader per distinct Result in items.tsv runs
first, one after another, with HYP_ACTOR reader-<arm>-<trial>-<R-id>, each stopped
after 5 minutes or $0.25. A reader that could end after minute 30 is not started,
and counts as failed. A reader answered when its final message has, for every
verdict of that Result in items.tsv, a line holding the R-id, then the H-id, then
holds, wrong or unclear. Each final message is saved unchanged as
reader-report-<R-id>.md in this recipe's tmp. Once every reader has ended they
are also written to wt/reader-reports/<R-id>.md, so no reader reads another's.

A reader that failed or did not answer ends the trial's sessions there: no later
reader, verifier or follow-up runs, and the trial is void by score. Drive still
exits 0, with empty transcript.jsonl, followup.jsonl, verifier.json and
followup.json, so that remove's --keep list does not depend on what ran.

The verifier is a claude -p session with brief.txt on stdin, project settings
only, the hyp on PATH (the sandbox's pinned hyp), the model the inputs file names,
and HYP_ACTOR set to verifier-<arm>-<trial> after every host HYP_* variable is
dropped. It is stopped after 20 minutes or $1.

The follow-up resumes that session with followup.txt on stdin and no tools, and
is stopped after 5 minutes or $0.50. It runs only when the verifier's turn ended well. A
resume may append to the session's file or write a new one, so the follow-up's
file is chosen from the session id its JSON result reports, and only records the
verifier's turn does not hold are kept.

Writes to this recipe's tmp, for remove's --keep: transcript.jsonl (the verifier's
turn), followup.jsonl (the follow-up's records), reader-<R-id>.jsonl and
reader-report-<R-id>.md (each reader that ran), verifier.json and followup.json
(each turn's stdout) and drive.json (exit, timed_out, subtype and cost of each
turn; each reader's exit, wall seconds, cost, failed and answered; readers_failed;
readers_no_answer; whether the follow-up was tool-free; and the sum of all costs).
Exits 1 when the verifier ran and no transcript of it was written.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import time
import uuid

VERIFIER_LIMIT = 20 * 60
FOLLOWUP_LIMIT = 5 * 60
READER_LIMIT = 5 * 60
READERS_WINDOW = 30 * 60
BUDGET = "1"
READER_BUDGET = "0.25"
FOLLOWUP_BUDGET = "0.50"
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


def answered(text, rid, hids):
    for hid in hids:
        pattern = re.escape(rid) + r"\W+" + re.escape(hid) + r"\b[^\n]*?\b(holds|wrong|unclear)\b"
        if not re.search(pattern, text or ""):
            return False
    return True


def readers(rtmp, wt, env, model, arm, trial, project):
    """Run one reader per distinct Result, in order, stopping at the first without
    an answer. Writes wt/reader-reports only after the last one has ended."""
    path = os.path.join(rtmp, "reader-brief.txt")
    if not os.path.isfile(path):
        return []
    with open(path) as f:
        brief = f.read()
    verdicts = {}
    with open(os.path.join(rtmp, "items.tsv")) as f:
        for line in f:
            if line.strip():
                rid, hid = line.rstrip("\n").split("\t")
                verdicts.setdefault(rid, []).append(hid)
    rows, reports, start = [], {}, time.monotonic()
    for rid, hids in verdicts.items():
        if time.monotonic() - start + READER_LIMIT > READERS_WINDOW:
            rows.append({"result": rid, "exit": None, "timed_out": False, "subtype": "not-started", "cost_usd": None,
                         "wall_seconds": 0, "failed": True, "answered": False})
            break
        sid = str(uuid.uuid4())
        began = time.monotonic()
        reader_env = dict(env, HYP_ACTOR=f"reader-{arm}-{trial}-{rid}")
        t = turn(["claude", "-p", "--model", model, "--session-id", sid, "--setting-sources", "project",
                  "--output-format", "json", *TOOLS, "--max-budget-usd", READER_BUDGET],
                 wt, reader_env, brief.replace("<R-id>", rid), READER_LIMIT, None)
        src = os.path.join(project, sid + ".jsonl")
        if os.path.isfile(src):
            shutil.copyfile(src, os.path.join(rtmp, f"reader-{rid}.jsonl"))
        text = t["result"] if isinstance(t["result"], str) else ""
        with open(os.path.join(rtmp, f"reader-report-{rid}.md"), "w") as f:
            f.write(text)
        reports[rid] = text
        row = {"result": rid, "exit": t["exit"], "timed_out": t["timed_out"], "subtype": t["subtype"],
               "cost_usd": t["cost_usd"], "wall_seconds": round(time.monotonic() - began, 1),
               "failed": failed(t) or not os.path.isfile(src), "answered": answered(text, rid, hids)}
        rows.append(row)
        if row["failed"] or not row["answered"]:
            break
    folder = os.path.join(wt, "reader-reports")
    os.makedirs(folder, exist_ok=True)
    for rid, text in reports.items():
        with open(os.path.join(folder, rid + ".md"), "w") as f:
            f.write(text)
    return rows


def main():
    wt = os.path.realpath(os.environ["HYP_SANDBOX_WT"])
    rtmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
    st = state(os.path.join(rtmp, "state.txt"))
    with open(os.path.join(rtmp, "brief.txt")) as f:
        brief = f.read()
    with open("followup.txt") as f:
        followup = f.read()
    sid = str(uuid.uuid4())
    env = {k: v for k, v in os.environ.items()
           if not (k.startswith("HYP_") or k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_"))}
    env["HYP_ACTOR"] = f"verifier-{st['arm']}-{st['trial']}"
    project = os.path.join(os.path.expanduser("~/.claude/projects"), re.sub(r"[^A-Za-z0-9]", "-", wt))

    rows = readers(rtmp, wt, env, st["model"], st["arm"], st["trial"], project)
    readers_failed = sum(1 for r in rows if r["failed"])
    readers_no_answer = sum(1 for r in rows if not r["failed"] and not r["answered"])

    not_run = {"exit": None, "timed_out": False, "subtype": None, "session_id": None, "result": None, "cost_usd": None}
    v, fu = dict(not_run), dict(not_run)
    kept = False
    later = []
    skipped = readers_failed + readers_no_answer > 0
    for name in ("transcript.jsonl", "verifier.json", "followup.json"):
        open(os.path.join(rtmp, name), "w").close()
    if not skipped:
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
        if kept and not failed(v):
            fu = turn(["claude", "-p", "--model", st["model"], "--resume", sid, "--setting-sources", "project",
                       "--tools", "", "--output-format", "json", "--max-budget-usd", FOLLOWUP_BUDGET],
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
                   "transcript_kept": kept, "readers": rows, "readers_failed": readers_failed,
                   "readers_no_answer": readers_no_answer, "verifier": v, "followup": fu,
                   "followup_tool_free": tool_uses(later) == 0, "cost_usd": cost}, f, indent=1)
    print(f"drive: arm {st['arm']} trial {st['trial']} readers {len(rows)} failed {readers_failed} no-answer {readers_no_answer} "
          f"verifier exit {v['exit']} followup exit {fu['exit']} transcript {kept} cost {cost:.2f}")
    if not skipped and not kept:
        print(f"drive: no transcript at {src}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
