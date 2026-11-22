package channel

import (
	"bytes"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

func run(t *testing.T, args ...string) (code int, stdout, stderr string) {
	t.Helper()
	var out, errb bytes.Buffer
	code = Run(args, &out, &errb)
	return code, out.String(), errb.String()
}

func inRepo(t *testing.T) *testutil.Repo {
	t.Helper()
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"main.go": "package main"})
	t.Chdir(r.Dir)
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	return r
}

func TestVersionLine(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != 0 || !strings.HasPrefix(out, "hyp 0.2.0-dev (events v2)") {
		t.Fatalf("code %d, stdout %q", code, out)
	}
}

func TestUnknownCommandExit2(t *testing.T) {
	code, out, errs := run(t, "frobnicate")
	if code != 2 || out != "" || !strings.Contains(errs, "usage: hyp") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	if code, _, _ := run(t); code != 2 {
		t.Fatalf("no command: code %d, want 2", code)
	}
	if code, _, _ := run(t, "add"); code != 2 {
		t.Fatalf("add alone: code %d, want 2", code)
	}
}

func TestHelpExit0(t *testing.T) {
	code, out, _ := run(t, "help")
	if code != 0 || !strings.Contains(out, "add goal") {
		t.Fatalf("code %d, stdout %q", code, out)
	}
}

func TestAddGoalRecordsExit0(t *testing.T) {
	r := inRepo(t)
	before := testutil.Snapshot(t, r.Dir)
	code, out, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "Find it.",
		"--line", "gap:it fails: with a colon", "--line", "guard:it holds")
	if code != 0 || !strings.Contains(out, "G-") || !strings.Contains(out, "line-1 gap: it fails: with a colon") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("add goal changed the repository's git state")
	}
}

func TestAddGoalRefusalExit1(t *testing.T) {
	inRepo(t)
	code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "guard:it holds")
	if code != 1 || !strings.Contains(errs, "Whole lines") {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
}

func TestAddGoalReplaceForm(t *testing.T) {
	inRepo(t)
	_, out, _ := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a")
	id := strings.Fields(out)[1]
	code, out, errs := run(t, "add", "goal", id, "--base", "HEAD", "--words", "new", "--line", "gap:a", "--line", "gap:b")
	if code != 0 || !strings.Contains(out, id) || !strings.Contains(out, "replaced") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
}

func TestWriteWithoutActorExit2(t *testing.T) {
	inRepo(t)
	t.Setenv("HYP_ACTOR", "")
	code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a")
	if code != 2 || !strings.Contains(errs, "actor") {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "")
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a"); code != 2 || !strings.Contains(errs, "role") {
		t.Fatalf("no role: code %d, stderr %q", code, errs)
	}
	// the flag beats the environment
	t.Setenv("HYP_ROLE", "king")
	if code, _, _ := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a", "--role", "owner"); code != 0 {
		t.Fatalf("--role owner: code %d", code)
	}
}

func TestAddGoalBadArgumentsExit2(t *testing.T) {
	inRepo(t)
	for _, args := range [][]string{
		{"add", "goal", "--words", "w", "--line", "gap:a"},                              // no --base
		{"add", "goal", "--base", "HEAD", "--line", "gap:a"},                            // no --words
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "maybe:a"},          // kind
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "nocolon"},          // no kind
		{"add", "goal", "--base", "nope", "--words", "w", "--line", "gap:a"},            // not a commit
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a", "--bogus"}, // unknown flag
		{"add", "goal", "a", "b", "--base", "HEAD", "--words", "w", "--line", "gap:a"},  // two positionals
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a", "--why", strings.Repeat("x", 601)},
	} {
		if code, _, errs := run(t, args...); code != 2 {
			t.Errorf("%v: code %d, stderr %q; want 2", args, code, errs)
		}
	}
}

