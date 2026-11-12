#!/usr/bin/env python3
"""How the score known answers were made; not a command of this recipe.

fixture.py DIR builds each score case's input/ under DIR/<case>/input with
the hyp on PATH and git: a three-commit repository (base, C1, C2, C3), a
store whose Goal "build" holds Q "q" and Hypotheses tied to those commits,
the trial's events after a cut, review files, and subagent transcripts that
show (or fail to show) each commit's diff. The expected outputs in each
case's expect.json were written by hand from score.py's rules, reading the
display ids this run printed to ids.txt.
"""
import json
import os
import subprocess
import sys
import tarfile

ENV = dict(os.environ, GIT_AUTHOR_DATE="2026-01-01T00:00:00Z", GIT_COMMITTER_DATE="2026-01-01T00:00:00Z",
           GIT_CONFIG_NOSYSTEM="1")


def sh(cwd, *argv, actor="claude-t"):
    env = dict(ENV, HYP_ACTOR=actor)
    r = subprocess.run(list(argv), cwd=cwd, env=env, capture_output=True, text=True)
    if r.returncode != 0:
        sys.exit("%s: %s" % (" ".join(argv), r.stderr))
    return r.stdout.strip()


def last_id(cwd):
    return json.loads(sh(cwd, "hyp", "events", "--json"))["events"][-1]["id"]


def disp(cwd, ref):
    return sh(cwd, "hyp", "show", ref).splitlines()[0].split()[1]


def repo(d):
    os.makedirs(d)
    sh(d, "git", "init", "-q", "-b", "main")
    sh(d, "git", "config", "user.email", "f@x")
    sh(d, "git", "config", "user.name", "f")
    out = []
    for name, files in [("base", {"a.txt": "a\n"}), ("c1", {"a.txt": "a1\n", "b.txt": "b\n"}),
                        ("c2", {"b.txt": "b2\n"}), ("c3", {"c.txt": "c\n"})]:
        for f, t in files.items():
            open(os.path.join(d, f), "w").write(t)
        sh(d, "git", "add", ".")
        sh(d, "git", "commit", "-qm", name)
        out.append(sh(d, "git", "rev-parse", "HEAD"))
    return out


def commits(d):
    out = {}
    for sha in sh(d, "git", "rev-list", "--all").split():
        parents = sh(d, "git", "rev-list", "--parents", "-n", "1", sha).split()[1:]
        files = []
        if parents:
            for line in sh(d, "git", "diff-tree", "-r", "--no-commit-id", "--no-renames", parents[0], sha).splitlines():
                meta, path = line.split("\t", 1)
                f = meta.split()
                files.append({"path": path, "status": f[4][0], "blob": f[3]})
        out[sha] = {"parents": parents, "files": files}
    return {"head": sh(d, "git", "rev-parse", "HEAD"), "commits": out}


REVIEW = ("Review of the candidate's diff against the Goal's lines. Finding 1: none of the changed files "
          "breaks a line; the change to the text files is what the Hypothesis states. Finding 2: no test "
          "covers the new file, which is a limit of the review and not a defect of the candidate.")


def subagent_jsonl(path, diff_text, final, write_to=None):
    rows = [{"type": "user", "message": {"role": "user", "content": "Review this candidate's diff."}},
            {"type": "assistant", "message": {"role": "assistant", "content": [
                {"type": "tool_use", "id": "t1", "name": "Bash", "input": {"command": "git diff"}}]}},
            {"type": "user", "message": {"role": "user", "content": [
                {"type": "tool_result", "tool_use_id": "t1", "content": diff_text}]}}]
    if write_to:
        rows.append({"type": "assistant", "message": {"role": "assistant", "content": [
            {"type": "tool_use", "id": "t2", "name": "Write", "input": {"file_path": write_to, "content": final}}]}})
    rows.append({"type": "assistant", "message": {"role": "assistant", "content": [{"type": "text", "text": final}]}})
    with open(path, "w") as f:
        for r in rows:
            f.write(json.dumps(r) + "\n")


