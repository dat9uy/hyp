package channel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

// treeFiles lists every file under dir/.tree with its bytes.
func treeFiles(t *testing.T, dir string) map[string]string {
	t.Helper()
	out := map[string]string{}
	root := filepath.Join(dir, ".tree")
	err := filepath.Walk(root, func(p string, fi os.FileInfo, err error) error {
		if err != nil || fi.IsDir() {
			return err
		}
		b, err := os.ReadFile(p)
		out[p] = string(b)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	return out
}

// TestFreshCloneReplaysFromGit: clone A runs and selects; git holds only the
// select, and the Result and outputs stay in A's trace. A fresh clone B shows
// the same state and no refusal, counts no run, says the selected row's cell
// is not in its trace with the try line that runs it again, and after that
// try prints the cell as run again in this clone.
func TestFreshCloneReplaysFromGit(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": "ok\n", f.spec(): "spec\n"})
	testutil.Git(t, f.r.Dir, "branch", "cand", c)

	before := treeFiles(t, f.r.Dir)
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	after := treeFiles(t, f.r.Dir)
	if len(after) != len(before) {
		t.Fatalf("run changed .tree: %d files before, %d after", len(before), len(after))
	}
	for p, b := range before {
		if after[p] != b {
			t.Fatalf("run changed %s", p)
		}
	}
	_, out, _ := run(t, "status")
	args := selectLine(t, out)
	if code, out, errs := run(t, args[1:]...); code != 0 {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, out, _ := run(t, "status", f.goal); strings.Contains(out, "on no branch or tag") {
		t.Fatalf("a warning for a commit on branch cand:\n%s", out)
	}

	testutil.Git(t, f.r.Dir, "add", "-f", ".tree")
	testutil.Git(t, f.r.Dir, "commit", "-q", "-m", "store")
	b := filepath.Join(t.TempDir(), "b")
	testutil.Git(t, f.r.Dir, "clone", "-q", "--no-local", f.r.Dir, b)
	t.Chdir(b)
	t.Setenv("HYP_ACTOR", "reader-1")
	t.Setenv("HYP_ROLE", "")

	_, out, _ = run(t, "status", f.goal)
	try := "hyp try " + f.goal + " line-1/a --commit " + c
	for _, want := range []string{f.goal + "  selected  runs: 0 of 3 used\n", "not in this clone's trace; run it again with: " + try + "\n"} {
		if !strings.Contains(out, want) {
			t.Fatalf("clone B lacks %q:\n%s", want, out)
		}
	}
	if strings.Contains(out, "refused at replay") || strings.Contains(out, "output, 3 bytes") {
		t.Fatalf("clone B refuses an event or shows A's output:\n%s", out)
	}
	if code, out, errs := run(t, strings.Fields(try)[1:]...); code != 0 || !strings.Contains(out, "saved in this clone's trace") {
		t.Fatalf("try: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ = run(t, "status", f.goal)
	if !strings.Contains(out, "run again in this clone") || strings.Contains(out, "not in this clone's trace") {
		t.Fatalf("clone B does not print the cell it ran again:\n%s", out)
	}
	if code, out, _ := run(t, strings.Fields(try)[1:]...); code != 0 || strings.Contains(out, "saved in") {
		t.Fatalf("a second try saved the cell again:\n%s", out)
	}
}

// TestWarnsWhenSelectedCommitIsOnNoRef: once no branch or tag holds the
// selected commit, the list and the Goal's state block warn, also when
// another commit with the same subject is on a branch.
func TestWarnsWhenSelectedCommitIsOnNoRef(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	files := map[string]string{"counter": "ok\n", f.spec(): "spec\n"}
	c := f.leaf(t, files)
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ := run(t, "status")
	args := selectLine(t, out)
	if code, out, errs := run(t, args[1:]...); code != 0 || !strings.Contains(out, "on no branch or tag") {
		t.Fatalf("select's return lacks the warning: code %d, stdout %q, stderr %q", code, out, errs)
	}
	// a twin: the same subject on a branch, another commit (a file under
	// goals/ is outside the subject)
	testutil.Git(t, f.r.Dir, "checkout", "-q", "-b", "twin", f.base)
	twinFiles := map[string]string{"goals/twin.txt": "twin\n"}
	for k, v := range files {
		twinFiles[k] = v
	}
	twin := f.r.Commit(t, twinFiles)
	testutil.Git(t, f.r.Dir, "checkout", "-q", "main")
	if twin == c {
		t.Fatal("test setup: the twin is the selected commit")
	}
	for _, args := range [][]string{{"status"}, {"status", f.goal}} {
		_, out, _ := run(t, args...)
		if !strings.Contains(out, "warning: the selected commit "+c+" is on no branch or tag") {
			t.Errorf("%s lacks the warning:\n%s", strings.Join(args, " "), out)
		}
	}
	testutil.Git(t, f.r.Dir, "tag", "kept", c)
	if _, out, _ := run(t, "status"); strings.Contains(out, "on no branch or tag") {
		t.Errorf("a warning for a commit a tag holds:\n%s", out)
	}
}

// TestSpecStoresEachKindFromOneRule: every event kind says whether it changes
// a Goal's state, and is stored in git exactly when it does; every guard says
// whether replay checks it, and the guards that read a Result do not.
func TestSpecStoresEachKindFromOneRule(t *testing.T) {
	_, out, _ := run(t, "spec")
	var kinds, guards []string
	block := ""
	for _, l := range strings.Split(out, "\n") {
		switch {
		case strings.HasPrefix(l, "events:"), strings.HasPrefix(l, "guards:"):
			block = l[:6]
		case l == "":
			block = ""
		case block == "events" && strings.HasPrefix(l, "  "):
			kinds = append(kinds, l)
		case block == "guards" && strings.HasPrefix(l, "  ") && strings.Contains(l, "  rule: "):
			guards = append(guards, l)
		}
	}
	if len(kinds) < 12 || len(guards) < 20 {
		t.Fatalf("found %d kinds and %d guards in:\n%s", len(kinds), len(guards), out)
	}
	for _, k := range kinds {
		yes := strings.Contains(k, "changes a Goal's state: yes  stored in git  ")
		no := strings.Contains(k, "changes a Goal's state: no  stored in trace  ")
		if yes == no {
			t.Errorf("kind row does not store by the rule: %s", k)
		}
	}
	for _, name := range []string{"result.record", "cells.cache"} {
		found := false
		for _, k := range kinds {
			found = found || strings.HasPrefix(k, "  "+name+"  changes a Goal's state: no  stored in trace  ")
		}
		if !found {
			t.Errorf("%s is not stored in the trace", name)
		}
	}
	for _, g := range guards {
		if !strings.Contains(g, "  checked at replay: yes  ") && !strings.Contains(g, "  checked at replay: no  ") {
			t.Errorf("guard row says nothing about replay: %s", g)
		}
	}
	for _, name := range []string{"Candidate", "New cell", "Budget", "Stable"} {
		found := false
		for _, g := range guards {
			found = found || (strings.HasPrefix(g, "  "+name+"  rule: ") && strings.Contains(g, "  checked at replay: no  "))
		}
		if !found {
			t.Errorf("%s reads the Results of the trace, and replay checks it", name)
		}
	}
}
