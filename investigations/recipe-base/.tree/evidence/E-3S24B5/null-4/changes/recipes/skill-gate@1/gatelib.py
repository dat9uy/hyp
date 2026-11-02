"""Shared helpers of skill-gate@1: inputs, state, shell parsing, scoring, verdict."""
import hashlib
import json
import os
import re
import sys

SKILLS = ("hyp-prototype", "hyp-verification")
ARM_RE = re.compile(r"^[a-z][a-z0-9-]{0,31}$")
FINGERPRINTS = ("f_reading_pin", "f_body_limit")
REFUSE = 30


def die(msg, code=REFUSE):
    sys.stderr.write(msg + "\n")
    sys.exit(code)


def sha256_file(path):
    h = hashlib.sha256()
    with open(path, "rb") as f:
        h.update(f.read())
    return h.hexdigest()


def read_inputs(path):
    d = {}
    with open(path) as f:
        for line in f:
            line = line.rstrip("\n")
            if not line.strip() or line.lstrip().startswith("#"):
                continue
            k, sep, v = line.partition("=")
            if not sep:
                die("inputs: line without '=': %r" % line, 1)
            d[k.strip()] = v.strip()
    arm = d.get("arm", "")
    if not ARM_RE.match(arm) or arm == "repro":
        die("inputs: arm %r is not [a-z][a-z0-9-]{0,31}" % arm, 1)
    if not d.get("trial", "").isdigit() or int(d["trial"]) < 1:
        die("inputs: trial must be a positive integer", 1)
    if not os.path.isabs(d.get("brief", "")) or not os.path.isfile(d["brief"]):
        die("inputs: brief must be the absolute path of a file", 1)
    v = d.get("variant", "")
    if arm == "base":
        if v:
            die("inputs: arm base takes no variant", 1)
    else:
        if not os.path.isabs(v) or not os.path.isdir(v):
            die("inputs: arm %s needs variant, the absolute path of a directory" % arm, 1)
        found = 0
        for name in os.listdir(v):
            if name not in SKILLS:
                die("inputs: variant holds %r, not a hyp skill" % name, 1)
            if not os.path.isfile(os.path.join(v, name, "SKILL.md")):
                die("inputs: variant %s has no SKILL.md" % name, 1)
            found += 1
        if not found:
            die("inputs: variant holds no skill", 1)
    return d


def skill_hashes(wt):
    """sha256 of every file under .claude/skills, by path relative to wt."""
    out = {}
    root = os.path.join(wt, ".claude", "skills")
    for dp, dn, fn in os.walk(root):
        for f in fn:
            p = os.path.join(dp, f)
            out[os.path.relpath(p, wt)] = sha256_file(p)
    return out


# ---- shell: the command whose exit status a Bash call returns -------------

def last_command(cmd):
    """Last simple command of a shell line (after ; && || | & and newlines)."""
    segs = []
    cur = []
    i, n = 0, len(cmd)
    paren = 0          # depth of $( ) and ( ) outside quotes
    heredocs = []      # pending delimiters
    while i < n:
        c = cmd[i]
        if c == "\\" and i + 1 < n:
            cur.append(cmd[i:i + 2]); i += 2; continue
        if c == "'":
            j = cmd.find("'", i + 1)
            j = n - 1 if j < 0 else j
            cur.append(cmd[i:j + 1]); i = j + 1; continue
        if c == '"':
            j, depth = i + 1, 0
            while j < n:
                if cmd[j] == "\\":
                    j += 2; continue
                if cmd.startswith("$(", j):
                    depth += 1; j += 2; continue
                if cmd[j] == ")" and depth:
                    depth -= 1
                elif cmd[j] == '"' and not depth:
                    break
                j += 1
            cur.append(cmd[i:j + 1]); i = j + 1; continue
        if c == "`":
            j = cmd.find("`", i + 1)
            j = n - 1 if j < 0 else j
            cur.append(cmd[i:j + 1]); i = j + 1; continue
        if cmd.startswith("$(", i):
            paren += 1; cur.append("$("); i += 2; continue
        if c == "(":
            paren += 1
        elif c == ")" and paren:
            paren -= 1
        if cmd.startswith("<<", i) and not cmd.startswith("<<<", i):
            m = re.compile(r"<<-?\s*(?:'([^']*)'|\"([^\"]*)\"|(\\?[A-Za-z_][A-Za-z0-9_]*))").match(cmd, i)
            if m:
                heredocs.append((m.group(1) or m.group(2) or m.group(3).lstrip("\\"), cmd[i + 2:i + 3] == "-"))
                cur.append(m.group(0)); i = m.end(); continue
        if c == "\n":
            if heredocs:
                for delim, strip in heredocs:
                    i += 1
                    while i < n:
                        e = cmd.find("\n", i)
                        e = n if e < 0 else e
                        line = cmd[i:e]
                        i = e
                        if (line.strip() if strip else line) == delim:
                            break
                        i += 1
                heredocs = []
                continue
            if not paren:
                segs.append("".join(cur)); cur = []
            else:
                cur.append(" ")
            i += 1; continue
        if not paren and c in ";&|":
            j = i + 1
            if j < n and cmd[j] == c and c != ";":
                j += 1
            if c == "&" and i + 1 < n and cmd[i + 1] == ">":
                cur.append(c); i += 1; continue
            if c == "&" and cur and cur[-1].endswith((">", "<")):
                cur.append(c); i += 1; continue
            if c == "|" and j == i + 1 and cur and cur[-1].endswith(">"):
                cur.append(c); i += 1; continue
            segs.append("".join(cur)); cur = []
            i = j; continue
        cur.append(c); i += 1
    segs.append("".join(cur))
    segs = [s.strip() for s in segs if s.strip()]
    return segs[-1] if segs else ""


