#!/usr/bin/env python3
"""(c) One removal per rule from ded3e4b, each in a fresh detached worktree.
Each removal must give a non-empty diff and fail its test. R1 to R6 as
E-HSEK8J's c-remove.py; R4 also drops the imports its removal leaves unused;
R5 matches the append; R7, R9 to R11 as E-HJ8SQ9; R8 went with
replay.Merge; R12 and R13 are the rules of ded3e4b.""" 
import subprocess, sys, os
REPO = "/Users/datguy/codingProjects/tree-based-agent-memory"
base_dir = sys.argv[1]
KWAY = '''func MergeOrder(files map[string][]event.Event, tie func(a, b At) int) []At {
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
   lambda s: s.replace(between(s, "func MergeOrder(", "\n}\n") + "\n}\n", KWAY).replace('\t"strings"\n', '', 1).replace('\t"bytes"\n', '', 1).replace('\t"encoding/json"\n', '', 1), "./internal/cli", "TestReplayIgnoresLineOrder"),
  ("R4b", "line 4: the cut writes lines in file order", "internal/store/cut.go",
   lambda s: s.replace("raw[pos.Actor][pos.Index])", "raw[ev.Actor][len(c.lines[ev.Actor])])"), "./internal/cli", "TestReplayIgnoresLineOrder"),
  ("R5", "line 5: append id from the clock alone", "internal/store/store.go",
   lambda s: s.replace('if ev.ID = idAfter(now, last); ev.ID == "" {', 'if ev.ID = newID(now); ev.ID == "" && last != "" {'), "./internal/store", "TestAppendSortsAfterNewest"),
  ("R6", "line 6: init writes no .gitattributes", "internal/store/store.go",
   lambda s: s.replace('''	if err := os.WriteFile(filepath.Join(Dir(dir), ".gitattributes"), []byte(GitAttributes), 0o644); err != nil {
		return err
	}
''', ''), "./internal/store", "TestUnionMergeKeepsBothBranches"),
  ("R7", "R-ZN1VJJ#1: idAfter back to newest+1", "internal/store/store.go",
   lambda s: s.replace('''	if ms := u.Time(); ms < ulid.MaxTime() {
		return newID(ulid.Time(ms + 1))
	}
''', '''	for i := len(u) - 1; i >= 0; i-- {
		u[i]++
		if u[i] != 0 {
			break
		}
	}
	if true {
		return u.String()
	}
'''), "./internal/store", "TestClockBehindOnTwoClones"),
  ("R9", "R-SXCDZW#1, #2: the store's tie to file order", "internal/store/store.go",
   lambda s: s.replace("return func(a, b replay.At) int { return bytes.Compare(raw[a.Actor][a.Index], raw[b.Actor][b.Index]) }", "return func(a, b replay.At) int { return a.Index - b.Index }"), "./internal/store", "TestSameIDLinesAnyOrder"),
  ("R10", "R-SXCDZW#1, #2: the cut's tie to file order", "internal/store/cut.go",
   lambda s: s.replace('''		la, lb := raw[a.Actor][a.Index], raw[b.Actor][b.Index]
		if c := bytes.Compare(bytes.TrimSpace(la), bytes.TrimSpace(lb)); c != 0 {
			return c
		}
		return bytes.Compare(la, lb)
''', "\t\treturn a.Index - b.Index\n"), "./internal/store", "TestSameIDLinesAnyOrder"),
  ("R11", "R-SXCDZW#3: a constant step of one in the greatest millisecond", "internal/store/store.go",
   lambda s: s.replace("{1 + int64(binary.BigEndian.Uint32(step[:])), 1}", "{1 + 0*int64(binary.BigEndian.Uint32(step[:])), 1}"), "./internal/store", "TestAppendInGreatestMillisecond"),
  ("R12", "R-RZMF87#1: no newest+1 fallback in the greatest millisecond", "internal/store/store.go",
   lambda s: s.replace("[]int64{1 + int64(binary.BigEndian.Uint32(step[:])), 1}", "[]int64{1 + int64(binary.BigEndian.Uint32(step[:]))}"), "./internal/store", "TestAppendInGreatestMillisecond"),
  ("R13", "R-RZMF87#1: an unparseable newest id lets the clock id through", "internal/store/store.go",
   lambda s: s.replace('''	u, err := ulid.ParseStrict(last)
	if err != nil {
		return ""
	}
''', '''	u, err := ulid.ParseStrict(last)
	if err != nil {
		return id
	}
'''), "./internal/store", "TestAppendInGreatestMillisecond"),
]
def sh(*a, cwd=None):
    return subprocess.run(a, cwd=cwd, capture_output=True, text=True)
ok = True
for rid, what, path, edit, pkg, test in RULES:
    wt = os.path.join(base_dir, "wt-" + rid)
    if sh("git", "-C", REPO, "worktree", "add", "--detach", wt, "ded3e4b").returncode:
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
