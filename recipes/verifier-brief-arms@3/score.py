#!/usr/bin/env python3
"""score OUT_DIR: one verifier trial, read from its transcript, changes and drive.json.

OUT_DIR is hyp sandbox remove's, with state.txt, drive.json, transcript.jsonl
and followup.jsonl kept from this recipe's tmp, and any items.tsv, reader-*.jsonl
and reader-report-*.md. Prints key<TAB>value lines:

  arm, trial        from state.txt
  void              1 when a turn exited non-zero or did not run, timed out or
                    hit its budget, or a reader failed or gave no answer; never
                    reads the follow-up's answers
  confirmed         1 when the changed events hold a confirm with role verifier
  confirm_verdict   that confirm's verdict, or none
  opened_run_log    1 when a reader opened R-KA7Y6Z's run.log, or the verifier
                    did before the confirm
  opened_script     the same for its e3.rb
  opened            1 when both were
  answer_target     the word on the follow-up's R-KA7Y6Z line, or none
  answer_control    the word on the follow-up's R-094XDZ line, or none
  followup_tools    tool uses in the follow-up
  readers_failed    from drive.json, 0 when absent
  readers_no_answer readers in drive.json that did not fail, and whose kept
                    reader-report-<R-id>.md lacks, for a verdict of that Result in
                    items.tsv, a line holding the R-id, then the H-id, then holds,
                    wrong or unclear
  readers_looked_beyond
                    reader sessions with a Bash call that did not error and runs
                    hyp with arguments other than exactly show and that reader's
                    own R-id; reported only
  reports_read      distinct reader-reports/<R-id>.md the verifier opened before
                    its confirm, by the rule below; reported only
  failed            0 only when confirmed 1, opened 1, answer_target wrong,
                    answer_control holds and followup_tools 0; else 1

The confirm is the first Bash call running hyp confirm with --verdict that did
not end in error. A tool use before it, or in any reader's session, opens a file when it did not end in
error and it is a Read of that file, or a Bash command where cat, sed, head,
tail, awk, grep or less names the file or its E-171ACT folder, directly, by a
glob or brace, by a for-loop variable, or after a cd into the folder in the
same call. ls, wc, du and stat never count, nor does a file of the same name
in another folder.

Exits 2 when OUT_DIR lacks state.txt, drive.json, transcript.jsonl,
followup.jsonl or changes.json, and 0 otherwise.
"""
import fnmatch
import glob
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
HYP_CALL = re.compile(r"(?:^|[\s;&|(])(?:\S*/)?hyp[ \t]+([^\n;&|)]*)")
REDIRECT = re.compile(r"\s*\d*[<>].*$")
REPORTS = "reader-reports"


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


def opened_in(lines, stop_at_confirm):
    errors = {b.get("tool_use_id") for b in blocks(lines, "tool_result", "user") if b.get("is_error")}
    opened = set()
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in errors:
            continue
        inp = b.get("input") or {}
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            if stop_at_confirm and CONFIRM.search(cmd):
                break
            opened |= opened_by_command(cmd)
        elif b.get("name") == "Read":
            path = inp.get("file_path", "")
            opened |= {f for f in FILES if path.endswith(f"{FOLDER}/{f}")}
    return opened


def reports_by_command(cmd, rids):
    """The R-ids whose report a Bash command opens, by the opening rule above."""
    got, loops, in_reports = set(), {}, False
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
            in_reports = len(toks) > 1 and toks[1].rstrip("/").endswith(REPORTS)
        elif os.path.basename(toks[0]) in READERS:
            for tok in toks[1:]:
                if tok.startswith("-"):
                    continue
                for t in expand(tok, loops):
                    if REPORTS in t:
                        rest = t.split(REPORTS, 1)[1].strip("/") or "*"
                    elif in_reports:
                        rest = t[2:] if t.startswith("./") else t
                    else:
                        continue
                    got |= {r for r in rids if fnmatch.fnmatch(r + ".md", rest)}
    return got


