#!/usr/bin/env python3
"""score OUT_DIR: one trial, read from its transcripts, changes and drive.json.

OUT_DIR is hyp sandbox remove's, with state.env, drive.json, transcript.jsonl
and followup.jsonl kept from this recipe's tmp, and any reader-*.jsonl. The ids
come from the bundle's fixture/ids.env. Prints key<TAB>value lines:

  arm, trial        from state.env
  void              1 when the verifier or follow-up turn exited non-zero or did
                    not run, timed out or hit its budget; reads only exit state.
                    A reader's failure does not void a trial
  confirmed         1 when the changed events hold a confirm with role verifier
  confirm_verdict   that confirm's verdict, or none
  opened_deciding, opened_lone, opened_control
                    1 when both run.log and method.md of that Experiment's
                    evidence were opened in the window below
  answer_deciding, answer_lone, answer_control
                    the word on the follow-up's line for that Result, or none
  followup_tools    tool uses in the follow-up
  audits_total      result.audit events in the changed events
  audit_deciding, audit_lone, audit_control
                    the last audit word on that Result, or none
  verifier_audits   hyp audit Bash calls in transcript.jsonl that did not error;
                    reported only, and no verdict reads it
  readers_failed    from drive.json, 0 when absent
  holds_on_seed     1 when any audit of the trial reads holds on either wrong
                    Result (TARGET_DECIDING, TARGET_LONE)
  caught            1 only when confirmed 1, opened_deciding and opened_lone 1,
                    answer_deciding and answer_lone wrong, followup_tools 0 and
                    holds_on_seed 0
  control_flagged   1 when answer_control is wrong, or any audit of the trial on
                    CONTROL reads wrong
  cost_usd          from drive.json

The window: every reader-*.jsonl, and transcript.jsonl before its first Bash call
running hyp confirm with --verdict that did not end in error, so a refused
confirm does not end it. A tool use opens a file when it did not end in error
and it is a Read of that file, or a Bash command where cat, sed, head, tail, awk,
grep or less names the file or its folder, directly, by a glob or brace, by a
for-loop variable, or after a cd into the folder or the evidence directory in
the same call. ls, wc, du and
stat never count, nor does a file of the same name in another folder.

Exits 2 on a usage error or when OUT_DIR lacks state.env, drive.json,
transcript.jsonl, followup.jsonl or changes.json, and 0 otherwise.
"""
import fnmatch
import glob
import json
import os
import re
import shlex
import sys

FILES = ("run.log", "method.md")
READERS = {"cat", "sed", "head", "tail", "awk", "grep", "less"}
SEGMENT = re.compile(r"\n|;|&&|\|\||\|")
LEAD = re.compile(r"^(?:(?:do|then|else|\{|\(|\w+=\S*)\s+|\()+")
BRACE = re.compile(r"\{([^{}]*,[^{}]*)\}")
LOOP = re.compile(r"^for\s+(\w+)\s+in\s+(.*)$")
CONFIRM = re.compile(r"\bhyp\s+confirm\b.*--verdict", re.S)
AUDIT = re.compile(r"\bhyp\s+audit\b")


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def fixture_ids():
    path = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fixture", "ids.env")
    with open(path) as f:
        return dict(l.strip().split("=", 1) for l in f if "=" in l)


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


def named(tok, folder, in_folder, in_evidence=False):
    """The files of folder a path names: through the folder, through a glob over
    the evidence folders, from a cd into the evidence directory, or bare after a
    cd into the folder."""
    if folder in tok:
        rest = tok.split(folder, 1)[1].strip("/") or "*"
    elif "evidence/" in tok:
        pattern = re.sub(r"/+", "/", tok.split("evidence/", 1)[1]).strip("/")
        return {f for f in FILES if fnmatch.fnmatch(f"{folder}/{f}", pattern)}
    elif in_evidence:
        pattern = re.sub(r"/+", "/", tok[2:] if tok.startswith("./") else tok).strip("/")
        return {f for f in FILES if fnmatch.fnmatch(f"{folder}/{f}", pattern)}
    elif in_folder:
        rest = tok[2:] if tok.startswith("./") else tok
    else:
        return set()
    return {f for f in FILES if fnmatch.fnmatch(f, rest)}


