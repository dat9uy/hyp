package example

import (
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/herr"
)

func mustParse(t *testing.T, s string) Example {
	t.Helper()
	ex, err := Parse([]byte(s))
	if err != nil {
		t.Fatalf("Parse(%q): %v", s, err)
	}
	return ex
}

const base = "-- words --\nw\n-- run --\necho hi\n-- floor: contains --\nhi\n"

func TestParseLineOneExample(t *testing.T) {
	data, err := os.ReadFile("testdata/line1.txtar")
	if err != nil {
		t.Fatal(err)
	}
	ex, err := Parse(data)
	if err != nil {
		t.Fatal(err)
	}
	if ex.Words != "No Go file outside internal/render builds text for the agent." {
		t.Fatalf("words = %q", ex.Words)
	}
	if len(ex.Inputs) != 1 || ex.Inputs[0].Path != "patterns.txt" || ex.Inputs[0].Body != "herr.Refused\nrefusedErr\nusageErr\nFindRefuse" {
		t.Fatalf("inputs = %+v", ex.Inputs)
	}
	if !strings.HasPrefix(ex.Run, `grep -rlE -f "$HYP_EXAMPLE/input/patterns.txt"`) || strings.Contains(ex.Run, "\n") {
		t.Fatalf("run = %q", ex.Run)
	}
	if len(ex.Floors) != 1 || ex.Floors[0] != (Floor{Op: OpEquals, Text: ""}) {
		t.Fatalf("floors = %+v, want one empty equals floor", ex.Floors)
	}
	if len(ex.Ceilings) != 0 || ex.HasExternal {
		t.Fatalf("ceilings %v external %v", ex.Ceilings, ex.HasExternal)
	}
	if got := Compare(nil, ex); got != Met {
		t.Fatalf("the empty output = %s, want met", got)
	}
	if got := Compare([]byte("./x.go"), ex); got != Failed {
		t.Fatalf("a file name = %s, want failed", got)
	}
}

func TestParseDropsOneFinalNewline(t *testing.T) {
	ex := mustParse(t, "-- words --\n\nw\n\n-- run --\nr\n\n-- floor: equals --\nabc")
	if ex.Words != "\nw\n" || ex.Run != "r\n" {
		t.Fatalf("words %q run %q; want one final newline dropped each", ex.Words, ex.Run)
	}
	if ex.Floors[0].Text != "abc" {
		t.Fatalf("a body with no newline = %q", ex.Floors[0].Text)
	}
	ex = mustParse(t, "-- words --\nw\n-- run --\nr\n-- floor: equals --\n\n")
	if ex.Floors[0].Text != "" {
		t.Fatalf("a body of one blank line = %q, want empty", ex.Floors[0].Text)
	}
}

func TestParseKeepsEveryOtherByte(t *testing.T) {
	ex := mustParse(t, "-- words --\nw\n-- run --\nr\n-- floor: contains --\n  a\tb \r\n-- ceiling --\n--not a marker\n-- external --\nFOO\nBAR\n")
	if ex.Floors[0].Text != "  a\tb \r" {
		t.Fatalf("floor = %q", ex.Floors[0].Text)
	}
	if !slices.Equal(ex.Ceilings, []string{"--not a marker"}) {
		t.Fatalf("ceilings = %q", ex.Ceilings)
	}
	if !ex.HasExternal || !slices.Equal(ex.ExternalNames(), []string{"FOO", "BAR"}) {
		t.Fatalf("external = %q names %v", ex.External, ex.ExternalNames())
	}
}

func refusal(t *testing.T, s, want string) {
	t.Helper()
	_, err := Parse([]byte(s))
	if err == nil || herr.KindOf(err) != herr.Refusal || !strings.Contains(err.Error(), want) {
		t.Fatalf("Parse(%q) error = %v, want a refusal containing %q", s, err, want)
	}
}

func TestParseRefusesPreamble(t *testing.T) {
	refusal(t, "text first\n"+base, "line 1")
	refusal(t, "\n"+base, "line 1")
	refusal(t, "", "words")
}

