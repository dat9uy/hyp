#!/usr/bin/env python3
"""(d) Removals: each takes one rule out of 8a12324 in a detached worktree,
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
    ("internal/replay/validate.go", "\tif err := st.checkStore(ev); err != nil {\n\t\treturn err\n\t}\n", "")]),
  ("R8b", "TestEventNeverCrossesStores", [
    ("internal/store/history.go", 'return "", herr.Refused("hyp add goal needs a store: store discovery finds none from here, and a Goal has no parent whose store holds it; run it in the store\'s folder")', 'return "", nil')]),
  ("R8c", "TestEventNeverCrossesStores", [
    ("internal/store/history.go", "if len(stores) > 1 {", "if false && len(stores) > 1 {")]),
  ("R8d store rule first", "TestEventNeverCrossesStores", [
    ("internal/replay/validate.go", "\tif err := st.checkStore(ev); err != nil {\n\t\treturn err\n\t}\n\treturn rules[ev.Type].validate(&st.Facts, ev)", "\tif err := rules[ev.Type].validate(&st.Facts, ev); err != nil {\n\t\treturn err\n\t}\n\treturn st.checkStore(ev)"),
    ("internal/store/history.go", "\t\tif ev.Type != \"\" {\n\t\t\tt, terr := h.target(st, ev)", "\t\tif false {\n\t\t\tt, terr := h.target(st, ev)")]),
  ("RF fresh store joins", "TestCiteAcrossStores", [
    ("internal/store/history.go", 'case rel != "." && !strings.HasPrefix(rel, "../") && rel != ".." && storeOfHistory(repo, rel):', "case false:")]),
  ("RA --all reads the repository", "TestStatusAllListsTheHistory", [
    ("internal/cli/read.go", "\t\topen = store.OpenRepoHistory\n", "\t\topen = store.OpenHistory\n")]),
  ("RG G-id lookup", "TestStatusAllListsTheHistory", [
    ("internal/cli/read.go", "if err != nil && len(pos) == 1 && !*all {", "if false {")]),
  ("RR store view rejections", "TestDuplicateIDAcrossStores", [
    ("internal/cli/read.go", 'func onlyStoreRejections(p replay.Projection, s string) replay.Projection {\n\tif s == "" {', 'func onlyStoreRejections(p replay.Projection, s string) replay.Projection {\n\tif true {'),
    ("internal/cli/read.go", 'if cur == "" || x.Store == cur {', "if true {")]),
  ("Rroot root store read alone", "TestUntrackedStoreAllAndRoot", [
    ("internal/store/history.go", 'case rel != "." && !strings.HasPrefix(rel, "../")', 'case !strings.HasPrefix(rel, "../")')]),
  ("Rlock one lock root", "store:TestOneLockRoot", [
    ("internal/store/history.go", "\tf, err := os.Open(h.Root)", "\tdir := h.Root\n\tif len(h.Stores) == 1 && h.Stores[0] == \".\" {\n\t\tdir = eventsDir(h.Root)\n\t}\n\tf, err := os.Open(dir)")]),
  ("Rasof --as-of fallback", "TestStatusAllListsTheHistory", [
    ("internal/cli/read.go", "\tp, id, err := resolve(r)\n", "\tp, id, err := resolve(r)\n\tif _, aerr := r.at(*asOf); aerr != nil {\n\t\treturn aerr\n\t}\n")]),
  ("Rsame same-ask store warning", "TestSameAskAcrossStores", [
    ("internal/cli/write.go", "ms != nil && ks != nil && ms.Store != ks.Store", "false")]),
  ("Rpath same store by another path", "TestStoreByAnotherPath", [
    ("internal/store/history.go", "\t\tif s := sameStore(repo, stores, cur); s != \"\" {", "\t\tif s := sameStore(repo, stores, cur); false && s != \"\" {")]),
  ("Runtracked untracked ids", "TestAppendSortsAfterUntracked", [
    ("internal/store/history.go", "\tif u := h.unreadLast(); u > last {", "\tif u := h.unreadLast(); false && u > last {")]),
  ("Rcutfiles cut keeps .tree files", "TestSandboxHistoryCut", [
    ("internal/sandbox/sandbox.go", 'for _, f := range []string{".gitattributes", ".gitignore"} {\n\t\t\tif data', 'for _, f := range []string{} {\n\t\t\tif data')]),
  ("Routside outside link", "TestStoreByAnotherPath", [
    ("internal/store/history.go", "\tif _, ok := recipe.RepoRoot(abs); !ok {\n\t\tif real", "\tif false {\n\t\tif real")]),
  ("Rconfirm confirm keys", "TestEventNeverCrossesStores", [
    ("internal/replay/stores.go", "for _, k := range p.Accepts {\n\t\t\tif err := cross(", "for _, k := range p.Accepts[:0] {\n\t\t\tif err := cross(")]),
  ("Ralias --as-of in the Goal's store", "TestStatusAllListsTheHistory", [
    ("internal/cli/read.go", "p, err := r.atIn(*asOf, r.full.Nodes[id].Store)", "p, err := r.at(*asOf)")]),
  ("Rroot-id unread root store", "store:TestAppendSortsAfterUnreadRoot", [
    ("internal/store/history.go", "if read[filepath.Clean(", "if dir == \".\" || read[filepath.Clean(")]),
]
def sh(*a, cwd=None):
    return subprocess.run(a, cwd=cwd, capture_output=True, text=True)
for name, test, subs in R:
    wt = f"{OUT}/wt-{name.split()[0]}"
    sh("git", "-C", REPO, "worktree", "add", "--detach", wt, "8a12324")
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
    if test.startswith("store:"):
        pkg, test = "./internal/store", test[len("store:"):]
    r = sh("go", "test", pkg, "-count=1", "-run", f"^{test}$", cwd=wt)
    print("\n".join(l for l in (r.stdout + r.stderr).splitlines() if "_test.go" in l or l.startswith(("---", "ok", "FAIL")))[:3000])
    print(f"exit {r.returncode} -> {'fails as required' if r.returncode != 0 else 'PASSES: removal not caught'}")
    sh("git", "-C", REPO, "worktree", "remove", "--force", wt)
