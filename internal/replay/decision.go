package replay

import (
	"slices"

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
	// Chain is the verifier-loop signal's view: the insufficient verdicts in a
	// row on the current select.
	Chain []Verdict
	// Limits and Dossier are derived when a Projection is built; a record
	// stored in State holds neither.
	Limits  []Limit  // the limits recorded under the question, in creation order
	Dossier *Dossier // of the current select, or nil when there is none
}

// ConfirmRec is one applied confirm on a select. Items is the Dossier as it
// stood at the confirm's place in the merged log.
type ConfirmRec struct {
	ID, Role, Verdict, Actor, Why string
	Accepts                       []string
	Items                         []DossierItem
}

func newDecision() *Decision {
	return &Decision{Selects: map[string]SelectRec{}, Confirms: map[string][]ConfirmRec{}}
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
func (s *Decision) onSelect(rec SelectRec, prev string, continues bool) {
	if !continues {
		s.Chain = nil
	}
	s.Selects[rec.ID] = rec
	s.setCurrent(rec.ID, rec.ID, prev)
}

// onUnselect applies a dq.unselect and forgets the chain.
func (s *Decision) onUnselect(evID, prev string) {
	s.setCurrent(evID, "", prev)
	s.Chain = nil
}

// onAmend runs on an amend of the current select's Hypothesis. The Dossier
// holds the body, so the Confirms need nothing here. The chain restarts, so
// verdicts on the old body never count towards a loop.
func (s *Decision) onAmend() { s.Chain = nil }

// onConfirm applies confirm ev on select p.Select, whose Dossier just then
// was items. A verifier confirm on the current select extends the chain when
// insufficient and clears it when sufficient.
func (s *Decision) onConfirm(ev event.Event, p event.Confirm, items []DossierItem) {
	s.Confirms[p.Select] = append(s.Confirms[p.Select], ConfirmRec{ID: ev.ID, Role: p.Role, Verdict: p.Verdict,
		Actor: ev.Actor, Why: ev.Why, Accepts: p.Accepts, Items: items})
	if p.Role != event.RoleVerifier || p.Select != s.Current {
		return
	}
	switch p.Verdict {
	case event.ConfirmInsufficient:
		s.Chain = append(s.Chain, Verdict{ev.ID, ev.Actor})
	case event.ConfirmSufficient:
		s.Chain = nil
	}
}

// counting is the Confirm by role that counts on select id when the Dossier is
// now: the latest one that judged this Dossier (spec §7.4), or nil. A Confirm
// lapses when the Dossier differs from the one it judged and counts again if
// the Dossier returns to it.
func (s *Decision) counting(id, role string, now []DossierItem) *ConfirmRec {
	cs := s.Confirms[id]
	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].Role == role && slices.Equal(cs[i].Items, now) {
			return &cs[i]
		}
	}
	return nil
}

// latest is the last Confirm by role on select id, counting or not, or nil.
func (s *Decision) latest(id, role string) *ConfirmRec {
	cs := s.Confirms[id]
	for i := len(cs) - 1; i >= 0; i-- {
		if cs[i].Role == role {
			return &cs[i]
		}
	}
	return nil
}

// Counting is the Confirm by role that counts on the current select of a
// projected record, or nil.
func (s *Decision) Counting(role string) *ConfirmRec {
	if s == nil || s.Dossier == nil {
		return nil
	}
	return s.counting(s.Current, role, s.Dossier.Items)
}

// Latest is the last Confirm by role on the current select, or nil.
func (s *Decision) Latest(role string) *ConfirmRec {
	if s == nil {
		return nil
	}
	return s.latest(s.Current, role)
}

// ChangedSince lists what the current Dossier of a projected record holds and
// c's does not, then the reverse.
func (s *Decision) ChangedSince(c *ConfirmRec) (added, removed []DossierItem) {
	return diffItems(c.Items, s.Dossier.Items)
}