func TestOutsideRepoExit3(t *testing.T) {
	t.Setenv("GIT_CEILING_DIRECTORIES", t.TempDir())
	t.Chdir(t.TempDir())
	t.Setenv("HYP_ACTOR", "owner-1")
	t.Setenv("HYP_ROLE", "owner")
	if code, _, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a"); code != 3 {
		t.Fatalf("code %d, stderr %q; want 3", code, errs)
	}
}

func TestV1CommandGone(t *testing.T) {
	for _, c := range []string{"sandbox", "init", "signals", "events"} {
		if code, _, _ := run(t, c); code != 2 {
			t.Errorf("hyp %s: code %d, want 2", c, code)
		}
	}
}

func TestCommandsAreSorted(t *testing.T) {
	got := Commands()
	if !slices.Equal(got, []string{"add", "help", "land", "place", "prune", "run", "select", "show", "status", "try", "unselect", "version"}) {
		t.Fatalf("Commands = %v", got)
	}
}

const fixtureProbe = `#!/bin/sh
o=$(mktemp); sh -c "$HYP_RUN" >"$o"; c=$?; cat "$o"; [ -s "$o" ] && [ -n "$(tail -c1 "$o")" ] && echo; rm -f "$o"; echo "exit $c"
`

// goalWithExample adds a one-line gap Goal as the owner, commits an example
// folder for line-1/a whose floor fails on H0, and switches to the grill. It
// returns the Goal id.
func goalWithExample(t *testing.T, r *testutil.Repo, commit bool) string {
	t.Helper()
	_, out, errs := run(t, "add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:it fails")
	fields := strings.Fields(out)
	if len(fields) < 2 {
		t.Fatalf("add goal printed %q, %q", out, errs)
	}
	goal := fields[1]
	dir := filepath.Join("goals", goal, "line-1", "a")
	if err := os.MkdirAll(filepath.Join(r.Dir, dir), 0o755); err != nil {
		t.Fatal(err)
	}
	txtar := "-- words --\nw\n-- run --\necho hi\n-- floor: contains --\nnope\n"
	if err := os.WriteFile(filepath.Join(r.Dir, dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
		t.Fatal(err)
	}
	r.WriteExec(t, filepath.Join(dir, "probe", "run"), fixtureProbe)
	if commit {
		testutil.Git(t, r.Dir, "add", "-A")
		testutil.Git(t, r.Dir, "commit", "-q", "-m", "example")
	}
	t.Setenv("HYP_ACTOR", "grill-1")
	t.Setenv("HYP_ROLE", "grill")
	return goal
}

func TestPlaceRecordsExit0(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, true)
	before := testutil.Snapshot(t, r.Dir)
	code, out, errs := run(t, "place", goal)
	if code != 0 || !strings.Contains(out, "placed "+goal+": 1 example\nline-1/a @1  hash: ") || !strings.Contains(out, "H0: failed (exit 0)") || !strings.Contains(out, goal+" is open") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("place changed the repository's git state")
	}
}

func TestPlaceFlagsMayFollowThePositionals(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, true)
	t.Setenv("HYP_ROLE", "runner")
	code, out, errs := run(t, "place", goal, "--role", "grill", "--why", "first")
	if code != 0 {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
}

func TestPlaceRefusalExit1(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, false)
	code, out, errs := run(t, "place", goal)
	if code != 1 || !strings.HasPrefix(out, goal+"  written  runs: 0 of 3 used\n") || !strings.Contains(errs, "Committed") {
		t.Fatalf("code %d, stdout %q, stderr %q; want the refusal on stderr and the state block on stdout", code, out, errs)
	}
	t.Setenv("HYP_ROLE", "runner")
	if code, _, errs := run(t, "place", goal); code != 1 || !strings.Contains(errs, "Grill places") {
		t.Fatalf("runner: code %d, stderr %q", code, errs)
	}
}