func TestParseRefusesUnknownSection(t *testing.T) {
	refusal(t, base+"-- bogus --\nx\n", "line 7")
	refusal(t, base+"-- floor: matches --\nx\n", "line 7")
	refusal(t, base+"-- input/ --\nx\n", "line 7")
}

func TestParseRefusesDotDotInput(t *testing.T) {
	for _, name := range []string{"../x", "a/../../x", "/abs", "a//b", "./x", "a/", "a/./b"} {
		refusal(t, base+"-- input/"+name+" --\nx\n", "input")
	}
	ex := mustParse(t, base+"-- input/a/b.txt --\nx\n-- input/c --\ny\n")
	if len(ex.Inputs) != 2 || ex.Inputs[0].Path != "a/b.txt" {
		t.Fatalf("inputs = %+v", ex.Inputs)
	}
	refusal(t, base+"-- input/a --\nx\n-- input/a --\ny\n", "input/a")
}

func TestParseNeedsAFloor(t *testing.T) {
	refusal(t, "-- words --\nw\n-- run --\nr\n", "floor")
	refusal(t, "-- words --\nw\n-- run --\nr\n-- ceiling --\nx\n", "floor")
}

func TestParseCountsSections(t *testing.T) {
	refusal(t, base+"-- words --\nw2\n", "words")
	refusal(t, base+"-- run --\nr2\n", "run")
	refusal(t, "-- run --\nr\n-- floor: contains --\nx\n", "words")
	refusal(t, "-- words --\nw\n-- floor: contains --\nx\n", "run")
	refusal(t, base+"-- external --\nA\n-- external --\nB\n", "external")
	refusal(t, base+"-- external --\nA B\n", "external")
	refusal(t, base+"-- external --\n1A\n", "external")
	if ex := mustParse(t, base+"-- floor: equals --\nx\n-- floor: contains --\ny\n-- ceiling --\nz\n-- ceiling --\nz\n"); len(ex.Floors) != 3 || len(ex.Ceilings) != 2 {
		t.Fatalf("floors %+v ceilings %q", ex.Floors, ex.Ceilings)
	}
}

func TestCompareContainsEqualsCeiling(t *testing.T) {
	ex := func(floors []Floor, ceilings ...string) Example { return Example{Floors: floors, Ceilings: ceilings} }
	tests := []struct {
		name   string
		out    string
		ex     Example
		expect Verdict
	}{
		{"contains met", "say hi there", ex([]Floor{{OpContains, "hi"}}), Met},
		{"contains failed", "say ho there", ex([]Floor{{OpContains, "hi"}}), Failed},
		{"contains is case sensitive", "HI", ex([]Floor{{OpContains, "hi"}}), Failed},
		{"equals met", "abc", ex([]Floor{{OpEquals, "abc"}}), Met},
		{"equals failed on extra", "abc ", ex([]Floor{{OpEquals, "abc"}}), Failed},
		{"empty equals empty", "", ex([]Floor{{OpEquals, ""}}), Met},
		{"empty equals not text", "x", ex([]Floor{{OpEquals, ""}}), Failed},
		{"every floor must hold", "a", ex([]Floor{{OpContains, "a"}, {OpContains, "b"}}), Failed},
		{"all floors hold", "ab", ex([]Floor{{OpContains, "a"}, {OpContains, "b"}}), Met},
		{"ceiling present fails", "ab", ex([]Floor{{OpContains, "a"}}, "b"), Failed},
		{"ceiling absent holds", "a", ex([]Floor{{OpContains, "a"}}, "b"), Met},
		{"a ceiling beats an equals floor", "a", ex([]Floor{{OpEquals, "a"}}, "a"), Failed},
		{"empty ceiling is always contained", "a", ex([]Floor{{OpContains, "a"}}, ""), Failed},
	}
	for _, tc := range tests {
		if got := Compare([]byte(tc.out), tc.ex); got != tc.expect {
			t.Errorf("%s: Compare = %q, want %q", tc.name, got, tc.expect)
		}
	}
}

const (
	probeA = "1111111111111111111111111111111111111111"
	probeB = "2222222222222222222222222222222222222222"
)

func addsOnly(t *testing.T, old, new string) error {
	t.Helper()
	return AddsOnly(mustParse(t, old), mustParse(t, new), probeA, probeA)
}

