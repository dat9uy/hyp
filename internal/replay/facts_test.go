package replay

import (
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// An open reading is the owner's answer item in the ledger, on its Result's
// question and whatever that question's state, so AWAITING OWNER reads it
// from the ledger as it reads every other owner item. An answer settles it,
// and the runner then owes its apply.
func TestOpenReadingIsTheOwnersAnswerItem(t *testing.T) {
	b := testutil.New()
	g := b.Goal("g")
	q := b.DQ(g, "q")
	h := b.Hyp(q, "h")
	r := b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: b.Exp(h, event.RecipeNone),
		Verdicts: map[string]string{h: event.VerdictSupports}, Observation: "o",
		Readings: []event.ReadingItem{{Text: "is a reader a logic?"}}})
	k := event.ReadingKey(r, 1)
	p := replayB(b)
	wantOwed(t, p, q, "owner answer "+k)
	if got := p.OwedOf(g); len(got) != 1 || got[0].Kind != OwedAnswer {
		t.Fatalf("OwedOf(g) = %+v, want the answer item", got)
	}
	wantRefusal(t, p.SelectRefusal(q, h), "waits reading "+k)

	b.Withdraw(q)
	wantOwed(t, replayB(b), q, "owner answer "+k)
	b.Resume(q)

	b.As("owner").Emit(event.TypeCheckpoint, g, event.Checkpoint{Summary: "yes", Role: event.RoleOwner, Closes: []string{k}})
	wantOwed(t, replayB(b), q, "runner reading "+k)
}

// Three candidates where one Result judges two of them: the rule the
// question's rivals read (Unseparated) and the rule its candidates read
// (UnseparatedCandidates) name the same pairs.
func TestUnseparatedCandidatesReadsTheRivalRule(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	a, h2, c := b.Hyp(q, "a"), b.Hyp(q, "b"), b.Hyp(q, "c")
	for _, h := range []string{a, h2, c} {
		b.Supports(h)
	}
	b.Result(b.Exp(a, event.RecipeNone, h2), map[string]string{a: event.VerdictSupports, h2: inc})
	p := replayB(b)
	if got := p.Questions[q].Candidates; len(got) != 3 {
		t.Fatalf("candidates = %v, want three", got)
	}
	want := [][2]string{{a, c}, {h2, c}}
	got := p.UnseparatedCandidates(q)
	rivals := p.Unseparated(q)
	if len(got) != 2 || got[0] != want[0] || got[1] != want[1] || len(rivals) != 2 || rivals[0] != want[0] || rivals[1] != want[1] {
		t.Fatalf("unseparated candidates %v, rivals %v, want %v", got, rivals, want)
	}
}

// A depends_on edge between two Hypotheses of one question is no premise,
// for the ledger and for the facts hyp status prints.
func TestPremisesLeaveOutAnEdgeInsideTheQuestion(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	a, h2 := b.Hyp(q, "a"), b.Hyp(q, "b")
	b.Edge(a, event.EdgeDependsOn, h2)
	b.Supports(a)
	b.Supports(h2)
	p := replayB(b)
	if got := p.Premises(q); len(got) != 0 {
		t.Fatalf("premises = %+v, want none", got)
	}
	wantOwed(t, p, q, "owner choose "+q)
}
