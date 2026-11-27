package acts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/gitread"
	"github.com/dat9uy/hyp/internal/logstore"
	"github.com/dat9uy/hyp/internal/tables"
	"github.com/dat9uy/hyp/internal/testutil"
)

func opts(dir, base string, lines ...LineInput) AddGoalOpts {
	return AddGoalOpts{Dir: dir, Actor: "owner-1", Role: "owner", Words: "Find the gap.", Base: base, Lines: lines}
}

var gapLine = LineInput{Kind: events.KindGap, Text: "it fails"}

func TestAddGoalRecords(t *testing.T) {
	r := testutil.NewRepo(t)
	c := r.Commit(t, map[string]string{"main.go": "package main", "goals/x": "x"})
	sub := filepath.Join(r.Dir, "internal")
	if err := os.MkdirAll(sub, 0o755); err != nil {
		t.Fatal(err)
	}
	rep, err := AddGoal(opts(sub, "HEAD", gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"}))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(rep.Goal, "G-") || len(rep.Lines) != 2 || rep.Lines[1].Key != "line-2" {
		t.Fatalf("report = %+v", rep)
	}
	if rep.Base != c {
		t.Fatalf("base = %s, want the full sha %s", rep.Base, c)
	}
	files, err := filepath.Glob(filepath.Join(r.Dir, ".tree", "events", "owner-1", "*.json"))
	if err != nil || len(files) != 1 {
		t.Fatalf("events files = %v, %v; want one file of the actor at the top level", files, err)
	}
	data, err := os.ReadFile(files[0])
	if err != nil || strings.Count(string(data), "\n") != 1 {
		t.Fatalf("events file = %q, %v; want one line", data, err)
	}
	evs, errs, err := logstore.Events(r.Dir)
	if err != nil || len(errs) != 0 || len(evs) != 1 {
		t.Fatalf("Events = %d, %v, %v", len(evs), errs, err)
	}
	var p events.GoalAdd
	if err := events.Decode(evs[0], &p); err != nil {
		t.Fatal(err)
	}
	if evs[0].Node != evs[0].ID || p.Budget != 3 || p.Role != "owner" || p.Base != c || len(p.BaseTree) != 40 {
		t.Fatalf("event = %+v payload = %+v", evs[0], p)
	}
	tb := tables.Replay(evs)
	if g := tb.ByID(rep.Goal); g == nil || g.Words != "Find the gap." {
		t.Fatalf("replay does not hold %s: %+v", rep.Goal, tb.Goals)
	}
}

func TestAddGoalRefusesNoGapLine(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	_, err := AddGoal(opts(r.Dir, "HEAD", LineInput{Kind: events.KindGuard, Text: "it holds"}))
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Whole lines" {
		t.Fatalf("err = %v, want a Whole lines refusal", err)
	}
	if _, err := os.Stat(filepath.Join(r.Dir, ".tree")); !os.IsNotExist(err) {
		t.Fatalf("a refused first act left a store behind: %v", err)
	}
}

func TestAddGoalBadBaseIsBadArgs(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	for _, base := range []string{"nope", strings.Repeat("0", 40), "", "--output=x"} {
		_, err := AddGoal(opts(r.Dir, base, gapLine))
		if err == nil || fault.KindOf(err) != fault.BadArguments {
			t.Errorf("base %q: err = %v, want bad arguments", base, err)
		}
	}
}

func TestAddGoalBadActorOrRoleIsBadArgs(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	for _, mod := range []func(*AddGoalOpts){
		func(o *AddGoalOpts) { o.Actor = "" },
		func(o *AddGoalOpts) { o.Actor = "Bad Actor" },
		func(o *AddGoalOpts) { o.Role = "" },
		func(o *AddGoalOpts) { o.Role = "king" },
	} {
		o := opts(r.Dir, "HEAD", gapLine)
		mod(&o)
		if _, err := AddGoal(o); fault.KindOf(err) != fault.BadArguments {
			t.Errorf("%+v: err = %v, want bad arguments", o, err)
		}
	}
}

func TestAddGoalOutsideRepoIsEnvironment(t *testing.T) {
	t.Setenv("GIT_CEILING_DIRECTORIES", filepath.Dir(t.TempDir()))
	_, err := AddGoal(opts(t.TempDir(), "HEAD", gapLine))
	if fault.KindOf(err) != fault.Environment {
		t.Fatalf("err = %v, want an environment error", err)
	}
}

func TestAddGoalReplacesWrittenGoal(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	first, err := AddGoal(opts(r.Dir, "HEAD", gapLine))
	if err != nil {
		t.Fatal(err)
	}
	o := opts(r.Dir, "HEAD", gapLine, LineInput{Kind: events.KindGuard, Text: "keeps"})
	o.Goal, o.Words = first.Goal, "Better words."
	rep, err := AddGoal(o)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Goal != first.Goal || !rep.Replaced {
		t.Fatalf("report = %+v, want the replaced Goal %s", rep, first.Goal)
	}
	evs, _, _ := logstore.Events(r.Dir)
	tb := tables.Replay(evs)
	g := tb.ByID(first.Goal)
	if len(tb.Goals) != 1 || len(tb.RefusedAtReplay) != 0 || g == nil || g.Words != "Better words." || len(g.Lines) != 2 {
		t.Fatalf("tables = %+v", tb)
	}
	o.Goal = "G-NOPE00"
	if _, err := AddGoal(o); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("replace of an unknown Goal: %v", err)
	}
}

