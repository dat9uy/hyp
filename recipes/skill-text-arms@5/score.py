#!/usr/bin/env python3
"""score OUT_DIR: one trial's fingerprints, read from its transcript and changes.

OUT_DIR is hyp sandbox remove's, with state.env and transcript.jsonl kept
from this recipe's tmp. Prints key<TAB>value lines:

  arm, trial        from state.env
  skill_loaded      1 when the runner made a Skill call or a Read of a file
                    named SKILL.md
  failing_hyp       Bash calls that end in error with hyp as their last command
                    (the one whose exit status the call returns, heredoc
                    bodies dropped)
  f_reading_pin     P6: 1 when an errored tool result holds hyp's refusal
                    "--decided-by reading contradicts recipe"
  f_body_limit      S1: 1 when an errored tool result holds hyp's refusal
                    "body has N runes (max 600)", or the runner added or
                    modified a .md file outside .tree/ and recipes/
  f_launch_product  P5: 1 when a script named on a Run: line under ## Launch,
                    in a bundle the runner added under recipes/, has a line
                    naming HYP_SANDBOX_WT with >, tee, cp, mv or touch; or the
                    runner ran hyp sandbox remove more than once on one ID
  r_keep_tree       P4, a reading: the most --keep values that one hyp
                    sandbox remove call gives under one directory
  exp_added         1 when the runner's events add an Experiment pinned to
                    id-citation-arms@1
  result_recorded   1 when they record a Result under that Experiment
  end_state         recorded, registered (an Experiment but no Result) or none

Only the files under changes/ are read for the tree and bundles: events a
runner appends land there whole, because an events file is copied when it
is added or modified. Exits 2 when OUT_DIR lacks state.env, transcript.jsonl
or changes.json, and 0 otherwise.
"""
import json
import os
import re
import shlex
import sys

LAST_HYP = re.compile(r"^(\w+=\S*\s+)*(\S*/)?hyp\s+[a-z-]")
HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n.*?\n\s*\1\s*(?=\n|$)", re.S)
READING_PIN = "--decided-by reading contradicts recipe"
BODY_LIMIT = re.compile(r"body has \d+ runes \(max 600\)")
REMOVE = re.compile(r"\bhyp\s+sandbox\s+remove\b([^\n;&|]*)")
LAUNCH_WRITE = re.compile(r"HYP_SANDBOX_WT.*(>|\btee\b|\bcp\b|\bmv\b|\btouch\b)|(>|\btee\b|\bcp\b|\bmv\b|\btouch\b).*HYP_SANDBOX_WT")
BUNDLE = re.compile(r"^recipes/([a-z0-9][a-z0-9._-]*@[A-Za-z0-9._-]+)/")
PINNED = "id-citation-arms@1"


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def last_command(cmd):
    """The command whose exit status a Bash call returns: heredoc bodies dropped, then
    the last part after a newline, ;, && or ||, and after its last |."""
    parts = [p.strip() for p in re.split(r"\n|;|&&|\|\|", HEREDOC.sub("", cmd)) if p.strip()]
    return parts[-1].split("|")[-1].strip() if parts else ""


def blocks(lines, kind, role):
    for l in lines:
        content = (l.get("message") or {}).get("content")
        if l.get("type") == role and isinstance(content, list):
            for b in content:
                if isinstance(b, dict) and b.get("type") == kind:
                    yield b


def result_text(b):
    c = b.get("content")
    if isinstance(c, str):
        return c
    if isinstance(c, list):
        return "\n".join(x.get("text", "") for x in c if isinstance(x, dict))
    return ""


def remove_args(rest):
    """The sandbox ID and the --keep values of one hyp sandbox remove call."""
    try:
        toks = shlex.split(rest)
    except ValueError:
        toks = rest.split()
    keeps, pos, i = [], [], 0
    while i < len(toks):
        t = toks[i]
        if t in ("--keep", "-keep") and i + 1 < len(toks):
            keeps.append(toks[i + 1])
            i += 2
            continue
        if t.startswith(("--keep=", "-keep=")):
            keeps.append(t.split("=", 1)[1])
        elif not t.startswith("-") and not t.startswith(">") and t not in ("2>&1",):
            pos.append(t)
        i += 1
    return (pos[0] if pos else ""), keeps


