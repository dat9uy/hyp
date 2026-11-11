package replay

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"strings"

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
			if p.Kind == event.KindExperiment && st.Nodes[p.Parent].Kind == event.KindHypothesis {
				// The derived records still hold the state before this event.
				// An Experiment under a DesignQuestion has no rivals: it
				// gives no verdict.
				st.RivalsAtAdd[ev.ID] = st.LiveRivals(st.Nodes[p.Parent].Parent, p.Parent)
			}
			rubric := p.Rubric
			if p.DecidedBy == event.DecidedReading {
				rubric = event.RubricReading // the older spelling (ADR 0018)
			}
			st.addNode(&Node{ID: ev.ID, Kind: p.Kind, Parent: p.Parent, Title: p.Title, Body: p.Body,
				Alias: p.Alias, Recipe: p.Recipe, RecipeHash: p.RecipeHash, DecidedBy: p.DecidedBy, Actor: ev.Actor,
				AddedBy: ev.ID, Targets: p.Targets, Aims: p.Aims, Rubric: rubric, RubricHash: p.RubricHash, Records: p.Records})
			if p.Alias != "" {
				st.Aliases[p.Alias] = ev.ID
			}
			if p.Kind == event.KindGoal {
				st.Goals = append(st.Goals, ev.ID)
				if len(p.Lines) > 0 {
					// The Goal's add counts as the one write of its lines, and
					// each cite keeps what it meant here (ADR 0020).
					cites := make([][]Cite, len(p.Lines))
					for i, l := range p.Lines {
						st.Lines[ev.ID] = append(st.Lines[ev.ID], l.Text)
						for _, c := range l.Cites {
							m, _ := st.citeMeaning(c)
							cites[i] = append(cites[i], Cite{Key: c, Meaning: m})
						}
					}
					st.LinesBy[ev.ID], st.LineCites[ev.ID] = ev.ID, cites
				}
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
				EvidenceHashes: p.EvidenceHashes, Limits: p.Limits, Closes: p.Closes, Readings: p.Readings, Applies: p.Applies,
				Subjects: runSubjects(p)})
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
			// Superseding already took it out of the live set; an abandon
			// would change nothing a status reads.
			if f.superseded(ev.Node) {
				return fmt.Errorf("bad abandon: %s is superseded", ev.Node)
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
			st.Nodes[ev.Node].Body = p.Body
		},
	},
	event.TypeHypothesisSubject: {
		validate: (*Facts).validateSubject,
		apply: func(st *State, ev event.Event) {
			var p event.HypothesisSubject
			_ = event.DecodePayload(ev, &p)
			st.Subjects[ev.Node] = p
		},
	},
	event.TypeDQSelect: {
		validate: (*Facts).validateSelect,
		apply: func(st *State, ev event.Event) {
			var p event.DQSelect
			_ = event.DecodePayload(ev, &p)
			// candidates reads stored facts only, so they are the candidates
			// just before this select. A scope_change flag in an old log is
			// history and changes nothing (ADR 0019).
			st.Decisions[ev.Node].onSelect(SelectRec{ID: ev.ID, Hyp: p.Hypothesis, Actor: ev.Actor,
				Why: ev.Why, Candidates: st.candidates(ev.Node)}, p.PrevSelect)
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
		validate: func(f *Facts, ev event.Event) error {
			if err := validateHold(f, ev); err != nil {
				return err
			}
			// A withdrawn question no longer matters, so parking it says
			// nothing; a resume brings it back first.
			if f.Hold[ev.Node] == DQWithdrawn {
				return fmt.Errorf("bad defer: %s is withdrawn; resume it first", ev.Node)
			}
			return nil
		},
		apply: func(st *State, ev event.Event) { st.Hold[ev.Node] = DQDeferred },
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
			// Read-only history: no command writes a confirm (ADR 0019). An
			// old owner confirm is replayed only as the accepts it made, at
			// its place in the log, and an old verifier confirm applies
			// nothing (ADR 0015).
			if p.Role == event.RoleVerifier {
				return
			}
			for _, k := range p.Accepts {
				st.applyLimitMove(HandAccept, k, ev)
			}
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
			// Stale is computed only for live nodes, so a reaffirm of one
			// that is not live would change nothing.
			if f.superseded(ev.Node) || f.Abandoned[ev.Node] || f.Hold[ev.Node] == DQWithdrawn {
				return fmt.Errorf("bad reaffirm: %s is not live", ev.Node)
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
			if err := checkCheckpointShape(p); err != nil {
				return err
			}
			if err := textLimit("summary", p.Summary, 0, event.MaxText); err != nil {
				return err
			}
			if err := f.checkLines(ev.Node, p); err != nil {
				return err
			}
			return f.checkClosesReadings(ev, p)
		},
		apply: func(st *State, ev event.Event) {
			var p event.Checkpoint
			_ = event.DecodePayload(ev, &p)
			st.applyReadingMoves(ev, p)
			st.applyLimitMoves(ev, event.Keys(p.Refers), p.Returns, p.Accepts)
			if len(p.Lines) > 0 {
				st.Lines[ev.Node], st.LinesBy[ev.Node] = p.Lines, ev.ID
			}
		},
	},
	event.TypeAnswer: {
		validate: (*Facts).validateAnswer,
		apply: func(st *State, ev event.Event) {
			var p event.Answer
			_ = event.DecodePayload(ev, &p)
			st.applyAnswers(ev, p.Answers)
			st.applyLimitMoves(ev, nil, event.Keys(p.Returns), event.Keys(p.Accepts))
		},
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

// ownerFields are the checkpoint fields only an old owner checkpoint holds:
// read-only history that replay applies and no command writes. The owner's
// acts are an answer now (ADR 0020).
var ownerFields = []string{"role", "closes", "revises", "accepts", "returns", "lines"}

// checkCheckpointShape refuses an empty checkpoint, and one that mixes the
// runner's refers with an owner field: owner and runner fields never mix in
// one event (ADR 0020).
func checkCheckpointShape(p event.Checkpoint) error {
	owner := map[string]bool{"role": p.Role != "", "closes": len(p.Closes) > 0, "revises": len(p.Revises) > 0,
		"accepts": len(p.Accepts) > 0, "returns": len(p.Returns) > 0, "lines": len(p.Lines) > 0}
	if len(p.Refers) > 0 {
		for _, f := range ownerFields {
			if owner[f] {
				return fmt.Errorf("bad checkpoint: owner and runner fields never mix in one event: refers is the runner's and %s the owner's; the owner's acts are an answer", f)
			}
		}
	}
	for _, f := range ownerFields {
		if owner[f] {
			// An old owner checkpoint is checked as it was written: its
			// summary held the owner's words.
			return textLimit("summary", p.Summary, 1, event.MaxText)
		}
	}
	if strings.TrimSpace(p.Summary) == "" && len(p.Refers) == 0 {
		return errors.New("bad checkpoint: empty; a checkpoint holds a summary, refers, or both")
	}
	for _, it := range p.Refers {
		if it.Text == "" && strings.TrimSpace(p.Summary) == "" {
			return fmt.Errorf("bad refers: %s has no reason; a refer says why no test within the Goal's lines can settle it, in its text or the summary", it.Key)
		}
		if it.Text != "" {
			if err := textLimit("text of "+it.Key, it.Text, 1, event.MaxText); err != nil {
				return err
			}
		}
	}
	return nil
}

// runnerFields are the fields a runner's checkpoint or event envelope holds
// and an answer never does: owner and runner fields never mix in one event.
var runnerFields = []string{"summary", "refers", "role", "closes", "revises", "lines", "node", "why"}

// validateAnswer is the replay rule of answer, the owner's act (ADR 0020).
// It has no node and no why, holds at least one key, and every item's text
// is the owner's words for that key.
func (f *Facts) validateAnswer(ev event.Event) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(bytes.TrimSpace(ev.Payload), &raw); err == nil {
		for _, k := range runnerFields {
			if _, ok := raw[k]; ok {
				return fmt.Errorf("bad answer: owner and runner fields never mix in one event: %s is the runner's; an answer holds only answers, accepts and returns", k)
			}
		}
	}
	if ev.Node != "" {
		return errors.New("bad answer: owner and runner fields never mix in one event: node is the runner's; an answer is on no node")
	}
	if ev.Why != "" {
		return errors.New("bad answer: owner and runner fields never mix in one event: why is the runner's; the owner's reason is each item's text")
	}
	var p event.Answer
	if err := decode(ev, &p); err != nil {
		return err
	}
	if len(p.Answers)+len(p.Accepts)+len(p.Returns) == 0 {
		return errors.New("bad answer: empty; an answer settles at least one key")
	}
	for _, items := range [][]event.KeyText{p.Answers, p.Accepts, p.Returns} {
		for _, it := range items {
			if err := textLimit("text of "+it.Key, it.Text, 1, event.MaxText); err != nil {
				return err
			}
		}
	}
	if err := f.checkAnswers(ev.Actor, p.Answers); err != nil {
		return err
	}
	return f.checkLimitMoves(ev.Actor, true, nil, event.Keys(p.Returns), event.Keys(p.Accepts))
}
