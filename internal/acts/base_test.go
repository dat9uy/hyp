package acts

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
)

// own commits the runner's change on top of from: the counter, and a file
// in the Goal's folder that names the value, which makes the commit the
// runner's. It holds no spec, so no select shows on it.
func (e *runEnv) own(t *testing.T, from, value string) string {
	t.Helper()
	return e.child(t, from, map[string]string{"counter": value + "\n", "goals/" + e.goal + "/change": value + "\n"})
}

// ownSpec is own with a spec in the Goal's spec folder, so the commit can
// be selected.
func (e *runEnv) ownSpec(t *testing.T, from, value string) string {
	t.Helper()
	return e.child(t, from, map[string]string{"counter": value + "\n", "goals/" + e.goal + "/specs/spec.md": "spec " + value + "\n"})
}

// calculated is the Goal as hyp status shows it, with the select this
// clone's Results show.
func (e *runEnv) calculated(t *testing.T) *tables.Goal {
	t.Helper()
	tb, _, err := ReadTables(e.root)
	if err != nil {
		t.Fatal(err)
	}
	return tb.ByID(e.goal)
}

func TestRunOnANewerBaseThatMeasuresTheSameIsRecorded(t *testing.T) {
	e := newRunEnv(t)
	n := e.child(t, e.base, map[string]string{"other": "x\n"})
	cn := e.ownSpec(t, n, "ok")
	if _, err := e.run(cn); err != nil {
		t.Fatal(err)
	}
	res := e.results(t)
	if len(res) != 1 || res[0].Refused != "" || res[0].Base == nil || res[0].Base.Commit != n || len(res[0].Base.Differing) != 0 {
		t.Fatalf("results = %+v, want one Result that names the base %s", res, n)
	}
	g := e.calculated(t)
	if used := tables.RunsUsed(g); used != 1 {
		t.Fatalf("runs used = %d, want 1: measuring a base uses no run", used)
	}
	if g.State != tables.StateSelected || g.Select.Commit != cn {
		t.Fatalf("state = %s, want selected on %s", g.State, cn)
	}
}

// TestACommitOfNoCodeIsNoBase: a commit newer than H0 that is not the
// runner's and changes no code, as an owner's act does, is no base: a run
// that stands on it and on one other fix has that fix as its one base.
func TestACommitOfNoCodeIsNoBase(t *testing.T) {
	e := newRunEnv(t)
	act := e.child(t, e.base, map[string]string{"goals/model/note": "x\n"})
	n := e.child(t, e.base, map[string]string{"other": "x\n"})
	cn := e.ownSpec(t, act, "ok")
	testutil.Git(t, e.root, "checkout", "-q", cn)
	testutil.Git(t, e.root, "merge", "-q", "--no-ff", "-m", "take in the fix", n)
	m := strings.TrimSpace(testutil.Git(t, e.root, "rev-parse", "HEAD"))
	testutil.Git(t, e.root, "checkout", "-q", "main")
	if _, err := e.run(m); err != nil {
		t.Fatal(err)
	}
	res := e.results(t)
	if len(res) != 1 || res[0].Refused != "" || res[0].Base == nil || res[0].Base.Commit != n {
		t.Fatalf("results = %+v, want one Result that names the base %s", res, n)
	}
}

func TestRunOnANewerBaseThatMeasuresDifferentlyIsRefused(t *testing.T) {
	e := newRunEnv(t)
	v := e.child(t, e.base, map[string]string{"counter": "ok\n"})
	cv := e.own(t, v, "ok")
	_, err := e.run(cv)
	refusedBy(t, err, "Same base")
	res := e.results(t)
	if len(res) != 1 || res[0].Refused != tables.RefusedSameBase || len(res[0].Cells) != 0 ||
		res[0].Base == nil || res[0].Base.Commit != v || strings.Join(res[0].Base.Differing, ",") != "line-1/a" || len(res[0].Base.Environment) != 0 {
		t.Fatalf("results = %+v, want one Result refused by Same base that names line-1/a and %s", res, v)
	}
	g := e.calculated(t)
	if tables.RunsUsed(g) != 0 || g.State != tables.StateOpen || len(g.RefusedMoves()) != 1 {
		t.Fatalf("runs used %d, state %s, refused moves %d; want 0, open, 1", tables.RunsUsed(g), g.State, len(g.RefusedMoves()))
	}
	// the Goal stays in work on H0: a run on H0's code is recorded
	if _, err := e.run(e.own(t, e.base, "ok")); err != nil {
		t.Fatal(err)
	}
}

func TestRunOnAMergeOfTheSelectMovesTheGoal(t *testing.T) {
	e := newRunEnv(t)
	c1 := e.ownSpec(t, e.base, "ok")
	if _, err := e.run(c1); err != nil {
		t.Fatal(err)
	}
	if g := e.calculated(t); g.State != tables.StateSelected || g.Select.Commit != c1 {
		t.Fatalf("test setup: the Goal is not selected on c1")
	}
	n := e.child(t, e.base, map[string]string{"other": "x\n"})
	testutil.Git(t, e.root, "checkout", "-q", "--detach", c1)
	testutil.Git(t, e.root, "merge", "-q", "--no-ff", "-m", "merge", n)
	m := testutil.Git(t, e.root, "rev-parse", "HEAD")
	testutil.Git(t, e.root, "checkout", "-q", "main")
	if _, err := e.run(m); err != nil {
		t.Fatal(err)
	}
	res := e.results(t)
	if len(res) != 2 || res[1].Move != c1 || res[1].Base == nil || res[1].Base.Commit != n {
		t.Fatalf("results = %+v, want the move from %s onto the base %s", res, c1, n)
	}
	if g := e.calculated(t); g.State != tables.StateSelected || g.Select.Commit != m {
		t.Fatalf("state = %s, want selected on the merge %s", g.State, m)
	}
	// a commit that is no merge of the select is still refused
	_, err := e.run(e.own(t, e.base, "ok2"))
	refusedBy(t, err, "Selected")
}
