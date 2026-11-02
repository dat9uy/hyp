package replay

import (
	"reflect"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// Results holds one record per Result, built once: status, questionable
// flag and pins; Uncontained is left for contain.Mark.
func TestResultsRecord(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h1 := b.Hyp(q, "h1")
	flagged := b.Result(b.Exp(h1, "boot@1"), map[string]string{h1: sup})
	clean := b.Result(b.Exp(h1, "tool@1"), map[string]string{h1: sup})
	plain := b.Result(b.Exp(h1, event.RecipeNone), map[string]string{h1: sup})
	gone := b.Result(b.Exp(h1, event.RecipeNone), map[string]string{h1: sup})
	b.Retract(gone)
	p := Replay(b.Evs, map[string]string{"boot@1": "changed", "tool@1": "aaa"}, "")
	wantNoRejections(t, p)

	if len(p.Results) != 4 {
		t.Fatalf("Results has %d records, want 4", len(p.Results))
	}
	for id, r := range p.Results {
		if r.Status != p.Status[id].Status {
			t.Errorf("%s: Res.Status %s, Status %s", id, r.Status, p.Status[id].Status)
		}
		if r.Uncontained != nil {
			t.Errorf("%s: replay set Uncontained %q", id, r.Uncontained)
		}
	}
	if r := p.Results[flagged]; !r.Questionable || !reflect.DeepEqual(r.Pins, []Pin{{Ref: event.Ref{Name: "boot", Version: "1"}, Hash: "aaa"}}) {
		t.Errorf("flagged = %+v", r)
	}
	if r := p.Results[clean]; r.Questionable || len(r.Pins) != 1 {
		t.Errorf("clean = %+v", r)
	}
	if r := p.Results[plain]; r.Questionable || r.Pins != nil || r.Status != ResRecorded {
		t.Errorf("plain = %+v", r)
	}
	if r := p.Results[gone]; r.Status != ResRetracted {
		t.Errorf("retracted = %+v", r)
	}
}
