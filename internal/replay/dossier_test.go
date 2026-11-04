package replay

import (
	"slices"
	"testing"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

func hasItem(d *Dossier, it DossierItem) bool { return slices.Contains(d.Items, it) }

// selectedDossier is a question with a selected supported Hypothesis and one
// rival, ready for each test to add to.
func selectedDossier() (b *testutil.B, q, h1, h2, sel string) {
	b = testutil.New()
	q = b.DQ(b.Goal("g"), "q")
	h1 = b.Hyp(q, "h1")
	h2 = b.Hyp(q, "h2")
	b.Supports(h1)
	sel = b.Select(q, h1, "", false)
	return b, q, h1, h2, sel
}

func TestDossierIsPerCurrentSelect(t *testing.T) {
	b := testutil.New()
	q := b.DQ(b.Goal("g"), "q")
	h := b.Hyp(q, "h")
	if d := replayB(b).Decisions[q].Dossier; d != nil {
		t.Fatalf("no select, got a Dossier: %+v", d)
	}
	sel := b.Select(q, h, "", false)
	if replayB(b).Decisions[q].Dossier == nil {
		t.Fatal("a current select has no Dossier")
	}
	b.Unselect(q, sel)
	if d := replayB(b).Decisions[q].Dossier; d != nil {
		t.Fatalf("after an unselect, got a Dossier: %+v", d)
	}
}

func TestDossierHoldsBodyAndLiveHypotheses(t *testing.T) {
	b, q, h1, h2, _ := selectedDossier()
	h3 := b.Hyp(q, "h3")
	b.Abandon(h3)
	b.Emit(event.TypeHypothesisAmend, h1, event.HypothesisAmend{Body: "new body"})
	d := replayB(b).Decisions[q].Dossier
	for _, it := range []DossierItem{{Kind: DossierBody, Node: h1, Body: "new body"}, {Kind: DossierHyp, Node: h1}, {Kind: DossierHyp, Node: h2}} {
		if !hasItem(d, it) {
			t.Fatalf("missing %+v in %+v", it, d.Items)
		}
	}
	if hasItem(d, DossierItem{Kind: DossierHyp, Node: h3}) {
		t.Fatalf("an abandoned Hypothesis is listed: %+v", d.Items)
	}
}

func TestDossierHoldsAResultOutsideTheTreeWalk(t *testing.T) {
	b, q, h1, h2, _ := selectedDossier()
	// The Result sits under h1's Experiment, and its verdict is on h2.
	r := b.Result(b.Exp(h1, event.RecipeNone, h2), map[string]string{h2: con})
	d := replayB(b).Decisions[q].Dossier
	if !hasItem(d, DossierItem{Kind: DossierResult, Node: r, To: h2, Word: con}) {
		t.Fatalf("a verdict on the rival is missing: %+v", d.Items)
	}
}

func TestDossierDropsARetractedResult(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	r := b.Contradicts(h2)
	b.Retract(r)
	d := replayB(b).Decisions[q].Dossier
	if hasItem(d, DossierItem{Kind: DossierResult, Node: r, To: h2, Word: con}) {
		t.Fatalf("a retracted Result is listed: %+v", d.Items)
	}
}

func TestDossierHoldsOutgoingLinksOnly(t *testing.T) {
	b, q, h1, _, _ := selectedDossier()
	q2 := b.DQ(b.Evs[0].ID, "other")
	other := b.Hyp(q2, "other")
	b.Edge(h1, event.EdgeDependsOn, other)
	h1b := b.Hyp(q2, "points in")
	b.Edge(h1b, event.EdgeDependsOn, h1) // a link another question adds, pointing in
	b.Edge(h1, event.EdgeExternalRef, "gh:1")
	p := replayB(b)
	wantNoRejections(t, p)
	d := p.Decisions[q].Dossier
	if !hasItem(d, DossierItem{Kind: DossierLink, Node: h1, To: other, Word: event.EdgeDependsOn}) {
		t.Fatalf("the outgoing link is missing: %+v", d.Items)
	}
	for _, it := range d.Items {
		if it.Kind == DossierLink && (it.Node == h1b || it.Word == event.EdgeExternalRef) {
			t.Fatalf("listed %+v, an incoming link or external_ref", it)
		}
	}
}

func TestDossierDiffNamesAddedAndRemoved(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	r := b.Contradicts(h2)
	was := replayB(b).Decisions[q].Dossier
	if !was.Equal(replayB(b).Decisions[q].Dossier) {
		t.Fatal("the same log gave two different Dossiers")
	}
	b.Retract(r)
	h3 := b.Hyp(q, "h3")
	now := replayB(b).Decisions[q].Dossier
	if was.Equal(now) {
		t.Fatal("different Dossiers compare equal")
	}
	added, removed := was.Diff(now)
	if want := []DossierItem{{Kind: DossierHyp, Node: h3}}; !slices.Equal(added, want) {
		t.Fatalf("added %+v, want %+v", added, want)
	}
	if want := []DossierItem{{Kind: DossierResult, Node: r, To: h2, Word: con}}; !slices.Equal(removed, want) {
		t.Fatalf("removed %+v, want %+v", removed, want)
	}
}

func TestDecidingMarksTheOneContradicts(t *testing.T) {
	b, q, h1, h2, _ := selectedDossier()
	b.Supports(h2)
	r := b.Contradicts(h2)
	d := replayB(b).Decisions[q].Dossier
	if got := d.Deciding[h2]; !slices.Equal(got, []string{r}) {
		t.Fatalf("Deciding[h2] = %v, want [%s]", got, r)
	}
	if len(d.Deciding[h1]) != 0 {
		t.Fatalf("the selected Hypothesis is marked: %v", d.Deciding[h1])
	}
}

func TestDecidingLeavesOutTwoContradicts(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	b.Supports(h2)
	b.Contradicts(h2)
	b.Contradicts(h2)
	if got := replayB(b).Decisions[q].Dossier.Deciding[h2]; len(got) != 0 {
		t.Fatalf("two contradicts are marked: %v", got)
	}
}

func TestDecidingNeedsASupports(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	b.Contradicts(h2)
	if got := replayB(b).Decisions[q].Dossier.Deciding[h2]; len(got) != 0 {
		t.Fatalf("a rival with no supports is marked: %v", got)
	}
}

func TestDecidingFollowsRuleFourOfHypothesisStatus(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	b.Supports(h2)
	b.Contradicts(h2)
	b.DQ(h2, "open child") // without the contradicts it would await this, not be supported
	if got := replayB(b).Decisions[q].Dossier.Deciding[h2]; len(got) != 0 {
		t.Fatalf("a rival awaiting a sub-question is marked: %v", got)
	}
}

func TestDecisionKeepsEveryConfirmWithItsDossier(t *testing.T) {
	b, q, _, h2, sel := selectedDossier()
	v1 := b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmInsufficient)
	r := b.As("a").Supports(h2)
	o := b.As("o").Confirm(q, sel, event.RoleOwner, event.ConfirmInsufficient)
	v2 := b.As("v").Confirm(q, sel, event.RoleVerifier, event.ConfirmSufficient)
	p := replayB(b)
	dec := p.Decisions[q]
	got := dec.Confirms[sel]
	if len(got) != 3 || got[0].ID != v1 || got[1].ID != o || got[2].ID != v2 {
		t.Fatalf("confirms = %+v, want v1, o, v2 in log order", got)
	}
	it := DossierItem{Kind: DossierResult, Node: r, To: h2, Word: sup}
	if slices.Contains(got[0].Items, it) {
		t.Fatalf("the first confirm's Dossier holds a later Result: %+v", got[0].Items)
	}
	if !slices.Equal(got[2].Items, dec.Dossier.Items) {
		t.Fatalf("the last confirm's Dossier %+v differs from the current %+v", got[2].Items, dec.Dossier.Items)
	}
}

func TestAProjectionKeepsItsLimitsAndDossierAsTheLogGoesOn(t *testing.T) {
	b, q, _, h2, _ := selectedDossier()
	st, p := Build(b.Evs, nil, "")
	was := slices.Clone(p.Decisions[q].Dossier.Items)
	n := b.N
	b.Supports(h2)
	for _, ev := range b.Evs[n:] {
		_ = st.Step(ev, nil)
	}
	if !slices.Equal(p.Decisions[q].Dossier.Items, was) {
		t.Fatalf("an earlier projection's Dossier changed: %+v", p.Decisions[q].Dossier.Items)
	}
	if st.Decisions[q].Dossier != nil {
		t.Fatal("the stored record holds a Dossier")
	}
}
