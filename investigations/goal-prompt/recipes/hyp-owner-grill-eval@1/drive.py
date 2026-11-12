#!/usr/bin/env python3
"""Drive: run one grill trial, the grill and the stand-in owner taking turns.

The grill is a claude -p session in wt that starts from prompt.txt (the
/hyp-owner-grill command) and is resumed with each stand-in reply. The
stand-in is a second claude -p session, with no tools, in its own empty
directory, whose system prompt is standin-system.txt; it is resumed with
each grill message. The loop ends when the stand-in replies <<END>>, when a
call fails, after max_exchanges grill calls, or at limit_min minutes.

Every call runs on the inputs file's model, with every model alias and the
subagent model mapped to it. The grill gets project settings only,
Claude Code's sandbox (Bash writes only inside wt and reaches no network),
Edit and Write inside wt only, the hyp first on PATH (the sandbox's pinned
hyp), no HYP_*, CLAUDE_CODE_* or gh token variables, and a git config with
no credential helper and dead push URLs.

It writes to this recipe's tmp: calls.json (one row per call), drive.json,
transcript.jsonl (the grill's session), subagents.tar (its subagents'
transcripts), standin.jsonl (the stand-in's session) and stores.tar (the events
of every .tree directory under wt at the end). It exits 1 when
the grill wrote no transcript.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import time
import uuid

TOOLS = ["Bash", "Read", "Edit(./**)", "Write(./**)", "Glob", "Grep", "Skill", "Agent", "Task", "TodoWrite"]
END = "<<END>>"


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
MODEL = ST["model"]


def base_env():
    env = {k: v for k, v in os.environ.items()
           if not (k.startswith("HYP_") or k == "CLAUDECODE" or k.startswith("CLAUDE_CODE_")
                   or k.startswith("ANTHROPIC_DEFAULT_")
                   or k in ("GH_TOKEN", "GITHUB_TOKEN", "GH_ENTERPRISE_TOKEN"))}
    env.update({"CLAUDE_CODE_SUBAGENT_MODEL": MODEL, "ANTHROPIC_DEFAULT_OPUS_MODEL": MODEL,
                "ANTHROPIC_DEFAULT_SONNET_MODEL": MODEL, "ANTHROPIC_DEFAULT_HAIKU_MODEL": MODEL,
                "ANTHROPIC_DEFAULT_FABLE_MODEL": MODEL, "DISABLE_AUTOUPDATER": "1"})
    return env


def grill_env():
    env = base_env()
    env.update({"GIT_CONFIG_GLOBAL": os.path.join(RT, "gitconfig"), "GIT_CONFIG_NOSYSTEM": "1",
                "GH_CONFIG_DIR": os.path.join(RT, "gh"), "GIT_TERMINAL_PROMPT": "0",
                "GIT_SSH_COMMAND": "false", "GOCACHE": os.path.join(WT, ".git", "gocache")})
    return env


def call(who, argv, cwd, env, text, deadline):
    """Runs one claude -p call with text on stdin; returns its result record."""
    t0 = time.time()
    p = subprocess.Popen(argv, cwd=cwd, env=env, stdin=subprocess.PIPE, stdout=subprocess.PIPE,
                         stderr=open(os.path.join(RT, "claude.err"), "a"), start_new_session=True, text=True)
    with open(os.path.join(RT, "run.json"), "w") as f:
        json.dump({"pid": p.pid, "who": who}, f)
    timed_out = False
    try:
        out, _ = p.communicate(text, timeout=max(1, deadline - time.time()))
    except subprocess.TimeoutExpired:
        timed_out = True
        for sig in (15, 9):
            try:
                os.killpg(p.pid, sig)
            except OSError:
                pass
            try:
                p.wait(timeout=30)
                break
            except subprocess.TimeoutExpired:
                pass
        out = ""
    try:
        rec = json.loads(out) if out.strip() else {}
        rec = rec if isinstance(rec, dict) else {}
    except ValueError:
        rec = {}
    row = {"who": who, "exit": p.returncode, "timed_out": timed_out, "seconds": round(time.time() - t0, 1),
           "session_id": rec.get("session_id"), "total_cost_usd": rec.get("total_cost_usd"),
           "num_turns": rec.get("num_turns"), "is_error": rec.get("is_error"),
           "models": sorted((rec.get("modelUsage") or {}).keys()),
           "permission_denials": len(rec.get("permission_denials") or [])}
    return row, (rec.get("result") or "")


def main():
    start = time.time()
    deadline = start + int(ST["limit_min"]) * 60
    gsid, ssid = str(uuid.uuid4()), str(uuid.uuid4())
    version = subprocess.run(["claude", "--version"], capture_output=True, text=True).stdout.strip()
    gargs = ["claude", "-p", "--model", MODEL, "--setting-sources", "project", "--settings",
             os.path.join(RT, "claude-settings.json"), "--output-format", "json",
             "--permission-mode", "dontAsk", "--allowedTools"] + TOOLS
    sargs = ["claude", "-p", "--model", MODEL, "--setting-sources", "project", "--tools", "",
             "--output-format", "json"]
    system = open(os.path.join(RT, "standin-system.txt")).read()
    calls, end = [], "max_exchanges"
    message = open(os.path.join(RT, "prompt.txt")).read()
    for n in range(int(ST["max_exchanges"])):
        argv = gargs + (["--session-id", gsid] if n == 0 else ["--resume", gsid])
        row, said = call("grill", argv, WT, grill_env(), message, deadline)
        calls.append(row)
        if row["timed_out"]:
            end = "limit_min"
            break
        if row["exit"] != 0 or row["is_error"] or not said.strip():
            end = "grill_error"
            break
        ask = "The grill's message:\n\n" + said
        sv = sargs + ["--system-prompt", system] + (["--session-id", ssid] if n == 0 else ["--resume", ssid])
        row, reply = call("standin", sv, os.path.join(RT, "standin"), base_env(), ask, deadline)
        calls.append(row)
        if row["timed_out"]:
            end = "limit_min"
            break
        if row["exit"] != 0 or row["is_error"] or not reply.strip():
            end = "standin_error"
            break
        if reply.strip() == END:
            end = "end"
            break
        message = reply
    collect(start, gsid, ssid, calls, end, version)


def session_cost(calls, who):
    """A resumed session reports its whole spend so far: its cost is its last call's."""
    costs = [c["total_cost_usd"] for c in calls if c["who"] == who and c["total_cost_usd"] is not None]
    return costs[-1] if costs else None


