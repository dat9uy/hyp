#!/usr/bin/env python3
"""score OUT_DIR: one trial's score, read from its transcript and the files it changed.

OUT_DIR is hyp sandbox remove's, with state.env and transcript.jsonl kept
from the runner's tmp, and changes.json and changes/ written by remove.
Prints key<TAB>value lines, in this order:

  arm, trial      from state.env
  skill_loaded    1 when the runner made a Skill call or a Read of a file
                  named SKILL.md, else 0
  failing_hyp     the number of Bash calls that end in error with hyp as the
                  command whose exit status the call returns (heredoc bodies
                  dropped)
  f_reading_pin   1 when an errored tool result holds hyp's refusal of an
                  Experiment that declares --decided-by reading while it pins
                  a recipe: "--decided-by reading contradicts recipe REF, an
                  instrument ..."; else 0
  f_body_limit    1 when an errored tool result holds hyp's refusal of a body
                  over 600 runes: "over limit: body has N runes (max 600)";
                  else 0
  end_state       recorded   the runner's events hold a result.record
                  registered they hold a node.add of an Experiment, no Result
                  none       neither

The runner's events are the lines of every .tree/events/*.jsonl that remove
lists as changed and copied: all lines of an added file, and of a modified
file only those whose actor is runner-<arm>-<trial>, the actor Drive sets.
This script reads files only and calls no hyp. Exits 2 when OUT_DIR lacks
state.env, transcript.jsonl or changes.json, or state.env lacks arm or
trial; it exits 0 otherwise.
"""
import json
import os
import re
import sys

FINGERPRINTS = [
    ("reading_pin", r"--decided-by reading contradicts recipe \S+, an instrument"),
    ("body_limit", r"over limit: body has \d+ runes \(max 600\)"),
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
    for k in ("arm", "trial"):
        if not st.get(k):
            fail("state.env has no " + k)
    changes = json.load(open(os.path.join(out, "changes.json")))["changes"]
    actor = f"runner-{st['arm']}-{st['trial']}"
    events = []
    for c in changes:
        p = c.get("path", "")
        if c.get("change") not in ("added", "modified") or not c.get("copied"):
            continue
        if not re.fullmatch(r"\.tree/events/[^/]+\.jsonl", p):
            continue
        with open(os.path.join(out, "changes", p), errors="replace") as f:
            for l in f:
                try:
                    ev = json.loads(l)
                except ValueError:
                    continue
                if isinstance(ev, dict) and (c["change"] == "added" or ev.get("actor") == actor):
                    events.append(ev)

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
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            if LAST_HYP.match(last_command(cmd)) and b.get("id") in errors:
                failing += 1
    texts = list(errors.values())
    rows = [("arm", st["arm"]), ("trial", st["trial"]), ("skill_loaded", loaded), ("failing_hyp", failing)]
    for name, pat in FINGERPRINTS:
        rows.append(("f_" + name, int(any(re.search(pat, t) for t in texts))))
    if any(e.get("type") == "result.record" for e in events):
        end = "recorded"
    elif any(e.get("type") == "node.add" and (e.get("payload") or {}).get("kind") == "Experiment" for e in events):
        end = "registered"
    else:
        end = "none"
    rows.append(("end_state", end))
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
