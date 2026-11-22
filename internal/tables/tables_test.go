package tables

import (
	"encoding/json"
	"sort"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/fault"
	"github.com/dat9uy/hyp/internal/testutil"
)

func goalAdd(t *testing.T, id, node, words string, lines ...events.Line) events.Event {
	t.Helper()
	p, err := json.Marshal(events.GoalAdd{Role: "owner", Words: words, Base: "c", BaseTree: "t", Budget: 3, Lines: lines})
	if err != nil {
		t.Fatal(err)
	}
	if node == "" {
		node = id
	}
	return events.Event{V: 2, ID: id, Actor: "a", Type: events.TypeGoalAdd, Node: node, Payload: p}
}

var (
	gap   = events.Line{Key: "line-1", Text: "it fails", Kind: events.KindGap}
	guard = events.Line{Key: "line-2", Text: "it holds", Kind: events.KindGuard}
)

const (
	id1 = "01M661QG600000000000NXHHVA"
	id2 = "01M661QG700000000000ABCDEF"
)

func TestReplayGoalIDFromULID(t *testing.T) {
	tb := Replay([]events.Event{goalAdd(t, id1, "", "w", gap, guard)})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g, ok := tb.Goals[id1]
	if !ok {
		t.Fatalf("no Goal for the event id; Goals = %v", tb.Goals)
	}
	if g.ID != "G-NXHHVA" || GoalID(id1) != "G-NXHHVA" {
		t.Fatalf("Goal id = %s, want G-NXHHVA", g.ID)
	}
	if g.State != StateWritten || g.Actor != "a" || g.Budget != 3 || g.Words != "w" || len(g.Lines) != 2 {
		t.Fatalf("Goal = %+v", g)
	}
	if got := tb.ByID("G-NXHHVA"); got != g {
		t.Fatalf("ByID = %v", got)
	}
	if tb.ByID("G-NOPE00") != nil {
		t.Fatal("ByID found a Goal that is not there")
	}
}

func TestWholeLinesRefusesNoGap(t *testing.T) {
	if err := GuardWholeLines([]events.Line{gap, guard}); err != nil {
		t.Fatalf("a Goal with a gap line was refused: %v", err)
	}
	err := GuardWholeLines([]events.Line{{Key: "line-1", Text: "x", Kind: events.KindGuard}})
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Whole lines" {
		t.Fatalf("err = %v, want a Whole lines refusal", err)
	}
	if err := GuardWholeLines(nil); err == nil {
		t.Fatal("a Goal with no line was allowed")
	}
}

func TestWholeLinesRefusesMissingKind(t *testing.T) {
	err := GuardWholeLines([]events.Line{gap, {Key: "line-2", Text: "x"}})
	if err == nil || string(fault.GuardOf(err)) != "Whole lines" || !testutil.Mentions(err, "line-2") {
		t.Fatalf("err = %v, want a Whole lines refusal naming line-2", err)
	}
	err = GuardWholeLines([]events.Line{gap, {Key: "line-2", Text: "x", Kind: "maybe"}})
	if err == nil {
		t.Fatal("an unknown kind was allowed")
	}
}

func TestWholeLinesRefusesMoreThanTen(t *testing.T) {
	var lines []events.Line
	for i := 0; i < 11; i++ {
		lines = append(lines, events.Line{Key: "k", Text: "x", Kind: events.KindGap})
	}
	if err := GuardWholeLines(lines[:10]); err != nil {
		t.Fatalf("ten lines refused: %v", err)
	}
	if err := GuardWholeLines(lines); err == nil || string(fault.GuardOf(err)) != "Whole lines" {
		t.Fatalf("eleven lines: %v", err)
	}
}

func TestDraftReplacesWrittenGoal(t *testing.T) {
	first := goalAdd(t, id1, "", "old", gap)
	replace := goalAdd(t, id2, id1, "new", gap, guard)
	replace.Actor = "b"
	tb := Replay([]events.Event{first, replace})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	if len(tb.Goals) != 1 {
		t.Fatalf("a replace made %d Goals, want 1", len(tb.Goals))
	}
	g := tb.Goals[id1]
	if g.Words != "new" || len(g.Lines) != 2 || g.ID != "G-NXHHVA" || g.Actor != "a" {
		t.Fatalf("Goal after replace = %+v; want new words and lines, the first id and actor", g)
	}
}

