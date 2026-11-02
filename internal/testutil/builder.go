// Package testutil builds deterministic hyp event scripts for tests.
package testutil

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/dat9uy/hyp/internal/event"
)

// B builds well-formed event scripts with deterministic ids that sort in
// creation order.
type B struct {
	Evs   []event.Event
	N     int
	actor string
}

// New returns a builder whose events are written by actor "a".
func New() *B { return &B{actor: "a"} }

// As switches the actor for following events.
func (b *B) As(actor string) *B { b.actor = actor; return b }

// Emit appends one event of type typ on node (the new id when node is "")
// and returns its id.
func (b *B) Emit(typ, node string, payload any) string {
	b.N++
	id := fmt.Sprintf("01JB%022d", b.N)
	if node == "" {
		node = id
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		panic(err)
	}
	ts := time.Date(2026, 10, 24, 0, 0, 0, 0, time.UTC).Add(time.Duration(b.N) * time.Second)
	b.Evs = append(b.Evs, event.Event{V: 1, ID: id, TS: ts.Format("2006-01-02T15:04:05.000000000Z"),
		Actor: b.actor, Type: typ, Node: node, Payload: raw, Why: "test"})
	return id
}

func (b *B) add(kind, parent, title string) string {
	return b.Emit(event.TypeNodeAdd, "", event.NodeAdd{Kind: kind, Parent: parent, Title: title})
}

func (b *B) Goal(title string) string        { return b.add(event.KindGoal, "", title) }
func (b *B) DQ(parent, title string) string  { return b.add(event.KindDesignQuestion, parent, title) }
func (b *B) Hyp(parent, title string) string { return b.add(event.KindHypothesis, parent, title) }

func (b *B) Exp(parent, recipe string, targets ...string) string {
	p := event.NodeAdd{Kind: event.KindExperiment, Parent: parent, Title: "exp", Recipe: recipe, Targets: targets}
	if recipe != event.RecipeNone {
		p.RecipeHash = "aaa"
	}
	return b.Emit(event.TypeNodeAdd, "", p)
}

func (b *B) Result(exp string, verdicts map[string]string) string {
	return b.Emit(event.TypeResultRecord, "", event.ResultRecord{Parent: exp, Verdicts: verdicts, Observation: "obs"})
}

func (b *B) Retract(r string) string { return b.Emit(event.TypeResultRetract, r, event.Empty{}) }
func (b *B) Abandon(h string) string { return b.Emit(event.TypeHypothesisAbandon, h, event.Empty{}) }

func (b *B) Select(dq, h, prev string, tradeoff bool) string {
	return b.Emit(event.TypeDQSelect, dq, event.DQSelect{Hypothesis: h, PrevSelect: prev, Tradeoff: tradeoff})
}

func (b *B) Unselect(dq, prev string) string {
	return b.Emit(event.TypeDQUnselect, dq, event.DQUnselect{PrevSelect: prev})
}

func (b *B) RequestVerify(dq, sel string) string {
	return b.Emit(event.TypeDQRequestVerify, dq, event.DQRequestVerify{Select: sel})
}

func (b *B) Defer(dq string) string    { return b.Emit(event.TypeDQDefer, dq, event.Empty{}) }
func (b *B) Withdraw(dq string) string { return b.Emit(event.TypeDQWithdraw, dq, event.Empty{}) }
func (b *B) Resume(dq string) string   { return b.Emit(event.TypeDQResume, dq, event.Empty{}) }

func (b *B) Confirm(dq, sel, role, verdict string) string {
	return b.Emit(event.TypeConfirm, dq, event.Confirm{Select: sel, Role: role, Verdict: verdict})
}

func (b *B) Reaffirm(n string) string { return b.Emit(event.TypeReaffirm, n, event.Empty{}) }

func (b *B) Edge(from, typ, to string) string {
	return b.Emit(event.TypeEdgeAdd, from, event.Edge{Type: typ, To: to})
}

func (b *B) Unedge(from, typ, to string) string {
	return b.Emit(event.TypeEdgeRemove, from, event.Edge{Type: typ, To: to})
}

// Supports records a supporting Result for h through a fresh Experiment.
func (b *B) Supports(h string) string {
	return b.Result(b.Exp(h, event.RecipeNone), map[string]string{h: event.VerdictSupports})
}

// Contradicts records a contradicting Result for h through a fresh Experiment.
func (b *B) Contradicts(h string) string {
	return b.Result(b.Exp(h, event.RecipeNone), map[string]string{h: event.VerdictContradicts})
}

// Answer selects h on dq as actor a and confirms it as verifier b, then
// restores the previous actor. It returns the select id.
func (b *B) Answer(dq, h, prev string) string {
	saved := b.actor
	sel := b.As("a").Select(dq, h, prev, false)
	b.As("b").Confirm(dq, sel, event.RoleVerifier, event.ConfirmSufficient)
	b.As(saved)
	return sel
}

// Chain is three answered questions under one Goal whose selected Hypotheses
// depend on each other: H1 depends_on H2 depends_on H3.
type Chain struct {
	B              *B
	G, Q1, Q2, Q3  string
	H1, H2, H3, E3 string
}

// BuildChain builds the Chain with every question answered.
func BuildChain() Chain {
	b := New()
	c := Chain{B: b}
	c.G = b.Goal("ship the data design")
	c.Q1, c.Q2, c.Q3 = b.DQ(c.G, "which API shape"), b.DQ(c.G, "which storage"), b.DQ(c.G, "which ingest path")
	c.H1, c.H2, c.H3 = b.Hyp(c.Q1, "REST over the ledger"), b.Hyp(c.Q2, "ledger table in Postgres"), b.Hyp(c.Q3, "nightly CSV import")
	b.Supports(c.H1)
	b.Supports(c.H2)
	c.E3 = b.Exp(c.H3, event.RecipeNone)
	b.Result(c.E3, map[string]string{c.H3: event.VerdictSupports})
	b.Edge(c.H2, event.EdgeDependsOn, c.H3)
	b.Edge(c.H1, event.EdgeDependsOn, c.H2)
	b.Answer(c.Q1, c.H1, "")
	b.Answer(c.Q2, c.H2, "")
	b.Answer(c.Q3, c.H3, "")
	return c
}
