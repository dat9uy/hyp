package event

import (
	"encoding/json"
	"testing"
)

// A reading with no Of is stored as a bare string, so a Result written before
// readings could read an answer decodes and encodes byte for byte the same.
func TestReadingItemKeepsTheStringForm(t *testing.T) {
	old := `{"parent":"E","verdicts":{},"observation":"o","readings":["a Goal line"]}`
	var p ResultRecord
	if err := json.Unmarshal([]byte(old), &p); err != nil {
		t.Fatal(err)
	}
	if len(p.Readings) != 1 || p.Readings[0] != (ReadingItem{Text: "a Goal line"}) {
		t.Fatalf("decoded %+v", p.Readings)
	}
	b, _ := json.Marshal(p)
	if string(b) != old {
		t.Fatalf("encoded %s, want %s", b, old)
	}

	p.Readings = append(p.Readings, ReadingItem{Of: "R#r1", Text: "an answer"})
	b, _ = json.Marshal(p)
	var back ResultRecord
	if err := json.Unmarshal(b, &back); err != nil || back.Readings[1] != p.Readings[1] {
		t.Fatalf("round trip %s: %+v, %v", b, back.Readings, err)
	}
	if err := json.Unmarshal([]byte(`{"parent":"E","verdicts":{},"observation":"o","readings":[{"text":"t","extra":1}]}`), &back); err == nil {
		t.Fatal("unknown field in a reading decoded")
	}
}

// A limit key and a reading key never parse as each other.
func TestParseKeyKeepsTheKindsApart(t *testing.T) {
	for s, want := range map[string]Key{
		"R-ABC#1":  {"R-ABC", KeyLimit, 1},
		"R-ABC#r2": {"R-ABC", KeyReading, 2},
	} {
		if got, ok := ParseKey(s); !ok || got != want || got.String() != s {
			t.Errorf("ParseKey(%q) = %+v, %v; want %+v", s, got, ok, want)
		}
	}
	for _, s := range []string{"R-ABC", "R-ABC#", "#1", "R-ABC#0", "R-ABC#r0", "R-ABC#rx", "R-ABC#x1"} {
		if k, ok := ParseKey(s); ok {
			t.Errorf("ParseKey(%q) = %+v, want no key", s, k)
		}
	}
}