func TestDraftRefusesUnknownGoal(t *testing.T) {
	if err := GuardDraft(Tables{}, ""); err != nil {
		t.Fatalf("a new Goal was refused: %v", err)
	}
	err := GuardDraft(Tables{Goals: map[string]*Goal{}}, id1)
	if err == nil || string(fault.GuardOf(err)) != "Draft" {
		t.Fatalf("err = %v, want a Draft refusal", err)
	}
	tb := Replay([]events.Event{goalAdd(t, id2, id1, "w", gap)})
	if len(tb.Goals) != 0 || len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != id2 {
		t.Fatalf("tables = %+v; want the event refused at replay", tb)
	}
}

func TestDraftRefusesFrozenGoal(t *testing.T) {
	tb := Replay([]events.Event{goalAdd(t, id1, "", "w", gap)})
	tb.Goals[id1].State = "open"
	err := GuardDraft(tb, id1)
	if err == nil || string(fault.GuardOf(err)) != "Draft" || fault.KindOf(err) != fault.Refusal {
		t.Fatalf("err = %v, want a Draft refusal", err)
	}
}

func TestReplayKeepsRefusedEventAndGoesOn(t *testing.T) {
	bad := goalAdd(t, id1, "", "w", guard)
	good := goalAdd(t, id2, "", "w2", gap)
	tb := Replay([]events.Event{bad, good})
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != id1 ||
		string(fault.GuardOf(tb.RefusedAtReplay[0].Err)) != "Whole lines" {
		t.Fatalf("RefusedAtReplay = %+v", tb.RefusedAtReplay)
	}
	if _, ok := tb.Goals[id2]; !ok || len(tb.Goals) != 1 {
		t.Fatalf("Goals = %v", tb.Goals)
	}
}

func TestFindResolvesShortAndRefusesAmbiguous(t *testing.T) {
	a := "01M661QG600000000000AAAAAA"
	b := "01M661QG700000000000BBAAAA"
	c := "01M661QG800000000000CCAAAA"
	tb := Replay([]events.Event{goalAdd(t, a, "", "w", gap), goalAdd(t, b, "", "w", gap), goalAdd(t, c, "", "w", gap)})
	g, err := tb.Find("G-AAAAAA")
	if err != nil || g.EventID != a {
		t.Fatalf("Find = %v, %v", g, err)
	}
	if g, err := tb.Find(b); err != nil || g.EventID != b {
		t.Fatalf("Find by event id = %v, %v", g, err)
	}
	if _, err := tb.Find("G-NOPE00"); fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("unknown Goal: %v", err)
	}
	// two Goals sharing their last six characters
	d := "01M661QG900000000000XXAAAA"
	tb = Replay([]events.Event{goalAdd(t, c, "", "w", gap), goalAdd(t, d, "", "w", gap)})
	if _, err := tb.Find("G-AAAAAA"); err == nil || fault.KindOf(err) != fault.BadArguments {
		t.Fatalf("ambiguous id: %v", err)
	}
}

const (
	idP1 = "01M661QG800000000000PLACE1"
	idP2 = "01M661QG900000000000PLACE2"
	idP3 = "01M661QGA00000000000PLACE3"
	idP4 = "01M661QGB00000000000PLACE4"
)

// placeEv is an example.place event on the Goal id1 whose H0 row tree is "t".
func placeEv(t *testing.T, id, line, example, hash, verdict string) events.Event {
	t.Helper()
	return placeOn(t, id, id1, "g", "grill", line, example, hash, "t", verdict)
}

func placeOn(t *testing.T, id, goal, actor, role, line, example, hash, rowTree, verdict string) events.Event {
	t.Helper()
	p, err := json.Marshal(events.ExamplePlace{
		Role: role, Line: line, Example: example, Hash: hash,
		H0: events.Cell{Commit: "c", RowTree: rowTree, Probe: "p", Output: "o", Stderr: "e", Exit: 1, Verdict: verdict},
	})
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: id, TS: "2026-11-19T00:00:00Z", Actor: actor, Type: events.TypeExamplePlace, Node: goal, Payload: p}
}

