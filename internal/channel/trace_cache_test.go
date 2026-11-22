package channel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

// selectedFx runs a met candidate on its own branch and selects it, as the
// runner. It returns the candidate commit.
func selectedFx(t *testing.T) (*fx, string) {
	t.Helper()
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": "ok\n", f.spec(): "spec\n"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	_, out, _ := run(t, "status")
	if code, out, errs := run(t, selectLine(t, out)[1:]...); code != 0 {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	return f, c
}

// freshClone commits the store of f's repository and clones it into a
// folder of its own, whose trace is empty, and changes into the clone.
func freshClone(t *testing.T, f *fx) string {
	t.Helper()
	testutil.Git(t, f.r.Dir, "add", "-f", ".tree")
	testutil.Git(t, f.r.Dir, "commit", "-q", "-m", "store")
	dir := filepath.Join(t.TempDir(), "clone")
	testutil.Git(t, f.r.Dir, "clone", "-q", f.r.Dir, dir)
	t.Chdir(dir)
	return dir
}

// Git holds no cell, Result or output: the store holds the place and the
// select, and every output is in the clone's trace.
func TestGitKeepsNoCell(t *testing.T) {
	f, c := selectedFx(t)
	if entries, err := os.ReadDir(filepath.Join(f.r.Dir, ".tree", "outputs")); err == nil && len(entries) > 0 {
		t.Fatalf(".tree/outputs holds %d files, want none", len(entries))
	}
	var all strings.Builder
	paths, _ := filepath.Glob(filepath.Join(f.r.Dir, ".tree", "events", "*.jsonl"))
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err != nil {
			t.Fatal(err)
		}
		all.Write(b)
	}
	for _, word := range []string{`"verdict"`, `"output"`, `"stderr"`, `"result.record"`} {
		if strings.Contains(all.String(), word) {
			t.Fatalf("the store holds %s:\n%s", word, all.String())
		}
	}
	if !strings.Contains(all.String(), `"type":"select"`) || !strings.Contains(all.String(), c) {
		t.Fatalf("the store holds no select of %s:\n%s", c, all.String())
	}
}

// In a fresh clone of a selected Goal the candidate's cell is not in the
// trace: hyp status G says so and prints the try line, and that try line,
// run as printed, saves the cell, marked as run again in this clone. Git sees
// nothing of it, and the first clone shows no mark.
func TestFreshCloneRunsAMissingCellAgain(t *testing.T) {
	f, c := selectedFx(t)
	first := f.r.Dir
	dir := freshClone(t, f)
	t.Setenv("HYP_ACTOR", "reader")
	t.Setenv("HYP_ROLE", "")
	_, out, _ := run(t, "status", f.goal)
	want := "verdict: not in this clone's trace; run it again with: hyp try " + f.goal + " line-1/a --commit " + c
	if !strings.Contains(out, f.goal+"  selected") || !strings.Contains(out, want) || strings.Contains(out, "verdict: met") {
		t.Fatalf("status in the fresh clone:\n%s\nwant %q", out, want)
	}
	at := strings.Index(out, "hyp try ")
	line, _, _ := strings.Cut(out[at:], "\n")
	code, tried, errs := run(t, strings.Fields(line)[1:]...)
	if code != 0 || !strings.Contains(tried, "saved in this clone's trace") {
		t.Fatalf("try: code %d, stdout %q, stderr %q", code, tried, errs)
	}
	_, out, _ = run(t, "status", f.goal)
	if !strings.Contains(out, "verdict: met, exit 0, run again in this clone") || strings.Contains(out, "not in this clone's trace") {
		t.Fatalf("status after the try:\n%s", out)
	}
	if st := testutil.Git(t, dir, "status", "--porcelain", "--untracked-files=all"); st != "" {
		t.Fatalf("the try changed the clone: %q", st)
	}
	t.Chdir(first)
	if _, out, _ = run(t, "status", f.goal); strings.Contains(out, "run again in this clone") {
		t.Fatalf("the first clone shows a mark:\n%s", out)
	}
}

// A second try of a cell the trace holds saves nothing.
func TestTryOfACellTheTraceHoldsSavesNothing(t *testing.T) {
	f, c := selectedFx(t)
	code, tried, errs := run(t, "try", f.goal, "line-1/a", "--commit", c)
	if code != 0 || !strings.Contains(tried, "nothing is recorded") {
		t.Fatalf("try: code %d, stdout %q, stderr %q", code, tried, errs)
	}
}

// In a fresh clone of an open Goal, H0's cell is not in the trace, the Goal
// is open from the place in git, and the budget counts only this clone's
// runs.
func TestFreshCloneOfAnOpenGoal(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	c := f.leaf(t, map[string]string{"counter": "ok\n"})
	if code, out, errs := run(t, "run", c, f.goal+"/line-1"); code != 0 {
		t.Fatalf("run: code %d, stdout %q, stderr %q", code, out, errs)
	}
	freshClone(t, f)
	_, out, _ := run(t, "status", f.goal)
	if !strings.Contains(out, f.goal+"  open  runs: 0 of 3 used") ||
		!strings.Contains(out, "verdict on H0 at place: not in this clone's trace; run it again with: hyp try "+f.goal+" line-1/a\n") {
		t.Fatalf("status in the fresh clone:\n%s", out)
	}
	if code, out, errs := run(t, "try", f.goal, "line-1/a"); code != 0 || !strings.Contains(out, "saved in this clone's trace") {
		t.Fatalf("try: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, out, _ = run(t, "status", f.goal); !strings.Contains(out, "verdict on H0: failed, exit 0, run again in this clone") {
		t.Fatalf("status after the try:\n%s", out)
	}
}

// Every state block of a selected Goal warns once the selected commit is on
// no branch or tag, and not before.
func TestWarnsWhenTheSelectedCommitIsOnNoRef(t *testing.T) {
	f, c := selectedFx(t)
	const warn = "is on no branch or tag"
	if _, out, _ := run(t, "status"); strings.Contains(out, warn) {
		t.Fatalf("a warning while the branch holds the commit:\n%s", out)
	}
	testutil.Git(t, f.r.Dir, "branch", "-D", "cand-"+c[:12])
	for _, args := range [][]string{{"status"}, {"status", f.goal}} {
		if _, out, _ := run(t, args...); !strings.Contains(out, "warning: the selected commit "+c+" "+warn) {
			t.Fatalf("hyp %v gives no warning:\n%s", args, out)
		}
	}
	testutil.Git(t, f.r.Dir, "tag", "kept", c)
	if _, out, _ := run(t, "status"); strings.Contains(out, warn) {
		t.Fatalf("a warning while a tag holds the commit:\n%s", out)
	}
}

func TestTryTakesOneSubject(t *testing.T) {
	f := newFx(t, map[string]string{"line-1/a": counterEx})
	if code, _, errs := run(t, "try", f.goal, "line-1/a", "--worktree", "--commit", "HEAD"); code != 2 {
		t.Fatalf("code %d, stderr %q; want 2", code, errs)
	}
	if code, _, errs := run(t, "try", f.goal, "line-1/a", "--commit", "nope"); code != 2 {
		t.Fatalf("code %d, stderr %q; want 2", code, errs)
	}
}
