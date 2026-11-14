#!/usr/bin/env python3
"""(c) Removals: each takes one rule out of a9963fc in a detached worktree,
logs the diff, and runs the line test that must then fail."""
import subprocess, sys
REPO = "/Users/datguy/codingProjects/tree-based-agent-memory"
OUT = sys.argv[1]
R = [
  ("R3", "TestHistoryLeavesOutStores", [
    ("internal/store/history.go", '\tif rel == "." {\n\t\treturn true\n\t}\n\tparts', '\treturn true\n\tparts'),
    ("internal/store/history.go", '"ls-files", "-z", "--cached",', '"ls-files", "-z", "--cached", "--others",')]),
  ("R6", "TestDuplicateIDAcrossStores", [
    ("internal/replay/validate.go", "if f.SeenIDs[ev.ID] {", "if false && f.SeenIDs[ev.ID] {")]),
  ("R7a", "TestAliasLocalToStore", [
    ("internal/cli/resolve.go", '(scope == "" || n.Store == scope)', "true")]),
  ("R7b", "TestAliasLocalToStore", [
    ("internal/replay/validate.go", "f.Aliases[AliasKey{ev.Store, p.Alias}]", 'f.Aliases[AliasKey{"", p.Alias}]'),
    ("internal/replay/rules.go", "st.Aliases[AliasKey{ev.Store, p.Alias}]", 'st.Aliases[AliasKey{"", p.Alias}]')]),
  ("R7c", "TestAliasLocalToStore", [
    ("internal/claim/claim.go", "n.Kind != event.KindHypothesis || n.Store != store", "n.Kind != event.KindHypothesis")]),
  ("R8a", "TestEventNeverCrossesStores", [
    ("internal/replay/validate.go", "return st.checkStore(ev)", "return nil")]),
  ("R8b", "TestEventNeverCrossesStores", [
    ("internal/store/history.go", 'return "", herr.Refused("hyp add goal needs a store: store discovery finds none from here, and a Goal has no parent whose store holds it; run it in the store\'s folder")', 'return "", nil')]),
  ("R8c", "TestEventNeverCrossesStores", [
    ("internal/store/history.go", "if len(stores) > 1 {", "if false && len(stores) > 1 {")]),
]
def sh(*a, cwd=None):
    return subprocess.run(a, cwd=cwd, capture_output=True, text=True)
for name, test, subs in R:
    wt = f"{OUT}/wt-{name}"
    sh("git", "-C", REPO, "worktree", "add", "--detach", wt, "a9963fc")
    for f, old, new in subs:
        p = f"{wt}/{f}"
        s = open(p).read()
        if old not in s:
            print(f"== {name}: pattern missing in {f}"); break
        open(p, "w").write(s.replace(old, new, 1))
    print(f"== {name} ({test})")
    d = sh("git", "diff", cwd=wt).stdout
    print(d if d else "EMPTY DIFF")
    pkg = "./internal/cli"
    r = sh("go", "test", pkg, "-count=1", "-run", f"^{test}$", cwd=wt)
    print("\n".join(l for l in (r.stdout + r.stderr).splitlines() if "history_test.go" in l or l.startswith(("---", "ok", "FAIL")))[:3000])
    print(f"exit {r.returncode} -> {'fails as required' if r.returncode != 0 else 'PASSES: removal not caught'}")
    sh("git", "-C", REPO, "worktree", "remove", "--force", wt)
