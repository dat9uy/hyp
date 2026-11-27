package channel

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/testutil"
)

// try prints where the example ran, that nothing was recorded, the verdict
// and exit code, the output and stderr as byte blocks under the names they
// would have in .tree/outputs, and the Goal's state block. It is a read: it
// leaves its trace, which show prints, and changes nothing in the
// repository.
func TestTryPrintsTheCellAndRecordsNothing(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, false)
	before := testutil.Snapshot(t, r.Dir)
	code, out, errs := run(t, "try", goal, "line-1/a")
	h0 := strings.TrimSpace(testutil.Git(t, r.Dir, "rev-parse", "HEAD"))
	for _, want := range []string{
		"tried " + goal + " line-1/a on H0 " + h0 + "\n",
		"nothing is recorded: no event, no output file, no run of the budget\n",
		"example: goals/" + goal + "/line-1/a as it is on disk, hash ",
		"verdict: failed, exit 0\n",
		"output 8f434346648f6b96df89dda901c5176b10a6d83961dd3c1ac88b59b2dc327aa4, 2 bytes:\nhi\n",
		"stderr e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855, 0 bytes\n",
		goal + "  written  runs: 0 of 3 used\n",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("code %d, stdout lacks %q:\n%s\nstderr %q", code, want, out, errs)
		}
	}
	if code != 0 || errs != "" {
		t.Fatalf("code %d, stderr %q", code, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("try changed the repository's git state")
	}
	if _, show, _ := run(t, "show", "owner-1"); !strings.Contains(show, "command: hyp try "+goal+" line-1/a\n") || !strings.Contains(show, "verdict: failed, exit 0\n") {
		t.Fatalf("show does not hold the try:\n%s", show)
	}
}

// A draft folder is tried on the commit named, before any Goal holds it:
// try needs no store and records nothing.
func TestTryDraftNeedsNoGoal(t *testing.T) {
	r := inRepo(t)
	if err := os.MkdirAll(filepath.Join(r.Dir, "goals/draft-x/line-1/a"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(r.Dir, "goals/draft-x/line-1/a/example.txtar"), []byte("-- words --\nw\n-- run --\necho hi\n-- floor: equals --\nbye\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	r.WriteExec(t, "goals/draft-x/line-1/a/probe/run", fixtureProbe)
	base := strings.TrimSpace(testutil.Git(t, r.Dir, "rev-parse", "HEAD"))
	before := testutil.Snapshot(t, r.Dir)
	code, out, errs := run(t, "try", "--examples", "goals/draft-x", "--base", base, "line-1/a")
	if code != 0 || !strings.Contains(out, "tried the draft goals/draft-x line-1/a on H0 "+base+"\n") ||
		!strings.Contains(out, "nothing is recorded") || !strings.Contains(out, "verdict: failed, exit 0\n") {
		t.Fatalf("code %d, stdout %q, stderr %q", code, out, errs)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("a draft try changed the repository's git state")
	}
	if code, _, _ := run(t, "try", "--examples", "goals/draft-x", "--base", base, "line-1/a", "--worktree"); code != 2 {
		t.Fatalf("a draft try on another commit: code %d, want 2", code)
	}
}

func TestTryArguments(t *testing.T) {
	r := inRepo(t)
	goal := goalWithExample(t, r, false)
	for _, args := range [][]string{
		{"try", goal},
		{"try", goal, "line-1/a", "extra"},
		{"try", goal, "line-1/b"},
		{"try", goal, "line-1/a", "--worktree=maybe"},
	} {
		if code, out, errs := run(t, args...); code != 2 || out != "" {
			t.Errorf("%v: code %d, stdout %q, stderr %q; want exit 2 and no state block", args, code, out, errs)
		}
	}
	t.Setenv("HYP_ACTOR", "")
	if code, _, errs := run(t, "try", goal, "line-1/a"); code != 2 || !strings.Contains(errs, "try needs an actor") {
		t.Fatalf("no actor: code %d, stderr %q", code, errs)
	}
}
