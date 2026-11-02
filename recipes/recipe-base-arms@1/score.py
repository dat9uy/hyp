#!/usr/bin/env python3
"""score OUT_DIR: count one trial's stops, judge the bundle it left, and measure its copy.

OUT_DIR is hyp sandbox remove's, with state.env, drive.json and
transcript.jsonl kept from this recipe's tmp. Prints key<TAB>value lines:

  arm, trial          from state.env
  model, cost_usd     from drive.json
  failing_hyp         Bash calls that end in error with hyp as their last command
  f_sandbox_json      1 when the runner wrote a sandbox.json itself
  f_tar_input         1 when a known-answers input under recipes/ holds an archive
                      that method.md does not name
  f_go_file           1 when a bundle under recipes/ holds a .go or .go.txt file
  f_recipe_none       1 when the runner ran hyp add exp with --recipe none
  f_weak_cases        1 when hyp recipe check warns about the bundle's cases
  stops               failing_hyp plus every f_ value
  bundle              the ref judged, or none
  known_answers_exit  hyp recipe known-answers on the bundle, rebuilt from changes/
  add_exp_exit        hyp add exp --recipe on it, in a scratch tree
  end_state           pass when both exits are 0, else fail
  base_from           the bundle's from: lines that name a base (from: base:...)
  edited_lines        lines of the bundle's files that its named bases do not
                      hold: for each file, every non-blank line (stripped)
                      counted as often as it occurs beyond its count in the
                      base file at the same path, or in the bases' recipe.md
                      files together for recipe.md; a file no base holds, and
                      every file on a bundle with no base, counts whole.
                      Files under known-answers/*/input/ and files that are
                      not UTF-8 are not counted.
  unused_base_files   files of the named bases, recipe.md aside, that the
                      bundle does not hold at the same path
  seam                1 when a known-answers case input of the bundle holds a
                      transcript.jsonl, so a case runs the scorer on a trial
                      laid out as the runner's drive writes it

The bases are read from bases.tar in this bundle, the candidates the coarse
and fine arms embed, at <name>@<N>/<path>; a from: base the archive does
not hold has no files. A file is made executable when its changes.json mode (ls form, for
example -rwxr-xr-x) has an owner x. Every hyp call is the hyp first on PATH. Exits 2 when OUT_DIR lacks a kept
file (state.env, drive.json, transcript.jsonl, changes.json), and 0
otherwise.
"""
import json
import os
import re
import shutil
import subprocess
import sys
import tarfile
import tempfile
from collections import Counter

LAST_HYP = re.compile(r"^(\w+=\S*\s+)*(\S*/)?hyp\s+[a-z-]")
HEREDOC = re.compile(r"<<-?\s*['\"]?(\w+)['\"]?[^\n]*\n.*?\n\s*\1\s*(?=\n|$)", re.S)


def last_command(cmd):
    """The command whose exit status a Bash call returns: heredoc bodies dropped, then
    the last part after a newline, ;, && or ||, and after its last |."""
    parts = [p.strip() for p in re.split(r"\n|;|&&|\|\|", HEREDOC.sub("", cmd)) if p.strip()]
    return parts[-1].split("|")[-1].strip() if parts else ""


def method_names():
    """Base names of the files the bundle's own method.md names (an archive it asks for is no workaround)."""
    with open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "method.md")) as f:
        return set(re.findall(r"[\w.-]+\.(?:tar|tgz|tar\.gz)\b", f.read()))
WRITE_SANDBOX = re.compile(r"(>|\btee\s|\bcp\s|\bmv\s)[^;&|\n]*sandbox\.json")
ADD_EXP_NONE = re.compile(r"\bhyp\s+add\s+exp\b[^\n;&|]*--recipe[ =]none\b")
BUNDLE = re.compile(r"^recipes/([a-z0-9][a-z0-9._-]*)@([A-Za-z0-9._-]+)/")