def pack(case, store, subs, st, n, cost, repo_dir):
    inp = os.path.join(case, "input")
    os.makedirs(inp)
    with open(os.path.join(inp, "state.txt"), "w") as f:
        for k, v in st.items():
            f.write("%s=%s\n" % (k, v))
    json.dump({"cost_usd": cost, "duration_ms": 1000, "num_turns": 9, "timed_out": cost is None},
              open(os.path.join(inp, "drive.json"), "w"))
    open(os.path.join(inp, "prompt.txt"), "w").write(
        "/goal Run the tree. Done when the latest printed hyp status G shows \"runner owes nothing\", "
        "or %d Results have been recorded under %s, or you cannot.\n" % (n, st["actor"]))
    open(os.path.join(inp, "transcript.jsonl"), "w").write("{}\n")
    with tarfile.open(os.path.join(inp, "subagents.tar"), "w") as t:
        for p in subs:
            t.add(p, arcname="subagents/" + os.path.basename(p))
    with tarfile.open(os.path.join(inp, "tree.tar"), "w") as t:
        t.add(os.path.join(store, ".tree"), arcname=".tree")
    json.dump(commits(repo_dir), open(os.path.join(inp, "commits.json"), "w"))


def review_exp(s, h, alias, final, subpath, diff_text, write=True, verdict="supports"):
    sh(s, "hyp", "add", "exp", "--parent", h, "--title", "review " + h, "--recipe", "none", "--rubric", "reading",
       "--alias", alias, "--why", "fx")
    e = disp(s, alias)
    rel = ".tree/evidence/%s/review.md" % e
    os.makedirs(os.path.join(s, os.path.dirname(rel)))
    open(os.path.join(s, rel), "w").write("# Review\n\n" + final + "\n")
    subagent_jsonl(subpath, diff_text, final, write_to=os.path.join("/wt/store", rel) if write else None)
    sh(s, "hyp", "record", "--exp", alias, "--verdict", "%s=%s" % (h, verdict), "--observation", "review",
       "--evidence", rel, "--why", "fx")


