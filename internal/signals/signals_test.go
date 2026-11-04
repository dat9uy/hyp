package signals

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/replay"
	"github.com/dat9uy/hyp/internal/store"
)

// synth reads the synthetic store's events (internal/testutil/testdata/synth,
// written by its generate.sh with the current hyp), cut after the event with
// id cut, or whole when cut is "".
func synth(t *testing.T, cut string) []event.Event {
	t.Helper()
	evs, _, err := store.Events("../testutil/testdata/synth")
	if err != nil {
		t.Fatal(err)
	}
	if cut == "" {
		return evs
	}
	for i, ev := range evs {
		if ev.ID == cut {
			return evs[:i+1]
		}
	}
	t.Fatalf("no event %s", cut)
	return nil
}

func wantGolden(t *testing.T, name, got string) {
	t.Helper()
	want, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("%s differs:\n--- got\n%s--- want\n%s", name, got, want)
	}
}

func signalLines(r Report) string {
	var b strings.Builder
	for _, s := range r.Signals {
		fmt.Fprintf(&b, "%s\t%d\t%s\n", s.Key, s.Round, s.Detail)
	}
	return b.String()
}

// The golden files hold every signal generate.sh's comment names for the
// synthetic store: one-rival, unjudged, wasted, no-movement and spin over the
// whole log, and verifier-loop at the cut after the second insufficient verdict.
func TestSynthMatchesGolden(t *testing.T) {
	for _, c := range []struct{ name, cut string }{
		{"full.golden", ""},
		{"cut-loop.golden", "01M6A4QYZCM5RGD6AZ58XJTSBW"},
	} {
		wantGolden(t, c.name, signalLines(Compute(synth(t, c.cut))))
	}
	var b strings.Builder
	for _, r := range Compute(synth(t, "")).Rounds {
		b.WriteString(r.Line + "\n")
	}
	wantGolden(t, "rounds-full.golden", b.String())
}

// A refused event changes no round and no signal: Compute steps the whole log
// and skips what replay refuses.
func TestRefusedEventChangesNothing(t *testing.T) {
	evs := synth(t, "")
	dup := evs[0]
	dup.ID = "01M5ZZZZZZZZZZZZZZZZZZZZZZ"
	mid := len(evs) / 2
	withDup := append(append(append([]event.Event{}, evs[:mid]...), dup), evs[mid:]...)
	if _, p := replay.Build(withDup, nil, ""); len(p.Rejections) != 1 {
		t.Fatalf("want the duplicate refused, got rejections %v", p.Rejections)
	}
	want, _ := json.Marshal(Compute(evs))
	got, _ := json.Marshal(Compute(withDup))
	if string(got) != string(want) {
		t.Errorf("a refused event changed the report:\n--- got\n%s\n--- want\n%s", got, want)
	}
}
