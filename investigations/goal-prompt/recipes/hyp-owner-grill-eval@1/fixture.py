#!/usr/bin/env python3
"""Makes the known-answer inputs of this bundle with hyp and git; not a
command. Run once from the bundle directory: ./fixture.py. It writes
known-answers/*/input/ and leaves each expect.json to be written by hand.
"""
import io
import json
import os
import shutil
import subprocess
import tarfile
import tempfile
import time

HERE = os.path.dirname(os.path.abspath(__file__))
KA = os.path.join(HERE, "known-answers")
SHA = "d81f883a9ddbb87fa2b015781e42041811cb14750dd3a703cfaeca409c1d7247"


def hyp(cwd, *args, actor="g"):
    env = dict(os.environ, HYP_ACTOR=actor)
    r = subprocess.run(["hyp"] + list(args), cwd=cwd, env=env, capture_output=True, text=True)
    if r.returncode != 0:
        raise SystemExit("hyp %s: %s" % (args, r.stderr))
    return r.stdout.split()[0] if r.stdout.split() else ""


def tar_dir(path, src, arc):
    with tarfile.open(path, "w") as t:
        for name in sorted(os.listdir(src)) if arc is None else [arc]:
            t.add(os.path.join(src, name), arcname=name)


def jsonl(path, rows):
    with open(path, "w") as f:
        for r in rows:
            f.write(json.dumps(r) + "\n")


def transcript(root):
    """A small grill session: the command, a read of the loaded skill, a
    hyp init in a subdirectory, a question and the owner's answer."""
    def ev(typ, content, cwd=root, meta=False):
        e = {"type": typ, "cwd": cwd, "message": {"role": typ, "content": content}}
        if typ == "assistant":
            e["message"]["model"] = "claude-sonnet-5-5"
        if meta:
            e["isMeta"] = True
        return e
    return [
        ev("user", "<command-name>/hyp-owner-grill</command-name>\n<command-args>#141</command-args>"),
        ev("user", [{"type": "text", "text": "Base directory for this skill: the loaded skill text"}], meta=True),
        ev("assistant", [{"type": "tool_use", "id": "t1", "name": "Read", "input": {"file_path": root + "/.claude/skills/hyp-owner-grill/SKILL.md"}}]),
        ev("user", [{"type": "tool_result", "tool_use_id": "t1", "content": "the arm's skill text"}]),
        ev("assistant", [{"type": "tool_use", "id": "t2", "name": "Bash", "input": {"command": "hyp status"}}], cwd=root + "/investigations/goal-prompt"),
        ev("user", [{"type": "tool_result", "tool_use_id": "t2", "content": [{"type": "text", "text": "GOAL G-1 ..."}]}]),
        ev("assistant", [{"type": "text", "text": "Round 1. 1. Where does the Goal go? (a) investigations/goal-prompt."}]),
        ev("user", "1 a"),
        ev("assistant", [{"type": "text", "text": "The grill is finished."}]),
    ]


def out_dir(d, fixture, stores, start):
    os.makedirs(d)
    with open(os.path.join(d, "state.txt"), "w") as f:
        f.write("arm=base\nfixture=%s\ntrial=1\nmodel=claude-sonnet-5-5\nskills_commit=0000000\nskill_sha=x\n"
                "skill_creator=/x\nlimit_min=50\nmax_exchanges=20\n" % fixture)
    json.dump({"arm": "base", "fixture": fixture, "trial": "1", "model": "claude-sonnet-5-5", "claude_version": "2.1.322 (Claude Code)",
               "start": start, "end": "end", "timed_out": False, "grill_session": "s1", "standin_session": "s2",
               "exchanges": 2, "transcript_kept": True, "subagents": 0, "cost_usd": 1.5, "cost_usd_literal": 2.25,
               "models": ["claude-sonnet-5-5"], "home_hyp_unchanged": True}, open(os.path.join(d, "drive.json"), "w"))
    json.dump([], open(os.path.join(d, "calls.json"), "w"))
    jsonl(os.path.join(d, "transcript.jsonl"), transcript("/sandbox/wt"))
    jsonl(os.path.join(d, "standin.jsonl"), [{"type": "assistant", "message": {"model": "claude-sonnet-5-5", "content": [{"type": "text", "text": "1 a"}]}}])
    with tarfile.open(os.path.join(d, "subagents.tar"), "w"):
        pass
    with tarfile.open(os.path.join(d, "stores.tar"), "w") as t:
        for rel, src in stores:
            t.add(src, arcname=os.path.join(rel, ".tree") if rel != "." else ".tree",
                  filter=lambda ti: None if os.path.basename(ti.name).startswith("index.sqlite") else ti)
    sk = tempfile.mkdtemp()
    os.makedirs(os.path.join(sk, "hyp-owner-grill"))
    open(os.path.join(sk, "hyp-owner-grill", "SKILL.md"), "w").write("---\nname: hyp-owner-grill\ndescription: x\n---\n\nThe arm's skill.\n")
    with tarfile.open(os.path.join(d, "skill.tar"), "w") as t:
        t.add(os.path.join(sk, "hyp-owner-grill"), arcname="hyp-owner-grill")


