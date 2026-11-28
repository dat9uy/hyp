package tables

import (
	"encoding/json"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/testutil"
)

func refusedWith(t *testing.T, err error, guard string) {
	t.Helper()
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != guard {
		t.Fatalf("err = %v, want a %s refusal", err, guard)
	}
}

// openGoal is an open Goal on the abstract commit "c" and H0 row tree "t",
// with line-1/a placed under hashA and failing on H0, followed by evs.
func openGoal(t *testing.T, evs ...events.Event) (Tables, *Goal) {
	t.Helper()
	all := append([]events.Event{goalAdd(t, id1, "", "w", gap), placeEv(t, idP1, "line-1", "a", hashA, "failed")}, evs...)
	tb := Replay(all)
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if g.State != StateOpen {
		t.Fatalf("test setup: Goal is %s, want open", g.State)
	}
	return tb, g
}

func run(t *testing.T, tree, verdict string) events.Event {
	t.Helper()
	return result(t, id1, "line-1", "commit-"+tree, tree, cell(hashA, verdict))
}

func TestGuardOpenSkipsH0(t *testing.T) {
	// line-2 has no placed example, so the Goal is written.
	tb := Replay([]events.Event{goalAdd(t, id1, "", "w", gap, guard), placeEv(t, idP1, "line-1", "a", hashA, "failed")})
	g := tb.Goals[id1]
	if g.State != StateWritten {
		t.Fatalf("test setup: Goal is %s, want written", g.State)
	}
	if err := GuardOpen(g, "t", "line-1"); err != nil {
		t.Fatalf("a run on H0 with a placed example was refused: %v", err)
	}
	refusedWith(t, GuardOpen(g, "other", "line-1"), "Open")
	refusedWith(t, GuardOpen(g, "t", "line-2"), "Open") // nothing to run, even on H0
	_, open := openGoal(t)
	if err := GuardOpen(open, "other", "line-1"); err != nil {
		t.Fatalf("an open Goal refused a run: %v", err)
	}
}

type fakeAnc struct {
	has  map[string]bool
	desc map[[2]string]bool // [a, b]: a is an ancestor of b
}

func (f fakeAnc) HasCommit(c string) (bool, error) { return f.has[c], nil }

func (f fakeAnc) IsAncestor(a, b string) (bool, error) {
	if !f.has[a] || !f.has[b] {
		return false, fault.EnvFault(fault.Code("missing"))
	}
	return a == b || f.desc[[2]string{a, b}], nil
}

func TestGuardBytesMissingCommit(t *testing.T) {
	anc := fakeAnc{has: map[string]bool{"c1": true}}
	if err := GuardBytes(anc, "c1"); err != nil {
		t.Fatalf("a commit this clone holds was refused: %v", err)
	}
	refusedWith(t, GuardBytes(anc, "c2"), "Bytes")
}

func TestGuardLeafOnlyRefusesAncestor(t *testing.T) {
	_, g := openGoal(t, run(t, "t1", "met")) // a row at commit-t1; H0 is "c"
	anc := fakeAnc{
		has:  map[string]bool{"c": true, "commit-t1": true, "older": true, "other": true},
		desc: map[[2]string]bool{{"older", "commit-t1"}: true, {"c", "commit-t1"}: true},
	}
	if err := GuardLeafOnly(g, anc, "other"); err != nil {
		t.Fatalf("a commit no row descends from was refused: %v", err)
	}
	if err := GuardLeafOnly(g, anc, "c"); err != nil {
		t.Fatalf("a run on H0 was refused: %v", err)
	}
	refusedWith(t, GuardLeafOnly(g, anc, "older"), "Leaf only")
	// A commit of a row that this clone lacks is an environment problem.
	anc.has["commit-t1"] = false
	err := GuardLeafOnly(g, anc, "other")
	if err == nil || fault.KindOf(err) != fault.Environment || !testutil.Mentions(err, "commit-t1") {
		t.Fatalf("err = %v, want an environment error naming commit-t1", err)
	}
}

