#!/usr/bin/env python3
"""Removal checks of h-rec2-2's gate: each removal alone in a detached checkout
of the candidate. A removal passes when every pattern applies whole and the
named test fails on an assertion (go test exits 1, prints --- FAIL, and no
build failure). Usage: removals.py COMMIT OUTDIR"""
import os, subprocess, sys

S = "/private/tmp/claude-501/-Users-datguy-codingProjects-tree-based-agent-memory/55d8dd85-ee7f-47ed-9ca9-1d64f08971d0/scratchpad"
MAIN = "/Users/datguy/codingProjects/tree-based-agent-memory-feat-record"
commit, out = sys.argv[1], sys.argv[2]
wt = os.path.join(S, "removal-wt")

SKIPFILE = ('if key := src.Name + "\\x00" + rel; !adding[key] {',
            'if key := src.Name + "\\x00" + rel; f.Path != "" && !adding[key] {')
R = [
 ("r1", "missing named file is refused", "./internal/cli/", "TestImportRefusesAnEventWithoutItsBytes/removed", False,
  [("internal/store/import.go", 'return record.File{}, herr.Refused("%s, which %s names, is not in %s; hyp import adds an event with the bytes it names or not at all, so it adds nothing", rel, ev.ID, src.Src)', 'return record.File{}, nil'), ("internal/store/import.go",) + SKIPFILE]),
 ("r2", "named file with other bytes is refused", "./internal/cli/", "TestImportRefusesAnEventWithoutItsBytes/changed|TestImportRefusesAPathTwoEventsNameWithOtherBytes", False,
  [("internal/store/import.go", 'return record.File{}, herr.Refused("%s, which %s names, hashes %s in %s, not %s as the event stores; hyp import adds an event with the bytes it names or not at all, so it adds nothing", rel, ev.ID, got[:12], src.Src, sum[:min(12, len(sum))])', 'return record.File{}, nil'), ("internal/store/import.go",) + SKIPFILE]),
 ("r3", "pinned bundle the source lacks is refused", "./internal/cli/", "TestImportRefusesAnEventWithoutItsBundle", False,
  [("internal/store/import.go", 'return nil, "", herr.Refused("bundle %s, which %s pins, is not in %s with its pinned hash; hyp import adds an event with what it pins or not at all, so it adds nothing", b.Ref, ev.ID, src.Src)', 'return nil, "", nil')]),
 ("r4", "subject commit the repository lacks is refused", "./internal/cli/", "TestImportRefusesAnEventWithoutItsSubjectCommit", False,
  [("internal/store/import.go", 'return nil, herr.Refused("%s, the subject commit that %s names, is not in %s; fetch it, then run hyp import again, which adds nothing until then", c, ev.ID, repo.Top)', 'continue')]),
 ("r5", "id another store holds is refused", "./internal/cli/", "TestImportRefusesAnIDAnotherStoreHolds", False,
  [("internal/store/import.go", 'return nil, herr.Refused("event %s of the store %s is also an event of the store %s; an id is one event of one store, so hyp import adds nothing", ev.ID, src.Name, o)', 'continue')]),
 ("r6", "a name character outside the plain set is refused", "./internal/cli/", "TestStoreFolderNameIsPlain", False,
  [("internal/record/name.go", '\t\t\tif !plainRune(c) {', '\t\t\tif false && !plainRune(c) {')]),
 ("r7", "a part named as git's folder is refused", "./internal/cli/", "TestStoreFolderNameIsPlain", False,
  [("internal/record/name.go", 'if part == "" || part == "." || part == ".." || gitPart(part) {', 'if part == "" || part == "." || part == ".." {')]),
 ("r8", "a part that differs only in case from a sibling is refused", "./internal/cli/", "TestStoreFolderNameIsPlain", True,
  [("internal/record/name.go", 'if e.Name() != part && strings.EqualFold(e.Name(), part) {', 'if false && e.Name() != part && strings.EqualFold(e.Name(), part) {')]),
 ("r9", "discovery refuses a store folder outside the repository", "./internal/cli/", "TestStoreFolderOutsideTheRepository", False,
  [("internal/record/record.go", '\t\t\tif !inside {\n\t\t\t\treturn nil, outside(dir, top)\n\t\t\t}\n', '')]),
 ("r10", "init and import refuse a store folder outside the repository", "./internal/cli/", "TestStoreFolderOutsideTheRepository", False,
  [("internal/record/name.go", '\tname, ok := storeName(top, dir)\n\tif !ok {\n\t\treturn outside(dir, top)\n\t}', '\tname, ok := storeName(top, dir)\n\tif !ok {\n\t\treturn nil\n\t}')]),
 ("r11", "a new id sorts after every store's ids", "./internal/cli/", "TestNewIDSortsAfterEveryStoresID", False,
  [("internal/store/store.go", '\tfor _, ids := range all {\n\t\tfor _, id := range ids {', '\tfor _, ids := range all {\n\t\tfor _, id := range ids[:0] {')]),
 ("r12", "a write's index lines end in NUL", "./internal/cli/", "TestNamedPathWithANewline", False,
  [("internal/record/commit.go", '"update-index", "-z", "--add", "--index-info"', '"update-index", "--add", "--index-info"'),
   ("internal/record/commit.go", '\\t%s\\x00", sha, p)', '\\t%s\\n", sha, p)'),
   ("internal/record/commit.go", '\\t%s\\x00", a.mode, a.sha, p)', '\\t%s\\n", a.mode, a.sha, p)')]),
 ("r13", "a sync join's index lines end in NUL", "./internal/cli/", "TestNamedPathWithANewline", False,
  [("internal/record/sync.go", '"update-index", "-z", "--add", "--index-info"', '"update-index", "--add", "--index-info"'),
   ("internal/record/sync.go", '\\t%s\\x00", e.mode, e.sha, p)', '\\t%s\\n", e.mode, e.sha, p)'),
   ("internal/record/sync.go", '\\t%s\\x00", kept[k].mode, kept[k].sha, k)', '\\t%s\\n", kept[k].mode, kept[k].sha, k)')]),
 ("r14", "a read holds its snapshot", "./internal/record/", "TestPruneLeavesASnapshotAReadHolds", False,
  [("internal/record/snapshot.go", '\tif err := r.holdSnapshot(key); err != nil {\n\t\treturn "", err\n\t}\n', '\tif err := os.MkdirAll(r.hypDir("snaplock"), 0o755); err != nil {\n\t\treturn "", err\n\t}\n')]),
 ("r15", "a prune takes the snapshot's lock first", "./internal/record/", "TestPruneLeavesASnapshotAReadHolds", False,
  [("internal/record/snapshot.go", '\t\tif syscall.Flock(int(f.Fd()), syscall.LOCK_EX|syscall.LOCK_NB) == nil {', '\t\tif true {')]),
 ("r16", "a part named as git's folder in another letter case is refused", "./internal/cli/", "TestStoreBelowAFolderNamedAsGitsInOtherCase", False,
  [("internal/record/name.go", 'if part == "" || part == "." || part == ".." || gitPart(part) {', 'if part == "" || part == "." || part == ".." || part == GitName {')]),
 ("r17", "discovery refuses a store below a folder named as git's", "./internal/cli/", "TestStoreBelowAFolderNamedAsGitsInOtherCase", False,
  [("internal/record/record.go", '\t\t\tif gitPart(name) {\n\t\t\t\treturn nil, belowGit(dir, name)\n\t\t\t}\n', '')]),
 ("r18", "sandbox create keeps a refusal of --store at exit 1", "./internal/cli/", "TestSandboxStoreOutsideTheRepository", False,
  [("internal/sandbox/sandbox.go", '\t\t\tif herr.KindOf(err) == herr.Refusal {\n\t\t\t\treturn nil, herr.Refused("--store %s: %s; nothing created", o.Store, herr.Text(err))\n\t\t\t}\n', '')]),
 ("r19", "a store cut keeps the refusal of the store folder", "./internal/cli/", "TestSandboxStoreOutsideTheRepository", False,
  [("internal/store/cut.go", '\tif err != nil {\n\t\t// A refusal, such as a store folder outside the repository, or an\n\t\t// environment problem keeps its own exit code.\n\t\treturn nil, err\n\t}', '\tif err != nil {\n\t\treturn nil, fmt.Errorf("%s: %w", src, ErrNoStore)\n\t}')]),
 ("r20", "a prune removes a pruned snapshot's lock file", "./internal/record/", "TestPruneLeavesASnapshotAReadHolds", False,
  [("internal/record/snapshot.go", '\t\t\t\t_ = os.Remove(r.snapLock(d.Name()))\n', '')]),
 ("p1", "line 10: a store-folder literal outside package record", "./internal/cli/", "TestOnlyRecordNamesTheFoldersAndRunsGit", False,
  [("internal/render/probe.go", None, 'package render\n\nvar probe = "a/.tree/b"\n')]),
 ("p2", "line 10: exec.Command with the literal git outside record and sandbox", "./internal/cli/", "TestOnlyRecordNamesTheFoldersAndRunsGit", False,
  [("internal/render/probe.go", None, 'package render\n\nimport "os/exec"\n\nfunc probe() { _ = exec.Command("git", "status") }\n')]),
]