def fail(msg):
    print("score: " + msg, file=sys.stderr)
    sys.exit(2)


def read_lines(path):
    with open(path) as f:
        return [json.loads(l) for l in f if l.strip()]


def blocks(lines, kind, role):
    for l in lines:
        content = (l.get("message") or {}).get("content")
        if l.get("type") == role and isinstance(content, list):
            for b in content:
                if b.get("type") == kind:
                    yield b


def run(argv, cwd):
    r = subprocess.run(argv, cwd=cwd, capture_output=True, text=True)
    return r.returncode, r.stdout + r.stderr


def lines_of(data):
    try:
        text = data.decode("utf-8")
    except UnicodeDecodeError:
        return None
    return Counter(l.strip() for l in text.splitlines() if l.strip())


def bases_archive():
    """{"<name>@<N>": {path: bytes}} from bases.tar beside this script."""
    out = {}
    with tarfile.open(os.path.join(os.path.dirname(os.path.abspath(__file__)), "bases.tar")) as t:
        for m in t.getmembers():
            if m.isfile():
                ref, _, rel = m.name.partition("/")
                out.setdefault(ref, {})[rel] = t.extractfile(m).read()
    return out


def measure(out, live, bundle):
    """base_from, edited_lines, unused_base_files and seam of the judged bundle."""
    if bundle == "none":
        return 0, 0, 0, 0
    prefix = f"recipes/{bundle}/"
    files = {}
    for c in live:
        if c["path"].startswith(prefix) and c.get("copied"):
            with open(os.path.join(out, "changes", c["path"]), "rb") as f:
                files[c["path"][len(prefix):]] = f.read()
    md = files.get("recipe.md", b"").decode("utf-8", "replace")
    fm = md.split("\n---", 1)[0] if md.startswith("---") else ""
    named = [l.split()[1][len("base:"):] for l in fm.splitlines()
             if l.startswith("from:") and len(l.split()) > 1 and l.split()[1].startswith("base:")]
    arch = bases_archive()
    base_files = {}
    base_md = Counter()
    for ref in named:
        for rel, data in arch.get(ref, {}).items():
            if rel == "recipe.md":
                base_md += lines_of(data) or Counter()
            else:
                base_files.setdefault(rel, data)
    edited = 0
    for rel, data in files.items():
        if re.match(r"known-answers/[^/]+/input/", rel):
            continue
        have = lines_of(data)
        if have is None:
            continue
        if rel == "recipe.md":
            base = base_md
        else:
            base = lines_of(base_files[rel]) or Counter() if rel in base_files else Counter()
        edited += sum((have - base).values())
    unused = sum(1 for rel in base_files if rel not in files)
    seam = int(any(re.match(r"known-answers/[^/]+/input/(.+/)?transcript\.jsonl$", rel) for rel in files))
    return len(named), edited, unused, seam