func TestGuardStableRefusesDifferingVerdict(t *testing.T) {
	_, g := openGoal(t, run(t, "t1", "met"))
	same := []events.ExampleCell{cell(hashA, "met")}
	if err := GuardStable(g, "line-1", "t1", same); err != nil {
		t.Fatalf("a repeat with the same verdict was refused: %v", err)
	}
	refusedWith(t, GuardStable(g, "line-1", "t1", []events.ExampleCell{cell(hashA, "failed")}), "Stable")
	// A tree with no cell yet has nothing to differ from.
	if err := GuardStable(g, "line-1", "t2", []events.ExampleCell{cell(hashA, "failed")}); err != nil {
		t.Fatalf("a first run was refused: %v", err)
	}
}

func TestGuardStableCountsPlacementCellOnH0(t *testing.T) {
	_, g := openGoal(t) // the placement's H0 cell: failed on tree "t"
	if err := GuardStable(g, "line-1", "t", []events.ExampleCell{cell(hashA, "failed")}); err != nil {
		t.Fatalf("a repeat on H0 with the same verdict was refused: %v", err)
	}
	refusedWith(t, GuardStable(g, "line-1", "t", []events.ExampleCell{cell(hashA, "met")}), "Stable")
}

func TestGuardBudgetRefusesFourthTree(t *testing.T) {
	_, g := openGoal(t, run(t, "t1", "met"), run(t, "t2", "met"))
	if err := GuardBudget(g, "t3"); err != nil {
		t.Fatalf("the third tree was refused: %v", err)
	}
	_, g = openGoal(t, run(t, "t1", "met"), run(t, "t2", "met"), run(t, "t3", "met"))
	if n := RunsUsed(g); n != 3 {
		t.Fatalf("RunsUsed = %d, want 3", n)
	}
	if err := GuardBudget(g, "t2"); err != nil {
		t.Fatalf("a tree already counted was refused: %v", err)
	}
	refusedWith(t, GuardBudget(g, "t4"), "Budget")
}

func TestGuardBudgetIgnoresH0(t *testing.T) {
	_, g := openGoal(t, run(t, "t", "failed"), run(t, "t1", "met"), run(t, "t", "failed"))
	if n := RunsUsed(g); n != 1 {
		t.Fatalf("RunsUsed = %d, want 1: H0's tree never counts", n)
	}
	_, g = openGoal(t, run(t, "t1", "met"), run(t, "t2", "met"), run(t, "t3", "met"))
	if err := GuardBudget(g, "t"); err != nil {
		t.Fatalf("a run on H0 was refused by Budget: %v", err)
	}
}

func TestGuardBudgetResetsOnChangedExample(t *testing.T) {
	_, g := openGoal(t, run(t, "t1", "met"), run(t, "t2", "met"), run(t, "t3", "met"),
		placeEv(t, idP2, "line-1", "b", hashB, "failed"))
	if n := RunsUsed(g); n != 0 {
		t.Fatalf("RunsUsed = %d after a new example key, want 0", n)
	}
	if err := GuardBudget(g, "t4"); err != nil {
		t.Fatalf("a fourth tree after the boundary was refused: %v", err)
	}
}

func TestGuardBudgetCountsOldTreeRerunAfterBoundary(t *testing.T) {
	_, g := openGoal(t, run(t, "t1", "met"), placeEv(t, idP2, "line-1", "b", hashB, "failed"))
	if n := RunsUsed(g); n != 0 {
		t.Fatalf("RunsUsed = %d right after the boundary, want 0", n)
	}
	_, g = openGoal(t, run(t, "t1", "met"), placeEv(t, idP2, "line-1", "b", hashB, "failed"), run(t, "t1", "met"))
	if n := RunsUsed(g); n != 1 {
		t.Fatalf("RunsUsed = %d after t1 ran again, want 1: it is run against the new evidence", n)
	}
}