func TestPlaceBadArgumentsExit2(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, true)
	for _, args := range [][]string{
		{"place"},
		{"place", goal, "line-1/a"}, // the old form, one example per place
		{"place", goal, "line-1"},
		{"place", "G-NOPE00"},
		{"place", goal, "--bogus"},
		{"place", goal, "--why", strings.Repeat("x", 601)},
	} {
		if code, _, errs := run(t, args...); code != 2 {
			t.Errorf("%v: code %d, stderr %q; want 2", args, code, errs)
		}
	}
	t.Setenv("HYP_ROLE", "")
	if code, _, errs := run(t, "place", goal); code != 2 || !strings.Contains(errs, "role") {
		t.Fatalf("no role: code %d, stderr %q", code, errs)
	}
}

func TestPlaceWithoutStoreExit1(t *testing.T) {
	inRepo(t)
	t.Setenv("HYP_ROLE", "grill")
	if code, _, errs := run(t, "place", "G-AAAAAA"); code != 1 || !strings.Contains(errs, "no store; run hyp add goal") {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
}

func TestPlaceProbeFailureExit3(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, false)
	r.WriteExec(t, filepath.Join("goals", goal, "line-1", "a", "probe", "run"), "#!/bin/sh\nexit 7\n")
	testutil.Git(t, r.Dir, "add", "-A")
	testutil.Git(t, r.Dir, "commit", "-q", "-m", "example")
	if code, _, errs := run(t, "place", goal); code != 3 {
		t.Fatalf("code %d, stderr %q; want 3", code, errs)
	}
}

// placed adds a Goal with a placed example and returns its id and a new commit
// on top of it, ready for run.
func placed(t *testing.T) (r *testutil.Repo, goal, commit string) {
	t.Helper()
	r = inRepo(t)
	goal = goalWithExample(t, r, true)
	if code, out, errs := run(t, "place", goal); code != 0 {
		t.Fatalf("place: code %d, stdout %q, stderr %q", code, out, errs)
	}
	commit = r.Commit(t, map[string]string{"x.go": "package x"})
	t.Setenv("HYP_ACTOR", "runner-1")
	t.Setenv("HYP_ROLE", "runner")
	return r, goal, commit
}

func TestRunRecordsExit0(t *testing.T) {
	r, goal, commit := placed(t)
	before := testutil.Snapshot(t, r.Dir)
	code, out, errs := run(t, "run", commit, goal+"/line-1")
	if code != 0 || !strings.Contains(out, "line-1/a") || !strings.Contains(out, "failed") || !strings.Contains(out, "runs: 1 of 3 used\n") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("run changed the repository's git state")
	}
}

func TestRunFlagsMayFollowThePositionals(t *testing.T) {
	_, goal, commit := placed(t)
	t.Setenv("HYP_ROLE", "")
	if code, out, errs := run(t, "run", commit, goal+"/line-1", "--role", "runner", "--why", "first"); code != 0 {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
}

func TestRunRefusalExit1(t *testing.T) {
	_, goal, _ := placed(t)
	code, out, errs := run(t, "run", "no-such-commit", goal+"/line-1")
	if code != 1 || !strings.HasPrefix(out, goal+"  open  runs: 0 of 3 used\n") || !strings.Contains(errs, "Bytes") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
}

func TestRunBadArgumentsExit2(t *testing.T) {
	_, goal, commit := placed(t)
	for _, args := range [][]string{
		{"run"},
		{"run", commit},
		{"run", commit, goal + "/line-1", "extra"},
		{"run", commit, goal},
		{"run", commit, "G-NOPE00/line-1"},
		{"run", commit, goal + "/line-9"},
		{"run", commit, goal + "/line-1", "--bogus"},
		{"run", commit, goal + "/line-1", "--why", strings.Repeat("x", 601)},
	} {
		if code, _, errs := run(t, args...); code != 2 {
			t.Errorf("%v: code %d, stderr %q; want 2", args, code, errs)
		}
	}
}

func TestRunWithoutStoreExit1(t *testing.T) {
	inRepo(t)
	t.Setenv("HYP_ROLE", "runner")
	if code, _, errs := run(t, "run", "HEAD", "G-AAAAAA/line-1"); code != 1 || !strings.Contains(errs, "no store; run hyp add goal") {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
}
