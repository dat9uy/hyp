#!/usr/bin/env python3
"""Drive: run the runner, a claude -p session, in wt, and keep its transcript.

The runner gets brief.txt on stdin, project settings only (the hyp skills
Launch put in wt/.claude/skills, no user settings, skills or plugins), the
hyp on PATH (the sandbox's pinned hyp), the model the inputs file names, and
HYP_ACTOR set to runner-<arm>-<trial> after every host HYP_* variable is
dropped, so its tree writes have an actor of their own.
It is stopped after 50 minutes, inside Drive's 60. The transcript and a
drive.json summary go to this recipe's tmp, for remove's --keep.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import uuid

LIMIT = 50 * 60


def state(path):
    out = {}
    with open(path) as f:
        for line in f:
            if "=" in line:
                k, v = line.rstrip("\n").split("=", 1)
                out[k] = v
    return out


def main():
    wt = os.path.realpath(os.environ["HYP_SANDBOX_WT"])
    rtmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
    st = state(os.path.join(rtmp, "state.env"))
    with open(os.path.join(rtmp, "brief.txt")) as f:
        brief = f.read()
    sid = str(uuid.uuid4())
    argv = ["claude", "-p", "--model", st["model"], "--session-id", sid, "--setting-sources", "project",
            "--output-format", "json", "--permission-mode", "dontAsk",
            "--allowedTools", "Bash", "Read", "Edit", "Write", "Glob", "Grep", "Skill"]
    env = {k: v for k, v in os.environ.items()
           if not (k.startswith("HYP_") or k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_"))}
    env["HYP_ACTOR"] = f"runner-{st['arm']}-{st['trial']}"
    timed_out = False
    try:
        r = subprocess.run(argv, cwd=wt, env=env, input=brief, capture_output=True, text=True, timeout=LIMIT)
        code, stdout = r.returncode, r.stdout
    except subprocess.TimeoutExpired as e:
        timed_out, code = True, None
        stdout = e.stdout.decode() if isinstance(e.stdout, bytes) else (e.stdout or "")
    try:
        out = json.loads(stdout)
        out = out if isinstance(out, dict) else {}
    except ValueError:
        out = {}
    src = os.path.join(os.path.expanduser("~/.claude/projects"), re.sub(r"[^A-Za-z0-9]", "-", wt), sid + ".jsonl")
    kept = os.path.isfile(src)
    if kept:
        shutil.copyfile(src, os.path.join(rtmp, "transcript.jsonl"))
    with open(os.path.join(rtmp, "drive.json"), "w") as f:
        json.dump({"arm": st["arm"], "trial": st["trial"], "model": st["model"], "session_id": sid,
                   "argv": argv, "exit": code, "timed_out": timed_out, "transcript_kept": kept,
                   "result": out.get("result"), "num_turns": out.get("num_turns"),
                   "cost_usd": out.get("total_cost_usd"),
                   "permission_denials": len(out.get("permission_denials") or [])}, f, indent=1)
    print(f"drive: arm {st['arm']} trial {st['trial']} exit {code} timed_out {timed_out} transcript {kept}")
    if not kept:
        print(f"drive: no transcript at {src}", file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
