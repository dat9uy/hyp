package acts

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
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
	data, err := os.ReadFile(filepath.Join(r.Dir, ".tree", "events", "owner-1.jsonl"))
	if err != nil || strings.Count(string(data), "\n") != 1 {
		t.Fatalf("events file = %q, %v; want one line at the top level", data, err)
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
