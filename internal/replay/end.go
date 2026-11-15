package replay

import (
	"fmt"

	"github.com/dat9uy/hyp/internal/event"
)

// The end is the one terminal act on a Goal (ADR 0023): the actor that
// added the Goal says, once and in their own words, that the Goal is wrong
// or outgrown. It is an item of an answer, keyed by the Goal's id. Replay
// stores it in Facts.Ended; the Goal's state reads ended from then on, and
// no later event moves it again. Nothing of an ended Goal is owed or waits,
// and every key of it is settled, so a later Goal may cite it.

// End is one applied end of a Goal: the answer event that made it, its
// actor and time, and the actor's words.
type End struct {
	Event, Actor, TS, Text string
}

// GoalEnded is the state of a Goal its adder ended. It stands in place of
// in_progress or complete, whatever the Goal's questions are.
const GoalEnded GoalState = "ended"

// CiteEnded is what a cite means when it names an ended Goal, or a key of
// one that was not otherwise settled.
const CiteEnded = "ended"

// checkEnds reports why an answer by actor may not make its ends: each
// names a Goal by its full id, the actor added that Goal, the Goal is not
// already ended, it is named once, and the same answer moves no key of it.
// moves are the answer's other items: answers, accepts and returns.
func (f *Facts) checkEnds(actor string, ends []event.KeyText, moves [][]event.KeyText) error {
	if len(ends) > event.MaxLimits {
		return fmt.Errorf("over limit: ends has %d items (max %d)", len(ends), event.MaxLimits)
	}
	seen := map[string]bool{}
	for _, it := range ends {
		n := f.Nodes[it.Key]
		if n == nil || n.Kind != event.KindGoal {
			return fmt.Errorf("bad end: %s is not a Goal; an end names a Goal by its id", it.Key)
		}
		g := event.DisplayID(event.KindGoal, it.Key)
		if seen[it.Key] {
			return fmt.Errorf("bad end: %s is named twice", g)
		}
		seen[it.Key] = true
		if e, ok := f.Ended[it.Key]; ok {
			return fmt.Errorf("bad end: %s is already ended, by %s in answer %s; a Goal ends once", g, e.Actor, e.Event)
		}
		if actor != n.Actor {
			return fmt.Errorf("bad end: only the actor that added a Goal ends it: %s added %s, and %s did not", n.Actor, g, actor)
		}
		for _, items := range moves {
			for _, m := range items {
				if f.keyGoal(m.Key) == it.Key {
					return fmt.Errorf("bad end: one answer ends %s and moves its key %s; an end is an answer of its own", g, f.keyDisplay(m.Key))
				}
			}
		}
	}
	return nil
}

// applyEnds records the ends of answer ev.
func (f *Facts) applyEnds(ev event.Event, ends []event.KeyText) {
	for _, it := range ends {
		f.Ended[it.Key] = End{Event: ev.ID, Actor: ev.Actor, TS: ev.TS, Text: it.Text}
	}
}

// EndOf returns the end of Goal g, and false when g is not ended.
func (p Projection) EndOf(g string) (End, bool) {
	e, ok := p.Ends[g]
	return e, ok
}

// Ended reports whether node id leads to an ended Goal.
func (p Projection) Ended(id string) bool {
	_, ok := p.Ends[p.GoalOf(id)]
	return ok
}

// EndedGoals lists the ended Goals that a write of ev would lead to: the
// Goal of its node, of its parent for a node.add or a result.record, or of
// each key an answer moves. A Goal's own add leads to no Goal before it.
// An answer's ends are left out: replay's rule refuses an end of an ended
// Goal by its own words.
func (p Projection) EndedGoals(ev event.Event) []string {
	var nodes []string
	switch ev.Type {
	case event.TypeAnswer:
		var a event.Answer
		if event.DecodePayload(ev, &a) == nil {
			for _, items := range a.Items() {
				for _, it := range items {
					nodes = append(nodes, event.KeyResult(it.Key))
				}
			}
		}
	case event.TypeNodeAdd:
		var a event.NodeAdd
		if event.DecodePayload(ev, &a) == nil {
			nodes = append(nodes, a.Parent)
		}
	case event.TypeResultRecord:
		var r event.ResultRecord
		if event.DecodePayload(ev, &r) == nil {
			nodes = append(nodes, r.Parent)
		}
	default:
		nodes = append(nodes, ev.Node)
	}
	var out []string
	for _, n := range nodes {
		if g := p.GoalOf(n); g != "" && !contains(out, g) {
			if _, ok := p.Ends[g]; ok {
				out = append(out, g)
			}
		}
	}
	return out
}

// AfterEnd lists, in merge order, each applied event that leads to a Goal
// after that Goal's end: a write the binary refuses, which replay applies
// when a merged log holds it. hyp check warns about each one.
func (p Projection) AfterEnd() []AfterEndEvent {
	var out []AfterEndEvent
	ended := map[string]string{}
	for _, ev := range p.Applied {
		for _, g := range p.endedSoFar(ev, ended) {
			out = append(out, AfterEndEvent{Event: ev, Goal: g, End: ended[g]})
		}
		if ev.Type == event.TypeAnswer {
			var a event.Answer
			_ = event.DecodePayload(ev, &a)
			for _, it := range a.Ends {
				ended[it.Key] = ev.ID
			}
		}
	}
	return out
}

// AfterEndEvent is one event AfterEnd lists: the event, the ended Goal it
// leads to, and the id of the answer that ended it.
type AfterEndEvent struct {
	Event     event.Event
	Goal, End string
}

// endedSoFar is EndedGoals of ev against the ends applied before it.
func (p Projection) endedSoFar(ev event.Event, ended map[string]string) []string {
	if len(ended) == 0 {
		return nil
	}
	q := p
	q.Ends = map[string]End{}
	for g := range ended {
		q.Ends[g] = End{}
	}
	return q.EndedGoals(ev)
}

// EndedRefusal is the refusal of a write that leads to ended Goal g: it
// names the end and its words, and what stays admitted.
func (p Projection) EndedRefusal(g string) string { return endedText(g, p.Ends[g]) }

func endedText(g string, e End) string {
	return fmt.Sprintf("%s is ended, by %s in answer %s: %q; nothing more is written under an ended Goal, and a new Goal may cite it",
		event.DisplayID(event.KindGoal, g), e.Actor, e.Event, e.Text)
}

// admitEnded is the first rule of every admission: a write whose node id
// leads to an ended Goal is refused, before any rule that reads the Goal's
// open work, so the refusal names the end.
func admitEnded(f *Facts, id string) []Finding {
	g := goalOf(f.Nodes, id)
	if e, ok := f.Ended[g]; ok {
		return []Finding{{FindRefuse, endedText(g, e)}}
	}
	return nil
}

// CitedGoal is the Goal that cite c names or leads to: a Goal's own id, a
// Hypothesis's Goal, or the Goal of a key; "" for any other cite.
func (p Projection) CitedGoal(c string) string {
	if n := p.Nodes[c]; n != nil {
		if n.Kind == event.KindGoal || n.Kind == event.KindHypothesis {
			return p.GoalOf(c)
		}
		return ""
	}
	if _, ok := event.ParseKey(c); ok {
		return p.KeyGoal(c)
	}
	return ""
}

// admitEnded is admitEnded on the projection's nodes and ends.
func (p Projection) admitEnded(id string) []Finding {
	return admitEnded(&Facts{Nodes: p.Nodes, Ended: p.Ends}, id)
}