func TestAddsOnlyAllowsNewFloor(t *testing.T) {
	if err := addsOnly(t, base, base+"-- floor: contains --\nmore\n"); err != nil {
		t.Fatalf("a new floor was refused: %v", err)
	}
}

func TestAddsOnlyAllowsNewCeiling(t *testing.T) {
	if err := addsOnly(t, base, base+"-- ceiling --\nbad\n"); err != nil {
		t.Fatalf("a new ceiling was refused: %v", err)
	}
	// the sections may sit in another order
	if err := addsOnly(t, base+"-- ceiling --\nx\n", "-- ceiling --\nbad\n"+base+"-- ceiling --\nx\n"); err != nil {
		t.Fatalf("a reordered file with a new ceiling was refused: %v", err)
	}
}

func adds(t *testing.T, old, new, probeOld, probeNew, want string) {
	t.Helper()
	err := AddsOnly(mustParse(t, old), mustParse(t, new), probeOld, probeNew)
	if err == nil || herr.KindOf(err) != herr.Refusal || !strings.HasPrefix(herr.Text(err), "Adds only") {
		t.Fatalf("err = %v, want an Adds only refusal", err)
	}
	if !strings.Contains(err.Error(), want) {
		t.Fatalf("err = %v, want it to name %q", err, want)
	}
}

func TestAddsOnlyRefusesDroppedCeiling(t *testing.T) {
	old := base + "-- ceiling --\nx\n"
	adds(t, old, base+"-- floor: contains --\nmore\n", probeA, probeA, "ceiling")
	// a dropped floor, with something else added
	adds(t, base+"-- floor: contains --\nmore\n", "-- words --\nw\n-- run --\necho hi\n-- floor: contains --\nmore\n-- ceiling --\nc\n", probeA, probeA, "floor")
}

func TestAddsOnlyCountsAsMultiset(t *testing.T) {
	old := base + "-- floor: contains --\nhi\n"
	adds(t, old, base+"-- floor: equals --\nz\n", probeA, probeA, "floor")
	if err := addsOnly(t, old, old+"-- floor: contains --\nhi\n"); err != nil {
		t.Fatalf("a third copy of a floor adds one: %v", err)
	}
	// the same text under another op is another floor
	adds(t, base, "-- words --\nw\n-- run --\necho hi\n-- floor: equals --\nhi\n", probeA, probeA, "floor")
}

func TestAddsOnlyRefusesChangedRun(t *testing.T) {
	adds(t, base, strings.Replace(base, "echo hi", "echo ho", 1)+"-- ceiling --\nx\n", probeA, probeA, "run")
}

func TestAddsOnlyRefusesChangedInput(t *testing.T) {
	old := base + "-- input/p.txt --\nA\n"
	adds(t, old, base+"-- input/p.txt --\nB\n-- ceiling --\nx\n", probeA, probeA, "input/p.txt")
	adds(t, old, old+"-- input/q.txt --\nB\n-- ceiling --\nx\n", probeA, probeA, "input/q.txt")
	adds(t, old, base+"-- ceiling --\nx\n", probeA, probeA, "input/p.txt")
}

func TestAddsOnlyRefusesChangedWords(t *testing.T) {
	adds(t, base, strings.Replace(base, "\nw\n", "\nother\n", 1)+"-- ceiling --\nx\n", probeA, probeA, "words")
}

func TestAddsOnlyRefusesChangedExternal(t *testing.T) {
	adds(t, base, base+"-- external --\nFOO\n-- ceiling --\nx\n", probeA, probeA, "external")
	adds(t, base+"-- external --\nFOO\n", base+"-- external --\nBAR\n-- ceiling --\nx\n", probeA, probeA, "external")
}

func TestAddsOnlyRefusesChangedProbe(t *testing.T) {
	adds(t, base, base+"-- ceiling --\nx\n", probeA, probeB, "probe/")
}

func TestAddsOnlyRefusesNothingAdded(t *testing.T) {
	adds(t, base, base, probeA, probeA, "nothing")
	adds(t, base, "-- run --\necho hi\n-- words --\nw\n-- floor: contains --\nhi\n", probeA, probeA, "nothing")
}