def main():
    out = os.path.abspath(sys.argv[1])
    work = os.path.join(out, "_work")
    ids = []
    r = os.path.join(work, "repo")
    base, c1, c2, c3 = repo(r)
    d12 = sh(r, "git", "diff", base, c1)
    d23 = sh(r, "git", "diff", c1, c2)

    # score-a-pass: cut after h1 is contradicted; the trial builds h2 (c2)
    # and a subagent that read c2's diff writes its review.
    for name, wrong in (("score-a-pass", False), ("score-a-unreviewed", True)):
        s = os.path.join(work, name)
        os.makedirs(s)
        sh(s, "git", "init", "-q")
        sh(s, "git", "fetch", "-q", r, "main")
        sh(s, "hyp", "init")
        sh(s, "hyp", "add", "goal", "--title", "Fixture build goal", "--body", "Fixture.", "--alias", "build",
           "--why", "fx", actor="owner-x")
        sh(s, "hyp", "add", "dq", "--parent", "build", "--title", "Which commit?", "--alias", "q", "--why", "fx")
        sh(s, "hyp", "add", "hyp", "--parent", "q", "--title", "c1", "--alias", "h1", "--why", "fx")
        sh(s, "hyp", "subject", "h1", "--base", c1, "--why", "fx")
        sd = os.path.join(work, name + "-sub")
        os.makedirs(sd)
        review_exp(s, "h1", "e1", REVIEW, os.path.join(sd, "agent-pre.jsonl"), d12, verdict="contradicts")
        cut = last_id(s)
        sh(s, "hyp", "add", "hyp", "--parent", "q", "--title", "c2", "--alias", "h2", "--why", "fx")
        sh(s, "hyp", "subject", "h2", "--base", c2, "--why", "fx")
        review_exp(s, "h2", "e2", REVIEW, os.path.join(sd, "agent-a.jsonl"), d12 if wrong else d23)
        if wrong:
            sh(s, "hyp", "add", "hyp", "--parent", "q", "--title", "c3", "--alias", "h3", "--why", "fx")
            sh(s, "hyp", "subject", "h3", "--base", c3, "--why", "fx")
        sh(s, "hyp", "checkpoint", "--goal", "build", "--summary", "Stopped: cap reached.", "--why", "fx")
        st = {"arm": "base", "cut": "a", "cut_event": cut, "trial": "1", "actor": "claude-t", "goal": disp(s, "build")}
        subs = [os.path.join(sd, "agent-a.jsonl")]
        pack(os.path.join(out, name), s, subs, st, 5, 2.5, r)
        ids.append("%s h1=%s h2=%s h3=%s c1=%s c2=%s c3=%s cut=%s" % (
            name, disp(s, "h1"), disp(s, "h2"), disp(s, "h3") if wrong else "-", c1, c2, c3, cut))

    # score-b-*: cut after h1 (c1) is tied; the trial reviews h1 (or not) and selects it.
    for name, n, reviewed, cp in (("score-b-pass", 5, True, ""), ("score-b-late", 1, True, None),
                                  ("score-b-unreviewed", 5, False, "Stopped: %s has no review." % c1[:9])):
        s = os.path.join(work, name)
        os.makedirs(s)
        sh(s, "git", "init", "-q")
        sh(s, "git", "fetch", "-q", r, "main")
        sh(s, "hyp", "init")
        sh(s, "hyp", "add", "goal", "--title", "Fixture build goal", "--body", "Fixture.", "--alias", "build",
           "--why", "fx", actor="owner-x")
        sh(s, "hyp", "add", "dq", "--parent", "build", "--title", "Which commit?", "--alias", "q", "--why", "fx",
           actor="claude-pre")
        sh(s, "hyp", "add", "hyp", "--parent", "q", "--title", "c1", "--alias", "h1", "--why", "fx", actor="claude-pre")
        sh(s, "hyp", "subject", "h1", "--base", c1, "--why", "fx", actor="claude-pre")
        cut = last_id(s)
        sd = os.path.join(work, name + "-sub")
        os.makedirs(sd)
        if reviewed:
            review_exp(s, "h1", "e1", REVIEW, os.path.join(sd, "agent-b.jsonl"), d12, write=False)
        else:
            sh(s, "hyp", "add", "exp", "--parent", "h1", "--title", "gate", "--recipe", "none", "--rubric", "reading",
               "--alias", "e1", "--why", "fx")
            e = disp(s, "e1")
            rel = ".tree/evidence/%s/gate.log" % e
            os.makedirs(os.path.join(s, os.path.dirname(rel)))
            open(os.path.join(s, rel), "w").write("go test ./... ok\n")
            subagent_jsonl(os.path.join(sd, "agent-b.jsonl"), d12, REVIEW)
            sh(s, "hyp", "record", "--exp", "e1", "--verdict", "h1=supports", "--observation", "gate",
               "--evidence", rel, "--why", "fx")
        sh(s, "hyp", "select", "q", "h1", "--why", "fx")
        if cp:
            sh(s, "hyp", "checkpoint", "--goal", "build", "--summary", cp, "--why", "fx")
        st = {"arm": "base", "cut": "b", "cut_event": cut, "trial": "2", "actor": "claude-t", "goal": disp(s, "build")}
        pack(os.path.join(out, name), s, [os.path.join(sd, "agent-b.jsonl")], st, n, None if name == "score-b-late" else 3.0, r)
        ids.append("%s h1=%s c1=%s cut=%s" % (name, disp(s, "h1"), c1, cut))
    open(os.path.join(out, "ids.txt"), "w").write("\n".join(ids) + "\n")


if __name__ == "__main__":
    main()
