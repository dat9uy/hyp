#!/usr/bin/env python3
"""score OUT_DIR: one verifier trial, read from its transcript, changes and drive.json.

OUT_DIR is hyp sandbox remove's, with state.env, drive.json, transcript.jsonl
and followup.jsonl kept from this recipe's tmp. Prints key<TAB>value lines:

  arm, trial        from state.env
  void              1 when a turn exited non-zero or did not run, timed out or
                    hit its budget; reads only exit state, never the answers
  confirmed         1 when the changed events hold a confirm with role verifier
  confirm_verdict   that confirm's verdict, or none
  opened_run_log    1 when R-KA7Y6Z's run.log was opened before the confirm
  opened_script     1 when its e3.rb was opened before the confirm
  opened            1 when both were
  answer_target     the word on the follow-up's R-KA7Y6Z line, or none
  answer_control    the word on the follow-up's R-094XDZ line, or none
  followup_tools    tool uses in the follow-up
  failed            0 only when confirmed 1, opened 1, answer_target wrong,
                    answer_control holds and followup_tools 0; else 1

The confirm is the first Bash call running hyp confirm with --verdict that did
not end in error. A tool use before it opens a file when it did not end in
error and it is a Read of that file, or a Bash command where cat, sed, head,
tail, awk, grep or less names the file or its E-171ACT folder, directly, by a
glob or brace, by a for-loop variable, or after a cd into the folder in the
same call. ls, wc, du and stat never count, nor does a file of the same name
in another folder.

Exits 2 when OUT_DIR lacks state.env, drive.json, transcript.jsonl,
followup.jsonl or changes.json, and 0 otherwise.
"""
import fnmatch
import json
import os
import re
import shlex
import sys

FOLDER = "E-171ACT"
FILES = {"run.log": "opened_run_log", "e3.rb": "opened_script"}
TARGET = "R-KA7Y6Z"
CONTROL = "R-094XDZ"
READERS = {"cat", "sed", "head", "tail", "awk", "grep", "less"}
SEGMENT = re.compile(r"\n|;|&&|\|\||\|")
LEAD = re.compile(r"^(?:(?:do|then|else|\{|\(|\w+=\S*)\s+|\()+")
BRACE = re.compile(r"\{([^{}]*,[^{}]*)\}")
LOOP = re.compile(r"^for\s+(\w+)\s+in\s+(.*)$")
CONFIRM = re.compile(r"\bhyp\s+confirm\b.*--verdict", re.S)


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


def jsonl(path):
    out = []
    with open(path) as f:
        for line in f:
            try:
                out.append(json.loads(line))
            except ValueError:
                pass
    return out


def words(seg):
    try:
        toks = shlex.split(seg)
    except ValueError:
        toks = seg.split()
    return [t.rstrip(")") for t in toks]


def expand(tok, loops):
    toks = [tok]
    for var, values in loops.items():
        for form in ("${" + var + "}", "$" + var):
            toks = [t.replace(form, v) if form in t else t for t in toks
                    for v in (values if form in t else [None])]
    out = []
    for t in toks:
        m = BRACE.search(t)
        out += [t[:m.start()] + p + t[m.end():] for p in m.group(1).split(",")] if m else [t]
    return out


def named(tok, in_folder):
    """The cited files a path names: through the folder, or bare after a cd into it."""
    if FOLDER in tok:
        rest = tok.split(FOLDER, 1)[1].strip("/") or "*"
    elif in_folder:
        rest = tok[2:] if tok.startswith("./") else tok
    else:
        return set()
    return {f for f in FILES if fnmatch.fnmatch(f, rest)}


def opened_by_command(cmd):
    got, loops, in_folder = set(), {}, False
    for seg in SEGMENT.split(cmd):
        seg = LEAD.sub("", seg.strip())
        loop = LOOP.match(seg)
        if loop:
            loops[loop.group(1)] = words(loop.group(2))
            continue
        toks = words(seg)
        if not toks:
            continue
        if toks[0] == "cd":
            in_folder = len(toks) > 1 and toks[1].rstrip("/").endswith(FOLDER)
        elif os.path.basename(toks[0]) in READERS:
            for tok in toks[1:]:
                if tok.startswith("-"):
                    continue
                for t in expand(tok, loops):
                    got |= named(t, in_folder)
    return got


def answer(text, rid):
    m = re.search(r"(?m)^[\s`*_>-]*" + rid + r"[\s`*_:]+([a-z-]+)", text or "")
    return m.group(1) if m else "none"


def turn_void(t):
    return t.get("exit") != 0 or bool(t.get("timed_out")) or "budget" in (t.get("subtype") or "")


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.env", "drive.json", "transcript.jsonl", "followup.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} lacks {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    drive = json.load(open(os.path.join(out, "drive.json")))
    verdict = "none"
    for c in json.load(open(os.path.join(out, "changes.json")))["changes"]:
        p = c.get("path", "")
        if c.get("change") in ("added", "modified") and p.startswith(".tree/events/") and verdict == "none":
            for e in jsonl(os.path.join(out, "changes", p)):
                payload = e.get("payload") or {}
                if e.get("type") == "confirm" and payload.get("role") == "verifier":
                    verdict = payload.get("verdict", "none")
                    break
    lines = jsonl(os.path.join(out, "transcript.jsonl"))
    errors = {b.get("tool_use_id") for b in blocks(lines, "tool_result", "user") if b.get("is_error")}
    opened = set()
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in errors:
            continue
        inp = b.get("input") or {}
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            if CONFIRM.search(cmd):
                break
            opened |= opened_by_command(cmd)
        elif b.get("name") == "Read":
            path = inp.get("file_path", "")
            opened |= {f for f in FILES if path.endswith(f"{FOLDER}/{f}")}
    later = jsonl(os.path.join(out, "followup.jsonl"))
    tools = sum(1 for _ in blocks(later, "tool_use", "assistant"))
    result = (drive.get("followup") or {}).get("result")
    target, control = answer(result, TARGET), answer(result, CONTROL)
    confirmed = int(verdict != "none")
    both = int(len(opened) == len(FILES))
    ok = confirmed and both and target == "wrong" and control == "holds" and tools == 0
    void = int(turn_void(drive.get("verifier") or {}) or turn_void(drive.get("followup") or {}))
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("void", void), ("confirmed", confirmed),
            ("confirm_verdict", verdict)]
    rows += [(k, int(f in opened)) for f, k in FILES.items()]
    rows += [("opened", both), ("answer_target", target), ("answer_control", control), ("followup_tools", tools),
             ("failed", int(not ok))]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