func TestGuardBudgetSkipsStableRefusal(t *testing.T) {
	refused := resultRefused(t, id1, "line-1", "commit-t3", "t3", "Stable", cell(hashA, "failed"))
	_, g := openGoal(t, run(t, "t1", "met"), run(t, "t2", "met"), refused)
	if n := RunsUsed(g); n != 2 {
		t.Fatalf("RunsUsed = %d, want 2: a Result refused by Stable counts nothing", n)
	}
	if err := GuardBudget(g, "t4"); err != nil {
		t.Fatalf("a third counted tree was refused: %v", err)
	}
}

// fakeTree is a GoalTree with fixed tree hashes and writers.
type fakeTree struct {
	hashes  map[string]string
	writers map[string]string
	owner   map[string]bool
}

func (f fakeTree) TreeHash(p string) (string, error) { return f.hashes[p], nil }
func (f fakeTree) Writer(p string) (string, bool, error) {
	return f.writers[p], f.owner[p], nil
}

func TestGuardOwnPathRefusesOtherGoal(t *testing.T) {
	g := &Goal{ID: "G-NXHHVA"}
	tree := fakeTree{writers: map[string]string{"goals/G-OTHER00/line-1/a/probe/run": "c2"}}
	own := []gitread.OwnPath{{Path: "goals/G-NXHHVA/line-1/a/example.txtar"}, {Path: "goals/model/m.tla"}}
	if err := GuardOwnPath(g, ownGoals, tree, own); err != nil {
		t.Fatalf("the Goal's own folder was refused: %v", err)
	}
	err := GuardOwnPath(g, ownGoals, tree, append(own, gitread.OwnPath{Path: "goals/G-OTHER00/line-1/a/probe/run"}))
	refusedWith(t, err, "Own path")
	if !testutil.Mentions(err, "goals/G-OTHER00/line-1/a/probe/run") || !testutil.Mentions(err, "c2") {
		t.Fatalf("err = %v, want the path and the commit that wrote it", err)
	}
	// a file the commit lacks is no write
	if err := GuardOwnPath(g, ownGoals, tree, []gitread.OwnPath{{Path: "goals/G-OTHER00/specs/s", Deleted: true}}); err != nil {
		t.Fatalf("a deleted file of another Goal was refused: %v", err)
	}
}

// TestGuardOwnPathReadsTheStore: another Goal's example folder passes when
// it holds a tree an act of that Goal recorded, and its other files pass when
// an owner's act wrote them.
func TestGuardOwnPathReadsTheStore(t *testing.T) {
	other := &Goal{ID: "G-OTHER00", Lines: []Line{{Key: "line-1"}},
		Placements: map[string][]Placement{"line-1/a": {{Key: "line-1/a", Hash: "h1"}}}}
	tb := Tables{Goals: map[string]*Goal{"a": {ID: "G-NXHHVA"}, "b": other}}
	g := tb.Goals["a"]
	ex := gitread.OwnPath{Path: "goals/G-OTHER00/line-1/a/example.txtar"}
	spec := gitread.OwnPath{Path: "goals/G-OTHER00/specs/model.tla"}
	tree := fakeTree{
		hashes:  map[string]string{"goals/G-OTHER00/line-1/a": "h1"},
		writers: map[string]string{ex.Path: "w1", spec.Path: "w2"},
		owner:   map[string]bool{spec.Path: true},
	}
	if err := GuardOwnPath(g, tb, tree, []gitread.OwnPath{ex, spec}); err != nil {
		t.Fatalf("bytes the owner's acts recorded were refused: %v", err)
	}
	tree.hashes["goals/G-OTHER00/line-1/a"] = "h2"
	refusedWith(t, GuardOwnPath(g, tb, tree, []gitread.OwnPath{ex}), "Own path")
	tree.owner[spec.Path] = false
	err := GuardOwnPath(g, tb, tree, []gitread.OwnPath{spec})
	refusedWith(t, err, "Own path")
	if !testutil.Mentions(err, "w2") {
		t.Fatalf("err = %v, want the commit that wrote the spec", err)
	}
	refusedWith(t, GuardOwnPath(g, tb, tree, []gitread.OwnPath{{Path: "goals/draft-x/f"}}), "Own path")
}