def main():
    if len(sys.argv) != 2:
        fail("usage: score OUT_DIR")
    out = sys.argv[1]
    for f in ("state.env", "drive.json", "transcript.jsonl", "changes.json"):
        if not os.path.isfile(os.path.join(out, f)):
            fail(f"{out} holds no {f}")
    st = dict(l.rstrip("\n").split("=", 1) for l in open(os.path.join(out, "state.env")) if "=" in l)
    lines = read_lines(os.path.join(out, "transcript.jsonl"))
    changes = json.load(open(os.path.join(out, "changes.json")))["changes"]
    drive = json.load(open(os.path.join(out, "drive.json")))

    errors = {b.get("tool_use_id") for b in blocks(lines, "tool_result", "user") if b.get("is_error")}
    seen, failing, sandbox_json, recipe_none = set(), 0, 0, 0
    for b in blocks(lines, "tool_use", "assistant"):
        if b.get("id") in seen:
            continue
        seen.add(b.get("id"))
        inp = b.get("input") or {}
        if b.get("name") == "Bash":
            cmd = inp.get("command", "")
            if LAST_HYP.match(last_command(cmd)) and b.get("id") in errors:
                failing += 1
            if WRITE_SANDBOX.search(cmd):
                sandbox_json = 1
            if ADD_EXP_NONE.search(cmd):
                recipe_none = 1
        elif b.get("name") in ("Write", "Edit") and os.path.basename(inp.get("file_path", "")) == "sandbox.json":
            sandbox_json = 1

    live = [c for c in changes if c.get("change") in ("added", "modified")]
    tar_input = int(any(BUNDLE.match(c["path"]) and "/known-answers/" in c["path"] and "/input/" in c["path"]
                        and re.search(r"\.(tar|tgz|tar\.gz)$", c["path"])
                        and os.path.basename(c["path"]) not in method_names() for c in live))
    go_file = int(any(BUNDLE.match(c["path"]) and re.search(r"\.go(\.txt)?$", c["path"]) for c in live))

    refs = sorted({"%s@%s" % BUNDLE.match(c["path"]).groups() for c in live
                   if BUNDLE.match(c["path"]) and c["path"].endswith("/recipe.md")})
    bundle, ka, ae, weak = "none", "none", "none", 0
    work = tempfile.mkdtemp()
    try:
        repo = os.path.join(work, "repo")
        os.makedirs(repo)
        run(["git", "init", "-q", "."], repo)
        for c in live:
            if BUNDLE.match(c["path"]) and c.get("copied"):
                dst = os.path.join(repo, c["path"])
                os.makedirs(os.path.dirname(dst), exist_ok=True)
                shutil.copyfile(os.path.join(out, "changes", c["path"]), dst)
                os.chmod(dst, 0o755 if "x" in c.get("mode", "-rw-r--r--")[1:4] else 0o644)
        os.environ["HYP_ACTOR"] = "score"
        run(["hyp", "init"], repo)
        g = run(["hyp", "add", "goal", "--title", "g", "--why", "w"], repo)[1].split()[0]
        q = run(["hyp", "add", "dq", "--parent", g, "--title", "q", "--why", "w"], repo)[1].split()[0]
        h = run(["hyp", "add", "hyp", "--parent", q, "--title", "h", "--why", "w"], repo)[1].split()[0]
        for ref in refs:
            k, _ = run(["hyp", "recipe", "known-answers", ref], repo)
            a, _ = run(["hyp", "add", "exp", "--parent", h, "--title", "e", "--recipe", ref,
                        "--target", "none", "--why", "w"], repo)
            _, chk = run(["hyp", "recipe", "check"], repo)
            w = int(any(l.startswith("warn") and ref in l for l in chk.splitlines()))
            if bundle == "none" or (k == 0 and a == 0):
                bundle, ka, ae, weak = ref, k, a, w
            if k == 0 and a == 0:
                break
    finally:
        shutil.rmtree(work, ignore_errors=True)

    base_from, edited, unused, seam = measure(out, live, bundle)
    stops = failing + sandbox_json + tar_input + go_file + recipe_none + weak
    rows = [("arm", st.get("arm", "")), ("trial", st.get("trial", "")), ("model", drive.get("model")),
            ("cost_usd", drive.get("cost_usd")), ("failing_hyp", failing),
            ("f_sandbox_json", sandbox_json), ("f_tar_input", tar_input), ("f_go_file", go_file),
            ("f_recipe_none", recipe_none), ("f_weak_cases", weak), ("stops", stops), ("bundle", bundle),
            ("known_answers_exit", ka), ("add_exp_exit", ae),
            ("end_state", "pass" if ka == 0 and ae == 0 else "fail"), ("base_from", base_from),
            ("edited_lines", edited), ("unused_base_files", unused), ("seam", seam)]
    for k, v in rows:
        print(f"{k}\t{v}")


if __name__ == "__main__":
    main()
