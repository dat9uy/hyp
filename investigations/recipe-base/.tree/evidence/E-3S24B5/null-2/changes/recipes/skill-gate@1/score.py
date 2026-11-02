#!/usr/bin/env python3
"""score OUT_DIR: one trial's fingerprints, read from its transcript and changes.

OUT_DIR is hyp sandbox remove's, with state.env and transcript.jsonl kept
from this recipe's tmp. Prints key<TAB>value lines:

  arm, trial      from state.env
  skill_loaded    1 when the runner made a Skill call or a Read of a file
                  named SKILL.md, else 0
  failing_hyp     the number of Bash calls that end in error with hyp as the
                  command whose exit status the call returns (heredoc bodies
                  dropped; the last command after a newline, ; && or ||, and
                  after its last |)
  f_reading_pin   1 when an errored tool result holds hyp's refusal
                  "--decided-by reading contradicts recipe NAME@V", an
                  Experiment that declares reading while it pins a recipe
  f_body_limit    1 when an errored tool result holds hyp's refusal
                  "body has N runes (max 600)"
  end_state       recorded when the runner's events record a Result
                  (result.record), registered when they add an Experiment
                  (node.add of kind Experiment) but no Result, else none

The runner's events are the lines of the files under changes/.tree/events/
that remove copied: every line of an added file, and the lines of a
modified file whose actor is runner-<arm>-<trial>. Exits 2 when OUT_DIR
lacks state.env, transcript.jsonl or changes.json, and 0 otherwise.
"""
import json
import os
import re
import sys

LAST_HYP = re.compile(r"^(\w+=\S*\s+)*(\S*/)?hyp\s+[a-z-]")
HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n.*?\n\s*\1\s*(?=\n|$)", re.S)
READING_PIN = re.compile(r"--decided-by reading contradicts recipe \S+")
BODY_LIMIT = re.compile(r"body has \d+ runes \(max 600\)")


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def last_command(cmd):
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

    errors, reading_pin, body_limit = set(), 0, 0
    for b in blocks(lines, "tool_result", "user"):
        if b.get("is_error"):
            errors.add(b.get("tool_use_id"))
            text = result_text(b)
            reading_pin |= int(bool(READING_PIN.search(text)))
            body_limit |= int(bool(BODY_LIMIT.search(text)))

    seen, skill, failing = set(), 0, 0
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in seen:
            continue
        seen.add(b.get("id"))
        inp = b.get("input") or {}
        name = b.get("name")
        if name == "Skill" or (name == "Read" and os.path.basename(inp.get("file_path", "")) == "SKILL.md"):
            skill = 1
        if name == "Bash" and b.get("id") in errors and LAST_HYP.match(last_command(inp.get("command", ""))):
            failing += 1

    actor = f"runner-{st.get('arm', '')}-{st.get('trial', '')}"
    experiments, recorded = 0, 0
    for c in changes:
        path = c.get("path", "")
        if c.get("change") not in ("added", "modified"):
            continue
        if not (path.startswith(".tree/events/") and path.endswith(".jsonl")):
            continue
        p = os.path.join(out, "changes", path)
        if not os.path.isfile(p):
            continue
        for l in open(p):
            if not l.strip():
                continue
            e = json.loads(l)
            if c["change"] == "modified" and e.get("actor") != actor:
                continue
            if e.get("type") == "result.record":
                recorded = 1
            elif e.get("type") == "node.add" and (e.get("payload") or {}).get("kind") == "Experiment":
                experiments += 1
    end = "recorded" if recorded else ("registered" if experiments else "none")

    for k, v in [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("skill_loaded", skill),
                 ("failing_hyp", failing), ("f_reading_pin", reading_pin), ("f_body_limit", body_limit),
                 ("end_state", end)]:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
