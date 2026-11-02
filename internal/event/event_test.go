package event

import "testing"

func TestParseLineRejectsUnknownField(t *testing.T) {
	line := []byte(`{"v":1,"id":"01JB0000000000000000ABCDEF","ts":"2026-10-24T00:00:00Z","actor":"a","type":"node.add","node":"x","payload":{},"why":"w","extra":1}`)
	if _, err := ParseLine(line); err == nil {
		t.Fatal("expected error for unknown top-level field")
	}
	ok := []byte(`{"v":1,"id":"01JB0000000000000000ABCDEF","ts":"2026-10-24T00:00:00Z","actor":"a","type":"node.add","node":"x","payload":{},"why":"w"}`)
	if _, err := ParseLine(ok); err != nil {
		t.Fatalf("valid line rejected: %v", err)
	}
}

func TestDecodePayloadRejectsUnknownField(t *testing.T) {
	e := Event{Type: TypeNodeAdd, Payload: []byte(`{"kind":"Hypothesis","parent":"p","title":"t","status":"supported"}`)}
	var p NodeAdd
	if err := DecodePayload(e, &p); err == nil {
		t.Fatal("expected error for status field in payload")
	}
}

func TestDisplayID(t *testing.T) {
	id := "01JB0000000000000000ABCDEF"
	cases := map[string]string{
		KindGoal:           "G-ABCDEF",
		KindDesignQuestion: "Q-ABCDEF",
		KindHypothesis:     "H-ABCDEF",
		KindExperiment:     "E-ABCDEF",
		KindResult:         "R-ABCDEF",
	}
	for kind, want := range cases {
		if got := DisplayID(kind, id); got != want {
			t.Errorf("DisplayID(%s) = %s, want %s", kind, got, want)
		}
	}
}

func TestRunesCountsRunes(t *testing.T) {
	if got := Runes("ếếế"); got != 3 {
		t.Fatalf("Runes = %d, want 3", got)
	}
}

func TestClosedSets(t *testing.T) {
	cases := []struct {
		name string
		set  map[string]bool
		want int
	}{
		{"EventTypes", EventTypes, 17},
		{"Kinds", Kinds, 5},
		{"EdgeTypes", EdgeTypes, 3},
		{"Verdicts", Verdicts, 3},
	}
	for _, c := range cases {
		if len(c.set) != c.want {
			t.Errorf("len(%s) = %d, want %d", c.name, len(c.set), c.want)
		}
	}
}
