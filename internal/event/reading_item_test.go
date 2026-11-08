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