def store(old_goal):
    s = tempfile.mkdtemp()
    hyp(s, "init")
    if old_goal:
        hyp(s, "add", "goal", "--title", "An older Goal", "--line", "[P] an older line", "--actor", "owner-dat", "--why", "before the trial")
    return s


def main():
    for name in os.listdir(KA):
        shutil.rmtree(os.path.join(KA, name))
    # score-b-pass: a Goal in investigations/goal-prompt, after the start.
    s = store(True)
    time.sleep(1.2)
    start = time.time()
    time.sleep(1.2)
    g = hyp(s, "add", "goal", "--title", "A trial Goal", "--body", "For issue #141.\nMay change: skills/ and the text the change makes wrong.\nRead only: plans/.\nNever: merge a PR.",
            "--line", "[P] each SKILL.md passes quick_validate.py (sha256 %s) with 0 errors." % SHA, "--line", "[N] hyp prompt exits 1 on a body with no block.",
            "--actor", "owner-standin", "--why", "the stand-in approved it")
    q = hyp(s, "add", "dq", "--parent", g, "--title", "Which change?", "--why", "contested")
    hyp(s, "add", "hyp", "--parent", q, "--title", "Rival one", "--body", "A body.", "--why", "a rival")
    out_dir(os.path.join(KA, "score-b-pass", "input", "od"), "b", [("investigations/goal-prompt", os.path.join(s, ".tree"))], round(start, 3))

    # score-a-bad: a store at the top directory, and a line naming a cut hash.
    s1, s2 = store(False), store(False)
    time.sleep(1.2)
    start = time.time()
    time.sleep(1.2)
    hyp(s1, "add", "goal", "--title", "Root Goal", "--line", "[N] no write is refused.", "--line", "[P] graded by eval_skill.py (sha256 c47a38de…) with no error.", "--actor", "owner-standin", "--why", "w")
    os.makedirs(os.path.join(KA, "score-a-bad", "input"))
    out_dir(os.path.join(KA, "score-a-bad", "input", "od"), "a", [(".", os.path.join(s1, ".tree")), ("investigations/x", os.path.join(s2, ".tree"))], round(start, 3))

    # score-missing: no stores.tar.
    shutil.copytree(os.path.join(KA, "score-b-pass", "input", "od"), os.path.join(KA, "score-missing", "input", "od"))
    os.remove(os.path.join(KA, "score-missing", "input", "od", "stores.tar"))

    # score-bad-eval: an eval_skill.py that is not the pinned one.
    shutil.copytree(os.path.join(KA, "score-b-pass", "input", "od"), os.path.join(KA, "score-bad-eval", "input", "od"))
    os.makedirs(os.path.join(KA, "score-bad-eval", "input", "scripts"))
    open(os.path.join(KA, "score-bad-eval", "input", "scripts", "eval_skill.py"), "w").write("print('not the pinned grader')\n")

    # table cases: synthetic score records.
    def rec(d, arm, fx, trial, cost, checks):
        os.makedirs(os.path.join(d, "out"))
        with open(os.path.join(d, "stdout"), "w") as f:
            for k, v in (("arm", arm), ("fixture", fx), ("trial", trial), ("cost_usd", cost), ("cost_usd_literal", "none" if cost == "none" else "%.4f" % (float(cost) * 2))):
                f.write("%s\t%s\n" % (k, v))
        json.dump({"run_id": "%s-%s-%s" % (arm, fx, trial), "checks": [{"id": i, "passed": p} for i, p in checks]},
                  open(os.path.join(d, "out", "record.json"), "w"))
    A = [("has-goal", True), ("l1-store", True), ("l1-init", None), ("l2", None), ("l3", None), ("l4-hash", True), ("l4-grader", None), ("l5", None)]
    B = [("has-goal", True), ("l1-store", False), ("l1-init", None), ("l4-hash", True), ("l4-grader", None), ("l5", None), ("l6", None)]
    tp = os.path.join(KA, "table-pass", "input")
    runs = [("base", "a", "1", "2.0", A), ("base", "a", "2", "none", A), ("base", "a", "3", "1.0", A),
            ("base", "b", "1", "3.0", B), ("x", "a", "1", "1.0", A), ("x", "b", "1", "6.0", B), ("x", "b", "2", "3.0", B)]
    labels, reviews = {}, {}
    for i, (arm, fx, trial, cost, checks) in enumerate(runs):
        rec(os.path.join(tp, "s%d" % i), arm, fx, trial, cost, checks)
        lab = "trial-%02d" % (i + 1)
        labels[lab] = "%s-%s-%s" % (arm, fx, trial)
        reviews[lab] = {c: ("fail" if (arm, fx, trial, c) in {("base", "a", "2", "l2"), ("x", "b", "2", "l6")} else "pass") for c, p in checks if p is None}
    json.dump(labels, open(os.path.join(tp, "map.json"), "w"), indent=1)
    json.dump(reviews, open(os.path.join(tp, "reviews.json"), "w"), indent=1)
    tn = os.path.join(KA, "table-no-base", "input")
    rec(os.path.join(tn, "s0"), "x", "a", "1", "1.0", A)
    json.dump({"trial-01": "x-a-1"}, open(os.path.join(tn, "map.json"), "w"))
    json.dump({"trial-01": {c: "pass" for c, p in A if p is None}}, open(os.path.join(tn, "reviews.json"), "w"))


if __name__ == "__main__":
    main()
