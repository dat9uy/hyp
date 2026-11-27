package channel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

const (
	titleBefore = "#!/bin/sh\nprintf '%s\\n' \"$1\" | cut -c1-10\n"
	titleAfter  = "#!/bin/sh\nprintf '%s\\n' \"$1\"\n"
	wholeTitle  = "-- words --\nthe whole title shows\n-- run --\nsh title.sh abcdefghijklmnopqrst\n-- floor: equals --\nabcdefghijklmnopqrst\n"
	shortTitle  = "-- words --\na short title prints as today\n-- run --\nsh title.sh short\n-- floor: equals --\nshort\n"
)

// readyLine finds the printed line of role that starts with prefix, as the
// argument list to hand to hyp.
func readyLine(t *testing.T, out, role, prefix string) []string {
	t.Helper()
	for _, l := range readyLines(out) {
		if l.role == role && strings.HasPrefix(l.line, prefix) {
			return strings.Fields(l.line)[1:]
		}
	}
	t.Fatalf("%s may no %q in:\n%s", role, prefix, out)
	return nil
}

// TestOneGoalEndToEnd runs one Goal from add goal to land, each next command
// the line hyp printed. Git holds only the owner's acts: the add, which
// carries the examples, and the land, which carries the evidence.
func TestOneGoalEndToEnd(t *testing.T) {
	r := testutil.NewRepo(t)
	keepStoreOut(t, r.Dir)
	r.Commit(t, map[string]string{"title.sh": titleBefore})
	t.Chdir(r.Dir)

	// the grill commits a draft folder with one example per line; the commit
	// that holds it is H0
	for key, txtar := range map[string]string{"line-1/a": wholeTitle, "line-2/a": shortTitle} {
		dir := filepath.Join("goals", "draft-title", key)
		if err := os.MkdirAll(filepath.Join(r.Dir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.Dir, dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
			t.Fatal(err)
		}
		r.WriteExec(t, filepath.Join(dir, "probe", "run"), fixtureProbe)
	}
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "draft examples")
	base := testutil.Git(t, r.Dir, "rev-parse", "HEAD")

	// the owner adds the Goal with the draft folder: it is open at once
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	code, out, errs := run(t, "add", "goal", "--base", base, "--words", "Show the whole title.",
		"--line", "gap:status shows the whole title", "--line", "guard:short titles print as today",
		"--examples", "goals/draft-title")
	fields := strings.Fields(out)
	if code != 0 || len(fields) < 2 || !strings.Contains(out, " is open") || !strings.Contains(out, "copied the examples to goals/") {
		t.Fatalf("add goal: code %d, stdout %q, stderr %q", code, out, errs)
	}
	goal := fields[1]
	if _, err := os.Stat(filepath.Join(r.Dir, "goals", goal, "line-1", "a", "example.txtar")); err != nil {
		t.Fatalf("the examples are not copied to the Goal's folder: %v", err)
	}
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "add the Goal")
	if _, out, _ = run(t, "status"); !strings.Contains(out, goal+"  open  runs: 0 of 3 used\nstop: none\n") {
		t.Fatalf("the Goal is not open after the add:\n%s", out)
	}

	// the runner commits the fix and the spec, then runs both lines on it
	fix := r.Commit(t, map[string]string{"title.sh": titleAfter, "goals/" + goal + "/specs/spec.md": "the title says bye\n"})
	afterCommit := testutil.Snapshot(t, r.Dir)
	t.Setenv("HYP_ACTOR", "runner-1")
	t.Setenv("HYP_ROLE", "runner")
	for _, line := range []string{"line-1", "line-2"} {
		args := readyLine(t, mustStatus(t), "runner", "hyp run <commit> "+goal+"/"+line)
		args[1] = fix
		code, out, errs := run(t, args...)
		if code != 0 || !strings.Contains(out, "met") {
			t.Fatalf("%s: code %d, stdout %q, stderr %q", line, code, out, errs)
		}
	}
	if got := testutil.Snapshot(t, r.Dir); got != afterCommit {
		t.Fatal("hyp changed the repository's git state between the last commit and the select")
	}

	// the only candidate is the fix: this clone shows the Goal selected, and
	// status prints the PR body
	_, status, _ := run(t, "status", goal)
	if !strings.Contains(status, "stop: selected") || !strings.Contains(status, "selected commit: "+fix) ||
		!strings.Contains(status, "## line-1 (gap): status shows the whole title") {
		t.Fatalf("status after the runs:\n%s", status)
	}

	// the owner lands with the printed line
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	landArgs := readyLine(t, mustStatus(t), "owner", "hyp land "+goal+" ")
	code, out, errs = run(t, landArgs...)
	if code != 0 || !strings.Contains(out, fix) {
		t.Fatalf("land: code %d, stdout %q, stderr %q", code, out, errs)
	}
	if _, status, _ = run(t, "status", goal); !strings.Contains(status, "ended") {
		t.Fatalf("status after the land:\n%s", status)
	}
}

func mustStatus(t *testing.T) string {
	t.Helper()
	code, out, errs := run(t, "status")
	if code != 0 {
		t.Fatalf("status: code %d, stderr %q", code, errs)
	}
	return out
}