// accepted maps each limit key an owner Confirm accepted on select id with
// the Dossier now to that owner's actor: a sufficient owner Confirm with this
// Dossier accepted it, and no later owner insufficient has this Dossier.
func (s *Decision) accepted(id string, now []DossierItem) map[string]string {
	out := map[string]string{}
	for _, c := range s.Confirms[id] {
		if c.Role != event.RoleOwner || !slices.Equal(c.Items, now) {
			continue
		}
		if c.Verdict == event.ConfirmInsufficient {
			clear(out)
		}
		for _, k := range c.Accepts {
			out[k] = c.Actor
		}
	}
	return out
}

// ownerRequired derives whether select rec needs the owner: the question has
// two or more candidates now, or the select says so.
func ownerRequired(rec SelectRec, candidates []string) bool {
	return len(candidates) >= 2 || rec.Tradeoff || rec.ScopeChange
}

// reason is the first reason question id is not answered (spec §7.4), or ""
// when it is. live lists its live child Hypotheses, candidates its trade-off
// candidates, and stale says the question itself is stale. The Dossier is read
// only for a select with a Confirm to compare, since this runs on every event.
func (d *derivation) reason(id string, live, candidates []string, stale bool) string {
	st := d.st
	s := st.Decisions[id]
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
	case st.awaitsSubquestions(rec.Hyp, func(c string) QuestionState { return d.dqState(c).State }):
		return ReasonAwaitingSubquestions
	case len(s.Confirms[rec.ID]) == 0:
		return ReasonAwaitingVerifier
	}
	now := st.dossierItems(id, rec.Hyp)
	v := s.counting(rec.ID, event.RoleVerifier, now)
	switch {
	case v == nil:
		return ReasonAwaitingVerifier
	case v.Verdict == event.ConfirmInsufficient:
		return ReasonVerifierInsufficient
	}
	if ownerRequired(rec, candidates) {
		if o := s.counting(rec.ID, event.RoleOwner, now); o == nil || o.Verdict != event.ConfirmSufficient {
			return ReasonAwaitingOwner
		}
	}
	return ""
}

// candidateReason gives the open reason of a DQ with no current selection.
func (d *derivation) candidateReason(live []string) string {
	if len(live) == 0 {
		return ReasonNoCandidates
	}
	count := map[HypStatus]int{}
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

// readings derives every reading, in creation order. The first owner checkpoint
// that names a reading closes it. Readings of a retracted Result are left out.
func (st *State) readings(applied []event.Event) []Reading {
	closedBy := map[string]string{}
	for _, ev := range applied {
		if ev.Type != event.TypeCheckpoint {
			continue
		}
		var c event.Checkpoint
		_ = event.DecodePayload(ev, &c)
		for _, k := range c.Closes {
			if _, ok := closedBy[k]; !ok {
				closedBy[k] = ev.ID
			}
		}
	}
	var out []Reading
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] {
			continue
		}
		for i, text := range n.Readings {
			k := event.ReadingKey(id, i+1)
			out = append(out, Reading{Key: k, Result: id, Text: text, By: closedBy[k]})
		}
	}
	return out
}

// limits derives each question's limits. A limit is closed when a
// non-retracted Result closes it, accepted when a counting owner confirm on
// the question's current select accepted it (Decision.accepted), and open
// otherwise. Limits of a retracted Result are left out.
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
	accepted := map[string]map[string]string{}
	for _, id := range st.Order {
		n := st.Nodes[id]
		if n.Kind != event.KindResult || st.Retracted[id] || len(n.Limits) == 0 {
			continue
		}
		q := st.resultQuestion(id)
		s := st.Decisions[q]
		if _, ok := accepted[q]; !ok && s.Current != "" {
			accepted[q] = s.accepted(s.Current, st.dossierItems(q, s.Selects[s.Current].Hyp))
		}
		for i, text := range n.Limits {
			l := Limit{Key: event.LimitKey(id, i+1), Result: id, Text: text, Hyps: sortedKeys(n.Verdicts), State: LimitOpen}
			if by, ok := closedBy[l.Key]; ok {
				l.State, l.By = LimitClosed, by
			} else if actor, ok := accepted[q][l.Key]; ok {
				l.State, l.By = LimitAccepted, actor
			}
			if s.Current != "" {
				l.Gating = l.Gates(s.Selects[s.Current].Hyp)
			}
			out[q] = append(out[q], l)
		}
	}
	return out
}
