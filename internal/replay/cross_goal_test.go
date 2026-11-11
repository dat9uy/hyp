package replay

import (
	"encoding/json"
	"fmt"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/dat9uy/hyp/internal/event"
	"github.com/dat9uy/hyp/internal/testutil"
)

// probeEvent is an event that would come next after b's script, by actor,
// for asking the move tables whether it may apply. An answer has no node
// and no why.
func probeEvent(t *testing.T, b *testutil.B, actor, typ, node string, payload any) event.Event {
	t.Helper()
	raw, err := json.Marshal(payload)
	if err != nil {
		t.Fatal(err)
	}
	n := b.N + 1
	ts := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC).Add(time.Duration(n) * time.Second)
	ev := event.Event{V: 1, ID: fmt.Sprintf("01JB%022d", n), TS: ts.Format("2006-01-02T15:04:05.000000000Z"),
		Actor: actor, Type: typ, Node: node, Payload: raw, Why: "test"}
	if typ == event.TypeAnswer {
		ev.Why = ""
	}
	return ev
}

// emitAnswer appends the owner o's answer to b's script.
func emitAnswer(t *testing.T, b *testutil.B, a event.Answer) {
	t.Helper()
	ev := probeEvent(t, b, "o", event.TypeAnswer, "", a)
	b.N++
	b.Evs = append(b.Evs, ev)
}

// goalWithLines adds a Goal whose add carries these lines.
func goalWithLines(b *testutil.B, title string, lines ...string) string {
	var ls []event.GoalLine
	for _, l := range lines {
		ls = append(ls, event.GoalLine{Text: l})
	}
	return b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: event.KindGoal, Title: title, Lines: ls})
}

func refer(b *testutil.B, g string, keys ...string) string {
	var items []event.KeyText
	for _, k := range keys {
		items = append(items, event.KeyText{Key: k, Text: "no test within the Goal can run it"})
	}
	return b.As("a").Emit(event.TypeCheckpoint, g, event.Checkpoint{Refers: items})
}

// crossGoalStore is a two-Goal store where every hand a limit or a line can
// be in occurs, and both named asymmetries occur.
type crossGoalStore struct {
	b              *testutil.B
	gA, gB         string
	runner, owner  []string // keys OWED must give the runner and the owner
	nobody         []string // keys in no one's hand: accepted, or closed
	asymmetries    map[string]string
	runnerNotOwned []string // keys only a runner moves that OWED owes nobody
}

func buildCrossGoalStore(t *testing.T) crossGoalStore {
	b := testutil.New()
	s := crossGoalStore{b: b, asymmetries: map[string]string{}}

	// Goal A: a live question with a candidate, whose Result holds a limit
	// in each hand, and a deferred and a withdrawn question, each with a
	// referred limit.
	s.gA = goalWithLines(b, "A", "line run on every candidate", "line no test can run")
	qA := b.DQ(s.gA, "qA")
	hA := b.Hyp(qA, "hA")
	rA := limitResult(b, hA, []string{"owed", "referred", "accepted", "closed", "returned"}, nil)
	k := func(r string, n int) string { return event.LimitKey(r, n) }
	refer(b, s.gA, k(rA, 2), k(rA, 3), k(rA, 5), event.LimitKey(s.gA, 2))
	limitResult(b, hA, nil, []string{k(rA, 4)})
	qDeferred := b.DQ(s.gA, "deferred")
	rD := limitResult(b, b.Hyp(qDeferred, "hD"), []string{"case on a deferred question"}, nil)
	qWithdrawn := b.DQ(s.gA, "withdrawn")
	rW := limitResult(b, b.Hyp(qWithdrawn, "hW"), []string{"case on a withdrawn question"}, nil)
	refer(b, s.gA, k(rD, 1), k(rW, 1))
	b.Defer(qDeferred)
	b.Withdraw(qWithdrawn)

	// Goal B: its one top-level question is withdrawn, so it has no live
	// top-level question; one line is referred, and one is not.
	s.gB = goalWithLines(b, "B", "line referred", "line not referred")
	qB := b.DQ(s.gB, "qB")
	b.Hyp(qB, "hB")
	refer(b, s.gB, event.LimitKey(s.gB, 1))
	b.Withdraw(qB)

	emitAnswer(t, b, event.Answer{
		Accepts: []event.KeyText{{Key: k(rA, 3), Text: "not this Goal"}},
		Returns: []event.KeyText{{Key: k(rA, 5), Text: "the Goal covers it"}}})

	s.runner = []string{k(rA, 1), k(rA, 5), event.LimitKey(s.gA, 1)}
	s.owner = []string{k(rA, 2), event.LimitKey(s.gA, 2)}
	s.nobody = []string{k(rA, 3), k(rA, 4)}
	// The two named asymmetries: each is accept-able by the move table and
	// owed by nobody.
	s.asymmetries[k(rD, 1)] = "a referred limit on a deferred question"
	s.asymmetries[k(rW, 1)] = "a referred limit on a withdrawn question"
	s.asymmetries[event.LimitKey(s.gB, 1)] = "a referred line on a Goal with no live top-level question"
	s.runnerNotOwned = []string{event.LimitKey(s.gB, 2)}
	return s
}

