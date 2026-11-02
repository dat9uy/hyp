#!/usr/bin/env python3
"""trial.py REPO COMMIT STORE CUT ARM VARIANT TRIAL BRIEF HYP CLAUDE [TIMEOUT]

Run one trial of the skill-text gate; see recipe.md."""
import json
import os
import re
import shutil
import signal
import subprocess
import sys

ARM = re.compile(r"^[a-z][a-z0-9-]{0,31}$")
SKILLS = ("hyp-prototype", "hyp-verification")


def usage(msg):
    print("trial.py: " + msg, file=sys.stderr)
    sys.exit(2)


def fail(msg):
    print("trial.py: " + msg, file=sys.stderr)
    sys.exit(1)


def run(args, **kw):
    return subprocess.run(args, capture_output=True, text=True, **kw)


def hyp_events(hyp, wt):
    p = run([hyp, "events", "--json"], cwd=wt)
    if p.returncode != 0:
        fail("hyp events --json: " + p.stderr.strip())
    return json.loads(p.stdout)["events"]


def rm(path):
    if os.path.isdir(path) and not os.path.islink(path):
        shutil.rmtree(path)
    else:
        os.remove(path)


def main():
    a = sys.argv[1:]
    if len(a) not in (10, 11):
        usage("want REPO COMMIT STORE CUT ARM VARIANT TRIAL BRIEF HYP CLAUDE [TIMEOUT]")
    repo, commit, store, cut, arm, variant, trial, brief, hyp, claude = a[:10]
    timeout = 3000
    if len(a) == 11:
        if not a[10].isdigit() or int(a[10]) < 1:
            usage("TIMEOUT must be a positive number of seconds")
        timeout = int(a[10])
    if not ARM.match(arm) or arm == "repro":
        usage("bad ARM %r" % arm)
    if not trial.isdigit() or int(trial) < 1:
        usage("TRIAL must be a positive integer")
    for p in (repo, store, brief):
        if not os.path.isabs(p) or not os.path.exists(p):
            usage("%s: want an existing absolute path" % p)
    hyp, claude = shutil.which(hyp), shutil.which(claude)
    if not hyp or not claude:
        usage("HYP and CLAUDE must be executables")
    hyp, claude = os.path.abspath(hyp), os.path.abspath(claude)
    laid = []
    if arm == "base":
        if variant != "none":
            usage("the arm base takes VARIANT none")
    else:
        if not os.path.isabs(variant) or not os.path.isdir(variant):
            usage("VARIANT must be an absolute directory")
        for e in sorted(os.listdir(variant)):
            d = os.path.join(variant, e)
            if e not in SKILLS or os.path.islink(d) or not os.path.isdir(d) \
                    or os.listdir(d) != ["SKILL.md"] or not os.path.isfile(os.path.join(d, "SKILL.md")):
                usage("VARIANT holds %r; want only <skill>/SKILL.md for %s" % (e, " or ".join(SKILLS)))
            laid.append(e)
        if not laid:
            usage("VARIANT is empty")

    out = os.environ["HYP_OUT"]
    tmp = os.environ.get("TMPDIR", "/tmp")
    sid = "sg-%s-%s" % (arm, trial)
    removed = os.path.join(tmp, "removed")
    p = run([hyp, "sandbox", "create", sid, "--host", repo, "--base", commit,
             "--recipe", "none", "--store", store, "--cut", cut, "--hyp", hyp])
    if p.returncode != 0:
        fail("sandbox create: " + (p.stderr or p.stdout).strip())
    wt = os.path.join(os.path.expanduser("~"), ".hyp-sandbox", sid, "wt")
    bindir = os.path.join(os.path.dirname(wt), "bin")

    before = {e["id"] for e in hyp_events(hyp, wt)}

    setup = []
    for rel in ("AGENTS.md", "CLAUDE.md", ".agentkit", "plans", os.path.join("docs", "plans")):
        path = os.path.join(wt, rel)
        if os.path.lexists(path):
            rm(path)
            setup.append("deleted " + rel)
    claude_dir = os.path.join(wt, ".claude")
    if os.path.lexists(claude_dir):
        for e in sorted(os.listdir(claude_dir)):
            if e != "skills":
                rm(os.path.join(claude_dir, e))
                setup.append("deleted .claude/" + e)
    cs = os.path.join(claude_dir, "skills")
    if os.path.lexists(cs):
        rm(cs)
        setup.append("deleted .claude/skills")
    for s in SKILLS:
        src = os.path.join(wt, "skills", s)
        if os.path.isdir(src):
            shutil.copytree(src, os.path.join(cs, s))
            setup.append("laid .claude/skills/" + s)
    for s in laid:
        for dst in (os.path.join(wt, "skills", s), os.path.join(cs, s)):
            os.makedirs(dst, exist_ok=True)
            shutil.copyfile(os.path.join(variant, s, "SKILL.md"), os.path.join(dst, "SKILL.md"))
        setup.append("laid variant skills/%s/SKILL.md and .claude/skills/%s/SKILL.md" % (s, s))

    env = dict(os.environ)
    env["PATH"] = bindir + os.pathsep + env.get("PATH", "")
    env.setdefault("HYP_ACTOR", "runner")
    transcript = os.path.join(out, "transcript.jsonl")
    timed_out = False
    with open(brief, "rb") as bf, open(transcript, "wb") as tf:
        proc = subprocess.Popen(
            [claude, "-p", "--model", "sonnet", "--setting-sources", "project",
             "--permission-mode", "bypassPermissions",
             "--output-format", "stream-json", "--verbose"],
            cwd=wt, env=env, stdin=bf, stdout=tf, stderr=subprocess.DEVNULL,
            start_new_session=True)
        try:
            code = proc.wait(timeout=timeout)
        except subprocess.TimeoutExpired:
            timed_out = True
            os.killpg(proc.pid, signal.SIGKILL)
            code = proc.wait()
        else:
            try:
                os.killpg(proc.pid, signal.SIGKILL)  # strays of the runner
            except OSError:
                pass

    turns = cost = None
    with open(transcript, encoding="utf-8", errors="replace") as tf:
        for line in tf:
            try:
                ev = json.loads(line)
            except ValueError:
                continue
            if isinstance(ev, dict) and ev.get("type") == "result":
                turns, cost = ev.get("num_turns"), ev.get("total_cost_usd")

    new = []
    for e in hyp_events(hyp, wt):
        if e["id"] not in before:
            new.append({"type": e["type"], "kind": (e.get("payload") or {}).get("kind", "")})

    p = run([hyp, "sandbox", "remove", sid, removed])
    if p.returncode != 0:
        fail("sandbox remove: " + (p.stderr or p.stdout).strip())

    def put(name, v):
        with open(os.path.join(out, name), "w") as f:
            json.dump(v, f, indent=1, sort_keys=True)
            f.write("\n")

    put("meta.json", {"arm": arm, "trial": int(trial), "commit": commit, "cut": cut})
    put("setup.json", setup)
    put("summary.json", {"exit": code, "timed_out": timed_out, "turns": turns, "cost": cost})
    put("new-events.json", new)
    shutil.copyfile(os.path.join(removed, "changes.json"), os.path.join(out, "changes.json"))
    if os.path.isdir(os.path.join(removed, "changes")):
        shutil.copytree(os.path.join(removed, "changes"), os.path.join(out, "changes"))


main()