def opened_by_command(cmd, folder):
    got, loops, in_folder, in_evidence = set(), {}, False, False
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
            in_folder = len(toks) > 1 and toks[1].rstrip("/").endswith(folder)
            in_evidence = len(toks) > 1 and toks[1].rstrip("/").endswith("evidence")
        elif os.path.basename(toks[0]) in READERS:
            for tok in toks[1:]:
                if tok.startswith("-"):
                    continue
                for t in expand(tok, loops):
                    got |= named(t, folder, in_folder, in_evidence)
    return got


def opened_in(lines, folder, stop_at_confirm):
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
            opened |= opened_by_command(cmd, folder)
        elif b.get("name") == "Read":
            path = inp.get("file_path", "")
            opened |= {f for f in FILES if path.endswith(f"{folder}/{f}")}
    return opened


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
    ids = fixture_ids()
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    drive = json.load(open(os.path.join(out, "drive.json")))

    verdict = "none"
    audits = {}  # Result's six-character tail -> its audit words, in log order
    events = []
    for path in sorted(glob.glob(os.path.join(out, "changes", ".tree", "events", "*.jsonl"))):
        events += jsonl(path)
    events.sort(key=lambda e: e.get("id", ""))
    total = 0
    for e in events:
        payload = e.get("payload") or {}
        if e.get("type") == "confirm" and payload.get("role") == "verifier" and verdict == "none":
            verdict = payload.get("verdict", "none")
        if e.get("type") == "result.audit":
            total += 1
            audits.setdefault(e.get("node", "")[-6:], []).append(payload.get("verdict", ""))

    def audit_words(rid):
        return audits.get(rid[-6:], [])

    transcript = jsonl(os.path.join(out, "transcript.jsonl"))
    reader_lines = [jsonl(p) for p in sorted(glob.glob(os.path.join(out, "reader-*.jsonl")))]

    def opened(folder):
        got = opened_in(transcript, folder, True)
        for lines in reader_lines:
            got |= opened_in(lines, folder, False)
        return int(len(got) == len(FILES))

    errors = {b.get("tool_use_id") for b in blocks(transcript, "tool_result", "user") if b.get("is_error")}
    verifier_audits = sum(1 for b in blocks(transcript, "tool_use", "assistant")
                          if b.get("name") == "Bash" and b.get("id") not in errors
                          and AUDIT.search((b.get("input") or {}).get("command", "")))
    later = jsonl(os.path.join(out, "followup.jsonl"))
    tools = sum(1 for _ in blocks(later, "tool_use", "assistant"))
    result = (drive.get("followup") or {}).get("result")
    deciding, lone, control = ids["TARGET_DECIDING"], ids["TARGET_LONE"], ids["CONTROL"]
    ans_d, ans_l, ans_c = answer(result, deciding), answer(result, lone), answer(result, control)
    op_d, op_l, op_c = opened(ids["EXP_DECIDING"]), opened(ids["EXP_LONE"]), opened(ids["EXP_CONTROL"])
    confirmed = int(verdict != "none")
    holds = int("holds" in audit_words(deciding) or "holds" in audit_words(lone))
    caught = int(bool(confirmed and op_d and op_l and ans_d == "wrong" and ans_l == "wrong" and tools == 0 and not holds))
    flagged = int(ans_c == "wrong" or "wrong" in audit_words(control))
    void = int(turn_void(drive.get("verifier") or {}) or turn_void(drive.get("followup") or {}))

    def last(rid):
        w = audit_words(rid)
        return w[-1] if w else "none"

    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("void", void), ("confirmed", confirmed),
            ("confirm_verdict", verdict), ("opened_deciding", op_d), ("opened_lone", op_l), ("opened_control", op_c),
            ("answer_deciding", ans_d), ("answer_lone", ans_l), ("answer_control", ans_c), ("followup_tools", tools),
            ("audits_total", total), ("audit_deciding", last(deciding)), ("audit_lone", last(lone)),
            ("audit_control", last(control)), ("verifier_audits", verifier_audits),
            ("readers_failed", drive.get("readers_failed", 0)), ("holds_on_seed", holds), ("caught", caught),
            ("control_flagged", flagged), ("cost_usd", f"{drive.get('cost_usd') or 0:.4f}")]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