def launch_scripts(recipe_md):
    """Script paths on Run: lines under ## Launch of one recipe.md."""
    out, inside = [], False
    for line in recipe_md.splitlines():
        if line.startswith("## "):
            inside = line.strip() == "## Launch"
        elif inside and line.startswith("Run:"):
            words = line[4:].split()
            if words:
                out.append(os.path.normpath(words[0]))
    return out


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.env", "transcript.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} holds no {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    lines = [json.loads(l) for l in open(os.path.join(out, "transcript.jsonl")) if l.strip()]
    changes = json.load(open(os.path.join(out, "changes.json")))["changes"]
    live = [c for c in changes if c.get("change") in ("added", "modified")]

    errors, reading_pin, body_limit = set(), 0, 0
    for b in blocks(lines, "tool_result", "user"):
        if b.get("is_error"):
            errors.add(b.get("tool_use_id"))
            text = result_text(b)
            reading_pin |= int(READING_PIN in text)
            body_limit |= int(bool(BODY_LIMIT.search(text)))

    seen, skill, failing, removes, keep_tree = set(), 0, 0, {}, 0
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in seen:
            continue
        seen.add(b.get("id"))
        inp = b.get("input") or {}
        name = b.get("name")
        if name == "Skill" or (name == "Read" and os.path.basename(inp.get("file_path", "")) == "SKILL.md"):
            skill = 1
        if name != "Bash":
            continue
        cmd = inp.get("command", "")
        if LAST_HYP.match(last_command(cmd)) and b.get("id") in errors:
            failing += 1
        for m in REMOVE.finditer(HEREDOC.sub("", cmd)):
            sid, keeps = remove_args(m.group(1))
            if sid:
                removes[sid] = removes.get(sid, 0) + 1
            dirs = {}
            for k in keeps:
                d = os.path.dirname(os.path.normpath(k))
                dirs[d] = dirs.get(d, 0) + 1
            keep_tree = max([keep_tree] + list(dirs.values()))

    md_outside = any(c["path"].endswith(".md") and not c["path"].startswith((".tree/", "recipes/")) for c in live)
    body_limit |= int(md_outside)

    def changed(path):
        p = os.path.join(out, "changes", path)
        return p if os.path.isfile(p) else None

    launch_product = int(any(n > 1 for n in removes.values()))
    added = {c["path"] for c in changes if c.get("change") == "added"}
    for c in live:
        m = BUNDLE.match(c["path"])
        if not (m and c["path"] == f"recipes/{m.group(1)}/recipe.md" and changed(c["path"])):
            continue
        for s in launch_scripts(open(changed(c["path"])).read()):
            sp = f"recipes/{m.group(1)}/{s}"
            if sp in added and changed(sp) and any(LAUNCH_WRITE.search(l) for l in open(changed(sp))):
                launch_product = 1

    exps, recorded = set(), 0
    events = []
    for c in live:
        if c["path"].startswith(".tree/events/") and c["path"].endswith(".jsonl") and changed(c["path"]):
            events += [json.loads(l) for l in open(changed(c["path"])) if l.strip()]
    for e in events:
        p = e.get("payload") or {}
        if e.get("type") == "node.add" and p.get("kind") == "Experiment" and p.get("recipe") == PINNED:
            exps.add(e.get("node"))
    for e in events:
        if e.get("type") == "result.record" and (e.get("payload") or {}).get("parent") in exps:
            recorded = 1
    end = "recorded" if recorded else ("registered" if exps else "none")

    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("skill_loaded", skill),
            ("failing_hyp", failing), ("f_reading_pin", reading_pin), ("f_body_limit", body_limit),
            ("f_launch_product", launch_product), ("r_keep_tree", keep_tree),
            ("exp_added", int(bool(exps))), ("result_recorded", recorded), ("end_state", end)]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
