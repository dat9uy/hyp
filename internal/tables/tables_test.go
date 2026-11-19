package tables

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/events"
	"github.com/dat9uy/hyp/internal/herr"
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
	if err == nil || herr.KindOf(err) != herr.Refusal || !strings.HasPrefix(herr.Text(err), "Whole lines") {
		t.Fatalf("err = %v, want a Whole lines refusal", err)
	}
	if err := GuardWholeLines(nil); err == nil {
		t.Fatal("a Goal with no line was allowed")
	}
}

func TestWholeLinesRefusesMissingKind(t *testing.T) {
	err := GuardWholeLines([]events.Line{gap, {Key: "line-2", Text: "x"}})
	if err == nil || !strings.HasPrefix(herr.Text(err), "Whole lines") || !strings.Contains(err.Error(), "line-2") {
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
	if err := GuardWholeLines(lines); err == nil || !strings.HasPrefix(herr.Text(err), "Whole lines") {
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
	if err == nil || !strings.HasPrefix(herr.Text(err), "Draft") {
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
	if err == nil || !strings.HasPrefix(herr.Text(err), "Draft") || herr.KindOf(err) != herr.Refusal {
		t.Fatalf("err = %v, want a Draft refusal", err)
	}
}

func TestReplayKeepsRefusedEventAndGoesOn(t *testing.T) {
	bad := goalAdd(t, id1, "", "w", guard)
	good := goalAdd(t, id2, "", "w2", gap)
	tb := Replay([]events.Event{bad, good})
	if len(tb.RefusedAtReplay) != 1 || tb.RefusedAtReplay[0].EventID != id1 ||
		!strings.HasPrefix(tb.RefusedAtReplay[0].Reason, "Whole lines") {
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
	if _, err := tb.Find("G-NOPE00"); herr.KindOf(err) != herr.BadArguments {
		t.Fatalf("unknown Goal: %v", err)
	}
	// two Goals sharing their last six characters
	d := "01M661QG900000000000XXAAAA"
	tb = Replay([]events.Event{goalAdd(t, c, "", "w", gap), goalAdd(t, d, "", "w", gap)})
	if _, err := tb.Find("G-AAAAAA"); err == nil || herr.KindOf(err) != herr.BadArguments {
		t.Fatalf("ambiguous id: %v", err)
	}
}
