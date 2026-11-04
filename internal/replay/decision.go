package replay

import (
	"fmt"

	"github.com/dat9uy/hyp/internal/event"
)

// Decision is the record of one DesignQuestion's selection (spec §7.4), the
// one owner of its select history, conflicts, Confirms, owner requirement,
// limit acceptances, reason order and verifier-loop chain. Only the methods in
// this file write it. Render, signals, check and the CLI read it, through
// Projection.Decisions.
type Decision struct {
	Current   string
	Conflicts []string
	Selects   map[string]SelectRec
	// Confirms holds every applied confirm on each select, verifier and owner
	// in one list, in log order, each with the Dossier at its place in the log.
	Confirms map[string][]ConfirmRec
	// Chain and ChainRequest are the verifier-loop signal's view: the
	// insufficient verdicts on the current select and a request made after
	// the last of them.
	Chain        []Verdict
	ChainRequest string
	// Limits and Dossier are derived when a Projection is built; a record
	// stored in State holds neither.
	Limits  []Limit  // the limits recorded under the question, in creation order
	Dossier *Dossier // of the current select, or nil when there is none

	verdict   map[string]string     // select id → latest verifier verdict that still counts
	requested map[string]bool       // select id → re-verify asked since its last verifier confirm
	owner     map[string]string     // select id → latest owner verdict that still counts
	accepted  map[string]Acceptance // select id+"|"+limit key → the owner confirm that accepted it
}

// ConfirmRec is one applied confirm on a select. Items is the Dossier as it
// stood at the confirm's place in the merged log.
type ConfirmRec struct {
	ID, Role, Verdict, Actor, Why string
	Accepts                       []string
	Items                         []DossierItem
}

func newDecision() *Decision {
	return &Decision{Selects: map[string]SelectRec{}, Confirms: map[string][]ConfirmRec{}, verdict: map[string]string{},
		requested: map[string]bool{}, owner: map[string]string{}, accepted: map[string]Acceptance{}}
}

// setCurrent applies the conflict rule of spec §7.4 for the select or
// unselect event evID: a write whose prev_select is not the current selection
// records both as conflicting; a write that names the current selection clears
// the conflict set. The write still applies and current becomes next.
func (s *Decision) setCurrent(evID, next, prev string) {
	if prev == s.Current {
		s.Conflicts = nil
	} else {
		if s.Current != "" {
			s.Conflicts = appendUnique(s.Conflicts, s.Current)
		}
		s.Conflicts = appendUnique(s.Conflicts, evID)
	}
	s.Current = next
}

// onSelect applies dq.select rec. The chain survives only when the select
// continues the current one: the same Hypothesis or one that supersedes it.
// An open chain request is forgotten either way.
func (s *Decision) onSelect(rec SelectRec, prev string, continues bool) {
	if !continues {
		s.Chain = nil
	}
	s.ChainRequest = ""
	s.Selects[rec.ID] = rec
	s.setCurrent(rec.ID, rec.ID, prev)
}

// onUnselect applies a dq.unselect and forgets the chain and its request.
func (s *Decision) onUnselect(evID, prev string) {
	s.setCurrent(evID, "", prev)
	s.Chain, s.ChainRequest = nil, ""
}

// onAmend runs on an amend of the current select's Hypothesis. It holds every
// amend rule. The owner confirm and its acceptances judged the old body and
// stop counting. The counting verifier verdict drops only when it is
// sufficient: an insufficient verdict and an open request stay with the
// select, so a request is still possible after an amend. The chain is
// stricter and restarts with its request forgotten, so verdicts on the old
// body never count towards a loop. The record keeps the latest verifier
// confirm, so hyp show still prints it.
func (s *Decision) onAmend() {
	delete(s.owner, s.Current)
	s.dropAcceptances(s.Current)
	if s.verdict[s.Current] == event.ConfirmSufficient {
		delete(s.verdict, s.Current)
	}
	s.Chain, s.ChainRequest = nil, ""
}

// onConfirm applies confirm ev on select p.Select, whose Dossier just then
// was items.
func (s *Decision) onConfirm(ev event.Event, p event.Confirm, items []DossierItem) {
	s.Confirms[p.Select] = append(s.Confirms[p.Select], ConfirmRec{ID: ev.ID, Role: p.Role, Verdict: p.Verdict,
		Actor: ev.Actor, Why: ev.Why, Accepts: p.Accepts, Items: items})
	switch p.Role {
	case event.RoleVerifier:
		s.onVerifier(ev, p.Select, p.Verdict)
	case event.RoleOwner:
		s.owner[p.Select] = p.Verdict
		// An owner's no withdraws their sign-off on this select's limits. Only
		// a sufficient owner confirm carries accepts, so nothing is re-added.
		if p.Verdict == event.ConfirmInsufficient {
			s.dropAcceptances(p.Select)
		}
		for _, k := range p.Accepts {
			s.accepted[p.Select+"|"+k] = Acceptance{Select: p.Select, Actor: ev.Actor}
		}
	}
}

