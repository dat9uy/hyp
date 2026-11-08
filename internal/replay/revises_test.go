package replay

import (
	"strings"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// A merged log keeps the first close of a reading and rejects a second one,
// rejects a revision that sorts before the close, and keeps every revision
// after it in merge order, the last in force.
func TestRevisionsInMergeOrder(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	e := b.Exp(h, "none")
	r := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: e,
		Verdicts: map[string]string{h: event.VerdictSupports}, Observation: "o",
		Readings: []event.ReadingItem{{Text: "is a reader a logic?"}}})
	k := event.ReadingKey(r, 1)
	owner := func(c event.Checkpoint) string {
		c.Role = event.RoleOwner
		return b.As("owner").Emit(event.TypeCheckpoint, g, c)
	}
	early := owner(event.Checkpoint{Summary: "too early", Revises: []string{k}})
	first := owner(event.Checkpoint{Summary: "yes", Closes: []string{k}})
	second := owner(event.Checkpoint{Summary: "a second close", Closes: []string{k}})
	owner(event.Checkpoint{Summary: "of the file", Revises: []string{k}})
	last := owner(event.Checkpoint{Summary: "of the file, not a Section", Revises: []string{k}})

	p := Replay(b.Evs, nil, "")
	rejected := map[string]string{}
	for _, x := range p.Rejections {
		rejected[x.EventID] = x.Reason
	}
	if !strings.Contains(rejected[early], "is open; an owner answers an open reading with closes") {
		t.Fatalf("revision before the close not rejected: %+v", p.Rejections)
	}
	if !strings.Contains(rejected[second], "is answered; an owner revises an answer with revises") {
		t.Fatalf("second close not rejected: %+v", p.Rejections)
	}
	if len(p.Rejections) != 2 {
		t.Fatalf("want 2 rejections, got %+v", p.Rejections)
	}
	l := p.Readings[0]
	if l.By != first || l.Answer != "of the file, not a Section" || len(l.Answers) != 3 || l.Answers[2].By != last {
		t.Fatalf("reading %+v", l)
	}
}
