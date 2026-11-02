#!/usr/bin/env python3
"""score OUT_DIR: one trial's score, read from its kept evidence.

OUT_DIR is hyp sandbox remove's, with state.env and transcript.jsonl kept from
the runner's tmp. It reads state.env, transcript.jsonl, sandbox.json and the
events files remove copied under changes/.tree/events/. It prints these
key<TAB>value lines, in this order:

  arm, trial     from state.env
  skill_loaded   1 when the runner made a Skill tool call, or a Read of a file
                 named SKILL.md; else 0
  failing_hyp    the number of Bash calls that ended in error (is_error) with
                 hyp as the command whose exit status the call returns: the
                 last command of the call, after heredoc bodies are dropped
                 and splitting at newlines, ; && || outside quotes, and after
                 the last | outside quotes
  f_reading_pin  1 when an errored tool result holds hyp's refusal of an
                 Experiment that declares --decided-by reading while it pins
                 a recipe ("--decided-by reading contradicts recipe"), else 0
  f_body_limit   1 when an errored tool result holds hyp's refusal of a body
                 over 600 runes ("over limit: body has N runes (max 600)"),
                 else 0
  end_state      the runner's events are the lines of changes/.tree/events/
                 *.jsonl whose ts is at or after the sandbox's created time.
                 "recorded" when one has type result.record; else
                 "registered" when one is a node.add with payload kind
                 Experiment; else "none"

A tool result counts once per call, in any case. Reads files only and calls no
hyp. Exits 2 when OUT_DIR lacks state.env, transcript.jsonl or sandbox.json,
or a file does not parse, and 0 otherwise.
"""
import glob
import json
import os
import re
import sys
from datetime import datetime

FINGERPRINTS = [
    ("reading_pin", r"--decided-by reading contradicts recipe"),
    ("body_limit", r"over limit: body has \d+ runes \(max 600\)"),
]

LAST_HYP = re.compile(r"^(\w+=\S*\s+)*(\S*/)?hyp(\s|$)")
HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n.*?\n\s*\1\s*(?=\n|$)", re.S)


def last_command(cmd):
    """The command whose exit status a Bash call returns."""
    cmd = HEREDOC.sub("", cmd)
    parts, cur, q, i = [], [], None, 0
    while i < len(cmd):
        c = cmd[i]
        if q:
            cur.append(c)
            if c == "\\" and q == '"' and i + 1 < len(cmd):
                i += 1
                cur.append(cmd[i])
            elif c == q:
                q = None
        elif c == "\\" and i + 1 < len(cmd):
            cur.append(c + cmd[i + 1])
            i += 1
        elif c in "'\"":
            q = c
            cur.append(c)
        elif c in "\n;" or cmd.startswith("&&", i) or cmd.startswith("||", i):
            parts.append("".join(cur))
            cur = []
            if c not in "\n;":
                i += 1
        elif c == "|":
            cur = []  # the status of a pipeline is its last command's
        else:
            cur.append(c)
        i += 1
    parts.append("".join(cur))
    parts = [p.strip() for p in parts if p.strip()]
    return parts[-1] if parts else ""


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


def when(s):
    m = re.match(r"(\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d)(\.\d+)?(Z|[+-]\d\d:\d\d)$", s or "")
    if not m:
        fail(f"bad time {s!r}")
    frac = (m.group(2) or ".0")[1:7].ljust(6, "0")
    return datetime.fromisoformat(f"{m.group(1)}.{frac}" + ("+00:00" if m.group(3) == "Z" else m.group(3)))


def end_state(out):
    created = when(json.load(open(os.path.join(out, "sandbox.json"))).get("created"))
    recorded = registered = False
    for p in sorted(glob.glob(os.path.join(out, "changes", ".tree", "events", "*.jsonl"))):
        for l in open(p):
            if not l.strip():
                continue
            ev = json.loads(l)
            if when(ev.get("ts")) < created:
                continue
            if ev.get("type") == "result.record":
                recorded = True
            elif ev.get("type") == "node.add" and (ev.get("payload") or {}).get("kind") == "Experiment":
                registered = True
    return "recorded" if recorded else "registered" if registered else "none"


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.env", "transcript.jsonl", "sandbox.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} holds no {f}")
    try:
        st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
        with open(os.path.join(out, "transcript.jsonl")) as f:
            lines = [json.loads(l) for l in f if l.strip()]
        end = end_state(out)
    except ValueError as e:
        fail(f"{out}: {e}")

    errors = {}
    for b in blocks(lines, "tool_result", "user"):
        if b.get("is_error"):
            errors[b.get("tool_use_id")] = text_of(b)
    seen, failing, loaded = set(), 0, 0
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in seen:
            continue
        seen.add(b.get("id"))
        inp = b.get("input") or {}
        if b.get("name") == "Skill" or (b.get("name") == "Read" and os.path.basename(inp.get("file_path", "")) == "SKILL.md"):
            loaded = 1
        if b.get("name") == "Bash" and b.get("id") in errors and LAST_HYP.match(last_command(inp.get("command", ""))):
            failing += 1
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("skill_loaded", loaded),
            ("failing_hyp", failing)]
    for name, pat in FINGERPRINTS:
        rows.append(("f_" + name, int(any(re.search(pat, t) for t in errors.values()))))
    rows.append(("end_state", end))
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
