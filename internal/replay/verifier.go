package replay

import (
	"fmt"

	"github.com/dat9uy/hyp/internal/event"
)

// The methods below are the only writers of a question's verifier state:
// each select's counting verdict and open request, the Verifier, VerifierWhy
// and RequestWhy fields of its SelectRec, and the verifier-loop chain
// (spec §7.4).

func newSel() *Sel {
	return &Sel{Selects: map[string]SelectRec{}, verdict: map[string]string{}, requested: map[string]bool{}}
}

// onSelect runs before a dq.select becomes current. The chain survives only
// when the select continues the current one: the same Hypothesis or one that
// supersedes it. An open chain request is forgotten either way.
func (s *Sel) onSelect(continues bool) {
	if !continues {
		s.Chain = nil
	}
	s.ChainRequest = ""
}

// onUnselect forgets the chain and its request.
func (s *Sel) onUnselect() {
	s.Chain, s.ChainRequest = nil, ""
}

// onAmend runs on an amend of the current select's Hypothesis. It holds both
// amend rules. The current select's counting verdict drops only when it is
// sufficient: an insufficient verdict and an open request stay with the
// select, so a request is still possible after an amend. The chain is
// stricter and restarts with its request forgotten, so verdicts on the old
// body never count towards a loop. The record keeps the latest verifier
// confirm, so hyp show still prints it.
func (s *Sel) onAmend() {
	if s.verdict[s.Current] == event.ConfirmSufficient {
		delete(s.verdict, s.Current)
	}
	s.Chain, s.ChainRequest = nil, ""
}

// onConfirm applies a verifier confirm ev on select id with verdict v. It closes
// any open request on that select, and on the current select it extends the
// chain when insufficient and clears it when sufficient.
func (s *Sel) onConfirm(ev event.Event, id, v string) {
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
func (s *Sel) onRequest(ev event.Event, id string) {
	s.requested[id] = true
	if rec, ok := s.Selects[id]; ok {
		rec.RequestWhy = ev.Why
		s.Selects[id] = rec
	}
	if id == s.Current && len(s.Chain) > 0 {
		s.ChainRequest = ev.ID
	}
}

// Verdict is the latest verifier verdict on select id that still counts, or "".
func (s *Sel) Verdict(id string) string { return s.verdict[id] }

// Requested reports whether a re-verify request is open on select id.
func (s *Sel) Requested(id string) bool { return s.requested[id] }

// CanRequest refuses a dq.request_verify on select id unless its counting
// verdict is insufficient and no request is open on it.
func (s *Sel) CanRequest(id string) error {
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
