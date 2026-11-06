package replay

import (
	"errors"
	"fmt"

	"github.com/dat9uy/hyp/internal/event"
)

// rule is one event type's definition. validate is its replay rule: it takes
// Facts, so reading derived status from it does not compile (spec §5.1). apply
// changes stored state only, and takes State because node.add stores the live
// rivals at the add; every derived field is left to recompute. basis lists the
// nodes the applied event is a basis event for (spec §7.3), and is nil for a
// type that is never one. Admission stays out of this table, because its rules
// read derived status (ADR 0013).
type rule struct {
	validate func(*Facts, event.Event) error
	apply    func(*State, event.Event)
	basis    func(*Facts, event.Event) []string
}

// noApply is the apply of an event that changes no stored state.
func noApply(*State, event.Event) {}

// onNode is the basis of an event whose own node is the basis node.
func onNode(_ *Facts, ev event.Event) []string { return []string{ev.Node} }

// rules holds one definition per event type (spec §4).
var rules = map[string]rule{
	event.TypeNodeAdd: {
		validate: (*Facts).validateNodeAdd,
		apply: func(st *State, ev event.Event) {
			var p event.NodeAdd
			_ = event.DecodePayload(ev, &p)
			if p.Kind == event.KindExperiment {
				// The derived records still hold the state before this event.
				st.RivalsAtAdd[ev.ID] = st.LiveRivals(st.Nodes[p.Parent].Parent, p.Parent)
			}
			st.addNode(&Node{ID: ev.ID, Kind: p.Kind, Parent: p.Parent, Title: p.Title, Body: p.Body,
				Alias: p.Alias, Recipe: p.Recipe, RecipeHash: p.RecipeHash, DecidedBy: p.DecidedBy, Actor: ev.Actor,
				AddedBy: ev.ID, Targets: p.Targets})
			if p.Alias != "" {
				st.Aliases[p.Alias] = ev.ID
			}
			if p.Kind == event.KindGoal {
				st.GoalID = ev.ID
			}
			if p.Kind == event.KindDesignQuestion {
				st.Decisions[ev.ID] = newDecision()
			}
		},
	},
	event.TypeResultRecord: {
		validate: (*Facts).validateResult,
		apply: func(st *State, ev event.Event) {
			var p event.ResultRecord
			_ = event.DecodePayload(ev, &p)
			st.addNode(&Node{ID: ev.ID, Kind: event.KindResult, Parent: p.Parent, Actor: ev.Actor,
				AddedBy: ev.ID, Verdicts: p.Verdicts, VerdictOrder: event.VerdictOrder(ev.Payload), Observation: p.Observation, Evidence: p.Evidence,
				EvidenceHashes: p.EvidenceHashes, Limits: p.Limits, Closes: p.Closes, Readings: p.Readings})
			for h := range p.Verdicts {
				st.ResultsByHyp[h] = append(st.ResultsByHyp[h], ev.ID)
			}
		},
		basis: func(f *Facts, ev event.Event) []string {
			var out []string
			for _, h := range sortedKeys(f.Nodes[ev.ID].Verdicts) {
				if f.Nodes[ev.ID].Verdicts[h] == event.VerdictSupports {
					out = append(out, h)
				}
			}
			return out
		},
	},
	event.TypeResultRetract: {
		validate: func(f *Facts, ev event.Event) error {
			if err := decodeEmpty(ev); err != nil {
				return err
			}
			if _, err := f.need(ev.Node, event.KindResult); err != nil {
				return err
			}
			if f.Retracted[ev.Node] {
				return errors.New("already retracted: " + ev.Node)
			}
			return nil
		},
		apply: func(st *State, ev event.Event) { st.Retracted[ev.Node] = true },
	},
	event.TypeHypothesisAbandon: {
		validate: func(f *Facts, ev event.Event) error {
			if err := decodeEmpty(ev); err != nil {
				return err
			}
			if _, err := f.need(ev.Node, event.KindHypothesis); err != nil {
				return err
			}
			if f.Abandoned[ev.Node] {
				return errors.New("already abandoned: " + ev.Node)
			}
			return nil
		},
		apply: func(st *State, ev event.Event) { st.Abandoned[ev.Node] = true },
	},
	event.TypeHypothesisAmend: {
		validate: (*Facts).validateAmend,
		apply: func(st *State, ev event.Event) {
			var p event.HypothesisAmend
			_ = event.DecodePayload(ev, &p)
			n := st.Nodes[ev.Node]
			n.Body = p.Body
			if sel := st.Decisions[n.Parent]; sel != nil && sel.Current != "" && sel.Selects[sel.Current].Hyp == ev.Node {
				sel.onAmend()
			}
		},
	},
	event.TypeDQSelect: {
		validate: (*Facts).validateSelect,
		apply: func(st *State, ev event.Event) {
			var p event.DQSelect
			_ = event.DecodePayload(ev, &p)
			sel := st.Decisions[ev.Node]
			continues := sel.Current != "" && (p.Hypothesis == sel.Selects[sel.Current].Hyp || st.supersedes(p.Hypothesis, sel.Selects[sel.Current].Hyp))
			sel.onSelect(SelectRec{ID: ev.ID, Hyp: p.Hypothesis, Actor: ev.Actor, Tradeoff: p.Tradeoff,
				ScopeChange: p.ScopeChange, Why: ev.Why}, p.PrevSelect, continues)
		},
		basis: onNode,
	},
	event.TypeDQUnselect: {
		validate: func(f *Facts, ev event.Event) error {
			var p event.DQUnselect
			if err := decode(ev, &p); err != nil {
				return err
			}
			_, err := f.need(ev.Node, event.KindDesignQuestion)
			return err
		},
		apply: func(st *State, ev event.Event) {
			var p event.DQUnselect
			_ = event.DecodePayload(ev, &p)
			st.Decisions[ev.Node].onUnselect(ev.ID, p.PrevSelect)
		},
	},
	event.TypeExperimentWithdraw: {
		validate: func(f *Facts, ev event.Event) error {
			if err := decodeEmpty(ev); err != nil {
				return err
			}
			if _, err := f.need(ev.Node, event.KindExperiment); err != nil {
				return err
			}
			if f.Withdrawn[ev.Node] {
				return errors.New("already withdrawn: " + ev.Node)
			}
			if f.superseded(ev.Node) {
				return errors.New("already superseded: " + ev.Node)
			}
			if r := f.liveResult(ev.Node); r != "" {
				return fmt.Errorf("bad withdraw: %s has the non-retracted Result %s", ev.Node, r)
			}
			return nil
		},
		apply: func(st *State, ev event.Event) { st.Withdrawn[ev.Node] = true },
	},
	event.TypeDQDefer: {
		validate: validateHold,
		apply:    func(st *State, ev event.Event) { st.Hold[ev.Node] = DQDeferred },
	},
	event.TypeDQWithdraw: {
		validate: validateHold,
		apply:    func(st *State, ev event.Event) { st.Hold[ev.Node] = DQWithdrawn },
	},
	event.TypeDQResume: {
		validate: validateHold,
		apply:    func(st *State, ev event.Event) { st.Hold[ev.Node] = "" },
	},
	event.TypeConfirm: {
		validate: (*Facts).validateConfirm,
		apply: func(st *State, ev event.Event) {
			var p event.Confirm
			_ = event.DecodePayload(ev, &p)
			sel := st.Decisions[ev.Node]
			sel.onConfirm(ev, p, st.dossierItems(ev.Node, sel.Selects[p.Select].Hyp))
		},
	},
	event.TypeDQRequestVerify: {
		// Read-only history: replay accepts it, whatever it holds, and it does
		// nothing (spec §4).
		validate: func(*Facts, event.Event) error { return nil },
		apply:    noApply,
	},
	event.TypeReaffirm: {
		validate: func(f *Facts, ev event.Event) error {
			if err := decodeEmpty(ev); err != nil {
				return err
			}
			n, err := f.need(ev.Node, "")
			if err != nil {
				return err
			}
			if n.Kind != event.KindHypothesis && n.Kind != event.KindDesignQuestion {
				return fmt.Errorf("wrong kind: reaffirm on %s", n.Kind)
			}
			return nil
		},
		apply: noApply,
		basis: onNode,
	},
	event.TypeCheckpoint: {
		validate: func(f *Facts, ev event.Event) error {
			var p event.Checkpoint
			if err := decode(ev, &p); err != nil {
				return err
			}
			if _, err := f.need(ev.Node, event.KindGoal); err != nil {
				return err
			}
			if err := textLimit("summary", p.Summary, 1, event.MaxText); err != nil {
				return err
			}
			return f.checkClosesReadings(ev, p)
		},
		apply: noApply,
	},
	event.TypeEdgeAdd: {
		validate: (*Facts).validateEdge,
		apply: func(st *State, ev event.Event) {
			var p event.Edge
			_ = event.DecodePayload(ev, &p)
			st.EdgesFrom[ev.Node] = append(st.EdgesFrom[ev.Node], len(st.Edges))
			st.Edges = append(st.Edges, Edge{ID: ev.ID, From: ev.Node, Type: p.Type, To: p.To, Live: true})
			if p.Type == event.EdgeSupersedes {
				st.SupersededN[p.To]++
			}
		},
		basis: func(f *Facts, ev event.Event) []string {
			if f.Edges[len(f.Edges)-1].Type == event.EdgeDependsOn {
				return []string{ev.Node}
			}
			return nil
		},
	},
	event.TypeEdgeRemove: {
		validate: (*Facts).validateEdge,
		apply: func(st *State, ev event.Event) {
			var p event.Edge
			_ = event.DecodePayload(ev, &p)
			st.Edges[st.liveEdge(ev.Node, p.Type, p.To)].Live = false
			if p.Type == event.EdgeSupersedes {
				st.SupersededN[p.To]--
			}
		},
	},
}

// validateHold is the replay rule of dq.defer, dq.withdraw and dq.resume.
func validateHold(f *Facts, ev event.Event) error {
	if err := decodeEmpty(ev); err != nil {
		return err
	}
	_, err := f.need(ev.Node, event.KindDesignQuestion)
	return err
}
