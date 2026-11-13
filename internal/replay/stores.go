package replay

import (
	"encoding/json"
	"fmt"

	"github.com/dat9uy/hyp/internal/event"
)

// AliasKey is an alias in one store. An alias is a name local to its store:
// two stores of one history may hold the same alias, and a store holds each
// alias once.
type AliasKey struct{ Store, Alias string }

// Hashes are the current recipe hashes by store: each store reads its own
// recipe libraries, so a pin is checked against the hashes of the store
// that holds its Result. Default serves a store with no entry in ByStore.
type Hashes struct {
	Default map[string]string
	ByStore map[string]map[string]string
}

// Of returns the recipe hashes of store s.
func (h Hashes) Of(s string) map[string]string {
	if m, ok := h.ByStore[s]; ok {
		return m
	}
	return h.Default
}

// StoreName prints a store's path for a message: "." is the history's root.
func StoreName(s string) string {
	if s == "" {
		return "."
	}
	return s
}

// checkStore reports why ev crosses stores: an event is appended to the
// store that holds its node, so every node and key it moves lies in its own
// store. A Goal line's cite and a reading's of may name a key in another
// store; they are not moves, and are not checked here.
func (st *State) checkStore(ev event.Event) error {
	cross := func(what, id string) error {
		n := st.Nodes[id]
		if n == nil || n.Store == ev.Store {
			return nil
		}
		return fmt.Errorf("event crosses stores: its %s %s is in store %s and the event in store %s; an event is appended to the store that holds its node",
			what, event.DisplayID(n.Kind, id), StoreName(n.Store), StoreName(ev.Store))
	}
	switch ev.Type {
	case event.TypeNodeAdd:
		var p event.NodeAdd
		_ = json.Unmarshal(ev.Payload, &p)
		if err := cross("parent", p.Parent); err != nil {
			return err
		}
		for _, t := range p.Targets {
			if err := cross("target", t); err != nil {
				return err
			}
		}
		for _, k := range p.Aims {
			if err := cross("aimed key's node", event.KeyResult(k)); err != nil {
				return err
			}
		}
		return nil
	case event.TypeResultRecord:
		var p event.ResultRecord
		_ = json.Unmarshal(ev.Payload, &p)
		if err := cross("Experiment", p.Parent); err != nil {
			return err
		}
		for h := range p.Verdicts {
			if err := cross("judged Hypothesis", h); err != nil {
				return err
			}
		}
		for _, ks := range [][]string{p.Closes, p.Applies} {
			for _, k := range ks {
				if err := cross("key's node", event.KeyResult(k)); err != nil {
					return err
				}
			}
		}
		return nil
	case event.TypeAnswer:
		var p event.Answer
		_ = json.Unmarshal(ev.Payload, &p)
		for _, items := range [][]event.KeyText{p.Answers, p.Accepts, p.Returns} {
			for _, it := range items {
				if err := cross("key's node", event.KeyResult(it.Key)); err != nil {
					return err
				}
			}
		}
		return nil
	}
	if err := cross("node", ev.Node); err != nil {
		return err
	}
	switch ev.Type {
	case event.TypeEdgeAdd, event.TypeEdgeRemove:
		var p event.Edge
		_ = json.Unmarshal(ev.Payload, &p)
		if p.Type != event.EdgeExternalRef {
			return cross("edge's end", p.To)
		}
	case event.TypeCheckpoint:
		var p event.Checkpoint
		_ = json.Unmarshal(ev.Payload, &p)
		keys := append([]string{}, p.Closes...)
		for _, it := range p.Refers {
			keys = append(keys, it.Key)
		}
		for _, ks := range [][]string{keys, p.Revises, p.Accepts, p.Returns} {
			for _, k := range ks {
				if err := cross("key's node", event.KeyResult(k)); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// CitedGoals lists the Goals Goal g cites, in the history's Goal order: the
// Goals of the keys its lines cite, and of the keys its Results' readings
// read (of). A cite may name a key of a Goal in another store; g itself is
// left out.
func (p Projection) CitedGoals(g string) []string {
	cited := map[string]bool{}
	for _, l := range p.LinesOf(g) {
		for _, c := range l.Cites {
			cited[goalOf(p.Nodes, event.KeyResult(c.Key))] = true
		}
	}
	for _, r := range p.Readings {
		if r.Of != "" && goalOf(p.Nodes, r.Result) == g {
			cited[goalOf(p.Nodes, event.KeyResult(r.Of))] = true
		}
	}
	var out []string
	for _, gs := range p.Goals {
		if gs.GoalID != g && cited[gs.GoalID] {
			out = append(out, gs.GoalID)
		}
	}
	return out
}

// CitingGoals lists the Goals that cite Goal g (CitedGoals), in the
// history's Goal order.
func (p Projection) CitingGoals(g string) []string {
	var out []string
	for _, gs := range p.Goals {
		for _, c := range p.CitedGoals(gs.GoalID) {
			if c == g {
				out = append(out, gs.GoalID)
			}
		}
	}
	return out
}