const hashA, hashB, hashC = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb", "cccccccccccccccccccccccccccccccccccccccc"

func TestPlacementVersionCounts(t *testing.T) {
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap, guard),
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		placeEv(t, idP2, "line-1", "a", hashB, "failed"),
		placeEv(t, idP3, "line-1", "b", hashC, "failed"),
	})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	hist := g.Placements["line-1/a"]
	if len(hist) != 2 || hist[0].Version != 1 || hist[1].Version != 2 || hist[0].Hash != hashA || hist[1].Hash != hashB {
		t.Fatalf("history of line-1/a = %+v, want @1 then @2", hist)
	}
	cur, ok := g.Current("line-1/a")
	if !ok || cur.Hash != hashB || cur.Version != 2 || cur.Actor != "g" || cur.Line != "line-1" || cur.Example != "a" || cur.EventID != idP2 {
		t.Fatalf("current = %+v, %v; want hashB at @2", cur, ok)
	}
	if b, ok := g.Current("line-1/b"); !ok || b.Version != 1 {
		t.Fatalf("line-1/b = %+v, %v; want @1: versions count per key", b, ok)
	}
	if _, ok := g.Current("line-1/none"); ok {
		t.Fatal("Current found an example key that was never placed")
	}
	if got := g.Placed(); len(got) != 2 || got[0].Key != "line-1/a" || got[1].Key != "line-1/b" {
		t.Fatalf("Placed = %+v, want the current placements in key order", got)
	}
}

func TestGoalOpensWhenEveryLineHasKindVerdict(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap, guard)
	// a gap line needs failed, a guard line needs met
	opens := Replay([]events.Event{goal,
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		placeEv(t, idP2, "line-2", "a", hashB, "met"),
	})
	if g := opens.Goals[id1]; g.State != StateOpen {
		t.Fatalf("state = %s, want open", g.State)
	}
	for name, evs := range map[string][]events.Event{
		"gap example met":      {placeEv(t, idP1, "line-1", "a", hashA, "met"), placeEv(t, idP2, "line-2", "a", hashB, "met")},
		"guard example failed": {placeEv(t, idP1, "line-1", "a", hashA, "failed"), placeEv(t, idP2, "line-2", "a", hashB, "failed")},
		"second gap example met": {placeEv(t, idP1, "line-1", "a", hashA, "failed"), placeEv(t, idP2, "line-1", "b", hashC, "met"),
			placeEv(t, idP3, "line-2", "a", hashB, "met")},
	} {
		tb := Replay(append([]events.Event{goal}, evs...))
		if g := tb.Goals[id1]; g.State != StateWritten {
			t.Errorf("%s: state = %s, want written", name, g.State)
		}
	}
}

func TestGoalStaysWrittenWithUnplacedLine(t *testing.T) {
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap, guard),
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
	})
	g := tb.Goals[id1]
	if len(g.Placed()) != 1 || g.State != StateWritten {
		t.Fatalf("state = %s with %d placed; line-2 has no example, so the Goal is written", g.State, len(g.Placed()))
	}
	tb = Replay([]events.Event{goalAdd(t, id1, "", "w", gap)})
	if g := tb.Goals[id1]; g.State != StateWritten {
		t.Fatalf("a Goal with no placement is %s, want written", g.State)
	}
}

func TestOpenGoalRefusesDraftReplace(t *testing.T) {
	tb := Replay([]events.Event{
		goalAdd(t, id1, "", "w", gap),
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
	})
	if err := GuardDraft(tb, id1); err == nil || string(fault.GuardOf(err)) != "Draft" {
		t.Fatalf("GuardDraft on an open Goal = %v, want a Draft refusal", err)
	}
}

