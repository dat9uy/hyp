#!/usr/bin/env python3
"""(c) One removal per rule from 65a1cc5, each in a fresh detached worktree.
Each removal must give a non-empty diff and fail its line's test.
Attempt 2: R4 also drops the strings import its removal leaves unused.
Attempt 1 (c-remove-attempt1.py, c-removals-attempt1-r4-build-failed.log)
did not build on R4 and is not counted."""
import subprocess, sys, os
REPO = "/Users/datguy/codingProjects/tree-based-agent-memory"
base_dir = sys.argv[1]
KWAY = '''func MergeOrder(files map[string][]event.Event) []At {
	names := make([]string, 0, len(files))
	total := 0
	for name, evs := range files {
		names = append(names, name)
		total += len(evs)
	}
	sort.Strings(names)
	cursor := make([]int, len(names))
	out := make([]At, 0, total)
	for len(out) < total {
		best := -1
		for i, name := range names {
			if cursor[i] >= len(files[name]) {
				continue
			}
			if best < 0 || files[name][cursor[i]].ID < files[names[best]][cursor[best]].ID {
				best = i
			}
		}
		out = append(out, At{names[best], cursor[best]})
		cursor[best]++
	}
	return out
}
'''
def between(s, start, end):
    i = s.index(start); j = s.index(end, i)
    return s[i:j]
RULES = [
  ("R1", "line 1: the cut keeps the base's .tree (RemoveAll dropped)", "internal/sandbox/sandbox.go",
   lambda s: s.replace('''		if err := os.RemoveAll(filepath.Join(wt, ".tree")); err != nil {
			return nil, nil, err
		}
''', ''), "./internal/cli", "TestSandboxCutOverBaseTree"),
  ("R1b", "line 1: the base refusal restored", "internal/sandbox/sandbox.go",
   lambda s: s.replace('''	if o.Store != "" {
		if cut, err''', '''	if o.Store != "" {
		if _, err := git("-C", o.Host, "cat-file", "-e", o.Base+":.tree"); err == nil {
			return nil, herr.BadArgs("%s already holds .tree at %s; a store cut needs a base without one", o.Host, o.Base)
		}
		if cut, err'''), "./internal/cli", "TestSandboxCutOverBaseTree"),
  ("R2", "line 2: init without the discovery refusal", "internal/store/store.go",
   lambda s: s.replace(between(s, "	root, err := FindRoot(dir)\n", "	for _, sub := range"), ""), "./internal/cli", "TestInitRefusedInsideStore"),
  ("R3", "line 3: discovery without the repository-root stop", "internal/store/store.go",
   lambda s: s.replace('''		if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
			return "", ErrNoStore
		}
''', ''), "./internal/cli", "TestDiscoveryStopsAtRepositoryRoot"),
  ("R4", "line 4: Merge by file order (k-way)", "internal/replay/replay.go",
   lambda s: s.replace(between(s, "func MergeOrder(", "\n}\n") + "\n}\n", KWAY).replace('\t"strings"\n', '', 1), "./internal/cli", "TestReplayIgnoresLineOrder"),
  ("R4b", "line 4: the cut writes lines in file order", "internal/store/cut.go",
   lambda s: s.replace("raw[pos.Actor][pos.Index])", "raw[ev.Actor][len(c.lines[ev.Actor])])"), "./internal/cli", "TestReplayIgnoresLineOrder"),
  ("R5", "line 5: append id from the clock alone", "internal/store/store.go",
   lambda s: s.replace("ev.ID = idAfter(now, last)", "ev.ID = newID(now)\n\t_ = last"), "./internal/store", "TestAppendSortsAfterNewest"),
  ("R6", "line 6: init writes no .gitattributes", "internal/store/store.go",
   lambda s: s.replace('''	if err := os.WriteFile(filepath.Join(Dir(dir), ".gitattributes"), []byte(GitAttributes), 0o644); err != nil {
		return err
	}
''', ''), "./internal/store", "TestUnionMergeKeepsBothBranches"),
]
def sh(*a, cwd=None):
    return subprocess.run(a, cwd=cwd, capture_output=True, text=True)
ok = True
for rid, what, path, edit, pkg, test in RULES:
    wt = os.path.join(base_dir, "wt-" + rid)
    if sh("git", "-C", REPO, "worktree", "add", "--detach", wt, "65a1cc5").returncode:
        print("worktree failed", rid); sys.exit(3)
    p = os.path.join(wt, path)
    s = open(p).read(); n = edit(s)
    open(p, "w").write(n)
    diff = sh("git", "diff", cwd=wt).stdout
    print(f"== {rid} {what}\n-- diff ({len(diff.splitlines())} lines)\n{diff}")
    r = sh("go", "test", pkg, "-count=1", "-run", f"^{test}$", cwd=wt)
    out = r.stdout + r.stderr
    print(f"-- go test {pkg} -run {test}: exit {r.returncode}\n{out}")
    failed = r.returncode != 0 and f"--- FAIL: {test}" in out
    print(f"-- {rid}: diff {'non-empty' if diff else 'EMPTY'}, test {'fails' if failed else 'DOES NOT FAIL'}\n")
    ok &= bool(diff) and failed
print("== all removals fail their test with a non-empty diff:", ok)
