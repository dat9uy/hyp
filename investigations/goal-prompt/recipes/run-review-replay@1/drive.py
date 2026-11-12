#!/usr/bin/env python3
"""Drive: run the replayed /goal run, a claude -p session, in wt.

drive.py start: starts the runner in its own session, detached, with the
prompt Launch printed on stdin, and waits for it up to 55 minutes, or
limit_min when that is shorter.
drive.py finish: waits for it until limit_min minutes after its start,
stops it then (SIGTERM, then SIGKILL after 30 seconds), and collects the
trial into this recipe's tmp: drive.json, transcript.jsonl, subagents.tar,
tree.tar (the store at its end) and commits.json. It exits 1 when no
transcript was written.

The runner gets project settings only, the hyp first on PATH (the
sandbox's pinned arm), no HYP_* or CLAUDE_CODE_* variables, no gh token, a
git config with no credential helper and dead push URLs, and the model the
inputs file names.
"""
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import tarfile
import time
import uuid

TOOLS = ["Bash", "Read", "Edit", "Write", "Glob", "Grep", "Skill", "Agent", "Task"]


def state(path):
    out = {}
    with open(path) as f:
        for line in f:
            if "=" in line:
                k, v = line.rstrip("\n").split("=", 1)
                out[k] = v
    return out


RT = os.environ["HYP_SANDBOX_RECIPE_TMP"]
WT = os.path.realpath(os.environ["HYP_SANDBOX_WT"])
ST = state(os.path.join(RT, "state.txt"))
RUN = os.path.join(RT, "run.json")


def alive(pid):
    try:
        os.kill(pid, 0)
    except OSError:
        return False
    r = subprocess.run(["ps", "-o", "stat=", "-p", str(pid)], capture_output=True, text=True)
    return r.stdout.strip() != "" and not r.stdout.strip().startswith("Z")


def wait(pid, until):
    while alive(pid) and time.time() < until:
        time.sleep(5)
    return not alive(pid)


def start():
    sid = str(uuid.uuid4())
    env = {k: v for k, v in os.environ.items()
           if not (k.startswith("HYP_") or k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_")
                   or k in ("GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN"))}
    env.update({"GIT_CONFIG_GLOBAL": os.path.join(RT, "gitconfig"), "GIT_CONFIG_NOSYSTEM": "1",
                "GH_CONFIG_DIR": os.path.join(RT, "gh"), "GIT_TERMINAL_PROMPT": "0",
                "GIT_SSH_COMMAND": "false", "DISABLE_AUTOUPDATER": "1"})
    version = subprocess.run(["claude", "--version"], env=env, capture_output=True, text=True).stdout.strip()
    argv = ["claude", "-p", "--model", ST["model"], "--session-id", sid, "--setting-sources", "project",
            "--output-format", "json", "--permission-mode", "dontAsk", "--allowedTools"] + TOOLS
    p = subprocess.Popen(argv, cwd=WT, env=env, stdin=open(os.path.join(RT, "prompt.txt")),
                         stdout=open(os.path.join(RT, "claude.out"), "w"),
                         stderr=open(os.path.join(RT, "claude.err"), "w"), start_new_session=True)
    t0 = time.time()
    run = {"pid": p.pid, "session_id": sid, "start": t0, "argv": argv, "claude_version": version}
    with open(RUN, "w") as f:
        json.dump(run, f)
    try:
        run["exit"] = p.wait(timeout=min(55, int(ST["limit_min"])) * 60)
        with open(RUN, "w") as f:
            json.dump(run, f)
    except subprocess.TimeoutExpired:
        pass
    print("drive start: session %s pid %d %s" % (sid, p.pid, "ended" if p.poll() is not None else "still running"))