func TestMergedLogSkipsRefusedEvent(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap)
	// Clone A replaces the Goal's draft; clone B places the example that
	// opens it, with an id lower than A's.
	replaceA := goalAdd(t, "01M661QGC00000000000REPLAC", id1, "replaced", gap, guard)
	replaceA.Actor = "a"
	placeB := placeEv(t, idP1, "line-1", "a", hashA, "failed")
	cloneA := []events.Event{goal, replaceA}
	cloneB := []events.Event{goal, placeB}
	seen := map[string]bool{}
	var merged []events.Event
	for _, ev := range append(cloneA, cloneB...) {
		if !seen[ev.ID] {
			seen[ev.ID] = true
			merged = append(merged, ev)
		}
	}
	sort.SliceStable(merged, func(i, j int) bool { return merged[i].ID < merged[j].ID })
	tb := Replay(merged)
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != replaceA.ID ||
		string(fault.GuardOf(tb.RefusedAtReplay[0].Err)) != "Draft" {
		t.Fatalf("RefusedAtReplay = %+v, want A's replace refused by Draft", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if g.State != StateOpen || g.Words != "w" || len(g.Lines) != 1 {
		t.Fatalf("Goal = state %s words %q lines %d; want B's placement applied and A's replace skipped", g.State, g.Words, len(g.Lines))
	}
	if _, ok := g.Current("line-1/a"); !ok {
		t.Fatal("B's placement is missing")
	}
}

func TestPlacementOfGoneLineIsHistory(t *testing.T) {
	two := goalAdd(t, id1, "", "w", gap, events.Line{Key: "line-2", Text: "second", Kind: events.KindGap})
	replace := goalAdd(t, idP2, id1, "w", gap) // line-2 is gone
	replace.Actor = "a"
	oneFailed := placeEv(t, idP1, "line-1", "a", hashA, "failed")
	goneMet := placeEv(t, "01M661QG800000000000PLACE0", "line-2", "a", hashB, "met")

	// after a replace, in the order the events happened
	tb := Replay([]events.Event{two, goneMet, replace, oneFailed})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if g.State != StateOpen {
		t.Fatalf("state = %s; a placement whose line is gone must not hold the Goal back", g.State)
	}
	if got := g.Placed(); len(got) != 1 || got[0].Line != "line-1" {
		t.Fatalf("Placed = %+v, want only line-1's", got)
	}
	if _, ok := g.Current("line-2/a"); !ok {
		t.Fatal("the placement on the gone line was deleted; it is history, kept")
	}

	// a merged log that puts the replace before a placement on the gone line
	late := placeEv(t, "01M661QGD00000000000LATE00", "line-2", "a", hashB, "met")
	tb = Replay([]events.Event{two, replace, oneFailed, late})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("a placement on a gone line was refused at replay: %+v", tb.RefusedAtReplay)
	}
	g = tb.Goals[id1]
	if g.State != StateOpen || len(g.Placed()) != 1 {
		t.Fatalf("state %s, placed %d; the late placement is history and counts for nothing", g.State, len(g.Placed()))
	}

	// the line comes back under the same key: its old placement counts again
	// only through the rule, here a met example on a gap line keeps it written
	back := goalAdd(t, "01M661QGE00000000000BACK00", id1, "w", gap, events.Line{Key: "line-2", Text: "second", Kind: events.KindGap})
	back.Actor = "a"
	if g := Replay([]events.Event{two, goneMet, replace, back, oneFailed}).Goals[id1]; g.State != StateWritten {
		t.Fatalf("state = %s, want written: line-2's current example is met on a gap line", g.State)
	}
}

