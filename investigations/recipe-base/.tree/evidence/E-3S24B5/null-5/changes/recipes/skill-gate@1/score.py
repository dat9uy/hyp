#!/usr/bin/env python3
"""score.py DIR: print key<TAB>value lines for the trial evidence in DIR."""
import json
import os
import re
import sys

PIN = re.compile(r"--decided-by reading contradicts recipe")
LIMIT = re.compile(r"over limit: body has \d+ runes \(max 600\)")
ARM = re.compile(r"^[a-z][a-z0-9-]{0,31}$")
OPS = (";", "&&", "||", "|", "&", "\n")


def die(msg):
    print("score.py: " + msg, file=sys.stderr)
    sys.exit(2)


def load(path):
    try:
        with open(path, encoding="utf-8") as f:
            return json.load(f)
    except (OSError, ValueError) as e:
        die("%s: %s" % (path, e))


def strip_heredocs(cmd):
    out, delim = [], None
    for line in cmd.split("\n"):
        if delim is not None:
            if line.strip() == delim:
                delim = None
            continue
        out.append(line)
        m = re.search(r"<<-?\s*['\"]?([A-Za-z_][A-Za-z0-9_]*)['\"]?", line)
        if m:
            delim = m.group(1)
    return "\n".join(out)


def segments(cmd):
    """Split on control operators outside quotes. Returns (segments, ends_in_background)."""
    segs, cur, i, n, q = [], [], 0, len(cmd), None
    bg = False
    while i < n:
        c = cmd[i]
        if q:
            cur.append(c)
            if c == "\\" and q == '"' and i + 1 < n:
                cur.append(cmd[i + 1])
                i += 1
            elif c == q:
                q = None
        elif c in "'\"":
            q = c
            cur.append(c)
        elif c == "\\" and i + 1 < n:
            cur.append(c + cmd[i + 1])
            i += 1
        elif c == "#" and (not cur or cur[-1] in " \t"):
            while i < n and cmd[i] != "\n":
                i += 1
            continue
        else:
            op = None
            for o in ("&&", "||"):
                if cmd.startswith(o, i):
                    op = o
            if op is None and c in ";|\n":
                op = c
            if op is None and c == "&":
                prev = cmd[i - 1] if i else ""
                nxt = cmd[i + 1] if i + 1 < n else ""
                if prev not in "<>" and nxt != ">":
                    op = "&"
            if op:
                segs.append("".join(cur))
                cur = []
                bg = op == "&"
                i += len(op)
                continue
            cur.append(c)
        i += 1
    tail = "".join(cur)
    if tail.strip():
        segs.append(tail)
        bg = False
    return segs, bg


def returns_hyp(cmd):
    cmd = strip_heredocs(cmd)
    segs, bg = segments(cmd)
    if bg:
        return False
    segs = [s for s in segs if s.strip()]
    if not segs:
        return False
    words = segs[-1].strip().lstrip("({ \t").split()
    skip = {"!", "time", "command", "exec", "env"}
    for w in words:
        if w in skip or re.match(r"^[A-Za-z_][A-Za-z0-9_]*=", w):
            continue
        return os.path.basename(w.strip("'\"")) == "hyp"
    return False


def text_of(content):
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return "\n".join(b.get("text", "") for b in content if isinstance(b, dict))
    return ""


def main():
    if len(sys.argv) != 2:
        die("want one argument: the trial evidence directory")
    d = sys.argv[1]
    meta = load(os.path.join(d, "meta.json"))
    ne = load(os.path.join(d, "new-events.json"))
    arm, trial = meta.get("arm"), meta.get("trial")
    if not isinstance(arm, str) or not ARM.match(arm):
        die("meta.json: bad arm %r" % (arm,))
    if not isinstance(trial, int) or isinstance(trial, bool) or trial < 1:
        die("meta.json: bad trial %r" % (trial,))
    if not isinstance(ne, list):
        die("new-events.json: want a list")

    calls, errors = {}, []
    loaded = 0
    try:
        f = open(os.path.join(d, "transcript.jsonl"), encoding="utf-8")
    except OSError as e:
        die(str(e))
    with f:
        for ln, line in enumerate(f, 1):
            if not line.strip():
                continue
            try:
                ev = json.loads(line)
            except ValueError as e:
                die("transcript.jsonl line %d: %s" % (ln, e))
            msg = ev.get("message") if isinstance(ev, dict) else None
            content = msg.get("content") if isinstance(msg, dict) else None
            if not isinstance(content, list):
                continue
            for b in content:
                if not isinstance(b, dict):
                    continue
                if b.get("type") == "tool_use":
                    inp = b.get("input") if isinstance(b.get("input"), dict) else {}
                    name = b.get("name")
                    if name == "Skill" or (
                        name == "Read"
                        and os.path.basename(str(inp.get("file_path", ""))) == "SKILL.md"
                    ):
                        loaded = 1
                    if name == "Bash":
                        calls[b.get("id")] = str(inp.get("command", ""))
                elif b.get("type") == "tool_result" and b.get("is_error") is True:
                    errors.append((b.get("tool_use_id"), text_of(b.get("content"))))

    failing = sum(1 for tid, _ in errors if tid in calls and returns_hyp(calls[tid]))
    pin = int(any(PIN.search(t) for _, t in errors))
    lim = int(any(LIMIT.search(t) for _, t in errors))
    types = [(e.get("type"), e.get("kind")) for e in ne if isinstance(e, dict)]
    if any(t == "result.record" for t, _ in types):
        end = "recorded"
    elif any(t == "node.add" and k == "Experiment" for t, k in types):
        end = "registered"
    else:
        end = "none"
    for k, v in (("arm", arm), ("trial", trial), ("skill_loaded", loaded),
                 ("failing_hyp", failing), ("f_reading_pin", pin),
                 ("f_body_limit", lim), ("end_state", end)):
        print("%s\t%s" % (k, v))


main()