def collect(start, gsid, ssid, calls, end, version):
    with open(os.path.join(RT, "calls.json"), "w") as f:
        json.dump(calls, f, indent=1)
    projects = os.path.expanduser("~/.claude/projects")
    gdir = os.path.join(projects, re.sub(r"[^A-Za-z0-9]", "-", WT))
    sdir = os.path.join(projects, re.sub(r"[^A-Za-z0-9]", "-", os.path.realpath(os.path.join(RT, "standin"))))
    kept = os.path.isfile(os.path.join(gdir, gsid + ".jsonl"))
    if kept:
        shutil.copyfile(os.path.join(gdir, gsid + ".jsonl"), os.path.join(RT, "transcript.jsonl"))
    if os.path.isfile(os.path.join(sdir, ssid + ".jsonl")):
        shutil.copyfile(os.path.join(sdir, ssid + ".jsonl"), os.path.join(RT, "standin.jsonl"))
    else:
        open(os.path.join(RT, "standin.jsonl"), "w").close()
    subs = 0
    with tarfile.open(os.path.join(RT, "subagents.tar"), "w") as t:
        sd = os.path.join(gdir, gsid, "subagents")
        if os.path.isdir(sd):
            for name in sorted(os.listdir(sd)):
                t.add(os.path.join(sd, name), arcname="subagents/" + name)
                subs += name.endswith(".jsonl")
    with tarfile.open(os.path.join(RT, "stores.tar"), "w") as t:
        for root, dirs, _ in os.walk(WT):
            dirs[:] = sorted(d for d in dirs if d not in (".git", ".claude", "node_modules"))
            if ".tree" in dirs:
                ev = os.path.join(root, ".tree", "events")
                if os.path.isdir(ev):
                    t.add(ev, arcname=os.path.relpath(ev, WT))
                dirs.remove(".tree")
    home = os.path.expanduser("~/.local/bin/hyp")
    after = subprocess.run(["shasum", "-a", "256", home], capture_output=True, text=True).stdout.split()[:1] if os.path.isfile(home) else ["none"]
    before = open(os.path.join(RT, "home-hyp.before")).read().strip()
    g, s = session_cost(calls, "grill"), session_cost(calls, "standin")
    literal = [c["total_cost_usd"] for c in calls if c["total_cost_usd"] is not None]
    timed_out = end == "limit_min"
    with open(os.path.join(RT, "drive.json"), "w") as f:
        json.dump({"arm": ST["arm"], "fixture": ST["fixture"], "trial": ST["trial"], "model": MODEL,
                   "claude_version": version, "start": round(start, 3), "end": end, "timed_out": timed_out,
                   "grill_session": gsid, "standin_session": ssid, "exchanges": sum(c["who"] == "grill" for c in calls),
                   "transcript_kept": kept, "subagents": subs,
                   "cost_usd": None if timed_out or g is None else round(g + (s or 0), 6),
                   "cost_usd_literal": None if timed_out or not literal else round(sum(literal), 6),
                   "models": sorted({m for c in calls for m in c["models"]}),
                   "home_hyp_unchanged": (after[0] if after else "none") == before}, f, indent=1)
    print("drive: arm %s fixture %s trial %s end %s exchanges %d subagents %d" % (
        ST["arm"], ST["fixture"], ST["trial"], end, sum(c["who"] == "grill" for c in calls), subs))
    if not kept:
        print("drive: no grill transcript in " + gdir, file=sys.stderr)
        sys.exit(1)


if __name__ == "__main__":
    main()