func TestReplayRefusesPlacementByTables(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap, guard)
	for name, tc := range map[string]struct {
		ev   events.Event
		want fault.Code
	}{
		"runner role":  {placeOn(t, idP1, id1, "r", "runner", "line-1", "a", hashA, "t", "failed"), fault.NotGrill},
		"unknown goal": {placeOn(t, idP1, id2, "g", "grill", "line-1", "a", hashA, "t", "failed"), fault.NoGoalToPlace},
	} {
		tb := Replay([]events.Event{goal, tc.ev})
		if len(tb.RefusedAtReplay) != 1 || fault.CodeOf(tb.RefusedAtReplay[0].Err) != tc.want {
			t.Errorf("%s: RefusedAtReplay = %+v, want a %s refusal", name, tb.RefusedAtReplay, tc.want)
		}
		if len(tb.Goals[id1].Placements) != 0 {
			t.Errorf("%s: a refused placement was applied", name)
		}
	}
	// the table half of Adds only: the hash already current
	tb := Replay([]events.Event{goal,
		placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		placeEv(t, idP2, "line-1", "a", hashA, "failed"),
	})
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != idP2 || string(fault.GuardOf(tb.RefusedAtReplay[0].Err)) != "Adds only" {
		t.Fatalf("RefusedAtReplay = %+v, want the repeat refused by Adds only", tb.RefusedAtReplay)
	}
	if h := tb.Goals[id1].Placements["line-1/a"]; len(h) != 1 {
		t.Fatalf("history = %+v, want one placement", h)
	}
}

func TestGuardsOfPlace(t *testing.T) {
	g := &Goal{ID: "G-X", State: StateWritten}
	if err := GuardNotSelected(g); err != nil {
		t.Fatalf("a written Goal was refused: %v", err)
	}
	g.State = StateSelected
	err := GuardNotSelected(g)
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Selected" {
		t.Fatalf("a selected Goal: %v, want a Selected refusal", err)
	}
	if err := GuardGrillPlaces("grill"); err != nil {
		t.Fatalf("grill refused: %v", err)
	}
	for _, role := range []string{"runner", "owner", "probe-writer", ""} {
		err := GuardGrillPlaces(role)
		if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Grill places" {
			t.Errorf("role %q: %v, want an Grill places refusal", role, err)
		}
	}
	g = &Goal{Placements: map[string][]Placement{"line-1/a": {{Key: "line-1/a", Hash: hashA, Version: 1}}}}
	if err := GuardAddsOnlyTable(g, "line-1/a", hashB); err != nil {
		t.Fatalf("a new hash was refused: %v", err)
	}
	if err := GuardAddsOnlyTable(g, "line-1/new", hashA); err != nil {
		t.Fatalf("a new key was refused: %v", err)
	}
	err = GuardAddsOnlyTable(g, "line-1/a", hashA)
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Adds only" {
		t.Fatalf("the current hash: %v, want an Adds only refusal", err)
	}
}

func TestH0CellIsNotRunAfterTheBaseMoves(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap, guard)
	replace := goalAdd(t, idP2, id1, "w", gap, guard)
	replace.Actor = "a"
	var p events.GoalAdd
	if err := events.Decode(replace, &p); err != nil {
		t.Fatal(err)
	}
	p.BaseTree = "other"
	replace.Payload, _ = json.Marshal(p)

	tb := Replay([]events.Event{goal, placeEv(t, idP1, "line-1", "a", hashA, "failed")})
	if v := tb.Goals[id1].H0Verdict("line-1/a"); v != "failed" {
		t.Fatalf("H0 verdict = %q, want failed", v)
	}
	tb = Replay([]events.Event{goal, placeEv(t, idP1, "line-1", "a", hashA, "failed"), replace})
	if len(tb.RefusedAtReplay) != 0 {
		t.Fatalf("refused at replay: %+v", tb.RefusedAtReplay)
	}
	g := tb.Goals[id1]
	if v := g.H0Verdict("line-1/a"); v != NotRun {
		t.Fatalf("H0 verdict after the base moved = %q, want %q", v, NotRun)
	}
	if v := g.H0Verdict("line-1/none"); v != NotRun {
		t.Fatalf("H0 verdict of an unplaced key = %q, want %q", v, NotRun)
	}
	if g.State != StateWritten {
		t.Fatalf("state = %s, want written: the cell is not live", g.State)
	}
}

func TestAddsOnlyTableRefusesAnEarlierHash(t *testing.T) {
	g := &Goal{Placements: map[string][]Placement{"line-1/a": {
		{Key: "line-1/a", Hash: hashA, Version: 1},
		{Key: "line-1/a", Hash: hashB, Version: 2},
	}}}
	err := GuardAddsOnlyTable(g, "line-1/a", hashA)
	if err == nil || fault.KindOf(err) != fault.Refusal || string(fault.GuardOf(err)) != "Adds only" {
		t.Fatalf("placing @1 again after @2: %v, want an Adds only refusal", err)
	}
}