func TestAddGoalFromAnUnknownGoalIsBadArgs(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	if _, err := AddGoal(opts(r.Dir, "HEAD", gapLine)); err != nil {
		t.Fatal(err)
	}
	o := opts(r.Dir, "HEAD", gapLine)
	o.From = "G-NOPE00"
	if _, err := AddGoal(o); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("from an unknown Goal: %v", err)
	}
}

func TestAddGoalWritesNoGitState(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"a": "a"})
	before := testutil.Snapshot(t, r.Dir)
	rep, err := AddGoal(opts(r.Dir, "HEAD", gapLine))
	if err != nil || !strings.HasPrefix(rep.Goal, "G-") {
		t.Fatalf("the act did not run: %+v, %v", rep, err)
	}
	if testutil.Snapshot(t, r.Dir) != before {
		t.Fatal("add goal changed the repository's git state")
	}
}

// draft writes a draft example folder x in the clone's trace for each key and
// returns HEAD, the commit that is H0.
func draft(t *testing.T, r *testutil.Repo, examples map[string]string) string {
	t.Helper()
	tr, err := logstore.TraceOf(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	for key, txtar := range examples {
		dir := filepath.Join(tr.DraftDir("x"), key)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "example.txtar"), []byte(txtar), 0o644); err != nil {
			t.Fatal(err)
		}
		if err := os.MkdirAll(filepath.Join(dir, "probe"), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "probe", "run"), []byte(fixtureProbe), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	return strings.TrimSpace(testutil.Git(t, r.Dir, "rev-parse", "HEAD"))
}

// An add with a draft carries every example with its H0 verdict,
// so the Goal is open from git alone, and copies the draft to the Goal's folder.
func TestAddGoalWithExamplesIsOpen(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"main.go": "package main\n"})
	base := draft(t, r, map[string]string{"line-1/a": failingOnH0, "line-2/a": metOnH0})
	o := opts(r.Dir, base, gapLine, LineInput{Kind: events.KindGuard, Text: "it holds"})
	o.Draft = "x"
	rep, err := AddGoal(o)
	if err != nil {
		t.Fatal(err)
	}
	if rep.State != tables.StateOpen || len(rep.Placed) != 2 || rep.Folder != "goals/"+rep.Goal {
		t.Fatalf("report = %+v", rep)
	}
	evs, _, err := logstore.Load(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	g := tables.Replay(evs).ByID(rep.Goal)
	if g == nil || g.State != tables.StateOpen || len(g.Placed()) != 2 {
		t.Fatalf("replay of git alone: %+v", g)
	}
	for _, ev := range evs {
		if ev.Type == events.TypeExamplesPlace {
			t.Fatal("the add wrote a place event")
		}
	}
	for _, f := range []string{"line-1/a/example.txtar", "line-2/a/probe/run"} {
		if _, err := os.Stat(filepath.Join(r.Dir, "goals", rep.Goal, f)); err != nil {
			t.Fatalf("not copied: %s: %v", f, err)
		}
	}
}

// An example whose H0 verdict is not what its line needs refuses the add,
// and nothing is written.
func TestAddGoalWithExamplesRefusesWrongVerdict(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"main.go": "package main\n"})
	base := draft(t, r, map[string]string{"line-1/a": metOnH0})
	o := opts(r.Dir, base, gapLine)
	o.Draft = "x"
	_, err := AddGoal(o)
	if fault.KindOf(err) != fault.Refusal {
		t.Fatalf("err = %v, want a refusal", err)
	}
	if logstore.Find(r.Dir) == nil {
		evs, _, _ := logstore.Load(r.Dir)
		if len(evs) != 0 {
			t.Fatalf("%d events after a refused add", len(evs))
		}
	}
	o.Draft = "../outside"
	if _, err := AddGoal(o); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("err = %v, want a bad argument for a folder outside the repository", err)
	}
}

// The draft is read once: a Probe that changes the draft while the add runs
// H0 does not change what the Goal's folder holds.
func TestAddGoalFolderHoldsTheBytesItHashed(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"main.go": "package main\n"})
	base := draft(t, r, map[string]string{"line-1/a": failingOnH0})
	tr, err := logstore.TraceOf(r.Dir)
	if err != nil {
		t.Fatal(err)
	}
	folder := filepath.Join(tr.DraftDir("x"), "line-1", "a")
	// the run section adds a file to the draft while the Probe runs
	changing := "-- words --\nw\n-- run --\necho late > " + filepath.Join(folder, "late") + "\n-- floor: contains --\nnope\n"
	if err := os.WriteFile(filepath.Join(folder, "example.txtar"), []byte(changing), 0o644); err != nil {
		t.Fatal(err)
	}
	o := opts(r.Dir, base, gapLine)
	o.Draft = "x"
	rep, err := AddGoal(o)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(filepath.Join(folder, "late")); err != nil {
		t.Fatalf("test setup: the Probe did not change the draft: %v", err)
	}
	if _, err := os.Stat(filepath.Join(r.Dir, "goals", rep.Goal, "line-1", "a", "late")); err == nil {
		t.Fatal("the Goal's folder holds a file the draft got after the add read it")
	}
	got, err := gitread.DirTreeHash(filepath.Join(r.Dir, "goals", rep.Goal, "line-1", "a"))
	if err != nil || got != rep.Placed[0].Hash {
		t.Fatalf("the folder hashes to %s (%v), the add carries %s", got, err, rep.Placed[0].Hash)
	}
}

func TestAddGoalNeedsTheDraft(t *testing.T) {
	r := testutil.NewRepo(t)
	r.Commit(t, map[string]string{"main.go": "package main\n"})
	o := opts(r.Dir, "HEAD", gapLine)
	o.Draft = "nope"
	if _, err := AddGoal(o); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("err = %v, want a bad argument for a draft that is not there", err)
	}
}