def sh(*a, **k):
    return subprocess.run(a, capture_output=True, text=True, **k)

os.makedirs(out, exist_ok=True)
sh("rm", "-rf", wt); sh("git", "-C", MAIN, "worktree", "prune")
r = sh("git", "-C", MAIN, "worktree", "add", "-q", "--detach", wt, commit)
if r.returncode:
    sys.exit(r.stderr)
summary = []
only = set(sys.argv[3].split(",")) if len(sys.argv) > 3 else None
for rid, what, pkg, test, cs, edits in R:
    if only and rid not in only:
        continue
    sh("git", "-C", wt, "checkout", "-q", "--", "."); sh("git", "-C", wt, "clean", "-qfd")
    log = [f"removal {rid}: {what}", f"candidate {commit}", f"test {pkg} -run '{test}'" + (" with TMPDIR on a case-sensitive volume" if cs else "")]
    ok = True
    for e in edits:
        path = os.path.join(wt, e[0])
        if e[1] is None:
            open(path, "w").write(e[2]); log.append(f"wrote {e[0]}:\n{e[2]}"); continue
        s = open(path).read()
        n = s.count(e[1])
        if n == 0:
            ok = False; log.append(f"pattern does not apply in {e[0]}: {e[1]!r}"); break
        open(path, "w").write(s.replace(e[1], e[2]))
        log.append(f"applied in {e[0]} ({n}x): {e[1]!r} -> {e[2]!r}")
    log.append("diff:\n" + sh("git", "-C", wt, "diff").stdout)
    if ok:
        env = dict(os.environ)
        if cs:
            env["TMPDIR"] = os.path.join(S, "cs")
        t = sh("go", "test", pkg, "-count=1", "-v", "-run", test, cwd=wt, env=env)
        log.append(t.stdout + t.stderr); log.append(f"exit {t.returncode}")
        built = "build failed" not in t.stdout + t.stderr and "[setup failed]" not in t.stdout
        verdict = "removal ok: the test fails on an assertion" if t.returncode == 1 and "--- FAIL" in t.stdout and built else "FAILED removal"
    else:
        verdict = "FAILED removal: a pattern does not apply whole"
    log.append(verdict)
    open(os.path.join(out, f"removal-{rid}.txt"), "w").write("\n".join(log) + "\n")
    summary.append(f"{rid}: {verdict}")
sh("git", "-C", MAIN, "worktree", "remove", "--force", wt)
open(os.path.join(out, "removal-summary.txt"), "a").write("\n".join(summary) + "\n")
print("\n".join(summary))
