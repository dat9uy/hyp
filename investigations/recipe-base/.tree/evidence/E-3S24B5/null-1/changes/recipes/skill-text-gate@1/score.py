#!/usr/bin/env python3
"""Score one trial from its kept files (OUT_DIR of hyp sandbox remove).

Prints key<TAB>value lines: arm, trial, skill_loaded, failing_hyp,
f_reading_pin, f_body_limit, end_state.
"""
import glob, json, os, re, sys

# hyp's refusals (internal/replay/admit.go, internal/replay/validate.go).
READING_PIN = re.compile(r"--decided-by reading contradicts recipe \S+, an instrument")
BODY_LIMIT = re.compile(r"over limit: body has \d+ runes \(max 600\)")
SEPARATORS = (";", "&&", "||", "|", "\n")


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(1)


def final_command(cmd):
    """The words of the last simple command of a shell line, whose exit
    status the line returns; None when the line ends in a background job."""
    cmd = cmd.replace("\\\n", " ")
    segs, cur, bg = [], [], False
    i, n = 0, len(cmd)
    quote = None
    depth = 0
    heredocs = []
    while i < n:
        c = cmd[i]
        if quote:
            cur.append(c)
            if c == quote:
                quote = None
            elif c == "\\" and quote == '"' and i + 1 < n:
                i += 1
                cur.append(cmd[i])
            i += 1
            continue
        if c == "\\" and i + 1 < n:
            cur.append(c + cmd[i + 1])
            i += 2
            continue
        if c in "'\"`":
            quote = c
            cur.append(c)
            i += 1
            continue
        if c == "#" and (not cur or cur[-1] in " \t"):
            while i < n and cmd[i] != "\n":
                i += 1
            continue
        if c == "(":
            depth += 1
        elif c == ")":
            depth = max(0, depth - 1)
        if depth == 0 and cmd.startswith("<<", i) and not cmd.startswith("<<<", i):
            m = re.match(r"<<-?\s*(?:'([^']*)'|\"([^\"]*)\"|([^\s;&|<>()]+))", cmd[i:])
            if m:
                heredocs.append(m.group(1) or m.group(2) or m.group(3))
                cur.append(m.group(0))
                i += len(m.group(0))
                continue
        if c == "\n" and heredocs:
            segs.append(("".join(cur), "\n"))
            cur, bg = [], False
            i += 1
            for delim in heredocs:
                while i < n:
                    j = cmd.find("\n", i)
                    line = cmd[i:] if j < 0 else cmd[i:j]
                    i = n if j < 0 else j + 1
                    if line.strip() == delim:
                        break
            heredocs = []
            continue
        if depth == 0:
            sep = None
            for s in SEPARATORS:
                if cmd.startswith(s, i):
                    sep = s
                    break
            if sep is None and c == "&":
                prev = cmd[i - 1] if i else ""
                if prev not in "<>" and not cmd.startswith("&>", i):
                    sep = "&"
            if sep:
                segs.append(("".join(cur), sep))
                cur = []
                i += len(sep)
                continue
        cur.append(c)
        i += 1
    segs.append(("".join(cur), ""))
    # last non-empty segment
    for text, sep in reversed(segs):
        if text.strip():
            if sep == "&":
                return None
            return strip_prefix(text.strip())
        if sep == "&":
            # "cmd &" ends in a background job
            continue
    return []


def strip_prefix(text):
    text = text.strip()
    if text.startswith("(") and text.endswith(")"):
        return final_command(text[1:-1]) or []
    words = re.findall(r"""(?:[^\s'"]|'[^']*'|"[^"]*")+""", text)
    while words and (re.match(r"^[A-Za-z_][A-Za-z0-9_]*=", words[0])
                     or words[0] in ("time", "command", "exec", "nohup", "env", "builtin", "!")):
        words.pop(0)
    return words


def is_hyp(cmd):
    words = final_command(cmd)
    if not words:
        return False
    w = words[0].strip("'\"")
    return os.path.basename(w) == "hyp"


def read_header(path):
    h = {}
    with open(path, encoding="utf-8") as f:
        for line in f:
            line = line.rstrip("\n")
            if not line.strip():
                break
            if "=" in line:
                k, v = line.split("=", 1)
                h.setdefault(k, v)
    return h


def text_of(content):
    if isinstance(content, str):
        return content
    if isinstance(content, list):
        return "\n".join(b.get("text", "") for b in content if isinstance(b, dict))
    return ""


def main():
    if len(sys.argv) != 2:
        fail("want one argument, OUT_DIR")
    out = sys.argv[1]
    inputs = [p for p in sorted(glob.glob(os.path.join(out, "inputs", "*"))) if os.path.isfile(p)]
    if len(inputs) != 1:
        fail("want exactly one file under inputs/, found %d" % len(inputs))
    hdr = read_header(inputs[0])
    arm, trial = hdr.get("arm", ""), hdr.get("trial", "")
    if not arm or not re.fullmatch(r"[1-9][0-9]*", trial):
        fail("the inputs file names no arm or no trial")
    tpath = os.path.join(out, "transcript.jsonl")
    cpath = os.path.join(out, "cut-events.txt")
    if not os.path.isfile(tpath):
        fail("no transcript.jsonl")
    if not os.path.isfile(cpath):
        fail("no cut-events.txt")
    with open(cpath, encoding="utf-8") as f:
        cut = {l.strip() for l in f if l.strip()}

    uses = {}  # tool_use id -> (name, input)
    results = []  # (tool_use_id, is_error, text)
    with open(tpath, encoding="utf-8") as f:
        for line in f:
            try:
                rec = json.loads(line)
            except ValueError:
                continue
            msg = rec.get("message") if isinstance(rec, dict) else None
            content = msg.get("content") if isinstance(msg, dict) else None
            if not isinstance(content, list):
                continue
            for b in content:
                if not isinstance(b, dict):
                    continue
                if b.get("type") == "tool_use":
                    uses[b.get("id")] = (b.get("name"), b.get("input") or {})
                elif b.get("type") == "tool_result":
                    results.append((b.get("tool_use_id"), b.get("is_error") is True,
                                    text_of(b.get("content"))))

    skill_loaded = 0
    for name, inp in uses.values():
        if name == "Skill":
            skill_loaded = 1
        elif name == "Read" and os.path.basename(str(inp.get("file_path", ""))) == "SKILL.md":
            skill_loaded = 1
    failing, reading, body = 0, 0, 0
    for uid, is_err, text in results:
        if not is_err:
            continue
        if READING_PIN.search(text):
            reading = 1
        if BODY_LIMIT.search(text):
            body = 1
        name, inp = uses.get(uid, (None, {}))
        if name == "Bash" and is_hyp(str(inp.get("command", ""))):
            failing += 1

    # The runner's own events: every event of a kept events file but the cut's.
    state = "none"
    for path in sorted(glob.glob(os.path.join(out, "changes", ".tree", "events", "*.jsonl"))):
        with open(path, encoding="utf-8") as f:
            for line in f:
                try:
                    ev = json.loads(line)
                except ValueError:
                    continue
                if not isinstance(ev, dict) or ev.get("id") in cut:
                    continue
                if ev.get("type") == "result.record":
                    state = "recorded"
                elif (ev.get("type") == "node.add" and state == "none"
                      and (ev.get("payload") or {}).get("kind") == "Experiment"):
                    state = "registered"

    for k, v in (("arm", arm), ("trial", trial), ("skill_loaded", skill_loaded),
                 ("failing_hyp", failing), ("f_reading_pin", reading),
                 ("f_body_limit", body), ("end_state", state)):
        print("%s\t%s" % (k, v))


main()
