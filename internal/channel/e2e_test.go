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

// TestOneGoalEndToEnd runs one Goal from add goal to select, each next command
// the line hyp printed.
func TestOneGoalEndToEnd(t *testing.T) {
	r := testutil.NewRepo(t)
	keepStoreOut(t, r.Dir)
	base := r.Commit(t, map[string]string{"title.sh": titleBefore})
	t.Chdir(r.Dir)

	// the owner adds the Goal
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	code, out, errs := run(t, "add", "goal", "--base", base, "--words", "Show the whole title.",
		"--line", "gap:status shows the whole title", "--line", "guard:short titles print as today")
	fields := strings.Fields(out)
	if code != 0 || len(fields) < 2 || !strings.Contains(out, "stop: none") || !strings.Contains(out, "waits: grill") {
		t.Fatalf("add goal: code %d, stdout %q, stderr %q", code, out, errs)
	}
	goal := fields[1]

	// the grill commits one example per line and places each
	for key, txtar := range map[string]string{"line-1/a": wholeTitle, "line-2/a": shortTitle} {
		dir := filepath.Join("goals", goal, key)
		if err := os.MkdirAll(filepath.Join(r.Dir, dir), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(r.Dir, dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
			t.Fatal(err)
		}
		r.WriteExec(t, filepath.Join(dir, "probe", "run"), fixtureProbe)
	}
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "examples")
	t.Setenv("HYP_ACTOR", "grill-1")
	t.Setenv("HYP_ROLE", "grill")
	for _, key := range []string{"line-1/a", "line-2/a"} {
		if code, out, errs := run(t, "place", goal, key); code != 0 {
			t.Fatalf("place %s: code %d, stdout %q, stderr %q", key, code, out, errs)
		}
	}
	_, out, _ = run(t, "status")
	if !strings.Contains(out, goal+"  open  runs: 0 of 3 used\nstop: none\n") {
		t.Fatalf("the Goal is not open after both placements:\n%s", out)
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

	// the only candidate is the fix: the runner selects with the printed line
	status := mustStatus(t)
	if !strings.Contains(status, "stop: none") {
		t.Fatalf("status before the select:\n%s", status)
	}
	selectArgs := readyLine(t, status, "runner", "hyp select "+goal+" ")
	code, out, errs = run(t, selectArgs...)
	if code != 0 || !strings.HasPrefix(out, "# "+goal+": Show the whole title.") ||
		!strings.Contains(out, "## line-1 (gap): status shows the whole title") ||
		!strings.Contains(out, "selected commit: "+fix) {
		t.Fatalf("select: code %d, stdout %q, stderr %q", code, out, errs)
	}
	status = mustStatus(t)
	if !strings.Contains(status, "stop: selected") || !strings.Contains(status, "selected commit: "+fix) {
		t.Fatalf("status after the select:\n%s", status)
	}
	if got := testutil.Snapshot(t, r.Dir); got != afterCommit {
		t.Fatal("hyp changed the repository's git state between the last commit and the select")
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