def commits():
    out = {}
    shas = subprocess.run(["git", "rev-list", "--all", "HEAD"], cwd=WT, capture_output=True, text=True).stdout.split()
    for sha in dict.fromkeys(shas):
        parents = subprocess.run(["git", "rev-list", "--parents", "-n", "1", sha], cwd=WT,
                                 capture_output=True, text=True).stdout.split()[1:]
        files = []
        if parents:
            d = subprocess.run(["git", "diff-tree", "-r", "--no-commit-id", "--no-renames", parents[0], sha],
                               cwd=WT, capture_output=True, text=True).stdout
            for line in d.splitlines():
                meta, path = line.split("\t", 1)
                f = meta.split()
                files.append({"path": path, "status": f[4][0], "blob": f[3]})
        out[sha] = {"parents": parents, "files": files}
    return {"head": subprocess.run(["git", "rev-parse", "HEAD"], cwd=WT, capture_output=True, text=True).stdout.strip(),
            "commits": out}


def finish():
    run = json.load(open(RUN))
    pid, sid = run["pid"], run["session_id"]
    timed_out = not wait(pid, run["start"] + int(ST["limit_min"]) * 60)
    if timed_out:
        for sig, grace in ((signal.SIGTERM, 30), (signal.SIGKILL, 10)):
            try:
                os.killpg(pid, sig)
            except OSError:
                pass
            if wait(pid, time.time() + grace):
                break
    try:
        out = json.load(open(os.path.join(RT, "claude.out")))
        out = out if isinstance(out, dict) else {}
    except ValueError:
        out = {}
    proj = os.path.join(os.path.expanduser("~/.claude/projects"), re.sub(r"[^A-Za-z0-9]", "-", WT))
    src = os.path.join(proj, sid + ".jsonl")
    kept = os.path.isfile(src)
    if kept:
        shutil.copyfile(src, os.path.join(RT, "transcript.jsonl"))
    subs = 0
    with tarfile.open(os.path.join(RT, "subagents.tar"), "w") as t:
        sd = os.path.join(proj, sid, "subagents")
        if os.path.isdir(sd):
            for f in sorted(os.listdir(sd)):
                t.add(os.path.join(sd, f), arcname="subagents/" + f)
                subs += f.endswith(".jsonl")
    store = os.path.join(WT, ST["store"])
    with tarfile.open(os.path.join(RT, "tree.tar"), "w") as t:
        t.add(os.path.join(store, ".tree"), arcname=".tree",
              filter=lambda ti: None if os.path.basename(ti.name).startswith("index.sqlite") else ti)
    with open(os.path.join(RT, "commits.json"), "w") as f:
        json.dump(commits(), f)
    home = os.path.expanduser("~/.local/bin/hyp")
    after = subprocess.run(["shasum", "-a", "256", home], capture_output=True, text=True).stdout.split()[:1] if os.path.isfile(home) else ["none"]
    before = open(os.path.join(RT, "home-hyp.before")).read().strip()
    with open(os.path.join(RT, "drive.json"), "w") as f:
        json.dump({"arm": ST["arm"], "cut": ST["cut"], "trial": ST["trial"], "model": ST["model"],
                   "actor": ST["actor"], "session_id": sid, "argv": run["argv"], "exit": run.get("exit"),
                   "claude_version": run.get("claude_version"),
                   "timed_out": timed_out, "transcript_kept": kept, "subagents": subs,
                   "result": (out.get("result") or "")[-4000:], "num_turns": out.get("num_turns"),
                   "cost_usd": out.get("total_cost_usd"), "duration_ms": out.get("duration_ms"),
                   "is_error": out.get("is_error"),
                   "permission_denials": len(out.get("permission_denials") or []),
                   "home_hyp_unchanged": (after[0] if after else "none") == before}, f, indent=1)
    print("drive finish: arm %s cut %s trial %s timed_out %s transcript %s subagents %d cost %s" % (
        ST["arm"], ST["cut"], ST["trial"], timed_out, kept, subs, out.get("total_cost_usd")))
    if not kept:
        print("drive: no transcript at " + src, file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    {"start": start, "finish": finish}[sys.argv[1]]()
