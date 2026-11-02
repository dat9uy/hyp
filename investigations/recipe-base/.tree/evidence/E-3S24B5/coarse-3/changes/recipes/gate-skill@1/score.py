#!/usr/bin/env python3
"""score OUT_DIR: one trial's stops, read from its transcript and changes.

OUT_DIR is hyp sandbox remove's, with state.env, drive.json and
transcript.jsonl kept from this recipe's tmp (Drive writes the last two),
and changes.json and changes/ written by remove. Prints key<TAB>value lines:

  arm, trial      from state.env
  skill_loaded    1 when the runner made a Skill call or a Read of a file
                  named SKILL.md, else 0
  failing_hyp     Bash calls that end in error with hyp as their last command
                  (the one whose exit status the call returns, heredoc
                  bodies dropped)
  f_reading_pin   1 when an errored tool result holds hyp's refusal of an
                  Experiment that declares --decided-by reading while it
                  pins a recipe ("--decided-by reading contradicts recipe
                  NAME@V, an instrument ..."), else 0
  f_body_limit    1 when an errored tool result holds hyp's refusal of a body
                  over 600 runes ("over limit: body has N runes (max 600)"),
                  else 0
  end_state       recorded   when the runner's events hold a result.record
                  registered when they hold a node.add of an Experiment but
                             no result.record
                  none       otherwise

The runner's events are the lines of the event files (.tree/events/*.jsonl)
under changes/ that the runner added or changed: every line of an added
file, and for a modified file (the store at the cut was copied in) only the
lines whose actor is runner-<arm>-<trial>, the actor Drive sets. This script
reads files only and calls no hyp. Exits 2 when OUT_DIR lacks state.env,
drive.json, transcript.jsonl or changes.json, and 0 otherwise.
"""
import json
import os
import re
import sys

FINGERPRINTS = [
    ("reading_pin", "error", r"--decided-by reading contradicts recipe \S+, an instrument"),
    ("body_limit", "error", r"over limit: body has \d+ runes \(max 600\)"),
]
EVENT_FILE = re.compile(r"(^|/)\.tree/events/[^/]+\.jsonl$")

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
    for f in ("state.env", "drive.json", "transcript.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} holds no {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    with open(os.path.join(out, "transcript.jsonl")) as f:
        lines = [json.loads(l) for l in f if l.strip()]
    drive = json.load(open(os.path.join(out, "drive.json")))
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
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")),
            ("skill_loaded", loaded), ("failing_hyp", failing)]
    for name, w, pat in FINGERPRINTS:
        rows.append(("f_" + name, int(any(re.search(pat, t, re.M) for t in where[w]))))
    rows.append(("end_state", end_state(out, changes, "runner-%s-%s" % (st.get("arm", ""), st.get("trial", "")))))
    for k, v in rows:
        print(f"{k}\t{v}")


def end_state(out, changes, actor):
    registered = recorded = False
    for c in changes:
        p = c.get("path", "")
        if c.get("change") not in ("added", "modified") or not EVENT_FILE.search(p):
            continue
        try:
            with open(os.path.join(out, "changes", p)) as f:
                text = f.read()
        except OSError:
            continue
        for line in text.splitlines():
            try:
                ev = json.loads(line)
            except ValueError:
                continue
            if not isinstance(ev, dict) or (c["change"] == "modified" and ev.get("actor") != actor):
                continue
            if ev.get("type") == "result.record":
                recorded = True
            elif ev.get("type") == "node.add" and (ev.get("payload") or {}).get("kind") == "Experiment":
                registered = True
    return "recorded" if recorded else "registered" if registered else "none"


if __name__ == "__main__":
    main()