func TestOwnPathAllowsMainWithOwnerActs(t *testing.T) {
	r := testutil.NewRepo(t)
	h0 := r.Commit(t, map[string]string{"main.go": "1"})
	// the owner's act on another Goal, on main: its spec and an event
	r.Commit(t, map[string]string{".tree/events/e1": "{}", "goals/G-OTHER00/specs/spec.md": "x"})
	testutil.Git(t, r.Dir, "checkout", "-q", "-b", "runner")
	tip := r.Commit(t, map[string]string{"own.go": "1", "goals/G-NXHHVA/specs/spec.md": "s"})
	g := &Goal{ID: "G-NXHHVA"}
	own, err := gitread.GoalChange(r.Dir, tip, h0)
	if err != nil {
		t.Fatal(err)
	}
	if len(own) != 2 {
		t.Fatalf("goal change = %v, want the two specs", own)
	}
	if err := GuardOwnPath(g, ownGoals, &gitread.CommitTree{Dir: r.Dir, Commit: tip, H0: h0}, own); err != nil {
		t.Fatalf("an owner's act on another Goal refused the runner: %v", err)
	}
	// A runner commit that writes another Goal's folder itself is refused.
	bad := r.Commit(t, map[string]string{"goals/G-OTHER00/specs/spec.md": "y"})
	own, err = gitread.GoalChange(r.Dir, bad, h0)
	if err != nil {
		t.Fatal(err)
	}
	err = GuardOwnPath(g, ownGoals, &gitread.CommitTree{Dir: r.Dir, Commit: bad, H0: h0}, own)
	refusedWith(t, err, "Own path")
	if !testutil.Mentions(err, bad) {
		t.Fatalf("err = %v, want the runner's commit %s", err, bad)
	}
}

func TestGuardBudgetIgnoresAnEarlierH0(t *testing.T) {
	// line-2 has no example, so the Goal stays written and a Draft replace may
	// move its base from tree "t" to tree "n".
	p, err := json.Marshal(events.GoalAdd{Role: "owner", Words: "w", Base: "c2", BaseTree: "n", Budget: 3, Lines: []events.Line{gap}})
	if err != nil {
		t.Fatal(err)
	}
	replace := events.Event{V: 2, ID: "01M661QG800000000000REPLAC", Actor: "a", Type: events.TypeGoalAdd, Node: id1, Payload: p}
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap, guard),
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		run(t, "t", "failed"), // a run on H0 when it ran
		replace,
	})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	if n := RunsUsed(tb.Goals[id1]); n != 0 {
		t.Fatalf("RunsUsed = %d, want 0: a tree that was H0 when it ran never counts", n)
	}
}

var ownGoals = Tables{Goals: map[string]*Goal{"a": {ID: "G-NXHHVA"}, "b": {ID: "G-OTHER00"}}}

func TestGuardOwnPathFolderOfNoGoal(t *testing.T) {
	g := &Goal{ID: "G-NXHHVA"}
	tree := fakeTree{writers: map[string]string{"goals/draft-x/f": "c1"}}
	err := GuardOwnPath(g, ownGoals, tree, []gitread.OwnPath{{Path: "goals/draft-x/f"}})
	refusedWith(t, err, "Own path")
	if !testutil.Mentions(err, "no-goal-folder") || testutil.Mentions(err, "other-goal-path") {
		t.Fatalf("err = %v, want the folder of no Goal, not another Goal's path", err)
	}
	if err := GuardOwnPath(g, ownGoals, tree, []gitread.OwnPath{{Path: "goals/draft-x/f", Deleted: true}}); err != nil {
		t.Fatalf("a delete of a folder of no Goal was refused: %v", err)
	}
}
