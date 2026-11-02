#!/usr/bin/env -S python3 -I
"""score.py OUT_DIR: score one trial from the evidence `hyp sandbox remove` kept.

Prints key<TAB>value lines: arm, trial, skill_loaded, failing_hyp,
f_reading_pin, f_body_limit, end_state. Exits 2 when the evidence lacks a
file it needs. See recipe.md, Evidence.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import tempfile
from datetime import datetime, timezone


def die(msg):
    sys.stderr.write("score.py: %s\n" % msg)
    sys.exit(2)


def load_json(path):
    try:
        with open(path, encoding="utf-8") as f:
            return json.load(f)
    except (OSError, ValueError) as e:
        die("cannot read %s: %s" % (path, e))


def parse_time(s):
    m = re.fullmatch(r"(\d{4}-\d\d-\d\dT\d\d:\d\d:\d\d)(\.\d+)?(Z|[+-]\d\d:\d\d)", str(s))
    if not m:
        die("bad time %r" % (s,))
    frac = (m.group(2) or ".0")[1:7].ljust(6, "0")
    zone = "+00:00" if m.group(3) == "Z" else m.group(3)
    return datetime.fromisoformat("%s.%s%s" % (m.group(1), frac, zone)).astimezone(timezone.utc)


# ---- the exit status of a Bash call -------------------------------------

def elements(cmd):
    """Split a shell command into and-or list elements.

    Returns a list of (separator_before, text); separators are '', '&&',
    '||', ';' (also newline and '&'). A pipeline stays inside one element.
    Quotes, $( ), backticks, comments and here-documents are skipped over.
    """
    out = []
    cur = []
    sep = ""
    i, n = 0, len(cmd)
    depth = 0          # open ( or $( outside quotes
    heredocs = []      # (delimiter, strip_tabs) waiting for the next newline
    word_start = True

    def flush(nextsep):
        nonlocal cur, sep
        out.append((sep, "".join(cur)))
        cur = []
        sep = nextsep

    while i < n:
        c = cmd[i]
        if c == "\\" and i + 1 < n:
            cur.append(cmd[i:i + 2]); i += 2; word_start = False; continue
        if c == "'":
            j = cmd.find("'", i + 1)
            j = n - 1 if j < 0 else j
            cur.append(cmd[i:j + 1]); i = j + 1; word_start = False; continue
        if c == '"':
            j = i + 1
            while j < n and cmd[j] != '"':
                j += 2 if cmd[j] == "\\" else 1
            cur.append(cmd[i:j + 1]); i = j + 1; word_start = False; continue
        if c == "`":
            j = cmd.find("`", i + 1)
            j = n - 1 if j < 0 else j
            cur.append(cmd[i:j + 1]); i = j + 1; word_start = False; continue
        if c == "#" and word_start:
            while i < n and cmd[i] != "\n":
                i += 1
            continue
        if c == "$" and cmd[i:i + 2] == "$(":
            depth += 1; cur.append("$("); i += 2; word_start = False; continue
        if c == "(":
            depth += 1; cur.append(c); i += 1; word_start = True; continue
        if c == ")":
            depth = max(0, depth - 1); cur.append(c); i += 1; word_start = False; continue
        if c == "<" and cmd[i:i + 2] == "<<" and cmd[i:i + 3] != "<<<":
            m = re.compile(r"<<(-?)[ \t]*(?:'([^']*)'|\"([^\"]*)\"|\\?([A-Za-z0-9_.-]+))").match(cmd, i)
            if m:
                heredocs.append((m.group(2) or m.group(3) or m.group(4), m.group(1) == "-"))
                cur.append(m.group(0)); i = m.end(); word_start = False; continue
        if c == "\n":
            if heredocs:
                i += 1
                for delim, strip in heredocs:
                    while i < n:
                        j = cmd.find("\n", i)
                        j = n if j < 0 else j
                        line = cmd[i:j]
                        i = min(j + 1, n)
                        if (line.lstrip("\t") if strip else line) == delim:
                            break
                heredocs = []
            else:
                i += 1
            if depth == 0:
                flush(";")
            else:
                cur.append(" ")
            word_start = True
            continue
        if depth == 0:
            if cmd[i:i + 2] == "&&":
                flush("&&"); i += 2; word_start = True; continue
            if cmd[i:i + 2] == "||":
                flush("||"); i += 2; word_start = True; continue
            if c == ";":
                flush(";"); i += 1; word_start = True; continue
            if c == "&" and cmd[i:i + 2] not in ("&>",) and (i == 0 or cmd[i - 1] not in "<>"):
                flush(";"); i += 1; word_start = True; continue
        cur.append(c)
        word_start = c in " \t"
        i += 1
    flush(";")
    return [(s, t) for s, t in out if t.strip()]


def last_stage(text):
    """The text of the last stage of a pipeline (split on a bare |)."""
    stages, cur = [], []
    i, n = 0, len(text)
    depth = 0
    while i < n:
        c = text[i]
        if c == "\\" and i + 1 < n:
            cur.append(text[i:i + 2]); i += 2; continue
        if c in "'\"`":
            j = i + 1
            while j < n and text[j] != c:
                j += 2 if (c == '"' and text[j] == "\\") else 1
            cur.append(text[i:j + 1]); i = j + 1; continue
        if c == "(":
            depth += 1
        elif c == ")":
            depth = max(0, depth - 1)
        if c == "|" and depth == 0 and text[i - 1:i] not in (">",):
            stages.append("".join(cur)); cur = []
            i += 2 if text[i:i + 2] == "|&" else 1
            continue
        cur.append(c)
        i += 1
    stages.append("".join(cur))
    return stages[-1]


def words(text):
    ws, cur, started = [], [], False
    i, n = 0, len(text)
    while i < n:
        c = text[i]
        if c in " \t\n":
            if started:
                ws.append("".join(cur)); cur = []; started = False
            i += 1
        elif c == "\\" and i + 1 < n:
            cur.append(text[i + 1]); started = True; i += 2
        elif c in "'\"":
            j = i + 1
            while j < n and text[j] != c:
                if c == '"' and text[j] == "\\" and j + 1 < n:
                    j += 1
                cur.append(text[j]); j += 1
            started = True; i = j + 1
        else:
            cur.append(c); started = True; i += 1
    if started:
        ws.append("".join(cur))
    return ws


WRAPPERS = {"command", "exec", "time", "env", "builtin", "nohup"}


def is_hyp(text):
    ws = words(last_stage(text))
    while ws:
        w = ws[0]
        if re.fullmatch(r"[A-Za-z_][A-Za-z0-9_]*=.*", w) or w in WRAPPERS or w.startswith("-") and len(ws) > 1 and ws[0] != "hyp":
            ws.pop(0)
            continue
        break
    return bool(ws) and os.path.basename(ws[0]) == "hyp"


def status_is_hyp(cmd):
    """True when the call's exit status, if non-zero, can be a hyp's.

    The status is the last command run, so when it fails it is the last
    element of the final and-or list or an element before it joined to it
    by && only; a pipeline's status is its last stage.
    """
    els = elements(cmd)
    k = len(els) - 1
    while k >= 0:
        if is_hyp(els[k][1]):
            return True
        if els[k][0] != "&&":
            break
        k -= 1
    return False


# ---- the transcript -----------------------------------------------------

def text_of(content):
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return "\n".join(b.get("text", "") for b in content if isinstance(b, dict) and b.get("type") == "text")
    return ""


READING_PIN = re.compile(r"--decided-by reading contradicts recipe \S+, an instrument whose checks decide the verdicts")
BODY_LIMIT = re.compile(r"over limit: body has \d+ runes \(max 600\)")


def read_transcript(path):
    uses = {}      # tool_use id -> (name, input)
    results = {}   # tool_use id -> (is_error, text)
    try:
        f = open(path, encoding="utf-8")
    except OSError as e:
        die("cannot read %s: %s" % (path, e))
    with f:
        for ln, line in enumerate(f, 1):
            if not line.strip():
                continue
            try:
                d = json.loads(line)
            except ValueError:
                die("%s line %d is not JSON" % (path, ln))
            # Only the conversation's own lines; api-request-blob and
            # other bookkeeping lines repeat the same tool calls.
            t = d.get("type")
            content = (d.get("message") or {}).get("content")
            if t not in ("assistant", "user") or not isinstance(content, list):
                continue
            for b in content:
                if not isinstance(b, dict):
                    continue
                if t == "assistant" and b.get("type") == "tool_use" and b.get("id"):
                    uses.setdefault(b["id"], (b.get("name"), b.get("input") or {}))
                elif t == "user" and b.get("type") == "tool_result" and b.get("tool_use_id"):
                    results.setdefault(b["tool_use_id"], (bool(b.get("is_error")), text_of(b.get("content"))))
    return uses, results


def applied_events(tree):
    """The applied events of the kept store OUT_DIR/tree, from `hyp events --json`."""
    if not os.path.isdir(os.path.join(tree, "events")):
        die("%s holds no events" % tree)
    tmp = tempfile.mkdtemp(prefix="score-")
    try:
        shutil.copytree(tree, os.path.join(tmp, ".tree"), symlinks=True)
        p = subprocess.run(["hyp", "events", "--json"], cwd=tmp, stdin=subprocess.DEVNULL,
                           stdout=subprocess.PIPE, stderr=subprocess.PIPE)
    finally:
        shutil.rmtree(tmp, ignore_errors=True)
    if p.returncode != 0:
        die("hyp events --json failed: %s" % p.stderr.decode("utf-8", "replace").strip())
    try:
        return json.loads(p.stdout)["events"]
    except (ValueError, KeyError):
        die("hyp events --json printed no events")


def main():
    if len(sys.argv) != 2:
        die("usage: score.py OUT_DIR")
    out = sys.argv[1]
    trial = load_json(os.path.join(out, "trial.json"))
    sandbox = load_json(os.path.join(out, "sandbox.json"))
    arm, num = trial.get("arm"), trial.get("trial")
    if not isinstance(arm, str) or not arm or not isinstance(num, int) or isinstance(num, bool):
        die("trial.json holds no arm and trial")
    created = parse_time(sandbox.get("created"))
    uses, results = read_transcript(os.path.join(out, "transcript.jsonl"))

    skill_loaded = 0
    failing = 0
    reading_pin = 0
    body_limit = 0
    for tid, (name, inp) in uses.items():
        if name == "Skill":
            skill_loaded = 1
        elif name == "Read" and os.path.basename(str(inp.get("file_path", ""))) == "SKILL.md":
            skill_loaded = 1
        err, text = results.get(tid, (False, ""))
        if not err:
            continue
        if READING_PIN.search(text):
            reading_pin = 1
        if BODY_LIMIT.search(text):
            body_limit = 1
        if name == "Bash" and status_is_hyp(str(inp.get("command", ""))):
            failing += 1

    # The runner's events: the applied events of the store the sandbox kept
    # (OUT_DIR/tree is wt/.tree after the run) stamped after the sandbox was
    # created; the cut's own events are older. Read through hyp events.
    added_exp = recorded = False
    for e in applied_events(os.path.join(out, "tree")):
        try:
            ts = parse_time(e["ts"])
        except (KeyError, TypeError):
            die("an applied event has no ts")
        if ts <= created:
            continue
        if e.get("type") == "result.record":
            recorded = True
        elif e.get("type") == "node.add" and (e.get("payload") or {}).get("kind") == "Experiment":
            added_exp = True
    end_state = "recorded" if recorded else "registered" if added_exp else "none"

    for k, v in (("arm", arm), ("trial", num), ("skill_loaded", skill_loaded), ("failing_hyp", failing),
                 ("f_reading_pin", reading_pin), ("f_body_limit", body_limit), ("end_state", end_state)):
        sys.stdout.write("%s\t%s\n" % (k, v))


main()
