package events

import (
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func TestParseLineRefusesV1(t *testing.T) {
	const v1 = `{"v":1,"id":"01M661QG600000000000000001","ts":"2026-10-30T09:00:01Z","actor":"a","type":"node.add","node":"01M661QG600000000000000001","payload":{"kind":"Goal","parent":"","title":"G","alias":"g"},"why":"start"}`
	if _, err := ParseLine([]byte(v1)); err == nil {
		t.Fatal("a v1 line parsed")
	}
}

func TestParseLineRefusesUnknownType(t *testing.T) {
	line := `{"v":2,"id":"01M661QG600000000000000001","ts":"2026-10-30T09:00:01Z","actor":"a","type":"node.add","node":"x","payload":{},"why":""}`
	if _, err := ParseLine([]byte(line)); err == nil {
		t.Fatal("an unknown type parsed")
	}
}

func TestParseLineRefusesUnknownEnvelopeField(t *testing.T) {
	line := `{"v":2,"id":"01M661QG600000000000000001","ts":"t","actor":"a","type":"budget.reset","node":"x","payload":{},"why":"","extra":1}`
	if _, err := ParseLine([]byte(line)); err == nil {
		t.Fatal("an unknown envelope field parsed")
	}
}

func TestGoalAddRoundTrip(t *testing.T) {
	want := GoalAdd{
		Role:     "owner",
		Words:    "Find the missing case.",
		Base:     strings.Repeat("a", 40),
		BaseTree: strings.Repeat("b", 40),
		Budget:   3,
		From:     "G-ABCDEF",
		Lines: []Line{
			{Key: "line-1", Text: "it fails", Kind: KindGap},
			{Key: "line-2", Text: "it holds", Kind: KindGuard},
		},
	}
	payload, err := json.Marshal(want)
	if err != nil {
		t.Fatal(err)
	}
	ev := Event{V: Version, ID: "01M661QG600000000000000001", TS: "2026-10-30T09:00:01Z", Actor: "a",
		Type: TypeGoalAdd, Node: "01M661QG600000000000000001", Payload: payload, Why: "start"}
	line, err := json.Marshal(ev)
	if err != nil {
		t.Fatal(err)
	}
	got, err := ParseLine(line)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(got, ev) {
		t.Fatalf("envelope: got %+v, want %+v", got, ev)
	}
	var back GoalAdd
	if err := Decode(got, &back); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(back, want) {
		t.Fatalf("payload: got %+v, want %+v", back, want)
	}
}

func TestDecodeRefusesUnknownField(t *testing.T) {
	ev := Event{Type: TypeBudgetReset, Payload: json.RawMessage(`{"role":"owner","surprise":1}`)}
	var p BudgetReset
	if err := Decode(ev, &p); err == nil {
		t.Fatal("an unknown payload field decoded")
	}
	ok := Event{Type: TypeBudgetReset, Payload: json.RawMessage(`{"role":"owner"}`)}
	if err := Decode(ok, &p); err != nil || p.Role != "owner" {
		t.Fatalf("Decode = %v, %+v", err, p)
	}
}

func TestEveryTypeIsKnown(t *testing.T) {
	for _, typ := range []string{TypeGoalAdd, TypeExamplePlace, TypeResultRecord, TypeResultRetract, TypeSelect, TypeBudgetReset} {
		line := `{"v":2,"id":"i","ts":"t","actor":"a","type":"` + typ + `","node":"n","payload":{},"why":""}`
		if _, err := ParseLine([]byte(line)); err != nil {
			t.Errorf("%s: %v", typ, err)
		}
	}
}