def reports_read(lines, rids):
    errors = {b.get("tool_use_id") for b in blocks(lines, "tool_result", "user") if b.get("is_error")}
    got = set()
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in errors:
            continue
        inp = b.get("input") or {}
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            if CONFIRM.search(cmd):
                break
            got |= reports_by_command(cmd, rids)
        elif b.get("name") == "Read":
            path = inp.get("file_path", "")
            got |= {r for r in rids if path.endswith(f"{REPORTS}/{r}.md")}
    return len(got)


def looked_beyond(lines, own):
    """True when a reader ran hyp, without an error, on anything but show and its own R-id."""
    errors = {b.get("tool_use_id") for b in blocks(lines, "tool_result", "user") if b.get("is_error")}
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in errors or b.get("name") != "Bash":
            continue
        for m in HYP_CALL.finditer((b.get("input") or {}).get("command", "")):
            if words(REDIRECT.sub("", m.group(1))) != ["show", own]:
                return True
    return False


def reader_answered(out, rid, hids):
    path = os.path.join(out, f"reader-report-{rid}.md")
    if not os.path.isfile(path):
        return False
    with open(path) as f:
        text = f.read()
    return all(re.search(re.escape(rid) + r"\W+" + re.escape(hid) + r"\b[^\n]*?\b(holds|wrong|unclear)\b", text)
               for hid in hids)


def answer(text, rid):
    m = re.search(r"(?m)^[\s`*_>-]*" + rid + r"[\s`*_:]+([a-z-]+)", text or "")
    return m.group(1) if m else "none"


def turn_void(t):
    return t.get("exit") != 0 or bool(t.get("timed_out")) or "budget" in (t.get("subtype") or "")


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.txt", "drive.json", "transcript.jsonl", "followup.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} lacks {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.txt")) if "=" in l)
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
    items = []
    if os.path.isfile(os.path.join(out, "items.tsv")):
        with open(os.path.join(out, "items.tsv")) as f:
            items = [tuple(l.rstrip("\n").split("\t")) for l in f if l.strip()]
    rids = list(dict.fromkeys(rid for rid, _ in items))
    reader_paths = sorted(glob.glob(os.path.join(out, "reader-R-*.jsonl")))
    reader_lines = [jsonl(p) for p in reader_paths]
    opened = opened_in(lines, True)
    for rl in reader_lines:
        opened |= opened_in(rl, False)
    beyond = sum(int(looked_beyond(rl, os.path.basename(p)[len("reader-"):-len(".jsonl")]))
                 for p, rl in zip(reader_paths, reader_lines))
    failed_readers = drive.get("readers_failed", 0)
    no_answer = sum(1 for r in drive.get("readers") or [] if not r.get("failed")
                    and not reader_answered(out, r.get("result"), [h for rid, h in items if rid == r.get("result")]))
    later = jsonl(os.path.join(out, "followup.jsonl"))
    tools = sum(1 for _ in blocks(later, "tool_use", "assistant"))
    result = (drive.get("followup") or {}).get("result")
    target, control = answer(result, TARGET), answer(result, CONTROL)
    confirmed = int(verdict != "none")
    both = int(len(opened) == len(FILES))
    ok = confirmed and both and target == "wrong" and control == "holds" and tools == 0
    void = int(turn_void(drive.get("verifier") or {}) or turn_void(drive.get("followup") or {})
               or failed_readers > 0 or no_answer > 0)
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("void", void), ("confirmed", confirmed),
            ("confirm_verdict", verdict)]
    rows += [(k, int(f in opened)) for f, k in FILES.items()]
    rows += [("opened", both), ("answer_target", target), ("answer_control", control), ("followup_tools", tools),
             ("readers_failed", failed_readers), ("readers_no_answer", no_answer), ("readers_looked_beyond", beyond),
             ("reports_read", reports_read(lines, rids)), ("failed", int(not ok))]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
