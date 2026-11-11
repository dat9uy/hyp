#!/usr/bin/env python3
"""One removal per rule from a commit, each in a fresh detached worktree.

usage: remove.py REPO COMMIT OUTDIR
Writes OUTDIR/<rule>.diff and OUTDIR/<rule>.log for each rule.
"""
import os, subprocess, sys, shutil

REPO, COMMIT, OUT = sys.argv[1], sys.argv[2], sys.argv[3]
SCRATCH = os.path.dirname(os.path.abspath(__file__))
os.makedirs(OUT, exist_ok=True)

SUBJ = "internal/cli/subject.go"
VAL = "internal/replay/validate.go"
RULES = "internal/replay/rules.go"

EDITS = {
    # R3a: merge every run into one subject.
    "R3a": (SUBJ, "\t\tout = append(out, run)\n",
            "\t\tif len(out) == 0 {\n\t\t\tout = append(out, run)\n\t\t} else {\n\t\t\tout[0].Files = append(out[0].Files, run.Files...)\n\t\t\tslices.Sort(out[0].Files)\n\t\t\tout[0].Files = slices.Compact(out[0].Files)\n\t\t}\n",
            ["TestResultKeepsOneSubjectPerCitedRun"]),
    # R3b: drop the base compare.
    "R3b": (VAL, "(run.Base != sub.Base || (run.Patch", "((run.Patch",
            ["TestBaseSubjectIsCoveredOnlyAtItsBase", "TestHoldsSubject"]),
    # R3c: drop the old fingerprints read.
    "R3c": (RULES, "Subjects: runSubjects(p)})", "Subjects: p.Subjects})",
            ["TestOldResultReadsAsOneSubjectWithNoBase"]),
    # R3d: drop the patch compare, so a patched run holds bare B.
    "R3d": (VAL, "run.Base != sub.Base || (run.Patch != \"\" && !slices.Contains(want, run.Patch))", "run.Base != sub.Base",
            ["TestPatchedRunNeverCoversBareBase", "TestPatchedRunWithInputsNeverCoversBareBase", "TestHoldsSubject"]),
    # R7: drop the patch's own field, so the patch is kept among the inputs.
    "R7": (SUBJ, "add(base, p.SHA256, inputs)", "add(base, \"\", append([]string{p.SHA256}, inputs...))",
           ["TestPatchedRunWithInputsNeverCoversBareBase", "TestPatchedRunNeverCoversBareBase", "TestResultKeepsOneSubjectPerCitedRun", "TestSandboxKeepsOneSubjectPerArm"]),
    # R7 also: restore 3ec4979's exact-files compare for a based subject.
    "R7-restore": (VAL, "\t\tfor _, h := range want {\n",
                   "\t\tif sub.Base != \"\" {\n\t\t\tfor _, f := range run.Files {\n\t\t\t\tif !slices.Contains(want, f) {\n\t\t\t\t\treturn false\n\t\t\t\t}\n\t\t\t}\n\t\t}\n\t\tfor _, h := range want {\n",
                   ["TestRunWithInputsCoversBareBase", "TestHoldsSubject"]),
    # F1: merge every Arm into one subject.
    "F1": (SUBJ, "\t\t\tif len(r.Patches) == 0 {\n\t\t\t\tadd(base, \"\", inputs)\n\t\t\t}\n\t\t\tfor _, p := range r.Patches {\n\t\t\t\tadd(base, p.SHA256, inputs)\n\t\t\t}\n",
           "\t\t\tfiles := inputs\n\t\t\tfor _, p := range r.Patches {\n\t\t\t\tfiles = append(files, p.SHA256)\n\t\t\t}\n\t\t\tadd(base, \"\", files)\n",
           ["TestSandboxKeepsOneSubjectPerArm"]),
    # F3: drop the subject-ran lock.
    "F3": (VAL, "if r := f.subjectRan(ev.Node); r != \"\" {", "if r := f.subjectRan(ev.Node); false && r != \"\" {",
           ["TestBaseSubjectLocksOnlyWhenItRan"]),
    # F4: compare a based subject's files to the run's as lists.
    "F4": (VAL, "\t\tfor _, h := range want {\n",
           "\t\tif sub.Base != \"\" {\n\t\t\tgiven := append(slices.Clone(run.Files), run.Patch)\n\t\t\tslices.Sort(given)\n\t\t\tw := slices.Clone(want)\n\t\t\tslices.Sort(w)\n\t\t\treturn slices.Equal(given, w)\n\t\t}\n\t\tfor _, h := range want {\n",
           ["TestBaseSubjectWithOneFileTwiceIsHeld"]),
    # F5: drop the git-missing exit 3.
    "F5": (SUBJ, "\tif errors.Is(err, exec.ErrNotFound) {\n\t\treturn \"\", herr.Env(\"--base %s: git is not on PATH\", rev)\n\t}\n",
           "\t_ = errors.Is\n\t_ = herr.Env\n",
           ["TestSubjectBaseExitCodes"]),
    # F6: drop the no-evidence rejection.
    "F6": (VAL, "if len(p.Subjects) > 0 && len(p.Evidence) == 0 {", "if false && len(p.Subjects) > 0 && len(p.Evidence) == 0 {",
           ["TestRunSubjectReplayRejections"]),
}


def sh(args, cwd, **kw):
    return subprocess.run(args, cwd=cwd, capture_output=True, text=True, **kw)


for rule, (path, old, new, tests) in EDITS.items():
    wt = os.path.join(SCRATCH, "wt-" + rule)
    if os.path.exists(wt):
        sh(["git", "worktree", "remove", "--force", wt], REPO)
    r = sh(["git", "worktree", "add", "--detach", wt, COMMIT], REPO)
    if r.returncode:
        sys.exit("worktree add failed: " + r.stderr)
    head = sh(["git", "rev-parse", "HEAD"], wt).stdout.strip()
    f = os.path.join(wt, path)
    src = open(f).read()
    n = src.count(old)
    if n != 1:
        sys.exit(f"{rule}: old text occurs {n} times in {path}")
    open(f, "w").write(src.replace(old, new))
    diff = sh(["git", "diff"], wt).stdout
    open(os.path.join(OUT, rule + ".diff"), "w").write(diff)
    run = "^(" + "|".join(tests) + ")$"
    t = sh(["go", "test", "./internal/cli/", "./internal/replay/", "-count=1", "-run", run], wt,
           env={**os.environ, "CGO_ENABLED": "0"})
    with open(os.path.join(OUT, rule + ".log"), "w") as log:
        log.write(f"== rule {rule} on {COMMIT}, worktree {head}\n")
        log.write(f"== diff lines: {len(diff.splitlines())}\n{diff}")
        log.write(f"== go test -run '{run}' exit {t.returncode}\n{t.stdout}{t.stderr}")
    print(f"{rule}: diff {len(diff.splitlines())} lines, go test exit {t.returncode}")
    sh(["git", "worktree", "remove", "--force", wt], REPO)