// placeSetEv is an examples.place event on the Goal id1 by the grill, each
// example given as line, example, hash and H0's verdict.
func placeSetEv(t *testing.T, id string, xs ...[4]string) events.Event {
	t.Helper()
	p := events.ExamplesPlace{Role: "grill"}
	for _, x := range xs {
		p.Examples = append(p.Examples, events.PlacedExample{Line: x[0], Example: x[1], Hash: x[2],
			H0: &events.Cell{Commit: "c", RowTree: "t", Probe: "p", Output: "o", Stderr: "e", Exit: 0, Verdict: x[3]}})
	}
	b, err := json.Marshal(p)
	if err != nil {
		t.Fatal(err)
	}
	return events.Event{V: 2, ID: id, TS: "2026-11-22T00:00:00Z", Actor: "g", Type: events.TypeExamplesPlace, Node: id1, Payload: b}
}

// TestPlaceSetPlacesEveryExampleAtOnce: one examples.place event places each
// example it names, opens the Goal, and starts one budget for the set.
func TestPlaceSetPlacesEveryExampleAtOnce(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap, guard)
	tb := Replay([]events.Event{goal, placeSetEv(t, idP1,
		[4]string{"line-1", "a", hashA, "failed"}, [4]string{"line-1", "b", hashB, "failed"}, [4]string{"line-2", "a", hashC, "met"})})
	g := tb.Goals[id1]
	if len(tb.RefusedAtReplay) != 0 || len(g.Placed()) != 3 || g.State != StateOpen || g.epoch != 1 {
		t.Fatalf("refused %+v, placed %+v, state %s, epoch %d; want three placed, open, one budget",
			tb.RefusedAtReplay, g.Placed(), g.State, g.epoch)
	}
	// a later set holds only what changed: a new version of line-1/a
	tb = Replay([]events.Event{goal,
		placeSetEv(t, idP1, [4]string{"line-1", "a", hashA, "failed"}, [4]string{"line-2", "a", hashC, "met"}),
		placeSetEv(t, idP2, [4]string{"line-1", "a", hashB, "failed"})})
	g = tb.Goals[id1]
	if cur, _ := g.Current("line-1/a"); cur.Version != 2 || cur.Hash != hashB || g.epoch != 2 {
		t.Fatalf("line-1/a = %+v, epoch %d; want @2 and a second budget", cur, g.epoch)
	}
	// an old-form example.place and a set replay in one log
	tb = Replay([]events.Event{goal, placeEv(t, idP1, "line-1", "a", hashA, "failed"),
		placeSetEv(t, idP2, [4]string{"line-2", "a", hashC, "met"})})
	if g := tb.Goals[id1]; len(tb.RefusedAtReplay) != 0 || g.State != StateOpen {
		t.Fatalf("mixed log: refused %+v, state %s", tb.RefusedAtReplay, g.State)
	}
}

// TestReplayRefusesABadSet: a set with no example, a key twice, or a hash the
// key held is refused whole at replay, and places nothing.
func TestReplayRefusesABadSet(t *testing.T) {
	goal := goalAdd(t, id1, "", "w", gap)
	first := placeSetEv(t, idP1, [4]string{"line-1", "a", hashA, "failed"})
	for name, ev := range map[string]events.Event{
		"empty":     placeSetEv(t, idP2),
		"key twice": placeSetEv(t, idP2, [4]string{"line-1", "b", hashB, "failed"}, [4]string{"line-1", "b", hashC, "failed"}),
		"held hash": placeSetEv(t, idP2, [4]string{"line-1", "b", hashB, "failed"}, [4]string{"line-1", "a", hashA, "failed"}),
	} {
		tb := Replay([]events.Event{goal, first, ev})
		g := tb.Goals[id1]
		if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != idP2 || len(g.Placed()) != 1 || g.epoch != 1 {
			t.Errorf("%s: refused %+v, placed %+v, epoch %d; want the set refused whole", name, tb.RefusedAtReplay, g.Placed(), g.epoch)
		}
	}
}