// In a two-Goal store the move tables and OWED never disagree on whose hand
// an item is in, except the two named asymmetries: a referred limit on a
// deferred or withdrawn question, and a referred line on a Goal with no live
// top-level question, each accept-able by the table but owed by nobody.
// OWED owes an item only in the Goal it leads to, and only that Goal's
// runner may move an item OWED gives the runner.
func TestMoveTablesAndOwedAgreeAcrossGoals(t *testing.T) {
	s := buildCrossGoalStore(t)
	st, p := Build(s.b.Evs, nil, "")
	if len(p.Rejections) > 0 {
		t.Fatalf("the fixture's events were rejected: %+v", p.Rejections)
	}
	owed := p.Owed()
	owedTo := func(k string, who OwedWho) bool {
		return slices.ContainsFunc(owed, func(o Owed) bool {
			return o.Key == k && o.Who == who && (o.Kind == OwedLimit || o.Kind == OwedReferred)
		})
	}
	admits := func(actor, typ, node string, payload any) error {
		return Validate(st, probeEvent(t, s.b, actor, typ, node, payload))
	}
	ownerAccepts := func(k string) bool {
		return admits("o", event.TypeAnswer, "", event.Answer{Accepts: []event.KeyText{{Key: k, Text: "t"}}}) == nil
	}
	runnerRefers := func(g, k string) error {
		return admits("a", event.TypeCheckpoint, g, event.Checkpoint{Refers: []event.KeyText{{Key: k, Text: "t"}}})
	}
	other := map[string]string{s.gA: s.gB, s.gB: s.gA}

	var keys []string
	keys = append(keys, s.runner...)
	keys = append(keys, s.owner...)
	keys = append(keys, s.nobody...)
	keys = append(keys, s.runnerNotOwned...)
	for k := range s.asymmetries {
		keys = append(keys, k)
	}
	slices.Sort(keys)
	for _, k := range keys {
		g := goalOf(p.Nodes, event.KeyResult(k))
		owner, runner, accept := owedTo(k, OwedOwner), owedTo(k, OwedRunner), ownerAccepts(k)
		switch {
		case owner && !accept:
			t.Errorf("%s: OWED gives it to the owner, but the move table refuses the owner's accept", k)
		case accept && !owner && s.asymmetries[k] == "":
			t.Errorf("%s: the move table admits the owner's accept, but OWED owes it to nobody, and it is not a named asymmetry", k)
		case runner && accept:
			t.Errorf("%s: OWED gives it to the runner, but the move table admits the owner's accept", k)
		}
		if why := s.asymmetries[k]; why != "" && (!accept || owner || runner) {
			t.Errorf("%s (%s): want accept-able and owed by nobody; accept %v, owner %v, runner %v", k, why, accept, owner, runner)
		}
		if slices.Contains(s.runner, k) != runner || slices.Contains(s.owner, k) != owner {
			t.Errorf("%s: OWED gives runner %v, owner %v; want runner %v, owner %v", k, runner, owner, slices.Contains(s.runner, k), slices.Contains(s.owner, k))
		}
		if runner {
			if err := runnerRefers(g, k); err != nil {
				t.Errorf("%s: OWED gives it to %s's runner, but the table refuses that runner's refer: %v", k, g, err)
			}
			if err := runnerRefers(other[g], k); err == nil {
				t.Errorf("%s: OWED gives it to %s's runner, but the table admits a refer from %s's checkpoint", k, g, other[g])
			}
		}
	}
	for _, o := range owed {
		if _, ok := event.ParseKey(o.Key); !ok {
			continue
		}
		if kg, qg := goalOf(p.Nodes, event.KeyResult(o.Key)), p.GoalOf(o.Question); kg != qg {
			t.Errorf("OWED owes %s on %s of %s, but the key leads to %s", o.Key, o.Question, qg, kg)
		}
	}
}

// The forbidden event of the agreement: a refer, on one Goal's checkpoint,
// of a limit OWED gives the other Goal's runner. The write is refused.
func TestCrossGoalReferOfAnOwedLimitRefusedWrite(t *testing.T) {
	b := testutil.New()
	gA := b.Goal("A")
	rA := limitResult(b, b.Hyp(b.DQ(gA, "qA"), "hA"), []string{"owed"}, nil)
	gB := b.Goal("B")
	st, _ := Build(b.Evs, nil, "")
	ev := probeEvent(t, b, "a", event.TypeCheckpoint, gB, event.Checkpoint{Refers: []event.KeyText{{Key: event.LimitKey(rA, 1), Text: "t"}}})
	if err := Validate(st, ev); err == nil || !strings.Contains(err.Error(), "a move never crosses Goals") {
		t.Fatalf("refer of %s from %s's checkpoint: err %v, want a refusal naming the rule", event.LimitKey(rA, 1), gB, err)
	}
}

// The replay half: a log holding that refer rejects it, and the limit stays
// the runner's.
func TestCrossGoalReferOfAnOwedLimitRejectedReplay(t *testing.T) {
	b := testutil.New()
	gA := b.Goal("A")
	rA := limitResult(b, b.Hyp(b.DQ(gA, "qA"), "hA"), []string{"owed"}, nil)
	gB := b.Goal("B")
	k := event.LimitKey(rA, 1)
	bad := refer(b, gB, k)
	p := replayB(b)
	wantRejected(t, p, bad, "a move never crosses Goals")
	if l, _ := p.Limit(k); l.Referred() {
		t.Fatalf("a rejected refer moved %s: %+v", k, l)
	}
}
