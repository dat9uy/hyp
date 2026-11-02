#!/usr/bin/env -S python3 -I
"""drive.py: run the trial's runner, a fresh `claude -p` session, once.

Keeps, under $HYP_SANDBOX_RECIPE_TMP: run.json (the summary) and
transcript.jsonl (the session transcript). Exits 1 only when it cannot run
the runner or find its transcript; a runner that fails or times out is data.
"""
import glob
import json
import os
import shutil
import signal
import subprocess
import sys
import uuid

TIMEOUT = 50 * 60


def die(msg):
    sys.stderr.write("drive: %s\n" % msg)
    sys.exit(1)


def main():
    wt = os.environ.get("HYP_SANDBOX_WT")
    tmp = os.environ.get("HYP_SANDBOX_RECIPE_TMP")
    if not wt or not tmp:
        die("not run by hyp sandbox drive")
    if os.path.exists(os.path.join(tmp, "run.json")):
        die("this trial was driven already; a trial is one run in a fresh sandbox")
    with open(os.path.join(tmp, "trial.json"), encoding="utf-8") as f:
        trial = json.load(f)
    sid = str(uuid.uuid4())
    env = dict(os.environ)
    for k in list(env):
        if k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_"):
            del env[k]
    env["HYP_ACTOR"] = "runner-%s-%d" % (trial["arm"], trial["trial"])
    argv = ["claude", "-p", "--model", "sonnet", "--setting-sources", "project",
            "--session-id", sid, "--output-format", "json",
            "--permission-mode", "bypassPermissions"]
    timed_out = False
    try:
        with open(os.path.join(tmp, "brief.md"), "rb") as brief:
            p = subprocess.Popen(argv, cwd=wt, env=env, stdin=brief, stdout=subprocess.PIPE,
                                 stderr=subprocess.PIPE, start_new_session=True)
    except OSError as e:
        die("cannot start claude: %s" % e)
    with open(os.path.join(tmp, "pgid"), "w") as f:
        f.write("%d\n" % p.pid)
    try:
        out, err = p.communicate(timeout=TIMEOUT)
    except subprocess.TimeoutExpired:
        timed_out = True
        os.killpg(p.pid, signal.SIGKILL)
        out, err = p.communicate()
    try:
        os.killpg(p.pid, signal.SIGKILL)   # anything the runner left behind
    except OSError:
        pass
    result = None
    try:
        result = json.loads(out.decode("utf-8", "replace"))
    except ValueError:
        pass
    r = result if isinstance(result, dict) else {}
    summary = {
        "session_id": sid,
        "exit": None if timed_out else p.returncode,
        "timed_out": timed_out,
        "turns": r.get("num_turns"),
        "cost_usd": r.get("total_cost_usd"),
        "duration_ms": r.get("duration_ms"),
        "is_error": r.get("is_error"),
        "stderr_tail": err.decode("utf-8", "replace")[-2000:],
    }
    with open(os.path.join(tmp, "run.json"), "w", encoding="utf-8") as f:
        json.dump(summary, f, indent=1, sort_keys=True)
        f.write("\n")
    found = glob.glob(os.path.join(os.path.expanduser("~"), ".claude", "projects", "*", sid + ".jsonl"))
    if len(found) != 1:
        die("session transcript %s.jsonl not found under ~/.claude/projects" % sid)
    shutil.copyfile(found[0], os.path.join(tmp, "transcript.jsonl"))
    print("drive: arm %s trial %d exit %s turns %s" % (trial["arm"], trial["trial"], summary["exit"], summary["turns"]))


main()
