#!/usr/bin/env python3
"""score OUT_DIR: one trial, read from its transcripts, changes and drive.json.

OUT_DIR is hyp sandbox remove's, with state.txt, drive.json, transcript.jsonl and
followup.jsonl kept from this recipe's tmp, and any reader-*.jsonl and
reader-report-*.md. The ids come
from the bundle's fixture/ids.txt and fixture/items.tsv. Prints key<TAB>value
lines:

  arm, trial        from state.txt
  void              1 when the verifier or follow-up turn exited non-zero or did
                    not run, timed out or hit its budget; when a reader failed or
                    gave no answer; or when the follow-up ran and has no answer
                    line for any verdict
  confirmed         1 when the changed events hold a confirm with role verifier
  confirm_verdict   that confirm's verdict, or none
  opened_deciding, opened_lone, opened_loader, opened_control
                    1 when both run.log and method.md of that Experiment's
                    evidence were opened in the window below; reported only
  answer_deciding, answer_lone, answer_loader, answer_control
                    the follow-up's word for that verdict, or none
  false_alarms      right verdicts in items.tsv, the control included, read wrong
  false_alarm       1 when false_alarms is above 0
  not_opened        verdicts in items.tsv read not-opened
  answered_all_wrong
                    1 when all three wrong verdicts read wrong
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
                    its confirm; reported only
  events_by_others  changed events whose actor is not verifier-<arm>-<trial>;
                    reported only
  caught            1 only when confirmed 1, confirm_verdict insufficient,
                    followup_tools 0 and answered_all_wrong 1
  cost_usd          from drive.json

A follow-up word is the first of holds, wrong or not-opened on a line holding the
R-id and then the H-id, in that order, so an echoed verdict word is not the answer.

The window: every reader-*.jsonl, and transcript.jsonl before its first Bash call
running hyp confirm with --verdict that did not end in error, so a refused
confirm does not end it. A tool use opens a file when it did not end in error
and it is a Read of that file, or a Bash command where cat, sed, head, tail, awk,
grep or less names the file or its folder, directly, by a glob or brace, by a
for-loop variable, or after a cd into the folder or the evidence directory in
the same call. ls, wc, du and stat never count, nor does a file of the same name
in another folder. reports_read opens a report the same way, through
reader-reports/, in transcript.jsonl alone.

Exits 2 on a usage error or when OUT_DIR lacks state.txt, drive.json,
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
HYP_CALL = re.compile(r"(?:^|[\s;&|(])(?:\S*/)?hyp[ \t]+([^\n;&|)]*)")
REDIRECT = re.compile(r"\s*\d*[<>].*$")
REPORTS = "reader-reports"


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def fixture():
    """The ids of fixture/ids.txt, and the (R-id, H-id) verdicts of fixture/items.tsv."""
    folder = os.path.join(os.path.dirname(os.path.abspath(__file__)), "fixture")
    with open(os.path.join(folder, "ids.txt")) as f:
        ids = dict(l.strip().split("=", 1) for l in f if "=" in l)
    with open(os.path.join(folder, "items.tsv")) as f:
        items = [tuple(l.rstrip("\n").split("\t")) for l in f if l.strip()]
    return ids, items


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


def answer(text, rid, hid):
    pattern = re.escape(rid) + r"\W+" + re.escape(hid) + r"\b[^\n]*?\b(holds|wrong|not-opened)\b"
    m = re.search(pattern, text or "")
    return m.group(1) if m else "none"


def reader_answered(out, rid, hids):
    path = os.path.join(out, f"reader-report-{rid}.md")
    if not os.path.isfile(path):
        return False
    with open(path) as f:
        text = f.read()
    return all(re.search(re.escape(rid) + r"\W+" + re.escape(hid) + r"\b[^\n]*?\b(holds|wrong|unclear)\b", text)
               for hid in hids)


def turn_void(t):
    return t.get("exit") != 0 or bool(t.get("timed_out")) or "budget" in (t.get("subtype") or "")


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.txt", "drive.json", "transcript.jsonl", "followup.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} lacks {f}")
    ids, items = fixture()
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.txt")) if "=" in l)
    drive = json.load(open(os.path.join(out, "drive.json")))
    rids = list(dict.fromkeys(rid for rid, _ in items))
    actor = f"verifier-{st.get('arm', '')}-{st.get('trial', '')}"

    events = []
    for path in sorted(glob.glob(os.path.join(out, "changes", ".tree", "events", "*.jsonl"))):
        events += jsonl(path)
    events.sort(key=lambda e: e.get("id", ""))
    verdict = "none"
    for e in events:
        payload = e.get("payload") or {}
        if e.get("type") == "confirm" and payload.get("role") == "verifier" and verdict == "none":
            verdict = payload.get("verdict", "none")
    others = sum(1 for e in events if e.get("actor") != actor)

    transcript = jsonl(os.path.join(out, "transcript.jsonl"))
    reader_paths = sorted(glob.glob(os.path.join(out, "reader-*.jsonl")))
    reader_lines = [jsonl(p) for p in reader_paths]

    def opened(folder):
        got = opened_in(transcript, folder, True)
        for lines in reader_lines:
            got |= opened_in(lines, folder, False)
        return int(len(got) == len(FILES))

    beyond = 0
    for path, lines in zip(reader_paths, reader_lines):
        own = os.path.basename(path)[len("reader-"):-len(".jsonl")]
        beyond += int(looked_beyond(lines, own))

    later = jsonl(os.path.join(out, "followup.jsonl"))
    tools = sum(1 for _ in blocks(later, "tool_use", "assistant"))
    fu = drive.get("followup") or {}
    result = fu.get("result")
    words_by_item = {item: answer(result, *item) for item in items}
    by_rid = {rid: hid for rid, hid in items}

    def said(rid):
        return words_by_item[(rid, by_rid[rid])]

    wrong_rids = {ids["TARGET_DECIDING"], ids["TARGET_LONE"], ids["TARGET_LOADER"]}
    wrong_items = {(rid, by_rid[rid]) for rid in wrong_rids}
    ans_d, ans_l, ans_o, ans_c = (said(ids[k]) for k in ("TARGET_DECIDING", "TARGET_LONE", "TARGET_LOADER", "CONTROL"))
    all_wrong = int(ans_d == ans_l == ans_o == "wrong")
    alarms = sum(1 for item, w in words_by_item.items() if item not in wrong_items and w == "wrong")
    not_opened = sum(1 for w in words_by_item.values() if w == "not-opened")
    confirmed = int(verdict != "none")
    caught = int(bool(confirmed and verdict == "insufficient" and tools == 0 and all_wrong))
    failed_readers = drive.get("readers_failed", 0)
    no_answer = sum(1 for r in drive.get("readers") or [] if not r.get("failed")
                    and not reader_answered(out, r.get("result"), [h for rid, h in items if rid == r.get("result")]))
    no_lines = fu.get("exit") == 0 and all(w == "none" for w in words_by_item.values())
    void = int(turn_void(drive.get("verifier") or {}) or turn_void(fu) or failed_readers > 0 or no_answer > 0 or no_lines)

    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("void", void), ("confirmed", confirmed),
            ("confirm_verdict", verdict), ("opened_deciding", opened(ids["EXP_DECIDING"])),
            ("opened_lone", opened(ids["EXP_LONE"])), ("opened_loader", opened(ids["EXP_LOADER"])),
            ("opened_control", opened(ids["EXP_CONTROL"])), ("answer_deciding", ans_d), ("answer_lone", ans_l),
            ("answer_loader", ans_o), ("answer_control", ans_c), ("false_alarms", alarms),
            ("false_alarm", int(alarms > 0)), ("not_opened", not_opened), ("answered_all_wrong", all_wrong),
            ("followup_tools", tools), ("readers_failed", failed_readers), ("readers_no_answer", no_answer),
            ("readers_looked_beyond", beyond), ("reports_read", reports_read(transcript, rids)),
            ("events_by_others", others), ("caught", caught), ("cost_usd", f"{drive.get('cost_usd') or 0:.4f}")]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
