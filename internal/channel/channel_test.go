package channel

import (
	"bytes"
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
		{"add", "goal", "--words", "w", "--line", "gap:a"},                                 // no --base
		{"add", "goal", "--base", "HEAD", "--line", "gap:a"},                               // no --words
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "maybe:a"},             // kind
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "nocolon"},             // no kind
		{"add", "goal", "--base", "nope", "--words", "w", "--line", "gap:a"},               // not a commit
		{"add", "goal", "--base", "HEAD", "--words", "w", "--line", "gap:a", "--bogus"},    // unknown flag
		{"add", "goal", "a", "b", "--base", "HEAD", "--words", "w", "--line", "gap:a"},     // two positionals
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
	for _, c := range []string{"sandbox", "init", "status", "signals", "events"} {
		if code, _, _ := run(t, c); code != 2 {
			t.Errorf("hyp %s: code %d, want 2", c, code)
		}
	}
}

func TestCommandsAreSorted(t *testing.T) {
	got := Commands()
	if !slices.Equal(got, []string{"add", "help", "version"}) {
		t.Fatalf("Commands = %v", got)
	}
}
