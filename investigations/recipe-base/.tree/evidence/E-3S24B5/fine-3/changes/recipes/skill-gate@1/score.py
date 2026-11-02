#!/usr/bin/env python3
"""score OUT_DIR: one trial of the skill-text gate, read from its evidence.

OUT_DIR is hyp sandbox remove's, with state.env and transcript.jsonl kept
from the runner's tmp, and changes.json (and changes/) written by remove.
Prints key<TAB>value lines, in this order:

  arm, trial       from state.env
  skill_loaded     1 when the runner made a Skill call or a Read of a file
                   named SKILL.md, else 0
  failing_hyp      Bash calls that end in error with hyp as their last command
                   (the one whose exit status the call returns, heredoc
                   bodies dropped)
  f_reading_pin    1 when an errored tool result holds hyp's refusal of an
                   Experiment that declares --decided-by reading while it
                   pins a recipe ("--decided-by reading contradicts recipe"),
                   else 0
  f_body_limit     1 when an errored tool result holds hyp's refusal of a
                   body over 600 runes ("over limit: body has N runes (max
                   600)"), else 0
  end_state        among the events in changes/.tree/events/*.jsonl whose
                   actor is runner-<arm>-<trial>: recorded when one is a
                   result.record; else registered when one is a node.add of
                   kind Experiment; else none

Exits 2 when OUT_DIR lacks state.env, transcript.jsonl or changes.json, and
0 otherwise. Reads files only and calls no hyp.
"""
import json
import os
import re
import sys

FINGERPRINTS = [
    ("reading_pin", "error", r"--decided-by reading contradicts recipe"),
    ("body_limit", "error", r"over limit: body has \d+ runes \(max 600\)"),
]

LAST_HYP = re.compile(r"^(\w+=\S*\s+)*(\S*/)?hyp\s+[a-z-]")
HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n.*?\n\s*\1\s*(?=\n|$)", re.S)


def last_command(cmd):
    """The command whose exit status a Bash call returns: heredoc bodies dropped, then
    the last part after a newline, ;, && or ||, and after its last |."""
    parts = [p.strip() for p in re.split(r"\n|;|&&|\|\|", HEREDOC.sub("", cmd)) if p.strip()]
    return parts[-1].split("|")[-1].strip() if parts else ""


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def blocks(lines, kind, role):
    for l in lines:
        content = (l.get("message") or {}).get("content")
        if l.get("type") == role and isinstance(content, list):
            for b in content:
                if isinstance(b, dict) and b.get("type") == kind:
                    yield b


def text_of(block):
    c = block.get("content")
    if isinstance(c, list):
        return "\n".join(x.get("text", "") for x in c if isinstance(x, dict))
    return c if isinstance(c, str) else ""


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.env", "transcript.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} holds no {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    with open(os.path.join(out, "transcript.jsonl")) as f:
        lines = [json.loads(l) for l in f if l.strip()]
    changes = json.load(open(os.path.join(out, "changes.json")))["changes"]
    live = [c["path"] for c in changes if c.get("change") in ("added", "modified")]

    errors = {}
    for b in blocks(lines, "tool_result", "user"):
        if b.get("is_error"):
            errors[b.get("tool_use_id")] = text_of(b)
    seen, failing, calls, loaded, bash = set(), 0, 0, 0, []
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in seen:
            continue
        seen.add(b.get("id"))
        calls += 1
        inp = b.get("input") or {}
        if b.get("name") == "Skill" or (b.get("name") == "Read" and os.path.basename(inp.get("file_path", "")) == "SKILL.md"):
            loaded = 1
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            bash.append(cmd)
            if LAST_HYP.match(last_command(cmd)) and b.get("id") in errors:
                failing += 1
    where = {"error": list(errors.values()), "bash": bash, "path": live}
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("skill_loaded", loaded),
            ("failing_hyp", failing)]
    for name, w, pat in FINGERPRINTS:
        rows.append(("f_" + name, int(any(re.search(pat, t, re.M) for t in where[w]))))
    rows.append(("end_state", end_state(out, live, f"runner-{st.get('arm', '')}-{st.get('trial', '')}")))
    for k, v in rows:
        print(f"{k}\t{v}")


def end_state(out, live, actor):
    state = "none"
    for p in live:
        if not re.fullmatch(r"\.tree/events/[^/]+\.jsonl", p):
            continue
        f = os.path.join(out, "changes", p)
        if not os.path.isfile(f):
            continue
        for l in open(f):
            try:
                e = json.loads(l)
            except ValueError:
                continue
            if not isinstance(e, dict) or e.get("actor") != actor:
                continue
            if e.get("type") == "result.record":
                return "recorded"
            if e.get("type") == "node.add" and (e.get("payload") or {}).get("kind") == "Experiment":
                state = "registered"
    return state


if __name__ == "__main__":
    main()