// onVerifier applies a verifier confirm ev on select id with verdict v. It
// closes any open request on that select, and on the current select it extends
// the chain when insufficient and clears it when sufficient.
func (s *Decision) onVerifier(ev event.Event, id, v string) {
	s.verdict[id] = v
	delete(s.requested, id)
	if rec, ok := s.Selects[id]; ok {
		rec.Verifier, rec.VerifierWhy, rec.RequestWhy = v, ev.Why, ""
		s.Selects[id] = rec
	}
	if id == s.Current {
		switch v {
		case event.ConfirmInsufficient:
			s.Chain = append(s.Chain, Verdict{ev.ID, ev.Actor})
		case event.ConfirmSufficient:
			s.Chain = nil
		}
		s.ChainRequest = ""
	}
}

// onRequest applies a dq.request_verify ev on select id. The chain records it
// only on the current select with a chain to answer.
func (s *Decision) onRequest(ev event.Event, id string) {
	s.requested[id] = true
	if rec, ok := s.Selects[id]; ok {
		rec.RequestWhy = ev.Why
		s.Selects[id] = rec
	}
	if id == s.Current && len(s.Chain) > 0 {
		s.ChainRequest = ev.ID
	}
}

// dropAcceptances forgets the limit acceptances recorded on select id.
func (s *Decision) dropAcceptances(id string) {
	for k, a := range s.accepted {
		if a.Select == id {
			delete(s.accepted, k)
		}
	}
}

// Verdict is the latest verifier verdict on select id that still counts, or "".
func (s *Decision) Verdict(id string) string { return s.verdict[id] }

// Requested reports whether a re-verify request is open on select id.
func (s *Decision) Requested(id string) bool { return s.requested[id] }

// CanRequest refuses a dq.request_verify on select id unless its counting
// verdict is insufficient and no request is open on it.
func (s *Decision) CanRequest(id string) error {
	switch s.verdict[id] {
	case event.ConfirmInsufficient:
	case "":
		return fmt.Errorf("bad request-verify: %s has no verifier confirm yet", id)
	default:
		return fmt.Errorf("bad request-verify: the latest verifier confirm on %s is not insufficient", id)
	}
	if s.requested[id] {
		return fmt.Errorf("bad request-verify: a request is already open on %s", id)
	}
	return nil
}

// reason is the first reason question id is not answered (spec §7.4), or ""
// when it is. live lists its live child Hypotheses, and stale says the
// question itself is stale.
func (d *derivation) reason(id string, live []string, stale bool) string {
	s := d.st.Decisions[id]
	if len(s.Conflicts) > 0 {
		return ReasonConflicting
	}
	if s.Current == "" {
		return d.candidateReason(live)
	}
	rec := s.Selects[s.Current]
	switch {
	case d.hypStatus(rec.Hyp) != HypSupported:
		return ReasonSelectedNotSupport
	case d.staleOf(rec.Hyp) != nil || stale:
		return ReasonSelectedStale
	case s.verdict[rec.ID] == event.ConfirmInsufficient && !s.requested[rec.ID]:
		return ReasonVerifierInsufficient
	case s.verdict[rec.ID] != event.ConfirmSufficient:
		return ReasonAwaitingVerifier
	case rec.OwnerRequired && s.owner[rec.ID] != event.ConfirmSufficient:
		return ReasonAwaitingOwner
	}
	return ""
}

// candidateReason gives the open reason of a DQ with no current selection.
func (d *derivation) candidateReason(live []string) string {
	if len(live) == 0 {
		return ReasonNoCandidates
	}
	count := map[string]int{}
	for _, c := range live {
		count[d.hypStatus(c)]++
	}
	switch {
	case count[HypProposed] == len(live):
		return ReasonUntested
	case count[HypContradicted] == len(live):
		return ReasonAllContradicted
	case count[HypSupported] == 0:
		return ReasonInconclusive
	}
	return ReasonSupportedUnsel
}

// decisions builds what Projection.Decisions holds: each question's stored
// record with its limits and the Dossier of its current select. Each is a
// copy, so a projection keeps its limits and Dossier as the log goes on.
func (st *State) decisions() map[string]*Decision {
	limits := st.limits()
	out := make(map[string]*Decision, len(st.Decisions))
	for q, s := range st.Decisions {
		c := *s
		c.Limits = limits[q]
		if s.Current != "" {
			c.Dossier = st.dossier(q, s.Selects[s.Current].Hyp)
		}
		out[q] = &c
	}
	return out
}

// limits derives each question's limits. A limit is closed when a
// non-retracted Result closes it, accepted when an owner confirm on the
// question's current select accepted it, and open otherwise. Limits of a
// retracted Result are left out.
func (st *State) limits() map[string][]Limit {
	closedBy := map[string]string{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] {
			continue
		}
		for _, k := range n.Closes {
			if _, ok := closedBy[k]; !ok {
				closedBy[k] = id
			}
		}
	}
	out := map[string][]Limit{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] || len(n.Limits) == 0 {
			continue
		}
		q := st.resultQuestion(id)
		s := st.Decisions[q]
		for i, text := range n.Limits {
			l := Limit{Key: event.LimitKey(id, i+1), Result: id, Text: text, Hyps: sortedKeys(n.Verdicts), State: LimitOpen}
			if by, ok := closedBy[l.Key]; ok {
				l.State, l.By = LimitClosed, by
			} else if a, ok := s.accepted[s.Current+"|"+l.Key]; ok {
				l.State, l.By = LimitAccepted, a.Actor
			}
			if s.Current != "" {
				l.Gating = l.Gates(s.Selects[s.Current].Hyp)
			}
			out[q] = append(out[q], l)
		}
	}
	return out
}
