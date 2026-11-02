#!/usr/bin/env python3
"""Run every case of CASES with every arm, each on a fresh copy of FIXTURE.

usage: cases.py CASES FIXTURE ARM=/abs/path...

CASES is a TSV with a header and rows `case  class  steps`, where steps is a
JSON list of argv lists (hyp arguments without the binary), run in order as
`step1 && step2`. FIXTURE is a tar holding fixture/ (the store the steps run
in) and, optionally, other entries. Each ARM is NAME=PATH, PATH an absolute
executable; this script never runs the hyp on PATH.

For each arm and case it extracts FIXTURE to a new directory, runs the steps
in fixture/ with HYP_ACTOR=arm-<name>, and prints one row:

    arm  case  exit  appended  warned

for the last step that ran: exit is its exit code; appended is y when the
bytes under fixture/.tree/events changed during that step; warned is y when
its stderr holds a line starting with "warn:". A case whose last step exits
3 is run once more from a new copy, and the second row is printed. Each
step's stderr is kept under $HYP_OUT/<arm>/<case>.<n>.stderr.

Exit 0, or 2 for bad arguments.
"""
import hashlib
import json
import os
import subprocess
import sys
import tarfile
import tempfile


def die(msg):
    print(f"cases.py: {msg}", file=sys.stderr)
    sys.exit(2)


def events_digest(store):
    h = hashlib.sha256()
    root = os.path.join(store, ".tree", "events")
    for dirpath, dirs, files in os.walk(root):
        dirs.sort()
        for f in sorted(files):
            p = os.path.join(dirpath, f)
            h.update(os.path.relpath(p, root).encode() + b"\0")
            with open(p, "rb") as fh:
                h.update(hashlib.sha256(fh.read()).digest())
    return h.hexdigest()


def run_case(arm, path, name, steps, fixture, out):
    with tempfile.TemporaryDirectory() as tmp:
        with tarfile.open(fixture) as t:
            t.extractall(tmp, filter="data")
        store = os.path.join(tmp, "fixture")
        env = dict(os.environ, HYP_ACTOR=f"arm-{arm}")
        row = None
        for n, argv in enumerate(steps, 1):
            before = events_digest(store)
            p = subprocess.run([path, *argv], cwd=store, env=env, stdin=subprocess.DEVNULL,
                               capture_output=True, text=True)
            appended = "y" if events_digest(store) != before else "n"
            warned = "y" if any(l.startswith("warn:") for l in p.stderr.splitlines()) else "n"
            os.makedirs(os.path.join(out, arm), exist_ok=True)
            with open(os.path.join(out, arm, f"{name}.{n}.stderr"), "w") as f:
                f.write(p.stderr)
            row = (p.returncode, appended, warned)
            if p.returncode != 0:
                break
        return row


def main(argv):
    if len(argv) < 3:
        die("usage: cases.py CASES FIXTURE ARM=/abs/path...")
    cases_path, fixture, specs = argv[0], argv[1], argv[2:]
    for p in (cases_path, fixture):
        if not os.path.isabs(p) or not os.path.isfile(p):
            die(f"{p}: want an absolute path to a file")
    arms = []
    for s in specs:
        name, sep, path = s.partition("=")
        if not sep or not name or not os.path.isabs(path) or not os.access(path, os.X_OK):
            die(f"{s}: want NAME=/abs/path to an executable")
        arms.append((name, path))
    with open(cases_path) as f:
        lines = f.read().splitlines()
    if not lines or lines[0].split("\t") != ["case", "class", "steps"]:
        die(f"{cases_path}: want header case, class, steps")
    cases = []
    for l in lines[1:]:
        cols = l.split("\t")
        if len(cols) != 3:
            die(f"{cases_path}: bad row {l!r}")
        cases.append((cols[0], json.loads(cols[2])))
    out = os.environ.get("HYP_OUT") or die("HYP_OUT is not set")
    print("arm\tcase\texit\tappended\twarned")
    for arm, path in arms:
        for name, steps in cases:
            row = run_case(arm, path, name, steps, fixture, out)
            if row[0] == 3:
                row = run_case(arm, path, name, steps, fixture, out)
            print(f"{arm}\t{name}\t{row[0]}\t{row[1]}\t{row[2]}")


if __name__ == "__main__":
    main(sys.argv[1:])