WRAPPERS = {"time", "exec", "command", "env", "nice", "!", "builtin"}


def is_hyp(segment):
    s = segment.lstrip("({ \t")
    words = s.split()
    k = 0
    while k < len(words):
        w = words[k]
        if re.match(r"^[A-Za-z_][A-Za-z0-9_]*=", w) or w in WRAPPERS:
            k += 1; continue
        break
    if k >= len(words):
        return False
    w = words[k].strip("\"'")
    return os.path.basename(w) == "hyp"


# ---- the transcript -------------------------------------------------------

def read_jsonl(path):
    out = []
    with open(path) as f:
        for line in f:
            line = line.strip()
            if not line:
                continue
            try:
                out.append(json.loads(line))
            except ValueError:
                continue
    return out


def result_text(block):
    c = block.get("content")
    if isinstance(c, str):
        return c
    if isinstance(c, list):
        return "\n".join(x.get("text", "") for x in c if isinstance(x, dict))
    return ""


REF_PIN = re.compile(r"--decided-by reading contradicts recipe \S+, an instrument whose checks decide the verdicts")
BODY_LIM = re.compile(r"over limit: body has \d+ runes \(max 600\)")


def score_transcript(msgs):
    uses = {}
    skill_loaded = 0
    failing = 0
    f_pin = f_body = 0
    results = []
    for m in msgs:
        content = (m.get("message") or {}).get("content")
        if not isinstance(content, list):
            continue
        for b in content:
            if not isinstance(b, dict):
                continue
            if m.get("type") == "assistant" and b.get("type") == "tool_use":
                uses[b.get("id")] = b
                name, inp = b.get("name"), b.get("input") or {}
                if name == "Skill":
                    skill_loaded = 1
                elif name == "Read" and os.path.basename(str(inp.get("file_path", ""))) == "SKILL.md":
                    skill_loaded = 1
            elif m.get("type") == "user" and b.get("type") == "tool_result":
                results.append(b)
    for b in results:
        if not b.get("is_error"):
            continue
        text = result_text(b)
        if REF_PIN.search(text):
            f_pin = 1
        if BODY_LIM.search(text):
            f_body = 1
        use = uses.get(b.get("tool_use_id"))
        if use and use.get("name") == "Bash":
            if is_hyp(last_command(str((use.get("input") or {}).get("command", "")))):
                failing += 1
    return skill_loaded, failing, f_pin, f_body


def end_state(events_json, cut_ids):
    cut = set(cut_ids)
    new = [e for e in events_json.get("events", []) if e.get("id") not in cut]
    if any(e.get("type") == "result.record" for e in new):
        return "recorded"
    for e in new:
        if e.get("type") == "node.add" and (e.get("payload") or {}).get("kind") == "Experiment":
            return "registered"
    return "none"


# ---- the verdict ----------------------------------------------------------

def parse_score(path):
    d = {}
    with open(path) as f:
        for line in f:
            line = line.rstrip("\n")
            if not line:
                continue
            k, sep, v = line.partition("\t")
            if not sep:
                die("score %s: line without a tab: %r" % (path, line))
            d[k] = v
    return d


def verdict(rows, scores):
    arms = {}
    if any(s.get("arm") == "repro" for s in scores):
        die("verdict: repro is the verdict label of the reproduction, not an arm")
    for s in scores:
        arms.setdefault(s["arm"], []).append(s)
    if "base" not in arms:
        die("verdict: no base arm")
    for arm, ts in sorted(arms.items()):
        nums = [t["trial"] for t in ts]
        if len(ts) != 5 or len(set(nums)) != 5:
            die("verdict: arm %s needs exactly 5 trials with distinct trial numbers, has %s" % (arm, nums))
    if not rows:
        die("verdict: no row to judge")
    for r in rows:
        for s in scores:
            if s.get(r) not in ("0", "1"):
                die("verdict: row %s is not 0/1 in arm %s trial %s" % (r, s["arm"], s["trial"]))
    for s in scores:
        if s.get("skill_loaded") not in ("0", "1"):
            die("verdict: skill_loaded is not 0/1 in arm %s trial %s" % (s["arm"], s["trial"]))

    def k(arm, row):
        n = 0
        for t in arms[arm]:
            loaded = t["skill_loaded"] == "1"
            if arm == "base":
                n += 1 if (loaded and t[row] == "1") else 0
            else:
                n += 1 if (not loaded or t[row] == "1") else 0
        return n

    repro_rows = [r for r in rows if k("base", r) >= 3]
    out = ["verdict: repro=%s" % ("supports" if repro_rows else "contradicts")]
    for arm in sorted(a for a in arms if a != "base"):
        if not repro_rows:
            w = "inconclusive"
        elif all(k(arm, r) <= 1 for r in repro_rows):
            w = "supports"
        else:
            w = "contradicts"
        out.append("verdict: %s=%s" % (arm, w))
    return out
