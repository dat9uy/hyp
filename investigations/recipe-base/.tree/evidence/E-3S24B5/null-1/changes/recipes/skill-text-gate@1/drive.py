#!/usr/bin/env python3
"""Drive: one fresh `claude -p` session in wt/, stopped after 50 minutes."""
import glob, json, os, shutil, signal, subprocess, sys, time, uuid

TIMEOUT = 50 * 60


def main():
    wt = os.environ["HYP_SANDBOX_WT"]
    tmp = os.environ["HYP_SANDBOX_RECIPE_TMP"]
    with open(os.path.join(tmp, "brief.txt"), "rb") as f:
        brief = f.read()
    sid = str(uuid.uuid4())
    env = {k: v for k, v in os.environ.items() if not k.startswith("HYP_SANDBOX_")}
    cmd = ["claude", "-p", "--model", "sonnet", "--setting-sources", "project",
           "--session-id", sid, "--output-format", "json",
           "--dangerously-skip-permissions"]
    started = time.time()
    try:
        p = subprocess.Popen(cmd, cwd=wt, env=env, stdin=subprocess.PIPE,
                             stdout=subprocess.PIPE, stderr=subprocess.PIPE,
                             start_new_session=True)
    except OSError as e:
        print("drive: cannot start claude: %s" % e, file=sys.stderr)
        return 1
    with open(os.path.join(tmp, "runner.pgid"), "w") as f:
        f.write(str(p.pid))
    timed_out = False
    try:
        out, err = p.communicate(brief, timeout=TIMEOUT)
    except subprocess.TimeoutExpired:
        timed_out = True
        os.killpg(p.pid, signal.SIGKILL)
        out, err = p.communicate()
    try:
        os.killpg(p.pid, signal.SIGKILL)  # anything the runner left behind
    except ProcessLookupError:
        pass
    os.remove(os.path.join(tmp, "runner.pgid"))
    result = {}
    try:
        r = json.loads(out)
        if isinstance(r, dict):
            result = r
    except ValueError:
        pass
    summary = {
        "session_id": sid,
        "exit": p.returncode,
        "timed_out": timed_out,
        "turns": result.get("num_turns"),
        "cost_usd": result.get("total_cost_usd"),
        "is_error": result.get("is_error"),
        "seconds": round(time.time() - started),
    }
    with open(os.path.join(tmp, "summary.json"), "w") as f:
        json.dump(summary, f, indent=1, sort_keys=True)
        f.write("\n")
    sys.stderr.write(err.decode("utf-8", "replace")[-4000:])
    found = glob.glob(os.path.join(os.path.expanduser("~"), ".claude", "projects", "*", sid + ".jsonl"))
    if not found:
        print("drive: no transcript for session %s" % sid, file=sys.stderr)
        return 1
    shutil.copyfile(found[0], os.path.join(tmp, "transcript.jsonl"))
    return 0


sys.exit(main())
